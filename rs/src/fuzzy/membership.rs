/// Membership functions for fuzzy logic.
///
/// Each function maps a crisp value to a membership degree μ ∈ [0, 1].
/// Two shapes are supported: SIGMOID (default, smooth) and LINEAR (trapezoidal ramp).
/// All functions degrade to crisp step functions when width = 0.

/// Shape of the fuzzy membership transition curve.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum MembershipShape {
    /// Smooth logistic curve. Default for most applications.
    Sigmoid = 0,
    /// Piecewise-linear ramp (trapezoidal/triangular).
    Linear = 1,
}

/// Steepness constant for sigmoid shape.
/// k = SIGMOID_K / width gives ≈0.997 at threshold ± width/2.
const SIGMOID_K: f64 = 12.0;

/// Logistic sigmoid: 1 / (1 + exp(k * (x - threshold))).
///
/// Returns the "less-than" membership: high when x << threshold,
/// low when x >> threshold, exactly 0.5 at x == threshold.
fn sigmoid(x: f64, threshold: f64, k: f64) -> f64 {
    let exponent = k * (x - threshold);
    // Clamp to avoid overflow in exp().
    if exponent > 500.0 {
        return 0.0;
    }
    if exponent < -500.0 {
        return 1.0;
    }
    1.0 / (1.0 + exponent.exp())
}

/// Degree to which `x` is less than `threshold`.
///
/// At `threshold`: μ = 0.5.
/// At `threshold - width/2`: μ ≈ 0.997 (sigmoid) or 1.0 (linear).
/// At `threshold + width/2`: μ ≈ 0.003 (sigmoid) or 0.0 (linear).
///
/// When `width` = 0 (crisp): 1.0 if x < threshold, 0.5 if x == threshold,
/// 0.0 if x > threshold.
pub fn mu_less(x: f64, threshold: f64, width: f64, shape: MembershipShape) -> f64 {
    if width <= 0.0 {
        if x < threshold {
            return 1.0;
        }
        if x > threshold {
            return 0.0;
        }
        return 0.5;
    }

    if shape == MembershipShape::Linear {
        let half = width * 0.5;
        if x <= threshold - half {
            return 1.0;
        }
        if x >= threshold + half {
            return 0.0;
        }
        return (threshold + half - x) / width;
    }
    // sigmoid
    sigmoid(x, threshold, SIGMOID_K / width)
}

/// Degree to which `x` ≤ `threshold`.
/// Identical to `mu_less` for continuous values — the distinction is conceptual.
pub fn mu_less_equal(x: f64, threshold: f64, width: f64, shape: MembershipShape) -> f64 {
    mu_less(x, threshold, width, shape)
}

/// Degree to which `x` > `threshold`. Complement of `mu_less`.
pub fn mu_greater(x: f64, threshold: f64, width: f64, shape: MembershipShape) -> f64 {
    1.0 - mu_less(x, threshold, width, shape)
}

/// Degree to which `x` ≥ `threshold`. Complement of `mu_less_equal`.
pub fn mu_greater_equal(x: f64, threshold: f64, width: f64, shape: MembershipShape) -> f64 {
    1.0 - mu_less_equal(x, threshold, width, shape)
}

/// Bell-shaped membership: degree to which `x` ≈ `target`.
///
/// μ = 1.0 at `x == target`.
/// μ ≈ 0 at `|x - target| ≥ width`.
///
/// For sigmoid shape: Gaussian bell `exp(-k * (x - target)²)`.
/// For linear shape: triangular peak at target with base `2 * width`.
pub fn mu_near(x: f64, target: f64, width: f64, shape: MembershipShape) -> f64 {
    if width <= 0.0 {
        return if x == target { 1.0 } else { 0.0 };
    }

    if shape == MembershipShape::Linear {
        let dist = (x - target).abs();
        if dist >= width {
            return 0.0;
        }
        return 1.0 - dist / width;
    }
    // sigmoid → Gaussian bell
    // σ chosen so that μ ≈ 0.003 at |x - target| = width.
    let sigma = width / 2.41;
    let d = (x - target) / sigma;
    (-d * d).exp()
}

