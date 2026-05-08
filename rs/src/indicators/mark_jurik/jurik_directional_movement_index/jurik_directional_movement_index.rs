use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;

use super::super::jurik_moving_average::{JurikMovingAverage, JurikMovingAverageParams};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Jurik Directional Movement Index indicator.
pub struct JurikDirectionalMovementIndexParams {
    /// Length (number of time periods). Must be >= 1. Default 14.
    pub length: usize,
}

impl Default for JurikDirectionalMovementIndexParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Jurik Directional Movement Index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum JurikDirectionalMovementIndexOutput {
    Bipolar = 1,
    Plus = 2,
    Minus = 3,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Mark Jurik's Jurik Directional Movement Index (DMX).
///
/// Produces three output lines:
/// - Bipolar: 100*(Plus-Minus)/(Plus+Minus)
/// - Plus: JMA(upward) / JMA(TrueRange)
/// - Minus: JMA(downward) / JMA(TrueRange)
///
/// The internal JMA instances use phase=-100 (maximum lag, no overshoot).
#[derive(Debug)]
pub struct JurikDirectionalMovementIndex {
    primed: bool,
    bar: usize,
    prev_high: f64,
    prev_low: f64,
    prev_close: f64,
    jma_plus: JurikMovingAverage,
    jma_minus: JurikMovingAverage,
    jma_denom: JurikMovingAverage,
    plus_val: f64,
    minus_val: f64,
    bipolar_val: f64,
    mnemonic: String,
    description: String,
}

impl JurikDirectionalMovementIndex {
    pub fn new(params: &JurikDirectionalMovementIndexParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err(
                "invalid jurik directional movement index parameters: length should be positive"
                    .to_string(),
            );
        }

        let jma_params = JurikMovingAverageParams {
            length: params.length,
            phase: -100,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        };

        let jma_plus = JurikMovingAverage::new(&jma_params)
            .map_err(|e| format!("invalid jurik directional movement index parameters: {}", e))?;
        let jma_minus = JurikMovingAverage::new(&jma_params)
            .map_err(|e| format!("invalid jurik directional movement index parameters: {}", e))?;
        let jma_denom = JurikMovingAverage::new(&jma_params)
            .map_err(|e| format!("invalid jurik directional movement index parameters: {}", e))?;

        let mnemonic = format!("jdmx({})", params.length);
        let description = format!("Jurik directional movement index {}", mnemonic);

        Ok(Self {
            primed: false,
            bar: 0,
            prev_high: f64::NAN,
            prev_low: f64::NAN,
            prev_close: f64::NAN,
            jma_plus,
            jma_minus,
            jma_denom,
            plus_val: f64::NAN,
            minus_val: f64::NAN,
            bipolar_val: f64::NAN,
            mnemonic,
            description,
        })
    }

    /// Core update. Returns (bipolar, plus, minus).
    pub fn update_hlc(&mut self, high: f64, low: f64, close: f64) -> (f64, f64, f64) {
        const WARMUP: usize = 41;
        const EPSILON: f64 = 0.00001;
        const HUNDRED: f64 = 100.0;

        self.bar += 1;

        let mut true_range = 0.0;
        let mut upward = 0.0;
        let mut downward = 0.0;

        if self.bar >= 2 {
            let v1 = HUNDRED * (high - self.prev_high);
            let v2 = HUNDRED * (self.prev_low - low);

            if v1 > v2 && v1 > 0.0 {
                upward = v1;
            }

            if v2 > v1 && v2 > 0.0 {
                downward = v2;
            }
        }

        if self.bar >= 3 {
            let m1 = (high - low).abs();
            let m2 = (high - self.prev_close).abs();
            let m3 = (low - self.prev_close).abs();
            true_range = m1.max(m2).max(m3);
        }

        self.prev_high = high;
        self.prev_low = low;
        self.prev_close = close;

        // Feed into JMA instances.
        let numer_plus = self.jma_plus.update(upward);
        let numer_minus = self.jma_minus.update(downward);
        let denom = self.jma_denom.update(true_range);

        if self.bar <= WARMUP {
            self.bipolar_val = f64::NAN;
            self.plus_val = f64::NAN;
            self.minus_val = f64::NAN;
            return (f64::NAN, f64::NAN, f64::NAN);
        }

        self.primed = true;

        // Compute Plus and Minus.
        if denom > EPSILON {
            self.plus_val = numer_plus / denom;
        } else {
            self.plus_val = 0.0;
        }

        if denom > EPSILON {
            self.minus_val = numer_minus / denom;
        } else {
            self.minus_val = 0.0;
        }

        // Compute Bipolar.
        let sum = self.plus_val + self.minus_val;
        if sum > EPSILON {
            self.bipolar_val = HUNDRED * (self.plus_val - self.minus_val) / sum;
        } else {
            self.bipolar_val = 0.0;
        }

        (self.bipolar_val, self.plus_val, self.minus_val)
    }

