const std = @import("std");
const math = std.math;
const ht = @import("hilbert_transformer.zig");

/// Hilbert transformer cycle estimator using the Homodyne Discriminator technique (unrolled).
///
/// Copied from the TA-Lib implementation with unrolled loops.
///
/// John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 52-77.
pub const HomodyneDiscriminatorEstimatorUnrolled = struct {
    smoothing_length: usize,
    min_period: usize,
    max_period: usize,
    alpha_ema_quadrature_in_phase: f64,
    alpha_ema_period: f64,
    warm_up_period: usize,
    one_min_alpha_ema_quadrature_in_phase: f64,
    one_min_alpha_ema_period: f64,
    _smoothed: f64,
    _detrended: f64,
    _in_phase: f64,
    _quadrature: f64,
    smoothing_multiplier: f64,
    adjusted_period: f64,
    _count: usize,
    index: usize,
    i2_previous: f64,
    q2_previous: f64,
    re: f64,
    im: f64,
    _period: f64,
    is_primed: bool,
    is_warmed_up: bool,

    // WMA smoother.
    wma_sum: f64,
    wma_sub: f64,
    wma_input1: f64,
    wma_input2: f64,
    wma_input3: f64,
    wma_input4: f64,

    // Detrender.
    detrender_odd0: f64,
    detrender_odd1: f64,
    detrender_odd2: f64,
    detrender_previous_odd: f64,
    detrender_previous_input_odd: f64,
    detrender_even0: f64,
    detrender_even1: f64,
    detrender_even2: f64,
    detrender_previous_even: f64,
    detrender_previous_input_even: f64,

    // Q1.
    q1_odd0: f64,
    q1_odd1: f64,
    q1_odd2: f64,
    q1_previous_odd: f64,
    q1_previous_input_odd: f64,
    q1_even0: f64,
    q1_even1: f64,
    q1_even2: f64,
    q1_previous_even: f64,
    q1_previous_input_even: f64,

    // I1.
    i1_previous1_odd: f64,
    i1_previous2_odd: f64,
    i1_previous1_even: f64,
    i1_previous2_even: f64,

    // jI.
    ji_odd0: f64,
    ji_odd1: f64,
    ji_odd2: f64,
    ji_previous_odd: f64,
    ji_previous_input_odd: f64,
    ji_even0: f64,
    ji_even1: f64,
    ji_even2: f64,
    ji_previous_even: f64,
    ji_previous_input_even: f64,

    // jQ.
    jq_odd0: f64,
    jq_odd1: f64,
    jq_odd2: f64,
    jq_previous_odd: f64,
    jq_previous_input_odd: f64,
    jq_even0: f64,
    jq_even1: f64,
    jq_even2: f64,
    jq_previous_even: f64,
    jq_previous_input_even: f64,

    const primed_count: usize = 23;

    pub fn init(p: *const ht.CycleEstimatorParams) ht.VerifyError!HomodyneDiscriminatorEstimatorUnrolled {
        try ht.verifyParameters(p);

        const length = p.smoothing_length;
        const alpha_quad = p.alpha_ema_quadrature_in_phase;
        const alpha_period = p.alpha_ema_period;

        const sm: f64 = if (length == 4) 1.0 / 10.0 else if (length == 3) 1.0 / 6.0 else 1.0 / 3.0;

        return .{
            .smoothing_length = length,
            .min_period = ht.default_min_period,
            .max_period = ht.default_max_period,
            .alpha_ema_quadrature_in_phase = alpha_quad,
            .alpha_ema_period = alpha_period,
            .warm_up_period = @max(p.warm_up_period, primed_count),
            .one_min_alpha_ema_quadrature_in_phase = 1.0 - alpha_quad,
            .one_min_alpha_ema_period = 1.0 - alpha_period,
            ._smoothed = 0,
            ._detrended = 0,
            ._in_phase = 0,
            ._quadrature = 0,
            .smoothing_multiplier = sm,
            .adjusted_period = 0,
            ._count = 0,
            .index = 0,
            .i2_previous = 0,
            .q2_previous = 0,
            .re = 0,
            .im = 0,
            ._period = ht.default_min_period,
            .is_primed = false,
            .is_warmed_up = false,
            .wma_sum = 0,
            .wma_sub = 0,
            .wma_input1 = 0,
            .wma_input2 = 0,
            .wma_input3 = 0,
            .wma_input4 = 0,
            .detrender_odd0 = 0,
            .detrender_odd1 = 0,
            .detrender_odd2 = 0,
            .detrender_previous_odd = 0,
            .detrender_previous_input_odd = 0,
            .detrender_even0 = 0,
            .detrender_even1 = 0,
            .detrender_even2 = 0,
            .detrender_previous_even = 0,
            .detrender_previous_input_even = 0,
            .q1_odd0 = 0,
            .q1_odd1 = 0,
            .q1_odd2 = 0,
            .q1_previous_odd = 0,
            .q1_previous_input_odd = 0,
            .q1_even0 = 0,
            .q1_even1 = 0,
            .q1_even2 = 0,
            .q1_previous_even = 0,
            .q1_previous_input_even = 0,
            .i1_previous1_odd = 0,
            .i1_previous2_odd = 0,
            .i1_previous1_even = 0,
            .i1_previous2_even = 0,
            .ji_odd0 = 0,
            .ji_odd1 = 0,
            .ji_odd2 = 0,
            .ji_previous_odd = 0,
            .ji_previous_input_odd = 0,
            .ji_even0 = 0,
            .ji_even1 = 0,
            .ji_even2 = 0,
            .ji_previous_even = 0,
            .ji_previous_input_even = 0,
            .jq_odd0 = 0,
            .jq_odd1 = 0,
            .jq_odd2 = 0,
            .jq_previous_odd = 0,
            .jq_previous_input_odd = 0,
            .jq_even0 = 0,
            .jq_even1 = 0,
            .jq_even2 = 0,
            .jq_previous_even = 0,
            .jq_previous_input_even = 0,
        };
    }

    pub fn smoothingLength(self: *const HomodyneDiscriminatorEstimatorUnrolled) usize {
        return self.smoothing_length;
    }

    pub fn minPeriod(self: *const HomodyneDiscriminatorEstimatorUnrolled) usize {
        return self.min_period;
    }

    pub fn maxPeriod(self: *const HomodyneDiscriminatorEstimatorUnrolled) usize {
        return self.max_period;
    }

    pub fn warmUpPeriod(self: *const HomodyneDiscriminatorEstimatorUnrolled) usize {
        return self.warm_up_period;
    }

    pub fn alphaEmaQuadratureInPhase(self: *const HomodyneDiscriminatorEstimatorUnrolled) f64 {
        return self.alpha_ema_quadrature_in_phase;
    }

    pub fn alphaEmaPeriod(self: *const HomodyneDiscriminatorEstimatorUnrolled) f64 {
        return self.alpha_ema_period;
    }

    pub fn count(self: *const HomodyneDiscriminatorEstimatorUnrolled) usize {
        return self._count;
    }

    pub fn primed(self: *const HomodyneDiscriminatorEstimatorUnrolled) bool {
        return self.is_warmed_up;
    }

    pub fn period(self: *const HomodyneDiscriminatorEstimatorUnrolled) f64 {
        return self._period;
    }

    pub fn inPhase(self: *const HomodyneDiscriminatorEstimatorUnrolled) f64 {
        return self._in_phase;
    }

    pub fn quadrature(self: *const HomodyneDiscriminatorEstimatorUnrolled) f64 {
        return self._quadrature;
    }

    pub fn detrended(self: *const HomodyneDiscriminatorEstimatorUnrolled) f64 {
        return self._detrended;
    }

    pub fn smoothed(self: *const HomodyneDiscriminatorEstimatorUnrolled) f64 {
        return self._smoothed;
    }

    pub fn update(self: *HomodyneDiscriminatorEstimatorUnrolled, sample: f64) void {
        if (math.isNan(sample)) return;

        const a = 0.0962;
        const b = 0.5769;

        var value: f64 = undefined;

        self._count += 1;

        // WMA accumulation phase.
        if (self.smoothing_length >= self._count) {
            if (1 == self._count) {
                self.wma_sub = sample;
                self.wma_input1 = sample;
                self.wma_sum = sample;
            } else if (2 == self._count) {
                self.wma_sub += sample;
                self.wma_input2 = sample;
                self.wma_sum += sample * 2;
                if (2 == self.smoothing_length) {
                    value = self.wma_sum * self.smoothing_multiplier;
                    // fall through to detrend
                } else return;
            } else if (3 == self._count) {
                self.wma_sub += sample;
                self.wma_input3 = sample;
                self.wma_sum += sample * 3;
                if (3 == self.smoothing_length) {
                    value = self.wma_sum * self.smoothing_multiplier;
                    // fall through to detrend
                } else return;
            } else { // 4 == self._count
                self.wma_sub += sample;
                self.wma_input4 = sample;
                self.wma_sum += sample * 4;
                value = self.wma_sum * self.smoothing_multiplier;
                // fall through to detrend
            }
        } else {
            // Normal WMA computation.
            self.wma_sum -= self.wma_sub;
            self.wma_sum += sample * @as(f64, @floatFromInt(self.smoothing_length));
            value = self.wma_sum * self.smoothing_multiplier;
            self.wma_sub += sample;
            self.wma_sub -= self.wma_input1;
            self.wma_input1 = self.wma_input2;

            if (4 == self.smoothing_length) {
                self.wma_input2 = self.wma_input3;
                self.wma_input3 = self.wma_input4;
                self.wma_input4 = sample;
            } else if (3 == self.smoothing_length) {
                self.wma_input2 = self.wma_input3;
                self.wma_input3 = sample;
            } else {
                self.wma_input2 = sample;
            }
        }

        // Detrend label.
        self._smoothed = value;

        if (!self.is_warmed_up) {
            self.is_warmed_up = self._count > self.warm_up_period;
            if (!self.is_primed) {
                self.is_primed = self._count > primed_count;
            }
        }

        var detrender: f64 = undefined;
        var ji: f64 = undefined;
        var jq: f64 = undefined;

        var temp = a * self._smoothed;
        self.adjusted_period = 0.075 * self._period + 0.54;

        // Even value count.
        if (0 == self._count % 2) {
            if (0 == self.index) {
                self.index = 1;
                detrender = -self.detrender_even0;
                self.detrender_even0 = temp;
                detrender += temp;
                detrender -= self.detrender_previous_even;
                self.detrender_previous_even = b * self.detrender_previous_input_even;
                self.detrender_previous_input_even = value;
                detrender += self.detrender_previous_even;
                detrender *= self.adjusted_period;

                temp = a * detrender;
                self._quadrature = -self.q1_even0;
                self.q1_even0 = temp;
                self._quadrature += temp;
                self._quadrature -= self.q1_previous_even;
                self.q1_previous_even = b * self.q1_previous_input_even;
                self.q1_previous_input_even = detrender;
                self._quadrature += self.q1_previous_even;
                self._quadrature *= self.adjusted_period;

                temp = a * self.i1_previous2_even;
                ji = -self.ji_even0;
                self.ji_even0 = temp;
                ji += temp;
                ji -= self.ji_previous_even;
                self.ji_previous_even = b * self.ji_previous_input_even;
                self.ji_previous_input_even = self.i1_previous2_even;
                ji += self.ji_previous_even;
                ji *= self.adjusted_period;

                temp = a * self._quadrature;
                jq = -self.jq_even0;
                self.jq_even0 = temp;
            } else if (1 == self.index) {
                self.index = 2;
                detrender = -self.detrender_even1;
                self.detrender_even1 = temp;
                detrender += temp;
                detrender -= self.detrender_previous_even;
                self.detrender_previous_even = b * self.detrender_previous_input_even;
                self.detrender_previous_input_even = value;
                detrender += self.detrender_previous_even;
                detrender *= self.adjusted_period;

                temp = a * detrender;
                self._quadrature = -self.q1_even1;
                self.q1_even1 = temp;
                self._quadrature += temp;
                self._quadrature -= self.q1_previous_even;
                self.q1_previous_even = b * self.q1_previous_input_even;
                self.q1_previous_input_even = detrender;
                self._quadrature += self.q1_previous_even;
                self._quadrature *= self.adjusted_period;

                temp = a * self.i1_previous2_even;
                ji = -self.ji_even1;
                self.ji_even1 = temp;
                ji += temp;
                ji -= self.ji_previous_even;
                self.ji_previous_even = b * self.ji_previous_input_even;
                self.ji_previous_input_even = self.i1_previous2_even;
                ji += self.ji_previous_even;
                ji *= self.adjusted_period;

                temp = a * self._quadrature;
                jq = -self.jq_even1;
                self.jq_even1 = temp;
            } else { // 2 == self.index
                self.index = 0;
                detrender = -self.detrender_even2;
                self.detrender_even2 = temp;
                detrender += temp;
                detrender -= self.detrender_previous_even;
                self.detrender_previous_even = b * self.detrender_previous_input_even;
                self.detrender_previous_input_even = value;
                detrender += self.detrender_previous_even;
                detrender *= self.adjusted_period;

                temp = a * detrender;
                self._quadrature = -self.q1_even2;
                self.q1_even2 = temp;
                self._quadrature += temp;
                self._quadrature -= self.q1_previous_even;
                self.q1_previous_even = b * self.q1_previous_input_even;
                self.q1_previous_input_even = detrender;
                self._quadrature += self.q1_previous_even;
                self._quadrature *= self.adjusted_period;

                temp = a * self.i1_previous2_even;
                ji = -self.ji_even2;
                self.ji_even2 = temp;
                ji += temp;
                ji -= self.ji_previous_even;
                self.ji_previous_even = b * self.ji_previous_input_even;
                self.ji_previous_input_even = self.i1_previous2_even;
                ji += self.ji_previous_even;
                ji *= self.adjusted_period;

                temp = a * self._quadrature;
                jq = -self.jq_even2;
                self.jq_even2 = temp;
            }

            // jQ continued.
            jq += temp;
            jq -= self.jq_previous_even;
            self.jq_previous_even = b * self.jq_previous_input_even;
            self.jq_previous_input_even = self._quadrature;
            jq += self.jq_previous_even;
            jq *= self.adjusted_period;

            // InPhase.
            self._in_phase = self.i1_previous2_even;

            // Update odd i1 history.
            self.i1_previous2_odd = self.i1_previous1_odd;
            self.i1_previous1_odd = detrender;
        } else {
            // Odd value count.
            if (0 == self.index) {
                self.index = 1;
                detrender = -self.detrender_odd0;
                self.detrender_odd0 = temp;
                detrender += temp;
                detrender -= self.detrender_previous_odd;
                self.detrender_previous_odd = b * self.detrender_previous_input_odd;
                self.detrender_previous_input_odd = value;
                detrender += self.detrender_previous_odd;
                detrender *= self.adjusted_period;

                temp = a * detrender;
                self._quadrature = -self.q1_odd0;
                self.q1_odd0 = temp;
                self._quadrature += temp;
                self._quadrature -= self.q1_previous_odd;
                self.q1_previous_odd = b * self.q1_previous_input_odd;
                self.q1_previous_input_odd = detrender;
                self._quadrature += self.q1_previous_odd;
                self._quadrature *= self.adjusted_period;

                temp = a * self.i1_previous2_odd;
                ji = -self.ji_odd0;
                self.ji_odd0 = temp;
                ji += temp;
                ji -= self.ji_previous_odd;
                self.ji_previous_odd = b * self.ji_previous_input_odd;
                self.ji_previous_input_odd = self.i1_previous2_odd;
                ji += self.ji_previous_odd;
                ji *= self.adjusted_period;

                temp = a * self._quadrature;
                jq = -self.jq_odd0;
                self.jq_odd0 = temp;
            } else if (1 == self.index) {
                self.index = 2;
                detrender = -self.detrender_odd1;
                self.detrender_odd1 = temp;
                detrender += temp;
                detrender -= self.detrender_previous_odd;
                self.detrender_previous_odd = b * self.detrender_previous_input_odd;
                self.detrender_previous_input_odd = value;
                detrender += self.detrender_previous_odd;
                detrender *= self.adjusted_period;

                temp = a * detrender;
                self._quadrature = -self.q1_odd1;
                self.q1_odd1 = temp;
                self._quadrature += temp;
                self._quadrature -= self.q1_previous_odd;
                self.q1_previous_odd = b * self.q1_previous_input_odd;
                self.q1_previous_input_odd = detrender;
                self._quadrature += self.q1_previous_odd;
                self._quadrature *= self.adjusted_period;

                temp = a * self.i1_previous2_odd;
                ji = -self.ji_odd1;
                self.ji_odd1 = temp;
                ji += temp;
                ji -= self.ji_previous_odd;
                self.ji_previous_odd = b * self.ji_previous_input_odd;
                self.ji_previous_input_odd = self.i1_previous2_odd;
                ji += self.ji_previous_odd;
                ji *= self.adjusted_period;

                temp = a * self._quadrature;
                jq = -self.jq_odd1;
                self.jq_odd1 = temp;
            } else { // 2 == self.index
                self.index = 0;
                detrender = -self.detrender_odd2;
                self.detrender_odd2 = temp;
                detrender += temp;
                detrender -= self.detrender_previous_odd;
                self.detrender_previous_odd = b * self.detrender_previous_input_odd;
                self.detrender_previous_input_odd = value;
                detrender += self.detrender_previous_odd;
                detrender *= self.adjusted_period;

                temp = a * detrender;
                self._quadrature = -self.q1_odd2;
                self.q1_odd2 = temp;
                self._quadrature += temp;
                self._quadrature -= self.q1_previous_odd;
                self.q1_previous_odd = b * self.q1_previous_input_odd;
                self.q1_previous_input_odd = detrender;
                self._quadrature += self.q1_previous_odd;
                self._quadrature *= self.adjusted_period;

                temp = a * self.i1_previous2_odd;
                ji = -self.ji_odd2;
                self.ji_odd2 = temp;
                ji += temp;
                ji -= self.ji_previous_odd;
                self.ji_previous_odd = b * self.ji_previous_input_odd;
                self.ji_previous_input_odd = self.i1_previous2_odd;
                ji += self.ji_previous_odd;
                ji *= self.adjusted_period;

                temp = a * self._quadrature;
                jq = -self.jq_odd2;
                self.jq_odd2 = temp;
            }

            // jQ continued.
            jq += temp;
            jq -= self.jq_previous_odd;
            self.jq_previous_odd = b * self.jq_previous_input_odd;
            self.jq_previous_input_odd = self._quadrature;
            jq += self.jq_previous_odd;
            jq *= self.adjusted_period;

            // InPhase.
            self._in_phase = self.i1_previous2_odd;

            // Update even i1 history.
            self.i1_previous2_even = self.i1_previous1_even;
            self.i1_previous1_even = detrender;
        }

        self._detrended = detrender;

        // Phasor addition for 3 bar averaging.
        var smoothed_i2 = self._in_phase - jq;
        var smoothed_q2 = self._quadrature + ji;

        // Smooth the InPhase and Quadrature components.
        smoothed_i2 = self.alpha_ema_quadrature_in_phase * smoothed_i2 + self.one_min_alpha_ema_quadrature_in_phase * self.i2_previous;
        smoothed_q2 = self.alpha_ema_quadrature_in_phase * smoothed_q2 + self.one_min_alpha_ema_quadrature_in_phase * self.q2_previous;

        // Homodyne discriminator.
        self.re = self.alpha_ema_quadrature_in_phase * (smoothed_i2 * self.i2_previous + smoothed_q2 * self.q2_previous) + self.one_min_alpha_ema_quadrature_in_phase * self.re;
        self.im = self.alpha_ema_quadrature_in_phase * (smoothed_i2 * self.q2_previous - smoothed_q2 * self.i2_previous) + self.one_min_alpha_ema_quadrature_in_phase * self.im;
        self.q2_previous = smoothed_q2;
        self.i2_previous = smoothed_i2;
        const period_previous = self._period;

        const period_new = 2.0 * math.pi / math.atan2(self.im, self.re);
        if (!math.isNan(period_new) and !math.isInf(period_new)) {
            self._period = period_new;
        }

        self._period = ht.adjustPeriod(self._period, period_previous);
        self._period = self.alpha_ema_period * self._period + self.one_min_alpha_ema_period * period_previous;
    }
};

