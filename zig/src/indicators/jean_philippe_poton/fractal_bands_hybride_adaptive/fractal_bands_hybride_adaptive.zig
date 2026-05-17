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

/// Enumerates the outputs of the fractal bands hybride adaptive indicator.
pub const FractalBandsHybrideAdaptiveOutput = enum(u8) {
    /// The FRASMA2 center line.
    frasma2 = 0,
    /// The upper band.
    upper = 1,
    /// The lower band.
    lower = 2,
    /// The lower/upper band pair.
    band = 3,
};

/// Parameters to create an instance of the fractal bands hybride adaptive indicator.
pub const FractalBandsHybrideAdaptiveParams = struct {
    /// The lookback period for FGDI computation. Must be > 1.
    period: usize,
    /// Fallback SMA period when CyclePeriod is unavailable. Must be > 0.
    normal_speed_fallback: usize,
    /// Band width multiplier raised to power H. Must be > 0.
    alpha: f64,
    /// Nyquist multiplier applied to the estimated cycle period. Must be > 0.
    nyquist: f64,
    /// High-pass filter alpha for Ehlers CyclePeriod. Must be between 0 and 1.
    alpha_hp: f64,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Fractal Bands Hybride Adaptive indicator.
pub const FractalBandsHybrideAdaptive = struct {
    line: LineIndicator,
    window: []f64,
    closes: std.ArrayList(f64),
    period: usize,
    window_size: usize,
    normal_speed_fallback: usize,
    alpha: f64,
    nyquist: f64,
    alpha_hp: f64,
    window_count: usize,
    primed: bool,
    log_denom: f64,
    ln2: f64,
    inv_period_sq: f64,
    // Ehlers CyclePeriod buffers.
    smooth_buf: std.ArrayList(f64),
    cycle_buf: std.ArrayList(f64),
    q1_buf: std.ArrayList(f64),
    i1_buf: std.ArrayList(f64),
    dp_buf: std.ArrayList(f64),
    inst_period_buf: std.ArrayList(f64),
    // Last computed values.
    frasma2: f64,
    upper_band: f64,
    lower_band: f64,
    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: FractalBandsHybrideAdaptiveParams) !FractalBandsHybrideAdaptive {
        if (params.period < 2) {
            return error.InvalidPeriod;
        }
        if (params.normal_speed_fallback < 1) {
            return error.InvalidNormalSpeedFallback;
        }
        if (params.alpha <= 0.0) {
            return error.InvalidAlpha;
        }
        if (params.nyquist <= 0.0) {
            return error.InvalidNyquist;
        }
        if (params.alpha_hp <= 0.0 or params.alpha_hp >= 1.0) {
            return error.InvalidAlphaHP;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "fbanha({d},{d},{d},{d},{d}{s})", .{
            params.period, params.normal_speed_fallback, params.alpha, params.nyquist, params.alpha_hp, triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Fractal bands hybride adaptive {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window_size = params.period + 1;
        const window = try allocator.alloc(f64, window_size);
        @memset(window, 0.0);

        const period_f: f64 = @floatFromInt(params.period);
        const period_minus_1_f: f64 = @floatFromInt(params.period - 1);

        var closes: std.ArrayList(f64) = .empty;
        try closes.ensureTotalCapacity(allocator, 256);

        var smooth_buf: std.ArrayList(f64) = .empty;
        try smooth_buf.ensureTotalCapacity(allocator, 256);
        var cycle_buf: std.ArrayList(f64) = .empty;
        try cycle_buf.ensureTotalCapacity(allocator, 256);
        var q1_buf: std.ArrayList(f64) = .empty;
        try q1_buf.ensureTotalCapacity(allocator, 256);
        var i1_buf: std.ArrayList(f64) = .empty;
        try i1_buf.ensureTotalCapacity(allocator, 256);
        var dp_buf: std.ArrayList(f64) = .empty;
        try dp_buf.ensureTotalCapacity(allocator, 256);
        var inst_period_buf: std.ArrayList(f64) = .empty;
        try inst_period_buf.ensureTotalCapacity(allocator, 256);

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
            .normal_speed_fallback = params.normal_speed_fallback,
            .alpha = params.alpha,
            .nyquist = params.nyquist,
            .alpha_hp = params.alpha_hp,
            .window_count = 0,
            .primed = false,
            .log_denom = @log(2.0 * period_minus_1_f),
            .ln2 = @log(2.0),
            .inv_period_sq = 1.0 / (period_f * period_f),
            .smooth_buf = smooth_buf,
            .cycle_buf = cycle_buf,
            .q1_buf = q1_buf,
            .i1_buf = i1_buf,
            .dp_buf = dp_buf,
            .inst_period_buf = inst_period_buf,
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

    pub fn deinit(self: *FractalBandsHybrideAdaptive) void {
        self.allocator.free(self.window);
        self.closes.deinit(self.allocator);
        self.smooth_buf.deinit(self.allocator);
        self.cycle_buf.deinit(self.allocator);
        self.q1_buf.deinit(self.allocator);
        self.i1_buf.deinit(self.allocator);
        self.dp_buf.deinit(self.allocator);
        self.inst_period_buf.deinit(self.allocator);
    }

    pub fn fixSlices(self: *FractalBandsHybrideAdaptive) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    fn getCyclePeriod(self: *FractalBandsHybrideAdaptive) f64 {
        const t = self.closes.items.len - 1;
        const prices = self.closes.items;

        // Extend buffers to index t.
        while (self.smooth_buf.items.len <= t) {
            self.smooth_buf.append(self.allocator, 0.0) catch return math.nan(f64);
        }
        while (self.cycle_buf.items.len <= t) {
            self.cycle_buf.append(self.allocator, 0.0) catch return math.nan(f64);
        }
        while (self.q1_buf.items.len <= t) {
            self.q1_buf.append(self.allocator, 0.0) catch return math.nan(f64);
        }
        while (self.i1_buf.items.len <= t) {
            self.i1_buf.append(self.allocator, 0.0) catch return math.nan(f64);
        }
        while (self.dp_buf.items.len <= t) {
            self.dp_buf.append(self.allocator, 0.0) catch return math.nan(f64);
        }
        while (self.inst_period_buf.items.len <= t) {
            self.inst_period_buf.append(self.allocator, 6.0) catch return math.nan(f64);
        }

        if (t < 6) {
            return math.nan(f64);
        }

        // 4-bar weighted smoother.
        self.smooth_buf.items[t] = (prices[t] + 2.0 * prices[t - 1] +
            2.0 * prices[t - 2] + prices[t - 3]) / 6.0;

        // High-pass filter.
        const alpha_hp = self.alpha_hp;
        const hp_coeff = (1.0 - 0.5 * alpha_hp) * (1.0 - 0.5 * alpha_hp);
        const one_minus_alpha = 1.0 - alpha_hp;

        self.cycle_buf.items[t] = hp_coeff * (self.smooth_buf.items[t] - 2.0 * self.smooth_buf.items[t - 1] + self.smooth_buf.items[t - 2]) +
            2.0 * one_minus_alpha * self.cycle_buf.items[t - 1] - one_minus_alpha * one_minus_alpha * self.cycle_buf.items[t - 2];

        // Quadrature component.
        self.q1_buf.items[t] = (0.0962 * self.cycle_buf.items[t] + 0.5769 * self.cycle_buf.items[t - 2] -
            0.5769 * self.cycle_buf.items[t - 4] - 0.0962 * self.cycle_buf.items[t - 6]) *
            (0.5 + 0.08 * self.inst_period_buf.items[t - 1]);

        // In-phase component.
        self.i1_buf.items[t] = self.cycle_buf.items[t - 3];

        // Smooth I and Q with EMA.
        if (t > 6) {
            self.i1_buf.items[t] = 0.15 * self.i1_buf.items[t] + 0.85 * self.i1_buf.items[t - 1];
            self.q1_buf.items[t] = 0.15 * self.q1_buf.items[t] + 0.85 * self.q1_buf.items[t - 1];
        }

        // Compute delta phase.
        var dp: f64 = undefined;
        if (@abs(self.i1_buf.items[t]) > 1e-10) {
            dp = math.atan(self.q1_buf.items[t] / self.i1_buf.items[t]);
        } else {
            dp = self.dp_buf.items[t - 1];
        }

        // Clamp delta phase.
        if (dp < 0.1) dp = 0.1;
        if (dp > 1.1) dp = 1.1;
        self.dp_buf.items[t] = dp;

        // Median delta phase over 5 bars.
        var median_dp: f64 = undefined;
        if (t >= 10) {
            var w = [5]f64{ self.dp_buf.items[t - 4], self.dp_buf.items[t - 3], self.dp_buf.items[t - 2], self.dp_buf.items[t - 1], self.dp_buf.items[t] };
            // Sort 5 elements.
            for (0..4) |ii| {
                for ((ii + 1)..5) |jj| {
                    if (w[jj] < w[ii]) {
                        const tmp = w[ii];
                        w[ii] = w[jj];
                        w[jj] = tmp;
                    }
                }
            }
            median_dp = w[2];
        } else {
            median_dp = dp;
        }

        // Instantaneous period.
        var dc: f64 = undefined;
        if (@abs(median_dp) > 1e-10) {
            dc = 6.2832 / median_dp + 0.5;
        } else {
            dc = self.inst_period_buf.items[t - 1];
        }

        // Clamp and smooth.
        if (dc < 6.0) dc = 6.0;
        if (dc > 50.0) dc = 50.0;
        self.inst_period_buf.items[t] = 0.33 * dc + 0.67 * self.inst_period_buf.items[t - 1];

        return self.inst_period_buf.items[t];
    }

    pub fn update(self: *FractalBandsHybrideAdaptive, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const period = self.period;
        const window_size = self.window_size;

        // Accumulate close history.
        self.closes.append(self.allocator, sample) catch return math.nan(f64);

        // Update Ehlers CyclePeriod.
        const cp = self.getCyclePeriod();

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
            var length: f64 = 0.0;
            var idx: usize = 1;
            while (idx < window_size) : (idx += 1) {
                const norm_cur = (self.window[idx] - price_min) / price_range;
                const norm_prev = (self.window[idx - 1] - price_min) / price_range;
                const diff = norm_cur - norm_prev;
                length += @sqrt(diff * diff + self.inv_period_sq);
            }
            fgdi = 1.0 + (@log(length) + self.ln2) / self.log_denom;
        }

        // Hurst exponent.
        var hurst = 2.0 - fgdi;
        if (hurst < 0.01) hurst = 0.01;

        const trail_dim = 1.0 / hurst;
        const beta = trail_dim / 2.0;

        // Adaptive normal_speed from CyclePeriod.
        var ns: f64 = undefined;
        if (math.isNan(cp) or cp < 1.0) {
            ns = @floatFromInt(self.normal_speed_fallback);
        } else {
            ns = cp * self.nyquist;
        }

        const speed_f = @round(ns * beta);
        var speed: usize = if (speed_f < 1.0) 1 else @intFromFloat(speed_f);
        if (speed < 1) speed = 1;

        // FRASMA2: SMA of close over 'speed' bars ending at current position.
        const n_closes = self.closes.items.len;
        if (speed > n_closes) {
            self.frasma2 = math.nan(f64);
            self.upper_band = math.nan(f64);
            self.lower_band = math.nan(f64);
            return math.nan(f64);
        }

        var sma_sum: f64 = 0.0;
        var si: usize = n_closes - speed;
        while (si < n_closes) : (si += 1) {
            sma_sum += self.closes.items[si];
        }
        const frasma2_val = sma_sum / @as(f64, @floatFromInt(speed));

        // Deviation over the last period closes.
        var sq_sum: f64 = 0.0;
        const dev_start = if (n_closes > period) n_closes - period else 0;
        var di: usize = dev_start;
        while (di < n_closes) : (di += 1) {
            const res = self.closes.items[di] - frasma2_val;
            sq_sum += res * res;
        }
        const period_f: f64 = @floatFromInt(period);
        const deviation = 2.0 * @sqrt(sq_sum / period_f);

        // Fractal bands.
        const band_mult = deviation * math.pow(f64, self.alpha, hurst);
        const upper_band_val = frasma2_val + band_mult;
        const lower_band_val = frasma2_val - band_mult;

        self.frasma2 = frasma2_val;
        self.upper_band = upper_band_val;
        self.lower_band = lower_band_val;

        return frasma2_val;
    }

    /// Updates and returns all three outputs: frasma2, upper_band, lower_band.
    pub fn updateAll(self: *FractalBandsHybrideAdaptive, sample: f64) struct { frasma2: f64, upper: f64, lower: f64 } {
        const frasma2_val = self.update(sample);
        return .{ .frasma2 = frasma2_val, .upper = self.upper_band, .lower = self.lower_band };
    }

    /// Returns the indicator metadata.
    pub fn getMetadata(self: *const FractalBandsHybrideAdaptive, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .fractal_bands_hybride_adaptive,
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

    /// Returns true if the indicator is primed.
    pub fn isPrimed(self: *const FractalBandsHybrideAdaptive) bool {
        return self.primed;
    }

    pub fn updateScalar(self: *FractalBandsHybrideAdaptive, sample: *const Scalar) OutputArray {
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

    pub fn updateBar(self: *FractalBandsHybrideAdaptive, sample: *const Bar) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractBar(sample) });
    }

    pub fn updateQuote(self: *FractalBandsHybrideAdaptive, sample: *const Quote) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractQuote(sample) });
    }

    pub fn updateTrade(self: *FractalBandsHybrideAdaptive, sample: *const Trade) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.extractTrade(sample) });
    }

    pub fn indicator(self: *FractalBandsHybrideAdaptive) indicator_mod.Indicator {
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
        const self: *FractalBandsHybrideAdaptive = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const FractalBandsHybrideAdaptive = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *FractalBandsHybrideAdaptive = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *FractalBandsHybrideAdaptive = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *FractalBandsHybrideAdaptive = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *FractalBandsHybrideAdaptive = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

const testing = std.testing;
const testdata = @import("testdata.zig");

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    if (math.isNan(a) and math.isNan(b)) return true;
    if (math.isNan(a) or math.isNan(b)) return false;
    return @abs(a - b) <= epsilon;
}

