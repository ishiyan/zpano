use std::any::Any;

use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent};
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
use crate::indicators::john_ehlers::corona::corona::{Corona, CoronaParams};

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

/// Bandpass filter delta factor in γ = 1/cos(ω · 2 · 0.1).
const BP_DELTA: f64 = 0.1;

/// Ratio EMA coefficients (0.33 * current + 0.67 * previous).
const RATIO_NEW_COEF: f64 = 0.33;
const RATIO_PREVIOUS_COEF: f64 = 0.67;

/// Vigor band thresholds.
const VIGOR_MID_LOW: f64 = 0.3;
const VIGOR_MID_HIGH: f64 = 0.7;
const VIGOR_MID: f64 = 0.5;
const WIDTH_EDGE: f64 = 0.01;

/// Raster update blend factors.
const RASTER_BLEND_SCALE: f64 = 0.8;
const RASTER_BLEND_PREVIOUS: f64 = 0.2;
const RASTER_BLEND_HALF: f64 = 0.5;
const RASTER_BLEND_EXPONENT: f64 = 0.85;

/// Ratio clamp bounds.
const RATIO_LIMIT: f64 = 10.0;

/// vigor = vigorScale*(ratio+ratioLimit); maps ratio ∈ [-10,10] to vigor ∈ [0,1].
const VIGOR_SCALE: f64 = 0.05;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Corona Trend Vigor indicator.
pub struct CoronaTrendVigorParams {
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

impl Default for CoronaTrendVigorParams {
    fn default() -> Self {
        Self {
            raster_length: 50,
            max_raster_value: 20.0,
            min_parameter_value: 0.0,
            max_parameter_value: 0.0,
            high_pass_filter_cutoff: 30,
            minimal_period: 6,
            maximal_period: 30,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output enum (1-based, matching Go iota+1)
// ---------------------------------------------------------------------------

/// Output indices for the Corona Trend Vigor indicator.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CoronaTrendVigorOutput {
    /// The Corona trend vigor heatmap column.
    Value = 1,
    /// The scalar trend vigor in [MinParameterValue, MaxParameterValue].
    TrendVigor = 2,
}

// ---------------------------------------------------------------------------
// Indicator struct
// ---------------------------------------------------------------------------

/// Ehlers' Corona Trend Vigor heatmap indicator.
pub struct CoronaTrendVigor {
    mnemonic: String,
    description: String,
    mnemonic_tv: String,
    description_tv: String,
    c: Corona,
    raster_length: usize,
    raster_step: f64,
    max_raster_value: f64,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    raster: Vec<f64>,
    sample_buffer: Vec<f64>,
    sample_count: usize,
    sample_previous: f64,
    sample_previous2: f64,
    band_pass_previous: f64,
    band_pass_previous2: f64,
    ratio_previous: f64,
    trend_vigor: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CoronaTrendVigor {
    /// Creates a new Corona Trend Vigor indicator with the given parameters.
    pub fn new(p: &CoronaTrendVigorParams) -> Result<Self, String> {
        let invalid = "invalid corona trend vigor parameters";

        let mut raster_length = p.raster_length;
        if raster_length == 0 {
            raster_length = 50;
        }

        let mut max_raster_value = p.max_raster_value;
        if max_raster_value == 0.0 {
            max_raster_value = 20.0;
        }

        let mut min_parameter_value = p.min_parameter_value;
        let mut max_parameter_value = p.max_parameter_value;
        if min_parameter_value == 0.0 && max_parameter_value == 0.0 {
            min_parameter_value = -10.0;
            max_parameter_value = 10.0;
        }

        let mut high_pass_filter_cutoff = p.high_pass_filter_cutoff;
        if high_pass_filter_cutoff == 0 {
            high_pass_filter_cutoff = 30;
        }

        let mut minimal_period = p.minimal_period;
        if minimal_period == 0 {
            minimal_period = 6;
        }

        let mut maximal_period = p.maximal_period;
        if maximal_period == 0 {
            maximal_period = 30;
        }

        if raster_length < 2 {
            return Err(format!("{}: RasterLength should be >= 2", invalid));
        }
        if max_raster_value <= 0.0 {
            return Err(format!("{}: MaxRasterValue should be > 0", invalid));
        }
        if max_parameter_value <= min_parameter_value {
            return Err(format!(
                "{}: MaxParameterValue should be > MinParameterValue",
                invalid
            ));
        }
        if high_pass_filter_cutoff < 2 {
            return Err(format!(
                "{}: HighPassFilterCutoff should be >= 2",
                invalid
            ));
        }
        if minimal_period < 2 {
            return Err(format!("{}: MinimalPeriod should be >= 2", invalid));
        }
        if maximal_period <= minimal_period {
            return Err(format!(
                "{}: MaximalPeriod should be > MinimalPeriod",
                invalid
            ));
        }

        let bc = p.bar_component.unwrap_or(BarComponent::Median);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let c = Corona::new(&CoronaParams {
            high_pass_filter_cutoff,
            minimal_period,
            maximal_period,
            ..CoronaParams::default()
        })
        .map_err(|e| format!("{}: {}", invalid, e))?;

        let buf_size = c.maximal_period_times_two() as usize;

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let component_mnemonic = component_triple_mnemonic(bc, qc, tc);

        let mnemonic = format!(
            "ctv({}, {}, {}, {}, {}{})",
            raster_length, max_raster_value, min_parameter_value,
            max_parameter_value, high_pass_filter_cutoff, component_mnemonic,
        );
        let mnemonic_tv = format!(
            "ctv-tv({}{})",
            high_pass_filter_cutoff, component_mnemonic,
        );
        let description = format!("Corona trend vigor {}", mnemonic);
        let description_tv = format!("Corona trend vigor scalar {}", mnemonic_tv);

        let rl = raster_length as usize;
        let parameter_resolution =
            (rl as f64 - 1.0) / (max_parameter_value - min_parameter_value);

        Ok(Self {
            mnemonic,
            description,
            mnemonic_tv,
            description_tv,
            c,
            raster_length: rl,
            raster_step: max_raster_value / rl as f64,
            max_raster_value,
            min_parameter_value,
            max_parameter_value,
            parameter_resolution,
            raster: vec![0.0; rl],
            sample_buffer: vec![0.0; buf_size],
            sample_count: 0,
            sample_previous: 0.0,
            sample_previous2: 0.0,
            band_pass_previous: 0.0,
            band_pass_previous2: 0.0,
            ratio_previous: 0.0,
            trend_vigor: f64::NAN,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Feeds the next sample and returns (heatmap, trend_vigor).
    pub fn update(&mut self, sample: f64, time: i64) -> (Heatmap, f64) {
        if sample.is_nan() {
            return (
                Heatmap::empty(
                    time,
                    self.min_parameter_value,
                    self.max_parameter_value,
                    self.parameter_resolution,
                ),
                f64::NAN,
            );
        }

        let primed = self.c.update(sample);
        self.sample_count += 1;

        let buf_last = self.sample_buffer.len() - 1;

        if self.sample_count == 1 {
            self.sample_previous = sample;
            self.sample_buffer[buf_last] = sample;

            return (
                Heatmap::empty(
                    time,
                    self.min_parameter_value,
                    self.max_parameter_value,
                    self.parameter_resolution,
                ),
                f64::NAN,
            );
        }

        // Bandpass InPhase filter at the dominant cycle median period.
        let omega = 2.0 * std::f64::consts::PI / self.c.dominant_cycle_median();
        let beta2 = omega.cos();
        let gamma2 = 1.0 / (omega * 2.0 * BP_DELTA).cos();
        let alpha2 = gamma2 - (gamma2 * gamma2 - 1.0).sqrt();
        let band_pass = 0.5 * (1.0 - alpha2) * (sample - self.sample_previous2)
            + beta2 * (1.0 + alpha2) * self.band_pass_previous
            - alpha2 * self.band_pass_previous2;

        // Quadrature = derivative / omega.
        let quadrature2 = (band_pass - self.band_pass_previous) / omega;

        self.band_pass_previous2 = self.band_pass_previous;
        self.band_pass_previous = band_pass;
        self.sample_previous2 = self.sample_previous;
        self.sample_previous = sample;

        // Left-shift sampleBuffer and append the new sample.
        for i in 0..buf_last {
            self.sample_buffer[i] = self.sample_buffer[i + 1];
        }
        self.sample_buffer[buf_last] = sample;

        // Cycle amplitude.
        let amplitude2 = (band_pass * band_pass + quadrature2 * quadrature2).sqrt();

        // Trend amplitude taken over the cycle period.
        let mut cycle_period = (self.c.dominant_cycle_median() - 1.0) as usize;
        if cycle_period > self.sample_buffer.len() {
            cycle_period = self.sample_buffer.len();
        }
        if cycle_period < 1 {
            cycle_period = 1;
        }

        let mut lookback = cycle_period;
        if self.sample_count < lookback {
            lookback = self.sample_count;
        }

        let trend = sample - self.sample_buffer[self.sample_buffer.len() - lookback];

        let mut ratio = 0.0;
        if trend.abs() > 0.0 && amplitude2 > 0.0 {
            ratio = RATIO_NEW_COEF * trend / amplitude2
                + RATIO_PREVIOUS_COEF * self.ratio_previous;
        }

        if ratio > RATIO_LIMIT {
            ratio = RATIO_LIMIT;
        } else if ratio < -RATIO_LIMIT {
            ratio = -RATIO_LIMIT;
        }

        self.ratio_previous = ratio;

        // ratio ∈ [-10, 10] ⇒ vigor ∈ [0, 1].
        let vigor = VIGOR_SCALE * (ratio + RATIO_LIMIT);

        let width = if vigor >= VIGOR_MID_LOW && vigor < VIGOR_MID {
            vigor - (VIGOR_MID_LOW - WIDTH_EDGE)
        } else if vigor >= VIGOR_MID && vigor <= VIGOR_MID_HIGH {
            (VIGOR_MID_HIGH + WIDTH_EDGE) - vigor
        } else {
            WIDTH_EDGE
        };

        self.trend_vigor = (self.max_parameter_value - self.min_parameter_value) * vigor
            + self.min_parameter_value;

        let vigor_scaled_to_raster_length =
            (self.raster_length as f64 * vigor).round() as i32;
        let vigor_scaled_to_max_raster_value = vigor * self.max_raster_value;

        for i in 0..self.raster_length {
            let mut value = self.raster[i];

            if i as i32 == vigor_scaled_to_raster_length {
                value *= RASTER_BLEND_HALF;
            } else {
                let mut argument =
                    vigor_scaled_to_max_raster_value - self.raster_step * i as f64;
                if (i as i32) > vigor_scaled_to_raster_length {
                    argument = -argument;
                }

                if width > 0.0 {
                    value = RASTER_BLEND_SCALE
                        * ((argument / width).powf(RASTER_BLEND_EXPONENT)
                            + RASTER_BLEND_PREVIOUS * value);
                }
            }

            // Clamp and saturate.
            if value < 0.0 {
                value = 0.0;
            } else if value > self.max_raster_value
                || vigor < VIGOR_MID_LOW
                || vigor > VIGOR_MID_HIGH
            {
                value = self.max_raster_value;
            }

            if value.is_nan() {
                value = 0.0;
            }

            self.raster[i] = value;
        }

        if !primed {
            return (
                Heatmap::empty(
                    time,
                    self.min_parameter_value,
                    self.max_parameter_value,
                    self.parameter_resolution,
                ),
                f64::NAN,
            );
        }

        let mut value_min = f64::INFINITY;
        let mut value_max = f64::NEG_INFINITY;
        let mut values = vec![0.0; self.raster_length];

        for i in 0..self.raster_length {
            let v = self.raster[i];
            values[i] = v;
            if v < value_min {
                value_min = v;
            }
            if v > value_max {
                value_max = v;
            }
        }

        let heatmap = Heatmap::new(
            time,
            self.min_parameter_value,
            self.max_parameter_value,
            self.parameter_resolution,
            value_min,
            value_max,
            values,
        );

        (heatmap, self.trend_vigor)
    }
}

// ---------------------------------------------------------------------------
// Indicator trait
// ---------------------------------------------------------------------------

impl Indicator for CoronaTrendVigor {
    fn is_primed(&self) -> bool {
        self.c.is_primed()
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CoronaTrendVigor,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: self.mnemonic.clone(),
                    description: self.description.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_tv.clone(),
                    description: self.description_tv.clone(),
                },
            ],
        )
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        let time = bar.time;
        let (heatmap, tv) = self.update(sample, time);
        vec![
            Box::new(heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, tv)) as Box<dyn Any>,
        ]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.quote_func)(quote);
        let time = quote.time;
        let (heatmap, tv) = self.update(sample, time);
        vec![
            Box::new(heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, tv)) as Box<dyn Any>,
        ]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.trade_func)(trade);
        let time = trade.time;
        let (heatmap, tv) = self.update(sample, time);
        vec![
            Box::new(heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, tv)) as Box<dyn Any>,
        ]
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let time = scalar.time;
        let (heatmap, tv) = self.update(scalar.value, time);
        vec![
            Box::new(heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, tv)) as Box<dyn Any>,
        ]
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    const TOLERANCE: f64 = 1e-4;

    #[test]
    fn test_update_snapshots() {
        let input = testdata::test_input();
        let mut x =
            CoronaTrendVigor::new(&CoronaTrendVigorParams::default()).unwrap();

        struct Snap {
            i: usize,
            tv: f64,
            vmn: f64,
            vmx: f64,
        }

        let snapshots = vec![
            Snap { i: 11, tv: 5.6512200755, vmn: 20.0, vmx: 20.0 },
            Snap { i: 12, tv: 6.8379492897, vmn: 20.0, vmx: 20.0 },
            Snap { i: 50, tv: 2.6145116709, vmn: 2.3773561485, vmx: 20.0 },
            Snap { i: 100, tv: 2.7536803664, vmn: 2.4892742850, vmx: 20.0 },
            Snap { i: 150, tv: -6.4606404251, vmn: 20.0, vmx: 20.0 },
            Snap { i: 200, tv: -10.0, vmn: 20.0, vmx: 20.0 },
            Snap { i: 251, tv: -0.1894989954, vmn: 0.5847573715, vmx: 20.0 },
        ];

        let mut si = 0;
        for i in 0..input.len() {
            let (h, tv) = x.update(input[i], i as i64);

            assert_eq!(h.parameter_first, -10.0, "[{}] parameter_first", i);
            assert_eq!(h.parameter_last, 10.0, "[{}] parameter_last", i);
            assert!(
                (h.parameter_resolution - 2.45).abs() < 1e-9,
                "[{}] parameter_resolution: {}",
                i,
                h.parameter_resolution
            );

            if !x.is_primed() {
                assert!(h.is_empty(), "[{}] expected empty heatmap before priming", i);
                assert!(tv.is_nan(), "[{}] expected NaN tv before priming", i);
                continue;
            }

            assert_eq!(
                h.values.len(),
                50,
                "[{}] heatmap values length",
                i
            );

            if si < snapshots.len() && snapshots[si].i == i {
                assert!(
                    (snapshots[si].tv - tv).abs() < TOLERANCE,
                    "[{}] tv: expected {}, got {}",
                    i,
                    snapshots[si].tv,
                    tv
                );
                assert!(
                    (snapshots[si].vmn - h.value_min).abs() < TOLERANCE,
                    "[{}] vmin: expected {}, got {}",
                    i,
                    snapshots[si].vmn,
                    h.value_min
                );
                assert!(
                    (snapshots[si].vmx - h.value_max).abs() < TOLERANCE,
                    "[{}] vmax: expected {}, got {}",
                    i,
                    snapshots[si].vmx,
                    h.value_max
                );
                si += 1;
            }
        }

        assert_eq!(si, snapshots.len(), "did not hit all snapshots");
    }

    #[test]
    fn test_primes_at_bar_11() {
        let input = testdata::test_input();
        let mut x =
            CoronaTrendVigor::new(&CoronaTrendVigorParams::default()).unwrap();

        assert!(!x.is_primed());

        let mut primed_at: Option<usize> = None;
        for i in 0..input.len() {
            x.update(input[i], i as i64);
            if x.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert_eq!(primed_at, Some(11), "expected priming at index 11");
    }

    #[test]
    fn test_nan_input() {
        let mut x =
            CoronaTrendVigor::new(&CoronaTrendVigorParams::default()).unwrap();

        let (h, tv) = x.update(f64::NAN, 0);
        assert!(h.is_empty());
        assert!(tv.is_nan());
        assert!(!x.is_primed());
    }

    #[test]
    fn test_metadata() {
        let x =
            CoronaTrendVigor::new(&CoronaTrendVigorParams::default()).unwrap();
        let md = x.metadata();

        let mn_value = "ctv(50, 20, -10, 10, 30, hl/2)";
        let mn_tv = "ctv-tv(30, hl/2)";

        assert_eq!(md.identifier, Identifier::CoronaTrendVigor);
        assert_eq!(md.mnemonic, mn_value);
        assert_eq!(md.description, format!("Corona trend vigor {}", mn_value));
        assert_eq!(md.outputs.len(), 2);

        assert_eq!(md.outputs[0].kind, CoronaTrendVigorOutput::Value as i32);
        assert_eq!(md.outputs[0].mnemonic, mn_value);

        assert_eq!(
            md.outputs[1].kind,
            CoronaTrendVigorOutput::TrendVigor as i32
        );
        assert_eq!(md.outputs[1].mnemonic, mn_tv);
    }

    #[test]
    fn test_update_entity() {
        let input = testdata::test_input();
        let time = 1000i64;

        let prime = |x: &mut CoronaTrendVigor| {
            for i in 0..50 {
                x.update(input[i % input.len()], time);
            }
        };

        // Scalar
        {
            let mut x =
                CoronaTrendVigor::new(&CoronaTrendVigorParams::default()).unwrap();
            prime(&mut x);
            let out = x.update_scalar(&Scalar::new(time, 100.0));
            assert_eq!(out.len(), 2);
            assert!(out[0].downcast_ref::<Heatmap>().is_some());
            assert!(out[1].downcast_ref::<Scalar>().is_some());
        }

        // Bar
        {
            let mut x =
                CoronaTrendVigor::new(&CoronaTrendVigorParams::default()).unwrap();
            prime(&mut x);
            let bar = Bar::new(time, 0.0, 100.5, 99.5, 100.0, 0.0);
            let out = x.update_bar(&bar);
            assert_eq!(out.len(), 2);
            assert!(out[0].downcast_ref::<Heatmap>().is_some());
            assert!(out[1].downcast_ref::<Scalar>().is_some());
        }

        // Quote
        {
            let mut x =
                CoronaTrendVigor::new(&CoronaTrendVigorParams::default()).unwrap();
            prime(&mut x);
            let quote = Quote::new(time, 100.0, 100.0, 0.0, 0.0);
            let out = x.update_quote(&quote);
            assert_eq!(out.len(), 2);
            assert!(out[0].downcast_ref::<Heatmap>().is_some());
            assert!(out[1].downcast_ref::<Scalar>().is_some());
        }

        // Trade
        {
            let mut x =
                CoronaTrendVigor::new(&CoronaTrendVigorParams::default()).unwrap();
            prime(&mut x);
            let trade = Trade::new(time, 100.0, 0.0);
            let out = x.update_trade(&trade);
            assert_eq!(out.len(), 2);
            assert!(out[0].downcast_ref::<Heatmap>().is_some());
            assert!(out[1].downcast_ref::<Scalar>().is_some());
        }
    }

    #[test]
    fn test_validation_errors() {
        assert!(CoronaTrendVigor::new(&CoronaTrendVigorParams {
            raster_length: 1,
            ..Default::default()
        })
        .is_err());

        assert!(CoronaTrendVigor::new(&CoronaTrendVigorParams {
            min_parameter_value: 5.0,
            max_parameter_value: 5.0,
            ..Default::default()
        })
        .is_err());

        assert!(CoronaTrendVigor::new(&CoronaTrendVigorParams {
            high_pass_filter_cutoff: 1,
            ..Default::default()
        })
        .is_err());

        assert!(CoronaTrendVigor::new(&CoronaTrendVigorParams {
            minimal_period: 1,
            ..Default::default()
        })
        .is_err());

        assert!(CoronaTrendVigor::new(&CoronaTrendVigorParams {
            minimal_period: 10,
            maximal_period: 10,
            ..Default::default()
        })
        .is_err());
    }
}
