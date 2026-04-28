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

fn testInput() [252]f64 {
    return .{
        92.0000,  93.1725,  95.3125,  94.8450,  94.4075,  94.1100,  93.5000,  91.7350,  90.9550,  91.6875,
        94.5000,  97.9700,  97.5775,  90.7825,  89.0325,  92.0950,  91.1550,  89.7175,  90.6100,  91.0000,
        88.9225,  87.5150,  86.4375,  83.8900,  83.0025,  82.8125,  82.8450,  86.7350,  86.8600,  87.5475,
        85.7800,  86.1725,  86.4375,  87.2500,  88.9375,  88.2050,  85.8125,  84.5950,  83.6575,  84.4550,
        83.5000,  86.7825,  88.1725,  89.2650,  90.8600,  90.7825,  91.8600,  90.3600,  89.8600,  90.9225,
        89.5000,  87.6725,  86.5000,  84.2825,  82.9075,  84.2500,  85.6875,  86.6100,  88.2825,  89.5325,
        89.5000,  88.0950,  90.6250,  92.2350,  91.6725,  92.5925,  93.0150,  91.1725,  90.9850,  90.3775,
        88.2500,  86.9075,  84.0925,  83.1875,  84.2525,  97.8600,  99.8750,  103.2650, 105.9375, 103.5000,
        103.1100, 103.6100, 104.6400, 106.8150, 104.9525, 105.5000, 107.1400, 109.7350, 109.8450, 110.9850,
        120.0000, 119.8750, 117.9075, 119.4075, 117.9525, 117.2200, 115.6425, 113.1100, 111.7500, 114.5175,
        114.7450, 115.4700, 112.5300, 112.0300, 113.4350, 114.2200, 119.5950, 117.9650, 118.7150, 115.0300,
        114.5300, 115.0000, 116.5300, 120.1850, 120.5000, 120.5950, 124.1850, 125.3750, 122.9700, 123.0000,
        124.4350, 123.4400, 124.0300, 128.1850, 129.6550, 130.8750, 132.3450, 132.0650, 133.8150, 135.6600,
        137.0350, 137.4700, 137.3450, 136.3150, 136.4400, 136.2850, 129.0950, 128.3100, 126.0000, 124.0300,
        123.9350, 125.0300, 127.2500, 125.6200, 125.5300, 123.9050, 120.6550, 119.9650, 120.7800, 124.0000,
        122.7800, 120.7200, 121.7800, 122.4050, 123.2500, 126.1850, 127.5600, 126.5650, 123.0600, 122.7150,
        123.5900, 122.3100, 122.4650, 123.9650, 123.9700, 124.1550, 124.4350, 127.0000, 125.5000, 128.8750,
        130.5350, 132.3150, 134.0650, 136.0350, 133.7800, 132.7500, 133.4700, 130.9700, 127.5950, 128.4400,
        127.9400, 125.8100, 124.6250, 122.7200, 124.0900, 123.2200, 121.4050, 120.9350, 118.2800, 118.3750,
        121.1550, 120.9050, 117.1250, 113.0600, 114.9050, 112.4350, 107.9350, 105.9700, 106.3700, 106.8450,
        106.9700, 110.0300, 91.0000,  93.5600,  93.6200,  95.3100,  94.1850,  94.7800,  97.6250,  97.5900,
        95.2500,  94.7200,  92.2200,  91.5650,  92.2200,  93.8100,  95.5900,  96.1850,  94.6250,  95.1200,
        94.0000,  93.7450,  95.9050,  101.7450, 106.4400, 107.9350, 103.4050, 105.0600, 104.1550, 103.3100,
        103.3450, 104.8400, 110.4050, 114.5000, 117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300,
        106.2200, 107.7200, 109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
        109.5300, 108.0600,
    };
}

