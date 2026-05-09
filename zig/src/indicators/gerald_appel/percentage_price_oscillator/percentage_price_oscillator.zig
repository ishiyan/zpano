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
const sma_mod = @import("../../common/simple_moving_average/simple_moving_average.zig");
const ema_mod = @import("../../common/exponential_moving_average/exponential_moving_average.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the PPO indicator.
pub const PercentagePriceOscillatorOutput = enum(u8) {
    value = 1,
};

/// Specifies the type of moving average to use.
pub const MovingAverageType = enum(u8) {
    sma = 0,
    ema = 1,
};

/// Parameters to create an instance of the PPO indicator.
pub const PercentagePriceOscillatorParams = struct {
    fast_length: usize,
    slow_length: usize,
    moving_average_type: MovingAverageType = .sma,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

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

/// Percentage Price Oscillator (PPO) by Gerald Appel.
///
/// PPO = 100 * (fast_ma - slow_ma) / slow_ma
pub const PercentagePriceOscillator = struct {
    line: LineIndicator,
    fast_ma: MaUnion,
    slow_ma: MaUnion,
    value: f64,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: PercentagePriceOscillatorParams) !PercentagePriceOscillator {
        if (params.fast_length < 2) return error.InvalidFastLength;
        if (params.slow_length < 2) return error.InvalidSlowLength;

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
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "ppo({s}{d}/{s}{d}{s})", .{
            ma_label, params.fast_length, ma_label, params.slow_length, triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Percentage Price Oscillator {s}", .{mnemonic_slice}) catch
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

    pub fn deinit(self: *PercentagePriceOscillator) void {
        self.fast_ma.deinit();
        self.slow_ma.deinit();
    }

    pub fn fixSlices(self: *PercentagePriceOscillator) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *PercentagePriceOscillator, sample: f64) f64 {
        const epsilon = 1e-8;

        if (math.isNan(sample)) return sample;

        const slow = self.slow_ma.update(sample);
        const fast = self.fast_ma.update(sample);
        self.primed = self.slow_ma.isPrimed() and self.fast_ma.isPrimed();

        if (math.isNan(fast) or math.isNan(slow)) {
            self.value = math.nan(f64);
            return self.value;
        }

        if (@abs(slow) < epsilon) {
            self.value = 0;
        } else {
            self.value = 100.0 * (fast - slow) / slow;
        }

        return self.value;
    }

    pub fn isPrimed(self: *const PercentagePriceOscillator) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const PercentagePriceOscillator, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .percentage_price_oscillator,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *PercentagePriceOscillator, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *PercentagePriceOscillator, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *PercentagePriceOscillator, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *PercentagePriceOscillator, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *PercentagePriceOscillator) indicator_mod.Indicator {
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
        const self: *PercentagePriceOscillator = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const PercentagePriceOscillator = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *PercentagePriceOscillator = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *PercentagePriceOscillator = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *PercentagePriceOscillator = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *PercentagePriceOscillator = @ptrCast(@alignCast(ptr));
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


test "PPO SMA 2/3 spot checks" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const tolerance = 5e-4;

    var ppo = try PercentagePriceOscillator.init(allocator, .{
        .fast_length = 2,
        .slow_length = 3,
    });
    defer ppo.deinit();
    ppo.fixSlices();

    // First 2 values should be NaN.
    for (0..2) |i| {
        const v = ppo.update(input[i]);
        try testing.expect(math.isNan(v));
    }

    // Index 2: first value.
    const v2 = ppo.update(input[2]);
    try testing.expect(!math.isNan(v2));
    try testing.expect(@abs(v2 - 1.10264) < tolerance);

    // Index 3.
    const v3 = ppo.update(input[3]);
    try testing.expect(@abs(v3 - (-0.02813)) < tolerance);

    // Feed remaining.
    for (4..251) |i| {
        _ = ppo.update(input[i]);
    }
    const v251 = ppo.update(input[251]);
    try testing.expect(@abs(v251 - (-0.21191)) < tolerance);
    try testing.expect(ppo.isPrimed());
}

test "PPO SMA 12/26 spot checks" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const tolerance = 5e-4;

    var ppo = try PercentagePriceOscillator.init(allocator, .{
        .fast_length = 12,
        .slow_length = 26,
    });
    defer ppo.deinit();
    ppo.fixSlices();

    // First 25 values should be NaN.
    for (0..25) |i| {
        const v = ppo.update(input[i]);
        try testing.expect(math.isNan(v));
    }

    // Index 25: first value.
    const v25 = ppo.update(input[25]);
    try testing.expect(!math.isNan(v25));
    try testing.expect(@abs(v25 - (-3.6393)) < tolerance);

    // Index 26.
    const v26 = ppo.update(input[26]);
    try testing.expect(@abs(v26 - (-3.9534)) < tolerance);

    // Feed remaining.
    for (27..251) |i| {
        _ = ppo.update(input[i]);
    }
    const v251 = ppo.update(input[251]);
    try testing.expect(@abs(v251 - (-0.15281)) < tolerance);
}

test "PPO EMA 12/26 spot checks" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const tolerance = 5e-3;

    var ppo = try PercentagePriceOscillator.init(allocator, .{
        .fast_length = 12,
        .slow_length = 26,
        .moving_average_type = .ema,
        .first_is_average = false,
    });
    defer ppo.deinit();
    ppo.fixSlices();

    // First 25 values should be NaN.
    for (0..25) |i| {
        const v = ppo.update(input[i]);
        try testing.expect(math.isNan(v));
    }

    // Index 25.
    const v25 = ppo.update(input[25]);
    try testing.expect(!math.isNan(v25));
    try testing.expect(@abs(v25 - (-2.7083)) < tolerance);

    // Index 26.
    const v26 = ppo.update(input[26]);
    try testing.expect(@abs(v26 - (-2.7390)) < tolerance);

    // Feed remaining.
    for (27..251) |i| {
        _ = ppo.update(input[i]);
    }
    const v251 = ppo.update(input[251]);
    try testing.expect(@abs(v251 - 0.83644) < tolerance);
}

