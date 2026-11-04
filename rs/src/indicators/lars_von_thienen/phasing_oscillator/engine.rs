use super::super::cycle_scanner::scanner as cyclescanner;
use super::super::harmonic_cycle::model as harmoniccycle;
use super::PhasingOscillatorParams;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

#[derive(Clone, Default)]
pub struct Snapshot {
    pub value: f64,
    pub short_score: f64,
    pub long_score: f64,
    pub short_period: f64,
    pub long_period: f64,
}
impl Snapshot {}
pub fn empty_snapshot() -> Snapshot {
    let nan: f64 = f64::NAN;
    return Snapshot {
        value: nan,
        short_score: nan,
        long_score: nan,
        short_period: nan,
        long_period: nan,
    }
    .clone();
}
#[derive(Clone, Default)]
pub struct Engine {
    pub p: PhasingOscillatorParams,
    pub short: cyclescanner::RollingScanner,
    pub long: cyclescanner::RollingScanner,
}
impl Engine {
    pub fn reset(&mut self) {
        (self.short, _) = cyclescanner::new_rolling_scanner(
            cyclescanner::ScanConfig {
                window: (self.p.window as i64),
                min_period: (self.p.short_min as i64),
                max_period: (self.p.short_max as i64),
                smoothing: self.p.smoothing,
                genuine_threshold: self.p.genuine_threshold,
                phase_cycles: self.p.phase_cycles,
                top_k: 1,
                peak_distance: 2,
                max_segments: 5,
            }
            .clone(),
        );
        (self.long, _) = cyclescanner::new_rolling_scanner(
            cyclescanner::ScanConfig {
                window: (self.p.window as i64),
                min_period: (self.p.long_min as i64),
                max_period: (self.p.long_max as i64),
                smoothing: self.p.smoothing,
                genuine_threshold: self.p.genuine_threshold,
                phase_cycles: self.p.phase_cycles,
                top_k: 1,
                peak_distance: 2,
                max_segments: 5,
            }
            .clone(),
        );
    }
    pub fn primed(&self) -> bool {
        return self.short.is_primed() && self.long.is_primed();
    }
    pub fn warmup(&self) -> i64 {
        return (self.p.window as i64) - 1;
    }
    pub fn update(&mut self, price: f64) -> Snapshot {
        harmoniccycle::validate_sample(price);
        if (price).is_nan() {
            return empty_snapshot().clone();
        }
        let short = self.short.update(price).cycles[(0) as usize].clone();
        let long = self.long.update(price).cycles[(0) as usize].clone();
        let (short_score, long_score) = (
            harmoniccycle::phase_state(short.phase),
            harmoniccycle::phase_state(long.phase),
        );
        return Snapshot {
            value: (short_score + long_score),
            short_score: short_score,
            long_score: long_score,
            short_period: short.period,
            long_period: long.period,
        }
        .clone();
    }
}
