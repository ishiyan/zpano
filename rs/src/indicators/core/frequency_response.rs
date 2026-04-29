/// Frequency response analysis for digital filters.
///
/// Computes power, amplitude, and phase spectra using a direct real FFT
/// of the filter's impulse response.

use std::f64::consts::PI;

use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

/// A single spectrum component (power, amplitude, or phase) with data and
/// min/max bounds.
#[derive(Debug, Clone)]
pub struct Component {
    pub data: Vec<f64>,
    pub min: f64,
    pub max: f64,
}

impl Component {
    fn new(length: usize) -> Self {
        Self {
            data: vec![0.0; length],
            min: f64::NEG_INFINITY,
            max: f64::INFINITY,
        }
    }
}

// ---------------------------------------------------------------------------
// FrequencyResponse
// ---------------------------------------------------------------------------

/// Calculated filter frequency response data.
///
/// All vectors have the same spectrum length (`signal_length / 2 - 1`).
#[derive(Debug, Clone)]
pub struct FrequencyResponse {
    /// Mnemonic of the filter used to calculate the frequency response.
    pub label: String,

    /// Normalized frequency in cycles per 2 samples (1 = Nyquist).
    pub normalized_frequency: Vec<f64>,

    /// Spectrum power in percentages from a maximum value.
    pub power_percent: Component,

    /// Spectrum power in decibels.
    pub power_decibel: Component,

    /// Spectrum amplitude in percentages from a maximum value.
    pub amplitude_percent: Component,

    /// Spectrum amplitude in decibels.
    pub amplitude_decibel: Component,

    /// Phase in degrees in range [-180, 180].
    pub phase_degrees: Component,

    /// Phase in degrees, unwrapped.
    pub phase_degrees_unwrapped: Component,
}

// ---------------------------------------------------------------------------
// Updater trait
// ---------------------------------------------------------------------------

/// Describes a filter whose frequency response is to be calculated.
pub trait Updater {
    fn metadata(&self) -> Metadata;
    fn update(&mut self, sample: f64) -> f64;
}

// ---------------------------------------------------------------------------
// Calculate
// ---------------------------------------------------------------------------

