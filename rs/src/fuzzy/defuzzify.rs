/// Defuzzification utilities.
///
/// Provides alpha-cut conversion from continuous fuzzy output back to crisp
/// discrete values for backward compatibility with TA-Lib-style integer outputs.

/// Convert a continuous fuzzy output to a crisp discrete value.
///
/// The confidence is `abs(value) / scale`. If confidence ≥ `alpha`,
/// the output is rounded to the nearest multiple of `scale` with the
/// original sign preserved. Otherwise 0 is returned.
pub fn alpha_cut(value: f64, alpha: f64, scale: f64) -> i32 {
    if scale <= 0.0 {
        return 0;
    }
    let confidence = value.abs() / scale;
    if confidence < alpha - 1e-10 {
        return 0;
    }
    let sign = if value >= 0.0 { 1 } else { -1 };
    // Round to nearest multiple of scale.
    let level = confidence.round().max(1.0);
    sign * (level * scale) as i32
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_strong_bearish() { assert_eq!(alpha_cut(-87.3, 0.5, 100.0), -100); }
    #[test]
    fn test_weak_bearish() { assert_eq!(alpha_cut(-32.1, 0.5, 100.0), 0); }
    #[test]
    fn test_strong_bullish() { assert_eq!(alpha_cut(92.5, 0.5, 100.0), 100); }
    #[test]
    fn test_weak_bullish() { assert_eq!(alpha_cut(15.0, 0.5, 100.0), 0); }
    #[test]
    fn test_zero() { assert_eq!(alpha_cut(0.0, 0.5, 100.0), 0); }
    #[test]
    fn test_strong_confirmation() { assert_eq!(alpha_cut(156.8, 0.5, 100.0), 200); }
    #[test]
    fn test_negative_confirmation() { assert_eq!(alpha_cut(-180.0, 0.5, 100.0), -200); }
    #[test]
    fn test_high_alpha_filters_more() { assert_eq!(alpha_cut(-87.3, 0.9, 100.0), 0); }
    #[test]
    fn test_high_alpha_passes_strong() { assert_eq!(alpha_cut(-95.0, 0.9, 100.0), -100); }
    #[test]
    fn test_low_alpha_passes_more() { assert_eq!(alpha_cut(-15.0, 0.1, 100.0), -100); }
    #[test]
    fn test_alpha_zero_passes_all() { assert_eq!(alpha_cut(-1.0, 0.0, 100.0), -100); }
    #[test]
    fn test_exactly_at_threshold() { assert_eq!(alpha_cut(50.0, 0.5, 100.0), 100); }
    #[test]
    fn test_just_below_threshold() { assert_eq!(alpha_cut(49.9, 0.5, 100.0), 0); }
    #[test]
    fn test_exactly_100() { assert_eq!(alpha_cut(100.0, 0.5, 100.0), 100); }
    #[test]
    fn test_exactly_minus_100() { assert_eq!(alpha_cut(-100.0, 0.5, 100.0), -100); }
    #[test]
    fn test_custom_scale() { assert_eq!(alpha_cut(-40.0, 0.5, 50.0), -50); }
    #[test]
    fn test_invalid_scale() { assert_eq!(alpha_cut(-87.3, 0.5, 0.0), 0); }
}
