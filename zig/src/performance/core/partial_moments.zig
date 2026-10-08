//! Streaming partial moments about a threshold.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const RawMomentsKleinKBN = @import("raw_moments_klein_kbn").RawMomentsKleinKBN;

/// Errors returned by `PartialMoments.revert`.
pub const Error = error{EmptyRevert};

/// Streaming low/high partial moments.
///
/// Construct with `PartialMoments.init(threshold)`; no allocation.
pub const PartialMoments = struct {
    /// Target return or minimum acceptable return (MAR) in same periodicity as returns.
    threshold: f64,
    count_total: usize = 0,

    // ddof=1, bias=true, fisher=true matches scipy's default behavior for kurtosis.
    upper_excess_kbn: RawMomentsKleinKBN = .{ .ddof = 1, .bias = true, .fisher = true },
    lower_excess_kbn: RawMomentsKleinKBN = .{ .ddof = 1, .bias = true, .fisher = true },
    lpm_kbn: RawMomentsKleinKBN = .{ .ddof = 1, .bias = true, .fisher = true },
    hpm_kbn: RawMomentsKleinKBN = .{ .ddof = 1, .bias = true, .fisher = true },

    const Self = @This();

    /// Streaming low/high partial moments.
    ///
    /// Args:
    ///     threshold: Target return or minimum acceptable return (MAR) in same periodicity as returns
    pub fn init(threshold: f64) Self {
        return .{ .threshold = threshold };
    }

    pub fn reset(self: *Self) void {
        self.count_total = 0;
        self.upper_excess_kbn.reset();
        self.lower_excess_kbn.reset();
        self.lpm_kbn.reset();
        self.hpm_kbn.reset();
    }

    /// Removes a previously added return.
    ///
    /// Returns `error.EmptyRevert` when empty (Python raises from the
    /// underlying moments).
    pub fn revert(self: *Self, ret: f64) Error!void {
        if (self.count_total == 0) return error.EmptyRevert;
        self.count_total -= 1;
        // Lower partial moments for the raw returns less target return
        var pm = self.threshold - ret;
        if (pm < 0) {
            try self.upper_excess_kbn.revert(-pm);
            pm = 0;
        }
        try self.lpm_kbn.revert(pm);

        // Higher partial moments for the raw returns less required return
        pm = ret - self.threshold;
        if (pm < 0) {
            try self.lower_excess_kbn.revert(-pm);
            pm = 0;
        }
        try self.hpm_kbn.revert(pm);
    }

    /// Adds a return.
    pub fn update(self: *Self, ret: f64) void {
        self.count_total += 1;
        // Lower partial moments for the raw returns less target return
        var pm = self.threshold - ret;
        if (pm < 0) {
            self.upper_excess_kbn.update(-pm);
            pm = 0;
        }
        self.lpm_kbn.update(pm);

        // Higher partial moments for the raw returns less required return
        pm = ret - self.threshold;
        if (pm < 0) {
            self.lower_excess_kbn.update(-pm);
            pm = 0;
        }
        self.hpm_kbn.update(pm);
    }

    /// mean(max(threshold - r, 0)) (NaN when empty).
    pub fn lowerPartialMoment1(self: *const Self) f64 {
        return self.lpm_kbn.x1();
    }

    /// mean(max(threshold - r, 0)²) (NaN when empty).
    pub fn lowerPartialMoment2(self: *const Self) f64 {
        return self.lpm_kbn.x2();
    }

    /// mean(max(threshold - r, 0)³) (NaN when empty).
    pub fn lowerPartialMoment3(self: *const Self) f64 {
        return self.lpm_kbn.x3();
    }

    /// mean(max(threshold - r, 0)⁴) (NaN when empty).
    pub fn lowerPartialMoment4(self: *const Self) f64 {
        return self.lpm_kbn.x4();
    }

    /// mean(max(r - threshold, 0)) (NaN when empty).
    pub fn higherPartialMoment1(self: *const Self) f64 {
        return self.hpm_kbn.x1();
    }

    /// mean(max(r - threshold, 0)²) (NaN when empty).
    pub fn higherPartialMoment2(self: *const Self) f64 {
        return self.hpm_kbn.x2();
    }

    /// mean(max(r - threshold, 0)³) (NaN when empty).
    pub fn higherPartialMoment3(self: *const Self) f64 {
        return self.hpm_kbn.x3();
    }

    /// mean(max(r - threshold, 0)⁴) (NaN when empty).
    pub fn higherPartialMoment4(self: *const Self) f64 {
        return self.hpm_kbn.x4();
    }

    /// Proportion of returns below threshold (NaN when empty).
    pub fn downsideFrequency(self: *const Self) f64 {
        const total = self.count_total;
        if (total == 0) return math.nan(f64);
        return @as(f64, @floatFromInt(self.lower_excess_kbn.n())) / @as(f64, @floatFromInt(total));
    }

    /// Proportion of returns above threshold (NaN when empty).
    pub fn upsideFrequency(self: *const Self) f64 {
        const total = self.count_total;
        if (total == 0) return math.nan(f64);
        return @as(f64, @floatFromInt(self.upper_excess_kbn.n())) / @as(f64, @floatFromInt(total));
    }

    /// Mean of lower partial moments (also called shortfall); 0.0 when empty.
    pub fn downsidePotential(self: *const Self) f64 {
        return self.lpm_kbn.mean();
    }

    /// Number of returns.
    pub fn totalCount(self: *const Self) usize {
        return self.count_total;
    }

    /// Number of returns above threshold.
    pub fn upperExcessCount(self: *const Self) usize {
        return self.upper_excess_kbn.n();
    }

    /// Number of returns below threshold.
    pub fn lowerExcessCount(self: *const Self) usize {
        return self.lower_excess_kbn.n();
    }

    /// mean(r - threshold) over returns above threshold (NaN when none).
    pub fn upperExcessMoment1(self: *const Self) f64 {
        return self.upper_excess_kbn.x1();
    }

    /// sum(r - threshold) over returns above threshold.
    pub fn upperExcessMoment1Sum(self: *const Self) f64 {
        return self.upper_excess_kbn.x1Sum();
    }

    /// mean((r - threshold)²) over returns above threshold (NaN when none).
    pub fn upperExcessMoment2(self: *const Self) f64 {
        return self.upper_excess_kbn.x2();
    }

    /// sum((r - threshold)²) over returns above threshold.
    pub fn upperExcessMoment2Sum(self: *const Self) f64 {
        return self.upper_excess_kbn.x2Sum();
    }

    /// mean((r - threshold)³) over returns above threshold (NaN when none).
    pub fn upperExcessMoment3(self: *const Self) f64 {
        return self.upper_excess_kbn.x3();
    }

    /// mean((r - threshold)⁴) over returns above threshold (NaN when none).
    pub fn upperExcessMoment4(self: *const Self) f64 {
        return self.upper_excess_kbn.x4();
    }

    /// mean(threshold - r) over returns below threshold (NaN when none).
    pub fn lowerExcessMoment1(self: *const Self) f64 {
        return self.lower_excess_kbn.x1();
    }

    /// mean((threshold - r)²) over returns below threshold (NaN when none).
    pub fn lowerExcessMoment2(self: *const Self) f64 {
        return self.lower_excess_kbn.x2();
    }

    /// sum((threshold - r)²) over returns below threshold.
    pub fn lowerExcessMoment2Sum(self: *const Self) f64 {
        return self.lower_excess_kbn.x2Sum();
    }

    /// mean((threshold - r)³) over returns below threshold (NaN when none).
    pub fn lowerExcessMoment3(self: *const Self) f64 {
        return self.lower_excess_kbn.x3();
    }

    /// mean((threshold - r)⁴) over returns below threshold (NaN when none).
    pub fn lowerExcessMoment4(self: *const Self) f64 {
        return self.lower_excess_kbn.x4();
    }
};

