use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::hodrick_prescott_smoothing_and_detrending::solver as hodrickprescottsmoothinganddetrending;
use super::GoertzelCycleSpectrumParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub amplitude: Vec<f64>,
    pub phases: Vec<f64>,
    pub trough_offsets: Vec<f64>,
    pub dominant_period: f64,
    pub dominant_amplitude: f64,
    pub dominant_phase: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        amplitude: Vec::new(),
        phases: Vec::new(),
        trough_offsets: Vec::new(),
        dominant_period: nan,
        dominant_amplitude: nan,
        dominant_phase: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: GoertzelCycleSpectrumParams,
    pub prices: Vec<f64>,
}
impl Engine {
    pub fn reset(&mut self) {
        self.prices = Vec::new();
    }
    pub fn primed(&self) -> bool {
        return (self.prices.len() as i64) == (self.p.window as i64);
    }
    pub fn warmup(&self) -> i64 {
        return (self.p.window as i64) - 1;
    }
    pub fn update(&mut self, price: f64) -> Snapshot {
        harmoniccycle::validate_sample(price);
        if (price).is_nan() {
            return empty_snapshot().clone();
        }
        let prices: Vec<f64> = harmoniccycle::push(
            {
                let mut __values = vec![].clone();
                __values.extend(self.prices.clone());
                __values
            }
            .clone(),
            price,
            self.p.window as i64,
        )
        .clone();
        let mut output = empty_snapshot().clone();
        if (prices.len() as i64) == (self.p.window as i64) {
            let values: Vec<f64> =
                hodrickprescottsmoothinganddetrending::cycle_data(prices.clone(), self.p.smoothing)
                    .clone();
            let count: i64 = (((self.p.max_period as i64) - (self.p.min_period as i64))
                / (self.p.step as i64))
                + 1;
            output.amplitude = vec![Default::default(); (count) as usize].clone();
            output.phases = vec![Default::default(); (count) as usize].clone();
            output.trough_offsets = vec![Default::default(); (count) as usize].clone();
            let mut best: i64 = 0;
            for i in 0..count {
                let model = harmoniccycle::goertzel(
                    values.clone(),
                    ((self.p.min_period as i64) + (i * (self.p.step as i64))) as f64,
                )
                .clone();
                output.amplitude[(i) as usize] = model.amplitude;
                output.phases[(i) as usize] = model.phase;
                output.trough_offsets[(i) as usize] = model.trough_offset;
                if model.amplitude > output.amplitude[(best) as usize] {
                    best = i;
                }
            }
            output.dominant_amplitude = output.amplitude[(best) as usize];
            output.dominant_phase = output.phases[(best) as usize];
            if output.dominant_amplitude > 0.0 {
                output.dominant_period =
                    ((self.p.min_period as i64) + (best * (self.p.step as i64))) as f64;
            }
        }
        self.prices = prices.clone();
        return output.clone();
    }
}
