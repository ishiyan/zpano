/// Weighted signal ensemble aggregator.
///
/// Blends multiple independent signal sources into a single confidence
/// value in [0, 1]. Adaptive methods update weights based on observed
/// outcomes with a configurable feedback delay.
use std::collections::VecDeque;

use super::method::AggregationMethod;
use super::error_metric::ErrorMetric;

/// Parameters for constructing an Aggregator.
pub struct AggregatorParams {
    /// Number of signal sources (>= 1).
    pub n_signals: usize,
    /// Aggregation method to use. Defaults to `Equal`.
    pub method: AggregationMethod,
    /// Number of bars between signal observation and outcome availability (>= 1).
    pub feedback_delay: usize,
    /// Required for `Fixed` method. Normalized to sum to 1.0.
    pub weights: Option<Vec<f64>>,
    /// Rolling window size for `InverseVariance` and `RankBased` (>= 2).
    pub window: usize,
    /// Decay rate for `ExponentialDecay` (0 < alpha <= 1).
    pub alpha: f64,
    /// Learning rate for `MultiplicativeWeights` (> 0).
    pub eta: f64,
    /// Prior weights for `Bayesian`. Defaults to uniform.
    pub prior: Option<Vec<f64>>,
    /// Error metric for `InverseVariance` and `RankBased`.
    pub error_metric: ErrorMetric,
}

impl Default for AggregatorParams {
    fn default() -> Self {
        Self {
            n_signals: 1,
            method: AggregationMethod::Equal,
            feedback_delay: 1,
            weights: None,
            window: 50,
            alpha: 0.1,
            eta: 0.5,
            prior: None,
            error_metric: ErrorMetric::Absolute,
        }
    }
}

/// History entry for warmup: signals at bar T and outcome for bar T.
pub struct HistoryEntry {
    pub signals: Vec<f64>,
    pub outcome: f64,
}

/// Method-specific internal state.
enum MethodState {
    None,
    InverseVariance {
        errors: Vec<RollingWindow>,
        error_metric: ErrorMetric,
    },
    ExponentialDecay {
        ema: Vec<f64>,
        alpha: f64,
    },
    MultiplicativeWeights {
        log_weights: Vec<f64>,
        eta: f64,
    },
    RankBased {
        errors: Vec<RollingWindow>,
        error_metric: ErrorMetric,
    },
    Bayesian {
        log_posterior: Vec<f64>,
    },
}

/// Fixed-capacity ring buffer for rolling error windows.
struct RollingWindow {
    data: Vec<f64>,
    len: usize,
    start: usize,
    capacity: usize,
}

impl RollingWindow {
    fn new(capacity: usize) -> Self {
        Self {
            data: vec![0.0; capacity],
            len: 0,
            start: 0,
            capacity,
        }
    }

    fn append(&mut self, value: f64) {
        if self.len < self.capacity {
            self.data[self.len] = value;
            self.len += 1;
        } else {
            self.data[self.start] = value;
            self.start = (self.start + 1) % self.capacity;
        }
    }

    fn get(&self, index: usize) -> f64 {
        self.data[(self.start + index) % self.capacity]
    }

    fn sum(&self) -> f64 {
        let mut s = 0.0;
        for i in 0..self.len {
            s += self.get(i);
        }
        s
    }

    fn population_variance(&self) -> f64 {
        let n = self.len as f64;
        let mean = self.sum() / n;
        let mut v = 0.0;
        for i in 0..self.len {
            let diff = self.get(i) - mean;
            v += diff * diff;
        }
        v / n
    }
}

/// Weighted signal ensemble aggregator.
pub struct Aggregator {
    n: usize,
    method: AggregationMethod,
    feedback_delay: usize,
    count: usize,
    weights: Vec<f64>,
    ring: VecDeque<Vec<f64>>,
    ring_capacity: usize,
    state: MethodState,
}

