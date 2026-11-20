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

// ---------------------------------------------------------------------------
// Output enum
// ---------------------------------------------------------------------------

/// Output indices for Ehlers' February 2025 Autocorrelation Heatmap indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AutoCorrelationHeatmapOutput {
    Value = 1,
}

impl AutoCorrelationHeatmapOutput {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Value => "value",
        }
    }

    pub fn is_known(&self) -> bool {
        matches!(self, Self::Value)
    }
}

impl std::fmt::Display for AutoCorrelationHeatmapOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

// ---------------------------------------------------------------------------
// Params & mnemonics
// ---------------------------------------------------------------------------

/// Parameters to create an `AutoCorrelationHeatmap` instance.
///
/// Defaults correspond to Ehlers' February 2025 article settings.
/// Explicit values (including `max_lag = 0`) are validated directly without silent substitution.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct AutoCorrelationHeatmapParams {
    /// Number of aligned sample pairs per Pearson correlation comparison (>= 2). Default is 20.
    pub length: i32,
    /// Ultimate Smoother critical period in bars (>= 1.0). Default is 20.0.
    pub smoothing_period: f64,
    /// Inclusive upper correlation lag in `[0, 99]`; minimum lag is always zero. Default is 99.
    pub max_lag: i32,
    /// Bar component to extract (`None` defaults to `Close` and is omitted from the mnemonic).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract (`None` defaults to `Mid` and is omitted from the mnemonic).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract (`None` defaults to `Price` and is omitted from the mnemonic).
    pub trade_component: Option<TradeComponent>,
}

pub type Params = AutoCorrelationHeatmapParams;

