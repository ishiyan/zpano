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
pub struct CycleStrengthParams {
    /// The window; reference default 100.
    pub window: usize,
    /// The period; reference default 20.0.
    pub period: f64,
    /// The smoothing; reference default 1600.0.
    pub smoothing: f64,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for CycleStrengthParams {
    fn default() -> Self {
        Self {
            window: 100,
            period: 20.0,
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
pub enum CycleStrengthOutput {
    Strength = 1,
    Amplitude = 2,
    Phase = 3,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct CycleStrength {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl CycleStrength {
    /// Construct a validated streaming indicator.
    pub fn new(p: &CycleStrengthParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.window as i64, "window", 6) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.period, "period", 2.0000000000000004) {
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
            (p.window as i64).to_string(),
            p.period.to_string(),
            p.smoothing.to_string(),
        ]
        .join(",");
        let mnemonic = format!("cstr({parameters}{suffix})");
        let description = format!("Cycle strength {mnemonic}");
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
        output.push(Box::new(Scalar::new(time, result.strength)));
        output.push(Box::new(Scalar::new(time, result.amplitude)));
        output.push(Box::new(Scalar::new(time, result.phase)));
        output
    }
}
impl Indicator for CycleStrength {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CycleStrength,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} strength", self.mnemonic),
                    description: format!("{} strength", self.description),
                },
                OutputText {
                    mnemonic: format!("{} amplitude", self.mnemonic),
                    description: format!("{} amplitude", self.description),
                },
                OutputText {
                    mnemonic: format!("{} phase", self.mnemonic),
                    description: format!("{} phase", self.description),
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
        let p = CycleStrengthParams {
            window: 100,
            period: 20.0,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = CycleStrength::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("STRENGTH", data::expected_w100_p20_l1600_strength()),
            ("AMPLITUDE", data::expected_w100_p20_l1600_amplitude()),
            ("PHASE", data::expected_w100_p20_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("STRENGTH", result.strength),
                ("AMPLITUDE", result.amplitude),
                ("PHASE", result.phase),
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
        let p = CycleStrengthParams {
            window: 100,
            period: 20.0,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = CycleStrength::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("STRENGTH", data::expected_w100_p20_l0_strength()),
            ("AMPLITUDE", data::expected_w100_p20_l0_amplitude()),
            ("PHASE", data::expected_w100_p20_l0_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("STRENGTH", result.strength),
                ("AMPLITUDE", result.amplitude),
                ("PHASE", result.phase),
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
        let p = CycleStrengthParams {
            window: 60,
            period: 10.0,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = CycleStrength::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("STRENGTH", data::expected_w60_p10_l1600_strength()),
            ("AMPLITUDE", data::expected_w60_p10_l1600_amplitude()),
            ("PHASE", data::expected_w60_p10_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("STRENGTH", result.strength),
                ("AMPLITUDE", result.amplitude),
                ("PHASE", result.phase),
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
        let p = CycleStrengthParams {
            window: 200,
            period: 40.0,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = CycleStrength::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("STRENGTH", data::expected_w200_p40_l1600_strength()),
            ("AMPLITUDE", data::expected_w200_p40_l1600_amplitude()),
            ("PHASE", data::expected_w200_p40_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("STRENGTH", result.strength),
                ("AMPLITUDE", result.amplitude),
                ("PHASE", result.phase),
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
        let p = CycleStrengthParams {
            window: 100,
            period: 20.5,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = CycleStrength::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("STRENGTH", data::expected_w100_p20_5_l1600_strength()),
            ("AMPLITUDE", data::expected_w100_p20_5_l1600_amplitude()),
            ("PHASE", data::expected_w100_p20_5_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("STRENGTH", result.strength),
                ("AMPLITUDE", result.amplitude),
                ("PHASE", result.phase),
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
        let p = CycleStrengthParams {
            window: 100,
            period: 20.0,
            smoothing: 16000.0,
            ..Default::default()
        };
        let mut indicator = CycleStrength::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("STRENGTH", data::expected_w100_p20_l16000_strength()),
            ("AMPLITUDE", data::expected_w100_p20_l16000_amplitude()),
            ("PHASE", data::expected_w100_p20_l16000_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("STRENGTH", result.strength),
                ("AMPLITUDE", result.amplitude),
                ("PHASE", result.phase),
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
        let p = CycleStrengthParams {
            window: 20,
            period: 40.0,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = CycleStrength::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("STRENGTH", data::expected_w20_p40_l1600_strength()),
            ("AMPLITUDE", data::expected_w20_p40_l1600_amplitude()),
            ("PHASE", data::expected_w20_p40_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("STRENGTH", result.strength),
                ("AMPLITUDE", result.amplitude),
                ("PHASE", result.phase),
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
        let p = CycleStrengthParams {
            window: 252,
            period: 20.0,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = CycleStrength::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("STRENGTH", data::expected_w252_p20_l1600_strength()),
            ("AMPLITUDE", data::expected_w252_p20_l1600_amplitude()),
            ("PHASE", data::expected_w252_p20_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("STRENGTH", result.strength),
                ("AMPLITUDE", result.amplitude),
                ("PHASE", result.phase),
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
            let p = CycleStrengthParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = CycleStrength::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
