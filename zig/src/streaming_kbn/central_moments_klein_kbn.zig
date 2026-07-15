const std = @import("std");
const math = std.math;
const testing = std.testing;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;

/// Streaming mean, variance, skewness, kurtosis via Pébay's central moment
/// update with KBN (Kahan-Babuška-Neumaier) double-compensated accumulation.
///
/// Maintains running sums of central moments m2, m3, m4 (as KleinKBNAccumulators)
/// updated in O(1) per sample.  Preferred over RawMomentsKleinKBN for forward-only
/// computation (no revert) because it avoids the numerical cancellation
/// inherent in converting raw power sums to central moments.
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
pub const CentralMomentsKleinKBN = struct {
    n: u32 = 0,
    m1: KleinKBNAccumulator = .{},
    m2: KleinKBNAccumulator = .{},
    m3: KleinKBNAccumulator = .{},
    m4: KleinKBNAccumulator = .{},
    ddof: u32 = 1,
    bias: bool = true,
    fisher: bool = true,

    /// Clears all accumulated state.
    pub fn reset(self: *CentralMomentsKleinKBN) void {
        self.n = 0;
        self.m1.reset();
        self.m2.reset();
        self.m3.reset();
        self.m4.reset();
    }

    /// Adds a new sample x using Pébay's central moment update formulas.
    pub fn update(self: *CentralMomentsKleinKBN, x: f64) void {
        const n_old = self.n;
        const n_new = n_old + 1;
        self.n = n_new;
        const delta = x - self.m1.value();
        const delta_n = delta / @as(f64, @floatFromInt(n_new));
        const delta_n2 = delta_n * delta_n;
        const term = delta * delta_n * @as(f64, @floatFromInt(n_old));

        self.m1.update(delta_n);
        const nf_new = @as(f64, @floatFromInt(n_new));
        self.m4.update(term * delta_n2 * (nf_new * nf_new - 3.0 * nf_new + 3.0) + 6.0 * delta_n2 * self.m2.value() - 4.0 * delta_n * self.m3.value());
        self.m3.update(term * delta_n * (nf_new - 2.0) - 3.0 * delta_n * self.m2.value());
        self.m2.update(term);
    }

    /// Removes the most recently added sample x (LIFO).
    ///
    /// Uses inverse Pébay formulas to restore prior state. The
    /// KleinKBNAccumulator.set() method resets compensation terms to zero,
    /// so subsequent updates rebuild compensation from the restored value.
    ///
    /// Only the most recent sample can be reverted (LIFO stack, not FIFO
    /// queue). For rolling-window FIFO removal use RawMomentsKleinKBN.
    pub fn revert(self: *CentralMomentsKleinKBN, x: f64) void {
        const n_new = self.n;
        if (n_new == 0) @panic("cannot revert below 0");
        const n_old = n_new - 1;
        if (n_old == 0) {
            self.n = 0;
            self.m1.reset();
            self.m2.reset();
            self.m3.reset();
            self.m4.reset();
            return;
        }

        const m1_new = self.m1.value();
        const m2_new = self.m2.value();
        const m3_new = self.m3.value();
        const m4_new = self.m4.value();

        const nf_new = @as(f64, @floatFromInt(n_new));
        const nf_old = @as(f64, @floatFromInt(n_old));

        const m1_old = (nf_new * m1_new - x) / nf_old;
        const delta = x - m1_old;
        const delta_n = delta / nf_new;
        const delta_n2 = delta_n * delta_n;
        const term = delta * delta_n * nf_old;

        const m2_old = m2_new - term;
        const m3_old = m3_new - (term * delta_n * (nf_new - 2.0) - 3.0 * delta_n * m2_old);
        const m4_old = m4_new - (term * delta_n2 * (nf_new * nf_new - 3.0 * nf_new + 3.0) + 6.0 * delta_n2 * m2_old - 4.0 * delta_n * m3_old);

        self.n = n_old;
        self.m1.set(m1_old);
        self.m2.set(m2_old);
        self.m3.set(m3_old);
        self.m4.set(m4_old);
    }

    /// Returns the current arithmetic mean.
    pub fn mean(self: *const CentralMomentsKleinKBN) f64 {
        return self.m1.value();
    }

    /// Returns the current variance. Returns NaN if n <= ddof.
    pub fn variance(self: *const CentralMomentsKleinKBN) f64 {
        const nf = @as(f64, @floatFromInt(self.n)) - @as(f64, @floatFromInt(self.ddof));
        if (nf <= 0.0) return math.nan(f64);
        return self.m2.value() / nf;
    }

    /// Returns the current standard deviation. Returns NaN if n <= ddof.
    pub fn standardDeviation(self: *const CentralMomentsKleinKBN) f64 {
        const nf = @as(f64, @floatFromInt(self.n)) - @as(f64, @floatFromInt(self.ddof));
        if (nf <= 0.0) return math.nan(f64);
        return math.sqrt(self.m2.value() / nf);
    }

    /// Returns the current skewness. Returns NaN if n < 3 or m2 <= 0.
    pub fn skewness(self: *const CentralMomentsKleinKBN) f64 {
        const N = @as(f64, @floatFromInt(self.n));
        if (self.n < 3 or self.m2.value() <= 0.0) return math.nan(f64);
        const g1 = math.sqrt(N) * self.m3.value() / math.pow(f64, self.m2.value(), 1.5);
        if (self.bias) return g1;
        return g1 * math.sqrt(N * (N - 1.0)) / (N - 2.0);
    }

    /// Returns the current kurtosis. Returns NaN if n <= 3 or m2 <= 0.
    pub fn kurtosis(self: *const CentralMomentsKleinKBN) f64 {
        const N = @as(f64, @floatFromInt(self.n));
        if (self.n <= 3 or self.m2.value() <= 0.0) return math.nan(f64);
        const raw = N * self.m4.value() / (self.m2.value() * self.m2.value());
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
    var m = CentralMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    for ([_]f64{ 1.0, 2.0, 3.0, 4.0 }) |x| {
        m.update(x);
    }
    try testing.expect(almostEqual(m.mean(), 2.5, 1e-15));
    try testing.expect(almostEqual(m.variance(), 1.25, 1e-15));
    try testing.expect(almostEqual(m.skewness(), 0.0, 1e-14));
    try testing.expect(almostEqual(m.kurtosis(), -1.36, 1e-13));
}

test "compare scipy" {
    var m = CentralMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    for (baconData) |x| {
        m.update(x);
    }
    try testing.expect(almostEqual(m.mean(), 0.009000000000000001, 1e-15));
    try testing.expect(almostEqual(m.variance(), 0.0014989166666666668, 1e-15));
    try testing.expect(almostEqual(m.skewness(), -0.08256245520856803, 1e-14));
    try testing.expect(almostEqual(m.kurtosis(), -0.5675462058921261, 1e-13));
}

test "compare scipy bias false" {
    var m = CentralMomentsKleinKBN{ .ddof = 0, .bias = false, .fisher = true };
    for (baconData) |x| {
        m.update(x);
    }
    try testing.expect(almostEqual(m.skewness(), -0.08817174934967532, 1e-14));
    try testing.expect(almostEqual(m.kurtosis(), -0.4076603211860876, 1e-13));
}

test "ddof" {
    var m = CentralMomentsKleinKBN{ .ddof = 1, .bias = true, .fisher = true };
    for ([_]f64{ 1.0, 2.0, 3.0 }) |x| {
        m.update(x);
    }
    try testing.expect(almostEqual(m.variance(), 1.0, 1e-15));
}

test "revert lifo simple" {
    const data = [_]f64{ 10.0, 18.0, 5.0 };
    var m_full = CentralMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    var m_part = CentralMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    for (data) |x| {
        m_full.update(x);
    }
    for (data[0..2]) |x| {
        m_part.update(x);
    }
    m_full.revert(data[2]);

    try testing.expect(almostEqual(m_full.mean(), m_part.mean(), 1e-15));
    try testing.expect(almostEqual(m_full.variance(), m_part.variance(), 1e-15));
    try testing.expect(math.isNan(m_full.skewness()));
    try testing.expect(math.isNan(m_full.kurtosis()));
}

test "revert lifo bacon" {
    var m_full = CentralMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    var m_part = CentralMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    for (baconData) |x| {
        m_full.update(x);
    }
    for (baconData[0 .. baconData.len - 1]) |x| {
        m_part.update(x);
    }
    m_full.revert(baconData[baconData.len - 1]);

    try testing.expect(almostEqual(m_full.mean(), m_part.mean(), 1e-15));
    try testing.expect(almostEqual(m_full.variance(), m_part.variance(), 1e-15));
    try testing.expect(almostEqual(m_full.skewness(), m_part.skewness(), 1e-14));
    try testing.expect(almostEqual(m_full.kurtosis(), m_part.kurtosis(), 1e-12));
}

test "revert lifo roundtrip" {
    var m = CentralMomentsKleinKBN{ .ddof = 0, .bias = true, .fisher = true };
    for (baconData) |x| {
        m.update(x);
    }
    var i: usize = baconData.len;
    while (i > 0) : (i -= 1) {
        m.revert(baconData[i - 1]);
    }
    try testing.expect(m.n == 0);
    try testing.expect(m.mean() == 0.0);
    try testing.expect(math.isNan(m.variance()));
}

test "reset" {
    var m = CentralMomentsKleinKBN{};
    m.update(10.0);
    m.reset();
    try testing.expect(m.n == 0);
    try testing.expect(m.mean() == 0.0);
    try testing.expect(math.isNan(m.variance()));
}
