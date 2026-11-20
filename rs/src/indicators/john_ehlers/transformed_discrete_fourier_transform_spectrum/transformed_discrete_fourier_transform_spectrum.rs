use std::any::Any;
use std::f64::consts::{LN_10, PI};

use crate::entities::bar::Bar;
use crate::entities::bar_component::{
    component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT,
};
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

const TWO_PI: f64 = 2.0 * PI;
const LOG_TEN: f64 = LN_10;
const DB_CAP: f64 = 20.0;
const CENTROID_DB_LIMIT: f64 = 3.0;

// ---------------------------------------------------------------------------
// Output enum
// ---------------------------------------------------------------------------

/// Output indices for Ehlers' Transformed Discrete Fourier Transform Spectrum indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum TransformedDiscreteFourierTransformSpectrumOutput {
    /// Spectrum decibel heatmap column across `[min_period, max_period]`.
    Value = 1,
    /// 3 dB center-of-gravity dominant cycle period estimate.
    DominantCycle = 2,
}

impl TransformedDiscreteFourierTransformSpectrumOutput {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Value => "value",
            Self::DominantCycle => "dominantCycle",
        }
    }

    pub fn is_known(&self) -> bool {
        matches!(self, Self::Value | Self::DominantCycle)
    }
}

impl std::fmt::Display for TransformedDiscreteFourierTransformSpectrumOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

// ---------------------------------------------------------------------------
// Params & mnemonics
// ---------------------------------------------------------------------------

/// Parameters to create a `TransformedDiscreteFourierTransformSpectrum` instance.
///
/// Defaults correspond to Ehlers' January 2007 article settings.
/// Explicit values (including `0`) are validated directly without silent substitution.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct TransformedDiscreteFourierTransformSpectrumParams {
    /// Shortest evaluated integer cycle period in bars (>= 8). Default is 8.
    pub min_period: i32,
    /// Longest evaluated integer cycle period in bars (>= min_period). Default is 50.
    pub max_period: i32,
    /// Sliding observation window length in cleaned samples (>= 1). Default is 50.
    pub window_length: i32,
    /// Single-pole high-pass cutoff period in bars (> 4.0). Default is 40.0.
    pub highpass_period: f64,
    /// Bar component to extract (`None` defaults to `Close` and is omitted from the mnemonic).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract (`None` defaults to `Mid` and is omitted from the mnemonic).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract (`None` defaults to `Price` and is omitted from the mnemonic).
    pub trade_component: Option<TradeComponent>,
}

pub type Params = TransformedDiscreteFourierTransformSpectrumParams;

