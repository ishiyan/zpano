use std::f64::consts::PI;

use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent};
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
use crate::indicators::core::metadata::Metadata;
use crate::indicators::core::outputs::band::Band;
use crate::indicators::john_ehlers::hilbert_transformer::{
    new_cycle_estimator, estimator_moniker, CycleEstimator, CycleEstimatorParams,
    CycleEstimatorType,
};

const DEG2RAD: f64 = PI / 180.0;
const RAD2DEG: f64 = 180.0 / PI;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Sine Wave indicator.
pub struct SineWaveParams {
    /// The type of cycle estimator to use.
    pub estimator_type: CycleEstimatorType,
    /// Parameters for the Hilbert transformer cycle estimator.
    pub estimator_params: CycleEstimatorParams,
    /// Alpha for additional EMA smoothing of the instantaneous period. Must be in (0, 1].
    /// Default is 0.33.
    pub alpha_ema_period_additional: f64,
    /// Bar component. `None` means default (Median for SineWave).
    pub bar_component: Option<BarComponent>,
    /// Quote component. `None` means default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component. `None` means default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for SineWaveParams {
    fn default() -> Self {
        Self {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: CycleEstimatorParams {
                smoothing_length: 4,
                alpha_ema_quadrature_in_phase: 0.2,
                alpha_ema_period: 0.2,
                warm_up_period: 0,
            },
            alpha_ema_period_additional: 0.33,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Sine Wave indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SineWaveOutput {
    /// The sine wave value, sin(phase * Deg2Rad).
    Value = 1,
    /// The sine wave lead value, sin((phase + 45) * Deg2Rad).
    Lead = 2,
    /// The band formed by the sine wave (upper) and the lead sine wave (lower).
    Band = 3,
    /// The smoothed dominant cycle period.
    DominantCyclePeriod = 4,
    /// The dominant cycle phase, in degrees.
    DominantCyclePhase = 5,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Sine Wave indicator.
///
/// Exposes five outputs: Value, Lead, Band, DominantCyclePeriod, DominantCyclePhase.
///
/// Reference: John Ehlers, Rocket Science for Traders, Wiley, 2001, pp 95-105.
pub struct SineWave {
    mnemonic: String,
    description: String,
    mnemonic_lead: String,
    description_lead: String,
    mnemonic_band: String,
    description_band: String,
    mnemonic_dcp: String,
    description_dcp: String,
    mnemonic_dc_phase: String,
    description_dc_phase: String,
    // Dominant cycle state
    alpha_ema_period_additional: f64,
    one_min_alpha: f64,
    smoothed_period: f64,
    smoothed_phase: f64,
    smoothed_input: Vec<f64>,
    smoothed_input_len_min1: usize,
    htce: Box<dyn CycleEstimator>,
    dc_primed: bool,
    // Sine wave state
    primed: bool,
    value: f64,
    lead: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl SineWave {
    /// Creates a new Sine Wave with default parameters.
    pub fn new_default() -> Result<Self, String> {
        Self::new(&SineWaveParams {
            estimator_params: CycleEstimatorParams {
                smoothing_length: 4,
                alpha_ema_quadrature_in_phase: 0.2,
                alpha_ema_period: 0.2,
                warm_up_period: 100,
            },
            alpha_ema_period_additional: 0.33,
            ..Default::default()
        })
    }

    /// Creates a new Sine Wave with the given parameters.
    pub fn new(p: &SineWaveParams) -> Result<Self, String> {
        const INVALID: &str = "invalid sine wave parameters";

        let alpha = p.alpha_ema_period_additional;
        if alpha <= 0.0 || alpha > 1.0 {
            return Err(format!(
                "{}: α for additional smoothing should be in range (0, 1]",
                INVALID
            ));
        }

        // SineWave defaults to BarMedianPrice (not the framework default).
        let bc = p.bar_component.unwrap_or(BarComponent::Median);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        // Build the inner cycle estimator. Pass the params through dominant cycle
        // which validates them and forwards errors.
        let estimator = new_cycle_estimator(p.estimator_type, &p.estimator_params)
            .map_err(|e| format!("{}: {}", INVALID, e))?;

        // Compose estimator moniker.
        let estimator_moniker = {
            let default_type = p.estimator_type == CycleEstimatorType::HomodyneDiscriminator
                && p.estimator_params.smoothing_length == 4
                && p.estimator_params.alpha_ema_quadrature_in_phase == 0.2
                && p.estimator_params.alpha_ema_period == 0.2;
            if default_type {
                String::new()
            } else {
                let m = estimator_moniker(p.estimator_type, estimator.as_ref());
                if m.is_empty() {
                    String::new()
                } else {
                    format!(", {}", m)
                }
            }
        };

        let comp_mn = component_triple_mnemonic(bc, qc, tc);

        // SineWave always shows hl/2 in mnemonic (since it differs from framework default Close).
        let mnemonic = format!("sw({:.3}{}{})", alpha, estimator_moniker, comp_mn);
        let mnemonic_lead = format!("sw-lead({:.3}{}{})", alpha, estimator_moniker, comp_mn);
        let mnemonic_band = format!("sw-band({:.3}{}{})", alpha, estimator_moniker, comp_mn);
        let mnemonic_dcp = format!("dcp({:.3}{}{})", alpha, estimator_moniker, comp_mn);
        let mnemonic_dc_phase = format!("dcph({:.3}{}{})", alpha, estimator_moniker, comp_mn);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let max_period = estimator.max_period();

        Ok(Self {
            description: format!("Sine wave {}", mnemonic),
            description_lead: format!("Sine wave lead {}", mnemonic_lead),
            description_band: format!("Sine wave band {}", mnemonic_band),
            description_dcp: format!("Dominant cycle period {}", mnemonic_dcp),
            description_dc_phase: format!("Dominant cycle phase {}", mnemonic_dc_phase),
            mnemonic,
            mnemonic_lead,
            mnemonic_band,
            mnemonic_dcp,
            mnemonic_dc_phase,
            alpha_ema_period_additional: alpha,
            one_min_alpha: 1.0 - alpha,
            smoothed_period: 0.0,
            smoothed_phase: 0.0,
            smoothed_input: vec![0.0; max_period],
            smoothed_input_len_min1: max_period - 1,
            htce: estimator,
            dc_primed: false,
            primed: false,
            value: f64::NAN,
            lead: f64::NAN,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Core update. Returns (value, lead, period, phase). NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> (f64, f64, f64, f64) {
        if sample.is_nan() {
            return (f64::NAN, f64::NAN, f64::NAN, f64::NAN);
        }

        // Update the inner dominant cycle.
        let (period, phase) = self.update_dominant_cycle(sample);

        if phase.is_nan() {
            return (f64::NAN, f64::NAN, f64::NAN, f64::NAN);
        }

        const LEAD_OFFSET: f64 = 45.0;

        self.primed = true;
        self.value = (phase * DEG2RAD).sin();
        self.lead = ((phase + LEAD_OFFSET) * DEG2RAD).sin();

        (self.value, self.lead, period, phase)
    }

    // --- DominantCycle logic inlined ---

    fn update_dominant_cycle(&mut self, sample: f64) -> (f64, f64) {
        self.htce.update(sample);
        self.push_smoothed_input(self.htce.smoothed());

        if self.dc_primed {
            self.smoothed_period = self.alpha_ema_period_additional * self.htce.period()
                + self.one_min_alpha * self.smoothed_period;
            self.calculate_smoothed_phase();
            return (self.smoothed_period, self.smoothed_phase);
        }

        if self.htce.primed() {
            self.dc_primed = true;
            self.smoothed_period = self.htce.period();
            self.calculate_smoothed_phase();
            return (self.smoothed_period, self.smoothed_phase);
        }

        (f64::NAN, f64::NAN)
    }

    fn push_smoothed_input(&mut self, value: f64) {
        let len_min1 = self.smoothed_input_len_min1;
        // Shift right by 1.
        for i in (1..=len_min1).rev() {
            self.smoothed_input[i] = self.smoothed_input[i - 1];
        }
        self.smoothed_input[0] = value;
    }

    fn calculate_smoothed_phase(&mut self) {
        const TWO_PI: f64 = 2.0 * PI;
        const EPSILON: f64 = 0.01;
        const NINETY: f64 = 90.0;
        const ONE_EIGHTY: f64 = 180.0;
        const THREE_SIXTY: f64 = 360.0;

        let mut length = (self.smoothed_period + 0.5).floor() as usize;
        if length > self.smoothed_input_len_min1 {
            length = self.smoothed_input_len_min1;
        }

        let mut real_part = 0.0_f64;
        let mut imag_part = 0.0_f64;

        for i in 0..length {
            let temp = TWO_PI * (i as f64) / (length as f64);
            let smoothed = self.smoothed_input[i];
            real_part += smoothed * temp.sin();
            imag_part += smoothed * temp.cos();
        }

        let previous = self.smoothed_phase;
        let mut phase = (real_part / imag_part).atan() * RAD2DEG;
        if phase.is_nan() || phase.is_infinite() {
            phase = previous;
        }

        if imag_part.abs() <= EPSILON {
            if real_part > 0.0 {
                phase += NINETY;
            } else if real_part < 0.0 {
                phase -= NINETY;
            }
        }

        // Introduce the 90 degree reference shift.
        phase += NINETY;
        // Compensate for one bar lag.
        phase += THREE_SIXTY / self.smoothed_period;
        // Resolve phase ambiguity.
        if imag_part < 0.0 {
            phase += ONE_EIGHTY;
        }
        // Cycle wraparound.
        if phase > THREE_SIXTY {
            phase -= THREE_SIXTY;
        }

        self.smoothed_phase = phase;
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let (value, lead, period, phase) = self.update(sample);
        vec![
            Box::new(Scalar::new(time, value)),
            Box::new(Scalar::new(time, lead)),
            Box::new(Band { time, upper: value, lower: lead }),
            Box::new(Scalar::new(time, period)),
            Box::new(Scalar::new(time, phase)),
        ]
    }
}

impl Indicator for SineWave {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::SineWave,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: self.mnemonic.clone(),
                    description: self.description.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_lead.clone(),
                    description: self.description_lead.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_band.clone(),
                    description: self.description_band.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_dcp.clone(),
                    description: self.description_dcp.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_dc_phase.clone(),
                    description: self.description_dc_phase.clone(),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        self.update_entity(sample.time, (self.bar_func)(sample))
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        self.update_entity(sample.time, (self.quote_func)(sample))
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        self.update_entity(sample.time, (self.trade_func)(sample))
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

    const TOLERANCE: f64 = 1e-4;
    const SKIP: usize = 9;
    const SETTLE_SKIP: usize = 177;

    fn create_default() -> SineWave {
        SineWave::new_default().unwrap()
    }

    fn create_alpha(alpha: f64, estimator_type: CycleEstimatorType) -> SineWave {
        SineWave::new(&SineWaveParams {
            estimator_type,
            estimator_params: CycleEstimatorParams {
                smoothing_length: 4,
                alpha_ema_quadrature_in_phase: 0.2,
                alpha_ema_period: 0.2,
                warm_up_period: 0,
            },
            alpha_ema_period_additional: alpha,
            ..Default::default()
        })
        .unwrap()
    }

    /// Returns the shortest signed angular difference in (-180, 180].
    fn phase_diff(a: f64, b: f64) -> f64 {
        let mut d = (a - b) % 360.0;
        if d > 180.0 {
            d -= 360.0;
        } else if d <= -180.0 {
            d += 360.0;
        }
        d
    }
    #[test]
    fn test_reference_sine() {
        let input = testdata::test_input();
        let exp = testdata::test_expected_sine();
        let mut sw = create_default();

        for i in SKIP..input.len() {
            let (value, _, _, _) = sw.update(input[i]);
            if value.is_nan() || i < SETTLE_SKIP {
                continue;
            }
            if exp[i].is_nan() {
                continue;
            }
            assert!(
                (exp[i] - value).abs() <= TOLERANCE,
                "[{}] sine: expected {}, got {}",
                i, exp[i], value
            );
        }
    }

    #[test]
    fn test_reference_sine_lead() {
        let input = testdata::test_input();
        let exp = testdata::test_expected_sine_lead();
        let mut sw = create_default();

        for i in SKIP..input.len() {
            let (_, lead, _, _) = sw.update(input[i]);
            if lead.is_nan() || i < SETTLE_SKIP {
                continue;
            }
            if exp[i].is_nan() {
                continue;
            }
            assert!(
                (exp[i] - lead).abs() <= TOLERANCE,
                "[{}] sine lead: expected {}, got {}",
                i, exp[i], lead
            );
        }
    }

    #[test]
    fn test_reference_period() {
        let input = testdata::test_input();
        let exp = testdata::test_expected_period();
        let mut sw = create_default();

        for i in SKIP..input.len() {
            let (_, _, period, _) = sw.update(input[i]);
            if period.is_nan() || i < SETTLE_SKIP {
                continue;
            }
            assert!(
                (exp[i] - period).abs() <= TOLERANCE,
                "[{}] period: expected {}, got {}",
                i, exp[i], period
            );
        }
    }

    #[test]
    fn test_reference_phase() {
        let input = testdata::test_input();
        let exp = testdata::test_expected_phase();
        let mut sw = create_default();

        for i in SKIP..input.len() {
            let (_, _, _, phase) = sw.update(input[i]);
            if phase.is_nan() || i < SETTLE_SKIP {
                continue;
            }
            if exp[i].is_nan() {
                continue;
            }
            assert!(
                phase_diff(exp[i], phase).abs() <= TOLERANCE,
                "[{}] phase: expected {}, got {}",
                i, exp[i], phase
            );
        }
    }

    #[test]
    fn test_nan_input() {
        let mut sw = create_default();
        let (v, l, p, ph) = sw.update(f64::NAN);
        assert!(v.is_nan());
        assert!(l.is_nan());
        assert!(p.is_nan());
        assert!(ph.is_nan());
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut sw = create_default();

        assert!(!sw.is_primed());

        let mut primed_at: Option<usize> = None;
        for i in 0..input.len() {
            sw.update(input[i]);
            if sw.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert!(primed_at.is_some(), "should become primed");
        assert!(sw.is_primed());
    }

    #[test]
    fn test_update_entity() {
        let input = testdata::test_input();
        let time = 1617235200_i64;
        let prime_count = 200;

        // Scalar
        {
            let mut sw = create_default();
            for i in 0..prime_count {
                sw.update(input[i % input.len()]);
            }
            let s = Scalar::new(time, 100.0);
            let out = sw.update_scalar(&s);
            assert_eq!(out.len(), 5);
            // Outputs 0, 1, 3, 4 are Scalar; output 2 is Band.
            for &idx in &[0usize, 1, 3, 4] {
                let sc = out[idx].downcast_ref::<Scalar>().unwrap();
                assert_eq!(sc.time, time);
            }
            let band = out[2].downcast_ref::<Band>().unwrap();
            assert_eq!(band.time, time);
        }

        // Bar
        {
            let mut sw = create_default();
            for i in 0..prime_count {
                sw.update(input[i % input.len()]);
            }
            let bar = Bar::new(time, 0.0, 100.0, 100.0, 100.0, 0.0);
            let out = sw.update_bar(&bar);
            assert_eq!(out.len(), 5);
        }

        // Quote
        {
            let mut sw = create_default();
            for i in 0..prime_count {
                sw.update(input[i % input.len()]);
            }
            let q = Quote::new(time, 100.0, 100.0, 0.0, 0.0);
            let out = sw.update_quote(&q);
            assert_eq!(out.len(), 5);
        }

        // Trade
        {
            let mut sw = create_default();
            for i in 0..prime_count {
                sw.update(input[i % input.len()]);
            }
            let t = Trade::new(time, 100.0, 0.0);
            let out = sw.update_trade(&t);
            assert_eq!(out.len(), 5);
        }
    }

    #[test]
    fn test_band_ordering() {
        let input = testdata::test_input();
        let mut sw = create_default();
        let time = 1617235200_i64;

        for i in 0..200 {
            sw.update(input[i % input.len()]);
        }

        let s = Scalar::new(time, input[0]);
        let out = sw.update_scalar(&s);

        let value = out[0].downcast_ref::<Scalar>().unwrap().value;
        let lead = out[1].downcast_ref::<Scalar>().unwrap().value;
        let band = out[2].downcast_ref::<Band>().unwrap();

        assert_eq!(band.upper, value);
        assert_eq!(band.lower, lead);
    }

    #[test]
    fn test_metadata_default() {
        let sw = create_default();
        let m = sw.metadata();

        assert_eq!(m.identifier, Identifier::SineWave);
        assert_eq!(m.mnemonic, "sw(0.330, hl/2)");
        assert_eq!(m.description, "Sine wave sw(0.330, hl/2)");
        assert_eq!(m.outputs.len(), 5);

        assert_eq!(m.outputs[0].kind, SineWaveOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "sw(0.330, hl/2)");

        assert_eq!(m.outputs[1].kind, SineWaveOutput::Lead as i32);
        assert_eq!(m.outputs[1].shape, Shape::Scalar);
        assert_eq!(m.outputs[1].mnemonic, "sw-lead(0.330, hl/2)");

        assert_eq!(m.outputs[2].kind, SineWaveOutput::Band as i32);
        assert_eq!(m.outputs[2].shape, Shape::Band);
        assert_eq!(m.outputs[2].mnemonic, "sw-band(0.330, hl/2)");

        assert_eq!(m.outputs[3].kind, SineWaveOutput::DominantCyclePeriod as i32);
        assert_eq!(m.outputs[3].shape, Shape::Scalar);
        assert_eq!(m.outputs[3].mnemonic, "dcp(0.330, hl/2)");

        assert_eq!(m.outputs[4].kind, SineWaveOutput::DominantCyclePhase as i32);
        assert_eq!(m.outputs[4].shape, Shape::Scalar);
        assert_eq!(m.outputs[4].mnemonic, "dcph(0.330, hl/2)");
    }

    #[test]
    fn test_metadata_phase_accumulator() {
        let sw = create_alpha(0.5, CycleEstimatorType::PhaseAccumulator);
        let m = sw.metadata();
        assert_eq!(m.mnemonic, "sw(0.500, pa(4, 0.200, 0.200), hl/2)");
    }

    #[test]
    fn test_new_validation() {
        // alpha <= 0
        let r = SineWave::new(&SineWaveParams {
            alpha_ema_period_additional: 0.0,
            ..Default::default()
        });
        assert!(r.is_err());
        assert!(r.err().unwrap().contains("α for additional smoothing"));

        // alpha > 1
        let r = SineWave::new(&SineWaveParams {
            alpha_ema_period_additional: 1.00000001,
            ..Default::default()
        });
        assert!(r.is_err());
        assert!(r.err().unwrap().contains("α for additional smoothing"));

        // valid alpha = 1.0 (boundary)
        let r = SineWave::new(&SineWaveParams {
            alpha_ema_period_additional: 1.0,
            ..Default::default()
        });
        assert!(r.is_ok());
    }
}
