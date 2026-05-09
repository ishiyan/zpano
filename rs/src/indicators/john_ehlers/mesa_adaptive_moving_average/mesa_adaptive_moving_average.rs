use crate::entities::bar::Bar;
use crate::entities::bar_component::{
    component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT,
};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{
    component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT,
};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{
    component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT,
};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;
use crate::indicators::core::outputs::band::Band;
use crate::indicators::john_ehlers::hilbert_transformer::{
    new_cycle_estimator, estimator_moniker, CycleEstimator, CycleEstimatorParams,
    CycleEstimatorType,
};

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for MAMA based on lengths.
pub struct MesaAdaptiveMovingAverageLengthParams {
    pub estimator_type: CycleEstimatorType,
    pub estimator_params: CycleEstimatorParams,
    pub fast_limit_length: i64,
    pub slow_limit_length: i64,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

/// Parameters for MAMA based on smoothing factors.
pub struct MesaAdaptiveMovingAverageSmoothingFactorParams {
    pub estimator_type: CycleEstimatorType,
    pub estimator_params: CycleEstimatorParams,
    pub fast_limit_smoothing_factor: f64,
    pub slow_limit_smoothing_factor: f64,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the MAMA indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MesaAdaptiveMovingAverageOutput {
    /// The scalar value of the MAMA.
    Value = 1,
    /// The scalar value of the FAMA.
    Fama = 2,
    /// The band output (MAMA upper, FAMA lower).
    Band = 3,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Mesa Adaptive Moving Average (MAMA).
pub struct MesaAdaptiveMovingAverage {
    line: LineIndicator,
    mnemonic_fama: String,
    description_fama: String,
    mnemonic_band: String,
    description_band: String,
    alpha_fast_limit: f64,
    alpha_slow_limit: f64,
    previous_phase: f64,
    mama: f64,
    fama: f64,
    htce: Box<dyn CycleEstimator>,
    is_phase_cached: bool,
    primed: bool,
}

impl std::fmt::Debug for MesaAdaptiveMovingAverage {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("MesaAdaptiveMovingAverage").finish()
    }
}

impl MesaAdaptiveMovingAverage {
    /// Creates a new MAMA with default parameters (fast=3, slow=39, homodyne discriminator).
    pub fn new_default() -> Result<Self, String> {
        Self::new_internal(
            CycleEstimatorType::HomodyneDiscriminator,
            &CycleEstimatorParams {
                smoothing_length: 4,
                alpha_ema_quadrature_in_phase: 0.2,
                alpha_ema_period: 0.2,
                warm_up_period: 0,
            },
            3, 39,
            f64::NAN, f64::NAN,
            None, None, None,
        )
    }

    /// Creates a new MAMA from length-based parameters.
    pub fn new_length(p: &MesaAdaptiveMovingAverageLengthParams) -> Result<Self, String> {
        Self::new_internal(
            p.estimator_type, &p.estimator_params,
            p.fast_limit_length, p.slow_limit_length,
            f64::NAN, f64::NAN,
            p.bar_component, p.quote_component, p.trade_component,
        )
    }

    /// Creates a new MAMA from smoothing-factor-based parameters.
    pub fn new_smoothing_factor(p: &MesaAdaptiveMovingAverageSmoothingFactorParams) -> Result<Self, String> {
        Self::new_internal(
            p.estimator_type, &p.estimator_params,
            0, 0,
            p.fast_limit_smoothing_factor, p.slow_limit_smoothing_factor,
            p.bar_component, p.quote_component, p.trade_component,
        )
    }

