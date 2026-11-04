const Params = @import("goertzel_cycle_spectrum.zig").GoertzelCycleSpectrumParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const hodrickprescottsmoothinganddetrending = @import("../hodrick_prescott_smoothing_and_detrending/solver.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    amplitude: []f64 = &.{},
    phases: []f64 = &.{},
    trough_offsets: []f64 = &.{},
    dominant_period: f64 = 0,
    dominant_amplitude: f64 = 0,
    dominant_phase: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .amplitude = &.{}, .phases = &.{}, .trough_offsets = &.{}, .dominant_period = nan, .dominant_amplitude = nan, .dominant_phase = nan };
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
            const count = (@divTrunc(((self.p.max_period - self.p.min_period)), self.p.step) + 1);
            output.amplitude = (try memory.make(allocator, f64, @intCast(count)));
            output.phases = (try memory.make(allocator, f64, @intCast(count)));
            output.trough_offsets = (try memory.make(allocator, f64, @intCast(count)));
            var best: isize = 0;
            {
                var i: isize = 0;
                while (i < count) : (i += 1) {
                    const model = (try harmoniccycle.goertzel(allocator, values, @as(f64, @floatFromInt((self.p.min_period + (i * self.p.step))))));
                    output.amplitude[@intCast(i)] = model.amplitude;
                    output.phases[@intCast(i)] = model.phase;
                    output.trough_offsets[@intCast(i)] = model.trough_offset;
                    if ((model.amplitude > output.amplitude[@intCast(best)])) {
                        best = i;
                    }
                }
            }
            output.dominant_amplitude = output.amplitude[@intCast(best)];
            output.dominant_phase = output.phases[@intCast(best)];
            if ((output.dominant_amplitude > @as(f64, 0))) {
                output.dominant_period = @as(f64, @floatFromInt((self.p.min_period + (best * self.p.step))));
            }
        }
        self.prices = prices;
        return output;
    }
};
