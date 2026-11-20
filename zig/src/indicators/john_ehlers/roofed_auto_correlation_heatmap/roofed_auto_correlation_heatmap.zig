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
const role_mod = @import("../../core/role.zig");
const shape_mod = @import("../../core/outputs/shape.zig");
const heatmap_mod = @import("../../core/outputs/heatmap.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Heatmap = heatmap_mod.Heatmap;

const two_pi: f64 = 2.0 * math.pi;
const sqrt_two: f64 = math.sqrt2;
const max_state_correlations: usize = 512;

/// Output indices for Ehlers' 2013 Roofed Autocorrelation Heatmap indicator.
pub const RoofedAutoCorrelationHeatmapOutput = enum(u8) {
    value = 1,
};

/// Parameters to create a `RoofedAutoCorrelationHeatmap`.
///
/// Defaults correspond to Ehlers' 2013 Book Listing 8-2 settings.
pub const RoofedAutoCorrelationHeatmapParams = struct {
    min_lag: i32 = 3,
    max_lag: i32 = 48,
    averaging_length: i32 = 0,
    highpass_period: f64 = 48.0,
    smoothing_period: f64 = 10.0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Params = RoofedAutoCorrelationHeatmapParams;

pub fn defaultParams() RoofedAutoCorrelationHeatmapParams {
    return .{};
}

pub const Error = error{
    InvalidMinLag,
    InvalidMaxLag,
    LagRangeTooLarge,
    InvalidAveragingLength,
    InvalidHighpassPeriod,
    InvalidSmoothingPeriod,
    UnstableFilterCoefficients,
    InvalidSample,
    NumericOverflow,
    MnemonicTooLong,
    OutOfMemory,
};

fn normalizeZero(number: f64) f64 {
    return if (number == 0.0) 0.0 else number;
}

/// Builds the mnemonic for `RoofedAutoCorrelationHeatmap` into `buffer`.
pub fn mnemonic(buffer: []u8, params: RoofedAutoCorrelationHeatmapParams) Error![]const u8 {
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

    return std.fmt.bufPrint(buffer, "rach({d}, {d}, {d}, {d}, {d}{s})", .{
        params.min_lag,
        params.max_lag,
        params.averaging_length,
        normalizeZero(params.highpass_period),
        normalizeZero(params.smoothing_period),
        comp_suffix,
    }) catch return error.MnemonicTooLong;
}

/// Snapshot of committed internal roofing and correlation state.
pub const AutocorrelationState = struct {
    high_pass: f64,
    filtered: f64,
    correlations_storage: [max_state_correlations]f64 = [_]f64{0.0} ** max_state_correlations,
    bins_len: usize,

    pub fn correlationsSlice(self: *const AutocorrelationState) []const f64 {
        return self.correlations_storage[0..self.bins_len];
    }

    pub fn eql(self: *const AutocorrelationState, other: *const AutocorrelationState) bool {
        if (self.bins_len != other.bins_len) return false;
        if (self.high_pass != other.high_pass) return false;
        if (self.filtered != other.filtered) return false;
        return std.mem.eql(f64, self.correlationsSlice(), other.correlationsSlice());
    }
};

/// Validated numeric parameters resolved at construction time.
pub const ResolvedParams = struct {
    min_lag: usize,
    max_lag: usize,
    averaging_length: usize,
    highpass_period: f64,
    smoothing_period: f64,
};

const RoofFilter = struct {
    hp0: f64,
    hp1: f64,
    hp2: f64,
    ss1: f64,
    ss2: f64,
    ss3: f64,
    x1: f64 = 0.0,
    x2: f64 = 0.0,
    h1: f64 = 0.0,
    h2: f64 = 0.0,
    f1: f64 = 0.0,
    f2: f64 = 0.0,

    fn init(highpass_period: f64, smoothing_period: f64) Error!RoofFilter {
        if (!math.isFinite(highpass_period) or highpass_period < 5.0) {
            return error.InvalidHighpassPeriod;
        }
        if (!math.isFinite(smoothing_period) or smoothing_period < 2.0) {
            return error.InvalidSmoothingPeriod;
        }

        const angle = two_pi / (sqrt_two * highpass_period);
        const cosine = @cos(angle);
        const alpha = (cosine + @sin(angle) - 1.0) / cosine;
        const pole = 1.0 - alpha;
        const attenuation = @exp(-sqrt_two * math.pi / smoothing_period);
        const hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
        const hp1 = 2.0 * pole;
        const hp2 = pole * pole;
        const ss2 = 2.0 * attenuation * @cos(sqrt_two * math.pi / smoothing_period);
        const ss3 = -attenuation * attenuation;
        const ss1 = 1.0 - ss2 - ss3;

        if (!(pole > 0.0 and pole < 1.0 and attenuation > 0.0 and attenuation < 1.0 and ss1 > 0.0) or
            !math.isFinite(hp0) or
            !math.isFinite(hp1) or
            !math.isFinite(hp2) or
            !math.isFinite(ss1) or
            !math.isFinite(ss2) or
            !math.isFinite(ss3))
        {
            return error.UnstableFilterCoefficients;
        }

        return .{
            .hp0 = hp0,
            .hp1 = hp1,
            .hp2 = hp2,
            .ss1 = ss1,
            .ss2 = ss2,
            .ss3 = ss3,
        };
    }

    fn reset(self: *RoofFilter) void {
        self.x1 = 0.0;
        self.x2 = 0.0;
        self.h1 = 0.0;
        self.h2 = 0.0;
        self.f1 = 0.0;
        self.f2 = 0.0;
    }

    fn candidate(self: *const RoofFilter, price: f64) Error!struct { high: f64, filtered: f64 } {
        const high = self.hp0 * (price - 2.0 * self.x1 + self.x2) + self.hp1 * self.h1 - self.hp2 * self.h2;
        const filtered = self.ss1 * (high + self.h1) / 2.0 + self.ss2 * self.f1 + self.ss3 * self.f2;
        if (!math.isFinite(high) or !math.isFinite(filtered)) {
            return error.NumericOverflow;
        }
        return .{ .high = high, .filtered = filtered };
    }

    fn commit(self: *RoofFilter, price: f64, high: f64, filtered: f64) void {
        self.x2 = self.x1;
        self.x1 = price;
        self.h2 = self.h1;
        self.h1 = high;
        self.f2 = self.f1;
        self.f1 = filtered;
    }
};

/// Ehlers' 2013 Roofed Autocorrelation Heatmap indicator (RACH, Book Listing 8-2).
pub const RoofedAutoCorrelationHeatmap = struct {
    allocator: std.mem.Allocator,
    resolved: ResolvedParams,
    total_lags: usize,
    displayed_bins: usize,
    history_len: usize,
    roof: RoofFilter,
    history: []f64,
    next_history: []f64,
    correlations: []f64,
    next_correlations: []f64,
    sample_count: usize = 0,
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
        params: RoofedAutoCorrelationHeatmapParams,
    ) Error!RoofedAutoCorrelationHeatmap {
        if (params.min_lag < 0) {
            return error.InvalidMinLag;
        }
        if (params.max_lag < params.min_lag) {
            return error.InvalidMaxLag;
        }
        if (params.averaging_length < 0) {
            return error.InvalidAveragingLength;
        }

        const lower_lag: usize = @intCast(params.min_lag);
        const upper_lag: usize = @intCast(params.max_lag);
        const displayed_bins = upper_lag - lower_lag + 1;
        if (displayed_bins > heatmap_mod.max_heatmap_values) {
            return error.LagRangeTooLarge;
        }

        const roof = try RoofFilter.init(params.highpass_period, params.smoothing_period);
        const averaging_window: usize = @intCast(params.averaging_length);
        const effective_window: usize = if (averaging_window > 0) averaging_window else upper_lag;
        const raw_history = math.add(usize, upper_lag, effective_window) catch return error.OutOfMemory;
        const history_len: usize = @max(@as(usize, 1), raw_history);
        const total_lags = math.add(usize, upper_lag, 1) catch return error.OutOfMemory;

        const history = allocator.alloc(f64, history_len) catch return error.OutOfMemory;
        errdefer allocator.free(history);
        const next_history = allocator.alloc(f64, history_len) catch return error.OutOfMemory;
        errdefer allocator.free(next_history);
        const correlations = allocator.alloc(f64, total_lags) catch return error.OutOfMemory;
        errdefer allocator.free(correlations);
        const next_correlations = allocator.alloc(f64, total_lags) catch return error.OutOfMemory;
        errdefer allocator.free(next_correlations);

        @memset(history, 0.0);
        @memset(next_history, 0.0);
        @memset(correlations, 0.0);
        @memset(next_correlations, 0.0);

        const bar_comp = params.bar_component orelse bar_component.default_bar_component;
        const quote_comp = params.quote_component orelse quote_component.default_quote_component;
        const trade_comp = params.trade_component orelse trade_component.default_trade_component;

        var mnemonic_storage: [output_metadata_mod.max_mnemonic_len]u8 = undefined;
        const mn = try mnemonic(&mnemonic_storage, params);

        var description_storage: [output_metadata_mod.max_description_len]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Roofed autocorrelation heatmap {s}", .{mn}) catch
            return error.MnemonicTooLong;

        return .{
            .allocator = allocator,
            .resolved = .{
                .min_lag = lower_lag,
                .max_lag = upper_lag,
                .averaging_length = averaging_window,
                .highpass_period = params.highpass_period,
                .smoothing_period = params.smoothing_period,
            },
            .total_lags = total_lags,
            .displayed_bins = displayed_bins,
            .history_len = history_len,
            .roof = roof,
            .history = history,
            .next_history = next_history,
            .correlations = correlations,
            .next_correlations = next_correlations,
            .last_output = Heatmap.empty(0, @floatFromInt(lower_lag), @floatFromInt(upper_lag), 1.0),
            .bar_func = bar_component.componentValue(bar_comp),
            .quote_func = quote_component.componentValue(quote_comp),
            .trade_func = trade_component.componentValue(trade_comp),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mn.len,
            .description_storage = description_storage,
            .description_len = desc.len,
        };
    }

    pub fn initDefault(allocator: std.mem.Allocator) Error!RoofedAutoCorrelationHeatmap {
        return init(allocator, defaultParams());
    }

    pub fn deinit(self: *RoofedAutoCorrelationHeatmap) void {
        self.allocator.free(self.history);
        self.allocator.free(self.next_history);
        self.allocator.free(self.correlations);
        self.allocator.free(self.next_correlations);
    }

    pub fn reset(self: *RoofedAutoCorrelationHeatmap) void {
        self.roof.reset();
        self.sample_count = 0;
        @memset(self.history, 0.0);
        @memset(self.next_history, 0.0);
        @memset(self.correlations, 0.0);
        @memset(self.next_correlations, 0.0);
        self.last_output = Heatmap.empty(
            0,
            @floatFromInt(self.resolved.min_lag),
            @floatFromInt(self.resolved.max_lag),
            1.0,
        );
    }

    pub fn fixSlices(self: *RoofedAutoCorrelationHeatmap) void {
        _ = self;
    }

    pub fn mnemonicText(self: *const RoofedAutoCorrelationHeatmap) []const u8 {
        return self.mnemonic_storage[0..self.mnemonic_len];
    }

    fn descriptionText(self: *const RoofedAutoCorrelationHeatmap) []const u8 {
        return self.description_storage[0..self.description_len];
    }

    pub fn parameters(self: *const RoofedAutoCorrelationHeatmap) ResolvedParams {
        return self.resolved;
    }

    pub fn minLag(self: *const RoofedAutoCorrelationHeatmap) usize {
        return self.resolved.min_lag;
    }

    pub fn maxLag(self: *const RoofedAutoCorrelationHeatmap) usize {
        return self.resolved.max_lag;
    }

    pub fn minPeriod(self: *const RoofedAutoCorrelationHeatmap) usize {
        return self.resolved.min_lag;
    }

    pub fn maxPeriod(self: *const RoofedAutoCorrelationHeatmap) usize {
        return self.resolved.max_lag;
    }

    pub fn averagingLength(self: *const RoofedAutoCorrelationHeatmap) usize {
        return self.resolved.averaging_length;
    }

    pub fn highpassPeriod(self: *const RoofedAutoCorrelationHeatmap) f64 {
        return self.resolved.highpass_period;
    }

    pub fn smoothingPeriod(self: *const RoofedAutoCorrelationHeatmap) f64 {
        return self.resolved.smoothing_period;
    }

    pub fn warmupBars(self: *const RoofedAutoCorrelationHeatmap) usize {
        return self.history_len - 1;
    }

    pub fn samplesSeen(self: *const RoofedAutoCorrelationHeatmap) usize {
        return self.sample_count;
    }

    pub fn state(self: *const RoofedAutoCorrelationHeatmap) AutocorrelationState {
        const count = @min(self.total_lags, max_state_correlations);
        var snapshot = AutocorrelationState{
            .high_pass = self.roof.h1,
            .filtered = self.roof.f1,
            .bins_len = count,
        };
        @memcpy(snapshot.correlations_storage[0..count], self.correlations[0..count]);
        return snapshot;
    }

    pub fn output(self: *const RoofedAutoCorrelationHeatmap) Heatmap {
        return self.last_output;
    }

    pub fn tryUpdate(self: *RoofedAutoCorrelationHeatmap, sample: f64, time: i64) Error!Heatmap {
        if (math.isInf(sample)) {
            return error.InvalidSample;
        }
        const min_lag_float: f64 = @floatFromInt(self.resolved.min_lag);
        const max_lag_float: f64 = @floatFromInt(self.resolved.max_lag);
        if (math.isNan(sample)) {
            return Heatmap.empty(time, min_lag_float, max_lag_float, 1.0);
        }

        const candidate_out = try self.roof.candidate(sample);

        self.next_history[0] = candidate_out.filtered;
        if (self.history_len > 1) {
            @memcpy(self.next_history[1..self.history_len], self.history[0 .. self.history_len - 1]);
        }

        @memcpy(self.next_correlations, self.correlations);

        for (0..self.total_lags) |lag| {
            const window = if (self.resolved.averaging_length > 0) self.resolved.averaging_length else lag;
            const window_float: f64 = @floatFromInt(window);
            var sum_x: f64 = 0.0;
            var sum_y: f64 = 0.0;
            var sum_xx: f64 = 0.0;
            var sum_xy: f64 = 0.0;
            var sum_yy: f64 = 0.0;
            for (0..window) |offset| {
                const sample_x = self.next_history[offset];
                const sample_y = self.next_history[lag + offset];
                sum_x += sample_x;
                sum_y += sample_y;
                sum_xx += sample_x * sample_x;
                sum_xy += sample_x * sample_y;
                sum_yy += sample_y * sample_y;
            }
            const denominator = (window_float * sum_xx - sum_x * sum_x) * (window_float * sum_yy - sum_y * sum_y);
            if (!math.isFinite(sum_x) or !math.isFinite(sum_y) or
                !math.isFinite(sum_xx) or !math.isFinite(sum_xy) or !math.isFinite(sum_yy) or
                !math.isFinite(denominator))
            {
                return error.NumericOverflow;
            }
            if (denominator > 0.0) {
                self.next_correlations[lag] = (window_float * sum_xy - sum_x * sum_y) / @sqrt(denominator);
            }
            self.next_correlations[lag] = 0.5 * (self.next_correlations[lag] + 1.0);
            if (!math.isFinite(self.next_correlations[lag])) {
                return error.NumericOverflow;
            }
        }

        self.roof.commit(sample, candidate_out.high, candidate_out.filtered);
        @memcpy(self.history, self.next_history);
        @memcpy(self.correlations, self.next_correlations);
        self.sample_count += 1;

        const current_output: Heatmap = if (self.sample_count >= self.history_len) blk: {
            const active_values = self.next_correlations[self.resolved.min_lag .. self.resolved.max_lag + 1];
            var value_min = active_values[0];
            var value_max = active_values[0];
            for (active_values[1..]) |item| {
                if (item < value_min) value_min = item;
                if (item > value_max) value_max = item;
            }
            break :blk Heatmap.new(time, min_lag_float, max_lag_float, 1.0, value_min, value_max, active_values);
        } else Heatmap.empty(time, min_lag_float, max_lag_float, 1.0);

        self.last_output = current_output;
        return current_output;
    }

    pub fn update(self: *RoofedAutoCorrelationHeatmap, sample: f64, time: i64) Heatmap {
        return self.tryUpdate(sample, time) catch |err| {
            std.debug.panic("RoofedAutoCorrelationHeatmap.update failed: {s}", .{@errorName(err)});
        };
    }

    pub const updateSample = update;

    fn wrapEntity(out_value: Heatmap) OutputArray {
        var arr = OutputArray{};
        arr.append(.{ .heatmap = out_value });
        return arr;
    }

    pub fn updateScalar(self: *RoofedAutoCorrelationHeatmap, sample: *const Scalar) OutputArray {
        return wrapEntity(self.update(sample.value, sample.time));
    }

    pub fn updateBar(self: *RoofedAutoCorrelationHeatmap, sample: *const Bar) OutputArray {
        return wrapEntity(self.update(self.bar_func(sample.*), sample.time));
    }

    pub fn updateQuote(self: *RoofedAutoCorrelationHeatmap, sample: *const Quote) OutputArray {
        return wrapEntity(self.update(self.quote_func(sample.*), sample.time));
    }

    pub fn updateTrade(self: *RoofedAutoCorrelationHeatmap, sample: *const Trade) OutputArray {
        return wrapEntity(self.update(self.trade_func(sample.*), sample.time));
    }

    pub fn isPrimed(self: *const RoofedAutoCorrelationHeatmap) bool {
        return self.sample_count >= self.history_len;
    }

    pub fn getMetadata(self: *const RoofedAutoCorrelationHeatmap, out: *Metadata) void {
        const mn = self.mnemonicText();
        const desc = self.descriptionText();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
        };
        build_metadata_mod.buildMetadata(out, .roofed_auto_correlation_heatmap, mn, desc, &texts);
    }

    pub fn indicator(self: *RoofedAutoCorrelationHeatmap) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(RoofedAutoCorrelationHeatmap);
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
const testing = std.testing;
const testdata = @import("testdata.zig");

const fixture_tolerance: f64 = 2e-11;
const a2_fixture_tolerance: f64 = 1e-10;

fn expectCloseTol(actual: f64, expected: f64, tolerance: f64) !void {
    if (math.isNan(expected)) {
        try testing.expect(math.isNan(actual));
        return;
    }
    try testing.expect(math.isFinite(actual));
    const bound = @max(tolerance, @abs(expected) * tolerance);
    try testing.expect(@abs(actual - expected) <= bound);
}

fn verifyReferenceCase(
    comptime cols: usize,
    params: Params,
    input_close: *const [252]f64,
    expected_heatmap: *const [252][cols]f64,
    expected_min: *const [252]f64,
    expected_max: *const [252]f64,
) !void {
    var indicator_instance = try RoofedAutoCorrelationHeatmap.init(testing.allocator, params);
    defer indicator_instance.deinit();

    const min_lag_usize: usize = @intCast(params.min_lag);
    const max_lag_usize: usize = @intCast(params.max_lag);
    const avg_usize: usize = @intCast(params.averaging_length);
    const effective_window: usize = if (avg_usize > 0) avg_usize else max_lag_usize;
    const history_len = @max(@as(usize, 1), max_lag_usize + effective_window);
    try testing.expectEqual(history_len - 1, indicator_instance.warmupBars());
    try testing.expectEqual(cols, max_lag_usize - min_lag_usize + 1);

    const case_tolerance = if (params.averaging_length == 2) a2_fixture_tolerance else fixture_tolerance;

    for (0..252) |bar_index| {
        const timestamp: i64 = @intCast(bar_index + 1);
        const result = indicator_instance.update(input_close[bar_index], timestamp);
        try testing.expectEqual(@as(f64, @floatFromInt(min_lag_usize)), result.parameter_first);
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
            try expectCloseTol(result.value_min, expected_min[bar_index], case_tolerance);
            try expectCloseTol(result.value_max, expected_max[bar_index], case_tolerance);
            for (0..cols) |bin_index| {
                try expectCloseTol(values[bin_index], expected_heatmap[bar_index][bin_index], case_tolerance);
            }
        }
    }
}

