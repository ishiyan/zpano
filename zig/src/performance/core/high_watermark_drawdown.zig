//! Rolling high-water-mark drawdown.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const Allocator = std.mem.Allocator;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;
const FifoBuffer = @import("fifo_buffer.zig").FifoBuffer;

/// Rolling high-water-mark drawdown.
///
/// Drawdown at each observation is measured from the high-water mark,
/// the highest equity value reached up to that observation within the
/// current rolling window, including the equity at the start of the window:
///
///     drawdown_t = equity_t / max(equity_start, equity_1, ..., equity_t) - 1
///
/// This matches R PerformanceAnalytics ``Drawdowns()``, which uses
/// ``cummax(c(1, cumprod(1 + R)))``: a first negative return already
/// produces a drawdown.  For a rolling window, ``equity_start`` is the
/// equity just before the first observation in the window, so the result
/// equals a fresh calculation over the window's returns.
///
/// Drawdowns are expressed as decimals and are non-positive.
///
/// Cumulative log-equity is maintained internally so that returns can be
/// accumulated accurately. Running sums of drawdowns and squared drawdowns
/// are maintained using compensated floating-point accumulation.
///
/// When an observation leaves the window, the window's starting equity
/// changes.  The drawdowns in the window are recomputed only when this
/// changes their high-water marks; otherwise, the update is O(1).
///
/// A window size of zero (or negative) means an expanding (unbounded) window.
///
/// Holds growable state: create with `init(allocator, window_size)`, free
/// with `deinit()`.
pub const HighWaterMarkDrawdown = struct {
    allocator: Allocator,
    /// Window size, 0 for an expanding window.
    window_size: usize,
    /// Cumulative log-equity at each observation.
    cumlog: FifoBuffer(f64) = .{},
    /// Drawdown at each observation, as a decimal (<= 0).
    dd: FifoBuffer(f64) = .{},
    /// Cumulative log return.
    c: KleinKBNAccumulator = .{},
    /// Log-equity just before the first observation in the window.
    base: f64 = 0.0,
    /// Current high-water mark in log-equity space.
    peak: f64 = 0.0,
    /// Running drawdown aggregates.
    sum_dd: KleinKBNAccumulator = .{},
    sum_dd2: KleinKBNAccumulator = .{},

    const Self = @This();

    /// Creates an empty tracker. `window_size <= 0` means expanding.
    pub fn init(allocator: Allocator, window_size: i64) Self {
        return .{
            .allocator = allocator,
            .window_size = if (window_size > 0) @intCast(window_size) else 0,
        };
    }

    pub fn deinit(self: *Self) void {
        self.cumlog.deinit(self.allocator);
        self.dd.deinit(self.allocator);
    }

    /// Reset the accumulator to its initial empty state.
    pub fn reset(self: *Self) void {
        self.cumlog.clear();
        self.dd.clear();
        self.sum_dd.reset();
        self.sum_dd2.reset();
        self.c.reset();
        self.base = 0.0;
        self.peak = 0.0;
    }

    /// Recompute all drawdowns from the cumulative log-equity values.
    ///
    /// This is required when an observation leaving the rolling window
    /// changes the high-water marks of the remaining observations.
    /// Never allocates: capacity for all values is already reserved.
    fn recompute(self: *Self) void {
        self.dd.clear();
        self.sum_dd.reset();
        self.sum_dd2.reset();
        var peak = self.base;
        for (self.cumlog.slice()) |c| {
            var d: f64 = undefined;
            if (c >= peak) {
                peak = c;
                d = 0.0;
            } else {
                d = math.expm1(c - peak);
            }
            self.dd.list.appendAssumeCapacity(d);
            self.sum_dd.update(d);
            self.sum_dd2.update(d * d);
        }
        self.peak = peak;
    }

    /// Add a return observation.
    ///
    /// If the rolling window is full, the oldest observation is removed
    /// before the new observation is added.
    ///
    /// Args:
    ///     ret: Period return expressed as a decimal. For example, ``0.02``
    ///         represents a 2% return and ``-0.015`` a -1.5% return.
    ///
    /// Returns true if the rolling window required a drawdown
    /// recomputation, otherwise false; `error.OutOfMemory` if storage could
    /// not grow (the state is then unchanged).
    pub fn update(self: *Self, ret: f64) Allocator.Error!bool {
        // Reserve space first so a failed allocation leaves the state intact.
        try self.cumlog.ensureUnusedCapacity(self.allocator, 1);
        try self.dd.ensureUnusedCapacity(self.allocator, 1);

        var old_base: ?f64 = null;
        if (self.window_size != 0 and self.cumlog.len() == self.window_size) {
            const old_c = self.cumlog.popFront();
            const old_dd = self.dd.popFront();

            self.sum_dd.revert(old_dd);
            self.sum_dd2.revert(old_dd * old_dd);

            // The evicted observation's equity is the new starting equity.
            // High-water marks of the remaining observations can only
            // change if the old starting equity was above the evicted one.
            if (old_c < self.base) old_base = self.base;
            self.base = old_c;
        }

        // Global cumulative log-equity.
        self.c.update(math.log1p(ret));
        const c = self.c.value();
        self.cumlog.list.appendAssumeCapacity(c);

        // Peaks of all remaining observations were max(old_base, c0, ..., cj);
        // without old_base they are max(c0, c1, ..., cj).  They differ only
        // if the new first observation is also below old_base.
        if (old_base) |ob| {
            if (self.cumlog.slice()[0] < ob) {
                self.recompute();
                return true;
            }
        }

        var d: f64 = undefined;
        if (c >= self.peak) {
            self.peak = c;
            d = 0.0;
        } else {
            d = math.expm1(c - self.peak);
        }

        self.dd.list.appendAssumeCapacity(d);
        self.sum_dd.update(d);
        self.sum_dd2.update(d * d);
        return false;
    }

    /// Drawdowns for observations currently in the window, oldest first.
    ///
    /// The slice points into internal storage (no copy); it is valid until
    /// the next call to `update`, `reset` or `deinit`. Do not modify.
    pub fn drawdowns(self: *const Self) []const f64 {
        return self.dd.slice();
    }

    /// The most recent drawdown in the current window (NaN when empty).
    pub fn drawdown(self: *const Self) f64 {
        const s = self.dd.slice();
        return if (s.len > 0) s[s.len - 1] else math.nan(f64);
    }

    /// Maximum drawdown in the current window (NaN when empty).
    ///
    /// Drawdowns are non-positive, the largest loss is the minimum
    /// drawdown value.
    pub fn maximumDrawdown(self: *const Self) f64 {
        const s = self.dd.slice();
        if (s.len == 0) return math.nan(f64);
        var m = s[0];
        for (s[1..]) |x| {
            if (x < m) m = x;
        }
        return m;
    }

    /// Arithmetic mean of drawdowns in the current window (NaN when empty).
    pub fn drawdownsMean(self: *const Self) f64 {
        const n = self.dd.len();
        return if (n != 0) self.sum_dd.value() / @as(f64, @floatFromInt(n)) else math.nan(f64);
    }

    /// Mean squared drawdown in the current window (NaN when empty).
    pub fn drawdownsSquaredMean(self: *const Self) f64 {
        const n = self.dd.len();
        return if (n != 0) self.sum_dd2.value() / @as(f64, @floatFromInt(n)) else math.nan(f64);
    }

    /// Number of observations in the current window.
    pub fn drawdownsCount(self: *const Self) usize {
        return self.dd.len();
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

/// Independent reference implementation of chronological high-water-mark
/// drawdowns, matching R PerformanceAnalytics Drawdowns(): the high-water
/// mark starts at the initial equity 1. Writes into `out`.
fn expectedDrawdowns(returns: []const f64, out: []f64) []f64 {
    var equity: f64 = 1.0;
    var peak: f64 = 1.0;
    for (returns, 0..) |ret, i| {
        equity *= 1.0 + ret;
        if (equity >= peak) {
            peak = equity;
            out[i] = 0.0;
        } else {
            out[i] = equity / peak - 1.0;
        }
    }
    return out[0..returns.len];
}

/// Drawdowns of the rolling window ending at index i: a fresh calculation
/// over the returns in the window.
fn expectedRollingDrawdowns(returns: []const f64, window_size: usize, i: usize, out: []f64) []f64 {
    const lo = if (window_size == 0) 0 else (if (i + 1 > window_size) i + 1 - window_size else 0);
    return expectedDrawdowns(returns[lo .. i + 1], out);
}

fn assertState(acc: *const HighWaterMarkDrawdown, expected: []const f64, places: u5) !void {
    const actual = acc.drawdowns();
    try testing.expectEqual(expected.len, actual.len);
    for (actual, expected) |a, e| try expectAlmostEqual(e, a, places);
    try testing.expectEqual(expected.len, acc.drawdownsCount());
    if (expected.len > 0) {
        try expectAlmostEqual(expected[expected.len - 1], acc.drawdown(), places);
        var mn = expected[0];
        var s: f64 = 0;
        var s2: f64 = 0;
        for (expected) |x| {
            mn = @min(mn, x);
            s += x;
            s2 += x * x;
        }
        const n: f64 = @floatFromInt(expected.len);
        try expectAlmostEqual(mn, acc.maximumDrawdown(), places);
        try expectAlmostEqual(s / n, acc.drawdownsMean(), places);
        try expectAlmostEqual(s2 / n, acc.drawdownsSquaredMean(), places);
    } else {
        try testing.expect(math.isNan(acc.drawdown()));
        try testing.expect(math.isNan(acc.maximumDrawdown()));
        try testing.expect(math.isNan(acc.drawdownsMean()));
        try testing.expect(math.isNan(acc.drawdownsSquaredMean()));
    }
}

fn feed(acc: *HighWaterMarkDrawdown, returns: []const f64) !void {
    for (returns) |r| _ = try acc.update(r);
}

// Expanding-window tests

test "expanding empty" {
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 0);
    defer acc.deinit();
    try assertState(&acc, &.{}, 14);
}

test "expanding all positive returns" {
    // Every observation creates a new high-water mark.
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 0);
    defer acc.deinit();
    try feed(&acc, &.{ 0.10, 0.05, 0.20 });
    try assertState(&acc, &.{ 0.0, 0.0, 0.0 }, 14);
}

