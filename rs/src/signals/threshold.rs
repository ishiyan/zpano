/// Threshold crossing signals.
use crate::fuzzy::{MembershipShape, mu_greater, mu_less};

/// Degree to which `value` is above `threshold`.
pub fn mu_above(value: f64, threshold: f64, width: f64, shape: MembershipShape) -> f64 {
    mu_greater(value, threshold, width, shape)
}

/// Degree to which `value` is below `threshold`. Complement of `mu_above`.
pub fn mu_below(value: f64, threshold: f64, width: f64, shape: MembershipShape) -> f64 {
    mu_less(value, threshold, width, shape)
}

/// Degree of overbought condition.
pub fn mu_overbought(value: f64, level: f64, width: f64, shape: MembershipShape) -> f64 {
    mu_greater(value, level, width, shape)
}

/// Degree of oversold condition.
pub fn mu_oversold(value: f64, level: f64, width: f64, shape: MembershipShape) -> f64 {
    mu_less(value, level, width, shape)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, eps: f64) -> bool { (a - b).abs() < eps }

    #[test]
    fn test_above_well_above() { assert!((mu_above(80.0, 70.0, 5.0, MembershipShape::Sigmoid) - 1.0).abs() < 0.01); }
    #[test]
    fn test_above_well_below() { assert!((mu_above(60.0, 70.0, 5.0, MembershipShape::Sigmoid)).abs() < 0.01); }
    #[test]
    fn test_above_at_threshold() { assert!(almost_equal(mu_above(70.0, 70.0, 5.0, MembershipShape::Sigmoid), 0.5, 1e-10)); }
    #[test]
    fn test_above_zero_width() { assert_eq!(mu_above(70.1, 70.0, 0.0, MembershipShape::Sigmoid), 1.0); }
    #[test]
    fn test_below_well_below() { assert!((mu_below(20.0, 30.0, 5.0, MembershipShape::Sigmoid) - 1.0).abs() < 0.01); }
    #[test]
    fn test_below_at_threshold() { assert!(almost_equal(mu_below(30.0, 30.0, 5.0, MembershipShape::Sigmoid), 0.5, 1e-10)); }
    #[test]
    fn test_complement() {
        for v in [25.0, 30.0, 35.0, 50.0] {
            let total = mu_below(v, 30.0, 5.0, MembershipShape::Sigmoid) + mu_above(v, 30.0, 5.0, MembershipShape::Sigmoid);
            assert!(almost_equal(total, 1.0, 1e-10));
        }
    }
    #[test]
    fn test_overbought_high() { assert!(mu_overbought(85.0, 70.0, 5.0, MembershipShape::Sigmoid) > 0.95); }
    #[test]
    fn test_oversold_low() { assert!(mu_oversold(15.0, 30.0, 5.0, MembershipShape::Sigmoid) > 0.95); }
    #[test]
    fn test_overbought_custom() { assert!(almost_equal(mu_overbought(80.0, 80.0, 5.0, MembershipShape::Sigmoid), 0.5, 1e-10)); }
    #[test]
    fn test_oversold_custom() { assert!(almost_equal(mu_oversold(20.0, 20.0, 5.0, MembershipShape::Sigmoid), 0.5, 1e-10)); }
}
