// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
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
pub struct DetrendedRhythmOscillatorsParams {
    /// The fast length; reference default 3.
    pub fast_length: usize,
    /// The slow length; reference default 50.
    pub slow_length: usize,
    /// The zig period; reference default 30.
    pub zig_period: usize,
    /// The cycle memory; reference default 15.
    pub cycle_memory: usize,
    /// The detrend; reference default True.
    pub detrend: bool,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for DetrendedRhythmOscillatorsParams {
    fn default() -> Self {
        Self {
            fast_length: 3,
            slow_length: 50,
            zig_period: 30,
            cycle_memory: 15,
            detrend: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DetrendedRhythmOscillatorsOutput {
    Value = 1,
    Direction = 2,
    CycleLength = 3,
    PivotValue = 4,
    PivotOffset = 5,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct DetrendedRhythmOscillators {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl DetrendedRhythmOscillators {
    /// Construct a validated streaming indicator.
    pub fn new(p: &DetrendedRhythmOscillatorsParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.fast_length as i64, "fast_length", 1) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.slow_length as i64, "slow_length", 2) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.zig_period as i64, "zig_period", 5) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.cycle_memory as i64, "cycle_memory", 3) {
            return Err(error);
        }
        if (p.fast_length as i64) >= (p.slow_length as i64) {
            return Err("fast_length must be smaller than slow_length".into());
        }
        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let suffix = component_triple_mnemonic(bc, qc, tc);
        let parameters = [
            (p.fast_length as i64).to_string(),
            (p.slow_length as i64).to_string(),
            (p.zig_period as i64).to_string(),
            (p.cycle_memory as i64).to_string(),
            (p.detrend as u8).to_string(),
        ]
        .join(",");
        let mnemonic = format!("dro({parameters}{suffix})");
        let description = format!("Detrended rhythm oscillators {mnemonic}");
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
        output.push(Box::new(Scalar::new(time, result.direction)));
        output.push(Box::new(Scalar::new(time, result.cycle_length)));
        output.push(Box::new(Scalar::new(time, result.pivot_value)));
        output.push(Box::new(Scalar::new(time, result.pivot_offset)));
        output
    }
}
impl Indicator for DetrendedRhythmOscillators {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::DetrendedRhythmOscillators,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} value", self.mnemonic),
                    description: format!("{} value", self.description),
                },
                OutputText {
                    mnemonic: format!("{} direction", self.mnemonic),
                    description: format!("{} direction", self.description),
                },
                OutputText {
                    mnemonic: format!("{} cycle_length", self.mnemonic),
                    description: format!("{} cycle_length", self.description),
                },
                OutputText {
                    mnemonic: format!("{} pivot_value", self.mnemonic),
                    description: format!("{} pivot_value", self.description),
                },
                OutputText {
                    mnemonic: format!("{} pivot_offset", self.mnemonic),
                    description: format!("{} pivot_offset", self.description),
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
        let p = DetrendedRhythmOscillatorsParams {
            fast_length: 3,
            slow_length: 50,
            zig_period: 30,
            cycle_memory: 15,
            detrend: true,
            ..Default::default()
        };
        let mut indicator = DetrendedRhythmOscillators::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_f3_s50_z30_m15_d1_value()),
            ("DIRECTION", data::expected_f3_s50_z30_m15_d1_direction()),
            (
                "CYCLE_LENGTH",
                data::expected_f3_s50_z30_m15_d1_cycle_length(),
            ),
            (
                "PIVOT_VALUE",
                data::expected_f3_s50_z30_m15_d1_pivot_value(),
            ),
            (
                "PIVOT_OFFSET",
                data::expected_f3_s50_z30_m15_d1_pivot_offset(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DIRECTION", result.direction),
                ("CYCLE_LENGTH", result.cycle_length),
                ("PIVOT_VALUE", result.pivot_value),
                ("PIVOT_OFFSET", result.pivot_offset),
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
        let p = DetrendedRhythmOscillatorsParams {
            fast_length: 3,
            slow_length: 50,
            zig_period: 5,
            cycle_memory: 15,
            detrend: true,
            ..Default::default()
        };
        let mut indicator = DetrendedRhythmOscillators::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_f3_s50_z5_m15_d1_value()),
            ("DIRECTION", data::expected_f3_s50_z5_m15_d1_direction()),
            (
                "CYCLE_LENGTH",
                data::expected_f3_s50_z5_m15_d1_cycle_length(),
            ),
            ("PIVOT_VALUE", data::expected_f3_s50_z5_m15_d1_pivot_value()),
            (
                "PIVOT_OFFSET",
                data::expected_f3_s50_z5_m15_d1_pivot_offset(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DIRECTION", result.direction),
                ("CYCLE_LENGTH", result.cycle_length),
                ("PIVOT_VALUE", result.pivot_value),
                ("PIVOT_OFFSET", result.pivot_offset),
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
        let p = DetrendedRhythmOscillatorsParams {
            fast_length: 3,
            slow_length: 50,
            zig_period: 10,
            cycle_memory: 15,
            detrend: true,
            ..Default::default()
        };
        let mut indicator = DetrendedRhythmOscillators::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_f3_s50_z10_m15_d1_value()),
            ("DIRECTION", data::expected_f3_s50_z10_m15_d1_direction()),
            (
                "CYCLE_LENGTH",
                data::expected_f3_s50_z10_m15_d1_cycle_length(),
            ),
            (
                "PIVOT_VALUE",
                data::expected_f3_s50_z10_m15_d1_pivot_value(),
            ),
            (
                "PIVOT_OFFSET",
                data::expected_f3_s50_z10_m15_d1_pivot_offset(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DIRECTION", result.direction),
                ("CYCLE_LENGTH", result.cycle_length),
                ("PIVOT_VALUE", result.pivot_value),
                ("PIVOT_OFFSET", result.pivot_offset),
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
        let p = DetrendedRhythmOscillatorsParams {
            fast_length: 2,
            slow_length: 20,
            zig_period: 10,
            cycle_memory: 15,
            detrend: true,
            ..Default::default()
        };
        let mut indicator = DetrendedRhythmOscillators::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_f2_s20_z10_m15_d1_value()),
            ("DIRECTION", data::expected_f2_s20_z10_m15_d1_direction()),
            (
                "CYCLE_LENGTH",
                data::expected_f2_s20_z10_m15_d1_cycle_length(),
            ),
            (
                "PIVOT_VALUE",
                data::expected_f2_s20_z10_m15_d1_pivot_value(),
            ),
            (
                "PIVOT_OFFSET",
                data::expected_f2_s20_z10_m15_d1_pivot_offset(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DIRECTION", result.direction),
                ("CYCLE_LENGTH", result.cycle_length),
                ("PIVOT_VALUE", result.pivot_value),
                ("PIVOT_OFFSET", result.pivot_offset),
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
        let p = DetrendedRhythmOscillatorsParams {
            fast_length: 5,
            slow_length: 100,
            zig_period: 30,
            cycle_memory: 15,
            detrend: true,
            ..Default::default()
        };
        let mut indicator = DetrendedRhythmOscillators::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_f5_s100_z30_m15_d1_value()),
            ("DIRECTION", data::expected_f5_s100_z30_m15_d1_direction()),
            (
                "CYCLE_LENGTH",
                data::expected_f5_s100_z30_m15_d1_cycle_length(),
            ),
            (
                "PIVOT_VALUE",
                data::expected_f5_s100_z30_m15_d1_pivot_value(),
            ),
            (
                "PIVOT_OFFSET",
                data::expected_f5_s100_z30_m15_d1_pivot_offset(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DIRECTION", result.direction),
                ("CYCLE_LENGTH", result.cycle_length),
                ("PIVOT_VALUE", result.pivot_value),
                ("PIVOT_OFFSET", result.pivot_offset),
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
        let p = DetrendedRhythmOscillatorsParams {
            fast_length: 3,
            slow_length: 50,
            zig_period: 5,
            cycle_memory: 3,
            detrend: true,
            ..Default::default()
        };
        let mut indicator = DetrendedRhythmOscillators::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_f3_s50_z5_m3_d1_value()),
            ("DIRECTION", data::expected_f3_s50_z5_m3_d1_direction()),
            (
                "CYCLE_LENGTH",
                data::expected_f3_s50_z5_m3_d1_cycle_length(),
            ),
            ("PIVOT_VALUE", data::expected_f3_s50_z5_m3_d1_pivot_value()),
            (
                "PIVOT_OFFSET",
                data::expected_f3_s50_z5_m3_d1_pivot_offset(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DIRECTION", result.direction),
                ("CYCLE_LENGTH", result.cycle_length),
                ("PIVOT_VALUE", result.pivot_value),
                ("PIVOT_OFFSET", result.pivot_offset),
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
        let p = DetrendedRhythmOscillatorsParams {
            fast_length: 3,
            slow_length: 50,
            zig_period: 10,
            cycle_memory: 15,
            detrend: false,
            ..Default::default()
        };
        let mut indicator = DetrendedRhythmOscillators::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_f3_s50_z10_m15_d0_value()),
            ("DIRECTION", data::expected_f3_s50_z10_m15_d0_direction()),
            (
                "CYCLE_LENGTH",
                data::expected_f3_s50_z10_m15_d0_cycle_length(),
            ),
            (
                "PIVOT_VALUE",
                data::expected_f3_s50_z10_m15_d0_pivot_value(),
            ),
            (
                "PIVOT_OFFSET",
                data::expected_f3_s50_z10_m15_d0_pivot_offset(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DIRECTION", result.direction),
                ("CYCLE_LENGTH", result.cycle_length),
                ("PIVOT_VALUE", result.pivot_value),
                ("PIVOT_OFFSET", result.pivot_offset),
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
        let p = DetrendedRhythmOscillatorsParams {
            fast_length: 3,
            slow_length: 300,
            zig_period: 30,
            cycle_memory: 15,
            detrend: true,
            ..Default::default()
        };
        let mut indicator = DetrendedRhythmOscillators::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("VALUE", data::expected_f3_s300_z30_m15_d1_value()),
            ("DIRECTION", data::expected_f3_s300_z30_m15_d1_direction()),
            (
                "CYCLE_LENGTH",
                data::expected_f3_s300_z30_m15_d1_cycle_length(),
            ),
            (
                "PIVOT_VALUE",
                data::expected_f3_s300_z30_m15_d1_pivot_value(),
            ),
            (
                "PIVOT_OFFSET",
                data::expected_f3_s300_z30_m15_d1_pivot_offset(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("DIRECTION", result.direction),
                ("CYCLE_LENGTH", result.cycle_length),
                ("PIVOT_VALUE", result.pivot_value),
                ("PIVOT_OFFSET", result.pivot_offset),
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
            let p = DetrendedRhythmOscillatorsParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = DetrendedRhythmOscillators::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
