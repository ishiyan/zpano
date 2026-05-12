/// Signal composition utilities.
const fuzzy = @import("fuzzy");
const operators = fuzzy.operators;

/// Combine signals with product t-norm (fuzzy AND).
pub fn signalAnd(signals: []const f64) f64 {
    return operators.tProductAll(signals);
}

/// Combine two signals with probabilistic s-norm (fuzzy OR).
pub fn signalOr(a: f64, b: f64) f64 {
    return operators.sProbabilistic(a, b);
}

/// Negate a signal (fuzzy complement). Returns 1 - signal.
pub fn signalNot(signal: f64) f64 {
    return operators.fNot(signal);
}

/// Filter weak signals below min_strength to zero.
/// Signals at or above the threshold pass through unchanged.
pub fn signalStrength(signal: f64, min_strength: f64) f64 {
    return if (signal >= min_strength) signal else 0.0;
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------
const std = @import("std");

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

test "and all high" { try std.testing.expect(almostEqual(signalAnd(&[_]f64{ 0.9, 0.8, 0.95 }), 0.9 * 0.8 * 0.95, 1e-10)); }
test "and one zero" { try std.testing.expect(almostEqual(signalAnd(&[_]f64{ 0.9, 0.0, 0.8 }), 0.0, 1e-10)); }
test "and all one" { try std.testing.expect(almostEqual(signalAnd(&[_]f64{ 1.0, 1.0, 1.0 }), 1.0, 1e-10)); }
test "and two" { try std.testing.expect(almostEqual(signalAnd(&[_]f64{ 0.6, 0.7 }), 0.42, 1e-10)); }
test "or both high" { try std.testing.expect(almostEqual(signalOr(0.8, 0.9), 0.8 + 0.9 - 0.8 * 0.9, 1e-10)); }
test "or one zero" { try std.testing.expect(almostEqual(signalOr(0.0, 0.7), 0.7, 1e-10)); }
test "or both one" { try std.testing.expect(almostEqual(signalOr(1.0, 1.0), 1.0, 1e-10)); }
test "not zero" { try std.testing.expect(almostEqual(signalNot(0.0), 1.0, 1e-10)); }
test "not one" { try std.testing.expect(almostEqual(signalNot(1.0), 0.0, 1e-10)); }
test "not half" { try std.testing.expect(almostEqual(signalNot(0.5), 0.5, 1e-10)); }
test "strength above" { try std.testing.expectEqual(@as(f64, 0.8), signalStrength(0.8, 0.5)); }
test "strength below" { try std.testing.expectEqual(@as(f64, 0.0), signalStrength(0.3, 0.5)); }
test "strength at" { try std.testing.expectEqual(@as(f64, 0.5), signalStrength(0.5, 0.5)); }
test "strength just below" { try std.testing.expectEqual(@as(f64, 0.0), signalStrength(0.499, 0.5)); }
