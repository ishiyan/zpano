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

/// Enumerates the outputs of the triple exponential moving average indicator.
pub const TripleExponentialMovingAverageOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the TEMA based on length.
pub const TripleExponentialMovingAverageLengthParams = struct {
    length: usize,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an instance of the TEMA based on smoothing factor.
pub const TripleExponentialMovingAverageSmoothingFactorParams = struct {
    smoothing_factor: f64,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Triple Exponential Moving Average (TEMA).
///
///   EMA¹ᵢ = EMA(Pᵢ)
///   EMA²ᵢ = EMA(EMA¹ᵢ)
///   EMA³ᵢ = EMA(EMA²ᵢ)
///   TEMAᵢ = 3·(EMA¹ᵢ - EMA²ᵢ) + EMA³ᵢ
///
/// Warmup period: 3·length - 3 samples before primed.
pub const TripleExponentialMovingAverage = struct {
    line: LineIndicator,
    smoothing_factor: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    ema3: f64,
    length: usize,
    length2: usize,
    length3: usize,
    count: usize,
    first_is_average: bool,
    primed: bool,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    const epsilon: f64 = 0.00000001;

    /// Create TEMA from length (min 2).
    pub fn initLength(params: TripleExponentialMovingAverageLengthParams) !TripleExponentialMovingAverage {
        if (params.length < 2) {
            return error.InvalidLength;
        }
        const alpha = 2.0 / @as(f64, @floatFromInt(1 + params.length));
        return initInternal(params.length, alpha, false, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    /// Create TEMA from smoothing factor.
    pub fn initSmoothingFactor(params: TripleExponentialMovingAverageSmoothingFactorParams) !TripleExponentialMovingAverage {
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
    ) !TripleExponentialMovingAverage {
        const bc = bc_opt orelse bar_component.default_bar_component;
        const qc = qc_opt orelse quote_component.default_quote_component;
        const tc = tc_opt orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        var mnemonic_slice: []u8 = undefined;
        if (is_alpha_mode) {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "tema({d}, {d:.8}{s})", .{ length, alpha, triple }) catch
                return error.MnemonicTooLong;
        } else {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "tema({d}{s})", .{ length, triple }) catch
                return error.MnemonicTooLong;
        }
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Triple exponential moving average {s}", .{mnemonic_slice}) catch
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
            .ema3 = 0.0,
            .length = length,
            .length2 = 2 * length - 1,
            .length3 = 3 * length - 2,
            .count = 0,
            .first_is_average = first_is_average,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *TripleExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *TripleExponentialMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            const sf = self.smoothing_factor;
            var v1 = self.ema1;
            var v2 = self.ema2;
            var v3 = self.ema3;
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            v3 += (v2 - v3) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            self.ema3 = v3;
            return 3.0 * (v1 - v2) + v3;
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
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.sum += self.ema2;

                if (self.length3 == self.count) {
                    self.primed = true;
                    self.ema3 = self.sum / @as(f64, @floatFromInt(self.length));
                    return 3.0 * (self.ema1 - self.ema2) + self.ema3;
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
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;

                if (self.length3 == self.count) {
                    self.primed = true;
                    return 3.0 * (self.ema1 - self.ema2) + self.ema3;
                }
            }
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const TripleExponentialMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const TripleExponentialMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .triple_exponential_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *TripleExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *TripleExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *TripleExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *TripleExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *TripleExponentialMovingAverage) indicator_mod.Indicator {
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
        const self: *TripleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const TripleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *TripleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *TripleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *TripleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *TripleExponentialMovingAverage = @ptrCast(@alignCast(ptr));
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


fn createTemaLength(length: usize, first_is_average: bool) !TripleExponentialMovingAverage {
    var tema = try TripleExponentialMovingAverage.initLength(.{
        .length = length,
        .first_is_average = first_is_average,
    });
    tema.fixSlices();
    return tema;
}

fn createTemaAlpha(alpha: f64, first_is_average: bool) !TripleExponentialMovingAverage {
    var tema = try TripleExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = alpha,
        .first_is_average = first_is_average,
    });
    tema.fixSlices();
    return tema;
}

test "tema update length 14 firstIsAverage true" {
    const input = testdata.testInput();
    var tema = try createTemaLength(14, true);
    const lprimed = 3 * 14 - 3;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(tema.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = tema.update(input[i]);
        switch (i) {
            39 => try testing.expect(@abs(84.8629 - act) < 1e-3),
            40 => try testing.expect(@abs(84.2246 - act) < 1e-3),
            251 => try testing.expect(@abs(108.418 - act) < 1e-3),
            else => {},
        }
    }

    try testing.expect(math.isNan(tema.update(math.nan(f64))));
}

test "tema update length 14 firstIsAverage false (Metastock)" {
    const input = testdata.testInput();
    var tema = try createTemaLength(14, false);
    const lprimed = 3 * 14 - 3;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(tema.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = tema.update(input[i]);
        switch (i) {
            39 => try testing.expect(@abs(84.721 - act) < 1e-3),
            40 => try testing.expect(@abs(84.089 - act) < 1e-3),
            251 => try testing.expect(@abs(108.418 - act) < 1e-3),
            else => {},
        }
    }

    try testing.expect(math.isNan(tema.update(math.nan(f64))));
}

test "tema update length 26 firstIsAverage false (Metastock) TASC" {
    const l = 26;
    const lprimed = 3 * l - 3;

    var tema = try createTemaLength(l, false);

    const in = testdata.testTascInput();
    const exp = testdata.testTascExpected();

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(tema.update(in[i])));
    }

    // Expected array is indexed by full input index; only validate from index 216 onward
    // (matching Go's firstCheck = 216).
    const first_check = 216;
    for (lprimed..316) |i| {
        const act = tema.update(in[i]);
        if (i >= first_check) {
            try testing.expect(@abs(exp[i] - act) < 1e-3);
        }
    }

    try testing.expect(math.isNan(tema.update(math.nan(f64))));
}

