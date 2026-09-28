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

/// Parameters to create an instance of the Directional Trend Index indicator.
///
/// The field names `q`, `r`, `s`, `u` and `ul` are the canonical symbols from
/// William Blau's *Momentum, Direction, and Divergence* (Wiley, 1995).
///
/// The indicator consumes the high and low prices of a bar, so it has no
/// configurable price-component fields.
pub struct DirectionalTrendIndexParams {
    /// High-low momentum look-back period; the high and the low are compared with their values q-1 bars ago. Must be > 0. Default 2.
    pub q: usize,
    /// Period of the 1st (innermost) EMA, applied to the high-low momentum. Must be > 0. Default 20.
    pub r: usize,
    /// Period of the 2nd EMA in the cascade. Must be > 0. Default 5.
    pub s: usize,
    /// Period of the 3rd (outermost) EMA in the cascade; u = 1 switches it off. Must be > 0. Default 3.
    pub u: usize,
    /// Period of the signal-line EMA (second output); ul = 1 is a passthrough. Must be > 0. Default 3.
    pub ul: usize,
}

impl Default for DirectionalTrendIndexParams {
    fn default() -> Self {
        Self {
            q: 2,
            r: 20,
            s: 5,
            u: 3,
            ul: 3,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Directional Trend Index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DirectionalTrendIndexOutput {
    /// The Directional Trend Index oscillator value (range [-100, +100]).
    Dti = 1,
    /// The signal-line value: the ul-period EMA of the oscillator.
    Signal = 2,
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
    previous: f64,
    primed: bool,
}

impl Ema {
    fn new(period: usize) -> Self {
        Self {
            alpha: 2.0 / (period as f64 + 1.0),
            previous: 0.0,
            primed: false,
        }
    }

    fn update(&mut self, x: f64) -> f64 {
        if !self.primed {
            self.previous = x;
            self.primed = true;
            return self.previous;
        }
        self.previous = self.alpha * x + (1.0 - self.alpha) * self.previous;
        self.previous
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// William Blau's Directional Trend Index (DTI).
///
/// A double-/triple-smoothed High-Low Momentum oscillator bounded to [-100, +100],
/// paired with an EMA signal line (the Ergodic form):
///
///   dti_k    = 100 * TEMA(HLM, r, s, u)_k / TEMA(|HLM|, r, s, u)_k   (the oscillator)
///   signal_k = EMA(dti, ul)_k                                       (ul-period EMA)
///
/// where the High-Low Momentum is built from how far the high rose and the low
/// fell relative to q-1 bars ago:
///
///   HMU_k = max(high_k - high_(k-(q-1)), 0)      (upward high movement)
///   LMD_k = max(low_(k-(q-1)) - low_k, 0)        (downward low movement)
///   HLM_k = HMU_k - LMD_k                         (composite high-low momentum)
///   TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u)  (triple EMA cascade)
///
/// This is the True Strength Index structure applied to HLM instead of price
/// momentum. The inputs are the high and low prices only (no close).
///
/// The indicator produces two outputs:
///   - DTI: the oscillator, range [-100, +100];
///   - Signal: the ul-period EMA of the oscillator.
///
/// Priming convention (book / EasyLanguage): HLM is valid from bar q-1 (it needs a
/// high/low from q-1 bars ago), so all cascade stages seed there together; both
/// outputs are NaN for bars 0..q-2 and finite from bar q-1. For q = 1 there is no
/// NaN warm-up, but HLM == 0 on every bar, so the division guard yields dti == 0.0
/// for all bars. The signal EMA seeds on the first finite oscillator value.
/// Division guard: denominator == 0 -> oscillator 0.0.
pub struct DirectionalTrendIndex {
    q: usize,
    highs: Vec<f64>,
    lows: Vec<f64>,
    window_count: usize,
    window_index: usize,
    num_r: Ema,
    num_s: Ema,
    num_u: Ema,
    den_r: Ema,
    den_s: Ema,
    den_u: Ema,
    signal_ema: Ema,
    primed: bool,
    mnemonic: String,
}

impl DirectionalTrendIndex {
    /// Creates a new Directional Trend Index from the given parameters.
    pub fn new(params: &DirectionalTrendIndexParams) -> Result<Self, String> {
        let invalid = "invalid directional trend index parameters";

        let mut q = params.q;
        if q == 0 {
            q = 2;
        }
        let mut r = params.r;
        if r == 0 {
            r = 20;
        }
        let mut s = params.s;
        if s == 0 {
            s = 5;
        }
        let mut u = params.u;
        if u == 0 {
            u = 3;
        }
        let mut ul = params.ul;
        if ul == 0 {
            ul = 3;
        }

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
        if ul < 1 {
            return Err(format!("{}: ul should be greater than 0", invalid));
        }

        let mnemonic = format!("dti({},{},{},{},{})", q, r, s, u, ul);

        Ok(Self {
            q,
            highs: vec![0.0; q],
            lows: vec![0.0; q],
            window_count: 0,
            window_index: 0,
            num_r: Ema::new(r),
            num_s: Ema::new(s),
            num_u: Ema::new(u),
            den_r: Ema::new(r),
            den_s: Ema::new(s),
            den_u: Ema::new(u),
            signal_ema: Ema::new(ul),
            primed: false,
            mnemonic,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update taking one bar's high and low, returning (dti, signal).
    /// Both are NaN until q bars have been seen.
    pub fn update(&mut self, high: f64, low: f64) -> (f64, f64) {
        self.highs[self.window_index] = high;
        self.lows[self.window_index] = low;
        self.window_index = (self.window_index + 1) % self.q;

        if self.window_count < self.q {
            self.window_count += 1;
        }

        // HLM needs a high/low from q-1 bars ago. Until then neither output
        // exists -- do NOT advance the EMA cascades.
        if self.window_count < self.q {
            return (f64::NAN, f64::NAN);
        }

        // The oldest value in the full window sits at the next write position:
        // high_(k-(q-1)) and low_(k-(q-1)).
        let previous_high = self.highs[self.window_index];
        let previous_low = self.lows[self.window_index];

        // Upward high movement and downward low movement, each floored at 0.
        let hmu = (high - previous_high).max(0.0);
        let lmd = (previous_low - low).max(0.0);

        // Composite high-low momentum and its magnitude.
        let hlm = hmu - lmd;
        let abs_hlm = hlm.abs();

        // Numerator cascade: TEMA(HLM, r, s, u).
        let n = self.num_u.update(self.num_s.update(self.num_r.update(hlm)));
        // Denominator cascade: TEMA(|HLM|, r, s, u).
        let d = self.den_u.update(self.den_s.update(self.den_r.update(abs_hlm)));

        // Division guard: denominator 0 -> oscillator 0.0.
        let dti = if d != 0.0 { 100.0 * n / d } else { 0.0 };

        // Signal line = EMA(dti, ul); seeds on the first finite oscillator value.
        let signal = self.signal_ema.update(dti);
        self.primed = true;

        (dti, signal)
    }

    /// Updates the indicator and wraps the two outputs.
    fn update_entity(&mut self, time: i64, high: f64, low: f64) -> Output {
        let (dti, signal) = self.update(high, low);
        vec![
            Box::new(Scalar { time, value: dti }),
            Box::new(Scalar {
                time,
                value: signal,
            }),
        ]
    }
}

impl Indicator for DirectionalTrendIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Directional Trend Index {}", self.mnemonic);
        build_metadata(
            Identifier::DirectionalTrendIndex,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} dti", self.mnemonic),
                    description: format!("{} DTI", desc),
                },
                OutputText {
                    mnemonic: format!("{} signal", self.mnemonic),
                    description: format!("{} signal", desc),
                },
            ],
        )
    }

    /// A scalar carries a single value, used as both the high and the low.
    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = sample.value;
        self.update_entity(sample.time, v, v)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        self.update_entity(sample.time, sample.high, sample.low)
    }

    /// A quote maps the mid price to both the high and the low.
    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (sample.bid_price + sample.ask_price) / 2.0;
        self.update_entity(sample.time, v, v)
    }

    /// A trade carries a single price, used as both the high and the low.
    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = sample.price;
        self.update_entity(sample.time, v, v)
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::super::testdata;
    use super::*;

    const TOLERANCE: f64 = 1e-13;

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

    fn passthrough(q: usize) -> DirectionalTrendIndex {
        DirectionalTrendIndex::new(&DirectionalTrendIndexParams {
            q,
            r: 1,
            s: 1,
            u: 1,
            ul: 1,
        })
        .unwrap()
    }

    fn scalar_value(out: &Output, i: usize) -> f64 {
        out[i].downcast_ref::<Scalar>().unwrap().value
    }

    struct Combo {
        q: usize,
        r: usize,
        s: usize,
        u: usize,
        ul: usize,
        dti: Vec<f64>,
        signal: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { q: 2, r: 20, s: 5, u: 3, ul: 3, dti: testdata::expected_q2_r20_s5_u3(), signal: testdata::expected_q2_r20_s5_u3_sig_ul3() },
            Combo { q: 2, r: 25, s: 13, u: 1, ul: 3, dti: testdata::expected_q2_r25_s13_u1(), signal: testdata::expected_q2_r25_s13_u1_sig_ul3() },
            Combo { q: 2, r: 20, s: 5, u: 1, ul: 3, dti: testdata::expected_q2_r20_s5_u1(), signal: testdata::expected_q2_r20_s5_u1_sig_ul3() },
            Combo { q: 2, r: 28, s: 28, u: 5, ul: 3, dti: testdata::expected_q2_r28_s28_u5(), signal: testdata::expected_q2_r28_s28_u5_sig_ul3() },
            Combo { q: 2, r: 1, s: 1, u: 1, ul: 3, dti: testdata::expected_q2_r1_s1_u1(), signal: testdata::expected_q2_r1_s1_u1_sig_ul3() },
            Combo { q: 3, r: 20, s: 5, u: 3, ul: 3, dti: testdata::expected_q3_r20_s5_u3(), signal: testdata::expected_q3_r20_s5_u3_sig_ul3() },
            Combo { q: 5, r: 20, s: 5, u: 3, ul: 3, dti: testdata::expected_q5_r20_s5_u3(), signal: testdata::expected_q5_r20_s5_u3_sig_ul3() },
            Combo { q: 2, r: 13, s: 13, u: 1, ul: 3, dti: testdata::expected_q2_r13_s13_u1(), signal: testdata::expected_q2_r13_s13_u1_sig_ul3() },
            Combo { q: 2, r: 40, s: 20, u: 1, ul: 3, dti: testdata::expected_q2_r40_s20_u1(), signal: testdata::expected_q2_r40_s20_u1_sig_ul3() },
            Combo { q: 2, r: 5, s: 5, u: 5, ul: 3, dti: testdata::expected_q2_r5_s5_u5(), signal: testdata::expected_q2_r5_s5_u5_sig_ul3() },
            Combo { q: 1, r: 20, s: 5, u: 3, ul: 3, dti: testdata::expected_q1_r20_s5_u3(), signal: testdata::expected_q1_r20_s5_u3_sig_ul3() },
            Combo { q: 10, r: 20, s: 5, u: 1, ul: 3, dti: testdata::expected_q10_r20_s5_u1(), signal: testdata::expected_q10_r20_s5_u1_sig_ul3() },
            Combo { q: 2, r: 9, s: 3, u: 1, ul: 3, dti: testdata::expected_q2_r9_s3_u1(), signal: testdata::expected_q2_r9_s3_u1_sig_ul3() },
            Combo { q: 2, r: 64, s: 64, u: 1, ul: 3, dti: testdata::expected_q2_r64_s64_u1(), signal: testdata::expected_q2_r64_s64_u1_sig_ul3() },
            Combo { q: 4, r: 28, s: 28, u: 5, ul: 3, dti: testdata::expected_q4_r28_s28_u5(), signal: testdata::expected_q4_r28_s28_u5_sig_ul3() },
            Combo { q: 2, r: 7, s: 4, u: 2, ul: 3, dti: testdata::expected_q2_r7_s4_u2(), signal: testdata::expected_q2_r7_s4_u2_sig_ul3() },
        ];

        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        for combo in &combos {
            let mut ind = DirectionalTrendIndex::new(&DirectionalTrendIndexParams {
                q: combo.q,
                r: combo.r,
                s: combo.s,
                u: combo.u,
                ul: combo.ul,
            })
            .unwrap();

            for i in 0..high.len() {
                let (dti, signal) = ind.update(high[i], low[i]);
                check("dti", i, combo.dti[i], dti);
                check("signal", i, combo.signal[i], signal);
            }
        }
    }

