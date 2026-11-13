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
    InvalidSample,
    MnemonicTooLong,
};

pub const SuperSmootherOutput = enum(u8) {
    value = 1,
};

pub const SuperSmootherParams = struct {
    shortest_cycle_period: f64 = 10.0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Params = SuperSmootherParams;

pub fn defaultParams() SuperSmootherParams {
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

pub fn mnemonic(buffer: []u8, params: SuperSmootherParams) Error![]const u8 {
    var triple_storage: [64]u8 = undefined;
    const suffix = component_triple_mnemonic_mod.componentTripleMnemonic(
        &triple_storage,
        params.bar_component orelse bar_component.default_bar_component,
        params.quote_component orelse quote_component.default_quote_component,
        params.trade_component orelse trade_component.default_trade_component,
    );
    return std.fmt.bufPrint(buffer, "ss({d}{s})", .{
        normalizeZero(params.shortest_cycle_period),
        suffix,
    }) catch return error.MnemonicTooLong;
}

pub const formatMnemonic = mnemonic;

pub const SuperSmoother = struct {
    line: LineIndicator,
    warmup_count: usize,
    c1: f64,
    c2: f64,
    c3: f64,

    count: usize,
    input1: f64,
    output1: f64,
    output2: f64,
    current_value: f64,

    mnemonic_storage: [64]u8,
    mnemonic_len: usize,
    description_storage: [128]u8,
    description_len: usize,

    pub fn init(params: SuperSmootherParams) Error!SuperSmoother {
        const shortest = params.shortest_cycle_period;
        if (!math.isFinite(shortest) or shortest < 2.0) {
            return error.InvalidShortestPeriod;
        }

        const smoother = smootherCoefficients(shortest);

        var mnemonic_storage: [64]u8 = undefined;
        const mnemonic_slice = try mnemonic(&mnemonic_storage, params);
        const mnemonic_len = mnemonic_slice.len;

        var description_storage: [128]u8 = undefined;
        const description_slice = std.fmt.bufPrint(&description_storage, "Super Smoother {s}", .{mnemonic_slice}) catch
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
            .warmup_count = 2,
            .c1 = smoother.c1,
            .c2 = smoother.c2,
            .c3 = smoother.c3,
            .count = 0,
            .input1 = 0.0,
            .output1 = 0.0,
            .output2 = 0.0,
            .current_value = math.nan(f64),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mnemonic_len,
            .description_storage = description_storage,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *SuperSmoother) void {
        self.line.mnemonic = self.mnemonic_storage[0..self.mnemonic_len];
        self.line.description = self.description_storage[0..self.description_len];
    }

    pub fn reset(self: *SuperSmoother) void {
        self.count = 0;
        self.input1 = 0.0;
        self.output1 = 0.0;
        self.output2 = 0.0;
        self.current_value = math.nan(f64);
    }

    pub fn warmupBars(self: *const SuperSmoother) usize {
        return self.warmup_count;
    }

    pub fn value(self: *const SuperSmoother) f64 {
        return self.current_value;
    }

    pub fn isPrimed(self: *const SuperSmoother) bool {
        return self.count > self.warmup_count;
    }

    pub fn getMetadata(self: *const SuperSmoother, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .super_smoother,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn tryUpdate(self: *SuperSmoother, sample: f64) Error!f64 {
        if (math.isNan(sample)) {
            return math.nan(f64);
        }
        if (math.isInf(sample)) {
            return error.InvalidSample;
        }

        if (self.count == 0) {
            self.input1 = sample;
            self.output1 = sample;
            self.output2 = sample;
        }

        const output = self.c1 * (sample + self.input1) / 2.0 + self.c2 * self.output1 + self.c3 * self.output2;
        self.input1 = sample;
        self.output2 = self.output1;
        self.output1 = output;
        self.count += 1;

        if (self.count > self.warmup_count) {
            self.current_value = output;
            return output;
        }
        return math.nan(f64);
    }

    pub fn update(self: *SuperSmoother, sample: f64) f64 {
        return self.tryUpdate(sample) catch @panic("SuperSmoother.update: sample must be finite or NaN");
    }

    pub fn updateScalar(self: *SuperSmoother, sample: *const Scalar) OutputArray {
        const output = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, output);
    }

    pub fn updateBar(self: *SuperSmoother, sample: *const Bar) OutputArray {
        const output = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, output);
    }

    pub fn updateQuote(self: *SuperSmoother, sample: *const Quote) OutputArray {
        const output = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, output);
    }

    pub fn updateTrade(self: *SuperSmoother, sample: *const Trade) OutputArray {
        const output = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, output);
    }

    pub fn indicator(self: *SuperSmoother) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(SuperSmoother);
};