test "expanding first negative return" {
    // The high-water mark starts at the initial equity 1.0, so the first
    // negative return is already a drawdown (as in R).
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 0);
    defer acc.deinit();
    try feed(&acc, &.{ -0.05, -0.02, 0.10 });
    try assertState(&acc, &.{ -0.05, -0.069, 0.0 }, 14);
}

test "expanding simple drawdown and recovery" {
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 0);
    defer acc.deinit();
    try feed(&acc, &.{ 0.10, -0.05, 0.10 });
    try assertState(&acc, &.{ 0.0, -0.05, 0.0 }, 14);
}

test "expanding compounded drawdown" {
    // Relative to the 1.10 peak: 0.891 / 1.10 - 1 = -19%
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 0);
    defer acc.deinit();
    try feed(&acc, &.{ 0.10, -0.10, -0.10 });
    try assertState(&acc, &.{ 0.0, -0.10, -0.19 }, 14);
}

test "expanding new high water mark resets drawdown" {
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 0);
    defer acc.deinit();
    try feed(&acc, &.{ 0.10, -0.05, 0.06, -0.02 });
    try assertState(&acc, &.{ 0.0, -0.05, 0.0, -0.02 }, 14);
}

test "expanding reset" {
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 0);
    defer acc.deinit();
    try feed(&acc, &.{ 0.10, -0.05, -0.02 });
    try testing.expect(acc.drawdownsCount() > 0);
    acc.reset();
    try assertState(&acc, &.{}, 14);

    // The accumulator can be reused, starting from equity 1.0 again.
    _ = try acc.update(-0.05);
    try assertState(&acc, &.{-0.05}, 14);
}

