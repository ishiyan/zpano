const std = @import("std");
const data = @import("testdata.zig");
const annual = @import("../annual_cycle_model/model.zig");
const superposition = @import("../harmonic_energy_superposition_wave/model.zig");
fn checkValue(expected: f64, actual: f64) !void {
    try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
}
test "annual-cycle-pattern-decoder prepared model 0" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const a = arena.allocator();
    const input = data.testInput();
    const model = try annual.fitAnnualValues(a, &input, 2020, .{ .candidate_pool = 10, .genuine_threshold = 49.0, .max_components = 7, .max_period = 125, .min_period = 5, .peak_distance = 2, .smoothing = 1600.0 });
    try std.testing.expect(@abs(model.correlation - data.expectedP05_P1125_L1600_G49_C10_K7_D2_CORRELATION()[0]) < 1e-13);
    try std.testing.expectEqual(data.expectedP05_P1125_L1600_G49_C10_K7_D2_COMPONENT_COUNT()[0], @as(f64, @floatFromInt(model.components.len)));
    var reader = try annual.AnnualWaveReader.init(a, model);
    defer reader.deinit();
    for (data.expectedP05_P1125_L1600_G49_C10_K7_D2_WAVE()) |expected| try checkValue(expected, try reader.update());
    reader.reset();
    try checkValue(data.expectedP05_P1125_L1600_G49_C10_K7_D2_WAVE()[0], try reader.update());
}
test "annual-cycle-pattern-decoder prepared model 1" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const a = arena.allocator();
    const input = data.testInput();
    const model = try annual.fitAnnualValues(a, &input, 2020, .{ .candidate_pool = 10, .genuine_threshold = 49.0, .max_components = 1, .max_period = 125, .min_period = 5, .peak_distance = 2, .smoothing = 1600.0 });
    try std.testing.expect(@abs(model.correlation - data.expectedP05_P1125_L1600_G49_C10_K1_D2_CORRELATION()[0]) < 1e-13);
    try std.testing.expectEqual(data.expectedP05_P1125_L1600_G49_C10_K1_D2_COMPONENT_COUNT()[0], @as(f64, @floatFromInt(model.components.len)));
    var reader = try annual.AnnualWaveReader.init(a, model);
    defer reader.deinit();
    for (data.expectedP05_P1125_L1600_G49_C10_K1_D2_WAVE()) |expected| try checkValue(expected, try reader.update());
    reader.reset();
    try checkValue(data.expectedP05_P1125_L1600_G49_C10_K1_D2_WAVE()[0], try reader.update());
}
test "annual-cycle-pattern-decoder prepared model 2" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const a = arena.allocator();
    const input = data.testInput();
    const model = try annual.fitAnnualValues(a, &input, 2020, .{ .candidate_pool = 7, .genuine_threshold = 49.0, .max_components = 3, .max_period = 80, .min_period = 5, .peak_distance = 2, .smoothing = 1600.0 });
    try std.testing.expect(@abs(model.correlation - data.expectedP05_P180_L1600_G49_C7_K3_D2_CORRELATION()[0]) < 1e-13);
    try std.testing.expectEqual(data.expectedP05_P180_L1600_G49_C7_K3_D2_COMPONENT_COUNT()[0], @as(f64, @floatFromInt(model.components.len)));
    var reader = try annual.AnnualWaveReader.init(a, model);
    defer reader.deinit();
    for (data.expectedP05_P180_L1600_G49_C7_K3_D2_WAVE()) |expected| try checkValue(expected, try reader.update());
    reader.reset();
    try checkValue(data.expectedP05_P180_L1600_G49_C7_K3_D2_WAVE()[0], try reader.update());
}
test "annual-cycle-pattern-decoder prepared model 3" {
    var arena = std.heap.ArenaAllocator.init(std.testing.allocator);
    defer arena.deinit();
    const a = arena.allocator();
    const input = data.testInput();
    const model = try annual.fitAnnualValues(a, &input, 2020, .{ .candidate_pool = 10, .genuine_threshold = 49.0, .max_components = 7, .max_period = 80, .min_period = 5, .peak_distance = 2, .smoothing = 16000.0 });
    try std.testing.expect(@abs(model.correlation - data.expectedP05_P180_L16000_G49_C10_K7_D2_CORRELATION()[0]) < 1e-13);
    try std.testing.expectEqual(data.expectedP05_P180_L16000_G49_C10_K7_D2_COMPONENT_COUNT()[0], @as(f64, @floatFromInt(model.components.len)));
    var reader = try annual.AnnualWaveReader.init(a, model);
    defer reader.deinit();
    for (data.expectedP05_P180_L16000_G49_C10_K7_D2_WAVE()) |expected| try checkValue(expected, try reader.update());
    reader.reset();
    try checkValue(data.expectedP05_P180_L16000_G49_C10_K7_D2_WAVE()[0], try reader.update());
}
