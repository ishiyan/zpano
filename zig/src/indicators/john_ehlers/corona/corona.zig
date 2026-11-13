const std = @import("std");
const math = std.math;
const heatmap_mod = @import("../../core/outputs/heatmap.zig");
const Heatmap = heatmap_mod.Heatmap;

pub const first_bank_sample: usize = 13;
pub const two_pi: f64 = 2.0 * math.pi;
pub const sqrt_three_over_two: f64 = @sqrt(3.0) / 2.0;
pub const max_bank_length: usize = heatmap_mod.max_heatmap_values;

const fir_weights = [6]f64{ 1.0, 2.0, 3.0, 3.0, 2.0, 1.0 };

pub const Error = error{
    InvalidHighPassFilterCutoff,
    InvalidMinimalPeriod,
    InvalidMaximalPeriod,
    InvalidDecibelsLowerThreshold,
    InvalidDecibelsUpperThreshold,
    InvalidRasterLength,
    InvalidMaxRasterValue,
    InvalidMinParameterValue,
    InvalidMaxParameterValue,
    InvalidLeadWindow,
    InvalidPositionWindow,
    InvalidLookbackOffset,
    InvalidBarRange,
    SnapshotMismatch,
    SnapshotOutOfOrder,
    NonFiniteInput,
    NumericOverflow,
    MnemonicTooLong,
    OutOfMemory,
};

pub fn validateFinite(value: f64) bool {
    return !math.isNan(value) and !math.isInf(value);
}

pub fn validPrice(price: f64) Error!bool {
    if (math.isNan(price)) return false;
    if (math.isInf(price)) return error.NonFiniteInput;
    return true;
}

pub fn checked(value: f64) Error!f64 {
    if (!validateFinite(value)) return error.NumericOverflow;
    return value;
}

pub fn clip(value: f64, lower: f64, upper: f64) f64 {
    return @max(lower, @min(upper, value));
}

pub fn normalizeZero(value: f64) f64 {
    return if (value == 0.0) 0.0 else value;
}

pub fn rollingAppend(window: []f64, current_len: *usize, value: f64) struct { minimum: f64, maximum: f64 } {
    if (current_len.* < window.len) {
        window[current_len.*] = value;
        current_len.* += 1;
    } else {
        std.mem.copyForwards(f64, window[0 .. window.len - 1], window[1..window.len]);
        window[window.len - 1] = value;
    }
    const active = window[0..current_len.*];
    var minimum = active[0];
    var maximum = active[0];
    for (active[1..]) |item| {
        if (item < minimum) minimum = item;
        if (item > maximum) maximum = item;
    }
    return .{ .minimum = minimum, .maximum = maximum };
}

/// Parameters for creating a Corona spectral analysis engine.
pub const Params = struct {
    high_pass_filter_cutoff: i32 = 30,
    minimal_period: i32 = 6,
    maximal_period: i32 = 30,
    decibels_lower_threshold: f64 = 6.0,
    decibels_upper_threshold: f64 = 20.0,

    pub fn eql(self: Params, other: Params) bool {
        return self.high_pass_filter_cutoff == other.high_pass_filter_cutoff and
            self.minimal_period == other.minimal_period and
            self.maximal_period == other.maximal_period and
            self.decibels_lower_threshold == other.decibels_lower_threshold and
            self.decibels_upper_threshold == other.decibels_upper_threshold;
    }
};

pub fn defaultParams() Params {
    return .{};
}

/// Per-bin state of a single half-period bandpass channel.
pub const Filter = struct {
    in_phase1: f64 = 0.0,
    in_phase2: f64 = 0.0,
    quadrature1: f64 = 0.0,
    quadrature2: f64 = 0.0,
    real1: f64 = 0.0,
    real2: f64 = 0.0,
    imaginary1: f64 = 0.0,
    imaginary2: f64 = 0.0,
    power: f64 = 0.0,
    decibels: f64 = 0.0,
};

