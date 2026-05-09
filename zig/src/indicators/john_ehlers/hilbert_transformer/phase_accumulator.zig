const std = @import("std");
const math = std.math;
const ht = @import("hilbert_transformer.zig");

/// Phase Accumulator cycle estimator using the Hilbert transformer.
///
/// John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 63-66.
pub const PhaseAccumulatorEstimator = struct {
    smoothing_length: usize,
    min_period: usize,
    max_period: usize,
    alpha_ema_quadrature_in_phase: f64,
    alpha_ema_period: f64,
    warm_up_period: usize,
    smoothing_length_plus_ht_length_min1: usize,
    smoothing_length_plus_2ht_length_min2: usize,
    smoothing_length_plus_2ht_length_min1: usize,
    smoothing_length_plus_2ht_length: usize,
    one_min_alpha_ema_quadrature_in_phase: f64,
    one_min_alpha_ema_period: f64,
    raw_values: [4]f64,
    wma_factors: [4]f64,
    wma_smoothed: [ht.ht_length]f64,
    detrend: [ht.ht_length]f64,
    delta_phase: [ht.accumulation_length]f64,
    _in_phase: f64,
    _quadrature: f64,
    _count: usize,
    smoothed_in_phase_previous: f64,
    smoothed_quadrature_previous: f64,
    phase_previous: f64,
    _period: f64,
    is_primed: bool,
    is_warmed_up: bool,

    pub fn init(p: *const ht.CycleEstimatorParams) ht.VerifyError!PhaseAccumulatorEstimator {
        try ht.verifyParameters(p);

        const length = p.smoothing_length;
        const alpha_quad = p.alpha_ema_quadrature_in_phase;
        const alpha_period = p.alpha_ema_period;

        const sl_ht_1 = length + ht.ht_length - 1;
        const sl_2ht_2 = sl_ht_1 + ht.ht_length - 1;
        const sl_2ht_1 = sl_2ht_2 + 1;
        const sl_2ht = sl_2ht_1 + 1;

        var wma_factors: [4]f64 = .{ 0, 0, 0, 0 };
        ht.fillWmaFactors(length, &wma_factors);

        return .{
            .smoothing_length = length,
            .min_period = ht.default_min_period,
            .max_period = ht.default_max_period,
            .alpha_ema_quadrature_in_phase = alpha_quad,
            .alpha_ema_period = alpha_period,
            .warm_up_period = @max(p.warm_up_period, sl_2ht),
            .smoothing_length_plus_ht_length_min1 = sl_ht_1,
            .smoothing_length_plus_2ht_length_min2 = sl_2ht_2,
            .smoothing_length_plus_2ht_length_min1 = sl_2ht_1,
            .smoothing_length_plus_2ht_length = sl_2ht,
            .one_min_alpha_ema_quadrature_in_phase = 1.0 - alpha_quad,
            .one_min_alpha_ema_period = 1.0 - alpha_period,
            .raw_values = .{ 0, 0, 0, 0 },
            .wma_factors = wma_factors,
            .wma_smoothed = .{ 0, 0, 0, 0, 0, 0, 0 },
            .detrend = .{ 0, 0, 0, 0, 0, 0, 0 },
            .delta_phase = .{0} ** ht.accumulation_length,
            ._in_phase = 0,
            ._quadrature = 0,
            ._count = 0,
            .smoothed_in_phase_previous = 0,
            .smoothed_quadrature_previous = 0,
            .phase_previous = 0,
            ._period = ht.default_min_period,
            .is_primed = false,
            .is_warmed_up = false,
        };
    }

    pub fn update(self: *PhaseAccumulatorEstimator, sample: f64) void {
        if (math.isNan(sample)) return;

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

            // Compute both the in-phase and quadrature components of the detrended signal.
            self._quadrature = ht.htTransform(&self.detrend) * acf;
            self._in_phase = self.detrend[ht.quadrature_index];

            // Exponential moving average smoothing.
            const smoothed_ip = self.emaQuad(self._in_phase, self.smoothed_in_phase_previous);
            const smoothed_q = self.emaQuad(self._quadrature, self.smoothed_quadrature_previous);
            self.smoothed_in_phase_previous = smoothed_ip;
            self.smoothed_quadrature_previous = smoothed_q;

            // Compute an instantaneous phase.
            const phase_val = instantaneousPhase(smoothed_ip, smoothed_q, self.phase_previous);

            // Compute a differential phase.
            ht.push(ht.accumulation_length, &self.delta_phase, calculateDifferentialPhase(phase_val, self.phase_previous));
            self.phase_previous = phase_val;

            // Compute an instantaneous period.
            const period_previous = self._period;
            self._period = instantaneousPeriod(&self.delta_phase, period_previous);

            // Exponential moving average smoothing of the period.
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

            self._quadrature = ht.htTransform(&self.detrend) * acf; // count >= 16
            self._in_phase = self.detrend[ht.quadrature_index];

            if (self.smoothing_length_plus_2ht_length_min2 == self._count) { // count == 16
                self.smoothed_in_phase_previous = self._in_phase;
                self.smoothed_quadrature_previous = self._quadrature;
                return;
            }

            // count >= 17
            const smoothed_ip = self.emaQuad(self._in_phase, self.smoothed_in_phase_previous);
            const smoothed_q = self.emaQuad(self._quadrature, self.smoothed_quadrature_previous);
            self.smoothed_in_phase_previous = smoothed_ip;
            self.smoothed_quadrature_previous = smoothed_q;

            const phase_val = instantaneousPhase(smoothed_ip, smoothed_q, self.phase_previous);
            ht.push(ht.accumulation_length, &self.delta_phase, calculateDifferentialPhase(phase_val, self.phase_previous));
            self.phase_previous = phase_val;

            const period_previous = self._period;
            self._period = instantaneousPeriod(&self.delta_phase, period_previous);

            if (self.smoothing_length_plus_2ht_length_min1 < self._count) { // count >= 18
                self._period = self.emaPeriodStep(self._period, period_previous);
                self.is_primed = true;
            }
        }
    }

    fn emaQuad(self: *const PhaseAccumulatorEstimator, value: f64, previous: f64) f64 {
        return self.alpha_ema_quadrature_in_phase * value + self.one_min_alpha_ema_quadrature_in_phase * previous;
    }

    fn emaPeriodStep(self: *const PhaseAccumulatorEstimator, value: f64, previous: f64) f64 {
        return self.alpha_ema_period * value + self.one_min_alpha_ema_period * previous;
    }

    pub fn smoothingLength(self: *const PhaseAccumulatorEstimator) usize {
        return self.smoothing_length;
    }

    pub fn minPeriod(self: *const PhaseAccumulatorEstimator) usize {
        return self.min_period;
    }

    pub fn maxPeriod(self: *const PhaseAccumulatorEstimator) usize {
        return self.max_period;
    }

    pub fn warmUpPeriod(self: *const PhaseAccumulatorEstimator) usize {
        return self.warm_up_period;
    }

    pub fn alphaEmaQuadratureInPhase(self: *const PhaseAccumulatorEstimator) f64 {
        return self.alpha_ema_quadrature_in_phase;
    }

    pub fn alphaEmaPeriod(self: *const PhaseAccumulatorEstimator) f64 {
        return self.alpha_ema_period;
    }

    pub fn count(self: *const PhaseAccumulatorEstimator) usize {
        return self._count;
    }

    pub fn primed(self: *const PhaseAccumulatorEstimator) bool {
        return self.is_warmed_up;
    }

    pub fn period(self: *const PhaseAccumulatorEstimator) f64 {
        return self._period;
    }

    pub fn inPhase(self: *const PhaseAccumulatorEstimator) f64 {
        return self._in_phase;
    }

    pub fn quadrature(self: *const PhaseAccumulatorEstimator) f64 {
        return self._quadrature;
    }

    pub fn detrended(self: *const PhaseAccumulatorEstimator) f64 {
        return self.detrend[0];
    }

    pub fn smoothed(self: *const PhaseAccumulatorEstimator) f64 {
        return self.wma_smoothed[0];
    }
};

