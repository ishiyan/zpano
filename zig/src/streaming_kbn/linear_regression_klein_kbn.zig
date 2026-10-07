const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;
const RawMomentsKleinKBN = @import("raw_moments_klein_kbn").RawMomentsKleinKBN;

/// Errors returned by `LinearRegressionKleinKBN`.
pub const Error = error{EmptyRevert};

/// Streaming ordinary least squares (OLS) regression y = a + b·x with
/// Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.
///
/// Tracks the means and variances of x and y (as RawMomentsKleinKBN with
/// ddof=0) and the co-moment S_xy = Σ(x − x̄)(y − ȳ), updated in O(1) per
/// sample (n = count before adding the sample):
///
///     S_xy += (x − x̄)·(y − ȳ)·n / (n + 1)
///
/// where x̄ and ȳ are the means before adding the sample.
///
/// Like RawMomentsKleinKBN, revert() may remove any previously added
/// (x, y) pair, not only the most recent one, so the type works for FIFO
/// rolling windows.
///
/// Derived quantities, with S_xx = Σ(x − x̄)² and S_yy = Σ(y − ȳ)²:
///
///     slope       b = S_xy / S_xx
///     intercept   a = ȳ − b·x̄
///     correlation r = S_xy / √(S_xx·S_yy)
///     covariance      S_xy / n   (population)
pub const LinearRegressionKleinKBN = struct {
    count: usize = 0,
    x_moments: RawMomentsKleinKBN = .{ .ddof = 0 },
    y_moments: RawMomentsKleinKBN = .{ .ddof = 0 },
    s_xy: KleinKBNAccumulator = .{},

    const Self = @This();

    /// Clears all accumulated state.
    pub fn reset(self: *Self) void {
        self.count = 0;
        self.x_moments.reset();
        self.y_moments.reset();
        self.s_xy.reset();
    }

    /// Adds a sample (x, y).
    pub fn update(self: *Self, x: f64, y: f64) void {
        const n_old: f64 = @floatFromInt(self.count);
        self.count += 1;
        const term = (self.x_moments.mean() - x) * (self.y_moments.mean() - y) * n_old / (n_old + 1);
        self.s_xy.update(term);
        self.x_moments.update(x);
        self.y_moments.update(y);
    }

    /// Removes a previously added sample (x, y), not necessarily the most
    /// recent one.
    ///
    /// Returns `error.EmptyRevert` if there are no samples.
    pub fn revert(self: *Self, x: f64, y: f64) Error!void {
        if (self.count == 0) return error.EmptyRevert;
        if (self.count == 1) {
            self.reset();
            return;
        }
        try self.x_moments.revert(x);
        try self.y_moments.revert(y);
        // The means are now those without (x, y), as in update().
        const count = self.count - 1;
        const nf: f64 = @floatFromInt(count);
        const term = (self.x_moments.mean() - x) * (self.y_moments.mean() - y) * nf / (nf + 1);
        self.s_xy.revert(term);
        self.count = count;
    }

    /// The number of samples.
    pub fn n(self: *const Self) usize {
        return self.count;
    }

    /// The mean of x (0.0 when empty).
    pub fn meanX(self: *const Self) f64 {
        return self.x_moments.mean();
    }

    /// The mean of y (0.0 when empty).
    pub fn meanY(self: *const Self) f64 {
        return self.y_moments.mean();
    }

    /// The population variance of x, S_xx / n (NaN when empty).
    pub fn varianceX(self: *const Self) f64 {
        return self.x_moments.variance();
    }

    /// The population variance of y, S_yy / n (NaN when empty).
    pub fn varianceY(self: *const Self) f64 {
        return self.y_moments.variance();
    }

    /// The co-moment S_xy = Σ(x − x̄)(y − ȳ) (0.0 when empty).
    pub fn coMoment(self: *const Self) f64 {
        return self.s_xy.value();
    }

    /// The population covariance S_xy / n (NaN when empty).
    pub fn covariance(self: *const Self) f64 {
        if (self.count < 1) return math.nan(f64);
        return self.s_xy.value() / @as(f64, @floatFromInt(self.count));
    }

    /// The OLS slope b = S_xy / S_xx.
    ///
    /// NaN when n < 2 or all x are equal (S_xx = 0).
    pub fn slope(self: *const Self) f64 {
        if (self.count < 2) return math.nan(f64);
        const s_xx = self.x_moments.variance() * @as(f64, @floatFromInt(self.count));
        return if (s_xx != 0) self.s_xy.value() / s_xx else math.nan(f64);
    }

    /// The OLS intercept a = ȳ − b·x̄ (NaN when the slope is NaN).
    pub fn intercept(self: *const Self) f64 {
        return self.y_moments.mean() - self.slope() * self.x_moments.mean();
    }

    /// The Pearson correlation coefficient r = S_xy / √(S_xx·S_yy),
    /// clamped to [−1, 1] to absorb rounding.
    ///
    /// NaN when n < 2 or either x or y is constant.
    pub fn correlation(self: *const Self) f64 {
        if (self.count < 2) return math.nan(f64);
        const t = self.x_moments.standardDeviation() * self.y_moments.standardDeviation();
        if (t == 0) return math.nan(f64);
        const r = self.s_xy.value() / (t * @as(f64, @floatFromInt(self.count)));
        return @max(-1.0, @min(1.0, r));
    }
};

