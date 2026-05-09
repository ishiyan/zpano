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

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the fractal adaptive moving average.
pub const FractalAdaptiveMovingAverageOutput = enum(u8) {
    value = 1,
    fdim = 2,
};

/// Parameters to create a FractalAdaptiveMovingAverage.
pub const Params = struct {
    /// Length (window size). Must be >= 2. Default is 16.
    /// If odd, it is rounded up to the next even number.
    length: i32 = 16,
    /// Slowest smoothing factor, alpha_s in [0, 1]. Default is 0.01.
    slowest_smoothing_factor: f64 = 0.01,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's Fractal Adaptive Moving Average (FRAMA).
///
/// An EMA with the smoothing factor alpha changed with each new sample based on
/// the estimated fractal dimension of the price series.
///
/// Two outputs: the FRAMA value and the estimated fractal dimension.
pub const FractalAdaptiveMovingAverage = struct {
    allocator: std.mem.Allocator,
    alpha_slowest: f64,
    scaling_factor: f64,
    fractal_dimension: f64,
    value: f64,
    length: usize,
    length_min_one: usize,
    half_length: usize,
    window_count: usize,
    window_high: []f64,
    window_low: []f64,
    primed: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [256]u8,
    description_len: usize,
    mnemonic_fdim_buf: [128]u8,
    mnemonic_fdim_len: usize,
    description_fdim_buf: [256]u8,
    description_fdim_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: Params) !FractalAdaptiveMovingAverage {
        if (params.length < 2) {
            return error.InvalidLength;
        }

        if (params.slowest_smoothing_factor < 0.0 or params.slowest_smoothing_factor > 1.0) {
            return error.InvalidSmoothingFactor;
        }

        var length_i: i32 = params.length;
        if (@mod(length_i, 2) != 0) {
            length_i += 1;
        }
        const length: usize = @intCast(length_i);

        const window_high = try allocator.alloc(f64, length);
        @memset(window_high, 0.0);
        errdefer allocator.free(window_high);

        const window_low = try allocator.alloc(f64, length);
        @memset(window_low, 0.0);

        // Resolve component defaults. FRAMA defaults to ClosePrice (not MedianPrice).
        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Build component triple mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            params.bar_component orelse bar_component.default_bar_component,
            params.quote_component orelse quote_component.default_quote_component,
            params.trade_component orelse trade_component.default_trade_component,
        );

        // Build mnemonics: frama(16, 0.010) or frama(16, 0.010, hl/2)
        var mnemonic_buf: [128]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "frama({d}, {d:.3}{s})", .{ length_i, params.slowest_smoothing_factor, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [256]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Fractal adaptive moving average {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_fdim_buf: [128]u8 = undefined;
        const mn_fdim = std.fmt.bufPrint(&mnemonic_fdim_buf, "framaDim({d}, {d:.3}{s})", .{ length_i, params.slowest_smoothing_factor, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_fdim_len = mn_fdim.len;

        var description_fdim_buf: [256]u8 = undefined;
        const desc_fdim = std.fmt.bufPrint(&description_fdim_buf, "Fractal adaptive moving average {s}", .{mn_fdim}) catch
            return error.MnemonicTooLong;
        const description_fdim_len = desc_fdim.len;

        return .{
            .allocator = allocator,
            .alpha_slowest = params.slowest_smoothing_factor,
            .scaling_factor = @log(params.slowest_smoothing_factor),
            .fractal_dimension = math.nan(f64),
            .value = math.nan(f64),
            .length = length,
            .length_min_one = length - 1,
            .half_length = length / 2,
            .window_count = 0,
            .window_high = window_high,
            .window_low = window_low,
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .mnemonic_fdim_buf = mnemonic_fdim_buf,
            .mnemonic_fdim_len = mnemonic_fdim_len,
            .description_fdim_buf = description_fdim_buf,
            .description_fdim_len = description_fdim_len,
        };
    }

    pub fn deinit(self: *FractalAdaptiveMovingAverage) void {
        self.allocator.free(self.window_high);
        self.allocator.free(self.window_low);
    }

    pub fn fixSlices(self: *FractalAdaptiveMovingAverage) void {
        _ = self;
    }

    /// Update the FRAMA given the next sample (value, high, low).
    pub fn update(self: *FractalAdaptiveMovingAverage, sample: f64, sample_high: f64, sample_low: f64) f64 {
        if (math.isNan(sample_high) or math.isNan(sample_low) or math.isNan(sample)) {
            return math.nan(f64);
        }

        if (self.primed) {
            // Shift windows left.
            std.mem.copyForwards(f64, self.window_high[0..self.length_min_one], self.window_high[1..self.length]);
            std.mem.copyForwards(f64, self.window_low[0..self.length_min_one], self.window_low[1..self.length]);

            self.window_high[self.length_min_one] = sample_high;
            self.window_low[self.length_min_one] = sample_low;

            self.fractal_dimension = self.estimateFractalDimension();
            const alpha = self.estimateAlpha();
            self.value += (sample - self.value) * alpha;

            return self.value;
        } else {
            self.window_high[self.window_count] = sample_high;
            self.window_low[self.window_count] = sample_low;

            self.window_count += 1;
            if (self.window_count == self.length_min_one) {
                self.value = sample;
            } else if (self.window_count == self.length) {
                self.fractal_dimension = self.estimateFractalDimension();
                const alpha = self.estimateAlpha();
                self.value += (sample - self.value) * alpha;
                self.primed = true;

                return self.value;
            }
        }

        return math.nan(f64);
    }

    fn estimateFractalDimension(self: *const FractalAdaptiveMovingAverage) f64 {
        var min_low_half: f64 = math.floatMax(f64);
        var max_high_half: f64 = math.floatMin(f64);

        for (0..self.half_length) |i| {
            const l = self.window_low[i];
            if (min_low_half > l) min_low_half = l;

            const h = self.window_high[i];
            if (max_high_half < h) max_high_half = h;
        }

        const range_n1 = max_high_half - min_low_half;
        var min_low_full = min_low_half;
        var max_high_full = max_high_half;
        min_low_half = math.floatMax(f64);
        max_high_half = math.floatMin(f64);

        for (0..self.half_length) |j| {
            const i = j + self.half_length;
            const l = self.window_low[i];

            if (min_low_full > l) min_low_full = l;
            if (min_low_half > l) min_low_half = l;

            const h = self.window_high[i];
            if (max_high_full < h) max_high_full = h;
            if (max_high_half < h) max_high_half = h;
        }

        const range_n2 = max_high_half - min_low_half;
        const range_n3 = max_high_full - min_low_full;

        const half_len_f: f64 = @floatFromInt(self.half_length);
        const len_f: f64 = @floatFromInt(self.length);

        const fdim = (@log((range_n1 + range_n2) / half_len_f) -
            @log(range_n3 / len_f)) * math.log2e;

        return @min(@max(fdim, 1.0), 2.0);
    }

    fn estimateAlpha(self: *const FractalAdaptiveMovingAverage) f64 {
        const alpha = @exp(self.scaling_factor * (self.fractal_dimension - 1.0));
        return @min(@max(alpha, self.alpha_slowest), 1.0);
    }

    pub fn isPrimed(self: *const FractalAdaptiveMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const FractalAdaptiveMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .fractal_adaptive_moving_average,
            self.mnemonic_buf[0..self.mnemonic_len],
            self.description_buf[0..self.description_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mnemonic_buf[0..self.mnemonic_len], .description = self.description_buf[0..self.description_len] },
                .{ .mnemonic = self.mnemonic_fdim_buf[0..self.mnemonic_fdim_len], .description = self.description_fdim_buf[0..self.description_fdim_len] },
            },
        );
    }

    pub fn updateScalar(self: *FractalAdaptiveMovingAverage, sample: *const Scalar) OutputArray {
        const v = sample.value;
        return self.updateEntity(sample.time, v, v, v);
    }

    pub fn updateBar(self: *FractalAdaptiveMovingAverage, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateEntity(sample.time, v, sample.high, sample.low);
    }

    pub fn updateQuote(self: *FractalAdaptiveMovingAverage, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateEntity(sample.time, v, sample.ask_price, sample.bid_price);
    }

    pub fn updateTrade(self: *FractalAdaptiveMovingAverage, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateEntity(sample.time, v, v, v);
    }

    fn updateEntity(self: *FractalAdaptiveMovingAverage, time: i64, sample: f64, sample_high: f64, sample_low: f64) OutputArray {
        const frama = self.update(sample, sample_high, sample_low);
        var fdim = self.fractal_dimension;
        if (math.isNan(frama)) {
            fdim = math.nan(f64);
        }

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = frama } });
        out.append(.{ .scalar = .{ .time = time, .value = fdim } });
        return out;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *FractalAdaptiveMovingAverage) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(FractalAdaptiveMovingAverage);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

