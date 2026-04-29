use std::f64::consts::{FRAC_PI_2, FRAC_PI_3, PI};

// Constants
pub const DEFAULT_MIN_PERIOD: usize = 6;
pub const DEFAULT_MAX_PERIOD: usize = 50;
pub(crate) const HT_LENGTH: usize = 7;
pub(crate) const QUADRATURE_INDEX: usize = HT_LENGTH / 2;
pub(crate) const ACCUMULATION_LENGTH: usize = 40;

// CycleEstimatorType
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CycleEstimatorType {
    HomodyneDiscriminator = 1,
    HomodyneDiscriminatorUnrolled = 2,
    PhaseAccumulator = 3,
    DualDifferentiator = 4,
}

impl CycleEstimatorType {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::HomodyneDiscriminator => "homodyneDiscriminator",
            Self::HomodyneDiscriminatorUnrolled => "homodyneDiscriminatorUnrolled",
            Self::PhaseAccumulator => "phaseAccumulator",
            Self::DualDifferentiator => "dualDifferentiator",
        }
    }
}

impl std::fmt::Display for CycleEstimatorType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

// CycleEstimatorParams
pub struct CycleEstimatorParams {
    pub smoothing_length: usize,
    pub alpha_ema_quadrature_in_phase: f64,
    pub alpha_ema_period: f64,
    pub warm_up_period: usize,
}

impl Default for CycleEstimatorParams {
    fn default() -> Self {
        Self {
            smoothing_length: 4,
            alpha_ema_quadrature_in_phase: 0.2,
            alpha_ema_period: 0.2,
            warm_up_period: 0,
        }
    }
}

// CycleEstimator trait
pub trait CycleEstimator {
    fn smoothing_length(&self) -> usize;
    fn smoothed(&self) -> f64;
    fn detrended(&self) -> f64;
    fn quadrature(&self) -> f64;
    fn in_phase(&self) -> f64;
    fn period(&self) -> f64;
    fn count(&self) -> usize;
    fn primed(&self) -> bool;
    fn min_period(&self) -> usize;
    fn max_period(&self) -> usize;
    fn alpha_ema_quadrature_in_phase(&self) -> f64;
    fn alpha_ema_period(&self) -> f64;
    fn warm_up_period(&self) -> usize;
    fn update(&mut self, sample: f64);
}

// Factory
pub fn new_cycle_estimator(
    typ: CycleEstimatorType,
    params: &CycleEstimatorParams,
) -> Result<Box<dyn CycleEstimator>, String> {
    use super::homodyne_discriminator::HomodyneDiscriminatorEstimator;
    use super::homodyne_discriminator_unrolled::HomodyneDiscriminatorEstimatorUnrolled;
    use super::phase_accumulator::PhaseAccumulatorEstimator;
    use super::dual_differentiator::DualDifferentiatorEstimator;

    match typ {
        CycleEstimatorType::HomodyneDiscriminator => {
            let e = HomodyneDiscriminatorEstimator::new(params)?;
            Ok(Box::new(e))
        }
        CycleEstimatorType::HomodyneDiscriminatorUnrolled => {
            let e = HomodyneDiscriminatorEstimatorUnrolled::new(params)?;
            Ok(Box::new(e))
        }
        CycleEstimatorType::PhaseAccumulator => {
            let e = PhaseAccumulatorEstimator::new(params)?;
            Ok(Box::new(e))
        }
        CycleEstimatorType::DualDifferentiator => {
            let e = DualDifferentiatorEstimator::new(params)?;
            Ok(Box::new(e))
        }
    }
}

pub fn estimator_moniker(typ: CycleEstimatorType, estimator: &dyn CycleEstimator) -> String {
    let prefix = match typ {
        CycleEstimatorType::HomodyneDiscriminator => "hd",
        CycleEstimatorType::HomodyneDiscriminatorUnrolled => "hdu",
        CycleEstimatorType::PhaseAccumulator => "pa",
        CycleEstimatorType::DualDifferentiator => "dd",
    };
    format!(
        "{}({}, {:.3}, {:.3})",
        prefix,
        estimator.smoothing_length(),
        estimator.alpha_ema_quadrature_in_phase(),
        estimator.alpha_ema_period()
    )
}

// Shared helpers
pub(crate) fn verify_parameters(p: &CycleEstimatorParams) -> Result<(), String> {
    const INVALID: &str = "invalid cycle estimator parameters";

    if p.smoothing_length < 2 || p.smoothing_length > 4 {
        return Err(format!("{}: SmoothingLength should be in range [2, 4]", INVALID));
    }
    if p.alpha_ema_quadrature_in_phase <= 0.0 || p.alpha_ema_quadrature_in_phase >= 1.0 {
        return Err(format!("{}: AlphaEmaQuadratureInPhase should be in range (0, 1)", INVALID));
    }
    if p.alpha_ema_period <= 0.0 || p.alpha_ema_period >= 1.0 {
        return Err(format!("{}: AlphaEmaPeriod should be in range (0, 1)", INVALID));
    }
    Ok(())
}