fn testExpectedSmoothed() [252]f64 {
    return .{
        0,          0,          0,          94.366250,  94.596250,
        94.466500,  93.999000,  93.006750,  92.013500,  91.658500,
        92.670750,  94.971000,  96.490750,  94.630250,  92.160250,
        91.462000,  90.975250,  90.555750,  90.599750,  90.642000,
        89.962750,  88.943750,  87.714000,  85.882500,  84.407000,
        83.447500,  82.971250,  84.410250,  85.614750,  86.708500,
        86.621750,  86.398500,  86.337500,  86.643750,  87.654750,
        88.057000,  87.299000,  86.116500,  84.824500,  84.379500,
        83.927500,  85.019750,  86.449250,  87.864250,  89.436250,
        90.241250,  91.077250,  90.944500,  90.502250,  90.585000,
        90.084750,  89.089500,  87.894000,  86.147500,  84.515000,
        84.078750,  84.559750,  85.491000,  86.858500,  88.188500,
        88.977250,  88.822750,  89.531750,  90.650500,  91.274000,
        92.048250,  92.541750,  92.059250,  91.608000,  90.982500,
        89.727500,  88.412000,  86.397000,  84.709250,  84.166500,
        89.466500,  94.477250,  99.265750,  103.115500, 103.821750,
        103.808000, 103.670750, 103.911000, 105.151000, 105.314500,
        105.512750, 106.178000, 107.631250, 108.836500, 110.008500,
        114.238000, 117.131500, 118.224000, 119.110250, 118.572250,
        117.946000, 116.954250, 115.176000, 113.483500, 113.518250,
        113.914250, 114.690000, 114.053750, 113.139500, 113.036000,
        113.377500, 115.994000, 117.252000, 118.216500, 117.179000,
        115.860500, 115.236500, 115.521000, 117.486000, 119.061500,
        120.078000, 121.971000, 123.574500, 123.697000, 123.584500,
        123.805500, 123.603500, 123.831000, 125.614500, 127.467500,
        129.286500, 130.950000, 131.670000, 132.702000, 134.056000,
        135.481500, 136.612000, 137.152000, 136.927000, 136.686500,
        136.443500, 133.443000, 130.953500, 128.340500, 125.983500,
        124.814000, 124.598500, 125.599000, 125.822500, 125.851000,
        125.070000, 123.101500, 121.516500, 120.823000, 121.892500,
        122.464500, 122.000000, 121.884000, 121.918000, 122.449500,
        124.108000, 125.770000, 126.456000, 125.324000, 124.073000,
        123.519000, 122.850000, 122.668500, 123.146500, 123.501500,
        123.892500, 124.211000, 125.358500, 125.602500, 127.043500,
        128.676500, 130.411500, 132.315000, 134.150000, 134.367000,
        133.847500, 133.572500, 132.357000, 130.298000, 129.195500,
        128.324000, 127.153500, 126.025000, 124.431500, 123.958000,
        123.521500, 122.618000, 121.848500, 120.195500, 119.161500,
        119.724000, 120.211500, 119.190000, 116.658000, 115.395500,
        113.770000, 111.191500, 108.746000, 107.169500, 106.636500,
        106.712500, 108.109000, 101.487500, 97.427000,  94.719000,
        94.022000,  94.347000,  94.591500,  95.852000,  96.698000,
        96.380000,  95.743500,  94.113000,  92.761000,  92.273500,
        92.725000,  93.979500,  95.135000,  95.204500,  95.231500,
        94.679500,  94.184500,  94.797500,  97.618500,  101.655000,
        105.045500, 105.205000, 105.276500, 104.654500, 103.923000,
        103.668000, 104.017000, 106.614000, 110.224000, 113.841000,
        116.435000, 117.262000, 114.437000, 112.049000, 109.977000,
        107.953000, 107.625500, 108.349000, 108.756000, 109.104500,
        109.192000, 109.200000, 109.158500, 109.458500, 109.565000,
        109.586500, 108.998500,
    };
}

