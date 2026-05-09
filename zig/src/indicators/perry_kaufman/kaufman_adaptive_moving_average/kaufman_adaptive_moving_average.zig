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
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Kaufman Adaptive Moving Average indicator.
pub const KaufmanAdaptiveMovingAverageOutput = enum(u8) {
    /// The scalar value of the moving average.
    value = 1,
};

/// Parameters to create an instance of the indicator based on lengths.
pub const KaufmanAdaptiveMovingAverageLengthParams = struct {
    /// Efficiency ratio length. Must be >= 2.
    efficiency_ratio_length: u32 = 10,
    /// Fastest boundary length. Must be >= 2.
    fastest_length: u32 = 2,
    /// Slowest boundary length. Must be >= 2.
    slowest_length: u32 = 30,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an instance of the indicator based on smoothing factors.
pub const KaufmanAdaptiveMovingAverageSmoothingFactorParams = struct {
    /// Efficiency ratio length. Must be >= 2.
    efficiency_ratio_length: u32 = 10,
    /// Fastest smoothing factor in [0, 1].
    fastest_smoothing_factor: f64 = 2.0 / 3.0,
    /// Slowest smoothing factor in [0, 1].
    slowest_smoothing_factor: f64 = 2.0 / 31.0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Kaufman's Adaptive Moving Average (KAMA).
///
/// An EMA with the smoothing factor adapted by the efficiency ratio.
pub const KaufmanAdaptiveMovingAverage = struct {
    line: LineIndicator,
    efficiency_ratio_length: u32,
    window: []f64,
    absolute_delta: []f64,
    absolute_delta_sum: f64,
    alpha_fastest: f64,
    alpha_slowest: f64,
    alpha_diff: f64,
    value: f64,
    efficiency_ratio: f64,
    window_count: u32,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,

    pub fn initLength(allocator: std.mem.Allocator, params: KaufmanAdaptiveMovingAverageLengthParams) !KaufmanAdaptiveMovingAverage {
        if (params.efficiency_ratio_length < 2) return error.InvalidEfficiencyRatioLength;
        if (params.fastest_length < 2) return error.InvalidFastestLength;
        if (params.slowest_length < 2) return error.InvalidSlowestLength;

        const fastest_alpha = 2.0 / @as(f64, @floatFromInt(1 + params.fastest_length));
        const slowest_alpha = 2.0 / @as(f64, @floatFromInt(1 + params.slowest_length));

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "kama({d}, {d}, {d}{s})", .{
            params.efficiency_ratio_length, params.fastest_length, params.slowest_length, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        return initInternal(allocator, params.efficiency_ratio_length, fastest_alpha, slowest_alpha, params.bar_component, params.quote_component, params.trade_component, mnemonic_buf, mnemonic_len);
    }

    pub fn initSmoothingFactor(allocator: std.mem.Allocator, params: KaufmanAdaptiveMovingAverageSmoothingFactorParams) !KaufmanAdaptiveMovingAverage {
        if (params.efficiency_ratio_length < 2) return error.InvalidEfficiencyRatioLength;
        if (params.fastest_smoothing_factor < 0.0 or params.fastest_smoothing_factor > 1.0) return error.InvalidFastestSmoothingFactor;
        if (params.slowest_smoothing_factor < 0.0 or params.slowest_smoothing_factor > 1.0) return error.InvalidSlowestSmoothingFactor;

        const epsilon = 0.00000001;
        var fastest_alpha = params.fastest_smoothing_factor;
        var slowest_alpha = params.slowest_smoothing_factor;
        if (fastest_alpha < epsilon) fastest_alpha = epsilon;
        if (slowest_alpha < epsilon) slowest_alpha = epsilon;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "kama({d}, {d:.4}, {d:.4}{s})", .{
            params.efficiency_ratio_length, fastest_alpha, slowest_alpha, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        return initInternal(allocator, params.efficiency_ratio_length, fastest_alpha, slowest_alpha, params.bar_component, params.quote_component, params.trade_component, mnemonic_buf, mnemonic_len);
    }

    fn initInternal(
        allocator: std.mem.Allocator,
        efficiency_ratio_length: u32,
        fastest_alpha: f64,
        slowest_alpha: f64,
        bc_opt: ?bar_component.BarComponent,
        qc_opt: ?quote_component.QuoteComponent,
        tc_opt: ?trade_component.TradeComponent,
        mnemonic_buf: [128]u8,
        mnemonic_len: usize,
    ) !KaufmanAdaptiveMovingAverage {
        const buf_len = efficiency_ratio_length + 1;
        const window = try allocator.alloc(f64, buf_len);
        @memset(window, 0);
        const absolute_delta = try allocator.alloc(f64, buf_len);
        @memset(absolute_delta, 0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Kaufman adaptive moving average ",
                bc_opt,
                qc_opt,
                tc_opt,
            ),
            .efficiency_ratio_length = efficiency_ratio_length,
            .window = window,
            .absolute_delta = absolute_delta,
            .absolute_delta_sum = 0,
            .alpha_fastest = fastest_alpha,
            .alpha_slowest = slowest_alpha,
            .alpha_diff = fastest_alpha - slowest_alpha,
            .value = math.nan(f64),
            .efficiency_ratio = math.nan(f64),
            .window_count = 0,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn deinit(self: *KaufmanAdaptiveMovingAverage) void {
        self.allocator.free(self.window);
        self.allocator.free(self.absolute_delta);
    }

    pub fn fixSlices(self: *KaufmanAdaptiveMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn update(self: *KaufmanAdaptiveMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        const epsilon = 0.00000001;
        const er_len = self.efficiency_ratio_length;

        if (self.primed) {
            var temp = @abs(sample - self.window[er_len]);
            self.absolute_delta_sum += temp - self.absolute_delta[1];

            // Shift window and absolute_delta left by 1.
            for (0..er_len) |i| {
                const j = i + 1;
                self.window[i] = self.window[j];
                self.absolute_delta[i] = self.absolute_delta[j];
            }

            self.window[er_len] = sample;
            self.absolute_delta[er_len] = temp;
            const delta = @abs(sample - self.window[0]);

            if (self.absolute_delta_sum <= delta or self.absolute_delta_sum < epsilon) {
                temp = 1.0;
            } else {
                temp = delta / self.absolute_delta_sum;
            }

            self.efficiency_ratio = temp;
            temp = self.alpha_slowest + temp * self.alpha_diff;
            self.value += (sample - self.value) * temp * temp;

            return self.value;
        } else {
            self.window[self.window_count] = sample;
            if (self.window_count > 0) {
                const temp = @abs(sample - self.window[self.window_count - 1]);
                self.absolute_delta[self.window_count] = temp;
                self.absolute_delta_sum += temp;
            }

            if (er_len == self.window_count) {
                self.primed = true;
                const delta = @abs(sample - self.window[0]);

                var temp: f64 = undefined;
                if (self.absolute_delta_sum <= delta or self.absolute_delta_sum < epsilon) {
                    temp = 1.0;
                } else {
                    temp = delta / self.absolute_delta_sum;
                }

                self.efficiency_ratio = temp;
                temp = self.alpha_slowest + temp * self.alpha_diff;
                self.value = self.window[er_len - 1];
                self.value += (sample - self.value) * temp * temp;

                return self.value;
            } else {
                self.window_count += 1;
            }
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const KaufmanAdaptiveMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const KaufmanAdaptiveMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .kaufman_adaptive_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *KaufmanAdaptiveMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *KaufmanAdaptiveMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *KaufmanAdaptiveMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *KaufmanAdaptiveMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *KaufmanAdaptiveMovingAverage) indicator_mod.Indicator {
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
        const self: *KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *KaufmanAdaptiveMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidEfficiencyRatioLength,
        InvalidFastestLength,
        InvalidSlowestLength,
        InvalidFastestSmoothingFactor,
        InvalidSlowestSmoothingFactor,
    } || std.mem.Allocator.Error;
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn createKamaLength(allocator: std.mem.Allocator, er_len: u32, fastest: u32, slowest: u32) !KaufmanAdaptiveMovingAverage {
    var kama = try KaufmanAdaptiveMovingAverage.initLength(allocator, .{
        .efficiency_ratio_length = er_len,
        .fastest_length = fastest,
        .slowest_length = slowest,
    });
    kama.fixSlices();
    return kama;
}

fn createKamaAlpha(allocator: std.mem.Allocator, er_len: u32, fastest_alpha: f64, slowest_alpha: f64) !KaufmanAdaptiveMovingAverage {
    var kama = try KaufmanAdaptiveMovingAverage.initSmoothingFactor(allocator, .{
        .efficiency_ratio_length = er_len,
        .fastest_smoothing_factor = fastest_alpha,
        .slowest_smoothing_factor = slowest_alpha,
    });
    kama.fixSlices();
    return kama;
}

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

test "kaufman adaptive moving average value" {
    const input = testdata.testInput();
    const expected = testdata.testExpected();

    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    for (0..10) |i| {
        const v = kama.update(input[i]);
        try testing.expect(math.isNan(v));
    }

    for (10..252) |i| {
        const v = kama.update(input[i]);
        try testing.expect(!math.isNan(v));
        try testing.expect(almostEqual(v, expected[i], 1e-8));
    }

    // NaN passthrough
    try testing.expect(math.isNan(kama.update(math.nan(f64))));
}

test "kaufman adaptive moving average efficiency ratio" {
    const input = testdata.testInput();
    const expected_er = testdata.testExpectedEr();

    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    for (0..10) |_| {
        _ = kama.update(input[0]);
    }

    // Re-create to get clean state
    kama.deinit();
    kama = try createKamaLength(testing.allocator, 10, 2, 30);

    for (0..10) |i| {
        _ = kama.update(input[i]);
    }

    for (10..252) |i| {
        _ = kama.update(input[i]);
        try testing.expect(almostEqual(kama.efficiency_ratio, expected_er[i], 1e-8));
    }
}

test "kaufman adaptive moving average is primed" {
    const input = testdata.testInput();

    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    try testing.expect(!kama.isPrimed());

    for (0..10) |_| {
        _ = kama.update(input[0]);
        try testing.expect(!kama.isPrimed());
    }

    // Re-create for clean state
    kama.deinit();
    kama = try createKamaLength(testing.allocator, 10, 2, 30);

    for (0..10) |i| {
        _ = kama.update(input[i]);
        try testing.expect(!kama.isPrimed());
    }

    _ = kama.update(input[10]);
    try testing.expect(kama.isPrimed());
}

test "kaufman adaptive moving average metadata length" {
    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    var m: Metadata = undefined;
    kama.getMetadata(&m);

    try testing.expectEqual(@import("../../core/identifier.zig").Identifier.kaufman_adaptive_moving_average, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("kama(10, 2, 30)", m.mnemonic);
}

test "kaufman adaptive moving average metadata alpha" {
    var kama = try createKamaAlpha(testing.allocator, 10, 0.666666666, 0.064516129);
    defer kama.deinit();

    var m: Metadata = undefined;
    kama.getMetadata(&m);

    try testing.expectEqualStrings("kama(10, 0.6667, 0.0645)", m.mnemonic);
}

test "kaufman adaptive moving average update scalar" {
    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    for (0..10) |_| {
        _ = kama.update(0);
    }

    const s = Scalar{ .time = 1617235200, .value = 3.0 };
    const out = kama.updateScalar(&s);
    const slice = out.slice();
    try testing.expectEqual(@as(usize, 1), slice.len);
    try testing.expectEqual(@as(i64, 1617235200), slice[0].scalar.time);
    try testing.expectEqual(@as(f64, 1.3333333333333328), slice[0].scalar.value);
}

test "kaufman adaptive moving average update bar" {
    var kama = try createKamaLength(testing.allocator, 10, 2, 30);
    defer kama.deinit();

    for (0..10) |_| {
        _ = kama.update(0);
    }

    const bar = Bar{ .time = 1617235200, .open = 0, .high = 0, .low = 0, .close = 3.0, .volume = 0 };
    const out = kama.updateBar(&bar);
    const slice = out.slice();
    try testing.expectEqual(@as(usize, 1), slice.len);
    try testing.expectEqual(@as(f64, 1.3333333333333328), slice[0].scalar.value);
}

test "kaufman adaptive moving average invalid params" {
    // ER length < 2
    try testing.expectError(error.InvalidEfficiencyRatioLength, KaufmanAdaptiveMovingAverage.initLength(testing.allocator, .{
        .efficiency_ratio_length = 1,
    }));

    // Fastest length < 2
    try testing.expectError(error.InvalidFastestLength, KaufmanAdaptiveMovingAverage.initLength(testing.allocator, .{
        .fastest_length = 1,
    }));

    // Slowest length < 2
    try testing.expectError(error.InvalidSlowestLength, KaufmanAdaptiveMovingAverage.initLength(testing.allocator, .{
        .slowest_length = 1,
    }));

    // Fastest alpha out of range
    try testing.expectError(error.InvalidFastestSmoothingFactor, KaufmanAdaptiveMovingAverage.initSmoothingFactor(testing.allocator, .{
        .fastest_smoothing_factor = -0.00000001,
    }));
    try testing.expectError(error.InvalidFastestSmoothingFactor, KaufmanAdaptiveMovingAverage.initSmoothingFactor(testing.allocator, .{
        .fastest_smoothing_factor = 1.00000001,
    }));

    // Slowest alpha out of range
    try testing.expectError(error.InvalidSlowestSmoothingFactor, KaufmanAdaptiveMovingAverage.initSmoothingFactor(testing.allocator, .{
        .slowest_smoothing_factor = -0.00000001,
    }));
    try testing.expectError(error.InvalidSlowestSmoothingFactor, KaufmanAdaptiveMovingAverage.initSmoothingFactor(testing.allocator, .{
        .slowest_smoothing_factor = 1.00000001,
    }));
}
