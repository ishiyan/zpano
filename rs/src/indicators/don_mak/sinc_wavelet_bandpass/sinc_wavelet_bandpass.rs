use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Band
// ---------------------------------------------------------------------------

/// Selects the frequency band of the Sinc Wavelet Band-Pass filter.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum Band {
    /// High-frequency band (periods 8-16 bars).
    High = 0,
    /// Mid-frequency band (periods 16-32 bars).
    Mid = 1,
    /// Low-frequency band (periods 32-64 bars).
    Low = 2,
    /// Full band (periods 8-64 bars).
    Full = 3,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Sinc Wavelet Band-Pass indicator.
pub struct SincWaveletBandpassParams {
    /// Frequency band selection. Default Mid.
    pub band: Band,
    /// Whether a cubic velocity kernel is applied to the band-pass output. Default false.
    pub velocity: bool,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for SincWaveletBandpassParams {
    fn default() -> Self {
        Self {
            band: Band::Mid,
            velocity: false,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Sinc Wavelet Band-Pass indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SincWaveletBandpassOutput {
    /// The band-passed price component (or its velocity).
    Value = 1,
}

// ---------------------------------------------------------------------------
// Coefficient computation
// ---------------------------------------------------------------------------

const VELOCITY_TAPS: usize = 4;

// Cubic velocity kernel (PFD degree=3, order=1, smoothing=0).
const VELOCITY_KERNEL: [f64; VELOCITY_TAPS] = [11.0 / 6.0, -3.0, 3.0 / 2.0, -1.0 / 3.0];

/// Returns (omega0, omega1, num_taps) for a band.
fn band_params(band: Band) -> (f64, f64, usize) {
    let pi = std::f64::consts::PI;
    match band {
        Band::High => (pi / 4.0, pi / 8.0, 121),
        Band::Mid => (pi / 8.0, pi / 16.0, 121),
        Band::Low => (pi / 16.0, pi / 32.0, 201),
        Band::Full => (pi / 4.0, pi / 32.0, 201),
    }
}

fn band_name(band: Band) -> &'static str {
    match band {
        Band::High => "high",
        Band::Mid => "mid",
        Band::Low => "low",
        Band::Full => "full",
    }
}

/// Computes sinc band-pass filter coefficients (difference of two sinc functions).
fn compute_coefficients(omega0: f64, omega1: f64, num_taps: usize) -> Vec<f64> {
    let pi = std::f64::consts::PI;
    let mut coeffs = vec![0.0_f64; num_taps];
    coeffs[0] = (omega0 - omega1) / pi;
    for k in 1..num_taps {
        let kf = k as f64;
        let pi_k = pi * kf;
        coeffs[k] = (omega0 * kf).sin() / pi_k - (omega1 * kf).sin() / pi_k;
    }
    coeffs
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Don Mak's Sinc Wavelet Band-Pass (SWB) filter.
///
/// A causal FIR band-pass filter derived from the sinc wavelet system, decomposing
/// price into frequency bands (HIGH, MID, LOW, FULL). Optionally a cubic velocity
/// kernel is applied to produce a momentum oscillator.
pub struct SincWaveletBandpass {
    line: LineIndicator,
    velocity: bool,
    coefficients: Vec<f64>,
    num_taps: usize,
    price_buffer: Vec<f64>,
    price_count: usize,
    price_index: usize,
    vel_buffer: [f64; VELOCITY_TAPS],
    vel_count: usize,
    vel_index: usize,
    primed: bool,
}

impl SincWaveletBandpass {
    /// Creates a new Sinc Wavelet Band-Pass from the given parameters.
    pub fn new(params: &SincWaveletBandpassParams) -> Result<Self, String> {
        let (omega0, omega1, num_taps) = band_params(params.band);

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let cfg = if params.velocity {
            format!("{},v", band_name(params.band))
        } else {
            band_name(params.band).to_string()
        };
        let mnemonic = format!("swb({}{})", cfg, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Sinc wavelet band-pass {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let coefficients = compute_coefficients(omega0, omega1, num_taps);

        Ok(Self {
            line,
            velocity: params.velocity,
            coefficients,
            num_taps,
            price_buffer: vec![0.0; num_taps],
            price_count: 0,
            price_index: 0,
            vel_buffer: [0.0; VELOCITY_TAPS],
            vel_count: 0,
            vel_index: 0,
            primed: false,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update returning the filter output.
    pub fn update(&mut self, sample: f64) -> f64 {
        // Store price in the ring buffer.
        self.price_buffer[self.price_index] = sample;
        self.price_index = (self.price_index + 1) % self.num_taps;
        self.price_count += 1;

        if self.price_count < self.num_taps {
            self.primed = false;
            return f64::NAN;
        }

        // Band-pass convolution: coefficients[k] multiplies the k-th most recent price.
        let mut bp_value = 0.0;
        let n = self.num_taps as i64;
        for k in 0..self.num_taps {
            let offset = self.price_index as i64 - 1 - k as i64;
            let buf_idx = offset.rem_euclid(n) as usize;
            bp_value += self.coefficients[k] * self.price_buffer[buf_idx];
        }

        if !self.velocity {
            self.primed = true;
            return bp_value;
        }

        // Store band-pass output in the velocity ring buffer.
        self.vel_buffer[self.vel_index] = bp_value;
        self.vel_index = (self.vel_index + 1) % VELOCITY_TAPS;
        self.vel_count += 1;

        if self.vel_count < VELOCITY_TAPS {
            self.primed = false;
            return f64::NAN;
        }

        // Cubic velocity: kernel[k] multiplies the k-th most recent band-pass value.
        let mut vel_value = 0.0;
        let vn = VELOCITY_TAPS as i64;
        for k in 0..VELOCITY_TAPS {
            let offset = self.vel_index as i64 - 1 - k as i64;
            let buf_idx = offset.rem_euclid(vn) as usize;
            vel_value += VELOCITY_KERNEL[k] * self.vel_buffer[buf_idx];
        }

        self.primed = true;
        vel_value
    }
}

impl Indicator for SincWaveletBandpass {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::SincWaveletBandpass,
            &self.line.mnemonic,
            &self.line.description,
            &[OutputText {
                mnemonic: self.line.mnemonic.clone(),
                description: self.line.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let sample_value = (self.line.bar_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let sample_value = (self.line.quote_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let sample_value = (self.line.trade_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::test_data;

    const TOLERANCE: f64 = 1e-9;

    fn check_series(name: &str, params: &SincWaveletBandpassParams, inputs: &[f64], expected: &[f64]) {
        let mut ind = SincWaveletBandpass::new(params).unwrap();
        assert_eq!(inputs.len(), expected.len(), "{}: length mismatch", name);
        for i in 0..inputs.len() {
            let value = ind.update(inputs[i]);
            let exp = expected[i];
            if exp.is_nan() {
                assert!(value.is_nan(), "{}[{}]: expected NaN, got {}", name, i, value);
            } else {
                assert!(
                    (value - exp).abs() <= TOLERANCE,
                    "{}[{}]: expected {}, got {}",
                    name, i, exp, value
                );
            }
        }
    }

    #[test]
    fn test_reference_data() {
        let input = test_data::test_input();
        let sine = test_data::test1_input_sine();
        let mixed = test_data::test2_input_mixed();

        check_series("HIGH", &SincWaveletBandpassParams { band: Band::High, ..Default::default() }, &input, &test_data::expected_high());
        check_series("MID", &SincWaveletBandpassParams { band: Band::Mid, ..Default::default() }, &input, &test_data::expected_mid());
        check_series("LOW", &SincWaveletBandpassParams { band: Band::Low, ..Default::default() }, &input, &test_data::expected_low());
        check_series("FULL", &SincWaveletBandpassParams { band: Band::Full, ..Default::default() }, &input, &test_data::expected_full());
        check_series("HIGH_V", &SincWaveletBandpassParams { band: Band::High, velocity: true, ..Default::default() }, &input, &test_data::expected_high_v());
        check_series("MID_V", &SincWaveletBandpassParams { band: Band::Mid, velocity: true, ..Default::default() }, &input, &test_data::expected_mid_v());
        check_series("LOW_V", &SincWaveletBandpassParams { band: Band::Low, velocity: true, ..Default::default() }, &input, &test_data::expected_low_v());
        check_series("FULL_V", &SincWaveletBandpassParams { band: Band::Full, velocity: true, ..Default::default() }, &input, &test_data::expected_full_v());

        check_series("TEST1_MID", &SincWaveletBandpassParams { band: Band::Mid, ..Default::default() }, &sine, &test_data::test1_expected_mid());

        check_series("TEST2_HIGH_V", &SincWaveletBandpassParams { band: Band::High, velocity: true, ..Default::default() }, &mixed, &test_data::test2_expected_high_v());
        check_series("TEST2_MID_V", &SincWaveletBandpassParams { band: Band::Mid, velocity: true, ..Default::default() }, &mixed, &test_data::test2_expected_mid_v());
        check_series("TEST2_LOW_V", &SincWaveletBandpassParams { band: Band::Low, velocity: true, ..Default::default() }, &mixed, &test_data::test2_expected_low_v());
    }

    #[test]
    fn test_mnemonic() {
        assert_eq!(SincWaveletBandpass::new(&SincWaveletBandpassParams::default()).unwrap().metadata().mnemonic, "swb(mid)");
        assert_eq!(SincWaveletBandpass::new(&SincWaveletBandpassParams { band: Band::High, ..Default::default() }).unwrap().metadata().mnemonic, "swb(high)");
        assert_eq!(SincWaveletBandpass::new(&SincWaveletBandpassParams { band: Band::Full, ..Default::default() }).unwrap().metadata().mnemonic, "swb(full)");
        assert_eq!(SincWaveletBandpass::new(&SincWaveletBandpassParams { band: Band::Mid, velocity: true, ..Default::default() }).unwrap().metadata().mnemonic, "swb(mid,v)");
        assert_eq!(SincWaveletBandpass::new(&SincWaveletBandpassParams { band: Band::Full, velocity: true, ..Default::default() }).unwrap().metadata().mnemonic, "swb(full,v)");
    }

    #[test]
    fn test_metadata() {
        let ind = SincWaveletBandpass::new(&SincWaveletBandpassParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::SincWaveletBandpass);
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, SincWaveletBandpassOutput::Value as i32);
    }

    #[test]
    fn test_update_scalar() {
        let input = test_data::test_input();
        let expected = test_data::expected_high();

        let mut ind = SincWaveletBandpass::new(&SincWaveletBandpassParams { band: Band::High, ..Default::default() }).unwrap();
        let mut out: Output = vec![];
        for i in 0..input.len() {
            out = ind.update_scalar(&Scalar { time: 0, value: input[i] });
        }
        assert_eq!(out.len(), 1);
        let last = input.len() - 1;
        let value = out[0].downcast_ref::<Scalar>().unwrap().value;
        assert!((value - expected[last]).abs() <= TOLERANCE);
    }
}
