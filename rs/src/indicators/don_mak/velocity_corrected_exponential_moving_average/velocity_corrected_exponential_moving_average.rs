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

/// Parameters to create an instance of the Velocity-Corrected Exponential Moving Average indicator.
pub struct VelocityCorrectedExponentialMovingAverageParams {
    /// EMA smoothing period. >= 2. Default 6.
    pub period: usize,
    /// Polynomial degree for the velocity estimation. >= 2. Default 3.
    pub degree: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for VelocityCorrectedExponentialMovingAverageParams {
    fn default() -> Self {
        Self {
            period: 6,
            degree: 3,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Velocity-Corrected Exponential Moving Average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum VelocityCorrectedExponentialMovingAverageOutput {
    /// The velocity-corrected EMA value.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Coefficient computation
// ---------------------------------------------------------------------------

/// Computes FIR coefficients for the first derivative of a degree-`degree`
/// polynomial fit evaluated at the most recent point (Lagrange basis, order=1).
fn compute_velocity_coefficients(degree: usize) -> Vec<f64> {
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
        for ell in 0..others.len() {
            let mut term = 1.0_f64;
            for m in 0..others.len() {
                if m != ell {
                    term *= others[m];
                }
            }
            numerator += term;
        }

        coefficients.push(numerator / denom);
    }

    coefficients
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Don Mak's Velocity-Corrected Exponential Moving Average (VCEMA).
///
/// A reduced-lag EMA that pre-corrects price by adding its polynomial velocity
/// before smoothing:
///   corrected = price + PFD(price, degree, order=1)
///   VCEMA(n)  = EMA(corrected, n)
pub struct VelocityCorrectedExponentialMovingAverage {
    line: LineIndicator,
    n_points: usize,
    coefficients: Vec<f64>,
    ema_alpha: f64,
    ema_value: f64,
    ema_initialized: bool,
    buf: Vec<f64>,
    buf_pos: usize,
    buf_count: usize,
    primed: bool,
}

impl VelocityCorrectedExponentialMovingAverage {
    /// Creates a new Velocity-Corrected Exponential Moving Average from the given parameters.
    pub fn new(params: &VelocityCorrectedExponentialMovingAverageParams) -> Result<Self, String> {
        let invalid = "invalid velocity-corrected exponential moving average parameters";

        let period = params.period;
        let degree = params.degree;

        if period < 2 {
            return Err(format!("{}: period should be >= 2", invalid));
        }
        if degree < 2 {
            return Err(format!("{}: degree should be >= 2", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("vcema({},{}{})", period, degree, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Velocity-corrected exponential moving average {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let n_points = degree + 1;

        Ok(Self {
            line,
            n_points,
            coefficients: compute_velocity_coefficients(degree),
            ema_alpha: 2.0 / (period as f64 + 1.0),
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
        // Store the raw price in the ring buffer.
        self.buf[self.buf_pos] = sample;
        self.buf_pos = (self.buf_pos + 1) % self.n_points;
        self.buf_count += 1;

        if self.buf_count < self.n_points {
            self.primed = false;
            return f64::NAN;
        }

        self.primed = true;

        // Compute the velocity from the raw prices.
        let mut velocity = 0.0;
        let n = self.n_points as i64;
        for k in 0..self.n_points {
            let idx = (self.buf_pos as i64 - 1 - k as i64).rem_euclid(n) as usize;
            velocity += self.coefficients[k] * self.buf[idx];
        }

        // Corrected price = price + velocity.
        let corrected = sample + velocity;

        // Apply the EMA to the corrected price (seed at the first corrected value).
        if !self.ema_initialized {
            self.ema_value = corrected;
            self.ema_initialized = true;
        } else {
            self.ema_value = self.ema_alpha * corrected + (1.0 - self.ema_alpha) * self.ema_value;
        }

        self.ema_value
    }
}

impl Indicator for VelocityCorrectedExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::VelocityCorrectedExponentialMovingAverage,
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

    fn check_series(name: &str, params: &VelocityCorrectedExponentialMovingAverageParams, inputs: &[f64], expected: &[f64]) {
        let mut ind = VelocityCorrectedExponentialMovingAverage::new(params).unwrap();
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
        let combos: &[(usize, usize, &[f64])] = &[
            (3, 2, testdata::EXPECTED_P3_D2),
            (3, 3, testdata::EXPECTED_P3_D3),
            (3, 4, testdata::EXPECTED_P3_D4),
            (3, 5, testdata::EXPECTED_P3_D5),
            (6, 2, testdata::EXPECTED_P6_D2),
            (6, 3, testdata::EXPECTED_P6_D3),
            (6, 4, testdata::EXPECTED_P6_D4),
            (6, 5, testdata::EXPECTED_P6_D5),
            (12, 2, testdata::EXPECTED_P12_D2),
            (12, 3, testdata::EXPECTED_P12_D3),
            (12, 4, testdata::EXPECTED_P12_D4),
            (12, 5, testdata::EXPECTED_P12_D5),
        ];

        for &(period, degree, expected) in combos {
            let name = format!("vcema({},{})", period, degree);
            check_series(
                &name,
                &VelocityCorrectedExponentialMovingAverageParams { period, degree, ..Default::default() },
                testdata::INPUT_CLOSE,
                expected,
            );
        }

        check_series(
            "TEST1",
            &VelocityCorrectedExponentialMovingAverageParams { period: 6, degree: 3, ..Default::default() },
            testdata::TEST1_INPUT_LINEAR,
            testdata::TEST1_EXPECTED_P6_D3,
        );
    }

    #[test]
    fn test_mnemonic() {
        assert_eq!(VelocityCorrectedExponentialMovingAverage::new(&VelocityCorrectedExponentialMovingAverageParams::default()).unwrap().metadata().mnemonic, "vcema(6,3)");
        assert_eq!(
            VelocityCorrectedExponentialMovingAverage::new(&VelocityCorrectedExponentialMovingAverageParams { period: 12, degree: 5, ..Default::default() }).unwrap().metadata().mnemonic,
            "vcema(12,5)"
        );
    }

    #[test]
    fn test_metadata() {
        let ind = VelocityCorrectedExponentialMovingAverage::new(&VelocityCorrectedExponentialMovingAverageParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::VelocityCorrectedExponentialMovingAverage);
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, VelocityCorrectedExponentialMovingAverageOutput::Value as i32);
    }

    #[test]
    fn test_update_scalar() {
        let mut ind = VelocityCorrectedExponentialMovingAverage::new(&VelocityCorrectedExponentialMovingAverageParams::default()).unwrap();
        let mut out: Output = vec![];
        for &c in testdata::INPUT_CLOSE {
            out = ind.update_scalar(&Scalar { time: 0, value: c });
        }
        assert_eq!(out.len(), 1);
        let last = testdata::INPUT_CLOSE.len() - 1;
        let value = out[0].downcast_ref::<Scalar>().unwrap().value;
        assert!((value - testdata::EXPECTED_P6_D3[last]).abs() <= TOLERANCE);
    }

    #[test]
    fn test_invalid_params() {
        assert!(VelocityCorrectedExponentialMovingAverage::new(&VelocityCorrectedExponentialMovingAverageParams { period: 1, ..Default::default() }).is_err());
        assert!(VelocityCorrectedExponentialMovingAverage::new(&VelocityCorrectedExponentialMovingAverageParams { degree: 1, ..Default::default() }).is_err());
    }
}
