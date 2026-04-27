use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the triangular moving average indicator.
pub struct TriangularMovingAverageParams {
    /// The length (number of time periods) of the moving window.
    /// Must be greater than 1.
    pub length: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for TriangularMovingAverageParams {
    fn default() -> Self {
        Self {
            length: 20,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the triangular moving average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum TriangularMovingAverageOutput {
    /// The scalar value of the moving average.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the triangular moving average (TRIMA).
///
/// The TRIMA puts more weight on the data in the middle of the window,
/// equivalent to computing an SMA of an SMA.
///
/// Uses an optimised incremental algorithm with four adjustments per step.
pub struct TriangularMovingAverage {
    line: LineIndicator,
    factor: f64,
    numerator: f64,
    numerator_sub: f64,
    numerator_add: f64,
    window: Vec<f64>,
    window_length: usize,
    window_length_half: usize,
    window_count: usize,
    is_odd: bool,
    primed: bool,
}

impl TriangularMovingAverage {
    /// Creates a new TriangularMovingAverage from the given parameters.
    pub fn new(params: &TriangularMovingAverageParams) -> Result<Self, String> {
        if params.length < 2 {
            return Err("invalid triangular moving average parameters: length should be greater than 1".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let length = params.length;
        let length_half = length >> 1;
        let l = 1 + length_half;
        let is_odd = length % 2 == 1;

        let (factor, window_length_half) = if is_odd {
            // Odd: 1+2+...+(l)+...+2+1 = l*l where l = (length+1)/2 = length_half+1.
            (1.0 / (l * l) as f64, length_half)
        } else {
            // Even: 1+2+...+l+l+...+2+1 = length_half * l.
            (1.0 / (length_half * l) as f64, length_half - 1)
        };

        let mnemonic = format!("trima({}{})", length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Triangular moving average {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            factor,
            numerator: 0.0,
            numerator_sub: 0.0,
            numerator_add: 0.0,
            window: vec![0.0; length],
            window_length: length,
            window_length_half,
            window_count: 0,
            is_odd,
            primed: false,
        })
    }

    /// Core update logic. Returns the TRIMA value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        let temp = sample;

        if self.primed {
            self.numerator -= self.numerator_sub;
            self.numerator_sub -= self.window[0];

            let j = self.window_length - 1;
            for i in 0..j {
                self.window[i] = self.window[i + 1];
            }

            self.window[j] = temp;
            let mid = self.window[self.window_length_half];
            self.numerator_sub += mid;

            if self.is_odd {
                self.numerator += self.numerator_add;
                self.numerator_add -= mid;
            } else {
                self.numerator_add -= mid;
                self.numerator += self.numerator_add;
            }

            self.numerator_add += sample;
            self.numerator += sample;
        } else {
            self.window[self.window_count] = temp;
            self.window_count += 1;

            if self.window_length > self.window_count {
                return f64::NAN;
            }

            // Initialise numerator_sub from the middle going left.
            let half = self.window_length_half;
            for i in (0..=half).rev() {
                self.numerator_sub += self.window[i];
                self.numerator += self.numerator_sub;
            }

            // Initialise numerator_add from the middle+1 going right.
            for i in (half + 1)..self.window_length {
                self.numerator_add += self.window[i];
                self.numerator += self.numerator_add;
            }

            self.primed = true;
        }

        self.numerator * self.factor
    }
}

impl Indicator for TriangularMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::TriangularMovingAverage,
            &self.line.mnemonic,
            &self.line.description,
            &[OutputText {
                mnemonic: self.line.mnemonic.clone(),
                description: self.line.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let sample_value = (self.line.bar_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let sample_value = (self.line.quote_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let sample_value = (self.line.trade_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::entities::bar_component::BarComponent;
    use crate::indicators::core::outputs::shape::Shape;

    fn test_input() -> Vec<f64> {
        vec![
            91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
            96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
            88.375000, 87.625000, 84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000,
            85.250000, 87.125000, 85.815000, 88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000,
            83.375000, 85.500000, 89.190000, 89.440000, 91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000,
            89.030000, 88.815000, 84.280000, 83.500000, 82.690000, 84.750000, 85.655000, 86.190000, 88.940000, 89.280000,
            88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000, 93.155000, 91.720000, 90.000000, 89.690000,
            88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000, 104.940000, 106.000000, 102.500000,
            102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000,
            112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000,
            110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
            116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000,
            124.750000, 123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000,
            132.250000, 131.000000, 132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000,
            136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000,
            125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000, 122.190000, 119.310000,
            123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000, 124.440000,
            122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
            132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000,
            130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000,
            117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
            107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
            94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000,
            95.000000, 95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000,
            105.000000, 104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000,
            113.370000, 109.000000, 109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000,
            108.000000, 108.620000, 109.750000, 109.810000, 109.000000, 108.750000, 107.870000,
        ]
    }

    fn expected_xls12() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN,
            93.5329761904762, 93.6096428571428, 93.5933333333333, 93.6425000000000, 93.7965476190476, 93.8471428571429,
            93.6536904761905, 93.2340476190476, 92.6722619047619, 92.1164285714286, 91.4207142857143, 90.6126190476190,
            89.8194047619047, 89.0209523809524, 88.1838095238095, 87.2529761904762, 86.4233333333333, 85.7552380952381,
            85.2686904761905, 84.9748809523809, 85.0114285714286, 85.2830952380952, 85.6417857142857, 86.0116666666667,
            86.3584523809524, 86.6651190476190, 86.8765476190476, 86.9123809523810, 86.7941666666667, 86.5613095238095,
            86.2458333333333, 85.9720238095238, 85.7696428571429, 85.7852380952381, 86.0032142857143, 86.5345238095238,
            87.2704761904762, 88.0822619047619, 88.8627380952381, 89.4853571428571, 89.8975000000000, 89.9754761904762,
            89.7304761904762, 89.2042857142857, 88.4980952380952, 87.6863095238095, 86.8611904761905, 86.1930952380952,
            85.8330952380952, 85.7453571428571, 85.9447619047619, 86.4314285714286, 87.1248809523809, 87.9834523809524,
            88.8315476190476, 89.6498809523810, 90.4035714285714, 91.0210714285714, 91.4451190476190, 91.6385714285714,
            91.5315476190476, 91.0911904761905, 90.3800000000000, 89.4954761904762, 88.8378571428571, 88.4852380952381,
            88.7070238095238, 89.6653571428571, 91.2761904761905, 93.4420238095238, 95.8794047619048, 98.2855952380953,
            100.4551190476190, 102.1559523809520, 103.3686904761900, 104.3098809523810, 104.9714285714290, 105.5622619047620,
            106.1650000000000, 107.1457142857140, 108.4820238095240, 110.0088095238100, 111.6240476190480, 113.3040476190480,
            114.9677380952380, 116.2847619047620, 117.0140476190480, 117.1920238095240, 117.1021428571430, 116.7295238095240,
            116.1692857142860, 115.4452380952380, 114.9517857142860, 114.6986904761900, 114.5891666666670, 114.6135714285710,
            114.6989285714290, 114.9138095238100, 115.2403571428570, 115.5548809523810, 115.8016666666670, 115.9888095238100,
            116.1657142857140, 116.4038095238090, 116.6538095238100, 117.1166666666670, 117.7342857142860, 118.5321428571430,
            119.4847619047620, 120.4102380952380, 121.3028571428570, 122.0614285714290, 122.7114285714290, 123.3659523809520,
            124.0828571428570, 124.9428571428570, 125.9771428571430, 127.1916666666670, 128.6028571428570, 130.0361904761900,
            131.4116666666670, 132.7052380952380, 133.8945238095240, 134.8933333333330, 135.6033333333330, 135.8921428571430,
            135.8073809523810, 135.2700000000000, 134.3100000000000, 132.9511904761900, 131.3392857142860, 129.7959523809520,
            128.3938095238100, 127.2464285714290, 126.3566666666670, 125.6542857142860, 125.0828571428570, 124.5873809523810,
            124.0442857142860, 123.5042857142860, 122.8509523809520, 122.3523809523810, 122.0026190476190, 121.8416666666670,
            121.8964285714290, 122.1459523809520, 122.5873809523810, 123.0900000000000, 123.5138095238100, 123.9007142857140,
            124.1554761904760, 124.1721428571430, 124.0164285714290, 123.7773809523810, 123.5814285714290, 123.3733333333330,
            123.2647619047620, 123.3673809523810, 123.7569047619050, 124.3590476190480, 125.1159523809520, 126.0811904761900,
            127.2280952380950, 128.4050000000000, 129.6045238095240, 130.6616666666670, 131.5104761904760, 131.9559523809520,
            132.0428571428570, 131.8200000000000, 131.2488095238090, 130.3350000000000, 129.3035714285710, 128.2335714285710,
            127.2290476190480, 126.1723809523810, 125.1411904761900, 124.2021428571430, 123.3776190476190, 122.6483333333330,
            121.8728571428570, 121.1673809523810, 120.4514285714290, 119.7519047619050, 118.9185714285710, 117.8040476190480,
            116.4230952380950, 114.9423809523810, 113.3947619047620, 111.8559523809520, 110.3290476190480, 108.7023809523810,
            107.1680952380950, 105.5907142857140, 103.9419047619050, 102.1116666666670, 100.1640476190480, 98.4604761904762,
            97.1585714285714, 96.1900000000000, 95.5278571428571, 95.1052380952381, 94.9071428571429, 94.8935714285714,
            94.6328571428571, 94.3573809523810, 94.0745238095238, 93.9211904761905, 93.8928571428571, 93.9923809523810,
            94.1976190476191, 94.5011904761905, 94.9654761904762, 95.7004761904762, 96.6185714285715, 97.6811904761905,
            98.9954761904762, 100.4540476190480, 101.8678571428570, 102.9628571428570, 103.7533333333330, 104.4335714285710,
            105.1404761904760, 105.8754761904760, 106.8254761904760, 108.0333333333330, 109.4359523809520, 110.8057142857140,
            111.8392857142860, 112.3819047619050, 112.4121428571430, 111.9997619047620, 111.3552380952380, 110.6319047619050,
            109.9304761904760, 109.4283333333330, 109.1685714285710, 109.1207142857140, 109.1483333333330, 109.1385714285710,
            109.1157142857140,
        ]
    }

    fn create_trima(length: usize) -> TriangularMovingAverage {
        TriangularMovingAverage::new(&TriangularMovingAverageParams { length, ..Default::default() }).unwrap()
    }

    #[test]
    fn test_update_length_9() {
        let mut trima = create_trima(9);
        let input = test_input();

        for i in 0..8 {
            assert!(trima.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let act = trima.update(input[8]);
        assert!((93.8176 - act).abs() < 1e-4, "[8] expected 93.8176, got {}", act);

        for i in 9..input.len() - 1 {
            trima.update(input[i]);
        }

        let act = trima.update(input[251]);
        assert!((109.1312 - act).abs() < 1e-4, "[251] expected 109.1312, got {}", act);

        assert!(trima.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_10() {
        let mut trima = create_trima(10);
        let input = test_input();

        for i in 0..9 {
            assert!(trima.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let act = trima.update(input[9]);
        assert!((93.6043 - act).abs() < 1e-4, "[9] expected 93.6043, got {}", act);

        let act = trima.update(input[10]);
        assert!((93.4252 - act).abs() < 1e-4, "[10] expected 93.4252, got {}", act);

        for i in 11..250 {
            trima.update(input[i]);
        }

        let act = trima.update(input[250]);
        assert!((109.1850 - act).abs() < 1e-4, "[250] expected 109.1850, got {}", act);

        let act = trima.update(input[251]);
        assert!((109.1407 - act).abs() < 1e-4, "[251] expected 109.1407, got {}", act);

        assert!(trima.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_12() {
        let mut trima = create_trima(12);
        let input = test_input();

        for i in 0..10 {
            assert!(trima.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        // index 10 is still NaN for length 12 (need 11 NaN values: indices 0..=10)
        assert!(trima.update(input[10]).is_nan(), "[10] expected NaN");

        let act = trima.update(input[11]);
        assert!((93.5329 - act).abs() < 1e-4, "[11] expected 93.5329, got {}", act);

        for i in 12..251 {
            trima.update(input[i]);
        }

        let act = trima.update(input[251]);
        assert!((109.1157 - act).abs() < 1e-4, "[251] expected 109.1157, got {}", act);

        assert!(trima.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_xls12() {
        let mut trima = create_trima(12);
        let input = test_input();
        let expected = expected_xls12();

        for i in 0..11 {
            assert!(trima.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in 11..input.len() {
            let act = trima.update(input[i]);
            assert!(
                (expected[i] - act).abs() < 1e-12,
                "[{}] expected {}, got {}", i, expected[i], act
            );
        }
    }

    #[test]
    fn test_update_entity() {
        let length = 12;
        let input = test_input();
        let inp = 97.250000; // input[11]
        let exp = 93.5329761904762;
        let time = 1617235200_i64;

        // scalar
        let mut trima = create_trima(length);
        for i in 0..11 {
            trima.update(input[i]);
        }
        let out = trima.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!((exp - s.value).abs() < 1e-12);

        // bar
        let mut trima = create_trima(length);
        for i in 0..11 {
            trima.update(input[i]);
        }
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = trima.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((exp - s.value).abs() < 1e-12);

        // quote
        let mut trima = create_trima(length);
        for i in 0..11 {
            trima.update(input[i]);
        }
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = trima.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((exp - s.value).abs() < 1e-12);

        // trade
        let mut trima = create_trima(length);
        for i in 0..11 {
            trima.update(input[i]);
        }
        let trade = Trade::new(time, inp, 0.0);
        let out = trima.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((exp - s.value).abs() < 1e-12);
    }

    #[test]
    fn test_is_primed() {
        let input = test_input();

        // length 9: primed after 9 samples (index 8)
        let mut trima = create_trima(9);
        assert!(!trima.is_primed());
        for i in 0..8 {
            trima.update(input[i]);
            assert!(!trima.is_primed(), "[{}] should not be primed", i);
        }
        for i in 8..input.len() {
            trima.update(input[i]);
            assert!(trima.is_primed(), "[{}] should be primed", i);
        }

        // length 12: primed after 12 samples (index 11)
        let mut trima = create_trima(12);
        assert!(!trima.is_primed());
        for i in 0..11 {
            trima.update(input[i]);
            assert!(!trima.is_primed(), "[{}] should not be primed", i);
        }
        for i in 11..input.len() {
            trima.update(input[i]);
            assert!(trima.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_metadata() {
        let trima = create_trima(5);
        let m = trima.metadata();
        assert_eq!(m.identifier, Identifier::TriangularMovingAverage);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, TriangularMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "trima(5)");
        assert_eq!(m.outputs[0].description, "Triangular moving average trima(5)");
    }

    #[test]
    fn test_new_invalid_length() {
        let r = TriangularMovingAverage::new(&TriangularMovingAverageParams { length: 1, ..Default::default() });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid triangular moving average parameters: length should be greater than 1");

        let r = TriangularMovingAverage::new(&TriangularMovingAverageParams { length: 0, ..Default::default() });
        assert!(r.is_err());
    }

    #[test]
    fn test_mnemonic_components() {
        // all defaults -> no component suffix
        let trima = create_trima(5);
        assert_eq!(trima.line.mnemonic, "trima(5)");

        // bar component set
        let trima = TriangularMovingAverage::new(&TriangularMovingAverageParams {
            length: 5, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(trima.line.mnemonic, "trima(5, hl/2)");

        // bar and trade set
        let trima = TriangularMovingAverage::new(&TriangularMovingAverageParams {
            length: 5,
            bar_component: Some(BarComponent::High),
            quote_component: None,
            trade_component: Some(TradeComponent::Volume),
        }).unwrap();
        assert_eq!(trima.line.mnemonic, "trima(5, h, v)");
    }
}
