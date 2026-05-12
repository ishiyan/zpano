/// Threshold crossing signals.
///
/// Fuzzy membership for indicator values relative to fixed thresholds.
const fuzzy = @import("fuzzy");
const membership = fuzzy.membership;

pub const MembershipShape = membership.MembershipShape;

/// Degree to which value is above threshold.
pub fn muAbove(value: f64, threshold: f64, width: f64, shape: MembershipShape) f64 {
    return membership.muGreater(value, threshold, width, shape);
}

/// Degree to which value is below threshold. Complement of muAbove.
pub fn muBelow(value: f64, threshold: f64, width: f64, shape: MembershipShape) f64 {
    return membership.muLess(value, threshold, width, shape);
}

/// Degree of overbought condition.
pub fn muOverbought(value: f64, level: f64, width: f64, shape: MembershipShape) f64 {
    return membership.muGreater(value, level, width, shape);
}

/// Degree of oversold condition.
pub fn muOversold(value: f64, level: f64, width: f64, shape: MembershipShape) f64 {
    return membership.muLess(value, level, width, shape);
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------
const std = @import("std");

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

test "muAbove well above" { try std.testing.expect(almostEqual(muAbove(80.0, 70.0, 5.0, .sigmoid), 1.0, 0.01)); }
test "muAbove well below" { try std.testing.expect(almostEqual(muAbove(60.0, 70.0, 5.0, .sigmoid), 0.0, 0.01)); }
test "muAbove at threshold" { try std.testing.expect(almostEqual(muAbove(70.0, 70.0, 5.0, .sigmoid), 0.5, 1e-10)); }
test "muAbove zero width" { try std.testing.expectEqual(@as(f64, 1.0), muAbove(70.1, 70.0, 0.0, .sigmoid)); }
test "muBelow well below" { try std.testing.expect(almostEqual(muBelow(20.0, 30.0, 5.0, .sigmoid), 1.0, 0.01)); }
test "muBelow at threshold" { try std.testing.expect(almostEqual(muBelow(30.0, 30.0, 5.0, .sigmoid), 0.5, 1e-10)); }
test "complement" {
    for ([_]f64{ 25.0, 30.0, 35.0, 50.0 }) |v| {
        const total = muBelow(v, 30.0, 5.0, .sigmoid) + muAbove(v, 30.0, 5.0, .sigmoid);
        try std.testing.expect(almostEqual(total, 1.0, 1e-10));
    }
}
test "overbought high" { try std.testing.expect(muOverbought(85.0, 70.0, 5.0, .sigmoid) > 0.95); }
test "oversold low" { try std.testing.expect(muOversold(15.0, 30.0, 5.0, .sigmoid) > 0.95); }
test "overbought custom" { try std.testing.expect(almostEqual(muOverbought(80.0, 80.0, 5.0, .sigmoid), 0.5, 1e-10)); }
test "oversold custom" { try std.testing.expect(almostEqual(muOversold(20.0, 20.0, 5.0, .sigmoid), 0.5, 1e-10)); }
