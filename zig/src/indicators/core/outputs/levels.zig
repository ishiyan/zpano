const std = @import("std");

/// A single entry of a Levels output, expressed as a value with an optional bar
/// offset and an optional strength.
pub const Level = struct {
    /// The value (e.g. a price level or a multiplier) at this entry.
    value: f64,
    /// The number of bars back from the Levels' time (0 = current bar).
    offset: i32 = 0,
    /// Optional significance measure (higher = more significant); NaN when not applicable.
    strength: f64 = std.math.nan(f64),
};

/// Maximum number of entries in a Levels output.
pub const max_levels = 256;

/// Holds a time stamp and a variable-length set of levels.
pub const Levels = struct {
    time: i64,
    levels: [max_levels]Level = undefined,
    levels_len: usize = 0,

    /// Creates a new Levels from a slice of entries.
    pub fn new(time: i64, levels: []const Level) Levels {
        var l = Levels{
            .time = time,
            .levels_len = @min(levels.len, max_levels),
        };
        @memcpy(l.levels[0..l.levels_len], levels[0..l.levels_len]);
        return l;
    }

    /// Creates a new empty Levels with no entries.
    pub fn empty(time: i64) Levels {
        return .{ .time = time };
    }

    /// Indicates whether this Levels has no entries.
    pub fn isEmpty(self: Levels) bool {
        return self.levels_len == 0;
    }

    /// Returns the entries as a slice.
    pub fn levelsSlice(self: *const Levels) []const Level {
        return self.levels[0..self.levels_len];
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

test "Levels new and slice" {
    const entries = [_]Level{
        .{ .value = 105.5, .offset = 3, .strength = 0.8 },
        .{ .value = 102.0, .offset = 1, .strength = 0.6 },
    };
    const l = Levels.new(42, &entries);
    try testing.expectEqual(@as(i64, 42), l.time);
    try testing.expectEqual(@as(usize, 2), l.levels_len);
    try testing.expectEqual(@as(f64, 105.5), l.levelsSlice()[0].value);
    try testing.expectEqual(@as(i32, 3), l.levelsSlice()[0].offset);
    try testing.expectEqual(@as(f64, 0.8), l.levelsSlice()[0].strength);
    try testing.expect(!l.isEmpty());
}

test "Levels empty" {
    const l = Levels.empty(7);
    try testing.expectEqual(@as(i64, 7), l.time);
    try testing.expect(l.isEmpty());
    try testing.expectEqual(@as(usize, 0), l.levelsSlice().len);
}

test "Level defaults" {
    const lv = Level{ .value = 42.0 };
    try testing.expectEqual(@as(i32, 0), lv.offset);
    try testing.expect(std.math.isNan(lv.strength));
}
