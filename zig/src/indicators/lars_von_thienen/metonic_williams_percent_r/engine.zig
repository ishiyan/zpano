const Params = @import("metonic_williams_percent_r.zig").MetonicWilliamsPercentRParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const metonic = @import("../metonic/calendar.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    value: f64 = 0,
    linear_percent_r: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .value = nan, .linear_percent_r = nan };
}
pub const Engine = struct {
    p: Params = .{},
    calendar: metonic.Calendar = .{},
    delay: ?*metonic.DelayMean = null,
    last_value: f64 = 0,
    samples_seen: isize = 0,
    highs: []f64 = &.{},
    lows: []f64 = &.{},
    pub fn reset(self: *Engine, allocator: std.mem.Allocator) anyerror!void {
        self.calendar = metonic.Calendar{};
        self.delay = (try metonic.newDelayMean(allocator, (try memory.dupe(allocator, isize, &[_]isize{ self.p.delay1, self.p.delay2, self.p.delay3 }))));
        self.last_value = std.math.nan(f64);
        self.samples_seen = 0;
        self.highs = &.{};
        self.lows = &.{};
    }
    pub fn primed(self: *Engine, _: std.mem.Allocator) anyerror!bool {
        return !std.math.isNan(self.last_value);
    }
    pub fn warmup(self: *Engine, _: std.mem.Allocator) anyerror!isize {
        return ((@max(@max(self.p.delay1, self.p.delay2), self.p.delay3) + self.p.length) - 1);
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64, day: ?i64, high: f64, low: f64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        (try harmoniccycle.sample(allocator, high));
        (try harmoniccycle.sample(allocator, low));
        if ((std.math.isNan(high) or std.math.isNan(low))) {
            return (try emptySnapshot(allocator));
        }
        if (((low > price) or (price > high))) {
            return error.InvalidInput;
        }
        try self.calendar.validate(allocator, day);
        self.highs = (try harmoniccycle.push(allocator, self.highs, high, self.p.length));
        self.lows = (try harmoniccycle.push(allocator, self.lows, low, self.p.length));
        var linear = std.math.nan(f64);
        if ((@as(isize, @intCast(self.highs.len)) == self.p.length)) {
            const temporary1_0 = self.highs[@intCast(0)];
            const temporary1_1 = self.lows[@intCast(0)];
            var maximum = temporary1_0;
            var minimum = temporary1_1;
            for (self.highs) |value| {
                maximum = @max(maximum, value);
            }
            for (self.lows) |value| {
                minimum = @min(minimum, value);
            }
            if ((maximum > minimum)) {
                linear = ((@as(f64, @floatFromInt(-100)) * ((maximum - price))) / ((maximum - minimum)));
            }
        }
        self.last_value = (try (self.delay orelse unreachable).update(allocator, linear));
        self.samples_seen += 1;
        (try self.calendar.commit(allocator, day));
        return Snapshot{ .value = self.last_value, .linear_percent_r = linear };
    }
};
