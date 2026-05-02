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

/// Enumerates the outputs of the Jurik Relative Trend Strength Index indicator.
pub const JurikRelativeTrendStrengthIndexOutput = enum(u8) {
    /// The RSX value.
    rsx = 1,
};

/// Parameters for the Jurik Relative Trend Strength Index.
pub const JurikRelativeTrendStrengthIndexParams = struct {
    /// Length of the indicator. Must be >= 2.
    length: u32 = 14,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Jurik Relative Trend Strength Index (RSX), see http://jurikres.com/.
/// RSX is a noise-free version of RSI based on triple-smoothed EMA of momentum and absolute momentum.
pub const JurikRelativeTrendStrengthIndex = struct {
    line: LineIndicator,
    primed: bool,
    param_len: u32,

    // State variables.
    f0: i32,
    f88: i32,
    f90: i32,

    f8: f64,
    f10: f64,
    f18: f64,
    f20: f64,
    f28: f64,
    f30: f64,
    f38: f64,
    f40: f64,
    f48: f64,
    f50: f64,
    f58: f64,
    f60: f64,
    f68: f64,
    f70: f64,
    f78: f64,
    f80: f64,

    mnemonic_buf: [96]u8,
    mnemonic_len: usize,

    pub fn init(params: JurikRelativeTrendStrengthIndexParams) !JurikRelativeTrendStrengthIndex {
        const length = params.length;

        if (length < 2) return error.InvalidLength;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "rsx({d}{s})", .{
            length, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Jurik relative trend strength index ",
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .primed = false,
            .param_len = length,
            .f0 = 0,
            .f88 = 0,
            .f90 = 0,
            .f8 = 0,
            .f10 = 0,
            .f18 = 0,
            .f20 = 0,
            .f28 = 0,
            .f30 = 0,
            .f38 = 0,
            .f40 = 0,
            .f48 = 0,
            .f50 = 0,
            .f58 = 0,
            .f60 = 0,
            .f68 = 0,
            .f70 = 0,
            .f78 = 0,
            .f80 = 0,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *JurikRelativeTrendStrengthIndex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn update(self: *JurikRelativeTrendStrengthIndex, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        const hundred: f64 = 100.0;
        const fifty: f64 = 50.0;
        const one_five: f64 = 1.5;
        const half: f64 = 0.5;
        const min_len: i32 = 5;
        const eps: f64 = 1e-10;

        const length: i32 = @intCast(self.param_len);

        if (self.f90 == 0) {
            // First call: initialize.
            self.f90 = 1;
            self.f0 = 0;

            if (length - 1 >= min_len) {
                self.f88 = length - 1;
            } else {
                self.f88 = min_len;
            }

            self.f8 = hundred * sample;
            self.f18 = 3.0 / @as(f64, @floatFromInt(length + 2));
            self.f20 = 1.0 - self.f18;
        } else {
            if (self.f88 <= self.f90) {
                self.f90 = self.f88 + 1;
            } else {
                self.f90 += 1;
            }

            self.f10 = self.f8;
            self.f8 = hundred * sample;
            const v8 = self.f8 - self.f10;

            self.f28 = self.f20 * self.f28 + self.f18 * v8;
            self.f30 = self.f18 * self.f28 + self.f20 * self.f30;
            const vC = self.f28 * one_five - self.f30 * half;

            self.f38 = self.f20 * self.f38 + self.f18 * vC;
            self.f40 = self.f18 * self.f38 + self.f20 * self.f40;
            const v10 = self.f38 * one_five - self.f40 * half;

            self.f48 = self.f20 * self.f48 + self.f18 * v10;
            self.f50 = self.f18 * self.f48 + self.f20 * self.f50;
            const v14 = self.f48 * one_five - self.f50 * half;

            self.f58 = self.f20 * self.f58 + self.f18 * @abs(v8);
            self.f60 = self.f18 * self.f58 + self.f20 * self.f60;
            const v18 = self.f58 * one_five - self.f60 * half;

            self.f68 = self.f20 * self.f68 + self.f18 * v18;
            self.f70 = self.f18 * self.f68 + self.f20 * self.f70;
            const v1C = self.f68 * one_five - self.f70 * half;

            self.f78 = self.f20 * self.f78 + self.f18 * v1C;
            self.f80 = self.f18 * self.f78 + self.f20 * self.f80;
            const v20 = self.f78 * one_five - self.f80 * half;

            if (self.f88 >= self.f90 and self.f8 != self.f10) {
                self.f0 = 1;
            }

            if (self.f88 == self.f90 and self.f0 == 0) {
                self.f90 = 0;
            }

            if (self.f88 < self.f90 and v20 > eps) {
                var v4 = (v14 / v20 + 1.0) * fifty;
                if (v4 > hundred) {
                    v4 = hundred;
                }
                if (v4 < 0.0) {
                    v4 = 0.0;
                }

                self.primed = true;
                return v4;
            }
        }

        // During warmup or when denominator is too small.
        if (self.f88 < self.f90) {
            self.primed = true;
        }

        if (!self.primed) {
            return math.nan(f64);
        }

        return fifty;
    }

    pub fn isPrimed(self: *const JurikRelativeTrendStrengthIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikRelativeTrendStrengthIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .jurik_relative_trend_strength_index,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *JurikRelativeTrendStrengthIndex, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *JurikRelativeTrendStrengthIndex, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *JurikRelativeTrendStrengthIndex, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *JurikRelativeTrendStrengthIndex, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *JurikRelativeTrendStrengthIndex) indicator_mod.Indicator {
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
        const self: *JurikRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikRelativeTrendStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidLength,
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

fn runRsxTest(length: u32, expected: [252]f64) !void {
    var rsx = JurikRelativeTrendStrengthIndex.init(.{ .length = length }) catch unreachable;
    rsx.fixSlices();
    const input = testdata.testInput();
    const eps = 1e-13;

    for (0..252) |i| {
        const v = rsx.update(input[i]);
        if (math.isNan(expected[i])) {
            try testing.expect(math.isNan(v));
        } else {
            try testing.expect(!math.isNan(v));
            if (!almostEqual(v, expected[i], eps)) {
                std.debug.print("FAIL [{d}] len={d}: expected {d}, got {d}, diff {d}\n", .{ i, length, expected[i], v, @abs(v - expected[i]) });
                return error.TestUnexpectedResult;
            }
        }
    }

    // NaN passthrough
    try testing.expect(math.isNan(rsx.update(math.nan(f64))));
}

test "rsx length 2" {
    try runRsxTest(2, testdata.expectedLength2());
}
test "rsx length 3" {
    try runRsxTest(3, testdata.expectedLength3());
}
test "rsx length 4" {
    try runRsxTest(4, testdata.expectedLength4());
}
test "rsx length 5" {
    try runRsxTest(5, testdata.expectedLength5());
}
test "rsx length 6" {
    try runRsxTest(6, testdata.expectedLength6());
}
test "rsx length 7" {
    try runRsxTest(7, testdata.expectedLength7());
}
test "rsx length 8" {
    try runRsxTest(8, testdata.expectedLength8());
}
test "rsx length 9" {
    try runRsxTest(9, testdata.expectedLength9());
}
test "rsx length 10" {
    try runRsxTest(10, testdata.expectedLength10());
}
test "rsx length 11" {
    try runRsxTest(11, testdata.expectedLength11());
}
test "rsx length 12" {
    try runRsxTest(12, testdata.expectedLength12());
}
test "rsx length 13" {
    try runRsxTest(13, testdata.expectedLength13());
}
test "rsx length 14" {
    try runRsxTest(14, testdata.expectedLength14());
}
test "rsx length 15" {
    try runRsxTest(15, testdata.expectedLength15());
}
