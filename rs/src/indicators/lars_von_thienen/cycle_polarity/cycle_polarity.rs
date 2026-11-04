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
pub struct CyclePolarityParams {
    /// The window; reference default 100.
    pub window: usize,
    /// The short period; reference default 20.0.
    pub short_period: f64,
    /// The long period; reference default 40.0.
    pub long_period: f64,
    /// The smoothing; reference default 1600.0.
    pub smoothing: f64,
    /// The tolerance; reference default 0.0.
    pub tolerance: f64,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for CyclePolarityParams {
    fn default() -> Self {
        Self {
            window: 100,
            short_period: 20.0,
            long_period: 40.0,
            smoothing: 1600.0,
            tolerance: 0.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CyclePolarityOutput {
    Polarity = 1,
    ShortValue = 2,
    LongValue = 3,
    ShortSlope = 4,
    LongSlope = 5,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct CyclePolarity {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl CyclePolarity {
    /// Construct a validated streaming indicator.
    pub fn new(p: &CyclePolarityParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.window as i64, "window", 6) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.short_period, "short_period", 2.0000000000000004) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.long_period, "long_period", 2.0000000000000004) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.smoothing, "smoothing", 0.0) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.tolerance, "tolerance", 0.0) {
            return Err(error);
        }
        if p.short_period >= p.long_period {
            return Err("short_period must be smaller than long_period".into());
        }
        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let suffix = component_triple_mnemonic(bc, qc, tc);
        let parameters = [
            (p.window as i64).to_string(),
            p.short_period.to_string(),
            p.long_period.to_string(),
            p.smoothing.to_string(),
            p.tolerance.to_string(),
        ]
        .join(",");
        let mnemonic = format!("cpol({parameters}{suffix})");
        let description = format!("Cycle polarity {mnemonic}");
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
        output.push(Box::new(Scalar::new(time, result.polarity)));
        output.push(Box::new(Scalar::new(time, result.short_value)));
        output.push(Box::new(Scalar::new(time, result.long_value)));
        output.push(Box::new(Scalar::new(time, result.short_slope)));
        output.push(Box::new(Scalar::new(time, result.long_slope)));
        output
    }
}
impl Indicator for CyclePolarity {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CyclePolarity,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} polarity", self.mnemonic),
                    description: format!("{} polarity", self.description),
                },
                OutputText {
                    mnemonic: format!("{} short_value", self.mnemonic),
                    description: format!("{} short_value", self.description),
                },
                OutputText {
                    mnemonic: format!("{} long_value", self.mnemonic),
                    description: format!("{} long_value", self.description),
                },
                OutputText {
                    mnemonic: format!("{} short_slope", self.mnemonic),
                    description: format!("{} short_slope", self.description),
                },
                OutputText {
                    mnemonic: format!("{} long_slope", self.mnemonic),
                    description: format!("{} long_slope", self.description),
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
        let p = CyclePolarityParams {
            window: 100,
            short_period: 20.0,
            long_period: 40.0,
            smoothing: 1600.0,
            tolerance: 0.0,
            ..Default::default()
        };
        let mut indicator = CyclePolarity::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("POLARITY", data::expected_w100_s20_l40_hp1600_t0_polarity()),
            (
                "SHORT_VALUE",
                data::expected_w100_s20_l40_hp1600_t0_short_value(),
            ),
            (
                "LONG_VALUE",
                data::expected_w100_s20_l40_hp1600_t0_long_value(),
            ),
            (
                "SHORT_SLOPE",
                data::expected_w100_s20_l40_hp1600_t0_short_slope(),
            ),
            (
                "LONG_SLOPE",
                data::expected_w100_s20_l40_hp1600_t0_long_slope(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("POLARITY", result.polarity),
                ("SHORT_VALUE", result.short_value),
                ("LONG_VALUE", result.long_value),
                ("SHORT_SLOPE", result.short_slope),
                ("LONG_SLOPE", result.long_slope),
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
        let p = CyclePolarityParams {
            window: 100,
            short_period: 20.0,
            long_period: 40.0,
            smoothing: 0.0,
            tolerance: 0.0,
            ..Default::default()
        };
        let mut indicator = CyclePolarity::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("POLARITY", data::expected_w100_s20_l40_hp0_t0_polarity()),
            (
                "SHORT_VALUE",
                data::expected_w100_s20_l40_hp0_t0_short_value(),
            ),
            (
                "LONG_VALUE",
                data::expected_w100_s20_l40_hp0_t0_long_value(),
            ),
            (
                "SHORT_SLOPE",
                data::expected_w100_s20_l40_hp0_t0_short_slope(),
            ),
            (
                "LONG_SLOPE",
                data::expected_w100_s20_l40_hp0_t0_long_slope(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("POLARITY", result.polarity),
                ("SHORT_VALUE", result.short_value),
                ("LONG_VALUE", result.long_value),
                ("SHORT_SLOPE", result.short_slope),
                ("LONG_SLOPE", result.long_slope),
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
        let p = CyclePolarityParams {
            window: 60,
            short_period: 10.0,
            long_period: 20.0,
            smoothing: 1600.0,
            tolerance: 0.0,
            ..Default::default()
        };
        let mut indicator = CyclePolarity::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("POLARITY", data::expected_w60_s10_l20_hp1600_t0_polarity()),
            (
                "SHORT_VALUE",
                data::expected_w60_s10_l20_hp1600_t0_short_value(),
            ),
            (
                "LONG_VALUE",
                data::expected_w60_s10_l20_hp1600_t0_long_value(),
            ),
            (
                "SHORT_SLOPE",
                data::expected_w60_s10_l20_hp1600_t0_short_slope(),
            ),
            (
                "LONG_SLOPE",
                data::expected_w60_s10_l20_hp1600_t0_long_slope(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("POLARITY", result.polarity),
                ("SHORT_VALUE", result.short_value),
                ("LONG_VALUE", result.long_value),
                ("SHORT_SLOPE", result.short_slope),
                ("LONG_SLOPE", result.long_slope),
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
        let p = CyclePolarityParams {
            window: 200,
            short_period: 40.0,
            long_period: 80.0,
            smoothing: 1600.0,
            tolerance: 0.0,
            ..Default::default()
        };
        let mut indicator = CyclePolarity::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("POLARITY", data::expected_w200_s40_l80_hp1600_t0_polarity()),
            (
                "SHORT_VALUE",
                data::expected_w200_s40_l80_hp1600_t0_short_value(),
            ),
            (
                "LONG_VALUE",
                data::expected_w200_s40_l80_hp1600_t0_long_value(),
            ),
            (
                "SHORT_SLOPE",
                data::expected_w200_s40_l80_hp1600_t0_short_slope(),
            ),
            (
                "LONG_SLOPE",
                data::expected_w200_s40_l80_hp1600_t0_long_slope(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("POLARITY", result.polarity),
                ("SHORT_VALUE", result.short_value),
                ("LONG_VALUE", result.long_value),
                ("SHORT_SLOPE", result.short_slope),
                ("LONG_SLOPE", result.long_slope),
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
        let p = CyclePolarityParams {
            window: 100,
            short_period: 17.5,
            long_period: 31.0,
            smoothing: 1600.0,
            tolerance: 0.0,
            ..Default::default()
        };
        let mut indicator = CyclePolarity::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "POLARITY",
                data::expected_w100_s17_5_l31_hp1600_t0_polarity(),
            ),
            (
                "SHORT_VALUE",
                data::expected_w100_s17_5_l31_hp1600_t0_short_value(),
            ),
            (
                "LONG_VALUE",
                data::expected_w100_s17_5_l31_hp1600_t0_long_value(),
            ),
            (
                "SHORT_SLOPE",
                data::expected_w100_s17_5_l31_hp1600_t0_short_slope(),
            ),
            (
                "LONG_SLOPE",
                data::expected_w100_s17_5_l31_hp1600_t0_long_slope(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("POLARITY", result.polarity),
                ("SHORT_VALUE", result.short_value),
                ("LONG_VALUE", result.long_value),
                ("SHORT_SLOPE", result.short_slope),
                ("LONG_SLOPE", result.long_slope),
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
        let p = CyclePolarityParams {
            window: 100,
            short_period: 20.0,
            long_period: 40.0,
            smoothing: 1600.0,
            tolerance: 0.1,
            ..Default::default()
        };
        let mut indicator = CyclePolarity::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "POLARITY",
                data::expected_w100_s20_l40_hp1600_t0_1_polarity(),
            ),
            (
                "SHORT_VALUE",
                data::expected_w100_s20_l40_hp1600_t0_1_short_value(),
            ),
            (
                "LONG_VALUE",
                data::expected_w100_s20_l40_hp1600_t0_1_long_value(),
            ),
            (
                "SHORT_SLOPE",
                data::expected_w100_s20_l40_hp1600_t0_1_short_slope(),
            ),
            (
                "LONG_SLOPE",
                data::expected_w100_s20_l40_hp1600_t0_1_long_slope(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("POLARITY", result.polarity),
                ("SHORT_VALUE", result.short_value),
                ("LONG_VALUE", result.long_value),
                ("SHORT_SLOPE", result.short_slope),
                ("LONG_SLOPE", result.long_slope),
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
        let p = CyclePolarityParams {
            window: 100,
            short_period: 20.0,
            long_period: 40.0,
            smoothing: 1600.0,
            tolerance: 1000.0,
            ..Default::default()
        };
        let mut indicator = CyclePolarity::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "POLARITY",
                data::expected_w100_s20_l40_hp1600_t1000_polarity(),
            ),
            (
                "SHORT_VALUE",
                data::expected_w100_s20_l40_hp1600_t1000_short_value(),
            ),
            (
                "LONG_VALUE",
                data::expected_w100_s20_l40_hp1600_t1000_long_value(),
            ),
            (
                "SHORT_SLOPE",
                data::expected_w100_s20_l40_hp1600_t1000_short_slope(),
            ),
            (
                "LONG_SLOPE",
                data::expected_w100_s20_l40_hp1600_t1000_long_slope(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("POLARITY", result.polarity),
                ("SHORT_VALUE", result.short_value),
                ("LONG_VALUE", result.long_value),
                ("SHORT_SLOPE", result.short_slope),
                ("LONG_SLOPE", result.long_slope),
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
        let p = CyclePolarityParams {
            window: 253,
            short_period: 20.0,
            long_period: 40.0,
            smoothing: 1600.0,
            tolerance: 0.0,
            ..Default::default()
        };
        let mut indicator = CyclePolarity::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("POLARITY", data::expected_w253_s20_l40_hp1600_t0_polarity()),
            (
                "SHORT_VALUE",
                data::expected_w253_s20_l40_hp1600_t0_short_value(),
            ),
            (
                "LONG_VALUE",
                data::expected_w253_s20_l40_hp1600_t0_long_value(),
            ),
            (
                "SHORT_SLOPE",
                data::expected_w253_s20_l40_hp1600_t0_short_slope(),
            ),
            (
                "LONG_SLOPE",
                data::expected_w253_s20_l40_hp1600_t0_long_slope(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("POLARITY", result.polarity),
                ("SHORT_VALUE", result.short_value),
                ("LONG_VALUE", result.long_value),
                ("SHORT_SLOPE", result.short_slope),
                ("LONG_SLOPE", result.long_slope),
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
            let p = CyclePolarityParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = CyclePolarity::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
