const Params = @import("cycle_swing_momentum.zig").CycleSwingMomentumParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const hodrickprescottsmoothinganddetrending = @import("../hodrick_prescott_smoothing_and_detrending/solver.zig");
const adaptivebands = @import("../adaptive_bands/bands.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    value: f64 = 0,
    lower: f64 = 0,
    upper: f64 = 0,
    score: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .value = nan, .lower = nan, .upper = nan, .score = nan };
}
pub const Engine = struct {
    p: Params = .{},
    prices: []f64 = &.{},
    values: []f64 = &.{},
    pub fn reset(self: *Engine, _: std.mem.Allocator) anyerror!void {
        self.prices = &.{};
        self.values = &.{};
    }
    pub fn primed(self: *Engine, _: std.mem.Allocator) anyerror!bool {
        return (@as(isize, @intCast(self.values.len)) == self.p.band_memory);
    }
    pub fn warmup(self: *Engine, _: std.mem.Allocator) anyerror!isize {
        return ((self.p.window + self.p.band_memory) - 2);
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        const prices = (try harmoniccycle.push(allocator, (try memory.concat(allocator, (try memory.dupe(allocator, f64, &[_]f64{})), self.prices)), price, self.p.window));
        var output = (try emptySnapshot(allocator));
        if ((@as(isize, @intCast(prices.len)) == self.p.window)) {
            const fast = (try hodrickprescottsmoothinganddetrending.hpTrend(allocator, prices, self.p.fast_smoothing));
            const slow = (try hodrickprescottsmoothinganddetrending.hpTrend(allocator, prices, self.p.slow_smoothing));
            const value = (try harmoniccycle.finite(allocator, (fast[@intCast((@as(isize, @intCast(fast.len)) - 1))] - slow[@intCast((@as(isize, @intCast(slow.len)) - 1))])));
            self.values = (try harmoniccycle.push(allocator, self.values, value, self.p.band_memory));
            output.value = value;
            if ((try self.primed(allocator))) {
                const tuple1 = (try adaptivebands.gridBands(allocator, self.values, self.p.leveling));
                output.lower = tuple1[0];
                output.upper = tuple1[1];
                output.score = @as(f64, 0);
                if ((value >= output.upper)) {
                    output.score = @as(f64, 1);
                } else if ((value <= output.lower)) {
                    output.score = @as(f64, @floatFromInt(-1));
                }
            }
        }
        self.prices = prices;
        return output;
    }
};
