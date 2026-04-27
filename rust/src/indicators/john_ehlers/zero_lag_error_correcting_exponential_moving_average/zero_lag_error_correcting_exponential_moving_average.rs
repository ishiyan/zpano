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
use crate::indicators::core::line_indicator::{BarFunc, QuoteFunc, TradeFunc};
use crate::indicators::core::metadata::Metadata;

/// Formats a float with 4 significant digits, matching Go's `%.4g`.
fn format_4g(v: f64) -> String {
    let s = format!("{:.3e}", v);
    let parts: Vec<&str> = s.split('e').collect();
    let exp: i32 = parts[1].parse().unwrap();
    if exp >= 0 && exp < 4 {
        let decimals = if 3 - exp > 0 { (3 - exp) as usize } else { 0 };
        let fixed = format!("{:.*}", decimals, v);
        fixed.trim_end_matches('0').trim_end_matches('.').to_string()
    } else if exp < 0 && exp >= -3 {
        let decimals = (3 - exp) as usize;
        let fixed = format!("{:.*}", decimals, v);
        fixed.trim_end_matches('0').trim_end_matches('.').to_string()
    } else {
        format!("{:.3e}", v)
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the zero-lag error-correcting exponential moving average.
pub struct ZeroLagErrorCorrectingExponentialMovingAverageParams {
    /// Smoothing factor (alpha) of the EMA. Must be in (0, 1].
    /// Default: 0.095 (equivalent to length 20).
    pub smoothing_factor: f64,
    /// Gain limit defines the range [-g, g] for finding the best gain factor.
    /// Must be positive. Default: 5.
    pub gain_limit: f64,
    /// Gain step defines the iteration step for finding the best gain factor.
    /// Must be positive. Default: 0.1.
    pub gain_step: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for ZeroLagErrorCorrectingExponentialMovingAverageParams {
    fn default() -> Self {
        Self {
            smoothing_factor: 0.095,
            gain_limit: 5.0,
            gain_step: 0.1,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum ZeroLagErrorCorrectingExponentialMovingAverageOutput {
    /// The scalar value of the zero-lag error-correcting exponential moving average.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Zero-lag Error-Correcting Exponential Moving Average (ZECEMA).
///
/// The algorithm iterates over gain values in [-gainLimit, gainLimit] with the given
/// gainStep to find the gain that minimizes the error between the sample and the
/// error-corrected EMA value.
///
/// The indicator is not primed during the first two updates; it primes on the third.
pub struct ZeroLagErrorCorrectingExponentialMovingAverage {
    mnemonic: String,
    description: String,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    alpha: f64,
    one_min_alpha: f64,
    gain_limit: f64,
    gain_step: f64,
    count: usize,
    value: f64,
    ema_value: f64,
    primed: bool,
}

impl ZeroLagErrorCorrectingExponentialMovingAverage {
    /// Creates a new instance from the given parameters.
    pub fn new(p: &ZeroLagErrorCorrectingExponentialMovingAverageParams) -> Result<Self, String> {
        let invalid = "invalid zero-lag error-correcting exponential moving average parameters";

        let sf = p.smoothing_factor;
        if sf <= 0.0 || sf > 1.0 {
            return Err(format!("{}: smoothing factor should be in (0, 1]", invalid));
        }

        let gl = p.gain_limit;
        if gl <= 0.0 {
            return Err(format!("{}: gain limit should be positive", invalid));
        }

        let gs = p.gain_step;
        if gs <= 0.0 {
            return Err(format!("{}: gain step should be positive", invalid));
        }

        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let component_mnemonic = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!(
            "zecema({}, {}, {}{})",
            format_4g(sf),
            format_4g(gl),
            format_4g(gs),
            component_mnemonic
        );
        let description = format!(
            "Zero-lag Error-Correcting Exponential Moving Average {}",
            mnemonic
        );

        Ok(Self {
            mnemonic,
            description,
            bar_func,
            quote_func,
            trade_func,
            alpha: sf,
            one_min_alpha: 1.0 - sf,
            gain_limit: gl,
            gain_step: gs,
            count: 0,
            value: f64::NAN,
            ema_value: f64::NAN,
            primed: false,
        })
    }

    /// Core update logic. Returns the ZECEMA value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        if self.primed {
            self.value = self.calculate(sample);
            return self.value;
        }

        self.count += 1;

        if self.count == 1 {
            self.ema_value = sample;
            return f64::NAN;
        }

        if self.count == 2 {
            self.ema_value = self.calculate_ema(sample);
            self.value = self.ema_value;
            return f64::NAN;
        }

        // count == 3: prime the indicator.
        self.value = self.calculate(sample);
        self.primed = true;
        self.value
    }

    fn calculate_ema(&self, sample: f64) -> f64 {
        self.alpha * sample + self.one_min_alpha * self.ema_value
    }

    fn calculate(&mut self, sample: f64) -> f64 {
        self.ema_value = self.calculate_ema(sample);

        let mut least_error = f64::MAX;
        let mut best_ec = 0.0;

        let mut gain = -self.gain_limit;
        while gain <= self.gain_limit {
            let ec = self.alpha * (self.ema_value + gain * (sample - self.value))
                + self.one_min_alpha * self.value;
            let err = (sample - ec).abs();

            if least_error > err {
                least_error = err;
                best_ec = ec;
            }

            gain += self.gain_step;
        }

        best_ec
    }
}

impl Indicator for ZeroLagErrorCorrectingExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::ZeroLagErrorCorrectingExponentialMovingAverage,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        let value = self.update(v);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        let value = self.update(v);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        let value = self.update(v);
        vec![Box::new(Scalar::new(sample.time, value))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    fn create_default() -> ZeroLagErrorCorrectingExponentialMovingAverage {
        ZeroLagErrorCorrectingExponentialMovingAverage::new(
            &ZeroLagErrorCorrectingExponentialMovingAverageParams::default(),
        )
        .unwrap()
    }

    #[test]
    fn test_is_primed() {
        let mut z = create_default();
        assert!(!z.is_primed());

        // First 2 updates should not prime.
        for _ in 0..2 {
            z.update(100.0);
            assert!(!z.is_primed());
        }

        // 3rd update should prime.
        z.update(100.0);
        assert!(z.is_primed());
    }

    #[test]
    fn test_update_nan() {
        let mut z = create_default();
        assert!(z.update(f64::NAN).is_nan());
        assert!(!z.is_primed());
    }

    #[test]
    fn test_update_constant() {
        let value = 42.0;
        let mut z = create_default();

        // First 2 should return NaN.
        for _ in 0..2 {
            assert!(z.update(value).is_nan());
        }

        // 3rd should return value (constant input).
        let act = z.update(value);
        assert!((act - value).abs() < 1e-6);

        // Further updates should stay close to value.
        for _ in 0..10 {
            let act = z.update(value);
            assert!((act - value).abs() < 1e-6);
        }
    }

    #[test]
    fn test_metadata() {
        let z = create_default();
        let md = z.metadata();
        assert_eq!(md.identifier, Identifier::ZeroLagErrorCorrectingExponentialMovingAverage);
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].mnemonic, "zecema(0.095, 5, 0.1)");
        assert_eq!(
            md.outputs[0].description,
            "Zero-lag Error-Correcting Exponential Moving Average zecema(0.095, 5, 0.1)"
        );
    }

    #[test]
    fn test_new_invalid_smoothing_factor() {
        let p = ZeroLagErrorCorrectingExponentialMovingAverageParams {
            smoothing_factor: 0.0,
            ..Default::default()
        };
        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::new(&p).is_err());

        let p = ZeroLagErrorCorrectingExponentialMovingAverageParams {
            smoothing_factor: -0.1,
            ..Default::default()
        };
        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::new(&p).is_err());

        let p = ZeroLagErrorCorrectingExponentialMovingAverageParams {
            smoothing_factor: 1.1,
            ..Default::default()
        };
        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::new(&p).is_err());

        // sf=1 should be valid.
        let p = ZeroLagErrorCorrectingExponentialMovingAverageParams {
            smoothing_factor: 1.0,
            ..Default::default()
        };
        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::new(&p).is_ok());
    }

    #[test]
    fn test_new_invalid_gain_limit() {
        let p = ZeroLagErrorCorrectingExponentialMovingAverageParams {
            gain_limit: 0.0,
            ..Default::default()
        };
        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::new(&p).is_err());

        let p = ZeroLagErrorCorrectingExponentialMovingAverageParams {
            gain_limit: -1.0,
            ..Default::default()
        };
        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::new(&p).is_err());
    }

    #[test]
    fn test_new_invalid_gain_step() {
        let p = ZeroLagErrorCorrectingExponentialMovingAverageParams {
            gain_step: 0.0,
            ..Default::default()
        };
        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::new(&p).is_err());

        let p = ZeroLagErrorCorrectingExponentialMovingAverageParams {
            gain_step: -0.1,
            ..Default::default()
        };
        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::new(&p).is_err());
    }

    #[test]
    fn test_non_default_bar_component() {
        let p = ZeroLagErrorCorrectingExponentialMovingAverageParams {
            bar_component: Some(BarComponent::Open),
            ..Default::default()
        };
        let z = ZeroLagErrorCorrectingExponentialMovingAverage::new(&p).unwrap();
        let md = z.metadata();
        assert_eq!(md.outputs[0].mnemonic, "zecema(0.095, 5, 0.1, o)");
    }

    #[test]
    fn test_update_entity() {
        let mut z = create_default();
        let inp = 100.0;
        let tm = 1617235200_i64;

        // Prime (3 updates).
        z.update(inp);
        z.update(inp);
        z.update(inp);
        assert!(z.is_primed());

        let s = Scalar::new(tm, inp);
        let out = z.update_scalar(&s);
        assert_eq!(out.len(), 1);
        let scalar = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(scalar.time, tm);
        assert!(!scalar.value.is_nan());
    }
}
