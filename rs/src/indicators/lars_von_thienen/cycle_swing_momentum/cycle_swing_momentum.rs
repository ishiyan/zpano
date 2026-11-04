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
use crate::indicators::core::outputs::band::Band;
/// Parameters matching the prepared reference; omitted components use entity defaults.
#[derive(Clone)]
pub struct CycleSwingMomentumParams {
    /// The window; reference default 50.
    pub window: usize,
    /// The fast smoothing; reference default 160.0.
    pub fast_smoothing: f64,
    /// The slow smoothing; reference default 1600.0.
    pub slow_smoothing: f64,
    /// The band memory; reference default 34.
    pub band_memory: usize,
    /// The leveling; reference default 10.0.
    pub leveling: f64,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for CycleSwingMomentumParams {
    fn default() -> Self {
        Self {
            window: 50,
            fast_smoothing: 160.0,
            slow_smoothing: 1600.0,
            band_memory: 34,
            leveling: 10.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CycleSwingMomentumOutput {
    Value = 1,
    Lower = 2,
    Upper = 3,
    Score = 4,
    Band = 5,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct CycleSwingMomentum {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl CycleSwingMomentum {
    /// Construct a validated streaming indicator.
    pub fn new(p: &CycleSwingMomentumParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.window as i64, "window", 6) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.fast_smoothing, "fast_smoothing", 0.0) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.slow_smoothing, "slow_smoothing", 0.0) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.band_memory as i64, "band_memory", 2) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.leveling, "leveling", 0.0) {
            return Err(error);
        }
        if p.leveling <= 0.0 || p.leveling >= 50.0 {
            return Err("leveling must be in (0,50)".into());
        }
        if p.fast_smoothing >= p.slow_smoothing {
            return Err("fast_smoothing must be smaller than slow_smoothing".into());
        }
        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let suffix = component_triple_mnemonic(bc, qc, tc);
        let parameters = [
            (p.window as i64).to_string(),
            p.fast_smoothing.to_string(),
            p.slow_smoothing.to_string(),
            (p.band_memory as i64).to_string(),
            p.leveling.to_string(),
        ]
        .join(",");
        let mnemonic = format!("csmo({parameters}{suffix})");
        let description = format!("Cycle swing momentum {mnemonic}");
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
        output.push(Box::new(Scalar::new(time, result.value)));
        output.push(Box::new(Scalar::new(time, result.lower)));
        output.push(Box::new(Scalar::new(time, result.upper)));
        output.push(Box::new(Scalar::new(time, result.score)));
        output.push(Box::new(Band::new(time, result.lower, result.upper)));
        output
    }
}
impl Indicator for CycleSwingMomentum {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CycleSwingMomentum,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} value", self.mnemonic),
                    description: format!("{} value", self.description),
                },
                OutputText {
                    mnemonic: format!("{} lower", self.mnemonic),
                    description: format!("{} lower", self.description),
                },
                OutputText {
                    mnemonic: format!("{} upper", self.mnemonic),
                    description: format!("{} upper", self.description),
                },
                OutputText {
                    mnemonic: format!("{} score", self.mnemonic),
                    description: format!("{} score", self.description),
                },
                OutputText {
                    mnemonic: format!("{} band", self.mnemonic),
                    description: format!("{} band", self.description),
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
        let p = CycleSwingMomentumParams {
            window: 50,
            fast_smoothing: 160.0,
            slow_smoothing: 1600.0,
            band_memory: 34,
            leveling: 10.0,
            ..Default::default()
        };
        let mut indicator = CycleSwingMomentum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_w50_f160_s1600_m34_q10_value()),
            ("LOWER", data::expected_w50_f160_s1600_m34_q10_lower()),
            ("UPPER", data::expected_w50_f160_s1600_m34_q10_upper()),
            ("SCORE", data::expected_w50_f160_s1600_m34_q10_score()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LOWER", result.lower),
                ("UPPER", result.upper),
                ("SCORE", result.score),
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
        let p = CycleSwingMomentumParams {
            window: 6,
            fast_smoothing: 16.0,
            slow_smoothing: 160.0,
            band_memory: 5,
            leveling: 10.0,
            ..Default::default()
        };
        let mut indicator = CycleSwingMomentum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_w6_f16_s160_m5_q10_value()),
            ("LOWER", data::expected_w6_f16_s160_m5_q10_lower()),
            ("UPPER", data::expected_w6_f16_s160_m5_q10_upper()),
            ("SCORE", data::expected_w6_f16_s160_m5_q10_score()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LOWER", result.lower),
                ("UPPER", result.upper),
                ("SCORE", result.score),
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
        let p = CycleSwingMomentumParams {
            window: 20,
            fast_smoothing: 160.0,
            slow_smoothing: 1600.0,
            band_memory: 20,
            leveling: 10.0,
            ..Default::default()
        };
        let mut indicator = CycleSwingMomentum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_w20_f160_s1600_m20_q10_value()),
            ("LOWER", data::expected_w20_f160_s1600_m20_q10_lower()),
            ("UPPER", data::expected_w20_f160_s1600_m20_q10_upper()),
            ("SCORE", data::expected_w20_f160_s1600_m20_q10_score()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LOWER", result.lower),
                ("UPPER", result.upper),
                ("SCORE", result.score),
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
        let p = CycleSwingMomentumParams {
            window: 100,
            fast_smoothing: 160.0,
            slow_smoothing: 1600.0,
            band_memory: 34,
            leveling: 10.0,
            ..Default::default()
        };
        let mut indicator = CycleSwingMomentum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_w100_f160_s1600_m34_q10_value()),
            ("LOWER", data::expected_w100_f160_s1600_m34_q10_lower()),
            ("UPPER", data::expected_w100_f160_s1600_m34_q10_upper()),
            ("SCORE", data::expected_w100_f160_s1600_m34_q10_score()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LOWER", result.lower),
                ("UPPER", result.upper),
                ("SCORE", result.score),
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
        let p = CycleSwingMomentumParams {
            window: 50,
            fast_smoothing: 0.0,
            slow_smoothing: 1600.0,
            band_memory: 34,
            leveling: 10.0,
            ..Default::default()
        };
        let mut indicator = CycleSwingMomentum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_w50_f0_s1600_m34_q10_value()),
            ("LOWER", data::expected_w50_f0_s1600_m34_q10_lower()),
            ("UPPER", data::expected_w50_f0_s1600_m34_q10_upper()),
            ("SCORE", data::expected_w50_f0_s1600_m34_q10_score()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LOWER", result.lower),
                ("UPPER", result.upper),
                ("SCORE", result.score),
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
        let p = CycleSwingMomentumParams {
            window: 50,
            fast_smoothing: 1600.0,
            slow_smoothing: 16000.0,
            band_memory: 34,
            leveling: 10.0,
            ..Default::default()
        };
        let mut indicator = CycleSwingMomentum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_w50_f1600_s16000_m34_q10_value()),
            ("LOWER", data::expected_w50_f1600_s16000_m34_q10_lower()),
            ("UPPER", data::expected_w50_f1600_s16000_m34_q10_upper()),
            ("SCORE", data::expected_w50_f1600_s16000_m34_q10_score()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LOWER", result.lower),
                ("UPPER", result.upper),
                ("SCORE", result.score),
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
        let p = CycleSwingMomentumParams {
            window: 50,
            fast_smoothing: 160.0,
            slow_smoothing: 1600.0,
            band_memory: 34,
            leveling: 5.0,
            ..Default::default()
        };
        let mut indicator = CycleSwingMomentum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_w50_f160_s1600_m34_q5_value()),
            ("LOWER", data::expected_w50_f160_s1600_m34_q5_lower()),
            ("UPPER", data::expected_w50_f160_s1600_m34_q5_upper()),
            ("SCORE", data::expected_w50_f160_s1600_m34_q5_score()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LOWER", result.lower),
                ("UPPER", result.upper),
                ("SCORE", result.score),
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
        let p = CycleSwingMomentumParams {
            window: 252,
            fast_smoothing: 160.0,
            slow_smoothing: 1600.0,
            band_memory: 34,
            leveling: 10.0,
            ..Default::default()
        };
        let mut indicator = CycleSwingMomentum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_w252_f160_s1600_m34_q10_value()),
            ("LOWER", data::expected_w252_f160_s1600_m34_q10_lower()),
            ("UPPER", data::expected_w252_f160_s1600_m34_q10_upper()),
            ("SCORE", data::expected_w252_f160_s1600_m34_q10_score()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("LOWER", result.lower),
                ("UPPER", result.upper),
                ("SCORE", result.score),
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
            let p = CycleSwingMomentumParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = CycleSwingMomentum::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