impl Default for TransformedDiscreteFourierTransformSpectrumParams {
    fn default() -> Self {
        Self {
            min_period: 8,
            max_period: 50,
            window_length: 50,
            highpass_period: 40.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Returns default Ehlers January 2007 article parameters.
pub fn default_params() -> TransformedDiscreteFourierTransformSpectrumParams {
    TransformedDiscreteFourierTransformSpectrumParams::default()
}

/// Formats a float parameter cleanly for mnemonics without `-0`.
pub fn format_number(value: f64) -> String {
    let normalized = if value == 0.0 { 0.0 } else { value };
    format!("{normalized}")
}

fn format_mnemonic(
    prefix: &str,
    params: &TransformedDiscreteFourierTransformSpectrumParams,
) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let comp_suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    let hp_text = format_number(params.highpass_period);
    format!(
        "{prefix}({}, {}, {}, {hp_text}{comp_suffix})",
        params.min_period, params.max_period, params.window_length
    )
}

/// Builds the primary heatmap mnemonic for `TransformedDiscreteFourierTransformSpectrum`.
pub fn mnemonic(params: &TransformedDiscreteFourierTransformSpectrumParams) -> String {
    format_mnemonic("tdfts", params)
}

/// Builds the dominant-cycle mnemonic for `TransformedDiscreteFourierTransformSpectrum`.
pub fn dominant_cycle_mnemonic(
    params: &TransformedDiscreteFourierTransformSpectrumParams,
) -> String {
    format_mnemonic("tdfts-dc", params)
}

// ---------------------------------------------------------------------------
// Output & state snapshots
// ---------------------------------------------------------------------------

/// Spectrum heatmap column and dominant cycle period returned per observation.
#[derive(Debug, Clone)]
pub struct SpectrumOutput {
    pub heatmap: Heatmap,
    pub dominant_cycle: f64,
}

/// Snapshot of committed internal detrending, spectral, and centroid state.
#[derive(Debug, Clone, PartialEq)]
pub struct SpectrumState {
    pub high_pass: f64,
    pub cleaned_data: f64,
    pub maximum_power: f64,
    pub dominant_cycle: f64,
    pub cosine_parts: Vec<f64>,
    pub sine_parts: Vec<f64>,
    pub powers: Vec<f64>,
    pub decibels: Vec<f64>,
}

/// Validated numeric parameters resolved at construction time.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct ResolvedParams {
    pub min_period: usize,
    pub max_period: usize,
    pub window_length: usize,
    pub highpass_period: f64,
}

// ---------------------------------------------------------------------------
// Indicator implementation
// ---------------------------------------------------------------------------

/// Ehlers' January 2007 Transformed Discrete Fourier Transform Spectrum (`TDFTS`).
pub struct TransformedDiscreteFourierTransformSpectrum {
    resolved: ResolvedParams,
    highpass_alpha: f64,
    warmup_bars: usize,
    metadata: Metadata,
    periods: Vec<usize>,
    cosines: Vec<Vec<f64>>,
    sines: Vec<Vec<f64>>,
    sample_count: usize,
    previous_price: f64,
    hp_history: [f64; 6],
    cleaned: f64,
    window: Vec<f64>,
    cosine_parts: Vec<f64>,
    sine_parts: Vec<f64>,
    powers: Vec<f64>,
    decibels: Vec<f64>,
    maximum: f64,
    dominant: f64,
    last_output: SpectrumOutput,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl TransformedDiscreteFourierTransformSpectrum {
    /// Creates a new `TransformedDiscreteFourierTransformSpectrum` with the supplied parameters.
    pub fn new(params: &TransformedDiscreteFourierTransformSpectrumParams) -> Result<Self, String> {
        const INVALID: &str = "invalid transformed discrete Fourier transform spectrum parameters";

        if params.min_period < 8 {
            return Err(format!("{INVALID}: min_period must be an integer >= 8"));
        }
        if params.max_period < params.min_period {
            return Err(format!(
                "{INVALID}: max_period must be an integer >= min_period"
            ));
        }
        if params.window_length < 1 {
            return Err(format!("{INVALID}: window_length must be an integer >= 1"));
        }
        if !params.highpass_period.is_finite() || params.highpass_period <= 4.0 {
            return Err(format!(
                "{INVALID}: highpass_period must be a finite real number > 4.0"
            ));
        }

        let omega = TWO_PI / params.highpass_period;
        let cosine_omega = omega.cos();
        if cosine_omega == 0.0 {
            return Err(format!(
                "{INVALID}: highpass_period produces a singular high-pass pole"
            ));
        }
        let highpass_alpha = (1.0 - omega.sin()) / cosine_omega;
        if !highpass_alpha.is_finite() {
            return Err(format!(
                "{INVALID}: highpass_period must produce a finite high-pass pole"
            ));
        }

        let min_period = params.min_period as usize;
        let max_period = params.max_period as usize;
        let window_length = params.window_length as usize;
        let warmup_bars = 5_usize.max(window_length - 1);

        let resolved = ResolvedParams {
            min_period,
            max_period,
            window_length,
            highpass_period: params.highpass_period,
        };

        let periods: Vec<usize> = (min_period..=max_period).collect();
        let channel_count = periods.len();
        let mut cosines = Vec::with_capacity(channel_count);
        let mut sines = Vec::with_capacity(channel_count);
        for &period in &periods {
            let period_float = period as f64;
            let mut cosine_row = Vec::with_capacity(window_length);
            let mut sine_row = Vec::with_capacity(window_length);
            for lag in 0..window_length {
                let angle = TWO_PI * (lag as f64) / period_float;
                cosine_row.push(angle.cos());
                sine_row.push(angle.sin());
            }
            cosines.push(cosine_row);
            sines.push(sine_row);
        }

        let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let mnemonic_text = mnemonic(params);
        let description_text =
            format!("Transformed discrete Fourier transform spectrum {mnemonic_text}");
        let mnemonic_dc = dominant_cycle_mnemonic(params);
        let description_dc = format!(
            "Transformed discrete Fourier transform spectrum dominant cycle {mnemonic_dc}"
        );

        let metadata = build_metadata(
            Identifier::TransformedDiscreteFourierTransformSpectrum,
            &mnemonic_text,
            &description_text,
            &[
                OutputText {
                    mnemonic: mnemonic_text.clone(),
                    description: description_text.clone(),
                },
                OutputText {
                    mnemonic: mnemonic_dc,
                    description: description_dc,
                },
            ],
        );

        let initial_output = SpectrumOutput {
            heatmap: Heatmap::empty(0, min_period as f64, max_period as f64, 1.0),
            dominant_cycle: f64::NAN,
        };

        Ok(Self {
            resolved,
            highpass_alpha,
            warmup_bars,
            metadata,
            periods,
            cosines,
            sines,
            sample_count: 0,
            previous_price: 0.0,
            hp_history: [0.0; 6],
            cleaned: 0.0,
            window: vec![0.0; window_length],
            cosine_parts: vec![0.0; channel_count],
            sine_parts: vec![0.0; channel_count],
            powers: vec![0.0; channel_count],
            decibels: vec![0.0; channel_count],
            maximum: 0.0,
            dominant: 0.0,
            last_output: initial_output,
            bar_func: bar_component_value(bar_comp),
            quote_func: quote_component_value(quote_comp),
            trade_func: trade_component_value(trade_comp),
        })
    }

    /// Creates a new `TransformedDiscreteFourierTransformSpectrum` with default article parameters.
    pub fn new_default() -> Result<Self, String> {
        Self::new(&TransformedDiscreteFourierTransformSpectrumParams::default())
    }

    /// Returns the validated resolved parameters.
    pub fn parameters(&self) -> ResolvedParams {
        self.resolved
    }

    /// Returns the shortest evaluated cycle period.
    pub fn min_period(&self) -> usize {
        self.resolved.min_period
    }

    /// Returns the longest evaluated cycle period.
    pub fn max_period(&self) -> usize {
        self.resolved.max_period
    }

    /// Returns the precomputed single-pole high-pass feedback coefficient.
    pub fn highpass_alpha(&self) -> f64 {
        self.highpass_alpha
    }

    /// Returns the number of warmup observations (`max(5, window_length - 1)`).
    pub fn warmup_bars(&self) -> usize {
        self.warmup_bars
    }

    /// Returns the number of accepted finite observations seen since construction or reset.
    pub fn samples_seen(&self) -> usize {
        self.sample_count
    }

    /// Returns an owned snapshot of committed internal state.
    pub fn state(&self) -> SpectrumState {
        SpectrumState {
            high_pass: self.hp_history[0],
            cleaned_data: self.cleaned,
            maximum_power: self.maximum,
            dominant_cycle: self.dominant,
            cosine_parts: self.cosine_parts.clone(),
            sine_parts: self.sine_parts.clone(),
            powers: self.powers.clone(),
            decibels: self.decibels.clone(),
        }
    }

    /// Returns an owned copy of the most recent accepted output.
    pub fn output(&self) -> SpectrumOutput {
        self.last_output.clone()
    }

    /// Resets all filter, window, spectral, nonlinear transform, and centroid state.
    pub fn reset(&mut self) {
        self.sample_count = 0;
        self.previous_price = 0.0;
        self.hp_history = [0.0; 6];
        self.cleaned = 0.0;
        self.window.fill(0.0);
        self.cosine_parts.fill(0.0);
        self.sine_parts.fill(0.0);
        self.powers.fill(0.0);
        self.decibels.fill(0.0);
        self.maximum = 0.0;
        self.dominant = 0.0;
        self.last_output = SpectrumOutput {
            heatmap: Heatmap::empty(
                0,
                self.resolved.min_period as f64,
                self.resolved.max_period as f64,
                1.0,
            ),
            dominant_cycle: f64::NAN,
        };
    }

    /// Fallibly updates the indicator with a scalar sample and timestamp.
    ///
    /// - `NaN` samples are skipped without mutating state or aging `samples_seen`.
    /// - `±Inf` samples or candidate arithmetic overflows return `Err` without mutating state.
    pub fn try_update(&mut self, sample: f64, time: i64) -> Result<SpectrumOutput, String> {
        if sample.is_infinite() {
            return Err("input must be finite or NaN (missing)".to_string());
        }
        let min_float = self.resolved.min_period as f64;
        let max_float = self.resolved.max_period as f64;
        if sample.is_nan() {
            return Ok(SpectrumOutput {
                heatmap: Heatmap::empty(time, min_float, max_float, 1.0),
                dominant_cycle: f64::NAN,
            });
        }

        let bar_number = self.sample_count + 1;
        let (high_pass, cleaned) = if bar_number <= 5 {
            (sample, sample)
        } else {
            let next_hp = 0.5 * (1.0 + self.highpass_alpha) * (sample - self.previous_price)
                + self.highpass_alpha * self.hp_history[0];
            let next_cleaned = (next_hp
                + 2.0 * self.hp_history[0]
                + 3.0 * self.hp_history[1]
                + 3.0 * self.hp_history[2]
                + 2.0 * self.hp_history[3]
                + self.hp_history[4])
                / 12.0;
            (next_hp, next_cleaned)
        };

        if !high_pass.is_finite() || !cleaned.is_finite() {
            return Err("detrending arithmetic overflowed".to_string());
        }

        let window_length = self.resolved.window_length;
        let mut next_window = Vec::with_capacity(window_length);
        next_window.push(cleaned);
        next_window.extend_from_slice(&self.window[..window_length - 1]);

        let channel_count = self.periods.len();
        let mut next_cosines = Vec::with_capacity(channel_count);
        let mut next_sines = Vec::with_capacity(channel_count);
        let mut next_powers = Vec::with_capacity(channel_count);
        let mut maximum = 0.0_f64;

        for channel_index in 0..channel_count {
            let mut cosine_sum = 0.0;
            let mut sine_sum = 0.0;
            let cosine_row = &self.cosines[channel_index];
            let sine_row = &self.sines[channel_index];
            for lag in 0..window_length {
                cosine_sum += next_window[lag] * cosine_row[lag];
                sine_sum += next_window[lag] * sine_row[lag];
            }
            let power = cosine_sum * cosine_sum + sine_sum * sine_sum;
            if !cosine_sum.is_finite() || !sine_sum.is_finite() || !power.is_finite() || power < 0.0
            {
                return Err("spectral power arithmetic overflowed".to_string());
            }
            if power > maximum {
                maximum = power;
            }
            next_cosines.push(cosine_sum);
            next_sines.push(sine_sum);
            next_powers.push(power);
        }

        let mut next_decibels = self.decibels.clone();
        for (channel_index, &power) in next_powers.iter().enumerate() {
            if maximum > 0.0 && power > 0.0 {
                next_decibels[channel_index] =
                    -10.0 * (0.01 / (1.0 - 0.99 * power / maximum)).ln() / LOG_TEN;
            }
            if next_decibels[channel_index] > DB_CAP {
                next_decibels[channel_index] = DB_CAP;
            }
            if !next_decibels[channel_index].is_finite() {
                return Err("decibel transform arithmetic overflowed".to_string());
            }
        }

        let mut numerator = 0.0;
        let mut denominator = 0.0;
        for (&period, &channel_db) in self.periods.iter().zip(next_decibels.iter()) {
            if channel_db < CENTROID_DB_LIMIT {
                let weight = CENTROID_DB_LIMIT - channel_db;
                numerator += (period as f64) * weight;
                denominator += weight;
            }
        }
        let dominant = if denominator != 0.0 {
            numerator / denominator
        } else {
            self.dominant
        };
        if !dominant.is_finite() {
            return Err("centroid arithmetic overflowed".to_string());
        }

        self.previous_price = sample;
        for slot in (1..6).rev() {
            self.hp_history[slot] = self.hp_history[slot - 1];
        }
        self.hp_history[0] = high_pass;
        self.cleaned = cleaned;
        self.window = next_window;
        self.cosine_parts = next_cosines;
        self.sine_parts = next_sines;
        self.powers = next_powers;
        self.decibels = next_decibels.clone();
        self.maximum = maximum;
        self.dominant = dominant;
        self.sample_count = bar_number;

        let current_output = if self.is_primed() {
            let mut value_min = next_decibels[0];
            let mut value_max = next_decibels[0];
            for &item in &next_decibels[1..] {
                if item < value_min {
                    value_min = item;
                }
                if item > value_max {
                    value_max = item;
                }
            }
            SpectrumOutput {
                heatmap: Heatmap::new(
                    time,
                    min_float,
                    max_float,
                    1.0,
                    value_min,
                    value_max,
                    next_decibels,
                ),
                dominant_cycle: dominant,
            }
        } else {
            SpectrumOutput {
                heatmap: Heatmap::empty(time, min_float, max_float, 1.0),
                dominant_cycle: f64::NAN,
            }
        };

        self.last_output = current_output.clone();
        Ok(current_output)
    }

    /// Updates the indicator with a scalar sample and timestamp, panicking on invalid input/overflow.
    pub fn update(&mut self, sample: f64, time: i64) -> SpectrumOutput {
        self.try_update(sample, time).unwrap_or_else(|err| {
            panic!("TransformedDiscreteFourierTransformSpectrum::update failed: {err}")
        })
    }

    fn wrap_output(time: i64, output: SpectrumOutput) -> Output {
        vec![
            Box::new(output.heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, output.dominant_cycle)),
        ]
    }
}

impl Indicator for TransformedDiscreteFourierTransformSpectrum {
    fn is_primed(&self) -> bool {
        self.sample_count > self.warmup_bars
    }

