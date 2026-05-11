/// Membership functions for fuzzy logic.
///
/// Each function maps a crisp value to a membership degree μ ∈ [0, 1].
/// Two shapes are supported: sigmoid (default, smooth) and linear (trapezoidal ramp).
/// All functions degrade to crisp step functions when width = 0.
const std = @import("std");

/// Shape of the fuzzy membership transition curve.
pub const MembershipShape = enum(u8) {
    /// Smooth logistic curve. Default for most applications.
    sigmoid = 0,
    /// Piecewise-linear ramp (trapezoidal/triangular).
    linear = 1,
};

/// Steepness constant for sigmoid shape.
/// k = sigmoid_k / width gives ≈0.997 at threshold ± width/2.
const sigmoid_k: f64 = 12.0;

/// Logistic sigmoid: 1 / (1 + exp(k * (x - threshold))).
///
/// Returns the "less-than" membership: high when x << threshold,
/// low when x >> threshold, exactly 0.5 at x == threshold.
fn sigmoid(x: f64, threshold: f64, k: f64) f64 {
    const exponent = k * (x - threshold);
    // Clamp to avoid overflow in exp().
    if (exponent > 500.0) return 0.0;
    if (exponent < -500.0) return 1.0;
    return 1.0 / (1.0 + @exp(exponent));
}

/// Degree to which x is less than threshold.
///
/// At threshold: μ = 0.5.
/// When width = 0 (crisp): 1.0 if x < threshold, 0.5 if x == threshold, 0.0 if x > threshold.
pub fn muLess(x: f64, threshold: f64, width: f64, shape: MembershipShape) f64 {
    if (width <= 0.0) {
        if (x < threshold) return 1.0;
        if (x > threshold) return 0.0;
        return 0.5;
    }

    if (shape == .linear) {
        const half = width * 0.5;
        if (x <= threshold - half) return 1.0;
        if (x >= threshold + half) return 0.0;
        return (threshold + half - x) / width;
    }
    // sigmoid
    return sigmoid(x, threshold, sigmoid_k / width);
}

/// Degree to which x ≤ threshold.
/// Identical to muLess for continuous values — the distinction is conceptual.
pub fn muLessEqual(x: f64, threshold: f64, width: f64, shape: MembershipShape) f64 {
    return muLess(x, threshold, width, shape);
}

/// Degree to which x > threshold. Complement of muLess.
pub fn muGreater(x: f64, threshold: f64, width: f64, shape: MembershipShape) f64 {
    return 1.0 - muLess(x, threshold, width, shape);
}

/// Degree to which x ≥ threshold. Complement of muLessEqual.
pub fn muGreaterEqual(x: f64, threshold: f64, width: f64, shape: MembershipShape) f64 {
    return 1.0 - muLessEqual(x, threshold, width, shape);
}

/// Bell-shaped membership: degree to which x ≈ target.
///
/// μ = 1.0 at x == target.
/// μ ≈ 0 at |x - target| ≥ width.
pub fn muNear(x: f64, target: f64, width: f64, shape: MembershipShape) f64 {
    if (width <= 0.0) {
        return if (x == target) 1.0 else 0.0;
    }

    if (shape == .linear) {
        const dist = @abs(x - target);
        if (dist >= width) return 0.0;
        return 1.0 - dist / width;
    }
    // sigmoid → Gaussian bell
    // σ chosen so that μ ≈ 0.003 at |x - target| = width.
    const s = width / 2.41;
    const d = (x - target) / s;
    return @exp(-d * d);
}

