const std = @import("std");
const testing = std.testing;
const Measures = @import("measures.zig").Measures;

test "review gain-to-pain repetition and rolling" {
    for ([_]usize{ 0, 2, 4 }) |window| {
        var m = try Measures.init(testing.allocator, 1, 0, 0, window);
        defer m.deinit();
        try testing.expect(std.math.isNan(m.gainToPainRatio()));
        for (0..4) |_| {
            try m.addReturn(0.1, 0);
            try m.addReturn(-0.05, 0);
            try testing.expectApproxEqAbs(@as(f64, 1), m.gainToPainRatio(), 5e-14);
        }
        m.reset();
        try m.addReturn(0.1, 0);
        try testing.expect(std.math.isNan(m.gainToPainRatio()));
    }
}

test "review modified information geometric sign" {
    var m = try Measures.init(testing.allocator, 1, 0, 0, 0);
    defer m.deinit();
    try m.addReturn(0.5, 0.05);
    try m.addReturn(-0.3, 0.05);
    try testing.expect(m.activePremium() < 0);
    try testing.expectApproxEqAbs(@as(f64, 0.04473320734100739), m.informationRatioModified(), 5e-14);
}

test "review validation before data fallback" {
    var m = try Measures.init(testing.allocator, 1, 0, 0, 0);
    defer m.deinit();
    for ([_]bool{ false, true }) |populated| {
        if (populated) {
            try m.addReturn(-0.1, 0);
            try m.addReturn(0.2, 0);
        }
        for ([_]f64{ -1e20, -1, 0, 1, 2, std.math.nan(f64), std.math.inf(f64), -std.math.inf(f64) }) |confidence| {
            try testing.expectError(error.InvalidArgument, m.isNormalDistribution(confidence));
            try testing.expectError(error.InvalidArgument, m.rewardToConditionalDrawdown(confidence));
        }
        for ([_]f64{ 0, -1, std.math.nan(f64), -std.math.inf(f64) }) |multiplier| {
            try testing.expectError(error.InvalidArgument, m.biasRatio(multiplier));
        }
    }
    m.reset();
    try testing.expect(std.math.isNan(try m.rewardToConditionalDrawdown(0.95)));
}
