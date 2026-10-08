//! Streaming cumulative (geometric) returns.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;

/// Errors returned by `CumulativeReturn.revert`.
pub const Error = error{EmptyRevert};

/// Streaming cumulative (geometric) returns.
///
/// Accumulates the sum of log returns, log(1 + r), with compensated
/// summation.  Because only a sum is stored, revert() may remove any
/// previously added return, so the type works for FIFO rolling windows:
/// the caller owns the window and feeds evicted returns to revert().
///
/// Construct with `CumulativeReturn{}`; no allocation.
pub const CumulativeReturn = struct {
    cumlogret_sum: KleinKBNAccumulator = .{},
    num: usize = 0,

    const Self = @This();

    pub fn reset(self: *Self) void {
        self.cumlogret_sum.reset();
        self.num = 0;
    }

    /// Removes a previously added return.
    ///
    /// Returns `error.EmptyRevert` if there are no returns (Python ValueError).
    pub fn revert(self: *Self, ret: f64) Error!void {
        if (self.num <= 0) return error.EmptyRevert;
        self.num -= 1;
        self.cumlogret_sum.revert(if (ret != 0) math.log1p(ret) else 0);
    }

    /// Adds a return, expressed as a decimal (must be > -1).
    pub fn update(self: *Self, ret: f64) void {
        self.num += 1;
        self.cumlogret_sum.update(if (ret != 0) math.log1p(ret) else 0);
    }

    /// The number of returns.
    pub fn count(self: *const Self) usize {
        return self.num;
    }

    /// Cumulative geometric return, prod(1 + r) - 1 (0.0 when empty).
    pub fn cumulativeGeometricReturn(self: *const Self) f64 {
        return math.expm1(self.cumlogret_sum.value());
    }

    /// The geometric mean of the returns, prod(1 + r)^(1/n) - 1
    /// (NaN when empty).
    pub fn geometricMeanReturn(self: *const Self) f64 {
        if (self.num == 0) return math.nan(f64);
        return math.expm1(self.cumlogret_sum.value() / @as(f64, @floatFromInt(self.num)));
    }

    /// The annualized geometric mean, prod(1 + r)^(periods_per_year/n) - 1
    /// (NaN when empty).
    pub fn annualizedGeometricMeanReturn(self: *const Self, periods_per_year: f64) f64 {
        if (self.num == 0) return math.nan(f64);
        return math.expm1(self.cumlogret_sum.value() * periods_per_year / @as(f64, @floatFromInt(self.num)));
    }
};

// ── Tests ──────────────────────────────────────────────────────────────────

fn expectAlmostEqual(expected: f64, actual: f64, places: u5) !void {
    if (expected == actual) return;
    const tol = 0.5 * math.pow(f64, 10.0, -@as(f64, @floatFromInt(places)));
    if (!(@abs(expected - actual) <= tol)) {
        std.debug.print("expected {d}, got {d} (places {d})\n", .{ expected, actual, places });
        return error.TestExpectedApproxEq;
    }
}

fn gauss(r: std.Random, mu: f64, sigma: f64) f64 {
    return mu + sigma * r.floatNorm(f64);
}

test "annualized return definition" {
    const returns = [_]f64{ 0.10, -0.05, 0.03, 0.08 };
    const periods_per_year = 12;
    var cr = CumulativeReturn{};
    for (returns) |r| cr.update(r);
    var growth: f64 = 1;
    for (returns) |r| growth *= 1 + r;
    const expected = math.pow(f64, growth, @as(f64, periods_per_year) / @as(f64, returns.len)) - 1;
    try expectAlmostEqual(expected, cr.annualizedGeometricMeanReturn(periods_per_year), 15);
}

test "one month return" {
    // One monthly return of 1%, expected (1.01)^12-1.
    var cr = CumulativeReturn{};
    cr.update(0.01);
    const expected = math.pow(f64, 1.01, 12) - 1;
    try expectAlmostEqual(expected, cr.annualizedGeometricMeanReturn(12), 15);
}

test "yearly returns" {
    // If the observations are yearly, the annualized geometric mean return
    // equals the geometric mean return.
    var cr = CumulativeReturn{};
    for ([_]f64{ 0.12, -0.04, 0.08 }) |r| cr.update(r);
    try expectAlmostEqual(cr.geometricMeanReturn(), cr.annualizedGeometricMeanReturn(1), 15);
}

test "constant monthly return" {
    // ((1+r)^n)^(12/n) = (1+r)^12: the number of observations cancels.
    var cr = CumulativeReturn{};
    for (0..60) |_| cr.update(0.01);
    const expected = math.pow(f64, 1.01, 12) - 1;
    try expectAlmostEqual(expected, cr.annualizedGeometricMeanReturn(12), 15);
}

test "empty" {
    const cr = CumulativeReturn{};
    try testing.expect(math.isNan(cr.annualizedGeometricMeanReturn(12)));
}

test "zero returns" {
    // log1p(0) == 0 and expm1(0) == 0, so the result is exactly zero.
    var cr = CumulativeReturn{};
    for (0..100) |_| cr.update(0.0);
    try expectAlmostEqual(0, cr.annualizedGeometricMeanReturn(252), 15);
}

test "consistency with geometric mean return" {
    // 1 + annualized = (1 + geometric mean)^p
    const returns = [_]f64{ 0.0010, -0.0005, 0.0003, 0.0008 };
    const periods_per_year = 252;
    var cr = CumulativeReturn{};
    for (returns) |r| cr.update(r);
    const expected = math.pow(f64, 1 + cr.geometricMeanReturn(), periods_per_year) - 1;
    try expectAlmostEqual(expected, cr.annualizedGeometricMeanReturn(periods_per_year), 13);
}

test "rolling window matches fresh calculation" {
    var prng = std.Random.DefaultPrng.init(42);
    const rng = prng.random();
    var returns: [100]f64 = undefined;
    for (&returns) |*r| {
        const g = gauss(rng, 0.0, 0.03);
        r.* = if (rng.boolean()) 0.0 else g;
    }
    const window_size = 7;
    var cr = CumulativeReturn{};
    for (returns, 0..) |r, i| {
        if (i >= window_size) try cr.revert(returns[i - window_size]);
        cr.update(r);
        const window = returns[(if (i + 1 > window_size) i + 1 - window_size else 0) .. i + 1];
        var growth: f64 = 1;
        for (window) |x| growth *= 1 + x;
        try testing.expectEqual(window.len, cr.count());
        try expectAlmostEqual(growth - 1, cr.cumulativeGeometricReturn(), 14);
        try expectAlmostEqual(math.pow(f64, growth, 1 / @as(f64, @floatFromInt(window.len))) - 1, cr.geometricMeanReturn(), 14);
    }
}

test "revert empty raises" {
    var cr = CumulativeReturn{};
    try testing.expectError(error.EmptyRevert, cr.revert(0.01));
}

test "reset" {
    var cr = CumulativeReturn{};
    for ([_]f64{ 0.1, -0.2 }) |r| cr.update(r);
    cr.reset();
    try testing.expectEqual(@as(usize, 0), cr.count());
    try testing.expectEqual(@as(f64, 0.0), cr.cumulativeGeometricReturn());
    try testing.expect(math.isNan(cr.geometricMeanReturn()));
}
