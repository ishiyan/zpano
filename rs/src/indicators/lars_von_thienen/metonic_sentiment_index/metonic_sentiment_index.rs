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
pub struct MetonicSentimentIndexParams {
    /// The length; reference default 13.
    pub length: usize,
    /// The delay1; reference default 6939.
    pub delay1: usize,
    /// The delay2; reference default 4016.
    pub delay2: usize,
    /// The delay3; reference default 2922.
    pub delay3: usize,
    /// The mode; reference default 0.
    pub mode: usize,
}
impl Default for MetonicSentimentIndexParams {
    fn default() -> Self {
        Self {
            length: 13,
            delay1: 6939,
            delay2: 4016,
            delay3: 2922,
            mode: 0,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MetonicSentimentIndexOutput {
    Value = 1,
    DailySentiment = 2,
    SmoothedClose = 3,
    SmoothedOpen = 4,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct MetonicSentimentIndex {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl MetonicSentimentIndex {
    /// Construct a validated streaming indicator.
    pub fn new(p: &MetonicSentimentIndexParams) -> Result<Self, String> {
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
        if let Some(error) = validate_integer(p.mode as i64, "mode", 0) {
            return Err(error);
        }
        if (p.delay1 as i64) == (p.delay2 as i64)
            || (p.delay1 as i64) == (p.delay3 as i64)
            || (p.delay2 as i64) == (p.delay3 as i64)
        {
            return Err("three delays must be distinct".into());
        }
        if (p.mode as i64) > 1 {
            return Err("mode must be 0 or 1".into());
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
            (p.mode as i64).to_string(),
        ]
        .join(",");
        let mnemonic = format!("msent({parameters}{suffix})");
        let description = format!("Metonic sentiment index {mnemonic}");
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
    pub fn update(&mut self, price: f64, day: Option<i64>, open_price: f64) -> Snapshot {
        self.engine.update(price, day, open_price)
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
        output.push(Box::new(Scalar::new(time, result.daily_sentiment)));
        output.push(Box::new(Scalar::new(time, result.smoothed_close)));
        output.push(Box::new(Scalar::new(time, result.smoothed_open)));
        output
    }
}
impl Indicator for MetonicSentimentIndex {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::MetonicSentimentIndex,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} value", self.mnemonic),
                    description: format!("{} value", self.description),
                },
                OutputText {
                    mnemonic: format!("{} daily_sentiment", self.mnemonic),
                    description: format!("{} daily_sentiment", self.description),
                },
                OutputText {
                    mnemonic: format!("{} smoothed_close", self.mnemonic),
                    description: format!("{} smoothed_close", self.description),
                },
                OutputText {
                    mnemonic: format!("{} smoothed_open", self.mnemonic),
                    description: format!("{} smoothed_open", self.description),
                },
            ],
        )
    }
    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        assert_eq!(
            (self.engine.p.mode as i64),
            1,
            "original sentiment requires OHLC"
        );
        let result = self.update(sample.value, Some(super::super::metonic::calendar::day_from_unix_seconds(sample.time)), f64::NAN);
        self.wrap(sample.time, result)
    }
    fn update_bar(&mut self, sample: &Bar) -> Output {
        let value = self.update(
            sample.close,
            Some(super::super::metonic::calendar::day_from_unix_seconds(sample.time)),
            sample.open,
        );
        self.wrap(sample.time, value)
    }
    fn update_quote(&mut self, sample: &Quote) -> Output {
        assert_eq!(
            (self.engine.p.mode as i64),
            1,
            "original sentiment requires OHLC"
        );
        let result = self.update(sample.mid(), Some(super::super::metonic::calendar::day_from_unix_seconds(sample.time)), f64::NAN);
        self.wrap(sample.time, result)
    }
    fn update_trade(&mut self, sample: &Trade) -> Output {
        assert_eq!(
            (self.engine.p.mode as i64),
            1,
            "original sentiment requires OHLC"
        );
        let result = self.update(sample.price, Some(super::super::metonic::calendar::day_from_unix_seconds(sample.time)), f64::NAN);
        self.wrap(sample.time, result)
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
        let p = MetonicSentimentIndexParams {
            length: 13,
            delay1: 6939,
            delay2: 4016,
            delay3: 2922,
            mode: 0,
            ..Default::default()
        };
        let mut indicator = MetonicSentimentIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d16939_d24016_d32922_m0_value()),
            (
                "DAILY_SENTIMENT",
                data::expected_l13_d16939_d24016_d32922_m0_daily_sentiment(),
            ),
            (
                "SMOOTHED_CLOSE",
                data::expected_l13_d16939_d24016_d32922_m0_smoothed_close(),
            ),
            (
                "SMOOTHED_OPEN",
                data::expected_l13_d16939_d24016_d32922_m0_smoothed_open(),
            ),
        ];
        let open = data::test_open();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), open[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DAILY_SENTIMENT", result.daily_sentiment),
                ("SMOOTHED_CLOSE", result.smoothed_close),
                ("SMOOTHED_OPEN", result.smoothed_open),
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
        let p = MetonicSentimentIndexParams {
            length: 13,
            delay1: 6939,
            delay2: 4016,
            delay3: 2922,
            mode: 1,
            ..Default::default()
        };
        let mut indicator = MetonicSentimentIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d16939_d24016_d32922_m1_value()),
            (
                "DAILY_SENTIMENT",
                data::expected_l13_d16939_d24016_d32922_m1_daily_sentiment(),
            ),
            (
                "SMOOTHED_CLOSE",
                data::expected_l13_d16939_d24016_d32922_m1_smoothed_close(),
            ),
            (
                "SMOOTHED_OPEN",
                data::expected_l13_d16939_d24016_d32922_m1_smoothed_open(),
            ),
        ];
        let open = data::test_open();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), open[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DAILY_SENTIMENT", result.daily_sentiment),
                ("SMOOTHED_CLOSE", result.smoothed_close),
                ("SMOOTHED_OPEN", result.smoothed_open),
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
        let p = MetonicSentimentIndexParams {
            length: 13,
            delay1: 19,
            delay2: 11,
            delay3: 8,
            mode: 0,
            ..Default::default()
        };
        let mut indicator = MetonicSentimentIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d119_d211_d38_m0_value()),
            (
                "DAILY_SENTIMENT",
                data::expected_l13_d119_d211_d38_m0_daily_sentiment(),
            ),
            (
                "SMOOTHED_CLOSE",
                data::expected_l13_d119_d211_d38_m0_smoothed_close(),
            ),
            (
                "SMOOTHED_OPEN",
                data::expected_l13_d119_d211_d38_m0_smoothed_open(),
            ),
        ];
        let open = data::test_open();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), open[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DAILY_SENTIMENT", result.daily_sentiment),
                ("SMOOTHED_CLOSE", result.smoothed_close),
                ("SMOOTHED_OPEN", result.smoothed_open),
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
        let p = MetonicSentimentIndexParams {
            length: 13,
            delay1: 19,
            delay2: 11,
            delay3: 8,
            mode: 1,
            ..Default::default()
        };
        let mut indicator = MetonicSentimentIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d119_d211_d38_m1_value()),
            (
                "DAILY_SENTIMENT",
                data::expected_l13_d119_d211_d38_m1_daily_sentiment(),
            ),
            (
                "SMOOTHED_CLOSE",
                data::expected_l13_d119_d211_d38_m1_smoothed_close(),
            ),
            (
                "SMOOTHED_OPEN",
                data::expected_l13_d119_d211_d38_m1_smoothed_open(),
            ),
        ];
        let open = data::test_open();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), open[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DAILY_SENTIMENT", result.daily_sentiment),
                ("SMOOTHED_CLOSE", result.smoothed_close),
                ("SMOOTHED_OPEN", result.smoothed_open),
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
        let p = MetonicSentimentIndexParams {
            length: 1,
            delay1: 19,
            delay2: 11,
            delay3: 8,
            mode: 0,
            ..Default::default()
        };
        let mut indicator = MetonicSentimentIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l1_d119_d211_d38_m0_value()),
            (
                "DAILY_SENTIMENT",
                data::expected_l1_d119_d211_d38_m0_daily_sentiment(),
            ),
            (
                "SMOOTHED_CLOSE",
                data::expected_l1_d119_d211_d38_m0_smoothed_close(),
            ),
            (
                "SMOOTHED_OPEN",
                data::expected_l1_d119_d211_d38_m0_smoothed_open(),
            ),
        ];
        let open = data::test_open();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), open[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DAILY_SENTIMENT", result.daily_sentiment),
                ("SMOOTHED_CLOSE", result.smoothed_close),
                ("SMOOTHED_OPEN", result.smoothed_open),
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
        let p = MetonicSentimentIndexParams {
            length: 10,
            delay1: 19,
            delay2: 11,
            delay3: 8,
            mode: 1,
            ..Default::default()
        };
        let mut indicator = MetonicSentimentIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l10_d119_d211_d38_m1_value()),
            (
                "DAILY_SENTIMENT",
                data::expected_l10_d119_d211_d38_m1_daily_sentiment(),
            ),
            (
                "SMOOTHED_CLOSE",
                data::expected_l10_d119_d211_d38_m1_smoothed_close(),
            ),
            (
                "SMOOTHED_OPEN",
                data::expected_l10_d119_d211_d38_m1_smoothed_open(),
            ),
        ];
        let open = data::test_open();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), open[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DAILY_SENTIMENT", result.daily_sentiment),
                ("SMOOTHED_CLOSE", result.smoothed_close),
                ("SMOOTHED_OPEN", result.smoothed_open),
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
        let p = MetonicSentimentIndexParams {
            length: 20,
            delay1: 100,
            delay2: 60,
            delay3: 40,
            mode: 0,
            ..Default::default()
        };
        let mut indicator = MetonicSentimentIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l20_d1100_d260_d340_m0_value()),
            (
                "DAILY_SENTIMENT",
                data::expected_l20_d1100_d260_d340_m0_daily_sentiment(),
            ),
            (
                "SMOOTHED_CLOSE",
                data::expected_l20_d1100_d260_d340_m0_smoothed_close(),
            ),
            (
                "SMOOTHED_OPEN",
                data::expected_l20_d1100_d260_d340_m0_smoothed_open(),
            ),
        ];
        let open = data::test_open();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), open[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DAILY_SENTIMENT", result.daily_sentiment),
                ("SMOOTHED_CLOSE", result.smoothed_close),
                ("SMOOTHED_OPEN", result.smoothed_open),
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
        let p = MetonicSentimentIndexParams {
            length: 13,
            delay1: 8,
            delay2: 19,
            delay3: 11,
            mode: 1,
            ..Default::default()
        };
        let mut indicator = MetonicSentimentIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l13_d18_d219_d311_m1_value()),
            (
                "DAILY_SENTIMENT",
                data::expected_l13_d18_d219_d311_m1_daily_sentiment(),
            ),
            (
                "SMOOTHED_CLOSE",
                data::expected_l13_d18_d219_d311_m1_smoothed_close(),
            ),
            (
                "SMOOTHED_OPEN",
                data::expected_l13_d18_d219_d311_m1_smoothed_open(),
            ),
        ];
        let open = data::test_open();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), open[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DAILY_SENTIMENT", result.daily_sentiment),
                ("SMOOTHED_CLOSE", result.smoothed_close),
                ("SMOOTHED_OPEN", result.smoothed_open),
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
    fn prepared_case_8() {
        let p = MetonicSentimentIndexParams {
            length: 1,
            delay1: 251,
            delay2: 100,
            delay3: 50,
            mode: 0,
            ..Default::default()
        };
        let mut indicator = MetonicSentimentIndex::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_l1_d1251_d2100_d350_m0_value()),
            (
                "DAILY_SENTIMENT",
                data::expected_l1_d1251_d2100_d350_m0_daily_sentiment(),
            ),
            (
                "SMOOTHED_CLOSE",
                data::expected_l1_d1251_d2100_d350_m0_smoothed_close(),
            ),
            (
                "SMOOTHED_OPEN",
                data::expected_l1_d1251_d2100_d350_m0_smoothed_open(),
            ),
        ];
        let open = data::test_open();
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price, Some(index as i64), open[index]);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DAILY_SENTIMENT", result.daily_sentiment),
                ("SMOOTHED_CLOSE", result.smoothed_close),
                ("SMOOTHED_OPEN", result.smoothed_open),
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
