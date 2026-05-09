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
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Cyber Cycle indicator based on length.
pub struct CyberCycleLengthParams {
    /// The length of the cyber cycle. Must be >= 1. Default is 28.
    pub length: i64,
    /// The lag of the signal line (EMA). Must be >= 1. Default is 9.
    pub signal_lag: i64,
    /// Bar component. `None` means default (Median).
    pub bar_component: Option<BarComponent>,
    /// Quote component. `None` means default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component. `None` means default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for CyberCycleLengthParams {
    fn default() -> Self {
        Self {
            length: 28,
            signal_lag: 9,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Parameters for the Cyber Cycle indicator based on smoothing factor.
pub struct CyberCycleSmoothingFactorParams {
    /// The smoothing factor alpha. Must be in [0, 1]. Default is 0.07.
    pub smoothing_factor: f64,
    /// The lag of the signal line (EMA). Must be >= 1. Default is 9.
    pub signal_lag: i64,
    /// Bar component. `None` means default (Median).
    pub bar_component: Option<BarComponent>,
    /// Quote component. `None` means default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component. `None` means default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for CyberCycleSmoothingFactorParams {
    fn default() -> Self {
        Self {
            smoothing_factor: 0.07,
            signal_lag: 9,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Cyber Cycle indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CyberCycleOutput {
    /// The scalar value of the cyber cycle.
    Value = 1,
    /// The scalar value of the signal line (EMA of the cyber cycle value).
    Signal = 2,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Cyber Cycle indicator.
///
/// H(z) = ((1-alpha/2)^2 (1 - 2z^-1 + z^-2)) / (1 - 2(1-alpha)z^-1 + (1-alpha)^2 z^-2)
///
/// Two outputs: cycle value and signal line (EMA of cycle value).
pub struct CyberCycle {
    length: i64,
    smoothing_factor: f64,
    signal_lag: i64,
    mnemonic: String,
    description: String,
    mnemonic_signal: String,
    description_signal: String,
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    coeff4: f64,
    coeff5: f64,
    count: i64,
    previous_sample1: f64,
    previous_sample2: f64,
    previous_sample3: f64,
    smoothed: f64,
    previous_smoothed1: f64,
    previous_smoothed2: f64,
    value: f64,
    previous_value1: f64,
    previous_value2: f64,
    signal: f64,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CyberCycle {
    /// Creates a new Cyber Cycle from length-based parameters.
    pub fn new_length(params: &CyberCycleLengthParams) -> Result<Self, String> {
        Self::new_inner(
            params.length,
            f64::NAN,
            params.signal_lag,
            params.bar_component,
            params.quote_component,
            params.trade_component,
        )
    }

    /// Creates a new Cyber Cycle from smoothing-factor-based parameters.
    pub fn new_smoothing_factor(params: &CyberCycleSmoothingFactorParams) -> Result<Self, String> {
        Self::new_inner(
            0,
            params.smoothing_factor,
            params.signal_lag,
            params.bar_component,
            params.quote_component,
            params.trade_component,
        )
    }

    fn new_inner(
        mut length: i64,
        mut alpha: f64,
        signal_lag: i64,
        bc: Option<BarComponent>,
        qc: Option<QuoteComponent>,
        tc: Option<TradeComponent>,
    ) -> Result<Self, String> {
        const INVALID: &str = "invalid cyber cycle parameters";
        const EPSILON: f64 = 0.00000001;

        if alpha.is_nan() {
            // Length-based construction.
            if length < 1 {
                return Err(format!("{}: length should be a positive integer", INVALID));
            }
            alpha = 2.0 / (1 + length) as f64;
        } else {
            // Smoothing-factor-based construction.
            if alpha < 0.0 || alpha > 1.0 {
                return Err(format!(
                    "{}: smoothing factor should be in range [0, 1]",
                    INVALID
                ));
            }
            if alpha < EPSILON {
                length = i64::MAX;
            } else {
                length = (2.0 / alpha).round() as i64 - 1;
            }
        }

        if signal_lag < 1 {
            return Err(format!(
                "{}: signal lag should be a positive integer",
                INVALID
            ));
        }

        let bc = bc.unwrap_or(BarComponent::Median);
        let qc = qc.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = tc.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let comp_mn = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!("cc({}{})", length, comp_mn);
        let mnemonic_signal = format!("ccSignal({}{})", length, comp_mn);
        let description = format!("Cyber Cycle {}", mnemonic);
        let description_signal = format!("Cyber Cycle signal {}", mnemonic_signal);

        // Coefficients.
        let x = 1.0 - alpha / 2.0;
        let c1 = x * x;
        let x = 1.0 - alpha;
        let c2 = 2.0 * x;
        let c3 = -(x * x);
        let x = 1.0 / (1 + signal_lag) as f64;
        let c4 = x;
        let c5 = 1.0 - x;

        Ok(Self {
            length,
            smoothing_factor: alpha,
            signal_lag,
            mnemonic,
            description,
            mnemonic_signal,
            description_signal,
            coeff1: c1,
            coeff2: c2,
            coeff3: c3,
            coeff4: c4,
            coeff5: c5,
            count: 0,
            previous_sample1: 0.0,
            previous_sample2: 0.0,
            previous_sample3: 0.0,
            smoothed: 0.0,
            previous_smoothed1: 0.0,
            previous_smoothed2: 0.0,
            value: f64::NAN,
            previous_value1: 0.0,
            previous_value2: 0.0,
            signal: f64::NAN,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Core update logic. Returns the cycle value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return f64::NAN;
        }

        if self.primed {
            self.previous_smoothed2 = self.previous_smoothed1;
            self.previous_smoothed1 = self.smoothed;
            self.smoothed = (sample
                + 2.0 * self.previous_sample1
                + 2.0 * self.previous_sample2
                + self.previous_sample3)
                / 6.0;

            self.previous_value2 = self.previous_value1;
            self.previous_value1 = self.value;
            self.value = self.coeff1
                * (self.smoothed - 2.0 * self.previous_smoothed1 + self.previous_smoothed2)
                + self.coeff2 * self.previous_value1
                + self.coeff3 * self.previous_value2;

            self.signal = self.coeff4 * self.value + self.coeff5 * self.signal;

            self.previous_sample3 = self.previous_sample2;
            self.previous_sample2 = self.previous_sample1;
            self.previous_sample1 = sample;

            return self.value;
        }

        self.count += 1;

        match self.count {
            1 => {
                self.previous_sample3 = sample;
                f64::NAN
            }
            2 => {
                self.previous_sample2 = sample;
                f64::NAN
            }
            3 => {
                self.signal =
                    self.coeff4 * (sample - 2.0 * self.previous_sample2 + self.previous_sample3)
                        / 4.0;
                self.previous_sample1 = sample;
                f64::NAN
            }
            4 => {
                self.previous_smoothed2 = (sample
                    + 2.0 * self.previous_sample1
                    + 2.0 * self.previous_sample2
                    + self.previous_sample3)
                    / 6.0;
                self.signal = self.coeff4
                    * (sample - 2.0 * self.previous_sample1 + self.previous_sample2)
                    / 4.0
                    + self.coeff5 * self.signal;

                self.previous_sample3 = self.previous_sample2;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                f64::NAN
            }
            5 => {
                self.previous_smoothed1 = (sample
                    + 2.0 * self.previous_sample1
                    + 2.0 * self.previous_sample2
                    + self.previous_sample3)
                    / 6.0;
                self.signal = self.coeff4
                    * (sample - 2.0 * self.previous_sample1 + self.previous_sample2)
                    / 4.0
                    + self.coeff5 * self.signal;

                self.previous_sample3 = self.previous_sample2;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                f64::NAN
            }
            6 => {
                self.smoothed = (sample
                    + 2.0 * self.previous_sample1
                    + 2.0 * self.previous_sample2
                    + self.previous_sample3)
                    / 6.0;
                self.previous_value2 =
                    (sample - 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0;
                self.signal = self.coeff4 * self.previous_value2 + self.coeff5 * self.signal;

                self.previous_sample3 = self.previous_sample2;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                f64::NAN
            }
            7 => {
                self.previous_smoothed2 = self.previous_smoothed1;
                self.previous_smoothed1 = self.smoothed;
                self.smoothed = (sample
                    + 2.0 * self.previous_sample1
                    + 2.0 * self.previous_sample2
                    + self.previous_sample3)
                    / 6.0;
                self.previous_value1 =
                    (sample - 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0;
                self.signal = self.coeff4 * self.previous_value1 + self.coeff5 * self.signal;

                self.previous_sample3 = self.previous_sample2;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                f64::NAN
            }
            8 => {
                self.previous_smoothed2 = self.previous_smoothed1;
                self.previous_smoothed1 = self.smoothed;
                self.smoothed = (sample
                    + 2.0 * self.previous_sample1
                    + 2.0 * self.previous_sample2
                    + self.previous_sample3)
                    / 6.0;

                self.value = self.coeff1
                    * (self.smoothed - 2.0 * self.previous_smoothed1 + self.previous_smoothed2)
                    + self.coeff2 * self.previous_value1
                    + self.coeff3 * self.previous_value2;

                self.signal = self.coeff4 * self.value + self.coeff5 * self.signal;

                self.previous_sample3 = self.previous_sample2;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                self.primed = true;

                self.value
            }
            _ => f64::NAN,
        }
    }

    /// Returns the current signal value.
    pub fn signal(&self) -> f64 {
        self.signal
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let v = self.update(sample);
        let sig = if v.is_nan() { f64::NAN } else { self.signal };
        vec![
            Box::new(Scalar::new(time, v)),
            Box::new(Scalar::new(time, sig)),
        ]
    }
}

impl Indicator for CyberCycle {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CyberCycle,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: self.mnemonic.clone(),
                    description: self.description.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_signal.clone(),
                    description: self.description_signal.clone(),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        self.update_entity(sample.time, (self.bar_func)(sample))
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        self.update_entity(sample.time, (self.quote_func)(sample))
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        self.update_entity(sample.time, (self.trade_func)(sample))
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

    const TOLERANCE: f64 = 1e-8;
    const L_PRIMED: usize = 7;

    fn create_default() -> CyberCycle {
        CyberCycle::new_smoothing_factor(&CyberCycleSmoothingFactorParams {
            smoothing_factor: 0.07,
            signal_lag: 9,
            ..Default::default()
        })
        .unwrap()
    }
    #[test]
    fn test_update_cycle_value() {
        let input = testdata::test_input();
        let expected = testdata::test_expected_cycle();
        let mut cc = create_default();

        for i in 0..L_PRIMED {
            let act = cc.update(input[i]);
            assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
        }

        for i in L_PRIMED..input.len() {
            let act = cc.update(input[i]);
            assert!(
                (act - expected[i]).abs() <= TOLERANCE,
                "[{}] expected {}, got {}",
                i,
                expected[i],
                act
            );
        }

        // NaN passthrough
        assert!(cc.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_signal() {
        let input = testdata::test_input();
        let expected_signal = testdata::test_expected_signal();
        let mut cc = create_default();

        for i in 0..L_PRIMED {
            cc.update(input[i]);
        }

        for i in L_PRIMED..input.len() {
            cc.update(input[i]);
            let act = cc.signal();
            assert!(
                (act - expected_signal[i]).abs() <= TOLERANCE,
                "[{}] signal expected {}, got {}",
                i,
                expected_signal[i],
                act
            );
        }
    }

    #[test]
    fn test_update_entity() {
        let input = testdata::test_input();
        let input_high = testdata::test_input_high();
        let input_low = testdata::test_input_low();
        let exp_cycle = testdata::test_expected_cycle();
        let exp_signal = testdata::test_expected_signal();
        let time = 1617235200_i64;

        // Scalar
        {
            let mut cc = create_default();
            for i in 0..input.len() {
                let s = Scalar::new(time, input[i]);
                let out = cc.update_scalar(&s);
                assert_eq!(out.len(), 2);
                let s0 = out[0].downcast_ref::<Scalar>().unwrap();
                let s1 = out[1].downcast_ref::<Scalar>().unwrap();
                assert_eq!(s0.time, time);
                assert_eq!(s1.time, time);

                if exp_cycle[i].is_nan() {
                    assert!(s0.value.is_nan(), "[{}] expected NaN value", i);
                    assert!(s1.value.is_nan(), "[{}] expected NaN signal", i);
                } else {
                    assert!(
                        (s0.value - exp_cycle[i]).abs() <= TOLERANCE,
                        "[{}] value expected {}, got {}",
                        i,
                        exp_cycle[i],
                        s0.value
                    );
                    assert!(
                        (s1.value - exp_signal[i]).abs() <= TOLERANCE,
                        "[{}] signal expected {}, got {}",
                        i,
                        exp_signal[i],
                        s1.value
                    );
                }
            }
        }

        // Bar (default = Median = (high+low)/2)
        {
            let mut cc = create_default();
            for i in 0..input.len() {
                let bar = Bar::new(time, 0.0, input_high[i], input_low[i], 0.0, 0.0);
                let out = cc.update_bar(&bar);
                let s0 = out[0].downcast_ref::<Scalar>().unwrap();
                let s1 = out[1].downcast_ref::<Scalar>().unwrap();

                if exp_cycle[i].is_nan() {
                    assert!(s0.value.is_nan());
                    assert!(s1.value.is_nan());
                } else {
                    assert!(
                        (s0.value - exp_cycle[i]).abs() <= TOLERANCE,
                        "[{}] bar value expected {}, got {}",
                        i,
                        exp_cycle[i],
                        s0.value
                    );
                    assert!(
                        (s1.value - exp_signal[i]).abs() <= TOLERANCE,
                        "[{}] bar signal expected {}, got {}",
                        i,
                        exp_signal[i],
                        s1.value
                    );
                }
            }
        }

        // Trade
        {
            let mut cc = create_default();
            for i in 0..input.len() {
                let trade = Trade::new(time, input[i], 0.0);
                let out = cc.update_trade(&trade);
                let s0 = out[0].downcast_ref::<Scalar>().unwrap();

                if exp_cycle[i].is_nan() {
                    assert!(s0.value.is_nan());
                } else {
                    assert!(
                        (s0.value - exp_cycle[i]).abs() <= TOLERANCE,
                        "[{}] trade value expected {}, got {}",
                        i,
                        exp_cycle[i],
                        s0.value
                    );
                }
            }
        }
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut cc = create_default();

        assert!(!cc.is_primed());

        for i in 0..L_PRIMED {
            cc.update(input[i]);
            assert!(!cc.is_primed(), "[{}] should not be primed", i);
        }

        cc.update(input[L_PRIMED]);
        assert!(cc.is_primed(), "should be primed after sample 8");
    }

    #[test]
    fn test_metadata() {
        let cc = create_default();
        let m = cc.metadata();

        assert_eq!(m.identifier, Identifier::CyberCycle);
        assert_eq!(m.mnemonic, "cc(28, hl/2)");
        assert_eq!(m.description, "Cyber Cycle cc(28, hl/2)");
        assert_eq!(m.outputs.len(), 2);

        assert_eq!(m.outputs[0].kind, CyberCycleOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "cc(28, hl/2)");

        assert_eq!(m.outputs[1].kind, CyberCycleOutput::Signal as i32);
        assert_eq!(m.outputs[1].shape, Shape::Scalar);
        assert_eq!(m.outputs[1].mnemonic, "ccSignal(28, hl/2)");
        assert_eq!(
            m.outputs[1].description,
            "Cyber Cycle signal ccSignal(28, hl/2)"
        );
    }

    #[test]
    fn test_new_length_validation() {
        let invalid_length = "invalid cyber cycle parameters: length should be a positive integer";
        let invalid_signal =
            "invalid cyber cycle parameters: signal lag should be a positive integer";

        // Valid
        let r = CyberCycle::new_length(&CyberCycleLengthParams {
            length: 28,
            signal_lag: 14,
            ..Default::default()
        });
        assert!(r.is_ok());

        // length=0
        let r = CyberCycle::new_length(&CyberCycleLengthParams {
            length: 0,
            signal_lag: 1,
            ..Default::default()
        });
        assert_eq!(r.err().unwrap(), invalid_length);

        // length=-8
        let r = CyberCycle::new_length(&CyberCycleLengthParams {
            length: -8,
            signal_lag: 1,
            ..Default::default()
        });
        assert_eq!(r.err().unwrap(), invalid_length);

        // signal_lag=0
        let r = CyberCycle::new_length(&CyberCycleLengthParams {
            length: 1,
            signal_lag: 0,
            ..Default::default()
        });
        assert_eq!(r.err().unwrap(), invalid_signal);

        // signal_lag=-8
        let r = CyberCycle::new_length(&CyberCycleLengthParams {
            length: 1,
            signal_lag: -8,
            ..Default::default()
        });
        assert_eq!(r.err().unwrap(), invalid_signal);
    }

    #[test]
    fn test_new_smoothing_factor_validation() {
        let invalid_alpha =
            "invalid cyber cycle parameters: smoothing factor should be in range [0, 1]";
        let invalid_signal =
            "invalid cyber cycle parameters: signal lag should be a positive integer";

        // Valid default
        let r = CyberCycle::new_smoothing_factor(&CyberCycleSmoothingFactorParams {
            smoothing_factor: 0.07,
            signal_lag: 9,
            ..Default::default()
        });
        assert!(r.is_ok());
        let cc = r.unwrap();
        assert_eq!(cc.length, 28);
        assert_eq!(cc.smoothing_factor, 0.07);

        // alpha=0.06 => length=32
        let r = CyberCycle::new_smoothing_factor(&CyberCycleSmoothingFactorParams {
            smoothing_factor: 0.06,
            signal_lag: 11,
            ..Default::default()
        });
        let cc = r.unwrap();
        assert_eq!(cc.length, 32);

        // near-zero alpha => length=MAX
        let r = CyberCycle::new_smoothing_factor(&CyberCycleSmoothingFactorParams {
            smoothing_factor: 0.000000001,
            signal_lag: 9,
            ..Default::default()
        });
        let cc = r.unwrap();
        assert_eq!(cc.length, i64::MAX);

        // alpha=0 => length=MAX
        let r = CyberCycle::new_smoothing_factor(&CyberCycleSmoothingFactorParams {
            smoothing_factor: 0.0,
            signal_lag: 9,
            ..Default::default()
        });
        let cc = r.unwrap();
        assert_eq!(cc.length, i64::MAX);

        // alpha=1 => length=1
        let r = CyberCycle::new_smoothing_factor(&CyberCycleSmoothingFactorParams {
            smoothing_factor: 1.0,
            signal_lag: 9,
            ..Default::default()
        });
        let cc = r.unwrap();
        assert_eq!(cc.length, 1);

        // alpha=-0.0001
        let r = CyberCycle::new_smoothing_factor(&CyberCycleSmoothingFactorParams {
            smoothing_factor: -0.0001,
            signal_lag: 8,
            ..Default::default()
        });
        assert_eq!(r.err().unwrap(), invalid_alpha);

        // alpha=1.0001
        let r = CyberCycle::new_smoothing_factor(&CyberCycleSmoothingFactorParams {
            smoothing_factor: 1.0001,
            signal_lag: 8,
            ..Default::default()
        });
        assert_eq!(r.err().unwrap(), invalid_alpha);

        // signal_lag=0
        let r = CyberCycle::new_smoothing_factor(&CyberCycleSmoothingFactorParams {
            smoothing_factor: 0.07,
            signal_lag: 0,
            ..Default::default()
        });
        assert_eq!(r.err().unwrap(), invalid_signal);
    }

    #[test]
    fn test_metadata_with_non_default_trade_component() {
        let cc = CyberCycle::new_length(&CyberCycleLengthParams {
            length: 3,
            signal_lag: 2,
            trade_component: Some(TradeComponent::Volume),
            ..Default::default()
        })
        .unwrap();
        let m = cc.metadata();
        assert_eq!(m.mnemonic, "cc(3, hl/2, v)");
        assert_eq!(m.description, "Cyber Cycle cc(3, hl/2, v)");
    }
}
