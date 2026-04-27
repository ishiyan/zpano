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

/// Parameters for the True Range indicator.
/// True Range has no configurable parameters.
pub struct TrueRangeParams;

impl Default for TrueRangeParams {
    fn default() -> Self {
        Self
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the True Range indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum TrueRangeOutput {
    /// The scalar value of the true range.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

const TR_MNEMONIC: &str = "tr";
const TR_DESCRIPTION: &str = "True Range";

/// Welles Wilder's True Range indicator.
///
/// The True Range is defined as the largest of:
/// - the distance from today's high to today's low
/// - the distance from yesterday's close to today's high
/// - the distance from yesterday's close to today's low
///
/// The first update stores the close and returns NaN (not primed).
/// The indicator is primed from the second update onward.
pub struct TrueRange {
    previous_close: f64,
    value: f64,
    primed: bool,
}

impl TrueRange {
    /// Creates a new TrueRange indicator.
    pub fn new(_params: &TrueRangeParams) -> Result<Self, String> {
        Ok(Self {
            previous_close: f64::NAN,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Core update with close, high, low values.
    pub fn update(&mut self, close: f64, high: f64, low: f64) -> f64 {
        if close.is_nan() || high.is_nan() || low.is_nan() {
            return f64::NAN;
        }

        if !self.primed {
            if self.previous_close.is_nan() {
                self.previous_close = close;
                return f64::NAN;
            }
            self.primed = true;
        }

        let mut greatest = high - low;

        let temp = (high - self.previous_close).abs();
        if greatest < temp {
            greatest = temp;
        }

        let temp = (low - self.previous_close).abs();
        if greatest < temp {
            greatest = temp;
        }

        self.value = greatest;
        self.previous_close = close;

        self.value
    }

    /// Updates using a single sample value as substitute for high, low, and close.
    pub fn update_sample(&mut self, sample: f64) -> f64 {
        self.update(sample, sample, sample)
    }
}

impl Indicator for TrueRange {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::TrueRange,
            TR_MNEMONIC,
            TR_DESCRIPTION,
            &[OutputText {
                mnemonic: TR_MNEMONIC.to_string(),
                description: TR_DESCRIPTION.to_string(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let v = sample.value;
        let result = self.update(v, v, v);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let result = self.update(sample.close, sample.high, sample.low);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (sample.bid_price + sample.ask_price) / 2.0;
        let result = self.update(v, v, v);
        vec![Box::new(Scalar::new(sample.time, result))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = sample.price;
        let result = self.update(v, v, v);
        vec![Box::new(Scalar::new(sample.time, result))]
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

    fn test_input_close() -> Vec<f64> {
        vec![
            91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,
            96.125, 97.25, 98.5, 89.875, 91.0, 92.815, 89.155, 89.345, 91.625, 89.875,
            88.375, 87.625, 84.78, 83.0, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,
            85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,
            83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89.0, 91.0, 90.5,
            89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,
            88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90.0, 89.69,
            88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106.0, 102.5,
            102.405, 104.595, 106.125, 106.0, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,
            123.0, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,
            116.0, 116.0, 112.0, 113.75, 112.94, 116.0, 120.5, 116.62, 117.0, 115.25,
            114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,
            123.12, 122.56, 124.62, 129.25, 131.0, 132.25, 131.0, 132.81, 134.0, 137.38,
            137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129.0, 123.87, 124.81,
            123.0, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,
            122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,
            124.44, 122.0, 122.37, 122.94, 124.0, 123.19, 124.56, 127.25, 125.87, 128.86,
            132.0, 130.75, 134.75, 135.0, 132.38, 133.31, 131.94, 130.0, 125.37, 130.13,
            127.12, 125.19, 122.0, 125.0, 123.0, 123.5, 120.06, 121.0, 117.75, 119.87,
            122.0, 119.19, 116.37, 113.5, 114.25, 110.0, 105.06, 107.0, 107.87, 107.0,
            107.12, 107.0, 91.0, 93.94, 93.87, 95.5, 93.0, 94.94, 98.25, 96.75,
            94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97.0, 95.0, 95.87, 94.06,
            94.62, 93.75, 98.0, 103.94, 107.87, 106.06, 104.5, 105.0, 104.19, 103.06,
            103.42, 105.27, 111.87, 116.0, 116.62, 118.28, 113.37, 109.0, 109.7, 109.25,
            107.0, 109.19, 110.0, 109.2, 110.12, 108.0, 108.62, 109.75, 109.81, 109.0,
            108.75, 107.87,
        ]
    }

    fn test_expected_tr() -> Vec<f64> {
        vec![
            f64::NAN, 3.535, 2.125, 2.69, 3.185, 1.22, 3.0, 3.97, 3.31, 2.435,
            3.78, 3.5, 3.095, 9.685, 4.565, 2.31, 4.5, 1.875, 2.72, 2.5,
            2.845, 1.97, 3.625, 3.22, 2.875, 3.875, 3.19, 5.34, 3.655, 3.155,
            2.75, 2.155, 1.875, 3.44, 2.125, 3.28, 2.315, 3.565, 2.31, 2.03,
            1.94, 5.125, 3.97, 1.47, 3.16, 1.315, 2.22, 2.72, 2.59, 1.655,
            1.5, 2.56, 5.0, 1.935, 1.815, 2.56, 1.875, 2.66, 3.185, 2.185,
            2.5, 1.5, 3.47, 2.28, 4.285, 2.875, 2.03, 2.625, 2.03, 3.125,
            3.0, 3.81, 4.125, 3.375, 3.375, 13.435, 6.81, 5.5, 3.375, 4.25,
            2.78, 3.78, 2.97, 2.25, 2.595, 3.0, 4.125, 2.41, 2.19, 6.47,
            10.25, 5.0, 3.815, 1.815, 2.845, 1.94, 2.095, 4.47, 2.5, 7.72,
            5.505, 2.56, 4.81, 5.18, 3.75, 3.56, 5.69, 4.44, 4.31, 3.69,
            3.94, 2.0, 3.56, 5.63, 3.0, 1.69, 5.0, 4.75, 2.44, 2.88,
            3.19, 2.5, 2.75, 7.63, 3.31, 2.75, 3.07, 2.87, 3.37, 3.56,
            3.31, 3.44, 2.31, 3.63, 2.12, 5.19, 8.57, 2.88, 5.5, 2.82,
            2.37, 3.5, 2.88, 3.76, 2.32, 4.19, 4.93, 3.81, 6.44, 6.0,
            3.32, 3.56, 4.44, 2.81, 3.38, 4.5, 2.0, 6.13, 3.62, 3.31,
            3.12, 3.44, 3.19, 3.81, 2.44, 2.93, 3.13, 3.94, 3.0, 3.88,
            4.31, 3.25, 5.75, 3.31, 3.56, 1.62, 3.06, 2.82, 7.69, 5.51,
            3.88, 2.88, 6.37, 4.56, 5.44, 3.06, 3.5, 2.63, 5.56, 3.25,
            4.19, 3.69, 4.13, 5.31, 3.57, 4.87, 5.87, 2.56, 5.0, 4.31,
            3.06, 6.06, 18.0, 3.62, 3.0, 2.13, 2.75, 2.31, 4.06, 2.44,
            3.12, 2.44, 4.44, 2.75, 3.69, 3.38, 3.44, 2.63, 3.25, 2.5,
            2.0, 2.25, 4.69, 7.12, 4.5, 3.87, 4.25, 1.88, 1.63, 2.38,
            2.19, 2.94, 7.6, 4.63, 3.75, 5.5, 9.87, 5.81, 6.19, 3.32,
            4.75, 3.94, 2.44, 2.69, 2.06, 2.31, 2.44, 1.88, 1.69, 1.75,
            1.94, 2.88,
        ]
    }

    #[test]
    fn test_true_range_update() {
        let tolerance = 1e-3;
        let high = test_input_high();
        let low = test_input_low();
        let close = test_input_close();
        let expected = test_expected_tr();
        let mut tr = TrueRange::new(&TrueRangeParams).unwrap();

        for i in 0..close.len() {
            let act = tr.update(close[i], high[i], low[i]);

            if i == 0 {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
                continue;
            }

            assert!(!act.is_nan(), "[{}] expected {}, got NaN", i, expected[i]);
            assert!(
                (act - expected[i]).abs() < tolerance,
                "[{}] expected {}, got {}",
                i,
                expected[i],
                act
            );
        }
    }

    #[test]
    fn test_true_range_nan_passthrough() {
        let mut tr = TrueRange::new(&TrueRangeParams).unwrap();
        assert!(tr.update(f64::NAN, 1.0, 1.0).is_nan());
        assert!(tr.update(1.0, f64::NAN, 1.0).is_nan());
        assert!(tr.update(1.0, 1.0, f64::NAN).is_nan());
    }

    #[test]
    fn test_true_range_is_primed() {
        let high = test_input_high();
        let low = test_input_low();
        let close = test_input_close();
        let mut tr = TrueRange::new(&TrueRangeParams).unwrap();

        assert!(!tr.is_primed());
        tr.update(close[0], high[0], low[0]);
        assert!(!tr.is_primed());
        tr.update(close[1], high[1], low[1]);
        assert!(tr.is_primed());
        tr.update(close[2], high[2], low[2]);
        assert!(tr.is_primed());
    }

    #[test]
    fn test_true_range_update_sample() {
        let mut tr = TrueRange::new(&TrueRangeParams).unwrap();

        let v = tr.update_sample(100.0);
        assert!(v.is_nan());

        let v = tr.update_sample(105.0);
        assert!((v - 5.0).abs() < 1e-10);

        let v = tr.update_sample(102.0);
        assert!((v - 3.0).abs() < 1e-10);
    }

    #[test]
    fn test_true_range_update_bar() {
        let mut tr = TrueRange::new(&TrueRangeParams).unwrap();
        tr.update(100.0, 105.0, 95.0);

        let bar = Bar {
            time: 1_000_000,
            open: 0.0,
            high: 110.0,
            low: 98.0,
            close: 108.0,
            volume: 0.0,
        };
        let out = tr.update_bar(&bar);
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, 1_000_000);
        assert!(!s.value.is_nan());
    }

    #[test]
    fn test_true_range_metadata() {
        let tr = TrueRange::new(&TrueRangeParams).unwrap();
        let meta = tr.metadata();
        assert_eq!(meta.identifier, Identifier::TrueRange);
        assert_eq!(meta.mnemonic, "tr");
        assert_eq!(meta.description, "True Range");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].mnemonic, "tr");
        assert_eq!(meta.outputs[0].description, "True Range");
    }
}
