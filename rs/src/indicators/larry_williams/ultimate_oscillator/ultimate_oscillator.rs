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

/// Parameters for the Ultimate Oscillator indicator.
pub struct UltimateOscillatorParams {
    /// First (shortest) period. Default 7, minimum 2.
    pub length1: usize,
    /// Second (medium) period. Default 14, minimum 2.
    pub length2: usize,
    /// Third (longest) period. Default 28, minimum 2.
    pub length3: usize,
}

impl Default for UltimateOscillatorParams {
    fn default() -> Self {
        Self { length1: 7, length2: 14, length3: 28 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum UltimateOscillatorOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Larry Williams' Ultimate Oscillator.
///
/// Combines three time periods (weighted 4:2:1) into a single oscillator
/// measuring buying pressure relative to true range.
pub struct UltimateOscillator {
    p1: usize,
    p2: usize,
    p3: usize,
    previous_close: f64,
    bp_buffer: Vec<f64>,
    tr_buffer: Vec<f64>,
    buffer_index: usize,
    bp_sum1: f64,
    bp_sum2: f64,
    bp_sum3: f64,
    tr_sum1: f64,
    tr_sum2: f64,
    tr_sum3: f64,
    count: usize,
    primed: bool,
    mnemonic: String,
}

const WEIGHT1: f64 = 4.0;
const WEIGHT2: f64 = 2.0;
const WEIGHT3: f64 = 1.0;
const TOTAL_WEIGHT: f64 = WEIGHT1 + WEIGHT2 + WEIGHT3;

fn sort_three(a: usize, b: usize, c: usize) -> (usize, usize, usize) {
    let (mut a, mut b, mut c) = (a, b, c);
    if a > b { std::mem::swap(&mut a, &mut b); }
    if b > c { std::mem::swap(&mut b, &mut c); }
    if a > b { std::mem::swap(&mut a, &mut b); }
    (a, b, c)
}

impl UltimateOscillator {
    pub fn new(params: &UltimateOscillatorParams) -> Result<Self, String> {
        let l1 = if params.length1 == 0 { 7 } else { params.length1 };
        let l2 = if params.length2 == 0 { 14 } else { params.length2 };
        let l3 = if params.length3 == 0 { 28 } else { params.length3 };

        if l1 < 2 { return Err(format!("length1 must be >= 2, got {}", l1)); }
        if l2 < 2 { return Err(format!("length2 must be >= 2, got {}", l2)); }
        if l3 < 2 { return Err(format!("length3 must be >= 2, got {}", l3)); }

        let (s1, s2, s3) = sort_three(l1, l2, l3);
        let mnemonic = format!("ultosc({}, {}, {})", l1, l2, l3);

        Ok(Self {
            p1: s1,
            p2: s2,
            p3: s3,
            previous_close: f64::NAN,
            bp_buffer: vec![0.0; s3],
            tr_buffer: vec![0.0; s3],
            buffer_index: 0,
            bp_sum1: 0.0, bp_sum2: 0.0, bp_sum3: 0.0,
            tr_sum1: 0.0, tr_sum2: 0.0, tr_sum3: 0.0,
            count: 0,
            primed: false,
            mnemonic,
        })
    }

    /// Core update with close, high, low.
    pub fn update(&mut self, close: f64, high: f64, low: f64) -> f64 {
        if close.is_nan() || high.is_nan() || low.is_nan() {
            return f64::NAN;
        }

        // First bar: just store close.
        if self.previous_close.is_nan() {
            self.previous_close = close;
            return f64::NAN;
        }

        let true_low = low.min(self.previous_close);
        let bp = close - true_low;

        let mut tr = high - low;
        let d1 = (self.previous_close - high).abs();
        if d1 > tr { tr = d1; }
        let d2 = (self.previous_close - low).abs();
        if d2 > tr { tr = d2; }

        self.previous_close = close;
        self.count += 1;

        // Remove trailing values BEFORE storing new value.
        if self.count > self.p1 {
            let old = (self.buffer_index + self.p3 - self.p1) % self.p3;
            self.bp_sum1 -= self.bp_buffer[old];
            self.tr_sum1 -= self.tr_buffer[old];
        }
        if self.count > self.p2 {
            let old = (self.buffer_index + self.p3 - self.p2) % self.p3;
            self.bp_sum2 -= self.bp_buffer[old];
            self.tr_sum2 -= self.tr_buffer[old];
        }
        if self.count > self.p3 {
            let old = (self.buffer_index + self.p3 - self.p3) % self.p3;
            self.bp_sum3 -= self.bp_buffer[old];
            self.tr_sum3 -= self.tr_buffer[old];
        }

        self.bp_sum1 += bp;
        self.bp_sum2 += bp;
        self.bp_sum3 += bp;
        self.tr_sum1 += tr;
        self.tr_sum2 += tr;
        self.tr_sum3 += tr;

        self.bp_buffer[self.buffer_index] = bp;
        self.tr_buffer[self.buffer_index] = tr;

        self.buffer_index = (self.buffer_index + 1) % self.p3;

        if self.count < self.p3 {
            return f64::NAN;
        }

        self.primed = true;

        let mut output = 0.0;
        if self.tr_sum1 != 0.0 { output += WEIGHT1 * (self.bp_sum1 / self.tr_sum1); }
        if self.tr_sum2 != 0.0 { output += WEIGHT2 * (self.bp_sum2 / self.tr_sum2); }
        if self.tr_sum3 != 0.0 { output += WEIGHT3 * (self.bp_sum3 / self.tr_sum3); }

        100.0 * (output / TOTAL_WEIGHT)
    }
}

impl Indicator for UltimateOscillator {
    fn is_primed(&self) -> bool { self.primed }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::UltimateOscillator,
            &self.mnemonic,
            &format!("Ultimate Oscillator {}", self.mnemonic),
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: format!("Ultimate Oscillator {}", self.mnemonic),
            }],
        )
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let v = scalar.value;
        vec![Box::new(Scalar { time: scalar.time, value: self.update(v, v, v) })]
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        vec![Box::new(Scalar { time: bar.time, value: self.update(bar.close, bar.high, bar.low) })]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let v = (quote.bid_price + quote.ask_price) / 2.0;
        vec![Box::new(Scalar { time: quote.time, value: self.update(v, v, v) })]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let v = trade.price;
        vec![Box::new(Scalar { time: trade.time, value: self.update(v, v, v) })]
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    const TOLERANCE: f64 = 1e-4;

