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

/// Parameters to create an instance of the Double-Smoothed Momenta indicator.
///
/// The field names `a`, `y` and `z` are the canonical symbols from William Blau's
/// double-smoothed momentum family.
pub struct DoubleSmoothedMomentaParams {
    /// The highest/lowest close look-back. Must be > 0. Default 2.
    /// `a=2` gives the one-bar momentum of the RSI family; `a>2` gives a
    /// double-smoothed stochastic of the close.
    pub a: usize,
    /// The period of the inner (1st) smoothing EMA. Must be > 0. Default 2.
    /// `y=1` makes the inner stage a passthrough, so `DM(2,1,z)` is the EMA-form `RSI(z)`.
    pub y: usize,
    /// The period of the outer (2nd) smoothing EMA. Must be > 0. Default 14.
    pub z: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for DoubleSmoothedMomentaParams {
    fn default() -> Self {
        Self {
            a: 2,
            y: 2,
            z: 14,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Double-Smoothed Momenta indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DoubleSmoothedMomentaOutput {
    /// The Double-Smoothed Momenta oscillator value (range [0, 100]).
    Value = 1,
}

// ---------------------------------------------------------------------------
// Inlined Blau EMA
// ---------------------------------------------------------------------------

/// Stateful streaming EMA: alpha = 2/(period+1), seeds e0 = x0.
///
/// Inlined verbatim from the Blau exponential moving average so the indicator is a
/// standalone porting unit. Do NOT change its numerics.
///
/// period == 1 -> alpha == 1 -> pure passthrough (output == input).
struct Ema {
    alpha: f64,
    prev: f64,
    primed: bool,
}

impl Ema {
    fn new(period: usize) -> Self {
        Self {
            alpha: 2.0 / (period as f64 + 1.0),
            prev: 0.0,
            primed: false,
        }
    }

    fn update(&mut self, x: f64) -> f64 {
        if !self.primed {
            self.prev = x;
            self.primed = true;
            return self.prev;
        }
        self.prev = self.alpha * x + (1.0 - self.alpha) * self.prev;
        self.prev
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// William Blau's Double-Smoothed Momenta (DM), also known as the
/// Double-Smoothed RSI (DRSI) when the look-back is fixed at 2.
///
/// A close-based, double-smoothed momentum oscillator bounded to [0, 100]:
///
///   LCa_k = min(Close over the last a bars)          (lowest close)
///   HCa_k = max(Close over the last a bars)          (highest close)
///   st_k  = Close_k - LCa_k                          (close above the low close)
///   rng_k = HCa_k - LCa_k                            (a-bar close range)
///
///   DM(a, y, z) = 100 * EMA(EMA(st, y), z) / EMA(EMA(rng, y), z)
///
/// Each of the numerator (st) and denominator (rng) series is double-smoothed by an
/// inner EMA of period y then an outer EMA of period z (Blau's Ez(Ey(.)) ), and the
/// ratio is scaled by 100.
///
/// This is structurally the Double-Smoothed Stochastic computed on the CLOSE -- it
/// uses the highest/lowest close over a bars instead of the high/low of the bar --
/// and it has no signal line, so it produces a single scalar output per bar.
///
/// Named instances:
///   - RSI equivalence:     DM(2, 1, z) == RSI(z), the EMA-form RSI;
///   - Double-smoothed RSI: DRSI(y, z) = DM(2, y, z).
///
/// Priming: st/rng are valid once a closes exist (bar a-1); all four EMA stages seed
/// there. DM is NaN for bars 0..a-2 and finite from bar a-1. For a == 1 there is no
/// NaN warm-up, but the a-bar close range is then always 0, so DM is 0.0 on every
/// bar via the guard (a degenerate setting).
///
/// Division guard: EMA(EMA(rng)) <= 0 -> DM = 0.0.
pub struct DoubleSmoothedMomenta {
    line: LineIndicator,

    // Rolling window of the last a closes (for the highest/lowest close).
    window: Vec<f64>,
    window_length: usize,
    window_count: usize,
    last_index: usize,

    // Two independent 2-stage EMA cascades (double smoothing), each wired
    // inner(y) -> outer(z): EMA(EMA(x, y), z).
    numerator_y: Ema,
    numerator_z: Ema,
    denominator_y: Ema,
    denominator_z: Ema,

    primed: bool,
}

impl DoubleSmoothedMomenta {
    /// Creates a new Double-Smoothed Momenta from the given parameters.
    pub fn new(params: &DoubleSmoothedMomentaParams) -> Result<Self, String> {
        let invalid = "invalid double smoothed momenta parameters";

        let a = params.a;
        let y = params.y;
        let z = params.z;

        if a < 1 {
            return Err(format!("{}: a should be greater than 0", invalid));
        }
        if y < 1 {
            return Err(format!("{}: y should be greater than 0", invalid));
        }
        if z < 1 {
            return Err(format!("{}: z should be greater than 0", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("dm({},{},{}{})", a, y, z, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Double-Smoothed Momenta {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            window: vec![0.0; a],
            window_length: a,
            window_count: 0,
            last_index: a - 1,
            numerator_y: Ema::new(y),
            numerator_z: Ema::new(z),
            denominator_y: Ema::new(y),
            denominator_z: Ema::new(z),
            primed: false,
        })
    }

    /// Core update logic. Returns the DM value or NaN during the a-bar warm-up.
    pub fn update(&mut self, sample: f64) -> f64 {
        if self.primed {
            for i in 0..self.last_index {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = sample;
        } else {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            // Need a closes before the highest/lowest close is defined. While
            // unprimed, the EMA cascades must NOT advance (they seed at bar a-1).
            if self.window_length > self.window_count {
                return f64::NAN;
            }

            self.primed = true;
        }

        // Highest/lowest close over the last a bars.
        let mut hc = self.window[0];
        let mut lc = self.window[0];

        for i in 1..self.window_length {
            let v = self.window[i];

            if v > hc {
                hc = v;
            }

            if v < lc {
                lc = v;
            }
        }

        // Raw close-above-low (>= 0) and a-bar close range (>= 0).
        let st = sample - lc;
        let rng = hc - lc;

        // Double-smooth each separately (inner y, then outer z), then divide.
        let num = self.numerator_z.update(self.numerator_y.update(st));
        let den = self.denominator_z.update(self.denominator_y.update(rng));

        // Division guard: smoothed range <= 0 -> DM = 0.0.
        if den <= 0.0 {
            return 0.0;
        }

        100.0 * num / den
    }
}

impl Indicator for DoubleSmoothedMomenta {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::DoubleSmoothedMomenta,
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
    use crate::indicators::core::outputs::shape::Shape;

    const TOLERANCE: f64 = 1e-9;

    struct Combo {
        a: usize,
        y: usize,
        z: usize,
        expected: Vec<f64>,
    }

    fn check(name: &str, i: usize, exp: f64, act: f64) {
        if exp.is_nan() {
            assert!(act.is_nan(), "{}[{}]: expected NaN, got {}", name, i, act);
            return;
        }

        assert!(
            (act - exp).abs() <= TOLERANCE,
            "{}[{}]: expected {}, got {}",
            name,
            i,
            exp,
            act
        );
    }

    fn create(a: usize, y: usize, z: usize) -> DoubleSmoothedMomenta {
        DoubleSmoothedMomenta::new(&DoubleSmoothedMomentaParams {
            a,
            y,
            z,
            ..Default::default()
        })
        .unwrap()
    }

    /// Independently-coded EMA-form RSI: 100 * EMA(up, z) / EMA(up + dn, z).
    fn ema_form_rsi(closes: &[f64], z: usize) -> Vec<f64> {
        let alpha = 2.0 / (z as f64 + 1.0);
        let mut result = vec![f64::NAN];

        let mut numerator = 0.0;
        let mut denominator = 0.0;
        let mut primed = false;

        for k in 1..closes.len() {
            let diff = closes[k] - closes[k - 1];
            let up = if diff > 0.0 { diff } else { 0.0 };
            let dn = if diff < 0.0 { -diff } else { 0.0 };

            if primed {
                numerator = alpha * up + (1.0 - alpha) * numerator;
                denominator = alpha * (up + dn) + (1.0 - alpha) * denominator;
            } else {
                numerator = up;
                denominator = up + dn;
                primed = true;
            }

            result.push(if denominator <= 0.0 { 0.0 } else { 100.0 * numerator / denominator });
        }

        result
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { a: 2, y: 2, z: 14, expected: testdata::expected_a2_y2_z14() },
            Combo { a: 2, y: 1, z: 14, expected: testdata::expected_a2_y1_z14() },
            Combo { a: 2, y: 1, z: 9, expected: testdata::expected_a2_y1_z9() },
            Combo { a: 2, y: 1, z: 2, expected: testdata::expected_a2_y1_z2() },
            Combo { a: 2, y: 3, z: 9, expected: testdata::expected_a2_y3_z9() },
            Combo { a: 2, y: 5, z: 5, expected: testdata::expected_a2_y5_z5() },
            Combo { a: 2, y: 2, z: 5, expected: testdata::expected_a2_y2_z5() },
            Combo { a: 2, y: 1, z: 1, expected: testdata::expected_a2_y1_z1() },
            Combo { a: 1, y: 1, z: 1, expected: testdata::expected_a1_y1_z1() },
            Combo { a: 5, y: 2, z: 14, expected: testdata::expected_a5_y2_z14() },
            Combo { a: 10, y: 3, z: 5, expected: testdata::expected_a10_y3_z5() },
            Combo { a: 14, y: 2, z: 9, expected: testdata::expected_a14_y2_z9() },
            Combo { a: 20, y: 5, z: 3, expected: testdata::expected_a20_y5_z3() },
            Combo { a: 3, y: 3, z: 3, expected: testdata::expected_a3_y3_z3() },
            Combo { a: 7, y: 4, z: 2, expected: testdata::expected_a7_y4_z2() },
            Combo { a: 32, y: 2, z: 7, expected: testdata::expected_a32_y2_z7() },
        ];

        let input = testdata::testdata::test_input();

        for combo in &combos {
            let mut ind = create(combo.a, combo.y, combo.z);

            for i in 0..input.len() {
                check("dm", i, combo.expected[i], ind.update(input[i]));
            }
        }
    }

    #[test]
    fn test_warm_up_region() {
        let input = testdata::testdata::test_input();
        let a = 5;
        let mut ind = create(a, 2, 14);

        for i in 0..a - 1 {
            assert!(ind.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in a - 1..input.len() {
            assert!(!ind.update(input[i]).is_nan(), "[{}] unexpected NaN", i);
        }
    }

    #[test]
    fn test_no_warm_up_when_a_is_one() {
        let input = testdata::testdata::test_input();
        let mut ind = create(1, 1, 1);

        for i in 0..input.len() {
            assert!(!ind.update(input[i]).is_nan(), "[{}] unexpected NaN", i);
        }
    }

    #[test]
    fn test_bounded_to_zero_hundred() {
        let input = testdata::testdata::test_input();
        let mut ind = DoubleSmoothedMomenta::new(&DoubleSmoothedMomentaParams::default()).unwrap();

        for i in 0..input.len() {
            let value = ind.update(input[i]);

            if !value.is_nan() {
                assert!((0.0..=100.0).contains(&value), "[{}] out of range: {}", i, value);
            }
        }
    }

    #[test]
    fn test_degenerate_a_one_is_identically_zero() {
        let input = testdata::testdata::test_input();
        let mut ind = create(1, 1, 1);

        for i in 0..input.len() {
            assert_eq!(ind.update(input[i]), 0.0, "[{}] must be 0 when a=1", i);
        }
    }

    #[test]
    fn test_rsi_equivalence() {
        let input = testdata::testdata::test_input();

        for z in [1usize, 2, 9, 14] {
            let expected = ema_form_rsi(&input, z);
            let mut ind = create(2, 1, z);

            for i in 0..input.len() {
                check("rsi", i, expected[i], ind.update(input[i]));
            }
        }
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::testdata::test_input();
        let a = 5;
        let mut ind = create(a, 2, 14);

        for i in 0..a - 1 {
            ind.update(input[i]);
            assert!(!ind.is_primed(), "[{}] must not be primed", i);
        }

        for i in a - 1..input.len() {
            ind.update(input[i]);
            assert!(ind.is_primed(), "[{}] must be primed", i);
        }
    }

    #[test]
    fn test_entity_updates() {
        let input = testdata::testdata::test_input();
        let expected = testdata::expected_a2_y2_z14();

        let mut scalar_ind = create(2, 2, 14);
        let mut bar_ind = create(2, 2, 14);
        let mut quote_ind = create(2, 2, 14);
        let mut trade_ind = create(2, 2, 14);

        for i in 0..input.len() {
            let time = i as i64;

            let scalar_out = scalar_ind.update_scalar(&Scalar::new(time, input[i]));
            check("scalar", i, expected[i], scalar_out[0].downcast_ref::<Scalar>().unwrap().value);

            let bar_out = bar_ind.update_bar(&Bar {
                time,
                open: input[i],
                high: input[i],
                low: input[i],
                close: input[i],
                volume: 0.0,
            });
            check("bar", i, expected[i], bar_out[0].downcast_ref::<Scalar>().unwrap().value);

            let quote_out = quote_ind.update_quote(&Quote {
                time,
                bid_price: input[i],
                bid_size: 0.0,
                ask_price: input[i],
                ask_size: 0.0,
            });
            check("quote", i, expected[i], quote_out[0].downcast_ref::<Scalar>().unwrap().value);

            let trade_out = trade_ind.update_trade(&Trade {
                time,
                price: input[i],
                volume: 0.0,
            });
            check("trade", i, expected[i], trade_out[0].downcast_ref::<Scalar>().unwrap().value);
        }
    }

    #[test]
    fn test_metadata() {
        let ind = DoubleSmoothedMomenta::new(&DoubleSmoothedMomentaParams::default()).unwrap();
        let meta = ind.metadata();

        assert_eq!(meta.identifier, Identifier::DoubleSmoothedMomenta);
        assert_eq!(meta.mnemonic, "dm(2,2,14)");
        assert_eq!(meta.description, "Double-Smoothed Momenta dm(2,2,14)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
        assert_eq!(meta.outputs[0].mnemonic, "dm(2,2,14)");
    }

    #[test]
    fn test_mnemonic_component_combinations() {
        let cases: Vec<(Option<BarComponent>, Option<QuoteComponent>, Option<TradeComponent>, &str)> = vec![
            (None, None, None, "dm(2,2,14)"),
            (Some(BarComponent::Median), None, None, "dm(2,2,14, hl/2)"),
            (None, Some(QuoteComponent::Bid), None, "dm(2,2,14, b)"),
            (None, None, Some(TradeComponent::Volume), "dm(2,2,14, v)"),
            (Some(BarComponent::Open), Some(QuoteComponent::Bid), None, "dm(2,2,14, o, b)"),
            (Some(BarComponent::High), None, Some(TradeComponent::Volume), "dm(2,2,14, h, v)"),
            (None, Some(QuoteComponent::Ask), Some(TradeComponent::Volume), "dm(2,2,14, a, v)"),
        ];

        for (bc, qc, tc, expected) in cases {
            let ind = DoubleSmoothedMomenta::new(&DoubleSmoothedMomentaParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            })
            .unwrap();

            assert_eq!(ind.metadata().mnemonic, expected);
        }
    }

    #[test]
    fn test_custom_mnemonic() {
        let ind = create(10, 3, 5);
        assert_eq!(ind.metadata().mnemonic, "dm(10,3,5)");
    }

    #[test]
    fn test_invalid_params() {
        assert!(DoubleSmoothedMomenta::new(&DoubleSmoothedMomentaParams { a: 0, ..Default::default() }).is_err());
        assert!(DoubleSmoothedMomenta::new(&DoubleSmoothedMomentaParams { y: 0, ..Default::default() }).is_err());
        assert!(DoubleSmoothedMomenta::new(&DoubleSmoothedMomentaParams { z: 0, ..Default::default() }).is_err());
    }
}
