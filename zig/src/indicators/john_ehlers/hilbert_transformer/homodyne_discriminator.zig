const std = @import("std");
const math = std.math;
const ht = @import("hilbert_transformer.zig");

/// Hilbert transformer cycle estimator using the Homodyne Discriminator technique.
///
/// John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 52-77.
pub const HomodyneDiscriminatorEstimator = struct {
    smoothing_length: usize,
    min_period: usize,
    max_period: usize,
    alpha_ema_quadrature_in_phase: f64,
    alpha_ema_period: f64,
    warm_up_period: usize,
    smoothing_length_plus_ht_length_min1: usize,
    smoothing_length_plus_2ht_length_min2: usize,
    smoothing_length_plus_3ht_length_min3: usize,
    smoothing_length_plus_3ht_length_min2: usize,
    smoothing_length_plus_3ht_length_min1: usize,
    smoothing_length_plus_3ht_length: usize,
    one_min_alpha_ema_quadrature_in_phase: f64,
    one_min_alpha_ema_period: f64,
    raw_values: [4]f64,
    wma_factors: [4]f64,
    wma_smoothed: [ht.ht_length]f64,
    detrend: [ht.ht_length]f64,
    in_phase: [ht.ht_length]f64,
    quadrt: [ht.ht_length]f64,
    j_in_phase: [ht.ht_length]f64,
    j_quadrature: [ht.ht_length]f64,
    _count: usize,
    smoothed_in_phase_previous: f64,
    smoothed_quadrature_previous: f64,
    re_previous: f64,
    im_previous: f64,
    _period: f64,
    is_primed: bool,
    is_warmed_up: bool,

    pub fn init(p: *const ht.CycleEstimatorParams) ht.VerifyError!HomodyneDiscriminatorEstimator {
        try ht.verifyParameters(p);

        const length = p.smoothing_length;
        const alpha_quad = p.alpha_ema_quadrature_in_phase;
        const alpha_period = p.alpha_ema_period;

        const sl_ht_1 = length + ht.ht_length - 1;
        const sl_2ht_2 = sl_ht_1 + ht.ht_length - 1;
        const sl_3ht_3 = sl_2ht_2 + ht.ht_length - 1;
        const sl_3ht_2 = sl_3ht_3 + 1;
        const sl_3ht_1 = sl_3ht_2 + 1;
        const sl_3ht = sl_3ht_1 + 1;

        var wma_factors: [4]f64 = .{ 0, 0, 0, 0 };
        ht.fillWmaFactors(length, &wma_factors);

        return .{
            .smoothing_length = length,
            .min_period = ht.default_min_period,
            .max_period = ht.default_max_period,
            .alpha_ema_quadrature_in_phase = alpha_quad,
            .alpha_ema_period = alpha_period,
            .warm_up_period = @max(p.warm_up_period, sl_3ht),
            .smoothing_length_plus_ht_length_min1 = sl_ht_1,
            .smoothing_length_plus_2ht_length_min2 = sl_2ht_2,
            .smoothing_length_plus_3ht_length_min3 = sl_3ht_3,
            .smoothing_length_plus_3ht_length_min2 = sl_3ht_2,
            .smoothing_length_plus_3ht_length_min1 = sl_3ht_1,
            .smoothing_length_plus_3ht_length = sl_3ht,
            .one_min_alpha_ema_quadrature_in_phase = 1.0 - alpha_quad,
            .one_min_alpha_ema_period = 1.0 - alpha_period,
            .raw_values = .{ 0, 0, 0, 0 },
            .wma_factors = wma_factors,
            .wma_smoothed = .{ 0, 0, 0, 0, 0, 0, 0 },
            .detrend = .{ 0, 0, 0, 0, 0, 0, 0 },
            .in_phase = .{ 0, 0, 0, 0, 0, 0, 0 },
            .quadrt = .{ 0, 0, 0, 0, 0, 0, 0 },
            .j_in_phase = .{ 0, 0, 0, 0, 0, 0, 0 },
            .j_quadrature = .{ 0, 0, 0, 0, 0, 0, 0 },
            ._count = 0,
            .smoothed_in_phase_previous = 0,
            .smoothed_quadrature_previous = 0,
            .re_previous = 0,
            .im_previous = 0,
            ._period = ht.default_min_period,
            .is_primed = false,
            .is_warmed_up = false,
        };
    }

    pub fn smoothingLength(self: *const HomodyneDiscriminatorEstimator) usize {
        return self.smoothing_length;
    }

    pub fn minPeriod(self: *const HomodyneDiscriminatorEstimator) usize {
        return self.min_period;
    }

    pub fn maxPeriod(self: *const HomodyneDiscriminatorEstimator) usize {
        return self.max_period;
    }

    pub fn warmUpPeriod(self: *const HomodyneDiscriminatorEstimator) usize {
        return self.warm_up_period;
    }

    pub fn alphaEmaQuadratureInPhase(self: *const HomodyneDiscriminatorEstimator) f64 {
        return self.alpha_ema_quadrature_in_phase;
    }

    pub fn alphaEmaPeriod(self: *const HomodyneDiscriminatorEstimator) f64 {
        return self.alpha_ema_period;
    }

    pub fn count(self: *const HomodyneDiscriminatorEstimator) usize {
        return self._count;
    }

    pub fn primed(self: *const HomodyneDiscriminatorEstimator) bool {
        return self.is_warmed_up;
    }

    pub fn period(self: *const HomodyneDiscriminatorEstimator) f64 {
        return self._period;
    }

    pub fn inPhase(self: *const HomodyneDiscriminatorEstimator) f64 {
        return self.in_phase[0];
    }

    pub fn quadrature(self: *const HomodyneDiscriminatorEstimator) f64 {
        return self.quadrt[0];
    }

    pub fn detrended(self: *const HomodyneDiscriminatorEstimator) f64 {
        return self.detrend[0];
    }

    pub fn smoothed(self: *const HomodyneDiscriminatorEstimator) f64 {
        return self.wma_smoothed[0];
    }

    pub fn update(self: *HomodyneDiscriminatorEstimator, sample: f64) void {
        if (math.isNan(sample)) return;

        const two_pi = 2.0 * math.pi;

        ht.push(4, &self.raw_values, sample);

        if (self.is_primed) {
            if (!self.is_warmed_up) {
                self._count += 1;
                if (self.warm_up_period < self._count) {
                    self.is_warmed_up = true;
                }
            }

            ht.push(ht.ht_length, &self.wma_smoothed, ht.wma(&self.raw_values, &self.wma_factors, self.smoothing_length));

            const acf = ht.correctAmplitude(self._period);

            ht.push(ht.ht_length, &self.detrend, ht.htTransform(&self.wma_smoothed) * acf);

            ht.push(ht.ht_length, &self.quadrt, ht.htTransform(&self.detrend) * acf);
            ht.push(ht.ht_length, &self.in_phase, self.detrend[ht.quadrature_index]);

            ht.push(ht.ht_length, &self.j_in_phase, ht.htTransform(&self.in_phase) * acf);
            ht.push(ht.ht_length, &self.j_quadrature, ht.htTransform(&self.quadrt) * acf);

            const smoothed_ip = self.emaQuad(self.in_phase[0] - self.j_quadrature[0], self.smoothed_in_phase_previous);
            const smoothed_q = self.emaQuad(self.quadrt[0] + self.j_in_phase[0], self.smoothed_quadrature_previous);

            var re = smoothed_ip * self.smoothed_in_phase_previous + smoothed_q * self.smoothed_quadrature_previous;
            var im = smoothed_ip * self.smoothed_quadrature_previous - smoothed_q * self.smoothed_in_phase_previous;
            self.smoothed_in_phase_previous = smoothed_ip;
            self.smoothed_quadrature_previous = smoothed_q;

            re = self.emaQuad(re, self.re_previous);
            im = self.emaQuad(im, self.im_previous);
            self.re_previous = re;
            self.im_previous = im;
            const period_previous = self._period;
            const period_new = two_pi / math.atan2(im, re);

            if (!math.isNan(period_new) and !math.isInf(period_new)) {
                self._period = period_new;
            }

            self._period = ht.adjustPeriod(self._period, period_previous);
            self._period = self.emaPeriodStep(self._period, period_previous);
        } else {
            // Not primed.
            self._count += 1;
            if (self.smoothing_length > self._count) return; // count < 4

            ht.push(ht.ht_length, &self.wma_smoothed, ht.wma(&self.raw_values, &self.wma_factors, self.smoothing_length)); // count >= 4

            if (self.smoothing_length_plus_ht_length_min1 > self._count) return; // count < 10

            const acf = ht.correctAmplitude(self._period); // count >= 10
            ht.push(ht.ht_length, &self.detrend, ht.htTransform(&self.wma_smoothed) * acf);

            if (self.smoothing_length_plus_2ht_length_min2 > self._count) return; // count < 16

            ht.push(ht.ht_length, &self.quadrt, ht.htTransform(&self.detrend) * acf); // count >= 16
            ht.push(ht.ht_length, &self.in_phase, self.detrend[ht.quadrature_index]);

            if (self.smoothing_length_plus_3ht_length_min3 > self._count) return; // count < 22

            ht.push(ht.ht_length, &self.j_in_phase, ht.htTransform(&self.in_phase) * acf); // count >= 22
            ht.push(ht.ht_length, &self.j_quadrature, ht.htTransform(&self.quadrt) * acf);

            if (self.smoothing_length_plus_3ht_length_min3 == self._count) { // count == 22
                self.smoothed_in_phase_previous = self.in_phase[0] - self.j_quadrature[0];
                self.smoothed_quadrature_previous = self.quadrt[0] + self.j_in_phase[0];
                return;
            }

            // count >= 23
            const smoothed_ip = self.emaQuad(self.in_phase[0] - self.j_quadrature[0], self.smoothed_in_phase_previous);
            const smoothed_q = self.emaQuad(self.quadrt[0] + self.j_in_phase[0], self.smoothed_quadrature_previous);

            const re = smoothed_ip * self.smoothed_in_phase_previous + smoothed_q * self.smoothed_quadrature_previous;
            const im_val = smoothed_ip * self.smoothed_quadrature_previous - smoothed_q * self.smoothed_in_phase_previous;
            self.smoothed_in_phase_previous = smoothed_ip;
            self.smoothed_quadrature_previous = smoothed_q;

            if (self.smoothing_length_plus_3ht_length_min2 == self._count) { // count == 23
                self.re_previous = re;
                self.im_previous = im_val;
                return;
            }

            // count >= 24
            const re_s = self.emaQuad(re, self.re_previous);
            const im_s = self.emaQuad(im_val, self.im_previous);
            self.re_previous = re_s;
            self.im_previous = im_s;
            const period_previous = self._period;

            const period_new = two_pi / math.atan2(im_s, re_s);
            if (!math.isNan(period_new) and !math.isInf(period_new)) {
                self._period = period_new;
            }

            self._period = ht.adjustPeriod(self._period, period_previous);

            if (self.smoothing_length_plus_3ht_length_min1 < self._count) { // count > 24
                self._period = self.emaPeriodStep(self._period, period_previous);
                self.is_primed = true;
            }
        }
    }

    fn emaQuad(self: *const HomodyneDiscriminatorEstimator, value: f64, previous: f64) f64 {
        return self.alpha_ema_quadrature_in_phase * value + self.one_min_alpha_ema_quadrature_in_phase * previous;
    }

    fn emaPeriodStep(self: *const HomodyneDiscriminatorEstimator, value: f64, previous: f64) f64 {
        return self.alpha_ema_period * value + self.one_min_alpha_ema_period * previous;
    }
};

