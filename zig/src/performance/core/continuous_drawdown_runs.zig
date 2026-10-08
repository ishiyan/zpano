//! Streaming 'continuous' drawdown runs for Burke-type measures.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const Allocator = std.mem.Allocator;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;
const FifoBuffer = @import("fifo_buffer.zig").FifoBuffer;

/// Errors returned by `ContinuousDrawdownRuns.revert`.
pub const Error = error{EmptyRevert};

/// Convert a compounded log return into a percentage drawdown.
pub fn ddPercent(logsum: f64) f64 {
    return math.expm1(logsum) * 100.0;
}

/// Python `max(x, 0.0)`: keeps x unless 0.0 is greater.
fn maxZero(x: f64) f64 {
    return if (0.0 > x) 0.0 else x;
}

/// Streaming 'continuous' drawdown runs for Burke-type measures.
///
/// A continuous drawdown is the compounded loss over a maximal run of
/// consecutive negative returns.  Following PerformanceAnalytics
/// ``BurkeRatio``, returns are compounded as if they were percentages:
///
///     DD = (prod(1 + r_i * 0.01) - 1) * 100
///
/// For decimal returns this is close to the sum of the run's returns,
/// not their compounded return; the quirk is kept to match R.
///
/// The Burke denominator is:
///
///     sqrt(sum(DD_j^2))
///
/// where the sum is taken over all continuous losing runs in the current
/// window.
///
/// This type is a pure accumulator: the caller owns the rolling window
/// and feeds evicted values to `revert()` and new values to `update()`.
/// Within one step, call `revert(old)` BEFORE `update(new)` so run
/// adjacency stays correct.
///
/// The complexity is amortized O(1) per call.
///
/// Holds growable state: create with `init(allocator)`, free with `deinit()`.
pub const ContinuousDrawdownRuns = struct {
    /// One losing run: compounded log return and number of returns.
    pub const Run = struct {
        logsum: f64,
        count: usize,
    };

    allocator: Allocator,
    runs: FifoBuffer(Run) = .{},
    /// Sum of squared continuous drawdowns.
    sum_sq: KleinKBNAccumulator = .{},
    /// Whether the most recent return is negative.
    last_was_negative: bool = false,

    const Self = @This();

    pub fn init(allocator: Allocator) Self {
        return .{ .allocator = allocator };
    }

    pub fn deinit(self: *Self) void {
        self.runs.deinit(self.allocator);
    }

    pub fn reset(self: *Self) void {
        self.runs.clear();
        self.sum_sq.reset();
        self.last_was_negative = false;
    }

    /// Remove the oldest return from the left edge of the window.
    ///
    /// Returns `error.EmptyRevert` if `old_ret` is negative but there are
    /// no runs (Python raises IndexError).
    pub fn revert(self: *Self, old_ret: f64) Error!void {
        if (old_ret < 0) {
            if (self.runs.len() == 0) return error.EmptyRevert;
            // oldest negative is the front of the left-most run, shrink it
            const run = self.runs.first();
            var d = ddPercent(run.logsum);
            self.sum_sq.revert(d * d);
            run.logsum -= math.log1p(old_ret * 0.01);
            run.count -= 1;
            if (run.count == 0) {
                _ = self.runs.popFront(); // run fully evicted
            } else {
                d = ddPercent(run.logsum);
                self.sum_sq.update(d * d);
            }
        }
        // old_ret >= 0 is a separator, nothing to update
    }

    /// Add a new (most-recent) return at the right edge of the window.
    ///
    /// May allocate (`error.OutOfMemory`) when a new run starts.
    pub fn update(self: *Self, ret: f64) Allocator.Error!void {
        if (ret < 0) {
            const logr = math.log1p(ret * 0.01);
            if (self.last_was_negative and self.runs.len() > 0) {
                // Extend the currently-open (right-most) run.
                const run = self.runs.last();
                var d = ddPercent(run.logsum);
                self.sum_sq.revert(d * d);
                run.logsum += logr;
                run.count += 1;
                d = ddPercent(run.logsum);
                self.sum_sq.update(d * d);
            } else {
                // Start a new run.
                try self.runs.append(self.allocator, .{ .logsum = logr, .count = 1 });
                const d = ddPercent(logr);
                self.sum_sq.update(d * d);
            }
            self.last_was_negative = true;
        } else {
            // Non-negative return closes any open run (already counted) — a separator.
            self.last_was_negative = false;
        }
    }

    /// Continuous drawdowns (negative percentages), one value for each
    /// losing run, oldest first.
    ///
    /// Returns a new slice owned by the caller (free with `allocator`).
    pub fn drawdowns(self: *const Self, allocator: Allocator) Allocator.Error![]f64 {
        const runs = self.runs.slice();
        const out = try allocator.alloc(f64, runs.len);
        for (runs, out) |r, *o| o.* = ddPercent(r.logsum);
        return out;
    }

    /// The runs currently in the window, oldest first. The slice points
    /// into internal storage and is valid until the next mutating call.
    pub fn runsSlice(self: *const Self) []const Run {
        return self.runs.slice();
    }

    /// Sum of squared continuous drawdowns.
    pub fn sumDrawdownsSquared(self: *const Self) f64 {
        return maxZero(self.sum_sq.value());
    }

    /// Square root of the sum of squared continuous drawdowns.
    ///
    /// This is the denominator used by the Burke ratio.
    pub fn sqrtSumDrawdownsSquared(self: *const Self) f64 {
        return @sqrt(maxZero(self.sum_sq.value()));
    }

    /// Number of continuous losing runs in the current window.
    pub fn runCount(self: *const Self) usize {
        return self.runs.len();
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

/// Compounded drawdown of one continuous losing run.
fn dd(returns: []const f64) f64 {
    var compounded: f64 = 1.0;
    for (returns) |r| compounded *= 1.0 + r * 0.01;
    return (compounded - 1.0) * 100.0;
}

fn assertState(acc: *const ContinuousDrawdownRuns, expected_drawdowns: []const f64, expected_run_count: ?usize, places: u5) !void {
    const actual = try acc.drawdowns(testing.allocator);
    defer testing.allocator.free(actual);
    try testing.expectEqual(expected_drawdowns.len, actual.len);
    for (actual, expected_drawdowns) |a, e| try expectAlmostEqual(e, a, places);
    if (expected_run_count) |c| try testing.expectEqual(c, acc.runCount());
    var expected_sum_sq: f64 = 0;
    for (expected_drawdowns) |x| expected_sum_sq += x * x;
    try expectAlmostEqual(expected_sum_sq, acc.sumDrawdownsSquared(), places);
    try expectAlmostEqual(@sqrt(expected_sum_sq), acc.sqrtSumDrawdownsSquared(), places);
}

fn feed(acc: *ContinuousDrawdownRuns, returns: []const f64) !void {
    for (returns) |r| try acc.update(r);
}

// Expanding-window tests

test "expanding single losing run" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try acc.update(-1.0);
    try assertState(&acc, &.{dd(&.{-1.0})}, 1, 12);
    try acc.update(-2.0);
    try assertState(&acc, &.{dd(&.{ -1.0, -2.0 })}, 1, 12);
    try acc.update(-3.0);
    try assertState(&acc, &.{dd(&.{ -1.0, -2.0, -3.0 })}, 1, 12);
}

test "expanding multiple losing runs" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ -1.0, -2.0, 1.0, -3.0, -4.0, 2.0, -5.0 });
    try assertState(&acc, &.{ dd(&.{ -1.0, -2.0 }), dd(&.{ -3.0, -4.0 }), dd(&.{-5.0}) }, 3, 12);
}

