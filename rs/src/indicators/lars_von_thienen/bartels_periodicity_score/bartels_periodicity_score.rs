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
pub struct BartelsPeriodicityScoreParams {
    /// The period; reference default 20.
    pub period: usize,
    /// The segments; reference default 5.
    pub segments: usize,
    /// The smoothing; reference default 0.0.
    pub smoothing: f64,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for BartelsPeriodicityScoreParams {
    fn default() -> Self {
        Self {
            period: 20,
            segments: 5,
            smoothing: 0.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum BartelsPeriodicityScoreOutput {
    Score = 1,
    GenuinePercent = 2,
    Segments = 3,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct BartelsPeriodicityScore {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl BartelsPeriodicityScore {
    /// Construct a validated streaming indicator.
    pub fn new(p: &BartelsPeriodicityScoreParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.period as i64, "period", 3) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.segments as i64, "segments", 2) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.smoothing, "smoothing", 0.0) {
            return Err(error);
        }
        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let suffix = component_triple_mnemonic(bc, qc, tc);
        let parameters = [
            (p.period as i64).to_string(),
            (p.segments as i64).to_string(),
            p.smoothing.to_string(),
        ]
        .join(",");
        let mnemonic = format!("bps({parameters}{suffix})");
        let description = format!("Bartels periodicity score {mnemonic}");
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
        output.push(Box::new(Scalar::new(time, result.score)));
        output.push(Box::new(Scalar::new(time, result.genuine_percent)));
        output.push(Box::new(Scalar::new(time, result.segments)));
        output
    }
}
impl Indicator for BartelsPeriodicityScore {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::BartelsPeriodicityScore,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} score", self.mnemonic),
                    description: format!("{} score", self.description),
                },
                OutputText {
                    mnemonic: format!("{} genuine_percent", self.mnemonic),
                    description: format!("{} genuine_percent", self.description),
                },
                OutputText {
                    mnemonic: format!("{} segments", self.mnemonic),
                    description: format!("{} segments", self.description),
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
        let p = BartelsPeriodicityScoreParams {
            period: 20,
            segments: 5,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = BartelsPeriodicityScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("SCORE", data::expected_p20_m5_l0_score()),
            (
                "GENUINE_PERCENT",
                data::expected_p20_m5_l0_genuine_percent(),
            ),
            ("SEGMENTS", data::expected_p20_m5_l0_segments()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("SCORE", result.score),
                ("GENUINE_PERCENT", result.genuine_percent),
                ("SEGMENTS", result.segments),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-10_f64.max(values[index].abs() * 5e-15),
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
        let p = BartelsPeriodicityScoreParams {
            period: 3,
            segments: 2,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = BartelsPeriodicityScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("SCORE", data::expected_p3_m2_l0_score()),
            ("GENUINE_PERCENT", data::expected_p3_m2_l0_genuine_percent()),
            ("SEGMENTS", data::expected_p3_m2_l0_segments()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("SCORE", result.score),
                ("GENUINE_PERCENT", result.genuine_percent),
                ("SEGMENTS", result.segments),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-10_f64.max(values[index].abs() * 5e-15),
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
        let p = BartelsPeriodicityScoreParams {
            period: 10,
            segments: 5,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = BartelsPeriodicityScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("SCORE", data::expected_p10_m5_l0_score()),
            (
                "GENUINE_PERCENT",
                data::expected_p10_m5_l0_genuine_percent(),
            ),
            ("SEGMENTS", data::expected_p10_m5_l0_segments()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("SCORE", result.score),
                ("GENUINE_PERCENT", result.genuine_percent),
                ("SEGMENTS", result.segments),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-10_f64.max(values[index].abs() * 5e-15),
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
        let p = BartelsPeriodicityScoreParams {
            period: 20,
            segments: 2,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = BartelsPeriodicityScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("SCORE", data::expected_p20_m2_l0_score()),
            (
                "GENUINE_PERCENT",
                data::expected_p20_m2_l0_genuine_percent(),
            ),
            ("SEGMENTS", data::expected_p20_m2_l0_segments()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("SCORE", result.score),
                ("GENUINE_PERCENT", result.genuine_percent),
                ("SEGMENTS", result.segments),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-10_f64.max(values[index].abs() * 5e-15),
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
        let p = BartelsPeriodicityScoreParams {
            period: 20,
            segments: 5,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = BartelsPeriodicityScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("SCORE", data::expected_p20_m5_l1600_score()),
            (
                "GENUINE_PERCENT",
                data::expected_p20_m5_l1600_genuine_percent(),
            ),
            ("SEGMENTS", data::expected_p20_m5_l1600_segments()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("SCORE", result.score),
                ("GENUINE_PERCENT", result.genuine_percent),
                ("SEGMENTS", result.segments),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-10_f64.max(values[index].abs() * 5e-15),
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
        let p = BartelsPeriodicityScoreParams {
            period: 40,
            segments: 5,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = BartelsPeriodicityScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("SCORE", data::expected_p40_m5_l0_score()),
            (
                "GENUINE_PERCENT",
                data::expected_p40_m5_l0_genuine_percent(),
            ),
            ("SEGMENTS", data::expected_p40_m5_l0_segments()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("SCORE", result.score),
                ("GENUINE_PERCENT", result.genuine_percent),
                ("SEGMENTS", result.segments),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-10_f64.max(values[index].abs() * 5e-15),
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
        let p = BartelsPeriodicityScoreParams {
            period: 42,
            segments: 6,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = BartelsPeriodicityScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("SCORE", data::expected_p42_m6_l0_score()),
            (
                "GENUINE_PERCENT",
                data::expected_p42_m6_l0_genuine_percent(),
            ),
            ("SEGMENTS", data::expected_p42_m6_l0_segments()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("SCORE", result.score),
                ("GENUINE_PERCENT", result.genuine_percent),
                ("SEGMENTS", result.segments),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-10_f64.max(values[index].abs() * 5e-15),
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
        let p = BartelsPeriodicityScoreParams {
            period: 60,
            segments: 5,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = BartelsPeriodicityScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("SCORE", data::expected_p60_m5_l0_score()),
            (
                "GENUINE_PERCENT",
                data::expected_p60_m5_l0_genuine_percent(),
            ),
            ("SEGMENTS", data::expected_p60_m5_l0_segments()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("SCORE", result.score),
                ("GENUINE_PERCENT", result.genuine_percent),
                ("SEGMENTS", result.segments),
            ]);
            for (field, values) in &expected {
                let value = actual[field];
                if values[index].is_nan() {
                    assert!(value.is_nan(), "bar {index} {field}");
                } else {
                    assert!(
                        !value.is_nan()
                            && (value - values[index]).abs()
                                <= 1e-10_f64.max(values[index].abs() * 5e-15),
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
            let p = BartelsPeriodicityScoreParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = BartelsPeriodicityScore::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
