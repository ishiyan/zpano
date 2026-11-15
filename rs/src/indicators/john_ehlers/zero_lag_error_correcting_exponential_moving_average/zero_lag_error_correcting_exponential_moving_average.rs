use std::any::Any;

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

const ALPHA_EPSILON: f64 = 1e-8;
const LEAST_ERROR_SENTINEL: f64 = 1_000_000.0;

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Output indices for Ehlers and Way's Zero-Lag Error-Correcting Exponential Moving Average.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ZeroLagErrorCorrectingExponentialMovingAverageOutput {
    /// The scalar value of the error-corrected exponential moving average (EC).
    Value = 1,
    /// The scalar value of the baseline exponential moving average (EMA).
    Ema = 2,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Length-based parameters to create an instance of ZeroLagErrorCorrectingExponentialMovingAverage.
#[derive(Debug, Clone, Copy)]
pub struct ZeroLagErrorCorrectingExponentialMovingAverageLengthParams {
    /// Equivalent EMA length L (>= 1), defining alpha = 2 / (L + 1). Default is 20.
    pub length: i64,
    /// Nonnegative integer gain bound in tenths (candidates span [-G/10, +G/10]). Default is 50.
    pub gain_limit: i32,
    /// Bar component to extract. `None` means default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means default (Price).
    pub trade_component: Option<TradeComponent>,
}

/// Smoothing-factor-based parameters to create an instance of ZeroLagErrorCorrectingExponentialMovingAverage.
#[derive(Debug, Clone, Copy)]
pub struct ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams {
    /// Baseline EMA smoothing factor alpha in (0, 1]. Default is 2 / 21.
    pub smoothing_factor: f64,
    /// Nonnegative integer gain bound in tenths (candidates span [-G/10, +G/10]). Default is 50.
    pub gain_limit: i32,
    /// Bar component to extract. `None` means default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means default (Price).
    pub trade_component: Option<TradeComponent>,
}

pub type LengthParams = ZeroLagErrorCorrectingExponentialMovingAverageLengthParams;
pub type SmoothingFactorParams = ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams;
pub type ZeroLagErrorCorrectingExponentialMovingAverageParams =
    ZeroLagErrorCorrectingExponentialMovingAverageLengthParams;
pub type Params = ZeroLagErrorCorrectingExponentialMovingAverageLengthParams;

