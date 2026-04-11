const std = @import("std");
const testing = std.testing;

/// Represents an OHLCV price bar.
pub const Bar = struct {
    time: i64,
    open: f64,
    high: f64,
    low: f64,
    close: f64,
    volume: f64,

    /// Indicates whether this is a rising bar (open < close).
    pub fn isRising(self: Bar) bool {
        return self.open < self.close;
    }

    /// Indicates whether this is a falling bar (close < open).
    pub fn isFalling(self: Bar) bool {
        return self.close < self.open;
    }

    /// The median price: (low + high) / 2.
    pub fn median(self: Bar) f64 {
        return (self.low + self.high) / 2.0;
    }

    /// The typical price: (low + high + close) / 3.
    pub fn typical(self: Bar) f64 {
        return (self.low + self.high + self.close) / 3.0;
    }

    /// The weighted price: (low + high + 2*close) / 4.
    pub fn weighted(self: Bar) f64 {
        return (self.low + self.high + self.close + self.close) / 4.0;
    }

    /// The average price: (low + high + open + close) / 4.
    pub fn average(self: Bar) f64 {
        return (self.low + self.high + self.open + self.close) / 4.0;
    }
};

fn makeBar(o: f64, h: f64, l: f64, c: f64, v: f64) Bar {
    return Bar{ .time = 0, .open = o, .high = h, .low = l, .close = c, .volume = v };
}

test "bar median" {
    const b = makeBar(0, 3, 2, 0, 0);
    try testing.expectEqual((b.low + b.high) / 2.0, b.median());
}

test "bar typical" {
    const b = makeBar(0, 4, 2, 3, 0);
    try testing.expectEqual((b.low + b.high + b.close) / 3.0, b.typical());
}

test "bar weighted" {
    const b = makeBar(0, 4, 2, 3, 0);
    try testing.expectEqual((b.low + b.high + b.close + b.close) / 4.0, b.weighted());
}

test "bar average" {
    const b = makeBar(3, 5, 2, 4, 0);
    try testing.expectEqual((b.low + b.high + b.open + b.close) / 4.0, b.average());
}

test "bar isRising" {
    try testing.expect(makeBar(2, 0, 0, 3, 0).isRising());
    try testing.expect(!makeBar(3, 0, 0, 2, 0).isRising());
    try testing.expect(!makeBar(0, 0, 0, 0, 0).isRising());
}

test "bar isFalling" {
    try testing.expect(!makeBar(2, 0, 0, 3, 0).isFalling());
    try testing.expect(makeBar(3, 0, 0, 2, 0).isFalling());
    try testing.expect(!makeBar(0, 0, 0, 0, 0).isFalling());
}
