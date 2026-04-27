const std = @import("std");

/// Represents two band values and a time stamp.
pub const Band = struct {
    time: i64,
    lower: f64,
    upper: f64,

    /// Creates a new band. Values are sorted so lower <= upper.
    pub fn new(time: i64, lower: f64, upper: f64) Band {
        if (lower < upper) {
            return .{ .time = time, .lower = lower, .upper = upper };
        } else {
            return .{ .time = time, .lower = upper, .upper = lower };
        }
    }

    /// Creates a new empty band with NaN values.
    pub fn empty(time: i64) Band {
        return .{ .time = time, .lower = std.math.nan(f64), .upper = std.math.nan(f64) };
    }

    /// Indicates whether this band is not initialized.
    pub fn isEmpty(self: Band) bool {
        return std.math.isNan(self.lower) or std.math.isNan(self.upper);
    }
};
