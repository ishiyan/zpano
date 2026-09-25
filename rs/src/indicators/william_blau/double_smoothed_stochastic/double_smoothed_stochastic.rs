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

/// Parameters to create an instance of the Double Smoothed Stochastic indicator.
///
/// The field names `q`, `r`, `s` and `g` are the canonical symbols from William
/// Blau's *Momentum, Direction, and Divergence* (Wiley, 1995).
///
/// The indicator consumes the high, low and close prices of a bar, so it has no
/// configurable price-component fields.
pub struct DoubleSmoothedStochasticParams {
    /// Stochastic look-back period (bars for the highest high and lowest low). Must be > 0. Default 5.
    pub q: usize,
    /// Period of the 1st (inner) EMA, applied to the raw stochastic and the range. Must be > 0. Default 7.
    pub r: usize,
    /// Period of the 2nd (outer) EMA in the cascade. Must be > 0. Default 3.
    pub s: usize,
    /// Period of the signal-line SMA (second output). Must be > 0. Default 3.
    pub g: usize,
}

impl Default for DoubleSmoothedStochasticParams {
    fn default() -> Self {
        Self {
            q: 5,
            r: 7,
            s: 3,
            g: 3,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Double Smoothed Stochastic indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DoubleSmoothedStochasticOutput {
    /// The Double Smoothed Stochastic oscillator value (range [0, 100]).
    Dss = 1,
    /// The signal-line value: the g-period SMA of the oscillator.
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
// Inlined SMA (signal line)
// ---------------------------------------------------------------------------

/// Stateful streaming SMA over the last period inputs.
///
/// Returns the mean of the window's current contents on every update: an expanding
/// window while fewer than period values have arrived, then a rolling period-bar
/// window. There is no NaN warm-up (finite from the first input).
///
/// period == 1 -> pure passthrough (output == input).
struct Sma {
    period: usize,
    window: Vec<f64>,
    count: usize,
    index: usize,
}

impl Sma {
    fn new(period: usize) -> Self {
        Self {
            period,
            window: vec![0.0; period],
            count: 0,
            index: 0,
        }
    }

    fn update(&mut self, x: f64) -> f64 {
        self.window[self.index] = x;
        self.index = (self.index + 1) % self.period;

        if self.count < self.period {
            self.count += 1;
        }

        // Naive left-to-right sum from the oldest to the newest value (NOT a
        // compensated sum), so that every port reproduces the same values.
        // Once full, the oldest value sits at the next write position.
        let start = if self.count == self.period { self.index } else { 0 };

        let mut sum = 0.0;
        for i in 0..self.count {
            sum += self.window[(start + i) % self.period];
        }

        sum / self.count as f64
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// William Blau's Double Smoothed Stochastic (DSS).
///
/// A classic double-smoothed stochastic oscillator bounded to [0, 100], paired
/// with a short simple-moving-average signal line:
///
///   dss_k    = 100 * EMA(EMA(st, r), s) / EMA(EMA(rng, r), s)   (the oscillator)
///   signal_k = SMA(dss, g)_k                                    (g-period SMA)
///
/// where, over the last q bars, HH_k is the highest high and LL_k is the lowest low,
/// st_k = close_k - LL_k >= 0 is the raw stochastic (the close above the low), and
/// rng_k = HH_k - LL_k >= 0 is the q-bar range.
///
/// The raw stochastic and the range are smoothed separately with the same two-stage
/// EMA cascade (r then s), then divided. Because 0 <= st <= rng on every bar, the
/// ratio is bounded to [0, 100]. With q = 1 it is Blau's one-bar HLC index. It is
/// exactly the MQL5 Blau_TStochI with its third EMA period u = 1. The inputs are the
/// high, low and close prices.
///
/// The indicator produces two outputs:
///   - DSS: the oscillator, range [0, 100];
///   - Signal: the g-period SMA of the oscillator.
///
/// Priming convention (book / EasyLanguage): st and rng become valid once q bars of
/// high/low exist, i.e. at bar q-1. All four cascade stages seed there together, so
/// both outputs are NaN for bars 0..q-2 and finite from bar q-1; for q = 1 there is
/// no NaN warm-up. The signal SMA seeds on the first finite oscillator value and
/// returns the mean of the oscillator values seen so far (expanding window <= g),
/// then the full g-bar rolling mean. Division guard: denominator <= 0 -> oscillator 0.0.
pub struct DoubleSmoothedStochastic {
    q: usize,
    highs: Vec<f64>,
    lows: Vec<f64>,
    window_count: usize,
    window_index: usize,
    num_r: Ema,
    num_s: Ema,
    den_r: Ema,
    den_s: Ema,
    signal_sma: Sma,
    primed: bool,
    mnemonic: String,
}

impl DoubleSmoothedStochastic {
    /// Creates a new Double Smoothed Stochastic from the given parameters.
    pub fn new(params: &DoubleSmoothedStochasticParams) -> Result<Self, String> {
        let invalid = "invalid double smoothed stochastic parameters";

        let mut q = params.q;
        if q == 0 {
            q = 5;
        }
        let mut r = params.r;
        if r == 0 {
            r = 7;
        }
        let mut s = params.s;
        if s == 0 {
            s = 3;
        }
        let mut g = params.g;
        if g == 0 {
            g = 3;
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
        if g < 1 {
            return Err(format!("{}: g should be greater than 0", invalid));
        }

        let mnemonic = format!("dss({},{},{},{})", q, r, s, g);

        Ok(Self {
            q,
            highs: vec![0.0; q],
            lows: vec![0.0; q],
            window_count: 0,
            window_index: 0,
            num_r: Ema::new(r),
            num_s: Ema::new(s),
            den_r: Ema::new(r),
            den_s: Ema::new(s),
            signal_sma: Sma::new(g),
            primed: false,
            mnemonic,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update taking one bar's high, low and close, returning (dss, signal).
    /// Both are NaN until q bars have been seen.
    pub fn update(&mut self, high: f64, low: f64, close: f64) -> (f64, f64) {
        self.highs[self.window_index] = high;
        self.lows[self.window_index] = low;
        self.window_index = (self.window_index + 1) % self.q;

        if self.window_count < self.q {
            self.window_count += 1;
        }

        // Need q bars of high/low before the stochastic is defined. Until then
        // neither output exists -- do NOT advance the EMA cascades or the SMA.
        if self.window_count < self.q {
            return (f64::NAN, f64::NAN);
        }

        // Rolling extremes over the last q bars.
        let mut hh = self.highs[0];
        let mut ll = self.lows[0];
        for i in 1..self.q {
            hh = hh.max(self.highs[i]);
            ll = ll.min(self.lows[i]);
        }

        // Raw stochastic and range (both non-negative).
        let st = close - ll;
        let rng = hh - ll;

        // Numerator cascade: EMA(EMA(st, r), s).
        let n = self.num_s.update(self.num_r.update(st));
        // Denominator cascade: EMA(EMA(rng, r), s).
        let d = self.den_s.update(self.den_r.update(rng));

        // Division guard: flat window so far -> oscillator 0.0.
        let dss = if d > 0.0 { 100.0 * n / d } else { 0.0 };

        // Signal line = SMA(dss, g); seeds on the first finite oscillator value.
        let signal = self.signal_sma.update(dss);
        self.primed = true;

        (dss, signal)
    }

    /// Updates the indicator and wraps the two outputs.
    fn update_entity(&mut self, time: i64, high: f64, low: f64, close: f64) -> Output {
        let (dss, signal) = self.update(high, low, close);
        vec![
            Box::new(Scalar { time, value: dss }),
            Box::new(Scalar {
                time,
                value: signal,
            }),
        ]
    }
}

impl Indicator for DoubleSmoothedStochastic {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Double Smoothed Stochastic {}", self.mnemonic);
        build_metadata(
            Identifier::DoubleSmoothedStochastic,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} dss", self.mnemonic),
                    description: format!("{} DSS", desc),
                },
                OutputText {
                    mnemonic: format!("{} signal", self.mnemonic),
                    description: format!("{} signal", desc),
                },
            ],
        )
    }

    /// A scalar carries a single value, used as the high, the low and the close.
    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = sample.value;
        self.update_entity(sample.time, v, v, v)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        self.update_entity(sample.time, sample.high, sample.low, sample.close)
    }

    /// A quote maps the mid price to the high, the low and the close.
    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (sample.bid_price + sample.ask_price) / 2.0;
        self.update_entity(sample.time, v, v, v)
    }

    /// A trade carries a single price, used as the high, the low and the close.
    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = sample.price;
        self.update_entity(sample.time, v, v, v)
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::super::testdata;
    use super::*;

    const TOLERANCE: f64 = 1e-10;

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

    fn passthrough(q: usize) -> DoubleSmoothedStochastic {
        DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams {
            q,
            r: 1,
            s: 1,
            g: 1,
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
        g: usize,
        dss: Vec<f64>,
        signal: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { q: 5, r: 7, s: 3, g: 3, dss: testdata::expected_ds_q5_r7_s3_g3(), signal: testdata::expected_sig_q5_r7_s3_g3() },
            Combo { q: 2, r: 3, s: 15, g: 3, dss: testdata::expected_ds_q2_r3_s15_g3(), signal: testdata::expected_sig_q2_r3_s15_g3() },
            Combo { q: 5, r: 20, s: 5, g: 3, dss: testdata::expected_ds_q5_r20_s5_g3(), signal: testdata::expected_sig_q5_r20_s5_g3() },
            Combo { q: 5, r: 7, s: 3, g: 1, dss: testdata::expected_ds_q5_r7_s3_g1(), signal: testdata::expected_sig_q5_r7_s3_g1() },
            Combo { q: 2, r: 3, s: 15, g: 1, dss: testdata::expected_ds_q2_r3_s15_g1(), signal: testdata::expected_sig_q2_r3_s15_g1() },
            Combo { q: 1, r: 1, s: 1, g: 1, dss: testdata::expected_ds_q1_r1_s1_g1(), signal: testdata::expected_sig_q1_r1_s1_g1() },
            Combo { q: 1, r: 5, s: 5, g: 3, dss: testdata::expected_ds_q1_r5_s5_g3(), signal: testdata::expected_sig_q1_r5_s5_g3() },
            Combo { q: 8, r: 5, s: 3, g: 3, dss: testdata::expected_ds_q8_r5_s3_g3(), signal: testdata::expected_sig_q8_r5_s3_g3() },
            Combo { q: 21, r: 13, s: 4, g: 3, dss: testdata::expected_ds_q21_r13_s4_g3(), signal: testdata::expected_sig_q21_r13_s4_g3() },
            Combo { q: 5, r: 1, s: 1, g: 3, dss: testdata::expected_ds_q5_r1_s1_g3(), signal: testdata::expected_sig_q5_r1_s1_g3() },
            Combo { q: 3, r: 10, s: 10, g: 5, dss: testdata::expected_ds_q3_r10_s10_g5(), signal: testdata::expected_sig_q3_r10_s10_g5() },
            Combo { q: 34, r: 5, s: 5, g: 3, dss: testdata::expected_ds_q34_r5_s5_g3(), signal: testdata::expected_sig_q34_r5_s5_g3() },
            Combo { q: 2, r: 3, s: 15, g: 5, dss: testdata::expected_ds_q2_r3_s15_g5(), signal: testdata::expected_sig_q2_r3_s15_g5() },
            Combo { q: 10, r: 7, s: 3, g: 3, dss: testdata::expected_ds_q10_r7_s3_g3(), signal: testdata::expected_sig_q10_r7_s3_g3() },
        ];

        let input = testdata::testdata::test_input();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        for combo in &combos {
            let mut ind = DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams {
                q: combo.q,
                r: combo.r,
                s: combo.s,
                g: combo.g,
            })
            .unwrap();

            for i in 0..input.len() {
                let (dss, signal) = ind.update(high[i], low[i], input[i]);
                check("dss", i, combo.dss[i], dss);
                check("signal", i, combo.signal[i], signal);
            }
        }
    }

    #[test]
    fn test_passthrough() {
        let mut ind = passthrough(1);

        // Close at high.
        assert_eq!(ind.update(12.0, 10.0, 12.0), (100.0, 100.0));
        // Close at low.
        assert_eq!(ind.update(12.0, 10.0, 10.0), (0.0, 0.0));
        // Exact midpoint.
        assert_eq!(ind.update(12.0, 10.0, 11.0), (50.0, 50.0));
        // Flat window -> division guard.
        assert_eq!(ind.update(11.0, 11.0, 11.0), (0.0, 0.0));
    }

    #[test]
    fn test_signal_expanding_then_rolling_window() {
        let mut ind = DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams {
            q: 1,
            r: 1,
            s: 1,
            g: 3,
        })
        .unwrap();

        // dss 100.
        check("signal", 0, 100.0, ind.update(12.0, 10.0, 12.0).1);
        // dss 0.
        check("signal", 1, 50.0, ind.update(12.0, 10.0, 10.0).1);
        // dss 50.
        check("signal", 2, 50.0, ind.update(12.0, 10.0, 11.0).1);
        // dss 50, 100 dropped.
        check("signal", 3, 100.0 / 3.0, ind.update(12.0, 10.0, 11.0).1);
    }

    #[test]
    fn test_signal_equals_dss_when_g_is_one() {
        let input = testdata::testdata::test_input();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        let mut ind = DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams {
            q: 5,
            r: 7,
            s: 3,
            g: 1,
        })
        .unwrap();

        for i in 0..input.len() {
            let (dss, signal) = ind.update(high[i], low[i], input[i]);
            if dss.is_nan() {
                assert!(signal.is_nan(), "[{}] expected NaN signal", i);
            } else {
                assert_eq!(dss, signal, "[{}]", i);
            }
        }
    }

    #[test]
    fn test_is_primed_from_bar_q_minus_one() {
        let input = testdata::testdata::test_input();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();
        let q = 5;

        let mut ind = DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams {
            q,
            ..Default::default()
        })
        .unwrap();
        assert!(!DoubleSmoothedStochastic::is_primed(&ind));

        for i in 0..q - 1 {
            let (dss, signal) = ind.update(high[i], low[i], input[i]);
            assert!(!DoubleSmoothedStochastic::is_primed(&ind), "[{}] primed too early", i);
            assert!(dss.is_nan());
            assert!(signal.is_nan());
        }

        for i in q - 1..input.len() {
            ind.update(high[i], low[i], input[i]);
            assert!(DoubleSmoothedStochastic::is_primed(&ind), "[{}] not primed", i);
        }
    }

