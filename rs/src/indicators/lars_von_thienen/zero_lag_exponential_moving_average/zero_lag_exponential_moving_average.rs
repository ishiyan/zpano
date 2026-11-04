use super::super::harmonic_cycle::validation::validate_integer;
use super::engine::Engine;
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
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;
/// Parameters matching the prepared reference; omitted components use entity defaults.
#[derive(Clone)]
pub struct LarsVonThienenZeroLagExponentialMovingAverageParams {
    /// The length; reference default 13.
    pub length: usize,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for LarsVonThienenZeroLagExponentialMovingAverageParams {
    fn default() -> Self {
        Self {
            length: 13,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum LarsVonThienenZeroLagExponentialMovingAverageOutput {
    Value = 1,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct LarsVonThienenZeroLagExponentialMovingAverage {
    engine: Engine,
    line: LineIndicator,
}
impl LarsVonThienenZeroLagExponentialMovingAverage {
    /// Construct a validated streaming indicator.
    pub fn new(p: &LarsVonThienenZeroLagExponentialMovingAverageParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.length as i64, "length", 1) {
            return Err(error);
        }
        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let suffix = component_triple_mnemonic(bc, qc, tc);
        let parameters = [(p.length as i64).to_string()].join(",");
        let mnemonic = format!("lzlema({parameters}{suffix})");
        let description = format!("Zero lag exponential moving average {mnemonic}");
        let line = LineIndicator::new(
            mnemonic,
            description,
            bar_component_value(bc),
            quote_component_value(qc),
            trade_component_value(tc),
        );
        let mut engine = Engine {
            p: p.clone(),
            ..Default::default()
        };
        engine.reset();
        Ok(Self { engine, line })
    }
    /// Accept one completed observation.
    pub fn update(&mut self, price: f64) -> f64 {
        self.engine.update(price).value
    }
    /// Clear history and retain parameters.
    pub fn reset(&mut self) {
        self.engine.reset();
    }
    /// Minimum initial observations before base readiness.
    pub fn warmup_bars(&self) -> i64 {
        self.engine.warmup()
    }
}
impl Indicator for LarsVonThienenZeroLagExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::LarsVonThienenZeroLagExponentialMovingAverage,
            &self.line.mnemonic,
            &self.line.description,
            &[OutputText {
                mnemonic: format!("{} value", self.line.mnemonic),
                description: format!("{} value", self.line.description),
            }],
        )
    }
    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }
    fn update_bar(&mut self, sample: &Bar) -> Output {
        let value = self.update((self.line.bar_func)(sample));
        vec![Box::new(Scalar::new(sample.time, value))]
    }
    fn update_quote(&mut self, sample: &Quote) -> Output {
        let value = self.update((self.line.quote_func)(sample));
        vec![Box::new(Scalar::new(sample.time, value))]
    }
    fn update_trade(&mut self, sample: &Trade) -> Output {
        let value = self.update((self.line.trade_func)(sample));
        vec![Box::new(Scalar::new(sample.time, value))]
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
        let p = LarsVonThienenZeroLagExponentialMovingAverageParams {
            length: 13,
            ..Default::default()
        };
        let mut indicator = LarsVonThienenZeroLagExponentialMovingAverage::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![("VALUE", data::expected_l13_value())];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([("VALUE", result)]);
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
        let p = LarsVonThienenZeroLagExponentialMovingAverageParams {
            length: 1,
            ..Default::default()
        };
        let mut indicator = LarsVonThienenZeroLagExponentialMovingAverage::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![("VALUE", data::expected_l1_value())];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([("VALUE", result)]);
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
        let p = LarsVonThienenZeroLagExponentialMovingAverageParams {
            length: 2,
            ..Default::default()
        };
        let mut indicator = LarsVonThienenZeroLagExponentialMovingAverage::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![("VALUE", data::expected_l2_value())];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([("VALUE", result)]);
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
        let p = LarsVonThienenZeroLagExponentialMovingAverageParams {
            length: 3,
            ..Default::default()
        };
        let mut indicator = LarsVonThienenZeroLagExponentialMovingAverage::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![("VALUE", data::expected_l3_value())];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([("VALUE", result)]);
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
        let p = LarsVonThienenZeroLagExponentialMovingAverageParams {
            length: 10,
            ..Default::default()
        };
        let mut indicator = LarsVonThienenZeroLagExponentialMovingAverage::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![("VALUE", data::expected_l10_value())];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([("VALUE", result)]);
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
        let p = LarsVonThienenZeroLagExponentialMovingAverageParams {
            length: 20,
            ..Default::default()
        };
        let mut indicator = LarsVonThienenZeroLagExponentialMovingAverage::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![("VALUE", data::expected_l20_value())];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([("VALUE", result)]);
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
        let p = LarsVonThienenZeroLagExponentialMovingAverageParams {
            length: 50,
            ..Default::default()
        };
        let mut indicator = LarsVonThienenZeroLagExponentialMovingAverage::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![("VALUE", data::expected_l50_value())];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([("VALUE", result)]);
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
        let p = LarsVonThienenZeroLagExponentialMovingAverageParams {
            length: 252,
            ..Default::default()
        };
        let mut indicator = LarsVonThienenZeroLagExponentialMovingAverage::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![("VALUE", data::expected_l252_value())];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([("VALUE", result)]);
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
        let p = LarsVonThienenZeroLagExponentialMovingAverageParams {
            length: 600,
            ..Default::default()
        };
        let mut indicator = LarsVonThienenZeroLagExponentialMovingAverage::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![("VALUE", data::expected_l600_value())];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([("VALUE", result)]);
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
            let p = LarsVonThienenZeroLagExponentialMovingAverageParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = LarsVonThienenZeroLagExponentialMovingAverage::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
