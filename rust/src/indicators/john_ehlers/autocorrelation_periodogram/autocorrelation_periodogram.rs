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
pub enum AutoCorrelationPeriodogramOutput {
    Value = 1,
}

impl AutoCorrelationPeriodogramOutput {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Value => "value",
        }
    }

    pub fn is_known(&self) -> bool {
        matches!(self, Self::Value)
    }
}

impl std::fmt::Display for AutoCorrelationPeriodogramOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

pub struct AutoCorrelationPeriodogramParams {
    pub min_period: i32,
    pub max_period: i32,
    pub averaging_length: i32,
    pub disable_spectral_squaring: bool,
    pub disable_smoothing: bool,
    pub disable_automatic_gain_control: bool,
    pub automatic_gain_control_decay_factor: f64,
    pub fixed_normalization: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for AutoCorrelationPeriodogramParams {
    fn default() -> Self {
        Self {
            min_period: 0,
            max_period: 0,
            averaging_length: 0,
            disable_spectral_squaring: false,
            disable_smoothing: false,
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
    min_period: usize,
    max_period: usize,
    averaging_length: usize,
    length_spectrum: usize,
    filt_buffer_len: usize,

    is_spectral_squaring: bool,
    is_smoothing: bool,
    is_automatic_gain_control: bool,
    automatic_gain_control_decay_factor: f64,

    // Pre-filter coefficients.
    coeff_hp0: f64,
    coeff_hp1: f64,
    coeff_hp2: f64,
    ss_c1: f64,
    ss_c2: f64,
    ss_c3: f64,

    // DFT basis tables: cos_tab[i][n], sin_tab[i][n].
    cos_tab: Vec<Vec<f64>>,
    sin_tab: Vec<Vec<f64>>,

    // Pre-filter state.
    close0: f64,
    close1: f64,
    close2: f64,
    hp0: f64,
    hp1: f64,
    hp2: f64,

    // Filt history: filt[k] = Filt k bars ago (0 = current).
    filt: Vec<f64>,

    // Per-lag Pearson correlation coefficients.
    corr: Vec<f64>,

    // Smoothed power per period bin.
    r_previous: Vec<f64>,

    // Normalized spectrum values (output).
    spectrum: Vec<f64>,
    spectrum_min: f64,
    spectrum_max: f64,
    previous_spectrum_max: f64,
}

impl Estimator {
    fn new(
        min_period: usize,
        max_period: usize,
        averaging_length: usize,
        is_spectral_squaring: bool,
        is_smoothing: bool,
        is_agc: bool,
        agc_decay: f64,
    ) -> Self {
        let two_pi = 2.0 * PI;
        let dft_lag_start: usize = 3;

        let length_spectrum = max_period - min_period + 1;
        let filt_buffer_len = max_period + averaging_length;
        let corr_len = max_period + 1;

        // Highpass coefficients, cutoff at MaxPeriod.
        let omega_hp = 0.707 * two_pi / max_period as f64;
        let alpha_hp = (omega_hp.cos() + omega_hp.sin() - 1.0) / omega_hp.cos();
        let c_hp0 = (1.0 - alpha_hp / 2.0) * (1.0 - alpha_hp / 2.0);
        let c_hp1 = 2.0 * (1.0 - alpha_hp);
        let c_hp2 = (1.0 - alpha_hp) * (1.0 - alpha_hp);

        // SuperSmoother coefficients, period = MinPeriod.
        let a1 = (-1.414 * PI / min_period as f64).exp();
        let b1 = 2.0 * a1 * (1.414 * PI / min_period as f64).cos();
        let ss_c2 = b1;
        let ss_c3 = -a1 * a1;
        let ss_c1 = 1.0 - ss_c2 - ss_c3;

        // DFT basis tables.
        let mut cos_tab = Vec::with_capacity(length_spectrum);
        let mut sin_tab = Vec::with_capacity(length_spectrum);

        for i in 0..length_spectrum {
            let period = min_period + i;
            let mut cos_row = vec![0.0; corr_len];
            let mut sin_row = vec![0.0; corr_len];

            for n in dft_lag_start..corr_len {
                let angle = two_pi * n as f64 / period as f64;
                cos_row[n] = angle.cos();
                sin_row[n] = angle.sin();
            }

            cos_tab.push(cos_row);
            sin_tab.push(sin_row);
        }

        Self {
            min_period,
            max_period,
            averaging_length,
            length_spectrum,
            filt_buffer_len,
            is_spectral_squaring,
            is_smoothing,
            is_automatic_gain_control: is_agc,
            automatic_gain_control_decay_factor: agc_decay,
            coeff_hp0: c_hp0,
            coeff_hp1: c_hp1,
            coeff_hp2: c_hp2,
            ss_c1,
            ss_c2,
            ss_c3,
            cos_tab,
            sin_tab,
            close0: 0.0,
            close1: 0.0,
            close2: 0.0,
            hp0: 0.0,
            hp1: 0.0,
            hp2: 0.0,
            filt: vec![0.0; filt_buffer_len],
            corr: vec![0.0; corr_len],
            r_previous: vec![0.0; length_spectrum],
            spectrum: vec![0.0; length_spectrum],
            spectrum_min: 0.0,
            spectrum_max: 0.0,
            previous_spectrum_max: 0.0,
        }
    }

    fn update(&mut self, sample: f64) {
        let dft_lag_start: usize = 3;

        // Pre-filter cascade.
        self.close2 = self.close1;
        self.close1 = self.close0;
        self.close0 = sample;

        self.hp2 = self.hp1;
        self.hp1 = self.hp0;
        self.hp0 = self.coeff_hp0 * (self.close0 - 2.0 * self.close1 + self.close2)
            + self.coeff_hp1 * self.hp1
            - self.coeff_hp2 * self.hp2;

        // Shift Filt history rightward; new Filt at filt[0].
        for k in (1..self.filt_buffer_len).rev() {
            self.filt[k] = self.filt[k - 1];
        }

        self.filt[0] =
            self.ss_c1 * (self.hp0 + self.hp1) / 2.0 + self.ss_c2 * self.filt[1] + self.ss_c3 * self.filt[2];

        // Pearson correlation per lag [0..max_period], fixed M = averaging_length.
        let m = self.averaging_length;

        for lag in 0..=self.max_period {
            let mut sx = 0.0;
            let mut sy = 0.0;
            let mut sxx = 0.0;
            let mut syy = 0.0;
            let mut sxy = 0.0;

            for c in 0..m {
                let x = self.filt[c];
                let y = self.filt[lag + c];
                sx += x;
                sy += y;
                sxx += x * x;
                syy += y * y;
                sxy += x * y;
            }

            let denom = (m as f64 * sxx - sx * sx) * (m as f64 * syy - sy * sy);

            let r = if denom > 0.0 {
                (m as f64 * sxy - sx * sy) / denom.sqrt()
            } else {
                0.0
            };

            self.corr[lag] = r;
        }

        // DFT of the correlation function, per period bin.
        self.spectrum_min = f64::MAX;
        if self.is_automatic_gain_control {
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max;
        } else {
            self.spectrum_max = f64::MIN;
        }

        // Pass 1: compute raw R values and track the running max for AGC.
        for i in 0..self.length_spectrum {
            let cos_row = &self.cos_tab[i];
            let sin_row = &self.sin_tab[i];

            let mut cos_part = 0.0;
            let mut sin_part = 0.0;

            for n in dft_lag_start..=self.max_period {
                cos_part += self.corr[n] * cos_row[n];
                sin_part += self.corr[n] * sin_row[n];
            }

            let sq_sum = cos_part * cos_part + sin_part * sin_part;

            let raw = if self.is_spectral_squaring {
                sq_sum * sq_sum
            } else {
                sq_sum
            };

            let r = if self.is_smoothing {
                0.2 * raw + 0.8 * self.r_previous[i]
            } else {
                raw
            };

            self.r_previous[i] = r;
            self.spectrum[i] = r;

            if self.spectrum_max < r {
                self.spectrum_max = r;
            }
        }

        self.previous_spectrum_max = self.spectrum_max;

        // Pass 2: normalize against the running max and track the (normalized) min.
        if self.spectrum_max > 0.0 {
            for i in 0..self.length_spectrum {
                let v = self.spectrum[i] / self.spectrum_max;
                self.spectrum[i] = v;

                if self.spectrum_min > v {
                    self.spectrum_min = v;
                }
            }
        } else {
            for i in 0..self.length_spectrum {
                self.spectrum[i] = 0.0;
            }
            self.spectrum_min = 0.0;
        }
    }
}

// ---------------------------------------------------------------------------
// AutoCorrelationPeriodogram indicator
// ---------------------------------------------------------------------------

pub struct AutoCorrelationPeriodogram {
    mnemonic: String,
    description: String,
    estimator: Estimator,
    window_count: usize,
    prime_count: usize,
    primed: bool,
    floating_normalization: bool,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl AutoCorrelationPeriodogram {
    pub fn default_params() -> Result<Self, String> {
        Self::new(&AutoCorrelationPeriodogramParams::default())
    }

    pub fn new(p: &AutoCorrelationPeriodogramParams) -> Result<Self, String> {
        let invalid = "invalid autocorrelation periodogram parameters";

        let def_min_period: i32 = 10;
        let def_max_period: i32 = 48;
        let def_averaging_len: i32 = 3;
        let def_agc_decay: f64 = 0.995;
        let agc_decay_epsilon: f64 = 1e-12;

        let min_period = if p.min_period == 0 { def_min_period } else { p.min_period };
        let max_period = if p.max_period == 0 { def_max_period } else { p.max_period };
        let averaging_length = if p.averaging_length == 0 { def_averaging_len } else { p.averaging_length };
        let agc_decay = if p.automatic_gain_control_decay_factor == 0.0 {
            def_agc_decay
        } else {
            p.automatic_gain_control_decay_factor
        };

        let squaring_on = !p.disable_spectral_squaring;
        let smoothing_on = !p.disable_smoothing;
        let agc_on = !p.disable_automatic_gain_control;
        let floating_norm = !p.fixed_normalization;

        if min_period < 2 {
            return Err(format!("{}: MinPeriod should be >= 2", invalid));
        }
        if max_period <= min_period {
            return Err(format!("{}: MaxPeriod should be > MinPeriod", invalid));
        }
        if averaging_length < 1 {
            return Err(format!("{}: AveragingLength should be >= 1", invalid));
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

        let flags = build_flag_tags(
            averaging_length, squaring_on, smoothing_on, agc_on, agc_decay, floating_norm,
            def_averaging_len, def_agc_decay, agc_decay_epsilon,
        );

        let mnemonic = format!(
            "acp({}, {}{}{})",
            min_period, max_period, flags, component_mnemonic
        );
        let description = format!("Autocorrelation periodogram {}", mnemonic);

        let estimator = Estimator::new(
            min_period as usize,
            max_period as usize,
            averaging_length as usize,
            squaring_on,
            smoothing_on,
            agc_on,
            agc_decay,
        );

        let prime_count = estimator.filt_buffer_len;

        Ok(Self {
            mnemonic,
            description,
            estimator,
            window_count: 0,
            prime_count,
            primed: false,
            floating_normalization: floating_norm,
            min_parameter_value: min_period as f64,
            max_parameter_value: max_period as f64,
            parameter_resolution: 1.0,
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

        self.estimator.update(sample);

        if !self.primed {
            self.window_count += 1;
            if self.window_count >= self.prime_count {
                self.primed = true;
            } else {
                return Heatmap::empty(
                    time,
                    self.min_parameter_value,
                    self.max_parameter_value,
                    self.parameter_resolution,
                );
            }
        }

        let length_spectrum = self.estimator.length_spectrum;

        let min_ref = if self.floating_normalization {
            self.estimator.spectrum_min
        } else {
            0.0
        };

        // Estimator spectrum is already AGC-normalized in [0, 1]. Apply optional
        // floating-minimum subtraction for display.
        let max_ref = 1.0;
        let spectrum_range = max_ref - min_ref;

        let mut values = vec![0.0; length_spectrum];
        let mut value_min = f64::INFINITY;
        let mut value_max = f64::NEG_INFINITY;

        for i in 0..length_spectrum {
            let v = if spectrum_range > 0.0 {
                (self.estimator.spectrum[i] - min_ref) / spectrum_range
            } else {
                0.0
            };

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
    averaging_length: i32,
    squaring_on: bool,
    smoothing_on: bool,
    agc_on: bool,
    agc_decay: f64,
    floating_norm: bool,
    def_averaging: i32,
    def_agc: f64,
    agc_eps: f64,
) -> String {
    let mut s = String::new();

    if averaging_length != def_averaging {
        s.push_str(&format!(", average={}", averaging_length));
    }
    if !squaring_on {
        s.push_str(", no-sqr");
    }
    if !smoothing_on {
        s.push_str(", no-smooth");
    }
    if !agc_on {
        s.push_str(", no-agc");
    }
    if agc_on && (agc_decay - def_agc).abs() > agc_eps {
        s.push_str(&format!(", agc={}", agc_decay));
    }
    if !floating_norm {
        s.push_str(", no-fn");
    }

    s
}

impl Indicator for AutoCorrelationPeriodogram {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::AutoCorrelationPeriodogram,
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

    fn test_acp_time() -> i64 {
        0
    }

    fn test_acp_input() -> Vec<f64> {
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

    struct AcpSnap {
        input_index: usize,
        value_min: f64,
        value_max: f64,
        spots: Vec<SpotValue>,
    }

    fn acp_snapshots() -> Vec<AcpSnap> {
        vec![
            AcpSnap {
                input_index: 120,
                value_min: 0.0,
                value_max: 0.587896156859073,
                spots: vec![
                    SpotValue { i: 0, v: 0.005178925655054 },
                    SpotValue { i: 9, v: 0.002809468053362 },
                    SpotValue { i: 19, v: 0.003090504825683 },
                    SpotValue { i: 28, v: 0.000546592441807 },
                    SpotValue { i: 38, v: 0.000986630131293 },
                ],
            },
            AcpSnap {
                input_index: 150,
                value_min: 0.0,
                value_max: 0.176709059744408,
                spots: vec![
                    SpotValue { i: 0, v: 0.026721265051811 },
                    SpotValue { i: 9, v: 0.092237395343329 },
                    SpotValue { i: 19, v: 0.000025883994424 },
                    SpotValue { i: 28, v: 0.023733698891158 },
                    SpotValue { i: 38, v: 0.028877871176187 },
                ],
            },
            AcpSnap {
                input_index: 200,
                value_min: 0.0,
                value_max: 0.691497315922981,
                spots: vec![
                    SpotValue { i: 0, v: 0.000000000000000 },
                    SpotValue { i: 9, v: 0.664761990766748 },
                    SpotValue { i: 19, v: 0.004025474045612 },
                    SpotValue { i: 28, v: 0.017601099472114 },
                    SpotValue { i: 38, v: 0.055697077605643 },
                ],
            },
            AcpSnap {
                input_index: 250,
                value_min: 0.0,
                value_max: 0.133271103774289,
                spots: vec![
                    SpotValue { i: 0, v: 0.042309937132732 },
                    SpotValue { i: 9, v: 0.001944553293214 },
                    SpotValue { i: 19, v: 0.003966252606748 },
                    SpotValue { i: 28, v: 0.029986716389868 },
                    SpotValue { i: 38, v: 0.052695592829157 },
                ],
            },
        ]
    }

    #[test]
    fn test_acp_update() {
        let input = test_acp_input();
        let t0 = test_acp_time();

        let mut x = AutoCorrelationPeriodogram::default_params().unwrap();
        let snaps = acp_snapshots();
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
    fn test_acp_nan_input() {
        let mut x = AutoCorrelationPeriodogram::default_params().unwrap();

        let h = x.update(f64::NAN, test_acp_time());
        assert!(h.is_empty());
        assert!(!x.is_primed());
    }

    #[test]
    fn test_acp_synthetic_sine() {
        let period: f64 = 20.0;
        let bars = 600;

        let mut x = AutoCorrelationPeriodogram::new(&AutoCorrelationPeriodogramParams {
            disable_automatic_gain_control: true,
            fixed_normalization: true,
            ..Default::default()
        })
        .unwrap();

        let t0 = test_acp_time();
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

        // Bin k corresponds to period MinPeriod+k. MinPeriod=10, period=20 -> bin 10.
        let expected_bin = (period - last.parameter_first) as usize;
        assert_eq!(
            peak_bin, expected_bin,
            "peak bin: expected {} (period {}), got {} (period {})",
            expected_bin, period, peak_bin, last.parameter_first + peak_bin as f64
        );
    }

    #[test]
    fn test_acp_metadata() {
        let x = AutoCorrelationPeriodogram::default_params().unwrap();
        let md = x.metadata();

        let mn = "acp(10, 48, hl/2)";

        assert_eq!(md.identifier, Identifier::AutoCorrelationPeriodogram);
        assert_eq!(md.mnemonic, mn);
        assert_eq!(md.description, format!("Autocorrelation periodogram {}", mn));
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].kind, AutoCorrelationPeriodogramOutput::Value as i32);
        assert_eq!(md.outputs[0].shape, Shape::Heatmap);
        assert_eq!(md.outputs[0].mnemonic, mn);
    }

    #[test]
    fn test_acp_mnemonic_flags() {
        let cases: Vec<(AutoCorrelationPeriodogramParams, &str)> = vec![
            (AutoCorrelationPeriodogramParams::default(), "acp(10, 48, hl/2)"),
            (
                AutoCorrelationPeriodogramParams {
                    averaging_length: 5,
                    ..Default::default()
                },
                "acp(10, 48, average=5, hl/2)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    disable_spectral_squaring: true,
                    ..Default::default()
                },
                "acp(10, 48, no-sqr, hl/2)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    disable_smoothing: true,
                    ..Default::default()
                },
                "acp(10, 48, no-smooth, hl/2)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    disable_automatic_gain_control: true,
                    ..Default::default()
                },
                "acp(10, 48, no-agc, hl/2)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    automatic_gain_control_decay_factor: 0.8,
                    ..Default::default()
                },
                "acp(10, 48, agc=0.8, hl/2)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    fixed_normalization: true,
                    ..Default::default()
                },
                "acp(10, 48, no-fn, hl/2)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    averaging_length: 5,
                    disable_spectral_squaring: true,
                    disable_smoothing: true,
                    disable_automatic_gain_control: true,
                    fixed_normalization: true,
                    ..Default::default()
                },
                "acp(10, 48, average=5, no-sqr, no-smooth, no-agc, no-fn, hl/2)",
            ),
        ];

        for (p, expected_mn) in &cases {
            let x = AutoCorrelationPeriodogram::new(p).unwrap();
            assert_eq!(x.mnemonic(), *expected_mn, "params produced wrong mnemonic");
        }
    }

    #[test]
    fn test_acp_validation() {
        let cases: Vec<(AutoCorrelationPeriodogramParams, &str)> = vec![
            (
                AutoCorrelationPeriodogramParams {
                    min_period: 1, max_period: 48, averaging_length: 3,
                    ..Default::default()
                },
                "invalid autocorrelation periodogram parameters: MinPeriod should be >= 2",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    min_period: 10, max_period: 10, averaging_length: 3,
                    ..Default::default()
                },
                "invalid autocorrelation periodogram parameters: MaxPeriod should be > MinPeriod",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    averaging_length: -1,
                    ..Default::default()
                },
                "invalid autocorrelation periodogram parameters: AveragingLength should be >= 1",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    automatic_gain_control_decay_factor: -0.1,
                    ..Default::default()
                },
                "invalid autocorrelation periodogram parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    automatic_gain_control_decay_factor: 1.0,
                    ..Default::default()
                },
                "invalid autocorrelation periodogram parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
            ),
        ];

        for (p, expected_msg) in &cases {
            match AutoCorrelationPeriodogram::new(p) {
                Ok(_) => panic!("expected error for: {}", expected_msg),
                Err(e) => assert_eq!(e, *expected_msg),
            }
        }
    }

    #[test]
    fn test_acp_update_entity() {
        let input = test_acp_input();
        let t0 = test_acp_time();
        let prime_count = 100;

        // Test update_scalar
        {
            let mut x = AutoCorrelationPeriodogram::default_params().unwrap();
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
            let mut x = AutoCorrelationPeriodogram::default_params().unwrap();
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
            let mut x = AutoCorrelationPeriodogram::default_params().unwrap();
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
            let mut x = AutoCorrelationPeriodogram::default_params().unwrap();
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
        assert_eq!(AutoCorrelationPeriodogramOutput::Value.as_str(), "value");
        assert!(AutoCorrelationPeriodogramOutput::Value.is_known());
    }
}
