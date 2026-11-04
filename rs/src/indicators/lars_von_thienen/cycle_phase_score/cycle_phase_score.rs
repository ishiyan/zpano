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
pub struct CyclePhaseScoreParams {
    /// The window; reference default 200.
    pub window: usize,
    /// The min period; reference default 5.
    pub min_period: usize,
    /// The max period; reference default 80.
    pub max_period: usize,
    /// The smoothing; reference default 1600.0.
    pub smoothing: f64,
    /// The genuine threshold; reference default 49.0.
    pub genuine_threshold: f64,
    /// The phase cycles; reference default 2.0.
    pub phase_cycles: f64,
    /// The top k; reference default 3.
    pub top_k: usize,
    /// The peak distance; reference default 2.
    pub peak_distance: usize,
    /// The max segments; reference default 5.
    pub max_segments: usize,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for CyclePhaseScoreParams {
    fn default() -> Self {
        Self {
            window: 200,
            min_period: 5,
            max_period: 80,
            smoothing: 1600.0,
            genuine_threshold: 49.0,
            phase_cycles: 2.0,
            top_k: 3,
            peak_distance: 2,
            max_segments: 5,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}
/// Registered plot components.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CyclePhaseScoreOutput {
    Value = 1,
    Period = 2,
    Phase = 3,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct CyclePhaseScore {
    engine: Engine,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
    description: String,
}
impl CyclePhaseScore {
    /// Construct a validated streaming indicator.
    pub fn new(p: &CyclePhaseScoreParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.window as i64, "window", 6) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.min_period as i64, "min_period", 3) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.max_period as i64, "max_period", 3) {
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
        if let Some(error) = validate_integer(p.top_k as i64, "top_k", 1) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.peak_distance as i64, "peak_distance", 1) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.max_segments as i64, "max_segments", 2) {
            return Err(error);
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
            (p.min_period as i64).to_string(),
            (p.max_period as i64).to_string(),
            p.smoothing.to_string(),
            p.genuine_threshold.to_string(),
            p.phase_cycles.to_string(),
            (p.top_k as i64).to_string(),
            (p.peak_distance as i64).to_string(),
            (p.max_segments as i64).to_string(),
        ]
        .join(",");
        let mnemonic = format!("cphase({parameters}{suffix})");
        let description = format!("Cycle phase score {mnemonic}");
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
        output.push(Box::new(Scalar::new(time, result.period)));
        output.push(Box::new(Scalar::new(time, result.phase)));
        output
    }
}
impl Indicator for CyclePhaseScore {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CyclePhaseScore,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: format!("{} value", self.mnemonic),
                    description: format!("{} value", self.description),
                },
                OutputText {
                    mnemonic: format!("{} period", self.mnemonic),
                    description: format!("{} period", self.description),
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
        let p = CyclePhaseScoreParams {
            genuine_threshold: 49.0,
            max_period: 80,
            max_segments: 5,
            min_period: 5,
            peak_distance: 2,
            phase_cycles: 2.0,
            smoothing: 1600.0,
            top_k: 3,
            window: 200,
            ..Default::default()
        };
        let mut indicator = CyclePhaseScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "VALUE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_value(),
            ),
            (
                "PERIOD",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_period(),
            ),
            (
                "PHASE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_phase(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("PERIOD", result.period),
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
        let p = CyclePhaseScoreParams {
            genuine_threshold: 49.0,
            max_period: 12,
            max_segments: 5,
            min_period: 12,
            peak_distance: 2,
            phase_cycles: 0.0,
            smoothing: 1600.0,
            top_k: 1,
            window: 60,
            ..Default::default()
        };
        let mut indicator = CyclePhaseScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "VALUE",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_value(),
            ),
            (
                "PERIOD",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_period(),
            ),
            (
                "PHASE",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_phase(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("PERIOD", result.period),
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
        let p = CyclePhaseScoreParams {
            genuine_threshold: 49.0,
            max_period: 30,
            max_segments: 5,
            min_period: 5,
            peak_distance: 2,
            phase_cycles: 2.0,
            smoothing: 0.0,
            top_k: 2,
            window: 100,
            ..Default::default()
        };
        let mut indicator = CyclePhaseScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "VALUE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_value(),
            ),
            (
                "PERIOD",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_period(),
            ),
            (
                "PHASE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_phase(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("PERIOD", result.period),
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
        let p = CyclePhaseScoreParams {
            genuine_threshold: 100.0,
            max_period: 30,
            max_segments: 5,
            min_period: 5,
            peak_distance: 2,
            phase_cycles: 2.0,
            smoothing: 1600.0,
            top_k: 2,
            window: 100,
            ..Default::default()
        };
        let mut indicator = CyclePhaseScore::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "VALUE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_value(),
            ),
            (
                "PERIOD",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_period(),
            ),
            (
                "PHASE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_phase(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let actual = std::collections::HashMap::from([
                ("VALUE", result.value),
                ("PERIOD", result.period),
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
            let p = CyclePhaseScoreParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = CyclePhaseScore::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
