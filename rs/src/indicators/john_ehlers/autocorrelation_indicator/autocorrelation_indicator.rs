use std::f64::consts::PI;

use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::{BarFunc, QuoteFunc, TradeFunc};
use crate::indicators::core::metadata::Metadata;
use crate::indicators::core::outputs::heatmap::Heatmap;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

pub struct AutoCorrelationIndicatorParams {
    pub min_lag: i32,
    pub max_lag: i32,
    pub smoothing_period: i32,
    pub averaging_length: i32,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for AutoCorrelationIndicatorParams {
    fn default() -> Self {
        Self {
            min_lag: 3,
            max_lag: 48,
            smoothing_period: 10,
            averaging_length: 0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output enum
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AutoCorrelationIndicatorOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Estimator
// ---------------------------------------------------------------------------

#[derive(Debug)]
struct Estimator {
    min_lag: i32,
    max_lag: i32,
    averaging_length: i32,
    length_spectrum: usize,
    filt_buffer_len: usize,

    coeff_hp0: f64,
    coeff_hp1: f64,
    coeff_hp2: f64,
    ss_c1: f64,
    ss_c2: f64,
    ss_c3: f64,

    close0: f64, close1: f64, close2: f64,
    hp0: f64, hp1: f64, hp2: f64,

    filt: Vec<f64>,
    spectrum: Vec<f64>,
}

impl Estimator {
    fn new(min_lag: i32, max_lag: i32, smoothing_period: i32, averaging_length: i32) -> Self {
        let two_pi = 2.0 * PI;
        let length_spectrum = (max_lag - min_lag + 1) as usize;

        let m_max = if averaging_length == 0 { max_lag } else { averaging_length };
        let filt_buffer_len = (max_lag + m_max) as usize;

        let omega_hp = 0.707 * two_pi / max_lag as f64;
        let alpha_hp = (omega_hp.cos() + omega_hp.sin() - 1.0) / omega_hp.cos();
        let c_hp0 = (1.0 - alpha_hp / 2.0) * (1.0 - alpha_hp / 2.0);
        let c_hp1 = 2.0 * (1.0 - alpha_hp);
        let c_hp2 = (1.0 - alpha_hp) * (1.0 - alpha_hp);

        let a1 = (-1.414 * PI / smoothing_period as f64).exp();
        let b1 = 2.0 * a1 * (1.414 * PI / smoothing_period as f64).cos();
        let ss_c2 = b1;
        let ss_c3 = -a1 * a1;
        let ss_c1 = 1.0 - ss_c2 - ss_c3;

        Self {
            min_lag, max_lag, averaging_length,
            length_spectrum, filt_buffer_len,
            coeff_hp0: c_hp0, coeff_hp1: c_hp1, coeff_hp2: c_hp2,
            ss_c1, ss_c2, ss_c3,
            close0: 0.0, close1: 0.0, close2: 0.0,
            hp0: 0.0, hp1: 0.0, hp2: 0.0,
            filt: vec![0.0; filt_buffer_len],
            spectrum: vec![0.0; length_spectrum],
        }
    }

    fn update(&mut self, sample: f64) {
        self.close2 = self.close1;
        self.close1 = self.close0;
        self.close0 = sample;

        self.hp2 = self.hp1;
        self.hp1 = self.hp0;
        self.hp0 = self.coeff_hp0 * (self.close0 - 2.0 * self.close1 + self.close2)
            + self.coeff_hp1 * self.hp1
            - self.coeff_hp2 * self.hp2;

        for k in (1..self.filt_buffer_len).rev() {
            self.filt[k] = self.filt[k - 1];
        }

        self.filt[0] = self.ss_c1 * (self.hp0 + self.hp1) / 2.0
            + self.ss_c2 * self.filt[1]
            + self.ss_c3 * self.filt[2];

        for i in 0..self.length_spectrum {
            let lag = self.min_lag + i as i32;
            let m = if self.averaging_length == 0 { lag } else { self.averaging_length };

            let mut sx = 0.0;
            let mut sy = 0.0;
            let mut sxx = 0.0;
            let mut syy = 0.0;
            let mut sxy = 0.0;

            for c in 0..m as usize {
                let x = self.filt[c];
                let y = self.filt[lag as usize + c];
                sx += x;
                sy += y;
                sxx += x * x;
                syy += y * y;
                sxy += x * y;
            }

            let mf = m as f64;
            let denom = (mf * sxx - sx * sx) * (mf * syy - sy * sy);

            let r = if denom > 0.0 {
                (mf * sxy - sx * sy) / denom.sqrt()
            } else {
                0.0
            };

            self.spectrum[i] = 0.5 * (r + 1.0);
        }
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

#[derive(Debug)]
pub struct AutoCorrelationIndicator {
    mnemonic: String,
    description: String,
    estimator: Estimator,
    window_count: usize,
    prime_count: usize,
    primed: bool,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
}

impl AutoCorrelationIndicator {
    pub fn new(params: &AutoCorrelationIndicatorParams) -> Result<Self, String> {
        let invalid = "invalid autocorrelation indicator parameters";

        let mut min_lag = params.min_lag;
        let mut max_lag = params.max_lag;
        let mut smoothing_period = params.smoothing_period;
        let averaging_length = params.averaging_length;

        if min_lag == 0 { min_lag = 3; }
        if max_lag == 0 { max_lag = 48; }
        if smoothing_period == 0 { smoothing_period = 10; }

        if min_lag < 1 {
            return Err(format!("{}: MinLag should be >= 1", invalid));
        }
        if max_lag <= min_lag {
            return Err(format!("{}: MaxLag should be > MinLag", invalid));
        }
        if smoothing_period < 2 {
            return Err(format!("{}: SmoothingPeriod should be >= 2", invalid));
        }
        if averaging_length < 0 {
            return Err(format!("{}: AveragingLength should be >= 0", invalid));
        }

        // Default BarMedianPrice for Ehlers.
        let bc = params.bar_component.unwrap_or(BarComponent::Median);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let component_mnemonic = component_triple_mnemonic(bc, qc, tc);

        let mut flags = String::new();
        if averaging_length != 0 {
            flags.push_str(&format!(", average={}", averaging_length));
        }

        let mnemonic = format!("aci({}, {}, {}{}{})", min_lag, max_lag, smoothing_period, flags, component_mnemonic);
        let description = format!("Autocorrelation indicator {}", mnemonic);

        let est = Estimator::new(min_lag, max_lag, smoothing_period, averaging_length);
        let prime_count = est.filt_buffer_len;

        Ok(Self {
            mnemonic,
            description,
            prime_count,
            estimator: est,
            window_count: 0,
            primed: false,
            min_parameter_value: min_lag as f64,
            max_parameter_value: max_lag as f64,
            parameter_resolution: 1.0,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    pub fn update(&mut self, sample: f64, time: i64) -> Heatmap {
        if sample.is_nan() {
            return Heatmap::empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution);
        }

        self.estimator.update(sample);

        if !self.primed {
            self.window_count += 1;
            if self.window_count >= self.prime_count {
                self.primed = true;
            } else {
                return Heatmap::empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution);
            }
        }

        let length_spectrum = self.estimator.length_spectrum;
        let mut values = vec![0.0; length_spectrum];
        let mut value_min = f64::INFINITY;
        let mut value_max = f64::NEG_INFINITY;

        for i in 0..length_spectrum {
            let v = self.estimator.spectrum[i];
            values[i] = v;
            if v < value_min { value_min = v; }
            if v > value_max { value_max = v; }
        }

        Heatmap::new(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution,
            value_min, value_max, values)
    }
}

impl Indicator for AutoCorrelationIndicator {
    fn is_primed(&self) -> bool { self.primed }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::AutoCorrelationIndicator,
            &self.mnemonic,
            &self.description,
            &[OutputText { mnemonic: self.mnemonic.clone(), description: self.description.clone() }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let h = self.update(sample.value, sample.time);
        vec![Box::new(h)]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        let h = self.update(v, sample.time);
        vec![Box::new(h)]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        let h = self.update(v, sample.time);
        vec![Box::new(h)]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        let h = self.update(v, sample.time);
        vec![Box::new(h)]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    use crate::indicators::core::outputs::shape::Shape;
    const TOL: f64 = 1e-12;
    const MINMAX_TOL: f64 = 1e-10;
    #[test]
    fn test_update() {
        let input = testdata::test_input();
        let mut x = AutoCorrelationIndicator::new(&AutoCorrelationIndicatorParams::default()).unwrap();
        let snaps = testdata::snapshots();
        let mut si = 0;

        for (i, &v) in input.iter().enumerate() {
            let h = x.update(v, i as i64);
            assert_eq!(h.parameter_first, 3.0);
            assert_eq!(h.parameter_last, 48.0);
            assert_eq!(h.parameter_resolution, 1.0);

            if !x.is_primed() {
                assert!(h.is_empty(), "[{}] expected empty before priming", i);
                continue;
            }
            assert_eq!(h.values.len(), 46, "[{}] values len", i);

            if si < snaps.len() && snaps[si].i == i {
                let snap = &snaps[si];
                assert!((h.value_min - snap.value_min).abs() < MINMAX_TOL, "[{}] ValueMin: expected {}, got {}", i, snap.value_min, h.value_min);
                assert!((h.value_max - snap.value_max).abs() < MINMAX_TOL, "[{}] ValueMax: expected {}, got {}", i, snap.value_max, h.value_max);
                for &(bin, expected) in &snap.spots {
                    assert!((h.values[bin] - expected).abs() < TOL, "[{}] Values[{}]: expected {}, got {}", i, bin, expected, h.values[bin]);
                }
                si += 1;
            }
        }
        assert_eq!(si, snaps.len(), "did not hit all snapshots");
    }

    #[test]
    fn test_synthetic_sine() {
        let period = 35.0_f64;
        let bars = 600;
        let mut x = AutoCorrelationIndicator::new(&AutoCorrelationIndicatorParams::default()).unwrap();
        let mut last = None;
        for i in 0..bars {
            let sample = 100.0 + (2.0 * PI * i as f64 / period).sin();
            last = Some(x.update(sample, i as i64));
        }
        let h = last.unwrap();
        assert!(!h.is_empty());
        let peak_bin = h.values.iter().enumerate().max_by(|a, b| a.1.partial_cmp(b.1).unwrap()).unwrap().0;
        let expected_bin = (period - h.parameter_first) as usize;
        assert_eq!(peak_bin, expected_bin);
    }

    #[test]
    fn test_nan_input() {
        let mut x = AutoCorrelationIndicator::new(&AutoCorrelationIndicatorParams::default()).unwrap();
        let h = x.update(f64::NAN, 0);
        assert!(h.is_empty());
        assert!(!x.is_primed());
    }

    #[test]
    fn test_metadata() {
        let x = AutoCorrelationIndicator::new(&AutoCorrelationIndicatorParams::default()).unwrap();
        let md = x.metadata();
        let mn = "aci(3, 48, 10, hl/2)";
        assert_eq!(md.identifier, Identifier::AutoCorrelationIndicator);
        assert_eq!(md.mnemonic, mn);
        assert_eq!(md.description, format!("Autocorrelation indicator {}", mn));
        assert_eq!(md.outputs.len(), 1);
        assert_eq!(md.outputs[0].kind, AutoCorrelationIndicatorOutput::Value as i32);
        assert_eq!(md.outputs[0].shape, Shape::Heatmap);
        assert_eq!(md.outputs[0].mnemonic, mn);
    }

    #[test]
    fn test_mnemonic_flags() {
        let cases = vec![
            (AutoCorrelationIndicatorParams::default(), "aci(3, 48, 10, hl/2)"),
            (AutoCorrelationIndicatorParams { averaging_length: 5, ..Default::default() }, "aci(3, 48, 10, average=5, hl/2)"),
            (AutoCorrelationIndicatorParams { min_lag: 5, max_lag: 30, smoothing_period: 8, ..Default::default() }, "aci(5, 30, 8, hl/2)"),
        ];
        for (p, expected) in cases {
            let x = AutoCorrelationIndicator::new(&p).unwrap();
            assert_eq!(x.mnemonic, expected);
        }
    }

    #[test]
    fn test_validation() {
        let cases = vec![
            (AutoCorrelationIndicatorParams { min_lag: -1, max_lag: 48, smoothing_period: 10, ..Default::default() },
             "invalid autocorrelation indicator parameters: MinLag should be >= 1"),
            (AutoCorrelationIndicatorParams { min_lag: 10, max_lag: 10, smoothing_period: 10, ..Default::default() },
             "invalid autocorrelation indicator parameters: MaxLag should be > MinLag"),
            (AutoCorrelationIndicatorParams { min_lag: 3, max_lag: 48, smoothing_period: 1, ..Default::default() },
             "invalid autocorrelation indicator parameters: SmoothingPeriod should be >= 2"),
            (AutoCorrelationIndicatorParams { averaging_length: -1, ..Default::default() },
             "invalid autocorrelation indicator parameters: AveragingLength should be >= 0"),
        ];
        for (p, expected_msg) in cases {
            let err = AutoCorrelationIndicator::new(&p).unwrap_err();
            assert_eq!(err, expected_msg);
        }
    }

    #[test]
    fn test_update_entity() {
        let input = testdata::test_input();
        let prime_count = 200;

        // Scalar
        let mut x = AutoCorrelationIndicator::new(&AutoCorrelationIndicatorParams::default()).unwrap();
        for i in 0..prime_count { x.update(input[i % input.len()], 0); }
        let out = x.update_scalar(&Scalar::new(42, 100.0));
        assert_eq!(out.len(), 1);
        assert!(out[0].downcast_ref::<Heatmap>().is_some());

        // Bar
        let mut x = AutoCorrelationIndicator::new(&AutoCorrelationIndicatorParams::default()).unwrap();
        for i in 0..prime_count { x.update(input[i % input.len()], 0); }
        let out = x.update_bar(&Bar { time: 42, open: 100.0, high: 100.0, low: 100.0, close: 100.0, volume: 0.0 });
        assert_eq!(out.len(), 1);
        assert!(out[0].downcast_ref::<Heatmap>().is_some());

        // Quote
        let mut x = AutoCorrelationIndicator::new(&AutoCorrelationIndicatorParams::default()).unwrap();
        for i in 0..prime_count { x.update(input[i % input.len()], 0); }
        let out = x.update_quote(&Quote { time: 42, bid_price: 100.0, ask_price: 100.0, bid_size: 1.0, ask_size: 1.0 });
        assert_eq!(out.len(), 1);
        assert!(out[0].downcast_ref::<Heatmap>().is_some());

        // Trade
        let mut x = AutoCorrelationIndicator::new(&AutoCorrelationIndicatorParams::default()).unwrap();
        for i in 0..prime_count { x.update(input[i % input.len()], 0); }
        let out = x.update_trade(&Trade { time: 42, price: 100.0, volume: 1.0 });
        assert_eq!(out.len(), 1);
        assert!(out[0].downcast_ref::<Heatmap>().is_some());
    }
}