// 252-entry mid-price input data from test_FRAMA.xls.
// 252-entry high price input data from test_FRAMA.xls.
// 252-entry low price input data from test_FRAMA.xls.
// Expected FRAMA values, 252 entries.
// Expected fractal dimension values, 252 entries.
test "FRAMA update value" {
    const tolerance = 1e-9;
    const l_primed = 15;

    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    for (0..l_primed) |i| {
        try testing.expect(math.isNan(frama.update(testdata.test_input_mid[i], testdata.test_input_high[i], testdata.test_input_low[i])));
    }

    for (l_primed..testdata.test_input_mid.len) |i| {
        const act = frama.update(testdata.test_input_mid[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(almostEqual(act, testdata.test_expected_frama[i], tolerance));
    }

    // NaN passthrough.
    try testing.expect(math.isNan(frama.update(math.nan(f64), math.nan(f64), math.nan(f64))));
}

test "FRAMA update fdim" {
    const tolerance = 1e-9;
    const l_primed = 15;

    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    for (0..l_primed) |i| {
        _ = frama.update(testdata.test_input_mid[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(math.isNan(frama.fractal_dimension));
    }

    for (l_primed..testdata.test_input_mid.len) |i| {
        _ = frama.update(testdata.test_input_mid[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(almostEqual(frama.fractal_dimension, testdata.test_expected_fdim[i], tolerance));
    }
}

test "FRAMA isPrimed" {
    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    try testing.expect(!frama.isPrimed());

    const l_primed = 15;
    for (0..l_primed) |i| {
        _ = frama.update(testdata.test_input_mid[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(!frama.isPrimed());
    }

    _ = frama.update(testdata.test_input_mid[l_primed], testdata.test_input_high[l_primed], testdata.test_input_low[l_primed]);
    try testing.expect(frama.isPrimed());
}

test "FRAMA metadata" {
    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    var meta: Metadata = undefined;
    frama.getMetadata(&meta);

    try testing.expectEqual(Identifier.fractal_adaptive_moving_average, meta.identifier);
    try testing.expectEqualStrings("frama(16, 0.010)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqualStrings("frama(16, 0.010)", meta.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("framaDim(16, 0.010)", meta.outputs_buf[1].mnemonic);
}

test "FRAMA constructor" {
    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();

    // Odd length rounds up.
    var frama2 = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 17, .slowest_smoothing_factor = 0.01 });
    defer frama2.deinit();
    try testing.expectEqual(@as(usize, 18), frama2.length);

    try testing.expectError(error.InvalidLength, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 1, .slowest_smoothing_factor = 0.01 }));
    try testing.expectError(error.InvalidLength, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 0, .slowest_smoothing_factor = 0.01 }));
    try testing.expectError(error.InvalidLength, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = -1, .slowest_smoothing_factor = 0.01 }));
    try testing.expectError(error.InvalidSmoothingFactor, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = -0.01 }));
    try testing.expectError(error.InvalidSmoothingFactor, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 1.01 }));
}