impl Default for AutoCorrelationHeatmapParams {
    fn default() -> Self {
        Self {
            length: 20,
            smoothing_period: 20.0,
            max_lag: 99,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Returns default Ehlers February 2025 article parameters.
pub fn default_params() -> AutoCorrelationHeatmapParams {
    AutoCorrelationHeatmapParams::default()
}

/// Formats a float parameter cleanly for mnemonics without `-0`.
pub fn format_number(value: f64) -> String {
    let normalized = if value == 0.0 { 0.0 } else { value };
    format!("{normalized}")
}

/// Builds the mnemonic for `AutoCorrelationHeatmap`.
pub fn mnemonic(params: &AutoCorrelationHeatmapParams) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let comp_suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    let period_text = format_number(params.smoothing_period);
    format!("ach({}, {period_text}, {}{comp_suffix})", params.length, params.max_lag)
}

// ---------------------------------------------------------------------------
// State & resolved parameter types
// ---------------------------------------------------------------------------

/// Snapshot of committed internal filter and correlation state.
#[derive(Debug, Clone, PartialEq)]
pub struct AutocorrelationState {
    pub accepted_samples: usize,
    pub filtered: f64,
    pub correlations: Vec<f64>,
}

/// Validated numeric parameters resolved at construction time.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct ResolvedParams {
    pub length: usize,
    pub smoothing_period: f64,
    pub max_lag: usize,
}

// ---------------------------------------------------------------------------
// Indicator implementation
// ---------------------------------------------------------------------------

/// Ehlers' February 2025 Autocorrelation Heatmap (`ACH`).
pub struct AutoCorrelationHeatmap {
    resolved: ResolvedParams,
    history_length: usize,
    prime_count: usize,
    coeff_c1: f64,
    coeff_c2: f64,
    coeff_c3: f64,
    metadata: Metadata,
    sample_count: usize,
    price1: f64,
    price2: f64,
    filtered1: f64,
    filtered2: f64,
    history: Vec<f64>,
    correlations: Vec<f64>,
    last_output: Heatmap,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl AutoCorrelationHeatmap {
    /// Creates a new `AutoCorrelationHeatmap` with the supplied parameters.
    pub fn new(params: &AutoCorrelationHeatmapParams) -> Result<Self, String> {
        const INVALID: &str = "invalid autocorrelation heatmap parameters";

        if params.length < 2 {
            return Err(format!("{INVALID}: length must be an integer >= 2"));
        }
        if !params.smoothing_period.is_finite() || params.smoothing_period < 1.0 {
            return Err(format!(
                "{INVALID}: smoothing_period must be a finite real number >= 1.0"
            ));
        }
        if params.max_lag < 0 || params.max_lag > 99 {
            return Err(format!("{INVALID}: max_lag must be an integer in [0, 99]"));
        }

        let angle = SQRT_2 * PI / params.smoothing_period;
        let attenuation = (-angle).exp();
        let coeff_c2 = 2.0 * attenuation * angle.cos();
        let coeff_c3 = -attenuation * attenuation;
        let coeff_c1 = (1.0 + coeff_c2 - coeff_c3) / 4.0;
        if !coeff_c1.is_finite() || !coeff_c2.is_finite() || !coeff_c3.is_finite() {
            return Err(format!(
                "{INVALID}: smoothing_period produces non-finite filter coefficients"
            ));
        }

        let window_length = params.length as usize;
        let upper_lag = params.max_lag as usize;
        let history_length = window_length
            .checked_add(upper_lag)
            .ok_or_else(|| format!("{INVALID}: history length overflow"))?;
        let prime_count = 4_usize.max(history_length);

        let resolved = ResolvedParams {
            length: window_length,
            smoothing_period: params.smoothing_period,
            max_lag: upper_lag,
        };

        let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let mnemonic_value = mnemonic(params);
        let description_value = format!("Autocorrelation heatmap {mnemonic_value}");
        let metadata = build_metadata(
            Identifier::AutoCorrelationHeatmap,
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
            prime_count,
            coeff_c1,
            coeff_c2,
            coeff_c3,
            metadata,
            sample_count: 0,
            price1: 0.0,
            price2: 0.0,
            filtered1: 0.0,
            filtered2: 0.0,
            history: vec![0.0; history_length],
            correlations: vec![0.0; upper_lag + 1],
            last_output: Heatmap::empty(0, 0.0, upper_lag as f64, 1.0),
            bar_func: bar_component_value(bar_comp),
            quote_func: quote_component_value(quote_comp),
            trade_func: trade_component_value(trade_comp),
        };
        instance.reset();
        Ok(instance)
    }

    /// Alias for `new`.
    pub fn create(params: &AutoCorrelationHeatmapParams) -> Result<Self, String> {
        Self::new(params)
    }

    /// Creates an `AutoCorrelationHeatmap` using Ehlers' February 2025 article defaults.
    pub fn new_default() -> Self {
        Self::new(&default_params()).expect("default AutoCorrelationHeatmap params must be valid")
    }

    /// Returns the resolved numeric parameters.
    pub fn parameters(&self) -> ResolvedParams {
        self.resolved
    }

    /// Returns the Pearson correlation sample pair count.
    pub fn length(&self) -> usize {
        self.resolved.length
    }

    /// Returns the Ultimate Smoother critical period.
    pub fn smoothing_period(&self) -> f64 {
        self.resolved.smoothing_period
    }

    /// Returns the inclusive upper correlation lag.
    pub fn max_lag(&self) -> usize {
        self.resolved.max_lag
    }

    /// Returns the number of leading empty columns (`max(4, length + max_lag) - 1`).
    pub fn warmup_bars(&self) -> usize {
        self.prime_count - 1
    }

    /// Returns the number of accepted finite observations seen since construction or reset.
    pub fn samples_seen(&self) -> usize {
        self.sample_count
    }

    /// Returns an owned snapshot of committed internal state.
    pub fn state(&self) -> AutocorrelationState {
        AutocorrelationState {
            accepted_samples: self.sample_count,
            filtered: self.filtered1,
            correlations: self.correlations.clone(),
        }
    }

    /// Returns an owned copy of the most recent accepted output.
    pub fn output(&self) -> Heatmap {
        self.last_output.clone()
    }

    /// Restores explicit source initial conditions while preserving configuration.
    pub fn reset(&mut self) {
        self.sample_count = 0;
        self.price1 = 0.0;
        self.price2 = 0.0;
        self.filtered1 = 0.0;
        self.filtered2 = 0.0;
        self.history.fill(0.0);
        self.correlations.fill(0.0);
        self.last_output = Heatmap::empty(0, 0.0, self.resolved.max_lag as f64, 1.0);
    }

    /// Fallibly updates the indicator with a scalar sample and timestamp.
    ///
    /// - `NaN` samples return an empty heatmap without mutating state or aging `samples_seen`.
    /// - `±Inf` samples or candidate arithmetic overflows return `Err` without mutating state.
    pub fn try_update(&mut self, sample: f64, time: i64) -> Result<Heatmap, String> {
        if sample.is_infinite() {
            return Err("price must be finite or NaN (missing sample)".to_string());
        }
        let max_lag_float = self.resolved.max_lag as f64;
        if sample.is_nan() {
            return Ok(Heatmap::empty(time, 0.0, max_lag_float, 1.0));
        }

        let filtered = if self.sample_count < 3 {
            sample
        } else {
            (1.0 - self.coeff_c1) * sample
                + (2.0 * self.coeff_c1 - self.coeff_c2) * self.price1
                - (self.coeff_c1 + self.coeff_c3) * self.price2
                + self.coeff_c2 * self.filtered1
                + self.coeff_c3 * self.filtered2
        };
        if !filtered.is_finite() {
            return Err("Ultimate Smoother calculation overflowed".to_string());
        }

        let mut next_history = Vec::with_capacity(self.history_length);
        next_history.push(filtered);
        next_history.extend_from_slice(&self.history[..self.history_length - 1]);

        let mut next_correlations = self.correlations.clone();
        let length_float = self.resolved.length as f64;
        let bin_count = self.resolved.max_lag + 1;

        for lag in 0..bin_count {
            let mut sum_x = 0.0;
            let mut sum_y = 0.0;
            let mut sum_xx = 0.0;
            let mut sum_xy = 0.0;
            let mut sum_yy = 0.0;
            for offset in 0..self.resolved.length {
                let sample_x = next_history[offset];
                let sample_y = next_history[lag + offset];
                sum_x += sample_x;
                sum_y += sample_y;
                sum_xx += sample_x * sample_x;
                sum_xy += sample_x * sample_y;
                sum_yy += sample_y * sample_y;
            }
            let var_x = length_float * sum_xx - sum_x * sum_x;
            let var_y = length_float * sum_yy - sum_y * sum_y;
            if !sum_x.is_finite()
                || !sum_y.is_finite()
                || !sum_xx.is_finite()
                || !sum_xy.is_finite()
                || !sum_yy.is_finite()
                || !var_x.is_finite()
                || !var_y.is_finite()
            {
                return Err("Pearson sums overflowed; reduce input magnitude".to_string());
            }
            if var_x > 0.0 && var_y > 0.0 {
                let product = var_x * var_y;
                let numerator = length_float * sum_xy - sum_x * sum_y;
                if product <= 0.0 || !product.is_finite() || !numerator.is_finite() {
                    return Err(
                        "Pearson denominator or numerator overflowed/underflowed".to_string(),
                    );
                }
                let correlation = numerator / product.sqrt();
                if !correlation.is_finite() {
                    return Err("Pearson correlation is not finite".to_string());
                }
                next_correlations[lag] = correlation;
            }
        }

        self.price2 = self.price1;
        self.price1 = sample;
        self.filtered2 = self.filtered1;
        self.filtered1 = filtered;
        self.history = next_history;
        self.correlations = next_correlations.clone();
        self.sample_count += 1;

        let current_output = if self.is_primed() {
            let mut value_min = next_correlations[0];
            let mut value_max = next_correlations[0];
            for &item in &next_correlations[1..] {
                if item < value_min {
                    value_min = item;
                }
                if item > value_max {
                    value_max = item;
                }
            }
            Heatmap::new(
                time,
                0.0,
                max_lag_float,
                1.0,
                value_min,
                value_max,
                next_correlations,
            )
        } else {
            Heatmap::empty(time, 0.0, max_lag_float, 1.0)
        };

        self.last_output = current_output.clone();
        Ok(current_output)
    }

    /// Updates the indicator with a scalar sample and timestamp, panicking on invalid input/overflow.
    pub fn update(&mut self, sample: f64, time: i64) -> Heatmap {
        self.try_update(sample, time)
            .unwrap_or_else(|err| panic!("AutoCorrelationHeatmap::update failed: {err}"))
    }

    fn wrap_output(output: Heatmap) -> Output {
        vec![Box::new(output) as Box<dyn Any>]
    }
}

impl Indicator for AutoCorrelationHeatmap {
    fn is_primed(&self) -> bool {
        self.sample_count >= self.prime_count
    }

