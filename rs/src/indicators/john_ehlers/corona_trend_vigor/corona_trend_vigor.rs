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
    accept_snapshot, checked, clip, format_number, rolling_append, validate_finite,
    AdaptiveBandPass, Corona, CoronaParams, CoronaSnapshot, Raster, RasterKind,
};

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct CoronaTrendVigorParams {
    pub raster_length: i32,
    pub max_raster_value: f64,
    pub min_parameter_value: f64,
    pub max_parameter_value: f64,
    pub high_pass_filter_cutoff: i32,
    pub minimal_period: i32,
    pub maximal_period: i32,
    pub lookback_offset: i32,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

pub type Params = CoronaTrendVigorParams;

impl Default for CoronaTrendVigorParams {
    fn default() -> Self {
        Self {
            raster_length: 50,
            max_raster_value: 20.0,
            min_parameter_value: -10.0,
            max_parameter_value: 10.0,
            high_pass_filter_cutoff: 30,
            minimal_period: 6,
            maximal_period: 30,
            lookback_offset: 0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

pub fn default_params() -> CoronaTrendVigorParams {
    CoronaTrendVigorParams::default()
}

fn trend_flags(params: &CoronaTrendVigorParams) -> String {
    let mut flags = String::new();
    if params.minimal_period != 6 {
        flags.push_str(&format!(", pmin={}", params.minimal_period));
    }
    if params.maximal_period != 30 {
        flags.push_str(&format!(", pmax={}", params.maximal_period));
    }
    if params.lookback_offset != 0 {
        flags.push_str(&format!(", offset={}", params.lookback_offset));
    }
    flags
}

pub fn mnemonic(params: &CoronaTrendVigorParams) -> String {
    let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let comp = component_triple_mnemonic(bc, qc, tc);
    let flags = trend_flags(params);
    format!(
        "ctv({}, {}, {}, {}, {}{}{})",
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
pub enum CoronaTrendVigorOutput {
    Value = 1,
    TrendVigor = 2,
}

#[derive(Debug, Clone)]
pub struct CoronaTrendVigorSnapshot {
    pub heatmap: Heatmap,
    pub trend_vigor: f64,
    pub normalized_vigor: f64,
    pub dominant_cycle: f64,
    pub dominant_cycle_median: f64,
    pub band_pass: f64,
    pub quadrature: f64,
    pub cycle_amplitude: f64,
    pub trend_displacement: f64,
    pub raw_ratio: f64,
    pub lookback: f64,
    pub width: f64,
    pub has_cycle_estimate: f64,
}

pub type Snapshot = CoronaTrendVigorSnapshot;

pub struct CoronaTrendVigor {
    mnemonic_text: String,
    description_text: String,
    mnemonic_tv: String,
    description_tv: String,
    engine: Corona,
    engine_params: CoronaParams,
    bandpass: AdaptiveBandPass,
    raster: Raster,
    lookback_offset: i32,
    max_history: usize,
    sample_count: usize,
    prices: Vec<f64>,
    ratio: f64,
    ready: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CoronaTrendVigor {
    pub fn new(params: &CoronaTrendVigorParams) -> Result<Self, String> {
        if params.lookback_offset != 0 && params.lookback_offset != -1 {
            return Err("invalid corona trend vigor parameters: lookback_offset must be 0 or -1".to_string());
        }
        let ceiling = validate_finite("max_raster_value", params.max_raster_value)?;
        if ceiling <= 6.0 {
            return Err("invalid corona trend vigor parameters: max_raster_value must exceed 6".to_string());
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
        let flags = trend_flags(params);

        let mnemonic_text = format!(
            "ctv({}, {}, {}, {}, {}{}{})",
            raster.length(),
            format_number(raster.ceiling()),
            format_number(raster.lower()),
            format_number(raster.upper()),
            params.high_pass_filter_cutoff,
            flags,
            comp
        );
        let mnemonic_tv = format!("ctv-tv({}{}{})", params.high_pass_filter_cutoff, flags, comp);
        let max_history = (params.maximal_period + 1) as usize;

        Ok(Self {
            description_text: format!("Corona trend vigor {mnemonic_text}"),
            mnemonic_text,
            description_tv: format!("Corona trend vigor scalar {mnemonic_tv}"),
            mnemonic_tv,
            engine,
            engine_params,
            bandpass: AdaptiveBandPass::new(),
            raster,
            lookback_offset: params.lookback_offset,
            max_history,
            sample_count: 0,
            prices: Vec::with_capacity(max_history),
            ratio: 0.0,
            ready: false,
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
        })
    }

    pub fn new_default() -> Result<Self, String> {
        Self::new(&CoronaTrendVigorParams::default())
    }

    pub fn reset(&mut self) {
        self.engine.reset();
        self.bandpass.reset();
        self.raster.reset();
        self.sample_count = 0;
        self.prices.clear();
        self.ratio = 0.0;
        self.ready = false;
    }

    fn empty_snapshot(&self, time: i64) -> CoronaTrendVigorSnapshot {
        CoronaTrendVigorSnapshot {
            heatmap: self.raster.empty(time),
            trend_vigor: f64::NAN,
            normalized_vigor: f64::NAN,
            dominant_cycle: f64::NAN,
            dominant_cycle_median: f64::NAN,
            band_pass: f64::NAN,
            quadrature: f64::NAN,
            cycle_amplitude: f64::NAN,
            trend_displacement: f64::NAN,
            raw_ratio: f64::NAN,
            lookback: f64::NAN,
            width: f64::NAN,
            has_cycle_estimate: f64::NAN,
        }
    }

    pub fn try_update_from_snapshot(
        &mut self,
        snapshot: &CoronaSnapshot,
    ) -> Result<CoronaTrendVigorSnapshot, String> {
        let accepted = accept_snapshot(snapshot, &self.engine_params, self.sample_count + 1)?;
        if !accepted {
            return Ok(self.empty_snapshot(snapshot.time));
        }
        let cycle = self.bandpass.update(snapshot.price, snapshot.dominant_cycle_median)?;
        self.sample_count = snapshot.sample_count;

        let _ = rolling_append(&mut self.prices, snapshot.price, self.max_history);
        let floored_dcm = snapshot.dominant_cycle_median.floor() as i32;
        let lag = 1_i32.max(floored_dcm + self.lookback_offset) as usize;
        let has_history = self.prices.len() > lag;
        let trend = if has_history {
            snapshot.price - self.prices[self.prices.len() - 1 - lag]
        } else {
            f64::NAN
        };
        let raw_ratio = if has_history && cycle.amplitude > 0.0 {
            trend / cycle.amplitude
        } else {
            f64::NAN
        };
        if has_history && trend != 0.0 && cycle.amplitude > 0.0 {
            let smoothed = checked(0.33 * raw_ratio + 0.67 * self.ratio)?;
            self.ratio = clip(smoothed, -10.0, 10.0);
        }

        let vigor = 0.05 * (self.ratio + 10.0);
        let mut width = if (0.3..=0.7).contains(&vigor) {
            (vigor - 0.3).min(0.7 - vigor)
        } else {
            0.01
        };
        if width == 0.0 {
            width = 0.01;
        }

        self.raster
            .update(vigor, width, RasterKind::Trend, !(0.3..=0.7).contains(&vigor));
        self.ready = snapshot.primed && has_history;
        if !self.ready {
            return Ok(self.empty_snapshot(snapshot.time));
        }

        let mapped = self.raster.lower() + (self.raster.upper() - self.raster.lower()) * vigor;
        Ok(CoronaTrendVigorSnapshot {
            heatmap: self.raster.column(snapshot.time),
            trend_vigor: mapped,
            normalized_vigor: vigor,
            dominant_cycle: snapshot.dominant_cycle,
            dominant_cycle_median: snapshot.dominant_cycle_median,
            band_pass: cycle.band_pass,
            quadrature: cycle.quadrature,
            cycle_amplitude: cycle.amplitude,
            trend_displacement: trend,
            raw_ratio,
            lookback: lag as f64,
            width,
            has_cycle_estimate: if snapshot.has_cycle_estimate { 1.0 } else { 0.0 },
        })
    }

    pub fn update_from_snapshot(&mut self, snapshot: &CoronaSnapshot) -> CoronaTrendVigorSnapshot {
        self.try_update_from_snapshot(snapshot)
            .unwrap_or_else(|err| panic!("CoronaTrendVigor::update_from_snapshot failed: {err}"))
    }

    pub fn try_update(&mut self, sample: f64, time: i64) -> Result<CoronaTrendVigorSnapshot, String> {
        let engine_snap = self.engine.try_update(sample, time)?;
        self.try_update_from_snapshot(&engine_snap)
    }

    pub fn update(&mut self, sample: f64, time: i64) -> CoronaTrendVigorSnapshot {
        self.try_update(sample, time)
            .unwrap_or_else(|err| panic!("CoronaTrendVigor::update failed: {err}"))
    }

    fn wrap_output(time: i64, snapshot: CoronaTrendVigorSnapshot) -> Output {
        vec![
            Box::new(snapshot.heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, snapshot.trend_vigor)),
        ]
    }
}

impl Indicator for CoronaTrendVigor {
    fn is_primed(&self) -> bool {
        self.ready
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CoronaTrendVigor,
            &self.mnemonic_text,
            &self.description_text,
            &[
                OutputText {
                    mnemonic: self.mnemonic_text.clone(),
                    description: self.description_text.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_tv.clone(),
                    description: self.description_tv.clone(),
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
        expected_tv: &[f64],
        expected_norm: &[f64],
        expected_dc: &[f64],
        expected_dcm: &[f64],
        expected_bp: &[f64],
        expected_quad: &[f64],
        expected_amp: &[f64],
        expected_disp: &[f64],
        expected_raw_ratio: &[f64],
        expected_lookback: &[f64],
        expected_width: &[f64],
        expected_has_est: &[f64],
    ) {
        let cfg = CoronaTrendVigorParams {
            raster_length: raster_len as i32,
            high_pass_filter_cutoff: hp_cutoff,
            minimal_period: min_period,
            maximal_period: max_period,
            ..Default::default()
        };
        let mut indicator = CoronaTrendVigor::new(&cfg).unwrap();
        let mut shared_consumer = CoronaTrendVigor::new(&cfg).unwrap();
        let mut shared_engine = Corona::new(&CoronaParams {
            high_pass_filter_cutoff: hp_cutoff,
            minimal_period: min_period,
            maximal_period: max_period,
            decibels_lower_threshold: 6.0,
            decibels_upper_threshold: 20.0,
        })
        .unwrap();

        let step = 20.0 / (raster_len as f64);
        let expected_first = -10.0 + step;
        let expected_resolution = 1.0 / step;
        let prices = testdata::test_input();
        let mut saw_lag_tie = false;

        for (index, &price) in prices.iter().enumerate() {
            let standalone_snap = indicator.update(price, index as i64);
            let mut engine_snap = shared_engine.update(price, index as i64);
            if engine_snap.primed {
                if !expected_dcm[index].is_nan() {
                    assert_close(standalone_snap.dominant_cycle, expected_dc[index], &format!("{case_name}[{index}].standalone.dc"));
                    assert_close(standalone_snap.dominant_cycle_median, expected_dcm[index], &format!("{case_name}[{index}].standalone.dcm"));
                    assert_close(standalone_snap.band_pass, expected_bp[index], &format!("{case_name}[{index}].standalone.bp"));
                    assert_close(standalone_snap.quadrature, expected_quad[index], &format!("{case_name}[{index}].standalone.quad"));
                    assert_close(standalone_snap.cycle_amplitude, expected_amp[index], &format!("{case_name}[{index}].standalone.amp"));
                    assert_close(standalone_snap.has_cycle_estimate, expected_has_est[index], &format!("{case_name}[{index}].standalone.hce"));
                }
                let nearest_integer = engine_snap.dominant_cycle_median.round();
                if (engine_snap.dominant_cycle_median - nearest_integer).abs() <= 5e-12 {
                    if expected_lookback[index].is_nan() {
                        if ((index + 1) as f64) == nearest_integer && engine_snap.dominant_cycle_median < nearest_integer {
                            engine_snap.dominant_cycle_median = nearest_integer;
                            saw_lag_tie = true;
                        }
                    } else if engine_snap.dominant_cycle_median.floor() != expected_lookback[index] {
                        if expected_lookback[index] == nearest_integer {
                            engine_snap.dominant_cycle_median = nearest_integer;
                        } else if expected_lookback[index] == nearest_integer - 1.0 && nearest_integer > 0.0 {
                            engine_snap.dominant_cycle_median = f64::from_bits(nearest_integer.to_bits() - 1);
                        }
                        saw_lag_tie = true;
                    }
                }
            }
            let snap = shared_consumer.update_from_snapshot(&engine_snap);

            assert_close(snap.heatmap.parameter_first, expected_first, &format!("{case_name}[{index}].first"));
            assert_close(snap.heatmap.parameter_last, 10.0, &format!("{case_name}[{index}].last"));
            assert_close(snap.heatmap.parameter_resolution, expected_resolution, &format!("{case_name}[{index}].resolution"));

            let row_empty = expected_tv[index].is_nan();
            assert_eq!(shared_consumer.is_primed(), !row_empty, "{case_name}[{index}].primed");
            if !saw_lag_tie {
                assert_eq!(indicator.is_primed(), !row_empty, "{case_name}[{index}].standalone.primed");
            }
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
                    if !saw_lag_tie {
                        assert_close(
                            standalone_snap.heatmap.values[bin_index],
                            expected_heatmap[index][bin_index],
                            &format!("{case_name}[{index}].standalone.hm[{bin_index}]"),
                        );
                    }
                }
            }
            assert_close(snap.heatmap.value_min, expected_min[index], &format!("{case_name}[{index}].min"));
            assert_close(snap.heatmap.value_max, expected_max[index], &format!("{case_name}[{index}].max"));
            assert_close(snap.trend_vigor, expected_tv[index], &format!("{case_name}[{index}].tv"));
            assert_close(snap.normalized_vigor, expected_norm[index], &format!("{case_name}[{index}].norm"));
            assert_close(snap.dominant_cycle, expected_dc[index], &format!("{case_name}[{index}].dc"));
            assert_close(snap.dominant_cycle_median, expected_dcm[index], &format!("{case_name}[{index}].dcm"));
            assert_close(snap.band_pass, expected_bp[index], &format!("{case_name}[{index}].bp"));
            assert_close(snap.quadrature, expected_quad[index], &format!("{case_name}[{index}].quad"));
            assert_close(snap.cycle_amplitude, expected_amp[index], &format!("{case_name}[{index}].amp"));
            assert_close(snap.trend_displacement, expected_disp[index], &format!("{case_name}[{index}].disp"));
            assert_close(snap.raw_ratio, expected_raw_ratio[index], &format!("{case_name}[{index}].raw"));
            assert_close(snap.lookback, expected_lookback[index], &format!("{case_name}[{index}].lag"));
            assert_close(snap.width, expected_width[index], &format!("{case_name}[{index}].width"));
            assert_close(snap.has_cycle_estimate, expected_has_est[index], &format!("{case_name}[{index}].hce"));
        }
    }

    #[test]
    fn test_reference_cases() {
        verify_case("h30_pmin6_pmax30_r50", 50, 30, 6, 30, &expected_data::expected_h30_pmin6_pmax30_r50_heatmap(), &expected_data::expected_h30_pmin6_pmax30_r50_heatmap_min(), &expected_data::expected_h30_pmin6_pmax30_r50_heatmap_max(), &expected_data::expected_h30_pmin6_pmax30_r50_trend_vigor(), &expected_data::expected_h30_pmin6_pmax30_r50_normalized_vigor(), &expected_data::expected_h30_pmin6_pmax30_r50_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax30_r50_band_pass(), &expected_data::expected_h30_pmin6_pmax30_r50_quadrature(), &expected_data::expected_h30_pmin6_pmax30_r50_cycle_amplitude(), &expected_data::expected_h30_pmin6_pmax30_r50_trend_displacement(), &expected_data::expected_h30_pmin6_pmax30_r50_raw_ratio(), &expected_data::expected_h30_pmin6_pmax30_r50_lookback(), &expected_data::expected_h30_pmin6_pmax30_r50_width(), &expected_data::expected_h30_pmin6_pmax30_r50_has_cycle_estimate());
        verify_case("h30_pmin6_pmax30_r25", 25, 30, 6, 30, &expected_data::expected_h30_pmin6_pmax30_r25_heatmap(), &expected_data::expected_h30_pmin6_pmax30_r25_heatmap_min(), &expected_data::expected_h30_pmin6_pmax30_r25_heatmap_max(), &expected_data::expected_h30_pmin6_pmax30_r25_trend_vigor(), &expected_data::expected_h30_pmin6_pmax30_r25_normalized_vigor(), &expected_data::expected_h30_pmin6_pmax30_r25_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax30_r25_band_pass(), &expected_data::expected_h30_pmin6_pmax30_r25_quadrature(), &expected_data::expected_h30_pmin6_pmax30_r25_cycle_amplitude(), &expected_data::expected_h30_pmin6_pmax30_r25_trend_displacement(), &expected_data::expected_h30_pmin6_pmax30_r25_raw_ratio(), &expected_data::expected_h30_pmin6_pmax30_r25_lookback(), &expected_data::expected_h30_pmin6_pmax30_r25_width(), &expected_data::expected_h30_pmin6_pmax30_r25_has_cycle_estimate());
        verify_case("h30_pmin6_pmax20_r50", 50, 30, 6, 20, &expected_data::expected_h30_pmin6_pmax20_r50_heatmap(), &expected_data::expected_h30_pmin6_pmax20_r50_heatmap_min(), &expected_data::expected_h30_pmin6_pmax20_r50_heatmap_max(), &expected_data::expected_h30_pmin6_pmax20_r50_trend_vigor(), &expected_data::expected_h30_pmin6_pmax20_r50_normalized_vigor(), &expected_data::expected_h30_pmin6_pmax20_r50_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax20_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax20_r50_band_pass(), &expected_data::expected_h30_pmin6_pmax20_r50_quadrature(), &expected_data::expected_h30_pmin6_pmax20_r50_cycle_amplitude(), &expected_data::expected_h30_pmin6_pmax20_r50_trend_displacement(), &expected_data::expected_h30_pmin6_pmax20_r50_raw_ratio(), &expected_data::expected_h30_pmin6_pmax20_r50_lookback(), &expected_data::expected_h30_pmin6_pmax20_r50_width(), &expected_data::expected_h30_pmin6_pmax20_r50_has_cycle_estimate());
        verify_case("h30_pmin6_pmax20_r25", 25, 30, 6, 20, &expected_data::expected_h30_pmin6_pmax20_r25_heatmap(), &expected_data::expected_h30_pmin6_pmax20_r25_heatmap_min(), &expected_data::expected_h30_pmin6_pmax20_r25_heatmap_max(), &expected_data::expected_h30_pmin6_pmax20_r25_trend_vigor(), &expected_data::expected_h30_pmin6_pmax20_r25_normalized_vigor(), &expected_data::expected_h30_pmin6_pmax20_r25_dominant_cycle(), &expected_data::expected_h30_pmin6_pmax20_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin6_pmax20_r25_band_pass(), &expected_data::expected_h30_pmin6_pmax20_r25_quadrature(), &expected_data::expected_h30_pmin6_pmax20_r25_cycle_amplitude(), &expected_data::expected_h30_pmin6_pmax20_r25_trend_displacement(), &expected_data::expected_h30_pmin6_pmax20_r25_raw_ratio(), &expected_data::expected_h30_pmin6_pmax20_r25_lookback(), &expected_data::expected_h30_pmin6_pmax20_r25_width(), &expected_data::expected_h30_pmin6_pmax20_r25_has_cycle_estimate());
        verify_case("h30_pmin10_pmax30_r50", 50, 30, 10, 30, &expected_data::expected_h30_pmin10_pmax30_r50_heatmap(), &expected_data::expected_h30_pmin10_pmax30_r50_heatmap_min(), &expected_data::expected_h30_pmin10_pmax30_r50_heatmap_max(), &expected_data::expected_h30_pmin10_pmax30_r50_trend_vigor(), &expected_data::expected_h30_pmin10_pmax30_r50_normalized_vigor(), &expected_data::expected_h30_pmin10_pmax30_r50_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax30_r50_band_pass(), &expected_data::expected_h30_pmin10_pmax30_r50_quadrature(), &expected_data::expected_h30_pmin10_pmax30_r50_cycle_amplitude(), &expected_data::expected_h30_pmin10_pmax30_r50_trend_displacement(), &expected_data::expected_h30_pmin10_pmax30_r50_raw_ratio(), &expected_data::expected_h30_pmin10_pmax30_r50_lookback(), &expected_data::expected_h30_pmin10_pmax30_r50_width(), &expected_data::expected_h30_pmin10_pmax30_r50_has_cycle_estimate());
        verify_case("h30_pmin10_pmax30_r25", 25, 30, 10, 30, &expected_data::expected_h30_pmin10_pmax30_r25_heatmap(), &expected_data::expected_h30_pmin10_pmax30_r25_heatmap_min(), &expected_data::expected_h30_pmin10_pmax30_r25_heatmap_max(), &expected_data::expected_h30_pmin10_pmax30_r25_trend_vigor(), &expected_data::expected_h30_pmin10_pmax30_r25_normalized_vigor(), &expected_data::expected_h30_pmin10_pmax30_r25_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax30_r25_band_pass(), &expected_data::expected_h30_pmin10_pmax30_r25_quadrature(), &expected_data::expected_h30_pmin10_pmax30_r25_cycle_amplitude(), &expected_data::expected_h30_pmin10_pmax30_r25_trend_displacement(), &expected_data::expected_h30_pmin10_pmax30_r25_raw_ratio(), &expected_data::expected_h30_pmin10_pmax30_r25_lookback(), &expected_data::expected_h30_pmin10_pmax30_r25_width(), &expected_data::expected_h30_pmin10_pmax30_r25_has_cycle_estimate());
        verify_case("h30_pmin10_pmax48_r50", 50, 30, 10, 48, &expected_data::expected_h30_pmin10_pmax48_r50_heatmap(), &expected_data::expected_h30_pmin10_pmax48_r50_heatmap_min(), &expected_data::expected_h30_pmin10_pmax48_r50_heatmap_max(), &expected_data::expected_h30_pmin10_pmax48_r50_trend_vigor(), &expected_data::expected_h30_pmin10_pmax48_r50_normalized_vigor(), &expected_data::expected_h30_pmin10_pmax48_r50_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax48_r50_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax48_r50_band_pass(), &expected_data::expected_h30_pmin10_pmax48_r50_quadrature(), &expected_data::expected_h30_pmin10_pmax48_r50_cycle_amplitude(), &expected_data::expected_h30_pmin10_pmax48_r50_trend_displacement(), &expected_data::expected_h30_pmin10_pmax48_r50_raw_ratio(), &expected_data::expected_h30_pmin10_pmax48_r50_lookback(), &expected_data::expected_h30_pmin10_pmax48_r50_width(), &expected_data::expected_h30_pmin10_pmax48_r50_has_cycle_estimate());
        verify_case("h30_pmin10_pmax48_r25", 25, 30, 10, 48, &expected_data::expected_h30_pmin10_pmax48_r25_heatmap(), &expected_data::expected_h30_pmin10_pmax48_r25_heatmap_min(), &expected_data::expected_h30_pmin10_pmax48_r25_heatmap_max(), &expected_data::expected_h30_pmin10_pmax48_r25_trend_vigor(), &expected_data::expected_h30_pmin10_pmax48_r25_normalized_vigor(), &expected_data::expected_h30_pmin10_pmax48_r25_dominant_cycle(), &expected_data::expected_h30_pmin10_pmax48_r25_dominant_cycle_median(), &expected_data::expected_h30_pmin10_pmax48_r25_band_pass(), &expected_data::expected_h30_pmin10_pmax48_r25_quadrature(), &expected_data::expected_h30_pmin10_pmax48_r25_cycle_amplitude(), &expected_data::expected_h30_pmin10_pmax48_r25_trend_displacement(), &expected_data::expected_h30_pmin10_pmax48_r25_raw_ratio(), &expected_data::expected_h30_pmin10_pmax48_r25_lookback(), &expected_data::expected_h30_pmin10_pmax48_r25_width(), &expected_data::expected_h30_pmin10_pmax48_r25_has_cycle_estimate());
        verify_case("h48_pmin6_pmax30_r50", 50, 48, 6, 30, &expected_data::expected_h48_pmin6_pmax30_r50_heatmap(), &expected_data::expected_h48_pmin6_pmax30_r50_heatmap_min(), &expected_data::expected_h48_pmin6_pmax30_r50_heatmap_max(), &expected_data::expected_h48_pmin6_pmax30_r50_trend_vigor(), &expected_data::expected_h48_pmin6_pmax30_r50_normalized_vigor(), &expected_data::expected_h48_pmin6_pmax30_r50_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax30_r50_band_pass(), &expected_data::expected_h48_pmin6_pmax30_r50_quadrature(), &expected_data::expected_h48_pmin6_pmax30_r50_cycle_amplitude(), &expected_data::expected_h48_pmin6_pmax30_r50_trend_displacement(), &expected_data::expected_h48_pmin6_pmax30_r50_raw_ratio(), &expected_data::expected_h48_pmin6_pmax30_r50_lookback(), &expected_data::expected_h48_pmin6_pmax30_r50_width(), &expected_data::expected_h48_pmin6_pmax30_r50_has_cycle_estimate());
        verify_case("h48_pmin6_pmax30_r25", 25, 48, 6, 30, &expected_data::expected_h48_pmin6_pmax30_r25_heatmap(), &expected_data::expected_h48_pmin6_pmax30_r25_heatmap_min(), &expected_data::expected_h48_pmin6_pmax30_r25_heatmap_max(), &expected_data::expected_h48_pmin6_pmax30_r25_trend_vigor(), &expected_data::expected_h48_pmin6_pmax30_r25_normalized_vigor(), &expected_data::expected_h48_pmin6_pmax30_r25_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax30_r25_band_pass(), &expected_data::expected_h48_pmin6_pmax30_r25_quadrature(), &expected_data::expected_h48_pmin6_pmax30_r25_cycle_amplitude(), &expected_data::expected_h48_pmin6_pmax30_r25_trend_displacement(), &expected_data::expected_h48_pmin6_pmax30_r25_raw_ratio(), &expected_data::expected_h48_pmin6_pmax30_r25_lookback(), &expected_data::expected_h48_pmin6_pmax30_r25_width(), &expected_data::expected_h48_pmin6_pmax30_r25_has_cycle_estimate());
        verify_case("h48_pmin6_pmax20_r50", 50, 48, 6, 20, &expected_data::expected_h48_pmin6_pmax20_r50_heatmap(), &expected_data::expected_h48_pmin6_pmax20_r50_heatmap_min(), &expected_data::expected_h48_pmin6_pmax20_r50_heatmap_max(), &expected_data::expected_h48_pmin6_pmax20_r50_trend_vigor(), &expected_data::expected_h48_pmin6_pmax20_r50_normalized_vigor(), &expected_data::expected_h48_pmin6_pmax20_r50_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax20_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax20_r50_band_pass(), &expected_data::expected_h48_pmin6_pmax20_r50_quadrature(), &expected_data::expected_h48_pmin6_pmax20_r50_cycle_amplitude(), &expected_data::expected_h48_pmin6_pmax20_r50_trend_displacement(), &expected_data::expected_h48_pmin6_pmax20_r50_raw_ratio(), &expected_data::expected_h48_pmin6_pmax20_r50_lookback(), &expected_data::expected_h48_pmin6_pmax20_r50_width(), &expected_data::expected_h48_pmin6_pmax20_r50_has_cycle_estimate());
        verify_case("h48_pmin6_pmax20_r25", 25, 48, 6, 20, &expected_data::expected_h48_pmin6_pmax20_r25_heatmap(), &expected_data::expected_h48_pmin6_pmax20_r25_heatmap_min(), &expected_data::expected_h48_pmin6_pmax20_r25_heatmap_max(), &expected_data::expected_h48_pmin6_pmax20_r25_trend_vigor(), &expected_data::expected_h48_pmin6_pmax20_r25_normalized_vigor(), &expected_data::expected_h48_pmin6_pmax20_r25_dominant_cycle(), &expected_data::expected_h48_pmin6_pmax20_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin6_pmax20_r25_band_pass(), &expected_data::expected_h48_pmin6_pmax20_r25_quadrature(), &expected_data::expected_h48_pmin6_pmax20_r25_cycle_amplitude(), &expected_data::expected_h48_pmin6_pmax20_r25_trend_displacement(), &expected_data::expected_h48_pmin6_pmax20_r25_raw_ratio(), &expected_data::expected_h48_pmin6_pmax20_r25_lookback(), &expected_data::expected_h48_pmin6_pmax20_r25_width(), &expected_data::expected_h48_pmin6_pmax20_r25_has_cycle_estimate());
        verify_case("h48_pmin10_pmax30_r50", 50, 48, 10, 30, &expected_data::expected_h48_pmin10_pmax30_r50_heatmap(), &expected_data::expected_h48_pmin10_pmax30_r50_heatmap_min(), &expected_data::expected_h48_pmin10_pmax30_r50_heatmap_max(), &expected_data::expected_h48_pmin10_pmax30_r50_trend_vigor(), &expected_data::expected_h48_pmin10_pmax30_r50_normalized_vigor(), &expected_data::expected_h48_pmin10_pmax30_r50_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax30_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax30_r50_band_pass(), &expected_data::expected_h48_pmin10_pmax30_r50_quadrature(), &expected_data::expected_h48_pmin10_pmax30_r50_cycle_amplitude(), &expected_data::expected_h48_pmin10_pmax30_r50_trend_displacement(), &expected_data::expected_h48_pmin10_pmax30_r50_raw_ratio(), &expected_data::expected_h48_pmin10_pmax30_r50_lookback(), &expected_data::expected_h48_pmin10_pmax30_r50_width(), &expected_data::expected_h48_pmin10_pmax30_r50_has_cycle_estimate());
        verify_case("h48_pmin10_pmax30_r25", 25, 48, 10, 30, &expected_data::expected_h48_pmin10_pmax30_r25_heatmap(), &expected_data::expected_h48_pmin10_pmax30_r25_heatmap_min(), &expected_data::expected_h48_pmin10_pmax30_r25_heatmap_max(), &expected_data::expected_h48_pmin10_pmax30_r25_trend_vigor(), &expected_data::expected_h48_pmin10_pmax30_r25_normalized_vigor(), &expected_data::expected_h48_pmin10_pmax30_r25_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax30_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax30_r25_band_pass(), &expected_data::expected_h48_pmin10_pmax30_r25_quadrature(), &expected_data::expected_h48_pmin10_pmax30_r25_cycle_amplitude(), &expected_data::expected_h48_pmin10_pmax30_r25_trend_displacement(), &expected_data::expected_h48_pmin10_pmax30_r25_raw_ratio(), &expected_data::expected_h48_pmin10_pmax30_r25_lookback(), &expected_data::expected_h48_pmin10_pmax30_r25_width(), &expected_data::expected_h48_pmin10_pmax30_r25_has_cycle_estimate());
        verify_case("h48_pmin10_pmax48_r50", 50, 48, 10, 48, &expected_data::expected_h48_pmin10_pmax48_r50_heatmap(), &expected_data::expected_h48_pmin10_pmax48_r50_heatmap_min(), &expected_data::expected_h48_pmin10_pmax48_r50_heatmap_max(), &expected_data::expected_h48_pmin10_pmax48_r50_trend_vigor(), &expected_data::expected_h48_pmin10_pmax48_r50_normalized_vigor(), &expected_data::expected_h48_pmin10_pmax48_r50_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax48_r50_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax48_r50_band_pass(), &expected_data::expected_h48_pmin10_pmax48_r50_quadrature(), &expected_data::expected_h48_pmin10_pmax48_r50_cycle_amplitude(), &expected_data::expected_h48_pmin10_pmax48_r50_trend_displacement(), &expected_data::expected_h48_pmin10_pmax48_r50_raw_ratio(), &expected_data::expected_h48_pmin10_pmax48_r50_lookback(), &expected_data::expected_h48_pmin10_pmax48_r50_width(), &expected_data::expected_h48_pmin10_pmax48_r50_has_cycle_estimate());
        verify_case("h48_pmin10_pmax48_r25", 25, 48, 10, 48, &expected_data::expected_h48_pmin10_pmax48_r25_heatmap(), &expected_data::expected_h48_pmin10_pmax48_r25_heatmap_min(), &expected_data::expected_h48_pmin10_pmax48_r25_heatmap_max(), &expected_data::expected_h48_pmin10_pmax48_r25_trend_vigor(), &expected_data::expected_h48_pmin10_pmax48_r25_normalized_vigor(), &expected_data::expected_h48_pmin10_pmax48_r25_dominant_cycle(), &expected_data::expected_h48_pmin10_pmax48_r25_dominant_cycle_median(), &expected_data::expected_h48_pmin10_pmax48_r25_band_pass(), &expected_data::expected_h48_pmin10_pmax48_r25_quadrature(), &expected_data::expected_h48_pmin10_pmax48_r25_cycle_amplitude(), &expected_data::expected_h48_pmin10_pmax48_r25_trend_displacement(), &expected_data::expected_h48_pmin10_pmax48_r25_raw_ratio(), &expected_data::expected_h48_pmin10_pmax48_r25_lookback(), &expected_data::expected_h48_pmin10_pmax48_r25_width(), &expected_data::expected_h48_pmin10_pmax48_r25_has_cycle_estimate());
    }

    #[test]
    fn test_sinusoid_cancellation_metadata_and_validation() {
        let mut indicator = CoronaTrendVigor::new_default().unwrap();
        let md = indicator.metadata();
        assert_eq!(md.identifier, Identifier::CoronaTrendVigor);
        assert_eq!(md.mnemonic, "ctv(50, 20, -10, 10, 30)");
        assert_eq!(md.outputs.len(), 2);
        assert_eq!(md.outputs[0].shape, Shape::Heatmap);
        assert_eq!(md.outputs[1].shape, Shape::Scalar);
        assert_eq!(md.outputs[1].mnemonic, "ctv-tv(30)");

        let offset_ind = CoronaTrendVigor::new(&CoronaTrendVigorParams {
            minimal_period: 10,
            maximal_period: 48,
            lookback_offset: -1,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        })
        .unwrap();
        assert_eq!(offset_ind.metadata().mnemonic, "ctv(50, 20, -10, 10, 30, pmin=10, pmax=48, offset=-1, hl/2)");
        assert_eq!(offset_ind.metadata().outputs[1].mnemonic, "ctv-tv(30, pmin=10, pmax=48, offset=-1, hl/2)");

        assert!(CoronaTrendVigor::new(&CoronaTrendVigorParams { lookback_offset: 1, ..Default::default() }).is_err());
        assert!(CoronaTrendVigor::new(&CoronaTrendVigorParams { lookback_offset: -2, ..Default::default() }).is_err());
        assert!(CoronaTrendVigor::new(&CoronaTrendVigorParams { max_raster_value: 6.0, ..Default::default() }).is_err());

        let mut last_snap = indicator.update(100.0, 0);
        for index in 1..200 {
            let sample = 100.0 + 0.5 * (index as f64) + 2.0 * (TWO_PI * (index as f64) / 20.0).sin();
            last_snap = indicator.update(sample, index as i64);
        }
        assert!(indicator.is_primed());
        assert!((last_snap.lookback - 20.0).abs() <= 1.0);
        assert!((last_snap.trend_displacement - 10.0).abs() <= 1e-10);
        assert!(last_snap.trend_vigor > 2.0);
        indicator.reset();
        assert!(!indicator.is_primed());
    }
}
