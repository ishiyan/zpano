//! Single factor model (SFM) regression of excess returns.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const LinearRegressionKleinKBN = @import("linear_regression_klein_kbn").LinearRegressionKleinKBN;

/// Errors returned by `SFMRegression.revert`.
pub const Error = error{EmptyRevert};

/// Streaming single factor model regression of the portfolio excess
/// returns y = ret - risk_free_rate on the benchmark excess returns
/// x = benchmark - risk_free_rate, with separate bull (x > 0) and
/// bear (x < 0) fits.
///
/// Construct with `SFMRegression.init(risk_free_rate)`; no allocation.
pub const SFMRegression = struct {
    risk_free_rate: f64,
    full: LinearRegressionKleinKBN = .{},
    bull: LinearRegressionKleinKBN = .{},
    bear: LinearRegressionKleinKBN = .{},

    const Self = @This();

    pub fn init(risk_free_rate: f64) Self {
        return .{ .risk_free_rate = risk_free_rate };
    }

    pub fn reset(self: *Self) void {
        self.full.reset();
        self.bull.reset();
        self.bear.reset();
    }

    /// Removes a previously added (portfolio, benchmark) pair.
    ///
    /// Returns `error.EmptyRevert` if a regression is empty.
    pub fn revert(self: *Self, ret: f64, benchmark: f64) Error!void {
        const x = benchmark - self.risk_free_rate;
        const y = ret - self.risk_free_rate;

        try self.full.revert(x, y);

        if (x > 0) {
            try self.bull.revert(x, y);
        } else if (x < 0) {
            try self.bear.revert(x, y);
        }
    }

    /// Adds a (portfolio, benchmark) return pair.
    pub fn update(self: *Self, ret: f64, benchmark: f64) void {
        const x = benchmark - self.risk_free_rate;
        const y = ret - self.risk_free_rate;

        self.full.update(x, y);

        if (x > 0) {
            self.bull.update(x, y);
        } else if (x < 0) {
            self.bear.update(x, y);
        }
    }

    /// Intercept of the full fit (NaN when the slope is NaN).
    pub fn alpha(self: *const Self) f64 {
        return self.full.intercept();
    }

    /// Slope of the full fit (NaN when n < 2 or x is constant).
    pub fn beta(self: *const Self) f64 {
        return self.full.slope();
    }

    /// Slope of the bull (x > 0) fit.
    pub fn betaBull(self: *const Self) f64 {
        return self.bull.slope();
    }

    /// Slope of the bear (x < 0) fit.
    pub fn betaBear(self: *const Self) f64 {
        return self.bear.slope();
    }

    /// Coefficient of determination r² of the full fit (NaN when the
    /// correlation is NaN).
    pub fn r2(self: *const Self) f64 {
        const corr = self.full.correlation();
        return if (!math.isNan(corr)) corr * corr else math.nan(f64);
    }
};

// ── Tests ──────────────────────────────────────────────────────────────────

fn expectAlmostEqual(expected: f64, actual: f64, places: u5) !void {
    if (expected == actual) return;
    const tol = 0.5 * math.pow(f64, 10.0, -@as(f64, @floatFromInt(places)));
    if (!(@abs(expected - actual) <= tol)) {
        std.debug.print("expected {d}, got {d} (places {d})\n", .{ expected, actual, places });
        return error.TestExpectedApproxEq;
    }
}

test "full bull and bear fits use excess returns" {
    var regression = SFMRegression.init(0.01);
    for ([_]f64{ -0.04, -0.02, 0.0, 0.02, 0.05 }) |excess_benchmark| {
        const benchmark = 0.01 + excess_benchmark;
        const portfolio = 0.01 + 0.005 + 2 * excess_benchmark;
        regression.update(portfolio, benchmark);
    }

    try expectAlmostEqual(0.005, regression.alpha(), 14);
    try expectAlmostEqual(2.0, regression.beta(), 14);
    try expectAlmostEqual(2.0, regression.betaBull(), 14);
    try expectAlmostEqual(2.0, regression.betaBear(), 14);
    try expectAlmostEqual(1.0, regression.r2(), 14);

    // Removing an older bear observation leaves too few bear points
    // for a slope, while the full and bull fits remain defined.
    try regression.revert(0.01 + 0.005 + 2 * -0.04, 0.01 - 0.04);
    try expectAlmostEqual(2.0, regression.beta(), 14);
    try expectAlmostEqual(2.0, regression.betaBull(), 14);
    try testing.expect(math.isNan(regression.betaBear()));
}

test "reset and zero excess benchmark" {
    var regression = SFMRegression.init(0.01);
    regression.update(0.02, 0.01);
    try testing.expect(math.isNan(regression.betaBull()));
    try testing.expect(math.isNan(regression.betaBear()));
    regression.reset();
    try testing.expect(math.isNan(regression.alpha()));
    try testing.expect(math.isNan(regression.beta()));
    try testing.expect(math.isNan(regression.betaBull()));
    try testing.expect(math.isNan(regression.betaBear()));
    try testing.expect(math.isNan(regression.r2()));
}
