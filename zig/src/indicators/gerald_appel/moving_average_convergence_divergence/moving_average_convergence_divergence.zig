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
const sma_mod = @import("../../common/simple_moving_average/simple_moving_average.zig");
const ema_mod = @import("../../common/exponential_moving_average/exponential_moving_average.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the MACD indicator.
pub const MacdOutput = enum(u8) {
    /// MACD line (fast MA - slow MA).
    macd = 1,
    /// Signal line (MA of MACD).
    signal = 2,
    /// Histogram (MACD - Signal).
    histogram = 3,
};

/// Specifies the type of moving average.
pub const MovingAverageType = enum(u8) {
    ema = 0,
    sma = 1,
};

/// Parameters to create a MACD indicator.
pub const MacdParams = struct {
    fast_length: usize = 12,
    slow_length: usize = 26,
    signal_length: usize = 9,
    moving_average_type: MovingAverageType = .ema,
    signal_moving_average_type: MovingAverageType = .ema,
    first_is_average: ?bool = null,
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

fn newMa(allocator: std.mem.Allocator, ma_type: MovingAverageType, length: usize, first_is_average: bool) !MaUnion {
    switch (ma_type) {
        .sma => {
            var sma = try sma_mod.SimpleMovingAverage.init(allocator, .{ .length = length });
            sma.fixSlices();
            return .{ .sma = sma };
        },
        .ema => {
            var ema = try ema_mod.ExponentialMovingAverage.initLength(.{
                .length = length,
                .first_is_average = first_is_average,
            });
            ema.fixSlices();
            return .{ .ema = ema };
        },
    }
}

fn maLabel(ma_type: MovingAverageType) []const u8 {
    return switch (ma_type) {
        .sma => "SMA",
        .ema => "EMA",
    };
}

/// Moving Average Convergence Divergence (MACD) by Gerald Appel.
///
/// MACD = fast MA - slow MA
/// Signal = MA of MACD
/// Histogram = MACD - Signal
pub const MovingAverageConvergenceDivergence = struct {
    fast_ma: MaUnion,
    slow_ma: MaUnion,
    signal_ma: MaUnion,

    macd_value: f64,
    signal_value: f64,
    histogram_value: f64,
    primed: bool,

    fast_delay: usize,
    fast_count: usize,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: MacdParams) !MovingAverageConvergenceDivergence {
        var fast_length = params.fast_length;
        var slow_length = params.slow_length;
        const signal_length = params.signal_length;

        if (fast_length < 2) return error.InvalidFastLength;
        if (slow_length < 2) return error.InvalidSlowLength;
        if (signal_length < 1) return error.InvalidSignalLength;

        // Auto-swap fast/slow if needed (matches TaLib behavior).
        if (slow_length < fast_length) {
            const tmp = fast_length;
            fast_length = slow_length;
            slow_length = tmp;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Default FirstIsAverage to true (TA-Lib compatible).
        const first_is_average = params.first_is_average orelse true;

        var fast_ma = try newMa(allocator, params.moving_average_type, fast_length, first_is_average);
        var slow_ma = try newMa(allocator, params.moving_average_type, slow_length, first_is_average);
        var signal_ma = try newMa(allocator, params.signal_moving_average_type, signal_length, first_is_average);

        // Fix slices after moving into struct fields below.
        _ = &fast_ma;
        _ = &slow_ma;
        _ = &signal_ma;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        // Build mnemonic: macd(12,26,9) or macd(12,26,9,SMA,EMA) if non-default types.
        var mnemonic_buf: [128]u8 = undefined;
        var mnemonic_slice: []u8 = undefined;

        if (params.moving_average_type != .ema or params.signal_moving_average_type != .ema) {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "macd({d},{d},{d},{s},{s}{s})", .{
                fast_length,
                slow_length,
                signal_length,
                maLabel(params.moving_average_type),
                maLabel(params.signal_moving_average_type),
                triple,
            }) catch return error.MnemonicTooLong;
        } else {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "macd({d},{d},{d}{s})", .{
                fast_length,
                slow_length,
                signal_length,
                triple,
            }) catch return error.MnemonicTooLong;
        }
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Moving Average Convergence Divergence {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .fast_ma = fast_ma,
            .slow_ma = slow_ma,
            .signal_ma = signal_ma,
            .macd_value = math.nan(f64),
            .signal_value = math.nan(f64),
            .histogram_value = math.nan(f64),
            .primed = false,
            .fast_delay = slow_length - fast_length,
            .fast_count = 0,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *MovingAverageConvergenceDivergence) void {
        self.fast_ma.deinit();
        self.slow_ma.deinit();
        self.signal_ma.deinit();
    }

    pub fn fixSlices(self: *MovingAverageConvergenceDivergence) void {
        _ = self;
        // MACD doesn't use LineIndicator, so no slice fixup needed for mnemonic/description.
        // The mnemonic/description are read from the buffers directly.
    }

    /// Returns macd, signal, histogram.
    pub fn updateValues(self: *MovingAverageConvergenceDivergence, sample: f64) struct { macd: f64, signal: f64, histogram: f64 } {
        const nan = math.nan(f64);

        if (math.isNan(sample)) {
            return .{ .macd = nan, .signal = nan, .histogram = nan };
        }

        // Feed slow MA every sample.
        const slow = self.slow_ma.update(sample);

        // Delay fast MA to align SMA seed windows.
        var fast: f64 = nan;
        if (self.fast_count < self.fast_delay) {
            self.fast_count += 1;
        } else {
            fast = self.fast_ma.update(sample);
        }

        if (math.isNan(fast) or math.isNan(slow)) {
            self.macd_value = nan;
            self.signal_value = nan;
            self.histogram_value = nan;
            return .{ .macd = nan, .signal = nan, .histogram = nan };
        }

        const macd = fast - slow;
        self.macd_value = macd;

        const sig = self.signal_ma.update(macd);

        if (math.isNan(sig)) {
            self.signal_value = nan;
            self.histogram_value = nan;
            return .{ .macd = macd, .signal = nan, .histogram = nan };
        }

        self.signal_value = sig;
        const hist = macd - sig;
        self.histogram_value = hist;
        self.primed = self.fast_ma.isPrimed() and self.slow_ma.isPrimed() and self.signal_ma.isPrimed();

        return .{ .macd = macd, .signal = sig, .histogram = hist };
    }

    pub fn isPrimed(self: *const MovingAverageConvergenceDivergence) bool {
        return self.primed;
    }

    fn mnemonic(self: *const MovingAverageConvergenceDivergence) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const MovingAverageConvergenceDivergence) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const MovingAverageConvergenceDivergence, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var macd_mn_buf: [160]u8 = undefined;
        const macd_mn = std.fmt.bufPrint(&macd_mn_buf, "{s} macd", .{mn}) catch mn;
        var signal_mn_buf: [160]u8 = undefined;
        const signal_mn = std.fmt.bufPrint(&signal_mn_buf, "{s} signal", .{mn}) catch mn;
        var hist_mn_buf: [160]u8 = undefined;
        const hist_mn = std.fmt.bufPrint(&hist_mn_buf, "{s} histogram", .{mn}) catch mn;

        var macd_desc_buf: [256]u8 = undefined;
        const macd_desc = std.fmt.bufPrint(&macd_desc_buf, "{s} MACD", .{desc}) catch desc;
        var signal_desc_buf: [256]u8 = undefined;
        const signal_desc = std.fmt.bufPrint(&signal_desc_buf, "{s} Signal", .{desc}) catch desc;
        var hist_desc_buf: [256]u8 = undefined;
        const hist_desc = std.fmt.bufPrint(&hist_desc_buf, "{s} Histogram", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .moving_average_convergence_divergence,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = macd_mn, .description = macd_desc },
                .{ .mnemonic = signal_mn, .description = signal_desc },
                .{ .mnemonic = hist_mn, .description = hist_desc },
            },
        );
    }

    pub fn updateScalar(self: *MovingAverageConvergenceDivergence, sample: *const Scalar) OutputArray {
        const result = self.updateValues(sample.value);
        return makeOutput(sample.time, result.macd, result.signal, result.histogram);
    }

    pub fn updateBar(self: *MovingAverageConvergenceDivergence, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *MovingAverageConvergenceDivergence, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *MovingAverageConvergenceDivergence, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    fn makeOutput(time: i64, macd_v: f64, signal_v: f64, hist_v: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = macd_v } });
        out.append(.{ .scalar = .{ .time = time, .value = signal_v } });
        out.append(.{ .scalar = .{ .time = time, .value = hist_v } });
        return out;
    }

    pub fn indicator(self: *MovingAverageConvergenceDivergence) indicator_mod.Indicator {
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
        const self: *MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidFastLength,
        InvalidSlowLength,
        InvalidSignalLength,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


test "MACD default params full validation" {
    const allocator = testing.allocator;
    const tolerance = 1e-8;

    const input = testdata.testInput();
    const exp_macd = testdata.testMacdExpected();
    const exp_signal = testdata.testSignalExpected();
    const exp_histogram = testdata.testHistogramExpected();

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    for (0..252) |i| {
        const result = ind.updateValues(input[i]);

        if (math.isNan(exp_macd[i])) {
            try testing.expect(math.isNan(result.macd));
            try testing.expect(math.isNan(result.signal));
            try testing.expect(math.isNan(result.histogram));
            continue;
        }

        if (!math.isNan(exp_macd[i])) {
            try testing.expect(@abs(result.macd - exp_macd[i]) <= tolerance);
        }

        if (math.isNan(exp_signal[i])) {
            try testing.expect(math.isNan(result.signal));
            try testing.expect(math.isNan(result.histogram));
            continue;
        }

        try testing.expect(@abs(result.signal - exp_signal[i]) <= tolerance);
        try testing.expect(@abs(result.histogram - exp_histogram[i]) <= tolerance);
    }
}

test "MACD TaLib spot check" {
    const allocator = testing.allocator;
    const tolerance = 5e-4;
    const input = testdata.testInput();

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    var result: @TypeOf(ind.updateValues(0)) = undefined;
    for (0..34) |i| {
        result = ind.updateValues(input[i]);
    }

    try testing.expect(@abs(result.macd - (-1.9738)) < tolerance);
    try testing.expect(@abs(result.signal - (-2.7071)) < tolerance);
    const exp_hist = (-1.9738) - (-2.7071);
    try testing.expect(@abs(result.histogram - exp_hist) < tolerance);
}

test "MACD period inversion" {
    const allocator = testing.allocator;
    const tolerance = 5e-4;
    const input = testdata.testInput();

    // fast=26, slow=12 should auto-swap.
    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{
        .fast_length = 26,
        .slow_length = 12,
    });
    defer ind.deinit();
    ind.fixSlices();

    var result: @TypeOf(ind.updateValues(0)) = undefined;
    for (0..34) |i| {
        result = ind.updateValues(input[i]);
    }

    try testing.expect(@abs(result.macd - (-1.9738)) < tolerance);
    try testing.expect(@abs(result.signal - (-2.7071)) < tolerance);
}