    #[test]
    fn test_update_all_252() {
        let highs = testdata::test_input_high();
        let lows = testdata::test_input_low();
        let closes = testdata::test_input_close();
        let expected = testdata::test_expected();

        let mut ind = UltimateOscillator::new(&UltimateOscillatorParams::default()).unwrap();

        for i in 0..highs.len() {
            let result = ind.update(closes[i], highs[i], lows[i]);
            if expected[i].is_nan() {
                assert!(result.is_nan(), "index {}: expected NaN, got {}", i, result);
            } else {
                assert!(!result.is_nan(), "index {}: expected {}, got NaN", i, expected[i]);
                assert!((result - expected[i]).abs() <= TOLERANCE,
                    "index {}: expected {}, got {} (diff {})", i, expected[i], result, (result - expected[i]).abs());
            }
        }
    }

    #[test]
    fn test_is_primed() {
        let highs = testdata::test_input_high();
        let lows = testdata::test_input_low();
        let closes = testdata::test_input_close();

        let mut ind = UltimateOscillator::new(&UltimateOscillatorParams::default()).unwrap();
        for i in 0..28 {
            ind.update(closes[i], highs[i], lows[i]);
            assert!(!ind.is_primed(), "should not be primed at index {}", i);
        }
        ind.update(closes[28], highs[28], lows[28]);
        assert!(ind.is_primed(), "should be primed at index 28");

        let mut ind2 = UltimateOscillator::new(&UltimateOscillatorParams { length1: 2, length2: 3, length3: 4 }).unwrap();
        for i in 0..4 {
            ind2.update(closes[i], highs[i], lows[i]);
            assert!(!ind2.is_primed());
        }
        ind2.update(closes[4], highs[4], lows[4]);
        assert!(ind2.is_primed());
    }

    #[test]
    fn test_nan_passthrough() {
        let mut ind = UltimateOscillator::new(&UltimateOscillatorParams::default()).unwrap();
        assert!(ind.update(f64::NAN, 100.0, 90.0).is_nan());
        assert!(ind.update(95.0, f64::NAN, 90.0).is_nan());
        assert!(ind.update(95.0, 100.0, f64::NAN).is_nan());
    }

    #[test]
    fn test_constructor_validation() {
        assert!(UltimateOscillator::new(&UltimateOscillatorParams::default()).is_ok());
        assert!(UltimateOscillator::new(&UltimateOscillatorParams { length1: 5, length2: 10, length3: 20 }).is_ok());
        assert!(UltimateOscillator::new(&UltimateOscillatorParams { length1: 1, length2: 14, length3: 28 }).is_err());
        assert!(UltimateOscillator::new(&UltimateOscillatorParams { length1: 7, length2: 1, length3: 28 }).is_err());
        assert!(UltimateOscillator::new(&UltimateOscillatorParams { length1: 7, length2: 14, length3: 1 }).is_err());
    }

    #[test]
    fn test_metadata() {
        let ind = UltimateOscillator::new(&UltimateOscillatorParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::UltimateOscillator);
        assert_eq!(meta.mnemonic, "ultosc(7, 14, 28)");
        assert_eq!(meta.outputs.len(), 1);
    }

    #[test]
    fn test_spot_checks() {
        let highs = testdata::test_input_high();
        let lows = testdata::test_input_low();
        let closes = testdata::test_input_close();

        let mut ind = UltimateOscillator::new(&UltimateOscillatorParams::default()).unwrap();
        let mut results = Vec::new();
        for i in 0..highs.len() {
            results.push(ind.update(closes[i], highs[i], lows[i]));
        }

        assert!((results[28] - 47.1713).abs() <= TOLERANCE);
        assert!((results[29] - 46.2802).abs() <= TOLERANCE);
        assert!((results[251] - 40.0854).abs() <= TOLERANCE);
    }
}
