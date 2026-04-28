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

const test_input = [_]f64{
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

test "Corona default smoke" {
    var c = try Corona.init(testing.allocator, null);
    defer c.deinit();

    try testing.expectEqual(@as(usize, 49), c.getFilterBankLength());
    try testing.expectEqual(@as(i32, 12), c.getMinimalPeriodTimesTwo());
    try testing.expectEqual(@as(i32, 60), c.getMaximalPeriodTimesTwo());

    var primed_at: ?usize = null;
    for (0..test_input.len) |idx| {
        _ = c.update(test_input[idx]);
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
        _ = c.update(test_input[idx]);
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

    for (0..test_input.len) |idx| {
        _ = c.update(test_input[idx]);
        if (c.isPrimed() and c.getDominantCycle() > min_f) {
            saw_above_min = true;
        }
    }

    try testing.expect(saw_above_min);
}
