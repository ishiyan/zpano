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

/// Parameters to create an instance of the fractional bands indicator.
pub struct FractionalBandsParams {
    /// The lookback period for FGDI computation. Must be greater than 1.
    pub period: usize,
    /// Price-to-working-space multiplier. Must be greater than 0.
    pub price_scale: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for FractionalBandsParams {
    fn default() -> Self {
        Self {
            period: 30,
            price_scale: 1.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the fractional bands indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum FractionalBandsOutput {
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

/// Computes the Fractional Bands indicator.
///
/// Fractal-adaptive moving average with FBM-scaled volatility bands.
/// Uses fractional Brownian motion power law: band_width = 2 * deviation^(2*H)
/// where H is the Hurst exponent derived from the Fractal Graph Dimension Index.
///
/// The indicator is not primed during the first `period` updates.
pub struct FractionalBands {
    line: LineIndicator,
    window: Vec<f64>,
    closes: Vec<f64>,
    period: usize,
    window_size: usize,
    price_scale: f64,
    window_count: usize,
    primed: bool,
    log_denom: f64,
    ln2: f64,
    inv_period_sq: f64,
    frasma2: f64,
    upper_band: f64,
    lower_band: f64,
}

impl FractionalBands {
    /// Creates a new FractionalBands from the given parameters.
    pub fn new(params: &FractionalBandsParams) -> Result<Self, String> {
        if params.period < 2 {
            return Err("invalid fractional bands parameters: period should be greater than 1".to_string());
        }
        if params.price_scale <= 0.0 {
            return Err("invalid fractional bands parameters: price_scale should be greater than 0".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("fctban({},{}{})", params.period, params.price_scale, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Fractional bands {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);
        let period_f = params.period as f64;
        let window_size = params.period + 1;
        let period_minus_1_f = (params.period - 1) as f64;

        Ok(Self {
            line,
            window: vec![0.0; window_size],
            closes: Vec::with_capacity(256),
            period: params.period,
            window_size,
            price_scale: params.price_scale,
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
        let window_size = self.window_size;
        let p = self.price_scale;

        // Accumulate close history.
        self.closes.push(sample);

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

        let fgdi;

        if price_range < 1e-10 {
            fgdi = 1.0;
        } else {
            let inv_range = 1.0 / price_range;
            let mut prev_norm = (self.window[0] - price_min) * inv_range;
            let mut length = 0.0;

            for i in 1..period { // period-1 segments
                let cur_norm = (self.window[i] - price_min) * inv_range;
                let diff = cur_norm - prev_norm;
                length += (diff * diff + self.inv_period_sq).sqrt();
                prev_norm = cur_norm;
            }

            if length > 0.0 {
                fgdi = 1.0 + (length.ln() + self.ln2) / self.log_denom;
            } else {
                fgdi = 1.0;
            }
        }

        // Hurst exponent and adaptive speed.
        let mut hurst = 2.0 - fgdi;
        if hurst < 0.01 {
            hurst = 0.01;
        }

        let trail_dim = 1.0 / hurst;
        let beta = trail_dim / 2.0;
        let speed_f = (period as f64 * beta).round();
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

        // Deviation in scaled space over last *period* closes.
        let dev_start = n_closes - period;
        let frasma2_scaled = p * frasma2_val;
        let mut sq_sum = 0.0;

        for k in dev_start..n_closes {
            let res = p * self.closes[k] - frasma2_scaled;
            sq_sum += res * res;
        }

        let deviation = (sq_sum / period as f64).sqrt();

        // FBM band offset: 2 * sigma^(2H).
        let two_h = 2.0 * hurst;
        let band_offset = 2.0 * deviation.powf(two_h);
        let ub = (frasma2_scaled + band_offset) / p;
        let lb = (frasma2_scaled - band_offset) / p;

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

impl Indicator for FractionalBands {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::FractionalBands,
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
    use crate::indicators::jean_philippe_poton::fractional_bands::testdata::testdata;

    const EPSILON: f64 = 1e-11;

    fn check_value(exp: f64, act: f64, label: &str, i: usize) {
        if exp.is_nan() {
            assert!(act.is_nan(), "[{}] {}: expected NaN, got {}", i, label, act);
        } else {
            assert!((exp - act).abs() < EPSILON, "[{}] {}: expected {}, got {}", i, label, exp, act);
        }
    }

    fn run_test(period: usize, price_scale: f64, exp_frasma2: &[f64], exp_upper: &[f64], exp_lower: &[f64]) {
        let mut ind = FractionalBands::new(&FractionalBandsParams { period, price_scale, ..Default::default() }).unwrap();
        let input = testdata::test_input();

        for i in 0..input.len() {
            let (frasma2, upper, lower) = ind.update_all(input[i]);
            check_value(exp_frasma2[i], frasma2, "frasma2", i);
            check_value(exp_upper[i], upper, "upper", i);
            check_value(exp_lower[i], lower, "lower", i);
        }
    }

    #[test]
    fn test_p5_s1() {
        run_test(5, 1.0, &testdata::expected_frasma2_p5_s1(), &testdata::expected_upper_p5_s1(), &testdata::expected_lower_p5_s1());
    }

    #[test]
    fn test_p10_s1() {
        run_test(10, 1.0, &testdata::expected_frasma2_p10_s1(), &testdata::expected_upper_p10_s1(), &testdata::expected_lower_p10_s1());
    }

    #[test]
    fn test_p20_s1() {
        run_test(20, 1.0, &testdata::expected_frasma2_p20_s1(), &testdata::expected_upper_p20_s1(), &testdata::expected_lower_p20_s1());
    }

    #[test]
    fn test_p30_s1() {
        run_test(30, 1.0, &testdata::expected_frasma2_p30_s1(), &testdata::expected_upper_p30_s1(), &testdata::expected_lower_p30_s1());
    }

    #[test]
    fn test_p50_s1() {
        run_test(50, 1.0, &testdata::expected_frasma2_p50_s1(), &testdata::expected_upper_p50_s1(), &testdata::expected_lower_p50_s1());
    }

    #[test]
    fn test_p80_s1() {
        run_test(80, 1.0, &testdata::expected_frasma2_p80_s1(), &testdata::expected_upper_p80_s1(), &testdata::expected_lower_p80_s1());
    }

    #[test]
    fn test_p30_s100() {
        run_test(30, 100.0, &testdata::expected_frasma2_p30_s100(), &testdata::expected_upper_p30_s100(), &testdata::expected_lower_p30_s100());
    }

    #[test]
    fn test_p30_s10000() {
        run_test(30, 10000.0, &testdata::expected_frasma2_p30_s10000(), &testdata::expected_upper_p30_s10000(), &testdata::expected_lower_p30_s10000());
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut ind = FractionalBands::new(&FractionalBandsParams { period: 30, price_scale: 1.0, ..Default::default() }).unwrap();

        for i in 0..30 {
            ind.update(input[i]);
            assert!(!ind.is_primed(), "expected not primed at index {}", i);
        }
        ind.update(input[30]);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_nan_passthrough() {
        let mut ind = FractionalBands::new(&FractionalBandsParams { period: 5, price_scale: 1.0, ..Default::default() }).unwrap();
        let (frasma2, upper, lower) = ind.update_all(f64::NAN);
        assert!(frasma2.is_nan());
        assert!(upper.is_nan());
        assert!(lower.is_nan());
    }

    #[test]
    fn test_invalid_period() {
        let result = FractionalBands::new(&FractionalBandsParams { period: 1, price_scale: 1.0, ..Default::default() });
        assert!(result.is_err());
    }

    #[test]
    fn test_invalid_price_scale() {
        let result = FractionalBands::new(&FractionalBandsParams { period: 30, price_scale: 0.0, ..Default::default() });
        assert!(result.is_err());
    }

    #[test]
    fn test_metadata() {
        let ind = FractionalBands::new(&FractionalBandsParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::FractionalBands);
        assert!(meta.mnemonic.contains("fctban(30"));
    }
}
