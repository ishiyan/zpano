const std = @import("std");
const math = std.math;
const testing = std.testing;

/// Klein second-order Kahan-Babuška-Neumaier (KBN) floating-point compensated
/// summation.
///
/// Maintains sum + cs + ccs where sum is the primary sum, cs is the
/// first-level KBN correction, and ccs is a second-level KBN correction
/// applied to the first correction term (Klein's generalisation).
///
/// Unlike naive summation, KBN correctly sums sequences with extreme
/// magnitude differences (e.g. Peters' example [1.0, 1e100, 1.0, -1e100]
/// → 2.0, while naive and standard Kahan return 0.0).
///
/// Use set(x) to overwrite the accumulator value (resets both
/// compensation terms to zero). Prefer set over constructing a
/// new instance when the accumulator is stored in an object slot.
///
/// Level 1 (Kahan-Babuška-Neumaier):
///   t = sum + x
///   if |sum| >= |x|:  c = (sum - t) + x
///   else:             c = (x - t) + sum
///   sum = t
///
/// Level 2 (Klein generalisation): reapplies the same technique
/// to the correction term c itself.
pub const KleinKBNAccumulator = struct {
    sum: f64 = 0.0,
    cs: f64 = 0.0,
    ccs: f64 = 0.0,

    /// Overwrites the accumulator value and resets both compensation terms to zero.
    pub fn set(self: *KleinKBNAccumulator, x: f64) void {
        self.sum = x;
        self.cs = 0.0;
        self.ccs = 0.0;
    }

    /// Resets the accumulator to zero.
    pub fn reset(self: *KleinKBNAccumulator) void {
        self.set(0.0);
    }

    /// Removes x from the accumulator by adding -x.
    pub fn revert(self: *KleinKBNAccumulator, x: f64) void {
        self.update(-x);
    }

    /// Adds x using Klein second-order KBN compensated summation.
    pub fn update(self: *KleinKBNAccumulator, x: f64) void {
        var t = self.sum + x;
        var c: f64 = undefined;
        if (@abs(self.sum) >= @abs(x)) {
            c = (self.sum - t) + x;
        } else {
            c = (x - t) + self.sum;
        }
        self.sum = t;

        t = self.cs + c;
        var cc: f64 = undefined;
        if (@abs(self.cs) >= @abs(c)) {
            cc = (self.cs - t) + c;
        } else {
            cc = (c - t) + self.cs;
        }
        self.cs = t;
        self.ccs = cc;
    }

    /// Returns the current compensated sum: sum + cs + ccs.
    pub fn value(self: *const KleinKBNAccumulator) f64 {
        return self.sum + self.cs + self.ccs;
    }
};

// ── Tests ──────────────────────────────────────────────────────────────────

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

test "peters" {
    const data = [_]f64{ 1.0, 1e100, 1.0, -1e100 };
    var kbn = KleinKBNAccumulator{};
    for (data) |x| {
        kbn.update(x);
    }
    try testing.expect(almostEqual(kbn.value(), 2.0, 1e-15));
}

test "numpy" {
    const data = [_]f64{
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
    const expected = -0.377392919181026;
    var kbn = KleinKBNAccumulator{};
    for (data) |x| {
        kbn.update(x);
    }
    try testing.expect(almostEqual(kbn.value(), expected, 1e-16));
}

test "better accuracy than naive" {
    const spread = 1e7;
    var kbn = KleinKBNAccumulator{};
    var naive: f64 = 0.0;

    var prng = std.Random.DefaultPrng.init(42);
    const rand = prng.random();
    var i: usize = 0;
    while (i < 1000000) : (i += 1) {
        const x = rand.float(f64) * spread;
        naive += x;
        kbn.update(x);
    }

    var prng2 = std.Random.DefaultPrng.init(42);
    const rand2 = prng2.random();
    i = 0;
    while (i < 1000000) : (i += 1) {
        const x = rand2.float(f64) * spread;
        naive -= x;
        kbn.update(-x);
    }

    try testing.expect(@abs(kbn.value()) <= @abs(naive));
}

test "revert" {
    var kbn = KleinKBNAccumulator{};
    try testing.expect(almostEqual(kbn.value(), 0.0, 1e-15));

    kbn.update(1.5);
    kbn.update(2.5);
    kbn.revert(2.5);
    try testing.expect(almostEqual(kbn.value(), 1.5, 1e-15));
    kbn.revert(1.5);
    try testing.expect(almostEqual(kbn.value(), 0.0, 1e-15));
}

test "reset" {
    var kbn = KleinKBNAccumulator{};
    kbn.update(1.5);
    kbn.reset();
    try testing.expect(almostEqual(kbn.value(), 0.0, 1e-15));

    kbn.update(1.5);
    try testing.expect(almostEqual(kbn.value(), 1.5, 1e-15));
}
