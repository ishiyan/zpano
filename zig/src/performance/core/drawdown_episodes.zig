//! Streaming high-water-mark drawdown episode tracker.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const Allocator = std.mem.Allocator;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;

/// A single high-water-mark drawdown episode.
pub const DrawdownEpisode = struct {
    /// Maximum drawdown depth, expressed as a decimal return.
    depth: f64,
    /// Index of the first underwater observation.
    from_idx: usize,
    /// Index of the deepest observation.
    trough_idx: usize,
    /// Recovery observation index if the episode recovered; otherwise the
    /// last observation index currently available.
    to_idx: usize,
    /// True if the high-water mark was recovered at `to_idx`. False if the
    /// series ended while the drawdown was still open.
    recovered: bool,
};

/// Streaming high-water-mark drawdown episode tracker.
///
/// The type consumes drawdown observations produced by
/// `HighWaterMarkDrawdown` and maintains the corresponding drawdown
/// episodes incrementally.
///
/// Drawdowns are expected as decimals, for example ``-0.025`` for a 2.5%
/// drawdown and ``0.0`` for an observation at a high-water mark.
///
/// Indices are positions in the sequence of observations passed to
/// `update()` since the last `reset()`/`recalculate()`.  For a rolling
/// window, call `recalculate()` with the window's drawdowns whenever they
/// are recomputed, so indices refer to the window.
///
/// A drawdown episode begins with the first negative drawdown and remains
/// open until a non-negative drawdown is observed.
///
/// The episode is considered recovered when the drawdown reaches zero or
/// becomes positive.
///
/// Holds growable state: create with `init(allocator)`, free with `deinit()`.
pub const DrawdownEpisodes = struct {
    allocator: Allocator,
    /// Closed (recovered) episodes.
    closed: std.ArrayList(DrawdownEpisode) = .empty,

    // Current open drawdown episode.
    current_from: ?usize = null,
    current_trough: usize = 0,
    current_depth: f64 = 0.0,

    /// Number of observations processed.
    count: usize = 0,

    // Running depth aggregates.
    sum_depth: KleinKBNAccumulator = .{},
    sum_depth_squared: KleinKBNAccumulator = .{},
    // Running episode length/peak-to-trough/recovery aggregates
    sum_length: usize = 0,
    sum_peak_to_trough: usize = 0,
    sum_recovery: usize = 0,

    const Self = @This();

    pub fn init(allocator: Allocator) Self {
        return .{ .allocator = allocator };
    }

    pub fn deinit(self: *Self) void {
        self.closed.deinit(self.allocator);
    }

    /// Reset the episode tracker to its initial empty state.
    pub fn reset(self: *Self) void {
        self.closed.clearRetainingCapacity();
        self.current_from = null;
        self.current_trough = 0;
        self.current_depth = 0.0;
        self.count = 0;
        self.sum_depth.reset();
        self.sum_depth_squared.reset();
        self.sum_length = 0;
        self.sum_peak_to_trough = 0;
        self.sum_recovery = 0;
    }

    /// Add one drawdown observation.
    ///
    /// Drawdowns must be non-positive, although non-negative values are
    /// accepted and treated as recovery or high-water-mark observations.
    ///
    /// May allocate when an episode closes (`error.OutOfMemory`; the state
    /// is then unchanged).
    pub fn update(self: *Self, drawdown: f64) Allocator.Error!void {
        const idx = self.count;

        if (drawdown < 0.0) {
            self.count += 1;
            // We are underwater.
            if (self.current_from == null) {
                // Start a new drawdown episode.
                self.current_from = idx;
                self.current_trough = idx;
                self.current_depth = drawdown;
            } else if (drawdown < self.current_depth) {
                // New trough within the current episode.
                self.current_trough = idx;
                self.current_depth = drawdown;
            }
            return;
        }

        // We are at or above the high-water mark.
        if (self.current_from) |from| {
            // Close the current drawdown episode.
            try self.closed.append(self.allocator, .{
                .depth = self.current_depth,
                .from_idx = from,
                .trough_idx = self.current_trough,
                .to_idx = idx,
                .recovered = true,
            });
            self.sum_depth.update(self.current_depth);
            self.sum_depth_squared.update(self.current_depth * self.current_depth);
            self.sum_length += idx - from + 1;
            self.sum_peak_to_trough += self.current_trough - from + 1;
            self.sum_recovery += idx - self.current_trough + 1;
            self.current_from = null;
            self.current_trough = 0;
            self.current_depth = 0.0;
        }
        self.count += 1;
    }

    /// Rebuild all episodes from drawdown observations history.
    pub fn recalculate(self: *Self, drawdowns: []const f64) Allocator.Error!void {
        self.reset();
        for (drawdowns) |d| try self.update(d);
    }

    fn openEpisode(self: *const Self) ?DrawdownEpisode {
        const from = self.current_from orelse return null;
        return .{
            .depth = self.current_depth,
            .from_idx = from,
            .trough_idx = self.current_trough,
            .to_idx = self.count - 1,
            .recovered = false,
        };
    }

    /// Number of episodes, including an open one.
    pub fn episodeCount(self: *const Self) usize {
        return self.closed.items.len + @intFromBool(self.current_from != null);
    }

    /// Drawdown episodes currently known to the tracker.
    ///
    /// If the latest drawdown episode is still open, it is included using
    /// the last processed observation as `to_idx` and with
    /// `recovered = false`.
    ///
    /// Returns a new slice owned by the caller (free with `allocator`).
    pub fn episodes(self: *const Self, allocator: Allocator) Allocator.Error![]DrawdownEpisode {
        const out = try allocator.alloc(DrawdownEpisode, self.episodeCount());
        @memcpy(out[0..self.closed.items.len], self.closed.items);
        if (self.openEpisode()) |e| out[out.len - 1] = e;
        return out;
    }

    /// Drawdown episode depths currently known to the tracker, including
    /// the depth so far of an open episode.
    ///
    /// Returns a new slice owned by the caller (free with `allocator`).
    pub fn depths(self: *const Self, allocator: Allocator) Allocator.Error![]f64 {
        const out = try allocator.alloc(f64, self.episodeCount());
        for (self.closed.items, 0..) |e, i| out[i] = e.depth;
        if (self.current_from != null) out[out.len - 1] = self.current_depth;
        return out;
    }

    /// The mean magnitude of the observed discrete episode drawdowns
    /// (0.0 when there are none).
    pub fn averageEpisodeDrawdown(self: *const Self) f64 {
        var sum_depth = self.sum_depth.value();
        var count = self.closed.items.len;
        if (self.current_from != null) {
            sum_depth += self.current_depth;
            count += 1;
        }
        return if (count > 0) -sum_depth / @as(f64, @floatFromInt(count)) else 0.0;
    }

    /// The sum of squared episode depths divided by the number of
    /// observations (not the number of episodes), as in
    /// PerformanceAnalytics ``DrawdownDeviation`` (0.0 when empty).
    pub fn averageEpisodeDrawdownSquared(self: *const Self) f64 {
        var sum_depth_squared = self.sum_depth_squared.value();
        const count = self.count;
        if (count == 0) return 0.0;
        if (self.current_from != null) sum_depth_squared += self.current_depth * self.current_depth;
        return sum_depth_squared / @as(f64, @floatFromInt(count));
    }

    /// The mean length of the observed discrete drawdown episodes
    /// (0.0 when there are none).
    pub fn averageEpisodeLength(self: *const Self) f64 {
        var sum_length = self.sum_length;
        var count = self.closed.items.len;
        if (self.current_from) |from| {
            sum_length += self.count - from;
            count += 1;
        }
        return if (count > 0) @as(f64, @floatFromInt(sum_length)) / @as(f64, @floatFromInt(count)) else 0.0;
    }

    /// The mean peak-to-trough length of the observed discrete drawdown
    /// episodes (0.0 when there are none).
    pub fn averageEpisodePeakToTrough(self: *const Self) f64 {
        var sum_peak_to_trough = self.sum_peak_to_trough;
        var count = self.closed.items.len;
        if (self.current_from) |from| {
            sum_peak_to_trough += self.current_trough - from + 1;
            count += 1;
        }
        return if (count > 0) @as(f64, @floatFromInt(sum_peak_to_trough)) / @as(f64, @floatFromInt(count)) else 0.0;
    }

    /// The mean recovery length of the observed discrete drawdown episodes
    /// (0.0 when there are none).
    pub fn averageEpisodeRecovery(self: *const Self) f64 {
        var sum_recovery = self.sum_recovery;
        var count = self.closed.items.len;
        if (self.current_from != null) {
            sum_recovery += self.count - self.current_trough;
            count += 1;
        }
        return if (count > 0) @as(f64, @floatFromInt(sum_recovery)) / @as(f64, @floatFromInt(count)) else 0.0;
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

// Drawdown observations (decimals) with two recovered episodes and an
// open one at the end:
//
//   idx:   0     1     2     3     4     5     6     7     8
//   dd:    0  -0.10 -0.20 -0.05    0  -0.03    0  -0.04 -0.06
//              |--- episode 1 ---|     |-- 2 --|     |-- 3 (open)
const test_drawdowns = [_]f64{ 0.0, -0.10, -0.20, -0.05, 0.0, -0.03, 0.0, -0.04, -0.06 };

const test_episodes = [_]DrawdownEpisode{
    .{ .depth = -0.20, .from_idx = 1, .trough_idx = 2, .to_idx = 4, .recovered = true },
    .{ .depth = -0.03, .from_idx = 5, .trough_idx = 5, .to_idx = 6, .recovered = true },
    .{ .depth = -0.06, .from_idx = 7, .trough_idx = 8, .to_idx = 8, .recovered = false },
};

fn ep(depth: f64, from: usize, trough: usize, to: usize, recovered: bool) DrawdownEpisode {
    return .{ .depth = depth, .from_idx = from, .trough_idx = trough, .to_idx = to, .recovered = recovered };
}

fn feed(t: *DrawdownEpisodes, drawdowns: []const f64) !void {
    for (drawdowns) |d| try t.update(d);
}

fn expectEpisodes(t: *const DrawdownEpisodes, expected: []const DrawdownEpisode) !void {
    const actual = try t.episodes(testing.allocator);
    defer testing.allocator.free(actual);
    try testing.expectEqualSlices(DrawdownEpisode, expected, actual);
}

fn expectDepths(t: *const DrawdownEpisodes, expected: []const f64) !void {
    const actual = try t.depths(testing.allocator);
    defer testing.allocator.free(actual);
    try testing.expectEqualSlices(f64, expected, actual);
}

fn assertEmpty(t: *const DrawdownEpisodes) !void {
    try expectEpisodes(t, &.{});
    try expectDepths(t, &.{});
    try testing.expectEqual(@as(f64, 0.0), t.averageEpisodeDrawdown());
    try testing.expectEqual(@as(f64, 0.0), t.averageEpisodeDrawdownSquared());
    try testing.expectEqual(@as(f64, 0.0), t.averageEpisodeLength());
    try testing.expectEqual(@as(f64, 0.0), t.averageEpisodePeakToTrough());
    try testing.expectEqual(@as(f64, 0.0), t.averageEpisodeRecovery());
}

test "empty" {
    var t = DrawdownEpisodes.init(testing.allocator);
    defer t.deinit();
    try assertEmpty(&t);
}

test "no drawdowns" {
    var t = DrawdownEpisodes.init(testing.allocator);
    defer t.deinit();
    try feed(&t, &.{ 0.0, 0.0, 0.0 });
    try expectEpisodes(&t, &.{});
    try testing.expectEqual(@as(f64, 0.0), t.averageEpisodeDrawdown());
    try testing.expectEqual(@as(f64, 0.0), t.averageEpisodeDrawdownSquared());
}

test "episodes" {
    var t = DrawdownEpisodes.init(testing.allocator);
    defer t.deinit();
    try feed(&t, &test_drawdowns);
    try expectEpisodes(&t, &test_episodes);
    try expectDepths(&t, &.{ -0.20, -0.03, -0.06 });
}

test "averages" {
    var t = DrawdownEpisodes.init(testing.allocator);
    defer t.deinit();
    try feed(&t, &test_drawdowns);
    try expectAlmostEqual((0.20 + 0.03 + 0.06) / 3.0, t.averageEpisodeDrawdown(), 16);
    // Divided by the number of observations, not episodes.
    try expectAlmostEqual((0.04 + 0.0009 + 0.0036) / @as(f64, test_drawdowns.len), t.averageEpisodeDrawdownSquared(), 16);
    // Lengths 4, 2, 2 (the open episode has no recovery observation).
    try expectAlmostEqual(8.0 / 3.0, t.averageEpisodeLength(), 15);
    // Peak to trough 2, 1, 2.
    try expectAlmostEqual(5.0 / 3.0, t.averageEpisodePeakToTrough(), 15);
    // Recovery 3, 2, 1.
    try expectAlmostEqual(2.0, t.averageEpisodeRecovery(), 15);
}

test "open episode progress" {
    var t = DrawdownEpisodes.init(testing.allocator);
    defer t.deinit();
    try t.update(-0.05);
    try expectEpisodes(&t, &.{ep(-0.05, 0, 0, 0, false)});
    try t.update(-0.02); // shallower, the trough doesn't move
    try expectEpisodes(&t, &.{ep(-0.05, 0, 0, 1, false)});
    try t.update(-0.08); // new trough
    try expectEpisodes(&t, &.{ep(-0.08, 0, 2, 2, false)});
    try t.update(0.0); // recovery closes the episode
    try expectEpisodes(&t, &.{ep(-0.08, 0, 2, 3, true)});
}

test "recalculate matches incremental" {
    var incremental = DrawdownEpisodes.init(testing.allocator);
    defer incremental.deinit();
    try feed(&incremental, &test_drawdowns);
    var rebuilt = DrawdownEpisodes.init(testing.allocator);
    defer rebuilt.deinit();
    try feed(&rebuilt, &.{ -0.5, -0.7, 0.0 });
    try rebuilt.recalculate(&test_drawdowns);
    const inc_eps = try incremental.episodes(testing.allocator);
    defer testing.allocator.free(inc_eps);
    try expectEpisodes(&rebuilt, inc_eps);
    try testing.expectEqual(incremental.averageEpisodeDrawdown(), rebuilt.averageEpisodeDrawdown());
    try testing.expectEqual(incremental.averageEpisodeDrawdownSquared(), rebuilt.averageEpisodeDrawdownSquared());
    try testing.expectEqual(incremental.averageEpisodeLength(), rebuilt.averageEpisodeLength());
    try testing.expectEqual(incremental.averageEpisodeRecovery(), rebuilt.averageEpisodeRecovery());
}

test "reset" {
    var t = DrawdownEpisodes.init(testing.allocator);
    defer t.deinit();
    try feed(&t, &test_drawdowns);
    t.reset();
    try assertEmpty(&t);
    try t.update(-0.01);
    try expectEpisodes(&t, &.{ep(-0.01, 0, 0, 0, false)});
}
