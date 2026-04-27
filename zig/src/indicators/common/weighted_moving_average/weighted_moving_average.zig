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

/// Enumerates the outputs of the weighted moving average indicator.
pub const WeightedMovingAverageOutput = enum(u8) {
    /// The scalar value of the weighted moving average.
    value = 1,
};

/// Parameters to create an instance of the weighted moving average indicator.
pub const WeightedMovingAverageParams = struct {
    /// The length (number of time periods) of the moving window. Must be > 1.
    length: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the weighted moving average (WMA) that has multiplying factors
/// to give arithmetically decreasing weights to the samples in the look back window.
///
///   WMAᵢ = (ℓPᵢ + (ℓ-1)Pᵢ₋₁ + ... + Pᵢ₋ℓ) / (ℓ + (ℓ-1) + ... + 2 + 1)
///
/// The indicator is not primed during the first ℓ−1 updates.
pub const WeightedMovingAverage = struct {
    line: LineIndicator,
    window: []f64,
    window_sum: f64,
    window_sub: f64,
    divider: f64,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: WeightedMovingAverageParams) !WeightedMovingAverage {
        if (params.length < 2) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "wma({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Weighted moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const length_f: f64 = @floatFromInt(params.length);
        const divider = length_f * (length_f + 1.0) / 2.0;

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
            .window_sub = 0.0,
            .divider = divider,
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

    pub fn deinit(self: *WeightedMovingAverage) void {
        self.allocator.free(self.window);
    }

    fn fixSlices(self: *WeightedMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *WeightedMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const temp = sample;

        if (self.primed) {
            self.window_sum -= self.window_sub;
            self.window_sum += temp * @as(f64, @floatFromInt(self.window_length));
            self.window_sub -= self.window[0];
            self.window_sub += temp;

            var i: usize = 0;
            while (i < self.last_index) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = temp;
        } else {
            self.window[self.window_count] = temp;
            self.window_sub += temp;
            self.window_count += 1;
            self.window_sum += temp * @as(f64, @floatFromInt(self.window_count));

            if (self.window_length > self.window_count) {
                return math.nan(f64);
            }

            self.primed = true;
        }

        return self.window_sum / self.divider;
    }

    pub fn isPrimed(self: *const WeightedMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const WeightedMovingAverage) Metadata {
        return build_metadata_mod.buildMetadata(
            .weighted_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *WeightedMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *WeightedMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *WeightedMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *WeightedMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *WeightedMovingAverage) indicator_mod.Indicator {
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
        const self: *WeightedMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque) Metadata {
        const self: *const WeightedMovingAverage = @ptrCast(@alignCast(ptr));
        return self.getMetadata();
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *WeightedMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *WeightedMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *WeightedMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *WeightedMovingAverage = @ptrCast(@alignCast(ptr));
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

fn createWma(allocator: std.mem.Allocator, length: usize) !WeightedMovingAverage {
    var wma = try WeightedMovingAverage.init(allocator, .{ .length = length });
    wma.fixSlices();
    return wma;
}

test "wma update length 2" {
    const input = testInput();
    var wma = try createWma(testing.allocator, 2);
    defer wma.deinit();

    // Index 0: NaN
    try testing.expect(math.isNan(wma.update(input[0])));
    // Index 1: 93.71
    try testing.expect(@abs(93.71 - wma.update(input[1])) < 1e-2);
    // Index 2: 94.52
    try testing.expect(@abs(94.52 - wma.update(input[2])) < 1e-2);
    // Index 3: 94.855
    try testing.expect(@abs(94.855 - wma.update(input[3])) < 1e-2);

    // Run remaining
    for (4..252) |i| {
        _ = wma.update(input[i]);
    }
    // We need to check the last value separately — recompute from scratch
    var wma2 = try createWma(testing.allocator, 2);
    defer wma2.deinit();
    var last: f64 = undefined;
    for (0..252) |i| {
        last = wma2.update(input[i]);
    }
    try testing.expect(@abs(108.16 - last) < 1e-2);

    // NaN passthrough
    try testing.expect(math.isNan(wma2.update(math.nan(f64))));
}

test "wma update length 30" {
    const input = testInput();
    var wma = try createWma(testing.allocator, 30);
    defer wma.deinit();

    for (0..29) |i| {
        try testing.expect(math.isNan(wma.update(input[i])));
    }

    var results: [252]f64 = undefined;
    // Re-init to get all results
    var wma2 = try createWma(testing.allocator, 30);
    defer wma2.deinit();
    for (0..252) |i| {
        results[i] = wma2.update(input[i]);
    }

    try testing.expect(@abs(88.567 - results[29]) < 1e-2);
    try testing.expect(@abs(88.233 - results[30]) < 1e-2);
    try testing.expect(@abs(88.034 - results[31]) < 1e-2);
    try testing.expect(@abs(87.191 - results[58]) < 1e-2);
    try testing.expect(@abs(109.3466 - results[250]) < 1e-2);
    try testing.expect(@abs(109.3413 - results[251]) < 1e-2);

    try testing.expect(math.isNan(wma2.update(math.nan(f64))));
}

test "wma is primed length 2" {
    const input = testInput();
    var wma = try createWma(testing.allocator, 2);
    defer wma.deinit();

    try testing.expect(!wma.isPrimed());
    _ = wma.update(input[0]);
    try testing.expect(!wma.isPrimed());
    _ = wma.update(input[1]);
    try testing.expect(wma.isPrimed());
}

test "wma is primed length 30" {
    const input = testInput();
    var wma = try createWma(testing.allocator, 30);
    defer wma.deinit();

    try testing.expect(!wma.isPrimed());
    for (0..29) |i| {
        _ = wma.update(input[i]);
        try testing.expect(!wma.isPrimed());
    }
    _ = wma.update(input[29]);
    try testing.expect(wma.isPrimed());
}

test "wma metadata" {
    var wma = try createWma(testing.allocator, 5);
    defer wma.deinit();
    const m = wma.getMetadata();

    try testing.expectEqual(Identifier.weighted_moving_average, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs.len);
    try testing.expectEqual(@as(i32, 1), m.outputs[0].kind);
    try testing.expectEqualStrings("wma(5)", m.outputs[0].mnemonic);
    try testing.expectEqualStrings("Weighted moving average wma(5)", m.outputs[0].description);
}

test "wma update entity" {
    const input = testInput();
    const time: i64 = 1617235200;

    // scalar
    {
        var wma = try createWma(testing.allocator, 2);
        defer wma.deinit();
        _ = wma.update(input[0]);
        const out = wma.updateScalar(&.{ .time = time, .value = input[1] });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(@abs(93.71 - s.value) < 1e-2);
    }

    // bar
    {
        var wma = try createWma(testing.allocator, 2);
        defer wma.deinit();
        _ = wma.update(input[0]);
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = input[1], .volume = 0 };
        const out = wma.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(93.71 - s.value) < 1e-2);
    }

    // quote
    {
        var wma = try createWma(testing.allocator, 2);
        defer wma.deinit();
        _ = wma.update(input[0]);
        const quote = Quote{ .time = time, .bid_price = input[1], .ask_price = input[1], .bid_size = 0, .ask_size = 0 };
        const out = wma.updateQuote(&quote);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(93.71 - s.value) < 1e-2);
    }

    // trade
    {
        var wma = try createWma(testing.allocator, 2);
        defer wma.deinit();
        _ = wma.update(input[0]);
        const trade = Trade{ .time = time, .price = input[1], .volume = 0 };
        const out = wma.updateTrade(&trade);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(93.71 - s.value) < 1e-2);
    }
}

test "wma init invalid length" {
    const result = WeightedMovingAverage.init(testing.allocator, .{ .length = 1 });
    try testing.expectError(error.InvalidLength, result);

    const result0 = WeightedMovingAverage.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, result0);
}

test "wma mnemonic components" {
    {
        var wma = try createWma(testing.allocator, 5);
        defer wma.deinit();
        try testing.expectEqualStrings("wma(5)", wma.line.mnemonic);
    }
    {
        var wma = try WeightedMovingAverage.init(testing.allocator, .{
            .length = 5,
            .bar_component = .median,
        });
        defer wma.deinit();
        wma.fixSlices();
        try testing.expectEqualStrings("wma(5, hl/2)", wma.line.mnemonic);
        try testing.expectEqualStrings("Weighted moving average wma(5, hl/2)", wma.line.description);
    }
}
