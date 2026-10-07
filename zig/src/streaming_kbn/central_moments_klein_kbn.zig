const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;

// Central moments with Klein KBN (Kahan-Babuška-Neumaier)
// compensated summation for improved numerical stability.
//
// References:
//   P. Pébay, "Formulas for Robust, One-Pass Parallel Computation
//     of Covariances and Arbitrary-Order Statistical Moments",
//     Sandia Report SAND2008-6212 (2008).
//   https://www.johndcook.com/skewness_kurtosis.html
//   https://github.com/kuiperzone/Compensated-Accumulators

/// Errors returned by `CentralMomentsKleinKBN`.
pub const Error = error{EmptyRevert};

/// Streaming mean, variance, skewness, kurtosis via Pébay's central moment
/// update with Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.
///
/// Maintains the mean M₁ and the sums of central powers
///
///     M₂ = Σ(x - x̄)²,   M₃ = Σ(x - x̄)³,   M₄ = Σ(x - x̄)⁴
///
/// (each as a KleinKBNAccumulator), updated in O(1) per sample.
/// The population central moments are μₖ = Mₖ / n.
///
/// Avoids the catastrophic cancellation inherent in converting raw power
/// sums Σxᵏ to central moments.  This matters for data with a large mean
/// relative to its spread.  Inverse updates can remove any previously
/// added sample, including the oldest sample in a FIFO rolling window.
/// Reversion clears the compensation terms, so repeated removals can
/// accumulate rounding error.
///
/// Parameters (public fields)
/// --------------------------
/// ddof : nonnegative int, default=1
///     Delta degrees of freedom for variance.
///     variance = M₂ / (n - ddof).  ddof=0 gives population, ddof=1 gives sample.
/// bias : bool, default=true
///     If true, return the biased (population) skewness and kurtosis.
///     If false, apply the bias corrections (see Notes).
/// fisher : bool, default=true
///     If true, return excess kurtosis (subtract 3 so Gaussian→0).
///     If false, return raw (Pearson) kurtosis (Gaussian→3).
///     Applied after the bias correction when bias=false.
///
/// Notes
/// -----
/// The results match scipy.stats.skew(bias=...) and
/// scipy.stats.kurtosis(bias=..., fisher=...).
///
/// Skewness (bias=true), requires n ≥ 2:
///     g₁ = μ₃ / μ₂^1.5 = √n · M₃ / M₂^1.5
///
/// Skewness (bias=false), requires n ≥ 3:
///     G₁ = g₁ · √(n·(n-1)) / (n-2)
///
/// Kurtosis (bias=true), requires n ≥ 2:
///     β₂ = μ₄ / μ₂² = n · M₄ / M₂²
///     fisher=true:  g₂ = β₂ - 3
///     fisher=false: β₂
///
/// Kurtosis (bias=false), requires n ≥ 4:
///     G₂ = ((n²-1) · β₂  -  3·(n-1)²) / ((n-2)·(n-3))
///     fisher=true:  G₂
///     fisher=false: G₂ + 3
///
/// Skewness and kurtosis are NaN when M₂ = 0 (constant data).
pub const CentralMomentsKleinKBN = struct {
    ddof: u32 = 1,
    bias: bool = true,
    fisher: bool = true,
    count: usize = 0,
    m1: KleinKBNAccumulator = .{},
    m2: KleinKBNAccumulator = .{},
    m3: KleinKBNAccumulator = .{},
    m4: KleinKBNAccumulator = .{},

    const Self = @This();

    /// Clears all accumulated state.
    pub fn reset(self: *Self) void {
        self.count = 0;
        self.m1.reset();
        self.m2.reset();
        self.m3.reset();
        self.m4.reset();
    }

    /// Adds a sample x using Pébay's update (n = count after adding x):
    ///
    ///     δ    = x − M₁
    ///     δₙ   = δ / n
    ///     term = δ · δₙ · (n − 1)
    ///
    ///     M₁ += δₙ
    ///     M₄ += term·δₙ²·(n²−3n+3) + 6·δₙ²·M₂ − 4·δₙ·M₃
    ///     M₃ += term·δₙ·(n−2) − 3·δₙ·M₂
    ///     M₂ += term
    ///
    /// M₄ and M₃ are updated before M₂ and M₃ respectively, because
    /// they use the values from before x was added.
    pub fn update(self: *Self, x: f64) void {
        const n_old: f64 = @floatFromInt(self.count);
        self.count += 1;
        const n_new: f64 = @floatFromInt(self.count);
        const delta = x - self.m1.value();
        const delta_n = delta / n_new;
        const delta_n2 = delta_n * delta_n;
        const term = delta * delta_n * n_old;
        const m2v = self.m2.value();
        const m3v = self.m3.value();
        self.m1.update(delta_n);
        self.m4.update(term * delta_n2 * (n_new * n_new - 3 * n_new + 3) + 6 * delta_n2 * m2v - 4 * delta_n * m3v);
        self.m3.update(term * delta_n * (n_new - 2) - 3 * delta_n * m2v);
        self.m2.update(term);
    }

    /// Removes a previously added sample x, regardless of insertion order.
    /// Reverting a value that was never added corrupts the state.
    ///
    /// The restored M₁–M₄ are written with KleinKBNAccumulator.set(), which
    /// clears their compensation terms.  Subsequent updates rebuild the
    /// compensation from the restored values.  Repeated reverts can
    /// accumulate rounding error, especially for large-offset data.
    ///
    /// Inverse formulas (where nₙ = count before revert, nₒ = nₙ − 1):
    ///
    ///     M₁_old = (nₙ · M₁_new − x) / nₒ            [mean undo]
    ///     δ      = x − M₁_old
    ///     δₙ     = δ / nₙ
    ///     term   = δ · δₙ · nₒ
    ///
    ///     M₂_old = M₂_new − term
    ///     M₃_old = M₃_new − (term·δₙ·(nₙ−2) − 3·δₙ·M₂_old)
    ///     M₄_old = M₄_new − (term·δₙ²·(nₙ²−3nₙ+3)
    ///                         + 6·δₙ²·M₂_old − 4·δₙ·M₃_old)
    ///
    /// Returns `error.EmptyRevert` if there are no samples.
    pub fn revert(self: *Self, x: f64) Error!void {
        if (self.count == 0) return error.EmptyRevert;
        if (self.count == 1) {
            self.reset();
            return;
        }
        const n_new: f64 = @floatFromInt(self.count);
        const n_old: f64 = @floatFromInt(self.count - 1);

        const m1_new = self.m1.value();
        const m2_new = self.m2.value();
        const m3_new = self.m3.value();
        const m4_new = self.m4.value();

        const m1_old = (n_new * m1_new - x) / n_old;
        const delta = x - m1_old;
        const delta_n = delta / n_new;
        const delta_n2 = delta_n * delta_n;
        const term = delta * delta_n * n_old;

        const m2_old = m2_new - term;
        const m3_old = m3_new - (term * delta_n * (n_new - 2) - 3 * delta_n * m2_old);
        const m4_old = m4_new - (term * delta_n2 * (n_new * n_new - 3 * n_new + 3) + 6 * delta_n2 * m2_old - 4 * delta_n * m3_old);

        self.count -= 1;
        self.m1.set(m1_old);
        self.m2.set(m2_old);
        self.m3.set(m3_old);
        self.m4.set(m4_old);
    }

    /// The number of samples.
    pub fn n(self: *const Self) usize {
        return self.count;
    }

    /// The arithmetic mean (0.0 when empty).
    pub fn mean(self: *const Self) f64 {
        return self.m1.value();
    }

    /// The variance M₂ / (n - ddof), NaN when n ≤ ddof.
    ///
    /// A slightly negative M₂ caused by rounding after revert()
    /// is clamped to zero.
    pub fn variance(self: *const Self) f64 {
        if (self.count <= self.ddof) return math.nan(f64);
        const d: f64 = @floatFromInt(self.count - self.ddof);
        return @max(self.m2.value(), 0.0) / d;
    }

    /// The square root of the variance, NaN when n ≤ ddof.
    pub fn standardDeviation(self: *const Self) f64 {
        const v = self.variance();
        return if (math.isNan(v)) v else @sqrt(v);
    }

    /// The skewness g₁ (bias=true) or G₁ (bias=false); see the type Notes.
    pub fn skewness(self: *const Self) f64 {
        const m2v = self.m2.value();
        if (self.count < 2 or m2v <= 0) return math.nan(f64);
        const nf: f64 = @floatFromInt(self.count);
        const g1 = @sqrt(nf) * self.m3.value() / (m2v * @sqrt(m2v));
        if (self.bias) return g1;
        if (self.count < 3) return math.nan(f64);
        return g1 * @sqrt(nf * (nf - 1)) / (nf - 2);
    }

    /// The kurtosis selected by bias and fisher; see the type Notes.
    pub fn kurtosis(self: *const Self) f64 {
        const m2v = self.m2.value();
        if (self.count < 2 or m2v <= 0) return math.nan(f64);
        const nf: f64 = @floatFromInt(self.count);
        const b2 = nf * self.m4.value() / (m2v * m2v);
        if (self.bias) return if (self.fisher) b2 - 3.0 else b2;
        if (self.count < 4) return math.nan(f64);
        const g2 = ((nf * nf - 1) * b2 - 3 * ((nf - 1) * (nf - 1))) / ((nf - 2) * (nf - 3));
        return if (self.fisher) g2 else g2 + 3.0;
    }
};

