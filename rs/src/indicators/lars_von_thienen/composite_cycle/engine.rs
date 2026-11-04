use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::harmonic_cycle::model::CycleModel;
use super::super::hodrick_prescott_smoothing_and_detrending::solver as hodrickprescottsmoothinganddetrending;
use super::CompositeCycleParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub value: f64,
    pub slope: f64,
    pub next_value: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        value: nan,
        slope: nan,
        next_value: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: CompositeCycleParams,
    pub prices: Vec<f64>,
    pub models: Vec<harmoniccycle::CycleModel>,
}
impl Engine {
    pub fn reset(&mut self) {
        self.prices = Vec::new();
        self.models = Vec::new();
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
            let mut models: Vec<CycleModel> =
                vec![Default::default(); (self.p.periods.len() as i64) as usize].clone();
            let mut sum: f64 = Default::default();

            for (__index, period) in self.p.periods.clone().into_iter().enumerate() {
                let i = __index as i64;
                models[(i) as usize] = harmoniccycle::goertzel(values.clone(), period).clone();
                sum += harmoniccycle::cycle_slope(
                    models[(i) as usize].clone(),
                    ((self.p.window as i64) - 1) as f64,
                );
            }
            output = Snapshot {
                value: harmoniccycle::composite_value(
                    models.clone(),
                    ((self.p.window as i64) - 1) as f64,
                ),
                slope: harmoniccycle::finite(sum),
                next_value: harmoniccycle::composite_value(
                    models.clone(),
                    (self.p.window as i64) as f64,
                ),
            }
            .clone();
            self.models = models.clone();
        }
        self.prices = prices.clone();
        return output.clone();
    }
}
