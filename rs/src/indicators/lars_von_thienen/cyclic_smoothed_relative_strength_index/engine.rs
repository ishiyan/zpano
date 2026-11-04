use super::super::adaptive_bands::bands as adaptivebands;
use super::super::harmonic_cycle::model as harmoniccycle;
use super::CyclicSmoothedRelativeStrengthIndexParams;
use super::rsi::{WilderRSI, new_wilder_rsi};
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub value: f64,
    pub lower: f64,
    pub upper: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        value: nan,
        lower: nan,
        upper: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: CyclicSmoothedRelativeStrengthIndexParams,
    pub rsi: WilderRSI,
    pub rsi_history: Vec<f64>,
    pub values: Vec<f64>,
    pub previous_value: f64,
}
impl Engine {
    pub fn reset(&mut self) {
        self.rsi = new_wilder_rsi((self.p.dominant_cycle as i64) / 2).clone();
        self.rsi_history = Vec::new();
        self.values = Vec::new();
        self.previous_value = 0.0;
    }
    pub fn primed(&self) -> bool {
        return (self.values.len() as i64) == (2 * (self.p.dominant_cycle as i64));
    }
    pub fn warmup(&self) -> i64 {
        return ((((self.p.dominant_cycle as i64) / 2) + (((self.p.vibration as i64) - 1) / 2))
            + (2 * (self.p.dominant_cycle as i64)))
            - 1;
    }
    pub fn update(&mut self, price: f64) -> Snapshot {
        harmoniccycle::validate_sample(price);
        if (price).is_nan() {
            return empty_snapshot().clone();
        }
        let rsi: f64 = self.rsi.update(price);
        if (rsi).is_nan() {
            return empty_snapshot().clone();
        }
        let lag: i64 = ((self.p.vibration as i64) - 1) / 2;
        self.rsi_history = harmoniccycle::push(self.rsi_history.clone(), rsi, lag + 1).clone();
        if (self.rsi_history.len() as i64) < (lag + 1) {
            return empty_snapshot().clone();
        }
        let torque: f64 = 2.0 / (((self.p.vibration as i64) + 1) as f64);
        let value: f64 = harmoniccycle::finite(
            (torque * ((2.0 * rsi) - self.rsi_history[(0) as usize]))
                + ((1.0 - torque) * self.previous_value),
        );
        self.previous_value = value;
        self.values = harmoniccycle::push(
            self.values.clone(),
            value,
            2 * (self.p.dominant_cycle as i64),
        )
        .clone();
        let mut output = empty_snapshot().clone();
        output.value = value;
        if self.primed() {
            (output.lower, output.upper) =
                adaptivebands::grid_bands(self.values.clone(), self.p.leveling);
        }
        return output.clone();
    }
}
