use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::hodrick_prescott_smoothing_and_detrending::solver as hodrickprescottsmoothinganddetrending;
use super::CyclePolarityParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub polarity: f64,
    pub short_value: f64,
    pub long_value: f64,
    pub short_slope: f64,
    pub long_slope: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        polarity: nan,
        short_value: nan,
        long_value: nan,
        short_slope: nan,
        long_slope: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: CyclePolarityParams,
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
            let short = harmoniccycle::goertzel(values.clone(), self.p.short_period).clone();
            let long = harmoniccycle::goertzel(values.clone(), self.p.long_period).clone();
            let short_slope: f64 =
                harmoniccycle::cycle_slope(short.clone(), ((self.p.window as i64) - 1) as f64);
            let long_slope: f64 =
                harmoniccycle::cycle_slope(long.clone(), ((self.p.window as i64) - 1) as f64);
            output = Snapshot {
                polarity: harmoniccycle::cycle_polarity(short_slope, long_slope, self.p.tolerance),
                short_value: harmoniccycle::cycle_value(
                    short.clone(),
                    ((self.p.window as i64) - 1) as f64,
                ),
                long_value: harmoniccycle::cycle_value(
                    long.clone(),
                    ((self.p.window as i64) - 1) as f64,
                ),
                short_slope: short_slope,
                long_slope: long_slope,
            }
            .clone();
        }
        self.prices = prices.clone();
        return output.clone();
    }
}
