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

use super::super::jurik_moving_average::{JurikMovingAverage, JurikMovingAverageParams};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

pub struct JurikCommodityChannelIndexParams {
    pub length: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for JurikCommodityChannelIndexParams {
    fn default() -> Self {
        Self {
            length: 20,
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
pub enum JurikCommodityChannelIndexOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

#[derive(Debug)]
pub struct JurikCommodityChannelIndex {
    primed: bool,
    fast_jma: JurikMovingAverage,
    slow_jma: JurikMovingAverage,
    diff_buffer: Vec<f64>,
    diff_buf_size: usize,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
    description: String,
}

impl JurikCommodityChannelIndex {
    pub fn new(params: &JurikCommodityChannelIndexParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid jurik commodity channel index parameters: length must be >= 2".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let triple = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!("jccx({}{})", params.length, triple);
        let description = format!("Jurik commodity channel index {}", mnemonic);

        let fast_params = JurikMovingAverageParams {
            length: 4,
            phase: 0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        };
        let slow_params = JurikMovingAverageParams {
            length: params.length,
            phase: 0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        };

        let fast_jma = JurikMovingAverage::new(&fast_params)
            .map_err(|e| format!("fast JMA: {}", e))?;
        let slow_jma = JurikMovingAverage::new(&slow_params)
            .map_err(|e| format!("slow JMA: {}", e))?;

        Ok(Self {
            primed: false,
            fast_jma,
            slow_jma,
            diff_buffer: Vec::new(),
            diff_buf_size: 3 * params.length,
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

        let fast_val = self.fast_jma.update(sample);
        let slow_val = self.slow_jma.update(sample);

        if fast_val.is_nan() || slow_val.is_nan() {
            return f64::NAN;
        }

        let diff = fast_val - slow_val;

        self.diff_buffer.push(diff);
        if self.diff_buffer.len() > self.diff_buf_size {
            self.diff_buffer.remove(0);
        }

        self.primed = true;

        // Compute MAD.
        let n = self.diff_buffer.len();
        let mut mad = 0.0_f64;
        for &d in &self.diff_buffer {
            mad += d.abs();
        }
        mad /= n as f64;

        if mad < 0.00001 {
            return 0.0;
        }

        diff / (1.5 * mad)
    }
}

impl Indicator for JurikCommodityChannelIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikCommodityChannelIndex,
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
    use crate::indicators::mark_jurik::jurik_commodity_channel_index::testdata::testdata;

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

    fn run_test(length: usize, expected: &[f64]) {
        let params = JurikCommodityChannelIndexParams {
            length,
            ..Default::default()
        };
        let mut ind = JurikCommodityChannelIndex::new(&params).unwrap();
        let input = testdata::test_input();
        for (i, &val) in input.iter().enumerate() {
            let result = ind.update(val);
            assert!(
                almost_equal(result, expected[i]),
                "bar {}: expected {}, got {} (length={})",
                i, expected[i], result, length
            );
        }
    }

    #[test] fn test_len10() { run_test(10, &testdata::expected_len10()); }
    #[test] fn test_len14() { run_test(14, &testdata::expected_len14()); }
    #[test] fn test_len20() { run_test(20, &testdata::expected_len20()); }
    #[test] fn test_len30() { run_test(30, &testdata::expected_len30()); }
    #[test] fn test_len40() { run_test(40, &testdata::expected_len40()); }
    #[test] fn test_len50() { run_test(50, &testdata::expected_len50()); }
    #[test] fn test_len60() { run_test(60, &testdata::expected_len60()); }
    #[test] fn test_len80() { run_test(80, &testdata::expected_len80()); }
    #[test] fn test_len100() { run_test(100, &testdata::expected_len100()); }
}
