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
pub enum CombBandPassSpectrumOutput {
    Value = 1,
}

impl CombBandPassSpectrumOutput {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Value => "value",
        }
    }

    pub fn is_known(&self) -> bool {
        matches!(self, Self::Value)
    }
}

impl std::fmt::Display for CombBandPassSpectrumOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

pub struct CombBandPassSpectrumParams {
    pub min_period: i32,
    pub max_period: i32,
    pub bandwidth: f64,
    pub disable_spectral_dilation_compensation: bool,
    pub disable_automatic_gain_control: bool,
    pub automatic_gain_control_decay_factor: f64,
    pub fixed_normalization: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for CombBandPassSpectrumParams {
    fn default() -> Self {
        Self {
            min_period: 0,
            max_period: 0,
            bandwidth: 0.0,
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
    min_period: usize,
    max_period: usize,
    length_spectrum: usize,
    is_spectral_dilation_compensation: bool,
    is_automatic_gain_control: bool,
    automatic_gain_control_decay_factor: f64,

    // Pre-filter coefficients.
    coeff_hp0: f64,
    coeff_hp1: f64,
    coeff_hp2: f64,
    ss_c1: f64,
    ss_c2: f64,
    ss_c3: f64,

    // Per-bin band-pass coefficients.
    periods: Vec<usize>,
    beta: Vec<f64>,
    alpha: Vec<f64>,
    comp: Vec<f64>,

    // Pre-filter state.
    close0: f64,
    close1: f64,
    close2: f64,
    hp0: f64,
    hp1: f64,
    hp2: f64,
    filt0: f64,
    filt1: f64,
    filt2: f64,

    // Band-pass filter state: bp[i][m] for bin i, lag m.
    bp: Vec<Vec<f64>>,

    // Raw spectrum values.
    spectrum: Vec<f64>,
    spectrum_min: f64,
    spectrum_max: f64,
    previous_spectrum_max: f64,
}

impl Estimator {
    fn new(
        min_period: usize,
        max_period: usize,
        bandwidth: f64,
        is_sdc: bool,
        is_agc: bool,
        agc_decay: f64,
    ) -> Self {
        let two_pi = 2.0 * PI;
        let length_spectrum = max_period - min_period + 1;

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

        let mut periods = vec![0usize; length_spectrum];
        let mut beta_vec = vec![0.0; length_spectrum];
        let mut alpha_vec = vec![0.0; length_spectrum];
        let mut comp_vec = vec![0.0; length_spectrum];
        let mut bp = Vec::with_capacity(length_spectrum);

        for i in 0..length_spectrum {
            let n = min_period + i;
            let b = (two_pi / n as f64).cos();
            let gamma = 1.0 / (two_pi * bandwidth / n as f64).cos();
            let a = gamma - (gamma * gamma - 1.0).sqrt();

            periods[i] = n;
            beta_vec[i] = b;
            alpha_vec[i] = a;
            comp_vec[i] = if is_sdc { n as f64 } else { 1.0 };

            bp.push(vec![0.0; max_period]);
        }

        Self {
            min_period,
            max_period,
            length_spectrum,
            is_spectral_dilation_compensation: is_sdc,
            is_automatic_gain_control: is_agc,
            automatic_gain_control_decay_factor: agc_decay,
            coeff_hp0: c_hp0,
            coeff_hp1: c_hp1,
            coeff_hp2: c_hp2,
            ss_c1,
            ss_c2,
            ss_c3,
            periods,
            beta: beta_vec,
            alpha: alpha_vec,
            comp: comp_vec,
            close0: 0.0,
            close1: 0.0,
            close2: 0.0,
            hp0: 0.0,
            hp1: 0.0,
            hp2: 0.0,
            filt0: 0.0,
            filt1: 0.0,
            filt2: 0.0,
            bp,
            spectrum: vec![0.0; length_spectrum],
            spectrum_min: 0.0,
            spectrum_max: 0.0,
            previous_spectrum_max: 0.0,
        }
    }

    fn update(&mut self, sample: f64) {
        // Shift close history.
        self.close2 = self.close1;
        self.close1 = self.close0;
        self.close0 = sample;

        // Shift HP history and compute new HP.
        self.hp2 = self.hp1;
        self.hp1 = self.hp0;
        self.hp0 = self.coeff_hp0 * (self.close0 - 2.0 * self.close1 + self.close2)
            + self.coeff_hp1 * self.hp1
            - self.coeff_hp2 * self.hp2;

        // Shift Filt history and compute new Filt (SuperSmoother on HP).
        self.filt2 = self.filt1;
        self.filt1 = self.filt0;
        self.filt0 =
            self.ss_c1 * (self.hp0 + self.hp1) / 2.0 + self.ss_c2 * self.filt1 + self.ss_c3 * self.filt2;

        let diff_filt = self.filt0 - self.filt2;

        // AGC seeds the running max with the decayed previous max.
        self.spectrum_min = f64::MAX;
        if self.is_automatic_gain_control {
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max;
        } else {
            self.spectrum_max = f64::MIN;
        }

        for i in 0..self.length_spectrum {
            let bp_row = &mut self.bp[i];

            // Rightward shift.
            for m in (1..self.max_period).rev() {
                bp_row[m] = bp_row[m - 1];
            }

            let a = self.alpha[i];
            let b = self.beta[i];
            bp_row[0] = 0.5 * (1.0 - a) * diff_filt + b * (1.0 + a) * bp_row[1] - a * bp_row[2];

            // Power = sum over m in [0..N) of (BP[i,m] / Comp[i])^2.
            let n = self.periods[i];
            let c = self.comp[i];
            let mut pwr = 0.0;

            for m in 0..n {
                let v = bp_row[m] / c;
                pwr += v * v;
            }

            self.spectrum[i] = pwr;

            if self.spectrum_max < pwr {
                self.spectrum_max = pwr;
            }
            if self.spectrum_min > pwr {
                self.spectrum_min = pwr;
            }
        }

        self.previous_spectrum_max = self.spectrum_max;
    }
}

// ---------------------------------------------------------------------------
// CombBandPassSpectrum indicator
// ---------------------------------------------------------------------------

pub struct CombBandPassSpectrum {
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

impl CombBandPassSpectrum {
    pub fn default_params() -> Result<Self, String> {
        Self::new(&CombBandPassSpectrumParams::default())
    }

    pub fn new(p: &CombBandPassSpectrumParams) -> Result<Self, String> {
        let invalid = "invalid comb band-pass spectrum parameters";

        let def_min_period: i32 = 10;
        let def_max_period: i32 = 48;
        let def_bandwidth: f64 = 0.3;
        let def_agc_decay: f64 = 0.995;
        let agc_decay_epsilon: f64 = 1e-12;
        let bandwidth_epsilon: f64 = 1e-12;

        let min_period = if p.min_period == 0 { def_min_period } else { p.min_period };
        let max_period = if p.max_period == 0 { def_max_period } else { p.max_period };
        let bandwidth = if p.bandwidth == 0.0 { def_bandwidth } else { p.bandwidth };
        let agc_decay = if p.automatic_gain_control_decay_factor == 0.0 {
            def_agc_decay
        } else {
            p.automatic_gain_control_decay_factor
        };

        let sdc_on = !p.disable_spectral_dilation_compensation;
        let agc_on = !p.disable_automatic_gain_control;
        let floating_norm = !p.fixed_normalization;

        if min_period < 2 {
            return Err(format!("{}: MinPeriod should be >= 2", invalid));
        }
        if max_period <= min_period {
            return Err(format!("{}: MaxPeriod should be > MinPeriod", invalid));
        }
        if bandwidth <= 0.0 || bandwidth >= 1.0 {
            return Err(format!("{}: Bandwidth should be in (0, 1)", invalid));
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
            bandwidth, sdc_on, agc_on, agc_decay, floating_norm,
            def_bandwidth, def_agc_decay, bandwidth_epsilon, agc_decay_epsilon,
        );

        let mnemonic = format!(
            "cbps({}, {}{}{})",
            min_period, max_period, flags, component_mnemonic
        );
        let description = format!("Comb band-pass spectrum {}", mnemonic);

        let estimator = Estimator::new(
            min_period as usize,
            max_period as usize,
            bandwidth,
            sdc_on,
            agc_on,
            agc_decay,
        );

        Ok(Self {
            mnemonic,
            description,
            estimator,
            window_count: 0,
            prime_count: max_period as usize,
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

        let max_ref = self.estimator.spectrum_max;
        let spectrum_range = max_ref - min_ref;

        let mut values = vec![0.0; length_spectrum];
        let mut value_min = f64::INFINITY;
        let mut value_max = f64::NEG_INFINITY;

        // Spectrum is already in axis order (bin 0 = MinPeriod, bin last = MaxPeriod).
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
    bandwidth: f64,
    sdc_on: bool,
    agc_on: bool,
    agc_decay: f64,
    floating_norm: bool,
    def_bandwidth: f64,
    def_agc: f64,
    bw_eps: f64,
    agc_eps: f64,
) -> String {
    let mut s = String::new();

    if (bandwidth - def_bandwidth).abs() > bw_eps {
        s.push_str(&format!(", bw={}", bandwidth));
    }
    if !sdc_on {
        s.push_str(", no-sdc");
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

impl Indicator for CombBandPassSpectrum {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CombBandPassSpectrum,
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

    fn test_cbps_time() -> i64 {
        0
    }
    const TOLERANCE: f64 = 1e-12;
    const MIN_MAX_TOL: f64 = 1e-10;
    #[test]
    fn test_cbps_update() {
        let input = testdata::test_cbps_input();
        let t0 = test_cbps_time();

        let mut x = CombBandPassSpectrum::default_params().unwrap();
        let snaps = testdata::cbps_snapshots();
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
    fn test_cbps_primes_at_bar_47() {
        let mut x = CombBandPassSpectrum::default_params().unwrap();
        assert!(!x.is_primed());

        let input = testdata::test_cbps_input();
        let t0 = test_cbps_time();
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
    fn test_cbps_nan_input() {
        let mut x = CombBandPassSpectrum::default_params().unwrap();

        let h = x.update(f64::NAN, test_cbps_time());
        assert!(h.is_empty());
        assert!(!x.is_primed());
    }

    #[test]
    fn test_cbps_synthetic_sine() {
        let period: f64 = 20.0;
        let bars = 400;

        let mut x = CombBandPassSpectrum::new(&CombBandPassSpectrumParams {
            disable_spectral_dilation_compensation: true,
            disable_automatic_gain_control: true,
            fixed_normalization: true,
            ..Default::default()
        })
        .unwrap();

        let t0 = test_cbps_time();
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
    fn test_cbps_metadata() {
        let x = CombBandPassSpectrum::default_params().unwrap();
        let md = x.metadata();

        let mn = "cbps(10, 48, hl/2)";

        assert_eq!(md.identifier, Identifier::CombBandPassSpectrum);
        assert_eq!(md.mnemonic, mn);
        assert_eq!(md.description, format!("Comb band-pass spectrum {}", mn));
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].kind, CombBandPassSpectrumOutput::Value as i32);
        assert_eq!(md.outputs[0].shape, Shape::Heatmap);
        assert_eq!(md.outputs[0].mnemonic, mn);
    }

    #[test]
    fn test_cbps_mnemonic_flags() {
        let cases: Vec<(CombBandPassSpectrumParams, &str)> = vec![
            (CombBandPassSpectrumParams::default(), "cbps(10, 48, hl/2)"),
            (
                CombBandPassSpectrumParams {
                    bandwidth: 0.5,
                    ..Default::default()
                },
                "cbps(10, 48, bw=0.5, hl/2)",
            ),
            (
                CombBandPassSpectrumParams {
                    disable_spectral_dilation_compensation: true,
                    ..Default::default()
                },
                "cbps(10, 48, no-sdc, hl/2)",
            ),
            (
                CombBandPassSpectrumParams {
                    disable_automatic_gain_control: true,
                    ..Default::default()
                },
                "cbps(10, 48, no-agc, hl/2)",
            ),
            (
                CombBandPassSpectrumParams {
                    automatic_gain_control_decay_factor: 0.8,
                    ..Default::default()
                },
                "cbps(10, 48, agc=0.8, hl/2)",
            ),
            (
                CombBandPassSpectrumParams {
                    fixed_normalization: true,
                    ..Default::default()
                },
                "cbps(10, 48, no-fn, hl/2)",
            ),
            (
                CombBandPassSpectrumParams {
                    bandwidth: 0.5,
                    disable_spectral_dilation_compensation: true,
                    disable_automatic_gain_control: true,
                    fixed_normalization: true,
                    ..Default::default()
                },
                "cbps(10, 48, bw=0.5, no-sdc, no-agc, no-fn, hl/2)",
            ),
        ];

        for (p, expected_mn) in &cases {
            let x = CombBandPassSpectrum::new(p).unwrap();
            assert_eq!(x.mnemonic(), *expected_mn, "params produced wrong mnemonic");
        }
    }

    #[test]
    fn test_cbps_validation() {
        let cases: Vec<(CombBandPassSpectrumParams, &str)> = vec![
            (
                CombBandPassSpectrumParams {
                    min_period: 1, max_period: 48, bandwidth: 0.3,
                    ..Default::default()
                },
                "invalid comb band-pass spectrum parameters: MinPeriod should be >= 2",
            ),
            (
                CombBandPassSpectrumParams {
                    min_period: 10, max_period: 10, bandwidth: 0.3,
                    ..Default::default()
                },
                "invalid comb band-pass spectrum parameters: MaxPeriod should be > MinPeriod",
            ),
            (
                CombBandPassSpectrumParams {
                    bandwidth: -0.1,
                    ..Default::default()
                },
                "invalid comb band-pass spectrum parameters: Bandwidth should be in (0, 1)",
            ),
            (
                CombBandPassSpectrumParams {
                    bandwidth: 1.0,
                    ..Default::default()
                },
                "invalid comb band-pass spectrum parameters: Bandwidth should be in (0, 1)",
            ),
            (
                CombBandPassSpectrumParams {
                    automatic_gain_control_decay_factor: -0.1,
                    ..Default::default()
                },
                "invalid comb band-pass spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
            ),
            (
                CombBandPassSpectrumParams {
                    automatic_gain_control_decay_factor: 1.0,
                    ..Default::default()
                },
                "invalid comb band-pass spectrum parameters: AutomaticGainControlDecayFactor should be in (0, 1)",
            ),
        ];

        for (p, expected_msg) in &cases {
            match CombBandPassSpectrum::new(p) {
                Ok(_) => panic!("expected error for: {}", expected_msg),
                Err(e) => assert_eq!(e, *expected_msg),
            }
        }
    }

    #[test]
    fn test_cbps_update_entity() {
        let input = testdata::test_cbps_input();
        let t0 = test_cbps_time();
        let prime_count = 60;

        // Test update_scalar
        {
            let mut x = CombBandPassSpectrum::default_params().unwrap();
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
            let mut x = CombBandPassSpectrum::default_params().unwrap();
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
            let mut x = CombBandPassSpectrum::default_params().unwrap();
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
            let mut x = CombBandPassSpectrum::default_params().unwrap();
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
        assert_eq!(CombBandPassSpectrumOutput::Value.as_str(), "value");
        assert!(CombBandPassSpectrumOutput::Value.is_known());
    }
}
