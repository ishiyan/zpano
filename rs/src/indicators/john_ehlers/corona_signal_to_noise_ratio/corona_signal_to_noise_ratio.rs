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
    accept_snapshot, checked, clip, format_number, valid_price, validate_finite, Corona,
    CoronaParams, CoronaSnapshot, Raster, RasterKind, FIRST_BANK_SAMPLE,
};

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct CoronaSignalToNoiseRatioParams {
    pub raster_length: i32,
    pub max_raster_value: f64,
    pub min_parameter_value: f64,
    pub max_parameter_value: f64,
    pub high_pass_filter_cutoff: i32,
    pub minimal_period: i32,
    pub maximal_period: i32,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

pub type Params = CoronaSignalToNoiseRatioParams;

impl Default for CoronaSignalToNoiseRatioParams {
    fn default() -> Self {
        Self {
            raster_length: 50,
            max_raster_value: 20.0,
            min_parameter_value: 1.0,
            max_parameter_value: 11.0,
            high_pass_filter_cutoff: 30,
            minimal_period: 6,
            maximal_period: 30,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

pub fn default_params() -> CoronaSignalToNoiseRatioParams {
    CoronaSignalToNoiseRatioParams::default()
}

fn snr_flags(params: &CoronaSignalToNoiseRatioParams) -> String {
    let mut flags = String::new();
    if params.minimal_period != 6 {
        flags.push_str(&format!(", pmin={}", params.minimal_period));
    }
    if params.maximal_period != 30 {
        flags.push_str(&format!(", pmax={}", params.maximal_period));
    }
    flags
}

pub fn mnemonic(params: &CoronaSignalToNoiseRatioParams) -> String {
    let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let comp = component_triple_mnemonic(bc, qc, tc);
    let flags = snr_flags(params);
    format!(
        "csnr({}, {}, {}, {}, {}{}{})",
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
pub enum CoronaSignalToNoiseRatioOutput {
    Value = 1,
    SignalToNoiseRatio = 2,
}

#[derive(Debug, Clone)]
pub struct CoronaSignalToNoiseRatioSnapshot {
    pub heatmap: Heatmap,
    pub signal_to_noise_ratio: f64,
    pub physical_decibels: f64,
    pub calibrated_decibels: f64,
    pub normalized_position: f64,
    pub signal: f64,
    pub noise: f64,
    pub median_bar_range: f64,
    pub width: f64,
    pub dominant_cycle: f64,
    pub dominant_cycle_median: f64,
    pub has_cycle_estimate: f64,
    pub ratio_is_finite: f64,
}

pub type Snapshot = CoronaSignalToNoiseRatioSnapshot;

fn validate_bar_range(bar_range: f64) -> Result<f64, String> {
    if bar_range.is_nan() {
        return Ok(f64::NAN);
    }
    if !bar_range.is_finite() || bar_range < 0.0 {
        return Err("bar_range must be finite and nonnegative".to_string());
    }
    Ok(bar_range)
}

fn validate_high_low(low: f64, high: f64) -> Result<f64, String> {
    if low.is_nan() || high.is_nan() {
        return Ok(f64::NAN);
    }
    if !low.is_finite() || !high.is_finite() {
        return Err("high and low must be finite".to_string());
    }
    let span = high - low;
    if !span.is_finite() || span < 0.0 {
        return Err("require high >= low with finite difference".to_string());
    }
    Ok(span)
}

pub struct CoronaSignalToNoiseRatio {
    mnemonic_text: String,
    description_text: String,
    mnemonic_snr: String,
    description_snr: String,
    engine: Corona,
    engine_params: CoronaParams,
    raster: Raster,
    sample_count: usize,
    ranges: [f64; 5],
    average_price: f64,
    signal: f64,
    noise: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CoronaSignalToNoiseRatio {
    pub fn new(params: &CoronaSignalToNoiseRatioParams) -> Result<Self, String> {
        let ceiling = validate_finite("max_raster_value", params.max_raster_value)?;
        if ceiling <= 6.0 {
            return Err("invalid corona signal to noise ratio parameters: max_raster_value must exceed 6".to_string());
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
        let flags = snr_flags(params);

        let mnemonic_text = format!(
            "csnr({}, {}, {}, {}, {}{}{})",
            raster.length(),
            format_number(raster.ceiling()),
            format_number(raster.lower()),
            format_number(raster.upper()),
            params.high_pass_filter_cutoff,
            flags,
            comp
        );
        let mnemonic_snr = format!("csnr-snr({}{}{})", params.high_pass_filter_cutoff, flags, comp);

        Ok(Self {
            description_text: format!("Corona signal to noise ratio {mnemonic_text}"),
            mnemonic_text,
            description_snr: format!("Corona signal to noise ratio scalar {mnemonic_snr}"),
            mnemonic_snr,
            engine,
            engine_params,
            raster,
            sample_count: 0,
            ranges: [0.0; 5],
            average_price: 0.0,
            signal: 0.0,
            noise: 0.0,
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
        })
    }

    pub fn new_default() -> Result<Self, String> {
        Self::new(&CoronaSignalToNoiseRatioParams::default())
    }

    pub fn reset(&mut self) {
        self.engine.reset();
        self.raster.reset();
        self.sample_count = 0;
        self.ranges = [0.0; 5];
        self.average_price = 0.0;
        self.signal = 0.0;
        self.noise = 0.0;
    }

    fn empty_snapshot(&self, time: i64) -> CoronaSignalToNoiseRatioSnapshot {
        CoronaSignalToNoiseRatioSnapshot {
            heatmap: self.raster.empty(time),
            signal_to_noise_ratio: f64::NAN,
            physical_decibels: f64::NAN,
            calibrated_decibels: f64::NAN,
            normalized_position: f64::NAN,
            signal: f64::NAN,
            noise: f64::NAN,
            median_bar_range: f64::NAN,
            width: f64::NAN,
            dominant_cycle: f64::NAN,
            dominant_cycle_median: f64::NAN,
            has_cycle_estimate: f64::NAN,
            ratio_is_finite: f64::NAN,
        }
    }

    pub fn try_update_from_snapshot_range(
        &mut self,
        snapshot: &CoronaSnapshot,
        bar_range: f64,
    ) -> Result<CoronaSignalToNoiseRatioSnapshot, String> {
        let accepted = accept_snapshot(snapshot, &self.engine_params, self.sample_count + 1)?;
        if !accepted {
            return Ok(self.empty_snapshot(snapshot.time));
        }
        let range_value = validate_bar_range(bar_range)?;
        if range_value.is_nan() {
            return Err("accepted shared snapshot requires a finite bar range".to_string());
        }
        self.sample_count = snapshot.sample_count;

        self.ranges.copy_within(1..5, 0);
        self.ranges[4] = range_value;
        let mut sorted_ranges = self.ranges;
        sorted_ranges.sort_by(|left, right| left.total_cmp(right));
        let median_range = sorted_ranges[2];

        if self.sample_count == 1 {
            self.average_price = snapshot.price;
        } else {
            self.average_price = checked(0.1 * snapshot.price + 0.9 * self.average_price)?;
            if self.average_price != 0.0 || snapshot.maximum_power > 0.0 {
                self.signal = checked(0.2 * snapshot.maximum_power.sqrt() + 0.9 * self.signal)?;
            }
            if self.average_price != 0.0 && self.sample_count > 5 {
                self.noise = checked(0.1 * median_range + 0.9 * self.noise)?;
            }
        }

        let physical_decibels = if self.signal > 0.0 && self.noise > 0.0 {
            20.0 * (self.signal.log10() - self.noise.log10())
        } else if self.signal > 0.0 {
            f64::INFINITY
        } else if self.noise > 0.0 {
            f64::NEG_INFINITY
        } else {
            f64::NAN
        };

        let calibrated_decibels = physical_decibels + 3.5;
        let position = if calibrated_decibels.is_nan() {
            0.0
        } else {
            clip(calibrated_decibels, 0.0, 10.0) / 10.0
        };
        let width = (0.2 - 0.4 * position).max(0.0);
        self.raster
            .update(position, width, RasterKind::Snr, position > 0.5);

        if !snapshot.primed {
            return Ok(self.empty_snapshot(snapshot.time));
        }

        let mapped = self.raster.lower() + (self.raster.upper() - self.raster.lower()) * position;
        Ok(CoronaSignalToNoiseRatioSnapshot {
            heatmap: self.raster.column(snapshot.time),
            signal_to_noise_ratio: mapped,
            physical_decibels,
            calibrated_decibels,
            normalized_position: position,
            signal: self.signal,
            noise: self.noise,
            median_bar_range: median_range,
            width,
            dominant_cycle: snapshot.dominant_cycle,
            dominant_cycle_median: snapshot.dominant_cycle_median,
            has_cycle_estimate: if snapshot.has_cycle_estimate { 1.0 } else { 0.0 },
            ratio_is_finite: if physical_decibels.is_finite() { 1.0 } else { 0.0 },
        })
    }

    pub fn update_from_snapshot_range(
        &mut self,
        snapshot: &CoronaSnapshot,
        bar_range: f64,
    ) -> CoronaSignalToNoiseRatioSnapshot {
        self.try_update_from_snapshot_range(snapshot, bar_range)
            .unwrap_or_else(|err| {
                panic!("CoronaSignalToNoiseRatio::update_from_snapshot_range failed: {err}")
            })
    }

    pub fn try_update_from_snapshot(
        &mut self,
        snapshot: &CoronaSnapshot,
        sample_low: f64,
        sample_high: f64,
    ) -> Result<CoronaSignalToNoiseRatioSnapshot, String> {
        if !snapshot.accepted {
            return self.try_update_from_snapshot_range(snapshot, 0.0);
        }
        let range_value = validate_high_low(sample_low, sample_high)?;
        self.try_update_from_snapshot_range(snapshot, range_value)
    }

    pub fn update_from_snapshot(
        &mut self,
        snapshot: &CoronaSnapshot,
        sample_low: f64,
        sample_high: f64,
    ) -> CoronaSignalToNoiseRatioSnapshot {
        self.try_update_from_snapshot(snapshot, sample_low, sample_high)
            .unwrap_or_else(|err| {
                panic!("CoronaSignalToNoiseRatio::update_from_snapshot failed: {err}")
            })
    }

    pub fn try_update_range(
        &mut self,
        sample: f64,
        bar_range: f64,
        time: i64,
    ) -> Result<CoronaSignalToNoiseRatioSnapshot, String> {
        let price_ok = valid_price(sample)?;
        if !price_ok {
            return Ok(self.empty_snapshot(time));
        }
        let range_value = validate_bar_range(bar_range)?;
        if range_value.is_nan() {
            return Ok(self.empty_snapshot(time));
        }
        let engine_snap = self.engine.try_update(sample, time)?;
        self.try_update_from_snapshot_range(&engine_snap, range_value)
    }

    pub fn update_range(
        &mut self,
        sample: f64,
        bar_range: f64,
        time: i64,
    ) -> CoronaSignalToNoiseRatioSnapshot {
        self.try_update_range(sample, bar_range, time)
            .unwrap_or_else(|err| panic!("CoronaSignalToNoiseRatio::update_range failed: {err}"))
    }

    pub fn try_update(
        &mut self,
        sample: f64,
        sample_low: f64,
        sample_high: f64,
        time: i64,
    ) -> Result<CoronaSignalToNoiseRatioSnapshot, String> {
        let price_ok = valid_price(sample)?;
        if !price_ok {
            return Ok(self.empty_snapshot(time));
        }
        let range_value = validate_high_low(sample_low, sample_high)?;
        if range_value.is_nan() {
            return Ok(self.empty_snapshot(time));
        }
        let engine_snap = self.engine.try_update(sample, time)?;
        self.try_update_from_snapshot_range(&engine_snap, range_value)
    }

    pub fn update(
        &mut self,
        sample: f64,
        sample_low: f64,
        sample_high: f64,
        time: i64,
    ) -> CoronaSignalToNoiseRatioSnapshot {
        self.try_update(sample, sample_low, sample_high, time)
            .unwrap_or_else(|err| panic!("CoronaSignalToNoiseRatio::update failed: {err}"))
    }

    fn wrap_output(time: i64, snapshot: CoronaSignalToNoiseRatioSnapshot) -> Output {
        vec![
            Box::new(snapshot.heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, snapshot.signal_to_noise_ratio)),
        ]
    }
}

impl Indicator for CoronaSignalToNoiseRatio {
    fn is_primed(&self) -> bool {
        self.sample_count >= FIRST_BANK_SAMPLE
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CoronaSignalToNoiseRatio,
            &self.mnemonic_text,
            &self.description_text,
            &[
                OutputText {
                    mnemonic: self.mnemonic_text.clone(),
                    description: self.description_text.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_snr.clone(),
                    description: self.description_snr.clone(),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        Self::wrap_output(
            sample.time,
            self.update(sample.value, sample.value, sample.value, sample.time),
        )
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        Self::wrap_output(bar.time, self.update(sample, bar.low, bar.high, bar.time))
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.quote_func)(quote);
        Self::wrap_output(quote.time, self.update(sample, sample, sample, quote.time))
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.trade_func)(trade);
        Self::wrap_output(trade.time, self.update(sample, sample, sample, trade.time))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::{expected_data, testdata};
    use crate::indicators::core::outputs::shape::Shape;
    use crate::indicators::john_ehlers::corona::TWO_PI;

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
        expected_snr: &[f64],
        expected_phys: &[f64],
        expected_cal: &[f64],
        expected_norm: &[f64],
        expected_sig: &[f64],
        expected_noise: &[f64],
        expected_med_range: &[f64],
        expected_width: &[f64],
        expected_dc: &[f64],
        expected_dcm: &[f64],
        expected_has_est: &[f64],
        expected_finite: &[f64],
    ) {
        let mut indicator = CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams {
            raster_length: raster_len as i32,
            high_pass_filter_cutoff: hp_cutoff,
            minimal_period: min_period,
            maximal_period: max_period,
            ..Default::default()
        })
        .unwrap();

        let step = 10.0 / (raster_len as f64);
        let expected_first = 1.0 + step;
        let expected_resolution = 1.0 / step;
        let prices = testdata::test_input();
        let highs = testdata::test_high();
        let lows = testdata::test_low();

        for (index, &price) in prices.iter().enumerate() {
            let snap = indicator.update(price, lows[index], highs[index], index as i64);
            assert_close(snap.heatmap.parameter_first, expected_first, &format!("{case_name}[{index}].first"));
            assert_close(snap.heatmap.parameter_last, 11.0, &format!("{case_name}[{index}].last"));
            assert_close(snap.heatmap.parameter_resolution, expected_resolution, &format!("{case_name}[{index}].res"));

            let row_empty = expected_snr[index].is_nan();
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
            assert_close(snap.signal_to_noise_ratio, expected_snr[index], &format!("{case_name}[{index}].snr"));
            assert_close(snap.physical_decibels, expected_phys[index], &format!("{case_name}[{index}].phys"));
            assert_close(snap.calibrated_decibels, expected_cal[index], &format!("{case_name}[{index}].cal"));
            assert_close(snap.normalized_position, expected_norm[index], &format!("{case_name}[{index}].norm"));
            assert_close(snap.signal, expected_sig[index], &format!("{case_name}[{index}].sig"));
            assert_close(snap.noise, expected_noise[index], &format!("{case_name}[{index}].noise"));
            assert_close(snap.median_bar_range, expected_med_range[index], &format!("{case_name}[{index}].medr"));
            assert_close(snap.width, expected_width[index], &format!("{case_name}[{index}].width"));
            assert_close(snap.dominant_cycle, expected_dc[index], &format!("{case_name}[{index}].dc"));
            assert_close(snap.dominant_cycle_median, expected_dcm[index], &format!("{case_name}[{index}].dcm"));
            assert_close(snap.has_cycle_estimate, expected_has_est[index], &format!("{case_name}[{index}].hce"));
            assert_close(snap.ratio_is_finite, expected_finite[index], &format!("{case_name}[{index}].fin"));
        }
    }

    #[test]
    fn test_reference_cases() {
        verify_case("h30_pmin6_pmax30_r50", 50, 30, 6, 30, &expected_data::expected_h30_pmin6_pmax30_r50_heatmap(), &expected_data::expected_h30_pmin6_pmax30_r50_heatmap_min(), &expected_data::expected_h30_pmin6_pmax30_r50_heatmap_max(), &expected_data::expected_h30_pmin6_pmax30_r50_signal_to_noise_ratio(), &expected_data::expected_h30_pmin6_pmax30_r50_physical_decibels(), &expected_data::expected_h30_pmin6_pmax30_r50_calibrated_decibels(), &expected_data::expected_h30_pmin6_pmax30_r50_normalized_position(), &expected_data::expected_h30_pmin6_pmax30_r50_signal(), &expected_data::expected_h30_pmin6_pmax30_r50_noise(), &expected_data::expected_h30_pmin6_pmax30_r50_median_bar_range(), &expected_data::expected_h30_pmin6_pmax30_r50_width(), &expected_data::expected_h30_pmin6_pmax30_r50_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax30_r50_has_cycle_estimate(), &expected_data::expected_h30_pmin6_pmax30_r50_ratio_is_finite());
        verify_case("h30_pmin6_pmax30_r25", 25, 30, 6, 30, &expected_data::expected_h30_pmin6_pmax30_r25_heatmap(), &expected_data::expected_h30_pmin6_pmax30_r25_heatmap_min(), &expected_data::expected_h30_pmin6_pmax30_r25_heatmap_max(), &expected_data::expected_h30_pmin6_pmax30_r25_signal_to_noise_ratio(), &expected_data::expected_h30_pmin6_pmax30_r25_physical_decibels(), &expected_data::expected_h30_pmin6_pmax30_r25_calibrated_decibels(), &expected_data::expected_h30_pmin6_pmax30_r25_normalized_position(), &expected_data::expected_h30_pmin6_pmax30_r25_signal(), &expected_data::expected_h30_pmin6_pmax30_r25_noise(), &expected_data::expected_h30_pmin6_pmax30_r25_median_bar_range(), &expected_data::expected_h30_pmin6_pmax30_r25_width(), &expected_data::expected_h30_pmin6_pmax30_r25_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax30_r25_has_cycle_estimate(), &expected_data::expected_h30_pmin6_pmax30_r25_ratio_is_finite());
        verify_case("h30_pmin6_pmax20_r50", 50, 30, 6, 20, &expected_data::expected_h30_pmin6_pmax20_r50_heatmap(), &expected_data::expected_h30_pmin6_pmax20_r50_heatmap_min(), &expected_data::expected_h30_pmin6_pmax20_r50_heatmap_max(), &expected_data::expected_h30_pmin6_pmax20_r50_signal_to_noise_ratio(), &expected_data::expected_h30_pmin6_pmax20_r50_physical_decibels(), &expected_data::expected_h30_pmin6_pmax20_r50_calibrated_decibels(), &expected_data::expected_h30_pmin6_pmax20_r50_normalized_position(), &expected_data::expected_h30_pmin6_pmax20_r50_signal(), &expected_data::expected_h30_pmin6_pmax20_r50_noise(), &expected_data::expected_h30_pmin6_pmax20_r50_median_bar_range(), &expected_data::expected_h30_pmin6_pmax20_r50_width(), &expected_data::expected_h30_pmin6_pmax20_r50_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax20_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax20_r50_has_cycle_estimate(), &expected_data::expected_h30_pmin6_pmax20_r50_ratio_is_finite());
        verify_case("h30_pmin6_pmax20_r25", 25, 30, 6, 20, &expected_data::expected_h30_pmin6_pmax20_r25_heatmap(), &expected_data::expected_h30_pmin6_pmax20_r25_heatmap_min(), &expected_data::expected_h30_pmin6_pmax20_r25_heatmap_max(), &expected_data::expected_h30_pmin6_pmax20_r25_signal_to_noise_ratio(), &expected_data::expected_h30_pmin6_pmax20_r25_physical_decibels(), &expected_data::expected_h30_pmin6_pmax20_r25_calibrated_decibels(), &expected_data::expected_h30_pmin6_pmax20_r25_normalized_position(), &expected_data::expected_h30_pmin6_pmax20_r25_signal(), &expected_data::expected_h30_pmin6_pmax20_r25_noise(), &expected_data::expected_h30_pmin6_pmax20_r25_median_bar_range(), &expected_data::expected_h30_pmin6_pmax20_r25_width(), &expected_data::expected_h30_pmin6_pmax20_r25_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax20_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax20_r25_has_cycle_estimate(), &expected_data::expected_h30_pmin6_pmax20_r25_ratio_is_finite());
        verify_case("h30_pmin10_pmax30_r50", 50, 30, 10, 30, &expected_data::expected_h30_pmin10_pmax30_r50_heatmap(), &expected_data::expected_h30_pmin10_pmax30_r50_heatmap_min(), &expected_data::expected_h30_pmin10_pmax30_r50_heatmap_max(), &expected_data::expected_h30_pmin10_pmax30_r50_signal_to_noise_ratio(), &expected_data::expected_h30_pmin10_pmax30_r50_physical_decibels(), &expected_data::expected_h30_pmin10_pmax30_r50_calibrated_decibels(), &expected_data::expected_h30_pmin10_pmax30_r50_normalized_position(), &expected_data::expected_h30_pmin10_pmax30_r50_signal(), &expected_data::expected_h30_pmin10_pmax30_r50_noise(), &expected_data::expected_h30_pmin10_pmax30_r50_median_bar_range(), &expected_data::expected_h30_pmin10_pmax30_r50_width(), &expected_data::expected_h30_pmin10_pmax30_r50_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax30_r50_has_cycle_estimate(), &expected_data::expected_h30_pmin10_pmax30_r50_ratio_is_finite());
        verify_case("h30_pmin10_pmax30_r25", 25, 30, 10, 30, &expected_data::expected_h30_pmin10_pmax30_r25_heatmap(), &expected_data::expected_h30_pmin10_pmax30_r25_heatmap_min(), &expected_data::expected_h30_pmin10_pmax30_r25_heatmap_max(), &expected_data::expected_h30_pmin10_pmax30_r25_signal_to_noise_ratio(), &expected_data::expected_h30_pmin10_pmax30_r25_physical_decibels(), &expected_data::expected_h30_pmin10_pmax30_r25_calibrated_decibels(), &expected_data::expected_h30_pmin10_pmax30_r25_normalized_position(), &expected_data::expected_h30_pmin10_pmax30_r25_signal(), &expected_data::expected_h30_pmin10_pmax30_r25_noise(), &expected_data::expected_h30_pmin10_pmax30_r25_median_bar_range(), &expected_data::expected_h30_pmin10_pmax30_r25_width(), &expected_data::expected_h30_pmin10_pmax30_r25_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax30_r25_has_cycle_estimate(), &expected_data::expected_h30_pmin10_pmax30_r25_ratio_is_finite());
        verify_case("h30_pmin10_pmax48_r50", 50, 30, 10, 48, &expected_data::expected_h30_pmin10_pmax48_r50_heatmap(), &expected_data::expected_h30_pmin10_pmax48_r50_heatmap_min(), &expected_data::expected_h30_pmin10_pmax48_r50_heatmap_max(), &expected_data::expected_h30_pmin10_pmax48_r50_signal_to_noise_ratio(), &expected_data::expected_h30_pmin10_pmax48_r50_physical_decibels(), &expected_data::expected_h30_pmin10_pmax48_r50_calibrated_decibels(), &expected_data::expected_h30_pmin10_pmax48_r50_normalized_position(), &expected_data::expected_h30_pmin10_pmax48_r50_signal(), &expected_data::expected_h30_pmin10_pmax48_r50_noise(), &expected_data::expected_h30_pmin10_pmax48_r50_median_bar_range(), &expected_data::expected_h30_pmin10_pmax48_r50_width(), &expected_data::expected_h30_pmin10_pmax48_r50_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax48_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax48_r50_has_cycle_estimate(), &expected_data::expected_h30_pmin10_pmax48_r50_ratio_is_finite());
        verify_case("h30_pmin10_pmax48_r25", 25, 30, 10, 48, &expected_data::expected_h30_pmin10_pmax48_r25_heatmap(), &expected_data::expected_h30_pmin10_pmax48_r25_heatmap_min(), &expected_data::expected_h30_pmin10_pmax48_r25_heatmap_max(), &expected_data::expected_h30_pmin10_pmax48_r25_signal_to_noise_ratio(), &expected_data::expected_h30_pmin10_pmax48_r25_physical_decibels(), &expected_data::expected_h30_pmin10_pmax48_r25_calibrated_decibels(), &expected_data::expected_h30_pmin10_pmax48_r25_normalized_position(), &expected_data::expected_h30_pmin10_pmax48_r25_signal(), &expected_data::expected_h30_pmin10_pmax48_r25_noise(), &expected_data::expected_h30_pmin10_pmax48_r25_median_bar_range(), &expected_data::expected_h30_pmin10_pmax48_r25_width(), &expected_data::expected_h30_pmin10_pmax48_r25_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax48_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax48_r25_has_cycle_estimate(), &expected_data::expected_h30_pmin10_pmax48_r25_ratio_is_finite());
        verify_case("h48_pmin6_pmax30_r50", 50, 48, 6, 30, &expected_data::expected_h48_pmin6_pmax30_r50_heatmap(), &expected_data::expected_h48_pmin6_pmax30_r50_heatmap_min(), &expected_data::expected_h48_pmin6_pmax30_r50_heatmap_max(), &expected_data::expected_h48_pmin6_pmax30_r50_signal_to_noise_ratio(), &expected_data::expected_h48_pmin6_pmax30_r50_physical_decibels(), &expected_data::expected_h48_pmin6_pmax30_r50_calibrated_decibels(), &expected_data::expected_h48_pmin6_pmax30_r50_normalized_position(), &expected_data::expected_h48_pmin6_pmax30_r50_signal(), &expected_data::expected_h48_pmin6_pmax30_r50_noise(), &expected_data::expected_h48_pmin6_pmax30_r50_median_bar_range(), &expected_data::expected_h48_pmin6_pmax30_r50_width(), &expected_data::expected_h48_pmin6_pmax30_r50_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax30_r50_has_cycle_estimate(), &expected_data::expected_h48_pmin6_pmax30_r50_ratio_is_finite());
        verify_case("h48_pmin6_pmax30_r25", 25, 48, 6, 30, &expected_data::expected_h48_pmin6_pmax30_r25_heatmap(), &expected_data::expected_h48_pmin6_pmax30_r25_heatmap_min(), &expected_data::expected_h48_pmin6_pmax30_r25_heatmap_max(), &expected_data::expected_h48_pmin6_pmax30_r25_signal_to_noise_ratio(), &expected_data::expected_h48_pmin6_pmax30_r25_physical_decibels(), &expected_data::expected_h48_pmin6_pmax30_r25_calibrated_decibels(), &expected_data::expected_h48_pmin6_pmax30_r25_normalized_position(), &expected_data::expected_h48_pmin6_pmax30_r25_signal(), &expected_data::expected_h48_pmin6_pmax30_r25_noise(), &expected_data::expected_h48_pmin6_pmax30_r25_median_bar_range(), &expected_data::expected_h48_pmin6_pmax30_r25_width(), &expected_data::expected_h48_pmin6_pmax30_r25_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax30_r25_has_cycle_estimate(), &expected_data::expected_h48_pmin6_pmax30_r25_ratio_is_finite());
        verify_case("h48_pmin6_pmax20_r50", 50, 48, 6, 20, &expected_data::expected_h48_pmin6_pmax20_r50_heatmap(), &expected_data::expected_h48_pmin6_pmax20_r50_heatmap_min(), &expected_data::expected_h48_pmin6_pmax20_r50_heatmap_max(), &expected_data::expected_h48_pmin6_pmax20_r50_signal_to_noise_ratio(), &expected_data::expected_h48_pmin6_pmax20_r50_physical_decibels(), &expected_data::expected_h48_pmin6_pmax20_r50_calibrated_decibels(), &expected_data::expected_h48_pmin6_pmax20_r50_normalized_position(), &expected_data::expected_h48_pmin6_pmax20_r50_signal(), &expected_data::expected_h48_pmin6_pmax20_r50_noise(), &expected_data::expected_h48_pmin6_pmax20_r50_median_bar_range(), &expected_data::expected_h48_pmin6_pmax20_r50_width(), &expected_data::expected_h48_pmin6_pmax20_r50_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax20_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax20_r50_has_cycle_estimate(), &expected_data::expected_h48_pmin6_pmax20_r50_ratio_is_finite());
        verify_case("h48_pmin6_pmax20_r25", 25, 48, 6, 20, &expected_data::expected_h48_pmin6_pmax20_r25_heatmap(), &expected_data::expected_h48_pmin6_pmax20_r25_heatmap_min(), &expected_data::expected_h48_pmin6_pmax20_r25_heatmap_max(), &expected_data::expected_h48_pmin6_pmax20_r25_signal_to_noise_ratio(), &expected_data::expected_h48_pmin6_pmax20_r25_physical_decibels(), &expected_data::expected_h48_pmin6_pmax20_r25_calibrated_decibels(), &expected_data::expected_h48_pmin6_pmax20_r25_normalized_position(), &expected_data::expected_h48_pmin6_pmax20_r25_signal(), &expected_data::expected_h48_pmin6_pmax20_r25_noise(), &expected_data::expected_h48_pmin6_pmax20_r25_median_bar_range(), &expected_data::expected_h48_pmin6_pmax20_r25_width(), &expected_data::expected_h48_pmin6_pmax20_r25_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax20_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax20_r25_has_cycle_estimate(), &expected_data::expected_h48_pmin6_pmax20_r25_ratio_is_finite());
        verify_case("h48_pmin10_pmax30_r50", 50, 48, 10, 30, &expected_data::expected_h48_pmin10_pmax30_r50_heatmap(), &expected_data::expected_h48_pmin10_pmax30_r50_heatmap_min(), &expected_data::expected_h48_pmin10_pmax30_r50_heatmap_max(), &expected_data::expected_h48_pmin10_pmax30_r50_signal_to_noise_ratio(), &expected_data::expected_h48_pmin10_pmax30_r50_physical_decibels(), &expected_data::expected_h48_pmin10_pmax30_r50_calibrated_decibels(), &expected_data::expected_h48_pmin10_pmax30_r50_normalized_position(), &expected_data::expected_h48_pmin10_pmax30_r50_signal(), &expected_data::expected_h48_pmin10_pmax30_r50_noise(), &expected_data::expected_h48_pmin10_pmax30_r50_median_bar_range(), &expected_data::expected_h48_pmin10_pmax30_r50_width(), &expected_data::expected_h48_pmin10_pmax30_r50_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax30_r50_has_cycle_estimate(), &expected_data::expected_h48_pmin10_pmax30_r50_ratio_is_finite());
        verify_case("h48_pmin10_pmax30_r25", 25, 48, 10, 30, &expected_data::expected_h48_pmin10_pmax30_r25_heatmap(), &expected_data::expected_h48_pmin10_pmax30_r25_heatmap_min(), &expected_data::expected_h48_pmin10_pmax30_r25_heatmap_max(), &expected_data::expected_h48_pmin10_pmax30_r25_signal_to_noise_ratio(), &expected_data::expected_h48_pmin10_pmax30_r25_physical_decibels(), &expected_data::expected_h48_pmin10_pmax30_r25_calibrated_decibels(), &expected_data::expected_h48_pmin10_pmax30_r25_normalized_position(), &expected_data::expected_h48_pmin10_pmax30_r25_signal(), &expected_data::expected_h48_pmin10_pmax30_r25_noise(), &expected_data::expected_h48_pmin10_pmax30_r25_median_bar_range(), &expected_data::expected_h48_pmin10_pmax30_r25_width(), &expected_data::expected_h48_pmin10_pmax30_r25_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax30_r25_has_cycle_estimate(), &expected_data::expected_h48_pmin10_pmax30_r25_ratio_is_finite());
        verify_case("h48_pmin10_pmax48_r50", 50, 48, 10, 48, &expected_data::expected_h48_pmin10_pmax48_r50_heatmap(), &expected_data::expected_h48_pmin10_pmax48_r50_heatmap_min(), &expected_data::expected_h48_pmin10_pmax48_r50_heatmap_max(), &expected_data::expected_h48_pmin10_pmax48_r50_signal_to_noise_ratio(), &expected_data::expected_h48_pmin10_pmax48_r50_physical_decibels(), &expected_data::expected_h48_pmin10_pmax48_r50_calibrated_decibels(), &expected_data::expected_h48_pmin10_pmax48_r50_normalized_position(), &expected_data::expected_h48_pmin10_pmax48_r50_signal(), &expected_data::expected_h48_pmin10_pmax48_r50_noise(), &expected_data::expected_h48_pmin10_pmax48_r50_median_bar_range(), &expected_data::expected_h48_pmin10_pmax48_r50_width(), &expected_data::expected_h48_pmin10_pmax48_r50_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax48_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax48_r50_has_cycle_estimate(), &expected_data::expected_h48_pmin10_pmax48_r50_ratio_is_finite());
        verify_case("h48_pmin10_pmax48_r25", 25, 48, 10, 48, &expected_data::expected_h48_pmin10_pmax48_r25_heatmap(), &expected_data::expected_h48_pmin10_pmax48_r25_heatmap_min(), &expected_data::expected_h48_pmin10_pmax48_r25_heatmap_max(), &expected_data::expected_h48_pmin10_pmax48_r25_signal_to_noise_ratio(), &expected_data::expected_h48_pmin10_pmax48_r25_physical_decibels(), &expected_data::expected_h48_pmin10_pmax48_r25_calibrated_decibels(), &expected_data::expected_h48_pmin10_pmax48_r25_normalized_position(), &expected_data::expected_h48_pmin10_pmax48_r25_signal(), &expected_data::expected_h48_pmin10_pmax48_r25_noise(), &expected_data::expected_h48_pmin10_pmax48_r25_median_bar_range(), &expected_data::expected_h48_pmin10_pmax48_r25_width(), &expected_data::expected_h48_pmin10_pmax48_r25_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax48_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax48_r25_has_cycle_estimate(), &expected_data::expected_h48_pmin10_pmax48_r25_ratio_is_finite());
    }

    #[test]
    fn test_metadata_mnemonics_range_validation_and_entities() {
        let mut indicator = CoronaSignalToNoiseRatio::new_default().unwrap();
        let md = indicator.metadata();
        assert_eq!(md.identifier, Identifier::CoronaSignalToNoiseRatio);
        assert_eq!(md.mnemonic, "csnr(50, 20, 1, 11, 30)");
        assert_eq!(md.outputs.len(), 2);
        assert_eq!(md.outputs[0].shape, Shape::Heatmap);
        assert_eq!(md.outputs[1].shape, Shape::Scalar);
        assert_eq!(md.outputs[1].mnemonic, "csnr-snr(30)");

        let flagged = CoronaSignalToNoiseRatio::new(&CoronaSignalToNoiseRatioParams {
            minimal_period: 10,
            maximal_period: 48,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        })
        .unwrap();
        assert_eq!(flagged.metadata().mnemonic, "csnr(50, 20, 1, 11, 30, pmin=10, pmax=48, hl/2)");
        assert_eq!(flagged.metadata().outputs[1].mnemonic, "csnr-snr(30, pmin=10, pmax=48, hl/2)");

        assert!(indicator.try_update_range(100.0, -0.1, 0).is_err());
        assert!(indicator.try_update(100.0, 102.0, 101.0, 0).is_err());
        assert!(!indicator.is_primed());

        // NaN range skips bar without advancing sample_count
        let skipped = indicator.update(100.0, f64::NAN, 101.0, 0);
        assert!(skipped.signal_to_noise_ratio.is_nan());
        assert!(!indicator.is_primed());

        // Zero range on sinusoid -> +inf physical dB, mapped = 11.0
        let mut last_snap = indicator.update_range(100.0, 0.0, 0);
        for index in 1..40 {
            let sample = 100.0 + 2.0 * (TWO_PI * (index as f64) / 20.0).sin();
            last_snap = indicator.update_range(sample, 0.0, index as i64);
        }
        assert_eq!(last_snap.physical_decibels, f64::INFINITY);
        assert_close(last_snap.signal_to_noise_ratio, 11.0, "zero-noise snr");
        assert_eq!(last_snap.ratio_is_finite, 0.0);

        indicator.reset();
        assert!(!indicator.is_primed());
    }
}