test "PPO isPrimed" {
    const allocator = testing.allocator;

    var ppo = try PercentagePriceOscillator.init(allocator, .{
        .fast_length = 3,
        .slow_length = 5,
    });
    defer ppo.deinit();
    ppo.fixSlices();

    try testing.expect(!ppo.isPrimed());

    for (1..5) |i| {
        _ = ppo.update(@floatFromInt(i));
        try testing.expect(!ppo.isPrimed());
    }

    _ = ppo.update(5);
    try testing.expect(ppo.isPrimed());

    for (6..10) |i| {
        _ = ppo.update(@floatFromInt(i));
        try testing.expect(ppo.isPrimed());
    }
}

test "PPO NaN passthrough" {
    const allocator = testing.allocator;

    var ppo = try PercentagePriceOscillator.init(allocator, .{
        .fast_length = 2,
        .slow_length = 3,
    });
    defer ppo.deinit();
    ppo.fixSlices();

    const v = ppo.update(math.nan(f64));
    try testing.expect(math.isNan(v));
}

test "PPO metadata SMA" {
    const allocator = testing.allocator;

    var ppo = try PercentagePriceOscillator.init(allocator, .{
        .fast_length = 12,
        .slow_length = 26,
    });
    defer ppo.deinit();
    ppo.fixSlices();

    var meta: Metadata = undefined;
    ppo.getMetadata(&meta);

    try testing.expectEqual(Identifier.percentage_price_oscillator, meta.identifier);
    try testing.expectEqualStrings("ppo(SMA12/SMA26)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "PPO metadata EMA" {
    const allocator = testing.allocator;

    var ppo = try PercentagePriceOscillator.init(allocator, .{
        .fast_length = 12,
        .slow_length = 26,
        .moving_average_type = .ema,
    });
    defer ppo.deinit();
    ppo.fixSlices();

    var meta: Metadata = undefined;
    ppo.getMetadata(&meta);

    try testing.expectEqualStrings("ppo(EMA12/EMA26)", meta.mnemonic);
}

test "PPO invalid params" {
    const allocator = testing.allocator;

    const r1 = PercentagePriceOscillator.init(allocator, .{ .fast_length = 1, .slow_length = 26 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = PercentagePriceOscillator.init(allocator, .{ .fast_length = 12, .slow_length = 1 });
    try testing.expect(if (r2) |_| false else |_| true);
}

test "PPO entity update" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const tolerance = 5e-4;

    var ppo = try PercentagePriceOscillator.init(allocator, .{
        .fast_length = 2,
        .slow_length = 3,
    });
    defer ppo.deinit();
    ppo.fixSlices();

    for (0..2) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        const out = ppo.updateScalar(&scalar);
        const v = out.slice()[0].scalar.value;
        try testing.expect(math.isNan(v));
    }

    const scalar = Scalar{ .time = 0, .value = input[2] };
    const out = ppo.updateScalar(&scalar);
    const v = out.slice()[0].scalar.value;
    try testing.expect(@abs(v - 1.10264) < tolerance);
}
