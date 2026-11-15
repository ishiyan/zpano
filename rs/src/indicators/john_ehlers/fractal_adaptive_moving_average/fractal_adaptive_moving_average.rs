use std::any::Any;
use std::f64::consts::LN_2;

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

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Fractal Adaptive Moving Average indicator.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum FractalAdaptiveMovingAverageOutput {
    /// The scalar value of the fractal adaptive moving average.
    Value = 1,
    /// The scalar value of the raw guarded fractal-dimension estimate.
    Fdim = 2,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Fractal Adaptive Moving Average indicator.
#[derive(Debug, Clone, Copy)]
pub struct FractalAdaptiveMovingAverageParams {
    /// Lookback length (must be an even integer >= 2). Default is 16.
    pub length: i64,
    /// Slowest boundary smoothing factor (finite, 0 < A <= 1). Default is 0.01.
    pub slowest_smoothing_factor: f64,
    /// Alpha exponent scale mode: 1 for full-precision ln(A), 0 for literal 2005 -4.6 (requires A == 0.01). Default is 1.
    pub coefficient_mode: i32,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

pub type Params = FractalAdaptiveMovingAverageParams;

impl Default for FractalAdaptiveMovingAverageParams {
    fn default() -> Self {
        Self {
            length: 16,
            slowest_smoothing_factor: 0.01,
            coefficient_mode: 1,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Returns default parameters for the Fractal Adaptive Moving Average indicator.
pub fn default_params() -> FractalAdaptiveMovingAverageParams {
    FractalAdaptiveMovingAverageParams::default()
}

fn format_number(number: f64) -> String {
    let normalized = if number == 0.0 { 0.0 } else { number };
    format!("{normalized}")
}

fn format_mnemonic(prefix: &str, params: &FractalAdaptiveMovingAverageParams) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let mode_tag = if params.coefficient_mode == 0 {
        ", mode=0"
    } else {
        ""
    };
    let suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    let slowest_text = format_number(params.slowest_smoothing_factor);
    format!("{prefix}({}, {slowest_text}{mode_tag}{suffix})", params.length)
}

/// Builds the primary FractalAdaptiveMovingAverage mnemonic string.
pub fn mnemonic(params: &FractalAdaptiveMovingAverageParams) -> String {
    format_mnemonic("frama", params)
}

/// Builds the FractalAdaptiveMovingAverage dimension mnemonic string.
pub fn dimension_mnemonic(params: &FractalAdaptiveMovingAverageParams) -> String {
    format_mnemonic("framaDim", params)
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

fn box_range(highs: &[f64], lows: &[f64], start: usize, end: usize) -> Result<f64, String> {
    let mut highest = highs[start];
    let mut lowest = lows[start];
    for index in (start + 1)..end {
        if highs[index] > highest {
            highest = highs[index];
        }
        if lows[index] < lowest {
            lowest = lows[index];
        }
    }
    let span = highest - lowest;
    if !span.is_finite() {
        return Err("price-range arithmetic overflowed".to_string());
    }
    Ok(span)
}

/// Ehlers' Fractal Adaptive Moving Average (FRAMA).
pub struct FractalAdaptiveMovingAverage {
    metadata: Metadata,
    length: usize,
    half_length: usize,
    warmup_count: usize,
    slowest_smoothing_factor: f64,
    exponent_scale: f64,
    count: usize,
    highs: Vec<f64>,
    lows: Vec<f64>,
    dimension_state: f64,
    alpha_state: f64,
    filter_state: f64,
    current_value: f64,
    current_dimension: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl FractalAdaptiveMovingAverage {
    /// Creates a new FRAMA with Ehlers' default parameters.
    pub fn default_instance() -> Result<Self, String> {
        Self::new(&FractalAdaptiveMovingAverageParams::default())
    }

    /// Creates a new FRAMA from the supplied parameters.
    pub fn new(params: &FractalAdaptiveMovingAverageParams) -> Result<Self, String> {
        const INVALID: &str = "invalid fractal adaptive moving average parameters";

        if params.length < 2 || params.length % 2 != 0 {
            return Err(format!("{INVALID}: length must be an even integer >= 2"));
        }
        let length = usize::try_from(params.length)
            .map_err(|_| format!("{INVALID}: length must be an even integer >= 2"))?;

        let slowest = params.slowest_smoothing_factor;
        if !slowest.is_finite() || slowest <= 0.0 || slowest > 1.0 {
            return Err(format!(
                "{INVALID}: slowest_smoothing_factor must satisfy 0 < A <= 1"
            ));
        }

        if params.coefficient_mode != 0 && params.coefficient_mode != 1 {
            return Err(format!(
                "{INVALID}: coefficient_mode must be 0 (article) or 1 (full precision)"
            ));
        }
        if params.coefficient_mode == 0 && slowest != 0.01 {
            return Err(format!(
                "{INVALID}: literal article mode requires slowest_smoothing_factor=0.01"
            ));
        }

        let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bar_comp);
        let quote_func = quote_component_value(quote_comp);
        let trade_func = trade_component_value(trade_comp);

        let exponent_scale = if params.coefficient_mode == 1 {
            slowest.ln()
        } else {
            -4.6
        };

        let mnemonic_text = mnemonic(params);
        let mnemonic_fdim = dimension_mnemonic(params);
        let prefix = "Fractal adaptive moving average ";
        let description_text = format!("{prefix}{mnemonic_text}");
        let description_fdim = format!("{prefix}{mnemonic_fdim}");

        let metadata = build_metadata(
            Identifier::FractalAdaptiveMovingAverage,
            &mnemonic_text,
            &description_text,
            &[
                OutputText {
                    mnemonic: mnemonic_text.clone(),
                    description: description_text.clone(),
                },
                OutputText {
                    mnemonic: mnemonic_fdim,
                    description: description_fdim,
                },
            ],
        );

        Ok(Self {
            metadata,
            length,
            half_length: length / 2,
            warmup_count: length - 1,
            slowest_smoothing_factor: slowest,
            exponent_scale,
            count: 0,
            highs: Vec::with_capacity(length),
            lows: Vec::with_capacity(length),
            dimension_state: 0.0,
            alpha_state: 1.0,
            filter_state: 0.0,
            current_value: f64::NAN,
            current_dimension: f64::NAN,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Clears all streaming histories and priming state while preserving configuration.
    pub fn reset(&mut self) {
        self.count = 0;
        self.highs.clear();
        self.lows.clear();
        self.dimension_state = 0.0;
        self.alpha_state = 1.0;
        self.filter_state = 0.0;
        self.current_value = f64::NAN;
        self.current_dimension = f64::NAN;
    }

    /// Returns the number of initial NaN outputs before priming (`length - 1`).
    pub fn warmup_bars(&self) -> usize {
        self.warmup_count
    }

    /// Returns the number of accepted finite observations seen since start or reset.
    pub fn samples_seen(&self) -> usize {
        self.count
    }

    /// Returns the latest accepted FRAMA value (`NaN` before priming).
    pub fn value(&self) -> f64 {
        self.current_value
    }

    /// Returns the latest accepted raw fractal-dimension estimate (`NaN` before priming).
    pub fn fractal_dimension(&self) -> f64 {
        self.current_dimension
    }

    /// Returns the latest diagnostic smoothing factor (`NaN` before priming).
    pub fn alpha(&self) -> f64 {
        if self.is_primed() {
            self.alpha_state
        } else {
            f64::NAN
        }
    }

    /// Fallible update taking `(sample, sample_high, sample_low)`.
    /// Validates inputs atomically without mutating state on error or `NaN`.
    pub fn try_update(
        &mut self,
        sample: f64,
        sample_high: f64,
        sample_low: f64,
    ) -> Result<f64, String> {
        if sample.is_infinite() {
            return Err("price must be finite or NaN".to_string());
        }
        if sample_high.is_infinite() {
            return Err("high must be finite or NaN".to_string());
        }
        if sample_low.is_infinite() {
            return Err("low must be finite or NaN".to_string());
        }
        if sample.is_nan() || sample_high.is_nan() || sample_low.is_nan() {
            return Ok(f64::NAN);
        }
        if sample_high < sample_low {
            return Err("high must be >= low".to_string());
        }

        let keep = self.highs.len().min(self.length - 1);
        let mut next_highs = Vec::with_capacity(keep + 1);
        let mut next_lows = Vec::with_capacity(keep + 1);
        next_highs.push(sample_high);
        next_highs.extend_from_slice(&self.highs[..keep]);
        next_lows.push(sample_low);
        next_lows.extend_from_slice(&self.lows[..keep]);

        let next_count = self.count + 1;
        let mut dimension = self.dimension_state;
        let mut alpha = self.alpha_state;

        if next_count >= self.length {
            let range1 = box_range(&next_highs, &next_lows, 0, self.half_length)?;
            let range2 = box_range(&next_highs, &next_lows, self.half_length, self.length)?;
            let range3 = box_range(&next_highs, &next_lows, 0, self.length)?;

            let half_float = self.half_length as f64;
            let length_float = self.length as f64;
            let n1 = range1 / half_float;
            let n2 = range2 / half_float;
            let n3 = range3 / length_float;

            if n1 > 0.0 && n2 > 0.0 && n3 > 0.0 {
                let combined = n1 + n2;
                if !combined.is_finite() {
                    return Err("combined box-range arithmetic overflowed".to_string());
                }
                dimension = (combined.ln() - n3.ln()) / LN_2;
                if !dimension.is_finite() {
                    return Err("dimension arithmetic overflowed".to_string());
                }
            }

            if dimension <= 1.0 {
                alpha = 1.0;
            } else {
                let raw_alpha = (self.exponent_scale * (dimension - 1.0)).exp();
                alpha = raw_alpha.max(self.slowest_smoothing_factor).min(1.0);
            }
        }

        let next_value = if next_count <= self.length {
            sample
        } else {
            let filtered = alpha * sample + (1.0 - alpha) * self.filter_state;
            if !filtered.is_finite() {
                return Err("FRAMA arithmetic overflowed".to_string());
            }
            filtered
        };

        self.highs = next_highs;
        self.lows = next_lows;
        self.count = next_count;
        self.dimension_state = dimension;
        self.alpha_state = alpha;
        self.filter_state = next_value;
        if next_count >= self.length {
            self.current_value = next_value;
            self.current_dimension = dimension;
            Ok(next_value)
        } else {
            Ok(f64::NAN)
        }
    }

    /// Updates the indicator with `(sample, sample_high, sample_low)`, panicking on invalid input.
    pub fn update(&mut self, sample: f64, sample_high: f64, sample_low: f64) -> f64 {
        self.try_update(sample, sample_high, sample_low)
            .expect("FractalAdaptiveMovingAverage::update failed")
    }

    /// Fallible scalar update using `price` for price, high, and low.
    pub fn try_update_price(&mut self, price: f64) -> Result<f64, String> {
        self.try_update(price, price, price)
    }

    /// Scalar update using `price` for price, high, and low, panicking on invalid input.
    pub fn update_price(&mut self, price: f64) -> f64 {
        self.update(price, price, price)
    }

    fn update_entity(
        &mut self,
        time: i64,
        sample: f64,
        sample_high: f64,
        sample_low: f64,
    ) -> Output {
        let frama = self.update(sample, sample_high, sample_low);
        let fdim = if frama.is_nan() {
            f64::NAN
        } else {
            self.current_dimension
        };
        vec![
            Box::new(Scalar::new(time, frama)) as Box<dyn Any>,
            Box::new(Scalar::new(time, fdim)) as Box<dyn Any>,
        ]
    }
}

impl Indicator for FractalAdaptiveMovingAverage {
    fn is_primed(&self) -> bool {
        self.count >= self.length
    }

    fn metadata(&self) -> Metadata {
        self.metadata.clone()
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let extracted = sample.value;
        self.update_entity(sample.time, extracted, extracted, extracted)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let extracted = (self.bar_func)(sample);
        self.update_entity(sample.time, extracted, sample.high, sample.low)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let extracted = (self.quote_func)(sample);
        self.update_entity(sample.time, extracted, sample.ask_price, sample.bid_price)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let extracted = (self.trade_func)(sample);
        self.update_entity(sample.time, extracted, extracted, extracted)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::core::outputs::shape::Shape;
    use crate::indicators::factory::create_indicator;
    use crate::indicators::john_ehlers::fractal_adaptive_moving_average::testdata::testdata as fixture;

    fn assert_close(expected: f64, actual: f64, context: &str) {
        if expected.is_nan() {
            assert!(actual.is_nan(), "{context}: expected NaN, got {actual}");
        } else {
            let diff = (expected - actual).abs();
            let bound = 1e-12_f64.max(expected.abs() * 1e-12);
            assert!(
                diff <= bound,
                "{context}: expected {expected}, got {actual} (diff={diff}, bound={bound})"
            );
        }
    }

    fn independent_batch_frama(
        prices: &[f64],
        highs: &[f64],
        lows: &[f64],
        length: usize,
        slowest: f64,
        coefficient_mode: i32,
    ) -> (Vec<f64>, Vec<f64>) {
        let half = length / 2;
        let scale = if coefficient_mode == 1 {
            slowest.ln()
        } else {
            -4.6
        };
        let mut history_high: Vec<f64> = Vec::new();
        let mut history_low: Vec<f64> = Vec::new();
        let mut accepted_count = 0usize;
        let mut dimension_state = 0.0;
        let mut alpha_state = 1.0;
        let mut filter_state = 0.0;
        let mut values = Vec::with_capacity(prices.len());
        let mut dimensions = Vec::with_capacity(prices.len());

        for step in 0..prices.len() {
            let price = prices[step];
            let high = highs[step];
            let low = lows[step];
            if price.is_nan() || high.is_nan() || low.is_nan() {
                values.push(f64::NAN);
                dimensions.push(f64::NAN);
                continue;
            }
            accepted_count += 1;
            history_high.insert(0, high);
            history_low.insert(0, low);
            if history_high.len() > length {
                history_high.truncate(length);
                history_low.truncate(length);
            }

            if accepted_count >= length {
                let r1 = box_range(&history_high, &history_low, 0, half).unwrap();
                let r2 = box_range(&history_high, &history_low, half, length).unwrap();
                let r3 = box_range(&history_high, &history_low, 0, length).unwrap();
                let n1 = r1 / (half as f64);
                let n2 = r2 / (half as f64);
                let n3 = r3 / (length as f64);
                if n1 > 0.0 && n2 > 0.0 && n3 > 0.0 {
                    dimension_state = ((n1 + n2).ln() - n3.ln()) / LN_2;
                }
                alpha_state = if dimension_state <= 1.0 {
                    1.0
                } else {
                    (scale * (dimension_state - 1.0)).exp().max(slowest).min(1.0)
                };
            }

            filter_state = if accepted_count <= length {
                price
            } else {
                alpha_state * price + (1.0 - alpha_state) * filter_state
            };

            if accepted_count >= length {
                values.push(filter_state);
                dimensions.push(dimension_state);
            } else {
                values.push(f64::NAN);
                dimensions.push(f64::NAN);
            }
        }
        (values, dimensions)
    }

    fn reference_cases() -> Vec<(FractalAdaptiveMovingAverageParams, Vec<f64>, Vec<f64>)> {
        vec![
            (
                FractalAdaptiveMovingAverageParams {
                    length: 16,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l16_a0_01_c1_value(),
                fixture::expected_l16_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 2,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l2_a0_01_c1_value(),
                fixture::expected_l2_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 4,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l4_a0_01_c1_value(),
                fixture::expected_l4_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 8,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l8_a0_01_c1_value(),
                fixture::expected_l8_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 20,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l20_a0_01_c1_value(),
                fixture::expected_l20_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 30,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l30_a0_01_c1_value(),
                fixture::expected_l30_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 64,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l64_a0_01_c1_value(),
                fixture::expected_l64_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 128,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l128_a0_01_c1_value(),
                fixture::expected_l128_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 250,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l250_a0_01_c1_value(),
                fixture::expected_l250_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 252,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l252_a0_01_c1_value(),
                fixture::expected_l252_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 254,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l254_a0_01_c1_value(),
                fixture::expected_l254_a0_01_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 16,
                    slowest_smoothing_factor: 0.001,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l16_a0_001_c1_value(),
                fixture::expected_l16_a0_001_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 16,
                    slowest_smoothing_factor: 0.1,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l16_a0_1_c1_value(),
                fixture::expected_l16_a0_1_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 16,
                    slowest_smoothing_factor: 1.0,
                    coefficient_mode: 1,
                    ..Default::default()
                },
                fixture::expected_l16_a1_c1_value(),
                fixture::expected_l16_a1_c1_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 16,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 0,
                    ..Default::default()
                },
                fixture::expected_l16_a0_01_c0_value(),
                fixture::expected_l16_a0_01_c0_fdim(),
            ),
            (
                FractalAdaptiveMovingAverageParams {
                    length: 30,
                    slowest_smoothing_factor: 0.01,
                    coefficient_mode: 0,
                    ..Default::default()
                },
                fixture::expected_l30_a0_01_c0_value(),
                fixture::expected_l30_a0_01_c0_fdim(),
            ),
        ]
    }

    #[test]
    fn test_reference_cases_and_batch_oracle() {
        let closes = fixture::test_input();
        let highs = fixture::test_high();
        let lows = fixture::test_low();
        let opens = fixture::test_open();
        assert_eq!(closes.len(), 252);

        for (case_index, (options, expected_values, expected_fdims)) in
            reference_cases().into_iter().enumerate()
        {
            let mut indicator = FractalAdaptiveMovingAverage::new(&options).unwrap();
            assert_eq!(indicator.warmup_bars(), (options.length - 1) as usize);

            let (oracle_values, oracle_fdims) = independent_batch_frama(
                &closes,
                &closes,
                &closes,
                options.length as usize,
                options.slowest_smoothing_factor,
                options.coefficient_mode,
            );

            for bar_index in 0..closes.len() {
                let actual_value = indicator.update_price(closes[bar_index]);
                let actual_fdim = indicator.fractal_dimension();
                let primed_expected = (bar_index + 1) >= (options.length as usize);
                assert_eq!(indicator.is_primed(), primed_expected);
                assert_eq!(indicator.samples_seen(), bar_index + 1);

                assert_close(
                    expected_values[bar_index],
                    actual_value,
                    &format!("case {case_index} value[{bar_index}]"),
                );
                assert_close(
                    expected_fdims[bar_index],
                    actual_fdim,
                    &format!("case {case_index} fdim[{bar_index}]"),
                );
                assert_close(
                    oracle_values[bar_index],
                    actual_value,
                    &format!("case {case_index} oracle value[{bar_index}]"),
                );
                assert_close(
                    oracle_fdims[bar_index],
                    actual_fdim,
                    &format!("case {case_index} oracle fdim[{bar_index}]"),
                );
            }

            indicator.reset();
            assert!(!indicator.is_primed());
            assert_eq!(indicator.samples_seen(), 0);
            assert!(indicator.value().is_nan());
            assert!(indicator.fractal_dimension().is_nan());
            assert!(indicator.alpha().is_nan());

            let midpoints: Vec<f64> = highs
                .iter()
                .zip(lows.iter())
                .map(|(h, l)| 0.5 * (h + l))
                .collect();
            let (bar_oracle_values, bar_oracle_fdims) = independent_batch_frama(
                &midpoints,
                &highs,
                &lows,
                options.length as usize,
                options.slowest_smoothing_factor,
                options.coefficient_mode,
            );

            let mut bar_indicator =
                FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
                    bar_component: Some(BarComponent::Median),
                    ..options
                })
                .unwrap();
            for bar_index in 0..closes.len() {
                let bar = Bar::new(
                    bar_index as i64,
                    opens[bar_index],
                    highs[bar_index],
                    lows[bar_index],
                    closes[bar_index],
                    1000.0,
                );
                let outputs = bar_indicator.update_bar(&bar);
                let out_value = outputs[0].downcast_ref::<Scalar>().unwrap().value;
                let out_fdim = outputs[1].downcast_ref::<Scalar>().unwrap().value;
                assert_close(
                    bar_oracle_values[bar_index],
                    out_value,
                    &format!("case {case_index} bar value[{bar_index}]"),
                );
                assert_close(
                    bar_oracle_fdims[bar_index],
                    out_fdim,
                    &format!("case {case_index} bar fdim[{bar_index}]"),
                );
            }
        }
    }

    #[test]
    fn test_invariants_and_unclamped_dimension() {
        let mut constant_ind = FractalAdaptiveMovingAverage::default_instance().unwrap();
        for step in 0..32 {
            let out = constant_ind.update_price(-42.5);
            if step < 15 {
                assert!(out.is_nan());
            } else {
                assert_close(-42.5, out, "constant value");
                assert_close(0.0, constant_ind.fractal_dimension(), "constant fdim");
                assert_close(1.0, constant_ind.alpha(), "constant alpha");
            }
        }

        let closes = fixture::test_input();
        let highs = fixture::test_high();
        let lows = fixture::test_low();
        let scale = 2.75;
        let shift = -500.0;
        let mut base_ind = FractalAdaptiveMovingAverage::default_instance().unwrap();
        let mut affine_ind = FractalAdaptiveMovingAverage::default_instance().unwrap();
        for step in 0..closes.len() {
            let v1 = base_ind.update(closes[step], highs[step], lows[step]);
            let v2 = affine_ind.update(
                scale * closes[step] + shift,
                scale * highs[step] + shift,
                scale * lows[step] + shift,
            );
            if step >= 15 {
                assert_close(scale * v1 + shift, v2, "affine value");
                assert_close(
                    base_ind.fractal_dimension(),
                    affine_ind.fractal_dimension(),
                    "affine fdim",
                );
            }
        }

        let mut gap_ind = FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length: 4,
            ..Default::default()
        })
        .unwrap();
        gap_ind.update(0.5, 1.0, 0.0);
        gap_ind.update(0.5, 1.0, 0.0);
        gap_ind.update(100.5, 101.0, 100.0);
        let primed_seed = gap_ind.update(100.5, 101.0, 100.0);
        assert_close(100.5, primed_seed, "bar length outputs unfiltered seed");
        assert!(gap_ind.fractal_dimension() < 1.0);
        assert_close(1.0, gap_ind.alpha(), "dimension <= 1 clamps alpha to 1");

        gap_ind.update(50.0, 50.0, 50.0);
        let held_dimension = gap_ind.fractal_dimension();
        gap_ind.update(50.0, 50.0, 50.0);
        assert_close(
            held_dimension,
            gap_ind.fractal_dimension(),
            "degenerate half holds previous dimension",
        );
    }

    #[test]
    fn test_nan_and_error_atomicity() {
        let mut indicator = FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length: 4,
            ..Default::default()
        })
        .unwrap();
        for step in 1..=4 {
            let price = (step * 10) as f64;
            indicator.update(price, price + 2.0, price - 2.0);
        }
        let saved_value = indicator.value();
        let saved_dimension = indicator.fractal_dimension();
        let saved_alpha = indicator.alpha();
        let saved_count = indicator.samples_seen();

        assert!(indicator.update(f64::NAN, 50.0, 40.0).is_nan());
        assert!(indicator.update(45.0, f64::NAN, 40.0).is_nan());
        assert!(indicator.update(45.0, 50.0, f64::NAN).is_nan());
        let scalar_nan = indicator.update_scalar(&Scalar::new(99, f64::NAN));
        assert!(scalar_nan[0].downcast_ref::<Scalar>().unwrap().value.is_nan());
        assert!(scalar_nan[1].downcast_ref::<Scalar>().unwrap().value.is_nan());

        assert!(indicator.try_update(f64::INFINITY, 50.0, 40.0).is_err());
        assert!(indicator.try_update(45.0, f64::INFINITY, 40.0).is_err());
        assert!(indicator.try_update(45.0, 50.0, f64::NEG_INFINITY).is_err());
        assert!(indicator.try_update(45.0, 39.0, 40.0).is_err());
        assert!(indicator.try_update(0.0, f64::MAX, -f64::MAX).is_err());

        assert_eq!(indicator.samples_seen(), saved_count);
        assert_close(saved_value, indicator.value(), "atomic value");
        assert_close(saved_dimension, indicator.fractal_dimension(), "atomic fdim");
        assert_close(saved_alpha, indicator.alpha(), "atomic alpha");
    }

    #[test]
    fn test_validation_metadata_mnemonics_entities_and_factory() {
        for bad_length in [-4, -1, 0, 1, 3, 15, 17] {
            assert!(FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
                length: bad_length,
                ..Default::default()
            })
            .is_err());
        }
        for bad_slowest in [0.0, -0.01, 1.01, f64::NAN, f64::INFINITY, f64::NEG_INFINITY] {
            assert!(FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
                slowest_smoothing_factor: bad_slowest,
                ..Default::default()
            })
            .is_err());
        }
        for bad_mode in [-1, 2, 99] {
            assert!(FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
                coefficient_mode: bad_mode,
                ..Default::default()
            })
            .is_err());
        }
        assert!(FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            slowest_smoothing_factor: 0.05,
            coefficient_mode: 0,
            ..Default::default()
        })
        .is_err());

        let default_ind = FractalAdaptiveMovingAverage::default_instance().unwrap();
        let meta = default_ind.metadata();
        assert_eq!(meta.identifier, Identifier::FractalAdaptiveMovingAverage);
        assert_eq!(meta.mnemonic, "frama(16, 0.01)");
        assert_eq!(
            meta.description,
            "Fractal adaptive moving average frama(16, 0.01)"
        );
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(
            meta.outputs[0].kind,
            FractalAdaptiveMovingAverageOutput::Value as i32
        );
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
        assert_eq!(meta.outputs[0].mnemonic, "frama(16, 0.01)");
        assert_eq!(
            meta.outputs[1].kind,
            FractalAdaptiveMovingAverageOutput::Fdim as i32
        );
        assert_eq!(meta.outputs[1].shape, Shape::Scalar);
        assert_eq!(meta.outputs[1].mnemonic, "framaDim(16, 0.01)");

        assert_eq!(format_number(-0.0), "0");
        assert_eq!(
            mnemonic(&FractalAdaptiveMovingAverageParams {
                length: 30,
                slowest_smoothing_factor: 0.01,
                coefficient_mode: 0,
                ..Default::default()
            }),
            "frama(30, 0.01, mode=0)"
        );
        assert_eq!(
            mnemonic(&FractalAdaptiveMovingAverageParams {
                bar_component: Some(BarComponent::Median),
                ..Default::default()
            }),
            "frama(16, 0.01, hl/2)"
        );
        assert_eq!(
            mnemonic(&FractalAdaptiveMovingAverageParams {
                quote_component: Some(QuoteComponent::Weighted),
                ..Default::default()
            }),
            "frama(16, 0.01, (bbs+aas)/(bs+as))"
        );
        assert_eq!(
            mnemonic(&FractalAdaptiveMovingAverageParams {
                trade_component: Some(TradeComponent::Volume),
                ..Default::default()
            }),
            "frama(16, 0.01, v)"
        );
        assert_eq!(
            mnemonic(&FractalAdaptiveMovingAverageParams {
                bar_component: Some(BarComponent::Typical),
                quote_component: Some(QuoteComponent::Bid),
                ..Default::default()
            }),
            "frama(16, 0.01, hlc/3, b)"
        );
        assert_eq!(
            dimension_mnemonic(&FractalAdaptiveMovingAverageParams {
                length: 20,
                slowest_smoothing_factor: 0.01,
                coefficient_mode: 0,
                bar_component: Some(BarComponent::Open),
                trade_component: Some(TradeComponent::Volume),
                ..Default::default()
            }),
            "framaDim(20, 0.01, mode=0, o, v)"
        );

        let mut quote_ind = FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length: 2,
            ..Default::default()
        })
        .unwrap();
        quote_ind.update_quote(&Quote::new(1, 9.0, 11.0, 10.0, 10.0));
        let quote_out = quote_ind.update_quote(&Quote::new(2, 19.0, 21.0, 10.0, 10.0));
        assert_close(
            20.0,
            quote_out[0].downcast_ref::<Scalar>().unwrap().value,
            "quote mid seed",
        );

        let mut trade_ind = FractalAdaptiveMovingAverage::new(&FractalAdaptiveMovingAverageParams {
            length: 2,
            trade_component: Some(TradeComponent::Volume),
            ..Default::default()
        })
        .unwrap();
        trade_ind.update_trade(&Trade::new(1, 100.0, 50.0));
        let trade_out = trade_ind.update_trade(&Trade::new(2, 105.0, 75.0));
        assert_close(
            75.0,
            trade_out[0].downcast_ref::<Scalar>().unwrap().value,
            "trade volume seed",
        );

        let from_default =
            create_indicator(Identifier::FractalAdaptiveMovingAverage, "{}").unwrap();
        assert_eq!(from_default.metadata().mnemonic, "frama(16, 0.01)");

        let from_custom = create_indicator(
            Identifier::FractalAdaptiveMovingAverage,
            r#"{"length": 30, "slowestSmoothingFactor": 0.01, "coefficientMode": 0, "barComponent": 5}"#,
        )
        .unwrap();
        assert_eq!(
            from_custom.metadata().mnemonic,
            "frama(30, 0.01, mode=0, hl/2)"
        );

        for bad_json in [
            r#"{"length": 17}"#,
            r#"{"length": 16.5}"#,
            r#"{"length": true}"#,
            r#"{"slowestSmoothingFactor": 0}"#,
            r#"{"slowestSmoothingFactor": "0.01"}"#,
            r#"{"coefficientMode": 2}"#,
            r#"{"coefficientMode": false}"#,
            r#"{"slowestSmoothingFactor": 0.05, "coefficientMode": 0}"#,
        ] {
            assert!(
                create_indicator(Identifier::FractalAdaptiveMovingAverage, bad_json).is_err(),
                "expected error for {bad_json}"
            );
        }
    }
}
