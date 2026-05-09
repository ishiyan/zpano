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

/// Describes the outputs of the Goertzel Spectrum indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum GoertzelSpectrumOutput {
    /// The Goertzel spectrum heatmap column.
    Value = 1,
}

impl GoertzelSpectrumOutput {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Value => "value",
        }
    }

    pub fn is_known(&self) -> bool {
        matches!(self, Self::Value)
    }
}

impl std::fmt::Display for GoertzelSpectrumOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Goertzel Spectrum indicator.
pub struct GoertzelSpectrumParams {
    /// Number of time periods in the spectrum window. Default 64.
    pub length: usize,
    /// Minimum cycle period (>= 2). Default 2.
    pub min_period: f64,
    /// Maximum cycle period (> min_period, <= 2*length). Default 64.
    pub max_period: f64,
    /// Spectrum resolution (positive integer). Default 1.
    pub spectrum_resolution: usize,
    /// Use first-order Goertzel algorithm. Default false.
    pub is_first_order: bool,
    /// Disable spectral dilation compensation. Default false (compensation on).
    pub disable_spectral_dilation_compensation: bool,
    /// Disable automatic gain control. Default false (AGC on).
    pub disable_automatic_gain_control: bool,
    /// AGC decay factor in (0, 1). Default 0.991.
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

impl Default for GoertzelSpectrumParams {
    fn default() -> Self {
        Self {
            length: 0,
            min_period: 0.0,
            max_period: 0.0,
            spectrum_resolution: 0,
            is_first_order: false,
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
    is_first_order: bool,
    is_spectral_dilation_compensation: bool,
    is_automatic_gain_control: bool,
    automatic_gain_control_decay_factor: f64,

    input_series: Vec<f64>,
    input_series_minus_mean: Vec<f64>,
    spectrum: Vec<f64>,
    period: Vec<f64>,

    // Pre-computed trig tables.
    frequency_sin: Vec<f64>,  // first-order only
    frequency_cos: Vec<f64>,  // first-order only
    frequency_cos2: Vec<f64>, // second-order only

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
        is_first_order: bool,
        is_spectral_dilation_compensation: bool,
        is_automatic_gain_control: bool,
        automatic_gain_control_decay_factor: f64,
    ) -> Self {
        let two_pi = 2.0 * PI;
        let length_spectrum =
            ((max_period - min_period) * spectrum_resolution as f64) as usize + 1;
        let res = spectrum_resolution as f64;

        let mut period = vec![0.0; length_spectrum];
        let mut frequency_sin = Vec::new();
        let mut frequency_cos = Vec::new();
        let mut frequency_cos2 = Vec::new();

        if is_first_order {
            frequency_sin = vec![0.0; length_spectrum];
            frequency_cos = vec![0.0; length_spectrum];

            for i in 0..length_spectrum {
                let p = max_period - i as f64 / res;
                period[i] = p;
                let theta = two_pi / p;
                frequency_sin[i] = theta.sin();
                frequency_cos[i] = theta.cos();
            }
        } else {
            frequency_cos2 = vec![0.0; length_spectrum];

            for i in 0..length_spectrum {
                let p = max_period - i as f64 / res;
                period[i] = p;
                frequency_cos2[i] = 2.0 * (two_pi / p).cos();
            }
        }

        Self {
            length,
            length_spectrum,
            is_first_order,
            is_spectral_dilation_compensation,
            is_automatic_gain_control,
            automatic_gain_control_decay_factor,
            input_series: vec![0.0; length],
            input_series_minus_mean: vec![0.0; length],
            spectrum: vec![0.0; length_spectrum],
            period,
            frequency_sin,
            frequency_cos,
            frequency_cos2,
            spectrum_min: 0.0,
            spectrum_max: 0.0,
            previous_spectrum_max: 0.0,
        }
    }

    fn calculate(&mut self) {
        // Subtract the mean.
        let mut mean = 0.0;
        for i in 0..self.length {
            mean += self.input_series[i];
        }
        mean /= self.length as f64;

        for i in 0..self.length {
            self.input_series_minus_mean[i] = self.input_series[i] - mean;
        }

        // Seed with the first bin.
        let mut spectrum = self.goertzel_estimate(0);
        if self.is_spectral_dilation_compensation {
            spectrum /= self.period[0];
        }
        self.spectrum[0] = spectrum;
        self.spectrum_min = spectrum;

        if self.is_automatic_gain_control {
            self.spectrum_max =
                self.automatic_gain_control_decay_factor * self.previous_spectrum_max;
            if self.spectrum_max < spectrum {
                self.spectrum_max = spectrum;
            }
        } else {
            self.spectrum_max = spectrum;
        }

        for i in 1..self.length_spectrum {
            spectrum = self.goertzel_estimate(i);
            if self.is_spectral_dilation_compensation {
                spectrum /= self.period[i];
            }
            self.spectrum[i] = spectrum;

            if self.spectrum_max < spectrum {
                self.spectrum_max = spectrum;
            } else if self.spectrum_min > spectrum {
                self.spectrum_min = spectrum;
            }
        }

        self.previous_spectrum_max = self.spectrum_max;
    }

    fn goertzel_estimate(&self, j: usize) -> f64 {
        if self.is_first_order {
            self.goertzel_first_order_estimate(j)
        } else {
            self.goertzel_second_order_estimate(j)
        }
    }

    fn goertzel_second_order_estimate(&self, j: usize) -> f64 {
        let cos2 = self.frequency_cos2[j];
        let mut s1 = 0.0_f64;
        let mut s2 = 0.0_f64;

        for i in 0..self.length {
            let s0 = self.input_series_minus_mean[i] + cos2 * s1 - s2;
            s2 = s1;
            s1 = s0;
        }

        let spectrum = s1 * s1 + s2 * s2 - cos2 * s1 * s2;
        if spectrum < 0.0 {
            return 0.0;
        }
        spectrum
    }

    fn goertzel_first_order_estimate(&self, j: usize) -> f64 {
        let cos_theta = self.frequency_cos[j];
        let sin_theta = self.frequency_sin[j];
        let mut yre = 0.0_f64;
        let mut yim = 0.0_f64;

        for i in 0..self.length {
            let re = self.input_series_minus_mean[i] + cos_theta * yre - sin_theta * yim;
            let im = self.input_series_minus_mean[i] + cos_theta * yim + sin_theta * yre;
            yre = re;
            yim = im;
        }

        yre * yre + yim * yim
    }
}

// ---------------------------------------------------------------------------
// GoertzelSpectrum indicator
// ---------------------------------------------------------------------------

/// MBST's Goertzel Spectrum heatmap indicator.
///
/// Displays a power heatmap of cyclic activity over a configurable cycle-period
/// range using the Goertzel algorithm.
pub struct GoertzelSpectrum {
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

impl GoertzelSpectrum {
    /// Creates an instance with default parameters.
    pub fn default_params() -> Result<Self, String> {
        Self::new(&GoertzelSpectrumParams::default())
    }

