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

/// Enumerates the outputs of the fractional bands indicator.
pub const FractionalBandsOutput = enum(u8) {
    /// The FRASMA2 center line.
    frasma2 = 0,
    /// The upper band.
    upper = 1,
    /// The lower band.
    lower = 2,
    /// The lower/upper band pair.
    band = 3,
};

/// Parameters to create an instance of the fractional bands indicator.
pub const FractionalBandsParams = struct {
    /// The lookback period for FGDI computation. Must be > 1.
    period: usize,
    /// Price-to-working-space multiplier. Must be > 0.
    price_scale: f64,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Fractional Bands indicator.
///
/// Fractal-adaptive moving average with FBM-scaled volatility bands.
/// Uses fractional Brownian motion power law: band_width = 2 * deviation^(2*H)
/// where H is the Hurst exponent derived from the Fractal Graph Dimension Index.
///
/// The indicator is not primed during the first `period` updates.
pub const FractionalBands = struct {
    line: LineIndicator,
    window: []f64,
    closes: std.ArrayList(f64),
    period: usize,
    window_size: usize,
    price_scale: f64,
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

    pub fn init(allocator: std.mem.Allocator, params: FractionalBandsParams) !FractionalBands {
        if (params.period < 2) {
            return error.InvalidPeriod;
        }
        if (params.price_scale <= 0.0) {
            return error.InvalidPriceScale;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "fctban({d},{d}{s})", .{ params.period, params.price_scale, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Fractional bands {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window_size = params.period + 1;
        const window = try allocator.alloc(f64, window_size);
        @memset(window, 0.0);

        const period_f: f64 = @floatFromInt(params.period);
        const period_minus_1: usize = params.period - 1;
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
            .window_size = window_size,
            .price_scale = params.price_scale,
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

    pub fn deinit(self: *FractionalBands) void {
        self.allocator.free(self.window);
        self.closes.deinit(self.allocator);
    }

    pub fn fixSlices(self: *FractionalBands) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *FractionalBands, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const period = self.period;
        const window_size = self.window_size;
        const p = self.price_scale;

        // Accumulate close history.
        self.closes.append(self.allocator, sample) catch return math.nan(f64);

        // Fill the FGDI window (period+1 elements).
        if (self.window_count < window_size) {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if (self.window_count < window_size) {
                return math.nan(f64);
            }

            self.primed = true;
        } else {
            var i: usize = 0;
            while (i < window_size - 1) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }
            self.window[window_size - 1] = sample;
        }

        // FGDI computation over period+1 points.
        var price_max = self.window[0];
        var price_min = self.window[0];

        var k: usize = 1;
        while (k < window_size) : (k += 1) {
            if (self.window[k] > price_max) price_max = self.window[k];
            if (self.window[k] < price_min) price_min = self.window[k];
        }

        const price_range = price_max - price_min;

        var fgdi: f64 = undefined;

        if (price_range < 1e-10) {
            fgdi = 1.0;
        } else {
            const inv_range = 1.0 / price_range;
            var prev_norm = (self.window[0] - price_min) * inv_range;
            var length: f64 = 0.0;

            var ii: usize = 1;
            while (ii < period) : (ii += 1) { // period-1 segments
                const cur_norm = (self.window[ii] - price_min) * inv_range;
                const diff = cur_norm - prev_norm;
                length += @sqrt(diff * diff + self.inv_period_sq);
                prev_norm = cur_norm;
            }

            if (length > 0.0) {
                fgdi = 1.0 + (@log(length) + self.ln2) / self.log_denom;
            } else {
                fgdi = 1.0;
            }
        }

        // Hurst exponent and adaptive speed.
        var hurst = 2.0 - fgdi;
        if (hurst < 0.01) {
            hurst = 0.01;
        }

        const trail_dim = 1.0 / hurst;
        const beta = trail_dim / 2.0;
        const period_f: f64 = @floatFromInt(period);
        const speed_f = @round(period_f * beta);
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

        // Deviation in scaled space over last *period* closes.
        const dev_start = n_closes - period;
        const frasma2_scaled = p * frasma2_val;
        var sq_sum: f64 = 0.0;

        k = dev_start;
        while (k < n_closes) : (k += 1) {
            const res = p * self.closes.items[k] - frasma2_scaled;
            sq_sum += res * res;
        }

        const deviation = @sqrt(sq_sum / period_f);

        // FBM band offset: 2 * sigma^(2H).
        const two_h = 2.0 * hurst;
        const band_offset = 2.0 * math.pow(f64, deviation, two_h);
        const ub = (frasma2_scaled + band_offset) / p;
        const lb = (frasma2_scaled - band_offset) / p;

        self.frasma2 = frasma2_val;
        self.upper_band = ub;
        self.lower_band = lb;

        return frasma2_val;
    }

    pub fn updateAll(self: *FractionalBands, sample: f64) struct { frasma2: f64, upper_band: f64, lower_band: f64 } {
        const frasma2_val = self.update(sample);
        return .{ .frasma2 = frasma2_val, .upper_band = self.upper_band, .lower_band = self.lower_band };
    }

    pub fn isPrimed(self: *const FractionalBands) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const FractionalBands, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .fractional_bands,
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

    pub fn updateScalar(self: *FractionalBands, sample: *const Scalar) OutputArray {
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

    pub fn updateBar(self: *FractionalBands, sample: *const Bar) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractBar(sample) });
    }

    pub fn updateQuote(self: *FractionalBands, sample: *const Quote) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractQuote(sample) });
    }

    pub fn updateTrade(self: *FractionalBands, sample: *const Trade) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractTrade(sample) });
    }

    pub fn indicator(self: *FractionalBands) indicator_mod.Indicator {
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
        const self: *FractionalBands = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const FractionalBands = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *FractionalBands = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *FractionalBands = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *FractionalBands = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *FractionalBands = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidPeriod,
        InvalidPriceScale,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn createFctban(allocator: std.mem.Allocator, period: usize, price_scale: f64) !FractionalBands {
    var ind = try FractionalBands.init(allocator, .{ .period = period, .price_scale = price_scale });
    ind.fixSlices();
    return ind;
}

fn checkValue(exp: f64, act: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
    } else {
        try testing.expect(@abs(exp - act) < 1e-11);
    }
}

