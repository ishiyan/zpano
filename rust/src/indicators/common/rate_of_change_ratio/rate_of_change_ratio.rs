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

/// Parameters to create an instance of the rate of change ratio indicator.
pub struct RateOfChangeRatioParams {
    /// The length (number of time periods).
    /// Must be greater than 0.
    pub length: usize,
    /// If true, multiply result by 100 (ROCR100 mode).
    pub hundred_scale: bool,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for RateOfChangeRatioParams {
    fn default() -> Self {
        Self {
            length: 10,
            hundred_scale: false,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the rate of change ratio indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum RateOfChangeRatioOutput {
    /// The scalar value of the rate of change ratio.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Rate of Change Ratio (ROCR).
///
/// ROCRi = Pi / Pi-l
///
/// Optionally scaled by 100 (ROCR100).
///
/// where l is the length.
///
/// The indicator is not primed during the first l updates.
pub struct RateOfChangeRatio {
    line: LineIndicator,
    window: Vec<f64>,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    hundred_scale: bool,
    primed: bool,
}

impl RateOfChangeRatio {
    /// Creates a new RateOfChangeRatio from the given parameters.
    pub fn new(params: &RateOfChangeRatioParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err("invalid rate of change ratio parameters: length should be positive".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let triple = component_triple_mnemonic(bc, qc, tc);

        let (mnemonic, description) = if params.hundred_scale {
            (
                format!("rocr100({}{})", params.length, triple),
                format!("Rate of Change Ratio 100 Scale rocr100({}{})", params.length, triple),
            )
        } else {
            (
                format!("rocr({}{})", params.length, triple),
                format!("Rate of Change Ratio rocr({}{})", params.length, triple),
            )
        };

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let window_length = params.length + 1;

        Ok(Self {
            line,
            window: vec![0.0; window_length],
            window_length,
            window_count: 0,
            last_index: params.length,
            hundred_scale: params.hundred_scale,
            primed: false,
        })
    }

    /// Core update logic. Returns the rate of change ratio value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        const EPSILON: f64 = 1e-13;
        let scale = if self.hundred_scale { 100.0 } else { 1.0 };

        if self.primed {
            for i in 0..self.last_index {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = sample;
            let previous = self.window[0];

            if previous.abs() > EPSILON {
                return (sample / previous) * scale;
            }

            return 0.0;
        }

        self.window[self.window_count] = sample;
        self.window_count += 1;

        if self.window_length == self.window_count {
            self.primed = true;
            let previous = self.window[0];

            if previous.abs() > EPSILON {
                return (sample / previous) * scale;
            }

            return 0.0;
        }

        f64::NAN
    }
}

impl Indicator for RateOfChangeRatio {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::RateOfChangeRatio,
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

    fn create_rocr(length: usize, hundred_scale: bool) -> RateOfChangeRatio {
        RateOfChangeRatio::new(&RateOfChangeRatioParams { length, hundred_scale, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_update_rocr_length_14() {
        let mut rocr = create_rocr(14, false);
        let input = test_input();

        for i in 0..13 {
            assert!(rocr.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 13..input.len() {
            let act = rocr.update(input[i]);

            match i {
                14 => assert!((act - 0.994536).abs() < 1e-4, "[14] expected 0.994536, got {}", act),
                15 => assert!((act - 0.978906).abs() < 1e-4, "[15] expected 0.978906, got {}", act),
                16 => assert!((act - 0.944689).abs() < 1e-4, "[16] expected 0.944689, got {}", act),
                251 => assert!((act - 0.989633).abs() < 1e-4, "[251] expected 0.989633, got {}", act),
                _ => {}
            }
        }

        assert!(rocr.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_rocr100_length_14() {
        let mut rocr100 = create_rocr(14, true);
        let input = test_input();

        for i in 0..13 {
            assert!(rocr100.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 13..input.len() {
            let act = rocr100.update(input[i]);

            match i {
                14 => assert!((act - 99.4536).abs() < 1e-2, "[14] expected 99.4536, got {}", act),
                15 => assert!((act - 97.8906).abs() < 1e-2, "[15] expected 97.8906, got {}", act),
                16 => assert!((act - 94.4689).abs() < 1e-2, "[16] expected 94.4689, got {}", act),
                251 => assert!((act - 98.9633).abs() < 1e-2, "[251] expected 98.9633, got {}", act),
                _ => {}
            }
        }

        assert!(rocr100.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_entity() {
        let length = 2;
        let inp = 3.0_f64;
        let exp = 1.0_f64; // rocr = 3/3 = 1
        let time = 1617235200;

        // scalar
        let mut rocr = create_rocr(length, false);
        rocr.update(inp);
        rocr.update(inp);
        let out = rocr.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!((s.value - exp).abs() < 1e-13);

        // bar
        let mut rocr = create_rocr(length, false);
        rocr.update(inp);
        rocr.update(inp);
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = rocr.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp).abs() < 1e-13);

        // quote
        let mut rocr = create_rocr(length, false);
        rocr.update(inp);
        rocr.update(inp);
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = rocr.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp).abs() < 1e-13);

        // trade
        let mut rocr = create_rocr(length, false);
        rocr.update(inp);
        rocr.update(inp);
        let trade = Trade::new(time, inp, 0.0);
        let out = rocr.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp).abs() < 1e-13);
    }

    #[test]
    fn test_is_primed() {
        let input = test_input();

        for &length in &[1_usize, 2, 3, 5, 10] {
            let mut rocr = create_rocr(length, false);
            assert!(!rocr.is_primed());

            for i in 0..length {
                rocr.update(input[i]);
                assert!(!rocr.is_primed(), "length={}, [{}] should not be primed", length, i);
            }

            for i in length..input.len() {
                rocr.update(input[i]);
                assert!(rocr.is_primed(), "length={}, [{}] should be primed", length, i);
            }
        }
    }

    #[test]
    fn test_metadata_rocr() {
        let rocr = create_rocr(5, false);
        let m = rocr.metadata();
        assert_eq!(m.identifier, Identifier::RateOfChangeRatio);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, RateOfChangeRatioOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "rocr(5)");
        assert_eq!(m.outputs[0].description, "Rate of Change Ratio rocr(5)");
    }

    #[test]
    fn test_metadata_rocr100() {
        let rocr100 = create_rocr(5, true);
        let m = rocr100.metadata();
        assert_eq!(m.identifier, Identifier::RateOfChangeRatio);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].mnemonic, "rocr100(5)");
        assert_eq!(m.outputs[0].description, "Rate of Change Ratio 100 Scale rocr100(5)");
    }

    #[test]
    fn test_new_invalid_length() {
        let r = RateOfChangeRatio::new(&RateOfChangeRatioParams { length: 0, ..Default::default() });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid rate of change ratio parameters: length should be positive");
    }

    #[test]
    fn test_mnemonic_components() {
        let rocr = create_rocr(5, false);
        assert_eq!(rocr.line.mnemonic, "rocr(5)");

        let rocr = RateOfChangeRatio::new(&RateOfChangeRatioParams {
            length: 5, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(rocr.line.mnemonic, "rocr(5, hl/2)");

        let rocr100 = RateOfChangeRatio::new(&RateOfChangeRatioParams {
            length: 5, hundred_scale: true, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(rocr100.line.mnemonic, "rocr100(5, hl/2)");

        let rocr = RateOfChangeRatio::new(&RateOfChangeRatioParams {
            length: 5, quote_component: Some(QuoteComponent::Bid), ..Default::default()
        }).unwrap();
        assert_eq!(rocr.line.mnemonic, "rocr(5, b)");

        let rocr = RateOfChangeRatio::new(&RateOfChangeRatioParams {
            length: 5, trade_component: Some(TradeComponent::Volume), ..Default::default()
        }).unwrap();
        assert_eq!(rocr.line.mnemonic, "rocr(5, v)");
    }
}
