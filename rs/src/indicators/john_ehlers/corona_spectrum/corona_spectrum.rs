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
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Corona Spectrum indicator.
pub struct CoronaSpectrumParams {
    /// Minimal raster value (z) of the heatmap, in decibels. Default 6.
    pub min_raster_value: f64,
    /// Maximal raster value (z) of the heatmap, in decibels. Default 20.
    pub max_raster_value: f64,
    /// Minimal ordinate (y) value — minimal cycle period. Default 6.
    pub min_parameter_value: f64,
    /// Maximal ordinate (y) value — maximal cycle period. Default 30.
    pub max_parameter_value: f64,
    /// High-pass filter cutoff (de-trending period). Default 30.
    pub high_pass_filter_cutoff: i32,
    /// Bar component. `None` → Median (hl/2).
    pub bar_component: Option<BarComponent>,
    /// Quote component. `None` → Mid.
    pub quote_component: Option<QuoteComponent>,
    /// Trade component. `None` → Price.
    pub trade_component: Option<TradeComponent>,
}

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

// ---------------------------------------------------------------------------
// Output enum
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CoronaSpectrumOutput {
    Value = 1,
    DominantCycle = 2,
    DominantCycleMedian = 3,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Corona Spectrum heatmap indicator.
///
/// Measures cyclic activity over a cycle period range in a bank of contiguous
/// bandpass filters. Outputs a heatmap column (decibels), dominant cycle, and
/// dominant cycle median.
pub struct CoronaSpectrum {
    mnemonic: String,
    description: String,
    mnemonic_dc: String,
    description_dc: String,
    mnemonic_dcm: String,
    description_dcm: String,
    corona: Corona,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CoronaSpectrum {
    pub fn new(p: &CoronaSpectrumParams) -> Result<Self, String> {
        let invalid = "invalid corona spectrum parameters";

        let min_raster = if p.min_raster_value == 0.0 { 6.0 } else { p.min_raster_value };
        let max_raster = if p.max_raster_value == 0.0 { 20.0 } else { p.max_raster_value };
        let min_pv = if p.min_parameter_value == 0.0 { 6.0 } else { p.min_parameter_value };
        let max_pv = if p.max_parameter_value == 0.0 { 30.0 } else { p.max_parameter_value };
        let hp = if p.high_pass_filter_cutoff == 0 { 30 } else { p.high_pass_filter_cutoff };

        if min_raster < 0.0 {
            return Err(format!("{}: MinRasterValue should be >= 0", invalid));
        }
        if max_raster <= min_raster {
            return Err(format!("{}: MaxRasterValue should be > MinRasterValue", invalid));
        }

        let min_param = min_pv.ceil();
        let max_param = max_pv.floor();

        if min_param < 2.0 {
            return Err(format!("{}: MinParameterValue should be >= 2", invalid));
        }
        if max_param <= min_param {
            return Err(format!("{}: MaxParameterValue should be > MinParameterValue", invalid));
        }
        if hp < 2 {
            return Err(format!("{}: HighPassFilterCutoff should be >= 2", invalid));
        }

        let bc = p.bar_component.unwrap_or(BarComponent::Median);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let corona = Corona::new(&CoronaParams {
            high_pass_filter_cutoff: hp,
            minimal_period: min_param as i32,
            maximal_period: max_param as i32,
            decibels_lower_threshold: min_raster,
            decibels_upper_threshold: max_raster,
        })?;

        let comp_mn = component_triple_mnemonic(bc, qc, tc);
        let parameter_resolution =
            (corona.filter_bank_length() as f64 - 1.0) / (max_param - min_param);

        let mnemonic = format!(
            "cspect({}, {}, {}, {}, {}{})",
            min_raster, max_raster, min_param, max_param, hp, comp_mn
        );
        let mnemonic_dc = format!("cspect-dc({}{})", hp, comp_mn);
        let mnemonic_dcm = format!("cspect-dcm({}{})", hp, comp_mn);

        Ok(Self {
            description: format!("Corona spectrum {}", mnemonic),
            mnemonic,
            description_dc: format!("Corona spectrum dominant cycle {}", mnemonic_dc),
            mnemonic_dc,
            description_dcm: format!("Corona spectrum dominant cycle median {}", mnemonic_dcm),
            mnemonic_dcm,
            corona,
            min_parameter_value: min_param,
            max_parameter_value: max_param,
            parameter_resolution,
            bar_func: bar_component_value(bc),
            quote_func: quote_component_value(qc),
            trade_func: trade_component_value(tc),
        })
    }

    /// Feed the next sample and return (heatmap, dominant_cycle, dominant_cycle_median).
    pub fn update(&mut self, sample: f64, time: i64) -> (Heatmap, f64, f64) {
        if sample.is_nan() {
            return (
                Heatmap::empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                f64::NAN,
                f64::NAN,
            );
        }

        let primed = self.corona.update(sample);
        if !primed {
            return (
                Heatmap::empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                f64::NAN,
                f64::NAN,
            );
        }

        let bank = self.corona.filter_bank();
        let mut values = Vec::with_capacity(bank.len());
        let mut value_min = f64::INFINITY;
        let mut value_max = f64::NEG_INFINITY;

        for f in bank {
            let v = f.decibels;
            values.push(v);
            if v < value_min { value_min = v; }
            if v > value_max { value_max = v; }
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

        (heatmap, self.corona.dominant_cycle(), self.corona.dominant_cycle_median())
    }
}

impl Indicator for CoronaSpectrum {
    fn is_primed(&self) -> bool {
        self.corona.is_primed()
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::CoronaSpectrum,
            &self.mnemonic,
            &self.description,
            &[
                OutputText { mnemonic: self.mnemonic.clone(), description: self.description.clone() },
                OutputText { mnemonic: self.mnemonic_dc.clone(), description: self.description_dc.clone() },
                OutputText { mnemonic: self.mnemonic_dcm.clone(), description: self.description_dcm.clone() },
            ],
        )
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        let (h, dc, dcm) = self.update(sample, bar.time);
        vec![
            Box::new(h) as Box<dyn Any>,
            Box::new(Scalar::new(bar.time, dc)),
            Box::new(Scalar::new(bar.time, dcm)),
        ]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.quote_func)(quote);
        let (h, dc, dcm) = self.update(sample, quote.time);
        vec![
            Box::new(h) as Box<dyn Any>,
            Box::new(Scalar::new(quote.time, dc)),
            Box::new(Scalar::new(quote.time, dcm)),
        ]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.trade_func)(trade);
        let (h, dc, dcm) = self.update(sample, trade.time);
        vec![
            Box::new(h) as Box<dyn Any>,
            Box::new(Scalar::new(trade.time, dc)),
            Box::new(Scalar::new(trade.time, dcm)),
        ]
    }

    fn update_scalar(&mut self, scalar: &Scalar) -> Output {
        let (h, dc, dcm) = self.update(scalar.value, scalar.time);
        vec![
            Box::new(h) as Box<dyn Any>,
            Box::new(Scalar::new(scalar.time, dc)),
            Box::new(Scalar::new(scalar.time, dcm)),
        ]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;
    const TOLERANCE: f64 = 1e-4;

    #[test]
    fn test_corona_spectrum_update() {
        let input = testdata::talib_input();

        struct Snap { i: usize, dc: f64, dcm: f64 }
        let snapshots = [
            Snap { i: 11,  dc: 17.7604672565, dcm: 17.7604672565 },
            Snap { i: 12,  dc: 6.0000000000,  dcm: 6.0000000000 },
            Snap { i: 50,  dc: 15.9989078712, dcm: 15.9989078712 },
            Snap { i: 100, dc: 14.7455497547, dcm: 14.7455497547 },
            Snap { i: 150, dc: 17.5000000000, dcm: 17.2826036069 },
            Snap { i: 200, dc: 19.7557338512, dcm: 20.0000000000 },
            Snap { i: 251, dc: 6.0000000000,  dcm: 6.0000000000 },
        ];

        let mut x = CoronaSpectrum::new(&CoronaSpectrumParams::default()).unwrap();

        let mut si = 0;
        for (i, &v) in input.iter().enumerate() {
            let (h, dc, dcm) = x.update(v, i as i64);

            // Heatmap axis invariants.
            assert_eq!(h.parameter_first, 6.0, "[{}] parameter_first", i);
            assert_eq!(h.parameter_last, 30.0, "[{}] parameter_last", i);
            assert_eq!(h.parameter_resolution, 2.0, "[{}] parameter_resolution", i);

            if !x.is_primed() {
                assert!(h.is_empty(), "[{}] expected empty heatmap before priming", i);
                assert!(dc.is_nan(), "[{}] expected NaN dc before priming", i);
                assert!(dcm.is_nan(), "[{}] expected NaN dcm before priming", i);
                continue;
            }

            assert_eq!(h.values.len(), 49, "[{}] heatmap values length", i);

            if si < snapshots.len() && snapshots[si].i == i {
                assert!(
                    (snapshots[si].dc - dc).abs() < TOLERANCE,
                    "[{}] dc: expected {}, got {}", i, snapshots[si].dc, dc
                );
                assert!(
                    (snapshots[si].dcm - dcm).abs() < TOLERANCE,
                    "[{}] dcm: expected {}, got {}", i, snapshots[si].dcm, dcm
                );
                si += 1;
            }
        }

        assert_eq!(si, snapshots.len(), "did not hit all snapshots");
    }

    #[test]
    fn test_corona_spectrum_primes_at_bar_11() {
        let mut x = CoronaSpectrum::new(&CoronaSpectrumParams::default()).unwrap();
        assert!(!x.is_primed());

        let input = testdata::talib_input();
        let mut primed_at: Option<usize> = None;

        for (i, &v) in input.iter().enumerate() {
            x.update(v, i as i64);
            if x.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }

        assert_eq!(primed_at, Some(11), "expected priming at index 11");
    }

    #[test]
    fn test_corona_spectrum_nan_input() {
        let mut x = CoronaSpectrum::new(&CoronaSpectrumParams::default()).unwrap();
        let (h, dc, dcm) = x.update(f64::NAN, 0);
        assert!(h.is_empty());
        assert!(dc.is_nan());
        assert!(dcm.is_nan());
        assert!(!x.is_primed());
    }

    #[test]
    fn test_corona_spectrum_metadata() {
        let x = CoronaSpectrum::new(&CoronaSpectrumParams::default()).unwrap();
        let md = x.metadata();

        assert_eq!(md.identifier, Identifier::CoronaSpectrum);
        assert_eq!(md.mnemonic, "cspect(6, 20, 6, 30, 30, hl/2)");
        assert_eq!(md.description, "Corona spectrum cspect(6, 20, 6, 30, 30, hl/2)");
        assert_eq!(md.outputs.len(), 3);

        assert_eq!(md.outputs[0].kind, 1); // Value
        assert_eq!(md.outputs[0].mnemonic, "cspect(6, 20, 6, 30, 30, hl/2)");
        assert_eq!(md.outputs[1].kind, 2); // DominantCycle
        assert_eq!(md.outputs[1].mnemonic, "cspect-dc(30, hl/2)");
        assert_eq!(md.outputs[2].kind, 3); // DominantCycleMedian
        assert_eq!(md.outputs[2].mnemonic, "cspect-dcm(30, hl/2)");
    }

    #[test]
    fn test_corona_spectrum_update_bar() {
        let input = testdata::talib_input();
        let mut x = CoronaSpectrum::new(&CoronaSpectrumParams::default()).unwrap();

        // Prime with 50 samples.
        for (i, &v) in input.iter().take(50).enumerate() {
            x.update(v, i as i64);
        }

        let bar = Bar::new(100, 100.0, 100.0, 100.0, 100.0, 0.0);
        let out = x.update_bar(&bar);
        assert_eq!(out.len(), 3);
    }

    #[test]
    fn test_corona_spectrum_invalid_params() {
        // MaxRasterValue <= MinRasterValue
        assert!(CoronaSpectrum::new(&CoronaSpectrumParams {
            min_raster_value: 10.0,
            max_raster_value: 10.0,
            ..CoronaSpectrumParams::default()
        }).is_err());

        // MinParameterValue < 2
        assert!(CoronaSpectrum::new(&CoronaSpectrumParams {
            min_parameter_value: 1.0,
            ..CoronaSpectrumParams::default()
        }).is_err());

        // MaxParameterValue <= MinParameterValue
        assert!(CoronaSpectrum::new(&CoronaSpectrumParams {
            min_parameter_value: 20.0,
            max_parameter_value: 20.0,
            ..CoronaSpectrumParams::default()
        }).is_err());

        // HighPassFilterCutoff < 2
        assert!(CoronaSpectrum::new(&CoronaSpectrumParams {
            high_pass_filter_cutoff: 1,
            ..CoronaSpectrumParams::default()
        }).is_err());
    }

    #[test]
    fn test_corona_spectrum_custom_ranges() {
        let x = CoronaSpectrum::new(&CoronaSpectrumParams {
            min_raster_value: 4.0,
            max_raster_value: 25.0,
            min_parameter_value: 8.7,  // ceils to 9
            max_parameter_value: 40.4, // floors to 40
            high_pass_filter_cutoff: 20,
            ..CoronaSpectrumParams::default()
        }).unwrap();

        assert_eq!(x.min_parameter_value, 9.0);
        assert_eq!(x.max_parameter_value, 40.0);
        assert_eq!(x.mnemonic, "cspect(4, 25, 9, 40, 20, hl/2)");
    }
}
