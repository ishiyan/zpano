const std = @import("std");
const math = std.math;

pub const ht_length: usize = 7;
pub const quadrature_index: usize = ht_length / 2; // 3
pub const default_min_period = 6.0;
pub const default_max_period = 50.0;
pub const accumulation_length: usize = 40;

/// Enumerates types of techniques to estimate an instantaneous period
/// using a Hilbert transformer.
pub const CycleEstimatorType = enum(u8) {
    homodyne_discriminator = 1,
    homodyne_discriminator_unrolled = 2,
    phase_accumulator = 3,
    dual_differentiator = 4,
};

/// Parameters to create an instance of the Hilbert transformer cycle estimator.
pub const CycleEstimatorParams = struct {
    /// The smoothing length of the underlying linear-Weighted Moving Average (WMA).
    /// Valid values are 2, 3, 4. Default is 4.
    smoothing_length: usize = 4,
    /// The value of α (0 < α < 1) used in EMA to smooth the in-phase and quadrature components.
    alpha_ema_quadrature_in_phase: f64 = 0.2,
    /// The value of α (0 < α < 1) used in EMA to smooth the instantaneous period.
    alpha_ema_period: f64 = 0.2,
    /// The number of updates before the estimator is primed.
    warm_up_period: usize = 0,
};

pub const VerifyError = error{
    InvalidSmoothingLength,
    InvalidAlphaEmaQuadratureInPhase,
    InvalidAlphaEmaPeriod,
};

/// Validates cycle estimator parameters.
pub fn verifyParameters(p: *const CycleEstimatorParams) VerifyError!void {
    if (p.smoothing_length < 2 or p.smoothing_length > 4) {
        return VerifyError.InvalidSmoothingLength;
    }
    if (p.alpha_ema_quadrature_in_phase <= 0 or p.alpha_ema_quadrature_in_phase >= 1) {
        return VerifyError.InvalidAlphaEmaQuadratureInPhase;
    }
    if (p.alpha_ema_period <= 0 or p.alpha_ema_period >= 1) {
        return VerifyError.InvalidAlphaEmaPeriod;
    }
}

/// Shifts all elements to the right and places the new value at index zero.
pub fn push(comptime N: usize, array: *[N]f64, value: f64) void {
    comptime var i: usize = N - 1;
    inline while (i > 0) : (i -= 1) {
        array[i] = array[i - 1];
    }
    array[0] = value;
}

/// Amplitude correction factor: 0.54 + 0.075 * period.
pub fn correctAmplitude(previous_period: f64) f64 {
    return 0.54 + 0.075 * previous_period;
}

/// Hilbert transform of a 7-element array.
pub fn htTransform(array: *const [ht_length]f64) f64 {
    const a = 0.0962;
    const b = 0.5769;
    var value: f64 = 0;
    value += a * array[0];
    value += b * array[2];
    value -= b * array[4];
    value -= a * array[6];
    return value;
}

/// Adjusts period to be within [0.67*prev, 1.5*prev] then [minPeriod, maxPeriod].
pub fn adjustPeriod(period: f64, period_previous: f64) f64 {
    var p = period;
    const max_prev = 1.5 * period_previous;
    if (p > max_prev) {
        p = max_prev;
    } else {
        const min_prev = 0.67 * period_previous;
        if (p < min_prev) {
            p = min_prev;
        }
    }
    if (p < default_min_period) {
        p = default_min_period;
    } else if (p > default_max_period) {
        p = default_max_period;
    }
    return p;
}

/// Fills WMA factors based on the smoothing length.
pub fn fillWmaFactors(length: usize, factors: *[4]f64) void {
    if (length == 4) {
        factors[0] = 4.0 / 10.0;
        factors[1] = 3.0 / 10.0;
        factors[2] = 2.0 / 10.0;
        factors[3] = 1.0 / 10.0;
    } else if (length == 3) {
        factors[0] = 3.0 / 6.0;
        factors[1] = 2.0 / 6.0;
        factors[2] = 1.0 / 6.0;
        factors[3] = 0;
    } else { // length == 2
        factors[0] = 2.0 / 3.0;
        factors[1] = 1.0 / 3.0;
        factors[2] = 0;
        factors[3] = 0;
    }
}

/// Computes WMA of raw_values using wma_factors for the given smoothing_length.
pub fn wma(raw_values: *const [4]f64, wma_factors: *const [4]f64, smoothing_length: usize) f64 {
    var value: f64 = 0;
    for (0..smoothing_length) |i| {
        value += wma_factors[i] * raw_values[i];
    }
    return value;
}

/// EMA step: alpha * value + (1 - alpha) * previous.
pub fn emaStep(alpha: f64, value: f64, previous: f64) f64 {
    return alpha * value + (1.0 - alpha) * previous;
}

