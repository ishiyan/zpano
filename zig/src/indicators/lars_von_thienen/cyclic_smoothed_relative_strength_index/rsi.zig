const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
// Native Go numerical port. Source conventions: description.md.
pub const WilderRSI = struct {
    length: isize = 0,
    count: isize = 0,
    previous: f64 = 0,
    up: f64 = 0,
    down: f64 = 0,
    pub fn update(self: *WilderRSI, allocator: std.mem.Allocator, price: f64) anyerror!f64 {
        if (std.math.isNan(self.previous)) {
            self.previous = price;
            return std.math.nan(f64);
        }
        const change = (try harmoniccycle.finite(allocator, (price - self.previous)));
        const temporary1_0 = @max(change, @as(f64, 0));
        const temporary1_1 = @max(-change, @as(f64, 0));
        const gain = temporary1_0;
        const loss = temporary1_1;
        if ((self.count < self.length)) {
            self.up = (try harmoniccycle.finite(allocator, (self.up + gain)));
            self.down = (try harmoniccycle.finite(allocator, (self.down + loss)));
            self.count += 1;
            self.previous = price;
            if ((self.count < self.length)) {
                return std.math.nan(f64);
            }
            self.up /= @as(f64, @floatFromInt(self.length));
            self.down /= @as(f64, @floatFromInt(self.length));
        } else {
            self.up = (try harmoniccycle.finite(allocator, ((((self.up * @as(f64, @floatFromInt((self.length - 1)))) + gain)) / @as(f64, @floatFromInt(self.length)))));
            self.down = (try harmoniccycle.finite(allocator, ((((self.down * @as(f64, @floatFromInt((self.length - 1)))) + loss)) / @as(f64, @floatFromInt(self.length)))));
            self.previous = price;
        }
        if ((self.down == @as(f64, 0))) {
            return 100;
        }
        if ((self.up == @as(f64, 0))) {
            return 0;
        }
        return (@as(f64, 100) - (@as(f64, 100) / ((@as(f64, 1) + (self.up / self.down)))));
    }
};
pub fn newWilderRSI(allocator: std.mem.Allocator, length: isize) anyerror!?*WilderRSI {
    try harmoniccycle.integer(allocator, length, "length", 1);
    return (try memory.create(allocator, WilderRSI{ .length = length, .previous = std.math.nan(f64) }));
}
