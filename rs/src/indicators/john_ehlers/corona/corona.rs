/// Corona is the shared spectral-analysis engine and helper primitives used by
/// CoronaSpectrum, CoronaSignalToNoiseRatio, CoronaSwingPosition, and CoronaTrendVigor.

pub const FIRST_BANK_SAMPLE: usize = 13;
pub const TWO_PI: f64 = 2.0 * std::f64::consts::PI;

const FIR_WEIGHTS: [f64; 6] = [1.0, 2.0, 3.0, 3.0, 2.0, 1.0];

use crate::indicators::core::outputs::heatmap::Heatmap;

pub fn sqrt_three_over_two() -> f64 {
    3.0_f64.sqrt() / 2.0
}

pub fn validate_integer(name: &str, value: i32, minimum: i32) -> Result<(), String> {
    if value < minimum {
        return Err(format!("{name} must be an integer >= {minimum}, got {value}"));
    }
    Ok(())
}

pub fn validate_finite(name: &str, value: f64) -> Result<f64, String> {
    if !value.is_finite() {
        return Err(format!("{name} must be finite"));
    }
    Ok(value)
}

pub fn valid_price(price: f64) -> Result<bool, String> {
    if price.is_nan() {
        return Ok(false);
    }
    if !price.is_finite() {
        return Err("price must be finite or NaN".to_string());
    }
    Ok(true)
}

pub fn checked(value: f64) -> Result<f64, String> {
    if !value.is_finite() {
        return Err("calculation overflow; use finite, reasonably scaled prices".to_string());
    }
    Ok(value)
}

pub fn clip(value: f64, lower: f64, upper: f64) -> f64 {
    value.min(upper).max(lower)
}

pub fn format_number(value: f64) -> String {
    let normalized = if value == 0.0 { 0.0 } else { value };
    format!("{normalized}")
}

pub fn rolling_append(history: &mut Vec<f64>, value: f64, window_len: usize) -> (f64, f64) {
    history.push(value);
    if history.len() > window_len {
        history.remove(0);
    }
    let mut minimum = history[0];
    let mut maximum = history[0];
    for &item in &history[1..] {
        if item < minimum {
            minimum = item;
        }
        if item > maximum {
            maximum = item;
        }
    }
    (minimum, maximum)
}

/// Parameters for the Corona spectral analysis engine.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct CoronaParams {
    pub high_pass_filter_cutoff: i32,
    pub minimal_period: i32,
    pub maximal_period: i32,
    pub decibels_lower_threshold: f64,
    pub decibels_upper_threshold: f64,
}

pub type Params = CoronaParams;

impl Default for CoronaParams {
    fn default() -> Self {
        Self {
            high_pass_filter_cutoff: 30,
            minimal_period: 6,
            maximal_period: 30,
            decibels_lower_threshold: 6.0,
            decibels_upper_threshold: 20.0,
        }
    }
}

pub fn default_params() -> CoronaParams {
    CoronaParams::default()
}

/// Per-bin state of a single half-period bandpass channel.
#[derive(Debug, Clone, Copy, Default)]
pub struct Filter {
    pub in_phase1: f64,
    pub in_phase2: f64,
    pub quadrature1: f64,
    pub quadrature2: f64,
    pub real1: f64,
    pub real2: f64,
    pub imaginary1: f64,
    pub imaginary2: f64,
    pub power: f64,
    pub decibels: f64,
}

/// Owned 18-field snapshot produced by one Corona engine update.
#[derive(Debug, Clone)]
pub struct CoronaSnapshot {
    pub parameters: CoronaParams,
    pub sample_count: usize,
    pub accepted: bool,
    pub primed: bool,
    pub price: f64,
    pub time: i64,
    pub high_pass: f64,
    pub filtered: f64,
    pub bandwidth: f64,
    pub dominant_cycle: f64,
    pub dominant_cycle_median: f64,
    pub maximum_power: f64,
    pub normalization_power: f64,
    pub selected_bins: usize,
    pub has_cycle_estimate: bool,
    pub power: Vec<f64>,
    pub relative_power: Vec<f64>,
    pub decibels: Vec<f64>,
}

pub type Snapshot = CoronaSnapshot;

