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
const INVERSE_SQRT_TWO: f64 = 1.0 / SQRT_TWO;
const DFT_LAG_START: usize = 3;
const POWER_SMOOTHING: f64 = 0.2;
const POWER_RETENTION: f64 = 0.8;
const BOOK_AGC_DECAY: f64 = 0.995;
const ARTICLE_THRESHOLD: f64 = 0.25;
const BOOK_THRESHOLD: f64 = 0.5;

/// Output indices for Ehlers' Autocorrelation Periodogram indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AutoCorrelationPeriodogramOutput {
    /// Periodogram heatmap column across `[min_period, max_period]`.
    Value = 1,
    /// Thresholded center-of-gravity dominant cycle period estimate.
    DominantCycle = 2,
}

impl AutoCorrelationPeriodogramOutput {
    /// Returns the camelCase name of this output kind.
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Value => "value",
            Self::DominantCycle => "dominantCycle",
        }
    }

    /// Indicates whether this output variant is recognized.
    pub fn is_known(&self) -> bool {
        matches!(self, Self::Value | Self::DominantCycle)
    }
}

impl std::fmt::Display for AutoCorrelationPeriodogramOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

/// Selects the Ehlers source profile for `AutoCorrelationPeriodogram`.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AutoCorrelationPeriodogramSource {
    /// TASC September 2016 "Measuring Market Cycles" sidebar profile (default).
    Article = 0,
    /// Corrected Cycle Analytics for Traders (2013), Listing 8-3 profile.
    Book = 1,
}

pub type Source = AutoCorrelationPeriodogramSource;

impl Default for AutoCorrelationPeriodogramSource {
    fn default() -> Self {
        Self::Article
    }
}

impl AutoCorrelationPeriodogramSource {
    /// Parses `"article"` or `"book"` into a source profile.
    pub fn from_str(text: &str) -> Option<Self> {
        match text {
            "article" => Some(Self::Article),
            "book" => Some(Self::Book),
            _ => None,
        }
    }

    /// Returns `"article"` or `"book"`.
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Article => "article",
            Self::Book => "book",
        }
    }
}

impl std::fmt::Display for AutoCorrelationPeriodogramSource {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

/// Parameters to create an `AutoCorrelationPeriodogram` instance.
///
/// `min_period: None` resolves to `8` for `Article` and `10` for `Book`.
/// `averaging_length: 0` selects lag-dependent Pearson windows (`M = lag`).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct AutoCorrelationPeriodogramParams {
    /// Source profile (`Article` default or `Book`).
    pub source: AutoCorrelationPeriodogramSource,
    /// Shortest spectrum bin and SuperSmoother cutoff in bars (`>= 3`), or `None` for source default.
    pub min_period: Option<i32>,
    /// Longest spectrum bin, high-pass cutoff, and largest lag in bars (`>= 5`, `> min_period`). Default is `48`.
    pub max_period: i32,
    /// Pearson correlation window length: `0` (`M = lag`) or `>= 2`. Default is `3`.
    pub averaging_length: i32,
    /// Cube normalized power before centroid (`Article`) or square display power after centroid (`Book`). Default is `false`.
    pub enhance_resolution: bool,
    /// Bar component to extract (`None` defaults to `Close` and is omitted from the mnemonic).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract (`None` defaults to `Mid` and is omitted from the mnemonic).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract (`None` defaults to `Price` and is omitted from the mnemonic).
    pub trade_component: Option<TradeComponent>,
}

pub type Params = AutoCorrelationPeriodogramParams;

