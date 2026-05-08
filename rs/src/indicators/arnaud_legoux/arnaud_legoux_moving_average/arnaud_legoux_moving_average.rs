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

/// Parameters to create an instance of the Arnaud Legoux moving average indicator.
pub struct ArnaudLegouxMovingAverageParams {
    /// The window size. Must be >= 1.
    pub window: usize,
    /// The Gaussian sigma parameter. Must be > 0.
    pub sigma: f64,
    /// The offset parameter. Must be between 0 and 1 inclusive.
    pub offset: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for ArnaudLegouxMovingAverageParams {
    fn default() -> Self {
        Self {
            window: 9,
            sigma: 6.0,
            offset: 0.85,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Arnaud Legoux moving average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum ArnaudLegouxMovingAverageOutput {
    /// The scalar value of the moving average.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Format a float like Go's `%g`: integers print without decimal, others use default Display.
fn format_g(v: f64) -> String {
    if v == v.trunc() && v.abs() < 1e15 {
        format!("{}", v as i64)
    } else {
        format!("{}", v)
    }
}

/// Computes the Arnaud Legoux Moving Average (ALMA).
///
/// ALMA is a Gaussian-weighted moving average that reduces lag while maintaining
/// smoothness. It applies a Gaussian bell curve as its kernel, shifted toward
/// recent bars via an adjustable offset parameter.
///
/// The indicator is not primed during the first (window - 1) updates.
pub struct ArnaudLegouxMovingAverage {
    line: LineIndicator,
    weights: Vec<f64>,
    buffer: Vec<f64>,
    window_length: usize,
    buffer_count: usize,
    buffer_index: usize,
    primed: bool,
}

impl ArnaudLegouxMovingAverage {
    /// Creates a new ArnaudLegouxMovingAverage from the given parameters.
    pub fn new(params: &ArnaudLegouxMovingAverageParams) -> Result<Self, String> {
        let window = params.window;
        if window < 1 {
            return Err("invalid Arnaud Legoux moving average parameters: window should be greater than 0".to_string());
        }

        let sigma = params.sigma;
        if sigma <= 0.0 {
            return Err("invalid Arnaud Legoux moving average parameters: sigma should be greater than 0".to_string());
        }

        let offset = params.offset;
        if !(0.0..=1.0).contains(&offset) {
            return Err("invalid Arnaud Legoux moving average parameters: offset should be between 0 and 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "alma({}, {}, {}{})",
            window,
            format_g(sigma),
            format_g(offset),
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Arnaud Legoux moving average {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        // Precompute Gaussian weights.
        let m = offset * (window - 1) as f64;
        let s = window as f64 / sigma;

        let mut weights = vec![0.0; window];
        let mut norm = 0.0;

        for i in 0..window {
            let diff = i as f64 - m;
            let w = (-(diff * diff) / (2.0 * s * s)).exp();
            weights[i] = w;
            norm += w;
        }

        for w in weights.iter_mut() {
            *w /= norm;
        }

        Ok(Self {
            line,
            weights,
            buffer: vec![0.0; window],
            window_length: window,
            buffer_count: 0,
            buffer_index: 0,
            primed: false,
        })
    }

    /// Core update logic. Returns the ALMA value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let window = self.window_length;

        if window == 1 {
            self.primed = true;
            return sample;
        }

        // Fill the circular buffer.
        self.buffer[self.buffer_index] = sample;
        self.buffer_index = (self.buffer_index + 1) % window;

        if !self.primed {
            self.buffer_count += 1;
            if self.buffer_count < window {
                return f64::NAN;
            }

            self.primed = true;
        }

        // Compute weighted sum.
        // Weight[0] applies to oldest sample, weight[N-1] to newest.
        // The oldest sample is at self.buffer_index (circular buffer).
        let mut result = 0.0;
        let mut index = self.buffer_index;

        for i in 0..window {
            result += self.weights[i] * self.buffer[index];
            index = (index + 1) % window;
        }

        result
    }
}

impl Indicator for ArnaudLegouxMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::ArnaudLegouxMovingAverage,
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
    use super::super::testdata::testdata;
    use crate::indicators::core::outputs::shape::Shape;

    fn create_alma(window: usize, sigma: f64, offset: f64) -> ArnaudLegouxMovingAverage {
        ArnaudLegouxMovingAverage::new(&ArnaudLegouxMovingAverageParams {
            window,
            sigma,
            offset,
            ..Default::default()
        })
        .unwrap()
    }

    fn check_alma(alma: &mut ArnaudLegouxMovingAverage, window: usize, input: &[f64], expected: &[f64]) {
        let warmup = if window <= 1 { 0 } else { window - 1 };

        for i in 0..warmup {
            let act = alma.update(input[i]);
            assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
        }

        for i in warmup..input.len() {
            let act = alma.update(input[i]);
            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(
                    (act - expected[i]).abs() < 1e-13,
                    "[{}] expected {}, got {}",
                    i,
                    expected[i],
                    act
                );
            }
        }

        // NaN passthrough.
        assert!(alma.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_w9_s6_o0_85() {
        let input = testdata::test_input();
        let exp = testdata::expected_w9_s6_o0_85();
        let mut alma = create_alma(9, 6.0, 0.85);
        check_alma(&mut alma, 9, &input, &exp);
    }

    #[test]
    fn test_w9_s6_o0_5() {
        let input = testdata::test_input();
        let exp = testdata::expected_w9_s6_o0_5();
        let mut alma = create_alma(9, 6.0, 0.5);
        check_alma(&mut alma, 9, &input, &exp);
    }

    #[test]
    fn test_w10_s6_o0_85() {
        let input = testdata::test_input();
        let exp = testdata::expected_w10_s6_o0_85();
        let mut alma = create_alma(10, 6.0, 0.85);
        check_alma(&mut alma, 10, &input, &exp);
    }

    #[test]
    fn test_w5_s6_o0_9() {
        let input = testdata::test_input();
        let exp = testdata::expected_w5_s6_o0_9();
        let mut alma = create_alma(5, 6.0, 0.9);
        check_alma(&mut alma, 5, &input, &exp);
    }

    #[test]
    fn test_w1_s6_o0_85() {
        let input = testdata::test_input();
        let exp = testdata::expected_w1_s6_o0_85();
        let mut alma = create_alma(1, 6.0, 0.85);
        check_alma(&mut alma, 1, &input, &exp);
    }

    #[test]
    fn test_w3_s6_o0_85() {
        let input = testdata::test_input();
        let exp = testdata::expected_w3_s6_o0_85();
        let mut alma = create_alma(3, 6.0, 0.85);
        check_alma(&mut alma, 3, &input, &exp);
    }

    #[test]
    fn test_w21_s6_o0_85() {
        let input = testdata::test_input();
        let exp = testdata::expected_w21_s6_o0_85();
        let mut alma = create_alma(21, 6.0, 0.85);
        check_alma(&mut alma, 21, &input, &exp);
    }

    #[test]
    fn test_w50_s6_o0_85() {
        let input = testdata::test_input();
        let exp = testdata::expected_w50_s6_o0_85();
        let mut alma = create_alma(50, 6.0, 0.85);
        check_alma(&mut alma, 50, &input, &exp);
    }

    #[test]
    fn test_w9_s6_o0() {
        let input = testdata::test_input();
        let exp = testdata::expected_w9_s6_o0();
        let mut alma = create_alma(9, 6.0, 0.0);
        check_alma(&mut alma, 9, &input, &exp);
    }

    #[test]
    fn test_w9_s6_o1() {
        let input = testdata::test_input();
        let exp = testdata::expected_w9_s6_o1();
        let mut alma = create_alma(9, 6.0, 1.0);
        check_alma(&mut alma, 9, &input, &exp);
    }

    #[test]
    fn test_w9_s2_o0_85() {
        let input = testdata::test_input();
        let exp = testdata::expected_w9_s2_o0_85();
        let mut alma = create_alma(9, 2.0, 0.85);
        check_alma(&mut alma, 9, &input, &exp);
    }

    #[test]
    fn test_w9_s20_o0_85() {
        let input = testdata::test_input();
        let exp = testdata::expected_w9_s20_o0_85();
        let mut alma = create_alma(9, 20.0, 0.85);
        check_alma(&mut alma, 9, &input, &exp);
    }

    #[test]
    fn test_w9_s0_5_o0_85() {
        let input = testdata::test_input();
        let exp = testdata::expected_w9_s0_5_o0_85();
        let mut alma = create_alma(9, 0.5, 0.85);
        check_alma(&mut alma, 9, &input, &exp);
    }

    #[test]
    fn test_w15_s4_o0_7() {
        let input = testdata::test_input();
        let exp = testdata::expected_w15_s4_o0_7();
        let mut alma = create_alma(15, 4.0, 0.7);
        check_alma(&mut alma, 15, &input, &exp);
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut alma = create_alma(9, 6.0, 0.85);

        assert!(!alma.is_primed());
        for i in 0..8 {
            alma.update(input[i]);
            assert!(!alma.is_primed(), "[{}] should not be primed", i);
        }
        for i in 8..input.len() {
            alma.update(input[i]);
            assert!(alma.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_is_primed_window_1() {
        let mut alma = create_alma(1, 6.0, 0.85);
        assert!(!alma.is_primed());
        alma.update(42.0);
        assert!(alma.is_primed());
    }

    #[test]
    fn test_metadata() {
        let alma = create_alma(9, 6.0, 0.85);
        let m = alma.metadata();
        assert_eq!(m.identifier, Identifier::ArnaudLegouxMovingAverage);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, ArnaudLegouxMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "alma(9, 6, 0.85)");
        assert_eq!(m.outputs[0].description, "Arnaud Legoux moving average alma(9, 6, 0.85)");
    }

    #[test]
    fn test_update_entity() {
        let window = 9;
        let time = 1617235200_i64;
        let input = testdata::test_input();
        let exp = testdata::expected_w9_s6_o0_85();

        // scalar
        let mut alma = create_alma(window, 6.0, 0.85);
        for i in 0..8 {
            alma.update(input[i]);
        }
        let out = alma.update_scalar(&Scalar::new(time, input[8]));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!((s.value - exp[8]).abs() < 1e-13);

        // bar
        let mut alma = create_alma(window, 6.0, 0.85);
        for i in 0..8 {
            alma.update(input[i]);
        }
        let bar = Bar::new(time, 0.0, 0.0, 0.0, input[8], 0.0);
        let out = alma.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp[8]).abs() < 1e-13);

        // quote
        let mut alma = create_alma(window, 6.0, 0.85);
        for i in 0..8 {
            alma.update(input[i]);
        }
        let quote = Quote::new(time, input[8], input[8], 0.0, 0.0);
        let out = alma.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp[8]).abs() < 1e-13);

        // trade
        let mut alma = create_alma(window, 6.0, 0.85);
        for i in 0..8 {
            alma.update(input[i]);
        }
        let trade = Trade::new(time, input[8], 0.0);
        let out = alma.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp[8]).abs() < 1e-13);
    }

    #[test]
    fn test_new_invalid_params() {
        let r = ArnaudLegouxMovingAverage::new(&ArnaudLegouxMovingAverageParams {
            window: 0,
            ..Default::default()
        });
        assert!(r.is_err());

        let r = ArnaudLegouxMovingAverage::new(&ArnaudLegouxMovingAverageParams {
            sigma: 0.0,
            ..Default::default()
        });
        assert!(r.is_err());

        let r = ArnaudLegouxMovingAverage::new(&ArnaudLegouxMovingAverageParams {
            sigma: -1.0,
            ..Default::default()
        });
        assert!(r.is_err());

        let r = ArnaudLegouxMovingAverage::new(&ArnaudLegouxMovingAverageParams {
            offset: -0.1,
            ..Default::default()
        });
        assert!(r.is_err());

        let r = ArnaudLegouxMovingAverage::new(&ArnaudLegouxMovingAverageParams {
            offset: 1.1,
            ..Default::default()
        });
        assert!(r.is_err());
    }

    #[test]
    fn test_mnemonic_components() {
        // all defaults -> no component suffix
        let alma = create_alma(9, 6.0, 0.85);
        assert_eq!(alma.line.mnemonic, "alma(9, 6, 0.85)");

        // bar component set to Median
        let alma = ArnaudLegouxMovingAverage::new(&ArnaudLegouxMovingAverageParams {
            window: 9,
            sigma: 6.0,
            offset: 0.85,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        })
        .unwrap();
        assert_eq!(alma.line.mnemonic, "alma(9, 6, 0.85, hl/2)");
        assert_eq!(alma.line.description, "Arnaud Legoux moving average alma(9, 6, 0.85, hl/2)");

        // bar=high, trade=volume
        let alma = ArnaudLegouxMovingAverage::new(&ArnaudLegouxMovingAverageParams {
            window: 9,
            sigma: 6.0,
            offset: 0.85,
            bar_component: Some(BarComponent::High),
            trade_component: Some(TradeComponent::Volume),
            ..Default::default()
        })
        .unwrap();
        assert_eq!(alma.line.mnemonic, "alma(9, 6, 0.85, h, v)");
    }
}