test "tema isPrimed length 14" {
    const input = testdata.testInput();
    const l = 14;
    const lprimed = 3 * l - 3;

    // firstIsAverage = true
    {
        var tema = try createTemaLength(l, true);
        try testing.expect(!tema.isPrimed());
        for (0..lprimed) |i| {
            _ = tema.update(input[i]);
            try testing.expect(!tema.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = tema.update(input[i]);
            try testing.expect(tema.isPrimed());
        }
    }

    // firstIsAverage = false
    {
        var tema = try createTemaLength(l, false);
        try testing.expect(!tema.isPrimed());
        for (0..lprimed) |i| {
            _ = tema.update(input[i]);
            try testing.expect(!tema.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = tema.update(input[i]);
            try testing.expect(tema.isPrimed());
        }
    }
}

test "tema metadata length" {
    var tema = try createTemaLength(10, true);
    var m: Metadata = undefined;
    tema.getMetadata(&m);

    try testing.expectEqual(Identifier.triple_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("tema(10)", m.mnemonic);
    try testing.expectEqualStrings("Triple exponential moving average tema(10)", m.description);
}

test "tema metadata alpha" {
    const alpha: f64 = 2.0 / 11.0;
    var tema = try createTemaAlpha(alpha, false);
    var m: Metadata = undefined;
    tema.getMetadata(&m);

    try testing.expectEqual(Identifier.triple_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("tema(10, 0.18181818)", m.mnemonic);
    try testing.expectEqualStrings("Triple exponential moving average tema(10, 0.18181818)", m.description);
}

test "tema update entity" {
    const inp: f64 = 3.0;
    const exp_false: f64 = 2.888888888888889;
    const exp_true: f64 = 2.6666666666666665;
    const time: i64 = 1617235200;

    // scalar (firstIsAverage=false)
    {
        var tema = try createTemaLength(2, false);
        // lprimed = 3*2-3 = 3, feed 3 zeros
        _ = tema.update(0.0);
        _ = tema.update(0.0);
        _ = tema.update(0.0);
        const out = tema.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(@abs(exp_false - s.value) < 1e-13);
    }

    // bar (firstIsAverage=true)
    {
        var tema = try createTemaLength(2, true);
        _ = tema.update(0.0);
        _ = tema.update(0.0);
        _ = tema.update(0.0);
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = tema.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(exp_true - s.value) < 1e-13);
    }
}

test "tema init invalid" {
    // length = 1 (min is 2)
    try testing.expectError(error.InvalidLength, TripleExponentialMovingAverage.initLength(.{ .length = 1 }));
    // length = 0
    try testing.expectError(error.InvalidLength, TripleExponentialMovingAverage.initLength(.{ .length = 0 }));

    // alpha < 0
    try testing.expectError(error.InvalidSmoothingFactor, TripleExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = -1.0 }));
    // alpha > 1
    try testing.expectError(error.InvalidSmoothingFactor, TripleExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = 2.0 }));
}

test "tema mnemonic with bar component" {
    var tema = try TripleExponentialMovingAverage.initLength(.{
        .length = 10,
        .bar_component = .median,
    });
    tema.fixSlices();
    try testing.expectEqualStrings("tema(10, hl/2)", tema.line.mnemonic);
    try testing.expectEqualStrings("Triple exponential moving average tema(10, hl/2)", tema.line.description);
}

test "tema alpha with quote component" {
    var tema = try TripleExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = 2.0 / 11.0,
        .quote_component = .bid,
    });
    tema.fixSlices();
    try testing.expectEqualStrings("tema(10, 0.18181818, b)", tema.line.mnemonic);
}
