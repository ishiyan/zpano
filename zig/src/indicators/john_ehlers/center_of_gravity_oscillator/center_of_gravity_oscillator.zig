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
    InvalidSample,
    MnemonicTooLong,
    OutOfMemory,
};

/// Enumerates the outputs of the center of gravity oscillator.
pub const CenterOfGravityOscillatorOutput = enum(u8) {
    value = 1,
    trigger = 2,
};

/// Parameters to create a CenterOfGravityOscillator.
pub const CenterOfGravityOscillatorParams = struct {
    /// Observation window length (>= 1). Default is 10.
    length: i32 = 10,
    /// Zero-centered 2004 book baseline (-Num/Denom + (length+1)/2) when true; May 2002 article baseline (-Num/Denom) when false.
    centered: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Params = CenterOfGravityOscillatorParams;

pub fn defaultParams() CenterOfGravityOscillatorParams {
    return .{};
}

fn formatMnemonicWithPrefix(
    prefix: []const u8,
    buffer: []u8,
    params: CenterOfGravityOscillatorParams,
) Error![]const u8 {
    var triple_storage: [64]u8 = undefined;
    const suffix = component_triple_mnemonic_mod.componentTripleMnemonic(
        &triple_storage,
        params.bar_component orelse bar_component.default_bar_component,
        params.quote_component orelse quote_component.default_quote_component,
        params.trade_component orelse trade_component.default_trade_component,
    );
    const centered_tag: []const u8 = if (params.centered) ", centered" else "";
    return std.fmt.bufPrint(buffer, "{s}({d}{s}{s})", .{
        prefix,
        params.length,
        centered_tag,
        suffix,
    }) catch return error.MnemonicTooLong;
}

pub fn mnemonic(buffer: []u8, params: CenterOfGravityOscillatorParams) Error![]const u8 {
    return formatMnemonicWithPrefix("cog", buffer, params);
}

pub fn triggerMnemonic(buffer: []u8, params: CenterOfGravityOscillatorParams) Error![]const u8 {
    return formatMnemonicWithPrefix("cogTrig", buffer, params);
}

/// Ehlers' Center of Gravity Oscillator (COG).
///
/// Supports both the May 2002 TASC / Mesa convention (`centered = false`,
/// `-Num / Denom`) and the 2004 book convention (`centered = true`,
/// `-Num / Denom + (length + 1) / 2`), exposing the oscillator value on
/// observation `length` and its one-observation delayed trigger on
/// observation `length + 1`.
pub const CenterOfGravityOscillator = struct {
    allocator: std.mem.Allocator,
    window_length: usize,
    is_centered: bool,
    offset: f64,

    window: []f64,
    next: usize,
    count: usize,
    last_value: f64,
    current_value: f64,
    current_trigger: f64,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    mnemonic_storage: [128]u8,
    mnemonic_len: usize,
    description_storage: [256]u8,
    description_len: usize,
    mnemonic_trig_storage: [128]u8,
    mnemonic_trig_len: usize,
    description_trig_storage: [256]u8,
    description_trig_len: usize,

    pub fn initDefault(allocator: std.mem.Allocator) Error!CenterOfGravityOscillator {
        return init(allocator, defaultParams());
    }

    pub fn init(allocator: std.mem.Allocator, params: CenterOfGravityOscillatorParams) Error!CenterOfGravityOscillator {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const window_length: usize = @intCast(params.length);
        const window = allocator.alloc(f64, window_length) catch return error.OutOfMemory;
        errdefer allocator.free(window);

        const bar_comp = params.bar_component orelse bar_component.default_bar_component;
        const quote_comp = params.quote_component orelse quote_component.default_quote_component;
        const trade_comp = params.trade_component orelse trade_component.default_trade_component;

        var mnemonic_storage: [128]u8 = undefined;
        const mn = try mnemonic(&mnemonic_storage, params);
        const mnemonic_len = mn.len;

        var description_storage: [256]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Center of Gravity oscillator {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_trig_storage: [128]u8 = undefined;
        const mn_trig = try triggerMnemonic(&mnemonic_trig_storage, params);
        const mnemonic_trig_len = mn_trig.len;

        var description_trig_storage: [256]u8 = undefined;
        const desc_trig = std.fmt.bufPrint(&description_trig_storage, "Center of Gravity trigger {s}", .{mn_trig}) catch
            return error.MnemonicTooLong;
        const description_trig_len = desc_trig.len;

        const offset: f64 = if (params.centered)
            @as(f64, @floatFromInt(window_length + 1)) / 2.0
        else
            0.0;

        var instance = CenterOfGravityOscillator{
            .allocator = allocator,
            .window_length = window_length,
            .is_centered = params.centered,
            .offset = offset,
            .window = window,
            .next = 0,
            .count = 0,
            .last_value = 0.0,
            .current_value = math.nan(f64),
            .current_trigger = math.nan(f64),
            .bar_func = bar_component.componentValue(bar_comp),
            .quote_func = quote_component.componentValue(quote_comp),
            .trade_func = trade_component.componentValue(trade_comp),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mnemonic_len,
            .description_storage = description_storage,
            .description_len = description_len,
            .mnemonic_trig_storage = mnemonic_trig_storage,
            .mnemonic_trig_len = mnemonic_trig_len,
            .description_trig_storage = description_trig_storage,
            .description_trig_len = description_trig_len,
        };
        instance.reset();
        return instance;
    }

    pub fn deinit(self: *CenterOfGravityOscillator) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *CenterOfGravityOscillator) void {
        _ = self;
    }

    pub fn reset(self: *CenterOfGravityOscillator) void {
        @memset(self.window, 0.0);
        self.next = 0;
        self.count = 0;
        self.last_value = 0.0;
        self.current_value = math.nan(f64);
        self.current_trigger = math.nan(f64);
    }

    pub fn length(self: *const CenterOfGravityOscillator) usize {
        return self.window_length;
    }

    pub fn centered(self: *const CenterOfGravityOscillator) bool {
        return self.is_centered;
    }

    pub fn value(self: *const CenterOfGravityOscillator) f64 {
        return self.current_value;
    }

    pub fn trigger(self: *const CenterOfGravityOscillator) f64 {
        return self.current_trigger;
    }

    pub fn isPrimed(self: *const CenterOfGravityOscillator) bool {
        return self.count >= self.window_length;
    }

    pub fn isTriggerPrimed(self: *const CenterOfGravityOscillator) bool {
        return self.count > self.window_length;
    }

    pub fn getMetadata(self: *const CenterOfGravityOscillator, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .center_of_gravity_oscillator,
            self.mnemonic_storage[0..self.mnemonic_len],
            self.description_storage[0..self.description_len],
            &[_]build_metadata_mod.OutputText{
                .{
                    .mnemonic = self.mnemonic_storage[0..self.mnemonic_len],
                    .description = self.description_storage[0..self.description_len],
                },
                .{
                    .mnemonic = self.mnemonic_trig_storage[0..self.mnemonic_trig_len],
                    .description = self.description_trig_storage[0..self.description_trig_len],
                },
            },
        );
    }

    pub fn tryUpdate(self: *CenterOfGravityOscillator, sample: f64) Error!f64 {
        if (math.isInf(sample)) {
            return error.InvalidSample;
        }
        if (math.isNan(sample)) {
            return math.nan(f64);
        }

        const next_trigger = if (self.count >= self.window_length)
            self.last_value
        else
            math.nan(f64);

        self.window[self.next] = sample;
        self.next = (self.next + 1) % self.window_length;
        if (self.count <= self.window_length) {
            self.count += 1;
        }
        if (self.count < self.window_length) {
            return math.nan(f64);
        }

        var numerator: f64 = 0.0;
        var denominator: f64 = 0.0;
        for (0..self.window_length) |lag| {
            const index = (self.next + self.window_length - 1 - lag) % self.window_length;
            const item = self.window[index];
            numerator += @as(f64, @floatFromInt(1 + lag)) * item;
            denominator += item;
        }

        if (denominator != 0.0) {
            self.last_value = -numerator / denominator + self.offset;
        }
        self.current_value = self.last_value;
        self.current_trigger = next_trigger;
        return self.current_value;
    }

    pub fn update(self: *CenterOfGravityOscillator, sample: f64) f64 {
        return self.tryUpdate(sample) catch @panic("invalid center of gravity oscillator sample");
    }

    pub fn updateScalar(self: *CenterOfGravityOscillator, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *CenterOfGravityOscillator, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *CenterOfGravityOscillator, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *CenterOfGravityOscillator, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *CenterOfGravityOscillator, stamp: i64, sample: f64) OutputArray {
        const cog = self.update(sample);
        const trig = if (math.isNan(cog)) math.nan(f64) else self.current_trigger;

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = stamp, .value = cog } });
        out.append(.{ .scalar = .{ .time = stamp, .value = trig } });
        return out;
    }

    pub fn indicator(self: *CenterOfGravityOscillator) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(CenterOfGravityOscillator);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");

