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

/// Parameters to create an instance of the Candlestick Strength Index indicator.
///
/// The field names `r`, `s`, `u` and `ul` are the canonical symbols from William
/// Blau's *Momentum, Direction, and Divergence* (Wiley, 1995), chapter 6.
///
/// The indicator consumes the open, high, low and close prices of a bar, so it has
/// no configurable price-component fields.
pub struct CandlestickStrengthIndexParams {
    /// Period of the 1st (innermost) EMA, applied to the candle body and the bar range. Must be > 0. Default 20.
    pub r: usize,
    /// Period of the 2nd EMA in the cascade. Must be > 0. Default 5.
    pub s: usize,
    /// Period of the 3rd (outermost) EMA in the cascade. Must be > 0. Default 3.
    pub u: usize,
    /// Period of the signal-line EMA (second output). Must be > 0. Default 3.
    pub ul: usize,
}

impl Default for CandlestickStrengthIndexParams {
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

/// Enumerates the outputs of the Candlestick Strength Index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CandlestickStrengthIndexOutput {
    /// The Candlestick Strength Index oscillator value (range [-100, +100]).
    Csi = 1,
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

/// William Blau's Candlestick Strength Index (CSI), known in the book as the
/// CandleStick Indicator.
///
/// A double-/triple-smoothed candle-body-vs-range oscillator bounded to [-100, +100],
/// paired with an EMA signal line (the Ergodic form, Blau ch.6.4):
///
///   csi_k    = 100 * TEMA(close-open, r, s, u) / TEMA(high-low, r, s, u)   (the oscillator)
///   signal_k = EMA(csi, ul)_k                                              (ul-period EMA)
///
/// where the two intra-bar quantities are the signed candle body co_k = close_k - open_k
/// and the bar range hl_k = high_k - low_k >= 0, and TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
///
/// It is the range-normalized sibling of the Candlestick Momentum Index (CMI): both share
/// the signed numerator TEMA(close-open), but the CSI divides by the smoothed range while
/// the CMI divides by the smoothed absolute body. Because every bar has |close-open| <= high-low,
/// the ratio is bounded to [-100, +100]. The inputs are the open, high, low and close prices.
///
/// The indicator produces two outputs:
///   - CSI: the oscillator, range [-100, +100];
///   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
///
/// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
/// received value. Both intra-bar series are defined from bar 0, so there is no NaN
/// warm-up region. Division guard: denominator <= 0 -> oscillator 0.0.
pub struct CandlestickStrengthIndex {
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

impl CandlestickStrengthIndex {
    /// Creates a new Candlestick Strength Index from the given parameters.
    pub fn new(params: &CandlestickStrengthIndexParams) -> Result<Self, String> {
        let invalid = "invalid candlestick strength index parameters";

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

        let mnemonic = format!("csi({},{},{},{})", r, s, u, ul);

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

    /// Core update taking one bar's open, high, low and close, returning (csi, signal).
    pub fn update(&mut self, open: f64, high: f64, low: f64, close: f64) -> (f64, f64) {
        // Two intra-bar quantities: the signed candle body and the (non-negative) range.
        let co = close - open;
        let hl = high - low;

        // Numerator cascade: TEMA(close-open, r, s, u).
        let n = self.num_u.update(self.num_s.update(self.num_r.update(co)));
        // Denominator cascade: TEMA(high-low, r, s, u).
        let d = self.den_u.update(self.den_s.update(self.den_r.update(hl)));

        // Division guard: zero range so far -> oscillator 0.0.
        let csi = if d > 0.0 { 100.0 * n / d } else { 0.0 };

        // Signal line = EMA(csi, ul); seeds on bar 0's oscillator value.
        let signal = self.signal_ema.update(csi);
        self.primed = true;

        (csi, signal)
    }

    /// Updates the indicator and wraps the two outputs.
    fn update_entity(&mut self, time: i64, open: f64, high: f64, low: f64, close: f64) -> Output {
        let (csi, signal) = self.update(open, high, low, close);
        vec![
            Box::new(Scalar { time, value: csi }),
            Box::new(Scalar {
                time,
                value: signal,
            }),
        ]
    }
}

impl Indicator for CandlestickStrengthIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Candlestick Strength Index {}", self.mnemonic);
        build_metadata(
            Identifier::CandlestickStrengthIndex,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} csi", self.mnemonic),
                    description: format!("{} CSI", desc),
                },
                OutputText {
                    mnemonic: format!("{} signal", self.mnemonic),
                    description: format!("{} signal", desc),
                },
            ],
        )
    }

    /// A scalar carries a single value, so the candle body and the bar range are both zero.
    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(
            sample.time,
            sample.value,
            sample.value,
            sample.value,
            sample.value,
        )
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        self.update_entity(
            sample.time,
            sample.open,
            sample.high,
            sample.low,
            sample.close,
        )
    }

    /// A quote maps the bid to the open and the low, and the ask to the close and the high,
    /// so the candle body and the bar range are both the spread.
    fn update_quote(&mut self, sample: &Quote) -> Output {
        self.update_entity(
            sample.time,
            sample.bid_price,
            sample.ask_price,
            sample.bid_price,
            sample.ask_price,
        )
    }

    /// A trade carries a single price, so the candle body and the bar range are both zero.
    fn update_trade(&mut self, sample: &Trade) -> Output {
        self.update_entity(
            sample.time,
            sample.price,
            sample.price,
            sample.price,
            sample.price,
        )
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
        csi: Vec<f64>,
        signal: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { r: 20, s: 5, u: 3, csi: testdata::expected_r20_s5_u3(), signal: testdata::expected_r20_s5_u3_sig_ul3() },
            Combo { r: 32, s: 32, u: 1, csi: testdata::expected_r32_s32_u1(), signal: testdata::expected_r32_s32_u1_sig_ul3() },
            Combo { r: 1, s: 1, u: 1, csi: testdata::expected_r1_s1_u1(), signal: testdata::expected_r1_s1_u1_sig_ul3() },
            Combo { r: 25, s: 13, u: 1, csi: testdata::expected_r25_s13_u1(), signal: testdata::expected_r25_s13_u1_sig_ul3() },
            Combo { r: 13, s: 13, u: 1, csi: testdata::expected_r13_s13_u1(), signal: testdata::expected_r13_s13_u1_sig_ul3() },
            Combo { r: 5, s: 5, u: 5, csi: testdata::expected_r5_s5_u5(), signal: testdata::expected_r5_s5_u5_sig_ul3() },
            Combo { r: 9, s: 3, u: 1, csi: testdata::expected_r9_s3_u1(), signal: testdata::expected_r9_s3_u1_sig_ul3() },
            Combo { r: 64, s: 64, u: 1, csi: testdata::expected_r64_s64_u1(), signal: testdata::expected_r64_s64_u1_sig_ul3() },
            Combo { r: 32, s: 32, u: 3, csi: testdata::expected_r32_s32_u3(), signal: testdata::expected_r32_s32_u3_sig_ul3() },
            Combo { r: 40, s: 20, u: 1, csi: testdata::expected_r40_s20_u1(), signal: testdata::expected_r40_s20_u1_sig_ul3() },
            Combo { r: 2, s: 2, u: 2, csi: testdata::expected_r2_s2_u2(), signal: testdata::expected_r2_s2_u2_sig_ul3() },
            Combo { r: 7, s: 4, u: 2, csi: testdata::expected_r7_s4_u2(), signal: testdata::expected_r7_s4_u2_sig_ul3() },
            Combo { r: 12, s: 12, u: 12, csi: testdata::expected_r12_s12_u12(), signal: testdata::expected_r12_s12_u12_sig_ul3() },
            Combo { r: 3, s: 10, u: 10, csi: testdata::expected_r3_s10_u10(), signal: testdata::expected_r3_s10_u10_sig_ul3() },
            Combo { r: 50, s: 1, u: 1, csi: testdata::expected_r50_s1_u1(), signal: testdata::expected_r50_s1_u1_sig_ul3() },
            Combo { r: 32, s: 5, u: 3, csi: testdata::expected_r32_s5_u3(), signal: testdata::expected_r32_s5_u3_sig_ul3() },
        ];

        let input = testdata::testdata::test_input();
        let open = testdata::testdata::test_open();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        for combo in &combos {
            let mut ind = CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams {
                r: combo.r,
                s: combo.s,
                u: combo.u,
                ul: UL,
            })
            .unwrap();

            for i in 0..input.len() {
                let (csi, signal) = ind.update(open[i], high[i], low[i], input[i]);
                check("csi", i, combo.csi[i], csi);
                check("signal", i, combo.signal[i], signal);
            }
        }
    }

    #[test]
    fn test_passthrough() {
        let mut ind = CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams {
            r: 1,
            s: 1,
            u: 1,
            ul: 1,
        })
        .unwrap();

        // Body spans the whole range up.
        assert_eq!(ind.update(10.0, 12.0, 10.0, 12.0), (100.0, 100.0));
        // Body spans the whole range down.
        assert_eq!(ind.update(12.0, 12.0, 10.0, 10.0), (-100.0, -100.0));
        // Zero range -> division guard.
        assert_eq!(ind.update(11.0, 11.0, 11.0, 11.0).0, 0.0);
    }

    #[test]
    fn test_is_primed_after_first_bar() {
        let input = testdata::testdata::test_input();
        let open = testdata::testdata::test_open();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        let mut ind =
            CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams::default()).unwrap();
        assert!(!CandlestickStrengthIndex::is_primed(&ind));

        ind.update(open[0], high[0], low[0], input[0]);
        assert!(CandlestickStrengthIndex::is_primed(&ind));
    }

    #[test]
    fn test_mnemonic() {
        let ind =
            CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "csi(20,5,3,3)");
        assert_eq!(
            ind.metadata().description,
            "Candlestick Strength Index csi(20,5,3,3)"
        );

        let ind2 = CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams {
            r: 25,
            s: 13,
            u: 1,
            ul: 7,
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "csi(25,13,1,7)");
    }

    #[test]
    fn test_metadata() {
        let ind =
            CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::CandlestickStrengthIndex);
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(
            meta.outputs[0].kind,
            CandlestickStrengthIndexOutput::Csi as i32
        );
        assert_eq!(
            meta.outputs[1].kind,
            CandlestickStrengthIndexOutput::Signal as i32
        );
    }

    #[test]
    fn test_update_bar_ordering() {
        let input = testdata::testdata::test_input();
        let open = testdata::testdata::test_open();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();
        let exp_csi = testdata::expected_r20_s5_u3();
        let exp_signal = testdata::expected_r20_s5_u3_sig_ul3();

        let mut ind = CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams {
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
        let csi = out[0].downcast_ref::<Scalar>().unwrap().value;
        let signal = out[1].downcast_ref::<Scalar>().unwrap().value;
        check("csi", last, exp_csi[last], csi);
        check("signal", last, exp_signal[last], signal);
    }

    #[test]
    fn test_update_quote_maps_bid_to_open_and_low_and_ask_to_close_and_high() {
        let mut ind = CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams {
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
    fn test_update_scalar_and_trade_have_zero_candle_body_and_range() {
        let mut scalar_ind = CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams {
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

        let mut trade_ind = CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams {
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