// ── Tests ──────────────────────────────────────────────────────────────────

const Getter = *const fn (*const LinearRegressionKleinKBN) f64;

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

fn feed(reg: *LinearRegressionKleinKBN, xs: []const f64, ys: []const f64) void {
    for (xs, ys) |x, y| reg.update(x, y);
}

fn expectAllNan(reg: *const LinearRegressionKleinKBN) !void {
    try testing.expect(math.isNan(reg.slope()));
    try testing.expect(math.isNan(reg.intercept()));
    try testing.expect(math.isNan(reg.correlation()));
}

// Bacon, Carl R., Practical Portfolio Performance Measurement and
// Attribution, 2nd ed., Wiley, 2008, p. 65 (portfolio) and p. 66 (benchmark).
const portfolio = [_]f64{
    0.003,  0.026,  0.011,  -0.010,
    0.015,  0.025,  0.016,  0.067,
    -0.014, 0.040,  -0.005, 0.081,
    0.040,  -0.037, -0.061, 0.017,
    -0.049, -0.022, 0.070,  0.058,
    -0.065, 0.024,  -0.005, -0.009,
};
const benchmark = [_]f64{
    0.002,  0.025,  0.018,  -0.011,
    0.014,  0.018,  0.014,  0.065,
    -0.015, 0.042,  -0.006, 0.083,
    0.039,  -0.038, -0.062, 0.015,
    -0.048, 0.021,  0.060,  0.056,
    -0.067, 0.019,  -0.003, 0.000,
};

// Reference values for x = benchmark, y = portfolio, computed with exact
// rational arithmetic (fractions.Fraction on the binary float inputs,
// square roots with 50-digit decimal.Decimal), rounded to the nearest float.
const slope_ref: f64 = 0.9988502086225746;
const intercept_ref: f64 = -0.001030120844918352;
const correlation_ref: f64 = 0.9693858148753051;
const co_moment_ref: f64 = 0.033844;
const covariance_ref: f64 = 0.0014101666666666668;
const variance_x_ref: f64 = 0.0014117899305555557;
const variance_y_ref: f64 = 0.0014989166666666668;

