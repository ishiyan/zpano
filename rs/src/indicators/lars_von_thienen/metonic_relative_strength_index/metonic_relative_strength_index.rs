use super::super::harmonic_cycle::validation::validate_integer;
use super::engine::{Engine, Snapshot};
use crate::entities::bar::Bar;
use crate::entities::bar_component::{
    BarComponent, DEFAULT_BAR_COMPONENT, component_value as bar_component_value,
};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{
    DEFAULT_QUOTE_COMPONENT, QuoteComponent, component_value as quote_component_value,
};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{
    DEFAULT_TRADE_COMPONENT, TradeComponent, component_value as trade_component_value,
};
use crate::indicators::core::build_metadata::{OutputText, build_metadata};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
/// Parameters matching the prepared reference; omitted components use entity defaults.
#[derive(Clone)]
pub struct MetonicRelativeStrengthIndexParams {
    /// The length; reference default 13.
    pub length: usize,
    /// The delay1; reference default 6939.
    pub delay1: usize,
    /// The delay2; reference default 4016.
    pub delay2: usize,
    /// The delay3; reference default 2922.
    pub delay3: usize,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for MetonicRelativeStrengthIndexParams {
    fn default() -> Self {
        Self {
            length: 13,
            delay1: 6939,
            delay2: 4016,
            delay3: 2922,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MetonicRelativeStrengthIndexOutput {
    Value = 1,
    LinearRsi = 2,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct MetonicRelativeStrengthIndex {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl MetonicRelativeStrengthIndex {
    /// Construct a validated streaming indicator.
    pub fn new(p: &MetonicRelativeStrengthIndexParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.length as i64, "length", 1) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.delay1 as i64, "delay1", 1) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.delay2 as i64, "delay2", 1) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.delay3 as i64, "delay3", 1) {
            return Err(error);
        }
        if (p.delay1 as i64) == (p.delay2 as i64)
            || (p.delay1 as i64) == (p.delay3 as i64)
            || (p.delay2 as i64) == (p.delay3 as i64)
        {
            return Err("three delays must be distinct".into());
        }
        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let suffix = component_triple_mnemonic(bc, qc, tc);
        let parameters = [
            (p.length as i64).to_string(),
            (p.delay1 as i64).to_string(),
            (p.delay2 as i64).to_string(),
            (p.delay3 as i64).to_string(),
        ]
        .join(",");
        let mnemonic = format!("mrsi({parameters}{suffix})");
        let description = format!("Metonic relative strength index {mnemonic}");
        let mut engine = Engine {
            p: p.clone(),
            ..Default::default()
        };
        engine.reset();
        Ok(Self {
            engine,
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
            mnemonic,
            description,
        })
    }
    /// Accept one completed observation.
    pub fn update(&mut self, price: f64, day: Option<i64>) -> Snapshot {
        self.engine.update(price, day)
    }
    /// Clear history and retain parameters.
    pub fn reset(&mut self) {
        self.engine.reset();
    }
    /// Minimum initial observations before base readiness.
    pub fn warmup_bars(&self) -> i64 {
        self.engine.warmup()
    }
    fn wrap(&self, time: i64, result: Snapshot) -> Output {
        let mut output: Output = Vec::new();
        output.push(Box::new(Scalar::new(time, result.value)));
        output.push(Box::new(Scalar::new(time, result.linear_rsi)));
        output
    }
}
impl Indicator for MetonicRelativeStrengthIndex {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::MetonicRelativeStrengthIndex,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} value", self.mnemonic),
                    description: format!("{} value", self.description),
                },
                OutputText {
                    mnemonic: format!("{} linear_rsi", self.mnemonic),
                    description: format!("{} linear_rsi", self.description),
                },
            ],
        )
    }
    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value, Some(super::super::metonic::calendar::day_from_unix_seconds(sample.time)));
        self.wrap(sample.time, value)
    }
    fn update_bar(&mut self, sample: &Bar) -> Output {
        let value = self.update((self.bar_func)(sample), Some(super::super::metonic::calendar::day_from_unix_seconds(sample.time)));
        self.wrap(sample.time, value)
    }
    fn update_quote(&mut self, sample: &Quote) -> Output {
        let value = self.update(
            (self.quote_func)(sample),
            Some(super::super::metonic::calendar::day_from_unix_seconds(sample.time)),
        );
        self.wrap(sample.time, value)
    }
    fn update_trade(&mut self, sample: &Trade) -> Output {
        let value = self.update(
            (self.trade_func)(sample),
            Some(super::super::metonic::calendar::day_from_unix_seconds(sample.time)),
        );
        self.wrap(sample.time, value)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    mod saved {
        include!("testdata.rs");
    }
    use saved::testdata as data;
    #[test]
    fn production_calendar_delays() {
        let input = data::test_input();
        let mut indicator =
            MetonicRelativeStrengthIndex::new(&MetonicRelativeStrengthIndexParams {
                length: 1,
                ..Default::default()
            })
            .unwrap();
        for i in 0..6945 {
            let actual = indicator
                .update(input[i % input.len()], Some(i as i64))
                .value;
            if i <= 6939 {
                assert!(actual.is_nan());
            } else {
                let mut sum = 0.0;
                for lag in [6939, 4016, 2922] {
                    if input[(i - lag) % input.len()] >= input[(i - lag - 1) % input.len()] {
                        sum += 100.0;
                    }
                }
                assert!((actual - sum / 3.0).abs() < 1e-13);
            }
        }
    }

    #[test]
    fn prepared_case_0() {
        let p = MetonicRelativeStrengthIndexParams {
            length: 13,
            delay1: 6939,
            delay2: 4016,
            delay3: 2922,
            ..Default::default()
        };
        let mut indicator = MetonicRelativeStrengthIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d16939_d24016_d32922_value()),
            (
                "LINEAR_RSI",
                data::expected_l13_d16939_d24016_d32922_linear_rsi(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64));
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_RSI", result.linear_rsi),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-13_f64.max(values[index].abs() * 5e-15),
                        "bar {index} {field}: {value:.17} vs {:.17}",
                        values[index]
                    );
                }
            }
        }
        indicator.reset();
        assert!(!indicator.is_primed());
    }
    #[test]
    fn prepared_case_1() {
        let p = MetonicRelativeStrengthIndexParams {
            length: 13,
            delay1: 19,
            delay2: 11,
            delay3: 8,
            ..Default::default()
        };
        let mut indicator = MetonicRelativeStrengthIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d119_d211_d38_value()),
            ("LINEAR_RSI", data::expected_l13_d119_d211_d38_linear_rsi()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64));
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_RSI", result.linear_rsi),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-13_f64.max(values[index].abs() * 5e-15),
                        "bar {index} {field}: {value:.17} vs {:.17}",
                        values[index]
                    );
                }
            }
        }
        indicator.reset();
        assert!(!indicator.is_primed());
    }
    #[test]
    fn prepared_case_2() {
        let p = MetonicRelativeStrengthIndexParams {
            length: 1,
            delay1: 19,
            delay2: 11,
            delay3: 8,
            ..Default::default()
        };
        let mut indicator = MetonicRelativeStrengthIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l1_d119_d211_d38_value()),
            ("LINEAR_RSI", data::expected_l1_d119_d211_d38_linear_rsi()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64));
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_RSI", result.linear_rsi),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-13_f64.max(values[index].abs() * 5e-15),
                        "bar {index} {field}: {value:.17} vs {:.17}",
                        values[index]
                    );
                }
            }
        }
        indicator.reset();
        assert!(!indicator.is_primed());
    }
    #[test]
    fn prepared_case_3() {
        let p = MetonicRelativeStrengthIndexParams {
            length: 14,
            delay1: 19,
            delay2: 11,
            delay3: 8,
            ..Default::default()
        };
        let mut indicator = MetonicRelativeStrengthIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l14_d119_d211_d38_value()),
            ("LINEAR_RSI", data::expected_l14_d119_d211_d38_linear_rsi()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64));
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_RSI", result.linear_rsi),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-13_f64.max(values[index].abs() * 5e-15),
                        "bar {index} {field}: {value:.17} vs {:.17}",
                        values[index]
                    );
                }
            }
        }
        indicator.reset();
        assert!(!indicator.is_primed());
    }
    #[test]
    fn prepared_case_4() {
        let p = MetonicRelativeStrengthIndexParams {
            length: 50,
            delay1: 100,
            delay2: 60,
            delay3: 40,
            ..Default::default()
        };
        let mut indicator = MetonicRelativeStrengthIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l50_d1100_d260_d340_value()),
            (
                "LINEAR_RSI",
                data::expected_l50_d1100_d260_d340_linear_rsi(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64));
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_RSI", result.linear_rsi),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-13_f64.max(values[index].abs() * 5e-15),
                        "bar {index} {field}: {value:.17} vs {:.17}",
                        values[index]
                    );
                }
            }
        }
        indicator.reset();
        assert!(!indicator.is_primed());
    }
    #[test]
    fn prepared_case_5() {
        let p = MetonicRelativeStrengthIndexParams {
            length: 13,
            delay1: 8,
            delay2: 19,
            delay3: 11,
            ..Default::default()
        };
        let mut indicator = MetonicRelativeStrengthIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d18_d219_d311_value()),
            ("LINEAR_RSI", data::expected_l13_d18_d219_d311_linear_rsi()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64));
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_RSI", result.linear_rsi),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-13_f64.max(values[index].abs() * 5e-15),
                        "bar {index} {field}: {value:.17} vs {:.17}",
                        values[index]
                    );
                }
            }
        }
        indicator.reset();
        assert!(!indicator.is_primed());
    }
    #[test]
    fn prepared_case_6() {
        let p = MetonicRelativeStrengthIndexParams {
            length: 13,
            delay1: 238,
            delay2: 100,
            delay3: 50,
            ..Default::default()
        };
        let mut indicator = MetonicRelativeStrengthIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d1238_d2100_d350_value()),
            (
                "LINEAR_RSI",
                data::expected_l13_d1238_d2100_d350_linear_rsi(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64));
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_RSI", result.linear_rsi),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-13_f64.max(values[index].abs() * 5e-15),
                        "bar {index} {field}: {value:.17} vs {:.17}",
                        values[index]
                    );
                }
            }
        }
        indicator.reset();
        assert!(!indicator.is_primed());
    }
    #[test]
    fn prepared_case_7() {
        let p = MetonicRelativeStrengthIndexParams {
            length: 13,
            delay1: 239,
            delay2: 100,
            delay3: 50,
            ..Default::default()
        };
        let mut indicator = MetonicRelativeStrengthIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d1239_d2100_d350_value()),
            (
                "LINEAR_RSI",
                data::expected_l13_d1239_d2100_d350_linear_rsi(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64));
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_RSI", result.linear_rsi),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-13_f64.max(values[index].abs() * 5e-15),
                        "bar {index} {field}: {value:.17} vs {:.17}",
                        values[index]
                    );
                }
            }
        }
        indicator.reset();
        assert!(!indicator.is_primed());
    }
}

#[cfg(test)]
mod component_contract {
    use super::*;
    #[test]
    fn mnemonic_suffixes() {
        for (bc, qc, tc) in [
            (None, None, None),
            (Some(BarComponent::Open), None, None),
            (None, Some(QuoteComponent::Bid), None),
            (None, None, Some(TradeComponent::Volume)),
            (Some(BarComponent::Open), Some(QuoteComponent::Bid), None),
            (Some(BarComponent::Open), None, Some(TradeComponent::Volume)),
            (
                None,
                Some(QuoteComponent::Bid),
                Some(TradeComponent::Volume),
            ),
        ] {
            let p = MetonicRelativeStrengthIndexParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = MetonicRelativeStrengthIndex::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
