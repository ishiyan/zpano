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

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the simple moving average indicator.
pub const SimpleMovingAverageOutput = enum(u8) {
    /// The scalar value of the moving average.
    value = 1,
};

/// Parameters to create an instance of the simple moving average indicator.
pub const SimpleMovingAverageParams = struct {
    /// The length (number of time periods) of the moving window. Must be > 1.
    length: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the simple (arithmetic) moving average (SMA).
///
/// SMAᵢ = SMAᵢ₋₁ + (Pᵢ − Pᵢ₋ℓ) / ℓ
///
/// The indicator is not primed during the first ℓ−1 updates.
pub const SimpleMovingAverage = struct {
    line: LineIndicator,
    window: []f64,
    window_sum: f64,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    primed: bool,
    allocator: std.mem.Allocator,
    // Fixed buffers for mnemonic and description strings.
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: SimpleMovingAverageParams) !SimpleMovingAverage {
        if (params.length < 2) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Build mnemonic: "sma({length}{triple})"
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "sma({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Simple moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window = try allocator.alloc(f64, params.length);
        @memset(window, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .window = window,
            .window_sum = 0.0,
            .window_length = params.length,
            .window_count = 0,
            .last_index = params.length - 1,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *SimpleMovingAverage) void {
        self.allocator.free(self.window);
    }

    /// After init, fix up the line's mnemonic/description slices to point into
    /// `self`'s own buffers (not the stack-local ones from `init`).
    fn fixSlices(self: *SimpleMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update logic. Returns the SMA value or NaN if not yet primed.
    pub fn update(self: *SimpleMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            self.window_sum += sample - self.window[0];

            var i: usize = 0;
            while (i < self.last_index) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = sample;
        } else {
            self.window_sum += sample;
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if (self.window_length > self.window_count) {
                return math.nan(f64);
            }

            self.primed = true;
        }

        return self.window_sum / @as(f64, @floatFromInt(self.window_length));
    }

    /// Returns whether the indicator has accumulated enough data.
    pub fn isPrimed(self: *const SimpleMovingAverage) bool {
        return self.primed;
    }

    /// Returns metadata for this indicator.
    pub fn getMetadata(self: *const SimpleMovingAverage) Metadata {
        return build_metadata_mod.buildMetadata(
            .simple_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *SimpleMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *SimpleMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *SimpleMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *SimpleMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *SimpleMovingAverage) indicator_mod.Indicator {
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
        const self: *SimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque) Metadata {
        const self: *const SimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.getMetadata();
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *SimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *SimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *SimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *SimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidLength,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn testInput() [51]f64 {
    return .{
        64.59, 64.23, 65.26, 65.24, 65.07, 65.14, 64.98, 64.76, 65.11, 65.46,
        65.94, 66.10, 66.87, 66.56, 66.71, 66.19, 66.14, 66.64, 67.33, 68.18,
        67.48, 67.19, 66.46, 67.20, 67.62, 67.66, 67.89, 69.19, 69.68, 69.31,
        69.11, 69.27, 68.97, 69.11, 69.50, 69.70, 69.94, 69.11, 67.64, 67.75,
        67.47, 67.50, 68.18, 67.35, 66.74, 67.00, 67.46, 67.36, 67.37, 67.78,
        67.96,
    };
}

fn expected3() [51]f64 {
    return .{
        math.nan(f64), math.nan(f64), 64.69, 64.91, 65.19, 65.15, 65.06, 64.96, 64.95, 65.11,
        65.50,         65.83,         66.30, 66.51, 66.71, 66.49, 66.35, 66.32, 66.70, 67.38,
        67.66,         67.62,         67.04, 66.95, 67.09, 67.49, 67.72, 68.25, 68.92, 69.39,
        69.37,         69.23,         69.12, 69.12, 69.19, 69.44, 69.71, 69.58, 68.90, 68.17,
        67.62,         67.57,         67.72, 67.68, 67.42, 67.03, 67.07, 67.27, 67.40, 67.50,
        67.70,
    };
}

fn expected5() [51]f64 {
    return .{
        math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), 64.88, 64.99, 65.14, 65.04, 65.01, 65.09,
        65.25,         65.47,         65.90,         66.19,         66.44, 66.49, 66.49, 66.45, 66.60, 66.90,
        67.15,         67.36,         67.33,         67.30,         67.19, 67.23, 67.37, 67.91, 68.41, 68.75,
        69.04,         69.31,         69.27,         69.15,         69.19, 69.31, 69.44, 69.47, 69.18, 68.83,
        68.38,         67.89,         67.71,         67.65,         67.45, 67.35, 67.35, 67.18, 67.19, 67.39,
        67.59,
    };
}

fn expected10() [51]f64 {
    return .{
        math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), math.nan(f64), 64.98,
        65.12,         65.31,         65.47,         65.60,         65.76,         65.87,         65.98,         66.17,         66.39,         66.67,
        66.82,         66.93,         66.89,         66.95,         67.04,         67.19,         67.37,         67.62,         67.86,         67.97,
        68.13,         68.34,         68.59,         68.78,         68.97,         69.17,         69.38,         69.37,         69.17,         69.01,
        68.85,         68.67,         68.59,         68.41,         68.14,         67.87,         67.62,         67.45,         67.42,         67.42,
        67.47,
    };
}

fn createSma(allocator: std.mem.Allocator, length: usize) !SimpleMovingAverage {
    var sma = try SimpleMovingAverage.init(allocator, .{ .length = length });
    sma.fixSlices();
    return sma;
}

fn checkUpdate(comptime nan_count: usize, comptime total: usize, sma: *SimpleMovingAverage, input: *const [total]f64, exp: *const [total]f64) !void {
    for (0..nan_count) |i| {
        const act = sma.update(input[i]);
        try testing.expect(math.isNan(act));
    }
    for (nan_count..total) |i| {
        const act = sma.update(input[i]);
        try testing.expect(@abs(exp[i] - act) < 1e-2);
    }
    // NaN passthrough
    try testing.expect(math.isNan(sma.update(math.nan(f64))));
}

test "sma update length 3" {
    const input = testInput();
    const exp = expected3();
    var sma = try createSma(testing.allocator, 3);
    defer sma.deinit();
    try checkUpdate(2, 51, &sma, &input, &exp);
}

test "sma update length 5" {
    const input = testInput();
    const exp = expected5();
    var sma = try createSma(testing.allocator, 5);
    defer sma.deinit();
    try checkUpdate(4, 51, &sma, &input, &exp);
}

test "sma update length 10" {
    const input = testInput();
    const exp = expected10();
    var sma = try createSma(testing.allocator, 10);
    defer sma.deinit();
    try checkUpdate(9, 51, &sma, &input, &exp);
}

test "sma is primed" {
    const input = testInput();
    var sma = try createSma(testing.allocator, 3);
    defer sma.deinit();

    try testing.expect(!sma.isPrimed());
    for (0..2) |i| {
        _ = sma.update(input[i]);
        try testing.expect(!sma.isPrimed());
    }
    for (2..51) |i| {
        _ = sma.update(input[i]);
        try testing.expect(sma.isPrimed());
    }
}

test "sma metadata" {
    var sma = try createSma(testing.allocator, 5);
    defer sma.deinit();
    const m = sma.getMetadata();

    try testing.expectEqual(Identifier.simple_moving_average, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs.len);
    try testing.expectEqual(@as(i32, 1), m.outputs[0].kind);
    try testing.expectEqualStrings("sma(5)", m.outputs[0].mnemonic);
    try testing.expectEqualStrings("Simple moving average sma(5)", m.outputs[0].description);
}

test "sma update entity" {
    const length: usize = 2;
    const inp: f64 = 3.0;
    const exp: f64 = inp / @as(f64, @floatFromInt(length));
    const time: i64 = 1617235200;

    // scalar
    {
        var sma = try createSma(testing.allocator, length);
        defer sma.deinit();
        _ = sma.update(0.0);
        const out = sma.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expectEqual(exp, s.value);
    }

    // bar
    {
        var sma = try createSma(testing.allocator, length);
        defer sma.deinit();
        _ = sma.update(0.0);
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = sma.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(exp, s.value);
    }

    // quote
    {
        var sma = try createSma(testing.allocator, length);
        defer sma.deinit();
        _ = sma.update(0.0);
        const quote = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };
        const out = sma.updateQuote(&quote);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(exp, s.value);
    }

    // trade
    {
        var sma = try createSma(testing.allocator, length);
        defer sma.deinit();
        _ = sma.update(0.0);
        const trade = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = sma.updateTrade(&trade);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(exp, s.value);
    }
}

test "sma init invalid length" {
    const result = SimpleMovingAverage.init(testing.allocator, .{ .length = 1 });
    try testing.expectError(error.InvalidLength, result);

    const result0 = SimpleMovingAverage.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, result0);
}

test "sma mnemonic components" {
    // all defaults -> no component suffix
    {
        var sma = try createSma(testing.allocator, 5);
        defer sma.deinit();
        try testing.expectEqualStrings("sma(5)", sma.line.mnemonic);
    }

    // bar component set to Median
    {
        var sma = try SimpleMovingAverage.init(testing.allocator, .{
            .length = 5,
            .bar_component = .median,
        });
        defer sma.deinit();
        sma.fixSlices();
        try testing.expectEqualStrings("sma(5, hl/2)", sma.line.mnemonic);
        try testing.expectEqualStrings("Simple moving average sma(5, hl/2)", sma.line.description);
    }

    // bar=high, trade=volume
    {
        var sma = try SimpleMovingAverage.init(testing.allocator, .{
            .length = 5,
            .bar_component = .high,
            .trade_component = .volume,
        });
        defer sma.deinit();
        sma.fixSlices();
        try testing.expectEqualStrings("sma(5, h, v)", sma.line.mnemonic);
    }
}
