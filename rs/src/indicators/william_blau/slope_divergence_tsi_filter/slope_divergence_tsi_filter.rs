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
use crate::indicators::william_blau::true_strength_index::true_strength_index::{
    TrueStrengthIndex, TrueStrengthIndexParams,
};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Slope Divergence TSI Filter indicator.
///
/// The field names `q`, `r`, `s`, `u`, `x` and `y` are the canonical symbols from
/// William Blau's Momentum, Direction, and Divergence (Wiley, 1995), chapter 12 and
/// Appendix B, Figure B-25.
pub struct SlopeDivergenceTsiFilterParams {
    /// The TSI momentum look-back period; momentum is `C_k - C_(k-(q-1))`. Must be > 0. Default 2.
    pub q: usize,
    /// The period of the 1st (innermost) EMA of the TSI smoothing cascade. Must be > 0. Default 32.
    pub r: usize,
    /// The period of the 2nd EMA of the TSI smoothing cascade. Must be > 0. Default 32.
    pub s: usize,
    /// The period of the 3rd (outermost) EMA of the TSI smoothing cascade. Must be > 0. Default 7.
    /// `u=1` switches the 3rd stage off, yielding the book's raw double-smoothed TSI.
    pub u: usize,
    /// The period of the 1st EMA of the price reference `DEMA(close, x, y)`. Must be > 0. Default 32.
    pub x: usize,
    /// The period of the 2nd EMA of the price reference `DEMA(close, x, y)`. Must be > 0. Default 7.
    /// `y=1` makes the price reference a single EMA.
    pub y: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for SlopeDivergenceTsiFilterParams {
    fn default() -> Self {
        Self {
            q: 2,
            r: 32,
            s: 32,
            u: 7,
            x: 32,
            y: 7,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Slope Divergence TSI Filter indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SlopeDivergenceTsiFilterOutput {
    /// The Slope Divergence TSI Filter value (range [-100, +100]).
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

/// William Blau's Slope Divergence TSI Filter (SD_TSI).
///
/// A trend/congestion prefilter built on the True Strength Index. It keeps the
/// TSI value only when the slope of the TSI agrees in sign with the slope of a
/// separate double EMA of price; otherwise it outputs 0 (a slope divergence, or
/// congestion zone):
///
///   ind_k = TSI(close, q, r, s, u)_k
///   ref_k = DEMA(close, x, y)_k = EMA(EMA(close, x), y)_k
///
///   SD_TSI_k = ind_k   if ind_k - ind_(k-1) > 0 and ref_k - ref_(k-1) > 0
///            = ind_k   if ind_k - ind_(k-1) < 0 and ref_k - ref_(k-1) < 0
///            = 0       otherwise
///
/// The gate is strict (book Fig. B-25): a flat slope on either series yields 0.
/// The output range is [-100, +100].
///
/// Priming: the TSI is NaN for bars 0..q-2 (momentum look-back), so SD_TSI is NaN
/// there. The price DEMA seeds at bar 0 and advances every bar, including through
/// the TSI warm-up. At the first finite TSI bar there is no prior TSI value, hence
/// no slope, so the output is 0.0.
pub struct SlopeDivergenceTsiFilter {
    line: LineIndicator,

    // The TSI oscillator (its signal line is unused, so ul=1).
    tsi: TrueStrengthIndex,

    // Price reference: DEMA(close, x, y) = EMA(EMA(close, x), y).
    reference_x: Ema,
    reference_y: Ema,

    // Slope state: the previous finite TSI and the previous-bar reference.
    previous_tsi: f64,
    has_previous_tsi: bool,
    previous_reference: f64,

    primed: bool,
}

impl SlopeDivergenceTsiFilter {
    /// Creates a new Slope Divergence TSI Filter from the given parameters.
    pub fn new(params: &SlopeDivergenceTsiFilterParams) -> Result<Self, String> {
        let invalid = "invalid slope divergence tsi filter parameters";

        let q = params.q;
        let r = params.r;
        let s = params.s;
        let u = params.u;
        let x = params.x;
        let y = params.y;

        if q < 1 {
            return Err(format!("{}: q should be greater than 0", invalid));
        }
        if r < 1 {
            return Err(format!("{}: r should be greater than 0", invalid));
        }
        if s < 1 {
            return Err(format!("{}: s should be greater than 0", invalid));
        }
        if u < 1 {
            return Err(format!("{}: u should be greater than 0", invalid));
        }
        if x < 1 {
            return Err(format!("{}: x should be greater than 0", invalid));
        }
        if y < 1 {
            return Err(format!("{}: y should be greater than 0", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "sdtsi({},{},{},{},{},{}{})",
            q, r, s, u, x, y,
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Slope Divergence TSI Filter {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let tsi = TrueStrengthIndex::new(&TrueStrengthIndexParams {
            q,
            r,
            s,
            u,
            ul: 1,
            ..Default::default()
        })?;

        Ok(Self {
            line,
            tsi,
            reference_x: Ema::new(x),
            reference_y: Ema::new(y),
            previous_tsi: 0.0,
            has_previous_tsi: false,
            previous_reference: 0.0,
            primed: false,
        })
    }

    /// Core update logic. Returns the SD_TSI value or NaN during the TSI warm-up.
    pub fn update(&mut self, sample: f64) -> f64 {
        // The price reference advances every bar (it has no NaN warm-up).
        let reference = self.reference_y.update(self.reference_x.update(sample));

        let (tsi, _) = self.tsi.update(sample);
        if tsi.is_nan() {
            // TSI momentum warm-up: keep the previous-bar reference current.
            self.previous_reference = reference;
            return f64::NAN;
        }

        let mut result = 0.0;

        // At the first finite TSI there is no prior TSI, hence no slope.
        if self.has_previous_tsi {
            let delta_tsi = tsi - self.previous_tsi;
            let delta_reference = reference - self.previous_reference;

            // Keep the TSI only when both slopes are strictly same-signed.
            if (delta_tsi > 0.0 && delta_reference > 0.0) || (delta_tsi < 0.0 && delta_reference < 0.0) {
                result = tsi;
            }
        }

        self.previous_tsi = tsi;
        self.has_previous_tsi = true;
        self.previous_reference = reference;
        self.primed = true;

        result
    }
}

impl Indicator for SlopeDivergenceTsiFilter {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::SlopeDivergenceTsiFilter,
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

    const TOLERANCE: f64 = 1e-13;

    struct Combo {
        r: usize,
        s: usize,
        u: usize,
        x: usize,
        y: usize,
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

    fn create(params: SlopeDivergenceTsiFilterParams) -> SlopeDivergenceTsiFilter {
        SlopeDivergenceTsiFilter::new(&params).unwrap()
    }

    #[test]
    fn test_reference_data_all_combos() {
        // Every expected array uses the book momentum q=2.
        let combos = vec![
            Combo { r: 32, s: 32, u: 7, x: 32, y: 7, expected: testdata::expected_r32_s32_u7_x32_y7() },
            Combo { r: 32, s: 32, u: 1, x: 32, y: 1, expected: testdata::expected_r32_s32_u1_x32_y1() },
            Combo { r: 32, s: 32, u: 7, x: 32, y: 1, expected: testdata::expected_r32_s32_u7_x32_y1() },
            Combo { r: 32, s: 32, u: 1, x: 32, y: 7, expected: testdata::expected_r32_s32_u1_x32_y7() },
            Combo { r: 1, s: 1, u: 1, x: 1, y: 1, expected: testdata::expected_r1_s1_u1_x1_y1() },
            Combo { r: 20, s: 5, u: 3, x: 20, y: 3, expected: testdata::expected_r20_s5_u3_x20_y3() },
            Combo { r: 32, s: 13, u: 3, x: 32, y: 7, expected: testdata::expected_r32_s13_u3_x32_y7() },
            Combo { r: 12, s: 12, u: 1, x: 12, y: 1, expected: testdata::expected_r12_s12_u1_x12_y1() },
            Combo { r: 25, s: 13, u: 1, x: 25, y: 1, expected: testdata::expected_r25_s13_u1_x25_y1() },
            Combo { r: 64, s: 64, u: 7, x: 32, y: 7, expected: testdata::expected_r64_s64_u7_x32_y7() },
            Combo { r: 32, s: 32, u: 7, x: 16, y: 3, expected: testdata::expected_r32_s32_u7_x16_y3() },
            Combo { r: 5, s: 5, u: 5, x: 5, y: 5, expected: testdata::expected_r5_s5_u5_x5_y5() },
            Combo { r: 10, s: 10, u: 1, x: 10, y: 1, expected: testdata::expected_r10_s10_u1_x10_y1() },
            Combo { r: 40, s: 20, u: 5, x: 32, y: 7, expected: testdata::expected_r40_s20_u5_x32_y7() },
            Combo { r: 32, s: 5, u: 1, x: 32, y: 1, expected: testdata::expected_r32_s5_u1_x32_y1() },
            Combo { r: 50, s: 25, u: 1, x: 50, y: 1, expected: testdata::expected_r50_s25_u1_x50_y1() },
        ];

        let input = testdata::testdata::test_input();

        for combo in &combos {
            let mut ind = create(SlopeDivergenceTsiFilterParams {
                q: 2,
                r: combo.r,
                s: combo.s,
                u: combo.u,
                x: combo.x,
                y: combo.y,
                ..Default::default()
            });

            for i in 0..input.len() {
                check("sdtsi", i, combo.expected[i], ind.update(input[i]));
            }
        }
    }

    #[test]
    fn test_passthrough() {
        let mut ind = create(SlopeDivergenceTsiFilterParams { q: 2, r: 1, s: 1, u: 1, x: 1, y: 1, ..Default::default() });

        assert!(ind.update(10.0).is_nan()); // momentum undefined
        assert_eq!(ind.update(12.0), 0.0); // first finite TSI, no slope
        assert_eq!(ind.update(11.0), -100.0); // both falling
        assert_eq!(ind.update(13.0), 100.0); // both rising
        assert_eq!(ind.update(14.0), 0.0); // TSI flat
    }

    #[test]
    fn test_warm_up_region() {
        let input = testdata::testdata::test_input();
        let q = 5;
        let mut ind = create(SlopeDivergenceTsiFilterParams { q, ..Default::default() });

        for i in 0..q - 1 {
            assert!(ind.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        assert_eq!(ind.update(input[q - 1]), 0.0);

        for i in q..input.len() {
            assert!(!ind.update(input[i]).is_nan(), "[{}] unexpected NaN", i);
        }
    }

    #[test]
    fn test_bounded() {
        let input = testdata::testdata::test_input();
        let mut ind = create(SlopeDivergenceTsiFilterParams::default());

        for i in 0..input.len() {
            let value = ind.update(input[i]);
            if !value.is_nan() {
                assert!(value >= -100.0, "[{}] below -100", i);
                assert!(value <= 100.0, "[{}] above 100", i);
            }
        }
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::testdata::test_input();
        let q = 5;
        let mut ind = create(SlopeDivergenceTsiFilterParams { q, ..Default::default() });

        for i in 0..q - 1 {
            ind.update(input[i]);
            assert!(!ind.is_primed(), "[{}] must not be primed", i);
        }

        for i in q - 1..input.len() {
            ind.update(input[i]);
            assert!(ind.is_primed(), "[{}] must be primed", i);
        }
    }

    #[test]
    fn test_entity_updates() {
        let input = testdata::testdata::test_input();
        let expected = testdata::expected_r32_s32_u7_x32_y7();

        let mut scalar_ind = create(SlopeDivergenceTsiFilterParams::default());
        let mut bar_ind = create(SlopeDivergenceTsiFilterParams::default());
        let mut quote_ind = create(SlopeDivergenceTsiFilterParams::default());
        let mut trade_ind = create(SlopeDivergenceTsiFilterParams::default());

        for i in 0..input.len() {
            let time = i as i64;

            let scalar_out = scalar_ind.update_scalar(&Scalar::new(time, input[i]));
            assert_eq!(scalar_out.len(), 1);
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
        let ind = create(SlopeDivergenceTsiFilterParams::default());
        let meta = ind.metadata();

        assert_eq!(meta.identifier, Identifier::SlopeDivergenceTsiFilter);
        assert_eq!(meta.mnemonic, "sdtsi(2,32,32,7,32,7)");
        assert_eq!(meta.description, "Slope Divergence TSI Filter sdtsi(2,32,32,7,32,7)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, SlopeDivergenceTsiFilterOutput::Value as i32);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
        assert_eq!(meta.outputs[0].mnemonic, "sdtsi(2,32,32,7,32,7)");
    }

    #[test]
    fn test_mnemonic_component_combinations() {
        let cases: Vec<(Option<BarComponent>, Option<QuoteComponent>, Option<TradeComponent>, &str)> = vec![
            (None, None, None, "sdtsi(2,32,32,7,32,7)"),
            (Some(BarComponent::Median), None, None, "sdtsi(2,32,32,7,32,7, hl/2)"),
            (None, Some(QuoteComponent::Bid), None, "sdtsi(2,32,32,7,32,7, b)"),
            (None, None, Some(TradeComponent::Volume), "sdtsi(2,32,32,7,32,7, v)"),
            (Some(BarComponent::Open), Some(QuoteComponent::Bid), None, "sdtsi(2,32,32,7,32,7, o, b)"),
            (Some(BarComponent::High), None, Some(TradeComponent::Volume), "sdtsi(2,32,32,7,32,7, h, v)"),
            (None, Some(QuoteComponent::Ask), Some(TradeComponent::Volume), "sdtsi(2,32,32,7,32,7, a, v)"),
        ];

        for (bc, qc, tc, expected) in cases {
            let ind = create(SlopeDivergenceTsiFilterParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            });

            assert_eq!(ind.metadata().mnemonic, expected);
        }
    }

    #[test]
    fn test_custom_mnemonic() {
        let ind = create(SlopeDivergenceTsiFilterParams { q: 3, r: 20, s: 5, u: 3, x: 20, y: 3, ..Default::default() });
        assert_eq!(ind.metadata().mnemonic, "sdtsi(3,20,5,3,20,3)");
    }

    #[test]
    fn test_invalid_params() {
        let new = |p: SlopeDivergenceTsiFilterParams| SlopeDivergenceTsiFilter::new(&p);

        assert!(new(SlopeDivergenceTsiFilterParams { q: 0, ..Default::default() }).is_err());
        assert!(new(SlopeDivergenceTsiFilterParams { r: 0, ..Default::default() }).is_err());
        assert!(new(SlopeDivergenceTsiFilterParams { s: 0, ..Default::default() }).is_err());
        assert!(new(SlopeDivergenceTsiFilterParams { u: 0, ..Default::default() }).is_err());
        assert!(new(SlopeDivergenceTsiFilterParams { x: 0, ..Default::default() }).is_err());
        assert!(new(SlopeDivergenceTsiFilterParams { y: 0, ..Default::default() }).is_err());
    }
}
