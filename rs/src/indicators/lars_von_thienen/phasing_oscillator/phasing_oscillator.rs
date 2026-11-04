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
pub struct PhasingOscillatorParams {
    /// The window; reference default 200.
    pub window: usize,
    /// The short min; reference default 5.
    pub short_min: usize,
    /// The short max; reference default 30.
    pub short_max: usize,
    /// The long min; reference default 31.
    pub long_min: usize,
    /// The long max; reference default 80.
    pub long_max: usize,
    /// The smoothing; reference default 1600.0.
    pub smoothing: f64,
    /// The genuine threshold; reference default 49.0.
    pub genuine_threshold: f64,
    /// The phase cycles; reference default 2.0.
    pub phase_cycles: f64,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for PhasingOscillatorParams {
    fn default() -> Self {
        Self {
            window: 200,
            short_min: 5,
            short_max: 30,
            long_min: 31,
            long_max: 80,
            smoothing: 1600.0,
            genuine_threshold: 49.0,
            phase_cycles: 2.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum PhasingOscillatorOutput {
    Value = 1,
    ShortScore = 2,
    LongScore = 3,
    ShortPeriod = 4,
    LongPeriod = 5,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct PhasingOscillator {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl PhasingOscillator {
    /// Construct a validated streaming indicator.
    pub fn new(p: &PhasingOscillatorParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.window as i64, "window", 6) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.short_min as i64, "short_min", 3) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.short_max as i64, "short_max", 3) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.long_min as i64, "long_min", 3) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.long_max as i64, "long_max", 3) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.smoothing, "smoothing", 0.0) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.genuine_threshold, "genuine_threshold", 0.0) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.phase_cycles, "phase_cycles", 0.0) {
            return Err(error);
        }
        if (p.short_max as i64) >= (p.long_min as i64) {
            return Err("short_max must be smaller than long_min".into());
        }
        if p.genuine_threshold > 100.0 || p.phase_cycles != 0.0 && p.phase_cycles < 1.0 {
            return Err("invalid scanner threshold or phase cycles".into());
        }
        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let suffix = component_triple_mnemonic(bc, qc, tc);
        let parameters = [
            (p.window as i64).to_string(),
            (p.short_min as i64).to_string(),
            (p.short_max as i64).to_string(),
            (p.long_min as i64).to_string(),
            (p.long_max as i64).to_string(),
            p.smoothing.to_string(),
            p.genuine_threshold.to_string(),
            p.phase_cycles.to_string(),
        ]
        .join(",");
        let mnemonic = format!("phosc({parameters}{suffix})");
        let description = format!("Phasing oscillator {mnemonic}");
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
        output.push(Box::new(Scalar::new(time, result.short_score)));
        output.push(Box::new(Scalar::new(time, result.long_score)));
        output.push(Box::new(Scalar::new(time, result.short_period)));
        output.push(Box::new(Scalar::new(time, result.long_period)));
        output
    }
}
impl Indicator for PhasingOscillator {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::PhasingOscillator,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} value", self.mnemonic),
                    description: format!("{} value", self.description),
                },
                OutputText {
                    mnemonic: format!("{} short_score", self.mnemonic),
                    description: format!("{} short_score", self.description),
                },
                OutputText {
                    mnemonic: format!("{} long_score", self.mnemonic),
                    description: format!("{} long_score", self.description),
                },
                OutputText {
                    mnemonic: format!("{} short_period", self.mnemonic),
                    description: format!("{} short_period", self.description),
                },
                OutputText {
                    mnemonic: format!("{} long_period", self.mnemonic),
                    description: format!("{} long_period", self.description),
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
        let p = PhasingOscillatorParams {
            genuine_threshold: 49.0,
            long_max: 80,
            long_min: 31,
            phase_cycles: 2.0,
            short_max: 30,
            short_min: 5,
            smoothing: 1600.0,
            window: 200,
            ..Default::default()
        };
        let mut indicator = PhasingOscillator::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "VALUE",
                data::expected_w200_s05_s130_l031_l180_hp1600_g49_f2_value(),
            ),
            (
                "SHORT_SCORE",
                data::expected_w200_s05_s130_l031_l180_hp1600_g49_f2_short_score(),
            ),
            (
                "LONG_SCORE",
                data::expected_w200_s05_s130_l031_l180_hp1600_g49_f2_long_score(),
            ),
            (
                "SHORT_PERIOD",
                data::expected_w200_s05_s130_l031_l180_hp1600_g49_f2_short_period(),
            ),
            (
                "LONG_PERIOD",
                data::expected_w200_s05_s130_l031_l180_hp1600_g49_f2_long_period(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("SHORT_SCORE", result.short_score),
                ("LONG_SCORE", result.long_score),
                ("SHORT_PERIOD", result.short_period),
                ("LONG_PERIOD", result.long_period),
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
        let p = PhasingOscillatorParams {
            genuine_threshold: 49.0,
            long_max: 40,
            long_min: 16,
            phase_cycles: 0.0,
            short_max: 15,
            short_min: 5,
            smoothing: 1600.0,
            window: 100,
            ..Default::default()
        };
        let mut indicator = PhasingOscillator::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "VALUE",
                data::expected_w100_s05_s115_l016_l140_hp1600_g49_f0_value(),
            ),
            (
                "SHORT_SCORE",
                data::expected_w100_s05_s115_l016_l140_hp1600_g49_f0_short_score(),
            ),
            (
                "LONG_SCORE",
                data::expected_w100_s05_s115_l016_l140_hp1600_g49_f0_long_score(),
            ),
            (
                "SHORT_PERIOD",
                data::expected_w100_s05_s115_l016_l140_hp1600_g49_f0_short_period(),
            ),
            (
                "LONG_PERIOD",
                data::expected_w100_s05_s115_l016_l140_hp1600_g49_f0_long_period(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("SHORT_SCORE", result.short_score),
                ("LONG_SCORE", result.long_score),
                ("SHORT_PERIOD", result.short_period),
                ("LONG_PERIOD", result.long_period),
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
        let p = PhasingOscillatorParams {
            genuine_threshold: 49.0,
            long_max: 40,
            long_min: 16,
            phase_cycles: 2.0,
            short_max: 15,
            short_min: 5,
            smoothing: 0.0,
            window: 100,
            ..Default::default()
        };
        let mut indicator = PhasingOscillator::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "VALUE",
                data::expected_w100_s05_s115_l016_l140_hp0_g49_f2_value(),
            ),
            (
                "SHORT_SCORE",
                data::expected_w100_s05_s115_l016_l140_hp0_g49_f2_short_score(),
            ),
            (
                "LONG_SCORE",
                data::expected_w100_s05_s115_l016_l140_hp0_g49_f2_long_score(),
            ),
            (
                "SHORT_PERIOD",
                data::expected_w100_s05_s115_l016_l140_hp0_g49_f2_short_period(),
            ),
            (
                "LONG_PERIOD",
                data::expected_w100_s05_s115_l016_l140_hp0_g49_f2_long_period(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("SHORT_SCORE", result.short_score),
                ("LONG_SCORE", result.long_score),
                ("SHORT_PERIOD", result.short_period),
                ("LONG_PERIOD", result.long_period),
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
        let p = PhasingOscillatorParams {
            genuine_threshold: 100.0,
            long_max: 80,
            long_min: 31,
            phase_cycles: 2.0,
            short_max: 30,
            short_min: 5,
            smoothing: 1600.0,
            window: 200,
            ..Default::default()
        };
        let mut indicator = PhasingOscillator::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "VALUE",
                data::expected_w200_s05_s130_l031_l180_hp1600_g100_f2_value(),
            ),
            (
                "SHORT_SCORE",
                data::expected_w200_s05_s130_l031_l180_hp1600_g100_f2_short_score(),
            ),
            (
                "LONG_SCORE",
                data::expected_w200_s05_s130_l031_l180_hp1600_g100_f2_long_score(),
            ),
            (
                "SHORT_PERIOD",
                data::expected_w200_s05_s130_l031_l180_hp1600_g100_f2_short_period(),
            ),
            (
                "LONG_PERIOD",
                data::expected_w200_s05_s130_l031_l180_hp1600_g100_f2_long_period(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("SHORT_SCORE", result.short_score),
                ("LONG_SCORE", result.long_score),
                ("SHORT_PERIOD", result.short_period),
                ("LONG_PERIOD", result.long_period),
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
            let p = PhasingOscillatorParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = PhasingOscillator::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
