//! Expected Shortfall (ES, CVaR): historical, Gaussian and Cornish-Fisher.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const RawMomentsKleinKBN = @import("raw_moments_klein_kbn").RawMomentsKleinKBN;

const norm = @import("norm.zig");
const var_mod = @import("var.zig");

/// Errors returned by `esHistorical` (see `var.HistoricalError`).
pub const HistoricalError = var_mod.HistoricalError;

/// Errors returned by `esGaussian` and `esCornishFisher`
/// (see `var.ParametricError`).
pub const ParametricError = var_mod.ParametricError;

/// Historical ES: the negated mean of the excess returns r - risk_free_rate
/// at or below the historical VaR threshold (excess <= -VaR).
///
/// NaN when `returns` is empty, VaR is NaN, or the tail is empty.
/// `allocator` is used for scratch space only.
pub fn esHistorical(allocator: std.mem.Allocator, returns: []const f64, risk_free_rate: f64, confidence: f64) HistoricalError!f64 {
    if (returns.len == 0) return math.nan(f64);
    const var_ = try var_mod.varHistorical(allocator, returns, risk_free_rate, confidence);
    if (math.isNan(var_)) return math.nan(f64);

    var sum_tail: f64 = 0.0;
    var count: usize = 0;
    for (returns) |r| {
        const excess = r - risk_free_rate;
        if (excess <= -var_) {
            sum_tail += excess;
            count += 1;
        }
    }

    return if (count != 0) -sum_tail / @as(f64, @floatFromInt(count)) else math.nan(f64);
}

/// Gaussian ES: -mean + φ(z)·σ / (1 - confidence) with z = Φ⁻¹(confidence)
/// and the population standard deviation σ. NaN when σ is NaN.
pub fn esGaussian(returns_kbn: *const RawMomentsKleinKBN, confidence: f64) ParametricError!f64 {
    const mean = returns_kbn.mean();
    const std_ = returns_kbn.standardDeviationDdof0();
    if (math.isNan(std_)) return math.nan(f64);
    const z = try norm.normPpf(confidence);
    const phi_z = norm.normPdf(z);
    return -mean + phi_z * std_ / (1 - confidence);
}

/// Cornish-Fisher (modified) ES using population skewness and excess
/// kurtosis (bias=True, fisher=True). Falls back to Gaussian ES when
/// skewness or kurtosis are unavailable (very small samples).
pub fn esCornishFisher(returns_kbn: *const RawMomentsKleinKBN, confidence: f64) ParametricError!f64 {
    const alpha = 1.0 - confidence;
    const z = try norm.normPpf(alpha);
    const mean = returns_kbn.mean();
    const sigma = returns_kbn.standardDeviationDdof0();
    const skew = returns_kbn.skewnessMoment(); // bias=True
    const kurtosis = returns_kbn.kurtosisExcess(); // bias=True, fisher=True
    // Skewness and kurtosis are unavailable for very small samples.
    // Fall back to Gaussian ES.
    if (math.isNan(skew) or math.isNan(kurtosis)) return esGaussian(returns_kbn, confidence);
    const z2 = z * z;
    const z3 = z2 * z;
    const h = (z + (z2 - 1) * skew / 6 + (z3 - 3 * z) * kurtosis / 24 - (2 * z3 - 5 * z) * skew * skew / 36);
    const h2 = h * h;
    const h4 = h2 * h2;
    const mes = (norm.normPdf(h) * (1 + h2 * h * skew / 6 + (h4 * h2 - 9 * h4 + 9 * h2 + 3) * skew * skew / 72 + (h4 - 2 * h2 - 1) * kurtosis / 24));
    // Python min(): keeps the first argument unless the second is smaller.
    const tail = -mes / alpha;
    return -mean - sigma * (if (h < tail) h else tail);
}

// ── Tests (from test_risk_helpers.py) ──────────────────────────────────────

fn expectAlmostEqual(expected: f64, actual: f64, places: u5) !void {
    if (expected == actual) return;
    const tol = 0.5 * math.pow(f64, 10.0, -@as(f64, @floatFromInt(places)));
    if (!(@abs(expected - actual) <= tol)) {
        std.debug.print("expected {d}, got {d} (places {d})\n", .{ expected, actual, places });
        return error.TestExpectedApproxEq;
    }
}

test "historical tail and risk free rate" {
    const alloc = testing.allocator;
    const returns = [_]f64{ -0.2, -0.1, 0.0, 0.1 };
    try expectAlmostEqual(0.2, try esHistorical(alloc, &returns, 0.0, 0.75), 15);
    try expectAlmostEqual(0.21, try esHistorical(alloc, &returns, 0.01, 0.75), 15);
    try testing.expectEqualSlices(f64, &.{ -0.2, -0.1, 0.0, 0.1 }, &returns);
}

test "empty historical input" {
    try testing.expect(math.isNan(try esHistorical(testing.allocator, &.{}, 0.0, 0.95)));
}

test "cornish fisher falls back for one sample" {
    var moments = RawMomentsKleinKBN{};
    moments.update(0.02);
    try expectAlmostEqual(-0.02, try esGaussian(&moments, 0.95), 15);
    try testing.expectEqual(try esGaussian(&moments, 0.95), try esCornishFisher(&moments, 0.95));
}
