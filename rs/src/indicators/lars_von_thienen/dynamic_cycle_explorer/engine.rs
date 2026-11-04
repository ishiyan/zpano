use super::super::cycle_scanner::scanner as cyclescanner;
use super::super::harmonic_cycle::model as harmoniccycle;
use super::DynamicCycleExplorerParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub value: f64,
    pub period: f64,
    pub phase: f64,
    pub eta_top: f64,
    pub eta_bottom: f64,
    pub highlighter: f64,
    pub highlighter_index: f64,
    pub fitted_at_highlighter: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        value: nan,
        period: nan,
        phase: nan,
        eta_top: nan,
        eta_bottom: nan,
        highlighter: nan,
        highlighter_index: nan,
        fitted_at_highlighter: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: DynamicCycleExplorerParams,
    pub scanner: cyclescanner::RollingScanner,
}
impl Engine {
    pub fn reset(&mut self) {
        (self.scanner, _) = cyclescanner::new_rolling_scanner(
            cyclescanner::ScanConfig {
                window: (self.p.window as i64),
                min_period: (self.p.min_period as i64),
                max_period: (self.p.max_period as i64),
                smoothing: self.p.smoothing,
                genuine_threshold: self.p.genuine_threshold,
                phase_cycles: self.p.phase_cycles,
                top_k: (self.p.top_k as i64),
                peak_distance: (self.p.peak_distance as i64),
                max_segments: (self.p.max_segments as i64),
            }
            .clone(),
        );
    }
    pub fn primed(&self) -> bool {
        return self.scanner.is_primed();
    }
    pub fn warmup(&self) -> i64 {
        return (self.p.window as i64) - 1;
    }
    pub fn update(&mut self, price: f64) -> Snapshot {
        harmoniccycle::validate_sample(price);
        if (price).is_nan() {
            return empty_snapshot().clone();
        }
        let scan = self.scanner.update(price).clone();
        let cycle = scan.cycles[(0) as usize].clone();
        if (cycle.period).is_nan() {
            return empty_snapshot().clone();
        }
        let long_length: i64 = (cycle.period) as i64;
        let short_length: i64 = (2).max((((cycle.period / 2.0) + 0.5).floor()) as i64);
        let delay: i64 = long_length / 2;
        let center: i64 = ((self.scanner.prices.len() as i64) - 1) - delay;
        let highlighter: f64 = centered_mean(self.scanner.prices.clone(), center, short_length)
            - centered_mean(self.scanner.prices.clone(), center, long_length);
        let (mut index, mut fitted) = (f64::NAN, f64::NAN);
        if !(highlighter).is_nan() && !(highlighter).is_infinite() {
            index = ((self.scanner.samples_seen - 1) - delay) as f64;
            fitted = cyclescanner::modeled_value(cycle.clone(), -(delay as f64));
        }
        return Snapshot {
            value: cycle.value,
            period: cycle.period,
            phase: cycle.phase,
            eta_top: cycle.eta_top,
            eta_bottom: cycle.eta_bottom,
            highlighter: highlighter,
            highlighter_index: index,
            fitted_at_highlighter: fitted,
        }
        .clone();
    }
}
pub fn centered_mean(values: Vec<f64>, center: i64, length: i64) -> f64 {
    let half: i64 = length / 2;
    if ((center - half) < 0) || ((center + half) >= (values.len() as i64)) {
        return f64::NAN;
    }
    let mut sum: f64 = Default::default();

    if (length % 2) != 0 {
        for value in values[(center - half) as usize..((center + half) + 1) as usize]
            .to_vec()
            .clone()
        {
            sum += value;
        }
        return sum / (length as f64);
    }
    for value in values[((center - half) + 1) as usize..(center + half) as usize]
        .to_vec()
        .clone()
    {
        sum += value;
    }
    return (((0.5 * values[(center - half) as usize]) + sum)
        + (0.5 * values[(center + half) as usize]))
        / (length as f64);
}
