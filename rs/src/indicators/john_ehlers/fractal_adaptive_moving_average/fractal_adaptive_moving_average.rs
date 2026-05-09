use crate::entities::bar::Bar;
use crate::entities::bar_component::{
    component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT,
};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{
    component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT,
};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{
    component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT,
};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Fractal Adaptive Moving Average indicator.
pub struct FractalAdaptiveMovingAverageParams {
    /// The length (number of time periods). Must be >= 2; odd values are rounded up to the next even integer.
    /// Default is 16.
    pub length: i64,
    /// The slowest boundary smoothing factor, αs in [0,1]. Default is 0.01.
    pub slowest_smoothing_factor: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for FractalAdaptiveMovingAverageParams {
    fn default() -> Self {
        Self {
            length: 16,
            slowest_smoothing_factor: 0.01,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Fractal Adaptive Moving Average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum FractalAdaptiveMovingAverageOutput {
    /// The scalar value of the fractal adaptive moving average.
    Value = 1,
    /// The scalar value of the estimated fractal dimension.
    Fdim = 2,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Fractal Adaptive Moving Average (FRAMA).
pub struct FractalAdaptiveMovingAverage {
    line: LineIndicator,
    mnemonic_fdim: String,
    description_fdim: String,
    alpha_slowest: f64,
    scaling_factor: f64,
    fractal_dimension: f64,
    value: f64,
    length: usize,
    length_min_one: usize,
    half_length: usize,
    window_count: usize,
    window_high: Vec<f64>,
    window_low: Vec<f64>,
    primed: bool,
}

impl FractalAdaptiveMovingAverage {
    /// Creates a new FRAMA from the supplied parameters.
    pub fn new(params: &FractalAdaptiveMovingAverageParams) -> Result<Self, String> {
        const INVALID: &str = "invalid fractal adaptive moving average parameters";

        if params.length < 2 {
            return Err(format!(
                "{}: length should be an even integer larger than 1",
                INVALID
            ));
        }

        if params.slowest_smoothing_factor < 0.0 || params.slowest_smoothing_factor > 1.0 {
            return Err(format!(
                "{}: slowest smoothing factor should be in range [0, 1]",
                INVALID
            ));
        }

        let mut length = params.length as usize;
        if length % 2 != 0 {
            length += 1;
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let comp_mnemonic = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!("frama({}, {:.3}{})", length, params.slowest_smoothing_factor, comp_mnemonic);
        let mnemonic_fdim = format!("framaDim({}, {:.3}{})", length, params.slowest_smoothing_factor, comp_mnemonic);
        let descr = "Fractal adaptive moving average ";
        let description = format!("{}{}", descr, mnemonic);
        let description_fdim = format!("{}{}", descr, mnemonic_fdim);

        let line = LineIndicator::new(
            mnemonic,
            description,
            bar_func,
            quote_func,
            trade_func,
        );

        Ok(Self {
            line,
            mnemonic_fdim,
            description_fdim,
            alpha_slowest: params.slowest_smoothing_factor,
            scaling_factor: params.slowest_smoothing_factor.ln(),
            fractal_dimension: f64::NAN,
            value: f64::NAN,
            length,
            length_min_one: length - 1,
            half_length: length / 2,
            window_count: 0,
            window_high: vec![0.0; length],
            window_low: vec![0.0; length],
            primed: false,
        })
    }

    /// Core update logic. Takes sample, high, and low values.
    /// Returns the FRAMA value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64, sample_high: f64, sample_low: f64) -> f64 {
        if sample_high.is_nan() || sample_low.is_nan() || sample.is_nan() {
            return f64::NAN;
        }

        if self.primed {
            for i in 0..self.length_min_one {
                let j = i + 1;
                self.window_high[i] = self.window_high[j];
                self.window_low[i] = self.window_low[j];
            }

            self.window_high[self.length_min_one] = sample_high;
            self.window_low[self.length_min_one] = sample_low;

            self.fractal_dimension = self.estimate_fractal_dimension();
            let alpha = self.estimate_alpha();
            self.value += (sample - self.value) * alpha;

            return self.value;
        }

        self.window_high[self.window_count] = sample_high;
        self.window_low[self.window_count] = sample_low;

        self.window_count += 1;
        if self.window_count == self.length_min_one {
            self.value = sample;
        } else if self.window_count == self.length {
            self.fractal_dimension = self.estimate_fractal_dimension();
            let alpha = self.estimate_alpha();
            self.value += (sample - self.value) * alpha;
            self.primed = true;

            return self.value;
        }

        f64::NAN
    }

    fn estimate_fractal_dimension(&self) -> f64 {
        let mut min_low_half = f64::MAX;
        let mut max_high_half = f64::MIN_POSITIVE; // SmallestNonzeroFloat64 equivalent

        for i in 0..self.half_length {
            let l = self.window_low[i];
            if min_low_half > l {
                min_low_half = l;
            }
            let h = self.window_high[i];
            if max_high_half < h {
                max_high_half = h;
            }
        }

        let range_n1 = max_high_half - min_low_half;
        let mut min_low_full = min_low_half;
        let mut max_high_full = max_high_half;
        min_low_half = f64::MAX;
        max_high_half = f64::MIN_POSITIVE;

        for j in 0..self.half_length {
            let i = j + self.half_length;
            let l = self.window_low[i];

            if min_low_full > l {
                min_low_full = l;
            }
            if min_low_half > l {
                min_low_half = l;
            }

            let h = self.window_high[i];
            if max_high_full < h {
                max_high_full = h;
            }
            if max_high_half < h {
                max_high_half = h;
            }
        }

        let range_n2 = max_high_half - min_low_half;
        let range_n3 = max_high_full - min_low_full;

        let fdim = (((range_n1 + range_n2) / self.half_length as f64).ln()
            - (range_n3 / self.length as f64).ln())
            * std::f64::consts::LOG2_E;

        fdim.clamp(1.0, 2.0)
    }

    fn estimate_alpha(&self) -> f64 {
        let alpha = (self.scaling_factor * (self.fractal_dimension - 1.0)).exp();
        alpha.clamp(self.alpha_slowest, 1.0)
    }

    fn update_entity(&mut self, time: i64, sample: f64, sample_high: f64, sample_low: f64) -> Output {
        let frama = self.update(sample, sample_high, sample_low);

        let fdim = if frama.is_nan() {
            f64::NAN
        } else {
            self.fractal_dimension
        };

        vec![
            Box::new(Scalar::new(time, frama)),
            Box::new(Scalar::new(time, fdim)),
        ]
    }
}

impl Indicator for FractalAdaptiveMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::FractalAdaptiveMovingAverage,
            &self.line.mnemonic,
            &self.line.description,
            &[
                OutputText {
                    mnemonic: self.line.mnemonic.clone(),
                    description: self.line.description.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_fdim.clone(),
                    description: self.description_fdim.clone(),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = sample.value;
        self.update_entity(sample.time, v, v, v)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.line.bar_func)(sample);
        self.update_entity(sample.time, v, sample.high, sample.low)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.line.quote_func)(sample);
        self.update_entity(sample.time, v, sample.ask_price, sample.bid_price)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.line.trade_func)(sample);
        self.update_entity(sample.time, v, v, v)
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

    const EPSILON: f64 = 1e-9;

    fn create_frama(length: i64, slowest: f64) -> FractalAdaptiveMovingAverage {
        FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length,
            slowest_smoothing_factor: slowest,
            ..Default::default()
        })
        .unwrap()
    }
    #[test]
    fn test_update_frama() {
        let input_mid = testdata::test_input_mid();
        let input_high = testdata::test_input_high();
        let input_low = testdata::test_input_low();
        let expected = testdata::test_expected_frama();

        let mut frama = create_frama(16, 0.01);

        for i in 0..15 {
            let v = frama.update(input_mid[i], input_high[i], input_low[i]);
            assert!(v.is_nan(), "[{}] expected NaN, got {}", i, v);
        }

        for i in 15..input_mid.len() {
            let act = frama.update(input_mid[i], input_high[i], input_low[i]);
            assert!(
                (act - expected[i]).abs() <= EPSILON,
                "[{}] expected {}, got {}",
                i,
                expected[i],
                act
            );
        }

        // NaN passthrough
        assert!(frama.update(f64::NAN, f64::NAN, f64::NAN).is_nan());
    }

    #[test]
    fn test_update_fdim() {
        let input_mid = testdata::test_input_mid();
        let input_high = testdata::test_input_high();
        let input_low = testdata::test_input_low();
        let expected_fdim = testdata::test_expected_fdim();

        let mut frama = create_frama(16, 0.01);

        for i in 0..15 {
            frama.update(input_mid[i], input_high[i], input_low[i]);
            assert!(
                frama.fractal_dimension.is_nan(),
                "[{}] expected NaN fdim, got {}",
                i,
                frama.fractal_dimension
            );
        }

        for i in 15..input_mid.len() {
            frama.update(input_mid[i], input_high[i], input_low[i]);
            let act = frama.fractal_dimension;
            assert!(
                (act - expected_fdim[i]).abs() <= EPSILON,
                "[{}] fdim expected {}, got {}",
                i,
                expected_fdim[i],
                act
            );
        }
    }

    #[test]
    fn test_is_primed() {
        let input_mid = testdata::test_input_mid();
        let input_high = testdata::test_input_high();
        let input_low = testdata::test_input_low();

        let mut frama = create_frama(16, 0.01);

        assert!(!frama.is_primed());

        for i in 0..15 {
            frama.update(input_mid[i], input_high[i], input_low[i]);
            assert!(!frama.is_primed(), "[{}] should not be primed", i);
        }

        for i in 15..input_mid.len() {
            frama.update(input_mid[i], input_high[i], input_low[i]);
            assert!(frama.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_update_entity() {
        let time = 1617235200_i64;
        let inp = 3.0_f64;
        let expected_frama_val = 2.999999999999997_f64;
        let expected_fdim_val = 1.0000000000000002_f64;

        // Scalar
        let mut frama = create_frama(16, 0.01);
        for _ in 0..15 {
            frama.update(0.0, 0.0, 0.0);
        }
        let out = frama.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 2);
        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        let s1 = out[1].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.time, time);
        assert_eq!(s0.value, expected_frama_val);
        assert_eq!(s1.time, time);
        assert_eq!(s1.value, expected_fdim_val);

        // Bar
        let mut frama = create_frama(16, 0.01);
        for _ in 0..15 {
            frama.update(0.0, 0.0, 0.0);
        }
        let bar = Bar::new(time, 0.0, inp, inp, inp, 0.0);
        let out = frama.update_bar(&bar);
        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        let s1 = out[1].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.value, expected_frama_val);
        assert_eq!(s1.value, expected_fdim_val);

        // Quote
        let mut frama = create_frama(16, 0.01);
        for _ in 0..15 {
            frama.update(0.0, 0.0, 0.0);
        }
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = frama.update_quote(&quote);
        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        let s1 = out[1].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.value, expected_frama_val);
        assert_eq!(s1.value, expected_fdim_val);

        // Trade
        let mut frama = create_frama(16, 0.01);
        for _ in 0..15 {
            frama.update(0.0, 0.0, 0.0);
        }
        let trade = Trade::new(time, inp, 0.0);
        let out = frama.update_trade(&trade);
        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        let s1 = out[1].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.value, expected_frama_val);
        assert_eq!(s1.value, expected_fdim_val);
    }

    #[test]
    fn test_metadata() {
        let frama = create_frama(16, 0.01);
        let m = frama.metadata();

        assert_eq!(m.identifier, Identifier::FractalAdaptiveMovingAverage);
        assert_eq!(m.mnemonic, "frama(16, 0.010)");
        assert_eq!(m.description, "Fractal adaptive moving average frama(16, 0.010)");
        assert_eq!(m.outputs.len(), 2);
        assert_eq!(m.outputs[0].kind, FractalAdaptiveMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "frama(16, 0.010)");
        assert_eq!(m.outputs[1].kind, FractalAdaptiveMovingAverageOutput::Fdim as i32);
        assert_eq!(m.outputs[1].shape, Shape::Scalar);
        assert_eq!(m.outputs[1].mnemonic, "framaDim(16, 0.010)");
    }

    #[test]
    fn test_metadata_with_bar_component() {
        let frama = FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length: 16,
            slowest_smoothing_factor: 0.01,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        })
        .unwrap();
        let m = frama.metadata();
        assert_eq!(m.mnemonic, "frama(16, 0.010, hl/2)");
        assert_eq!(
            m.description,
            "Fractal adaptive moving average frama(16, 0.010, hl/2)"
        );
    }

    #[test]
    fn test_new_validation() {
        // Length < 2
        assert!(FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length: 1,
            ..Default::default()
        })
        .is_err());

        assert!(FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length: 0,
            ..Default::default()
        })
        .is_err());

        // Alpha out of range
        assert!(FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length: 16,
            slowest_smoothing_factor: -0.01,
            ..Default::default()
        })
        .is_err());

        assert!(FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length: 16,
            slowest_smoothing_factor: 1.01,
            ..Default::default()
        })
        .is_err());
    }

    #[test]
    fn test_odd_length_rounded_up() {
        let frama = FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length: 17,
            slowest_smoothing_factor: 0.01,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        })
        .unwrap();
        assert_eq!(frama.length, 18);
        assert_eq!(frama.line.mnemonic, "frama(18, 0.010, hl/2)");
    }
}
