use crate::daycounting::conventions::DayCountConvention;
use crate::daycounting::fractional::{self, DateTime};
use super::periodicity::Periodicity;

const SQRT2: f64 = 1.4142135623730950488016887242097;

/// Ratios accumulates portfolio returns incrementally and computes
/// various financial performance ratios at each step.
pub struct Ratios {
    periodicity: Periodicity,
    periods_per_annum: i32,
    days_per_period: f64,
    risk_free_rate: f64,
    required_return: f64,
    day_count_convention: DayCountConvention,
    rolling_window: Option<usize>,
    min_periods: Option<usize>,

    fractional_periods: Vec<f64>,
    returns: Vec<f64>,
    sample_count: usize,

    logret_sum: f64,
    drawdowns_cumulative: Vec<f64>,
    drawdowns_cumulative_min: f64,
    drawdowns_peaks: Vec<f64>,
    drawdowns_peaks_peak: usize,
    drawdown_continuous: Vec<f64>,
    drawdown_continuous_final: Vec<f64>,
    drawdown_continuous_finalized: bool,
    drawdown_continuous_peak: usize,
    drawdown_continuous_inside: bool,
    cumulative_return_plus1: f64,
    cumulative_return_plus1_max: f64,
    cumulative_return_geometric_mean: Option<f64>,
    returns_mean: Option<f64>,
    returns_std: Option<f64>,
    returns_autocorr_penalty: f64,
    excess_mean: Option<f64>,
    excess_std: Option<f64>,
    excess_autocorr_penalty: f64,
    required_mean: Option<f64>,
    required_lpm1: Option<f64>,
    required_lpm2: Option<f64>,
    required_lpm3: Option<f64>,
    required_hpm1: Option<f64>,
    required_hpm2: Option<f64>,
    required_hpm3: Option<f64>,
    required_autocorr_penalty: f64,
    avg_return: Option<f64>,
    avg_win: Option<f64>,
    avg_loss: Option<f64>,
    win_rate: Option<f64>,
    total_duration: f64,

    reset_called: bool,
}

impl Ratios {
    /// Creates a new Ratios instance with the specified parameters.
    /// Annual rates are de-annualized to per-period rates based on periodicity.
    /// rolling_window, if Some, limits computations to the last N returns.
    /// min_periods, if Some and > 0, causes all ratio methods to return None
    /// until at least that many samples have been added.
    pub fn new(
        periodicity: Periodicity,
        annual_risk_free_rate: f64,
        annual_target_return: f64,
        day_count_convention: DayCountConvention,
        rolling_window: Option<usize>,
        min_periods: Option<usize>,
    ) -> Self {
        let ppa = periodicity.periods_per_annum();
        let dpp = periodicity.days_per_period();

        let rfr = if annual_risk_free_rate != 0.0 && ppa != 1 {
            (1.0 + annual_risk_free_rate).powf(1.0 / ppa as f64) - 1.0
        } else {
            annual_risk_free_rate
        };

        let rr = if annual_target_return != 0.0 && ppa != 1 {
            (1.0 + annual_target_return).powf(1.0 / ppa as f64) - 1.0
        } else {
            annual_target_return
        };

        // Treat None or <=0 min_periods as no minimum
        let mp = match min_periods {
            Some(v) if v > 0 => Some(v),
            _ => None,
        };

        let mut r = Self {
            periodicity,
            periods_per_annum: ppa,
            days_per_period: dpp,
            risk_free_rate: rfr,
            required_return: rr,
            day_count_convention,
            rolling_window,
            min_periods: mp,
            fractional_periods: Vec::new(),
            returns: Vec::new(),
            sample_count: 0,
            logret_sum: 0.0,
            drawdowns_cumulative: Vec::new(),
            drawdowns_cumulative_min: f64::INFINITY,
            drawdowns_peaks: Vec::new(),
            drawdowns_peaks_peak: 0,
            drawdown_continuous: Vec::new(),
            drawdown_continuous_final: Vec::new(),
            drawdown_continuous_finalized: false,
            drawdown_continuous_peak: 1,
            drawdown_continuous_inside: false,
            cumulative_return_plus1: 1.0,
            cumulative_return_plus1_max: f64::NEG_INFINITY,
            cumulative_return_geometric_mean: None,
            returns_mean: None,
            returns_std: None,
            returns_autocorr_penalty: 1.0,
            excess_mean: None,
            excess_std: None,
            excess_autocorr_penalty: 1.0,
            required_mean: None,
            required_lpm1: None,
            required_lpm2: None,
            required_lpm3: None,
            required_hpm1: None,
            required_hpm2: None,
            required_hpm3: None,
            required_autocorr_penalty: 1.0,
            avg_return: None,
            avg_win: None,
            avg_loss: None,
            win_rate: None,
            total_duration: 0.0,
            reset_called: false,
        };
        r.reset();
        r
    }

    /// Initializes/resets all internal state for accumulation.
    pub fn reset(&mut self) {
        self.fractional_periods = Vec::new();
        self.returns = Vec::new();
        self.sample_count = 0;
        self.logret_sum = 0.0;
        self.drawdowns_cumulative = Vec::new();
        self.drawdowns_cumulative_min = f64::INFINITY;
        self.drawdowns_peaks = Vec::new();
        self.drawdowns_peaks_peak = 0;
        self.drawdown_continuous = Vec::new();
        self.drawdown_continuous_final = Vec::new();
        self.drawdown_continuous_finalized = false;
        self.drawdown_continuous_peak = 1;
        self.drawdown_continuous_inside = false;
        self.cumulative_return_plus1 = 1.0;
        self.cumulative_return_plus1_max = f64::NEG_INFINITY;
        self.total_duration = 0.0;
        self.cumulative_return_geometric_mean = None;
        self.returns_mean = None;
        self.returns_std = None;
        self.returns_autocorr_penalty = 1.0;
        self.excess_mean = None;
        self.excess_std = None;
        self.excess_autocorr_penalty = 1.0;
        self.required_mean = None;
        self.required_lpm1 = None;
        self.required_lpm2 = None;
        self.required_lpm3 = None;
        self.required_hpm1 = None;
        self.required_hpm2 = None;
        self.required_hpm3 = None;
        self.required_autocorr_penalty = 1.0;
        self.avg_return = None;
        self.avg_win = None;
        self.avg_loss = None;
        self.win_rate = None;
        self.reset_called = true;
    }

