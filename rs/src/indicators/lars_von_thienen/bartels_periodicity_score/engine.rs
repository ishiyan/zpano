use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::hodrick_prescott_smoothing_and_detrending::solver as hodrickprescottsmoothinganddetrending;
use super::BartelsPeriodicityScoreParams;
use super::scorer::bartels_score;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub score: f64,
    pub genuine_percent: f64,
    pub segments: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        score: nan,
        genuine_percent: nan,
        segments: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: BartelsPeriodicityScoreParams,
    pub prices: Vec<f64>,
}
impl Engine {
    pub fn reset(&mut self) {
        self.prices = Vec::new();
    }
    pub fn primed(&self) -> bool {
        return (self.prices.len() as i64) == ((self.p.period as i64) * (self.p.segments as i64));
    }
    pub fn warmup(&self) -> i64 {
        return ((self.p.period as i64) * (self.p.segments as i64)) - 1;
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
            (self.p.period as i64) * (self.p.segments as i64),
        )
        .clone();
        let mut output = empty_snapshot().clone();
        if (prices.len() as i64) == ((self.p.period as i64) * (self.p.segments as i64)) {
            let score = bartels_score(
                hodrickprescottsmoothinganddetrending::cycle_data(prices.clone(), self.p.smoothing)
                    .clone(),
                self.p.period as i64,
                self.p.segments as i64,
                self.p.segments as i64,
            )
            .clone();
            output = Snapshot {
                score: score.score,
                genuine_percent: score.genuine_percent,
                segments: score.segments,
            }
            .clone();
        }
        self.prices = prices.clone();
        return output.clone();
    }
}
