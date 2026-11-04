// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct CycleModel {
    pub period: f64,
    pub amplitude: f64,
    pub phase: f64,
    pub trough_offset: f64,
}
impl CycleModel {}
pub fn goertzel(values: Vec<f64>, period: f64) -> CycleModel {
    {
        let err = validate_real(period, "period", 2.0000000000000004);
        if err.is_some() {
            panic!("{:?}", err);
        }
    }
    let n: i64 = values.len() as i64;
    if n < 1 {
        panic!("{:?}", "Goertzel needs observations");
    }
    let omega: f64 = (2.0 * std::f64::consts::PI) / period;
    let (sine, cosine) = ((omega).sin(), (omega).cos());
    let coefficient: f64 = 2.0 * cosine;
    let mut q1: f64 = Default::default();
    let mut q2: f64 = Default::default();

    for value in values.clone() {
        validate_sample(value);
        if (value).is_nan() {
            panic!("{:?}", "Goertzel batch input must be finite");
        }
        let q0: f64 = finite(((coefficient * q1) - q2) + value);
        (q2, q1) = (q1, q0);
    }
    let (r1, i1) = ((q1 - (q2 * cosine)), (q2 * sine));
    let (cn, sn) = (
        (((n - 1) as f64) * omega).cos(),
        (((n - 1) as f64) * omega).sin(),
    );
    let real: f64 = finite((r1 * cn) + (i1 * sn));
    let imaginary: f64 = finite((i1 * cn) - (r1 * sn));
    let amplitude: f64 =
        finite((2.0 * (finite((real * real) + (imaginary * imaginary))).sqrt()) / (n as f64));
    if amplitude == 0.0 {
        return CycleModel {
            period: period,
            amplitude: 0.0,
            phase: f64::NAN,
            trough_offset: f64::NAN,
        }
        .clone();
    }
    let phase: f64 = (std::f64::consts::PI / 2.0) + (imaginary).atan2(real);
    let offset: f64 = period * (0.75 - (phase / (2.0 * std::f64::consts::PI)));
    return CycleModel {
        period: period,
        amplitude: amplitude,
        phase: phase,
        trough_offset: offset,
    }
    .clone();
}
pub fn cycle_value(model: CycleModel, offset: f64) -> f64 {
    if model.amplitude == 0.0 {
        return 0.0;
    }
    return model.amplitude
        * ((((2.0 * std::f64::consts::PI) * offset) / model.period) + model.phase).sin();
}
pub fn cycle_slope(model: CycleModel, offset: f64) -> f64 {
    if model.amplitude == 0.0 {
        return 0.0;
    }
    let omega: f64 = (2.0 * std::f64::consts::PI) / model.period;
    return (model.amplitude * omega) * ((omega * offset) + model.phase).cos();
}
pub fn composite_value(models: Vec<CycleModel>, offset: f64) -> f64 {
    let mut sum: f64 = Default::default();

    for model in models.clone() {
        sum += cycle_value(model.clone(), offset);
    }
    return finite(sum);
}
pub fn cycle_strength(amplitude: f64, period: f64) -> f64 {
    {
        let err = validate_real(amplitude, "amplitude", 0.0);
        if err.is_some() {
            panic!("{:?}", err);
        }
    }
    {
        let err = validate_real(period, "period", 2.0000000000000004);
        if err.is_some() {
            panic!("{:?}", err);
        }
    }
    return amplitude / period;
}
pub fn cycle_polarity(short_slope: f64, long_slope: f64, tolerance: f64) -> f64 {
    validate_sample(short_slope);
    validate_sample(long_slope);
    {
        let err = validate_real(tolerance, "tolerance", 0.0);
        if err.is_some() {
            panic!("{:?}", err);
        }
    }
    if (short_slope).is_nan() || (long_slope).is_nan() {
        return f64::NAN;
    }
    if (short_slope > tolerance) && (long_slope > tolerance) {
        return 1.0;
    }
    if (short_slope < -tolerance) && (long_slope < -tolerance) {
        return -1 as f64;
    }
    return 0.0;
}
pub fn phase_state(phase: f64) -> f64 {
    validate_sample(phase);
    return 100.0 * (phase).sin();
}

pub use super::validation::{finite, push, validate_integer, validate_real, validate_sample};

pub use super::summation::sum;
