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

/// Parameters to create an instance of the Polynomial Forecast indicator.
pub struct PolynomialForecastParams {
    /// Polynomial degree for the local fit (uses degree+1 bars). >= 2. Default 3.
    pub degree: usize,
    /// Taylor expansion order: 1 = velocity only (F1V), 2 = velocity + acceleration (F1VA). Default 1.
    pub order: usize,
    /// EMA pre-smoothing period applied to price before fitting (0 = none). >= 0. Default 0.
    pub smoothing: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for PolynomialForecastParams {
    fn default() -> Self {
        Self {
            degree: 3,
            order: 1,
            smoothing: 0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Polynomial Forecast indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum PolynomialForecastOutput {
    /// The 1-bar-ahead price forecast value.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Coefficient computation
// ---------------------------------------------------------------------------

/// Computes FIR coefficients for the `order`-th derivative of a degree-`degree`
/// polynomial fit evaluated at the most recent point (Lagrange basis).
fn compute_coefficients(degree: usize, order: usize) -> Vec<f64> {
    let n_points = degree + 1;
    let mut coefficients: Vec<f64> = Vec::with_capacity(n_points);

    for i in 0..n_points {
        let mut denom = 1.0_f64;
        for j in 0..n_points {
            if j != i {
                denom *= (j as i64 - i as i64) as f64;
            }
        }

        let others: Vec<f64> = (0..n_points).filter(|&j| j != i).map(|j| j as f64).collect();

        let mut numerator = 0.0_f64;
        if order == 1 {
            for ell in 0..others.len() {
                let mut term = 1.0_f64;
                for m in 0..others.len() {
                    if m != ell {
                        term *= others[m];
                    }
                }
                numerator += term;
            }
        } else {
            for ell in 0..others.len() {
                for r in (ell + 1)..others.len() {
                    let mut term = 2.0_f64;
                    for m in 0..others.len() {
                        if m != ell && m != r {
                            term *= others[m];
                        }
                    }
                    numerator += term;
                }
            }
        }

        coefficients.push(numerator / denom);
    }

    coefficients
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Don Mak's Polynomial Forecast (POF).
///
/// A one-step-ahead price forecast using a Taylor series expansion built on
/// polynomial fit derivatives (PFD):
///   velocity     = PFD(price, degree, order=1)
///   acceleration = PFD(price, degree, order=2)
///   order=1:  forecast = price + velocity
///   order=2:  forecast = price + velocity + 0.5*acceleration
pub struct PolynomialForecast {
    line: LineIndicator,
    order: usize,
    smoothing: usize,
    n_points: usize,
    coeff_vel: Vec<f64>,
    coeff_acc: Option<Vec<f64>>,
    ema_alpha: f64,
    ema_value: f64,
    ema_initialized: bool,
    buf: Vec<f64>,
    buf_pos: usize,
    buf_count: usize,
    primed: bool,
}

impl PolynomialForecast {
    /// Creates a new Polynomial Forecast from the given parameters.
    pub fn new(params: &PolynomialForecastParams) -> Result<Self, String> {
        let invalid = "invalid polynomial forecast parameters";

        let degree = params.degree;
        let order = params.order;
        let smoothing = params.smoothing;

        if degree < 2 {
            return Err(format!("{}: degree should be >= 2", invalid));
        }
        if order < 1 || order > 2 {
            return Err(format!("{}: order should be 1 or 2", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("pof({},{},{}{})", degree, order, smoothing, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Polynomial forecast {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let n_points = degree + 1;
        let ema_alpha = if smoothing > 0 { 2.0 / (smoothing as f64 + 1.0) } else { 0.0 };
        let coeff_acc = if order == 2 { Some(compute_coefficients(degree, 2)) } else { None };

        Ok(Self {
            line,
            order,
            smoothing,
            n_points,
            coeff_vel: compute_coefficients(degree, 1),
            coeff_acc,
            ema_alpha,
            ema_value: 0.0,
            ema_initialized: false,
            buf: vec![0.0; n_points],
            buf_pos: 0,
            buf_count: 0,
            primed: false,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update returning the filter output.
    pub fn update(&mut self, sample: f64) -> f64 {
        // Optional EMA pre-smoothing.
        let mut smoothed = sample;
        if self.smoothing > 0 {
            if !self.ema_initialized {
                self.ema_value = sample;
                self.ema_initialized = true;
            } else {
                self.ema_value = self.ema_alpha * sample + (1.0 - self.ema_alpha) * self.ema_value;
            }
            smoothed = self.ema_value;
        }

        // Store the smoothed price in the ring buffer.
        self.buf[self.buf_pos] = smoothed;
        self.buf_pos = (self.buf_pos + 1) % self.n_points;
        self.buf_count += 1;

        if self.buf_count < self.n_points {
            self.primed = false;
            return f64::NAN;
        }

        self.primed = true;

        // Read buffer most-recent-first and compute velocity (and acceleration).
        let mut velocity = 0.0;
        let mut acceleration = 0.0;
        let n = self.n_points as i64;
        for k in 0..self.n_points {
            let idx = (self.buf_pos as i64 - 1 - k as i64).rem_euclid(n) as usize;
            let value = self.buf[idx];
            velocity += self.coeff_vel[k] * value;
            if let Some(ref acc) = self.coeff_acc {
                acceleration += acc[k] * value;
            }
        }

        let mut forecast = smoothed + velocity;
        if self.order == 2 {
            forecast += 0.5 * acceleration;
        }

        forecast
    }
}

impl Indicator for PolynomialForecast {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::PolynomialForecast,
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

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::super::testdata;
    use super::*;

    const TOLERANCE: f64 = 1e-9;

    fn check_series(name: &str, params: &PolynomialForecastParams, inputs: &[f64], expected: &[f64]) {
        let mut ind = PolynomialForecast::new(params).unwrap();
        assert_eq!(inputs.len(), expected.len(), "{}: length mismatch", name);
        for i in 0..inputs.len() {
            let value = ind.update(inputs[i]);
            let exp = expected[i];
            if exp.is_nan() {
                assert!(value.is_nan(), "{}[{}]: expected NaN, got {}", name, i, value);
            } else {
                assert!(
                    (value - exp).abs() <= TOLERANCE,
                    "{}[{}]: expected {}, got {}",
                    name, i, exp, value
                );
            }
        }
    }

    #[test]
    fn test_reference_data() {
        let combos: &[(usize, usize, usize, &[f64])] = &[
            (2, 1, 0, testdata::EXPECTED_D2_O1_S0),
            (2, 1, 3, testdata::EXPECTED_D2_O1_S3),
            (2, 1, 6, testdata::EXPECTED_D2_O1_S6),
            (2, 2, 0, testdata::EXPECTED_D2_O2_S0),
            (2, 2, 3, testdata::EXPECTED_D2_O2_S3),
            (2, 2, 6, testdata::EXPECTED_D2_O2_S6),
            (3, 1, 0, testdata::EXPECTED_D3_O1_S0),
            (3, 1, 3, testdata::EXPECTED_D3_O1_S3),
            (3, 1, 6, testdata::EXPECTED_D3_O1_S6),
            (3, 2, 0, testdata::EXPECTED_D3_O2_S0),
            (3, 2, 3, testdata::EXPECTED_D3_O2_S3),
            (3, 2, 6, testdata::EXPECTED_D3_O2_S6),
            (4, 1, 0, testdata::EXPECTED_D4_O1_S0),
            (4, 1, 3, testdata::EXPECTED_D4_O1_S3),
            (4, 1, 6, testdata::EXPECTED_D4_O1_S6),
            (4, 2, 0, testdata::EXPECTED_D4_O2_S0),
            (4, 2, 3, testdata::EXPECTED_D4_O2_S3),
            (4, 2, 6, testdata::EXPECTED_D4_O2_S6),
            (5, 1, 0, testdata::EXPECTED_D5_O1_S0),
            (5, 1, 3, testdata::EXPECTED_D5_O1_S3),
            (5, 1, 6, testdata::EXPECTED_D5_O1_S6),
            (5, 2, 0, testdata::EXPECTED_D5_O2_S0),
            (5, 2, 3, testdata::EXPECTED_D5_O2_S3),
            (5, 2, 6, testdata::EXPECTED_D5_O2_S6),
        ];

        for &(degree, order, smoothing, expected) in combos {
            let name = format!("pof({},{},{})", degree, order, smoothing);
            check_series(
                &name,
                &PolynomialForecastParams { degree, order, smoothing, ..Default::default() },
                testdata::INPUT_CLOSE,
                expected,
            );
        }

        check_series(
            "TEST1_O1",
            &PolynomialForecastParams { degree: 3, order: 1, smoothing: 0, ..Default::default() },
            testdata::TEST1_INPUT_LINEAR,
            testdata::TEST1_EXPECTED_D3_O1_S0,
        );
        check_series(
            "TEST1_O2",
            &PolynomialForecastParams { degree: 3, order: 2, smoothing: 0, ..Default::default() },
            testdata::TEST1_INPUT_LINEAR,
            testdata::TEST1_EXPECTED_D3_O2_S0,
        );
    }

    #[test]
    fn test_mnemonic() {
        assert_eq!(PolynomialForecast::new(&PolynomialForecastParams::default()).unwrap().metadata().mnemonic, "pof(3,1,0)");
        assert_eq!(
            PolynomialForecast::new(&PolynomialForecastParams { degree: 5, order: 2, smoothing: 6, ..Default::default() }).unwrap().metadata().mnemonic,
            "pof(5,2,6)"
        );
    }

    #[test]
    fn test_metadata() {
        let ind = PolynomialForecast::new(&PolynomialForecastParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::PolynomialForecast);
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, PolynomialForecastOutput::Value as i32);
    }

    #[test]
    fn test_update_scalar() {
        let mut ind = PolynomialForecast::new(&PolynomialForecastParams::default()).unwrap();
        let mut out: Output = vec![];
        for &c in testdata::INPUT_CLOSE {
            out = ind.update_scalar(&Scalar { time: 0, value: c });
        }
        assert_eq!(out.len(), 1);
        let last = testdata::INPUT_CLOSE.len() - 1;
        let value = out[0].downcast_ref::<Scalar>().unwrap().value;
        assert!((value - testdata::EXPECTED_D3_O1_S0[last]).abs() <= TOLERANCE);
    }

    #[test]
    fn test_invalid_params() {
        assert!(PolynomialForecast::new(&PolynomialForecastParams { degree: 1, ..Default::default() }).is_err());
        assert!(PolynomialForecast::new(&PolynomialForecastParams { order: 0, ..Default::default() }).is_err());
        assert!(PolynomialForecast::new(&PolynomialForecastParams { order: 3, ..Default::default() }).is_err());
    }
}