    /// Adds a new return observation and updates all internal state.
    pub fn add_return(
        &mut self,
        return_val: f64,
        _return_benchmark: f64,
        _value: f64,
        time_start: &DateTime,
        time_end: &DateTime,
    ) {
        let fractional_period = if self.periodicity == Periodicity::Annual {
            match fractional::year_frac(time_start, time_end, self.day_count_convention) {
                Ok(fp) => fp,
                Err(_) => return,
            }
        } else {
            match fractional::day_frac(time_start, time_end, self.day_count_convention) {
                Ok(fp) => fp / self.days_per_period,
                Err(_) => return,
            }
        };

        self.fractional_periods.push(fractional_period);
        if fractional_period == 0.0 {
            return;
        }
        self.total_duration += fractional_period;
        self.sample_count += 1;

        // Normalized return
        let ret = return_val / fractional_period;
        self.returns.push(ret);

        // Window slice: use last rolling_window returns, or all if not set
        let all_len = self.returns.len();
        let w_start = match self.rolling_window {
            Some(n) if all_len > n => all_len - n,
            _ => 0,
        };
        let w = &self.returns[w_start..];
        let l = w.len();
        let lf = l as f64;

        // Returns mean
        let mean = slice_mean(w);
        self.returns_mean = Some(mean);

        // Returns std (ddof=1, sample)
        if l > 1 {
            self.returns_std = Some(slice_std_ddof1(w, mean));
        } else {
            self.returns_std = None;
        }

        self.returns_autocorr_penalty = autocorr_penalty_fn(w);

        // Average return, win rate, avg win, avg loss
        let non_zero = filter_non_zero(w);
        let len_non_zero = non_zero.len();
        if len_non_zero > 0 {
            self.avg_return = Some(slice_mean(&non_zero));

            let positive = filter_positive(w);
            let len_pos = positive.len();
            self.win_rate = Some(len_pos as f64 / len_non_zero as f64);

            self.avg_win = if len_pos > 0 {
                Some(slice_mean(&positive))
            } else {
                None
            };

            let negative = filter_negative(w);
            self.avg_loss = if !negative.is_empty() {
                Some(slice_mean(&negative))
            } else {
                None
            };
        } else {
            self.avg_return = None;
            self.win_rate = None;
            self.avg_win = None;
            self.avg_loss = None;
        }

        // Excess returns (returns less risk-free rate)
        if self.risk_free_rate == 0.0 {
            self.excess_mean = self.returns_mean;
            self.excess_std = self.returns_std;
            self.excess_autocorr_penalty = self.returns_autocorr_penalty;
        } else {
            let excess: Vec<f64> = w.iter().map(|v| v - self.risk_free_rate).collect();
            let em = slice_mean(&excess);
            self.excess_mean = Some(em);
            if l > 1 {
                self.excess_std = Some(slice_std_ddof1(&excess, em));
            } else {
                self.excess_std = None;
            }
            self.excess_autocorr_penalty = autocorr_penalty_fn(&excess);
        }

        // Lower partial moments for the raw returns (less required return)
        let mut tmp2: Vec<f64> = if self.required_return == 0.0 {
            w.iter().map(|v| -v).collect()
        } else {
            w.iter().map(|v| self.required_return - v).collect()
        };
        // Clip to min 0
        for v in tmp2.iter_mut() {
            if *v < 0.0 {
                *v = 0.0;
            }
        }
        self.required_lpm1 = Some(slice_sum(&tmp2) / lf);
        self.required_lpm2 = Some(slice_sum_pow(&tmp2, 2.0) / lf);
        self.required_lpm3 = Some(slice_sum_pow(&tmp2, 3.0) / lf);

        // Higher partial moments for the raw returns (less required return)
        let mut tmp3: Vec<f64>;
        if self.required_return == 0.0 {
            tmp3 = w.to_vec();
            self.required_mean = self.returns_mean;
            self.required_autocorr_penalty = self.returns_autocorr_penalty;
        } else {
            tmp3 = w.iter().map(|v| v - self.required_return).collect();
            let rm = slice_mean(&tmp3);
            self.required_mean = Some(rm);
            self.required_autocorr_penalty = autocorr_penalty_fn(&tmp3);
        }
        // Clip to min 0
        for v in tmp3.iter_mut() {
            if *v < 0.0 {
                *v = 0.0;
            }
        }
        self.required_hpm1 = Some(slice_sum(&tmp3) / lf);
        self.required_hpm2 = Some(slice_sum_pow(&tmp3, 2.0) / lf);
        self.required_hpm3 = Some(slice_sum_pow(&tmp3, 3.0) / lf);

        // Cumulative returns — recompute from window
        let mut logret_sum_val = 0.0;
        for j in 0..l {
            let fp_j = self.fractional_periods[w_start + j];
            if fp_j != 0.0 {
                logret_sum_val += (w[j] + 1.0).ln();
            }
        }
        self.logret_sum = logret_sum_val;
        let cmr = logret_sum_val.exp();
        self.cumulative_return_plus1 = cmr;
        if l >= 1 {
            self.cumulative_return_geometric_mean = Some(cmr.powf(1.0 / lf) - 1.0);
        }
        self.cumulative_return_plus1_max = f64::NEG_INFINITY;
        let mut cumr = 1.0;
        for j in 0..l {
            cumr *= w[j] + 1.0;
            if cumr > self.cumulative_return_plus1_max {
                self.cumulative_return_plus1_max = cumr;
            }
        }

        // Drawdowns from peaks to valleys (cumulative returns) — recompute from window
        self.drawdowns_cumulative.clear();
        self.drawdowns_cumulative_min = f64::INFINITY;
        cumr = 1.0;
        let mut cumr_max = f64::NEG_INFINITY;
        for j in 0..l {
            cumr *= w[j] + 1.0;
            if cumr > cumr_max {
                cumr_max = cumr;
            }
            let dd = cumr / cumr_max - 1.0;
            self.drawdowns_cumulative.push(dd);
            if self.drawdowns_cumulative_min > dd {
                self.drawdowns_cumulative_min = dd;
            }
        }

        // Drawdown peaks (used in pain index, ulcer index) — recompute from window
        self.drawdowns_peaks.clear();
        self.drawdowns_peaks_peak = 0;
        for j in 0..l {
            let mut dd_peak = 1.0;
            for k in (self.drawdowns_peaks_peak + 1)..=j {
                dd_peak *= 1.0 + w[k] * 0.01;
            }
            if dd_peak > 1.0 {
                self.drawdowns_peaks_peak = j;
                self.drawdowns_peaks.push(0.0);
            } else {
                self.drawdowns_peaks.push((dd_peak - 1.0) * 100.0);
            }
        }

        // Drawdown continuous (used in Burke ratio) — recompute from window
        self.drawdown_continuous.clear();
        self.drawdown_continuous_final.clear();
        self.drawdown_continuous_finalized = false;
        self.drawdown_continuous_peak = 1;
        self.drawdown_continuous_inside = false;
        for j in 1..l {
            if w[j] < 0.0 {
                if !self.drawdown_continuous_inside {
                    self.drawdown_continuous_inside = true;
                    self.drawdown_continuous_peak = j - 1;
                }
                self.drawdown_continuous.push(0.0);
            } else {
                if self.drawdown_continuous_inside {
                    let mut dd_c = 1.0;
                    let j1 = self.drawdown_continuous_peak + 1;
                    for k in j1..j {
                        dd_c *= 1.0 + w[k] * 0.01;
                    }
                    self.drawdown_continuous.push((dd_c - 1.0) * 100.0);
                    self.drawdown_continuous_inside = false;
                } else {
                    self.drawdown_continuous.push(0.0);
                }
            }
        }
    }

    fn is_primed(&self) -> bool {
        match self.min_periods {
            Some(mp) => self.sample_count >= mp,
            None => true,
        }
    }

    fn window_returns(&self) -> &[f64] {
        let all_len = self.returns.len();
        let start = match self.rolling_window {
            Some(n) if all_len > n => all_len - n,
            _ => 0,
        };
        &self.returns[start..]
    }

    fn finalize_continuous_drawdown(&mut self) {
        if self.drawdown_continuous_finalized {
            return;
        }
        if self.drawdown_continuous_inside {
            let w = self.window_returns();
            let mut dd_c = 1.0;
            let j1 = self.drawdown_continuous_peak + 1;
            for j in j1..w.len() {
                dd_c *= 1.0 + w[j] * 0.01;
            }
            let mut final_vec = self.drawdown_continuous.clone();
            final_vec.push((dd_c - 1.0) * 100.0);
            self.drawdown_continuous_final = final_vec;
        } else {
            let mut final_vec = self.drawdown_continuous.clone();
            final_vec.push(0.0);
            self.drawdown_continuous_final = final_vec;
        }
        self.drawdown_continuous_finalized = true;
    }

    /// Returns the cumulative geometric return.
    pub fn cumulative_return(&self) -> f64 {
        self.cumulative_return_plus1 - 1.0
    }

    /// Returns the drawdowns from peaks to valleys on cumulative geometric returns.
    pub fn drawdowns_cumulative(&self) -> &[f64] {
        &self.drawdowns_cumulative
    }

    /// Returns the minimum (most negative) cumulative drawdown.
    pub fn min_drawdowns_cumulative(&self) -> f64 {
        self.drawdowns_cumulative_min
    }

    /// Returns the absolute value of the worst cumulative drawdown.
    pub fn worst_drawdowns_cumulative(&self) -> f64 {
        self.drawdowns_cumulative_min.abs()
    }

    /// Returns the drawdowns from peaks (used in pain/ulcer indices).
    pub fn drawdowns_peaks(&self) -> &[f64] {
        &self.drawdowns_peaks
    }

    /// Returns drawdowns on continuous uninterrupted losing regions.
    pub fn drawdowns_continuous(&mut self, peaks_only: bool, max_peaks: usize) -> Vec<f64> {
        self.finalize_continuous_drawdown();
        if !peaks_only {
            return self.drawdown_continuous_final.clone();
        }
        let mut drawdowns: Vec<f64> = self.drawdown_continuous_final.iter()
            .filter(|&&v| v != 0.0)
            .copied()
            .collect();
        if max_peaks > 0 && !drawdowns.is_empty() {
            drawdowns.sort_by(|a, b| a.partial_cmp(b).unwrap());
            if drawdowns.len() > max_peaks {
                drawdowns.truncate(max_peaks);
            }
        }
        drawdowns
    }

    /// Returns the population skewness of the returns.
    pub fn skew(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let w = self.window_returns();
        if w.len() < 2 {
            return None;
        }
        Some(population_skewness(w))
    }

    /// Returns the population excess kurtosis of the returns.
    pub fn kurtosis(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let w = self.window_returns();
        if w.len() < 2 {
            return None;
        }
        Some(population_excess_kurtosis(w))
    }

