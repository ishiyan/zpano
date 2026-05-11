/// Band signals.
const membership = @import("membership");
const operators = @import("operators");

const MembershipShape = membership.MembershipShape;

/// Degree to which value is above the upper band.
pub fn muAboveBand(value: f64, upper_band: f64, width: f64, shape: MembershipShape) f64 {
    return membership.muGreater(value, upper_band, width, shape);
}

/// Degree to which value is below the lower band.
pub fn muBelowBand(value: f64, lower_band: f64, width: f64, shape: MembershipShape) f64 {
    return membership.muLess(value, lower_band, width, shape);
}

/// Degree to which value is inside the band channel.
pub fn muBetweenBands(value: f64, lower_band: f64, upper_band: f64, shape: MembershipShape) f64 {
    if (upper_band <= lower_band) return 0.0;
    const spread = upper_band - lower_band;
    const width = spread * 0.5;
    const above_lower = membership.muGreater(value, lower_band, width, shape);
    const below_upper = membership.muLess(value, upper_band, width, shape);
    return operators.tProduct(above_lower, below_upper);
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------
const std = @import("std");

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

test "above well above" { try std.testing.expect(almostEqual(muAboveBand(110.0, 100.0, 5.0, .sigmoid), 1.0, 0.01)); }
test "above at band" { try std.testing.expect(almostEqual(muAboveBand(100.0, 100.0, 5.0, .sigmoid), 0.5, 1e-10)); }
test "below well below" { try std.testing.expect(almostEqual(muBelowBand(85.0, 90.0, 5.0, .sigmoid), 1.0, 0.01)); }
test "between centered" { try std.testing.expect(muBetweenBands(100.0, 90.0, 110.0, .sigmoid) > 0.8); }
test "between at upper" { try std.testing.expect(muBetweenBands(110.0, 90.0, 110.0, .sigmoid) < 0.6); }
test "between outside" { try std.testing.expect(muBetweenBands(130.0, 90.0, 110.0, .sigmoid) < 0.1); }
test "between degenerate" {
    try std.testing.expectEqual(@as(f64, 0.0), muBetweenBands(100.0, 110.0, 90.0, .sigmoid));
    try std.testing.expectEqual(@as(f64, 0.0), muBetweenBands(100.0, 100.0, 100.0, .sigmoid));
}
test "between monotonic" {
    const center = muBetweenBands(100.0, 90.0, 110.0, .sigmoid);
    const edge = muBetweenBands(108.0, 90.0, 110.0, .sigmoid);
    const outside = muBetweenBands(115.0, 90.0, 110.0, .sigmoid);
    try std.testing.expect(center > edge);
    try std.testing.expect(edge > outside);
}
