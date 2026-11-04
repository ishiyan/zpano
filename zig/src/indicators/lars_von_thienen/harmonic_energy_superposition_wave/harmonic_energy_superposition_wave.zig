const std = @import("std");
const data = @import("testdata.zig");
const annual = @import("../annual_cycle_model/model.zig");
const superposition = @import("../harmonic_energy_superposition_wave/model.zig");
fn checkValue(expected: f64, actual: f64) !void {
    try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
}
fn historicalModels(a: std.mem.Allocator) ![6]annual.Model {
    var models: [6]annual.Model = undefined;
    const input = data.testInput();
    for (superposition.metonic_lags, 0..) |lag, j| {
        var values: [252]f64 = undefined;
        for (&values, 0..) |*v, i| v.* = (1 + 0.1 * @as(f64, @floatFromInt(j))) * input[(i + 7 * j) % 252];
        models[j] = try annual.fitAnnualValues(a, &values, 2021 - lag, .{ .max_period = 80 });
    }
    return models;
}
test "harmonic-energy-superposition-wave prepared model 0" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const a = arena.allocator();
    var models = try historicalModels(a);
    _ = &models;
    var reader = try superposition.SuperpositionWaveReader.init(a, &models, 2021, 252, "none", "strict", false);
    defer reader.deinit();
    for (data.expectedN0_C0_I0_WAVE()) |expected| try checkValue(expected, reader.update().?);
    reader.reset();
    try checkValue(data.expectedN0_C0_I0_WAVE()[0], reader.update().?);
}
test "harmonic-energy-superposition-wave prepared model 1" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const a = arena.allocator();
    var models = try historicalModels(a);
    _ = &models;
    var reader = try superposition.SuperpositionWaveReader.init(a, &models, 2021, 252, "rms", "strict", false);
    defer reader.deinit();
    for (data.expectedN1_C0_I0_WAVE()) |expected| try checkValue(expected, reader.update().?);
    reader.reset();
    try checkValue(data.expectedN1_C0_I0_WAVE()[0], reader.update().?);
}
test "harmonic-energy-superposition-wave prepared model 2" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const a = arena.allocator();
    var models = try historicalModels(a);
    _ = &models;
    var reader = try superposition.SuperpositionWaveReader.init(a, &models, 2021, 252, "rms", "strict", true);
    defer reader.deinit();
    for (data.expectedN1_C0_I1_WAVE()) |expected| try checkValue(expected, reader.update().?);
    reader.reset();
    try checkValue(data.expectedN1_C0_I1_WAVE()[0], reader.update().?);
}
test "harmonic-energy-superposition-wave prepared model 3" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const a = arena.allocator();
    var models = try historicalModels(a);
    for (&models) |*m| m.sample_count = 250;
    var reader = try superposition.SuperpositionWaveReader.init(a, &models, 2021, 252, "rms", "extend", false);
    defer reader.deinit();
    for (data.expectedN1_C1_I0_WAVE()) |expected| try checkValue(expected, reader.update().?);
    reader.reset();
    try checkValue(data.expectedN1_C1_I0_WAVE()[0], reader.update().?);
}
