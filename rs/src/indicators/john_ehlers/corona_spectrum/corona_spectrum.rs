use std::any::Any;

use crate::entities::bar::Bar;
use crate::entities::bar_component::{
    component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT,
};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{
    component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT,
};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{
    component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT,
};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
use crate::indicators::core::outputs::heatmap::Heatmap;
use crate::indicators::john_ehlers::corona::{
    accept_snapshot, format_number, validate_finite, validate_integer, Corona, CoronaParams,
    CoronaSnapshot, FIRST_BANK_SAMPLE,
};

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct CoronaSpectrumParams {
    pub min_raster_value: f64,
    pub max_raster_value: f64,
    pub min_parameter_value: f64,
    pub max_parameter_value: f64,
    pub high_pass_filter_cutoff: i32,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

pub type Params = CoronaSpectrumParams;

impl Default for CoronaSpectrumParams {
    fn default() -> Self {
        Self {
            min_raster_value: 6.0,
            max_raster_value: 20.0,
            min_parameter_value: 6.0,
            max_parameter_value: 30.0,
            high_pass_filter_cutoff: 30,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

pub fn default_params() -> CoronaSpectrumParams {
    CoronaSpectrumParams::default()
}

pub fn mnemonic(params: &CoronaSpectrumParams) -> String {
    let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let comp = component_triple_mnemonic(bc, qc, tc);
    let min_period = params.min_parameter_value.ceil() as i32;
    let max_period = params.max_parameter_value.floor() as i32;
    format!(
        "cspect({}, {}, {}, {}, {}{})",
        format_number(params.min_raster_value),
        format_number(params.max_raster_value),
        min_period,
        max_period,
        params.high_pass_filter_cutoff,
        comp
    )
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CoronaSpectrumOutput {
    Value = 1,
    DominantCycle = 2,
    DominantCycleMedian = 3,
}

#[derive(Debug, Clone)]
pub struct CoronaSpectrumSnapshot {
    pub heatmap: Heatmap,
    pub dominant_cycle: f64,
    pub dominant_cycle_median: f64,
    pub maximum_power: f64,
    pub normalization_power: f64,
    pub selected_bins: f64,
    pub has_cycle_estimate: f64,
    pub power: Vec<f64>,
    pub relative_power: Vec<f64>,
}

pub type Snapshot = CoronaSpectrumSnapshot;

pub struct CoronaSpectrum {
    mnemonic_text: String,
    description_text: String,
    mnemonic_dc: String,
    description_dc: String,
    mnemonic_dcm: String,
    description_dcm: String,
    engine: Corona,
    engine_params: CoronaParams,
    sample_count: usize,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CoronaSpectrum {
    pub fn new(params: &CoronaSpectrumParams) -> Result<Self, String> {
        let min_raster = validate_finite("min_raster_value", params.min_raster_value)?;
        let max_raster = validate_finite("max_raster_value", params.max_raster_value)?;
        if min_raster < 0.0 || max_raster <= 0.0 || max_raster <= min_raster {
            return Err("invalid corona spectrum parameters: require 0 <= min_raster_value < max_raster_value".to_string());
        }
        let min_pv = validate_finite("min_parameter_value", params.min_parameter_value)?;
        let max_pv = validate_finite("max_parameter_value", params.max_parameter_value)?;
        let min_param_f = min_pv.ceil();
        let max_param_f = max_pv.floor();
        if min_param_f < 6.0 {
            return Err("invalid corona spectrum parameters: min_parameter_value must round up to >= 6".to_string());
        }
        if max_param_f <= min_param_f {
            return Err("invalid corona spectrum parameters: max_parameter_value must exceed min_parameter_value".to_string());
        }
        validate_integer("high_pass_filter_cutoff", params.high_pass_filter_cutoff, 5)?;

        let min_period = min_param_f as i32;
        let max_period = max_param_f as i32;
        let engine = Corona::new(&CoronaParams {
            high_pass_filter_cutoff: params.high_pass_filter_cutoff,
            minimal_period: min_period,
            maximal_period: max_period,
            decibels_lower_threshold: min_raster,
            decibels_upper_threshold: max_raster,
        })?;
        let engine_params = engine.parameters();

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let comp = component_triple_mnemonic(bc, qc, tc);

        let mnemonic_text = format!(
            "cspect({}, {}, {}, {}, {}{})",
            format_number(min_raster),
            format_number(max_raster),
            min_period,
            max_period,
            params.high_pass_filter_cutoff,
            comp
        );
        let mnemonic_dc = format!("cspect-dc({}{})", params.high_pass_filter_cutoff, comp);
        let mnemonic_dcm = format!("cspect-dcm({}{})", params.high_pass_filter_cutoff, comp);

        Ok(Self {
            description_text: format!("Corona spectrum {mnemonic_text}"),
            mnemonic_text,
            description_dc: format!("Corona spectrum dominant cycle {mnemonic_dc}"),
            mnemonic_dc,
            description_dcm: format!("Corona spectrum dominant cycle median {mnemonic_dcm}"),
            mnemonic_dcm,
            engine,
            engine_params,
            sample_count: 0,
            min_parameter_value: min_param_f,
            max_parameter_value: max_param_f,
            parameter_resolution: 2.0,
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
        })
    }

    pub fn new_default() -> Result<Self, String> {
        Self::new(&CoronaSpectrumParams::default())
    }

    pub fn reset(&mut self) {
        self.engine.reset();
        self.sample_count = 0;
    }

    fn empty_snapshot(&self, time: i64) -> CoronaSpectrumSnapshot {
        let bank_len = self.engine.filter_bank_length();
        CoronaSpectrumSnapshot {
            heatmap: Heatmap::empty(
                time,
                self.min_parameter_value,
                self.max_parameter_value,
                self.parameter_resolution,
            ),
            dominant_cycle: f64::NAN,
            dominant_cycle_median: f64::NAN,
            maximum_power: f64::NAN,
            normalization_power: f64::NAN,
            selected_bins: f64::NAN,
            has_cycle_estimate: f64::NAN,
            power: vec![f64::NAN; bank_len],
            relative_power: vec![f64::NAN; bank_len],
        }
    }

    pub fn try_update_from_snapshot(
        &mut self,
        snapshot: &CoronaSnapshot,
    ) -> Result<CoronaSpectrumSnapshot, String> {
        let accepted = accept_snapshot(snapshot, &self.engine_params, self.sample_count + 1)?;
        if !accepted {
            return Ok(self.empty_snapshot(snapshot.time));
        }
        self.sample_count = snapshot.sample_count;
        if !snapshot.primed {
            return Ok(self.empty_snapshot(snapshot.time));
        }

        let values = snapshot.decibels.clone();
        let mut minimum = values[0];
        let mut maximum = values[0];
        for &item in &values[1..] {
            if item < minimum {
                minimum = item;
            }
            if item > maximum {
                maximum = item;
            }
        }
        let heatmap = Heatmap::new(
            snapshot.time,
            self.min_parameter_value,
            self.max_parameter_value,
            self.parameter_resolution,
            minimum,
            maximum,
            values,
        );

        Ok(CoronaSpectrumSnapshot {
            heatmap,
            dominant_cycle: snapshot.dominant_cycle,
            dominant_cycle_median: snapshot.dominant_cycle_median,
            maximum_power: snapshot.maximum_power,
            normalization_power: snapshot.normalization_power,
            selected_bins: snapshot.selected_bins as f64,
            has_cycle_estimate: if snapshot.has_cycle_estimate { 1.0 } else { 0.0 },
            power: snapshot.power.clone(),
            relative_power: snapshot.relative_power.clone(),
        })
    }

    pub fn update_from_snapshot(&mut self, snapshot: &CoronaSnapshot) -> CoronaSpectrumSnapshot {
        self.try_update_from_snapshot(snapshot)
            .unwrap_or_else(|err| panic!("CoronaSpectrum::update_from_snapshot failed: {err}"))
    }

    pub fn try_update(&mut self, sample: f64, time: i64) -> Result<CoronaSpectrumSnapshot, String> {
        let engine_snap = self.engine.try_update(sample, time)?;
        self.try_update_from_snapshot(&engine_snap)
    }

    pub fn update(&mut self, sample: f64, time: i64) -> CoronaSpectrumSnapshot {
        self.try_update(sample, time)
            .unwrap_or_else(|err| panic!("CoronaSpectrum::update failed: {err}"))
    }

    fn wrap_output(time: i64, snapshot: CoronaSpectrumSnapshot) -> Output {
        vec![
            Box::new(snapshot.heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, snapshot.dominant_cycle)),
            Box::new(Scalar::new(time, snapshot.dominant_cycle_median)),
        ]
    }
}

impl Indicator for CoronaSpectrum {
    fn is_primed(&self) -> bool {
        self.sample_count >= FIRST_BANK_SAMPLE
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CoronaSpectrum,
            &self.mnemonic_text,
            &self.description_text,
            &[
                OutputText {
                    mnemonic: self.mnemonic_text.clone(),
                    description: self.description_text.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_dc.clone(),
                    description: self.description_dc.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_dcm.clone(),
                    description: self.description_dcm.clone(),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        Self::wrap_output(sample.time, self.update(sample.value, sample.time))
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        Self::wrap_output(bar.time, self.update(sample, bar.time))
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.quote_func)(quote);
        Self::wrap_output(quote.time, self.update(sample, quote.time))
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.trade_func)(trade);
        Self::wrap_output(trade.time, self.update(sample, trade.time))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::{expected_data, testdata};
    use crate::indicators::core::outputs::shape::Shape;

    const TOLERANCE: f64 = 1e-10;

    fn assert_close(actual: f64, expected: f64, label: &str) {
        if expected.is_nan() {
            assert!(actual.is_nan(), "{label}: expected NaN, got {actual}");
            return;
        }
        if expected.is_infinite() {
            assert_eq!(actual, expected, "{label}: expected {expected}, got {actual}");
            return;
        }
        assert!(actual.is_finite(), "{label}: expected finite {expected}, got {actual}");
        let bound = TOLERANCE.max(expected.abs() * TOLERANCE);
        assert!(
            (actual - expected).abs() <= bound,
            "{label}: actual={actual}, expected={expected}, diff={}",
            (actual - expected).abs()
        );
    }

    #[allow(clippy::too_many_arguments)]
    fn verify_case(
        case_name: &str,
        hp_cutoff: i32,
        min_period: i32,
        max_period: i32,
        db_threshold: f64,
        expected_heatmap: &[Vec<f64>],
        expected_min: &[f64],
        expected_max: &[f64],
        expected_dc: &[f64],
        expected_dcm: &[f64],
        expected_max_power: &[f64],
        expected_norm_power: &[f64],
        expected_selected: &[f64],
        expected_has_est: &[f64],
        expected_power: &[Vec<f64>],
        expected_rel_power: &[Vec<f64>],
    ) {
        let mut indicator = CoronaSpectrum::new(&CoronaSpectrumParams {
            min_raster_value: db_threshold,
            max_raster_value: 20.0,
            min_parameter_value: min_period as f64,
            max_parameter_value: max_period as f64,
            high_pass_filter_cutoff: hp_cutoff,
            ..Default::default()
        })
        .unwrap();

        let prices = testdata::test_input();
        let bank_len = (2 * (max_period - min_period) + 1) as usize;
        for (index, &price) in prices.iter().enumerate() {
            let snap = indicator.update(price, index as i64);
            assert_close(snap.heatmap.parameter_first, min_period as f64, &format!("{case_name}[{index}].first"));
            assert_close(snap.heatmap.parameter_last, max_period as f64, &format!("{case_name}[{index}].last"));
            assert_close(snap.heatmap.parameter_resolution, 2.0, &format!("{case_name}[{index}].res"));

            let row_empty = expected_dc[index].is_nan();
            assert_eq!(indicator.is_primed(), !row_empty, "{case_name}[{index}].primed");
            if row_empty {
                assert!(snap.heatmap.is_empty(), "{case_name}[{index}] expected empty heatmap");
            } else {
                assert_eq!(snap.heatmap.values.len(), bank_len);
                for bin_index in 0..bank_len {
                    assert_close(
                        snap.heatmap.values[bin_index],
                        expected_heatmap[index][bin_index],
                        &format!("{case_name}[{index}].hm[{bin_index}]"),
                    );
                }
            }
            assert_close(snap.heatmap.value_min, expected_min[index], &format!("{case_name}[{index}].min"));
            assert_close(snap.heatmap.value_max, expected_max[index], &format!("{case_name}[{index}].max"));
            assert_close(snap.dominant_cycle, expected_dc[index], &format!("{case_name}[{index}].dc"));
            assert_close(snap.dominant_cycle_median, expected_dcm[index], &format!("{case_name}[{index}].dcm"));
            assert_close(snap.maximum_power, expected_max_power[index], &format!("{case_name}[{index}].mp"));
            assert_close(snap.normalization_power, expected_norm_power[index], &format!("{case_name}[{index}].np"));
            assert_close(snap.selected_bins, expected_selected[index], &format!("{case_name}[{index}].sb"));
            assert_close(snap.has_cycle_estimate, expected_has_est[index], &format!("{case_name}[{index}].hce"));
            for bin_index in 0..bank_len {
                assert_close(snap.power[bin_index], expected_power[index][bin_index], &format!("{case_name}[{index}].pw[{bin_index}]"));
                assert_close(snap.relative_power[bin_index], expected_rel_power[index][bin_index], &format!("{case_name}[{index}].rp[{bin_index}]"));
            }
        }
    }

    #[test]
    fn test_reference_cases() {
        verify_case("h30_pmin6_pmax30_db6", 30, 6, 30, 6.0, &expected_data::expected_h30_pmin6_pmax30_db6_heatmap(), &expected_data::expected_h30_pmin6_pmax30_db6_heatmap_min(), &expected_data::expected_h30_pmin6_pmax30_db6_heatmap_max(), &expected_data::expected_h30_pmin6_pmax30_db6_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax30_db6_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax30_db6_maximum_power(), &expected_data::expected_h30_pmin6_pmax30_db6_normalization_power(), &expected_data::expected_h30_pmin6_pmax30_db6_selected_bins(), &expected_data::expected_h30_pmin6_pmax30_db6_has_cycle_estimate(), &expected_data::expected_h30_pmin6_pmax30_db6_power(), &expected_data::expected_h30_pmin6_pmax30_db6_relative_power());
        verify_case("h30_pmin6_pmax30_db3", 30, 6, 30, 3.0, &expected_data::expected_h30_pmin6_pmax30_db3_heatmap(), &expected_data::expected_h30_pmin6_pmax30_db3_heatmap_min(), &expected_data::expected_h30_pmin6_pmax30_db3_heatmap_max(), &expected_data::expected_h30_pmin6_pmax30_db3_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax30_db3_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax30_db3_maximum_power(), &expected_data::expected_h30_pmin6_pmax30_db3_normalization_power(), &expected_data::expected_h30_pmin6_pmax30_db3_selected_bins(), &expected_data::expected_h30_pmin6_pmax30_db3_has_cycle_estimate(), &expected_data::expected_h30_pmin6_pmax30_db3_power(), &expected_data::expected_h30_pmin6_pmax30_db3_relative_power());
        verify_case("h30_pmin6_pmax20_db6", 30, 6, 20, 6.0, &expected_data::expected_h30_pmin6_pmax20_db6_heatmap(), &expected_data::expected_h30_pmin6_pmax20_db6_heatmap_min(), &expected_data::expected_h30_pmin6_pmax20_db6_heatmap_max(), &expected_data::expected_h30_pmin6_pmax20_db6_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax20_db6_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax20_db6_maximum_power(), &expected_data::expected_h30_pmin6_pmax20_db6_normalization_power(), &expected_data::expected_h30_pmin6_pmax20_db6_selected_bins(), &expected_data::expected_h30_pmin6_pmax20_db6_has_cycle_estimate(), &expected_data::expected_h30_pmin6_pmax20_db6_power(), &expected_data::expected_h30_pmin6_pmax20_db6_relative_power());
        verify_case("h30_pmin6_pmax20_db3", 30, 6, 20, 3.0, &expected_data::expected_h30_pmin6_pmax20_db3_heatmap(), &expected_data::expected_h30_pmin6_pmax20_db3_heatmap_min(), &expected_data::expected_h30_pmin6_pmax20_db3_heatmap_max(), &expected_data::expected_h30_pmin6_pmax20_db3_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax20_db3_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax20_db3_maximum_power(), &expected_data::expected_h30_pmin6_pmax20_db3_normalization_power(), &expected_data::expected_h30_pmin6_pmax20_db3_selected_bins(), &expected_data::expected_h30_pmin6_pmax20_db3_has_cycle_estimate(), &expected_data::expected_h30_pmin6_pmax20_db3_power(), &expected_data::expected_h30_pmin6_pmax20_db3_relative_power());
        verify_case("h30_pmin10_pmax30_db6", 30, 10, 30, 6.0, &expected_data::expected_h30_pmin10_pmax30_db6_heatmap(), &expected_data::expected_h30_pmin10_pmax30_db6_heatmap_min(), &expected_data::expected_h30_pmin10_pmax30_db6_heatmap_max(), &expected_data::expected_h30_pmin10_pmax30_db6_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax30_db6_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax30_db6_maximum_power(), &expected_data::expected_h30_pmin10_pmax30_db6_normalization_power(), &expected_data::expected_h30_pmin10_pmax30_db6_selected_bins(), &expected_data::expected_h30_pmin10_pmax30_db6_has_cycle_estimate(), &expected_data::expected_h30_pmin10_pmax30_db6_power(), &expected_data::expected_h30_pmin10_pmax30_db6_relative_power());
        verify_case("h30_pmin10_pmax30_db3", 30, 10, 30, 3.0, &expected_data::expected_h30_pmin10_pmax30_db3_heatmap(), &expected_data::expected_h30_pmin10_pmax30_db3_heatmap_min(), &expected_data::expected_h30_pmin10_pmax30_db3_heatmap_max(), &expected_data::expected_h30_pmin10_pmax30_db3_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax30_db3_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax30_db3_maximum_power(), &expected_data::expected_h30_pmin10_pmax30_db3_normalization_power(), &expected_data::expected_h30_pmin10_pmax30_db3_selected_bins(), &expected_data::expected_h30_pmin10_pmax30_db3_has_cycle_estimate(), &expected_data::expected_h30_pmin10_pmax30_db3_power(), &expected_data::expected_h30_pmin10_pmax30_db3_relative_power());
        verify_case("h30_pmin10_pmax48_db6", 30, 10, 48, 6.0, &expected_data::expected_h30_pmin10_pmax48_db6_heatmap(), &expected_data::expected_h30_pmin10_pmax48_db6_heatmap_min(), &expected_data::expected_h30_pmin10_pmax48_db6_heatmap_max(), &expected_data::expected_h30_pmin10_pmax48_db6_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax48_db6_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax48_db6_maximum_power(), &expected_data::expected_h30_pmin10_pmax48_db6_normalization_power(), &expected_data::expected_h30_pmin10_pmax48_db6_selected_bins(), &expected_data::expected_h30_pmin10_pmax48_db6_has_cycle_estimate(), &expected_data::expected_h30_pmin10_pmax48_db6_power(), &expected_data::expected_h30_pmin10_pmax48_db6_relative_power());
        verify_case("h30_pmin10_pmax48_db3", 30, 10, 48, 3.0, &expected_data::expected_h30_pmin10_pmax48_db3_heatmap(), &expected_data::expected_h30_pmin10_pmax48_db3_heatmap_min(), &expected_data::expected_h30_pmin10_pmax48_db3_heatmap_max(), &expected_data::expected_h30_pmin10_pmax48_db3_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax48_db3_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax48_db3_maximum_power(), &expected_data::expected_h30_pmin10_pmax48_db3_normalization_power(), &expected_data::expected_h30_pmin10_pmax48_db3_selected_bins(), &expected_data::expected_h30_pmin10_pmax48_db3_has_cycle_estimate(), &expected_data::expected_h30_pmin10_pmax48_db3_power(), &expected_data::expected_h30_pmin10_pmax48_db3_relative_power());
        verify_case("h48_pmin6_pmax30_db6", 48, 6, 30, 6.0, &expected_data::expected_h48_pmin6_pmax30_db6_heatmap(), &expected_data::expected_h48_pmin6_pmax30_db6_heatmap_min(), &expected_data::expected_h48_pmin6_pmax30_db6_heatmap_max(), &expected_data::expected_h48_pmin6_pmax30_db6_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax30_db6_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax30_db6_maximum_power(), &expected_data::expected_h48_pmin6_pmax30_db6_normalization_power(), &expected_data::expected_h48_pmin6_pmax30_db6_selected_bins(), &expected_data::expected_h48_pmin6_pmax30_db6_has_cycle_estimate(), &expected_data::expected_h48_pmin6_pmax30_db6_power(), &expected_data::expected_h48_pmin6_pmax30_db6_relative_power());
        verify_case("h48_pmin6_pmax30_db3", 48, 6, 30, 3.0, &expected_data::expected_h48_pmin6_pmax30_db3_heatmap(), &expected_data::expected_h48_pmin6_pmax30_db3_heatmap_min(), &expected_data::expected_h48_pmin6_pmax30_db3_heatmap_max(), &expected_data::expected_h48_pmin6_pmax30_db3_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax30_db3_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax30_db3_maximum_power(), &expected_data::expected_h48_pmin6_pmax30_db3_normalization_power(), &expected_data::expected_h48_pmin6_pmax30_db3_selected_bins(), &expected_data::expected_h48_pmin6_pmax30_db3_has_cycle_estimate(), &expected_data::expected_h48_pmin6_pmax30_db3_power(), &expected_data::expected_h48_pmin6_pmax30_db3_relative_power());
        verify_case("h48_pmin6_pmax20_db6", 48, 6, 20, 6.0, &expected_data::expected_h48_pmin6_pmax20_db6_heatmap(), &expected_data::expected_h48_pmin6_pmax20_db6_heatmap_min(), &expected_data::expected_h48_pmin6_pmax20_db6_heatmap_max(), &expected_data::expected_h48_pmin6_pmax20_db6_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax20_db6_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax20_db6_maximum_power(), &expected_data::expected_h48_pmin6_pmax20_db6_normalization_power(), &expected_data::expected_h48_pmin6_pmax20_db6_selected_bins(), &expected_data::expected_h48_pmin6_pmax20_db6_has_cycle_estimate(), &expected_data::expected_h48_pmin6_pmax20_db6_power(), &expected_data::expected_h48_pmin6_pmax20_db6_relative_power());
        verify_case("h48_pmin6_pmax20_db3", 48, 6, 20, 3.0, &expected_data::expected_h48_pmin6_pmax20_db3_heatmap(), &expected_data::expected_h48_pmin6_pmax20_db3_heatmap_min(), &expected_data::expected_h48_pmin6_pmax20_db3_heatmap_max(), &expected_data::expected_h48_pmin6_pmax20_db3_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax20_db3_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax20_db3_maximum_power(), &expected_data::expected_h48_pmin6_pmax20_db3_normalization_power(), &expected_data::expected_h48_pmin6_pmax20_db3_selected_bins(), &expected_data::expected_h48_pmin6_pmax20_db3_has_cycle_estimate(), &expected_data::expected_h48_pmin6_pmax20_db3_power(), &expected_data::expected_h48_pmin6_pmax20_db3_relative_power());
        verify_case("h48_pmin10_pmax30_db6", 48, 10, 30, 6.0, &expected_data::expected_h48_pmin10_pmax30_db6_heatmap(), &expected_data::expected_h48_pmin10_pmax30_db6_heatmap_min(), &expected_data::expected_h48_pmin10_pmax30_db6_heatmap_max(), &expected_data::expected_h48_pmin10_pmax30_db6_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax30_db6_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax30_db6_maximum_power(), &expected_data::expected_h48_pmin10_pmax30_db6_normalization_power(), &expected_data::expected_h48_pmin10_pmax30_db6_selected_bins(), &expected_data::expected_h48_pmin10_pmax30_db6_has_cycle_estimate(), &expected_data::expected_h48_pmin10_pmax30_db6_power(), &expected_data::expected_h48_pmin10_pmax30_db6_relative_power());
        verify_case("h48_pmin10_pmax30_db3", 48, 10, 30, 3.0, &expected_data::expected_h48_pmin10_pmax30_db3_heatmap(), &expected_data::expected_h48_pmin10_pmax30_db3_heatmap_min(), &expected_data::expected_h48_pmin10_pmax30_db3_heatmap_max(), &expected_data::expected_h48_pmin10_pmax30_db3_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax30_db3_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax30_db3_maximum_power(), &expected_data::expected_h48_pmin10_pmax30_db3_normalization_power(), &expected_data::expected_h48_pmin10_pmax30_db3_selected_bins(), &expected_data::expected_h48_pmin10_pmax30_db3_has_cycle_estimate(), &expected_data::expected_h48_pmin10_pmax30_db3_power(), &expected_data::expected_h48_pmin10_pmax30_db3_relative_power());
        verify_case("h48_pmin10_pmax48_db6", 48, 10, 48, 6.0, &expected_data::expected_h48_pmin10_pmax48_db6_heatmap(), &expected_data::expected_h48_pmin10_pmax48_db6_heatmap_min(), &expected_data::expected_h48_pmin10_pmax48_db6_heatmap_max(), &expected_data::expected_h48_pmin10_pmax48_db6_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax48_db6_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax48_db6_maximum_power(), &expected_data::expected_h48_pmin10_pmax48_db6_normalization_power(), &expected_data::expected_h48_pmin10_pmax48_db6_selected_bins(), &expected_data::expected_h48_pmin10_pmax48_db6_has_cycle_estimate(), &expected_data::expected_h48_pmin10_pmax48_db6_power(), &expected_data::expected_h48_pmin10_pmax48_db6_relative_power());
        verify_case("h48_pmin10_pmax48_db3", 48, 10, 48, 3.0, &expected_data::expected_h48_pmin10_pmax48_db3_heatmap(), &expected_data::expected_h48_pmin10_pmax48_db3_heatmap_min(), &expected_data::expected_h48_pmin10_pmax48_db3_heatmap_max(), &expected_data::expected_h48_pmin10_pmax48_db3_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax48_db3_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax48_db3_maximum_power(), &expected_data::expected_h48_pmin10_pmax48_db3_normalization_power(), &expected_data::expected_h48_pmin10_pmax48_db3_selected_bins(), &expected_data::expected_h48_pmin10_pmax48_db3_has_cycle_estimate(), &expected_data::expected_h48_pmin10_pmax48_db3_power(), &expected_data::expected_h48_pmin10_pmax48_db3_relative_power());
    }

    #[test]
    fn test_metadata_mnemonics_entities_and_validation() {
        let mut instance = CoronaSpectrum::new_default().unwrap();
        let md = instance.metadata();
        assert_eq!(md.identifier, Identifier::CoronaSpectrum);
        assert_eq!(md.mnemonic, "cspect(6, 20, 6, 30, 30)");
        assert_eq!(md.outputs.len(), 3);
        assert_eq!(md.outputs[0].shape, Shape::Heatmap);
        assert_eq!(md.outputs[1].shape, Shape::Scalar);
        assert_eq!(md.outputs[1].mnemonic, "cspect-dc(30)");
        assert_eq!(md.outputs[2].shape, Shape::Scalar);
        assert_eq!(md.outputs[2].mnemonic, "cspect-dcm(30)");

        let zero_min = CoronaSpectrum::new(&CoronaSpectrumParams {
            min_raster_value: 0.0,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(zero_min.metadata().mnemonic, "cspect(0, 20, 6, 30, 30)");

        let combos = [
            (None, None, None),
            (Some(BarComponent::Open), None, None),
            (None, Some(QuoteComponent::Bid), None),
            (None, None, Some(TradeComponent::Volume)),
            (Some(BarComponent::Median), Some(QuoteComponent::Ask), None),
            (Some(BarComponent::Median), None, Some(TradeComponent::Volume)),
            (None, Some(QuoteComponent::Ask), Some(TradeComponent::Volume)),
            (Some(BarComponent::Open), Some(QuoteComponent::Bid), Some(TradeComponent::Volume)),
        ];
        for (bc, qc, tc) in combos {
            let cfg = CoronaSpectrumParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let ind = CoronaSpectrum::new(&cfg).unwrap();
            assert_eq!(ind.metadata().mnemonic, mnemonic(&cfg));
        }

        assert!(CoronaSpectrum::new(&CoronaSpectrumParams {
            high_pass_filter_cutoff: 0,
            ..Default::default()
        })
        .is_err());
        assert!(CoronaSpectrum::new(&CoronaSpectrumParams {
            min_parameter_value: 5.0,
            ..Default::default()
        })
        .is_err());
        assert!(CoronaSpectrum::new(&CoronaSpectrumParams {
            min_raster_value: -1.0,
            ..Default::default()
        })
        .is_err());

        let prices = testdata::test_input();
        let highs = testdata::test_high();
        let lows = testdata::test_low();
        let opens = testdata::test_open();
        let mut scalar_ind = CoronaSpectrum::new_default().unwrap();
        for index in 0..20 {
            let t = index as i64;
            let bar_out = instance.update_bar(&Bar::new(t, opens[index], highs[index], lows[index], prices[index], 100.0));
            let scalar_out = scalar_ind.update_scalar(&Scalar::new(t, prices[index]));
            let bar_dc = bar_out[1].downcast_ref::<Scalar>().unwrap().value;
            let scalar_dc = scalar_out[1].downcast_ref::<Scalar>().unwrap().value;
            if bar_dc.is_nan() {
                assert!(scalar_dc.is_nan());
            } else {
                assert_eq!(bar_dc.to_bits(), scalar_dc.to_bits());
            }
        }
        instance.reset();
        assert!(!instance.is_primed());
    }
}