/// Value-owned 18-field snapshot produced by one Corona engine update.
pub const CoronaSnapshot = struct {
    parameters: Params,
    sample_count: usize,
    accepted: bool,
    primed: bool,
    price: f64,
    time: i64,
    high_pass: f64,
    filtered: f64,
    bandwidth: f64,
    dominant_cycle: f64,
    dominant_cycle_median: f64,
    maximum_power: f64,
    normalization_power: f64,
    selected_bins: usize,
    has_cycle_estimate: bool,
    bank_length: usize,
    power_storage: [max_bank_length]f64 = [_]f64{0.0} ** max_bank_length,
    relative_power_storage: [max_bank_length]f64 = [_]f64{0.0} ** max_bank_length,
    decibels_storage: [max_bank_length]f64 = [_]f64{0.0} ** max_bank_length,

    pub fn powerSlice(self: *const CoronaSnapshot) []const f64 {
        return self.power_storage[0..self.bank_length];
    }

    pub fn relativePowerSlice(self: *const CoronaSnapshot) []const f64 {
        return self.relative_power_storage[0..self.bank_length];
    }

    pub fn decibelsSlice(self: *const CoronaSnapshot) []const f64 {
        return self.decibels_storage[0..self.bank_length];
    }
};

pub const Snapshot = CoronaSnapshot;

pub fn acceptSnapshot(
    snapshot: CoronaSnapshot,
    expected_parameters: Params,
    expected_count: usize,
) Error!bool {
    if (!snapshot.parameters.eql(expected_parameters)) {
        return error.SnapshotMismatch;
    }
    if (!snapshot.accepted) {
        return false;
    }
    if (snapshot.sample_count != expected_count) {
        return error.SnapshotOutOfOrder;
    }
    return true;
}