/// Computes instantaneous phase using arctangent with quadrant resolution.
fn instantaneousPhase(smoothed_in_phase: f64, smoothed_quadrature: f64, phase_previous: f64) f64 {
    const pi = math.pi;
    const two_pi = 2.0 * pi;

    // Use arctangent to compute the instantaneous phase in radians.
    const phase_raw = math.atan(@abs(smoothed_quadrature / smoothed_in_phase));
    if (math.isNan(phase_raw) or math.isInf(phase_raw)) {
        return phase_previous;
    }

    var phase_val = phase_raw;

    // Resolve the ambiguity for quadrants 2, 3, and 4.
    if (smoothed_in_phase < 0) {
        if (smoothed_quadrature > 0) {
            phase_val = pi - phase_val; // 2nd quadrant.
        } else if (smoothed_quadrature < 0) {
            phase_val = pi + phase_val; // 3rd quadrant.
        }
    } else if (smoothed_in_phase > 0 and smoothed_quadrature < 0) {
        phase_val = two_pi - phase_val; // 4th quadrant.
    }

    return phase_val;
}

/// Computes differential phase with wraparound fix and clamping.
fn calculateDifferentialPhase(phase_val: f64, phase_previous: f64) f64 {
    const two_pi = 2.0 * math.pi;
    const pi_over_2 = math.pi / 2.0;
    const three_pi_over_4 = 3.0 * math.pi / 4.0;
    const min_delta_phase = two_pi / ht.default_max_period;
    const max_delta_phase = two_pi / ht.default_min_period;

    // Compute a differential phase.
    var dp = phase_previous - phase_val;

    // Resolve phase wraparound from 1st quadrant to 4th quadrant.
    if (phase_previous < pi_over_2 and phase_val > three_pi_over_4) {
        dp += two_pi;
    }

    // Limit deltaPhase to be within [minDeltaPhase, maxDeltaPhase].
    if (dp < min_delta_phase) {
        dp = min_delta_phase;
    } else if (dp > max_delta_phase) {
        dp = max_delta_phase;
    }

    return dp;
}