const ReferenceCase = struct {
    length: i32,
    centered: bool,
    expected_value: *const fn () [252]f64,
    expected_trigger: *const fn () [252]f64,
};

const reference_cases = [_]ReferenceCase{
    .{ .length = 1, .centered = false, .expected_value = &testdata.expectedL1_C0_CG, .expected_trigger = &testdata.expectedL1_C0_TR },
    .{ .length = 1, .centered = true, .expected_value = &testdata.expectedL1_C1_CG, .expected_trigger = &testdata.expectedL1_C1_TR },
    .{ .length = 2, .centered = false, .expected_value = &testdata.expectedL2_C0_CG, .expected_trigger = &testdata.expectedL2_C0_TR },
    .{ .length = 2, .centered = true, .expected_value = &testdata.expectedL2_C1_CG, .expected_trigger = &testdata.expectedL2_C1_TR },
    .{ .length = 3, .centered = false, .expected_value = &testdata.expectedL3_C0_CG, .expected_trigger = &testdata.expectedL3_C0_TR },
    .{ .length = 3, .centered = true, .expected_value = &testdata.expectedL3_C1_CG, .expected_trigger = &testdata.expectedL3_C1_TR },
    .{ .length = 5, .centered = false, .expected_value = &testdata.expectedL5_C0_CG, .expected_trigger = &testdata.expectedL5_C0_TR },
    .{ .length = 5, .centered = true, .expected_value = &testdata.expectedL5_C1_CG, .expected_trigger = &testdata.expectedL5_C1_TR },
    .{ .length = 8, .centered = false, .expected_value = &testdata.expectedL8_C0_CG, .expected_trigger = &testdata.expectedL8_C0_TR },
    .{ .length = 8, .centered = true, .expected_value = &testdata.expectedL8_C1_CG, .expected_trigger = &testdata.expectedL8_C1_TR },
    .{ .length = 10, .centered = false, .expected_value = &testdata.expectedL10_C0_CG, .expected_trigger = &testdata.expectedL10_C0_TR },
    .{ .length = 10, .centered = true, .expected_value = &testdata.expectedL10_C1_CG, .expected_trigger = &testdata.expectedL10_C1_TR },
    .{ .length = 14, .centered = false, .expected_value = &testdata.expectedL14_C0_CG, .expected_trigger = &testdata.expectedL14_C0_TR },
    .{ .length = 14, .centered = true, .expected_value = &testdata.expectedL14_C1_CG, .expected_trigger = &testdata.expectedL14_C1_TR },
    .{ .length = 20, .centered = false, .expected_value = &testdata.expectedL20_C0_CG, .expected_trigger = &testdata.expectedL20_C0_TR },
    .{ .length = 20, .centered = true, .expected_value = &testdata.expectedL20_C1_CG, .expected_trigger = &testdata.expectedL20_C1_TR },
    .{ .length = 32, .centered = false, .expected_value = &testdata.expectedL32_C0_CG, .expected_trigger = &testdata.expectedL32_C0_TR },
    .{ .length = 32, .centered = true, .expected_value = &testdata.expectedL32_C1_CG, .expected_trigger = &testdata.expectedL32_C1_TR },
    .{ .length = 64, .centered = false, .expected_value = &testdata.expectedL64_C0_CG, .expected_trigger = &testdata.expectedL64_C0_TR },
    .{ .length = 64, .centered = true, .expected_value = &testdata.expectedL64_C1_CG, .expected_trigger = &testdata.expectedL64_C1_TR },
    .{ .length = 126, .centered = false, .expected_value = &testdata.expectedL126_C0_CG, .expected_trigger = &testdata.expectedL126_C0_TR },
    .{ .length = 126, .centered = true, .expected_value = &testdata.expectedL126_C1_CG, .expected_trigger = &testdata.expectedL126_C1_TR },
    .{ .length = 252, .centered = false, .expected_value = &testdata.expectedL252_C0_CG, .expected_trigger = &testdata.expectedL252_C0_TR },
    .{ .length = 252, .centered = true, .expected_value = &testdata.expectedL252_C1_CG, .expected_trigger = &testdata.expectedL252_C1_TR },
};

