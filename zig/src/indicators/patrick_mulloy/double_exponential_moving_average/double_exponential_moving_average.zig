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

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the double exponential moving average indicator.
pub const DoubleExponentialMovingAverageOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the DEMA based on length.
pub const DoubleExponentialMovingAverageLengthParams = struct {
    length: usize,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an instance of the DEMA based on smoothing factor.
pub const DoubleExponentialMovingAverageSmoothingFactorParams = struct {
    smoothing_factor: f64,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Double Exponential Moving Average (DEMA).
///
///   EMA¹ᵢ = EMA(Pᵢ)
///   EMA²ᵢ = EMA(EMA¹ᵢ)
///   DEMAᵢ = 2·EMA¹ᵢ - EMA²ᵢ
///
/// Warmup period: 2·length - 2 samples before primed.
pub const DoubleExponentialMovingAverage = struct {
    line: LineIndicator,
    smoothing_factor: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    length: usize,
    length2: usize,
    count: usize,
    first_is_average: bool,
    primed: bool,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    const epsilon: f64 = 0.00000001;

    /// Create DEMA from length.
    pub fn initLength(params: DoubleExponentialMovingAverageLengthParams) !DoubleExponentialMovingAverage {
        if (params.length < 1) {
            return error.InvalidLength;
        }
        const alpha = 2.0 / @as(f64, @floatFromInt(1 + params.length));
        return initInternal(params.length, alpha, false, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    /// Create DEMA from smoothing factor.
    pub fn initSmoothingFactor(params: DoubleExponentialMovingAverageSmoothingFactorParams) !DoubleExponentialMovingAverage {
        var alpha = params.smoothing_factor;
        if (alpha < 0.0 or alpha > 1.0) {
            return error.InvalidSmoothingFactor;
        }
        if (alpha < epsilon) {
            alpha = epsilon;
        }
        const length: usize = @intFromFloat(@round(2.0 / alpha) - 1.0);
        return initInternal(length, alpha, true, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    fn initInternal(
        length: usize,
        alpha: f64,
        is_alpha_mode: bool,
        first_is_average: bool,
        bc_opt: ?bar_component.BarComponent,
        qc_opt: ?quote_component.QuoteComponent,
        tc_opt: ?trade_component.TradeComponent,
    ) !DoubleExponentialMovingAverage {
        const bc = bc_opt orelse bar_component.default_bar_component;
        const qc = qc_opt orelse quote_component.default_quote_component;
        const tc = tc_opt orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        var mnemonic_slice: []u8 = undefined;
        if (is_alpha_mode) {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "dema({d}, {d:.8}{s})", .{ length, alpha, triple }) catch
                return error.MnemonicTooLong;
        } else {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "dema({d}{s})", .{ length, triple }) catch
                return error.MnemonicTooLong;
        }
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Double exponential moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                bc_opt,
                qc_opt,
                tc_opt,
            ),
            .smoothing_factor = alpha,
            .sum = 0.0,
            .ema1 = 0.0,
            .ema2 = 0.0,
            .length = length,
            .length2 = 2 * length - 1,
            .count = 0,
            .first_is_average = first_is_average,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *DoubleExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *DoubleExponentialMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            var v1 = self.ema1;
            var v2 = self.ema2;
            const sf = self.smoothing_factor;
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            return 2.0 * v1 - v2;
        }

        self.count += 1;
        if (self.first_is_average) {
            if (self.count == 1) {
                self.sum = sample;
            } else if (self.length >= self.count) {
                self.sum += sample;
                if (self.length == self.count) {
                    self.ema1 = self.sum / @as(f64, @floatFromInt(self.length));
                    self.sum = self.ema1;
                }
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.sum += self.ema1;

                if (self.length2 == self.count) {
                    self.primed = true;
                    self.ema2 = self.sum / @as(f64, @floatFromInt(self.length));
                    return 2.0 * self.ema1 - self.ema2;
                }
            }
        } else {
            // Metastock
            if (self.count == 1) {
                self.ema1 = sample;
            } else if (self.length >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                if (self.length == self.count) {
                    self.ema2 = self.ema1;
                }
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;

                if (self.length2 == self.count) {
                    self.primed = true;
                    return 2.0 * self.ema1 - self.ema2;
                }
            }
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const DoubleExponentialMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const DoubleExponentialMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .double_exponential_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *DoubleExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *DoubleExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *DoubleExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *DoubleExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *DoubleExponentialMovingAverage) indicator_mod.Indicator {
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
        const self: *DoubleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const DoubleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DoubleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DoubleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DoubleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DoubleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidLength,
        InvalidSmoothingFactor,
        MnemonicTooLong,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn createDemaLength(length: usize, first_is_average: bool) !DoubleExponentialMovingAverage {
    var dema = try DoubleExponentialMovingAverage.initLength(.{
        .length = length,
        .first_is_average = first_is_average,
    });
    dema.fixSlices();
    return dema;
}

fn createDemaAlpha(alpha: f64, first_is_average: bool) !DoubleExponentialMovingAverage {
    var dema = try DoubleExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = alpha,
        .first_is_average = first_is_average,
    });
    dema.fixSlices();
    return dema;
}

test "dema update length 2 firstIsAverage true" {
    const input = testdata.testInput();
    var dema = try createDemaLength(2, true);

    // length2 = 2*2-1 = 3, lprimed = 2*2-2 = 2
    // Indices 0,1: NaN
    try testing.expect(math.isNan(dema.update(input[0])));
    try testing.expect(math.isNan(dema.update(input[1])));

    // From index 2 onward, primed
    for (2..252) |i| {
        const act = dema.update(input[i]);
        switch (i) {
            4 => try testing.expect(@abs(94.013 - act) < 1e-2),
            5 => try testing.expect(@abs(94.539 - act) < 1e-2),
            251 => try testing.expect(@abs(107.94 - act) < 1e-2),
            else => {},
        }
    }

    try testing.expect(math.isNan(dema.update(math.nan(f64))));
}

test "dema update length 14 firstIsAverage true" {
    const input = testdata.testInput();
    var dema = try createDemaLength(14, true);
    const lprimed = 2 * 14 - 2;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(dema.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = dema.update(input[i]);
        switch (i) {
            28 => try testing.expect(@abs(84.347 - act) < 1e-2),
            29 => try testing.expect(@abs(84.487 - act) < 1e-2),
            30 => try testing.expect(@abs(84.374 - act) < 1e-2),
            31 => try testing.expect(@abs(84.772 - act) < 1e-2),
            48 => try testing.expect(@abs(89.803 - act) < 1e-2),
            251 => try testing.expect(@abs(109.4676 - act) < 1e-2),
            else => {},
        }
    }

    try testing.expect(math.isNan(dema.update(math.nan(f64))));
}

test "dema update length 2 firstIsAverage false (Metastock)" {
    const input = testdata.testInput();
    var dema = try createDemaLength(2, false);
    const lprimed = 2 * 2 - 2;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(dema.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = dema.update(input[i]);
        switch (i) {
            4 => try testing.expect(@abs(93.977 - act) < 1e-2),
            5 => try testing.expect(@abs(94.522 - act) < 1e-2),
            251 => try testing.expect(@abs(107.94 - act) < 1e-2),
            else => {},
        }
    }

    try testing.expect(math.isNan(dema.update(math.nan(f64))));
}

test "dema update length 14 firstIsAverage false (Metastock)" {
    const input = testdata.testInput();
    var dema = try createDemaLength(14, false);
    const lprimed = 2 * 14 - 2;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(dema.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = dema.update(input[i]);
        switch (i) {
            28 => try testing.expect(@abs(84.87 - act) < 1e-2),
            29 => try testing.expect(@abs(84.94 - act) < 1e-2),
            30 => try testing.expect(@abs(84.77 - act) < 1e-2),
            31 => try testing.expect(@abs(85.12 - act) < 1e-2),
            48 => try testing.expect(@abs(89.83 - act) < 1e-2),
            251 => try testing.expect(@abs(109.4676 - act) < 1e-2),
            else => {},
        }
    }

    try testing.expect(math.isNan(dema.update(math.nan(f64))));
}

test "dema isPrimed length 14" {
    const input = testdata.testInput();
    const l = 14;
    const lprimed = 2 * l - 2;

    // firstIsAverage = true
    {
        var dema = try createDemaLength(l, true);
        try testing.expect(!dema.isPrimed());
        for (0..lprimed) |i| {
            _ = dema.update(input[i]);
            try testing.expect(!dema.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = dema.update(input[i]);
            try testing.expect(dema.isPrimed());
        }
    }

    // firstIsAverage = false
    {
        var dema = try createDemaLength(l, false);
        try testing.expect(!dema.isPrimed());
        for (0..lprimed) |i| {
            _ = dema.update(input[i]);
            try testing.expect(!dema.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = dema.update(input[i]);
            try testing.expect(dema.isPrimed());
        }
    }
}

test "dema metadata length" {
    var dema = try createDemaLength(10, true);
    var m: Metadata = undefined;
    dema.getMetadata(&m);

    try testing.expectEqual(Identifier.double_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("dema(10)", m.mnemonic);
    try testing.expectEqualStrings("Double exponential moving average dema(10)", m.description);
}

test "dema metadata alpha" {
    const alpha: f64 = 2.0 / 11.0;
    var dema = try createDemaAlpha(alpha, false);
    var m: Metadata = undefined;
    dema.getMetadata(&m);

    try testing.expectEqual(Identifier.double_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("dema(10, 0.18181818)", m.mnemonic);
    try testing.expectEqualStrings("Double exponential moving average dema(10, 0.18181818)", m.description);
}

test "dema update entity" {
    const inp: f64 = 3.0;
    const exp_false: f64 = 2.666666666666667;
    const time: i64 = 1617235200;

    // scalar
    {
        var dema = try createDemaLength(2, false);
        _ = dema.update(0.0);
        _ = dema.update(0.0);
        const out = dema.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(@abs(exp_false - s.value) < 1e-13);
    }

    // bar
    {
        var dema = try createDemaLength(2, false);
        _ = dema.update(0.0);
        _ = dema.update(0.0);
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = dema.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(exp_false - s.value) < 1e-13);
    }
}

test "dema init invalid" {
    // length = 0
    try testing.expectError(error.InvalidLength, DoubleExponentialMovingAverage.initLength(.{ .length = 0 }));

    // alpha < 0
    try testing.expectError(error.InvalidSmoothingFactor, DoubleExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = -1.0 }));

    // alpha > 1
    try testing.expectError(error.InvalidSmoothingFactor, DoubleExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = 2.0 }));
}

test "dema mnemonic with bar component" {
    var dema = try DoubleExponentialMovingAverage.initLength(.{
        .length = 10,
        .bar_component = .median,
    });
    dema.fixSlices();
    try testing.expectEqualStrings("dema(10, hl/2)", dema.line.mnemonic);
    try testing.expectEqualStrings("Double exponential moving average dema(10, hl/2)", dema.line.description);
}

test "dema alpha with quote component" {
    var dema = try DoubleExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = 2.0 / 11.0,
        .quote_component = .bid,
    });
    dema.fixSlices();
    try testing.expectEqualStrings("dema(10, 0.18181818, b)", dema.line.mnemonic);
}
