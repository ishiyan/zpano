/// Fuzzy logic operators: t-norms, s-norms, and negation.
///
/// T-norms implement fuzzy AND. S-norms implement fuzzy OR.
/// All operators take membership degrees in [0, 1] and return [0, 1].
const std = @import("std");

// -----------------------------------------------------------------------
// T-norms (fuzzy AND)
// -----------------------------------------------------------------------

/// Product t-norm: a * b. All conditions contribute proportionally.
pub fn tProduct(a: f64, b: f64) f64 {
    return a * b;
}

/// Minimum t-norm (Zadeh): min(a, b). Dominated by the weakest condition.
pub fn tMin(a: f64, b: f64) f64 {
    return @min(a, b);
}

/// Łukasiewicz t-norm: max(0, a + b - 1). Very strict.
pub fn tLukasiewicz(a: f64, b: f64) f64 {
    return @max(0.0, a + b - 1.0);
}

// -----------------------------------------------------------------------
// S-norms (fuzzy OR)
// -----------------------------------------------------------------------

/// Probabilistic sum: a + b - a*b. Dual of the product t-norm.
pub fn sProbabilistic(a: f64, b: f64) f64 {
    return a + b - a * b;
}

/// Maximum s-norm (Zadeh): max(a, b). Dual of the minimum t-norm.
pub fn sMax(a: f64, b: f64) f64 {
    return @max(a, b);
}

// -----------------------------------------------------------------------
// Negation
// -----------------------------------------------------------------------

/// Standard fuzzy negation: 1 - a.
pub fn fNot(a: f64) f64 {
    return 1.0 - a;
}

// -----------------------------------------------------------------------
// Variadic helpers
// -----------------------------------------------------------------------

/// Product t-norm over a slice of arguments.
/// Returns 1.0 for empty slice (identity element of product).
pub fn tProductAll(args: []const f64) f64 {
    var result: f64 = 1.0;
    for (args) |a| {
        result *= a;
    }
    return result;
}

/// Minimum t-norm over a slice of arguments.
/// Returns 1.0 for empty slice (identity element of min).
pub fn tMinAll(args: []const f64) f64 {
    if (args.len == 0) return 1.0;
    var result = args[0];
    for (args[1..]) |a| {
        if (a < result) result = a;
    }
    return result;
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

test "product basic" { try std.testing.expect(almostEqual(tProduct(0.8, 0.6), 0.48, 1e-10)); }
test "product identity" { try std.testing.expect(almostEqual(tProduct(0.7, 1.0), 0.7, 1e-10)); }
test "product annihilator" { try std.testing.expect(almostEqual(tProduct(0.7, 0.0), 0.0, 1e-10)); }
test "product commutativity" { try std.testing.expect(almostEqual(tProduct(0.3, 0.8), tProduct(0.8, 0.3), 1e-10)); }
test "min basic" { try std.testing.expectEqual(@as(f64, 0.6), tMin(0.8, 0.6)); }
test "min identity" { try std.testing.expectEqual(@as(f64, 0.7), tMin(0.7, 1.0)); }
test "min annihilator" { try std.testing.expectEqual(@as(f64, 0.0), tMin(0.7, 0.0)); }
test "lukasiewicz both high" { try std.testing.expect(almostEqual(tLukasiewicz(0.9, 0.8), 0.7, 1e-10)); }
test "lukasiewicz one low" { try std.testing.expect(almostEqual(tLukasiewicz(0.3, 0.5), 0.0, 1e-10)); }
test "lukasiewicz clamp" { try std.testing.expectEqual(@as(f64, 0.0), tLukasiewicz(0.1, 0.2)); }
test "lukasiewicz identity" { try std.testing.expect(almostEqual(tLukasiewicz(0.7, 1.0), 0.7, 1e-10)); }

test "probabilistic basic" { try std.testing.expect(almostEqual(sProbabilistic(0.8, 0.6), 0.92, 1e-10)); }
test "probabilistic identity" { try std.testing.expect(almostEqual(sProbabilistic(0.7, 0.0), 0.7, 1e-10)); }
test "probabilistic annihilator" { try std.testing.expect(almostEqual(sProbabilistic(0.7, 1.0), 1.0, 1e-10)); }
test "max basic" { try std.testing.expectEqual(@as(f64, 0.8), sMax(0.8, 0.6)); }
test "max identity" { try std.testing.expectEqual(@as(f64, 0.7), sMax(0.7, 0.0)); }

test "not basic" { try std.testing.expect(almostEqual(fNot(0.3), 0.7, 1e-10)); }
test "not zero" { try std.testing.expect(almostEqual(fNot(0.0), 1.0, 1e-10)); }
test "not one" { try std.testing.expect(almostEqual(fNot(1.0), 0.0, 1e-10)); }
test "double negation" { try std.testing.expect(almostEqual(fNot(fNot(0.4)), 0.4, 1e-10)); }

test "product all three" { try std.testing.expect(almostEqual(tProductAll(&[_]f64{ 0.8, 0.6, 0.5 }), 0.24, 1e-10)); }
test "product all single" { try std.testing.expect(almostEqual(tProductAll(&[_]f64{0.7}), 0.7, 1e-10)); }
test "product all empty" { try std.testing.expect(almostEqual(tProductAll(&[_]f64{}), 1.0, 1e-10)); }
test "min all three" { try std.testing.expectEqual(@as(f64, 0.6), tMinAll(&[_]f64{ 0.8, 0.6, 0.9 })); }
test "min all empty" { try std.testing.expectEqual(@as(f64, 1.0), tMinAll(&[_]f64{})); }
test "product all five" {
    const result = tProductAll(&[_]f64{ 0.9, 0.9, 0.9, 0.9, 0.9 });
    const expected = 0.9 * 0.9 * 0.9 * 0.9 * 0.9;
    try std.testing.expect(almostEqual(result, expected, 1e-10));
}

test "product/probabilistic De Morgan" {
    const a: f64 = 0.7;
    const b: f64 = 0.4;
    try std.testing.expect(almostEqual(tProduct(a, b), fNot(sProbabilistic(fNot(a), fNot(b))), 1e-10));
}

test "min/max De Morgan" {
    const a: f64 = 0.7;
    const b: f64 = 0.4;
    try std.testing.expect(almostEqual(tMin(a, b), fNot(sMax(fNot(a), fNot(b))), 1e-10));
}