fn assertCloseOrNaN(expected: f64, actual: f64, tolerance: f64) !void {
    if (math.isNan(expected)) {
        try testing.expect(math.isNan(actual));
        return;
    }
    try testing.expect(!math.isNan(actual));
    const bound = @max(tolerance, @abs(expected) * tolerance);
    try testing.expect(@abs(actual - expected) <= bound);
}

fn publishedLoopOutputs(prices: []const f64, window_length: usize, is_centered: bool, out_value: *[252]f64, out_trigger: *[252]f64) void {
    var previous: f64 = 0.0;
    for (0..prices.len) |step| {
        if (step < window_length - 1) {
            out_value[step] = math.nan(f64);
            out_trigger[step] = math.nan(f64);
            continue;
        }
        const trig = if (step >= window_length) previous else math.nan(f64);
        var numerator: f64 = 0.0;
        var denominator: f64 = 0.0;
        for (0..window_length) |lag| {
            numerator += @as(f64, @floatFromInt(1 + lag)) * prices[step - lag];
            denominator += prices[step - lag];
        }
        if (denominator != 0.0) {
            previous = -numerator / denominator;
            if (is_centered) {
                previous += @as(f64, @floatFromInt(window_length + 1)) / 2.0;
            }
        }
        out_value[step] = previous;
        out_trigger[step] = trig;
    }
}

