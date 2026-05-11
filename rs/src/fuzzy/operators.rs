/// Fuzzy logic operators: t-norms, s-norms, and negation.
///
/// T-norms implement fuzzy AND. S-norms implement fuzzy OR.
/// All operators take membership degrees in [0, 1] and return [0, 1].

// -----------------------------------------------------------------------
// T-norms (fuzzy AND)
// -----------------------------------------------------------------------

/// Product t-norm: `a * b`.
/// All conditions contribute proportionally. The default choice.
pub fn t_product(a: f64, b: f64) -> f64 {
    a * b
}

/// Minimum t-norm (Zadeh): `min(a, b)`.
/// Dominated by the weakest condition.
pub fn t_min(a: f64, b: f64) -> f64 {
    a.min(b)
}

/// Łukasiewicz t-norm: `max(0, a + b - 1)`.
/// Very strict — both conditions must have high membership.
pub fn t_lukasiewicz(a: f64, b: f64) -> f64 {
    (a + b - 1.0).max(0.0)
}

// -----------------------------------------------------------------------
// S-norms (fuzzy OR)
// -----------------------------------------------------------------------

/// Probabilistic sum: `a + b - a*b`.
/// Dual of the product t-norm.
pub fn s_probabilistic(a: f64, b: f64) -> f64 {
    a + b - a * b
}

/// Maximum s-norm (Zadeh): `max(a, b)`.
/// Dual of the minimum t-norm.
pub fn s_max(a: f64, b: f64) -> f64 {
    a.max(b)
}

// -----------------------------------------------------------------------
// Negation
// -----------------------------------------------------------------------

/// Standard fuzzy negation: `1 - a`.
pub fn f_not(a: f64) -> f64 {
    1.0 - a
}

// -----------------------------------------------------------------------
// Variadic helpers
// -----------------------------------------------------------------------

/// Product t-norm over a slice of arguments.
/// Returns 1.0 for empty slice (identity element of product).
pub fn t_product_all(args: &[f64]) -> f64 {
    let mut result = 1.0;
    for &a in args {
        result *= a;
    }
    result
}

/// Minimum t-norm over a slice of arguments.
/// Returns 1.0 for empty slice (identity element of min).
pub fn t_min_all(args: &[f64]) -> f64 {
    if args.is_empty() {
        return 1.0;
    }
    let mut result = args[0];
    for &a in &args[1..] {
        if a < result {
            result = a;
        }
    }
    result
}

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, epsilon: f64) -> bool {
        (a - b).abs() < epsilon
    }

    // T-norms
    #[test]
    fn test_product_basic() { assert!(almost_equal(t_product(0.8, 0.6), 0.48, 1e-10)); }
    #[test]
    fn test_product_identity() { assert!(almost_equal(t_product(0.7, 1.0), 0.7, 1e-10)); }
    #[test]
    fn test_product_annihilator() { assert!(almost_equal(t_product(0.7, 0.0), 0.0, 1e-10)); }
    #[test]
    fn test_product_commutativity() { assert!(almost_equal(t_product(0.3, 0.8), t_product(0.8, 0.3), 1e-10)); }
    #[test]
    fn test_min_basic() { assert_eq!(t_min(0.8, 0.6), 0.6); }
    #[test]
    fn test_min_identity() { assert_eq!(t_min(0.7, 1.0), 0.7); }
    #[test]
    fn test_min_annihilator() { assert_eq!(t_min(0.7, 0.0), 0.0); }
    #[test]
    fn test_lukasiewicz_both_high() { assert!(almost_equal(t_lukasiewicz(0.9, 0.8), 0.7, 1e-10)); }
    #[test]
    fn test_lukasiewicz_one_low() { assert!(almost_equal(t_lukasiewicz(0.3, 0.5), 0.0, 1e-10)); }
    #[test]
    fn test_lukasiewicz_clamp() { assert_eq!(t_lukasiewicz(0.1, 0.2), 0.0); }
    #[test]
    fn test_lukasiewicz_identity() { assert!(almost_equal(t_lukasiewicz(0.7, 1.0), 0.7, 1e-10)); }

    // S-norms
    #[test]
    fn test_probabilistic_basic() { assert!(almost_equal(s_probabilistic(0.8, 0.6), 0.92, 1e-10)); }
    #[test]
    fn test_probabilistic_identity() { assert!(almost_equal(s_probabilistic(0.7, 0.0), 0.7, 1e-10)); }
    #[test]
    fn test_probabilistic_annihilator() { assert!(almost_equal(s_probabilistic(0.7, 1.0), 1.0, 1e-10)); }
    #[test]
    fn test_max_basic() { assert_eq!(s_max(0.8, 0.6), 0.8); }
    #[test]
    fn test_max_identity() { assert_eq!(s_max(0.7, 0.0), 0.7); }

    // Negation
    #[test]
    fn test_not_basic() { assert!(almost_equal(f_not(0.3), 0.7, 1e-10)); }
    #[test]
    fn test_not_zero() { assert!(almost_equal(f_not(0.0), 1.0, 1e-10)); }
    #[test]
    fn test_not_one() { assert!(almost_equal(f_not(1.0), 0.0, 1e-10)); }
    #[test]
    fn test_double_negation() { assert!(almost_equal(f_not(f_not(0.4)), 0.4, 1e-10)); }

    // Variadic
    #[test]
    fn test_product_all_three() { assert!(almost_equal(t_product_all(&[0.8, 0.6, 0.5]), 0.24, 1e-10)); }
    #[test]
    fn test_product_all_single() { assert!(almost_equal(t_product_all(&[0.7]), 0.7, 1e-10)); }
    #[test]
    fn test_product_all_empty() { assert!(almost_equal(t_product_all(&[]), 1.0, 1e-10)); }
    #[test]
    fn test_min_all_three() { assert_eq!(t_min_all(&[0.8, 0.6, 0.9]), 0.6); }
    #[test]
    fn test_min_all_empty() { assert_eq!(t_min_all(&[]), 1.0); }
    #[test]
    fn test_product_all_five() {
        let result = t_product_all(&[0.9, 0.9, 0.9, 0.9, 0.9]);
        assert!(almost_equal(result, 0.9_f64.powi(5), 1e-10));
    }

    // Duality
    #[test]
    fn test_product_probabilistic_duality() {
        let (a, b) = (0.7, 0.4);
        assert!(almost_equal(t_product(a, b), f_not(s_probabilistic(f_not(a), f_not(b))), 1e-10));
    }
    #[test]
    fn test_min_max_duality() {
        let (a, b) = (0.7, 0.4);
        assert!(almost_equal(t_min(a, b), f_not(s_max(f_not(a), f_not(b))), 1e-10));
    }
}
