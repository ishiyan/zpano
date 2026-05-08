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

const jma_mod = @import("../jurik_moving_average/jurik_moving_average.zig");
const JurikMovingAverage = jma_mod.JurikMovingAverage;

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Jurik Commodity Channel Index indicator.
pub const JurikCommodityChannelIndexOutput = enum(u8) {
    /// The JCCX value.
    value = 1,
};

/// Parameters for the Jurik Commodity Channel Index.
pub const JurikCommodityChannelIndexParams = struct {
    /// Length for the slow JMA. Must be >= 2.
    length: u32 = 20,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

const MAX_DIFF_BUF = 1024;

/// Jurik Commodity Channel Index (JCCX).
/// Uses fast JMA(4) and slow JMA(length), normalizes their difference by 1.5× MAD.
pub const JurikCommodityChannelIndex = struct {
    line: LineIndicator,
    primed: bool,
    fast_jma: JurikMovingAverage,
    slow_jma: JurikMovingAverage,
    diff_buffer: [MAX_DIFF_BUF]f64,
    diff_count: u32,
    diff_start: u32,
    diff_buf_size: u32,

    mnemonic_buf: [96]u8,
    mnemonic_len: usize,

    pub fn init(params: JurikCommodityChannelIndexParams) !JurikCommodityChannelIndex {
        const length = params.length;

        if (length < 2) return error.InvalidLength;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "jccx({d}{s})", .{
            length, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        const fast_jma = jma_mod.JurikMovingAverage.init(.{ .length = 4, .phase = 0 }) catch return error.InvalidLength;
        const slow_jma = jma_mod.JurikMovingAverage.init(.{ .length = length, .phase = 0 }) catch return error.InvalidLength;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Jurik commodity channel index ",
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .primed = false,
            .fast_jma = fast_jma,
            .slow_jma = slow_jma,
            .diff_buffer = [_]f64{0} ** MAX_DIFF_BUF,
            .diff_count = 0,
            .diff_start = 0,
            .diff_buf_size = 3 * length,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *JurikCommodityChannelIndex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn update(self: *JurikCommodityChannelIndex, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        const fast_val = self.fast_jma.update(sample);
        const slow_val = self.slow_jma.update(sample);

        if (math.isNan(fast_val) or math.isNan(slow_val)) {
            return math.nan(f64);
        }

        const diff = fast_val - slow_val;

        // Append to circular diff buffer.
        if (self.diff_count < self.diff_buf_size) {
            self.diff_buffer[self.diff_count] = diff;
            self.diff_count += 1;
        } else {
            self.diff_buffer[self.diff_start] = diff;
            self.diff_start = (self.diff_start + 1) % self.diff_buf_size;
        }

        self.primed = true;

        // Compute MAD.
        const n = self.diff_count;
        var mad: f64 = 0;

        for (0..n) |i| {
            const idx = (self.diff_start + @as(u32, @intCast(i))) % self.diff_buf_size;
            mad += @abs(self.diff_buffer[idx]);
        }

        mad /= @as(f64, @floatFromInt(n));

        if (mad < 0.00001) {
            return 0.0;
        }

        return diff / (1.5 * mad);
    }

    pub fn isPrimed(self: *const JurikCommodityChannelIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikCommodityChannelIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .jurik_commodity_channel_index,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *JurikCommodityChannelIndex, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *JurikCommodityChannelIndex, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *JurikCommodityChannelIndex, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *JurikCommodityChannelIndex, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *JurikCommodityChannelIndex) indicator_mod.Indicator {
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
        const self: *JurikCommodityChannelIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikCommodityChannelIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikCommodityChannelIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikCommodityChannelIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikCommodityChannelIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikCommodityChannelIndex = @ptrCast(@alignCast(ptr));
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

fn runJccxTest(length: u32, expected: [252]f64) !void {
    var ind = JurikCommodityChannelIndex.init(.{ .length = length }) catch unreachable;
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
                std.debug.print("FAIL [{d}] length={d}: expected {d}, got {d}, diff {d}\n", .{ i, length, expected[i], v, @abs(v - expected[i]) });
                return error.TestUnexpectedResult;
            }
        }
    }

    try testing.expect(math.isNan(ind.update(math.nan(f64))));
}

test "jccx length 10" {
    try runJccxTest(10, testdata.expectedLen10());
}
test "jccx length 14" {
    try runJccxTest(14, testdata.expectedLen14());
}
test "jccx length 20" {
    try runJccxTest(20, testdata.expectedLen20());
}
test "jccx length 30" {
    try runJccxTest(30, testdata.expectedLen30());
}
test "jccx length 40" {
    try runJccxTest(40, testdata.expectedLen40());
}
test "jccx length 50" {
    try runJccxTest(50, testdata.expectedLen50());
}
test "jccx length 60" {
    try runJccxTest(60, testdata.expectedLen60());
}
test "jccx length 80" {
    try runJccxTest(80, testdata.expectedLen80());
}
test "jccx length 100" {
    try runJccxTest(100, testdata.expectedLen100());
}
