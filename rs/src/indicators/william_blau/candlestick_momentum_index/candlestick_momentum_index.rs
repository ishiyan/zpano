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

/// Parameters to create an instance of the Candlestick Momentum Index indicator.
///
/// The field names `r`, `s`, `u` and `ul` are the canonical symbols from William
/// Blau's *Momentum, Direction, and Divergence* (Wiley, 1995), chapter 6.
///
/// The indicator consumes the open and close prices of a bar, so it has no
/// configurable price-component fields.
pub struct CandlestickMomentumIndexParams {
    /// Period of the 1st (innermost) EMA, applied to the candle momentum. Must be > 0. Default 20.
    pub r: usize,
    /// Period of the 2nd EMA in the cascade. Must be > 0. Default 5.
    pub s: usize,
    /// Period of the 3rd (outermost) EMA in the cascade. Must be > 0. Default 3.
    pub u: usize,
    /// Period of the signal-line EMA (second output). Must be > 0. Default 3.
    pub ul: usize,
}

impl Default for CandlestickMomentumIndexParams {
    fn default() -> Self {
        Self {
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

/// Enumerates the outputs of the Candlestick Momentum Index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CandlestickMomentumIndexOutput {
    /// The Candlestick Momentum Index oscillator value (range [-100, +100]).
    Cmi = 1,
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

/// William Blau's Candlestick Momentum Index (CMI).
///
/// A double-/triple-smoothed intra-bar momentum oscillator bounded to [-100, +100],
/// paired with an EMA signal line (the Ergodic form, Blau ch.6.4):
///
///   cmi_k    = 100 * TEMA(cmtm, r, s, u) / TEMA(|cmtm|, r, s, u)   (the oscillator)
///   signal_k = EMA(cmi, ul)_k                                      (ul-period EMA)
///
/// where the candle momentum is the signed candle body cmtm_k = close_k - open_k
/// and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
///
/// This is the True Strength Index structure applied to the candle body instead of
/// price momentum. Because it only looks inside each bar it is immune to inter-bar
/// gaps. The inputs are the open and close prices only.
///
/// The indicator produces two outputs:
///   - CMI: the oscillator, range [-100, +100];
///   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
///
/// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
/// received value. The candle momentum is defined from bar 0, so there is no NaN
/// warm-up region. Division guard: denominator 0 -> oscillator 0.0.
pub struct CandlestickMomentumIndex {
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

impl CandlestickMomentumIndex {
    /// Creates a new Candlestick Momentum Index from the given parameters.
    pub fn new(params: &CandlestickMomentumIndexParams) -> Result<Self, String> {
        let invalid = "invalid candlestick momentum index parameters";

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

        let mnemonic = format!("cmi({},{},{},{})", r, s, u, ul);

        Ok(Self {
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

    /// Core update taking one bar's open and close, returning (cmi, signal).
    pub fn update(&mut self, open: f64, close: f64) -> (f64, f64) {
        // Candle momentum: the signed body of the candle.
        let cmtm = close - open;
        let abs_cmtm = cmtm.abs();

        // Numerator cascade: TEMA(cmtm, r, s, u).
        let n = self.num_u.update(self.num_s.update(self.num_r.update(cmtm)));
        // Denominator cascade: TEMA(|cmtm|, r, s, u).
        let d = self.den_u.update(self.den_s.update(self.den_r.update(abs_cmtm)));

        // Division guard (Blau_CMI.mq5): denominator 0 -> oscillator 0.0.
        let cmi = if d == 0.0 { 0.0 } else { 100.0 * n / d };

        // Signal line = EMA(cmi, ul); seeds on bar 0's oscillator value.
        let signal = self.signal_ema.update(cmi);
        self.primed = true;

        (cmi, signal)
    }

    /// Updates the indicator and wraps the two outputs.
    fn update_entity(&mut self, time: i64, open: f64, close: f64) -> Output {
        let (cmi, signal) = self.update(open, close);
        vec![
            Box::new(Scalar { time, value: cmi }),
            Box::new(Scalar { time, value: signal }),
        ]
    }
}

impl Indicator for CandlestickMomentumIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Candlestick Momentum Index {}", self.mnemonic);
        build_metadata(
            Identifier::CandlestickMomentumIndex,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} cmi", self.mnemonic),
                    description: format!("{} CMI", desc),
                },
                OutputText {
                    mnemonic: format!("{} signal", self.mnemonic),
                    description: format!("{} signal", desc),
                },
            ],
        )
    }

    /// A scalar carries a single value, so open == close and the candle momentum is zero.
    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        self.update_entity(sample.time, sample.open, sample.close)
    }

    /// A quote maps the bid to the open and the ask to the close, so the candle body is the spread.
    fn update_quote(&mut self, sample: &Quote) -> Output {
        self.update_entity(sample.time, sample.bid_price, sample.ask_price)
    }

    /// A trade carries a single price, so open == close and the candle momentum is zero.
    fn update_trade(&mut self, sample: &Trade) -> Output {
        self.update_entity(sample.time, sample.price, sample.price)
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
        r: usize,
        s: usize,
        u: usize,
        cmi: Vec<f64>,
        signal: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { r: 20, s: 5, u: 3, cmi: testdata::expected_r20_s5_u3(), signal: testdata::expected_r20_s5_u3_sig_ul3() },
            Combo { r: 20, s: 5, u: 1, cmi: testdata::expected_r20_s5_u1(), signal: testdata::expected_r20_s5_u1_sig_ul3() },
            Combo { r: 1, s: 1, u: 1, cmi: testdata::expected_r1_s1_u1(), signal: testdata::expected_r1_s1_u1_sig_ul3() },
            Combo { r: 25, s: 13, u: 1, cmi: testdata::expected_r25_s13_u1(), signal: testdata::expected_r25_s13_u1_sig_ul3() },
            Combo { r: 13, s: 13, u: 1, cmi: testdata::expected_r13_s13_u1(), signal: testdata::expected_r13_s13_u1_sig_ul3() },
            Combo { r: 5, s: 5, u: 5, cmi: testdata::expected_r5_s5_u5(), signal: testdata::expected_r5_s5_u5_sig_ul3() },
            Combo { r: 9, s: 3, u: 1, cmi: testdata::expected_r9_s3_u1(), signal: testdata::expected_r9_s3_u1_sig_ul3() },
            Combo { r: 64, s: 64, u: 1, cmi: testdata::expected_r64_s64_u1(), signal: testdata::expected_r64_s64_u1_sig_ul3() },
            Combo { r: 32, s: 5, u: 3, cmi: testdata::expected_r32_s5_u3(), signal: testdata::expected_r32_s5_u3_sig_ul3() },
            Combo { r: 40, s: 20, u: 1, cmi: testdata::expected_r40_s20_u1(), signal: testdata::expected_r40_s20_u1_sig_ul3() },
            Combo { r: 2, s: 2, u: 2, cmi: testdata::expected_r2_s2_u2(), signal: testdata::expected_r2_s2_u2_sig_ul3() },
            Combo { r: 7, s: 4, u: 2, cmi: testdata::expected_r7_s4_u2(), signal: testdata::expected_r7_s4_u2_sig_ul3() },
            Combo { r: 12, s: 12, u: 12, cmi: testdata::expected_r12_s12_u12(), signal: testdata::expected_r12_s12_u12_sig_ul3() },
            Combo { r: 3, s: 10, u: 10, cmi: testdata::expected_r3_s10_u10(), signal: testdata::expected_r3_s10_u10_sig_ul3() },
            Combo { r: 50, s: 1, u: 1, cmi: testdata::expected_r50_s1_u1(), signal: testdata::expected_r50_s1_u1_sig_ul3() },
            Combo { r: 20, s: 5, u: 5, cmi: testdata::expected_r20_s5_u5(), signal: testdata::expected_r20_s5_u5_sig_ul3() },
        ];

        let input = testdata::testdata::test_input();
        let open = testdata::testdata::test_open();

        for combo in &combos {
            let mut ind = CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams {
                r: combo.r,
                s: combo.s,
                u: combo.u,
                ul: UL,
            })
            .unwrap();

            for i in 0..input.len() {
                let (cmi, signal) = ind.update(open[i], input[i]);
                check("cmi", i, combo.cmi[i], cmi);
                check("signal", i, combo.signal[i], signal);
            }
        }
    }

    #[test]
    fn test_passthrough() {
        let mut ind = CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams {
            r: 1,
            s: 1,
            u: 1,
            ul: 1,
        })
        .unwrap();

        assert_eq!(ind.update(10.0, 12.0), (100.0, 100.0)); // up candle
        assert_eq!(ind.update(12.0, 11.0), (-100.0, -100.0)); // down candle
        assert_eq!(ind.update(11.0, 11.0).0, 0.0); // doji -> division guard
    }

    #[test]
    fn test_is_primed_after_first_bar() {
        let input = testdata::testdata::test_input();
        let open = testdata::testdata::test_open();

        let mut ind =
            CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams::default()).unwrap();
        assert!(!CandlestickMomentumIndex::is_primed(&ind));

        ind.update(open[0], input[0]);
        assert!(CandlestickMomentumIndex::is_primed(&ind));
    }

    #[test]
    fn test_mnemonic() {
        let ind =
            CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "cmi(20,5,3,3)");
        assert_eq!(
            ind.metadata().description,
            "Candlestick Momentum Index cmi(20,5,3,3)"
        );

        let ind2 = CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams {
            r: 25,
            s: 13,
            u: 1,
            ul: 7,
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "cmi(25,13,1,7)");
    }

    #[test]
    fn test_metadata() {
        let ind =
            CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::CandlestickMomentumIndex);
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(
            meta.outputs[0].kind,
            CandlestickMomentumIndexOutput::Cmi as i32
        );
        assert_eq!(
            meta.outputs[1].kind,
            CandlestickMomentumIndexOutput::Signal as i32
        );
    }

    #[test]
    fn test_update_bar_ordering() {
        let input = testdata::testdata::test_input();
        let open = testdata::testdata::test_open();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();
        let exp_cmi = testdata::expected_r20_s5_u3();
        let exp_signal = testdata::expected_r20_s5_u3_sig_ul3();

        let mut ind = CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams {
            r: 20,
            s: 5,
            u: 3,
            ul: UL,
        })
        .unwrap();

        let mut out: Output = Vec::new();
        for i in 0..input.len() {
            out = ind.update_bar(&Bar {
                time: 0,
                open: open[i],
                high: high[i],
                low: low[i],
                close: input[i],
                volume: 0.0,
            });
        }

        let last = input.len() - 1;
        let cmi = out[0].downcast_ref::<Scalar>().unwrap().value;
        let signal = out[1].downcast_ref::<Scalar>().unwrap().value;
        check("cmi", last, exp_cmi[last], cmi);
        check("signal", last, exp_signal[last], signal);
    }

    #[test]
    fn test_update_quote_maps_bid_to_open_and_ask_to_close() {
        let mut ind = CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams {
            r: 1,
            s: 1,
            u: 1,
            ul: 1,
        })
        .unwrap();

        let out = ind.update_quote(&Quote {
            time: 0,
            bid_price: 10.0,
            ask_price: 12.0,
            bid_size: 1.0,
            ask_size: 1.0,
        });

        assert_eq!(out[0].downcast_ref::<Scalar>().unwrap().value, 100.0);
        assert_eq!(out[1].downcast_ref::<Scalar>().unwrap().value, 100.0);
    }

    #[test]
    fn test_update_scalar_and_trade_have_zero_candle_body() {
        let mut scalar_ind = CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams {
            r: 1,
            s: 1,
            u: 1,
            ul: 1,
        })
        .unwrap();
        let scalar_out = scalar_ind.update_scalar(&Scalar {
            time: 0,
            value: 10.0,
        });
        assert_eq!(scalar_out[0].downcast_ref::<Scalar>().unwrap().value, 0.0);
        assert_eq!(scalar_out[1].downcast_ref::<Scalar>().unwrap().value, 0.0);

        let mut trade_ind = CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams {
            r: 1,
            s: 1,
            u: 1,
            ul: 1,
        })
        .unwrap();
        let trade_out = trade_ind.update_trade(&Trade {
            time: 0,
            price: 10.0,
            volume: 1.0,
        });
        assert_eq!(trade_out[0].downcast_ref::<Scalar>().unwrap().value, 0.0);
        assert_eq!(trade_out[1].downcast_ref::<Scalar>().unwrap().value, 0.0);
    }
}