test "MACD isPrimed" {
    const allocator = testing.allocator;

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{
        .fast_length = 3,
        .slow_length = 5,
        .signal_length = 2,
    });
    defer ind.deinit();
    ind.fixSlices();

    try testing.expect(!ind.isPrimed());

    for (0..6) |i| {
        _ = ind.updateValues(@as(f64, @floatFromInt(i + 1)));
        if (i < 5) {
            try testing.expect(!ind.isPrimed());
        }
    }

    try testing.expect(ind.isPrimed());
}

test "MACD NaN passthrough" {
    const allocator = testing.allocator;

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    const result = ind.updateValues(math.nan(f64));
    try testing.expect(math.isNan(result.macd));
    try testing.expect(math.isNan(result.signal));
    try testing.expect(math.isNan(result.histogram));
}

test "MACD metadata default" {
    const allocator = testing.allocator;

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.moving_average_convergence_divergence, meta.identifier);
    try testing.expectEqualStrings("macd(12,26,9)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 3), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
    try testing.expectEqual(@as(u8, 3), meta.outputs_buf[2].kind);
}

test "MACD metadata SMA" {
    const allocator = testing.allocator;

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{
        .moving_average_type = .sma,
    });
    defer ind.deinit();
    ind.fixSlices();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("macd(12,26,9,SMA,EMA)", meta.mnemonic);
}