// ============================================================================
// Tests
// ============================================================================

const testing = std.testing;

fn almostEqual(a_val: f64, b_val: f64, epsilon: f64) bool {
    return @abs(a_val - b_val) <= epsilon;
}

const testdata = @import("testdata_unrolled.zig");

const testInput = testdata.testInput;
const testExpectedSmoothed = testdata.testExpectedSmoothed;
const testExpectedDetrended = testdata.testExpectedDetrended;
const testExpectedQuadrature = testdata.testExpectedQuadrature;
const testExpectedInPhase = testdata.testExpectedInPhase;
const testExpectedPeriod = testdata.testExpectedPeriod;

fn createDefault() HomodyneDiscriminatorEstimatorUnrolled {
    const params = ht.CycleEstimatorParams{
        .smoothing_length = 4,
        .alpha_ema_quadrature_in_phase = 0.2,
        .alpha_ema_period = 0.2,
    };
    return HomodyneDiscriminatorEstimatorUnrolled.init(&params) catch unreachable;
}

fn createWarmUp(warm_up: usize) HomodyneDiscriminatorEstimatorUnrolled {
    const params = ht.CycleEstimatorParams{
        .smoothing_length = 4,
        .alpha_ema_quadrature_in_phase = 0.2,
        .alpha_ema_period = 0.2,
        .warm_up_period = warm_up,
    };
    return HomodyneDiscriminatorEstimatorUnrolled.init(&params) catch unreachable;
}