    #[test]
    fn test_passthrough() {
        let mut ind = passthrough(2);

        let (dti, signal) = ind.update(10.0, 9.0);
        assert!(dti.is_nan());
        assert!(signal.is_nan());
        // HMU=+2, LMD=0.
        assert_eq!(ind.update(12.0, 11.0), (100.0, 100.0));
        // HMU=0, LMD=3.
        assert_eq!(ind.update(11.0, 8.0), (-100.0, -100.0));
        // Inside bar -> division guard.
        assert_eq!(ind.update(10.0, 9.0), (0.0, 0.0));
    }

    #[test]
    fn test_q_one_yields_zero_on_every_bar() {
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        let mut ind = DirectionalTrendIndex::new(&DirectionalTrendIndexParams {
            q: 1,
            ..Default::default()
        })
        .unwrap();

        for i in 0..high.len() {
            assert_eq!(ind.update(high[i], low[i]), (0.0, 0.0), "[{}]", i);
        }
    }

    #[test]
    fn test_signal_equals_dti_when_ul_is_one() {
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        let mut ind = DirectionalTrendIndex::new(&DirectionalTrendIndexParams {
            ul: 1,
            ..Default::default()
        })
        .unwrap();

        for i in 0..high.len() {
            let (dti, signal) = ind.update(high[i], low[i]);
            if dti.is_nan() {
                assert!(signal.is_nan(), "[{}] expected NaN signal", i);
            } else {
                assert_eq!(dti, signal, "[{}]", i);
            }
        }
    }