pub fn accept_snapshot(
    snapshot: &CoronaSnapshot,
    expected_parameters: &CoronaParams,
    expected_count: usize,
) -> Result<bool, String> {
    if &snapshot.parameters != expected_parameters {
        return Err("snapshot must use this consumer's Corona parameters".to_string());
    }
    if !snapshot.accepted {
        return Ok(false);
    }
    if snapshot.sample_count != expected_count {
        return Err("consume every accepted snapshot exactly once, in order from sample 1".to_string());
    }
    Ok(true)
}

/// Shared Corona spectral-analysis engine.
pub struct Corona {
    parameters: CoronaParams,
    minimal_period: i32,
    maximal_period: i32,
    minimal_period_times_two: i32,
    maximal_period_times_two: i32,
    filter_bank_length: usize,
    pole: f64,
    periods: Vec<f64>,
    betas: Vec<f64>,
    bank: Vec<Filter>,
    sample_count: usize,
    price_previous: f64,
    hp_history: [f64; 6],
    filtered_previous: f64,
    dc: f64,
    dc_history: [f64; 5],
    last_snapshot: Option<CoronaSnapshot>,
}

impl Corona {
    pub fn new(params: &CoronaParams) -> Result<Self, String> {
        validate_integer("high_pass_filter_cutoff", params.high_pass_filter_cutoff, 5)?;
        validate_integer("minimal_period", params.minimal_period, 6)?;
        validate_integer("maximal_period", params.maximal_period, params.minimal_period + 1)?;
        let threshold = validate_finite("decibels_lower_threshold", params.decibels_lower_threshold)?;
        let ceiling = validate_finite("decibels_upper_threshold", params.decibels_upper_threshold)?;
        if !(0.0 <= threshold && threshold < ceiling && ceiling > 0.0) {
            return Err("require 0 <= decibels_lower_threshold < decibels_upper_threshold".to_string());
        }

        let angle = TWO_PI / (params.high_pass_filter_cutoff as f64);
        let pole = (1.0 - angle.sin()) / angle.cos();
        if !pole.is_finite() || !(0.0 < pole && pole < 1.0) {
            return Err("highpass coefficient must have a stable positive pole".to_string());
        }

        let min_p2 = 2 * params.minimal_period;
        let max_p2 = 2 * params.maximal_period;
        let bank_len = (max_p2 - min_p2 + 1) as usize;
        let mut periods = Vec::with_capacity(bank_len);
        let mut betas = Vec::with_capacity(bank_len);
        for step_index in min_p2..=max_p2 {
            let period = (step_index as f64) / 2.0;
            periods.push(period);
            betas.push((TWO_PI / period).cos());
        }

        let mut engine = Self {
            parameters: *params,
            minimal_period: params.minimal_period,
            maximal_period: params.maximal_period,
            minimal_period_times_two: min_p2,
            maximal_period_times_two: max_p2,
            filter_bank_length: bank_len,
            pole,
            periods,
            betas,
            bank: vec![Filter::default(); bank_len],
            sample_count: 0,
            price_previous: 0.0,
            hp_history: [0.0; 6],
            filtered_previous: 0.0,
            dc: 0.0,
            dc_history: [0.0; 5],
            last_snapshot: None,
        };
        engine.reset();
        Ok(engine)
    }

    pub fn new_default() -> Result<Self, String> {
        Self::new(&CoronaParams::default())
    }

    pub fn reset(&mut self) {
        self.sample_count = 0;
        self.price_previous = 0.0;
        self.hp_history = [0.0; 6];
        self.filtered_previous = 0.0;
        self.bank.fill(Filter::default());
        self.dc = 0.0;
        self.dc_history = [0.0; 5];
        self.last_snapshot = None;
    }

    pub fn parameters(&self) -> CoronaParams {
        self.parameters
    }

    pub fn minimal_period(&self) -> i32 {
        self.minimal_period
    }

    pub fn maximal_period(&self) -> i32 {
        self.maximal_period
    }

    pub fn minimal_period_times_two(&self) -> i32 {
        self.minimal_period_times_two
    }

    pub fn maximal_period_times_two(&self) -> i32 {
        self.maximal_period_times_two
    }

