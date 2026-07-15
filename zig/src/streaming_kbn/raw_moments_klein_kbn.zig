const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;

/// Streaming mean, variance, skewness, kurtosis via raw power sums (x¹..x⁴)
/// with KBN (Kahan-Babuška-Neumaier) double-compensated accumulation.
///
/// Accumulates Σx, Σx², Σx³, Σx⁴ using KleinKBNAccumulator for each,
/// plus a separate Welford-style variance tracker (also KBN-compensated).
/// Converts raw sums to central moments at query time.
///
/// Supports both LIFO revert (undo the most recent update) and FIFO
/// rolling window (via the revert/update cycle) because subtracting
/// from a linear sum preserves the KBN compensation state.
///
/// Parameters:
///   ddof – Delta degrees of freedom for variance.
///          variance = m2 / (n - ddof). ddof=0 gives population, ddof=1 gives sample.
///   bias  – If true, compute population standardized moments (m3/m2^1.5).
///           If false, apply the Fisher-Pearson adjusted (bias-corrected) factor:
///           skewness_bcf = skewness_pop * sqrt(n*(n-1)) / (n-2)
///   fisher – If true, return excess kurtosis (subtract 3 so Gaussian→0).
///            If false, return raw kurtosis (Gaussian→3).
///            Applied after the bias correction when bias=false.
pub const RawMomentsKleinKBN = struct {
    n: u32 = 0,
    x1: KleinKBNAccumulator = .{},
    x2: KleinKBNAccumulator = .{},
    x3: KleinKBNAccumulator = .{},
    x4: KleinKBNAccumulator = .{},
    mean_tracker: KleinKBNAccumulator = .{},
    s: KleinKBNAccumulator = .{},
    ddof: u32 = 1,
    bias: bool = true,
    fisher: bool = true,

    /// Clears all accumulated state.
    pub fn reset(self: *RawMomentsKleinKBN) void {
        self.n = 0;
        self.x1.reset();
        self.x2.reset();
        self.x3.reset();
        self.x4.reset();
        self.mean_tracker.reset();
        self.s.reset();
    }

    /// Adds a new sample x.
    pub fn update(self: *RawMomentsKleinKBN, x: f64) void {
        self.n += 1;
        self.x1.update(x);
        const x2 = x * x;
        self.x2.update(x2);
        const x3 = x2 * x;
        self.x3.update(x3);
        const x4 = x3 * x;
        self.x4.update(x4);

        const nf = @as(f64, @floatFromInt(self.n));
        const delta = x - self.mean_tracker.value();
        self.mean_tracker.update(delta / nf);
        self.s.update(delta * (x - self.mean_tracker.value()));
    }

    /// Removes a previously added sample x.
    pub fn revert(self: *RawMomentsKleinKBN, x: f64) void {
        self.n -= 1;
        self.x1.update(-x);
        const x2 = x * x;
        self.x2.update(-x2);
        const x3 = x2 * x;
        self.x3.update(-x3);
        const x4 = x3 * x;
        self.x4.update(-x4);

        const delta = x - self.mean_tracker.value();
        const nf = @as(f64, @floatFromInt(self.n));
        self.mean_tracker.update(-delta / nf);
        self.s.update(-delta * (x - self.mean_tracker.value()));
    }

    /// Returns the current arithmetic mean.
    pub fn mean(self: *const RawMomentsKleinKBN) f64 {
        return self.mean_tracker.value();
    }

    /// Returns the current variance. Returns NaN if n <= ddof.
    pub fn variance(self: *const RawMomentsKleinKBN) f64 {
        const nf = @as(f64, @floatFromInt(self.n)) - @as(f64, @floatFromInt(self.ddof));
        if (nf <= 0.0) return math.nan(f64);
        const sv = self.s.value();
        if (sv < 0.0) return math.nan(f64);
        return sv / nf;
    }

    /// Returns the current standard deviation. Returns NaN if n <= ddof.
    pub fn standardDeviation(self: *const RawMomentsKleinKBN) f64 {
        const nf = @as(f64, @floatFromInt(self.n)) - @as(f64, @floatFromInt(self.ddof));
        if (nf <= 0.0) return math.nan(f64);
        return math.sqrt(self.s.value() / nf);
    }

    /// Returns the current skewness. Returns NaN if n < 3.
    pub fn skewness(self: *const RawMomentsKleinKBN) f64 {
        const N = @as(f64, @floatFromInt(self.n));
        if (self.n < 3) return math.nan(f64);
        const A = self.x1.value() / N;
        const B = self.x2.value() / N - A * A;
        if (B <= 1e-14) return math.nan(f64);
        const R = math.sqrt(B);
        const C = self.x3.value() / N - A * A * A - 3.0 * A * B;
        const g1 = C / (R * R * R);
        if (self.bias) return g1;
        return g1 * math.sqrt(N * (N - 1.0)) / (N - 2.0);
    }

    /// Returns the current kurtosis. Returns NaN if n <= 3.
    pub fn kurtosis(self: *const RawMomentsKleinKBN) f64 {
        const N = @as(f64, @floatFromInt(self.n));
        if (self.n <= 3) return math.nan(f64);
        const A = self.x1.value() / N;
        var R = A * A;
        const B = self.x2.value() / N - R;
        if (B <= 1e-14) return math.nan(f64);
        R *= A;
        const C = self.x3.value() / N - R - 3.0 * A * B;
        R *= A;
        const D = self.x4.value() / N - R - 6.0 * B * A * A - 4.0 * C * A;
        const raw = D / (B * B);
        if (!self.bias) {
            const adj = ((N * N - 1.0) * raw - 3.0 * (N - 1.0) * (N - 1.0)) / ((N - 2.0) * (N - 3.0));
            if (self.fisher) return adj;
            return adj + 3.0;
        }
        if (self.fisher) return raw - 3.0;
        return raw;
    }
};