/// Shared Corona spectral-analysis engine consumed by all Corona indicators.
pub const Corona = struct {
    allocator: std.mem.Allocator,
    parameters: Params,
    minimal_period: i32,
    maximal_period: i32,
    minimal_period_times_two: i32,
    maximal_period_times_two: i32,
    filter_bank_length: usize,
    pole: f64,
    periods: []f64,
    betas: []f64,
    bank: []Filter,
    next_bank: []Filter,
    sample_count: usize,
    price_previous: f64,
    hp_history: [6]f64,
    filtered_previous: f64,
    dc: f64,
    dc_history: [5]f64,
    last_snapshot: ?CoronaSnapshot,

    pub fn init(allocator: std.mem.Allocator, maybe_params: ?Params) Error!Corona {
        const cfg = maybe_params orelse Params{};
        if (cfg.high_pass_filter_cutoff < 5) return error.InvalidHighPassFilterCutoff;
        if (cfg.minimal_period < 6) return error.InvalidMinimalPeriod;
        if (cfg.maximal_period <= cfg.minimal_period) return error.InvalidMaximalPeriod;
        if (!validateFinite(cfg.decibels_lower_threshold) or cfg.decibels_lower_threshold < 0.0) {
            return error.InvalidDecibelsLowerThreshold;
        }
        if (!validateFinite(cfg.decibels_upper_threshold) or
            cfg.decibels_upper_threshold <= 0.0 or
            cfg.decibels_upper_threshold <= cfg.decibels_lower_threshold)
        {
            return error.InvalidDecibelsUpperThreshold;
        }

        const angle = two_pi / @as(f64, @floatFromInt(cfg.high_pass_filter_cutoff));
        const pole = (1.0 - @sin(angle)) / @cos(angle);
        if (!validateFinite(pole) or !(0.0 < pole and pole < 1.0)) {
            return error.InvalidHighPassFilterCutoff;
        }

        const min_p2 = 2 * cfg.minimal_period;
        const max_p2 = 2 * cfg.maximal_period;
        const bank_len: usize = @intCast(max_p2 - min_p2 + 1);
        if (bank_len > max_bank_length) {
            return error.InvalidMaximalPeriod;
        }

        const periods = allocator.alloc(f64, bank_len) catch return error.OutOfMemory;
        errdefer allocator.free(periods);
        const betas = allocator.alloc(f64, bank_len) catch return error.OutOfMemory;
        errdefer allocator.free(betas);
        const bank = allocator.alloc(Filter, bank_len) catch return error.OutOfMemory;
        errdefer allocator.free(bank);
        const next_bank = allocator.alloc(Filter, bank_len) catch return error.OutOfMemory;
        errdefer allocator.free(next_bank);

        for (0..bank_len) |index| {
            const half_period = min_p2 + @as(i32, @intCast(index));
            const period = @as(f64, @floatFromInt(half_period)) / 2.0;
            periods[index] = period;
            betas[index] = @cos(two_pi / period);
        }
        @memset(bank, Filter{});
        @memset(next_bank, Filter{});

        return .{
            .allocator = allocator,
            .parameters = cfg,
            .minimal_period = cfg.minimal_period,
            .maximal_period = cfg.maximal_period,
            .minimal_period_times_two = min_p2,
            .maximal_period_times_two = max_p2,
            .filter_bank_length = bank_len,
            .pole = pole,
            .periods = periods,
            .betas = betas,
            .bank = bank,
            .next_bank = next_bank,
            .sample_count = 0,
            .price_previous = 0.0,
            .hp_history = [_]f64{0.0} ** 6,
            .filtered_previous = 0.0,
            .dc = 0.0,
            .dc_history = [_]f64{0.0} ** 5,
            .last_snapshot = null,
        };
    }

    pub fn deinit(self: *Corona) void {
        self.allocator.free(self.periods);
        self.allocator.free(self.betas);
        self.allocator.free(self.bank);
        self.allocator.free(self.next_bank);
    }

    pub fn reset(self: *Corona) void {
        self.sample_count = 0;
        self.price_previous = 0.0;
        self.hp_history = [_]f64{0.0} ** 6;
        self.filtered_previous = 0.0;
        @memset(self.bank, Filter{});
        @memset(self.next_bank, Filter{});
        self.dc = 0.0;
        self.dc_history = [_]f64{0.0} ** 5;
        self.last_snapshot = null;
    }

    pub fn isPrimed(self: *const Corona) bool {
        return self.sample_count >= first_bank_sample;
    }

    pub fn getFilterBankLength(self: *const Corona) usize {
        return self.filter_bank_length;
    }

    pub fn getDominantCycle(self: *const Corona) f64 {
        return self.dc;
    }

    pub fn getDominantCycleMedian(self: *const Corona) f64 {
        if (self.last_snapshot) |snap| return snap.dominant_cycle_median;
        return @floatFromInt(self.minimal_period);
    }

    pub fn tryUpdate(self: *Corona, price: f64, time: i64) Error!CoronaSnapshot {
        const ok = try validPrice(price);
        if (!ok) {
            if (self.last_snapshot) |previous_snap| {
                var skipped = previous_snap;
                skipped.accepted = false;
                skipped.price = price;
                skipped.time = time;
                return skipped;
            }
            return CoronaSnapshot{
                .parameters = self.parameters,
                .sample_count = 0,
                .accepted = false,
                .primed = false,
                .price = price,
                .time = time,
                .high_pass = 0.0,
                .filtered = 0.0,
                .bandwidth = 0.5,
                .dominant_cycle = 0.0,
                .dominant_cycle_median = @floatFromInt(self.minimal_period),
                .maximum_power = 0.0,
                .normalization_power = 0.01,
                .selected_bins = 0,
                .has_cycle_estimate = false,
                .bank_length = self.filter_bank_length,
            };
        }

        const count = self.sample_count + 1;
        const previous_price = if (count == 1) price else self.price_previous;
        const hp = try checked(0.5 * (1.0 + self.pole) * (price - previous_price) + self.pole * self.hp_history[5]);
        var hp_history: [6]f64 = undefined;
        @memcpy(hp_history[0..5], self.hp_history[1..6]);
        hp_history[5] = hp;

        var fir_sum: f64 = 0.0;
        for (0..6) |index| {
            fir_sum += hp_history[index] * fir_weights[index];
        }
        var filtered = try checked(fir_sum / 12.0);
        if (count < 7) {
            filtered = try checked(price - previous_price);
        }
        if (count == 1) {
            filtered = 0.0;
        }
        const bandwidth = @max(0.1, 0.5 - 0.015 * @as(f64, @floatFromInt(count)));

        for (0..self.filter_bank_length) |index| {
            const old = self.bank[index];
            if (count < first_bank_sample) {
                self.next_bank[index] = old;
                continue;
            }
            const period = self.periods[index];
            const beta = self.betas[index];
            const gamma = 1.0 / @cos(2.0 * two_pi * bandwidth / period);
            const alpha = gamma - @sqrt(gamma * gamma - 1.0);
            const in_phase = filtered;
            const quadrature = try checked((period / two_pi) * (filtered - self.filtered_previous));
            const real_part = try checked(
                0.5 * (1.0 - alpha) * (in_phase - old.in_phase2) +
                    beta * (1.0 + alpha) * old.real1 -
                    alpha * old.real2,
            );
            const imaginary_part = try checked(
                0.5 * (1.0 - alpha) * (quadrature - old.quadrature2) +
                    beta * (1.0 + alpha) * old.imaginary1 -
                    alpha * old.imaginary2,
            );
            const power = try checked(real_part * real_part + imaginary_part * imaginary_part);
            self.next_bank[index] = .{
                .in_phase1 = in_phase,
                .in_phase2 = old.in_phase1,
                .quadrature1 = quadrature,
                .quadrature2 = old.quadrature1,
                .real1 = real_part,
                .real2 = old.real1,
                .imaginary1 = imaginary_part,
                .imaginary2 = old.imaginary1,
                .power = power,
                .decibels = old.decibels,
            };
        }

        var snapshot = CoronaSnapshot{
            .parameters = self.parameters,
            .sample_count = count,
            .accepted = true,
            .primed = count >= first_bank_sample,
            .price = price,
            .time = time,
            .high_pass = hp,
            .filtered = filtered,
            .bandwidth = bandwidth,
            .dominant_cycle = self.dc,
            .dominant_cycle_median = @floatFromInt(self.minimal_period),
            .maximum_power = 0.0,
            .normalization_power = 0.01,
            .selected_bins = 0,
            .has_cycle_estimate = false,
            .bank_length = self.filter_bank_length,
        };

        var maximum_power: f64 = self.next_bank[0].power;
        for (self.next_bank[1..self.filter_bank_length]) |channel| {
            if (channel.power > maximum_power) maximum_power = channel.power;
        }
        const first_power = if (self.next_bank[0].power == 0.0) 0.01 else self.next_bank[0].power;
        const normalization_power = @max(first_power, maximum_power);
        snapshot.maximum_power = maximum_power;
        snapshot.normalization_power = normalization_power;

        const upper_threshold = self.parameters.decibels_upper_threshold;
        const lower_threshold = self.parameters.decibels_lower_threshold;
        var numerator: f64 = 0.0;
        var denominator: f64 = 0.0;
        var selected_bins: usize = 0;

        for (0..self.filter_bank_length) |index| {
            var channel = &self.next_bank[index];
            const relative_power = channel.power / normalization_power;
            var raw_decibels = channel.decibels;
            if (relative_power > 0.0) {
                raw_decibels = 10.0 * @log10((1.0 - 0.99 * relative_power) / 0.01);
            }
            const smoothed = 0.33 * raw_decibels + 0.67 * channel.decibels;
            channel.decibels = @min(smoothed, upper_threshold);
            if (channel.decibels <= lower_threshold) {
                const weight = upper_threshold - channel.decibels;
                numerator += self.periods[index] * weight;
                denominator += weight;
                selected_bins += 1;
            }
            snapshot.power_storage[index] = channel.power;
            snapshot.relative_power_storage[index] = relative_power;
            snapshot.decibels_storage[index] = channel.decibels;
        }

        var dc = self.dc;
        if (denominator > 0.0) {
            dc = numerator / denominator;
        }
        var dc_history: [5]f64 = undefined;
        @memcpy(dc_history[0..4], self.dc_history[1..5]);
        dc_history[4] = dc;
        var sorted_dc = dc_history;
        std.mem.sort(f64, &sorted_dc, {}, std.sort.asc(f64));
        const dc_median = @max(@as(f64, @floatFromInt(self.minimal_period)), sorted_dc[2]);

        snapshot.dominant_cycle = dc;
        snapshot.dominant_cycle_median = dc_median;
        snapshot.selected_bins = selected_bins;
        snapshot.has_cycle_estimate = maximum_power > 0.0 and selected_bins > 0;

        self.sample_count = count;
        self.price_previous = price;
        self.hp_history = hp_history;
        self.filtered_previous = filtered;
        @memcpy(self.bank, self.next_bank);
        self.dc = dc;
        self.dc_history = dc_history;
        self.last_snapshot = snapshot;
        return snapshot;
    }

    pub fn update(self: *Corona, price: f64, time: i64) CoronaSnapshot {
        return self.tryUpdate(price, time) catch |err| {
            std.debug.panic("Corona.update failed: {s}", .{@errorName(err)});
        };
    }
};

