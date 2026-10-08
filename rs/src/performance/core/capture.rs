//! Streaming up/down capture, number and percentage ratios.

use crate::streaming_kbn::KleinKbnAccumulator;

/// Streaming capture ratios of an asset relative to a benchmark,
/// following the PerformanceAnalytics conventions: upside periods have
/// benchmark > 0; downside capture and down-number use benchmark ≤ 0;
/// down-percentage uses benchmark < 0.
///
/// Only sums and counts are stored, so `revert()` may remove any
/// previously added pair (FIFO rolling windows).
#[derive(Debug, Clone, Default)]
pub struct Capture {
    logret_a_sum_up: KleinKbnAccumulator,
    logret_b_sum_up: KleinKbnAccumulator,
    logret_a_sum_dn: KleinKbnAccumulator,
    logret_b_sum_dn: KleinKbnAccumulator,
    a_sum_up: KleinKbnAccumulator,
    b_sum_up: KleinKbnAccumulator,
    a_sum_dn: KleinKbnAccumulator,
    b_sum_dn: KleinKbnAccumulator,
    a_num_up: i64,
    b_num_up: i64,
    a_num_dn: i64,
    b_num_dn: i64,
    a_perc_up: i64,
    b_perc_up: i64,
    a_perc_dn: i64,
    b_perc_dn: i64,
}

#[inline]
fn logret(ret: f64) -> f64 {
    if ret != 0.0 { ret.ln_1p() } else { 0.0 }
}

#[inline]
fn nan_div(a: f64, b: f64) -> f64 {
    if b != 0.0 { a / b } else { f64::NAN }
}

impl Capture {
    /// Creates an empty accumulator.
    pub fn new() -> Self {
        Self::default()
    }

    /// Resets to the initial empty state.
    pub fn reset(&mut self) {
        *self = Self::default();
    }

    /// Removes a previously added (asset, benchmark) return pair.
    pub fn revert(&mut self, ret_asset: f64, ret_benchmark: f64) {
        if ret_benchmark > 0.0 {
            // Upside
            // Geometric
            self.logret_a_sum_up.revert(logret(ret_asset));
            self.logret_b_sum_up.revert(logret(ret_benchmark));
            // Arithmetic
            self.a_sum_up.revert(ret_asset);
            self.b_sum_up.revert(ret_benchmark);
            // Number
            self.b_num_up -= 1;
            if ret_asset > 0.0 {
                self.a_num_up -= 1;
            }
            // Percentage
            self.b_perc_up -= 1;
            if ret_asset > ret_benchmark {
                self.a_perc_up -= 1;
            }
        } else {
            // Downside
            // Geometric
            self.logret_a_sum_dn.revert(logret(ret_asset));
            self.logret_b_sum_dn.revert(logret(ret_benchmark));
            // Arithmetic
            self.a_sum_dn.revert(ret_asset);
            self.b_sum_dn.revert(ret_benchmark);
            // Number
            self.b_num_dn -= 1;
            if ret_asset < 0.0 {
                self.a_num_dn -= 1;
            }
            // Percentage
            if ret_benchmark < 0.0 {
                self.b_perc_dn -= 1;
                if ret_asset > ret_benchmark {
                    self.a_perc_dn -= 1;
                }
            }
        }
    }

    /// Adds an (asset, benchmark) return pair.
    pub fn update(&mut self, ret_asset: f64, ret_benchmark: f64) {
        if ret_benchmark > 0.0 {
            // Upside
            // Geometric
            self.logret_a_sum_up.update(logret(ret_asset));
            self.logret_b_sum_up.update(logret(ret_benchmark));
            // Arithmetic
            self.a_sum_up.update(ret_asset);
            self.b_sum_up.update(ret_benchmark);
            // Counts
            self.b_num_up += 1;
            if ret_asset > 0.0 {
                self.a_num_up += 1;
            }
            // Perc
            self.b_perc_up += 1;
            if ret_asset > ret_benchmark {
                self.a_perc_up += 1;
            }
        } else {
            // Downside
            // Geometric
            self.logret_a_sum_dn.update(logret(ret_asset));
            self.logret_b_sum_dn.update(logret(ret_benchmark));
            // Arithmetic
            self.a_sum_dn.update(ret_asset);
            self.b_sum_dn.update(ret_benchmark);
            // Counts
            self.b_num_dn += 1;
            if ret_asset < 0.0 {
                self.a_num_dn += 1;
            }
            // Perc
            if ret_benchmark < 0.0 {
                self.b_perc_dn += 1;
                if ret_asset > ret_benchmark {
                    self.a_perc_dn += 1;
                }
            }
        }
    }

