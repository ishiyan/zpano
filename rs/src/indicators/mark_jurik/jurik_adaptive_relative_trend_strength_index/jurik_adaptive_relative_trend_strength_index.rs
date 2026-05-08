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

pub struct JurikAdaptiveRelativeTrendStrengthIndexParams {
    pub lo_length: usize,
    pub hi_length: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for JurikAdaptiveRelativeTrendStrengthIndexParams {
    fn default() -> Self {
        Self {
            lo_length: 5,
            hi_length: 30,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum JurikAdaptiveRelativeTrendStrengthIndexOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

#[derive(Debug)]
pub struct JurikAdaptiveRelativeTrendStrengthIndex {
    primed: bool,
    lo_length: usize,
    hi_length: usize,
    eps: f64,
    bar_count: usize,
    previous_price: f64,

    // Rolling buffers for adaptive length.
    long_buffer: [f64; 100],
    long_index: usize,
    long_sum: f64,
    long_count: usize,
    short_buffer: [f64; 10],
    short_index: usize,
    short_sum: f64,
    short_count: usize,

    // RSX core state.
    kg: f64,
    c: f64,
    warmup: usize,
    sig1a: f64,
    sig1b: f64,
    sig2a: f64,
    sig2b: f64,
    sig3a: f64,
    sig3b: f64,
    den1a: f64,
    den1b: f64,
    den2a: f64,
    den2b: f64,
    den3a: f64,
    den3b: f64,

    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
    description: String,
}

impl JurikAdaptiveRelativeTrendStrengthIndex {
    pub fn new(params: &JurikAdaptiveRelativeTrendStrengthIndexParams) -> Result<Self, String> {
        if params.lo_length < 2 {
            return Err("invalid jurik adaptive relative trend strength index parameters: lo_length should be at least 2".to_string());
        }
        if params.hi_length < params.lo_length {
            return Err("invalid jurik adaptive relative trend strength index parameters: hi_length should be at least lo_length".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let triple = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!("jarsx({}, {}{})", params.lo_length, params.hi_length, triple);
        let description = format!("Jurik adaptive relative trend strength index {}", mnemonic);

        Ok(Self {
            primed: false,
            lo_length: params.lo_length,
            hi_length: params.hi_length,
            eps: 0.001,
            bar_count: 0,
            previous_price: 0.0,
            long_buffer: [0.0; 100],
            long_index: 0,
            long_sum: 0.0,
            long_count: 0,
            short_buffer: [0.0; 10],
            short_index: 0,
            short_sum: 0.0,
            short_count: 0,
            kg: 0.0,
            c: 0.0,
            warmup: 0,
            sig1a: 0.0,
            sig1b: 0.0,
            sig2a: 0.0,
            sig2b: 0.0,
            sig3a: 0.0,
            sig3b: 0.0,
            den1a: 0.0,
            den1b: 0.0,
            den2a: 0.0,
            den2b: 0.0,
            den3a: 0.0,
            den3b: 0.0,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
            description,
        })
    }

    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let bar = self.bar_count;
        self.bar_count += 1;

        if bar == 0 {
            self.previous_price = sample;

            // First bar: add 0 to both buffers.
            self.long_buffer[0] = 0.0;
            self.long_sum = 0.0;
            self.long_count = 1;
            self.short_buffer[0] = 0.0;
            self.short_sum = 0.0;
            self.short_count = 1;

            // Compute adaptive length from bar 0.
            let avg1 = 0.0_f64;
            let avg2 = 0.0_f64;
            let value2 = ((self.eps + avg1) / (self.eps + avg2)).ln();
            let value3 = value2 / (1.0 + value2.abs());
            let adaptive_length = self.lo_length as f64
                + (self.hi_length - self.lo_length) as f64 * (1.0 + value3) / 2.0;
            let mut length = adaptive_length as usize;
            if length < 2 {
                length = 2;
            }

            self.kg = 3.0 / (length + 2) as f64;
            self.c = 1.0 - self.kg;
            self.warmup = if length - 1 > 5 { length - 1 } else { 5 };

            return f64::NAN;
        }

        // Bars 1+
        let old_price = self.previous_price;
        self.previous_price = sample;
        let value1 = (sample - old_price).abs();

        // Update long rolling buffer.
        if self.long_count < 100 {
            self.long_buffer[self.long_count] = value1;
            self.long_sum += value1;
            self.long_count += 1;
        } else {
            self.long_sum -= self.long_buffer[self.long_index];
            self.long_buffer[self.long_index] = value1;
            self.long_sum += value1;
            self.long_index = (self.long_index + 1) % 100;
        }

        // Update short rolling buffer.
        if self.short_count < 10 {
            self.short_buffer[self.short_count] = value1;
            self.short_sum += value1;
            self.short_count += 1;
        } else {
            self.short_sum -= self.short_buffer[self.short_index];
            self.short_buffer[self.short_index] = value1;
            self.short_sum += value1;
            self.short_index = (self.short_index + 1) % 10;
        }

        // RSX core computation.
        let mom = 100.0 * (sample - old_price);
        let abs_mom = mom.abs();

        let kg = self.kg;
        let c = self.c;

        // Signal path — Stage 1.
        self.sig1a = c * self.sig1a + kg * mom;
        self.sig1b = kg * self.sig1a + c * self.sig1b;
        let s1 = 1.5 * self.sig1a - 0.5 * self.sig1b;

        // Signal path — Stage 2.
        self.sig2a = c * self.sig2a + kg * s1;
        self.sig2b = kg * self.sig2a + c * self.sig2b;
        let s2 = 1.5 * self.sig2a - 0.5 * self.sig2b;

        // Signal path — Stage 3.
        self.sig3a = c * self.sig3a + kg * s2;
        self.sig3b = kg * self.sig3a + c * self.sig3b;
        let numerator = 1.5 * self.sig3a - 0.5 * self.sig3b;

        // Denominator path — Stage 1.
        self.den1a = c * self.den1a + kg * abs_mom;
        self.den1b = kg * self.den1a + c * self.den1b;
        let d1 = 1.5 * self.den1a - 0.5 * self.den1b;

        // Denominator path — Stage 2.
        self.den2a = c * self.den2a + kg * d1;
        self.den2b = kg * self.den2a + c * self.den2b;
        let d2 = 1.5 * self.den2a - 0.5 * self.den2b;

        // Denominator path — Stage 3.
        self.den3a = c * self.den3a + kg * d2;
        self.den3b = kg * self.den3a + c * self.den3b;
        let denominator = 1.5 * self.den3a - 0.5 * self.den3b;

        // Output after warmup.
        if bar >= self.warmup {
            self.primed = true;
            let value = if denominator != 0.0 {
                (numerator / denominator + 1.0) * 50.0
            } else {
                50.0
            };
            return value.max(0.0).min(100.0);
        }

        f64::NAN
    }
}

impl Indicator for JurikAdaptiveRelativeTrendStrengthIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikAdaptiveRelativeTrendStrengthIndex,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let val = self.update((self.bar_func)(bar));
        vec![Box::new(Scalar::new(bar.time, val))]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let val = self.update((self.quote_func)(quote));
        vec![Box::new(Scalar::new(quote.time, val))]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let val = self.update((self.trade_func)(trade));
        vec![Box::new(Scalar::new(trade.time, val))]
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let val = self.update(scalar.value);
        vec![Box::new(Scalar::new(scalar.time, val))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::mark_jurik::jurik_adaptive_relative_trend_strength_index::testdata::testdata;

    const EPSILON: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64) -> bool {
        if a.is_nan() && b.is_nan() {
            return true;
        }
        if a.is_nan() || b.is_nan() {
            return false;
        }
        (a - b).abs() < EPSILON
    }

    fn run_test(lo: usize, hi: usize, expected: &[f64]) {
        let params = JurikAdaptiveRelativeTrendStrengthIndexParams {
            lo_length: lo,
            hi_length: hi,
            ..Default::default()
        };
        let mut ind = JurikAdaptiveRelativeTrendStrengthIndex::new(&params).unwrap();
        let input = testdata::test_input();
        for (i, &val) in input.iter().enumerate() {
            let result = ind.update(val);
            assert!(
                almost_equal(result, expected[i]),
                "bar {}: expected {}, got {} (lo={}, hi={})",
                i, expected[i], result, lo, hi
            );
        }
    }

    #[test]
    fn test_jarsx_lo2_hi15() {
        run_test(2, 15, &testdata::expected_lo2_hi15());
    }

    #[test]
    fn test_jarsx_lo2_hi30() {
        run_test(2, 30, &testdata::expected_lo2_hi30());
    }

    #[test]
    fn test_jarsx_lo2_hi60() {
        run_test(2, 60, &testdata::expected_lo2_hi60());
    }

    #[test]
    fn test_jarsx_lo5_hi15() {
        run_test(5, 15, &testdata::expected_lo5_hi15());
    }

    #[test]
    fn test_jarsx_lo5_hi30() {
        run_test(5, 30, &testdata::expected_lo5_hi30());
    }

    #[test]
    fn test_jarsx_lo5_hi60() {
        run_test(5, 60, &testdata::expected_lo5_hi60());
    }

    #[test]
    fn test_jarsx_lo10_hi15() {
        run_test(10, 15, &testdata::expected_lo10_hi15());
    }

    #[test]
    fn test_jarsx_lo10_hi30() {
        run_test(10, 30, &testdata::expected_lo10_hi30());
    }

    #[test]
    fn test_jarsx_lo10_hi60() {
        run_test(10, 60, &testdata::expected_lo10_hi60());
    }
}
