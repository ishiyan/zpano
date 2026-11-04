const Params = @import("hodrick_prescott_smoothing_and_detrending.zig").HodrickPrescottSmoothingAndDetrendingParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const hpTrend = @import("solver.zig").hpTrend;
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    trend: f64 = 0,
    cycle: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .trend = nan, .cycle = nan };
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
            const trend = (try hpTrend(allocator, prices, self.p.smoothing));
            output = Snapshot{ .trend = trend[@intCast((@as(isize, @intCast(trend.len)) - 1))], .cycle = (try harmoniccycle.finite(allocator, (price - trend[@intCast((@as(isize, @intCast(trend.len)) - 1))]))) };
        }
        self.prices = prices;
        return output;
    }
};