fn testExpectedDetrended() [252]f64 {
    return .{
        0,                     0,                     0,                    8.98725291750000000,   9.00915765750000000,
        62.89229125575000000,  62.97912762075000000,  9.32026150012499000,  8.72833094955000000,   -1.18487042824499000,
        -1.50107340048300000,  -0.86554876465665300,  0.77455320732833200,  2.73425135915793000,   3.10719600483979000,
        -0.32043818348019300,  -4.20971694786669000,  -3.76607919981557000, -2.19840953444192000,  -1.66730405611637000,
        -0.82421664344656100,  -0.38972085057669300,  -1.44032300054316000, -3.09190726835819000,  -4.14005072156989000,
        -5.43783089745035000,  -5.80420011549713000,  -4.24476734097920000, -2.46885943954737000,  1.58710671977183000,
        4.43981858459690000,   4.09313796580003000,   2.22579638727991000,  0.08537893440106290,   0.07614965962751420,
        0.65627035940797800,   1.97596326645869000,   1.80442306972465000,  -0.76683833161150600,  -2.83984842772889000,
        -3.74697339934701000,  -2.68826283652387000,  -1.23064642254623000, 1.09946840528635000,   3.91972422308773000,
        4.57846163358676000,   4.91688320830916000,   3.81997889869049000,  2.55676637716418000,   1.26507167176780000,
        -0.50288443520152700,  -0.57617141736822500,  -0.96103685665957100, -2.27638430534806000,  -3.13716623486099000,
        -3.98079947738033000,  -4.26583247323698000,  -2.63959640039743000, -0.12692619104035100,  1.76433489230392000,
        3.10091823508968000,   3.56487192559447000,   3.00083561141979000,  1.51806394525404000,   1.31280489765731000,
        2.50935714439499000,   2.35255338597973000,   1.92841612829923000,  1.60036265582390000,   0.06634301061101840,
        -1.21377249429583000,  -1.76687971650801000,  -3.13318310164546000, -4.15128566296215000,  -5.08219214506417000,
        -4.50751700586996000,  -1.65504949371185000,  7.60380811494012000,  15.28911596036860000,  14.88477549627070000,
        12.99886906351420000,  7.19915926099098000,   2.38351083241801000,  0.92952210211501300,   0.51773586549053500,
        1.85518088811070000,   1.83829239536756000,   1.04306155568002000,  1.74675295275361000,   3.09896531604222000,
        4.46274347582128000,   4.71023535541215000,   8.22206979394878000,  10.20471942482790000,  6.45067137892342000,
        3.88610429720727000,   0.96991350698374700,   -1.82963426900828000, -2.98641863998011000,  -4.61916641640855000,
        -5.25745083982017000,  -2.72176646217929000,  -0.06609252040257180, 1.04039032886564000,   0.08017265813270270,
        -1.91662698969654000,  -0.80130442487024800,  0.77964217713282100,  4.25230033030034000,   5.28350296383388000,
        3.10502321031030000,   0.27115917330546200,   -2.79324779318748000, -2.21668023651026000,  -0.23487376211077200,
        3.19066742289005000,   5.20999393848774000,   4.59373390141711000,  5.07623775499030000,   5.52166161646234000,
        3.18346552831830000,   0.77584943073788900,   0.52199624043629500,  0.42565791776485100,   0.73761865440863100,
        3.24552740845475000,   5.31574324639712000,   5.63785437853044000,  5.76387266052843000,   4.63860064871392000,
        4.01639198580676000,   5.04837026370218000,   5.48653081543948000,  4.79457953146280000,   3.23349400610759000,
        1.01449886376103000,   -1.18833174543853000,  -2.17786864601495000, -7.34513291852914000,  -11.50426011352260000,
        -11.37221338511770000, -11.33559485984170000, -7.74619267543083000, -3.59741713526470000,  0.63099180612244000,
        1.78298570503918000,   -0.05284814696186360,  -1.88843174477108000, -5.01597392573878000,  -5.83342657377768000,
        -3.92922937752374000,  -0.18683843476507100,  1.95607761821063000,  0.23472360933602400,   -0.41309548234113700,
        0.38266745173604600,   1.49862889415607000,   4.02155722687378000,  5.33120420836294000,   3.61850994475616000,
        -0.35289594825326000,  -3.46092056067680000,  -3.18921954436038000, -2.49325013638114000,  -1.59946761080849000,
        0.36121674507225200,   1.28701295443815000,   1.55559487384262000,  1.53303420008103000,   2.57255681225572000,
        2.62870085569429000,   3.13397071228133000,   4.85862146022901000,  5.14708489427225000,   5.30680757348722000,
        5.04382996210236000,   2.99916871645245000,   0.02363191358535060,  -1.28708550920351000,  -2.71738883445928000,
        -5.02798528811439000,  -5.03676785247540000,  -3.88604064591759000, -4.08252818775635000,  -4.04110378590922000,
        -4.40039573137898000,  -3.62218844045606000,  -2.13018013334558000, -2.72853159066207000,  -3.04893739344158000,
        -3.81053374031141000,  -4.06902280248562000,  -1.37278934716961000, 0.25445656490840700,   -1.94767628533836000,
        -6.78718331682070000,  -8.15194602296862000,  -8.10019390801706000, -11.33153249871230000, -11.55092444836040000,
        -8.92673363548058000,  -4.71709009204003000,  -3.18121994360306000, -0.60270707950848200,  -10.05426111933760000,
        -16.75228621791540000, -11.36608161740570000, -7.69128618504172000, -1.92020294611357000,  0.70733464629796300,
        3.03978872650678000,   4.41680905312702000,   0.85901441370150100,  -2.16944309441368000,  -4.94188633543843000,
        -6.20700437333476000,  -3.70132242074857000,  -0.21667627392717300, 2.81825540058030000,   4.03015489676071000,
        2.25939281328576000,   0.46386462324942300,   -0.52712246258679300, -0.86458622397821600,  1.63986141262330000,
        6.85043506002733000,   11.27655864736460000,  11.91772714034700000, 6.82845339282070000,   1.81455653743714000,
        -0.32770330788056300,  -2.51142625160202000,  -1.28074066811762000, 1.48166065755279000,   6.83496943904234000,
        12.29335145847880000,  14.03376432173650000,  11.76231007817030000, 6.36898977425706000,   -2.99161050039115000,
        -9.14966540170821000,  -8.53172032731284000,  -7.61738045385190000, -4.27219556885827000,  -0.12066829540402900,
        1.28701281691344000,   1.21839253427298000,   0.83864843431442200,  0.32581834936928400,   0.11396996331983700,
        0.37278108469203800,   0.40559600219908000,
    };
}

