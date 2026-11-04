use super::super::harmonic_cycle::validation::push;
#[derive(Clone, Default)]
pub struct Calendar {
    last: Option<i64>,
    dated: Option<bool>,
}
impl Calendar {
    pub fn validate(&self, day: Option<i64>) -> Option<String> {
        if self.dated.is_some() && self.dated != Some(day.is_some()) {
            return Some("cannot mix dated and undated observations".into());
        }
        if let Some(previous) = self.last {
            if day != Some(previous + 1) {
                return Some("calendar gap: fill full OHLC before update".into());
            }
        }
        None
    }
    pub fn commit(&mut self, day: Option<i64>) {
        self.last = day;
        self.dated = Some(day.is_some());
    }
}
#[derive(Clone, Default)]
pub struct DelayMean {
    delays: Vec<i64>,
    history: Vec<f64>,
    capacity: i64,
}
impl DelayMean {
    pub fn update(&mut self, value: f64) -> f64 {
        self.history = push(self.history.clone(), value, self.capacity);
        if self.history.len() < self.capacity as usize {
            return f64::NAN;
        }
        let mut sum = 0.0;
        for &delay in &self.delays {
            let value = self.history[self.history.len() - 1 - delay as usize];
            if value.is_nan() {
                return f64::NAN;
            }
            sum += value;
        }
        sum / self.delays.len() as f64
    }
}
pub fn new_delay_mean(delays: Vec<i64>) -> DelayMean {
    let capacity = *delays.iter().max().expect("nonempty delays") + 1;
    DelayMean {
        delays,
        capacity,
        history: Vec::new(),
    }
}
/// Fill missing integer calendar-day ordinals with the preceding full payload.
pub fn fill_calendar_days<T: Clone>(records: &[(i64, T)]) -> Result<Vec<(i64, T)>, String> {
    let mut output = Vec::new();
    for (i, record) in records.iter().enumerate() {
        if i > 0 {
            let previous = &records[i - 1];
            if record.0 <= previous.0 {
                return Err("dates must increase strictly".into());
            }
            for day in previous.0 + 1..record.0 {
                output.push((day, previous.1.clone()));
            }
        }
        output.push(record.clone());
    }
    Ok(output)
}

/// Entity timestamps must represent UTC midnight on the daily calendar.
pub fn day_from_unix_seconds(time: i64) -> i64 {
    assert_eq!(time.rem_euclid(86400), 0, "use UTC midnight calendar timestamps");
    time.div_euclid(86400)
}
