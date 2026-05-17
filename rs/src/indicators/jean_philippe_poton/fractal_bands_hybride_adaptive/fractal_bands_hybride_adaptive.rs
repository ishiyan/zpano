use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;
use crate::indicators::core::outputs::band::Band;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the fractal bands hybride adaptive indicator.
pub struct FractalBandsHybrideAdaptiveParams {
    /// The lookback period for FGDI computation. Must be greater than 1.
    pub period: usize,
    /// Fallback SMA period when CyclePeriod is unavailable. Must be greater than 0.
    pub normal_speed_fallback: usize,
    /// Band width multiplier raised to power H. Must be greater than 0.
    pub alpha: f64,
    /// Nyquist multiplier applied to the estimated cycle period. Must be greater than 0.
    pub nyquist: f64,
    /// High-pass filter alpha for Ehlers CyclePeriod. Must be between 0 and 1.
    pub alpha_hp: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for FractalBandsHybrideAdaptiveParams {
    fn default() -> Self {
        Self {
            period: 30,
            normal_speed_fallback: 30,
            alpha: 2.0,
            nyquist: 0.5,
            alpha_hp: 0.07,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the fractal bands hybride adaptive indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum FractalBandsHybrideAdaptiveOutput {
    /// The FRASMA2 center line.
    Frasma2 = 0,
    /// The upper band.
    Upper = 1,
    /// The lower band.
    Lower = 2,
    /// The lower/upper band pair.
    Band = 3,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Fractal Bands Hybride Adaptive indicator.
pub struct FractalBandsHybrideAdaptive {
    line: LineIndicator,
    window: Vec<f64>,
    closes: Vec<f64>,
    period: usize,
    window_size: usize,
    normal_speed_fallback: usize,
    alpha: f64,
    nyquist: f64,
    alpha_hp: f64,
    window_count: usize,
    primed: bool,
    log_denom: f64,
    ln2: f64,
    inv_period_sq: f64,
    // Ehlers CyclePeriod buffers.
    smooth_buf: Vec<f64>,
    cycle_buf: Vec<f64>,
    q1_buf: Vec<f64>,
    i1_buf: Vec<f64>,
    dp_buf: Vec<f64>,
    inst_period_buf: Vec<f64>,
    // Last computed values.
    frasma2: f64,
    upper_band: f64,
    lower_band: f64,
}

impl FractalBandsHybrideAdaptive {
    /// Creates a new FractalBandsHybrideAdaptive from the given parameters.
    pub fn new(params: &FractalBandsHybrideAdaptiveParams) -> Result<Self, String> {
        if params.period < 2 {
            return Err("invalid fractal bands hybride adaptive parameters: period should be greater than 1".to_string());
        }
        if params.normal_speed_fallback < 1 {
            return Err("invalid fractal bands hybride adaptive parameters: normal_speed_fallback should be greater than 0".to_string());
        }
        if params.alpha <= 0.0 {
            return Err("invalid fractal bands hybride adaptive parameters: alpha should be greater than 0".to_string());
        }
        if params.nyquist <= 0.0 {
            return Err("invalid fractal bands hybride adaptive parameters: nyquist should be greater than 0".to_string());
        }
        if params.alpha_hp <= 0.0 || params.alpha_hp >= 1.0 {
            return Err("invalid fractal bands hybride adaptive parameters: alpha_hp should be between 0 and 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("fbanha({},{},{},{},{}{})", params.period, params.normal_speed_fallback,
            params.alpha, params.nyquist, params.alpha_hp, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Fractal bands hybride adaptive {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);
        let period_f = params.period as f64;
        let period_minus_1_f = (params.period - 1) as f64;
        let window_size = params.period + 1;

        Ok(Self {
            line,
            window: vec![0.0; window_size],
            closes: Vec::with_capacity(256),
            period: params.period,
            window_size,
            normal_speed_fallback: params.normal_speed_fallback,
            alpha: params.alpha,
            nyquist: params.nyquist,
            alpha_hp: params.alpha_hp,
            window_count: 0,
            primed: false,
            log_denom: (2.0 * period_minus_1_f).ln(),
            ln2: 2.0_f64.ln(),
            inv_period_sq: 1.0 / (period_f * period_f),
            smooth_buf: Vec::with_capacity(256),
            cycle_buf: Vec::with_capacity(256),
            q1_buf: Vec::with_capacity(256),
            i1_buf: Vec::with_capacity(256),
            dp_buf: Vec::with_capacity(256),
            inst_period_buf: Vec::with_capacity(256),
            frasma2: f64::NAN,
            upper_band: f64::NAN,
            lower_band: f64::NAN,
        })
    }

    fn get_cycle_period(&mut self) -> f64 {
        let t = self.closes.len() - 1;

        // Extend buffers to index t.
        while self.smooth_buf.len() <= t { self.smooth_buf.push(0.0); }
        while self.cycle_buf.len() <= t { self.cycle_buf.push(0.0); }
        while self.q1_buf.len() <= t { self.q1_buf.push(0.0); }
        while self.i1_buf.len() <= t { self.i1_buf.push(0.0); }
        while self.dp_buf.len() <= t { self.dp_buf.push(0.0); }
        while self.inst_period_buf.len() <= t { self.inst_period_buf.push(6.0); }

        if t < 6 {
            return f64::NAN;
        }

        let prices = &self.closes;

        // 4-bar weighted smoother.
        self.smooth_buf[t] = (prices[t] + 2.0 * prices[t - 1] +
            2.0 * prices[t - 2] + prices[t - 3]) / 6.0;

        // High-pass filter.
        let alpha_hp = self.alpha_hp;
        let hp_coeff = (1.0 - 0.5 * alpha_hp) * (1.0 - 0.5 * alpha_hp);
        let one_minus_alpha = 1.0 - alpha_hp;

        self.cycle_buf[t] = hp_coeff * (self.smooth_buf[t] - 2.0 * self.smooth_buf[t - 1] + self.smooth_buf[t - 2]) +
            2.0 * one_minus_alpha * self.cycle_buf[t - 1] - one_minus_alpha * one_minus_alpha * self.cycle_buf[t - 2];

        // Quadrature component.
        self.q1_buf[t] = (0.0962 * self.cycle_buf[t] + 0.5769 * self.cycle_buf[t - 2] -
            0.5769 * self.cycle_buf[t - 4] - 0.0962 * self.cycle_buf[t - 6]) *
            (0.5 + 0.08 * self.inst_period_buf[t - 1]);

        // In-phase component.
        self.i1_buf[t] = self.cycle_buf[t - 3];

        // Smooth I and Q with EMA.
        if t > 6 {
            self.i1_buf[t] = 0.15 * self.i1_buf[t] + 0.85 * self.i1_buf[t - 1];
            self.q1_buf[t] = 0.15 * self.q1_buf[t] + 0.85 * self.q1_buf[t - 1];
        }

        // Compute delta phase.
        let dp = if self.i1_buf[t].abs() > 1e-10 {
            (self.q1_buf[t] / self.i1_buf[t]).atan()
        } else {
            self.dp_buf[t - 1]
        };

        // Clamp delta phase.
        let dp = dp.max(0.1).min(1.1);
        self.dp_buf[t] = dp;

        // Median delta phase over 5 bars.
        let median_dp = if t >= 10 {
            let mut w = [self.dp_buf[t - 4], self.dp_buf[t - 3], self.dp_buf[t - 2],
                self.dp_buf[t - 1], self.dp_buf[t]];
            w.sort_by(|a, b| a.partial_cmp(b).unwrap());
            w[2]
        } else {
            dp
        };

        // Instantaneous period.
        let dc = if median_dp.abs() > 1e-10 {
            6.2832 / median_dp + 0.5
        } else {
            self.inst_period_buf[t - 1]
        };

        // Clamp and smooth.
        let dc = dc.max(6.0).min(50.0);
        self.inst_period_buf[t] = 0.33 * dc + 0.67 * self.inst_period_buf[t - 1];

        self.inst_period_buf[t]
    }

    /// Core update logic. Returns the FRASMA2 value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let period = self.period;
        let window_size = self.window_size;

        // Accumulate close history.
        self.closes.push(sample);

        // Update Ehlers CyclePeriod.
        let cp = self.get_cycle_period();

        // Fill the FGDI window (period+1 elements).
        if self.window_count < window_size {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if self.window_count < window_size {
                return f64::NAN;
            }

            self.primed = true;
        } else {
            for i in 0..(window_size - 1) {
                self.window[i] = self.window[i + 1];
            }
            self.window[window_size - 1] = sample;
        }

        // FGDI computation over period+1 points.
        let mut price_max = self.window[0];
        let mut price_min = self.window[0];

        for k in 1..window_size {
            if self.window[k] > price_max { price_max = self.window[k]; }
            if self.window[k] < price_min { price_min = self.window[k]; }
        }

        let price_range = price_max - price_min;

        let fgdi = if price_range < 1e-10 {
            1.0
        } else {
            let mut length = 0.0;
            for i in 1..window_size {
                let norm_cur = (self.window[i] - price_min) / price_range;
                let norm_prev = (self.window[i - 1] - price_min) / price_range;
                let diff = norm_cur - norm_prev;
                length += (diff * diff + self.inv_period_sq).sqrt();
            }
            1.0 + (length.ln() + self.ln2) / self.log_denom
        };

        // Hurst exponent.
        let mut hurst = 2.0 - fgdi;
        if hurst < 0.01 { hurst = 0.01; }

        let trail_dim = 1.0 / hurst;
        let beta = trail_dim / 2.0;

        // Adaptive normal_speed from CyclePeriod.
        let ns = if cp.is_nan() || cp < 1.0 {
            self.normal_speed_fallback as f64
        } else {
            cp * self.nyquist
        };

        let speed = (ns * beta).round().max(1.0) as usize;

        // FRASMA2: SMA of close over 'speed' bars ending at current position.
        let n_closes = self.closes.len();
        if speed > n_closes {
            self.frasma2 = f64::NAN;
            self.upper_band = f64::NAN;
            self.lower_band = f64::NAN;
            return f64::NAN;
        }

        let mut sma_sum = 0.0;
        for k in (n_closes - speed)..n_closes {
            sma_sum += self.closes[k];
        }
        let frasma2_val = sma_sum / speed as f64;

        // Deviation over the last period closes.
        let dev_start = if n_closes > period { n_closes - period } else { 0 };
        let mut sq_sum = 0.0;
        for k in dev_start..n_closes {
            let res = self.closes[k] - frasma2_val;
            sq_sum += res * res;
        }
        let deviation = 2.0 * (sq_sum / period as f64).sqrt();

        // Fractal bands.
        let band_mult = deviation * self.alpha.powf(hurst);
        let upper_band_val = frasma2_val + band_mult;
        let lower_band_val = frasma2_val - band_mult;

        self.frasma2 = frasma2_val;
        self.upper_band = upper_band_val;
        self.lower_band = lower_band_val;

        frasma2_val
    }

    /// Updates and returns all three outputs.
    pub fn update_all(&mut self, sample: f64) -> (f64, f64, f64) {
        let frasma2 = self.update(sample);
        (frasma2, self.upper_band, self.lower_band)
    }
}

impl Indicator for FractalBandsHybrideAdaptive {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::FractalBandsHybrideAdaptive,
            &self.line.mnemonic,
            &self.line.description,
            &[
                OutputText { mnemonic: self.line.mnemonic.clone(), description: self.line.description.clone() },
                OutputText { mnemonic: format!("{} upper", self.line.mnemonic), description: format!("{} Upper Band", self.line.description) },
                OutputText { mnemonic: format!("{} lower", self.line.mnemonic), description: format!("{} Lower Band", self.line.description) },
                OutputText { mnemonic: format!("{} band", self.line.mnemonic), description: format!("{} Band", self.line.description) },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (frasma2, upper, lower) = self.update_all(sample.value);
        let t = sample.time;

        let band: Box<dyn std::any::Any> = if lower.is_nan() || upper.is_nan() {
            Box::new(Band::empty(t))
        } else {
            Box::new(Band::new(t, lower, upper))
        };

        vec![
            Box::new(Scalar::new(t, frasma2)),
            Box::new(Scalar::new(t, upper)),
            Box::new(Scalar::new(t, lower)),
            band,
        ]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.line.bar_func)(sample);
        self.update_scalar(&Scalar::new(sample.time, v))
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.line.quote_func)(sample);
        self.update_scalar(&Scalar::new(sample.time, v))
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.line.trade_func)(sample);
        self.update_scalar(&Scalar::new(sample.time, v))
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;

    fn almost_equal(a: f64, b: f64, epsilon: f64) -> bool {
        if a.is_nan() && b.is_nan() { return true; }
        if a.is_nan() || b.is_nan() { return false; }
        (a - b).abs() <= epsilon
    }

    #[test]
    fn test_p10_ny05_ahp007() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 10, normal_speed_fallback: 30, alpha: 2.0, nyquist: 0.5, alpha_hp: 0.07,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p10_ny05_ahp007();
        let exp_upper = testdata::expected_upper_p10_ny05_ahp007();
        let exp_lower = testdata::expected_lower_p10_ny05_ahp007();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p10_ny05_ahp015() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 10, normal_speed_fallback: 30, alpha: 2.0, nyquist: 0.5, alpha_hp: 0.15,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p10_ny05_ahp015();
        let exp_upper = testdata::expected_upper_p10_ny05_ahp015();
        let exp_lower = testdata::expected_lower_p10_ny05_ahp015();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p10_ny10_ahp007() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 10, normal_speed_fallback: 30, alpha: 2.0, nyquist: 1.0, alpha_hp: 0.07,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p10_ny10_ahp007();
        let exp_upper = testdata::expected_upper_p10_ny10_ahp007();
        let exp_lower = testdata::expected_lower_p10_ny10_ahp007();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p10_ny10_ahp015() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 10, normal_speed_fallback: 30, alpha: 2.0, nyquist: 1.0, alpha_hp: 0.15,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p10_ny10_ahp015();
        let exp_upper = testdata::expected_upper_p10_ny10_ahp015();
        let exp_lower = testdata::expected_lower_p10_ny10_ahp015();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p20_ny05_ahp007() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 20, normal_speed_fallback: 30, alpha: 2.0, nyquist: 0.5, alpha_hp: 0.07,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p20_ny05_ahp007();
        let exp_upper = testdata::expected_upper_p20_ny05_ahp007();
        let exp_lower = testdata::expected_lower_p20_ny05_ahp007();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p20_ny05_ahp015() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 20, normal_speed_fallback: 30, alpha: 2.0, nyquist: 0.5, alpha_hp: 0.15,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p20_ny05_ahp015();
        let exp_upper = testdata::expected_upper_p20_ny05_ahp015();
        let exp_lower = testdata::expected_lower_p20_ny05_ahp015();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p20_ny10_ahp007() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 20, normal_speed_fallback: 30, alpha: 2.0, nyquist: 1.0, alpha_hp: 0.07,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p20_ny10_ahp007();
        let exp_upper = testdata::expected_upper_p20_ny10_ahp007();
        let exp_lower = testdata::expected_lower_p20_ny10_ahp007();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p20_ny10_ahp015() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 20, normal_speed_fallback: 30, alpha: 2.0, nyquist: 1.0, alpha_hp: 0.15,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p20_ny10_ahp015();
        let exp_upper = testdata::expected_upper_p20_ny10_ahp015();
        let exp_lower = testdata::expected_lower_p20_ny10_ahp015();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p30_ny05_ahp007() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 30, normal_speed_fallback: 30, alpha: 2.0, nyquist: 0.5, alpha_hp: 0.07,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p30_ny05_ahp007();
        let exp_upper = testdata::expected_upper_p30_ny05_ahp007();
        let exp_lower = testdata::expected_lower_p30_ny05_ahp007();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p30_ny05_ahp015() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 30, normal_speed_fallback: 30, alpha: 2.0, nyquist: 0.5, alpha_hp: 0.15,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p30_ny05_ahp015();
        let exp_upper = testdata::expected_upper_p30_ny05_ahp015();
        let exp_lower = testdata::expected_lower_p30_ny05_ahp015();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p30_ny10_ahp007() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 30, normal_speed_fallback: 30, alpha: 2.0, nyquist: 1.0, alpha_hp: 0.07,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p30_ny10_ahp007();
        let exp_upper = testdata::expected_upper_p30_ny10_ahp007();
        let exp_lower = testdata::expected_lower_p30_ny10_ahp007();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p30_ny10_ahp015() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 30, normal_speed_fallback: 30, alpha: 2.0, nyquist: 1.0, alpha_hp: 0.15,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p30_ny10_ahp015();
        let exp_upper = testdata::expected_upper_p30_ny10_ahp015();
        let exp_lower = testdata::expected_lower_p30_ny10_ahp015();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p50_ny05_ahp007() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 50, normal_speed_fallback: 30, alpha: 2.0, nyquist: 0.5, alpha_hp: 0.07,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p50_ny05_ahp007();
        let exp_upper = testdata::expected_upper_p50_ny05_ahp007();
        let exp_lower = testdata::expected_lower_p50_ny05_ahp007();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p50_ny05_ahp015() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 50, normal_speed_fallback: 30, alpha: 2.0, nyquist: 0.5, alpha_hp: 0.15,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p50_ny05_ahp015();
        let exp_upper = testdata::expected_upper_p50_ny05_ahp015();
        let exp_lower = testdata::expected_lower_p50_ny05_ahp015();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p50_ny10_ahp007() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 50, normal_speed_fallback: 30, alpha: 2.0, nyquist: 1.0, alpha_hp: 0.07,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p50_ny10_ahp007();
        let exp_upper = testdata::expected_upper_p50_ny10_ahp007();
        let exp_lower = testdata::expected_lower_p50_ny10_ahp007();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_p50_ny10_ahp015() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 50, normal_speed_fallback: 30, alpha: 2.0, nyquist: 1.0, alpha_hp: 0.15,
            ..Default::default()
        }).unwrap();

        let input = testdata::test_input();
        let exp_frasma = testdata::expected_frasma_p50_ny10_ahp015();
        let exp_upper = testdata::expected_upper_p50_ny10_ahp015();
        let exp_lower = testdata::expected_lower_p50_ny10_ahp015();

        let epsilon = 2e-13;
        for i in 0..252 {
            let (f, u, l) = ind.update_all(input[i]);
            assert!(almost_equal(f, exp_frasma[i], epsilon), "frasma at {}: {} vs {}", i, f, exp_frasma[i]);
            assert!(almost_equal(u, exp_upper[i], epsilon), "upper at {}: {} vs {}", i, u, exp_upper[i]);
            assert!(almost_equal(l, exp_lower[i], epsilon), "lower at {}: {} vs {}", i, l, exp_lower[i]);
        }
    }

    #[test]
    fn test_is_primed() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams::default()).unwrap();
        let input = testdata::test_input();
        for i in 0..30 {
            ind.update(input[i]);
            assert!(!ind.is_primed(), "should not be primed at index {}", i);
        }
        ind.update(input[30]);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_nan_passthrough() {
        let mut ind = FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams {
            period: 5, ..Default::default()
        }).unwrap();
        let (f, u, l) = ind.update_all(f64::NAN);
        assert!(f.is_nan());
        assert!(u.is_nan());
        assert!(l.is_nan());
    }

    #[test]
    fn test_invalid_params() {
        assert!(FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams { period: 1, ..Default::default() }).is_err());
        assert!(FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams { normal_speed_fallback: 0, ..Default::default() }).is_err());
        assert!(FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams { alpha: 0.0, ..Default::default() }).is_err());
        assert!(FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams { nyquist: 0.0, ..Default::default() }).is_err());
        assert!(FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams { alpha_hp: 0.0, ..Default::default() }).is_err());
        assert!(FractalBandsHybrideAdaptive::new(&FractalBandsHybrideAdaptiveParams { alpha_hp: 1.0, ..Default::default() }).is_err());
    }
}
