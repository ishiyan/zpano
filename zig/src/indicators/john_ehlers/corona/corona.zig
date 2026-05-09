const std = @import("std");
const math = std.math;

// Constants matching Go's corona package.
const default_high_pass_filter_cutoff = 30;
const default_minimal_period = 6;
const default_maximal_period = 30;
const default_decibels_lower_threshold = 6.0;
const default_decibels_upper_threshold = 20.0;

const high_pass_filter_buffer_size = 6;
const fir_coef_sum = 12.0;

const delta_lower_threshold = 0.1;
const delta_factor = -0.015;
const delta_summand = 0.5;

const dominant_cycle_buffer_size = 5;
const dominant_cycle_median_index = 2;

const decibels_smoothing_alpha = 0.33;
const decibels_smoothing_one_minus = 0.67;

const normalized_amplitude_factor = 0.99;
const decibels_floor = 0.01;
const decibels_gain = 10.0;

/// Per-bin state of a single bandpass filter in the bank.
pub const Filter = struct {
    in_phase: f64 = 0,
    in_phase_previous: f64 = 0,
    quadrature: f64 = 0,
    quadrature_previous: f64 = 0,
    real: f64 = 0,
    real_previous: f64 = 0,
    imaginary: f64 = 0,
    imaginary_previous: f64 = 0,
    amplitude_squared: f64 = 0,
    decibels: f64 = 0,
};

/// Parameters for creating a Corona engine.
pub const Params = struct {
    high_pass_filter_cutoff: i32 = 0,
    minimal_period: i32 = 0,
    maximal_period: i32 = 0,
    decibels_lower_threshold: f64 = 0,
    decibels_upper_threshold: f64 = 0,
};

pub const Error = error{
    InvalidHighPassFilterCutoff,
    InvalidMinimalPeriod,
    InvalidMaximalPeriod,
    InvalidDecibelsLowerThreshold,
    InvalidDecibelsUpperThreshold,
    OutOfMemory,
};