    fn update_entity(&mut self, time: i64, high: f64, low: f64, close: f64) -> Output {
        let (bipolar, plus, minus) = self.update_hlc(high, low, close);
        vec![
            Box::new(Scalar::new(time, bipolar)),
            Box::new(Scalar::new(time, plus)),
            Box::new(Scalar::new(time, minus)),
        ]
    }
}

impl Indicator for JurikDirectionalMovementIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikDirectionalMovementIndex,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{}:bipolar", self.mnemonic),
                    description: format!("{} bipolar", self.description),
                },
                OutputText {
                    mnemonic: format!("{}:plus", self.mnemonic),
                    description: format!("{} plus", self.description),
                },
                OutputText {
                    mnemonic: format!("{}:minus", self.mnemonic),
                    description: format!("{} minus", self.description),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = sample.value;
        self.update_entity(sample.time, v, v, v)
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        self.update_entity(bar.time, bar.high, bar.low, bar.close)
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        self.update_entity(
            quote.time,
            quote.ask_price,
            quote.bid_price,
            (quote.ask_price + quote.bid_price) / 2.0,
        )
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let v = trade.price;
        self.update_entity(trade.time, v, v, v)
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::super::testdata::testdata;
    use super::*;
    use crate::indicators::core::indicator::Indicator;
    use crate::indicators::core::outputs::shape::Shape;

    const TOLERANCE: f64 = 1e-10;

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() <= TOLERANCE
    }

    fn run_dmx_test(
        length: usize,
        expected_bipolar: &[f64],
        expected_plus: &[f64],
        check_minus: bool,
        expected_minus: &[f64],
    ) {
        let params = JurikDirectionalMovementIndexParams { length };
        let mut dmx = JurikDirectionalMovementIndex::new(&params).unwrap();
        let close = testdata::test_input_close();
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        for i in 0..252 {
            let (bipolar, plus, minus) = dmx.update_hlc(high[i], low[i], close[i]);
            if i <= 40 {
                assert!(
                    bipolar.is_nan(),
                    "bar {}: bipolar expected NaN, got {}",
                    i,
                    bipolar
                );
                continue;
            }
            assert!(
                almost_equal(bipolar, expected_bipolar[i]),
                "bar {} bipolar: expected {}, got {}, diff {}",
                i,
                expected_bipolar[i],
                bipolar,
                (bipolar - expected_bipolar[i]).abs()
            );
            assert!(
                almost_equal(plus, expected_plus[i]),
                "bar {} plus: expected {}, got {}, diff {}",
                i,
                expected_plus[i],
                plus,
                (plus - expected_plus[i]).abs()
            );
            if check_minus {
                assert!(
                    almost_equal(minus, expected_minus[i]),
                    "bar {} minus: expected {}, got {}, diff {}",
                    i,
                    expected_minus[i],
                    minus,
                    (minus - expected_minus[i]).abs()
                );
            }
        }
    }

    #[test]
    fn test_dmx_length_2() {
        run_dmx_test(
            2,
            &testdata::dmx_bipolar_len2(),
            &testdata::dmx_plus_len2(),
            true,
            &testdata::dmx_minus_len2(),
        );
    }
    #[test]
    fn test_dmx_length_3() {
        run_dmx_test(
            3,
            &testdata::dmx_bipolar_len3(),
            &testdata::dmx_plus_len3(),
            true,
            &testdata::dmx_minus_len3(),
        );
    }
    #[test]
    fn test_dmx_length_4() {
        run_dmx_test(
            4,
            &testdata::dmx_bipolar_len4(),
            &testdata::dmx_plus_len4(),
            true,
            &testdata::dmx_minus_len4(),
        );
    }
    #[test]
    fn test_dmx_length_5() {
        run_dmx_test(
            5,
            &testdata::dmx_bipolar_len5(),
            &testdata::dmx_plus_len5(),
            true,
            &testdata::dmx_minus_len5(),
        );
    }
    #[test]
    fn test_dmx_length_6() {
        run_dmx_test(
            6,
            &testdata::dmx_bipolar_len6(),
            &testdata::dmx_plus_len6(),
            true,
            &testdata::dmx_minus_len6(),
        );
    }
    #[test]
    fn test_dmx_length_7() {
        run_dmx_test(
            7,
            &testdata::dmx_bipolar_len7(),
            &testdata::dmx_plus_len7(),
            true,
            &testdata::dmx_minus_len7(),
        );
    }
    #[test]
    fn test_dmx_length_8() {
        run_dmx_test(
            8,
            &testdata::dmx_bipolar_len8(),
            &testdata::dmx_plus_len8(),
            true,
            &testdata::dmx_minus_len8(),
        );
    }
    #[test]
    fn test_dmx_length_9() {
        run_dmx_test(
            9,
            &testdata::dmx_bipolar_len9(),
            &testdata::dmx_plus_len9(),
            true,
            &testdata::dmx_minus_len9(),
        );
    }
    #[test]
    fn test_dmx_length_10() {
        run_dmx_test(
            10,
            &testdata::dmx_bipolar_len10(),
            &testdata::dmx_plus_len10(),
            true,
            &testdata::dmx_minus_len10(),
        );
    }
    #[test]
    fn test_dmx_length_11() {
        run_dmx_test(
            11,
            &testdata::dmx_bipolar_len11(),
            &testdata::dmx_plus_len11(),
            true,
            &testdata::dmx_minus_len11(),
        );
    }
    #[test]
    fn test_dmx_length_12() {
        run_dmx_test(
            12,
            &testdata::dmx_bipolar_len12(),
            &testdata::dmx_plus_len12(),
            true,
            &testdata::dmx_minus_len12(),
        );
    }
    #[test]
    fn test_dmx_length_13() {
        run_dmx_test(
            13,
            &testdata::dmx_bipolar_len13(),
            &testdata::dmx_plus_len13(),
            true,
            &testdata::dmx_minus_len13(),
        );
    }
    #[test]
    fn test_dmx_length_14() {
        run_dmx_test(
            14,
            &testdata::dmx_bipolar_len14(),
            &testdata::dmx_plus_len14(),
            false,
            &testdata::dmx_minus_len14(),
        );
    }
    #[test]
    fn test_dmx_length_15() {
        run_dmx_test(
            15,
            &testdata::dmx_bipolar_len15(),
            &testdata::dmx_plus_len15(),
            true,
            &testdata::dmx_minus_len15(),
        );
    }
    #[test]
    fn test_dmx_length_16() {
        run_dmx_test(
            16,
            &testdata::dmx_bipolar_len16(),
            &testdata::dmx_plus_len16(),
            true,
            &testdata::dmx_minus_len16(),
        );
    }
    #[test]
    fn test_dmx_length_17() {
        run_dmx_test(
            17,
            &testdata::dmx_bipolar_len17(),
            &testdata::dmx_plus_len17(),
            true,
            &testdata::dmx_minus_len17(),
        );
    }
    #[test]
    fn test_dmx_length_18() {
        run_dmx_test(
            18,
            &testdata::dmx_bipolar_len18(),
            &testdata::dmx_plus_len18(),
            true,
            &testdata::dmx_minus_len18(),
        );
    }
    #[test]
    fn test_dmx_length_19() {
        run_dmx_test(
            19,
            &testdata::dmx_bipolar_len19(),
            &testdata::dmx_plus_len19(),
            true,
            &testdata::dmx_minus_len19(),
        );
    }
    #[test]
    fn test_dmx_length_20() {
        run_dmx_test(
            20,
            &testdata::dmx_bipolar_len20(),
            &testdata::dmx_plus_len20(),
            true,
            &testdata::dmx_minus_len20(),
        );
    }

    #[test]
    fn test_dmx_metadata() {
        let params = JurikDirectionalMovementIndexParams { length: 14 };
        let dmx = JurikDirectionalMovementIndex::new(&params).unwrap();
        let md = dmx.metadata();
        assert_eq!(md.outputs.len(), 3);
        assert_eq!(md.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_dmx_invalid_params() {
        let params = JurikDirectionalMovementIndexParams { length: 0 };
        assert!(JurikDirectionalMovementIndex::new(&params).is_err());
    }
}
