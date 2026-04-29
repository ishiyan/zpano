use std::any::Any;
use std::f64::consts::PI;

use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{
    component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT,
};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{
    component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT,
};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
use crate::indicators::core::outputs::heatmap::Heatmap;

// ---------------------------------------------------------------------------
// Output enum
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DiscreteFourierTransformSpectrumOutput {
    Value = 1,
}

impl DiscreteFourierTransformSpectrumOutput {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Value => "value",
        }
    }

    pub fn is_known(&self) -> bool {
        matches!(self, Self::Value)
    }
}

impl std::fmt::Display for DiscreteFourierTransformSpectrumOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

pub struct DiscreteFourierTransformSpectrumParams {
    pub length: usize,
    pub min_period: f64,
    pub max_period: f64,
    pub spectrum_resolution: usize,
    pub disable_spectral_dilation_compensation: bool,
    pub disable_automatic_gain_control: bool,
    pub automatic_gain_control_decay_factor: f64,
    pub fixed_normalization: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for DiscreteFourierTransformSpectrumParams {
    fn default() -> Self {
        Self {
            length: 0,
            min_period: 0.0,
            max_period: 0.0,
            spectrum_resolution: 0,
            disable_spectral_dilation_compensation: false,
            disable_automatic_gain_control: false,
            automatic_gain_control_decay_factor: 0.0,
            fixed_normalization: false,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Estimator (internal)
// ---------------------------------------------------------------------------

struct Estimator {
    length: usize,
    length_spectrum: usize,
    max_omega_length: usize,
    is_spectral_dilation_compensation: bool,
    is_automatic_gain_control: bool,
    automatic_gain_control_decay_factor: f64,

    input_series: Vec<f64>,
    input_series_minus_mean: Vec<f64>,
    spectrum: Vec<f64>,
    period: Vec<f64>,

    frequency_sin_omega: Vec<Vec<f64>>,
    frequency_cos_omega: Vec<Vec<f64>>,

    spectrum_min: f64,
    spectrum_max: f64,
    previous_spectrum_max: f64,
}

impl Estimator {
    fn new(
        length: usize,
        min_period: f64,
        max_period: f64,
        spectrum_resolution: usize,
        is_spectral_dilation_compensation: bool,
        is_automatic_gain_control: bool,
        automatic_gain_control_decay_factor: f64,
    ) -> Self {
        let two_pi = 2.0 * PI;
        let length_spectrum =
            ((max_period - min_period) * spectrum_resolution as f64) as usize + 1;
        let max_omega_length = length;
        let res = spectrum_resolution as f64;

        let mut period = vec![0.0; length_spectrum];
        let mut frequency_sin_omega = Vec::with_capacity(length_spectrum);
        let mut frequency_cos_omega = Vec::with_capacity(length_spectrum);

        for i in 0..length_spectrum {
            let p = max_period - i as f64 / res;
            period[i] = p;
            let theta = two_pi / p;

            let mut sin_row = vec![0.0; max_omega_length];
            let mut cos_row = vec![0.0; max_omega_length];

            for j in 0..max_omega_length {
                let omega = j as f64 * theta;
                sin_row[j] = omega.sin();
                cos_row[j] = omega.cos();
            }

            frequency_sin_omega.push(sin_row);
            frequency_cos_omega.push(cos_row);
        }

        Self {
            length,
            length_spectrum,
            max_omega_length,
            is_spectral_dilation_compensation,
            is_automatic_gain_control,
            automatic_gain_control_decay_factor,
            input_series: vec![0.0; length],
            input_series_minus_mean: vec![0.0; length],
            spectrum: vec![0.0; length_spectrum],
            period,
            frequency_sin_omega,
            frequency_cos_omega,
            spectrum_min: 0.0,
            spectrum_max: 0.0,
            previous_spectrum_max: 0.0,
        }
    }

    fn calculate(&mut self) {
        let mut mean = 0.0;
        for i in 0..self.length {
            mean += self.input_series[i];
        }
        mean /= self.length as f64;

        for i in 0..self.length {
            self.input_series_minus_mean[i] = self.input_series[i] - mean;
        }

        self.spectrum_min = f64::MAX;
        if self.is_automatic_gain_control {
            self.spectrum_max =
                self.automatic_gain_control_decay_factor * self.previous_spectrum_max;
        } else {
            self.spectrum_max = f64::MIN;
        }

        for i in 0..self.length_spectrum {
            let sin_row = &self.frequency_sin_omega[i];
            let cos_row = &self.frequency_cos_omega[i];

            let mut sum_sin = 0.0;
            let mut sum_cos = 0.0;

            for j in 0..self.max_omega_length {
                let sample = self.input_series_minus_mean[j];
                sum_sin += sample * sin_row[j];
                sum_cos += sample * cos_row[j];
            }

            let mut s = sum_sin * sum_sin + sum_cos * sum_cos;
            if self.is_spectral_dilation_compensation {
                s /= self.period[i];
            }

            self.spectrum[i] = s;

            if self.spectrum_max < s {
                self.spectrum_max = s;
            }

            if self.spectrum_min > s {
                self.spectrum_min = s;
            }
        }

        self.previous_spectrum_max = self.spectrum_max;
    }
}

// ---------------------------------------------------------------------------
// DiscreteFourierTransformSpectrum indicator
// ---------------------------------------------------------------------------

pub struct DiscreteFourierTransformSpectrum {
    mnemonic: String,
    description: String,
    estimator: Estimator,
    window_count: usize,
    last_index: usize,
    primed: bool,
    floating_normalization: bool,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl DiscreteFourierTransformSpectrum {
    pub fn default_params() -> Result<Self, String> {
        Self::new(&DiscreteFourierTransformSpectrumParams::default())
    }

    pub fn new(p: &DiscreteFourierTransformSpectrumParams) -> Result<Self, String> {
        let invalid = "invalid discrete Fourier transform spectrum parameters";

        let def_length: usize = 48;
        let def_min_period: f64 = 10.0;
        let def_max_period: f64 = 48.0;
        let def_spectrum_resolution: usize = 1;
        let def_agc_decay: f64 = 0.995;
        let agc_decay_epsilon: f64 = 1e-12;

        let length = if p.length == 0 { def_length } else { p.length };
        let min_period = if p.min_period == 0.0 { def_min_period } else { p.min_period };
        let max_period = if p.max_period == 0.0 { def_max_period } else { p.max_period };
        let spectrum_resolution = if p.spectrum_resolution == 0 {
            def_spectrum_resolution
        } else {
            p.spectrum_resolution
        };
        let agc_decay = if p.automatic_gain_control_decay_factor == 0.0 {
            def_agc_decay
        } else {
            p.automatic_gain_control_decay_factor
        };

        let sdc_on = !p.disable_spectral_dilation_compensation;
        let agc_on = !p.disable_automatic_gain_control;
        let floating_norm = !p.fixed_normalization;

        if length < 2 {
            return Err(format!("{}: Length should be >= 2", invalid));
        }
        if min_period < 2.0 {
            return Err(format!("{}: MinPeriod should be >= 2", invalid));
        }
        if max_period <= min_period {
            return Err(format!("{}: MaxPeriod should be > MinPeriod", invalid));
        }
        if max_period > 2.0 * length as f64 {
            return Err(format!("{}: MaxPeriod should be <= 2 * Length", invalid));
        }
        if spectrum_resolution < 1 {
            return Err(format!("{}: SpectrumResolution should be >= 1", invalid));
        }
        if agc_on && (agc_decay <= 0.0 || agc_decay >= 1.0) {
            return Err(format!(
                "{}: AutomaticGainControlDecayFactor should be in (0, 1)",
                invalid
            ));
        }

        let bc = p.bar_component.unwrap_or(BarComponent::Median);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let component_mnemonic = component_triple_mnemonic(bc, qc, tc);

        let flags = build_flag_tags(sdc_on, agc_on, agc_decay, floating_norm, def_agc_decay, agc_decay_epsilon);

        let mnemonic = format!(
            "dftps({}, {}, {}, {}{}{})",
            length, min_period, max_period, spectrum_resolution, flags, component_mnemonic
        );
        let description = format!("Discrete Fourier transform spectrum {}", mnemonic);

        let estimator = Estimator::new(
            length,
            min_period,
            max_period,
            spectrum_resolution,
            sdc_on,
            agc_on,
            agc_decay,
        );

        Ok(Self {
            mnemonic,
            description,
            estimator,
            window_count: 0,
            last_index: length - 1,
            primed: false,
            floating_normalization: floating_norm,
            min_parameter_value: min_period,
            max_parameter_value: max_period,
            parameter_resolution: spectrum_resolution as f64,
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
        })
    }

    pub fn mnemonic(&self) -> &str {
        &self.mnemonic
    }

    pub fn update(&mut self, sample: f64, time: i64) -> Heatmap {
        if sample.is_nan() {
            return Heatmap::empty(
                time,
                self.min_parameter_value,
                self.max_parameter_value,
                self.parameter_resolution,
            );
        }

        let window = &mut self.estimator.input_series;

        if self.primed {
            window.copy_within(1.., 0);
            window[self.last_index] = sample;
        } else {
            window[self.window_count] = sample;
            self.window_count += 1;
            if self.window_count == self.estimator.length {
                self.primed = true;
            }
        }

        if !self.primed {
            return Heatmap::empty(
                time,
                self.min_parameter_value,
                self.max_parameter_value,
                self.parameter_resolution,
            );
        }

        self.estimator.calculate();

        let length_spectrum = self.estimator.length_spectrum;

        let min_ref = if self.floating_normalization {
            self.estimator.spectrum_min
        } else {
            0.0
        };

        let max_ref = self.estimator.spectrum_max;
        let spectrum_range = max_ref - min_ref;

        let mut values = vec![0.0; length_spectrum];
        let mut value_min = f64::INFINITY;
        let mut value_max = f64::NEG_INFINITY;

        for i in 0..length_spectrum {
            let v =
                (self.estimator.spectrum[length_spectrum - 1 - i] - min_ref) / spectrum_range;
            values[i] = v;
            if v < value_min {
                value_min = v;
            }
            if v > value_max {
                value_max = v;
            }
        }

        Heatmap::new(
            time,
            self.min_parameter_value,
            self.max_parameter_value,
            self.parameter_resolution,
            value_min,
            value_max,
            values,
        )
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let heatmap = self.update(sample, time);
        vec![Box::new(heatmap) as Box<dyn Any>]
    }
}

fn build_flag_tags(
    sdc_on: bool,
    agc_on: bool,
    agc_decay: f64,
    floating_norm: bool,
    def_agc: f64,
    eps: f64,
) -> String {
    let mut s = String::new();

    if !sdc_on {
        s.push_str(", no-sdc");
    }
    if !agc_on {
        s.push_str(", no-agc");
    }
    if agc_on && (agc_decay - def_agc).abs() > eps {
        s.push_str(&format!(", agc={}", agc_decay));
    }
    if !floating_norm {
        s.push_str(", no-fn");
    }

    s
}

impl Indicator for DiscreteFourierTransformSpectrum {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::DiscreteFourierTransformSpectrum,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        self.update_entity(sample.time, v)
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::core::outputs::shape::Shape;

    fn test_dfts_time() -> i64 {
        0
    }

    fn test_dfts_input() -> Vec<f64> {
        vec![
            92.0000, 93.1725, 95.3125, 94.8450, 94.4075, 94.1100, 93.5000, 91.7350, 90.9550,
            91.6875, 94.5000, 97.9700, 97.5775, 90.7825, 89.0325, 92.0950, 91.1550, 89.7175,
            90.6100, 91.0000, 88.9225, 87.5150, 86.4375, 83.8900, 83.0025, 82.8125, 82.8450,
            86.7350, 86.8600, 87.5475, 85.7800, 86.1725, 86.4375, 87.2500, 88.9375, 88.2050,
            85.8125, 84.5950, 83.6575, 84.4550, 83.5000, 86.7825, 88.1725, 89.2650, 90.8600,
            90.7825, 91.8600, 90.3600, 89.8600, 90.9225, 89.5000, 87.6725, 86.5000, 84.2825,
            82.9075, 84.2500, 85.6875, 86.6100, 88.2825, 89.5325, 89.5000, 88.0950, 90.6250,
            92.2350, 91.6725, 92.5925, 93.0150, 91.1725, 90.9850, 90.3775, 88.2500, 86.9075,
            84.0925, 83.1875, 84.2525, 97.8600, 99.8750, 103.2650, 105.9375, 103.5000,
            103.1100, 103.6100, 104.6400, 106.8150, 104.9525, 105.5000, 107.1400, 109.7350,
            109.8450, 110.9850, 120.0000, 119.8750, 117.9075, 119.4075, 117.9525, 117.2200,
            115.6425, 113.1100, 111.7500, 114.5175, 114.7450, 115.4700, 112.5300, 112.0300,
            113.4350, 114.2200, 119.5950, 117.9650, 118.7150, 115.0300, 114.5300, 115.0000,
            116.5300, 120.1850, 120.5000, 120.5950, 124.1850, 125.3750, 122.9700, 123.0000,
            124.4350, 123.4400, 124.0300, 128.1850, 129.6550, 130.8750, 132.3450, 132.0650,
            133.8150, 135.6600, 137.0350, 137.4700, 137.3450, 136.3150, 136.4400, 136.2850,
            129.0950, 128.3100, 126.0000, 124.0300, 123.9350, 125.0300, 127.2500, 125.6200,
            125.5300, 123.9050, 120.6550, 119.9650, 120.7800, 124.0000, 122.7800, 120.7200,
            121.7800, 122.4050, 123.2500, 126.1850, 127.5600, 126.5650, 123.0600, 122.7150,
            123.5900, 122.3100, 122.4650, 123.9650, 123.9700, 124.1550, 124.4350, 127.0000,
            125.5000, 128.8750, 130.5350, 132.3150, 134.0650, 136.0350, 133.7800, 132.7500,
            133.4700, 130.9700, 127.5950, 128.4400, 127.9400, 125.8100, 124.6250, 122.7200,
            124.0900, 123.2200, 121.4050, 120.9350, 118.2800, 118.3750, 121.1550, 120.9050,
            117.1250, 113.0600, 114.9050, 112.4350, 107.9350, 105.9700, 106.3700, 106.8450,
            106.9700, 110.0300, 91.0000, 93.5600, 93.6200, 95.3100, 94.1850, 94.7800, 97.6250,
            97.5900, 95.2500, 94.7200, 92.2200, 91.5650, 92.2200, 93.8100, 95.5900, 96.1850,
            94.6250, 95.1200, 94.0000, 93.7450, 95.9050, 101.7450, 106.4400, 107.9350,
            103.4050, 105.0600, 104.1550, 103.3100, 103.3450, 104.8400, 110.4050, 114.5000,
            117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300, 106.2200, 107.7200,
            109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
            109.5300, 108.0600,
        ]
    }

    const TOLERANCE: f64 = 1e-12;
    const MIN_MAX_TOL: f64 = 1e-10;

    struct SpotValue {
        i: usize,
        v: f64,
    }

    struct DftsSnap {
        input_index: usize,
        value_min: f64,
        value_max: f64,
        spots: Vec<SpotValue>,
    }

    fn dfts_snapshots() -> Vec<DftsSnap> {
        vec![
            DftsSnap {
                input_index: 47,
                value_min: 0.0,
                value_max: 1.0,
                spots: vec![
                    SpotValue { i: 0, v: 0.092590275198666 },
                    SpotValue { i: 9, v: 0.027548278511410 },
                    SpotValue { i: 19, v: 0.157582436454944 },
                    SpotValue { i: 28, v: 0.818873586056752 },
                    SpotValue { i: 38, v: 0.994469163657834 },
                ],
            },
            DftsSnap {
                input_index: 60,
                value_min: 0.0,
                value_max: 1.0,
                spots: vec![
                    SpotValue { i: 0, v: 0.051096698360059 },
                    SpotValue { i: 9, v: 0.058377239289306 },
                    SpotValue { i: 19, v: 0.248187339190831 },
                    SpotValue { i: 28, v: 0.136096222727665 },
                    SpotValue { i: 38, v: 0.059069344951534 },
                ],
            },
            DftsSnap {
                input_index: 100,
                value_min: 0.0,
                value_max: 1.0,
                spots: vec![
                    SpotValue { i: 0, v: 0.150638336509665 },
                    SpotValue { i: 9, v: 0.046435987728045 },
                    SpotValue { i: 19, v: 0.103832850895319 },
                    SpotValue { i: 28, v: 0.402801132284104 },
                    SpotValue { i: 38, v: 1.000000000000000 },
                ],
            },
            DftsSnap {
                input_index: 150,
                value_min: 0.0,
                value_max: 0.5272269971142493,
                spots: vec![
                    SpotValue { i: 0, v: 0.000000000000000 },
                    SpotValue { i: 9, v: 0.091857989427651 },
                    SpotValue { i: 19, v: 0.219395988856534 },
                    SpotValue { i: 28, v: 0.516960894560452 },
                    SpotValue { i: 38, v: 0.468682020733700 },
                ],
            },
            DftsSnap {
                input_index: 200,
                value_min: 0.0,
                value_max: 0.6015223942655807,
                spots: vec![
                    SpotValue { i: 0, v: 0.107853213261092 },
                    SpotValue { i: 9, v: 0.164118955219278 },
                    SpotValue { i: 19, v: 0.306440501928972 },
                    SpotValue { i: 28, v: 0.569768020155262 },
                    SpotValue { i: 38, v: 0.585690371992475 },
                ],
            },
        ]
    }

    #[test]
    fn test_dfts_update() {
        let input = test_dfts_input();
        let t0 = test_dfts_time();

        let mut x = DiscreteFourierTransformSpectrum::default_params().unwrap();
        let snaps = dfts_snapshots();
        let mut si = 0;

        for i in 0..input.len() {
            let h = x.update(input[i], t0 + i as i64);

            assert_eq!(h.parameter_first, 10.0, "[{}] parameter_first", i);
            assert_eq!(h.parameter_last, 48.0, "[{}] parameter_last", i);
            assert_eq!(h.parameter_resolution, 1.0, "[{}] parameter_resolution", i);

            if !x.is_primed() {
                assert!(h.is_empty(), "[{}] expected empty heatmap before priming", i);
                continue;
            }

            assert_eq!(h.values.len(), 39, "[{}] values len", i);

            if si < snaps.len() && snaps[si].input_index == i {
                let snap = &snaps[si];
                assert!(
                    (h.value_min - snap.value_min).abs() < MIN_MAX_TOL,
                    "[{}] ValueMin: expected {}, got {}",
                    i, snap.value_min, h.value_min
                );
                assert!(
                    (h.value_max - snap.value_max).abs() < MIN_MAX_TOL,
                    "[{}] ValueMax: expected {}, got {}",
                    i, snap.value_max, h.value_max
                );
                for sp in &snap.spots {
                    assert!(
                        (h.values[sp.i] - sp.v).abs() < TOLERANCE,
                        "[{}] Values[{}]: expected {}, got {}",
                        i, sp.i, sp.v, h.values[sp.i]
                    );
                }
                si += 1;
            }
        }

        assert_eq!(si, snaps.len(), "did not hit all snapshots");
    }

    #[test]
    fn test_dfts_primes_at_bar_47() {
        let mut x = DiscreteFourierTransformSpectrum::default_params().unwrap();
        assert!(!x.is_primed());

        let input = test_dfts_input();
        let t0 = test_dfts_time();
        let mut primed_at: Option<usize> = None;

        for i in 0..input.len() {
            x.update(input[i], t0 + i as i64);
            if x.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert_eq!(primed_at, Some(47));
    }

    #[test]
    fn test_dfts_nan_input() {
        let mut x = DiscreteFourierTransformSpectrum::default_params().unwrap();

        let h = x.update(f64::NAN, test_dfts_time());
        assert!(h.is_empty());
        assert!(!x.is_primed());
    }

    #[test]
    fn test_dfts_synthetic_sine() {
        let period: f64 = 16.0;
        let bars = 200;

        let mut x = DiscreteFourierTransformSpectrum::new(
            &DiscreteFourierTransformSpectrumParams {
                disable_spectral_dilation_compensation: true,
                disable_automatic_gain_control: true,
                fixed_normalization: true,
                ..Default::default()
            },
        )
        .unwrap();

        let t0 = test_dfts_time();
        let mut last = Heatmap::empty(0, 0.0, 0.0, 0.0);

        for i in 0..bars {
            let sample = 100.0 + (2.0 * PI * i as f64 / period).sin();
            last = x.update(sample, t0 + i as i64);
        }

        assert!(!last.is_empty());

        let mut peak_bin = 0;
        for i in 0..last.values.len() {
            if last.values[i] > last.values[peak_bin] {
                peak_bin = i;
            }
        }

        let expected_bin = (period - last.parameter_first) as usize;
        assert_eq!(
            peak_bin, expected_bin,
            "peak bin: expected {} (period {}), got {} (period {})",
            expected_bin, period, peak_bin, last.parameter_first + peak_bin as f64
        );
    }

    #[test]
    fn test_dfts_metadata() {
        let x = DiscreteFourierTransformSpectrum::default_params().unwrap();
        let md = x.metadata();

        let mn = "dftps(48, 10, 48, 1, hl/2)";

        assert_eq!(md.identifier, Identifier::DiscreteFourierTransformSpectrum);
        assert_eq!(md.mnemonic, mn);
        assert_eq!(md.description, format!("Discrete Fourier transform spectrum {}", mn));
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].kind, DiscreteFourierTransformSpectrumOutput::Value as i32);
        assert_eq!(md.outputs[0].shape, Shape::Heatmap);
        assert_eq!(md.outputs[0].mnemonic, mn);
    }

    #[test]
    fn test_dfts_mnemonic_flags() {
        let cases: Vec<(DiscreteFourierTransformSpectrumParams, &str)> = vec![
            (DiscreteFourierTransformSpectrumParams::default(), "dftps(48, 10, 48, 1, hl/2)"),
            (
                DiscreteFourierTransformSpectrumParams {
                    disable_spectral_dilation_compensation: true,
                    ..Default::default()
                },
                "dftps(48, 10, 48, 1, no-sdc, hl/2)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    disable_automatic_gain_control: true,
                    ..Default::default()
                },
                "dftps(48, 10, 48, 1, no-agc, hl/2)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    automatic_gain_control_decay_factor: 0.8,
                    ..Default::default()
                },
                "dftps(48, 10, 48, 1, agc=0.8, hl/2)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    fixed_normalization: true,
                    ..Default::default()
                },
                "dftps(48, 10, 48, 1, no-fn, hl/2)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    disable_spectral_dilation_compensation: true,
                    disable_automatic_gain_control: true,
                    fixed_normalization: true,
                    ..Default::default()
                },
                "dftps(48, 10, 48, 1, no-sdc, no-agc, no-fn, hl/2)",
            ),
        ];

        for (p, expected_mn) in &cases {
            let x = DiscreteFourierTransformSpectrum::new(p).unwrap();
            assert_eq!(x.mnemonic(), *expected_mn, "params produced wrong mnemonic");
        }
    }

