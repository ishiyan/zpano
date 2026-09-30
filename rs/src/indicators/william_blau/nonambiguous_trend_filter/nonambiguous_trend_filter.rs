use crate::entities::bar::Bar;
use crate::entities::bar_component::{BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
use crate::indicators::william_blau::candlestick_momentum_index::candlestick_momentum_index::{
    CandlestickMomentumIndex, CandlestickMomentumIndexParams,
};
use crate::indicators::william_blau::candlestick_strength_index::candlestick_strength_index::{
    CandlestickStrengthIndex, CandlestickStrengthIndexParams,
};
use crate::indicators::william_blau::directional_trend_index::directional_trend_index::{
    DirectionalTrendIndex, DirectionalTrendIndexParams,
};
use crate::indicators::william_blau::mean_deviation_index::mean_deviation_index::{
    MeanDeviationIndex, MeanDeviationIndexParams,
};
use crate::indicators::william_blau::stochastic_momentum_index::stochastic_momentum_index::{
    StochasticMomentumIndex, StochasticMomentumIndexParams,
};
use crate::indicators::william_blau::tick_volume_indicator::tick_volume_indicator::{
    TickVolumeIndicator, TickVolumeIndicatorParams,
};
use crate::indicators::william_blau::true_strength_index::true_strength_index::{
    TrueStrengthIndex, TrueStrengthIndexParams,
};

// ---------------------------------------------------------------------------
// Base
// ---------------------------------------------------------------------------

/// Specifies the base oscillator the Nonambiguous Trend Filter is applied to.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum NonambiguousTrendFilterBase {
    /// True Strength Index (named instance TSI_Trade); defaults q=2, r=32, s=13, u=3.
    Tsi = 0,
    /// Stochastic Momentum Index (named instance SMI_Trade); defaults q=32, r=64, s=7, u=1.
    Smi = 1,
    /// Directional Trend Index (named instance DTI_Trade); defaults q=2, r=28, s=28, u=5.
    Dti = 2,
    /// Tick Volume Indicator (named instance TVI_Trade); defaults r=32, s=32, u=5.
    Tvi = 3,
    /// Mean Deviation Index (MDI_Trade); defaults r=20, s=5, u=3.
    Mdi = 4,
    /// Candlestick Momentum Index (CMI_Trade); defaults r=20, s=5, u=3.
    Cmi = 5,
    /// Candlestick Strength Index (CSI_Trade); defaults r=32, s=32, u=1.
    Csi = 6,
}

impl NonambiguousTrendFilterBase {
    /// Converts an integer to a base, returning `None` for unknown values.
    pub fn from_i64(value: i64) -> Option<Self> {
        match value {
            0 => Some(Self::Tsi),
            1 => Some(Self::Smi),
            2 => Some(Self::Dti),
            3 => Some(Self::Tvi),
            4 => Some(Self::Mdi),
            5 => Some(Self::Cmi),
            6 => Some(Self::Csi),
            _ => None,
        }
    }