/// Computes instantaneous period by accumulating delta phases until sum >= 2π.
fn instantaneousPeriod(delta_phase: *const [ht.accumulation_length]f64, period_previous: f64) f64 {
    const two_pi = 2.0 * math.pi;

    var sum_phase: f64 = 0;
    var period_val: usize = 0;

    for (0..ht.accumulation_length) |i| {
        sum_phase += delta_phase[i];
        if (sum_phase >= two_pi) {
            period_val = i + 1;
            break;
        }
    }

    // Resolve instantaneous period errors.
    if (period_val == 0) {
        return period_previous;
    }

    return @floatFromInt(period_val);
}

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

fn createDefault() PhaseAccumulatorEstimator {
    const params = ht.CycleEstimatorParams{
        .smoothing_length = 4,
        .alpha_ema_quadrature_in_phase = 0.15,
        .alpha_ema_period = 0.25,
    };
    return PhaseAccumulatorEstimator.init(&params) catch unreachable;
}

fn createWarmUp(warm_up: usize) PhaseAccumulatorEstimator {
    const params = ht.CycleEstimatorParams{
        .smoothing_length = 4,
        .alpha_ema_quadrature_in_phase = 0.15,
        .alpha_ema_period = 0.25,
        .warm_up_period = warm_up,
    };
    return PhaseAccumulatorEstimator.init(&params) catch unreachable;
}

test "reference implementation: wma smoothed" {
    var pae = createDefault();
    const input = testInput();
    const exp = testExpectedSmoothed();
    const epsilon = 1e-8;

    for (0..3) |i| {
        pae.update(input[i]);
        try testing.expect(almostEqual(0, pae.smoothed(), epsilon));
    }

    for (3..252) |i| {
        pae.update(input[i]);
        try testing.expect(almostEqual(exp[i], pae.smoothed(), epsilon));
    }

    const previous = pae.smoothed();
    pae.update(math.nan(f64));
    try testing.expect(almostEqual(previous, pae.smoothed(), epsilon));
}

test "reference implementation: detrended" {
    var pae = createDefault();
    const input = testInput();
    const exp = testExpectedDetrended();
    const epsilon = 1e-8;
    const last = 24;

    for (0..9) |i| {
        pae.update(input[i]);
        try testing.expect(almostEqual(0, pae.detrended(), epsilon));
    }

    for (9..last) |i| {
        pae.update(input[i]);
        try testing.expect(almostEqual(exp[i], pae.detrended(), epsilon));
    }

    const previous = pae.detrended();
    pae.update(math.nan(f64));
    try testing.expect(almostEqual(previous, pae.detrended(), epsilon));
}

test "reference implementation: quadrature" {
    var pae = createDefault();
    const input = testInput();
    const exp = testExpectedQuadrature();
    const epsilon = 1e-8;
    const last = 24;

    for (0..15) |i| {
        pae.update(input[i]);
        try testing.expect(almostEqual(0, pae.quadrature(), epsilon));
    }

    for (15..last) |i| {
        pae.update(input[i]);
        try testing.expect(almostEqual(exp[i], pae.quadrature(), epsilon));
    }

    const previous = pae.quadrature();
    pae.update(math.nan(f64));
    try testing.expect(almostEqual(previous, pae.quadrature(), epsilon));
}

