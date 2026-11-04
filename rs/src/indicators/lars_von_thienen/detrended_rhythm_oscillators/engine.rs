// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
use super::super::harmonic_cycle::model as harmoniccycle;
use super::DetrendedRhythmOscillatorsParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, MPL-2.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub value: f64,
    pub direction: f64,
    pub cycle_length: f64,
    pub pivot_value: f64,
    pub pivot_offset: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        value: nan,
        direction: nan,
        cycle_length: nan,
        pivot_value: nan,
        pivot_offset: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: DetrendedRhythmOscillatorsParams,
    pub fast: f64,
    pub slow: f64,
    pub window: Vec<f64>,
    pub pivots: Vec<Pivot>,
    pub direction: i64,
    pub samples_seen: i64,
    pub last_cycle_length: f64,
}
impl Engine {
    pub fn reset(&mut self) {
        self.fast = f64::NAN;
        self.slow = f64::NAN;
        self.window = Vec::new();
        self.pivots = Vec::new();
        self.direction = 0;
        self.samples_seen = 0;
        self.last_cycle_length = f64::NAN;
    }
    pub fn primed(&self) -> bool {
        return self.samples_seen > self.warmup();
    }
    pub fn warmup(&self) -> i64 {
        return rhythm_warmup(self.p.clone());
    }
    pub fn update(&mut self, price: f64) -> Snapshot {
        harmoniccycle::validate_sample(price);
        if (price).is_nan() {
            return empty_snapshot().clone();
        }
        let mut value: f64 = price;
        if self.p.detrend {
            self.fast = ema(price, self.fast, self.p.fast_length as i64);
            self.slow = ema(price, self.slow, self.p.slow_length as i64);
            value = f64::NAN;
            if self.slow != 0.0 {
                value = harmoniccycle::finite((100.0 * (self.fast - self.slow)) / self.slow);
            }
        }
        let index: i64 = self.samples_seen;
        self.samples_seen += 1;
        if (value).is_nan() {
            return empty_snapshot().clone();
        }
        self.window =
            harmoniccycle::push(self.window.clone(), value, self.p.zig_period as i64).clone();
        let (mut maximum, mut minimum) = (self.window[(0) as usize], self.window[(0) as usize]);
        for item in self.window.clone() {
            maximum = (maximum).max(item);
            minimum = (minimum).min(item);
        }
        let (high, low) = ((value == maximum), (value == minimum));
        let mut direction: i64 = self.direction;
        if high && !low {
            direction = 1;
        } else if low && !high {
            direction = -1;
        }
        if (direction != 0) && (high || low) {
            if (direction != self.direction) || self.pivots.is_empty() {
                self.pivots = {
                    let mut __values = vec![Pivot {
                        value: value,
                        index: index,
                    }]
                    .clone();
                    __values.extend(self.pivots.clone());
                    __values
                }
                .clone();
                if (self.pivots.len() as i64) > ((self.p.cycle_memory as i64) + 3) {
                    self.pivots = self.pivots[0..((self.p.cycle_memory as i64) + 3) as usize]
                        .to_vec()
                        .clone();
                }
            } else if ((direction == 1) && (value > self.pivots[(0) as usize].value))
                || ((direction == -1) && (value < self.pivots[(0) as usize].value))
            {
                self.pivots[(0) as usize] = Pivot {
                    value: value,
                    index: index,
                }
                .clone();
            }
        }
        self.direction = direction;
        let mut cycles: Vec<f64> = Vec::<f64>::new().clone();
        if (self.pivots.len() as i64) >= 5 {
            {
                let mut i: i64 = 1;
                while i < ((self.pivots.len() as i64) - 2) {
                    cycles = {
                        let mut __values = cycles.clone();
                        __values.push(
                            (self.pivots[(i) as usize].index - self.pivots[(i + 2) as usize].index)
                                as f64,
                        );
                        __values
                    }
                    .clone();
                    i += 1;
                }
            }
        }
        if (cycles.len() as i64) > (self.p.cycle_memory as i64) {
            cycles = cycles[((cycles.len() as i64) - (self.p.cycle_memory as i64)) as usize..]
                .to_vec()
                .clone();
        }
        self.last_cycle_length = median(cycles.clone());
        if !self.primed() {
            return empty_snapshot().clone();
        }
        let (mut pivot_value, mut offset) = (f64::NAN, f64::NAN);
        if !self.pivots.is_empty() {
            pivot_value = self.pivots[(0) as usize].value;
            offset = (self.pivots[(0) as usize].index - index) as f64;
        }
        return Snapshot {
            value: value,
            direction: (direction as f64),
            cycle_length: self.last_cycle_length,
            pivot_value: pivot_value,
            pivot_offset: offset,
        }
        .clone();
    }
}
#[derive(Clone, Default)]
pub struct Pivot {
    pub value: f64,
    pub index: i64,
}
impl Pivot {}
pub fn ema(value: f64, previous: f64, length: i64) -> f64 {
    if (previous).is_nan() {
        return value;
    }
    let alpha: f64 = 2.0 / ((length + 1) as f64);
    return harmoniccycle::finite((alpha * value) + ((1.0 - alpha) * previous));
}
pub fn rhythm_warmup(p: DetrendedRhythmOscillatorsParams) -> i64 {
    if p.detrend {
        return p.slow_length as i64 - 1;
    }
    return 0;
}
pub fn median(mut values: Vec<f64>) -> f64 {
    if values.is_empty() || ((values.len() as i64) < 1) {
        return f64::NAN;
    }
    values.sort_by(|a, b| a.total_cmp(b));
    let middle: i64 = (values.len() as i64) / 2;
    if ((values.len() as i64) % 2) != 0 {
        return values[(middle) as usize];
    }
    return (values[(middle - 1) as usize] + values[(middle) as usize]) / 2.0;
}
