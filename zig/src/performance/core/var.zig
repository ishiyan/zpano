//! Value at Risk (VaR): historical, Gaussian and Cornish-Fisher.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const RawMomentsKleinKBN = @import("raw_moments_klein_kbn").RawMomentsKleinKBN;

const percentile_mod = @import("percentile.zig");
const norm = @import("norm.zig");

/// Errors returned by `varHistorical`: `error.InvalidArgument` when
/// 1 - confidence is outside [0, 1] (Python ValueError from percentile),
/// or `error.OutOfMemory`.
pub const HistoricalError = percentile_mod.Error;

/// Errors returned by `varGaussian` and `varCornishFisher`:
/// `error.InvalidArgument` when 1 - confidence is not in (0, 1)
/// (Python ValueError from norm_ppf).
pub const ParametricError = norm.Error;

/// Historical VaR: the negated (1 - confidence) percentile of the excess
/// returns r - risk_free_rate (NumPy linear interpolation).
///
/// NaN when `returns` is empty. `returns` is not modified; a scratch copy
/// is allocated with `allocator` and freed before returning.
pub fn varHistorical(allocator: std.mem.Allocator, returns: []const f64, risk_free_rate: f64, confidence: f64) HistoricalError!f64 {
    const w = returns;
    if (w.len < 1) return math.nan(f64);
    const q = 1 - confidence;
    const values = try allocator.dupe(f64, w);
    defer allocator.free(values);
    if (risk_free_rate != 0) {
        for (values) |*v| v.* = v.* - risk_free_rate;
    }
    return -(try percentile_mod.percentileInPlace(values, q));
}

/// Gaussian VaR: -(mean + z·σ) with z = Φ⁻¹(1 - confidence) and the
/// population standard deviation σ. NaN when σ is NaN (empty moments).
pub fn varGaussian(returns_kbn: *const RawMomentsKleinKBN, confidence: f64) ParametricError!f64 {
    const mean = returns_kbn.mean();
    const std_ = returns_kbn.standardDeviationDdof0();
    if (math.isNan(std_)) return math.nan(f64);
    const z = try norm.normPpf(1 - confidence);
    return -(mean + z * std_);
}

/// Cornish-Fisher (modified) VaR: Gaussian VaR with the z-score adjusted
/// for the population skewness and excess kurtosis (bias=True,
/// fisher=True). Falls back to Gaussian VaR when skewness or kurtosis are
/// unavailable (very small samples). NaN when σ is NaN.
pub fn varCornishFisher(returns_kbn: *const RawMomentsKleinKBN, confidence: f64) ParametricError!f64 {
    const mean = returns_kbn.mean();
    const std_ = returns_kbn.standardDeviationDdof0();
    if (math.isNan(std_)) return math.nan(f64);
    // Cornish-Fisher expansion for z-score adjustment
    var z = try norm.normPpf(1 - confidence);
    const skew = returns_kbn.skewnessMoment(); // bias=True
    const kurtosis = returns_kbn.kurtosisExcess(); // bias=True, fisher=True
    // Skewness and kurtosis are unavailable for very small samples.
    // Fall back to Gaussian VaR.
    if (math.isNan(skew) or math.isNan(kurtosis)) return -(mean + z * std_);
    // Cornish-Fisher expansion
    const z2 = z * z;
    const z3 = z2 * z;
    z = (z + (z2 - 1) * skew / 6 + (z3 - 3 * z) * kurtosis / 24 - (2 * z3 - 5 * z) * skew * skew / 36);
    return -(mean + z * std_);
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

test "historical quantile and risk free rate" {
    const alloc = testing.allocator;
    const returns = [_]f64{ -0.2, -0.1, 0.0, 0.1 };
    try expectAlmostEqual(0.125, try varHistorical(alloc, &returns, 0.0, 0.75), 15);
    try expectAlmostEqual(0.135, try varHistorical(alloc, &returns, 0.01, 0.75), 15);
    try testing.expectEqualSlices(f64, &.{ -0.2, -0.1, 0.0, 0.1 }, &returns);
}

test "empty historical input" {
    try testing.expect(math.isNan(try varHistorical(testing.allocator, &.{}, 0.0, 0.95)));
}

test "cornish fisher falls back for one sample" {
    var moments = RawMomentsKleinKBN{};
    moments.update(0.02);
    try expectAlmostEqual(-0.02, try varGaussian(&moments, 0.95), 15);
    try testing.expectEqual(try varGaussian(&moments, 0.95), try varCornishFisher(&moments, 0.95));
}