/// Fuzzy candle direction ∈ [-1, +1].
///
/// +1 = fully bullish. 0 = neutral. -1 = fully bearish.
/// Uses tanh(steepness * (c - o) / body_avg).
/// When body_avg ≤ 0: returns +1.0 if c ≥ o, else -1.0 (crisp).
pub fn muDirection(o: f64, c: f64, body_avg: f64, steepness: f64) f64 {
    if (body_avg <= 0.0) {
        return if (c >= o) 1.0 else -1.0;
    }
    return std.math.tanh(steepness * (c - o) / body_avg);
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

test "muLess crossover at threshold" {
    try std.testing.expect(almostEqual(muLess(10.0, 10.0, 2.0, .sigmoid), 0.5, 1e-10));
}

test "muLess well below threshold" {
    try std.testing.expect(muLess(8.0, 10.0, 2.0, .sigmoid) > 0.99);
}

test "muLess well above threshold" {
    try std.testing.expect(muLess(12.0, 10.0, 2.0, .sigmoid) < 0.01);
}

test "muLess monotonically decreasing" {
    const xs = [_]f64{ 8.0, 9.0, 10.0, 11.0, 12.0 };
    for (0..xs.len - 1) |i| {
        try std.testing.expect(muLess(xs[i], 10.0, 2.0, .sigmoid) > muLess(xs[i + 1], 10.0, 2.0, .sigmoid));
    }
}

test "muLess symmetry" {
    const below = muLess(9.0, 10.0, 2.0, .sigmoid);
    const above = muLess(11.0, 10.0, 2.0, .sigmoid);
    try std.testing.expect(almostEqual(below + above, 1.0, 1e-10));
}

test "muLess linear crossover" {
    try std.testing.expect(almostEqual(muLess(10.0, 10.0, 4.0, .linear), 0.5, 1e-10));
}

test "muLess linear below range" {
    try std.testing.expectEqual(@as(f64, 1.0), muLess(7.0, 10.0, 4.0, .linear));
}

test "muLess linear above range" {
    try std.testing.expectEqual(@as(f64, 0.0), muLess(13.0, 10.0, 4.0, .linear));
}

test "muLess linear midpoint" {
    try std.testing.expect(almostEqual(muLess(9.0, 10.0, 4.0, .linear), 0.75, 1e-10));
}

test "muLess crisp below" {
    try std.testing.expectEqual(@as(f64, 1.0), muLess(9.0, 10.0, 0.0, .sigmoid));
}

test "muLess crisp above" {
    try std.testing.expectEqual(@as(f64, 0.0), muLess(11.0, 10.0, 0.0, .sigmoid));
}

test "muLess crisp at threshold" {
    try std.testing.expectEqual(@as(f64, 0.5), muLess(10.0, 10.0, 0.0, .sigmoid));
}

test "muLessEqual same as muLess" {
    try std.testing.expectEqual(muLess(9.5, 10.0, 2.0, .sigmoid), muLessEqual(9.5, 10.0, 2.0, .sigmoid));
}

test "muGreater complement of muLess" {
    for ([_]f64{ 8.0, 9.0, 10.0, 11.0, 12.0 }) |x| {
        const sum = muGreater(x, 10.0, 2.0, .sigmoid) + muLess(x, 10.0, 2.0, .sigmoid);
        try std.testing.expect(almostEqual(sum, 1.0, 1e-10));
    }
}

test "muGreater crossover" {
    try std.testing.expect(almostEqual(muGreater(10.0, 10.0, 2.0, .sigmoid), 0.5, 1e-10));
}

test "muGreater well above" {
    try std.testing.expect(muGreater(12.0, 10.0, 2.0, .sigmoid) > 0.99);
}

test "muGreater well below" {
    try std.testing.expect(muGreater(8.0, 10.0, 2.0, .sigmoid) < 0.01);
}

test "muGreaterEqual complement" {
    const sum = muGreaterEqual(9.5, 10.0, 2.0, .sigmoid) + muLessEqual(9.5, 10.0, 2.0, .sigmoid);
    try std.testing.expect(almostEqual(sum, 1.0, 1e-10));
}

test "muNear peak at target" {
    try std.testing.expect(almostEqual(muNear(10.0, 10.0, 2.0, .sigmoid), 1.0, 1e-10));
}

test "muNear falls off" {
    try std.testing.expect(muNear(12.0, 10.0, 2.0, .sigmoid) < 0.05);
}

test "muNear symmetric" {
    const below = muNear(9.0, 10.0, 2.0, .sigmoid);
    const above = muNear(11.0, 10.0, 2.0, .sigmoid);
    try std.testing.expect(almostEqual(below, above, 1e-10));
}

test "muNear monotonic from center" {
    const offsets = [_]f64{ 0, 0.5, 1.0, 1.5, 2.0 };
    for (0..offsets.len - 1) |i| {
        try std.testing.expect(muNear(10.0 + offsets[i], 10.0, 2.0, .sigmoid) >
            muNear(10.0 + offsets[i + 1], 10.0, 2.0, .sigmoid));
    }
}

test "muNear linear peak" {
    try std.testing.expect(almostEqual(muNear(10.0, 10.0, 2.0, .linear), 1.0, 1e-10));
}

test "muNear linear at boundary" {
    try std.testing.expectEqual(@as(f64, 0.0), muNear(12.0, 10.0, 2.0, .linear));
}

test "muNear linear midpoint" {
    try std.testing.expect(almostEqual(muNear(11.0, 10.0, 2.0, .linear), 0.5, 1e-10));
}

test "muNear crisp exact" {
    try std.testing.expectEqual(@as(f64, 1.0), muNear(10.0, 10.0, 0.0, .sigmoid));
}

test "muNear crisp any distance" {
    try std.testing.expectEqual(@as(f64, 0.0), muNear(10.1, 10.0, 0.0, .sigmoid));
}

test "muDirection large white body" {
    try std.testing.expect(muDirection(100.0, 110.0, 5.0, 2.0) > 0.95);
}

test "muDirection large black body" {
    try std.testing.expect(muDirection(110.0, 100.0, 5.0, 2.0) < -0.95);
}

test "muDirection doji" {
    try std.testing.expect(almostEqual(muDirection(100.0, 100.0, 5.0, 2.0), 0.0, 1e-10));
}

test "muDirection tiny white" {
    const d = muDirection(100.0, 100.1, 5.0, 2.0);
    try std.testing.expect(d > 0.0 and d < 0.1);
}

test "muDirection antisymmetric" {
    const d1 = muDirection(100.0, 105.0, 5.0, 2.0);
    const d2 = muDirection(105.0, 100.0, 5.0, 2.0);
    try std.testing.expect(almostEqual(d1, -d2, 1e-10));
}

test "muDirection zero body avg white" {
    try std.testing.expectEqual(@as(f64, 1.0), muDirection(100.0, 101.0, 0.0, 2.0));
}

test "muDirection zero body avg black" {
    try std.testing.expectEqual(@as(f64, -1.0), muDirection(101.0, 100.0, 0.0, 2.0));
}

test "muDirection zero body avg doji" {
    try std.testing.expectEqual(@as(f64, 1.0), muDirection(100.0, 100.0, 0.0, 2.0));
}

test "muDirection range bounded" {
    const cases = [_][3]f64{ .{ 0, 1000, 1 }, .{ 1000, 0, 1 }, .{ 50, 50, 100 } };
    for (cases) |case| {
        const d = muDirection(case[0], case[1], case[2], 2.0);
        try std.testing.expect(d >= -1.0 and d <= 1.0);
    }
}

test "very large x" {
    try std.testing.expectEqual(@as(f64, 0.0), muLess(1e10, 0.0, 1.0, .sigmoid));
}

test "very small x" {
    try std.testing.expectEqual(@as(f64, 1.0), muLess(-1e10, 0.0, 1.0, .sigmoid));
}

test "tiny width" {
    try std.testing.expect(muLess(9.999, 10.0, 0.001, .sigmoid) > 0.99);
}

test "huge width" {
    const val = muLess(0.0, 10.0, 1000.0, .sigmoid);
    try std.testing.expect(val > 0.49 and val < 0.60);
}
