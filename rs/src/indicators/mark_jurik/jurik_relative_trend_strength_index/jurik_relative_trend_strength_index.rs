use crate::entities::bar::Bar;
use crate::entities::bar_component::{
    component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT,
};
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
use crate::indicators::core::line_indicator::{BarFunc, QuoteFunc, TradeFunc};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Jurik Relative Trend Strength Index indicator.
pub struct JurikRelativeTrendStrengthIndexParams {
    /// Length (number of time periods). Must be >= 2. Default 14.
    pub length: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for JurikRelativeTrendStrengthIndexParams {
    fn default() -> Self {
        Self {
            length: 14,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Jurik Relative Trend Strength Index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum JurikRelativeTrendStrengthIndexOutput {
    RelativeTrendStrengthIndex = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Mark Jurik's Relative Trend Strength Index (RSX/JRSX).
#[derive(Debug)]
pub struct JurikRelativeTrendStrengthIndex {
    primed: bool,
    param_len: usize,
    f0: i32,
    f88: i32,
    f90: i32,
    f8: f64,
    f10: f64,
    f18: f64,
    f20: f64,
    f28: f64,
    f30: f64,
    f38: f64,
    f40: f64,
    f48: f64,
    f50: f64,
    f58: f64,
    f60: f64,
    f68: f64,
    f70: f64,
    f78: f64,
    f80: f64,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
    description: String,
}

impl JurikRelativeTrendStrengthIndex {
    pub fn new(params: &JurikRelativeTrendStrengthIndexParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err(
                "invalid jurik relative trend strength index parameters: length must be >= 2"
                    .to_string(),
            );
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "jrsx({}{})",
            params.length,
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Jurik relative trend strength index {}", mnemonic);

        Ok(Self {
            primed: false,
            param_len: params.length,
            f0: 0,
            f88: 0,
            f90: 0,
            f8: 0.0,
            f10: 0.0,
            f18: 0.0,
            f20: 0.0,
            f28: 0.0,
            f30: 0.0,
            f38: 0.0,
            f40: 0.0,
            f48: 0.0,
            f50: 0.0,
            f58: 0.0,
            f60: 0.0,
            f68: 0.0,
            f70: 0.0,
            f78: 0.0,
            f80: 0.0,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
            description,
        })
    }

    /// Core update. Returns the RSX value.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        const HUNDRED: f64 = 100.0;
        const FIFTY: f64 = 50.0;
        const ONE_FIVE: f64 = 1.5;
        const HALF: f64 = 0.5;
        const MIN_LEN: i32 = 5;
        const EPS: f64 = 1e-10;

        let length = self.param_len;

        if self.f90 == 0 {
            self.f90 = 1;
            self.f0 = 0;
            if (length as i32 - 1) >= MIN_LEN {
                self.f88 = length as i32 - 1;
            } else {
                self.f88 = MIN_LEN;
            }
            self.f8 = HUNDRED * sample;
            self.f18 = 3.0 / (length as f64 + 2.0);
            self.f20 = 1.0 - self.f18;
        } else {
            if self.f88 <= self.f90 {
                self.f90 = self.f88 + 1;
            } else {
                self.f90 += 1;
            }
            self.f10 = self.f8;
            self.f8 = HUNDRED * sample;
            let v8 = self.f8 - self.f10;

            self.f28 = self.f20 * self.f28 + self.f18 * v8;
            self.f30 = self.f18 * self.f28 + self.f20 * self.f30;
            let v_c = self.f28 * ONE_FIVE - self.f30 * HALF;

            self.f38 = self.f20 * self.f38 + self.f18 * v_c;
            self.f40 = self.f18 * self.f38 + self.f20 * self.f40;
            let v10 = self.f38 * ONE_FIVE - self.f40 * HALF;

            self.f48 = self.f20 * self.f48 + self.f18 * v10;
            self.f50 = self.f18 * self.f48 + self.f20 * self.f50;
            let v14 = self.f48 * ONE_FIVE - self.f50 * HALF;

            self.f58 = self.f20 * self.f58 + self.f18 * v8.abs();
            self.f60 = self.f18 * self.f58 + self.f20 * self.f60;
            let v18 = self.f58 * ONE_FIVE - self.f60 * HALF;

            self.f68 = self.f20 * self.f68 + self.f18 * v18;
            self.f70 = self.f18 * self.f68 + self.f20 * self.f70;
            let v1c = self.f68 * ONE_FIVE - self.f70 * HALF;

            self.f78 = self.f20 * self.f78 + self.f18 * v1c;
            self.f80 = self.f18 * self.f78 + self.f20 * self.f80;
            let v20 = self.f78 * ONE_FIVE - self.f80 * HALF;

            if self.f88 >= self.f90 && self.f8 != self.f10 {
                self.f0 = 1;
            }
            if self.f88 == self.f90 && self.f0 == 0 {
                self.f90 = 0;
            }

            if self.f88 < self.f90 && v20 > EPS {
                let mut v4 = (v14 / v20 + 1.0) * FIFTY;
                if v4 > HUNDRED {
                    v4 = HUNDRED;
                }
                if v4 < 0.0 {
                    v4 = 0.0;
                }
                self.primed = true;
                return v4;
            }
        }

        if self.f88 < self.f90 {
            self.primed = true;
        }
        if !self.primed {
            return f64::NAN;
        }
        FIFTY
    }
}

impl Indicator for JurikRelativeTrendStrengthIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikRelativeTrendStrengthIndex,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let val = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, val))]
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        let val = self.update(sample);
        vec![Box::new(Scalar::new(bar.time, val))]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.quote_func)(quote);
        let val = self.update(sample);
        vec![Box::new(Scalar::new(quote.time, val))]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.trade_func)(trade);
        let val = self.update(sample);
        vec![Box::new(Scalar::new(trade.time, val))]
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    use crate::indicators::core::indicator::Indicator;
    use crate::indicators::core::outputs::shape::Shape;

    const TOLERANCE: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() <= TOLERANCE
    }

    fn run_rsx_test(length: usize, expected: &[f64]) {
        let params = JurikRelativeTrendStrengthIndexParams { length, ..Default::default() };
        let mut rsx = JurikRelativeTrendStrengthIndex::new(&params).unwrap();
        let input = testdata::test_input();
        for (i, &val) in input.iter().enumerate() {
            let result = rsx.update(val);
            if expected[i].is_nan() {
                assert!(result.is_nan(), "bar {}: expected NaN, got {}", i, result);
            } else {
                assert!(
                    almost_equal(result, expected[i]),
                    "bar {}: expected {}, got {}, diff {}",
                    i, expected[i], result, (result - expected[i]).abs()
                );
            }
        }
    }

    #[test] fn test_rsx_length_2() { run_rsx_test(2, &testdata::expected_length2()); }
    #[test] fn test_rsx_length_3() { run_rsx_test(3, &testdata::expected_length3()); }
    #[test] fn test_rsx_length_4() { run_rsx_test(4, &testdata::expected_length4()); }
    #[test] fn test_rsx_length_5() { run_rsx_test(5, &testdata::expected_length5()); }
    #[test] fn test_rsx_length_6() { run_rsx_test(6, &testdata::expected_length6()); }
    #[test] fn test_rsx_length_7() { run_rsx_test(7, &testdata::expected_length7()); }
    #[test] fn test_rsx_length_8() { run_rsx_test(8, &testdata::expected_length8()); }
    #[test] fn test_rsx_length_9() { run_rsx_test(9, &testdata::expected_length9()); }
    #[test] fn test_rsx_length_10() { run_rsx_test(10, &testdata::expected_length10()); }
    #[test] fn test_rsx_length_11() { run_rsx_test(11, &testdata::expected_length11()); }
    #[test] fn test_rsx_length_12() { run_rsx_test(12, &testdata::expected_length12()); }
    #[test] fn test_rsx_length_13() { run_rsx_test(13, &testdata::expected_length13()); }
    #[test] fn test_rsx_length_14() { run_rsx_test(14, &testdata::expected_length14()); }
    #[test] fn test_rsx_length_15() { run_rsx_test(15, &testdata::expected_length15()); }

    #[test]
    fn test_rsx_metadata() {
        let params = JurikRelativeTrendStrengthIndexParams::default();
        let rsx = JurikRelativeTrendStrengthIndex::new(&params).unwrap();
        let md = rsx.metadata();
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_rsx_invalid_params() {
        let params = JurikRelativeTrendStrengthIndexParams { length: 1, ..Default::default() };
        assert!(JurikRelativeTrendStrengthIndex::new(&params).is_err());
    }
}