    #[test]
    fn test_is_primed_after_first_bar_when_q_is_one() {
        let input = testdata::testdata::test_input();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        let mut ind = DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams {
            q: 1,
            ..Default::default()
        })
        .unwrap();
        assert!(!DoubleSmoothedStochastic::is_primed(&ind));

        ind.update(high[0], low[0], input[0]);
        assert!(DoubleSmoothedStochastic::is_primed(&ind));
    }

    #[test]
    fn test_mnemonic() {
        let ind =
            DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "dss(5,7,3,3)");
        assert_eq!(
            ind.metadata().description,
            "Double Smoothed Stochastic dss(5,7,3,3)"
        );

        let ind2 = DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams {
            q: 2,
            r: 3,
            s: 15,
            g: 5,
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "dss(2,3,15,5)");
    }

    #[test]
    fn test_metadata() {
        let ind =
            DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::DoubleSmoothedStochastic);
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(
            meta.outputs[0].kind,
            DoubleSmoothedStochasticOutput::Dss as i32
        );
        assert_eq!(
            meta.outputs[1].kind,
            DoubleSmoothedStochasticOutput::Signal as i32
        );
        assert_eq!(meta.outputs[0].mnemonic, "dss(5,7,3,3) dss");
        assert_eq!(
            meta.outputs[0].description,
            "Double Smoothed Stochastic dss(5,7,3,3) DSS"
        );
        assert_eq!(meta.outputs[1].mnemonic, "dss(5,7,3,3) signal");
        assert_eq!(
            meta.outputs[1].description,
            "Double Smoothed Stochastic dss(5,7,3,3) signal"
        );
    }

    #[test]
    fn test_zero_params_resolve_to_defaults() {
        // Zero means "use default", so an all-zero params value is valid.
        let ind = DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams {
            q: 0,
            r: 0,
            s: 0,
            g: 0,
        })
        .unwrap();
        assert_eq!(ind.metadata().mnemonic, "dss(5,7,3,3)");
    }

    #[test]
    fn test_update_bar_ordering() {
        let input = testdata::testdata::test_input();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();
        let exp_dss = testdata::expected_ds_q5_r7_s3_g3();
        let exp_signal = testdata::expected_sig_q5_r7_s3_g3();

        let mut ind =
            DoubleSmoothedStochastic::new(&DoubleSmoothedStochasticParams::default()).unwrap();

        let mut out: Output = Vec::new();
        for i in 0..input.len() {
            out = ind.update_bar(&Bar {
                time: 0,
                open: 0.0,
                high: high[i],
                low: low[i],
                close: input[i],
                volume: 0.0,
            });
        }

        let last = input.len() - 1;
        assert_eq!(out.len(), 2);
        check("dss", last, exp_dss[last], scalar_value(&out, 0));
        check("signal", last, exp_signal[last], scalar_value(&out, 1));
    }

    #[test]
    fn test_update_scalar_uses_value_as_high_low_and_close() {
        let mut ind = passthrough(2);

        let first = ind.update_scalar(&Scalar {
            time: 0,
            value: 10.0,
        });
        assert!(scalar_value(&first, 0).is_nan());
        assert!(scalar_value(&first, 1).is_nan());

        // Close at the 2-bar high.
        let out = ind.update_scalar(&Scalar {
            time: 0,
            value: 12.0,
        });
        assert_eq!(scalar_value(&out, 0), 100.0);
        assert_eq!(scalar_value(&out, 1), 100.0);
    }

    #[test]
    fn test_update_quote_uses_mid_price_as_high_low_and_close() {
        let mut ind = passthrough(2);

        ind.update_quote(&Quote {
            time: 0,
            bid_price: 12.0,
            ask_price: 14.0,
            bid_size: 1.0,
            ask_size: 1.0,
        });

        // Mid 11 is at the 2-bar low.
        let out = ind.update_quote(&Quote {
            time: 0,
            bid_price: 10.0,
            ask_price: 12.0,
            bid_size: 1.0,
            ask_size: 1.0,
        });
        assert_eq!(scalar_value(&out, 0), 0.0);
        assert_eq!(scalar_value(&out, 1), 0.0);
    }

    #[test]
    fn test_update_trade_uses_price_as_high_low_and_close() {
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