test "expanding matches reference" {
    var prng = std.Random.DefaultPrng.init(42);
    const rng = prng.random();
    var returns: [200]f64 = undefined;
    for (&returns) |*r| r.* = rng.floatNorm(f64) * 0.03;
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 0);
    defer acc.deinit();
    try feed(&acc, &returns);
    var buf: [200]f64 = undefined;
    try assertState(&acc, expectedDrawdowns(&returns, &buf), 13);
}

test "zero size means expanding" {
    const returns = [_]f64{ 0.10, -0.05, -0.02, 0.05 };
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 0);
    defer acc.deinit();
    try feed(&acc, &returns);
    var buf: [4]f64 = undefined;
    try assertState(&acc, expectedDrawdowns(&returns, &buf), 14);
}

test "negative size means expanding" {
    // Non-positive window sizes are normalized to zero (expanding mode).
    const returns = [_]f64{ 0.10, -0.05, -0.02 };
    var acc = HighWaterMarkDrawdown.init(testing.allocator, -10);
    defer acc.deinit();
    try feed(&acc, &returns);
    var buf: [3]f64 = undefined;
    try assertState(&acc, expectedDrawdowns(&returns, &buf), 14);
}

// Rolling-window tests: the window equals a fresh calculation over its
// returns, starting from the equity just before the window.