    /// Geometric upside capture ratio: compounded asset return over
    /// compounded benchmark return in upside periods (NaN if the latter is 0).
    pub fn upside_capture_ratio_geometric(&self) -> f64 {
        let a_cum = self.logret_a_sum_up.value().exp_m1();
        let b_cum = self.logret_b_sum_up.value().exp_m1();
        nan_div(a_cum, b_cum)
    }

    /// Arithmetic upside capture ratio: sum of asset returns over sum of
    /// benchmark returns in upside periods (NaN if the latter is 0).
    pub fn upside_capture_ratio_arithmetic(&self) -> f64 {
        nan_div(self.a_sum_up.value(), self.b_sum_up.value())
    }

    /// Geometric downside capture ratio (benchmark ≤ 0 periods).
    pub fn downside_capture_ratio_geometric(&self) -> f64 {
        let a_cum = self.logret_a_sum_dn.value().exp_m1();
        let b_cum = self.logret_b_sum_dn.value().exp_m1();
        nan_div(a_cum, b_cum)
    }

    /// Arithmetic downside capture ratio (benchmark ≤ 0 periods).
    pub fn downside_capture_ratio_arithmetic(&self) -> f64 {
        nan_div(self.a_sum_dn.value(), self.b_sum_dn.value())
    }

    /// Fraction of upside periods in which the asset return is positive.
    pub fn up_number_ratio(&self) -> f64 {
        nan_div(self.a_num_up as f64, self.b_num_up as f64)
    }

    /// Fraction of downside (benchmark ≤ 0) periods in which the asset
    /// return is negative.
    pub fn down_number_ratio(&self) -> f64 {
        nan_div(self.a_num_dn as f64, self.b_num_dn as f64)
    }

    /// Fraction of upside periods in which the asset outperforms the benchmark.
    pub fn up_percentage_ratio(&self) -> f64 {
        nan_div(self.a_perc_up as f64, self.b_perc_up as f64)
    }