pub const AdaptiveCycle = struct {
    band_pass: f64,
    quadrature: f64,
    amplitude: f64,
    lead60: f64,
};

pub const AdaptiveBandPass = struct {
    count: usize = 0,
    prices: [2]f64 = .{ 0.0, 0.0 },
    bp1: f64 = 0.0,
    bp2: f64 = 0.0,

    pub fn reset(self: *AdaptiveBandPass) void {
        self.count = 0;
        self.prices = .{ 0.0, 0.0 };
        self.bp1 = 0.0;
        self.bp2 = 0.0;
    }

    pub fn update(self: *AdaptiveBandPass, price: f64, period: f64) Error!AdaptiveCycle {
        self.count += 1;
        if (self.count <= 2) {
            self.prices[self.count - 1] = price;
            return .{ .band_pass = 0.0, .quadrature = 0.0, .amplitude = 0.0, .lead60 = 0.0 };
        }
        const omega = two_pi / period;
        const beta = @cos(omega);
        const gamma = 1.0 / @cos(2.0 * omega * 0.1);
        const alpha = gamma - @sqrt(gamma * gamma - 1.0);
        const bp = try checked(
            0.5 * (1.0 - alpha) * (price - self.prices[0]) +
                beta * (1.0 + alpha) * self.bp1 -
                alpha * self.bp2,
        );
        const quadrature = try checked((bp - self.bp1) / omega);
        const amplitude = try checked(@sqrt(bp * bp + quadrature * quadrature));
        const lead60 = try checked(0.5 * bp + sqrt_three_over_two * quadrature);
        self.prices = .{ self.prices[1], price };
        self.bp2 = self.bp1;
        self.bp1 = bp;
        return .{
            .band_pass = bp,
            .quadrature = quadrature,
            .amplitude = amplitude,
            .lead60 = lead60,
        };
    }
};

