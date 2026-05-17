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
const band_mod = @import("../../core/outputs/band.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Band = band_mod.Band;

/// Enumerates the outputs of the fractal graph dimension index indicator.
pub const FractalGraphDimensionIndexOutput = enum(u8) {
    /// The fractal graph dimension value.
    fgdi = 1,
    /// The upper band (fgdi + stddev).
    upper = 2,
    /// The lower band (fgdi - stddev).
    lower = 3,
    /// The standard deviation of the dimension estimate.
    stddev = 4,
    /// The lower/upper band pair.
    band = 5,
};

/// Parameters to create an instance of the fractal graph dimension index indicator.
pub const FractalGraphDimensionIndexParams = struct {
    /// The lookback period N. Must be > 1.
    period: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Fractal Graph Dimension Index (FGDI).
///
/// This is Poton's corrected and enhanced version of the Fractal Dimension
/// Index (FDI). It fixes loop boundary and denominator bugs in the original
/// and adds standard deviation bands around the estimated dimension.
///
/// The indicator is not primed during the first `period - 1` updates.
pub const FractalGraphDimensionIndex = struct {
    line: LineIndicator,
    window: []f64,
    segments: []f64,
    period: usize,
    n_minus_1: usize,
    window_count: usize,
    primed: bool,
    log_2n1: f64,
    ln2: f64,
    inv_n_sq: f64,
    fgdi: f64,
    upper: f64,
    lower: f64,
    stddev_val: f64,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: FractalGraphDimensionIndexParams) !FractalGraphDimensionIndex {
        if (params.period < 2) {
            return error.InvalidPeriod;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "fgdi({d}{s})", .{ params.period, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Fractal graph dimension index {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window = try allocator.alloc(f64, params.period);
        @memset(window, 0.0);

        const n_minus_1 = params.period - 1;
        const segments = try allocator.alloc(f64, n_minus_1);
        @memset(segments, 0.0);

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
            .segments = segments,
            .period = params.period,
            .n_minus_1 = n_minus_1,
            .window_count = 0,
            .primed = false,
            .log_2n1 = @log(2.0 * n_minus_1_f),
            .ln2 = @log(2.0),
            .inv_n_sq = 1.0 / (period_f * period_f),
            .fgdi = math.nan(f64),
            .upper = math.nan(f64),
            .lower = math.nan(f64),
            .stddev_val = math.nan(f64),
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *FractalGraphDimensionIndex) void {
        self.allocator.free(self.window);
        self.allocator.free(self.segments);
    }

    pub fn fixSlices(self: *FractalGraphDimensionIndex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *FractalGraphDimensionIndex, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const period = self.period;
        const n_minus_1 = self.n_minus_1;

        if (self.primed) {
            var i: usize = 0;
            while (i < n_minus_1) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }
            self.window[n_minus_1] = sample;
        } else {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if (self.window_count < period) {
                return math.nan(f64);
            }

            self.primed = true;
        }

        // Find min/max for normalization.
        var price_max = self.window[0];
        var price_min = self.window[0];

        var k: usize = 1;
        while (k < period) : (k += 1) {
            if (self.window[k] > price_max) price_max = self.window[k];
            if (self.window[k] < price_min) price_min = self.window[k];
        }

        const price_range = price_max - price_min;
        if (price_range < 1e-10) {
            self.fgdi = 1.0;
            self.stddev_val = 0.0;
            self.upper = 1.0;
            self.lower = 1.0;
            return 1.0;
        }

        // Normalize and compute path segments.
        var prior_norm = (self.window[0] - price_min) / price_range;
        var length: f64 = 0.0;

        k = 1;
        while (k < period) : (k += 1) {
            const curr_norm = (self.window[k] - price_min) / price_range;
            const diff = curr_norm - prior_norm;
            const seg = @sqrt(diff * diff + self.inv_n_sq);
            self.segments[k - 1] = seg;
            length += seg;
            prior_norm = curr_norm;
        }

        // FGDI = 1 + (ln(L) + ln(2)) / ln(2*(N-1))
        const fgdi_val = 1.0 + (@log(length) + self.ln2) / self.log_2n1;

        // Standard deviation of the estimate.
        const n_minus_1_f: f64 = @floatFromInt(n_minus_1);
        const mean_seg = length / n_minus_1_f;
        var sum_sq: f64 = 0.0;

        var j: usize = 0;
        while (j < n_minus_1) : (j += 1) {
            const d = self.segments[j] - mean_seg;
            sum_sq += d * d;
        }

        const variance = sum_sq / (length * length * self.log_2n1 * self.log_2n1);
        const stddev = @sqrt(variance);

        self.fgdi = fgdi_val;
        self.upper = fgdi_val + stddev;
        self.lower = fgdi_val - stddev;
        self.stddev_val = stddev;

        return fgdi_val;
    }

    pub fn updateAll(self: *FractalGraphDimensionIndex, sample: f64) struct { fgdi: f64, upper: f64, lower: f64, stddev: f64 } {
        const fgdi_val = self.update(sample);
        return .{ .fgdi = fgdi_val, .upper = self.upper, .lower = self.lower, .stddev = self.stddev_val };
    }

    pub fn isPrimed(self: *const FractalGraphDimensionIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const FractalGraphDimensionIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .fractal_graph_dimension_index,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
                .{ .mnemonic = "upper", .description = "Upper" },
                .{ .mnemonic = "lower", .description = "Lower" },
                .{ .mnemonic = "stddev", .description = "Stddev" },
                .{ .mnemonic = "band", .description = "Band" },
            },
        );
    }

    pub fn updateScalar(self: *FractalGraphDimensionIndex, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = sample.time, .value = value } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = self.upper } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = self.lower } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = self.stddev_val } });

        if (math.isNan(self.lower) or math.isNan(self.upper)) {
            out.append(.{ .band = Band.empty(sample.time) });
        } else {
            out.append(.{ .band = Band.new(sample.time, self.lower, self.upper) });
        }

        return out;
    }

    pub fn updateBar(self: *FractalGraphDimensionIndex, sample: *const Bar) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractBar(sample) });
    }

    pub fn updateQuote(self: *FractalGraphDimensionIndex, sample: *const Quote) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractQuote(sample) });
    }

    pub fn updateTrade(self: *FractalGraphDimensionIndex, sample: *const Trade) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractTrade(sample) });
    }

    pub fn indicator(self: *FractalGraphDimensionIndex) indicator_mod.Indicator {
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
        const self: *FractalGraphDimensionIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const FractalGraphDimensionIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *FractalGraphDimensionIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *FractalGraphDimensionIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *FractalGraphDimensionIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *FractalGraphDimensionIndex = @ptrCast(@alignCast(ptr));
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

fn createFgdi(allocator: std.mem.Allocator, period: usize) !FractalGraphDimensionIndex {
    var ind = try FractalGraphDimensionIndex.init(allocator, .{ .period = period });
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

test "fgdi update period 5" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFdiP5();
    const exp_upper = testdata.expectedUpperP5();
    const exp_lower = testdata.expectedLowerP5();
    const exp_stddev = testdata.expectedStddevP5();
    var ind = try createFgdi(testing.allocator, 5);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_upper[i], result.upper);
        try checkValue(exp_lower[i], result.lower);
        try checkValue(exp_stddev[i], result.stddev);
    }
}

