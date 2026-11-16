use std::any::Any;
use std::f64::consts::{PI, SQRT_2};

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
const SQRT_TWO: f64 = SQRT_2;
const AGC_DECAY: f64 = 0.995;
const CENTROID_THRESHOLD: f64 = 0.5;

// ---------------------------------------------------------------------------
// Output enum
// ---------------------------------------------------------------------------

/// Output indices for Ehlers' Discrete Fourier Transform Spectrum indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DiscreteFourierTransformSpectrumOutput {
    /// Spectrum heatmap column across `[min_period, max_period]`.
    Value = 1,
    /// Center-of-gravity dominant cycle period estimate.
    DominantCycle = 2,
}

impl DiscreteFourierTransformSpectrumOutput {
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

impl std::fmt::Display for DiscreteFourierTransformSpectrumOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

// ---------------------------------------------------------------------------
// Params & mnemonics
// ---------------------------------------------------------------------------

/// Parameters to create a `DiscreteFourierTransformSpectrum` instance.
///
/// Defaults correspond to Ehlers' 2013 Book Listing 9-1 settings.
/// Explicit values (including `0` or `false`) are validated directly without silent substitution.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct DiscreteFourierTransformSpectrumParams {
    /// Shortest evaluated integer cycle period in bars (>= 3). Default is 10.
    pub min_period: i32,
    /// Longest evaluated integer cycle period in bars (>= min_period). Default is 48.
    pub max_period: i32,
    /// Sliding roofed DFT window length in bars (>= 2, independent of max_period). Default is 48.
    pub window_length: i32,
    /// Two-pole roofing high-pass cutoff period in bars (>= 5.0). Default is 48.0.
    pub highpass_period: f64,
    /// Two-pole SuperSmoother cutoff period in bars (>= 2.0). Default is 10.0.
    pub smoothing_period: f64,
    /// Divide each cosine/sine contribution by period before summing/squaring. Default is true.
    pub spectral_dilation_compensation: bool,
    /// Cube normalized display heatmap intensities after centroid calculation. Default is false.
    pub enhance_resolution: bool,
    /// Bar component to extract (`None` defaults to `Close` and is omitted from the mnemonic).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract (`None` defaults to `Mid` and is omitted from the mnemonic).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract (`None` defaults to `Price` and is omitted from the mnemonic).
    pub trade_component: Option<TradeComponent>,
}

pub type Params = DiscreteFourierTransformSpectrumParams;