    fn metadata(&self) -> Metadata {
        self.metadata.clone()
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        Self::wrap_output(sample.time, self.update(sample.value, sample.time))
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        Self::wrap_output(bar.time, self.update(sample, bar.time))
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.quote_func)(quote);
        Self::wrap_output(quote.time, self.update(sample, quote.time))
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.trade_func)(trade);
        Self::wrap_output(trade.time, self.update(sample, trade.time))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::core::descriptor::descriptor_of;
    use crate::indicators::core::outputs::shape::Shape;
    use crate::indicators::john_ehlers::transformed_discrete_fourier_transform_spectrum::testdata::testdata as data;

    const TOLERANCE: f64 = 1e-12;

    fn assert_close(actual: f64, expected: f64, label: &str) {
        if expected.is_nan() {
            assert!(actual.is_nan(), "{label}: expected NaN, got {actual}");
        } else {
            assert!(
                !actual.is_nan(),
                "{label}: expected {expected}, got NaN"
            );
            let diff = (actual - expected).abs();
            assert!(
                diff <= TOLERANCE,
                "{label}: actual={actual} expected={expected} diff={diff}"
            );
        }
    }

    fn verify_fixture_case(
        case_name: &str,
        params: TransformedDiscreteFourierTransformSpectrumParams,
        expected_heatmap: &[Vec<f64>],
        expected_min: &[f64],
        expected_max: &[f64],
        expected_dc: &[f64],
    ) {
        let prices = data::test_input();
        let mut indicator =
            TransformedDiscreteFourierTransformSpectrum::new(&params).expect("valid params");
        let expected_bins = (params.max_period - params.min_period + 1) as usize;
        let expected_warmup = 5_usize.max((params.window_length - 1) as usize);
        assert_eq!(indicator.warmup_bars(), expected_warmup, "{case_name} warmup");

        for (bar_index, &price) in prices.iter().enumerate() {
            let timestamp = bar_index as i64;
            let output = indicator.update(price, timestamp);
            let primed_expected = bar_index >= expected_warmup;
            assert_eq!(
                indicator.is_primed(),
                primed_expected,
                "{case_name} bar {bar_index} primed"
            );

            assert_eq!(output.heatmap.time, timestamp);
            assert_eq!(output.heatmap.parameter_first, params.min_period as f64);
            assert_eq!(output.heatmap.parameter_last, params.max_period as f64);
            assert_eq!(output.heatmap.parameter_resolution, 1.0);

            let row = &expected_heatmap[bar_index];
            if row.is_empty() || row.iter().all(|item| item.is_nan()) {
                assert!(
                    output.heatmap.is_empty(),
                    "{case_name} bar {bar_index} expected empty heatmap"
                );
                assert!(expected_min[bar_index].is_nan());
                assert!(expected_max[bar_index].is_nan());
                assert!(output.heatmap.value_min.is_nan());
                assert!(output.heatmap.value_max.is_nan());
            } else {
                assert_eq!(output.heatmap.values.len(), expected_bins);
                assert_eq!(output.heatmap.values.len(), row.len());
                assert_close(
                    output.heatmap.value_min,
                    expected_min[bar_index],
                    &format!("{case_name} bar {bar_index} min"),
                );
                assert_close(
                    output.heatmap.value_max,
                    expected_max[bar_index],
                    &format!("{case_name} bar {bar_index} max"),
                );
                for (bin_index, &expected_db) in row.iter().enumerate() {
                    assert_close(
                        output.heatmap.values[bin_index],
                        expected_db,
                        &format!("{case_name} bar {bar_index} bin {bin_index}"),
                    );
                }
            }

            assert_close(
                output.dominant_cycle,
                expected_dc[bar_index],
                &format!("{case_name} bar {bar_index} dc"),
            );
        }
    }