test "MACD invalid params" {
    const allocator = testing.allocator;

    const r1 = MovingAverageConvergenceDivergence.init(allocator, .{ .fast_length = 1 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = MovingAverageConvergenceDivergence.init(allocator, .{ .slow_length = 1 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = MovingAverageConvergenceDivergence.init(allocator, .{ .signal_length = 0 });
    try testing.expect(if (r3) |_| false else |_| true);
}

test "MACD entity update" {
    const allocator = testing.allocator;
    const tolerance = 5e-4;
    const input = testdata.testInput();

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    // First 24 should have NaN MACD.
    for (0..25) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        const out = ind.updateScalar(&scalar);
        const items = out.slice();
        const m = items[0].scalar.value;
        try testing.expect(math.isNan(m));
    }

    // Feed through index 33.
    for (25..33) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        _ = ind.updateScalar(&scalar);
    }

    const scalar = Scalar{ .time = 0, .value = input[33] };
    const out = ind.updateScalar(&scalar);
    const items = out.slice();
    const macd_v = items[0].scalar.value;
    const signal_v = items[1].scalar.value;
    const hist_v = items[2].scalar.value;

    try testing.expect(@abs(macd_v - (-1.9738)) < tolerance);
    try testing.expect(@abs(signal_v - (-2.7071)) < tolerance);
    const exp_hist = (-1.9738) - (-2.7071);
    try testing.expect(@abs(hist_v - exp_hist) < tolerance);
}
