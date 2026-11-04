use super::super::adaptive_bands::bands as adaptivebands;
use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::hodrick_prescott_smoothing_and_detrending::solver as hodrickprescottsmoothinganddetrending;
use super::CycleSwingMomentumParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub value: f64,
    pub lower: f64,
    pub upper: f64,
    pub score: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        value: nan,
        lower: nan,
        upper: nan,
        score: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: CycleSwingMomentumParams,
    pub prices: Vec<f64>,
    pub values: Vec<f64>,
}
impl Engine {
    pub fn reset(&mut self) {
        self.prices = Vec::new();
        self.values = Vec::new();
    }
    pub fn primed(&self) -> bool {
        return (self.values.len() as i64) == (self.p.band_memory as i64);
    }
    pub fn warmup(&self) -> i64 {
        return ((self.p.window as i64) + (self.p.band_memory as i64)) - 2;
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
            let fast: Vec<f64> = hodrickprescottsmoothinganddetrending::hp_trend(
                prices.clone(),
                self.p.fast_smoothing,
            )
            .clone();
            let slow: Vec<f64> = hodrickprescottsmoothinganddetrending::hp_trend(
                prices.clone(),
                self.p.slow_smoothing,
            )
            .clone();
            let value: f64 = harmoniccycle::finite(
                fast[((fast.len() as i64) - 1) as usize] - slow[((slow.len() as i64) - 1) as usize],
            );
            self.values =
                harmoniccycle::push(self.values.clone(), value, self.p.band_memory as i64)
                    .clone();
            output.value = value;
            if self.primed() {
                (output.lower, output.upper) =
                    adaptivebands::grid_bands(self.values.clone(), self.p.leveling);
                output.score = 0.0;
                if value >= output.upper {
                    output.score = 1.0;
                } else if value <= output.lower {
                    output.score = -1 as f64;
                }
            }
        }
        self.prices = prices.clone();
        return output.clone();
    }
}
