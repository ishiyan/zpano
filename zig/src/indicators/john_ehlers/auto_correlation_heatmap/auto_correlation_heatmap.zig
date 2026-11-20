const std = @import("std");
const math = std.math;

const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const bar_component = entities.bar_component;
const quote_component = entities.quote_component;
const trade_component = entities.trade_component;
const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const descriptor_mod = @import("../../core/descriptor.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");
const output_metadata_mod = @import("../../core/output_metadata.zig");
const shape_mod = @import("../../core/outputs/shape.zig");
const heatmap_mod = @import("../../core/outputs/heatmap.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Heatmap = heatmap_mod.Heatmap;

const sqrt_two: f64 = math.sqrt2;
const max_correlation_bins: usize = 100;

/// Output indices for Ehlers' February 2025 Autocorrelation Heatmap indicator.
pub const AutoCorrelationHeatmapOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an `AutoCorrelationHeatmap`.
///
/// Defaults correspond to Ehlers' February 2025 article settings.
pub const AutoCorrelationHeatmapParams = struct {
    length: i32 = 20,
    smoothing_period: f64 = 20.0,
    max_lag: i32 = 99,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Params = AutoCorrelationHeatmapParams;

pub fn defaultParams() AutoCorrelationHeatmapParams {
    return .{};
}

pub const Error = error{
    InvalidLength,
    InvalidSmoothingPeriod,
    InvalidMaxLag,
    UnstableFilterCoefficients,
    InvalidSample,
    NumericOverflow,
    MnemonicTooLong,
    OutOfMemory,
};

fn normalizeZero(number: f64) f64 {
    return if (number == 0.0) 0.0 else number;
}

/// Builds the mnemonic for `AutoCorrelationHeatmap` into `buffer`.
pub fn mnemonic(buffer: []u8, params: AutoCorrelationHeatmapParams) Error![]const u8 {
    const bar_comp = params.bar_component orelse bar_component.default_bar_component;
    const quote_comp = params.quote_component orelse quote_component.default_quote_component;
    const trade_comp = params.trade_component orelse trade_component.default_trade_component;

    var triple_storage: [64]u8 = undefined;
    const comp_suffix = component_triple_mnemonic_mod.componentTripleMnemonic(
        &triple_storage,
        bar_comp,
        quote_comp,
        trade_comp,
    );

    return std.fmt.bufPrint(buffer, "ach({d}, {d}, {d}{s})", .{
        params.length,
        normalizeZero(params.smoothing_period),
        params.max_lag,
        comp_suffix,
    }) catch return error.MnemonicTooLong;
}

/// Snapshot of committed internal filter and correlation state.
pub const AutocorrelationState = struct {
    accepted_samples: usize,
    filtered: f64,
    correlations_storage: [max_correlation_bins]f64 = [_]f64{0.0} ** max_correlation_bins,
    bins_len: usize,

    pub fn correlationsSlice(self: *const AutocorrelationState) []const f64 {
        return self.correlations_storage[0..self.bins_len];
    }

    pub fn eql(self: *const AutocorrelationState, other: *const AutocorrelationState) bool {
        if (self.accepted_samples != other.accepted_samples) return false;
        if (self.bins_len != other.bins_len) return false;
        if (self.filtered != other.filtered) return false;
        return std.mem.eql(f64, self.correlationsSlice(), other.correlationsSlice());
    }
};

/// Validated numeric parameters resolved at construction time.
pub const ResolvedParams = struct {
    length: usize,
    smoothing_period: f64,
    max_lag: usize,
};

/// Ehlers' February 2025 Autocorrelation Heatmap indicator (ACH).
pub const AutoCorrelationHeatmap = struct {
    allocator: std.mem.Allocator,
    resolved: ResolvedParams,
    bins_len: usize,
    history_len: usize,
    prime_count: usize,
    coeff_c1: f64,
    coeff_c2: f64,
    coeff_c3: f64,
    history: []f64,
    next_history: []f64,
    correlations: [max_correlation_bins]f64 = [_]f64{0.0} ** max_correlation_bins,
    sample_count: usize = 0,
    price1: f64 = 0.0,
    price2: f64 = 0.0,
    filtered1: f64 = 0.0,
    filtered2: f64 = 0.0,
    last_output: Heatmap,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_storage: [output_metadata_mod.max_mnemonic_len]u8,
    mnemonic_len: usize,
    description_storage: [output_metadata_mod.max_description_len]u8,
    description_len: usize,

    pub fn init(
        allocator: std.mem.Allocator,
        params: AutoCorrelationHeatmapParams,
    ) Error!AutoCorrelationHeatmap {
        if (params.length < 2) {
            return error.InvalidLength;
        }
        if (!math.isFinite(params.smoothing_period) or params.smoothing_period < 1.0) {
            return error.InvalidSmoothingPeriod;
        }
        if (params.max_lag < 0 or params.max_lag > 99) {
            return error.InvalidMaxLag;
        }

        const angle = sqrt_two * math.pi / params.smoothing_period;
        const attenuation = @exp(-angle);
        const coeff_c2 = 2.0 * attenuation * @cos(angle);
        const coeff_c3 = -attenuation * attenuation;
        const coeff_c1 = (1.0 + coeff_c2 - coeff_c3) / 4.0;
        if (!math.isFinite(coeff_c1) or !math.isFinite(coeff_c2) or !math.isFinite(coeff_c3)) {
            return error.UnstableFilterCoefficients;
        }

        const window_length: usize = @intCast(params.length);
        const upper_lag: usize = @intCast(params.max_lag);
        const bins_len: usize = upper_lag + 1;
        const history_len = math.add(usize, window_length, upper_lag) catch return error.OutOfMemory;
        const prime_count: usize = @max(@as(usize, 4), history_len);

        const history = allocator.alloc(f64, history_len) catch return error.OutOfMemory;
        errdefer allocator.free(history);
        const next_history = allocator.alloc(f64, history_len) catch return error.OutOfMemory;
        errdefer allocator.free(next_history);

        @memset(history, 0.0);
        @memset(next_history, 0.0);

        const bar_comp = params.bar_component orelse bar_component.default_bar_component;
        const quote_comp = params.quote_component orelse quote_component.default_quote_component;
        const trade_comp = params.trade_component orelse trade_component.default_trade_component;

        var mnemonic_storage: [output_metadata_mod.max_mnemonic_len]u8 = undefined;
        const mn = try mnemonic(&mnemonic_storage, params);

        var description_storage: [output_metadata_mod.max_description_len]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Autocorrelation heatmap {s}", .{mn}) catch
            return error.MnemonicTooLong;

        return .{
            .allocator = allocator,
            .resolved = .{
                .length = window_length,
                .smoothing_period = params.smoothing_period,
                .max_lag = upper_lag,
            },
            .bins_len = bins_len,
            .history_len = history_len,
            .prime_count = prime_count,
            .coeff_c1 = coeff_c1,
            .coeff_c2 = coeff_c2,
            .coeff_c3 = coeff_c3,
            .history = history,
            .next_history = next_history,
            .last_output = Heatmap.empty(0, 0.0, @floatFromInt(upper_lag), 1.0),
            .bar_func = bar_component.componentValue(bar_comp),
            .quote_func = quote_component.componentValue(quote_comp),
            .trade_func = trade_component.componentValue(trade_comp),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mn.len,
            .description_storage = description_storage,
            .description_len = desc.len,
        };
    }

    pub fn initDefault(allocator: std.mem.Allocator) Error!AutoCorrelationHeatmap {
        return init(allocator, defaultParams());
    }

    pub fn deinit(self: *AutoCorrelationHeatmap) void {
        self.allocator.free(self.history);
        self.allocator.free(self.next_history);
    }

    pub fn reset(self: *AutoCorrelationHeatmap) void {
        self.sample_count = 0;
        self.price1 = 0.0;
        self.price2 = 0.0;
        self.filtered1 = 0.0;
        self.filtered2 = 0.0;
        @memset(self.history, 0.0);
        @memset(self.next_history, 0.0);
        @memset(&self.correlations, 0.0);
        self.last_output = Heatmap.empty(0, 0.0, @floatFromInt(self.resolved.max_lag), 1.0);
    }

    pub fn fixSlices(self: *AutoCorrelationHeatmap) void {
        _ = self;
    }

    pub fn mnemonicText(self: *const AutoCorrelationHeatmap) []const u8 {
        return self.mnemonic_storage[0..self.mnemonic_len];
    }

    fn descriptionText(self: *const AutoCorrelationHeatmap) []const u8 {
        return self.description_storage[0..self.description_len];
    }

    pub fn parameters(self: *const AutoCorrelationHeatmap) ResolvedParams {
        return self.resolved;
    }

    pub fn length(self: *const AutoCorrelationHeatmap) usize {
        return self.resolved.length;
    }

    pub fn smoothingPeriod(self: *const AutoCorrelationHeatmap) f64 {
        return self.resolved.smoothing_period;
    }

    pub fn maxLag(self: *const AutoCorrelationHeatmap) usize {
        return self.resolved.max_lag;
    }

    pub fn warmupBars(self: *const AutoCorrelationHeatmap) usize {
        return self.prime_count - 1;
    }

    pub fn samplesSeen(self: *const AutoCorrelationHeatmap) usize {
        return self.sample_count;
    }

    pub fn state(self: *const AutoCorrelationHeatmap) AutocorrelationState {
        var out_state = AutocorrelationState{
            .accepted_samples = self.sample_count,
            .filtered = self.filtered1,
            .bins_len = self.bins_len,
        };
        @memcpy(out_state.correlations_storage[0..self.bins_len], self.correlations[0..self.bins_len]);
        return out_state;
    }

    pub fn output(self: *const AutoCorrelationHeatmap) Heatmap {
        return self.last_output;
    }

    pub fn tryUpdate(self: *AutoCorrelationHeatmap, sample: f64, time: i64) Error!Heatmap {
        if (math.isInf(sample)) {
            return error.InvalidSample;
        }
        const max_lag_float: f64 = @floatFromInt(self.resolved.max_lag);
        if (math.isNan(sample)) {
            return Heatmap.empty(time, 0.0, max_lag_float, 1.0);
        }

        const filtered: f64 = if (self.sample_count < 3)
            sample
        else
            (1.0 - self.coeff_c1) * sample +
                (2.0 * self.coeff_c1 - self.coeff_c2) * self.price1 -
                (self.coeff_c1 + self.coeff_c3) * self.price2 +
                self.coeff_c2 * self.filtered1 +
                self.coeff_c3 * self.filtered2;
        if (!math.isFinite(filtered)) {
            return error.NumericOverflow;
        }

        self.next_history[0] = filtered;
        if (self.history_len > 1) {
            @memcpy(self.next_history[1..self.history_len], self.history[0 .. self.history_len - 1]);
        }

        var next_correlations: [max_correlation_bins]f64 = undefined;
        @memcpy(next_correlations[0..self.bins_len], self.correlations[0..self.bins_len]);
        const length_float: f64 = @floatFromInt(self.resolved.length);

        for (0..self.bins_len) |lag| {
            var sum_x: f64 = 0.0;
            var sum_y: f64 = 0.0;
            var sum_xx: f64 = 0.0;
            var sum_xy: f64 = 0.0;
            var sum_yy: f64 = 0.0;
            for (0..self.resolved.length) |offset| {
                const sample_x = self.next_history[offset];
                const sample_y = self.next_history[lag + offset];
                sum_x += sample_x;
                sum_y += sample_y;
                sum_xx += sample_x * sample_x;
                sum_xy += sample_x * sample_y;
                sum_yy += sample_y * sample_y;
            }
            const var_x = length_float * sum_xx - sum_x * sum_x;
            const var_y = length_float * sum_yy - sum_y * sum_y;
            if (!math.isFinite(sum_x) or !math.isFinite(sum_y) or
                !math.isFinite(sum_xx) or !math.isFinite(sum_xy) or !math.isFinite(sum_yy) or
                !math.isFinite(var_x) or !math.isFinite(var_y))
            {
                return error.NumericOverflow;
            }
            if (var_x > 0.0 and var_y > 0.0) {
                const product = var_x * var_y;
                const numerator = length_float * sum_xy - sum_x * sum_y;
                if (product <= 0.0 or !math.isFinite(product) or !math.isFinite(numerator)) {
                    return error.NumericOverflow;
                }
                const correlation = numerator / @sqrt(product);
                if (!math.isFinite(correlation)) {
                    return error.NumericOverflow;
                }
                next_correlations[lag] = correlation;
            }
        }

        self.price2 = self.price1;
        self.price1 = sample;
        self.filtered2 = self.filtered1;
        self.filtered1 = filtered;
        @memcpy(self.history, self.next_history);
        @memcpy(self.correlations[0..self.bins_len], next_correlations[0..self.bins_len]);
        self.sample_count += 1;

        const current_output: Heatmap = if (self.sample_count >= self.prime_count) blk: {
            const active_values = next_correlations[0..self.bins_len];
            var value_min = active_values[0];
            var value_max = active_values[0];
            for (active_values[1..]) |item| {
                if (item < value_min) value_min = item;
                if (item > value_max) value_max = item;
            }
            break :blk Heatmap.new(time, 0.0, max_lag_float, 1.0, value_min, value_max, active_values);
        } else Heatmap.empty(time, 0.0, max_lag_float, 1.0);

        self.last_output = current_output;
        return current_output;
    }

    pub fn update(self: *AutoCorrelationHeatmap, sample: f64, time: i64) Heatmap {
        return self.tryUpdate(sample, time) catch |err| {
            std.debug.panic("AutoCorrelationHeatmap.update failed: {s}", .{@errorName(err)});
        };
    }

    pub const updateSample = update;

    fn wrapEntity(out_value: Heatmap) OutputArray {
        var arr = OutputArray{};
        arr.append(.{ .heatmap = out_value });
        return arr;
    }

    pub fn updateScalar(self: *AutoCorrelationHeatmap, sample: *const Scalar) OutputArray {
        return wrapEntity(self.update(sample.value, sample.time));
    }

    pub fn updateBar(self: *AutoCorrelationHeatmap, sample: *const Bar) OutputArray {
        return wrapEntity(self.update(self.bar_func(sample.*), sample.time));
    }

    pub fn updateQuote(self: *AutoCorrelationHeatmap, sample: *const Quote) OutputArray {
        return wrapEntity(self.update(self.quote_func(sample.*), sample.time));
    }

    pub fn updateTrade(self: *AutoCorrelationHeatmap, sample: *const Trade) OutputArray {
        return wrapEntity(self.update(self.trade_func(sample.*), sample.time));
    }

    pub fn isPrimed(self: *const AutoCorrelationHeatmap) bool {
        return self.sample_count >= self.prime_count;
    }

    pub fn getMetadata(self: *const AutoCorrelationHeatmap, out: *Metadata) void {
        const mn = self.mnemonicText();
        const desc = self.descriptionText();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
        };
        build_metadata_mod.buildMetadata(out, .auto_correlation_heatmap, mn, desc, &texts);
    }

    pub fn indicator(self: *AutoCorrelationHeatmap) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(AutoCorrelationHeatmap);
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
const testing = std.testing;
const testdata = @import("testdata.zig");

const fixture_tolerance: f64 = 1e-13;

fn expectCloseTol(actual: f64, expected: f64, tolerance: f64) !void {
    if (math.isNan(expected)) {
        try testing.expect(math.isNan(actual));
        return;
    }
    try testing.expect(math.isFinite(actual));
    const bound = @max(tolerance, @abs(expected) * tolerance);
    try testing.expect(@abs(actual - expected) <= bound);
}

fn expectClose(actual: f64, expected: f64) !void {
    try expectCloseTol(actual, expected, fixture_tolerance);
}

fn verifyReferenceCase(
    comptime cols: usize,
    params: Params,
    input_close: *const [252]f64,
    expected_heatmap: *const [252][cols]f64,
    expected_min: *const [252]f64,
    expected_max: *const [252]f64,
) !void {
    var indicator_instance = try AutoCorrelationHeatmap.init(testing.allocator, params);
    defer indicator_instance.deinit();

    const length_usize: usize = @intCast(params.length);
    const max_lag_usize: usize = @intCast(params.max_lag);
    const warmup = @max(@as(usize, 4), length_usize + max_lag_usize) - 1;
    try testing.expectEqual(warmup, indicator_instance.warmupBars());
    try testing.expectEqual(cols, max_lag_usize + 1);

    for (0..252) |bar_index| {
        const timestamp: i64 = @intCast(bar_index + 1);
        const result = indicator_instance.update(input_close[bar_index], timestamp);
        try testing.expectEqual(@as(f64, 0.0), result.parameter_first);
        try testing.expectEqual(@as(f64, @floatFromInt(max_lag_usize)), result.parameter_last);
        try testing.expectEqual(@as(f64, 1.0), result.parameter_resolution);

        if (math.isNan(expected_min[bar_index])) {
            try testing.expect(!indicator_instance.isPrimed());
            try testing.expect(result.isEmpty());
            try testing.expect(math.isNan(result.value_min));
            try testing.expect(math.isNan(result.value_max));
            try testing.expect(math.isNan(expected_max[bar_index]));
            for (expected_heatmap[bar_index]) |cell| {
                try testing.expect(math.isNan(cell));
            }
        } else {
            try testing.expect(indicator_instance.isPrimed());
            try testing.expect(!result.isEmpty());
            const values = result.valuesSlice();
            try testing.expectEqual(cols, values.len);
            try expectClose(result.value_min, expected_min[bar_index]);
            try expectClose(result.value_max, expected_max[bar_index]);
            for (0..cols) |bin_index| {
                try expectClose(values[bin_index], expected_heatmap[bar_index][bin_index]);
            }
        }
    }
}

test "auto correlation heatmap all 12 reference cases" {
    const input_close = testdata.testInput();

    {
        const hm = testdata.expectedL20_P20_H99_HEATMAP();
        const mn = testdata.expectedL20_P20_H99_MIN();
        const mx = testdata.expectedL20_P20_H99_MAX();
        try verifyReferenceCase(100, .{ .length = 20, .smoothing_period = 20.0, .max_lag = 99 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL2_P20_H99_HEATMAP();
        const mn = testdata.expectedL2_P20_H99_MIN();
        const mx = testdata.expectedL2_P20_H99_MAX();
        try verifyReferenceCase(100, .{ .length = 2, .smoothing_period = 20.0, .max_lag = 99 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL3_P20_H99_HEATMAP();
        const mn = testdata.expectedL3_P20_H99_MIN();
        const mx = testdata.expectedL3_P20_H99_MAX();
        try verifyReferenceCase(100, .{ .length = 3, .smoothing_period = 20.0, .max_lag = 99 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL5_P20_H99_HEATMAP();
        const mn = testdata.expectedL5_P20_H99_MIN();
        const mx = testdata.expectedL5_P20_H99_MAX();
        try verifyReferenceCase(100, .{ .length = 5, .smoothing_period = 20.0, .max_lag = 99 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL40_P20_H99_HEATMAP();
        const mn = testdata.expectedL40_P20_H99_MIN();
        const mx = testdata.expectedL40_P20_H99_MAX();
        try verifyReferenceCase(100, .{ .length = 40, .smoothing_period = 20.0, .max_lag = 99 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL100_P20_H99_HEATMAP();
        const mn = testdata.expectedL100_P20_H99_MIN();
        const mx = testdata.expectedL100_P20_H99_MAX();
        try verifyReferenceCase(100, .{ .length = 100, .smoothing_period = 20.0, .max_lag = 99 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL20_P10_H99_HEATMAP();
        const mn = testdata.expectedL20_P10_H99_MIN();
        const mx = testdata.expectedL20_P10_H99_MAX();
        try verifyReferenceCase(100, .{ .length = 20, .smoothing_period = 10.0, .max_lag = 99 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL20_P10_5_H99_HEATMAP();
        const mn = testdata.expectedL20_P10_5_H99_MIN();
        const mx = testdata.expectedL20_P10_5_H99_MAX();
        try verifyReferenceCase(100, .{ .length = 20, .smoothing_period = 10.5, .max_lag = 99 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL20_P40_H99_HEATMAP();
        const mn = testdata.expectedL20_P40_H99_MIN();
        const mx = testdata.expectedL20_P40_H99_MAX();
        try verifyReferenceCase(100, .{ .length = 20, .smoothing_period = 40.0, .max_lag = 99 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL20_P20_H32_HEATMAP();
        const mn = testdata.expectedL20_P20_H32_MIN();
        const mx = testdata.expectedL20_P20_H32_MAX();
        try verifyReferenceCase(33, .{ .length = 20, .smoothing_period = 20.0, .max_lag = 32 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL2_P20_H32_HEATMAP();
        const mn = testdata.expectedL2_P20_H32_MIN();
        const mx = testdata.expectedL2_P20_H32_MAX();
        try verifyReferenceCase(33, .{ .length = 2, .smoothing_period = 20.0, .max_lag = 32 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL20_P20_H0_HEATMAP();
        const mn = testdata.expectedL20_P20_H0_MIN();
        const mx = testdata.expectedL20_P20_H0_MAX();
        try verifyReferenceCase(1, .{ .length = 20, .smoothing_period = 20.0, .max_lag = 0 }, &input_close, &hm, &mn, &mx);
    }
}

test "auto correlation heatmap edge cases mnemonics metadata and entities" {
    const input_close = testdata.testInput();
    const input_open = testdata.testOpen();
    const input_high = testdata.testHigh();
    const input_low = testdata.testLow();

    try testing.expectError(error.InvalidLength, AutoCorrelationHeatmap.init(testing.allocator, .{ .length = 1 }));
    try testing.expectError(error.InvalidSmoothingPeriod, AutoCorrelationHeatmap.init(testing.allocator, .{ .smoothing_period = 0.5 }));
    try testing.expectError(error.InvalidSmoothingPeriod, AutoCorrelationHeatmap.init(testing.allocator, .{ .smoothing_period = math.nan(f64) }));
    try testing.expectError(error.InvalidSmoothingPeriod, AutoCorrelationHeatmap.init(testing.allocator, .{ .smoothing_period = math.inf(f64) }));
    try testing.expectError(error.InvalidMaxLag, AutoCorrelationHeatmap.init(testing.allocator, .{ .max_lag = -1 }));
    try testing.expectError(error.InvalidMaxLag, AutoCorrelationHeatmap.init(testing.allocator, .{ .max_lag = 100 }));

    {
        var subject = try AutoCorrelationHeatmap.init(testing.allocator, .{ .length = 5, .smoothing_period = 10.0, .max_lag = 12 });
        defer subject.deinit();
        var clean = try AutoCorrelationHeatmap.init(testing.allocator, .{ .length = 5, .smoothing_period = 10.0, .max_lag = 12 });
        defer clean.deinit();

        for (0..50) |bar_index| {
            const before = subject.state();
            const skipped = try subject.tryUpdate(math.nan(f64), 100);
            try testing.expect(skipped.isEmpty());
            try testing.expect(subject.state().eql(&before));

            try testing.expectError(error.InvalidSample, subject.tryUpdate(math.inf(f64), 100));
            try testing.expectError(error.InvalidSample, subject.tryUpdate(-math.inf(f64), 100));
            try testing.expectError(error.NumericOverflow, subject.tryUpdate(1e308, 100));
            try testing.expect(subject.state().eql(&before));

            const actual = subject.update(input_close[bar_index], @intCast(bar_index));
            const expected = clean.update(input_close[bar_index], @intCast(bar_index));
            try testing.expectEqual(bar_index + 1 >= 17, subject.isPrimed());
            try testing.expectEqualSlices(f64, expected.valuesSlice(), actual.valuesSlice());
        }

        subject.reset();
        clean.reset();
        try testing.expectEqual(@as(usize, 0), subject.samplesSeen());
        try testing.expect(!subject.isPrimed());
        for (0..50) |bar_index| {
            const actual = subject.update(input_close[bar_index], @intCast(bar_index));
            const expected = clean.update(input_close[bar_index], @intCast(bar_index));
            try testing.expectEqualSlices(f64, expected.valuesSlice(), actual.valuesSlice());
        }
    }

    {
        var sine_ind = try AutoCorrelationHeatmap.init(testing.allocator, .{ .length = 20, .smoothing_period = 20.0, .max_lag = 40 });
        defer sine_ind.deinit();
        var column = Heatmap.empty(0, 0.0, 40.0, 1.0);
        for (0..1000) |step| {
            const step_float: f64 = @floatFromInt(step);
            column = sine_ind.update(@sin(2.0 * math.pi * step_float / 20.0), @intCast(step));
        }
        const values = column.valuesSlice();
        for (0..41) |lag| {
            const lag_float: f64 = @floatFromInt(lag);
            try expectCloseTol(values[lag], @cos(2.0 * math.pi * lag_float / 20.0), 1e-11);
        }
    }

    {
        var default_ind = try AutoCorrelationHeatmap.initDefault(testing.allocator);
        defer default_ind.deinit();
        var meta: Metadata = undefined;
        default_ind.getMetadata(&meta);
        try testing.expectEqual(Identifier.auto_correlation_heatmap, meta.identifier);
        try testing.expectEqualStrings("ach(20, 20, 99)", meta.mnemonic);
        try testing.expectEqualStrings("Autocorrelation heatmap ach(20, 20, 99)", meta.description);
        try testing.expectEqual(@as(usize, 1), meta.outputs_len);
        try testing.expectEqual(@intFromEnum(AutoCorrelationHeatmapOutput.value), meta.outputs_buf[0].kind);
        try testing.expectEqual(shape_mod.Shape.heatmap, meta.outputs_buf[0].shape);

        var mn_buffer: [64]u8 = undefined;
        try testing.expectEqualStrings(
            "ach(20, 10.5, 0, hl/2, a, v)",
            try mnemonic(&mn_buffer, .{
                .length = 20,
                .smoothing_period = 10.5,
                .max_lag = 0,
                .bar_component = .median,
                .quote_component = .ask,
                .trade_component = .volume,
            }),
        );
    }

    {
        var scalar_ind = try AutoCorrelationHeatmap.initDefault(testing.allocator);
        defer scalar_ind.deinit();
        var bar_ind = try AutoCorrelationHeatmap.initDefault(testing.allocator);
        defer bar_ind.deinit();
        var quote_ind = try AutoCorrelationHeatmap.initDefault(testing.allocator);
        defer quote_ind.deinit();
        var trade_ind = try AutoCorrelationHeatmap.initDefault(testing.allocator);
        defer trade_ind.deinit();
        var open_ind = try AutoCorrelationHeatmap.init(testing.allocator, .{ .bar_component = .open });
        defer open_ind.deinit();
        var open_ref = try AutoCorrelationHeatmap.initDefault(testing.allocator);
        defer open_ref.deinit();

        for (0..252) |bar_index| {
            const stamp: i64 = @intCast(bar_index + 1);
            const close_price = input_close[bar_index];
            const open_price = input_open[bar_index];
            const high_price = input_high[bar_index];
            const low_price = input_low[bar_index];

            const scalar_sample = Scalar{ .time = stamp, .value = close_price };
            const bar_sample = Bar{
                .time = stamp,
                .open = open_price,
                .high = high_price,
                .low = low_price,
                .close = close_price,
                .volume = 1000.0,
            };
            const quote_sample = Quote{
                .time = stamp,
                .bid_price = close_price,
                .ask_price = close_price,
                .bid_size = 10.0,
                .ask_size = 10.0,
            };
            const trade_sample = Trade{ .time = stamp, .price = close_price, .volume = 10.0 };

            const out_scalar = scalar_ind.updateScalar(&scalar_sample).values[0].heatmap;
            const out_bar = bar_ind.updateBar(&bar_sample).values[0].heatmap;
            const out_quote = quote_ind.updateQuote(&quote_sample).values[0].heatmap;
            const out_trade = trade_ind.updateTrade(&trade_sample).values[0].heatmap;
            try testing.expectEqualSlices(f64, out_scalar.valuesSlice(), out_bar.valuesSlice());
            try testing.expectEqualSlices(f64, out_scalar.valuesSlice(), out_quote.valuesSlice());
            try testing.expectEqualSlices(f64, out_scalar.valuesSlice(), out_trade.valuesSlice());

            const open_scalar = Scalar{ .time = stamp, .value = open_price };
            const out_open = open_ind.updateBar(&bar_sample).values[0].heatmap;
            const out_open_ref = open_ref.updateScalar(&open_scalar).values[0].heatmap;
            try testing.expectEqualSlices(f64, out_open.valuesSlice(), out_open_ref.valuesSlice());
        }
    }
}
