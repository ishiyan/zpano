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

/// Enumerates the outputs of the hurst difference indicator.
pub const HurstDifferenceOutput = enum(u8) {
    /// The first difference of FGDI.
    hurst_diff = 1,
    /// The raw FGDI value.
    fgdi = 2,
};

/// Parameters to create an instance of the hurst difference indicator.
pub const HurstDifferenceParams = struct {
    /// The lookback period N. Must be > 1.
    period: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Hurst Difference (first difference of the corrected FGDI).
///
/// Positive values indicate rising volatility (potential trade entry);
/// negative values indicate declining volatility.
///
/// The FGDI is computed using the corrected FGDI formula with (period-1)
/// segments and denominator ln(2*(period-1)).
///
/// The indicator is not primed during the first `period` updates.
/// The hurst_diff output requires one additional update beyond FGDI priming.
pub const HurstDifference = struct {
    line: LineIndicator,
    window: []f64,
    period: usize,
    n_minus_1: usize,
    window_count: usize,
    primed: bool,
    log_2pm1: f64,
    ln2: f64,
    inv_n_sq: f64,
    prev_fgdi: f64,
    last_fgdi: f64,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: HurstDifferenceParams) !HurstDifference {
        if (params.period < 2) {
            return error.InvalidPeriod;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "hurdif({d}{s})", .{ params.period, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Hurst difference {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window = try allocator.alloc(f64, params.period + 1);
        @memset(window, 0.0);

        const n_minus_1 = params.period - 1;
        const period_f: f64 = @floatFromInt(params.period);
        const n_minus_1_f: f64 = @floatFromInt(n_minus_1);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .window = window,
            .period = params.period,
            .n_minus_1 = n_minus_1,
            .window_count = 0,
            .primed = false,
            .log_2pm1 = @log(2.0 * n_minus_1_f),
            .ln2 = @log(2.0),
            .inv_n_sq = 1.0 / (period_f * period_f),
            .prev_fgdi = math.nan(f64),
            .last_fgdi = math.nan(f64),
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *HurstDifference) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *HurstDifference) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *HurstDifference, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const period = self.period;

        if (self.primed) {
            var i: usize = 0;
            while (i < period) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }
            self.window[period] = sample;
        } else {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if (self.window_count <= period) {
                return math.nan(f64);
            }

            self.primed = true;
        }

        // Use the last `period` elements of the window (indices 1..period inclusive).
        // Find min/max for normalization.
        var price_max = self.window[1];
        var price_min = self.window[1];

        var k: usize = 2;
        while (k <= period) : (k += 1) {
            if (self.window[k] > price_max) price_max = self.window[k];
            if (self.window[k] < price_min) price_min = self.window[k];
        }

        const price_range = price_max - price_min;

        var fgdi_val: f64 = undefined;

        if (price_range <= 0.0) {
            fgdi_val = 0.0;
        } else {
            // Normalize and compute path length.
            var prior_norm = (self.window[1] - price_min) / price_range;
            var length: f64 = 0.0;

            k = 2;
            while (k <= period) : (k += 1) {
                const curr_norm = (self.window[k] - price_min) / price_range;
                const diff = curr_norm - prior_norm;
                length += @sqrt(diff * diff + self.inv_n_sq);
                prior_norm = curr_norm;
            }

            if (length > 0.0) {
                fgdi_val = 1.0 + (@log(length) + self.ln2) / self.log_2pm1;
            } else {
                fgdi_val = 0.0;
            }
        }

        // First difference.
        var hurst_diff: f64 = undefined;
        if (math.isNan(self.prev_fgdi)) {
            hurst_diff = math.nan(f64);
        } else {
            hurst_diff = fgdi_val - self.prev_fgdi;
        }

        self.prev_fgdi = fgdi_val;
        self.last_fgdi = fgdi_val;

        return hurst_diff;
    }

    pub fn updateAll(self: *HurstDifference, sample: f64) struct { hurst_diff: f64, fgdi: f64 } {
        const hurst_diff = self.update(sample);
        return .{ .hurst_diff = hurst_diff, .fgdi = self.last_fgdi };
    }

    pub fn isPrimed(self: *const HurstDifference) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const HurstDifference, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .hurst_difference,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
                .{ .mnemonic = "fgdi", .description = "FGDI" },
            },
        );
    }

    pub fn updateScalar(self: *HurstDifference, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *HurstDifference, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *HurstDifference, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *HurstDifference, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *HurstDifference) indicator_mod.Indicator {
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
        const self: *HurstDifference = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const HurstDifference = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *HurstDifference = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *HurstDifference = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *HurstDifference = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *HurstDifference = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidPeriod,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn createHurdif(allocator: std.mem.Allocator, period: usize) !HurstDifference {
    var ind = try HurstDifference.init(allocator, .{ .period = period });
    ind.fixSlices();
    return ind;
}

fn checkValue(exp: f64, act: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
    } else {
        try testing.expect(@abs(exp - act) < 1e-13);
    }
}

test "hurdif update period 5" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFDIP5();
    const exp_hdiff = testdata.expectedHDIFFP5();
    var ind = try createHurdif(testing.allocator, 5);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_hdiff[i], result.hurst_diff);
    }
}

test "hurdif update period 10" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFDIP10();
    const exp_hdiff = testdata.expectedHDIFFP10();
    var ind = try createHurdif(testing.allocator, 10);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_hdiff[i], result.hurst_diff);
    }
}

test "hurdif update period 15" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFDIP15();
    const exp_hdiff = testdata.expectedHDIFFP15();
    var ind = try createHurdif(testing.allocator, 15);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_hdiff[i], result.hurst_diff);
    }
}

test "hurdif update period 20" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFDIP20();
    const exp_hdiff = testdata.expectedHDIFFP20();
    var ind = try createHurdif(testing.allocator, 20);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_hdiff[i], result.hurst_diff);
    }
}

test "hurdif update period 30" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFDIP30();
    const exp_hdiff = testdata.expectedHDIFFP30();
    var ind = try createHurdif(testing.allocator, 30);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_hdiff[i], result.hurst_diff);
    }
}

test "hurdif update period 50" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFDIP50();
    const exp_hdiff = testdata.expectedHDIFFP50();
    var ind = try createHurdif(testing.allocator, 50);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_hdiff[i], result.hurst_diff);
    }
}

test "hurdif update period 80" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFDIP80();
    const exp_hdiff = testdata.expectedHDIFFP80();
    var ind = try createHurdif(testing.allocator, 80);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_hdiff[i], result.hurst_diff);
    }
}

test "hurdif update period 120" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFDIP120();
    const exp_hdiff = testdata.expectedHDIFFP120();
    var ind = try createHurdif(testing.allocator, 120);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_hdiff[i], result.hurst_diff);
    }
}

test "hurdif is primed" {
    const input = testdata.testInput();
    var ind = try createHurdif(testing.allocator, 30);
    defer ind.deinit();

    for (0..30) |i| {
        _ = ind.update(input[i]);
        try testing.expect(!ind.isPrimed());
    }
    _ = ind.update(input[30]);
    try testing.expect(ind.isPrimed());
}

test "hurdif nan passthrough" {
    var ind = try createHurdif(testing.allocator, 5);
    defer ind.deinit();
    const result = ind.updateAll(math.nan(f64));
    try testing.expect(math.isNan(result.hurst_diff));
    try testing.expect(math.isNan(result.fgdi));
}

test "hurdif invalid period" {
    const result = HurstDifference.init(testing.allocator, .{ .period = 1 });
    try testing.expectError(error.InvalidPeriod, result);
}
