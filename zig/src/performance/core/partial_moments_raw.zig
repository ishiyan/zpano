//! Streaming raw partial moments (sums about zero).

const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;

/// Errors returned by `RawPartialMoments.revert`.
pub const Error = error{EmptyRevert};

/// Streaming raw low/high partial moments.
///
/// Construct with `RawPartialMoments{}`; no allocation.
pub const RawPartialMoments = struct {
    num: usize = 0,
    num_pos: usize = 0,
    num_neg: usize = 0,
    lpm_kbn: KleinKBNAccumulator = .{},
    hpm_kbn: KleinKBNAccumulator = .{},
    pos_kbn: KleinKBNAccumulator = .{},
    neg_kbn: KleinKBNAccumulator = .{},

    const Self = @This();

    pub fn reset(self: *Self) void {
        self.num = 0;
        self.num_pos = 0;
        self.num_neg = 0;
        self.lpm_kbn.reset();
        self.hpm_kbn.reset();
        self.pos_kbn.reset();
        self.neg_kbn.reset();
    }

    /// Removes a previously added return.
    ///
    /// Returns `error.EmptyRevert` when empty (Python has no check; its
    /// counters would go negative). Reverting a value that was never added
    /// is illegal.
    pub fn revert(self: *Self, ret: f64) Error!void {
        if (self.num == 0) return error.EmptyRevert;
        self.num -= 1;
        // Lower partial moment
        var pm = -ret;
        if (pm < 0) pm = 0;
        self.lpm_kbn.revert(pm);

        // Higher partial moment
        pm = ret;
        if (pm < 0) pm = 0;
        self.hpm_kbn.revert(pm);

        if (ret > 0) {
            self.num_pos -= 1;
            self.pos_kbn.revert(ret);
        } else if (ret < 0) {
            self.num_neg -= 1;
            self.neg_kbn.revert(ret);
        }
    }

    /// Adds a return.
    pub fn update(self: *Self, ret: f64) void {
        self.num += 1;
        // Lower partial moment
        var pm = -ret;
        if (pm < 0) pm = 0;
        self.lpm_kbn.update(pm);

        // Higher partial moment
        pm = ret;
        if (pm < 0) pm = 0;
        self.hpm_kbn.update(pm);

        if (ret > 0) {
            self.num_pos += 1;
            self.pos_kbn.update(ret);
        } else if (ret < 0) {
            self.num_neg += 1;
            self.neg_kbn.update(ret);
        }
    }

    /// Number of returns.
    pub fn count(self: *const Self) usize {
        return self.num;
    }

    /// sum(max(-r, 0)) (0.0 when empty).
    pub fn lowerPartialMoment1(self: *const Self) f64 {
        return self.lpm_kbn.value();
    }

    /// sum(max(r, 0)) (0.0 when empty).
    pub fn higherPartialMoment1(self: *const Self) f64 {
        return self.hpm_kbn.value();
    }

    /// Number of negative returns.
    pub fn countNegative(self: *const Self) usize {
        return self.num_neg;
    }

    /// Sum of negative returns.
    pub fn sumNegative(self: *const Self) f64 {
        return self.neg_kbn.value();
    }

    /// Number of positive returns.
    pub fn countPositive(self: *const Self) usize {
        return self.num_pos;
    }

    /// Sum of positive returns.
    pub fn sumPositive(self: *const Self) f64 {
        return self.pos_kbn.value();
    }
};

// ── Tests (TestRawPartialMoments from test_partial_moments.py) ─────────────

fn expectAlmostEqual(expected: f64, actual: f64, places: u5) !void {
    if (expected == actual) return;
    const tol = 0.5 * math.pow(f64, 10.0, -@as(f64, @floatFromInt(places)));
    if (!(@abs(expected - actual) <= tol)) {
        std.debug.print("expected {d}, got {d} (places {d})\n", .{ expected, actual, places });
        return error.TestExpectedApproxEq;
    }
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

/// Naive raw partial moments (sums, threshold 0), compared with `pm`.
fn assertMatchesRawReference(pm: *const RawPartialMoments, returns: []const f64, places: u5) !void {
    var lpm: f64 = 0;
    var hpm: f64 = 0;
    var cn: usize = 0;
    var sn: f64 = 0;
    var cp: usize = 0;
    var sp: f64 = 0;
    for (returns) |r| {
        lpm += @max(-r, 0.0);
        hpm += @max(r, 0.0);
        if (r < 0) {
            cn += 1;
            sn += r;
        }
        if (r > 0) {
            cp += 1;
            sp += r;
        }
    }
    try testing.expectEqual(returns.len, pm.count());
    try expectAlmostEqual(lpm, pm.lowerPartialMoment1(), places);
    try expectAlmostEqual(hpm, pm.higherPartialMoment1(), places);
    try testing.expectEqual(cn, pm.countNegative());
    try expectAlmostEqual(sn, pm.sumNegative(), places);
    try testing.expectEqual(cp, pm.countPositive());
    try expectAlmostEqual(sp, pm.sumPositive(), places);
}

test "empty" {
    try assertMatchesRawReference(&RawPartialMoments{}, &.{}, 14);
}

test "matches reference" {
    const returns = testRandomReturns(42, 200, 0.0);
    var pm = RawPartialMoments{};
    for (returns) |r| pm.update(r);
    try assertMatchesRawReference(&pm, &returns, 14);
}

test "rolling window matches reference" {
    const returns = testRandomReturns(7, 120, 0.0);
    const w = 8;
    var pm = RawPartialMoments{};
    for (returns, 0..) |r, i| {
        if (i >= w) try pm.revert(returns[i - w]);
        pm.update(r);
        try assertMatchesRawReference(&pm, returns[(if (i + 1 > w) i + 1 - w else 0) .. i + 1], 14);
    }
}

test "reset" {
    var pm = RawPartialMoments{};
    for ([_]f64{ 0.01, -0.02 }) |r| pm.update(r);
    pm.reset();
    try assertMatchesRawReference(&pm, &.{}, 14);
}

test "revert empty" {
    var pm = RawPartialMoments{};
    try testing.expectError(error.EmptyRevert, pm.revert(0.01));
}
