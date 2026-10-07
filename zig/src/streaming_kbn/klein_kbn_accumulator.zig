const std = @import("std");
const testing = std.testing;

// Klein second-order Kahan-Babuška-Neumaier (KBN) compensated summation.
//
// Kahan (1965) introduced single-level compensated summation.
// Neumaier (1974) improved it with a branch on |sum| >= |x|
// (the KBN algorithm proper).  Klein (2006) generalised KBN
// to arbitrary order; this is the second-order variant, which
// applies the same KBN trick to the correction term itself.
//
// Level 1 (KBN):      t = sum + x
//                     if |sum| >= |x|: c = (sum - t) + x
//                     else:            c = (x - t) + sum
//                     sum = t
// Level 2 (Klein):    t = cs + c
//                     if |cs| >= |c|:  cc = (cs - t) + c
//                     else:            cc = (c - t) + cs
//                     cs = t
//                     ccs += cc
//
// The corrected sum is: sum + cs + ccs.
//
// References:
//   A. Klein, "A Generalized Kahan-Babuška-Summation-Algorithm",
//     Computing 76, 279-293 (2006).
//   https://github.com/kuiperzone/Compensated-Accumulators
//   https://en.wikipedia.org/wiki/Kahan_summation_algorithm

/// Klein second-order Kahan-Babuška-Neumaier (KBN) floating-point accumulator.
///
/// Maintains three terms whose sum is the corrected total:
///
/// - `sum`: the primary (naive) running sum;
/// - `cs`:  the running sum of first-level KBN corrections;
/// - `ccs`: the running sum of second-level corrections, i.e. the
///   rounding errors made while accumulating `cs` (Klein's
///   generalisation).
///
/// Unlike naive summation, KBN correctly sums sequences with extreme
/// magnitude differences (e.g. Peters' example [1.0, 1e100, 1.0, -1e100]
/// → 2.0, while naive and standard Kahan summation return 0.0).
///
/// Level 1 (Kahan-Babuška-Neumaier):
///
///     t = sum + x
///     if |sum| >= |x|:  c = (sum - t) + x
///     else:             c = (x - t) + sum
///     sum = t
///
/// The branch makes sure the larger operand comes first, so the
/// expression recovers exactly the low-order bits that were lost
/// when rounding `sum + x` to `t`.
///
/// Level 2 (Klein generalisation) applies the same technique to the
/// addition `cs + c` and accumulates its rounding error `cc`
/// into `ccs`.
///
/// The accumulator only stores sums, so `revert(x)` (adding `-x`)
/// removes any previously added value, not only the most recent one.
/// This makes it suitable for FIFO rolling windows.
pub const KleinKBNAccumulator = struct {
    sum: f64 = 0.0,
    cs: f64 = 0.0,
    ccs: f64 = 0.0,

    const Self = @This();

    /// Sets the accumulator to zero.
    pub fn reset(self: *Self) void {
        self.set(0.0);
    }

    /// Overwrites the accumulated value with x and clears both
    /// compensation terms.
    ///
    /// Prefer set() over constructing a new instance when the
    /// accumulator is stored in an object slot.
    pub fn set(self: *Self, x: f64) void {
        self.sum = x;
        self.cs = 0.0;
        self.ccs = 0.0;
    }

    /// Removes a previously added value x (equivalent to update(-x)).
    pub fn revert(self: *Self, x: f64) void {
        self.update(-x);
    }

    /// Adds x to the accumulator.
    pub fn update(self: *Self, x: f64) void {
        const s = self.sum;
        const t1 = s + x;
        const c = if (@abs(s) >= @abs(x)) (s - t1) + x else (x - t1) + s;
        self.sum = t1;

        const cs = self.cs;
        const t2 = cs + c;
        const cc = if (@abs(cs) >= @abs(c)) (cs - t2) + c else (c - t2) + cs;
        self.cs = t2;
        self.ccs += cc;
    }

    /// The compensated sum of all added values.
    pub fn value(self: *const Self) f64 {
        return self.sum + self.cs + self.ccs;
    }
};

// ── Tests ──────────────────────────────────────────────────────────────────

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

fn naiveSum(data: []const f64) f64 {
    var s: f64 = 0.0;
    for (data) |x| s += x;
    return s;
}

fn kbnSum(data: []const f64) f64 {
    var kbn = KleinKBNAccumulator{};
    for (data) |x| kbn.update(x);
    return kbn.value();
}

