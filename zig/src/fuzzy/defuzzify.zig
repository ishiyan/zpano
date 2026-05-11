/// Defuzzification utilities.
///
/// Provides alpha-cut conversion from continuous fuzzy output back to crisp
/// discrete values for backward compatibility with TA-Lib-style integer outputs.
const std = @import("std");

/// Convert a continuous fuzzy output to a crisp discrete value.
///
/// The confidence is abs(value) / scale. If confidence ≥ alpha,
/// the output is rounded to the nearest multiple of scale with the
/// original sign preserved. Otherwise 0 is returned.
pub fn alphaCut(value: f64, alpha: f64, scale: f64) i32 {
    if (scale <= 0.0) return 0;
    const confidence = @abs(value) / scale;
    if (confidence < alpha - 1e-10) return 0;
    const sign: i32 = if (value >= 0.0) 1 else -1;
    // Round to nearest multiple of scale.
    const level = @max(1.0, @round(confidence));
    return sign * @as(i32, @intFromFloat(level * scale));
}

// -----------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------

test "alpha cut strong bearish" { try std.testing.expectEqual(@as(i32, -100), alphaCut(-87.3, 0.5, 100.0)); }
test "alpha cut weak bearish" { try std.testing.expectEqual(@as(i32, 0), alphaCut(-32.1, 0.5, 100.0)); }
test "alpha cut strong bullish" { try std.testing.expectEqual(@as(i32, 100), alphaCut(92.5, 0.5, 100.0)); }
test "alpha cut weak bullish" { try std.testing.expectEqual(@as(i32, 0), alphaCut(15.0, 0.5, 100.0)); }
test "alpha cut zero" { try std.testing.expectEqual(@as(i32, 0), alphaCut(0.0, 0.5, 100.0)); }
test "alpha cut strong confirmation" { try std.testing.expectEqual(@as(i32, 200), alphaCut(156.8, 0.5, 100.0)); }
test "alpha cut negative confirmation" { try std.testing.expectEqual(@as(i32, -200), alphaCut(-180.0, 0.5, 100.0)); }
test "alpha cut high alpha filters" { try std.testing.expectEqual(@as(i32, 0), alphaCut(-87.3, 0.9, 100.0)); }
test "alpha cut high alpha passes" { try std.testing.expectEqual(@as(i32, -100), alphaCut(-95.0, 0.9, 100.0)); }
test "alpha cut low alpha passes" { try std.testing.expectEqual(@as(i32, -100), alphaCut(-15.0, 0.1, 100.0)); }
test "alpha cut zero alpha" { try std.testing.expectEqual(@as(i32, -100), alphaCut(-1.0, 0.0, 100.0)); }
test "alpha cut exactly at threshold" { try std.testing.expectEqual(@as(i32, 100), alphaCut(50.0, 0.5, 100.0)); }
test "alpha cut just below threshold" { try std.testing.expectEqual(@as(i32, 0), alphaCut(49.9, 0.5, 100.0)); }
test "alpha cut exactly 100" { try std.testing.expectEqual(@as(i32, 100), alphaCut(100.0, 0.5, 100.0)); }
test "alpha cut exactly minus 100" { try std.testing.expectEqual(@as(i32, -100), alphaCut(-100.0, 0.5, 100.0)); }
test "alpha cut custom scale" { try std.testing.expectEqual(@as(i32, -50), alphaCut(-40.0, 0.5, 50.0)); }
test "alpha cut invalid scale" { try std.testing.expectEqual(@as(i32, 0), alphaCut(-87.3, 0.5, 0.0)); }
