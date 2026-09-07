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

/// Parameters to create an instance of the Mean Deviation Index indicator.
///
/// The field names `r`, `s`, `u` and `ul` are the canonical symbols from
/// William Blau's *Momentum, Direction, and Divergence* (Wiley, 1995), chapter 5.
pub struct MeanDeviationIndexParams {
    /// Period of the baseline (detrending) EMA subtracted from the price.
    /// Must be > 0. Default 20. `r=1` makes the detrend a passthrough, so the
    /// index is identically 0.
    pub r: usize,
    /// Period of the 1st smoothing EMA, applied to the mean deviation. Must be > 0. Default 5.
    pub s: usize,
    /// Period of the 2nd smoothing EMA in the cascade. Must be > 0. Default 3.
    /// `u=1` yields the book's pure double-smoothed form.
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

impl Default for MeanDeviationIndexParams {
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

/// Enumerates the outputs of the Mean Deviation Index indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MeanDeviationIndexOutput {
    /// The Mean Deviation Index line value, in raw price units (unbounded).
    Mdi = 1,
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

/// William Blau's Mean Deviation Index (MDI).
///
/// A detrended, double-/triple-smoothed momentum line in raw price units, paired
/// with an EMA signal line (the Ergodic form, Blau ch. 5):
///
///   md_k     = price_k - EMA(price, r)_k                  (deviation from trend)
///   mdi_k    = EMA(EMA(md, s), u)_k                       (the MDI line)
///   signal_k = EMA(mdi, ul)_k                             (ul-period EMA)
///
/// The price series is detrended by subtracting its own r-period EMA, then the
/// deviation is smoothed by an s-period EMA and an optional u-period EMA. Blau
/// notes the MDI approximates the MACD when r is long and s is short.
///
/// The index is NOT normalized: there is no 100 * TEMA/TEMA ratio and no fixed
/// range, so the output is in the same price units as the input and may take any
/// sign or magnitude. Because there is no division there is also no division guard.
///
/// The indicator produces two outputs:
///   - MDI: the index line, in raw price units, finite from bar 0;
///   - Signal: the ul-period EMA of the index (Blau's Ergodic signal line).
///
/// Priming: the detrending EMA is defined from bar 0, so md_0 = 0 and both
/// smoothing EMAs seed on that 0 -- there is no NaN warm-up region and bar 0 is
/// exactly 0.0. Degenerate case: r=1 makes the detrend a passthrough, so the
/// index is identically 0.0.
pub struct MeanDeviationIndex {
    trend: Ema,
    smooth_s: Ema,
    smooth_u: Ema,
    signal_ema: Ema,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
}

impl MeanDeviationIndex {
    /// Creates a new Mean Deviation Index from the given parameters.
    pub fn new(params: &MeanDeviationIndexParams) -> Result<Self, String> {
        let invalid = "invalid mean deviation index parameters";

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

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "mdi({},{},{}{})",
            r,
            s,
            u,
            component_triple_mnemonic(bc, qc, tc)
        );