// ============================================================================
// Tests
// ============================================================================

const testing = std.testing;

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

const testdata = @import("testdata.zig");

const testInput = testdata.testInput;
const testExpectedSmoothed = testdata.testExpectedSmoothed;
const testExpectedDetrended = testdata.testExpectedDetrended;
const testExpectedQuadrature = testdata.testExpectedQuadrature;
const testExpectedInPhase = testdata.testExpectedInPhase;
const testExpectedPeriod = testdata.testExpectedPeriod;

fn createDefault() HomodyneDiscriminatorEstimator {
    const params = ht.CycleEstimatorParams{
        .smoothing_length = 4,
        .alpha_ema_quadrature_in_phase = 0.2,
        .alpha_ema_period = 0.2,
    };
    return HomodyneDiscriminatorEstimator.init(&params) catch unreachable;
}

fn createWarmUp(warm_up: usize) HomodyneDiscriminatorEstimator {
    const params = ht.CycleEstimatorParams{
        .smoothing_length = 4,
        .alpha_ema_quadrature_in_phase = 0.2,
        .alpha_ema_period = 0.2,
        .warm_up_period = warm_up,
    };
    return HomodyneDiscriminatorEstimator.init(&params) catch unreachable;
}

test "reference implementation: wma smoothed" {
    var hde = createDefault();
    const input = testInput();
    const exp = testExpectedSmoothed();
    const epsilon = 1e-8;

    for (0..3) |i| {
        hde.update(input[i]);
        try testing.expect(almostEqual(0, hde.smoothed(), epsilon));
    }

    for (3..252) |i| {
        hde.update(input[i]);
        try testing.expect(almostEqual(exp[i], hde.smoothed(), epsilon));
    }

    const previous = hde.smoothed();
    hde.update(math.nan(f64));
    try testing.expect(almostEqual(previous, hde.smoothed(), epsilon));
}