    #[allow(clippy::too_many_arguments)]
    fn new_internal(
        estimator_type: CycleEstimatorType,
        estimator_params: &CycleEstimatorParams,
        fast_limit_length: i64,
        slow_limit_length: i64,
        mut fast_limit_sf: f64,
        mut slow_limit_sf: f64,
        bc_opt: Option<BarComponent>,
        qc_opt: Option<QuoteComponent>,
        tc_opt: Option<TradeComponent>,
    ) -> Result<Self, String> {
        const INVALID: &str = "invalid mesa adaptive moving average parameters";
        const EPSILON: f64 = 0.00000001;

        let estimator = new_cycle_estimator(estimator_type, estimator_params)?;

        // Build estimator moniker (only when non-default).
        let estimator_moniker_str = if estimator_type != CycleEstimatorType::HomodyneDiscriminator
            || estimator_params.smoothing_length != 4
            || estimator_params.alpha_ema_quadrature_in_phase != 0.2
            || estimator_params.alpha_ema_period != 0.2
        {
            let m = estimator_moniker(estimator_type, estimator.as_ref());
            if m.is_empty() { String::new() } else { format!(", {}", m) }
        } else {
            String::new()
        };

        let bc = bc_opt.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = qc_opt.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = tc_opt.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let comp_mnemonic = component_triple_mnemonic(bc, qc, tc);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let (mnemonic, mnemonic_fama, mnemonic_band);

        if fast_limit_sf.is_nan() {
            // Length-based
            if fast_limit_length < 2 {
                return Err(format!("{}: fast limit length should be larger than 1", INVALID));
            }
            if slow_limit_length < 2 {
                return Err(format!("{}: slow limit length should be larger than 1", INVALID));
            }
            fast_limit_sf = 2.0 / (1 + fast_limit_length) as f64;
            slow_limit_sf = 2.0 / (1 + slow_limit_length) as f64;

            mnemonic = format!("mama({}, {}{}{})", fast_limit_length, slow_limit_length, estimator_moniker_str, comp_mnemonic);
            mnemonic_fama = format!("fama({}, {}{}{})", fast_limit_length, slow_limit_length, estimator_moniker_str, comp_mnemonic);
            mnemonic_band = format!("mama-fama({}, {}{}{})", fast_limit_length, slow_limit_length, estimator_moniker_str, comp_mnemonic);
        } else {
            // Smoothing-factor-based
            if fast_limit_sf < 0.0 || fast_limit_sf > 1.0 {
                return Err(format!("{}: fast limit smoothing factor should be in range [0, 1]", INVALID));
            }
            if slow_limit_sf < 0.0 || slow_limit_sf > 1.0 {
                return Err(format!("{}: slow limit smoothing factor should be in range [0, 1]", INVALID));
            }
            if fast_limit_sf < EPSILON {
                fast_limit_sf = EPSILON;
            }
            if slow_limit_sf < EPSILON {
                slow_limit_sf = EPSILON;
            }

            mnemonic = format!("mama({:.3}, {:.3}{}{})", fast_limit_sf, slow_limit_sf, estimator_moniker_str, comp_mnemonic);
            mnemonic_fama = format!("fama({:.3}, {:.3}{}{})", fast_limit_sf, slow_limit_sf, estimator_moniker_str, comp_mnemonic);
            mnemonic_band = format!("mama-fama({:.3}, {:.3}{}{})", fast_limit_sf, slow_limit_sf, estimator_moniker_str, comp_mnemonic);
        }

        let descr = "Mesa adaptive moving average ";
        let description = format!("{}{}", descr, mnemonic);
        let description_fama = format!("{}{}", descr, mnemonic_fama);
        let description_band = format!("{}{}", descr, mnemonic_band);

        let line = LineIndicator::new(
            mnemonic,
            description,
            bar_func,
            quote_func,
            trade_func,
        );

        Ok(Self {
            line,
            mnemonic_fama,
            description_fama,
            mnemonic_band,
            description_band,
            alpha_fast_limit: fast_limit_sf,
            alpha_slow_limit: slow_limit_sf,
            previous_phase: 0.0,
            mama: 0.0,
            fama: 0.0,
            htce: estimator,
            is_phase_cached: false,
            primed: false,
        })
    }

    /// Core update. Returns MAMA value or NaN if not primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        self.htce.update(sample);

        if self.primed {
            return self.calculate(sample);
        }

        if self.htce.primed() {
            if self.is_phase_cached {
                self.primed = true;
                return self.calculate(sample);
            }

            self.is_phase_cached = true;
            self.previous_phase = self.calculate_phase();
            self.mama = sample;
            self.fama = sample;
        }

        f64::NAN
    }

    /// Returns the current FAMA value.
    pub fn fama(&self) -> f64 {
        self.fama
    }