pub const RasterKind = enum {
    swing,
    trend,
    snr,
};

pub const Raster = struct {
    allocator: std.mem.Allocator,
    length: usize,
    lower: f64,
    upper: f64,
    ceiling: f64,
    step: f64,
    first: f64,
    resolution: f64,
    values: []f64,

    pub fn init(
        allocator: std.mem.Allocator,
        length_raw: i32,
        lower: f64,
        upper: f64,
        ceiling: f64,
    ) Error!Raster {
        if (length_raw < 2 or length_raw > @as(i32, @intCast(max_bank_length))) {
            return error.InvalidRasterLength;
        }
        if (!validateFinite(lower)) return error.InvalidMinParameterValue;
        if (!validateFinite(upper) or upper <= lower) return error.InvalidMaxParameterValue;
        if (!validateFinite(ceiling) or ceiling <= 0.0) return error.InvalidMaxRasterValue;

        const length: usize = @intCast(length_raw);
        const step = (upper - lower) / @as(f64, @floatFromInt(length));
        const first = lower + step;
        const resolution = 1.0 / step;
        const values = allocator.alloc(f64, length) catch return error.OutOfMemory;
        @memset(values, 0.0);

        return .{
            .allocator = allocator,
            .length = length,
            .lower = lower,
            .upper = upper,
            .ceiling = ceiling,
            .step = step,
            .first = first,
            .resolution = resolution,
            .values = values,
        };
    }

    pub fn deinit(self: *Raster) void {
        self.allocator.free(self.values);
    }

    pub fn reset(self: *Raster) void {
        @memset(self.values, 0.0);
    }

    pub fn empty(self: *const Raster, time: i64) Heatmap {
        return Heatmap.empty(time, self.first, self.upper, self.resolution);
    }

    pub fn column(self: *const Raster, time: i64) Heatmap {
        var minimum = self.values[0];
        var maximum = self.values[0];
        for (self.values[1..]) |item| {
            if (item < minimum) minimum = item;
            if (item > maximum) maximum = item;
        }
        return Heatmap.new(time, self.first, self.upper, self.resolution, minimum, maximum, self.values);
    }

    pub fn update(self: *Raster, position: f64, width: f64, kind: RasterKind, suppress: bool) void {
        const length_f = @as(f64, @floatFromInt(self.length));
        const marker: isize = @intFromFloat(@floor(length_f * position + 0.5));
        const RasterCoefficients = struct { exponent: f64, scale: f64, history_scale: f64 };
        const coeffs: RasterCoefficients = switch (kind) {
            .swing => .{ .exponent = 0.95, .scale = 0.5, .history_scale = 0.5 },
            .trend => .{ .exponent = 0.85, .scale = 0.8, .history_scale = 0.2 },
            .snr => .{ .exponent = 0.8, .scale = 0.5, .history_scale = 1.0 },
        };
        for (0..self.length) |index| {
            const old = self.values[index];
            const cell_index: isize = @as(isize, @intCast(index)) + 1;
            const cell_f = @as(f64, @floatFromInt(cell_index));
            var value: f64 = undefined;
            if (suppress) {
                value = self.ceiling;
            } else if (cell_index == marker) {
                value = 0.5 * old;
            } else if (width <= 0.0) {
                value = self.ceiling;
            } else {
                const distance = @abs(self.ceiling * position - self.ceiling * cell_f / length_f) / width;
                if (kind == .snr and cell_index > marker and distance <= 1.0) {
                    value = self.ceiling;
                } else {
                    value = coeffs.scale * (math.pow(f64, distance, coeffs.exponent) + coeffs.history_scale * old);
                }
            }
            self.values[index] = clip(value, 0.0, self.ceiling);
        }
    }
};

