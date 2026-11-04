const Params = @import("zero_lag_exponential_moving_average.zig").LarsVonThienenZeroLagExponentialMovingAverageParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const ZLEMA = @import("smoother.zig").ZLEMA;
const newZLEMA = @import("smoother.zig").newZLEMA;
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    value: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .value = nan };
}
pub const Engine = struct {
    p: Params = .{},
    smoother: ?*ZLEMA = null,
    samples_seen: isize = 0,
    pub fn reset(self: *Engine, allocator: std.mem.Allocator) anyerror!void {
        self.smoother = (try newZLEMA(allocator, self.p.length));
        self.samples_seen = 0;
    }
    pub fn primed(self: *Engine, _: std.mem.Allocator) anyerror!bool {
        return (self.samples_seen >= (@divTrunc(((self.p.length - 1)), 2) + 1));
    }
    pub fn warmup(self: *Engine, _: std.mem.Allocator) anyerror!isize {
        return @divTrunc(((self.p.length - 1)), 2);
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        self.samples_seen += 1;
        return Snapshot{ .value = (try (self.smoother orelse unreachable).update(allocator, price)) };
    }
};