impl Aggregator {
    /// Create a new Aggregator from the given parameters.
    ///
    /// # Errors
    ///
    /// Returns an error string if parameters are invalid.
    pub fn new(params: &AggregatorParams) -> Result<Self, String> {
        if params.n_signals < 1 {
            return Err(format!("n_signals must be >= 1, got {}", params.n_signals));
        }
        if params.feedback_delay < 1 {
            return Err(format!("feedback_delay must be >= 1, got {}", params.feedback_delay));
        }

        let n = params.n_signals;
        let ring_capacity = params.feedback_delay + 1;
        let inv_n = 1.0 / n as f64;

        let (weights, state) = match params.method {
            AggregationMethod::Fixed => {
                let w = params.weights.as_ref()
                    .ok_or_else(|| "FIXED method requires weights".to_string())?;
                if w.len() != n {
                    return Err(format!("weights length {} != n_signals {}", w.len(), n));
                }
                let s: f64 = w.iter().sum();
                if s <= 0.0 {
                    return Err("weights must sum to a positive value".to_string());
                }
                let normalized: Vec<f64> = w.iter().map(|v| v / s).collect();
                (normalized, MethodState::None)
            }
            AggregationMethod::Equal => {
                (vec![inv_n; n], MethodState::None)
            }
            AggregationMethod::InverseVariance => {
                if params.window < 2 {
                    return Err(format!("window must be >= 2, got {}", params.window));
                }
                let errors: Vec<RollingWindow> = (0..n)
                    .map(|_| RollingWindow::new(params.window))
                    .collect();
                (vec![inv_n; n], MethodState::InverseVariance {
                    errors,
                    error_metric: params.error_metric,
                })
            }
            AggregationMethod::ExponentialDecay => {
                if params.alpha <= 0.0 || params.alpha > 1.0 {
                    return Err(format!("alpha must be in (0, 1], got {}", params.alpha));
                }
                (vec![inv_n; n], MethodState::ExponentialDecay {
                    ema: vec![0.5; n], // neutral prior
                    alpha: params.alpha,
                })
            }
            AggregationMethod::MultiplicativeWeights => {
                if params.eta <= 0.0 {
                    return Err(format!("eta must be > 0, got {}", params.eta));
                }
                (vec![inv_n; n], MethodState::MultiplicativeWeights {
                    log_weights: vec![0.0; n], // uniform in log-space
                    eta: params.eta,
                })
            }
            AggregationMethod::RankBased => {
                if params.window < 2 {
                    return Err(format!("window must be >= 2, got {}", params.window));
                }
                let errors: Vec<RollingWindow> = (0..n)
                    .map(|_| RollingWindow::new(params.window))
                    .collect();
                (vec![inv_n; n], MethodState::RankBased {
                    errors,
                    error_metric: params.error_metric,
                })
            }
            AggregationMethod::Bayesian => {
                if let Some(ref prior) = params.prior {
                    if prior.len() != n {
                        return Err(format!("prior length {} != n_signals {}", prior.len(), n));
                    }
                    let s: f64 = prior.iter().sum();
                    if s <= 0.0 {
                        return Err("prior must sum to a positive value".to_string());
                    }
                    let normalized: Vec<f64> = prior.iter().map(|v| v / s).collect();
                    let log_post: Vec<f64> = normalized.iter().map(|v| v.ln()).collect();
                    (normalized, MethodState::Bayesian { log_posterior: log_post })
                } else {
                    let log_post = vec![inv_n.ln(); n];
                    (vec![inv_n; n], MethodState::Bayesian { log_posterior: log_post })
                }
            }
        };

        Ok(Self {
            n,
            method: params.method,
            feedback_delay: params.feedback_delay,
            count: 0,
            weights,
            ring: VecDeque::with_capacity(ring_capacity),
            ring_capacity,
            state,
        })
    }

    /// Blend signal sources into a single confidence value.
    ///
    /// `signals` must contain one value per source in [0, 1].
    /// Returns the blended confidence in [0, 1].
    ///
    /// # Errors
    ///
    /// Returns an error if `signals.len() != n_signals`.
    pub fn blend(&mut self, signals: &[f64]) -> Result<f64, String> {
        if signals.len() != self.n {
            return Err(format!("expected {} signals, got {}", self.n, signals.len()));
        }

        let output: f64 = (0..self.n)
            .map(|i| self.weights[i] * signals[i])
            .sum();

        // Append to ring buffer.
        if self.ring.len() >= self.ring_capacity {
            self.ring.pop_front();
        }
        self.ring.push_back(signals.to_vec());

        self.count += 1;
        Ok(output)
    }

