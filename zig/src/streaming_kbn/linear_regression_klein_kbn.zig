const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;
const RawMomentsKleinKBN = @import("raw_moments_klein_kbn").RawMomentsKleinKBN;

/// Streaming simple linear regression (y = slope * x + intercept) with
/// KBN-compensated accumulation.
///
/// Internally uses two RawMomentsKleinKBN (ddof=0) for x and y moments,
/// and a KleinKBNAccumulator for the cross-product S_xy.
///
/// Supports both LIFO revert and FIFO rolling window via the revert/update cycle.
pub const LinearRegressionKleinKBN = struct {
    n: u32 = 0,
    x_moments: RawMomentsKleinKBN = .{ .ddof = 0, .bias = true, .fisher = true },
    y_moments: RawMomentsKleinKBN = .{ .ddof = 0, .bias = true, .fisher = true },
    s_xy: KleinKBNAccumulator = .{},

    /// Clears all accumulated state.
    pub fn reset(self: *LinearRegressionKleinKBN) void {
        self.n = 0;
        self.x_moments.reset();
        self.y_moments.reset();
        self.s_xy.reset();
    }

    /// Adds a new (x, y) observation.
    pub fn update(self: *LinearRegressionKleinKBN, x: f64, y: f64) void {
        const n_old = self.n;
        self.n += 1;
        const term = (self.x_moments.mean() - x) * (self.y_moments.mean() - y) *
            @as(f64, @floatFromInt(n_old)) / @as(f64, @floatFromInt(n_old + 1));
        self.s_xy.update(term);
        self.x_moments.update(x);
        self.y_moments.update(y);
    }

    /// Removes a previously added (x, y) observation.
    pub fn revert(self: *LinearRegressionKleinKBN, x: f64, y: f64) void {
        if (self.n == 0) return;
        if (self.n == 1) {
            self.reset();
            return;
        }
        self.x_moments.revert(x);
        self.y_moments.revert(y);
        const n = self.n - 1;
        const term = (self.x_moments.mean() - x) * (self.y_moments.mean() - y) *
            @as(f64, @floatFromInt(n)) / @as(f64, @floatFromInt(n + 1));
        self.s_xy.update(-term);
        self.n = n;
    }

    /// Returns the current slope coefficient. Returns NaN if n < 2 or S_xx == 0.
    pub fn slope(self: *const LinearRegressionKleinKBN) f64 {
        if (self.n < 2) return math.nan(f64);
        const Sxx = self.x_moments.variance() * @as(f64, @floatFromInt(self.n));
        if (Sxx == 0.0) return math.nan(f64);
        return self.s_xy.value() / Sxx;
    }

    /// Returns the current intercept coefficient.
    pub fn intercept(self: *const LinearRegressionKleinKBN) f64 {
        return self.y_moments.mean() - self.slope() * self.x_moments.mean();
    }

    /// Returns the current Pearson correlation coefficient. Returns NaN if n < 2
    /// or either standard deviation is zero.
    pub fn correlation(self: *const LinearRegressionKleinKBN) f64 {
        if (self.n < 2) return math.nan(f64);
        const t = self.x_moments.standardDeviation() * self.y_moments.standardDeviation();
        if (t == 0.0) return math.nan(f64);
        return self.s_xy.value() / (t * @as(f64, @floatFromInt(self.n)));
    }
};

// ── Tests ──────────────────────────────────────────────────────────────────

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

test "perfect fit" {
    var r = LinearRegressionKleinKBN{};
    var i: u32 = 0;
    while (i < 5) : (i += 1) {
        const x = @as(f64, @floatFromInt(i));
        r.update(x, 2.0 * x + 1.0);
    }
    try testing.expect(almostEqual(r.slope(), 2.0, 1e-13));
    try testing.expect(almostEqual(r.intercept(), 1.0, 1e-13));
    try testing.expect(almostEqual(r.correlation(), 1.0, 1e-13));
}

