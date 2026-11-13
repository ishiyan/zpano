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
pub enum SuperSmootherOutput {
    Value = 1,
}

#[derive(Debug, Clone, Copy)]
pub struct SuperSmootherParams {
    pub shortest_cycle_period: f64,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

pub type Params = SuperSmootherParams;

impl Default for SuperSmootherParams {
    fn default() -> Self {
        Self {
            shortest_cycle_period: 10.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

pub fn default_params() -> SuperSmootherParams {
    SuperSmootherParams::default()
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

pub fn mnemonic(params: &SuperSmootherParams) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    let shortest_text = format_number(params.shortest_cycle_period);
    format!("ss({shortest_text}{suffix})")
}

pub struct SuperSmoother {
    metadata: Metadata,
    warmup_count: usize,
    c1: f64,
    c2: f64,
    c3: f64,
    count: usize,
    input1: f64,
    output1: f64,
    output2: f64,
    current_value: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl SuperSmoother {
    pub fn default_instance() -> Result<Self, String> {
        Self::new(&SuperSmootherParams::default())
    }

    pub fn new(params: &SuperSmootherParams) -> Result<Self, String> {
        let shortest = params.shortest_cycle_period;
        if !shortest.is_finite() || shortest < 2.0 {
            return Err(
                "invalid super smoother parameters: shortest_cycle_period must be finite and at least 2 bars"
                    .to_string(),
            );
        }

        let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bar_comp);
        let quote_func = quote_component_value(quote_comp);
        let trade_func = trade_component_value(trade_comp);

        let (c1, c2, c3) = smoother_coefficients(shortest);
        let mnemonic_text = mnemonic(params);
        let description_text = format!("Super Smoother {mnemonic_text}");
        let metadata = build_metadata(
            Identifier::SuperSmoother,
            &mnemonic_text,
            &description_text,
            &[OutputText {
                mnemonic: mnemonic_text.clone(),
                description: description_text.clone(),
            }],
        );

        Ok(Self {
            metadata,
            warmup_count: 2,
            c1,
            c2,
            c3,
            count: 0,
            input1: 0.0,
            output1: 0.0,
            output2: 0.0,
            current_value: f64::NAN,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    pub fn reset(&mut self) {
        self.count = 0;
        self.input1 = 0.0;
        self.output1 = 0.0;
        self.output2 = 0.0;
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

        if self.count == 0 {
            self.input1 = sample;
            self.output1 = sample;
            self.output2 = sample;
        }

        let output =
            self.c1 * (sample + self.input1) / 2.0 + self.c2 * self.output1 + self.c3 * self.output2;
        self.input1 = sample;
        self.output2 = self.output1;
        self.output1 = output;
        self.count += 1;

        if self.count > self.warmup_count {
            self.current_value = output;
            Ok(output)
        } else {
            Ok(f64::NAN)
        }
    }

    pub fn update(&mut self, sample: f64) -> f64 {
        self.try_update(sample)
            .expect("SuperSmoother::update: sample must be finite or NaN")
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let output = self.update(sample);
        vec![Box::new(Scalar::new(time, output)) as Box<dyn Any>]
    }
}

impl Indicator for SuperSmoother {
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
    use crate::indicators::john_ehlers::super_smoother::testdata;

    fn reference_cases() -> Vec<(SuperSmootherParams, Vec<f64>, usize)> {
        vec![
            (
                SuperSmootherParams {
                    shortest_cycle_period: 2.0,
                    ..Default::default()
                },
                testdata::expected_p2(),
                2,
            ),
            (
                SuperSmootherParams {
                    shortest_cycle_period: 3.0,
                    ..Default::default()
                },
                testdata::expected_p3(),
                2,
            ),
            (
                SuperSmootherParams {
                    shortest_cycle_period: 5.0,
                    ..Default::default()
                },
                testdata::expected_p5(),
                2,
            ),
            (
                SuperSmootherParams {
                    shortest_cycle_period: 8.0,
                    ..Default::default()
                },
                testdata::expected_p8(),
                2,
            ),
            (
                SuperSmootherParams {
                    shortest_cycle_period: 10.0,
                    ..Default::default()
                },
                testdata::expected_p10(),
                2,
            ),
            (
                SuperSmootherParams {
                    shortest_cycle_period: 10.5,
                    ..Default::default()
                },
                testdata::expected_p10_5(),
                2,
            ),
            (
                SuperSmootherParams {
                    shortest_cycle_period: 20.0,
                    ..Default::default()
                },
                testdata::expected_p20(),
                2,
            ),
            (
                SuperSmootherParams {
                    shortest_cycle_period: 40.0,
                    ..Default::default()
                },
                testdata::expected_p40(),
                2,
            ),
            (
                SuperSmootherParams {
                    shortest_cycle_period: 80.0,
                    ..Default::default()
                },
                testdata::expected_p80(),
                2,
            ),
            (
                SuperSmootherParams {
                    shortest_cycle_period: 252.0,
                    ..Default::default()
                },
                testdata::expected_p252(),
                2,
            ),
        ]
    }

    #[test]
    fn test_all_ten_reference_cases() {
        let inputs = testdata::testdata::test_input();
        assert_eq!(inputs.len(), 252);
        for (params, expected, expected_warmup) in reference_cases() {
            let mut filter = SuperSmoother::new(&params).unwrap();
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
            let mut first = SuperSmoother::new(&params).unwrap();
            let mut second = SuperSmoother::new(&params).unwrap();
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
            let mut third = SuperSmoother::new(&params).unwrap();
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
    fn test_constant_preservation_and_nyquist_zero() {
        let mut constant_filter = SuperSmoother::default_instance().unwrap();
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
        assert!((last_constant - 7.0).abs() <= 1e-12);

        let mut nyquist_filter = SuperSmoother::default_instance().unwrap();
        let mut last_nyquist = f64::NAN;
        for index in 0..500 {
            let sample = if index % 2 == 0 { 1.0 } else { -1.0 };
            last_nyquist = nyquist_filter.update(sample);
        }
        assert!(last_nyquist.abs() <= 1e-12);
    }

    #[test]
    fn test_invalid_parameters_and_infinite_samples() {
        let invalid_cases = [
            SuperSmootherParams {
                shortest_cycle_period: 1.0,
                ..Default::default()
            },
            SuperSmootherParams {
                shortest_cycle_period: 1.999,
                ..Default::default()
            },
            SuperSmootherParams {
                shortest_cycle_period: 0.0,
                ..Default::default()
            },
            SuperSmootherParams {
                shortest_cycle_period: -2.0,
                ..Default::default()
            },
            SuperSmootherParams {
                shortest_cycle_period: f64::NAN,
                ..Default::default()
            },
            SuperSmootherParams {
                shortest_cycle_period: f64::INFINITY,
                ..Default::default()
            },
            SuperSmootherParams {
                shortest_cycle_period: f64::NEG_INFINITY,
                ..Default::default()
            },
        ];
        for candidate in invalid_cases {
            assert!(SuperSmoother::new(&candidate).is_err());
        }

        let mut filter = SuperSmoother::default_instance().unwrap();
        assert!(filter.try_update(f64::INFINITY).is_err());
        assert!(filter.try_update(f64::NEG_INFINITY).is_err());
        assert!(!filter.is_primed());
    }

    #[test]
    fn test_metadata_mnemonics_entities_and_factory() {
        let filter = SuperSmoother::new(&default_params()).unwrap();
        let meta = filter.metadata();
        assert_eq!(meta.identifier, Identifier::SuperSmoother);
        assert_eq!(meta.mnemonic, "ss(10)");
        assert_eq!(meta.description, "Super Smoother ss(10)");
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, SuperSmootherOutput::Value as i32);
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
        assert_eq!(meta.outputs[0].mnemonic, "ss(10)");
        assert_eq!(meta.outputs[0].description, "Super Smoother ss(10)");

        let mnemonic_cases = [
            (SuperSmootherParams::default(), "ss(10)"),
            (
                SuperSmootherParams {
                    shortest_cycle_period: 10.5,
                    ..Default::default()
                },
                "ss(10.5)",
            ),
            (
                SuperSmootherParams {
                    bar_component: Some(BarComponent::Median),
                    ..Default::default()
                },
                "ss(10, hl/2)",
            ),
            (
                SuperSmootherParams {
                    quote_component: Some(QuoteComponent::Bid),
                    ..Default::default()
                },
                "ss(10, b)",
            ),
            (
                SuperSmootherParams {
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "ss(10, v)",
            ),
            (
                SuperSmootherParams {
                    bar_component: Some(BarComponent::Open),
                    quote_component: Some(QuoteComponent::Ask),
                    ..Default::default()
                },
                "ss(10, o, a)",
            ),
            (
                SuperSmootherParams {
                    bar_component: Some(BarComponent::Typical),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "ss(10, hlc/3, v)",
            ),
            (
                SuperSmootherParams {
                    quote_component: Some(QuoteComponent::Weighted),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "ss(10, (bbs+aas)/(bs+as), v)",
            ),
        ];
        for (candidate, expected_mnemonic) in mnemonic_cases {
            assert_eq!(mnemonic(&candidate), expected_mnemonic);
            assert_eq!(
                SuperSmoother::new(&candidate).unwrap().metadata().mnemonic,
                expected_mnemonic
            );
        }

        let closes = testdata::testdata::test_input();
        let opens = testdata::testdata::test_open();
        let highs = testdata::testdata::test_high();
        let lows = testdata::testdata::test_low();

        let mut bar_filter = SuperSmoother::default_instance().unwrap();
        let mut scalar_filter = SuperSmoother::default_instance().unwrap();
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

        let mut quote_filter = SuperSmoother::default_instance().unwrap();
        let mut quote_out = Vec::new();
        for _ in 0..5 {
            let quote = Quote::new(2000, 99.0, 101.0, 1.0, 1.0);
            quote_out = quote_filter.update_quote(&quote);
        }
        assert_eq!(quote_out.len(), 1);
        assert!(!quote_out[0].downcast_ref::<Scalar>().unwrap().value.is_nan());

        let mut trade_filter = SuperSmoother::default_instance().unwrap();
        let mut trade_out = Vec::new();
        for _ in 0..5 {
            let trade = Trade::new(3000, 100.0, 50.0);
            trade_out = trade_filter.update_trade(&trade);
        }
        assert_eq!(trade_out.len(), 1);
        assert!(!trade_out[0].downcast_ref::<Scalar>().unwrap().value.is_nan());

        let created_default = create_indicator(Identifier::SuperSmoother, "{}").unwrap();
        assert_eq!(created_default.metadata().mnemonic, "ss(10)");

        let created_custom = create_indicator(
            Identifier::SuperSmoother,
            r#"{"shortestCyclePeriod": 10.5}"#,
        )
        .unwrap();
        assert_eq!(created_custom.metadata().mnemonic, "ss(10.5)");

        assert!(create_indicator(Identifier::SuperSmoother, r#"{"shortestCyclePeriod": "bad"}"#).is_err());
        assert!(create_indicator(Identifier::SuperSmoother, r#"{"shortestCyclePeriod": 0}"#).is_err());
    }
}
