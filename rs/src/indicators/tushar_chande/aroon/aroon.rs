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

    fn test_input_high() -> Vec<f64> {
        vec![
            93.25, 94.94, 96.375, 96.19, 96.0, 94.72, 95.0, 93.72, 92.47, 92.75,
            96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
            90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
            87.155, 87.25, 87.375, 88.97, 90.0, 89.845, 86.97, 85.94, 84.75, 85.47,
            84.47, 88.5, 89.47, 90.0, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
            90.0, 88.875, 89.0, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
            90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92.0, 91.94,
            89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
            104.5, 105.5, 106.125, 107.94, 106.25, 107.0, 108.75, 110.94, 110.94, 114.22,
            123.0, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113.0, 118.315,
            116.87, 116.75, 113.87, 114.62, 115.31, 116.0, 121.69, 119.87, 120.87, 116.75,
            116.5, 116.0, 118.31, 121.5, 122.0, 121.44, 125.75, 127.75, 124.19, 124.44,
            125.75, 124.69, 125.31, 132.0, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
            138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
            125.12, 126.5, 128.69, 126.62, 126.69, 126.0, 123.12, 121.87, 124.0, 127.0,
            124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
            124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126.0, 128.5, 126.75, 129.75,
            132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135.0, 132.38, 131.44, 130.88,
            129.63, 127.25, 127.81, 125.0, 126.81, 124.75, 122.81, 122.25, 121.06, 120.0,
            123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109.0,
            108.5, 113.06, 93.0, 94.62, 95.12, 96.0, 95.56, 95.31, 99.0, 98.81,
            96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
            95.0, 94.87, 98.25, 105.12, 108.44, 109.87, 105.0, 106.0, 104.94, 104.5,
            104.44, 106.31, 112.87, 116.5, 119.19, 121.0, 122.12, 111.94, 112.75, 110.19,
            107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110.0, 110.75, 110.5,
            110.5, 109.5,
        ]
    }

    fn test_input_low() -> Vec<f64> {
        vec![
            90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92.0, 89.75, 89.44, 90.625,
            92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
            87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
            84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
            82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89.0, 88.565, 90.095,
            89.0, 86.47, 84.0, 83.315, 82.0, 83.25, 84.75, 85.28, 87.19, 88.44,
            88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92.0, 90.53, 89.97, 88.815,
            86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
            101.72, 101.72, 103.155, 105.69, 103.655, 104.0, 105.53, 108.53, 108.75, 107.75,
            117.0, 118.0, 116.0, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
            112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
            112.56, 114.0, 114.75, 118.87, 119.0, 119.75, 122.62, 123.0, 121.75, 121.56,
            123.12, 122.19, 122.75, 124.37, 128.0, 129.5, 130.81, 130.63, 132.13, 133.88,
            135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
            122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121.0,
            121.12, 118.94, 119.81, 121.0, 122.0, 124.5, 126.56, 123.5, 121.25, 121.06,
            122.31, 121.0, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128.0,
            128.38, 130.69, 131.63, 134.38, 132.0, 131.94, 131.94, 129.56, 123.75, 126.0,
            126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120.0, 119.62, 115.5, 116.75,
            119.06, 119.06, 115.06, 111.06, 113.12, 110.0, 105.0, 104.69, 103.87, 104.69,
            105.44, 107.0, 89.0, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
            93.69, 93.5, 90.0, 90.19, 90.5, 92.12, 94.12, 94.87, 93.0, 93.87,
            93.0, 92.62, 93.56, 98.37, 104.44, 106.0, 101.81, 104.12, 103.37, 102.12,
            102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
            104.5, 105.75, 108.62, 107.75, 108.06, 108.0, 108.19, 108.12, 109.06, 108.75,
            108.56, 106.62,
        ]
    }

    struct ExpectedRow {
        up: f64,
        down: f64,
        osc: f64,
    }

    fn test_expected() -> Vec<ExpectedRow> {
        let nan = f64::NAN;
        vec![
            ExpectedRow { up: nan, down: nan, osc: nan }, // 0
            ExpectedRow { up: nan, down: nan, osc: nan }, // 1
            ExpectedRow { up: nan, down: nan, osc: nan }, // 2
            ExpectedRow { up: nan, down: nan, osc: nan }, // 3
            ExpectedRow { up: nan, down: nan, osc: nan }, // 4
            ExpectedRow { up: nan, down: nan, osc: nan }, // 5
            ExpectedRow { up: nan, down: nan, osc: nan }, // 6
            ExpectedRow { up: nan, down: nan, osc: nan }, // 7
            ExpectedRow { up: nan, down: nan, osc: nan }, // 8
            ExpectedRow { up: nan, down: nan, osc: nan }, // 9
            ExpectedRow { up: nan, down: nan, osc: nan }, // 10
            ExpectedRow { up: nan, down: nan, osc: nan }, // 11
            ExpectedRow { up: nan, down: nan, osc: nan }, // 12
            ExpectedRow { up: nan, down: nan, osc: nan }, // 13
            ExpectedRow { up: 78.571428571428600, down: 100.000000000000000, osc: -21.428571428571400 },
            ExpectedRow { up: 71.428571428571400, down: 92.857142857142900, osc: -21.428571428571400 },
            ExpectedRow { up: 64.285714285714300, down: 85.714285714285700, osc: -21.428571428571400 },
            ExpectedRow { up: 57.142857142857100, down: 78.571428571428600, osc: -21.428571428571400 },
            ExpectedRow { up: 50.000000000000000, down: 71.428571428571400, osc: -21.428571428571400 },
            ExpectedRow { up: 42.857142857142900, down: 64.285714285714300, osc: -21.428571428571400 },
            ExpectedRow { up: 35.714285714285700, down: 57.142857142857100, osc: -21.428571428571400 },
            ExpectedRow { up: 28.571428571428600, down: 100.000000000000000, osc: -71.428571428571400 },
            ExpectedRow { up: 21.428571428571400, down: 100.000000000000000, osc: -78.571428571428600 },
            ExpectedRow { up: 14.285714285714300, down: 100.000000000000000, osc: -85.714285714285700 },
            ExpectedRow { up: 7.142857142857140, down: 100.000000000000000, osc: -92.857142857142900 },
            ExpectedRow { up: 0.000000000000000, down: 100.000000000000000, osc: -100.000000000000000 },
            ExpectedRow { up: 0.000000000000000, down: 92.857142857142900, osc: -92.857142857142900 },
            ExpectedRow { up: 21.428571428571400, down: 85.714285714285700, osc: -64.285714285714300 },
            ExpectedRow { up: 14.285714285714300, down: 78.571428571428600, osc: -64.285714285714300 },
            ExpectedRow { up: 7.142857142857140, down: 71.428571428571400, osc: -64.285714285714300 },
            ExpectedRow { up: 0.000000000000000, down: 64.285714285714300, osc: -64.285714285714300 },
            ExpectedRow { up: 14.285714285714300, down: 57.142857142857100, osc: -42.857142857142900 },
            ExpectedRow { up: 7.142857142857140, down: 50.000000000000000, osc: -42.857142857142900 },
            ExpectedRow { up: 0.000000000000000, down: 42.857142857142900, osc: -42.857142857142900 },
            ExpectedRow { up: 0.000000000000000, down: 35.714285714285700, osc: -35.714285714285700 },
            ExpectedRow { up: 92.857142857142900, down: 28.571428571428600, osc: 64.285714285714300 },
            ExpectedRow { up: 85.714285714285700, down: 21.428571428571400, osc: 64.285714285714300 },
            ExpectedRow { up: 78.571428571428600, down: 14.285714285714300, osc: 64.285714285714300 },
            ExpectedRow { up: 71.428571428571400, down: 7.142857142857140, osc: 64.285714285714300 },
            ExpectedRow { up: 64.285714285714300, down: 0.000000000000000, osc: 64.285714285714300 },
            ExpectedRow { up: 57.142857142857100, down: 0.000000000000000, osc: 57.142857142857100 },
            ExpectedRow { up: 50.000000000000000, down: 92.857142857142900, osc: -42.857142857142900 },
            ExpectedRow { up: 42.857142857142900, down: 85.714285714285700, osc: -42.857142857142900 },
            ExpectedRow { up: 100.000000000000000, down: 78.571428571428600, osc: 21.428571428571400 },
            ExpectedRow { up: 100.000000000000000, down: 71.428571428571400, osc: 28.571428571428600 },
            ExpectedRow { up: 92.857142857142900, down: 64.285714285714300, osc: 28.571428571428600 },
            ExpectedRow { up: 100.000000000000000, down: 57.142857142857100, osc: 42.857142857142900 },
            ExpectedRow { up: 92.857142857142900, down: 50.000000000000000, osc: 42.857142857142900 },
            ExpectedRow { up: 85.714285714285700, down: 42.857142857142900, osc: 42.857142857142900 },
            ExpectedRow { up: 78.571428571428600, down: 35.714285714285700, osc: 42.857142857142900 },
            ExpectedRow { up: 71.428571428571400, down: 28.571428571428600, osc: 42.857142857142900 },
            ExpectedRow { up: 64.285714285714300, down: 21.428571428571400, osc: 42.857142857142900 },
            ExpectedRow { up: 57.142857142857100, down: 14.285714285714300, osc: 42.857142857142900 },
            ExpectedRow { up: 50.000000000000000, down: 7.142857142857140, osc: 42.857142857142900 },
            ExpectedRow { up: 42.857142857142900, down: 100.000000000000000, osc: -57.142857142857100 },
            ExpectedRow { up: 35.714285714285700, down: 92.857142857142900, osc: -57.142857142857100 },
            ExpectedRow { up: 28.571428571428600, down: 85.714285714285700, osc: -57.142857142857100 },
            ExpectedRow { up: 21.428571428571400, down: 78.571428571428600, osc: -57.142857142857100 },
            ExpectedRow { up: 14.285714285714300, down: 71.428571428571400, osc: -57.142857142857100 },
            ExpectedRow { up: 7.142857142857140, down: 64.285714285714300, osc: -57.142857142857200 },
            ExpectedRow { up: 0.000000000000000, down: 57.142857142857100, osc: -57.142857142857100 },
            ExpectedRow { up: 14.285714285714300, down: 50.000000000000000, osc: -35.714285714285700 },
            ExpectedRow { up: 100.000000000000000, down: 42.857142857142900, osc: 57.142857142857100 },
            ExpectedRow { up: 100.000000000000000, down: 35.714285714285700, osc: 64.285714285714300 },
            ExpectedRow { up: 100.000000000000000, down: 28.571428571428600, osc: 71.428571428571400 },
            ExpectedRow { up: 100.000000000000000, down: 21.428571428571400, osc: 78.571428571428600 },
            ExpectedRow { up: 100.000000000000000, down: 14.285714285714300, osc: 85.714285714285700 },
            ExpectedRow { up: 92.857142857142900, down: 7.142857142857140, osc: 85.714285714285700 },
            ExpectedRow { up: 85.714285714285700, down: 0.000000000000000, osc: 85.714285714285700 },
            ExpectedRow { up: 78.571428571428600, down: 0.000000000000000, osc: 78.571428571428600 },
            ExpectedRow { up: 71.428571428571400, down: 0.000000000000000, osc: 71.428571428571400 },
            ExpectedRow { up: 64.285714285714300, down: 100.000000000000000, osc: -35.714285714285700 },
            ExpectedRow { up: 57.142857142857100, down: 100.000000000000000, osc: -42.857142857142900 },
            ExpectedRow { up: 50.000000000000000, down: 100.000000000000000, osc: -50.000000000000000 },
            ExpectedRow { up: 42.857142857142900, down: 92.857142857142900, osc: -50.000000000000000 },
            ExpectedRow { up: 100.000000000000000, down: 85.714285714285700, osc: 14.285714285714300 },
            ExpectedRow { up: 100.000000000000000, down: 78.571428571428600, osc: 21.428571428571400 },
            ExpectedRow { up: 100.000000000000000, down: 71.428571428571400, osc: 28.571428571428600 },
            ExpectedRow { up: 100.000000000000000, down: 64.285714285714300, osc: 35.714285714285700 },
            ExpectedRow { up: 92.857142857142900, down: 57.142857142857100, osc: 35.714285714285700 },
            ExpectedRow { up: 85.714285714285700, down: 50.000000000000000, osc: 35.714285714285700 },
            ExpectedRow { up: 78.571428571428600, down: 42.857142857142900, osc: 35.714285714285700 },
            ExpectedRow { up: 71.428571428571400, down: 35.714285714285700, osc: 35.714285714285700 },
            ExpectedRow { up: 100.000000000000000, down: 28.571428571428600, osc: 71.428571428571400 },
            ExpectedRow { up: 92.857142857142900, down: 21.428571428571400, osc: 71.428571428571400 },
            ExpectedRow { up: 85.714285714285700, down: 14.285714285714300, osc: 71.428571428571400 },
            ExpectedRow { up: 100.000000000000000, down: 7.142857142857140, osc: 92.857142857142900 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 92.857142857142900, down: 0.000000000000000, osc: 92.857142857142900 },
            ExpectedRow { up: 85.714285714285700, down: 21.428571428571400, osc: 64.285714285714300 },
            ExpectedRow { up: 78.571428571428600, down: 14.285714285714300, osc: 64.285714285714300 },
            ExpectedRow { up: 71.428571428571400, down: 7.142857142857140, osc: 64.285714285714300 },
            ExpectedRow { up: 64.285714285714300, down: 0.000000000000000, osc: 64.285714285714300 },
            ExpectedRow { up: 57.142857142857100, down: 0.000000000000000, osc: 57.142857142857100 },
            ExpectedRow { up: 50.000000000000000, down: 7.142857142857140, osc: 42.857142857142900 },
            ExpectedRow { up: 42.857142857142900, down: 0.000000000000000, osc: 42.857142857142900 },
            ExpectedRow { up: 35.714285714285700, down: 0.000000000000000, osc: 35.714285714285700 },
            ExpectedRow { up: 28.571428571428600, down: 0.000000000000000, osc: 28.571428571428600 },
            ExpectedRow { up: 21.428571428571400, down: 14.285714285714300, osc: 7.142857142857140 },
            ExpectedRow { up: 14.285714285714300, down: 7.142857142857140, osc: 7.142857142857140 },
            ExpectedRow { up: 7.142857142857140, down: 0.000000000000000, osc: 7.142857142857140 },
            ExpectedRow { up: 0.000000000000000, down: 92.857142857142900, osc: -92.857142857142900 },
            ExpectedRow { up: 0.000000000000000, down: 85.714285714285700, osc: -85.714285714285700 },
            ExpectedRow { up: 100.000000000000000, down: 78.571428571428600, osc: 21.428571428571400 },
            ExpectedRow { up: 92.857142857142900, down: 71.428571428571400, osc: 21.428571428571400 },
            ExpectedRow { up: 85.714285714285700, down: 64.285714285714300, osc: 21.428571428571400 },
            ExpectedRow { up: 78.571428571428600, down: 57.142857142857100, osc: 21.428571428571400 },
            ExpectedRow { up: 71.428571428571400, down: 50.000000000000000, osc: 21.428571428571400 },
            ExpectedRow { up: 64.285714285714300, down: 42.857142857142900, osc: 21.428571428571400 },
            ExpectedRow { up: 57.142857142857100, down: 35.714285714285700, osc: 21.428571428571400 },
            ExpectedRow { up: 50.000000000000000, down: 28.571428571428600, osc: 21.428571428571400 },
            ExpectedRow { up: 100.000000000000000, down: 21.428571428571400, osc: 78.571428571428600 },
            ExpectedRow { up: 92.857142857142900, down: 14.285714285714300, osc: 78.571428571428600 },
            ExpectedRow { up: 100.000000000000000, down: 7.142857142857140, osc: 92.857142857142900 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 92.857142857142900, down: 0.000000000000000, osc: 92.857142857142900 },
            ExpectedRow { up: 85.714285714285700, down: 0.000000000000000, osc: 85.714285714285700 },
            ExpectedRow { up: 78.571428571428600, down: 28.571428571428600, osc: 50.000000000000000 },
            ExpectedRow { up: 71.428571428571400, down: 21.428571428571400, osc: 50.000000000000000 },
            ExpectedRow { up: 64.285714285714300, down: 14.285714285714300, osc: 50.000000000000000 },
            ExpectedRow { up: 100.000000000000000, down: 7.142857142857140, osc: 92.857142857142900 },
            ExpectedRow { up: 92.857142857142900, down: 0.000000000000000, osc: 92.857142857142900 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 92.857142857142900, down: 0.000000000000000, osc: 92.857142857142900 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 100.000000000000000, down: 21.428571428571400, osc: 78.571428571428600 },
            ExpectedRow { up: 100.000000000000000, down: 14.285714285714300, osc: 85.714285714285700 },
            ExpectedRow { up: 92.857142857142900, down: 7.142857142857140, osc: 85.714285714285700 },
            ExpectedRow { up: 85.714285714285700, down: 0.000000000000000, osc: 85.714285714285700 },
            ExpectedRow { up: 78.571428571428600, down: 7.142857142857140, osc: 71.428571428571400 },
            ExpectedRow { up: 71.428571428571400, down: 0.000000000000000, osc: 71.428571428571400 },
            ExpectedRow { up: 64.285714285714300, down: 0.000000000000000, osc: 64.285714285714300 },
            ExpectedRow { up: 57.142857142857100, down: 0.000000000000000, osc: 57.142857142857100 },
            ExpectedRow { up: 50.000000000000000, down: 100.000000000000000, osc: -50.000000000000000 },
            ExpectedRow { up: 42.857142857142900, down: 100.000000000000000, osc: -57.142857142857100 },
            ExpectedRow { up: 35.714285714285700, down: 92.857142857142900, osc: -57.142857142857100 },
            ExpectedRow { up: 28.571428571428600, down: 85.714285714285700, osc: -57.142857142857100 },
            ExpectedRow { up: 21.428571428571400, down: 78.571428571428600, osc: -57.142857142857100 },
            ExpectedRow { up: 14.285714285714300, down: 71.428571428571400, osc: -57.142857142857100 },
            ExpectedRow { up: 7.142857142857140, down: 64.285714285714300, osc: -57.142857142857200 },
            ExpectedRow { up: 0.000000000000000, down: 100.000000000000000, osc: -100.000000000000000 },
            ExpectedRow { up: 21.428571428571400, down: 100.000000000000000, osc: -78.571428571428600 },
            ExpectedRow { up: 14.285714285714300, down: 100.000000000000000, osc: -85.714285714285700 },
            ExpectedRow { up: 7.142857142857140, down: 100.000000000000000, osc: -92.857142857142900 },
            ExpectedRow { up: 0.000000000000000, down: 92.857142857142900, osc: -92.857142857142900 },
            ExpectedRow { up: 0.000000000000000, down: 85.714285714285700, osc: -85.714285714285700 },
            ExpectedRow { up: 0.000000000000000, down: 78.571428571428600, osc: -78.571428571428600 },
            ExpectedRow { up: 28.571428571428600, down: 71.428571428571400, osc: -42.857142857142900 },
            ExpectedRow { up: 21.428571428571400, down: 64.285714285714300, osc: -42.857142857142900 },
            ExpectedRow { up: 14.285714285714300, down: 57.142857142857100, osc: -42.857142857142900 },
            ExpectedRow { up: 7.142857142857140, down: 50.000000000000000, osc: -42.857142857142900 },
            ExpectedRow { up: 0.000000000000000, down: 42.857142857142900, osc: -42.857142857142900 },
            ExpectedRow { up: 100.000000000000000, down: 35.714285714285700, osc: 64.285714285714300 },
            ExpectedRow { up: 92.857142857142900, down: 28.571428571428600, osc: 64.285714285714300 },
            ExpectedRow { up: 85.714285714285700, down: 21.428571428571400, osc: 64.285714285714300 },
            ExpectedRow { up: 78.571428571428600, down: 14.285714285714300, osc: 64.285714285714300 },
            ExpectedRow { up: 71.428571428571400, down: 7.142857142857140, osc: 64.285714285714300 },
            ExpectedRow { up: 64.285714285714300, down: 0.000000000000000, osc: 64.285714285714300 },
            ExpectedRow { up: 57.142857142857100, down: 14.285714285714300, osc: 42.857142857142900 },
            ExpectedRow { up: 50.000000000000000, down: 7.142857142857140, osc: 42.857142857142900 },
            ExpectedRow { up: 42.857142857142900, down: 0.000000000000000, osc: 42.857142857142900 },
            ExpectedRow { up: 35.714285714285700, down: 0.000000000000000, osc: 35.714285714285700 },
            ExpectedRow { up: 28.571428571428600, down: 64.285714285714300, osc: -35.714285714285700 },
            ExpectedRow { up: 21.428571428571400, down: 57.142857142857100, osc: -35.714285714285700 },
            ExpectedRow { up: 100.000000000000000, down: 50.000000000000000, osc: 50.000000000000000 },
            ExpectedRow { up: 100.000000000000000, down: 42.857142857142900, osc: 57.142857142857100 },
            ExpectedRow { up: 100.000000000000000, down: 35.714285714285700, osc: 64.285714285714300 },
            ExpectedRow { up: 100.000000000000000, down: 28.571428571428600, osc: 71.428571428571400 },
            ExpectedRow { up: 100.000000000000000, down: 21.428571428571400, osc: 78.571428571428600 },
            ExpectedRow { up: 92.857142857142900, down: 14.285714285714300, osc: 78.571428571428600 },
            ExpectedRow { up: 85.714285714285700, down: 7.142857142857140, osc: 78.571428571428600 },
            ExpectedRow { up: 78.571428571428600, down: 0.000000000000000, osc: 78.571428571428600 },
            ExpectedRow { up: 71.428571428571400, down: 0.000000000000000, osc: 71.428571428571400 },
            ExpectedRow { up: 64.285714285714300, down: 7.142857142857140, osc: 57.142857142857200 },
            ExpectedRow { up: 57.142857142857100, down: 0.000000000000000, osc: 57.142857142857100 },
            ExpectedRow { up: 50.000000000000000, down: 0.000000000000000, osc: 50.000000000000000 },
            ExpectedRow { up: 42.857142857142900, down: 78.571428571428600, osc: -35.714285714285700 },
            ExpectedRow { up: 35.714285714285700, down: 100.000000000000000, osc: -64.285714285714300 },
            ExpectedRow { up: 28.571428571428600, down: 100.000000000000000, osc: -71.428571428571400 },
            ExpectedRow { up: 21.428571428571400, down: 92.857142857142900, osc: -71.428571428571400 },
            ExpectedRow { up: 14.285714285714300, down: 85.714285714285700, osc: -71.428571428571400 },
            ExpectedRow { up: 7.142857142857140, down: 100.000000000000000, osc: -92.857142857142900 },
            ExpectedRow { up: 0.000000000000000, down: 100.000000000000000, osc: -100.000000000000000 },
            ExpectedRow { up: 0.000000000000000, down: 100.000000000000000, osc: -100.000000000000000 },
            ExpectedRow { up: 7.142857142857140, down: 92.857142857142900, osc: -85.714285714285700 },
            ExpectedRow { up: 0.000000000000000, down: 85.714285714285700, osc: -85.714285714285700 },
            ExpectedRow { up: 0.000000000000000, down: 78.571428571428600, osc: -78.571428571428600 },
            ExpectedRow { up: 0.000000000000000, down: 100.000000000000000, osc: -100.000000000000000 },
            ExpectedRow { up: 0.000000000000000, down: 100.000000000000000, osc: -100.000000000000000 },
            ExpectedRow { up: 0.000000000000000, down: 92.857142857142900, osc: -92.857142857142900 },
            ExpectedRow { up: 7.142857142857140, down: 100.000000000000000, osc: -92.857142857142900 },
            ExpectedRow { up: 0.000000000000000, down: 100.000000000000000, osc: -100.000000000000000 },
            ExpectedRow { up: 7.142857142857140, down: 100.000000000000000, osc: -92.857142857142900 },
            ExpectedRow { up: 0.000000000000000, down: 100.000000000000000, osc: -100.000000000000000 },
            ExpectedRow { up: 0.000000000000000, down: 92.857142857142900, osc: -92.857142857142900 },
            ExpectedRow { up: 28.571428571428600, down: 85.714285714285700, osc: -57.142857142857100 },
            ExpectedRow { up: 21.428571428571400, down: 78.571428571428600, osc: -57.142857142857100 },
            ExpectedRow { up: 14.285714285714300, down: 100.000000000000000, osc: -85.714285714285700 },
            ExpectedRow { up: 7.142857142857140, down: 92.857142857142900, osc: -85.714285714285700 },
            ExpectedRow { up: 0.000000000000000, down: 85.714285714285700, osc: -85.714285714285700 },
            ExpectedRow { up: 0.000000000000000, down: 78.571428571428600, osc: -78.571428571428600 },
            ExpectedRow { up: 0.000000000000000, down: 71.428571428571400, osc: -71.428571428571400 },
            ExpectedRow { up: 7.142857142857140, down: 64.285714285714300, osc: -57.142857142857200 },
            ExpectedRow { up: 0.000000000000000, down: 57.142857142857100, osc: -57.142857142857100 },
            ExpectedRow { up: 0.000000000000000, down: 50.000000000000000, osc: -50.000000000000000 },
            ExpectedRow { up: 35.714285714285700, down: 42.857142857142900, osc: -7.142857142857140 },
            ExpectedRow { up: 28.571428571428600, down: 35.714285714285700, osc: -7.142857142857150 },
            ExpectedRow { up: 21.428571428571400, down: 28.571428571428600, osc: -7.142857142857140 },
            ExpectedRow { up: 14.285714285714300, down: 21.428571428571400, osc: -7.142857142857140 },
            ExpectedRow { up: 7.142857142857140, down: 14.285714285714300, osc: -7.142857142857140 },
            ExpectedRow { up: 0.000000000000000, down: 7.142857142857140, osc: -7.142857142857140 },
            ExpectedRow { up: 42.857142857142900, down: 0.000000000000000, osc: 42.857142857142900 },
            ExpectedRow { up: 35.714285714285700, down: 64.285714285714300, osc: -28.571428571428600 },
            ExpectedRow { up: 28.571428571428600, down: 57.142857142857100, osc: -28.571428571428600 },
            ExpectedRow { up: 21.428571428571400, down: 50.000000000000000, osc: -28.571428571428600 },
            ExpectedRow { up: 14.285714285714300, down: 42.857142857142900, osc: -28.571428571428600 },
            ExpectedRow { up: 7.142857142857140, down: 35.714285714285700, osc: -28.571428571428600 },
            ExpectedRow { up: 0.000000000000000, down: 28.571428571428600, osc: -28.571428571428600 },
            ExpectedRow { up: 100.000000000000000, down: 21.428571428571400, osc: 78.571428571428600 },
            ExpectedRow { up: 100.000000000000000, down: 14.285714285714300, osc: 85.714285714285700 },
            ExpectedRow { up: 100.000000000000000, down: 7.142857142857140, osc: 92.857142857142900 },
            ExpectedRow { up: 92.857142857142900, down: 0.000000000000000, osc: 92.857142857142900 },
            ExpectedRow { up: 85.714285714285700, down: 0.000000000000000, osc: 85.714285714285700 },
            ExpectedRow { up: 78.571428571428600, down: 0.000000000000000, osc: 78.571428571428600 },
            ExpectedRow { up: 71.428571428571400, down: 0.000000000000000, osc: 71.428571428571400 },
            ExpectedRow { up: 64.285714285714300, down: 35.714285714285700, osc: 28.571428571428600 },
            ExpectedRow { up: 57.142857142857100, down: 28.571428571428600, osc: 28.571428571428600 },
            ExpectedRow { up: 100.000000000000000, down: 21.428571428571400, osc: 78.571428571428600 },
            ExpectedRow { up: 100.000000000000000, down: 14.285714285714300, osc: 85.714285714285700 },
            ExpectedRow { up: 100.000000000000000, down: 7.142857142857140, osc: 92.857142857142900 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 100.000000000000000, down: 0.000000000000000, osc: 100.000000000000000 },
            ExpectedRow { up: 92.857142857142900, down: 0.000000000000000, osc: 92.857142857142900 },
            ExpectedRow { up: 85.714285714285700, down: 14.285714285714300, osc: 71.428571428571400 },
            ExpectedRow { up: 78.571428571428600, down: 7.142857142857140, osc: 71.428571428571400 },
            ExpectedRow { up: 71.428571428571400, down: 0.000000000000000, osc: 71.428571428571400 },
            ExpectedRow { up: 64.285714285714300, down: 14.285714285714300, osc: 50.000000000000000 },
            ExpectedRow { up: 57.142857142857100, down: 7.142857142857140, osc: 50.000000000000000 },
            ExpectedRow { up: 50.000000000000000, down: 0.000000000000000, osc: 50.000000000000000 },
            ExpectedRow { up: 42.857142857142900, down: 0.000000000000000, osc: 42.857142857142900 },
            ExpectedRow { up: 35.714285714285700, down: 0.000000000000000, osc: 35.714285714285700 },
            ExpectedRow { up: 28.571428571428600, down: 57.142857142857100, osc: -28.571428571428600 },
            ExpectedRow { up: 21.428571428571400, down: 50.000000000000000, osc: -28.571428571428600 },
            ExpectedRow { up: 14.285714285714300, down: 42.857142857142900, osc: -28.571428571428600 },
            ExpectedRow { up: 7.142857142857140, down: 35.714285714285700, osc: -28.571428571428600 },
            ExpectedRow { up: 0.000000000000000, down: 28.571428571428600, osc: -28.571428571428600 },
            ExpectedRow { up: 7.142857142857140, down: 21.428571428571400, osc: -14.285714285714300 },
        ]
    }

    #[test]
    fn test_aroon_length14_full_data() {
        let tolerance = 1e-6;
        let high = test_input_high();
        let low = test_input_low();
        let expected = test_expected();

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
        let high = test_input_high();
        let low = test_input_low();

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