    /// Provide outcome feedback for weight adaptation.
    ///
    /// For stateless methods (`Fixed`, `Equal`), this is a no-op.
    /// For adaptive methods, pairs the outcome with the buffered signals
    /// from `feedback_delay` bars ago and updates weights.
    /// `outcome` is the observed outcome in [0, 1].
    pub fn update(&mut self, outcome: f64) {
        if self.method == AggregationMethod::Fixed || self.method == AggregationMethod::Equal {
            return;
        }
        if self.ring.len() < self.feedback_delay + 1 {
            return;
        }

        // Retrieve signals from feedback_delay bars ago.
        let idx = self.ring.len() - 1 - self.feedback_delay;
        let past_signals = self.ring[idx].clone();

        match &mut self.state {
            MethodState::InverseVariance { errors, error_metric } => {
                Self::update_inverse_variance(
                    &past_signals, outcome, errors, *error_metric, self.n, &mut self.weights,
                );
            }
            MethodState::ExponentialDecay { ema, alpha } => {
                Self::update_exponential_decay(
                    &past_signals, outcome, ema, *alpha, self.n, &mut self.weights,
                );
            }
            MethodState::MultiplicativeWeights { log_weights, eta } => {
                Self::update_multiplicative_weights(
                    &past_signals, outcome, log_weights, *eta, self.n, &mut self.weights,
                );
            }
            MethodState::RankBased { errors, error_metric } => {
                Self::update_rank_based(
                    &past_signals, outcome, errors, *error_metric, self.n, &mut self.weights,
                );
            }
            MethodState::Bayesian { log_posterior } => {
                Self::update_bayesian(
                    &past_signals, outcome, log_posterior, self.n, &mut self.weights,
                );
            }
            MethodState::None => {}
        }
    }

    /// Replay historical data through `blend()` + `update()`.
    ///
    /// Each entry contains signals at bar T and the outcome for bar T.
    /// The method handles the feedback delay internally.
    pub fn warmup(&mut self, history: &[HistoryEntry]) -> Result<(), String> {
        let mut outcomes: Vec<f64> = Vec::with_capacity(history.len());
        for entry in history {
            self.blend(&entry.signals)?;
            outcomes.push(entry.outcome);
            let outcomes_len = outcomes.len();
            if outcomes_len > self.feedback_delay {
                let oidx = outcomes_len - 1 - self.feedback_delay;
                self.update(outcomes[oidx]);
            }
        }
        Ok(())
    }

    /// Current normalized weights (read-only copy).
    pub fn weights(&self) -> Vec<f64> {
        self.weights.clone()
    }

    /// Total number of `blend()` calls.
    pub fn count(&self) -> usize {
        self.count
    }

    // ── Private update methods ──────────────────────────────────────────

    /// Compute per-signal error using the configured metric.
    fn compute_error(error_metric: ErrorMetric, signal: f64, outcome: f64) -> f64 {
        let diff = signal - outcome;
        match error_metric {
            ErrorMetric::Absolute => diff.abs(),
            ErrorMetric::Squared => diff * diff,
        }
    }

    /// Update weights using inverse-variance of prediction errors.
    fn update_inverse_variance(
        signals: &[f64],
        outcome: f64,
        errors: &mut [RollingWindow],
        error_metric: ErrorMetric,
        n: usize,
        weights: &mut [f64],
    ) {
        let epsilon: f64 = 1e-15;

        for i in 0..n {
            let err = Self::compute_error(error_metric, signals[i], outcome);
            errors[i].append(err);
        }

        // Need at least 2 errors per signal to compute variance.
        for i in 0..n {
            if errors[i].len < 2 {
                return;
            }
        }

        let mut total: f64 = 0.0;
        for i in 0..n {
            let v = errors[i].population_variance();
            let raw = 1.0 / v.max(epsilon);
            weights[i] = raw;
            total += raw;
        }
        for i in 0..n {
            weights[i] /= total;
        }
    }

    /// Update weights using EMA of accuracy.
    fn update_exponential_decay(
        signals: &[f64],
        outcome: f64,
        ema: &mut [f64],
        alpha: f64,
        n: usize,
        weights: &mut [f64],
    ) {
        for i in 0..n {
            let err = (signals[i] - outcome).abs();
            let accuracy = 1.0 - err;
            ema[i] = alpha * accuracy + (1.0 - alpha) * ema[i];
        }

        // Normalize, clamping negative EMAs to 0.
        let mut total: f64 = 0.0;
        for i in 0..n {
            let clamped = ema[i].max(0.0);
            weights[i] = clamped;
            total += clamped;
        }
        if total > 0.0 {
            for i in 0..n {
                weights[i] /= total;
            }
        } else {
            let inv_n = 1.0 / n as f64;
            for i in 0..n {
                weights[i] = inv_n;
            }
        }
    }

    /// Update weights using the Hedge algorithm in log-space.
    fn update_multiplicative_weights(
        signals: &[f64],
        outcome: f64,
        log_weights: &mut [f64],
        eta: f64,
        n: usize,
        weights: &mut [f64],
    ) {
        for i in 0..n {
            let loss = (signals[i] - outcome).abs();
            log_weights[i] -= eta * loss;
        }

        // Softmax normalization (log-sum-exp trick).
        let max_log = log_weights.iter().cloned().fold(f64::NEG_INFINITY, f64::max);
        let mut total: f64 = 0.0;
        for i in 0..n {
            let exp_w = (log_weights[i] - max_log).exp();
            weights[i] = exp_w;
            total += exp_w;
        }
        for i in 0..n {
            weights[i] /= total;
        }
    }

