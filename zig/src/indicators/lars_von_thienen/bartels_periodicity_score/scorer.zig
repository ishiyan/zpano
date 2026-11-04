const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
// Native Go numerical port. Source conventions: description.md.
pub const BartelsOutput = struct {
    score: f64 = 0,
    genuine_percent: f64 = 0,
    segments: f64 = 0,
};
pub fn bartelsScore(allocator: std.mem.Allocator, values: []f64, period: isize, maxSegments: isize, minSegments: isize) anyerror!BartelsOutput {
    try harmoniccycle.integer(allocator, period, "period", 3);
    if ((((maxSegments < 2) or (minSegments < 2)) or (minSegments > maxSegments))) {
        return error.InvalidInput;
    }
    for (values) |value| {
        (try harmoniccycle.sample(allocator, value));
        if (std.math.isNan(value)) {
            return error.InvalidInput;
        }
    }
    const segments = @min(@divTrunc(@as(isize, @intCast(values.len)), period), maxSegments);
    const nan = std.math.nan(f64);
    if ((segments < minSegments)) {
        return BartelsOutput{ .score = nan, .genuine_percent = nan, .segments = @as(f64, @floatFromInt(segments)) };
    }
    var sumA: f64 = 0;
    var sumB: f64 = 0;
    var sumEnergy: f64 = 0;

    {
        var segment: isize = 0;
        while (segment < segments) : (segment += 1) {
            var a: f64 = 0;
            var b: f64 = 0;

            {
                var j: isize = 0;
                while (j < period) : (j += 1) {
                    const x = ((segment * period) + j);
                    const angle = (((((@as(f64, 360) / @as(f64, @floatFromInt(period))) * @as(f64, @floatFromInt((x + 1))))) / @as(f64, 180)) * std.math.pi);
                    a += (@sin(angle) * values[@intCast(x)]);
                    b += (@cos(angle) * values[@intCast(x)]);
                }
            }
            sumA += a;
            sumB += b;
            sumEnergy = (try harmoniccycle.finite(allocator, ((sumEnergy + (a * a)) + (b * b))));
        }
    }
    if ((sumEnergy == @as(f64, 0))) {
        return BartelsOutput{ .score = nan, .genuine_percent = nan, .segments = @as(f64, @floatFromInt(segments)) };
    }
    const temporary1_0 = (sumA / @as(f64, @floatFromInt(segments)));
    const temporary1_1 = (sumB / @as(f64, @floatFromInt(segments)));
    const averageA = temporary1_0;
    const averageB = temporary1_1;
    const amplitude = @sqrt(((averageA * averageA) + (averageB * averageB)));
    const rmsAmplitude = @sqrt((sumEnergy / @as(f64, @floatFromInt(segments))));
    const reference = (rmsAmplitude / @sqrt(@as(f64, @floatFromInt(segments))));
    const ratio = (amplitude / reference);
    const score = @exp((-ratio * ratio));
    return BartelsOutput{ .score = score, .genuine_percent = (@as(f64, 100) * ((@as(f64, 1) - score))), .segments = @as(f64, @floatFromInt(segments)) };
}
