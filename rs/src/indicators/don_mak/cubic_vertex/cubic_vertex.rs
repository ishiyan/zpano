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
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Cubic Vertex indicator.
///
/// The indicator has no numeric parameters; only the input price component is configurable.
#[derive(Default)]
pub struct CubicVertexParams {
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Cubic Vertex indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CubicVertexOutput {
    /// The number of bars to the more imminent turning point.
    BarsToNearTurn = 1,
    /// The number of bars to the more distant turning point.
    BarsToFarTurn = 2,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Don Mak's Cubic Vertex (CVTX).
///
/// Predicts turning points by fitting a cubic polynomial to the 4 most recent price
/// points and computing where the two vertices (extrema) occur relative to the current
/// bar. Given four consecutive prices x(n), x(n-1), x(n-2), x(n-3) (most recent first),
/// the cubic coefficients are (Eq 7.2a-c):
///   c = (x(n) - 3*x(n-1) + 3*x(n-2) - x(n-3)) / 6
///   d = (2*x(n) - 5*x(n-1) + 4*x(n-2) - x(n-3)) / 2
///   e = (11*x(n) - 18*x(n-1) + 9*x(n-2) - 2*x(n-3)) / 6
/// The vertex locations are the roots of 3c*t^2 + 2d*t + e = 0. near = smaller |t|,
/// far = larger |t|.
pub struct CubicVertex {
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    buffer: [f64; 4],
    index: usize,
    count: usize,
    primed: bool,
}

impl CubicVertex {
    /// Creates a new Cubic Vertex from the given parameters.
    pub fn new(params: &CubicVertexParams) -> Result<Self, String> {
        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let suffix = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = if suffix.is_empty() {
            "cvtx".to_string()
        } else {
            format!("cvtx({})", &suffix[2..]) // strip leading ", "
        };

        Ok(Self {
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
            mnemonic,
            buffer: [0.0; 4],
            index: 0,
            count: 0,
            primed: false,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update returning (bars_to_near_turn, bars_to_far_turn).
    pub fn update(&mut self, sample: f64) -> (f64, f64) {
        // Store the price in the ring buffer.
        self.buffer[self.index] = sample;
        self.index = (self.index + 1) % 4;
        self.count += 1;

        if self.count < 4 {
            self.primed = false;
            return (f64::NAN, f64::NAN);
        }

        self.primed = true;

        // Extract prices: x[n] (newest), x[n-1], x[n-2], x[n-3] (oldest).
        let i = self.index as i64;
        let xn = self.buffer[(i - 1).rem_euclid(4) as usize];
        let xn1 = self.buffer[(i - 2).rem_euclid(4) as usize];
        let xn2 = self.buffer[(i - 3).rem_euclid(4) as usize];
        let xn3 = self.buffer[(i - 4).rem_euclid(4) as usize];

        // Cubic polynomial coefficients (Eq 7.2a-c).
        let c = (xn - 3.0 * xn1 + 3.0 * xn2 - xn3) / 6.0;
        let d = (2.0 * xn - 5.0 * xn1 + 4.0 * xn2 - xn3) / 2.0;
        let e = (11.0 * xn - 18.0 * xn1 + 9.0 * xn2 - 2.0 * xn3) / 6.0;

        // Case: c == 0 -- cubic term vanishes, reduces to parabola or line.
        if c == 0.0 {
            if d == 0.0 {
                return (f64::NAN, f64::NAN);
            }
            let vertex = -e / (2.0 * d);
            return (vertex, f64::NAN);
        }

        // Full cubic: solve quadratic 3c*t^2 + 2d*t + e = 0.
        let disc = d * d - 3.0 * c * e;

        if disc < 0.0 {
            return (f64::NAN, f64::NAN);
        }

        if disc == 0.0 {
            let vertex = -d / (3.0 * c);
            return (vertex, vertex);
        }

        let sqrt_disc = disc.sqrt();
        let three_c = 3.0 * c;

        let t_plus = (-d + sqrt_disc) / three_c;
        let t_minus = (-d - sqrt_disc) / three_c;

        if t_plus.abs() <= t_minus.abs() {
            (t_plus, t_minus)
        } else {
            (t_minus, t_plus)
        }
    }
}

impl Indicator for CubicVertex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Cubic vertex {}", self.mnemonic);
        build_metadata(
            Identifier::CubicVertex,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} near", self.mnemonic),
                    description: format!("{} near turn", desc),
                },
                OutputText {
                    mnemonic: format!("{} far", self.mnemonic),
                    description: format!("{} far turn", desc),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (near, far) = self.update(sample.value);
        vec![
            Box::new(Scalar { time: sample.time, value: near }),
            Box::new(Scalar { time: sample.time, value: far }),
        ]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;

    include!("testdata.rs");

    const TOLERANCE: f64 = 1e-9;

    fn run_series(inputs: &[f64]) -> (Vec<f64>, Vec<f64>) {
        let mut ind = CubicVertex::new(&CubicVertexParams::default()).unwrap();
        let mut near = Vec::with_capacity(inputs.len());
        let mut far = Vec::with_capacity(inputs.len());
        for &c in inputs {
            let (n, f) = ind.update(c);
            near.push(n);
            far.push(f);
        }
        (near, far)
    }

    fn check_series(name: &str, actual: &[f64], expected: &[f64]) {
        assert_eq!(actual.len(), expected.len(), "{}: length mismatch", name);
        for i in 0..expected.len() {
            let exp = expected[i];
            if exp.is_nan() {
                assert!(actual[i].is_nan(), "{}[{}]: expected NaN, got {}", name, i, actual[i]);
            } else {
                // Combined absolute + relative tolerance (ill-conditioned near degenerate points).
                let delta = TOLERANCE * exp.abs().max(1.0);
                assert!(
                    (actual[i] - exp).abs() <= delta,
                    "{}[{}]: expected {}, got {}",
                    name, i, exp, actual[i]
                );
            }
        }
    }

    #[test]
    fn test_reference_data() {
        let (near, far) = run_series(&test_data::input_close());
        check_series("RAW_NEAR", &near, &test_data::expected_raw_near());
        check_series("RAW_FAR", &far, &test_data::expected_raw_far());

        let (near, far) = run_series(&test_data::input_ema6());
        check_series("EMA6_NEAR", &near, &test_data::expected_ema6_near());
        check_series("EMA6_FAR", &far, &test_data::expected_ema6_far());

        let (near, far) = run_series(&test_data::input_ema20());
        check_series("EMA20_NEAR", &near, &test_data::expected_ema20_near());
        check_series("EMA20_FAR", &far, &test_data::expected_ema20_far());

        let (near, far) = run_series(&test_data::test1_input_cubic());
        check_series("TEST1_NEAR", &near, &test_data::test1_expected_near());
        check_series("TEST1_FAR", &far, &test_data::test1_expected_far());
    }

    #[test]
    fn test_mnemonic() {
        assert_eq!(
            CubicVertex::new(&CubicVertexParams::default()).unwrap().metadata().mnemonic,
            "cvtx"
        );
        assert_eq!(
            CubicVertex::new(&CubicVertexParams { bar_component: Some(BarComponent::Median), ..Default::default() })
                .unwrap().metadata().mnemonic,
            "cvtx(hl/2)"
        );
    }

    #[test]
    fn test_metadata() {
        let ind = CubicVertex::new(&CubicVertexParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::CubicVertex);
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(meta.outputs[0].kind, CubicVertexOutput::BarsToNearTurn as i32);
        assert_eq!(meta.outputs[1].kind, CubicVertexOutput::BarsToFarTurn as i32);
    }

    #[test]
    fn test_priming() {
        let mut ind = CubicVertex::new(&CubicVertexParams::default()).unwrap();
        for _ in 0..3 {
            let (n, f) = ind.update(1.0);
            assert!(n.is_nan() && f.is_nan());
            assert!(!ind.is_primed());
        }
        // Four collinear points -> c == 0 and d == 0 -> both NaN, but primed.
        let (n, f) = ind.update(1.0);
        assert!(n.is_nan() && f.is_nan());
        assert!(ind.is_primed());
    }
}
