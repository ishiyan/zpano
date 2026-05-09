use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Aroon indicator.
pub struct AroonParams {
    /// Lookback period. Must be >= 2. Default is 14.
    pub length: usize,
}

impl Default for AroonParams {
    fn default() -> Self {
        Self { length: 14 }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Aroon indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AroonOutput {
    /// The Aroon Up line.
    Up = 1,
    /// The Aroon Down line.
    Down = 2,
    /// The Aroon Oscillator (Up - Down).
    Osc = 3,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Tushar Chande's Aroon indicator.
///
/// Measures the number of periods since the highest high and lowest low
/// within a lookback window. Produces Up, Down, and Oscillator outputs.
pub struct Aroon {
    length: usize,
    factor: f64,
    high_buf: Vec<f64>,
    low_buf: Vec<f64>,
    buffer_index: usize,
    count: usize,
    highest_index: usize,
    lowest_index: usize,
    up: f64,
    down: f64,
    osc: f64,
    primed: bool,
    mnemonic: String,
}

impl Aroon {
    /// Creates a new Aroon indicator from the given parameters.
    pub fn new(params: &AroonParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid aroon parameters: length should be greater than 1".to_string());
        }

        let window_size = params.length + 1;
        let mnemonic = format!("aroon({})", params.length);

        Ok(Self {
            length: params.length,
            factor: 100.0 / params.length as f64,
            high_buf: vec![0.0; window_size],
            low_buf: vec![0.0; window_size],
            buffer_index: 0,
            count: 0,
            highest_index: 0,
            lowest_index: 0,
            up: f64::NAN,
            down: f64::NAN,
            osc: f64::NAN,
            primed: false,
            mnemonic,
        })
    }

    /// Core update with high and low values. Returns (up, down, osc).
    pub fn update(&mut self, high: f64, low: f64) -> (f64, f64, f64) {
        if high.is_nan() || low.is_nan() {
            return (f64::NAN, f64::NAN, f64::NAN);
        }

        let window_size = self.length + 1;
        let today = self.count;

        let pos = self.buffer_index;
        self.high_buf[pos] = high;
        self.low_buf[pos] = low;
        self.buffer_index = (self.buffer_index + 1) % window_size;
        self.count += 1;

        if self.count < window_size {
            return (self.up, self.down, self.osc);
        }

        let trailing_index = today - self.length;

        if self.count == window_size {
            // First time: scan entire window.
            self.highest_index = trailing_index;
            self.lowest_index = trailing_index;

            for i in (trailing_index + 1)..=today {
                let buf_pos = i % window_size;
                if self.high_buf[buf_pos] >= self.high_buf[self.highest_index % window_size] {
                    self.highest_index = i;
                }
                if self.low_buf[buf_pos] <= self.low_buf[self.lowest_index % window_size] {
                    self.lowest_index = i;
                }
            }
        } else {
            // Subsequent: optimized update.
            if self.highest_index < trailing_index {
                self.highest_index = trailing_index;
                for i in (trailing_index + 1)..=today {
                    let buf_pos = i % window_size;
                    if self.high_buf[buf_pos] >= self.high_buf[self.highest_index % window_size] {
                        self.highest_index = i;
                    }
                }
            } else if high >= self.high_buf[self.highest_index % window_size] {
                self.highest_index = today;
            }

            if self.lowest_index < trailing_index {
                self.lowest_index = trailing_index;
                for i in (trailing_index + 1)..=today {
                    let buf_pos = i % window_size;
                    if self.low_buf[buf_pos] <= self.low_buf[self.lowest_index % window_size] {
                        self.lowest_index = i;
                    }
                }
            } else if low <= self.low_buf[self.lowest_index % window_size] {
                self.lowest_index = today;
            }
        }

        self.up = self.factor * (self.length - (today - self.highest_index)) as f64;
        self.down = self.factor * (self.length - (today - self.lowest_index)) as f64;
        self.osc = self.up - self.down;

        if !self.primed {
            self.primed = true;
        }

        (self.up, self.down, self.osc)
    }
}

impl Indicator for Aroon {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Aroon {}", self.mnemonic);
        build_metadata(
            Identifier::Aroon,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} up", self.mnemonic),
                    description: format!("{} Up", desc),
                },
                OutputText {
                    mnemonic: format!("{} down", self.mnemonic),
                    description: format!("{} Down", desc),
                },
                OutputText {
                    mnemonic: format!("{} osc", self.mnemonic),
                    description: format!("{} Oscillator", desc),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (up, down, osc) = self.update(sample.value, sample.value);
        vec![
            Box::new(Scalar::new(sample.time, up)),
            Box::new(Scalar::new(sample.time, down)),
            Box::new(Scalar::new(sample.time, osc)),
        ]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let (up, down, osc) = self.update(sample.high, sample.low);
        vec![
            Box::new(Scalar::new(sample.time, up)),
            Box::new(Scalar::new(sample.time, down)),
            Box::new(Scalar::new(sample.time, osc)),
        ]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (sample.bid_price + sample.ask_price) / 2.0;
        let (up, down, osc) = self.update(v, v);
        vec![
            Box::new(Scalar::new(sample.time, up)),
            Box::new(Scalar::new(sample.time, down)),
            Box::new(Scalar::new(sample.time, osc)),
        ]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let (up, down, osc) = self.update(sample.price, sample.price);
        vec![
            Box::new(Scalar::new(sample.time, up)),
            Box::new(Scalar::new(sample.time, down)),
            Box::new(Scalar::new(sample.time, osc)),
        ]
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    #[test]
    fn test_aroon_length14_full_data() {
        let tolerance = 1e-6;
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();
        let expected = testdata::test_expected();

        let mut ind = Aroon::new(&AroonParams { length: 14 }).unwrap();

        for i in 0..252 {
            let (up, down, osc) = ind.update(high[i], low[i]);

            if expected[i].up.is_nan() {
                assert!(up.is_nan(), "[{}] Up: expected NaN, got {}", i, up);
                continue;
            }

            assert!((up - expected[i].up).abs() < tolerance, "[{}] Up: expected {}, got {}", i, expected[i].up, up);
            assert!((down - expected[i].down).abs() < tolerance, "[{}] Down: expected {}, got {}", i, expected[i].down, down);
            assert!((osc - expected[i].osc).abs() < tolerance, "[{}] Osc: expected {}, got {}", i, expected[i].osc, osc);
        }
    }

    #[test]
    fn test_aroon_is_primed() {
        let mut ind = Aroon::new(&AroonParams { length: 14 }).unwrap();
        let high = testdata::test_input_high();
        let low = testdata::test_input_low();

        assert!(!ind.is_primed());

        for i in 0..14 {
            ind.update(high[i], low[i]);
            assert!(!ind.is_primed(), "[{}] expected not primed", i);
        }

        ind.update(high[14], low[14]);
        assert!(ind.is_primed());
    }

    #[test]
    fn test_aroon_nan() {
        let mut ind = Aroon::new(&AroonParams { length: 14 }).unwrap();
        let (up, down, osc) = ind.update(f64::NAN, 1.0);
        assert!(up.is_nan());
        assert!(down.is_nan());
        assert!(osc.is_nan());
    }

    #[test]
    fn test_aroon_metadata() {
        let ind = Aroon::new(&AroonParams { length: 14 }).unwrap();
        let meta = ind.metadata();

        assert_eq!(meta.identifier, Identifier::Aroon);
        assert_eq!(meta.mnemonic, "aroon(14)");
        assert_eq!(meta.outputs.len(), 3);
        assert_eq!(meta.outputs[0].kind, AroonOutput::Up as i32);
        assert_eq!(meta.outputs[1].kind, AroonOutput::Down as i32);
        assert_eq!(meta.outputs[2].kind, AroonOutput::Osc as i32);
    }

    #[test]
    fn test_aroon_invalid_params() {
        assert!(Aroon::new(&AroonParams { length: 1 }).is_err());
        assert!(Aroon::new(&AroonParams { length: 0 }).is_err());
    }
}