/// Returns the moniker string for a cycle estimator, e.g. "hd(4, 0.200, 0.200)".
pub fn estimatorMoniker(buf: []u8, typ: CycleEstimatorType, estimator: *const CycleEstimator) []const u8 {
    const prefix: []const u8 = switch (typ) {
        .homodyne_discriminator => "hd",
        .homodyne_discriminator_unrolled => "hdu",
        .phase_accumulator => "pa",
        .dual_differentiator => "dd",
    };

    const result = std.fmt.bufPrint(buf, "{s}({d}, {d:.3}, {d:.3})", .{
        prefix,
        estimator.smoothingLength(),
        estimator.alphaEmaQuadratureInPhase(),
        estimator.alphaEmaPeriod(),
    }) catch return "";
    return result;
}

// Re-export the estimator types.
pub const HomodyneDiscriminatorEstimator = @import("homodyne_discriminator.zig").HomodyneDiscriminatorEstimator;
pub const HomodyneDiscriminatorEstimatorUnrolled = @import("homodyne_discriminator_unrolled.zig").HomodyneDiscriminatorEstimatorUnrolled;
pub const PhaseAccumulatorEstimator = @import("phase_accumulator.zig").PhaseAccumulatorEstimator;
pub const DualDifferentiatorEstimator = @import("dual_differentiator.zig").DualDifferentiatorEstimator;

/// Creates a new cycle estimator based on the specified type and parameters.
pub fn newCycleEstimator(typ: CycleEstimatorType, params: *const CycleEstimatorParams) VerifyError!CycleEstimator {
    return switch (typ) {
        .homodyne_discriminator => .{ .homodyne_discriminator = try HomodyneDiscriminatorEstimator.init(params) },
        .homodyne_discriminator_unrolled => .{ .homodyne_discriminator_unrolled = try HomodyneDiscriminatorEstimatorUnrolled.init(params) },
        .phase_accumulator => .{ .phase_accumulator = try PhaseAccumulatorEstimator.init(params) },
        .dual_differentiator => .{ .dual_differentiator = try DualDifferentiatorEstimator.init(params) },
    };
}

/// Tagged union wrapping all cycle estimator types with a common interface.
pub const CycleEstimator = union(enum) {
    homodyne_discriminator: HomodyneDiscriminatorEstimator,
    homodyne_discriminator_unrolled: HomodyneDiscriminatorEstimatorUnrolled,
    phase_accumulator: PhaseAccumulatorEstimator,
    dual_differentiator: DualDifferentiatorEstimator,

    pub fn update(self: *CycleEstimator, sample: f64) void {
        switch (self.*) {
            inline else => |*e| e.update(sample),
        }
    }

    pub fn period(self: *const CycleEstimator) f64 {
        return switch (self.*) {
            inline else => |*e| e.period(),
        };
    }

    pub fn smoothed(self: *const CycleEstimator) f64 {
        return switch (self.*) {
            inline else => |*e| e.smoothed(),
        };
    }

    pub fn detrended(self: *const CycleEstimator) f64 {
        return switch (self.*) {
            inline else => |*e| e.detrended(),
        };
    }

    pub fn quadrature(self: *const CycleEstimator) f64 {
        return switch (self.*) {
            inline else => |*e| e.quadrature(),
        };
    }

    pub fn inPhase(self: *const CycleEstimator) f64 {
        return switch (self.*) {
            inline else => |*e| e.inPhase(),
        };
    }

    pub fn primed(self: *const CycleEstimator) bool {
        return switch (self.*) {
            inline else => |*e| e.primed(),
        };
    }

    pub fn count(self: *const CycleEstimator) usize {
        return switch (self.*) {
            inline else => |*e| e.count(),
        };
    }

    pub fn smoothingLength(self: *const CycleEstimator) usize {
        return switch (self.*) {
            inline else => |*e| e.smoothingLength(),
        };
    }

    pub fn minPeriod(self: *const CycleEstimator) usize {
        return switch (self.*) {
            inline else => |*e| e.minPeriod(),
        };
    }

    pub fn maxPeriod(self: *const CycleEstimator) usize {
        return switch (self.*) {
            inline else => |*e| e.maxPeriod(),
        };
    }

    pub fn warmUpPeriod(self: *const CycleEstimator) usize {
        return switch (self.*) {
            inline else => |*e| e.warmUpPeriod(),
        };
    }

    pub fn alphaEmaQuadratureInPhase(self: *const CycleEstimator) f64 {
        return switch (self.*) {
            inline else => |*e| e.alphaEmaQuadratureInPhase(),
        };
    }

    pub fn alphaEmaPeriod(self: *const CycleEstimator) f64 {
        return switch (self.*) {
            inline else => |*e| e.alphaEmaPeriod(),
        };
    }
};
