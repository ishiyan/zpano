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

/// Enumerates the outputs of the Arnaud Legoux moving average indicator.
pub const ArnaudLegouxMovingAverageOutput = enum(u8) {
    /// The scalar value of the moving average.
    value = 1,
};

/// Parameters to create an instance of the Arnaud Legoux moving average indicator.
pub const ArnaudLegouxMovingAverageParams = struct {
    /// The window size. Must be >= 1.
    window: usize,
    /// The Gaussian sigma parameter. Must be > 0.
    sigma: f64 = 6.0,
    /// The offset parameter. Must be between 0 and 1 inclusive.
    offset: f64 = 0.85,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Arnaud Legoux Moving Average (ALMA).
///
/// ALMA is a Gaussian-weighted moving average that reduces lag while maintaining
/// smoothness. It applies a Gaussian bell curve as its kernel, shifted toward
/// recent bars via an adjustable offset parameter.
///
/// The indicator is not primed during the first (window - 1) updates.
pub const ArnaudLegouxMovingAverage = struct {
    line: LineIndicator,
    weights: []f64,
    buffer: []f64,
    window_length: usize,
    buffer_count: usize,
    buffer_index: usize,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: ArnaudLegouxMovingAverageParams) !ArnaudLegouxMovingAverage {
        const window = params.window;
        if (window < 1) {
            return error.InvalidWindow;
        }

        const sigma = params.sigma;
        if (sigma <= 0) {
            return error.InvalidSigma;
        }

        const offset = params.offset;
        if (offset < 0 or offset > 1) {
            return error.InvalidOffset;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Build mnemonic: "alma({window}, {sigma}, {offset}{triple})"
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "alma({d}, {d}, {d}{s})", .{ window, sigma, offset, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Arnaud Legoux moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        // Precompute Gaussian weights.
        const m = offset * @as(f64, @floatFromInt(window - 1));
        const s = @as(f64, @floatFromInt(window)) / sigma;

        const weights = try allocator.alloc(f64, window);
        var norm: f64 = 0.0;

        for (0..window) |i| {
            const diff = @as(f64, @floatFromInt(i)) - m;
            const w = @exp(-(diff * diff) / (2.0 * s * s));
            weights[i] = w;
            norm += w;
        }

        for (weights) |*w| {
            w.* /= norm;
        }

        const buffer = try allocator.alloc(f64, window);
        @memset(buffer, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .weights = weights,
            .buffer = buffer,
            .window_length = window,
            .buffer_count = 0,
            .buffer_index = 0,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *ArnaudLegouxMovingAverage) void {
        self.allocator.free(self.weights);
        self.allocator.free(self.buffer);
    }

    /// After init, fix up the line's mnemonic/description slices to point into
    /// `self`'s own buffers (not the stack-local ones from `init`).
    pub fn fixSlices(self: *ArnaudLegouxMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update logic. Returns the ALMA value or NaN if not yet primed.
    pub fn update(self: *ArnaudLegouxMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const window = self.window_length;

        if (window == 1) {
            self.primed = true;
            return sample;
        }

        // Fill the circular buffer.
        self.buffer[self.buffer_index] = sample;
        self.buffer_index = (self.buffer_index + 1) % window;

        if (!self.primed) {
            self.buffer_count += 1;
            if (self.buffer_count < window) {
                return math.nan(f64);
            }

            self.primed = true;
        }

        // Compute weighted sum.
        // Weight[0] applies to oldest sample, weight[N-1] to newest.
        // The oldest sample is at self.buffer_index (circular buffer).
        var result: f64 = 0.0;
        var index = self.buffer_index;

        for (0..window) |i| {
            result += self.weights[i] * self.buffer[index];
            index = (index + 1) % window;
        }

        return result;
    }

    /// Returns whether the indicator has accumulated enough data.
    pub fn isPrimed(self: *const ArnaudLegouxMovingAverage) bool {
        return self.primed;
    }

    /// Returns metadata for this indicator.
    pub fn getMetadata(self: *const ArnaudLegouxMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .arnaud_legoux_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *ArnaudLegouxMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *ArnaudLegouxMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *ArnaudLegouxMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *ArnaudLegouxMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *ArnaudLegouxMovingAverage) indicator_mod.Indicator {
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
        const self: *ArnaudLegouxMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const ArnaudLegouxMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *ArnaudLegouxMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *ArnaudLegouxMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *ArnaudLegouxMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *ArnaudLegouxMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidWindow,
        InvalidSigma,
        InvalidOffset,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn createAlma(allocator: std.mem.Allocator, window: usize, sigma: f64, offset: f64) !ArnaudLegouxMovingAverage {
    var alma = try ArnaudLegouxMovingAverage.init(allocator, .{ .window = window, .sigma = sigma, .offset = offset });
    alma.fixSlices();
    return alma;
}

fn checkAlmaUpdate(comptime total: usize, alma: *ArnaudLegouxMovingAverage, window: usize, input: *const [total]f64, exp: *const [total]f64) !void {
    const warmup = if (window <= 1) 0 else window - 1;

    for (0..warmup) |i| {
        const act = alma.update(input[i]);
        try testing.expect(math.isNan(act));
    }

    for (warmup..total) |i| {
        const act = alma.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(act - exp[i]) < 1e-13);
        }
    }

    // NaN passthrough.
    try testing.expect(math.isNan(alma.update(math.nan(f64))));
}

test "alma w9 s6 o0.85 (default)" {
    const input = testdata.testInput();
    const exp = testdata.expectedW9_S6_O0_85();
    var alma = try createAlma(testing.allocator, 9, 6.0, 0.85);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 9, &input, &exp);
}

test "alma w9 s6 o0.5" {
    const input = testdata.testInput();
    const exp = testdata.expectedW9_S6_O0_5();
    var alma = try createAlma(testing.allocator, 9, 6.0, 0.5);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 9, &input, &exp);
}

test "alma w10 s6 o0.85" {
    const input = testdata.testInput();
    const exp = testdata.expectedW10_S6_O0_85();
    var alma = try createAlma(testing.allocator, 10, 6.0, 0.85);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 10, &input, &exp);
}

test "alma w5 s6 o0.9" {
    const input = testdata.testInput();
    const exp = testdata.expectedW5_S6_O0_9();
    var alma = try createAlma(testing.allocator, 5, 6.0, 0.9);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 5, &input, &exp);
}

