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

/// Parameters to create an instance of the rate of change percent indicator.
pub struct RateOfChangePercentParams {
    /// The length (number of time periods).
    /// Must be greater than 0.
    pub length: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for RateOfChangePercentParams {
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

/// Enumerates the outputs of the rate of change percent indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum RateOfChangePercentOutput {
    /// The scalar value of the rate of change percent.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Rate of Change Percent (ROCP).
///
/// ROCPi = (Pi - Pi-l) / Pi-l
///
/// where l is the length.
///
/// The indicator is not primed during the first l updates.
pub struct RateOfChangePercent {
    line: LineIndicator,
    window: Vec<f64>,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    primed: bool,
}

impl RateOfChangePercent {
    /// Creates a new RateOfChangePercent from the given parameters.
    pub fn new(params: &RateOfChangePercentParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err("invalid rate of change percent parameters: length should be positive".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("rocp({}{})", params.length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Rate of Change Percent {}", mnemonic);

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

    /// Core update logic. Returns the rate of change percent value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        const EPSILON: f64 = 1e-13;

        if self.primed {
            for i in 0..self.last_index {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = sample;
            let previous = self.window[0];

            if previous.abs() > EPSILON {
                return sample / previous - 1.0;
            }

            return 0.0;
        }

        self.window[self.window_count] = sample;
        self.window_count += 1;

        if self.window_length == self.window_count {
            self.primed = true;
            let previous = self.window[0];

            if previous.abs() > EPSILON {
                return sample / previous - 1.0;
            }

            return 0.0;
        }

        f64::NAN
    }
}

impl Indicator for RateOfChangePercent {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::RateOfChangePercent,
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
    fn create_rocp(length: usize) -> RateOfChangePercent {
        RateOfChangePercent::new(&RateOfChangePercentParams { length, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_update_length_14() {
        let mut rocp = create_rocp(14);
        let input = testdata::test_input();

        for i in 0..13 {
            assert!(rocp.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        // ROCP = ROC / 100
        for i in 13..input.len() {
            let act = rocp.update(input[i]);

            match i {
                14 => assert!((act - (-0.00546)).abs() < 1e-4, "[14] expected -0.00546, got {}", act),
                15 => assert!((act - (-0.02109)).abs() < 1e-4, "[15] expected -0.02109, got {}", act),
                16 => assert!((act - (-0.0553)).abs() < 1e-4, "[16] expected -0.0553, got {}", act),
                251 => assert!((act - (-0.010367)).abs() < 1e-4, "[251] expected -0.010367, got {}", act),
                _ => {}
            }
        }

        assert!(rocp.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_entity() {
        let length = 2;
        let inp = 3.0_f64;
        let exp = 0.0_f64; // rocp = (3-3)/3 = 0
        let time = 1617235200;

        // scalar
        let mut rocp = create_rocp(length);
        rocp.update(inp);
        rocp.update(inp);
        let out = rocp.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!((s.value - exp).abs() < 1e-13);

        // bar
        let mut rocp = create_rocp(length);
        rocp.update(inp);
        rocp.update(inp);
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = rocp.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp).abs() < 1e-13);

        // quote
        let mut rocp = create_rocp(length);
        rocp.update(inp);
        rocp.update(inp);
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = rocp.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp).abs() < 1e-13);

        // trade
        let mut rocp = create_rocp(length);
        rocp.update(inp);
        rocp.update(inp);
        let trade = Trade::new(time, inp, 0.0);
        let out = rocp.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp).abs() < 1e-13);
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();

        for &length in &[1_usize, 2, 3, 5, 10] {
            let mut rocp = create_rocp(length);
            assert!(!rocp.is_primed());

            for i in 0..length {
                rocp.update(input[i]);
                assert!(!rocp.is_primed(), "length={}, [{}] should not be primed", length, i);
            }

            for i in length..input.len() {
                rocp.update(input[i]);
                assert!(rocp.is_primed(), "length={}, [{}] should be primed", length, i);
            }
        }
    }

    #[test]
    fn test_metadata() {
        let rocp = create_rocp(5);
        let m = rocp.metadata();
        assert_eq!(m.identifier, Identifier::RateOfChangePercent);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, RateOfChangePercentOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "rocp(5)");
        assert_eq!(m.outputs[0].description, "Rate of Change Percent rocp(5)");
    }

    #[test]
    fn test_new_invalid_length() {
        let r = RateOfChangePercent::new(&RateOfChangePercentParams { length: 0, ..Default::default() });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid rate of change percent parameters: length should be positive");
    }

    #[test]
    fn test_mnemonic_components() {
        let rocp = create_rocp(5);
        assert_eq!(rocp.line.mnemonic, "rocp(5)");

        let rocp = RateOfChangePercent::new(&RateOfChangePercentParams {
            length: 5, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(rocp.line.mnemonic, "rocp(5, hl/2)");

        let rocp = RateOfChangePercent::new(&RateOfChangePercentParams {
            length: 5, quote_component: Some(QuoteComponent::Bid), ..Default::default()
        }).unwrap();
        assert_eq!(rocp.line.mnemonic, "rocp(5, b)");

        let rocp = RateOfChangePercent::new(&RateOfChangePercentParams {
            length: 5, trade_component: Some(TradeComponent::Volume), ..Default::default()
        }).unwrap();
        assert_eq!(rocp.line.mnemonic, "rocp(5, v)");
    }
}