test "fractal_bands_hybride_adaptive P10_NY05_AHP007" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 10,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 0.5,
        .alpha_hp = 0.07,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP10NY05AHP007();
    const exp_upper = testdata.expectedUpperP10NY05AHP007();
    const exp_lower = testdata.expectedLowerP10NY05AHP007();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P10_NY05_AHP015" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 10,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 0.5,
        .alpha_hp = 0.15,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP10NY05AHP015();
    const exp_upper = testdata.expectedUpperP10NY05AHP015();
    const exp_lower = testdata.expectedLowerP10NY05AHP015();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P10_NY10_AHP007" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 10,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 1.0,
        .alpha_hp = 0.07,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP10NY10AHP007();
    const exp_upper = testdata.expectedUpperP10NY10AHP007();
    const exp_lower = testdata.expectedLowerP10NY10AHP007();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P10_NY10_AHP015" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 10,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 1.0,
        .alpha_hp = 0.15,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP10NY10AHP015();
    const exp_upper = testdata.expectedUpperP10NY10AHP015();
    const exp_lower = testdata.expectedLowerP10NY10AHP015();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P20_NY05_AHP007" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 20,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 0.5,
        .alpha_hp = 0.07,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP20NY05AHP007();
    const exp_upper = testdata.expectedUpperP20NY05AHP007();
    const exp_lower = testdata.expectedLowerP20NY05AHP007();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P20_NY05_AHP015" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 20,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 0.5,
        .alpha_hp = 0.15,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP20NY05AHP015();
    const exp_upper = testdata.expectedUpperP20NY05AHP015();
    const exp_lower = testdata.expectedLowerP20NY05AHP015();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P20_NY10_AHP007" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 20,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 1.0,
        .alpha_hp = 0.07,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP20NY10AHP007();
    const exp_upper = testdata.expectedUpperP20NY10AHP007();
    const exp_lower = testdata.expectedLowerP20NY10AHP007();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P20_NY10_AHP015" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 20,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 1.0,
        .alpha_hp = 0.15,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP20NY10AHP015();
    const exp_upper = testdata.expectedUpperP20NY10AHP015();
    const exp_lower = testdata.expectedLowerP20NY10AHP015();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P30_NY05_AHP007" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 30,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 0.5,
        .alpha_hp = 0.07,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP30NY05AHP007();
    const exp_upper = testdata.expectedUpperP30NY05AHP007();
    const exp_lower = testdata.expectedLowerP30NY05AHP007();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P30_NY05_AHP015" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 30,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 0.5,
        .alpha_hp = 0.15,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP30NY05AHP015();
    const exp_upper = testdata.expectedUpperP30NY05AHP015();
    const exp_lower = testdata.expectedLowerP30NY05AHP015();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P30_NY10_AHP007" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 30,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 1.0,
        .alpha_hp = 0.07,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP30NY10AHP007();
    const exp_upper = testdata.expectedUpperP30NY10AHP007();
    const exp_lower = testdata.expectedLowerP30NY10AHP007();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P30_NY10_AHP015" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 30,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 1.0,
        .alpha_hp = 0.15,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP30NY10AHP015();
    const exp_upper = testdata.expectedUpperP30NY10AHP015();
    const exp_lower = testdata.expectedLowerP30NY10AHP015();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P50_NY05_AHP007" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 50,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 0.5,
        .alpha_hp = 0.07,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP50NY05AHP007();
    const exp_upper = testdata.expectedUpperP50NY05AHP007();
    const exp_lower = testdata.expectedLowerP50NY05AHP007();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P50_NY05_AHP015" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 50,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 0.5,
        .alpha_hp = 0.15,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP50NY05AHP015();
    const exp_upper = testdata.expectedUpperP50NY05AHP015();
    const exp_lower = testdata.expectedLowerP50NY05AHP015();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P50_NY10_AHP007" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 50,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 1.0,
        .alpha_hp = 0.07,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP50NY10AHP007();
    const exp_upper = testdata.expectedUpperP50NY10AHP007();
    const exp_lower = testdata.expectedLowerP50NY10AHP007();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive P50_NY10_AHP015" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 50,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 1.0,
        .alpha_hp = 0.15,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    const exp_frasma = testdata.expectedFrasmaP50NY10AHP015();
    const exp_upper = testdata.expectedUpperP50NY10AHP015();
    const exp_lower = testdata.expectedLowerP50NY10AHP015();

    const epsilon = 2e-13;
    for (0..252) |i| {
        const result = ind.updateAll(input[i]);
        try testing.expect(almostEqual(result.frasma2, exp_frasma[i], epsilon));
        try testing.expect(almostEqual(result.upper, exp_upper[i], epsilon));
        try testing.expect(almostEqual(result.lower, exp_lower[i], epsilon));
    }
}

test "fractal_bands_hybride_adaptive is_primed" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 30,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 0.5,
        .alpha_hp = 0.07,
    });
    defer ind.deinit();

    const input = testdata.testInput();
    for (0..30) |i| {
        _ = ind.update(input[i]);
        try testing.expect(!ind.isPrimed());
    }
    _ = ind.update(input[30]);
    try testing.expect(ind.isPrimed());
}

test "fractal_bands_hybride_adaptive nan_passthrough" {
    const allocator = testing.allocator;
    var ind = try FractalBandsHybrideAdaptive.init(allocator, .{
        .period = 5,
        .normal_speed_fallback = 30,
        .alpha = 2.0,
        .nyquist = 0.5,
        .alpha_hp = 0.07,
    });
    defer ind.deinit();

    const result = ind.updateAll(math.nan(f64));
    try testing.expect(math.isNan(result.frasma2));
    try testing.expect(math.isNan(result.upper));
    try testing.expect(math.isNan(result.lower));
}
