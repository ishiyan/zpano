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

/// Enumerates the outputs of the fractal adaptive simple moving average indicator.
pub const FractalAdaptiveSimpleMovingAverageOutput = enum(u8) {
    /// The scalar value of the FRASMA.
    value = 0,
};

/// Parameters to create an instance of the fractal adaptive simple moving average indicator.
pub const FractalAdaptiveSimpleMovingAverageParams = struct {
    /// The lookback period N for FDI computation. Must be > 1.
    period: usize,
    /// Base SMA period before fractal adaptation. Must be > 0.
    normal_speed: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Fractal Adaptive Simple Moving Average (FRASMA).
///
/// Uses the Fractal Dimension Index (FDI) formula to adaptively modify an SMA's period.
/// The indicator is not primed during the first `period - 1` updates.
pub const FractalAdaptiveSimpleMovingAverage = struct {
    line: LineIndicator,
    window: []f64,
    closes: std.ArrayList(f64),
    period: usize,
    normal_speed: usize,
    window_count: usize,
    primed: bool,
    log_2p: f64,
    ln2: f64,
    inv_p_sq: f64,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: FractalAdaptiveSimpleMovingAverageParams) !FractalAdaptiveSimpleMovingAverage {
        if (params.period < 2) {
            return error.InvalidPeriod;
        }
        if (params.normal_speed < 1) {
            return error.InvalidNormalSpeed;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "frasma({d},{d}{s})", .{ params.period, params.normal_speed, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Fractal adaptive simple moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window = try allocator.alloc(f64, params.period);
        @memset(window, 0.0);

        const period_f: f64 = @floatFromInt(params.period);

        var closes: std.ArrayList(f64) = .empty;
        try closes.ensureTotalCapacity(allocator, 256);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .window = window,
            .closes = closes,
            .period = params.period,
            .normal_speed = params.normal_speed,
            .window_count = 0,
            .primed = false,
            .log_2p = @log(2.0 * period_f),
            .ln2 = @log(2.0),
            .inv_p_sq = 1.0 / (period_f * period_f),
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *FractalAdaptiveSimpleMovingAverage) void {
        self.allocator.free(self.window);
        self.closes.deinit(self.allocator);
    }

    pub fn fixSlices(self: *FractalAdaptiveSimpleMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *FractalAdaptiveSimpleMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const period = self.period;

        // Accumulate close history for SMA computation.
        self.closes.append(self.allocator, sample) catch return math.nan(f64);

        // Fill the FDI window.
        if (self.window_count < period) {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if (self.window_count < period) {
                return math.nan(f64);
            }

            self.primed = true;
        } else {
            var i: usize = 0;
            while (i < period - 1) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }
            self.window[period - 1] = sample;
        }

        // --- Compute FDI using iliko's original formula (period-2 segments) ---
        var price_max = self.window[0];
        var price_min = self.window[0];

        var k: usize = 1;
        while (k < period) : (k += 1) {
            if (self.window[k] > price_max) price_max = self.window[k];
            if (self.window[k] < price_min) price_min = self.window[k];
        }

        const price_range = price_max - price_min;
        if (price_range < 1e-10) {
            return math.nan(f64);
        }

        // iliko skips iteration 0: prior_norm starts at window[1], loop from window[2].
        var prior_norm = (self.window[1] - price_min) / price_range;
        var length: f64 = 0.0;

        k = 2;
        while (k < period) : (k += 1) {
            const curr_norm = (self.window[k] - price_min) / price_range;
            const diff = curr_norm - prior_norm;
            length += @sqrt(diff * diff + self.inv_p_sq);
            prior_norm = curr_norm;
        }

        if (length <= 0.0) {
            return math.nan(f64);
        }

        const fdi = 1.0 + (@log(length) + self.ln2) / self.log_2p;

        // --- Adaptive speed ---
        const denom = 2.0 - fdi;
        if (@abs(denom) < 1e-10) {
            return math.nan(f64);
        }

        const trail_dim = 1.0 / denom;
        const alpha = trail_dim / 2.0;
        const ns_f: f64 = @floatFromInt(self.normal_speed);
        const speed_raw = @round(ns_f * alpha);
        const speed: usize = if (speed_raw < 1.0) 1 else @intFromFloat(speed_raw);

        // --- SMA of length `speed` ending at current position ---
        const n_closes = self.closes.items.len;
        if (speed > n_closes) {
            return math.nan(f64);
        }

        var sma_sum: f64 = 0.0;
        var idx: usize = n_closes - speed;
        while (idx < n_closes) : (idx += 1) {
            sma_sum += self.closes.items[idx];
        }

        return sma_sum / @as(f64, @floatFromInt(speed));
    }

    pub fn isPrimed(self: *const FractalAdaptiveSimpleMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const FractalAdaptiveSimpleMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .fractal_adaptive_simple_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *FractalAdaptiveSimpleMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *FractalAdaptiveSimpleMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *FractalAdaptiveSimpleMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *FractalAdaptiveSimpleMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *FractalAdaptiveSimpleMovingAverage) indicator_mod.Indicator {
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
        const self: *FractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const FractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *FractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *FractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *FractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *FractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidPeriod,
        InvalidNormalSpeed,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn createFrasma(allocator: std.mem.Allocator, period: usize, normal_speed: usize) !FractalAdaptiveSimpleMovingAverage {
    var f = try FractalAdaptiveSimpleMovingAverage.init(allocator, .{ .period = period, .normal_speed = normal_speed });
    f.fixSlices();
    return f;
}

test "frasma update period 5" {
    const input = testdata.testInput();
    const exp = testdata.expectedP5();
    var f = try createFrasma(testing.allocator, 5, 20);
    defer f.deinit();

    for (0..252) |i| {
        const act = f.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "frasma update period 10" {
    const input = testdata.testInput();
    const exp = testdata.expectedP10();
    var f = try createFrasma(testing.allocator, 10, 20);
    defer f.deinit();

    for (0..252) |i| {
        const act = f.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "frasma update period 15" {
    const input = testdata.testInput();
    const exp = testdata.expectedP15();
    var f = try createFrasma(testing.allocator, 15, 20);
    defer f.deinit();

    for (0..252) |i| {
        const act = f.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "frasma update period 20" {
    const input = testdata.testInput();
    const exp = testdata.expectedP20();
    var f = try createFrasma(testing.allocator, 20, 20);
    defer f.deinit();

    for (0..252) |i| {
        const act = f.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "frasma update period 30" {
    const input = testdata.testInput();
    const exp = testdata.expectedP30();
    var f = try createFrasma(testing.allocator, 30, 20);
    defer f.deinit();

    for (0..252) |i| {
        const act = f.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "frasma update period 50" {
    const input = testdata.testInput();
    const exp = testdata.expectedP50();
    var f = try createFrasma(testing.allocator, 50, 20);
    defer f.deinit();

    for (0..252) |i| {
        const act = f.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "frasma update period 80" {
    const input = testdata.testInput();
    const exp = testdata.expectedP80();
    var f = try createFrasma(testing.allocator, 80, 20);
    defer f.deinit();

    for (0..252) |i| {
        const act = f.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "frasma update period 120" {
    const input = testdata.testInput();
    const exp = testdata.expectedP120();
    var f = try createFrasma(testing.allocator, 120, 20);
    defer f.deinit();

    for (0..252) |i| {
        const act = f.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "frasma is primed" {
    const input = testdata.testInput();
    var f = try createFrasma(testing.allocator, 30, 20);
    defer f.deinit();

    for (0..29) |i| {
        _ = f.update(input[i]);
        try testing.expect(!f.isPrimed());
    }
    _ = f.update(input[29]);
    try testing.expect(f.isPrimed());
}

test "frasma nan passthrough" {
    var f = try createFrasma(testing.allocator, 5, 20);
    defer f.deinit();
    try testing.expect(math.isNan(f.update(math.nan(f64))));
}

test "frasma invalid period" {
    const result = FractalAdaptiveSimpleMovingAverage.init(testing.allocator, .{ .period = 1, .normal_speed = 20 });
    try testing.expectError(error.InvalidPeriod, result);
}

test "frasma invalid normal_speed" {
    const result = FractalAdaptiveSimpleMovingAverage.init(testing.allocator, .{ .period = 5, .normal_speed = 0 });
    try testing.expectError(error.InvalidNormalSpeed, result);
}
