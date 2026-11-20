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

// ---------------------------------------------------------------------------
// Output enum
// ---------------------------------------------------------------------------

/// Output indices for Ehlers' 2013 Roofed Autocorrelation Heatmap indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum RoofedAutoCorrelationHeatmapOutput {
    Value = 1,
}

impl RoofedAutoCorrelationHeatmapOutput {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Value => "value",
        }
    }

    pub fn is_known(&self) -> bool {
        matches!(self, Self::Value)
    }
}

impl std::fmt::Display for RoofedAutoCorrelationHeatmapOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

// ---------------------------------------------------------------------------
// Params & mnemonics
// ---------------------------------------------------------------------------

/// Parameters to create a `RoofedAutoCorrelationHeatmap` instance.
///
/// Defaults correspond to Ehlers' 2013 *Cycle Analytics for Traders* Listing 8-2 settings.
/// Explicit values (including `min_lag = 0`, `max_lag = 0`, and `averaging_length = 0`) are
/// validated directly without silent substitution.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct RoofedAutoCorrelationHeatmapParams {
    /// First displayed correlation lag on the heatmap axis (`>= 0`). Default is `3`.
    pub min_lag: i32,
    /// Last displayed correlation lag on the heatmap axis (`>= min_lag`). Default is `48`.
    pub max_lag: i32,
    /// Number of aligned sample pairs per Pearson correlation comparison (`>= 0`; `0` selects `M = lag`). Default is `0`.
    pub averaging_length: i32,
    /// Two-pole roofing high-pass cutoff period in bars (`>= 5.0`). Default is `48.0`.
    pub highpass_period: f64,
    /// Two-pole SuperSmoother cutoff period in bars (`>= 2.0`). Default is `10.0`.
    pub smoothing_period: f64,
    /// Bar component to extract (`None` defaults to `Close` and is omitted from the mnemonic).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract (`None` defaults to `Mid` and is omitted from the mnemonic).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract (`None` defaults to `Price` and is omitted from the mnemonic).
    pub trade_component: Option<TradeComponent>,
}

pub type Params = RoofedAutoCorrelationHeatmapParams;

