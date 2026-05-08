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

/// Parameters to create an instance of the Jurik Turning Point Oscillator.
pub struct JurikTurningPointOscillatorParams {
    /// Length controls the lookback window for the Spearman rank correlation. Must be >= 2. Default 14.
    pub length: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for JurikTurningPointOscillatorParams {
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

/// Enumerates the outputs of the Jurik Turning Point Oscillator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum JurikTurningPointOscillatorOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Mark Jurik's Turning Point Oscillator (JTPO).
/// Computes Spearman rank correlation between price ranks and time positions.
/// Output is in [-1, +1].
#[derive(Debug)]
pub struct JurikTurningPointOscillator {
    primed: bool,
    length: usize,
    buffer: Vec<f64>,
    buf_idx: usize,
    count: usize,
    f18: f64,
    mid: f64,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
    description: String,
}

impl JurikTurningPointOscillator {
    pub fn new(params: &JurikTurningPointOscillatorParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err(
                "invalid jurik turning point oscillator parameters: length should be at least 2"
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
            "jtpo({}{})",
            params.length,
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Jurik turning point oscillator {}", mnemonic);

        let n = params.length as f64;

        Ok(Self {
            primed: false,
            length: params.length,
            buffer: vec![0.0; params.length],
            buf_idx: 0,
            count: 0,
            f18: 12.0 / (n * (n - 1.0) * (n + 1.0)),
            mid: (n + 1.0) / 2.0,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
            description,
        })
    }

    /// Core update. Returns the JTPO value.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let length = self.length;

        self.buffer[self.buf_idx] = sample;
        self.buf_idx = (self.buf_idx + 1) % length;
        self.count += 1;

        if self.count < length {
            return f64::NAN;
        }

        // Extract window in chronological order.
        let mut window = vec![0.0; length];
        for i in 0..length {
            window[i] = self.buffer[(self.buf_idx + i) % length];
        }

        // Check if all values are identical.
        let all_same = window[1..].iter().all(|&v| v == window[0]);

        if all_same {
            if !self.primed {
                self.primed = true;
            }
            return f64::NAN;
        }

        // Build indices sorted by price.
        let mut items: Vec<(usize, f64)> = window.iter().enumerate().map(|(i, &p)| (i, p)).collect();
        items.sort_by(|a, b| a.1.partial_cmp(&b.1).unwrap());

        // arr2[i] = original time position (1-based) of the i-th sorted element.
        let arr2: Vec<f64> = items.iter().map(|(idx, _)| (*idx + 1) as f64).collect();

        // Assign fractional ranks for ties.
        let mut arr3 = vec![0.0; length];
        let mut i = 0;
        while i < length {
            let mut j = i;
            while j < length - 1 && items[j + 1].1 == items[j].1 {
                j += 1;
            }
            let avg_rank = (i + 1 + j + 1) as f64 / 2.0;
            for k in i..=j {
                arr3[k] = avg_rank;
            }
            i = j + 1;
        }

        // Compute correlation sum.
        let mut corr_sum = 0.0;
        for i in 0..length {
            corr_sum += (arr3[i] - self.mid) * (arr2[i] - self.mid);
        }

        if !self.primed {
            self.primed = true;
        }

        self.f18 * corr_sum
    }
}

impl Indicator for JurikTurningPointOscillator {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikTurningPointOscillator,
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

    fn run_jtpo_test(length: usize, expected: &[f64]) {
        let params = JurikTurningPointOscillatorParams { length, ..Default::default() };
        let mut ind = JurikTurningPointOscillator::new(&params).unwrap();
        let input = testdata::test_input();
        for (i, &val) in input.iter().enumerate() {
            let result = ind.update(val);
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

    #[test] fn test_jtpo_len5() { run_jtpo_test(5, &testdata::expected_len5()); }
    #[test] fn test_jtpo_len7() { run_jtpo_test(7, &testdata::expected_len7()); }
    #[test] fn test_jtpo_len10() { run_jtpo_test(10, &testdata::expected_len10()); }
    #[test] fn test_jtpo_len14() { run_jtpo_test(14, &testdata::expected_len14()); }
    #[test] fn test_jtpo_len20() { run_jtpo_test(20, &testdata::expected_len20()); }
    #[test] fn test_jtpo_len28() { run_jtpo_test(28, &testdata::expected_len28()); }
    #[test] fn test_jtpo_len40() { run_jtpo_test(40, &testdata::expected_len40()); }
    #[test] fn test_jtpo_len60() { run_jtpo_test(60, &testdata::expected_len60()); }
    #[test] fn test_jtpo_len80() { run_jtpo_test(80, &testdata::expected_len80()); }

    #[test]
    fn test_jtpo_metadata() {
        let params = JurikTurningPointOscillatorParams::default();
        let ind = JurikTurningPointOscillator::new(&params).unwrap();
        let md = ind.metadata();
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_jtpo_invalid_params() {
        let params = JurikTurningPointOscillatorParams { length: 1, ..Default::default() };
        assert!(JurikTurningPointOscillator::new(&params).is_err());
    }
}
