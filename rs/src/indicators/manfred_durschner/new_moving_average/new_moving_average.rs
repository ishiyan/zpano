use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// MAType
// ---------------------------------------------------------------------------

/// Enumerates the moving average types used in NMA.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MAType {
    SMA = 0,
    EMA = 1,
    SMMA = 2,
    LWMA = 3,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the New Moving Average indicator.
pub struct NewMovingAverageParams {
    pub primary_period: usize,
    pub secondary_period: usize,
    pub ma_type: MAType,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for NewMovingAverageParams {
    fn default() -> Self {
        Self {
            primary_period: 0,
            secondary_period: 8,
            ma_type: MAType::LWMA,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output enum
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the New Moving Average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum NewMovingAverageOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Streaming MA trait and implementations
// ---------------------------------------------------------------------------

trait StreamingMA {
    fn update(&mut self, sample: f64) -> f64;
}

struct StreamingSMA {
    period: usize,
    buffer: Vec<f64>,
    buffer_index: usize,
    buffer_count: usize,
    sum: f64,
    primed: bool,
}

impl StreamingSMA {
    fn new(period: usize) -> Self {
        Self {
            period,
            buffer: vec![0.0; period],
            buffer_index: 0,
            buffer_count: 0,
            sum: 0.0,
            primed: false,
        }
    }
}

impl StreamingMA for StreamingSMA {
    fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }
        if self.primed {
            self.sum -= self.buffer[self.buffer_index];
        }
        self.buffer[self.buffer_index] = sample;
        self.sum += sample;
        self.buffer_index = (self.buffer_index + 1) % self.period;
        if !self.primed {
            self.buffer_count += 1;
            if self.buffer_count < self.period {
                return f64::NAN;
            }
            self.primed = true;
        }
        self.sum / self.period as f64
    }
}

struct StreamingEMA {
    period: usize,
    multiplier: f64,
    count: usize,
    sum: f64,
    value: f64,
    primed: bool,
}

impl StreamingEMA {
    fn new(period: usize) -> Self {
        Self {
            period,
            multiplier: 2.0 / (period as f64 + 1.0),
            count: 0,
            sum: 0.0,
            value: f64::NAN,
            primed: false,
        }
    }
}

impl StreamingMA for StreamingEMA {
    fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }
        if !self.primed {
            self.count += 1;
            self.sum += sample;
            if self.count < self.period {
                return f64::NAN;
            }
            self.value = self.sum / self.period as f64;
            self.primed = true;
            return self.value;
        }
        self.value = (sample - self.value) * self.multiplier + self.value;
        self.value
    }
}

struct StreamingSMMA {
    period: usize,
    count: usize,
    sum: f64,
    value: f64,
    primed: bool,
}

impl StreamingSMMA {
    fn new(period: usize) -> Self {
        Self {
            period,
            count: 0,
            sum: 0.0,
            value: f64::NAN,
            primed: false,
        }
    }
}

impl StreamingMA for StreamingSMMA {
    fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }
        if !self.primed {
            self.count += 1;
            self.sum += sample;
            if self.count < self.period {
                return f64::NAN;
            }
            self.value = self.sum / self.period as f64;
            self.primed = true;
            return self.value;
        }
        self.value = (self.value * (self.period - 1) as f64 + sample) / self.period as f64;
        self.value
    }
}

struct StreamingLWMA {
    period: usize,
    buffer: Vec<f64>,
    buffer_index: usize,
    buffer_count: usize,
    weight_sum: f64,
    primed: bool,
}

impl StreamingLWMA {
    fn new(period: usize) -> Self {
        Self {
            period,
            buffer: vec![0.0; period],
            buffer_index: 0,
            buffer_count: 0,
            weight_sum: period as f64 * (period + 1) as f64 / 2.0,
            primed: false,
        }
    }
}

impl StreamingMA for StreamingLWMA {
    fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }
        self.buffer[self.buffer_index] = sample;
        self.buffer_index = (self.buffer_index + 1) % self.period;
        if !self.primed {
            self.buffer_count += 1;
            if self.buffer_count < self.period {
                return f64::NAN;
            }
            self.primed = true;
        }
        let mut result = 0.0;
        let mut index = self.buffer_index;
        for i in 0..self.period {
            result += (i + 1) as f64 * self.buffer[index];
            index = (index + 1) % self.period;
        }
        result / self.weight_sum
    }
}

