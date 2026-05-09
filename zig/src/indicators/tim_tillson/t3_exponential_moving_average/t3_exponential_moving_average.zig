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

/// Enumerates the outputs of the T3 exponential moving average indicator.
pub const T3ExponentialMovingAverageOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the T3 EMA based on length.
pub const T3ExponentialMovingAverageLengthParams = struct {
    length: usize,
    volume_factor: f64 = 0.7,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an instance of the T3 EMA based on smoothing factor.
pub const T3ExponentialMovingAverageSmoothingFactorParams = struct {
    smoothing_factor: f64,
    volume_factor: f64 = 0.7,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the T3 Exponential Moving Average.
///
/// T3 is a six-pole non-linear Kalman filter developed by Tim Tillson.
///
///   c1*EMA6 + c2*EMA5 + c3*EMA4 + c4*EMA3
///
/// Warmup period: 6*length - 6 samples before primed.
pub const T3ExponentialMovingAverage = struct {
    line: LineIndicator,
    smoothing_factor: f64,
    c1: f64,
    c2: f64,
    c3: f64,
    c4: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    ema3: f64,
    ema4: f64,
    ema5: f64,
    ema6: f64,
    length: usize,
    length2: usize,
    length3: usize,
    length4: usize,
    length5: usize,
    length6: usize,
    count: usize,
    first_is_average: bool,
    primed: bool,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    const epsilon: f64 = 0.00000001;

    /// Create T3 EMA from length.
    pub fn initLength(params: T3ExponentialMovingAverageLengthParams) !T3ExponentialMovingAverage {
        if (params.length < 2) {
            return error.InvalidLength;
        }
        if (params.volume_factor < 0.0 or params.volume_factor > 1.0) {
            return error.InvalidVolumeFactor;
        }
        const alpha = 2.0 / @as(f64, @floatFromInt(1 + params.length));
        return initInternal(params.length, alpha, false, params.volume_factor, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    /// Create T3 EMA from smoothing factor.
    pub fn initSmoothingFactor(params: T3ExponentialMovingAverageSmoothingFactorParams) !T3ExponentialMovingAverage {
        if (params.volume_factor < 0.0 or params.volume_factor > 1.0) {
            return error.InvalidVolumeFactor;
        }
        var alpha = params.smoothing_factor;
        if (alpha < 0.0 or alpha > 1.0) {
            return error.InvalidSmoothingFactor;
        }
        if (alpha < epsilon) {
            alpha = epsilon;
        }
        const length: usize = @intFromFloat(@round(2.0 / alpha) - 1.0);
        return initInternal(length, alpha, true, params.volume_factor, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    fn initInternal(
        length: usize,
        alpha: f64,
        is_alpha_mode: bool,
        v: f64,
        first_is_average: bool,
        bc_opt: ?bar_component.BarComponent,
        qc_opt: ?quote_component.QuoteComponent,
        tc_opt: ?trade_component.TradeComponent,
    ) !T3ExponentialMovingAverage {
        const bc = bc_opt orelse bar_component.default_bar_component;
        const qc = qc_opt orelse quote_component.default_quote_component;
        const tc = tc_opt orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        var mnemonic_slice: []u8 = undefined;
        if (is_alpha_mode) {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "t3({d}, {d:.8}, {d:.8}{s})", .{ length, alpha, v, triple }) catch
                return error.MnemonicTooLong;
        } else {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "t3({d}, {d:.8}{s})", .{ length, v, triple }) catch
                return error.MnemonicTooLong;
        }
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "T3 exponential moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const vv = v * v;
        const c1 = -vv * v;
        const c2 = 3.0 * (vv - c1);
        const c3 = -6.0 * vv - 3.0 * (v - c1);
        const c4 = 1.0 + 3.0 * v - c1 + 3.0 * vv;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                bc_opt,
                qc_opt,
                tc_opt,
            ),
            .smoothing_factor = alpha,
            .c1 = c1,
            .c2 = c2,
            .c3 = c3,
            .c4 = c4,
            .sum = 0.0,
            .ema1 = 0.0,
            .ema2 = 0.0,
            .ema3 = 0.0,
            .ema4 = 0.0,
            .ema5 = 0.0,
            .ema6 = 0.0,
            .length = length,
            .length2 = 2 * length - 1,
            .length3 = 3 * length - 2,
            .length4 = 4 * length - 3,
            .length5 = 5 * length - 4,
            .length6 = 6 * length - 5,
            .count = 0,
            .first_is_average = first_is_average,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *T3ExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *T3ExponentialMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            var v1 = self.ema1;
            var v2 = self.ema2;
            var v3 = self.ema3;
            var v4 = self.ema4;
            var v5 = self.ema5;
            var v6 = self.ema6;
            const sf = self.smoothing_factor;
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            v3 += (v2 - v3) * sf;
            v4 += (v3 - v4) * sf;
            v5 += (v4 - v5) * sf;
            v6 += (v5 - v6) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            self.ema3 = v3;
            self.ema4 = v4;
            self.ema5 = v5;
            self.ema6 = v6;
            return self.c1 * v6 + self.c2 * v5 + self.c3 * v4 + self.c4 * v3;
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
            } else if (self.length2 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.sum += self.ema1;

                if (self.length2 == self.count) {
                    self.ema2 = self.sum / @as(f64, @floatFromInt(self.length));
                    self.sum = self.ema2;
                }
            } else if (self.length3 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.sum += self.ema2;

                if (self.length3 == self.count) {
                    self.ema3 = self.sum / @as(f64, @floatFromInt(self.length));
                    self.sum = self.ema3;
                }
            } else if (self.length4 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.sum += self.ema3;

                if (self.length4 == self.count) {
                    self.ema4 = self.sum / @as(f64, @floatFromInt(self.length));
                    self.sum = self.ema4;
                }
            } else if (self.length5 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;
                self.sum += self.ema4;

                if (self.length5 == self.count) {
                    self.ema5 = self.sum / @as(f64, @floatFromInt(self.length));
                    self.sum = self.ema5;
                }
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;
                self.ema5 += (self.ema4 - self.ema5) * self.smoothing_factor;
                self.sum += self.ema5;

                if (self.length6 == self.count) {
                    self.primed = true;
                    self.ema6 = self.sum / @as(f64, @floatFromInt(self.length));
                    return self.c1 * self.ema6 + self.c2 * self.ema5 + self.c3 * self.ema4 + self.c4 * self.ema3;
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
            } else if (self.length2 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;

                if (self.length2 == self.count) {
                    self.ema3 = self.ema2;
                }
            } else if (self.length3 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;

                if (self.length3 == self.count) {
                    self.ema4 = self.ema3;
                }
            } else if (self.length4 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;

                if (self.length4 == self.count) {
                    self.ema5 = self.ema4;
                }
            } else if (self.length5 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;
                self.ema5 += (self.ema4 - self.ema5) * self.smoothing_factor;

                if (self.length5 == self.count) {
                    self.ema6 = self.ema5;
                }
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;
                self.ema5 += (self.ema4 - self.ema5) * self.smoothing_factor;
                self.ema6 += (self.ema5 - self.ema6) * self.smoothing_factor;

                if (self.length6 == self.count) {
                    self.primed = true;
                    return self.c1 * self.ema6 + self.c2 * self.ema5 + self.c3 * self.ema4 + self.c4 * self.ema3;
                }
            }
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const T3ExponentialMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const T3ExponentialMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .t3_exponential_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *T3ExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *T3ExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *T3ExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *T3ExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *T3ExponentialMovingAverage) indicator_mod.Indicator {
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
        const self: *T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidLength,
        InvalidSmoothingFactor,
        InvalidVolumeFactor,
        MnemonicTooLong,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


// Expected data from test_T3.xls, T3(5, 0.7) — firstIsAverage = true.
fn createT3Length(length: usize, first_is_average: bool, volume: f64) !T3ExponentialMovingAverage {
    var t3 = try T3ExponentialMovingAverage.initLength(.{
        .length = length,
        .volume_factor = volume,
        .first_is_average = first_is_average,
    });
    t3.fixSlices();
    return t3;
}

fn createT3Alpha(alpha: f64, first_is_average: bool, volume: f64) !T3ExponentialMovingAverage {
    var t3 = try T3ExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = alpha,
        .volume_factor = volume,
        .first_is_average = first_is_average,
    });
    t3.fixSlices();
    return t3;
}

test "t3 update length 5 firstIsAverage true (t3.xls)" {
    const input = testdata.testInput();
    const exp = testdata.testExpected();
    var t3 = try createT3Length(5, true, 0.7);
    const lprimed = 6 * 5 - 6;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(t3.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = t3.update(input[i]);
        try testing.expect(@abs(exp[i] - act) < 1e-3);
    }

    try testing.expect(math.isNan(t3.update(math.nan(f64))));
}

test "t3 update length 5 firstIsAverage false (Metastock)" {
    const input = testdata.testInput();
    var t3 = try createT3Length(5, false, 0.7);
    const lprimed = 6 * 5 - 6;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(t3.update(input[i])));
    }

    // Spot-check values from TA-Lib tests.
    for (lprimed..252) |i| {
        const act = t3.update(input[i]);
        switch (i) {
            24 => try testing.expect(@abs(85.749 - act) < 1e-3),
            25 => try testing.expect(@abs(84.380 - act) < 1e-3),
            250 => try testing.expect(@abs(109.032 - act) < 1e-3),
            251 => try testing.expect(@abs(108.88 - act) < 1e-3),
            else => {},
        }
    }

    try testing.expect(math.isNan(t3.update(math.nan(f64))));
}

test "t3 isPrimed length 5" {
    const input = testdata.testInput();
    const l = 5;
    const lprimed = 6 * l - 6;

    // firstIsAverage = true
    {
        var t3 = try createT3Length(l, true, 0.7);
        try testing.expect(!t3.isPrimed());
        for (0..lprimed) |i| {
            _ = t3.update(input[i]);
            try testing.expect(!t3.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = t3.update(input[i]);
            try testing.expect(t3.isPrimed());
        }
    }

    // firstIsAverage = false
    {
        var t3 = try createT3Length(l, false, 0.7);
        try testing.expect(!t3.isPrimed());
        for (0..lprimed) |i| {
            _ = t3.update(input[i]);
            try testing.expect(!t3.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = t3.update(input[i]);
            try testing.expect(t3.isPrimed());
        }
    }
}

test "t3 metadata length" {
    var t3 = try createT3Length(10, true, 0.3333);
    var m: Metadata = undefined;
    t3.getMetadata(&m);

    try testing.expectEqual(Identifier.t3_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("t3(10, 0.33330000)", m.mnemonic);
    try testing.expectEqualStrings("T3 exponential moving average t3(10, 0.33330000)", m.description);
}

test "t3 metadata alpha" {
    const alpha: f64 = 2.0 / 11.0;
    var t3 = try createT3Alpha(alpha, false, 0.3333333);
    var m: Metadata = undefined;
    t3.getMetadata(&m);

    try testing.expectEqual(Identifier.t3_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("t3(10, 0.18181818, 0.33333330)", m.mnemonic);
    try testing.expectEqualStrings("T3 exponential moving average t3(10, 0.18181818, 0.33333330)", m.description);
}

test "t3 update entity" {
    const inp: f64 = 3.0;
    const exp_false: f64 = 1.6675884773662544;
    const exp_true: f64 = 1.6901728395061721;
    const time: i64 = 1617235200;
    const l = 2;
    const lprimed = 6 * l - 6;

    // scalar
    {
        var t3 = try createT3Length(l, false, 0.7);
        for (0..lprimed) |_| {
            _ = t3.update(0.0);
        }
        const out = t3.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(@abs(exp_false - s.value) < 1e-13);
    }

    // bar
    {
        var t3 = try createT3Length(l, true, 0.7);
        for (0..lprimed) |_| {
            _ = t3.update(0.0);
        }
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = t3.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(exp_true - s.value) < 1e-13);
    }
}

test "t3 init invalid" {
    // length < 2
    try testing.expectError(error.InvalidLength, T3ExponentialMovingAverage.initLength(.{ .length = 1 }));
    try testing.expectError(error.InvalidLength, T3ExponentialMovingAverage.initLength(.{ .length = 0 }));

    // alpha out of range
    try testing.expectError(error.InvalidSmoothingFactor, T3ExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = -1.0 }));
    try testing.expectError(error.InvalidSmoothingFactor, T3ExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = 2.0 }));

    // volume factor out of range
    try testing.expectError(error.InvalidVolumeFactor, T3ExponentialMovingAverage.initLength(.{ .length = 5, .volume_factor = -0.7 }));
    try testing.expectError(error.InvalidVolumeFactor, T3ExponentialMovingAverage.initLength(.{ .length = 5, .volume_factor = 1.7 }));
}

test "t3 mnemonic with bar component" {
    var t3 = try T3ExponentialMovingAverage.initLength(.{
        .length = 10,
        .volume_factor = 0.7,
        .bar_component = .median,
    });
    t3.fixSlices();
    try testing.expectEqualStrings("t3(10, 0.70000000, hl/2)", t3.line.mnemonic);
    try testing.expectEqualStrings("T3 exponential moving average t3(10, 0.70000000, hl/2)", t3.line.description);
}

test "t3 alpha with quote component" {
    var t3 = try T3ExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = 2.0 / 11.0,
        .volume_factor = 0.7,
        .quote_component = .bid,
    });
    t3.fixSlices();
    try testing.expectEqualStrings("t3(10, 0.18181818, 0.70000000, b)", t3.line.mnemonic);
}