// ── Tests ──────────────────────────────────────────────────────────────────

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

fn feed(m: *CentralMomentsKleinKBN, data: []const f64) void {
    for (data) |x| m.update(x);
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

/// statistics.fmean: fsum(data) / len(data).
fn fmean(data: []const f64) f64 {
    return fsum(data) / @as(f64, @floatFromInt(data.len));
}

/// Population central moment fsum((x - mean)^k) / len(data) for k = 2..4
/// (also statistics.pvariance for k = 2, up to final rounding).
fn centralMoment(data: []const f64, mu: f64, k: u32) f64 {
    var buf: [64]f64 = undefined;
    for (data, 0..) |x, i| {
        const d = x - mu;
        buf[i] = switch (k) {
            2 => d * d,
            3 => d * d * d,
            else => d * d * d * d,
        };
    }
    return fsum(buf[0..data.len]) / @as(f64, @floatFromInt(data.len));
}

// Bacon, Carl R., Practical Portfolio Performance Measurement and
// Attribution, 2nd ed., Wiley, 2008, p. 65 (portfolio returns).
const bacon = [_]f64{
    0.003,  0.026,  0.011,  -0.010,
    0.015,  0.025,  0.016,  0.067,
    -0.014, 0.040,  -0.005, 0.081,
    0.040,  -0.037, -0.061, 0.017,
    -0.049, -0.022, 0.070,  0.058,
    -0.065, 0.024,  -0.005, -0.009,
};

// Reference values for bacon, computed with exact rational arithmetic
// (fractions.Fraction on the binary float inputs, square roots with
// 50-digit decimal.Decimal), rounded to the nearest float.  The names
// follow scipy.stats: skew(bias=...), kurtosis(bias=..., fisher=...).
const mean_ref: f64 = 0.009000000000000001;
const variance_ddof_0: f64 = 0.0014989166666666668;
const variance_ddof_1: f64 = 0.0015640869565217393;
const std_ddof_0: f64 = 0.03871584516275819;
const std_ddof_1: f64 = 0.039548539246370897;
const skew_biased: f64 = -0.08256245520856804; // skew(bias=True)
const skew_unbiased: f64 = -0.08817174934967535; // skew(bias=False)
const kurt_biased_fisher: f64 = -0.5675462058921257; // kurtosis(bias=True, fisher=True)
const kurt_biased_pearson: f64 = 2.4324537941078743; // kurtosis(bias=True, fisher=False)
const kurt_unbiased_fisher: f64 = -0.40766032118608714; // kurtosis(bias=False, fisher=True)
const kurt_unbiased_pearson: f64 = 2.592339678813913; // kurtosis(bias=False, fisher=False)

// Same statistics for [1e4 + x for x in bacon], exact for the shifted floats.
const offset_skew_biased: f64 = -0.08256245521966786;
const offset_kurt_biased_fisher: f64 = -0.5675462058934164;

test "simple update" {
    var m = CentralMomentsKleinKBN{ .ddof = 0 };
    feed(&m, &.{ 1.0, 2.0, 3.0, 4.0 });
    try testing.expectEqual(@as(usize, 4), m.n());
    try testing.expect(almostEqual(m.mean(), 2.5, 1e-15));
    try testing.expect(almostEqual(m.variance(), 1.25, 1e-15));
    try testing.expect(almostEqual(m.skewness(), 0.0, 1e-14));
    try testing.expect(almostEqual(m.kurtosis(), -1.36, 1e-13));
}

test "bacon mean variance" {
    var m0 = CentralMomentsKleinKBN{ .ddof = 0 };
    feed(&m0, &bacon);
    var m1 = CentralMomentsKleinKBN{ .ddof = 1 };
    feed(&m1, &bacon);
    try testing.expect(almostEqual(m0.mean(), mean_ref, 1e-16));
    try testing.expect(almostEqual(m0.variance(), variance_ddof_0, 1e-16));
    try testing.expect(almostEqual(m1.variance(), variance_ddof_1, 1e-16));
    try testing.expect(almostEqual(m0.standardDeviation(), std_ddof_0, 1e-15));
    try testing.expect(almostEqual(m1.standardDeviation(), std_ddof_1, 1e-15));
}

test "bacon skewness kurtosis" {
    const Case = struct { bias: bool, fisher: bool, skew: f64, kurt: f64 };
    const cases = [_]Case{
        .{ .bias = true, .fisher = true, .skew = skew_biased, .kurt = kurt_biased_fisher },
        .{ .bias = true, .fisher = false, .skew = skew_biased, .kurt = kurt_biased_pearson },
        .{ .bias = false, .fisher = true, .skew = skew_unbiased, .kurt = kurt_unbiased_fisher },
        .{ .bias = false, .fisher = false, .skew = skew_unbiased, .kurt = kurt_unbiased_pearson },
    };
    for (cases) |c| {
        var m = CentralMomentsKleinKBN{ .ddof = 0, .bias = c.bias, .fisher = c.fisher };
        feed(&m, &bacon);
        try testing.expect(almostEqual(m.skewness(), c.skew, 1e-14));
        try testing.expect(almostEqual(m.kurtosis(), c.kurt, 1e-13));
    }
}

test "large offset" {
    // Central moments don't suffer from the cancellation of raw power sums.
    var m = CentralMomentsKleinKBN{ .ddof = 0 };
    for (bacon) |x| m.update(1e4 + x);
    try testing.expect(almostEqual(m.mean(), 1e4 + mean_ref, 1e-11));
    try testing.expect(almostEqual(m.variance(), variance_ddof_0, 1e-13));
    try testing.expect(almostEqual(m.skewness(), offset_skew_biased, 1e-10));
    try testing.expect(almostEqual(m.kurtosis(), offset_kurt_biased_fisher, 1e-10));
}

test "scale invariance" {
    var m = CentralMomentsKleinKBN{ .ddof = 0 };
    for (bacon) |x| m.update(x * 1e-6);
    try testing.expect(almostEqual(m.skewness(), skew_biased, 1e-14));
    try testing.expect(almostEqual(m.kurtosis(), kurt_biased_fisher, 1e-13));
}

test "empty" {
    const m = CentralMomentsKleinKBN{};
    try testing.expectEqual(@as(usize, 0), m.n());
    try testing.expectEqual(@as(f64, 0.0), m.mean());
    try testing.expect(math.isNan(m.variance()));
    try testing.expect(math.isNan(m.standardDeviation()));
    try testing.expect(math.isNan(m.skewness()));
    try testing.expect(math.isNan(m.kurtosis()));
}

test "ddof" {
    var m = CentralMomentsKleinKBN{ .ddof = 1 };
    feed(&m, &.{ 1.0, 2.0, 3.0 });
    try testing.expect(almostEqual(m.variance(), 1.0, 1e-15));
    try testing.expect(almostEqual(m.standardDeviation(), 1.0, 1e-15));
    var m1 = CentralMomentsKleinKBN{ .ddof = 1 };
    m1.update(1.0);
    try testing.expect(math.isNan(m1.variance()));
    try testing.expect(math.isNan(m1.standardDeviation()));
}

// test_invalid_ddof is not ported: ddof is a u32, so negative and
// non-integer values are rejected at compile time.

test "minimum sample sizes" {
    const data = [_]f64{ 1.0, 2.0, 4.0, 8.0 };
    // (bias, fisher) -> minimum n for (skewness, kurtosis)
    const Case = struct { bias: bool, fisher: bool, skew_n: usize, kurt_n: usize };
    const cases = [_]Case{
        .{ .bias = true, .fisher = true, .skew_n = 2, .kurt_n = 2 },
        .{ .bias = true, .fisher = false, .skew_n = 2, .kurt_n = 2 },
        .{ .bias = false, .fisher = true, .skew_n = 3, .kurt_n = 4 },
        .{ .bias = false, .fisher = false, .skew_n = 3, .kurt_n = 4 },
    };
    for (cases) |c| {
        var m = CentralMomentsKleinKBN{ .ddof = 0, .bias = c.bias, .fisher = c.fisher };
        for (data, 0..) |x, i| {
            m.update(x);
            const count = i + 1;
            try testing.expectEqual(count < c.skew_n, math.isNan(m.skewness()));
            try testing.expectEqual(count < c.kurt_n, math.isNan(m.kurtosis()));
        }
    }
}

test "constant data" {
    var m = CentralMomentsKleinKBN{ .ddof = 0 };
    feed(&m, &([_]f64{3.0} ** 5));
    try testing.expectEqual(@as(f64, 3.0), m.mean());
    try testing.expectEqual(@as(f64, 0.0), m.variance());
    try testing.expectEqual(@as(f64, 0.0), m.standardDeviation());
    try testing.expect(math.isNan(m.skewness()));
    try testing.expect(math.isNan(m.kurtosis()));
}

test "revert lifo simple" {
    const data = [_]f64{ 10.0, 18.0, 5.0 };
    var m_full = CentralMomentsKleinKBN{ .ddof = 0 };
    feed(&m_full, &data);
    var m_part = CentralMomentsKleinKBN{ .ddof = 0 };
    feed(&m_part, data[0..2]);
    try m_full.revert(data[2]);

    try testing.expectEqual(@as(usize, 2), m_full.n());
    try testing.expect(almostEqual(m_full.mean(), m_part.mean(), 1e-15));
    try testing.expect(almostEqual(m_full.variance(), m_part.variance(), 1e-15));
    try testing.expect(almostEqual(m_full.skewness(), m_part.skewness(), 1e-14));
    try testing.expect(almostEqual(m_full.kurtosis(), m_part.kurtosis(), 1e-13));
}

test "revert lifo bacon" {
    const Case = struct { bias: bool, fisher: bool };
    const cases = [_]Case{ .{ .bias = true, .fisher = true }, .{ .bias = false, .fisher = false } };
    for (cases) |c| {
        var m_full = CentralMomentsKleinKBN{ .ddof = 0, .bias = c.bias, .fisher = c.fisher };
        feed(&m_full, &bacon);
        var m_part = CentralMomentsKleinKBN{ .ddof = 0, .bias = c.bias, .fisher = c.fisher };
        feed(&m_part, bacon[0 .. bacon.len - 1]);
        try m_full.revert(bacon[bacon.len - 1]);

        try testing.expect(almostEqual(m_full.mean(), m_part.mean(), 1e-15));
        try testing.expect(almostEqual(m_full.variance(), m_part.variance(), 1e-15));
        try testing.expect(almostEqual(m_full.skewness(), m_part.skewness(), 1e-13));
        try testing.expect(almostEqual(m_full.kurtosis(), m_part.kurtosis(), 1e-12));
    }
}

test "revert then update" {
    var m = CentralMomentsKleinKBN{ .ddof = 0 };
    feed(&m, &bacon);
    var i: usize = bacon.len;
    while (i > 12) {
        i -= 1;
        try m.revert(bacon[i]);
    }
    feed(&m, bacon[12..]);
    try testing.expect(almostEqual(m.mean(), mean_ref, 1e-15));
    try testing.expect(almostEqual(m.variance(), variance_ddof_0, 1e-15));
    try testing.expect(almostEqual(m.skewness(), skew_biased, 1e-12));
    try testing.expect(almostEqual(m.kurtosis(), kurt_biased_fisher, 1e-12));
}

test "revert lifo roundtrip" {
    var m = CentralMomentsKleinKBN{ .ddof = 0 };
    feed(&m, &bacon);
    var i: usize = bacon.len;
    while (i > 0) {
        i -= 1;
        try m.revert(bacon[i]);
    }
    try testing.expectEqual(@as(usize, 0), m.n());
    try testing.expectEqual(@as(f64, 0.0), m.mean());
    try testing.expect(math.isNan(m.variance()));
}

test "revert oldest and middle" {
    var m = CentralMomentsKleinKBN{ .ddof = 0 };
    feed(&m, &.{ 0.0, 1.0, 2.0, 4.0, 8.0 });
    const removed = [_]f64{ 0.0, 2.0 };
    // The data remaining after each removal.
    const remaining = [_][]const f64{ &.{ 1.0, 2.0, 4.0, 8.0 }, &.{ 1.0, 4.0, 8.0 } };
    for (removed, remaining) |r, data| {
        try m.revert(r);
        const mu = fmean(data);
        const mu2 = centralMoment(data, mu, 2);
        const mu3 = centralMoment(data, mu, 3);
        const mu4 = centralMoment(data, mu, 4);
        try testing.expectEqual(data.len, m.n());
        try testing.expect(almostEqual(m.mean(), mu, 1e-14));
        try testing.expect(almostEqual(m.variance(), mu2, 1e-14));
        try testing.expect(almostEqual(m.skewness(), mu3 / math.pow(f64, mu2, 1.5), 1e-13));
        try testing.expect(almostEqual(m.kurtosis(), mu4 / math.pow(f64, mu2, 2) - 3, 1e-13));
    }
}

test "fifo rolling window" {
    var m = CentralMomentsKleinKBN{ .ddof = 0 };
    const width: usize = 6;
    for (bacon, 0..) |x, i| {
        m.update(x);
        if (i >= width) try m.revert(bacon[i - width]);
        const lo = if (i + 1 >= width) i + 1 - width else 0;
        const window = bacon[lo .. i + 1];
        try testing.expectEqual(window.len, m.n());
        const mu = fmean(window);
        try testing.expect(almostEqual(m.mean(), mu, 1e-14));
        try testing.expect(almostEqual(m.variance(), centralMoment(window, mu, 2), 1e-14));
    }
}

test "revert empty raises" {
    var m = CentralMomentsKleinKBN{};
    try testing.expectError(error.EmptyRevert, m.revert(1.0));
}

test "standard deviation is real after revert" {
    // Reverting to two equal samples can leave a tiny negative M2.
    var m = CentralMomentsKleinKBN{ .ddof = 0 };
    for ([_]f64{ 0.1, 0.1, 0.7 }) |x| m.update(x);
    try m.revert(0.7);
    try testing.expect(!math.isNan(m.standardDeviation()));
    try testing.expect(m.variance() >= 0.0);
    try testing.expect(almostEqual(m.standardDeviation(), 0.0, 1e-15));
}

test "reset" {
    var m = CentralMomentsKleinKBN{};
    feed(&m, &bacon);
    m.reset();
    try testing.expectEqual(@as(usize, 0), m.n());
    try testing.expectEqual(@as(f64, 0.0), m.mean());
    try testing.expect(math.isNan(m.variance()));
    feed(&m, &.{ 1.0, 2.0, 3.0 });
    try testing.expect(almostEqual(m.variance(), 1.0, 1e-15));
}