test "reference implementation: wma smoothed" {
    var hdeu = createDefault();
    const input = testInput();
    const exp = testExpectedSmoothed();
    const epsilon = 1e-8;

    for (0..3) |idx| {
        hdeu.update(input[idx]);
        try testing.expect(almostEqual(0, hdeu.smoothed(), epsilon));
    }

    for (3..252) |i| {
        hdeu.update(input[i]);
        try testing.expect(almostEqual(exp[i], hdeu.smoothed(), epsilon));
    }

    const previous = hdeu.smoothed();
    hdeu.update(math.nan(f64));
    try testing.expect(almostEqual(previous, hdeu.smoothed(), epsilon));
}

test "reference implementation: detrended" {
    var hdeu = createDefault();
    const input = testInput();
    const exp = testExpectedDetrended();
    const epsilon = 1e-8;

    for (0..3) |idx| {
        hdeu.update(input[idx]);
        try testing.expect(almostEqual(0, hdeu.detrended(), epsilon));
    }

    for (3..252) |i| {
        hdeu.update(input[i]);
        try testing.expect(almostEqual(exp[i], hdeu.detrended(), epsilon));
    }

    const previous = hdeu.detrended();
    hdeu.update(math.nan(f64));
    try testing.expect(almostEqual(previous, hdeu.detrended(), epsilon));
}

