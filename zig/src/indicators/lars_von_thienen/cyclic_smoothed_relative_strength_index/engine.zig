const Params = @import("cyclic_smoothed_relative_strength_index.zig").CyclicSmoothedRelativeStrengthIndexParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const adaptivebands = @import("../adaptive_bands/bands.zig");
const WilderRSI = @import("rsi.zig").WilderRSI;
const newWilderRSI = @import("rsi.zig").newWilderRSI;
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    value: f64 = 0,
    lower: f64 = 0,
    upper: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .value = nan, .lower = nan, .upper = nan };
}
pub const Engine = struct {
    p: Params = .{},
    rsi: ?*WilderRSI = null,
    rsi_history: []f64 = &.{},
    values: []f64 = &.{},
    previous_value: f64 = 0,
    pub fn reset(self: *Engine, allocator: std.mem.Allocator) anyerror!void {
        self.rsi = (try newWilderRSI(allocator, @divTrunc(self.p.dominant_cycle, 2)));
        self.rsi_history = &.{};
        self.values = &.{};
        self.previous_value = @as(f64, 0);
    }
    pub fn primed(self: *Engine, _: std.mem.Allocator) anyerror!bool {
        return (@as(isize, @intCast(self.values.len)) == (2 * self.p.dominant_cycle));
    }
    pub fn warmup(self: *Engine, _: std.mem.Allocator) anyerror!isize {
        return (((@divTrunc(self.p.dominant_cycle, 2) + @divTrunc(((self.p.vibration - 1)), 2)) + (2 * self.p.dominant_cycle)) - 1);
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        const rsi = (try (self.rsi orelse unreachable).update(allocator, price));
        if (std.math.isNan(rsi)) {
            return (try emptySnapshot(allocator));
        }
        const lag = @divTrunc(((self.p.vibration - 1)), 2);
        self.rsi_history = (try harmoniccycle.push(allocator, self.rsi_history, rsi, (lag + 1)));
        if ((@as(isize, @intCast(self.rsi_history.len)) < (lag + 1))) {
            return (try emptySnapshot(allocator));
        }
        const torque = (@as(f64, 2) / @as(f64, @floatFromInt((self.p.vibration + 1))));
        const value = (try harmoniccycle.finite(allocator, ((torque * (((@as(f64, 2) * rsi) - self.rsi_history[@intCast(0)]))) + (((@as(f64, 1) - torque)) * self.previous_value))));
        self.previous_value = value;
        self.values = (try harmoniccycle.push(allocator, self.values, value, (2 * self.p.dominant_cycle)));
        var output = (try emptySnapshot(allocator));
        output.value = value;
        if ((try self.primed(allocator))) {
            const tuple1 = (try adaptivebands.gridBands(allocator, self.values, self.p.leveling));
            output.lower = tuple1[0];
            output.upper = tuple1[1];
        }
        return output;
    }
};
