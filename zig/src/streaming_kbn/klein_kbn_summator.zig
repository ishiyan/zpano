const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;

/// Errors returned by `KleinKBNSummator`.
pub const Error = error{EmptyRevert};

/// Counted compensated sum: a KleinKBNAccumulator plus a sample count,
/// which also gives the arithmetic mean.
///
/// See KleinKBNAccumulator for the summation algorithm.
///
/// Every call to update() increments the count, including calls with
/// x == 0; every call to revert() decrements it.  Because only sums are
/// stored, revert() may remove any previously added value (not only
/// the most recent one), so the summator works for FIFO rolling windows.
pub const KleinKBNSummator = struct {
    count: usize = 0,
    sum: KleinKBNAccumulator = .{},

    const Self = @This();

    /// Clears the count and the sum.
    pub fn reset(self: *Self) void {
        self.count = 0;
        self.sum.reset();
    }

    /// Removes a previously added value x.
    /// Removing the final sample clears the sum and its compensation terms.
    ///
    /// Returns `error.EmptyRevert` if the summator is empty.
    pub fn revert(self: *Self, x: f64) Error!void {
        if (self.count == 0) return error.EmptyRevert;
        if (self.count == 1) {
            self.reset();
            return;
        }
        self.count -= 1;
        // Adding zero leaves the accumulator unchanged, so skip it.
        if (x != 0) self.sum.revert(x);
    }

    /// Adds a value x.
    pub fn update(self: *Self, x: f64) void {
        self.count += 1;
        // Adding zero leaves the accumulator unchanged, so skip it.
        if (x != 0) self.sum.update(x);
    }

    /// The compensated sum of all added values (0.0 when empty).
    pub fn value(self: *const Self) f64 {
        return self.sum.value();
    }

    /// The arithmetic mean, sum / n (NaN when empty).
    pub fn mean(self: *const Self) f64 {
        if (self.count == 0) return math.nan(f64);
        return self.sum.value() / @as(f64, @floatFromInt(self.count));
    }

    /// The number of added values.
    pub fn n(self: *const Self) usize {
        return self.count;
    }
};

// ── Tests ──────────────────────────────────────────────────────────────────

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

/// Faithful port of CPython's math.fsum (Shewchuk's exact partials with
/// round-half-even correction). Inputs are assumed finite.
fn fsum(data: []const f64) f64 {
    var p: [64]f64 = undefined;
    var np: usize = 0;
    for (data) |xin| {
        var x = xin;
        var i: usize = 0;
        var j: usize = 0;
        while (j < np) : (j += 1) {
            var y = p[j];
            if (@abs(x) < @abs(y)) {
                const tmp = x;
                x = y;
                y = tmp;
            }
            const hi = x + y;
            const lo = y - (hi - x);
            if (lo != 0.0) {
                p[i] = lo;
                i += 1;
            }
            x = hi;
        }
        np = i;
        if (x != 0.0) {
            p[np] = x;
            np += 1;
        }
    }
    var hi: f64 = 0.0;
    if (np > 0) {
        np -= 1;
        hi = p[np];
        var lo: f64 = 0.0;
        while (np > 0) {
            const x = hi;
            np -= 1;
            const y = p[np];
            hi = x + y;
            const yr = hi - x;
            lo = y - yr;
            if (lo != 0.0) break;
        }
        if (np > 0 and ((lo < 0.0 and p[np - 1] < 0.0) or (lo > 0.0 and p[np - 1] > 0.0))) {
            const y = lo * 2.0;
            const x = hi + y;
            const yr = x - hi;
            if (y == yr) hi = x;
        }
    }
    return hi;
}

test "empty" {
    const s = KleinKBNSummator{};
    try testing.expectEqual(@as(usize, 0), s.n());
    try testing.expectEqual(@as(f64, 0.0), s.value());
    try testing.expect(math.isNan(s.mean()));
}

