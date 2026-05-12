/// Crossover signals.
const fuzzy = @import("fuzzy");
const membership = fuzzy.membership;
const operators = fuzzy.operators;

const MembershipShape = membership.MembershipShape;

/// Degree to which a value crossed above threshold from below.
pub fn muCrossesAbove(prev_value: f64, curr_value: f64, threshold: f64, width: f64, shape: MembershipShape) f64 {
    const was_below = membership.muLess(prev_value, threshold, width, shape);
    const is_above = membership.muGreater(curr_value, threshold, width, shape);
    return operators.tProduct(was_below, is_above);
}

/// Degree to which a value crossed below threshold from above.
pub fn muCrossesBelow(prev_value: f64, curr_value: f64, threshold: f64, width: f64, shape: MembershipShape) f64 {
    const was_above = membership.muGreater(prev_value, threshold, width, shape);
    const is_below = membership.muLess(curr_value, threshold, width, shape);
    return operators.tProduct(was_above, is_below);
}

/// Degree to which a fast line crossed above a slow line.
pub fn muLineCrossesAbove(prev_fast: f64, curr_fast: f64, prev_slow: f64, curr_slow: f64, width: f64, shape: MembershipShape) f64 {
    const prev_diff = prev_fast - prev_slow;
    const curr_diff = curr_fast - curr_slow;
    return muCrossesAbove(prev_diff, curr_diff, 0.0, width, shape);
}

/// Degree to which a fast line crossed below a slow line.
pub fn muLineCrossesBelow(prev_fast: f64, curr_fast: f64, prev_slow: f64, curr_slow: f64, width: f64, shape: MembershipShape) f64 {
    const prev_diff = prev_fast - prev_slow;
    const curr_diff = curr_fast - curr_slow;
    return muCrossesBelow(prev_diff, curr_diff, 0.0, width, shape);
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------
const std = @import("std");

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

test "clear cross above" { try std.testing.expect(almostEqual(muCrossesAbove(25.0, 35.0, 30.0, 0.0, .sigmoid), 1.0, 1e-10)); }
test "no cross both above" { try std.testing.expect(almostEqual(muCrossesAbove(35.0, 40.0, 30.0, 0.0, .sigmoid), 0.0, 1e-10)); }
test "no cross both below" { try std.testing.expect(almostEqual(muCrossesAbove(25.0, 28.0, 30.0, 0.0, .sigmoid), 0.0, 1e-10)); }
test "at threshold" { try std.testing.expect(almostEqual(muCrossesAbove(30.0, 30.0, 30.0, 0.0, .sigmoid), 0.25, 1e-10)); }
test "clear cross below" { try std.testing.expect(almostEqual(muCrossesBelow(35.0, 25.0, 30.0, 0.0, .sigmoid), 1.0, 1e-10)); }
test "symmetry" {
    const cb = muCrossesBelow(35.0, 25.0, 30.0, 2.0, .sigmoid);
    const ca = muCrossesAbove(25.0, 35.0, 30.0, 2.0, .sigmoid);
    try std.testing.expect(almostEqual(cb, ca, 1e-10));
}
test "golden cross" { try std.testing.expect(almostEqual(muLineCrossesAbove(49.0, 51.0, 50.0, 50.0, 0.0, .sigmoid), 1.0, 1e-10)); }
test "no line cross" { try std.testing.expect(almostEqual(muLineCrossesAbove(52.0, 53.0, 50.0, 50.0, 0.0, .sigmoid), 0.0, 1e-10)); }
test "death cross" { try std.testing.expect(almostEqual(muLineCrossesBelow(51.0, 49.0, 50.0, 50.0, 0.0, .sigmoid), 1.0, 1e-10)); }
