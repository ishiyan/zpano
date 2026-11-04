const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
// Native Go numerical port. Source conventions: description.md.
pub fn gridBands(_: std.mem.Allocator, values: []f64, leveling: f64) anyerror!struct { f64, f64 } {
    const temporary1_0 = values[@intCast(0)];
    const temporary1_1 = values[@intCast(0)];
    var low = temporary1_0;
    var high = temporary1_1;
    for (values) |value| {
        low = @min(low, value);
        high = @max(high, value);
    }
    const step = (((high - low)) / @as(f64, 100));
    const fraction = (leveling / @as(f64, 100));
    const temporary2_0 = low;
    const temporary2_1 = high;
    var lower = temporary2_0;
    var upper = temporary2_1;
    {
        var k: isize = 0;
        while (k < 101) : (k += 1) {
            const testValue = (low + (step * @as(f64, @floatFromInt(k))));
            var below: isize = 0;
            for (values) |value| {
                if ((value < testValue)) {
                    below += 1;
                }
            }
            if (((@as(f64, @floatFromInt(below)) / @as(f64, @floatFromInt(@as(isize, @intCast(values.len))))) >= fraction)) {
                lower = testValue;
                break;
            }
        }
    }
    {
        var k: isize = 0;
        while (k < 101) : (k += 1) {
            const testValue = (high - (step * @as(f64, @floatFromInt(k))));
            var above: isize = 0;
            for (values) |value| {
                if ((value >= testValue)) {
                    above += 1;
                }
            }
            if (((@as(f64, @floatFromInt(above)) / @as(f64, @floatFromInt(@as(isize, @intCast(values.len))))) >= fraction)) {
                upper = testValue;
                break;
            }
        }
    }
    return .{ lower, upper };
}