    /// Fraction of strictly negative benchmark periods in which the asset
    /// outperforms the benchmark.
    pub fn down_percentage_ratio(&self) -> f64 {
        nan_div(self.a_perc_dn as f64, self.b_perc_dn as f64)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::performance::core::test_support::{almost_equal, assert_close, choice2, gauss, prod, Rng};

    type Getter = fn(&Capture) -> f64;

    const NAMES: [(&str, Getter); 8] = [
        ("upside_capture_ratio_geometric", Capture::upside_capture_ratio_geometric),
        ("upside_capture_ratio_arithmetic", Capture::upside_capture_ratio_arithmetic),
        ("downside_capture_ratio_geometric", Capture::downside_capture_ratio_geometric),
        ("downside_capture_ratio_arithmetic", Capture::downside_capture_ratio_arithmetic),
        ("up_number_ratio", Capture::up_number_ratio),
        ("down_number_ratio", Capture::down_number_ratio),
        ("up_percentage_ratio", Capture::up_percentage_ratio),
        ("down_percentage_ratio", Capture::down_percentage_ratio),
    ];

    /// Naive capture ratios, in the order of `NAMES`.
    fn reference(pairs: &[(f64, f64)]) -> [f64; 8] {
        let up: Vec<(f64, f64)> = pairs.iter().copied().filter(|&(_, b)| b > 0.0).collect();
        let dn: Vec<(f64, f64)> = pairs.iter().copied().filter(|&(_, b)| b <= 0.0).collect();
        let dn_strict: Vec<(f64, f64)> = pairs.iter().copied().filter(|&(_, b)| b < 0.0).collect();
        let sum = |v: &[(f64, f64)], f: fn(&(f64, f64)) -> f64| v.iter().map(f).sum::<f64>();
        let count = |v: &[(f64, f64)], f: fn(&(f64, f64)) -> bool| v.iter().filter(|p| f(p)).count() as f64;
        [
            nan_div(
                prod(up.iter().map(|&(a, _)| 1.0 + a)) - 1.0,
                prod(up.iter().map(|&(_, b)| 1.0 + b)) - 1.0,
            ),
            nan_div(sum(&up, |p| p.0), sum(&up, |p| p.1)),
            nan_div(
                prod(dn.iter().map(|&(a, _)| 1.0 + a)) - 1.0,
                prod(dn.iter().map(|&(_, b)| 1.0 + b)) - 1.0,
            ),
            nan_div(sum(&dn, |p| p.0), sum(&dn, |p| p.1)),
            nan_div(count(&up, |p| p.0 > 0.0), up.len() as f64),
            nan_div(count(&dn, |p| p.0 < 0.0), dn.len() as f64),
            nan_div(count(&up, |p| p.0 > p.1), up.len() as f64),
            nan_div(count(&dn_strict, |p| p.0 > p.1), dn_strict.len() as f64),
        ]
    }

    fn assert_matches(c: &Capture, pairs: &[(f64, f64)], places: i32, msg: &str) {
        for ((name, getter), expected) in NAMES.iter().zip(reference(pairs)) {
            assert_close(getter(c), expected, places, &format!("{msg} {name}"));
        }
    }

    fn assert_all_nan(c: &Capture) {
        for (name, getter) in NAMES {
            assert!(getter(c).is_nan(), "{name}");
        }
    }

    fn random_pairs(seed: u64, n: usize) -> Vec<(f64, f64)> {
        let mut rng = Rng::new(seed);
        (0..n)
            .map(|_| {
                let g = gauss(&mut rng, 0.0, 0.03);
                let a = choice2(&mut rng, 0.0, g);
                let g = gauss(&mut rng, 0.0, 0.03);
                let b = choice2(&mut rng, 0.0, g);
                (a, b)
            })
            .collect()
    }

    #[test]
    fn test_empty() {
        assert_all_nan(&Capture::new());
    }

    #[test]
    fn test_hand_computed() {
        let pairs = [(0.02, 0.01), (-0.01, -0.02), (0.03, 0.04), (-0.03, 0.0), (0.01, -0.01)];
        let mut c = Capture::new();
        for (a, b) in pairs {
            c.update(a, b);
        }
        // Up periods (b > 0): (0.02, 0.01), (0.03, 0.04)
        assert!(almost_equal(c.upside_capture_ratio_arithmetic(), 0.05 / 0.05, 15));
        assert!(almost_equal(
            c.upside_capture_ratio_geometric(),
            (1.02 * 1.03 - 1.0) / (1.01 * 1.04 - 1.0),
            14
        ));
        assert_eq!(c.up_number_ratio(), 1.0);
        assert_eq!(c.up_percentage_ratio(), 0.5);
        // Down periods (b <= 0) include the zero-benchmark period.
        assert!(almost_equal(
            c.downside_capture_ratio_arithmetic(),
            (-0.01 - 0.03 + 0.01) / (-0.02 + 0.0 - 0.01),
            15
        ));
        assert!(almost_equal(c.down_number_ratio(), 2.0 / 3.0, 15));
        // Down-percentage only counts strictly negative benchmark periods.
        assert_eq!(c.down_percentage_ratio(), 1.0);
    }

    #[test]
    fn test_matches_reference() {
        let pairs = random_pairs(42, 200);
        let mut c = Capture::new();
        for &(a, b) in &pairs {
            c.update(a, b);
        }
        assert_matches(&c, &pairs, 12, "");
    }

    #[test]
    fn test_rolling_window_matches_reference() {
        let pairs = random_pairs(7, 120);
        let w = 9;
        let mut c = Capture::new();
        for (i, &(a, b)) in pairs.iter().enumerate() {
            if i >= w {
                let (oa, ob) = pairs[i - w];
                c.revert(oa, ob);
            }
            c.update(a, b);
            let lo = (i + 1).saturating_sub(w);
            assert_matches(&c, &pairs[lo..=i], 12, &format!("step {i}"));
        }
    }

    #[test]
    fn test_reset() {
        let mut c = Capture::new();
        c.update(0.01, 0.02);
        c.update(-0.01, -0.02);
        c.reset();
        assert_all_nan(&c);
    }
}