    fn calculate_phase(&self) -> f64 {
        if self.htce.in_phase() == 0.0 {
            return self.previous_phase;
        }

        const RAD2DEG: f64 = 180.0 / std::f64::consts::PI;

        let phase = (self.htce.quadrature() / self.htce.in_phase()).atan() * RAD2DEG;
        if !phase.is_nan() && !phase.is_infinite() {
            return phase;
        }

        self.previous_phase
    }

    fn calculate_mama(&mut self, sample: f64) -> f64 {
        let phase = self.calculate_phase();

        let mut phase_rate_of_change = self.previous_phase - phase;
        self.previous_phase = phase;

        if phase_rate_of_change < 1.0 {
            phase_rate_of_change = 1.0;
        }

        let alpha = (self.alpha_fast_limit / phase_rate_of_change)
            .max(self.alpha_slow_limit)
            .min(self.alpha_fast_limit);

        self.mama = alpha * sample + (1.0 - alpha) * self.mama;

        alpha
    }

    fn calculate(&mut self, sample: f64) -> f64 {
        let alpha = self.calculate_mama(sample) / 2.0;
        self.fama = alpha * self.mama + (1.0 - alpha) * self.fama;

        self.mama
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let mama_val = self.update(sample);

        let fama_val = if mama_val.is_nan() {
            f64::NAN
        } else {
            self.fama
        };

        vec![
            Box::new(Scalar::new(time, mama_val)),
            Box::new(Scalar::new(time, fama_val)),
            Box::new(Band { time, upper: mama_val, lower: fama_val }),
        ]
    }
}

impl Indicator for MesaAdaptiveMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::MesaAdaptiveMovingAverage,
            &self.line.mnemonic,
            &self.line.description,
            &[
                OutputText {
                    mnemonic: self.line.mnemonic.clone(),
                    description: self.line.description.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_fama.clone(),
                    description: self.description_fama.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_band.clone(),
                    description: self.description_band.clone(),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.line.bar_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.line.quote_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.line.trade_func)(sample);
        self.update_entity(sample.time, v)
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    use crate::indicators::core::outputs::shape::Shape;

    const EPSILON: f64 = 1e-9;
    const L_PRIMED: usize = 26;

    fn default_ce_params() -> CycleEstimatorParams {
        CycleEstimatorParams {
            smoothing_length: 4,
            alpha_ema_quadrature_in_phase: 0.2,
            alpha_ema_period: 0.2,
            warm_up_period: 0,
        }
    }

    fn create_length(fast: i64, slow: i64) -> MesaAdaptiveMovingAverage {
        MesaAdaptiveMovingAverage::new_length(&MesaAdaptiveMovingAverageLengthParams {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: default_ce_params(),
            fast_limit_length: fast,
            slow_limit_length: slow,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }).unwrap()
    }

    fn create_alpha(fast: f64, slow: f64) -> MesaAdaptiveMovingAverage {
        MesaAdaptiveMovingAverage::new_smoothing_factor(&MesaAdaptiveMovingAverageSmoothingFactorParams {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: default_ce_params(),
            fast_limit_smoothing_factor: fast,
            slow_limit_smoothing_factor: slow,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }).unwrap()
    }
    #[test]
    fn test_update_mama() {
        let input = testdata::test_input();
        let expected = testdata::test_expected_mama();
        let mut mama = create_length(3, 39);

        for i in 0..L_PRIMED {
            assert!(mama.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in L_PRIMED..input.len() {
            let act = mama.update(input[i]);
            assert!(
                (act - expected[i]).abs() <= EPSILON,
                "[{}] mama expected {}, got {}", i, expected[i], act
            );
        }

        assert!(mama.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_fama() {
        let input = testdata::test_input();
        let expected = testdata::test_expected_fama();
        let mut mama = create_length(3, 39);

        for i in 0..L_PRIMED {
            assert!(mama.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in L_PRIMED..input.len() {
            mama.update(input[i]);
            let act = mama.fama();
            assert!(
                (act - expected[i]).abs() <= EPSILON,
                "[{}] fama expected {}, got {}", i, expected[i], act
            );
        }
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut mama = create_length(3, 39);

        assert!(!mama.is_primed());

        for i in 0..L_PRIMED {
            mama.update(input[i]);
            assert!(!mama.is_primed(), "[{}] should not be primed", i);
        }

        for i in L_PRIMED..input.len() {
            mama.update(input[i]);
            assert!(mama.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_update_entity() {
        let time = 1617235200_i64;
        let inp = 3.0_f64;
        let expected_mama_val = 1.5;
        let expected_fama_val = 0.375;

        // Scalar
        let mut mama = create_length(3, 39);
        for _ in 0..L_PRIMED {
            mama.update(0.0);
        }
        let out = mama.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 3);
        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        let s1 = out[1].downcast_ref::<Scalar>().unwrap();
        let b = out[2].downcast_ref::<Band>().unwrap();
        assert_eq!(s0.time, time);
        assert_eq!(s0.value, expected_mama_val);
        assert_eq!(s1.time, time);
        assert_eq!(s1.value, expected_fama_val);
        assert_eq!(b.time, time);
        assert_eq!(b.upper, expected_mama_val);
        assert_eq!(b.lower, expected_fama_val);

        // Bar
        let mut mama = create_length(3, 39);
        for _ in 0..L_PRIMED {
            mama.update(0.0);
        }
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = mama.update_bar(&bar);
        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.value, expected_mama_val);

        // Quote
        let mut mama = create_length(3, 39);
        for _ in 0..L_PRIMED {
            mama.update(0.0);
        }
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = mama.update_quote(&quote);
        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.value, expected_mama_val);

        // Trade
        let mut mama = create_length(3, 39);
        for _ in 0..L_PRIMED {
            mama.update(0.0);
        }
        let trade = Trade::new(time, inp, 0.0);
        let out = mama.update_trade(&trade);
        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.value, expected_mama_val);
    }

    #[test]
    fn test_metadata_length() {
        let mama = create_length(2, 40);
        let m = mama.metadata();

        assert_eq!(m.identifier, Identifier::MesaAdaptiveMovingAverage);
        assert_eq!(m.mnemonic, "mama(2, 40)");
        assert_eq!(m.description, "Mesa adaptive moving average mama(2, 40)");
        assert_eq!(m.outputs.len(), 3);
        assert_eq!(m.outputs[0].kind, MesaAdaptiveMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "mama(2, 40)");
        assert_eq!(m.outputs[1].kind, MesaAdaptiveMovingAverageOutput::Fama as i32);
        assert_eq!(m.outputs[1].shape, Shape::Scalar);
        assert_eq!(m.outputs[1].mnemonic, "fama(2, 40)");
        assert_eq!(m.outputs[2].kind, MesaAdaptiveMovingAverageOutput::Band as i32);
        assert_eq!(m.outputs[2].shape, Shape::Band);
        assert_eq!(m.outputs[2].mnemonic, "mama-fama(2, 40)");
    }

    #[test]
    fn test_metadata_alpha() {
        let mama = create_alpha(0.666666666, 0.064516129);
        let m = mama.metadata();

        assert_eq!(m.mnemonic, "mama(0.667, 0.065)");
        assert_eq!(m.outputs[1].mnemonic, "fama(0.667, 0.065)");
        assert_eq!(m.outputs[2].mnemonic, "mama-fama(0.667, 0.065)");
    }

    #[test]
    fn test_new_default() {
        let mama = MesaAdaptiveMovingAverage::new_default().unwrap();
        assert_eq!(mama.line.mnemonic, "mama(3, 39)");
        assert!(!mama.primed);
    }

    #[test]
    fn test_new_length_validation() {
        // fast limit < 2
        let r = MesaAdaptiveMovingAverage::new_length(&MesaAdaptiveMovingAverageLengthParams {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: default_ce_params(),
            fast_limit_length: 1, slow_limit_length: 39,
            bar_component: None, quote_component: None, trade_component: None,
        });
        assert!(r.is_err());
        assert!(r.unwrap_err().contains("fast limit length"));

        // slow limit < 2
        let r = MesaAdaptiveMovingAverage::new_length(&MesaAdaptiveMovingAverageLengthParams {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: default_ce_params(),
            fast_limit_length: 3, slow_limit_length: 1,
            bar_component: None, quote_component: None, trade_component: None,
        });
        assert!(r.is_err());
        assert!(r.unwrap_err().contains("slow limit length"));
    }

    #[test]
    fn test_new_smoothing_factor_validation() {
        // fast < 0
        let r = MesaAdaptiveMovingAverage::new_smoothing_factor(&MesaAdaptiveMovingAverageSmoothingFactorParams {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: default_ce_params(),
            fast_limit_smoothing_factor: -0.00000001, slow_limit_smoothing_factor: 0.33,
            bar_component: None, quote_component: None, trade_component: None,
        });
        assert!(r.is_err());
        assert!(r.unwrap_err().contains("fast limit smoothing factor"));

        // fast > 1
        let r = MesaAdaptiveMovingAverage::new_smoothing_factor(&MesaAdaptiveMovingAverageSmoothingFactorParams {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: default_ce_params(),
            fast_limit_smoothing_factor: 1.00000001, slow_limit_smoothing_factor: 0.33,
            bar_component: None, quote_component: None, trade_component: None,
        });
        assert!(r.is_err());
        assert!(r.unwrap_err().contains("fast limit smoothing factor"));

        // slow < 0
        let r = MesaAdaptiveMovingAverage::new_smoothing_factor(&MesaAdaptiveMovingAverageSmoothingFactorParams {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: default_ce_params(),
            fast_limit_smoothing_factor: 0.66, slow_limit_smoothing_factor: -0.00000001,
            bar_component: None, quote_component: None, trade_component: None,
        });
        assert!(r.is_err());
        assert!(r.unwrap_err().contains("slow limit smoothing factor"));

        // slow > 1
        let r = MesaAdaptiveMovingAverage::new_smoothing_factor(&MesaAdaptiveMovingAverageSmoothingFactorParams {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: default_ce_params(),
            fast_limit_smoothing_factor: 0.66, slow_limit_smoothing_factor: 1.00000001,
            bar_component: None, quote_component: None, trade_component: None,
        });
        assert!(r.is_err());
        assert!(r.unwrap_err().contains("slow limit smoothing factor"));
    }

    #[test]
    fn test_estimator_moniker_in_mnemonic() {
        // Non-default smoothing length
        let mama = MesaAdaptiveMovingAverage::new_length(&MesaAdaptiveMovingAverageLengthParams {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: CycleEstimatorParams { smoothing_length: 3, ..default_ce_params() },
            fast_limit_length: 2, slow_limit_length: 40,
            bar_component: Some(BarComponent::Median), quote_component: None, trade_component: None,
        }).unwrap();
        assert_eq!(mama.line.mnemonic, "mama(2, 40, hd(3, 0.200, 0.200), hl/2)");

        // Unrolled
        let mama = MesaAdaptiveMovingAverage::new_length(&MesaAdaptiveMovingAverageLengthParams {
            estimator_type: CycleEstimatorType::HomodyneDiscriminatorUnrolled,
            estimator_params: default_ce_params(),
            fast_limit_length: 2, slow_limit_length: 40,
            bar_component: Some(BarComponent::Median), quote_component: None, trade_component: None,
        }).unwrap();
        assert_eq!(mama.line.mnemonic, "mama(2, 40, hdu(4, 0.200, 0.200), hl/2)");

        // Phase accumulator
        let mama = MesaAdaptiveMovingAverage::new_length(&MesaAdaptiveMovingAverageLengthParams {
            estimator_type: CycleEstimatorType::PhaseAccumulator,
            estimator_params: default_ce_params(),
            fast_limit_length: 2, slow_limit_length: 40,
            bar_component: Some(BarComponent::Median), quote_component: None, trade_component: None,
        }).unwrap();
        assert_eq!(mama.line.mnemonic, "mama(2, 40, pa(4, 0.200, 0.200), hl/2)");

        // Dual differentiator
        let mama = MesaAdaptiveMovingAverage::new_length(&MesaAdaptiveMovingAverageLengthParams {
            estimator_type: CycleEstimatorType::DualDifferentiator,
            estimator_params: default_ce_params(),
            fast_limit_length: 2, slow_limit_length: 40,
            bar_component: Some(BarComponent::Median), quote_component: None, trade_component: None,
        }).unwrap();
        assert_eq!(mama.line.mnemonic, "mama(2, 40, dd(4, 0.200, 0.200), hl/2)");
    }
}
