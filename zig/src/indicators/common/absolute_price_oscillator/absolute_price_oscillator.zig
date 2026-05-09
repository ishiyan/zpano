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
const sma_mod = @import("../simple_moving_average/simple_moving_average.zig");
const ema_mod = @import("../exponential_moving_average/exponential_moving_average.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the APO indicator.
pub const AbsolutePriceOscillatorOutput = enum(u8) {
    value = 1,
};

/// Specifies the type of moving average to use.
pub const MovingAverageType = enum(u8) {
    sma = 0,
    ema = 1,
};

/// Parameters to create an instance of the APO indicator.
pub const AbsolutePriceOscillatorParams = struct {
    fast_length: usize,
    slow_length: usize,
    moving_average_type: MovingAverageType = .sma,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Absolute Price Oscillator (APO).
///
/// APO = fast_ma - slow_ma
pub const AbsolutePriceOscillator = struct {
    line: LineIndicator,
    // We use a tagged union for the MA type to avoid dynamic dispatch overhead.
    fast_ma: MaUnion,
    slow_ma: MaUnion,
    value: f64,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    const MaUnion = union(enum) {
        sma: sma_mod.SimpleMovingAverage,
        ema: ema_mod.ExponentialMovingAverage,

        fn update(self: *MaUnion, sample: f64) f64 {
            return switch (self.*) {
                .sma => |*s| s.update(sample),
                .ema => |*e| e.update(sample),
            };
        }

        fn isPrimed(self: *const MaUnion) bool {
            return switch (self.*) {
                .sma => |*s| s.isPrimed(),
                .ema => |*e| e.isPrimed(),
            };
        }

        fn deinit(self: *MaUnion) void {
            switch (self.*) {
                .sma => |*s| s.deinit(),
                .ema => {},
            }
        }
    };

    pub fn init(allocator: std.mem.Allocator, params: AbsolutePriceOscillatorParams) !AbsolutePriceOscillator {
        if (params.fast_length < 2) {
            return error.InvalidFastLength;
        }
        if (params.slow_length < 2) {
            return error.InvalidSlowLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var fast_ma: MaUnion = undefined;
        var slow_ma: MaUnion = undefined;
        var ma_label: []const u8 = undefined;

        switch (params.moving_average_type) {
            .ema => {
                ma_label = "EMA";
                var fast_ema = try ema_mod.ExponentialMovingAverage.initLength(.{
                    .length = params.fast_length,
                    .first_is_average = params.first_is_average,
                });
                fast_ema.fixSlices();
                var slow_ema = try ema_mod.ExponentialMovingAverage.initLength(.{
                    .length = params.slow_length,
                    .first_is_average = params.first_is_average,
                });
                slow_ema.fixSlices();
                fast_ma = .{ .ema = fast_ema };
                slow_ma = .{ .ema = slow_ema };
            },
            .sma => {
                ma_label = "SMA";
                var fast_sma = try sma_mod.SimpleMovingAverage.init(allocator, .{
                    .length = params.fast_length,
                });
                fast_sma.fixSlices();
                var slow_sma = try sma_mod.SimpleMovingAverage.init(allocator, .{
                    .length = params.slow_length,
                });
                slow_sma.fixSlices();
                fast_ma = .{ .sma = fast_sma };
                slow_ma = .{ .sma = slow_sma };
            },
        }

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "apo({s}{d}/{s}{d}{s})", .{
            ma_label, params.fast_length, ma_label, params.slow_length, triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Absolute Price Oscillator {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .fast_ma = fast_ma,
            .slow_ma = slow_ma,
            .value = math.nan(f64),
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *AbsolutePriceOscillator) void {
        self.fast_ma.deinit();
        self.slow_ma.deinit();
    }

    pub fn fixSlices(self: *AbsolutePriceOscillator) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *AbsolutePriceOscillator, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const slow = self.slow_ma.update(sample);
        const fast = self.fast_ma.update(sample);
        self.primed = self.slow_ma.isPrimed() and self.fast_ma.isPrimed();

        if (math.isNan(fast) or math.isNan(slow)) {
            self.value = math.nan(f64);
            return self.value;
        }

        self.value = fast - slow;
        return self.value;
    }

    pub fn isPrimed(self: *const AbsolutePriceOscillator) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const AbsolutePriceOscillator, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .absolute_price_oscillator,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *AbsolutePriceOscillator, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *AbsolutePriceOscillator, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *AbsolutePriceOscillator, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *AbsolutePriceOscillator, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *AbsolutePriceOscillator) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.VTable{
        .isPrimed = vtableIsPrimed,
        .metadata = vtableMetadata,
        .updateScalar = vtableUpdateScalar,
        .updateBar = vtableUpdateBar,
        .updateQuote = vtableUpdateQuote,
        .updateTrade = vtableUpdateTrade,
    };

    fn vtableIsPrimed(ptr: *anyopaque) bool {
        const self: *AbsolutePriceOscillator = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const AbsolutePriceOscillator = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AbsolutePriceOscillator = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AbsolutePriceOscillator = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AbsolutePriceOscillator = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AbsolutePriceOscillator = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidFastLength,
        InvalidSlowLength,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn createApo(allocator: std.mem.Allocator, fast: usize, slow: usize, ma_type: MovingAverageType, first_is_avg: bool) !AbsolutePriceOscillator {
    var apo = try AbsolutePriceOscillator.init(allocator, .{
        .fast_length = fast,
        .slow_length = slow,
        .moving_average_type = ma_type,
        .first_is_average = first_is_avg,
    });
    apo.fixSlices();
    return apo;
}

test "apo sma 12/26" {
    const tolerance = 5e-4;
    const input = testdata.testInput();
    var apo = try createApo(testing.allocator, 12, 26, .sma, false);
    defer apo.deinit();

    for (0..25) |_i| {
        const v = apo.update(input[_i]);
        try testing.expect(math.isNan(v));
    }

    // Index 25: first value
    var v = apo.update(input[25]);
    try testing.expect(!math.isNan(v));
    try testing.expect(@abs(v - (-3.3124)) < tolerance);

    // Index 26
    v = apo.update(input[26]);
    try testing.expect(@abs(v - (-3.5876)) < tolerance);

    // Feed remaining
    for (27..251) |i| {
        _ = apo.update(input[i]);
    }

    v = apo.update(input[251]);
    try testing.expect(@abs(v - (-0.1667)) < tolerance);
    try testing.expect(apo.isPrimed());
}

test "apo ema 12/26" {
    const tolerance = 5e-4;
    const input = testdata.testInput();
    var apo = try createApo(testing.allocator, 12, 26, .ema, false);
    defer apo.deinit();

    for (0..25) |_i| {
        const v = apo.update(input[_i]);
        try testing.expect(math.isNan(v));
    }

    var v = apo.update(input[25]);
    try testing.expect(!math.isNan(v));
    try testing.expect(@abs(v - (-2.4193)) < tolerance);

    v = apo.update(input[26]);
    try testing.expect(@abs(v - (-2.4367)) < tolerance);

    for (27..251) |i| {
        _ = apo.update(input[i]);
    }

    v = apo.update(input[251]);
    try testing.expect(@abs(v - 0.90401) < tolerance);
}

test "apo is primed" {
    var apo = try createApo(testing.allocator, 3, 5, .sma, false);
    defer apo.deinit();

    try testing.expect(!apo.isPrimed());

    for (1..5) |i| {
        _ = apo.update(@floatFromInt(i));
        try testing.expect(!apo.isPrimed());
    }

    _ = apo.update(5.0);
    try testing.expect(apo.isPrimed());
}

test "apo nan passthrough" {
    var apo = try createApo(testing.allocator, 2, 3, .sma, false);
    defer apo.deinit();
    try testing.expect(math.isNan(apo.update(math.nan(f64))));
}

test "apo metadata sma" {
    var apo = try createApo(testing.allocator, 12, 26, .sma, false);
    defer apo.deinit();
    var m: Metadata = undefined;
    apo.getMetadata(&m);

    try testing.expectEqual(Identifier.absolute_price_oscillator, m.identifier);
    try testing.expectEqualStrings("apo(SMA12/SMA26)", m.outputs_buf[0].mnemonic);
}

test "apo metadata ema" {
    var apo = try createApo(testing.allocator, 12, 26, .ema, false);
    defer apo.deinit();
    var m: Metadata = undefined;
    apo.getMetadata(&m);
    try testing.expectEqualStrings("apo(EMA12/EMA26)", m.outputs_buf[0].mnemonic);
}

test "apo update entity" {
    const input = testdata.testInput();
    var apo = try createApo(testing.allocator, 2, 3, .sma, false);
    defer apo.deinit();

    const time: i64 = 1617235200;

    for (0..2) |i| {
        const out = apo.updateScalar(&.{ .time = time, .value = input[i] });
        const v = out.slice()[0].scalar.value;
        try testing.expect(math.isNan(v));
    }

    const out = apo.updateScalar(&.{ .time = time, .value = input[2] });
    const v = out.slice()[0].scalar.value;
    try testing.expect(!math.isNan(v));
}

test "apo invalid params" {
    // fast too small
    try testing.expectError(error.InvalidFastLength, AbsolutePriceOscillator.init(testing.allocator, .{
        .fast_length = 1,
        .slow_length = 26,
    }));
    // slow too small
    try testing.expectError(error.InvalidSlowLength, AbsolutePriceOscillator.init(testing.allocator, .{
        .fast_length = 12,
        .slow_length = 1,
    }));
}
