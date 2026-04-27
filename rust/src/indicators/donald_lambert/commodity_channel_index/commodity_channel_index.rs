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

/// The default inverse scaling factor.
/// The value of 0.015 ensures that approximately 70 to 80 percent of CCI values
/// fall between -100 and +100.
pub const DEFAULT_INVERSE_SCALING_FACTOR: f64 = 0.015;

/// Parameters to create an instance of the commodity channel index indicator.
pub struct CommodityChannelIndexParams {
    /// The number of time periods. Must be greater than 1.
    pub length: usize,
    /// Inverse scaling factor. 0 means use default (0.015).
    pub inverse_scaling_factor: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for CommodityChannelIndexParams {
    fn default() -> Self {
        Self {
            length: 20,
            inverse_scaling_factor: DEFAULT_INVERSE_SCALING_FACTOR,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the commodity channel index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CommodityChannelIndexOutput {
    /// The scalar value of the commodity channel index.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Donald Lambert's Commodity Channel Index (CCI).
///
/// CCI measures the deviation of the price from its statistical mean.
///
/// CCI = (typicalPrice - SMA) / (scalingFactor * meanDeviation)
///
/// where scalingFactor defaults to 0.015 so that approximately 70-80% of CCI
/// values fall between -100 and +100.
pub struct CommodityChannelIndex {
    line: LineIndicator,
    length: usize,
    scaling_factor: f64,
    window: Vec<f64>,
    window_count: usize,
    window_sum: f64,
    primed: bool,
}

impl CommodityChannelIndex {
    /// Creates a new CommodityChannelIndex from the given parameters.
    pub fn new(params: &CommodityChannelIndexParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid commodity channel index parameters: length should be greater than 1".to_string());
        }

        let inverse_factor = if params.inverse_scaling_factor == 0.0 {
            DEFAULT_INVERSE_SCALING_FACTOR
        } else {
            params.inverse_scaling_factor
        };

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("cci({}{})", params.length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Commodity Channel Index {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            length: params.length,
            scaling_factor: params.length as f64 / inverse_factor,
            window: vec![0.0; params.length],
            window_count: 0,
            window_sum: 0.0,
            primed: false,
        })
    }

    /// Core update logic. Returns the CCI value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let last_index = self.length - 1;

        if self.primed {
            self.window_sum += sample - self.window[0];

            for i in 0..last_index {
                self.window[i] = self.window[i + 1];
            }
            self.window[last_index] = sample;

            let average = self.window_sum / self.length as f64;

            let mut temp = 0.0_f64;
            for i in 0..self.length {
                temp += (self.window[i] - average).abs();
            }

            if temp.abs() < f64::MIN_POSITIVE {
                0.0
            } else {
                self.scaling_factor * (sample - average) / temp
            }
        } else {
            self.window_sum += sample;
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if self.window_count == self.length {
                self.primed = true;

                let average = self.window_sum / self.length as f64;

                let mut temp = 0.0_f64;
                for i in 0..self.length {
                    temp += (self.window[i] - average).abs();
                }

                if temp.abs() < f64::MIN_POSITIVE {
                    0.0
                } else {
                    self.scaling_factor * (sample - average) / temp
                }
            } else {
                f64::NAN
            }
        }
    }
}

impl Indicator for CommodityChannelIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CommodityChannelIndex,
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
    use crate::indicators::core::outputs::shape::Shape;

    fn test_input() -> Vec<f64> {
        vec![
            91.83333333333330, 93.72000000000000, 95.00000000000000, 94.92833333333330, 94.19833333333330, 94.28166666666670, 93.17666666666670, 92.07333333333330, 90.74166666666670, 91.94833333333330,
            95.04166666666670, 97.73000000000000, 97.88500000000000, 90.48000000000000, 89.68833333333330, 92.33500000000000, 90.48833333333330, 89.59333333333330, 90.94833333333330, 90.62500000000000,
            88.74000000000000, 87.55166666666670, 85.88500000000000, 83.59333333333330, 83.16833333333330, 82.33333333333330, 83.37666666666670, 87.57333333333330, 86.69833333333330, 87.11500000000000,
            85.60333333333330, 86.49000000000000, 86.23000000000000, 87.82333333333330, 88.78166666666670, 87.76166666666670, 86.14666666666670, 84.68833333333330, 83.83500000000000, 84.26166666666670,
            83.45833333333330, 86.35500000000000, 88.51166666666670, 89.32333333333330, 90.93833333333330, 90.77166666666670, 91.72000000000000, 89.90666666666670, 90.24000000000000, 90.78166666666670,
            89.34333333333330, 88.05333333333330, 85.76000000000000, 84.02166666666670, 82.83500000000000, 84.41666666666670, 85.67666666666670, 86.47000000000000, 88.50166666666670, 89.44833333333330,
            89.20833333333330, 88.23000000000000, 91.07333333333330, 91.99000000000000, 92.19833333333330, 92.89500000000000, 93.06166666666670, 91.35500000000000, 90.65666666666670, 90.14833333333330,
            88.45833333333330, 86.33500000000000, 83.85333333333330, 83.75000000000000, 84.81500000000000, 97.65666666666670, 99.87500000000000, 103.82333333333300, 105.95833333333300, 103.16666666666700,
            102.87500000000000, 103.93833333333300, 105.13500000000000, 106.54333333333300, 105.32333333333300, 105.20833333333300, 107.63500000000000, 109.59500000000000, 110.06333333333300, 111.57333333333300,
            121.00000000000000, 119.79166666666700, 118.18833333333300, 119.35500000000000, 117.94833333333300, 116.96000000000000, 115.49166666666700, 112.69833333333300, 111.36500000000000, 115.72000000000000,
            115.16333333333300, 115.64666666666700, 112.35333333333300, 112.60333333333300, 113.27000000000000, 114.81333333333300, 119.89666666666700, 117.51666666666700, 118.14333333333300, 115.10333333333300,
            114.45666666666700, 115.16666666666700, 116.31000000000000, 120.35333333333300, 120.39666666666700, 120.64666666666700, 124.37333333333300, 124.70666666666700, 122.96000000000000, 122.85333333333300,
            123.99666666666700, 123.14666666666700, 124.22666666666700, 128.54000000000000, 130.10333333333300, 131.33333333333300, 131.89666666666700, 132.31333333333300, 133.87666666666700, 136.23333333333300,
            137.29333333333300, 137.60666666666700, 137.31333333333300, 136.31333333333300, 136.37666666666700, 135.73333333333300, 128.81333333333300, 128.54000000000000, 125.29000000000000, 124.29000000000000,
            123.62333333333300, 125.43666666666700, 127.62666666666700, 125.53666666666700, 125.58333333333300, 123.35333333333300, 120.22666666666700, 119.47666666666700, 121.58333333333300, 123.83333333333300,
            122.58333333333300, 120.25000000000000, 122.29000000000000, 121.97666666666700, 123.29000000000000, 126.58000000000000, 127.87333333333300, 125.66666666666700, 123.02000000000000, 122.39333333333300,
            123.87333333333300, 122.20666666666700, 122.43333333333300, 123.62333333333300, 123.98000000000000, 123.83333333333300, 124.47666666666700, 127.08333333333300, 125.62333333333300, 128.87000000000000,
            131.02333333333300, 131.79333333333300, 134.29333333333300, 135.69000000000000, 133.31333333333300, 132.93666666666700, 132.96000000000000, 130.64666666666700, 126.85333333333300, 129.00333333333300,
            127.66666666666700, 125.60333333333300, 123.75000000000000, 123.48000000000000, 123.72666666666700, 123.31333333333300, 120.95666666666700, 120.95666666666700, 118.10333333333300, 118.87333333333300,
            121.43666666666700, 120.33333333333300, 116.87333333333300, 113.20666666666700, 114.68666666666700, 111.62333333333300, 106.97666666666700, 106.31333333333300, 106.87000000000000, 106.89666666666700,
            107.02000000000000, 109.02000000000000, 91.00000000000000, 93.68666666666670, 93.70333333333330, 95.37333333333330, 93.79000000000000, 94.83333333333330, 97.83333333333330, 97.31000000000000,
            95.10333333333330, 94.60333333333330, 92.00000000000000, 91.12666666666670, 92.79333333333330, 93.74666666666670, 96.06000000000000, 95.79000000000000, 95.04000000000000, 94.76666666666670,
            94.20666666666670, 93.74666666666670, 96.60333333333330, 102.47666666666700, 106.91666666666700, 107.31000000000000, 103.77000000000000, 105.04000000000000, 104.16666666666700, 103.22666666666700,
            103.37000000000000, 104.98333333333300, 110.89333333333300, 115.00000000000000, 117.08333333333300, 118.26000000000000, 115.91333333333300, 109.50000000000000, 109.67000000000000, 108.77000000000000,
            106.48000000000000, 108.21000000000000, 109.89333333333300, 109.13000000000000, 109.43333333333300, 108.77000000000000, 109.08333333333300, 109.29000000000000, 109.87333333333300, 109.41666666666700,
            109.27000000000000, 107.99666666666700,
        ]
    }

    #[test]
    fn test_length_11() {
        let tolerance = 5e-8;
        let input = test_input();

        let mut cci = CommodityChannelIndex::new(&CommodityChannelIndexParams {
            length: 11,
            ..Default::default()
        }).unwrap();

        // First 10 values should be NaN.
        for i in 0..10 {
            let v = cci.update(input[i]);
            assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
        }

        // Index 10: first value.
        let v = cci.update(input[10]);
        assert!(!v.is_nan(), "[10] expected non-NaN");
        assert!((v - 87.92686612269590).abs() < tolerance, "[10] expected ~87.9269, got {}", v);

        // Index 11.
        let v = cci.update(input[11]);
        assert!((v - 180.00543014506300).abs() < tolerance, "[11] expected ~180.0054, got {}", v);

        // Feed remaining and check last.
        for i in 12..251 {
            cci.update(input[i]);
        }

        let v = cci.update(input[251]);
        assert!((v - (-169.65514382823800)).abs() < tolerance, "[251] expected ~-169.6551, got {}", v);
        assert!(cci.is_primed());
    }

    #[test]
    fn test_length_2() {
        let tolerance = 5e-7;
        let input = test_input();

        let mut cci = CommodityChannelIndex::new(&CommodityChannelIndexParams {
            length: 2,
            ..Default::default()
        }).unwrap();

        // First value should be NaN.
        let v = cci.update(input[0]);
        assert!(v.is_nan());

        // Index 1: first value.
        let v = cci.update(input[1]);
        assert!(!v.is_nan());
        assert!((v - 66.66666666666670).abs() < tolerance, "[1] expected ~66.6667, got {}", v);

        // Feed remaining and check last.
        for i in 2..251 {
            cci.update(input[i]);
        }

        let v = cci.update(input[251]);
        assert!((v - (-66.66666666666590)).abs() < tolerance, "[251] expected ~-66.6667, got {}", v);
    }

    #[test]
    fn test_is_primed() {
        let mut cci = CommodityChannelIndex::new(&CommodityChannelIndexParams {
            length: 5,
            ..Default::default()
        }).unwrap();

        assert!(!cci.is_primed());

        for i in 1..=4 {
            cci.update(i as f64);
            assert!(!cci.is_primed(), "[{}] expected not primed", i);
        }

        cci.update(5.0);
        assert!(cci.is_primed());

        cci.update(6.0);
        assert!(cci.is_primed());
    }

    #[test]
    fn test_nan() {
        let mut cci = CommodityChannelIndex::new(&CommodityChannelIndexParams {
            length: 5,
            ..Default::default()
        }).unwrap();

        assert!(cci.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_metadata() {
        let cci = CommodityChannelIndex::new(&CommodityChannelIndexParams {
            length: 20,
            ..Default::default()
        }).unwrap();

        let meta = cci.metadata();
        assert_eq!(meta.identifier, Identifier::CommodityChannelIndex);
        assert_eq!(meta.mnemonic, "cci(20)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, CommodityChannelIndexOutput::Value as i32);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
    }

    #[test]
    fn test_update_entity() {
        let input = test_input();

        let mut cci = CommodityChannelIndex::new(&CommodityChannelIndexParams {
            length: 11,
            ..Default::default()
        }).unwrap();

        for i in 0..10 {
            let scalar = Scalar::new(0, input[i]);
            let out = cci.update_scalar(&scalar);
            let sv = out[0].downcast_ref::<Scalar>().unwrap();
            assert!(sv.value.is_nan(), "[{}] expected NaN", i);
        }

        let scalar = Scalar::new(0, input[10]);
        let out = cci.update_scalar(&scalar);
        let sv = out[0].downcast_ref::<Scalar>().unwrap();
        assert!(!sv.value.is_nan(), "[10] expected non-NaN");
    }

    #[test]
    fn test_invalid_params() {
        assert!(CommodityChannelIndex::new(&CommodityChannelIndexParams {
            length: 1,
            ..Default::default()
        }).is_err());

        assert!(CommodityChannelIndex::new(&CommodityChannelIndexParams {
            length: 0,
            ..Default::default()
        }).is_err());
    }

    #[test]
    fn test_custom_scaling_factor() {
        let mut cci = CommodityChannelIndex::new(&CommodityChannelIndexParams {
            length: 5,
            inverse_scaling_factor: 0.03,
            ..Default::default()
        }).unwrap();

        for i in 1..=5 {
            cci.update(i as f64);
        }

        assert!(cci.is_primed());
    }
}