fn testExpectedQuadrature() [252]f64 {
    return .{
        0,                     0,                     0,                     0.85592799335686500,   0.85801415698498500,
        11.12263478063980000,  11.14341537843060000,  33.11423067572620000,  32.80628957798340000,  -34.26282819437900000,
        -36.44577853942920000, -14.47254290608920000, -14.87523259354600000, -0.59436870756590500,  1.08163370103997000,
        3.21191138405851000,   1.71644163814649000,   -3.41291532315777000,  -7.92464724787348000,  -4.43497195796719000,
        1.50607267371026000,   2.43779146460479000,   2.23947875152554000,   1.73469635325904000,   -1.18581800238647000,
        -4.25093663922334000,  -4.58090657328628000,  -3.97577830705054000,  -2.53778253663147000,  2.84586883531416000,
        7.02347777802915000,   10.88612302323000000,  11.70523794304240000,  4.41376906994538000,   -2.43675590086246000,
        -5.81140549865358000,  -3.53762913951195000,  0.24993681636606400,   1.76659329560857000,   0.80900272922444700,
        -4.09103399086359000,  -6.23333019911698000,  -4.16437919056855000,  0.04016039234256580,   3.92812429899796000,
        6.01932747876865000,   7.75913530498010000,   5.18143806518233000,   1.79738056675322000,   -0.79917444960473200,
        -3.33661929425683000,  -3.56682961288299000,  -4.09399754549751000,  -2.83413751241676000,  -1.38469260666944000,
        -2.54547608904040000,  -2.78113427076003000,  -2.02607326491503000,  -0.98186385846887000,  2.02850582830166000,
        5.27797211699404000,   5.78602764229670000,   4.51938782080200000,   2.53296848492360000,   0.14238401319210600,
        -1.95211979228278000,  -1.82516727520038000,  0.71506000697959200,   0.79941696987506900,   -0.82278348322309200,
        -1.19530761185683000,  -2.70332223657632000,  -4.02158101434879000,  -3.11338826530719000,  -3.37273306595388000,
        -3.58626979123510000,  -2.32713075465888000,  1.39679273990094000,   7.58598007470388000,   17.52023981780450000,
        21.77754701293870000,  9.59917697735426000,   -1.70094139331706000,  -9.84900959965207000,  -14.41635598882900000,
        -8.88766855049639000,  -3.80886086039557000,  -0.10295243943533700,  1.25899710631473000,   -0.47657162828701400,
        0.60949091013001800,   2.76409564684112000,   4.19469719141599000,   3.54455169670085000,   5.22510210240912000,
        6.62019171404484000,   -2.84952383793055000,  -9.09597034418764000,  -9.11341691990470000,  -10.21482014541620000,
        -7.31366631380089000,  -4.81280173622443000,  -3.05219500231647000,  2.97053741068114000,   7.04712331791934000,
        5.12995214653134000,   1.06213522465482000,   -2.78133371852436000,  -0.18788825034758700,  3.95420177678643000,
        6.40719482727042000,   5.57054232367856000,   -1.69718961277596000,  -6.41893916460208000,  -7.86075190953845000,
        -3.31198040061866000,  3.32449971744085000,   7.06889624495551000,   8.02756916242835000,   3.29512786711032000,
        0.55176250761359500,   0.68158809564185800,   -3.33464127808794000,  -6.44791649349957000,  -3.81631812038157000,
        -0.79964336650902100,  0.62890738001252300,   4.07982118131883000,   6.33361436702288000,   3.78665473615762000,
        1.29382896273133000,   -0.97764671793158800,  -2.47301652148446000,  0.37594131404506000,   1.45157122041508000,
        -1.22077519109172000,  -4.60465786878931000,  -7.60777054441535000,  -10.22639724466520000, -9.29519426709912000,
        -13.79697921945210000, -18.58334726553210000, -8.20462196428375000,  -0.10925649063314200,  8.45501845525402000,
        16.56069010500390000,  16.16178539648960000,  10.37200446873600000,  -0.32332579486043600,  -5.60534588599990000,
        -7.90978168063863000,  -5.87384866566135000,  1.93331661517267000,   8.07382074869962000,   8.88612737690198000,
        1.94116961880012000,   -1.96478539796873000,  1.16504848164629000,   3.38773034724125000,   5.61771158983562000,
        5.06633969778975000,   -1.39353602473123000,  -8.87995480427041000,  -11.47090828573120000, -5.53108859416250000,
        0.57559330318340100,   2.52846391345203000,   4.93144178171995000,   4.69843040972786000,   2.47927695331917000,
        1.10896798915647000,   1.67246091046829000,   1.85682046699541000,   1.23545367358066000,   2.97514669920631000,
        2.51018011450836000,   0.53320725674937600,   -0.67052703620189600,  -3.79288866194180000,  -7.42678844558002000,
        -7.05626209907513000,  -5.20660956095194000,  -5.87735765622913000,  -3.64617943717568000,  0.82201895413307000,
        0.80805512373699700,   0.09522967861753010,   0.19800633087284200,   0.72215261913224100,   2.91843910043223000,
        1.13529807838677000,   -1.08479482518601000,  -0.93031018639129200,  -0.85860305008168300,  3.74811472172704000,
        5.64021766392183000,   -2.02939477563566000,  -13.01767841907970000, -14.35487330578800000, -5.66165957161537000,
        -7.08823729446843000,  -4.79749011205560000,  4.95720273139486000,   11.74107528578980000,  8.20531597011912000,
        4.25445053180659000,   -9.37014100711369000,  -22.62300531123650000, -1.61250412540437000,  14.65313624441180000,
        19.83795355234590000,  22.01172629030180000,  12.29472284560720000,  7.97392329630181000,   -4.63285658929020000,
        -13.17897372505700000, -11.44377596452480000, -7.58174858520786000,  2.33948714858062000,   10.03143122429620000,
        10.72645371017750000,  7.31428324634381000,   -0.04008962238147510,  -5.01965004665468000,  -4.09815661747972000,
        -1.15940191924281000,  4.80595108806550000,   12.36527079127330000,  14.28258636399240000,  7.80283863443487000,
        -7.28580805959545000,  -19.20911352815760000, -15.76355716486980000, -9.78017211736507000,  -1.45304367850065000,
        8.50909791922158000,   15.53662839007720000,  19.52137732735850000,  12.47276472234150000,  -1.87339740268066000,
        -15.25760462402560000, -26.22763533089890000, -26.10286332078850000, -10.63756123061160000, 0.57169865869596600,
        6.40288398255792000,   11.66695443018620000,  8.63556902693533000,   3.09381531978735000,   0.31811975854348500,
        -0.89136299286429000,  -0.94470730253761400,
    };
}