    /// Calculates the ex-post Sharpe ratio.
    pub fn sharpe_ratio(&self, ignore_risk_free_rate: bool, autocorrelation_penalty: bool) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        if ignore_risk_free_rate {
            let mean = self.returns_mean?;
            let std = self.returns_std?;
            if std == 0.0 {
                return None;
            }
            let mut denom = std;
            if autocorrelation_penalty {
                denom *= self.returns_autocorr_penalty;
            }
            return Some(mean / denom);
        }
        let mean = self.excess_mean?;
        let std = self.excess_std?;
        if std == 0.0 {
            return None;
        }
        let mut denom = std;
        if autocorrelation_penalty {
            denom *= self.excess_autocorr_penalty;
        }
        Some(mean / denom)
    }

    /// Calculates the Sortino ratio.
    pub fn sortino_ratio(&self, autocorrelation_penalty: bool, divide_by_sqrt2: bool) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let mean = self.required_mean?;
        let lpm2 = self.required_lpm2?;
        if lpm2 == 0.0 {
            return None;
        }
        let mut denom = lpm2.sqrt();
        if autocorrelation_penalty {
            denom *= self.required_autocorr_penalty;
        }
        if divide_by_sqrt2 {
            denom *= SQRT2;
        }
        Some(mean / denom)
    }

    /// Calculates the Omega ratio.
    pub fn omega_ratio(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let mean = self.required_mean?;
        let lpm1 = self.required_lpm1?;
        if lpm1 == 0.0 {
            return None;
        }
        Some(mean / lpm1 + 1.0)
    }

    /// Calculates the Kappa ratio of a given order.
    pub fn kappa_ratio(&self, order: i32) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let mean = self.required_mean?;
        match order {
            1 => {
                let lpm1 = self.required_lpm1?;
                if lpm1 == 0.0 { return None; }
                Some(mean / lpm1)
            }
            2 => {
                let lpm2 = self.required_lpm2?;
                if lpm2 == 0.0 { return None; }
                Some(mean / lpm2.sqrt())
            }
            3 => {
                let lpm3 = self.required_lpm3?;
                if lpm3 == 0.0 { return None; }
                Some(mean / lpm3.cbrt())
            }
            _ => {
                let w = self.window_returns();
                let l = w.len();
                if l == 0 {
                    return None;
                }
                let mut tmp: Vec<f64> = if self.required_return == 0.0 {
                    w.iter().map(|v| -v).collect()
                } else {
                    w.iter().map(|v| self.required_return - v).collect()
                };
                for v in tmp.iter_mut() {
                    if *v < 0.0 {
                        *v = 0.0;
                    }
                }
                let lpm = slice_sum_pow(&tmp, order as f64) / l as f64;
                if lpm == 0.0 {
                    return None;
                }
                Some(mean / lpm.powf(1.0 / order as f64))
            }
        }
    }

    /// Calculates the Kappa ratio of order 3.
    pub fn kappa3_ratio(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let mean = self.required_mean?;
        let lpm3 = self.required_lpm3?;
        if lpm3 == 0.0 {
            return None;
        }
        Some(mean / lpm3.cbrt())
    }

    /// Calculates the Bernardo-Ledoit ratio.
    pub fn bernardo_ledoit_ratio(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let w = self.window_returns();
        let l = w.len();
        if l < 1 {
            return None;
        }
        let lf = l as f64;

        // LPM1 with threshold=0
        let tmp: Vec<f64> = w.iter().map(|v| (-v).max(0.0)).collect();
        let lpm1 = slice_sum(&tmp) / lf;
        if lpm1 == 0.0 {
            return None;
        }

        // HPM1 with threshold=0
        let tmp2: Vec<f64> = w.iter().map(|v| v.max(0.0)).collect();
        let hpm1 = slice_sum(&tmp2) / lf;
        Some(hpm1 / lpm1)
    }

    /// Calculates the upside potential ratio.
    pub fn upside_potential_ratio(&self, full: bool) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        if full {
            let hpm1 = self.required_hpm1?;
            let lpm2 = self.required_lpm2?;
            if lpm2 == 0.0 {
                return None;
            }
            return Some(hpm1 / lpm2.sqrt());
        }
        // Subset version
        let w = self.window_returns();
        let below: Vec<f64> = w.iter()
            .filter(|&&v| v < self.required_return)
            .copied()
            .collect();
        let l = below.len();
        if l < 1 {
            return None;
        }
        let lf = l as f64;
        let tmp: Vec<f64> = below.iter().map(|v| v - self.required_return).collect();
        let lpm2 = slice_sum_pow(&tmp, 2.0) / lf;
        if lpm2 == 0.0 {
            return None;
        }
        let above: Vec<f64> = w.iter()
            .filter(|&&v| v > self.required_return)
            .map(|v| v - self.required_return)
            .collect();
        if above.is_empty() {
            return None;
        }
        let hpm1 = slice_mean(&above);
        Some(hpm1 / lpm2.sqrt())
    }

    /// Returns the compound (annual) growth rate (CAGR).
    pub fn compound_growth_rate(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        self.cumulative_return_geometric_mean
    }

    /// Calculates the Calmar ratio.
    pub fn calmar_ratio(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let wdd = self.worst_drawdowns_cumulative();
        if wdd == 0.0 {
            return None;
        }
        let gm = self.cumulative_return_geometric_mean?;
        Some(gm / wdd)
    }

    /// Calculates the Sterling ratio with the given annual excess rate.
    pub fn sterling_ratio(&self, annual_excess_rate: f64) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let excess_rate = if annual_excess_rate != 0.0 && self.periods_per_annum != 1 {
            (1.0 + annual_excess_rate).powf(1.0 / self.periods_per_annum as f64) - 1.0
        } else {
            annual_excess_rate
        };
        let wdd = self.worst_drawdowns_cumulative() + excess_rate;
        if wdd == 0.0 {
            return None;
        }
        let gm = self.cumulative_return_geometric_mean?;
        Some(gm / wdd)
    }

    /// Calculates the Burke ratio.
    pub fn burke_ratio(&mut self, modified: bool) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let gm = self.cumulative_return_geometric_mean?;
        let rate = gm - self.risk_free_rate;
        let drawdowns = self.drawdowns_continuous(true, 0);
        if drawdowns.is_empty() {
            return None;
        }
        let sum_sq: f64 = drawdowns.iter().map(|d| d * d).sum();
        let sqrt_sum_sq = sum_sq.sqrt();
        if sqrt_sum_sq == 0.0 {
            return None;
        }
        let mut burke = rate / sqrt_sum_sq;
        if modified {
            burke *= (self.window_returns().len() as f64).sqrt();
        }
        Some(burke)
    }

    /// Calculates the pain index.
    pub fn pain_index(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let l = self.drawdowns_peaks.len();
        if l < 1 {
            return None;
        }
        Some(-slice_sum(&self.drawdowns_peaks) / l as f64)
    }

    /// Calculates the pain ratio.
    pub fn pain_ratio(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let gm = self.cumulative_return_geometric_mean?;
        let rate = gm - self.risk_free_rate;
        let l = self.drawdowns_peaks.len();
        if l < 1 {
            return None;
        }
        let pain_index = -slice_sum(&self.drawdowns_peaks) / l as f64;
        if pain_index == 0.0 {
            return None;
        }
        Some(rate / pain_index)
    }

    /// Calculates the ulcer index.
    pub fn ulcer_index(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let l = self.drawdowns_peaks.len();
        if l < 1 {
            return None;
        }
        let sum_sq: f64 = self.drawdowns_peaks.iter().map(|d| d * d).sum();
        Some((sum_sq / l as f64).sqrt())
    }

    /// Calculates the Martin (Ulcer) ratio.
    pub fn martin_ratio(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let gm = self.cumulative_return_geometric_mean?;
        let rate = gm - self.risk_free_rate;
        let l = self.drawdowns_peaks.len();
        if l < 1 {
            return None;
        }
        let sum_sq: f64 = self.drawdowns_peaks.iter().map(|d| d * d).sum();
        let ulcer_index = (sum_sq / l as f64).sqrt();
        if ulcer_index == 0.0 {
            return None;
        }
        Some(rate / ulcer_index)
    }

    /// Returns Jack Schwager's gain-to-pain ratio.
    pub fn gain_to_pain_ratio(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let lpm1 = self.required_lpm1?;
        if lpm1 == 0.0 {
            return None;
        }
        let mean = self.returns_mean?;
        Some(mean / lpm1)
    }

    /// Calculates the risk of ruin.
    pub fn risk_of_ruin(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let wr = self.win_rate?;
        Some(((1.0 - wr) / (1.0 + wr)).powf(self.window_returns().len() as f64))
    }

    /// Calculates the return/risk ratio.
    pub fn risk_return_ratio(&self) -> Option<f64> {
        if !self.is_primed() {
            return None;
        }
        let mean = self.returns_mean?;
        let std = self.returns_std?;
        if std == 0.0 {
            return None;
        }
        Some(mean / std)
    }
}

// ---------- helper functions ----------

fn autocorr_penalty_fn(_returns: &[f64]) -> f64 {
    1.0
}

fn slice_sum(s: &[f64]) -> f64 {
    s.iter().sum()
}

fn slice_mean(s: &[f64]) -> f64 {
    if s.is_empty() {
        return 0.0;
    }
    slice_sum(s) / s.len() as f64
}

fn slice_std_ddof1(s: &[f64], mean: f64) -> f64 {
    if s.len() < 2 {
        return 0.0;
    }
    let sum: f64 = s.iter().map(|v| {
        let d = v - mean;
        d * d
    }).sum();
    (sum / (s.len() - 1) as f64).sqrt()
}

fn slice_sum_pow(s: &[f64], power: f64) -> f64 {
    s.iter().map(|v| v.powf(power)).sum()
}

fn filter_non_zero(s: &[f64]) -> Vec<f64> {
    s.iter().filter(|&&v| v != 0.0).copied().collect()
}

