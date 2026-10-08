//! Streaming upside/downside capture, number and percentage ratios.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;

/// Errors returned by `Capture.revert`.
pub const Error = error{EmptyRevert};

/// Streaming capture ratios of an asset against a benchmark, following
/// the PerformanceAnalytics conventions: upside periods have
/// benchmark > 0; downside capture and down-number use benchmark <= 0;
/// down-percentage uses benchmark < 0.
///
/// Construct with `Capture{}`; no allocation. All ratios are NaN when
/// their denominator is zero.
pub const Capture = struct {
    logret_a_sum_up: KleinKBNAccumulator = .{},
    logret_b_sum_up: KleinKBNAccumulator = .{},
    logret_a_sum_dn: KleinKBNAccumulator = .{},
    logret_b_sum_dn: KleinKBNAccumulator = .{},
    a_sum_up: KleinKBNAccumulator = .{},
    b_sum_up: KleinKBNAccumulator = .{},
    a_sum_dn: KleinKBNAccumulator = .{},
    b_sum_dn: KleinKBNAccumulator = .{},
    a_num_up: usize = 0,
    b_num_up: usize = 0,
    a_num_dn: usize = 0,
    b_num_dn: usize = 0,
    a_perc_up: usize = 0,
    b_perc_up: usize = 0,
    a_perc_dn: usize = 0,
    b_perc_dn: usize = 0,

    const Self = @This();

    pub fn reset(self: *Self) void {
        self.* = .{};
    }

    fn logret(ret: f64) f64 {
        return if (ret != 0) math.log1p(ret) else 0;
    }

    /// Removes a previously added (asset, benchmark) pair.
    ///
    /// Returns `error.EmptyRevert` if nothing was added. (Python has no
    /// check; its counters would go negative.) Reverting a pair that was
    /// never added is illegal.
    pub fn revert(self: *Self, ret_asset: f64, ret_benchmark: f64) Error!void {
        if (self.b_num_up + self.b_num_dn == 0) return error.EmptyRevert;
        if (ret_benchmark > 0) { // Upside
            // Geometric
            self.logret_a_sum_up.revert(logret(ret_asset));
            self.logret_b_sum_up.revert(logret(ret_benchmark));
            // Arithmetic
            self.a_sum_up.revert(ret_asset);
            self.b_sum_up.revert(ret_benchmark);
            // Number
            self.b_num_up -= 1;
            if (ret_asset > 0) self.a_num_up -= 1;
            // Percentage
            self.b_perc_up -= 1;
            if (ret_asset > ret_benchmark) self.a_perc_up -= 1;
        } else { // Downside
            // Geometric
            self.logret_a_sum_dn.revert(logret(ret_asset));
            self.logret_b_sum_dn.revert(logret(ret_benchmark));
            // Arithmetic
            self.a_sum_dn.revert(ret_asset);
            self.b_sum_dn.revert(ret_benchmark);
            // Number
            self.b_num_dn -= 1;
            if (ret_asset < 0) self.a_num_dn -= 1;
            // Percentage
            if (ret_benchmark < 0) {
                self.b_perc_dn -= 1;
                if (ret_asset > ret_benchmark) self.a_perc_dn -= 1;
            }
        }
    }

    /// Adds an (asset, benchmark) return pair.
    pub fn update(self: *Self, ret_asset: f64, ret_benchmark: f64) void {
        if (ret_benchmark > 0) { // Upside
            // Geometric
            self.logret_a_sum_up.update(logret(ret_asset));
            self.logret_b_sum_up.update(logret(ret_benchmark));
            // Arithmetic
            self.a_sum_up.update(ret_asset);
            self.b_sum_up.update(ret_benchmark);
            // Counts
            self.b_num_up += 1;
            if (ret_asset > 0) self.a_num_up += 1;
            // Perc
            self.b_perc_up += 1;
            if (ret_asset > ret_benchmark) self.a_perc_up += 1;
        } else { // Downside
            // Geometric
            self.logret_a_sum_dn.update(logret(ret_asset));
            self.logret_b_sum_dn.update(logret(ret_benchmark));
            // Arithmetic
            self.a_sum_dn.update(ret_asset);
            self.b_sum_dn.update(ret_benchmark);
            // Counts
            self.b_num_dn += 1;
            if (ret_asset < 0) self.a_num_dn += 1;
            // Perc
            if (ret_benchmark < 0) {
                self.b_perc_dn += 1;
                if (ret_asset > ret_benchmark) self.a_perc_dn += 1;
            }
        }
    }

    fn ratio(a: usize, b: usize) f64 {
        if (b == 0) return math.nan(f64);
        return @as(f64, @floatFromInt(a)) / @as(f64, @floatFromInt(b));
    }

    /// Geometric upside capture: (prod(1 + a) - 1) / (prod(1 + b) - 1)
    /// over periods with benchmark > 0.
    pub fn upsideCaptureRatioGeometric(self: *const Self) f64 {
        const a_cum = math.expm1(self.logret_a_sum_up.value());
        const b_cum = math.expm1(self.logret_b_sum_up.value());
        return if (b_cum != 0) a_cum / b_cum else math.nan(f64);
    }

    /// Arithmetic upside capture: sum(a) / sum(b) over periods with benchmark > 0.
    pub fn upsideCaptureRatioArithmetic(self: *const Self) f64 {
        const b_sum = self.b_sum_up.value();
        return if (b_sum != 0) self.a_sum_up.value() / b_sum else math.nan(f64);
    }

    /// Geometric downside capture over periods with benchmark <= 0.
    pub fn downsideCaptureRatioGeometric(self: *const Self) f64 {
        const a_cum = math.expm1(self.logret_a_sum_dn.value());
        const b_cum = math.expm1(self.logret_b_sum_dn.value());
        return if (b_cum != 0) a_cum / b_cum else math.nan(f64);
    }

    /// Arithmetic downside capture over periods with benchmark <= 0.
    pub fn downsideCaptureRatioArithmetic(self: *const Self) f64 {
        const b_sum = self.b_sum_dn.value();
        return if (b_sum != 0) self.a_sum_dn.value() / b_sum else math.nan(f64);
    }

    /// Fraction of up-benchmark periods with a positive asset return.
    pub fn upNumberRatio(self: *const Self) f64 {
        return ratio(self.a_num_up, self.b_num_up);
    }

    /// Fraction of down-benchmark (<= 0) periods with a negative asset return.
    pub fn downNumberRatio(self: *const Self) f64 {
        return ratio(self.a_num_dn, self.b_num_dn);
    }

    /// Fraction of up-benchmark periods where the asset beat the benchmark.
    pub fn upPercentageRatio(self: *const Self) f64 {
        return ratio(self.a_perc_up, self.b_perc_up);
    }

    /// Fraction of strictly negative benchmark periods where the asset beat
    /// the benchmark.
    pub fn downPercentageRatio(self: *const Self) f64 {
        return ratio(self.a_perc_dn, self.b_perc_dn);
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

fn gauss(r: std.Random, mu: f64, sigma: f64) f64 {
    return mu + sigma * r.floatNorm(f64);
}

const Pair = struct { a: f64, b: f64 };

fn nanDiv(a: f64, b: f64) f64 {
    return if (b != 0) a / b else math.nan(f64);
}

const getters = [_]*const fn (*const Capture) f64{
    Capture.upsideCaptureRatioGeometric,
    Capture.upsideCaptureRatioArithmetic,
    Capture.downsideCaptureRatioGeometric,
    Capture.downsideCaptureRatioArithmetic,
    Capture.upNumberRatio,
    Capture.downNumberRatio,
    Capture.upPercentageRatio,
    Capture.downPercentageRatio,
};

/// Naive capture ratios, in the order of `getters`.
fn reference(pairs: []const Pair) [8]f64 {
    var up_pa: f64 = 1;
    var up_pb: f64 = 1;
    var up_sa: f64 = 0;
    var up_sb: f64 = 0;
    var dn_pa: f64 = 1;
    var dn_pb: f64 = 1;
    var dn_sa: f64 = 0;
    var dn_sb: f64 = 0;
    var up_n: f64 = 0;
    var up_num: f64 = 0;
    var up_perc: f64 = 0;
    var dn_n: f64 = 0;
    var dn_num: f64 = 0;
    var dns_n: f64 = 0;
    var dns_perc: f64 = 0;
    for (pairs) |p| {
        if (p.b > 0) {
            up_pa *= 1 + p.a;
            up_pb *= 1 + p.b;
            up_sa += p.a;
            up_sb += p.b;
            up_n += 1;
            if (p.a > 0) up_num += 1;
            if (p.a > p.b) up_perc += 1;
        } else {
            dn_pa *= 1 + p.a;
            dn_pb *= 1 + p.b;
            dn_sa += p.a;
            dn_sb += p.b;
            dn_n += 1;
            if (p.a < 0) dn_num += 1;
        }
        if (p.b < 0) {
            dns_n += 1;
            if (p.a > p.b) dns_perc += 1;
        }
    }
    return .{
        nanDiv(up_pa - 1, up_pb - 1),
        nanDiv(up_sa, up_sb),
        nanDiv(dn_pa - 1, dn_pb - 1),
        nanDiv(dn_sa, dn_sb),
        nanDiv(up_num, up_n),
        nanDiv(dn_num, dn_n),
        nanDiv(up_perc, up_n),
        nanDiv(dns_perc, dns_n),
    };
}

fn assertMatches(c: *const Capture, pairs: []const Pair, places: u5) !void {
    const expected = reference(pairs);
    for (getters, expected) |g, e| {
        const actual = g(c);
        if (math.isNan(e)) {
            try testing.expect(math.isNan(actual));
        } else {
            try expectAlmostEqual(e, actual, places);
        }
    }
}

fn randomPairs(seed: u64, comptime n: usize) [n]Pair {
    var prng = std.Random.DefaultPrng.init(seed);
    const rng = prng.random();
    var pairs: [n]Pair = undefined;
    for (&pairs) |*p| {
        const ga = gauss(rng, 0.0, 0.03);
        p.a = if (rng.boolean()) 0.0 else ga;
        const gb = gauss(rng, 0.0, 0.03);
        p.b = if (rng.boolean()) 0.0 else gb;
    }
    return pairs;
}

test "empty" {
    const c = Capture{};
    for (getters) |g| try testing.expect(math.isNan(g(&c)));
}

test "hand computed" {
    const pairs = [_]Pair{ .{ .a = 0.02, .b = 0.01 }, .{ .a = -0.01, .b = -0.02 }, .{ .a = 0.03, .b = 0.04 }, .{ .a = -0.03, .b = 0.0 }, .{ .a = 0.01, .b = -0.01 } };
    var c = Capture{};
    for (pairs) |p| c.update(p.a, p.b);
    // Up periods (b > 0): (0.02, 0.01), (0.03, 0.04)
    try expectAlmostEqual(0.05 / 0.05, c.upsideCaptureRatioArithmetic(), 15);
    try expectAlmostEqual((1.02 * 1.03 - 1) / (1.01 * 1.04 - 1), c.upsideCaptureRatioGeometric(), 14);
    try testing.expectEqual(@as(f64, 1.0), c.upNumberRatio());
    try testing.expectEqual(@as(f64, 0.5), c.upPercentageRatio());
    // Down periods (b <= 0) include the zero-benchmark period.
    try expectAlmostEqual((-0.01 - 0.03 + 0.01) / (-0.02 + 0.0 - 0.01), c.downsideCaptureRatioArithmetic(), 15);
    try expectAlmostEqual(2.0 / 3.0, c.downNumberRatio(), 15);
    // Down-percentage only counts strictly negative benchmark periods.
    try testing.expectEqual(@as(f64, 1.0), c.downPercentageRatio());
}

test "matches reference" {
    const pairs = randomPairs(42, 200);
    var c = Capture{};
    for (pairs) |p| c.update(p.a, p.b);
    try assertMatches(&c, &pairs, 12);
}

test "rolling window matches reference" {
    const pairs = randomPairs(7, 120);
    const w = 9;
    var c = Capture{};
    for (pairs, 0..) |p, i| {
        if (i >= w) try c.revert(pairs[i - w].a, pairs[i - w].b);
        c.update(p.a, p.b);
        try assertMatches(&c, pairs[(if (i + 1 > w) i + 1 - w else 0) .. i + 1], 12);
    }
}

test "reset" {
    var c = Capture{};
    c.update(0.01, 0.02);
    c.update(-0.01, -0.02);
    c.reset();
    for (getters) |g| try testing.expect(math.isNan(g(&c)));
}

test "revert empty" {
    var c = Capture{};
    try testing.expectError(error.EmptyRevert, c.revert(0.01, 0.02));
}