    #[test]
    fn test_dfts_validation() {
        let cases: Vec<(DiscreteFourierTransformSpectrumParams, &str)> = vec![
            (
                DiscreteFourierTransformSpectrumParams {
                    length: 1, min_period: 10.0, max_period: 48.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid discrete Fourier transform spectrum parameters: Length should be >= 2",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    length: 48, min_period: 1.0, max_period: 48.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid discrete Fourier transform spectrum parameters: MinPeriod should be >= 2",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    length: 48, min_period: 10.0, max_period: 10.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid discrete Fourier transform spectrum parameters: MaxPeriod should be > MinPeriod",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    length: 10, min_period: 2.0, max_period: 48.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid discrete Fourier transform spectrum parameters: MaxPeriod should be <= 2 * Length",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    automatic_gain_control_decay_factor: -0.1,
                    ..Default::default()
                },
                "invalid discrete Fourier transform spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    automatic_gain_control_decay_factor: 1.0,
                    ..Default::default()
                },
                "invalid discrete Fourier transform spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
            ),
        ];

        for (p, expected_msg) in &cases {
            match DiscreteFourierTransformSpectrum::new(p) {
                Ok(_) => panic!("expected error for: {}", expected_msg),
                Err(e) => assert_eq!(e, *expected_msg),
            }
        }
    }

    #[test]
    fn test_dfts_update_entity() {
        let input = test_dfts_input();
        let t0 = test_dfts_time();
        let prime_count = 60;

        // Test update_scalar
        {
            let mut x = DiscreteFourierTransformSpectrum::default_params().unwrap();
            for i in 0..prime_count {
                x.update(input[i % input.len()], t0);
            }
            let s = Scalar { time: t0, value: 100.0 };
            let out = x.update_scalar(&s);
            assert_eq!(out.len(), 1);
            assert!(out[0].downcast_ref::<Heatmap>().is_some());
        }

        // Test update_bar
        {
            let mut x = DiscreteFourierTransformSpectrum::default_params().unwrap();
            for i in 0..prime_count {
                x.update(input[i % input.len()], t0);
            }
            let b = Bar {
                time: t0, open: 0.0, high: 100.0, low: 100.0, close: 100.0, volume: 0.0,
            };
            let out = x.update_bar(&b);
            assert_eq!(out.len(), 1);
            assert!(out[0].downcast_ref::<Heatmap>().is_some());
        }

        // Test update_quote
        {
            let mut x = DiscreteFourierTransformSpectrum::default_params().unwrap();
            for i in 0..prime_count {
                x.update(input[i % input.len()], t0);
            }
            let q = Quote {
                time: t0, bid_price: 100.0, ask_price: 100.0, bid_size: 0.0, ask_size: 0.0,
            };
            let out = x.update_quote(&q);
            assert_eq!(out.len(), 1);
            assert!(out[0].downcast_ref::<Heatmap>().is_some());
        }

        // Test update_trade
        {
            let mut x = DiscreteFourierTransformSpectrum::default_params().unwrap();
            for i in 0..prime_count {
                x.update(input[i % input.len()], t0);
            }
            let tr = Trade { time: t0, price: 100.0, volume: 0.0 };
            let out = x.update_trade(&tr);
            assert_eq!(out.len(), 1);
            assert!(out[0].downcast_ref::<Heatmap>().is_some());
        }
    }

    #[test]
    fn test_output_string() {
        assert_eq!(DiscreteFourierTransformSpectrumOutput::Value.as_str(), "value");
        assert!(DiscreteFourierTransformSpectrumOutput::Value.is_known());
    }
}
