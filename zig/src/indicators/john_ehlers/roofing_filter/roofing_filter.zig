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
const line_indicator_mod = @import("../../core/line_indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");
const shape_mod = @import("../../core/outputs/shape.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

pub const Error = error{
    InvalidShortestPeriod,
    InvalidLongestPeriod,
    InvalidHighpassPoles,
    InvalidZeroMean,
    InvalidSample,
    MnemonicTooLong,
};

pub const RoofingFilterOutput = enum(u8) {
    value = 1,
};

pub const RoofingFilterParams = struct {
    shortest_cycle_period: f64 = 10.0,
    longest_cycle_period: f64 = 48.0,
    highpass_poles: i32 = 2,
    zero_mean: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Params = RoofingFilterParams;

pub fn defaultParams() RoofingFilterParams {
    return .{};
}

fn normalizeZero(number: f64) f64 {
    return if (number == 0.0) 0.0 else number;
}

fn smootherCoefficients(period: f64) struct { c1: f64, c2: f64, c3: f64 } {
    const smoother_angle = math.sqrt2 * math.pi / period;
    const a1 = @exp(-smoother_angle);
    const c2 = 2.0 * a1 * @cos(smoother_angle);
    const c3 = -a1 * a1;
    return .{
        .c1 = 1.0 - c2 - c3,
        .c2 = c2,
        .c3 = c3,
    };
}

pub fn mnemonic(buffer: []u8, params: RoofingFilterParams) Error![]const u8 {
    var triple_storage: [64]u8 = undefined;
    const suffix = component_triple_mnemonic_mod.componentTripleMnemonic(
        &triple_storage,
        params.bar_component orelse bar_component.default_bar_component,
        params.quote_component orelse quote_component.default_quote_component,
        params.trade_component orelse trade_component.default_trade_component,
    );
    const zm_tag: []const u8 = if (params.zero_mean) "zm" else "";
    return std.fmt.bufPrint(buffer, "roof{d}hp{s}({d}, {d}{s})", .{
        params.highpass_poles,
        zm_tag,
        normalizeZero(params.shortest_cycle_period),
        normalizeZero(params.longest_cycle_period),
        suffix,
    }) catch return error.MnemonicTooLong;
}

pub const formatMnemonic = mnemonic;

pub const RoofingFilter = struct {
    line: LineIndicator,
    poles: usize,
    zero_mean: bool,
    warmup_count: usize,
    gain: f64,
    decay: f64,
    c1: f64,
    c2: f64,
    c3: f64,

    count: usize,
    input1: f64,
    input2: f64,
    hp1: f64,
    hp2: f64,
    filt1: f64,
    filt2: f64,
    zero1: f64,
    current_value: f64,

    mnemonic_storage: [64]u8,
    mnemonic_len: usize,
    description_storage: [128]u8,
    description_len: usize,

    pub fn init(params: RoofingFilterParams) Error!RoofingFilter {
        const shortest = params.shortest_cycle_period;
        if (!math.isFinite(shortest) or shortest < 2.0) {
            return error.InvalidShortestPeriod;
        }

        const longest = params.longest_cycle_period;
        if (!math.isFinite(longest) or longest <= shortest) {
            return error.InvalidLongestPeriod;
        }

        if (params.highpass_poles != 1 and params.highpass_poles != 2) {
            return error.InvalidHighpassPoles;
        }
        const poles: usize = @intCast(params.highpass_poles);

        if (params.zero_mean and poles != 1) {
            return error.InvalidZeroMean;
        }

        const angle = if (poles == 1)
            2.0 * math.pi / longest
        else
            math.sqrt2 * math.pi / longest;
        const tangent = @tan(angle / 2.0);
        const alpha = 2.0 * tangent / (1.0 + tangent);
        const gain = 1.0 - alpha / 2.0;
        const decay = 1.0 - alpha;
        const smoother = smootherCoefficients(shortest);
        const warmup_count: usize = if (poles == 2 or params.zero_mean) 4 else 3;

        var mnemonic_storage: [64]u8 = undefined;
        const mnemonic_slice = try mnemonic(&mnemonic_storage, params);
        const mnemonic_len = mnemonic_slice.len;

        var description_storage: [128]u8 = undefined;
        const description_slice = std.fmt.bufPrint(&description_storage, "Roofing Filter {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = description_slice.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_storage[0..mnemonic_len],
                description_storage[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .poles = poles,
            .zero_mean = params.zero_mean,
            .warmup_count = warmup_count,
            .gain = gain,
            .decay = decay,
            .c1 = smoother.c1,
            .c2 = smoother.c2,
            .c3 = smoother.c3,
            .count = 0,
            .input1 = 0.0,
            .input2 = 0.0,
            .hp1 = 0.0,
            .hp2 = 0.0,
            .filt1 = 0.0,
            .filt2 = 0.0,
            .zero1 = 0.0,
            .current_value = math.nan(f64),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mnemonic_len,
            .description_storage = description_storage,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *RoofingFilter) void {
        self.line.mnemonic = self.mnemonic_storage[0..self.mnemonic_len];
        self.line.description = self.description_storage[0..self.description_len];
    }

    pub fn reset(self: *RoofingFilter) void {
        self.count = 0;
        self.input1 = 0.0;
        self.input2 = 0.0;
        self.hp1 = 0.0;
        self.hp2 = 0.0;
        self.filt1 = 0.0;
        self.filt2 = 0.0;
        self.zero1 = 0.0;
        self.current_value = math.nan(f64);
    }

    pub fn warmupBars(self: *const RoofingFilter) usize {
        return self.warmup_count;
    }

    pub fn value(self: *const RoofingFilter) f64 {
        return self.current_value;
    }

    pub fn isPrimed(self: *const RoofingFilter) bool {
        return self.count > self.warmup_count;
    }

    pub fn getMetadata(self: *const RoofingFilter, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .roofing_filter,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn tryUpdate(self: *RoofingFilter, sample: f64) Error!f64 {
        if (math.isNan(sample)) {
            return math.nan(f64);
        }
        if (math.isInf(sample)) {
            return error.InvalidSample;
        }

        var hp: f64 = 0.0;
        var filt: f64 = 0.0;
        var zero: f64 = 0.0;

        if (self.count >= self.poles) {
            if (self.poles == 2) {
                const difference = sample - 2.0 * self.input1 + self.input2;
                hp = self.gain * self.gain * difference + 2.0 * self.decay * self.hp1 - self.decay * self.decay * self.hp2;
            } else {
                hp = self.gain * (sample - self.input1) + self.decay * self.hp1;
            }
            filt = self.c1 * (hp + self.hp1) / 2.0 + self.c2 * self.filt1 + self.c3 * self.filt2;
            if (self.zero_mean) {
                zero = self.gain * (filt - self.filt1) + self.decay * self.zero1;
            }
        }

        self.input2 = self.input1;
        self.input1 = sample;
        self.hp2 = self.hp1;
        self.hp1 = hp;
        self.filt2 = self.filt1;
        self.filt1 = filt;
        self.zero1 = zero;
        self.count += 1;

        const output = if (self.zero_mean) zero else filt;
        if (self.count > self.warmup_count) {
            self.current_value = output;
            return output;
        }
        return math.nan(f64);
    }

    pub fn update(self: *RoofingFilter, sample: f64) f64 {
        return self.tryUpdate(sample) catch @panic("RoofingFilter.update: sample must be finite or NaN");
    }

    pub fn updateScalar(self: *RoofingFilter, sample: *const Scalar) OutputArray {
        const output = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, output);
    }

    pub fn updateBar(self: *RoofingFilter, sample: *const Bar) OutputArray {
        const output = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, output);
    }

    pub fn updateQuote(self: *RoofingFilter, sample: *const Quote) OutputArray {
        const output = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, output);
    }

    pub fn updateTrade(self: *RoofingFilter, sample: *const Trade) OutputArray {
        const output = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, output);
    }

    pub fn indicator(self: *RoofingFilter) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(RoofingFilter);
};

const testing = std.testing;
const testdata = @import("testdata.zig");

const ReferenceCase = struct {
    params: RoofingFilterParams,
    expected: [252]f64,
    warmup: usize,
};

fn buildReferenceCases() [12]ReferenceCase {
    return .{
        .{ .params = .{ .shortest_cycle_period = 10.0, .longest_cycle_period = 48.0, .highpass_poles = 2, .zero_mean = false }, .expected = testdata.expectedLP10_HP48_PO2_ZM0(), .warmup = 4 },
        .{ .params = .{ .shortest_cycle_period = 40.0, .longest_cycle_period = 80.0, .highpass_poles = 2, .zero_mean = false }, .expected = testdata.expectedLP40_HP80_PO2_ZM0(), .warmup = 4 },
        .{ .params = .{ .shortest_cycle_period = 2.0, .longest_cycle_period = 3.0, .highpass_poles = 2, .zero_mean = false }, .expected = testdata.expectedLP2_HP3_PO2_ZM0(), .warmup = 4 },
        .{ .params = .{ .shortest_cycle_period = 5.0, .longest_cycle_period = 24.0, .highpass_poles = 2, .zero_mean = false }, .expected = testdata.expectedLP5_HP24_PO2_ZM0(), .warmup = 4 },
        .{ .params = .{ .shortest_cycle_period = 10.5, .longest_cycle_period = 48.0, .highpass_poles = 2, .zero_mean = false }, .expected = testdata.expectedLP10_5_HP48_PO2_ZM0(), .warmup = 4 },
        .{ .params = .{ .shortest_cycle_period = 10.0, .longest_cycle_period = 11.0, .highpass_poles = 2, .zero_mean = false }, .expected = testdata.expectedLP10_HP11_PO2_ZM0(), .warmup = 4 },
        .{ .params = .{ .shortest_cycle_period = 20.0, .longest_cycle_period = 80.0, .highpass_poles = 2, .zero_mean = false }, .expected = testdata.expectedLP20_HP80_PO2_ZM0(), .warmup = 4 },
        .{ .params = .{ .shortest_cycle_period = 40.0, .longest_cycle_period = 252.0, .highpass_poles = 2, .zero_mean = false }, .expected = testdata.expectedLP40_HP252_PO2_ZM0(), .warmup = 4 },
        .{ .params = .{ .shortest_cycle_period = 10.0, .longest_cycle_period = 48.0, .highpass_poles = 1, .zero_mean = false }, .expected = testdata.expectedLP10_HP48_PO1_ZM0(), .warmup = 3 },
        .{ .params = .{ .shortest_cycle_period = 2.0, .longest_cycle_period = 4.0, .highpass_poles = 1, .zero_mean = false }, .expected = testdata.expectedLP2_HP4_PO1_ZM0(), .warmup = 3 },
        .{ .params = .{ .shortest_cycle_period = 10.0, .longest_cycle_period = 48.0, .highpass_poles = 1, .zero_mean = true }, .expected = testdata.expectedLP10_HP48_PO1_ZM1(), .warmup = 4 },
        .{ .params = .{ .shortest_cycle_period = 2.0, .longest_cycle_period = 4.0, .highpass_poles = 1, .zero_mean = true }, .expected = testdata.expectedLP2_HP4_PO1_ZM1(), .warmup = 4 },
    };
}

test "RoofingFilter matches all twelve 252-bar reference configurations" {
    const inputs = testdata.testInput();
    const cases = buildReferenceCases();

    for (cases) |item| {
        var filter = try RoofingFilter.init(item.params);
        filter.fixSlices();
        try testing.expectEqual(item.warmup, filter.warmupBars());
        try testing.expect(!filter.isPrimed());
        try testing.expect(math.isNan(filter.value()));

        for (inputs, 0..) |sample, index| {
            const actual = filter.update(sample);
            const target = item.expected[index];
            if (index < item.warmup) {
                try testing.expect(!filter.isPrimed());
                try testing.expect(math.isNan(actual));
                try testing.expect(math.isNan(filter.value()));
            } else {
                try testing.expect(filter.isPrimed());
                try testing.expect(!math.isNan(target));
                const bound = @max(1e-11, @abs(target) * 1e-11);
                try testing.expect(@abs(actual - target) <= bound);
                try testing.expect(@abs(filter.value() - actual) <= 1e-15);
            }
        }
    }
}

test "RoofingFilter skips NaN without state mutation and reset reproduces start" {
    const inputs = testdata.testInput();
    const cases = buildReferenceCases();

    for (cases) |item| {
        var first = try RoofingFilter.init(item.params);
        first.fixSlices();
        var second = try RoofingFilter.init(item.params);
        second.fixSlices();

        for (inputs) |sample| {
            try testing.expect(math.isNan(first.update(math.nan(f64))));
            const left = first.update(sample);
            const right = second.update(sample);
            try testing.expectEqual(second.isPrimed(), first.isPrimed());
            if (math.isNan(right)) {
                try testing.expect(math.isNan(left));
            } else {
                try testing.expect(@abs(left - right) <= 1e-15);
            }
        }

        first.reset();
        try testing.expect(!first.isPrimed());
        try testing.expect(math.isNan(first.value()));

        var third = try RoofingFilter.init(item.params);
        third.fixSlices();
        for (inputs) |sample| {
            const left = first.update(sample);
            const right = third.update(sample);
            if (math.isNan(right)) {
                try testing.expect(math.isNan(left));
            } else {
                try testing.expect(@abs(left - right) <= 1e-15);
            }
        }
    }
}

test "RoofingFilter invariants on constant input, linear ramp, and 1-pole period 4" {
    var constant_filter = try RoofingFilter.init(defaultParams());
    constant_filter.fixSlices();
    var first_finite: ?usize = null;
    var last_constant: f64 = math.nan(f64);
    for (0..500) |index| {
        const output = constant_filter.update(7.0);
        if (first_finite == null and math.isFinite(output)) {
            first_finite = index;
        }
        last_constant = output;
    }
    try testing.expectEqual(@as(?usize, constant_filter.warmupBars()), first_finite);
    try testing.expect(@abs(last_constant) <= 1e-12);

    var ramp_filter = try RoofingFilter.init(defaultParams());
    ramp_filter.fixSlices();
    var last_ramp: f64 = math.nan(f64);
    for (0..1000) |index| {
        last_ramp = ramp_filter.update(@floatFromInt(index));
    }
    try testing.expect(@abs(last_ramp) <= 1e-12);

    const boundary = try RoofingFilter.init(.{
        .shortest_cycle_period = 2.0,
        .longest_cycle_period = 4.0,
        .highpass_poles = 1,
        .zero_mean = false,
    });
    try testing.expect(@abs(boundary.gain - 0.5) <= 1e-15);
    try testing.expect(@abs(boundary.decay - 0.0) <= 1e-15);
}

test "RoofingFilter rejects invalid parameters and infinite samples" {
    try testing.expectError(error.InvalidShortestPeriod, RoofingFilter.init(.{ .shortest_cycle_period = 1.0 }));
    try testing.expectError(error.InvalidShortestPeriod, RoofingFilter.init(.{ .shortest_cycle_period = 0.0 }));
    try testing.expectError(error.InvalidShortestPeriod, RoofingFilter.init(.{ .shortest_cycle_period = -2.0 }));
    try testing.expectError(error.InvalidShortestPeriod, RoofingFilter.init(.{ .shortest_cycle_period = math.nan(f64) }));
    try testing.expectError(error.InvalidShortestPeriod, RoofingFilter.init(.{ .shortest_cycle_period = math.inf(f64) }));
    try testing.expectError(error.InvalidLongestPeriod, RoofingFilter.init(.{ .shortest_cycle_period = 10.0, .longest_cycle_period = 10.0 }));
    try testing.expectError(error.InvalidLongestPeriod, RoofingFilter.init(.{ .shortest_cycle_period = 10.0, .longest_cycle_period = 5.0 }));
    try testing.expectError(error.InvalidLongestPeriod, RoofingFilter.init(.{ .shortest_cycle_period = 10.0, .longest_cycle_period = math.nan(f64) }));
    try testing.expectError(error.InvalidLongestPeriod, RoofingFilter.init(.{ .shortest_cycle_period = 10.0, .longest_cycle_period = math.inf(f64) }));
    try testing.expectError(error.InvalidHighpassPoles, RoofingFilter.init(.{ .highpass_poles = 0 }));
    try testing.expectError(error.InvalidHighpassPoles, RoofingFilter.init(.{ .highpass_poles = 3 }));
    try testing.expectError(error.InvalidZeroMean, RoofingFilter.init(.{ .highpass_poles = 2, .zero_mean = true }));

    var filter = try RoofingFilter.init(defaultParams());
    filter.fixSlices();
    try testing.expectError(error.InvalidSample, filter.tryUpdate(math.inf(f64)));
    try testing.expectError(error.InvalidSample, filter.tryUpdate(-math.inf(f64)));
    try testing.expect(!filter.isPrimed());
}

test "RoofingFilter metadata, mnemonics, and entity updates" {
    var filter = try RoofingFilter.init(defaultParams());
    filter.fixSlices();
    var meta: Metadata = undefined;
    filter.getMetadata(&meta);
    try testing.expectEqual(Identifier.roofing_filter, meta.identifier);
    try testing.expectEqualStrings("roof2hp(10, 48)", meta.mnemonic);
    try testing.expectEqualStrings("Roofing Filter roof2hp(10, 48)", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@intFromEnum(RoofingFilterOutput.value), meta.outputs_buf[0].kind);
    try testing.expectEqual(shape_mod.Shape.scalar, meta.outputs_buf[0].shape);
    try testing.expectEqualStrings("roof2hp(10, 48)", meta.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Roofing Filter roof2hp(10, 48)", meta.outputs_buf[0].description);

    const MnemonicCase = struct {
        params: RoofingFilterParams,
        expected: []const u8,
    };
    const mnemonic_cases = [_]MnemonicCase{
        .{ .params = .{}, .expected = "roof2hp(10, 48)" },
        .{ .params = .{ .shortest_cycle_period = 10.5, .longest_cycle_period = 48.0, .highpass_poles = 2, .zero_mean = false }, .expected = "roof2hp(10.5, 48)" },
        .{ .params = .{ .shortest_cycle_period = 10.0, .longest_cycle_period = 48.0, .highpass_poles = 1, .zero_mean = false }, .expected = "roof1hp(10, 48)" },
        .{ .params = .{ .shortest_cycle_period = 10.0, .longest_cycle_period = 48.0, .highpass_poles = 1, .zero_mean = true }, .expected = "roof1hpzm(10, 48)" },
        .{ .params = .{ .bar_component = .median }, .expected = "roof2hp(10, 48, hl/2)" },
        .{ .params = .{ .quote_component = .bid }, .expected = "roof2hp(10, 48, b)" },
        .{ .params = .{ .trade_component = .volume }, .expected = "roof2hp(10, 48, v)" },
        .{ .params = .{ .bar_component = .open, .quote_component = .ask }, .expected = "roof2hp(10, 48, o, a)" },
        .{ .params = .{ .bar_component = .typical, .trade_component = .volume }, .expected = "roof2hp(10, 48, hlc/3, v)" },
        .{ .params = .{ .quote_component = .weighted, .trade_component = .volume }, .expected = "roof2hp(10, 48, (bbs+aas)/(bs+as), v)" },
    };
    for (mnemonic_cases) |item| {
        var text_storage: [64]u8 = undefined;
        const formatted = try mnemonic(&text_storage, item.params);
        try testing.expectEqualStrings(item.expected, formatted);
        var instance = try RoofingFilter.init(item.params);
        instance.fixSlices();
        var item_meta: Metadata = undefined;
        instance.getMetadata(&item_meta);
        try testing.expectEqualStrings(item.expected, item_meta.mnemonic);
    }

    const closes = testdata.testInput();
    const opens = testdata.testOpen();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();

    var bar_filter = try RoofingFilter.init(defaultParams());
    bar_filter.fixSlices();
    var scalar_filter = try RoofingFilter.init(defaultParams());
    scalar_filter.fixSlices();

    for (0..10) |index| {
        const bar = Bar{
            .time = 1000 + @as(i64, @intCast(index)),
            .open = opens[index],
            .high = highs[index],
            .low = lows[index],
            .close = closes[index],
            .volume = 1000.0,
        };
        const scalar = Scalar{
            .time = bar.time,
            .value = closes[index],
        };
        const bar_out = bar_filter.updateBar(&bar);
        const scalar_out = scalar_filter.updateScalar(&scalar);
        try testing.expectEqual(@as(usize, 1), bar_out.len);
        const bar_scalar = bar_out.values[0].scalar;
        const ref_scalar = scalar_out.values[0].scalar;
        try testing.expectEqual(bar.time, bar_scalar.time);
        if (math.isNan(ref_scalar.value)) {
            try testing.expect(math.isNan(bar_scalar.value));
        } else {
            try testing.expect(@abs(bar_scalar.value - ref_scalar.value) <= 1e-15);
        }
    }

    var quote_filter = try RoofingFilter.init(defaultParams());
    quote_filter.fixSlices();
    var quote_out: OutputArray = undefined;
    for (0..5) |_| {
        const quote = Quote{ .time = 2000, .bid_price = 99.0, .ask_price = 101.0, .bid_size = 1.0, .ask_size = 1.0 };
        quote_out = quote_filter.updateQuote(&quote);
    }
    try testing.expectEqual(@as(usize, 1), quote_out.len);
    try testing.expect(!math.isNan(quote_out.values[0].scalar.value));

    var trade_filter = try RoofingFilter.init(defaultParams());
    trade_filter.fixSlices();
    var trade_out: OutputArray = undefined;
    for (0..5) |_| {
        const trade = Trade{ .time = 3000, .price = 100.0, .volume = 50.0 };
        trade_out = trade_filter.updateTrade(&trade);
    }
    try testing.expectEqual(@as(usize, 1), trade_out.len);
    try testing.expect(!math.isNan(trade_out.values[0].scalar.value));
}
