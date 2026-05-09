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
use crate::indicators::core::outputs::band::Band;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Bollinger Bands indicator.
pub struct BollingerBandsParams {
    /// The length (number of periods) of the moving window. Must be > 1. Default 5.
    pub length: usize,
    /// Number of standard deviations above the middle band. Default 2.0.
    pub upper_multiplier: f64,
    /// Number of standard deviations below the middle band. Default 2.0.
    pub lower_multiplier: f64,
    /// Whether to use unbiased sample variance (true) or population variance (false). Default true.
    pub is_unbiased: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for BollingerBandsParams {
    fn default() -> Self {
        Self {
            length: 5,
            upper_multiplier: 2.0,
            lower_multiplier: 2.0,
            is_unbiased: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Bollinger Bands indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum BollingerBandsOutput {
    Lower = 1,
    Middle = 2,
    Upper = 3,
    BandWidth = 4,
    PercentBand = 5,
    Band = 6,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// John Bollinger's Bollinger Bands indicator.
pub struct BollingerBands {
    ma: SimpleMovingAverage,
    variance: Variance,
    upper_multiplier: f64,
    lower_multiplier: f64,
    primed: bool,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
}

impl BollingerBands {
    pub fn new(params: &BollingerBandsParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid bollinger bands parameters: length should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let ma = SimpleMovingAverage::new(&SimpleMovingAverageParams {
            length: params.length,
            bar_component: Some(bc),
            quote_component: Some(qc),
            trade_component: Some(tc),
        })?;

        let variance = Variance::new(&VarianceParams {
            length: params.length,
            is_unbiased: params.is_unbiased,
            bar_component: Some(bc),
            quote_component: Some(qc),
            trade_component: Some(tc),
        })?;

        let mnemonic = format!(
            "bb({},{},{}{})",
            params.length,
            params.upper_multiplier as i64,
            params.lower_multiplier as i64,
            component_triple_mnemonic(bc, qc, tc)
        );

        Ok(Self {
            ma,
            variance,
            upper_multiplier: params.upper_multiplier,
            lower_multiplier: params.lower_multiplier,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
        })
    }

    /// Core update. Returns (lower, middle, upper, bandwidth, percentband).
    pub fn update(&mut self, sample: f64) -> (f64, f64, f64, f64, f64) {
        let nan = f64::NAN;

        if sample.is_nan() {
            return (nan, nan, nan, nan, nan);
        }

        let middle = self.ma.update(sample);
        let v = self.variance.update(sample);

        self.primed = self.ma.is_primed() && self.variance.is_primed();

        if middle.is_nan() || v.is_nan() {
            return (nan, nan, nan, nan, nan);
        }

        let stddev = v.sqrt();
        let upper = middle + self.upper_multiplier * stddev;
        let lower = middle - self.lower_multiplier * stddev;

        const EPSILON: f64 = 1e-10;

        let bw = if middle.abs() < EPSILON {
            0.0
        } else {
            (upper - lower) / middle
        };

        let spread = upper - lower;
        let pct_b = if spread.abs() < EPSILON {
            0.0
        } else {
            (sample - lower) / spread
        };

        (lower, middle, upper, bw, pct_b)
    }
}

impl Indicator for BollingerBands {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Bollinger Bands {}", self.mnemonic);
        build_metadata(
            Identifier::BollingerBands,
            &self.mnemonic,
            &desc,
            &[
                OutputText { mnemonic: format!("{} lower", self.mnemonic), description: format!("{} Lower", desc) },
                OutputText { mnemonic: format!("{} middle", self.mnemonic), description: format!("{} Middle", desc) },
                OutputText { mnemonic: format!("{} upper", self.mnemonic), description: format!("{} Upper", desc) },
                OutputText { mnemonic: format!("{} bandWidth", self.mnemonic), description: format!("{} Band Width", desc) },
                OutputText { mnemonic: format!("{} percentBand", self.mnemonic), description: format!("{} Percent Band", desc) },
                OutputText { mnemonic: format!("{} band", self.mnemonic), description: format!("{} Band", desc) },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (lower, middle, upper, bw, pct_b) = self.update(sample.value);
        let t = sample.time;

        let band: Box<dyn std::any::Any> = if lower.is_nan() || upper.is_nan() {
            Box::new(Band::empty(t))
        } else {
            Box::new(Band::new(t, lower, upper))
        };

        vec![
            Box::new(Scalar::new(t, lower)),
            Box::new(Scalar::new(t, middle)),
            Box::new(Scalar::new(t, upper)),
            Box::new(Scalar::new(t, bw)),
            Box::new(Scalar::new(t, pct_b)),
            band,
        ]
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
    fn create_bb(length: usize, is_unbiased: bool) -> BollingerBands {
        BollingerBands::new(&BollingerBandsParams {
            length,
            is_unbiased,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_sample_stddev_length20() {
        let mut bb = create_bb(20, true);
        let input = testdata::closing_prices();
        let exp_mid = testdata::sma20_expected();
        let exp_lo = testdata::sample_lower_expected();
        let exp_hi = testdata::sample_upper_expected();
        let exp_bw = testdata::sample_bandwidth_expected();
        let exp_pb = testdata::sample_pctb_expected();

        for i in 0..252 {
            let (lo, mid, hi, bw, pb) = bb.update(input[i]);

            if exp_mid[i].is_nan() {
                assert!(lo.is_nan(), "[{}] lower expected NaN, got {}", i, lo);
                assert!(mid.is_nan(), "[{}] middle expected NaN, got {}", i, mid);
                assert!(hi.is_nan(), "[{}] upper expected NaN, got {}", i, hi);
                continue;
            }

            assert!((mid - exp_mid[i]).abs() < TOLERANCE, "[{}] middle: expected {}, got {}", i, exp_mid[i], mid);
            assert!((lo - exp_lo[i]).abs() < TOLERANCE, "[{}] lower: expected {}, got {}", i, exp_lo[i], lo);
            assert!((hi - exp_hi[i]).abs() < TOLERANCE, "[{}] upper: expected {}, got {}", i, exp_hi[i], hi);
            assert!((bw - exp_bw[i]).abs() < TOLERANCE, "[{}] bandwidth: expected {}, got {}", i, exp_bw[i], bw);
            assert!((pb - exp_pb[i]).abs() < TOLERANCE, "[{}] percentband: expected {}, got {}", i, exp_pb[i], pb);
        }
    }

    #[test]
    fn test_is_primed() {
        let mut bb = create_bb(20, true);
        let input = testdata::closing_prices();

        assert!(!bb.is_primed());
        for i in 0..19 {
            bb.update(input[i]);
            assert!(!bb.is_primed(), "[{}] should not be primed", i);
        }
        bb.update(input[19]);
        assert!(bb.is_primed());
    }

    #[test]
    fn test_nan_passthrough() {
        let mut bb = create_bb(20, true);
        let (lo, mid, hi, bw, pb) = bb.update(f64::NAN);
        assert!(lo.is_nan());
        assert!(mid.is_nan());
        assert!(hi.is_nan());
        assert!(bw.is_nan());
        assert!(pb.is_nan());
    }

    #[test]
    fn test_metadata() {
        let bb = create_bb(20, true);
        let m = bb.metadata();
        assert_eq!(m.identifier, Identifier::BollingerBands);
        assert_eq!(m.outputs.len(), 6);
        assert_eq!(m.outputs[0].kind, BollingerBandsOutput::Lower as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[5].kind, BollingerBandsOutput::Band as i32);
        assert_eq!(m.outputs[5].shape, Shape::Band);
    }

    #[test]
    fn test_invalid_params() {
        assert!(BollingerBands::new(&BollingerBandsParams { length: 1, ..Default::default() }).is_err());
        assert!(BollingerBands::new(&BollingerBandsParams { length: 0, ..Default::default() }).is_err());
    }

    #[test]
    fn test_update_scalar_output() {
        let mut bb = create_bb(20, true);
        let input = testdata::closing_prices();

        // Feed 19 NaN-producing samples
        for i in 0..19 {
            let out = bb.update_scalar(&Scalar::new(1000 + i as i64, input[i]));
            assert_eq!(out.len(), 6);
            let s = out[0].downcast_ref::<Scalar>().unwrap();
            assert!(s.value.is_nan());
            let band = out[5].downcast_ref::<Band>().unwrap();
            assert!(band.is_empty());
        }

        // Feed index 19 - first primed
        let out = bb.update_scalar(&Scalar::new(1019, input[19]));
        assert_eq!(out.len(), 6);
        let mid = out[1].downcast_ref::<Scalar>().unwrap();
        assert!((mid.value - testdata::sma20_expected()[19]).abs() < TOLERANCE);
        let band = out[5].downcast_ref::<Band>().unwrap();
        assert!(!band.is_empty());
    }
}
