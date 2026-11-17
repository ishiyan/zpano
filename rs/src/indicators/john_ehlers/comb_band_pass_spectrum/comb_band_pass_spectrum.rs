use std::any::Any;
use std::f64::consts::{PI, SQRT_2};

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

const TWO_PI: f64 = 2.0 * PI;
const SQRT_TWO: f64 = SQRT_2;
const AGC_DECAY: f64 = 0.995;
const CENTROID_THRESHOLD: f64 = 0.5;
const BANK_STARTUP_BARS: usize = 12;

// ---------------------------------------------------------------------------
// Output enum
// ---------------------------------------------------------------------------

/// Output indices for Ehlers' Comb Band-Pass Spectrum indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum CombBandPassSpectrumOutput {
    /// Spectrum heatmap column across `[min_period, max_period]`.
    Value = 1,
    /// Center-of-gravity dominant cycle period estimate.
    DominantCycle = 2,
}

impl CombBandPassSpectrumOutput {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Value => "value",
            Self::DominantCycle => "dominantCycle",
        }
    }

    pub fn is_known(&self) -> bool {
        matches!(self, Self::Value | Self::DominantCycle)
    }
}

impl std::fmt::Display for CombBandPassSpectrumOutput {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}

// ---------------------------------------------------------------------------
// Params & mnemonics
// ---------------------------------------------------------------------------

/// Parameters to create a `CombBandPassSpectrum` instance.
///
/// Defaults correspond to Ehlers' 2013 Book Listing 10-1 settings.
/// Explicit values (including `0` or `false`) are validated directly without silent substitution.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct CombBandPassSpectrumParams {
    /// Shortest evaluated integer cycle period in bars (>= 3). Default is 10.
    pub min_period: i32,
    /// Longest evaluated integer cycle period in bars (>= min_period). Default is 48.
    pub max_period: i32,
    /// Fractional band-pass resonator bandwidth (`0 < bandwidth < min_period / 4`). Default is 0.3.
    pub bandwidth: f64,
    /// Two-pole roofing high-pass cutoff period in bars (>= 5.0). Default is 48.0.
    pub highpass_period: f64,
    /// Two-pole SuperSmoother cutoff period in bars (>= 2.0). Default is 10.0.
    pub smoothing_period: f64,
    /// Divide each resonator sample by its period before squaring and summing. Default is true.
    pub spectral_dilation_compensation: bool,
    /// Cube normalized display heatmap intensities after centroid calculation. Default is false.
    pub enhance_resolution: bool,
    /// Bar component to extract (`None` defaults to `Close` and is omitted from the mnemonic).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract (`None` defaults to `Mid` and is omitted from the mnemonic).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract (`None` defaults to `Price` and is omitted from the mnemonic).
    pub trade_component: Option<TradeComponent>,
}

pub type Params = CombBandPassSpectrumParams;

impl Default for CombBandPassSpectrumParams {
    fn default() -> Self {
        Self {
            min_period: 10,
            max_period: 48,
            bandwidth: 0.3,
            highpass_period: 48.0,
            smoothing_period: 10.0,
            spectral_dilation_compensation: true,
            enhance_resolution: false,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Returns default Ehlers Listing 10-1 parameters.
pub fn default_params() -> CombBandPassSpectrumParams {
    CombBandPassSpectrumParams::default()
}

/// Formats a float parameter cleanly for mnemonics without `-0`.
pub fn format_number(value: f64) -> String {
    let normalized = if value == 0.0 { 0.0 } else { value };
    format!("{normalized}")
}

fn format_mnemonic(
    prefix: &str,
    include_enhance: bool,
    params: &CombBandPassSpectrumParams,
) -> String {
    let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
    let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
    let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);
    let comp_suffix = component_triple_mnemonic(bar_comp, quote_comp, trade_comp);
    let sdc_suffix = if params.spectral_dilation_compensation {
        ""
    } else {
        ", no-sdc"
    };
    let enh_suffix = if include_enhance && params.enhance_resolution {
        ", enh"
    } else {
        ""
    };
    let bw_text = format_number(params.bandwidth);
    let hp_text = format_number(params.highpass_period);
    let lp_text = format_number(params.smoothing_period);
    format!(
        "{prefix}({}, {}, {bw_text}, {hp_text}, {lp_text}{sdc_suffix}{enh_suffix}{comp_suffix})",
        params.min_period, params.max_period
    )
}

/// Builds the primary heatmap mnemonic for `CombBandPassSpectrum`.
pub fn mnemonic(params: &CombBandPassSpectrumParams) -> String {
    format_mnemonic("cbps", true, params)
}

/// Builds the dominant-cycle mnemonic for `CombBandPassSpectrum`.
pub fn dominant_cycle_mnemonic(params: &CombBandPassSpectrumParams) -> String {
    format_mnemonic("cbps-dc", false, params)
}

// ---------------------------------------------------------------------------
// Output & state snapshots
// ---------------------------------------------------------------------------

/// Spectrum heatmap column and dominant cycle period returned per observation.
#[derive(Debug, Clone)]
pub struct SpectrumOutput {
    pub heatmap: Heatmap,
    pub dominant_cycle: f64,
}

/// Snapshot of committed internal roofing, spectral, and centroid state.
#[derive(Debug, Clone, PartialEq)]
pub struct SpectrumState {
    pub high_pass: f64,
    pub filtered: f64,
    pub raw_power: Vec<f64>,
    pub normalization_maximum: f64,
    pub internal_dominant_cycle: f64,
}

/// Validated numeric parameters resolved at construction time.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct ResolvedParams {
    pub min_period: usize,
    pub max_period: usize,
    pub bandwidth: f64,
    pub highpass_period: f64,
    pub smoothing_period: f64,
    pub spectral_dilation_compensation: bool,
    pub enhance_resolution: bool,
}

// ---------------------------------------------------------------------------
// Internal two-pole roofing filter (Book Equation 3-3 / Listing 10-1)
// ---------------------------------------------------------------------------

struct RoofFilter {
    hp0: f64,
    hp1: f64,
    hp2: f64,
    ss1: f64,
    ss2: f64,
    ss3: f64,
    x1: f64,
    x2: f64,
    h1: f64,
    h2: f64,
    f1: f64,
    f2: f64,
}

impl RoofFilter {
    fn new(highpass_period: f64, smoothing_period: f64) -> Result<Self, String> {
        if !highpass_period.is_finite() || highpass_period < 5.0 {
            return Err("highpass_period must be a finite real >= 5".to_string());
        }
        if !smoothing_period.is_finite() || smoothing_period < 2.0 {
            return Err("smoothing_period must be a finite real >= 2".to_string());
        }

        let angle = TWO_PI / (SQRT_TWO * highpass_period);
        let cosine = angle.cos();
        let alpha = (cosine + angle.sin() - 1.0) / cosine;
        let pole = 1.0 - alpha;
        let attenuation = (-SQRT_TWO * PI / smoothing_period).exp();
        let hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
        let hp1 = 2.0 * pole;
        let hp2 = pole * pole;
        let ss2 = 2.0 * attenuation * (SQRT_TWO * PI / smoothing_period).cos();
        let ss3 = -attenuation * attenuation;
        let ss1 = 1.0 - ss2 - ss3;

        if !(0.0 < pole && pole < 1.0 && 0.0 < attenuation && attenuation < 1.0 && ss1 > 0.0)
            || !hp0.is_finite()
            || !hp1.is_finite()
            || !hp2.is_finite()
            || !ss1.is_finite()
            || !ss2.is_finite()
            || !ss3.is_finite()
        {
            return Err(
                "periods must produce finite resolvable stable coefficients".to_string(),
            );
        }

        Ok(Self {
            hp0,
            hp1,
            hp2,
            ss1,
            ss2,
            ss3,
            x1: 0.0,
            x2: 0.0,
            h1: 0.0,
            h2: 0.0,
            f1: 0.0,
            f2: 0.0,
        })
    }