    pub fn filter_bank_length(&self) -> usize {
        self.filter_bank_length
    }

    pub fn periods(&self) -> &[f64] {
        &self.periods
    }

    pub fn filter_bank(&self) -> &[Filter] {
        &self.bank
    }

    pub fn is_primed(&self) -> bool {
        self.sample_count >= FIRST_BANK_SAMPLE
    }

    pub fn sample_count(&self) -> usize {
        self.sample_count
    }

    pub fn dominant_cycle(&self) -> f64 {
        self.dc
    }

    pub fn dominant_cycle_median(&self) -> f64 {
        if let Some(ref snap) = self.last_snapshot {
            snap.dominant_cycle_median
        } else {
            self.minimal_period as f64
        }
    }

    pub fn maximum_power(&self) -> f64 {
        self.last_snapshot.as_ref().map_or(0.0, |snap| snap.maximum_power)
    }

    pub fn try_update(&mut self, price: f64, time: i64) -> Result<CoronaSnapshot, String> {
        if !valid_price(price)? {
            if let Some(ref previous_snap) = self.last_snapshot {
                let mut skipped = previous_snap.clone();
                skipped.accepted = false;
                skipped.price = price;
                skipped.time = time;
                return Ok(skipped);
            }
            let zeros = vec![0.0; self.filter_bank_length];
            return Ok(CoronaSnapshot {
                parameters: self.parameters,
                sample_count: 0,
                accepted: false,
                primed: false,
                price,
                time,
                high_pass: 0.0,
                filtered: 0.0,
                bandwidth: 0.5,
                dominant_cycle: 0.0,
                dominant_cycle_median: self.minimal_period as f64,
                maximum_power: 0.0,
                normalization_power: 0.01,
                selected_bins: 0,
                has_cycle_estimate: false,
                power: zeros.clone(),
                relative_power: zeros.clone(),
                decibels: zeros,
            });
        }

        let count = self.sample_count + 1;
        let previous_price = if count == 1 { price } else { self.price_previous };
        let hp = checked(
            0.5 * (1.0 + self.pole) * (price - previous_price) + self.pole * self.hp_history[5],
        )?;
        let mut hp_history = [0.0; 6];
        hp_history[..5].copy_from_slice(&self.hp_history[1..6]);
        hp_history[5] = hp;

        let mut fir_sum = 0.0;
        for index in 0..6 {
            fir_sum += hp_history[index] * FIR_WEIGHTS[index];
        }
        let mut filtered = checked(fir_sum / 12.0)?;
        if count < 7 {
            filtered = checked(price - previous_price)?;
        }
        if count == 1 {
            filtered = 0.0;
        }
        let bandwidth = (0.5 - 0.015 * (count as f64)).max(0.1);

        let mut next_bank = Vec::with_capacity(self.filter_bank_length);
        for index in 0..self.filter_bank_length {
            let old = self.bank[index];
            if count < FIRST_BANK_SAMPLE {
                next_bank.push(old);
                continue;
            }
            let period = self.periods[index];
            let beta = self.betas[index];
            let gamma = 1.0 / (2.0 * TWO_PI * bandwidth / period).cos();
            let alpha = gamma - (gamma * gamma - 1.0).sqrt();
            let in_phase = filtered;
            let quadrature = checked((period / TWO_PI) * (filtered - self.filtered_previous))?;
            let real_part = checked(
                0.5 * (1.0 - alpha) * (in_phase - old.in_phase2)
                    + beta * (1.0 + alpha) * old.real1
                    - alpha * old.real2,
            )?;
            let imaginary_part = checked(
                0.5 * (1.0 - alpha) * (quadrature - old.quadrature2)
                    + beta * (1.0 + alpha) * old.imaginary1
                    - alpha * old.imaginary2,
            )?;
            let power = checked(real_part * real_part + imaginary_part * imaginary_part)?;
            next_bank.push(Filter {
                in_phase1: in_phase,
                in_phase2: old.in_phase1,
                quadrature1: quadrature,
                quadrature2: old.quadrature1,
                real1: real_part,
                real2: old.real1,
                imaginary1: imaginary_part,
                imaginary2: old.imaginary1,
                power,
                decibels: old.decibels,
            });
        }

        let mut maximum_power = next_bank[0].power;
        for channel in &next_bank[1..] {
            if channel.power > maximum_power {
                maximum_power = channel.power;
            }
        }
        let first_power = if next_bank[0].power == 0.0 { 0.01 } else { next_bank[0].power };
        let normalization_power = first_power.max(maximum_power);

        let upper_threshold = self.parameters.decibels_upper_threshold;
        let lower_threshold = self.parameters.decibels_lower_threshold;
        let mut numerator = 0.0;
        let mut denominator = 0.0;
        let mut selected_bins = 0usize;
        let mut powers = Vec::with_capacity(self.filter_bank_length);
        let mut relative_powers = Vec::with_capacity(self.filter_bank_length);
        let mut decibels_vec = Vec::with_capacity(self.filter_bank_length);

        for index in 0..self.filter_bank_length {
            let channel = &mut next_bank[index];
            let relative_power = channel.power / normalization_power;
            let mut raw_decibels = channel.decibels;
            if relative_power > 0.0 {
                raw_decibels = 10.0 * ((1.0 - 0.99 * relative_power) / 0.01).log10();
            }
            let smoothed = 0.33 * raw_decibels + 0.67 * channel.decibels;
            channel.decibels = smoothed.min(upper_threshold);
            if channel.decibels <= lower_threshold {
                let weight = upper_threshold - channel.decibels;
                numerator += self.periods[index] * weight;
                denominator += weight;
                selected_bins += 1;
            }
            powers.push(channel.power);
            relative_powers.push(relative_power);
            decibels_vec.push(channel.decibels);
        }

        let dc = if denominator > 0.0 {
            numerator / denominator
        } else {
            self.dc
        };
        let mut dc_history = [0.0; 5];
        dc_history[..4].copy_from_slice(&self.dc_history[1..5]);
        dc_history[4] = dc;
        let mut sorted_dc = dc_history;
        sorted_dc.sort_by(|left, right| left.total_cmp(right));
        let dc_median = (self.minimal_period as f64).max(sorted_dc[2]);

        let snapshot = CoronaSnapshot {
            parameters: self.parameters,
            sample_count: count,
            accepted: true,
            primed: count >= FIRST_BANK_SAMPLE,
            price,
            time,
            high_pass: hp,
            filtered,
            bandwidth,
            dominant_cycle: dc,
            dominant_cycle_median: dc_median,
            maximum_power,
            normalization_power,
            selected_bins,
            has_cycle_estimate: maximum_power > 0.0 && selected_bins > 0,
            power: powers,
            relative_power: relative_powers,
            decibels: decibels_vec,
        };

        self.sample_count = count;
        self.price_previous = price;
        self.hp_history = hp_history;
        self.filtered_previous = filtered;
        self.bank = next_bank;
        self.dc = dc;
        self.dc_history = dc_history;
        self.last_snapshot = Some(snapshot.clone());
        Ok(snapshot)
    }

