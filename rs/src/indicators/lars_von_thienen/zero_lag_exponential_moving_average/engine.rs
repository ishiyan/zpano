use super::super::harmonic_cycle::model as harmoniccycle;
use super::LarsVonThienenZeroLagExponentialMovingAverageParams;
use super::smoother::{ZLEMA, new_zlema};
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub value: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot { value: nan }.clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: LarsVonThienenZeroLagExponentialMovingAverageParams,
    pub smoother: ZLEMA,
    pub samples_seen: i64,
}
impl Engine {
    pub fn reset(&mut self) {
        self.smoother = new_zlema(self.p.length as i64).clone();
        self.samples_seen = 0;
    }
    pub fn primed(&self) -> bool {
        return self.samples_seen >= ((((self.p.length as i64) - 1) / 2) + 1);
    }
    pub fn warmup(&self) -> i64 {
        return ((self.p.length as i64) - 1) / 2;
    }
    pub fn update(&mut self, price: f64) -> Snapshot {
        harmoniccycle::validate_sample(price);
        if (price).is_nan() {
            return empty_snapshot().clone();
        }
        self.samples_seen += 1;
        return Snapshot {
            value: self.smoother.update(price),
        }
        .clone();
    }
}
