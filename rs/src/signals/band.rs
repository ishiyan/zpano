/// Band signals.
use crate::fuzzy::{MembershipShape, mu_greater, mu_less, t_product};

/// Degree to which `value` is above the upper band.
pub fn mu_above_band(value: f64, upper_band: f64, width: f64, shape: MembershipShape) -> f64 {
    mu_greater(value, upper_band, width, shape)
}

/// Degree to which `value` is below the lower band.
pub fn mu_below_band(value: f64, lower_band: f64, width: f64, shape: MembershipShape) -> f64 {
    mu_less(value, lower_band, width, shape)
}

/// Degree to which `value` is inside the band channel.
pub fn mu_between_bands(value: f64, lower_band: f64, upper_band: f64, shape: MembershipShape) -> f64 {
    if upper_band <= lower_band {
        return 0.0;
    }
    let spread = upper_band - lower_band;
    let width = spread * 0.5;
    let above_lower = mu_greater(value, lower_band, width, shape);
    let below_upper = mu_less(value, upper_band, width, shape);
    t_product(above_lower, below_upper)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, eps: f64) -> bool { (a - b).abs() < eps }

    #[test]
    fn test_above_well_above() { assert!(almost_equal(mu_above_band(110.0, 100.0, 5.0, MembershipShape::Sigmoid), 1.0, 0.01)); }
    #[test]
    fn test_above_at_band() { assert!(almost_equal(mu_above_band(100.0, 100.0, 5.0, MembershipShape::Sigmoid), 0.5, 1e-10)); }
    #[test]
    fn test_below_well_below() { assert!(almost_equal(mu_below_band(85.0, 90.0, 5.0, MembershipShape::Sigmoid), 1.0, 0.01)); }
    #[test]
    fn test_between_centered() { assert!(mu_between_bands(100.0, 90.0, 110.0, MembershipShape::Sigmoid) > 0.8); }
    #[test]
    fn test_between_at_upper() { assert!(mu_between_bands(110.0, 90.0, 110.0, MembershipShape::Sigmoid) < 0.6); }
    #[test]
    fn test_between_outside() { assert!(mu_between_bands(130.0, 90.0, 110.0, MembershipShape::Sigmoid) < 0.1); }
    #[test]
    fn test_between_degenerate() {
        assert_eq!(mu_between_bands(100.0, 110.0, 90.0, MembershipShape::Sigmoid), 0.0);
        assert_eq!(mu_between_bands(100.0, 100.0, 100.0, MembershipShape::Sigmoid), 0.0);
    }
    #[test]
    fn test_between_monotonic() {
        let center = mu_between_bands(100.0, 90.0, 110.0, MembershipShape::Sigmoid);
        let edge = mu_between_bands(108.0, 90.0, 110.0, MembershipShape::Sigmoid);
        let outside = mu_between_bands(115.0, 90.0, 110.0, MembershipShape::Sigmoid);
        assert!(center > edge);
        assert!(edge > outside);
    }
}
