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

/// Selects the frequency band of the Sinc Wavelet Band-Pass filter.
pub const Band = enum(u8) {
    /// High-frequency band (periods 8-16 bars).
    high = 0,
    /// Mid-frequency band (periods 16-32 bars).
    mid = 1,
    /// Low-frequency band (periods 32-64 bars).
    low = 2,
    /// Full band (periods 8-64 bars).
    full = 3,
};

/// Enumerates the outputs of the Sinc Wavelet Band-Pass indicator.
pub const SincWaveletBandpassOutput = enum(u8) {
    /// The band-passed price component (or its velocity).
    value = 1,
};

/// Parameters to create a Sinc Wavelet Band-Pass indicator.
pub const SincWaveletBandpassParams = struct {
    band: Band = .mid,
    velocity: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

const velocity_taps = 4;

// Cubic velocity kernel (PFD degree=3, order=1, smoothing=0).
const velocity_kernel = [velocity_taps]f64{ 11.0 / 6.0, -3.0, 3.0 / 2.0, -1.0 / 3.0 };

const BandParams = struct { omega0: f64, omega1: f64, num_taps: usize };

fn bandParams(band: Band) BandParams {
    return switch (band) {
        .high => .{ .omega0 = math.pi / 4.0, .omega1 = math.pi / 8.0, .num_taps = 121 },
        .mid => .{ .omega0 = math.pi / 8.0, .omega1 = math.pi / 16.0, .num_taps = 121 },
        .low => .{ .omega0 = math.pi / 16.0, .omega1 = math.pi / 32.0, .num_taps = 201 },
        .full => .{ .omega0 = math.pi / 4.0, .omega1 = math.pi / 32.0, .num_taps = 201 },
    };
}

fn bandName(band: Band) []const u8 {
    return switch (band) {
        .high => "high",
        .mid => "mid",
        .low => "low",
        .full => "full",
    };
}

/// Computes sinc band-pass filter coefficients (difference of two sinc functions).
fn computeCoefficients(allocator: std.mem.Allocator, omega0: f64, omega1: f64, num_taps: usize) ![]f64 {
    const coeffs = try allocator.alloc(f64, num_taps);
    errdefer allocator.free(coeffs);

    coeffs[0] = (omega0 - omega1) / math.pi;

    var k: usize = 1;
    while (k < num_taps) : (k += 1) {
        const kf: f64 = @floatFromInt(k);
        const pi_k = math.pi * kf;
        coeffs[k] = @sin(omega0 * kf) / pi_k - @sin(omega1 * kf) / pi_k;
    }

    return coeffs;
}

/// Sinc Wavelet Band-Pass (SWB) by Don Mak.
///
/// A causal FIR band-pass filter derived from the sinc wavelet system, decomposing
/// price into frequency bands (HIGH, MID, LOW, FULL). Optionally a cubic velocity
/// kernel is applied to produce a momentum oscillator.
pub const SincWaveletBandpass = struct {
    line: LineIndicator,

    velocity: bool,
    coefficients: []f64,
    num_taps: usize,

    price_buffer: []f64,
    price_count: usize = 0,
    price_index: usize = 0,

    vel_buffer: [velocity_taps]f64 = .{ 0.0, 0.0, 0.0, 0.0 },
    vel_count: usize = 0,
    vel_index: usize = 0,

    primed: bool = false,

    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: SincWaveletBandpassParams) !SincWaveletBandpass {
        const bp = bandParams(params.band);

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = if (params.velocity)
            std.fmt.bufPrint(&mnemonic_buf, "swb({s},v{s})", .{ bandName(params.band), triple }) catch return error.MnemonicTooLong
        else
            std.fmt.bufPrint(&mnemonic_buf, "swb({s}{s})", .{ bandName(params.band), triple }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Sinc wavelet band-pass {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const coefficients = try computeCoefficients(allocator, bp.omega0, bp.omega1, bp.num_taps);
        errdefer allocator.free(coefficients);

        const price_buffer = try allocator.alloc(f64, bp.num_taps);
        @memset(price_buffer, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .velocity = params.velocity,
            .coefficients = coefficients,
            .num_taps = bp.num_taps,
            .price_buffer = price_buffer,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *SincWaveletBandpass) void {
        self.allocator.free(self.coefficients);
        self.allocator.free(self.price_buffer);
    }

    pub fn fixSlices(self: *SincWaveletBandpass) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *SincWaveletBandpass, sample: f64) f64 {
        // Store price in the ring buffer.
        self.price_buffer[self.price_index] = sample;
        self.price_index = (self.price_index + 1) % self.num_taps;
        self.price_count += 1;

        if (self.price_count < self.num_taps) {
            self.primed = false;
            return math.nan(f64);
        }

        // Band-pass convolution: coefficients[k] multiplies the k-th most recent price.
        var bp_value: f64 = 0.0;
        const n: i64 = @intCast(self.num_taps);
        var k: usize = 0;
        while (k < self.num_taps) : (k += 1) {
            const offset = @as(i64, @intCast(self.price_index)) - 1 - @as(i64, @intCast(k));
            const buf_idx: usize = @intCast(@mod(offset, n));
            bp_value += self.coefficients[k] * self.price_buffer[buf_idx];
        }

        if (!self.velocity) {
            self.primed = true;
            return bp_value;
        }

        // Store band-pass output in the velocity ring buffer.
        self.vel_buffer[self.vel_index] = bp_value;
        self.vel_index = (self.vel_index + 1) % velocity_taps;
        self.vel_count += 1;

        if (self.vel_count < velocity_taps) {
            self.primed = false;
            return math.nan(f64);
        }

        // Cubic velocity: kernel[k] multiplies the k-th most recent band-pass value.
        var vel_value: f64 = 0.0;
        const vn: i64 = velocity_taps;
        var j: usize = 0;
        while (j < velocity_taps) : (j += 1) {
            const offset = @as(i64, @intCast(self.vel_index)) - 1 - @as(i64, @intCast(j));
            const buf_idx: usize = @intCast(@mod(offset, vn));
            vel_value += velocity_kernel[j] * self.vel_buffer[buf_idx];
        }

        self.primed = true;
        return vel_value;
    }

    pub fn isPrimed(self: *const SincWaveletBandpass) bool {
        return self.primed;
    }

    fn mnemonic(self: *const SincWaveletBandpass) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const SincWaveletBandpass) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const SincWaveletBandpass, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        build_metadata_mod.buildMetadata(
            out,
            .sinc_wavelet_bandpass,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    pub fn updateScalar(self: *SincWaveletBandpass, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *SincWaveletBandpass, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *SincWaveletBandpass, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *SincWaveletBandpass, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *SincWaveletBandpass) indicator_mod.Indicator {
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
        const self: *SincWaveletBandpass = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const SincWaveletBandpass = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *SincWaveletBandpass = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *SincWaveletBandpass = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *SincWaveletBandpass = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *SincWaveletBandpass = @ptrCast(@alignCast(ptr));
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

const tolerance = 1e-9;

fn checkSeries(params: SincWaveletBandpassParams, inputs: []const f64, expected: []const f64) !void {
    const allocator = testing.allocator;

    var swb = try SincWaveletBandpass.init(allocator, params);
    defer swb.deinit();

    try testing.expectEqual(inputs.len, expected.len);

    for (0..inputs.len) |i| {
        const value = swb.update(inputs[i]);
        const exp = expected[i];
        if (math.isNan(exp)) {
            try testing.expect(math.isNan(value));
        } else {
            try testing.expect(@abs(value - exp) <= tolerance);
        }
    }
}

test "SWB reference data all bands" {
    const input = testdata.testInput();
    const sine = testdata.test1InputSine();
    const mixed = testdata.test2InputMixed();

    try checkSeries(.{ .band = .high }, &input, &testdata.expectedHIGH());
    try checkSeries(.{ .band = .mid }, &input, &testdata.expectedMID());
    try checkSeries(.{ .band = .low }, &input, &testdata.expectedLOW());
    try checkSeries(.{ .band = .full }, &input, &testdata.expectedFULL());
    try checkSeries(.{ .band = .high, .velocity = true }, &input, &testdata.expectedHIGH_V());
    try checkSeries(.{ .band = .mid, .velocity = true }, &input, &testdata.expectedMID_V());
    try checkSeries(.{ .band = .low, .velocity = true }, &input, &testdata.expectedLOW_V());
    try checkSeries(.{ .band = .full, .velocity = true }, &input, &testdata.expectedFULL_V());

    try checkSeries(.{ .band = .mid }, &sine, &testdata.test1ExpectedMID());

    try checkSeries(.{ .band = .high, .velocity = true }, &mixed, &testdata.test2ExpectedHIGH_V());
    try checkSeries(.{ .band = .mid, .velocity = true }, &mixed, &testdata.test2ExpectedMID_V());
    try checkSeries(.{ .band = .low, .velocity = true }, &mixed, &testdata.test2ExpectedLOW_V());
}

test "SWB metadata default" {
    const allocator = testing.allocator;

    var swb = try SincWaveletBandpass.init(allocator, .{});
    defer swb.deinit();
    swb.fixSlices();

    var meta: Metadata = undefined;
    swb.getMetadata(&meta);

    try testing.expectEqual(Identifier.sinc_wavelet_bandpass, meta.identifier);
    try testing.expectEqualStrings("swb(mid)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "SWB mnemonics" {
    const allocator = testing.allocator;

    var high = try SincWaveletBandpass.init(allocator, .{ .band = .high });
    defer high.deinit();
    try testing.expectEqualStrings("swb(high)", high.mnemonic());

    var full = try SincWaveletBandpass.init(allocator, .{ .band = .full });
    defer full.deinit();
    try testing.expectEqualStrings("swb(full)", full.mnemonic());

    var mid_v = try SincWaveletBandpass.init(allocator, .{ .band = .mid, .velocity = true });
    defer mid_v.deinit();
    try testing.expectEqualStrings("swb(mid,v)", mid_v.mnemonic());

    var full_v = try SincWaveletBandpass.init(allocator, .{ .band = .full, .velocity = true });
    defer full_v.deinit();
    try testing.expectEqualStrings("swb(full,v)", full_v.mnemonic());
}