fn create_streaming_ma(ma_type: MAType, period: usize) -> Box<dyn StreamingMA> {
    match ma_type {
        MAType::SMA => Box::new(StreamingSMA::new(period)),
        MAType::EMA => Box::new(StreamingEMA::new(period)),
        MAType::SMMA => Box::new(StreamingSMMA::new(period)),
        MAType::LWMA => Box::new(StreamingLWMA::new(period)),
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the New Moving Average (NMA) by Manfred Dürschner.
///
/// NMA applies the Nyquist-Shannon sampling theorem to moving average design:
/// by cascading two moving averages whose period ratio satisfies the Nyquist
/// criterion (lambda = n1/n2 >= 2), the resulting lag can be extrapolated away
/// geometrically.
///
/// Formula: NMA = (1 + alpha) * MA1 - alpha * MA2
/// where: alpha = lambda * (n1-1) / (n1-lambda), lambda = n1 / n2 (integer division)
pub struct NewMovingAverage {
    line: LineIndicator,
    alpha: f64,
    ma_primary: Box<dyn StreamingMA>,
    ma_secondary: Box<dyn StreamingMA>,
    primed: bool,
}

impl NewMovingAverage {
    /// Creates a new NewMovingAverage from the given parameters.
    pub fn new(params: &NewMovingAverageParams) -> Result<Self, String> {
        let mut primary_period = params.primary_period;
        let secondary_period_input = params.secondary_period;

        // Enforce Nyquist constraint.
        if primary_period < 4 {
            primary_period = 4;
        }
        let mut secondary_period = secondary_period_input;
        if secondary_period < 2 {
            secondary_period = 2;
        }
        if primary_period < secondary_period * 2 {
            primary_period = secondary_period * 4;
        }

        // Compute alpha.
        let nyquist_ratio = primary_period / secondary_period;
        let alpha = nyquist_ratio as f64 * (primary_period - 1) as f64
            / (primary_period - nyquist_ratio) as f64;

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "nma({}, {}, {}{})",
            primary_period,
            secondary_period,
            params.ma_type as u8,
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("New moving average {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            alpha,
            ma_primary: create_streaming_ma(params.ma_type, primary_period),
            ma_secondary: create_streaming_ma(params.ma_type, secondary_period),
            primed: false,
        })
    }

    /// Core update logic.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let ma1_value = self.ma_primary.update(sample);
        if ma1_value.is_nan() {
            return f64::NAN;
        }

        let ma2_value = self.ma_secondary.update(ma1_value);
        if ma2_value.is_nan() {
            return f64::NAN;
        }

        self.primed = true;
        (1.0 + self.alpha) * ma1_value - self.alpha * ma2_value
    }
}

impl Indicator for NewMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::NewMovingAverage,
            &self.line.mnemonic,
            &self.line.description,
            &[OutputText {
                mnemonic: self.line.mnemonic.clone(),
                description: self.line.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let sample_value = (self.line.bar_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let sample_value = (self.line.quote_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let sample_value = (self.line.trade_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;

    fn create_nma(primary_period: usize, secondary_period: usize, ma_type: MAType) -> NewMovingAverage {
        NewMovingAverage::new(&NewMovingAverageParams {
            primary_period,
            secondary_period,
            ma_type,
            ..Default::default()
        })
        .unwrap()
    }

    fn check_nma(nma: &mut NewMovingAverage, input: &[f64], expected: &[f64]) {
        for i in 0..input.len() {
            let act = nma.update(input[i]);
            if expected[i].is_nan() {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(
                    (act - expected[i]).abs() < 1e-13,
                    "[{}] expected {}, got {}",
                    i,
                    expected[i],
                    act
                );
            }
        }
        // NaN passthrough.
        assert!(nma.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_sec4_pri_auto_lwma() {
        let input = testdata::test_input();
        let exp = testdata::expected_sec4_pri_auto_lwma();
        let mut nma = create_nma(0, 4, MAType::LWMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_sec8_pri_auto_lwma() {
        let input = testdata::test_input();
        let exp = testdata::expected_sec8_pri_auto_lwma();
        let mut nma = create_nma(0, 8, MAType::LWMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_sec16_pri_auto_lwma() {
        let input = testdata::test_input();
        let exp = testdata::expected_sec16_pri_auto_lwma();
        let mut nma = create_nma(0, 16, MAType::LWMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_pri16_sec8_lwma() {
        let input = testdata::test_input();
        let exp = testdata::expected_pri16_sec8_lwma();
        let mut nma = create_nma(16, 8, MAType::LWMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_pri32_sec8_lwma() {
        let input = testdata::test_input();
        let exp = testdata::expected_pri32_sec8_lwma();
        let mut nma = create_nma(32, 8, MAType::LWMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_pri64_sec8_lwma() {
        let input = testdata::test_input();
        let exp = testdata::expected_pri64_sec8_lwma();
        let mut nma = create_nma(64, 8, MAType::LWMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_pri8_sec4_lwma() {
        let input = testdata::test_input();
        let exp = testdata::expected_pri8_sec4_lwma();
        let mut nma = create_nma(8, 4, MAType::LWMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_pri16_sec4_lwma() {
        let input = testdata::test_input();
        let exp = testdata::expected_pri16_sec4_lwma();
        let mut nma = create_nma(16, 4, MAType::LWMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_pri32_sec4_lwma() {
        let input = testdata::test_input();
        let exp = testdata::expected_pri32_sec4_lwma();
        let mut nma = create_nma(32, 4, MAType::LWMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_sec8_sma() {
        let input = testdata::test_input();
        let exp = testdata::expected_sec8_sma();
        let mut nma = create_nma(0, 8, MAType::SMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_sec8_ema() {
        let input = testdata::test_input();
        let exp = testdata::expected_sec8_ema();
        let mut nma = create_nma(0, 8, MAType::EMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_sec8_smma() {
        let input = testdata::test_input();
        let exp = testdata::expected_sec8_smma();
        let mut nma = create_nma(0, 8, MAType::SMMA);
        check_nma(&mut nma, &input, &exp);
    }

    #[test]
    fn test_metadata() {
        let nma = create_nma(0, 8, MAType::LWMA);
        let m = nma.metadata();
        assert_eq!(m.identifier, Identifier::NewMovingAverage);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].mnemonic, "nma(32, 8, 3)");
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut nma = create_nma(0, 8, MAType::LWMA);
        assert!(!nma.is_primed());
        // Feed until primed (primary=32 + secondary=8 - 1 = 38 warmup)
        for i in 0..38 {
            nma.update(input[i]);
            assert!(!nma.is_primed(), "[{}] should not be primed", i);
        }
        nma.update(input[38]);
        assert!(nma.is_primed());
    }
}