/// Calculates a frequency response of a given impulse signal length using the
/// filter update function.
///
/// * `signal_length` — must be a power of 2 and >= 4 (e.g. 512, 1024, 2048).
/// * `warmup` — how many zero-valued updates before the impulse.
/// * `phase_degrees_unwrapping_limit` — threshold for phase unwrapping (use 179.0).
pub fn calculate(
    signal_length: usize,
    filter: &mut dyn Updater,
    warmup: usize,
    phase_degrees_unwrapping_limit: f64,
) -> Result<FrequencyResponse, String> {
    if !is_valid_signal_length(signal_length) {
        return Err(format!(
            "length should be power of 2 and not less than 4: {}",
            signal_length
        ));
    }

    let spectrum_length = signal_length / 2 - 1;

    let mut fr = FrequencyResponse {
        label: filter.metadata().mnemonic.clone(),
        normalized_frequency: vec![0.0; spectrum_length],
        power_percent: Component::new(spectrum_length),
        power_decibel: Component::new(spectrum_length),
        amplitude_percent: Component::new(spectrum_length),
        amplitude_decibel: Component::new(spectrum_length),
        phase_degrees: Component::new(spectrum_length),
        phase_degrees_unwrapped: Component::new(spectrum_length),
    };

    prepare_frequency_domain(spectrum_length, &mut fr.normalized_frequency);

    let mut signal = prepare_filtered_signal(signal_length, filter, warmup);
    direct_real_fast_fourier_transform(&mut signal);
    parse_spectrum(
        spectrum_length,
        &signal,
        &mut fr.power_percent,
        &mut fr.amplitude_percent,
        &mut fr.phase_degrees,
        &mut fr.phase_degrees_unwrapped,
        phase_degrees_unwrapping_limit,
    );
    to_decibels(spectrum_length, &fr.power_percent.clone(), &mut fr.power_decibel);
    to_percents(spectrum_length, &mut fr.power_percent);
    to_decibels(spectrum_length, &fr.amplitude_percent.clone(), &mut fr.amplitude_decibel);
    to_percents(spectrum_length, &mut fr.amplitude_percent);

    Ok(fr)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

fn is_valid_signal_length(mut length: usize) -> bool {
    if length < 4 {
        return false;
    }
    while length > 4 {
        if length % 2 != 0 {
            return false;
        }
        length /= 2;
    }
    length == 4
}

fn prepare_frequency_domain(spectrum_length: usize, freq: &mut [f64]) {
    for i in 0..spectrum_length {
        freq[i] = (1 + i) as f64 / spectrum_length as f64;
    }
}

fn prepare_filtered_signal(
    signal_length: usize,
    filter: &mut dyn Updater,
    warmup: usize,
) -> Vec<f64> {
    const ZERO: f64 = 0.0;
    const ONE: f64 = 1000.0;

    for _ in 0..warmup {
        filter.update(ZERO);
    }

    let mut signal = vec![0.0; signal_length];
    signal[0] = filter.update(ONE);

    for i in 1..signal_length {
        signal[i] = filter.update(ZERO);
    }

    signal
}

fn parse_spectrum(
    length: usize,
    signal: &[f64],
    power: &mut Component,
    amplitude: &mut Component,
    phase: &mut Component,
    phase_unwrapped: &mut Component,
    phase_degrees_unwrapping_limit: f64,
) {
    const RAD2DEG: f64 = 180.0 / PI;

    let mut pmin = f64::INFINITY;
    let mut pmax = f64::NEG_INFINITY;
    let mut amin = f64::INFINITY;
    let mut amax = f64::NEG_INFINITY;

    let mut k = 2;
    for i in 0..length {
        let re = signal[k];
        k += 1;
        let im = signal[k];
        k += 1;

        // Wrapped phase: atan2 returns radians in [-π, π], convert to [-180, 180].
        phase.data[i] = -im.atan2(re) * RAD2DEG;
        phase_unwrapped.data[i] = 0.0;

        let pwr = re * re + im * im;
        power.data[i] = pwr;
        pmin = pmin.min(pwr);
        pmax = pmax.max(pwr);

        let amp = pwr.sqrt();
        amplitude.data[i] = amp;
        amin = amin.min(amp);
        amax = amax.max(amp);
    }

    unwrap_phase_degrees(length, &phase.data, phase_unwrapped, phase_degrees_unwrapping_limit);
    phase.min = -180.0;
    phase.max = 180.0;
    power.min = pmin;
    power.max = pmax;
    amplitude.min = amin;
    amplitude.max = amax;
}

fn unwrap_phase_degrees(
    length: usize,
    wrapped: &[f64],
    unwrapped: &mut Component,
    limit: f64,
) {
    let mut k = 0.0;

    let mut min = wrapped[0];
    let mut max = min;
    unwrapped.data[0] = min;

    for i in 1..length {
        let mut w = wrapped[i];
        let increment = wrapped[i] - wrapped[i - 1];

        if increment > limit {
            k -= increment;
        } else if increment < -limit {
            k += increment;
        }

        w += k;
        min = min.min(w);
        max = max.max(w);
        unwrapped.data[i] = w;
    }

    unwrapped.min = min;
    unwrapped.max = max;
}

fn to_decibels(length: usize, src: &Component, tgt: &mut Component) {
    let mut dbmin = f64::INFINITY;
    let mut dbmax = f64::NEG_INFINITY;

    let mut base = src.data[0];
    if base < f64::MIN_POSITIVE {
        base = src.max;
    }

    for i in 0..length {
        let db = 20.0 * (src.data[i] / base).log10();
        dbmin = dbmin.min(db);
        dbmax = dbmax.max(db);
        tgt.data[i] = db;
    }

    // Snap dbmin to interval boundaries: [-100, -90), [-90, -80), ...
    for j in (1..=10).rev() {
        let lo = -(j as f64) * 10.0;
        let hi = -((j - 1) as f64) * 10.0;
        if dbmin >= lo && dbmin < hi {
            dbmin = lo;
            break;
        }
    }

    // Clamp to -100.
    if dbmin < -100.0 {
        dbmin = -100.0;
        for i in 0..length {
            if tgt.data[i] < -100.0 {
                tgt.data[i] = -100.0;
            }
        }
    }

    // Snap dbmax to interval boundaries: [0, 5), [5, 10).
    for j in (1..=2).rev() {
        let hi = (j as f64) * 5.0;
        let lo = ((j - 1) as f64) * 5.0;
        if dbmax >= lo && dbmax < hi {
            dbmax = hi;
            break;
        }
    }

    // Clamp to 10.
    if dbmax > 10.0 {
        dbmax = 10.0;
        for i in 0..length {
            if tgt.data[i] > 10.0 {
                tgt.data[i] = 10.0;
            }
        }
    }

    tgt.min = dbmin;
    tgt.max = dbmax;
}

fn to_percents(length: usize, component: &mut Component) {
    let mut pctmax = f64::NEG_INFINITY;

    let mut base = component.data[0];
    if base < f64::MIN_POSITIVE {
        base = component.max;
    }

    for i in 0..length {
        let pct = 100.0 * component.data[i] / base;
        pctmax = pctmax.max(pct);
        component.data[i] = pct;
    }

    // Snap pctmax to interval boundaries: [100, 110), [110, 120), ...
    for j in 0..10 {
        let lo = 100.0 + (j as f64) * 10.0;
        let hi = 100.0 + ((j + 1) as f64) * 10.0;
        if pctmax >= lo && pctmax < hi {
            pctmax = hi;
            break;
        }
    }

    // Clamp to 200.
    if pctmax > 200.0 {
        pctmax = 200.0;
        for i in 0..length {
            if component.data[i] > 200.0 {
                component.data[i] = 200.0;
            }
        }
    }

    component.min = 0.0;
    component.max = pctmax;
}

// ---------------------------------------------------------------------------
// Direct Real Fast Fourier Transform
// ---------------------------------------------------------------------------

/// Direct real FFT. Input is real data; output is {re, im} pairs.
/// Length must be a power of 2.
fn direct_real_fast_fourier_transform(array: &mut [f64]) {
    let two_pi = 2.0 * PI;

    let length = array.len();
    let ttheta = two_pi / length as f64;
    let nn = length / 2;
    let mut j: usize = 1;

    for ii in 1..=nn {
        let i = 2 * ii - 1;

        if j > i {
            let temp_r = array[j - 1];
            let temp_i = array[j];
            array[j - 1] = array[i - 1];
            array[j] = array[i];
            array[i - 1] = temp_r;
            array[i] = temp_i;
        }

        let mut m = nn;
        while m >= 2 && j > m {
            j -= m;
            m /= 2;
        }
        j += m;
    }

    let mut m_max = 2;
    let n = length;

    while n > m_max {
        let istep = 2 * m_max;
        let theta = two_pi / m_max as f64;
        let mut wp_r = (0.5 * theta).sin();
        wp_r = -2.0 * wp_r * wp_r;
        let wp_i = theta.sin();
        let mut w_r = 1.0;
        let mut w_i = 0.0;

        for ii in 1..=(m_max / 2) {
            let m = 2 * ii - 1;
            let mut jj = 0;
            while jj <= (n - m) / istep {
                let i = m + jj * istep;
                let j_idx = i + m_max;
                let temp_r = w_r * array[j_idx - 1] - w_i * array[j_idx];
                let temp_i = w_r * array[j_idx] + w_i * array[j_idx - 1];
                array[j_idx - 1] = array[i - 1] - temp_r;
                array[j_idx] = array[i] - temp_i;
                array[i - 1] += temp_r;
                array[i] += temp_i;
                jj += 1;
            }

            let w_temp = w_r;
            w_r = w_r * wp_r - w_i * wp_i + w_r;
            w_i = w_i * wp_r + w_temp * wp_i + w_i;
        }

        m_max = istep;
    }

    let mut twp_r = (0.5 * ttheta).sin();
    twp_r = -2.0 * twp_r * twp_r;
    let twp_i = ttheta.sin();
    let mut tw_r = 1.0 + twp_r;
    let mut tw_i = twp_i;
    let n_quarter = length / 4 + 1;

    for i in 2..=n_quarter {
        let i1 = i + i - 2;
        let i2 = i1 + 1;
        let i3 = length + 1 - i2;
        let i4 = i3 + 1;
        let w_rs = tw_r;
        let w_is = tw_i;
        let h1r = 0.5 * (array[i1] + array[i3]);
        let h1i = 0.5 * (array[i2] - array[i4]);
        let h2r = 0.5 * (array[i2] + array[i4]);
        let h2i = -0.5 * (array[i1] - array[i3]);
        array[i1] = h1r + w_rs * h2r - w_is * h2i;
        array[i2] = h1i + w_rs * h2i + w_is * h2r;
        array[i3] = h1r - w_rs * h2r + w_is * h2i;
        array[i4] = -h1i + w_rs * h2i + w_is * h2r;
        let tw_temp = tw_r;
        tw_r = tw_r * twp_r - tw_i * twp_i + tw_r;
        tw_i = tw_i * twp_r + tw_temp * twp_i + tw_i;
    }

    let tw_r_save = array[0];
    array[0] = tw_r_save + array[1];
    array[1] = tw_r_save - array[1];
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, epsilon: f64) -> bool {
        (a - b).abs() < epsilon
    }

    #[test]
    fn test_is_valid_signal_length() {
        let valid = [4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192];
        for i in 0..8199 {
            let expected = valid.contains(&i);
            let actual = is_valid_signal_length(i);
            assert_eq!(expected, actual, "is_valid_signal_length({})", i);
        }
    }

    #[test]
    fn test_prepare_frequency_domain() {
        let l = 7.0_f64;
        let expected = [1.0/l, 2.0/l, 3.0/l, 4.0/l, 5.0/l, 6.0/l, 7.0/l];
        let mut actual = [0.0; 7];
        prepare_frequency_domain(7, &mut actual);
        for i in 0..7 {
            assert!(
                almost_equal(expected[i], actual[i], f64::MIN_POSITIVE),
                "[{}] expected {}, got {}", i, expected[i], actual[i]
            );
        }
    }

    struct IdentityFilter;
    impl Updater for IdentityFilter {
        fn metadata(&self) -> Metadata {
            Metadata {
                identifier: crate::indicators::core::identifier::Identifier::SimpleMovingAverage,
                mnemonic: "identity".to_string(),
                description: "identity filter".to_string(),
                outputs: vec![],
            }
        }
        fn update(&mut self, sample: f64) -> f64 {
            sample
        }
    }

    #[test]
    fn test_prepare_filtered_signal() {
        let expected = [1000.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0];
        let mut filter = IdentityFilter;
        let actual = prepare_filtered_signal(7, &mut filter, 5);
        for i in 0..7 {
            assert!(
                almost_equal(expected[i], actual[i], f64::MIN_POSITIVE),
                "[{}] expected {}, got {}", i, expected[i], actual[i]
            );
        }
    }

    #[test]
    fn test_direct_real_fft() {
        let expected = [16.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
                        0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0];
        let mut actual = [1.0; 16];
        direct_real_fast_fourier_transform(&mut actual);
        for i in 0..16 {
            assert!(
                almost_equal(expected[i], actual[i], f64::MIN_POSITIVE),
                "[{}] expected {}, got {}", i, expected[i], actual[i]
            );
        }
    }

    #[test]
    fn test_calculate_identity() {
        let mut filter = IdentityFilter;
        let fr = calculate(512, &mut filter, 128, 179.0).unwrap();
        assert_eq!(fr.label, "identity");
        assert_eq!(fr.normalized_frequency.len(), 255);
        assert_eq!(fr.power_percent.data.len(), 255);
        assert_eq!(fr.power_decibel.data.len(), 255);
        assert_eq!(fr.amplitude_percent.data.len(), 255);
        assert_eq!(fr.amplitude_decibel.data.len(), 255);
        assert_eq!(fr.phase_degrees.data.len(), 255);
        assert_eq!(fr.phase_degrees_unwrapped.data.len(), 255);
    }

    #[test]
    fn test_calculate_invalid_signal_length() {
        let mut filter = IdentityFilter;
        assert!(calculate(3, &mut filter, 0, 179.0).is_err());
        assert!(calculate(5, &mut filter, 0, 179.0).is_err());
        assert!(calculate(6, &mut filter, 0, 179.0).is_err());
    }
}