test "zero correlation" {
    var r = LinearRegressionKleinKBN{};
    var i: u32 = 0;
    while (i < 5) : (i += 1) {
        r.update(@as(f64, @floatFromInt(i)), 0.0);
    }
    try testing.expect(almostEqual(r.slope(), 0.0, 1e-13));
    try testing.expect(math.isNan(r.correlation()));
}

test "single point" {
    var r = LinearRegressionKleinKBN{};
    r.update(1.0, 2.0);
    try testing.expect(math.isNan(r.slope()));
    try testing.expect(math.isNan(r.intercept()));
    try testing.expect(math.isNan(r.correlation()));
}

test "two points" {
    var r = LinearRegressionKleinKBN{};
    r.update(0.0, 1.0);
    r.update(2.0, 5.0);
    try testing.expect(almostEqual(r.slope(), 2.0, 1e-13));
    try testing.expect(almostEqual(r.intercept(), 1.0, 1e-13));
    try testing.expect(almostEqual(r.correlation(), 1.0, 1e-13));
}

test "revert matches single update" {
    var r = LinearRegressionKleinKBN{};
    r.update(1.0, 2.0);
    r.update(3.0, 4.0);
    r.revert(3.0, 4.0);

    var ref = LinearRegressionKleinKBN{};
    ref.update(1.0, 2.0);

    try testing.expect(r.n == ref.n);
    try testing.expect(math.isNan(r.slope()));
    try testing.expect(math.isNan(ref.slope()));
}

test "revert to empty" {
    var r = LinearRegressionKleinKBN{};
    r.update(1.0, 2.0);
    r.revert(1.0, 2.0);
    try testing.expect(r.n == 0);
    try testing.expect(math.isNan(r.slope()));
    try testing.expect(math.isNan(r.intercept()));
    try testing.expect(math.isNan(r.correlation()));
}

test "rolling window" {
    const data = [_][2]f64{
        .{ 0.0, 1.0 }, .{ 1.0, 3.0 }, .{ 2.0, 5.0 }, .{ 3.0, 7.0 }, .{ 4.0, 9.0 },
    };
    var r = LinearRegressionKleinKBN{};
    for (data) |p| {
        r.update(p[0], p[1]);
    }
    r.revert(data[0][0], data[0][1]);
    r.revert(data[1][0], data[1][1]);
    r.update(5.0, 11.0);
    r.update(6.0, 13.0);

    var ref = LinearRegressionKleinKBN{};
    for (data[2..]) |p| {
        ref.update(p[0], p[1]);
    }
    ref.update(5.0, 11.0);
    ref.update(6.0, 13.0);

    try testing.expect(r.n == ref.n);
    try testing.expect(almostEqual(r.slope(), ref.slope(), 1e-12));
    try testing.expect(almostEqual(r.intercept(), ref.intercept(), 1e-12));
    try testing.expect(almostEqual(r.correlation(), ref.correlation(), 1e-12));
}

test "negative correlation" {
    var r = LinearRegressionKleinKBN{};
    var i: u32 = 0;
    while (i < 5) : (i += 1) {
        const x = @as(f64, @floatFromInt(i));
        r.update(x, -2.0 * x + 1.0);
    }
    try testing.expect(almostEqual(r.slope(), -2.0, 1e-13));
    try testing.expect(almostEqual(r.intercept(), 1.0, 1e-13));
    try testing.expect(almostEqual(r.correlation(), -1.0, 1e-13));
}

test "reset" {
    var r = LinearRegressionKleinKBN{};
    var i: u32 = 0;
    while (i < 5) : (i += 1) {
        const x = @as(f64, @floatFromInt(i));
        r.update(x, 2.0 * x + 1.0);
    }
    r.reset();
    try testing.expect(r.n == 0);
    try testing.expect(math.isNan(r.slope()));
    try testing.expect(math.isNan(r.intercept()));
    try testing.expect(math.isNan(r.correlation()));

    r.update(0.0, 1.0);
    r.update(1.0, 3.0);
    try testing.expect(almostEqual(r.slope(), 2.0, 1e-13));
}