        Ok(Self {
            trend: Ema::new(r),
            smooth_s: Ema::new(s),
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

    /// Core update returning (mdi, signal).
    pub fn update(&mut self, sample: f64) -> (f64, f64) {
        // Mean deviation: the price minus its own r-period EMA trend. The
        // baseline EMA seeds on bar 0, so the bar-0 deviation is exactly 0.
        let md = sample - self.trend.update(sample);

        // Smooth the deviation: EMA(EMA(md, s), u). No normalization, no guard.
        let mdi = self.smooth_u.update(self.smooth_s.update(md));

        // Signal line = EMA(mdi, ul); seeds here on the bar-0 index value.
        let signal = self.signal_ema.update(mdi);
        self.primed = true;

        (mdi, signal)
    }
}

impl Indicator for MeanDeviationIndex {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Mean Deviation Index {}", self.mnemonic);
        build_metadata(
            Identifier::MeanDeviationIndex,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} mdi", self.mnemonic),
                    description: format!("{} MDI", desc),
                },
                OutputText {
                    mnemonic: format!("{} signal", self.mnemonic),
                    description: format!("{} signal", desc),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (mdi, signal) = self.update(sample.value);
        vec![
            Box::new(Scalar { time: sample.time, value: mdi }),
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
        mdi: Vec<f64>,
        signal: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { r: 20, s: 5, u: 3, mdi: testdata::expected_r20_s5_u3(), signal: testdata::expected_r20_s5_u3_sig_ul3() },
            Combo { r: 20, s: 5, u: 1, mdi: testdata::expected_r20_s5_u1(), signal: testdata::expected_r20_s5_u1_sig_ul3() },
            Combo { r: 1, s: 5, u: 3, mdi: testdata::expected_r1_s5_u3(), signal: testdata::expected_r1_s5_u3_sig_ul3() },
            Combo { r: 40, s: 5, u: 3, mdi: testdata::expected_r40_s5_u3(), signal: testdata::expected_r40_s5_u3_sig_ul3() },
            Combo { r: 10, s: 5, u: 3, mdi: testdata::expected_r10_s5_u3(), signal: testdata::expected_r10_s5_u3_sig_ul3() },
            Combo { r: 5, s: 5, u: 5, mdi: testdata::expected_r5_s5_u5(), signal: testdata::expected_r5_s5_u5_sig_ul3() },
            Combo { r: 20, s: 9, u: 1, mdi: testdata::expected_r20_s9_u1(), signal: testdata::expected_r20_s9_u1_sig_ul3() },
            Combo { r: 26, s: 12, u: 9, mdi: testdata::expected_r26_s12_u9(), signal: testdata::expected_r26_s12_u9_sig_ul3() },
            Combo { r: 50, s: 13, u: 1, mdi: testdata::expected_r50_s13_u1(), signal: testdata::expected_r50_s13_u1_sig_ul3() },
            Combo { r: 30, s: 5, u: 3, mdi: testdata::expected_r30_s5_u3(), signal: testdata::expected_r30_s5_u3_sig_ul3() },
            Combo { r: 3, s: 3, u: 3, mdi: testdata::expected_r3_s3_u3(), signal: testdata::expected_r3_s3_u3_sig_ul3() },
            Combo { r: 7, s: 4, u: 2, mdi: testdata::expected_r7_s4_u2(), signal: testdata::expected_r7_s4_u2_sig_ul3() },
            Combo { r: 2, s: 5, u: 3, mdi: testdata::expected_r2_s5_u3(), signal: testdata::expected_r2_s5_u3_sig_ul3() },
            Combo { r: 20, s: 1, u: 1, mdi: testdata::expected_r20_s1_u1(), signal: testdata::expected_r20_s1_u1_sig_ul3() },
            Combo { r: 20, s: 20, u: 5, mdi: testdata::expected_r20_s20_u5(), signal: testdata::expected_r20_s20_u5_sig_ul3() },
            Combo { r: 60, s: 30, u: 10, mdi: testdata::expected_r60_s30_u10(), signal: testdata::expected_r60_s30_u10_sig_ul3() },
        ];

        let input = testdata::testdata::test_input();

        for combo in &combos {
            let mut ind = MeanDeviationIndex::new(&MeanDeviationIndexParams {
                r: combo.r,
                s: combo.s,
                u: combo.u,
                ul: UL,
                ..Default::default()
            })
            .unwrap();

            for i in 0..input.len() {
                let (mdi, signal) = ind.update(input[i]);
                check("mdi", i, combo.mdi[i], mdi);
                check("signal", i, combo.signal[i], signal);
            }
        }
    }

    #[test]
    fn test_no_warm_up() {
        let input = testdata::testdata::test_input();

        let mut ind = MeanDeviationIndex::new(&MeanDeviationIndexParams::default()).unwrap();

        let (mdi0, signal0) = ind.update(input[0]);
        assert_eq!(mdi0, 0.0);
        assert_eq!(signal0, 0.0);

        for i in 1..input.len() {
            let (mdi, signal) = ind.update(input[i]);
            assert!(!mdi.is_nan(), "mdi[{}]: unexpected NaN", i);
            assert!(!signal.is_nan(), "signal[{}]: unexpected NaN", i);
        }
    }

    #[test]
    fn test_degenerate_r_one_is_identically_zero() {
        let input = testdata::testdata::test_input();

        let mut ind = MeanDeviationIndex::new(&MeanDeviationIndexParams {
            r: 1,
            s: 5,
            u: 3,
            ul: 3,
            ..Default::default()
        })
        .unwrap();

        for i in 0..input.len() {
            let (mdi, signal) = ind.update(input[i]);
            assert_eq!(mdi, 0.0, "mdi[{}] must be 0 when r = 1", i);
            assert_eq!(signal, 0.0, "signal[{}] must be 0 when r = 1", i);
        }
    }

    #[test]
    fn test_passthrough_smoothing() {
        let mut ind = MeanDeviationIndex::new(&MeanDeviationIndexParams {
            r: 2,
            s: 1,
            u: 1,
            ul: 1,
            ..Default::default()
        })
        .unwrap();

        // Bar 0 seeds the baseline, so the deviation is exactly 0.
        let (mdi0, signal0) = ind.update(10.0);
        check("mdi", 0, 0.0, mdi0);
        check("signal", 0, 0.0, signal0);

        // EMA(2) = (2/3)*13 + (1/3)*10 = 12, so the deviation is 13-12 = 1.
        let (mdi1, signal1) = ind.update(13.0);
        check("mdi", 1, 1.0, mdi1);
        check("signal", 1, 1.0, signal1);
    }

    #[test]
    fn test_signal_passthrough_when_ul_is_one() {
        let input = testdata::testdata::test_input();

        let mut ind = MeanDeviationIndex::new(&MeanDeviationIndexParams {
            r: 20,
            s: 5,
            u: 3,
            ul: 1,
            ..Default::default()
        })
        .unwrap();

        for i in 0..input.len() {
            let (mdi, signal) = ind.update(input[i]);
            assert_eq!(mdi, signal, "signal[{}] must equal mdi when ul = 1", i);
        }
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::testdata::test_input();

        let mut ind = MeanDeviationIndex::new(&MeanDeviationIndexParams::default()).unwrap();

        assert!(!ind.is_primed());
        ind.update(input[0]);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_mnemonic_excludes_ul() {
        let ind = MeanDeviationIndex::new(&MeanDeviationIndexParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "mdi(20,5,3)");

        let ind2 = MeanDeviationIndex::new(&MeanDeviationIndexParams {
            r: 26,
            s: 12,
            u: 9,
            ul: 7,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "mdi(26,12,9)");
    }

    #[test]
    fn test_metadata() {
        let ind = MeanDeviationIndex::new(&MeanDeviationIndexParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::MeanDeviationIndex);
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(meta.outputs[0].kind, MeanDeviationIndexOutput::Mdi as i32);
        assert_eq!(meta.outputs[1].kind, MeanDeviationIndexOutput::Signal as i32);
    }

    #[test]
    fn test_update_scalar_ordering() {
        let input = testdata::testdata::test_input();
        let exp_mdi = testdata::expected_r20_s5_u3();
        let exp_signal = testdata::expected_r20_s5_u3_sig_ul3();

        let mut ind = MeanDeviationIndex::new(&MeanDeviationIndexParams {
            ul: UL,
            ..Default::default()
        })
        .unwrap();
        let mut out: Output = vec![];
        for i in 0..input.len() {
            out = ind.update_scalar(&Scalar { time: 0, value: input[i] });
        }
        let last = input.len() - 1;
        let mdi = out[0].downcast_ref::<Scalar>().unwrap().value;
        let signal = out[1].downcast_ref::<Scalar>().unwrap().value;
        check("mdi", last, exp_mdi[last], mdi);
        check("signal", last, exp_signal[last], signal);
    }
}