    #[test]
    fn test_reference_fixture_parity_all_12_cases() {
        verify_fixture_case(
            "p8_q50_w50_hp40",
            default_params(),
            &data::expected_p8_q50_w50_hp40_heatmap(),
            &data::expected_p8_q50_w50_hp40_min(),
            &data::expected_p8_q50_w50_hp40_max(),
            &data::expected_p8_q50_w50_hp40_dominant_cycle(),
        );
        verify_fixture_case(
            "p8_q50_w1_hp40",
            Params {
                window_length: 1,
                ..default_params()
            },
            &data::expected_p8_q50_w1_hp40_heatmap(),
            &data::expected_p8_q50_w1_hp40_min(),
            &data::expected_p8_q50_w1_hp40_max(),
            &data::expected_p8_q50_w1_hp40_dominant_cycle(),
        );
        verify_fixture_case(
            "p8_q50_w20_hp40",
            Params {
                window_length: 20,
                ..default_params()
            },
            &data::expected_p8_q50_w20_hp40_heatmap(),
            &data::expected_p8_q50_w20_hp40_min(),
            &data::expected_p8_q50_w20_hp40_max(),
            &data::expected_p8_q50_w20_hp40_dominant_cycle(),
        );
        verify_fixture_case(
            "p8_q50_w48_hp40",
            Params {
                window_length: 48,
                ..default_params()
            },
            &data::expected_p8_q50_w48_hp40_heatmap(),
            &data::expected_p8_q50_w48_hp40_min(),
            &data::expected_p8_q50_w48_hp40_max(),
            &data::expected_p8_q50_w48_hp40_dominant_cycle(),
        );
        verify_fixture_case(
            "p8_q50_w64_hp40",
            Params {
                window_length: 64,
                ..default_params()
            },
            &data::expected_p8_q50_w64_hp40_heatmap(),
            &data::expected_p8_q50_w64_hp40_min(),
            &data::expected_p8_q50_w64_hp40_max(),
            &data::expected_p8_q50_w64_hp40_dominant_cycle(),
        );
        verify_fixture_case(
            "p8_q50_w252_hp40",
            Params {
                window_length: 252,
                ..default_params()
            },
            &data::expected_p8_q50_w252_hp40_heatmap(),
            &data::expected_p8_q50_w252_hp40_min(),
            &data::expected_p8_q50_w252_hp40_max(),
            &data::expected_p8_q50_w252_hp40_dominant_cycle(),
        );
        verify_fixture_case(
            "p8_q50_w50_hp20",
            Params {
                highpass_period: 20.0,
                ..default_params()
            },
            &data::expected_p8_q50_w50_hp20_heatmap(),
            &data::expected_p8_q50_w50_hp20_min(),
            &data::expected_p8_q50_w50_hp20_max(),
            &data::expected_p8_q50_w50_hp20_dominant_cycle(),
        );
        verify_fixture_case(
            "p8_q50_w50_hp80",
            Params {
                highpass_period: 80.0,
                ..default_params()
            },
            &data::expected_p8_q50_w50_hp80_heatmap(),
            &data::expected_p8_q50_w50_hp80_min(),
            &data::expected_p8_q50_w50_hp80_max(),
            &data::expected_p8_q50_w50_hp80_dominant_cycle(),
        );
        verify_fixture_case(
            "p8_q50_w50_hp40_5",
            Params {
                highpass_period: 40.5,
                ..default_params()
            },
            &data::expected_p8_q50_w50_hp40_5_heatmap(),
            &data::expected_p8_q50_w50_hp40_5_min(),
            &data::expected_p8_q50_w50_hp40_5_max(),
            &data::expected_p8_q50_w50_hp40_5_dominant_cycle(),
        );
        verify_fixture_case(
            "p10_q30_w50_hp40",
            Params {
                min_period: 10,
                max_period: 30,
                ..default_params()
            },
            &data::expected_p10_q30_w50_hp40_heatmap(),
            &data::expected_p10_q30_w50_hp40_min(),
            &data::expected_p10_q30_w50_hp40_max(),
            &data::expected_p10_q30_w50_hp40_dominant_cycle(),
        );
        verify_fixture_case(
            "p8_q80_w50_hp40",
            Params {
                min_period: 8,
                max_period: 80,
                ..default_params()
            },
            &data::expected_p8_q80_w50_hp40_heatmap(),
            &data::expected_p8_q80_w50_hp40_min(),
            &data::expected_p8_q80_w50_hp40_max(),
            &data::expected_p8_q80_w50_hp40_dominant_cycle(),
        );
        verify_fixture_case(
            "p20_q20_w50_hp40",
            Params {
                min_period: 20,
                max_period: 20,
                ..default_params()
            },
            &data::expected_p20_q20_w50_hp40_heatmap(),
            &data::expected_p20_q20_w50_hp40_min(),
            &data::expected_p20_q20_w50_hp40_max(),
            &data::expected_p20_q20_w50_hp40_dominant_cycle(),
        );
    }

