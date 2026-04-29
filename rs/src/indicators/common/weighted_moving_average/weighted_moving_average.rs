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

/// Parameters to create an instance of the weighted moving average indicator.
pub struct WeightedMovingAverageParams {
    /// The length (number of time periods) of the moving window. Must be > 1.
    pub length: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for WeightedMovingAverageParams {
    fn default() -> Self {
        Self { length: 20, bar_component: None, quote_component: None, trade_component: None }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the weighted moving average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum WeightedMovingAverageOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the weighted moving average (WMA) with arithmetically decreasing weights.
///
/// WMAᵢ = (ℓPᵢ + (ℓ-1)Pᵢ₋₁ + … + Pᵢ₋ℓ) / ½ℓ(ℓ+1)
pub struct WeightedMovingAverage {
    line: LineIndicator,
    window: Vec<f64>,
    window_sum: f64,
    window_sub: f64,
    divider: f64,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    primed: bool,
}

impl WeightedMovingAverage {
    /// Creates a new WeightedMovingAverage from the given parameters.
    pub fn new(params: &WeightedMovingAverageParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid weighted moving average parameters: length should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("wma({}{})", params.length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Weighted moving average {}", mnemonic);
        let divider = params.length as f64 * (params.length + 1) as f64 / 2.0;

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            window: vec![0.0; params.length],
            window_sum: 0.0,
            window_sub: 0.0,
            divider,
            window_length: params.length,
            window_count: 0,
            last_index: params.length - 1,
            primed: false,
        })
    }

    /// Core update logic.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let temp = sample;

        if self.primed {
            self.window_sum -= self.window_sub;
            self.window_sum += temp * self.window_length as f64;
            self.window_sub -= self.window[0];
            self.window_sub += temp;

            for i in 0..self.last_index {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = temp;
        } else {
            self.window[self.window_count] = temp;
            self.window_sub += temp;
            self.window_count += 1;
            self.window_sum += temp * self.window_count as f64;

            if self.window_length > self.window_count {
                return f64::NAN;
            }

            self.primed = true;
        }

        self.window_sum / self.divider
    }
}

impl Indicator for WeightedMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::WeightedMovingAverage,
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
        let v = (self.line.bar_func)(sample);
        let value = self.update(v);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.line.quote_func)(sample);
        let value = self.update(v);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.line.trade_func)(sample);
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
    use crate::indicators::core::outputs::shape::Shape;

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

    fn create_wma(length: usize) -> WeightedMovingAverage {
        WeightedMovingAverage::new(&WeightedMovingAverageParams { length, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_update_length_2() {
        let mut wma = create_wma(2);
        let input = test_input();

        assert!(wma.update(input[0]).is_nan());

        let act = wma.update(input[1]);
        assert!((93.71 - act).abs() < 1e-2, "[1] expected 93.71, got {}", act);

        let act = wma.update(input[2]);
        assert!((94.52 - act).abs() < 1e-2, "[2] expected 94.52, got {}", act);

        let act = wma.update(input[3]);
        assert!((94.855 - act).abs() < 1e-2, "[3] expected 94.855, got {}", act);

        // Feed rest and check last
        for i in 4..input.len() {
            wma.update(input[i]);
        }
        // Rewind not possible, just check NaN passthrough
        assert!(wma.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_30() {
        let mut wma = create_wma(30);
        let input = test_input();

        for i in 0..29 {
            assert!(wma.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let act = wma.update(input[29]);
        assert!((88.5677 - act).abs() < 1e-2, "[29] expected 88.5677, got {}", act);

        let act = wma.update(input[30]);
        assert!((88.2337 - act).abs() < 1e-2, "[30] expected 88.2337, got {}", act);

        let act = wma.update(input[31]);
        assert!((88.034 - act).abs() < 1e-2, "[31] expected 88.034, got {}", act);

        // Feed through to index 58
        for i in 32..58 {
            wma.update(input[i]);
        }

        let act = wma.update(input[58]);
        assert!((87.191 - act).abs() < 1e-2, "[58] expected 87.191, got {}", act);

        // Feed through to check last values
        for i in 59..250 {
            wma.update(input[i]);
        }

        let act = wma.update(input[250]);
        assert!((109.3466 - act).abs() < 1e-2, "[250] expected 109.3466, got {}", act);

        let act = wma.update(input[251]);
        assert!((109.3413 - act).abs() < 1e-2, "[251] expected 109.3413, got {}", act);

        assert!(wma.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_is_primed() {
        let input = test_input();

        let mut wma = create_wma(2);
        assert!(!wma.is_primed());
        wma.update(input[0]);
        assert!(!wma.is_primed());
        for i in 1..input.len() {
            wma.update(input[i]);
            assert!(wma.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_metadata() {
        let wma = create_wma(5);
        let m = wma.metadata();
        assert_eq!(m.identifier, Identifier::WeightedMovingAverage);
        assert_eq!(m.mnemonic, "wma(5)");
        assert_eq!(m.description, "Weighted moving average wma(5)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, WeightedMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "wma(5)");
    }

    #[test]
    fn test_new_invalid_length() {
        assert!(WeightedMovingAverage::new(&WeightedMovingAverageParams { length: 1, ..Default::default() }).is_err());
        assert!(WeightedMovingAverage::new(&WeightedMovingAverageParams { length: 0, ..Default::default() }).is_err());
    }

    #[test]
    fn test_update_entity() {
        let input = test_input();
        let time = 1617235200_i64;

        // scalar
        let mut wma = create_wma(2);
        wma.update(input[0]);
        let out = wma.update_scalar(&Scalar::new(time, input[1]));
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((93.71 - s.value).abs() < 1e-2);
        assert_eq!(s.time, time);

        // bar
        let mut wma = create_wma(2);
        wma.update(input[0]);
        let bar = Bar::new(time, 0.0, 0.0, 0.0, input[1], 0.0);
        let out = wma.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((93.71 - s.value).abs() < 1e-2);
    }
}
