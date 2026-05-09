use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

const DEFAULT_ACCELERATION_INIT: f64 = 0.02;
const DEFAULT_ACCELERATION_STEP: f64 = 0.02;
const DEFAULT_ACCELERATION_MAX: f64 = 0.20;

/// Parameters for the Parabolic Stop And Reverse (SAR) indicator.
pub struct ParabolicStopAndReverseParams {
    /// Start value: 0 = auto-detect, >0 = force long, <0 = force short.
    pub start_value: f64,
    /// Percent offset added/removed on reversal.
    pub offset_on_reverse: f64,
    /// Initial acceleration factor for long direction.
    pub acceleration_init_long: f64,
    /// Acceleration factor increment for long direction.
    pub acceleration_long: f64,
    /// Maximum acceleration factor for long direction.
    pub acceleration_max_long: f64,
    /// Initial acceleration factor for short direction.
    pub acceleration_init_short: f64,
    /// Acceleration factor increment for short direction.
    pub acceleration_short: f64,
    /// Maximum acceleration factor for short direction.
    pub acceleration_max_short: f64,
}

impl Default for ParabolicStopAndReverseParams {
    fn default() -> Self {
        Self {
            start_value: 0.0,
            offset_on_reverse: 0.0,
            acceleration_init_long: DEFAULT_ACCELERATION_INIT,
            acceleration_long: DEFAULT_ACCELERATION_STEP,
            acceleration_max_long: DEFAULT_ACCELERATION_MAX,
            acceleration_init_short: DEFAULT_ACCELERATION_INIT,
            acceleration_short: DEFAULT_ACCELERATION_STEP,
            acceleration_max_short: DEFAULT_ACCELERATION_MAX,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Parabolic Stop And Reverse indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum ParabolicStopAndReverseOutput {
    /// The scalar value of the Parabolic SAR.
    /// Positive values indicate a long position; negative values indicate a short position.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const SAR_MNEMONIC: &str = "sar()";
const SAR_DESCRIPTION: &str = "Parabolic Stop And Reverse sar()";

/// Welles Wilder's Parabolic Stop And Reverse (SAR) indicator.
pub struct ParabolicStopAndReverse {
    // Resolved parameters.
    start_value: f64,
    offset_on_reverse: f64,
    af_init_long: f64,
    af_step_long: f64,
    af_max_long: f64,
    af_init_short: f64,
    af_step_short: f64,
    af_max_short: f64,

    // State.
    count: usize,
    is_long: bool,
    sar: f64,
    ep: f64,
    af_long: f64,
    af_short: f64,
    previous_high: f64,
    previous_low: f64,
    new_high: f64,
    new_low: f64,
    primed: bool,
}

impl ParabolicStopAndReverse {
    /// Creates a new Parabolic Stop And Reverse indicator.
    pub fn new(p: &ParabolicStopAndReverseParams) -> Result<Self, String> {
        let mut af_init_long = if p.acceleration_init_long == 0.0 { DEFAULT_ACCELERATION_INIT } else { p.acceleration_init_long };
        let mut af_step_long = if p.acceleration_long == 0.0 { DEFAULT_ACCELERATION_STEP } else { p.acceleration_long };
        let mut af_max_long = if p.acceleration_max_long == 0.0 { DEFAULT_ACCELERATION_MAX } else { p.acceleration_max_long };
        let mut af_init_short = if p.acceleration_init_short == 0.0 { DEFAULT_ACCELERATION_INIT } else { p.acceleration_init_short };
        let mut af_step_short = if p.acceleration_short == 0.0 { DEFAULT_ACCELERATION_STEP } else { p.acceleration_short };
        let mut af_max_short = if p.acceleration_max_short == 0.0 { DEFAULT_ACCELERATION_MAX } else { p.acceleration_max_short };

        if af_init_long < 0.0 || af_step_long < 0.0 || af_max_long < 0.0 {
            return Err("invalid parabolic stop and reverse parameters: long acceleration factors must be non-negative".to_string());
        }
        if af_init_short < 0.0 || af_step_short < 0.0 || af_max_short < 0.0 {
            return Err("invalid parabolic stop and reverse parameters: short acceleration factors must be non-negative".to_string());
        }
        if p.offset_on_reverse < 0.0 {
            return Err("invalid parabolic stop and reverse parameters: offset on reverse must be non-negative".to_string());
        }

        // Clamp: init and step cannot exceed max.
        if af_init_long > af_max_long { af_init_long = af_max_long; }
        if af_step_long > af_max_long { af_step_long = af_max_long; }
        if af_init_short > af_max_short { af_init_short = af_max_short; }
        if af_step_short > af_max_short { af_step_short = af_max_short; }

        Ok(Self {
            start_value: p.start_value,
            offset_on_reverse: p.offset_on_reverse,
            af_init_long,
            af_step_long,
            af_max_long,
            af_init_short,
            af_step_short,
            af_max_short,
            af_long: af_init_long,
            af_short: af_init_short,
            count: 0,
            is_long: false,
            sar: 0.0,
            ep: 0.0,
            previous_high: 0.0,
            previous_low: 0.0,
            new_high: 0.0,
            new_low: 0.0,
            primed: false,
        })
    }

    /// Updates the indicator with high and low values.
    pub fn update_hl(&mut self, high: f64, low: f64) -> f64 {
        if high.is_nan() || low.is_nan() {
            return f64::NAN;
        }

        self.count += 1;

        // First bar: store high/low, no output yet.
        if self.count == 1 {
            self.new_high = high;
            self.new_low = low;
            return f64::NAN;
        }

        // Second bar: initialize SAR, EP, and direction.
        if self.count == 2 {
            let previous_high = self.new_high;
            let previous_low = self.new_low;

            if self.start_value == 0.0 {
                // Auto-detect direction.
                let mut minus_dm = previous_low - low;
                let mut plus_dm = high - previous_high;
                if minus_dm < 0.0 { minus_dm = 0.0; }
                if plus_dm < 0.0 { plus_dm = 0.0; }

                self.is_long = minus_dm <= plus_dm;

                if self.is_long {
                    self.ep = high;
                    self.sar = previous_low;
                } else {
                    self.ep = low;
                    self.sar = previous_high;
                }
            } else if self.start_value > 0.0 {
                self.is_long = true;
                self.ep = high;
                self.sar = self.start_value;
            } else {
                self.is_long = false;
                self.ep = low;
                self.sar = self.start_value.abs();
            }

            self.new_high = high;
            self.new_low = low;
            self.primed = true;
        }

        // Main SAR calculation (bars 2+).
        if self.count >= 2 {
            self.previous_low = self.new_low;
            self.previous_high = self.new_high;
            self.new_low = low;
            self.new_high = high;

            if self.count == 2 {
                self.previous_low = self.new_low;
                self.previous_high = self.new_high;
            }

            if self.is_long {
                return self.update_long();
            }
            return self.update_short();
        }

        f64::NAN
    }

    fn update_long(&mut self) -> f64 {
        // Switch to short if low penetrates SAR.
        if self.new_low <= self.sar {
            self.is_long = false;
            self.sar = self.ep;

            if self.sar < self.previous_high { self.sar = self.previous_high; }
            if self.sar < self.new_high { self.sar = self.new_high; }

            if self.offset_on_reverse != 0.0 {
                self.sar += self.sar * self.offset_on_reverse;
            }

            let result = -self.sar;

            // Reset short AF and set EP.
            self.af_short = self.af_init_short;
            self.ep = self.new_low;

            // Calculate new SAR.
            self.sar = self.sar + self.af_short * (self.ep - self.sar);

            if self.sar < self.previous_high { self.sar = self.previous_high; }
            if self.sar < self.new_high { self.sar = self.new_high; }

            return result;
        }

        // No switch — output current SAR.
        let result = self.sar;

        // Adjust AF and EP.
        if self.new_high > self.ep {
            self.ep = self.new_high;
            self.af_long += self.af_step_long;
            if self.af_long > self.af_max_long { self.af_long = self.af_max_long; }
        }

        // Calculate new SAR.
        self.sar = self.sar + self.af_long * (self.ep - self.sar);

        if self.sar > self.previous_low { self.sar = self.previous_low; }
        if self.sar > self.new_low { self.sar = self.new_low; }

        result
    }

    fn update_short(&mut self) -> f64 {
        // Switch to long if high penetrates SAR.
        if self.new_high >= self.sar {
            self.is_long = true;
            self.sar = self.ep;

            if self.sar > self.previous_low { self.sar = self.previous_low; }
            if self.sar > self.new_low { self.sar = self.new_low; }

            if self.offset_on_reverse != 0.0 {
                self.sar -= self.sar * self.offset_on_reverse;
            }

            let result = self.sar;

            // Reset long AF and set EP.
            self.af_long = self.af_init_long;
            self.ep = self.new_high;

            // Calculate new SAR.
            self.sar = self.sar + self.af_long * (self.ep - self.sar);

            if self.sar > self.previous_low { self.sar = self.previous_low; }
            if self.sar > self.new_low { self.sar = self.new_low; }

            return result;
        }

        // No switch — output negated SAR.
        let result = -self.sar;

        // Adjust AF and EP.
        if self.new_low < self.ep {
            self.ep = self.new_low;
            self.af_short += self.af_step_short;
            if self.af_short > self.af_max_short { self.af_short = self.af_max_short; }
        }

        // Calculate new SAR.
        self.sar = self.sar + self.af_short * (self.ep - self.sar);

        if self.sar < self.previous_high { self.sar = self.previous_high; }
        if self.sar < self.new_high { self.sar = self.new_high; }

        result
    }

    /// Updates using a single sample value (high = low = sample).
    pub fn update_sample(&mut self, sample: f64) -> f64 {
        self.update_hl(sample, sample)
    }
}

impl Indicator for ParabolicStopAndReverse {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::ParabolicStopAndReverse,
            SAR_MNEMONIC,
            SAR_DESCRIPTION,
            &[OutputText {
                mnemonic: SAR_MNEMONIC.to_string(),
                description: SAR_DESCRIPTION.to_string(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = self.update_sample(sample.value);
        vec![Box::new(Scalar::new(sample.time, v))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = self.update_hl(sample.high, sample.low);
        vec![Box::new(Scalar::new(sample.time, v))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (sample.bid_price + sample.ask_price) / 2.0;
        let result = self.update_sample(v);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let result = self.update_sample(sample.price);
        vec![Box::new(Scalar::new(sample.time, result))]
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    #[test]
    fn test_252_bar() {
        let tol = 1e-6;
        let mut sar = ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams::default()).unwrap();

        let highs = testdata::test_highs();
        let lows = testdata::test_lows();
        let expected = testdata::test_expected();

        for i in 0..highs.len() {
            let result = sar.update_hl(highs[i], lows[i]);

            if expected[i].is_nan() {
                assert!(result.is_nan(), "[{}] expected NaN, got {}", i, result);
                continue;
            }

            let diff = (result - expected[i]).abs();
            assert!(
                diff <= tol,
                "[{}] expected {:.10}, got {:.10}, diff {:.10}",
                i, expected[i], result, diff
            );
        }
    }

    #[test]
    fn test_is_primed() {
        let mut sar = ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams::default()).unwrap();

        assert!(!sar.is_primed());
        sar.update_hl(93.25, 90.75);
        assert!(!sar.is_primed());
        sar.update_hl(94.94, 91.405);
        assert!(sar.is_primed());
    }

    #[test]
    fn test_nan_passthrough() {
        let mut sar = ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams::default()).unwrap();

        // Prime.
        sar.update_hl(93.25, 90.75);
        sar.update_hl(94.94, 91.405);

        // NaN should not corrupt state.
        let result = sar.update_hl(f64::NAN, 92.0);
        assert!(result.is_nan());

        // Valid data should still work.
        let result = sar.update_hl(96.375, 94.25);
        assert!(!result.is_nan());
    }

    #[test]
    fn test_constructor_validation() {
        // Defaults: ok
        assert!(ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams::default()).is_ok());

        // Negative long init: err
        assert!(ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams {
            acceleration_init_long: -0.01,
            ..Default::default()
        }).is_err());

        // Negative short step: err
        assert!(ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams {
            acceleration_short: -0.01,
            ..Default::default()
        }).is_err());

        // Negative offset: err
        assert!(ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams {
            offset_on_reverse: -0.01,
            ..Default::default()
        }).is_err());

        // Custom valid
        assert!(ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams {
            acceleration_init_long: 0.01,
            acceleration_long: 0.01,
            acceleration_max_long: 0.10,
            acceleration_init_short: 0.03,
            acceleration_short: 0.03,
            acceleration_max_short: 0.30,
            ..Default::default()
        }).is_ok());

        // Start value positive/negative: ok
        assert!(ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams {
            start_value: 100.0,
            ..Default::default()
        }).is_ok());
        assert!(ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams {
            start_value: -100.0,
            ..Default::default()
        }).is_ok());
    }

    #[test]
    fn test_metadata() {
        let sar = ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams::default()).unwrap();
        let meta = sar.metadata();
        assert_eq!(meta.identifier, Identifier::ParabolicStopAndReverse);
        assert_eq!(meta.mnemonic, "sar()");
        assert_eq!(meta.description, "Parabolic Stop And Reverse sar()");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, ParabolicStopAndReverseOutput::Value as i32);
        assert_eq!(meta.outputs[0].mnemonic, "sar()");
    }

    #[test]
    fn test_update_bar() {
        let mut sar = ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams::default()).unwrap();

        let bar1 = Bar { time: 1000, open: 91.0, high: 93.25, low: 90.75, close: 91.5, volume: 1000.0 };
        let out1 = sar.update_bar(&bar1);
        let s1 = out1[0].downcast_ref::<Scalar>().unwrap();
        assert!(s1.value.is_nan());

        let bar2 = Bar { time: 2000, open: 92.0, high: 94.94, low: 91.405, close: 94.815, volume: 1000.0 };
        let out2 = sar.update_bar(&bar2);
        let s2 = out2[0].downcast_ref::<Scalar>().unwrap();
        assert!(!s2.value.is_nan());
    }

    #[test]
    fn test_forced_start_long() {
        let mut sar = ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams {
            start_value: 85.0,
            ..Default::default()
        }).unwrap();

        let highs = testdata::test_highs();
        let lows = testdata::test_lows();

        let result = sar.update_hl(highs[0], lows[0]);
        assert!(result.is_nan());

        let result = sar.update_hl(highs[1], lows[1]);
        assert!(result > 0.0, "expected positive (long) SAR, got {}", result);
    }

    #[test]
    fn test_forced_start_short() {
        let mut sar = ParabolicStopAndReverse::new(&ParabolicStopAndReverseParams {
            start_value: -100.0,
            ..Default::default()
        }).unwrap();

        let highs = testdata::test_highs();
        let lows = testdata::test_lows();

        let result = sar.update_hl(highs[0], lows[0]);
        assert!(result.is_nan());

        let result = sar.update_hl(highs[1], lows[1]);
        assert!(result < 0.0, "expected negative (short) SAR, got {}", result);
    }
}
