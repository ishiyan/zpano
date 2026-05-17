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

/// Enumerates the outputs of the fractal bands indicator.
pub const FractalBandsOutput = enum(u8) {
    /// The FRASMA2 center line.
    frasma2 = 0,
    /// The upper band.
    upper = 1,
    /// The lower band.
    lower = 2,
    /// The lower/upper band pair.
    band = 3,
};

/// Parameters to create an instance of the fractal bands indicator.
pub const FractalBandsParams = struct {
    /// The lookback period for FGDI computation. Must be > 1.
    period: usize,
    /// Base SMA period before fractal adaptation. Must be > 0.
    normal_speed: usize,
    /// Band width multiplier raised to power H. Must be > 0.
    alpha: f64,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Fractal Bands indicator.
///
/// FRASMA2 center line with upper/lower bands scaled by alpha^H where H is
/// the local Hurst exponent estimated from the Fractal Graph Dimension Index.
///
/// The indicator is not primed during the first `period - 1` updates.
pub const FractalBands = struct {
    line: LineIndicator,
    window: []f64,
    closes: std.ArrayList(f64),
    period: usize,
    period_minus_1: usize,
    normal_speed: usize,
    alpha: f64,
    window_count: usize,
    primed: bool,
    log_denom: f64,
    ln2: f64,
    inv_period_sq: f64,
    frasma2: f64,
    upper_band: f64,
    lower_band: f64,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: FractalBandsParams) !FractalBands {
        if (params.period < 2) {
            return error.InvalidPeriod;
        }
        if (params.normal_speed < 1) {
            return error.InvalidNormalSpeed;
        }
        if (params.alpha <= 0.0) {
            return error.InvalidAlpha;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "fban({d},{d},{d}{s})", .{ params.period, params.normal_speed, params.alpha, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Fractal bands {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window = try allocator.alloc(f64, params.period);
        @memset(window, 0.0);

        const period_f: f64 = @floatFromInt(params.period);
        const period_minus_1 = params.period - 1;
        const period_minus_1_f: f64 = @floatFromInt(period_minus_1);

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
            .period_minus_1 = period_minus_1,
            .normal_speed = params.normal_speed,
            .alpha = params.alpha,
            .window_count = 0,
            .primed = false,
            .log_denom = @log(2.0 * period_minus_1_f),
            .ln2 = @log(2.0),
            .inv_period_sq = 1.0 / (period_f * period_f),
            .frasma2 = math.nan(f64),
            .upper_band = math.nan(f64),
            .lower_band = math.nan(f64),
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *FractalBands) void {
        self.allocator.free(self.window);
        self.closes.deinit(self.allocator);
    }

    pub fn fixSlices(self: *FractalBands) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *FractalBands, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const period = self.period;
        const period_minus_1 = self.period_minus_1;

        // Accumulate close history for SMA computation.
        self.closes.append(self.allocator, sample) catch return math.nan(f64);

        // Fill the FGDI window.
        if (self.window_count < period) {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if (self.window_count < period) {
                return math.nan(f64);
            }

            self.primed = true;
        } else {
            var i: usize = 0;
            while (i < period_minus_1) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }
            self.window[period_minus_1] = sample;
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

        var fgdi: f64 = undefined;

        if (price_range <= 0.0) {
            fgdi = 0.0;
        } else {
            // Compute normalized path length: period points, period-1 segments.
            var prior_norm = (self.window[0] - price_min) / price_range;
            var length: f64 = 0.0;

            k = 1;
            while (k < period) : (k += 1) {
                const curr_norm = (self.window[k] - price_min) / price_range;
                const diff = curr_norm - prior_norm;
                length += @sqrt(diff * diff + self.inv_period_sq);
                prior_norm = curr_norm;
            }

            if (length > 0.0) {
                fgdi = 1.0 + (@log(length) + self.ln2) / self.log_denom;
            } else {
                fgdi = 0.0;
            }
        }

        // Hurst exponent.
        var hurst = 2.0 - fgdi;
        if (hurst < 0.01) {
            hurst = 0.01;
        }

        const trail_dim = 1.0 / hurst;
        const beta = trail_dim / 2.0;
        const normal_speed_f: f64 = @floatFromInt(self.normal_speed);
        const speed_f = @round(normal_speed_f * beta);
        const speed: usize = if (speed_f < 1.0) 1 else @intFromFloat(speed_f);

        // FRASMA2: SMA of close over 'speed' bars ending at current position.
        const n_closes = self.closes.items.len;
        if (speed > n_closes) {
            self.frasma2 = math.nan(f64);
            self.upper_band = math.nan(f64);
            self.lower_band = math.nan(f64);
            return math.nan(f64);
        }

        var sma_sum: f64 = 0.0;
        var idx: usize = n_closes - speed;
        while (idx < n_closes) : (idx += 1) {
            sma_sum += self.closes.items[idx];
        }

        const speed_f2: f64 = @floatFromInt(speed);
        const frasma2_val = sma_sum / speed_f2;

        // Deviation over the FGDI lookback window (period bars).
        var sq_sum: f64 = 0.0;
        k = 0;
        while (k < period) : (k += 1) {
            const res = self.window[k] - frasma2_val;
            sq_sum += res * res;
        }

        const period_f: f64 = @floatFromInt(period);
        const deviation = 2.0 * @sqrt(sq_sum / period_f);

        // Fractal bands.
        const band_mult = deviation * math.pow(f64, self.alpha, hurst);
        const ub = frasma2_val + band_mult;
        const lb = frasma2_val - band_mult;

        self.frasma2 = frasma2_val;
        self.upper_band = ub;
        self.lower_band = lb;

        return frasma2_val;
    }

    pub fn updateAll(self: *FractalBands, sample: f64) struct { frasma2: f64, upper_band: f64, lower_band: f64 } {
        const frasma2_val = self.update(sample);
        return .{ .frasma2 = frasma2_val, .upper_band = self.upper_band, .lower_band = self.lower_band };
    }

    pub fn isPrimed(self: *const FractalBands) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const FractalBands, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .fractal_bands,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
                .{ .mnemonic = "upper", .description = "Upper Band" },
                .{ .mnemonic = "lower", .description = "Lower Band" },
                .{ .mnemonic = "band", .description = "Band" },
            },
        );
    }

    pub fn updateScalar(self: *FractalBands, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = sample.time, .value = value } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = self.upper_band } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = self.lower_band } });

        if (math.isNan(self.lower_band) or math.isNan(self.upper_band)) {
            out.append(.{ .band = Band.empty(sample.time) });
        } else {
            out.append(.{ .band = Band.new(sample.time, self.lower_band, self.upper_band) });
        }

        return out;
    }

    pub fn updateBar(self: *FractalBands, sample: *const Bar) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractBar(sample) });
    }

    pub fn updateQuote(self: *FractalBands, sample: *const Quote) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractQuote(sample) });
    }

    pub fn updateTrade(self: *FractalBands, sample: *const Trade) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractTrade(sample) });
    }

    pub fn indicator(self: *FractalBands) indicator_mod.Indicator {
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
        const self: *FractalBands = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const FractalBands = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *FractalBands = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *FractalBands = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *FractalBands = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *FractalBands = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidPeriod,
        InvalidNormalSpeed,
        InvalidAlpha,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn createFban(allocator: std.mem.Allocator, period: usize, normal_speed: usize, alpha: f64) !FractalBands {
    var ind = try FractalBands.init(allocator, .{ .period = period, .normal_speed = normal_speed, .alpha = alpha });
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

test "fban P10_NS20_A2" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P10Ns20A2();
    const exp_upper = testdata.expectedUpperP10Ns20A2();
    const exp_lower = testdata.expectedLowerP10Ns20A2();
    var ind = try createFban(testing.allocator, 10, 20, 2.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fban P20_NS20_A2" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P20Ns20A2();
    const exp_upper = testdata.expectedUpperP20Ns20A2();
    const exp_lower = testdata.expectedLowerP20Ns20A2();
    var ind = try createFban(testing.allocator, 20, 20, 2.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fban P30_NS20_A2" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P30Ns20A2();
    const exp_upper = testdata.expectedUpperP30Ns20A2();
    const exp_lower = testdata.expectedLowerP30Ns20A2();
    var ind = try createFban(testing.allocator, 30, 20, 2.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fban P50_NS20_A2" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P50Ns20A2();
    const exp_upper = testdata.expectedUpperP50Ns20A2();
    const exp_lower = testdata.expectedLowerP50Ns20A2();
    var ind = try createFban(testing.allocator, 50, 20, 2.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fban P30_NS10_A2" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P30Ns10A2();
    const exp_upper = testdata.expectedUpperP30Ns10A2();
    const exp_lower = testdata.expectedLowerP30Ns10A2();
    var ind = try createFban(testing.allocator, 30, 10, 2.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fban P30_NS40_A2" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P30Ns40A2();
    const exp_upper = testdata.expectedUpperP30Ns40A2();
    const exp_lower = testdata.expectedLowerP30Ns40A2();
    var ind = try createFban(testing.allocator, 30, 40, 2.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fban P30_NS20_A1" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P30Ns20A1();
    const exp_upper = testdata.expectedUpperP30Ns20A1();
    const exp_lower = testdata.expectedLowerP30Ns20A1();
    var ind = try createFban(testing.allocator, 30, 20, 1.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fban P30_NS20_A3" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P30Ns20A3();
    const exp_upper = testdata.expectedUpperP30Ns20A3();
    const exp_lower = testdata.expectedLowerP30Ns20A3();
    var ind = try createFban(testing.allocator, 30, 20, 3.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fban is primed" {
    const input = testdata.testInput();
    var ind = try createFban(testing.allocator, 30, 20, 2.0);
    defer ind.deinit();

    for (0..29) |i| {
        _ = ind.update(input[i]);
        try testing.expect(!ind.isPrimed());
    }
    _ = ind.update(input[29]);
    try testing.expect(ind.isPrimed());
}

test "fban nan passthrough" {
    var ind = try createFban(testing.allocator, 5, 20, 2.0);
    defer ind.deinit();
    const result = ind.updateAll(math.nan(f64));
    try testing.expect(math.isNan(result.frasma2));
    try testing.expect(math.isNan(result.upper_band));
    try testing.expect(math.isNan(result.lower_band));
}

test "fban invalid period" {
    const result = FractalBands.init(testing.allocator, .{ .period = 1, .normal_speed = 20, .alpha = 2.0 });
    try testing.expectError(error.InvalidPeriod, result);
}

test "fban invalid normal speed" {
    const result = FractalBands.init(testing.allocator, .{ .period = 30, .normal_speed = 0, .alpha = 2.0 });
    try testing.expectError(error.InvalidNormalSpeed, result);
}

test "fban invalid alpha" {
    const result = FractalBands.init(testing.allocator, .{ .period = 30, .normal_speed = 20, .alpha = 0.0 });
    try testing.expectError(error.InvalidAlpha, result);
}
