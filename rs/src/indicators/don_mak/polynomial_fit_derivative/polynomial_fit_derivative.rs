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

/// Parameters to create an instance of the Polynomial Fit Derivative indicator.
pub struct PolynomialFitDerivativeParams {
    /// Polynomial degree (number of data points is degree + 1). >= 2. Default 3.
    pub degree: usize,
    /// Derivative order (1 = velocity, 2 = acceleration). >= 1 and <= degree. Default 1.
    pub order: usize,
    /// EMA pre-smoothing length applied before the FIR filter. >= 0. Default 6.
    pub smoothing: i64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for PolynomialFitDerivativeParams {
    fn default() -> Self {
        Self {
            degree: 3,
            order: 1,
            smoothing: 6,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Polynomial Fit Derivative indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum PolynomialFitDerivativeOutput {
    /// The order-th derivative of the polynomial fit at the current bar.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Coefficient computation
// ---------------------------------------------------------------------------

/// Computes the FIR filter coefficients for the order-th derivative of a
/// degree-`degree` polynomial fit, evaluated at the most recent point.
///
/// Uses the Lagrange basis with the elementary-symmetric-polynomial identity:
///   c_i = order! * e_{degree-order}(others) / prod_{j != i} (j - i)
/// where `others` is the set of point positions {0..degree} excluding i.
fn compute_coefficients(degree: usize, order: usize) -> Vec<f64> {
    let n_points = degree + 1;

    let mut factorial_order = 1.0_f64;
    for f in 2..=order {
        factorial_order *= f as f64;
    }

    let mut coefficients: Vec<f64> = Vec::with_capacity(n_points);

    for i in 0..n_points {
        let mut denom = 1.0_f64;
        for j in 0..n_points {
            if j != i {
                denom *= (j as i64 - i as i64) as f64;
            }
        }

        // Elementary symmetric polynomials e[0..degree] of the values {0..degree} \ {i}.
        let mut e = vec![0.0_f64; degree + 1];
        e[0] = 1.0;
        for j in 0..n_points {
            if j == i {
                continue;
            }
            let v = j as f64;
            let mut k = degree;
            while k >= 1 {
                e[k] += v * e[k - 1];
                k -= 1;
            }
        }

        let numerator = factorial_order * e[degree - order];
        coefficients.push(numerator / denom);
    }

    coefficients
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Don Mak's Polynomial Fit Derivative (PFD).
///
/// Fits a polynomial of degree `degree` to the most recent `degree + 1`
/// (optionally EMA-smoothed) prices and evaluates its `order`-th derivative at
/// the current bar. This is a FIR filter: a dot product of fixed Lagrange-derived
/// coefficients with the last `degree + 1` smoothed prices.
pub struct PolynomialFitDerivative {
    line: LineIndicator,
    coefficients: Vec<f64>,
    n_points: usize,
    smoothing: i64,
    ema_alpha: f64,
    ema_value: f64,
    ema_initialized: bool,
    buf: Vec<f64>,
    buf_pos: usize,
    buf_count: usize,
    primed: bool,
}

impl PolynomialFitDerivative {
    /// Creates a new Polynomial Fit Derivative from the given parameters.
    pub fn new(params: &PolynomialFitDerivativeParams) -> Result<Self, String> {
        let invalid = "invalid polynomial fit derivative parameters";

        let degree = params.degree;
        let order = params.order;
        let smoothing = params.smoothing;

        if degree < 2 {
            return Err(format!("{}: degree should be >= 2", invalid));
        }
        if order < 1 || order > degree {
            return Err(format!("{}: order should be >= 1 and <= degree", invalid));
        }
        if smoothing < 0 {
            return Err(format!("{}: smoothing should be >= 0", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("pfd({},{},{}{})", degree, order, smoothing, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Polynomial fit derivative {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let ema_alpha = if smoothing > 0 {
            2.0 / (smoothing as f64 + 1.0)
        } else {
            0.0
        };

        Ok(Self {
            line,
            coefficients: compute_coefficients(degree, order),
            n_points: degree + 1,
            smoothing,
            ema_alpha,
            ema_value: 0.0,
            ema_initialized: false,
            buf: vec![0.0; degree + 1],
            buf_pos: 0,
            buf_count: 0,
            primed: false,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update returning the FIR filter output.
    pub fn update(&mut self, sample: f64) -> f64 {
        // Step 1: optional EMA smoothing.
        let smoothed = if self.smoothing > 0 {
            if !self.ema_initialized {
                self.ema_value = sample;
                self.ema_initialized = true;
            } else {
                self.ema_value = self.ema_alpha * sample + (1.0 - self.ema_alpha) * self.ema_value;
            }
            self.ema_value
        } else {
            sample
        };

        // Step 2: push into the ring buffer.
        self.buf[self.buf_pos] = smoothed;
        self.buf_pos = (self.buf_pos + 1) % self.n_points;
        self.buf_count += 1;

        // Step 3: not enough data yet.
        if self.buf_count < self.n_points {
            self.primed = false;
            return f64::NAN;
        }

        // Step 4: FIR dot product (coefficients[j] multiplies the j-th most recent).
        let mut result = 0.0_f64;
        let n = self.n_points as i64;
        for j in 0..self.n_points {
            let offset = self.buf_pos as i64 - 1 - j as i64;
            let buf_idx = offset.rem_euclid(n) as usize;
            result += self.coefficients[j] * self.buf[buf_idx];
        }

        self.primed = true;
        result
    }
}

impl Indicator for PolynomialFitDerivative {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::PolynomialFitDerivative,
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
    use super::*;
    use super::super::testdata::testdata;

    const TOLERANCE: f64 = 1e-9;

    struct Combo {
        degree: usize,
        order: usize,
        smoothing: i64,
        expected: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { degree: 2, order: 1, smoothing: 0, expected: testdata::expected_d2_o1_s0() },
            Combo { degree: 2, order: 1, smoothing: 3, expected: testdata::expected_d2_o1_s3() },
            Combo { degree: 2, order: 1, smoothing: 6, expected: testdata::expected_d2_o1_s6() },
            Combo { degree: 2, order: 2, smoothing: 0, expected: testdata::expected_d2_o2_s0() },
            Combo { degree: 2, order: 2, smoothing: 3, expected: testdata::expected_d2_o2_s3() },
            Combo { degree: 2, order: 2, smoothing: 6, expected: testdata::expected_d2_o2_s6() },
            Combo { degree: 3, order: 1, smoothing: 0, expected: testdata::expected_d3_o1_s0() },
            Combo { degree: 3, order: 1, smoothing: 3, expected: testdata::expected_d3_o1_s3() },
            Combo { degree: 3, order: 1, smoothing: 6, expected: testdata::expected_d3_o1_s6() },
            Combo { degree: 3, order: 2, smoothing: 0, expected: testdata::expected_d3_o2_s0() },
            Combo { degree: 3, order: 2, smoothing: 3, expected: testdata::expected_d3_o2_s3() },
            Combo { degree: 3, order: 2, smoothing: 6, expected: testdata::expected_d3_o2_s6() },
            Combo { degree: 4, order: 1, smoothing: 0, expected: testdata::expected_d4_o1_s0() },
            Combo { degree: 4, order: 1, smoothing: 3, expected: testdata::expected_d4_o1_s3() },
            Combo { degree: 4, order: 1, smoothing: 6, expected: testdata::expected_d4_o1_s6() },
            Combo { degree: 4, order: 2, smoothing: 0, expected: testdata::expected_d4_o2_s0() },
            Combo { degree: 4, order: 2, smoothing: 3, expected: testdata::expected_d4_o2_s3() },
            Combo { degree: 4, order: 2, smoothing: 6, expected: testdata::expected_d4_o2_s6() },
            Combo { degree: 5, order: 1, smoothing: 0, expected: testdata::expected_d5_o1_s0() },
            Combo { degree: 5, order: 1, smoothing: 3, expected: testdata::expected_d5_o1_s3() },
            Combo { degree: 5, order: 1, smoothing: 6, expected: testdata::expected_d5_o1_s6() },
            Combo { degree: 5, order: 2, smoothing: 0, expected: testdata::expected_d5_o2_s0() },
            Combo { degree: 5, order: 2, smoothing: 3, expected: testdata::expected_d5_o2_s3() },
            Combo { degree: 5, order: 2, smoothing: 6, expected: testdata::expected_d5_o2_s6() },
            Combo { degree: 6, order: 1, smoothing: 0, expected: testdata::expected_d6_o1_s0() },
            Combo { degree: 6, order: 1, smoothing: 3, expected: testdata::expected_d6_o1_s3() },
            Combo { degree: 6, order: 1, smoothing: 6, expected: testdata::expected_d6_o1_s6() },
            Combo { degree: 6, order: 2, smoothing: 0, expected: testdata::expected_d6_o2_s0() },
            Combo { degree: 6, order: 2, smoothing: 3, expected: testdata::expected_d6_o2_s3() },
            Combo { degree: 6, order: 2, smoothing: 6, expected: testdata::expected_d6_o2_s6() },
            Combo { degree: 4, order: 3, smoothing: 6, expected: testdata::expected_d4_o3_s6() },
            Combo { degree: 5, order: 3, smoothing: 6, expected: testdata::expected_d5_o3_s6() },
            Combo { degree: 6, order: 3, smoothing: 6, expected: testdata::expected_d6_o3_s6() },
            Combo { degree: 6, order: 5, smoothing: 6, expected: testdata::expected_d6_o5_s6() },
        ];

        let input = testdata::test_input();

        for combo in &combos {
            let mut ind = PolynomialFitDerivative::new(&PolynomialFitDerivativeParams {
                degree: combo.degree,
                order: combo.order,
                smoothing: combo.smoothing,
                ..Default::default()
            })
            .unwrap();

            for i in 0..input.len() {
                let value = ind.update(input[i]);
                let exp = combo.expected[i];
                if exp.is_nan() {
                    assert!(value.is_nan(), "d{} o{} s{} [{}]: expected NaN, got {}", combo.degree, combo.order, combo.smoothing, i, value);
                } else {
                    assert!(
                        (value - exp).abs() <= TOLERANCE,
                        "d{} o{} s{} [{}]: expected {}, got {}",
                        combo.degree, combo.order, combo.smoothing, i, exp, value
                    );
                }
            }
        }
    }

    #[test]
    fn test_mnemonic() {
        let ind = PolynomialFitDerivative::new(&PolynomialFitDerivativeParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "pfd(3,1,6)");

        let ind2 = PolynomialFitDerivative::new(&PolynomialFitDerivativeParams {
            degree: 4,
            order: 2,
            smoothing: 3,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "pfd(4,2,3)");
    }

    #[test]
    fn test_metadata() {
        let ind = PolynomialFitDerivative::new(&PolynomialFitDerivativeParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::PolynomialFitDerivative);
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, PolynomialFitDerivativeOutput::Value as i32);
    }

    #[test]
    fn test_update_scalar() {
        let input = testdata::test_input();
        let expected = testdata::expected_d3_o1_s6();

        let mut ind = PolynomialFitDerivative::new(&PolynomialFitDerivativeParams::default()).unwrap();
        let mut out: Output = vec![];
        for i in 0..input.len() {
            out = ind.update_scalar(&Scalar { time: 0, value: input[i] });
        }
        assert_eq!(out.len(), 1);
        let last = input.len() - 1;
        let value = out[0].downcast_ref::<Scalar>().unwrap().value;
        assert!((value - expected[last]).abs() <= TOLERANCE);
    }

    #[test]
    fn test_invalid_params() {
        assert!(PolynomialFitDerivative::new(&PolynomialFitDerivativeParams { degree: 1, ..Default::default() }).is_err());
        assert!(PolynomialFitDerivative::new(&PolynomialFitDerivativeParams { order: 0, ..Default::default() }).is_err());
        assert!(PolynomialFitDerivative::new(&PolynomialFitDerivativeParams { degree: 3, order: 4, ..Default::default() }).is_err());
        assert!(PolynomialFitDerivative::new(&PolynomialFitDerivativeParams { smoothing: -1, ..Default::default() }).is_err());
    }
}
