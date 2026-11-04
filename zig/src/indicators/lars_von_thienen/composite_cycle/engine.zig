const Params = @import("composite_cycle.zig").CompositeCycleParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const hodrickprescottsmoothinganddetrending = @import("../hodrick_prescott_smoothing_and_detrending/solver.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    value: f64 = 0,
    slope: f64 = 0,
    next_value: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .value = nan, .slope = nan, .next_value = nan };
}
pub const Engine = struct {
    p: Params = .{},
    prices: []f64 = &.{},
    models: []harmoniccycle.CycleModel = &.{},
    pub fn reset(self: *Engine, _: std.mem.Allocator) anyerror!void {
        self.prices = &.{};
        self.models = &.{};
    }
    pub fn primed(self: *Engine, _: std.mem.Allocator) anyerror!bool {
        return (@as(isize, @intCast(self.prices.len)) == self.p.window);
    }
    pub fn warmup(self: *Engine, _: std.mem.Allocator) anyerror!isize {
        return (self.p.window - 1);
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        const prices = (try harmoniccycle.push(allocator, (try memory.concat(allocator, (try memory.dupe(allocator, f64, &[_]f64{})), self.prices)), price, self.p.window));
        var output = (try emptySnapshot(allocator));
        if ((@as(isize, @intCast(prices.len)) == self.p.window)) {
            const values = (try hodrickprescottsmoothinganddetrending.cycleData(allocator, prices, self.p.smoothing));
            var models = (try memory.make(allocator, harmoniccycle.CycleModel, @intCast(@as(isize, @intCast(self.p.periods.len)))));
            var sum: f64 = 0;

            for (self.p.periods, 0..) |period, ordinal| {
                const i: isize = @intCast(ordinal);
                models[@intCast(i)] = (try harmoniccycle.goertzel(allocator, values, period));
                sum += (try harmoniccycle.cycleSlope(allocator, models[@intCast(i)], @as(f64, @floatFromInt((self.p.window - 1)))));
            }
            output = Snapshot{ .value = (try harmoniccycle.compositeValue(allocator, models, @as(f64, @floatFromInt((self.p.window - 1))))), .slope = (try harmoniccycle.finite(allocator, sum)), .next_value = (try harmoniccycle.compositeValue(allocator, models, @as(f64, @floatFromInt(self.p.window)))) };
            self.models = models;
        }
        self.prices = prices;
        return output;
    }
};