    /// Returns the mnemonic name and the book defaults (q, r, s, u) of the base,
    /// and whether the base uses the look-back q.
    fn defaults(self) -> (&'static str, usize, usize, usize, usize, bool) {
        match self {
            Self::Tsi => ("tsi", 2, 32, 13, 3, true),
            Self::Smi => ("smi", 32, 64, 7, 1, true),
            Self::Dti => ("dti", 2, 28, 28, 5, true),
            Self::Tvi => ("tvi", 0, 32, 32, 5, false),
            Self::Mdi => ("mdi", 0, 20, 5, 3, false),
            Self::Cmi => ("cmi", 0, 20, 5, 3, false),
            Self::Csi => ("csi", 0, 32, 32, 1, false),
        }
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Nonambiguous Trend Filter indicator.
///
/// The filter itself is parameterless; `q`, `r`, `s` and `u` are the canonical
/// symbols of the base oscillator from William Blau's Momentum, Direction, and
/// Divergence (Wiley, 1995). A zero value selects the book default of the base.
pub struct NonambiguousTrendFilterParams {
    /// The base oscillator the filter is applied to. Default `Tsi`.
    pub base: NonambiguousTrendFilterBase,
    /// The momentum look-back period of the base (`Tsi`, `Smi` and `Dti` only). Zero selects the base default.
    pub q: usize,
    /// The period of the 1st EMA of the base smoothing cascade. Zero selects the base default.
    pub r: usize,
    /// The period of the 2nd EMA of the base smoothing cascade. Zero selects the base default.
    pub s: usize,
    /// The period of the 3rd EMA of the base smoothing cascade. Zero selects the base default.
    pub u: usize,
    /// Bar component to extract (`Tsi` and `Mdi` only). `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract (`Tsi` and `Mdi` only). `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract (`Tsi` and `Mdi` only). `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for NonambiguousTrendFilterParams {
    fn default() -> Self {
        Self {
            base: NonambiguousTrendFilterBase::Tsi,
            q: 0,
            r: 0,
            s: 0,
            u: 0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Nonambiguous Trend Filter indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum NonambiguousTrendFilterOutput {
    /// The Nonambiguous Trend Filter value (the base oscillator value or 0).
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// William Blau's Nonambiguous Trend Filter (_Trade).
///
/// A post-processing transform applied to a normalized, signed base oscillator X
/// (TSI, SMI, DTI, TVI, MDI, CMI, CSI). It keeps X only where its sign and slope
/// agree and zeroes every ambiguous bar (book Ch. 8, Appendix B Figs. B-20..B-23):
///
///   X_Trade[k] = X[k]   if X[k] > 0 and X[k] - X[k-1] > 0   (positive and rising)
///              = X[k]   if X[k] < 0 and X[k] - X[k-1] < 0   (negative and falling)
///              = 0      otherwise                           (ambiguous)
///
/// The nonzero stretches correspond one-to-one with genuine up/down trends;
/// congestion and flat regions are blanked to zero.
///
/// The filter wraps an instance of the base indicator: every sample is routed to
/// the base with its own entity mapping, and the base's primary output is filtered.
///
/// Conventions: a NaN base value (the base's own look-back warm-up) yields NaN and
/// leaves the filter state untouched; the first finite base value has no prior
/// slope, so the output is 0.0; a flat step (delta == 0) is neither rising nor
/// falling, so it is zeroed.
pub struct NonambiguousTrendFilter {
    base: Box<dyn Indicator>,

    previous: f64,
    primed: bool,

    mnemonic: String,
    description: String,
}

impl NonambiguousTrendFilter {
    /// Creates a new Nonambiguous Trend Filter from the given parameters.
    pub fn new(params: &NonambiguousTrendFilterParams) -> Result<Self, String> {
        let (name, dq, dr, ds, du, uses_q) = params.base.defaults();
        let q = if params.q != 0 { params.q } else { dq };
        let r = if params.r != 0 { params.r } else { dr };
        let s = if params.s != 0 { params.s } else { ds };
        let u = if params.u != 0 { params.u } else { du };

        // Price components are meaningful only for the single-price Tsi and Mdi bases.
        let uses_components =
            params.base == NonambiguousTrendFilterBase::Tsi || params.base == NonambiguousTrendFilterBase::Mdi;
        let (bc, qc, tc) = if uses_components {
            (params.bar_component, params.quote_component, params.trade_component)
        } else {
            (None, None, None)
        };

        // The base signal line is unused, so its period is 1.
        let base: Box<dyn Indicator> = match params.base {
            NonambiguousTrendFilterBase::Tsi => Box::new(TrueStrengthIndex::new(&TrueStrengthIndexParams {
                q,
                r,
                s,
                u,
                ul: 1,
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
            })?),
            NonambiguousTrendFilterBase::Smi => Box::new(StochasticMomentumIndex::new(&StochasticMomentumIndexParams {
                q,
                r,
                s,
                u,
                ul: 1,
                ..Default::default()
            })?),
            NonambiguousTrendFilterBase::Dti => Box::new(DirectionalTrendIndex::new(&DirectionalTrendIndexParams {
                q,
                r,
                s,
                u,
                ul: 1,
                ..Default::default()
            })?),
            NonambiguousTrendFilterBase::Tvi => Box::new(TickVolumeIndicator::new(&TickVolumeIndicatorParams {
                r,
                s,
                u,
                ..Default::default()
            })?),
            NonambiguousTrendFilterBase::Mdi => Box::new(MeanDeviationIndex::new(&MeanDeviationIndexParams {
                r,
                s,
                u,
                ul: 1,
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
            })?),
            NonambiguousTrendFilterBase::Cmi => Box::new(CandlestickMomentumIndex::new(&CandlestickMomentumIndexParams {
                r,
                s,
                u,
                ul: 1,
                ..Default::default()
            })?),
            NonambiguousTrendFilterBase::Csi => Box::new(CandlestickStrengthIndex::new(&CandlestickStrengthIndexParams {
                r,
                s,
                u,
                ul: 1,
                ..Default::default()
            })?),
        };

        let triple = component_triple_mnemonic(
            bc.unwrap_or(DEFAULT_BAR_COMPONENT),
            qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
            tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
        );
        let mnemonic = if uses_q {
            format!("ntf.{}({},{},{},{}{})", name, q, r, s, u, triple)
        } else {
            format!("ntf.{}({},{},{}{})", name, r, s, u, triple)
        };
        let description = format!("Nonambiguous Trend Filter {}", mnemonic);

        Ok(Self {
            base,
            previous: 0.0,
            primed: false,
            mnemonic,
            description,
        })
    }

    /// Keeps x when positive-and-rising or negative-and-falling, else 0.
    fn filter(&mut self, x: f64) -> f64 {
        if x.is_nan() {
            // The base is still warming up: do not touch the filter state.
            return f64::NAN;
        }

        if !self.primed {
            // First finite value: no prior slope, hence ambiguous.
            self.previous = x;
            self.primed = true;
            return 0.0;
        }

        let delta = x - self.previous;
        self.previous = x;

        if x > 0.0 && delta > 0.0 {
            return x; // positive and rising
        }
        if x < 0.0 && delta < 0.0 {
            return x; // negative and falling
        }
        0.0 // ambiguous / flat / congestion
    }

    /// Filters the primary output of the base.
    fn filter_output(&mut self, output: Output) -> f64 {
        let value = output.first().and_then(|o| o.downcast_ref::<Scalar>()).map_or(f64::NAN, |s| s.value);
        self.filter(value)
    }

    /// Updates the indicator given the next single sample value, fed to the base as a scalar.
    pub fn update(&mut self, sample: f64) -> f64 {
        let output = self.base.update_scalar(&Scalar::new(0, sample));
        self.filter_output(output)
    }
}

impl Indicator for NonambiguousTrendFilter {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::NonambiguousTrendFilter,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let output = self.base.update_scalar(sample);
        let value = self.filter_output(output);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let output = self.base.update_bar(sample);
        let value = self.filter_output(output);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let output = self.base.update_quote(sample);
        let value = self.filter_output(output);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let output = self.base.update_trade(sample);
        let value = self.filter_output(output);
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
        base: NonambiguousTrendFilterBase,
        q: usize,
        r: usize,
        s: usize,
        u: usize,
        expected: Vec<f64>,
    }

    // q=0 where the base does not use it.
    fn combos() -> Vec<Combo> {
        use NonambiguousTrendFilterBase::*;
        vec![
            Combo { base: Tsi, q: 2, r: 32, s: 13, u: 3, expected: testdata::expected_tsi_r32_s13_u3() },
            Combo { base: Tsi, q: 2, r: 20, s: 5, u: 3, expected: testdata::expected_tsi_r20_s5_u3() },
            Combo { base: Tsi, q: 2, r: 40, s: 20, u: 5, expected: testdata::expected_tsi_r40_s20_u5() },
            Combo { base: Smi, q: 32, r: 64, s: 7, u: 1, expected: testdata::expected_smi_q32_r64_s7_u1() },
            Combo { base: Smi, q: 5, r: 20, s: 5, u: 3, expected: testdata::expected_smi_q5_r20_s5_u3() },
            Combo { base: Smi, q: 13, r: 25, s: 2, u: 1, expected: testdata::expected_smi_q13_r25_s2_u1() },
            Combo { base: Dti, q: 2, r: 28, s: 28, u: 5, expected: testdata::expected_dti_q2_r28_s28_u5() },
            Combo { base: Dti, q: 2, r: 20, s: 5, u: 3, expected: testdata::expected_dti_q2_r20_s5_u3() },
            Combo { base: Dti, q: 4, r: 14, s: 14, u: 3, expected: testdata::expected_dti_q4_r14_s14_u3() },
            Combo { base: Mdi, q: 0, r: 20, s: 5, u: 3, expected: testdata::expected_mdi_r20_s5_u3() },
            Combo { base: Mdi, q: 0, r: 40, s: 5, u: 3, expected: testdata::expected_mdi_r40_s5_u3() },
            Combo { base: Cmi, q: 0, r: 20, s: 5, u: 3, expected: testdata::expected_cmi_r20_s5_u3() },
            Combo { base: Cmi, q: 0, r: 10, s: 5, u: 3, expected: testdata::expected_cmi_r10_s5_u3() },
            Combo { base: Csi, q: 0, r: 32, s: 32, u: 1, expected: testdata::expected_csi_r32_s32_u1() },
            Combo { base: Csi, q: 0, r: 20, s: 5, u: 3, expected: testdata::expected_csi_r20_s5_u3() },
            Combo { base: Csi, q: 0, r: 1, s: 1, u: 1, expected: testdata::expected_csi_r1_s1_u1() },
            Combo { base: Tvi, q: 0, r: 32, s: 32, u: 5, expected: testdata::expected_tvi_r32_s32_u5() },
            Combo { base: Tvi, q: 0, r: 12, s: 12, u: 1, expected: testdata::expected_tvi_r12_s12_u1() },
            Combo { base: Tvi, q: 0, r: 25, s: 13, u: 1, expected: testdata::expected_tvi_r25_s13_u1() },
        ]
    }

    struct Data {
        open: Vec<f64>,
        high: Vec<f64>,
        low: Vec<f64>,
        close: Vec<f64>,
    }

    impl Data {
        fn load() -> Self {
            Self {
                open: testdata::testdata::test_open(),
                high: testdata::testdata::test_high(),
                low: testdata::testdata::test_low(),
                close: testdata::testdata::test_input(),
            }
        }

        fn bar(&self, i: usize) -> Bar {
            Bar {
                time: i as i64,
                open: self.open[i],
                high: self.high[i],
                low: self.low[i],
                close: self.close[i],
                volume: 0.0,
            }
        }
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

    fn value(output: &Output) -> f64 {
        output[0].downcast_ref::<Scalar>().unwrap().value
    }

    fn create(base: NonambiguousTrendFilterBase, q: usize, r: usize, s: usize, u: usize) -> NonambiguousTrendFilter {
        NonambiguousTrendFilter::new(&NonambiguousTrendFilterParams { base, q, r, s, u, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_reference_data_from_bars() {
        let data = Data::load();

        for combo in &combos() {
            let mut ind = create(combo.base, combo.q, combo.r, combo.s, combo.u);

            for i in 0..data.close.len() {
                let out = ind.update_bar(&data.bar(i));
                assert_eq!(out.len(), 1);
                check("ntf", i, combo.expected[i], value(&out));
            }
        }
    }

    #[test]
    fn test_reference_data_from_samples() {
        let input = testdata::testdata::test_input();

        for combo in combos().iter().filter(|c| {
            c.base == NonambiguousTrendFilterBase::Tsi || c.base == NonambiguousTrendFilterBase::Mdi
        }) {
            let mut ind = create(combo.base, combo.q, combo.r, combo.s, combo.u);

            for i in 0..input.len() {
                check("ntf", i, combo.expected[i], ind.update(input[i]));
            }
        }
    }

    #[test]
    fn test_rule() {
        // A passthrough CSI base is 100*(C-O)/(H-L).
        let mut ind = create(NonambiguousTrendFilterBase::Csi, 0, 1, 1, 1);
        let cases = [
            (11.0, 0.0),   // first finite value
            (12.0, 50.0),  // positive and rising
            (11.0, 0.0),   // positive and falling
            (9.0, -25.0),  // negative and falling
            (9.0, 0.0),    // flat
            (9.5, 0.0),    // negative and rising
        ];

        for (close, expected) in cases {
            let out = ind.update_bar(&Bar { time: 0, open: 10.0, high: 12.0, low: 8.0, close, volume: 0.0 });
            assert_eq!(value(&out), expected, "close {}", close);
        }
    }

    #[test]
    fn test_warm_up_region() {
        let data = Data::load();
        let q = 13;
        let mut ind = create(NonambiguousTrendFilterBase::Smi, q, 25, 2, 1);

        for i in 0..q - 1 {
            assert!(value(&ind.update_bar(&data.bar(i))).is_nan(), "[{}] expected NaN", i);
            assert!(!ind.is_primed(), "[{}] must not be primed", i);
        }

        assert_eq!(value(&ind.update_bar(&data.bar(q - 1))), 0.0);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_entity_updates() {
        let data = Data::load();
        let expected = testdata::expected_tsi_r32_s13_u3();

        let default = || NonambiguousTrendFilter::new(&NonambiguousTrendFilterParams::default()).unwrap();
        let mut scalar_ind = default();
        let mut bar_ind = default();
        let mut quote_ind = default();
        let mut trade_ind = default();

        for i in 0..data.close.len() {
            let time = i as i64;
            let v = data.close[i];

            let scalar_out = scalar_ind.update_scalar(&Scalar::new(time, v));
            assert_eq!(scalar_out.len(), 1);
            assert_eq!(scalar_out[0].downcast_ref::<Scalar>().unwrap().time, time);
            check("scalar", i, expected[i], value(&scalar_out));

            check("bar", i, expected[i], value(&bar_ind.update_bar(&data.bar(i))));

            let quote = Quote { time, bid_price: v, bid_size: 0.0, ask_price: v, ask_size: 0.0 };
            check("quote", i, expected[i], value(&quote_ind.update_quote(&quote)));

            let trade = Trade { time, price: v, volume: 0.0 };
            check("trade", i, expected[i], value(&trade_ind.update_trade(&trade)));
        }
    }

    #[test]
    fn test_metadata() {
        let ind = NonambiguousTrendFilter::new(&NonambiguousTrendFilterParams::default()).unwrap();
        let meta = ind.metadata();

        assert_eq!(meta.identifier, Identifier::NonambiguousTrendFilter);
        assert_eq!(meta.mnemonic, "ntf.tsi(2,32,13,3)");
        assert_eq!(meta.description, "Nonambiguous Trend Filter ntf.tsi(2,32,13,3)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, NonambiguousTrendFilterOutput::Value as i32);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
        assert_eq!(meta.outputs[0].mnemonic, "ntf.tsi(2,32,13,3)");
    }

    #[test]
    fn test_mnemonics() {
        use NonambiguousTrendFilterBase::*;
        let p = |base: NonambiguousTrendFilterBase| NonambiguousTrendFilterParams { base, ..Default::default() };

        let cases: Vec<(NonambiguousTrendFilterParams, &str)> = vec![
            (p(Tsi), "ntf.tsi(2,32,13,3)"),
            (p(Smi), "ntf.smi(32,64,7,1)"),
            (p(Dti), "ntf.dti(2,28,28,5)"),
            (p(Tvi), "ntf.tvi(32,32,5)"),
            (p(Mdi), "ntf.mdi(20,5,3)"),
            (p(Cmi), "ntf.cmi(20,5,3)"),
            (p(Csi), "ntf.csi(32,32,1)"),
            (NonambiguousTrendFilterParams { base: Smi, q: 5, r: 20, s: 5, u: 3, ..Default::default() }, "ntf.smi(5,20,5,3)"),
            (
                NonambiguousTrendFilterParams { bar_component: Some(BarComponent::Median), ..Default::default() },
                "ntf.tsi(2,32,13,3, hl/2)",
            ),
            (
                NonambiguousTrendFilterParams { quote_component: Some(QuoteComponent::Bid), ..Default::default() },
                "ntf.tsi(2,32,13,3, b)",
            ),
            (
                NonambiguousTrendFilterParams { trade_component: Some(TradeComponent::Volume), ..Default::default() },
                "ntf.tsi(2,32,13,3, v)",
            ),
            (
                NonambiguousTrendFilterParams {
                    bar_component: Some(BarComponent::Open),
                    quote_component: Some(QuoteComponent::Bid),
                    ..Default::default()
                },
                "ntf.tsi(2,32,13,3, o, b)",
            ),
            (
                NonambiguousTrendFilterParams {
                    bar_component: Some(BarComponent::High),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "ntf.tsi(2,32,13,3, h, v)",
            ),
            (
                NonambiguousTrendFilterParams {
                    quote_component: Some(QuoteComponent::Ask),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "ntf.tsi(2,32,13,3, a, v)",
            ),
            (
                NonambiguousTrendFilterParams { base: Mdi, bar_component: Some(BarComponent::Median), ..Default::default() },
                "ntf.mdi(20,5,3, hl/2)",
            ),
            (
                NonambiguousTrendFilterParams { base: Csi, bar_component: Some(BarComponent::Median), ..Default::default() },
                "ntf.csi(32,32,1)",
            ),
        ];

        for (params, expected) in cases {
            let ind = NonambiguousTrendFilter::new(&params).unwrap();
            assert_eq!(ind.metadata().mnemonic, expected);
        }
    }

    #[test]
    fn test_base_from_i64() {
        assert_eq!(NonambiguousTrendFilterBase::from_i64(0), Some(NonambiguousTrendFilterBase::Tsi));
        assert_eq!(NonambiguousTrendFilterBase::from_i64(6), Some(NonambiguousTrendFilterBase::Csi));
        assert_eq!(NonambiguousTrendFilterBase::from_i64(7), None);
        assert_eq!(NonambiguousTrendFilterBase::from_i64(-1), None);
    }
}