test "center of gravity oscillator matches all 24 reference cases and published loop" {
    const inputs = testdata.testInput();
    const opens = testdata.testOpen();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();

    for (reference_cases) |case| {
        var indicator = try CenterOfGravityOscillator.init(testing.allocator, .{
            .length = case.length,
            .centered = case.centered,
        });
        defer indicator.deinit();

        var bar_indicator = try CenterOfGravityOscillator.init(testing.allocator, .{
            .length = case.length,
            .centered = case.centered,
        });
        defer bar_indicator.deinit();

        const target_length: usize = @intCast(case.length);
        try testing.expectEqual(target_length, indicator.length());
        try testing.expectEqual(case.centered, indicator.centered());
        try testing.expect(!indicator.isPrimed());
        try testing.expect(!indicator.isTriggerPrimed());
        try testing.expect(math.isNan(indicator.value()));
        try testing.expect(math.isNan(indicator.trigger()));

        const expected_values = case.expected_value();
        const expected_triggers = case.expected_trigger();
        var oracle_values: [252]f64 = undefined;
        var oracle_triggers: [252]f64 = undefined;
        publishedLoopOutputs(&inputs, target_length, case.centered, &oracle_values, &oracle_triggers);

        for (inputs, 0..) |price, index| {
            const actual_value = indicator.update(price);
            const actual_trigger = indicator.trigger();

            try testing.expectEqual(index >= target_length - 1, indicator.isPrimed());
            try testing.expectEqual(index >= target_length, indicator.isTriggerPrimed());
            try assertCloseOrNaN(expected_values[index], actual_value, 1e-13);
            try assertCloseOrNaN(expected_triggers[index], actual_trigger, 1e-13);
            try assertCloseOrNaN(actual_value, indicator.value(), 0.0);
            try assertCloseOrNaN(oracle_values[index], actual_value, 0.0);
            try assertCloseOrNaN(oracle_triggers[index], actual_trigger, 0.0);

            const stamp: i64 = @intCast(1_600_000_000 + index);
            const bar_out = bar_indicator.updateBar(&.{
                .time = stamp,
                .open = opens[index],
                .high = highs[index],
                .low = lows[index],
                .close = price,
                .volume = 1000.0,
            });
            try testing.expectEqual(@as(usize, 2), bar_out.len);
            try testing.expectEqual(stamp, bar_out.values[0].scalar.time);
            try testing.expectEqual(stamp, bar_out.values[1].scalar.time);
            try assertCloseOrNaN(actual_value, bar_out.values[0].scalar.value, 0.0);
            try assertCloseOrNaN(actual_trigger, bar_out.values[1].scalar.value, 0.0);
        }
    }
}