/// Corona is the shared spectral-analysis engine consumed by all corona indicators.
pub const Corona = struct {
    allocator: std.mem.Allocator,

    // Configuration (immutable after construction).
    minimal_period: i32,
    maximal_period: i32,
    minimal_period_times_two: i32,
    maximal_period_times_two: i32,
    filter_bank_length: usize,
    decibels_lower_threshold: f64,
    decibels_upper_threshold: f64,

    // High-pass filter coefficients.
    alpha: f64,
    half_one_plus_alpha: f64,

    // Pre-calculated cos(4π/n) for each half-period index.
    pre_calculated_beta: []f64,

    // HP ring buffer (oldest at index 0, current at index 5).
    high_pass_buffer: [high_pass_filter_buffer_size]f64,

    // Previous raw sample and previous smoothed HP.
    sample_previous: f64,
    smooth_hp_previous: f64,

    // Filter bank.
    filter_bank: []Filter,

    // Running maximum of amplitude-squared across filter bins for current bar.
    maximal_amplitude_squared: f64,

    // 5-sample ring buffer for dominant cycle median.
    dominant_cycle_buffer: [dominant_cycle_buffer_size]f64,

    // Sample counter.
    sample_count: i32,

    // Most recent dominant cycle estimate and its 5-sample median.
    dominant_cycle: f64,
    dominant_cycle_median: f64,

    primed: bool,

    pub fn init(allocator: std.mem.Allocator, p: ?Params) Error!Corona {
        var cfg = p orelse Params{};
        applyDefaults(&cfg);

        try verifyParameters(&cfg);

        const min_p = cfg.minimal_period;
        const max_p = cfg.maximal_period;
        const min_p2 = min_p * 2;
        const max_p2 = max_p * 2;
        const fbl: usize = @intCast(max_p2 - min_p2 + 1);

        // HP filter coefficients.
        const phi = 2.0 * math.pi / @as(f64, @floatFromInt(cfg.high_pass_filter_cutoff));
        const a = (1.0 - @sin(phi)) / @cos(phi);
        const half_one_plus_a = 0.5 * (1.0 + a);

        // Pre-calculate β = cos(4π / n) for each half-period index.
        const beta_buf = allocator.alloc(f64, fbl) catch return error.OutOfMemory;
        errdefer allocator.free(beta_buf);
        for (0..fbl) |index| {
            const n: f64 = @floatFromInt(min_p2 + @as(i32, @intCast(index)));
            beta_buf[index] = @cos(4.0 * math.pi / n);
        }

        const fb = allocator.alloc(Filter, fbl) catch return error.OutOfMemory;
        errdefer allocator.free(fb);
        @memset(fb, Filter{});

        var dc_buf: [dominant_cycle_buffer_size]f64 = undefined;
        @memset(&dc_buf, math.floatMax(f64));

        return .{
            .allocator = allocator,
            .minimal_period = min_p,
            .maximal_period = max_p,
            .minimal_period_times_two = min_p2,
            .maximal_period_times_two = max_p2,
            .filter_bank_length = fbl,
            .decibels_lower_threshold = cfg.decibels_lower_threshold,
            .decibels_upper_threshold = cfg.decibels_upper_threshold,
            .alpha = a,
            .half_one_plus_alpha = half_one_plus_a,
            .pre_calculated_beta = beta_buf,
            .high_pass_buffer = [_]f64{0} ** high_pass_filter_buffer_size,
            .sample_previous = 0,
            .smooth_hp_previous = 0,
            .filter_bank = fb,
            .maximal_amplitude_squared = 0,
            .dominant_cycle_buffer = dc_buf,
            .sample_count = 0,
            .dominant_cycle = math.floatMax(f64),
            .dominant_cycle_median = math.floatMax(f64),
            .primed = false,
        };
    }

    pub fn deinit(self: *Corona) void {
        self.allocator.free(self.pre_calculated_beta);
        self.allocator.free(self.filter_bank);
    }

    /// Feeds the next sample to the engine.
    /// Returns true once primed and outputs are meaningful.
    pub fn update(self: *Corona, sample: f64) bool {
        if (math.isNan(sample)) {
            return self.primed;
        }

        self.sample_count += 1;

        // First sample: just store as prior reference.
        if (self.sample_count == 1) {
            self.sample_previous = sample;
            return false;
        }

        // Step 1: High-pass filter.
        const hp = self.alpha * self.high_pass_buffer[high_pass_filter_buffer_size - 1] +
            self.half_one_plus_alpha * (sample - self.sample_previous);
        self.sample_previous = sample;

        // Shift buffer left.
        var i: usize = 0;
        while (i < high_pass_filter_buffer_size - 1) : (i += 1) {
            self.high_pass_buffer[i] = self.high_pass_buffer[i + 1];
        }
        self.high_pass_buffer[high_pass_filter_buffer_size - 1] = hp;

        // Step 2: 6-tap FIR smoothing {1, 2, 3, 3, 2, 1} / 12.
        const smooth_hp = (self.high_pass_buffer[0] +
            2.0 * self.high_pass_buffer[1] +
            3.0 * self.high_pass_buffer[2] +
            3.0 * self.high_pass_buffer[3] +
            2.0 * self.high_pass_buffer[4] +
            self.high_pass_buffer[5]) / fir_coef_sum;

        // Step 3: Momentum.
        const momentum = smooth_hp - self.smooth_hp_previous;
        self.smooth_hp_previous = smooth_hp;

        // Step 4: Adaptive delta.
        var delta = delta_factor * @as(f64, @floatFromInt(self.sample_count)) + delta_summand;
        if (delta < delta_lower_threshold) {
            delta = delta_lower_threshold;
        }

        // Step 5: Filter-bank update.
        self.maximal_amplitude_squared = 0;
        for (0..self.filter_bank_length) |index| {
            const n: f64 = @floatFromInt(self.minimal_period_times_two + @as(i32, @intCast(index)));

            const gamma = 1.0 / @cos(8.0 * math.pi * delta / n);
            const a_val = gamma - @sqrt(gamma * gamma - 1.0);

            const quadrature = momentum * (n / (4.0 * math.pi));
            const in_phase = smooth_hp;

            const half_one_min_a = 0.5 * (1.0 - a_val);
            const beta = self.pre_calculated_beta[index];
            const beta_one_plus_a = beta * (1.0 + a_val);

            var f = &self.filter_bank[index];

            const real = half_one_min_a * (in_phase - f.in_phase_previous) + beta_one_plus_a * f.real - a_val * f.real_previous;
            const imag = half_one_min_a * (quadrature - f.quadrature_previous) + beta_one_plus_a * f.imaginary - a_val * f.imaginary_previous;

            const amp_sq = real * real + imag * imag;

            f.in_phase_previous = f.in_phase;
            f.in_phase = in_phase;
            f.quadrature_previous = f.quadrature;
            f.quadrature = quadrature;
            f.real_previous = f.real;
            f.real = real;
            f.imaginary_previous = f.imaginary;
            f.imaginary = imag;
            f.amplitude_squared = amp_sq;

            if (amp_sq > self.maximal_amplitude_squared) {
                self.maximal_amplitude_squared = amp_sq;
            }
        }

        // Step 6: dB normalization and dominant-cycle weighted average.
        var numerator: f64 = 0;
        var denominator: f64 = 0;
        self.dominant_cycle = 0;

        for (0..self.filter_bank_length) |index| {
            var f = &self.filter_bank[index];

            var decibels: f64 = 0;
            if (self.maximal_amplitude_squared > 0) {
                const normalized = f.amplitude_squared / self.maximal_amplitude_squared;
                if (normalized > 0) {
                    const arg = (1.0 - normalized_amplitude_factor * normalized) / decibels_floor;
                    if (arg > 0) {
                        decibels = decibels_gain * @log10(arg);
                    }
                }
            }

            // EMA smoothing.
            decibels = decibels_smoothing_alpha * decibels + decibels_smoothing_one_minus * f.decibels;
            if (decibels > self.decibels_upper_threshold) {
                decibels = self.decibels_upper_threshold;
            }
            f.decibels = decibels;

            // Only bins at or below lower threshold contribute.
            if (decibels <= self.decibels_lower_threshold) {
                const n: f64 = @floatFromInt(self.minimal_period_times_two + @as(i32, @intCast(index)));
                const adjusted = self.decibels_upper_threshold - decibels;
                numerator += n * adjusted;
                denominator += adjusted;
            }
        }

        // DC = 0.5 * num / denom (converts half-period to period).
        if (denominator != 0) {
            self.dominant_cycle = 0.5 * numerator / denominator;
        }
        const min_period_f: f64 = @floatFromInt(self.minimal_period);
        if (self.dominant_cycle < min_period_f) {
            self.dominant_cycle = min_period_f;
        }

        // Step 7: 5-sample median of dominant cycle.
        {
            var j: usize = 0;
            while (j < dominant_cycle_buffer_size - 1) : (j += 1) {
                self.dominant_cycle_buffer[j] = self.dominant_cycle_buffer[j + 1];
            }
            self.dominant_cycle_buffer[dominant_cycle_buffer_size - 1] = self.dominant_cycle;
        }

        var sorted: [dominant_cycle_buffer_size]f64 = self.dominant_cycle_buffer;
        std.mem.sort(f64, &sorted, {}, std.sort.asc(f64));
        self.dominant_cycle_median = sorted[dominant_cycle_median_index];
        if (self.dominant_cycle_median < min_period_f) {
            self.dominant_cycle_median = min_period_f;
        }

        if (self.sample_count < self.minimal_period_times_two) {
            return false;
        }
        self.primed = true;

        return true;
    }

    // --- Accessors ---

    pub fn isPrimed(self: *const Corona) bool {
        return self.primed;
    }

    pub fn getDominantCycle(self: *const Corona) f64 {
        return self.dominant_cycle;
    }

    pub fn getDominantCycleMedian(self: *const Corona) f64 {
        return self.dominant_cycle_median;
    }

    pub fn getMaximalAmplitudeSquared(self: *const Corona) f64 {
        return self.maximal_amplitude_squared;
    }

    pub fn getFilterBank(self: *const Corona) []const Filter {
        return self.filter_bank;
    }

    pub fn getFilterBankLength(self: *const Corona) usize {
        return self.filter_bank_length;
    }

    pub fn getMinimalPeriod(self: *const Corona) i32 {
        return self.minimal_period;
    }

    pub fn getMaximalPeriod(self: *const Corona) i32 {
        return self.maximal_period;
    }

    pub fn getMinimalPeriodTimesTwo(self: *const Corona) i32 {
        return self.minimal_period_times_two;
    }

    pub fn getMaximalPeriodTimesTwo(self: *const Corona) i32 {
        return self.maximal_period_times_two;
    }
};

