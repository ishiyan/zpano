use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
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
    CycleEstimator, CycleEstimatorParams, CycleEstimatorType, estimator_moniker,
    new_cycle_estimator,
};

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Dominant Cycle indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DominantCycleOutput {
    /// The raw instantaneous cycle period.
    RawPeriod = 1,
    /// The dominant cycle period (EMA-smoothed).
    Period = 2,
    /// The dominant cycle phase, in degrees.
    Phase = 3,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Dominant Cycle indicator.
pub struct DominantCycleParams {
    /// The type of cycle estimator to use.
    pub estimator_type: CycleEstimatorType,
    /// Parameters for the cycle estimator.
    pub estimator_params: CycleEstimatorParams,
    /// α for additional EMA smoothing of the period. Must be in (0, 1].
    pub alpha_ema_period_additional: f64,
    /// Bar component. `None` means default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component. `None` means default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component. `None` means default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for DominantCycleParams {
    fn default() -> Self {
        Self {
            estimator_type: CycleEstimatorType::HomodyneDiscriminator,
            estimator_params: CycleEstimatorParams::default(),
            alpha_ema_period_additional: 0.33,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Dominant Cycle indicator.
///
/// Computes the instantaneous cycle period and phase derived from a Hilbert
/// transformer cycle estimator.
///
/// Three outputs: RawPeriod, Period (EMA-smoothed), Phase (degrees).
pub struct DominantCycle {
    mnemonic_raw_period: String,
    description_raw_period: String,
    mnemonic_period: String,
    description_period: String,
    mnemonic_phase: String,
    description_phase: String,
    alpha_ema_period_additional: f64,
    one_min_alpha_ema_period_additional: f64,
    smoothed_period: f64,
    smoothed_phase: f64,
    smoothed_input: Vec<f64>,
    smoothed_input_length_min1: usize,
    htce: Box<dyn CycleEstimator>,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl DominantCycle {
    /// Creates a new Dominant Cycle with default parameters.
    pub fn new_default() -> Result<Self, String> {
        let params = CycleEstimatorParams {
            smoothing_length: 4,
            alpha_ema_quadrature_in_phase: 0.2,
            alpha_ema_period: 0.2,
            warm_up_period: 100,
        };
        Self::new_inner(
            CycleEstimatorType::HomodyneDiscriminator,
            &params,
            0.33,
            None,
            None,
            None,
        )
    }

    /// Creates a new Dominant Cycle from supplied parameters.
    pub fn new(p: &DominantCycleParams) -> Result<Self, String> {
        Self::new_inner(
            p.estimator_type,
            &p.estimator_params,
            p.alpha_ema_period_additional,
            p.bar_component,
            p.quote_component,
            p.trade_component,
        )
    }

    fn new_inner(
        estimator_type: CycleEstimatorType,
        estimator_params: &CycleEstimatorParams,
        alpha_ema_period_additional: f64,
        bc: Option<BarComponent>,
        qc: Option<QuoteComponent>,
        tc: Option<TradeComponent>,
    ) -> Result<Self, String> {
        const INVALID: &str = "invalid dominant cycle parameters";

        if alpha_ema_period_additional <= 0.0 || alpha_ema_period_additional > 1.0 {
            return Err(format!(
                "{}: α for additional smoothing should be in range (0, 1]",
                INVALID
            ));
        }

        let estimator = new_cycle_estimator(estimator_type, estimator_params)?;

        let mut est_moniker = String::new();
        if estimator_type != CycleEstimatorType::HomodyneDiscriminator
            || estimator_params.smoothing_length != 4
            || estimator_params.alpha_ema_quadrature_in_phase != 0.2
            || estimator_params.alpha_ema_period != 0.2
        {
            let m = estimator_moniker(estimator_type, estimator.as_ref());
            if !m.is_empty() {
                est_moniker = format!(", {}", m);
            }
        }

        let bc = bc.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = qc.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = tc.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let comp_mn = component_triple_mnemonic(bc, qc, tc);

        let mnemonic_raw_period = format!("dcp-raw({:.3}{}{})", alpha_ema_period_additional, est_moniker, comp_mn);
        let mnemonic_period = format!("dcp({:.3}{}{})", alpha_ema_period_additional, est_moniker, comp_mn);
        let mnemonic_phase = format!("dcph({:.3}{}{})", alpha_ema_period_additional, est_moniker, comp_mn);

        let description_raw_period = format!("Dominant cycle raw period {}", mnemonic_raw_period);
        let description_period = format!("Dominant cycle period {}", mnemonic_period);
        let description_phase = format!("Dominant cycle phase {}", mnemonic_phase);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let max_period = estimator.max_period();

        Ok(Self {
            mnemonic_raw_period,
            description_raw_period,
            mnemonic_period,
            description_period,
            mnemonic_phase,
            description_phase,
            alpha_ema_period_additional,
            one_min_alpha_ema_period_additional: 1.0 - alpha_ema_period_additional,
            smoothed_period: 0.0,
            smoothed_phase: 0.0,
            smoothed_input: vec![0.0; max_period],
            smoothed_input_length_min1: max_period - 1,
            htce: estimator,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Returns the current WMA-smoothed price value from the underlying HTCE.
    /// Returns NaN if not yet primed.
    pub fn smoothed_price(&self) -> f64 {
        if !self.primed {
            return f64::NAN;
        }
        self.htce.smoothed()
    }

    /// Returns the maximum cycle period supported by the underlying HTCE.
    pub fn max_period(&self) -> usize {
        self.htce.max_period()
    }

    /// Core update. Returns (raw_period, period, phase). NaN triple if not primed.
    pub fn update(&mut self, sample: f64) -> (f64, f64, f64) {
        if sample.is_nan() {
            return (f64::NAN, f64::NAN, f64::NAN);
        }

        self.htce.update(sample);
        self.push_smoothed_input(self.htce.smoothed());

        if self.primed {
            self.smoothed_period = self.alpha_ema_period_additional * self.htce.period()
                + self.one_min_alpha_ema_period_additional * self.smoothed_period;
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
        let n = self.smoothed_input_length_min1;
        for i in (1..=n).rev() {
            self.smoothed_input[i] = self.smoothed_input[i - 1];
        }
        self.smoothed_input[0] = value;
    }

    fn calculate_smoothed_phase(&mut self) {
        const RAD2DEG: f64 = 180.0 / std::f64::consts::PI;
        const TWO_PI: f64 = 2.0 * std::f64::consts::PI;
        const EPSILON: f64 = 0.01;

        let length = {
            let l = (self.smoothed_period + 0.5).floor() as usize;
            if l > self.smoothed_input_length_min1 {
                self.smoothed_input_length_min1
            } else {
                l
            }
        };

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

        if imag_part.abs() <= EPSILON {
            if real_part > 0.0 {
                phase += 90.0;
            } else if real_part < 0.0 {
                phase -= 90.0;
            }
        }

        // 90 degree reference shift.
        phase += 90.0;
        // Compensate for one bar lag.
        phase += 360.0 / self.smoothed_period;
        // Resolve phase ambiguity.
        if imag_part < 0.0 {
            phase += 180.0;
        }
        // Cycle wraparound.
        if phase > 360.0 {
            phase -= 360.0;
        }

        self.smoothed_phase = phase;
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let (raw_period, period, phase) = self.update(sample);
        vec![
            Box::new(Scalar::new(time, raw_period)),
            Box::new(Scalar::new(time, period)),
            Box::new(Scalar::new(time, phase)),
        ]
    }
}

impl Indicator for DominantCycle {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::DominantCycle,
            &self.mnemonic_period,
            &self.description_period,
            &[
                OutputText {
                    mnemonic: self.mnemonic_raw_period.clone(),
                    description: self.description_raw_period.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_period.clone(),
                    description: self.description_period.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_phase.clone(),
                    description: self.description_phase.clone(),
                },
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

    fn phase_diff(a: f64, b: f64) -> f64 {
        let mut d = (a - b) % 360.0;
        if d > 180.0 {
            d -= 360.0;
        } else if d <= -180.0 {
            d += 360.0;
        }
        d
    }

    fn create_default() -> DominantCycle {
        DominantCycle::new_default().unwrap()
    }

    fn create_alpha(alpha: f64, estimator_type: CycleEstimatorType) -> DominantCycle {
        let params = DominantCycleParams {
            estimator_type,
            estimator_params: CycleEstimatorParams {
                smoothing_length: 4,
                alpha_ema_quadrature_in_phase: 0.2,
                alpha_ema_period: 0.2,
                warm_up_period: 0,
            },
            alpha_ema_period_additional: alpha,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        };
        DominantCycle::new(&params).unwrap()
    }

    fn create_cycle_estimator_params() -> CycleEstimatorParams {
        CycleEstimatorParams {
            smoothing_length: 4,
            alpha_ema_quadrature_in_phase: 0.2,
            alpha_ema_period: 0.2,
            warm_up_period: 0,
        }
    }
    #[test]
    fn test_reference_period() {
        let input = testdata::test_input();
        let exp_period = testdata::test_expected_period();
        let mut dc = create_default();

        for i in SKIP..input.len() {
            let (_, period, _) = dc.update(input[i]);
            if period.is_nan() || i < SETTLE_SKIP {
                continue;
            }
            assert!(
                (exp_period[i] - period).abs() <= TOLERANCE,
                "[{}] period: expected {}, actual {}",
                i, exp_period[i], period
            );
        }
    }

    #[test]
    fn test_reference_phase() {
        let input = testdata::test_input();
        let exp_phase = testdata::test_expected_phase();
        let mut dc = create_default();

        for i in SKIP..input.len() {
            let (_, _, phase) = dc.update(input[i]);
            if phase.is_nan() || i < SETTLE_SKIP {
                continue;
            }
            if exp_phase[i].is_nan() {
                continue;
            }
            assert!(
                phase_diff(exp_phase[i], phase).abs() <= TOLERANCE,
                "[{}] phase: expected {}, actual {}",
                i, exp_phase[i], phase
            );
        }
    }

    #[test]
    fn test_nan_input() {
        let mut dc = create_default();
        let (rp, p, ph) = dc.update(f64::NAN);
        assert!(rp.is_nan());
        assert!(p.is_nan());
        assert!(ph.is_nan());
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut dc = create_default();

        assert!(!dc.is_primed());

        let mut primed_at: Option<usize> = None;
        for i in 0..input.len() {
            dc.update(input[i]);
            if dc.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert!(primed_at.is_some(), "expected indicator to become primed");
        assert!(dc.is_primed());
    }

    #[test]
    fn test_update_entity() {
        let input = testdata::test_input();
        let time = 1617235200_i64;
        let inp = 100.0;

        // Scalar
        {
            let mut dc = create_default();
            for v in &input[..30] {
                dc.update(*v);
            }
            let s = Scalar::new(time, inp);
            let out = dc.update_scalar(&s);
            assert_eq!(out.len(), 3);
            for o in &out {
                let s = o.downcast_ref::<Scalar>().unwrap();
                assert_eq!(s.time, time);
            }
        }

        // Bar
        {
            let mut dc = create_default();
            for v in &input[..30] {
                dc.update(*v);
            }
            let b = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
            let out = dc.update_bar(&b);
            assert_eq!(out.len(), 3);
        }

        // Quote
        {
            let mut dc = create_default();
            for v in &input[..30] {
                dc.update(*v);
            }
            let q = Quote::new(time, inp, inp, 1.0, 1.0);
            let out = dc.update_quote(&q);
            assert_eq!(out.len(), 3);
        }

        // Trade
        {
            let mut dc = create_default();
            for v in &input[..30] {
                dc.update(*v);
            }
            let t = Trade::new(time, inp, 0.0);
            let out = dc.update_trade(&t);
            assert_eq!(out.len(), 3);
        }
    }

    #[test]
    fn test_metadata_default() {
        let dc = create_default();
        let m = dc.metadata();

        assert_eq!(m.identifier, Identifier::DominantCycle);
        assert_eq!(m.mnemonic, "dcp(0.330)");
        assert_eq!(m.description, "Dominant cycle period dcp(0.330)");
        assert_eq!(m.outputs.len(), 3);

        assert_eq!(m.outputs[0].kind, DominantCycleOutput::RawPeriod as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "dcp-raw(0.330)");
        assert_eq!(m.outputs[0].description, "Dominant cycle raw period dcp-raw(0.330)");

        assert_eq!(m.outputs[1].kind, DominantCycleOutput::Period as i32);
        assert_eq!(m.outputs[1].shape, Shape::Scalar);
        assert_eq!(m.outputs[1].mnemonic, "dcp(0.330)");

        assert_eq!(m.outputs[2].kind, DominantCycleOutput::Phase as i32);
        assert_eq!(m.outputs[2].shape, Shape::Scalar);
        assert_eq!(m.outputs[2].mnemonic, "dcph(0.330)");
    }

    #[test]
    fn test_metadata_phase_accumulator() {
        let dc = create_alpha(0.5, CycleEstimatorType::PhaseAccumulator);
        let m = dc.metadata();
        assert_eq!(m.mnemonic, "dcp(0.500, pa(4, 0.200, 0.200))");
    }

    #[test]
    fn test_smoothed_price() {
        let input = testdata::test_input();
        let mut dc = create_default();

        assert!(dc.smoothed_price().is_nan());

        for i in 0..input.len() {
            dc.update(input[i]);
            if dc.is_primed() {
                assert!(!dc.smoothed_price().is_nan());
                break;
            } else {
                assert!(dc.smoothed_price().is_nan());
            }
        }
    }

    #[test]
    fn test_max_period() {
        let dc = create_default();
        assert_eq!(dc.max_period(), dc.smoothed_input.len());
    }

    #[test]
    fn test_new_validation() {
        let err_alpha = "invalid dominant cycle parameters: α for additional smoothing should be in range (0, 1]";

        // alpha = 0
        let p = DominantCycleParams {
            alpha_ema_period_additional: 0.0,
            estimator_params: create_cycle_estimator_params(),
            ..Default::default()
        };
        assert_eq!(DominantCycle::new(&p).err().unwrap(), err_alpha);

        // alpha > 1
        let p = DominantCycleParams {
            alpha_ema_period_additional: 1.00000001,
            estimator_params: create_cycle_estimator_params(),
            ..Default::default()
        };
        assert_eq!(DominantCycle::new(&p).err().unwrap(), err_alpha);

        // alpha = 1.0 should succeed
        let p = DominantCycleParams {
            alpha_ema_period_additional: 1.0,
            estimator_params: create_cycle_estimator_params(),
            ..Default::default()
        };
        assert!(DominantCycle::new(&p).is_ok());
    }

    #[test]
    fn test_new_estimator_types() {
        // HomodyneDiscriminatorUnrolled
        let p = DominantCycleParams {
            alpha_ema_period_additional: 0.5,
            estimator_type: CycleEstimatorType::HomodyneDiscriminatorUnrolled,
            estimator_params: create_cycle_estimator_params(),
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        };
        let dc = DominantCycle::new(&p).unwrap();
        assert_eq!(dc.mnemonic_period, "dcp(0.500, hdu(4, 0.200, 0.200), hl/2)");

        // DualDifferentiator
        let p = DominantCycleParams {
            alpha_ema_period_additional: 0.5,
            estimator_type: CycleEstimatorType::DualDifferentiator,
            estimator_params: create_cycle_estimator_params(),
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        };
        let dc = DominantCycle::new(&p).unwrap();
        assert_eq!(dc.mnemonic_period, "dcp(0.500, dd(4, 0.200, 0.200), hl/2)");
    }
}
