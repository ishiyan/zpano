/// Corona is the shared spectral-analysis engine used by CoronaSpectrum,
/// CoronaSignalToNoiseRatio, CoronaSwingPosition, and CoronaTrendVigor.

const DEFAULT_HIGH_PASS_FILTER_CUTOFF: i32 = 30;
const DEFAULT_MINIMAL_PERIOD: i32 = 6;
const DEFAULT_MAXIMAL_PERIOD: i32 = 30;
const DEFAULT_DECIBELS_LOWER_THRESHOLD: f64 = 6.0;
const DEFAULT_DECIBELS_UPPER_THRESHOLD: f64 = 20.0;

const HIGH_PASS_FILTER_BUFFER_SIZE: usize = 6;
const FIR_COEF_SUM: f64 = 12.0;

const DELTA_LOWER_THRESHOLD: f64 = 0.1;
const DELTA_FACTOR: f64 = -0.015;
const DELTA_SUMMAND: f64 = 0.5;

const DOMINANT_CYCLE_BUFFER_SIZE: usize = 5;
const DOMINANT_CYCLE_MEDIAN_INDEX: usize = 2;

const DECIBELS_SMOOTHING_ALPHA: f64 = 0.33;
const DECIBELS_SMOOTHING_ONE_MINUS: f64 = 0.67;

const NORMALIZED_AMPLITUDE_FACTOR: f64 = 0.99;
const DECIBELS_FLOOR: f64 = 0.01;
const DECIBELS_GAIN: f64 = 10.0;

/// Parameters for the Corona spectral analysis engine.
pub struct CoronaParams {
    pub high_pass_filter_cutoff: i32,
    pub minimal_period: i32,
    pub maximal_period: i32,
    pub decibels_lower_threshold: f64,
    pub decibels_upper_threshold: f64,
}

impl Default for CoronaParams {
    fn default() -> Self {
        Self {
            high_pass_filter_cutoff: DEFAULT_HIGH_PASS_FILTER_CUTOFF,
            minimal_period: DEFAULT_MINIMAL_PERIOD,
            maximal_period: DEFAULT_MAXIMAL_PERIOD,
            decibels_lower_threshold: DEFAULT_DECIBELS_LOWER_THRESHOLD,
            decibels_upper_threshold: DEFAULT_DECIBELS_UPPER_THRESHOLD,
        }
    }
}

/// Per-bin state of a single bandpass filter in the bank.
#[derive(Debug, Clone)]
pub struct Filter {
    pub in_phase: f64,
    pub in_phase_previous: f64,
    pub quadrature: f64,
    pub quadrature_previous: f64,
    pub real: f64,
    pub real_previous: f64,
    pub imaginary: f64,
    pub imaginary_previous: f64,
    pub amplitude_squared: f64,
    pub decibels: f64,
}

impl Filter {
    fn new() -> Self {
        Self {
            in_phase: 0.0,
            in_phase_previous: 0.0,
            quadrature: 0.0,
            quadrature_previous: 0.0,
            real: 0.0,
            real_previous: 0.0,
            imaginary: 0.0,
            imaginary_previous: 0.0,
            amplitude_squared: 0.0,
            decibels: 0.0,
        }
    }
}

/// Corona spectral analysis engine.
pub struct Corona {
    minimal_period: i32,
    maximal_period: i32,
    minimal_period_times_two: i32,
    maximal_period_times_two: i32,
    filter_bank_length: usize,
    decibels_lower_threshold: f64,
    decibels_upper_threshold: f64,

    alpha: f64,
    half_one_plus_alpha: f64,

    pre_calculated_beta: Vec<f64>,

    high_pass_buffer: [f64; HIGH_PASS_FILTER_BUFFER_SIZE],

    sample_previous: f64,
    smooth_hp_previous: f64,

    filter_bank: Vec<Filter>,

    maximal_amplitude_squared: f64,

    dominant_cycle_buffer: [f64; DOMINANT_CYCLE_BUFFER_SIZE],

    sample_count: i32,