/// Fuzzy candle direction ∈ [-1, +1].
///
/// +1 = fully bullish (large white body).
///  0 = neutral (doji-like).
/// -1 = fully bearish (large black body).
///
/// Uses `tanh(steepness * (c - o) / body_avg)`.
/// When `body_avg` ≤ 0: returns +1.0 if `c ≥ o`, else -1.0 (crisp).
pub fn mu_direction(o: f64, c: f64, body_avg: f64, steepness: f64) -> f64 {
    if body_avg <= 0.0 {
        return if c >= o { 1.0 } else { -1.0 };
    }
    (steepness * (c - o) / body_avg).tanh()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, epsilon: f64) -> bool {
        (a - b).abs() < epsilon
    }

    #[test]
    fn test_mu_less_crossover_at_threshold() {
        assert!(almost_equal(mu_less(10.0, 10.0, 2.0, MembershipShape::Sigmoid), 0.5, 1e-10));
    }

    #[test]
    fn test_mu_less_well_below() {
        assert!(mu_less(8.0, 10.0, 2.0, MembershipShape::Sigmoid) > 0.99);
    }

    #[test]
    fn test_mu_less_well_above() {
        assert!(mu_less(12.0, 10.0, 2.0, MembershipShape::Sigmoid) < 0.01);
    }

    #[test]
    fn test_mu_less_monotonic() {
        let xs = [8.0, 9.0, 10.0, 11.0, 12.0];
        for i in 0..xs.len() - 1 {
            assert!(mu_less(xs[i], 10.0, 2.0, MembershipShape::Sigmoid) >
                    mu_less(xs[i + 1], 10.0, 2.0, MembershipShape::Sigmoid));
        }
    }

    #[test]
    fn test_mu_less_symmetry() {
        let below = mu_less(9.0, 10.0, 2.0, MembershipShape::Sigmoid);
        let above = mu_less(11.0, 10.0, 2.0, MembershipShape::Sigmoid);
        assert!(almost_equal(below + above, 1.0, 1e-10));
    }

    #[test]
    fn test_mu_less_linear_crossover() {
        assert!(almost_equal(mu_less(10.0, 10.0, 4.0, MembershipShape::Linear), 0.5, 1e-10));
    }

    #[test]
    fn test_mu_less_linear_below_range() {
        assert_eq!(mu_less(7.0, 10.0, 4.0, MembershipShape::Linear), 1.0);
    }

    #[test]
    fn test_mu_less_linear_above_range() {
        assert_eq!(mu_less(13.0, 10.0, 4.0, MembershipShape::Linear), 0.0);
    }

    #[test]
    fn test_mu_less_linear_midpoint() {
        assert!(almost_equal(mu_less(9.0, 10.0, 4.0, MembershipShape::Linear), 0.75, 1e-10));
    }

    #[test]
    fn test_mu_less_crisp_below() { assert_eq!(mu_less(9.0, 10.0, 0.0, MembershipShape::Sigmoid), 1.0); }

    #[test]
    fn test_mu_less_crisp_above() { assert_eq!(mu_less(11.0, 10.0, 0.0, MembershipShape::Sigmoid), 0.0); }

    #[test]
    fn test_mu_less_crisp_at_threshold() { assert_eq!(mu_less(10.0, 10.0, 0.0, MembershipShape::Sigmoid), 0.5); }

    #[test]
    fn test_mu_less_equal_same_as_less() {
        assert_eq!(mu_less_equal(9.5, 10.0, 2.0, MembershipShape::Sigmoid),
                   mu_less(9.5, 10.0, 2.0, MembershipShape::Sigmoid));
    }

    #[test]
    fn test_mu_greater_complement() {
        for x in [8.0, 9.0, 10.0, 11.0, 12.0] {
            let sum = mu_greater(x, 10.0, 2.0, MembershipShape::Sigmoid) +
                      mu_less(x, 10.0, 2.0, MembershipShape::Sigmoid);
            assert!(almost_equal(sum, 1.0, 1e-10));
        }
    }

    #[test]
    fn test_mu_greater_crossover() {
        assert!(almost_equal(mu_greater(10.0, 10.0, 2.0, MembershipShape::Sigmoid), 0.5, 1e-10));
    }

    #[test]
    fn test_mu_greater_well_above() { assert!(mu_greater(12.0, 10.0, 2.0, MembershipShape::Sigmoid) > 0.99); }

    #[test]
    fn test_mu_greater_well_below() { assert!(mu_greater(8.0, 10.0, 2.0, MembershipShape::Sigmoid) < 0.01); }

    #[test]
    fn test_mu_greater_equal_complement() {
        let sum = mu_greater_equal(9.5, 10.0, 2.0, MembershipShape::Sigmoid) +
                  mu_less_equal(9.5, 10.0, 2.0, MembershipShape::Sigmoid);
        assert!(almost_equal(sum, 1.0, 1e-10));
    }

    #[test]
    fn test_mu_near_peak() {
        assert!(almost_equal(mu_near(10.0, 10.0, 2.0, MembershipShape::Sigmoid), 1.0, 1e-10));
    }

    #[test]
    fn test_mu_near_falls_off() { assert!(mu_near(12.0, 10.0, 2.0, MembershipShape::Sigmoid) < 0.05); }

    #[test]
    fn test_mu_near_symmetric() {
        let below = mu_near(9.0, 10.0, 2.0, MembershipShape::Sigmoid);
        let above = mu_near(11.0, 10.0, 2.0, MembershipShape::Sigmoid);
        assert!(almost_equal(below, above, 1e-10));
    }

    #[test]
    fn test_mu_near_monotonic() {
        let offsets = [0.0, 0.5, 1.0, 1.5, 2.0];
        for i in 0..offsets.len() - 1 {
            assert!(mu_near(10.0 + offsets[i], 10.0, 2.0, MembershipShape::Sigmoid) >
                    mu_near(10.0 + offsets[i + 1], 10.0, 2.0, MembershipShape::Sigmoid));
        }
    }

    #[test]
    fn test_mu_near_linear_peak() {
        assert!(almost_equal(mu_near(10.0, 10.0, 2.0, MembershipShape::Linear), 1.0, 1e-10));
    }

    #[test]
    fn test_mu_near_linear_boundary() { assert_eq!(mu_near(12.0, 10.0, 2.0, MembershipShape::Linear), 0.0); }

    #[test]
    fn test_mu_near_linear_midpoint() {
        assert!(almost_equal(mu_near(11.0, 10.0, 2.0, MembershipShape::Linear), 0.5, 1e-10));
    }

    #[test]
    fn test_mu_near_crisp_exact() { assert_eq!(mu_near(10.0, 10.0, 0.0, MembershipShape::Sigmoid), 1.0); }

    #[test]
    fn test_mu_near_crisp_any_distance() { assert_eq!(mu_near(10.1, 10.0, 0.0, MembershipShape::Sigmoid), 0.0); }

    #[test]
    fn test_mu_direction_large_white() { assert!(mu_direction(100.0, 110.0, 5.0, 2.0) > 0.95); }

    #[test]
    fn test_mu_direction_large_black() { assert!(mu_direction(110.0, 100.0, 5.0, 2.0) < -0.95); }

    #[test]
    fn test_mu_direction_doji() {
        assert!(almost_equal(mu_direction(100.0, 100.0, 5.0, 2.0), 0.0, 1e-10));
    }

    #[test]
    fn test_mu_direction_tiny_white() {
        let d = mu_direction(100.0, 100.1, 5.0, 2.0);
        assert!(d > 0.0 && d < 0.1);
    }

    #[test]
    fn test_mu_direction_antisymmetric() {
        let d1 = mu_direction(100.0, 105.0, 5.0, 2.0);
        let d2 = mu_direction(105.0, 100.0, 5.0, 2.0);
        assert!(almost_equal(d1, -d2, 1e-10));
    }

    #[test]
    fn test_mu_direction_zero_body_avg_white() { assert_eq!(mu_direction(100.0, 101.0, 0.0, 2.0), 1.0); }

    #[test]
    fn test_mu_direction_zero_body_avg_black() { assert_eq!(mu_direction(101.0, 100.0, 0.0, 2.0), -1.0); }

    #[test]
    fn test_mu_direction_zero_body_avg_doji() { assert_eq!(mu_direction(100.0, 100.0, 0.0, 2.0), 1.0); }

    #[test]
    fn test_mu_direction_range_bounded() {
        for (o, c, avg) in [(0.0, 1000.0, 1.0), (1000.0, 0.0, 1.0), (50.0, 50.0, 100.0)] {
            let d = mu_direction(o, c, avg, 2.0);
            assert!(d >= -1.0 && d <= 1.0);
        }
    }

    #[test]
    fn test_very_large_x() { assert_eq!(mu_less(1e10, 0.0, 1.0, MembershipShape::Sigmoid), 0.0); }

    #[test]
    fn test_very_small_x() { assert_eq!(mu_less(-1e10, 0.0, 1.0, MembershipShape::Sigmoid), 1.0); }

    #[test]
    fn test_tiny_width() { assert!(mu_less(9.999, 10.0, 0.001, MembershipShape::Sigmoid) > 0.99); }

    #[test]
    fn test_huge_width() {
        let val = mu_less(0.0, 10.0, 1000.0, MembershipShape::Sigmoid);
        assert!(val > 0.49 && val < 0.60);
    }
}
