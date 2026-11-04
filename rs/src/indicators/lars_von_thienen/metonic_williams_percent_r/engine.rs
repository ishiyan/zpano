use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::metonic::calendar as metonic;
use super::MetonicWilliamsPercentRParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub value: f64,
    pub linear_percent_r: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        value: nan,
        linear_percent_r: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: MetonicWilliamsPercentRParams,
    pub calendar: metonic::Calendar,
    pub delay: metonic::DelayMean,
    pub last_value: f64,
    pub samples_seen: i64,
    pub highs: Vec<f64>,
    pub lows: Vec<f64>,
}
impl Engine {
    pub fn reset(&mut self) {
        self.calendar = metonic::Calendar::default();
        self.delay = metonic::new_delay_mean(
            vec![
                (self.p.delay1 as i64),
                (self.p.delay2 as i64),
                (self.p.delay3 as i64),
            ]
            .clone(),
        )
        .clone();
        self.last_value = f64::NAN;
        self.samples_seen = 0;
        self.highs = Vec::new();
        self.lows = Vec::new();
    }
    pub fn primed(&self) -> bool {
        return !(self.last_value).is_nan();
    }
    pub fn warmup(&self) -> i64 {
        return ((self.p.delay1 as i64)
            .max(self.p.delay2 as i64)
            .max(self.p.delay3 as i64)
            + (self.p.length as i64))
            - 1;
    }
    pub fn update(&mut self, price: f64, day: Option<i64>, high: f64, low: f64) -> Snapshot {
        harmoniccycle::validate_sample(price);
        if (price).is_nan() {
            return empty_snapshot().clone();
        }
        harmoniccycle::validate_sample(high);
        harmoniccycle::validate_sample(low);
        if (high).is_nan() || (low).is_nan() {
            return empty_snapshot().clone();
        }
        if (low > price) || (price > high) {
            panic!(
                "{:?}",
                "completed HLC bar must satisfy low <= close <= high"
            );
        }
        {
            let err = self.calendar.validate(day);
            if err.is_some() {
                panic!("{:?}", err);
            }
        }
        self.highs = harmoniccycle::push(self.highs.clone(), high, self.p.length as i64).clone();
        self.lows = harmoniccycle::push(self.lows.clone(), low, self.p.length as i64).clone();
        let mut linear: f64 = f64::NAN;
        if (self.highs.len() as i64) == (self.p.length as i64) {
            let (mut maximum, mut minimum) = (self.highs[(0) as usize], self.lows[(0) as usize]);
            for value in self.highs.clone() {
                maximum = (maximum).max(value);
            }
            for value in self.lows.clone() {
                minimum = (minimum).min(value);
            }
            if maximum > minimum {
                linear = ((-100 as f64) * (maximum - price)) / (maximum - minimum);
            }
        }
        self.last_value = self.delay.update(linear);
        self.samples_seen += 1;
        self.calendar.commit(day);
        return Snapshot {
            value: self.last_value,
            linear_percent_r: linear,
        }
        .clone();
    }
}
