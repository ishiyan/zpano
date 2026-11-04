const Params = @import("bartels_periodicity_score.zig").BartelsPeriodicityScoreParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const hodrickprescottsmoothinganddetrending = @import("../hodrick_prescott_smoothing_and_detrending/solver.zig");
const bartelsScore = @import("scorer.zig").bartelsScore;
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    score: f64 = 0,
    genuine_percent: f64 = 0,
    segments: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .score = nan, .genuine_percent = nan, .segments = nan };
}
pub const Engine = struct {
    p: Params = .{},
    prices: []f64 = &.{},
    pub fn reset(self: *Engine, _: std.mem.Allocator) anyerror!void {
        self.prices = &.{};
    }
    pub fn primed(self: *Engine, _: std.mem.Allocator) anyerror!bool {
        return (@as(isize, @intCast(self.prices.len)) == (self.p.period * self.p.segments));
    }
    pub fn warmup(self: *Engine, _: std.mem.Allocator) anyerror!isize {
        return ((self.p.period * self.p.segments) - 1);
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        const prices = (try harmoniccycle.push(allocator, (try memory.concat(allocator, (try memory.dupe(allocator, f64, &[_]f64{})), self.prices)), price, (self.p.period * self.p.segments)));
        var output = (try emptySnapshot(allocator));
        if ((@as(isize, @intCast(prices.len)) == (self.p.period * self.p.segments))) {
            const score = (try bartelsScore(allocator, (try hodrickprescottsmoothinganddetrending.cycleData(allocator, prices, self.p.smoothing)), self.p.period, self.p.segments, self.p.segments));
            output = Snapshot{ .score = score.score, .genuine_percent = score.genuine_percent, .segments = score.segments };
        }
        self.prices = prices;
        return output;
    }
};
