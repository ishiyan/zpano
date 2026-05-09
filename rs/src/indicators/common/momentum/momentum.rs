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

/// Parameters to create an instance of the momentum indicator.
pub struct MomentumParams {
    /// The length (number of time periods) defining the absolute difference
    /// between today's sample and the sample `length` periods ago.
    /// Must be greater than 0.
    pub length: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for MomentumParams {
    fn default() -> Self {
        Self {
            length: 10,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the momentum indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MomentumOutput {
    /// The scalar value of the momentum.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the momentum (MOM).
///
/// MOMi = Pi - Pi-l
///
/// where l is the length.
///
/// The indicator is not primed during the first l updates.
pub struct Momentum {
    line: LineIndicator,
    window: Vec<f64>,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    primed: bool,
}

impl Momentum {
    /// Creates a new Momentum from the given parameters.
    pub fn new(params: &MomentumParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err("invalid momentum parameters: length should be positive".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("mom({}{})", params.length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Momentum {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let window_length = params.length + 1;

        Ok(Self {
            line,
            window: vec![0.0; window_length],
            window_length,
            window_count: 0,
            last_index: params.length,
            primed: false,
        })
    }

    /// Core update logic. Returns the momentum value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        if self.primed {
            for i in 0..self.last_index {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = sample;

            return sample - self.window[0];
        }

        self.window[self.window_count] = sample;
        self.window_count += 1;

        if self.window_length == self.window_count {
            self.primed = true;

            return sample - self.window[0];
        }

        f64::NAN
    }
}

impl Indicator for Momentum {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::Momentum,
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
    use crate::entities::bar_component::BarComponent;
    use crate::entities::quote_component::QuoteComponent;
    use crate::entities::trade_component::TradeComponent;
    use crate::indicators::core::outputs::shape::Shape;
    fn create_momentum(length: usize) -> Momentum {
        Momentum::new(&MomentumParams { length, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_update_length_14() {
        let mut mom = create_momentum(14);
        let input = testdata::test_input();

        // First 13 updates (index 0..12) produce NaN (not yet primed).
        for i in 0..13 {
            assert!(mom.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        // From index 13 onward, the indicator is primed.
        // We check specific indices from the TA-Lib reference.
        for i in 13..input.len() {
            let act = mom.update(input[i]);

            match i {
                14 => assert!((act - (-0.50)).abs() < 1e-13, "[14] expected -0.50, got {}", act),
                15 => assert!((act - (-2.00)).abs() < 1e-13, "[15] expected -2.00, got {}", act),
                16 => assert!((act - (-5.22)).abs() < 1e-13, "[16] expected -5.22, got {}", act),
                251 => assert!((act - (-1.13)).abs() < 1e-13, "[251] expected -1.13, got {}", act),
                _ => {}
            }
        }

        assert!(mom.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_entity() {
        let length = 2;
        let inp = 3.0_f64;
        let exp = 3.0_f64; // mom = 3.0 - 0.0 = 3.0
        let time = 1617235200;

        // scalar
        let mut mom = create_momentum(length);
        mom.update(0.0);
        mom.update(0.0);
        let out = mom.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!((s.value - exp).abs() < 1e-13);

        // bar
        let mut mom = create_momentum(length);
        mom.update(0.0);
        mom.update(0.0);
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = mom.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp).abs() < 1e-13);

        // quote
        let mut mom = create_momentum(length);
        mom.update(0.0);
        mom.update(0.0);
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = mom.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp).abs() < 1e-13);

        // trade
        let mut mom = create_momentum(length);
        mom.update(0.0);
        mom.update(0.0);
        let trade = Trade::new(time, inp, 0.0);
        let out = mom.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp).abs() < 1e-13);
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();

        for &length in &[1_usize, 2, 3, 5, 10] {
            let mut mom = create_momentum(length);
            assert!(!mom.is_primed());

            for i in 0..length {
                mom.update(input[i]);
                assert!(!mom.is_primed(), "length={}, [{}] should not be primed", length, i);
            }

            for i in length..input.len() {
                mom.update(input[i]);
                assert!(mom.is_primed(), "length={}, [{}] should be primed", length, i);
            }
        }
    }

    #[test]
    fn test_metadata() {
        let mom = create_momentum(5);
        let m = mom.metadata();
        assert_eq!(m.identifier, Identifier::Momentum);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, MomentumOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "mom(5)");
        assert_eq!(m.outputs[0].description, "Momentum mom(5)");
    }

    #[test]
    fn test_new_invalid_length() {
        let r = Momentum::new(&MomentumParams { length: 0, ..Default::default() });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid momentum parameters: length should be positive");
    }

    #[test]
    fn test_mnemonic_components() {
        // all defaults -> no component suffix
        let mom = create_momentum(5);
        assert_eq!(mom.line.mnemonic, "mom(5)");

        // bar component set
        let mom = Momentum::new(&MomentumParams {
            length: 5, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(mom.line.mnemonic, "mom(5, hl/2)");

        // only quote component set
        let mom = Momentum::new(&MomentumParams {
            length: 5, quote_component: Some(QuoteComponent::Bid), ..Default::default()
        }).unwrap();
        assert_eq!(mom.line.mnemonic, "mom(5, b)");

        // only trade component set
        let mom = Momentum::new(&MomentumParams {
            length: 5, trade_component: Some(TradeComponent::Volume), ..Default::default()
        }).unwrap();
        assert_eq!(mom.line.mnemonic, "mom(5, v)");

        // bar non-default, trade non-default
        let mom = Momentum::new(&MomentumParams {
            length: 5,
            bar_component: Some(BarComponent::High),
            quote_component: None,
            trade_component: Some(TradeComponent::Volume),
        }).unwrap();
        assert_eq!(mom.line.mnemonic, "mom(5, h, v)");
    }
}