test "reference implementation: detrended" {
    var hde = createDefault();
    const input = testInput();
    const exp = testExpectedDetrended();
    const epsilon = 1e-8;

    for (0..9) |i| {
        hde.update(input[i]);
        try testing.expect(almostEqual(0, hde.detrended(), epsilon));
    }

    for (9..252) |i| {
        hde.update(input[i]);
        try testing.expect(almostEqual(exp[i], hde.detrended(), epsilon));
    }

    const previous = hde.detrended();
    hde.update(math.nan(f64));
    try testing.expect(almostEqual(previous, hde.detrended(), epsilon));
}

test "reference implementation: quadrature" {
    var hde = createDefault();
    const input = testInput();
    const exp = testExpectedQuadrature();
    const epsilon = 1e-8;

    for (0..15) |i| {
        hde.update(input[i]);
        try testing.expect(almostEqual(0, hde.quadrature(), epsilon));
    }

    for (15..252) |i| {
        hde.update(input[i]);
        try testing.expect(almostEqual(exp[i], hde.quadrature(), epsilon));
    }

    const previous = hde.quadrature();
    hde.update(math.nan(f64));
    try testing.expect(almostEqual(previous, hde.quadrature(), epsilon));
}

test "reference implementation: in-phase" {
    var hde = createDefault();
    const input = testInput();
    const exp = testExpectedInPhase();
    const epsilon = 1e-8;

    for (0..15) |i| {
        hde.update(input[i]);
        try testing.expect(almostEqual(0, hde.inPhase(), epsilon));
    }

    for (15..252) |i| {
        hde.update(input[i]);
        try testing.expect(almostEqual(exp[i], hde.inPhase(), epsilon));
    }

    const previous = hde.inPhase();
    hde.update(math.nan(f64));
    try testing.expect(almostEqual(previous, hde.inPhase(), epsilon));
}

