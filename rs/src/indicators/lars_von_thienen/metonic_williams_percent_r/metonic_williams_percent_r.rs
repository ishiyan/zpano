use super::super::harmonic_cycle::validation::validate_integer;
use super::engine::{Engine, Snapshot};
use crate::entities::bar::Bar;
use crate::entities::bar_component::{
    DEFAULT_BAR_COMPONENT, component_value as bar_component_value,
};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{
    DEFAULT_QUOTE_COMPONENT, component_value as quote_component_value,
};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{
    DEFAULT_TRADE_COMPONENT, component_value as trade_component_value,
};
use crate::indicators::core::build_metadata::{OutputText, build_metadata};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
/// Parameters matching the prepared reference; omitted components use entity defaults.
#[derive(Clone)]
pub struct MetonicWilliamsPercentRParams {
    /// The length; reference default 13.
    pub length: usize,
    /// The delay1; reference default 6939.
    pub delay1: usize,
    /// The delay2; reference default 4016.
    pub delay2: usize,
    /// The delay3; reference default 2922.
    pub delay3: usize,
}
impl Default for MetonicWilliamsPercentRParams {
    fn default() -> Self {
        Self {
            length: 13,
            delay1: 6939,
            delay2: 4016,
            delay3: 2922,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MetonicWilliamsPercentROutput {
    Value = 1,
    LinearPercentR = 2,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct MetonicWilliamsPercentR {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl MetonicWilliamsPercentR {
    /// Construct a validated streaming indicator.
    pub fn new(p: &MetonicWilliamsPercentRParams) -> Result<Self, String> {
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
        let bc = DEFAULT_BAR_COMPONENT;
        let qc = DEFAULT_QUOTE_COMPONENT;
        let tc = DEFAULT_TRADE_COMPONENT;
        let suffix = String::new();
        let parameters = [
            (p.length as i64).to_string(),
            (p.delay1 as i64).to_string(),
            (p.delay2 as i64).to_string(),
            (p.delay3 as i64).to_string(),
        ]
        .join(",");
        let mnemonic = format!("mwillr({parameters}{suffix})");
        let description = format!("Metonic williams percent r {mnemonic}");
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
    pub fn update(&mut self, price: f64, day: Option<i64>, high: f64, low: f64) -> Snapshot {
        self.engine.update(price, day, high, low)
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
        output.push(Box::new(Scalar::new(time, result.linear_percent_r)));
        output
    }
}
impl Indicator for MetonicWilliamsPercentR {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::MetonicWilliamsPercentR,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} value", self.mnemonic),
                    description: format!("{} value", self.description),
                },
                OutputText {
                    mnemonic: format!("{} linear_percent_r", self.mnemonic),
                    description: format!("{} linear_percent_r", self.description),
                },
            ],
        )
    }
    fn update_scalar(&mut self, _sample: &Scalar) -> Output {
        panic!("conventional Metonic Williams %R requires a completed HLC bar")
    }
    fn update_bar(&mut self, sample: &Bar) -> Output {
        let value = self.update(
            sample.close,
            Some(super::super::metonic::calendar::day_from_unix_seconds(sample.time)),
            sample.high,
            sample.low,
        );
        self.wrap(sample.time, value)
    }
    fn update_quote(&mut self, _sample: &Quote) -> Output {
        panic!("conventional Metonic Williams %R requires a completed HLC bar")
    }
    fn update_trade(&mut self, _sample: &Trade) -> Output {
        panic!("conventional Metonic Williams %R requires a completed HLC bar")
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
    fn prepared_case_0() {
        let p = MetonicWilliamsPercentRParams {
            length: 13,
            delay1: 6939,
            delay2: 4016,
            delay3: 2922,
            ..Default::default()
        };
        let mut indicator = MetonicWilliamsPercentR::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d16939_d24016_d32922_value()),
            (
                "LINEAR_PERCENT_R",
                data::expected_l13_d16939_d24016_d32922_linear_percent_r(),
            ),
        ];
        let high = data::test_high();
        let low = data::test_low();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), high[index], low[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_PERCENT_R", result.linear_percent_r),
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
        let p = MetonicWilliamsPercentRParams {
            length: 13,
            delay1: 19,
            delay2: 11,
            delay3: 8,
            ..Default::default()
        };
        let mut indicator = MetonicWilliamsPercentR::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d119_d211_d38_value()),
            (
                "LINEAR_PERCENT_R",
                data::expected_l13_d119_d211_d38_linear_percent_r(),
            ),
        ];
        let high = data::test_high();
        let low = data::test_low();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), high[index], low[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_PERCENT_R", result.linear_percent_r),
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
        let p = MetonicWilliamsPercentRParams {
            length: 1,
            delay1: 19,
            delay2: 11,
            delay3: 8,
            ..Default::default()
        };
        let mut indicator = MetonicWilliamsPercentR::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l1_d119_d211_d38_value()),
            (
                "LINEAR_PERCENT_R",
                data::expected_l1_d119_d211_d38_linear_percent_r(),
            ),
        ];
        let high = data::test_high();
        let low = data::test_low();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), high[index], low[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_PERCENT_R", result.linear_percent_r),
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
        let p = MetonicWilliamsPercentRParams {
            length: 14,
            delay1: 19,
            delay2: 11,
            delay3: 8,
            ..Default::default()
        };
        let mut indicator = MetonicWilliamsPercentR::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l14_d119_d211_d38_value()),
            (
                "LINEAR_PERCENT_R",
                data::expected_l14_d119_d211_d38_linear_percent_r(),
            ),
        ];
        let high = data::test_high();
        let low = data::test_low();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), high[index], low[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_PERCENT_R", result.linear_percent_r),
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
        let p = MetonicWilliamsPercentRParams {
            length: 50,
            delay1: 100,
            delay2: 60,
            delay3: 40,
            ..Default::default()
        };
        let mut indicator = MetonicWilliamsPercentR::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l50_d1100_d260_d340_value()),
            (
                "LINEAR_PERCENT_R",
                data::expected_l50_d1100_d260_d340_linear_percent_r(),
            ),
        ];
        let high = data::test_high();
        let low = data::test_low();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), high[index], low[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_PERCENT_R", result.linear_percent_r),
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
        let p = MetonicWilliamsPercentRParams {
            length: 13,
            delay1: 8,
            delay2: 19,
            delay3: 11,
            ..Default::default()
        };
        let mut indicator = MetonicWilliamsPercentR::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d18_d219_d311_value()),
            (
                "LINEAR_PERCENT_R",
                data::expected_l13_d18_d219_d311_linear_percent_r(),
            ),
        ];
        let high = data::test_high();
        let low = data::test_low();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), high[index], low[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_PERCENT_R", result.linear_percent_r),
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
        let p = MetonicWilliamsPercentRParams {
            length: 13,
            delay1: 239,
            delay2: 100,
            delay3: 50,
            ..Default::default()
        };
        let mut indicator = MetonicWilliamsPercentR::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d1239_d2100_d350_value()),
            (
                "LINEAR_PERCENT_R",
                data::expected_l13_d1239_d2100_d350_linear_percent_r(),
            ),
        ];
        let high = data::test_high();
        let low = data::test_low();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), high[index], low[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_PERCENT_R", result.linear_percent_r),
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
        let p = MetonicWilliamsPercentRParams {
            length: 13,
            delay1: 240,
            delay2: 100,
            delay3: 50,
            ..Default::default()
        };
        let mut indicator = MetonicWilliamsPercentR::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d1240_d2100_d350_value()),
            (
                "LINEAR_PERCENT_R",
                data::expected_l13_d1240_d2100_d350_linear_percent_r(),
            ),
        ];
        let high = data::test_high();
        let low = data::test_low();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), high[index], low[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LINEAR_PERCENT_R", result.linear_percent_r),
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
