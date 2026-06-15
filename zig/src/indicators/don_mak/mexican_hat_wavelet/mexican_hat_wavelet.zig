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

/// Selects the frequency band of the Mexican Hat Wavelet filter.
pub const Band = enum(u8) {
    /// High-frequency band (a_f = 1.483, period ~ 4.6 bars).
    high = 0,
    /// Mid-frequency band (a_f = 4.048, period ~ 13.5 bars).
    mid = 1,
    /// Low-frequency band (a_f = 15.97, period ~ 54 bars).
    low = 2,
    /// User-specified dilation or period.
    custom = 3,
};

/// Enumerates the outputs of the Mexican Hat Wavelet indicator.
pub const MexicanHatWaveletOutput = enum(u8) {
    /// The bandpass-filtered price component.
    value = 1,
};

/// Parameters to create a Mexican Hat Wavelet indicator.
pub const MexicanHatWaveletParams = struct {
    band: Band = .mid,
    dilation: f64 = 0.0,
    period: f64 = 0.0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

// Preset dilation values (a_f) for the three standard bands (Table 5.2).
const dilation_high = 1.483;
const dilation_mid = 4.048;
const dilation_low = 15.97;

/// Rounds half to even (banker's rounding), matching Python's round() for x > 0.
fn roundHalfEven(x: f64) f64 {
    const frac = @abs(x - math.trunc(x));
    if (frac == 0.5) {
        const f = math.floor(x);
        const fi: i64 = @intFromFloat(f);
        return if (@mod(fi, 2) == 0) f else f + 1.0;
    }
    return math.round(x);
}

/// Computes dilation a_f from a desired center period in bars (Eq 5.11).
fn dilationFromPeriod(period: f64) !f64 {
    const omega0 = 2.0 * math.pi / period;
    const two_over_a = 1.091 * omega0 - 0.071 * omega0 * omega0;
    if (two_over_a <= 0.0) return error.InvalidPeriod;
    return 2.0 / two_over_a;
}

/// Computes normalized Mexican Hat wavelet FIR coefficients for dilation a_f.
fn computeCoefficients(allocator: std.mem.Allocator, a_f: f64) ![]f64 {
    var k: usize = @intFromFloat(4.0 * roundHalfEven(a_f));
    if (k < 1) k = 1;

    const norm = 0.488 + 0.646 * a_f + 0.0001 * a_f * a_f;

    const coeffs = try allocator.alloc(f64, k + 1);
    errdefer allocator.free(coeffs);

    var n: usize = 0;
    while (n <= k) : (n += 1) {
        const t = @as(f64, @floatFromInt(n)) / a_f;
        const t2 = t * t;
        const h_n = (1.0 - 2.0 * t2) * @exp(-t2);
        coeffs[n] = h_n / norm;
    }

    return coeffs;
}

/// Mexican Hat Wavelet (MHW) by Don Mak.
///
/// A causal bandpass FIR filter derived from the Mexican Hat wavelet (the second
/// derivative of a Gaussian), decomposing price into frequency bands with zero
/// phase shift at the filter's center frequency.
pub const MexicanHatWavelet = struct {
    line: LineIndicator,

    coefficients: []f64,
    num_taps: usize,

    buffer: []f64,
    count: usize = 0,

    primed: bool = false,

    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: MexicanHatWaveletParams) !MexicanHatWavelet {
        var a_f: f64 = undefined;
        var cfg_buf: [32]u8 = undefined;
        var cfg: []const u8 = undefined;

        switch (params.band) {
            .high => {
                a_f = dilation_high;
                cfg = "high";
            },
            .mid => {
                a_f = dilation_mid;
                cfg = "mid";
            },
            .low => {
                a_f = dilation_low;
                cfg = "low";
            },
            .custom => {
                const has_dilation = params.dilation != 0.0;
                const has_period = params.period != 0.0;
                if (has_dilation and has_period) return error.BothDilationAndPeriod;
                if (!has_dilation and !has_period) return error.MissingDilationAndPeriod;
                if (has_period) {
                    if (params.period <= 2.0) return error.InvalidPeriod;
                    a_f = try dilationFromPeriod(params.period);
                    cfg = std.fmt.bufPrint(&cfg_buf, "p{d:.2}", .{params.period}) catch return error.MnemonicTooLong;
                } else {
                    if (params.dilation <= 0.0) return error.InvalidDilation;
                    a_f = params.dilation;
                    cfg = std.fmt.bufPrint(&cfg_buf, "d{d:.2}", .{params.dilation}) catch return error.MnemonicTooLong;
                }
            },
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "mhw({s}{s})", .{ cfg, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Mexican hat wavelet {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const coefficients = try computeCoefficients(allocator, a_f);
        errdefer allocator.free(coefficients);

        const buffer = try allocator.alloc(f64, coefficients.len);
        @memset(buffer, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .coefficients = coefficients,
            .num_taps = coefficients.len,
            .buffer = buffer,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *MexicanHatWavelet) void {
        self.allocator.free(self.coefficients);
        self.allocator.free(self.buffer);
    }

    pub fn fixSlices(self: *MexicanHatWavelet) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *MexicanHatWavelet, sample: f64) f64 {
        // Shift buffer right and insert the new price at position 0.
        var i: usize = self.num_taps - 1;
        while (i > 0) : (i -= 1) {
            self.buffer[i] = self.buffer[i - 1];
        }
        self.buffer[0] = sample;
        self.count += 1;

        if (self.count < self.num_taps) {
            self.primed = false;
            return math.nan(f64);
        }

        // FIR convolution: y = sum(coefficients[k] * buffer[k]).
        var y: f64 = 0.0;
        var k: usize = 0;
        while (k < self.num_taps) : (k += 1) {
            y += self.coefficients[k] * self.buffer[k];
        }

        self.primed = true;
        return y;
    }

    pub fn isPrimed(self: *const MexicanHatWavelet) bool {
        return self.primed;
    }

    fn mnemonic(self: *const MexicanHatWavelet) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const MexicanHatWavelet) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const MexicanHatWavelet, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        build_metadata_mod.buildMetadata(
            out,
            .mexican_hat_wavelet,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    pub fn updateScalar(self: *MexicanHatWavelet, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *MexicanHatWavelet, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *MexicanHatWavelet, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *MexicanHatWavelet, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *MexicanHatWavelet) indicator_mod.Indicator {
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
        const self: *MexicanHatWavelet = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const MexicanHatWavelet = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *MexicanHatWavelet = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *MexicanHatWavelet = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *MexicanHatWavelet = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *MexicanHatWavelet = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        BothDilationAndPeriod,
        MissingDilationAndPeriod,
        InvalidPeriod,
        InvalidDilation,
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

fn checkSeries(name: []const u8, params: MexicanHatWaveletParams, inputs: []const f64, expected: []const f64) !void {
    _ = name;
    const allocator = testing.allocator;

    var mhw = try MexicanHatWavelet.init(allocator, params);
    defer mhw.deinit();

    try testing.expectEqual(inputs.len, expected.len);

    for (0..inputs.len) |i| {
        const value = mhw.update(inputs[i]);
        const exp = expected[i];
        if (math.isNan(exp)) {
            try testing.expect(math.isNan(value));
        } else {
            try testing.expect(@abs(value - exp) <= tolerance);
        }
    }
}

test "MHW reference data all bands" {
    const input = testdata.testInput();
    const sine = testdata.test1InputSine();
    const mixed = testdata.test2InputMixed();

    try checkSeries("HIGH", .{ .band = .high }, &input, &testdata.expectedHIGH());
    try checkSeries("MID", .{ .band = .mid }, &input, &testdata.expectedMID());
    try checkSeries("LOW", .{ .band = .low }, &input, &testdata.expectedLOW());
    try checkSeries("P8", .{ .band = .custom, .period = 8.0 }, &input, &testdata.expectedP8());
    try checkSeries("P20", .{ .band = .custom, .period = 20.0 }, &input, &testdata.expectedP20());
    try checkSeries("P32", .{ .band = .custom, .period = 32.0 }, &input, &testdata.expectedP32());
    try checkSeries("D2_0", .{ .band = .custom, .dilation = 2.0 }, &input, &testdata.expectedD2_0());
    try checkSeries("D8_0", .{ .band = .custom, .dilation = 8.0 }, &input, &testdata.expectedD8_0());

    try checkSeries("TEST1_MID", .{ .band = .mid }, &sine, &testdata.test1ExpectedMID());

    try checkSeries("TEST2_HIGH", .{ .band = .high }, &mixed, &testdata.test2ExpectedHIGH());
    try checkSeries("TEST2_MID", .{ .band = .mid }, &mixed, &testdata.test2ExpectedMID());
    try checkSeries("TEST2_LOW", .{ .band = .low }, &mixed, &testdata.test2ExpectedLOW());
}

test "MHW metadata default" {
    const allocator = testing.allocator;

    var mhw = try MexicanHatWavelet.init(allocator, .{});
    defer mhw.deinit();
    mhw.fixSlices();

    var meta: Metadata = undefined;
    mhw.getMetadata(&meta);

    try testing.expectEqual(Identifier.mexican_hat_wavelet, meta.identifier);
    try testing.expectEqualStrings("mhw(mid)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "MHW mnemonics" {
    const allocator = testing.allocator;

    var high = try MexicanHatWavelet.init(allocator, .{ .band = .high });
    defer high.deinit();
    try testing.expectEqualStrings("mhw(high)", high.mnemonic());

    var low = try MexicanHatWavelet.init(allocator, .{ .band = .low });
    defer low.deinit();
    try testing.expectEqualStrings("mhw(low)", low.mnemonic());

    var d2 = try MexicanHatWavelet.init(allocator, .{ .band = .custom, .dilation = 2.0 });
    defer d2.deinit();
    try testing.expectEqualStrings("mhw(d2.00)", d2.mnemonic());

    var p20 = try MexicanHatWavelet.init(allocator, .{ .band = .custom, .period = 20.0 });
    defer p20.deinit();
    try testing.expectEqualStrings("mhw(p20.00)", p20.mnemonic());
}

test "MHW invalid params" {
    const allocator = testing.allocator;

    const r1 = MexicanHatWavelet.init(allocator, .{ .band = .custom });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = MexicanHatWavelet.init(allocator, .{ .band = .custom, .dilation = 2.0, .period = 20.0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = MexicanHatWavelet.init(allocator, .{ .band = .custom, .period = 2.0 });
    try testing.expect(if (r3) |_| false else |_| true);

    const r4 = MexicanHatWavelet.init(allocator, .{ .band = .custom, .dilation = -1.0 });
    try testing.expect(if (r4) |_| false else |_| true);
}