test "alma w1 s6 o0.85 (passthrough)" {
    const input = testdata.testInput();
    const exp = testdata.expectedW1_S6_O0_85();
    var alma = try createAlma(testing.allocator, 1, 6.0, 0.85);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 1, &input, &exp);
}

test "alma w3 s6 o0.85" {
    const input = testdata.testInput();
    const exp = testdata.expectedW3_S6_O0_85();
    var alma = try createAlma(testing.allocator, 3, 6.0, 0.85);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 3, &input, &exp);
}

test "alma w21 s6 o0.85" {
    const input = testdata.testInput();
    const exp = testdata.expectedW21_S6_O0_85();
    var alma = try createAlma(testing.allocator, 21, 6.0, 0.85);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 21, &input, &exp);
}

test "alma w50 s6 o0.85" {
    const input = testdata.testInput();
    const exp = testdata.expectedW50_S6_O0_85();
    var alma = try createAlma(testing.allocator, 50, 6.0, 0.85);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 50, &input, &exp);
}

test "alma w9 s6 o0 (left-aligned)" {
    const input = testdata.testInput();
    const exp = testdata.expectedW9_S6_O0();
    var alma = try createAlma(testing.allocator, 9, 6.0, 0.0);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 9, &input, &exp);
}

test "alma w9 s6 o1 (right-aligned)" {
    const input = testdata.testInput();
    const exp = testdata.expectedW9_S6_O1();
    var alma = try createAlma(testing.allocator, 9, 6.0, 1.0);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 9, &input, &exp);
}

test "alma w9 s2 o0.85 (narrow)" {
    const input = testdata.testInput();
    const exp = testdata.expectedW9_S2_O0_85();
    var alma = try createAlma(testing.allocator, 9, 2.0, 0.85);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 9, &input, &exp);
}

test "alma w9 s20 o0.85 (wide)" {
    const input = testdata.testInput();
    const exp = testdata.expectedW9_S20_O0_85();
    var alma = try createAlma(testing.allocator, 9, 20.0, 0.85);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 9, &input, &exp);
}

test "alma w9 s0.5 o0.85" {
    const input = testdata.testInput();
    const exp = testdata.expectedW9_S0_5_O0_85();
    var alma = try createAlma(testing.allocator, 9, 0.5, 0.85);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 9, &input, &exp);
}

test "alma w15 s4 o0.7" {
    const input = testdata.testInput();
    const exp = testdata.expectedW15_S4_O0_7();
    var alma = try createAlma(testing.allocator, 15, 4.0, 0.7);
    defer alma.deinit();
    try checkAlmaUpdate(252, &alma, 15, &input, &exp);
}

test "alma is primed" {
    const input = testdata.testInput();
    var alma = try createAlma(testing.allocator, 9, 6.0, 0.85);
    defer alma.deinit();

    try testing.expect(!alma.isPrimed());
    for (0..8) |i| {
        _ = alma.update(input[i]);
        try testing.expect(!alma.isPrimed());
    }
    for (8..252) |i| {
        _ = alma.update(input[i]);
        try testing.expect(alma.isPrimed());
    }
}

test "alma is primed window 1" {
    var alma = try createAlma(testing.allocator, 1, 6.0, 0.85);
    defer alma.deinit();

    try testing.expect(!alma.isPrimed());
    _ = alma.update(42.0);
    try testing.expect(alma.isPrimed());
}

