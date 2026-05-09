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

const MAX_LEAD_LIST_COUNT: usize = 50;
const MAX_POSITION_LIST_COUNT: usize = 20;

/// 60° phase-lead coefficients: lead60 = 0.5·BP + sin(60°)·Q ≈ 0.5·BP + 0.866·Q.
const LEAD60_COEF_BP: f64 = 0.5;
const LEAD60_COEF_Q: f64 = 0.866;

/// Bandpass filter delta factor in γ = 1/cos(ω · 2 · 0.1).
const BP_DELTA: f64 = 0.1;

const WIDTH_HIGH_THRESHOLD: f64 = 0.85;
const WIDTH_HIGH_SATURATE: f64 = 0.8;
const WIDTH_NARROW: f64 = 0.01;
const WIDTH_SCALE: f64 = 0.15;

const RASTER_BLEND_EXPONENT: f64 = 0.95;
const RASTER_BLEND_HALF: f64 = 0.5;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Corona Swing Position indicator.
pub struct CoronaSwingPositionParams {
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

impl Default for CoronaSwingPositionParams {
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

/// Output indices for the Corona Swing Position indicator.
#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CoronaSwingPositionOutput {
    /// The Corona swing position heatmap column.
    Value = 1,
    /// The scalar swing position in [MinParameterValue, MaxParameterValue].
    SwingPosition = 2,
}

// ---------------------------------------------------------------------------
// Indicator struct
// ---------------------------------------------------------------------------

/// Ehlers' Corona Swing Position heatmap indicator.
pub struct CoronaSwingPosition {
    mnemonic: String,
    description: String,
    mnemonic_sp: String,
    description_sp: String,
    c: Corona,
    raster_length: usize,
    raster_step: f64,
    max_raster_value: f64,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    raster: Vec<f64>,
    lead_list: Vec<f64>,
    position_list: Vec<f64>,
    sample_previous: f64,
    sample_previous2: f64,
    band_pass_previous: f64,
    band_pass_previous2: f64,
    swing_position: f64,
    is_started: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CoronaSwingPosition {
    /// Creates a new Corona Swing Position indicator with the given parameters.
    pub fn new(p: &CoronaSwingPositionParams) -> Result<Self, String> {
        let invalid = "invalid corona swing position parameters";

        let mut raster_length = p.raster_length;
        if raster_length == 0 {
            raster_length = 50;
        }

        let mut max_raster_value = p.max_raster_value;
        if max_raster_value == 0.0 {
            max_raster_value = 20.0;
        }

        // MinParameterValue and MaxParameterValue default only when BOTH are zero.
        let mut min_parameter_value = p.min_parameter_value;
        let mut max_parameter_value = p.max_parameter_value;
        if min_parameter_value == 0.0 && max_parameter_value == 0.0 {
            min_parameter_value = -5.0;
            max_parameter_value = 5.0;
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

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let component_mnemonic = component_triple_mnemonic(bc, qc, tc);

        let mnemonic = format!(
            "cswing({}, {}, {}, {}, {}{})",
            raster_length, max_raster_value, min_parameter_value,
            max_parameter_value, high_pass_filter_cutoff, component_mnemonic,
        );
        let mnemonic_sp = format!(
            "cswing-sp({}{})",
            high_pass_filter_cutoff, component_mnemonic,
        );
        let description = format!("Corona swing position {}", mnemonic);
        let description_sp = format!("Corona swing position scalar {}", mnemonic_sp);

        let rl = raster_length as usize;
        let parameter_resolution =
            (rl as f64 - 1.0) / (max_parameter_value - min_parameter_value);

        Ok(Self {
            mnemonic,
            description,
            mnemonic_sp,
            description_sp,
            c,
            raster_length: rl,
            raster_step: max_raster_value / rl as f64,
            max_raster_value,
            min_parameter_value,
            max_parameter_value,
            parameter_resolution,
            raster: vec![0.0; rl],
            lead_list: Vec::with_capacity(MAX_LEAD_LIST_COUNT),
            position_list: Vec::with_capacity(MAX_POSITION_LIST_COUNT),
            sample_previous: 0.0,
            sample_previous2: 0.0,
            band_pass_previous: 0.0,
            band_pass_previous2: 0.0,
            swing_position: f64::NAN,
            is_started: false,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    /// Feeds the next sample and returns (heatmap, swing_position).
    /// On unprimed bars the heatmap is empty and swing_position is NaN.
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

        if !self.is_started {
            self.sample_previous = sample;
            self.is_started = true;

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

        // 60° lead: lead60 = 0.5·BP_previous2 + 0.866·Q
        let lead60 =
            LEAD60_COEF_BP * self.band_pass_previous2 + LEAD60_COEF_Q * quadrature2;

        let (lowest_lead, highest_lead) =
            append_rolling(&mut self.lead_list, MAX_LEAD_LIST_COUNT, lead60);

        // Normalised lead position ∈ [0, 1].
        let range_lead = highest_lead - lowest_lead;
        let position = if range_lead > 0.0 {
            (lead60 - lowest_lead) / range_lead
        } else {
            range_lead
        };

        let (lowest_pos, highest_pos) =
            append_rolling(&mut self.position_list, MAX_POSITION_LIST_COUNT, position);
        let highest = highest_pos - lowest_pos;

        let width = if highest > WIDTH_HIGH_THRESHOLD {
            WIDTH_NARROW
        } else {
            WIDTH_SCALE * highest
        };

        self.swing_position = (self.max_parameter_value - self.min_parameter_value) * position
            + self.min_parameter_value;

        let position_scaled_to_raster_length =
            (position * self.raster_length as f64).round() as i32;
        let position_scaled_to_max_raster_value = position * self.max_raster_value;

        for i in 0..self.raster_length {
            let mut value = self.raster[i];

            if i as i32 == position_scaled_to_raster_length {
                value *= RASTER_BLEND_HALF;
            } else {
                let mut argument =
                    position_scaled_to_max_raster_value - self.raster_step * i as f64;
                if (i as i32) > position_scaled_to_raster_length {
                    argument = -argument;
                }

                if width > 0.0 {
                    value = RASTER_BLEND_HALF
                        * ((argument / width).powf(RASTER_BLEND_EXPONENT)
                            + RASTER_BLEND_HALF * value);
                }
            }

            if value < 0.0 {
                value = 0.0;
            } else if value > self.max_raster_value {
                value = self.max_raster_value;
            }

            if highest > WIDTH_HIGH_SATURATE {
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

        (heatmap, self.swing_position)
    }
}

// ---------------------------------------------------------------------------
// Indicator trait
// ---------------------------------------------------------------------------

impl Indicator for CoronaSwingPosition {
    fn is_primed(&self) -> bool {
        self.c.is_primed()
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CoronaSwingPosition,
            &self.mnemonic,
            &self.description,
            &[
                OutputText {
                    mnemonic: self.mnemonic.clone(),
                    description: self.description.clone(),
                },
                OutputText {
                    mnemonic: self.mnemonic_sp.clone(),
                    description: self.description_sp.clone(),
                },
            ],
        )
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        let time = bar.time;
        let (heatmap, sp) = self.update(sample, time);
        vec![
            Box::new(heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, sp)) as Box<dyn Any>,
        ]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.quote_func)(quote);
        let time = quote.time;
        let (heatmap, sp) = self.update(sample, time);
        vec![
            Box::new(heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, sp)) as Box<dyn Any>,
        ]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.trade_func)(trade);
        let time = trade.time;
        let (heatmap, sp) = self.update(sample, time);
        vec![
            Box::new(heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, sp)) as Box<dyn Any>,
        ]
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let time = scalar.time;
        let (heatmap, sp) = self.update(scalar.value, time);
        vec![
            Box::new(heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, sp)) as Box<dyn Any>,
        ]
    }
}

// ---------------------------------------------------------------------------
// Helper: appendRolling
// ---------------------------------------------------------------------------

/// Appends v to the list, drops the oldest element once the list reaches
/// max_count, and returns the current (lowest, highest) values.
fn append_rolling(list: &mut Vec<f64>, max_count: usize, v: f64) -> (f64, f64) {
    if list.len() >= max_count {
        list.remove(0);
    }
    list.push(v);

    let mut lowest = v;
    let mut highest = v;
    for &x in list.iter() {
        if x < lowest {
            lowest = x;
        }
        if x > highest {
            highest = x;
        }
    }

    (lowest, highest)
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
            CoronaSwingPosition::new(&CoronaSwingPositionParams::default()).unwrap();

        struct Snap {
            i: usize,
            sp: f64,
            vmn: f64,
            vmx: f64,
        }

        let snapshots = vec![
            Snap { i: 11, sp: 5.0, vmn: 20.0, vmx: 20.0 },
            Snap { i: 12, sp: 5.0, vmn: 20.0, vmx: 20.0 },
            Snap { i: 50, sp: 4.5384908349, vmn: 20.0, vmx: 20.0 },
            Snap { i: 100, sp: -3.8183742675, vmn: 3.4957777081, vmx: 20.0 },
            Snap { i: 150, sp: -1.8516194371, vmn: 5.3792287864, vmx: 20.0 },
            Snap { i: 200, sp: -3.6944428668, vmn: 4.2580825738, vmx: 20.0 },
            Snap { i: 251, sp: -0.8524812061, vmn: 4.4822539784, vmx: 20.0 },
        ];

        let mut si = 0;
        for i in 0..input.len() {
            let (h, sp) = x.update(input[i], i as i64);

            assert_eq!(h.parameter_first, -5.0, "[{}] parameter_first", i);
            assert_eq!(h.parameter_last, 5.0, "[{}] parameter_last", i);
            assert!(
                (h.parameter_resolution - 4.9).abs() < 1e-9,
                "[{}] parameter_resolution: {}",
                i,
                h.parameter_resolution
            );

            if !x.is_primed() {
                assert!(h.is_empty(), "[{}] expected empty heatmap before priming", i);
                assert!(sp.is_nan(), "[{}] expected NaN sp before priming", i);
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
                    (snapshots[si].sp - sp).abs() < TOLERANCE,
                    "[{}] sp: expected {}, got {}",
                    i,
                    snapshots[si].sp,
                    sp
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
            CoronaSwingPosition::new(&CoronaSwingPositionParams::default()).unwrap();

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
            CoronaSwingPosition::new(&CoronaSwingPositionParams::default()).unwrap();

        let (h, sp) = x.update(f64::NAN, 0);
        assert!(h.is_empty());
        assert!(sp.is_nan());
        assert!(!x.is_primed());
    }

    #[test]
    fn test_metadata() {
        let x =
            CoronaSwingPosition::new(&CoronaSwingPositionParams::default()).unwrap();
        let md = x.metadata();

        let mn_value = "cswing(50, 20, -5, 5, 30, hl/2)";
        let mn_sp = "cswing-sp(30, hl/2)";

        assert_eq!(md.identifier, Identifier::CoronaSwingPosition);
        assert_eq!(md.mnemonic, mn_value);
        assert_eq!(md.description, format!("Corona swing position {}", mn_value));
        assert_eq!(md.outputs.len(), 2);

        assert_eq!(md.outputs[0].kind, CoronaSwingPositionOutput::Value as i32);
        assert_eq!(md.outputs[0].mnemonic, mn_value);

        assert_eq!(
            md.outputs[1].kind,
            CoronaSwingPositionOutput::SwingPosition as i32
        );
        assert_eq!(md.outputs[1].mnemonic, mn_sp);
    }

    #[test]
    fn test_update_entity() {
        let input = testdata::test_input();
        let time = 1000i64;

        // Helper to prime an indicator.
        let prime = |x: &mut CoronaSwingPosition| {
            for i in 0..50 {
                x.update(input[i % input.len()], time);
            }
        };

        // Scalar
        {
            let mut x =
                CoronaSwingPosition::new(&CoronaSwingPositionParams::default()).unwrap();
            prime(&mut x);
            let out = x.update_scalar(&Scalar::new(time, 100.0));
            assert_eq!(out.len(), 2);
            assert!(out[0].downcast_ref::<Heatmap>().is_some());
            assert!(out[1].downcast_ref::<Scalar>().is_some());
        }

        // Bar
        {
            let mut x =
                CoronaSwingPosition::new(&CoronaSwingPositionParams::default()).unwrap();
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
                CoronaSwingPosition::new(&CoronaSwingPositionParams::default()).unwrap();
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
                CoronaSwingPosition::new(&CoronaSwingPositionParams::default()).unwrap();
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
        // RasterLength < 2
        assert!(CoronaSwingPosition::new(&CoronaSwingPositionParams {
            raster_length: 1,
            ..Default::default()
        })
        .is_err());

        // MaxParameterValue <= MinParameterValue
        assert!(CoronaSwingPosition::new(&CoronaSwingPositionParams {
            min_parameter_value: 5.0,
            max_parameter_value: 5.0,
            ..Default::default()
        })
        .is_err());

        // HighPassFilterCutoff < 2
        assert!(CoronaSwingPosition::new(&CoronaSwingPositionParams {
            high_pass_filter_cutoff: 1,
            ..Default::default()
        })
        .is_err());

        // MinimalPeriod < 2
        assert!(CoronaSwingPosition::new(&CoronaSwingPositionParams {
            minimal_period: 1,
            ..Default::default()
        })
        .is_err());

        // MaximalPeriod <= MinimalPeriod
        assert!(CoronaSwingPosition::new(&CoronaSwingPositionParams {
            minimal_period: 10,
            maximal_period: 10,
            ..Default::default()
        })
        .is_err());
    }
}
