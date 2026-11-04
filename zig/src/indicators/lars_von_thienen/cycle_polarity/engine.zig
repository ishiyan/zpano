const Params = @import("cycle_polarity.zig").CyclePolarityParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const hodrickprescottsmoothinganddetrending = @import("../hodrick_prescott_smoothing_and_detrending/solver.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    polarity: f64 = 0,
    short_value: f64 = 0,
    long_value: f64 = 0,
    short_slope: f64 = 0,
    long_slope: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .polarity = nan, .short_value = nan, .long_value = nan, .short_slope = nan, .long_slope = nan };
}
pub const Engine = struct {
    p: Params = .{},
    prices: []f64 = &.{},
    pub fn reset(self: *Engine, _: std.mem.Allocator) anyerror!void {
        self.prices = &.{};
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
            const short = (try harmoniccycle.goertzel(allocator, values, self.p.short_period));
            const long = (try harmoniccycle.goertzel(allocator, values, self.p.long_period));
            const shortSlope = (try harmoniccycle.cycleSlope(allocator, short, @as(f64, @floatFromInt((self.p.window - 1)))));
            const longSlope = (try harmoniccycle.cycleSlope(allocator, long, @as(f64, @floatFromInt((self.p.window - 1)))));
            output = Snapshot{ .polarity = (try harmoniccycle.cyclePolarity(allocator, shortSlope, longSlope, self.p.tolerance)), .short_value = (try harmoniccycle.cycleValue(allocator, short, @as(f64, @floatFromInt((self.p.window - 1))))), .long_value = (try harmoniccycle.cycleValue(allocator, long, @as(f64, @floatFromInt((self.p.window - 1))))), .short_slope = shortSlope, .long_slope = longSlope };
        }
        self.prices = prices;
        return output;
    }
};
