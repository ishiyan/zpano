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
    use super::super::testdata::testdata;
    use crate::indicators::core::outputs::shape::Shape;
    #[test]
    fn test_length_11() {
        let tolerance = 5e-8;
        let input = testdata::test_input();

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
        let input = testdata::test_input();

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
        let input = testdata::test_input();

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
