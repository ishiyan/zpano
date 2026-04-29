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

/// Describes the outputs of the Maximum Entropy Spectrum indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MaximumEntropySpectrumOutput {
    /// The maximum-entropy spectrum heatmap column.
    Value = 1,
}

impl MaximumEntropySpectrumOutput {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Value => "value",
        }
    }

    pub fn is_known(&self) -> bool {
        matches!(self, Self::Value)
    }
}

impl std::fmt::Display for MaximumEntropySpectrumOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Maximum Entropy Spectrum indicator.
pub struct MaximumEntropySpectrumParams {
    /// Number of time periods in the spectrum window. Default 60.
    pub length: usize,
    /// Order of the auto-regression model (Burg). Must be > 0 and < length. Default 30.
    pub degree: usize,
    /// Minimum cycle period (>= 2). Default 2.
    pub min_period: f64,
    /// Maximum cycle period (> min_period, <= 2*length). Default 59.
    pub max_period: f64,
    /// Spectrum resolution (positive integer). Default 1.
    pub spectrum_resolution: usize,
    /// Disable automatic gain control. Default false (AGC on).
    pub disable_automatic_gain_control: bool,
    /// AGC decay factor in (0, 1). Default 0.995.
    pub automatic_gain_control_decay_factor: f64,
    /// Use fixed (0-clamped) normalization. Default false (floating).
    pub fixed_normalization: bool,
    /// Bar component. `None` → Median (hl/2).
    pub bar_component: Option<BarComponent>,
    /// Quote component. `None` → Mid.
    pub quote_component: Option<QuoteComponent>,
    /// Trade component. `None` → Price.
    pub trade_component: Option<TradeComponent>,
}