fn applyDefaults(p: *Params) void {
    if (p.high_pass_filter_cutoff <= 0) {
        p.high_pass_filter_cutoff = default_high_pass_filter_cutoff;
    }
    if (p.minimal_period <= 0) {
        p.minimal_period = default_minimal_period;
    }
    if (p.maximal_period <= 0) {
        p.maximal_period = default_maximal_period;
    }
    if (p.decibels_lower_threshold == 0) {
        p.decibels_lower_threshold = default_decibels_lower_threshold;
    }
    if (p.decibels_upper_threshold == 0) {
        p.decibels_upper_threshold = default_decibels_upper_threshold;
    }
}

fn verifyParameters(p: *const Params) Error!void {
    if (p.high_pass_filter_cutoff < 2) {
        return error.InvalidHighPassFilterCutoff;
    }
    if (p.minimal_period < 2) {
        return error.InvalidMinimalPeriod;
    }
    if (p.maximal_period <= p.minimal_period) {
        return error.InvalidMaximalPeriod;
    }
    if (p.decibels_lower_threshold < 0) {
        return error.InvalidDecibelsLowerThreshold;
    }
    if (p.decibels_upper_threshold <= p.decibels_lower_threshold) {
        return error.InvalidDecibelsUpperThreshold;
    }
}

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