fn testExpectedInPhase() [252]f64 {
    return .{
        0,                    0,                     0,                     0.00000000000000000,   0.00000000000000000,
        0.00000000000000000,  8.98725291750000000,   9.00915765750000000,   62.89229125575000000,  62.97912762075000000,
        9.32026150012499000,  8.72833094955000000,   -1.18487042824499000,  -1.50107340048300000,  -0.86554876465665300,
        0.77455320732833200,  2.73425135915793000,   3.10719600483979000,   -0.32043818348019300,  -4.20971694786669000,
        -3.76607919981557000, -2.19840953444192000,  -1.66730405611637000,  -0.82421664344656100,  -0.38972085057669300,
        -1.44032300054316000, -3.09190726835819000,  -4.14005072156989000,  -5.43783089745035000,  -5.80420011549713000,
        -4.24476734097920000, -2.46885943954737000,  1.58710671977183000,   4.43981858459690000,   4.09313796580003000,
        2.22579638727991000,  0.08537893440106290,   0.07614965962751420,   0.65627035940797800,   1.97596326645869000,
        1.80442306972465000,  -0.76683833161150600,  -2.83984842772889000,  -3.74697339934701000,  -2.68826283652387000,
        -1.23064642254623000, 1.09946840528635000,   3.91972422308773000,   4.57846163358676000,   4.91688320830916000,
        3.81997889869049000,  2.55676637716418000,   1.26507167176780000,   -0.50288443520152700,  -0.57617141736822500,
        -0.96103685665957100, -2.27638430534806000,  -3.13716623486099000,  -3.98079947738033000,  -4.26583247323698000,
        -2.63959640039743000, -0.12692619104035100,  1.76433489230392000,   3.10091823508968000,   3.56487192559447000,
        3.00083561141979000,  1.51806394525404000,   1.31280489765731000,   2.50935714439499000,   2.35255338597973000,
        1.92841612829923000,  1.60036265582390000,   0.06634301061101840,   -1.21377249429583000,  -1.76687971650801000,
        -3.13318310164546000, -4.15128566296215000,  -5.08219214506417000,  -4.50751700586996000,  -1.65504949371185000,
        7.60380811494012000,  15.28911596036860000,  14.88477549627070000,  12.99886906351420000,  7.19915926099098000,
        2.38351083241801000,  0.92952210211501300,   0.51773586549053500,   1.85518088811070000,   1.83829239536756000,
        1.04306155568002000,  1.74675295275361000,   3.09896531604222000,   4.46274347582128000,   4.71023535541215000,
        8.22206979394878000,  10.20471942482790000,  6.45067137892342000,   3.88610429720727000,   0.96991350698374700,
        -1.82963426900828000, -2.98641863998011000,  -4.61916641640855000,  -5.25745083982017000,  -2.72176646217929000,
        -0.06609252040257180, 1.04039032886564000,   0.08017265813270270,   -1.91662698969654000,  -0.80130442487024800,
        0.77964217713282100,  4.25230033030034000,   5.28350296383388000,   3.10502321031030000,   0.27115917330546200,
        -2.79324779318748000, -2.21668023651026000,  -0.23487376211077200,  3.19066742289005000,   5.20999393848774000,
        4.59373390141711000,  5.07623775499030000,   5.52166161646234000,   3.18346552831830000,   0.77584943073788900,
        0.52199624043629500,  0.42565791776485100,   0.73761865440863100,   3.24552740845475000,   5.31574324639712000,
        5.63785437853044000,  5.76387266052843000,   4.63860064871392000,   4.01639198580676000,   5.04837026370218000,
        5.48653081543948000,  4.79457953146280000,   3.23349400610759000,   1.01449886376103000,   -1.18833174543853000,
        -2.17786864601495000, -7.34513291852914000,  -11.50426011352260000, -11.37221338511770000, -11.33559485984170000,
        -7.74619267543083000, -3.59741713526470000,  0.63099180612244000,   1.78298570503918000,   -0.05284814696186360,
        -1.88843174477108000, -5.01597392573878000,  -5.83342657377768000,  -3.92922937752374000,  -0.18683843476507100,
        1.95607761821063000,  0.23472360933602400,   -0.41309548234113700,  0.38266745173604600,   1.49862889415607000,
        4.02155722687378000,  5.33120420836294000,   3.61850994475616000,   -0.35289594825326000,  -3.46092056067680000,
        -3.18921954436038000, -2.49325013638114000,  -1.59946761080849000,  0.36121674507225200,   1.28701295443815000,
        1.55559487384262000,  1.53303420008103000,   2.57255681225572000,   2.62870085569429000,   3.13397071228133000,
        4.85862146022901000,  5.14708489427225000,   5.30680757348722000,   5.04382996210236000,   2.99916871645245000,
        0.02363191358535060,  -1.28708550920351000,  -2.71738883445928000,  -5.02798528811439000,  -5.03676785247540000,
        -3.88604064591759000, -4.08252818775635000,  -4.04110378590922000,  -4.40039573137898000,  -3.62218844045606000,
        -2.13018013334558000, -2.72853159066207000,  -3.04893739344158000,  -3.81053374031141000,  -4.06902280248562000,
        -1.37278934716961000, 0.25445656490840700,   -1.94767628533836000,  -6.78718331682070000,  -8.15194602296862000,
        -8.10019390801706000, -11.33153249871230000, -11.55092444836040000, -8.92673363548058000,  -4.71709009204003000,
        -3.18121994360306000, -0.60270707950848200,  -10.05426111933760000, -16.75228621791540000, -11.36608161740570000,
        -7.69128618504172000, -1.92020294611357000,  0.70733464629796300,   3.03978872650678000,   4.41680905312702000,
        0.85901441370150100,  -2.16944309441368000,  -4.94188633543843000,  -6.20700437333476000,  -3.70132242074857000,
        -0.21667627392717300, 2.81825540058030000,   4.03015489676071000,   2.25939281328576000,   0.46386462324942300,
        -0.52712246258679300, -0.86458622397821600,  1.63986141262330000,   6.85043506002733000,   11.27655864736460000,
        11.91772714034700000, 6.82845339282070000,   1.81455653743714000,   -0.32770330788056300,  -2.51142625160202000,
        -1.28074066811762000, 1.48166065755279000,   6.83496943904234000,   12.29335145847880000,  14.03376432173650000,
        11.76231007817030000, 6.36898977425706000,   -2.99161050039115000,  -9.14966540170821000,  -8.53172032731284000,
        -7.61738045385190000, -4.27219556885827000,  -0.12066829540402900,  1.28701281691344000,   1.21839253427298000,
        0.83864843431442200,  0.32581834936928400,
    };
}