test "center of gravity oscillator mathematical properties, zero/subnormal guards, NaN/Inf, reset, and metadata" {
    var article = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 3, .centered = false });
    defer article.deinit();
    var book = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 3, .centered = true });
    defer book.deinit();

    var article_value: f64 = math.nan(f64);
    var book_value: f64 = math.nan(f64);
    for ([_]f64{ 10.0, 20.0, 30.0 }) |price| {
        article_value = article.update(price);
        book_value = book.update(price);
    }
    try assertCloseOrNaN(-5.0 / 3.0, article_value, 1e-15);
    try assertCloseOrNaN(1.0 / 3.0, book_value, 1e-15);
    try testing.expect(math.isNan(article.trigger()));
    try assertCloseOrNaN(-18.0 / 11.0, article.update(60.0), 1e-15);
    try assertCloseOrNaN(-5.0 / 3.0, article.trigger(), 1e-15);

    for ([_]i32{ 1, 2, 3, 10 }) |len| {
        for ([_]bool{ false, true }) |is_centered| {
            var constant_ind = try CenterOfGravityOscillator.init(testing.allocator, .{
                .length = len,
                .centered = is_centered,
            });
            defer constant_ind.deinit();
            const baseline: f64 = if (is_centered) 0.0 else -@as(f64, @floatFromInt(len + 1)) / 2.0;
            const target_len: usize = @intCast(len);
            for (0..target_len + 2) |step| {
                try assertCloseOrNaN(if (step < target_len - 1) math.nan(f64) else baseline, constant_ind.update(8.0), 1e-14);
                try assertCloseOrNaN(if (step < target_len) math.nan(f64) else baseline, constant_ind.trigger(), 1e-14);
            }
        }
    }

    for ([_]bool{ false, true }) |is_centered| {
        var initial_zero = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 2, .centered = is_centered });
        defer initial_zero.deinit();
        _ = initial_zero.update(1.0);
        try testing.expectEqual(@as(f64, 0.0), initial_zero.update(-1.0));
        try testing.expect(math.isNan(initial_zero.trigger()));

        var held_zero = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 2, .centered = is_centered });
        defer held_zero.deinit();
        _ = held_zero.update(1.0);
        const previous = held_zero.update(2.0);
        try testing.expectEqual(previous, held_zero.update(-2.0));
        try testing.expectEqual(previous, held_zero.trigger());
        const resumed = held_zero.update(4.0);
        try testing.expectEqual(previous, held_zero.trigger());
        try testing.expectEqual(if (is_centered) @as(f64, 1.5) else @as(f64, 0.0), resumed);
    }

    var subnormal = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 2 });
    defer subnormal.deinit();
    _ = subnormal.update(1e-309);
    try assertCloseOrNaN(-1.25, subnormal.update(3e-309), 1e-14);

    var skipped = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 3 });
    defer skipped.deinit();
    var uninterrupted = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 3 });
    defer uninterrupted.deinit();
    for ([_]f64{ 1.0, 2.0, 3.0, 4.0, 5.0 }) |price| {
        try testing.expect(math.isNan(skipped.update(math.nan(f64))));
        const gap = skipped.updateScalar(&.{ .time = 100, .value = math.nan(f64) });
        try testing.expect(math.isNan(gap.values[0].scalar.value));
        try testing.expect(math.isNan(gap.values[1].scalar.value));
        const actual = skipped.update(price);
        const reference = uninterrupted.update(price);
        try assertCloseOrNaN(reference, actual, 0.0);
        try assertCloseOrNaN(uninterrupted.trigger(), skipped.trigger(), 0.0);
    }
    try testing.expect(skipped.isTriggerPrimed());

    var single = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 1 });
    defer single.deinit();
    _ = single.update(4.0);
    try testing.expectError(Error.InvalidSample, single.tryUpdate(math.inf(f64)));
    try testing.expectError(Error.InvalidSample, single.tryUpdate(-math.inf(f64)));
    try testing.expectEqual(@as(f64, -1.0), single.update(5.0));
    try testing.expectEqual(@as(f64, -1.0), single.trigger());

    var resettable = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 2, .centered = true });
    defer resettable.deinit();
    for ([_]f64{ 1.0, 2.0, 3.0 }) |price| {
        _ = resettable.update(price);
    }
    resettable.reset();
    try testing.expect(!resettable.isPrimed());
    try testing.expect(!resettable.isTriggerPrimed());
    try testing.expect(math.isNan(resettable.value()));
    try testing.expect(math.isNan(resettable.trigger()));
    try testing.expect(math.isNan(resettable.update(1.0)));
    try testing.expectEqual(@as(f64, 0.0), resettable.update(-1.0));

    try testing.expectError(Error.InvalidLength, CenterOfGravityOscillator.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(Error.InvalidLength, CenterOfGravityOscillator.init(testing.allocator, .{ .length = -5 }));

    var default_ind = try CenterOfGravityOscillator.initDefault(testing.allocator);
    defer default_ind.deinit();
    var meta: Metadata = undefined;
    default_ind.getMetadata(&meta);
    try testing.expectEqual(Identifier.center_of_gravity_oscillator, meta.identifier);
    try testing.expectEqualStrings("cog(10)", meta.mnemonic);
    try testing.expectEqualStrings("Center of Gravity oscillator cog(10)", meta.description);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@intFromEnum(CenterOfGravityOscillatorOutput.value), meta.outputs_buf[0].kind);
    try testing.expectEqual(shape_mod.Shape.scalar, meta.outputs_buf[0].shape);
    try testing.expectEqualStrings("cog(10)", meta.outputs_buf[0].mnemonic);
    try testing.expectEqual(@intFromEnum(CenterOfGravityOscillatorOutput.trigger), meta.outputs_buf[1].kind);
    try testing.expectEqual(shape_mod.Shape.scalar, meta.outputs_buf[1].shape);
    try testing.expectEqualStrings("cogTrig(10)", meta.outputs_buf[1].mnemonic);
    try testing.expectEqualStrings("Center of Gravity trigger cogTrig(10)", meta.outputs_buf[1].description);

    var mn_storage: [128]u8 = undefined;
    var mn_trig_storage: [128]u8 = undefined;
    const custom_params = CenterOfGravityOscillatorParams{
        .length = 14,
        .centered = true,
        .bar_component = .median,
        .quote_component = .bid,
        .trade_component = .volume,
    };
    try testing.expectEqualStrings("cog(14, centered, hl/2, b, v)", try mnemonic(&mn_storage, custom_params));
    try testing.expectEqualStrings("cogTrig(14, centered, hl/2, b, v)", try triggerMnemonic(&mn_trig_storage, custom_params));

    const inputs = testdata.testInput();
    const opens = testdata.testOpen();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();
    var bar_median = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 4, .centered = true, .bar_component = .median });
    defer bar_median.deinit();
    var quote_mid = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 4, .centered = true });
    defer quote_mid.deinit();
    var trade_price = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 4, .centered = true });
    defer trade_price.deinit();
    var scalar_direct = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 4, .centered = true });
    defer scalar_direct.deinit();

    for (0..12) |step| {
        const stamp: i64 = @intCast(1_700_000_000 + step);
        const median = (highs[step] + lows[step]) / 2.0;
        const bar_out = bar_median.updateBar(&.{ .time = stamp, .open = opens[step], .high = highs[step], .low = lows[step], .close = inputs[step], .volume = 1000.0 });
        const quote_out = quote_mid.updateQuote(&.{ .time = stamp, .bid_price = lows[step], .ask_price = highs[step], .bid_size = 1.0, .ask_size = 1.0 });
        const trade_out = trade_price.updateTrade(&.{ .time = stamp, .price = median, .volume = 10.0 });
        const scalar_out = scalar_direct.updateScalar(&.{ .time = stamp, .value = median });
        for (0..2) |slot| {
            const expected = scalar_out.values[slot].scalar.value;
            try assertCloseOrNaN(expected, bar_out.values[slot].scalar.value, 1e-15);
            try assertCloseOrNaN(expected, quote_out.values[slot].scalar.value, 1e-15);
            try assertCloseOrNaN(expected, trade_out.values[slot].scalar.value, 1e-15);
        }
    }
}
