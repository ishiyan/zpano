use std::any::Any;

use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{
    component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT,
};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{
    component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT,
};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Output indices for the Center of Gravity oscillator.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CenterOfGravityOscillatorOutput {
    /// The COG oscillator value.
    Value = 1,
    /// The trigger line (previous value of the oscillator).
    Trigger = 2,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create a Center of Gravity oscillator.
pub struct CenterOfGravityOscillatorParams {
    /// Length (number of time periods). Must be >= 1. Default 10.
    pub length: usize,
    /// Bar component. `None` means default (Median = hl/2).
    pub bar_component: Option<BarComponent>,
    /// Quote component. `None` means default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component. `None` means default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for CenterOfGravityOscillatorParams {
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
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Center of Gravity oscillator (COG).
///
/// The center of gravity in a FIR filter is the position of the average price
/// within the filter window length:
///
///   CGᵢ = Σ((i+1) * Priceᵢ) / Σ(Priceᵢ), where i = 0…ℓ-1.
///
/// Has two outputs: the oscillator value and a trigger line (previous value).
pub struct CenterOfGravityOscillator {
    mnemonic: String,
    description: String,
    mnemonic_trig: String,
    description_trig: String,
    value: f64,
    value_previous: f64,
    denominator_sum: f64,
    length: usize,
    length_min_one: usize,
    window_count: usize,
    window: Vec<f64>,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CenterOfGravityOscillator {
    /// Creates a new Center of Gravity oscillator from the given parameters.
    pub fn new(p: &CenterOfGravityOscillatorParams) -> Result<Self, String> {
        if p.length < 1 {
            return Err(
                "invalid center of gravity oscillator parameters: length should be a positive integer"
                    .to_string(),
            );
        }

        // COG default bar component is Median (hl/2), not Close.
        let bc = p.bar_component.unwrap_or(BarComponent::Median);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let component_mnemonic = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!("cog({}{component_mnemonic})", p.length);
        let mnemonic_trig = format!("cogTrig({}{component_mnemonic})", p.length);
        let description = format!("Center of Gravity oscillator {mnemonic}");
        let description_trig = format!("Center of Gravity trigger {mnemonic_trig}");

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        Ok(Self {
            mnemonic,
            description,
            mnemonic_trig,
            description_trig,
            value: f64::NAN,
            value_previous: f64::NAN,
            denominator_sum: 0.0,
            length: p.length,
            length_min_one: p.length - 1,
            window_count: 0,
            window: vec![0.0; p.length],
            primed: false,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Updates the COG oscillator with the next sample value.
    /// Returns the current COG value (NaN while not primed).
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return f64::NAN;
        }

        if self.primed {
            self.value_previous = self.value;
            self.value = self.calculate(sample);
            return self.value;
        }

        // Not primed.
        if self.length > self.window_count {
            self.denominator_sum += sample;
            self.window[self.window_count] = sample;

            if self.length_min_one == self.window_count {
                let mut sum = 0.0;
                if self.denominator_sum.abs() > f64::MIN_POSITIVE {
                    for i in 0..self.length {
                        sum += (1 + i) as f64 * self.window[i];
                    }
                    sum /= self.denominator_sum;
                }
                self.value_previous = sum;
            }
        } else {
            self.value = self.calculate(sample);
            self.primed = true;
            self.window_count += 1;
            return self.value;
        }

        self.window_count += 1;
        f64::NAN
    }

    fn calculate(&mut self, sample: f64) -> f64 {
        self.denominator_sum += sample - self.window[0];

        for i in 0..self.length_min_one {
            self.window[i] = self.window[i + 1];
        }
        self.window[self.length_min_one] = sample;

        let mut sum = 0.0;
        if self.denominator_sum.abs() > f64::MIN_POSITIVE {
            for i in 0..self.length {
                sum += (i + 1) as f64 * self.window[i];
            }
            sum /= self.denominator_sum;
        }

        sum
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let cog = self.update(sample);
        let trig = if cog.is_nan() {
            f64::NAN
        } else {
            self.value_previous
        };

        vec![
            Box::new(Scalar::new(time, cog)) as Box<dyn Any>,
            Box::new(Scalar::new(time, trig)) as Box<dyn Any>,
        ]
    }
}

impl Indicator for CenterOfGravityOscillator {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CenterOfGravityOscillator,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: self.mnemonic.clone(),
                    description: self.description.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_trig.clone(),
                    description: self.description_trig.clone(),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        self.update_entity(sample.time, v)
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;

    const TOLERANCE: f64 = 1e-8;

    fn create(length: usize) -> CenterOfGravityOscillator {
        let p = CenterOfGravityOscillatorParams {
            length,
            ..Default::default()
        };
        CenterOfGravityOscillator::new(&p).unwrap()
    }

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() < TOLERANCE
    }
    const L: usize = 10;
    const LPRIMED: usize = 10;

    #[test]
    fn test_cog_value() {
        let mut cog = create(L);
        let inp = testdata::input();
        let exp = testdata::expected_cog();

        for i in 0..LPRIMED {
            assert!(cog.update(inp[i]).is_nan(), "[{i}] expected NaN");
        }

        for i in LPRIMED..inp.len() {
            let act = cog.update(inp[i]);
            assert!(
                almost_equal(exp[i], act),
                "[{i}] cog: expected {}, actual {act}",
                exp[i]
            );
        }

        assert!(cog.update(f64::NAN).is_nan(), "NaN input should return NaN");
    }

    #[test]
    fn test_cog_trigger() {
        let mut cog = create(L);
        let inp = testdata::input();
        let exp_trig = testdata::expected_trigger();

        for i in 0..LPRIMED {
            cog.update(inp[i]);
        }

        for i in LPRIMED..inp.len() {
            cog.update(inp[i]);
            let act = cog.value_previous;
            assert!(
                almost_equal(exp_trig[i], act),
                "[{i}] trigger: expected {}, actual {act}",
                exp_trig[i]
            );
        }
    }

    #[test]
    fn test_cog_is_primed() {
        let mut cog = create(L);
        let inp = testdata::input();

        assert!(!cog.is_primed());

        for i in 0..LPRIMED {
            cog.update(inp[i]);
            assert!(!cog.is_primed(), "[{i}] should not be primed");
        }

        for i in LPRIMED..inp.len() {
            cog.update(inp[i]);
            assert!(cog.is_primed(), "[{i}] should be primed");
        }
    }

    #[test]
    fn test_cog_update_scalar() {
        let mut cog = create(L);
        let inp = testdata::input();
        let exp_cog = testdata::expected_cog();
        let exp_trig = testdata::expected_trigger();
        let time: i64 = 1617235200;

        for i in 0..inp.len() {
            let s = Scalar::new(time, inp[i]);
            let out = cog.update_scalar(&s);
            assert_eq!(out.len(), 2);

            let s0 = out[0].downcast_ref::<Scalar>().unwrap();
            let s1 = out[1].downcast_ref::<Scalar>().unwrap();
            assert_eq!(s0.time, time);
            assert_eq!(s1.time, time);

            if exp_cog[i].is_nan() {
                assert!(s0.value.is_nan());
                assert!(s1.value.is_nan());
            } else {
                assert!(
                    almost_equal(exp_cog[i], s0.value),
                    "[{i}] scalar cog: expected {}, actual {}",
                    exp_cog[i],
                    s0.value
                );
                assert!(
                    almost_equal(exp_trig[i], s1.value),
                    "[{i}] scalar trigger: expected {}, actual {}",
                    exp_trig[i],
                    s1.value
                );
            }
        }
    }

    #[test]
    fn test_cog_update_bar() {
        let mut cog = create(L);
        let hi = testdata::input_high();
        let lo = testdata::input_low();
        let exp_cog = testdata::expected_cog();
        let exp_trig = testdata::expected_trigger();
        let time: i64 = 1617235200;

        for i in 0..hi.len() {
            let b = Bar::new(time, 0.0, hi[i], lo[i], 0.0, 0.0);
            let out = cog.update_bar(&b);
            assert_eq!(out.len(), 2);

            let s0 = out[0].downcast_ref::<Scalar>().unwrap();
            let s1 = out[1].downcast_ref::<Scalar>().unwrap();

            if exp_cog[i].is_nan() {
                assert!(s0.value.is_nan());
                assert!(s1.value.is_nan());
            } else {
                assert!(
                    almost_equal(exp_cog[i], s0.value),
                    "[{i}] bar cog: expected {}, actual {}",
                    exp_cog[i],
                    s0.value
                );
                assert!(
                    almost_equal(exp_trig[i], s1.value),
                    "[{i}] bar trigger: expected {}, actual {}",
                    exp_trig[i],
                    s1.value
                );
            }
        }
    }

    #[test]
    fn test_cog_update_quote() {
        let mut cog = create(L);
        let hi = testdata::input_high();
        let lo = testdata::input_low();
        let exp_cog = testdata::expected_cog();
        let exp_trig = testdata::expected_trigger();
        let time: i64 = 1617235200;

        for i in 0..hi.len() {
            // QuoteMidPrice = (ask + bid) / 2; feed high as ask, low as bid.
            let q = Quote::new(time, lo[i], hi[i], 0.0, 0.0);
            let out = cog.update_quote(&q);
            assert_eq!(out.len(), 2);

            let s0 = out[0].downcast_ref::<Scalar>().unwrap();
            let s1 = out[1].downcast_ref::<Scalar>().unwrap();

            if exp_cog[i].is_nan() {
                assert!(s0.value.is_nan());
                assert!(s1.value.is_nan());
            } else {
                assert!(
                    almost_equal(exp_cog[i], s0.value),
                    "[{i}] quote cog: expected {}, actual {}",
                    exp_cog[i],
                    s0.value
                );
                assert!(
                    almost_equal(exp_trig[i], s1.value),
                    "[{i}] quote trigger: expected {}, actual {}",
                    exp_trig[i],
                    s1.value
                );
            }
        }
    }

    #[test]
    fn test_cog_update_trade() {
        let mut cog = create(L);
        let inp = testdata::input();
        let exp_cog = testdata::expected_cog();
        let exp_trig = testdata::expected_trigger();
        let time: i64 = 1617235200;

        for i in 0..inp.len() {
            let t = Trade::new(time, inp[i], 0.0);
            let out = cog.update_trade(&t);
            assert_eq!(out.len(), 2);

            let s0 = out[0].downcast_ref::<Scalar>().unwrap();
            let s1 = out[1].downcast_ref::<Scalar>().unwrap();

            if exp_cog[i].is_nan() {
                assert!(s0.value.is_nan());
                assert!(s1.value.is_nan());
            } else {
                assert!(
                    almost_equal(exp_cog[i], s0.value),
                    "[{i}] trade cog: expected {}, actual {}",
                    exp_cog[i],
                    s0.value
                );
                assert!(
                    almost_equal(exp_trig[i], s1.value),
                    "[{i}] trade trigger: expected {}, actual {}",
                    exp_trig[i],
                    s1.value
                );
            }
        }
    }

    #[test]
    fn test_cog_metadata() {
        let cog = create(L);
        let m = cog.metadata();

        assert_eq!(m.identifier, Identifier::CenterOfGravityOscillator);
        assert_eq!(m.mnemonic, "cog(10, hl/2)");
        assert_eq!(m.description, "Center of Gravity oscillator cog(10, hl/2)");
        assert_eq!(m.outputs.len(), 2);

        assert_eq!(m.outputs[0].kind, CenterOfGravityOscillatorOutput::Value as i32);
        assert_eq!(m.outputs[0].mnemonic, "cog(10, hl/2)");

        assert_eq!(m.outputs[1].kind, CenterOfGravityOscillatorOutput::Trigger as i32);
        assert_eq!(m.outputs[1].mnemonic, "cogTrig(10, hl/2)");
    }

    #[test]
    fn test_cog_new_errors() {
        let p = CenterOfGravityOscillatorParams {
            length: 0,
            ..Default::default()
        };
        assert!(CenterOfGravityOscillator::new(&p).is_err());

        let p2 = CenterOfGravityOscillatorParams {
            length: 10,
            ..Default::default()
        };
        assert!(CenterOfGravityOscillator::new(&p2).is_ok());
    }
}
