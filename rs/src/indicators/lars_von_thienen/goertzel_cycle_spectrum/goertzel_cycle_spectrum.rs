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
pub struct GoertzelCycleSpectrumParams {
    /// The window; reference default 128.
    pub window: usize,
    /// The min period; reference default 5.
    pub min_period: usize,
    /// The max period; reference default 50.
    pub max_period: usize,
    /// The step; reference default 1.
    pub step: usize,
    /// The smoothing; reference default 1600.0.
    pub smoothing: f64,
    /// Extraction component; None omits the default from the mnemonic.
    pub bar_component: Option<BarComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub quote_component: Option<QuoteComponent>,
    /// Extraction component; None omits the default from the mnemonic.
    pub trade_component: Option<TradeComponent>,
}
impl Default for GoertzelCycleSpectrumParams {
    fn default() -> Self {
        Self {
            window: 128,
            min_period: 5,
            max_period: 50,
            step: 1,
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
pub enum GoertzelCycleSpectrumOutput {
    Amplitude = 1,
    Phases = 2,
    TroughOffsets = 3,
    DominantPeriod = 4,
    DominantAmplitude = 5,
    DominantPhase = 6,
}
/// Lars von Thienen prepared streaming research indicator.
pub struct GoertzelCycleSpectrum {
    engine: Engine,
    line: LineIndicator,
}
impl GoertzelCycleSpectrum {
    /// Construct a validated streaming indicator.
    pub fn new(p: &GoertzelCycleSpectrumParams) -> Result<Self, String> {
        if let Some(error) = validate_integer(p.window as i64, "window", 6) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.min_period as i64, "min_period", 3) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.max_period as i64, "max_period", 3) {
            return Err(error);
        }
        if let Some(error) = validate_integer(p.step as i64, "step", 1) {
            return Err(error);
        }
        if let Some(error) = validate_real(p.smoothing, "smoothing", 0.0) {
            return Err(error);
        }
        if (p.min_period as i64) > (p.max_period as i64)
            || ((p.max_period as i64) - (p.min_period as i64)) % (p.step as i64) != 0
        {
            return Err("step must land on max_period".into());
        }
        let bc = p.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let suffix = component_triple_mnemonic(bc, qc, tc);
        let parameters = [
            (p.window as i64).to_string(),
            (p.min_period as i64).to_string(),
            (p.max_period as i64).to_string(),
            (p.step as i64).to_string(),
            p.smoothing.to_string(),
        ]
        .join(",");
        let mnemonic = format!("gcs({parameters}{suffix})");
        let description = format!("Goertzel cycle spectrum {mnemonic}");
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
        let values = result.amplitude;
        let minimum = values.iter().copied().reduce(f64::min).unwrap_or(f64::NAN);
        let maximum = values.iter().copied().reduce(f64::max).unwrap_or(f64::NAN);
        output.push(Box::new(Heatmap::new(
            time,
            (self.engine.p.min_period as i64) as f64,
            (self.engine.p.max_period as i64) as f64,
            1.0 / (self.engine.p.step as i64) as f64,
            minimum,
            maximum,
            values,
        )));
        let values = result.phases;
        let minimum = values.iter().copied().reduce(f64::min).unwrap_or(f64::NAN);
        let maximum = values.iter().copied().reduce(f64::max).unwrap_or(f64::NAN);
        output.push(Box::new(Heatmap::new(
            time,
            (self.engine.p.min_period as i64) as f64,
            (self.engine.p.max_period as i64) as f64,
            1.0 / (self.engine.p.step as i64) as f64,
            minimum,
            maximum,
            values,
        )));
        let values = result.trough_offsets;
        let minimum = values.iter().copied().reduce(f64::min).unwrap_or(f64::NAN);
        let maximum = values.iter().copied().reduce(f64::max).unwrap_or(f64::NAN);
        output.push(Box::new(Heatmap::new(
            time,
            (self.engine.p.min_period as i64) as f64,
            (self.engine.p.max_period as i64) as f64,
            1.0 / (self.engine.p.step as i64) as f64,
            minimum,
            maximum,
            values,
        )));
        output.push(Box::new(Scalar::new(time, result.dominant_period)));
        output.push(Box::new(Scalar::new(time, result.dominant_amplitude)));
        output.push(Box::new(Scalar::new(time, result.dominant_phase)));
        output
    }
}
impl Indicator for GoertzelCycleSpectrum {
    fn is_primed(&self) -> bool {
        self.engine.primed()
    }
    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::GoertzelCycleSpectrum,
            &self.line.mnemonic,
            &self.line.description,
            &[
                OutputText {
                    mnemonic: format!("{} amplitude", self.line.mnemonic),
                    description: format!("{} amplitude", self.line.description),
                },
                OutputText {
                    mnemonic: format!("{} phases", self.line.mnemonic),
                    description: format!("{} phases", self.line.description),
                },
                OutputText {
                    mnemonic: format!("{} trough_offsets", self.line.mnemonic),
                    description: format!("{} trough_offsets", self.line.description),
                },
                OutputText {
                    mnemonic: format!("{} dominant_period", self.line.mnemonic),
                    description: format!("{} dominant_period", self.line.description),
                },
                OutputText {
                    mnemonic: format!("{} dominant_amplitude", self.line.mnemonic),
                    description: format!("{} dominant_amplitude", self.line.description),
                },
                OutputText {
                    mnemonic: format!("{} dominant_phase", self.line.mnemonic),
                    description: format!("{} dominant_phase", self.line.description),
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
        let p = GoertzelCycleSpectrumParams {
            window: 128,
            min_period: 5,
            max_period: 50,
            step: 1,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = GoertzelCycleSpectrum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("A5", data::expected_w128_min5_max50_s1_l1600_a5()),
            ("PH5", data::expected_w128_min5_max50_s1_l1600_ph5()),
            ("TO5", data::expected_w128_min5_max50_s1_l1600_to5()),
            ("A6", data::expected_w128_min5_max50_s1_l1600_a6()),
            ("PH6", data::expected_w128_min5_max50_s1_l1600_ph6()),
            ("TO6", data::expected_w128_min5_max50_s1_l1600_to6()),
            ("A7", data::expected_w128_min5_max50_s1_l1600_a7()),
            ("PH7", data::expected_w128_min5_max50_s1_l1600_ph7()),
            ("TO7", data::expected_w128_min5_max50_s1_l1600_to7()),
            ("A8", data::expected_w128_min5_max50_s1_l1600_a8()),
            ("PH8", data::expected_w128_min5_max50_s1_l1600_ph8()),
            ("TO8", data::expected_w128_min5_max50_s1_l1600_to8()),
            ("A9", data::expected_w128_min5_max50_s1_l1600_a9()),
            ("PH9", data::expected_w128_min5_max50_s1_l1600_ph9()),
            ("TO9", data::expected_w128_min5_max50_s1_l1600_to9()),
            ("A10", data::expected_w128_min5_max50_s1_l1600_a10()),
            ("PH10", data::expected_w128_min5_max50_s1_l1600_ph10()),
            ("TO10", data::expected_w128_min5_max50_s1_l1600_to10()),
            ("A11", data::expected_w128_min5_max50_s1_l1600_a11()),
            ("PH11", data::expected_w128_min5_max50_s1_l1600_ph11()),
            ("TO11", data::expected_w128_min5_max50_s1_l1600_to11()),
            ("A12", data::expected_w128_min5_max50_s1_l1600_a12()),
            ("PH12", data::expected_w128_min5_max50_s1_l1600_ph12()),
            ("TO12", data::expected_w128_min5_max50_s1_l1600_to12()),
            ("A13", data::expected_w128_min5_max50_s1_l1600_a13()),
            ("PH13", data::expected_w128_min5_max50_s1_l1600_ph13()),
            ("TO13", data::expected_w128_min5_max50_s1_l1600_to13()),
            ("A14", data::expected_w128_min5_max50_s1_l1600_a14()),
            ("PH14", data::expected_w128_min5_max50_s1_l1600_ph14()),
            ("TO14", data::expected_w128_min5_max50_s1_l1600_to14()),
            ("A15", data::expected_w128_min5_max50_s1_l1600_a15()),
            ("PH15", data::expected_w128_min5_max50_s1_l1600_ph15()),
            ("TO15", data::expected_w128_min5_max50_s1_l1600_to15()),
            ("A16", data::expected_w128_min5_max50_s1_l1600_a16()),
            ("PH16", data::expected_w128_min5_max50_s1_l1600_ph16()),
            ("TO16", data::expected_w128_min5_max50_s1_l1600_to16()),
            ("A17", data::expected_w128_min5_max50_s1_l1600_a17()),
            ("PH17", data::expected_w128_min5_max50_s1_l1600_ph17()),
            ("TO17", data::expected_w128_min5_max50_s1_l1600_to17()),
            ("A18", data::expected_w128_min5_max50_s1_l1600_a18()),
            ("PH18", data::expected_w128_min5_max50_s1_l1600_ph18()),
            ("TO18", data::expected_w128_min5_max50_s1_l1600_to18()),
            ("A19", data::expected_w128_min5_max50_s1_l1600_a19()),
            ("PH19", data::expected_w128_min5_max50_s1_l1600_ph19()),
            ("TO19", data::expected_w128_min5_max50_s1_l1600_to19()),
            ("A20", data::expected_w128_min5_max50_s1_l1600_a20()),
            ("PH20", data::expected_w128_min5_max50_s1_l1600_ph20()),
            ("TO20", data::expected_w128_min5_max50_s1_l1600_to20()),
            ("A21", data::expected_w128_min5_max50_s1_l1600_a21()),
            ("PH21", data::expected_w128_min5_max50_s1_l1600_ph21()),
            ("TO21", data::expected_w128_min5_max50_s1_l1600_to21()),
            ("A22", data::expected_w128_min5_max50_s1_l1600_a22()),
            ("PH22", data::expected_w128_min5_max50_s1_l1600_ph22()),
            ("TO22", data::expected_w128_min5_max50_s1_l1600_to22()),
            ("A23", data::expected_w128_min5_max50_s1_l1600_a23()),
            ("PH23", data::expected_w128_min5_max50_s1_l1600_ph23()),
            ("TO23", data::expected_w128_min5_max50_s1_l1600_to23()),
            ("A24", data::expected_w128_min5_max50_s1_l1600_a24()),
            ("PH24", data::expected_w128_min5_max50_s1_l1600_ph24()),
            ("TO24", data::expected_w128_min5_max50_s1_l1600_to24()),
            ("A25", data::expected_w128_min5_max50_s1_l1600_a25()),
            ("PH25", data::expected_w128_min5_max50_s1_l1600_ph25()),
            ("TO25", data::expected_w128_min5_max50_s1_l1600_to25()),
            ("A26", data::expected_w128_min5_max50_s1_l1600_a26()),
            ("PH26", data::expected_w128_min5_max50_s1_l1600_ph26()),
            ("TO26", data::expected_w128_min5_max50_s1_l1600_to26()),
            ("A27", data::expected_w128_min5_max50_s1_l1600_a27()),
            ("PH27", data::expected_w128_min5_max50_s1_l1600_ph27()),
            ("TO27", data::expected_w128_min5_max50_s1_l1600_to27()),
            ("A28", data::expected_w128_min5_max50_s1_l1600_a28()),
            ("PH28", data::expected_w128_min5_max50_s1_l1600_ph28()),
            ("TO28", data::expected_w128_min5_max50_s1_l1600_to28()),
            ("A29", data::expected_w128_min5_max50_s1_l1600_a29()),
            ("PH29", data::expected_w128_min5_max50_s1_l1600_ph29()),
            ("TO29", data::expected_w128_min5_max50_s1_l1600_to29()),
            ("A30", data::expected_w128_min5_max50_s1_l1600_a30()),
            ("PH30", data::expected_w128_min5_max50_s1_l1600_ph30()),
            ("TO30", data::expected_w128_min5_max50_s1_l1600_to30()),
            ("A31", data::expected_w128_min5_max50_s1_l1600_a31()),
            ("PH31", data::expected_w128_min5_max50_s1_l1600_ph31()),
            ("TO31", data::expected_w128_min5_max50_s1_l1600_to31()),
            ("A32", data::expected_w128_min5_max50_s1_l1600_a32()),
            ("PH32", data::expected_w128_min5_max50_s1_l1600_ph32()),
            ("TO32", data::expected_w128_min5_max50_s1_l1600_to32()),
            ("A33", data::expected_w128_min5_max50_s1_l1600_a33()),
            ("PH33", data::expected_w128_min5_max50_s1_l1600_ph33()),
            ("TO33", data::expected_w128_min5_max50_s1_l1600_to33()),
            ("A34", data::expected_w128_min5_max50_s1_l1600_a34()),
            ("PH34", data::expected_w128_min5_max50_s1_l1600_ph34()),
            ("TO34", data::expected_w128_min5_max50_s1_l1600_to34()),
            ("A35", data::expected_w128_min5_max50_s1_l1600_a35()),
            ("PH35", data::expected_w128_min5_max50_s1_l1600_ph35()),
            ("TO35", data::expected_w128_min5_max50_s1_l1600_to35()),
            ("A36", data::expected_w128_min5_max50_s1_l1600_a36()),
            ("PH36", data::expected_w128_min5_max50_s1_l1600_ph36()),
            ("TO36", data::expected_w128_min5_max50_s1_l1600_to36()),
            ("A37", data::expected_w128_min5_max50_s1_l1600_a37()),
            ("PH37", data::expected_w128_min5_max50_s1_l1600_ph37()),
            ("TO37", data::expected_w128_min5_max50_s1_l1600_to37()),
            ("A38", data::expected_w128_min5_max50_s1_l1600_a38()),
            ("PH38", data::expected_w128_min5_max50_s1_l1600_ph38()),
            ("TO38", data::expected_w128_min5_max50_s1_l1600_to38()),
            ("A39", data::expected_w128_min5_max50_s1_l1600_a39()),
            ("PH39", data::expected_w128_min5_max50_s1_l1600_ph39()),
            ("TO39", data::expected_w128_min5_max50_s1_l1600_to39()),
            ("A40", data::expected_w128_min5_max50_s1_l1600_a40()),
            ("PH40", data::expected_w128_min5_max50_s1_l1600_ph40()),
            ("TO40", data::expected_w128_min5_max50_s1_l1600_to40()),
            ("A41", data::expected_w128_min5_max50_s1_l1600_a41()),
            ("PH41", data::expected_w128_min5_max50_s1_l1600_ph41()),
            ("TO41", data::expected_w128_min5_max50_s1_l1600_to41()),
            ("A42", data::expected_w128_min5_max50_s1_l1600_a42()),
            ("PH42", data::expected_w128_min5_max50_s1_l1600_ph42()),
            ("TO42", data::expected_w128_min5_max50_s1_l1600_to42()),
            ("A43", data::expected_w128_min5_max50_s1_l1600_a43()),
            ("PH43", data::expected_w128_min5_max50_s1_l1600_ph43()),
            ("TO43", data::expected_w128_min5_max50_s1_l1600_to43()),
            ("A44", data::expected_w128_min5_max50_s1_l1600_a44()),
            ("PH44", data::expected_w128_min5_max50_s1_l1600_ph44()),
            ("TO44", data::expected_w128_min5_max50_s1_l1600_to44()),
            ("A45", data::expected_w128_min5_max50_s1_l1600_a45()),
            ("PH45", data::expected_w128_min5_max50_s1_l1600_ph45()),
            ("TO45", data::expected_w128_min5_max50_s1_l1600_to45()),
            ("A46", data::expected_w128_min5_max50_s1_l1600_a46()),
            ("PH46", data::expected_w128_min5_max50_s1_l1600_ph46()),
            ("TO46", data::expected_w128_min5_max50_s1_l1600_to46()),
            ("A47", data::expected_w128_min5_max50_s1_l1600_a47()),
            ("PH47", data::expected_w128_min5_max50_s1_l1600_ph47()),
            ("TO47", data::expected_w128_min5_max50_s1_l1600_to47()),
            ("A48", data::expected_w128_min5_max50_s1_l1600_a48()),
            ("PH48", data::expected_w128_min5_max50_s1_l1600_ph48()),
            ("TO48", data::expected_w128_min5_max50_s1_l1600_to48()),
            ("A49", data::expected_w128_min5_max50_s1_l1600_a49()),
            ("PH49", data::expected_w128_min5_max50_s1_l1600_ph49()),
            ("TO49", data::expected_w128_min5_max50_s1_l1600_to49()),
            ("A50", data::expected_w128_min5_max50_s1_l1600_a50()),
            ("PH50", data::expected_w128_min5_max50_s1_l1600_ph50()),
            ("TO50", data::expected_w128_min5_max50_s1_l1600_to50()),
            ("DOM", data::expected_w128_min5_max50_s1_l1600_dom()),
            ("AMP", data::expected_w128_min5_max50_s1_l1600_amp()),
            ("PHASE", data::expected_w128_min5_max50_s1_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual = std::collections::HashMap::from([
                ("DOM", result.dominant_period),
                ("AMP", result.dominant_amplitude),
                ("PHASE", result.dominant_phase),
            ]);
            actual.insert("A5", result.amplitude.get(0).copied().unwrap_or(f64::NAN));
            actual.insert("PH5", result.phases.get(0).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO5",
                result.trough_offsets.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A6", result.amplitude.get(1).copied().unwrap_or(f64::NAN));
            actual.insert("PH6", result.phases.get(1).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO6",
                result.trough_offsets.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A7", result.amplitude.get(2).copied().unwrap_or(f64::NAN));
            actual.insert("PH7", result.phases.get(2).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO7",
                result.trough_offsets.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A8", result.amplitude.get(3).copied().unwrap_or(f64::NAN));
            actual.insert("PH8", result.phases.get(3).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO8",
                result.trough_offsets.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A9", result.amplitude.get(4).copied().unwrap_or(f64::NAN));
            actual.insert("PH9", result.phases.get(4).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO9",
                result.trough_offsets.get(4).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A10", result.amplitude.get(5).copied().unwrap_or(f64::NAN));
            actual.insert("PH10", result.phases.get(5).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO10",
                result.trough_offsets.get(5).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A11", result.amplitude.get(6).copied().unwrap_or(f64::NAN));
            actual.insert("PH11", result.phases.get(6).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO11",
                result.trough_offsets.get(6).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A12", result.amplitude.get(7).copied().unwrap_or(f64::NAN));
            actual.insert("PH12", result.phases.get(7).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO12",
                result.trough_offsets.get(7).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A13", result.amplitude.get(8).copied().unwrap_or(f64::NAN));
            actual.insert("PH13", result.phases.get(8).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO13",
                result.trough_offsets.get(8).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A14", result.amplitude.get(9).copied().unwrap_or(f64::NAN));
            actual.insert("PH14", result.phases.get(9).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO14",
                result.trough_offsets.get(9).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A15", result.amplitude.get(10).copied().unwrap_or(f64::NAN));
            actual.insert("PH15", result.phases.get(10).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO15",
                result.trough_offsets.get(10).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A16", result.amplitude.get(11).copied().unwrap_or(f64::NAN));
            actual.insert("PH16", result.phases.get(11).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO16",
                result.trough_offsets.get(11).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A17", result.amplitude.get(12).copied().unwrap_or(f64::NAN));
            actual.insert("PH17", result.phases.get(12).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO17",
                result.trough_offsets.get(12).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A18", result.amplitude.get(13).copied().unwrap_or(f64::NAN));
            actual.insert("PH18", result.phases.get(13).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO18",
                result.trough_offsets.get(13).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A19", result.amplitude.get(14).copied().unwrap_or(f64::NAN));
            actual.insert("PH19", result.phases.get(14).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO19",
                result.trough_offsets.get(14).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A20", result.amplitude.get(15).copied().unwrap_or(f64::NAN));
            actual.insert("PH20", result.phases.get(15).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO20",
                result.trough_offsets.get(15).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A21", result.amplitude.get(16).copied().unwrap_or(f64::NAN));
            actual.insert("PH21", result.phases.get(16).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO21",
                result.trough_offsets.get(16).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A22", result.amplitude.get(17).copied().unwrap_or(f64::NAN));
            actual.insert("PH22", result.phases.get(17).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO22",
                result.trough_offsets.get(17).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A23", result.amplitude.get(18).copied().unwrap_or(f64::NAN));
            actual.insert("PH23", result.phases.get(18).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO23",
                result.trough_offsets.get(18).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A24", result.amplitude.get(19).copied().unwrap_or(f64::NAN));
            actual.insert("PH24", result.phases.get(19).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO24",
                result.trough_offsets.get(19).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A25", result.amplitude.get(20).copied().unwrap_or(f64::NAN));
            actual.insert("PH25", result.phases.get(20).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO25",
                result.trough_offsets.get(20).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A26", result.amplitude.get(21).copied().unwrap_or(f64::NAN));
            actual.insert("PH26", result.phases.get(21).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO26",
                result.trough_offsets.get(21).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A27", result.amplitude.get(22).copied().unwrap_or(f64::NAN));
            actual.insert("PH27", result.phases.get(22).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO27",
                result.trough_offsets.get(22).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A28", result.amplitude.get(23).copied().unwrap_or(f64::NAN));
            actual.insert("PH28", result.phases.get(23).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO28",
                result.trough_offsets.get(23).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A29", result.amplitude.get(24).copied().unwrap_or(f64::NAN));
            actual.insert("PH29", result.phases.get(24).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO29",
                result.trough_offsets.get(24).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A30", result.amplitude.get(25).copied().unwrap_or(f64::NAN));
            actual.insert("PH30", result.phases.get(25).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO30",
                result.trough_offsets.get(25).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A31", result.amplitude.get(26).copied().unwrap_or(f64::NAN));
            actual.insert("PH31", result.phases.get(26).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO31",
                result.trough_offsets.get(26).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A32", result.amplitude.get(27).copied().unwrap_or(f64::NAN));
            actual.insert("PH32", result.phases.get(27).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO32",
                result.trough_offsets.get(27).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A33", result.amplitude.get(28).copied().unwrap_or(f64::NAN));
            actual.insert("PH33", result.phases.get(28).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO33",
                result.trough_offsets.get(28).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A34", result.amplitude.get(29).copied().unwrap_or(f64::NAN));
            actual.insert("PH34", result.phases.get(29).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO34",
                result.trough_offsets.get(29).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A35", result.amplitude.get(30).copied().unwrap_or(f64::NAN));
            actual.insert("PH35", result.phases.get(30).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO35",
                result.trough_offsets.get(30).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A36", result.amplitude.get(31).copied().unwrap_or(f64::NAN));
            actual.insert("PH36", result.phases.get(31).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO36",
                result.trough_offsets.get(31).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A37", result.amplitude.get(32).copied().unwrap_or(f64::NAN));
            actual.insert("PH37", result.phases.get(32).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO37",
                result.trough_offsets.get(32).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A38", result.amplitude.get(33).copied().unwrap_or(f64::NAN));
            actual.insert("PH38", result.phases.get(33).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO38",
                result.trough_offsets.get(33).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A39", result.amplitude.get(34).copied().unwrap_or(f64::NAN));
            actual.insert("PH39", result.phases.get(34).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO39",
                result.trough_offsets.get(34).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A40", result.amplitude.get(35).copied().unwrap_or(f64::NAN));
            actual.insert("PH40", result.phases.get(35).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO40",
                result.trough_offsets.get(35).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A41", result.amplitude.get(36).copied().unwrap_or(f64::NAN));
            actual.insert("PH41", result.phases.get(36).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO41",
                result.trough_offsets.get(36).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A42", result.amplitude.get(37).copied().unwrap_or(f64::NAN));
            actual.insert("PH42", result.phases.get(37).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO42",
                result.trough_offsets.get(37).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A43", result.amplitude.get(38).copied().unwrap_or(f64::NAN));
            actual.insert("PH43", result.phases.get(38).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO43",
                result.trough_offsets.get(38).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A44", result.amplitude.get(39).copied().unwrap_or(f64::NAN));
            actual.insert("PH44", result.phases.get(39).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO44",
                result.trough_offsets.get(39).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A45", result.amplitude.get(40).copied().unwrap_or(f64::NAN));
            actual.insert("PH45", result.phases.get(40).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO45",
                result.trough_offsets.get(40).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A46", result.amplitude.get(41).copied().unwrap_or(f64::NAN));
            actual.insert("PH46", result.phases.get(41).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO46",
                result.trough_offsets.get(41).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A47", result.amplitude.get(42).copied().unwrap_or(f64::NAN));
            actual.insert("PH47", result.phases.get(42).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO47",
                result.trough_offsets.get(42).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A48", result.amplitude.get(43).copied().unwrap_or(f64::NAN));
            actual.insert("PH48", result.phases.get(43).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO48",
                result.trough_offsets.get(43).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A49", result.amplitude.get(44).copied().unwrap_or(f64::NAN));
            actual.insert("PH49", result.phases.get(44).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO49",
                result.trough_offsets.get(44).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A50", result.amplitude.get(45).copied().unwrap_or(f64::NAN));
            actual.insert("PH50", result.phases.get(45).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO50",
                result.trough_offsets.get(45).copied().unwrap_or(f64::NAN),
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
        let p = GoertzelCycleSpectrumParams {
            window: 100,
            min_period: 20,
            max_period: 20,
            step: 1,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = GoertzelCycleSpectrum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("A20", data::expected_w100_min20_max20_s1_l0_a20()),
            ("PH20", data::expected_w100_min20_max20_s1_l0_ph20()),
            ("TO20", data::expected_w100_min20_max20_s1_l0_to20()),
            ("DOM", data::expected_w100_min20_max20_s1_l0_dom()),
            ("AMP", data::expected_w100_min20_max20_s1_l0_amp()),
            ("PHASE", data::expected_w100_min20_max20_s1_l0_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual = std::collections::HashMap::from([
                ("DOM", result.dominant_period),
                ("AMP", result.dominant_amplitude),
                ("PHASE", result.dominant_phase),
            ]);
            actual.insert("A20", result.amplitude.get(0).copied().unwrap_or(f64::NAN));
            actual.insert("PH20", result.phases.get(0).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO20",
                result.trough_offsets.get(0).copied().unwrap_or(f64::NAN),
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
        let p = GoertzelCycleSpectrumParams {
            window: 40,
            min_period: 5,
            max_period: 25,
            step: 5,
            smoothing: 0.0,
            ..Default::default()
        };
        let mut indicator = GoertzelCycleSpectrum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("A5", data::expected_w40_min5_max25_s5_l0_a5()),
            ("PH5", data::expected_w40_min5_max25_s5_l0_ph5()),
            ("TO5", data::expected_w40_min5_max25_s5_l0_to5()),
            ("A10", data::expected_w40_min5_max25_s5_l0_a10()),
            ("PH10", data::expected_w40_min5_max25_s5_l0_ph10()),
            ("TO10", data::expected_w40_min5_max25_s5_l0_to10()),
            ("A15", data::expected_w40_min5_max25_s5_l0_a15()),
            ("PH15", data::expected_w40_min5_max25_s5_l0_ph15()),
            ("TO15", data::expected_w40_min5_max25_s5_l0_to15()),
            ("A20", data::expected_w40_min5_max25_s5_l0_a20()),
            ("PH20", data::expected_w40_min5_max25_s5_l0_ph20()),
            ("TO20", data::expected_w40_min5_max25_s5_l0_to20()),
            ("A25", data::expected_w40_min5_max25_s5_l0_a25()),
            ("PH25", data::expected_w40_min5_max25_s5_l0_ph25()),
            ("TO25", data::expected_w40_min5_max25_s5_l0_to25()),
            ("DOM", data::expected_w40_min5_max25_s5_l0_dom()),
            ("AMP", data::expected_w40_min5_max25_s5_l0_amp()),
            ("PHASE", data::expected_w40_min5_max25_s5_l0_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual = std::collections::HashMap::from([
                ("DOM", result.dominant_period),
                ("AMP", result.dominant_amplitude),
                ("PHASE", result.dominant_phase),
            ]);
            actual.insert("A5", result.amplitude.get(0).copied().unwrap_or(f64::NAN));
            actual.insert("PH5", result.phases.get(0).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO5",
                result.trough_offsets.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A10", result.amplitude.get(1).copied().unwrap_or(f64::NAN));
            actual.insert("PH10", result.phases.get(1).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO10",
                result.trough_offsets.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A15", result.amplitude.get(2).copied().unwrap_or(f64::NAN));
            actual.insert("PH15", result.phases.get(2).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO15",
                result.trough_offsets.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A20", result.amplitude.get(3).copied().unwrap_or(f64::NAN));
            actual.insert("PH20", result.phases.get(3).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO20",
                result.trough_offsets.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A25", result.amplitude.get(4).copied().unwrap_or(f64::NAN));
            actual.insert("PH25", result.phases.get(4).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO25",
                result.trough_offsets.get(4).copied().unwrap_or(f64::NAN),
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
        let p = GoertzelCycleSpectrumParams {
            window: 40,
            min_period: 5,
            max_period: 25,
            step: 5,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = GoertzelCycleSpectrum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("A5", data::expected_w40_min5_max25_s5_l1600_a5()),
            ("PH5", data::expected_w40_min5_max25_s5_l1600_ph5()),
            ("TO5", data::expected_w40_min5_max25_s5_l1600_to5()),
            ("A10", data::expected_w40_min5_max25_s5_l1600_a10()),
            ("PH10", data::expected_w40_min5_max25_s5_l1600_ph10()),
            ("TO10", data::expected_w40_min5_max25_s5_l1600_to10()),
            ("A15", data::expected_w40_min5_max25_s5_l1600_a15()),
            ("PH15", data::expected_w40_min5_max25_s5_l1600_ph15()),
            ("TO15", data::expected_w40_min5_max25_s5_l1600_to15()),
            ("A20", data::expected_w40_min5_max25_s5_l1600_a20()),
            ("PH20", data::expected_w40_min5_max25_s5_l1600_ph20()),
            ("TO20", data::expected_w40_min5_max25_s5_l1600_to20()),
            ("A25", data::expected_w40_min5_max25_s5_l1600_a25()),
            ("PH25", data::expected_w40_min5_max25_s5_l1600_ph25()),
            ("TO25", data::expected_w40_min5_max25_s5_l1600_to25()),
            ("DOM", data::expected_w40_min5_max25_s5_l1600_dom()),
            ("AMP", data::expected_w40_min5_max25_s5_l1600_amp()),
            ("PHASE", data::expected_w40_min5_max25_s5_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual = std::collections::HashMap::from([
                ("DOM", result.dominant_period),
                ("AMP", result.dominant_amplitude),
                ("PHASE", result.dominant_phase),
            ]);
            actual.insert("A5", result.amplitude.get(0).copied().unwrap_or(f64::NAN));
            actual.insert("PH5", result.phases.get(0).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO5",
                result.trough_offsets.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A10", result.amplitude.get(1).copied().unwrap_or(f64::NAN));
            actual.insert("PH10", result.phases.get(1).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO10",
                result.trough_offsets.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A15", result.amplitude.get(2).copied().unwrap_or(f64::NAN));
            actual.insert("PH15", result.phases.get(2).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO15",
                result.trough_offsets.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A20", result.amplitude.get(3).copied().unwrap_or(f64::NAN));
            actual.insert("PH20", result.phases.get(3).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO20",
                result.trough_offsets.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A25", result.amplitude.get(4).copied().unwrap_or(f64::NAN));
            actual.insert("PH25", result.phases.get(4).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO25",
                result.trough_offsets.get(4).copied().unwrap_or(f64::NAN),
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
    fn prepared_case_4() {
        let p = GoertzelCycleSpectrumParams {
            window: 200,
            min_period: 10,
            max_period: 50,
            step: 10,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = GoertzelCycleSpectrum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("A10", data::expected_w200_min10_max50_s10_l1600_a10()),
            ("PH10", data::expected_w200_min10_max50_s10_l1600_ph10()),
            ("TO10", data::expected_w200_min10_max50_s10_l1600_to10()),
            ("A20", data::expected_w200_min10_max50_s10_l1600_a20()),
            ("PH20", data::expected_w200_min10_max50_s10_l1600_ph20()),
            ("TO20", data::expected_w200_min10_max50_s10_l1600_to20()),
            ("A30", data::expected_w200_min10_max50_s10_l1600_a30()),
            ("PH30", data::expected_w200_min10_max50_s10_l1600_ph30()),
            ("TO30", data::expected_w200_min10_max50_s10_l1600_to30()),
            ("A40", data::expected_w200_min10_max50_s10_l1600_a40()),
            ("PH40", data::expected_w200_min10_max50_s10_l1600_ph40()),
            ("TO40", data::expected_w200_min10_max50_s10_l1600_to40()),
            ("A50", data::expected_w200_min10_max50_s10_l1600_a50()),
            ("PH50", data::expected_w200_min10_max50_s10_l1600_ph50()),
            ("TO50", data::expected_w200_min10_max50_s10_l1600_to50()),
            ("DOM", data::expected_w200_min10_max50_s10_l1600_dom()),
            ("AMP", data::expected_w200_min10_max50_s10_l1600_amp()),
            ("PHASE", data::expected_w200_min10_max50_s10_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual = std::collections::HashMap::from([
                ("DOM", result.dominant_period),
                ("AMP", result.dominant_amplitude),
                ("PHASE", result.dominant_phase),
            ]);
            actual.insert("A10", result.amplitude.get(0).copied().unwrap_or(f64::NAN));
            actual.insert("PH10", result.phases.get(0).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO10",
                result.trough_offsets.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A20", result.amplitude.get(1).copied().unwrap_or(f64::NAN));
            actual.insert("PH20", result.phases.get(1).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO20",
                result.trough_offsets.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A30", result.amplitude.get(2).copied().unwrap_or(f64::NAN));
            actual.insert("PH30", result.phases.get(2).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO30",
                result.trough_offsets.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A40", result.amplitude.get(3).copied().unwrap_or(f64::NAN));
            actual.insert("PH40", result.phases.get(3).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO40",
                result.trough_offsets.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A50", result.amplitude.get(4).copied().unwrap_or(f64::NAN));
            actual.insert("PH50", result.phases.get(4).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO50",
                result.trough_offsets.get(4).copied().unwrap_or(f64::NAN),
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
    fn prepared_case_5() {
        let p = GoertzelCycleSpectrumParams {
            window: 20,
            min_period: 10,
            max_period: 50,
            step: 10,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = GoertzelCycleSpectrum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("A10", data::expected_w20_min10_max50_s10_l1600_a10()),
            ("PH10", data::expected_w20_min10_max50_s10_l1600_ph10()),
            ("TO10", data::expected_w20_min10_max50_s10_l1600_to10()),
            ("A20", data::expected_w20_min10_max50_s10_l1600_a20()),
            ("PH20", data::expected_w20_min10_max50_s10_l1600_ph20()),
            ("TO20", data::expected_w20_min10_max50_s10_l1600_to20()),
            ("A30", data::expected_w20_min10_max50_s10_l1600_a30()),
            ("PH30", data::expected_w20_min10_max50_s10_l1600_ph30()),
            ("TO30", data::expected_w20_min10_max50_s10_l1600_to30()),
            ("A40", data::expected_w20_min10_max50_s10_l1600_a40()),
            ("PH40", data::expected_w20_min10_max50_s10_l1600_ph40()),
            ("TO40", data::expected_w20_min10_max50_s10_l1600_to40()),
            ("A50", data::expected_w20_min10_max50_s10_l1600_a50()),
            ("PH50", data::expected_w20_min10_max50_s10_l1600_ph50()),
            ("TO50", data::expected_w20_min10_max50_s10_l1600_to50()),
            ("DOM", data::expected_w20_min10_max50_s10_l1600_dom()),
            ("AMP", data::expected_w20_min10_max50_s10_l1600_amp()),
            ("PHASE", data::expected_w20_min10_max50_s10_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual = std::collections::HashMap::from([
                ("DOM", result.dominant_period),
                ("AMP", result.dominant_amplitude),
                ("PHASE", result.dominant_phase),
            ]);
            actual.insert("A10", result.amplitude.get(0).copied().unwrap_or(f64::NAN));
            actual.insert("PH10", result.phases.get(0).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO10",
                result.trough_offsets.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A20", result.amplitude.get(1).copied().unwrap_or(f64::NAN));
            actual.insert("PH20", result.phases.get(1).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO20",
                result.trough_offsets.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A30", result.amplitude.get(2).copied().unwrap_or(f64::NAN));
            actual.insert("PH30", result.phases.get(2).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO30",
                result.trough_offsets.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A40", result.amplitude.get(3).copied().unwrap_or(f64::NAN));
            actual.insert("PH40", result.phases.get(3).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO40",
                result.trough_offsets.get(3).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A50", result.amplitude.get(4).copied().unwrap_or(f64::NAN));
            actual.insert("PH50", result.phases.get(4).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO50",
                result.trough_offsets.get(4).copied().unwrap_or(f64::NAN),
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
    fn prepared_case_6() {
        let p = GoertzelCycleSpectrumParams {
            window: 252,
            min_period: 8,
            max_period: 32,
            step: 8,
            smoothing: 160.0,
            ..Default::default()
        };
        let mut indicator = GoertzelCycleSpectrum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("A8", data::expected_w252_min8_max32_s8_l160_a8()),
            ("PH8", data::expected_w252_min8_max32_s8_l160_ph8()),
            ("TO8", data::expected_w252_min8_max32_s8_l160_to8()),
            ("A16", data::expected_w252_min8_max32_s8_l160_a16()),
            ("PH16", data::expected_w252_min8_max32_s8_l160_ph16()),
            ("TO16", data::expected_w252_min8_max32_s8_l160_to16()),
            ("A24", data::expected_w252_min8_max32_s8_l160_a24()),
            ("PH24", data::expected_w252_min8_max32_s8_l160_ph24()),
            ("TO24", data::expected_w252_min8_max32_s8_l160_to24()),
            ("A32", data::expected_w252_min8_max32_s8_l160_a32()),
            ("PH32", data::expected_w252_min8_max32_s8_l160_ph32()),
            ("TO32", data::expected_w252_min8_max32_s8_l160_to32()),
            ("DOM", data::expected_w252_min8_max32_s8_l160_dom()),
            ("AMP", data::expected_w252_min8_max32_s8_l160_amp()),
            ("PHASE", data::expected_w252_min8_max32_s8_l160_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual = std::collections::HashMap::from([
                ("DOM", result.dominant_period),
                ("AMP", result.dominant_amplitude),
                ("PHASE", result.dominant_phase),
            ]);
            actual.insert("A8", result.amplitude.get(0).copied().unwrap_or(f64::NAN));
            actual.insert("PH8", result.phases.get(0).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO8",
                result.trough_offsets.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A16", result.amplitude.get(1).copied().unwrap_or(f64::NAN));
            actual.insert("PH16", result.phases.get(1).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO16",
                result.trough_offsets.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A24", result.amplitude.get(2).copied().unwrap_or(f64::NAN));
            actual.insert("PH24", result.phases.get(2).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO24",
                result.trough_offsets.get(2).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A32", result.amplitude.get(3).copied().unwrap_or(f64::NAN));
            actual.insert("PH32", result.phases.get(3).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO32",
                result.trough_offsets.get(3).copied().unwrap_or(f64::NAN),
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
    fn prepared_case_7() {
        let p = GoertzelCycleSpectrumParams {
            window: 253,
            min_period: 10,
            max_period: 30,
            step: 10,
            smoothing: 1600.0,
            ..Default::default()
        };
        let mut indicator = GoertzelCycleSpectrum::new(&p).unwrap();
        let input = data::test_input();
        let expected = vec![
            ("A10", data::expected_w253_min10_max30_s10_l1600_a10()),
            ("PH10", data::expected_w253_min10_max30_s10_l1600_ph10()),
            ("TO10", data::expected_w253_min10_max30_s10_l1600_to10()),
            ("A20", data::expected_w253_min10_max30_s10_l1600_a20()),
            ("PH20", data::expected_w253_min10_max30_s10_l1600_ph20()),
            ("TO20", data::expected_w253_min10_max30_s10_l1600_to20()),
            ("A30", data::expected_w253_min10_max30_s10_l1600_a30()),
            ("PH30", data::expected_w253_min10_max30_s10_l1600_ph30()),
            ("TO30", data::expected_w253_min10_max30_s10_l1600_to30()),
            ("DOM", data::expected_w253_min10_max30_s10_l1600_dom()),
            ("AMP", data::expected_w253_min10_max30_s10_l1600_amp()),
            ("PHASE", data::expected_w253_min10_max30_s10_l1600_phase()),
        ];
        for (index, price) in input.iter().copied().enumerate() {
            let result = indicator.update(price);
            let mut actual = std::collections::HashMap::from([
                ("DOM", result.dominant_period),
                ("AMP", result.dominant_amplitude),
                ("PHASE", result.dominant_phase),
            ]);
            actual.insert("A10", result.amplitude.get(0).copied().unwrap_or(f64::NAN));
            actual.insert("PH10", result.phases.get(0).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO10",
                result.trough_offsets.get(0).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A20", result.amplitude.get(1).copied().unwrap_or(f64::NAN));
            actual.insert("PH20", result.phases.get(1).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO20",
                result.trough_offsets.get(1).copied().unwrap_or(f64::NAN),
            );
            actual.insert("A30", result.amplitude.get(2).copied().unwrap_or(f64::NAN));
            actual.insert("PH30", result.phases.get(2).copied().unwrap_or(f64::NAN));
            actual.insert(
                "TO30",
                result.trough_offsets.get(2).copied().unwrap_or(f64::NAN),
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
            let p = GoertzelCycleSpectrumParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let indicator = GoertzelCycleSpectrum::new(&p).unwrap();
            let suffix = component_triple_mnemonic(
                bc.unwrap_or(DEFAULT_BAR_COMPONENT),
                qc.unwrap_or(DEFAULT_QUOTE_COMPONENT),
                tc.unwrap_or(DEFAULT_TRADE_COMPONENT),
            );
            assert!(indicator.metadata().mnemonic.ends_with(&(suffix + ")")));
        }
    }
}
