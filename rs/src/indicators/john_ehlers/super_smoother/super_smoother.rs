use std::f64::consts::{PI, SQRT_2};

use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent};
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

/// Parameters for the Super Smoother indicator.
pub struct SuperSmootherParams {
    /// The shortest cycle period in bars. Must be >= 2. Default is 10.
    pub shortest_cycle_period: i64,
    /// Bar component to extract. `None` means use default (Median).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for SuperSmootherParams {
    fn default() -> Self {
        Self {
            shortest_cycle_period: 10,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Super Smoother indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SuperSmootherOutput {
    /// The scalar value of the super smoother.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' two-pole Super Smoother filter.
///
/// Given the shortest cycle period lambda, the filter attenuates all cycle
/// periods shorter than lambda.
///
///   beta  = sqrt(2) * pi / lambda
///   alpha = exp(-beta)
///   g2    = 2 * alpha * cos(beta)
///   g3    = -alpha^2
///   g1    = (1 - g2 - g3) / 2
///
///   SS_i  = g1*(x_i + x_{i-1}) + g2*SS_{i-1} + g3*SS_{i-2}
///
/// The indicator is not primed during the first 2 updates.
pub struct SuperSmoother {
    line: LineIndicator,
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    count: i64,
    sample_previous: f64,
    filter_previous: f64,
    filter_previous2: f64,
    value: f64,
    primed: bool,
}

impl SuperSmoother {
    /// Creates a new Super Smoother from the supplied parameters.
    pub fn new(params: &SuperSmootherParams) -> Result<Self, String> {
        const INVALID: &str = "invalid super smoother parameters";

        let period = params.shortest_cycle_period;
        if period < 2 {
            return Err(format!("{}: shortest cycle period should be greater than 1", INVALID));
        }

        // Default bar component is Median (not Close).
        let bc = params.bar_component.unwrap_or(BarComponent::Median);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        // Calculate coefficients.
        let beta = SQRT_2 * PI / period as f64;
        let alpha = (-beta).exp();
        let gamma2 = 2.0 * alpha * beta.cos();
        let gamma3 = -alpha * alpha;
        let gamma1 = (1.0 - gamma2 - gamma3) / 2.0;

        let mnemonic = format!("ss({}{})", period, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Super Smoother {}", mnemonic);
        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            coeff1: gamma1,
            coeff2: gamma2,
            coeff3: gamma3,
            count: 0,
            sample_previous: 0.0,
            filter_previous: 0.0,
            filter_previous2: 0.0,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Core update logic. Returns the filter value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        if self.primed {
            let filter = self.coeff1 * (sample + self.sample_previous)
                + self.coeff2 * self.filter_previous
                + self.coeff3 * self.filter_previous2;
            self.value = filter;
            self.sample_previous = sample;
            self.filter_previous2 = self.filter_previous;
            self.filter_previous = filter;
            return self.value;
        }

        self.count += 1;

        if self.count == 1 {
            self.sample_previous = sample;
            self.filter_previous = sample;
            self.filter_previous2 = sample;
        }

        let filter = self.coeff1 * (sample + self.sample_previous)
            + self.coeff2 * self.filter_previous
            + self.coeff3 * self.filter_previous2;

        if self.count == 3 {
            self.primed = true;
            self.value = filter;
        }

        self.sample_previous = sample;
        self.filter_previous2 = self.filter_previous;
        self.filter_previous = filter;

        self.value
    }
}

impl Indicator for SuperSmoother {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::SuperSmoother,
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
    use crate::indicators::core::outputs::shape::Shape;
    fn create_ss(period: i64) -> SuperSmoother {
        SuperSmoother::new(&SuperSmootherParams {
            shortest_cycle_period: period,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_update() {
        const SKIP_ROWS: usize = 60;
        const TOLERANCE: f64 = 0.5;

        let input = testdata::test_input();
        let expected = testdata::test_expected();
        let mut ss = create_ss(10);

        for i in 0..input.len() {
            let act = ss.update(input[i]);

            if i < 2 {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
                continue;
            }

            if i < SKIP_ROWS {
                continue;
            }

            assert!(
                (act - expected[i]).abs() <= TOLERANCE,
                "[{}] expected {}, got {}", i, expected[i], act
            );
        }

        // NaN passthrough
        assert!(ss.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut ss = create_ss(10);

        assert!(!ss.is_primed());

        for i in 0..2 {
            ss.update(input[i]);
            assert!(!ss.is_primed(), "[{}] should not be primed", i);
        }

        ss.update(input[2]);
        assert!(ss.is_primed(), "[2] should be primed");
    }

    #[test]
    fn test_update_entity() {
        let inp = 100.0_f64;
        let time = 1617235200_i64;
        let mut ss = create_ss(10);

        // Prime
        ss.update(inp);
        ss.update(inp);
        ss.update(inp);

        // Scalar
        let mut ss2 = create_ss(10);
        ss2.update(inp);
        ss2.update(inp);
        ss2.update(inp);
        let out = ss2.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!(!s.value.is_nan());

        // Bar (default component = Median = (high+low)/2)
        let mut ss2 = create_ss(10);
        ss2.update(inp);
        ss2.update(inp);
        ss2.update(inp);
        let bar = Bar::new(time, 0.0, inp, inp, 0.0, 0.0);
        let out = ss2.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!(!s.value.is_nan());

        // Quote (default component = Mid)
        let mut ss2 = create_ss(10);
        ss2.update(inp);
        ss2.update(inp);
        ss2.update(inp);
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = ss2.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!(!s.value.is_nan());

        // Trade (default component = Price)
        let mut ss2 = create_ss(10);
        ss2.update(inp);
        ss2.update(inp);
        ss2.update(inp);
        let trade = Trade::new(time, inp, 0.0);
        let out = ss2.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!(!s.value.is_nan());
    }

    #[test]
    fn test_metadata() {
        let ss = create_ss(10);
        let m = ss.metadata();

        assert_eq!(m.identifier, Identifier::SuperSmoother);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, SuperSmootherOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "ss(10, hl/2)");
        assert_eq!(m.outputs[0].description, "Super Smoother ss(10, hl/2)");
    }

    #[test]
    fn test_new_period_validation() {
        let err_msg = "invalid super smoother parameters: shortest cycle period should be greater than 1";

        let r = SuperSmoother::new(&SuperSmootherParams {
            shortest_cycle_period: 1,
            ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), err_msg);

        let r = SuperSmoother::new(&SuperSmootherParams {
            shortest_cycle_period: 0,
            ..Default::default()
        });
        assert!(r.is_err());

        let r = SuperSmoother::new(&SuperSmootherParams {
            shortest_cycle_period: -1,
            ..Default::default()
        });
        assert!(r.is_err());
    }

    #[test]
    fn test_new_all_defaults() {
        let ss = create_ss(10);
        assert_eq!(ss.line.mnemonic, "ss(10, hl/2)");
        assert_eq!(ss.line.description, "Super Smoother ss(10, hl/2)");
    }

    #[test]
    fn test_new_bar_component_open() {
        let ss = SuperSmoother::new(&SuperSmootherParams {
            shortest_cycle_period: 10,
            bar_component: Some(BarComponent::Open),
            ..Default::default()
        }).unwrap();
        assert_eq!(ss.line.mnemonic, "ss(10, o)");
    }
}