impl Default for AutoCorrelationPeriodogramParams {
    fn default() -> Self {
        Self {
            source: AutoCorrelationPeriodogramSource::Article,
            min_period: None,
            max_period: 48,
            averaging_length: 3,
            enhance_resolution: false,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Returns default Ehlers 2016 article parameters.
pub fn default_params() -> AutoCorrelationPeriodogramParams {
    AutoCorrelationPeriodogramParams::default()
}

/// Resolves `None` to `8` for `Article` or `10` for `Book`.
pub fn resolve_min_period(
    source: AutoCorrelationPeriodogramSource,
    min_period: Option<i32>,
) -> i32 {
    match min_period {
        Some(explicit_min) => explicit_min,
        None => match source {
            AutoCorrelationPeriodogramSource::Article => 8,
            AutoCorrelationPeriodogramSource::Book => 10,
        },
    }
}

fn format_mnemonic(
    prefix: &str,
    include_book_enhance: bool,
    params: &AutoCorrelationPeriodogramParams,
) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let comp_suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    let resolved_min = resolve_min_period(params.source, params.min_period);
    let is_book = params.source == AutoCorrelationPeriodogramSource::Book;
    let source_tag = if is_book { ", book" } else { "" };
    let enh_tag = if params.enhance_resolution && (!is_book || include_book_enhance) {
        ", enh"
    } else {
        ""
    };
    format!(
        "{prefix}({resolved_min}, {}, {}{source_tag}{enh_tag}{comp_suffix})",
        params.max_period, params.averaging_length
    )
}

/// Builds the primary heatmap mnemonic for `AutoCorrelationPeriodogram`.
pub fn mnemonic(params: &AutoCorrelationPeriodogramParams) -> String {
    format_mnemonic("acp", true, params)
}

/// Builds the dominant-cycle mnemonic for `AutoCorrelationPeriodogram`.
pub fn dominant_cycle_mnemonic(params: &AutoCorrelationPeriodogramParams) -> String {
    format_mnemonic("acp-dc", false, params)
}

/// Periodogram heatmap column and dominant cycle period returned per observation.
#[derive(Debug, Clone)]
pub struct PeriodogramOutput {
    /// Periodogram heatmap column across `[min_period, max_period]`.
    pub heatmap: Heatmap,
    /// Thresholded center-of-gravity dominant cycle period estimate.
    pub dominant_cycle: f64,
}

/// Snapshot of committed internal roofing, correlation, spectral, and centroid state.
#[derive(Debug, Clone, PartialEq)]
pub struct PeriodogramState {
    /// Most recent committed high-pass output.
    pub high_pass: f64,
    /// Most recent committed SuperSmoother output.
    pub filtered: f64,
    /// Committed Pearson correlations across lags `0..=max_period`.
    pub correlations: Vec<f64>,
    /// Committed fourth-power smoothed spectral power across `[min_period, max_period]`.
    pub smoothed_power: Vec<f64>,
    /// Current normalization maximum.
    pub normalization_maximum: f64,
    /// Committed internal dominant cycle estimate.
    pub internal_dominant_cycle: f64,
}

/// Validated numeric parameters resolved at construction time.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct ResolvedParams {
    /// Selected source profile.
    pub source: AutoCorrelationPeriodogramSource,
    /// Resolved shortest period in bars.
    pub min_period: usize,
    /// Resolved longest period in bars.
    pub max_period: usize,
    /// Pearson window length (`0` for lag-dependent `M = lag`).
    pub averaging_length: usize,
    /// Whether resolution enhancement is enabled.
    pub enhance_resolution: bool,
}

/// Ehlers' Autocorrelation Periodogram indicator (`Article` 2016 sidebar and `Book` Listing 8-3).
pub struct AutoCorrelationPeriodogram {
    resolved: ResolvedParams,
    metadata: Metadata,
    history_length: usize,
    bins_len: usize,
    lag_span: usize,
    hp0: f64,
    hp1: f64,
    hp2: f64,
    ss1: f64,
    ss2: f64,
    ss3: f64,
    cosines: Vec<f64>,
    sines: Vec<f64>,
    sample_count: usize,
    close1: f64,
    close2: f64,
    high1: f64,
    high2: f64,
    filtered: Vec<f64>,
    correlations: Vec<f64>,
    power: Vec<f64>,
    maximum: f64,
    dominant: f64,
    has_dominant: bool,
    last_output: PeriodogramOutput,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl AutoCorrelationPeriodogram {
    /// Creates a new `AutoCorrelationPeriodogram` from the supplied parameters.
    pub fn new(params: &AutoCorrelationPeriodogramParams) -> Result<Self, String> {
        const INVALID: &str = "invalid autocorrelation periodogram parameters";

        let resolved_min_i32 = resolve_min_period(params.source, params.min_period);
        if resolved_min_i32 < 3 {
            return Err(format!("{INVALID}: min_period must be an integer >= 3"));
        }
        if params.max_period < 5 || params.max_period <= resolved_min_i32 {
            return Err(format!(
                "{INVALID}: max_period must be an integer >= 5 and > min_period"
            ));
        }
        if params.averaging_length < 0 || params.averaging_length == 1 {
            return Err(format!(
                "{INVALID}: averaging_length must be 0 or an integer >= 2"
            ));
        }

        let min_period = resolved_min_i32 as usize;
        let max_period = params.max_period as usize;
        let averaging_length = params.averaging_length as usize;
        let bins_len = max_period - min_period + 1;
        let lag_span = max_period - DFT_LAG_START + 1;

        let min_f = min_period as f64;
        let max_f = max_period as f64;

        let (pole, hp0, hp1, hp2) = match params.source {
            AutoCorrelationPeriodogramSource::Article => {
                let theta = TWO_PI / max_f;
                let article_pole = (1.0 - theta.sin()) / theta.cos();
                (article_pole, 0.5 * (1.0 + article_pole), article_pole, 0.0)
            }
            AutoCorrelationPeriodogramSource::Book => {
                let theta = INVERSE_SQRT_TWO * TWO_PI / max_f;
                let cosine = theta.cos();
                let alpha = (cosine + theta.sin() - 1.0) / cosine;
                let book_pole = 1.0 - alpha;
                let half = 1.0 - 0.5 * alpha;
                (book_pole, half * half, 2.0 * book_pole, book_pole * book_pole)
            }
        };

        let decay = (-SQRT_TWO * PI / min_f).exp();
        let ss2 = 2.0 * decay * (SQRT_TWO * PI / min_f).cos();
        let ss3 = -decay * decay;
        let ss1 = 1.0 - ss2 - ss3;

        if !(0.0 < pole && pole < 1.0 && 0.0 < decay && decay < 1.0 && ss1 > 0.0)
            || !hp0.is_finite()
            || !hp1.is_finite()
            || !hp2.is_finite()
            || !ss1.is_finite()
            || !ss2.is_finite()
            || !ss3.is_finite()
        {
            return Err(format!(
                "{INVALID}: periods must yield finite stable coefficients"
            ));
        }

        let effective_window = if averaging_length == 0 {
            max_period
        } else {
            averaging_length
        };
        let history_length = 3usize.max(
            max_period
                .checked_add(effective_window)
                .ok_or_else(|| format!("{INVALID}: history length overflowed"))?,
        );
        let corr_len = max_period
            .checked_add(1)
            .ok_or_else(|| format!("{INVALID}: correlation length overflowed"))?;
        let table_size = bins_len
            .checked_mul(lag_span)
            .ok_or_else(|| format!("{INVALID}: basis table size overflowed"))?;

        let mut cosines = vec![0.0; table_size];
        let mut sines = vec![0.0; table_size];
        for bin_index in 0..bins_len {
            let period_f = (min_period + bin_index) as f64;
            let row_offset = bin_index * lag_span;
            for offset in 0..lag_span {
                let lag_f = (DFT_LAG_START + offset) as f64;
                let angle = TWO_PI * lag_f / period_f;
                cosines[row_offset + offset] = angle.cos();
                sines[row_offset + offset] = angle.sin();
            }
        }

        let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let mnemonic_text = mnemonic(params);
        let description_text = format!("Autocorrelation periodogram {mnemonic_text}");
        let mnemonic_dc = dominant_cycle_mnemonic(params);
        let description_dc = format!("Autocorrelation periodogram dominant cycle {mnemonic_dc}");

        let metadata = build_metadata(
            Identifier::AutoCorrelationPeriodogram,
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

        let initial_output = PeriodogramOutput {
            heatmap: Heatmap::empty(0, min_f, max_f, 1.0),
            dominant_cycle: f64::NAN,
        };

        Ok(Self {
            resolved: ResolvedParams {
                source: params.source,
                min_period,
                max_period,
                averaging_length,
                enhance_resolution: params.enhance_resolution,
            },
            metadata,
            history_length,
            bins_len,
            lag_span,
            hp0,
            hp1,
            hp2,
            ss1,
            ss2,
            ss3,
            cosines,
            sines,
            sample_count: 0,
            close1: 0.0,
            close2: 0.0,
            high1: 0.0,
            high2: 0.0,
            filtered: vec![0.0; history_length],
            correlations: vec![0.0; corr_len],
            power: vec![0.0; bins_len],
            maximum: 0.0,
            dominant: 0.0,
            has_dominant: false,
            last_output: initial_output,
            bar_func: bar_component_value(bar_comp),
            quote_func: quote_component_value(quote_comp),
            trade_func: trade_component_value(trade_comp),
        })
    }

    /// Creates a new `AutoCorrelationPeriodogram` with default article parameters.
    pub fn new_default() -> Result<Self, String> {
        Self::new(&AutoCorrelationPeriodogramParams::default())
    }

    /// Returns the validated resolved parameters.
    pub fn parameters(&self) -> ResolvedParams {
        self.resolved
    }

    /// Returns the resolved shortest evaluated cycle period.
    pub fn min_period(&self) -> usize {
        self.resolved.min_period
    }

    /// Returns the longest evaluated cycle period.
    pub fn max_period(&self) -> usize {
        self.resolved.max_period
    }

    /// Returns the number of warmup observations (`history_length - 1`).
    pub fn warmup_bars(&self) -> usize {
        self.history_length - 1
    }

    /// Returns the number of accepted finite observations seen since construction or reset.
    pub fn samples_seen(&self) -> usize {
        self.sample_count
    }

    /// Returns an owned snapshot of committed internal state.
    pub fn state(&self) -> PeriodogramState {
        PeriodogramState {
            high_pass: self.high1,
            filtered: self.filtered[0],
            correlations: self.correlations.clone(),
            smoothed_power: self.power.clone(),
            normalization_maximum: self.maximum,
            internal_dominant_cycle: self.dominant,
        }
    }

    /// Returns an owned copy of the most recent accepted output.
    pub fn output(&self) -> PeriodogramOutput {
        self.last_output.clone()
    }

    /// Resets all roofing, correlation, spectral, AGC, and centroid state to initial values.
    pub fn reset(&mut self) {
        self.sample_count = 0;
        self.close1 = 0.0;
        self.close2 = 0.0;
        self.high1 = 0.0;
        self.high2 = 0.0;
        self.filtered.fill(0.0);
        self.correlations.fill(0.0);
        self.power.fill(0.0);
        self.maximum = 0.0;
        self.dominant = 0.0;
        self.has_dominant = false;
        self.last_output = PeriodogramOutput {
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
    pub fn try_update(&mut self, sample: f64, time: i64) -> Result<PeriodogramOutput, String> {
        if sample.is_infinite() {
            return Err("sample must be finite or NaN (skipped)".to_string());
        }
        let min_f = self.resolved.min_period as f64;
        let max_f = self.resolved.max_period as f64;
        if sample.is_nan() {
            return Ok(PeriodogramOutput {
                heatmap: Heatmap::empty(time, min_f, max_f, 1.0),
                dominant_cycle: f64::NAN,
            });
        }

        let is_book = self.resolved.source == AutoCorrelationPeriodogramSource::Book;
        let high = if !is_book {
            self.hp0 * (sample - self.close1) + self.hp1 * self.high1
        } else {
            self.hp0 * (sample - 2.0 * self.close1 + self.close2)
                + self.hp1 * self.high1
                - self.hp2 * self.high2
        };
        let filtered_value = self.ss1 * (high + self.high1) / 2.0
            + self.ss2 * self.filtered[0]
            + self.ss3 * self.filtered[1];
        if !high.is_finite() || !filtered_value.is_finite() {
            return Err("roofing filter arithmetic overflowed".to_string());
        }

        let mut next_filtered = Vec::with_capacity(self.history_length);
        next_filtered.push(filtered_value);
        next_filtered.extend_from_slice(&self.filtered[..self.history_length - 1]);

        let mut next_correlations = self.correlations.clone();
        for lag in 0..=self.resolved.max_period {
            let window_size = if self.resolved.averaging_length != 0 {
                self.resolved.averaging_length
            } else {
                lag
            };
            let mut sx = 0.0;
            let mut sy = 0.0;
            let mut sxx = 0.0;
            let mut syy = 0.0;
            let mut sxy = 0.0;
            for offset in 0..window_size {
                let x_sample = next_filtered[offset];
                let y_sample = next_filtered[lag + offset];
                sx += x_sample;
                sy += y_sample;
                sxx += x_sample * x_sample;
                sxy += x_sample * y_sample;
                syy += y_sample * y_sample;
            }
            let window_f = window_size as f64;
            let denominator = (window_f * sxx - sx * sx) * (window_f * syy - sy * sy);
            if !sx.is_finite()
                || !sy.is_finite()
                || !sxx.is_finite()
                || !syy.is_finite()
                || !sxy.is_finite()
                || !denominator.is_finite()
            {
                return Err("correlation arithmetic overflowed".to_string());
            }
            if denominator > 0.0 {
                next_correlations[lag] = (window_f * sxy - sx * sy) / denominator.sqrt();
            }
            if !next_correlations[lag].is_finite() {
                return Err("correlation arithmetic overflowed".to_string());
            }
        }

        let mut next_power = vec![0.0; self.bins_len];
        for bin_index in 0..self.bins_len {
            let mut cosine_part = 0.0;
            let mut sine_part = 0.0;
            let row_offset = bin_index * self.lag_span;
            for offset in 0..self.lag_span {
                let corr_value = next_correlations[DFT_LAG_START + offset];
                cosine_part += corr_value * self.cosines[row_offset + offset];
                sine_part += corr_value * self.sines[row_offset + offset];
            }
            let squared_sum = cosine_part * cosine_part + sine_part * sine_part;
            let smoothed =
                POWER_SMOOTHING * squared_sum * squared_sum + POWER_RETENTION * self.power[bin_index];
            if !smoothed.is_finite() {
                return Err("spectral power arithmetic overflowed".to_string());
            }
            next_power[bin_index] = smoothed;
        }

        let mut maximum = if is_book {
            BOOK_AGC_DECAY * self.maximum
        } else {
            0.0
        };
        for &power_value in &next_power {
            if power_value > maximum {
                maximum = power_value;
            }
        }

        let normalized: Vec<f64> = next_power
            .iter()
            .map(|&power_value| {
                if maximum > 0.0 {
                    power_value / maximum
                } else {
                    0.0
                }
            })
            .collect();

        let weights: Vec<f64> = if !is_book && self.resolved.enhance_resolution {
            normalized
                .iter()
                .map(|&norm_value| norm_value * norm_value * norm_value)
                .collect()
        } else {
            normalized.clone()
        };

        let threshold = if is_book {
            BOOK_THRESHOLD
        } else {
            ARTICLE_THRESHOLD
        };
        let mut peak = weights[0];
        for &weight in &weights[1..] {
            if weight > peak {
                peak = weight;
            }
        }

        let mut sp = 0.0;
        let mut spx = 0.0;
        for (bin_index, &weight) in weights.iter().enumerate() {
            if weight >= threshold && (is_book || peak >= ARTICLE_THRESHOLD) {
                let period_f = (self.resolved.min_period + bin_index) as f64;
                spx += period_f * weight;
                sp += weight;
            }
        }

        let mut dominant = self.dominant;
        let mut has_dominant = self.has_dominant;
        if sp != 0.0 && (is_book || sp >= ARTICLE_THRESHOLD) {
            dominant = spx / sp;
            has_dominant = true;
        }
        if !dominant.is_finite() {
            return Err("centroid arithmetic overflowed".to_string());
        }

        let display_values: Vec<f64> = if is_book && self.resolved.enhance_resolution {
            normalized
                .iter()
                .map(|&norm_value| norm_value * norm_value)
                .collect()
        } else {
            weights
        };

        self.close2 = self.close1;
        self.close1 = sample;
        self.high2 = self.high1;
        self.high1 = high;
        self.filtered = next_filtered;
        self.correlations = next_correlations;
        self.power = next_power;
        self.maximum = maximum;
        self.dominant = dominant;
        self.has_dominant = has_dominant;
        self.sample_count += 1;

        let current_output = if self.sample_count >= self.history_length {
            let mut value_min = display_values[0];
            let mut value_max = display_values[0];
            for &item in &display_values[1..] {
                if item < value_min {
                    value_min = item;
                }
                if item > value_max {
                    value_max = item;
                }
            }
            PeriodogramOutput {
                heatmap: Heatmap::new(time, min_f, max_f, 1.0, value_min, value_max, display_values),
                dominant_cycle: if has_dominant { dominant } else { f64::NAN },
            }
        } else {
            PeriodogramOutput {
                heatmap: Heatmap::empty(time, min_f, max_f, 1.0),
                dominant_cycle: f64::NAN,
            }
        };

        self.last_output = current_output.clone();
        Ok(current_output)
    }

    /// Updates the indicator with a scalar sample and timestamp, panicking on invalid input or overflow.
    pub fn update(&mut self, sample: f64, time: i64) -> PeriodogramOutput {
        self.try_update(sample, time).unwrap_or_else(|err| {
            panic!("AutoCorrelationPeriodogram::update failed: {err}")
        })
    }

    fn wrap_output(time: i64, output: PeriodogramOutput) -> Output {
        vec![
            Box::new(output.heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, output.dominant_cycle)),
        ]
    }
}

impl Indicator for AutoCorrelationPeriodogram {
    fn is_primed(&self) -> bool {
        self.sample_count >= self.history_length
    }

    fn metadata(&self) -> Metadata {
        self.metadata.clone()
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let output = self.update(sample.value, sample.time);
        Self::wrap_output(sample.time, output)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let extracted = (self.bar_func)(sample);
        let output = self.update(extracted, sample.time);
        Self::wrap_output(sample.time, output)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let extracted = (self.quote_func)(sample);
        let output = self.update(extracted, sample.time);
        Self::wrap_output(sample.time, output)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let extracted = (self.trade_func)(sample);
        let output = self.update(extracted, sample.time);
        Self::wrap_output(sample.time, output)
    }
}

#[cfg(test)]
mod tests {
    use super::super::testdata;
    use super::*;
    use crate::indicators::core::adaptivity::Adaptivity;
    use crate::indicators::core::descriptors::DESCRIPTORS;
    use crate::indicators::core::outputs::shape::Shape;

    const FIXTURE_TOLERANCE: f64 = 1e-11;
    const ORACLE_TOLERANCE: f64 = 2e-10;
    const EXACT_TOLERANCE: f64 = 1e-13;

    fn assert_close_tol(actual: f64, expected: f64, tolerance: f64, label: &str) {
        if expected.is_nan() {
            assert!(
                actual.is_nan(),
                "{label}: expected NaN, got {actual}"
            );
            return;
        }
        assert!(
            !actual.is_nan(),
            "{label}: expected {expected}, got NaN"
        );
        let diff = (actual - expected).abs();
        let scale = 1.0_f64.max(actual.abs().max(expected.abs()));
        assert!(
            diff <= tolerance * scale,
            "{label}: expected {expected}, got {actual} (diff={diff})"
        );
    }

    fn degree_oracle(
        prices: &[f64],
        params: &AutoCorrelationPeriodogramParams,
    ) -> Vec<(Vec<f64>, f64, bool)> {
        let is_book = params.source == AutoCorrelationPeriodogramSource::Book;
        let min_period = resolve_min_period(params.source, params.min_period) as usize;
        let max_period = params.max_period as usize;
        let averaging_length = params.averaging_length as usize;
        let bins_len = max_period - min_period + 1;

        let mut hp_series: Vec<f64> = Vec::with_capacity(prices.len());
        let mut filtered_series: Vec<f64> = Vec::with_capacity(prices.len());
        let mut correlations = vec![0.0; max_period + 1];
        let mut powers = vec![0.0; bins_len];
        let mut maximum = 0.0;
        let mut dominant = 0.0;
        let mut has_dominant = false;
        let mut records = Vec::with_capacity(prices.len());

        let ss_angle = (SQRT_2 * 180.0 / (min_period as f64)).to_radians();
        let attenuation = (-SQRT_2 * PI / (min_period as f64)).exp();
        let c2 = 2.0 * attenuation * ss_angle.cos();
        let c3 = -attenuation * attenuation;
        let c1 = 1.0 - c2 - c3;

        let article_angle = (360.0 / (max_period as f64)).to_radians();
        let article_pole = (1.0 - article_angle.sin()) / article_angle.cos();

        let book_angle = ((1.0 / SQRT_2) * 360.0 / (max_period as f64)).to_radians();
        let alpha = (book_angle.cos() + book_angle.sin() - 1.0) / book_angle.cos();

        let old_at = |series: &[f64], signed_index: isize| -> f64 {
            if signed_index >= 0 {
                series[signed_index as usize]
            } else {
                0.0
            }
        };

        for (step_index, &price) in prices.iter().enumerate() {
            let t = step_index as isize;
            let h_value = if is_book {
                (1.0 - alpha / 2.0)
                    * (1.0 - alpha / 2.0)
                    * (price - 2.0 * old_at(prices, t - 1) + old_at(prices, t - 2))
                    + 2.0 * (1.0 - alpha) * old_at(&hp_series, t - 1)
                    - (1.0 - alpha) * (1.0 - alpha) * old_at(&hp_series, t - 2)
            } else {
                0.5 * (1.0 + article_pole) * (price - old_at(prices, t - 1))
                    + article_pole * old_at(&hp_series, t - 1)
            };
            hp_series.push(h_value);

            let f_value = c1 * (h_value + old_at(&hp_series, t - 1)) / 2.0
                + c2 * old_at(&filtered_series, t - 1)
                + c3 * old_at(&filtered_series, t - 2);
            filtered_series.push(f_value);

            for lag in 0..=max_period {
                let window_m = if averaging_length != 0 {
                    averaging_length
                } else {
                    lag
                };
                let mut sx = 0.0;
                let mut sy = 0.0;
                let mut sxx = 0.0;
                let mut syy = 0.0;
                let mut sxy = 0.0;
                for j in 0..window_m {
                    let x_item = old_at(&filtered_series, t - (j as isize));
                    let y_item = old_at(&filtered_series, t - (lag as isize) - (j as isize));
                    sx += x_item;
                    sy += y_item;
                    sxx += x_item * x_item;
                    sxy += x_item * y_item;
                    syy += y_item * y_item;
                }
                let m_f = window_m as f64;
                let denom = (m_f * sxx - sx * sx) * (m_f * syy - sy * sy);
                if denom > 0.0 {
                    correlations[lag] = (m_f * sxy - sx * sy) / denom.sqrt();
                }
            }

            for bin_index in 0..bins_len {
                let period_f = (min_period + bin_index) as f64;
                let mut cp = 0.0;
                let mut sp = 0.0;
                for lag in 3..=max_period {
                    let angle = (360.0 * (lag as f64) / period_f).to_radians();
                    cp += correlations[lag] * angle.cos();
                    sp += correlations[lag] * angle.sin();
                }
                let q = cp * cp + sp * sp;
                powers[bin_index] = 0.2 * q * q + 0.8 * powers[bin_index];
            }

            maximum = if is_book { 0.995 * maximum } else { 0.0 };
            for &power_item in &powers {
                if power_item > maximum {
                    maximum = power_item;
                }
            }

            let normalized: Vec<f64> = powers
                .iter()
                .map(|&p| if maximum > 0.0 { p / maximum } else { 0.0 })
                .collect();
            let weights: Vec<f64> = if !is_book && params.enhance_resolution {
                normalized.iter().map(|&p| p * p * p).collect()
            } else {
                normalized.clone()
            };

            let mut max_weight = weights[0];
            for &w in &weights[1..] {
                if w > max_weight {
                    max_weight = w;
                }
            }

            let mut spx = 0.0;
            let mut sp = 0.0;
            for (bin_index, &w) in weights.iter().enumerate() {
                if (is_book && w >= 0.5) || (!is_book && max_weight >= 0.25 && w >= 0.25) {
                    spx += ((min_period + bin_index) as f64) * w;
                    sp += w;
                }
            }

            let mut candidate = if sp != 0.0 { spx / sp } else { 0.0 };
            if (is_book && sp == 0.0) || (!is_book && sp < 0.25) {
                candidate = dominant;
            } else if sp != 0.0 {
                has_dominant = true;
            }
            dominant = candidate;

            let display = if is_book && params.enhance_resolution {
                normalized.iter().map(|&p| p * p).collect()
            } else {
                weights
            };
            records.push((display, dominant, has_dominant));
        }

        records
    }

    #[test]
    fn test_all_64_reference_cases_and_degree_oracle() {
        let prices = testdata::testdata::test_input();
        let cases: Vec<(
            &str,
            AutoCorrelationPeriodogramParams,
            Vec<Vec<f64>>,
            Vec<f64>,
            Vec<f64>,
            Vec<f64>,
        )> = vec![
            (
                "v0_l8_h24_a0_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 0,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h24_a0_e0_heatmap(),
                testdata::expected_v0_l8_h24_a0_e0_min(),
                testdata::expected_v0_l8_h24_a0_e0_max(),
                testdata::expected_v0_l8_h24_a0_e0_dc(),
            ),
            (
                "v0_l8_h24_a0_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 0,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h24_a0_e1_heatmap(),
                testdata::expected_v0_l8_h24_a0_e1_min(),
                testdata::expected_v0_l8_h24_a0_e1_max(),
                testdata::expected_v0_l8_h24_a0_e1_dc(),
            ),
            (
                "v0_l8_h24_a3_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 3,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h24_a3_e0_heatmap(),
                testdata::expected_v0_l8_h24_a3_e0_min(),
                testdata::expected_v0_l8_h24_a3_e0_max(),
                testdata::expected_v0_l8_h24_a3_e0_dc(),
            ),
            (
                "v0_l8_h24_a3_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 3,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h24_a3_e1_heatmap(),
                testdata::expected_v0_l8_h24_a3_e1_min(),
                testdata::expected_v0_l8_h24_a3_e1_max(),
                testdata::expected_v0_l8_h24_a3_e1_dc(),
            ),
            (
                "v0_l8_h24_a5_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 5,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h24_a5_e0_heatmap(),
                testdata::expected_v0_l8_h24_a5_e0_min(),
                testdata::expected_v0_l8_h24_a5_e0_max(),
                testdata::expected_v0_l8_h24_a5_e0_dc(),
            ),
            (
                "v0_l8_h24_a5_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 5,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h24_a5_e1_heatmap(),
                testdata::expected_v0_l8_h24_a5_e1_min(),
                testdata::expected_v0_l8_h24_a5_e1_max(),
                testdata::expected_v0_l8_h24_a5_e1_dc(),
            ),
            (
                "v0_l8_h24_a10_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 10,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h24_a10_e0_heatmap(),
                testdata::expected_v0_l8_h24_a10_e0_min(),
                testdata::expected_v0_l8_h24_a10_e0_max(),
                testdata::expected_v0_l8_h24_a10_e0_dc(),
            ),
            (
                "v0_l8_h24_a10_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 10,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h24_a10_e1_heatmap(),
                testdata::expected_v0_l8_h24_a10_e1_min(),
                testdata::expected_v0_l8_h24_a10_e1_max(),
                testdata::expected_v0_l8_h24_a10_e1_dc(),
            ),
            (
                "v0_l8_h48_a0_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 0,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h48_a0_e0_heatmap(),
                testdata::expected_v0_l8_h48_a0_e0_min(),
                testdata::expected_v0_l8_h48_a0_e0_max(),
                testdata::expected_v0_l8_h48_a0_e0_dc(),
            ),
            (
                "v0_l8_h48_a0_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 0,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h48_a0_e1_heatmap(),
                testdata::expected_v0_l8_h48_a0_e1_min(),
                testdata::expected_v0_l8_h48_a0_e1_max(),
                testdata::expected_v0_l8_h48_a0_e1_dc(),
            ),
            (
                "v0_l8_h48_a3_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 3,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h48_a3_e0_heatmap(),
                testdata::expected_v0_l8_h48_a3_e0_min(),
                testdata::expected_v0_l8_h48_a3_e0_max(),
                testdata::expected_v0_l8_h48_a3_e0_dc(),
            ),
            (
                "v0_l8_h48_a3_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 3,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h48_a3_e1_heatmap(),
                testdata::expected_v0_l8_h48_a3_e1_min(),
                testdata::expected_v0_l8_h48_a3_e1_max(),
                testdata::expected_v0_l8_h48_a3_e1_dc(),
            ),
            (
                "v0_l8_h48_a5_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 5,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h48_a5_e0_heatmap(),
                testdata::expected_v0_l8_h48_a5_e0_min(),
                testdata::expected_v0_l8_h48_a5_e0_max(),
                testdata::expected_v0_l8_h48_a5_e0_dc(),
            ),
            (
                "v0_l8_h48_a5_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 5,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h48_a5_e1_heatmap(),
                testdata::expected_v0_l8_h48_a5_e1_min(),
                testdata::expected_v0_l8_h48_a5_e1_max(),
                testdata::expected_v0_l8_h48_a5_e1_dc(),
            ),
            (
                "v0_l8_h48_a10_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 10,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h48_a10_e0_heatmap(),
                testdata::expected_v0_l8_h48_a10_e0_min(),
                testdata::expected_v0_l8_h48_a10_e0_max(),
                testdata::expected_v0_l8_h48_a10_e0_dc(),
            ),
            (
                "v0_l8_h48_a10_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 10,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l8_h48_a10_e1_heatmap(),
                testdata::expected_v0_l8_h48_a10_e1_min(),
                testdata::expected_v0_l8_h48_a10_e1_max(),
                testdata::expected_v0_l8_h48_a10_e1_dc(),
            ),
            (
                "v0_l10_h48_a0_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 0,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l10_h48_a0_e0_heatmap(),
                testdata::expected_v0_l10_h48_a0_e0_min(),
                testdata::expected_v0_l10_h48_a0_e0_max(),
                testdata::expected_v0_l10_h48_a0_e0_dc(),
            ),
            (
                "v0_l10_h48_a0_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 0,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l10_h48_a0_e1_heatmap(),
                testdata::expected_v0_l10_h48_a0_e1_min(),
                testdata::expected_v0_l10_h48_a0_e1_max(),
                testdata::expected_v0_l10_h48_a0_e1_dc(),
            ),
            (
                "v0_l10_h48_a3_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 3,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l10_h48_a3_e0_heatmap(),
                testdata::expected_v0_l10_h48_a3_e0_min(),
                testdata::expected_v0_l10_h48_a3_e0_max(),
                testdata::expected_v0_l10_h48_a3_e0_dc(),
            ),
            (
                "v0_l10_h48_a3_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 3,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l10_h48_a3_e1_heatmap(),
                testdata::expected_v0_l10_h48_a3_e1_min(),
                testdata::expected_v0_l10_h48_a3_e1_max(),
                testdata::expected_v0_l10_h48_a3_e1_dc(),
            ),
            (
                "v0_l10_h48_a5_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 5,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l10_h48_a5_e0_heatmap(),
                testdata::expected_v0_l10_h48_a5_e0_min(),
                testdata::expected_v0_l10_h48_a5_e0_max(),
                testdata::expected_v0_l10_h48_a5_e0_dc(),
            ),
            (
                "v0_l10_h48_a5_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 5,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l10_h48_a5_e1_heatmap(),
                testdata::expected_v0_l10_h48_a5_e1_min(),
                testdata::expected_v0_l10_h48_a5_e1_max(),
                testdata::expected_v0_l10_h48_a5_e1_dc(),
            ),
            (
                "v0_l10_h48_a10_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 10,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l10_h48_a10_e0_heatmap(),
                testdata::expected_v0_l10_h48_a10_e0_min(),
                testdata::expected_v0_l10_h48_a10_e0_max(),
                testdata::expected_v0_l10_h48_a10_e0_dc(),
            ),
            (
                "v0_l10_h48_a10_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 10,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l10_h48_a10_e1_heatmap(),
                testdata::expected_v0_l10_h48_a10_e1_min(),
                testdata::expected_v0_l10_h48_a10_e1_max(),
                testdata::expected_v0_l10_h48_a10_e1_dc(),
            ),
            (
                "v0_l16_h64_a0_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 0,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l16_h64_a0_e0_heatmap(),
                testdata::expected_v0_l16_h64_a0_e0_min(),
                testdata::expected_v0_l16_h64_a0_e0_max(),
                testdata::expected_v0_l16_h64_a0_e0_dc(),
            ),
            (
                "v0_l16_h64_a0_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 0,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l16_h64_a0_e1_heatmap(),
                testdata::expected_v0_l16_h64_a0_e1_min(),
                testdata::expected_v0_l16_h64_a0_e1_max(),
                testdata::expected_v0_l16_h64_a0_e1_dc(),
            ),
            (
                "v0_l16_h64_a3_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 3,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l16_h64_a3_e0_heatmap(),
                testdata::expected_v0_l16_h64_a3_e0_min(),
                testdata::expected_v0_l16_h64_a3_e0_max(),
                testdata::expected_v0_l16_h64_a3_e0_dc(),
            ),
            (
                "v0_l16_h64_a3_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 3,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l16_h64_a3_e1_heatmap(),
                testdata::expected_v0_l16_h64_a3_e1_min(),
                testdata::expected_v0_l16_h64_a3_e1_max(),
                testdata::expected_v0_l16_h64_a3_e1_dc(),
            ),
            (
                "v0_l16_h64_a5_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 5,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l16_h64_a5_e0_heatmap(),
                testdata::expected_v0_l16_h64_a5_e0_min(),
                testdata::expected_v0_l16_h64_a5_e0_max(),
                testdata::expected_v0_l16_h64_a5_e0_dc(),
            ),
            (
                "v0_l16_h64_a5_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 5,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l16_h64_a5_e1_heatmap(),
                testdata::expected_v0_l16_h64_a5_e1_min(),
                testdata::expected_v0_l16_h64_a5_e1_max(),
                testdata::expected_v0_l16_h64_a5_e1_dc(),
            ),
            (
                "v0_l16_h64_a10_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 10,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v0_l16_h64_a10_e0_heatmap(),
                testdata::expected_v0_l16_h64_a10_e0_min(),
                testdata::expected_v0_l16_h64_a10_e0_max(),
                testdata::expected_v0_l16_h64_a10_e0_dc(),
            ),
            (
                "v0_l16_h64_a10_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Article,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 10,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v0_l16_h64_a10_e1_heatmap(),
                testdata::expected_v0_l16_h64_a10_e1_min(),
                testdata::expected_v0_l16_h64_a10_e1_max(),
                testdata::expected_v0_l16_h64_a10_e1_dc(),
            ),
            (
                "v1_l8_h24_a0_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 0,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h24_a0_e0_heatmap(),
                testdata::expected_v1_l8_h24_a0_e0_min(),
                testdata::expected_v1_l8_h24_a0_e0_max(),
                testdata::expected_v1_l8_h24_a0_e0_dc(),
            ),
            (
                "v1_l8_h24_a0_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 0,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h24_a0_e1_heatmap(),
                testdata::expected_v1_l8_h24_a0_e1_min(),
                testdata::expected_v1_l8_h24_a0_e1_max(),
                testdata::expected_v1_l8_h24_a0_e1_dc(),
            ),
            (
                "v1_l8_h24_a3_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 3,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h24_a3_e0_heatmap(),
                testdata::expected_v1_l8_h24_a3_e0_min(),
                testdata::expected_v1_l8_h24_a3_e0_max(),
                testdata::expected_v1_l8_h24_a3_e0_dc(),
            ),
            (
                "v1_l8_h24_a3_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 3,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h24_a3_e1_heatmap(),
                testdata::expected_v1_l8_h24_a3_e1_min(),
                testdata::expected_v1_l8_h24_a3_e1_max(),
                testdata::expected_v1_l8_h24_a3_e1_dc(),
            ),
            (
                "v1_l8_h24_a5_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 5,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h24_a5_e0_heatmap(),
                testdata::expected_v1_l8_h24_a5_e0_min(),
                testdata::expected_v1_l8_h24_a5_e0_max(),
                testdata::expected_v1_l8_h24_a5_e0_dc(),
            ),
            (
                "v1_l8_h24_a5_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 5,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h24_a5_e1_heatmap(),
                testdata::expected_v1_l8_h24_a5_e1_min(),
                testdata::expected_v1_l8_h24_a5_e1_max(),
                testdata::expected_v1_l8_h24_a5_e1_dc(),
            ),
            (
                "v1_l8_h24_a10_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 10,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h24_a10_e0_heatmap(),
                testdata::expected_v1_l8_h24_a10_e0_min(),
                testdata::expected_v1_l8_h24_a10_e0_max(),
                testdata::expected_v1_l8_h24_a10_e0_dc(),
            ),
            (
                "v1_l8_h24_a10_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 24,
                    averaging_length: 10,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h24_a10_e1_heatmap(),
                testdata::expected_v1_l8_h24_a10_e1_min(),
                testdata::expected_v1_l8_h24_a10_e1_max(),
                testdata::expected_v1_l8_h24_a10_e1_dc(),
            ),
            (
                "v1_l8_h48_a0_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 0,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h48_a0_e0_heatmap(),
                testdata::expected_v1_l8_h48_a0_e0_min(),
                testdata::expected_v1_l8_h48_a0_e0_max(),
                testdata::expected_v1_l8_h48_a0_e0_dc(),
            ),
            (
                "v1_l8_h48_a0_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 0,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h48_a0_e1_heatmap(),
                testdata::expected_v1_l8_h48_a0_e1_min(),
                testdata::expected_v1_l8_h48_a0_e1_max(),
                testdata::expected_v1_l8_h48_a0_e1_dc(),
            ),
            (
                "v1_l8_h48_a3_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 3,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h48_a3_e0_heatmap(),
                testdata::expected_v1_l8_h48_a3_e0_min(),
                testdata::expected_v1_l8_h48_a3_e0_max(),
                testdata::expected_v1_l8_h48_a3_e0_dc(),
            ),
            (
                "v1_l8_h48_a3_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 3,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h48_a3_e1_heatmap(),
                testdata::expected_v1_l8_h48_a3_e1_min(),
                testdata::expected_v1_l8_h48_a3_e1_max(),
                testdata::expected_v1_l8_h48_a3_e1_dc(),
            ),
            (
                "v1_l8_h48_a5_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 5,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h48_a5_e0_heatmap(),
                testdata::expected_v1_l8_h48_a5_e0_min(),
                testdata::expected_v1_l8_h48_a5_e0_max(),
                testdata::expected_v1_l8_h48_a5_e0_dc(),
            ),
            (
                "v1_l8_h48_a5_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 5,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h48_a5_e1_heatmap(),
                testdata::expected_v1_l8_h48_a5_e1_min(),
                testdata::expected_v1_l8_h48_a5_e1_max(),
                testdata::expected_v1_l8_h48_a5_e1_dc(),
            ),
            (
                "v1_l8_h48_a10_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 10,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h48_a10_e0_heatmap(),
                testdata::expected_v1_l8_h48_a10_e0_min(),
                testdata::expected_v1_l8_h48_a10_e0_max(),
                testdata::expected_v1_l8_h48_a10_e0_dc(),
            ),
            (
                "v1_l8_h48_a10_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(8),
                    max_period: 48,
                    averaging_length: 10,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l8_h48_a10_e1_heatmap(),
                testdata::expected_v1_l8_h48_a10_e1_min(),
                testdata::expected_v1_l8_h48_a10_e1_max(),
                testdata::expected_v1_l8_h48_a10_e1_dc(),
            ),
            (
                "v1_l10_h48_a0_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 0,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l10_h48_a0_e0_heatmap(),
                testdata::expected_v1_l10_h48_a0_e0_min(),
                testdata::expected_v1_l10_h48_a0_e0_max(),
                testdata::expected_v1_l10_h48_a0_e0_dc(),
            ),
            (
                "v1_l10_h48_a0_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 0,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l10_h48_a0_e1_heatmap(),
                testdata::expected_v1_l10_h48_a0_e1_min(),
                testdata::expected_v1_l10_h48_a0_e1_max(),
                testdata::expected_v1_l10_h48_a0_e1_dc(),
            ),
            (
                "v1_l10_h48_a3_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 3,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l10_h48_a3_e0_heatmap(),
                testdata::expected_v1_l10_h48_a3_e0_min(),
                testdata::expected_v1_l10_h48_a3_e0_max(),
                testdata::expected_v1_l10_h48_a3_e0_dc(),
            ),
            (
                "v1_l10_h48_a3_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 3,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l10_h48_a3_e1_heatmap(),
                testdata::expected_v1_l10_h48_a3_e1_min(),
                testdata::expected_v1_l10_h48_a3_e1_max(),
                testdata::expected_v1_l10_h48_a3_e1_dc(),
            ),
            (
                "v1_l10_h48_a5_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 5,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l10_h48_a5_e0_heatmap(),
                testdata::expected_v1_l10_h48_a5_e0_min(),
                testdata::expected_v1_l10_h48_a5_e0_max(),
                testdata::expected_v1_l10_h48_a5_e0_dc(),
            ),
            (
                "v1_l10_h48_a5_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 5,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l10_h48_a5_e1_heatmap(),
                testdata::expected_v1_l10_h48_a5_e1_min(),
                testdata::expected_v1_l10_h48_a5_e1_max(),
                testdata::expected_v1_l10_h48_a5_e1_dc(),
            ),
            (
                "v1_l10_h48_a10_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 10,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l10_h48_a10_e0_heatmap(),
                testdata::expected_v1_l10_h48_a10_e0_min(),
                testdata::expected_v1_l10_h48_a10_e0_max(),
                testdata::expected_v1_l10_h48_a10_e0_dc(),
            ),
            (
                "v1_l10_h48_a10_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(10),
                    max_period: 48,
                    averaging_length: 10,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l10_h48_a10_e1_heatmap(),
                testdata::expected_v1_l10_h48_a10_e1_min(),
                testdata::expected_v1_l10_h48_a10_e1_max(),
                testdata::expected_v1_l10_h48_a10_e1_dc(),
            ),
            (
                "v1_l16_h64_a0_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 0,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l16_h64_a0_e0_heatmap(),
                testdata::expected_v1_l16_h64_a0_e0_min(),
                testdata::expected_v1_l16_h64_a0_e0_max(),
                testdata::expected_v1_l16_h64_a0_e0_dc(),
            ),
            (
                "v1_l16_h64_a0_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 0,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l16_h64_a0_e1_heatmap(),
                testdata::expected_v1_l16_h64_a0_e1_min(),
                testdata::expected_v1_l16_h64_a0_e1_max(),
                testdata::expected_v1_l16_h64_a0_e1_dc(),
            ),
            (
                "v1_l16_h64_a3_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 3,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l16_h64_a3_e0_heatmap(),
                testdata::expected_v1_l16_h64_a3_e0_min(),
                testdata::expected_v1_l16_h64_a3_e0_max(),
                testdata::expected_v1_l16_h64_a3_e0_dc(),
            ),
            (
                "v1_l16_h64_a3_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 3,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l16_h64_a3_e1_heatmap(),
                testdata::expected_v1_l16_h64_a3_e1_min(),
                testdata::expected_v1_l16_h64_a3_e1_max(),
                testdata::expected_v1_l16_h64_a3_e1_dc(),
            ),
            (
                "v1_l16_h64_a5_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 5,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l16_h64_a5_e0_heatmap(),
                testdata::expected_v1_l16_h64_a5_e0_min(),
                testdata::expected_v1_l16_h64_a5_e0_max(),
                testdata::expected_v1_l16_h64_a5_e0_dc(),
            ),
            (
                "v1_l16_h64_a5_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 5,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l16_h64_a5_e1_heatmap(),
                testdata::expected_v1_l16_h64_a5_e1_min(),
                testdata::expected_v1_l16_h64_a5_e1_max(),
                testdata::expected_v1_l16_h64_a5_e1_dc(),
            ),
            (
                "v1_l16_h64_a10_e0",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 10,
                    enhance_resolution: false,
                    ..Default::default()
                },
                testdata::expected_v1_l16_h64_a10_e0_heatmap(),
                testdata::expected_v1_l16_h64_a10_e0_min(),
                testdata::expected_v1_l16_h64_a10_e0_max(),
                testdata::expected_v1_l16_h64_a10_e0_dc(),
            ),
            (
                "v1_l16_h64_a10_e1",
                AutoCorrelationPeriodogramParams {
                    source: AutoCorrelationPeriodogramSource::Book,
                    min_period: Some(16),
                    max_period: 64,
                    averaging_length: 10,
                    enhance_resolution: true,
                    ..Default::default()
                },
                testdata::expected_v1_l16_h64_a10_e1_heatmap(),
                testdata::expected_v1_l16_h64_a10_e1_min(),
                testdata::expected_v1_l16_h64_a10_e1_max(),
                testdata::expected_v1_l16_h64_a10_e1_dc(),
            ),
        ];

        assert_eq!(cases.len(), 64);

        for (name, params, expected_hm, expected_min, expected_max, expected_dc) in &cases {
            let mut indicator = AutoCorrelationPeriodogram::new(params).unwrap();
            let warmup = indicator.warmup_bars();
            let bins_len = indicator.max_period() - indicator.min_period() + 1;
            let oracle = degree_oracle(&prices, params);

            for (bar_index, &price) in prices.iter().enumerate() {
                let result = indicator.update(price, bar_index as i64);
                assert_eq!(
                    indicator.is_primed(),
                    bar_index >= warmup,
                    "{name}[{bar_index}] primed"
                );
                assert_close_tol(
                    result.heatmap.parameter_first,
                    indicator.min_period() as f64,
                    EXACT_TOLERANCE,
                    name,
                );
                assert_close_tol(
                    result.heatmap.parameter_last,
                    indicator.max_period() as f64,
                    EXACT_TOLERANCE,
                    name,
                );
                assert_close_tol(result.heatmap.parameter_resolution, 1.0, EXACT_TOLERANCE, name);

                if bar_index < warmup {
                    assert!(result.heatmap.values.is_empty(), "{name}[{bar_index}] empty");
                    assert!(result.heatmap.value_min.is_nan(), "{name}[{bar_index}] min NaN");
                    assert!(result.heatmap.value_max.is_nan(), "{name}[{bar_index}] max NaN");
                    assert!(result.dominant_cycle.is_nan(), "{name}[{bar_index}] dc NaN");
                } else {
                    assert_eq!(result.heatmap.values.len(), bins_len, "{name}[{bar_index}] bins");
                    assert_close_tol(
                        result.heatmap.value_min,
                        expected_min[bar_index],
                        FIXTURE_TOLERANCE,
                        name,
                    );
                    assert_close_tol(
                        result.heatmap.value_max,
                        expected_max[bar_index],
                        FIXTURE_TOLERANCE,
                        name,
                    );
                    assert_close_tol(
                        result.dominant_cycle,
                        expected_dc[bar_index],
                        FIXTURE_TOLERANCE,
                        name,
                    );
                    for bin_index in 0..bins_len {
                        assert_close_tol(
                            result.heatmap.values[bin_index],
                            expected_hm[bar_index][bin_index],
                            FIXTURE_TOLERANCE,
                            name,
                        );
                        assert_close_tol(
                            result.heatmap.values[bin_index],
                            oracle[bar_index].0[bin_index],
                            ORACLE_TOLERANCE,
                            name,
                        );
                    }
                    let expected_oracle_dc = if oracle[bar_index].2 {
                        oracle[bar_index].1
                    } else {
                        f64::NAN
                    };
                    assert_close_tol(
                        result.dominant_cycle,
                        expected_oracle_dc,
                        ORACLE_TOLERANCE,
                        name,
                    );
                }
            }
        }
    }

    #[test]
    fn test_enhancement_order_book_hold_zero_sine_nan_overflow_and_reset() {
        let prices = testdata::testdata::test_input();

        let mut book_plain = AutoCorrelationPeriodogram::new(&AutoCorrelationPeriodogramParams {
            source: AutoCorrelationPeriodogramSource::Book,
            enhance_resolution: false,
            ..Default::default()
        })
        .unwrap();
        let mut book_enh = AutoCorrelationPeriodogram::new(&AutoCorrelationPeriodogramParams {
            source: AutoCorrelationPeriodogramSource::Book,
            enhance_resolution: true,
            ..Default::default()
        })
        .unwrap();
        let mut art_plain = AutoCorrelationPeriodogram::new(&AutoCorrelationPeriodogramParams {
            source: AutoCorrelationPeriodogramSource::Article,
            enhance_resolution: false,
            ..Default::default()
        })
        .unwrap();
        let mut art_enh = AutoCorrelationPeriodogram::new(&AutoCorrelationPeriodogramParams {
            source: AutoCorrelationPeriodogramSource::Article,
            enhance_resolution: true,
            ..Default::default()
        })
        .unwrap();

        let mut previous_book_dc = f64::NAN;
        let mut hold_count = 0usize;
        let mut last_art_plain_dc = 0.0;
        let mut last_art_enh_dc = 0.0;

        for (bar_index, &price) in prices.iter().enumerate() {
            let bp = book_plain.update(price, bar_index as i64);
            let be = book_enh.update(price, bar_index as i64);
            let ap = art_plain.update(price, bar_index as i64);
            let ae = art_enh.update(price, bar_index as i64);

            if bar_index >= 50 {
                assert_close_tol(
                    bp.dominant_cycle,
                    be.dominant_cycle,
                    EXACT_TOLERANCE,
                    "book dc enhance invariance",
                );
                for bin_index in 0..bp.heatmap.values.len() {
                    let squared = bp.heatmap.values[bin_index] * bp.heatmap.values[bin_index];
                    assert_close_tol(
                        be.heatmap.values[bin_index],
                        squared,
                        EXACT_TOLERANCE,
                        "book squared display",
                    );
                }
                assert_close_tol(ap.heatmap.value_max, 1.0, EXACT_TOLERANCE, "article max 1");
                for bin_index in 0..ap.heatmap.values.len() {
                    let cubed = ap.heatmap.values[bin_index]
                        * ap.heatmap.values[bin_index]
                        * ap.heatmap.values[bin_index];
                    assert_close_tol(
                        ae.heatmap.values[bin_index],
                        cubed,
                        EXACT_TOLERANCE,
                        "article cubed display",
                    );
                }
                if bar_index >= 51 && bp.heatmap.value_max < 0.5 {
                    hold_count += 1;
                    assert_close_tol(
                        bp.dominant_cycle,
                        previous_book_dc,
                        EXACT_TOLERANCE,
                        "book hold",
                    );
                }
            }
            previous_book_dc = bp.dominant_cycle;
            last_art_plain_dc = ap.dominant_cycle;
            last_art_enh_dc = ae.dominant_cycle;
        }
        assert!(hold_count > 0);
        assert!((last_art_plain_dc - last_art_enh_dc).abs() > 1e-9);

        let mut zero_ind = AutoCorrelationPeriodogram::new_default().unwrap();
        let mut zero_out = zero_ind.output();
        for step_index in 0..100 {
            zero_out = zero_ind.update(0.0, step_index);
        }
        assert!(zero_ind.is_primed());
        assert!(zero_out.dominant_cycle.is_nan());
        for &cell in &zero_out.heatmap.values {
            assert_eq!(cell, 0.0);
        }
        zero_ind.correlations[12] = 0.75;
        let _ = zero_ind.update(0.0, 100);
        assert_eq!(zero_ind.state().correlations[12], 0.75);

        let mut sine_ind = AutoCorrelationPeriodogram::new_default().unwrap();
        let mut sine_out = sine_ind.output();
        for step_index in 0..1000 {
            let sample = 100.0 + (TWO_PI * (step_index as f64) / 20.0).sin();
            sine_out = sine_ind.update(sample, step_index);
        }
        let mut best_bin = 0usize;
        for (bin_index, &cell) in sine_out.heatmap.values.iter().enumerate() {
            if cell > sine_out.heatmap.values[best_bin] {
                best_bin = bin_index;
            }
        }
        assert_close_tol(
            sine_out.heatmap.parameter_first + (best_bin as f64),
            20.0,
            EXACT_TOLERANCE,
            "sine peak 20",
        );
        assert!(sine_out.dominant_cycle > 19.0 && sine_out.dominant_cycle < 21.0);

        let saved_state = art_plain.state();
        let saved_count = art_plain.samples_seen();
        let saved_dc = art_plain.output().dominant_cycle;

        let nan_out = art_plain.try_update(f64::NAN, 999).unwrap();
        assert!(nan_out.dominant_cycle.is_nan());
        assert!(nan_out.heatmap.values.is_empty());
        assert_eq!(art_plain.samples_seen(), saved_count);
        assert_eq!(art_plain.state(), saved_state);

        assert!(art_plain.try_update(f64::INFINITY, 1000).is_err());
        assert!(art_plain.try_update(f64::NEG_INFINITY, 1001).is_err());
        assert!(art_plain.try_update(1e308, 1002).is_err());
        assert_eq!(art_plain.samples_seen(), saved_count);
        assert_eq!(art_plain.state(), saved_state);

        let mut mutated_state = art_plain.state();
        mutated_state.correlations[3] = 999.0;
        mutated_state.smoothed_power[0] = 999.0;
        assert_eq!(art_plain.state(), saved_state);

        let mut mutated_output = art_plain.output();
        mutated_output.heatmap.values[0] = 999.0;
        assert_ne!(art_plain.output().heatmap.values[0], 999.0);

        art_plain.reset();
        assert!(!art_plain.is_primed());
        assert_eq!(art_plain.samples_seen(), 0);
        let mut replay_dc = 0.0;
        for (bar_index, &price) in prices.iter().enumerate() {
            replay_dc = art_plain.update(price, bar_index as i64).dominant_cycle;
        }
        assert_eq!(replay_dc, saved_dc);
    }

    #[test]
    fn test_metadata_mnemonics_entities_and_validation() {
        let descriptor = DESCRIPTORS.iter().find(|item| item.identifier == Identifier::AutoCorrelationPeriodogram).unwrap();
        assert_eq!(descriptor.adaptivity, Adaptivity::Adaptive);
        assert_eq!(descriptor.outputs.len(), 2);

        let default_ind = AutoCorrelationPeriodogram::new_default().unwrap();
        let meta = default_ind.metadata();
        assert_eq!(meta.identifier, Identifier::AutoCorrelationPeriodogram);
        assert_eq!(meta.mnemonic, "acp(8, 48, 3)");
        assert_eq!(meta.description, "Autocorrelation periodogram acp(8, 48, 3)");
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(
            meta.outputs[0].kind,
            AutoCorrelationPeriodogramOutput::Value as i32
        );
        assert_eq!(meta.outputs[0].shape, Shape::Heatmap);
        assert_eq!(meta.outputs[0].mnemonic, "acp(8, 48, 3)");
        assert_eq!(
            meta.outputs[1].kind,
            AutoCorrelationPeriodogramOutput::DominantCycle as i32
        );
        assert_eq!(meta.outputs[1].shape, Shape::Scalar);
        assert_eq!(meta.outputs[1].mnemonic, "acp-dc(8, 48, 3)");
        assert_eq!(
            meta.outputs[1].description,
            "Autocorrelation periodogram dominant cycle acp-dc(8, 48, 3)"
        );

        let book_enh = AutoCorrelationPeriodogram::new(&AutoCorrelationPeriodogramParams {
            source: AutoCorrelationPeriodogramSource::Book,
            enhance_resolution: true,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(book_enh.min_period(), 10);
        assert_eq!(book_enh.metadata().mnemonic, "acp(10, 48, 3, book, enh)");
        assert_eq!(
            book_enh.metadata().outputs[1].mnemonic,
            "acp-dc(10, 48, 3, book)"
        );

        let mnemonic_cases = [
            (
                AutoCorrelationPeriodogramParams::default(),
                "acp(8, 48, 3)",
                "acp-dc(8, 48, 3)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    enhance_resolution: true,
                    ..Default::default()
                },
                "acp(8, 48, 3, enh)",
                "acp-dc(8, 48, 3, enh)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    bar_component: Some(BarComponent::Open),
                    ..Default::default()
                },
                "acp(8, 48, 3, o)",
                "acp-dc(8, 48, 3, o)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    quote_component: Some(QuoteComponent::Bid),
                    ..Default::default()
                },
                "acp(8, 48, 3, b)",
                "acp-dc(8, 48, 3, b)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "acp(8, 48, 3, v)",
                "acp-dc(8, 48, 3, v)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    bar_component: Some(BarComponent::High),
                    quote_component: Some(QuoteComponent::Ask),
                    ..Default::default()
                },
                "acp(8, 48, 3, h, a)",
                "acp-dc(8, 48, 3, h, a)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    bar_component: Some(BarComponent::Low),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "acp(8, 48, 3, l, v)",
                "acp-dc(8, 48, 3, l, v)",
            ),
            (
                AutoCorrelationPeriodogramParams {
                    quote_component: Some(QuoteComponent::SpreadBp),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "acp(8, 48, 3, spread bp, v)",
                "acp-dc(8, 48, 3, spread bp, v)",
            ),
        ];
        for (case_params, expected_m, expected_dc_m) in &mnemonic_cases {
            assert_eq!(&mnemonic(case_params), expected_m);
            assert_eq!(&dominant_cycle_mnemonic(case_params), expected_dc_m);
        }

        assert_eq!(AutoCorrelationPeriodogramOutput::Value.as_str(), "value");
        assert_eq!(
            AutoCorrelationPeriodogramOutput::DominantCycle.as_str(),
            "dominantCycle"
        );
        assert!(AutoCorrelationPeriodogramOutput::Value.is_known());
        assert!(AutoCorrelationPeriodogramOutput::DominantCycle.is_known());
        assert_eq!(
            AutoCorrelationPeriodogramSource::from_str("article"),
            Some(AutoCorrelationPeriodogramSource::Article)
        );
        assert_eq!(
            AutoCorrelationPeriodogramSource::from_str("book"),
            Some(AutoCorrelationPeriodogramSource::Book)
        );
        assert_eq!(AutoCorrelationPeriodogramSource::from_str("invalid"), None);

        let prices = testdata::testdata::test_input();
        let highs = testdata::testdata::test_high();
        let lows = testdata::testdata::test_low();
        let opens = testdata::testdata::test_open();
        let mut bar_ind = AutoCorrelationPeriodogram::new_default().unwrap();
        let mut scalar_ind = AutoCorrelationPeriodogram::new_default().unwrap();
        let mut quote_ind = AutoCorrelationPeriodogram::new_default().unwrap();
        let mut trade_ind = AutoCorrelationPeriodogram::new_default().unwrap();

        for step_index in 0..60 {
            let bar_item = Bar {
                time: step_index as i64,
                open: opens[step_index],
                high: highs[step_index],
                low: lows[step_index],
                close: prices[step_index],
                volume: 1000.0,
            };
            let scalar_item = Scalar::new(step_index as i64, prices[step_index]);
            let quote_item = Quote {
                time: step_index as i64,
                bid_price: prices[step_index] - 0.5,
                ask_price: prices[step_index] + 0.5,
                bid_size: 10.0,
                ask_size: 10.0,
            };
            let trade_item = Trade {
                time: step_index as i64,
                price: prices[step_index],
                volume: 50.0,
            };

            let out_bar = bar_ind.update_bar(&bar_item);
            let out_scalar = scalar_ind.update_scalar(&scalar_item);
            let out_quote = quote_ind.update_quote(&quote_item);
            let out_trade = trade_ind.update_trade(&trade_item);

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
            for bin_index in 0..hm_bar.values.len() {
                assert_close_tol(
                    hm_bar.values[bin_index],
                    hm_scalar.values[bin_index],
                    EXACT_TOLERANCE,
                    "bar vs scalar",
                );
                assert_close_tol(
                    hm_bar.values[bin_index],
                    hm_quote.values[bin_index],
                    EXACT_TOLERANCE,
                    "bar vs quote",
                );
                assert_close_tol(
                    hm_bar.values[bin_index],
                    hm_trade.values[bin_index],
                    EXACT_TOLERANCE,
                    "bar vs trade",
                );
            }
            assert_close_tol(dc_bar.value, dc_scalar.value, EXACT_TOLERANCE, "dc scalar");
            assert_close_tol(dc_bar.value, dc_quote.value, EXACT_TOLERANCE, "dc quote");
            assert_close_tol(dc_bar.value, dc_trade.value, EXACT_TOLERANCE, "dc trade");
        }

        for invalid_params in [
            AutoCorrelationPeriodogramParams {
                min_period: Some(0),
                ..Default::default()
            },
            AutoCorrelationPeriodogramParams {
                min_period: Some(2),
                ..Default::default()
            },
            AutoCorrelationPeriodogramParams {
                min_period: Some(3),
                max_period: 4,
                ..Default::default()
            },
            AutoCorrelationPeriodogramParams {
                min_period: Some(8),
                max_period: 8,
                ..Default::default()
            },
            AutoCorrelationPeriodogramParams {
                source: AutoCorrelationPeriodogramSource::Book,
                max_period: 10,
                ..Default::default()
            },
            AutoCorrelationPeriodogramParams {
                averaging_length: 1,
                ..Default::default()
            },
            AutoCorrelationPeriodogramParams {
                averaging_length: -1,
                ..Default::default()
            },
        ] {
            assert!(AutoCorrelationPeriodogram::new(&invalid_params).is_err());
        }
    }
}
