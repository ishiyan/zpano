use super::super::harmonic_cycle::validation::{validate_integer, validate_real};
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
pub struct HodrickPrescottSmoothingAndDetrendingParams {
    /// The window; reference default 50.
    pub window: usize,
    /// The smoothing; reference default 1600.0.
    pub smoothing: f64,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for HodrickPrescottSmoothingAndDetrendingParams {
    fn default() -> Self {
        Self {
            window: 50,
            smoothing: 1600.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum HodrickPrescottSmoothingAndDetrendingOutput {
    Trend = 1,
    Cycle = 2,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct HodrickPrescottSmoothingAndDetrending {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl HodrickPrescottSmoothingAndDetrending {
    /// Construct a validated streaming indicator.
    pub fn new(p: &HodrickPrescottSmoothingAndDetrendingParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.window as i64, "window", 6) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.smoothing, "smoothing", 0.0) {
            return Err(error);
        }
        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let suffix = component_triple_mnemonic(bc, qc, tc);
        let parameters = [(p.window as i64).to_string(), p.smoothing.to_string()].join(",");
        let mnemonic = format!("hpsd({parameters}{suffix})");
        let description = format!("Hodrick prescott smoothing and detrending {mnemonic}");
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
    pub fn update(&mut self, price: f64) -> Snapshot {
        self.engine.update(price)
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
        output.push(Box::new(Scalar::new(time, result.trend)));
        output.push(Box::new(Scalar::new(time, result.cycle)));
        output
    }
}
impl Indicator for HodrickPrescottSmoothingAndDetrending {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::HodrickPrescottSmoothingAndDetrending,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} trend", self.mnemonic),
                    description: format!("{} trend", self.description),
                },
                OutputText {
                    mnemonic: format!("{} cycle", self.mnemonic),
                    description: format!("{} cycle", self.description),
                },
            ],
        )
    }
    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        self.wrap(sample.time, value)
    }
    fn update_bar(&mut self, sample: &Bar) -> Output {
        let value = self.update((self.bar_func)(sample));
        self.wrap(sample.time, value)
    }
    fn update_quote(&mut self, sample: &Quote) -> Output {
        let value = self.update((self.quote_func)(sample));
        self.wrap(sample.time, value)
    }
    fn update_trade(&mut self, sample: &Trade) -> Output {
        let value = self.update((self.trade_func)(sample));
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
    fn prepared_case_0() {
        let p = HodrickPrescottSmoothingAndDetrendingParams {
            window: 50,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = HodrickPrescottSmoothingAndDetrending::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("TREND", data::expected_w50_l1600_trend()),
            ("CYCLE", data::expected_w50_l1600_cycle()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual =
                std::collections::HashMap::from([("TREND", result.trend), ("CYCLE", result.cycle)]);
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
        let p = HodrickPrescottSmoothingAndDetrendingParams {
            window: 6,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = HodrickPrescottSmoothingAndDetrending::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("TREND", data::expected_w6_l1600_trend()),
            ("CYCLE", data::expected_w6_l1600_cycle()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual =
                std::collections::HashMap::from([("TREND", result.trend), ("CYCLE", result.cycle)]);
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
        let p = HodrickPrescottSmoothingAndDetrendingParams {
            window: 20,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = HodrickPrescottSmoothingAndDetrending::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("TREND", data::expected_w20_l0_trend()),
            ("CYCLE", data::expected_w20_l0_cycle()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual =
                std::collections::HashMap::from([("TREND", result.trend), ("CYCLE", result.cycle)]);
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
        let p = HodrickPrescottSmoothingAndDetrendingParams {
            window: 20,
            smoothing: 1.0,
            ..Default::default()
        };
        let mut indicator = HodrickPrescottSmoothingAndDetrending::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("TREND", data::expected_w20_l1_trend()),
            ("CYCLE", data::expected_w20_l1_cycle()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual =
                std::collections::HashMap::from([("TREND", result.trend), ("CYCLE", result.cycle)]);
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
        let p = HodrickPrescottSmoothingAndDetrendingParams {
            window: 50,
            smoothing: 160.0,
            ..Default::default()
        };
        let mut indicator = HodrickPrescottSmoothingAndDetrending::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("TREND", data::expected_w50_l160_trend()),
            ("CYCLE", data::expected_w50_l160_cycle()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual =
                std::collections::HashMap::from([("TREND", result.trend), ("CYCLE", result.cycle)]);
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
        let p = HodrickPrescottSmoothingAndDetrendingParams {
            window: 100,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = HodrickPrescottSmoothingAndDetrending::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("TREND", data::expected_w100_l1600_trend()),
            ("CYCLE", data::expected_w100_l1600_cycle()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual =
                std::collections::HashMap::from([("TREND", result.trend), ("CYCLE", result.cycle)]);
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
        let p = HodrickPrescottSmoothingAndDetrendingParams {
            window: 100,
            smoothing: 100000.0,
            ..Default::default()
        };
        let mut indicator = HodrickPrescottSmoothingAndDetrending::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("TREND", data::expected_w100_l100000_trend()),
            ("CYCLE", data::expected_w100_l100000_cycle()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual =
                std::collections::HashMap::from([("TREND", result.trend), ("CYCLE", result.cycle)]);
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
        let p = HodrickPrescottSmoothingAndDetrendingParams {
            window: 252,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = HodrickPrescottSmoothingAndDetrending::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("TREND", data::expected_w252_l1600_trend()),
            ("CYCLE", data::expected_w252_l1600_cycle()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual =
                std::collections::HashMap::from([("TREND", result.trend), ("CYCLE", result.cycle)]);
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
        let p = HodrickPrescottSmoothingAndDetrendingParams {
            window: 253,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = HodrickPrescottSmoothingAndDetrending::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("TREND", data::expected_w253_l1600_trend()),
            ("CYCLE", data::expected_w253_l1600_cycle()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual =
                std::collections::HashMap::from([("TREND", result.trend), ("CYCLE", result.cycle)]);
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
            let p = HodrickPrescottSmoothingAndDetrendingParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = HodrickPrescottSmoothingAndDetrending::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
