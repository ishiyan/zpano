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
    accept_snapshot, clip, format_number, rolling_append, validate_finite, validate_integer,
    AdaptiveBandPass, Corona, CoronaParams, CoronaSnapshot, Raster, RasterKind, FIRST_BANK_SAMPLE,
};

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct CoronaSwingPositionParams {
    pub raster_length: i32,
    pub max_raster_value: f64,
    pub min_parameter_value: f64,
    pub max_parameter_value: f64,
    pub high_pass_filter_cutoff: i32,
    pub minimal_period: i32,
    pub maximal_period: i32,
    pub lead_window: i32,
    pub position_window: i32,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

pub type Params = CoronaSwingPositionParams;

impl Default for CoronaSwingPositionParams {
    fn default() -> Self {
        Self {
            raster_length: 50,
            max_raster_value: 20.0,
            min_parameter_value: -5.0,
            max_parameter_value: 5.0,
            high_pass_filter_cutoff: 30,
            minimal_period: 6,
            maximal_period: 30,
            lead_window: 51,
            position_window: 21,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

pub fn default_params() -> CoronaSwingPositionParams {
    CoronaSwingPositionParams::default()
}

fn swing_flags(params: &CoronaSwingPositionParams) -> String {
    let mut flags = String::new();
    if params.minimal_period != 6 {
        flags.push_str(&format!(", pmin={}", params.minimal_period));
    }
    if params.maximal_period != 30 {
        flags.push_str(&format!(", pmax={}", params.maximal_period));
    }
    if params.lead_window != 51 {
        flags.push_str(&format!(", lead={}", params.lead_window));
    }
    if params.position_window != 21 {
        flags.push_str(&format!(", pos={}", params.position_window));
    }
    flags
}

pub fn mnemonic(params: &CoronaSwingPositionParams) -> String {
    let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let comp = component_triple_mnemonic(bc, qc, tc);
    let flags = swing_flags(params);
    format!(
        "cswing({}, {}, {}, {}, {}{}{})",
        params.raster_length,
        format_number(params.max_raster_value),
        format_number(params.min_parameter_value),
        format_number(params.max_parameter_value),
        params.high_pass_filter_cutoff,
        flags,
        comp
    )
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CoronaSwingPositionOutput {
    Value = 1,
    SwingPosition = 2,
}

#[derive(Debug, Clone)]
pub struct CoronaSwingPositionSnapshot {
    pub heatmap: Heatmap,
    pub swing_position: f64,
    pub normalized_position: f64,
    pub dominant_cycle: f64,
    pub dominant_cycle_median: f64,
    pub band_pass: f64,
    pub quadrature: f64,
    pub cycle_amplitude: f64,
    pub lead60: f64,
    pub position_span: f64,
    pub width: f64,
    pub has_cycle_estimate: f64,
}

pub type Snapshot = CoronaSwingPositionSnapshot;

pub struct CoronaSwingPosition {
    mnemonic_text: String,
    description_text: String,
    mnemonic_sp: String,
    description_sp: String,
    engine: Corona,
    engine_params: CoronaParams,
    bandpass: AdaptiveBandPass,
    raster: Raster,
    lead_window: usize,
    position_window: usize,
    sample_count: usize,
    leads: Vec<f64>,
    positions: Vec<f64>,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CoronaSwingPosition {
    pub fn new(params: &CoronaSwingPositionParams) -> Result<Self, String> {
        validate_integer("lead_window", params.lead_window, 2)?;
        validate_integer("position_window", params.position_window, 2)?;
        let ceiling = validate_finite("max_raster_value", params.max_raster_value)?;
        if ceiling <= 6.0 {
            return Err("invalid corona swing position parameters: max_raster_value must exceed 6".to_string());
        }

        let raster = Raster::new(
            params.raster_length,
            params.min_parameter_value,
            params.max_parameter_value,
            ceiling,
        )?;
        let engine = Corona::new(&CoronaParams {
            high_pass_filter_cutoff: params.high_pass_filter_cutoff,
            minimal_period: params.minimal_period,
            maximal_period: params.maximal_period,
            decibels_lower_threshold: 6.0,
            decibels_upper_threshold: raster.ceiling(),
        })?;
        let engine_params = engine.parameters();

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
        let comp = component_triple_mnemonic(bc, qc, tc);
        let flags = swing_flags(params);

        let mnemonic_text = format!(
            "cswing({}, {}, {}, {}, {}{}{})",
            raster.length(),
            format_number(raster.ceiling()),
            format_number(raster.lower()),
            format_number(raster.upper()),
            params.high_pass_filter_cutoff,
            flags,
            comp
        );
        let mnemonic_sp = format!("cswing-sp({}{}{})", params.high_pass_filter_cutoff, flags, comp);
        let lead_window = params.lead_window as usize;
        let position_window = params.position_window as usize;

        Ok(Self {
            description_text: format!("Corona swing position {mnemonic_text}"),
            mnemonic_text,
            description_sp: format!("Corona swing position scalar {mnemonic_sp}"),
            mnemonic_sp,
            engine,
            engine_params,
            bandpass: AdaptiveBandPass::new(),
            raster,
            lead_window,
            position_window,
            sample_count: 0,
            leads: vec![0.0; lead_window],
            positions: vec![0.0; position_window],
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
        })
    }

    pub fn new_default() -> Result<Self, String> {
        Self::new(&CoronaSwingPositionParams::default())
    }

    pub fn reset(&mut self) {
        self.engine.reset();
        self.bandpass.reset();
        self.raster.reset();
        self.sample_count = 0;
        self.leads = vec![0.0; self.lead_window];
        self.positions = vec![0.0; self.position_window];
    }

    fn empty_snapshot(&self, time: i64) -> CoronaSwingPositionSnapshot {
        CoronaSwingPositionSnapshot {
            heatmap: self.raster.empty(time),
            swing_position: f64::NAN,
            normalized_position: f64::NAN,
            dominant_cycle: f64::NAN,
            dominant_cycle_median: f64::NAN,
            band_pass: f64::NAN,
            quadrature: f64::NAN,
            cycle_amplitude: f64::NAN,
            lead60: f64::NAN,
            position_span: f64::NAN,
            width: f64::NAN,
            has_cycle_estimate: f64::NAN,
        }
    }

    pub fn try_update_from_snapshot(
        &mut self,
        snapshot: &CoronaSnapshot,
    ) -> Result<CoronaSwingPositionSnapshot, String> {
        let accepted = accept_snapshot(snapshot, &self.engine_params, self.sample_count + 1)?;
        if !accepted {
            return Ok(self.empty_snapshot(snapshot.time));
        }
        let cycle = self.bandpass.update(snapshot.price, snapshot.dominant_cycle_median)?;
        self.sample_count = snapshot.sample_count;

        let (lowest, highest) = rolling_append(&mut self.leads, cycle.lead60, self.lead_window);
        let flat = highest == lowest;
        let raw_position = if flat {
            0.5
        } else {
            (cycle.lead60 - lowest) / (highest - lowest)
        };
        let position = clip(raw_position, 0.0, 1.0);
        let (low_position, high_position) =
            rolling_append(&mut self.positions, position, self.position_window);
        let span = high_position - low_position;
        let mut width = if span > 0.85 { 0.01 } else { 0.15 * span };
        if width == 0.0 {
            width = 0.01;
        }

        self.raster
            .update(position, width, RasterKind::Swing, span > 0.8 || flat);
        if !snapshot.primed {
            return Ok(self.empty_snapshot(snapshot.time));
        }

        let mapped = self.raster.lower() + (self.raster.upper() - self.raster.lower()) * position;
        Ok(CoronaSwingPositionSnapshot {
            heatmap: self.raster.column(snapshot.time),
            swing_position: mapped,
            normalized_position: position,
            dominant_cycle: snapshot.dominant_cycle,
            dominant_cycle_median: snapshot.dominant_cycle_median,
            band_pass: cycle.band_pass,
            quadrature: cycle.quadrature,
            cycle_amplitude: cycle.amplitude,
            lead60: cycle.lead60,
            position_span: span,
            width,
            has_cycle_estimate: if snapshot.has_cycle_estimate { 1.0 } else { 0.0 },
        })
    }

    pub fn update_from_snapshot(&mut self, snapshot: &CoronaSnapshot) -> CoronaSwingPositionSnapshot {
        self.try_update_from_snapshot(snapshot)
            .unwrap_or_else(|err| panic!("CoronaSwingPosition::update_from_snapshot failed: {err}"))
    }

    pub fn try_update(&mut self, sample: f64, time: i64) -> Result<CoronaSwingPositionSnapshot, String> {
        let engine_snap = self.engine.try_update(sample, time)?;
        self.try_update_from_snapshot(&engine_snap)
    }

    pub fn update(&mut self, sample: f64, time: i64) -> CoronaSwingPositionSnapshot {
        self.try_update(sample, time)
            .unwrap_or_else(|err| panic!("CoronaSwingPosition::update failed: {err}"))
    }

    fn wrap_output(time: i64, snapshot: CoronaSwingPositionSnapshot) -> Output {
        vec![
            Box::new(snapshot.heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, snapshot.swing_position)),
        ]
    }
}

impl Indicator for CoronaSwingPosition {
    fn is_primed(&self) -> bool {
        self.sample_count >= FIRST_BANK_SAMPLE
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CoronaSwingPosition,
            &self.mnemonic_text,
            &self.description_text,
            &[
                OutputText {
                    mnemonic: self.mnemonic_text.clone(),
                    description: self.description_text.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_sp.clone(),
                    description: self.description_sp.clone(),
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

    const TOLERANCE: f64 = 5e-12;

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
        raster_len: usize,
        hp_cutoff: i32,
        min_period: i32,
        max_period: i32,
        expected_heatmap: &[Vec<f64>],
        expected_min: &[f64],
        expected_max: &[f64],
        expected_sp: &[f64],
        expected_norm: &[f64],
        expected_dc: &[f64],
        expected_dcm: &[f64],
        expected_bp: &[f64],
        expected_quad: &[f64],
        expected_amp: &[f64],
        expected_lead: &[f64],
        expected_span: &[f64],
        expected_width: &[f64],
        expected_has_est: &[f64],
    ) {
        let mut indicator = CoronaSwingPosition::new(&CoronaSwingPositionParams {
            raster_length: raster_len as i32,
            high_pass_filter_cutoff: hp_cutoff,
            minimal_period: min_period,
            maximal_period: max_period,
            ..Default::default()
        })
        .unwrap();

        let step = 10.0 / (raster_len as f64);
        let expected_first = -5.0 + step;
        let expected_resolution = 1.0 / step;
        let prices = testdata::test_input();

        for (index, &price) in prices.iter().enumerate() {
            let snap = indicator.update(price, index as i64);
            assert_close(snap.heatmap.parameter_first, expected_first, &format!("{case_name}[{index}].first"));
            assert_close(snap.heatmap.parameter_last, 5.0, &format!("{case_name}[{index}].last"));
            assert_close(snap.heatmap.parameter_resolution, expected_resolution, &format!("{case_name}[{index}].res"));

            let row_empty = expected_sp[index].is_nan();
            assert_eq!(indicator.is_primed(), !row_empty);
            if row_empty {
                assert!(snap.heatmap.is_empty());
            } else {
                assert_eq!(snap.heatmap.values.len(), raster_len);
                for bin_index in 0..raster_len {
                    assert_close(
                        snap.heatmap.values[bin_index],
                        expected_heatmap[index][bin_index],
                        &format!("{case_name}[{index}].hm[{bin_index}]"),
                    );
                }
            }
            assert_close(snap.heatmap.value_min, expected_min[index], &format!("{case_name}[{index}].min"));
            assert_close(snap.heatmap.value_max, expected_max[index], &format!("{case_name}[{index}].max"));
            assert_close(snap.swing_position, expected_sp[index], &format!("{case_name}[{index}].sp"));
            assert_close(snap.normalized_position, expected_norm[index], &format!("{case_name}[{index}].norm"));
            assert_close(snap.dominant_cycle, expected_dc[index], &format!("{case_name}[{index}].dc"));
            assert_close(snap.dominant_cycle_median, expected_dcm[index], &format!("{case_name}[{index}].dcm"));
            assert_close(snap.band_pass, expected_bp[index], &format!("{case_name}[{index}].bp"));
            assert_close(snap.quadrature, expected_quad[index], &format!("{case_name}[{index}].quad"));
            assert_close(snap.cycle_amplitude, expected_amp[index], &format!("{case_name}[{index}].amp"));
            assert_close(snap.lead60, expected_lead[index], &format!("{case_name}[{index}].lead"));
            assert_close(snap.position_span, expected_span[index], &format!("{case_name}[{index}].span"));
            assert_close(snap.width, expected_width[index], &format!("{case_name}[{index}].width"));
            assert_close(snap.has_cycle_estimate, expected_has_est[index], &format!("{case_name}[{index}].hce"));
        }
    }

    #[test]
    fn test_reference_cases() {
        verify_case("h30_pmin6_pmax30_r50", 50, 30, 6, 30, &expected_data::expected_h30_pmin6_pmax30_r50_heatmap(), &expected_data::expected_h30_pmin6_pmax30_r50_heatmap_min(), &expected_data::expected_h30_pmin6_pmax30_r50_heatmap_max(), &expected_data::expected_h30_pmin6_pmax30_r50_swing_position(), &expected_data::expected_h30_pmin6_pmax30_r50_normalized_position(), &expected_data::expected_h30_pmin6_pmax30_r50_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax30_r50_band_pass(), &expected_data::expected_h30_pmin6_pmax30_r50_quadrature(), &expected_data::expected_h30_pmin6_pmax30_r50_cycle_amplitude(), &expected_data::expected_h30_pmin6_pmax30_r50_lead60(), &expected_data::expected_h30_pmin6_pmax30_r50_position_span(), &expected_data::expected_h30_pmin6_pmax30_r50_width(), &expected_data::expected_h30_pmin6_pmax30_r50_has_cycle_estimate());
        verify_case("h30_pmin6_pmax30_r25", 25, 30, 6, 30, &expected_data::expected_h30_pmin6_pmax30_r25_heatmap(), &expected_data::expected_h30_pmin6_pmax30_r25_heatmap_min(), &expected_data::expected_h30_pmin6_pmax30_r25_heatmap_max(), &expected_data::expected_h30_pmin6_pmax30_r25_swing_position(), &expected_data::expected_h30_pmin6_pmax30_r25_normalized_position(), &expected_data::expected_h30_pmin6_pmax30_r25_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax30_r25_band_pass(), &expected_data::expected_h30_pmin6_pmax30_r25_quadrature(), &expected_data::expected_h30_pmin6_pmax30_r25_cycle_amplitude(), &expected_data::expected_h30_pmin6_pmax30_r25_lead60(), &expected_data::expected_h30_pmin6_pmax30_r25_position_span(), &expected_data::expected_h30_pmin6_pmax30_r25_width(), &expected_data::expected_h30_pmin6_pmax30_r25_has_cycle_estimate());
        verify_case("h30_pmin6_pmax20_r50", 50, 30, 6, 20, &expected_data::expected_h30_pmin6_pmax20_r50_heatmap(), &expected_data::expected_h30_pmin6_pmax20_r50_heatmap_min(), &expected_data::expected_h30_pmin6_pmax20_r50_heatmap_max(), &expected_data::expected_h30_pmin6_pmax20_r50_swing_position(), &expected_data::expected_h30_pmin6_pmax20_r50_normalized_position(), &expected_data::expected_h30_pmin6_pmax20_r50_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax20_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax20_r50_band_pass(), &expected_data::expected_h30_pmin6_pmax20_r50_quadrature(), &expected_data::expected_h30_pmin6_pmax20_r50_cycle_amplitude(), &expected_data::expected_h30_pmin6_pmax20_r50_lead60(), &expected_data::expected_h30_pmin6_pmax20_r50_position_span(), &expected_data::expected_h30_pmin6_pmax20_r50_width(), &expected_data::expected_h30_pmin6_pmax20_r50_has_cycle_estimate());
        verify_case("h30_pmin6_pmax20_r25", 25, 30, 6, 20, &expected_data::expected_h30_pmin6_pmax20_r25_heatmap(), &expected_data::expected_h30_pmin6_pmax20_r25_heatmap_min(), &expected_data::expected_h30_pmin6_pmax20_r25_heatmap_max(), &expected_data::expected_h30_pmin6_pmax20_r25_swing_position(), &expected_data::expected_h30_pmin6_pmax20_r25_normalized_position(), &expected_data::expected_h30_pmin6_pmax20_r25_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax20_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax20_r25_band_pass(), &expected_data::expected_h30_pmin6_pmax20_r25_quadrature(), &expected_data::expected_h30_pmin6_pmax20_r25_cycle_amplitude(), &expected_data::expected_h30_pmin6_pmax20_r25_lead60(), &expected_data::expected_h30_pmin6_pmax20_r25_position_span(), &expected_data::expected_h30_pmin6_pmax20_r25_width(), &expected_data::expected_h30_pmin6_pmax20_r25_has_cycle_estimate());
        verify_case("h30_pmin10_pmax30_r50", 50, 30, 10, 30, &expected_data::expected_h30_pmin10_pmax30_r50_heatmap(), &expected_data::expected_h30_pmin10_pmax30_r50_heatmap_min(), &expected_data::expected_h30_pmin10_pmax30_r50_heatmap_max(), &expected_data::expected_h30_pmin10_pmax30_r50_swing_position(), &expected_data::expected_h30_pmin10_pmax30_r50_normalized_position(), &expected_data::expected_h30_pmin10_pmax30_r50_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax30_r50_band_pass(), &expected_data::expected_h30_pmin10_pmax30_r50_quadrature(), &expected_data::expected_h30_pmin10_pmax30_r50_cycle_amplitude(), &expected_data::expected_h30_pmin10_pmax30_r50_lead60(), &expected_data::expected_h30_pmin10_pmax30_r50_position_span(), &expected_data::expected_h30_pmin10_pmax30_r50_width(), &expected_data::expected_h30_pmin10_pmax30_r50_has_cycle_estimate());
        verify_case("h30_pmin10_pmax30_r25", 25, 30, 10, 30, &expected_data::expected_h30_pmin10_pmax30_r25_heatmap(), &expected_data::expected_h30_pmin10_pmax30_r25_heatmap_min(), &expected_data::expected_h30_pmin10_pmax30_r25_heatmap_max(), &expected_data::expected_h30_pmin10_pmax30_r25_swing_position(), &expected_data::expected_h30_pmin10_pmax30_r25_normalized_position(), &expected_data::expected_h30_pmin10_pmax30_r25_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax30_r25_band_pass(), &expected_data::expected_h30_pmin10_pmax30_r25_quadrature(), &expected_data::expected_h30_pmin10_pmax30_r25_cycle_amplitude(), &expected_data::expected_h30_pmin10_pmax30_r25_lead60(), &expected_data::expected_h30_pmin10_pmax30_r25_position_span(), &expected_data::expected_h30_pmin10_pmax30_r25_width(), &expected_data::expected_h30_pmin10_pmax30_r25_has_cycle_estimate());
        verify_case("h30_pmin10_pmax48_r50", 50, 30, 10, 48, &expected_data::expected_h30_pmin10_pmax48_r50_heatmap(), &expected_data::expected_h30_pmin10_pmax48_r50_heatmap_min(), &expected_data::expected_h30_pmin10_pmax48_r50_heatmap_max(), &expected_data::expected_h30_pmin10_pmax48_r50_swing_position(), &expected_data::expected_h30_pmin10_pmax48_r50_normalized_position(), &expected_data::expected_h30_pmin10_pmax48_r50_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax48_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax48_r50_band_pass(), &expected_data::expected_h30_pmin10_pmax48_r50_quadrature(), &expected_data::expected_h30_pmin10_pmax48_r50_cycle_amplitude(), &expected_data::expected_h30_pmin10_pmax48_r50_lead60(), &expected_data::expected_h30_pmin10_pmax48_r50_position_span(), &expected_data::expected_h30_pmin10_pmax48_r50_width(), &expected_data::expected_h30_pmin10_pmax48_r50_has_cycle_estimate());
        verify_case("h30_pmin10_pmax48_r25", 25, 30, 10, 48, &expected_data::expected_h30_pmin10_pmax48_r25_heatmap(), &expected_data::expected_h30_pmin10_pmax48_r25_heatmap_min(), &expected_data::expected_h30_pmin10_pmax48_r25_heatmap_max(), &expected_data::expected_h30_pmin10_pmax48_r25_swing_position(), &expected_data::expected_h30_pmin10_pmax48_r25_normalized_position(), &expected_data::expected_h30_pmin10_pmax48_r25_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax48_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax48_r25_band_pass(), &expected_data::expected_h30_pmin10_pmax48_r25_quadrature(), &expected_data::expected_h30_pmin10_pmax48_r25_cycle_amplitude(), &expected_data::expected_h30_pmin10_pmax48_r25_lead60(), &expected_data::expected_h30_pmin10_pmax48_r25_position_span(), &expected_data::expected_h30_pmin10_pmax48_r25_width(), &expected_data::expected_h30_pmin10_pmax48_r25_has_cycle_estimate());
        verify_case("h48_pmin6_pmax30_r50", 50, 48, 6, 30, &expected_data::expected_h48_pmin6_pmax30_r50_heatmap(), &expected_data::expected_h48_pmin6_pmax30_r50_heatmap_min(), &expected_data::expected_h48_pmin6_pmax30_r50_heatmap_max(), &expected_data::expected_h48_pmin6_pmax30_r50_swing_position(), &expected_data::expected_h48_pmin6_pmax30_r50_normalized_position(), &expected_data::expected_h48_pmin6_pmax30_r50_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax30_r50_band_pass(), &expected_data::expected_h48_pmin6_pmax30_r50_quadrature(), &expected_data::expected_h48_pmin6_pmax30_r50_cycle_amplitude(), &expected_data::expected_h48_pmin6_pmax30_r50_lead60(), &expected_data::expected_h48_pmin6_pmax30_r50_position_span(), &expected_data::expected_h48_pmin6_pmax30_r50_width(), &expected_data::expected_h48_pmin6_pmax30_r50_has_cycle_estimate());
        verify_case("h48_pmin6_pmax30_r25", 25, 48, 6, 30, &expected_data::expected_h48_pmin6_pmax30_r25_heatmap(), &expected_data::expected_h48_pmin6_pmax30_r25_heatmap_min(), &expected_data::expected_h48_pmin6_pmax30_r25_heatmap_max(), &expected_data::expected_h48_pmin6_pmax30_r25_swing_position(), &expected_data::expected_h48_pmin6_pmax30_r25_normalized_position(), &expected_data::expected_h48_pmin6_pmax30_r25_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax30_r25_band_pass(), &expected_data::expected_h48_pmin6_pmax30_r25_quadrature(), &expected_data::expected_h48_pmin6_pmax30_r25_cycle_amplitude(), &expected_data::expected_h48_pmin6_pmax30_r25_lead60(), &expected_data::expected_h48_pmin6_pmax30_r25_position_span(), &expected_data::expected_h48_pmin6_pmax30_r25_width(), &expected_data::expected_h48_pmin6_pmax30_r25_has_cycle_estimate());
        verify_case("h48_pmin6_pmax20_r50", 50, 48, 6, 20, &expected_data::expected_h48_pmin6_pmax20_r50_heatmap(), &expected_data::expected_h48_pmin6_pmax20_r50_heatmap_min(), &expected_data::expected_h48_pmin6_pmax20_r50_heatmap_max(), &expected_data::expected_h48_pmin6_pmax20_r50_swing_position(), &expected_data::expected_h48_pmin6_pmax20_r50_normalized_position(), &expected_data::expected_h48_pmin6_pmax20_r50_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax20_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax20_r50_band_pass(), &expected_data::expected_h48_pmin6_pmax20_r50_quadrature(), &expected_data::expected_h48_pmin6_pmax20_r50_cycle_amplitude(), &expected_data::expected_h48_pmin6_pmax20_r50_lead60(), &expected_data::expected_h48_pmin6_pmax20_r50_position_span(), &expected_data::expected_h48_pmin6_pmax20_r50_width(), &expected_data::expected_h48_pmin6_pmax20_r50_has_cycle_estimate());
        verify_case("h48_pmin6_pmax20_r25", 25, 48, 6, 20, &expected_data::expected_h48_pmin6_pmax20_r25_heatmap(), &expected_data::expected_h48_pmin6_pmax20_r25_heatmap_min(), &expected_data::expected_h48_pmin6_pmax20_r25_heatmap_max(), &expected_data::expected_h48_pmin6_pmax20_r25_swing_position(), &expected_data::expected_h48_pmin6_pmax20_r25_normalized_position(), &expected_data::expected_h48_pmin6_pmax20_r25_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax20_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax20_r25_band_pass(), &expected_data::expected_h48_pmin6_pmax20_r25_quadrature(), &expected_data::expected_h48_pmin6_pmax20_r25_cycle_amplitude(), &expected_data::expected_h48_pmin6_pmax20_r25_lead60(), &expected_data::expected_h48_pmin6_pmax20_r25_position_span(), &expected_data::expected_h48_pmin6_pmax20_r25_width(), &expected_data::expected_h48_pmin6_pmax20_r25_has_cycle_estimate());
        verify_case("h48_pmin10_pmax30_r50", 50, 48, 10, 30, &expected_data::expected_h48_pmin10_pmax30_r50_heatmap(), &expected_data::expected_h48_pmin10_pmax30_r50_heatmap_min(), &expected_data::expected_h48_pmin10_pmax30_r50_heatmap_max(), &expected_data::expected_h48_pmin10_pmax30_r50_swing_position(), &expected_data::expected_h48_pmin10_pmax30_r50_normalized_position(), &expected_data::expected_h48_pmin10_pmax30_r50_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax30_r50_band_pass(), &expected_data::expected_h48_pmin10_pmax30_r50_quadrature(), &expected_data::expected_h48_pmin10_pmax30_r50_cycle_amplitude(), &expected_data::expected_h48_pmin10_pmax30_r50_lead60(), &expected_data::expected_h48_pmin10_pmax30_r50_position_span(), &expected_data::expected_h48_pmin10_pmax30_r50_width(), &expected_data::expected_h48_pmin10_pmax30_r50_has_cycle_estimate());
        verify_case("h48_pmin10_pmax30_r25", 25, 48, 10, 30, &expected_data::expected_h48_pmin10_pmax30_r25_heatmap(), &expected_data::expected_h48_pmin10_pmax30_r25_heatmap_min(), &expected_data::expected_h48_pmin10_pmax30_r25_heatmap_max(), &expected_data::expected_h48_pmin10_pmax30_r25_swing_position(), &expected_data::expected_h48_pmin10_pmax30_r25_normalized_position(), &expected_data::expected_h48_pmin10_pmax30_r25_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax30_r25_band_pass(), &expected_data::expected_h48_pmin10_pmax30_r25_quadrature(), &expected_data::expected_h48_pmin10_pmax30_r25_cycle_amplitude(), &expected_data::expected_h48_pmin10_pmax30_r25_lead60(), &expected_data::expected_h48_pmin10_pmax30_r25_position_span(), &expected_data::expected_h48_pmin10_pmax30_r25_width(), &expected_data::expected_h48_pmin10_pmax30_r25_has_cycle_estimate());
        verify_case("h48_pmin10_pmax48_r50", 50, 48, 10, 48, &expected_data::expected_h48_pmin10_pmax48_r50_heatmap(), &expected_data::expected_h48_pmin10_pmax48_r50_heatmap_min(), &expected_data::expected_h48_pmin10_pmax48_r50_heatmap_max(), &expected_data::expected_h48_pmin10_pmax48_r50_swing_position(), &expected_data::expected_h48_pmin10_pmax48_r50_normalized_position(), &expected_data::expected_h48_pmin10_pmax48_r50_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax48_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax48_r50_band_pass(), &expected_data::expected_h48_pmin10_pmax48_r50_quadrature(), &expected_data::expected_h48_pmin10_pmax48_r50_cycle_amplitude(), &expected_data::expected_h48_pmin10_pmax48_r50_lead60(), &expected_data::expected_h48_pmin10_pmax48_r50_position_span(), &expected_data::expected_h48_pmin10_pmax48_r50_width(), &expected_data::expected_h48_pmin10_pmax48_r50_has_cycle_estimate());
        verify_case("h48_pmin10_pmax48_r25", 25, 48, 10, 48, &expected_data::expected_h48_pmin10_pmax48_r25_heatmap(), &expected_data::expected_h48_pmin10_pmax48_r25_heatmap_min(), &expected_data::expected_h48_pmin10_pmax48_r25_heatmap_max(), &expected_data::expected_h48_pmin10_pmax48_r25_swing_position(), &expected_data::expected_h48_pmin10_pmax48_r25_normalized_position(), &expected_data::expected_h48_pmin10_pmax48_r25_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax48_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax48_r25_band_pass(), &expected_data::expected_h48_pmin10_pmax48_r25_quadrature(), &expected_data::expected_h48_pmin10_pmax48_r25_cycle_amplitude(), &expected_data::expected_h48_pmin10_pmax48_r25_lead60(), &expected_data::expected_h48_pmin10_pmax48_r25_position_span(), &expected_data::expected_h48_pmin10_pmax48_r25_width(), &expected_data::expected_h48_pmin10_pmax48_r25_has_cycle_estimate());
    }

    #[test]
    fn test_metadata_mnemonics_validation_and_entities() {
        let mut instance = CoronaSwingPosition::new_default().unwrap();
        let md = instance.metadata();
        assert_eq!(md.identifier, Identifier::CoronaSwingPosition);
        assert_eq!(md.mnemonic, "cswing(50, 20, -5, 5, 30)");
        assert_eq!(md.outputs.len(), 2);
        assert_eq!(md.outputs[0].shape, Shape::Heatmap);
        assert_eq!(md.outputs[1].shape, Shape::Scalar);
        assert_eq!(md.outputs[1].mnemonic, "cswing-sp(30)");

        let flagged = CoronaSwingPosition::new(&CoronaSwingPositionParams {
            minimal_period: 10,
            maximal_period: 48,
            lead_window: 25,
            position_window: 15,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        })
        .unwrap();
        assert_eq!(flagged.metadata().mnemonic, "cswing(50, 20, -5, 5, 30, pmin=10, pmax=48, lead=25, pos=15, hl/2)");
        assert_eq!(flagged.metadata().outputs[1].mnemonic, "cswing-sp(30, pmin=10, pmax=48, lead=25, pos=15, hl/2)");

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
            let cfg = CoronaSwingPositionParams {
                bar_component: bc,
                quote_component: qc,
                trade_component: tc,
                ..Default::default()
            };
            let ind = CoronaSwingPosition::new(&cfg).unwrap();
            assert_eq!(ind.metadata().mnemonic, mnemonic(&cfg));
        }

        assert!(CoronaSwingPosition::new(&CoronaSwingPositionParams { lead_window: 1, ..Default::default() }).is_err());
        assert!(CoronaSwingPosition::new(&CoronaSwingPositionParams { position_window: 1, ..Default::default() }).is_err());
        assert!(CoronaSwingPosition::new(&CoronaSwingPositionParams { max_raster_value: 6.0, ..Default::default() }).is_err());
        assert!(CoronaSwingPosition::new(&CoronaSwingPositionParams { raster_length: 0, ..Default::default() }).is_err());

        for index in 0..20 {
            let out = instance.update_scalar(&Scalar::new(index as i64, 100.0));
            if index >= 12 {
                let sp = out[1].downcast_ref::<Scalar>().unwrap().value;
                assert_close(sp, 0.0, "flat swing scalar");
                let hm = out[0].downcast_ref::<Heatmap>().unwrap();
                assert!(hm.values.iter().all(|&cell| cell == 20.0));
            }
        }
        instance.reset();
        assert!(!instance.is_primed());
    }
}