fn testExpectedPeriod() [252]f64 {
    return .{
        0,                  0,                  0,                  6.000000000000000,  6.000000000000000,
        6.000000000000000,  6.600000000000000,  6.480000000000000,  7.128000000000000,  7.840800000000000,
        8.624880000000000,  9.487368000000010,  10.436104800000000, 11.479715280000000, 12.627686808000000,
        13.890455488800000, 15.092632992472200, 16.243311991597400, 17.324083690403400, 18.475004245551700,
        19.793928938691700, 20.991150791562600, 21.643352022039500, 21.959705519424200, 22.295567484037000,
        22.791325682745500, 23.544576614590200, 24.752147731605700, 26.131811393682200, 26.858904594382200,
        26.702958409577900, 25.600860243375200, 24.401023548778500, 24.269231586852900, 25.063772525606900,
        24.731935337132500, 23.324893884513200, 21.947699992262300, 21.114348897399200, 20.771351737837100,
        20.494233851604400, 20.183056650705500, 20.073950730222800, 20.329324851599500, 20.486682177547600,
        19.991320129284400, 19.037214654890800, 18.305338289364600, 18.072077074725200, 17.695410800682500,
        16.943409301403400, 16.222025478081200, 15.720412102044600, 15.535348292553700, 15.647225038303600,
        15.726724315115400, 15.659028152755600, 15.731926272152600, 16.066872450769700, 16.353132283152400,
        16.417488514615400, 16.327478042059500, 16.274657444500200, 16.306049424020600, 16.267229242589500,
        16.069414540211500, 15.804303177834000, 15.717155609449600, 15.905384578368700, 16.341416455846900,
        17.062385718849300, 17.726097022447900, 18.074288402433000, 18.494521449266000, 19.137157817464000,
        19.391962256800300, 19.560392220028600, 19.777094480192300, 19.290751623371500, 18.017562016229000,
        16.828402923157900, 17.114778109331900, 18.671307835949900, 18.276196598474400, 17.131108499737400,
        16.420699130133500, 16.384257384254200, 16.761284812463100, 17.257329929450600, 17.674593986434500,
        18.030662699220900, 18.444179030948200, 18.897999595547500, 19.377943804339100, 19.997497366949900,
        20.779094276230200, 21.173848640418500, 21.457828654575300, 21.634241987533300, 21.407595434921600,
        21.384357614634800, 21.666002925735200, 21.695440685430200, 21.360632335984100, 20.943570860148100,
        20.402535751359300, 19.887372972108100, 19.709970997965800, 19.649125076619000, 19.443232041094100,
        19.243827618745400, 19.313747204284000, 19.714913779763200, 20.134291666228900, 19.783181020847000,
        19.209690641670400, 19.460010571864000, 20.257433211912800, 21.074817453763800, 22.030867527260900,
        22.791291008915400, 21.616607563144900, 20.189911463977300, 18.857377307354800, 18.125831538562300,
        18.247592228820300, 18.771553474631800, 19.648253936845000, 21.078372413428500, 22.857039083933000,
        25.142742992326400, 26.049798360798400, 25.082168957220200, 24.806386106021400, 25.679481374169500,
        26.899647278044100, 28.071461927551400, 28.820896172665600, 29.148840897874100, 29.911641725893800,
        30.522301095492400, 29.831708133341600, 29.907565347137900, 32.228991190374700, 31.252506165184100,
        29.189840758281900, 27.263311268235300, 25.493790649485100, 24.833135719882200, 24.740125292429300,
        24.567302182123800, 24.228316198340700, 23.897133073236100, 23.670442010647200, 23.568722805351700,
        23.816626211788300, 24.489293689253900, 24.441197040242600, 23.690363029328200, 23.272337293810300,
        23.650689958418800, 24.541373135015500, 25.266648506194000, 24.822433791959300, 24.137921243243200,
        24.164085898425800, 23.679447801608500, 22.356967000947100, 20.900937564817300, 19.748775833847100,
        18.929417088080100, 18.174749782048300, 17.408877728644600, 16.849194933070500, 16.721820589710600,
        16.965495052172900, 17.718617550863000, 19.105336881323600, 19.909658447444800, 19.937274080580900,
        20.012914257426900, 20.584173964128200, 20.851919554908700, 20.628402644809500, 20.525657062631700,
        20.534404371242100, 20.233452244659500, 20.074871676384200, 20.414790876648000, 20.950309986790800,
        21.835118453187100, 23.216096664683200, 24.681124488027900, 26.533680552485700, 28.030508520296200,
        28.912019935050600, 31.803221928555700, 34.983544121411300, 32.674630209398100, 30.518104615577900,
        28.503909710949700, 28.230501292765300, 26.367288207442800, 24.627047185751600, 23.082674396417400,
        22.548443279717300, 24.211646687633100, 26.632811356396400, 29.296092492036100, 32.225701741239700,
        35.448271915363600, 33.402035063759300, 32.601996512141400, 32.684060257133300, 32.156947268423700,
        30.993331799166300, 29.228685633205700, 27.299592381414100, 25.807182056140100, 24.910990815824500,
        24.343386608043300, 24.151018312250900, 24.374269517606000, 24.552001576180600, 24.025117913609500,
        23.060588115528900, 22.492139589927800, 23.186782719829300, 25.505460991812300, 28.056007090993500,
        30.861607800092900, 32.186253240004200, 30.061960526164000, 28.077871131437100, 27.059021434006100,
        26.964120571638100, 27.001238343745600, 26.816631426266700, 26.706734370991900, 26.936033848694400,
        26.057711959547000, 24.337902970216900, 22.731601374182600, 22.155129320948500, 22.557429753923800,
        22.028068563989900, 20.824799623543400, 19.644851848783900, 18.776215429172100, 18.224498541335600,
        17.848949436739100, 17.538093820109200,
    };
}

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
