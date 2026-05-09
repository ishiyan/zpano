use std::any::Any;

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
use crate::indicators::john_ehlers::corona::corona::{Corona, CoronaParams};

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const HIGH_LOW_BUFFER_SIZE: usize = 5;
const HIGH_LOW_MEDIAN_INDEX: usize = 2;
const AVERAGE_SAMPLE_ALPHA: f64 = 0.1;
const AVERAGE_SAMPLE_ONE_MINUS: f64 = 0.9;
const SIGNAL_EMA_ALPHA: f64 = 0.2;
const SIGNAL_EMA_ONE_MINUS: f64 = 0.9; // Intentional: sums to 1.1, per Ehlers.
const NOISE_EMA_ALPHA: f64 = 0.1;
const NOISE_EMA_ONE_MINUS: f64 = 0.9;
const RATIO_OFFSET_DB: f64 = 3.5;
const RATIO_UPPER_DB: f64 = 10.0;
const DB_GAIN: f64 = 20.0;
const WIDTH_LOW_RATIO_THRESHOLD: f64 = 0.5;
const WIDTH_BASELINE: f64 = 0.2;
const WIDTH_SLOPE: f64 = 0.4;
const RASTER_BLEND_EXPONENT: f64 = 0.8;
const RASTER_BLEND_HALF: f64 = 0.5;
const RASTER_NEGATIVE_ARG_CUTOFF: f64 = 1.0;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Corona Signal-to-Noise Ratio indicator.
pub struct CoronaSignalToNoiseRatioParams {
    pub raster_length: i32,
    pub max_raster_value: f64,
    pub min_parameter_value: f64,
    pub max_parameter_value: f64,
    pub high_pass_filter_cutoff: i32,
    pub minimal_period: i32,
    pub maximal_period: i32,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for CoronaSignalToNoiseRatioParams {
    fn default() -> Self {
        Self {
            raster_length: 50,
            max_raster_value: 20.0,
            min_parameter_value: 1.0,
            max_parameter_value: 11.0,
            high_pass_filter_cutoff: 30,
            minimal_period: 6,
            maximal_period: 30,
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
pub enum CoronaSignalToNoiseRatioOutput {
    Value = 1,
    SignalToNoiseRatio = 2,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Corona Signal-to-Noise Ratio heatmap indicator.
pub struct CoronaSignalToNoiseRatio {
    mnemonic: String,
    description: String,
    mnemonic_snr: String,
    description_snr: String,
    corona: Corona,
    raster_length: usize,
    raster_step: f64,
    max_raster_value: f64,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    raster: Vec<f64>,
    high_low_buffer: [f64; HIGH_LOW_BUFFER_SIZE],
    hl_sorted: [f64; HIGH_LOW_BUFFER_SIZE],
    average_sample_previous: f64,
    signal_previous: f64,
    noise_previous: f64,
    signal_to_noise_ratio: f64,
    is_started: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CoronaSignalToNoiseRatio {
    pub fn new(p: &CoronaSignalToNoiseRatioParams) -> Result<Self, String> {
        let invalid = "invalid corona signal to noise ratio parameters";

        let raster_len = if p.raster_length == 0 { 50 } else { p.raster_length };
        let max_raster = if p.max_raster_value == 0.0 { 20.0 } else { p.max_raster_value };
        let min_pv = if p.min_parameter_value == 0.0 { 1.0 } else { p.min_parameter_value };
        let max_pv = if p.max_parameter_value == 0.0 { 11.0 } else { p.max_parameter_value };
        let hp = if p.high_pass_filter_cutoff == 0 { 30 } else { p.high_pass_filter_cutoff };
        let min_per = if p.minimal_period == 0 { 6 } else { p.minimal_period };
        let max_per = if p.maximal_period == 0 { 30 } else { p.maximal_period };

        if raster_len < 2 {
            return Err(format!("{}: RasterLength should be >= 2", invalid));
        }
        if max_raster <= 0.0 {
            return Err(format!("{}: MaxRasterValue should be > 0", invalid));
        }
        if min_pv < 0.0 {
            return Err(format!("{}: MinParameterValue should be >= 0", invalid));
        }
        if max_pv <= min_pv {
            return Err(format!("{}: MaxParameterValue should be > MinParameterValue", invalid));
        }
        if hp < 2 {
            return Err(format!("{}: HighPassFilterCutoff should be >= 2", invalid));
        }
        if min_per < 2 {
            return Err(format!("{}: MinimalPeriod should be >= 2", invalid));
        }
        if max_per <= min_per {
            return Err(format!("{}: MaximalPeriod should be > MinimalPeriod", invalid));
        }

        let bc = p.bar_component.unwrap_or(BarComponent::Median);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let corona = Corona::new(&CoronaParams {
            high_pass_filter_cutoff: hp,
            minimal_period: min_per,
            maximal_period: max_per,
            ..CoronaParams::default()
        })?;

        let comp_mn = component_triple_mnemonic(bc, qc, tc);
        let parameter_resolution = (raster_len as f64 - 1.0) / (max_pv - min_pv);

        let mnemonic = format!(
            "csnr({}, {}, {}, {}, {}{})",
            raster_len, max_raster, min_pv, max_pv, hp, comp_mn
        );
        let mnemonic_snr = format!("csnr-snr({}{})", hp, comp_mn);

        Ok(Self {
            description: format!("Corona signal to noise ratio {}", mnemonic),
            mnemonic,
            description_snr: format!("Corona signal to noise ratio scalar {}", mnemonic_snr),
            mnemonic_snr,
            corona,
            raster_length: raster_len as usize,
            raster_step: max_raster / raster_len as f64,
            max_raster_value: max_raster,
            min_parameter_value: min_pv,
            max_parameter_value: max_pv,
            parameter_resolution,
            raster: vec![0.0; raster_len as usize],
            high_low_buffer: [0.0; HIGH_LOW_BUFFER_SIZE],
            hl_sorted: [0.0; HIGH_LOW_BUFFER_SIZE],
            average_sample_previous: 0.0,
            signal_previous: 0.0,
            noise_previous: 0.0,
            signal_to_noise_ratio: f64::NAN,
            is_started: false,
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
        })
    }

    /// Feed the next sample plus bar extremes.
    /// Returns (heatmap, signal_to_noise_ratio).
    pub fn update(&mut self, sample: f64, sample_low: f64, sample_high: f64, time: i64) -> (Heatmap, f64) {
        if sample.is_nan() {
            return (
                Heatmap::empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                f64::NAN,
            );
        }

        let primed = self.corona.update(sample);

        if !self.is_started {
            self.average_sample_previous = sample;
            self.high_low_buffer[HIGH_LOW_BUFFER_SIZE - 1] = sample_high - sample_low;
            self.is_started = true;
            return (
                Heatmap::empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                f64::NAN,
            );
        }

        let max_amp_sq = self.corona.maximal_amplitude_squared();

        let average_sample = AVERAGE_SAMPLE_ALPHA * sample + AVERAGE_SAMPLE_ONE_MINUS * self.average_sample_previous;
        self.average_sample_previous = average_sample;

        if average_sample.abs() > 0.0 || max_amp_sq > 0.0 {
            self.signal_previous = SIGNAL_EMA_ALPHA * max_amp_sq.sqrt() + SIGNAL_EMA_ONE_MINUS * self.signal_previous;
        }

        // Shift H-L ring buffer left; push new value.
        for i in 0..(HIGH_LOW_BUFFER_SIZE - 1) {
            self.high_low_buffer[i] = self.high_low_buffer[i + 1];
        }
        self.high_low_buffer[HIGH_LOW_BUFFER_SIZE - 1] = sample_high - sample_low;

        let mut ratio = 0.0;
        if average_sample.abs() > 0.0 {
            self.hl_sorted = self.high_low_buffer;
            self.hl_sorted.sort_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal));
            self.noise_previous = NOISE_EMA_ALPHA * self.hl_sorted[HIGH_LOW_MEDIAN_INDEX]
                + NOISE_EMA_ONE_MINUS * self.noise_previous;

            if self.noise_previous.abs() > 0.0 {
                ratio = DB_GAIN * (self.signal_previous / self.noise_previous).log10() + RATIO_OFFSET_DB;
                if ratio < 0.0 {
                    ratio = 0.0;
                } else if ratio > RATIO_UPPER_DB {
                    ratio = RATIO_UPPER_DB;
                }
                ratio /= RATIO_UPPER_DB; // ∈ [0, 1]
            }
        }

        self.signal_to_noise_ratio =
            (self.max_parameter_value - self.min_parameter_value) * ratio + self.min_parameter_value;

        // Raster update.
        let width = if ratio <= WIDTH_LOW_RATIO_THRESHOLD {
            WIDTH_BASELINE - WIDTH_SLOPE * ratio
        } else {
            0.0
        };

        let ratio_scaled_to_raster_length = (ratio * self.raster_length as f64).round() as i32;
        let ratio_scaled_to_max_raster_value = ratio * self.max_raster_value;

        for i in 0..self.raster_length {
            let mut value = self.raster[i];

            if i as i32 == ratio_scaled_to_raster_length {
                value *= 0.5;
            } else if width == 0.0 {
                // Above the high-ratio threshold: handled by the ratio>0.5 override below.
            } else {
                let argument = (ratio_scaled_to_max_raster_value - self.raster_step * i as f64) / width;
                if (i as i32) < ratio_scaled_to_raster_length {
                    value = RASTER_BLEND_HALF * (argument.powf(RASTER_BLEND_EXPONENT) + value);
                } else {
                    let argument = -argument;
                    if argument > RASTER_NEGATIVE_ARG_CUTOFF {
                        value = RASTER_BLEND_HALF * (argument.powf(RASTER_BLEND_EXPONENT) + value);
                    } else {
                        value = self.max_raster_value;
                    }
                }
            }

            if value < 0.0 {
                value = 0.0;
            } else if value > self.max_raster_value {
                value = self.max_raster_value;
            }

            if ratio > WIDTH_LOW_RATIO_THRESHOLD {
                value = self.max_raster_value;
            }

            self.raster[i] = value;
        }

        if !primed {
            return (
                Heatmap::empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                f64::NAN,
            );
        }

        let mut value_min = f64::INFINITY;
        let mut value_max = f64::NEG_INFINITY;
        let values: Vec<f64> = self.raster.clone();

        for &v in &values {
            if v < value_min { value_min = v; }
            if v > value_max { value_max = v; }
        }

        let heatmap = Heatmap::new(
            time,
            self.min_parameter_value,
            self.max_parameter_value,
            self.parameter_resolution,
            value_min,
            value_max,
            values,
        );

        (heatmap, self.signal_to_noise_ratio)
    }
}

impl Indicator for CoronaSignalToNoiseRatio {
    fn is_primed(&self) -> bool {
        self.corona.is_primed()
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CoronaSignalToNoiseRatio,
            &self.mnemonic,
            &self.description,
            &[
                OutputText { mnemonic: self.mnemonic.clone(), description: self.description.clone() },
                OutputText { mnemonic: self.mnemonic_snr.clone(), description: self.description_snr.clone() },
            ],
        )
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        let (h, snr) = self.update(sample, bar.low, bar.high, bar.time);
        vec![Box::new(h) as Box<dyn Any>, Box::new(Scalar::new(bar.time, snr))]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let v = (self.quote_func)(quote);
        let (h, snr) = self.update(v, v, v, quote.time);
        vec![Box::new(h) as Box<dyn Any>, Box::new(Scalar::new(quote.time, snr))]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let v = (self.trade_func)(trade);
        let (h, snr) = self.update(v, v, v, trade.time);
        vec![Box::new(h) as Box<dyn Any>, Box::new(Scalar::new(trade.time, snr))]
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let (h, snr) = self.update(scalar.value, scalar.value, scalar.value, scalar.time);
        vec![Box::new(h) as Box<dyn Any>, Box::new(Scalar::new(scalar.time, snr))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    const TOLERANCE: f64 = 1e-4;

    fn make_hl(i: usize, sample: f64) -> (f64, f64) {
        let frac = 0.005 + 0.03 * (1.0 + (i as f64 * 0.37).sin());
        let half = sample * frac;
        (sample - half, sample + half)
    }

    #[test]
    fn test_csnr_update() {
        let input = testdata::talib_input();

        struct Snap { i: usize, snr: f64, vmn: f64, vmx: f64 }
        let snapshots = [
            Snap { i: 11,  snr: 1.0000000000, vmn: 0.0000000000, vmx: 20.0000000000 },
            Snap { i: 12,  snr: 1.0000000000, vmn: 0.0000000000, vmx: 20.0000000000 },
            Snap { i: 50,  snr: 1.0000000000, vmn: 0.0000000000, vmx: 20.0000000000 },
            Snap { i: 100, snr: 2.9986583538, vmn: 4.2011609652, vmx: 20.0000000000 },
            Snap { i: 150, snr: 1.0000000000, vmn: 0.0000000035, vmx: 20.0000000000 },
            Snap { i: 200, snr: 1.0000000000, vmn: 0.0000000000, vmx: 20.0000000000 },
            Snap { i: 251, snr: 1.0000000000, vmn: 0.0000000026, vmx: 20.0000000000 },
        ];

        let mut x = CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams::default()).unwrap();

        let mut si = 0;
        for (i, &v) in input.iter().enumerate() {
            let (low, high) = make_hl(i, v);
            let (h, snr) = x.update(v, low, high, i as i64);

            assert_eq!(h.parameter_first, 1.0, "[{}] parameter_first", i);
            assert_eq!(h.parameter_last, 11.0, "[{}] parameter_last", i);
            assert!((h.parameter_resolution - 4.9).abs() < 1e-9, "[{}] parameter_resolution", i);

            if !x.is_primed() {
                assert!(h.is_empty(), "[{}] expected empty heatmap before priming", i);
                assert!(snr.is_nan(), "[{}] expected NaN snr before priming", i);
                continue;
            }

            assert_eq!(h.values.len(), 50, "[{}] heatmap values length", i);

            if si < snapshots.len() && snapshots[si].i == i {
                assert!(
                    (snapshots[si].snr - snr).abs() < TOLERANCE,
                    "[{}] snr: expected {}, got {}", i, snapshots[si].snr, snr
                );
                assert!(
                    (snapshots[si].vmn - h.value_min).abs() < TOLERANCE,
                    "[{}] vmin: expected {}, got {}", i, snapshots[si].vmn, h.value_min
                );
                assert!(
                    (snapshots[si].vmx - h.value_max).abs() < TOLERANCE,
                    "[{}] vmax: expected {}, got {}", i, snapshots[si].vmx, h.value_max
                );
                si += 1;
            }
        }

        assert_eq!(si, snapshots.len(), "did not hit all snapshots");
    }

    #[test]
    fn test_csnr_primes_at_bar_11() {
        let mut x = CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams::default()).unwrap();
        assert!(!x.is_primed());

        let input = testdata::talib_input();
        let mut primed_at: Option<usize> = None;

        for (i, &v) in input.iter().enumerate() {
            let (low, high) = make_hl(i, v);
            x.update(v, low, high, i as i64);
            if x.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert_eq!(primed_at, Some(11), "expected priming at index 11");
    }

    #[test]
    fn test_csnr_nan_input() {
        let mut x = CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams::default()).unwrap();
        let (h, snr) = x.update(f64::NAN, f64::NAN, f64::NAN, 0);
        assert!(h.is_empty());
        assert!(snr.is_nan());
        assert!(!x.is_primed());
    }

    #[test]
    fn test_csnr_metadata() {
        let x = CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams::default()).unwrap();
        let md = x.metadata();

        assert_eq!(md.identifier, Identifier::CoronaSignalToNoiseRatio);
        assert_eq!(md.mnemonic, "csnr(50, 20, 1, 11, 30, hl/2)");
        assert_eq!(md.description, "Corona signal to noise ratio csnr(50, 20, 1, 11, 30, hl/2)");
        assert_eq!(md.outputs.len(), 2);

        assert_eq!(md.outputs[0].kind, 1);
        assert_eq!(md.outputs[0].mnemonic, "csnr(50, 20, 1, 11, 30, hl/2)");
        assert_eq!(md.outputs[1].kind, 2);
        assert_eq!(md.outputs[1].mnemonic, "csnr-snr(30, hl/2)");
    }

    #[test]
    fn test_csnr_update_bar() {
        let input = testdata::talib_input();
        let mut x = CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams::default()).unwrap();

        for (i, &v) in input.iter().take(50).enumerate() {
            let (low, high) = make_hl(i, v);
            x.update(v, low, high, i as i64);
        }

        let bar = Bar::new(100, 99.5, 100.5, 99.5, 100.0, 0.0);
        let out = x.update_bar(&bar);
        assert_eq!(out.len(), 2);
    }

    #[test]
    fn test_csnr_invalid_params() {
        // RasterLength < 2
        assert!(CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams {
            raster_length: 1,
            ..CoronaSignalToNoiseRatioParams::default()
        }).is_err());

        // MaxParameterValue <= MinParameterValue
        assert!(CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams {
            min_parameter_value: 5.0,
            max_parameter_value: 5.0,
            ..CoronaSignalToNoiseRatioParams::default()
        }).is_err());

        // HighPassFilterCutoff < 2
        assert!(CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams {
            high_pass_filter_cutoff: 1,
            ..CoronaSignalToNoiseRatioParams::default()
        }).is_err());

        // MinimalPeriod < 2
        assert!(CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams {
            minimal_period: 1,
            ..CoronaSignalToNoiseRatioParams::default()
        }).is_err());

        // MaximalPeriod <= MinimalPeriod
        assert!(CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams {
            minimal_period: 10,
            maximal_period: 10,
            ..CoronaSignalToNoiseRatioParams::default()
        }).is_err());
    }
}