test "reference implementation: in-phase" {
    var pae = createDefault();
    const input = testInput();
    const exp = testExpectedInPhase();
    const epsilon = 1e-8;
    const last = 24;

    for (0..15) |i| {
        pae.update(input[i]);
        try testing.expect(almostEqual(0, pae.inPhase(), epsilon));
    }

    for (15..last) |i| {
        pae.update(input[i]);
        try testing.expect(almostEqual(exp[i], pae.inPhase(), epsilon));
    }

    const previous = pae.inPhase();
    pae.update(math.nan(f64));
    try testing.expect(almostEqual(previous, pae.inPhase(), epsilon));
}

test "reference implementation: period" {
    var pae = createDefault();
    const input = testInput();
    // Period test only checks up to index 18 for PA.
    const epsilon = 1e-8;
    const last = 18;

    for (0..18) |i| {
        pae.update(input[i]);
        try testing.expect(almostEqual(6, pae.period(), epsilon));
    }

    for (18..last) |_| {
        // No iterations since last == 18.
    }

    const previous = pae.period();
    pae.update(math.nan(f64));
    try testing.expect(almostEqual(previous, pae.period(), epsilon));
}

test "reference implementation: primed" {
    var pae = createDefault();
    const input = testInput();
    const lprimed = 4 + 7 * 2; // 18

    try testing.expect(!pae.primed());

    for (0..lprimed) |i| {
        pae.update(input[i]);
        try testing.expect(!pae.primed());
    }

    for (lprimed..252) |i| {
        pae.update(input[i]);
        try testing.expect(pae.primed());
    }
}

test "reference implementation: primed with warmup" {
    const lprimed = 50;
    var pae = createWarmUp(lprimed);
    const input = testInput();

    try testing.expect(!pae.primed());

    for (0..lprimed) |i| {
        pae.update(input[i]);
        try testing.expect(!pae.primed());
    }

    for (lprimed..252) |i| {
        pae.update(input[i]);
        try testing.expect(pae.primed());
    }
}

test "period of sin input" {
    const period_val: f64 = 30;
    const omega = 2.0 * math.pi / period_val;
    var pae = createDefault();
    for (0..512) |i| {
        pae.update(@sin(omega * @as(f64, @floatFromInt(i))));
    }
    try testing.expect(almostEqual(period_val, pae.period(), 1e0));
}

test "min period of sin input" {
    const period_val: f64 = 3;
    const omega = 2.0 * math.pi / period_val;
    var pae = createDefault();
    for (0..512) |i| {
        pae.update(@sin(omega * @as(f64, @floatFromInt(i))));
    }
    try testing.expect(almostEqual(@as(f64, @floatFromInt(pae.minPeriod())), pae.period(), 1e-14));
}

test "max period of sin input" {
    const period_val: f64 = 60;
    const omega = 2.0 * math.pi / period_val;
    var pae = createDefault();
    for (0..512) |i| {
        pae.update(@sin(omega * @as(f64, @floatFromInt(i))));
    }
    try testing.expect(almostEqual(@as(f64, @floatFromInt(pae.maxPeriod())), pae.period(), 12.5e0));
}

test "constructor validation: smoothing length = 1" {
    const params = ht.CycleEstimatorParams{ .smoothing_length = 1 };
    const result = PhaseAccumulatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidSmoothingLength, result);
}

test "constructor validation: smoothing length = 5" {
    const params = ht.CycleEstimatorParams{ .smoothing_length = 5 };
    const result = PhaseAccumulatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidSmoothingLength, result);
}

test "constructor validation: alpha quad = 0" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_quadrature_in_phase = 0.0 };
    const result = PhaseAccumulatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaQuadratureInPhase, result);
}

test "constructor validation: alpha quad = 1" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_quadrature_in_phase = 1.0 };
    const result = PhaseAccumulatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaQuadratureInPhase, result);
}

test "constructor validation: alpha period = 0" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_period = 0.0 };
    const result = PhaseAccumulatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaPeriod, result);
}

test "constructor validation: alpha period = 1" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_period = 1.0 };
    const result = PhaseAccumulatorEstimator.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaPeriod, result);
}