impl Default for DiscreteFourierTransformSpectrumParams {
    fn default() -> Self {
        Self {
            min_period: 10,
            max_period: 48,
            window_length: 48,
            highpass_period: 48.0,
            smoothing_period: 10.0,
            spectral_dilation_compensation: true,
            enhance_resolution: false,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Returns default Ehlers Listing 9-1 parameters.
pub fn default_params() -> DiscreteFourierTransformSpectrumParams {
    DiscreteFourierTransformSpectrumParams::default()
}

/// Formats a float parameter cleanly for mnemonics without `-0`.
pub fn format_number(value: f64) -> String {
    let normalized = if value == 0.0 { 0.0 } else { value };
    format!("{normalized}")
}

fn format_mnemonic(
    prefix: &str,
    include_enhance: bool,
    params: &DiscreteFourierTransformSpectrumParams,
) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let comp_suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    let sdc_suffix = if params.spectral_dilation_compensation {
        ""
    } else {
        ", no-sdc"
    };
    let enh_suffix = if include_enhance && params.enhance_resolution {
        ", enh"
    } else {
        ""
    };
    let hp_text = format_number(params.highpass_period);
    let lp_text = format_number(params.smoothing_period);
    format!(
        "{prefix}({}, {}, {}, {hp_text}, {lp_text}{sdc_suffix}{enh_suffix}{comp_suffix})",
        params.min_period, params.max_period, params.window_length
    )
}

/// Builds the primary heatmap mnemonic for `DiscreteFourierTransformSpectrum`.
pub fn mnemonic(params: &DiscreteFourierTransformSpectrumParams) -> String {
    format_mnemonic("dftps", true, params)
}

/// Builds the dominant-cycle mnemonic for `DiscreteFourierTransformSpectrum`.
pub fn dominant_cycle_mnemonic(params: &DiscreteFourierTransformSpectrumParams) -> String {
    format_mnemonic("dftps-dc", false, params)
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

/// Snapshot of committed internal roofing, spectral, and centroid state.
#[derive(Debug, Clone, PartialEq)]
pub struct SpectrumState {
    pub high_pass: f64,
    pub filtered: f64,
    pub raw_power: Vec<f64>,
    pub normalization_maximum: f64,
    pub internal_dominant_cycle: f64,
}

/// Validated numeric parameters resolved at construction time.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct ResolvedParams {
    pub min_period: usize,
    pub max_period: usize,
    pub window_length: usize,
    pub highpass_period: f64,
    pub smoothing_period: f64,
    pub spectral_dilation_compensation: bool,
    pub enhance_resolution: bool,
}

// ---------------------------------------------------------------------------
// Internal two-pole roofing filter (Book Equation 3-3 / Listing 9-1)
// ---------------------------------------------------------------------------

struct RoofFilter {
    hp0: f64,
    hp1: f64,
    hp2: f64,
    ss1: f64,
    ss2: f64,
    ss3: f64,
    x1: f64,
    x2: f64,
    h1: f64,
    h2: f64,
    f1: f64,
    f2: f64,
}

impl RoofFilter {
    fn new(highpass_period: f64, smoothing_period: f64) -> Result<Self, String> {
        if !highpass_period.is_finite() || highpass_period < 5.0 {
            return Err("highpass_period must be a finite real >= 5".to_string());
        }
        if !smoothing_period.is_finite() || smoothing_period < 2.0 {
            return Err("smoothing_period must be a finite real >= 2".to_string());
        }

        let angle = TWO_PI / (SQRT_TWO * highpass_period);
        let cosine = angle.cos();
        let alpha = (cosine + angle.sin() - 1.0) / cosine;
        let pole = 1.0 - alpha;
        let attenuation = (-SQRT_TWO * PI / smoothing_period).exp();
        let hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
        let hp1 = 2.0 * pole;
        let hp2 = pole * pole;
        let ss2 = 2.0 * attenuation * (SQRT_TWO * PI / smoothing_period).cos();
        let ss3 = -attenuation * attenuation;
        let ss1 = 1.0 - ss2 - ss3;

        if !(0.0 < pole && pole < 1.0 && 0.0 < attenuation && attenuation < 1.0 && ss1 > 0.0)
            || !hp0.is_finite()
            || !hp1.is_finite()
            || !hp2.is_finite()
            || !ss1.is_finite()
            || !ss2.is_finite()
            || !ss3.is_finite()
        {
            return Err(
                "periods must produce finite resolvable stable coefficients".to_string(),
            );
        }

        Ok(Self {
            hp0,
            hp1,
            hp2,
            ss1,
            ss2,
            ss3,
            x1: 0.0,
            x2: 0.0,
            h1: 0.0,
            h2: 0.0,
            f1: 0.0,
            f2: 0.0,
        })
    }

    fn reset(&mut self) {
        self.x1 = 0.0;
        self.x2 = 0.0;
        self.h1 = 0.0;
        self.h2 = 0.0;
        self.f1 = 0.0;
        self.f2 = 0.0;
    }

    fn candidate(&self, price: f64) -> Result<(f64, f64), String> {
        let high = self.hp0 * (price - 2.0 * self.x1 + self.x2)
            + self.hp1 * self.h1
            - self.hp2 * self.h2;
        let filtered = self.ss1 * (high + self.h1) / 2.0
            + self.ss2 * self.f1
            + self.ss3 * self.f2;
        if !high.is_finite() || !filtered.is_finite() {
            return Err("roofing arithmetic overflowed".to_string());
        }
        Ok((high, filtered))
    }

    fn commit(&mut self, price: f64, high: f64, filtered: f64) {
        self.x2 = self.x1;
        self.x1 = price;
        self.h2 = self.h1;
        self.h1 = high;
        self.f2 = self.f1;
        self.f1 = filtered;
    }
}

// ---------------------------------------------------------------------------
// Indicator implementation
// ---------------------------------------------------------------------------

/// Ehlers' 2013 Book Listing 9-1 Discrete Fourier Transform Spectrum indicator.
pub struct DiscreteFourierTransformSpectrum {
    resolved: ResolvedParams,
    metadata: Metadata,
    roof: RoofFilter,
    periods: Vec<usize>,
    cosines: Vec<Vec<f64>>,
    sines: Vec<Vec<f64>>,
    sample_count: usize,
    history: Vec<f64>,
    raw: Vec<f64>,
    maximum: f64,
    dominant: f64,
    has_dominant: bool,
    last_output: SpectrumOutput,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl DiscreteFourierTransformSpectrum {
    /// Creates a new `DiscreteFourierTransformSpectrum` with the supplied parameters.
    pub fn new(params: &DiscreteFourierTransformSpectrumParams) -> Result<Self, String> {
        const INVALID: &str = "invalid discrete Fourier transform spectrum parameters";

        if params.min_period < 3 {
            return Err(format!("{INVALID}: min_period must be an integer >= 3"));
        }
        if params.max_period < params.min_period {
            return Err(format!(
                "{INVALID}: max_period must be an integer >= min_period"
            ));
        }
        if params.window_length < 2 {
            return Err(format!("{INVALID}: window_length must be an integer >= 2"));
        }

        let roof = RoofFilter::new(params.highpass_period, params.smoothing_period)
            .map_err(|err| format!("{INVALID}: {err}"))?;

        let min_period = params.min_period as usize;
        let max_period = params.max_period as usize;
        let window_length = params.window_length as usize;

        let resolved = ResolvedParams {
            min_period,
            max_period,
            window_length,
            highpass_period: params.highpass_period,
            smoothing_period: params.smoothing_period,
            spectral_dilation_compensation: params.spectral_dilation_compensation,
            enhance_resolution: params.enhance_resolution,
        };

        let periods: Vec<usize> = (min_period..=max_period).collect();
        let mut cosines = Vec::with_capacity(periods.len());
        let mut sines = Vec::with_capacity(periods.len());
        for &period in &periods {
            let period_f = period as f64;
            let mut cos_row = Vec::with_capacity(window_length);
            let mut sin_row = Vec::with_capacity(window_length);
            for lag in 0..window_length {
                let angle = TWO_PI * (lag as f64) / period_f;
                cos_row.push(angle.cos());
                sin_row.push(angle.sin());
            }
            cosines.push(cos_row);
            sines.push(sin_row);
        }

        let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let mnemonic_text = mnemonic(params);
        let description_text = format!("Discrete Fourier transform spectrum {mnemonic_text}");
        let mnemonic_dc = dominant_cycle_mnemonic(params);
        let description_dc =
            format!("Discrete Fourier transform spectrum dominant cycle {mnemonic_dc}");

        let metadata = build_metadata(
            Identifier::DiscreteFourierTransformSpectrum,
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
            metadata,
            roof,
            raw: vec![0.0; periods.len()],
            periods,
            cosines,
            sines,
            sample_count: 0,
            history: vec![0.0; window_length],
            maximum: 0.0,
            dominant: 0.0,
            has_dominant: false,
            last_output: initial_output,
            bar_func: bar_component_value(bar_comp),
            quote_func: quote_component_value(quote_comp),
            trade_func: trade_component_value(trade_comp),
        })
    }

    /// Creates a new `DiscreteFourierTransformSpectrum` with default Listing 9-1 parameters.
    pub fn new_default() -> Result<Self, String> {
        Self::new(&DiscreteFourierTransformSpectrumParams::default())
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

    /// Returns the number of warmup observations (`window_length - 1`).
    pub fn warmup_bars(&self) -> usize {
        self.resolved.window_length - 1
    }

    /// Returns the number of accepted finite observations seen since construction or reset.
    pub fn samples_seen(&self) -> usize {
        self.sample_count
    }

    /// Returns an owned snapshot of committed internal state.
    pub fn state(&self) -> SpectrumState {
        SpectrumState {
            high_pass: self.roof.h1,
            filtered: self.roof.f1,
            raw_power: self.raw.clone(),
            normalization_maximum: self.maximum,
            internal_dominant_cycle: self.dominant,
        }
    }

    /// Returns an owned copy of the most recent accepted output.
    pub fn output(&self) -> SpectrumOutput {
        self.last_output.clone()
    }

    /// Resets all filter, window, spectral, AGC, and centroid state to initial values.
    pub fn reset(&mut self) {
        self.roof.reset();
        self.sample_count = 0;
        self.history.fill(0.0);
        self.raw.fill(0.0);
        self.maximum = 0.0;
        self.dominant = 0.0;
        self.has_dominant = false;
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
            return Err("price must be finite or NaN (missing)".to_string());
        }
        let min_f = self.resolved.min_period as f64;
        let max_f = self.resolved.max_period as f64;
        if sample.is_nan() {
            return Ok(SpectrumOutput {
                heatmap: Heatmap::empty(time, min_f, max_f, 1.0),
                dominant_cycle: f64::NAN,
            });
        }

        let (high, filtered) = self.roof.candidate(sample)?;

        let mut next_history = Vec::with_capacity(self.resolved.window_length);
        next_history.push(filtered);
        next_history.extend_from_slice(&self.history[..self.resolved.window_length - 1]);

        let mut next_raw = Vec::with_capacity(self.periods.len());
        for (bin_index, &period) in self.periods.iter().enumerate() {
            let compensation = if self.resolved.spectral_dilation_compensation {
                period as f64
            } else {
                1.0
            };
            let mut cosine_part = 0.0;
            let mut sine_part = 0.0;
            let cos_row = &self.cosines[bin_index];
            let sin_row = &self.sines[bin_index];
            for lag in 0..self.resolved.window_length {
                cosine_part += next_history[lag] * cos_row[lag] / compensation;
                sine_part += next_history[lag] * sin_row[lag] / compensation;
            }
            next_raw.push(cosine_part * cosine_part + sine_part * sine_part);
        }

        let mut maximum = AGC_DECAY * self.maximum;
        for &power in &next_raw {
            if !power.is_finite() || power < 0.0 {
                return Err("spectral power arithmetic overflowed".to_string());
            }
            if power > maximum {
                maximum = power;
            }
        }

        let mut normalized = Vec::with_capacity(next_raw.len());
        for &power in &next_raw {
            normalized.push(if maximum > 0.0 { power / maximum } else { 0.0 });
        }

        let mut sp = 0.0;
        let mut spx = 0.0;
        for (&period, &power) in self.periods.iter().zip(normalized.iter()) {
            if power >= CENTROID_THRESHOLD {
                spx += (period as f64) * power;
                sp += power;
            }
        }

        let mut dominant = self.dominant;
        let mut has_dominant = self.has_dominant;
        if sp != 0.0 {
            dominant = spx / sp;
            has_dominant = true;
        }
        if !dominant.is_finite() {
            return Err("centroid arithmetic overflowed".to_string());
        }

        let values: Vec<f64> = if self.resolved.enhance_resolution {
            normalized.iter().map(|&item| item * item * item).collect()
        } else {
            normalized
        };

        self.roof.commit(sample, high, filtered);
        self.history = next_history;
        self.raw = next_raw;
        self.maximum = maximum;
        self.dominant = dominant;
        self.has_dominant = has_dominant;
        self.sample_count += 1;

        let current_output = if self.sample_count >= self.resolved.window_length {
            let mut value_min = values[0];
            let mut value_max = values[0];
            for &item in &values[1..] {
                if item < value_min {
                    value_min = item;
                }
                if item > value_max {
                    value_max = item;
                }
            }
            SpectrumOutput {
                heatmap: Heatmap::new(time, min_f, max_f, 1.0, value_min, value_max, values),
                dominant_cycle: if has_dominant { dominant } else { f64::NAN },
            }
        } else {
            SpectrumOutput {
                heatmap: Heatmap::empty(time, min_f, max_f, 1.0),
                dominant_cycle: f64::NAN,
            }
        };

        self.last_output = current_output.clone();
        Ok(current_output)
    }

    /// Updates the indicator with a scalar sample and timestamp, panicking on invalid input/overflow.
    pub fn update(&mut self, sample: f64, time: i64) -> SpectrumOutput {
        self.try_update(sample, time).unwrap_or_else(|err| {
            panic!("DiscreteFourierTransformSpectrum::update failed: {err}")
        })
    }

    fn wrap_output(time: i64, output: SpectrumOutput) -> Output {
        vec![
            Box::new(output.heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, output.dominant_cycle)),
        ]
    }
}

impl Indicator for DiscreteFourierTransformSpectrum {
    fn is_primed(&self) -> bool {
        self.sample_count >= self.resolved.window_length
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
    use crate::indicators::core::outputs::shape::Shape;
    use crate::indicators::john_ehlers::discrete_fourier_transform_spectrum::testdata;

    const TOLERANCE: f64 = 1e-13;
    const ORACLE_TOLERANCE: f64 = 2e-11;

    fn assert_close_tol(actual: f64, expected: f64, tolerance: f64, label: &str) {
        if expected.is_nan() {
            assert!(actual.is_nan(), "{label}: expected NaN, got {actual}");
            return;
        }
        assert!(
            actual.is_finite(),
            "{label}: expected finite {expected}, got {actual}"
        );
        let bound = tolerance.max(expected.abs() * tolerance);
        let diff = (actual - expected).abs();
        assert!(
            diff <= bound,
            "{label}: actual={actual}, expected={expected}, diff={diff}, bound={bound}"
        );
    }

    fn assert_close(actual: f64, expected: f64, label: &str) {
        assert_close_tol(actual, expected, TOLERANCE, label);
    }

    fn assert_same_output(actual: &SpectrumOutput, expected: &SpectrumOutput, label: &str) {
        assert_eq!(actual.heatmap.time, expected.heatmap.time, "{label}: time");
        assert_eq!(
            actual.heatmap.parameter_first, expected.heatmap.parameter_first,
            "{label}: parameter_first"
        );
        assert_eq!(
            actual.heatmap.parameter_last, expected.heatmap.parameter_last,
            "{label}: parameter_last"
        );
        assert_eq!(
            actual.heatmap.parameter_resolution, expected.heatmap.parameter_resolution,
            "{label}: parameter_resolution"
        );
        assert_eq!(
            actual.heatmap.is_empty(),
            expected.heatmap.is_empty(),
            "{label}: is_empty"
        );
        assert_close(actual.heatmap.value_min, expected.heatmap.value_min, label);
        assert_close(actual.heatmap.value_max, expected.heatmap.value_max, label);
        assert_eq!(
            actual.heatmap.values.len(),
            expected.heatmap.values.len(),
            "{label}: values len"
        );
        for (bin, (&lhs, &rhs)) in actual
            .heatmap
            .values
            .iter()
            .zip(expected.heatmap.values.iter())
            .enumerate()
        {
            assert_close(lhs, rhs, &format!("{label}[{bin}]"));
        }
        assert_close(actual.dominant_cycle, expected.dominant_cycle, label);
    }

    struct ListingOracle {
        params: ResolvedParams,
        prices: Vec<f64>,
        hp: Vec<f64>,
        filt: Vec<f64>,
        max_power: f64,
        dominant: f64,
        identified: bool,
    }

    impl ListingOracle {
        fn new(params: ResolvedParams) -> Self {
            Self {
                params,
                prices: Vec::new(),
                hp: Vec::new(),
                filt: Vec::new(),
                max_power: 0.0,
                dominant: 0.0,
                identified: false,
            }
        }

        fn lookback(series: &[f64], lag: usize) -> f64 {
            if series.len() > lag {
                series[series.len() - 1 - lag]
            } else {
                0.0
            }
        }

        fn update(&mut self, price: f64) -> (Vec<f64>, f64, f64, f64) {
            let theta = SQRT_2 * PI / self.params.highpass_period;
            let alpha = (theta.cos() + theta.sin() - 1.0) / theta.cos();
            let high_pass = (1.0 - alpha / 2.0)
                * (1.0 - alpha / 2.0)
                * (price - 2.0 * Self::lookback(&self.prices, 0) + Self::lookback(&self.prices, 1))
                + 2.0 * (1.0 - alpha) * Self::lookback(&self.hp, 0)
                - (1.0 - alpha) * (1.0 - alpha) * Self::lookback(&self.hp, 1);
            let a1 = (-SQRT_2 * PI / self.params.smoothing_period).exp();
            let c2 = 2.0 * a1 * (SQRT_2 * PI / self.params.smoothing_period).cos();
            let c3 = -a1 * a1;
            let c1 = 1.0 - c2 - c3;
            let filtered = c1 * (high_pass + Self::lookback(&self.hp, 0)) / 2.0
                + c2 * Self::lookback(&self.filt, 0)
                + c3 * Self::lookback(&self.filt, 1);
            self.prices.push(price);
            self.hp.push(high_pass);
            self.filt.push(filtered);

            let count_periods = self.params.max_period - self.params.min_period + 1;
            let mut powers = Vec::with_capacity(count_periods);
            for period in self.params.min_period..=self.params.max_period {
                let comp = if self.params.spectral_dilation_compensation {
                    period as f64
                } else {
                    1.0
                };
                let mut cosine = 0.0;
                let mut sine = 0.0;
                for lag in 0..self.params.window_length {
                    let sample = Self::lookback(&self.filt, lag);
                    let angle = 2.0 * PI * (lag as f64) / (period as f64);
                    cosine += sample * angle.cos() / comp;
                    sine += sample * angle.sin() / comp;
                }
                powers.push(cosine * cosine + sine * sine);
            }

            self.max_power = 0.995 * self.max_power;
            for &power in &powers {
                if power > self.max_power {
                    self.max_power = power;
                }
            }
            for power in &mut powers {
                *power = if self.max_power != 0.0 {
                    *power / self.max_power
                } else {
                    0.0
                };
            }

            let mut sp = 0.0;
            let mut spx = 0.0;
            for (offset, &power) in powers.iter().enumerate() {
                if power >= 0.5 {
                    let period = (self.params.min_period + offset) as f64;
                    spx += period * power;
                    sp += power;
                }
            }
            if sp != 0.0 {
                self.dominant = spx / sp;
                self.identified = true;
            }

            let values: Vec<f64> = if self.params.enhance_resolution {
                powers.iter().map(|&item| item.powi(3)).collect()
            } else {
                powers
            };
            let dominant = if self.identified {
                self.dominant
            } else {
                f64::NAN
            };
            (values, dominant, high_pass, filtered)
        }
    }

    struct ReferenceCase {
        name: &'static str,
        params: DiscreteFourierTransformSpectrumParams,
        expected_heatmap: Vec<Vec<f64>>,
        expected_min: Vec<f64>,
        expected_max: Vec<f64>,
        expected_dc: Vec<f64>,
    }

    fn reference_cases() -> Vec<ReferenceCase> {
        vec![
            ReferenceCase {
                name: "P010_P148_HP48_LP10_SDC1_E0_W48",
                params: DiscreteFourierTransformSpectrumParams::default(),
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w48_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w48_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w48_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w48_dominant_cycle(),
            },
            ReferenceCase {
                name: "P010_P148_HP48_LP10_SDC0_E0_W48",
                params: DiscreteFourierTransformSpectrumParams {
                    spectral_dilation_compensation: false,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc0_e0_w48_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc0_e0_w48_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc0_e0_w48_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc0_e0_w48_dominant_cycle(),
            },
            ReferenceCase {
                name: "P010_P148_HP48_LP10_SDC1_E1_W48",
                params: DiscreteFourierTransformSpectrumParams {
                    enhance_resolution: true,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc1_e1_w48_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc1_e1_w48_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc1_e1_w48_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc1_e1_w48_dominant_cycle(),
            },
            ReferenceCase {
                name: "P010_P148_HP48_LP10_SDC0_E1_W48",
                params: DiscreteFourierTransformSpectrumParams {
                    spectral_dilation_compensation: false,
                    enhance_resolution: true,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc0_e1_w48_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc0_e1_w48_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc0_e1_w48_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc0_e1_w48_dominant_cycle(),
            },
            ReferenceCase {
                name: "P010_P148_HP48_LP10_SDC1_E0_W24",
                params: DiscreteFourierTransformSpectrumParams {
                    window_length: 24,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w24_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w24_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w24_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w24_dominant_cycle(),
            },
            ReferenceCase {
                name: "P010_P148_HP48_LP10_SDC1_E0_W96",
                params: DiscreteFourierTransformSpectrumParams {
                    window_length: 96,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w96_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w96_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w96_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w96_dominant_cycle(),
            },
            ReferenceCase {
                name: "P010_P148_HP48_LP10_SDC1_E0_W252",
                params: DiscreteFourierTransformSpectrumParams {
                    window_length: 252,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w252_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w252_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w252_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_w252_dominant_cycle(),
            },
            ReferenceCase {
                name: "P010_P148_HP48_LP10_5_SDC1_E0_W48",
                params: DiscreteFourierTransformSpectrumParams {
                    smoothing_period: 10.5,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_5_sdc1_e0_w48_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_5_sdc1_e0_w48_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_5_sdc1_e0_w48_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_5_sdc1_e0_w48_dominant_cycle(),
            },
            ReferenceCase {
                name: "P010_P148_HP80_LP20_SDC1_E0_W48",
                params: DiscreteFourierTransformSpectrumParams {
                    highpass_period: 80.0,
                    smoothing_period: 20.0,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp80_lp20_sdc1_e0_w48_heatmap(),
                expected_min: testdata::expected_p010_p148_hp80_lp20_sdc1_e0_w48_min(),
                expected_max: testdata::expected_p010_p148_hp80_lp20_sdc1_e0_w48_max(),
                expected_dc: testdata::expected_p010_p148_hp80_lp20_sdc1_e0_w48_dominant_cycle(),
            },
            ReferenceCase {
                name: "P06_P124_HP24_LP6_SDC1_E0_W24",
                params: DiscreteFourierTransformSpectrumParams {
                    min_period: 6,
                    max_period: 24,
                    window_length: 24,
                    highpass_period: 24.0,
                    smoothing_period: 6.0,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p06_p124_hp24_lp6_sdc1_e0_w24_heatmap(),
                expected_min: testdata::expected_p06_p124_hp24_lp6_sdc1_e0_w24_min(),
                expected_max: testdata::expected_p06_p124_hp24_lp6_sdc1_e0_w24_max(),
                expected_dc: testdata::expected_p06_p124_hp24_lp6_sdc1_e0_w24_dominant_cycle(),
            },
            ReferenceCase {
                name: "P020_P120_HP48_LP10_SDC1_E0_W48",
                params: DiscreteFourierTransformSpectrumParams {
                    min_period: 20,
                    max_period: 20,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p020_p120_hp48_lp10_sdc1_e0_w48_heatmap(),
                expected_min: testdata::expected_p020_p120_hp48_lp10_sdc1_e0_w48_min(),
                expected_max: testdata::expected_p020_p120_hp48_lp10_sdc1_e0_w48_max(),
                expected_dc: testdata::expected_p020_p120_hp48_lp10_sdc1_e0_w48_dominant_cycle(),
            },
            ReferenceCase {
                name: "P024_P160_HP60_LP12_SDC1_E0_W60",
                params: DiscreteFourierTransformSpectrumParams {
                    min_period: 24,
                    max_period: 60,
                    window_length: 60,
                    highpass_period: 60.0,
                    smoothing_period: 12.0,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p024_p160_hp60_lp12_sdc1_e0_w60_heatmap(),
                expected_min: testdata::expected_p024_p160_hp60_lp12_sdc1_e0_w60_min(),
                expected_max: testdata::expected_p024_p160_hp60_lp12_sdc1_e0_w60_max(),
                expected_dc: testdata::expected_p024_p160_hp60_lp12_sdc1_e0_w60_dominant_cycle(),
            },
        ]
    }

    #[test]
    fn test_all_12_reference_cases_and_listing_oracle() {
        let prices = testdata::testdata::test_input();
        for case in reference_cases() {
            let mut indicator = DiscreteFourierTransformSpectrum::new(&case.params).unwrap();
            let mut oracle = ListingOracle::new(indicator.parameters());
            let warmup = indicator.warmup_bars();

            assert_eq!(case.expected_heatmap.len(), prices.len());
            assert_eq!(case.expected_min.len(), prices.len());
            assert_eq!(case.expected_max.len(), prices.len());
            assert_eq!(case.expected_dc.len(), prices.len());

            for (bar_index, &price) in prices.iter().enumerate() {
                let result = indicator.update(price, bar_index as i64);
                let (oracle_values, oracle_dc, oracle_h, oracle_f) = oracle.update(price);
                let state = indicator.state();

                assert_close_tol(
                    state.high_pass,
                    oracle_h,
                    ORACLE_TOLERANCE,
                    &format!("{}[{bar_index}] hp", case.name),
                );
                assert_close_tol(
                    state.filtered,
                    oracle_f,
                    ORACLE_TOLERANCE,
                    &format!("{}[{bar_index}] filt", case.name),
                );
                assert_eq!(
                    indicator.is_primed(),
                    bar_index >= warmup,
                    "{}[{bar_index}] is_primed",
                    case.name
                );
                assert_eq!(
                    result.heatmap.is_empty(),
                    bar_index < warmup,
                    "{}[{bar_index}] is_empty",
                    case.name
                );
                assert_eq!(result.heatmap.time, bar_index as i64);
                assert_eq!(result.heatmap.parameter_first, case.params.min_period as f64);
                assert_eq!(result.heatmap.parameter_last, case.params.max_period as f64);
                assert_eq!(result.heatmap.parameter_resolution, 1.0);

                if bar_index < warmup {
                    assert!(
                        case.expected_heatmap[bar_index].iter().all(|item| item.is_nan()),
                        "{}[{bar_index}] expected NaN warmup row",
                        case.name
                    );
                    assert!(result.heatmap.value_min.is_nan());
                    assert!(result.heatmap.value_max.is_nan());
                    assert!(result.dominant_cycle.is_nan());
                    assert!(case.expected_min[bar_index].is_nan());
                    assert!(case.expected_max[bar_index].is_nan());
                    assert!(case.expected_dc[bar_index].is_nan());
                    continue;
                }

                assert_eq!(
                    result.heatmap.values.len(),
                    case.expected_heatmap[bar_index].len()
                );
                for (bin_index, (&actual_v, &expected_v)) in result
                    .heatmap
                    .values
                    .iter()
                    .zip(case.expected_heatmap[bar_index].iter())
                    .enumerate()
                {
                    assert_close(
                        actual_v,
                        expected_v,
                        &format!("{}[{bar_index}][{bin_index}]", case.name),
                    );
                    assert_close_tol(
                        actual_v,
                        oracle_values[bin_index],
                        ORACLE_TOLERANCE,
                        &format!("{}[{bar_index}][{bin_index}] oracle", case.name),
                    );
                }
                assert_close(
                    result.heatmap.value_min,
                    case.expected_min[bar_index],
                    &format!("{}[{bar_index}] min", case.name),
                );
                assert_close(
                    result.heatmap.value_max,
                    case.expected_max[bar_index],
                    &format!("{}[{bar_index}] max", case.name),
                );
                assert_close(
                    result.dominant_cycle,
                    case.expected_dc[bar_index],
                    &format!("{}[{bar_index}] dc", case.name),
                );
                assert_close_tol(
                    result.dominant_cycle,
                    oracle_dc,
                    ORACLE_TOLERANCE,
                    &format!("{}[{bar_index}] oracle dc", case.name),
                );
            }
        }
    }

    #[test]
    fn test_missing_infinity_and_overflow_preserve_state() {
        let prices = testdata::testdata::test_input();
        let mut indicator = DiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut control = DiscreteFourierTransformSpectrum::new_default().unwrap();
        for (bar_index, &price) in prices[..120].iter().enumerate() {
            indicator.update(price, bar_index as i64);
            control.update(price, bar_index as i64);
        }
        let saved_state = indicator.state();
        let saved_output = indicator.output();
        let saved_count = indicator.samples_seen();

        let nan_result = indicator.update(f64::NAN, 999);
        assert!(nan_result.heatmap.is_empty());
        assert_eq!(nan_result.heatmap.time, 999);
        assert!(nan_result.dominant_cycle.is_nan());
        assert_eq!(indicator.state(), saved_state);
        assert_same_output(&indicator.output(), &saved_output, "after NaN");
        assert_eq!(indicator.samples_seen(), saved_count);

        for bad_sample in [f64::INFINITY, f64::NEG_INFINITY, 1e308] {
            assert!(indicator.try_update(bad_sample, 1000).is_err());
            assert_eq!(indicator.state(), saved_state);
            assert_same_output(&indicator.output(), &saved_output, "after invalid sample");
            assert_eq!(indicator.samples_seen(), saved_count);
        }

        let actual = indicator.update(prices[120], 120);
        let expected = control.update(prices[120], 120);
        assert_same_output(&actual, &expected, "resumed bar 120");
    }

    #[test]
    fn test_owned_snapshot_and_reset_replay() {
        let prices = testdata::testdata::test_input();
        let mut indicator = DiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut outputs: Vec<SpectrumOutput> = prices
            .iter()
            .enumerate()
            .map(|(index, &price)| indicator.update(price, index as i64))
            .collect();
        let snapshot = outputs.last().unwrap().heatmap.values.clone();
        indicator.update(101.0, 252);
        assert_eq!(outputs.last().unwrap().heatmap.values, snapshot);

        outputs.last_mut().unwrap().heatmap.values[0] = -999.0;
        let mut control = DiscreteFourierTransformSpectrum::new_default().unwrap();
        for (index, &price) in prices.iter().enumerate() {
            control.update(price, index as i64);
        }
        control.update(101.0, 252);
        let actual = indicator.update(102.0, 253);
        let expected = control.update(102.0, 253);
        assert_same_output(&actual, &expected, "post-mutation bar");

        indicator.reset();
        assert!(!indicator.is_primed());
        assert_eq!(indicator.samples_seen(), 0);
        let mut fresh = DiscreteFourierTransformSpectrum::new_default().unwrap();
        for (index, &price) in prices.iter().enumerate() {
            let replay_out = indicator.update(price, index as i64);
            let fresh_out = fresh.update(price, index as i64);
            assert_same_output(&replay_out, &fresh_out, &format!("replay[{index}]"));
        }
    }

    #[test]
    fn test_zero_input_and_single_bin_boundary() {
        let mut indicator = DiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut last = indicator.output();
        for step in 0..(indicator.warmup_bars() + 4) {
            last = indicator.update(0.0, step as i64);
        }
        assert!(indicator.is_primed());
        assert_eq!(last.heatmap.values, vec![0.0; 39]);
        assert!(last.dominant_cycle.is_nan());
        assert_eq!(indicator.state().normalization_maximum, 0.0);

        let prices = testdata::testdata::test_input();
        let mut single = DiscreteFourierTransformSpectrum::new(&DiscreteFourierTransformSpectrumParams {
            min_period: 20,
            max_period: 20,
            ..Default::default()
        })
        .unwrap();
        for (index, &price) in prices.iter().enumerate() {
            let out = single.update(price, index as i64);
            if !out.dominant_cycle.is_nan() {
                assert_close(out.dominant_cycle, 20.0, "single-bin dc");
            }
        }
    }

    #[test]
    fn test_fourier_compensation_and_display_enhancement() {
        let prices = testdata::testdata::test_input();
        let mut compensated = DiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut plain = DiscreteFourierTransformSpectrum::new(&DiscreteFourierTransformSpectrumParams {
            spectral_dilation_compensation: false,
            ..Default::default()
        })
        .unwrap();
        let mut enhanced = DiscreteFourierTransformSpectrum::new(&DiscreteFourierTransformSpectrumParams {
            enhance_resolution: true,
            ..Default::default()
        })
        .unwrap();

        for (index, &price) in prices.iter().enumerate() {
            let comp_out = compensated.update(price, index as i64);
            plain.update(price, index as i64);
            let enh_out = enhanced.update(price, index as i64);

            assert_eq!(compensated.state(), enhanced.state());
            assert_close(comp_out.dominant_cycle, enh_out.dominant_cycle, "enh dc");
            if !comp_out.heatmap.is_empty() {
                for (&base_v, &enh_v) in comp_out
                    .heatmap
                    .values
                    .iter()
                    .zip(enh_out.heatmap.values.iter())
                {
                    assert_close(enh_v, base_v * base_v * base_v, "enh cube");
                }
            }
        }

        let comp_state = compensated.state();
        let plain_state = plain.state();
        for (offset, period) in (10..=48).enumerate() {
            let scale = (period * period) as f64;
            assert_close_tol(
                comp_state.raw_power[offset],
                plain_state.raw_power[offset] / scale,
                ORACLE_TOLERANCE,
                &format!("sdc period {period}"),
            );
        }
    }

    #[test]
    fn test_sine_wave_selected_period() {
        let mut indicator = DiscreteFourierTransformSpectrum::new(&DiscreteFourierTransformSpectrumParams {
            window_length: 120,
            ..Default::default()
        })
        .unwrap();
        let mut last = indicator.output();
        for step in 0..1000 {
            let sample = (2.0 * PI * (step as f64) / 20.0).sin();
            last = indicator.update(sample, step);
        }
        let (peak_offset, _) = last
            .heatmap
            .values
            .iter()
            .enumerate()
            .max_by(|lhs, rhs| lhs.1.partial_cmp(rhs.1).unwrap())
            .unwrap();
        let peak_period = (peak_offset as i32) + 10;
        assert!(
            (peak_period - 20).abs() <= 1,
            "expected peak near 20, got {peak_period}"
        );
        assert!(
            last.dominant_cycle > 17.0 && last.dominant_cycle < 23.0,
            "expected dc near 20, got {}",
            last.dominant_cycle
        );
    }

    #[test]
    fn test_invalid_parameters() {
        let bad_cases = [
            DiscreteFourierTransformSpectrumParams {
                min_period: 2,
                ..Default::default()
            },
            DiscreteFourierTransformSpectrumParams {
                min_period: 0,
                ..Default::default()
            },
            DiscreteFourierTransformSpectrumParams {
                max_period: 9,
                ..Default::default()
            },
            DiscreteFourierTransformSpectrumParams {
                window_length: 1,
                ..Default::default()
            },
            DiscreteFourierTransformSpectrumParams {
                window_length: 0,
                ..Default::default()
            },
            DiscreteFourierTransformSpectrumParams {
                highpass_period: 4.0,
                ..Default::default()
            },
            DiscreteFourierTransformSpectrumParams {
                highpass_period: f64::NAN,
                ..Default::default()
            },
            DiscreteFourierTransformSpectrumParams {
                smoothing_period: 1.0,
                ..Default::default()
            },
            DiscreteFourierTransformSpectrumParams {
                smoothing_period: f64::INFINITY,
                ..Default::default()
            },
        ];
        for case in &bad_cases {
            assert!(
                DiscreteFourierTransformSpectrum::new(case).is_err(),
                "expected error for {case:?}"
            );
        }
    }

    #[test]
    fn test_metadata_mnemonics_and_entities() {
        let indicator = DiscreteFourierTransformSpectrum::new_default().unwrap();
        let meta = indicator.metadata();
        assert_eq!(meta.identifier, Identifier::DiscreteFourierTransformSpectrum);
        assert_eq!(meta.mnemonic, "dftps(10, 48, 48, 48, 10)");
        assert_eq!(
            meta.description,
            "Discrete Fourier transform spectrum dftps(10, 48, 48, 48, 10)"
        );
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(
            meta.outputs[0].kind,
            DiscreteFourierTransformSpectrumOutput::Value as i32
        );
        assert_eq!(meta.outputs[0].shape, Shape::Heatmap);
        assert_eq!(meta.outputs[0].mnemonic, "dftps(10, 48, 48, 48, 10)");
        assert_eq!(
            meta.outputs[1].kind,
            DiscreteFourierTransformSpectrumOutput::DominantCycle as i32
        );
        assert_eq!(meta.outputs[1].shape, Shape::Scalar);
        assert_eq!(meta.outputs[1].mnemonic, "dftps-dc(10, 48, 48, 48, 10)");
        assert_eq!(
            meta.outputs[1].description,
            "Discrete Fourier transform spectrum dominant cycle dftps-dc(10, 48, 48, 48, 10)"
        );

        let custom = DiscreteFourierTransformSpectrumParams {
            min_period: 6,
            max_period: 24,
            window_length: 30,
            highpass_period: 48.0,
            smoothing_period: 10.5,
            spectral_dilation_compensation: false,
            enhance_resolution: true,
            bar_component: Some(BarComponent::Median),
            quote_component: Some(QuoteComponent::Weighted),
            trade_component: Some(TradeComponent::Volume),
        };
        assert_eq!(
            mnemonic(&custom),
            "dftps(6, 24, 30, 48, 10.5, no-sdc, enh, hl/2, (bbs+aas)/(bs+as), v)"
        );
        assert_eq!(
            dominant_cycle_mnemonic(&custom),
            "dftps-dc(6, 24, 30, 48, 10.5, no-sdc, hl/2, (bbs+aas)/(bs+as), v)"
        );

        let mnemonic_cases: Vec<(DiscreteFourierTransformSpectrumParams, &str, &str)> = vec![
            (
                DiscreteFourierTransformSpectrumParams::default(),
                "dftps(10, 48, 48, 48, 10)",
                "dftps-dc(10, 48, 48, 48, 10)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    bar_component: Some(BarComponent::Open),
                    ..Default::default()
                },
                "dftps(10, 48, 48, 48, 10, o)",
                "dftps-dc(10, 48, 48, 48, 10, o)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    quote_component: Some(QuoteComponent::Bid),
                    ..Default::default()
                },
                "dftps(10, 48, 48, 48, 10, b)",
                "dftps-dc(10, 48, 48, 48, 10, b)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "dftps(10, 48, 48, 48, 10, v)",
                "dftps-dc(10, 48, 48, 48, 10, v)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    bar_component: Some(BarComponent::High),
                    quote_component: Some(QuoteComponent::Ask),
                    ..Default::default()
                },
                "dftps(10, 48, 48, 48, 10, h, a)",
                "dftps-dc(10, 48, 48, 48, 10, h, a)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    bar_component: Some(BarComponent::Low),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "dftps(10, 48, 48, 48, 10, l, v)",
                "dftps-dc(10, 48, 48, 48, 10, l, v)",
            ),
            (
                DiscreteFourierTransformSpectrumParams {
                    quote_component: Some(QuoteComponent::SpreadBp),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "dftps(10, 48, 48, 48, 10, spread bp, v)",
                "dftps-dc(10, 48, 48, 48, 10, spread bp, v)",
            ),
        ];
        for (params_case, expected_m, expected_dc_m) in &mnemonic_cases {
            assert_eq!(&mnemonic(params_case), expected_m);
            assert_eq!(&dominant_cycle_mnemonic(params_case), expected_dc_m);
        }

        assert_eq!(DiscreteFourierTransformSpectrumOutput::Value.as_str(), "value");
        assert_eq!(
            DiscreteFourierTransformSpectrumOutput::DominantCycle.as_str(),
            "dominantCycle"
        );
        assert!(DiscreteFourierTransformSpectrumOutput::Value.is_known());
        assert!(DiscreteFourierTransformSpectrumOutput::DominantCycle.is_known());

        let prices = testdata::testdata::test_input();
        let highs = testdata::testdata::test_high();
        let lows = testdata::testdata::test_low();
        let opens = testdata::testdata::test_open();
        let mut bar_ind = DiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut scalar_ind = DiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut quote_ind = DiscreteFourierTransformSpectrum::new_default().unwrap();
        let mut trade_ind = DiscreteFourierTransformSpectrum::new_default().unwrap();

        for index in 0..60 {
            let bar = Bar {
                time: index as i64,
                open: opens[index],
                high: highs[index],
                low: lows[index],
                close: prices[index],
                volume: 1000.0,
            };
            let scalar = Scalar::new(index as i64, prices[index]);
            let quote = Quote {
                time: index as i64,
                bid_price: prices[index] - 0.5,
                ask_price: prices[index] + 0.5,
                bid_size: 10.0,
                ask_size: 10.0,
            };
            let trade = Trade {
                time: index as i64,
                price: prices[index],
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
            for bin in 0..hm_bar.values.len() {
                assert_close(hm_bar.values[bin], hm_scalar.values[bin], "bar vs scalar");
                assert_close(hm_bar.values[bin], hm_quote.values[bin], "bar vs quote");
                assert_close(hm_bar.values[bin], hm_trade.values[bin], "bar vs trade");
            }
            assert_close(dc_bar.value, dc_scalar.value, "dc bar vs scalar");
            assert_close(dc_bar.value, dc_quote.value, "dc bar vs quote");
            assert_close(dc_bar.value, dc_trade.value, "dc bar vs trade");
        }
    }
}
