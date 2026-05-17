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

/// Parameters to create an instance of the fractal dimension index indicator.
pub struct FractalDimensionIndexParams {
    /// The lookback period N. Must be greater than 1.
    pub period: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for FractalDimensionIndexParams {
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

/// Enumerates the outputs of the fractal dimension indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum FractalDimensionIndexOutput {
    /// The scalar value of the fractal dimension.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Fractal Dimension Index (FDI).
///
/// Measures the fractal dimension of a price time series using normalized
/// path length. The indicator is not primed during the first `period` updates.
pub struct FractalDimensionIndex {
    line: LineIndicator,
    window: Vec<f64>,
    period: usize,
    window_count: usize,
    primed: bool,
    log_2n: f64,
    ln2: f64,
    inv_n_sq: f64,
}

impl FractalDimensionIndex {
    /// Creates a new FractalDimensionIndex from the given parameters.
    pub fn new(params: &FractalDimensionIndexParams) -> Result<Self, String> {
        if params.period < 2 {
            return Err("invalid fractal dimension parameters: period should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("fdi({}{})", params.period, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Fractal dimension index {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);
        let period_f = params.period as f64;

        Ok(Self {
            line,
            window: vec![0.0; params.period + 1],
            period: params.period,
            window_count: 0,
            primed: false,
            log_2n: (2.0 * period_f).ln(),
            ln2: 2.0_f64.ln(),
            inv_n_sq: 1.0 / (period_f * period_f),
        })
    }

    /// Core update logic. Returns the FDI value or NaN if not yet primed.
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

        // Find min/max for normalization.
        let mut price_max = self.window[0];
        let mut price_min = self.window[0];

        for k in 1..=period {
            if self.window[k] > price_max { price_max = self.window[k]; }
            if self.window[k] < price_min { price_min = self.window[k]; }
        }

        let price_range = price_max - price_min;
        if price_range < 1e-10 {
            return 1.0;
        }

        // Normalize and compute path length.
        let mut prior_norm = (self.window[0] - price_min) / price_range;
        let mut length = 0.0;

        for k in 1..=period {
            let curr_norm = (self.window[k] - price_min) / price_range;
            let diff = curr_norm - prior_norm;
            length += (diff * diff + self.inv_n_sq).sqrt();
            prior_norm = curr_norm;
        }

        1.0 + (length.ln() + self.ln2) / self.log_2n
    }
}

impl Indicator for FractalDimensionIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::FractalDimensionIndex,
            &self.line.mnemonic,
            &self.line.description,
            &[OutputText {
                mnemonic: self.line.mnemonic.clone(),
                description: self.line.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let sample_value = (self.line.bar_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let sample_value = (self.line.quote_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let sample_value = (self.line.trade_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::jean_philippe_poton::fractal_dimension_index::testdata::testdata;

    const EPSILON: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() < EPSILON
    }

    #[test]
    fn test_fdi_period_5() {
        let mut fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 5, ..Default::default() }).unwrap();
        let input = testdata::test_input();
        let expected = testdata::expected_p5();

        for (i, &v) in input.iter().enumerate() {
            let act = fdi.update(v);
            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(almost_equal(act, expected[i]), "[{}] expected {}, got {}", i, expected[i], act);
            }
        }
    }

    #[test]
    fn test_fdi_period_10() {
        let mut fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 10, ..Default::default() }).unwrap();
        let input = testdata::test_input();
        let expected = testdata::expected_p10();

        for (i, &v) in input.iter().enumerate() {
            let act = fdi.update(v);
            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(almost_equal(act, expected[i]), "[{}] expected {}, got {}", i, expected[i], act);
            }
        }
    }

    #[test]
    fn test_fdi_period_15() {
        let mut fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 15, ..Default::default() }).unwrap();
        let input = testdata::test_input();
        let expected = testdata::expected_p15();

        for (i, &v) in input.iter().enumerate() {
            let act = fdi.update(v);
            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(almost_equal(act, expected[i]), "[{}] expected {}, got {}", i, expected[i], act);
            }
        }
    }

    #[test]
    fn test_fdi_period_20() {
        let mut fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 20, ..Default::default() }).unwrap();
        let input = testdata::test_input();
        let expected = testdata::expected_p20();

        for (i, &v) in input.iter().enumerate() {
            let act = fdi.update(v);
            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(almost_equal(act, expected[i]), "[{}] expected {}, got {}", i, expected[i], act);
            }
        }
    }

    #[test]
    fn test_fdi_period_30() {
        let mut fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 30, ..Default::default() }).unwrap();
        let input = testdata::test_input();
        let expected = testdata::expected_p30();

        for (i, &v) in input.iter().enumerate() {
            let act = fdi.update(v);
            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(almost_equal(act, expected[i]), "[{}] expected {}, got {}", i, expected[i], act);
            }
        }
    }

    #[test]
    fn test_fdi_period_50() {
        let mut fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 50, ..Default::default() }).unwrap();
        let input = testdata::test_input();
        let expected = testdata::expected_p50();

        for (i, &v) in input.iter().enumerate() {
            let act = fdi.update(v);
            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(almost_equal(act, expected[i]), "[{}] expected {}, got {}", i, expected[i], act);
            }
        }
    }

    #[test]
    fn test_fdi_period_80() {
        let mut fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 80, ..Default::default() }).unwrap();
        let input = testdata::test_input();
        let expected = testdata::expected_p80();

        for (i, &v) in input.iter().enumerate() {
            let act = fdi.update(v);
            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(almost_equal(act, expected[i]), "[{}] expected {}, got {}", i, expected[i], act);
            }
        }
    }

    #[test]
    fn test_fdi_period_120() {
        let mut fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 120, ..Default::default() }).unwrap();
        let input = testdata::test_input();
        let expected = testdata::expected_p120();

        for (i, &v) in input.iter().enumerate() {
            let act = fdi.update(v);
            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(almost_equal(act, expected[i]), "[{}] expected {}, got {}", i, expected[i], act);
            }
        }
    }

    #[test]
    fn test_fdi_is_primed() {
        let mut fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 30, ..Default::default() }).unwrap();
        let input = testdata::test_input();

        for i in 0..30 {
            fdi.update(input[i]);
            assert!(!fdi.is_primed());
        }
        fdi.update(input[30]);
        assert!(fdi.is_primed());
    }

    #[test]
    fn test_fdi_nan_passthrough() {
        let mut fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 5, ..Default::default() }).unwrap();
        assert!(fdi.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_fdi_invalid_period() {
        let result = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 1, ..Default::default() });
        assert!(result.is_err());
    }

    #[test]
    fn test_fdi_metadata() {
        let fdi = FractalDimensionIndex::new(&FractalDimensionIndexParams { period: 30, ..Default::default() }).unwrap();
        let meta = fdi.metadata();
        assert_eq!(meta.identifier, Identifier::FractalDimensionIndex);
    }
}