test "FRAMA updateEntity" {
    const tolerance = 1e-9;

    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    // Prime with 15 zeros.
    for (0..15) |_| {
        _ = frama.update(0.0, 0.0, 0.0);
    }

    // 16th update with value 3.0.
    const inp: f64 = 3.0;
    const s = Scalar{ .time = 1, .value = inp };
    const out = frama.updateScalar(&s);
    const outputs = out.slice();
    try testing.expectEqual(@as(usize, 2), outputs.len);
    try testing.expect(almostEqual(outputs[0].scalar.value, 2.999999999999997, tolerance));
    try testing.expect(almostEqual(outputs[1].scalar.value, 1.0000000000000002, tolerance));
}

test "FRAMA updateBar" {
    const tolerance = 1e-9;

    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    for (0..testdata.test_input_mid.len) |i| {
        const b = Bar{
            .time = @intCast(i),
            .open = 0,
            .high = testdata.test_input_high[i],
            .low = testdata.test_input_low[i],
            .close = testdata.test_input_mid[i],
            .volume = 0,
        };
        const out = frama.updateBar(&b);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 2), outputs.len);

        if (i < 15) {
            try testing.expect(math.isNan(outputs[0].scalar.value));
            try testing.expect(math.isNan(outputs[1].scalar.value));
        } else {
            try testing.expect(almostEqual(outputs[0].scalar.value, testdata.test_expected_frama[i], tolerance));
            try testing.expect(almostEqual(outputs[1].scalar.value, testdata.test_expected_fdim[i], tolerance));
        }
    }
}