test "bacon" {
    var reg = LinearRegressionKleinKBN{};
    feed(&reg, &benchmark, &portfolio);
    try testing.expectEqual(@as(usize, portfolio.len), reg.n());
    try testing.expect(almostEqual(reg.slope(), slope_ref, 1e-14));
    try testing.expect(almostEqual(reg.intercept(), intercept_ref, 1e-15));
    try testing.expect(almostEqual(reg.correlation(), correlation_ref, 1e-14));
    try testing.expect(almostEqual(reg.coMoment(), co_moment_ref, 1e-16));
    try testing.expect(almostEqual(reg.covariance(), covariance_ref, 1e-16));
    try testing.expect(almostEqual(reg.varianceX(), variance_x_ref, 1e-16));
    try testing.expect(almostEqual(reg.varianceY(), variance_y_ref, 1e-16));
}

test "perfect fit" {
    var reg = LinearRegressionKleinKBN{};
    var i: usize = 0;
    while (i < 5) : (i += 1) {
        const x: f64 = @floatFromInt(i);
        reg.update(x, 2 * x + 1);
    }
    try testing.expect(almostEqual(reg.slope(), 2.0, 1e-13));
    try testing.expect(almostEqual(reg.intercept(), 1.0, 1e-13));
    try testing.expect(almostEqual(reg.correlation(), 1.0, 1e-15));
    try testing.expect(almostEqual(reg.meanX(), 2.0, 1e-15));
    try testing.expect(almostEqual(reg.meanY(), 5.0, 1e-15));
    try testing.expect(almostEqual(reg.varianceX(), 2.0, 1e-15));
    try testing.expect(almostEqual(reg.coMoment(), 20.0, 1e-13));
    try testing.expect(almostEqual(reg.covariance(), 4.0, 1e-13));
}

test "negative correlation" {
    var reg = LinearRegressionKleinKBN{};
    var i: usize = 0;
    while (i < 5) : (i += 1) {
        const x: f64 = @floatFromInt(i);
        reg.update(x, -2.0 * x + 1.0);
    }
    try testing.expect(almostEqual(reg.slope(), -2.0, 1e-13));
    try testing.expect(almostEqual(reg.intercept(), 1.0, 1e-13));
    try testing.expect(almostEqual(reg.correlation(), -1.0, 1e-15));
    try testing.expect(almostEqual(reg.covariance(), -4.0, 1e-13));
}

test "constant y" {
    var reg = LinearRegressionKleinKBN{};
    var i: usize = 0;
    while (i < 5) : (i += 1) {
        reg.update(@floatFromInt(i), 3.0);
    }
    try testing.expect(almostEqual(reg.slope(), 0.0, 1e-13));
    try testing.expect(almostEqual(reg.intercept(), 3.0, 1e-13));
    try testing.expectEqual(@as(f64, 0.0), reg.covariance());
    try testing.expect(math.isNan(reg.correlation()));
}

test "constant x" {
    var reg = LinearRegressionKleinKBN{};
    var i: usize = 0;
    while (i < 5) : (i += 1) {
        reg.update(3.0, @floatFromInt(i));
    }
    try testing.expectEqual(@as(f64, 0.0), reg.covariance());
    try expectAllNan(&reg);
}

test "empty" {
    const reg = LinearRegressionKleinKBN{};
    try testing.expectEqual(@as(usize, 0), reg.n());
    try testing.expectEqual(@as(f64, 0.0), reg.coMoment());
    try testing.expect(math.isNan(reg.covariance()));
    try testing.expect(math.isNan(reg.varianceX()));
    try expectAllNan(&reg);
}

test "single point" {
    var reg = LinearRegressionKleinKBN{};
    reg.update(1.0, 2.0);
    try testing.expectEqual(@as(f64, 0.0), reg.covariance());
    try expectAllNan(&reg);
}

test "two points" {
    var reg = LinearRegressionKleinKBN{};
    reg.update(0.0, 1.0);
    reg.update(2.0, 5.0);
    try testing.expect(almostEqual(reg.slope(), 2.0, 1e-13));
    try testing.expect(almostEqual(reg.intercept(), 1.0, 1e-13));
    try testing.expect(almostEqual(reg.correlation(), 1.0, 1e-13));
}