    fn reset(&mut self) {
        self.x1 = 0.0;
        self.x2 = 0.0;
        self.h1 = 0.0;
        self.h2 = 0.0;
        self.f1 = 0.0;
        self.f2 = 0.0;
    }

    fn candidate(&self, price: f64) -> Result<(f64, f64), String> {
        let high = self.hp0 * (price - 2.0 * self.x1 + self.x2)
            + self.hp1 * self.h1
            - self.hp2 * self.h2;
        let filtered = self.ss1 * (high + self.h1) / 2.0
            + self.ss2 * self.f1
            + self.ss3 * self.f2;
        if !high.is_finite() || !filtered.is_finite() {
            return Err("roofing arithmetic overflowed".to_string());
        }
        Ok((high, filtered))
    }

    fn commit(&mut self, price: f64, high: f64, filtered: f64) {
        self.x2 = self.x1;
        self.x1 = price;
        self.h2 = self.h1;
        self.h1 = high;
        self.f2 = self.f1;
        self.f1 = filtered;
    }
}

// ---------------------------------------------------------------------------
// Indicator implementation
// ---------------------------------------------------------------------------

/// Ehlers' 2013 Book Listing 10-1 Comb Band-Pass Spectrum indicator.
pub struct CombBandPassSpectrum {
    resolved: ResolvedParams,
    metadata: Metadata,
    roof: RoofFilter,
    periods: Vec<usize>,
    alpha: Vec<f64>,
    beta: Vec<f64>,
    bp: Vec<Vec<f64>>,
    raw: Vec<f64>,
    sample_count: usize,
    maximum: f64,
    dominant: f64,
    has_dominant: bool,
    last_output: SpectrumOutput,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl CombBandPassSpectrum {
    /// Creates a new `CombBandPassSpectrum` with the supplied parameters.
    pub fn new(params: &CombBandPassSpectrumParams) -> Result<Self, String> {
        const INVALID: &str = "invalid comb band-pass spectrum parameters";

        if params.min_period < 3 {
            return Err(format!("{INVALID}: min_period must be an integer >= 3"));
        }
        if params.max_period < params.min_period {
            return Err(format!(
                "{INVALID}: max_period must be an integer >= min_period"
            ));
        }
        if !params.bandwidth.is_finite()
            || params.bandwidth <= 0.0
            || params.bandwidth >= (params.min_period as f64) / 4.0
        {
            return Err(format!(
                "{INVALID}: bandwidth must satisfy 0 < bandwidth < min_period / 4"
            ));
        }

        let roof = RoofFilter::new(params.highpass_period, params.smoothing_period)
            .map_err(|err| format!("{INVALID}: {err}"))?;

        let min_period = params.min_period as usize;
        let max_period = params.max_period as usize;

        let periods: Vec<usize> = (min_period..=max_period).collect();
        let mut alpha_vec = Vec::with_capacity(periods.len());
        let mut beta_vec = Vec::with_capacity(periods.len());
        let mut bp_rows = Vec::with_capacity(periods.len());
        for &period in &periods {
            let period_f = period as f64;
            let beta_item = (TWO_PI / period_f).cos();
            let cosine_bw = (TWO_PI * params.bandwidth / period_f).cos();
            if !cosine_bw.is_finite() || cosine_bw <= 0.0 {
                return Err(format!(
                    "{INVALID}: bandwidth produces non-positive cosine term"
                ));
            }
            let gamma_item = 1.0 / cosine_bw;
            let radicand = gamma_item * gamma_item - 1.0;
            if !radicand.is_finite() || radicand < 0.0 {
                return Err(format!("{INVALID}: bandwidth produces invalid gamma term"));
            }
            let alpha_item = gamma_item - radicand.sqrt();
            if !alpha_item.is_finite()
                || !(0.0 < alpha_item && alpha_item < 1.0)
                || !beta_item.is_finite()
            {
                return Err(format!(
                    "{INVALID}: bandwidth must produce resolvable stable poles"
                ));
            }
            alpha_vec.push(alpha_item);
            beta_vec.push(beta_item);
            bp_rows.push(vec![0.0; period]);
        }

        let resolved = ResolvedParams {
            min_period,
            max_period,
            bandwidth: params.bandwidth,
            highpass_period: params.highpass_period,
            smoothing_period: params.smoothing_period,
            spectral_dilation_compensation: params.spectral_dilation_compensation,
            enhance_resolution: params.enhance_resolution,
        };

        let bar_comp = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let quote_comp = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let trade_comp = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let mnemonic_text = mnemonic(params);
        let description_text = format!("Comb band-pass spectrum {mnemonic_text}");
        let mnemonic_dc = dominant_cycle_mnemonic(params);
        let description_dc =
            format!("Comb band-pass spectrum dominant cycle {mnemonic_dc}");

        let metadata = build_metadata(
            Identifier::CombBandPassSpectrum,
            &mnemonic_text,
            &description_text,
            &[
                OutputText {
                    mnemonic: mnemonic_text.clone(),
                    description: description_text.clone(),
                },
                OutputText {
                    mnemonic: mnemonic_dc,
                    description: description_dc,
                },
            ],
        );

        let initial_output = SpectrumOutput {
            heatmap: Heatmap::empty(0, min_period as f64, max_period as f64, 1.0),
            dominant_cycle: f64::NAN,
        };

        Ok(Self {
            resolved,
            metadata,
            roof,
            raw: vec![0.0; periods.len()],
            periods,
            alpha: alpha_vec,
            beta: beta_vec,
            bp: bp_rows,
            sample_count: 0,
            maximum: 0.0,
            dominant: 0.0,
            has_dominant: false,
            last_output: initial_output,
            bar_func: bar_component_value(bar_comp),
            quote_func: quote_component_value(quote_comp),
            trade_func: trade_component_value(trade_comp),
        })
    }

    /// Creates a new `CombBandPassSpectrum` with default Listing 10-1 parameters.
    pub fn new_default() -> Result<Self, String> {
        Self::new(&CombBandPassSpectrumParams::default())
    }

    /// Returns the validated resolved parameters.
    pub fn parameters(&self) -> ResolvedParams {
        self.resolved
    }

    /// Returns the shortest evaluated cycle period.
    pub fn min_period(&self) -> usize {
        self.resolved.min_period
    }

    /// Returns the longest evaluated cycle period.
    pub fn max_period(&self) -> usize {
        self.resolved.max_period
    }

    /// Returns the number of warmup observations (`12 + max_period - 1`).
    pub fn warmup_bars(&self) -> usize {
        BANK_STARTUP_BARS + self.resolved.max_period - 1
    }

    /// Returns the number of accepted finite observations seen since construction or reset.
    pub fn samples_seen(&self) -> usize {
        self.sample_count
    }

    /// Returns an owned copy of each resonator's newest-first `P`-sample history.
    pub fn band_pass_history(&self) -> Vec<Vec<f64>> {
        self.bp.clone()
    }

    /// Returns an owned snapshot of committed internal state.
    pub fn state(&self) -> SpectrumState {
        SpectrumState {
            high_pass: self.roof.h1,
            filtered: self.roof.f1,
            raw_power: self.raw.clone(),
            normalization_maximum: self.maximum,
            internal_dominant_cycle: self.dominant,
        }
    }

    /// Returns an owned copy of the most recent accepted output.
    pub fn output(&self) -> SpectrumOutput {
        self.last_output.clone()
    }

    /// Resets all roofing, resonator, power, AGC, and centroid state to initial values.
    pub fn reset(&mut self) {
        self.roof.reset();
        self.sample_count = 0;
        for row in &mut self.bp {
            row.fill(0.0);
        }
        self.raw.fill(0.0);
        self.maximum = 0.0;
        self.dominant = 0.0;
        self.has_dominant = false;
        self.last_output = SpectrumOutput {
            heatmap: Heatmap::empty(
                0,
                self.resolved.min_period as f64,
                self.resolved.max_period as f64,
                1.0,
            ),
            dominant_cycle: f64::NAN,
        };
    }

    /// Fallibly updates the indicator with a scalar sample and timestamp.
    ///
    /// - `NaN` samples are skipped without mutating state or aging `samples_seen`.
    /// - `±Inf` samples or candidate arithmetic overflows return `Err` without mutating state.
    pub fn try_update(&mut self, sample: f64, time: i64) -> Result<SpectrumOutput, String> {
        if sample.is_infinite() {
            return Err("price must be finite or NaN (missing)".to_string());
        }
        let min_f = self.resolved.min_period as f64;
        let max_f = self.resolved.max_period as f64;
        if sample.is_nan() {
            return Ok(SpectrumOutput {
                heatmap: Heatmap::empty(time, min_f, max_f, 1.0),
                dominant_cycle: f64::NAN,
            });
        }

        let (high, filtered) = self.roof.candidate(sample)?;
        let difference = filtered - self.roof.f2;

        let mut next_bp = Vec::with_capacity(self.periods.len());
        let mut next_raw = Vec::with_capacity(self.periods.len());
        for (bin_index, &period) in self.periods.iter().enumerate() {
            let old_row = &self.bp[bin_index];
            let current = if self.sample_count + 1 > BANK_STARTUP_BARS {
                let a_coeff = self.alpha[bin_index];
                let b_coeff = self.beta[bin_index];
                0.5 * (1.0 - a_coeff) * difference + b_coeff * (1.0 + a_coeff) * old_row[0]
                    - a_coeff * old_row[1]
            } else {
                0.0
            };
            if !current.is_finite() {
                return Err("band-pass arithmetic overflowed".to_string());
            }

            let mut next_row = Vec::with_capacity(period);
            next_row.push(current);
            next_row.extend_from_slice(&old_row[..period - 1]);

            let mut power = 0.0;
            if self.sample_count + 1 > BANK_STARTUP_BARS {
                let compensation = if self.resolved.spectral_dilation_compensation {
                    period as f64
                } else {
                    1.0
                };
                for &sample_bp in &next_row {
                    let scaled = sample_bp / compensation;
                    power += scaled * scaled;
                }
            }
            if !power.is_finite() || power < 0.0 {
                return Err("spectral power arithmetic overflowed".to_string());
            }
            next_bp.push(next_row);
            next_raw.push(power);
        }

        let mut maximum = AGC_DECAY * self.maximum;
        for &power in &next_raw {
            if power > maximum {
                maximum = power;
            }
        }

        let mut normalized = Vec::with_capacity(next_raw.len());
        for &power in &next_raw {
            normalized.push(if maximum > 0.0 { power / maximum } else { 0.0 });
        }

        let mut sp = 0.0;
        let mut spx = 0.0;
        for (&period, &power) in self.periods.iter().zip(normalized.iter()) {
            if power >= CENTROID_THRESHOLD {
                spx += (period as f64) * power;
                sp += power;
            }
        }

        let mut dominant = self.dominant;
        let mut has_dominant = self.has_dominant;
        if sp != 0.0 {
            dominant = spx / sp;
            has_dominant = true;
        }
        if !dominant.is_finite() {
            return Err("centroid arithmetic overflowed".to_string());
        }

        let values: Vec<f64> = if self.resolved.enhance_resolution {
            normalized.iter().map(|&item| item * item * item).collect()
        } else {
            normalized
        };

        self.roof.commit(sample, high, filtered);
        self.bp = next_bp;
        self.raw = next_raw;
        self.maximum = maximum;
        self.dominant = dominant;
        self.has_dominant = has_dominant;
        self.sample_count += 1;

        let current_output = if self.sample_count > self.warmup_bars() {
            let mut value_min = values[0];
            let mut value_max = values[0];
            for &item in &values[1..] {
                if item < value_min {
                    value_min = item;
                }
                if item > value_max {
                    value_max = item;
                }
            }
            SpectrumOutput {
                heatmap: Heatmap::new(time, min_f, max_f, 1.0, value_min, value_max, values),
                dominant_cycle: if has_dominant { dominant } else { f64::NAN },
            }
        } else {
            SpectrumOutput {
                heatmap: Heatmap::empty(time, min_f, max_f, 1.0),
                dominant_cycle: f64::NAN,
            }
        };

        self.last_output = current_output.clone();
        Ok(current_output)
    }

    /// Updates the indicator with a scalar sample and timestamp, panicking on invalid input/overflow.
    pub fn update(&mut self, sample: f64, time: i64) -> SpectrumOutput {
        self.try_update(sample, time)
            .unwrap_or_else(|err| panic!("CombBandPassSpectrum::update failed: {err}"))
    }

    fn wrap_output(time: i64, output: SpectrumOutput) -> Output {
        vec![
            Box::new(output.heatmap) as Box<dyn Any>,
            Box::new(Scalar::new(time, output.dominant_cycle)),
        ]
    }
}

impl Indicator for CombBandPassSpectrum {
    fn is_primed(&self) -> bool {
        self.sample_count > self.warmup_bars()
    }