    #[test]
    fn test_is_primed_from_bar_q_minus_one() {
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        for q in [2, 3, 5, 10] {
            let mut ind = DirectionalTrendIndex::new(&DirectionalTrendIndexParams {
                q,
                ..Default::default()
            })
            .unwrap();
            assert!(!DirectionalTrendIndex::is_primed(&ind));

            for i in 0..q - 1 {
                let (dti, signal) = ind.update(high[i], low[i]);
                assert!(!DirectionalTrendIndex::is_primed(&ind), "q={} [{}] primed too early", q, i);
                assert!(dti.is_nan());
                assert!(signal.is_nan());
            }

            for i in q - 1..high.len() {
                ind.update(high[i], low[i]);
                assert!(DirectionalTrendIndex::is_primed(&ind), "q={} [{}] not primed", q, i);
            }
        }
    }

    #[test]
    fn test_is_primed_after_first_bar_when_q_is_one() {
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        let mut ind = DirectionalTrendIndex::new(&DirectionalTrendIndexParams {
            q: 1,
            ..Default::default()
        })
        .unwrap();
        assert!(!DirectionalTrendIndex::is_primed(&ind));

        ind.update(high[0], low[0]);
        assert!(DirectionalTrendIndex::is_primed(&ind));
    }

    #[test]
    fn test_mnemonic() {
        let ind = DirectionalTrendIndex::new(&DirectionalTrendIndexParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "dti(2,20,5,3,3)");
        assert_eq!(
            ind.metadata().description,
            "Directional Trend Index dti(2,20,5,3,3)"
        );

        let ind2 = DirectionalTrendIndex::new(&DirectionalTrendIndexParams {
            q: 4,
            r: 28,
            s: 28,
            u: 5,
            ul: 1,
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "dti(4,28,28,5,1)");
    }

    #[test]
    fn test_metadata() {
        let ind = DirectionalTrendIndex::new(&DirectionalTrendIndexParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::DirectionalTrendIndex);
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(meta.outputs[0].kind, DirectionalTrendIndexOutput::Dti as i32);
        assert_eq!(
            meta.outputs[1].kind,
            DirectionalTrendIndexOutput::Signal as i32
        );
        assert_eq!(meta.outputs[0].mnemonic, "dti(2,20,5,3,3) dti");
        assert_eq!(
            meta.outputs[0].description,
            "Directional Trend Index dti(2,20,5,3,3) DTI"
        );
        assert_eq!(meta.outputs[1].mnemonic, "dti(2,20,5,3,3) signal");
        assert_eq!(
            meta.outputs[1].description,
            "Directional Trend Index dti(2,20,5,3,3) signal"
        );
    }

    #[test]
    fn test_zero_params_resolve_to_defaults() {
        // Zero means "use default", so an all-zero params value is valid.
        let ind = DirectionalTrendIndex::new(&DirectionalTrendIndexParams {
            q: 0,
            r: 0,
            s: 0,
            u: 0,
            ul: 0,
        })
        .unwrap();
        assert_eq!(ind.metadata().mnemonic, "dti(2,20,5,3,3)");
    }

    #[test]
    fn test_update_bar_ordering() {
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();
        let exp_dti = testdata::expected_q2_r20_s5_u3();
        let exp_signal = testdata::expected_q2_r20_s5_u3_sig_ul3();

        let mut ind = DirectionalTrendIndex::new(&DirectionalTrendIndexParams::default()).unwrap();

        let mut out: Output = Vec::new();
        for i in 0..high.len() {
            out = ind.update_bar(&Bar {
                time: 0,
                open: 0.0,
                high: high[i],
                low: low[i],
                close: 0.0,
                volume: 0.0,
            });
        }

        let last = high.len() - 1;
        assert_eq!(out.len(), 2);
        check("dti", last, exp_dti[last], scalar_value(&out, 0));
        check("signal", last, exp_signal[last], scalar_value(&out, 1));
    }

    #[test]
    fn test_update_scalar_uses_value_as_high_and_low() {
        let mut ind = passthrough(2);

        let first = ind.update_scalar(&Scalar {
            time: 0,
            value: 10.0,
        });
        assert!(scalar_value(&first, 0).is_nan());
        assert!(scalar_value(&first, 1).is_nan());

        // Rising value: HLM is the plain one-bar momentum.
        let out = ind.update_scalar(&Scalar {
            time: 0,
            value: 12.0,
        });
        assert_eq!(scalar_value(&out, 0), 100.0);
        assert_eq!(scalar_value(&out, 1), 100.0);
    }

    #[test]
    fn test_update_quote_uses_mid_price_as_high_and_low() {
        let mut ind = passthrough(2);

        ind.update_quote(&Quote {
            time: 0,
            bid_price: 12.0,
            ask_price: 14.0,
            bid_size: 1.0,
            ask_size: 1.0,
        });

        // Mid falls from 13 to 11.
        let out = ind.update_quote(&Quote {
            time: 0,
            bid_price: 10.0,
            ask_price: 12.0,
            bid_size: 1.0,
            ask_size: 1.0,
        });
        assert_eq!(scalar_value(&out, 0), -100.0);
        assert_eq!(scalar_value(&out, 1), -100.0);
    }

    #[test]
    fn test_update_trade_uses_price_as_high_and_low() {
        let mut ind = passthrough(2);

        ind.update_trade(&Trade {
            time: 0,
            price: 10.0,
            volume: 1.0,
        });
        let out = ind.update_trade(&Trade {
            time: 0,
            price: 12.0,
            volume: 1.0,
        });
        assert_eq!(scalar_value(&out, 0), 100.0);
        assert_eq!(scalar_value(&out, 1), 100.0);
    }
}