test "alma metadata" {
    var alma = try createAlma(testing.allocator, 9, 6.0, 0.85);
    defer alma.deinit();
    var m: Metadata = undefined;
    alma.getMetadata(&m);

    try testing.expectEqual(Identifier.arnaud_legoux_moving_average, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("alma(9, 6, 0.85)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Arnaud Legoux moving average alma(9, 6, 0.85)", m.outputs_buf[0].description);
}

test "alma update entity" {
    const window: usize = 9;
    const time: i64 = 1617235200;
    const input = testdata.testInput();

    // scalar
    {
        var alma = try createAlma(testing.allocator, window, 6.0, 0.85);
        defer alma.deinit();
        for (0..8) |i| {
            _ = alma.update(input[i]);
        }
        const out = alma.updateScalar(&.{ .time = time, .value = input[8] });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        const exp = testdata.expectedW9_S6_O0_85();
        try testing.expect(@abs(s.value - exp[8]) < 1e-13);
    }

    // bar
    {
        var alma = try createAlma(testing.allocator, window, 6.0, 0.85);
        defer alma.deinit();
        for (0..8) |i| {
            _ = alma.update(input[i]);
        }
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = input[8], .volume = 0 };
        const out = alma.updateBar(&bar);
        const s = out.slice()[0].scalar;
        const exp = testdata.expectedW9_S6_O0_85();
        try testing.expect(@abs(s.value - exp[8]) < 1e-13);
    }

    // quote
    {
        var alma = try createAlma(testing.allocator, window, 6.0, 0.85);
        defer alma.deinit();
        for (0..8) |i| {
            _ = alma.update(input[i]);
        }
        const quote = Quote{ .time = time, .bid_price = input[8], .ask_price = input[8], .bid_size = 0, .ask_size = 0 };
        const out = alma.updateQuote(&quote);
        const s = out.slice()[0].scalar;
        const exp = testdata.expectedW9_S6_O0_85();
        try testing.expect(@abs(s.value - exp[8]) < 1e-13);
    }

    // trade
    {
        var alma = try createAlma(testing.allocator, window, 6.0, 0.85);
        defer alma.deinit();
        for (0..8) |i| {
            _ = alma.update(input[i]);
        }
        const trade = Trade{ .time = time, .price = input[8], .volume = 0 };
        const out = alma.updateTrade(&trade);
        const s = out.slice()[0].scalar;
        const exp = testdata.expectedW9_S6_O0_85();
        try testing.expect(@abs(s.value - exp[8]) < 1e-13);
    }
}

test "alma init invalid params" {
    const r1 = ArnaudLegouxMovingAverage.init(testing.allocator, .{ .window = 0 });
    try testing.expectError(error.InvalidWindow, r1);

    const r2 = ArnaudLegouxMovingAverage.init(testing.allocator, .{ .window = 9, .sigma = 0 });
    try testing.expectError(error.InvalidSigma, r2);

    const r3 = ArnaudLegouxMovingAverage.init(testing.allocator, .{ .window = 9, .sigma = -1.0 });
    try testing.expectError(error.InvalidSigma, r3);

    const r4 = ArnaudLegouxMovingAverage.init(testing.allocator, .{ .window = 9, .offset = -0.1 });
    try testing.expectError(error.InvalidOffset, r4);

    const r5 = ArnaudLegouxMovingAverage.init(testing.allocator, .{ .window = 9, .offset = 1.1 });
    try testing.expectError(error.InvalidOffset, r5);
}

test "alma mnemonic components" {
    // all defaults -> no component suffix
    {
        var alma = try createAlma(testing.allocator, 9, 6.0, 0.85);
        defer alma.deinit();
        try testing.expectEqualStrings("alma(9, 6, 0.85)", alma.line.mnemonic);
    }

    // bar component set to Median
    {
        var alma = try ArnaudLegouxMovingAverage.init(testing.allocator, .{
            .window = 9,
            .sigma = 6.0,
            .offset = 0.85,
            .bar_component = .median,
        });
        defer alma.deinit();
        alma.fixSlices();
        try testing.expectEqualStrings("alma(9, 6, 0.85, hl/2)", alma.line.mnemonic);
        try testing.expectEqualStrings("Arnaud Legoux moving average alma(9, 6, 0.85, hl/2)", alma.line.description);
    }

    // bar=high, trade=volume
    {
        var alma = try ArnaudLegouxMovingAverage.init(testing.allocator, .{
            .window = 9,
            .sigma = 6.0,
            .offset = 0.85,
            .bar_component = .high,
            .trade_component = .volume,
        });
        defer alma.deinit();
        alma.fixSlices();
        try testing.expectEqualStrings("alma(9, 6, 0.85, h, v)", alma.line.mnemonic);
    }
}
