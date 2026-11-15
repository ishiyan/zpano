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

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Output indices for Ehlers' Center of Gravity oscillator.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CenterOfGravityOscillatorOutput {
    /// The scalar value of the Center of Gravity oscillator.
    Value = 1,
    /// The one-observation delayed trigger line.
    Trigger = 2,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create a Center of Gravity oscillator.
#[derive(Debug, Clone, Copy)]
pub struct CenterOfGravityOscillatorParams {
    /// Observation window length (>= 1). Default is 10.
    pub length: i64,
    /// Zero-centered 2004 book baseline (-Num/Denom + (length+1)/2) when true; May 2002 article baseline (-Num/Denom) when false.
    pub centered: bool,
    /// Bar component to extract. `None` means default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means default (Price).
    pub trade_component: Option<TradeComponent>,
}

pub type Params = CenterOfGravityOscillatorParams;

impl Default for CenterOfGravityOscillatorParams {
    fn default() -> Self {
        Self {
            length: 10,
            centered: false,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Returns default parameters for the Center of Gravity oscillator.
pub fn default_params() -> CenterOfGravityOscillatorParams {
    CenterOfGravityOscillatorParams::default()
}

fn format_mnemonic(prefix: &str, params: &CenterOfGravityOscillatorParams) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let centered_tag = if params.centered { ", centered" } else { "" };
    let suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    format!("{prefix}({}{centered_tag}{suffix})", params.length)
}

/// Builds the primary CenterOfGravityOscillator mnemonic string.
pub fn mnemonic(params: &CenterOfGravityOscillatorParams) -> String {
    format_mnemonic("cog", params)
}

/// Builds the CenterOfGravityOscillator trigger mnemonic string.
pub fn trigger_mnemonic(params: &CenterOfGravityOscillatorParams) -> String {
    format_mnemonic("cogTrig", params)
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Center of Gravity oscillator (COG).
///
/// Supports both the May 2002 TASC / Mesa convention (`centered = false`,
/// `-Num / Denom`) and the 2004 book convention (`centered = true`,
/// `-Num / Denom + (length + 1) / 2`), exposing the oscillator value on
/// observation `length` and its one-observation delayed trigger on
/// observation `length + 1`.
pub struct CenterOfGravityOscillator {
    metadata: Metadata,
    window_length: usize,
    is_centered: bool,
    offset: f64,
    window: Vec<f64>,
    next: usize,
    count: usize,
    last_value: f64,
    current_value: f64,
    current_trigger: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CenterOfGravityOscillator {
    /// Creates a new Center of Gravity oscillator with Ehlers' default parameters.
    pub fn default_instance() -> Result<Self, String> {
        Self::new(&CenterOfGravityOscillatorParams::default())
    }

    /// Creates a new Center of Gravity oscillator from the supplied parameters.
    pub fn new(params: &CenterOfGravityOscillatorParams) -> Result<Self, String> {
        const INVALID: &str =
            "invalid center of gravity oscillator parameters: length should be a positive integer";

        if params.length < 1 {
            return Err(INVALID.to_string());
        }
        let window_length = usize::try_from(params.length).map_err(|_| INVALID.to_string())?;

        let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bar_comp);
        let quote_func = quote_component_value(quote_comp);
        let trade_func = trade_component_value(trade_comp);

        let mnemonic_text = mnemonic(params);
        let mnemonic_trig = trigger_mnemonic(params);
        let description_text = format!("Center of Gravity oscillator {mnemonic_text}");
        let description_trig = format!("Center of Gravity trigger {mnemonic_trig}");

        let metadata = build_metadata(
            Identifier::CenterOfGravityOscillator,
            &mnemonic_text,
            &description_text,
            &[
                OutputText {
                    mnemonic: mnemonic_text.clone(),
                    description: description_text.clone(),
                },
                OutputText {
                    mnemonic: mnemonic_trig,
                    description: description_trig,
                },
            ],
        );

        let offset = if params.centered {
            (window_length as f64 + 1.0) / 2.0
        } else {
            0.0
        };

        Ok(Self {
            metadata,
            window_length,
            is_centered: params.centered,
            offset,
            window: vec![0.0; window_length],
            next: 0,
            count: 0,
            last_value: 0.0,
            current_value: f64::NAN,
            current_trigger: f64::NAN,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Clears streaming history and priming state while preserving configuration.
    pub fn reset(&mut self) {
        self.window.fill(0.0);
        self.next = 0;
        self.count = 0;
        self.last_value = 0.0;
        self.current_value = f64::NAN;
        self.current_trigger = f64::NAN;
    }

    /// Returns the configured observation window length.
    pub fn length(&self) -> usize {
        self.window_length
    }

    /// Returns whether the zero-centered 2004 book baseline is enabled.
    pub fn centered(&self) -> bool {
        self.is_centered
    }

    /// Returns the latest oscillator value (`NaN` before priming).
    pub fn value(&self) -> f64 {
        self.current_value
    }

    /// Returns the latest one-observation delayed trigger (`NaN` before trigger priming).
    pub fn trigger(&self) -> f64 {
        self.current_trigger
    }

    /// Returns `true` once at least `length` finite observations have been processed.
    pub fn is_primed(&self) -> bool {
        self.count >= self.window_length
    }

    /// Returns `true` once at least `length + 1` finite observations have been processed.
    pub fn is_trigger_primed(&self) -> bool {
        self.count > self.window_length
    }

    /// Fallible scalar update. Rejects infinite inputs without mutating state and skips `NaN`.
    pub fn try_update(&mut self, sample: f64) -> Result<f64, String> {
        if sample.is_infinite() {
            return Err("sample must be finite or NaN".to_string());
        }
        if sample.is_nan() {
            return Ok(f64::NAN);
        }

        let next_trigger = if self.count >= self.window_length {
            self.last_value
        } else {
            f64::NAN
        };

        self.window[self.next] = sample;
        self.next = (self.next + 1) % self.window_length;
        if self.count <= self.window_length {
            self.count += 1;
        }
        if self.count < self.window_length {
            return Ok(f64::NAN);
        }

        let mut numerator = 0.0;
        let mut denominator = 0.0;
        for lag in 0..self.window_length {
            let index = (self.next + self.window_length - 1 - lag) % self.window_length;
            let item = self.window[index];
            numerator += ((1 + lag) as f64) * item;
            denominator += item;
        }

        if denominator != 0.0 {
            self.last_value = -numerator / denominator + self.offset;
        }
        self.current_value = self.last_value;
        self.current_trigger = next_trigger;
        Ok(self.current_value)
    }

    /// Updates the indicator with a scalar sample, panicking on infinite input.
    pub fn update(&mut self, sample: f64) -> f64 {
        self.try_update(sample)
            .expect("CenterOfGravityOscillator::update: sample must be finite or NaN")
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let cog = self.update(sample);
        let trig = if cog.is_nan() {
            f64::NAN
        } else {
            self.current_trigger
        };
        vec![
            Box::new(Scalar::new(time, cog)) as Box<dyn Any>,
            Box::new(Scalar::new(time, trig)) as Box<dyn Any>,
        ]
    }
}

impl Indicator for CenterOfGravityOscillator {
    fn is_primed(&self) -> bool {
        self.count >= self.window_length
    }

    fn metadata(&self) -> Metadata {
        self.metadata.clone()
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let extracted = (self.bar_func)(sample);
        self.update_entity(sample.time, extracted)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let extracted = (self.quote_func)(sample);
        self.update_entity(sample.time, extracted)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let extracted = (self.trade_func)(sample);
        self.update_entity(sample.time, extracted)
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::core::outputs::shape::Shape;
    use crate::indicators::factory::create_indicator;
    use crate::indicators::john_ehlers::center_of_gravity_oscillator::testdata::testdata as fixture;

    fn assert_number(actual: f64, expected: f64, tolerance: f64, context: &str) {
        if expected.is_nan() {
            assert!(actual.is_nan(), "{context}: expected NaN, got {actual}");
        } else {
            assert!(!actual.is_nan(), "{context}: expected {expected}, got NaN");
            assert!(
                (actual - expected).abs() <= tolerance,
                "{context}: expected {expected}, got {actual} (diff {})",
                (actual - expected).abs()
            );
        }
    }

    fn source_outputs(prices: &[f64], length: usize, centered: bool) -> Vec<(f64, f64)> {
        let mut previous = 0.0;
        let mut results = Vec::with_capacity(prices.len());
        for step in 0..prices.len() {
            if step + 1 < length {
                results.push((f64::NAN, f64::NAN));
                continue;
            }
            let trigger = if step >= length { previous } else { f64::NAN };
            let mut numerator = 0.0;
            let mut denominator = 0.0;
            for lag in 0..length {
                numerator += ((1 + lag) as f64) * prices[step - lag];
                denominator += prices[step - lag];
            }
            if denominator != 0.0 {
                previous = -numerator / denominator;
                if centered {
                    previous += (length as f64 + 1.0) / 2.0;
                }
            }
            results.push((previous, trigger));
        }
        results
    }

    fn reference_cases() -> Vec<(i64, bool, Vec<f64>, Vec<f64>)> {
        vec![
            (1, false, fixture::expected_l1_c0_cg(), fixture::expected_l1_c0_tr()),
            (1, true, fixture::expected_l1_c1_cg(), fixture::expected_l1_c1_tr()),
            (2, false, fixture::expected_l2_c0_cg(), fixture::expected_l2_c0_tr()),
            (2, true, fixture::expected_l2_c1_cg(), fixture::expected_l2_c1_tr()),
            (3, false, fixture::expected_l3_c0_cg(), fixture::expected_l3_c0_tr()),
            (3, true, fixture::expected_l3_c1_cg(), fixture::expected_l3_c1_tr()),
            (5, false, fixture::expected_l5_c0_cg(), fixture::expected_l5_c0_tr()),
            (5, true, fixture::expected_l5_c1_cg(), fixture::expected_l5_c1_tr()),
            (8, false, fixture::expected_l8_c0_cg(), fixture::expected_l8_c0_tr()),
            (8, true, fixture::expected_l8_c1_cg(), fixture::expected_l8_c1_tr()),
            (10, false, fixture::expected_l10_c0_cg(), fixture::expected_l10_c0_tr()),
            (10, true, fixture::expected_l10_c1_cg(), fixture::expected_l10_c1_tr()),
            (14, false, fixture::expected_l14_c0_cg(), fixture::expected_l14_c0_tr()),
            (14, true, fixture::expected_l14_c1_cg(), fixture::expected_l14_c1_tr()),
            (20, false, fixture::expected_l20_c0_cg(), fixture::expected_l20_c0_tr()),
            (20, true, fixture::expected_l20_c1_cg(), fixture::expected_l20_c1_tr()),
            (32, false, fixture::expected_l32_c0_cg(), fixture::expected_l32_c0_tr()),
            (32, true, fixture::expected_l32_c1_cg(), fixture::expected_l32_c1_tr()),
            (64, false, fixture::expected_l64_c0_cg(), fixture::expected_l64_c0_tr()),
            (64, true, fixture::expected_l64_c1_cg(), fixture::expected_l64_c1_tr()),
            (126, false, fixture::expected_l126_c0_cg(), fixture::expected_l126_c0_tr()),
            (126, true, fixture::expected_l126_c1_cg(), fixture::expected_l126_c1_tr()),
            (252, false, fixture::expected_l252_c0_cg(), fixture::expected_l252_c0_tr()),
            (252, true, fixture::expected_l252_c1_cg(), fixture::expected_l252_c1_tr()),
        ]
    }

    #[test]
    fn test_all_twenty_four_reference_cases_and_published_loop() {
        let close = fixture::test_input();
        let open = fixture::test_open();
        let high = fixture::test_high();
        let low = fixture::test_low();
        assert_eq!(close.len(), 252);

        for (length, centered, expected_values, expected_triggers) in reference_cases() {
            let params = CenterOfGravityOscillatorParams {
                length,
                centered,
                ..Default::default()
            };
            let label = mnemonic(&params);
            let mut indicator = CenterOfGravityOscillator::new(&params).unwrap();
            let mut bar_indicator = CenterOfGravityOscillator::new(&params).unwrap();
            let oracle = source_outputs(&close, length as usize, centered);

            assert_eq!(indicator.length(), length as usize);
            assert_eq!(indicator.centered(), centered);
            assert!(!indicator.is_primed());
            assert!(!indicator.is_trigger_primed());
            assert!(indicator.value().is_nan());
            assert!(indicator.trigger().is_nan());

            for step in 0..close.len() {
                let actual_value = indicator.update(close[step]);
                let actual_trigger = indicator.trigger();
                let ctx = format!("{label} bar {step}");
                assert_number(actual_value, expected_values[step], 1e-13, &ctx);
                assert_number(actual_trigger, expected_triggers[step], 1e-13, &ctx);
                assert_number(actual_value, oracle[step].0, 1e-14, &ctx);
                assert_number(actual_trigger, oracle[step].1, 1e-14, &ctx);
                assert_number(indicator.value(), expected_values[step], 1e-13, &ctx);
                assert_eq!(indicator.is_primed(), step + 1 >= length as usize);
                assert_eq!(indicator.is_trigger_primed(), step >= length as usize);

                let bar = Bar::new(
                    1_600_000_000 + step as i64,
                    open[step],
                    high[step],
                    low[step],
                    close[step],
                    1000.0,
                );
                let bar_out = bar_indicator.update_bar(&bar);
                assert_eq!(bar_out.len(), 2);
                let first_scalar = bar_out[0].downcast_ref::<Scalar>().unwrap();
                let second_scalar = bar_out[1].downcast_ref::<Scalar>().unwrap();
                assert_number(first_scalar.value, expected_values[step], 1e-13, &ctx);
                assert_number(second_scalar.value, expected_triggers[step], 1e-13, &ctx);
            }
        }
    }

    #[test]
    fn test_newest_price_has_weight_one_and_sign_is_negative() {
        let mut article = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 3,
            ..Default::default()
        })
        .unwrap();
        let mut book = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 3,
            centered: true,
            ..Default::default()
        })
        .unwrap();

        let mut article_value = f64::NAN;
        let mut book_value = f64::NAN;
        for price in [10.0, 20.0, 30.0] {
            article_value = article.update(price);
            book_value = book.update(price);
        }
        assert_number(article_value, -5.0 / 3.0, 1e-15, "article L=3");
        assert_number(book_value, 1.0 / 3.0, 1e-15, "book L=3");
        assert!(article.trigger().is_nan());
        assert_number(article.update(60.0), -18.0 / 11.0, 1e-15, "article step 4");
        assert_number(article.trigger(), -5.0 / 3.0, 1e-15, "article trigger step 4");
    }

    #[test]
    fn test_constant_price_length_one_scale_invariance_and_crossover() {
        for length in [1_i64, 2, 3, 10] {
            for centered in [false, true] {
                let mut indicator =
                    CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
                        length,
                        centered,
                        ..Default::default()
                    })
                    .unwrap();
                let baseline = if centered {
                    0.0
                } else {
                    -(length as f64 + 1.0) / 2.0
                };
                for step in 0..(length as usize + 2) {
                    let actual_value = indicator.update(8.0);
                    let expected_value = if step + 1 < length as usize {
                        f64::NAN
                    } else {
                        baseline
                    };
                    let expected_trigger = if step < length as usize {
                        f64::NAN
                    } else {
                        baseline
                    };
                    assert_number(actual_value, expected_value, 1e-13, "constant value");
                    assert_number(indicator.trigger(), expected_trigger, 1e-13, "constant trig");
                }
            }
        }

        let close = fixture::test_input();
        let mut base = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 5,
            centered: true,
            ..Default::default()
        })
        .unwrap();
        let mut scaled = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 5,
            centered: true,
            ..Default::default()
        })
        .unwrap();
        for price in close.iter().take(25) {
            assert_number(
                base.update(*price),
                scaled.update(3.75 * *price),
                1e-13,
                "scale invariance value",
            );
            assert_number(
                base.trigger(),
                scaled.trigger(),
                1e-13,
                "scale invariance trigger",
            );
        }

        let mut rising = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 4,
            centered: true,
            ..Default::default()
        })
        .unwrap();
        let mut falling = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 4,
            centered: true,
            ..Default::default()
        })
        .unwrap();
        for step in 0..4 {
            rising.update((10 + step) as f64);
            falling.update((13 - step) as f64);
        }
        assert!(rising.value() > 0.0);
        assert!(falling.value() < 0.0);
        let after_drop = rising.update(1.0);
        assert!(after_drop < rising.trigger());
    }

    #[test]
    fn test_zero_denominator_and_subnormal_guards() {
        for centered in [false, true] {
            let mut initial = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
                length: 2,
                centered,
                ..Default::default()
            })
            .unwrap();
            initial.update(1.0);
            assert_eq!(initial.update(-1.0), 0.0);
            assert!(initial.trigger().is_nan());

            let mut subsequent =
                CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
                    length: 2,
                    centered,
                    ..Default::default()
                })
                .unwrap();
            subsequent.update(1.0);
            let previous = subsequent.update(2.0);
            let held = subsequent.update(-2.0);
            assert_eq!(held, previous);
            assert_eq!(subsequent.trigger(), previous);
            let resumed = subsequent.update(4.0);
            assert_eq!(subsequent.trigger(), previous);
            assert_eq!(resumed, if centered { 1.5 } else { 0.0 });
        }

        let mut subnormal = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 2,
            ..Default::default()
        })
        .unwrap();
        subnormal.update(1e-309);
        assert_number(subnormal.update(3e-309), -1.25, 1e-14, "subnormal ratio");
    }

    #[test]
    fn test_nan_skip_infinity_rejection_and_reset() {
        let mut skipped = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 3,
            ..Default::default()
        })
        .unwrap();
        let mut uninterrupted = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 3,
            ..Default::default()
        })
        .unwrap();

        let stamp: i64 = 1_700_000_000;
        for price in [1.0, 2.0, 3.0, 4.0, 5.0] {
            assert!(skipped.update(f64::NAN).is_nan());
            let gap = skipped.update_scalar(&Scalar::new(stamp, f64::NAN));
            let gap_first = gap[0].downcast_ref::<Scalar>().unwrap();
            let gap_second = gap[1].downcast_ref::<Scalar>().unwrap();
            assert!(gap_first.value.is_nan() && gap_second.value.is_nan());
            let actual = skipped.update(price);
            let reference = uninterrupted.update(price);
            assert_number(actual, reference, 0.0, "NaN skip value");
            assert_number(skipped.trigger(), uninterrupted.trigger(), 0.0, "NaN skip trigger");
        }
        assert!(skipped.is_trigger_primed());

        let mut single = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 1,
            ..Default::default()
        })
        .unwrap();
        single.update(4.0);
        assert!(single.try_update(f64::INFINITY).is_err());
        assert!(single.try_update(f64::NEG_INFINITY).is_err());
        assert_eq!(single.update(5.0), -1.0);
        assert_eq!(single.trigger(), -1.0);

        let mut resettable = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 2,
            centered: true,
            ..Default::default()
        })
        .unwrap();
        for price in [1.0, 2.0, 3.0] {
            resettable.update(price);
        }
        resettable.reset();
        assert!(!resettable.is_primed());
        assert!(!resettable.is_trigger_primed());
        assert!(resettable.value().is_nan());
        assert!(resettable.trigger().is_nan());
        assert!(resettable.update(1.0).is_nan());
        assert_eq!(resettable.update(-1.0), 0.0);
    }

    #[test]
    #[should_panic(expected = "sample must be finite or NaN")]
    fn test_update_panics_on_infinity() {
        let mut indicator = CenterOfGravityOscillator::default_instance().unwrap();
        indicator.update(f64::INFINITY);
    }

    #[test]
    fn test_validation_metadata_entities_and_factory() {
        for bad_length in [0_i64, -1, -100] {
            assert!(
                CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
                    length: bad_length,
                    ..Default::default()
                })
                .is_err()
            );
        }

        let default_ind = CenterOfGravityOscillator::default_instance().unwrap();
        let meta = default_ind.metadata();
        assert_eq!(meta.identifier, Identifier::CenterOfGravityOscillator);
        assert_eq!(meta.mnemonic, "cog(10)");
        assert_eq!(meta.description, "Center of Gravity oscillator cog(10)");
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(
            meta.outputs[0].kind,
            CenterOfGravityOscillatorOutput::Value as i32
        );
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
        assert_eq!(meta.outputs[0].mnemonic, "cog(10)");
        assert_eq!(
            meta.outputs[1].kind,
            CenterOfGravityOscillatorOutput::Trigger as i32
        );
        assert_eq!(meta.outputs[1].shape, Shape::Scalar);
        assert_eq!(meta.outputs[1].mnemonic, "cogTrig(10)");
        assert_eq!(
            meta.outputs[1].description,
            "Center of Gravity trigger cogTrig(10)"
        );

        let custom_params = CenterOfGravityOscillatorParams {
            length: 14,
            centered: true,
            bar_component: Some(BarComponent::Median),
            quote_component: Some(QuoteComponent::Bid),
            trade_component: Some(TradeComponent::Volume),
        };
        assert_eq!(mnemonic(&custom_params), "cog(14, centered, hl/2, b, v)");
        assert_eq!(
            trigger_mnemonic(&custom_params),
            "cogTrig(14, centered, hl/2, b, v)"
        );

        let from_factory = create_indicator(
            Identifier::CenterOfGravityOscillator,
            r#"{"length": 8, "centered": true, "barComponent": 5}"#,
        )
        .unwrap();
        assert_eq!(from_factory.metadata().mnemonic, "cog(8, centered, hl/2)");
        assert!(
            create_indicator(Identifier::CenterOfGravityOscillator, r#"{"length": 1.5}"#).is_err()
        );
        assert!(
            create_indicator(Identifier::CenterOfGravityOscillator, r#"{"centered": 1}"#).is_err()
        );
        assert!(
            create_indicator(Identifier::CenterOfGravityOscillator, r#"{"length": 0}"#).is_err()
        );

        let open = fixture::test_open();
        let high = fixture::test_high();
        let low = fixture::test_low();
        let close = fixture::test_input();

        let mut bar_median = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 4,
            centered: true,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        })
        .unwrap();
        let mut quote_mid = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 4,
            centered: true,
            ..Default::default()
        })
        .unwrap();
        let mut trade_price = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 4,
            centered: true,
            ..Default::default()
        })
        .unwrap();
        let mut scalar_direct = CenterOfGravityOscillator::new(&CenterOfGravityOscillatorParams {
            length: 4,
            centered: true,
            ..Default::default()
        })
        .unwrap();

        for step in 0..12 {
            let stamp = 1_700_000_000 + step as i64;
            let median = (high[step] + low[step]) / 2.0;
            let bar_out = bar_median.update_bar(&Bar::new(
                stamp,
                open[step],
                high[step],
                low[step],
                close[step],
                1000.0,
            ));
            let quote_out = quote_mid.update_quote(&Quote::new(
                stamp, low[step], high[step], 1.0, 1.0,
            ));
            let trade_out = trade_price.update_trade(&Trade::new(stamp, median, 10.0));
            let scalar_out = scalar_direct.update_scalar(&Scalar::new(stamp, median));
            for slot in 0..2 {
                let bar_scalar = bar_out[slot].downcast_ref::<Scalar>().unwrap().value;
                let quote_scalar = quote_out[slot].downcast_ref::<Scalar>().unwrap().value;
                let trade_scalar = trade_out[slot].downcast_ref::<Scalar>().unwrap().value;
                let direct_scalar = scalar_out[slot].downcast_ref::<Scalar>().unwrap().value;
                assert_number(bar_scalar, direct_scalar, 1e-15, "bar vs scalar");
                assert_number(quote_scalar, direct_scalar, 1e-15, "quote vs scalar");
                assert_number(trade_scalar, direct_scalar, 1e-15, "trade vs scalar");
            }
        }
    }
}