    /// Update weights using rank of rolling accuracy.
    fn update_rank_based(
        signals: &[f64],
        outcome: f64,
        errors: &mut [RollingWindow],
        error_metric: ErrorMetric,
        n: usize,
        weights: &mut [f64],
    ) {
        for i in 0..n {
            let err = Self::compute_error(error_metric, signals[i], outcome);
            errors[i].append(err);
        }

        // Need at least 1 error per signal.
        for i in 0..n {
            if errors[i].len < 1 {
                return;
            }
        }

        // Compute mean accuracy per signal.
        let mut accuracies: Vec<f64> = Vec::with_capacity(n);
        for i in 0..n {
            let mean_error = errors[i].sum() / errors[i].len as f64;
            accuracies.push(1.0 - mean_error);
        }

        // Rank by accuracy (best = highest rank = n, worst = 1).
        // Ties get the average rank.
        let ranks = rank_with_ties(&accuracies);

        let total: f64 = ranks.iter().sum();
        if total > 0.0 {
            for i in 0..n {
                weights[i] = ranks[i] / total;
            }
        } else {
            let inv_n = 1.0 / n as f64;
            for i in 0..n {
                weights[i] = inv_n;
            }
        }
    }

    /// Update weights using Bayesian model averaging (Bernoulli likelihood).
    fn update_bayesian(
        signals: &[f64],
        outcome: f64,
        log_posterior: &mut [f64],
        n: usize,
        weights: &mut [f64],
    ) {
        let epsilon: f64 = 1e-15;

        for i in 0..n {
            // Clamp signal to [epsilon, 1 - epsilon] to avoid log(0).
            let sig = signals[i].max(epsilon).min(1.0 - epsilon);
            let log_lik = outcome * sig.ln() + (1.0 - outcome) * (1.0 - sig).ln();
            log_posterior[i] += log_lik;
        }

        // Softmax normalization.
        let max_log = log_posterior.iter().cloned().fold(f64::NEG_INFINITY, f64::max);
        let mut total: f64 = 0.0;
        for i in 0..n {
            let exp_w = (log_posterior[i] - max_log).exp();
            weights[i] = exp_w;
            total += exp_w;
        }
        for i in 0..n {
            weights[i] /= total;
        }
    }
}