const testing = std.testing;
const testdata = @import("testdata.zig");

const ReferenceCase = struct {
    params: SuperSmootherParams,
    expected: [252]f64,
    warmup: usize,
};

fn buildReferenceCases() [10]ReferenceCase {
    return .{
        .{ .params = .{ .shortest_cycle_period = 2.0 }, .expected = testdata.expectedP2(), .warmup = 2 },
        .{ .params = .{ .shortest_cycle_period = 3.0 }, .expected = testdata.expectedP3(), .warmup = 2 },
        .{ .params = .{ .shortest_cycle_period = 5.0 }, .expected = testdata.expectedP5(), .warmup = 2 },
        .{ .params = .{ .shortest_cycle_period = 8.0 }, .expected = testdata.expectedP8(), .warmup = 2 },
        .{ .params = .{ .shortest_cycle_period = 10.0 }, .expected = testdata.expectedP10(), .warmup = 2 },
        .{ .params = .{ .shortest_cycle_period = 10.5 }, .expected = testdata.expectedP10_5(), .warmup = 2 },
        .{ .params = .{ .shortest_cycle_period = 20.0 }, .expected = testdata.expectedP20(), .warmup = 2 },
        .{ .params = .{ .shortest_cycle_period = 40.0 }, .expected = testdata.expectedP40(), .warmup = 2 },
        .{ .params = .{ .shortest_cycle_period = 80.0 }, .expected = testdata.expectedP80(), .warmup = 2 },
        .{ .params = .{ .shortest_cycle_period = 252.0 }, .expected = testdata.expectedP252(), .warmup = 2 },
    };
}