test "fgdi update period 10" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFdiP10();
    const exp_upper = testdata.expectedUpperP10();
    const exp_lower = testdata.expectedLowerP10();
    const exp_stddev = testdata.expectedStddevP10();
    var ind = try createFgdi(testing.allocator, 10);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_upper[i], result.upper);
        try checkValue(exp_lower[i], result.lower);
        try checkValue(exp_stddev[i], result.stddev);
    }
}

test "fgdi update period 15" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFdiP15();
    const exp_upper = testdata.expectedUpperP15();
    const exp_lower = testdata.expectedLowerP15();
    const exp_stddev = testdata.expectedStddevP15();
    var ind = try createFgdi(testing.allocator, 15);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_upper[i], result.upper);
        try checkValue(exp_lower[i], result.lower);
        try checkValue(exp_stddev[i], result.stddev);
    }
}

test "fgdi update period 20" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFdiP20();
    const exp_upper = testdata.expectedUpperP20();
    const exp_lower = testdata.expectedLowerP20();
    const exp_stddev = testdata.expectedStddevP20();
    var ind = try createFgdi(testing.allocator, 20);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_upper[i], result.upper);
        try checkValue(exp_lower[i], result.lower);
        try checkValue(exp_stddev[i], result.stddev);
    }
}

test "fgdi update period 30" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFdiP30();
    const exp_upper = testdata.expectedUpperP30();
    const exp_lower = testdata.expectedLowerP30();
    const exp_stddev = testdata.expectedStddevP30();
    var ind = try createFgdi(testing.allocator, 30);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_upper[i], result.upper);
        try checkValue(exp_lower[i], result.lower);
        try checkValue(exp_stddev[i], result.stddev);
    }
}

test "fgdi update period 50" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFdiP50();
    const exp_upper = testdata.expectedUpperP50();
    const exp_lower = testdata.expectedLowerP50();
    const exp_stddev = testdata.expectedStddevP50();
    var ind = try createFgdi(testing.allocator, 50);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_upper[i], result.upper);
        try checkValue(exp_lower[i], result.lower);
        try checkValue(exp_stddev[i], result.stddev);
    }
}

test "fgdi update period 80" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFdiP80();
    const exp_upper = testdata.expectedUpperP80();
    const exp_lower = testdata.expectedLowerP80();
    const exp_stddev = testdata.expectedStddevP80();
    var ind = try createFgdi(testing.allocator, 80);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_upper[i], result.upper);
        try checkValue(exp_lower[i], result.lower);
        try checkValue(exp_stddev[i], result.stddev);
    }
}

test "fgdi update period 120" {
    const input = testdata.testInput();
    const exp_fgdi = testdata.expectedFdiP120();
    const exp_upper = testdata.expectedUpperP120();
    const exp_lower = testdata.expectedLowerP120();
    const exp_stddev = testdata.expectedStddevP120();
    var ind = try createFgdi(testing.allocator, 120);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_fgdi[i], result.fgdi);
        try checkValue(exp_upper[i], result.upper);
        try checkValue(exp_lower[i], result.lower);
        try checkValue(exp_stddev[i], result.stddev);
    }
}

test "fgdi is primed" {
    const input = testdata.testInput();
    var ind = try createFgdi(testing.allocator, 30);
    defer ind.deinit();

    for (0..29) |i| {
        _ = ind.update(input[i]);
        try testing.expect(!ind.isPrimed());
    }
    _ = ind.update(input[29]);
    try testing.expect(ind.isPrimed());
}

test "fgdi nan passthrough" {
    var ind = try createFgdi(testing.allocator, 5);
    defer ind.deinit();
    const result = ind.updateAll(math.nan(f64));
    try testing.expect(math.isNan(result.fgdi));
    try testing.expect(math.isNan(result.upper));
    try testing.expect(math.isNan(result.lower));
    try testing.expect(math.isNan(result.stddev));
}

test "fgdi invalid period" {
    const result = FractalGraphDimensionIndex.init(testing.allocator, .{ .period = 1 });
    try testing.expectError(error.InvalidPeriod, result);
}
