/// Crossover signals.
use crate::fuzzy::{MembershipShape, mu_greater, mu_less, t_product};

/// Degree to which a value crossed above `threshold` from below.
pub fn mu_crosses_above(prev_value: f64, curr_value: f64, threshold: f64, width: f64, shape: MembershipShape) -> f64 {
    let was_below = mu_less(prev_value, threshold, width, shape);
    let is_above = mu_greater(curr_value, threshold, width, shape);
    t_product(was_below, is_above)
}

/// Degree to which a value crossed below `threshold` from above.
pub fn mu_crosses_below(prev_value: f64, curr_value: f64, threshold: f64, width: f64, shape: MembershipShape) -> f64 {
    let was_above = mu_greater(prev_value, threshold, width, shape);
    let is_below = mu_less(curr_value, threshold, width, shape);
    t_product(was_above, is_below)
}

/// Degree to which a fast line crossed above a slow line.
pub fn mu_line_crosses_above(prev_fast: f64, curr_fast: f64, prev_slow: f64, curr_slow: f64, width: f64, shape: MembershipShape) -> f64 {
    let prev_diff = prev_fast - prev_slow;
    let curr_diff = curr_fast - curr_slow;
    mu_crosses_above(prev_diff, curr_diff, 0.0, width, shape)
}

/// Degree to which a fast line crossed below a slow line.
pub fn mu_line_crosses_below(prev_fast: f64, curr_fast: f64, prev_slow: f64, curr_slow: f64, width: f64, shape: MembershipShape) -> f64 {
    let prev_diff = prev_fast - prev_slow;
    let curr_diff = curr_fast - curr_slow;
    mu_crosses_below(prev_diff, curr_diff, 0.0, width, shape)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, eps: f64) -> bool { (a - b).abs() < eps }

    #[test]
    fn test_clear_cross_above() { assert!(almost_equal(mu_crosses_above(25.0, 35.0, 30.0, 0.0, MembershipShape::Sigmoid), 1.0, 1e-10)); }
    #[test]
    fn test_no_cross_both_above() { assert!(almost_equal(mu_crosses_above(35.0, 40.0, 30.0, 0.0, MembershipShape::Sigmoid), 0.0, 1e-10)); }
    #[test]
    fn test_no_cross_both_below() { assert!(almost_equal(mu_crosses_above(25.0, 28.0, 30.0, 0.0, MembershipShape::Sigmoid), 0.0, 1e-10)); }
    #[test]
    fn test_at_threshold() { assert!(almost_equal(mu_crosses_above(30.0, 30.0, 30.0, 0.0, MembershipShape::Sigmoid), 0.25, 1e-10)); }
    #[test]
    fn test_clear_cross_below() { assert!(almost_equal(mu_crosses_below(35.0, 25.0, 30.0, 0.0, MembershipShape::Sigmoid), 1.0, 1e-10)); }
    #[test]
    fn test_symmetry() {
        let cb = mu_crosses_below(35.0, 25.0, 30.0, 2.0, MembershipShape::Sigmoid);
        let ca = mu_crosses_above(25.0, 35.0, 30.0, 2.0, MembershipShape::Sigmoid);
        assert!(almost_equal(cb, ca, 1e-10));
    }
    #[test]
    fn test_golden_cross() { assert!(almost_equal(mu_line_crosses_above(49.0, 51.0, 50.0, 50.0, 0.0, MembershipShape::Sigmoid), 1.0, 1e-10)); }
    #[test]
    fn test_no_line_cross() { assert!(almost_equal(mu_line_crosses_above(52.0, 53.0, 50.0, 50.0, 0.0, MembershipShape::Sigmoid), 0.0, 1e-10)); }
    #[test]
    fn test_death_cross() { assert!(almost_equal(mu_line_crosses_below(51.0, 49.0, 50.0, 50.0, 0.0, MembershipShape::Sigmoid), 1.0, 1e-10)); }
}