// ── Tests (TestPartialMoments; TestRawPartialMoments is in partial_moments_raw.zig) ──

fn expectAlmostEqual(expected: f64, actual: f64, places: u5) !void {
    if (expected == actual) return;
    const tol = 0.5 * math.pow(f64, 10.0, -@as(f64, @floatFromInt(places)));
    if (!(@abs(expected - actual) <= tol)) {
        std.debug.print("expected {d}, got {d} (places {d})\n", .{ expected, actual, places });
        return error.TestExpectedApproxEq;
    }
}

fn expectNanOrAlmostEqual(expected: f64, actual: f64, places: u5) !void {
    if (math.isNan(expected)) return testing.expect(math.isNan(actual));
    try expectAlmostEqual(expected, actual, places);
}

/// Mixes in values exactly equal to the threshold and to zero.
fn testRandomReturns(seed: u64, comptime n: usize, threshold: f64) [n]f64 {
    var prng = std.Random.DefaultPrng.init(seed);
    const rng = prng.random();
    var out: [n]f64 = undefined;
    for (&out) |*r| {
        const choices = [_]f64{ threshold, 0.0, rng.floatNorm(f64) * 0.03, rng.floatNorm(f64) * 0.03 };
        r.* = choices[rng.uintLessThan(usize, choices.len)];
    }
    return out;
}

