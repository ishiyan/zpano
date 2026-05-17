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

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the hurst difference indicator.
pub struct HurstDifferenceParams {
    /// The lookback period N. Must be greater than 1.
    pub period: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for HurstDifferenceParams {
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

/// Enumerates the outputs of the hurst difference indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum HurstDifferenceOutput {
    /// The first difference of FGDI.
    HurstDiff = 0,
    /// The raw FGDI value.
    Fgdi = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Hurst Difference (first difference of the corrected FGDI).
///
/// Positive values indicate rising volatility (potential trade entry);
/// negative values indicate declining volatility.
///
/// The FGDI is computed using the corrected FGDI formula with (period-1)
/// segments and denominator ln(2*(period-1)).
///
/// The indicator is not primed during the first `period` updates.
/// The hurst_diff output requires one additional update beyond FGDI priming.
pub struct HurstDifference {
    line: LineIndicator,
    window: Vec<f64>,
    period: usize,
    window_count: usize,
    primed: bool,
    log_2pm1: f64,
    ln2: f64,
    inv_n_sq: f64,
    prev_fgdi: f64,
    last_fgdi: f64,
}

impl HurstDifference {
    /// Creates a new HurstDifference from the given parameters.
    pub fn new(params: &HurstDifferenceParams) -> Result<Self, String> {
        if params.period < 2 {
            return Err("invalid hurst difference parameters: period should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("hurdif({}{})", params.period, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Hurst difference {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);
        let period_f = params.period as f64;
        let n_minus_1 = params.period - 1;
        let n_minus_1_f = n_minus_1 as f64;

        Ok(Self {
            line,
            window: vec![0.0; params.period + 1],
            period: params.period,
            window_count: 0,
            primed: false,
            log_2pm1: (2.0 * n_minus_1_f).ln(),
            ln2: 2.0_f64.ln(),
            inv_n_sq: 1.0 / (period_f * period_f),
            prev_fgdi: f64::NAN,
            last_fgdi: f64::NAN,
        })
    }

    /// Core update logic. Returns the hurst_diff value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let period = self.period;

        if self.primed {
            for i in 0..period {
                self.window[i] = self.window[i + 1];
            }
            self.window[period] = sample;
        } else {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if self.window_count <= period {
                return f64::NAN;
            }

            self.primed = true;
        }

        // Use the last `period` elements of the window (indices 1..=period).
        // Find min/max for normalization.
        let mut price_max = self.window[1];
        let mut price_min = self.window[1];

        for k in 2..=period {
            if self.window[k] > price_max { price_max = self.window[k]; }
            if self.window[k] < price_min { price_min = self.window[k]; }
        }

        let price_range = price_max - price_min;

        let fgdi_val;

        if price_range <= 0.0 {
            fgdi_val = 0.0;
        } else {
            // Normalize and compute path length.
            let mut prior_norm = (self.window[1] - price_min) / price_range;
            let mut length = 0.0;

            for k in 2..=period {
                let curr_norm = (self.window[k] - price_min) / price_range;
                let diff = curr_norm - prior_norm;
                length += (diff * diff + self.inv_n_sq).sqrt();
                prior_norm = curr_norm;
            }

            if length > 0.0 {
                fgdi_val = 1.0 + (length.ln() + self.ln2) / self.log_2pm1;
            } else {
                fgdi_val = 0.0;
            }
        }

        // First difference.
        let hurst_diff = if self.prev_fgdi.is_nan() {
            f64::NAN
        } else {
            fgdi_val - self.prev_fgdi
        };

        self.prev_fgdi = fgdi_val;
        self.last_fgdi = fgdi_val;

        hurst_diff
    }

    /// Updates and returns both outputs: (hurst_diff, fgdi).
    pub fn update_all(&mut self, sample: f64) -> (f64, f64) {
        let hurst_diff = self.update(sample);
        (hurst_diff, self.last_fgdi)
    }

    /// Returns the last computed FGDI value.
    pub fn fgdi_value(&self) -> f64 { self.last_fgdi }
}

impl Indicator for HurstDifference {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::HurstDifference,
            &self.line.mnemonic,
            &self.line.description,
            &[
                OutputText { mnemonic: self.line.mnemonic.clone(), description: self.line.description.clone() },
                OutputText { mnemonic: format!("{} fgdi", self.line.mnemonic), description: format!("{} FGDI", self.line.description) },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.line.bar_func)(sample);
        let scalar = Scalar::new(sample.time, v);
        self.update_scalar(&scalar)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.line.quote_func)(sample);
        let scalar = Scalar::new(sample.time, v);
        self.update_scalar(&scalar)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.line.trade_func)(sample);
        let scalar = Scalar::new(sample.time, v);
        self.update_scalar(&scalar)
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::jean_philippe_poton::hurst_difference::testdata::testdata::*;

    const EPSILON: f64 = 1e-13;

    fn check(label: &str, index: usize, exp: f64, act: f64) {
        if exp.is_nan() {
            assert!(act.is_nan(), "[{}] {}: expected NaN, got {}", index, label, act);
        } else {
            assert!(
                (exp - act).abs() < EPSILON,
                "[{}] {}: expected {}, got {}",
                index, label, exp, act
            );
        }
    }

    fn run_test(period: usize, exp_fgdi: &[f64], exp_hdiff: &[f64]) {
        let mut ind = HurstDifference::new(&HurstDifferenceParams {
            period,
            ..Default::default()
        }).unwrap();

        let input = test_input();
        for i in 0..input.len() {
            let (hdiff, fgdi) = ind.update_all(input[i]);
            check("fgdi", i, exp_fgdi[i], fgdi);
            check("hdiff", i, exp_hdiff[i], hdiff);
        }
    }

    #[test]
    fn test_update_period_5() {
        run_test(5, &expected_fgdi_p5(), &expected_hdiff_p5());
    }

    #[test]
    fn test_update_period_10() {
        run_test(10, &expected_fgdi_p10(), &expected_hdiff_p10());
    }

    #[test]
    fn test_update_period_15() {
        run_test(15, &expected_fgdi_p15(), &expected_hdiff_p15());
    }

    #[test]
    fn test_update_period_20() {
        run_test(20, &expected_fgdi_p20(), &expected_hdiff_p20());
    }

    #[test]
    fn test_update_period_30() {
        run_test(30, &expected_fgdi_p30(), &expected_hdiff_p30());
    }

    #[test]
    fn test_update_period_50() {
        run_test(50, &expected_fgdi_p50(), &expected_hdiff_p50());
    }

    #[test]
    fn test_update_period_80() {
        run_test(80, &expected_fgdi_p80(), &expected_hdiff_p80());
    }

    #[test]
    fn test_update_period_120() {
        run_test(120, &expected_fgdi_p120(), &expected_hdiff_p120());
    }

    #[test]
    fn test_is_primed() {
        let mut ind = HurstDifference::new(&HurstDifferenceParams {
            period: 30,
            ..Default::default()
        }).unwrap();

        let input = test_input();
        for i in 0..30 {
            ind.update(input[i]);
            assert!(!ind.is_primed(), "should not be primed at index {}", i);
        }
        ind.update(input[30]);
        assert!(ind.is_primed(), "should be primed after 31 samples");
    }

    #[test]
    fn test_nan_passthrough() {
        let mut ind = HurstDifference::new(&HurstDifferenceParams {
            period: 5,
            ..Default::default()
        }).unwrap();

        let (hdiff, fgdi) = ind.update_all(f64::NAN);
        assert!(hdiff.is_nan());
        assert!(fgdi.is_nan());
    }

    #[test]
    fn test_invalid_period() {
        let result = HurstDifference::new(&HurstDifferenceParams {
            period: 1,
            ..Default::default()
        });
        assert!(result.is_err());
    }

    #[test]
    fn test_metadata() {
        let ind = HurstDifference::new(&HurstDifferenceParams {
            period: 30,
            ..Default::default()
        }).unwrap();

        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::HurstDifference);
        assert!(meta.mnemonic.contains("hurdif(30)"));
    }
}
