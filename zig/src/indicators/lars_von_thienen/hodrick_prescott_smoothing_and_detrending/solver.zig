const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
// Native Go numerical port. Source conventions: description.md.
pub fn hpTrend(allocator: std.mem.Allocator, values: []f64, smoothing: f64) anyerror![]f64 {
    const n = @as(isize, @intCast(values.len));
    if ((n < 6)) {
        return error.InvalidInput;
    }
    try harmoniccycle.real(allocator, smoothing, "smoothing", 0);
    var a = (try memory.make(allocator, f64, @intCast(n)));
    var b = (try memory.make(allocator, f64, @intCast(n)));
    var c = (try memory.make(allocator, f64, @intCast(n)));
    for (values, 0..) |value, ordinal| {
        const i: isize = @intCast(ordinal);
        (try harmoniccycle.sample(allocator, value));
        if (std.math.isNan(value)) {
            return error.InvalidInput;
        }
        a[@intCast(i)] = ((@as(f64, 6) * smoothing) + @as(f64, 1));
        b[@intCast(i)] = (@as(f64, @floatFromInt(-4)) * smoothing);
        c[@intCast(i)] = smoothing;
    }
    a[@intCast(0)] = (@as(f64, 1) + smoothing);
    a[@intCast((n - 1))] = (@as(f64, 1) + smoothing);
    a[@intCast(1)] = (@as(f64, 1) + (@as(f64, 5) * smoothing));
    a[@intCast((n - 2))] = (@as(f64, 1) + (@as(f64, 5) * smoothing));
    b[@intCast(0)] = (@as(f64, @floatFromInt(-2)) * smoothing);
    b[@intCast((n - 2))] = (@as(f64, @floatFromInt(-2)) * smoothing);
    b[@intCast((n - 1))] = @as(f64, 0);
    c[@intCast((n - 1))] = @as(f64, 0);
    c[@intCast((n - 2))] = @as(f64, 0);
    var h1: f64 = 0;
    var h2: f64 = 0;
    var h3: f64 = 0;
    var h4: f64 = 0;
    var h5: f64 = 0;
    var hh2: f64 = 0;
    var hh3: f64 = 0;
    var hh5: f64 = 0;

    {
        var i: isize = 0;
        while (i < n) : (i += 1) {
            const z = (try harmoniccycle.finite(allocator, ((a[@intCast(i)] - (h4 * h1)) - (hh5 * hh2))));
            if ((z <= @as(f64, 0))) {
                return error.InvalidInput;
            }
            const hb = b[@intCast(i)];
            const hh1 = h1;
            h1 = (((hb - (h4 * h2))) / z);
            b[@intCast(i)] = h1;
            const hc = c[@intCast(i)];
            hh2 = h2;
            h2 = (hc / z);
            c[@intCast(i)] = h2;
            a[@intCast(i)] = (try harmoniccycle.finite(allocator, ((((values[@intCast(i)] - (hh3 * hh5)) - (h3 * h4))) / z)));
            hh3 = h3;
            h3 = a[@intCast(i)];
            h4 = (hb - (h5 * hh1));
            hh5 = h5;
            h5 = hc;
        }
    }
    var output = (try memory.make(allocator, f64, @intCast(n)));
    h2 = @as(f64, 0);
    h1 = a[@intCast((n - 1))];
    output[@intCast((n - 1))] = h1;
    {
        var i = (n - 1);
        while ((i > 0)) : (i -= 1) {
            output[@intCast((i - 1))] = (try harmoniccycle.finite(allocator, ((a[@intCast((i - 1))] - (b[@intCast((i - 1))] * h1)) - (c[@intCast((i - 1))] * h2))));
            const temporary1_0 = h1;
            const temporary1_1 = output[@intCast((i - 1))];
            h2 = temporary1_0;
            h1 = temporary1_1;
        }
    }
    return output;
}
pub fn cycleData(allocator: std.mem.Allocator, values: []f64, smoothing: f64) anyerror![]f64 {
    if ((smoothing == @as(f64, 0))) {
        return (try memory.concat(allocator, (try memory.dupe(allocator, f64, &[_]f64{})), values));
    }
    var trend = (try hpTrend(allocator, values, smoothing));
    {
        var i: isize = 0;
        while (i < @as(isize, @intCast(trend.len))) : (i += 1) {
            trend[@intCast(i)] = (try harmoniccycle.finite(allocator, (values[@intCast(i)] - trend[@intCast(i)])));
        }
    }
    return trend;
}