fn mean(values: []const f64) f64 {
    if (values.len == 0) return math.nan(f64);
    var s: f64 = 0;
    for (values) |v| s += v;
    return s / @as(f64, @floatFromInt(values.len));
}

/// Naive partial moments about the threshold, compared with `pm`.
fn assertMatchesReference(pm: *const PartialMoments, returns: []const f64, threshold: f64, places: u5) !void {
    var lower_buf: [256]f64 = undefined;
    var upper_buf: [256]f64 = undefined;
    var nl: usize = 0;
    var nu: usize = 0;
    for (returns) |r| {
        if (r < threshold) {
            lower_buf[nl] = threshold - r;
            nl += 1;
        }
        if (r > threshold) {
            upper_buf[nu] = r - threshold;
            nu += 1;
        }
    }
    const lower = lower_buf[0..nl];
    const upper = upper_buf[0..nu];
    const n = returns.len;
    const nf: f64 = @floatFromInt(n);

    try testing.expectEqual(n, pm.totalCount());
    try testing.expectEqual(nl, pm.lowerExcessCount());
    try testing.expectEqual(nu, pm.upperExcessCount());
    try expectNanOrAlmostEqual(if (n > 0) @as(f64, @floatFromInt(nl)) / nf else math.nan(f64), pm.downsideFrequency(), places);
    try expectNanOrAlmostEqual(if (n > 0) @as(f64, @floatFromInt(nu)) / nf else math.nan(f64), pm.upsideFrequency(), places);

    var buf: [256]f64 = undefined;
    for (returns, 0..) |r, i| buf[i] = @max(threshold - r, 0.0);
    try expectNanOrAlmostEqual(mean(buf[0..n]), pm.downsidePotential(), places);

    var s1: f64 = 0;
    var s2: f64 = 0;
    for (upper) |x| {
        s1 += x;
        s2 += x * x;
    }
    var l2: f64 = 0;
    for (lower) |x| l2 += x * x;
    try expectAlmostEqual(s1, pm.upperExcessMoment1Sum(), places);
    try expectAlmostEqual(s2, pm.upperExcessMoment2Sum(), places);
    try expectAlmostEqual(l2, pm.lowerExcessMoment2Sum(), places);

    const lpm = [_]*const fn (*const PartialMoments) f64{ PartialMoments.lowerPartialMoment1, PartialMoments.lowerPartialMoment2, PartialMoments.lowerPartialMoment3, PartialMoments.lowerPartialMoment4 };
    const hpm = [_]*const fn (*const PartialMoments) f64{ PartialMoments.higherPartialMoment1, PartialMoments.higherPartialMoment2, PartialMoments.higherPartialMoment3, PartialMoments.higherPartialMoment4 };
    const uem = [_]*const fn (*const PartialMoments) f64{ PartialMoments.upperExcessMoment1, PartialMoments.upperExcessMoment2, PartialMoments.upperExcessMoment3, PartialMoments.upperExcessMoment4 };
    const lem = [_]*const fn (*const PartialMoments) f64{ PartialMoments.lowerExcessMoment1, PartialMoments.lowerExcessMoment2, PartialMoments.lowerExcessMoment3, PartialMoments.lowerExcessMoment4 };
    for (0..4) |ki| {
        const k: f64 = @floatFromInt(ki + 1);
        for (returns, 0..) |r, i| buf[i] = math.pow(f64, @max(threshold - r, 0.0), k);
        try expectNanOrAlmostEqual(mean(buf[0..n]), lpm[ki](pm), places);
        for (returns, 0..) |r, i| buf[i] = math.pow(f64, @max(r - threshold, 0.0), k);
        try expectNanOrAlmostEqual(mean(buf[0..n]), hpm[ki](pm), places);
        for (upper, 0..) |x, i| buf[i] = math.pow(f64, x, k);
        try expectNanOrAlmostEqual(mean(buf[0..nu]), uem[ki](pm), places);
        for (lower, 0..) |x, i| buf[i] = math.pow(f64, x, k);
        try expectNanOrAlmostEqual(mean(buf[0..nl]), lem[ki](pm), places);
    }
}