test "revert most recent" {
    var reg = LinearRegressionKleinKBN{};
    feed(&reg, &benchmark, &portfolio);
    reg.update(0.5, -0.5);
    try reg.revert(0.5, -0.5);
    try testing.expectEqual(@as(usize, portfolio.len), reg.n());
    try testing.expect(almostEqual(reg.slope(), slope_ref, 1e-13));
    try testing.expect(almostEqual(reg.intercept(), intercept_ref, 1e-14));
    try testing.expect(almostEqual(reg.correlation(), correlation_ref, 1e-13));
    try testing.expect(almostEqual(reg.covariance(), covariance_ref, 1e-15));
}

test "revert oldest" {
    var reg = LinearRegressionKleinKBN{};
    reg.update(0.5, -0.5);
    feed(&reg, &benchmark, &portfolio);
    try reg.revert(0.5, -0.5);
    try testing.expect(almostEqual(reg.slope(), slope_ref, 1e-13));
    try testing.expect(almostEqual(reg.intercept(), intercept_ref, 1e-14));
    try testing.expect(almostEqual(reg.correlation(), correlation_ref, 1e-13));
    try testing.expect(almostEqual(reg.covariance(), covariance_ref, 1e-15));
}

test "revert to single" {
    var reg = LinearRegressionKleinKBN{};
    reg.update(1.0, 2.0);
    reg.update(3.0, 4.0);
    try reg.revert(3.0, 4.0);
    try testing.expectEqual(@as(usize, 1), reg.n());
    try testing.expect(almostEqual(reg.meanX(), 1.0, 1e-15));
    try testing.expect(almostEqual(reg.meanY(), 2.0, 1e-15));
    try testing.expect(almostEqual(reg.coMoment(), 0.0, 1e-15));
    try expectAllNan(&reg);
}

test "revert to empty" {
    var reg = LinearRegressionKleinKBN{};
    reg.update(1.0, 2.0);
    try reg.revert(1.0, 2.0);
    try testing.expectEqual(@as(usize, 0), reg.n());
    try expectAllNan(&reg);
}

test "revert empty raises" {
    var reg = LinearRegressionKleinKBN{};
    try testing.expectError(error.EmptyRevert, reg.revert(1.0, 2.0));
}

test "rolling window" {
    const getters = [_]Getter{
        &LinearRegressionKleinKBN.slope,      &LinearRegressionKleinKBN.intercept,
        &LinearRegressionKleinKBN.correlation, &LinearRegressionKleinKBN.covariance,
        &LinearRegressionKleinKBN.coMoment,   &LinearRegressionKleinKBN.varianceX,
        &LinearRegressionKleinKBN.varianceY,
    };
    const w: usize = 6;
    var reg = LinearRegressionKleinKBN{};
    for (benchmark, portfolio, 0..) |x, y, i| {
        reg.update(x, y);
        if (i >= w) try reg.revert(benchmark[i - w], portfolio[i - w]);
        const lo = if (i + 1 >= w) i + 1 - w else 0;
        var ref = LinearRegressionKleinKBN{};
        feed(&ref, benchmark[lo .. i + 1], portfolio[lo .. i + 1]);
        try testing.expectEqual(ref.n(), reg.n());
        for (getters) |g| {
            const actual = g(&reg);
            const expected = g(&ref);
            if (math.isNan(expected)) {
                try testing.expect(math.isNan(actual));
            } else {
                try testing.expect(almostEqual(actual, expected, 1e-13));
            }
        }
    }
}

test "reset" {
    var reg = LinearRegressionKleinKBN{};
    var i: usize = 0;
    while (i < 5) : (i += 1) {
        const x: f64 = @floatFromInt(i);
        reg.update(x, 2 * x + 1);
    }
    reg.reset();
    try testing.expectEqual(@as(usize, 0), reg.n());
    try expectAllNan(&reg);
    reg.update(0.0, 1.0);
    reg.update(1.0, 3.0);
    try testing.expect(almostEqual(reg.slope(), 2.0, 1e-13));
}
