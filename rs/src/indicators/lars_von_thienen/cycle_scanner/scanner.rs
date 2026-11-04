use super::super::bartels_periodicity_score::scorer as bartelsperiodicityscore;
use super::super::bartels_periodicity_score::scorer::BartelsOutput;
use super::super::harmonic_cycle::model as harmoniccycle;
use super::super::harmonic_cycle::model::CycleModel;
use super::super::hodrick_prescott_smoothing_and_detrending::solver as hodrickprescottsmoothinganddetrending;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct ScanConfig {
    pub window: i64,
    pub min_period: i64,
    pub max_period: i64,
    pub smoothing: f64,
    pub genuine_threshold: f64,
    pub phase_cycles: f64,
    pub top_k: i64,
    pub peak_distance: i64,
    pub max_segments: i64,
}
impl ScanConfig {
    pub fn validate(&self) -> Option<String> {
        if (((((self.window < 6) || (self.min_period < 3)) || (self.max_period < self.min_period))
            || (self.top_k < 1))
            || (self.peak_distance < 1))
            || (self.max_segments < 2)
        {
            return Some("invalid scanner integer parameters".to_string());
        }
        for (name, value) in std::collections::HashMap::from([
            ("smoothing", self.smoothing),
            ("genuine_threshold", self.genuine_threshold),
            ("phase_cycles", self.phase_cycles),
        ]) {
            {
                let err = harmoniccycle::validate_real(value, name, 0.0);
                if err.is_some() {
                    return err;
                }
            }
        }
        if (self.genuine_threshold > 100.0)
            || ((self.phase_cycles != 0.0) && (self.phase_cycles < 1.0))
        {
            return Some("invalid scanner threshold or phase cycles".to_string());
        }
        return None;
    }
}
pub fn default_config() -> ScanConfig {
    return ScanConfig {
        window: 200,
        min_period: 5,
        max_period: 80,
        smoothing: 1600.0,
        genuine_threshold: 49.0,
        phase_cycles: 2.0,
        top_k: 3,
        peak_distance: 2,
        max_segments: 5,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct RankedCycle {
    pub period: f64,
    pub amplitude: f64,
    pub phase: f64,
    pub strength: f64,
    pub genuine_percent: f64,
    pub segments: f64,
    pub fit_amplitude: f64,
    pub fit_samples: f64,
    pub value: f64,
    pub slope: f64,
    pub eta_top: f64,
    pub eta_bottom: f64,
}
impl RankedCycle {}
#[derive(Clone, Default)]
pub struct ScanOutput {
    pub amplitude: Vec<f64>,
    pub genuine_percent: Vec<f64>,
    pub cycles: Vec<RankedCycle>,
    pub accepted_count: f64,
}
impl ScanOutput {}
pub fn empty_cycle() -> RankedCycle {
    let nan: f64 = f64::NAN;
    return RankedCycle {
        period: nan,
        amplitude: nan,
        phase: nan,
        strength: nan,
        genuine_percent: nan,
        segments: nan,
        fit_amplitude: nan,
        fit_samples: nan,
        value: nan,
        slope: nan,
        eta_top: nan,
        eta_bottom: nan,
    }
    .clone();
}
pub fn empty_scan(config: ScanConfig) -> ScanOutput {
    let count: i64 = (config.max_period - config.min_period) + 1;
    let mut output = ScanOutput {
        amplitude: vec![Default::default(); (count) as usize].clone(),
        genuine_percent: vec![Default::default(); (count) as usize].clone(),
        cycles: vec![Default::default(); (config.top_k) as usize].clone(),
        accepted_count: f64::NAN,
    }
    .clone();
    for i in 0..count {
        output.amplitude[(i) as usize] = f64::NAN;
        output.genuine_percent[(i) as usize] = f64::NAN;
    }
    for i in 0..(output.cycles.len() as i64) {
        output.cycles[(i) as usize] = empty_cycle().clone();
    }
    return output.clone();
}
pub fn modeled_value(cycle: RankedCycle, offset: f64) -> f64 {
    return cycle.fit_amplitude
        * (cycle.phase + (((2.0 * std::f64::consts::PI) * offset) / cycle.period)).sin();
}
pub fn modulo(mut value: f64) -> f64 {
    value = value % (2.0 * std::f64::consts::PI);
    if value < 0.0 {
        value += 2.0 * std::f64::consts::PI;
    }
    return value;
}
pub fn scan_cycles(prices: Vec<f64>, config: ScanConfig) -> ScanOutput {
    {
        let err = config.validate();
        if err.is_some() {
            panic!("{:?}", err);
        }
    }
    if (prices.len() as i64) < 6 {
        panic!("{:?}", "scan requires >= 6 prices");
    }
    for value in prices.clone() {
        harmoniccycle::validate_sample(value);
        if (value).is_nan() {
            panic!("{:?}", "scan batch data must be finite");
        }
    }
    let residual: Vec<f64> =
        hodrickprescottsmoothinganddetrending::cycle_data(prices.clone(), config.smoothing).clone();
    let mut output = empty_scan(config.clone()).clone();
    let count: i64 = output.amplitude.len() as i64;
    let mut models: Vec<CycleModel> = vec![Default::default(); (count) as usize].clone();
    let mut evidence: Vec<BartelsOutput> = vec![Default::default(); (count) as usize].clone();
    for i in 0..count {
        let period: i64 = config.min_period + i;
        models[(i) as usize] = harmoniccycle::goertzel(residual.clone(), period as f64).clone();
        let segments: i64 = ((prices.len() as i64) / period).min(config.max_segments);
        evidence[(i) as usize] = bartelsperiodicityscore::bartels_score(
            prices[((prices.len() as i64) - (segments * period)) as usize..]
                .to_vec()
                .clone(),
            period,
            config.max_segments,
            2,
        )
        .clone();
        output.amplitude[(i) as usize] = models[(i) as usize].amplitude;
        output.genuine_percent[(i) as usize] = evidence[(i) as usize].genuine_percent;
    }
    let mut candidates: Vec<i64> = Vec::<i64>::new().clone();
    for (__index, model) in models.clone().into_iter().enumerate() {
        let i = __index as i64;
        let (mut left, mut right) = (f64::NEG_INFINITY, f64::NEG_INFINITY);
        if i > 0 {
            left = output.amplitude[(i - 1) as usize];
        }
        if (i + 1) < count {
            right = output.amplitude[(i + 1) as usize];
        }
        let score = evidence[(i) as usize].clone();
        if ((((model.amplitude > 0.0) && (model.amplitude > left)) && (model.amplitude >= right))
            && !(score.genuine_percent).is_nan())
            && (score.genuine_percent > config.genuine_threshold)
        {
            candidates = {
                let mut __values = candidates.clone();
                __values.push(i);
                __values
            }
            .clone();
        }
    }
    candidates.sort_by(|&i, &j| {
        let less = |i: i64, j: i64| {
            let (left, right) = (models[(i) as usize].clone(), models[(j) as usize].clone());
            let (a, b) = (
                (left.amplitude / left.period),
                (right.amplitude / right.period),
            );
            if a == b {
                return left.period < right.period;
            }
            return a > b;
        };
        if less(i, j) {
            std::cmp::Ordering::Less
        } else if less(j, i) {
            std::cmp::Ordering::Greater
        } else {
            std::cmp::Ordering::Equal
        }
    });
    let mut accepted: Vec<RankedCycle> = Vec::<RankedCycle>::new().clone();
    for i in candidates.clone() {
        let (model, score) = (models[(i) as usize].clone(), evidence[(i) as usize].clone());
        let mut suppressed = false;
        for prior in accepted.clone() {
            if (model.period - prior.period).abs() < (config.peak_distance as f64) {
                suppressed = true;
                break;
            }
        }
        if suppressed {
            continue;
        }
        let mut fit_n: i64 = prices.len() as i64;
        if config.phase_cycles != 0.0 {
            fit_n = (fit_n).min((6).max(((config.phase_cycles * model.period).ceil()) as i64));
        }
        let mut fit = model.clone();
        if fit_n != (prices.len() as i64) {
            fit = harmoniccycle::goertzel(
                hodrickprescottsmoothinganddetrending::cycle_data(
                    prices[((prices.len() as i64) - fit_n) as usize..]
                        .to_vec()
                        .clone(),
                    config.smoothing,
                )
                .clone(),
                model.period,
            )
            .clone();
        }
        let phase: f64 = modulo(
            fit.phase + (((2.0 * std::f64::consts::PI) * ((fit_n - 1) as f64)) / model.period),
        );
        let omega: f64 = (2.0 * std::f64::consts::PI) / model.period;
        let value: f64 = fit.amplitude * (phase).sin();
        let slope: f64 = (fit.amplitude * omega) * (phase).cos();
        let eta_top: f64 = modulo((std::f64::consts::PI / 2.0) - phase) / omega;
        let eta_bottom: f64 = modulo(((3.0 * std::f64::consts::PI) / 2.0) - phase) / omega;
        accepted = {
            let mut __values = accepted.clone();
            __values.push(
                RankedCycle {
                    period: model.period,
                    amplitude: model.amplitude,
                    phase: phase,
                    strength: (model.amplitude / model.period),
                    genuine_percent: score.genuine_percent,
                    segments: score.segments,
                    fit_amplitude: fit.amplitude,
                    fit_samples: (fit_n as f64),
                    value: value,
                    slope: slope,
                    eta_top: eta_top,
                    eta_bottom: eta_bottom,
                }
                .clone(),
            );
            __values
        }
        .clone();
    }
    {
        let __copy = accepted.clone();
        let count = output.cycles.len().min(__copy.len());
        output.cycles[..count].clone_from_slice(&__copy[..count]);
    };
    output.accepted_count = (accepted.len() as i64) as f64;
    return output.clone();
}
#[derive(Clone, Default)]
pub struct RollingScanner {
    pub config: ScanConfig,
    pub prices: Vec<f64>,
    pub samples_seen: i64,
}
impl RollingScanner {
    pub fn reset(&mut self) {
        self.prices = self.prices[0..(0) as usize].to_vec().clone();
        self.samples_seen = 0;
    }
    pub fn is_primed(&self) -> bool {
        return (self.prices.len() as i64) == self.config.window;
    }
    pub fn update(&mut self, price: f64) -> ScanOutput {
        harmoniccycle::validate_sample(price);
        if (price).is_nan() {
            return empty_scan(self.config.clone()).clone();
        }
        let prices: Vec<f64> = harmoniccycle::push(
            {
                let mut __values = Vec::<f64>::new().clone();
                __values.extend(self.prices.clone());
                __values
            }
            .clone(),
            price,
            self.config.window,
        )
        .clone();
        let mut output = empty_scan(self.config.clone()).clone();
        if (prices.len() as i64) == self.config.window {
            output = scan_cycles(prices.clone(), self.config.clone()).clone();
        }
        self.prices = prices.clone();
        self.samples_seen += 1;
        return output.clone();
    }
}
pub fn new_rolling_scanner(config: ScanConfig) -> (RollingScanner, Option<String>) {
    {
        let err = config.validate();
        if err.is_some() {
            panic!("{:?}", err);
        }
    }
    return (
        RollingScanner {
            config: config.clone(),
            prices: Vec::<f64>::new().clone(),
            ..Default::default()
        }
        .clone(),
        None,
    );
}