test "empty" {
    const pm = PartialMoments.init(0.01);
    try testing.expectEqual(@as(usize, 0), pm.totalCount());
    try testing.expect(math.isNan(pm.downsideFrequency()));
    try testing.expect(math.isNan(pm.upsideFrequency()));
    try testing.expect(math.isNan(pm.lowerPartialMoment2()));
}

test "hand computed" {
    var pm = PartialMoments.init(0.01);
    for ([_]f64{ 0.03, -0.01, 0.01, 0.00 }) |r| pm.update(r);
    // Shortfalls below 1%: 0.02, 0.01; excesses above: 0.02.
    try expectAlmostEqual((0.02 + 0.01) / 4.0, pm.lowerPartialMoment1(), 16);
    try expectAlmostEqual((0.0004 + 0.0001) / 4.0, pm.lowerPartialMoment2(), 16);
    try expectAlmostEqual(0.02 / 4.0, pm.higherPartialMoment1(), 16);
    try testing.expectEqual(@as(usize, 2), pm.lowerExcessCount());
    try testing.expectEqual(@as(usize, 1), pm.upperExcessCount());
    try testing.expectEqual(@as(f64, 0.5), pm.downsideFrequency());
    try testing.expectEqual(@as(f64, 0.25), pm.upsideFrequency());
}

test "matches reference" {
    for ([_]f64{ 0.0, 0.005 }) |threshold| {
        const returns = testRandomReturns(42, 200, threshold);
        var pm = PartialMoments.init(threshold);
        for (returns) |r| pm.update(r);
        try assertMatchesReference(&pm, &returns, threshold, 14);
    }
}

test "rolling window matches reference" {
    const threshold = 0.005;
    const returns = testRandomReturns(7, 120, threshold);
    const w = 8;
    var pm = PartialMoments.init(threshold);
    for (returns, 0..) |r, i| {
        if (i >= w) try pm.revert(returns[i - w]);
        pm.update(r);
        try assertMatchesReference(&pm, returns[(if (i + 1 > w) i + 1 - w else 0) .. i + 1], threshold, 13);
    }
}

test "reset" {
    var pm = PartialMoments.init(0.0);
    for ([_]f64{ 0.01, -0.02 }) |r| pm.update(r);
    pm.reset();
    try testing.expectEqual(@as(usize, 0), pm.totalCount());
    try testing.expectEqual(@as(usize, 0), pm.lowerExcessCount());
    try testing.expectEqual(@as(usize, 0), pm.upperExcessCount());
}

test "revert empty" {
    var pm = PartialMoments.init(0.0);
    try testing.expectError(error.EmptyRevert, pm.revert(0.01));
}
