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

/// Enumerates the outputs of the T2 exponential moving average indicator.
pub const T2ExponentialMovingAverageOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the T2 EMA based on length.
pub const T2ExponentialMovingAverageLengthParams = struct {
    length: usize,
    volume_factor: f64 = 0.7,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an instance of the T2 EMA based on smoothing factor.
pub const T2ExponentialMovingAverageSmoothingFactorParams = struct {
    smoothing_factor: f64,
    volume_factor: f64 = 0.7,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the T2 Exponential Moving Average.
///
/// T2 is a four-pole non-linear Kalman filter developed by Tim Tillson.
///
///   v^2 * EMA4 - 2v(1+v) * EMA3 + (1+v)^2 * EMA2
///
/// Warmup period: 4*length - 4 samples before primed.
pub const T2ExponentialMovingAverage = struct {
    line: LineIndicator,
    smoothing_factor: f64,
    c1: f64,
    c2: f64,
    c3: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    ema3: f64,
    ema4: f64,
    length: usize,
    length2: usize,
    length3: usize,
    length4: usize,
    count: usize,
    first_is_average: bool,
    primed: bool,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    const epsilon: f64 = 0.00000001;

    /// Create T2 EMA from length.
    pub fn initLength(params: T2ExponentialMovingAverageLengthParams) !T2ExponentialMovingAverage {
        if (params.length < 2) {
            return error.InvalidLength;
        }
        if (params.volume_factor < 0.0 or params.volume_factor > 1.0) {
            return error.InvalidVolumeFactor;
        }
        const alpha = 2.0 / @as(f64, @floatFromInt(1 + params.length));
        return initInternal(params.length, alpha, false, params.volume_factor, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    /// Create T2 EMA from smoothing factor.
    pub fn initSmoothingFactor(params: T2ExponentialMovingAverageSmoothingFactorParams) !T2ExponentialMovingAverage {
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
    ) !T2ExponentialMovingAverage {
        const bc = bc_opt orelse bar_component.default_bar_component;
        const qc = qc_opt orelse quote_component.default_quote_component;
        const tc = tc_opt orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        var mnemonic_slice: []u8 = undefined;
        if (is_alpha_mode) {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "t2({d}, {d:.8}, {d:.8}{s})", .{ length, alpha, v, triple }) catch
                return error.MnemonicTooLong;
        } else {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "t2({d}, {d:.8}{s})", .{ length, v, triple }) catch
                return error.MnemonicTooLong;
        }
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "T2 exponential moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const v1 = v + 1.0;
        const c1 = v * v;
        const c2 = -2.0 * v * v1;
        const c3 = v1 * v1;

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
            .sum = 0.0,
            .ema1 = 0.0,
            .ema2 = 0.0,
            .ema3 = 0.0,
            .ema4 = 0.0,
            .length = length,
            .length2 = 2 * length - 1,
            .length3 = 3 * length - 2,
            .length4 = 4 * length - 3,
            .count = 0,
            .first_is_average = first_is_average,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *T2ExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *T2ExponentialMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            var v1 = self.ema1;
            var v2 = self.ema2;
            var v3 = self.ema3;
            var v4 = self.ema4;
            const sf = self.smoothing_factor;
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            v3 += (v2 - v3) * sf;
            v4 += (v3 - v4) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            self.ema3 = v3;
            self.ema4 = v4;
            return self.c1 * v4 + self.c2 * v3 + self.c3 * v2;
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
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.sum += self.ema3;

                if (self.length4 == self.count) {
                    self.primed = true;
                    self.ema4 = self.sum / @as(f64, @floatFromInt(self.length));
                    return self.c1 * self.ema4 + self.c2 * self.ema3 + self.c3 * self.ema2;
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
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;

                if (self.length4 == self.count) {
                    self.primed = true;
                    return self.c1 * self.ema4 + self.c2 * self.ema3 + self.c3 * self.ema2;
                }
            }
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const T2ExponentialMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const T2ExponentialMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .t2_exponential_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *T2ExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *T2ExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *T2ExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *T2ExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *T2ExponentialMovingAverage) indicator_mod.Indicator {
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
        const self: *T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
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


// Expected data from test_T2.xls, T2(5, 0.7) — firstIsAverage = true.
fn createT2Length(length: usize, first_is_average: bool, volume: f64) !T2ExponentialMovingAverage {
    var t2 = try T2ExponentialMovingAverage.initLength(.{
        .length = length,
        .volume_factor = volume,
        .first_is_average = first_is_average,
    });
    t2.fixSlices();
    return t2;
}

fn createT2Alpha(alpha: f64, first_is_average: bool, volume: f64) !T2ExponentialMovingAverage {
    var t2 = try T2ExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = alpha,
        .volume_factor = volume,
        .first_is_average = first_is_average,
    });
    t2.fixSlices();
    return t2;
}

test "t2 update length 5 firstIsAverage true (t2.xls)" {
    const input = testdata.testInput();
    const exp = testdata.testExpected();
    var t2 = try createT2Length(5, true, 0.7);
    const lprimed = 4 * 5 - 4;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(t2.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = t2.update(input[i]);
        try testing.expect(@abs(exp[i] - act) < 1e-8);
    }

    try testing.expect(math.isNan(t2.update(math.nan(f64))));
}

test "t2 update length 5 firstIsAverage false (Metastock)" {
    const input = testdata.testInput();
    const exp = testdata.testExpected();
    var t2 = try createT2Length(5, false, 0.7);
    const lprimed = 4 * 5 - 4;
    const first_check = lprimed + 43;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(t2.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = t2.update(input[i]);
        if (i >= first_check) {
            try testing.expect(@abs(exp[i] - act) < 1e-8);
        }
    }

    try testing.expect(math.isNan(t2.update(math.nan(f64))));
}

test "t2 isPrimed length 5" {
    const input = testdata.testInput();
    const l = 5;
    const lprimed = 4 * l - 4;

    // firstIsAverage = true
    {
        var t2 = try createT2Length(l, true, 0.7);
        try testing.expect(!t2.isPrimed());
        for (0..lprimed) |i| {
            _ = t2.update(input[i]);
            try testing.expect(!t2.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = t2.update(input[i]);
            try testing.expect(t2.isPrimed());
        }
    }

    // firstIsAverage = false
    {
        var t2 = try createT2Length(l, false, 0.7);
        try testing.expect(!t2.isPrimed());
        for (0..lprimed) |i| {
            _ = t2.update(input[i]);
            try testing.expect(!t2.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = t2.update(input[i]);
            try testing.expect(t2.isPrimed());
        }
    }
}

test "t2 metadata length" {
    var t2 = try createT2Length(10, true, 0.3333);
    var m: Metadata = undefined;
    t2.getMetadata(&m);

    try testing.expectEqual(Identifier.t2_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("t2(10, 0.33330000)", m.mnemonic);
    try testing.expectEqualStrings("T2 exponential moving average t2(10, 0.33330000)", m.description);
}

test "t2 metadata alpha" {
    const alpha: f64 = 2.0 / 11.0;
    var t2 = try createT2Alpha(alpha, false, 0.3333333);
    var m: Metadata = undefined;
    t2.getMetadata(&m);

    try testing.expectEqual(Identifier.t2_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("t2(10, 0.18181818, 0.33333330)", m.mnemonic);
    try testing.expectEqualStrings("T2 exponential moving average t2(10, 0.18181818, 0.33333330)", m.description);
}

test "t2 update entity" {
    const inp: f64 = 3.0;
    const exp_false: f64 = 2.0281481481481483;
    const exp_true: f64 = 1.9555555555555555;
    const time: i64 = 1617235200;
    const l = 2;
    const lprimed = 4 * l - 4;

    // scalar
    {
        var t2 = try createT2Length(l, false, 0.7);
        for (0..lprimed) |_| {
            _ = t2.update(0.0);
        }
        const out = t2.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(@abs(exp_false - s.value) < 1e-13);
    }

    // bar
    {
        var t2 = try createT2Length(l, true, 0.7);
        for (0..lprimed) |_| {
            _ = t2.update(0.0);
        }
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = t2.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(exp_true - s.value) < 1e-13);
    }
}

test "t2 init invalid" {
    // length < 2
    try testing.expectError(error.InvalidLength, T2ExponentialMovingAverage.initLength(.{ .length = 1 }));
    try testing.expectError(error.InvalidLength, T2ExponentialMovingAverage.initLength(.{ .length = 0 }));

    // alpha out of range
    try testing.expectError(error.InvalidSmoothingFactor, T2ExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = -1.0 }));
    try testing.expectError(error.InvalidSmoothingFactor, T2ExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = 2.0 }));

    // volume factor out of range
    try testing.expectError(error.InvalidVolumeFactor, T2ExponentialMovingAverage.initLength(.{ .length = 5, .volume_factor = -0.7 }));
    try testing.expectError(error.InvalidVolumeFactor, T2ExponentialMovingAverage.initLength(.{ .length = 5, .volume_factor = 1.7 }));
}

test "t2 mnemonic with bar component" {
    var t2 = try T2ExponentialMovingAverage.initLength(.{
        .length = 10,
        .volume_factor = 0.7,
        .bar_component = .median,
    });
    t2.fixSlices();
    try testing.expectEqualStrings("t2(10, 0.70000000, hl/2)", t2.line.mnemonic);
    try testing.expectEqualStrings("T2 exponential moving average t2(10, 0.70000000, hl/2)", t2.line.description);
}

test "t2 alpha with quote component" {
    var t2 = try T2ExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = 2.0 / 11.0,
        .volume_factor = 0.7,
        .quote_component = .bid,
    });
    t2.fixSlices();
    try testing.expectEqualStrings("t2(10, 0.18181818, 0.70000000, b)", t2.line.mnemonic);
}
