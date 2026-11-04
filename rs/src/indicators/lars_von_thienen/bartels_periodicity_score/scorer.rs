use super::super::harmonic_cycle::model as harmoniccycle;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct BartelsOutput {
    pub score: f64,
    pub genuine_percent: f64,
    pub segments: f64,
}
impl BartelsOutput {}
pub fn bartels_score(
    values: Vec<f64>,
    period: i64,
    max_segments: i64,
    min_segments: i64,
) -> BartelsOutput {
    {
        let err = harmoniccycle::validate_integer(period, "period", 3);
        if err.is_some() {
            panic!("{:?}", err);
        }
    }
    if ((max_segments < 2) || (min_segments < 2)) || (min_segments > max_segments) {
        panic!("{:?}", "invalid Bartels segments");
    }
    for value in values.clone() {
        harmoniccycle::validate_sample(value);
        if (value).is_nan() {
            panic!("{:?}", "Bartels batch data must be finite");
        }
    }
    let segments: i64 = ((values.len() as i64) / period).min(max_segments);
    let nan: f64 = f64::NAN;
    if segments < min_segments {
        return BartelsOutput {
            score: nan,
            genuine_percent: nan,
            segments: (segments as f64),
        }
        .clone();
    }
    let mut sum_a: f64 = Default::default();
    let mut sum_b: f64 = Default::default();
    let mut sum_energy: f64 = Default::default();

    for segment in 0..segments {
        let mut a: f64 = Default::default();
        let mut b: f64 = Default::default();

        for j in 0..period {
            let x: i64 = (segment * period) + j;
            let angle: f64 =
                (((360.0 / (period as f64)) * ((x + 1) as f64)) / 180.0) * std::f64::consts::PI;
            a += (angle).sin() * values[(x) as usize];
            b += (angle).cos() * values[(x) as usize];
        }
        sum_a += a;
        sum_b += b;
        sum_energy = harmoniccycle::finite((sum_energy + (a * a)) + (b * b));
    }
    if sum_energy == 0.0 {
        return BartelsOutput {
            score: nan,
            genuine_percent: nan,
            segments: (segments as f64),
        }
        .clone();
    }
    let (average_a, average_b) = ((sum_a / (segments as f64)), (sum_b / (segments as f64)));
    let amplitude: f64 = ((average_a * average_a) + (average_b * average_b)).sqrt();
    let rms_amplitude: f64 = (sum_energy / (segments as f64)).sqrt();
    let reference: f64 = rms_amplitude / (segments as f64).sqrt();
    let ratio: f64 = amplitude / reference;
    let score: f64 = (-ratio * ratio).exp();
    return BartelsOutput {
        score: score,
        genuine_percent: (100.0 * (1.0 - score)),
        segments: (segments as f64),
    }
    .clone();
}