    /// Creates an instance with the supplied parameters.
    pub fn new(p: &GoertzelSpectrumParams) -> Result<Self, String> {
        let invalid = "invalid goertzel spectrum parameters";

        let def_length: usize = 64;
        let def_min_period: f64 = 2.0;
        let def_max_period: f64 = 64.0;
        let def_spectrum_resolution: usize = 1;
        let def_agc_decay: f64 = 0.991;
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

        let flags = build_flag_tags(
            p.is_first_order,
            sdc_on,
            agc_on,
            agc_decay,
            floating_norm,
            def_agc_decay,
            agc_decay_epsilon,
        );

        let mnemonic = format!(
            "gspect({}, {}, {}, {}{}{})",
            length, min_period, max_period, spectrum_resolution, flags, component_mnemonic
        );
        let description = format!("Goertzel spectrum {}", mnemonic);

        let estimator = Estimator::new(
            length,
            min_period,
            max_period,
            spectrum_resolution,
            p.is_first_order,
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
    is_first_order: bool,
    sdc_on: bool,
    agc_on: bool,
    agc_decay: f64,
    floating_norm: bool,
    def_agc: f64,
    eps: f64,
) -> String {
    let mut s = String::new();

    if is_first_order {
        s.push_str(", fo");
    }
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

impl Indicator for GoertzelSpectrum {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::GoertzelSpectrum,
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
    use super::super::testdata::testdata;
    use crate::indicators::core::outputs::shape::Shape;

    fn test_gs_time() -> i64 {
        0
    }
    const TOLERANCE: f64 = 1e-10;
    const MIN_MAX_TOL: f64 = 1e-9;
    #[test]
    fn test_goertzel_spectrum_update() {
        let input = testdata::test_gs_input();
        let t0 = test_gs_time();

        let mut x = GoertzelSpectrum::default_params().unwrap();
        let snaps = testdata::goertzel_snapshots();
        let mut si = 0;

        for i in 0..input.len() {
            let h = x.update(input[i], t0 + i as i64);

            assert_eq!(h.parameter_first, 2.0, "[{}] parameter_first", i);
            assert_eq!(h.parameter_last, 64.0, "[{}] parameter_last", i);
            assert_eq!(h.parameter_resolution, 1.0, "[{}] parameter_resolution", i);

            if !x.is_primed() {
                assert!(h.is_empty(), "[{}] expected empty heatmap before priming", i);
                continue;
            }

            assert_eq!(h.values.len(), 63, "[{}] values len", i);

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
    fn test_goertzel_spectrum_primes_at_bar_63() {
        let mut x = GoertzelSpectrum::default_params().unwrap();
        assert!(!x.is_primed());

        let input = testdata::test_gs_input();
        let t0 = test_gs_time();
        let mut primed_at: Option<usize> = None;

        for i in 0..input.len() {
            x.update(input[i], t0 + i as i64);
            if x.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert_eq!(primed_at, Some(63));
    }

    #[test]
    fn test_goertzel_spectrum_nan_input() {
        let mut x = GoertzelSpectrum::default_params().unwrap();

        let h = x.update(f64::NAN, test_gs_time());
        assert!(h.is_empty());
        assert!(!x.is_primed());
    }

    #[test]
    fn test_goertzel_spectrum_metadata() {
        let x = GoertzelSpectrum::default_params().unwrap();
        let md = x.metadata();

        let mn = "gspect(64, 2, 64, 1, hl/2)";

        assert_eq!(md.identifier, Identifier::GoertzelSpectrum);
        assert_eq!(md.mnemonic, mn);
        assert_eq!(md.description, format!("Goertzel spectrum {}", mn));
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].kind, GoertzelSpectrumOutput::Value as i32);
        assert_eq!(md.outputs[0].shape, Shape::Heatmap);
        assert_eq!(md.outputs[0].mnemonic, mn);
    }

    #[test]
    fn test_goertzel_spectrum_mnemonic_flags() {
        let cases: Vec<(GoertzelSpectrumParams, &str)> = vec![
            (GoertzelSpectrumParams::default(), "gspect(64, 2, 64, 1, hl/2)"),
            (
                GoertzelSpectrumParams { is_first_order: true, ..Default::default() },
                "gspect(64, 2, 64, 1, fo, hl/2)",
            ),
            (
                GoertzelSpectrumParams {
                    disable_spectral_dilation_compensation: true,
                    ..Default::default()
                },
                "gspect(64, 2, 64, 1, no-sdc, hl/2)",
            ),
            (
                GoertzelSpectrumParams {
                    disable_automatic_gain_control: true,
                    ..Default::default()
                },
                "gspect(64, 2, 64, 1, no-agc, hl/2)",
            ),
            (
                GoertzelSpectrumParams {
                    automatic_gain_control_decay_factor: 0.8,
                    ..Default::default()
                },
                "gspect(64, 2, 64, 1, agc=0.8, hl/2)",
            ),
            (
                GoertzelSpectrumParams { fixed_normalization: true, ..Default::default() },
                "gspect(64, 2, 64, 1, no-fn, hl/2)",
            ),
            (
                GoertzelSpectrumParams {
                    is_first_order: true,
                    disable_spectral_dilation_compensation: true,
                    disable_automatic_gain_control: true,
                    fixed_normalization: true,
                    ..Default::default()
                },
                "gspect(64, 2, 64, 1, fo, no-sdc, no-agc, no-fn, hl/2)",
            ),
        ];

        for (p, expected_mn) in &cases {
            let x = GoertzelSpectrum::new(p).unwrap();
            assert_eq!(x.mnemonic(), *expected_mn, "params produced wrong mnemonic");
        }
    }

    #[test]
    fn test_goertzel_spectrum_validation() {
        let cases: Vec<(GoertzelSpectrumParams, &str)> = vec![
            (
                GoertzelSpectrumParams {
                    length: 1, min_period: 2.0, max_period: 64.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid goertzel spectrum parameters: Length should be >= 2",
            ),
            (
                GoertzelSpectrumParams {
                    length: 64, min_period: 1.0, max_period: 64.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid goertzel spectrum parameters: MinPeriod should be >= 2",
            ),
            (
                GoertzelSpectrumParams {
                    length: 64, min_period: 10.0, max_period: 10.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid goertzel spectrum parameters: MaxPeriod should be > MinPeriod",
            ),
            (
                GoertzelSpectrumParams {
                    length: 16, min_period: 2.0, max_period: 64.0, spectrum_resolution: 1,
                    ..Default::default()
                },
                "invalid goertzel spectrum parameters: MaxPeriod should be <= 2 * Length",
            ),
            (
                GoertzelSpectrumParams {
                    automatic_gain_control_decay_factor: -0.1,
                    ..Default::default()
                },
                "invalid goertzel spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
            ),
            (
                GoertzelSpectrumParams {
                    automatic_gain_control_decay_factor: 1.0,
                    ..Default::default()
                },
                "invalid goertzel spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
            ),
        ];

        for (p, expected_msg) in &cases {
            match GoertzelSpectrum::new(p) {
                Ok(_) => panic!("expected error for: {}", expected_msg),
                Err(e) => assert_eq!(e, *expected_msg),
            }
        }
    }

    #[test]
    fn test_goertzel_spectrum_update_entity() {
        let input = testdata::test_gs_input();
        let t0 = test_gs_time();
        let prime_count = 70;

        // Test update_scalar
        {
            let mut x = GoertzelSpectrum::default_params().unwrap();
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
            let mut x = GoertzelSpectrum::default_params().unwrap();
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
            let mut x = GoertzelSpectrum::default_params().unwrap();
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
            let mut x = GoertzelSpectrum::default_params().unwrap();
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
        assert_eq!(GoertzelSpectrumOutput::Value.as_str(), "value");
        assert!(GoertzelSpectrumOutput::Value.is_known());
    }
}