/// Rank values from 1 (worst) to n (best), averaging ties.
fn rank_with_ties(values: &[f64]) -> Vec<f64> {
    let n = values.len();
    if n == 0 {
        return vec![];
    }

    // Sort indices by value.
    let mut indices: Vec<usize> = (0..n).collect();
    indices.sort_by(|&a, &b| values[a].partial_cmp(&values[b]).unwrap_or(std::cmp::Ordering::Equal));

    let mut ranks = vec![0.0; n];
    let mut i = 0;
    while i < n {
        // Find the end of the tie group.
        let mut j = i + 1;
        while j < n && values[indices[j]] == values[indices[i]] {
            j += 1;
        }
        // Average rank for this group (1-based).
        let avg_rank = (i + 1 + j) as f64 / 2.0;
        for k in i..j {
            ranks[indices[k]] = avg_rank;
        }
        i = j;
    }

    ranks
}

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, epsilon: f64) -> bool {
        (a - b).abs() < epsilon
    }

    // ── Validation tests ────────────────────────────────────────────

    #[test]
    fn test_validation_n_signals_zero() {
        let params = AggregatorParams { n_signals: 0, ..Default::default() };
        assert!(Aggregator::new(&params).is_err());
    }

    #[test]
    fn test_validation_feedback_delay_zero() {
        let params = AggregatorParams { feedback_delay: 0, ..Default::default() };
        assert!(Aggregator::new(&params).is_err());
    }

    #[test]
    fn test_validation_fixed_requires_weights() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Fixed, ..Default::default()
        };
        assert!(Aggregator::new(&params).is_err());
    }

    #[test]
    fn test_validation_fixed_weights_wrong_length() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Fixed,
            weights: Some(vec![1.0]), ..Default::default()
        };
        assert!(Aggregator::new(&params).is_err());
    }

    #[test]
    fn test_validation_fixed_weights_zero_sum() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Fixed,
            weights: Some(vec![0.0, 0.0]), ..Default::default()
        };
        assert!(Aggregator::new(&params).is_err());
    }

    #[test]
    fn test_validation_inverse_variance_window_too_small() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::InverseVariance,
            window: 1, ..Default::default()
        };
        assert!(Aggregator::new(&params).is_err());
    }

    #[test]
    fn test_validation_exponential_decay_alpha_zero() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::ExponentialDecay,
            alpha: 0.0, ..Default::default()
        };
        assert!(Aggregator::new(&params).is_err());
    }

    #[test]
    fn test_validation_multiplicative_weights_eta_zero() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::MultiplicativeWeights,
            eta: 0.0, ..Default::default()
        };
        assert!(Aggregator::new(&params).is_err());
    }

    #[test]
    fn test_validation_bayesian_prior_wrong_length() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Bayesian,
            prior: Some(vec![1.0]), ..Default::default()
        };
        assert!(Aggregator::new(&params).is_err());
    }

    #[test]
    fn test_validation_blend_wrong_signal_count() {
        let params = AggregatorParams { n_signals: 3, ..Default::default() };
        let mut agg = Aggregator::new(&params).unwrap();
        assert!(agg.blend(&[0.5, 0.5]).is_err());
    }

    // ── Fixed weights tests ─────────────────────────────────────────

    #[test]
    fn test_fixed_basic_blend() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::Fixed,
            weights: Some(vec![0.5, 0.3, 0.2]), ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();
        let result = agg.blend(&[1.0, 0.0, 0.0]).unwrap();
        assert!(almost_equal(result, 0.5, 1e-13));
    }

    #[test]
    fn test_fixed_weights_normalized() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Fixed,
            weights: Some(vec![2.0, 8.0]), ..Default::default()
        };
        let agg = Aggregator::new(&params).unwrap();
        let w = agg.weights();
        assert!(almost_equal(w[0], 0.2, 1e-13));
        assert!(almost_equal(w[1], 0.8, 1e-13));
    }

    #[test]
    fn test_fixed_update_is_noop() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Fixed,
            weights: Some(vec![0.6, 0.4]), ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();
        agg.blend(&[0.8, 0.2]).unwrap();
        agg.blend(&[0.7, 0.3]).unwrap();
        let w0_before = agg.weights()[0];
        agg.update(0.9);
        assert!(almost_equal(agg.weights()[0], w0_before, 1e-15));
    }

    #[test]
    fn test_fixed_blend_all_ones() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::Fixed,
            weights: Some(vec![0.5, 0.3, 0.2]), ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();
        let result = agg.blend(&[1.0, 1.0, 1.0]).unwrap();
        assert!(almost_equal(result, 1.0, 1e-13));
    }

    #[test]
    fn test_fixed_blend_all_zeros() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::Fixed,
            weights: Some(vec![0.5, 0.3, 0.2]), ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();
        let result = agg.blend(&[0.0, 0.0, 0.0]).unwrap();
        assert!(almost_equal(result, 0.0, 1e-13));
    }

    #[test]
    fn test_fixed_count_increments() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Fixed,
            weights: Some(vec![0.5, 0.5]), ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();
        assert_eq!(agg.count(), 0);
        agg.blend(&[0.5, 0.5]).unwrap();
        assert_eq!(agg.count(), 1);
        agg.blend(&[0.5, 0.5]).unwrap();
        assert_eq!(agg.count(), 2);
    }

    // ── Equal weights tests ─────────────────────────────────────────

    #[test]
    fn test_equal_basic_blend() {
        let params = AggregatorParams { n_signals: 3, ..Default::default() };
        let mut agg = Aggregator::new(&params).unwrap();
        let result = agg.blend(&[0.9, 0.3, 0.6]).unwrap();
        assert!(almost_equal(result, 0.6, 1e-13));
    }

    #[test]
    fn test_equal_single_signal() {
        let params = AggregatorParams { n_signals: 1, ..Default::default() };
        let mut agg = Aggregator::new(&params).unwrap();
        let result = agg.blend(&[0.7]).unwrap();
        assert!(almost_equal(result, 0.7, 1e-13));
    }

    #[test]
    fn test_equal_weights_are_uniform() {
        let params = AggregatorParams { n_signals: 4, ..Default::default() };
        let agg = Aggregator::new(&params).unwrap();
        for w in agg.weights() {
            assert!(almost_equal(w, 0.25, 1e-13));
        }
    }

    #[test]
    fn test_equal_update_is_noop() {
        let params = AggregatorParams { n_signals: 2, ..Default::default() };
        let mut agg = Aggregator::new(&params).unwrap();
        agg.blend(&[0.8, 0.2]).unwrap();
        agg.blend(&[0.7, 0.3]).unwrap();
        let w0_before = agg.weights()[0];
        agg.update(0.5);
        assert!(almost_equal(agg.weights()[0], w0_before, 1e-15));
    }

    // ── Inverse-variance tests ──────────────────────────────────────

    #[test]
    fn test_inverse_variance_initial_weights_uniform() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::InverseVariance,
            window: 10, ..Default::default()
        };
        let agg = Aggregator::new(&params).unwrap();
        for w in agg.weights() {
            assert!(almost_equal(w, 1.0 / 3.0, 1e-13));
        }
    }

    #[test]
    fn test_inverse_variance_accurate_signal_gets_higher_weight() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::InverseVariance,
            feedback_delay: 1, window: 10, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        let outcomes = [0.5, 0.6, 0.4, 0.55, 0.45, 0.5, 0.6, 0.4, 0.55, 0.45];
        for (i, &outcome) in outcomes.iter().enumerate() {
            let noise = if i % 2 == 0 { 0.01 } else { -0.01 };
            let s0 = outcome + noise;
            let s1 = if i % 2 == 0 { 0.9 } else { 0.1 };
            agg.blend(&[s0, s1]).unwrap();
            agg.update(outcome);
        }

        let w = agg.weights();
        assert!(w[0] > w[1]);
    }

    #[test]
    fn test_inverse_variance_squared_error_metric() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::InverseVariance,
            feedback_delay: 1, window: 10,
            error_metric: ErrorMetric::Squared, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        for _ in 0..5 {
            agg.blend(&[0.5, 0.5]).unwrap();
            agg.update(0.5);
        }
        let w = agg.weights();
        assert!(almost_equal(w[0], w[1], 1e-10));
    }

    // ── Exponential decay tests ─────────────────────────────────────

    #[test]
    fn test_exponential_decay_initial_weights_uniform() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::ExponentialDecay, ..Default::default()
        };
        let agg = Aggregator::new(&params).unwrap();
        for w in agg.weights() {
            assert!(almost_equal(w, 1.0 / 3.0, 1e-13));
        }
    }

    #[test]
    fn test_exponential_decay_good_signal_weight_increases() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::ExponentialDecay,
            feedback_delay: 1, alpha: 0.3, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        for _ in 0..20 {
            agg.blend(&[0.8, 0.2]).unwrap();
            agg.update(0.8);
        }

        let w = agg.weights();
        assert!(w[0] > w[1]);
    }

    #[test]
    fn test_exponential_decay_alpha_one() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::ExponentialDecay,
            feedback_delay: 1, alpha: 1.0, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        agg.blend(&[0.9, 0.1]).unwrap();
        agg.blend(&[0.9, 0.1]).unwrap();
        agg.update(0.9);
        let w = agg.weights();
        assert!(almost_equal(w[0], 1.0 / 1.2, 1e-13));
        assert!(almost_equal(w[1], 0.2 / 1.2, 1e-13));
    }

    // ── Multiplicative weights tests ────────────────────────────────

    #[test]
    fn test_multiplicative_weights_initial_weights_uniform() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::MultiplicativeWeights, ..Default::default()
        };
        let agg = Aggregator::new(&params).unwrap();
        for w in agg.weights() {
            assert!(almost_equal(w, 1.0 / 3.0, 1e-13));
        }
    }

    #[test]
    fn test_multiplicative_weights_best_signal_converges() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::MultiplicativeWeights,
            feedback_delay: 1, eta: 0.5, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        for _ in 0..50 {
            agg.blend(&[0.8, 0.2, 0.3]).unwrap();
            agg.update(0.8);
        }

        let w = agg.weights();
        assert!(w[0] > 0.5);
        assert!(w[0] > w[1]);
        assert!(w[0] > w[2]);
    }

    #[test]
    fn test_multiplicative_weights_high_eta_faster_convergence() {
        let mut results = [0.0f64; 2];
        let etas = [0.1, 1.0];
        for (e, &eta) in etas.iter().enumerate() {
            let params = AggregatorParams {
                n_signals: 2, method: AggregationMethod::MultiplicativeWeights,
                feedback_delay: 1, eta, ..Default::default()
            };
            let mut agg = Aggregator::new(&params).unwrap();
            for _ in 0..10 {
                agg.blend(&[0.9, 0.1]).unwrap();
                agg.update(0.9);
            }
            results[e] = agg.weights()[0];
        }
        assert!(results[1] > results[0]);
    }

    // ── Rank-based tests ────────────────────────────────────────────

    #[test]
    fn test_rank_based_initial_weights_uniform() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::RankBased,
            window: 10, ..Default::default()
        };
        let agg = Aggregator::new(&params).unwrap();
        for w in agg.weights() {
            assert!(almost_equal(w, 1.0 / 3.0, 1e-13));
        }
    }

    #[test]
    fn test_rank_based_rank_ordering() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::RankBased,
            feedback_delay: 1, window: 10, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        for _ in 0..15 {
            agg.blend(&[0.7, 0.5, 0.2]).unwrap();
            agg.update(0.7);
        }

        let w = agg.weights();
        assert!(w[0] > w[1]);
        assert!(w[1] > w[2]);
    }

    #[test]
    fn test_rank_based_ties_get_average_rank() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::RankBased,
            feedback_delay: 1, window: 10, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        for _ in 0..5 {
            agg.blend(&[0.5, 0.5]).unwrap();
            agg.update(0.5);
        }

        let w = agg.weights();
        assert!(almost_equal(w[0], w[1], 1e-13));
    }

    // ── Bayesian tests ──────────────────────────────────────────────

    #[test]
    fn test_bayesian_uniform_prior() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::Bayesian, ..Default::default()
        };
        let agg = Aggregator::new(&params).unwrap();
        for w in agg.weights() {
            assert!(almost_equal(w, 1.0 / 3.0, 1e-13));
        }
    }

    #[test]
    fn test_bayesian_custom_prior() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::Bayesian,
            prior: Some(vec![0.5, 0.3, 0.2]), ..Default::default()
        };
        let agg = Aggregator::new(&params).unwrap();
        let w = agg.weights();
        assert!(almost_equal(w[0], 0.5, 1e-13));
        assert!(almost_equal(w[1], 0.3, 1e-13));
        assert!(almost_equal(w[2], 0.2, 1e-13));
    }

    #[test]
    fn test_bayesian_good_predictor_dominates() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Bayesian,
            feedback_delay: 1, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        for _ in 0..20 {
            agg.blend(&[0.9, 0.1]).unwrap();
            agg.update(0.9);
        }

        let w = agg.weights();
        assert!(w[0] > 0.9);
    }

    #[test]
    fn test_bayesian_evidence_overrides_prior() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Bayesian,
            feedback_delay: 1, prior: Some(vec![0.1, 0.9]), ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        for _ in 0..50 {
            agg.blend(&[0.8, 0.2]).unwrap();
            agg.update(0.8);
        }

        let w = agg.weights();
        assert!(w[0] > w[1]);
    }

    // ── Delayed feedback tests ──────────────────────────────────────

    #[test]
    fn test_delayed_feedback_delay_1() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::ExponentialDecay,
            feedback_delay: 1, alpha: 1.0, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        agg.blend(&[0.9, 0.1]).unwrap();
        agg.blend(&[0.5, 0.5]).unwrap();
        agg.update(0.9);

        let w = agg.weights();
        assert!(almost_equal(w[0], 1.0 / 1.2, 1e-13));
        assert!(almost_equal(w[1], 0.2 / 1.2, 1e-13));
    }

    #[test]
    fn test_delayed_feedback_delay_2() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::ExponentialDecay,
            feedback_delay: 2, alpha: 1.0, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        agg.blend(&[0.9, 0.1]).unwrap();
        agg.blend(&[0.5, 0.5]).unwrap();
        // Not enough history (need 3 entries).
        agg.update(0.9);
        for w in &agg.weights() {
            assert!(almost_equal(*w, 0.5, 1e-13));
        }

        agg.blend(&[0.3, 0.7]).unwrap();
        agg.update(0.9); // pairs with first blend [0.9, 0.1]

        let w = agg.weights();
        assert!(almost_equal(w[0], 1.0 / 1.2, 1e-13));
        assert!(almost_equal(w[1], 0.2 / 1.2, 1e-13));
    }

    #[test]
    fn test_delayed_feedback_update_without_enough_history() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::ExponentialDecay,
            feedback_delay: 3, alpha: 0.5, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        agg.blend(&[0.5, 0.5]).unwrap();
        let w0_before = agg.weights()[0];
        agg.update(0.5);
        assert!(almost_equal(agg.weights()[0], w0_before, 1e-15));
    }

    // ── Warmup tests ────────────────────────────────────────────────

    #[test]
    fn test_warmup_equals_live_replay() {
        let history = vec![
            HistoryEntry { signals: vec![0.8, 0.3], outcome: 0.7 },
            HistoryEntry { signals: vec![0.6, 0.5], outcome: 0.5 },
            HistoryEntry { signals: vec![0.9, 0.2], outcome: 0.8 },
            HistoryEntry { signals: vec![0.7, 0.4], outcome: 0.6 },
            HistoryEntry { signals: vec![0.5, 0.6], outcome: 0.4 },
        ];

        // Method 1: warmup.
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::ExponentialDecay,
            feedback_delay: 1, alpha: 0.2, ..Default::default()
        };
        let mut agg1 = Aggregator::new(&params).unwrap();
        agg1.warmup(&history).unwrap();

        // Method 2: manual replay.
        let mut agg2 = Aggregator::new(&params).unwrap();
        let outcomes = [0.7, 0.5, 0.8, 0.6, 0.4];
        for (i, entry) in history.iter().enumerate() {
            agg2.blend(&entry.signals).unwrap();
            if i >= 1 {
                agg2.update(outcomes[i - 1]);
            }
        }

        assert_eq!(agg1.count(), agg2.count());
        for (w1, w2) in agg1.weights().iter().zip(agg2.weights().iter()) {
            assert!(almost_equal(*w1, *w2, 1e-13));
        }
    }

    #[test]
    fn test_warmup_bayesian_calibrated() {
        let history: Vec<HistoryEntry> = (0..20)
            .map(|_| HistoryEntry { signals: vec![0.9, 0.1], outcome: 0.9 })
            .collect();

        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Bayesian,
            feedback_delay: 1, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();
        agg.warmup(&history).unwrap();

        assert!(agg.weights()[0] > 0.9);
    }

    // ── Edge case tests ─────────────────────────────────────────────

    #[test]
    fn test_edge_single_signal_all_methods() {
        let methods = [
            AggregationMethod::Fixed,
            AggregationMethod::Equal,
            AggregationMethod::InverseVariance,
            AggregationMethod::ExponentialDecay,
            AggregationMethod::MultiplicativeWeights,
            AggregationMethod::RankBased,
            AggregationMethod::Bayesian,
        ];
        for method in methods {
            let params = AggregatorParams {
                n_signals: 1, method,
                weights: if method == AggregationMethod::Fixed { Some(vec![1.0]) } else { None },
                ..Default::default()
            };
            let mut agg = Aggregator::new(&params).unwrap();
            let result = agg.blend(&[0.73]).unwrap();
            assert!(almost_equal(result, 0.73, 1e-13), "failed for {:?}", method);
        }
    }

    #[test]
    fn test_edge_extreme_signals() {
        let params = AggregatorParams { n_signals: 2, ..Default::default() };
        let mut agg = Aggregator::new(&params).unwrap();
        let result = agg.blend(&[0.0, 1.0]).unwrap();
        assert!(almost_equal(result, 0.5, 1e-13));
    }

    #[test]
    fn test_edge_bayesian_extreme_signals() {
        let params = AggregatorParams {
            n_signals: 2, method: AggregationMethod::Bayesian,
            feedback_delay: 1, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();
        agg.blend(&[0.0, 1.0]).unwrap();
        agg.blend(&[0.0, 1.0]).unwrap();
        agg.update(1.0);
        let w_sum: f64 = agg.weights().iter().sum();
        assert!(almost_equal(w_sum, 1.0, 1e-13));
    }

    #[test]
    fn test_edge_inverse_variance_identical_signals() {
        let params = AggregatorParams {
            n_signals: 3, method: AggregationMethod::InverseVariance,
            feedback_delay: 1, window: 10, ..Default::default()
        };
        let mut agg = Aggregator::new(&params).unwrap();

        for _ in 0..5 {
            agg.blend(&[0.5, 0.5, 0.5]).unwrap();
            agg.update(0.5);
        }

        for w in agg.weights() {
            assert!(almost_equal(w, 1.0 / 3.0, 1e-10));
        }
    }

    #[test]
    fn test_edge_weights_sum_to_one_all_methods() {
        let methods = [
            AggregationMethod::Fixed,
            AggregationMethod::Equal,
            AggregationMethod::InverseVariance,
            AggregationMethod::ExponentialDecay,
            AggregationMethod::MultiplicativeWeights,
            AggregationMethod::RankBased,
            AggregationMethod::Bayesian,
        ];
        for method in methods {
            let params = AggregatorParams {
                n_signals: 2, method,
                weights: if method == AggregationMethod::Fixed { Some(vec![0.3, 0.7]) } else { None },
                ..Default::default()
            };
            let agg = Aggregator::new(&params).unwrap();
            let w_sum: f64 = agg.weights().iter().sum();
            assert!(almost_equal(w_sum, 1.0, 1e-13), "weights don't sum to 1 for {:?}", method);
        }
    }
}
