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

/// Enumerates the outputs of the Jurik Turning Point Oscillator indicator.
pub const JurikTurningPointOscillatorOutput = enum(u8) {
    /// The turning point oscillator value.
    value = 1,
};

/// Parameters for the Jurik Turning Point Oscillator.
pub const JurikTurningPointOscillatorParams = struct {
    /// Length controls the lookback window for the Spearman rank correlation. Must be >= 2.
    length: u32 = 14,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Jurik Turning Point Oscillator (JTPO) computes Spearman rank correlation between
/// price ranks and time positions. Output is in [-1, +1].
pub const JurikTurningPointOscillator = struct {
    line: LineIndicator,
    primed: bool,
    length: u32,
    buffer: [256]f64,
    buf_idx: u32,
    count: u32,
    f18: f64,
    mid: f64,

    mnemonic_buf: [96]u8,
    mnemonic_len: usize,

    pub fn init(params: JurikTurningPointOscillatorParams) !JurikTurningPointOscillator {
        const length = params.length;

        if (length < 2) return error.InvalidLength;
        if (length > 256) return error.InvalidLength;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "jtpo({d}{s})", .{
            length, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        const n: f64 = @floatFromInt(length);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Jurik turning point oscillator ",
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .primed = false,
            .length = length,
            .buffer = [_]f64{0} ** 256,
            .buf_idx = 0,
            .count = 0,
            .f18 = 12.0 / (n * (n - 1.0) * (n + 1.0)),
            .mid = (n + 1.0) / 2.0,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *JurikTurningPointOscillator) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn update(self: *JurikTurningPointOscillator, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        const length = self.length;

        self.buffer[self.buf_idx] = sample;
        self.buf_idx = (self.buf_idx + 1) % length;
        self.count += 1;

        if (self.count < length) {
            return math.nan(f64);
        }

        // Extract window in chronological order.
        var window: [256]f64 = undefined;
        for (0..length) |i| {
            window[i] = self.buffer[(self.buf_idx + @as(u32, @intCast(i))) % length];
        }

        // Check if all values are identical.
        var all_same = true;
        for (1..length) |i| {
            if (window[i] != window[0]) {
                all_same = false;
                break;
            }
        }

        if (all_same) {
            if (!self.primed) {
                self.primed = true;
            }
            return math.nan(f64);
        }

        // Build indices sorted by price (stable sort).
        var indices: [256]u32 = undefined;
        for (0..length) |i| {
            indices[i] = @intCast(i);
        }

        // Stable sort by price using insertion sort (stable, O(n^2) but n<=256).
        for (1..length) |i_usize| {
            const i: u32 = @intCast(i_usize);
            const key_idx = indices[i];
            const key_price = window[key_idx];
            var j: i32 = @as(i32, @intCast(i)) - 1;
            while (j >= 0) {
                const j_usize: usize = @intCast(@as(u32, @intCast(j)));
                if (window[indices[j_usize]] <= key_price) break;
                indices[j_usize + 1] = indices[j_usize];
                j -= 1;
            }
            indices[@intCast(@as(u32, @intCast(j + 1)))] = key_idx;
        }

        // arr2[i] = original time position (1-based) of the i-th sorted element.
        var arr2: [256]f64 = undefined;
        for (0..length) |i| {
            arr2[i] = @as(f64, @floatFromInt(indices[i])) + 1.0;
        }

        // Assign fractional ranks for ties.
        var arr3: [256]f64 = undefined;
        var i: u32 = 0;
        while (i < length) {
            var j: u32 = i;
            while (j < length - 1 and window[indices[j + 1]] == window[indices[j]]) {
                j += 1;
            }
            const avg_rank = (@as(f64, @floatFromInt(i + 1)) + @as(f64, @floatFromInt(j + 1))) / 2.0;
            for (i..j + 1) |k| {
                arr3[k] = avg_rank;
            }
            i = j + 1;
        }

        // Compute correlation sum.
        var corr_sum: f64 = 0;
        for (0..length) |ci| {
            corr_sum += (arr3[ci] - self.mid) * (arr2[ci] - self.mid);
        }

        if (!self.primed) {
            self.primed = true;
        }

        return self.f18 * corr_sum;
    }

    pub fn isPrimed(self: *const JurikTurningPointOscillator) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikTurningPointOscillator, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .jurik_turning_point_oscillator,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *JurikTurningPointOscillator, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *JurikTurningPointOscillator, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *JurikTurningPointOscillator, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *JurikTurningPointOscillator, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *JurikTurningPointOscillator) indicator_mod.Indicator {
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
        const self: *JurikTurningPointOscillator = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikTurningPointOscillator = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikTurningPointOscillator = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikTurningPointOscillator = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikTurningPointOscillator = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikTurningPointOscillator = @ptrCast(@alignCast(ptr));
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

fn runTpoTest(length: u32, expected: [252]f64) !void {
    var tpo = JurikTurningPointOscillator.init(.{ .length = length }) catch unreachable;
    tpo.fixSlices();
    const input = testdata.testInput();
    const eps = 1e-13;

    for (0..252) |i| {
        const v = tpo.update(input[i]);
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

    // NaN passthrough
    try testing.expect(math.isNan(tpo.update(math.nan(f64))));
}

test "jtpo length 5" {
    try runTpoTest(5, testdata.expectedLen5());
}
test "jtpo length 7" {
    try runTpoTest(7, testdata.expectedLen7());
}
test "jtpo length 10" {
    try runTpoTest(10, testdata.expectedLen10());
}
test "jtpo length 14" {
    try runTpoTest(14, testdata.expectedLen14());
}
test "jtpo length 20" {
    try runTpoTest(20, testdata.expectedLen20());
}
test "jtpo length 28" {
    try runTpoTest(28, testdata.expectedLen28());
}
test "jtpo length 40" {
    try runTpoTest(40, testdata.expectedLen40());
}
test "jtpo length 60" {
    try runTpoTest(60, testdata.expectedLen60());
}
test "jtpo length 80" {
    try runTpoTest(80, testdata.expectedLen80());
}