/// Faithful port of CPython's math.fsum (Shewchuk's exact partials with
/// round-half-even correction). Inputs are assumed finite.
fn fsum(data: []const f64) f64 {
    var p: [64]f64 = undefined;
    var n: usize = 0;
    for (data) |xin| {
        var x = xin;
        var i: usize = 0;
        var j: usize = 0;
        while (j < n) : (j += 1) {
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
        n = i;
        if (x != 0.0) {
            p[n] = x;
            n += 1;
        }
    }
    var hi: f64 = 0.0;
    if (n > 0) {
        n -= 1;
        hi = p[n];
        var lo: f64 = 0.0;
        while (n > 0) {
            const x = hi;
            n -= 1;
            const y = p[n];
            hi = x + y;
            const yr = hi - x;
            lo = y - yr;
            if (lo != 0.0) break;
        }
        if (n > 0 and ((lo < 0.0 and p[n - 1] < 0.0) or (lo > 0.0 and p[n - 1] > 0.0))) {
            const y = lo * 2.0;
            const x = hi + y;
            const yr = x - hi;
            if (y == yr) hi = x;
        }
    }
    return hi;
}

// https://en.wikipedia.org/wiki/Kahan_summation_algorithm
// A simple example due to Peters: summing [1.0, +1e100, 1.0, -1e100]
// in double precision, Kahan's algorithm yields 0.0, whereas
// Neumaier's algorithm yields the correct value 2.0.
const peters_data = [_]f64{ 1.0, 1e100, 1.0, -1e100 };

// https://github.com/numpy/numpy/issues/8786
// A badly conditioned sum, condition number ~2.188e+14.
const numpy_data = [_]f64{
    -0.41253261766461263,
    41287272281118.43,
    -1.4727977348624173e-14,
    5670.3302557520055,
    2.119245229045646e-11,
    -0.003679264134906428,
    -6.892634568678797e-14,
    -0.0006984744181630712,
    -4054136.048352595,
    -1003.101760720037,
    -1.4436349910427172e-17,
    -41287268231649.57,
};
const numpy_expected: f64 = -0.377392919181026;

// A sequence where the second-level correction is non-zero at more
// than one step, so it must be accumulated (ccs += cc), not
// overwritten (ccs = cc).  Overwriting yields -1.0000000000000001e-16.
const klein_data = [_]f64{ 1e-16, -1e16, 1.0, 1e-16, -1.0, -1e-16, -1e-32, 1e16 };
const klein_expected: f64 = 9.999999999999999e-17; // math.fsum(klein_data)

test "initial value is zero" {
    try testing.expectEqual(@as(f64, 0.0), (KleinKBNAccumulator{}).value());
}

test "peters" {
    try testing.expectEqual(@as(f64, 0.0), naiveSum(&peters_data));
    try testing.expectEqual(@as(f64, 2.0), kbnSum(&peters_data));
}

test "numpy issue" {
    try testing.expect(almostEqual(kbnSum(&numpy_data), numpy_expected, 1e-16));
    try testing.expect(!almostEqual(naiveSum(&numpy_data), numpy_expected, 1e-3));
}

test "second level correction is accumulated" {
    try testing.expectEqual(klein_expected, kbnSum(&klein_data));
    try testing.expectEqual(klein_expected, fsum(&klein_data));
}

test "matches fsum on mixed magnitudes" {
    var prng = std.Random.DefaultPrng.init(42);
    const rng = prng.random();
    const scales = [_]f64{ 1e-8, 1.0, 1e8 };
    var data: [100]f64 = undefined;
    var k: usize = 0;
    while (k < 200) : (k += 1) {
        for (&data) |*d| {
            const u = -1.0 + 2.0 * rng.float(f64);
            d.* = u * scales[rng.uintLessThan(usize, scales.len)];
        }
        try testing.expectEqual(fsum(&data), kbnSum(&data));
    }
}

test "better accuracy than naive" {
    // Add and then subtract the same values, so the exact sum is 0.
    const count = 100000;
    var kbn = KleinKBNAccumulator{};
    var naive: f64 = 0.0;

    var prng = std.Random.DefaultPrng.init(42);
    const rng = prng.random();
    var i: usize = 0;
    while (i < count) : (i += 1) {
        const x = 1e7 * rng.float(f64);
        naive += x;
        kbn.update(x);
    }

    var prng2 = std.Random.DefaultPrng.init(42);
    const rng2 = prng2.random();
    i = 0;
    while (i < count) : (i += 1) {
        const x = 1e7 * rng2.float(f64);
        naive += -x;
        kbn.update(-x);
    }

    try testing.expectEqual(@as(f64, 0.0), kbn.value());
    try testing.expect(naive != 0.0);
}

test "update zero keeps compensation" {
    var kbn = KleinKBNAccumulator{};
    for (klein_data) |x| kbn.update(x);
    kbn.update(0.0);
    try testing.expectEqual(klein_expected, kbn.value());
}

test "revert" {
    var kbn = KleinKBNAccumulator{};
    kbn.update(1.5);
    kbn.update(2.5);
    kbn.revert(2.5);
    try testing.expectEqual(@as(f64, 1.5), kbn.value());
    kbn.revert(1.5);
    try testing.expectEqual(@as(f64, 0.0), kbn.value());
}

test "revert not most recent" {
    var kbn = KleinKBNAccumulator{};
    for (peters_data) |x| kbn.update(x);
    kbn.revert(1e100); // not the most recent value
    try testing.expectEqual(@as(f64, 2.0) - 1e100, kbn.value());
    kbn.revert(-1e100);
    try testing.expectEqual(@as(f64, 2.0), kbn.value());
}

test "revert restores compensated sum" {
    var kbn = KleinKBNAccumulator{};
    for (numpy_data) |x| kbn.update(x);
    kbn.update(1e20);
    kbn.revert(1e20);
    try testing.expect(almostEqual(kbn.value(), numpy_expected, 1e-16));
}

test "set" {
    var kbn = KleinKBNAccumulator{};
    for (peters_data) |x| kbn.update(x);
    kbn.set(5.0);
    try testing.expectEqual(@as(f64, 5.0), kbn.value());
    // Compensation terms are cleared, so only the new values count.
    kbn.update(1e100);
    kbn.update(-1e100);
    try testing.expectEqual(@as(f64, 5.0), kbn.value());
}

test "reset" {
    var kbn = KleinKBNAccumulator{};
    for (peters_data) |x| kbn.update(x);
    kbn.reset();
    try testing.expectEqual(@as(f64, 0.0), kbn.value());
    kbn.update(1.5);
    try testing.expectEqual(@as(f64, 1.5), kbn.value());
}
