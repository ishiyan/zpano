use super::super::harmonic_cycle::model as harmoniccycle;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct WilderRSI {
    pub length: i64,
    pub count: i64,
    pub previous: f64,
    pub up: f64,
    pub down: f64,
}
impl WilderRSI {
    pub fn update(&mut self, price: f64) -> f64 {
        if (self.previous).is_nan() {
            self.previous = price;
            return f64::NAN;
        }
        let change: f64 = harmoniccycle::finite(price - self.previous);
        let (gain, loss) = ((change).max(0.0), (-change).max(0.0));
        if self.count < self.length {
            self.up = harmoniccycle::finite(self.up + gain);
            self.down = harmoniccycle::finite(self.down + loss);
            self.count += 1;
            self.previous = price;
            if self.count < self.length {
                return f64::NAN;
            }
            self.up /= self.length as f64;
            self.down /= self.length as f64;
        } else {
            self.up = harmoniccycle::finite(
                ((self.up * ((self.length - 1) as f64)) + gain) / (self.length as f64),
            );
            self.down = harmoniccycle::finite(
                ((self.down * ((self.length - 1) as f64)) + loss) / (self.length as f64),
            );
            self.previous = price;
        }
        if self.down == 0.0 {
            return 100.0;
        }
        if self.up == 0.0 {
            return 0.0;
        }
        return 100.0 - (100.0 / (1.0 + (self.up / self.down)));
    }
}
pub fn new_wilder_rsi(length: i64) -> WilderRSI {
    {
        let err = harmoniccycle::validate_integer(length, "length", 1);
        if err.is_some() {
            panic!("{:?}", err);
        }
    }
    return WilderRSI {
        length: length,
        previous: f64::NAN,
        ..Default::default()
    }
    .clone();
}
