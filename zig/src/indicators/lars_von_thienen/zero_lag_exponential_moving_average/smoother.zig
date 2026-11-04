const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
// Native Go numerical port. Source conventions: description.md.
pub const ZLEMA = struct {
    alpha: f64 = 0,
    first: f64 = 0,
    value: f64 = 0,
    lag: isize = 0,
    history: []f64 = &.{},
    pub fn update(self: *ZLEMA, allocator: std.mem.Allocator, price: f64) anyerror!f64 {
        if (std.math.isNan(self.first)) {
            self.first = price;
            self.value = price;
            self.history = (try harmoniccycle.push(allocator, self.history, price, (self.lag + 1)));
            return price;
        }
        const history = (try memory.append(allocator, (try memory.concat(allocator, (try memory.dupe(allocator, f64, &[_]f64{})), self.history)), price));
        var lagged = self.first;
        if ((@as(isize, @intCast(history.len)) > self.lag)) {
            lagged = history[@intCast(((@as(isize, @intCast(history.len)) - 1) - self.lag))];
        }
        const adjusted = (try harmoniccycle.finite(allocator, ((@as(f64, 2) * price) - lagged)));
        const value = (try harmoniccycle.finite(allocator, ((self.alpha * adjusted) + (((@as(f64, 1) - self.alpha)) * self.value))));
        self.history = (try harmoniccycle.push(allocator, self.history, price, (self.lag + 1)));
        self.value = value;
        return value;
    }
};
pub fn newZLEMA(allocator: std.mem.Allocator, length: isize) anyerror!?*ZLEMA {
    try harmoniccycle.integer(allocator, length, "length", 1);
    return (try memory.create(allocator, ZLEMA{ .alpha = (@as(f64, 2) / @as(f64, @floatFromInt((length + 1)))), .lag = @divTrunc(((length - 1)), 2), .first = std.math.nan(f64), .value = std.math.nan(f64), .history = (try memory.make(allocator, f64, @intCast(0))) }));
}
