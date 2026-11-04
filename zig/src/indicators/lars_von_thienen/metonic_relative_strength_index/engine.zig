const Params = @import("metonic_relative_strength_index.zig").MetonicRelativeStrengthIndexParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const cyclicsmoothedrelativestrengthindex = @import("../cyclic_smoothed_relative_strength_index/rsi.zig");
const metonic = @import("../metonic/calendar.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    value: f64 = 0,
    linear_rsi: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .value = nan, .linear_rsi = nan };
}
pub const Engine = struct {
    p: Params = .{},
    calendar: metonic.Calendar = .{},
    delay: ?*metonic.DelayMean = null,
    last_value: f64 = 0,
    samples_seen: isize = 0,
    rsi: ?*cyclicsmoothedrelativestrengthindex.WilderRSI = null,
    pub fn reset(self: *Engine, allocator: std.mem.Allocator) anyerror!void {
        self.calendar = metonic.Calendar{};
        self.delay = (try metonic.newDelayMean(allocator, (try memory.dupe(allocator, isize, &[_]isize{ self.p.delay1, self.p.delay2, self.p.delay3 }))));
        self.last_value = std.math.nan(f64);
        self.samples_seen = 0;
        self.rsi = (try cyclicsmoothedrelativestrengthindex.newWilderRSI(allocator, self.p.length));
    }
    pub fn primed(self: *Engine, _: std.mem.Allocator) anyerror!bool {
        return !std.math.isNan(self.last_value);
    }
    pub fn warmup(self: *Engine, _: std.mem.Allocator) anyerror!isize {
        return (@max(@max(self.p.delay1, self.p.delay2), self.p.delay3) + self.p.length);
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64, day: ?i64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        try self.calendar.validate(allocator, day);
        const linear = (try (self.rsi orelse unreachable).update(allocator, price));
        self.last_value = (try (self.delay orelse unreachable).update(allocator, linear));
        self.samples_seen += 1;
        (try self.calendar.commit(allocator, day));
        return Snapshot{ .value = self.last_value, .linear_rsi = linear };
    }
};
