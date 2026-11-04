use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::hodrick_prescott_smoothing_and_detrending::solver as hodrickprescottsmoothinganddetrending;
use super::CycleStrengthParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub strength: f64,
    pub amplitude: f64,
    pub phase: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        strength: nan,
        amplitude: nan,
        phase: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: CycleStrengthParams,
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
            let model = harmoniccycle::goertzel(values.clone(), self.p.period).clone();
            output = Snapshot {
                strength: harmoniccycle::cycle_strength(model.amplitude, self.p.period),
                amplitude: model.amplitude,
                phase: model.phase,
            }
            .clone();
        }
        self.prices = prices.clone();
        return output.clone();
    }
}