// --- Engine Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");

test "Corona engine startup and empty selection holds" {
    var engine = try Corona.init(testing.allocator, defaultParams());
    defer engine.deinit();

    try testing.expectEqual(@as(usize, 49), engine.filter_bank_length);
    try testing.expectEqual(@as(i32, 12), engine.minimal_period_times_two);
    try testing.expectEqual(@as(i32, 60), engine.maximal_period_times_two);

    const prices = testdata.testInput();
    var previous_dc: f64 = 0.0;
    var holds: usize = 0;
    var final_dc: f64 = 0.0;

    for (prices, 0..) |price, index| {
        const snap = engine.update(price, @intCast(index));
        if (index < 12) {
            try testing.expect(!snap.primed);
            try testing.expectEqual(@as(f64, 0.0), snap.maximum_power);
        } else if (index == 12) {
            try testing.expect(snap.primed);
        }
        if (index > 0 and snap.selected_bins == 0) {
            holds += 1;
            try testing.expectEqual(previous_dc, snap.dominant_cycle);
            try testing.expect(!snap.has_cycle_estimate);
        }
        previous_dc = snap.dominant_cycle;
        final_dc = snap.dominant_cycle;
    }
    try testing.expect(holds > 0);
    try testing.expect(@abs(final_dc - 11.5) <= 5e-12);
}

