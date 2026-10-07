const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;

// Raw moments with Klein KBN (Kahan-Babuška-Neumaier)
// compensated summation for improved numerical stability.
//
// References:
//   https://github.com/kuiperzone/Compensated-Accumulators
//   https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance

/// Errors returned by `RawMomentsKleinKBN`.
pub const Error = error{EmptyRevert};

/// Relative threshold below which the population variance μ₂, computed
/// from raw power sums as Σx²/n − (Σx/n)², is considered to be lost
/// in rounding error (i.e. indistinguishable from zero).
const cancellation_epsilon: f64 = 1e-14;

/// Population central moments (μ₂, μ₃, μ₄).
const CentralMoments = struct {
    mu2: f64,
    mu3: f64,
    mu4: f64,
};

/// Streaming mean, variance, skewness, kurtosis via raw power sums (x¹..x⁴)
/// with Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.
///
/// Accumulates Σx, Σx², Σx³, Σx⁴ using a KleinKBNAccumulator for each,
/// plus a separate Welford mean/variance tracker (also KBN-compensated).
/// Mean and variance come from the Welford tracker; skewness and kurtosis
/// are converted from the raw power sums at query time.
///
/// Supports removal of any previously added sample with revert(), not
/// only the most recent one, because both the power sums and Welford's
/// mean/M₂ are symmetric functions of the samples.  This makes the type
/// suitable for FIFO rolling windows (update the new sample, revert the
/// oldest one).
///
/// Accuracy caveat: converting raw power sums to central moments suffers
/// catastrophic cancellation when the mean is large compared to the spread
/// (e.g. prices rather than returns).  For such data skewness and kurtosis
/// lose precision.  Use CentralMomentsKleinKBN for such data; it supports
/// FIFO removal, but repeated removals clear compensation and can accumulate
/// rounding error.
///
/// Notation
/// --------
/// μₖ' = Σxᵏ / n are the raw moments (methods x1..x4) and μₖ are the
/// population central moments derived from them:
///
///     μ₂ = μ₂' − μ₁'²
///     μ₃ = μ₃' − μ₁'³ − 3·μ₁'·μ₂
///     μ₄ = μ₄' − μ₁'⁴ − 6·μ₁'²·μ₂ − 4·μ₁'·μ₃
///
/// g₁ = μ₃ / μ₂^1.5 is the population skewness and β₂ = μ₄ / μ₂² is the
/// population (Pearson) kurtosis.  All skewness and kurtosis methods
/// return NaN when n < 2 or μ₂ is zero relative to μ₂' (constant data).
///
/// Parameters (public fields)
/// --------------------------
/// ddof : nonnegative int, default=1
///     Delta degrees of freedom for variance and standardDeviation.
///     variance = Σ(x - x̄)² / (n - ddof).  ddof=0 gives population, ddof=1 gives sample.
/// bias : bool, default=true
///     Selects the skewness and kurtosis methods, see below.
/// fisher : bool, default=true
///     Selects the kurtosis method, see below.
///
/// The `skewness` and `kurtosis` methods dispatch as follows
/// (matching scipy.stats.skew and scipy.stats.kurtosis):
///
///     bias   fisher  skewness               kurtosis
///     true   true    skewnessMoment, g₁     kurtosisExcess, β₂ − 3
///     true   false   skewnessMoment, g₁     kurtosisMoment, β₂
///     false  true    skewnessFisher, G₁     kurtosisSampleExcess, G₂
///     false  false   skewnessFisher, G₁     kurtosisSample, G₂ + 3
///
/// The remaining variants (skewnessSample, kurtosisSampleCorrected),
/// which match the R PerformanceAnalytics package, are available as
/// separate methods.
pub const RawMomentsKleinKBN = struct {
    ddof: u32 = 1,
    bias: bool = true,
    fisher: bool = true,
    count: usize = 0,
    x1_acc: KleinKBNAccumulator = .{},
    x2_acc: KleinKBNAccumulator = .{},
    x3_acc: KleinKBNAccumulator = .{},
    x4_acc: KleinKBNAccumulator = .{},
    // Welford's mean and sum of squared deviations Σ(x - x̄)².
    mean_acc: KleinKBNAccumulator = .{},
    s_acc: KleinKBNAccumulator = .{},

    const Self = @This();

    /// Clears all accumulated state.
    pub fn reset(self: *Self) void {
        self.count = 0;
        self.x1_acc.reset();
        self.x2_acc.reset();
        self.x3_acc.reset();
        self.x4_acc.reset();
        self.mean_acc.reset();
        self.s_acc.reset();
    }

    /// Adds a sample x.
    pub fn update(self: *Self, x: f64) void {
        self.count += 1;
        self.x1_acc.update(x);
        const p2 = x * x;
        self.x2_acc.update(p2);
        const p3 = p2 * x;
        self.x3_acc.update(p3);
        const p4 = p3 * x;
        self.x4_acc.update(p4);
        // Welford: mean += (x - mean_old) / n;  S += (x - mean_old)·(x - mean_new)
        const delta = x - self.mean_acc.value();
        self.mean_acc.update(delta / @as(f64, @floatFromInt(self.count)));
        self.s_acc.update(delta * (x - self.mean_acc.value()));
    }

    /// Removes a previously added sample x (any sample, not only the most
    /// recent one).  Reverting a value that was never added corrupts the
    /// state.
    ///
    /// Returns `error.EmptyRevert` if there are no samples.
    pub fn revert(self: *Self, x: f64) Error!void {
        if (self.count == 0) return error.EmptyRevert;
        if (self.count == 1) {
            self.reset();
            return;
        }
        self.count -= 1;
        self.x1_acc.revert(x);
        const p2 = x * x;
        self.x2_acc.revert(p2);
        const p3 = p2 * x;
        self.x3_acc.revert(p3);
        const p4 = p3 * x;
        self.x4_acc.revert(p4);
        // Inverse Welford: mean_old = mean_new - (x - mean_new) / (n - 1);
        // S -= (x - mean_new)·(x - mean_old)
        const delta = x - self.mean_acc.value();
        self.mean_acc.revert(delta / @as(f64, @floatFromInt(self.count)));
        self.s_acc.revert(delta * (x - self.mean_acc.value()));
    }

    fn varianceWithDdof(self: *const Self, ddof: u32) f64 {
        // A slightly negative S caused by rounding after revert() is clamped to zero.
        if (self.count <= ddof) return math.nan(f64);
        const d: f64 = @floatFromInt(self.count - ddof);
        return @max(self.s_acc.value(), 0.0) / d;
    }

    fn standardDeviationWithDdof(self: *const Self, ddof: u32) f64 {
        const v = self.varianceWithDdof(ddof);
        return if (math.isNan(v)) v else @sqrt(v);
    }

    /// The arithmetic mean (0.0 when empty).
    pub fn mean(self: *const Self) f64 {
        return self.mean_acc.value();
    }

    /// The variance Σ(x - x̄)² / (n - ddof), NaN when n ≤ ddof.
    pub fn variance(self: *const Self) f64 {
        return self.varianceWithDdof(self.ddof);
    }

    /// The population variance Σ(x - x̄)² / n, regardless of ddof.
    pub fn varianceDdof0(self: *const Self) f64 {
        return self.varianceWithDdof(0);
    }

    /// The sample variance Σ(x - x̄)² / (n - 1), regardless of ddof.
    pub fn varianceDdof1(self: *const Self) f64 {
        return self.varianceWithDdof(1);
    }

    /// The square root of variance.
    pub fn standardDeviation(self: *const Self) f64 {
        return self.standardDeviationWithDdof(self.ddof);
    }

    /// The population standard deviation, regardless of ddof.
    pub fn standardDeviationDdof0(self: *const Self) f64 {
        return self.standardDeviationWithDdof(0);
    }

    /// The sample standard deviation, regardless of ddof.
    pub fn standardDeviationDdof1(self: *const Self) f64 {
        return self.standardDeviationWithDdof(1);
    }

    /// Converts the raw power sums to population central moments
    /// (μ₂, μ₃, μ₄), see the type Notation.
    ///
    /// Returns null when n < 2 or μ₂ is lost in rounding error.
    fn centralMoments(self: *const Self) ?CentralMoments {
        if (self.count < 2) return null;
        const nf: f64 = @floatFromInt(self.count);
        const mu1 = self.x1_acc.value() / nf;
        var r = mu1 * mu1;
        const mean_x2 = self.x2_acc.value() / nf;
        const mu2 = mean_x2 - r;
        if (mu2 <= cancellation_epsilon * mean_x2) return null;
        r *= mu1;
        const mu3 = self.x3_acc.value() / nf - r - 3 * mu1 * mu2;
        r *= mu1;
        const mu4 = self.x4_acc.value() / nf - r - 6 * mu2 * mu1 * mu1 - 4 * mu3 * mu1;
        return .{ .mu2 = mu2, .mu3 = mu3, .mu4 = mu4 };
    }

    /// The population skewness g₁ = μ₃ / μ₂^1.5.
    fn g1(self: *const Self) f64 {
        const cm = self.centralMoments() orelse return math.nan(f64);
        return cm.mu3 / (cm.mu2 * @sqrt(cm.mu2));
    }

    /// The population (Pearson) kurtosis β₂ = μ₄ / μ₂².
    fn b2(self: *const Self) f64 {
        const cm = self.centralMoments() orelse return math.nan(f64);
        return cm.mu4 / (cm.mu2 * cm.mu2);
    }

    /// The 'moment' (biased, population) skewness, requires n ≥ 2:
    ///
    ///     g₁ = μ₃ / μ₂^1.5
    ///
    /// Matches scipy.stats.skew(bias=True) and PerformanceAnalytics
    /// skewness(method="moment").
    pub fn skewnessMoment(self: *const Self) f64 {
        return self.g1();
    }

    /// The 'fisher' (bias-adjusted Fisher-Pearson) skewness, requires n ≥ 3:
    ///
    ///     G₁ = g₁ · √(n(n−1)) / (n−2)
    ///
    /// Matches scipy.stats.skew(bias=False) and PerformanceAnalytics
    /// skewness(method="fisher").
    pub fn skewnessFisher(self: *const Self) f64 {
        const g = self.g1();
        if (math.isNan(g)) return math.nan(f64);
        if (self.count < 3) return math.nan(f64);
        const nf: f64 = @floatFromInt(self.count);
        return g * @sqrt(nf * (nf - 1)) / (nf - 2);
    }

    /// The 'sample' skewness, requires n ≥ 3:
    ///
    ///     g₁ · n² / ((n−1)(n−2))
    ///
    /// Matches PerformanceAnalytics skewness(method="sample").
    /// Doesn't depend on the bias parameter.
    pub fn skewnessSample(self: *const Self) f64 {
        const g = self.g1();
        if (math.isNan(g)) return math.nan(f64);
        if (self.count < 3) return math.nan(f64);
        const nf: f64 = @floatFromInt(self.count);
        return g * (nf * nf) / ((nf - 1) * (nf - 2));
    }

    /// The skewness selected by the bias parameter:
    ///
    /// - bias=true:  skewnessMoment, g₁
    /// - bias=false: skewnessFisher, G₁ = g₁ · √(n(n−1)) / (n−2)
    ///
    /// The third variant, skewnessSample, doesn't depend on bias.
    pub fn skewness(self: *const Self) f64 {
        return if (self.bias) self.skewnessMoment() else self.skewnessFisher();
    }

    /// The 'moment' (biased, population) Pearson kurtosis, requires n ≥ 2:
    ///
    ///     β₂ = μ₄ / μ₂²
    ///
    /// Matches scipy.stats.kurtosis(bias=True, fisher=False) and
    /// PerformanceAnalytics kurtosis(method="moment").
    pub fn kurtosisMoment(self: *const Self) f64 {
        return self.b2();
    }

    /// The 'excess' (biased, population) excess kurtosis, requires n ≥ 2:
    ///
    ///     β₂ − 3
    ///
    /// Matches scipy.stats.kurtosis(bias=True, fisher=True) and
    /// PerformanceAnalytics kurtosis(method="excess").
    pub fn kurtosisExcess(self: *const Self) f64 {
        const b = self.b2();
        if (math.isNan(b)) return math.nan(f64);
        return b - 3;
    }

    /// The 'sample excess' (unbiased) excess kurtosis, requires n ≥ 4:
    ///
    ///     G₂ = ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3))
    ///
    /// Matches scipy.stats.kurtosis(bias=False, fisher=True) and
    /// PerformanceAnalytics kurtosis(method="sample_excess").
    pub fn kurtosisSampleExcess(self: *const Self) f64 {
        const b = self.b2();
        if (math.isNan(b)) return math.nan(f64);
        if (self.count <= 3) return math.nan(f64);
        const nf: f64 = @floatFromInt(self.count);
        return ((nf * nf - 1) * b - 3 * ((nf - 1) * (nf - 1))) / ((nf - 2) * (nf - 3));
    }

    /// The 'sample' (unbiased) Pearson kurtosis, requires n ≥ 4, calculated
    /// as the 'sample excess' kurtosis plus 3:
    ///
    ///     G₂ + 3 = ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3)) + 3
    ///
    /// Matches scipy.stats.kurtosis(bias=False, fisher=False).
    ///
    /// The PerformanceAnalytics kurtosis(method="sample") variant is
    /// available as kurtosisSampleCorrected; it is larger than this one
    /// by (9n−15) / ((n−2)(n−3)), approximately 0.44 for n = 24.
    pub fn kurtosisSample(self: *const Self) f64 {
        const b = self.b2();
        if (math.isNan(b)) return math.nan(f64);
        if (self.count <= 3) return math.nan(f64);
        const nf: f64 = @floatFromInt(self.count);
        return ((nf * nf - 1) * b - 3 * ((nf - 1) * (nf - 1))) / ((nf - 2) * (nf - 3)) + 3;
    }

    /// The PerformanceAnalytics 'sample' (unbiased) Pearson kurtosis,
    /// requires n ≥ 4:
    ///
    ///     (n²−1)·β₂ / ((n−2)(n−3))
    ///
    /// Matches PerformanceAnalytics kurtosis(method="sample").
    /// Doesn't depend on the bias and fisher parameters.
    ///
    /// It differs from kurtosisSample (G₂ + 3) by (9n−15) / ((n−2)(n−3)),
    /// approximately 0.44 for n = 24.
    pub fn kurtosisSampleCorrected(self: *const Self) f64 {
        const b = self.b2();
        if (math.isNan(b)) return math.nan(f64);
        if (self.count <= 3) return math.nan(f64);
        const nf: f64 = @floatFromInt(self.count);
        return b * (nf * nf - 1) / ((nf - 2) * (nf - 3));
    }

    /// The kurtosis selected by the bias and fisher parameters:
    ///
    /// - bias=true,  fisher=true:  kurtosisExcess, β₂ − 3
    /// - bias=true,  fisher=false: kurtosisMoment, β₂
    /// - bias=false, fisher=true:  kurtosisSampleExcess, G₂
    /// - bias=false, fisher=false: kurtosisSample, G₂ + 3
    pub fn kurtosis(self: *const Self) f64 {
        if (self.bias) {
            return if (self.fisher) self.kurtosisExcess() else self.kurtosisMoment();
        }
        return if (self.fisher) self.kurtosisSampleExcess() else self.kurtosisSample();
    }

    /// The sum Σx.
    pub fn x1Sum(self: *const Self) f64 {
        return self.x1_acc.value();
    }

    /// The sum Σx².
    pub fn x2Sum(self: *const Self) f64 {
        return self.x2_acc.value();
    }

    /// The sum Σx³.
    pub fn x3Sum(self: *const Self) f64 {
        return self.x3_acc.value();
    }

    /// The sum Σx⁴.
    pub fn x4Sum(self: *const Self) f64 {
        return self.x4_acc.value();
    }

    fn rawMoment(self: *const Self, acc: *const KleinKBNAccumulator) f64 {
        if (self.count == 0) return math.nan(f64);
        return acc.value() / @as(f64, @floatFromInt(self.count));
    }

    /// The first raw moment Σx / n (NaN when empty).
    pub fn x1(self: *const Self) f64 {
        return self.rawMoment(&self.x1_acc);
    }

    /// The second raw moment Σx² / n (NaN when empty).
    pub fn x2(self: *const Self) f64 {
        return self.rawMoment(&self.x2_acc);
    }

    /// The third raw moment Σx³ / n (NaN when empty).
    pub fn x3(self: *const Self) f64 {
        return self.rawMoment(&self.x3_acc);
    }

    /// The fourth raw moment Σx⁴ / n (NaN when empty).
    pub fn x4(self: *const Self) f64 {
        return self.rawMoment(&self.x4_acc);
    }

    /// The number of samples.
    pub fn n(self: *const Self) usize {
        return self.count;
    }
};

