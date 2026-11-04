use super::super::harmonic_cycle::model as harmoniccycle;
use super::HodrickPrescottSmoothingAndDetrendingParams;
use super::solver::hp_trend;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub trend: f64,
    pub cycle: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        trend: nan,
        cycle: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: HodrickPrescottSmoothingAndDetrendingParams,
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
            let trend: Vec<f64> = hp_trend(prices.clone(), self.p.smoothing).clone();
            output = Snapshot {
                trend: trend[((trend.len() as i64) - 1) as usize],
                cycle: harmoniccycle::finite(price - trend[((trend.len() as i64) - 1) as usize]),
            }
            .clone();
        }
        self.prices = prices.clone();
        return output.clone();
    }
}