pub(crate) fn push(array: &mut [f64], value: f64) {
    let len = array.len();
    for i in (1..len).rev() {
        array[i] = array[i - 1];
    }
    array[0] = value;
}

pub(crate) fn correct_amplitude(previous_period: f64) -> f64 {
    0.54 + 0.075 * previous_period
}

pub(crate) fn ht(array: &[f64]) -> f64 {
    const A: f64 = 0.0962;
    const B: f64 = 0.5769;
    A * array[0] + B * array[2] - B * array[4] - A * array[6]
}

pub(crate) fn adjust_period(mut period: f64, period_previous: f64) -> f64 {
    const MIN_FACTOR: f64 = 0.67;
    const MAX_FACTOR: f64 = 1.5;

    let temp = MAX_FACTOR * period_previous;
    if period > temp {
        period = temp;
    } else {
        let temp = MIN_FACTOR * period_previous;
        if period < temp {
            period = temp;
        }
    }

    if period < DEFAULT_MIN_PERIOD as f64 {
        period = DEFAULT_MIN_PERIOD as f64;
    } else if period > DEFAULT_MAX_PERIOD as f64 {
        period = DEFAULT_MAX_PERIOD as f64;
    }

    period
}

pub(crate) fn fill_wma_factors(length: usize, factors: &mut [f64]) {
    match length {
        4 => {
            factors[0] = 4.0 / 10.0;
            factors[1] = 3.0 / 10.0;
            factors[2] = 2.0 / 10.0;
            factors[3] = 1.0 / 10.0;
        }
        3 => {
            factors[0] = 3.0 / 6.0;
            factors[1] = 2.0 / 6.0;
            factors[2] = 1.0 / 6.0;
        }
        _ => {
            // length == 2
            factors[0] = 2.0 / 3.0;
            factors[1] = 1.0 / 3.0;
        }
    }
}

pub(crate) fn wma(raw_values: &[f64], wma_factors: &[f64], length: usize) -> f64 {
    let mut value = 0.0;
    for i in 0..length {
        value += wma_factors[i] * raw_values[i];
    }
    value
}

pub(crate) fn ema(alpha: f64, one_minus_alpha: f64, value: f64, value_previous: f64) -> f64 {
    alpha * value + one_minus_alpha * value_previous
}

// Phase accumulator helpers
pub(crate) fn instantaneous_phase(smoothed_in_phase: f64, smoothed_quadrature: f64, phase_previous: f64) -> f64 {
    let phase = (smoothed_quadrature / smoothed_in_phase).abs().atan();
    if phase.is_nan() || phase.is_infinite() {
        return phase_previous;
    }

    if smoothed_in_phase < 0.0 {
        if smoothed_quadrature > 0.0 {
            PI - phase
        } else if smoothed_quadrature < 0.0 {
            PI + phase
        } else {
            phase
        }
    } else if smoothed_in_phase > 0.0 && smoothed_quadrature < 0.0 {
        2.0 * PI - phase
    } else {
        phase
    }
}

pub(crate) fn calculate_differential_phase(phase: f64, phase_previous: f64) -> f64 {
    const TWO_PI: f64 = 2.0 * PI;
    const PI_OVER_2: f64 = FRAC_PI_2;
    const THREE_PI_OVER_4: f64 = 3.0 * PI / 4.0;
    // Use correctly-rounded compile-time constants to match Go's arbitrary-precision
    // constant folding. Runtime TWO_PI / 6.0 gives a double-rounding error.
    const MIN_DELTA_PHASE: f64 = TWO_PI / DEFAULT_MAX_PERIOD as f64;
    const MAX_DELTA_PHASE: f64 = FRAC_PI_3; // 2π/6 = π/3

    let mut delta_phase = phase_previous - phase;

    if phase_previous < PI_OVER_2 && phase > THREE_PI_OVER_4 {
        delta_phase += TWO_PI;
    }

    if delta_phase < MIN_DELTA_PHASE {
        delta_phase = MIN_DELTA_PHASE;
    } else if delta_phase > MAX_DELTA_PHASE {
        delta_phase = MAX_DELTA_PHASE;
    }

    delta_phase
}

pub(crate) fn instantaneous_period(delta_phase: &[f64], period_previous: f64) -> f64 {
    const TWO_PI: f64 = 2.0 * PI;

    let mut sum_phase = 0.0;
    let mut period = 0usize;

    for i in 0..ACCUMULATION_LENGTH {
        sum_phase += delta_phase[i];
        if sum_phase >= TWO_PI {
            period = i + 1;
            break;
        }
    }

    if period == 0 {
        return period_previous;
    }

    period as f64
}
