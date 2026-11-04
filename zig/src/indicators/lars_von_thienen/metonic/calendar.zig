const std = @import("std");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const memory = @import("../harmonic_cycle/memory.zig");
/// Integer UTC day ordinals. Mixed dated/undated streams and gaps are rejected.
pub const Calendar = struct {
    last: ?i64 = null,
    dated: ?bool = null,
    pub fn validate(self: *Calendar, _: std.mem.Allocator, day: ?i64) !void {
        if (self.dated) |dated| {
            if (dated != (day != null)) return error.MixedCalendarModes;
        }
        if (self.last) |last| {
            if (day == null or day.? != last + 1) return error.CalendarGap;
        }
    }
    pub fn commit(self: *Calendar, _: std.mem.Allocator, day: ?i64) !void {
        self.last = day;
        self.dated = day != null;
    }
};
pub const DelayMean = struct {
    delays: []isize = &.{},
    history: []f64 = &.{},
    capacity: isize = 0,
    pub fn update(self: *DelayMean, a: std.mem.Allocator, value: f64) !f64 {
        self.history = try harmoniccycle.push(a, self.history, value, self.capacity);
        if (self.history.len < @as(usize, @intCast(self.capacity))) return std.math.nan(f64);
        var sum: f64 = 0;
        for (self.delays) |delay| {
            const v = self.history[self.history.len - 1 - @as(usize, @intCast(delay))];
            if (std.math.isNan(v)) return std.math.nan(f64);
            sum += v;
        }
        return sum / @as(f64, @floatFromInt(self.delays.len));
    }
};
pub fn newDelayMean(a: std.mem.Allocator, delays: []isize) !?*DelayMean {
    var capacity: isize = 0;
    for (delays) |delay| {
        if (delay < 1) return error.InvalidDelay;
        capacity = @max(capacity, delay + 1);
    }
    return try memory.create(a, DelayMean{ .delays = try a.dupe(isize, delays), .capacity = capacity });
}
/// Create the shared daily-record type for a complete payload.
pub fn dailyRecord(comptime T: type) type {
    return struct { day: i64, payload: T };
}

/// Fill missing calendar slots with the preceding full payload.
pub fn fillCalendarDays(a: std.mem.Allocator, comptime T: type, records: []const dailyRecord(T)) ![]dailyRecord(T) {
    const Record = dailyRecord(T);
    var output: std.ArrayList(Record) = .empty;
    errdefer output.deinit(a);
    for (records, 0..) |record, i| {
        if (i > 0) {
            const previous = records[i - 1];
            if (record.day <= previous.day) return error.NonIncreasingDates;
            var day = previous.day + 1;
            while (day < record.day) : (day += 1) try output.append(a, .{ .day = day, .payload = previous.payload });
        }
        try output.append(a, .{ .day = record.day, .payload = record.payload });
    }
    return output.toOwnedSlice(a);
}

/// Extract a UTC daily ordinal without silently accepting intraday observations.
pub fn dayFromUnixSeconds(time: i64) !i64 {
    if (@mod(time, 86400) != 0) return error.NonMidnightTimestamp;
    return @divFloor(time, 86400);
}

test "fill calendar gaps with the complete preceding OHLC payload" {
    const data = @import("../metonic_sentiment_index/testdata.zig");
    const open = data.testOpen();
    const high = data.testHigh();
    const low = data.testLow();
    const close = data.testInput();
    const Payload = struct { open: f64, high: f64, low: f64, close: f64 };
    const first: Payload = .{ .open = open[0], .high = high[0], .low = low[0], .close = close[0] };
    const last: Payload = .{ .open = open[1], .high = high[1], .low = low[1], .close = close[1] };
    const filled = try fillCalendarDays(std.testing.allocator, Payload, &.{
        .{ .day = 1, .payload = first },
        .{ .day = 4, .payload = last },
    });
    defer std.testing.allocator.free(filled);
    try std.testing.expectEqual(@as(usize, 4), filled.len);
    for (filled, 0..) |record, i| {
        try std.testing.expectEqual(@as(i64, @intCast(i + 1)), record.day);
        try std.testing.expectEqualDeep(if (i < 3) first else last, record.payload);
    }
}

test "calendar filling rejects duplicate and reversed days" {
    const data = @import("../metonic_sentiment_index/testdata.zig");
    const close = data.testInput();
    try std.testing.expectError(error.NonIncreasingDates, fillCalendarDays(std.testing.allocator, f64, &.{
        .{ .day = 2, .payload = close[0] },
        .{ .day = 2, .payload = close[1] },
    }));
    try std.testing.expectError(error.NonIncreasingDates, fillCalendarDays(std.testing.allocator, f64, &.{
        .{ .day = 2, .payload = close[0] },
        .{ .day = 1, .payload = close[1] },
    }));
}