test "expanding non-negative returns are separators" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ -2.0, 0.0, -3.0, 1.0, -4.0 });
    try assertState(&acc, &.{ dd(&.{-2.0}), dd(&.{-3.0}), dd(&.{-4.0}) }, 3, 12);
}

test "expanding positive return does not create drawdown" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ 2.0, 3.0, 0.0, 5.0 });
    try assertState(&acc, &.{}, 0, 12);
}

test "reset" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ -1.0, -2.0, 1.0, -3.0 });
    try testing.expect(acc.runCount() > 0);

    acc.reset();
    try assertState(&acc, &.{}, 0, 12);

    // It must also be possible to use it again after reset.
    try acc.update(-4.0);
    try assertState(&acc, &.{dd(&.{-4.0})}, 1, 12);
}

// Rolling-window tests

test "rolling window eviction from front of losing run" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ -1.0, -2.0, -3.0 });
    try assertState(&acc, &.{dd(&.{ -1.0, -2.0, -3.0 })}, 1, 12);
    try acc.revert(-1.0);
    try acc.update(-4.0);
    try assertState(&acc, &.{dd(&.{ -2.0, -3.0, -4.0 })}, 1, 12);
}

test "rolling window eviction of entire losing run" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ -1.0, -2.0, 1.0 });
    try assertState(&acc, &.{dd(&.{ -1.0, -2.0 })}, 1, 12);
    try acc.revert(-1.0);
    try acc.update(-3.0);
    try assertState(&acc, &.{ dd(&.{-2.0}), dd(&.{-3.0}) }, 2, 12);
}