    #[test]
    fn test_nan_skip_overflow_atomicity_and_reset() {
        let prices = data::test_input();
        let mut clean = TransformedDiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut noisy = TransformedDiscreteFourierTransformSpectrum::new_default().unwrap();

        for (bar_index, &price) in prices.iter().take(80).enumerate() {
            if bar_index == 0 || bar_index == 4 || bar_index == 49 || bar_index == 65 {
                let before_state = noisy.state();
                let before_seen = noisy.samples_seen();
                let skipped = noisy.update(f64::NAN, 9999);
                assert!(skipped.heatmap.is_empty());
                assert!(skipped.dominant_cycle.is_nan());
                assert_eq!(noisy.samples_seen(), before_seen);
                assert_eq!(noisy.state(), before_state);
            }

            let out_clean = clean.update(price, bar_index as i64);
            let out_noisy = noisy.update(price, bar_index as i64);
            assert_eq!(clean.is_primed(), noisy.is_primed());
            assert_eq!(out_clean.heatmap.values.len(), out_noisy.heatmap.values.len());
            for bin_index in 0..out_clean.heatmap.values.len() {
                assert_close(
                    out_noisy.heatmap.values[bin_index],
                    out_clean.heatmap.values[bin_index],
                    "nan skip parity",
                );
            }
            assert_close(
                out_noisy.dominant_cycle,
                out_clean.dominant_cycle,
                "nan skip dc parity",
            );
        }

        let snapshot = clean.state();
        let seen_before = clean.samples_seen();
        assert!(clean.try_update(f64::INFINITY, 100).is_err());
        assert!(clean.try_update(f64::NEG_INFINITY, 101).is_err());
        assert!(clean.try_update(1e308, 102).is_err());
        assert_eq!(clean.samples_seen(), seen_before);
        assert_eq!(clean.state(), snapshot);

        clean.reset();
        assert!(!clean.is_primed());
        assert_eq!(clean.samples_seen(), 0);

        let mut fresh = TransformedDiscreteFourierTransformSpectrum::new_default().unwrap();
        for (bar_index, &price) in prices.iter().take(70).enumerate() {
            let out_reset = clean.update(price, bar_index as i64);
            let out_fresh = fresh.update(price, bar_index as i64);
            assert_eq!(out_reset.heatmap.values.len(), out_fresh.heatmap.values.len());
            for bin_index in 0..out_reset.heatmap.values.len() {
                assert_close(
                    out_reset.heatmap.values[bin_index],
                    out_fresh.heatmap.values[bin_index],
                    "reset parity",
                );
            }
            assert_close(out_reset.dominant_cycle, out_fresh.dominant_cycle, "reset dc");
        }
    }

