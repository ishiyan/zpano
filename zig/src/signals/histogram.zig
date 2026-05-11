/// Histogram sign-change signals.
const membership = @import("membership");
const operators = @import("operators");

const MembershipShape = membership.MembershipShape;

/// Degree to which a histogram turned from non-positive to positive.
pub fn muTurnsPositive(prev_value: f64, curr_value: f64, width: f64, shape: MembershipShape) f64 {
    const was_nonpositive = membership.muLess(prev_value, 0.0, width, shape);
    const is_positive = membership.muGreater(curr_value, 0.0, width, shape);
    return operators.tProduct(was_nonpositive, is_positive);
}

/// Degree to which a histogram turned from non-negative to negative.
pub fn muTurnsNegative(prev_value: f64, curr_value: f64, width: f64, shape: MembershipShape) f64 {
    const was_nonnegative = membership.muGreater(prev_value, 0.0, width, shape);
    const is_negative = membership.muLess(curr_value, 0.0, width, shape);
    return operators.tProduct(was_nonnegative, is_negative);
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------
const std = @import("std");

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

test "clear turn positive" { try std.testing.expect(almostEqual(muTurnsPositive(-5.0, 5.0, 0.0, .sigmoid), 1.0, 1e-10)); }
test "stays positive" { try std.testing.expect(almostEqual(muTurnsPositive(3.0, 5.0, 0.0, .sigmoid), 0.0, 1e-10)); }
test "stays negative" { try std.testing.expect(almostEqual(muTurnsPositive(-5.0, -3.0, 0.0, .sigmoid), 0.0, 1e-10)); }
test "from zero" { try std.testing.expect(almostEqual(muTurnsPositive(0.0, 5.0, 0.0, .sigmoid), 0.5, 1e-10)); }
test "clear turn negative" { try std.testing.expect(almostEqual(muTurnsNegative(5.0, -5.0, 0.0, .sigmoid), 1.0, 1e-10)); }
test "symmetry" {
    const tn = muTurnsNegative(3.0, -3.0, 1.0, .sigmoid);
    const tp = muTurnsPositive(-3.0, 3.0, 1.0, .sigmoid);
    try std.testing.expect(almostEqual(tn, tp, 1e-10));
}