test "reference implementation: period" {
    var hde = createDefault();
    const input = testInput();
    const exp = testExpectedPeriod();
    const epsilon = 1e-8;

    for (0..23) |i| {
        hde.update(input[i]);
        try testing.expect(almostEqual(6, hde.period(), epsilon));
    }

    for (23..252) |i| {
        hde.update(input[i]);
        try testing.expect(almostEqual(exp[i], hde.period(), epsilon));
    }

    const previous = hde.period();
    hde.update(math.nan(f64));
    try testing.expect(almostEqual(previous, hde.period(), epsilon));
}

test "reference implementation: primed" {
    var hde = createDefault();
    const input = testInput();
    const lprimed = 4 + 7 * 3; // 25

    try testing.expect(!hde.primed());

    for (0..lprimed) |i| {
        hde.update(input[i]);
        try testing.expect(!hde.primed());
    }

    for (lprimed..252) |i| {
        hde.update(input[i]);
        try testing.expect(hde.primed());
    }
}

test "reference implementation: primed with warmup" {
    const lprimed = 50;
    var hde = createWarmUp(lprimed);
    const input = testInput();

    try testing.expect(!hde.primed());

    for (0..lprimed) |i| {
        hde.update(input[i]);
        try testing.expect(!hde.primed());
    }

    for (lprimed..252) |i| {
        hde.update(input[i]);
        try testing.expect(hde.primed());
    }
}

test "period of sin input" {
    const period_val: f64 = 30;
    const omega = 2.0 * math.pi / period_val;
    var hde = createDefault();
    for (0..512) |i| {
        hde.update(@sin(omega * @as(f64, @floatFromInt(i))));
    }
    try testing.expect(almostEqual(period_val, hde.period(), 1e-2));
}

test "min period of sin input" {
    const period_val: f64 = 3;
    const omega = 2.0 * math.pi / period_val;
    var hde = createDefault();
    for (0..512) |i| {
        hde.update(@sin(omega * @as(f64, @floatFromInt(i))));
    }
    try testing.expect(almostEqual(@as(f64, @floatFromInt(hde.minPeriod())), hde.period(), 1e-14));
}

test "max period of sin input" {
    const period_val: f64 = 60;
    const omega = 2.0 * math.pi / period_val;
    var hde = createDefault();
    for (0..512) |i| {
        hde.update(@sin(omega * @as(f64, @floatFromInt(i))));
    }
    try testing.expect(almostEqual(@as(f64, @floatFromInt(hde.maxPeriod())), hde.period(), 1e-14));
}

test "constructor validation: smoothing length = 1" {
    const params = ht.CycleEstimatorParams{ .smoothing_length = 1 };
    const result = HomodyneDiscriminatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidSmoothingLength, result);
}

test "constructor validation: smoothing length = 5" {
    const params = ht.CycleEstimatorParams{ .smoothing_length = 5 };
    const result = HomodyneDiscriminatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidSmoothingLength, result);
}

test "constructor validation: alpha quad = 0" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_quadrature_in_phase = 0.0 };
    const result = HomodyneDiscriminatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaQuadratureInPhase, result);
}

test "constructor validation: alpha quad = 1" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_quadrature_in_phase = 1.0 };
    const result = HomodyneDiscriminatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaQuadratureInPhase, result);
}

test "constructor validation: alpha period = 0" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_period = 0.0 };
    const result = HomodyneDiscriminatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaPeriod, result);
}

test "constructor validation: alpha period = 1" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_period = 1.0 };
    const result = HomodyneDiscriminatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaPeriod, result);
}