test "reference implementation: quadrature" {
    var hdeu = createDefault();
    const input = testInput();
    const exp = testExpectedQuadrature();
    const epsilon = 1e-8;

    for (0..3) |idx| {
        hdeu.update(input[idx]);
        try testing.expect(almostEqual(0, hdeu.quadrature(), epsilon));
    }

    for (3..252) |i| {
        hdeu.update(input[i]);
        try testing.expect(almostEqual(exp[i], hdeu.quadrature(), epsilon));
    }

    const previous = hdeu.quadrature();
    hdeu.update(math.nan(f64));
    try testing.expect(almostEqual(previous, hdeu.quadrature(), epsilon));
}

test "reference implementation: in-phase" {
    var hdeu = createDefault();
    const input = testInput();
    const exp = testExpectedInPhase();
    const epsilon = 1e-8;

    for (0..3) |idx| {
        hdeu.update(input[idx]);
        try testing.expect(almostEqual(0, hdeu.inPhase(), epsilon));
    }

    for (3..252) |i| {
        hdeu.update(input[i]);
        try testing.expect(almostEqual(exp[i], hdeu.inPhase(), epsilon));
    }

    const previous = hdeu.inPhase();
    hdeu.update(math.nan(f64));
    try testing.expect(almostEqual(previous, hdeu.inPhase(), epsilon));
}

