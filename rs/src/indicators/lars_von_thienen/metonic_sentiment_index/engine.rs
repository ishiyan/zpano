use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::metonic::calendar as metonic;
use super::super::zero_lag_exponential_moving_average::smoother as zerolagexponentialmovingaverage;
use super::MetonicSentimentIndexParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub value: f64,
    pub daily_sentiment: f64,
    pub smoothed_close: f64,
    pub smoothed_open: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        value: nan,
        daily_sentiment: nan,
        smoothed_close: nan,
        smoothed_open: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: MetonicSentimentIndexParams,
    pub calendar: metonic::Calendar,
    pub delay: metonic::DelayMean,
    pub last_value: f64,
    pub samples_seen: i64,
    pub close_smoother: zerolagexponentialmovingaverage::ZLEMA,
    pub open_smoother: zerolagexponentialmovingaverage::ZLEMA,
    pub previous_close: f64,
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
        self.close_smoother =
            zerolagexponentialmovingaverage::new_zlema(self.p.length as i64).clone();
        self.open_smoother =
            zerolagexponentialmovingaverage::new_zlema(self.p.length as i64).clone();
        self.previous_close = f64::NAN;
    }
    pub fn primed(&self) -> bool {
        return !(self.last_value).is_nan();
    }
    pub fn warmup(&self) -> i64 {
        return ((self.p.delay1 as i64)
            .max(self.p.delay2 as i64)
            .max(self.p.delay3 as i64)
            + (((self.p.length as i64) - 1) / 2))
            + (self.p.mode as i64);
    }
    pub fn update(&mut self, price: f64, day: Option<i64>, open_price: f64) -> Snapshot {
        harmoniccycle::validate_sample(price);
        if (price).is_nan() {
            return empty_snapshot().clone();
        }
        if (self.p.mode as i64) == 0 {
            harmoniccycle::validate_sample(open_price);
            if (open_price).is_nan() {
                return empty_snapshot().clone();
            }
        }
        {
            let err = self.calendar.validate(day);
            if err.is_some() {
                panic!("{:?}", err);
            }
        }
        let smoothed_close: f64 = self.close_smoother.update(price);
        let mut smoothed_open: f64 = f64::NAN;
        let mut comparison: f64 = self.previous_close;
        if (self.p.mode as i64) == 0 {
            smoothed_open = self.open_smoother.update(open_price);
            comparison = smoothed_open;
        }
        let mut sentiment: f64 = f64::NAN;
        if (self.samples_seen >= ((((self.p.length as i64) - 1) / 2) + (self.p.mode as i64)))
            && !(comparison).is_nan()
        {
            sentiment = -1 as f64;
            if smoothed_close > comparison {
                sentiment = 1.0;
            }
        }
        self.previous_close = smoothed_close;
        let mean: f64 = self.delay.update(sentiment);
        self.last_value = 3.0 * mean;
        self.samples_seen += 1;
        self.calendar.commit(day);
        return Snapshot {
            value: self.last_value,
            daily_sentiment: sentiment,
            smoothed_close: smoothed_close,
            smoothed_open: smoothed_open,
        }
        .clone();
    }
}
