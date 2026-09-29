use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Tick Volume Indicator.
///
/// The field names `r`, `s` and `u` are the canonical symbols from William
/// Blau's *Momentum, Direction, and Divergence* (Wiley, 1995), chapters 4 and 10.
///
/// The indicator consumes upticks and downticks (derived from the bar range,
/// or from consecutive prices for single-valued samples), so it has no
/// configurable price-component fields.
pub struct TickVolumeIndicatorParams {
    /// Period of the 1st (innermost) EMA, applied to the upticks and downticks. Must be > 0. Default 12.
    pub r: usize,
    /// Period of the 2nd EMA in the cascade. Must be > 0. Default 12.
    pub s: usize,
    /// Period of the 3rd (outermost) EMA in the cascade; 1 yields the book's double-smoothed TVI(r, s). Must be > 0. Default 1.
    pub u: usize,
}

impl Default for TickVolumeIndicatorParams {
    fn default() -> Self {
        Self { r: 12, s: 12, u: 1 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Tick Volume Indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum TickVolumeIndicatorOutput {
    /// The Tick Volume Indicator oscillator value (range [-100, +100]).
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

/// William Blau's Tick Volume Indicator (TVI).
///
/// A normalized, double-/triple-smoothed oscillator built from the balance of
/// upticks vs downticks inside each bar, bounded to [-100, +100] (Blau ch.4, ch.10):
///
///   tvi_k = 100 * (TEMA(up, r, s, u) - TEMA(down, r, s, u))
///               / (TEMA(up, r, s, u) + TEMA(down, r, s, u))
///
/// where TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u). Setting u=1 recovers the
/// book's double-smoothed TVI(r, s), because EMA(., 1) is a passthrough. Because it
/// is built from intra-bar tick direction rather than from the close vs a previous
/// close, the TVI is immune to opening gaps.
///
/// The inputs are two non-negative series, upticks and downticks. Genuine tick
/// counts are fed through `update`. The entity updates derive a deterministic proxy:
///   - Bar: up = close - low, down = high - close (the intra-bar range split);
///   - Scalar, Trade, Quote: a magnitude tick rule against the previous value
///     (value, price or mid price): up = max(x - previous, 0),
///     down = max(previous - x, 0). The first sample yields (0, 0). On such a
///     single-valued series the TVI reduces to a True Strength Index of the
///     one-step momentum.
///
/// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
/// received value, so there is no NaN warm-up region and the output is finite for
/// every update. Division guard: denominator 0 (a fully flat market) -> 0.0.
pub struct TickVolumeIndicator {
    up_r: Ema,
    up_s: Ema,
    up_u: Ema,
    down_r: Ema,
    down_s: Ema,
    down_u: Ema,
    previous: f64,
    has_previous: bool,
    primed: bool,
    mnemonic: String,
}

impl TickVolumeIndicator {
    /// Creates a new Tick Volume Indicator from the given parameters.
    pub fn new(params: &TickVolumeIndicatorParams) -> Result<Self, String> {
        let invalid = "invalid tick volume indicator parameters";

        let mut r = params.r;
        if r == 0 {
            r = 12;
        }
        let mut s = params.s;
        if s == 0 {
            s = 12;
        }
        let mut u = params.u;
        if u == 0 {
            u = 1;
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

        let mnemonic = format!("tvi({},{},{})", r, s, u);

        Ok(Self {
            up_r: Ema::new(r),
            up_s: Ema::new(s),
            up_u: Ema::new(u),
            down_r: Ema::new(r),
            down_s: Ema::new(s),
            down_u: Ema::new(u),
            previous: 0.0,
            has_previous: false,
            primed: false,
            mnemonic,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update taking one bar's upticks and downticks, returning the TVI value.
    pub fn update(&mut self, upticks: f64, downticks: f64) -> f64 {
        // Upticks cascade: TEMA(up, r, s, u).
        let up = self.up_u.update(self.up_s.update(self.up_r.update(upticks)));
        // Downticks cascade: TEMA(down, r, s, u).
        let down = self.down_u.update(self.down_s.update(self.down_r.update(downticks)));

        self.primed = true;

        // Division guard (Appendix B): fully flat smoothed volume -> 0.0.
        let denominator = up + down;
        if denominator == 0.0 {
            return 0.0;
        }

        100.0 * (up - down) / denominator
    }

    /// Derives the upticks and downticks from the change of the value vs the
    /// previous value and updates the indicator.
    fn update_tick_rule(&mut self, value: f64) -> f64 {
        let mut upticks = 0.0;
        let mut downticks = 0.0;
        if self.has_previous {
            let diff = value - self.previous;
            upticks = diff.max(0.0);
            downticks = (-diff).max(0.0);
        }

        self.previous = value;
        self.has_previous = true;

        self.update(upticks, downticks)
    }

    /// Wraps the TVI value into the output.
    fn wrap(time: i64, value: f64) -> Output {
        vec![Box::new(Scalar { time, value })]
    }
}

impl Indicator for TickVolumeIndicator {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Tick Volume Indicator {}", self.mnemonic);
        build_metadata(
            Identifier::TickVolumeIndicator,
            &self.mnemonic,
            &desc,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: desc.clone(),
            }],
        )
    }

    /// A scalar carries a single value, so the tick rule is applied to it.
    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        Self::wrap(sample.time, self.update_tick_rule(sample.value))
    }

    /// A bar splits its range: up = close - low, down = high - close.
    fn update_bar(&mut self, sample: &Bar) -> Output {
        Self::wrap(
            sample.time,
            self.update(sample.close - sample.low, sample.high - sample.close),
        )
    }

    /// A quote applies the tick rule to its mid price.
    fn update_quote(&mut self, sample: &Quote) -> Output {
        Self::wrap(sample.time, self.update_tick_rule(sample.mid()))
    }

    /// A trade applies the tick rule to its price.
    fn update_trade(&mut self, sample: &Trade) -> Output {
        Self::wrap(sample.time, self.update_tick_rule(sample.price))
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

    const TOLERANCE: f64 = 1e-10;

    // Tick rule with passthrough stages: first -> 0, up -> +100, down -> -100, flat -> 0.
    const TICK_VALUES: [f64; 4] = [10.0, 12.0, 11.0, 11.0];
    const TICK_EXPECTED: [f64; 4] = [0.0, 100.0, -100.0, 0.0];

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

    fn check_output(name: &str, i: usize, time: i64, exp: f64, out: &Output) {
        assert_eq!(out.len(), 1, "{}[{}]: expected 1 output", name, i);
        let scalar = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(scalar.time, time, "{}[{}]: time", name, i);
        check(name, i, exp, scalar.value);
    }

    fn passthrough() -> TickVolumeIndicator {
        TickVolumeIndicator::new(&TickVolumeIndicatorParams { r: 1, s: 1, u: 1 }).unwrap()
    }

    struct Combo {
        r: usize,
        s: usize,
        u: usize,
        expected: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { r: 12, s: 12, u: 1, expected: testdata::expected_r12_s12_u1() },
            Combo { r: 25, s: 13, u: 1, expected: testdata::expected_r25_s13_u1() },
            Combo { r: 32, s: 32, u: 5, expected: testdata::expected_r32_s32_u5() },
            Combo { r: 1, s: 1, u: 1, expected: testdata::expected_r1_s1_u1() },
            Combo { r: 32, s: 5, u: 1, expected: testdata::expected_r32_s5_u1() },
            Combo { r: 12, s: 12, u: 5, expected: testdata::expected_r12_s12_u5() },
            Combo { r: 20, s: 5, u: 3, expected: testdata::expected_r20_s5_u3() },
            Combo { r: 5, s: 5, u: 5, expected: testdata::expected_r5_s5_u5() },
            Combo { r: 32, s: 32, u: 1, expected: testdata::expected_r32_s32_u1() },
            Combo { r: 10, s: 10, u: 1, expected: testdata::expected_r10_s10_u1() },
            Combo { r: 50, s: 25, u: 1, expected: testdata::expected_r50_s25_u1() },
            Combo { r: 12, s: 26, u: 9, expected: testdata::expected_r12_s26_u9() },
            Combo { r: 3, s: 3, u: 3, expected: testdata::expected_r3_s3_u3() },
            Combo { r: 7, s: 4, u: 2, expected: testdata::expected_r7_s4_u2() },
            Combo { r: 64, s: 1, u: 1, expected: testdata::expected_r64_s1_u1() },
            Combo { r: 12, s: 12, u: 3, expected: testdata::expected_r12_s12_u3() },
        ];

        let upticks = testdata::testdata::test_upticks();
        let downticks = testdata::testdata::test_downticks();

        for combo in &combos {
            let mut ind = TickVolumeIndicator::new(&TickVolumeIndicatorParams {
                r: combo.r,
                s: combo.s,
                u: combo.u,
            })
            .unwrap();

            for i in 0..upticks.len() {
                check("tvi", i, combo.expected[i], ind.update(upticks[i], downticks[i]));
            }
        }
    }

    #[test]
    fn test_passthrough() {
        let mut ind = passthrough();

        check("mostly up", 0, 60.0, ind.update(8.0, 2.0));
        check("all down", 1, -100.0, ind.update(0.0, 5.0));
        // Flat market: denominator 0 -> division guard 0.0.
        assert_eq!(ind.update(0.0, 0.0), 0.0);
    }

    #[test]
    fn test_is_primed_after_first_update() {
        let upticks = testdata::testdata::test_upticks();
        let downticks = testdata::testdata::test_downticks();

        let mut ind = TickVolumeIndicator::new(&TickVolumeIndicatorParams::default()).unwrap();
        assert!(!TickVolumeIndicator::is_primed(&ind));
        ind.update(upticks[0], downticks[0]);
        assert!(TickVolumeIndicator::is_primed(&ind));

        let mut tick_ind = TickVolumeIndicator::new(&TickVolumeIndicatorParams::default()).unwrap();
        assert!(!TickVolumeIndicator::is_primed(&tick_ind));
        tick_ind.update_scalar(&Scalar { time: 0, value: 10.0 });
        assert!(TickVolumeIndicator::is_primed(&tick_ind));
    }

    #[test]
    fn test_mnemonic() {
        let ind = TickVolumeIndicator::new(&TickVolumeIndicatorParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "tvi(12,12,1)");
        assert_eq!(
            ind.metadata().description,
            "Tick Volume Indicator tvi(12,12,1)"
        );

        let ind2 = TickVolumeIndicator::new(&TickVolumeIndicatorParams { r: 32, s: 32, u: 5 }).unwrap();
        assert_eq!(ind2.metadata().mnemonic, "tvi(32,32,5)");

        // Zero parameters resolve to the defaults.
        let ind3 = TickVolumeIndicator::new(&TickVolumeIndicatorParams { r: 0, s: 0, u: 0 }).unwrap();
        assert_eq!(ind3.metadata().mnemonic, "tvi(12,12,1)");
    }

    #[test]
    fn test_metadata() {
        let ind = TickVolumeIndicator::new(&TickVolumeIndicatorParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::TickVolumeIndicator);
        assert_eq!(meta.mnemonic, "tvi(12,12,1)");
        assert_eq!(meta.description, "Tick Volume Indicator tvi(12,12,1)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, TickVolumeIndicatorOutput::Value as i32);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
        assert_eq!(meta.outputs[0].mnemonic, "tvi(12,12,1)");
        assert_eq!(
            meta.outputs[0].description,
            "Tick Volume Indicator tvi(12,12,1)"
        );
    }

    #[test]
    fn test_update_bar_maps_range() {
        let input = testdata::testdata::test_input();
        let open = testdata::testdata::test_open();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();
        let expected = testdata::expected_r12_s12_u1();

        let mut ind = TickVolumeIndicator::new(&TickVolumeIndicatorParams { r: 12, s: 12, u: 1 }).unwrap();

        for i in 0..input.len() {
            let out = ind.update_bar(&Bar {
                time: 42,
                open: open[i],
                high: high[i],
                low: low[i],
                close: input[i],
                volume: 0.0,
            });
            check_output("bar", i, 42, expected[i], &out);
        }
    }

    #[test]
    fn test_update_scalar_applies_tick_rule() {
        let mut ind = passthrough();
        for (i, value) in TICK_VALUES.iter().enumerate() {
            let out = ind.update_scalar(&Scalar { time: 42, value: *value });
            check_output("scalar", i, 42, TICK_EXPECTED[i], &out);
        }
    }

    #[test]
    fn test_update_trade_applies_tick_rule() {
        let mut ind = passthrough();
        for (i, price) in TICK_VALUES.iter().enumerate() {
            let out = ind.update_trade(&Trade { time: 42, price: *price, volume: 1.0 });
            check_output("trade", i, 42, TICK_EXPECTED[i], &out);
        }
    }

    #[test]
    fn test_update_quote_applies_tick_rule_to_mid() {
        let mut ind = passthrough();
        let quotes = [(9.0, 11.0), (11.0, 13.0), (10.0, 12.0), (10.5, 11.5)];
        for (i, (bid, ask)) in quotes.iter().enumerate() {
            let out = ind.update_quote(&Quote {
                time: 42,
                bid_price: *bid,
                ask_price: *ask,
                bid_size: 1.0,
                ask_size: 1.0,
            });
            check_output("quote", i, 42, TICK_EXPECTED[i], &out);
        }
    }

    #[test]
    fn test_update_scalar_matches_update_of_changes() {
        let input = testdata::testdata::test_input();

        let mut ind = TickVolumeIndicator::new(&TickVolumeIndicatorParams { r: 12, s: 12, u: 3 }).unwrap();
        let mut reference = TickVolumeIndicator::new(&TickVolumeIndicatorParams { r: 12, s: 12, u: 3 }).unwrap();

        for i in 0..input.len() {
            let diff = if i == 0 { 0.0 } else { input[i] - input[i - 1] };
            let exp = reference.update(diff.max(0.0), (-diff).max(0.0));
            let out = ind.update_scalar(&Scalar { time: 42, value: input[i] });
            check_output("scalar", i, 42, exp, &out);
        }
    }
}