// ── Tests ──────────────────────────────────────────────────────────────────

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

const baconData = [_]f64{
    0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067,
    -0.014, 0.040, -0.005, 0.081, 0.040, -0.037, -0.061, 0.017,
    -0.049, -0.022, 0.070, 0.058, -0.065, 0.024, -0.005, -0.009,
};

test "simple update" {
    var m = RawMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    for ([_]f64{ 1.0, 2.0, 3.0, 4.0 }) |x| {
        m.update(x);
    }
    try testing.expect(almostEqual(m.mean(), 2.5, 1e-15));
    try testing.expect(almostEqual(m.variance(), 1.25, 1e-15));
    try testing.expect(almostEqual(m.skewness(), 0.0, 1e-14));
    try testing.expect(almostEqual(m.kurtosis(), -1.36, 1e-13));
}

test "standard deviation" {
    var m = RawMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    for ([_]f64{ 1.0, 2.0, 3.0, 4.0 }) |x| {
        m.update(x);
    }
    const expected = math.sqrt(m.variance());
    try testing.expect(almostEqual(m.standardDeviation(), expected, 1e-15));
}

test "ddof" {
    var m = RawMomentsKleinKBN{ .ddof = 1, .bias = true, .fisher = true };
    for ([_]f64{ 1.0, 2.0, 3.0 }) |x| {
        m.update(x);
    }
    try testing.expect(almostEqual(m.variance(), 1.0, 1e-15));
}

test "compare scipy" {
    var m = RawMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    for (baconData) |x| {
        m.update(x);
    }
    try testing.expect(almostEqual(m.mean(), 0.009000000000000001, 1e-15));
    try testing.expect(almostEqual(m.variance(), 0.0014989166666666666, 1e-14));
    try testing.expect(almostEqual(m.skewness(), -0.08256245520856798, 1e-14));
    try testing.expect(almostEqual(m.kurtosis(), -0.5675462058921257, 1e-13));
}

test "compare scipy bias false" {
    var m = RawMomentsKleinKBN{ .ddof = 0, .bias = false, .fisher = true };
    for (baconData) |x| {
        m.update(x);
    }
    try testing.expect(almostEqual(m.skewness(), -0.08817174934967527, 1e-14));
    try testing.expect(almostEqual(m.kurtosis(), -0.40766032118608714, 1e-13));
}

test "revert roundtrip" {
    const data = [_]f64{ 1.0, 2.0, 3.0, 4.0, 5.0 };
    var m = RawMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    for (data) |x| {
        m.update(x);
    }
    var i: usize = data.len - 1;
    while (i > 0) : (i -= 1) {
        m.revert(data[i]);
    }
    try testing.expect(m.n == 1);
    try testing.expect(almostEqual(m.mean(), 1.0, 1e-15));
}

test "revert partial" {
    const data = [_]f64{ 10.0, 18.0, 5.0, 12.0, 7.0 };
    var m_full = RawMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    var m_part = RawMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    for (data) |x| {
        m_full.update(x);
    }
    for (data[0..4]) |x| {
        m_part.update(x);
    }
    m_full.revert(data[4]);

    try testing.expect(almostEqual(m_full.mean(), m_part.mean(), 1e-15));
    try testing.expect(almostEqual(m_full.variance(), m_part.variance(), 1e-15));
    try testing.expect(almostEqual(m_full.skewness(), m_part.skewness(), 1e-14));
    try testing.expect(almostEqual(m_full.kurtosis(), m_part.kurtosis(), 1e-13));
}

test "reset" {
    var m = RawMomentsKleinKBN{};
    m.update(10.0);
    m.reset();
    try testing.expect(m.n == 0);
    try testing.expect(math.isNan(m.variance()));
}
