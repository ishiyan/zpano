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

/// Parameters to create an instance of the zero-lag exponential moving average.
pub struct ZeroLagExponentialMovingAverageParams {
    /// Smoothing factor (alpha) of the EMA. Must be in (0, 1].
    /// Default: 0.25.
    pub smoothing_factor: f64,
    /// Gain factor used to estimate the velocity. Default: 0.5.
    pub velocity_gain_factor: f64,
    /// Length of the momentum used to estimate velocity. Must be >= 1.
    /// Default: 3.
    pub velocity_momentum_length: i32,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for ZeroLagExponentialMovingAverageParams {
    fn default() -> Self {
        Self {
            smoothing_factor: 0.25,
            velocity_gain_factor: 0.5,
            velocity_momentum_length: 3,
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
pub enum ZeroLagExponentialMovingAverageOutput {
    /// The scalar value of the zero-lag exponential moving average.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Zero-lag Exponential Moving Average (ZEMA).
///
/// ZEMA = alpha*(Price + gainFactor*(Price - Price[momentumLength ago])) + (1-alpha)*ZEMA_prev
pub struct ZeroLagExponentialMovingAverage {
    mnemonic: String,
    description: String,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    alpha: f64,
    one_min_alpha: f64,
    gain_factor: f64,
    momentum_length: usize,
    momentum_window: Vec<f64>,
    count: usize,
    value: f64,
    primed: bool,
}

impl ZeroLagExponentialMovingAverage {
    /// Creates a new instance from the given parameters.
    pub fn new(p: &ZeroLagExponentialMovingAverageParams) -> Result<Self, String> {
        let invalid = "invalid zero-lag exponential moving average parameters";

        let sf = p.smoothing_factor;
        if sf <= 0.0 || sf > 1.0 {
            return Err(format!("{}: smoothing factor should be in (0, 1]", invalid));
        }

        let ml = p.velocity_momentum_length;
        if ml < 1 {
            return Err(format!("{}: velocity momentum length should be positive", invalid));
        }

        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let component_mnemonic = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!("zema({}, {}, {}{})", format_4g(sf), format_4g(p.velocity_gain_factor), ml, component_mnemonic);
        let description = format!("Zero-lag Exponential Moving Average {}", mnemonic);

        let ml_usize = ml as usize;

        Ok(Self {
            mnemonic,
            description,
            bar_func,
            quote_func,
            trade_func,
            alpha: sf,
            one_min_alpha: 1.0 - sf,
            gain_factor: p.velocity_gain_factor,
            momentum_length: ml_usize,
            momentum_window: vec![0.0; ml_usize + 1],
            count: 0,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Core update logic. Returns the ZEMA value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        if self.primed {
            // Shift momentum window left by 1.
            for i in 0..self.momentum_length {
                self.momentum_window[i] = self.momentum_window[i + 1];
            }
            self.momentum_window[self.momentum_length] = sample;
            self.value = self.calculate(sample);
            return self.value;
        }

        self.momentum_window[self.count] = sample;
        self.count += 1;

        if self.count <= self.momentum_length {
            self.value = sample;
            return f64::NAN;
        }

        // count == momentum_length + 1: prime the indicator.
        self.value = self.calculate(sample);
        self.primed = true;
        self.value
    }

    fn calculate(&self, sample: f64) -> f64 {
        let momentum = sample - self.momentum_window[0];
        self.alpha * (sample + self.gain_factor * momentum) + self.one_min_alpha * self.value
    }
}

impl Indicator for ZeroLagExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::ZeroLagExponentialMovingAverage,
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

    fn create_default() -> ZeroLagExponentialMovingAverage {
        ZeroLagExponentialMovingAverage::new(&ZeroLagExponentialMovingAverageParams::default()).unwrap()
    }

    #[test]
    fn test_is_primed() {
        let mut z = create_default();
        assert!(!z.is_primed());

        // First 3 updates (momentum_length=3) should not prime.
        for _ in 0..3 {
            z.update(100.0);
            assert!(!z.is_primed());
        }

        // 4th update should prime.
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

        // First 3 should return NaN.
        for _ in 0..3 {
            assert!(z.update(value).is_nan());
        }

        // 4th should return value (momentum=0 for constant input).
        let act = z.update(value);
        assert!((act - value).abs() < 1e-10);

        // Further updates should stay at value.
        for _ in 0..10 {
            let act = z.update(value);
            assert!((act - value).abs() < 1e-10);
        }
    }

    #[test]
    fn test_metadata() {
        let z = create_default();
        let md = z.metadata();
        assert_eq!(md.identifier, Identifier::ZeroLagExponentialMovingAverage);
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].mnemonic, "zema(0.25, 0.5, 3)");
        assert_eq!(md.outputs[0].description, "Zero-lag Exponential Moving Average zema(0.25, 0.5, 3)");
    }

    #[test]
    fn test_new_invalid_smoothing_factor() {
        let p = ZeroLagExponentialMovingAverageParams { smoothing_factor: 0.0, ..Default::default() };
        assert!(ZeroLagExponentialMovingAverage::new(&p).is_err());

        let p = ZeroLagExponentialMovingAverageParams { smoothing_factor: -0.1, ..Default::default() };
        assert!(ZeroLagExponentialMovingAverage::new(&p).is_err());

        let p = ZeroLagExponentialMovingAverageParams { smoothing_factor: 1.1, ..Default::default() };
        assert!(ZeroLagExponentialMovingAverage::new(&p).is_err());

        // sf=1 should be valid.
        let p = ZeroLagExponentialMovingAverageParams { smoothing_factor: 1.0, ..Default::default() };
        assert!(ZeroLagExponentialMovingAverage::new(&p).is_ok());
    }

    #[test]
    fn test_new_invalid_momentum_length() {
        let p = ZeroLagExponentialMovingAverageParams { velocity_momentum_length: 0, ..Default::default() };
        assert!(ZeroLagExponentialMovingAverage::new(&p).is_err());

        let p = ZeroLagExponentialMovingAverageParams { velocity_momentum_length: -1, ..Default::default() };
        assert!(ZeroLagExponentialMovingAverage::new(&p).is_err());
    }

    #[test]
    fn test_non_default_bar_component() {
        let p = ZeroLagExponentialMovingAverageParams {
            bar_component: Some(BarComponent::Open),
            ..Default::default()
        };
        let z = ZeroLagExponentialMovingAverage::new(&p).unwrap();
        let md = z.metadata();
        assert_eq!(md.outputs[0].mnemonic, "zema(0.25, 0.5, 3, o)");
    }

    #[test]
    fn test_update_entity() {
        let mut z = create_default();
        let inp = 100.0;
        let tm = 1617235200_i64;

        // Prime.
        z.update(inp);
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
