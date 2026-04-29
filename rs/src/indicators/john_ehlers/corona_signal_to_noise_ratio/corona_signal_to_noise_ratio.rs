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

    fn talib_input() -> Vec<f64> {
        vec![
            92.0000, 93.1725, 95.3125, 94.8450, 94.4075, 94.1100, 93.5000, 91.7350, 90.9550, 91.6875,
            94.5000, 97.9700, 97.5775, 90.7825, 89.0325, 92.0950, 91.1550, 89.7175, 90.6100, 91.0000,
            88.9225, 87.5150, 86.4375, 83.8900, 83.0025, 82.8125, 82.8450, 86.7350, 86.8600, 87.5475,
            85.7800, 86.1725, 86.4375, 87.2500, 88.9375, 88.2050, 85.8125, 84.5950, 83.6575, 84.4550,
            83.5000, 86.7825, 88.1725, 89.2650, 90.8600, 90.7825, 91.8600, 90.3600, 89.8600, 90.9225,
            89.5000, 87.6725, 86.5000, 84.2825, 82.9075, 84.2500, 85.6875, 86.6100, 88.2825, 89.5325,
            89.5000, 88.0950, 90.6250, 92.2350, 91.6725, 92.5925, 93.0150, 91.1725, 90.9850, 90.3775,
            88.2500, 86.9075, 84.0925, 83.1875, 84.2525, 97.8600, 99.8750, 103.2650, 105.9375, 103.5000,
            103.1100, 103.6100, 104.6400, 106.8150, 104.9525, 105.5000, 107.1400, 109.7350, 109.8450, 110.9850,
            120.0000, 119.8750, 117.9075, 119.4075, 117.9525, 117.2200, 115.6425, 113.1100, 111.7500, 114.5175,
            114.7450, 115.4700, 112.5300, 112.0300, 113.4350, 114.2200, 119.5950, 117.9650, 118.7150, 115.0300,
            114.5300, 115.0000, 116.5300, 120.1850, 120.5000, 120.5950, 124.1850, 125.3750, 122.9700, 123.0000,
            124.4350, 123.4400, 124.0300, 128.1850, 129.6550, 130.8750, 132.3450, 132.0650, 133.8150, 135.6600,
            137.0350, 137.4700, 137.3450, 136.3150, 136.4400, 136.2850, 129.0950, 128.3100, 126.0000, 124.0300,
            123.9350, 125.0300, 127.2500, 125.6200, 125.5300, 123.9050, 120.6550, 119.9650, 120.7800, 124.0000,
            122.7800, 120.7200, 121.7800, 122.4050, 123.2500, 126.1850, 127.5600, 126.5650, 123.0600, 122.7150,
            123.5900, 122.3100, 122.4650, 123.9650, 123.9700, 124.1550, 124.4350, 127.0000, 125.5000, 128.8750,
            130.5350, 132.3150, 134.0650, 136.0350, 133.7800, 132.7500, 133.4700, 130.9700, 127.5950, 128.4400,
            127.9400, 125.8100, 124.6250, 122.7200, 124.0900, 123.2200, 121.4050, 120.9350, 118.2800, 118.3750,
            121.1550, 120.9050, 117.1250, 113.0600, 114.9050, 112.4350, 107.9350, 105.9700, 106.3700, 106.8450,
            106.9700, 110.0300, 91.0000, 93.5600, 93.6200, 95.3100, 94.1850, 94.7800, 97.6250, 97.5900,
            95.2500, 94.7200, 92.2200, 91.5650, 92.2200, 93.8100, 95.5900, 96.1850, 94.6250, 95.1200,
            94.0000, 93.7450, 95.9050, 101.7450, 106.4400, 107.9350, 103.4050, 105.0600, 104.1550, 103.3100,
            103.3450, 104.8400, 110.4050, 114.5000, 117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300,
            106.2200, 107.7200, 109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
            109.5300, 108.0600,
        ]
    }

    const TOLERANCE: f64 = 1e-4;

    fn make_hl(i: usize, sample: f64) -> (f64, f64) {
        let frac = 0.005 + 0.03 * (1.0 + (i as f64 * 0.37).sin());
        let half = sample * frac;
        (sample - half, sample + half)
    }

    #[test]
    fn test_csnr_update() {
        let input = talib_input();

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

        let input = talib_input();
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
        let input = talib_input();
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
