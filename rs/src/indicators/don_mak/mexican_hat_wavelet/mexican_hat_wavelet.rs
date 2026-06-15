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

/// Selects the frequency band of the Mexican Hat Wavelet filter.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum Band {
    /// High-frequency band (a_f = 1.483, period ~ 4.6 bars).
    High = 0,
    /// Mid-frequency band (a_f = 4.048, period ~ 13.5 bars).
    Mid = 1,
    /// Low-frequency band (a_f = 15.97, period ~ 54 bars).
    Low = 2,
    /// User-specified dilation or period.
    Custom = 3,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Mexican Hat Wavelet indicator.
pub struct MexicanHatWaveletParams {
    /// Frequency band selection. Default Mid.
    pub band: Band,
    /// Custom dilation a_f (used only when band is Custom). > 0. Zero means unset.
    pub dilation: f64,
    /// Custom center period in bars (used only when band is Custom). > 2. Zero means unset.
    pub period: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for MexicanHatWaveletParams {
    fn default() -> Self {
        Self {
            band: Band::Mid,
            dilation: 0.0,
            period: 0.0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Mexican Hat Wavelet indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MexicanHatWaveletOutput {
    /// The bandpass-filtered price component.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Coefficient computation
// ---------------------------------------------------------------------------

// Preset dilation values (a_f) for the three standard bands (Table 5.2).
const DILATION_HIGH: f64 = 1.483;
const DILATION_MID: f64 = 4.048;
const DILATION_LOW: f64 = 15.97;

/// Rounds half to even (banker's rounding), matching Python's round() for x > 0.
fn round_half_even(x: f64) -> f64 {
    let frac = (x - x.trunc()).abs();
    if frac == 0.5 {
        let f = x.floor();
        if (f as i64) % 2 == 0 {
            f
        } else {
            f + 1.0
        }
    } else {
        x.round()
    }
}

/// Computes dilation a_f from a desired center period in bars (Eq 5.11).
fn dilation_from_period(period: f64) -> Result<f64, String> {
    let omega0 = (2.0 * std::f64::consts::PI) / period;
    let two_over_a = 1.091 * omega0 - 0.071 * omega0 * omega0;
    if two_over_a <= 0.0 {
        return Err("invalid mexican hat wavelet parameters: period is too large for the fitting formula (2/a <= 0)".to_string());
    }
    Ok(2.0 / two_over_a)
}

/// Computes normalized Mexican Hat wavelet FIR coefficients for dilation a_f.
fn compute_coefficients(a_f: f64) -> Vec<f64> {
    let mut k = 4 * round_half_even(a_f) as i64;
    if k < 1 {
        k = 1;
    }
    let k = k as usize;

    let norm = 0.488 + 0.646 * a_f + 0.0001 * a_f * a_f;

    let mut coeffs: Vec<f64> = Vec::with_capacity(k + 1);
    for n in 0..=k {
        let t = n as f64 / a_f;
        let t2 = t * t;
        let h_n = (1.0 - 2.0 * t2) * (-t2).exp();
        coeffs.push(h_n / norm);
    }

    coeffs
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Don Mak's Mexican Hat Wavelet (MHW) bandpass filter.
///
/// A causal bandpass FIR filter derived from the Mexican Hat wavelet (the second
/// derivative of a Gaussian), decomposing price into frequency bands with zero
/// phase shift at the filter's center frequency.
pub struct MexicanHatWavelet {
    line: LineIndicator,
    coefficients: Vec<f64>,
    num_taps: usize,
    buffer: Vec<f64>,
    count: usize,
    primed: bool,
}

impl MexicanHatWavelet {
    /// Creates a new Mexican Hat Wavelet from the given parameters.
    pub fn new(params: &MexicanHatWaveletParams) -> Result<Self, String> {
        let invalid = "invalid mexican hat wavelet parameters";

        let (a_f, cfg) = match params.band {
            Band::High => (DILATION_HIGH, "high".to_string()),
            Band::Mid => (DILATION_MID, "mid".to_string()),
            Band::Low => (DILATION_LOW, "low".to_string()),
            Band::Custom => {
                let has_dilation = params.dilation != 0.0;
                let has_period = params.period != 0.0;
                if has_dilation && has_period {
                    return Err(format!("{}: provide only one of dilation or period, not both", invalid));
                }
                if !has_dilation && !has_period {
                    return Err(format!("{}: band=custom requires either dilation or period", invalid));
                }
                if has_period {
                    if params.period <= 2.0 {
                        return Err(format!("{}: period must be > 2", invalid));
                    }
                    (dilation_from_period(params.period)?, format!("p{:.2}", params.period))
                } else {
                    if params.dilation <= 0.0 {
                        return Err(format!("{}: dilation must be > 0", invalid));
                    }
                    (params.dilation, format!("d{:.2}", params.dilation))
                }
            }
        };

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("mhw({}{})", cfg, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Mexican hat wavelet {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        let coefficients = compute_coefficients(a_f);
        let num_taps = coefficients.len();

        Ok(Self {
            line,
            coefficients,
            num_taps,
            buffer: vec![0.0; num_taps],
            count: 0,
            primed: false,
        })
    }

    /// Returns true if the indicator has produced at least one valid output.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    /// Core update returning the filter output.
    pub fn update(&mut self, sample: f64) -> f64 {
        // Shift buffer right and insert the new price at position 0.
        let mut i = self.num_taps - 1;
        while i > 0 {
            self.buffer[i] = self.buffer[i - 1];
            i -= 1;
        }
        self.buffer[0] = sample;
        self.count += 1;

        if self.count < self.num_taps {
            self.primed = false;
            return f64::NAN;
        }

        // FIR convolution: y = sum(coefficients[k] * buffer[k]).
        let mut y = 0.0;
        for k in 0..self.num_taps {
            y += self.coefficients[k] * self.buffer[k];
        }

        self.primed = true;
        y
    }
}

impl Indicator for MexicanHatWavelet {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::MexicanHatWavelet,
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
    use super::super::testdata::testdata;

    const TOLERANCE: f64 = 1e-9;

    fn check_series(name: &str, params: &MexicanHatWaveletParams, inputs: &[f64], expected: &[f64]) {
        let mut ind = MexicanHatWavelet::new(params).unwrap();
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
        let input = testdata::test_input();
        let sine = testdata::test1_input_sine();
        let mixed = testdata::test2_input_mixed();

        check_series("HIGH", &MexicanHatWaveletParams { band: Band::High, ..Default::default() }, &input, &testdata::expected_high());
        check_series("MID", &MexicanHatWaveletParams { band: Band::Mid, ..Default::default() }, &input, &testdata::expected_mid());
        check_series("LOW", &MexicanHatWaveletParams { band: Band::Low, ..Default::default() }, &input, &testdata::expected_low());
        check_series("P8", &MexicanHatWaveletParams { band: Band::Custom, period: 8.0, ..Default::default() }, &input, &testdata::expected_p8());
        check_series("P20", &MexicanHatWaveletParams { band: Band::Custom, period: 20.0, ..Default::default() }, &input, &testdata::expected_p20());
        check_series("P32", &MexicanHatWaveletParams { band: Band::Custom, period: 32.0, ..Default::default() }, &input, &testdata::expected_p32());
        check_series("D2_0", &MexicanHatWaveletParams { band: Band::Custom, dilation: 2.0, ..Default::default() }, &input, &testdata::expected_d2_0());
        check_series("D8_0", &MexicanHatWaveletParams { band: Band::Custom, dilation: 8.0, ..Default::default() }, &input, &testdata::expected_d8_0());

        check_series("TEST1_MID", &MexicanHatWaveletParams { band: Band::Mid, ..Default::default() }, &sine, &testdata::test1_expected_mid());

        check_series("TEST2_HIGH", &MexicanHatWaveletParams { band: Band::High, ..Default::default() }, &mixed, &testdata::test2_expected_high());
        check_series("TEST2_MID", &MexicanHatWaveletParams { band: Band::Mid, ..Default::default() }, &mixed, &testdata::test2_expected_mid());
        check_series("TEST2_LOW", &MexicanHatWaveletParams { band: Band::Low, ..Default::default() }, &mixed, &testdata::test2_expected_low());
    }

    #[test]
    fn test_mnemonic() {
        assert_eq!(MexicanHatWavelet::new(&MexicanHatWaveletParams::default()).unwrap().metadata().mnemonic, "mhw(mid)");
        assert_eq!(MexicanHatWavelet::new(&MexicanHatWaveletParams { band: Band::High, ..Default::default() }).unwrap().metadata().mnemonic, "mhw(high)");
        assert_eq!(MexicanHatWavelet::new(&MexicanHatWaveletParams { band: Band::Low, ..Default::default() }).unwrap().metadata().mnemonic, "mhw(low)");
        assert_eq!(MexicanHatWavelet::new(&MexicanHatWaveletParams { band: Band::Custom, dilation: 2.0, ..Default::default() }).unwrap().metadata().mnemonic, "mhw(d2.00)");
        assert_eq!(MexicanHatWavelet::new(&MexicanHatWaveletParams { band: Band::Custom, period: 20.0, ..Default::default() }).unwrap().metadata().mnemonic, "mhw(p20.00)");
    }

    #[test]
    fn test_metadata() {
        let ind = MexicanHatWavelet::new(&MexicanHatWaveletParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::MexicanHatWavelet);
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(meta.outputs[0].kind, MexicanHatWaveletOutput::Value as i32);
    }

    #[test]
    fn test_update_scalar() {
        let input = testdata::test_input();
        let expected = testdata::expected_high();

        let mut ind = MexicanHatWavelet::new(&MexicanHatWaveletParams { band: Band::High, ..Default::default() }).unwrap();
        let mut out: Output = vec![];
        for i in 0..input.len() {
            out = ind.update_scalar(&Scalar { time: 0, value: input[i] });
        }
        assert_eq!(out.len(), 1);
        let last = input.len() - 1;
        let value = out[0].downcast_ref::<Scalar>().unwrap().value;
        assert!((value - expected[last]).abs() <= TOLERANCE);
    }

    #[test]
    fn test_invalid_params() {
        assert!(MexicanHatWavelet::new(&MexicanHatWaveletParams { band: Band::Custom, ..Default::default() }).is_err());
        assert!(MexicanHatWavelet::new(&MexicanHatWaveletParams { band: Band::Custom, dilation: 2.0, period: 20.0, ..Default::default() }).is_err());
        assert!(MexicanHatWavelet::new(&MexicanHatWaveletParams { band: Band::Custom, period: 2.0, ..Default::default() }).is_err());
        assert!(MexicanHatWavelet::new(&MexicanHatWaveletParams { band: Band::Custom, dilation: -1.0, ..Default::default() }).is_err());
    }
}