// ── Tests ──────────────────────────────────────────────────────────────────

// https://github.com/medo64/Medo/blob/main/tests/Tests.Medo/Math/WelfordVariance.cs
// https://github.com/andrewuhl/RollingWindow/blob/master/src/RollingWindow.cpp
// https://github.com/ajcr/rolling/blob/master/rolling/similarity.py

const Getter = *const fn (*const RawMomentsKleinKBN) f64;

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

fn feed(m: *RawMomentsKleinKBN, data: []const f64) void {
    for (data) |x| m.update(x);
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
// 50-digit decimal.Decimal), rounded to the nearest float.
const exp_mean: f64 = 0.009000000000000001;
const exp_variance_ddof_0: f64 = 0.0014989166666666668;
const exp_variance_ddof_1: f64 = 0.0015640869565217393;
const exp_standard_deviation_ddof_0: f64 = 0.03871584516275819;
const exp_standard_deviation_ddof_1: f64 = 0.039548539246370897;
const exp_skewness_moment: f64 = -0.08256245520856804; // scipy skew(bias=True)
const exp_skewness_fisher: f64 = -0.08817174934967535; // scipy skew(bias=False)
const exp_skewness_sample: f64 = -0.09398413873544505; // R PerformanceAnalytics "sample"
const exp_kurtosis_moment: f64 = 2.4324537941078743; // scipy kurtosis(bias=True, fisher=False)
const exp_kurtosis_excess: f64 = -0.5675462058921257; // scipy kurtosis(bias=True, fisher=True)
const exp_kurtosis_sample_excess: f64 = -0.40766032118608714; // scipy kurtosis(bias=False, fisher=True)
const exp_kurtosis_sample: f64 = 2.592339678813913; // scipy kurtosis(bias=False, fisher=False)
const exp_kurtosis_sample_corrected: f64 = 3.027404613878848; // R PerformanceAnalytics "sample"

const Expected = struct { get: Getter, value: f64 };

const expected_all = [_]Expected{
    .{ .get = &RawMomentsKleinKBN.mean, .value = exp_mean },
    .{ .get = &RawMomentsKleinKBN.varianceDdof0, .value = exp_variance_ddof_0 },
    .{ .get = &RawMomentsKleinKBN.varianceDdof1, .value = exp_variance_ddof_1 },
    .{ .get = &RawMomentsKleinKBN.standardDeviationDdof0, .value = exp_standard_deviation_ddof_0 },
    .{ .get = &RawMomentsKleinKBN.standardDeviationDdof1, .value = exp_standard_deviation_ddof_1 },
    .{ .get = &RawMomentsKleinKBN.skewnessMoment, .value = exp_skewness_moment },
    .{ .get = &RawMomentsKleinKBN.skewnessFisher, .value = exp_skewness_fisher },
    .{ .get = &RawMomentsKleinKBN.skewnessSample, .value = exp_skewness_sample },
    .{ .get = &RawMomentsKleinKBN.kurtosisMoment, .value = exp_kurtosis_moment },
    .{ .get = &RawMomentsKleinKBN.kurtosisExcess, .value = exp_kurtosis_excess },
    .{ .get = &RawMomentsKleinKBN.kurtosisSampleExcess, .value = exp_kurtosis_sample_excess },
    .{ .get = &RawMomentsKleinKBN.kurtosisSample, .value = exp_kurtosis_sample },
    .{ .get = &RawMomentsKleinKBN.kurtosisSampleCorrected, .value = exp_kurtosis_sample_corrected },
    .{ .get = &RawMomentsKleinKBN.x1Sum, .value = 0.21600000000000003 },
    .{ .get = &RawMomentsKleinKBN.x2Sum, .value = 0.037918 },
    .{ .get = &RawMomentsKleinKBN.x3Sum, .value = 0.0008738040000000003 },
    .{ .get = &RawMomentsKleinKBN.x4Sum, .value = 0.00014466403 },
    .{ .get = &RawMomentsKleinKBN.x1, .value = 0.009000000000000001 },
    .{ .get = &RawMomentsKleinKBN.x2, .value = 0.0015799166666666668 },
    .{ .get = &RawMomentsKleinKBN.x3, .value = 3.640850000000001e-05 },
    .{ .get = &RawMomentsKleinKBN.x4, .value = 6.0276679166666674e-06 },
};

// Defaults are ddof=1, bias=true, fisher=true.
const expected_by_dispatch = [_]Expected{
    .{ .get = &RawMomentsKleinKBN.mean, .value = exp_mean },
    .{ .get = &RawMomentsKleinKBN.variance, .value = exp_variance_ddof_1 },
    .{ .get = &RawMomentsKleinKBN.skewness, .value = exp_skewness_moment },
    .{ .get = &RawMomentsKleinKBN.kurtosis, .value = exp_kurtosis_excess },
};

test "simple update" {
    var m = RawMomentsKleinKBN{ .ddof = 0 };
    feed(&m, &.{ 1.0, 2.0, 3.0, 4.0 });
    try testing.expectEqual(@as(usize, 4), m.n());
    try testing.expect(almostEqual(m.mean(), 2.5, 1e-15));
    try testing.expect(almostEqual(m.variance(), 1.25, 1e-15));
    try testing.expect(almostEqual(m.skewness(), 0.0, 1e-14));
    try testing.expect(almostEqual(m.kurtosis(), -1.36, 1e-13));
}

test "bacon all properties" {
    var m = RawMomentsKleinKBN{};
    feed(&m, &bacon);
    for (expected_all) |e| {
        try testing.expect(almostEqual(e.get(&m), e.value, 1e-14));
    }
}

test "dispatch" {
    const Case = struct { bias: bool, fisher: bool, skew: Getter, skew_value: f64, kurt: Getter, kurt_value: f64 };
    const cases = [_]Case{
        .{ .bias = true, .fisher = true, .skew = &RawMomentsKleinKBN.skewnessMoment, .skew_value = exp_skewness_moment, .kurt = &RawMomentsKleinKBN.kurtosisExcess, .kurt_value = exp_kurtosis_excess },
        .{ .bias = true, .fisher = false, .skew = &RawMomentsKleinKBN.skewnessMoment, .skew_value = exp_skewness_moment, .kurt = &RawMomentsKleinKBN.kurtosisMoment, .kurt_value = exp_kurtosis_moment },
        .{ .bias = false, .fisher = true, .skew = &RawMomentsKleinKBN.skewnessFisher, .skew_value = exp_skewness_fisher, .kurt = &RawMomentsKleinKBN.kurtosisSampleExcess, .kurt_value = exp_kurtosis_sample_excess },
        .{ .bias = false, .fisher = false, .skew = &RawMomentsKleinKBN.skewnessFisher, .skew_value = exp_skewness_fisher, .kurt = &RawMomentsKleinKBN.kurtosisSample, .kurt_value = exp_kurtosis_sample },
    };
    for (cases) |c| {
        var m = RawMomentsKleinKBN{ .bias = c.bias, .fisher = c.fisher };
        feed(&m, &bacon);
        try testing.expectEqual(c.skew(&m), m.skewness());
        try testing.expectEqual(c.kurt(&m), m.kurtosis());
        try testing.expect(almostEqual(m.skewness(), c.skew_value, 1e-14));
        try testing.expect(almostEqual(m.kurtosis(), c.kurt_value, 1e-13));
    }
}

test "ddof" {
    {
        var m = RawMomentsKleinKBN{ .ddof = 0 };
        feed(&m, &bacon);
        try testing.expectEqual(m.varianceDdof0(), m.variance());
        try testing.expectEqual(m.standardDeviationDdof0(), m.standardDeviation());
    }
    {
        var m = RawMomentsKleinKBN{ .ddof = 1 };
        feed(&m, &bacon);
        try testing.expectEqual(m.varianceDdof1(), m.variance());
        try testing.expectEqual(m.standardDeviationDdof1(), m.standardDeviation());
    }
    var m = RawMomentsKleinKBN{ .ddof = 1 };
    feed(&m, &.{ 1.0, 2.0, 3.0 });
    try testing.expect(almostEqual(m.variance(), 1.0, 1e-15));
    try testing.expect(almostEqual(m.standardDeviation(), 1.0, 1e-15));
}

// test_invalid_ddof is not ported: ddof is a u32, so negative and
// non-integer values are rejected at compile time.

test "kurtosis sample corrected difference" {
    // kurtosisSampleCorrected - kurtosisSample = (9n-15) / ((n-2)(n-3))
    var m = RawMomentsKleinKBN{};
    feed(&m, &bacon);
    const nf: f64 = @floatFromInt(bacon.len);
    try testing.expect(almostEqual(
        m.kurtosisSampleCorrected() - m.kurtosisSample(),
        (9 * nf - 15) / ((nf - 2) * (nf - 3)),
        1e-14,
    ));
}

test "empty" {
    const m = RawMomentsKleinKBN{};
    try testing.expectEqual(@as(usize, 0), m.n());
    try testing.expectEqual(@as(f64, 0.0), m.mean());
    const getters = [_]Getter{
        &RawMomentsKleinKBN.variance, &RawMomentsKleinKBN.standardDeviation,
        &RawMomentsKleinKBN.skewness, &RawMomentsKleinKBN.kurtosis,
        &RawMomentsKleinKBN.x1,       &RawMomentsKleinKBN.x2,
        &RawMomentsKleinKBN.x3,       &RawMomentsKleinKBN.x4,
    };
    for (getters) |g| try testing.expect(math.isNan(g(&m)));
    try testing.expectEqual(@as(f64, 0.0), m.x1Sum());
}

test "minimum sample sizes" {
    const data = [_]f64{ 1.0, 2.0, 4.0, 8.0 };
    const MinN = struct { get: Getter, min_n: usize };
    const minimum_n = [_]MinN{
        .{ .get = &RawMomentsKleinKBN.skewnessMoment, .min_n = 2 },
        .{ .get = &RawMomentsKleinKBN.skewnessFisher, .min_n = 3 },
        .{ .get = &RawMomentsKleinKBN.skewnessSample, .min_n = 3 },
        .{ .get = &RawMomentsKleinKBN.kurtosisMoment, .min_n = 2 },
        .{ .get = &RawMomentsKleinKBN.kurtosisExcess, .min_n = 2 },
        .{ .get = &RawMomentsKleinKBN.kurtosisSampleExcess, .min_n = 4 },
        .{ .get = &RawMomentsKleinKBN.kurtosisSample, .min_n = 4 },
        .{ .get = &RawMomentsKleinKBN.kurtosisSampleCorrected, .min_n = 4 },
    };
    var m = RawMomentsKleinKBN{};
    for (data, 0..) |x, i| {
        m.update(x);
        const count = i + 1;
        for (minimum_n) |e| {
            try testing.expectEqual(count < e.min_n, math.isNan(e.get(&m)));
        }
    }
}

test "constant data" {
    var m = RawMomentsKleinKBN{ .ddof = 0 };
    feed(&m, &([_]f64{0.1} ** 5));
    try testing.expect(almostEqual(m.mean(), 0.1, 1e-16));
    try testing.expect(almostEqual(m.variance(), 0.0, 1e-16));
    try testing.expect(math.isNan(m.skewness()));
    try testing.expect(math.isNan(m.kurtosis()));
}

test "scale invariance" {
    // The cancellation threshold is relative, so tiny values work.
    var m = RawMomentsKleinKBN{};
    for (bacon) |x| m.update(x * 1e-6);
    try testing.expect(almostEqual(m.skewnessMoment(), exp_skewness_moment, 1e-13));
    try testing.expect(almostEqual(m.kurtosisExcess(), exp_kurtosis_excess, 1e-13));
}

test "large offset preserves variance but not higher moments" {
    // Welford's variance remains usable when raw-power cancellation
    // makes skewness and kurtosis unreliable.
    var m = RawMomentsKleinKBN{ .ddof = 0 };
    const big: f64 = 1e8;
    feed(&m, &.{ big, big + 1, big + 2 });
    try testing.expect(almostEqual(m.mean(), big + 1, 1e-10));
    try testing.expect(almostEqual(m.variance(), @as(f64, 2.0) / 3.0, 1e-14));
    try testing.expect(math.isNan(m.skewness()));
    try testing.expect(math.isNan(m.kurtosis()));
}

test "revert partial" {
    const data = [_]f64{ 10.0, 18.0, 5.0, 12.0, 7.0 };
    var m_full = RawMomentsKleinKBN{ .ddof = 0 };
    feed(&m_full, &data);
    var m_part = RawMomentsKleinKBN{ .ddof = 0 };
    feed(&m_part, data[0..4]);
    try m_full.revert(data[4]);
    try testing.expectEqual(@as(usize, 4), m_full.n());
    try testing.expect(almostEqual(m_full.mean(), m_part.mean(), 1e-15));
    try testing.expect(almostEqual(m_full.variance(), m_part.variance(), 1e-15));
    try testing.expect(almostEqual(m_full.skewness(), m_part.skewness(), 1e-14));
    try testing.expect(almostEqual(m_full.kurtosis(), m_part.kurtosis(), 1e-13));
}

test "revert not most recent" {
    var m = RawMomentsKleinKBN{};
    feed(&m, &bacon);
    m.update(0.5);
    try m.revert(0.5);
    var m2 = RawMomentsKleinKBN{};
    m2.update(0.5);
    feed(&m2, &bacon);
    try m2.revert(0.5); // the oldest sample
    for (expected_by_dispatch) |e| {
        try testing.expect(almostEqual(e.get(&m), e.value, 1e-13));
        try testing.expect(almostEqual(e.get(&m2), e.value, 1e-13));
    }
}

test "revert to empty" {
    var m = RawMomentsKleinKBN{ .ddof = 0 };
    feed(&m, &bacon);
    for (bacon) |x| try m.revert(x);
    try testing.expectEqual(@as(usize, 0), m.n());
    try testing.expectEqual(@as(f64, 0.0), m.mean());
    try testing.expectEqual(@as(f64, 0.0), m.x1Sum());
    try testing.expect(math.isNan(m.variance()));
    feed(&m, &.{ 1.0, 2.0, 3.0, 4.0 });
    try testing.expect(almostEqual(m.variance(), 1.25, 1e-15));
}

test "revert empty raises" {
    var m = RawMomentsKleinKBN{};
    try testing.expectError(error.EmptyRevert, m.revert(1.0));
}

test "rolling window" {
    const getters = [_]Getter{
        &RawMomentsKleinKBN.mean,                    &RawMomentsKleinKBN.variance,
        &RawMomentsKleinKBN.standardDeviation,       &RawMomentsKleinKBN.skewness,
        &RawMomentsKleinKBN.kurtosis,                &RawMomentsKleinKBN.skewnessSample,
        &RawMomentsKleinKBN.kurtosisSampleCorrected, &RawMomentsKleinKBN.x1,
        &RawMomentsKleinKBN.x2,                      &RawMomentsKleinKBN.x3,
        &RawMomentsKleinKBN.x4,
    };
    const w: usize = 5;
    var m = RawMomentsKleinKBN{ .ddof = 1, .bias = false, .fisher = true };
    for (bacon, 0..) |x, i| {
        m.update(x);
        if (i >= w) try m.revert(bacon[i - w]);
        const lo = if (i + 1 >= w) i + 1 - w else 0;
        var ref = RawMomentsKleinKBN{ .ddof = 1, .bias = false, .fisher = true };
        feed(&ref, bacon[lo .. i + 1]);
        try testing.expectEqual(ref.n(), m.n());
        for (getters) |g| {
            const actual = g(&m);
            const expected = g(&ref);
            if (math.isNan(expected)) {
                try testing.expect(math.isNan(actual));
            } else {
                try testing.expect(almostEqual(actual, expected, 1e-13));
            }
        }
    }
}

test "standard deviation is real after revert" {
    var m = RawMomentsKleinKBN{ .ddof = 0 };
    for ([_]f64{ 0.1, 0.1, 0.7 }) |x| m.update(x);
    try m.revert(0.7);
    try testing.expect(!math.isNan(m.standardDeviation()));
    try testing.expect(m.variance() >= 0.0);
    try testing.expect(almostEqual(m.standardDeviation(), 0.0, 1e-15));
}

test "variance getter has no side effects" {
    var m = RawMomentsKleinKBN{ .ddof = 0 };
    feed(&m, &bacon);
    const v = m.variance();
    _ = m.standardDeviation();
    try testing.expectEqual(v, m.variance());
    try testing.expectEqual(@as(usize, bacon.len), m.n());
}

test "reset" {
    var m = RawMomentsKleinKBN{};
    feed(&m, &bacon);
    m.reset();
    try testing.expectEqual(@as(usize, 0), m.n());
    try testing.expectEqual(@as(f64, 0.0), m.mean());
    try testing.expectEqual(@as(f64, 0.0), m.x4Sum());
    try testing.expect(math.isNan(m.variance()));
    feed(&m, &.{ 1.0, 2.0, 3.0 });
    try testing.expect(almostEqual(m.variance(), 1.0, 1e-15));
}