test "reference implementation: period" {
    var hdeu = createDefault();
    const input = testInput();
    const exp = testExpectedPeriod();
    const epsilon = 1e-8;

    for (0..3) |idx| {
        hdeu.update(input[idx]);
        try testing.expect(almostEqual(6, hdeu.period(), epsilon));
    }

    for (3..252) |i| {
        hdeu.update(input[i]);
        try testing.expect(almostEqual(exp[i], hdeu.period(), epsilon));
    }

    const previous = hdeu.period();
    hdeu.update(math.nan(f64));
    try testing.expect(almostEqual(previous, hdeu.period(), epsilon));
}

test "reference implementation: primed" {
    var hdeu = createDefault();
    const input = testInput();
    const lprimed = 2 + 7 * 3; // 23

    try testing.expect(!hdeu.primed());

    for (0..lprimed) |i| {
        hdeu.update(input[i]);
        try testing.expect(!hdeu.primed());
    }

    for (lprimed..252) |i| {
        hdeu.update(input[i]);
        try testing.expect(hdeu.primed());
    }
}

test "reference implementation: primed with warmup" {
    const lprimed = 50;
    var hdeu = createWarmUp(lprimed);
    const input = testInput();

    try testing.expect(!hdeu.primed());

    for (0..lprimed) |i| {
        hdeu.update(input[i]);
        try testing.expect(!hdeu.primed());
    }

    for (lprimed..252) |i| {
        hdeu.update(input[i]);
        try testing.expect(hdeu.primed());
    }
}