test "update" {
    var s = KleinKBNSummator{};
    for ([_]f64{ 1.0, 2.0, 3.0, 4.0 }) |x| s.update(x);
    try testing.expectEqual(@as(usize, 4), s.n());
    try testing.expectEqual(@as(f64, 10.0), s.value());
    try testing.expectEqual(@as(f64, 2.5), s.mean());
}

test "zero is counted" {
    var s = KleinKBNSummator{};
    for ([_]f64{ 0.0, 3.0, 0.0 }) |x| s.update(x);
    try testing.expectEqual(@as(usize, 3), s.n());
    try testing.expectEqual(@as(f64, 3.0), s.value());
    try testing.expectEqual(@as(f64, 1.0), s.mean());
}

test "compensated" {
    // Peters' example: naive summation yields 0.0.
    var s = KleinKBNSummator{};
    for ([_]f64{ 1.0, 1e100, 1.0, -1e100 }) |x| s.update(x);
    try testing.expectEqual(@as(f64, 2.0), s.value());
    try testing.expectEqual(@as(f64, 0.5), s.mean());
}

test "revert" {
    var s = KleinKBNSummator{};
    for ([_]f64{ 1.0, 2.0, 0.0, 4.0 }) |x| s.update(x);
    try s.revert(0.0);
    try testing.expectEqual(@as(usize, 3), s.n());
    try testing.expectEqual(@as(f64, 7.0), s.value());
    try s.revert(1.0); // not the most recent value
    try testing.expectEqual(@as(usize, 2), s.n());
    try testing.expectEqual(@as(f64, 6.0), s.value());
    try testing.expectEqual(@as(f64, 3.0), s.mean());
}

test "rolling window" {
    const data = [_]f64{ 0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067 };
    const w: usize = 3;
    var s = KleinKBNSummator{};
    for (data, 0..) |x, i| {
        s.update(x);
        if (i >= w) try s.revert(data[i - w]);
        const lo = if (i + 1 >= w) i + 1 - w else 0;
        const window = data[lo .. i + 1];
        try testing.expectEqual(window.len, s.n());
        try testing.expect(almostEqual(s.value(), fsum(window), 1e-17));
    }
}

test "revert to empty" {
    var s = KleinKBNSummator{};
    s.update(5.0);
    try s.revert(5.0);
    try testing.expectEqual(@as(usize, 0), s.n());
    try testing.expectEqual(@as(f64, 0.0), s.value());
    try testing.expect(math.isNan(s.mean()));
}

test "revert empty raises" {
    var s = KleinKBNSummator{};
    try testing.expectError(error.EmptyRevert, s.revert(1.0));
}

test "revert to empty clears compensation" {
    for ([_]bool{ false, true }) |final_zero| {
        var s = KleinKBNSummator{};
        for ([_]f64{ 0.1, 1e16, 1e32, 1e48 }) |x| s.update(x);
        if (final_zero) s.update(0.0);
        for ([_]f64{ 1e16, 0.1, 1e32, 1e48 }) |x| try s.revert(x);
        if (final_zero) {
            try testing.expectEqual(@as(usize, 1), s.n());
            try s.revert(0.0);
        }
        try testing.expectEqual(@as(usize, 0), s.n());
        try testing.expectEqual(@as(f64, 0.0), s.value());
        try testing.expect(math.isNan(s.mean()));
        s.update(3.0);
        try testing.expectEqual(@as(usize, 1), s.n());
        try testing.expectEqual(@as(f64, 3.0), s.value());
        try testing.expectEqual(@as(f64, 3.0), s.mean());
    }
}

test "reset" {
    var s = KleinKBNSummator{};
    for ([_]f64{ 1.0, 2.0 }) |x| s.update(x);
    s.reset();
    try testing.expectEqual(@as(usize, 0), s.n());
    try testing.expectEqual(@as(f64, 0.0), s.value());
    try testing.expect(math.isNan(s.mean()));
    s.update(3.0);
    try testing.expectEqual(@as(f64, 3.0), s.mean());
}