impl Default for RoofedAutoCorrelationHeatmapParams {
    fn default() -> Self {
        Self {
            min_lag: 3,
            max_lag: 48,
            averaging_length: 0,
            highpass_period: 48.0,
            smoothing_period: 10.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Returns default Ehlers 2013 Book Listing 8-2 parameters.
pub fn default_params() -> RoofedAutoCorrelationHeatmapParams {
    RoofedAutoCorrelationHeatmapParams::default()
}

/// Formats a float parameter cleanly for mnemonics without `-0`.
pub fn format_number(value: f64) -> String {
    let normalized = if value == 0.0 { 0.0 } else { value };
    format!("{normalized}")
}

/// Builds the mnemonic for `RoofedAutoCorrelationHeatmap`.
pub fn mnemonic(params: &RoofedAutoCorrelationHeatmapParams) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let comp_suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    let hp_text = format_number(params.highpass_period);
    let lp_text = format_number(params.smoothing_period);
    format!(
        "rach({}, {}, {}, {hp_text}, {lp_text}{comp_suffix})",
        params.min_lag, params.max_lag, params.averaging_length
    )
}

// ---------------------------------------------------------------------------
// State, roofing filter, and resolved parameter types
// ---------------------------------------------------------------------------

/// Snapshot of committed internal roofing and correlation state.
///
/// `correlations` stores the mapped `[0, 1]` values across all internal lags `0..=max_lag`.
#[derive(Debug, Clone, PartialEq)]
pub struct AutocorrelationState {
    pub high_pass: f64,
    pub filtered: f64,
    pub correlations: Vec<f64>,
}

pub type RoofedAutoCorrelationHeatmapState = AutocorrelationState;

/// Validated numeric parameters resolved at construction time.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct ResolvedParams {
    pub min_lag: usize,
    pub max_lag: usize,
    pub averaging_length: usize,
    pub highpass_period: f64,
    pub smoothing_period: f64,
}

#[derive(Debug, Clone)]
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
            return Err("highpass_period must be a finite real number >= 5.0".to_string());
        }
        if !smoothing_period.is_finite() || smoothing_period < 2.0 {
            return Err("smoothing_period must be a finite real number >= 2.0".to_string());
        }

        let angle = TWO_PI / (SQRT_2 * highpass_period);
        let cosine = angle.cos();
        let alpha = (cosine + angle.sin() - 1.0) / cosine;
        let pole = 1.0 - alpha;
        let attenuation = (-SQRT_2 * PI / smoothing_period).exp();
        let hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
        let hp1 = 2.0 * pole;
        let hp2 = pole * pole;
        let ss2 = 2.0 * attenuation * (SQRT_2 * PI / smoothing_period).cos();
        let ss3 = -attenuation * attenuation;
        let ss1 = 1.0 - ss2 - ss3;

        if !(pole > 0.0 && pole < 1.0 && attenuation > 0.0 && attenuation < 1.0 && ss1 > 0.0)
            || !hp0.is_finite()
            || !hp1.is_finite()
            || !hp2.is_finite()
            || !ss1.is_finite()
            || !ss2.is_finite()
            || !ss3.is_finite()
        {
            return Err("periods must produce finite resolvable stable coefficients".to_string());
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
        let high =
            self.hp0 * (price - 2.0 * self.x1 + self.x2) + self.hp1 * self.h1 - self.hp2 * self.h2;
        let filtered =
            self.ss1 * (high + self.h1) / 2.0 + self.ss2 * self.f1 + self.ss3 * self.f2;
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

/// Ehlers' 2013 Roofed Autocorrelation Heatmap indicator (RACH, Book Listing 8-2).
#[derive(Clone)]
pub struct RoofedAutoCorrelationHeatmap {
    resolved: ResolvedParams,
    history_length: usize,
    roof: RoofFilter,
    metadata: Metadata,
    sample_count: usize,
    history: Vec<f64>,
    correlations: Vec<f64>,
    last_output: Heatmap,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl RoofedAutoCorrelationHeatmap {
    /// Creates a new `RoofedAutoCorrelationHeatmap` from validated parameters.
    pub fn new(params: &RoofedAutoCorrelationHeatmapParams) -> Result<Self, String> {
        const INVALID: &str = "invalid roofed autocorrelation heatmap parameters";

        if params.min_lag < 0 {
            return Err(format!("{INVALID}: min_lag must be an integer >= 0"));
        }
        if params.max_lag < params.min_lag {
            return Err(format!("{INVALID}: max_lag must be an integer >= min_lag"));
        }
        if params.averaging_length < 0 {
            return Err(format!(
                "{INVALID}: averaging_length must be an integer >= 0"
            ));
        }

        let roof = RoofFilter::new(params.highpass_period, params.smoothing_period)
            .map_err(|detail| format!("{INVALID}: {detail}"))?;

        let lower_lag = params.min_lag as usize;
        let upper_lag = params.max_lag as usize;
        let average_window = params.averaging_length as usize;
        let effective_window = if average_window > 0 {
            average_window
        } else {
            upper_lag
        };
        let raw_history = upper_lag
            .checked_add(effective_window)
            .ok_or_else(|| format!("{INVALID}: history length overflow"))?;
        let history_length = 1_usize.max(raw_history);
        let bin_count = upper_lag
            .checked_add(1)
            .ok_or_else(|| format!("{INVALID}: lag count overflow"))?;

        let resolved = ResolvedParams {
            min_lag: lower_lag,
            max_lag: upper_lag,
            averaging_length: average_window,
            highpass_period: params.highpass_period,
            smoothing_period: params.smoothing_period,
        };

        let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let mnemonic_value = mnemonic(params);
        let description_value = format!("Roofed autocorrelation heatmap {mnemonic_value}");
        let metadata = build_metadata(
            Identifier::RoofedAutoCorrelationHeatmap,
            &mnemonic_value,
            &description_value,
            &[OutputText {
                mnemonic: mnemonic_value.clone(),
                description: description_value.clone(),
            }],
        );

        let mut instance = Self {
            resolved,
            history_length,
            roof,
            metadata,
            sample_count: 0,
            history: vec![0.0; history_length],
            correlations: vec![0.0; bin_count],
            last_output: Heatmap::empty(0, lower_lag as f64, upper_lag as f64, 1.0),
            bar_func: bar_component_value(bar_comp),
            quote_func: quote_component_value(quote_comp),
            trade_func: trade_component_value(trade_comp),
        };
        instance.reset();
        Ok(instance)
    }

    /// Alias for `new`.
    pub fn create(params: &RoofedAutoCorrelationHeatmapParams) -> Result<Self, String> {
        Self::new(params)
    }

    /// Creates a `RoofedAutoCorrelationHeatmap` using Ehlers' 2013 Book Listing 8-2 defaults.
    pub fn new_default() -> Self {
        Self::new(&default_params())
            .expect("default RoofedAutoCorrelationHeatmap params must be valid")
    }

    /// Returns the resolved numeric parameters.
    pub fn parameters(&self) -> ResolvedParams {
        self.resolved
    }

    /// Returns the inclusive lower displayed correlation lag.
    pub fn min_lag(&self) -> usize {
        self.resolved.min_lag
    }

    /// Returns the inclusive upper displayed correlation lag.
    pub fn max_lag(&self) -> usize {
        self.resolved.max_lag
    }

    /// Axis alias for `min_lag`.
    pub fn min_period(&self) -> usize {
        self.resolved.min_lag
    }

    /// Axis alias for `max_lag`.
    pub fn max_period(&self) -> usize {
        self.resolved.max_lag
    }

    /// Returns the Pearson averaging length (`0` indicates lag-dependent `M = lag`).
    pub fn averaging_length(&self) -> usize {
        self.resolved.averaging_length
    }

    /// Returns the two-pole high-pass cutoff period.
    pub fn highpass_period(&self) -> f64 {
        self.resolved.highpass_period
    }

    /// Returns the two-pole SuperSmoother cutoff period.
    pub fn smoothing_period(&self) -> f64 {
        self.resolved.smoothing_period
    }

    /// Returns the number of leading empty columns (`history_length - 1`).
    pub fn warmup_bars(&self) -> usize {
        self.history_length - 1
    }

    /// Returns the number of accepted finite observations seen since construction or reset.
    pub fn samples_seen(&self) -> usize {
        self.sample_count
    }

    /// Returns an owned snapshot of committed internal roofing and correlation state.
    pub fn state(&self) -> AutocorrelationState {
        AutocorrelationState {
            high_pass: self.roof.h1,
            filtered: self.roof.f1,
            correlations: self.correlations.clone(),
        }
    }

    /// Returns an owned copy of the most recent accepted output.
    pub fn output(&self) -> Heatmap {
        self.last_output.clone()
    }

    /// Restores explicit source initial conditions while preserving configuration.
    pub fn reset(&mut self) {
        self.roof.reset();
        self.sample_count = 0;
        self.history.fill(0.0);
        self.correlations.fill(0.0);
        self.last_output = Heatmap::empty(
            0,
            self.resolved.min_lag as f64,
            self.resolved.max_lag as f64,
            1.0,
        );
    }

    /// Fallibly updates the indicator with a scalar sample and timestamp.
    ///
    /// - `NaN` samples return an empty heatmap without mutating state or aging `samples_seen`.
    /// - `±Inf` samples or candidate arithmetic overflows return `Err` without mutating state.
    pub fn try_update(&mut self, sample: f64, time: i64) -> Result<Heatmap, String> {
        if sample.is_infinite() {
            return Err("price must be finite or NaN (missing sample)".to_string());
        }
        let min_lag_float = self.resolved.min_lag as f64;
        let max_lag_float = self.resolved.max_lag as f64;
        if sample.is_nan() {
            return Ok(Heatmap::empty(time, min_lag_float, max_lag_float, 1.0));
        }

        let (high, filtered) = self.roof.candidate(sample)?;

        let mut next_history = Vec::with_capacity(self.history_length);
        next_history.push(filtered);
        next_history.extend_from_slice(&self.history[..self.history_length - 1]);

        let mut next_correlations = self.correlations.clone();
        let bin_count = self.resolved.max_lag + 1;

        for lag in 0..bin_count {
            let window = if self.resolved.averaging_length > 0 {
                self.resolved.averaging_length
            } else {
                lag
            };
            let window_float = window as f64;

            let mut sum_x = 0.0;
            let mut sum_y = 0.0;
            let mut sum_xx = 0.0;
            let mut sum_xy = 0.0;
            let mut sum_yy = 0.0;
            for offset in 0..window {
                let sample_x = next_history[offset];
                let sample_y = next_history[lag + offset];
                sum_x += sample_x;
                sum_y += sample_y;
                sum_xx += sample_x * sample_x;
                sum_xy += sample_x * sample_y;
                sum_yy += sample_y * sample_y;
            }

            let denominator =
                (window_float * sum_xx - sum_x * sum_x) * (window_float * sum_yy - sum_y * sum_y);
            if !sum_x.is_finite()
                || !sum_y.is_finite()
                || !sum_xx.is_finite()
                || !sum_xy.is_finite()
                || !sum_yy.is_finite()
                || !denominator.is_finite()
            {
                return Err("Pearson arithmetic overflowed".to_string());
            }
            if denominator > 0.0 {
                let numerator = window_float * sum_xy - sum_x * sum_y;
                if !numerator.is_finite() {
                    return Err("Pearson arithmetic overflowed".to_string());
                }
                next_correlations[lag] = numerator / denominator.sqrt();
            }
            next_correlations[lag] = 0.5 * (next_correlations[lag] + 1.0);
            if !next_correlations[lag].is_finite() {
                return Err("correlation arithmetic overflowed".to_string());
            }
        }

        self.roof.commit(sample, high, filtered);
        self.history = next_history;
        self.correlations = next_correlations;
        self.sample_count += 1;

        let current_output = if self.is_primed() {
            let displayed =
                self.correlations[self.resolved.min_lag..=self.resolved.max_lag].to_vec();
            let mut value_min = displayed[0];
            let mut value_max = displayed[0];
            for &item in &displayed[1..] {
                if item < value_min {
                    value_min = item;
                }
                if item > value_max {
                    value_max = item;
                }
            }
            Heatmap::new(
                time,
                min_lag_float,
                max_lag_float,
                1.0,
                value_min,
                value_max,
                displayed,
            )
        } else {
            Heatmap::empty(time, min_lag_float, max_lag_float, 1.0)
        };

        self.last_output = current_output.clone();
        Ok(current_output)
    }

    /// Updates the indicator with a scalar sample and timestamp, panicking on invalid input/overflow.
    pub fn update(&mut self, sample: f64, time: i64) -> Heatmap {
        self.try_update(sample, time)
            .unwrap_or_else(|err| panic!("RoofedAutoCorrelationHeatmap::update failed: {err}"))
    }

    fn wrap_output(output: Heatmap) -> Output {
        vec![Box::new(output) as Box<dyn Any>]
    }
}

impl Indicator for RoofedAutoCorrelationHeatmap {
    fn metadata(&self) -> Metadata {
        self.metadata.clone()
    }

    fn is_primed(&self) -> bool {
        self.sample_count >= self.history_length
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        Self::wrap_output(self.update(scalar.value, scalar.time))
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        Self::wrap_output(self.update(sample, bar.time))
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.quote_func)(quote);
        Self::wrap_output(self.update(sample, quote.time))
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.trade_func)(trade);
        Self::wrap_output(self.update(sample, trade.time))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::core::adaptivity::Adaptivity;
    use crate::indicators::core::descriptor::descriptor_of;
    use crate::indicators::core::input_requirement::InputRequirement;
    use crate::indicators::core::outputs::shape::Shape;
    use crate::indicators::core::pane::Pane;
    use crate::indicators::core::role::Role;
    use crate::indicators::core::volume_usage::VolumeUsage;
    use crate::indicators::john_ehlers::roofed_auto_correlation_heatmap::testdata;
    use crate::indicators::john_ehlers::roofed_auto_correlation_heatmap::testdata::testdata as inputs;

    const ORACLE_TOLERANCE: f64 = 1e-13;
    const FIXTURE_TOLERANCE: f64 = 2e-11;
    const COLLINEAR_A2_TOLERANCE: f64 = 1e-10;

    fn assert_close_tol(actual: f64, expected: f64, tolerance: f64, label: &str) {
        if expected.is_nan() {
            assert!(actual.is_nan(), "{label}: expected NaN, got {actual}");
        } else {
            assert!(!actual.is_nan(), "{label}: expected {expected}, got NaN");
            let diff = (actual - expected).abs();
            assert!(
                diff <= tolerance,
                "{label}: actual={actual} expected={expected} diff={diff} tol={tolerance}"
            );
        }
    }

    fn batch_oracle(
        closes: &[f64],
        min_lag: usize,
        max_lag: usize,
        averaging_length: usize,
        highpass_period: f64,
        smoothing_period: f64,
    ) -> Vec<(f64, f64, Option<Vec<f64>>)> {
        let angle = TWO_PI / (SQRT_2 * highpass_period);
        let cosine = angle.cos();
        let alpha = (cosine + angle.sin() - 1.0) / cosine;
        let pole = 1.0 - alpha;
        let attenuation = (-SQRT_2 * PI / smoothing_period).exp();
        let hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
        let hp1 = 2.0 * pole;
        let hp2 = pole * pole;
        let ss2 = 2.0 * attenuation * (SQRT_2 * PI / smoothing_period).cos();
        let ss3 = -attenuation * attenuation;
        let ss1 = 1.0 - ss2 - ss3;

        let effective_window = if averaging_length > 0 {
            averaging_length
        } else {
            max_lag
        };
        let history_length = 1_usize.max(max_lag + effective_window);

        let mut high_series = vec![0.0_f64; closes.len()];
        let mut filtered_series = vec![0.0_f64; closes.len()];
        let mut correlations = vec![0.0_f64; max_lag + 1];
        let mut rows = Vec::with_capacity(closes.len());

        for (bar_index, &price) in closes.iter().enumerate() {
            let close_1 = if bar_index >= 1 { closes[bar_index - 1] } else { 0.0 };
            let close_2 = if bar_index >= 2 { closes[bar_index - 2] } else { 0.0 };
            let high_1 = if bar_index >= 1 { high_series[bar_index - 1] } else { 0.0 };
            let high_2 = if bar_index >= 2 { high_series[bar_index - 2] } else { 0.0 };
            let filt_1 = if bar_index >= 1 { filtered_series[bar_index - 1] } else { 0.0 };
            let filt_2 = if bar_index >= 2 { filtered_series[bar_index - 2] } else { 0.0 };

            let high = hp0 * (price - 2.0 * close_1 + close_2) + hp1 * high_1 - hp2 * high_2;
            let filtered = ss1 * (high + high_1) / 2.0 + ss2 * filt_1 + ss3 * filt_2;
            high_series[bar_index] = high;
            filtered_series[bar_index] = filtered;

            for lag in 0..=max_lag {
                let window = if averaging_length > 0 {
                    averaging_length
                } else {
                    lag
                };
                let window_float = window as f64;
                let mut sum_x = 0.0;
                let mut sum_y = 0.0;
                let mut sum_xx = 0.0;
                let mut sum_xy = 0.0;
                let mut sum_yy = 0.0;
                for offset in 0..window {
                    let sample_x = if bar_index >= offset {
                        filtered_series[bar_index - offset]
                    } else {
                        0.0
                    };
                    let sample_y = if bar_index >= lag + offset {
                        filtered_series[bar_index - lag - offset]
                    } else {
                        0.0
                    };
                    sum_x += sample_x;
                    sum_y += sample_y;
                    sum_xx += sample_x * sample_x;
                    sum_xy += sample_x * sample_y;
                    sum_yy += sample_y * sample_y;
                }
                let denominator = (window_float * sum_xx - sum_x * sum_x)
                    * (window_float * sum_yy - sum_y * sum_y);
                if denominator > 0.0 {
                    correlations[lag] =
                        (window_float * sum_xy - sum_x * sum_y) / denominator.sqrt();
                }
                correlations[lag] = 0.5 * (correlations[lag] + 1.0);
            }

            if bar_index + 1 < history_length {
                rows.push((high, filtered, None));
            } else {
                rows.push((high, filtered, Some(correlations[min_lag..=max_lag].to_vec())));
            }
        }

        rows
    }

    fn verify_fixture_case(
        case_name: &str,
        params: RoofedAutoCorrelationHeatmapParams,
        expected_heatmap: &[Vec<f64>],
        expected_min: &[f64],
        expected_max: &[f64],
    ) {
        let closes = inputs::test_input();
        assert_eq!(closes.len(), 252);
        let mut indicator = RoofedAutoCorrelationHeatmap::new(&params).unwrap();
        let min_lag = params.min_lag as usize;
        let max_lag = params.max_lag as usize;
        let average_window = params.averaging_length as usize;
        let effective_window = if average_window > 0 {
            average_window
        } else {
            max_lag
        };
        let warmup = 1_usize.max(max_lag + effective_window) - 1;
        assert_eq!(indicator.warmup_bars(), warmup);

        let oracle = batch_oracle(
            &closes,
            min_lag,
            max_lag,
            average_window,
            params.highpass_period,
            params.smoothing_period,
        );
        let bin_count = max_lag - min_lag + 1;
        let fixture_tolerance = if params.averaging_length == 2 {
            COLLINEAR_A2_TOLERANCE
        } else {
            FIXTURE_TOLERANCE
        };

        for (bar_index, &price) in closes.iter().enumerate() {
            let result = indicator.update(price, bar_index as i64);
            assert_eq!(result.parameter_first, min_lag as f64);
            assert_eq!(result.parameter_last, max_lag as f64);
            assert_eq!(result.parameter_resolution, 1.0);
            assert_eq!(indicator.is_primed(), bar_index >= warmup);

            let snapshot = indicator.state();
            assert_close_tol(
                snapshot.high_pass,
                oracle[bar_index].0,
                ORACLE_TOLERANCE,
                &format!("{case_name}[{bar_index}].high_pass"),
            );
            assert_close_tol(
                snapshot.filtered,
                oracle[bar_index].1,
                ORACLE_TOLERANCE,
                &format!("{case_name}[{bar_index}].filtered"),
            );

            let expected_row = &expected_heatmap[bar_index];
            if bar_index < warmup {
                assert!(result.is_empty());
                assert!(result.value_min.is_nan());
                assert!(result.value_max.is_nan());
                assert!(expected_min[bar_index].is_nan());
                assert!(expected_max[bar_index].is_nan());
                for &cell in expected_row {
                    assert!(cell.is_nan());
                }
                continue;
            }

            assert_eq!(result.values.len(), bin_count);
            assert_close_tol(
                result.value_min,
                expected_min[bar_index],
                fixture_tolerance,
                &format!("{case_name}[{bar_index}].min"),
            );
            assert_close_tol(
                result.value_max,
                expected_max[bar_index],
                fixture_tolerance,
                &format!("{case_name}[{bar_index}].max"),
            );
            let oracle_bins = oracle[bar_index].2.as_ref().unwrap();
            for bin_index in 0..bin_count {
                assert_close_tol(
                    result.values[bin_index],
                    expected_row[bin_index],
                    fixture_tolerance,
                    &format!("{case_name}[{bar_index}][{bin_index}]"),
                );
                assert_close_tol(
                    result.values[bin_index],
                    oracle_bins[bin_index],
                    ORACLE_TOLERANCE,
                    &format!("{case_name}[{bar_index}][{bin_index}].oracle"),
                );
            }
        }
    }

    #[test]
    fn test_all_12_reference_cases() {
        let cases: Vec<(
            &str,
            RoofedAutoCorrelationHeatmapParams,
            Vec<Vec<f64>>,
            Vec<f64>,
            Vec<f64>,
        )> = vec![
            (
                "L03_L148_A0_HP48_LP10",
                default_params(),
                testdata::expected_l03_l148_a0_hp48_lp10_heatmap(),
                testdata::expected_l03_l148_a0_hp48_lp10_min(),
                testdata::expected_l03_l148_a0_hp48_lp10_max(),
            ),
            (
                "L03_L148_A2_HP48_LP10",
                RoofedAutoCorrelationHeatmapParams {
                    averaging_length: 2,
                    ..Default::default()
                },
                testdata::expected_l03_l148_a2_hp48_lp10_heatmap(),
                testdata::expected_l03_l148_a2_hp48_lp10_min(),
                testdata::expected_l03_l148_a2_hp48_lp10_max(),
            ),
            (
                "L03_L148_A3_HP48_LP10",
                RoofedAutoCorrelationHeatmapParams {
                    averaging_length: 3,
                    ..Default::default()
                },
                testdata::expected_l03_l148_a3_hp48_lp10_heatmap(),
                testdata::expected_l03_l148_a3_hp48_lp10_min(),
                testdata::expected_l03_l148_a3_hp48_lp10_max(),
            ),
            (
                "L03_L148_A20_HP48_LP10",
                RoofedAutoCorrelationHeatmapParams {
                    averaging_length: 20,
                    ..Default::default()
                },
                testdata::expected_l03_l148_a20_hp48_lp10_heatmap(),
                testdata::expected_l03_l148_a20_hp48_lp10_min(),
                testdata::expected_l03_l148_a20_hp48_lp10_max(),
            ),
            (
                "L03_L148_A100_HP48_LP10",
                RoofedAutoCorrelationHeatmapParams {
                    averaging_length: 100,
                    ..Default::default()
                },
                testdata::expected_l03_l148_a100_hp48_lp10_heatmap(),
                testdata::expected_l03_l148_a100_hp48_lp10_min(),
                testdata::expected_l03_l148_a100_hp48_lp10_max(),
            ),
            (
                "L03_L148_A1_HP48_LP10",
                RoofedAutoCorrelationHeatmapParams {
                    averaging_length: 1,
                    ..Default::default()
                },
                testdata::expected_l03_l148_a1_hp48_lp10_heatmap(),
                testdata::expected_l03_l148_a1_hp48_lp10_min(),
                testdata::expected_l03_l148_a1_hp48_lp10_max(),
            ),
            (
                "L03_L148_A0_HP48_LP10_5",
                RoofedAutoCorrelationHeatmapParams {
                    smoothing_period: 10.5,
                    ..Default::default()
                },
                testdata::expected_l03_l148_a0_hp48_lp10_5_heatmap(),
                testdata::expected_l03_l148_a0_hp48_lp10_5_min(),
                testdata::expected_l03_l148_a0_hp48_lp10_5_max(),
            ),
            (
                "L03_L148_A0_HP48_LP20",
                RoofedAutoCorrelationHeatmapParams {
                    smoothing_period: 20.0,
                    ..Default::default()
                },
                testdata::expected_l03_l148_a0_hp48_lp20_heatmap(),
                testdata::expected_l03_l148_a0_hp48_lp20_min(),
                testdata::expected_l03_l148_a0_hp48_lp20_max(),
            ),
            (
                "L03_L148_A0_HP80_LP10",
                RoofedAutoCorrelationHeatmapParams {
                    highpass_period: 80.0,
                    ..Default::default()
                },
                testdata::expected_l03_l148_a0_hp80_lp10_heatmap(),
                testdata::expected_l03_l148_a0_hp80_lp10_min(),
                testdata::expected_l03_l148_a0_hp80_lp10_max(),
            ),
            (
                "L03_L124_A0_HP24_LP10",
                RoofedAutoCorrelationHeatmapParams {
                    min_lag: 3,
                    max_lag: 24,
                    averaging_length: 0,
                    highpass_period: 24.0,
                    smoothing_period: 10.0,
                    ..Default::default()
                },
                testdata::expected_l03_l124_a0_hp24_lp10_heatmap(),
                testdata::expected_l03_l124_a0_hp24_lp10_min(),
                testdata::expected_l03_l124_a0_hp24_lp10_max(),
            ),
            (
                "L012_L112_A12_HP48_LP10",
                RoofedAutoCorrelationHeatmapParams {
                    min_lag: 12,
                    max_lag: 12,
                    averaging_length: 12,
                    highpass_period: 48.0,
                    smoothing_period: 10.0,
                    ..Default::default()
                },
                testdata::expected_l012_l112_a12_hp48_lp10_heatmap(),
                testdata::expected_l012_l112_a12_hp48_lp10_min(),
                testdata::expected_l012_l112_a12_hp48_lp10_max(),
            ),
            (
                "L00_L10_A3_HP48_LP10",
                RoofedAutoCorrelationHeatmapParams {
                    min_lag: 0,
                    max_lag: 0,
                    averaging_length: 3,
                    highpass_period: 48.0,
                    smoothing_period: 10.0,
                    ..Default::default()
                },
                testdata::expected_l00_l10_a3_hp48_lp10_heatmap(),
                testdata::expected_l00_l10_a3_hp48_lp10_min(),
                testdata::expected_l00_l10_a3_hp48_lp10_max(),
            ),
        ];

        for (name, case_params, expected_hm, expected_min, expected_max) in cases {
            verify_fixture_case(name, case_params, &expected_hm, &expected_min, &expected_max);
        }
    }

    #[test]
    fn test_validation_and_output_enum() {
        assert_eq!(RoofedAutoCorrelationHeatmapOutput::Value.as_str(), "value");
        assert!(RoofedAutoCorrelationHeatmapOutput::Value.is_known());
        assert_eq!(
            format!("{}", RoofedAutoCorrelationHeatmapOutput::Value),
            "value"
        );

        for invalid in [
            RoofedAutoCorrelationHeatmapParams {
                min_lag: -1,
                ..Default::default()
            },
            RoofedAutoCorrelationHeatmapParams {
                min_lag: 10,
                max_lag: 9,
                ..Default::default()
            },
            RoofedAutoCorrelationHeatmapParams {
                averaging_length: -1,
                ..Default::default()
            },
            RoofedAutoCorrelationHeatmapParams {
                highpass_period: 4.9,
                ..Default::default()
            },
            RoofedAutoCorrelationHeatmapParams {
                highpass_period: f64::NAN,
                ..Default::default()
            },
            RoofedAutoCorrelationHeatmapParams {
                highpass_period: f64::INFINITY,
                ..Default::default()
            },
            RoofedAutoCorrelationHeatmapParams {
                smoothing_period: 1.9,
                ..Default::default()
            },
            RoofedAutoCorrelationHeatmapParams {
                smoothing_period: f64::NAN,
                ..Default::default()
            },
            RoofedAutoCorrelationHeatmapParams {
                smoothing_period: f64::NEG_INFINITY,
                ..Default::default()
            },
        ] {
            assert!(RoofedAutoCorrelationHeatmap::new(&invalid).is_err());
        }

        let zero_zero = RoofedAutoCorrelationHeatmap::create(&RoofedAutoCorrelationHeatmapParams {
            min_lag: 0,
            max_lag: 0,
            averaging_length: 0,
            highpass_period: 5.0,
            smoothing_period: 2.0,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(zero_zero.warmup_bars(), 0);
        assert_eq!(zero_zero.min_period(), 0);
        assert_eq!(zero_zero.max_period(), 0);
    }

    #[test]
    fn test_nan_overflow_reset_zero_variance_and_entities() {
        let closes = inputs::test_input();
        let opens = inputs::test_open();
        let highs = inputs::test_high();
        let lows = inputs::test_low();

        let params = RoofedAutoCorrelationHeatmapParams {
            min_lag: 3,
            max_lag: 12,
            averaging_length: 5,
            highpass_period: 24.0,
            smoothing_period: 10.0,
            ..Default::default()
        };
        let mut subject = RoofedAutoCorrelationHeatmap::new(&params).unwrap();
        let mut clean = RoofedAutoCorrelationHeatmap::new(&params).unwrap();

        for (bar_index, &price) in closes[..50].iter().enumerate() {
            let before = subject.state();
            let skipped = subject.try_update(f64::NAN, 100).unwrap();
            assert!(skipped.is_empty());
            assert_eq!(subject.state(), before);

            for bad_sample in [f64::INFINITY, f64::NEG_INFINITY, 1e308] {
                assert!(subject.try_update(bad_sample, 100).is_err());
                assert_eq!(subject.state(), before);
            }

            let actual = subject.update(price, bar_index as i64);
            let expected = clean.update(price, bar_index as i64);
            assert_eq!(subject.is_primed(), bar_index + 1 >= 17);
            assert_eq!(actual.values, expected.values);
        }

        subject.reset();
        clean.reset();
        assert_eq!(subject.samples_seen(), 0);
        assert!(!subject.is_primed());
        for (bar_index, &price) in closes[..50].iter().enumerate() {
            assert_eq!(
                subject.update(price, bar_index as i64).values,
                clean.update(price, bar_index as i64).values
            );
        }

        // Persistent zero-variance rescaling rule: c(t) = 0.5 * (c(t-1) + 1.0)
        let mut flat_ind = RoofedAutoCorrelationHeatmap::new(&RoofedAutoCorrelationHeatmapParams {
            min_lag: 0,
            max_lag: 2,
            averaging_length: 1,
            highpass_period: 48.0,
            smoothing_period: 10.0,
            ..Default::default()
        })
        .unwrap();
        let mut expected_cell = 0.0_f64;
        for step in 0..10 {
            expected_cell = 0.5 * (expected_cell + 1.0);
            let column = flat_ind.update(0.0, step);
            if step >= 2 {
                assert_eq!(column.values, vec![expected_cell, expected_cell, expected_cell]);
                assert_eq!(column.value_min, expected_cell);
                assert_eq!(column.value_max, expected_cell);
            } else {
                assert!(column.is_empty());
            }
        }

        // 20-bar sine wave periodicity: peaks near 1 at lag 20 & 40, troughs near 0 at lag 10 & 30
        let mut sine_ind = RoofedAutoCorrelationHeatmap::new(&RoofedAutoCorrelationHeatmapParams {
            min_lag: 3,
            max_lag: 45,
            averaging_length: 20,
            highpass_period: 48.0,
            smoothing_period: 10.0,
            ..Default::default()
        })
        .unwrap();
        let mut sine_column = Heatmap::empty(0, 3.0, 45.0, 1.0);
        for step in 0..800 {
            sine_column = sine_ind.update((TWO_PI * (step as f64) / 20.0).sin(), step);
        }
        assert_close_tol(sine_column.values[20 - 3], 1.0, 1e-9, "sine[lag=20]");
        assert_close_tol(sine_column.values[40 - 3], 1.0, 1e-9, "sine[lag=40]");
        assert_close_tol(sine_column.values[10 - 3], 0.0, 1e-9, "sine[lag=10]");
        assert_close_tol(sine_column.values[30 - 3], 0.0, 1e-9, "sine[lag=30]");

        let mnemonic_cases = [
            (default_params(), "rach(3, 48, 0, 48, 10)"),
            (
                RoofedAutoCorrelationHeatmapParams {
                    min_lag: 0,
                    max_lag: 0,
                    averaging_length: 3,
                    highpass_period: 24.0,
                    smoothing_period: 10.5,
                    ..Default::default()
                },
                "rach(0, 0, 3, 24, 10.5)",
            ),
            (
                RoofedAutoCorrelationHeatmapParams {
                    bar_component: Some(BarComponent::Median),
                    ..Default::default()
                },
                "rach(3, 48, 0, 48, 10, hl/2)",
            ),
            (
                RoofedAutoCorrelationHeatmapParams {
                    quote_component: Some(QuoteComponent::Bid),
                    ..Default::default()
                },
                "rach(3, 48, 0, 48, 10, b)",
            ),
            (
                RoofedAutoCorrelationHeatmapParams {
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "rach(3, 48, 0, 48, 10, v)",
            ),
            (
                RoofedAutoCorrelationHeatmapParams {
                    bar_component: Some(BarComponent::Typical),
                    quote_component: Some(QuoteComponent::Ask),
                    ..Default::default()
                },
                "rach(3, 48, 0, 48, 10, hlc/3, a)",
            ),
            (
                RoofedAutoCorrelationHeatmapParams {
                    bar_component: Some(BarComponent::Open),
                    quote_component: Some(QuoteComponent::Weighted),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "rach(3, 48, 0, 48, 10, o, (bbs+aas)/(bs+as), v)",
            ),
        ];
        for (case_params, want_mnemonic) in mnemonic_cases {
            assert_eq!(mnemonic(&case_params), want_mnemonic);
            let ind = RoofedAutoCorrelationHeatmap::new(&case_params).unwrap();
            let meta = ind.metadata();
            assert_eq!(meta.identifier, Identifier::RoofedAutoCorrelationHeatmap);
            assert_eq!(meta.mnemonic, want_mnemonic);
            assert_eq!(
                meta.description,
                format!("Roofed autocorrelation heatmap {want_mnemonic}")
            );
            assert_eq!(meta.outputs.len(), 1);
            assert_eq!(
                meta.outputs[0].kind,
                RoofedAutoCorrelationHeatmapOutput::Value as i32
            );
            assert_eq!(meta.outputs[0].shape, Shape::Heatmap);
            assert_eq!(meta.outputs[0].mnemonic, want_mnemonic);
        }

        let desc = descriptor_of(Identifier::RoofedAutoCorrelationHeatmap).unwrap();
        assert_eq!(desc.family, "John Ehlers");
        assert_eq!(desc.adaptivity, Adaptivity::Static);
        assert_eq!(desc.input_requirement, InputRequirement::ScalarInput);
        assert_eq!(desc.volume_usage, VolumeUsage::NoVolume);
        assert_eq!(desc.outputs.len(), 1);
        assert_eq!(desc.outputs[0].shape, Shape::Heatmap);
        assert_eq!(desc.outputs[0].role, Role::Correlation);
        assert_eq!(desc.outputs[0].pane, Pane::Own);

        let mut scalar_ind = RoofedAutoCorrelationHeatmap::new_default();
        let mut bar_ind = RoofedAutoCorrelationHeatmap::new_default();
        let mut quote_ind = RoofedAutoCorrelationHeatmap::new_default();
        let mut trade_ind = RoofedAutoCorrelationHeatmap::new_default();
        let mut open_ind =
            RoofedAutoCorrelationHeatmap::new(&RoofedAutoCorrelationHeatmapParams {
                bar_component: Some(BarComponent::Open),
                ..Default::default()
            })
            .unwrap();
        let mut open_ref = RoofedAutoCorrelationHeatmap::new_default();

        for bar_index in 0..closes.len() {
            let stamp = bar_index as i64;
            let close_price = closes[bar_index];
            let open_price = opens[bar_index];
            let bar = Bar::new(
                stamp,
                open_price,
                highs[bar_index],
                lows[bar_index],
                close_price,
                1000.0,
            );
            let scalar = Scalar::new(stamp, close_price);
            let quote = Quote::new(stamp, close_price, close_price, 10.0, 10.0);
            let trade = Trade::new(stamp, close_price, 10.0);

            let out_scalar = scalar_ind.update_scalar(&scalar)[0]
                .downcast_ref::<Heatmap>()
                .unwrap()
                .clone();
            let out_bar = bar_ind.update_bar(&bar)[0]
                .downcast_ref::<Heatmap>()
                .unwrap()
                .clone();
            let out_quote = quote_ind.update_quote(&quote)[0]
                .downcast_ref::<Heatmap>()
                .unwrap()
                .clone();
            let out_trade = trade_ind.update_trade(&trade)[0]
                .downcast_ref::<Heatmap>()
                .unwrap()
                .clone();
            assert_eq!(out_scalar.values, out_bar.values);
            assert_eq!(out_scalar.values, out_quote.values);
            assert_eq!(out_scalar.values, out_trade.values);

            let out_open = open_ind.update_bar(&bar)[0]
                .downcast_ref::<Heatmap>()
                .unwrap()
                .clone();
            let out_open_ref = open_ref
                .update_scalar(&Scalar::new(stamp, open_price))[0]
                .downcast_ref::<Heatmap>()
                .unwrap()
                .clone();
            assert_eq!(out_open.values, out_open_ref.values);
        }
    }
}