    dominant_cycle: f64,
    dominant_cycle_median: f64,

    primed: bool,
}

impl Corona {
    /// Creates a new Corona engine using the provided parameters.
    pub fn new(p: &CoronaParams) -> Result<Self, String> {
        let mut cfg = CoronaParams {
            high_pass_filter_cutoff: p.high_pass_filter_cutoff,
            minimal_period: p.minimal_period,
            maximal_period: p.maximal_period,
            decibels_lower_threshold: p.decibels_lower_threshold,
            decibels_upper_threshold: p.decibels_upper_threshold,
        };

        apply_defaults(&mut cfg);
        verify_parameters(&cfg)?;

        let minimal_period_times_two = cfg.minimal_period * 2;
        let maximal_period_times_two = cfg.maximal_period * 2;
        let filter_bank_length = (maximal_period_times_two - minimal_period_times_two + 1) as usize;

        let mut dominant_cycle_buffer = [0.0_f64; DOMINANT_CYCLE_BUFFER_SIZE];
        for v in dominant_cycle_buffer.iter_mut() {
            *v = f64::MAX;
        }

        let phi = 2.0 * std::f64::consts::PI / cfg.high_pass_filter_cutoff as f64;
        let alpha = (1.0 - phi.sin()) / phi.cos();
        let half_one_plus_alpha = 0.5 * (1.0 + alpha);

        let mut pre_calculated_beta = vec![0.0; filter_bank_length];
        for index in 0..filter_bank_length {
            let n = minimal_period_times_two as usize + index;
            pre_calculated_beta[index] = (4.0 * std::f64::consts::PI / n as f64).cos();
        }

        let filter_bank: Vec<Filter> = (0..filter_bank_length).map(|_| Filter::new()).collect();

        Ok(Self {
            minimal_period: cfg.minimal_period,
            maximal_period: cfg.maximal_period,
            minimal_period_times_two,
            maximal_period_times_two,
            filter_bank_length,
            decibels_lower_threshold: cfg.decibels_lower_threshold,
            decibels_upper_threshold: cfg.decibels_upper_threshold,
            alpha,
            half_one_plus_alpha,
            pre_calculated_beta,
            high_pass_buffer: [0.0; HIGH_PASS_FILTER_BUFFER_SIZE],
            sample_previous: 0.0,
            smooth_hp_previous: 0.0,
            filter_bank,
            maximal_amplitude_squared: 0.0,
            dominant_cycle_buffer,
            sample_count: 0,
            dominant_cycle: f64::MAX,
            dominant_cycle_median: f64::MAX,
            primed: false,
        })
    }

    pub fn minimal_period(&self) -> i32 { self.minimal_period }
    pub fn maximal_period(&self) -> i32 { self.maximal_period }
    pub fn minimal_period_times_two(&self) -> i32 { self.minimal_period_times_two }
    pub fn maximal_period_times_two(&self) -> i32 { self.maximal_period_times_two }
    pub fn filter_bank_length(&self) -> usize { self.filter_bank_length }
    pub fn filter_bank(&self) -> &[Filter] { &self.filter_bank }
    pub fn is_primed(&self) -> bool { self.primed }
    pub fn dominant_cycle(&self) -> f64 { self.dominant_cycle }
    pub fn dominant_cycle_median(&self) -> f64 { self.dominant_cycle_median }
    pub fn maximal_amplitude_squared(&self) -> f64 { self.maximal_amplitude_squared }