impl Default for ZeroLagErrorCorrectingExponentialMovingAverageLengthParams {
    fn default() -> Self {
        Self {
            length: 20,
            gain_limit: 50,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

impl Default for ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams {
    fn default() -> Self {
        Self {
            smoothing_factor: 2.0 / 21.0,
            gain_limit: 50,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Returns default length-based parameters for ZECEMA.
pub fn default_length_params() -> ZeroLagErrorCorrectingExponentialMovingAverageLengthParams {
    ZeroLagErrorCorrectingExponentialMovingAverageLengthParams::default()
}

/// Returns default smoothing-factor-based parameters for ZECEMA.
pub fn default_smoothing_factor_params(
) -> ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams {
    ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams::default()
}

/// Returns default parameters for ZECEMA.
pub fn default_params() -> ZeroLagErrorCorrectingExponentialMovingAverageLengthParams {
    default_length_params()
}

/// Computes the equivalent integer length `round_half_away_from_zero(2 / alpha) - 1`.
pub fn equivalent_length_from_alpha(alpha: f64) -> i64 {
    if !(alpha >= ALPHA_EPSILON) {
        return i64::MAX;
    }
    let rounded = (2.0_f64 / alpha).round();
    if rounded >= i64::MAX as f64 {
        return i64::MAX;
    }
    (rounded as i64) - 1
}

fn format_length_mnemonic(
    prefix: &str,
    params: &ZeroLagErrorCorrectingExponentialMovingAverageLengthParams,
) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    format!("{prefix}({}, {}{suffix})", params.length, params.gain_limit)
}

fn format_smoothing_factor_mnemonic(
    prefix: &str,
    params: &ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    let equiv_length = equivalent_length_from_alpha(params.smoothing_factor);
    format!(
        "{prefix}({equiv_length}, {:.8}, {}{suffix})",
        params.smoothing_factor, params.gain_limit
    )
}

/// Builds the primary EC mnemonic string for length-based parameters.
pub fn mnemonic(params: &ZeroLagErrorCorrectingExponentialMovingAverageLengthParams) -> String {
    format_length_mnemonic("zecema", params)
}

/// Builds the baseline EMA mnemonic string for length-based parameters.
pub fn ema_mnemonic(params: &ZeroLagErrorCorrectingExponentialMovingAverageLengthParams) -> String {
    format_length_mnemonic("zecemaEma", params)
}

/// Builds the primary EC mnemonic string for smoothing-factor parameters.
pub fn smoothing_factor_mnemonic(
    params: &ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
) -> String {
    format_smoothing_factor_mnemonic("zecema", params)
}

/// Builds the baseline EMA mnemonic string for smoothing-factor parameters.
pub fn smoothing_factor_ema_mnemonic(
    params: &ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
) -> String {
    format_smoothing_factor_mnemonic("zecemaEma", params)
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers and Way's Zero-Lag Error-Correcting Exponential Moving Average (ZECEMA).
///
/// Follows the November 2010 TASC EasyLanguage listing ("Zero Lag (Well, Almost)"):
/// starting from explicit zero EMA and EC states, each accepted finite sample updates
/// the baseline EMA once and searches integer-tenths gains `index / 10` for inclusive
/// integers `[-gain_limit, +gain_limit]` against the frozen previous-bar EC using
/// strict improvement (`error_value < least_error`) from the literal `1_000_000`
/// sentinel. Both EC (`Value`) and baseline EMA (`Ema`) outputs are available from
/// the first accepted finite observation.
pub struct ZeroLagErrorCorrectingExponentialMovingAverage {
    metadata: Metadata,
    configured_length: i64,
    configured_gain_limit: i32,
    configured_alpha: f64,
    complement: f64,
    count: usize,
    ema_state: f64,
    ec_state: f64,
    best_gain_state: f64,
    least_error_state: f64,
    current_value: f64,
    current_ema: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl ZeroLagErrorCorrectingExponentialMovingAverage {
    /// Creates a new ZECEMA instance with Ehlers' default parameters (`length = 20, gain_limit = 50`).
    pub fn default_instance() -> Result<Self, String> {
        Self::from_length(&default_length_params())
    }

    /// Creates a new ZECEMA instance from length-based parameters.
    pub fn new(params: &ZeroLagErrorCorrectingExponentialMovingAverageLengthParams) -> Result<Self, String> {
        Self::from_length(params)
    }

    /// Creates a new ZECEMA instance from length-based parameters.
    pub fn new_from_length(
        params: &ZeroLagErrorCorrectingExponentialMovingAverageLengthParams,
    ) -> Result<Self, String> {
        Self::from_length(params)
    }

    /// Creates a new ZECEMA instance from length-based parameters.
    pub fn from_length(
        params: &ZeroLagErrorCorrectingExponentialMovingAverageLengthParams,
    ) -> Result<Self, String> {
        const INVALID: &str = "invalid zero-lag error-correcting exponential moving average parameters";

        if params.length < 1 {
            return Err(format!("{INVALID}: length must be an integer >= 1"));
        }
        if params.gain_limit < 0 {
            return Err(format!(
                "{INVALID}: gain_limit must be a nonnegative integer in tenths"
            ));
        }

        let alpha_value = 2.0 / ((params.length as f64) + 1.0);
        let complement_value = 1.0 - alpha_value;
        if !alpha_value.is_finite()
            || !(alpha_value > 0.0 && alpha_value <= 1.0)
            || !(complement_value >= 0.0 && complement_value < 1.0)
        {
            return Err(format!(
                "{INVALID}: length must yield representably stable EMA coefficients"
            ));
        }

        Self::new_common(
            params.length,
            params.gain_limit,
            alpha_value,
            complement_value,
            mnemonic(params),
            ema_mnemonic(params),
            params.bar_component,
            params.quote_component,
            params.trade_component,
        )
    }

    /// Creates a new ZECEMA instance from smoothing-factor parameters.
    pub fn new_from_smoothing_factor(
        params: &ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
    ) -> Result<Self, String> {
        Self::from_smoothing_factor(params)
    }

    /// Creates a new ZECEMA instance from smoothing-factor parameters.
    pub fn from_smoothing_factor(
        params: &ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
    ) -> Result<Self, String> {
        const INVALID: &str = "invalid zero-lag error-correcting exponential moving average parameters";

        let alpha_value = params.smoothing_factor;
        if !alpha_value.is_finite() || !(alpha_value > 0.0 && alpha_value <= 1.0) {
            return Err(format!("{INVALID}: smoothing_factor must be in (0, 1]"));
        }
        let complement_value = 1.0 - alpha_value;
        if !(complement_value >= 0.0 && complement_value < 1.0) {
            return Err(format!(
                "{INVALID}: smoothing_factor must yield representably stable EMA coefficients"
            ));
        }
        if params.gain_limit < 0 {
            return Err(format!(
                "{INVALID}: gain_limit must be a nonnegative integer in tenths"
            ));
        }

        Self::new_common(
            equivalent_length_from_alpha(alpha_value),
            params.gain_limit,
            alpha_value,
            complement_value,
            smoothing_factor_mnemonic(params),
            smoothing_factor_ema_mnemonic(params),
            params.bar_component,
            params.quote_component,
            params.trade_component,
        )
    }

    #[allow(clippy::too_many_arguments)]
    fn new_common(
        length_value: i64,
        gain_limit_value: i32,
        alpha_value: f64,
        complement_value: f64,
        mnemonic_text: String,
        mnemonic_ema_text: String,
        bar_comp_opt: Option<BarComponent>,
        quote_comp_opt: Option<QuoteComponent>,
        trade_comp_opt: Option<TradeComponent>,
    ) -> Result<Self, String> {
        let bar_comp = bar_comp_opt.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = quote_comp_opt.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = trade_comp_opt.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let description_text =
            format!("Zero-lag Error-Correcting Exponential Moving Average {mnemonic_text}");
        let description_ema_text = format!(
            "Zero-lag Error-Correcting Exponential Moving Average baseline EMA {mnemonic_ema_text}"
        );

        let metadata = build_metadata(
            Identifier::ZeroLagErrorCorrectingExponentialMovingAverage,
            &mnemonic_text,
            &description_text,
            &[
                OutputText {
                    mnemonic: mnemonic_text.clone(),
                    description: description_text.clone(),
                },
                OutputText {
                    mnemonic: mnemonic_ema_text,
                    description: description_ema_text,
                },
            ],
        );

        Ok(Self {
            metadata,
            configured_length: length_value,
            configured_gain_limit: gain_limit_value,
            configured_alpha: alpha_value,
            complement: complement_value,
            count: 0,
            ema_state: 0.0,
            ec_state: 0.0,
            best_gain_state: 0.0,
            least_error_state: 0.0,
            current_value: f64::NAN,
            current_ema: f64::NAN,
            bar_func: bar_component_value(bar_comp),
            quote_func: quote_component_value(quote_comp),
            trade_func: trade_component_value(trade_comp),
        })
    }

    /// Restores zero startup states and NaN public outputs while preserving configuration.
    pub fn reset(&mut self) {
        self.count = 0;
        self.ema_state = 0.0;
        self.ec_state = 0.0;
        self.best_gain_state = 0.0;
        self.least_error_state = 0.0;
        self.current_value = f64::NAN;
        self.current_ema = f64::NAN;
    }

    /// Returns the configured or equivalent EMA length.
    pub fn length(&self) -> i64 {
        self.configured_length
    }

    /// Returns the configured integer gain limit in tenths.
    pub fn gain_limit(&self) -> i32 {
        self.configured_gain_limit
    }

    /// Returns the baseline EMA smoothing factor alpha.
    pub fn alpha(&self) -> f64 {
        self.configured_alpha
    }

    /// Returns the number of accepted finite observations processed since construction or reset.
    pub fn samples_seen(&self) -> usize {
        self.count
    }

    /// Returns the warmup bar count (`0` for ZECEMA's explicit zero-seeded recursion).
    pub fn warmup_bars(&self) -> usize {
        0
    }

    /// Returns `true` once at least one finite observation has been accepted.
    pub fn is_primed(&self) -> bool {
        self.count > 0
    }

    /// Returns the latest accepted EC value (`NaN` before priming).
    pub fn value(&self) -> f64 {
        self.current_value
    }

    /// Returns the latest accepted baseline EMA value (`NaN` before priming).
    pub fn ema(&self) -> f64 {
        self.current_ema
    }

    /// Returns the selected gain from the latest accepted observation (`NaN` before priming).
    pub fn best_gain(&self) -> f64 {
        if self.count > 0 {
            self.best_gain_state
        } else {
            f64::NAN
        }
    }

    /// Returns the literal least-error diagnostic from the latest accepted observation (`NaN` before priming).
    pub fn least_error(&self) -> f64 {
        if self.count > 0 {
            self.least_error_state
        } else {
            f64::NAN
        }
    }

    fn step(&mut self, sample: f64) -> Result<(f64, f64), String> {
        if sample.is_infinite() {
            return Err("price must be a finite real or NaN".to_string());
        }
        if sample.is_nan() {
            return Ok((f64::NAN, f64::NAN));
        }

        let alpha_value = self.configured_alpha;
        let complement_value = self.complement;
        let previous_ec = self.ec_state;

        let ema_value = alpha_value * sample + complement_value * self.ema_state;
        if !ema_value.is_finite() {
            return Err("baseline EMA arithmetic overflowed".to_string());
        }

        let mut least_error_value = LEAST_ERROR_SENTINEL;
        let mut best_gain_value = self.best_gain_state;
        let limit = i64::from(self.configured_gain_limit);
        for index in -limit..=limit {
            let gain = (index as f64) / 10.0;
            let candidate =
                alpha_value * (ema_value + gain * (sample - previous_ec)) + complement_value * previous_ec;
            let error_value = (sample - candidate).abs();
            if !candidate.is_finite() || !error_value.is_finite() {
                return Err("gain-search arithmetic overflowed".to_string());
            }
            if error_value < least_error_value {
                least_error_value = error_value;
                best_gain_value = gain;
            }
        }

        let ec_value = alpha_value * (ema_value + best_gain_value * (sample - previous_ec))
            + complement_value * previous_ec;
        if !ec_value.is_finite() {
            return Err("selected EC arithmetic overflowed".to_string());
        }

        self.ema_state = ema_value;
        self.ec_state = ec_value;
        self.best_gain_state = best_gain_value;
        self.least_error_state = least_error_value;
        self.count += 1;
        self.current_value = ec_value;
        self.current_ema = ema_value;
        Ok((ec_value, ema_value))
    }

    /// Fallible scalar update. Rejects infinite or overflowing inputs before mutating state and skips `NaN`.
    pub fn try_update(&mut self, sample: f64) -> Result<f64, String> {
        let (ec_value, _) = self.step(sample)?;
        Ok(ec_value)
    }

    /// Updates the indicator with a scalar sample, panicking on infinite input or overflow.
    pub fn update(&mut self, sample: f64) -> f64 {
        self.try_update(sample)
            .expect("ZeroLagErrorCorrectingExponentialMovingAverage::update: invalid sample")
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let (ec_value, ema_value) = self
            .step(sample)
            .expect("ZeroLagErrorCorrectingExponentialMovingAverage::update_entity: invalid sample");
        vec![
            Box::new(Scalar::new(time, ec_value)) as Box<dyn Any>,
            Box::new(Scalar::new(time, ema_value)) as Box<dyn Any>,
        ]
    }
}

impl Indicator for ZeroLagErrorCorrectingExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        ZeroLagErrorCorrectingExponentialMovingAverage::is_primed(self)
    }

    fn metadata(&self) -> Metadata {
        self.metadata.clone()
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let price = (self.bar_func)(sample);
        self.update_entity(sample.time, price)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let price = (self.quote_func)(sample);
        self.update_entity(sample.time, price)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let price = (self.trade_func)(sample);
        self.update_entity(sample.time, price)
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::super::testdata::testdata;
    use super::*;
    use crate::indicators::core::adaptivity::Adaptivity;
    use crate::indicators::core::descriptor::descriptor_of;
    use crate::indicators::core::outputs::shape::Shape;

    type FixtureFn = fn() -> Vec<f64>;

    const REFERENCE_CASES: &[(i64, i32, FixtureFn, FixtureFn)] = &[
        (20, 50, testdata::expected_l20_g50_ec, testdata::expected_l20_g50_ema),
        (12, 50, testdata::expected_l12_g50_ec, testdata::expected_l12_g50_ema),
        (32, 22, testdata::expected_l32_g22_ec, testdata::expected_l32_g22_ema),
        (20, 0, testdata::expected_l20_g0_ec, testdata::expected_l20_g0_ema),
        (20, 1, testdata::expected_l20_g1_ec, testdata::expected_l20_g1_ema),
        (20, 10, testdata::expected_l20_g10_ec, testdata::expected_l20_g10_ema),
        (20, 22, testdata::expected_l20_g22_ec, testdata::expected_l20_g22_ema),
        (20, 100, testdata::expected_l20_g100_ec, testdata::expected_l20_g100_ema),
        (1, 50, testdata::expected_l1_g50_ec, testdata::expected_l1_g50_ema),
        (2, 50, testdata::expected_l2_g50_ec, testdata::expected_l2_g50_ema),
        (5, 50, testdata::expected_l5_g50_ec, testdata::expected_l5_g50_ema),
        (10, 50, testdata::expected_l10_g50_ec, testdata::expected_l10_g50_ema),
        (40, 50, testdata::expected_l40_g50_ec, testdata::expected_l40_g50_ema),
        (100, 50, testdata::expected_l100_g50_ec, testdata::expected_l100_g50_ema),
        (252, 50, testdata::expected_l252_g50_ec, testdata::expected_l252_g50_ema),
        (20, 49, testdata::expected_l20_g49_ec, testdata::expected_l20_g49_ema),
    ];

    fn article_batch(prices: &[f64], length_value: i64, gain_limit_value: i32) -> Vec<(f64, f64, f64, f64)> {
        let alpha_value = 2.0 / ((length_value as f64) + 1.0);
        let mut ema_series = vec![0.0; prices.len() + 1];
        let mut ec_series = vec![0.0; prices.len() + 1];
        let mut best_gain = 0.0;
        let mut outputs = Vec::with_capacity(prices.len());
        let limit = i64::from(gain_limit_value);

        for (offset, &price) in prices.iter().enumerate() {
            let step_index = offset + 1;
            ema_series[step_index] = alpha_value * price + (1.0 - alpha_value) * ema_series[step_index - 1];
            let mut least_error = 1_000_000.0;
            for integer_gain in -limit..=limit {
                let gain = (integer_gain as f64) / 10.0;
                let trial = alpha_value * (ema_series[step_index] + gain * (price - ec_series[step_index - 1]))
                    + (1.0 - alpha_value) * ec_series[step_index - 1];
                let error_value = price - trial;
                if error_value.abs() < least_error {
                    least_error = error_value.abs();
                    best_gain = gain;
                }
            }
            ec_series[step_index] =
                alpha_value * (ema_series[step_index] + best_gain * (price - ec_series[step_index - 1]))
                    + (1.0 - alpha_value) * ec_series[step_index - 1];
            outputs.push((ec_series[step_index], ema_series[step_index], best_gain, least_error));
        }
        outputs
    }

    fn assert_close_or_nan(actual: f64, expected: f64) {
        if expected.is_nan() {
            assert!(actual.is_nan(), "expected NaN, got {actual}");
        } else {
            let scale = 1.0_f64.max(expected.abs());
            assert!(
                (actual - expected).abs() <= 2e-12 * scale,
                "expected {expected}, got {actual}"
            );
        }
    }

    #[test]
    fn test_all_reference_fixtures_and_article_batch() {
        let prices = testdata::test_input();
        for &(length_value, gain_limit_value, ec_fn, ema_fn) in REFERENCE_CASES {
            let expected_ec = ec_fn();
            let expected_ema = ema_fn();
            let reference = article_batch(&prices, length_value, gain_limit_value);

            let params = ZeroLagErrorCorrectingExponentialMovingAverageLengthParams {
                length: length_value,
                gain_limit: gain_limit_value,
                ..Default::default()
            };
            let mut stream = ZeroLagErrorCorrectingExponentialMovingAverage::from_length(&params).unwrap();
            let mut entity_stream =
                ZeroLagErrorCorrectingExponentialMovingAverage::from_length(&params).unwrap();

            for (bar_index, &price) in prices.iter().enumerate() {
                let ec_value = stream.update(price);
                let ema_value = stream.ema();
                assert_close_or_nan(ec_value, expected_ec[bar_index]);
                assert_close_or_nan(ema_value, expected_ema[bar_index]);
                assert_close_or_nan(ec_value, reference[bar_index].0);
                assert_close_or_nan(ema_value, reference[bar_index].1);
                assert_eq!(stream.best_gain(), reference[bar_index].2);
                assert_eq!(stream.least_error(), reference[bar_index].3);

                let stamp = 1_600_000_000_i64 + (bar_index as i64) * 86_400;
                let entity_out = entity_stream.update_scalar(&Scalar::new(stamp, price));
                assert_eq!(entity_out.len(), 2);
                let first_scalar = entity_out[0].downcast_ref::<Scalar>().unwrap();
                let second_scalar = entity_out[1].downcast_ref::<Scalar>().unwrap();
                assert_eq!(first_scalar.time, stamp);
                assert_eq!(second_scalar.time, stamp);
                assert_eq!(first_scalar.value, ec_value);
                assert_eq!(second_scalar.value, ema_value);
            }
        }
    }

    #[test]
    fn test_zero_state_startup_priming_and_reset() {
        let mut stream = ZeroLagErrorCorrectingExponentialMovingAverage::default_instance().unwrap();
        assert!(!stream.is_primed());
        assert_eq!(stream.warmup_bars(), 0);
        assert_eq!(stream.samples_seen(), 0);
        assert!(stream.value().is_nan());
        assert!(stream.ema().is_nan());
        assert!(stream.best_gain().is_nan());
        assert!(stream.least_error().is_nan());
        assert_eq!(stream.alpha(), 2.0 / 21.0);

        let ec_first = stream.update(100.0);
        let alpha_value = 2.0 / 21.0;
        let expected_ema = alpha_value * 100.0;
        let expected_ec = alpha_value * (expected_ema + 5.0 * 100.0);
        assert!(stream.is_primed());
        assert_eq!(stream.samples_seen(), 1);
        assert_eq!(ec_first, expected_ec);
        assert_eq!(stream.ema(), expected_ema);
        assert_eq!(stream.best_gain(), 5.0);
        assert_eq!(stream.least_error(), (100.0 - expected_ec).abs());

        stream.reset();
        assert!(!stream.is_primed());
        assert_eq!(stream.samples_seen(), 0);
        assert!(stream.value().is_nan());
        assert!(stream.ema().is_nan());
        assert!(stream.best_gain().is_nan());
        assert!(stream.least_error().is_nan());
        assert_eq!(stream.update(100.0), expected_ec);
    }

    #[test]
    fn test_zero_gain_cascaded_ema_ties_sentinel_nan_and_overflow() {
        let prices = testdata::test_input();
        let mut zero_gain = ZeroLagErrorCorrectingExponentialMovingAverage::from_length(&LengthParams {
            length: 20,
            gain_limit: 0,
            ..Default::default()
        })
        .unwrap();
        let alpha_value = 2.0 / 21.0;
        let mut ema_acc = 0.0;
        let mut cascaded_acc = 0.0;
        for &price in &prices {
            ema_acc = alpha_value * price + (1.0 - alpha_value) * ema_acc;
            cascaded_acc = alpha_value * ema_acc + (1.0 - alpha_value) * cascaded_acc;
            assert_eq!(zero_gain.update(price), cascaded_acc);
            assert_eq!(zero_gain.ema(), ema_acc);
            assert_eq!(zero_gain.best_gain(), 0.0);
        }

        let mut tie_stream = ZeroLagErrorCorrectingExponentialMovingAverage::default_instance().unwrap();
        assert_eq!(tie_stream.update(0.0), 0.0);
        assert_eq!(tie_stream.best_gain(), -5.0);
        assert_eq!(tie_stream.least_error(), 0.0);

        let mut huge_stream = ZeroLagErrorCorrectingExponentialMovingAverage::default_instance().unwrap();
        let huge_price = 1e9;
        let huge_ec = huge_stream.update(huge_price);
        assert_eq!(huge_stream.best_gain(), 0.0);
        assert_eq!(huge_stream.least_error(), 1_000_000.0);
        assert_eq!(huge_ec, alpha_value * (alpha_value * huge_price));

        let mut nan_stream = ZeroLagErrorCorrectingExponentialMovingAverage::default_instance().unwrap();
        let pre_nan = nan_stream.update_scalar(&Scalar::new(10, f64::NAN));
        assert!(pre_nan[0].downcast_ref::<Scalar>().unwrap().value.is_nan());
        assert!(pre_nan[1].downcast_ref::<Scalar>().unwrap().value.is_nan());
        assert!(!nan_stream.is_primed());

        let accepted_ec = nan_stream.update(prices[0]);
        let accepted_ema = nan_stream.ema();
        let accepted_gain = nan_stream.best_gain();
        let accepted_error = nan_stream.least_error();

        assert!(nan_stream.update(f64::NAN).is_nan());
        assert_eq!(nan_stream.samples_seen(), 1);
        assert_eq!(nan_stream.value(), accepted_ec);
        assert_eq!(nan_stream.ema(), accepted_ema);
        assert_eq!(nan_stream.best_gain(), accepted_gain);
        assert_eq!(nan_stream.least_error(), accepted_error);

        for bad_sample in [f64::INFINITY, f64::NEG_INFINITY, 1e308] {
            assert!(nan_stream.try_update(bad_sample).is_err());
            assert_eq!(nan_stream.samples_seen(), 1);
            assert_eq!(nan_stream.value(), accepted_ec);
            assert_eq!(nan_stream.ema(), accepted_ema);
            assert_eq!(nan_stream.best_gain(), accepted_gain);
            assert_eq!(nan_stream.least_error(), accepted_error);
        }
    }

    #[test]
    fn test_smoothing_factor_metadata_entities_and_validation() {
        let prices = testdata::test_input();
        let mut from_length = ZeroLagErrorCorrectingExponentialMovingAverage::from_length(&LengthParams {
            length: 20,
            gain_limit: 50,
            ..Default::default()
        })
        .unwrap();
        let mut from_alpha =
            ZeroLagErrorCorrectingExponentialMovingAverage::from_smoothing_factor(&SmoothingFactorParams {
                smoothing_factor: 2.0 / 21.0,
                gain_limit: 50,
                ..Default::default()
            })
            .unwrap();
        assert_eq!(from_alpha.length(), 20);
        for &price in &prices {
            assert_eq!(from_alpha.update(price), from_length.update(price));
            assert_eq!(from_alpha.ema(), from_length.ema());
        }

        let mut unit_alpha =
            ZeroLagErrorCorrectingExponentialMovingAverage::from_smoothing_factor(&SmoothingFactorParams {
                smoothing_factor: 1.0,
                gain_limit: 50,
                ..Default::default()
            })
            .unwrap();
        assert_eq!(unit_alpha.length(), 1);
        assert_eq!(unit_alpha.update(91.5), 91.5);
        assert_eq!(equivalent_length_from_alpha(1e-9), i64::MAX);

        let md = from_length.metadata();
        let desc = descriptor_of(Identifier::ZeroLagErrorCorrectingExponentialMovingAverage).unwrap();
        assert_eq!(desc.adaptivity, Adaptivity::Adaptive);
        assert_eq!(md.identifier, Identifier::ZeroLagErrorCorrectingExponentialMovingAverage);
        assert_eq!(md.mnemonic, "zecema(20, 50)");
        assert_eq!(md.outputs.len(), 2);
        assert_eq!(
            md.outputs[0].kind,
            ZeroLagErrorCorrectingExponentialMovingAverageOutput::Value as i32
        );
        assert_eq!(md.outputs[0].shape, Shape::Scalar);
        assert_eq!(md.outputs[0].mnemonic, "zecema(20, 50)");
        assert_eq!(
            md.outputs[1].kind,
            ZeroLagErrorCorrectingExponentialMovingAverageOutput::Ema as i32
        );
        assert_eq!(md.outputs[1].shape, Shape::Scalar);
        assert_eq!(md.outputs[1].mnemonic, "zecemaEma(20, 50)");

        let mut bar_stream = ZeroLagErrorCorrectingExponentialMovingAverage::default_instance().unwrap();
        let bar_out = bar_stream.update_bar(&Bar::new(100, 100.0, 101.0, 99.0, 100.0, 10.0));
        assert_eq!(bar_out.len(), 2);
        assert!(bar_out[0].downcast_ref::<Scalar>().unwrap().value.is_finite());
        assert!(bar_out[1].downcast_ref::<Scalar>().unwrap().value.is_finite());

        let mut quote_stream = ZeroLagErrorCorrectingExponentialMovingAverage::default_instance().unwrap();
        let quote_out = quote_stream.update_quote(&Quote::new(100, 100.0, 100.0, 1.0, 1.0));
        assert_eq!(quote_out.len(), 2);

        let mut trade_stream = ZeroLagErrorCorrectingExponentialMovingAverage::default_instance().unwrap();
        let trade_out = trade_stream.update_trade(&Trade::new(100, 100.0, 1.0));
        assert_eq!(trade_out.len(), 2);

        assert_eq!(
            mnemonic(&LengthParams {
                length: 32,
                gain_limit: 22,
                bar_component: Some(BarComponent::Open),
                quote_component: Some(QuoteComponent::Ask),
                trade_component: None,
            }),
            "zecema(32, 22, o, a)"
        );
        assert_eq!(
            smoothing_factor_ema_mnemonic(&SmoothingFactorParams {
                smoothing_factor: 0.2,
                gain_limit: 22,
                bar_component: Some(BarComponent::Typical),
                quote_component: None,
                trade_component: Some(TradeComponent::Volume),
            }),
            "zecemaEma(9, 0.20000000, 22, hlc/3, v)"
        );

        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::from_length(&LengthParams {
            length: 0,
            ..Default::default()
        })
        .is_err());
        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::from_length(&LengthParams {
            length: i64::MAX,
            ..Default::default()
        })
        .is_err());
        assert!(ZeroLagErrorCorrectingExponentialMovingAverage::from_length(&LengthParams {
            gain_limit: -1,
            ..Default::default()
        })
        .is_err());
        for bad_alpha in [0.0, -0.1, 1.1, 1e-30, f64::NAN, f64::INFINITY] {
            assert!(
                ZeroLagErrorCorrectingExponentialMovingAverage::from_smoothing_factor(&SmoothingFactorParams {
                    smoothing_factor: bad_alpha,
                    ..Default::default()
                })
                .is_err()
            );
        }
        assert!(
            ZeroLagErrorCorrectingExponentialMovingAverage::from_smoothing_factor(&SmoothingFactorParams {
                gain_limit: -1,
                ..Default::default()
            })
            .is_err()
        );
    }
}