    pub fn update(&mut self, price: f64, time: i64) -> CoronaSnapshot {
        self.try_update(price, time)
            .unwrap_or_else(|err| panic!("Corona::update failed: {err}"))
    }
}

#[derive(Debug, Clone, Copy)]
pub struct AdaptiveCycle {
    pub band_pass: f64,
    pub quadrature: f64,
    pub amplitude: f64,
    pub lead60: f64,
}

#[derive(Debug, Clone, Default)]
pub struct AdaptiveBandPass {
    count: usize,
    prices: [f64; 2],
    bp1: f64,
    bp2: f64,
}

impl AdaptiveBandPass {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn reset(&mut self) {
        self.count = 0;
        self.prices = [0.0; 2];
        self.bp1 = 0.0;
        self.bp2 = 0.0;
    }

    pub fn update(&mut self, price: f64, period: f64) -> Result<AdaptiveCycle, String> {
        self.count += 1;
        if self.count <= 2 {
            self.prices[self.count - 1] = price;
            return Ok(AdaptiveCycle {
                band_pass: 0.0,
                quadrature: 0.0,
                amplitude: 0.0,
                lead60: 0.0,
            });
        }
        let omega = TWO_PI / period;
        let beta = omega.cos();
        let gamma = 1.0 / (2.0 * omega * 0.1).cos();
        let alpha = gamma - (gamma * gamma - 1.0).sqrt();
        let bp = checked(
            0.5 * (1.0 - alpha) * (price - self.prices[0])
                + beta * (1.0 + alpha) * self.bp1
                - alpha * self.bp2,
        )?;
        let quadrature = checked((bp - self.bp1) / omega)?;
        let amplitude = checked((bp * bp + quadrature * quadrature).sqrt())?;
        let lead60 = checked(0.5 * bp + sqrt_three_over_two() * quadrature)?;
        self.prices = [self.prices[1], price];
        self.bp2 = self.bp1;
        self.bp1 = bp;
        Ok(AdaptiveCycle {
            band_pass: bp,
            quadrature,
            amplitude,
            lead60,
        })
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RasterKind {
    Swing,
    Trend,
    Snr,
}

#[derive(Debug, Clone)]
pub struct Raster {
    length: usize,
    lower: f64,
    upper: f64,
    ceiling: f64,
    step: f64,
    first: f64,
    resolution: f64,
    values: Vec<f64>,
}

impl Raster {
    pub fn new(length_raw: i32, lower: f64, upper: f64, ceiling: f64) -> Result<Self, String> {
        validate_integer("raster_length", length_raw, 2)?;
        let lower_val = validate_finite("min_parameter_value", lower)?;
        let upper_val = validate_finite("max_parameter_value", upper)?;
        let ceiling_val = validate_finite("max_raster_value", ceiling)?;
        if upper_val <= lower_val || ceiling_val <= 0.0 {
            return Err("require min_parameter_value < max_parameter_value and max_raster_value > 0".to_string());
        }
        let length = length_raw as usize;
        let step = (upper_val - lower_val) / (length as f64);
        let first = lower_val + step;
        let resolution = 1.0 / step;
        Ok(Self {
            length,
            lower: lower_val,
            upper: upper_val,
            ceiling: ceiling_val,
            step,
            first,
            resolution,
            values: vec![0.0; length],
        })
    }

    pub fn reset(&mut self) {
        self.values.fill(0.0);
    }

    pub fn length(&self) -> usize {
        self.length
    }

    pub fn lower(&self) -> f64 {
        self.lower
    }

    pub fn upper(&self) -> f64 {
        self.upper
    }

    pub fn ceiling(&self) -> f64 {
        self.ceiling
    }

    pub fn step(&self) -> f64 {
        self.step
    }

    pub fn first(&self) -> f64 {
        self.first
    }

    pub fn resolution(&self) -> f64 {
        self.resolution
    }

    pub fn empty(&self, time: i64) -> Heatmap {
        Heatmap::empty(time, self.first, self.upper, self.resolution)
    }

    pub fn column(&self, time: i64) -> Heatmap {
        let mut minimum = self.values[0];
        let mut maximum = self.values[0];
        for &item in &self.values[1..] {
            if item < minimum {
                minimum = item;
            }
            if item > maximum {
                maximum = item;
            }
        }
        Heatmap::new(
            time,
            self.first,
            self.upper,
            self.resolution,
            minimum,
            maximum,
            self.values.clone(),
        )
    }

    pub fn update(&mut self, position: f64, width: f64, kind: RasterKind, suppress: bool) {
        let length_f = self.length as f64;
        let marker = (length_f * position + 0.5).floor() as isize;
        let (exponent, scale, history_scale) = match kind {
            RasterKind::Swing => (0.95, 0.5, 0.5),
            RasterKind::Trend => (0.85, 0.8, 0.2),
            RasterKind::Snr => (0.8, 0.5, 1.0),
        };
        for index in 0..self.length {
            let old = self.values[index];
            let cell_index = (index as isize) + 1;
            let cell_f = cell_index as f64;
            let value = if suppress {
                self.ceiling
            } else if cell_index == marker {
                0.5 * old
            } else if width <= 0.0 {
                self.ceiling
            } else {
                let distance = (self.ceiling * position - self.ceiling * cell_f / length_f).abs() / width;
                if kind == RasterKind::Snr && cell_index > marker && distance <= 1.0 {
                    self.ceiling
                } else {
                    scale * (distance.powf(exponent) + history_scale * old)
                }
            };
            self.values[index] = clip(value, 0.0, self.ceiling);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    use crate::indicators::john_ehlers::corona_signal_to_noise_ratio::CoronaSignalToNoiseRatio;
    use crate::indicators::john_ehlers::corona_spectrum::CoronaSpectrum;
    use crate::indicators::john_ehlers::corona_swing_position::CoronaSwingPosition;
    use crate::indicators::john_ehlers::corona_trend_vigor::CoronaTrendVigor;

    #[test]
    fn test_startup_and_empty_selection_holds() {
        let mut engine = Corona::new_default().unwrap();
        assert_eq!(engine.filter_bank_length(), 49);
        assert_eq!(engine.minimal_period_times_two(), 12);
        assert_eq!(engine.maximal_period_times_two(), 60);

        let prices = testdata::test_input();
        let mut previous_dc = 0.0;
        let mut holds = 0usize;
        let mut final_dc = 0.0;

        for (index, &price) in prices.iter().enumerate() {
            let snap = engine.update(price, index as i64);
            if index < 12 {
                assert!(!snap.primed, "bar {index} should not be primed");
                assert_eq!(snap.maximum_power, 0.0);
                assert!(!snap.has_cycle_estimate);
            } else if index == 12 {
                assert!(snap.primed, "bar 12 (sample 13) should be primed");
            }
            if index > 0 && snap.selected_bins == 0 {
                holds += 1;
                assert_eq!(snap.dominant_cycle, previous_dc);
                assert!(!snap.has_cycle_estimate);
            }
            previous_dc = snap.dominant_cycle;
            final_dc = snap.dominant_cycle;
        }
        assert!(holds > 0);
        assert!((final_dc - 11.5).abs() <= 5e-12);
    }

    #[test]
    fn test_shared_and_standalone_consumers_match() {
        let prices = testdata::test_input();
        let highs = testdata::test_high();
        let lows = testdata::test_low();

        let mut shared = Corona::new_default().unwrap();
        let mut spectrum_shared = CoronaSpectrum::new_default().unwrap();
        let mut swing_shared = CoronaSwingPosition::new_default().unwrap();
        let mut trend_shared = CoronaTrendVigor::new_default().unwrap();
        let mut snr_shared = CoronaSignalToNoiseRatio::new_default().unwrap();

        let mut spectrum_solo = CoronaSpectrum::new_default().unwrap();
        let mut swing_solo = CoronaSwingPosition::new_default().unwrap();
        let mut trend_solo = CoronaTrendVigor::new_default().unwrap();
        let mut snr_solo = CoronaSignalToNoiseRatio::new_default().unwrap();

        for index in 0..prices.len() {
            let t = index as i64;
            let snap = shared.update(prices[index], t);
            let sp_a = spectrum_shared.update_from_snapshot(&snap);
            let sp_b = spectrum_solo.update(prices[index], t);
            assert_eq!(sp_a.dominant_cycle.to_bits(), sp_b.dominant_cycle.to_bits());
            assert_eq!(sp_a.heatmap.values, sp_b.heatmap.values);

            let sw_a = swing_shared.update_from_snapshot(&snap);
            let sw_b = swing_solo.update(prices[index], t);
            assert_eq!(sw_a.swing_position.to_bits(), sw_b.swing_position.to_bits());
            assert_eq!(sw_a.heatmap.values, sw_b.heatmap.values);

            let tv_a = trend_shared.update_from_snapshot(&snap);
            let tv_b = trend_solo.update(prices[index], t);
            assert_eq!(tv_a.trend_vigor.to_bits(), tv_b.trend_vigor.to_bits());
            assert_eq!(tv_a.heatmap.values, tv_b.heatmap.values);

            let sn_a = snr_shared.update_from_snapshot(&snap, lows[index], highs[index]);
            let sn_b = snr_solo.update(prices[index], lows[index], highs[index], t);
            assert_eq!(sn_a.signal_to_noise_ratio.to_bits(), sn_b.signal_to_noise_ratio.to_bits());
            assert_eq!(sn_a.heatmap.values, sn_b.heatmap.values);
        }

        // Duplicate consumption of the same snapshot is rejected.
        let last_snap = shared.last_snapshot.clone().unwrap();
        assert!(spectrum_shared.try_update_from_snapshot(&last_snap).is_err());
    }

    #[test]
    fn test_nan_skips_and_snapshot_owns_arrays() {
        let mut engine = Corona::new_default().unwrap();
        let pre = engine.update(f64::NAN, 0);
        assert!(!pre.accepted);
        assert_eq!(engine.sample_count(), 0);

        let prices = testdata::test_input();
        let mut last = pre;
        for (index, &price) in prices[..20].iter().enumerate() {
            last = engine.update(price, index as i64);
        }
        assert!(last.primed);
        let saved_dc = last.dominant_cycle;
        let mut mutated = last.clone();
        mutated.decibels[0] = -999.0;
        assert_ne!(engine.filter_bank()[0].decibels, -999.0);

        let skipped = engine.update(f64::NAN, 99);
        assert!(!skipped.accepted);
        assert_eq!(skipped.sample_count, 20);
        assert_eq!(skipped.dominant_cycle, saved_dc);
    }

    #[test]
    fn test_validation_zero_threshold_overflow_and_reset() {
        assert!(Corona::new(&CoronaParams { minimal_period: 5, ..Default::default() }).is_err());
        assert!(Corona::new(&CoronaParams { minimal_period: 0, ..Default::default() }).is_err());
        assert!(Corona::new(&CoronaParams { minimal_period: 6, maximal_period: 6, ..Default::default() }).is_err());
        assert!(Corona::new(&CoronaParams { high_pass_filter_cutoff: 4, ..Default::default() }).is_err());
        assert!(Corona::new(&CoronaParams { high_pass_filter_cutoff: 0, ..Default::default() }).is_err());
        assert!(Corona::new(&CoronaParams { decibels_lower_threshold: -0.1, ..Default::default() }).is_err());
        assert!(Corona::new(&CoronaParams {
            decibels_lower_threshold: 20.0,
            decibels_upper_threshold: 20.0,
            ..Default::default()
        }).is_err());

        let zero_lower = Corona::new(&CoronaParams {
            decibels_lower_threshold: 0.0,
            ..Default::default()
        }).unwrap();
        assert_eq!(zero_lower.parameters().decibels_lower_threshold, 0.0);

        let mut engine = Corona::new_default().unwrap();
        for (index, &price) in testdata::test_input()[..30].iter().enumerate() {
            engine.update(price, index as i64);
        }
        assert!(engine.try_update(f64::INFINITY, 30).is_err());
        assert!(engine.try_update(1e308, 30).is_err());
        assert_eq!(engine.sample_count(), 30);

        engine.reset();
        assert_eq!(engine.sample_count(), 0);
        assert!(!engine.is_primed());
    }

    #[test]
    fn test_adaptive_bandpass_and_raster_geometry() {
        let mut bandpass = AdaptiveBandPass::new();
        let mut previous_bp = 0.0;
        for index in 0..100 {
            let sample = 100.0 + (TWO_PI * (index as f64) / 20.0).sin();
            let cycle = bandpass.update(sample, 20.0).unwrap();
            if index >= 2 {
                let expected_q = (cycle.band_pass - previous_bp) / (TWO_PI / 20.0);
                let expected_lead = 0.5 * cycle.band_pass + sqrt_three_over_two() * cycle.quadrature;
                assert!((cycle.quadrature - expected_q).abs() <= 1e-14);
                assert!((cycle.lead60 - expected_lead).abs() <= 1e-14);
            }
            previous_bp = cycle.band_pass;
        }

        let mut raster = Raster::new(50, -5.0, 5.0, 20.0).unwrap();
        raster.update(1.0, 0.1, RasterKind::Swing, false);
        let col = raster.column(0);
        assert!((col.parameter_first - (-4.8)).abs() <= 1e-14);
        assert!((col.parameter_last - 5.0).abs() <= 1e-14);
        assert!((col.parameter_resolution - 5.0).abs() <= 1e-14);
        assert_eq!(col.values[49], 0.0);
        assert!(col.values[0] > 0.0);

        raster.values.fill(8.0);
        raster.update(0.01, 0.1, RasterKind::Swing, false);
        assert_eq!(raster.values[0], 4.0);
    }
}