    /// Feeds the next sample to the engine. Returns true once primed.
    pub fn update(&mut self, sample: f64) -> bool {
        if sample.is_nan() {
            return self.primed;
        }

        self.sample_count += 1;

        if self.sample_count == 1 {
            self.sample_previous = sample;
            return false;
        }

        // Step 1: High-pass filter.
        let hp = self.alpha * self.high_pass_buffer[HIGH_PASS_FILTER_BUFFER_SIZE - 1]
            + self.half_one_plus_alpha * (sample - self.sample_previous);
        self.sample_previous = sample;

        for i in 0..HIGH_PASS_FILTER_BUFFER_SIZE - 1 {
            self.high_pass_buffer[i] = self.high_pass_buffer[i + 1];
        }
        self.high_pass_buffer[HIGH_PASS_FILTER_BUFFER_SIZE - 1] = hp;

        // Step 2: 6-tap FIR smoothing.
        let smooth_hp = (self.high_pass_buffer[0]
            + 2.0 * self.high_pass_buffer[1]
            + 3.0 * self.high_pass_buffer[2]
            + 3.0 * self.high_pass_buffer[3]
            + 2.0 * self.high_pass_buffer[4]
            + self.high_pass_buffer[5])
            / FIR_COEF_SUM;

        // Step 3: Momentum.
        let momentum = smooth_hp - self.smooth_hp_previous;
        self.smooth_hp_previous = smooth_hp;

        // Step 4: Adaptive delta.
        let mut delta = DELTA_FACTOR * self.sample_count as f64 + DELTA_SUMMAND;
        if delta < DELTA_LOWER_THRESHOLD {
            delta = DELTA_LOWER_THRESHOLD;
        }

        // Step 5: Filter-bank update.
        self.maximal_amplitude_squared = 0.0;
        for index in 0..self.filter_bank_length {
            let n = self.minimal_period_times_two as usize + index;
            let nf = n as f64;

            let gamma = 1.0 / (8.0 * std::f64::consts::PI * delta / nf).cos();
            let a = gamma - (gamma * gamma - 1.0).sqrt();

            let quadrature = momentum * (nf / (4.0 * std::f64::consts::PI));
            let in_phase = smooth_hp;

            let half_one_min_a = 0.5 * (1.0 - a);
            let beta = self.pre_calculated_beta[index];
            let beta_one_plus_a = beta * (1.0 + a);

            let f = &self.filter_bank[index];
            let real = half_one_min_a * (in_phase - f.in_phase_previous)
                + beta_one_plus_a * f.real
                - a * f.real_previous;
            let imag = half_one_min_a * (quadrature - f.quadrature_previous)
                + beta_one_plus_a * f.imaginary
                - a * f.imaginary_previous;

            let amp_sq = real * real + imag * imag;

            let f = &mut self.filter_bank[index];
            f.in_phase_previous = f.in_phase;
            f.in_phase = in_phase;
            f.quadrature_previous = f.quadrature;
            f.quadrature = quadrature;
            f.real_previous = f.real;
            f.real = real;
            f.imaginary_previous = f.imaginary;
            f.imaginary = imag;
            f.amplitude_squared = amp_sq;

            if amp_sq > self.maximal_amplitude_squared {
                self.maximal_amplitude_squared = amp_sq;
            }
        }

        // Step 6: dB normalization and dominant-cycle weighted average.
        let mut numerator = 0.0_f64;
        let mut denominator = 0.0_f64;
        self.dominant_cycle = 0.0;

        for index in 0..self.filter_bank_length {
            let f = &mut self.filter_bank[index];
            let mut decibels = 0.0;

            if self.maximal_amplitude_squared > 0.0 {
                let normalized = f.amplitude_squared / self.maximal_amplitude_squared;
                if normalized > 0.0 {
                    let arg = (1.0 - NORMALIZED_AMPLITUDE_FACTOR * normalized) / DECIBELS_FLOOR;
                    if arg > 0.0 {
                        decibels = DECIBELS_GAIN * arg.log10();
                    }
                }
            }

            decibels = DECIBELS_SMOOTHING_ALPHA * decibels + DECIBELS_SMOOTHING_ONE_MINUS * f.decibels;
            if decibels > self.decibels_upper_threshold {
                decibels = self.decibels_upper_threshold;
            }
            f.decibels = decibels;

            if decibels <= self.decibels_lower_threshold {
                let n = (self.minimal_period_times_two as usize + index) as f64;
                let adjusted = self.decibels_upper_threshold - decibels;
                numerator += n * adjusted;
                denominator += adjusted;
            }
        }

        if denominator != 0.0 {
            self.dominant_cycle = 0.5 * numerator / denominator;
        }
        if self.dominant_cycle < self.minimal_period as f64 {
            self.dominant_cycle = self.minimal_period as f64;
        }

        // Step 7: 5-sample median.
        for i in 0..DOMINANT_CYCLE_BUFFER_SIZE - 1 {
            self.dominant_cycle_buffer[i] = self.dominant_cycle_buffer[i + 1];
        }
        self.dominant_cycle_buffer[DOMINANT_CYCLE_BUFFER_SIZE - 1] = self.dominant_cycle;

        let mut sorted = self.dominant_cycle_buffer;
        sorted.sort_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal));
        self.dominant_cycle_median = sorted[DOMINANT_CYCLE_MEDIAN_INDEX];
        if self.dominant_cycle_median < self.minimal_period as f64 {
            self.dominant_cycle_median = self.minimal_period as f64;
        }

        if self.sample_count < self.minimal_period_times_two {
            return false;
        }
        self.primed = true;

        true
    }
}

