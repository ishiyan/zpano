const Params = @import("metonic_sentiment_index.zig").MetonicSentimentIndexParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const zerolagexponentialmovingaverage = @import("../zero_lag_exponential_moving_average/smoother.zig");
const metonic = @import("../metonic/calendar.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    value: f64 = 0,
    daily_sentiment: f64 = 0,
    smoothed_close: f64 = 0,
    smoothed_open: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .value = nan, .daily_sentiment = nan, .smoothed_close = nan, .smoothed_open = nan };
}
pub const Engine = struct {
    p: Params = .{},
    calendar: metonic.Calendar = .{},
    delay: ?*metonic.DelayMean = null,
    last_value: f64 = 0,
    samples_seen: isize = 0,
    close_smoother: ?*zerolagexponentialmovingaverage.ZLEMA = null,
    open_smoother: ?*zerolagexponentialmovingaverage.ZLEMA = null,
    previous_close: f64 = 0,
    pub fn reset(self: *Engine, allocator: std.mem.Allocator) anyerror!void {
        self.calendar = metonic.Calendar{};
        self.delay = (try metonic.newDelayMean(allocator, (try memory.dupe(allocator, isize, &[_]isize{ self.p.delay1, self.p.delay2, self.p.delay3 }))));
        self.last_value = std.math.nan(f64);
        self.samples_seen = 0;
        self.close_smoother = (try zerolagexponentialmovingaverage.newZLEMA(allocator, self.p.length));
        self.open_smoother = (try zerolagexponentialmovingaverage.newZLEMA(allocator, self.p.length));
        self.previous_close = std.math.nan(f64);
    }
    pub fn primed(self: *Engine, _: std.mem.Allocator) anyerror!bool {
        return !std.math.isNan(self.last_value);
    }
    pub fn warmup(self: *Engine, _: std.mem.Allocator) anyerror!isize {
        return ((@max(@max(self.p.delay1, self.p.delay2), self.p.delay3) + @divTrunc(((self.p.length - 1)), 2)) + self.p.mode);
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64, day: ?i64, openPrice: f64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        if ((self.p.mode == 0)) {
            (try harmoniccycle.sample(allocator, openPrice));
            if (std.math.isNan(openPrice)) {
                return (try emptySnapshot(allocator));
            }
        }
        try self.calendar.validate(allocator, day);
        const smoothedClose = (try (self.close_smoother orelse unreachable).update(allocator, price));
        var smoothedOpen = std.math.nan(f64);
        var comparison = self.previous_close;
        if ((self.p.mode == 0)) {
            smoothedOpen = (try (self.open_smoother orelse unreachable).update(allocator, openPrice));
            comparison = smoothedOpen;
        }
        var sentiment = std.math.nan(f64);
        if (((self.samples_seen >= (@divTrunc(((self.p.length - 1)), 2) + self.p.mode)) and !std.math.isNan(comparison))) {
            sentiment = @as(f64, @floatFromInt(-1));
            if ((smoothedClose > comparison)) {
                sentiment = @as(f64, 1);
            }
        }
        self.previous_close = smoothedClose;
        const mean = (try (self.delay orelse unreachable).update(allocator, sentiment));
        self.last_value = (@as(f64, 3) * mean);
        self.samples_seen += 1;
        (try self.calendar.commit(allocator, day));
        return Snapshot{ .value = self.last_value, .daily_sentiment = sentiment, .smoothed_close = smoothedClose, .smoothed_open = smoothedOpen };
    }
};
