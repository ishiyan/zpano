/// Histogram sign-change signals.
use crate::fuzzy::{MembershipShape, mu_greater, mu_less, t_product};

/// Degree to which a histogram turned from non-positive to positive.
pub fn mu_turns_positive(prev_value: f64, curr_value: f64, width: f64, shape: MembershipShape) -> f64 {
    let was_nonpositive = mu_less(prev_value, 0.0, width, shape);
    let is_positive = mu_greater(curr_value, 0.0, width, shape);
    t_product(was_nonpositive, is_positive)
}

/// Degree to which a histogram turned from non-negative to negative.
pub fn mu_turns_negative(prev_value: f64, curr_value: f64, width: f64, shape: MembershipShape) -> f64 {
    let was_nonnegative = mu_greater(prev_value, 0.0, width, shape);
    let is_negative = mu_less(curr_value, 0.0, width, shape);
    t_product(was_nonnegative, is_negative)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, eps: f64) -> bool { (a - b).abs() < eps }

    #[test]
    fn test_clear_turn_positive() { assert!(almost_equal(mu_turns_positive(-5.0, 5.0, 0.0, MembershipShape::Sigmoid), 1.0, 1e-10)); }
    #[test]
    fn test_stays_positive() { assert!(almost_equal(mu_turns_positive(3.0, 5.0, 0.0, MembershipShape::Sigmoid), 0.0, 1e-10)); }
    #[test]
    fn test_stays_negative() { assert!(almost_equal(mu_turns_positive(-5.0, -3.0, 0.0, MembershipShape::Sigmoid), 0.0, 1e-10)); }
    #[test]
    fn test_from_zero() { assert!(almost_equal(mu_turns_positive(0.0, 5.0, 0.0, MembershipShape::Sigmoid), 0.5, 1e-10)); }
    #[test]
    fn test_clear_turn_negative() { assert!(almost_equal(mu_turns_negative(5.0, -5.0, 0.0, MembershipShape::Sigmoid), 1.0, 1e-10)); }
    #[test]
    fn test_symmetry() {
        let tn = mu_turns_negative(3.0, -3.0, 1.0, MembershipShape::Sigmoid);
        let tp = mu_turns_positive(-3.0, 3.0, 1.0, MembershipShape::Sigmoid);
        assert!(almost_equal(tn, tp, 1e-10));
    }
}