impl Default for MaximumEntropySpectrumParams {
    fn default() -> Self {
        Self {
            length: 0,
            degree: 0,
            min_period: 0.0,
            max_period: 0.0,
            spectrum_resolution: 0,
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
    degree: usize,
    length_spectrum: usize,
    is_automatic_gain_control: bool,
    automatic_gain_control_decay_factor: f64,

    input_series: Vec<f64>,
    input_series_minus_mean: Vec<f64>,
    coefficients: Vec<f64>,
    spectrum: Vec<f64>,
    period: Vec<f64>,

    // Pre-computed trig tables [length_spectrum][degree].
    frequency_sin_omega: Vec<Vec<f64>>,
    frequency_cos_omega: Vec<Vec<f64>>,

    // Burg working buffers.
    h: Vec<f64>,
    g: Vec<f64>,
    per: Vec<f64>,
    pef: Vec<f64>,

    spectrum_min: f64,
    spectrum_max: f64,
    previous_spectrum_max: f64,
}

impl Estimator {
    fn new(
        length: usize,
        degree: usize,
        min_period: f64,
        max_period: f64,
        spectrum_resolution: usize,
        is_automatic_gain_control: bool,
        automatic_gain_control_decay_factor: f64,
    ) -> Self {
        let two_pi = 2.0 * PI;
        let length_spectrum =
            ((max_period - min_period) * spectrum_resolution as f64) as usize + 1;
        let res = spectrum_resolution as f64;

        let mut period = vec![0.0; length_spectrum];
        let mut frequency_sin_omega = Vec::with_capacity(length_spectrum);
        let mut frequency_cos_omega = Vec::with_capacity(length_spectrum);

        // Spectrum is evaluated from MaxPeriod down to MinPeriod.
        for i in 0..length_spectrum {
            let p = max_period - i as f64 / res;
            period[i] = p;
            let theta = two_pi / p;

            let mut sin_row = vec![0.0; degree];
            let mut cos_row = vec![0.0; degree];

            for j in 0..degree {
                let omega = -(j as f64 + 1.0) * theta;
                sin_row[j] = omega.sin();
                cos_row[j] = omega.cos();
            }

            frequency_sin_omega.push(sin_row);
            frequency_cos_omega.push(cos_row);
        }

        Self {
            length,
            degree,
            length_spectrum,
            is_automatic_gain_control,
            automatic_gain_control_decay_factor,
            input_series: vec![0.0; length],
            input_series_minus_mean: vec![0.0; length],
            coefficients: vec![0.0; degree],
            spectrum: vec![0.0; length_spectrum],
            period,
            frequency_sin_omega,
            frequency_cos_omega,
            h: vec![0.0; degree + 1],
            g: vec![0.0; degree + 2],
            per: vec![0.0; length + 1],
            pef: vec![0.0; length + 1],
            spectrum_min: 0.0,
            spectrum_max: 0.0,
            previous_spectrum_max: 0.0,
        }
    }

    fn calculate(&mut self) {
        // Subtract the mean from the input series.
        let mut mean = 0.0;
        for i in 0..self.length {
            mean += self.input_series[i];
        }
        mean /= self.length as f64;

        for i in 0..self.length {
            self.input_series_minus_mean[i] = self.input_series[i] - mean;
        }

        self.burg_estimate();

        // Evaluate the spectrum from the AR coefficients.
        self.spectrum_min = f64::MAX;
        if self.is_automatic_gain_control {
            self.spectrum_max =
                self.automatic_gain_control_decay_factor * self.previous_spectrum_max;
        } else {
            self.spectrum_max = f64::MIN;
        }

        for i in 0..self.length_spectrum {
            let mut real = 1.0;
            let mut imag = 0.0;

            let cos_row = &self.frequency_cos_omega[i];
            let sin_row = &self.frequency_sin_omega[i];

            for j in 0..self.degree {
                real -= self.coefficients[j] * cos_row[j];
                imag -= self.coefficients[j] * sin_row[j];
            }

            let s = 1.0 / (real * real + imag * imag);
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

    /// Burg maximum-entropy AR coefficient estimation.
    /// Direct port of the zero-based C reference from Paul Bourke's ar.h suite.
    fn burg_estimate(&mut self) {
        let length = self.length;
        let degree = self.degree;

        for i in 1..=length {
            self.pef[i] = 0.0;
            self.per[i] = 0.0;
        }

        for i in 1..=degree {
            let mut sn: f64 = 0.0;
            let mut sd: f64 = 0.0;

            let jj = length - i;

            for j in 0..jj {
                let t1 = self.input_series_minus_mean[j + i] + self.pef[j];
                let t2 = self.input_series_minus_mean[j] + self.per[j];
                sn -= 2.0 * t1 * t2;
                sd += t1 * t1 + t2 * t2;
            }

            let t = sn / sd;
            self.g[i] = t;

            if i != 1 {
                for j in 1..i {
                    self.h[j] = self.g[j] + t * self.g[i - j];
                }

                for j in 1..i {
                    self.g[j] = self.h[j];
                }
            }

            let jj2 = if i != 1 { jj - 1 } else { jj };

            for j in 0..jj2 {
                self.per[j] += t * self.pef[j] + t * self.input_series_minus_mean[j + i];
                self.pef[j] = self.pef[j + 1]
                    + t * self.per[j + 1]
                    + t * self.input_series_minus_mean[j + 1];
            }
        }

        for i in 0..degree {
            self.coefficients[i] = -self.g[i + 1];
        }
    }
}

// ---------------------------------------------------------------------------
// MaximumEntropySpectrum indicator
// ---------------------------------------------------------------------------

/// MBST's Maximum Entropy Spectrum heatmap indicator.
///
/// Displays a power heatmap of cyclic activity over a configurable cycle-period
/// range using Burg's maximum-entropy auto-regressive method.
pub struct MaximumEntropySpectrum {
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

impl MaximumEntropySpectrum {
    /// Creates an instance with default parameters.
    pub fn default_params() -> Result<Self, String> {
        Self::new(&MaximumEntropySpectrumParams::default())
    }

    /// Creates an instance with the supplied parameters.
    pub fn new(p: &MaximumEntropySpectrumParams) -> Result<Self, String> {
        let invalid = "invalid maximum entropy spectrum parameters";

        let def_length: usize = 60;
        let def_degree: usize = 30;
        let def_min_period: f64 = 2.0;
        let def_max_period: f64 = 59.0;
        let def_spectrum_resolution: usize = 1;
        let def_agc_decay: f64 = 0.995;
        let agc_decay_epsilon: f64 = 1e-12;

        let length = if p.length == 0 { def_length } else { p.length };
        let degree = if p.degree == 0 { def_degree } else { p.degree };
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

        let agc_on = !p.disable_automatic_gain_control;
        let floating_norm = !p.fixed_normalization;

        if length < 2 {
            return Err(format!("{}: Length should be >= 2", invalid));
        }
        if degree == 0 || degree >= length {
            return Err(format!("{}: Degree should be > 0 and < Length", invalid));
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

        let flags = build_flag_tags(agc_on, agc_decay, floating_norm, def_agc_decay, agc_decay_epsilon);

        let mnemonic = format!(
            "mespect({}, {}, {}, {}, {}{}{})",
            length, degree, min_period, max_period, spectrum_resolution, flags, component_mnemonic
        );
        let description = format!("Maximum entropy spectrum {}", mnemonic);

        let estimator = Estimator::new(
            length,
            degree,
            min_period,
            max_period,
            spectrum_resolution,
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

    /// Returns the mnemonic string.
    pub fn mnemonic(&self) -> &str {
        &self.mnemonic
    }

    /// Feeds the next sample and returns the heatmap column.
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

        // MBST fills spectrum[0] at MaxPeriod and spectrum[last] at MinPeriod.
        // The heatmap axis runs MinPeriod -> MaxPeriod, so reverse on output.
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
    agc_on: bool,
    agc_decay: f64,
    floating_norm: bool,
    def_agc: f64,
    eps: f64,
) -> String {
    let mut s = String::new();

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

impl Indicator for MaximumEntropySpectrum {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::MaximumEntropySpectrum,
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

    fn test_mes_time() -> i64 {
        0
    }

    fn test_mes_input() -> Vec<f64> {
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

    struct MesSnap {
        input_index: usize,
        value_min: f64,
        value_max: f64,
        spots: Vec<SpotValue>,
    }

    fn mes_snapshots() -> Vec<MesSnap> {
        vec![
            MesSnap {
                input_index: 59,
                value_min: 0.0,
                value_max: 1.0,
                spots: vec![
                    SpotValue { i: 0, v: 0.000000000000000 },
                    SpotValue { i: 14, v: 0.124709393535801 },
                    SpotValue { i: 28, v: 0.021259483287733 },
                    SpotValue { i: 42, v: 0.726759100473496 },
                    SpotValue { i: 57, v: 0.260829244402141 },
                ],
            },
            MesSnap {
                input_index: 60,
                value_min: 0.0,
                value_max: 0.3803558166,
                spots: vec![
                    SpotValue { i: 0, v: 0.000000000000000 },
                    SpotValue { i: 14, v: 0.047532484316402 },
                    SpotValue { i: 28, v: 0.156007210177695 },
                    SpotValue { i: 42, v: 0.204392941920655 },
                    SpotValue { i: 57, v: 0.099988829337396 },
                ],
            },
            MesSnap {
                input_index: 100,
                value_min: 0.0,
                value_max: 0.7767627734,
                spots: vec![
                    SpotValue { i: 0, v: 0.000000000000000 },
                    SpotValue { i: 14, v: 0.005541589459818 },
                    SpotValue { i: 28, v: 0.019544065000896 },
                    SpotValue { i: 42, v: 0.045342308770863 },
                    SpotValue { i: 57, v: 0.776762773404885 },
                ],
            },
            MesSnap {
                input_index: 150,
                value_min: 0.0,
                value_max: 0.0126783313,
                spots: vec![
                    SpotValue { i: 0, v: 0.000347619185321 },
                    SpotValue { i: 14, v: 0.001211800388686 },
                    SpotValue { i: 28, v: 0.001749939543675 },
                    SpotValue { i: 42, v: 0.010949450171300 },
                    SpotValue { i: 57, v: 0.001418701588812 },
                ],
            },
            MesSnap {
                input_index: 200,
                value_min: 0.0,
                value_max: 0.5729940203,
                spots: vec![
                    SpotValue { i: 0, v: 0.000000000000000 },
                    SpotValue { i: 14, v: 0.047607367831419 },
                    SpotValue { i: 28, v: 0.013304430092822 },
                    SpotValue { i: 42, v: 0.137193402225458 },
                    SpotValue { i: 57, v: 0.506646287515276 },
                ],
            },
        ]
    }

    #[test]
    fn test_maximum_entropy_spectrum_update() {
        let input = test_mes_input();
        let t0 = test_mes_time();

        let mut x = MaximumEntropySpectrum::default_params().unwrap();
        let snaps = mes_snapshots();
        let mut si = 0;

        for i in 0..input.len() {
            let h = x.update(input[i], t0 + i as i64);

            assert_eq!(h.parameter_first, 2.0, "[{}] parameter_first", i);
            assert_eq!(h.parameter_last, 59.0, "[{}] parameter_last", i);
            assert_eq!(h.parameter_resolution, 1.0, "[{}] parameter_resolution", i);

            if !x.is_primed() {
                assert!(h.is_empty(), "[{}] expected empty heatmap before priming", i);
                continue;
            }

            assert_eq!(h.values.len(), 58, "[{}] values len", i);

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
    fn test_maximum_entropy_spectrum_primes_at_bar_59() {
        let mut x = MaximumEntropySpectrum::default_params().unwrap();
        assert!(!x.is_primed());

        let input = test_mes_input();
        let t0 = test_mes_time();
        let mut primed_at: Option<usize> = None;

        for i in 0..input.len() {
            x.update(input[i], t0 + i as i64);
            if x.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert_eq!(primed_at, Some(59));
    }

    #[test]
    fn test_maximum_entropy_spectrum_nan_input() {
        let mut x = MaximumEntropySpectrum::default_params().unwrap();

        let h = x.update(f64::NAN, test_mes_time());
        assert!(h.is_empty());
        assert!(!x.is_primed());
    }

    #[test]
    fn test_maximum_entropy_spectrum_metadata() {
        let x = MaximumEntropySpectrum::default_params().unwrap();
        let md = x.metadata();

        let mn = "mespect(60, 30, 2, 59, 1, hl/2)";

        assert_eq!(md.identifier, Identifier::MaximumEntropySpectrum);
        assert_eq!(md.mnemonic, mn);
        assert_eq!(md.description, format!("Maximum entropy spectrum {}", mn));
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].kind, MaximumEntropySpectrumOutput::Value as i32);
        assert_eq!(md.outputs[0].shape, Shape::Heatmap);
        assert_eq!(md.outputs[0].mnemonic, mn);
    }

    #[test]
    fn test_maximum_entropy_spectrum_mnemonic_flags() {
        let cases: Vec<(MaximumEntropySpectrumParams, &str)> = vec![
            (MaximumEntropySpectrumParams::default(), "mespect(60, 30, 2, 59, 1, hl/2)"),
            (
                MaximumEntropySpectrumParams {
                    disable_automatic_gain_control: true,
                    ..Default::default()
                },
                "mespect(60, 30, 2, 59, 1, no-agc, hl/2)",
            ),
            (
                MaximumEntropySpectrumParams {
                    automatic_gain_control_decay_factor: 0.8,
                    ..Default::default()
                },
                "mespect(60, 30, 2, 59, 1, agc=0.8, hl/2)",
            ),
            (
                MaximumEntropySpectrumParams {
                    fixed_normalization: true,
                    ..Default::default()
                },
                "mespect(60, 30, 2, 59, 1, no-fn, hl/2)",
            ),
            (
                MaximumEntropySpectrumParams {
                    disable_automatic_gain_control: true,
                    fixed_normalization: true,
                    ..Default::default()
                },
                "mespect(60, 30, 2, 59, 1, no-agc, no-fn, hl/2)",
            ),
        ];

        for (p, expected_mn) in &cases {
            let x = MaximumEntropySpectrum::new(p).unwrap();
            assert_eq!(x.mnemonic(), *expected_mn, "params produced wrong mnemonic");
        }
    }

    #[test]
    fn test_maximum_entropy_spectrum_validation() {
        let cases: Vec<(MaximumEntropySpectrumParams, &str)> = vec![
            (
                MaximumEntropySpectrumParams {
                    length: 1, degree: 1, min_period: 2.0, max_period: 4.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid maximum entropy spectrum parameters: Length should be >= 2",
            ),
            (
                MaximumEntropySpectrumParams {
                    length: 4, degree: 4, min_period: 2.0, max_period: 4.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid maximum entropy spectrum parameters: Degree should be > 0 and < Length",
            ),
            (
                MaximumEntropySpectrumParams {
                    length: 60, degree: 30, min_period: 1.0, max_period: 59.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid maximum entropy spectrum parameters: MinPeriod should be >= 2",
            ),
            (
                MaximumEntropySpectrumParams {
                    length: 60, degree: 30, min_period: 10.0, max_period: 10.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid maximum entropy spectrum parameters: MaxPeriod should be > MinPeriod",
            ),
            (
                MaximumEntropySpectrumParams {
                    length: 10, degree: 5, min_period: 2.0, max_period: 59.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid maximum entropy spectrum parameters: MaxPeriod should be <= 2 * Length",
            ),
            (
                MaximumEntropySpectrumParams {
                    automatic_gain_control_decay_factor: -0.1,
                    ..Default::default()
                },
                "invalid maximum entropy spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
            ),
            (
                MaximumEntropySpectrumParams {
                    automatic_gain_control_decay_factor: 1.0,
                    ..Default::default()
                },
                "invalid maximum entropy spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
            ),
        ];

        for (p, expected_msg) in &cases {
            match MaximumEntropySpectrum::new(p) {
                Ok(_) => panic!("expected error for: {}", expected_msg),
                Err(e) => assert_eq!(e, *expected_msg),
            }
        }
    }

    #[test]
    fn test_maximum_entropy_spectrum_update_entity() {
        let input = test_mes_input();
        let t0 = test_mes_time();
        let prime_count = 70;

        // Test update_scalar
        {
            let mut x = MaximumEntropySpectrum::default_params().unwrap();
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
            let mut x = MaximumEntropySpectrum::default_params().unwrap();
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
            let mut x = MaximumEntropySpectrum::default_params().unwrap();
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
            let mut x = MaximumEntropySpectrum::default_params().unwrap();
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
        assert_eq!(MaximumEntropySpectrumOutput::Value.as_str(), "value");
        assert!(MaximumEntropySpectrumOutput::Value.is_known());
    }
}