test "rolling window eviction of separator" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ -1.0, 1.0, -2.0 });
    try assertState(&acc, &.{ dd(&.{-1.0}), dd(&.{-2.0}) }, 2, 12);
    try acc.revert(-1.0);
    try acc.update(-3.0);
    try assertState(&acc, &.{dd(&.{ -2.0, -3.0 })}, 1, 12);
}

test "rolling window multiple runs" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ -1.0, -2.0, 1.0, -3.0, -4.0 });
    try assertState(&acc, &.{ dd(&.{ -1.0, -2.0 }), dd(&.{ -3.0, -4.0 }) }, 2, 12);

    // Slide 1: remove -1%, add +2%
    try acc.revert(-1.0);
    try acc.update(2.0);
    try assertState(&acc, &.{ dd(&.{-2.0}), dd(&.{ -3.0, -4.0 }) }, 2, 12);

    // Slide 2: remove -2%, add -5%
    try acc.revert(-2.0);
    try acc.update(-5.0);
    try assertState(&acc, &.{ dd(&.{ -3.0, -4.0 }), dd(&.{-5.0}) }, 2, 12);
}

test "rolling window new return extends existing run" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ 1.0, -2.0, -3.0 });
    try assertState(&acc, &.{dd(&.{ -2.0, -3.0 })}, 1, 12);
    try acc.revert(1.0);
    try acc.update(-4.0);
    try assertState(&acc, &.{dd(&.{ -2.0, -3.0, -4.0 })}, 1, 12);
}

test "rolling window new negative starts new run after separator" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ -2.0, 1.0, -3.0 });
    try assertState(&acc, &.{ dd(&.{-2.0}), dd(&.{-3.0}) }, 2, 12);
    try acc.revert(-2.0);
    try acc.update(-4.0);
    try assertState(&acc, &.{dd(&.{ -3.0, -4.0 })}, 1, 12);
}

test "revert then update order is required" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ -1.0, -2.0, 2.0 });
    try assertState(&acc, &.{dd(&.{ -1.0, -2.0 })}, 1, 12);

    // New window: [-2%, +2%, -3%]
    try acc.revert(-1.0);
    try acc.update(-3.0);
    try assertState(&acc, &.{ dd(&.{-2.0}), dd(&.{-3.0}) }, 2, 12);
}

// Numerical consistency

test "sqrt sum drawdowns squared" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try feed(&acc, &.{ -1.0, -2.0, 1.0, -3.0 });

    const dd1 = dd(&.{ -1.0, -2.0 });
    const dd2 = dd(&.{-3.0});

    const expected_sum_sq = dd1 * dd1 + dd2 * dd2;
    const expected_sqrt = @sqrt(expected_sum_sq);

    try expectAlmostEqual(expected_sum_sq, acc.sumDrawdownsSquared(), 12);
    try expectAlmostEqual(expected_sqrt, acc.sqrtSumDrawdownsSquared(), 12);
}

// Brute-force rolling-window test

fn referenceRuns(window: []const f64, out: []f64) usize {
    var n: usize = 0;
    var start: ?usize = null;
    for (window, 0..) |r, i| {
        if (r < 0) {
            if (start == null) start = i;
        } else if (start) |s| {
            out[n] = dd(window[s..i]);
            n += 1;
            start = null;
        }
    }
    if (start) |s| {
        out[n] = dd(window[s..]);
        n += 1;
    }
    return n;
}

test "rolling window matches fresh calculation" {
    // At every step, the runs equal those of the returns currently in the
    // window, computed from scratch. Returns are percentages, with exact
    // zeros mixed in as separators.
    var prng = std.Random.DefaultPrng.init(42);
    const rng = prng.random();
    for ([_]usize{ 1, 2, 3, 5, 12 }) |window_size| {
        var returns: [150]f64 = undefined;
        for (&returns) |*r| {
            const g = rng.floatNorm(f64) * 3.0;
            r.* = if (rng.boolean()) 0.0 else g;
        }
        var acc = ContinuousDrawdownRuns.init(testing.allocator);
        defer acc.deinit();
        for (returns, 0..) |ret, i| {
            if (i >= window_size) try acc.revert(returns[i - window_size]);
            try acc.update(ret);
            const window = returns[(if (i + 1 > window_size) i + 1 - window_size else 0) .. i + 1];
            var buf: [16]f64 = undefined;
            const n = referenceRuns(window, &buf);
            try assertState(&acc, buf[0..n], n, 12);
        }
    }
}

test "revert negative when empty" {
    var acc = ContinuousDrawdownRuns.init(testing.allocator);
    defer acc.deinit();
    try acc.revert(1.0); // separator: nothing to do
    try testing.expectError(error.EmptyRevert, acc.revert(-1.0));
}
