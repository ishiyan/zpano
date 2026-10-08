//! Streaming winning/losing return averages and counts.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNSummator = @import("klein_kbn_summator").KleinKBNSummator;

/// Errors returned by `WinLoss.revert`.
pub const Error = error{EmptyRevert};

/// Streaming winning/loosing return averages and counts.
///
/// Construct with `WinLoss{}`; no allocation.
pub const WinLoss = struct {
    non_zero_sum: KleinKBNSummator = .{},
    win_sum: KleinKBNSummator = .{},
    loss_sum: KleinKBNSummator = .{},

    const Self = @This();

    pub fn reset(self: *Self) void {
        self.non_zero_sum.reset();
        self.win_sum.reset();
        self.loss_sum.reset();
    }

    /// Removes a previously added return.
    ///
    /// Returns `error.EmptyRevert` if a non-zero return is reverted from an
    /// empty summator (as the Python summator raises).
    pub fn revert(self: *Self, ret: f64) Error!void {
        if (ret != 0) try self.non_zero_sum.revert(ret);
        if (ret > 0) try self.win_sum.revert(ret);
        if (ret < 0) try self.loss_sum.revert(ret);
    }

    /// Adds a return.
    pub fn update(self: *Self, ret: f64) void {
        if (ret != 0) self.non_zero_sum.update(ret);
        if (ret > 0) self.win_sum.update(ret);
        if (ret < 0) self.loss_sum.update(ret);
    }

    /// Arithmetic mean (average) of non-zero returns (NaN when none).
    pub fn nonZeroReturnsMean(self: *const Self) f64 {
        return self.non_zero_sum.mean();
    }

    /// The number of non-zero returns.
    pub fn nonZeroReturnsCount(self: *const Self) usize {
        return self.non_zero_sum.n();
    }

    /// Sum of winning (positive) returns.
    pub fn winningReturnsSum(self: *const Self) f64 {
        return self.win_sum.value();
    }

    /// Arithmetic mean (average) of winning (positive) returns (NaN when none).
    pub fn winningReturnsMean(self: *const Self) f64 {
        return self.win_sum.mean();
    }

    /// The number of winning (positive) returns.
    pub fn winningReturnsCount(self: *const Self) usize {
        return self.win_sum.n();
    }

    /// Sum of losing (negative) returns.
    pub fn losingReturnsSum(self: *const Self) f64 {
        return self.loss_sum.value();
    }

    /// Arithmetic mean (average) of losing (negative) returns (NaN when none).
    pub fn losingReturnsMean(self: *const Self) f64 {
        return self.loss_sum.mean();
    }

    /// The number of losing (negative) returns.
    pub fn losingReturnsCount(self: *const Self) usize {
        return self.loss_sum.n();
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

fn expectNanOrAlmostEqual(expected: f64, actual: f64, places: u5) !void {
    if (math.isNan(expected)) return testing.expect(math.isNan(actual));
    try expectAlmostEqual(expected, actual, places);
}

fn assertMatches(wl: *const WinLoss, returns: []const f64, places: u5) !void {
    var non_zero_n: usize = 0;
    var non_zero_s: f64 = 0;
    var win_n: usize = 0;
    var win_s: f64 = 0;
    var loss_n: usize = 0;
    var loss_s: f64 = 0;
    for (returns) |r| {
        if (r != 0) {
            non_zero_n += 1;
            non_zero_s += r;
        }
        if (r > 0) {
            win_n += 1;
            win_s += r;
        }
        if (r < 0) {
            loss_n += 1;
            loss_s += r;
        }
    }
    const mean = struct {
        fn f(s: f64, n: usize) f64 {
            return if (n > 0) s / @as(f64, @floatFromInt(n)) else math.nan(f64);
        }
    }.f;
    try testing.expectEqual(non_zero_n, wl.nonZeroReturnsCount());
    try expectNanOrAlmostEqual(mean(non_zero_s, non_zero_n), wl.nonZeroReturnsMean(), places);
    try testing.expectEqual(win_n, wl.winningReturnsCount());
    try expectAlmostEqual(win_s, wl.winningReturnsSum(), places);
    try expectNanOrAlmostEqual(mean(win_s, win_n), wl.winningReturnsMean(), places);
    try testing.expectEqual(loss_n, wl.losingReturnsCount());
    try expectAlmostEqual(loss_s, wl.losingReturnsSum(), places);
    try expectNanOrAlmostEqual(mean(loss_s, loss_n), wl.losingReturnsMean(), places);
}

test "empty" {
    try assertMatches(&WinLoss{}, &.{}, 15);
}

test "hand computed" {
    var wl = WinLoss{};
    for ([_]f64{ 0.02, 0.0, -0.01, 0.04, 0.0, -0.03 }) |r| wl.update(r);
    try testing.expectEqual(@as(usize, 4), wl.nonZeroReturnsCount());
    try expectAlmostEqual(0.02 / 4.0, wl.nonZeroReturnsMean(), 16);
    try testing.expectEqual(@as(usize, 2), wl.winningReturnsCount());
    try expectAlmostEqual(0.03, wl.winningReturnsMean(), 16);
    try testing.expectEqual(@as(usize, 2), wl.losingReturnsCount());
    try expectAlmostEqual(-0.02, wl.losingReturnsMean(), 16);
}

test "rolling window matches reference" {
    var prng = std.Random.DefaultPrng.init(42);
    const rng = prng.random();
    var returns: [120]f64 = undefined;
    for (&returns) |*r| {
        const g = gauss(rng, 0.0, 0.03);
        r.* = if (rng.boolean()) 0.0 else g;
    }
    const w = 6;
    var wl = WinLoss{};
    for (returns, 0..) |r, i| {
        if (i >= w) try wl.revert(returns[i - w]);
        wl.update(r);
        try assertMatches(&wl, returns[(if (i + 1 > w) i + 1 - w else 0) .. i + 1], 15);
    }
}

test "reset" {
    var wl = WinLoss{};
    for ([_]f64{ 0.01, -0.02 }) |r| wl.update(r);
    wl.reset();
    try assertMatches(&wl, &.{}, 15);
}
