use super::super::cycle_scanner::scanner as cyclescanner;
use super::super::harmonic_cycle::model as harmoniccycle;
use super::CompleteCycleScannerParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub spectrum: Vec<f64>,
    pub genuine_percent: Vec<f64>,
    pub dominant_period: f64,
    pub accepted_count: f64,
    pub cycles: Vec<cyclescanner::RankedCycle>,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        spectrum: Vec::new(),
        genuine_percent: Vec::new(),
        dominant_period: nan,
        accepted_count: nan,
        cycles: Vec::new(),
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: CompleteCycleScannerParams,
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
        return Snapshot {
            spectrum: scan.amplitude.clone(),
            genuine_percent: scan.genuine_percent.clone(),
            dominant_period: scan.cycles[(0) as usize].period,
            accepted_count: scan.accepted_count,
            cycles: scan.cycles.clone(),
        }
        .clone();
    }
}
