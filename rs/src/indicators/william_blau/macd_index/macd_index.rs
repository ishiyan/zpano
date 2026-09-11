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

/// Parameters to create an instance of the MACD Index indicator.
///
/// The field names `r`, `s`, `u` and `ul` are the canonical symbols from
/// William Blau's *Momentum, Direction, and Divergence* (Wiley, 1995), chapter 5.
pub struct MacdIndexParams {
    /// Period of the slow EMA in the MACD line (`EMA(close, s) - EMA(close, r)`).
    /// Must be > 0 and strictly greater than `s`. Default 20.
    pub r: usize,
    /// Period of the fast EMA in the MACD line. Must be > 0 and strictly less
    /// than `r`. Default 5.
    pub s: usize,
    /// Period of the smoothing EMA applied to the MACD line. Must be > 0.
    /// Default 3. `u=1` yields the book's pure two-EMA MACD line.
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

impl Default for MacdIndexParams {
    fn default() -> Self {
        Self {
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

/// Enumerates the outputs of the MACD Index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MacdIndexOutput {
    /// The MACD Index line value, in raw price units (unbounded).
    Macdi = 1,
    /// The signal-line value: the ul-period EMA of the index.
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

/// William Blau's MACD Index (MACD_I).
///
/// Blau's MACD line is the difference of two EMAs of the close, optionally
/// smoothed by a third EMA, paired with an EMA signal line (the Ergodic form,
/// Blau ch. 5):
///
///   macd_k   = EMA(close, s)_k - EMA(close, r)_k          (MACD line; s fast, r slow)
///   macdi_k  = EMA(macd, u)_k                             (the MACD_I line)
///   signal_k = EMA(macdi, ul)_k                           (ul-period EMA)
///
/// with the fast period s strictly shorter than the slow period r (s < r).
/// Setting u=1 recovers the book's pure two-EMA MACD line. Blau notes the MACD
/// and the MDI are both double-smoothed momentum indicators with nearly
/// interchangeable shapes (within a scale factor).
///
/// The index is NOT normalized: there is no 100 * TEMA/TEMA ratio and no fixed
/// range, so the output is in the same price units as the input and may take any
/// sign or magnitude. Because there is no division there is also no division guard.
///
/// The indicator produces two outputs:
///   - MACDI: the index line, in raw price units, finite from bar 0;
///   - Signal: the ul-period EMA of the index (Blau's Ergodic signal line).
///
/// Priming: both price EMAs are defined from bar 0, so macd_0 = 0 and the u
/// smoothing and signal EMAs seed on that 0 -- there is no NaN warm-up region
/// and bar 0 is exactly 0.0.
pub struct MacdIndex {
    ema_fast: Ema,
    ema_slow: Ema,
    smooth_u: Ema,
    signal_ema: Ema,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
}

impl MacdIndex {
    /// Creates a new MACD Index from the given parameters.
    pub fn new(params: &MacdIndexParams) -> Result<Self, String> {
        let invalid = "invalid macd index parameters";

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
        if s >= r {
            return Err(format!("{}: s (fast) should be less than r (slow)", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "macdi({},{},{}{})",
            r,
            s,
            u,
            component_triple_mnemonic(bc, qc, tc)
        );

        Ok(Self {
            ema_fast: Ema::new(s),
            ema_slow: Ema::new(r),
            smooth_u: Ema::new(u),
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

    /// Core update returning (macdi, signal).
    pub fn update(&mut self, sample: f64) -> (f64, f64) {
        // MACD line = fast EMA - slow EMA. Both seed at bar 0 (to close_0), so
        // the line is 0.0 on bar 0 and defined on every bar thereafter.
        let macd = self.ema_fast.update(sample) - self.ema_slow.update(sample);

        // Smooth the MACD line: EMA(macd, u). No normalization, no guard.
        let macdi = self.smooth_u.update(macd);

        // Signal line = EMA(macdi, ul); seeds here on the bar-0 index value.
        let signal = self.signal_ema.update(macdi);
        self.primed = true;

        (macdi, signal)
    }
}

impl Indicator for MacdIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("MACD Index {}", self.mnemonic);
        build_metadata(
            Identifier::MacdIndex,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} macdi", self.mnemonic),
                    description: format!("{} MACDI", desc),
                },
                OutputText {
                    mnemonic: format!("{} signal", self.mnemonic),
                    description: format!("{} signal", desc),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (macdi, signal) = self.update(sample.value);
        vec![
            Box::new(Scalar { time: sample.time, value: macdi }),
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

    // Signal-line EMA period used for every expected signal array.
    const UL: usize = 3;

    fn check(name: &str, i: usize, exp: f64, act: f64) {
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
        macdi: Vec<f64>,
        signal: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { r: 20, s: 5, u: 3, macdi: testdata::expected_r20_s5_u3(), signal: testdata::expected_r20_s5_u3_sig_ul3() },
            Combo { r: 20, s: 5, u: 1, macdi: testdata::expected_r20_s5_u1(), signal: testdata::expected_r20_s5_u1_sig_ul3() },
            Combo { r: 26, s: 12, u: 3, macdi: testdata::expected_r26_s12_u3(), signal: testdata::expected_r26_s12_u3_sig_ul3() },
            Combo { r: 26, s: 12, u: 1, macdi: testdata::expected_r26_s12_u1(), signal: testdata::expected_r26_s12_u1_sig_ul3() },
            Combo { r: 35, s: 5, u: 3, macdi: testdata::expected_r35_s5_u3(), signal: testdata::expected_r35_s5_u3_sig_ul3() },
            Combo { r: 10, s: 3, u: 5, macdi: testdata::expected_r10_s3_u5(), signal: testdata::expected_r10_s3_u5_sig_ul3() },
            Combo { r: 32, s: 12, u: 5, macdi: testdata::expected_r32_s12_u5(), signal: testdata::expected_r32_s12_u5_sig_ul3() },
            Combo { r: 17, s: 8, u: 1, macdi: testdata::expected_r17_s8_u1(), signal: testdata::expected_r17_s8_u1_sig_ul3() },
            Combo { r: 20, s: 10, u: 3, macdi: testdata::expected_r20_s10_u3(), signal: testdata::expected_r20_s10_u3_sig_ul3() },
            Combo { r: 8, s: 4, u: 2, macdi: testdata::expected_r8_s4_u2(), signal: testdata::expected_r8_s4_u2_sig_ul3() },
            Combo { r: 30, s: 15, u: 1, macdi: testdata::expected_r30_s15_u1(), signal: testdata::expected_r30_s15_u1_sig_ul3() },
            Combo { r: 3, s: 2, u: 3, macdi: testdata::expected_r3_s2_u3(), signal: testdata::expected_r3_s2_u3_sig_ul3() },
            Combo { r: 50, s: 12, u: 1, macdi: testdata::expected_r50_s12_u1(), signal: testdata::expected_r50_s12_u1_sig_ul3() },
            Combo { r: 19, s: 6, u: 3, macdi: testdata::expected_r19_s6_u3(), signal: testdata::expected_r19_s6_u3_sig_ul3() },
            Combo { r: 20, s: 5, u: 5, macdi: testdata::expected_r20_s5_u5(), signal: testdata::expected_r20_s5_u5_sig_ul3() },
            Combo { r: 60, s: 30, u: 10, macdi: testdata::expected_r60_s30_u10(), signal: testdata::expected_r60_s30_u10_sig_ul3() },
        ];

        let input = testdata::testdata::test_input();

        for combo in &combos {
            let mut ind = MacdIndex::new(&MacdIndexParams {
                r: combo.r,
                s: combo.s,
                u: combo.u,
                ul: UL,
                ..Default::default()
            })
            .unwrap();

            for i in 0..input.len() {
                let (macdi, signal) = ind.update(input[i]);
                check("macdi", i, combo.macdi[i], macdi);
                check("signal", i, combo.signal[i], signal);
            }
        }
    }

    #[test]
    fn test_no_warm_up() {
        let input = testdata::testdata::test_input();

        let mut ind = MacdIndex::new(&MacdIndexParams::default()).unwrap();

        let (macdi0, signal0) = ind.update(input[0]);
        assert_eq!(macdi0, 0.0);
        assert_eq!(signal0, 0.0);

        for i in 1..input.len() {
            let (macdi, signal) = ind.update(input[i]);
            assert!(!macdi.is_nan(), "macdi[{}]: unexpected NaN", i);
            assert!(!signal.is_nan(), "signal[{}]: unexpected NaN", i);
        }
    }

    #[test]
    fn test_passthrough_smoothing() {
        let mut ind = MacdIndex::new(&MacdIndexParams {
            r: 2,
            s: 1,
            u: 1,
            ul: 1,
            ..Default::default()
        })
        .unwrap();

        // Bar 0 seeds both price EMAs, so the MACD line is exactly 0.
        let (macdi0, signal0) = ind.update(10.0);
        check("macdi", 0, 0.0, macdi0);
        check("signal", 0, 0.0, signal0);

        // EMA(1) is a passthrough -> fast = 12. EMA(2) = (2/3)*12 + (1/3)*10 = 11.3333.
        let (macdi1, signal1) = ind.update(12.0);
        check("macdi", 1, 12.0 - (2.0 / 3.0 * 12.0 + 1.0 / 3.0 * 10.0), macdi1);
        check("signal", 1, macdi1, signal1);
    }

    #[test]
    fn test_signal_passthrough_when_ul_is_one() {
        let input = testdata::testdata::test_input();

        let mut ind = MacdIndex::new(&MacdIndexParams {
            r: 20,
            s: 5,
            u: 3,
            ul: 1,
            ..Default::default()
        })
        .unwrap();

        for i in 0..input.len() {
            let (macdi, signal) = ind.update(input[i]);
            assert_eq!(macdi, signal, "signal[{}] must equal macdi when ul = 1", i);
        }
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::testdata::test_input();

        let mut ind = MacdIndex::new(&MacdIndexParams::default()).unwrap();

        assert!(!ind.is_primed());
        ind.update(input[0]);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_mnemonic_excludes_ul() {
        let ind = MacdIndex::new(&MacdIndexParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "macdi(20,5,3)");

        let ind2 = MacdIndex::new(&MacdIndexParams {
            r: 26,
            s: 12,
            u: 9,
            ul: 7,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "macdi(26,12,9)");
    }

    #[test]
    fn test_metadata() {
        let ind = MacdIndex::new(&MacdIndexParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::MacdIndex);
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(meta.outputs[0].kind, MacdIndexOutput::Macdi as i32);
        assert_eq!(meta.outputs[1].kind, MacdIndexOutput::Signal as i32);
    }

    #[test]
    fn test_update_scalar_ordering() {
        let input = testdata::testdata::test_input();
        let exp_macdi = testdata::expected_r20_s5_u3();
        let exp_signal = testdata::expected_r20_s5_u3_sig_ul3();

        let mut ind = MacdIndex::new(&MacdIndexParams {
            ul: UL,
            ..Default::default()
        })
        .unwrap();
        let mut out: Output = vec![];
        for i in 0..input.len() {
            out = ind.update_scalar(&Scalar { time: 0, value: input[i] });
        }
        let last = input.len() - 1;
        let macdi = out[0].downcast_ref::<Scalar>().unwrap().value;
        let signal = out[1].downcast_ref::<Scalar>().unwrap().value;
        check("macdi", last, exp_macdi[last], macdi);
        check("signal", last, exp_signal[last], signal);
    }

    #[test]
    fn test_invalid_params() {
        assert!(MacdIndex::new(&MacdIndexParams { r: 20, s: 20, u: 3, ul: 3, ..Default::default() }).is_err());
        assert!(MacdIndex::new(&MacdIndexParams { r: 20, s: 25, u: 3, ul: 3, ..Default::default() }).is_err());
    }
}
