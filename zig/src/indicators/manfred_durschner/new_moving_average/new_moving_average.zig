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

/// Enumerates the outputs of the New Moving Average indicator.
pub const NewMovingAverageOutput = enum(u8) {
    /// The scalar value of the moving average.
    value = 1,
};

/// MA type used by the NMA indicator.
pub const MAType = enum(u8) {
    sma = 0,
    ema = 1,
    smma = 2,
    lwma = 3,
};

/// Parameters to create an instance of the New Moving Average indicator.
pub const NewMovingAverageParams = struct {
    /// Primary period. 0 means auto-derive from secondary.
    primary_period: usize = 0,
    /// Secondary period. Must be >= 2.
    secondary_period: usize = 8,
    /// Type of moving average to use.
    ma_type: MAType = .lwma,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

// ---------------------------------------------------------------------------
// Streaming MA implementations
// ---------------------------------------------------------------------------

const StreamingSMA = struct {
    period: usize,
    buffer: []f64,
    buffer_index: usize,
    buffer_count: usize,
    sum: f64,
    primed: bool,
    allocator: std.mem.Allocator,

    fn init(allocator: std.mem.Allocator, period: usize) !StreamingSMA {
        const buffer = try allocator.alloc(f64, period);
        @memset(buffer, 0.0);
        return .{
            .period = period,
            .buffer = buffer,
            .buffer_index = 0,
            .buffer_count = 0,
            .sum = 0.0,
            .primed = false,
            .allocator = allocator,
        };
    }

    fn deinit(self: *StreamingSMA) void {
        self.allocator.free(self.buffer);
    }

    fn update(self: *StreamingSMA, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        if (self.primed) {
            self.sum -= self.buffer[self.buffer_index];
        }

        self.buffer[self.buffer_index] = sample;
        self.sum += sample;
        self.buffer_index = (self.buffer_index + 1) % self.period;

        if (!self.primed) {
            self.buffer_count += 1;
            if (self.buffer_count < self.period) return math.nan(f64);
            self.primed = true;
        }

        return self.sum / @as(f64, @floatFromInt(self.period));
    }
};

const StreamingEMA = struct {
    period: usize,
    multiplier: f64,
    count: usize,
    sum: f64,
    value: f64,
    primed: bool,

    fn init(period: usize) StreamingEMA {
        return .{
            .period = period,
            .multiplier = 2.0 / @as(f64, @floatFromInt(period + 1)),
            .count = 0,
            .sum = 0.0,
            .value = math.nan(f64),
            .primed = false,
        };
    }

    fn update(self: *StreamingEMA, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        if (!self.primed) {
            self.count += 1;
            self.sum += sample;
            if (self.count < self.period) return math.nan(f64);
            self.value = self.sum / @as(f64, @floatFromInt(self.period));
            self.primed = true;
            return self.value;
        }

        self.value = (sample - self.value) * self.multiplier + self.value;
        return self.value;
    }
};

const StreamingSMMA = struct {
    period: usize,
    count: usize,
    sum: f64,
    value: f64,
    primed: bool,

    fn init(period: usize) StreamingSMMA {
        return .{
            .period = period,
            .count = 0,
            .sum = 0.0,
            .value = math.nan(f64),
            .primed = false,
        };
    }

    fn update(self: *StreamingSMMA, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        if (!self.primed) {
            self.count += 1;
            self.sum += sample;
            if (self.count < self.period) return math.nan(f64);
            self.value = self.sum / @as(f64, @floatFromInt(self.period));
            self.primed = true;
            return self.value;
        }

        const p = @as(f64, @floatFromInt(self.period));
        self.value = (self.value * (p - 1.0) + sample) / p;
        return self.value;
    }
};

const StreamingLWMA = struct {
    period: usize,
    buffer: []f64,
    buffer_index: usize,
    buffer_count: usize,
    weight_sum: f64,
    primed: bool,
    allocator: std.mem.Allocator,

    fn init(allocator: std.mem.Allocator, period: usize) !StreamingLWMA {
        const buffer = try allocator.alloc(f64, period);
        @memset(buffer, 0.0);
        const p = @as(f64, @floatFromInt(period));
        return .{
            .period = period,
            .buffer = buffer,
            .buffer_index = 0,
            .buffer_count = 0,
            .weight_sum = p * (p + 1.0) / 2.0,
            .primed = false,
            .allocator = allocator,
        };
    }

    fn deinit(self: *StreamingLWMA) void {
        self.allocator.free(self.buffer);
    }

    fn update(self: *StreamingLWMA, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        self.buffer[self.buffer_index] = sample;
        self.buffer_index = (self.buffer_index + 1) % self.period;

        if (!self.primed) {
            self.buffer_count += 1;
            if (self.buffer_count < self.period) return math.nan(f64);
            self.primed = true;
        }

        var result: f64 = 0.0;
        var index = self.buffer_index;
        for (0..self.period) |i| {
            result += @as(f64, @floatFromInt(i + 1)) * self.buffer[index];
            index = (index + 1) % self.period;
        }

        return result / self.weight_sum;
    }
};

const StreamingMA = union(enum) {
    sma: StreamingSMA,
    ema: StreamingEMA,
    smma: StreamingSMMA,
    lwma: StreamingLWMA,

    fn init(allocator: std.mem.Allocator, ma_type: MAType, period: usize) !StreamingMA {
        return switch (ma_type) {
            .sma => .{ .sma = try StreamingSMA.init(allocator, period) },
            .ema => .{ .ema = StreamingEMA.init(period) },
            .smma => .{ .smma = StreamingSMMA.init(period) },
            .lwma => .{ .lwma = try StreamingLWMA.init(allocator, period) },
        };
    }

    fn deinit(self: *StreamingMA) void {
        switch (self.*) {
            .sma => |*s| s.deinit(),
            .lwma => |*s| s.deinit(),
            .ema, .smma => {},
        }
    }

    fn update(self: *StreamingMA, sample: f64) f64 {
        return switch (self.*) {
            .sma => |*s| s.update(sample),
            .ema => |*s| s.update(sample),
            .smma => |*s| s.update(sample),
            .lwma => |*s| s.update(sample),
        };
    }
};

// ---------------------------------------------------------------------------
// NewMovingAverage
// ---------------------------------------------------------------------------

/// Computes the New Moving Average (NMA) by Manfred Dürschner.
///
/// NMA applies the Nyquist-Shannon sampling theorem to moving average design:
/// by cascading two moving averages whose period ratio satisfies the Nyquist
/// criterion (lambda = n1/n2 >= 2), the resulting lag can be extrapolated away
/// geometrically.
///
/// Formula: NMA = (1 + alpha) * MA1 - alpha * MA2
/// where: alpha = lambda * (n1-1) / (n1-lambda), lambda = n1 / n2
pub const NewMovingAverage = struct {
    line: LineIndicator,
    alpha: f64,
    ma_primary: StreamingMA,
    ma_secondary: StreamingMA,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: NewMovingAverageParams) !NewMovingAverage {
        var primary_period = params.primary_period;
        var secondary_period = params.secondary_period;

        // Enforce Nyquist constraint.
        if (primary_period < 4) primary_period = 4;
        if (secondary_period < 2) secondary_period = 2;
        if (primary_period < secondary_period * 2) primary_period = secondary_period * 4;

        // Compute alpha.
        const nyquist_ratio = primary_period / secondary_period;
        const alpha = @as(f64, @floatFromInt(nyquist_ratio)) * @as(f64, @floatFromInt(primary_period - 1)) / @as(f64, @floatFromInt(primary_period - nyquist_ratio));

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Build mnemonic: "nma({pri}, {sec}, {maType}{triple})"
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "nma({d}, {d}, {d}{s})", .{ primary_period, secondary_period, @intFromEnum(params.ma_type), triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "New moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        var ma_primary = try StreamingMA.init(allocator, params.ma_type, primary_period);
        errdefer ma_primary.deinit();

        var ma_secondary = try StreamingMA.init(allocator, params.ma_type, secondary_period);
        errdefer ma_secondary.deinit();

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .alpha = alpha,
            .ma_primary = ma_primary,
            .ma_secondary = ma_secondary,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *NewMovingAverage) void {
        self.ma_primary.deinit();
        self.ma_secondary.deinit();
    }

    /// Fix up the line's mnemonic/description slices to point into self's own buffers.
    pub fn fixSlices(self: *NewMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update logic. Returns the NMA value or NaN if not yet primed.
    pub fn update(self: *NewMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        // First filter: MA of raw price.
        const ma1_value = self.ma_primary.update(sample);
        if (math.isNan(ma1_value)) return math.nan(f64);

        // Second filter: MA of MA1 output.
        const ma2_value = self.ma_secondary.update(ma1_value);
        if (math.isNan(ma2_value)) return math.nan(f64);

        self.primed = true;

        // Geometric extrapolation.
        return (1.0 + self.alpha) * ma1_value - self.alpha * ma2_value;
    }

    /// Returns whether the indicator has accumulated enough data.
    pub fn isPrimed(self: *const NewMovingAverage) bool {
        return self.primed;
    }

    /// Returns metadata for this indicator.
    pub fn getMetadata(self: *const NewMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .new_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *NewMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *NewMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *NewMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *NewMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *NewMovingAverage) indicator_mod.Indicator {
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
        const self: *NewMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const NewMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *NewMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *NewMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *NewMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *NewMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn createNma(allocator: std.mem.Allocator, primary_period: usize, secondary_period: usize, ma_type: MAType) !NewMovingAverage {
    var nma = try NewMovingAverage.init(allocator, .{ .primary_period = primary_period, .secondary_period = secondary_period, .ma_type = ma_type });
    nma.fixSlices();
    return nma;
}

fn checkNmaUpdate(nma: *NewMovingAverage, input: []const f64, expected: []const f64) !void {
    for (0..input.len) |i| {
        const act = nma.update(input[i]);
        const exp = expected[i];
        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(act - exp) < 1e-13);
        }
    }

    // NaN passthrough.
    try testing.expect(math.isNan(nma.update(math.nan(f64))));
}

test "nma pri8 sec4 LWMA" {
    const input = &testdata.test_input;
    const exp = &testdata.expected_pri8_sec4_lwma;
    var nma = try createNma(testing.allocator, 8, 4, .lwma);
    defer nma.deinit();
    try checkNmaUpdate(&nma, input, exp);
}

test "nma sec8 pri_auto LWMA (default)" {
    const input = &testdata.test_input;
    const exp = &testdata.expected_sec8_pri_auto_lwma;
    var nma = try createNma(testing.allocator, 0, 8, .lwma);
    defer nma.deinit();
    try checkNmaUpdate(&nma, input, exp);
}

test "nma sec8 pri_auto SMA" {
    const input = &testdata.test_input;
    const exp = &testdata.expected_sec8_sma;
    var nma = try createNma(testing.allocator, 0, 8, .sma);
    defer nma.deinit();
    try checkNmaUpdate(&nma, input, exp);
}

test "nma sec8 pri_auto EMA" {
    const input = &testdata.test_input;
    const exp = &testdata.expected_sec8_ema;
    var nma = try createNma(testing.allocator, 0, 8, .ema);
    defer nma.deinit();
    try checkNmaUpdate(&nma, input, exp);
}

test "nma sec8 pri_auto SMMA" {
    const input = &testdata.test_input;
    const exp = &testdata.expected_sec8_smma;
    var nma = try createNma(testing.allocator, 0, 8, .smma);
    defer nma.deinit();
    try checkNmaUpdate(&nma, input, exp);
}

test "nma is primed" {
    const input = &testdata.test_input;
    var nma = try createNma(testing.allocator, 0, 8, .lwma);
    defer nma.deinit();

    // With default params: pri=32, sec=8. Warmup = 32 + 8 - 2 = 38 bars.
    for (0..38) |i| {
        _ = nma.update(input[i]);
        try testing.expect(!nma.isPrimed());
    }

    _ = nma.update(input[38]);
    try testing.expect(nma.isPrimed());
}

test "nma metadata" {
    var nma = try createNma(testing.allocator, 0, 8, .lwma);
    defer nma.deinit();
    var m: Metadata = undefined;
    nma.getMetadata(&m);

    try testing.expectEqual(Identifier.new_moving_average, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqualStrings("nma(32, 8, 3)", nma.line.mnemonic);
}
