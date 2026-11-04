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
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;
use crate::indicators::core::outputs::heatmap::Heatmap;
/// Parameters matching the prepared reference; omitted components use entity defaults.
#[derive(Clone)]
pub struct CompleteCycleScannerParams {
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
impl Default for CompleteCycleScannerParams {
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
pub enum CompleteCycleScannerOutput {
    Spectrum = 1,
    GenuinePercent = 2,
    DominantPeriod = 3,
    AcceptedCount = 4,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct CompleteCycleScanner {
    engine: Engine,
    line: LineIndicator,
}
impl CompleteCycleScanner {
    /// Construct a validated streaming indicator.
    pub fn new(p: &CompleteCycleScannerParams) -> Result<Self, String> {
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
        let mnemonic = format!("cscan({parameters}{suffix})");
        let description = format!("Complete cycle scanner {mnemonic}");
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
        let values = result.spectrum;
        let minimum = values.iter().copied().reduce(f64::min).unwrap_or(f64::NAN);
        let maximum = values.iter().copied().reduce(f64::max).unwrap_or(f64::NAN);
        output.push(Box::new(Heatmap::new(
            time,
            (self.engine.p.min_period as i64) as f64,
            (self.engine.p.max_period as i64) as f64,
            1.0,
            minimum,
            maximum,
            values,
        )));
        let values = result.genuine_percent;
        let minimum = values.iter().copied().reduce(f64::min).unwrap_or(f64::NAN);
        let maximum = values.iter().copied().reduce(f64::max).unwrap_or(f64::NAN);
        output.push(Box::new(Heatmap::new(
            time,
            (self.engine.p.min_period as i64) as f64,
            (self.engine.p.max_period as i64) as f64,
            1.0,
            minimum,
            maximum,
            values,
        )));
        output.push(Box::new(Scalar::new(time, result.dominant_period)));
        output.push(Box::new(Scalar::new(time, result.accepted_count)));
        output
    }
}
impl Indicator for CompleteCycleScanner {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CompleteCycleScanner,
            &self.line.mnemonic,
            &self.line.description,
            &[
                OutputText {
                    mnemonic: format!("{} spectrum", self.line.mnemonic),
                    description: format!("{} spectrum", self.line.description),
                },
                OutputText {
                    mnemonic: format!("{} genuine_percent", self.line.mnemonic),
                    description: format!("{} genuine_percent", self.line.description),
                },
                OutputText {
                    mnemonic: format!("{} dominant_period", self.line.mnemonic),
                    description: format!("{} dominant_period", self.line.description),
                },
                OutputText {
                    mnemonic: format!("{} accepted_count", self.line.mnemonic),
                    description: format!("{} accepted_count", self.line.description),
                },
            ],
        )
    }
    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        self.wrap(sample.time, value)
    }
    fn update_bar(&mut self, sample: &Bar) -> Output {
        let value = self.update((self.line.bar_func)(sample));
        self.wrap(sample.time, value)
    }
    fn update_quote(&mut self, sample: &Quote) -> Output {
        let value = self.update((self.line.quote_func)(sample));
        self.wrap(sample.time, value)
    }
    fn update_trade(&mut self, sample: &Trade) -> Output {
        let value = self.update((self.line.trade_func)(sample));
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
        let p = CompleteCycleScannerParams {
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
        let mut indicator = CompleteCycleScanner::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "ACCEPTED_COUNT",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_accepted_count(),
            ),
            (
                "AMPLITUDE_P5",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p5(),
            ),
            (
                "GENUINE_P5",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p5(),
            ),
            (
                "AMPLITUDE_P6",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p6(),
            ),
            (
                "GENUINE_P6",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p6(),
            ),
            (
                "AMPLITUDE_P7",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p7(),
            ),
            (
                "GENUINE_P7",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p7(),
            ),
            (
                "AMPLITUDE_P8",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p8(),
            ),
            (
                "GENUINE_P8",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p8(),
            ),
            (
                "AMPLITUDE_P9",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p9(),
            ),
            (
                "GENUINE_P9",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p9(),
            ),
            (
                "AMPLITUDE_P10",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p10(),
            ),
            (
                "GENUINE_P10",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p10(),
            ),
            (
                "AMPLITUDE_P11",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p11(),
            ),
            (
                "GENUINE_P11",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p11(),
            ),
            (
                "AMPLITUDE_P12",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p12(),
            ),
            (
                "GENUINE_P12",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p12(),
            ),
            (
                "AMPLITUDE_P13",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p13(),
            ),
            (
                "GENUINE_P13",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p13(),
            ),
            (
                "AMPLITUDE_P14",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p14(),
            ),
            (
                "GENUINE_P14",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p14(),
            ),
            (
                "AMPLITUDE_P15",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p15(),
            ),
            (
                "GENUINE_P15",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p15(),
            ),
            (
                "AMPLITUDE_P16",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p16(),
            ),
            (
                "GENUINE_P16",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p16(),
            ),
            (
                "AMPLITUDE_P17",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p17(),
            ),
            (
                "GENUINE_P17",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p17(),
            ),
            (
                "AMPLITUDE_P18",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p18(),
            ),
            (
                "GENUINE_P18",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p18(),
            ),
            (
                "AMPLITUDE_P19",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p19(),
            ),
            (
                "GENUINE_P19",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p19(),
            ),
            (
                "AMPLITUDE_P20",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p20(),
            ),
            (
                "GENUINE_P20",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p20(),
            ),
            (
                "AMPLITUDE_P21",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p21(),
            ),
            (
                "GENUINE_P21",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p21(),
            ),
            (
                "AMPLITUDE_P22",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p22(),
            ),
            (
                "GENUINE_P22",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p22(),
            ),
            (
                "AMPLITUDE_P23",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p23(),
            ),
            (
                "GENUINE_P23",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p23(),
            ),
            (
                "AMPLITUDE_P24",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p24(),
            ),
            (
                "GENUINE_P24",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p24(),
            ),
            (
                "AMPLITUDE_P25",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p25(),
            ),
            (
                "GENUINE_P25",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p25(),
            ),
            (
                "AMPLITUDE_P26",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p26(),
            ),
            (
                "GENUINE_P26",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p26(),
            ),
            (
                "AMPLITUDE_P27",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p27(),
            ),
            (
                "GENUINE_P27",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p27(),
            ),
            (
                "AMPLITUDE_P28",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p28(),
            ),
            (
                "GENUINE_P28",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p28(),
            ),
            (
                "AMPLITUDE_P29",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p29(),
            ),
            (
                "GENUINE_P29",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p29(),
            ),
            (
                "AMPLITUDE_P30",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p30(),
            ),
            (
                "GENUINE_P30",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p30(),
            ),
            (
                "AMPLITUDE_P31",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p31(),
            ),
            (
                "GENUINE_P31",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p31(),
            ),
            (
                "AMPLITUDE_P32",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p32(),
            ),
            (
                "GENUINE_P32",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p32(),
            ),
            (
                "AMPLITUDE_P33",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p33(),
            ),
            (
                "GENUINE_P33",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p33(),
            ),
            (
                "AMPLITUDE_P34",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p34(),
            ),
            (
                "GENUINE_P34",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p34(),
            ),
            (
                "AMPLITUDE_P35",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p35(),
            ),
            (
                "GENUINE_P35",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p35(),
            ),
            (
                "AMPLITUDE_P36",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p36(),
            ),
            (
                "GENUINE_P36",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p36(),
            ),
            (
                "AMPLITUDE_P37",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p37(),
            ),
            (
                "GENUINE_P37",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p37(),
            ),
            (
                "AMPLITUDE_P38",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p38(),
            ),
            (
                "GENUINE_P38",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p38(),
            ),
            (
                "AMPLITUDE_P39",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p39(),
            ),
            (
                "GENUINE_P39",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p39(),
            ),
            (
                "AMPLITUDE_P40",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p40(),
            ),
            (
                "GENUINE_P40",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p40(),
            ),
            (
                "AMPLITUDE_P41",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p41(),
            ),
            (
                "GENUINE_P41",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p41(),
            ),
            (
                "AMPLITUDE_P42",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p42(),
            ),
            (
                "GENUINE_P42",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p42(),
            ),
            (
                "AMPLITUDE_P43",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p43(),
            ),
            (
                "GENUINE_P43",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p43(),
            ),
            (
                "AMPLITUDE_P44",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p44(),
            ),
            (
                "GENUINE_P44",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p44(),
            ),
            (
                "AMPLITUDE_P45",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p45(),
            ),
            (
                "GENUINE_P45",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p45(),
            ),
            (
                "AMPLITUDE_P46",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p46(),
            ),
            (
                "GENUINE_P46",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p46(),
            ),
            (
                "AMPLITUDE_P47",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p47(),
            ),
            (
                "GENUINE_P47",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p47(),
            ),
            (
                "AMPLITUDE_P48",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p48(),
            ),
            (
                "GENUINE_P48",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p48(),
            ),
            (
                "AMPLITUDE_P49",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p49(),
            ),
            (
                "GENUINE_P49",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p49(),
            ),
            (
                "AMPLITUDE_P50",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p50(),
            ),
            (
                "GENUINE_P50",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p50(),
            ),
            (
                "AMPLITUDE_P51",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p51(),
            ),
            (
                "GENUINE_P51",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p51(),
            ),
            (
                "AMPLITUDE_P52",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p52(),
            ),
            (
                "GENUINE_P52",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p52(),
            ),
            (
                "AMPLITUDE_P53",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p53(),
            ),
            (
                "GENUINE_P53",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p53(),
            ),
            (
                "AMPLITUDE_P54",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p54(),
            ),
            (
                "GENUINE_P54",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p54(),
            ),
            (
                "AMPLITUDE_P55",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p55(),
            ),
            (
                "GENUINE_P55",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p55(),
            ),
            (
                "AMPLITUDE_P56",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p56(),
            ),
            (
                "GENUINE_P56",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p56(),
            ),
            (
                "AMPLITUDE_P57",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p57(),
            ),
            (
                "GENUINE_P57",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p57(),
            ),
            (
                "AMPLITUDE_P58",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p58(),
            ),
            (
                "GENUINE_P58",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p58(),
            ),
            (
                "AMPLITUDE_P59",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p59(),
            ),
            (
                "GENUINE_P59",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p59(),
            ),
            (
                "AMPLITUDE_P60",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p60(),
            ),
            (
                "GENUINE_P60",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p60(),
            ),
            (
                "AMPLITUDE_P61",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p61(),
            ),
            (
                "GENUINE_P61",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p61(),
            ),
            (
                "AMPLITUDE_P62",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p62(),
            ),
            (
                "GENUINE_P62",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p62(),
            ),
            (
                "AMPLITUDE_P63",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p63(),
            ),
            (
                "GENUINE_P63",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p63(),
            ),
            (
                "AMPLITUDE_P64",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p64(),
            ),
            (
                "GENUINE_P64",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p64(),
            ),
            (
                "AMPLITUDE_P65",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p65(),
            ),
            (
                "GENUINE_P65",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p65(),
            ),
            (
                "AMPLITUDE_P66",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p66(),
            ),
            (
                "GENUINE_P66",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p66(),
            ),
            (
                "AMPLITUDE_P67",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p67(),
            ),
            (
                "GENUINE_P67",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p67(),
            ),
            (
                "AMPLITUDE_P68",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p68(),
            ),
            (
                "GENUINE_P68",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p68(),
            ),
            (
                "AMPLITUDE_P69",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p69(),
            ),
            (
                "GENUINE_P69",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p69(),
            ),
            (
                "AMPLITUDE_P70",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p70(),
            ),
            (
                "GENUINE_P70",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p70(),
            ),
            (
                "AMPLITUDE_P71",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p71(),
            ),
            (
                "GENUINE_P71",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p71(),
            ),
            (
                "AMPLITUDE_P72",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p72(),
            ),
            (
                "GENUINE_P72",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p72(),
            ),
            (
                "AMPLITUDE_P73",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p73(),
            ),
            (
                "GENUINE_P73",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p73(),
            ),
            (
                "AMPLITUDE_P74",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p74(),
            ),
            (
                "GENUINE_P74",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p74(),
            ),
            (
                "AMPLITUDE_P75",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p75(),
            ),
            (
                "GENUINE_P75",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p75(),
            ),
            (
                "AMPLITUDE_P76",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p76(),
            ),
            (
                "GENUINE_P76",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p76(),
            ),
            (
                "AMPLITUDE_P77",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p77(),
            ),
            (
                "GENUINE_P77",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p77(),
            ),
            (
                "AMPLITUDE_P78",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p78(),
            ),
            (
                "GENUINE_P78",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p78(),
            ),
            (
                "AMPLITUDE_P79",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p79(),
            ),
            (
                "GENUINE_P79",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p79(),
            ),
            (
                "AMPLITUDE_P80",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_amplitude_p80(),
            ),
            (
                "GENUINE_P80",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_genuine_p80(),
            ),
            (
                "RANK1_PERIOD",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_period(),
            ),
            (
                "RANK1_AMPLITUDE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_amplitude(),
            ),
            (
                "RANK1_PHASE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_phase(),
            ),
            (
                "RANK1_STRENGTH",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_strength(),
            ),
            (
                "RANK1_GENUINE_PERCENT",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_genuine_percent(),
            ),
            (
                "RANK1_SEGMENTS",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_segments(),
            ),
            (
                "RANK1_FIT_AMPLITUDE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_fit_amplitude(),
            ),
            (
                "RANK1_FIT_SAMPLES",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_fit_samples(),
            ),
            (
                "RANK1_VALUE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_value(),
            ),
            (
                "RANK1_SLOPE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_slope(),
            ),
            (
                "RANK1_ETA_TOP",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_eta_top(),
            ),
            (
                "RANK1_ETA_BOTTOM",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank1_eta_bottom(),
            ),
            (
                "RANK2_PERIOD",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_period(),
            ),
            (
                "RANK2_AMPLITUDE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_amplitude(),
            ),
            (
                "RANK2_PHASE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_phase(),
            ),
            (
                "RANK2_STRENGTH",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_strength(),
            ),
            (
                "RANK2_GENUINE_PERCENT",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_genuine_percent(),
            ),
            (
                "RANK2_SEGMENTS",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_segments(),
            ),
            (
                "RANK2_FIT_AMPLITUDE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_fit_amplitude(),
            ),
            (
                "RANK2_FIT_SAMPLES",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_fit_samples(),
            ),
            (
                "RANK2_VALUE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_value(),
            ),
            (
                "RANK2_SLOPE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_slope(),
            ),
            (
                "RANK2_ETA_TOP",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_eta_top(),
            ),
            (
                "RANK2_ETA_BOTTOM",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank2_eta_bottom(),
            ),
            (
                "RANK3_PERIOD",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_period(),
            ),
            (
                "RANK3_AMPLITUDE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_amplitude(),
            ),
            (
                "RANK3_PHASE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_phase(),
            ),
            (
                "RANK3_STRENGTH",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_strength(),
            ),
            (
                "RANK3_GENUINE_PERCENT",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_genuine_percent(),
            ),
            (
                "RANK3_SEGMENTS",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_segments(),
            ),
            (
                "RANK3_FIT_AMPLITUDE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_fit_amplitude(),
            ),
            (
                "RANK3_FIT_SAMPLES",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_fit_samples(),
            ),
            (
                "RANK3_VALUE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_value(),
            ),
            (
                "RANK3_SLOPE",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_slope(),
            ),
            (
                "RANK3_ETA_TOP",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_eta_top(),
            ),
            (
                "RANK3_ETA_BOTTOM",
                data::expected_w200_p05_p180_l1600_g49_f2_k3_d2_m5_rank3_eta_bottom(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual =
                std::collections::HashMap::from([("ACCEPTED_COUNT", result.accepted_count)]);
            actual.insert(
                "AMPLITUDE_P5",
                result.spectrum.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P5",
                result.genuine_percent.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P6",
                result.spectrum.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P6",
                result.genuine_percent.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P7",
                result.spectrum.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P7",
                result.genuine_percent.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P8",
                result.spectrum.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P8",
                result.genuine_percent.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P9",
                result.spectrum.get(4).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P9",
                result.genuine_percent.get(4).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P10",
                result.spectrum.get(5).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P10",
                result.genuine_percent.get(5).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P11",
                result.spectrum.get(6).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P11",
                result.genuine_percent.get(6).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P12",
                result.spectrum.get(7).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P12",
                result.genuine_percent.get(7).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P13",
                result.spectrum.get(8).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P13",
                result.genuine_percent.get(8).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P14",
                result.spectrum.get(9).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P14",
                result.genuine_percent.get(9).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P15",
                result.spectrum.get(10).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P15",
                result.genuine_percent.get(10).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P16",
                result.spectrum.get(11).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P16",
                result.genuine_percent.get(11).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P17",
                result.spectrum.get(12).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P17",
                result.genuine_percent.get(12).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P18",
                result.spectrum.get(13).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P18",
                result.genuine_percent.get(13).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P19",
                result.spectrum.get(14).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P19",
                result.genuine_percent.get(14).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P20",
                result.spectrum.get(15).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P20",
                result.genuine_percent.get(15).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P21",
                result.spectrum.get(16).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P21",
                result.genuine_percent.get(16).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P22",
                result.spectrum.get(17).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P22",
                result.genuine_percent.get(17).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P23",
                result.spectrum.get(18).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P23",
                result.genuine_percent.get(18).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P24",
                result.spectrum.get(19).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P24",
                result.genuine_percent.get(19).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P25",
                result.spectrum.get(20).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P25",
                result.genuine_percent.get(20).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P26",
                result.spectrum.get(21).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P26",
                result.genuine_percent.get(21).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P27",
                result.spectrum.get(22).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P27",
                result.genuine_percent.get(22).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P28",
                result.spectrum.get(23).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P28",
                result.genuine_percent.get(23).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P29",
                result.spectrum.get(24).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P29",
                result.genuine_percent.get(24).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P30",
                result.spectrum.get(25).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P30",
                result.genuine_percent.get(25).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P31",
                result.spectrum.get(26).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P31",
                result.genuine_percent.get(26).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P32",
                result.spectrum.get(27).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P32",
                result.genuine_percent.get(27).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P33",
                result.spectrum.get(28).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P33",
                result.genuine_percent.get(28).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P34",
                result.spectrum.get(29).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P34",
                result.genuine_percent.get(29).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P35",
                result.spectrum.get(30).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P35",
                result.genuine_percent.get(30).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P36",
                result.spectrum.get(31).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P36",
                result.genuine_percent.get(31).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P37",
                result.spectrum.get(32).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P37",
                result.genuine_percent.get(32).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P38",
                result.spectrum.get(33).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P38",
                result.genuine_percent.get(33).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P39",
                result.spectrum.get(34).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P39",
                result.genuine_percent.get(34).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P40",
                result.spectrum.get(35).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P40",
                result.genuine_percent.get(35).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P41",
                result.spectrum.get(36).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P41",
                result.genuine_percent.get(36).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P42",
                result.spectrum.get(37).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P42",
                result.genuine_percent.get(37).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P43",
                result.spectrum.get(38).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P43",
                result.genuine_percent.get(38).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P44",
                result.spectrum.get(39).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P44",
                result.genuine_percent.get(39).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P45",
                result.spectrum.get(40).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P45",
                result.genuine_percent.get(40).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P46",
                result.spectrum.get(41).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P46",
                result.genuine_percent.get(41).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P47",
                result.spectrum.get(42).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P47",
                result.genuine_percent.get(42).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P48",
                result.spectrum.get(43).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P48",
                result.genuine_percent.get(43).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P49",
                result.spectrum.get(44).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P49",
                result.genuine_percent.get(44).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P50",
                result.spectrum.get(45).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P50",
                result.genuine_percent.get(45).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P51",
                result.spectrum.get(46).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P51",
                result.genuine_percent.get(46).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P52",
                result.spectrum.get(47).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P52",
                result.genuine_percent.get(47).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P53",
                result.spectrum.get(48).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P53",
                result.genuine_percent.get(48).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P54",
                result.spectrum.get(49).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P54",
                result.genuine_percent.get(49).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P55",
                result.spectrum.get(50).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P55",
                result.genuine_percent.get(50).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P56",
                result.spectrum.get(51).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P56",
                result.genuine_percent.get(51).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P57",
                result.spectrum.get(52).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P57",
                result.genuine_percent.get(52).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P58",
                result.spectrum.get(53).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P58",
                result.genuine_percent.get(53).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P59",
                result.spectrum.get(54).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P59",
                result.genuine_percent.get(54).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P60",
                result.spectrum.get(55).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P60",
                result.genuine_percent.get(55).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P61",
                result.spectrum.get(56).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P61",
                result.genuine_percent.get(56).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P62",
                result.spectrum.get(57).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P62",
                result.genuine_percent.get(57).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P63",
                result.spectrum.get(58).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P63",
                result.genuine_percent.get(58).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P64",
                result.spectrum.get(59).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P64",
                result.genuine_percent.get(59).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P65",
                result.spectrum.get(60).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P65",
                result.genuine_percent.get(60).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P66",
                result.spectrum.get(61).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P66",
                result.genuine_percent.get(61).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P67",
                result.spectrum.get(62).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P67",
                result.genuine_percent.get(62).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P68",
                result.spectrum.get(63).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P68",
                result.genuine_percent.get(63).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P69",
                result.spectrum.get(64).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P69",
                result.genuine_percent.get(64).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P70",
                result.spectrum.get(65).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P70",
                result.genuine_percent.get(65).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P71",
                result.spectrum.get(66).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P71",
                result.genuine_percent.get(66).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P72",
                result.spectrum.get(67).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P72",
                result.genuine_percent.get(67).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P73",
                result.spectrum.get(68).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P73",
                result.genuine_percent.get(68).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P74",
                result.spectrum.get(69).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P74",
                result.genuine_percent.get(69).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P75",
                result.spectrum.get(70).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P75",
                result.genuine_percent.get(70).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P76",
                result.spectrum.get(71).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P76",
                result.genuine_percent.get(71).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P77",
                result.spectrum.get(72).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P77",
                result.genuine_percent.get(72).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P78",
                result.spectrum.get(73).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P78",
                result.genuine_percent.get(73).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P79",
                result.spectrum.get(74).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P79",
                result.genuine_percent.get(74).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P80",
                result.spectrum.get(75).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P80",
                result.genuine_percent.get(75).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_PERIOD",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.period)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_AMPLITUDE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_PHASE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.phase)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_STRENGTH",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.strength)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_GENUINE_PERCENT",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.genuine_percent)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_SEGMENTS",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.segments)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_FIT_AMPLITUDE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.fit_amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_FIT_SAMPLES",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.fit_samples)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_VALUE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.value)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_SLOPE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.slope)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_ETA_TOP",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.eta_top)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_ETA_BOTTOM",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.eta_bottom)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_PERIOD",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.period)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_AMPLITUDE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_PHASE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.phase)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_STRENGTH",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.strength)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_GENUINE_PERCENT",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.genuine_percent)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_SEGMENTS",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.segments)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_FIT_AMPLITUDE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.fit_amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_FIT_SAMPLES",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.fit_samples)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_VALUE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.value)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_SLOPE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.slope)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_ETA_TOP",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.eta_top)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_ETA_BOTTOM",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.eta_bottom)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_PERIOD",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.period)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_AMPLITUDE",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_PHASE",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.phase)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_STRENGTH",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.strength)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_GENUINE_PERCENT",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.genuine_percent)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_SEGMENTS",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.segments)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_FIT_AMPLITUDE",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.fit_amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_FIT_SAMPLES",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.fit_samples)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_VALUE",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.value)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_SLOPE",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.slope)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_ETA_TOP",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.eta_top)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK3_ETA_BOTTOM",
                result
                    .cycles
                    .get(2)
                    .map(|cycle| cycle.eta_bottom)
                    .unwrap_or(f64::NAN),
            );
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
        let p = CompleteCycleScannerParams {
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
        let mut indicator = CompleteCycleScanner::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "ACCEPTED_COUNT",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_accepted_count(),
            ),
            (
                "AMPLITUDE_P12",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_amplitude_p12(),
            ),
            (
                "GENUINE_P12",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_genuine_p12(),
            ),
            (
                "RANK1_PERIOD",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_period(),
            ),
            (
                "RANK1_AMPLITUDE",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_amplitude(),
            ),
            (
                "RANK1_PHASE",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_phase(),
            ),
            (
                "RANK1_STRENGTH",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_strength(),
            ),
            (
                "RANK1_GENUINE_PERCENT",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_genuine_percent(),
            ),
            (
                "RANK1_SEGMENTS",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_segments(),
            ),
            (
                "RANK1_FIT_AMPLITUDE",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_fit_amplitude(),
            ),
            (
                "RANK1_FIT_SAMPLES",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_fit_samples(),
            ),
            (
                "RANK1_VALUE",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_value(),
            ),
            (
                "RANK1_SLOPE",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_slope(),
            ),
            (
                "RANK1_ETA_TOP",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_eta_top(),
            ),
            (
                "RANK1_ETA_BOTTOM",
                data::expected_w60_p012_p112_l1600_g49_f0_k1_d2_m5_rank1_eta_bottom(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual =
                std::collections::HashMap::from([("ACCEPTED_COUNT", result.accepted_count)]);
            actual.insert(
                "AMPLITUDE_P12",
                result.spectrum.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P12",
                result.genuine_percent.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_PERIOD",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.period)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_AMPLITUDE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_PHASE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.phase)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_STRENGTH",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.strength)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_GENUINE_PERCENT",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.genuine_percent)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_SEGMENTS",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.segments)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_FIT_AMPLITUDE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.fit_amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_FIT_SAMPLES",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.fit_samples)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_VALUE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.value)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_SLOPE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.slope)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_ETA_TOP",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.eta_top)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_ETA_BOTTOM",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.eta_bottom)
                    .unwrap_or(f64::NAN),
            );
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
        let p = CompleteCycleScannerParams {
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
        let mut indicator = CompleteCycleScanner::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "ACCEPTED_COUNT",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_accepted_count(),
            ),
            (
                "AMPLITUDE_P5",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p5(),
            ),
            (
                "GENUINE_P5",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p5(),
            ),
            (
                "AMPLITUDE_P6",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p6(),
            ),
            (
                "GENUINE_P6",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p6(),
            ),
            (
                "AMPLITUDE_P7",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p7(),
            ),
            (
                "GENUINE_P7",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p7(),
            ),
            (
                "AMPLITUDE_P8",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p8(),
            ),
            (
                "GENUINE_P8",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p8(),
            ),
            (
                "AMPLITUDE_P9",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p9(),
            ),
            (
                "GENUINE_P9",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p9(),
            ),
            (
                "AMPLITUDE_P10",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p10(),
            ),
            (
                "GENUINE_P10",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p10(),
            ),
            (
                "AMPLITUDE_P11",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p11(),
            ),
            (
                "GENUINE_P11",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p11(),
            ),
            (
                "AMPLITUDE_P12",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p12(),
            ),
            (
                "GENUINE_P12",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p12(),
            ),
            (
                "AMPLITUDE_P13",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p13(),
            ),
            (
                "GENUINE_P13",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p13(),
            ),
            (
                "AMPLITUDE_P14",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p14(),
            ),
            (
                "GENUINE_P14",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p14(),
            ),
            (
                "AMPLITUDE_P15",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p15(),
            ),
            (
                "GENUINE_P15",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p15(),
            ),
            (
                "AMPLITUDE_P16",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p16(),
            ),
            (
                "GENUINE_P16",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p16(),
            ),
            (
                "AMPLITUDE_P17",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p17(),
            ),
            (
                "GENUINE_P17",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p17(),
            ),
            (
                "AMPLITUDE_P18",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p18(),
            ),
            (
                "GENUINE_P18",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p18(),
            ),
            (
                "AMPLITUDE_P19",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p19(),
            ),
            (
                "GENUINE_P19",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p19(),
            ),
            (
                "AMPLITUDE_P20",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p20(),
            ),
            (
                "GENUINE_P20",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p20(),
            ),
            (
                "AMPLITUDE_P21",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p21(),
            ),
            (
                "GENUINE_P21",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p21(),
            ),
            (
                "AMPLITUDE_P22",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p22(),
            ),
            (
                "GENUINE_P22",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p22(),
            ),
            (
                "AMPLITUDE_P23",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p23(),
            ),
            (
                "GENUINE_P23",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p23(),
            ),
            (
                "AMPLITUDE_P24",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p24(),
            ),
            (
                "GENUINE_P24",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p24(),
            ),
            (
                "AMPLITUDE_P25",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p25(),
            ),
            (
                "GENUINE_P25",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p25(),
            ),
            (
                "AMPLITUDE_P26",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p26(),
            ),
            (
                "GENUINE_P26",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p26(),
            ),
            (
                "AMPLITUDE_P27",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p27(),
            ),
            (
                "GENUINE_P27",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p27(),
            ),
            (
                "AMPLITUDE_P28",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p28(),
            ),
            (
                "GENUINE_P28",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p28(),
            ),
            (
                "AMPLITUDE_P29",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p29(),
            ),
            (
                "GENUINE_P29",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p29(),
            ),
            (
                "AMPLITUDE_P30",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_amplitude_p30(),
            ),
            (
                "GENUINE_P30",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_genuine_p30(),
            ),
            (
                "RANK1_PERIOD",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_period(),
            ),
            (
                "RANK1_AMPLITUDE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_amplitude(),
            ),
            (
                "RANK1_PHASE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_phase(),
            ),
            (
                "RANK1_STRENGTH",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_strength(),
            ),
            (
                "RANK1_GENUINE_PERCENT",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_genuine_percent(),
            ),
            (
                "RANK1_SEGMENTS",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_segments(),
            ),
            (
                "RANK1_FIT_AMPLITUDE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_fit_amplitude(),
            ),
            (
                "RANK1_FIT_SAMPLES",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_fit_samples(),
            ),
            (
                "RANK1_VALUE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_value(),
            ),
            (
                "RANK1_SLOPE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_slope(),
            ),
            (
                "RANK1_ETA_TOP",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_eta_top(),
            ),
            (
                "RANK1_ETA_BOTTOM",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank1_eta_bottom(),
            ),
            (
                "RANK2_PERIOD",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_period(),
            ),
            (
                "RANK2_AMPLITUDE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_amplitude(),
            ),
            (
                "RANK2_PHASE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_phase(),
            ),
            (
                "RANK2_STRENGTH",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_strength(),
            ),
            (
                "RANK2_GENUINE_PERCENT",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_genuine_percent(),
            ),
            (
                "RANK2_SEGMENTS",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_segments(),
            ),
            (
                "RANK2_FIT_AMPLITUDE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_fit_amplitude(),
            ),
            (
                "RANK2_FIT_SAMPLES",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_fit_samples(),
            ),
            (
                "RANK2_VALUE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_value(),
            ),
            (
                "RANK2_SLOPE",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_slope(),
            ),
            (
                "RANK2_ETA_TOP",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_eta_top(),
            ),
            (
                "RANK2_ETA_BOTTOM",
                data::expected_w100_p05_p130_l0_g49_f2_k2_d2_m5_rank2_eta_bottom(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual =
                std::collections::HashMap::from([("ACCEPTED_COUNT", result.accepted_count)]);
            actual.insert(
                "AMPLITUDE_P5",
                result.spectrum.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P5",
                result.genuine_percent.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P6",
                result.spectrum.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P6",
                result.genuine_percent.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P7",
                result.spectrum.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P7",
                result.genuine_percent.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P8",
                result.spectrum.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P8",
                result.genuine_percent.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P9",
                result.spectrum.get(4).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P9",
                result.genuine_percent.get(4).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P10",
                result.spectrum.get(5).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P10",
                result.genuine_percent.get(5).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P11",
                result.spectrum.get(6).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P11",
                result.genuine_percent.get(6).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P12",
                result.spectrum.get(7).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P12",
                result.genuine_percent.get(7).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P13",
                result.spectrum.get(8).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P13",
                result.genuine_percent.get(8).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P14",
                result.spectrum.get(9).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P14",
                result.genuine_percent.get(9).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P15",
                result.spectrum.get(10).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P15",
                result.genuine_percent.get(10).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P16",
                result.spectrum.get(11).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P16",
                result.genuine_percent.get(11).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P17",
                result.spectrum.get(12).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P17",
                result.genuine_percent.get(12).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P18",
                result.spectrum.get(13).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P18",
                result.genuine_percent.get(13).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P19",
                result.spectrum.get(14).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P19",
                result.genuine_percent.get(14).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P20",
                result.spectrum.get(15).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P20",
                result.genuine_percent.get(15).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P21",
                result.spectrum.get(16).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P21",
                result.genuine_percent.get(16).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P22",
                result.spectrum.get(17).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P22",
                result.genuine_percent.get(17).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P23",
                result.spectrum.get(18).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P23",
                result.genuine_percent.get(18).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P24",
                result.spectrum.get(19).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P24",
                result.genuine_percent.get(19).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P25",
                result.spectrum.get(20).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P25",
                result.genuine_percent.get(20).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P26",
                result.spectrum.get(21).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P26",
                result.genuine_percent.get(21).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P27",
                result.spectrum.get(22).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P27",
                result.genuine_percent.get(22).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P28",
                result.spectrum.get(23).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P28",
                result.genuine_percent.get(23).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P29",
                result.spectrum.get(24).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P29",
                result.genuine_percent.get(24).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P30",
                result.spectrum.get(25).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P30",
                result.genuine_percent.get(25).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_PERIOD",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.period)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_AMPLITUDE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_PHASE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.phase)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_STRENGTH",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.strength)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_GENUINE_PERCENT",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.genuine_percent)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_SEGMENTS",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.segments)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_FIT_AMPLITUDE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.fit_amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_FIT_SAMPLES",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.fit_samples)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_VALUE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.value)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_SLOPE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.slope)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_ETA_TOP",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.eta_top)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_ETA_BOTTOM",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.eta_bottom)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_PERIOD",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.period)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_AMPLITUDE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_PHASE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.phase)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_STRENGTH",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.strength)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_GENUINE_PERCENT",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.genuine_percent)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_SEGMENTS",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.segments)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_FIT_AMPLITUDE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.fit_amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_FIT_SAMPLES",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.fit_samples)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_VALUE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.value)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_SLOPE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.slope)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_ETA_TOP",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.eta_top)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_ETA_BOTTOM",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.eta_bottom)
                    .unwrap_or(f64::NAN),
            );
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
        let p = CompleteCycleScannerParams {
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
        let mut indicator = CompleteCycleScanner::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            (
                "ACCEPTED_COUNT",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_accepted_count(),
            ),
            (
                "AMPLITUDE_P5",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p5(),
            ),
            (
                "GENUINE_P5",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p5(),
            ),
            (
                "AMPLITUDE_P6",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p6(),
            ),
            (
                "GENUINE_P6",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p6(),
            ),
            (
                "AMPLITUDE_P7",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p7(),
            ),
            (
                "GENUINE_P7",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p7(),
            ),
            (
                "AMPLITUDE_P8",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p8(),
            ),
            (
                "GENUINE_P8",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p8(),
            ),
            (
                "AMPLITUDE_P9",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p9(),
            ),
            (
                "GENUINE_P9",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p9(),
            ),
            (
                "AMPLITUDE_P10",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p10(),
            ),
            (
                "GENUINE_P10",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p10(),
            ),
            (
                "AMPLITUDE_P11",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p11(),
            ),
            (
                "GENUINE_P11",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p11(),
            ),
            (
                "AMPLITUDE_P12",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p12(),
            ),
            (
                "GENUINE_P12",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p12(),
            ),
            (
                "AMPLITUDE_P13",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p13(),
            ),
            (
                "GENUINE_P13",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p13(),
            ),
            (
                "AMPLITUDE_P14",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p14(),
            ),
            (
                "GENUINE_P14",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p14(),
            ),
            (
                "AMPLITUDE_P15",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p15(),
            ),
            (
                "GENUINE_P15",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p15(),
            ),
            (
                "AMPLITUDE_P16",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p16(),
            ),
            (
                "GENUINE_P16",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p16(),
            ),
            (
                "AMPLITUDE_P17",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p17(),
            ),
            (
                "GENUINE_P17",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p17(),
            ),
            (
                "AMPLITUDE_P18",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p18(),
            ),
            (
                "GENUINE_P18",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p18(),
            ),
            (
                "AMPLITUDE_P19",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p19(),
            ),
            (
                "GENUINE_P19",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p19(),
            ),
            (
                "AMPLITUDE_P20",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p20(),
            ),
            (
                "GENUINE_P20",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p20(),
            ),
            (
                "AMPLITUDE_P21",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p21(),
            ),
            (
                "GENUINE_P21",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p21(),
            ),
            (
                "AMPLITUDE_P22",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p22(),
            ),
            (
                "GENUINE_P22",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p22(),
            ),
            (
                "AMPLITUDE_P23",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p23(),
            ),
            (
                "GENUINE_P23",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p23(),
            ),
            (
                "AMPLITUDE_P24",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p24(),
            ),
            (
                "GENUINE_P24",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p24(),
            ),
            (
                "AMPLITUDE_P25",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p25(),
            ),
            (
                "GENUINE_P25",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p25(),
            ),
            (
                "AMPLITUDE_P26",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p26(),
            ),
            (
                "GENUINE_P26",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p26(),
            ),
            (
                "AMPLITUDE_P27",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p27(),
            ),
            (
                "GENUINE_P27",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p27(),
            ),
            (
                "AMPLITUDE_P28",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p28(),
            ),
            (
                "GENUINE_P28",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p28(),
            ),
            (
                "AMPLITUDE_P29",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p29(),
            ),
            (
                "GENUINE_P29",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p29(),
            ),
            (
                "AMPLITUDE_P30",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_amplitude_p30(),
            ),
            (
                "GENUINE_P30",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_genuine_p30(),
            ),
            (
                "RANK1_PERIOD",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_period(),
            ),
            (
                "RANK1_AMPLITUDE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_amplitude(),
            ),
            (
                "RANK1_PHASE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_phase(),
            ),
            (
                "RANK1_STRENGTH",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_strength(),
            ),
            (
                "RANK1_GENUINE_PERCENT",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_genuine_percent(),
            ),
            (
                "RANK1_SEGMENTS",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_segments(),
            ),
            (
                "RANK1_FIT_AMPLITUDE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_fit_amplitude(),
            ),
            (
                "RANK1_FIT_SAMPLES",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_fit_samples(),
            ),
            (
                "RANK1_VALUE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_value(),
            ),
            (
                "RANK1_SLOPE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_slope(),
            ),
            (
                "RANK1_ETA_TOP",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_eta_top(),
            ),
            (
                "RANK1_ETA_BOTTOM",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank1_eta_bottom(),
            ),
            (
                "RANK2_PERIOD",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_period(),
            ),
            (
                "RANK2_AMPLITUDE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_amplitude(),
            ),
            (
                "RANK2_PHASE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_phase(),
            ),
            (
                "RANK2_STRENGTH",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_strength(),
            ),
            (
                "RANK2_GENUINE_PERCENT",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_genuine_percent(),
            ),
            (
                "RANK2_SEGMENTS",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_segments(),
            ),
            (
                "RANK2_FIT_AMPLITUDE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_fit_amplitude(),
            ),
            (
                "RANK2_FIT_SAMPLES",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_fit_samples(),
            ),
            (
                "RANK2_VALUE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_value(),
            ),
            (
                "RANK2_SLOPE",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_slope(),
            ),
            (
                "RANK2_ETA_TOP",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_eta_top(),
            ),
            (
                "RANK2_ETA_BOTTOM",
                data::expected_w100_p05_p130_l1600_g100_f2_k2_d2_m5_rank2_eta_bottom(),
            ),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual =
                std::collections::HashMap::from([("ACCEPTED_COUNT", result.accepted_count)]);
            actual.insert(
                "AMPLITUDE_P5",
                result.spectrum.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P5",
                result.genuine_percent.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P6",
                result.spectrum.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P6",
                result.genuine_percent.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P7",
                result.spectrum.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P7",
                result.genuine_percent.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P8",
                result.spectrum.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P8",
                result.genuine_percent.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P9",
                result.spectrum.get(4).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P9",
                result.genuine_percent.get(4).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P10",
                result.spectrum.get(5).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P10",
                result.genuine_percent.get(5).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P11",
                result.spectrum.get(6).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P11",
                result.genuine_percent.get(6).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P12",
                result.spectrum.get(7).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P12",
                result.genuine_percent.get(7).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P13",
                result.spectrum.get(8).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P13",
                result.genuine_percent.get(8).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P14",
                result.spectrum.get(9).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P14",
                result.genuine_percent.get(9).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P15",
                result.spectrum.get(10).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P15",
                result.genuine_percent.get(10).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P16",
                result.spectrum.get(11).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P16",
                result.genuine_percent.get(11).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P17",
                result.spectrum.get(12).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P17",
                result.genuine_percent.get(12).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P18",
                result.spectrum.get(13).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P18",
                result.genuine_percent.get(13).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P19",
                result.spectrum.get(14).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P19",
                result.genuine_percent.get(14).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P20",
                result.spectrum.get(15).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P20",
                result.genuine_percent.get(15).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P21",
                result.spectrum.get(16).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P21",
                result.genuine_percent.get(16).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P22",
                result.spectrum.get(17).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P22",
                result.genuine_percent.get(17).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P23",
                result.spectrum.get(18).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P23",
                result.genuine_percent.get(18).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P24",
                result.spectrum.get(19).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P24",
                result.genuine_percent.get(19).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P25",
                result.spectrum.get(20).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P25",
                result.genuine_percent.get(20).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P26",
                result.spectrum.get(21).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P26",
                result.genuine_percent.get(21).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P27",
                result.spectrum.get(22).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P27",
                result.genuine_percent.get(22).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P28",
                result.spectrum.get(23).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P28",
                result.genuine_percent.get(23).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P29",
                result.spectrum.get(24).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P29",
                result.genuine_percent.get(24).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "AMPLITUDE_P30",
                result.spectrum.get(25).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "GENUINE_P30",
                result.genuine_percent.get(25).copied().unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_PERIOD",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.period)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_AMPLITUDE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_PHASE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.phase)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_STRENGTH",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.strength)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_GENUINE_PERCENT",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.genuine_percent)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_SEGMENTS",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.segments)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_FIT_AMPLITUDE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.fit_amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_FIT_SAMPLES",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.fit_samples)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_VALUE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.value)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_SLOPE",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.slope)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_ETA_TOP",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.eta_top)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK1_ETA_BOTTOM",
                result
                    .cycles
                    .get(0)
                    .map(|cycle| cycle.eta_bottom)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_PERIOD",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.period)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_AMPLITUDE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_PHASE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.phase)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_STRENGTH",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.strength)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_GENUINE_PERCENT",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.genuine_percent)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_SEGMENTS",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.segments)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_FIT_AMPLITUDE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.fit_amplitude)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_FIT_SAMPLES",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.fit_samples)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_VALUE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.value)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_SLOPE",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.slope)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_ETA_TOP",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.eta_top)
                    .unwrap_or(f64::NAN),
            );
            actual.insert(
                "RANK2_ETA_BOTTOM",
                result
                    .cycles
                    .get(1)
                    .map(|cycle| cycle.eta_bottom)
                    .unwrap_or(f64::NAN),
            );
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
            let p = CompleteCycleScannerParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = CompleteCycleScanner::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