    fn metadata(&self) -> Metadata {
        self.metadata.clone()
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        Self::wrap_output(self.update(sample.value, sample.time))
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
    use crate::indicators::john_ehlers::auto_correlation_heatmap::testdata;
    use crate::indicators::john_ehlers::auto_correlation_heatmap::testdata::testdata as inputs;

    const TOLERANCE: f64 = 1e-13;

    fn assert_close(actual: f64, expected: f64, label: &str) {
        assert_close_tol(actual, expected, TOLERANCE, label);
    }

    fn assert_close_tol(actual: f64, expected: f64, tolerance: f64, label: &str) {
        if expected.is_nan() {
            assert!(actual.is_nan(), "{label}: expected NaN, got {actual}");
        } else {
            assert!(!actual.is_nan(), "{label}: expected {expected}, got NaN");
            let diff = (actual - expected).abs();
            assert!(
                diff <= tolerance,
                "{label}: actual={actual} expected={expected} diff={diff}"
            );
        }
    }

    fn batch_oracle(
        closes: &[f64],
        length: usize,
        smoothing_period: f64,
        max_lag: usize,
    ) -> Vec<(f64, Option<Vec<f64>>)> {
        let mut filtered_series = vec![0.0_f64; closes.len()];
        let mut correlations = [0.0_f64; 101];
        let angle = SQRT_2 * PI / smoothing_period;
        let attenuation = (-angle).exp();
        let coeff_c2 = 2.0 * attenuation * angle.cos();
        let coeff_c3 = -attenuation * attenuation;
        let coeff_c1 = (1.0 + coeff_c2 - coeff_c3) / 4.0;
        let prime_count = 4_usize.max(length + max_lag);
        let length_float = length as f64;
        let mut rows = Vec::with_capacity(closes.len());

        for (bar_index, &price) in closes.iter().enumerate() {
            if bar_index < 3 {
                filtered_series[bar_index] = price;
            } else {
                filtered_series[bar_index] = (1.0 - coeff_c1) * price
                    + (2.0 * coeff_c1 - coeff_c2) * closes[bar_index - 1]
                    - (coeff_c1 + coeff_c3) * closes[bar_index - 2]
                    + coeff_c2 * filtered_series[bar_index - 1]
                    + coeff_c3 * filtered_series[bar_index - 2];
            }
            for lag in 0..=max_lag {
                let mut sum_x = 0.0;
                let mut sum_y = 0.0;
                let mut sum_xx = 0.0;
                let mut sum_xy = 0.0;
                let mut sum_yy = 0.0;
                for offset in 0..length {
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
                let var_x = length_float * sum_xx - sum_x * sum_x;
                let var_y = length_float * sum_yy - sum_y * sum_y;
                if var_x > 0.0 && var_y > 0.0 {
                    correlations[lag + 1] =
                        (length_float * sum_xy - sum_x * sum_y) / (var_x * var_y).sqrt();
                }
            }
            if bar_index + 1 < prime_count {
                rows.push((filtered_series[bar_index], None));
            } else {
                rows.push((
                    filtered_series[bar_index],
                    Some(correlations[1..=max_lag + 1].to_vec()),
                ));
            }
        }
        rows
    }

    fn verify_fixture_case(
        case_name: &str,
        params: AutoCorrelationHeatmapParams,
        expected_heatmap: &[Vec<f64>],
        expected_min: &[f64],
        expected_max: &[f64],
    ) {
        let closes = inputs::test_input();
        assert_eq!(closes.len(), 252);
        let mut indicator = AutoCorrelationHeatmap::new(&params).unwrap();
        let length = params.length as usize;
        let max_lag = params.max_lag as usize;
        let warmup = 4_usize.max(length + max_lag) - 1;
        assert_eq!(indicator.warmup_bars(), warmup);

        let oracle = batch_oracle(&closes, length, params.smoothing_period, max_lag);
        let bin_count = max_lag + 1;

        for (bar_index, &price) in closes.iter().enumerate() {
            let result = indicator.update(price, bar_index as i64);
            assert_eq!(result.parameter_first, 0.0);
            assert_eq!(result.parameter_last, max_lag as f64);
            assert_eq!(result.parameter_resolution, 1.0);
            assert_eq!(indicator.is_primed(), bar_index >= warmup);
            assert_close(
                indicator.state().filtered,
                oracle[bar_index].0,
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
            assert_close(
                result.value_min,
                expected_min[bar_index],
                &format!("{case_name}[{bar_index}].min"),
            );
            assert_close(
                result.value_max,
                expected_max[bar_index],
                &format!("{case_name}[{bar_index}].max"),
            );
            let oracle_bins = oracle[bar_index].1.as_ref().unwrap();
            for lag in 0..bin_count {
                assert_close(
                    result.values[lag],
                    expected_row[lag],
                    &format!("{case_name}[{bar_index}][{lag}]"),
                );
                assert_close(
                    result.values[lag],
                    oracle_bins[lag],
                    &format!("{case_name}[{bar_index}][{lag}].oracle"),
                );
            }
        }
    }

    #[test]
    fn test_all_12_reference_cases() {
        let cases: Vec<(
            &str,
            AutoCorrelationHeatmapParams,
            Vec<Vec<f64>>,
            Vec<f64>,
            Vec<f64>,
        )> = vec![
            (
                "L20_P20_H99",
                AutoCorrelationHeatmapParams {
                    length: 20,
                    smoothing_period: 20.0,
                    max_lag: 99,
                    ..Default::default()
                },
                testdata::expected_l20_p20_h99_heatmap(),
                testdata::expected_l20_p20_h99_min(),
                testdata::expected_l20_p20_h99_max(),
            ),
            (
                "L2_P20_H99",
                AutoCorrelationHeatmapParams {
                    length: 2,
                    smoothing_period: 20.0,
                    max_lag: 99,
                    ..Default::default()
                },
                testdata::expected_l2_p20_h99_heatmap(),
                testdata::expected_l2_p20_h99_min(),
                testdata::expected_l2_p20_h99_max(),
            ),
            (
                "L3_P20_H99",
                AutoCorrelationHeatmapParams {
                    length: 3,
                    smoothing_period: 20.0,
                    max_lag: 99,
                    ..Default::default()
                },
                testdata::expected_l3_p20_h99_heatmap(),
                testdata::expected_l3_p20_h99_min(),
                testdata::expected_l3_p20_h99_max(),
            ),
            (
                "L5_P20_H99",
                AutoCorrelationHeatmapParams {
                    length: 5,
                    smoothing_period: 20.0,
                    max_lag: 99,
                    ..Default::default()
                },
                testdata::expected_l5_p20_h99_heatmap(),
                testdata::expected_l5_p20_h99_min(),
                testdata::expected_l5_p20_h99_max(),
            ),
            (
                "L40_P20_H99",
                AutoCorrelationHeatmapParams {
                    length: 40,
                    smoothing_period: 20.0,
                    max_lag: 99,
                    ..Default::default()
                },
                testdata::expected_l40_p20_h99_heatmap(),
                testdata::expected_l40_p20_h99_min(),
                testdata::expected_l40_p20_h99_max(),
            ),
            (
                "L100_P20_H99",
                AutoCorrelationHeatmapParams {
                    length: 100,
                    smoothing_period: 20.0,
                    max_lag: 99,
                    ..Default::default()
                },
                testdata::expected_l100_p20_h99_heatmap(),
                testdata::expected_l100_p20_h99_min(),
                testdata::expected_l100_p20_h99_max(),
            ),
            (
                "L20_P10_H99",
                AutoCorrelationHeatmapParams {
                    length: 20,
                    smoothing_period: 10.0,
                    max_lag: 99,
                    ..Default::default()
                },
                testdata::expected_l20_p10_h99_heatmap(),
                testdata::expected_l20_p10_h99_min(),
                testdata::expected_l20_p10_h99_max(),
            ),
            (
                "L20_P10_5_H99",
                AutoCorrelationHeatmapParams {
                    length: 20,
                    smoothing_period: 10.5,
                    max_lag: 99,
                    ..Default::default()
                },
                testdata::expected_l20_p10_5_h99_heatmap(),
                testdata::expected_l20_p10_5_h99_min(),
                testdata::expected_l20_p10_5_h99_max(),
            ),
            (
                "L20_P40_H99",
                AutoCorrelationHeatmapParams {
                    length: 20,
                    smoothing_period: 40.0,
                    max_lag: 99,
                    ..Default::default()
                },
                testdata::expected_l20_p40_h99_heatmap(),
                testdata::expected_l20_p40_h99_min(),
                testdata::expected_l20_p40_h99_max(),
            ),
            (
                "L20_P20_H32",
                AutoCorrelationHeatmapParams {
                    length: 20,
                    smoothing_period: 20.0,
                    max_lag: 32,
                    ..Default::default()
                },
                testdata::expected_l20_p20_h32_heatmap(),
                testdata::expected_l20_p20_h32_min(),
                testdata::expected_l20_p20_h32_max(),
            ),
            (
                "L2_P20_H32",
                AutoCorrelationHeatmapParams {
                    length: 2,
                    smoothing_period: 20.0,
                    max_lag: 32,
                    ..Default::default()
                },
                testdata::expected_l2_p20_h32_heatmap(),
                testdata::expected_l2_p20_h32_min(),
                testdata::expected_l2_p20_h32_max(),
            ),
            (
                "L20_P20_H0",
                AutoCorrelationHeatmapParams {
                    length: 20,
                    smoothing_period: 20.0,
                    max_lag: 0,
                    ..Default::default()
                },
                testdata::expected_l20_p20_h0_heatmap(),
                testdata::expected_l20_p20_h0_min(),
                testdata::expected_l20_p20_h0_max(),
            ),
        ];

        for (name, case_params, expected_hm, expected_min, expected_max) in cases {
            verify_fixture_case(name, case_params, &expected_hm, &expected_min, &expected_max);
        }
    }

    #[test]
    fn test_edge_cases_sine_retention_mnemonics_metadata_and_entities() {
        for bad_params in [
            AutoCorrelationHeatmapParams {
                length: 1,
                ..Default::default()
            },
            AutoCorrelationHeatmapParams {
                smoothing_period: 0.5,
                ..Default::default()
            },
            AutoCorrelationHeatmapParams {
                smoothing_period: f64::NAN,
                ..Default::default()
            },
            AutoCorrelationHeatmapParams {
                smoothing_period: f64::INFINITY,
                ..Default::default()
            },
            AutoCorrelationHeatmapParams {
                max_lag: -1,
                ..Default::default()
            },
            AutoCorrelationHeatmapParams {
                max_lag: 100,
                ..Default::default()
            },
        ] {
            assert!(AutoCorrelationHeatmap::new(&bad_params).is_err());
        }

        let closes = inputs::test_input();
        let opens = inputs::test_open();
        let highs = inputs::test_high();
        let lows = inputs::test_low();

        let params = AutoCorrelationHeatmapParams {
            length: 5,
            smoothing_period: 10.0,
            max_lag: 12,
            ..Default::default()
        };
        let mut subject = AutoCorrelationHeatmap::new(&params).unwrap();
        let mut clean = AutoCorrelationHeatmap::new(&params).unwrap();

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

        let mut sine_ind = AutoCorrelationHeatmap::new(&AutoCorrelationHeatmapParams {
            length: 20,
            smoothing_period: 20.0,
            max_lag: 40,
            ..Default::default()
        })
        .unwrap();
        let mut column = Heatmap::empty(0, 0.0, 40.0, 1.0);
        for step in 0..1000 {
            column = sine_ind.update((2.0 * PI * (step as f64) / 20.0).sin(), step);
        }
        for lag in 0..=40 {
            assert_close_tol(
                column.values[lag],
                (2.0 * PI * (lag as f64) / 20.0).cos(),
                1e-11,
                &format!("sine[{lag}]"),
            );
        }

        let mut flat_ind = AutoCorrelationHeatmap::new(&AutoCorrelationHeatmapParams {
            length: 3,
            smoothing_period: 20.0,
            max_lag: 2,
            ..Default::default()
        })
        .unwrap();
        let mut flat_out = Heatmap::empty(0, 0.0, 2.0, 1.0);
        for _ in 0..30 {
            flat_out = flat_ind.update(0.0, 0);
        }
        assert_eq!(flat_out.values, vec![0.0, 0.0, 0.0]);
        flat_ind.correlations = vec![0.25, -0.5, 0.75];
        let retained = flat_ind.update(0.0, 0);
        assert_eq!(retained.values, vec![0.25, -0.5, 0.75]);
        assert_eq!(retained.value_min, -0.5);
        assert_eq!(retained.value_max, 0.75);

        let mnemonic_cases = [
            (default_params(), "ach(20, 20, 99)"),
            (
                AutoCorrelationHeatmapParams {
                    length: 20,
                    smoothing_period: 10.5,
                    max_lag: 0,
                    ..Default::default()
                },
                "ach(20, 10.5, 0)",
            ),
            (
                AutoCorrelationHeatmapParams {
                    bar_component: Some(BarComponent::Median),
                    ..Default::default()
                },
                "ach(20, 20, 99, hl/2)",
            ),
            (
                AutoCorrelationHeatmapParams {
                    quote_component: Some(QuoteComponent::Bid),
                    ..Default::default()
                },
                "ach(20, 20, 99, b)",
            ),
            (
                AutoCorrelationHeatmapParams {
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "ach(20, 20, 99, v)",
            ),
            (
                AutoCorrelationHeatmapParams {
                    bar_component: Some(BarComponent::Typical),
                    quote_component: Some(QuoteComponent::Ask),
                    ..Default::default()
                },
                "ach(20, 20, 99, hlc/3, a)",
            ),
            (
                AutoCorrelationHeatmapParams {
                    bar_component: Some(BarComponent::Open),
                    quote_component: Some(QuoteComponent::Weighted),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "ach(20, 20, 99, o, (bbs+aas)/(bs+as), v)",
            ),
        ];
        for (case_params, want_mnemonic) in mnemonic_cases {
            assert_eq!(mnemonic(&case_params), want_mnemonic);
            let ind = AutoCorrelationHeatmap::new(&case_params).unwrap();
            let meta = ind.metadata();
            assert_eq!(meta.identifier, Identifier::AutoCorrelationHeatmap);
            assert_eq!(meta.mnemonic, want_mnemonic);
            assert_eq!(meta.description, format!("Autocorrelation heatmap {want_mnemonic}"));
            assert_eq!(meta.outputs.len(), 1);
            assert_eq!(meta.outputs[0].kind, AutoCorrelationHeatmapOutput::Value as i32);
            assert_eq!(meta.outputs[0].shape, Shape::Heatmap);
            assert_eq!(meta.outputs[0].mnemonic, want_mnemonic);
        }

        let desc = descriptor_of(Identifier::AutoCorrelationHeatmap).unwrap();
        assert_eq!(desc.family, "John Ehlers");
        assert_eq!(desc.adaptivity, Adaptivity::Static);
        assert_eq!(desc.input_requirement, InputRequirement::ScalarInput);
        assert_eq!(desc.volume_usage, VolumeUsage::NoVolume);
        assert_eq!(desc.outputs.len(), 1);
        assert_eq!(desc.outputs[0].shape, Shape::Heatmap);
        assert_eq!(desc.outputs[0].role, Role::Correlation);
        assert_eq!(desc.outputs[0].pane, Pane::Own);

        let mut scalar_ind = AutoCorrelationHeatmap::new_default();
        let mut bar_ind = AutoCorrelationHeatmap::new_default();
        let mut quote_ind = AutoCorrelationHeatmap::new_default();
        let mut trade_ind = AutoCorrelationHeatmap::new_default();
        let mut open_ind = AutoCorrelationHeatmap::new(&AutoCorrelationHeatmapParams {
            bar_component: Some(BarComponent::Open),
            ..Default::default()
        })
        .unwrap();
        let mut open_ref = AutoCorrelationHeatmap::new_default();

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