fn apply_defaults(p: &mut CoronaParams) {
    if p.high_pass_filter_cutoff <= 0 {
        p.high_pass_filter_cutoff = DEFAULT_HIGH_PASS_FILTER_CUTOFF;
    }
    if p.minimal_period <= 0 {
        p.minimal_period = DEFAULT_MINIMAL_PERIOD;
    }
    if p.maximal_period <= 0 {
        p.maximal_period = DEFAULT_MAXIMAL_PERIOD;
    }
    if p.decibels_lower_threshold == 0.0 {
        p.decibels_lower_threshold = DEFAULT_DECIBELS_LOWER_THRESHOLD;
    }
    if p.decibels_upper_threshold == 0.0 {
        p.decibels_upper_threshold = DEFAULT_DECIBELS_UPPER_THRESHOLD;
    }
}

fn verify_parameters(p: &CoronaParams) -> Result<(), String> {
    let invalid = "invalid corona parameters";
    if p.high_pass_filter_cutoff < 2 {
        return Err(format!("{}: HighPassFilterCutoff should be >= 2", invalid));
    }
    if p.minimal_period < 2 {
        return Err(format!("{}: MinimalPeriod should be >= 2", invalid));
    }
    if p.maximal_period <= p.minimal_period {
        return Err(format!("{}: MaximalPeriod should be > MinimalPeriod", invalid));
    }
    if p.decibels_lower_threshold < 0.0 {
        return Err(format!("{}: DecibelsLowerThreshold should be >= 0", invalid));
    }
    if p.decibels_upper_threshold <= p.decibels_lower_threshold {
        return Err(format!("{}: DecibelsUpperThreshold should be > DecibelsLowerThreshold", invalid));
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::testdata;

    #[test]
    fn test_corona_default_params() {
        let c = Corona::new(&CoronaParams::default()).unwrap();
        assert_eq!(c.minimal_period(), 6);
        assert_eq!(c.maximal_period(), 30);
        assert_eq!(c.filter_bank_length(), 49);
        assert!(!c.is_primed());
    }

    #[test]
    fn test_corona_priming() {
        let mut c = Corona::new(&CoronaParams::default()).unwrap();
        // MinimalPeriodTimesTwo = 12, primes at sample index 11 (1-indexed sample_count=12).
        for i in 0..11 {
            let primed = c.update(100.0 + i as f64);
            assert!(!primed, "should not be primed at sample {}", i);
        }
        let primed = c.update(111.0);
        assert!(primed, "should be primed at sample 11");
        assert!(c.is_primed());
    }

    #[test]
    fn test_corona_nan_passthrough() {
        let mut c = Corona::new(&CoronaParams::default()).unwrap();
        assert!(!c.update(f64::NAN));
        assert_eq!(c.sample_count, 0);
    }
    #[test]
    fn test_corona_default_smoke() {
        let mut c = Corona::new(&CoronaParams::default()).unwrap();

        assert_eq!(c.filter_bank_length(), 49);
        assert_eq!(c.minimal_period_times_two(), 12);
        assert_eq!(c.maximal_period_times_two(), 60);

        let input = testdata::talib_input();
        let mut primed_at: Option<usize> = None;
        for (i, &v) in input.iter().enumerate() {
            c.update(v);
            if c.is_primed() && primed_at.is_none() {
                primed_at = Some(i);
            }
        }
        let primed_at = primed_at.expect("engine never primed over 252 samples");
        assert_eq!(primed_at + 1, c.minimal_period_times_two() as usize,
            "primedAt (1-based) = {}, want {}", primed_at + 1, c.minimal_period_times_two());

        let dc = c.dominant_cycle();
        let dc_med = c.dominant_cycle_median();
        assert!(dc.is_finite(), "DominantCycle should be finite, got {}", dc);
        assert!(dc_med.is_finite(), "DominantCycleMedian should be finite, got {}", dc_med);

        let min = c.minimal_period() as f64;
        let max = c.maximal_period() as f64;
        assert!(dc >= min && dc <= max, "DominantCycle = {}, want in [{}, {}]", dc, min, max);
        assert!(dc_med >= min && dc_med <= max, "DominantCycleMedian = {}, want in [{}, {}]", dc_med, min, max);

        let m = c.maximal_amplitude_squared();
        assert!(m > 0.0 && m.is_finite(), "MaximalAmplitudeSquared = {}, want positive finite", m);

        // Second pass: spot-check and verify DC exceeds MinimalPeriod at some point.
        let mut c2 = Corona::new(&CoronaParams::default()).unwrap();
        let mut saw_above_min = false;
        for (i, &v) in input.iter().enumerate() {
            c2.update(v);
            if i == 11 || i == 30 || i == 60 || i == 100 || i == 150 || i == 200 || i == 251 {
                // Spot-check: just ensure values are finite at these indices.
                assert!(c2.dominant_cycle().is_finite(), "bar {}: DC not finite", i);
                assert!(c2.dominant_cycle_median().is_finite(), "bar {}: DCmed not finite", i);
            }
            if c2.is_primed() && c2.dominant_cycle() > min {
                saw_above_min = true;
            }
        }
        assert!(saw_above_min, "DominantCycle never exceeded MinimalPeriod across 252 samples");
    }

    #[test]
    fn test_corona_nan_after_primed() {
        let mut c = Corona::new(&CoronaParams::default()).unwrap();
        let input = testdata::talib_input();
        for &v in &input[..20] {
            c.update(v);
        }
        assert!(c.is_primed(), "expected primed after 20 samples");

        let dc_before = c.dominant_cycle();
        let dc_med_before = c.dominant_cycle_median();

        let got = c.update(f64::NAN);
        assert!(got, "Update(NaN) should return true (preserves primed)");
        assert_eq!(c.dominant_cycle(), dc_before, "NaN input mutated DominantCycle");
        assert_eq!(c.dominant_cycle_median(), dc_med_before, "NaN input mutated DominantCycleMedian");
    }

    #[test]
    fn test_corona_invalid_params() {
        assert!(Corona::new(&CoronaParams {
            high_pass_filter_cutoff: 1,
            ..CoronaParams::default()
        }).is_err());

        assert!(Corona::new(&CoronaParams {
            minimal_period: 1,
            ..CoronaParams::default()
        }).is_err());

        assert!(Corona::new(&CoronaParams {
            minimal_period: 6,
            maximal_period: 6,
            ..CoronaParams::default()
        }).is_err());

        // negative dB lower
        assert!(Corona::new(&CoronaParams {
            decibels_lower_threshold: -1.0,
            ..CoronaParams::default()
        }).is_err());

        // dB upper <= lower
        assert!(Corona::new(&CoronaParams {
            decibels_lower_threshold: 6.0,
            decibels_upper_threshold: 6.0,
            ..CoronaParams::default()
        }).is_err());
    }
}
