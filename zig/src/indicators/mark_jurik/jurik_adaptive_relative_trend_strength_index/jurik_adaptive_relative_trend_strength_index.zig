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

/// Enumerates the outputs of the Jurik Adaptive Relative Trend Strength Index indicator.
pub const JurikAdaptiveRelativeTrendStrengthIndexOutput = enum(u8) {
    /// The JARSX value in [0, 100].
    value = 1,
};

/// Parameters for the Jurik Adaptive Relative Trend Strength Index.
pub const JurikAdaptiveRelativeTrendStrengthIndexParams = struct {
    /// Minimum adaptive RSX length. Must be >= 2.
    lo_length: u32 = 5,
    /// Maximum adaptive RSX length. Must be >= lo_length.
    hi_length: u32 = 30,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Jurik Adaptive Relative Trend Strength Index (JARSX).
/// Combines adaptive length selection (volatility regime detection) with
/// the RSX core (triple-cascaded lag-reduced EMA oscillator). Output in [0, 100].
pub const JurikAdaptiveRelativeTrendStrengthIndex = struct {
    line: LineIndicator,
    primed: bool,
    lo_length: u32,
    hi_length: u32,
    eps: f64,

    bar_count: u32,
    previous_price: f64,

    // Rolling buffers for adaptive length.
    long_buffer: [100]f64,
    long_index: u32,
    long_sum: f64,
    long_count: u32,
    short_buffer: [10]f64,
    short_index: u32,
    short_sum: f64,
    short_count: u32,

    // RSX core state.
    kg: f64,
    c: f64,
    warmup: u32,
    // Signal path (3 cascaded stages).
    sig1a: f64,
    sig1b: f64,
    sig2a: f64,
    sig2b: f64,
    sig3a: f64,
    sig3b: f64,
    // Denominator path (3 cascaded stages).
    den1a: f64,
    den1b: f64,
    den2a: f64,
    den2b: f64,
    den3a: f64,
    den3b: f64,

    mnemonic_buf: [96]u8,
    mnemonic_len: usize,

    pub fn init(params: JurikAdaptiveRelativeTrendStrengthIndexParams) !JurikAdaptiveRelativeTrendStrengthIndex {
        const lo_length = params.lo_length;
        const hi_length = params.hi_length;

        if (lo_length < 2) return error.InvalidLoLength;
        if (hi_length < lo_length) return error.InvalidHiLength;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "jarsx({d}, {d}{s})", .{
            lo_length, hi_length, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Jurik adaptive relative trend strength index ",
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .primed = false,
            .lo_length = lo_length,
            .hi_length = hi_length,
            .eps = 0.001,
            .bar_count = 0,
            .previous_price = 0,
            .long_buffer = [_]f64{0} ** 100,
            .long_index = 0,
            .long_sum = 0,
            .long_count = 0,
            .short_buffer = [_]f64{0} ** 10,
            .short_index = 0,
            .short_sum = 0,
            .short_count = 0,
            .kg = 0,
            .c = 0,
            .warmup = 0,
            .sig1a = 0,
            .sig1b = 0,
            .sig2a = 0,
            .sig2b = 0,
            .sig3a = 0,
            .sig3b = 0,
            .den1a = 0,
            .den1b = 0,
            .den2a = 0,
            .den2b = 0,
            .den3a = 0,
            .den3b = 0,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *JurikAdaptiveRelativeTrendStrengthIndex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn update(self: *JurikAdaptiveRelativeTrendStrengthIndex, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        const bar = self.bar_count;
        self.bar_count += 1;

        if (bar == 0) {
            self.previous_price = sample;

            // First bar: add 0 to both buffers.
            self.long_buffer[0] = 0.0;
            self.long_sum = 0.0;
            self.long_count = 1;
            self.short_buffer[0] = 0.0;
            self.short_sum = 0.0;
            self.short_count = 1;

            // Compute adaptive length from bar 0.
            const avg1: f64 = 0.0;
            const avg2: f64 = 0.0;
            const value2 = @log(self.eps + avg1) - @log(self.eps + avg2);
            const value3 = value2 / (1.0 + @abs(value2));
            const adaptive_length = @as(f64, @floatFromInt(self.lo_length)) +
                @as(f64, @floatFromInt(self.hi_length - self.lo_length)) * (1.0 + value3) / 2.0;
            var length: u32 = @intFromFloat(adaptive_length);
            if (length < 2) length = 2;

            self.kg = 3.0 / @as(f64, @floatFromInt(length + 2));
            self.c = 1.0 - self.kg;
            self.warmup = length - 1;
            if (self.warmup < 5) self.warmup = 5;

            return math.nan(f64);
        }

        // Bars 1+
        const old_price = self.previous_price;
        self.previous_price = sample;
        const value1 = @abs(sample - old_price);

        // Update long rolling buffer.
        if (self.long_count < 100) {
            self.long_buffer[self.long_count] = value1;
            self.long_sum += value1;
            self.long_count += 1;
        } else {
            self.long_sum -= self.long_buffer[self.long_index];
            self.long_buffer[self.long_index] = value1;
            self.long_sum += value1;
            self.long_index = (self.long_index + 1) % 100;
        }

        // Update short rolling buffer.
        if (self.short_count < 10) {
            self.short_buffer[self.short_count] = value1;
            self.short_sum += value1;
            self.short_count += 1;
        } else {
            self.short_sum -= self.short_buffer[self.short_index];
            self.short_buffer[self.short_index] = value1;
            self.short_sum += value1;
            self.short_index = (self.short_index + 1) % 10;
        }

        // RSX core computation.
        const mom = 100.0 * (sample - old_price);
        const abs_mom = @abs(mom);

        const kg = self.kg;
        const c = self.c;

        // Signal path — Stage 1.
        self.sig1a = c * self.sig1a + kg * mom;
        self.sig1b = kg * self.sig1a + c * self.sig1b;
        const s1 = 1.5 * self.sig1a - 0.5 * self.sig1b;

        // Signal path — Stage 2.
        self.sig2a = c * self.sig2a + kg * s1;
        self.sig2b = kg * self.sig2a + c * self.sig2b;
        const s2 = 1.5 * self.sig2a - 0.5 * self.sig2b;

        // Signal path — Stage 3.
        self.sig3a = c * self.sig3a + kg * s2;
        self.sig3b = kg * self.sig3a + c * self.sig3b;
        const numerator = 1.5 * self.sig3a - 0.5 * self.sig3b;

        // Denominator path — Stage 1.
        self.den1a = c * self.den1a + kg * abs_mom;
        self.den1b = kg * self.den1a + c * self.den1b;
        const d1 = 1.5 * self.den1a - 0.5 * self.den1b;

        // Denominator path — Stage 2.
        self.den2a = c * self.den2a + kg * d1;
        self.den2b = kg * self.den2a + c * self.den2b;
        const d2 = 1.5 * self.den2a - 0.5 * self.den2b;

        // Denominator path — Stage 3.
        self.den3a = c * self.den3a + kg * d2;
        self.den3b = kg * self.den3a + c * self.den3b;
        const denominator = 1.5 * self.den3a - 0.5 * self.den3b;

        // Output after warmup.
        if (bar >= self.warmup) {
            self.primed = true;

            var value: f64 = undefined;
            if (denominator != 0.0) {
                value = (numerator / denominator + 1.0) * 50.0;
            } else {
                value = 50.0;
            }

            return @max(0.0, @min(100.0, value));
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const JurikAdaptiveRelativeTrendStrengthIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikAdaptiveRelativeTrendStrengthIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .jurik_adaptive_relative_trend_strength_index,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *JurikAdaptiveRelativeTrendStrengthIndex, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *JurikAdaptiveRelativeTrendStrengthIndex, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *JurikAdaptiveRelativeTrendStrengthIndex, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *JurikAdaptiveRelativeTrendStrengthIndex, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *JurikAdaptiveRelativeTrendStrengthIndex) indicator_mod.Indicator {
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
        const self: *JurikAdaptiveRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikAdaptiveRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikAdaptiveRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikAdaptiveRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikAdaptiveRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikAdaptiveRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidLoLength,
        InvalidHiLength,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

fn runJarsxTest(lo_length: u32, hi_length: u32, expected: [252]f64) !void {
    var ind = JurikAdaptiveRelativeTrendStrengthIndex.init(.{ .lo_length = lo_length, .hi_length = hi_length }) catch unreachable;
    ind.fixSlices();
    const input = testdata.testInput();
    const eps = 1e-13;

    for (0..252) |i| {
        const v = ind.update(input[i]);
        if (math.isNan(expected[i])) {
            try testing.expect(math.isNan(v));
        } else {
            try testing.expect(!math.isNan(v));
            if (!almostEqual(v, expected[i], eps)) {
                std.debug.print("FAIL [{d}] lo={d} hi={d}: expected {d}, got {d}, diff {d}\n", .{ i, lo_length, hi_length, expected[i], v, @abs(v - expected[i]) });
                return error.TestUnexpectedResult;
            }
        }
    }

    try testing.expect(math.isNan(ind.update(math.nan(f64))));
}

test "jarsx lo=2 hi=15" {
    try runJarsxTest(2, 15, testdata.expectedLo2Hi15());
}
test "jarsx lo=2 hi=30" {
    try runJarsxTest(2, 30, testdata.expectedLo2Hi30());
}
test "jarsx lo=2 hi=60" {
    try runJarsxTest(2, 60, testdata.expectedLo2Hi60());
}
test "jarsx lo=5 hi=15" {
    try runJarsxTest(5, 15, testdata.expectedLo5Hi15());
}
test "jarsx lo=5 hi=30" {
    try runJarsxTest(5, 30, testdata.expectedLo5Hi30());
}
test "jarsx lo=5 hi=60" {
    try runJarsxTest(5, 60, testdata.expectedLo5Hi60());
}
test "jarsx lo=10 hi=15" {
    try runJarsxTest(10, 15, testdata.expectedLo10Hi15());
}
test "jarsx lo=10 hi=30" {
    try runJarsxTest(10, 30, testdata.expectedLo10Hi30());
}
test "jarsx lo=10 hi=60" {
    try runJarsxTest(10, 60, testdata.expectedLo10Hi60());
}
