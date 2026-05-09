use std::any::Any;
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
use crate::indicators::john_ehlers::hilbert_transformer::{
    new_cycle_estimator, estimator_moniker, CycleEstimator, CycleEstimatorParams,
    CycleEstimatorType,
};

const DEG2RAD: f64 = PI / 180.0;
const EPSILON: f64 = 1e-308;

/// Output describes the outputs of the indicator.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum TrendCycleModeOutput {
    Value = 1,
    IsTrendMode = 2,
    IsCycleMode = 3,
    InstantaneousTrendLine = 4,
    SineWave = 5,
    SineWaveLead = 6,
    DominantCyclePeriod = 7,
    DominantCyclePhase = 8,
}

/// Params for creating a TrendCycleMode indicator.
pub struct TrendCycleModeParams {
    pub estimator_type: CycleEstimatorType,
    pub estimator_params: CycleEstimatorParams,
    pub alpha_ema_period_additional: f64,
    pub trend_line_smoothing_length: usize,
    pub cycle_part_multiplier: f64,
    pub separation_percentage: f64,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for TrendCycleModeParams {
    fn default() -> Self {
        Self {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: CycleEstimatorParams {
                smoothing_length: 4,
                alpha_ema_quadrature_in_phase: 0.2,
                alpha_ema_period: 0.2,
                warm_up_period: 100,
            },
            alpha_ema_period_additional: 0.33,
            trend_line_smoothing_length: 4,
            cycle_part_multiplier: 1.0,
            separation_percentage: 1.5,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Internal DominantCycle helper that wraps a CycleEstimator.
struct DominantCycleInner {
    alpha_ema_period_additional: f64,
    one_min_alpha: f64,
    smoothed_period: f64,
    smoothed_phase: f64,
    smoothed_input: Vec<f64>,
    smoothed_input_length_min1: usize,
    htce: Box<dyn CycleEstimator>,
    primed: bool,
}

impl DominantCycleInner {
    fn new(
        estimator_type: CycleEstimatorType,
        estimator_params: &CycleEstimatorParams,
        alpha_ema_period_additional: f64,
    ) -> Result<Self, String> {
        let htce = new_cycle_estimator(estimator_type, estimator_params)?;
        let max_period = htce.max_period();
        Ok(Self {
            alpha_ema_period_additional,
            one_min_alpha: 1.0 - alpha_ema_period_additional,
            smoothed_period: 0.0,
            smoothed_phase: 0.0,
            smoothed_input: vec![0.0; max_period],
            smoothed_input_length_min1: max_period - 1,
            htce,
            primed: false,
        })
    }

    fn is_primed(&self) -> bool {
        self.primed
    }

    fn smoothed_price(&self) -> f64 {
        if !self.primed {
            return f64::NAN;
        }
        self.htce.smoothed()
    }

    fn max_period(&self) -> usize {
        self.htce.max_period()
    }

    /// Returns (raw_period, smoothed_period, phase). NaN if not primed.
    fn update(&mut self, sample: f64) -> (f64, f64, f64) {
        self.htce.update(sample);
        self.push_smoothed_input(self.htce.smoothed());

        if self.primed {
            self.smoothed_period = self.alpha_ema_period_additional * self.htce.period()
                + self.one_min_alpha * self.smoothed_period;
            self.calculate_smoothed_phase();
            return (self.htce.period(), self.smoothed_period, self.smoothed_phase);
        }

        if self.htce.primed() {
            self.primed = true;
            self.smoothed_period = self.htce.period();
            self.calculate_smoothed_phase();
            return (self.htce.period(), self.smoothed_period, self.smoothed_phase);
        }

        (f64::NAN, f64::NAN, f64::NAN)
    }

    fn push_smoothed_input(&mut self, value: f64) {
        let len_min1 = self.smoothed_input_length_min1;
        for i in (1..=len_min1).rev() {
            self.smoothed_input[i] = self.smoothed_input[i - 1];
        }
        self.smoothed_input[0] = value;
    }

    fn calculate_smoothed_phase(&mut self) {
        const RAD2DEG: f64 = 180.0 / PI;
        const TWO_PI: f64 = 2.0 * PI;
        const PHASE_EPSILON: f64 = 0.01;

        let length = (self.smoothed_period + 0.5).floor() as usize;
        let length = length.min(self.smoothed_input_length_min1);

        let mut real_part = 0.0_f64;
        let mut imag_part = 0.0_f64;

        for i in 0..length {
            let temp = TWO_PI * i as f64 / length as f64;
            let smoothed = self.smoothed_input[i];
            real_part += smoothed * temp.sin();
            imag_part += smoothed * temp.cos();
        }

        let previous = self.smoothed_phase;
        let mut phase = (real_part / imag_part).atan() * RAD2DEG;
        if phase.is_nan() || phase.is_infinite() {
            phase = previous;
        }

        if imag_part.abs() <= PHASE_EPSILON {
            if real_part > 0.0 {
                phase += 90.0;
            } else if real_part < 0.0 {
                phase -= 90.0;
            }
        }

        phase += 90.0;
        phase += 360.0 / self.smoothed_period;

        if imag_part < 0.0 {
            phase += 180.0;
        }
        if phase > 360.0 {
            phase -= 360.0;
        }

        self.smoothed_phase = phase;
    }
}

/// TrendCycleMode is Ehlers' Trend-versus-Cycle Mode indicator.
pub struct TrendCycleMode {
    mnemonic: String,
    description: String,
    mnemonic_trend: String,
    description_trend: String,
    mnemonic_cycle: String,
    description_cycle: String,
    mnemonic_itl: String,
    description_itl: String,
    mnemonic_sine: String,
    description_sine: String,
    mnemonic_sine_lead: String,
    description_sine_lead: String,
    mnemonic_dcp: String,
    description_dcp: String,
    mnemonic_dc_phase: String,
    description_dc_phase: String,
    dc: DominantCycleInner,
    cycle_part_multiplier: f64,
    separation_factor: f64,
    coeff0: f64,
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    trendline: f64,
    trend_average1: f64,
    trend_average2: f64,
    trend_average3: f64,
    sin_wave: f64,
    sin_wave_lead: f64,
    previous_dc_phase: f64,
    previous_sine_lead_wave_difference: f64,
    samples_in_trend: i32,
    is_trend_mode: bool,
    input: Vec<f64>,
    input_length: usize,
    input_length_min1: usize,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl TrendCycleMode {
    /// Creates an instance with default parameters.
    pub fn new_default() -> Result<Self, String> {
        Self::new(&TrendCycleModeParams::default())
    }

    /// Creates an instance from supplied parameters.
    pub fn new(p: &TrendCycleModeParams) -> Result<Self, String> {
        let invalid = "invalid trend cycle mode parameters";

        if p.alpha_ema_period_additional <= 0.0 || p.alpha_ema_period_additional > 1.0 {
            return Err(format!(
                "{}: α for additional smoothing should be in range (0, 1]",
                invalid
            ));
        }

        if p.trend_line_smoothing_length < 2 || p.trend_line_smoothing_length > 4 {
            return Err(format!(
                "{}: trend line smoothing length should be 2, 3, or 4",
                invalid
            ));
        }

        if p.cycle_part_multiplier <= 0.0 || p.cycle_part_multiplier > 10.0 {
            return Err(format!(
                "{}: cycle part multiplier should be in range (0, 10]",
                invalid
            ));
        }

        if p.separation_percentage <= 0.0 || p.separation_percentage > 100.0 {
            return Err(format!(
                "{}: separation percentage should be in range (0, 100]",
                invalid
            ));
        }

        // Resolve bar component default: BarMedianPrice for this indicator.
        let bc = p.bar_component.unwrap_or(BarComponent::Median);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        // Build inner dominant cycle (passes explicit components).
        let dc = DominantCycleInner::new(
            p.estimator_type,
            &CycleEstimatorParams {
                smoothing_length: p.estimator_params.smoothing_length,
                alpha_ema_quadrature_in_phase: p.estimator_params.alpha_ema_quadrature_in_phase,
                alpha_ema_period: p.estimator_params.alpha_ema_period,
                warm_up_period: p.estimator_params.warm_up_period,
            },
            p.alpha_ema_period_additional,
        ).map_err(|e| format!("{}: {}", invalid, e))?;

        // Build estimator moniker for mnemonic.
        let est_temp = new_cycle_estimator(p.estimator_type, &CycleEstimatorParams {
            smoothing_length: p.estimator_params.smoothing_length,
            alpha_ema_quadrature_in_phase: p.estimator_params.alpha_ema_quadrature_in_phase,
            alpha_ema_period: p.estimator_params.alpha_ema_period,
            warm_up_period: p.estimator_params.warm_up_period,
        }).map_err(|e| format!("{}: {}", invalid, e))?;

        let est_moniker = if p.estimator_type != CycleEstimatorType::HomodyneDiscriminator
            || p.estimator_params.smoothing_length != 4
            || p.estimator_params.alpha_ema_quadrature_in_phase != 0.2
            || p.estimator_params.alpha_ema_period != 0.2
        {
            let m = estimator_moniker(p.estimator_type, est_temp.as_ref());
            if m.is_empty() { String::new() } else { format!(", {}", m) }
        } else {
            String::new()
        };

        let component_mnemonic = component_triple_mnemonic(bc, qc, tc);

        let mn_value = format!(
            "tcm({:.3}, {}, {:.3}, {:.3}%{}{})",
            p.alpha_ema_period_additional, p.trend_line_smoothing_length,
            p.cycle_part_multiplier, p.separation_percentage,
            est_moniker, component_mnemonic
        );
        let mn_trend = format!(
            "tcm-trend({:.3}, {}, {:.3}, {:.3}%{}{})",
            p.alpha_ema_period_additional, p.trend_line_smoothing_length,
            p.cycle_part_multiplier, p.separation_percentage,
            est_moniker, component_mnemonic
        );
        let mn_cycle = format!(
            "tcm-cycle({:.3}, {}, {:.3}, {:.3}%{}{})",
            p.alpha_ema_period_additional, p.trend_line_smoothing_length,
            p.cycle_part_multiplier, p.separation_percentage,
            est_moniker, component_mnemonic
        );
        let mn_itl = format!(
            "tcm-itl({:.3}, {}, {:.3}, {:.3}%{}{})",
            p.alpha_ema_period_additional, p.trend_line_smoothing_length,
            p.cycle_part_multiplier, p.separation_percentage,
            est_moniker, component_mnemonic
        );
        let mn_sine = format!(
            "tcm-sine({:.3}, {}, {:.3}, {:.3}%{}{})",
            p.alpha_ema_period_additional, p.trend_line_smoothing_length,
            p.cycle_part_multiplier, p.separation_percentage,
            est_moniker, component_mnemonic
        );
        let mn_sine_lead = format!(
            "tcm-sineLead({:.3}, {}, {:.3}, {:.3}%{}{})",
            p.alpha_ema_period_additional, p.trend_line_smoothing_length,
            p.cycle_part_multiplier, p.separation_percentage,
            est_moniker, component_mnemonic
        );
        let mn_dcp = format!(
            "dcp({:.3}{}{})",
            p.alpha_ema_period_additional, est_moniker, component_mnemonic
        );
        let mn_dc_phase = format!(
            "dcph({:.3}{}{})",
            p.alpha_ema_period_additional, est_moniker, component_mnemonic
        );

        let desc_value = format!("Trend versus cycle mode {}", mn_value);
        let desc_trend = format!("Trend versus cycle mode, is-trend flag {}", mn_trend);
        let desc_cycle = format!("Trend versus cycle mode, is-cycle flag {}", mn_cycle);
        let desc_itl = format!("Trend versus cycle mode instantaneous trend line {}", mn_itl);
        let desc_sine = format!("Trend versus cycle mode sine wave {}", mn_sine);
        let desc_sine_lead = format!("Trend versus cycle mode sine wave lead {}", mn_sine_lead);
        let desc_dcp = format!("Dominant cycle period {}", mn_dcp);
        let desc_dc_phase = format!("Dominant cycle phase {}", mn_dc_phase);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let (c0, c1, c2, c3) = match p.trend_line_smoothing_length {
            2 => (2.0 / 3.0, 1.0 / 3.0, 0.0, 0.0),
            3 => (3.0 / 6.0, 2.0 / 6.0, 1.0 / 6.0, 0.0),
            _ => (4.0 / 10.0, 3.0 / 10.0, 2.0 / 10.0, 1.0 / 10.0),
        };

        let max_period = dc.max_period();

        Ok(Self {
            mnemonic: mn_value,
            description: desc_value,
            mnemonic_trend: mn_trend,
            description_trend: desc_trend,
            mnemonic_cycle: mn_cycle,
            description_cycle: desc_cycle,
            mnemonic_itl: mn_itl,
            description_itl: desc_itl,
            mnemonic_sine: mn_sine,
            description_sine: desc_sine,
            mnemonic_sine_lead: mn_sine_lead,
            description_sine_lead: desc_sine_lead,
            mnemonic_dcp: mn_dcp,
            description_dcp: desc_dcp,
            mnemonic_dc_phase: mn_dc_phase,
            description_dc_phase: desc_dc_phase,
            dc,
            cycle_part_multiplier: p.cycle_part_multiplier,
            separation_factor: p.separation_percentage / 100.0,
            coeff0: c0,
            coeff1: c1,
            coeff2: c2,
            coeff3: c3,
            trendline: f64::NAN,
            trend_average1: 0.0,
            trend_average2: 0.0,
            trend_average3: 0.0,
            sin_wave: f64::NAN,
            sin_wave_lead: f64::NAN,
            previous_dc_phase: 0.0,
            previous_sine_lead_wave_difference: 0.0,
            samples_in_trend: 0,
            is_trend_mode: true,
            input: vec![0.0; max_period],
            input_length: max_period,
            input_length_min1: max_period - 1,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Updates the indicator, returning (value, is_trend, is_cycle, trendline, sine, sine_lead, period, phase).
    pub fn update(
        &mut self,
        sample: f64,
    ) -> (f64, f64, f64, f64, f64, f64, f64, f64) {
        if sample.is_nan() {
            return (sample, sample, sample, sample, sample, sample, sample, sample);
        }

        let (_, period, phase) = self.dc.update(sample);
        let smoothed_price = self.dc.smoothed_price();

        self.push_input(sample);

        if self.primed {
            let smoothed_period = period;
            let average = self.calculate_trend_average(smoothed_period);
            self.trendline = self.coeff0 * average
                + self.coeff1 * self.trend_average1
                + self.coeff2 * self.trend_average2
                + self.coeff3 * self.trend_average3;
            self.trend_average3 = self.trend_average2;
            self.trend_average2 = self.trend_average1;
            self.trend_average1 = average;

            let diff = self.calculate_sine_lead_wave_difference(phase);

            // Condition 1
            self.is_trend_mode = true;
            if (diff > 0.0 && self.previous_sine_lead_wave_difference < 0.0)
                || (diff < 0.0 && self.previous_sine_lead_wave_difference > 0.0)
            {
                self.is_trend_mode = false;
                self.samples_in_trend = 0;
            }
            self.previous_sine_lead_wave_difference = diff;
            self.samples_in_trend += 1;

            if (self.samples_in_trend as f64) < 0.5 * smoothed_period {
                self.is_trend_mode = false;
            }

            // Condition 2
            let phase_delta = phase - self.previous_dc_phase;
            self.previous_dc_phase = phase;

            if smoothed_period.abs() > EPSILON {
                let dc_rate = 360.0 / smoothed_period;
                if phase_delta > (2.0 / 3.0) * dc_rate && phase_delta < 1.5 * dc_rate {
                    self.is_trend_mode = false;
                }
            }

            // Condition 3
            if self.trendline.abs() > EPSILON
                && ((smoothed_price - self.trendline) / self.trendline).abs()
                    >= self.separation_factor
            {
                self.is_trend_mode = true;
            }

            return (
                self.mode_value(),
                self.is_trend_float(),
                self.is_cycle_float(),
                self.trendline,
                self.sin_wave,
                self.sin_wave_lead,
                period,
                phase,
            );
        }

        if self.dc.is_primed() {
            self.primed = true;
            let smoothed_period = period;
            self.trendline = self.calculate_trend_average(smoothed_period);
            self.trend_average1 = self.trendline;
            self.trend_average2 = self.trendline;
            self.trend_average3 = self.trendline;

            self.previous_dc_phase = phase;
            self.previous_sine_lead_wave_difference =
                self.calculate_sine_lead_wave_difference(phase);

            self.is_trend_mode = true;
            self.samples_in_trend += 1;

            if (self.samples_in_trend as f64) < 0.5 * smoothed_period {
                self.is_trend_mode = false;
            }

            return (
                self.mode_value(),
                self.is_trend_float(),
                self.is_cycle_float(),
                self.trendline,
                self.sin_wave,
                self.sin_wave_lead,
                period,
                phase,
            );
        }

        let nan = f64::NAN;
        (nan, nan, nan, nan, nan, nan, nan, nan)
    }

    fn push_input(&mut self, value: f64) {
        let len_min1 = self.input_length_min1;
        for i in (1..=len_min1).rev() {
            self.input[i] = self.input[i - 1];
        }
        self.input[0] = value;
    }

    fn calculate_trend_average(&self, smoothed_period: f64) -> f64 {
        let mut length = (smoothed_period * self.cycle_part_multiplier + 0.5).floor() as usize;
        if length > self.input_length {
            length = self.input_length;
        } else if length < 1 {
            length = 1;
        }

        let sum: f64 = self.input[..length].iter().sum();
        sum / length as f64
    }

    fn calculate_sine_lead_wave_difference(&mut self, phase: f64) -> f64 {
        let p = phase * DEG2RAD;
        self.sin_wave = p.sin();
        self.sin_wave_lead = (p + 45.0 * DEG2RAD).sin();
        self.sin_wave - self.sin_wave_lead
    }

    fn mode_value(&self) -> f64 {
        if self.is_trend_mode { 1.0 } else { -1.0 }
    }

    fn is_trend_float(&self) -> f64 {
        if self.is_trend_mode { 1.0 } else { 0.0 }
    }

    fn is_cycle_float(&self) -> f64 {
        if self.is_trend_mode { 0.0 } else { 1.0 }
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let (value, trend, cycle, itl, sine, sine_lead, period, phase) = self.update(sample);
        vec![
            Box::new(Scalar::new(time, value)) as Box<dyn Any>,
            Box::new(Scalar::new(time, trend)) as Box<dyn Any>,
            Box::new(Scalar::new(time, cycle)) as Box<dyn Any>,
            Box::new(Scalar::new(time, itl)) as Box<dyn Any>,
            Box::new(Scalar::new(time, sine)) as Box<dyn Any>,
            Box::new(Scalar::new(time, sine_lead)) as Box<dyn Any>,
            Box::new(Scalar::new(time, period)) as Box<dyn Any>,
            Box::new(Scalar::new(time, phase)) as Box<dyn Any>,
        ]
    }
}

impl Indicator for TrendCycleMode {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::TrendCycleMode,
            &self.mnemonic,
            &self.description,
            &[
                OutputText { mnemonic: self.mnemonic.clone(), description: self.description.clone() },
                OutputText { mnemonic: self.mnemonic_trend.clone(), description: self.description_trend.clone() },
                OutputText { mnemonic: self.mnemonic_cycle.clone(), description: self.description_cycle.clone() },
                OutputText { mnemonic: self.mnemonic_itl.clone(), description: self.description_itl.clone() },
                OutputText { mnemonic: self.mnemonic_sine.clone(), description: self.description_sine.clone() },
                OutputText { mnemonic: self.mnemonic_sine_lead.clone(), description: self.description_sine_lead.clone() },
                OutputText { mnemonic: self.mnemonic_dcp.clone(), description: self.description_dcp.clone() },
                OutputText { mnemonic: self.mnemonic_dc_phase.clone(), description: self.description_dc_phase.clone() },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        self.update_entity(sample.time, v)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;

    const TOLERANCE: f64 = 1e-4;
    const SKIP: usize = 9;
    const SETTLE_SKIP: usize = 177;

    fn create_default() -> TrendCycleMode {
        TrendCycleMode::new_default().unwrap()
    }
    #[test]
    fn test_reference_period() {
        let mut x = create_default();
        let input = testdata::test_input();
        let exp = testdata::test_expected_period();

        for i in SKIP..input.len() {
            let (_, _, _, _, _, _, period, _) = x.update(input[i]);
            if period.is_nan() || i < SETTLE_SKIP {
                continue;
            }
            assert!(
                (exp[i] - period).abs() < TOLERANCE,
                "[{}] period: expected {}, actual {}", i, exp[i], period
            );
        }
    }

    #[test]
    fn test_reference_phase() {
        let mut x = create_default();
        let input = testdata::test_input();
        let exp = testdata::test_expected_phase();

        for i in SKIP..input.len() {
            let (_, _, _, _, _, _, _, phase) = x.update(input[i]);
            if phase.is_nan() || exp[i].is_nan() || i < SETTLE_SKIP {
                continue;
            }
            let mut d = (exp[i] - phase) % 360.0;
            if d > 180.0 { d -= 360.0; }
            else if d < -180.0 { d += 360.0; }
            assert!(
                d.abs() < TOLERANCE,
                "[{}] phase: expected {}, actual {}", i, exp[i], phase
            );
        }
    }

    #[test]
    fn test_reference_sine_wave() {
        let mut x = create_default();
        let input = testdata::test_input();
        let exp = testdata::test_expected_sine();

        for i in SKIP..input.len() {
            let (_, _, _, _, sine, _, _, _) = x.update(input[i]);
            if sine.is_nan() || exp[i].is_nan() || i < SETTLE_SKIP {
                continue;
            }
            assert!(
                (exp[i] - sine).abs() < TOLERANCE,
                "[{}] sine: expected {}, actual {}", i, exp[i], sine
            );
        }
    }

    #[test]
    fn test_reference_sine_wave_lead() {
        let mut x = create_default();
        let input = testdata::test_input();
        let exp = testdata::test_expected_sine_lead();

        for i in SKIP..input.len() {
            let (_, _, _, _, _, sine_lead, _, _) = x.update(input[i]);
            if sine_lead.is_nan() || exp[i].is_nan() || i < SETTLE_SKIP {
                continue;
            }
            assert!(
                (exp[i] - sine_lead).abs() < TOLERANCE,
                "[{}] sine_lead: expected {}, actual {}", i, exp[i], sine_lead
            );
        }
    }

    #[test]
    fn test_reference_itl() {
        let mut x = create_default();
        let input = testdata::test_input();
        let exp = testdata::test_expected_itl();

        for i in SKIP..input.len() {
            let (_, _, _, itl, _, _, _, _) = x.update(input[i]);
            if itl.is_nan() || exp[i].is_nan() || i < SETTLE_SKIP {
                continue;
            }
            assert!(
                (exp[i] - itl).abs() < TOLERANCE,
                "[{}] itl: expected {}, actual {}", i, exp[i], itl
            );
        }
    }

    #[test]
    fn test_reference_value() {
        let mut x = create_default();
        let input = testdata::test_input();
        let exp = testdata::test_expected_value();
        let limit = exp.len();

        for i in SKIP..input.len() {
            let (value, _, _, _, _, _, _, _) = x.update(input[i]);
            if i >= limit { continue; }
            // MBST known mismatches.
            if i == 70 || i == 71 { continue; }
            if value.is_nan() || exp[i].is_nan() { continue; }
            assert!(
                (exp[i] - value).abs() < TOLERANCE,
                "[{}] value: expected {}, actual {}", i, exp[i], value
            );
        }
    }

    #[test]
    fn test_trend_cycle_complementary() {
        let mut x = create_default();
        let input = testdata::test_input();

        for i in SKIP..input.len() {
            let (value, trend, cycle, _, _, _, _, _) = x.update(input[i]);
            if value.is_nan() { continue; }
            assert_eq!(trend + cycle, 1.0, "[{}] trend+cycle should be 1", i);
            if value > 0.0 { assert_eq!(trend, 1.0, "[{}] trend should be 1 when value>0", i); }
            if value < 0.0 { assert_eq!(trend, 0.0, "[{}] trend should be 0 when value<0", i); }
        }
    }

    #[test]
    fn test_nan_input() {
        let mut x = create_default();
        let (v, t, c, itl, s, sl, p, ph) = x.update(f64::NAN);
        assert!(v.is_nan());
        assert!(t.is_nan());
        assert!(c.is_nan());
        assert!(itl.is_nan());
        assert!(s.is_nan());
        assert!(sl.is_nan());
        assert!(p.is_nan());
        assert!(ph.is_nan());
    }

    #[test]
    fn test_is_primed() {
        let mut x = create_default();
        let input = testdata::test_input();

        assert!(!x.is_primed());

        let mut primed_at: Option<usize> = None;
        for i in 0..input.len() {
            x.update(input[i]);
            if x.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert!(primed_at.is_some(), "should become primed");
        assert!(x.is_primed());
    }

    #[test]
    fn test_metadata() {
        let x = create_default();
        let m = x.metadata();

        assert_eq!(m.identifier, Identifier::TrendCycleMode);
        assert_eq!(m.mnemonic, "tcm(0.330, 4, 1.000, 1.500%, hl/2)");
        assert_eq!(m.description, "Trend versus cycle mode tcm(0.330, 4, 1.000, 1.500%, hl/2)");
        assert_eq!(m.outputs.len(), 8);
        assert_eq!(m.outputs[0].mnemonic, "tcm(0.330, 4, 1.000, 1.500%, hl/2)");
        assert_eq!(m.outputs[1].mnemonic, "tcm-trend(0.330, 4, 1.000, 1.500%, hl/2)");
        assert_eq!(m.outputs[2].mnemonic, "tcm-cycle(0.330, 4, 1.000, 1.500%, hl/2)");
        assert_eq!(m.outputs[3].mnemonic, "tcm-itl(0.330, 4, 1.000, 1.500%, hl/2)");
        assert_eq!(m.outputs[4].mnemonic, "tcm-sine(0.330, 4, 1.000, 1.500%, hl/2)");
        assert_eq!(m.outputs[5].mnemonic, "tcm-sineLead(0.330, 4, 1.000, 1.500%, hl/2)");
        assert_eq!(m.outputs[6].mnemonic, "dcp(0.330, hl/2)");
        assert_eq!(m.outputs[7].mnemonic, "dcph(0.330, hl/2)");
    }

    #[test]
    fn test_update_entity_scalar() {
        let mut x = create_default();
        let input = testdata::test_input();

        for i in 0..200 {
            x.update(input[i % input.len()]);
        }

        let s = Scalar::new(1000, 100.0);
        let out = x.update_scalar(&s);
        assert_eq!(out.len(), 8);
        for j in 0..8 {
            let sc = out[j].downcast_ref::<Scalar>().unwrap();
            assert_eq!(sc.time, 1000);
        }
    }

    #[test]
    fn test_update_entity_bar() {
        let mut x = create_default();
        let input = testdata::test_input();

        for i in 0..200 {
            x.update(input[i % input.len()]);
        }

        let bar = Bar::new(1000, 0.0, 100.0, 100.0, 0.0, 0.0);
        let out = x.update_bar(&bar);
        assert_eq!(out.len(), 8);
        for j in 0..8 {
            let sc = out[j].downcast_ref::<Scalar>().unwrap();
            assert_eq!(sc.time, 1000);
        }
    }

    #[test]
    fn test_validation_errors() {
        // alpha <= 0
        let p = TrendCycleModeParams { alpha_ema_period_additional: 0.0, ..Default::default() };
        assert!(TrendCycleMode::new(&p).is_err());

        // alpha > 1
        let p = TrendCycleModeParams { alpha_ema_period_additional: 1.0001, ..Default::default() };
        assert!(TrendCycleMode::new(&p).is_err());

        // tlsl < 2
        let p = TrendCycleModeParams { trend_line_smoothing_length: 1, ..Default::default() };
        assert!(TrendCycleMode::new(&p).is_err());

        // tlsl > 4
        let p = TrendCycleModeParams { trend_line_smoothing_length: 5, ..Default::default() };
        assert!(TrendCycleMode::new(&p).is_err());

        // cpm <= 0
        let p = TrendCycleModeParams { cycle_part_multiplier: 0.0, ..Default::default() };
        assert!(TrendCycleMode::new(&p).is_err());

        // cpm > 10
        let p = TrendCycleModeParams { cycle_part_multiplier: 10.0001, ..Default::default() };
        assert!(TrendCycleMode::new(&p).is_err());

        // sep <= 0
        let p = TrendCycleModeParams { separation_percentage: 0.0, ..Default::default() };
        assert!(TrendCycleMode::new(&p).is_err());

        // sep > 100
        let p = TrendCycleModeParams { separation_percentage: 100.0001, ..Default::default() };
        assert!(TrendCycleMode::new(&p).is_err());
    }
}