    fn metadata(&self) -> Metadata {
        self.metadata.clone()
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
    use crate::indicators::core::outputs::shape::Shape;
    use crate::indicators::john_ehlers::comb_band_pass_spectrum::testdata;

    const FIXTURE_TOLERANCE: f64 = 1e-11;
    const ORACLE_TOLERANCE: f64 = 2e-11;
    const EXACT_TOLERANCE: f64 = 1e-13;

    fn assert_close_tol(actual: f64, expected: f64, tolerance: f64, label: &str) {
        if expected.is_nan() {
            assert!(actual.is_nan(), "{label}: expected NaN, got {actual}");
            return;
        }
        assert!(
            actual.is_finite(),
            "{label}: expected finite {expected}, got {actual}"
        );
        let bound = tolerance.max(expected.abs() * tolerance);
        let diff = (actual - expected).abs();
        assert!(
            diff <= bound,
            "{label}: actual={actual}, expected={expected}, diff={diff}, bound={bound}"
        );
    }

    fn assert_close(actual: f64, expected: f64, label: &str) {
        assert_close_tol(actual, expected, EXACT_TOLERANCE, label);
    }

    fn assert_same_output(actual: &SpectrumOutput, expected: &SpectrumOutput, label: &str) {
        assert_eq!(actual.heatmap.time, expected.heatmap.time, "{label}: time");
        assert_eq!(
            actual.heatmap.parameter_first, expected.heatmap.parameter_first,
            "{label}: parameter_first"
        );
        assert_eq!(
            actual.heatmap.parameter_last, expected.heatmap.parameter_last,
            "{label}: parameter_last"
        );
        assert_eq!(
            actual.heatmap.parameter_resolution, expected.heatmap.parameter_resolution,
            "{label}: parameter_resolution"
        );
        assert_eq!(
            actual.heatmap.is_empty(),
            expected.heatmap.is_empty(),
            "{label}: is_empty"
        );
        assert_close(actual.heatmap.value_min, expected.heatmap.value_min, label);
        assert_close(actual.heatmap.value_max, expected.heatmap.value_max, label);
        assert_eq!(
            actual.heatmap.values.len(),
            expected.heatmap.values.len(),
            "{label}: values len"
        );
        for (bin_index, (&lhs, &rhs)) in actual
            .heatmap
            .values
            .iter()
            .zip(expected.heatmap.values.iter())
            .enumerate()
        {
            assert_close(lhs, rhs, &format!("{label}[{bin_index}]"));
        }
        assert_close(actual.dominant_cycle, expected.dominant_cycle, label);
    }

    struct ListingOracle {
        params: ResolvedParams,
        bp: Vec<Vec<f64>>,
        prices: Vec<f64>,
        hp: Vec<f64>,
        filt: Vec<f64>,
        max_power: f64,
        dominant: f64,
        identified: bool,
    }

    impl ListingOracle {
        fn new(params: ResolvedParams) -> Self {
            let bins_len = params.max_period - params.min_period + 1;
            let stride = params.max_period + 1;
            Self {
                params,
                bp: vec![vec![0.0; stride]; bins_len],
                prices: Vec::new(),
                hp: Vec::new(),
                filt: Vec::new(),
                max_power: 0.0,
                dominant: 0.0,
                identified: false,
            }
        }

        fn lookback(series: &[f64], lag: usize) -> f64 {
            if series.len() > lag {
                series[series.len() - 1 - lag]
            } else {
                0.0
            }
        }

        fn update(&mut self, price: f64) -> (Vec<f64>, f64, f64, f64) {
            let theta = SQRT_2 * PI / self.params.highpass_period;
            let alpha = (theta.cos() + theta.sin() - 1.0) / theta.cos();
            let high_pass = (1.0 - alpha / 2.0)
                * (1.0 - alpha / 2.0)
                * (price - 2.0 * Self::lookback(&self.prices, 0) + Self::lookback(&self.prices, 1))
                + 2.0 * (1.0 - alpha) * Self::lookback(&self.hp, 0)
                - (1.0 - alpha) * (1.0 - alpha) * Self::lookback(&self.hp, 1);
            let a1 = (-SQRT_2 * PI / self.params.smoothing_period).exp();
            let c2 = 2.0 * a1 * (SQRT_2 * PI / self.params.smoothing_period).cos();
            let c3 = -a1 * a1;
            let c1 = 1.0 - c2 - c3;
            let filtered = c1 * (high_pass + Self::lookback(&self.hp, 0)) / 2.0
                + c2 * Self::lookback(&self.filt, 0)
                + c3 * Self::lookback(&self.filt, 1);

            self.prices.push(price);
            self.hp.push(high_pass);
            self.filt.push(filtered);

            let bins_len = self.params.max_period - self.params.min_period + 1;
            let mut powers = Vec::with_capacity(bins_len);
            for offset in 0..bins_len {
                let period = self.params.min_period + offset;
                let period_f = period as f64;
                let comp = if self.params.spectral_dilation_compensation {
                    period_f
                } else {
                    1.0
                };
                let row = &mut self.bp[offset];
                for slot in (2..=self.params.max_period).rev() {
                    row[slot] = row[slot - 1];
                }
                row[1] = 0.0;
                if self.prices.len() > BANK_STARTUP_BARS {
                    let beta = (2.0 * PI / period_f).cos();
                    let gamma = 1.0 / (2.0 * PI * self.params.bandwidth / period_f).cos();
                    let pole = gamma - (gamma * gamma - 1.0).sqrt();
                    row[1] = 0.5 * (1.0 - pole) * (filtered - Self::lookback(&self.filt, 2))
                        + beta * (1.0 + pole) * row[2]
                        - pole * row[3];
                }

                let mut power = 0.0;
                for slot in 1..=period {
                    power += (row[slot] / comp) * (row[slot] / comp);
                }
                powers.push(power);
            }

            self.max_power = 0.995 * self.max_power;
            for &power in &powers {
                if power > self.max_power {
                    self.max_power = power;
                }
            }

            let normalized: Vec<f64> = powers
                .iter()
                .map(|&power| {
                    if self.max_power > 0.0 {
                        power / self.max_power
                    } else {
                        0.0
                    }
                })
                .collect();

            let mut sp = 0.0;
            let mut spx = 0.0;
            for (offset, &power) in normalized.iter().enumerate() {
                if power >= 0.5 {
                    spx += ((self.params.min_period + offset) as f64) * power;
                    sp += power;
                }
            }
            if sp != 0.0 {
                self.dominant = spx / sp;
                self.identified = true;
            }

            let values: Vec<f64> = if self.params.enhance_resolution {
                normalized.iter().map(|&item| item * item * item).collect()
            } else {
                normalized
            };

            if self.prices.len() < BANK_STARTUP_BARS + self.params.max_period {
                return (Vec::new(), f64::NAN, high_pass, filtered);
            }
            let dc = if self.identified {
                self.dominant
            } else {
                f64::NAN
            };
            (values, dc, high_pass, filtered)
        }
    }

    struct ReferenceCase {
        name: &'static str,
        params: CombBandPassSpectrumParams,
        expected_heatmap: Vec<Vec<f64>>,
        expected_min: Vec<f64>,
        expected_max: Vec<f64>,
        expected_dc: Vec<f64>,
    }

    fn verify_reference_case(case: &ReferenceCase, prices: &[f64]) {
        let mut indicator = CombBandPassSpectrum::new(&case.params).unwrap();
        let mut oracle = ListingOracle::new(indicator.parameters());
        let expected_bins = (case.params.max_period - case.params.min_period + 1) as usize;
        let expected_warmup = BANK_STARTUP_BARS + (case.params.max_period as usize) - 1;
        assert_eq!(indicator.warmup_bars(), expected_warmup, "{}", case.name);

        for (bar_index, &price) in prices.iter().enumerate() {
            let output = indicator.update(price, bar_index as i64);
            let (oracle_values, oracle_dc, oracle_hp, oracle_filt) = oracle.update(price);
            let snapshot = indicator.state();
            assert_close_tol(
                snapshot.high_pass,
                oracle_hp,
                ORACLE_TOLERANCE,
                &format!("{}:hp[{bar_index}]", case.name),
            );
            assert_close_tol(
                snapshot.filtered,
                oracle_filt,
                ORACLE_TOLERANCE,
                &format!("{}:filt[{bar_index}]", case.name),
            );

            assert_eq!(indicator.samples_seen(), bar_index + 1);
            assert_eq!(output.heatmap.time, bar_index as i64);
            assert_eq!(output.heatmap.parameter_first, case.params.min_period as f64);
            assert_eq!(output.heatmap.parameter_last, case.params.max_period as f64);
            assert_eq!(output.heatmap.parameter_resolution, 1.0);

            let exp_row = &case.expected_heatmap[bar_index];
            assert_eq!(exp_row.len(), expected_bins);

            if bar_index < expected_warmup {
                assert!(!indicator.is_primed(), "{}: premature prime", case.name);
                assert!(output.heatmap.is_empty(), "{}: expected empty", case.name);
                assert!(output.heatmap.value_min.is_nan());
                assert!(output.heatmap.value_max.is_nan());
                assert!(output.dominant_cycle.is_nan());
                assert!(case.expected_min[bar_index].is_nan());
                assert!(case.expected_max[bar_index].is_nan());
                assert!(case.expected_dc[bar_index].is_nan());
                for &cell in exp_row {
                    assert!(cell.is_nan());
                }
            } else {
                assert!(indicator.is_primed(), "{}: expected primed", case.name);
                assert!(!output.heatmap.is_empty());
                assert_eq!(output.heatmap.values.len(), expected_bins);
                for bin_index in 0..expected_bins {
                    assert_close_tol(
                        output.heatmap.values[bin_index],
                        exp_row[bin_index],
                        FIXTURE_TOLERANCE,
                        &format!("{}:fixture[{bar_index}][{bin_index}]", case.name),
                    );
                    assert_close_tol(
                        output.heatmap.values[bin_index],
                        oracle_values[bin_index],
                        ORACLE_TOLERANCE,
                        &format!("{}:oracle[{bar_index}][{bin_index}]", case.name),
                    );
                }
                assert_close_tol(
                    output.heatmap.value_min,
                    case.expected_min[bar_index],
                    FIXTURE_TOLERANCE,
                    &format!("{}:min[{bar_index}]", case.name),
                );
                assert_close_tol(
                    output.heatmap.value_max,
                    case.expected_max[bar_index],
                    FIXTURE_TOLERANCE,
                    &format!("{}:max[{bar_index}]", case.name),
                );
                assert_close_tol(
                    output.dominant_cycle,
                    case.expected_dc[bar_index],
                    FIXTURE_TOLERANCE,
                    &format!("{}:dc[{bar_index}]", case.name),
                );
                assert_close_tol(
                    output.dominant_cycle,
                    oracle_dc,
                    ORACLE_TOLERANCE,
                    &format!("{}:oracle_dc[{bar_index}]", case.name),
                );
            }
        }
    }

    #[test]
    fn test_reference_fixtures_and_listing_oracle() {
        let prices = testdata::testdata::test_input();
        let cases = [
            ReferenceCase {
                name: "default_p10_48_hp48_lp10_sdc1_e0_b0_3",
                params: CombBandPassSpectrumParams::default(),
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_3_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_3_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_3_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_3_dominant_cycle(),
            },
            ReferenceCase {
                name: "uncompensated_sdc0_e0_b0_3",
                params: CombBandPassSpectrumParams {
                    spectral_dilation_compensation: false,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc0_e0_b0_3_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc0_e0_b0_3_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc0_e0_b0_3_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc0_e0_b0_3_dominant_cycle(),
            },
            ReferenceCase {
                name: "enhanced_sdc1_e1_b0_3",
                params: CombBandPassSpectrumParams {
                    enhance_resolution: true,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc1_e1_b0_3_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc1_e1_b0_3_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc1_e1_b0_3_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc1_e1_b0_3_dominant_cycle(),
            },
            ReferenceCase {
                name: "uncomp_enhanced_sdc0_e1_b0_3",
                params: CombBandPassSpectrumParams {
                    spectral_dilation_compensation: false,
                    enhance_resolution: true,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc0_e1_b0_3_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc0_e1_b0_3_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc0_e1_b0_3_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc0_e1_b0_3_dominant_cycle(),
            },
            ReferenceCase {
                name: "narrow_bandwidth_b0_1",
                params: CombBandPassSpectrumParams {
                    bandwidth: 0.1,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_1_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_1_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_1_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_1_dominant_cycle(),
            },
            ReferenceCase {
                name: "wide_bandwidth_b0_7",
                params: CombBandPassSpectrumParams {
                    bandwidth: 0.7,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_7_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_7_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_7_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_sdc1_e0_b0_7_dominant_cycle(),
            },
            ReferenceCase {
                name: "fractional_smoothing_lp10_5",
                params: CombBandPassSpectrumParams {
                    smoothing_period: 10.5,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp48_lp10_5_sdc1_e0_b0_3_heatmap(),
                expected_min: testdata::expected_p010_p148_hp48_lp10_5_sdc1_e0_b0_3_min(),
                expected_max: testdata::expected_p010_p148_hp48_lp10_5_sdc1_e0_b0_3_max(),
                expected_dc: testdata::expected_p010_p148_hp48_lp10_5_sdc1_e0_b0_3_dominant_cycle(),
            },
            ReferenceCase {
                name: "independent_roof_hp80_lp20",
                params: CombBandPassSpectrumParams {
                    highpass_period: 80.0,
                    smoothing_period: 20.0,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p148_hp80_lp20_sdc1_e0_b0_3_heatmap(),
                expected_min: testdata::expected_p010_p148_hp80_lp20_sdc1_e0_b0_3_min(),
                expected_max: testdata::expected_p010_p148_hp80_lp20_sdc1_e0_b0_3_max(),
                expected_dc: testdata::expected_p010_p148_hp80_lp20_sdc1_e0_b0_3_dominant_cycle(),
            },
            ReferenceCase {
                name: "short_range_p6_24",
                params: CombBandPassSpectrumParams {
                    min_period: 6,
                    max_period: 24,
                    highpass_period: 24.0,
                    smoothing_period: 6.0,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p06_p124_hp24_lp6_sdc1_e0_b0_3_heatmap(),
                expected_min: testdata::expected_p06_p124_hp24_lp6_sdc1_e0_b0_3_min(),
                expected_max: testdata::expected_p06_p124_hp24_lp6_sdc1_e0_b0_3_max(),
                expected_dc: testdata::expected_p06_p124_hp24_lp6_sdc1_e0_b0_3_dominant_cycle(),
            },
            ReferenceCase {
                name: "single_bin_p20_20",
                params: CombBandPassSpectrumParams {
                    min_period: 20,
                    max_period: 20,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p020_p120_hp48_lp10_sdc1_e0_b0_3_heatmap(),
                expected_min: testdata::expected_p020_p120_hp48_lp10_sdc1_e0_b0_3_min(),
                expected_max: testdata::expected_p020_p120_hp48_lp10_sdc1_e0_b0_3_max(),
                expected_dc: testdata::expected_p020_p120_hp48_lp10_sdc1_e0_b0_3_dominant_cycle(),
            },
            ReferenceCase {
                name: "long_range_p24_60",
                params: CombBandPassSpectrumParams {
                    min_period: 24,
                    max_period: 60,
                    highpass_period: 60.0,
                    smoothing_period: 12.0,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p024_p160_hp60_lp12_sdc1_e0_b0_3_heatmap(),
                expected_min: testdata::expected_p024_p160_hp60_lp12_sdc1_e0_b0_3_min(),
                expected_max: testdata::expected_p024_p160_hp60_lp12_sdc1_e0_b0_3_max(),
                expected_dc: testdata::expected_p024_p160_hp60_lp12_sdc1_e0_b0_3_dominant_cycle(),
            },
            ReferenceCase {
                name: "boundary_p10_240",
                params: CombBandPassSpectrumParams {
                    min_period: 10,
                    max_period: 240,
                    highpass_period: 240.0,
                    ..Default::default()
                },
                expected_heatmap: testdata::expected_p010_p1240_hp240_lp10_sdc1_e0_b0_3_heatmap(),
                expected_min: testdata::expected_p010_p1240_hp240_lp10_sdc1_e0_b0_3_min(),
                expected_max: testdata::expected_p010_p1240_hp240_lp10_sdc1_e0_b0_3_max(),
                expected_dc: testdata::expected_p010_p1240_hp240_lp10_sdc1_e0_b0_3_dominant_cycle(),
            },
        ];
        for case in &cases {
            verify_reference_case(case, &prices);
        }
    }

    #[test]
    fn test_startup_gate_zero_series_nan_overflow_and_reset() {
        let prices = testdata::testdata::test_input();

        let mut gated = CombBandPassSpectrum::new_default().unwrap();
        for bar_index in 0..BANK_STARTUP_BARS {
            let out = gated.update(prices[bar_index], bar_index as i64);
            assert!(out.heatmap.is_empty());
            assert!(out.dominant_cycle.is_nan());
            let state_now = gated.state();
            assert!(state_now.raw_power.iter().all(|&power| power == 0.0));
            for row in gated.band_pass_history() {
                assert!(row.iter().all(|&sample_bp| sample_bp == 0.0));
            }
        }
        let _ = gated.update(prices[BANK_STARTUP_BARS], BANK_STARTUP_BARS as i64);
        assert!(gated
            .band_pass_history()
            .iter()
            .any(|row| row[0] != 0.0));

        let mut zero_ind = CombBandPassSpectrum::new_default().unwrap();
        for step in 0..80 {
            let out = zero_ind.update(0.0, step as i64);
            if step >= zero_ind.warmup_bars() {
                assert!(zero_ind.is_primed());
                assert!(!out.heatmap.is_empty());
                assert_eq!(out.heatmap.value_min, 0.0);
                assert_eq!(out.heatmap.value_max, 0.0);
                assert!(out.heatmap.values.iter().all(|&item| item == 0.0));
                assert!(out.dominant_cycle.is_nan());
            }
        }
        for step in 80..180 {
            let sample = (2.0 * PI * (step as f64) / 20.0).sin();
            let _ = zero_ind.update(sample, step as i64);
        }
        let held_dc = zero_ind.output().dominant_cycle;
        assert!(held_dc.is_finite());
        let silent_out = zero_ind.update(0.0, 180);
        assert!(silent_out.dominant_cycle.is_finite());

        let mut instance = CombBandPassSpectrum::new_default().unwrap();
        let mut control = CombBandPassSpectrum::new_default().unwrap();
        for bar_index in 0..120 {
            let _ = instance.update(prices[bar_index], bar_index as i64);
            let _ = control.update(prices[bar_index], bar_index as i64);
        }

        let saved_state = instance.state();
        let saved_bp = instance.band_pass_history();
        let saved_output = instance.output();
        let saved_count = instance.samples_seen();

        let nan_out = instance.update(f64::NAN, 999);
        assert!(nan_out.heatmap.is_empty());
        assert_eq!(nan_out.heatmap.time, 999);
        assert!(nan_out.dominant_cycle.is_nan());
        assert_eq!(instance.state(), saved_state);
        assert_eq!(instance.band_pass_history(), saved_bp);
        assert_same_output(&instance.output(), &saved_output, "after_nan");
        assert_eq!(instance.samples_seen(), saved_count);

        assert!(instance.try_update(f64::INFINITY, 1000).is_err());
        assert!(instance.try_update(f64::NEG_INFINITY, 1001).is_err());
        assert!(instance.try_update(1e308, 1002).is_err());
        assert_eq!(instance.state(), saved_state);
        assert_eq!(instance.band_pass_history(), saved_bp);
        assert_same_output(&instance.output(), &saved_output, "after_err");
        assert_eq!(instance.samples_seen(), saved_count);

        for bar_index in 120..150 {
            let lhs = instance.update(prices[bar_index], bar_index as i64);
            let rhs = control.update(prices[bar_index], bar_index as i64);
            assert_same_output(&lhs, &rhs, &format!("resume[{bar_index}]"));
        }

        instance.reset();
        assert!(!instance.is_primed());
        assert_eq!(instance.samples_seen(), 0);
        assert!(instance.output().heatmap.is_empty());
        assert!(instance.output().dominant_cycle.is_nan());

        let mut fresh = CombBandPassSpectrum::new_default().unwrap();
        for bar_index in 0..80 {
            let lhs = instance.update(prices[bar_index], bar_index as i64);
            let rhs = fresh.update(prices[bar_index], bar_index as i64);
            assert_same_output(&lhs, &rhs, &format!("post_reset[{bar_index}]"));
        }
    }

    #[test]
    fn test_compensation_enhancement_and_sine_wave_peak() {
        let prices = testdata::testdata::test_input();
        let mut compensated = CombBandPassSpectrum::new_default().unwrap();
        let mut uncompensated = CombBandPassSpectrum::new(&CombBandPassSpectrumParams {
            spectral_dilation_compensation: false,
            ..Default::default()
        })
        .unwrap();
        let mut enhanced = CombBandPassSpectrum::new(&CombBandPassSpectrumParams {
            enhance_resolution: true,
            ..Default::default()
        })
        .unwrap();

        for (bar_index, &price) in prices.iter().enumerate() {
            let out_comp = compensated.update(price, bar_index as i64);
            let _ = uncompensated.update(price, bar_index as i64);
            let out_enh = enhanced.update(price, bar_index as i64);
            assert_eq!(compensated.state(), enhanced.state());
            assert_close(out_comp.dominant_cycle, out_enh.dominant_cycle, "dc_enh");
            let comp_raw = compensated.state().raw_power;
            let uncomp_raw = uncompensated.state().raw_power;
            for bin_index in 0..39 {
                let period_f = (10 + bin_index) as f64;
                assert_close(
                    comp_raw[bin_index],
                    uncomp_raw[bin_index] / (period_f * period_f),
                    "sdc_ratio",
                );
            }
            if !out_comp.heatmap.is_empty() {
                for bin_index in 0..39 {
                    let item = out_comp.heatmap.values[bin_index];
                    assert_close(
                        out_enh.heatmap.values[bin_index],
                        item * item * item,
                        "enh_cube",
                    );
                }
            }
        }

        let mut sine_ind = CombBandPassSpectrum::new_default().unwrap();
        let mut sine_out = sine_ind.output();
        for step in 0..1000 {
            let sample = (2.0 * PI * (step as f64) / 20.0).sin();
            sine_out = sine_ind.update(sample, step as i64);
        }
        let mut best_bin = 0usize;
        for bin_index in 1..sine_out.heatmap.values.len() {
            if sine_out.heatmap.values[bin_index] > sine_out.heatmap.values[best_bin] {
                best_bin = bin_index;
            }
        }
        let peak_period = 10 + best_bin as i32;
        assert!((peak_period - 20).abs() <= 2, "peak_period={peak_period}");
        assert!(
            sine_out.dominant_cycle > 17.0 && sine_out.dominant_cycle < 23.0,
            "dominant_cycle={}",
            sine_out.dominant_cycle
        );
    }

    #[test]
    fn test_metadata_mnemonics_entities_and_validation() {
        let indicator = CombBandPassSpectrum::new_default().unwrap();
        let meta = indicator.metadata();
        assert_eq!(meta.identifier, Identifier::CombBandPassSpectrum);
        assert_eq!(meta.mnemonic, "cbps(10, 48, 0.3, 48, 10)");
        assert_eq!(
            meta.description,
            "Comb band-pass spectrum cbps(10, 48, 0.3, 48, 10)"
        );
        assert_eq!(meta.outputs.len(), 2);
        assert_eq!(
            meta.outputs[0].kind,
            CombBandPassSpectrumOutput::Value as i32
        );
        assert_eq!(meta.outputs[0].shape, Shape::Heatmap);
        assert_eq!(meta.outputs[0].mnemonic, "cbps(10, 48, 0.3, 48, 10)");
        assert_eq!(
            meta.outputs[1].kind,
            CombBandPassSpectrumOutput::DominantCycle as i32
        );
        assert_eq!(meta.outputs[1].shape, Shape::Scalar);
        assert_eq!(meta.outputs[1].mnemonic, "cbps-dc(10, 48, 0.3, 48, 10)");
        assert_eq!(
            meta.outputs[1].description,
            "Comb band-pass spectrum dominant cycle cbps-dc(10, 48, 0.3, 48, 10)"
        );

        let custom = CombBandPassSpectrumParams {
            min_period: 6,
            max_period: 24,
            bandwidth: 0.4,
            highpass_period: 48.0,
            smoothing_period: 10.5,
            spectral_dilation_compensation: false,
            enhance_resolution: true,
            bar_component: Some(BarComponent::Median),
            quote_component: Some(QuoteComponent::Weighted),
            trade_component: Some(TradeComponent::Volume),
        };
        assert_eq!(
            mnemonic(&custom),
            "cbps(6, 24, 0.4, 48, 10.5, no-sdc, enh, hl/2, (bbs+aas)/(bs+as), v)"
        );
        assert_eq!(
            dominant_cycle_mnemonic(&custom),
            "cbps-dc(6, 24, 0.4, 48, 10.5, no-sdc, hl/2, (bbs+aas)/(bs+as), v)"
        );

        let mnemonic_cases: Vec<(CombBandPassSpectrumParams, &str, &str)> = vec![
            (
                default_params(),
                "cbps(10, 48, 0.3, 48, 10)",
                "cbps-dc(10, 48, 0.3, 48, 10)",
            ),
            (
                CombBandPassSpectrumParams {
                    bar_component: Some(BarComponent::Open),
                    ..Default::default()
                },
                "cbps(10, 48, 0.3, 48, 10, o)",
                "cbps-dc(10, 48, 0.3, 48, 10, o)",
            ),
            (
                CombBandPassSpectrumParams {
                    quote_component: Some(QuoteComponent::Bid),
                    ..Default::default()
                },
                "cbps(10, 48, 0.3, 48, 10, b)",
                "cbps-dc(10, 48, 0.3, 48, 10, b)",
            ),
            (
                CombBandPassSpectrumParams {
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "cbps(10, 48, 0.3, 48, 10, v)",
                "cbps-dc(10, 48, 0.3, 48, 10, v)",
            ),
            (
                CombBandPassSpectrumParams {
                    bar_component: Some(BarComponent::High),
                    quote_component: Some(QuoteComponent::Ask),
                    ..Default::default()
                },
                "cbps(10, 48, 0.3, 48, 10, h, a)",
                "cbps-dc(10, 48, 0.3, 48, 10, h, a)",
            ),
            (
                CombBandPassSpectrumParams {
                    bar_component: Some(BarComponent::Low),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "cbps(10, 48, 0.3, 48, 10, l, v)",
                "cbps-dc(10, 48, 0.3, 48, 10, l, v)",
            ),
            (
                CombBandPassSpectrumParams {
                    quote_component: Some(QuoteComponent::SpreadBp),
                    trade_component: Some(TradeComponent::Volume),
                    ..Default::default()
                },
                "cbps(10, 48, 0.3, 48, 10, spread bp, v)",
                "cbps-dc(10, 48, 0.3, 48, 10, spread bp, v)",
            ),
        ];
        for (params_case, expected_m, expected_dc_m) in &mnemonic_cases {
            assert_eq!(&mnemonic(params_case), expected_m);
            assert_eq!(&dominant_cycle_mnemonic(params_case), expected_dc_m);
        }

        assert_eq!(CombBandPassSpectrumOutput::Value.as_str(), "value");
        assert_eq!(
            CombBandPassSpectrumOutput::DominantCycle.as_str(),
            "dominantCycle"
        );
        assert!(CombBandPassSpectrumOutput::Value.is_known());
        assert!(CombBandPassSpectrumOutput::DominantCycle.is_known());
        assert_eq!(format!("{}", CombBandPassSpectrumOutput::Value), "value");
        assert_eq!(
            format!("{}", CombBandPassSpectrumOutput::DominantCycle),
            "dominantCycle"
        );

        let prices = testdata::testdata::test_input();
        let highs = testdata::testdata::test_high();
        let lows = testdata::testdata::test_low();
        let opens = testdata::testdata::test_open();
        let mut bar_ind = CombBandPassSpectrum::new_default().unwrap();
        let mut scalar_ind = CombBandPassSpectrum::new_default().unwrap();
        let mut quote_ind = CombBandPassSpectrum::new_default().unwrap();
        let mut trade_ind = CombBandPassSpectrum::new_default().unwrap();

        for index in 0..70 {
            let bar = Bar {
                time: index as i64,
                open: opens[index],
                high: highs[index],
                low: lows[index],
                close: prices[index],
                volume: 1000.0,
            };
            let scalar = Scalar::new(index as i64, prices[index]);
            let quote = Quote {
                time: index as i64,
                bid_price: prices[index] - 0.5,
                ask_price: prices[index] + 0.5,
                bid_size: 10.0,
                ask_size: 10.0,
            };
            let trade = Trade {
                time: index as i64,
                price: prices[index],
                volume: 50.0,
            };
            let out_bar = bar_ind.update_bar(&bar);
            let out_scalar = scalar_ind.update_scalar(&scalar);
            let out_quote = quote_ind.update_quote(&quote);
            let out_trade = trade_ind.update_trade(&trade);

            assert_eq!(out_bar.len(), 2);
            let hm_bar = out_bar[0].downcast_ref::<Heatmap>().unwrap();
            let hm_scalar = out_scalar[0].downcast_ref::<Heatmap>().unwrap();
            let hm_quote = out_quote[0].downcast_ref::<Heatmap>().unwrap();
            let hm_trade = out_trade[0].downcast_ref::<Heatmap>().unwrap();
            let dc_bar = out_bar[1].downcast_ref::<Scalar>().unwrap();
            let dc_scalar = out_scalar[1].downcast_ref::<Scalar>().unwrap();
            let dc_quote = out_quote[1].downcast_ref::<Scalar>().unwrap();
            let dc_trade = out_trade[1].downcast_ref::<Scalar>().unwrap();

            assert_eq!(hm_bar.values.len(), hm_scalar.values.len());
            assert_eq!(hm_bar.values.len(), hm_quote.values.len());
            assert_eq!(hm_bar.values.len(), hm_trade.values.len());
            for bin_index in 0..hm_bar.values.len() {
                assert_close(hm_bar.values[bin_index], hm_scalar.values[bin_index], "bar vs scalar");
                assert_close(hm_bar.values[bin_index], hm_quote.values[bin_index], "bar vs quote");
                assert_close(hm_bar.values[bin_index], hm_trade.values[bin_index], "bar vs trade");
            }
            assert_close(dc_bar.value, dc_scalar.value, "dc bar vs scalar");
            assert_close(dc_bar.value, dc_quote.value, "dc bar vs quote");
            assert_close(dc_bar.value, dc_trade.value, "dc bar vs trade");
        }

        let bad_cases = [
            CombBandPassSpectrumParams {
                min_period: 2,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                min_period: 0,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                min_period: 10,
                max_period: 9,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                bandwidth: 0.0,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                bandwidth: -0.1,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                bandwidth: 2.5,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                min_period: 3,
                max_period: 10,
                bandwidth: 0.75,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                bandwidth: f64::NAN,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                bandwidth: f64::INFINITY,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                highpass_period: 4.0,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                highpass_period: f64::NAN,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                highpass_period: f64::INFINITY,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                smoothing_period: 1.0,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                smoothing_period: f64::NAN,
                ..Default::default()
            },
            CombBandPassSpectrumParams {
                smoothing_period: f64::INFINITY,
                ..Default::default()
            },
        ];
        for case in &bad_cases {
            assert!(
                CombBandPassSpectrum::new(case).is_err(),
                "expected error for {case:?}"
            );
        }
    }
}