fn filter_positive(s: &[f64]) -> Vec<f64> {
    s.iter().filter(|&&v| v > 0.0).copied().collect()
}

fn filter_negative(s: &[f64]) -> Vec<f64> {
    s.iter().filter(|&&v| v < 0.0).copied().collect()
}

fn population_skewness(s: &[f64]) -> f64 {
    let n = s.len() as f64;
    let mean = slice_mean(s);
    let mut m2 = 0.0;
    let mut m3 = 0.0;
    for v in s {
        let d = v - mean;
        m2 += d * d;
        m3 += d * d * d;
    }
    m2 /= n;
    m3 /= n;
    if m2 == 0.0 {
        return 0.0;
    }
    m3 / m2.powf(1.5)
}

fn population_excess_kurtosis(s: &[f64]) -> f64 {
    let n = s.len() as f64;
    let mean = slice_mean(s);
    let mut m2 = 0.0;
    let mut m4 = 0.0;
    for v in s {
        let d = v - mean;
        let d2 = d * d;
        m2 += d2;
        m4 += d2 * d2;
    }
    m2 /= n;
    m4 /= n;
    if m2 == 0.0 {
        return 0.0;
    }
    m4 / (m2 * m2) - 3.0
}

#[cfg(test)]
mod tests {
    use super::*;

    const EPSILON: f64 = 1e-13;
    const EPSILON_12: f64 = 1e-12;

    fn almost_equal(a: f64, b: f64, eps: f64) -> bool {
        (a - b).abs() < eps
    }

    // Bacon dataset from Go's ratios_test.go (PerformanceAnalytics R package).
    const BACON_PORTFOLIO_RETURNS: [f64; 24] = [
        0.003, 0.026, 0.011, -0.010,
        0.015, 0.025, 0.016, 0.067,
        -0.014, 0.040, -0.005, 0.081,
        0.040, -0.037, -0.061, 0.017,
        -0.049, -0.022, 0.070, 0.058,
        -0.065, 0.024, -0.005, -0.009,
    ];
    const BACON_BENCHMARK_RETURNS: [f64; 24] = [
        0.002, 0.025, 0.018, -0.011,
        0.014, 0.018, 0.014, 0.065,
        -0.015, 0.042, -0.006, 0.083,
        0.039, -0.038, -0.062, 0.015,
        -0.048, 0.021, 0.060, 0.056,
        -0.067, 0.019, -0.003, 0.000,
    ];

    fn bacon_dates() -> Vec<DateTime> {
        vec![
            DateTime::date(2024, 7, 1), DateTime::date(2024, 7, 2),
            DateTime::date(2024, 7, 3), DateTime::date(2024, 7, 4),
            DateTime::date(2024, 7, 5), DateTime::date(2024, 7, 6),
            DateTime::date(2024, 7, 7), DateTime::date(2024, 7, 8),
            DateTime::date(2024, 7, 9), DateTime::date(2024, 7, 10),
            DateTime::date(2024, 7, 11), DateTime::date(2024, 7, 12),
            DateTime::date(2024, 7, 13), DateTime::date(2024, 7, 14),
            DateTime::date(2024, 7, 15), DateTime::date(2024, 7, 16),
            DateTime::date(2024, 7, 17), DateTime::date(2024, 7, 18),
            DateTime::date(2024, 7, 19), DateTime::date(2024, 7, 20),
            DateTime::date(2024, 7, 21), DateTime::date(2024, 7, 22),
            DateTime::date(2024, 7, 23), DateTime::date(2024, 7, 24),
        ]
    }

    fn bacon_dates_previous() -> Vec<DateTime> {
        vec![
            DateTime::date(2024, 6, 30), DateTime::date(2024, 7, 1),
            DateTime::date(2024, 7, 2), DateTime::date(2024, 7, 3),
            DateTime::date(2024, 7, 4), DateTime::date(2024, 7, 5),
            DateTime::date(2024, 7, 6), DateTime::date(2024, 7, 7),
            DateTime::date(2024, 7, 8), DateTime::date(2024, 7, 9),
            DateTime::date(2024, 7, 10), DateTime::date(2024, 7, 11),
            DateTime::date(2024, 7, 12), DateTime::date(2024, 7, 13),
            DateTime::date(2024, 7, 14), DateTime::date(2024, 7, 15),
            DateTime::date(2024, 7, 16), DateTime::date(2024, 7, 17),
            DateTime::date(2024, 7, 18), DateTime::date(2024, 7, 19),
            DateTime::date(2024, 7, 20), DateTime::date(2024, 7, 21),
            DateTime::date(2024, 7, 22), DateTime::date(2024, 7, 23),
        ]
    }

    fn new_ratios_with_rf(rf: f64) -> Ratios {
        let annual_rf = (1.0 + rf).powf(252.0) - 1.0;
        Ratios::new(Periodicity::Daily, annual_rf, 0.0, DayCountConvention::Raw, None, None)
    }

    fn new_ratios_with_mar(mar: f64) -> Ratios {
        let annual_mar = (1.0 + mar).powf(252.0) - 1.0;
        Ratios::new(Periodicity::Daily, 0.0, annual_mar, DayCountConvention::Raw, None, None)
    }

    fn add_bacon_return(r: &mut Ratios, i: usize) {
        let dates = bacon_dates();
        let prev = bacon_dates_previous();
        r.add_return(
            BACON_PORTFOLIO_RETURNS[i],
            BACON_BENCHMARK_RETURNS[i],
            1.0,
            &prev[i],
            &dates[i],
        );
    }

    fn assert_nullable(step: usize, expected: Option<f64>, actual: Option<f64>, eps: f64, label: &str) {
        match expected {
            None => {
                assert!(actual.is_none(), "{} step {}: expected None, got {:?}", label, step, actual);
            }
            Some(exp) => {
                let act = actual.unwrap_or_else(|| panic!("{} step {}: expected {:.16}, got None", label, step, exp));
                assert!(almost_equal(act, exp, eps),
                    "{} step {}: expected {:.16e}, got {:.16e}, diff={:.2e}",
                    label, step, exp, act, (act - exp).abs());
            }
        }
    }