test "Corona engine validation, explicit zero lower threshold, overflow guard, and reset" {
    try testing.expectError(error.InvalidMinimalPeriod, Corona.init(testing.allocator, .{ .minimal_period = 2 }));
    try testing.expectError(error.InvalidMinimalPeriod, Corona.init(testing.allocator, .{ .minimal_period = 0 }));
    try testing.expectError(error.InvalidMaximalPeriod, Corona.init(testing.allocator, .{ .minimal_period = 6, .maximal_period = 6 }));
    try testing.expectError(error.InvalidHighPassFilterCutoff, Corona.init(testing.allocator, .{ .high_pass_filter_cutoff = 4 }));
    try testing.expectError(error.InvalidHighPassFilterCutoff, Corona.init(testing.allocator, .{ .high_pass_filter_cutoff = 0 }));
    try testing.expectError(error.InvalidDecibelsLowerThreshold, Corona.init(testing.allocator, .{ .decibels_lower_threshold = -1.0 }));
    try testing.expectError(error.InvalidDecibelsUpperThreshold, Corona.init(testing.allocator, .{
        .decibels_lower_threshold = 20.0,
        .decibels_upper_threshold = 20.0,
    }));

    var zero_lower = try Corona.init(testing.allocator, .{ .decibels_lower_threshold = 0.0 });
    defer zero_lower.deinit();
    try testing.expectEqual(@as(f64, 0.0), zero_lower.parameters.decibels_lower_threshold);

    var engine = try Corona.init(testing.allocator, null);
    defer engine.deinit();
    const prices = testdata.testInput();
    for (prices[0..30], 0..) |price, index| {
        _ = engine.update(price, @intCast(index));
    }
    try testing.expectError(error.NonFiniteInput, engine.tryUpdate(math.inf(f64), 30));
    try testing.expectError(error.NumericOverflow, engine.tryUpdate(1e308, 30));
    try testing.expectEqual(@as(usize, 30), engine.sample_count);

    engine.reset();
    try testing.expectEqual(@as(usize, 0), engine.sample_count);
    try testing.expect(engine.last_snapshot == null);
}

test "AdaptiveBandPass quadrature and lead60 and Raster geometry" {
    var bandpass = AdaptiveBandPass{};
    var previous_bp: f64 = 0.0;
    for (0..100) |index| {
        const sample = 100.0 + @sin(two_pi * @as(f64, @floatFromInt(index)) / 20.0);
        const cycle = try bandpass.update(sample, 20.0);
        if (index >= 2) {
            try testing.expect(@abs(cycle.quadrature - (cycle.band_pass - previous_bp) / (two_pi / 20.0)) <= 1e-14);
            try testing.expect(@abs(cycle.lead60 - (0.5 * cycle.band_pass + sqrt_three_over_two * cycle.quadrature)) <= 1e-14);
        }
        previous_bp = cycle.band_pass;
    }

    var raster = try Raster.init(testing.allocator, 50, -5.0, 5.0, 20.0);
    defer raster.deinit();
    raster.update(1.0, 0.1, .swing, false);
    const column = raster.column(0);
    try testing.expect(@abs(column.parameter_first - (-4.8)) <= 1e-14);
    try testing.expect(@abs(column.parameter_last - 5.0) <= 1e-14);
    try testing.expect(@abs(column.parameter_resolution - 5.0) <= 1e-14);
    try testing.expectEqual(@as(f64, 0.0), column.valuesSlice()[49]);
    try testing.expect(column.valuesSlice()[0] > 0.0);

    @memset(raster.values, 8.0);
    raster.update(0.01, 0.1, .swing, false);
    try testing.expectEqual(@as(f64, 4.0), raster.values[0]);
}
