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

/// Parameters to create an instance of the linear regression indicator.
pub struct LinearRegressionParams {
    /// The lookback period (number of bars used for the regression).
    /// Must be greater than 1.
    pub length: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for LinearRegressionParams {
    fn default() -> Self {
        Self {
            length: 20,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the linear regression indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum LinearRegressionOutput {
    /// The linear regression value: b + m*(period-1).
    Value = 1,
    /// The time series forecast: b + m*period.
    Forecast = 2,
    /// The y-intercept of the regression line: b.
    Intercept = 3,
    /// The slope of the regression line: m.
    SlopeRad = 4,
    /// The slope in degrees: atan(m) * 180/pi.
    SlopeDeg = 5,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the least-squares regression line over a rolling window
/// and produces five outputs per sample:
///
/// - Value:     b + m*(period-1)  -- the regression value at the last bar
/// - Forecast:  b + m*period      -- the time series forecast (one bar ahead)
/// - Intercept: b                 -- the y-intercept of the regression line
/// - SlopeRad:  m                 -- the slope of the regression line
/// - SlopeDeg:  atan(m)*180/pi   -- the slope expressed in degrees
///
/// The indicator is not primed during the first (period-1) updates.
pub struct LinearRegression {
    line: LineIndicator,
    length: usize,
    length_f: f64,
    sum_x: f64,
    divisor: f64,
    window: Vec<f64>,
    window_count: usize,
    primed: bool,
    // Current output values.
    cur_value: f64,
    cur_forecast: f64,
    cur_intercept: f64,
    cur_slope_rad: f64,
    cur_slope_deg: f64,
}

impl LinearRegression {
    /// Creates a new LinearRegression from the given parameters.
    pub fn new(params: &LinearRegressionParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid linear regression parameters: length should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("linreg({}{})", params.length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Linear Regression {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let n = params.length as f64;
        let sum_x = n * (n - 1.0) * 0.5;
        let sum_x_sqr = n * (n - 1.0) * (2.0 * n - 1.0) / 6.0;
        let divisor = sum_x * sum_x - n * sum_x_sqr;

        Ok(Self {
            line,
            length: params.length,
            length_f: n,
            sum_x,
            divisor,
            window: vec![0.0; params.length],
            window_count: 0,
            primed: false,
            cur_value: 0.0,
            cur_forecast: 0.0,
            cur_intercept: 0.0,
            cur_slope_rad: 0.0,
            cur_slope_deg: 0.0,
        })
    }

    /// Core update logic. Returns the Value output or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        if self.primed {
            self.calculate(sample);
            return self.cur_value;
        }

        self.window[self.window_count] = sample;
        self.window_count += 1;

        if self.window_count == self.length {
            self.primed = true;
            self.compute_from_window();
            return self.cur_value;
        }

        f64::NAN
    }

    fn calculate(&mut self, sample: f64) {
        for i in 0..self.length - 1 {
            self.window[i] = self.window[i + 1];
        }
        self.window[self.length - 1] = sample;
        self.compute_from_window();
    }

    fn compute_from_window(&mut self) {
        const RAD_TO_DEG: f64 = 180.0 / std::f64::consts::PI;

        let mut sum_xy: f64 = 0.0;
        let mut sum_y: f64 = 0.0;

        for i in (1..=self.length).rev() {
            let v = self.window[self.length - i];
            sum_y += v;
            sum_xy += (i - 1) as f64 * v;
        }

        let m = (self.length_f * sum_xy - self.sum_x * sum_y) / self.divisor;
        let b = (sum_y - m * self.sum_x) / self.length_f;

        self.cur_slope_rad = m;
        self.cur_slope_deg = m.atan() * RAD_TO_DEG;
        self.cur_intercept = b;
        self.cur_value = b + m * (self.length_f - 1.0);
        self.cur_forecast = b + m * self.length_f;
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let value = self.update(sample);

        if value.is_nan() {
            let nan = f64::NAN;
            return vec![
                Box::new(Scalar::new(time, nan)),
                Box::new(Scalar::new(time, nan)),
                Box::new(Scalar::new(time, nan)),
                Box::new(Scalar::new(time, nan)),
                Box::new(Scalar::new(time, nan)),
            ];
        }

        vec![
            Box::new(Scalar::new(time, self.cur_value)),
            Box::new(Scalar::new(time, self.cur_forecast)),
            Box::new(Scalar::new(time, self.cur_intercept)),
            Box::new(Scalar::new(time, self.cur_slope_rad)),
            Box::new(Scalar::new(time, self.cur_slope_deg)),
        ]
    }
}

impl Indicator for LinearRegression {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::LinearRegression,
            &self.line.mnemonic,
            &self.line.description,
            &[
                OutputText {
                    mnemonic: self.line.mnemonic.clone(),
                    description: format!("{} value", self.line.description),
                },
                OutputText {
                    mnemonic: self.line.mnemonic.clone(),
                    description: format!("{} forecast", self.line.description),
                },
                OutputText {
                    mnemonic: self.line.mnemonic.clone(),
                    description: format!("{} intercept", self.line.description),
                },
                OutputText {
                    mnemonic: self.line.mnemonic.clone(),
                    description: format!("{} slope", self.line.description),
                },
                OutputText {
                    mnemonic: self.line.mnemonic.clone(),
                    description: format!("{} angle", self.line.description),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let sample_value = (self.line.bar_func)(sample);
        self.update_entity(sample.time, sample_value)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let sample_value = (self.line.quote_func)(sample);
        self.update_entity(sample.time, sample_value)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let sample_value = (self.line.trade_func)(sample);
        self.update_entity(sample.time, sample_value)
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    use crate::entities::bar_component::BarComponent;
    use crate::indicators::core::outputs::shape::Shape;

    const TOLERANCE: f64 = 1e-4;
    fn create_linreg(length: usize) -> LinearRegression {
        LinearRegression::new(&LinearRegressionParams { length, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_update_value_period_14() {
        let mut lr = create_linreg(14);
        let input = testdata::test_input();
        let expected = testdata::expected_value();

        for i in 0..13 {
            assert!(lr.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 13..input.len() {
            let act = lr.update(input[i]);
            assert!((expected[i] - act).abs() < TOLERANCE, "[{}] expected {}, got {}", i, expected[i], act);
        }

        assert!(lr.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_all_outputs_period_14() {
        let mut lr = create_linreg(14);
        let input = testdata::test_input();
        let exp_val = testdata::expected_value();
        let exp_fcast = testdata::expected_forecast();
        let exp_intcpt = testdata::expected_intercept();
        let exp_slope = testdata::expected_slope_rad();
        let exp_deg = testdata::expected_slope_deg();
        let time: i64 = 1617235200;

        // Feed first 12 via update.
        for i in 0..12 {
            lr.update(input[i]);
        }

        // Feed index 12 via update_scalar - should be NaN.
        let out = lr.update_scalar(&Scalar::new(time, input[12]));
        assert_eq!(out.len(), 5);
        for j in 0..5 {
            let s = out[j].downcast_ref::<Scalar>().unwrap();
            assert!(s.value.is_nan(), "output[{}] expected NaN", j);
        }

        // Feed indices 13-251 via update_scalar and verify all 5 outputs.
        for i in 13..input.len() {
            let out = lr.update_scalar(&Scalar::new(time, input[i]));
            assert_eq!(out.len(), 5);
            let s0 = out[0].downcast_ref::<Scalar>().unwrap();
            let s1 = out[1].downcast_ref::<Scalar>().unwrap();
            let s2 = out[2].downcast_ref::<Scalar>().unwrap();
            let s3 = out[3].downcast_ref::<Scalar>().unwrap();
            let s4 = out[4].downcast_ref::<Scalar>().unwrap();

            assert!((exp_val[i] - s0.value).abs() < TOLERANCE, "[{}] Value: expected {}, got {}", i, exp_val[i], s0.value);
            assert!((exp_fcast[i] - s1.value).abs() < TOLERANCE, "[{}] Forecast: expected {}, got {}", i, exp_fcast[i], s1.value);
            assert!((exp_intcpt[i] - s2.value).abs() < TOLERANCE, "[{}] Intercept: expected {}, got {}", i, exp_intcpt[i], s2.value);
            assert!((exp_slope[i] - s3.value).abs() < TOLERANCE, "[{}] SlopeRad: expected {}, got {}", i, exp_slope[i], s3.value);
            assert!((exp_deg[i] - s4.value).abs() < TOLERANCE, "[{}] SlopeDeg: expected {}, got {}", i, exp_deg[i], s4.value);
        }
    }

    #[test]
    fn test_update_entity() {
        let input = testdata::test_input();
        let time: i64 = 1617235200;

        let setup = || -> LinearRegression {
            let mut lr = create_linreg(14);
            for i in 0..14 {
                lr.update(input[i]);
            }
            lr
        };

        // bar
        let mut lr = setup();
        let bar = Bar::new(time, 0.0, 0.0, 0.0, input[14], 0.0);
        let out = lr.update_bar(&bar);
        assert_eq!(out.len(), 5);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);

        // quote
        let mut lr = setup();
        let quote = Quote::new(time, input[14], input[14], 0.0, 0.0);
        let out = lr.update_quote(&quote);
        assert_eq!(out.len(), 5);

        // trade
        let mut lr = setup();
        let trade = Trade::new(time, input[14], 0.0);
        let out = lr.update_trade(&trade);
        assert_eq!(out.len(), 5);
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();

        let mut lr = create_linreg(14);
        assert!(!lr.is_primed());
        for i in 0..13 {
            lr.update(input[i]);
            assert!(!lr.is_primed(), "[{}] should not be primed", i);
        }
        for i in 13..input.len() {
            lr.update(input[i]);
            assert!(lr.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_metadata() {
        let lr = create_linreg(14);
        let m = lr.metadata();
        assert_eq!(m.identifier, Identifier::LinearRegression);
        assert_eq!(m.mnemonic, "linreg(14)");
        assert_eq!(m.description, "Linear Regression linreg(14)");
        assert_eq!(m.outputs.len(), 5);
        assert_eq!(m.outputs[0].kind, LinearRegressionOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[1].kind, LinearRegressionOutput::Forecast as i32);
        assert_eq!(m.outputs[2].kind, LinearRegressionOutput::Intercept as i32);
        assert_eq!(m.outputs[3].kind, LinearRegressionOutput::SlopeRad as i32);
        assert_eq!(m.outputs[4].kind, LinearRegressionOutput::SlopeDeg as i32);
    }

    #[test]
    fn test_new_invalid_length() {
        let r = LinearRegression::new(&LinearRegressionParams { length: 1, ..Default::default() });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid linear regression parameters: length should be greater than 1");

        let r = LinearRegression::new(&LinearRegressionParams { length: 0, ..Default::default() });
        assert!(r.is_err());
    }

    #[test]
    fn test_mnemonic_components() {
        let lr = create_linreg(14);
        assert_eq!(lr.line.mnemonic, "linreg(14)");

        let lr = LinearRegression::new(&LinearRegressionParams {
            length: 14, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(lr.line.mnemonic, "linreg(14, hl/2)");
    }
}
