use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::common::simple_moving_average::{SimpleMovingAverage, SimpleMovingAverageParams};
use crate::indicators::common::variance::{Variance, VarianceParams};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::{BarFunc, QuoteFunc, TradeFunc};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Bollinger Bands Trend indicator.
pub struct BollingerBandsTrendParams {
    pub fast_length: usize,
    pub slow_length: usize,
    pub upper_multiplier: f64,
    pub lower_multiplier: f64,
    /// None defaults to true (sample/unbiased).
    pub is_unbiased: Option<bool>,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for BollingerBandsTrendParams {
    fn default() -> Self {
        Self {
            fast_length: 20,
            slow_length: 50,
            upper_multiplier: 2.0,
            lower_multiplier: 2.0,
            is_unbiased: None,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Bollinger Bands Trend indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum BollingerBandsTrendOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Internal bbLine
// ---------------------------------------------------------------------------

struct BbLine {
    ma: SimpleMovingAverage,
    variance: Variance,
    upper_multiplier: f64,
    lower_multiplier: f64,
}

impl BbLine {
    fn new(length: usize, upper_multiplier: f64, lower_multiplier: f64, is_unbiased: bool,
           bc: BarComponent, qc: QuoteComponent, tc: TradeComponent) -> Result<Self, String> {
        let ma = SimpleMovingAverage::new(&SimpleMovingAverageParams {
            length,
            bar_component: Some(bc),
            quote_component: Some(qc),
            trade_component: Some(tc),
        })?;

        let variance = Variance::new(&VarianceParams {
            length,
            is_unbiased,
            bar_component: Some(bc),
            quote_component: Some(qc),
            trade_component: Some(tc),
        })?;

        Ok(Self { ma, variance, upper_multiplier, lower_multiplier })
    }

    /// Returns (lower, middle, upper, primed).
    fn update(&mut self, sample: f64) -> (f64, f64, f64, bool) {
        let middle = self.ma.update(sample);
        let v = self.variance.update(sample);
        let primed = self.ma.is_primed() && self.variance.is_primed();

        if middle.is_nan() || v.is_nan() {
            return (f64::NAN, f64::NAN, f64::NAN, primed);
        }

        let stddev = v.sqrt();
        let upper = middle + self.upper_multiplier * stddev;
        let lower = middle - self.lower_multiplier * stddev;

        (lower, middle, upper, primed)
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// John Bollinger's Bollinger Bands Trend indicator.
pub struct BollingerBandsTrend {
    fast_bb: BbLine,
    slow_bb: BbLine,
    value: f64,
    primed: bool,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
}

impl BollingerBandsTrend {
    pub fn new(params: &BollingerBandsTrendParams) -> Result<Self, String> {
        let fast_length = if params.fast_length == 0 { 20 } else { params.fast_length };
        let slow_length = if params.slow_length == 0 { 50 } else { params.slow_length };
        let upper_multiplier = if params.upper_multiplier == 0.0 { 2.0 } else { params.upper_multiplier };
        let lower_multiplier = if params.lower_multiplier == 0.0 { 2.0 } else { params.lower_multiplier };
        let is_unbiased = params.is_unbiased.unwrap_or(true);

        const INVALID: &str = "invalid bollinger bands trend parameters";

        if fast_length < 2 {
            return Err(format!("{}: fast length should be greater than 1", INVALID));
        }
        if slow_length < 2 {
            return Err(format!("{}: slow length should be greater than 1", INVALID));
        }
        if slow_length <= fast_length {
            return Err(format!("{}: slow length should be greater than fast length", INVALID));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let fast_bb = BbLine::new(fast_length, upper_multiplier, lower_multiplier, is_unbiased, bc, qc, tc)?;
        let slow_bb = BbLine::new(slow_length, upper_multiplier, lower_multiplier, is_unbiased, bc, qc, tc)?;

        let mnemonic = format!(
            "bbtrend({},{},{},{}{})",
            fast_length, slow_length,
            upper_multiplier as i64, lower_multiplier as i64,
            component_triple_mnemonic(bc, qc, tc)
        );

        Ok(Self {
            fast_bb,
            slow_bb,
            value: f64::NAN,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
        })
    }

    /// Core update. Returns the BBTrend value.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return f64::NAN;
        }

        let (fast_lower, fast_middle, fast_upper, fast_primed) = self.fast_bb.update(sample);
        let (slow_lower, _, slow_upper, slow_primed) = self.slow_bb.update(sample);

        self.primed = fast_primed && slow_primed;

        if !self.primed || fast_middle.is_nan() || fast_lower.is_nan() || slow_lower.is_nan() {
            self.value = f64::NAN;
            return f64::NAN;
        }

        const EPSILON: f64 = 1e-10;

        let lower_diff = (fast_lower - slow_lower).abs();
        let upper_diff = (fast_upper - slow_upper).abs();

        if fast_middle.abs() < EPSILON {
            self.value = 0.0;
            return 0.0;
        }

        let result = (lower_diff - upper_diff) / fast_middle;
        self.value = result;
        result
    }
}

impl Indicator for BollingerBandsTrend {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Bollinger Bands Trend {}", self.mnemonic);
        build_metadata(
            Identifier::BollingerBandsTrend,
            &self.mnemonic,
            &desc,
            &[OutputText { mnemonic: self.mnemonic.clone(), description: desc.clone() }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, v))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        self.update_scalar(&Scalar::new(sample.time, v))
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        self.update_scalar(&Scalar::new(sample.time, v))
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        self.update_scalar(&Scalar::new(sample.time, v))
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
    #[test]
    fn test_sample_stddev_full_data() {
        let input = testdata::closing_prices();
        let expected = testdata::sample_expected();

        let mut ind = BollingerBandsTrend::new(&BollingerBandsTrendParams {
            is_unbiased: Some(true),
            ..Default::default()
        }).unwrap();

        for i in 0..252 {
            let v = ind.update(input[i]);
            if expected[i].is_nan() {
                assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
            } else {
                assert!((v - expected[i]).abs() < TOLERANCE, "[{}] expected {}, got {}", i, expected[i], v);
            }
        }
    }

    #[test]
    fn test_population_stddev_full_data() {
        let input = testdata::closing_prices();
        let expected = testdata::population_expected();

        let mut ind = BollingerBandsTrend::new(&BollingerBandsTrendParams {
            is_unbiased: Some(false),
            ..Default::default()
        }).unwrap();

        for i in 0..252 {
            let v = ind.update(input[i]);
            if expected[i].is_nan() {
                assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
            } else {
                assert!((v - expected[i]).abs() < TOLERANCE, "[{}] expected {}, got {}", i, expected[i], v);
            }
        }
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::closing_prices();
        let mut ind = BollingerBandsTrend::new(&Default::default()).unwrap();

        assert!(!ind.is_primed());
        for i in 0..49 {
            ind.update(input[i]);
            assert!(!ind.is_primed(), "[{}] should not be primed", i);
        }
        ind.update(input[49]);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_nan_passthrough() {
        let mut ind = BollingerBandsTrend::new(&Default::default()).unwrap();
        let v = ind.update(f64::NAN);
        assert!(v.is_nan());
    }

    #[test]
    fn test_metadata() {
        let ind = BollingerBandsTrend::new(&Default::default()).unwrap();
        let m = ind.metadata();
        assert_eq!(m.identifier, Identifier::BollingerBandsTrend);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, BollingerBandsTrendOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_update_scalar_output() {
        let input = testdata::closing_prices();
        let expected = testdata::sample_expected();

        let mut ind = BollingerBandsTrend::new(&BollingerBandsTrendParams {
            is_unbiased: Some(true),
            ..Default::default()
        }).unwrap();

        for i in 0..49 {
            let out = ind.update_scalar(&Scalar::new(1000 + i as i64, input[i]));
            assert_eq!(out.len(), 1);
            let s = out[0].downcast_ref::<Scalar>().unwrap();
            assert!(s.value.is_nan());
        }

        let out = ind.update_scalar(&Scalar::new(1049, input[49]));
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - expected[49]).abs() < TOLERANCE);
    }

    #[test]
    fn test_invalid_params() {
        assert!(BollingerBandsTrend::new(&BollingerBandsTrendParams {
            fast_length: 1, slow_length: 50, ..Default::default()
        }).is_err());
        assert!(BollingerBandsTrend::new(&BollingerBandsTrendParams {
            fast_length: 20, slow_length: 1, ..Default::default()
        }).is_err());
        assert!(BollingerBandsTrend::new(&BollingerBandsTrendParams {
            fast_length: 20, slow_length: 20, ..Default::default()
        }).is_err());
        assert!(BollingerBandsTrend::new(&BollingerBandsTrendParams {
            fast_length: 50, slow_length: 20, ..Default::default()
        }).is_err());
    }
}