    #[test]
    fn test_validation_mnemonics_metadata_and_entities() {
        let bad_cases = [
            Params {
                min_period: 7,
                ..default_params()
            },
            Params {
                min_period: 25,
                max_period: 24,
                ..default_params()
            },
            Params {
                window_length: 0,
                ..default_params()
            },
            Params {
                highpass_period: 4.0,
                ..default_params()
            },
            Params {
                highpass_period: -10.0,
                ..default_params()
            },
            Params {
                highpass_period: f64::NAN,
                ..default_params()
            },
            Params {
                highpass_period: f64::INFINITY,
                ..default_params()
            },
        ];
        for case_params in &bad_cases {
            assert!(
                TransformedDiscreteFourierTransformSpectrum::new(case_params).is_err(),
                "expected error for {case_params:?}"
            );
        }

        let mnemonic_cases: Vec<(Params, &str, &str)> = vec![
            (
                default_params(),
                "tdfts(8, 50, 50, 40)",
                "tdfts-dc(8, 50, 50, 40)",
            ),
            (
                Params {
                    min_period: 10,
                    max_period: 30,
                    window_length: 48,
                    highpass_period: 40.5,
                    bar_component: Some(BarComponent::Median),
                    ..default_params()
                },
                "tdfts(10, 30, 48, 40.5, hl/2)",
                "tdfts-dc(10, 30, 48, 40.5, hl/2)",
            ),
            (
                Params {
                    quote_component: Some(QuoteComponent::Ask),
                    ..default_params()
                },
                "tdfts(8, 50, 50, 40, a)",
                "tdfts-dc(8, 50, 50, 40, a)",
            ),
            (
                Params {
                    trade_component: Some(TradeComponent::Volume),
                    ..default_params()
                },
                "tdfts(8, 50, 50, 40, v)",
                "tdfts-dc(8, 50, 50, 40, v)",
            ),
            (
                Params {
                    bar_component: Some(BarComponent::High),
                    quote_component: Some(QuoteComponent::Ask),
                    ..default_params()
                },
                "tdfts(8, 50, 50, 40, h, a)",
                "tdfts-dc(8, 50, 50, 40, h, a)",
            ),
            (
                Params {
                    bar_component: Some(BarComponent::Low),
                    trade_component: Some(TradeComponent::Volume),
                    ..default_params()
                },
                "tdfts(8, 50, 50, 40, l, v)",
                "tdfts-dc(8, 50, 50, 40, l, v)",
            ),
            (
                Params {
                    quote_component: Some(QuoteComponent::SpreadBp),
                    trade_component: Some(TradeComponent::Volume),
                    ..default_params()
                },
                "tdfts(8, 50, 50, 40, spread bp, v)",
                "tdfts-dc(8, 50, 50, 40, spread bp, v)",
            ),
        ];

        for (case_params, expected_main, expected_dc) in &mnemonic_cases {
            assert_eq!(&mnemonic(case_params), expected_main);
            assert_eq!(&dominant_cycle_mnemonic(case_params), expected_dc);

            let instance =
                TransformedDiscreteFourierTransformSpectrum::new(case_params).unwrap();
            let meta = instance.metadata();
            assert_eq!(
                meta.identifier,
                Identifier::TransformedDiscreteFourierTransformSpectrum
            );
            assert_eq!(&meta.mnemonic, expected_main);
            assert_eq!(meta.outputs.len(), 2);
            assert_eq!(
                meta.outputs[0].kind,
                TransformedDiscreteFourierTransformSpectrumOutput::Value as i32
            );
            assert_eq!(meta.outputs[0].shape, Shape::Heatmap);
            assert_eq!(&meta.outputs[0].mnemonic, expected_main);
            assert_eq!(
                meta.outputs[1].kind,
                TransformedDiscreteFourierTransformSpectrumOutput::DominantCycle as i32
            );
            assert_eq!(meta.outputs[1].shape, Shape::Scalar);
            assert_eq!(&meta.outputs[1].mnemonic, expected_dc);
        }

        let desc = descriptor_of(Identifier::TransformedDiscreteFourierTransformSpectrum)
            .expect("descriptor registered");
        assert_eq!(desc.family, "John Ehlers");
        assert_eq!(desc.outputs.len(), 2);

        assert_eq!(
            TransformedDiscreteFourierTransformSpectrumOutput::Value.as_str(),
            "value"
        );
        assert_eq!(
            TransformedDiscreteFourierTransformSpectrumOutput::DominantCycle.as_str(),
            "dominantCycle"
        );
        assert!(TransformedDiscreteFourierTransformSpectrumOutput::Value.is_known());
        assert!(TransformedDiscreteFourierTransformSpectrumOutput::DominantCycle.is_known());

        let prices = data::test_input();
        let highs = data::test_high();
        let lows = data::test_low();
        let opens = data::test_open();
        let mut bar_ind = TransformedDiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut scalar_ind = TransformedDiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut quote_ind = TransformedDiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut trade_ind = TransformedDiscreteFourierTransformSpectrum::new_default().unwrap();

        for bar_index in 0..70 {
            let bar = Bar {
                time: bar_index as i64,
                open: opens[bar_index],
                high: highs[bar_index],
                low: lows[bar_index],
                close: prices[bar_index],
                volume: 1000.0,
            };
            let scalar = Scalar::new(bar_index as i64, prices[bar_index]);
            let quote = Quote {
                time: bar_index as i64,
                bid_price: prices[bar_index] - 0.5,
                ask_price: prices[bar_index] + 0.5,
                bid_size: 10.0,
                ask_size: 10.0,
            };
            let trade = Trade {
                time: bar_index as i64,
                price: prices[bar_index],
                volume: 50.0,
            };
            let out_bar = bar_ind.update_bar(&bar);
            let out_scalar = scalar_ind.update_scalar(&scalar);
            let out_quote = quote_ind.update_quote(&quote);
            let out_trade = trade_ind.update_trade(&trade);

            assert_eq!(out_bar.len(), 2);
            let hm_bar = out_bar[0].downcast_ref::<Heatmap>().unwrap();
            let hm_scalar = out_scalar[0].downcast_ref::<Heatmap>().unwrap();
            let hm_quote = out_quote[0].downcast_ref::<Heatmap>().unwrap();
            let hm_trade = out_trade[0].downcast_ref::<Heatmap>().unwrap();
            let dc_bar = out_bar[1].downcast_ref::<Scalar>().unwrap();
            let dc_scalar = out_scalar[1].downcast_ref::<Scalar>().unwrap();
            let dc_quote = out_quote[1].downcast_ref::<Scalar>().unwrap();
            let dc_trade = out_trade[1].downcast_ref::<Scalar>().unwrap();

            assert_eq!(hm_bar.values.len(), hm_scalar.values.len());
            assert_eq!(hm_bar.values.len(), hm_quote.values.len());
            assert_eq!(hm_bar.values.len(), hm_trade.values.len());
            for bin_index in 0..hm_bar.values.len() {
                assert_close(hm_bar.values[bin_index], hm_scalar.values[bin_index], "bar vs scalar");
                assert_close(hm_bar.values[bin_index], hm_quote.values[bin_index], "bar vs quote");
                assert_close(hm_bar.values[bin_index], hm_trade.values[bin_index], "bar vs trade");
            }
            assert_close(dc_bar.value, dc_scalar.value, "dc bar vs scalar");
            assert_close(dc_bar.value, dc_quote.value, "dc bar vs quote");
            assert_close(dc_bar.value, dc_trade.value, "dc bar vs trade");
        }
    }
}
