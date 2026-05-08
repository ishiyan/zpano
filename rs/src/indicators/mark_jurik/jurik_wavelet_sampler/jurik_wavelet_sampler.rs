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

/// Parameters to create an instance of the Jurik Wavelet Sampler.
pub struct JurikWaveletSamplerParams {
    /// Index is the number of wavelet columns (1-18). Default 12.
    pub index: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for JurikWaveletSamplerParams {
    fn default() -> Self {
        Self {
            index: 12,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Jurik Wavelet Sampler.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum JurikWaveletSamplerOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// nmTable
// ---------------------------------------------------------------------------

/// (n, M) parameters for columns 0..17.
const NM_TABLE: [(usize, usize); 18] = [
    (1, 0), (2, 0), (3, 0), (4, 0), (5, 0),
    (7, 2), (10, 2), (14, 4), (19, 4), (26, 8),
    (35, 8), (48, 16), (65, 16), (90, 32), (123, 32),
    (172, 64), (237, 64), (334, 128),
];

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Jurik Wavelet Sampler (WAV).
/// Produces `index` output columns per bar, each representing a different
/// multi-resolution scale. The framework output is the first column value.
#[derive(Debug)]
pub struct JurikWaveletSampler {
    primed: bool,
    index: usize,
    max_lookback: usize,
    prices: Vec<f64>,
    bar_count: usize,
    columns: Vec<f64>,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
    description: String,
}

impl JurikWaveletSampler {
    pub fn new(params: &JurikWaveletSamplerParams) -> Result<Self, String> {
        if params.index < 1 || params.index > 18 {
            return Err(
                "invalid jurik wavelet sampler parameters: index must be in range [1, 18]"
                    .to_string(),
            );
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        // Compute max lookback.
        let mut max_lookback = 0usize;
        for c in 0..params.index {
            let lb = NM_TABLE[c].0 + NM_TABLE[c].1 / 2;
            if lb > max_lookback {
                max_lookback = lb;
            }
        }

        let mnemonic = format!(
            "jwav({}{})",
            params.index,
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Jurik wavelet sampler {}", mnemonic);

        Ok(Self {
            primed: false,
            index: params.index,
            max_lookback,
            prices: Vec::new(),
            bar_count: 0,
            columns: vec![f64::NAN; params.index],
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
            description,
        })
    }

    /// Core update. Returns the first column value.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        self.prices.push(sample);
        self.bar_count += 1;

        let mut all_valid = true;

        for c in 0..self.index {
            let n = NM_TABLE[c].0;
            let m = NM_TABLE[c].1;
            let dead_zone = n + m / 2;

            if self.bar_count <= dead_zone {
                self.columns[c] = f64::NAN;
                all_valid = false;
            } else if m == 0 {
                // Simple lag.
                self.columns[c] = self.prices[self.bar_count - 1 - n];
            } else {
                // Mean of (M+1) prices centered at lag n.
                let half = m / 2;
                let center_idx = self.bar_count - 1 - n;
                let mut total = 0.0;

                for k in (center_idx - half)..=(center_idx + half) {
                    total += self.prices[k];
                }

                self.columns[c] = total / (m + 1) as f64;
            }
        }

        if all_valid {
            self.primed = true;
        }

        // Return first column as the framework output.
        self.columns[0]
    }

    /// Returns a copy of the current column values after the last update.
    pub fn columns(&self) -> Vec<f64> {
        self.columns.clone()
    }
}

impl Indicator for JurikWaveletSampler {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikWaveletSampler,
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
    use crate::indicators::core::outputs::shape::Shape;

    const TOLERANCE: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64) -> bool {
        if a.is_nan() && b.is_nan() {
            return true;
        }
        (a - b).abs() <= TOLERANCE
    }

    fn run_wav_test(index: usize, expected_cols: &[&[f64]]) {
        let params = JurikWaveletSamplerParams { index, ..Default::default() };
        let mut ind = JurikWaveletSampler::new(&params).unwrap();
        let input = testdata::test_input();
        for (i, &val) in input.iter().enumerate() {
            ind.update(val);
            let cols = ind.columns();
            for (c, exp_col) in expected_cols.iter().enumerate() {
                assert!(
                    almost_equal(cols[c], exp_col[i]),
                    "bar {} col {}: expected {}, got {}, diff {}",
                    i, c, exp_col[i], cols[c], (cols[c] - exp_col[i]).abs()
                );
            }
        }
    }

    #[test]
    fn test_wav_index12() {
        let exp = vec![
            testdata::expected_wav_col0(), testdata::expected_wav_col1(),
            testdata::expected_wav_col2(), testdata::expected_wav_col3(),
            testdata::expected_wav_col4(), testdata::expected_wav_col5(),
            testdata::expected_wav_col6(), testdata::expected_wav_col7(),
            testdata::expected_wav_col8(), testdata::expected_wav_col9(),
            testdata::expected_wav_col10(), testdata::expected_wav_col11(),
        ];
        let refs: Vec<&[f64]> = exp.iter().map(|v| v.as_slice()).collect();
        run_wav_test(12, &refs);
    }

    #[test]
    fn test_wav_index6() {
        let exp = vec![
            testdata::expected_index6_col0(), testdata::expected_index6_col1(),
            testdata::expected_index6_col2(), testdata::expected_index6_col3(),
            testdata::expected_index6_col4(), testdata::expected_index6_col5(),
        ];
        let refs: Vec<&[f64]> = exp.iter().map(|v| v.as_slice()).collect();
        run_wav_test(6, &refs);
    }

    #[test]
    fn test_wav_index16() {
        let exp = vec![
            testdata::expected_index16_col0(), testdata::expected_index16_col1(),
            testdata::expected_index16_col2(), testdata::expected_index16_col3(),
            testdata::expected_index16_col4(), testdata::expected_index16_col5(),
            testdata::expected_index16_col6(), testdata::expected_index16_col7(),
            testdata::expected_index16_col8(), testdata::expected_index16_col9(),
            testdata::expected_index16_col10(), testdata::expected_index16_col11(),
            testdata::expected_index16_col12(), testdata::expected_index16_col13(),
            testdata::expected_index16_col14(), testdata::expected_index16_col15(),
        ];
        let refs: Vec<&[f64]> = exp.iter().map(|v| v.as_slice()).collect();
        run_wav_test(16, &refs);
    }

    #[test]
    fn test_wav_metadata() {
        let params = JurikWaveletSamplerParams::default();
        let ind = JurikWaveletSampler::new(&params).unwrap();
        let md = ind.metadata();
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_wav_invalid_params() {
        let params = JurikWaveletSamplerParams { index: 0, ..Default::default() };
        assert!(JurikWaveletSampler::new(&params).is_err());
        let params = JurikWaveletSamplerParams { index: 19, ..Default::default() };
        assert!(JurikWaveletSampler::new(&params).is_err());
    }
}
