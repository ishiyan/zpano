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

#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RoofingFilterOutput {
    Value = 1,
}

#[derive(Debug, Clone, Copy)]
pub struct RoofingFilterParams {
    pub shortest_cycle_period: f64,
    pub longest_cycle_period: f64,
    pub highpass_poles: i32,
    pub zero_mean: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

pub type Params = RoofingFilterParams;

impl Default for RoofingFilterParams {
    fn default() -> Self {
        Self {
            shortest_cycle_period: 10.0,
            longest_cycle_period: 48.0,
            highpass_poles: 2,
            zero_mean: false,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

pub fn default_params() -> RoofingFilterParams {
    RoofingFilterParams::default()
}

fn format_number(number: f64) -> String {
    let normalized = if number == 0.0 { 0.0 } else { number };
    format!("{normalized}")
}

fn smoother_coefficients(period: f64) -> (f64, f64, f64) {
    let smoother_angle = SQRT_2 * PI / period;
    let a1 = (-smoother_angle).exp();
    let c2 = 2.0 * a1 * smoother_angle.cos();
    let c3 = -a1 * a1;
    (1.0 - c2 - c3, c2, c3)
}

pub fn mnemonic(params: &RoofingFilterParams) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    let zm_tag = if params.zero_mean { "zm" } else { "" };
    let shortest_text = format_number(params.shortest_cycle_period);
    let longest_text = format_number(params.longest_cycle_period);
    format!(
        "roof{}hp{}({}, {}{})",
        params.highpass_poles, zm_tag, shortest_text, longest_text, suffix
    )
}

pub struct RoofingFilter {
    metadata: Metadata,
    poles: usize,
    zero_mean: bool,
    warmup_count: usize,
    gain: f64,
    decay: f64,
    c1: f64,
    c2: f64,
    c3: f64,
    count: usize,
    input1: f64,
    input2: f64,
    hp1: f64,
    hp2: f64,
    filt1: f64,
    filt2: f64,
    zero1: f64,
    current_value: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl RoofingFilter {
    pub fn default_instance() -> Result<Self, String> {
        Self::new(&RoofingFilterParams::default())
    }

    pub fn new(params: &RoofingFilterParams) -> Result<Self, String> {
        let shortest = params.shortest_cycle_period;
        if !shortest.is_finite() || shortest < 2.0 {
            return Err(
                "invalid roofing filter parameters: shortest_cycle_period must be finite and at least 2 bars"
                    .to_string(),
            );
        }

        let longest = params.longest_cycle_period;
        if !longest.is_finite() || longest <= shortest {
            return Err(
                "invalid roofing filter parameters: longest_cycle_period must be finite and greater than shortest"
                    .to_string(),
            );
        }

        let poles_raw = params.highpass_poles;
        if poles_raw != 1 && poles_raw != 2 {
            return Err(
                "invalid roofing filter parameters: highpass_poles must be integer 1 or 2"
                    .to_string(),
            );
        }
        let poles = poles_raw as usize;

        if params.zero_mean && poles != 1 {
            return Err(
                "invalid roofing filter parameters: zero_mean requires highpass_poles=1"
                    .to_string(),
            );
        }

        let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bar_comp);
        let quote_func = quote_component_value(quote_comp);
        let trade_func = trade_component_value(trade_comp);

        let angle = if poles == 1 {
            2.0 * PI / longest
        } else {
            SQRT_2 * PI / longest
        };
        let tangent = (angle / 2.0).tan();
        let alpha = 2.0 * tangent / (1.0 + tangent);
        let gain = 1.0 - alpha / 2.0;
        let decay = 1.0 - alpha;
        let (c1, c2, c3) = smoother_coefficients(shortest);
        let warmup_count = if poles == 2 || params.zero_mean { 4 } else { 3 };

        let mnemonic_text = mnemonic(params);
        let description_text = format!("Roofing Filter {mnemonic_text}");
        let metadata = build_metadata(
            Identifier::RoofingFilter,
            &mnemonic_text,
            &description_text,
            &[OutputText {
                mnemonic: mnemonic_text.clone(),
                description: description_text.clone(),
            }],
        );

        Ok(Self {
            metadata,
            poles,
            zero_mean: params.zero_mean,
            warmup_count,
            gain,
            decay,
            c1,
            c2,
            c3,
            count: 0,
            input1: 0.0,
            input2: 0.0,
            hp1: 0.0,
            hp2: 0.0,
            filt1: 0.0,
            filt2: 0.0,
            zero1: 0.0,
            current_value: f64::NAN,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    pub fn reset(&mut self) {
        self.count = 0;
        self.input1 = 0.0;
        self.input2 = 0.0;
        self.hp1 = 0.0;
        self.hp2 = 0.0;
        self.filt1 = 0.0;
        self.filt2 = 0.0;
        self.zero1 = 0.0;
        self.current_value = f64::NAN;
    }

    pub fn warmup_bars(&self) -> usize {
        self.warmup_count
    }

    pub fn value(&self) -> f64 {
        self.current_value
    }

    pub fn try_update(&mut self, sample: f64) -> Result<f64, String> {
        if sample.is_nan() {
            return Ok(f64::NAN);
        }
        if sample.is_infinite() {
            return Err("sample must be finite or NaN".to_string());
        }

        let (hp, filt, zero) = if self.count < self.poles {
            (0.0, 0.0, 0.0)
        } else {
            let hp_next = if self.poles == 2 {
                let difference = sample - 2.0 * self.input1 + self.input2;
                self.gain * self.gain * difference + 2.0 * self.decay * self.hp1
                    - self.decay * self.decay * self.hp2
            } else {
                self.gain * (sample - self.input1) + self.decay * self.hp1
            };
            let filt_next =
                self.c1 * (hp_next + self.hp1) / 2.0 + self.c2 * self.filt1 + self.c3 * self.filt2;
            let zero_next = if self.zero_mean {
                self.gain * (filt_next - self.filt1) + self.decay * self.zero1
            } else {
                0.0
            };
            (hp_next, filt_next, zero_next)
        };

        self.input2 = self.input1;
        self.input1 = sample;
        self.hp2 = self.hp1;
        self.hp1 = hp;
        self.filt2 = self.filt1;
        self.filt1 = filt;
        self.zero1 = zero;
        self.count += 1;

        let output = if self.zero_mean { zero } else { filt };
        if self.count > self.warmup_count {
            self.current_value = output;
            Ok(output)
        } else {
            Ok(f64::NAN)
        }
    }

    pub fn update(&mut self, sample: f64) -> f64 {
        self.try_update(sample)
            .expect("RoofingFilter::update: sample must be finite or NaN")
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let output = self.update(sample);
        vec![Box::new(Scalar::new(time, output)) as Box<dyn Any>]
    }
}

impl Indicator for RoofingFilter {
    fn is_primed(&self) -> bool {
        self.count > self.warmup_count
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

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::core::outputs::shape::Shape;
    use crate::indicators::factory::create_indicator;
    use crate::indicators::john_ehlers::roofing_filter::testdata;

    fn reference_cases() -> Vec<(RoofingFilterParams, Vec<f64>, usize)> {
        vec![
            (
                RoofingFilterParams {
                    shortest_cycle_period: 10.0,
                    longest_cycle_period: 48.0,
                    highpass_poles: 2,
                    zero_mean: false,
                    ..Default::default()
                },
                testdata::expected_lp10_hp48_po2_zm0(),
                4,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 40.0,
                    longest_cycle_period: 80.0,
                    highpass_poles: 2,
                    zero_mean: false,
                    ..Default::default()
                },
                testdata::expected_lp40_hp80_po2_zm0(),
                4,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 2.0,
                    longest_cycle_period: 3.0,
                    highpass_poles: 2,
                    zero_mean: false,
                    ..Default::default()
                },
                testdata::expected_lp2_hp3_po2_zm0(),
                4,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 5.0,
                    longest_cycle_period: 24.0,
                    highpass_poles: 2,
                    zero_mean: false,
                    ..Default::default()
                },
                testdata::expected_lp5_hp24_po2_zm0(),
                4,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 10.5,
                    longest_cycle_period: 48.0,
                    highpass_poles: 2,
                    zero_mean: false,
                    ..Default::default()
                },
                testdata::expected_lp10_5_hp48_po2_zm0(),
                4,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 10.0,
                    longest_cycle_period: 11.0,
                    highpass_poles: 2,
                    zero_mean: false,
                    ..Default::default()
                },
                testdata::expected_lp10_hp11_po2_zm0(),
                4,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 20.0,
                    longest_cycle_period: 80.0,
                    highpass_poles: 2,
                    zero_mean: false,
                    ..Default::default()
                },
                testdata::expected_lp20_hp80_po2_zm0(),
                4,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 40.0,
                    longest_cycle_period: 252.0,
                    highpass_poles: 2,
                    zero_mean: false,
                    ..Default::default()
                },
                testdata::expected_lp40_hp252_po2_zm0(),
                4,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 10.0,
                    longest_cycle_period: 48.0,
                    highpass_poles: 1,
                    zero_mean: false,
                    ..Default::default()
                },
                testdata::expected_lp10_hp48_po1_zm0(),
                3,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 2.0,
                    longest_cycle_period: 4.0,
                    highpass_poles: 1,
                    zero_mean: false,
                    ..Default::default()
                },
                testdata::expected_lp2_hp4_po1_zm0(),
                3,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 10.0,
                    longest_cycle_period: 48.0,
                    highpass_poles: 1,
                    zero_mean: true,
                    ..Default::default()
                },
                testdata::expected_lp10_hp48_po1_zm1(),
                4,
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 2.0,
                    longest_cycle_period: 4.0,
                    highpass_poles: 1,
                    zero_mean: true,
                    ..Default::default()
                },
                testdata::expected_lp2_hp4_po1_zm1(),
                4,
            ),
        ]
    }

    #[test]
    fn test_all_twelve_reference_cases() {
        let inputs = testdata::testdata::test_input();
        assert_eq!(inputs.len(), 252);
        for (params, expected, expected_warmup) in reference_cases() {
            let mut filter = RoofingFilter::new(&params).unwrap();
            assert_eq!(filter.warmup_bars(), expected_warmup);
            assert!(!filter.is_primed());
            assert!(filter.value().is_nan());
            assert_eq!(expected.len(), 252);

            for (index, &sample) in inputs.iter().enumerate() {
                let actual = filter.update(sample);
                let target = expected[index];
                if index < expected_warmup {
                    assert!(!filter.is_primed());
                    assert!(actual.is_nan());
                    assert!(filter.value().is_nan());
                } else {
                    assert!(filter.is_primed());
                    assert!(!target.is_nan());
                    let bound = 1e-11_f64.max(target.abs() * 1e-11);
                    assert!(
                        (actual - target).abs() <= bound,
                        "bar {index} ({}) actual={actual} target={target}",
                        mnemonic(&params)
                    );
                    assert!((filter.value() - actual).abs() <= 1e-15);
                }
            }
        }
    }

    #[test]
    fn test_nan_skips_state_and_reset_reproduces_start() {
        let inputs = testdata::testdata::test_input();
        for (params, _, _) in reference_cases() {
            let mut first = RoofingFilter::new(&params).unwrap();
            let mut second = RoofingFilter::new(&params).unwrap();
            for &sample in &inputs {
                assert!(first.update(f64::NAN).is_nan());
                let left = first.update(sample);
                let right = second.update(sample);
                assert_eq!(first.is_primed(), second.is_primed());
                if right.is_nan() {
                    assert!(left.is_nan());
                } else {
                    assert!((left - right).abs() <= 1e-15);
                }
            }

            first.reset();
            assert!(!first.is_primed());
            assert!(first.value().is_nan());
            let mut third = RoofingFilter::new(&params).unwrap();
            for &sample in &inputs {
                let left = first.update(sample);
                let right = third.update(sample);
                if right.is_nan() {
                    assert!(left.is_nan());
                } else {
                    assert!((left - right).abs() <= 1e-15);
                }
            }
        }
    }

    #[test]
    fn test_constant_ramp_and_period_four_boundary() {
        let mut constant_filter = RoofingFilter::default_instance().unwrap();
        let mut first_finite: Option<usize> = None;
        let mut last_constant = f64::NAN;
        for index in 0..500 {
            let output = constant_filter.update(7.0);
            if first_finite.is_none() && output.is_finite() {
                first_finite = Some(index);
            }
            last_constant = output;
        }
        assert_eq!(first_finite, Some(constant_filter.warmup_bars()));
        assert!(last_constant.abs() <= 1e-12);

        let mut ramp_filter = RoofingFilter::default_instance().unwrap();
        let mut last_ramp = f64::NAN;
        for index in 0..1000 {
            last_ramp = ramp_filter.update(index as f64);
        }
        assert!(last_ramp.abs() <= 1e-12);

        let boundary = RoofingFilter::new(&RoofingFilterParams {
            shortest_cycle_period: 2.0,
            longest_cycle_period: 4.0,
            highpass_poles: 1,
            zero_mean: false,
            ..Default::default()
        })
        .unwrap();
        assert!((boundary.gain - 0.5).abs() <= 1e-15);
        assert!((boundary.decay - 0.0).abs() <= 1e-15);
    }

    #[test]
    fn test_invalid_parameters_and_infinite_samples() {
        let invalid_cases = [
            RoofingFilterParams {
                shortest_cycle_period: 1.0,
                ..Default::default()
            },
            RoofingFilterParams {
                shortest_cycle_period: 0.0,
                ..Default::default()
            },
            RoofingFilterParams {
                shortest_cycle_period: -2.0,
                ..Default::default()
            },
            RoofingFilterParams {
                shortest_cycle_period: f64::NAN,
                ..Default::default()
            },
            RoofingFilterParams {
                shortest_cycle_period: f64::INFINITY,
                ..Default::default()
            },
            RoofingFilterParams {
                shortest_cycle_period: 10.0,
                longest_cycle_period: 10.0,
                ..Default::default()
            },
            RoofingFilterParams {
                shortest_cycle_period: 10.0,
                longest_cycle_period: 5.0,
                ..Default::default()
            },
            RoofingFilterParams {
                shortest_cycle_period: 10.0,
                longest_cycle_period: f64::NAN,
                ..Default::default()
            },
            RoofingFilterParams {
                shortest_cycle_period: 10.0,
                longest_cycle_period: f64::INFINITY,
                ..Default::default()
            },
            RoofingFilterParams {
                highpass_poles: 0,
                ..Default::default()
            },
            RoofingFilterParams {
                highpass_poles: 3,
                ..Default::default()
            },
            RoofingFilterParams {
                highpass_poles: 2,
                zero_mean: true,
                ..Default::default()
            },
        ];
        for bad_params in invalid_cases {
            assert!(RoofingFilter::new(&bad_params).is_err());
        }

        let mut filter = RoofingFilter::default_instance().unwrap();
        assert!(filter.try_update(f64::INFINITY).is_err());
        assert!(filter.try_update(f64::NEG_INFINITY).is_err());
        assert!(!filter.is_primed());
    }

    #[test]
    fn test_metadata_mnemonics_entities_and_factory() {
        let filter = RoofingFilter::new(&default_params()).unwrap();
        let meta = filter.metadata();
        assert_eq!(meta.identifier, Identifier::RoofingFilter);
        assert_eq!(meta.mnemonic, "roof2hp(10, 48)");
        assert_eq!(meta.description, "Roofing Filter roof2hp(10, 48)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, RoofingFilterOutput::Value as i32);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
        assert_eq!(meta.outputs[0].mnemonic, "roof2hp(10, 48)");
        assert_eq!(meta.outputs[0].description, "Roofing Filter roof2hp(10, 48)");

        let mnemonic_cases = [
            (RoofingFilterParams::default(), "roof2hp(10, 48)"),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 10.5,
                    longest_cycle_period: 48.0,
                    highpass_poles: 2,
                    zero_mean: false,
                    ..Default::default()
                },
                "roof2hp(10.5, 48)",
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 10.0,
                    longest_cycle_period: 48.0,
                    highpass_poles: 1,
                    zero_mean: false,
                    ..Default::default()
                },
                "roof1hp(10, 48)",
            ),
            (
                RoofingFilterParams {
                    shortest_cycle_period: 10.0,
                    longest_cycle_period: 48.0,
                    highpass_poles: 1,
                    zero_mean: true,
                    ..Default::default()
                },
                "roof1hpzm(10, 48)",
            ),
            (
                RoofingFilterParams {
                    bar_component: Some(BarComponent::Median),
                    ..Default::default()
                },
                "roof2hp(10, 48, hl/2)",
            ),
            (
                RoofingFilterParams {
                    quote_component: Some(QuoteComponent::Bid),
                    ..Default::default()
                },
                "roof2hp(10, 48, b)",
            ),
            (
                RoofingFilterParams {
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "roof2hp(10, 48, v)",
            ),
            (
                RoofingFilterParams {
                    bar_component: Some(BarComponent::Open),
                    quote_component: Some(QuoteComponent::Ask),
                    ..Default::default()
                },
                "roof2hp(10, 48, o, a)",
            ),
            (
                RoofingFilterParams {
                    bar_component: Some(BarComponent::Typical),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "roof2hp(10, 48, hlc/3, v)",
            ),
            (
                RoofingFilterParams {
                    quote_component: Some(QuoteComponent::Weighted),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "roof2hp(10, 48, (bbs+aas)/(bs+as), v)",
            ),
        ];
        for (candidate, expected_mnemonic) in mnemonic_cases {
            assert_eq!(mnemonic(&candidate), expected_mnemonic);
            assert_eq!(
                RoofingFilter::new(&candidate).unwrap().metadata().mnemonic,
                expected_mnemonic
            );
        }

        let closes = testdata::testdata::test_input();
        let opens = testdata::testdata::test_open();
        let highs = testdata::testdata::test_high();
        let lows = testdata::testdata::test_low();

        let mut bar_filter = RoofingFilter::default_instance().unwrap();
        let mut scalar_filter = RoofingFilter::default_instance().unwrap();
        for index in 0..10 {
            let timestamp = 1000 + index as i64;
            let bar = Bar::new(
                timestamp,
                opens[index],
                highs[index],
                lows[index],
                closes[index],
                1000.0,
            );
            let scalar = Scalar::new(timestamp, closes[index]);
            let bar_out = bar_filter.update_bar(&bar);
            let scalar_out = scalar_filter.update_scalar(&scalar);
            assert_eq!(bar_out.len(), 1);
            let bar_scalar = bar_out[0].downcast_ref::<Scalar>().unwrap();
            let ref_scalar = scalar_out[0].downcast_ref::<Scalar>().unwrap();
            assert_eq!(bar_scalar.time, timestamp);
            if ref_scalar.value.is_nan() {
                assert!(bar_scalar.value.is_nan());
            } else {
                assert!((bar_scalar.value - ref_scalar.value).abs() <= 1e-15);
            }
        }

        let mut quote_filter = RoofingFilter::default_instance().unwrap();
        let mut quote_out = Vec::new();
        for _ in 0..5 {
            let quote = Quote::new(2000, 99.0, 101.0, 1.0, 1.0);
            quote_out = quote_filter.update_quote(&quote);
        }
        assert_eq!(quote_out.len(), 1);
        assert!(!quote_out[0].downcast_ref::<Scalar>().unwrap().value.is_nan());

        let mut trade_filter = RoofingFilter::default_instance().unwrap();
        let mut trade_out = Vec::new();
        for _ in 0..5 {
            let trade = Trade::new(3000, 100.0, 50.0);
            trade_out = trade_filter.update_trade(&trade);
        }
        assert_eq!(trade_out.len(), 1);
        assert!(!trade_out[0].downcast_ref::<Scalar>().unwrap().value.is_nan());

        let created_default = create_indicator(Identifier::RoofingFilter, "{}").unwrap();
        assert_eq!(created_default.metadata().mnemonic, "roof2hp(10, 48)");

        let created_custom = create_indicator(
            Identifier::RoofingFilter,
            r#"{"shortestCyclePeriod": 10.5, "longestCyclePeriod": 48, "highpassPoles": 1, "zeroMean": true}"#,
        )
        .unwrap();
        assert_eq!(created_custom.metadata().mnemonic, "roof1hpzm(10.5, 48)");

        assert!(create_indicator(Identifier::RoofingFilter, r#"{"highpassPoles": 1.5}"#).is_err());
        assert!(create_indicator(Identifier::RoofingFilter, r#"{"highpassPoles": 2, "zeroMean": true}"#).is_err());
    }
}
