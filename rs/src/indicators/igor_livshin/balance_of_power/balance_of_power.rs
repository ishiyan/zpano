use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the balance of power indicator.
/// Balance of Power requires OHLC bar data and has no configurable parameters.
pub struct BalanceOfPowerParams;

impl Default for BalanceOfPowerParams {
    fn default() -> Self {
        Self
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the balance of power indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum BalanceOfPowerOutput {
    /// The scalar value of the balance of power.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const EPSILON: f64 = 1e-8;

/// Igor Livshin's Balance of Power (BOP).
///
/// The Balance of Market Power captures the struggles of bulls vs. bears
/// throughout the trading day.
///
/// BOP = (Close - Open) / (High - Low)
///
/// When the range (High - Low) is less than epsilon, the value is 0.
pub struct BalanceOfPower {
    line: LineIndicator,
}

impl BalanceOfPower {
    /// Creates a new BalanceOfPower from the given parameters.
    pub fn new(_params: &BalanceOfPowerParams) -> Result<Self, String> {
        let bc = DEFAULT_BAR_COMPONENT;
        let qc = DEFAULT_QUOTE_COMPONENT;
        let tc = DEFAULT_TRADE_COMPONENT;

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = "bop".to_string();
        let description = "Balance of Power".to_string();

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self { line })
    }

    /// Core update with a single scalar. Since O=H=L=C, BOP is always 0.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return f64::NAN;
        }
        self.update_ohlc(sample, sample, sample, sample)
    }

    /// Updates the indicator with the given OHLC values.
    pub fn update_ohlc(&mut self, open: f64, high: f64, low: f64, close: f64) -> f64 {
        if open.is_nan() || high.is_nan() || low.is_nan() || close.is_nan() {
            return f64::NAN;
        }

        let r = high - low;
        if r < EPSILON {
            0.0
        } else {
            (close - open) / r
        }
    }
}

impl Indicator for BalanceOfPower {
    fn is_primed(&self) -> bool {
        true // Always primed.
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::BalanceOfPower,
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
        let value = self.update_ohlc(sample.open, sample.high, sample.low, sample.close);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let sample_value = (self.line.quote_func)(sample);
        let value = self.update_ohlc(sample_value, sample_value, sample_value, sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let sample_value = (self.line.trade_func)(sample);
        let value = self.update_ohlc(sample_value, sample_value, sample_value, sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::core::outputs::shape::Shape;

    fn test_open() -> Vec<f64> {
        vec![
            92.500, 91.500, 95.155, 93.970, 95.500, 94.500, 95.000, 91.500, 91.815, 91.125,
            93.875, 97.500, 98.815, 92.000, 91.125, 91.875, 93.405, 89.750, 89.345, 92.250,
        ]
    }

    fn test_high() -> Vec<f64> {
        vec![
            93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000,
            96.250000, 99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000,
        ]
    }

    fn test_low() -> Vec<f64> {
        vec![
            90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000,
            92.750000, 96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000,
        ]
    }

    fn test_close() -> Vec<f64> {
        vec![
            91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
            96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
        ]
    }

    fn test_expected() -> Vec<f64> {
        vec![
            -0.400000000000000, 0.937765205091938, -0.367058823529412, 0.418215613382900, -0.540031397174254,
            0.102459016393443, -0.823333333333333, 0.314861460957179, -0.495049504950495, 0.632941176470588,
            0.642857142857143, -0.075528700906344, -0.101777059773828, -0.540025412960610, -0.027382256297919,
            0.406926406926406, -0.944444444444444, -0.216000000000001, 0.838235294117648, -0.950000000000000,
        ]
    }

    #[test]
    fn test_ohlc() {
        let open = test_open();
        let high = test_high();
        let low = test_low();
        let close = test_close();
        let expected = test_expected();

        let mut bop = BalanceOfPower::new(&BalanceOfPowerParams).unwrap();

        for i in 0..open.len() {
            let v = bop.update_ohlc(open[i], high[i], low[i], close[i]);
            assert!(!v.is_nan(), "[{}] expected non-NaN", i);
            assert!(bop.is_primed(), "[{}] expected primed", i);
            assert!((v - expected[i]).abs() < 1e-13, "[{}] expected {}, got {}", i, expected[i], v);
        }
    }

    #[test]
    fn test_is_primed() {
        let mut bop = BalanceOfPower::new(&BalanceOfPowerParams).unwrap();
        assert!(bop.is_primed()); // Always primed.

        bop.update_ohlc(92.5, 93.25, 90.75, 91.5);
        assert!(bop.is_primed());
    }

    #[test]
    fn test_nan() {
        let mut bop = BalanceOfPower::new(&BalanceOfPowerParams).unwrap();

        assert!(bop.update(f64::NAN).is_nan());
        assert!(bop.update_ohlc(f64::NAN, 1.0, 2.0, 3.0).is_nan());
        assert!(bop.update_ohlc(1.0, f64::NAN, 2.0, 3.0).is_nan());
        assert!(bop.update_ohlc(1.0, 2.0, f64::NAN, 3.0).is_nan());
        assert!(bop.update_ohlc(1.0, 2.0, 3.0, f64::NAN).is_nan());
    }

    #[test]
    fn test_zero_range() {
        let mut bop = BalanceOfPower::new(&BalanceOfPowerParams).unwrap();
        let v = bop.update_ohlc(0.001, 0.001, 0.001, 0.001);
        assert_eq!(v, 0.0);
    }

    #[test]
    fn test_scalar_always_zero() {
        let mut bop = BalanceOfPower::new(&BalanceOfPowerParams).unwrap();
        assert_eq!(bop.update(50.0), 0.0);
        assert_eq!(bop.update(100.0), 0.0);
    }

    #[test]
    fn test_metadata() {
        let bop = BalanceOfPower::new(&BalanceOfPowerParams).unwrap();
        let meta = bop.metadata();

        assert_eq!(meta.identifier, Identifier::BalanceOfPower);
        assert_eq!(meta.mnemonic, "bop");
        assert_eq!(meta.description, "Balance of Power");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, BalanceOfPowerOutput::Value as i32);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_update_bar() {
        let open = test_open();
        let high = test_high();
        let low = test_low();
        let close = test_close();
        let expected = test_expected();

        let mut bop = BalanceOfPower::new(&BalanceOfPowerParams).unwrap();

        for i in 0..open.len() {
            let bar = Bar {
                time: 0,
                open: open[i],
                high: high[i],
                low: low[i],
                close: close[i],
                volume: 0.0,
            };
            let out = bop.update_bar(&bar);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert!((sv.value - expected[i]).abs() < 1e-13, "[{}] expected {}, got {}", i, expected[i], sv.value);
        }
    }
}