test "SuperSmoother matches all ten 252-bar reference configurations" {
    const inputs = testdata.testInput();
    const cases = buildReferenceCases();

    for (cases) |item| {
        var filter = try SuperSmoother.init(item.params);
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

test "SuperSmoother skips NaN without state mutation and reset reproduces start" {
    const inputs = testdata.testInput();
    const cases = buildReferenceCases();

    for (cases) |item| {
        var first = try SuperSmoother.init(item.params);
        first.fixSlices();
        var second = try SuperSmoother.init(item.params);
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

        var third = try SuperSmoother.init(item.params);
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

test "SuperSmoother invariants on constant input and Nyquist zero" {
    var constant_filter = try SuperSmoother.init(defaultParams());
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
    try testing.expect(@abs(last_constant - 7.0) <= 1e-12);

    var nyquist_filter = try SuperSmoother.init(defaultParams());
    nyquist_filter.fixSlices();
    var last_nyquist: f64 = math.nan(f64);
    for (0..500) |index| {
        const sample: f64 = if (index % 2 == 0) 1.0 else -1.0;
        last_nyquist = nyquist_filter.update(sample);
    }
    try testing.expect(@abs(last_nyquist) <= 1e-12);
}

test "SuperSmoother rejects invalid parameters and infinite samples" {
    try testing.expectError(error.InvalidShortestPeriod, SuperSmoother.init(.{ .shortest_cycle_period = 1.0 }));
    try testing.expectError(error.InvalidShortestPeriod, SuperSmoother.init(.{ .shortest_cycle_period = 1.999 }));
    try testing.expectError(error.InvalidShortestPeriod, SuperSmoother.init(.{ .shortest_cycle_period = 0.0 }));
    try testing.expectError(error.InvalidShortestPeriod, SuperSmoother.init(.{ .shortest_cycle_period = -2.0 }));
    try testing.expectError(error.InvalidShortestPeriod, SuperSmoother.init(.{ .shortest_cycle_period = math.nan(f64) }));
    try testing.expectError(error.InvalidShortestPeriod, SuperSmoother.init(.{ .shortest_cycle_period = math.inf(f64) }));
    try testing.expectError(error.InvalidShortestPeriod, SuperSmoother.init(.{ .shortest_cycle_period = -math.inf(f64) }));

    var filter = try SuperSmoother.init(defaultParams());
    filter.fixSlices();
    try testing.expectError(error.InvalidSample, filter.tryUpdate(math.inf(f64)));
    try testing.expectError(error.InvalidSample, filter.tryUpdate(-math.inf(f64)));
    try testing.expect(!filter.isPrimed());
}

test "SuperSmoother metadata, mnemonics, and entity updates" {
    var filter = try SuperSmoother.init(defaultParams());
    filter.fixSlices();
    var meta: Metadata = undefined;
    filter.getMetadata(&meta);
    try testing.expectEqual(Identifier.super_smoother, meta.identifier);
    try testing.expectEqualStrings("ss(10)", meta.mnemonic);
    try testing.expectEqualStrings("Super Smoother ss(10)", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@intFromEnum(SuperSmootherOutput.value), meta.outputs_buf[0].kind);
    try testing.expectEqual(shape_mod.Shape.scalar, meta.outputs_buf[0].shape);
    try testing.expectEqualStrings("ss(10)", meta.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Super Smoother ss(10)", meta.outputs_buf[0].description);

    const MnemonicCase = struct {
        params: SuperSmootherParams,
        expected: []const u8,
    };
    const mnemonic_cases = [_]MnemonicCase{
        .{ .params = .{}, .expected = "ss(10)" },
        .{ .params = .{ .shortest_cycle_period = 10.5 }, .expected = "ss(10.5)" },
        .{ .params = .{ .bar_component = .median }, .expected = "ss(10, hl/2)" },
        .{ .params = .{ .quote_component = .bid }, .expected = "ss(10, b)" },
        .{ .params = .{ .trade_component = .volume }, .expected = "ss(10, v)" },
        .{ .params = .{ .bar_component = .open, .quote_component = .ask }, .expected = "ss(10, o, a)" },
        .{ .params = .{ .bar_component = .typical, .trade_component = .volume }, .expected = "ss(10, hlc/3, v)" },
        .{ .params = .{ .quote_component = .weighted, .trade_component = .volume }, .expected = "ss(10, (bbs+aas)/(bs+as), v)" },
    };
    for (mnemonic_cases) |item| {
        var text_storage: [64]u8 = undefined;
        const formatted = try mnemonic(&text_storage, item.params);
        try testing.expectEqualStrings(item.expected, formatted);
        var instance = try SuperSmoother.init(item.params);
        instance.fixSlices();
        var item_meta: Metadata = undefined;
        instance.getMetadata(&item_meta);
        try testing.expectEqualStrings(item.expected, item_meta.mnemonic);
    }

    const closes = testdata.testInput();
    const opens = testdata.testOpen();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();

    var bar_filter = try SuperSmoother.init(defaultParams());
    bar_filter.fixSlices();
    var scalar_filter = try SuperSmoother.init(defaultParams());
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

    var quote_filter = try SuperSmoother.init(defaultParams());
    quote_filter.fixSlices();
    var quote_out: OutputArray = undefined;
    for (0..5) |_| {
        const quote = Quote{ .time = 2000, .bid_price = 99.0, .ask_price = 101.0, .bid_size = 1.0, .ask_size = 1.0 };
        quote_out = quote_filter.updateQuote(&quote);
    }
    try testing.expectEqual(@as(usize, 1), quote_out.len);
    try testing.expect(!math.isNan(quote_out.values[0].scalar.value));

    var trade_filter = try SuperSmoother.init(defaultParams());
    trade_filter.fixSlices();
    var trade_out: OutputArray = undefined;
    for (0..5) |_| {
        const trade = Trade{ .time = 3000, .price = 100.0, .volume = 50.0 };
        trade_out = trade_filter.updateTrade(&trade);
    }
    try testing.expectEqual(@as(usize, 1), trade_out.len);
    try testing.expect(!math.isNan(trade_out.values[0].scalar.value));
}
