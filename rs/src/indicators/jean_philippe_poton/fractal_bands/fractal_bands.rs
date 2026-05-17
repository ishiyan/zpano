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

/// Parameters to create an instance of the fractal bands indicator.
pub struct FractalBandsParams {
    /// The lookback period for FGDI computation. Must be greater than 1.
    pub period: usize,
    /// Base SMA period before fractal adaptation. Must be greater than 0.
    pub normal_speed: usize,
    /// Band width multiplier raised to power H. Must be greater than 0.
    pub alpha: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for FractalBandsParams {
    fn default() -> Self {
        Self {
            period: 30,
            normal_speed: 20,
            alpha: 2.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the fractal bands indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum FractalBandsOutput {
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

/// Computes the Fractal Bands indicator.
///
/// FRASMA2 center line with upper/lower bands scaled by alpha^H where H is
/// the local Hurst exponent estimated from the Fractal Graph Dimension Index.
///
/// The indicator is not primed during the first `period - 1` updates.
pub struct FractalBands {
    line: LineIndicator,
    window: Vec<f64>,
    closes: Vec<f64>,
    period: usize,
    period_minus_1: usize,
    normal_speed: usize,
    alpha: f64,
    window_count: usize,
    primed: bool,
    log_denom: f64,
    ln2: f64,
    inv_period_sq: f64,
    frasma2: f64,
    upper_band: f64,
    lower_band: f64,
}

impl FractalBands {
    /// Creates a new FractalBands from the given parameters.
    pub fn new(params: &FractalBandsParams) -> Result<Self, String> {
        if params.period < 2 {
            return Err("invalid fractal bands parameters: period should be greater than 1".to_string());
        }
        if params.normal_speed < 1 {
            return Err("invalid fractal bands parameters: normal_speed should be greater than 0".to_string());
        }
        if params.alpha <= 0.0 {
            return Err("invalid fractal bands parameters: alpha should be greater than 0".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("fban({},{},{}{})", params.period, params.normal_speed, params.alpha, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Fractal bands {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);
        let period_f = params.period as f64;
        let period_minus_1 = params.period - 1;
        let period_minus_1_f = period_minus_1 as f64;

        Ok(Self {
            line,
            window: vec![0.0; params.period],
            closes: Vec::with_capacity(256),
            period: params.period,
            period_minus_1,
            normal_speed: params.normal_speed,
            alpha: params.alpha,
            window_count: 0,
            primed: false,
            log_denom: (2.0 * period_minus_1_f).ln(),
            ln2: 2.0_f64.ln(),
            inv_period_sq: 1.0 / (period_f * period_f),
            frasma2: f64::NAN,
            upper_band: f64::NAN,
            lower_band: f64::NAN,
        })
    }

    /// Core update logic. Returns the FRASMA2 value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let period = self.period;
        let period_minus_1 = self.period_minus_1;

        // Accumulate close history for SMA computation.
        self.closes.push(sample);

        // Fill the FGDI window.
        if self.window_count < period {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if self.window_count < period {
                return f64::NAN;
            }

            self.primed = true;
        } else {
            for i in 0..period_minus_1 {
                self.window[i] = self.window[i + 1];
            }
            self.window[period_minus_1] = sample;
        }

        // Find min/max for normalization.
        let mut price_max = self.window[0];
        let mut price_min = self.window[0];

        for k in 1..period {
            if self.window[k] > price_max { price_max = self.window[k]; }
            if self.window[k] < price_min { price_min = self.window[k]; }
        }

        let price_range = price_max - price_min;

        let fgdi;

        if price_range <= 0.0 {
            fgdi = 0.0;
        } else {
            // Compute normalized path length: period points, period-1 segments.
            let mut prior_norm = (self.window[0] - price_min) / price_range;
            let mut length = 0.0;

            for k in 1..period {
                let curr_norm = (self.window[k] - price_min) / price_range;
                let diff = curr_norm - prior_norm;
                length += (diff * diff + self.inv_period_sq).sqrt();
                prior_norm = curr_norm;
            }

            if length > 0.0 {
                fgdi = 1.0 + (length.ln() + self.ln2) / self.log_denom;
            } else {
                fgdi = 0.0;
            }
        }

        // Hurst exponent.
        let mut hurst = 2.0 - fgdi;
        if hurst < 0.01 {
            hurst = 0.01;
        }

        let trail_dim = 1.0 / hurst;
        let beta = trail_dim / 2.0;
        let speed_f = (self.normal_speed as f64 * beta).round();
        let speed = if speed_f < 1.0 { 1_usize } else { speed_f as usize };

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

        // Deviation over the FGDI lookback window (period bars).
        let mut sq_sum = 0.0;
        for k in 0..period {
            let res = self.window[k] - frasma2_val;
            sq_sum += res * res;
        }

        let deviation = 2.0 * (sq_sum / period as f64).sqrt();

        // Fractal bands.
        let band_mult = deviation * self.alpha.powf(hurst);
        let ub = frasma2_val + band_mult;
        let lb = frasma2_val - band_mult;

        self.frasma2 = frasma2_val;
        self.upper_band = ub;
        self.lower_band = lb;

        frasma2_val
    }

    /// Updates and returns all three outputs: (frasma2, upper_band, lower_band).
    pub fn update_all(&mut self, sample: f64) -> (f64, f64, f64) {
        let frasma2_val = self.update(sample);
        (frasma2_val, self.upper_band, self.lower_band)
    }

    /// Returns the last computed FRASMA2 value.
    pub fn frasma2_value(&self) -> f64 { self.frasma2 }

    /// Returns the last computed upper band value.
    pub fn upper_band_value(&self) -> f64 { self.upper_band }

    /// Returns the last computed lower band value.
    pub fn lower_band_value(&self) -> f64 { self.lower_band }
}

impl Indicator for FractalBands {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::FractalBands,
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
    use crate::indicators::jean_philippe_poton::fractal_bands::testdata::testdata;

    const EPSILON: f64 = 1e-13;

    fn check_value(exp: f64, act: f64, label: &str, i: usize) {
        if exp.is_nan() {
            assert!(act.is_nan(), "[{}] {}: expected NaN, got {}", i, label, act);
        } else {
            assert!((exp - act).abs() < EPSILON, "[{}] {}: expected {}, got {}", i, label, exp, act);
        }
    }

    fn run_test(period: usize, normal_speed: usize, alpha: f64, exp_frasma2: &[f64], exp_upper: &[f64], exp_lower: &[f64]) {
        let mut ind = FractalBands::new(&FractalBandsParams { period, normal_speed, alpha, ..Default::default() }).unwrap();
        let input = testdata::test_input();

        for i in 0..input.len() {
            let (frasma2, upper, lower) = ind.update_all(input[i]);
            check_value(exp_frasma2[i], frasma2, "frasma2", i);
            check_value(exp_upper[i], upper, "upper", i);
            check_value(exp_lower[i], lower, "lower", i);
        }
    }

    #[test]
    fn test_p10_ns20_a2() {
        run_test(10, 20, 2.0, &testdata::expected_frasma2_p10_ns20_a2(), &testdata::expected_upper_p10_ns20_a2(), &testdata::expected_lower_p10_ns20_a2());
    }

    #[test]
    fn test_p20_ns20_a2() {
        run_test(20, 20, 2.0, &testdata::expected_frasma2_p20_ns20_a2(), &testdata::expected_upper_p20_ns20_a2(), &testdata::expected_lower_p20_ns20_a2());
    }

    #[test]
    fn test_p30_ns20_a2() {
        run_test(30, 20, 2.0, &testdata::expected_frasma2_p30_ns20_a2(), &testdata::expected_upper_p30_ns20_a2(), &testdata::expected_lower_p30_ns20_a2());
    }

    #[test]
    fn test_p50_ns20_a2() {
        run_test(50, 20, 2.0, &testdata::expected_frasma2_p50_ns20_a2(), &testdata::expected_upper_p50_ns20_a2(), &testdata::expected_lower_p50_ns20_a2());
    }

    #[test]
    fn test_p30_ns10_a2() {
        run_test(30, 10, 2.0, &testdata::expected_frasma2_p30_ns10_a2(), &testdata::expected_upper_p30_ns10_a2(), &testdata::expected_lower_p30_ns10_a2());
    }

    #[test]
    fn test_p30_ns40_a2() {
        run_test(30, 40, 2.0, &testdata::expected_frasma2_p30_ns40_a2(), &testdata::expected_upper_p30_ns40_a2(), &testdata::expected_lower_p30_ns40_a2());
    }

    #[test]
    fn test_p30_ns20_a1() {
        run_test(30, 20, 1.0, &testdata::expected_frasma2_p30_ns20_a1(), &testdata::expected_upper_p30_ns20_a1(), &testdata::expected_lower_p30_ns20_a1());
    }

    #[test]
    fn test_p30_ns20_a3() {
        run_test(30, 20, 3.0, &testdata::expected_frasma2_p30_ns20_a3(), &testdata::expected_upper_p30_ns20_a3(), &testdata::expected_lower_p30_ns20_a3());
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut ind = FractalBands::new(&FractalBandsParams { period: 30, normal_speed: 20, alpha: 2.0, ..Default::default() }).unwrap();

        for i in 0..29 {
            ind.update(input[i]);
            assert!(!ind.is_primed(), "expected not primed at index {}", i);
        }
        ind.update(input[29]);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_nan_passthrough() {
        let mut ind = FractalBands::new(&FractalBandsParams { period: 5, normal_speed: 20, alpha: 2.0, ..Default::default() }).unwrap();
        let (frasma2, upper, lower) = ind.update_all(f64::NAN);
        assert!(frasma2.is_nan());
        assert!(upper.is_nan());
        assert!(lower.is_nan());
    }

    #[test]
    fn test_invalid_period() {
        let result = FractalBands::new(&FractalBandsParams { period: 1, normal_speed: 20, alpha: 2.0, ..Default::default() });
        assert!(result.is_err());
    }

    #[test]
    fn test_invalid_normal_speed() {
        let result = FractalBands::new(&FractalBandsParams { period: 30, normal_speed: 0, alpha: 2.0, ..Default::default() });
        assert!(result.is_err());
    }

    #[test]
    fn test_invalid_alpha() {
        let result = FractalBands::new(&FractalBandsParams { period: 30, normal_speed: 20, alpha: 0.0, ..Default::default() });
        assert!(result.is_err());
    }

    #[test]
    fn test_metadata() {
        let ind = FractalBands::new(&FractalBandsParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::FractalBands);
        assert!(meta.mnemonic.contains("fban(30"));
    }
}