test "roofed auto correlation heatmap all 12 reference cases" {
    const input_close = testdata.testInput();

    {
        const hm = testdata.expectedL03_L148_A0_HP48_LP10_HEATMAP();
        const mn = testdata.expectedL03_L148_A0_HP48_LP10_MIN();
        const mx = testdata.expectedL03_L148_A0_HP48_LP10_MAX();
        try verifyReferenceCase(46, .{ .min_lag = 3, .max_lag = 48, .averaging_length = 0, .highpass_period = 48.0, .smoothing_period = 10.0 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL03_L148_A2_HP48_LP10_HEATMAP();
        const mn = testdata.expectedL03_L148_A2_HP48_LP10_MIN();
        const mx = testdata.expectedL03_L148_A2_HP48_LP10_MAX();
        try verifyReferenceCase(46, .{ .min_lag = 3, .max_lag = 48, .averaging_length = 2, .highpass_period = 48.0, .smoothing_period = 10.0 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL03_L148_A3_HP48_LP10_HEATMAP();
        const mn = testdata.expectedL03_L148_A3_HP48_LP10_MIN();
        const mx = testdata.expectedL03_L148_A3_HP48_LP10_MAX();
        try verifyReferenceCase(46, .{ .min_lag = 3, .max_lag = 48, .averaging_length = 3, .highpass_period = 48.0, .smoothing_period = 10.0 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL03_L148_A20_HP48_LP10_HEATMAP();
        const mn = testdata.expectedL03_L148_A20_HP48_LP10_MIN();
        const mx = testdata.expectedL03_L148_A20_HP48_LP10_MAX();
        try verifyReferenceCase(46, .{ .min_lag = 3, .max_lag = 48, .averaging_length = 20, .highpass_period = 48.0, .smoothing_period = 10.0 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL03_L148_A100_HP48_LP10_HEATMAP();
        const mn = testdata.expectedL03_L148_A100_HP48_LP10_MIN();
        const mx = testdata.expectedL03_L148_A100_HP48_LP10_MAX();
        try verifyReferenceCase(46, .{ .min_lag = 3, .max_lag = 48, .averaging_length = 100, .highpass_period = 48.0, .smoothing_period = 10.0 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL03_L148_A1_HP48_LP10_HEATMAP();
        const mn = testdata.expectedL03_L148_A1_HP48_LP10_MIN();
        const mx = testdata.expectedL03_L148_A1_HP48_LP10_MAX();
        try verifyReferenceCase(46, .{ .min_lag = 3, .max_lag = 48, .averaging_length = 1, .highpass_period = 48.0, .smoothing_period = 10.0 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL03_L148_A0_HP48_LP10_5_HEATMAP();
        const mn = testdata.expectedL03_L148_A0_HP48_LP10_5_MIN();
        const mx = testdata.expectedL03_L148_A0_HP48_LP10_5_MAX();
        try verifyReferenceCase(46, .{ .min_lag = 3, .max_lag = 48, .averaging_length = 0, .highpass_period = 48.0, .smoothing_period = 10.5 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL03_L148_A0_HP48_LP20_HEATMAP();
        const mn = testdata.expectedL03_L148_A0_HP48_LP20_MIN();
        const mx = testdata.expectedL03_L148_A0_HP48_LP20_MAX();
        try verifyReferenceCase(46, .{ .min_lag = 3, .max_lag = 48, .averaging_length = 0, .highpass_period = 48.0, .smoothing_period = 20.0 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL03_L148_A0_HP80_LP10_HEATMAP();
        const mn = testdata.expectedL03_L148_A0_HP80_LP10_MIN();
        const mx = testdata.expectedL03_L148_A0_HP80_LP10_MAX();
        try verifyReferenceCase(46, .{ .min_lag = 3, .max_lag = 48, .averaging_length = 0, .highpass_period = 80.0, .smoothing_period = 10.0 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL03_L124_A0_HP24_LP10_HEATMAP();
        const mn = testdata.expectedL03_L124_A0_HP24_LP10_MIN();
        const mx = testdata.expectedL03_L124_A0_HP24_LP10_MAX();
        try verifyReferenceCase(22, .{ .min_lag = 3, .max_lag = 24, .averaging_length = 0, .highpass_period = 24.0, .smoothing_period = 10.0 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL012_L112_A12_HP48_LP10_HEATMAP();
        const mn = testdata.expectedL012_L112_A12_HP48_LP10_MIN();
        const mx = testdata.expectedL012_L112_A12_HP48_LP10_MAX();
        try verifyReferenceCase(1, .{ .min_lag = 12, .max_lag = 12, .averaging_length = 12, .highpass_period = 48.0, .smoothing_period = 10.0 }, &input_close, &hm, &mn, &mx);
    }
    {
        const hm = testdata.expectedL00_L10_A3_HP48_LP10_HEATMAP();
        const mn = testdata.expectedL00_L10_A3_HP48_LP10_MIN();
        const mx = testdata.expectedL00_L10_A3_HP48_LP10_MAX();
        try verifyReferenceCase(1, .{ .min_lag = 0, .max_lag = 0, .averaging_length = 3, .highpass_period = 48.0, .smoothing_period = 10.0 }, &input_close, &hm, &mn, &mx);
    }
}

test "roofed auto correlation heatmap validation, edge behavior, metadata, and entities" {
    const input_close = testdata.testInput();
    const input_open = testdata.testOpen();
    const input_high = testdata.testHigh();
    const input_low = testdata.testLow();

    try testing.expectError(error.InvalidMinLag, RoofedAutoCorrelationHeatmap.init(testing.allocator, .{ .min_lag = -1 }));
    try testing.expectError(error.InvalidMaxLag, RoofedAutoCorrelationHeatmap.init(testing.allocator, .{ .min_lag = 10, .max_lag = 9 }));
    try testing.expectError(error.LagRangeTooLarge, RoofedAutoCorrelationHeatmap.init(testing.allocator, .{ .min_lag = 0, .max_lag = 256 }));
    try testing.expectError(error.InvalidAveragingLength, RoofedAutoCorrelationHeatmap.init(testing.allocator, .{ .averaging_length = -1 }));
    try testing.expectError(error.InvalidHighpassPeriod, RoofedAutoCorrelationHeatmap.init(testing.allocator, .{ .highpass_period = 4.9 }));
    try testing.expectError(error.InvalidSmoothingPeriod, RoofedAutoCorrelationHeatmap.init(testing.allocator, .{ .smoothing_period = 1.9 }));

    {
        var subject = try RoofedAutoCorrelationHeatmap.init(testing.allocator, .{
            .min_lag = 3,
            .max_lag = 12,
            .averaging_length = 5,
            .highpass_period = 24.0,
            .smoothing_period = 10.0,
        });
        defer subject.deinit();
        var clean = try RoofedAutoCorrelationHeatmap.init(testing.allocator, .{
            .min_lag = 3,
            .max_lag = 12,
            .averaging_length = 5,
            .highpass_period = 24.0,
            .smoothing_period = 10.0,
        });
        defer clean.deinit();

        for (0..40) |bar_index| {
            const before = subject.state();
            const skipped = try subject.tryUpdate(math.nan(f64), @intCast(bar_index));
            try testing.expect(skipped.isEmpty());
            try testing.expectError(error.InvalidSample, subject.tryUpdate(math.inf(f64), @intCast(bar_index)));
            try testing.expectError(error.NumericOverflow, subject.tryUpdate(1e308, @intCast(bar_index)));
            const after = subject.state();
            try testing.expect(before.eql(&after));

            const actual = subject.update(input_close[bar_index], @intCast(bar_index));
            const expected = clean.update(input_close[bar_index], @intCast(bar_index));
            try testing.expectEqualSlices(f64, expected.valuesSlice(), actual.valuesSlice());
        }

        subject.reset();
        try testing.expectEqual(@as(usize, 0), subject.samplesSeen());
        try testing.expect(!subject.isPrimed());
    }

    {
        var default_ind = try RoofedAutoCorrelationHeatmap.initDefault(testing.allocator);
        defer default_ind.deinit();
        var meta: Metadata = undefined;
        default_ind.getMetadata(&meta);
        try testing.expectEqual(Identifier.roofed_auto_correlation_heatmap, meta.identifier);
        try testing.expectEqualStrings("rach(3, 48, 0, 48, 10)", meta.mnemonic);
        try testing.expectEqualStrings("Roofed autocorrelation heatmap rach(3, 48, 0, 48, 10)", meta.description);
        try testing.expectEqual(@as(usize, 1), meta.outputs_len);
        try testing.expectEqual(@intFromEnum(RoofedAutoCorrelationHeatmapOutput.value), meta.outputs_buf[0].kind);
        try testing.expectEqual(shape_mod.Shape.heatmap, meta.outputs_buf[0].shape);

        const desc = descriptor_mod.descriptorOf(.roofed_auto_correlation_heatmap);
        try testing.expect(desc != null);
        try testing.expectEqual(role_mod.Role.correlation, desc.?.outputs[0].role);

        var mn_buffer: [64]u8 = undefined;
        try testing.expectEqualStrings(
            "rach(0, 0, 3, 24, 10.5, hl/2, a, v)",
            try mnemonic(&mn_buffer, .{
                .min_lag = 0,
                .max_lag = 0,
                .averaging_length = 3,
                .highpass_period = 24.0,
                .smoothing_period = 10.5,
                .bar_component = .median,
                .quote_component = .ask,
                .trade_component = .volume,
            }),
        );
    }

    {
        var scalar_ind = try RoofedAutoCorrelationHeatmap.initDefault(testing.allocator);
        defer scalar_ind.deinit();
        var bar_ind = try RoofedAutoCorrelationHeatmap.initDefault(testing.allocator);
        defer bar_ind.deinit();
        var quote_ind = try RoofedAutoCorrelationHeatmap.initDefault(testing.allocator);
        defer quote_ind.deinit();
        var trade_ind = try RoofedAutoCorrelationHeatmap.initDefault(testing.allocator);
        defer trade_ind.deinit();

        for (0..252) |bar_index| {
            const stamp: i64 = @intCast(bar_index + 1);
            const close_price = input_close[bar_index];
            const scalar_sample = Scalar{ .time = stamp, .value = close_price };
            const bar_sample = Bar{
                .time = stamp,
                .open = input_open[bar_index],
                .high = input_high[bar_index],
                .low = input_low[bar_index],
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
        }
    }
}