test "rolling window peak eviction" {
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 3);
    defer acc.deinit();
    try feed(&acc, &.{ 0.10, -0.05, -0.02 });
    try assertState(&acc, &.{ 0.0, -0.05, -0.069 }, 14);
    _ = try acc.update(0.03);
    try assertState(&acc, &.{ -0.05, -0.069, -0.04107 }, 14);
}

test "rolling window evicted peak followed by new peak" {
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 3);
    defer acc.deinit();
    try feed(&acc, &.{ 0.05, -0.02, 0.10 });
    try assertState(&acc, &.{ 0.0, -0.02, 0.0 }, 14);
    _ = try acc.update(-0.03);
    try assertState(&acc, &.{ -0.02, 0.0, -0.03 }, 14);
}

test "rolling window peak eviction recomputes drawdowns" {
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 3);
    defer acc.deinit();
    try feed(&acc, &.{ 0.10, -0.05, -0.05 });
    try assertState(&acc, &.{ 0.0, -0.05, -0.0975 }, 14);
    _ = try acc.update(0.01);
    try assertState(&acc, &.{ -0.05, -0.0975, -0.088475 }, 14);
}

test "rolling window all negative returns" {
    const returns = [_]f64{ -0.01, -0.02, -0.03, -0.04 };
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 3);
    defer acc.deinit();
    var buf: [4]f64 = undefined;
    for (returns, 0..) |ret, i| {
        _ = try acc.update(ret);
        try assertState(&acc, expectedRollingDrawdowns(&returns, 3, i, &buf), 14);
    }
}

test "rolling window size one" {
    const returns = [_]f64{ 0.10, -0.05, -0.02, 0.03, -0.04 };
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 1);
    defer acc.deinit();
    for (returns) |ret| {
        _ = try acc.update(ret);
        try assertState(&acc, &.{@min(ret, 0.0)}, 14);
    }
}

test "rolling window matches fresh calculation" {
    var prng = std.Random.DefaultPrng.init(7);
    const rng = prng.random();
    for ([_]usize{ 2, 3, 4, 7, 20 }) |window_size| {
        for (0..10) |_| {
            var returns: [80]f64 = undefined;
            for (&returns) |*r| r.* = rng.floatNorm(f64) * 0.03;
            var acc = HighWaterMarkDrawdown.init(testing.allocator, @intCast(window_size));
            defer acc.deinit();
            var buf: [80]f64 = undefined;
            for (returns, 0..) |ret, i| {
                _ = try acc.update(ret);
                try assertState(&acc, expectedRollingDrawdowns(&returns, window_size, i, &buf), 13);
            }
        }
    }
}

test "rolling window recompute flag" {
    // Evicting an observation can only change the remaining high-water
    // marks if its return was negative (its equity is below the window's
    // starting equity).
    var acc = HighWaterMarkDrawdown.init(testing.allocator, 2);
    defer acc.deinit();
    try testing.expect(!try acc.update(0.10));
    try testing.expect(!try acc.update(-0.05));
    // Evicts +10%: the remaining peaks don't change.
    try testing.expect(!try acc.update(-0.02));
    // Evicts -5% and the new first observation (-2%) is also below
    // the old starting equity: recompute.
    try testing.expect(try acc.update(0.01));
    try assertState(&acc, &.{ -0.02, -0.0102 }, 14);
    // Evicts -2%; the new first observation (equity 1.0343) is still
    // below the old starting equity (1.045): recompute.
    try testing.expect(try acc.update(0.03));
    try assertState(&acc, &.{ 0.0, 0.0 }, 14);
    // Evicts +1%: its equity is above the window's starting equity,
    // so the remaining peaks don't change.
    try testing.expect(!try acc.update(-0.01));
    try assertState(&acc, &.{ 0.0, -0.01 }, 14);
}
