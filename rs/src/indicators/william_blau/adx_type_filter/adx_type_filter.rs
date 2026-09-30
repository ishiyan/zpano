use std::collections::VecDeque;

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
// Source
// ---------------------------------------------------------------------------

/// Specifies the bipolar momentum the ADX-Type Filter is applied to.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AdxTypeFilterSource {
    /// The TSI numerator `C - C[q-1]` (named instance TSI_ATF).
    TsiMomentum = 0,
    /// The SMI raw stochastic momentum `C - 0.5*(HH(q) + LL(q))` (named instance SMI_ATF).
    SmiMomentum = 1,
    /// The DTI numerator `max(H - H[q-1], 0) - max(L[q-1] - L, 0)`.
    DtiMomentum = 2,
    /// The TVI tick balance `upticks - downticks` (`2C - H - L` for a bar).
    TviBalance = 3,
    /// The single-smoothed normalized `TSI(q, r, 1, 1)`, replacing the inner EMA.
    TsiNormalized = 4,
}

impl AdxTypeFilterSource {
    /// Converts an integer to a source, returning `None` for unknown values.
    pub fn from_i64(value: i64) -> Option<Self> {
        match value {
            0 => Some(Self::TsiMomentum),
            1 => Some(Self::SmiMomentum),
            2 => Some(Self::DtiMomentum),
            3 => Some(Self::TviBalance),
            4 => Some(Self::TsiNormalized),
            _ => None,
        }
    }

