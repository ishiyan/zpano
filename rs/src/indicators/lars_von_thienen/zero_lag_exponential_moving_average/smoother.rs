use super::super::harmonic_cycle::model as harmoniccycle;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct ZLEMA {
    pub alpha: f64,
    pub first: f64,
    pub value: f64,
    pub lag: i64,
    pub history: Vec<f64>,
}
impl ZLEMA {
    pub fn update(&mut self, price: f64) -> f64 {
        if (self.first).is_nan() {
            self.first = price;
            self.value = price;
            self.history = harmoniccycle::push(self.history.clone(), price, self.lag + 1).clone();
            return price;
        }
        let history: Vec<f64> = {
            let mut __values = {
                let mut __values = vec![].clone();
                __values.extend(self.history.clone());
                __values
            }
            .clone();
            __values.push(price);
            __values
        }
        .clone();
        let mut lagged: f64 = self.first;
        if (history.len() as i64) > self.lag {
            lagged = history[(((history.len() as i64) - 1) - self.lag) as usize];
        }
        let adjusted: f64 = harmoniccycle::finite((2.0 * price) - lagged);
        let value: f64 =
            harmoniccycle::finite((self.alpha * adjusted) + ((1.0 - self.alpha) * self.value));
        self.history = harmoniccycle::push(self.history.clone(), price, self.lag + 1).clone();
        self.value = value;
        return value;
    }
}
pub fn new_zlema(length: i64) -> ZLEMA {
    {
        let err = harmoniccycle::validate_integer(length, "length", 1);
        if err.is_some() {
            panic!("{:?}", err);
        }
    }
    return ZLEMA {
        alpha: (2.0 / ((length + 1) as f64)),
        lag: ((length - 1) / 2),
        first: f64::NAN,
        value: f64::NAN,
        history: Vec::<f64>::new().clone(),
    }
    .clone();
}
