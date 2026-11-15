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
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");
const shape_mod = @import("../../core/outputs/shape.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

pub const Error = error{
    InvalidLength,
    InvalidSmoothingFactor,
    InvalidCoefficientMode,
    InvalidSample,
    InvalidRangeOrder,
    ArithmeticOverflow,
    MnemonicTooLong,
    OutOfMemory,
};

/// Enumerates the outputs of the fractal adaptive moving average.
pub const FractalAdaptiveMovingAverageOutput = enum(u8) {
    value = 1,
    fdim = 2,
};

/// Parameters to create a FractalAdaptiveMovingAverage.
pub const FractalAdaptiveMovingAverageParams = struct {
    /// Even lookback length (>= 2). Default is 16.
    length: i32 = 16,
    /// Slowest boundary smoothing factor (finite, 0 < A <= 1). Default is 0.01.
    slowest_smoothing_factor: f64 = 0.01,
    /// Alpha exponent scale mode: 1 for ln(A), 0 for literal -4.6 (requires A == 0.01). Default is 1.
    coefficient_mode: i32 = 1,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Params = FractalAdaptiveMovingAverageParams;

pub fn defaultParams() FractalAdaptiveMovingAverageParams {
    return .{};
}

fn normalizeZero(number: f64) f64 {
    return if (number == 0.0) 0.0 else number;
}

fn formatMnemonicWithPrefix(
    prefix: []const u8,
    buffer: []u8,
    params: FractalAdaptiveMovingAverageParams,
) Error![]const u8 {
    var triple_storage: [64]u8 = undefined;
    const suffix = component_triple_mnemonic_mod.componentTripleMnemonic(
        &triple_storage,
        params.bar_component orelse bar_component.default_bar_component,
        params.quote_component orelse quote_component.default_quote_component,
        params.trade_component orelse trade_component.default_trade_component,
    );
    const mode_tag: []const u8 = if (params.coefficient_mode == 0) ", mode=0" else "";
    return std.fmt.bufPrint(buffer, "{s}({d}, {d}{s}{s})", .{
        prefix,
        params.length,
        normalizeZero(params.slowest_smoothing_factor),
        mode_tag,
        suffix,
    }) catch return error.MnemonicTooLong;
}

pub fn mnemonic(buffer: []u8, params: FractalAdaptiveMovingAverageParams) Error![]const u8 {
    return formatMnemonicWithPrefix("frama", buffer, params);
}

pub fn dimensionMnemonic(buffer: []u8, params: FractalAdaptiveMovingAverageParams) Error![]const u8 {
    return formatMnemonicWithPrefix("framaDim", buffer, params);
}

fn boxRange(highs: []const f64, lows: []const f64, start: usize, end: usize) Error!f64 {
    var highest = highs[start];
    var lowest = lows[start];
    for (start + 1..end) |index| {
        if (highs[index] > highest) {
            highest = highs[index];
        }
        if (lows[index] < lowest) {
            lowest = lows[index];
        }
    }
    const span = highest - lowest;
    if (!math.isFinite(span)) {
        return error.ArithmeticOverflow;
    }
    return span;
}

pub const FractalAdaptiveMovingAverage = struct {
    allocator: std.mem.Allocator,
    length: usize,
    half_length: usize,
    warmup_count: usize,
    slowest_smoothing_factor: f64,
    coefficient_mode: i32,
    exponent_scale: f64,

    count: usize,
    highs: []f64,
    lows: []f64,
    scratch_highs: []f64,
    scratch_lows: []f64,
    dimension_state: f64,
    alpha_state: f64,
    filter_state: f64,
    current_value: f64,
    current_dimension: f64,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    mnemonic_storage: [128]u8,
    mnemonic_len: usize,
    description_storage: [256]u8,
    description_len: usize,
    mnemonic_fdim_storage: [128]u8,
    mnemonic_fdim_len: usize,
    description_fdim_storage: [256]u8,
    description_fdim_len: usize,

    pub fn initDefault(allocator: std.mem.Allocator) Error!FractalAdaptiveMovingAverage {
        return init(allocator, defaultParams());
    }

    pub fn init(allocator: std.mem.Allocator, params: FractalAdaptiveMovingAverageParams) Error!FractalAdaptiveMovingAverage {
        if (params.length < 2 or @rem(params.length, 2) != 0) {
            return error.InvalidLength;
        }

        const slowest = params.slowest_smoothing_factor;
        if (!math.isFinite(slowest) or slowest <= 0.0 or slowest > 1.0) {
            return error.InvalidSmoothingFactor;
        }

        if (params.coefficient_mode != 0 and params.coefficient_mode != 1) {
            return error.InvalidCoefficientMode;
        }
        if (params.coefficient_mode == 0 and slowest != 0.01) {
            return error.InvalidCoefficientMode;
        }

        const length: usize = @intCast(params.length);
        const half_length = length / 2;

        const highs = allocator.alloc(f64, length) catch return error.OutOfMemory;
        errdefer allocator.free(highs);
        const lows = allocator.alloc(f64, length) catch return error.OutOfMemory;
        errdefer allocator.free(lows);
        const scratch_highs = allocator.alloc(f64, length) catch return error.OutOfMemory;
        errdefer allocator.free(scratch_highs);
        const scratch_lows = allocator.alloc(f64, length) catch return error.OutOfMemory;
        errdefer allocator.free(scratch_lows);

        const bar_comp = params.bar_component orelse bar_component.default_bar_component;
        const quote_comp = params.quote_component orelse quote_component.default_quote_component;
        const trade_comp = params.trade_component orelse trade_component.default_trade_component;

        var mnemonic_storage: [128]u8 = undefined;
        const mn = try mnemonic(&mnemonic_storage, params);
        const mnemonic_len = mn.len;

        var description_storage: [256]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Fractal adaptive moving average {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_fdim_storage: [128]u8 = undefined;
        const mn_fdim = try dimensionMnemonic(&mnemonic_fdim_storage, params);
        const mnemonic_fdim_len = mn_fdim.len;

        var description_fdim_storage: [256]u8 = undefined;
        const desc_fdim = std.fmt.bufPrint(&description_fdim_storage, "Fractal adaptive moving average {s}", .{mn_fdim}) catch
            return error.MnemonicTooLong;
        const description_fdim_len = desc_fdim.len;

        const exponent_scale = if (params.coefficient_mode == 1) @log(slowest) else -4.6;

        var instance = FractalAdaptiveMovingAverage{
            .allocator = allocator,
            .length = length,
            .half_length = half_length,
            .warmup_count = length - 1,
            .slowest_smoothing_factor = slowest,
            .coefficient_mode = params.coefficient_mode,
            .exponent_scale = exponent_scale,
            .count = 0,
            .highs = highs,
            .lows = lows,
            .scratch_highs = scratch_highs,
            .scratch_lows = scratch_lows,
            .dimension_state = 0.0,
            .alpha_state = 1.0,
            .filter_state = 0.0,
            .current_value = math.nan(f64),
            .current_dimension = math.nan(f64),
            .bar_func = bar_component.componentValue(bar_comp),
            .quote_func = quote_component.componentValue(quote_comp),
            .trade_func = trade_component.componentValue(trade_comp),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mnemonic_len,
            .description_storage = description_storage,
            .description_len = description_len,
            .mnemonic_fdim_storage = mnemonic_fdim_storage,
            .mnemonic_fdim_len = mnemonic_fdim_len,
            .description_fdim_storage = description_fdim_storage,
            .description_fdim_len = description_fdim_len,
        };
        instance.reset();
        return instance;
    }

    pub fn deinit(self: *FractalAdaptiveMovingAverage) void {
        self.allocator.free(self.highs);
        self.allocator.free(self.lows);
        self.allocator.free(self.scratch_highs);
        self.allocator.free(self.scratch_lows);
    }

    pub fn fixSlices(self: *FractalAdaptiveMovingAverage) void {
        _ = self;
    }

    pub fn reset(self: *FractalAdaptiveMovingAverage) void {
        self.count = 0;
        @memset(self.highs, 0.0);
        @memset(self.lows, 0.0);
        @memset(self.scratch_highs, 0.0);
        @memset(self.scratch_lows, 0.0);
        self.dimension_state = 0.0;
        self.alpha_state = 1.0;
        self.filter_state = 0.0;
        self.current_value = math.nan(f64);
        self.current_dimension = math.nan(f64);
    }

    pub fn warmupBars(self: *const FractalAdaptiveMovingAverage) usize {
        return self.warmup_count;
    }

    pub fn samplesSeen(self: *const FractalAdaptiveMovingAverage) usize {
        return self.count;
    }

    pub fn value(self: *const FractalAdaptiveMovingAverage) f64 {
        return self.current_value;
    }

    pub fn fractalDimension(self: *const FractalAdaptiveMovingAverage) f64 {
        return self.current_dimension;
    }

    pub fn alpha(self: *const FractalAdaptiveMovingAverage) f64 {
        return if (self.isPrimed()) self.alpha_state else math.nan(f64);
    }

    pub fn isPrimed(self: *const FractalAdaptiveMovingAverage) bool {
        return self.count >= self.length;
    }

    pub fn getMetadata(self: *const FractalAdaptiveMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .fractal_adaptive_moving_average,
            self.mnemonic_storage[0..self.mnemonic_len],
            self.description_storage[0..self.description_len],
            &[_]build_metadata_mod.OutputText{
                .{
                    .mnemonic = self.mnemonic_storage[0..self.mnemonic_len],
                    .description = self.description_storage[0..self.description_len],
                },
                .{
                    .mnemonic = self.mnemonic_fdim_storage[0..self.mnemonic_fdim_len],
                    .description = self.description_fdim_storage[0..self.description_fdim_len],
                },
            },
        );
    }

    pub fn tryUpdate(
        self: *FractalAdaptiveMovingAverage,
        sample: f64,
        sample_high: f64,
        sample_low: f64,
    ) Error!f64 {
        if (math.isInf(sample) or math.isInf(sample_high) or math.isInf(sample_low)) {
            return error.InvalidSample;
        }
        if (math.isNan(sample) or math.isNan(sample_high) or math.isNan(sample_low)) {
            return math.nan(f64);
        }
        if (sample_high < sample_low) {
            return error.InvalidRangeOrder;
        }

        const keep = @min(self.count, self.length - 1);
        self.scratch_highs[0] = sample_high;
        self.scratch_lows[0] = sample_low;
        if (keep > 0) {
            @memcpy(self.scratch_highs[1 .. keep + 1], self.highs[0..keep]);
            @memcpy(self.scratch_lows[1 .. keep + 1], self.lows[0..keep]);
        }

        const next_count = self.count + 1;
        var dimension = self.dimension_state;
        var next_alpha = self.alpha_state;

        if (next_count >= self.length) {
            const range1 = try boxRange(self.scratch_highs, self.scratch_lows, 0, self.half_length);
            const range2 = try boxRange(self.scratch_highs, self.scratch_lows, self.half_length, self.length);
            const range3 = try boxRange(self.scratch_highs, self.scratch_lows, 0, self.length);

            const half_float: f64 = @floatFromInt(self.half_length);
            const length_float: f64 = @floatFromInt(self.length);
            const n1 = range1 / half_float;
            const n2 = range2 / half_float;
            const n3 = range3 / length_float;

            if (n1 > 0.0 and n2 > 0.0 and n3 > 0.0) {
                const combined = n1 + n2;
                if (!math.isFinite(combined)) {
                    return error.ArithmeticOverflow;
                }
                dimension = (@log(combined) - @log(n3)) / @log(@as(f64, 2.0));
                if (!math.isFinite(dimension)) {
                    return error.ArithmeticOverflow;
                }
            }

            if (dimension <= 1.0) {
                next_alpha = 1.0;
            } else {
                next_alpha = @exp(self.exponent_scale * (dimension - 1.0));
                next_alpha = @min(1.0, @max(self.slowest_smoothing_factor, next_alpha));
            }
        }

        const next_value = if (next_count <= self.length)
            sample
        else blk: {
            const computed = next_alpha * sample + (1.0 - next_alpha) * self.filter_state;
            if (!math.isFinite(computed)) {
                return error.ArithmeticOverflow;
            }
            break :blk computed;
        };

        const stored_len = keep + 1;
        @memcpy(self.highs[0..stored_len], self.scratch_highs[0..stored_len]);
        @memcpy(self.lows[0..stored_len], self.scratch_lows[0..stored_len]);
        self.count = next_count;
        self.dimension_state = dimension;
        self.alpha_state = next_alpha;
        self.filter_state = next_value;
        if (next_count >= self.length) {
            self.current_value = next_value;
            self.current_dimension = dimension;
            return next_value;
        }
        return math.nan(f64);
    }

    pub fn update(
        self: *FractalAdaptiveMovingAverage,
        sample: f64,
        sample_high: f64,
        sample_low: f64,
    ) f64 {
        return self.tryUpdate(sample, sample_high, sample_low) catch @panic("invalid FRAMA sample");
    }

    pub fn tryUpdatePrice(self: *FractalAdaptiveMovingAverage, price: f64) Error!f64 {
        return self.tryUpdate(price, price, price);
    }

    pub fn updatePrice(self: *FractalAdaptiveMovingAverage, price: f64) f64 {
        return self.update(price, price, price);
    }

    pub fn updateScalar(self: *FractalAdaptiveMovingAverage, sample: *const Scalar) OutputArray {
        const extracted = sample.value;
        return self.updateEntity(sample.time, extracted, extracted, extracted);
    }

    pub fn updateBar(self: *FractalAdaptiveMovingAverage, sample: *const Bar) OutputArray {
        const extracted = self.bar_func(sample.*);
        return self.updateEntity(sample.time, extracted, sample.high, sample.low);
    }

    pub fn updateQuote(self: *FractalAdaptiveMovingAverage, sample: *const Quote) OutputArray {
        const extracted = self.quote_func(sample.*);
        return self.updateEntity(sample.time, extracted, sample.ask_price, sample.bid_price);
    }

    pub fn updateTrade(self: *FractalAdaptiveMovingAverage, sample: *const Trade) OutputArray {
        const extracted = self.trade_func(sample.*);
        return self.updateEntity(sample.time, extracted, extracted, extracted);
    }

    fn updateEntity(
        self: *FractalAdaptiveMovingAverage,
        time: i64,
        sample: f64,
        sample_high: f64,
        sample_low: f64,
    ) OutputArray {
        const frama = self.update(sample, sample_high, sample_low);
        const fdim = if (math.isNan(frama)) math.nan(f64) else self.current_dimension;

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = frama } });
        out.append(.{ .scalar = .{ .time = time, .value = fdim } });
        return out;
    }

    pub fn indicator(self: *FractalAdaptiveMovingAverage) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(FractalAdaptiveMovingAverage);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");

