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
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an EMA from length.
pub struct ExponentialMovingAverageLengthParams {
    /// The length (number of time periods) of the moving window.
    /// Must be >= 1.
    pub length: i64,
    /// Whether the first EMA value is the simple average of the first `length` samples.
    pub first_is_average: bool,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for ExponentialMovingAverageLengthParams {
    fn default() -> Self {
        Self {
            length: 20,
            first_is_average: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Parameters to create an EMA from smoothing factor (alpha).
pub struct ExponentialMovingAverageSmoothingFactorParams {
    /// The smoothing factor alpha in [0, 1].
    pub smoothing_factor: f64,
    /// Whether the first EMA value is the simple average of the first `length` samples.
    pub first_is_average: bool,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for ExponentialMovingAverageSmoothingFactorParams {
    fn default() -> Self {
        Self {
            smoothing_factor: 0.0952,
            first_is_average: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the exponential moving average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum ExponentialMovingAverageOutput {
    /// The scalar value of the moving average.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the exponential (exponentially weighted) moving average (EMA).
///
/// EMAᵢ = EMAᵢ₋₁ + α(Pᵢ − EMAᵢ₋₁), 0 < α ≤ 1.
///
/// The indicator is not primed during the first ℓ−1 updates.
pub struct ExponentialMovingAverage {
    line: LineIndicator,
    value: f64,
    sum: f64,
    smoothing_factor: f64,
    length: i64,
    count: i64,
    first_is_average: bool,
    primed: bool,
}

impl ExponentialMovingAverage {
    /// Creates a new EMA from length-based parameters.
    pub fn new_from_length(params: &ExponentialMovingAverageLengthParams) -> Result<Self, String> {
        Self::new_internal(params.length, f64::NAN, params.first_is_average,
            params.bar_component, params.quote_component, params.trade_component)
    }

    /// Creates a new EMA from smoothing-factor-based parameters.
    pub fn new_from_smoothing_factor(params: &ExponentialMovingAverageSmoothingFactorParams) -> Result<Self, String> {
        Self::new_internal(0, params.smoothing_factor, params.first_is_average,
            params.bar_component, params.quote_component, params.trade_component)
    }

    fn new_internal(
        length: i64,
        alpha: f64,
        first_is_average: bool,
        bc_opt: Option<BarComponent>,
        qc_opt: Option<QuoteComponent>,
        tc_opt: Option<TradeComponent>,
    ) -> Result<Self, String> {
        const INVALID: &str = "invalid exponential moving average parameters";
        const EPSILON: f64 = 0.00000001;

        let bc = bc_opt.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = qc_opt.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = tc_opt.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let (actual_length, actual_alpha, mnemonic);

        if alpha.is_nan() {
            // Length-based construction.
            if length < 1 {
                return Err(format!("{}: length should be positive", INVALID));
            }
            actual_alpha = 2.0 / (1 + length) as f64;
            actual_length = length;
            mnemonic = format!("ema({}{})", length, component_triple_mnemonic(bc, qc, tc));
        } else {
            // Smoothing-factor-based construction.
            if alpha < 0.0 || alpha > 1.0 {
                return Err(format!("{}: smoothing factor should be in range [0, 1]", INVALID));
            }
            let clamped = if alpha < EPSILON { EPSILON } else { alpha };
            actual_length = (2.0_f64 / clamped).round() as i64 - 1;
            actual_alpha = clamped;
            mnemonic = format!("ema({}, {:.8}{})", actual_length, clamped, component_triple_mnemonic(bc, qc, tc));
        }

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let description = format!("Exponential moving average {}", mnemonic);
        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            value: 0.0,
            sum: 0.0,
            smoothing_factor: actual_alpha,
            length: actual_length,
            count: 0,
            first_is_average,
            primed: false,
        })
    }

    /// Core update logic. Returns the EMA value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let temp = sample;

        if self.primed {
            self.value += (temp - self.value) * self.smoothing_factor;
        } else {
            self.count += 1;
            if self.first_is_average {
                self.sum += temp;
                if self.count < self.length {
                    return f64::NAN;
                }
                self.value = self.sum / self.length as f64;
            } else {
                if self.count == 1 {
                    self.value = temp;
                } else {
                    self.value += (temp - self.value) * self.smoothing_factor;
                }
                if self.count < self.length {
                    return f64::NAN;
                }
            }
            self.primed = true;
        }

        self.value
    }
}

impl Indicator for ExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::ExponentialMovingAverage,
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
    use crate::entities::bar_component::BarComponent;
    use crate::entities::quote_component::QuoteComponent;
    use crate::entities::trade_component::TradeComponent;
    use crate::indicators::core::outputs::shape::Shape;

    #[allow(clippy::excessive_precision)]
    fn test_input() -> Vec<f64> {
        vec![
            91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
            96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
            88.375000, 87.625000, 84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000,
            85.250000, 87.125000, 85.815000, 88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000,
            83.375000, 85.500000, 89.190000, 89.440000, 91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000,
            89.030000, 88.815000, 84.280000, 83.500000, 82.690000, 84.750000, 85.655000, 86.190000, 88.940000, 89.280000,
            88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000, 93.155000, 91.720000, 90.000000, 89.690000,
            88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000, 104.940000, 106.000000, 102.500000,
            102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000,
            112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000,
            110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
            116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000,
            124.750000, 123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000,
            132.250000, 131.000000, 132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000,
            136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000,
            125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000, 122.190000, 119.310000,
            123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000, 124.440000,
            122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
            132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000,
            130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000,
            117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
            107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
            94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000,
            95.000000, 95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000,
            105.000000, 104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000,
            113.370000, 109.000000, 109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000,
            108.000000, 108.620000, 109.750000, 109.810000, 109.000000, 108.750000, 107.870000,
        ]
    }

    fn create_ema_length(length: i64, first_is_average: bool) -> ExponentialMovingAverage {
        ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
            length,
            first_is_average,
            ..Default::default()
        }).unwrap()
    }

    fn create_ema_alpha(alpha: f64, first_is_average: bool) -> ExponentialMovingAverage {
        ExponentialMovingAverage::new_from_smoothing_factor(&ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: alpha,
            first_is_average,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_update_length_2_first_is_average_true() {
        let mut ema = create_ema_length(2, true);
        let input = test_input();

        // Index 0: not primed, should be NaN
        assert!(ema.update(input[0]).is_nan(), "[0] expected NaN");

        // Index 1: primed, first value = average of first 2 = (91.5 + 94.815) / 2 = 93.1575
        let act = ema.update(input[1]);
        assert!((93.15 - act).abs() < 1e-2, "[1] expected 93.15, got {}", act);

        // Index 2
        let act = ema.update(input[2]);
        assert!((93.96 - act).abs() < 1e-2, "[2] expected 93.96, got {}", act);

        // Index 3
        let act = ema.update(input[3]);
        assert!((94.71 - act).abs() < 1e-2, "[3] expected 94.71, got {}", act);

        // Feed remaining
        for i in 4..input.len() {
            ema.update(input[i]);
        }

        // Recompute from scratch for last value check
        let mut ema2 = create_ema_length(2, true);
        let mut last = f64::NAN;
        for i in 0..input.len() {
            last = ema2.update(input[i]);
        }
        assert!((108.21 - last).abs() < 1e-2, "[251] expected 108.21, got {}", last);

        assert!(ema2.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_10_first_is_average_true() {
        let mut ema = create_ema_length(10, true);
        let input = test_input();

        for i in 0..9 {
            assert!(ema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let act = ema.update(input[9]);
        assert!((93.22 - act).abs() < 1e-2, "[9] expected 93.22, got {}", act);

        let act = ema.update(input[10]);
        assert!((93.75 - act).abs() < 1e-2, "[10] expected 93.75, got {}", act);

        for i in 11..input.len() {
            ema.update(input[i]);
        }

        // Recompute for index 29 and 251
        let mut ema2 = create_ema_length(10, true);
        for i in 0..input.len() {
            let act = ema2.update(input[i]);
            match i {
                29 => assert!((86.46 - act).abs() < 1e-2, "[29] expected 86.46, got {}", act),
                251 => assert!((108.97 - act).abs() < 1e-2, "[251] expected 108.97, got {}", act),
                _ => {}
            }
        }

        assert!(ema2.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_2_first_is_average_false() {
        let mut ema = create_ema_length(2, false);
        let input = test_input();

        // Index 0: not primed (length=2, count=1 < 2), NaN
        assert!(ema.update(input[0]).is_nan(), "[0] expected NaN");

        // Index 1
        let act = ema.update(input[1]);
        assert!((93.71 - act).abs() < 1e-2, "[1] expected 93.71, got {}", act);

        // Index 2
        let act = ema.update(input[2]);
        assert!((94.15 - act).abs() < 1e-2, "[2] expected 94.15, got {}", act);

        // Index 3
        let act = ema.update(input[3]);
        assert!((94.78 - act).abs() < 1e-2, "[3] expected 94.78, got {}", act);

        for i in 4..input.len() {
            ema.update(input[i]);
        }

        let mut ema2 = create_ema_length(2, false);
        let mut last = f64::NAN;
        for i in 0..input.len() {
            last = ema2.update(input[i]);
        }
        assert!((108.21 - last).abs() < 1e-2, "[251] expected 108.21, got {}", last);

        assert!(ema2.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_10_first_is_average_false() {
        let mut ema = create_ema_length(10, false);
        let input = test_input();

        for i in 0..9 {
            assert!(ema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let act = ema.update(input[9]);
        assert!((92.60 - act).abs() < 1e-2, "[9] expected 92.60, got {}", act);

        let act = ema.update(input[10]);
        assert!((93.24 - act).abs() < 1e-2, "[10] expected 93.24, got {}", act);

        let act = ema.update(input[11]);
        assert!((93.97 - act).abs() < 1e-2, "[11] expected 93.97, got {}", act);

        for i in 12..input.len() {
            ema.update(input[i]);
        }

        let mut ema2 = create_ema_length(10, false);
        for i in 0..input.len() {
            let act = ema2.update(input[i]);
            match i {
                30 => assert!((86.23 - act).abs() < 1e-2, "[30] expected 86.23, got {}", act),
                251 => assert!((108.97 - act).abs() < 1e-2, "[251] expected 108.97, got {}", act),
                _ => {}
            }
        }

        assert!(ema2.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_entity() {
        let l: i64 = 2;
        let alpha = 2.0 / (l + 1) as f64;
        let inp = 3.0_f64;
        let exp = alpha * inp;
        let time = 1617235200;

        // scalar
        let mut ema = create_ema_length(l, false);
        ema.update(0.0);
        ema.update(0.0);
        let out = ema.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert_eq!(s.value, exp);

        // bar (default component = Close)
        let mut ema = create_ema_length(l, false);
        ema.update(0.0);
        ema.update(0.0);
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = ema.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.value, exp);

        // quote (default component = Mid = (bid+ask)/2)
        let mut ema = create_ema_length(l, false);
        ema.update(0.0);
        ema.update(0.0);
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = ema.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.value, exp);

        // trade (default component = Price)
        let mut ema = create_ema_length(l, false);
        ema.update(0.0);
        ema.update(0.0);
        let trade = Trade::new(time, inp, 0.0);
        let out = ema.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.value, exp);
    }

    #[test]
    fn test_is_primed() {
        let input = test_input();

        // firstIsAverage = true
        let mut ema = create_ema_length(10, true);
        assert!(!ema.is_primed());
        for i in 0..9 {
            ema.update(input[i]);
            assert!(!ema.is_primed(), "[{}] should not be primed", i);
        }
        for i in 9..input.len() {
            ema.update(input[i]);
            assert!(ema.is_primed(), "[{}] should be primed", i);
        }

        // firstIsAverage = false
        let mut ema = create_ema_length(10, false);
        assert!(!ema.is_primed());
        for i in 0..9 {
            ema.update(input[i]);
            assert!(!ema.is_primed(), "[{}] should not be primed", i);
        }
        for i in 9..input.len() {
            ema.update(input[i]);
            assert!(ema.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_metadata_length() {
        let ema = create_ema_length(10, true);
        let m = ema.metadata();
        assert_eq!(m.identifier, Identifier::ExponentialMovingAverage);
        assert_eq!(m.mnemonic, "ema(10)");
        assert_eq!(m.description, "Exponential moving average ema(10)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, ExponentialMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "ema(10)");
        assert_eq!(m.outputs[0].description, "Exponential moving average ema(10)");
    }

    #[test]
    fn test_metadata_alpha() {
        let alpha = 2.0 / 11.0;
        let ema = create_ema_alpha(alpha, false);
        let m = ema.metadata();
        assert_eq!(m.identifier, Identifier::ExponentialMovingAverage);
        assert_eq!(m.mnemonic, "ema(10, 0.18181818)");
        assert_eq!(m.description, "Exponential moving average ema(10, 0.18181818)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].mnemonic, "ema(10, 0.18181818)");
    }

    #[test]
    fn test_metadata_length_with_bar_component() {
        let ema = ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
            length: 10,
            first_is_average: true,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        }).unwrap();
        let m = ema.metadata();
        assert_eq!(m.mnemonic, "ema(10, hl/2)");
        assert_eq!(m.description, "Exponential moving average ema(10, hl/2)");
    }

    #[test]
    fn test_metadata_alpha_with_quote_component() {
        let ema = ExponentialMovingAverage::new_from_smoothing_factor(&ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 2.0 / 11.0,
            first_is_average: false,
            quote_component: Some(QuoteComponent::Bid),
            ..Default::default()
        }).unwrap();
        let m = ema.metadata();
        assert_eq!(m.mnemonic, "ema(10, 0.18181818, b)");
        assert_eq!(m.description, "Exponential moving average ema(10, 0.18181818, b)");
    }

    #[test]
    fn test_new_length_zero() {
        let r = ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
            length: 0, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid exponential moving average parameters: length should be positive");
    }

    #[test]
    fn test_new_length_negative() {
        let r = ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
            length: -1, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid exponential moving average parameters: length should be positive");
    }

    #[test]
    fn test_new_alpha_negative() {
        let r = ExponentialMovingAverage::new_from_smoothing_factor(&ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: -1.0, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid exponential moving average parameters: smoothing factor should be in range [0, 1]");
    }

    #[test]
    fn test_new_alpha_greater_than_1() {
        let r = ExponentialMovingAverage::new_from_smoothing_factor(&ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 2.0, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid exponential moving average parameters: smoothing factor should be in range [0, 1]");
    }

    #[test]
    fn test_new_alpha_zero_clamped() {
        let ema = ExponentialMovingAverage::new_from_smoothing_factor(&ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 0.0, ..Default::default()
        }).unwrap();
        assert_eq!(ema.smoothing_factor, 0.00000001);
        assert_eq!(ema.length, 199999999);
        assert_eq!(ema.line.mnemonic, "ema(199999999, 0.00000001)");
    }

    #[test]
    fn test_new_alpha_below_epsilon_clamped() {
        let ema = ExponentialMovingAverage::new_from_smoothing_factor(&ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 0.000000001, ..Default::default()
        }).unwrap();
        assert_eq!(ema.smoothing_factor, 0.00000001);
        assert_eq!(ema.length, 199999999);
    }

    #[test]
    fn test_new_length_1() {
        let ema = ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
            length: 1, first_is_average: true, ..Default::default()
        }).unwrap();
        assert_eq!(ema.length, 1);
        assert_eq!(ema.smoothing_factor, 1.0);
        assert_eq!(ema.line.mnemonic, "ema(1)");
    }

    #[test]
    fn test_new_alpha_1() {
        let ema = ExponentialMovingAverage::new_from_smoothing_factor(&ExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 1.0, first_is_average: true, ..Default::default()
        }).unwrap();
        assert_eq!(ema.smoothing_factor, 1.0);
        assert_eq!(ema.length, 1);
        assert_eq!(ema.line.mnemonic, "ema(1, 1.00000000)");
    }

    #[test]
    fn test_mnemonic_components() {
        // all defaults -> no component suffix
        let ema = create_ema_length(10, true);
        assert_eq!(ema.line.mnemonic, "ema(10)");

        // bar component set
        let ema = ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
            length: 10,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        }).unwrap();
        assert_eq!(ema.line.mnemonic, "ema(10, hl/2)");

        // quote component set
        let ema = ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
            length: 10,
            quote_component: Some(QuoteComponent::Bid),
            ..Default::default()
        }).unwrap();
        assert_eq!(ema.line.mnemonic, "ema(10, b)");

        // trade component set
        let ema = ExponentialMovingAverage::new_from_length(&ExponentialMovingAverageLengthParams {
            length: 10,
            trade_component: Some(TradeComponent::Volume),
            ..Default::default()
        }).unwrap();
        assert_eq!(ema.line.mnemonic, "ema(10, v)");
    }
}
