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

/// Parameters to create an instance of the True Strength Index indicator.
///
/// The field names `q`, `r`, `s`, `u` and `ul` are the canonical symbols from
/// William Blau's *Momentum, Direction, and Divergence* (Wiley, 1995), chapter 2.
pub struct TrueStrengthIndexParams {
    /// Momentum look-back period; momentum is `C_k - C_(k-(q-1))`. Must be > 0. Default 2.
    pub q: usize,
    /// Period of the 1st (innermost) EMA, applied to the momentum. Must be > 0. Default 20.
    pub r: usize,
    /// Period of the 2nd EMA in the cascade. Must be > 0. Default 5.
    pub s: usize,
    /// Period of the 3rd (outermost) EMA in the cascade. Must be > 0. Default 3.
    pub u: usize,
    /// Period of the signal-line EMA (second output). Must be > 0. Default 3.
    /// Not shown in the indicator mnemonic.
    pub ul: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for TrueStrengthIndexParams {
    fn default() -> Self {
        Self {
            q: 2,
            r: 20,
            s: 5,
            u: 3,
            ul: 3,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the True Strength Index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum TrueStrengthIndexOutput {
    /// The True Strength Index oscillator value (range [-100, +100]).
    Tsi = 1,
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

/// William Blau's True Strength Index (TSI).
///
/// A double-/triple-smoothed momentum oscillator bounded to [-100, +100], paired
/// with an EMA signal line (the Ergodic form, Blau ch.1.4):
///
///   tsi_k    = 100 * TEMA(mtm, r, s, u) / TEMA(|mtm|, r, s, u)   (the oscillator)
///   signal_k = EMA(tsi, ul)_k                                    (ul-period EMA)
///
/// where mtm_k = C_k - C_(k-(q-1)) and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
///
/// The indicator produces two outputs:
///   - TSI: the oscillator, range [-100, +100], NaN during warm-up (bars 0..q-2);
///   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
pub struct TrueStrengthIndex {
    q: usize,
    history: Vec<f64>,
    num_r: Ema,
    num_s: Ema,
    num_u: Ema,
    den_r: Ema,
    den_s: Ema,
    den_u: Ema,
    signal_ema: Ema,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
}

impl TrueStrengthIndex {
    /// Creates a new True Strength Index from the given parameters.
    pub fn new(params: &TrueStrengthIndexParams) -> Result<Self, String> {
        let invalid = "invalid true strength index parameters";

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

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "tsi({},{},{},{}{})",
            q,
            r,
            s,
            u,
            component_triple_mnemonic(bc, qc, tc)
        );

        Ok(Self {
            q,
            history: Vec::with_capacity(q),
            num_r: Ema::new(r),
            num_s: Ema::new(s),
            num_u: Ema::new(u),
            den_r: Ema::new(r),
            den_s: Ema::new(s),
            den_u: Ema::new(u),
            signal_ema: Ema::new(ul),
            primed: false,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update returning (tsi, signal).
    pub fn update(&mut self, sample: f64) -> (f64, f64) {
        // Maintain a rolling window of the last q prices; the leftmost element is
        // C_(k-(q-1)).
        if self.history.len() < self.q {
            self.history.push(sample);
        } else {
            self.history.remove(0);
            self.history.push(sample);
        }

        // Momentum needs a price from q-1 bars ago, available only once the window
        // holds q prices. Before then neither output is defined and the signal EMA
        // is NOT advanced.
        if self.history.len() < self.q {
            return (f64::NAN, f64::NAN);
        }

        // mtm_k = C_k - C_(k-(q-1)); the leftmost history element is C_(k-(q-1)).
        let mtm = sample - self.history[0];
        let abs_mtm = mtm.abs();

        // Numerator cascade: TEMA(mtm, r, s, u).
        let n = self.num_u.update(self.num_s.update(self.num_r.update(mtm)));
        // Denominator cascade: TEMA(|mtm|, r, s, u).
        let d = self.den_u.update(self.den_s.update(self.den_r.update(abs_mtm)));

        // Division guard (Blau_TSI.mq5): denominator 0 -> oscillator 0.0.
        let tsi = if d == 0.0 { 0.0 } else { 100.0 * n / d };

        // Signal line = EMA(tsi, ul); seeds here on the first finite oscillator.
        let signal = self.signal_ema.update(tsi);
        self.primed = true;

        (tsi, signal)
    }
}

impl Indicator for TrueStrengthIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("True Strength Index {}", self.mnemonic);
        build_metadata(
            Identifier::TrueStrengthIndex,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} tsi", self.mnemonic),
                    description: format!("{} TSI", desc),
                },
                OutputText {
                    mnemonic: format!("{} signal", self.mnemonic),
                    description: format!("{} signal", desc),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (tsi, signal) = self.update(sample.value);
        vec![
            Box::new(Scalar { time: sample.time, value: tsi }),
            Box::new(Scalar { time: sample.time, value: signal }),
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
    use super::super::testdata;

    const TOLERANCE: f64 = 1e-9;

    // Signal-line EMA period used for every expected signal array.
    const UL: usize = 3;

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

    struct Combo {
        q: usize,
        r: usize,
        s: usize,
        u: usize,
        tsi: Vec<f64>,
        signal: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { q: 2, r: 20, s: 5, u: 3, tsi: testdata::expected_q2_r20_s5_u3(), signal: testdata::expected_q2_r20_s5_u3_sig_ul3() },
            Combo { q: 2, r: 25, s: 13, u: 1, tsi: testdata::expected_q2_r25_s13_u1(), signal: testdata::expected_q2_r25_s13_u1_sig_ul3() },
            Combo { q: 2, r: 20, s: 5, u: 1, tsi: testdata::expected_q2_r20_s5_u1(), signal: testdata::expected_q2_r20_s5_u1_sig_ul3() },
            Combo { q: 2, r: 32, s: 5, u: 1, tsi: testdata::expected_q2_r32_s5_u1(), signal: testdata::expected_q2_r32_s5_u1_sig_ul3() },
            Combo { q: 2, r: 13, s: 13, u: 1, tsi: testdata::expected_q2_r13_s13_u1(), signal: testdata::expected_q2_r13_s13_u1_sig_ul3() },
            Combo { q: 2, r: 20, s: 40, u: 1, tsi: testdata::expected_q2_r20_s40_u1(), signal: testdata::expected_q2_r20_s40_u1_sig_ul3() },
            Combo { q: 2, r: 40, s: 20, u: 1, tsi: testdata::expected_q2_r40_s20_u1(), signal: testdata::expected_q2_r40_s20_u1_sig_ul3() },
            Combo { q: 2, r: 64, s: 64, u: 1, tsi: testdata::expected_q2_r64_s64_u1(), signal: testdata::expected_q2_r64_s64_u1_sig_ul3() },
            Combo { q: 2, r: 100, s: 5, u: 1, tsi: testdata::expected_q2_r100_s5_u1(), signal: testdata::expected_q2_r100_s5_u1_sig_ul3() },
            Combo { q: 2, r: 1, s: 1, u: 1, tsi: testdata::expected_q2_r1_s1_u1(), signal: testdata::expected_q2_r1_s1_u1_sig_ul3() },
            Combo { q: 2, r: 1, s: 5, u: 3, tsi: testdata::expected_q2_r1_s5_u3(), signal: testdata::expected_q2_r1_s5_u3_sig_ul3() },
            Combo { q: 2, r: 20, s: 1, u: 1, tsi: testdata::expected_q2_r20_s1_u1(), signal: testdata::expected_q2_r20_s1_u1_sig_ul3() },
            Combo { q: 2, r: 5, s: 5, u: 5, tsi: testdata::expected_q2_r5_s5_u5(), signal: testdata::expected_q2_r5_s5_u5_sig_ul3() },
            Combo { q: 3, r: 20, s: 5, u: 3, tsi: testdata::expected_q3_r20_s5_u3(), signal: testdata::expected_q3_r20_s5_u3_sig_ul3() },
            Combo { q: 5, r: 20, s: 5, u: 3, tsi: testdata::expected_q5_r20_s5_u3(), signal: testdata::expected_q5_r20_s5_u3_sig_ul3() },
            Combo { q: 10, r: 20, s: 5, u: 1, tsi: testdata::expected_q10_r20_s5_u1(), signal: testdata::expected_q10_r20_s5_u1_sig_ul3() },
            Combo { q: 2, r: 9, s: 3, u: 1, tsi: testdata::expected_q2_r9_s3_u1(), signal: testdata::expected_q2_r9_s3_u1_sig_ul3() },
            Combo { q: 2, r: 7, s: 4, u: 2, tsi: testdata::expected_q2_r7_s4_u2(), signal: testdata::expected_q2_r7_s4_u2_sig_ul3() },
        ];

        let input = testdata::testdata::test_input();

        for combo in &combos {
            let mut ind = TrueStrengthIndex::new(&TrueStrengthIndexParams {
                q: combo.q,
                r: combo.r,
                s: combo.s,
                u: combo.u,
                ul: UL,
                ..Default::default()
            })
            .unwrap();

            for i in 0..input.len() {
                let (tsi, signal) = ind.update(input[i]);
                check("tsi", i, combo.tsi[i], tsi);
                check("signal", i, combo.signal[i], signal);
            }
        }
    }

    #[test]
    fn test_mnemonic_excludes_ul() {
        let ind = TrueStrengthIndex::new(&TrueStrengthIndexParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "tsi(2,20,5,3)");

        let ind2 = TrueStrengthIndex::new(&TrueStrengthIndexParams {
            q: 2,
            r: 25,
            s: 13,
            u: 1,
            ul: 7,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "tsi(2,25,13,1)");
    }

    #[test]
    fn test_metadata() {
        let ind = TrueStrengthIndex::new(&TrueStrengthIndexParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::TrueStrengthIndex);
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(meta.outputs[0].kind, TrueStrengthIndexOutput::Tsi as i32);
        assert_eq!(meta.outputs[1].kind, TrueStrengthIndexOutput::Signal as i32);
    }

    #[test]
    fn test_update_scalar_ordering() {
        let input = testdata::testdata::test_input();
        let exp_tsi = testdata::expected_q2_r20_s5_u3();
        let exp_signal = testdata::expected_q2_r20_s5_u3_sig_ul3();

        let mut ind = TrueStrengthIndex::new(&TrueStrengthIndexParams {
            ul: UL,
            ..Default::default()
        })
        .unwrap();
        let mut out: Output = vec![];
        for i in 0..input.len() {
            out = ind.update_scalar(&Scalar { time: 0, value: input[i] });
        }
        let last = input.len() - 1;
        let tsi = out[0].downcast_ref::<Scalar>().unwrap().value;
        let signal = out[1].downcast_ref::<Scalar>().unwrap().value;
        check("tsi", last, exp_tsi[last], tsi);
        check("signal", last, exp_signal[last], signal);
    }

    #[test]
    fn test_invalid_params() {
        // q/r/s/u/ul = 0 are treated as defaults (matching other languages), so
        // they are not invalid. There is no out-of-range invalid case for TSI
        // beyond zero, so this test simply confirms default construction works.
        assert!(TrueStrengthIndex::new(&TrueStrengthIndexParams::default()).is_ok());
    }

    #[test]
    fn test_nan_warmup() {
        let mut ind = TrueStrengthIndex::new(&TrueStrengthIndexParams::default()).unwrap();
        let (tsi, signal) = ind.update(f64::NAN);
        assert!(tsi.is_nan());
        assert!(signal.is_nan());
    }
}
