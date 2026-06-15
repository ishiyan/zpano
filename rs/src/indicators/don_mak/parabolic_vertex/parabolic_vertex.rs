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

/// Parameters to create an instance of the Parabolic Vertex indicator.
///
/// The indicator has no numeric parameters; only the input price component is configurable.
#[derive(Default)]
pub struct ParabolicVertexParams {
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

/// Enumerates the outputs of the Parabolic Vertex indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum ParabolicVertexOutput {
    /// The bars-to-near-turn value.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Don Mak's Parabolic Vertex (PVTX).
///
/// Predicts turning points by fitting a parabola to the 3 most recent price points
/// and computing where the vertex (extremum) occurs relative to the current bar:
///   t_v = -(1.5*x(n) - 2*x(n-1) + 0.5*x(n-2)) / (x(n) - 2*x(n-1) + x(n-2))
/// The output is the number of bars from the current bar to the predicted turning point.
pub struct ParabolicVertex {
    line: LineIndicator,
    buffer: [f64; 3],
    index: usize,
    count: usize,
    primed: bool,
}

impl ParabolicVertex {
    /// Creates a new Parabolic Vertex from the given parameters.
    pub fn new(params: &ParabolicVertexParams) -> Result<Self, String> {
        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let suffix = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = if suffix.is_empty() {
            "pvtx".to_string()
        } else {
            format!("pvtx({})", &suffix[2..]) // strip leading ", "
        };
        let description = format!("Parabolic vertex {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            buffer: [0.0; 3],
            index: 0,
            count: 0,
            primed: false,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update returning the filter output.
    pub fn update(&mut self, sample: f64) -> f64 {
        // Store the price in the ring buffer.
        self.buffer[self.index] = sample;
        self.index = (self.index + 1) % 3;
        self.count += 1;

        if self.count < 3 {
            self.primed = false;
            return f64::NAN;
        }

        self.primed = true;

        // Extract prices: x[n] (newest), x[n-1], x[n-2] (oldest).
        let i = self.index as i64;
        let xn = self.buffer[(i - 1).rem_euclid(3) as usize];
        let xn1 = self.buffer[(i - 2).rem_euclid(3) as usize];
        let xn2 = self.buffer[(i - 3).rem_euclid(3) as usize];

        // Denominator = second-order finite difference (proportional to curvature).
        let denom = xn - 2.0 * xn1 + xn2;
        if denom == 0.0 {
            return f64::NAN;
        }

        let numer = 1.5 * xn - 2.0 * xn1 + 0.5 * xn2;

        -numer / denom
    }
}

impl Indicator for ParabolicVertex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::ParabolicVertex,
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

    include!("testdata.rs");

    const TOLERANCE: f64 = 1e-9;

    fn check_series(name: &str, inputs: &[f64], expected: &[f64]) {
        let mut ind = ParabolicVertex::new(&ParabolicVertexParams::default()).unwrap();
        assert_eq!(inputs.len(), expected.len(), "{}: length mismatch", name);
        for i in 0..inputs.len() {
            let value = ind.update(inputs[i]);
            let exp = expected[i];
            if exp.is_nan() {
                assert!(value.is_nan(), "{}[{}]: expected NaN, got {}", name, i, value);
            } else {
                // Combined absolute + relative tolerance (ill-conditioned near collinear points).
                let delta = TOLERANCE * exp.abs().max(1.0);
                assert!(
                    (value - exp).abs() <= delta,
                    "{}[{}]: expected {}, got {}",
                    name, i, exp, value
                );
            }
        }
    }

    #[test]
    fn test_reference_data() {
        check_series("RAW", &test_data::input_close(), &test_data::expected_raw());
        check_series("EMA6", &test_data::input_ema6(), &test_data::expected_ema6());
        check_series("EMA20", &test_data::input_ema20(), &test_data::expected_ema20());
        check_series("TEST1", &test_data::test1_input_parabola(), &test_data::test1_expected());
    }

    #[test]
    fn test_mnemonic() {
        assert_eq!(
            ParabolicVertex::new(&ParabolicVertexParams::default()).unwrap().metadata().mnemonic,
            "pvtx"
        );
        assert_eq!(
            ParabolicVertex::new(&ParabolicVertexParams { bar_component: Some(BarComponent::Median), ..Default::default() })
                .unwrap().metadata().mnemonic,
            "pvtx(hl/2)"
        );
    }

    #[test]
    fn test_metadata() {
        let ind = ParabolicVertex::new(&ParabolicVertexParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::ParabolicVertex);
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, ParabolicVertexOutput::Value as i32);
    }

    #[test]
    fn test_priming() {
        let mut ind = ParabolicVertex::new(&ParabolicVertexParams::default()).unwrap();
        assert!(ind.update(1.0).is_nan());
        assert!(!ind.is_primed());
        assert!(ind.update(2.0).is_nan());
        assert!(!ind.is_primed());
        // Three collinear points -> zero curvature -> NaN, but primed.
        assert!(ind.update(3.0).is_nan());
        assert!(ind.is_primed());
    }
}