test "Corona default smoke" {
    var c = try Corona.init(testing.allocator, null);
    defer c.deinit();

    try testing.expectEqual(@as(usize, 49), c.getFilterBankLength());
    try testing.expectEqual(@as(i32, 12), c.getMinimalPeriodTimesTwo());
    try testing.expectEqual(@as(i32, 60), c.getMaximalPeriodTimesTwo());

    var primed_at: ?usize = null;
    for (0..testdata.test_input.len) |idx| {
        _ = c.update(testdata.test_input[idx]);
        if (c.isPrimed() and primed_at == null) {
            primed_at = idx;
        }
    }

    try testing.expect(primed_at != null);
    // Primes at sample index 11 (0-indexed), which is the 12th sample = MinimalPeriodTimesTwo.
    try testing.expectEqual(@as(usize, 11), primed_at.?);

    const dc = c.getDominantCycle();
    const dc_med = c.getDominantCycleMedian();
    const min_f: f64 = @floatFromInt(c.getMinimalPeriod());
    const max_f: f64 = @floatFromInt(c.getMaximalPeriod());

    try testing.expect(!math.isNan(dc) and !math.isInf(dc));
    try testing.expect(!math.isNan(dc_med) and !math.isInf(dc_med));
    try testing.expect(dc >= min_f and dc <= max_f);
    try testing.expect(dc_med >= min_f and dc_med <= max_f);

    const m = c.getMaximalAmplitudeSquared();
    try testing.expect(m > 0 and !math.isNan(m) and !math.isInf(m));
}

test "Corona NaN input is no-op" {
    var c = try Corona.init(testing.allocator, null);
    defer c.deinit();

    // Warm past priming.
    for (0..20) |idx| {
        _ = c.update(testdata.test_input[idx]);
    }
    try testing.expect(c.isPrimed());

    const dc_before = c.getDominantCycle();
    const dcm_before = c.getDominantCycleMedian();

    const got = c.update(math.nan(f64));
    try testing.expect(got); // preserves primed
    try testing.expectEqual(dc_before, c.getDominantCycle());
    try testing.expectEqual(dcm_before, c.getDominantCycleMedian());
}

test "Corona invalid params" {
    try testing.expectError(error.InvalidHighPassFilterCutoff, Corona.init(testing.allocator, .{ .high_pass_filter_cutoff = 1 }));
    try testing.expectError(error.InvalidMinimalPeriod, Corona.init(testing.allocator, .{ .minimal_period = 1 }));
    try testing.expectError(error.InvalidMaximalPeriod, Corona.init(testing.allocator, .{ .minimal_period = 10, .maximal_period = 10 }));
    try testing.expectError(error.InvalidDecibelsLowerThreshold, Corona.init(testing.allocator, .{ .decibels_lower_threshold = -1 }));
    try testing.expectError(error.InvalidDecibelsUpperThreshold, Corona.init(testing.allocator, .{ .decibels_lower_threshold = 6, .decibels_upper_threshold = 6 }));
}

test "Corona DC above min period" {
    var c = try Corona.init(testing.allocator, null);
    defer c.deinit();

    var saw_above_min = false;
    const min_f: f64 = @floatFromInt(c.getMinimalPeriod());

    for (0..testdata.test_input.len) |idx| {
        _ = c.update(testdata.test_input[idx]);
        if (c.isPrimed() and c.getDominantCycle() > min_f) {
            saw_above_min = true;
        }
    }

    try testing.expect(saw_above_min);
}
