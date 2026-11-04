use super::super::cycle_scanner::scanner as cyclescanner;
use super::super::harmonic_cycle::model as harmoniccycle;
use super::DominantMarketVibrationParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
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
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
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
#[derive(Clone, Default)]
pub struct Engine {
    pub p: DominantMarketVibrationParams,
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
        return Snapshot {
            period: cycle.period,
            amplitude: cycle.amplitude,
            phase: cycle.phase,
            strength: cycle.strength,
            genuine_percent: cycle.genuine_percent,
            segments: cycle.segments,
            fit_amplitude: cycle.fit_amplitude,
            fit_samples: cycle.fit_samples,
            value: cycle.value,
            slope: cycle.slope,
            eta_top: cycle.eta_top,
            eta_bottom: cycle.eta_bottom,
        }
        .clone();
    }
}
