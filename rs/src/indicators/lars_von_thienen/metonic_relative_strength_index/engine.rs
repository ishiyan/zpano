use super::super::cyclic_smoothed_relative_strength_index::rsi as cyclicsmoothedrelativestrengthindex;
use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::metonic::calendar as metonic;
use super::MetonicRelativeStrengthIndexParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub value: f64,
    pub linear_rsi: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        value: nan,
        linear_rsi: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: MetonicRelativeStrengthIndexParams,
    pub calendar: metonic::Calendar,
    pub delay: metonic::DelayMean,
    pub last_value: f64,
    pub samples_seen: i64,
    pub rsi: cyclicsmoothedrelativestrengthindex::WilderRSI,
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
        self.rsi =
            cyclicsmoothedrelativestrengthindex::new_wilder_rsi(self.p.length as i64).clone();
    }
    pub fn primed(&self) -> bool {
        return !(self.last_value).is_nan();
    }
    pub fn warmup(&self) -> i64 {
        return (self.p.delay1 as i64)
            .max(self.p.delay2 as i64)
            .max(self.p.delay3 as i64)
            + (self.p.length as i64);
    }
    pub fn update(&mut self, price: f64, day: Option<i64>) -> Snapshot {
        harmoniccycle::validate_sample(price);
        if (price).is_nan() {
            return empty_snapshot().clone();
        }
        {
            let err = self.calendar.validate(day);
            if err.is_some() {
                panic!("{:?}", err);
            }
        }
        let linear: f64 = self.rsi.update(price);
        self.last_value = self.delay.update(linear);
        self.samples_seen += 1;
        self.calendar.commit(day);
        return Snapshot {
            value: self.last_value,
            linear_rsi: linear,
        }
        .clone();
    }
}