test "fctban P5_S1" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P5S1();
    const exp_upper = testdata.expectedUpperP5S1();
    const exp_lower = testdata.expectedLowerP5S1();
    var ind = try createFctban(testing.allocator, 5, 1.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fctban P10_S1" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P10S1();
    const exp_upper = testdata.expectedUpperP10S1();
    const exp_lower = testdata.expectedLowerP10S1();
    var ind = try createFctban(testing.allocator, 10, 1.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fctban P20_S1" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P20S1();
    const exp_upper = testdata.expectedUpperP20S1();
    const exp_lower = testdata.expectedLowerP20S1();
    var ind = try createFctban(testing.allocator, 20, 1.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fctban P30_S1" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P30S1();
    const exp_upper = testdata.expectedUpperP30S1();
    const exp_lower = testdata.expectedLowerP30S1();
    var ind = try createFctban(testing.allocator, 30, 1.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fctban P50_S1" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P50S1();
    const exp_upper = testdata.expectedUpperP50S1();
    const exp_lower = testdata.expectedLowerP50S1();
    var ind = try createFctban(testing.allocator, 50, 1.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fctban P80_S1" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P80S1();
    const exp_upper = testdata.expectedUpperP80S1();
    const exp_lower = testdata.expectedLowerP80S1();
    var ind = try createFctban(testing.allocator, 80, 1.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fctban P30_S100" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P30S100();
    const exp_upper = testdata.expectedUpperP30S100();
    const exp_lower = testdata.expectedLowerP30S100();
    var ind = try createFctban(testing.allocator, 30, 100.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fctban P30_S10000" {
    const input = testdata.testInput();
    const exp_frasma2 = testdata.expectedFrasma2P30S10000();
    const exp_upper = testdata.expectedUpperP30S10000();
    const exp_lower = testdata.expectedLowerP30S10000();
    var ind = try createFctban(testing.allocator, 30, 10000.0);
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try checkValue(exp_frasma2[i], result.frasma2);
        try checkValue(exp_upper[i], result.upper_band);
        try checkValue(exp_lower[i], result.lower_band);
    }
}

test "fctban is primed" {
    const input = testdata.testInput();
    var ind = try createFctban(testing.allocator, 30, 1.0);
    defer ind.deinit();

    for (0..30) |i| {
        _ = ind.update(input[i]);
        try testing.expect(!ind.isPrimed());
    }
    _ = ind.update(input[30]);
    try testing.expect(ind.isPrimed());
}

test "fctban nan passthrough" {
    var ind = try createFctban(testing.allocator, 5, 1.0);
    defer ind.deinit();
    const result = ind.updateAll(math.nan(f64));
    try testing.expect(math.isNan(result.frasma2));
    try testing.expect(math.isNan(result.upper_band));
    try testing.expect(math.isNan(result.lower_band));
}

test "fctban invalid period" {
    const result = FractionalBands.init(testing.allocator, .{ .period = 1, .price_scale = 1.0 });
    try testing.expectError(error.InvalidPeriod, result);
}

test "fctban invalid price scale" {
    const result = FractionalBands.init(testing.allocator, .{ .period = 30, .price_scale = 0.0 });
    try testing.expectError(error.InvalidPriceScale, result);
}
