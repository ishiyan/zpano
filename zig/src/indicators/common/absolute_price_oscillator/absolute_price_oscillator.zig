const std = @import("std");
const math = std.math;

const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;
const bar_component = @import("bar_component");
const quote_component = @import("quote_component");
const trade_component = @import("trade_component");

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

fn testInput() [252]f64 {
    return .{
        91.500000,  94.815000,  94.375000,  95.095000,  93.780000,  94.625000,  92.530000,  92.750000,  90.315000,  92.470000,
        96.125000,  97.250000,  98.500000,  89.875000,  91.000000,  92.815000,  89.155000,  89.345000,  91.625000,  89.875000,
        88.375000,  87.625000,  84.780000,  83.000000,  83.500000,  81.375000,  84.440000,  89.250000,  86.375000,  86.250000,
        85.250000,  87.125000,  85.815000,  88.970000,  88.470000,  86.875000,  86.815000,  84.875000,  84.190000,  83.875000,
        83.375000,  85.500000,  89.190000,  89.440000,  91.095000,  90.750000,  91.440000,  89.000000,  91.000000,  90.500000,
        89.030000,  88.815000,  84.280000,  83.500000,  82.690000,  84.750000,  85.655000,  86.190000,  88.940000,  89.280000,
        88.625000,  88.500000,  91.970000,  91.500000,  93.250000,  93.500000,  93.155000,  91.720000,  90.000000,  89.690000,
        88.875000,  85.190000,  83.375000,  84.875000,  85.940000,  97.250000,  99.875000,  104.940000, 106.000000, 102.500000,
        102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000, 112.750000,
        123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000, 110.595000, 118.125000,
        116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000, 116.620000, 117.000000, 115.250000,
        114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000, 123.370000, 122.940000, 122.560000,
        123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000, 132.810000, 134.000000, 137.380000,
        137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000,
        123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000,
        122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000,
        124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
        132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000, 130.130000,
        127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000, 117.750000, 119.870000,
        122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000, 107.000000, 107.870000, 107.000000,
        107.120000, 107.000000, 91.000000,  93.940000,  93.870000,  95.500000,  93.000000,  94.940000,  98.250000,  96.750000,
        94.810000,  94.370000,  91.560000,  90.250000,  93.940000,  93.620000,  97.000000,  95.000000,  95.870000,  94.060000,
        94.620000,  93.750000,  98.000000,  103.940000, 107.870000, 106.060000, 104.500000, 105.000000, 104.190000, 103.060000,
        103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000, 109.700000, 109.250000,
        107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000, 109.810000, 109.000000,
        108.750000, 107.870000,
    };
}

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
    const input = testInput();
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
    const input = testInput();
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
    const input = testInput();
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
