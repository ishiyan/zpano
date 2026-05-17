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

/// Parameters to create an instance of the fractal graph dimension index indicator.
pub struct FractalGraphDimensionIndexParams {
    /// The lookback period N. Must be greater than 1.
    pub period: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for FractalGraphDimensionIndexParams {
    fn default() -> Self {
        Self {
            period: 30,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the fractal graph dimension index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum FractalGraphDimensionIndexOutput {
    /// The fractal graph dimension value.
    Fgdi = 1,
    /// The upper band (fgdi + stddev).
    Upper = 2,
    /// The lower band (fgdi - stddev).
    Lower = 3,
    /// The standard deviation of the dimension estimate.
    Stddev = 4,
    /// The lower/upper band pair.
    Band = 5,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Fractal Graph Dimension Index (FGDI).
///
/// This is Poton's corrected and enhanced version of the Fractal Dimension
/// Index (FDI). It fixes loop boundary and denominator bugs in the original
/// and adds standard deviation bands around the estimated dimension.
///
/// The indicator is not primed during the first `period - 1` updates.
pub struct FractalGraphDimensionIndex {
    line: LineIndicator,
    window: Vec<f64>,
    period: usize,
    n_minus_1: usize,
    window_count: usize,
    primed: bool,
    log_2n1: f64,
    ln2: f64,
    inv_n_sq: f64,
    fgdi: f64,
    upper: f64,
    lower: f64,
    stddev_val: f64,
}

impl FractalGraphDimensionIndex {
    /// Creates a new FractalGraphDimensionIndex from the given parameters.
    pub fn new(params: &FractalGraphDimensionIndexParams) -> Result<Self, String> {
        if params.period < 2 {
            return Err("invalid fractal graph dimension index parameters: period should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("fgdi({}{})", params.period, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Fractal graph dimension index {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);
        let period_f = params.period as f64;
        let n_minus_1 = params.period - 1;
        let n_minus_1_f = n_minus_1 as f64;

        Ok(Self {
            line,
            window: vec![0.0; params.period],
            period: params.period,
            n_minus_1,
            window_count: 0,
            primed: false,
            log_2n1: (2.0 * n_minus_1_f).ln(),
            ln2: 2.0_f64.ln(),
            inv_n_sq: 1.0 / (period_f * period_f),
            fgdi: f64::NAN,
            upper: f64::NAN,
            lower: f64::NAN,
            stddev_val: f64::NAN,
        })
    }

    /// Core update logic. Returns the FGDI value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let period = self.period;
        let n_minus_1 = self.n_minus_1;

        if self.primed {
            for i in 0..n_minus_1 {
                self.window[i] = self.window[i + 1];
            }
            self.window[n_minus_1] = sample;
        } else {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if self.window_count < period {
                return f64::NAN;
            }

            self.primed = true;
        }

        // Find min/max for normalization.
        let mut price_max = self.window[0];
        let mut price_min = self.window[0];

        for k in 1..period {
            if self.window[k] > price_max { price_max = self.window[k]; }
            if self.window[k] < price_min { price_min = self.window[k]; }
        }

        let price_range = price_max - price_min;
        if price_range < 1e-10 {
            self.fgdi = 1.0;
            self.stddev_val = 0.0;
            self.upper = 1.0;
            self.lower = 1.0;
            return 1.0;
        }

        // Normalize and compute path segments.
        let mut prior_norm = (self.window[0] - price_min) / price_range;
        let mut length = 0.0;
        let mut segments = vec![0.0; n_minus_1];

        for k in 1..period {
            let curr_norm = (self.window[k] - price_min) / price_range;
            let diff = curr_norm - prior_norm;
            let seg = (diff * diff + self.inv_n_sq).sqrt();
            segments[k - 1] = seg;
            length += seg;
            prior_norm = curr_norm;
        }

        // FGDI = 1 + (ln(L) + ln(2)) / ln(2*(N-1))
        let fgdi_val = 1.0 + (length.ln() + self.ln2) / self.log_2n1;

        // Standard deviation of the estimate.
        let mean_seg = length / n_minus_1 as f64;
        let mut sum_sq = 0.0;

        for k in 0..n_minus_1 {
            let d = segments[k] - mean_seg;
            sum_sq += d * d;
        }

        let variance = sum_sq / (length * length * self.log_2n1 * self.log_2n1);
        let stddev = variance.sqrt();

        self.fgdi = fgdi_val;
        self.upper = fgdi_val + stddev;
        self.lower = fgdi_val - stddev;
        self.stddev_val = stddev;

        fgdi_val
    }

    /// Updates and returns all four outputs: (fgdi, upper, lower, stddev).
    pub fn update_all(&mut self, sample: f64) -> (f64, f64, f64, f64) {
        let fgdi_val = self.update(sample);
        (fgdi_val, self.upper, self.lower, self.stddev_val)
    }

    /// Returns the last computed FGDI value.
    pub fn fgdi_value(&self) -> f64 { self.fgdi }

    /// Returns the last computed upper band value.
    pub fn upper_value(&self) -> f64 { self.upper }

    /// Returns the last computed lower band value.
    pub fn lower_value(&self) -> f64 { self.lower }

    /// Returns the last computed stddev value.
    pub fn stddev_value(&self) -> f64 { self.stddev_val }
}

impl Indicator for FractalGraphDimensionIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::FractalGraphDimensionIndex,
            &self.line.mnemonic,
            &self.line.description,
            &[
                OutputText { mnemonic: self.line.mnemonic.clone(), description: self.line.description.clone() },
                OutputText { mnemonic: format!("{} upper", self.line.mnemonic), description: format!("{} Upper", self.line.description) },
                OutputText { mnemonic: format!("{} lower", self.line.mnemonic), description: format!("{} Lower", self.line.description) },
                OutputText { mnemonic: format!("{} stddev", self.line.mnemonic), description: format!("{} Stddev", self.line.description) },
                OutputText { mnemonic: format!("{} band", self.line.mnemonic), description: format!("{} Band", self.line.description) },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (fgdi, upper, lower, stddev) = self.update_all(sample.value);
        let t = sample.time;

        let band: Box<dyn std::any::Any> = if lower.is_nan() || upper.is_nan() {
            Box::new(Band::empty(t))
        } else {
            Box::new(Band::new(t, lower, upper))
        };

        vec![
            Box::new(Scalar::new(t, fgdi)),
            Box::new(Scalar::new(t, upper)),
            Box::new(Scalar::new(t, lower)),
            Box::new(Scalar::new(t, stddev)),
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
    use crate::indicators::jean_philippe_poton::fractal_graph_dimension_index::testdata::testdata;

    const EPSILON: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() < EPSILON
    }

    fn check(i: usize, label: &str, exp: f64, act: f64) {
        if exp.is_nan() {
            assert!(act.is_nan(), "[{}] {} expected NaN, got {}", i, label, act);
        } else {
            assert!(almost_equal(act, exp), "[{}] {} expected {}, got {}", i, label, exp, act);
        }
    }

    fn run_test(period: usize, exp_fgdi: &[f64], exp_upper: &[f64], exp_lower: &[f64], exp_stddev: &[f64]) {
        let mut ind = FractalGraphDimensionIndex::new(&FractalGraphDimensionIndexParams { period, ..Default::default() }).unwrap();
        let input = testdata::test_input();

        for (i, &v) in input.iter().enumerate() {
            let (fgdi, upper, lower, stddev) = ind.update_all(v);
            check(i, "fgdi", exp_fgdi[i], fgdi);
            check(i, "upper", exp_upper[i], upper);
            check(i, "lower", exp_lower[i], lower);
            check(i, "stddev", exp_stddev[i], stddev);
        }
    }

    #[test]
    fn test_fgdi_period_5() {
        run_test(5, &testdata::expected_fgdi_p5(), &testdata::expected_upper_p5(), &testdata::expected_lower_p5(), &testdata::expected_stddev_p5());
    }

    #[test]
    fn test_fgdi_period_10() {
        run_test(10, &testdata::expected_fgdi_p10(), &testdata::expected_upper_p10(), &testdata::expected_lower_p10(), &testdata::expected_stddev_p10());
    }

    #[test]
    fn test_fgdi_period_15() {
        run_test(15, &testdata::expected_fgdi_p15(), &testdata::expected_upper_p15(), &testdata::expected_lower_p15(), &testdata::expected_stddev_p15());
    }

    #[test]
    fn test_fgdi_period_20() {
        run_test(20, &testdata::expected_fgdi_p20(), &testdata::expected_upper_p20(), &testdata::expected_lower_p20(), &testdata::expected_stddev_p20());
    }

    #[test]
    fn test_fgdi_period_30() {
        run_test(30, &testdata::expected_fgdi_p30(), &testdata::expected_upper_p30(), &testdata::expected_lower_p30(), &testdata::expected_stddev_p30());
    }

    #[test]
    fn test_fgdi_period_50() {
        run_test(50, &testdata::expected_fgdi_p50(), &testdata::expected_upper_p50(), &testdata::expected_lower_p50(), &testdata::expected_stddev_p50());
    }

    #[test]
    fn test_fgdi_period_80() {
        run_test(80, &testdata::expected_fgdi_p80(), &testdata::expected_upper_p80(), &testdata::expected_lower_p80(), &testdata::expected_stddev_p80());
    }

    #[test]
    fn test_fgdi_period_120() {
        run_test(120, &testdata::expected_fgdi_p120(), &testdata::expected_upper_p120(), &testdata::expected_lower_p120(), &testdata::expected_stddev_p120());
    }

    #[test]
    fn test_fgdi_is_primed() {
        let mut ind = FractalGraphDimensionIndex::new(&FractalGraphDimensionIndexParams { period: 30, ..Default::default() }).unwrap();
        let input = testdata::test_input();

        for i in 0..29 {
            ind.update(input[i]);
            assert!(!ind.is_primed());
        }
        ind.update(input[29]);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_fgdi_nan_passthrough() {
        let mut ind = FractalGraphDimensionIndex::new(&FractalGraphDimensionIndexParams { period: 5, ..Default::default() }).unwrap();
        let (fgdi, upper, lower, stddev) = ind.update_all(f64::NAN);
        assert!(fgdi.is_nan());
        assert!(upper.is_nan());
        assert!(lower.is_nan());
        assert!(stddev.is_nan());
    }

    #[test]
    fn test_fgdi_invalid_period() {
        let result = FractalGraphDimensionIndex::new(&FractalGraphDimensionIndexParams { period: 1, ..Default::default() });
        assert!(result.is_err());
    }

    #[test]
    fn test_fgdi_metadata() {
        let ind = FractalGraphDimensionIndex::new(&FractalGraphDimensionIndexParams { period: 30, ..Default::default() }).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::FractalGraphDimensionIndex);
    }
}