const ReferenceCase = struct {
    params: FractalAdaptiveMovingAverageParams,
    expected_values_fn: *const fn () [252]f64,
    expected_dimensions_fn: *const fn () [252]f64,
};

const reference_cases = [_]ReferenceCase{
    .{ .params = .{ .length = 16, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL16_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL16_A0_01_C1_FDIM },
    .{ .params = .{ .length = 2, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL2_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL2_A0_01_C1_FDIM },
    .{ .params = .{ .length = 4, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL4_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL4_A0_01_C1_FDIM },
    .{ .params = .{ .length = 8, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL8_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL8_A0_01_C1_FDIM },
    .{ .params = .{ .length = 20, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL20_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL20_A0_01_C1_FDIM },
    .{ .params = .{ .length = 30, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL30_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL30_A0_01_C1_FDIM },
    .{ .params = .{ .length = 64, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL64_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL64_A0_01_C1_FDIM },
    .{ .params = .{ .length = 128, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL128_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL128_A0_01_C1_FDIM },
    .{ .params = .{ .length = 250, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL250_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL250_A0_01_C1_FDIM },
    .{ .params = .{ .length = 252, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL252_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL252_A0_01_C1_FDIM },
    .{ .params = .{ .length = 254, .slowest_smoothing_factor = 0.01, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL254_A0_01_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL254_A0_01_C1_FDIM },
    .{ .params = .{ .length = 16, .slowest_smoothing_factor = 0.001, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL16_A0_001_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL16_A0_001_C1_FDIM },
    .{ .params = .{ .length = 16, .slowest_smoothing_factor = 0.1, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL16_A0_1_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL16_A0_1_C1_FDIM },
    .{ .params = .{ .length = 16, .slowest_smoothing_factor = 1.0, .coefficient_mode = 1 }, .expected_values_fn = &testdata.expectedL16_A1_C1_VALUE, .expected_dimensions_fn = &testdata.expectedL16_A1_C1_FDIM },
    .{ .params = .{ .length = 16, .slowest_smoothing_factor = 0.01, .coefficient_mode = 0 }, .expected_values_fn = &testdata.expectedL16_A0_01_C0_VALUE, .expected_dimensions_fn = &testdata.expectedL16_A0_01_C0_FDIM },
    .{ .params = .{ .length = 30, .slowest_smoothing_factor = 0.01, .coefficient_mode = 0 }, .expected_values_fn = &testdata.expectedL30_A0_01_C0_VALUE, .expected_dimensions_fn = &testdata.expectedL30_A0_01_C0_FDIM },
};

test "FRAMA all 16 confirmed reference cases" {
    const inputs = testdata.testInput();
    for (reference_cases) |case| {
        var indicator = try FractalAdaptiveMovingAverage.init(testing.allocator, case.params);
        defer indicator.deinit();

        const expected_values = case.expected_values_fn();
        const expected_dimensions = case.expected_dimensions_fn();
        const warmup: usize = @intCast(case.params.length - 1);
        try testing.expectEqual(warmup, indicator.warmupBars());
        try testing.expect(!indicator.isPrimed());
        try testing.expect(math.isNan(indicator.value()));
        try testing.expect(math.isNan(indicator.fractalDimension()));
        try testing.expect(math.isNan(indicator.alpha()));

        for (inputs, 0..) |sample, index| {
            const actual_value = indicator.updatePrice(sample);
            const actual_dimension = indicator.fractalDimension();
            const target_value = expected_values[index];
            const target_dimension = expected_dimensions[index];

            if (index < warmup) {
                try testing.expect(!indicator.isPrimed());
                try testing.expect(math.isNan(actual_value));
                try testing.expect(math.isNan(actual_dimension));
                try testing.expect(math.isNan(target_value));
                try testing.expect(math.isNan(target_dimension));
            } else {
                try testing.expect(indicator.isPrimed());
                const value_bound = @max(1e-12, @abs(target_value) * 1e-12);
                const dim_bound = @max(1e-12, @abs(target_dimension) * 1e-12);
                try testing.expect(@abs(actual_value - target_value) <= value_bound);
                try testing.expect(@abs(actual_dimension - target_dimension) <= dim_bound);
                try testing.expect(@abs(indicator.value() - actual_value) <= 1e-15);
            }
        }
    }
}

test "FRAMA invariants: startup seed, ramp, gap, flat guard, and modes" {
    const inputs = testdata.testInput();
    var seeded = try FractalAdaptiveMovingAverage.initDefault(testing.allocator);
    defer seeded.deinit();
    var seed_value: f64 = 0.0;
    for (inputs[0..20], 0..) |price, index| {
        const actual_value = seeded.updatePrice(price);
        if (index == 15) {
            try testing.expectEqual(price, actual_value);
            seed_value = actual_value;
        }
        if (index == 16) {
            const current_alpha = seeded.alpha();
            const expected = current_alpha * price + (1.0 - current_alpha) * seed_value;
            try testing.expect(@abs(actual_value - expected) <= 1e-15);
        }
    }

    const expected_ramp = @log2(@as(f64, 28.0) / 15.0);
    for ([_]f64{ 100.0, 0.0, -100.0 }) |offset| {
        var ramp = try FractalAdaptiveMovingAverage.initDefault(testing.allocator);
        defer ramp.deinit();
        for (0..40) |step| {
            const sample = offset + @as(f64, @floatFromInt(step));
            const actual_value = ramp.updatePrice(sample);
            if (step >= 15) {
                try testing.expect(@abs(ramp.fractalDimension() - expected_ramp) <= 1e-13);
                try testing.expect(ramp.fractalDimension() < 1.0);
                try testing.expectEqual(@as(f64, 1.0), ramp.alpha());
                try testing.expectEqual(sample, actual_value);
            }
        }
    }

    var gap = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 4 });
    defer gap.deinit();
    var gap_value: f64 = math.nan(f64);
    for ([_]f64{ 0.0, 1.0, 100.0, 101.0 }) |price| {
        gap_value = gap.updatePrice(price);
    }
    try testing.expect(@abs(gap.fractalDimension() - @log2(@as(f64, 4.0) / 101.0)) <= 1e-13);
    try testing.expect(gap.fractalDimension() < 0.0);
    try testing.expectEqual(@as(f64, 1.0), gap.alpha());
    try testing.expectEqual(@as(f64, 101.0), gap_value);

    var hold = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 4 });
    defer hold.deinit();
    for ([_]f64{ 0.0, 1.0, 0.0, 1.0 }) |price| {
        _ = hold.updatePrice(price);
    }
    try testing.expectEqual(@as(f64, 2.0), hold.fractalDimension());
    for (0..8) |_| {
        _ = hold.updatePrice(1.0);
        try testing.expectEqual(@as(f64, 2.0), hold.fractalDimension());
    }

    var precise = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 4, .coefficient_mode = 1 });
    defer precise.deinit();
    var article = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 4, .coefficient_mode = 0 });
    defer article.deinit();
    for (0..5) |_| {
        _ = precise.update(1.0, 2.0, 0.0);
        _ = article.update(1.0, 2.0, 0.0);
    }
    try testing.expect(@abs(precise.alpha() - 0.01) <= 1e-15);
    try testing.expectEqual(@exp(@as(f64, -4.6)), article.alpha());
    try testing.expect(article.alpha() > precise.alpha());
}

test "FRAMA affine invariance, NaN gap, reset, validation, metadata, and entities" {
    const inputs = testdata.testInput();
    var original = try FractalAdaptiveMovingAverage.initDefault(testing.allocator);
    defer original.deinit();
    var transformed = try FractalAdaptiveMovingAverage.initDefault(testing.allocator);
    defer transformed.deinit();
    var with_gaps = try FractalAdaptiveMovingAverage.initDefault(testing.allocator);
    defer with_gaps.deinit();

    for (inputs) |price| {
        const left_val = original.updatePrice(price);
        const right_val = transformed.updatePrice(3.0 * price - 1000.0);
        const gap_val = with_gaps.updatePrice(price);
        const seen_before = with_gaps.samplesSeen();
        const saved_val = with_gaps.value();

        try testing.expect(math.isNan(with_gaps.updatePrice(math.nan(f64))));
        try testing.expect(math.isNan(with_gaps.update(price, math.nan(f64), price)));
        try testing.expectEqual(seen_before, with_gaps.samplesSeen());
        if (math.isNan(saved_val)) {
            try testing.expect(math.isNan(with_gaps.value()));
        } else {
            try testing.expectEqual(saved_val, with_gaps.value());
        }

        if (!math.isNan(left_val)) {
            try testing.expectEqual(left_val, gap_val);
            try testing.expect(@abs(original.fractalDimension() - transformed.fractalDimension()) <= 1e-11);
            try testing.expect(@abs((3.0 * left_val - 1000.0) - right_val) <= 1e-10);
        }
    }

    with_gaps.reset();
    try testing.expect(!with_gaps.isPrimed());
    try testing.expectEqual(@as(usize, 0), with_gaps.samplesSeen());
    var fresh = try FractalAdaptiveMovingAverage.initDefault(testing.allocator);
    defer fresh.deinit();
    for (inputs) |price| {
        const a = with_gaps.updatePrice(price);
        const b = fresh.updatePrice(price);
        if (math.isNan(b)) {
            try testing.expect(math.isNan(a));
        } else {
            try testing.expectEqual(b, a);
        }
    }

    try testing.expectError(error.InvalidLength, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 1 }));
    try testing.expectError(error.InvalidLength, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 17 }));
    try testing.expectError(error.InvalidSmoothingFactor, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .slowest_smoothing_factor = 0.0 }));
    try testing.expectError(error.InvalidSmoothingFactor, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .slowest_smoothing_factor = 1.1 }));
    try testing.expectError(error.InvalidSmoothingFactor, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .slowest_smoothing_factor = math.nan(f64) }));
    try testing.expectError(error.InvalidCoefficientMode, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .coefficient_mode = 2 }));
    try testing.expectError(error.InvalidCoefficientMode, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .coefficient_mode = 0, .slowest_smoothing_factor = 0.1 }));

    var atomic = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 2 });
    defer atomic.deinit();
    _ = atomic.updatePrice(0.0);
    try testing.expectError(error.InvalidSample, atomic.tryUpdatePrice(math.inf(f64)));
    try testing.expectError(error.InvalidRangeOrder, atomic.tryUpdate(1.0, 0.0, 2.0));
    try testing.expectError(error.ArithmeticOverflow, atomic.tryUpdate(0.0, 1e308, -1e308));
    try testing.expectEqual(@as(usize, 1), atomic.samplesSeen());
    try testing.expect(!atomic.isPrimed());

    var meta: Metadata = undefined;
    original.getMetadata(&meta);
    try testing.expectEqual(Identifier.fractal_adaptive_moving_average, meta.identifier);
    try testing.expectEqualStrings("frama(16, 0.01)", meta.mnemonic);
    try testing.expectEqualStrings("Fractal adaptive moving average frama(16, 0.01)", meta.description);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(shape_mod.Shape.scalar, meta.outputs_buf[0].shape);
    try testing.expectEqualStrings("framaDim(16, 0.01)", meta.outputs_buf[1].mnemonic);

    var mn_storage: [128]u8 = undefined;
    const custom_mn = try mnemonic(&mn_storage, .{
        .length = 30,
        .slowest_smoothing_factor = 0.01,
        .coefficient_mode = 0,
        .bar_component = .median,
        .quote_component = .bid,
        .trade_component = .volume,
    });
    try testing.expectEqualStrings("frama(30, 0.01, mode=0, hl/2, b, v)", custom_mn);

    var entity_ind = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 2 });
    defer entity_ind.deinit();
    _ = entity_ind.updateBar(&Bar{ .time = 1, .open = 10, .high = 12, .low = 9, .close = 11, .volume = 100 });
    const bar_out = entity_ind.updateBar(&Bar{ .time = 2, .open = 11, .high = 14, .low = 10, .close = 13, .volume = 100 });
    try testing.expectEqual(@as(f64, 13.0), bar_out.values[0].scalar.value);

    const skipped = entity_ind.updateScalar(&Scalar{ .time = 3, .value = math.nan(f64) });
    try testing.expect(math.isNan(skipped.values[0].scalar.value));
    try testing.expect(math.isNan(skipped.values[1].scalar.value));
}