    // ===== Kurtosis tests =====
    #[test]
    fn test_kurtosis() {
        let expected: [Option<f64>; 24] = [
            None, Some(-2.00000000000000000), Some(-1.50000000000000000),
            Some(-1.17592035552795000), Some(-0.94669079980875600), Some(-0.96028723389787100),
            Some(-0.57793300076120100), Some(0.78641242115027200), Some(0.59954237086621500),
            Some(-0.01187577489273160), Some(0.07517391430462480), Some(-0.27406990671095100),
            Some(-0.38022416153835900), Some(-0.31560370425738600), Some(-0.16235155227201600),
            Some(0.02528905226985100), Some(-0.33285099821964000), Some(-0.37425348407483000),
            Some(-0.58502674157514900), Some(-0.69334606360953100), Some(-0.77381631285861200),
            Some(-0.68208349704651200), Some(-0.61779722177118000), Some(-0.56754620589212500),
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.kurtosis(), EPSILON, "Kurtosis");
        }
    }

    // ===== Sharpe Ratio tests =====
    #[test]
    fn test_sharpe_rf0() {
        let expected: [Option<f64>; 24] = [
            None, Some(0.8915694197569510), Some(1.1419253390798400),
            Some(0.4977924836999790), Some(0.6680426571226850), Some(0.8511810078441020),
            Some(0.9735918376312110), Some(0.8462916062735410), Some(0.6475912629068400),
            Some(0.7524743687246650), Some(0.6702597534059590), Some(0.7244562693337180),
            Some(0.7945207458232130), Some(0.5805910371128360), Some(0.3566360956461000),
            Some(0.3758075293232440), Some(0.2578994439571370), Some(0.2131725662300710),
            Some(0.2880753096781920), Some(0.3448210835211740), Some(0.2337747541463060),
            Some(0.2546053055676570), Some(0.2430648040410730), Some(0.2275684556623890),
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.sharpe_ratio(false, false), EPSILON, "Sharpe rf=0");
        }
    }

    #[test]
    fn test_sharpe_rf005() {
        let expected: [Option<f64>; 24] = [
            None, Some(-2.1828078897497800), Some(-3.1402946824695500),
            Some(-2.8208240742998800), Some(-3.0433054380033400), Some(-2.7967375972020500),
            Some(-2.9887005248213900), Some(-1.3662354689514000), Some(-1.4489272141296900),
            Some(-1.3494093427967500), Some(-1.4483773981646000), Some(-0.9801467173338540),
            Some(-0.9561181856516630), Some(-0.9946559628057110), Some(-1.0011155375243300),
            Some(-1.0290804307636500), Some(-1.0706734491553900), Some(-1.1284729554976500),
            Some(-0.9967676208113950), Some(-0.9275814386971810), Some(-0.9577955946576800),
            Some(-0.9630722427994000), Some(-0.9992664166133000), Some(-1.0367007424619900),
        ];
        let mut r = new_ratios_with_rf(0.05);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.sharpe_ratio(false, false), EPSILON, "Sharpe rf=0.05");
        }
    }

    #[test]
    fn test_sharpe_rf010() {
        let expected: [Option<f64>; 24] = [
            None, Some(-5.257185199256510), Some(-7.422514704018940),
            Some(-6.139440632299740), Some(-6.754653533129370), Some(-6.444656202248210),
            Some(-6.950992887274000), Some(-3.578762544176350), Some(-3.545445691166220),
            Some(-3.451293054318160), Some(-3.567014549735160), Some(-2.684749704001430),
            Some(-2.706757117126540), Some(-2.569902962724260), Some(-2.358867170694770),
            Some(-2.433968390850540), Some(-2.399246342267910), Some(-2.470118477225370),
            Some(-2.281610551300980), Some(-2.199983960915530), Some(-2.149365943461670),
            Some(-2.180749791166460), Some(-2.241597637267670), Some(-2.300969940586380),
        ];
        let mut r = new_ratios_with_rf(0.10);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.sharpe_ratio(false, false), EPSILON, "Sharpe rf=0.10");
        }
    }

    #[test]
    fn test_sharpe_ignore_rf() {
        // When ignoring RF, all RF values should produce the same Sharpe as RF=0.
        let expected_rf0: [Option<f64>; 24] = [
            None, Some(0.8915694197569510), Some(1.1419253390798400),
            Some(0.4977924836999790), Some(0.6680426571226850), Some(0.8511810078441020),
            Some(0.9735918376312110), Some(0.8462916062735410), Some(0.6475912629068400),
            Some(0.7524743687246650), Some(0.6702597534059590), Some(0.7244562693337180),
            Some(0.7945207458232130), Some(0.5805910371128360), Some(0.3566360956461000),
            Some(0.3758075293232440), Some(0.2578994439571370), Some(0.2131725662300710),
            Some(0.2880753096781920), Some(0.3448210835211740), Some(0.2337747541463060),
            Some(0.2546053055676570), Some(0.2430648040410730), Some(0.2275684556623890),
        ];
        for &rf in &[0.0, 0.05, 0.10] {
            let mut r = new_ratios_with_rf(rf);
            for i in 0..24 {
                add_bacon_return(&mut r, i);
                assert_nullable(i, expected_rf0[i], r.sharpe_ratio(true, false), EPSILON, "Sharpe ignore_rf");
            }
        }
    }

    // ===== Sortino Ratio tests =====
    #[test]
    fn test_sortino_mar0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(1.5), Some(2.01246117974981), Some(2.85773803324704),
            Some(3.25049446787935), Some(5.40936687607709), Some(2.69307029756515),
            Some(3.29008543386979), Some(2.92819766175444), Some(4.10863007844407),
            Some(4.56665101160337), Some(1.67730613630736), Some(0.691483512929973),
            Some(0.727302390567925), Some(0.452770753672167), Some(0.370054264368203),
            Some(0.536498400203865), Some(0.665303673385798), Some(0.401733515514418),
            Some(0.438224836666163), Some(0.418857174247308), Some(0.392372028795065),
        ];
        let mut r = new_ratios_with_mar(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.sortino_ratio(false, false), EPSILON, "Sortino mar=0");
        }
    }

    #[test]
    fn test_sortino_mar005() {
        let expected: [Option<f64>; 24] = [
            Some(-1.0), Some(-0.951329033501053), Some(-0.967821008377905),
            Some(-0.955961761235827), Some(-0.959422032420532), Some(-0.950640505399932),
            Some(-0.95521850710367), Some(-0.835987494907806), Some(-0.84620916319764),
            Some(-0.825850705880606), Some(-0.841892559996059), Some(-0.739594446201381),
            Some(-0.729168016460068), Some(-0.735413445987151), Some(-0.731283824494091),
            Some(-0.739823509257131), Some(-0.750430484501361), Some(-0.766429130761335),
            Some(-0.726278292165206), Some(-0.700204514919608), Some(-0.709303305303401),
            Some(-0.71078905810419), Some(-0.723223919287678), Some(-0.735374254070636),
        ];
        let mut r = new_ratios_with_mar(0.05);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.sortino_ratio(false, false), EPSILON, "Sortino mar=0.05");
        }
    }

    #[test]
    fn test_sortino_mar010() {
        let expected: [Option<f64>; 24] = [
            Some(-1.0), Some(-0.991075392350217), Some(-0.994004065367307),
            Some(-0.990197182430257), Some(-0.991346575643354), Some(-0.990116442284817),
            Some(-0.991246179848335), Some(-0.967496714088971), Some(-0.966414074414246),
            Some(-0.964235550350565), Some(-0.966082469415414), Some(-0.94189872360901),
            Some(-0.942394085487388), Some(-0.936339897881547), Some(-0.925395562084343),
            Some(-0.929178181525619), Some(-0.927078590338157), Some(-0.930569106181964),
            Some(-0.919801251226696), Some(-0.914287704738154), Some(-0.910539461748751),
            Some(-0.912598194566486), Some(-0.916559222172636), Some(-0.920182172840589),
        ];
        let mut r = new_ratios_with_mar(0.10);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.sortino_ratio(false, false), EPSILON, "Sortino mar=0.10");
        }
    }

    #[test]
    fn test_sortino_sqrt2() {
        // Sortino with divide_by_sqrt2 should give normal / sqrt(2).
        for &mar in &[0.0, 0.05, 0.10] {
            let mut r = new_ratios_with_mar(mar);
            for i in 0..24 {
                add_bacon_return(&mut r, i);
                let normal = r.sortino_ratio(false, false);
                let divided = r.sortino_ratio(false, true);
                match normal {
                    None => assert!(divided.is_none()),
                    Some(n) => {
                        let d = divided.unwrap();
                        assert!(almost_equal(d, n / SQRT2, EPSILON),
                            "mar={} step {}: sqrt2 mismatch", mar, i);
                    }
                }
            }
        }
    }

    // ===== Omega Ratio tests =====
    #[test]
    fn test_omega_threshold0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(4.000000000000000), Some(5.500000000000000), Some(8.000000000000000),
            Some(9.600000000000000), Some(16.300000000000000), Some(6.791666666666670),
            Some(8.458333333333330), Some(7.000000000000000), Some(9.793103448275860),
            Some(11.172413793103400), Some(4.909090909090910), Some(2.551181102362210),
            Some(2.685039370078740), Some(1.937500000000000), Some(1.722222222222220),
            Some(2.075757575757580), Some(2.368686868686870), Some(1.783269961977190),
            Some(1.874524714828900), Some(1.839552238805970), Some(1.779783393501810),
        ];
        let mut r = new_ratios_with_mar(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.omega_ratio(), EPSILON, "Omega t=0");
        }
    }

    #[test]
    fn test_omega_threshold002() {
        let expected: [Option<f64>; 24] = [
            Some(0.00000000000000000), Some(0.35294117647058800), Some(0.23076923076923100),
            Some(0.10714285714285700), Some(0.09836065573770490), Some(0.18032786885245900),
            Some(0.16923076923076900), Some(0.89230769230769200), Some(0.58585858585858600),
            Some(0.78787878787878800), Some(0.62903225806451600), Some(1.12096774193548000),
            Some(1.28225806451613000), Some(0.87845303867403300), Some(0.60687022900763400),
            Some(0.60000000000000000), Some(0.47604790419161700), Some(0.42287234042553200),
            Some(0.55585106382978700), Some(0.65691489361702100), Some(0.53579175704989200),
            Some(0.54446854663774400), Some(0.51646090534979400), Some(0.48737864077669900),
        ];
        let mut r = new_ratios_with_mar(0.02);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.omega_ratio(), EPSILON, "Omega t=0.02");
        }
    }

    #[test]
    fn test_omega_threshold004() {
        let expected: [Option<f64>; 24] = [
            Some(0.00000000000000000), Some(0.00000000000000000), Some(0.00000000000000000),
            Some(0.00000000000000000), Some(0.00000000000000000), Some(0.00000000000000000),
            Some(0.00000000000000000), Some(0.13917525773195900), Some(0.10887096774193500),
            Some(0.10887096774193500), Some(0.09215017064846420), Some(0.23208191126279900),
            Some(0.23208191126279900), Some(0.18378378378378400), Some(0.14437367303609300),
            Some(0.13765182186234800), Some(0.11663807890223000), Some(0.10542635658914700),
            Some(0.15193798449612400), Some(0.17984496124031000), Some(0.15466666666666700),
            Some(0.15143603133159300), Some(0.14303329223181300), Some(0.13488372093023300),
        ];
        let mut r = new_ratios_with_mar(0.04);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.omega_ratio(), EPSILON, "Omega t=0.04");
        }
    }

    #[test]
    fn test_omega_threshold006() {
        let expected: [Option<f64>; 24] = [
            Some(0.00000000000000000), Some(0.00000000000000000), Some(0.00000000000000000),
            Some(0.00000000000000000), Some(0.00000000000000000), Some(0.00000000000000000),
            Some(0.00000000000000000), Some(0.02095808383233530), Some(0.01715686274509810),
            Some(0.01635514018691590), Some(0.01419878296146050), Some(0.05679513184584180),
            Some(0.05458089668615990), Some(0.04590163934426230), Some(0.03830369357045140),
            Some(0.03617571059431530), Some(0.03171007927519820), Some(0.02901554404145080),
            Some(0.03937823834196890), Some(0.03929679420889350), Some(0.03479853479853480),
            Some(0.03368794326241140), Some(0.03185247275775360), Some(0.03011093502377180),
        ];
        let mut r = new_ratios_with_mar(0.06);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.omega_ratio(), EPSILON, "Omega t=0.06");
        }
    }

    // ===== Kappa Ratio tests =====
    #[test]
    fn test_kappa_order1_mar0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(3.0000000000000000), Some(4.5000000000000000), Some(7.0000000000000000),
            Some(8.6000000000000000), Some(15.300000000000000), Some(5.7916666666666700),
            Some(7.4583333333333300), Some(6.0000000000000000), Some(8.7931034482758600),
            Some(10.172413793103400), Some(3.9090909090909090), Some(1.5511811023622000),
            Some(1.6850393700787400), Some(0.9375000000000000), Some(0.7222222222222220),
            Some(1.0757575757575800), Some(1.3686868686868700), Some(0.7832699619771860),
            Some(0.8745247148288970), Some(0.8395522388059700), Some(0.7797833935018050),
        ];
        let mut r = new_ratios_with_mar(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.kappa_ratio(1), EPSILON, "Kappa1 mar=0");
        }
    }

    #[test]
    fn test_kappa_order2_mar0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(1.5000000000000000), Some(2.0124611797498100), Some(2.8577380332470400),
            Some(3.2504944678793500), Some(5.4093668760770900), Some(2.6930702975651500),
            Some(3.2900854338697900), Some(2.9281976617544400), Some(4.1086300784440700),
            Some(4.5666510116033700), Some(1.6773061363073600), Some(0.6914835129299730),
            Some(0.7273023905679250), Some(0.4527707536721670), Some(0.3700542643682030),
            Some(0.5364984002038650), Some(0.6653036733857980), Some(0.4017335155144180),
            Some(0.4382248366661630), Some(0.4188571742473080), Some(0.3923720287950650),
        ];
        let mut r = new_ratios_with_mar(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.kappa_ratio(2), EPSILON, "Kappa2 mar=0");
        }
    }

    #[test]
    fn test_kappa_order3_mar0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(1.1905507889761500), Some(1.5389783520090300), Some(2.1199740249708300),
            Some(2.3501725959775100), Some(3.8250000000000000), Some(2.0689080079822300),
            Some(2.4835586338430600), Some(2.2408934899599800), Some(3.0989871337864400),
            Some(3.3988098734763700), Some(1.1713241279859900), Some(0.4942094486331960),
            Some(0.5142481946830330), Some(0.3389522803724070), Some(0.2803047018509310),
            Some(0.4027354737116720), Some(0.4951749226471700), Some(0.3070994714658920),
            Some(0.3324074590706010), Some(0.3156667962042520), Some(0.2944582876612480),
        ];
        let mut r = new_ratios_with_mar(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.kappa_ratio(3), EPSILON, "Kappa3 mar=0");
        }
    }

    #[test]
    fn test_kappa3_ratio_matches_kappa_order3() {
        let mut r = new_ratios_with_mar(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            let k3 = r.kappa3_ratio();
            let kr3 = r.kappa_ratio(3);
            match k3 {
                None => assert!(kr3.is_none()),
                Some(a) => {
                    let b = kr3.unwrap();
                    assert!(almost_equal(a, b, 1e-15),
                        "step {}: kappa3={:.16}, kappa(3)={:.16}", i, a, b);
                }
            }
        }
    }

    #[test]
    fn test_kappa_order4_mar0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(1.0606601717798200), Some(1.3458139030991000), Some(1.8259320100855000),
            Some(1.9983654900858500), Some(3.2164287883454600), Some(1.8033735333115700),
            Some(2.1458818396425000), Some(1.9358196995813000), Some(2.6577517731212100),
            Some(2.8955073548113600), Some(0.9572325404178820), Some(0.4101535241803780),
            Some(0.4244948756831930), Some(0.2893122764499430), Some(0.2395667425326850),
            Some(0.3426567344926290), Some(0.4195093677307130), Some(0.2646839611869220),
            Some(0.2853879958557920), Some(0.2700286247384100), Some(0.2510732850424660),
        ];
        let mut r = new_ratios_with_mar(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.kappa_ratio(4), EPSILON, "Kappa4 mar=0");
        }
    }

    #[test]
    fn test_kappa_order1_mar005() {
        let expected: [Option<f64>; 24] = [
            Some(-1.0000000000000000), Some(-1.0000000000000000), Some(-1.0000000000000000),
            Some(-1.0000000000000000), Some(-1.0000000000000000), Some(-1.0000000000000000),
            Some(-1.0000000000000000), Some(-0.9356060606060610), Some(-0.9481707317073170),
            Some(-0.9497041420118340), Some(-0.9567430025445290), Some(-0.8778625954198470),
            Some(-0.8808933002481390), Some(-0.9020408163265300), Some(-0.9201331114808650),
            Some(-0.9242902208201890), Some(-0.9345156889495230), Some(-0.9403726708074530),
            Some(-0.9155279503105590), Some(-0.9055900621118010), Some(-0.9173913043478260),
            Some(-0.9196617336152220), Some(-0.9240759240759240), Some(-0.9283018867924530),
        ];
        let mut r = new_ratios_with_mar(0.05);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.kappa_ratio(1), EPSILON, "Kappa1 mar=0.05");
        }
    }

    // ===== Bernardo-Ledoit Ratio tests =====
    #[test]
    fn test_bernardo_ledoit() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(4.000000000000000), Some(5.500000000000000), Some(8.000000000000000),
            Some(9.600000000000000), Some(16.30000000000000), Some(6.791666666666670),
            Some(8.458333333333330), Some(7.000000000000000), Some(9.793103448275860),
            Some(11.17241379310340), Some(4.909090909090910), Some(2.551181102362200),
            Some(2.685039370078740), Some(1.937500000000000), Some(1.722222222222220),
            Some(2.075757575757580), Some(2.368686868686870), Some(1.783269961977190),
            Some(1.874524714828900), Some(1.839552238805970), Some(1.779783393501800),
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.bernardo_ledoit_ratio(), EPSILON, "BernardoLedoit");
        }
    }

    // ===== Upside Potential Ratio tests =====
    #[test]
    fn test_upside_potential_full_mar0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(2.0000000000000000), Some(2.4596747752497700), Some(3.2659863237109000),
            Some(3.6284589408885800), Some(5.7629202666703600), Some(3.1580608525404200),
            Some(3.7312142071260700), Some(3.4162306053801800), Some(4.5758860481494800),
            Some(5.0155760263033600), Some(2.1063844502464600), Some(1.1372622243112200),
            Some(1.1589257718862700), Some(0.9357262242558120), Some(0.8824370919549460),
            Some(1.0352152229285800), Some(1.1513927041252400), Some(0.9146263047391370),
            Some(0.9393254107670360), Some(0.9177626084618780), Some(0.8955528249813290),
        ];
        let mut r = new_ratios_with_mar(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.upside_potential_ratio(true), EPSILON, "UpsidePotential");
        }
    }

    // ===== Cumulative Return tests =====
    #[test]
    fn test_cumulative_return() {
        let expected: [f64; 24] = [
            0.00299999999999989, 0.02907799999999990, 0.04039785799999970,
            0.02999387941999990, 0.04544378761129960, 0.07157988230158210,
            0.08872516041840740, 0.16166974616644100, 0.14540636972011000,
            0.19122262450891500, 0.18526651138637000, 0.28127309880866600,
            0.33252402276101300, 0.28322063391885500, 0.20494417524980500,
            0.22542822622905200, 0.16538224314382800, 0.13974383379466400,
            0.21952590216029100, 0.29025840448558800, 0.20639160819402400,
            0.23534500679068100, 0.22916828175672800, 0.21810576722091700,
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            let cr = r.cumulative_return();
            assert!(almost_equal(cr, expected[i], EPSILON),
                "CumulativeReturn[{}]: got {:.16}, want {:.16}, diff={:.2e}",
                i, cr, expected[i], (cr - expected[i]).abs());
        }
    }

    // ===== Drawdowns Cumulative tests =====
    #[test]
    fn test_drawdowns_cumulative() {
        let expected: [f64; 24] = [
            0.000000000000000000, 0.000000000000000000, 0.000000000000000000,
            -0.009999999999999900, 0.000000000000000000, 0.000000000000000000,
            0.000000000000000000, 0.000000000000000000, -0.014000000000000000,
            0.000000000000000000, -0.005000000000000120, 0.000000000000000000,
            0.000000000000000000, -0.037000000000000100, -0.095743000000000000,
            -0.080370631000000200, -0.125432470081000000, -0.144672955739218000,
            -0.084800062640963400, -0.031718466274139200, -0.094656765966320100,
            -0.072928528349511800, -0.077563885707764200, -0.085865810736394400,
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
        }
        let dd = r.drawdowns_cumulative();
        assert_eq!(dd.len(), 24);
        for i in 0..24 {
            assert!(almost_equal(dd[i], expected[i], EPSILON),
                "Drawdown[{}]: got {:.18}, want {:.18}", i, dd[i], expected[i]);
        }
    }

    #[test]
    fn test_worst_drawdown() {
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
        }
        let expected_worst = 0.1446729557392180;
        let wdd = r.worst_drawdowns_cumulative();
        assert!(almost_equal(wdd, expected_worst, EPSILON),
            "WorstDrawdown: got {:.16}, want {:.16}", wdd, expected_worst);
    }

    // ===== Drawdown Peaks tests =====
    #[test]
    fn test_drawdowns_peaks() {
        let expected: [f64; 24] = [
            0.00000000000000000, 0.00000000000000000, 0.00000000000000000,
            -0.00999999999999890, 0.00000000000000000, 0.00000000000000000,
            0.00000000000000000, 0.00000000000000000, -0.01400000000000290,
            0.00000000000000000, -0.00499999999999945, 0.00000000000000000,
            0.00000000000000000, -0.03699999999999810, -0.09797742999999580,
            -0.08099408616309980, -0.12995439906088300, -0.15192580909309000,
            -0.08203215715946180, -0.02407973581061150, -0.08906408398233760,
            -0.06508545936249050, -0.07008220508951670, -0.07907589769106100,
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
        }
        let peaks = r.drawdowns_peaks();
        assert_eq!(peaks.len(), 24);
        for i in 0..24 {
            assert!(almost_equal(peaks[i], expected[i], EPSILON),
                "DrawdownPeaks[{}]: got {:.17}, want {:.17}, diff={:.2e}",
                i, peaks[i], expected[i], (peaks[i] - expected[i]).abs());
        }
    }

    // ===== Calmar Ratio tests =====
    #[test]
    fn test_calmar() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(0.74155751780918500), Some(0.89279126631479400), Some(1.15889854414036000),
            Some(1.22179559465027000), Some(1.89088510302246000), Some(1.08562360529801000),
            Some(1.26085762243604000), Some(1.11225700971196000), Some(1.49066405029967000),
            Some(1.59487944362032000), Some(0.48572827216522300), Some(0.13062513296618000),
            Some(0.13355239428276700), Some(0.07209886266479390), Some(0.05041253535660620),
            Some(0.07257832783270360), Some(0.08863890501902830), Some(0.06203631318696950),
            Some(0.06672377010548700), Some(0.06228923867560830), Some(0.05705690600200920),
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.calmar_ratio(), EPSILON_12, "Calmar");
        }
    }

    // ===== Sterling Ratio tests =====
    #[test]
    fn test_sterling_excess0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(0.74155751780918500), Some(0.89279126631479400), Some(1.15889854414036000),
            Some(1.22179559465027000), Some(1.89088510302246000), Some(1.08562360529801000),
            Some(1.26085762243604000), Some(1.11225700971196000), Some(1.49066405029967000),
            Some(1.59487944362032000), Some(0.48572827216522300), Some(0.13062513296618000),
            Some(0.13355239428276700), Some(0.07209886266479390), Some(0.05041253535660620),
            Some(0.07257832783270360), Some(0.08863890501902830), Some(0.06203631318696950),
            Some(0.06672377010548700), Some(0.06228923867560830), Some(0.05705690600200920),
        ];
        let mut r = Ratios::new(Periodicity::Daily, 0.0, 0.0, DayCountConvention::Raw, None, None);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.sterling_ratio(0.0), EPSILON_12, "Sterling ex=0");
        }
    }

    #[test]
    fn test_sterling_excess002() {
        let expected: [Option<f64>; 24] = [
            Some(0.14999999999999500), Some(0.72174090072224500), Some(0.66442920035313400),
            Some(0.24718583926972700), Some(0.29759708877159600), Some(0.38629951471345200),
            Some(0.40726519821675300), Some(0.63029503434081700), Some(0.44702148453447300),
            Some(0.51917666806190000), Some(0.45798818046963000), Some(0.61380284424104300),
            Some(0.65671506502013300), Some(0.31529729947567100), Some(0.10805355058691200),
            Some(0.11047499102161600), Some(0.06218376425181400), Some(0.04428978919828240),
            Some(0.06376348297771220), Some(0.07787345727187060), Some(0.05450182606873110),
            Some(0.05861997797933450), Some(0.05472403303561820), Some(0.05012718208397050),
        ];
        let excess_annual = (1.0 + 0.02_f64).powf(252.0) - 1.0;
        let mut r = Ratios::new(Periodicity::Daily, 0.0, 0.0, DayCountConvention::Raw, None, None);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.sterling_ratio(excess_annual), EPSILON_12, "Sterling ex=0.02");
        }
    }

    // ===== Burke Ratio tests =====
    #[test]
    fn test_burke_unmodified_rf0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(0.74155751780925900), Some(0.89279126631488400), Some(1.15889854414048000),
            Some(1.22179559465039000), Some(1.89088510302265000), Some(0.88340826476302900),
            Some(1.02600204980225000), Some(0.86912185514484500), Some(1.16481055500805000),
            Some(1.24624485947780000), Some(0.43717141205593600), Some(0.12556405008668100),
            Some(0.12837789439234000), Some(0.08147141226635310), Some(0.05962926105099300),
            Some(0.08584753824355520), Some(0.10484440763127000), Some(0.06479655403731050),
            Some(0.06969257444720940), Some(0.06501838430134020), Some(0.05929350254553110),
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.burke_ratio(false), EPSILON_12, "Burke unmod");
        }
    }

    #[test]
    fn test_burke_modified_rf0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(1.4831150356185200), Some(1.9963419611982000), Some(2.8387100967984600),
            Some(3.2325672963992100), Some(5.3482307151677600), Some(2.6502247942890900),
            Some(3.2445033613766200), Some(2.8825510906130700), Some(4.0350221249328900),
            Some(4.4933997426306100), Some(1.6357456432054900), Some(0.4863074748680710),
            Some(0.5135115775693600), Some(0.3359152382424160), Some(0.2529855290778000),
            Some(0.3742007437554000), Some(0.4688784450484330), Some(0.2969351136482720),
            Some(0.3268871495298580), Some(0.3118172170272280), Some(0.2904776525979330),
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.burke_ratio(true), EPSILON_12, "Burke mod");
        }
    }

    // ===== Pain Index tests =====
    #[test]
    fn test_pain_index() {
        let expected: [Option<f64>; 24] = [
            Some(0.000000000000000000), Some(0.000000000000000000), Some(0.000000000000000000),
            Some(0.002499999999999720), Some(0.001999999999999780), Some(0.001666666666666480),
            Some(0.001428571428571270), Some(0.001249999999999860), Some(0.002666666666666870),
            Some(0.002400000000000180), Some(0.002636363636363750), Some(0.002416666666666770),
            Some(0.002230769230769330), Some(0.004714285714285670), Some(0.010931828666666300),
            Some(0.015310719760193400), Some(0.022054465601410500), Some(0.029269540239837100),
            Some(0.032046520077712100), Some(0.031648180864357100), Some(0.034382271489022800),
            Some(0.035777870937816800), Some(0.037269363727021100), Some(0.039011302642189500),
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.pain_index(), EPSILON, "PainIndex");
        }
    }

    // ===== Pain Ratio tests =====
    #[test]
    fn test_pain_ratio_rf0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(2.9662300712370400), Some(4.4639563315744200), Some(6.9533912648428800),
            Some(8.5525691625527300), Some(15.127080824181200), Some(5.6995239278141000),
            Some(7.3550027975430300), Some(5.9064682584701500), Some(8.6355710500115400),
            Some(10.009243404789200), Some(3.8122309845695300), Some(1.1440393448276400),
            Some(0.8351473403007020), Some(0.4100547525167660), Some(0.2491781707736400),
            Some(0.3276524622550170), Some(0.4051939805814540), Some(0.2610350161067210),
            Some(0.2698071401733870), Some(0.2417956028428860), Some(0.2115948629646200),
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.pain_ratio(), EPSILON_12, "PainRatio");
        }
    }

    // ===== Ulcer Index tests =====
    #[test]
    fn test_ulcer_index() {
        let expected: [Option<f64>; 24] = [
            Some(0.000000000000000000), Some(0.000000000000000000), Some(0.000000000000000000),
            Some(0.004999999999999450), Some(0.004472135954999090), Some(0.004082482904638180),
            Some(0.003779644730091860), Some(0.003535533905932350), Some(0.005734883511362320),
            Some(0.005440588203494720), Some(0.005402019824271570), Some(0.005172040216394730),
            Some(0.004969135507541710), Some(0.010987005311470400), Some(0.027434256917710300),
            Some(0.033400616370435100), Some(0.045203959104378100), Some(0.056676085534212600),
            Some(0.058286267631993700), Some(0.057065017555245000), Some(0.058983749023907400),
            Some(0.059274727347882400), Some(0.059785256332054100), Some(0.060711532990550200),
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.ulcer_index(), EPSILON, "UlcerIndex");
        }
    }

    // ===== Martin Ratio tests =====
    #[test]
    fn test_martin_ratio_rf0() {
        let expected: [Option<f64>; 24] = [
            None, None, None,
            Some(1.4831150356185200), Some(1.9963419611982000), Some(2.8387100967984600),
            Some(3.2325672963992100), Some(5.3482307151677600), Some(2.6502247942890900),
            Some(3.2445033613766200), Some(2.8825510906130700), Some(4.0350221249328900),
            Some(4.4933997426306100), Some(1.6357456432054900), Some(0.4558695408843890),
            Some(0.3828284707085000), Some(0.2000607604567100), Some(0.1286844429639630),
            Some(0.1801474281465850), Some(0.2247200286966700), Some(0.1521601617470090),
            Some(0.1628539762413560), Some(0.1507322845601700), Some(0.1359641377846650),
        ];
        let mut r = new_ratios_with_rf(0.0);
        for i in 0..24 {
            add_bacon_return(&mut r, i);
            assert_nullable(i, expected[i], r.martin_ratio(), EPSILON_12, "MartinRatio");
        }
    }

    // ===== Rolling Window / Min Periods tests =====

    fn new_ratios_with_window(rolling_window: Option<usize>, min_periods: Option<usize>) -> Ratios {
        Ratios::new(Periodicity::Daily, 0.0, 0.0, DayCountConvention::Raw, rolling_window, min_periods)
    }

    fn add_all_bacon(r: &mut Ratios) {
        for i in 0..24 {
            add_bacon_return(r, i);
        }
    }

    #[test]
    fn test_min_periods_returns_none_before_threshold() {
        let mut r = new_ratios_with_window(None, Some(5));
        // First 4 returns: not primed yet
        for i in 0..4 {
            add_bacon_return(&mut r, i);
            assert!(r.sharpe_ratio(true, false).is_none(),
                "step {}: expected None before min_periods", i);
            assert!(r.sortino_ratio(false, false).is_none(),
                "step {}: sortino expected None before min_periods", i);
        }
        // 5th return: now primed
        add_bacon_return(&mut r, 4);
        assert!(r.sharpe_ratio(true, false).is_some(),
            "step 4: expected Some after min_periods");
    }

    #[test]
    fn test_min_periods_zero_or_negative_ignored() {
        // min_periods=0 should be treated as None (always primed)
        let mut r0 = new_ratios_with_window(None, Some(0));
        add_bacon_return(&mut r0, 0);
        add_bacon_return(&mut r0, 1);
        assert!(r0.sharpe_ratio(true, false).is_some());

        // min_periods=-1 should be treated as None
        // In Rust, Option<usize> can't be negative, but test the sanitization:
        // Actually usize can't be negative. Just test 0.
    }

    #[test]
    fn test_min_periods_none_always_primed() {
        let mut r = new_ratios_with_window(None, None);
        add_bacon_return(&mut r, 0);
        add_bacon_return(&mut r, 1);
        assert!(r.sharpe_ratio(true, false).is_some());
    }

    #[test]
    fn test_min_periods_full_dataset_matches() {
        // With min_periods=1 (effectively no delay), full dataset should match baseline
        let mut r_base = new_ratios_with_window(None, None);
        add_all_bacon(&mut r_base);

        let mut r_mp = new_ratios_with_window(None, Some(1));
        add_all_bacon(&mut r_mp);

        let s1 = r_base.sharpe_ratio(true, false).unwrap();
        let s2 = r_mp.sharpe_ratio(true, false).unwrap();
        assert!(almost_equal(s1, s2, 1e-15), "sharpe mismatch");

        let o1 = r_base.omega_ratio().unwrap();
        let o2 = r_mp.omega_ratio().unwrap();
        assert!(almost_equal(o1, o2, 1e-15), "omega mismatch");
    }

    #[test]
    fn test_rolling_window_matches_fresh_instance() {
        // A rolling window of 10 after 24 returns should match
        // a fresh instance fed only the last 10 returns.
        let mut r_rolling = new_ratios_with_window(Some(10), None);
        add_all_bacon(&mut r_rolling);

        // Fresh instance with last 10
        let mut r_fresh = new_ratios_with_window(None, None);
        for i in 14..24 {
            add_bacon_return(&mut r_fresh, i);
        }

        let eps = 1e-13;

        let s1 = r_rolling.sharpe_ratio(true, false).unwrap();
        let s2 = r_fresh.sharpe_ratio(true, false).unwrap();
        assert!(almost_equal(s1, s2, eps),
            "sharpe: rolling={:.16e}, fresh={:.16e}", s1, s2);

        let o1 = r_rolling.omega_ratio().unwrap();
        let o2 = r_fresh.omega_ratio().unwrap();
        assert!(almost_equal(o1, o2, eps),
            "omega: rolling={:.16e}, fresh={:.16e}", o1, o2);

        let sk1 = r_rolling.skew().unwrap();
        let sk2 = r_fresh.skew().unwrap();
        assert!(almost_equal(sk1, sk2, eps),
            "skew: rolling={:.16e}, fresh={:.16e}", sk1, sk2);

        let k1 = r_rolling.kurtosis().unwrap();
        let k2 = r_fresh.kurtosis().unwrap();
        assert!(almost_equal(k1, k2, eps),
            "kurtosis: rolling={:.16e}, fresh={:.16e}", k1, k2);

        let pi1 = r_rolling.pain_index().unwrap();
        let pi2 = r_fresh.pain_index().unwrap();
        assert!(almost_equal(pi1, pi2, eps),
            "pain_index: rolling={:.16e}, fresh={:.16e}", pi1, pi2);

        let ui1 = r_rolling.ulcer_index().unwrap();
        let ui2 = r_fresh.ulcer_index().unwrap();
        assert!(almost_equal(ui1, ui2, eps),
            "ulcer_index: rolling={:.16e}, fresh={:.16e}", ui1, ui2);

        let ror1 = r_rolling.risk_of_ruin().unwrap();
        let ror2 = r_fresh.risk_of_ruin().unwrap();
        assert!(almost_equal(ror1, ror2, eps),
            "risk_of_ruin: rolling={:.16e}, fresh={:.16e}", ror1, ror2);
    }

    #[test]
    fn test_rolling_window_none_is_expanding() {
        // rolling_window=None should give same results as no window at all
        let mut r1 = new_ratios_with_window(None, None);
        add_all_bacon(&mut r1);

        let mut r2 = new_ratios_with_rf(0.0);
        add_all_bacon(&mut r2);

        let s1 = r1.sharpe_ratio(true, false).unwrap();
        let s2 = r2.sharpe_ratio(true, false).unwrap();
        assert!(almost_equal(s1, s2, 1e-15));
    }

    #[test]
    fn test_rolling_window_larger_than_data() {
        // window=100 with 24 returns should behave like expanding
        let mut r_large = new_ratios_with_window(Some(100), None);
        add_all_bacon(&mut r_large);

        let mut r_base = new_ratios_with_window(None, None);
        add_all_bacon(&mut r_base);

        let s1 = r_large.sharpe_ratio(true, false).unwrap();
        let s2 = r_base.sharpe_ratio(true, false).unwrap();
        assert!(almost_equal(s1, s2, 1e-15));
    }

    #[test]
    fn test_rolling_window_with_min_periods() {
        // window=10, min_periods=5: first 4 should be None
        let mut r = new_ratios_with_window(Some(10), Some(5));
        for i in 0..4 {
            add_bacon_return(&mut r, i);
            assert!(r.sharpe_ratio(true, false).is_none(),
                "step {}: expected None before min_periods", i);
        }
        // 5th: now primed
        add_bacon_return(&mut r, 4);
        assert!(r.sharpe_ratio(true, false).is_some(),
            "step 4: expected Some after min_periods");

        // Continue to full dataset
        for i in 5..24 {
            add_bacon_return(&mut r, i);
        }

        // Should match a fresh instance fed last 10 returns
        let mut r_fresh = new_ratios_with_window(None, None);
        for i in 14..24 {
            add_bacon_return(&mut r_fresh, i);
        }

        let eps = 1e-13;
        let s1 = r.sharpe_ratio(true, false).unwrap();
        let s2 = r_fresh.sharpe_ratio(true, false).unwrap();
        assert!(almost_equal(s1, s2, eps),
            "sharpe: rolling={:.16e}, fresh={:.16e}", s1, s2);
    }

    #[test]
    fn test_rolling_window_with_min_periods_clamped() {
        // min_periods > window: min_periods dominates
        let mut r = new_ratios_with_window(Some(5), Some(10));
        for i in 0..9 {
            add_bacon_return(&mut r, i);
            assert!(r.sharpe_ratio(true, false).is_none(),
                "step {}: expected None before min_periods", i);
        }
        // 10th return: primed
        add_bacon_return(&mut r, 9);
        assert!(r.sharpe_ratio(true, false).is_some(),
            "step 9: expected Some after min_periods");
    }
}