    fn mnemonic(self) -> &'static str {
        match self {
            Self::TsiMomentum => "tsi",
            Self::SmiMomentum => "smi",
            Self::DtiMomentum => "dti",
            Self::TviBalance => "tvi",
            Self::TsiNormalized => "tsin",
        }
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the ADX-Type Filter indicator.
///
/// The field names `q`, `r` and `s` are the canonical symbols from William Blau's
/// Momentum, Direction, and Divergence (Wiley, 1995), Appendix B, Figure B-24.
pub struct AdxTypeFilterParams {
    /// The bipolar momentum the filter is applied to. Default `TsiMomentum`.
    pub source: AdxTypeFilterSource,
    /// The momentum look-back period. Zero selects the source default: 2 for
    /// `TsiMomentum`, `DtiMomentum` and `TsiNormalized`, 32 for `SmiMomentum`.
    /// Not used by `TviBalance`. Default 0.
    pub q: usize,
    /// The period of the inner EMA (for `TsiNormalized`, the smoothing period of the
    /// normalized TSI, which replaces the inner EMA). Must be > 0. Default 32.
    pub r: usize,
    /// The period of the outer EMA. Must be > 0. Default 32.
    pub s: usize,
    /// Bar component to extract (TSI sources only). `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract (TSI sources only). `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract (TSI sources only). `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for AdxTypeFilterParams {
    fn default() -> Self {
        Self {
            source: AdxTypeFilterSource::TsiMomentum,
            q: 0,
            r: 32,
            s: 32,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the ADX-Type Filter indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AdxTypeFilterOutput {
    /// The ADX-Type Filter value (non-negative).
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

/// Appends a sample to a rolling window holding at most `length` values.
fn push(window: &mut VecDeque<f64>, length: usize, sample: f64) {
    if window.len() >= length {
        window.pop_front();
    }
    window.push_back(sample);
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// William Blau's ADX-Type Filter (ATF).
///
/// A non-negative trend-strength filter, analogous to Wilder's ADX, built by
/// rectifying and double-smoothing a bipolar momentum series (book Fig. B-24):
///
///   ATF(Price, r, s) = EMA(|EMA(Price, r)|, s)
///
/// The inner EMA(r) smooths the signed momentum, the absolute value discards the
/// direction and keeps the amplitude, and the outer EMA(s) smooths the amplitude.
/// A rising ATF signals a strengthening trend, a falling ATF a ranging market.
///
/// The bipolar momentum is selected by the source:
///   - `TsiMomentum`:   C - C[q-1]                               (TSI_ATF);
///   - `SmiMomentum`:   C - 0.5*(HH(q) + LL(q))                  (SMI_ATF);
///   - `DtiMomentum`:   max(H - H[q-1], 0) - max(L[q-1] - L, 0);
///   - `TviBalance`:    upticks - downticks (2C - H - L for a bar);
///   - `TsiNormalized`: TSI(q, r, 1, 1), which replaces the inner EMA (r = 1).
///
/// Priming: each EMA stage seeds on its first finite momentum value. A NaN momentum
/// (the q-bar look-back warm-up) is propagated: the output is NaN and the EMAs do not
/// advance. The output is always >= 0.
pub struct AdxTypeFilter {
    source: AdxTypeFilterSource,
    uses_components: bool,
    q: usize,

    // Rolling windows of the last q values (for the q-bar look-back).
    closes: VecDeque<f64>,
    highs: VecDeque<f64>,
    lows: VecDeque<f64>,

    // Tick-rule state for single-valued samples of the TviBalance source.
    previous: f64,
    has_previous: bool,

    // The normalized TSI (TsiNormalized only) replaces the inner EMA.
    tsi: Option<TrueStrengthIndex>,
    inner: Ema,
    outer: Ema,

    primed: bool,

    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,

    mnemonic: String,
    description: String,
}

impl AdxTypeFilter {
    /// Creates a new ADX-Type Filter from the given parameters.
    pub fn new(params: &AdxTypeFilterParams) -> Result<Self, String> {
        let invalid = "invalid adx type filter parameters";

        let source = params.source;
        let q = if params.q != 0 {
            params.q
        } else if source == AdxTypeFilterSource::SmiMomentum {
            32
        } else {
            2
        };
        let r = params.r;
        let s = params.s;

        if r < 1 {
            return Err(format!("{}: r should be greater than 0", invalid));
        }
        if s < 1 {
            return Err(format!("{}: s should be greater than 0", invalid));
        }

        // Price components are meaningful only for the single-price TSI sources; the
        // other sources use the bar's high/low/close or the single value of the sample
        // (scalar value, quote mid price, trade price).
        let uses_components =
            source == AdxTypeFilterSource::TsiMomentum || source == AdxTypeFilterSource::TsiNormalized;

        let (bc, qc, tc) = if uses_components {
            (
                params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT),
                params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT),
            )
        } else {
            (DEFAULT_BAR_COMPONENT, DEFAULT_QUOTE_COMPONENT, DEFAULT_TRADE_COMPONENT)
        };

        let name = source.mnemonic();
        let mnemonic = if source == AdxTypeFilterSource::TviBalance {
            format!("atf.{}({},{})", name, r, s)
        } else {
            format!("atf.{}({},{},{}{})", name, q, r, s, component_triple_mnemonic(bc, qc, tc))
        };
        let description = format!("ADX-Type Filter {}", mnemonic);

        let (tsi, inner) = if source == AdxTypeFilterSource::TsiNormalized {
            let tsi = TrueStrengthIndex::new(&TrueStrengthIndexParams {
                q,
                r,
                s: 1,
                u: 1,
                ul: 1,
                ..Default::default()
            })?;
            (Some(tsi), Ema::new(1))
        } else {
            (None, Ema::new(r))
        };

        Ok(Self {
            source,
            uses_components,
            q,
            closes: VecDeque::with_capacity(q),
            highs: VecDeque::with_capacity(q),
            lows: VecDeque::with_capacity(q),
            previous: 0.0,
            has_previous: false,
            tsi,
            inner,
            outer: Ema::new(s),
            primed: false,
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
            mnemonic,
            description,
        })
    }

    /// Inner smooth -> rectify -> outer smooth; a NaN momentum propagates without
    /// advancing the EMAs.
    fn filter(&mut self, momentum: f64) -> f64 {
        if momentum.is_nan() {
            return f64::NAN;
        }

        self.primed = true;

        self.outer.update(self.inner.update(momentum).abs())
    }

    /// Updates the indicator given the next single sample value.
    ///
    /// The `SmiMomentum` and `DtiMomentum` sources use the value as the high, the low
    /// and the close; the `TviBalance` source applies the tick rule
    /// (balance = value - previous value).
    pub fn update(&mut self, sample: f64) -> f64 {
        match self.source {
            AdxTypeFilterSource::TsiMomentum => {
                push(&mut self.closes, self.q, sample);
                if self.closes.len() < self.q {
                    return f64::NAN;
                }

                // mtm_k = C_k - C_(k-(q-1)); the leftmost element is C_(k-(q-1)).
                let momentum = sample - self.closes[0];
                self.filter(momentum)
            }
            AdxTypeFilterSource::TsiNormalized => {
                let (tsi, _) = self.tsi.as_mut().map_or((f64::NAN, f64::NAN), |t| t.update(sample));
                self.filter(tsi)
            }
            AdxTypeFilterSource::TviBalance => {
                let balance = if self.has_previous { sample - self.previous } else { 0.0 };
                self.previous = sample;
                self.has_previous = true;
                self.filter(balance)
            }
            _ => self.update_high_low_close(sample, sample, sample),
        }
    }

    /// Updates the indicator given the next bar's high, low and close.
    ///
    /// The `TsiMomentum` and `TsiNormalized` sources use the close only.
    pub fn update_high_low_close(&mut self, high: f64, low: f64, close: f64) -> f64 {
        match self.source {
            AdxTypeFilterSource::SmiMomentum => {
                push(&mut self.highs, self.q, high);
                push(&mut self.lows, self.q, low);
                if self.highs.len() < self.q {
                    return f64::NAN;
                }

                // sm = C - 0.5*(HH(q) + LL(q)).
                let hh = self.highs.iter().copied().fold(f64::NEG_INFINITY, f64::max);
                let ll = self.lows.iter().copied().fold(f64::INFINITY, f64::min);
                self.filter(close - 0.5 * (hh + ll))
            }
            AdxTypeFilterSource::DtiMomentum => {
                push(&mut self.highs, self.q, high);
                push(&mut self.lows, self.q, low);
                if self.highs.len() < self.q {
                    return f64::NAN;
                }

                // HMU - LMD; the leftmost elements are H_(k-(q-1)) and L_(k-(q-1)).
                let hmu = (high - self.highs[0]).max(0.0);
                let lmd = (self.lows[0] - low).max(0.0);
                self.filter(hmu - lmd)
            }
            // up - down = (C - L) - (H - C) = 2C - H - L.
            AdxTypeFilterSource::TviBalance => self.filter(2.0 * close - high - low),
            _ => self.update(close),
        }
    }
}

impl Indicator for AdxTypeFilter {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::AdxTypeFilter,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    /// The TSI sources use the bar component; the other sources use the bar's high,
    /// low and close.
    fn update_bar(&mut self, sample: &Bar) -> Output {
        let value = if self.uses_components {
            let sample_value = (self.bar_func)(sample);
            self.update(sample_value)
        } else {
            self.update_high_low_close(sample.high, sample.low, sample.close)
        };
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let sample_value = (self.quote_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let sample_value = (self.trade_func)(sample);
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
        source: AdxTypeFilterSource,
        q: usize,
        r: usize,
        s: usize,
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

    fn create(source: AdxTypeFilterSource, q: usize, r: usize, s: usize) -> AdxTypeFilter {
        AdxTypeFilter::new(&AdxTypeFilterParams { source, q, r, s, ..Default::default() }).unwrap()
    }

    fn test_bar(i: usize, input: &[f64], high: &[f64], low: &[f64]) -> Bar {
        Bar {
            time: i as i64,
            open: input[i],
            high: high[i],
            low: low[i],
            close: input[i],
            volume: 0.0,
        }
    }

    fn bar_combos() -> Vec<Combo> {
        use AdxTypeFilterSource::*;
        vec![
            Combo { source: SmiMomentum, q: 32, r: 32, s: 32, expected: testdata::expected_smiraw_q32_r32_s32() },
            Combo { source: SmiMomentum, q: 32, r: 20, s: 5, expected: testdata::expected_smiraw_q32_r20_s5() },
            Combo { source: SmiMomentum, q: 5, r: 32, s: 32, expected: testdata::expected_smiraw_q5_r32_s32() },
            Combo { source: DtiMomentum, q: 2, r: 32, s: 32, expected: testdata::expected_dtinum_q2_r32_s32() },
            Combo { source: DtiMomentum, q: 2, r: 28, s: 28, expected: testdata::expected_dtinum_q2_r28_s28() },
            Combo { source: DtiMomentum, q: 5, r: 32, s: 32, expected: testdata::expected_dtinum_q5_r32_s32() },
            Combo { source: TviBalance, q: 0, r: 32, s: 32, expected: testdata::expected_tvi_r32_s32() },
            Combo { source: TviBalance, q: 0, r: 12, s: 12, expected: testdata::expected_tvi_r12_s12() },
            Combo { source: TviBalance, q: 0, r: 1, s: 1, expected: testdata::expected_tvi_r1_s1() },
        ]
    }

    #[test]
    fn test_reference_data_close_sources() {
        use AdxTypeFilterSource::*;
        let combos = vec![
            Combo { source: TsiMomentum, q: 2, r: 32, s: 32, expected: testdata::expected_tsimtm_q2_r32_s32() },
            Combo { source: TsiMomentum, q: 2, r: 20, s: 5, expected: testdata::expected_tsimtm_q2_r20_s5() },
            Combo { source: TsiMomentum, q: 2, r: 13, s: 1, expected: testdata::expected_tsimtm_q2_r13_s1() },
            Combo { source: TsiMomentum, q: 2, r: 1, s: 1, expected: testdata::expected_tsimtm_q2_r1_s1() },
            Combo { source: TsiMomentum, q: 5, r: 32, s: 32, expected: testdata::expected_tsimtm_q5_r32_s32() },
            Combo { source: TsiNormalized, q: 2, r: 32, s: 32, expected: testdata::expected_tsinorm_r32_s32() },
            Combo { source: TsiNormalized, q: 2, r: 20, s: 20, expected: testdata::expected_tsinorm_r20_s20() },
        ];

        let input = testdata::testdata::test_input();

        for combo in &combos {
            let mut ind = create(combo.source, combo.q, combo.r, combo.s);

            for i in 0..input.len() {
                check("atf", i, combo.expected[i], ind.update(input[i]));
            }
        }
    }

    #[test]
    fn test_reference_data_bar_sources() {
        let input = testdata::testdata::test_input();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        for combo in &bar_combos() {
            let mut ind = create(combo.source, combo.q, combo.r, combo.s);

            for i in 0..input.len() {
                let out = ind.update_bar(&test_bar(i, &input, &high, &low));
                assert_eq!(out.len(), 1);
                check("atf", i, combo.expected[i], out[0].downcast_ref::<Scalar>().unwrap().value);
            }
        }
    }

    #[test]
    fn test_reference_data_bar_sources_high_low_close() {
        let input = testdata::testdata::test_input();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        for combo in &bar_combos() {
            let mut ind = create(combo.source, combo.q, combo.r, combo.s);

            for i in 0..input.len() {
                check("atf", i, combo.expected[i], ind.update_high_low_close(high[i], low[i], input[i]));
            }
        }
    }

    #[test]
    fn test_passthrough() {
        let mut tsi = create(AdxTypeFilterSource::TsiMomentum, 2, 1, 1);
        assert!(tsi.update(10.0).is_nan());
        assert_eq!(tsi.update(12.0), 2.0);
        assert_eq!(tsi.update(11.0), 1.0);

        let mut smi = create(AdxTypeFilterSource::SmiMomentum, 1, 1, 1);
        assert_eq!(smi.update_high_low_close(11.0, 9.0, 10.5), 0.5);

        let mut tvi = create(AdxTypeFilterSource::TviBalance, 0, 1, 1);
        assert_eq!(tvi.update(10.0), 0.0);
        assert_eq!(tvi.update(12.0), 2.0);
        assert_eq!(tvi.update(9.0), 3.0);
    }

    #[test]
    fn test_warm_up_region() {
        let input = testdata::testdata::test_input();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        let mut smi = create(AdxTypeFilterSource::SmiMomentum, 0, 32, 32);
        let mut tvi = create(AdxTypeFilterSource::TviBalance, 0, 32, 32);

        for i in 0..input.len() {
            assert_eq!(smi.update_high_low_close(high[i], low[i], input[i]).is_nan(), i < 31, "[{}]", i);
            assert!(!tvi.update_high_low_close(high[i], low[i], input[i]).is_nan(), "[{}]", i);
        }
    }

    #[test]
    fn test_non_negative() {
        use AdxTypeFilterSource::*;
        let input = testdata::testdata::test_input();
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();

        for source in [TsiMomentum, SmiMomentum, DtiMomentum, TviBalance, TsiNormalized] {
            let mut ind = create(source, 0, 32, 32);

            for i in 0..input.len() {
                let out = ind.update_bar(&test_bar(i, &input, &high, &low));
                let value = out[0].downcast_ref::<Scalar>().unwrap().value;
                if !value.is_nan() {
                    assert!(value >= 0.0, "{:?}[{}] negative", source, i);
                }
            }
        }
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::testdata::test_input();
        let q = 5;
        let mut ind = create(AdxTypeFilterSource::TsiMomentum, q, 32, 32);

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
        let high = testdata::testdata::test_high();
        let low = testdata::testdata::test_low();
        let expected = testdata::expected_tsimtm_q2_r32_s32();

        let default = || AdxTypeFilter::new(&AdxTypeFilterParams::default()).unwrap();
        let mut scalar_ind = default();
        let mut bar_ind = default();
        let mut quote_ind = default();
        let mut trade_ind = default();

        for i in 0..input.len() {
            let time = i as i64;

            let scalar_out = scalar_ind.update_scalar(&Scalar::new(time, input[i]));
            assert_eq!(scalar_out.len(), 1);
            check("scalar", i, expected[i], scalar_out[0].downcast_ref::<Scalar>().unwrap().value);

            let bar_out = bar_ind.update_bar(&test_bar(i, &input, &high, &low));
            check("bar", i, expected[i], bar_out[0].downcast_ref::<Scalar>().unwrap().value);

            let quote_out = quote_ind.update_quote(&Quote {
                time,
                bid_price: input[i],
                bid_size: 0.0,
                ask_price: input[i],
                ask_size: 0.0,
            });
            check("quote", i, expected[i], quote_out[0].downcast_ref::<Scalar>().unwrap().value);

            let trade_out = trade_ind.update_trade(&Trade { time, price: input[i], volume: 0.0 });
            check("trade", i, expected[i], trade_out[0].downcast_ref::<Scalar>().unwrap().value);
        }
    }

    #[test]
    fn test_smi_single_values_use_value_as_high_low_close() {
        let input = testdata::testdata::test_input();
        let smi = || create(AdxTypeFilterSource::SmiMomentum, 5, 32, 32);
        let mut reference = smi();
        let mut scalar_ind = smi();
        let mut quote_ind = smi();
        let mut trade_ind = smi();

        for i in 0..input.len() {
            let v = input[i];
            let exp = reference.update_high_low_close(v, v, v);

            let scalar_out = scalar_ind.update_scalar(&Scalar::new(0, v));
            check("scalar", i, exp, scalar_out[0].downcast_ref::<Scalar>().unwrap().value);

            let quote_out = quote_ind.update_quote(&Quote {
                time: 0,
                bid_price: v,
                bid_size: 0.0,
                ask_price: v,
                ask_size: 0.0,
            });
            check("quote", i, exp, quote_out[0].downcast_ref::<Scalar>().unwrap().value);

            let trade_out = trade_ind.update_trade(&Trade { time: 0, price: v, volume: 0.0 });
            check("trade", i, exp, trade_out[0].downcast_ref::<Scalar>().unwrap().value);
        }
    }

    #[test]
    fn test_metadata() {
        let ind = AdxTypeFilter::new(&AdxTypeFilterParams::default()).unwrap();
        let meta = ind.metadata();

        assert_eq!(meta.identifier, Identifier::AdxTypeFilter);
        assert_eq!(meta.mnemonic, "atf.tsi(2,32,32)");
        assert_eq!(meta.description, "ADX-Type Filter atf.tsi(2,32,32)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, AdxTypeFilterOutput::Value as i32);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
        assert_eq!(meta.outputs[0].mnemonic, "atf.tsi(2,32,32)");
    }

    #[test]
    fn test_mnemonics() {
        use AdxTypeFilterSource::*;
        let p = |source: AdxTypeFilterSource| AdxTypeFilterParams { source, ..Default::default() };

        let cases: Vec<(AdxTypeFilterParams, &str)> = vec![
            (p(TsiMomentum), "atf.tsi(2,32,32)"),
            (p(SmiMomentum), "atf.smi(32,32,32)"),
            (p(DtiMomentum), "atf.dti(2,32,32)"),
            (p(TviBalance), "atf.tvi(32,32)"),
            (p(TsiNormalized), "atf.tsin(2,32,32)"),
            (AdxTypeFilterParams { source: SmiMomentum, q: 5, r: 20, s: 5, ..Default::default() }, "atf.smi(5,20,5)"),
            (AdxTypeFilterParams { bar_component: Some(BarComponent::Median), ..Default::default() }, "atf.tsi(2,32,32, hl/2)"),
            (AdxTypeFilterParams { quote_component: Some(QuoteComponent::Bid), ..Default::default() }, "atf.tsi(2,32,32, b)"),
            (AdxTypeFilterParams { trade_component: Some(TradeComponent::Volume), ..Default::default() }, "atf.tsi(2,32,32, v)"),
            (
                AdxTypeFilterParams {
                    bar_component: Some(BarComponent::Open),
                    quote_component: Some(QuoteComponent::Bid),
                    ..Default::default()
                },
                "atf.tsi(2,32,32, o, b)",
            ),
            (
                AdxTypeFilterParams {
                    bar_component: Some(BarComponent::High),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "atf.tsi(2,32,32, h, v)",
            ),
            (
                AdxTypeFilterParams {
                    quote_component: Some(QuoteComponent::Ask),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "atf.tsi(2,32,32, a, v)",
            ),
            (
                AdxTypeFilterParams { source: TsiNormalized, bar_component: Some(BarComponent::Median), ..Default::default() },
                "atf.tsin(2,32,32, hl/2)",
            ),
            (
                AdxTypeFilterParams { source: SmiMomentum, bar_component: Some(BarComponent::Median), ..Default::default() },
                "atf.smi(32,32,32)",
            ),
        ];

        for (params, expected) in cases {
            let ind = AdxTypeFilter::new(&params).unwrap();
            assert_eq!(ind.metadata().mnemonic, expected);
        }
    }

    #[test]
    fn test_source_from_i64() {
        assert_eq!(AdxTypeFilterSource::from_i64(0), Some(AdxTypeFilterSource::TsiMomentum));
        assert_eq!(AdxTypeFilterSource::from_i64(4), Some(AdxTypeFilterSource::TsiNormalized));
        assert_eq!(AdxTypeFilterSource::from_i64(5), None);
        assert_eq!(AdxTypeFilterSource::from_i64(-1), None);
    }

    #[test]
    fn test_invalid_params() {
        assert!(AdxTypeFilter::new(&AdxTypeFilterParams { r: 0, ..Default::default() }).is_err());
        assert!(AdxTypeFilter::new(&AdxTypeFilterParams { s: 0, ..Default::default() }).is_err());
    }
}
