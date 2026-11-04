const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
// Native Go numerical port. Source conventions: description.md.
pub const CycleModel = struct {
    period: f64 = 0,
    amplitude: f64 = 0,
    phase: f64 = 0,
    trough_offset: f64 = 0,
};
pub fn goertzel(allocator: std.mem.Allocator, values: []f64, period: f64) anyerror!CycleModel {
    try @import("validation.zig").real(allocator, period, "period", 2.0000000000000004);
    const n = @as(isize, @intCast(values.len));
    if ((n < 1)) {
        return error.InvalidInput;
    }
    const omega = ((@as(f64, 2) * std.math.pi) / period);
    const temporary1_0 = @sin(omega);
    const temporary1_1 = @cos(omega);
    const sine = temporary1_0;
    const cosine = temporary1_1;
    const coefficient = (@as(f64, 2) * cosine);
    var q1: f64 = 0;
    var q2: f64 = 0;

    for (values) |value| {
        (try @import("validation.zig").sample(allocator, value));
        if (std.math.isNan(value)) {
            return error.InvalidInput;
        }
        const q0 = (try @import("validation.zig").finite(allocator, (((coefficient * q1) - q2) + value)));
        const temporary2_0 = q1;
        const temporary2_1 = q0;
        q2 = temporary2_0;
        q1 = temporary2_1;
    }
    const temporary3_0 = (q1 - (q2 * cosine));
    const temporary3_1 = (q2 * sine);
    const r1 = temporary3_0;
    const quadrature = temporary3_1;
    const temporary4_0 = @cos((@as(f64, @floatFromInt((n - 1))) * omega));
    const temporary4_1 = @sin((@as(f64, @floatFromInt((n - 1))) * omega));
    const cn = temporary4_0;
    const sn = temporary4_1;
    const realPart = (try @import("validation.zig").finite(allocator, ((r1 * cn) + (quadrature * sn))));
    const imaginary = (try @import("validation.zig").finite(allocator, ((quadrature * cn) - (r1 * sn))));
    const amplitude = (try @import("validation.zig").finite(allocator, ((@as(f64, 2) * @sqrt((try @import("validation.zig").finite(allocator, ((realPart * realPart) + (imaginary * imaginary)))))) / @as(f64, @floatFromInt(n)))));
    if ((amplitude == @as(f64, 0))) {
        return CycleModel{ .period = period, .amplitude = @as(f64, 0), .phase = std.math.nan(f64), .trough_offset = std.math.nan(f64) };
    }
    const phase = ((std.math.pi / @as(f64, 2)) + std.math.atan2(imaginary, realPart));
    const offset = (period * ((0.75 - (phase / ((@as(f64, 2) * std.math.pi))))));
    return CycleModel{ .period = period, .amplitude = amplitude, .phase = phase, .trough_offset = offset };
}
pub fn cycleValue(_: std.mem.Allocator, model: CycleModel, offset: f64) anyerror!f64 {
    if ((model.amplitude == @as(f64, 0))) {
        return 0;
    }
    return (model.amplitude * @sin(((((@as(f64, 2) * std.math.pi) * offset) / model.period) + model.phase)));
}
pub fn cycleSlope(_: std.mem.Allocator, model: CycleModel, offset: f64) anyerror!f64 {
    if ((model.amplitude == @as(f64, 0))) {
        return 0;
    }
    const omega = ((@as(f64, 2) * std.math.pi) / model.period);
    return ((model.amplitude * omega) * @cos(((omega * offset) + model.phase)));
}
pub fn compositeValue(allocator: std.mem.Allocator, models: []CycleModel, offset: f64) anyerror!f64 {
    var sum: f64 = 0;

    for (models) |model| {
        sum += (try cycleValue(allocator, model, offset));
    }
    return (try @import("validation.zig").finite(allocator, sum));
}
pub fn cycleStrength(allocator: std.mem.Allocator, amplitude: f64, period: f64) anyerror!f64 {
    try @import("validation.zig").real(allocator, amplitude, "amplitude", 0);
    try @import("validation.zig").real(allocator, period, "period", 2.0000000000000004);
    return (amplitude / period);
}
pub fn cyclePolarity(allocator: std.mem.Allocator, shortSlope: f64, longSlope: f64, tolerance: f64) anyerror!f64 {
    (try @import("validation.zig").sample(allocator, shortSlope));
    (try @import("validation.zig").sample(allocator, longSlope));
    try @import("validation.zig").real(allocator, tolerance, "tolerance", 0);
    if ((std.math.isNan(shortSlope) or std.math.isNan(longSlope))) {
        return std.math.nan(f64);
    }
    if (((shortSlope > tolerance) and (longSlope > tolerance))) {
        return 1;
    }
    if (((shortSlope < -tolerance) and (longSlope < -tolerance))) {
        return -1;
    }
    return 0;
}
pub fn phaseState(allocator: std.mem.Allocator, phase: f64) anyerror!f64 {
    (try @import("validation.zig").sample(allocator, phase));
    return (@as(f64, 100) * @sin(phase));
}

pub const real = @import("validation.zig").real;
pub const integer = @import("validation.zig").integer;
pub const finite = @import("validation.zig").finite;
pub const sample = @import("validation.zig").sample;
pub const push = @import("validation.zig").push;