test "period of sin input" {
    const period_val: f64 = 30;
    const omega = 2.0 * math.pi / period_val;
    var hdeu = createDefault();
    for (0..512) |i| {
        hdeu.update(@sin(omega * @as(f64, @floatFromInt(i))));
    }
    try testing.expect(almostEqual(period_val, hdeu.period(), 1e-2));
}

test "min period of sin input" {
    const period_val: f64 = 3;
    const omega = 2.0 * math.pi / period_val;
    var hdeu = createDefault();
    for (0..512) |i| {
        hdeu.update(@sin(omega * @as(f64, @floatFromInt(i))));
    }
    try testing.expect(almostEqual(@as(f64, @floatFromInt(hdeu.minPeriod())), hdeu.period(), 1e-14));
}

test "max period of sin input" {
    const period_val: f64 = 60;
    const omega = 2.0 * math.pi / period_val;
    var hdeu = createDefault();
    for (0..512) |i| {
        hdeu.update(@sin(omega * @as(f64, @floatFromInt(i))));
    }
    try testing.expect(almostEqual(@as(f64, @floatFromInt(hdeu.maxPeriod())), hdeu.period(), 1e-14));
}

test "constructor validation: smoothing length = 1" {
    const params = ht.CycleEstimatorParams{ .smoothing_length = 1 };
    const result = HomodyneDiscriminatorEstimatorUnrolled.init(&params);
    try testing.expectError(ht.VerifyError.InvalidSmoothingLength, result);
}

test "constructor validation: smoothing length = 5" {
    const params = ht.CycleEstimatorParams{ .smoothing_length = 5 };
    const result = HomodyneDiscriminatorEstimatorUnrolled.init(&params);
    try testing.expectError(ht.VerifyError.InvalidSmoothingLength, result);
}

test "constructor validation: alpha quad = 0" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_quadrature_in_phase = 0.0 };
    const result = HomodyneDiscriminatorEstimatorUnrolled.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaQuadratureInPhase, result);
}

test "constructor validation: alpha quad = 1" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_quadrature_in_phase = 1.0 };
    const result = HomodyneDiscriminatorEstimatorUnrolled.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaQuadratureInPhase, result);
}

test "constructor validation: alpha period = 0" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_period = 0.0 };
    const result = HomodyneDiscriminatorEstimatorUnrolled.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaPeriod, result);
}

test "constructor validation: alpha period = 1" {
    const params = ht.CycleEstimatorParams{ .alpha_ema_period = 1.0 };
    const result = HomodyneDiscriminatorEstimatorUnrolled.init(&params);
    try testing.expectError(ht.VerifyError.InvalidAlphaEmaPeriod, result);
}
