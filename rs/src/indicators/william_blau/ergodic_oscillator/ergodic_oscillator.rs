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
use crate::indicators::william_blau::true_strength_index::true_strength_index::{
    TrueStrengthIndex, TrueStrengthIndexParams,
};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Ergodic Oscillator indicator.
///
/// The field names `q`, `r`, `s`, `u` and `ul` are the canonical symbols from
/// William Blau's *Momentum, Direction, and Divergence* (Wiley, 1995), chapter 2.
pub struct ErgodicOscillatorParams {
    /// Momentum look-back period; momentum is `C_k - C_(k-(q-1))`. Must be > 0. Default 2.
    pub q: usize,
    /// Period of the 1st (innermost) EMA, applied to the momentum. Must be > 0. Default 20.
    pub r: usize,
    /// Period of the 2nd EMA in the cascade. Must be > 0. Default 5.
    pub s: usize,
    /// Period of the 3rd (outermost) EMA in the cascade. Must be > 0. Default 3.
    pub u: usize,
    /// Period of the signal-line EMA (second output). Must be > 0. Default 3.
    /// `ul = 1` makes the signal a passthrough (signal == ergodic every bar).
    pub ul: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for ErgodicOscillatorParams {
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

/// Enumerates the outputs of the Ergodic Oscillator indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum ErgodicOscillatorOutput {
    /// The Ergodic oscillator value (the True Strength Index, range [-100, +100]).
    Ergodic = 1,
    /// The signal-line value: the ul-period EMA of the oscillator.
    Signal = 2,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// William Blau's Ergodic Oscillator.
///
/// The Ergodic is the True Strength Index plotted together with a signal line --
/// the EMA of the oscillator that Blau introduces as the trading vehicle for the
/// TSI (ch. 2, Fig. 2-14):
///
///   ergodic_k = TSI(q, r, s, u)_k                    (the oscillator)
///   signal_k  = EMA(ergodic, ul)_k                   (ul-period EMA of it)
///
/// The indicator produces two outputs:
///   - Ergodic: the oscillator, range [-100, +100], NaN during warm-up (bars 0..q-2);
///   - Signal: the ul-period EMA of the oscillator.
///
/// The numerics are exactly those of the True Strength Index, so this indicator
/// wraps a [`TrueStrengthIndex`] instance instead of duplicating the triple EMA
/// cascade. Only the mnemonic, the identifier, and the output naming differ.
pub struct ErgodicOscillator {
    tsi: TrueStrengthIndex,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
}

impl ErgodicOscillator {
    /// Creates a new Ergodic Oscillator from the given parameters.
    pub fn new(params: &ErgodicOscillatorParams) -> Result<Self, String> {
        let invalid = "invalid ergodic oscillator parameters";

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

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        // The oscillator and its signal line are exactly the True Strength Index
        // outputs; wrap an instance rather than duplicating its numerics. The
        // resolved components are passed through so both agree on the mnemonic.
        // Parameter validation is performed by the wrapped indicator.
        let tsi = TrueStrengthIndex::new(&TrueStrengthIndexParams {
            q,
            r,
            s,
            u,
            ul,
            bar_component: Some(bc),
            quote_component: Some(qc),
            trade_component: Some(tc),
        })
        .map_err(|err| format!("{}: {}", invalid, err))?;

        let mnemonic = format!(
            "ergodic({},{},{},{},{}{})",
            q,
            r,
            s,
            u,
            ul,
            component_triple_mnemonic(bc, qc, tc)
        );

        Ok(Self {
            tsi,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.tsi.is_primed()
    }

    /// Core update returning (ergodic, signal).
    pub fn update(&mut self, sample: f64) -> (f64, f64) {
        self.tsi.update(sample)
    }
}

impl Indicator for ErgodicOscillator {
    fn is_primed(&self) -> bool {
        self.tsi.is_primed()
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Ergodic Oscillator {}", self.mnemonic);
        build_metadata(
            Identifier::ErgodicOscillator,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} ergodic", self.mnemonic),
                    description: format!("{} ergodic", desc),
                },
                OutputText {
                    mnemonic: format!("{} signal", self.mnemonic),
                    description: format!("{} signal", desc),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (ergodic, signal) = self.update(sample.value);
        vec![
            Box::new(Scalar { time: sample.time, value: ergodic }),
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
    use super::super::testdata;
    use super::*;

    const TOLERANCE: f64 = 1e-9;

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
        ul: usize,
        ergodic: Vec<f64>,
        signal: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { q: 2, r: 20, s: 5, u: 3, ul: 3, ergodic: testdata::expected_erg_q2_r20_s5_u3_l3(), signal: testdata::expected_sig_q2_r20_s5_u3_l3() },
            Combo { q: 2, r: 32, s: 5, u: 1, ul: 5, ergodic: testdata::expected_erg_q2_r32_s5_u1_l5(), signal: testdata::expected_sig_q2_r32_s5_u1_l5() },
            Combo { q: 2, r: 20, s: 5, u: 1, ul: 5, ergodic: testdata::expected_erg_q2_r20_s5_u1_l5(), signal: testdata::expected_sig_q2_r20_s5_u1_l5() },
            Combo { q: 2, r: 32, s: 5, u: 1, ul: 7, ergodic: testdata::expected_erg_q2_r32_s5_u1_l7(), signal: testdata::expected_sig_q2_r32_s5_u1_l7() },
            Combo { q: 2, r: 25, s: 13, u: 1, ul: 5, ergodic: testdata::expected_erg_q2_r25_s13_u1_l5(), signal: testdata::expected_sig_q2_r25_s13_u1_l5() },
            Combo { q: 2, r: 20, s: 5, u: 3, ul: 1, ergodic: testdata::expected_erg_q2_r20_s5_u3_l1(), signal: testdata::expected_sig_q2_r20_s5_u3_l1() },
            Combo { q: 2, r: 1, s: 1, u: 1, ul: 1, ergodic: testdata::expected_erg_q2_r1_s1_u1_l1(), signal: testdata::expected_sig_q2_r1_s1_u1_l1() },
            Combo { q: 2, r: 20, s: 5, u: 3, ul: 9, ergodic: testdata::expected_erg_q2_r20_s5_u3_l9(), signal: testdata::expected_sig_q2_r20_s5_u3_l9() },
            Combo { q: 2, r: 64, s: 64, u: 1, ul: 5, ergodic: testdata::expected_erg_q2_r64_s64_u1_l5(), signal: testdata::expected_sig_q2_r64_s64_u1_l5() },
            Combo { q: 2, r: 9, s: 3, u: 1, ul: 3, ergodic: testdata::expected_erg_q2_r9_s3_u1_l3(), signal: testdata::expected_sig_q2_r9_s3_u1_l3() },
            Combo { q: 3, r: 20, s: 5, u: 3, ul: 3, ergodic: testdata::expected_erg_q3_r20_s5_u3_l3(), signal: testdata::expected_sig_q3_r20_s5_u3_l3() },
            Combo { q: 5, r: 20, s: 5, u: 3, ul: 5, ergodic: testdata::expected_erg_q5_r20_s5_u3_l5(), signal: testdata::expected_sig_q5_r20_s5_u3_l5() },
            Combo { q: 2, r: 13, s: 7, u: 1, ul: 7, ergodic: testdata::expected_erg_q2_r13_s7_u1_l7(), signal: testdata::expected_sig_q2_r13_s7_u1_l7() },
            Combo { q: 2, r: 20, s: 5, u: 1, ul: 3, ergodic: testdata::expected_erg_q2_r20_s5_u1_l3(), signal: testdata::expected_sig_q2_r20_s5_u1_l3() },
        ];

        let input = testdata::testdata::test_input();

        for combo in &combos {
            let mut ind = ErgodicOscillator::new(&ErgodicOscillatorParams {
                q: combo.q,
                r: combo.r,
                s: combo.s,
                u: combo.u,
                ul: combo.ul,
                ..Default::default()
            })
            .unwrap();

            for i in 0..input.len() {
                let (ergodic, signal) = ind.update(input[i]);
                check("ergodic", i, combo.ergodic[i], ergodic);
                check("signal", i, combo.signal[i], signal);
            }
        }
    }

    #[test]
    fn test_signal_passthrough_when_ul_is_one() {
        let input = testdata::testdata::test_input();

        let mut ind = ErgodicOscillator::new(&ErgodicOscillatorParams {
            q: 2,
            r: 20,
            s: 5,
            u: 3,
            ul: 1,
            ..Default::default()
        })
        .unwrap();

        for i in 0..input.len() {
            let (ergodic, signal) = ind.update(input[i]);
            if ergodic.is_nan() {
                assert!(signal.is_nan(), "signal[{}]: expected NaN, got {}", i, signal);
            } else {
                assert_eq!(ergodic, signal, "signal[{}] must equal ergodic when ul = 1", i);
            }
        }
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::testdata::test_input();

        let mut ind = ErgodicOscillator::new(&ErgodicOscillatorParams {
            q: 3,
            ..Default::default()
        })
        .unwrap();

        assert!(!ind.is_primed());
        ind.update(input[0]);
        assert!(!ind.is_primed());
        ind.update(input[1]);
        assert!(!ind.is_primed());
        ind.update(input[2]);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_mnemonic() {
        let ind = ErgodicOscillator::new(&ErgodicOscillatorParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "ergodic(2,20,5,3,3)");

        let ind2 = ErgodicOscillator::new(&ErgodicOscillatorParams {
            q: 2,
            r: 25,
            s: 13,
            u: 1,
            ul: 7,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "ergodic(2,25,13,1,7)");
    }

    #[test]
    fn test_metadata() {
        let ind = ErgodicOscillator::new(&ErgodicOscillatorParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::ErgodicOscillator);
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(meta.outputs[0].kind, ErgodicOscillatorOutput::Ergodic as i32);
        assert_eq!(meta.outputs[1].kind, ErgodicOscillatorOutput::Signal as i32);
    }

    #[test]
    fn test_update_scalar_ordering() {
        let input = testdata::testdata::test_input();
        let exp_ergodic = testdata::expected_erg_q2_r20_s5_u3_l3();
        let exp_signal = testdata::expected_sig_q2_r20_s5_u3_l3();

        let mut ind = ErgodicOscillator::new(&ErgodicOscillatorParams::default()).unwrap();
        let mut out: Output = vec![];
        for i in 0..input.len() {
            out = ind.update_scalar(&Scalar { time: 0, value: input[i] });
        }
        let last = input.len() - 1;
        let ergodic = out[0].downcast_ref::<Scalar>().unwrap().value;
        let signal = out[1].downcast_ref::<Scalar>().unwrap().value;
        check("ergodic", last, exp_ergodic[last], ergodic);
        check("signal", last, exp_signal[last], signal);
    }

    #[test]
    fn test_nan_warmup() {
        let mut ind = ErgodicOscillator::new(&ErgodicOscillatorParams::default()).unwrap();
        let (ergodic, signal) = ind.update(f64::NAN);
        assert!(ergodic.is_nan());
        assert!(signal.is_nan());
    }
}
