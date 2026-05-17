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

/// Enumerates the outputs of the fractal dimension indicator.
pub const FractalDimensionIndexOutput = enum(u8) {
    /// The scalar value of the fractal dimension.
    value = 1,
};

/// Parameters to create an instance of the fractal dimension indicator.
pub const FractalDimensionIndexParams = struct {
    /// The lookback period N. Must be > 1.
    period: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Fractal Dimension Index (FDI).
///
/// Measures the fractal dimension of a price time series using normalized
/// path length. The indicator is not primed during the first `period` updates.
pub const FractalDimensionIndex = struct {
    line: LineIndicator,
    window: []f64,
    period: usize,
    window_count: usize,
    primed: bool,
    log_2n: f64,
    ln2: f64,
    inv_n_sq: f64,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: FractalDimensionIndexParams) !FractalDimensionIndex {
        if (params.period < 2) {
            return error.InvalidPeriod;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "fdi({d}{s})", .{ params.period, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Fractal dimension index {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window = try allocator.alloc(f64, params.period + 1);
        @memset(window, 0.0);

        const period_f: f64 = @floatFromInt(params.period);

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
            .window_count = 0,
            .primed = false,
            .log_2n = @log(2.0 * period_f),
            .ln2 = @log(2.0),
            .inv_n_sq = 1.0 / (period_f * period_f),
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *FractalDimensionIndex) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *FractalDimensionIndex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *FractalDimensionIndex, sample: f64) f64 {
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

        // Find min/max for normalization.
        var price_max = self.window[0];
        var price_min = self.window[0];

        var k: usize = 1;
        while (k <= period) : (k += 1) {
            if (self.window[k] > price_max) price_max = self.window[k];
            if (self.window[k] < price_min) price_min = self.window[k];
        }

        const price_range = price_max - price_min;
        if (price_range < 1e-10) {
            return 1.0;
        }

        // Normalize and compute path length.
        var prior_norm = (self.window[0] - price_min) / price_range;
        var length: f64 = 0.0;

        k = 1;
        while (k <= period) : (k += 1) {
            const curr_norm = (self.window[k] - price_min) / price_range;
            const diff = curr_norm - prior_norm;
            length += @sqrt(diff * diff + self.inv_n_sq);
            prior_norm = curr_norm;
        }

        return 1.0 + (@log(length) + self.ln2) / self.log_2n;
    }

    pub fn isPrimed(self: *const FractalDimensionIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const FractalDimensionIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .fractal_dimension_index,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *FractalDimensionIndex, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *FractalDimensionIndex, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *FractalDimensionIndex, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *FractalDimensionIndex, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *FractalDimensionIndex) indicator_mod.Indicator {
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
        const self: *FractalDimensionIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const FractalDimensionIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *FractalDimensionIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *FractalDimensionIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *FractalDimensionIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *FractalDimensionIndex = @ptrCast(@alignCast(ptr));
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

fn createFdi(allocator: std.mem.Allocator, period: usize) !FractalDimensionIndex {
    var fdi = try FractalDimensionIndex.init(allocator, .{ .period = period });
    fdi.fixSlices();
    return fdi;
}

test "fdi update period 5" {
    const input = testdata.testInput();
    const exp = testdata.expected_P5();
    var fdi = try createFdi(testing.allocator, 5);
    defer fdi.deinit();

    for (0..252) |i| {
        const act = fdi.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "fdi update period 10" {
    const input = testdata.testInput();
    const exp = testdata.expected_P10();
    var fdi = try createFdi(testing.allocator, 10);
    defer fdi.deinit();

    for (0..252) |i| {
        const act = fdi.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "fdi update period 15" {
    const input = testdata.testInput();
    const exp = testdata.expected_P15();
    var fdi = try createFdi(testing.allocator, 15);
    defer fdi.deinit();

    for (0..252) |i| {
        const act = fdi.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "fdi update period 20" {
    const input = testdata.testInput();
    const exp = testdata.expected_P20();
    var fdi = try createFdi(testing.allocator, 20);
    defer fdi.deinit();

    for (0..252) |i| {
        const act = fdi.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "fdi update period 30" {
    const input = testdata.testInput();
    const exp = testdata.expected_P30();
    var fdi = try createFdi(testing.allocator, 30);
    defer fdi.deinit();

    for (0..252) |i| {
        const act = fdi.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "fdi update period 50" {
    const input = testdata.testInput();
    const exp = testdata.expected_P50();
    var fdi = try createFdi(testing.allocator, 50);
    defer fdi.deinit();

    for (0..252) |i| {
        const act = fdi.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "fdi update period 80" {
    const input = testdata.testInput();
    const exp = testdata.expected_P80();
    var fdi = try createFdi(testing.allocator, 80);
    defer fdi.deinit();

    for (0..252) |i| {
        const act = fdi.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "fdi update period 120" {
    const input = testdata.testInput();
    const exp = testdata.expected_P120();
    var fdi = try createFdi(testing.allocator, 120);
    defer fdi.deinit();

    for (0..252) |i| {
        const act = fdi.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "fdi is primed" {
    const input = testdata.testInput();
    var fdi = try createFdi(testing.allocator, 30);
    defer fdi.deinit();

    for (0..30) |i| {
        _ = fdi.update(input[i]);
        try testing.expect(!fdi.isPrimed());
    }
    _ = fdi.update(input[30]);
    try testing.expect(fdi.isPrimed());
}

test "fdi nan passthrough" {
    var fdi = try createFdi(testing.allocator, 5);
    defer fdi.deinit();
    try testing.expect(math.isNan(fdi.update(math.nan(f64))));
}

test "fdi invalid period" {
    const result = FractalDimensionIndex.init(testing.allocator, .{ .period = 1 });
    try testing.expectError(error.InvalidPeriod, result);
}
