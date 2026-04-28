const std = @import("std");
const math = std.math;

const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;
const bar_component = @import("bar_component");
const quote_component = @import("quote_component");
const trade_component = @import("trade_component");

const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the fractal adaptive moving average.
pub const FractalAdaptiveMovingAverageOutput = enum(u8) {
    value = 1,
    fdim = 2,
};

/// Parameters to create a FractalAdaptiveMovingAverage.
pub const Params = struct {
    /// Length (window size). Must be >= 2. Default is 16.
    /// If odd, it is rounded up to the next even number.
    length: i32 = 16,
    /// Slowest smoothing factor, alpha_s in [0, 1]. Default is 0.01.
    slowest_smoothing_factor: f64 = 0.01,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's Fractal Adaptive Moving Average (FRAMA).
///
/// An EMA with the smoothing factor alpha changed with each new sample based on
/// the estimated fractal dimension of the price series.
///
/// Two outputs: the FRAMA value and the estimated fractal dimension.
pub const FractalAdaptiveMovingAverage = struct {
    allocator: std.mem.Allocator,
    alpha_slowest: f64,
    scaling_factor: f64,
    fractal_dimension: f64,
    value: f64,
    length: usize,
    length_min_one: usize,
    half_length: usize,
    window_count: usize,
    window_high: []f64,
    window_low: []f64,
    primed: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [256]u8,
    description_len: usize,
    mnemonic_fdim_buf: [128]u8,
    mnemonic_fdim_len: usize,
    description_fdim_buf: [256]u8,
    description_fdim_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: Params) !FractalAdaptiveMovingAverage {
        if (params.length < 2) {
            return error.InvalidLength;
        }

        if (params.slowest_smoothing_factor < 0.0 or params.slowest_smoothing_factor > 1.0) {
            return error.InvalidSmoothingFactor;
        }

        var length_i: i32 = params.length;
        if (@mod(length_i, 2) != 0) {
            length_i += 1;
        }
        const length: usize = @intCast(length_i);

        const window_high = try allocator.alloc(f64, length);
        @memset(window_high, 0.0);
        errdefer allocator.free(window_high);

        const window_low = try allocator.alloc(f64, length);
        @memset(window_low, 0.0);

        // Resolve component defaults. FRAMA defaults to ClosePrice (not MedianPrice).
        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Build component triple mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            params.bar_component orelse bar_component.default_bar_component,
            params.quote_component orelse quote_component.default_quote_component,
            params.trade_component orelse trade_component.default_trade_component,
        );

        // Build mnemonics: frama(16, 0.010) or frama(16, 0.010, hl/2)
        var mnemonic_buf: [128]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "frama({d}, {d:.3}{s})", .{ length_i, params.slowest_smoothing_factor, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [256]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Fractal adaptive moving average {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_fdim_buf: [128]u8 = undefined;
        const mn_fdim = std.fmt.bufPrint(&mnemonic_fdim_buf, "framaDim({d}, {d:.3}{s})", .{ length_i, params.slowest_smoothing_factor, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_fdim_len = mn_fdim.len;

        var description_fdim_buf: [256]u8 = undefined;
        const desc_fdim = std.fmt.bufPrint(&description_fdim_buf, "Fractal adaptive moving average {s}", .{mn_fdim}) catch
            return error.MnemonicTooLong;
        const description_fdim_len = desc_fdim.len;

        return .{
            .allocator = allocator,
            .alpha_slowest = params.slowest_smoothing_factor,
            .scaling_factor = @log(params.slowest_smoothing_factor),
            .fractal_dimension = math.nan(f64),
            .value = math.nan(f64),
            .length = length,
            .length_min_one = length - 1,
            .half_length = length / 2,
            .window_count = 0,
            .window_high = window_high,
            .window_low = window_low,
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .mnemonic_fdim_buf = mnemonic_fdim_buf,
            .mnemonic_fdim_len = mnemonic_fdim_len,
            .description_fdim_buf = description_fdim_buf,
            .description_fdim_len = description_fdim_len,
        };
    }

    pub fn deinit(self: *FractalAdaptiveMovingAverage) void {
        self.allocator.free(self.window_high);
        self.allocator.free(self.window_low);
    }

    pub fn fixSlices(self: *FractalAdaptiveMovingAverage) void {
        _ = self;
    }

    /// Update the FRAMA given the next sample (value, high, low).
    pub fn update(self: *FractalAdaptiveMovingAverage, sample: f64, sample_high: f64, sample_low: f64) f64 {
        if (math.isNan(sample_high) or math.isNan(sample_low) or math.isNan(sample)) {
            return math.nan(f64);
        }

        if (self.primed) {
            // Shift windows left.
            std.mem.copyForwards(f64, self.window_high[0..self.length_min_one], self.window_high[1..self.length]);
            std.mem.copyForwards(f64, self.window_low[0..self.length_min_one], self.window_low[1..self.length]);

            self.window_high[self.length_min_one] = sample_high;
            self.window_low[self.length_min_one] = sample_low;

            self.fractal_dimension = self.estimateFractalDimension();
            const alpha = self.estimateAlpha();
            self.value += (sample - self.value) * alpha;

            return self.value;
        } else {
            self.window_high[self.window_count] = sample_high;
            self.window_low[self.window_count] = sample_low;

            self.window_count += 1;
            if (self.window_count == self.length_min_one) {
                self.value = sample;
            } else if (self.window_count == self.length) {
                self.fractal_dimension = self.estimateFractalDimension();
                const alpha = self.estimateAlpha();
                self.value += (sample - self.value) * alpha;
                self.primed = true;

                return self.value;
            }
        }

        return math.nan(f64);
    }

    fn estimateFractalDimension(self: *const FractalAdaptiveMovingAverage) f64 {
        var min_low_half: f64 = math.floatMax(f64);
        var max_high_half: f64 = math.floatMin(f64);

        for (0..self.half_length) |i| {
            const l = self.window_low[i];
            if (min_low_half > l) min_low_half = l;

            const h = self.window_high[i];
            if (max_high_half < h) max_high_half = h;
        }

        const range_n1 = max_high_half - min_low_half;
        var min_low_full = min_low_half;
        var max_high_full = max_high_half;
        min_low_half = math.floatMax(f64);
        max_high_half = math.floatMin(f64);

        for (0..self.half_length) |j| {
            const i = j + self.half_length;
            const l = self.window_low[i];

            if (min_low_full > l) min_low_full = l;
            if (min_low_half > l) min_low_half = l;

            const h = self.window_high[i];
            if (max_high_full < h) max_high_full = h;
            if (max_high_half < h) max_high_half = h;
        }

        const range_n2 = max_high_half - min_low_half;
        const range_n3 = max_high_full - min_low_full;

        const half_len_f: f64 = @floatFromInt(self.half_length);
        const len_f: f64 = @floatFromInt(self.length);

        const fdim = (@log((range_n1 + range_n2) / half_len_f) -
            @log(range_n3 / len_f)) * math.log2e;

        return @min(@max(fdim, 1.0), 2.0);
    }

    fn estimateAlpha(self: *const FractalAdaptiveMovingAverage) f64 {
        const alpha = @exp(self.scaling_factor * (self.fractal_dimension - 1.0));
        return @min(@max(alpha, self.alpha_slowest), 1.0);
    }

    pub fn isPrimed(self: *const FractalAdaptiveMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const FractalAdaptiveMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .fractal_adaptive_moving_average,
            self.mnemonic_buf[0..self.mnemonic_len],
            self.description_buf[0..self.description_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mnemonic_buf[0..self.mnemonic_len], .description = self.description_buf[0..self.description_len] },
                .{ .mnemonic = self.mnemonic_fdim_buf[0..self.mnemonic_fdim_len], .description = self.description_fdim_buf[0..self.description_fdim_len] },
            },
        );
    }

    pub fn updateScalar(self: *FractalAdaptiveMovingAverage, sample: *const Scalar) OutputArray {
        const v = sample.value;
        return self.updateEntity(sample.time, v, v, v);
    }

    pub fn updateBar(self: *FractalAdaptiveMovingAverage, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateEntity(sample.time, v, sample.high, sample.low);
    }

    pub fn updateQuote(self: *FractalAdaptiveMovingAverage, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateEntity(sample.time, v, sample.ask_price, sample.bid_price);
    }

    pub fn updateTrade(self: *FractalAdaptiveMovingAverage, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateEntity(sample.time, v, v, v);
    }

    fn updateEntity(self: *FractalAdaptiveMovingAverage, time: i64, sample: f64, sample_high: f64, sample_low: f64) OutputArray {
        const frama = self.update(sample, sample_high, sample_low);
        var fdim = self.fractal_dimension;
        if (math.isNan(frama)) {
            fdim = math.nan(f64);
        }

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = frama } });
        out.append(.{ .scalar = .{ .time = time, .value = fdim } });
        return out;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *FractalAdaptiveMovingAverage) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(FractalAdaptiveMovingAverage);
};

// --- Tests ---
const testing = std.testing;

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

// 252-entry mid-price input data from test_FRAMA.xls.
const test_input_mid = [_]f64{
    92.00000,  93.17250,  95.31250,  94.84500,  94.40750,  94.11000,  93.50000,  91.73500,  90.95500,  91.68750,
    94.50000,  97.97000,  97.57750,  90.78250,  89.03250,  92.09500,  91.15500,  89.71750,  90.61000,  91.00000,
    88.92250,  87.51500,  86.43750,  83.89000,  83.00250,  82.81250,  82.84500,  86.73500,  86.86000,  87.54750,
    85.78000,  86.17250,  86.43750,  87.25000,  88.93750,  88.20500,  85.81250,  84.59500,  83.65750,  84.45500,
    83.50000,  86.78250,  88.17250,  89.26500,  90.86000,  90.78250,  91.86000,  90.36000,  89.86000,  90.92250,
    89.50000,  87.67250,  86.50000,  84.28250,  82.90750,  84.25000,  85.68750,  86.61000,  88.28250,  89.53250,
    89.50000,  88.09500,  90.62500,  92.23500,  91.67250,  92.59250,  93.01500,  91.17250,  90.98500,  90.37750,
    88.25000,  86.90750,  84.09250,  83.18750,  84.25250,  97.86000,  99.87500,  103.26500, 105.93750, 103.50000,
    103.11000, 103.61000, 104.64000, 106.81500, 104.95250, 105.50000, 107.14000, 109.73500, 109.84500, 110.98500,
    120.00000, 119.87500, 117.90750, 119.40750, 117.95250, 117.22000, 115.64250, 113.11000, 111.75000, 114.51750,
    114.74500, 115.47000, 112.53000, 112.03000, 113.43500, 114.22000, 119.59500, 117.96500, 118.71500, 115.03000,
    114.53000, 115.00000, 116.53000, 120.18500, 120.50000, 120.59500, 124.18500, 125.37500, 122.97000, 123.00000,
    124.43500, 123.44000, 124.03000, 128.18500, 129.65500, 130.87500, 132.34500, 132.06500, 133.81500, 135.66000,
    137.03500, 137.47000, 137.34500, 136.31500, 136.44000, 136.28500, 129.09500, 128.31000, 126.00000, 124.03000,
    123.93500, 125.03000, 127.25000, 125.62000, 125.53000, 123.90500, 120.65500, 119.96500, 120.78000, 124.00000,
    122.78000, 120.72000, 121.78000, 122.40500, 123.25000, 126.18500, 127.56000, 126.56500, 123.06000, 122.71500,
    123.59000, 122.31000, 122.46500, 123.96500, 123.97000, 124.15500, 124.43500, 127.00000, 125.50000, 128.87500,
    130.53500, 132.31500, 134.06500, 136.03500, 133.78000, 132.75000, 133.47000, 130.97000, 127.59500, 128.44000,
    127.94000, 125.81000, 124.62500, 122.72000, 124.09000, 123.22000, 121.40500, 120.93500, 118.28000, 118.37500,
    121.15500, 120.90500, 117.12500, 113.06000, 114.90500, 112.43500, 107.93500, 105.97000, 106.37000, 106.84500,
    106.97000, 110.03000, 91.00000,  93.56000,  93.62000,  95.31000,  94.18500,  94.78000,  97.62500,  97.59000,
    95.25000,  94.72000,  92.22000,  91.56500,  92.22000,  93.81000,  95.59000,  96.18500,  94.62500,  95.12000,
    94.00000,  93.74500,  95.90500,  101.74500, 106.44000, 107.93500, 103.40500, 105.06000, 104.15500, 103.31000,
    103.34500, 104.84000, 110.40500, 114.50000, 117.31500, 118.25000, 117.18500, 109.75000, 109.65500, 108.53000,
    106.22000, 107.72000, 109.84000, 109.09500, 109.09000, 109.15500, 109.31500, 109.06000, 109.90500, 109.62500,
    109.53000, 108.06000,
};

// 252-entry high price input data from test_FRAMA.xls.
const test_input_high = [_]f64{
    93.2500,  94.9400,  96.3750,  96.1900,  96.0000,  94.7200,  95.0000,  93.7200,  92.4700,  92.7500,
    96.2500,  99.6250,  99.1250,  92.7500,  91.3150,  93.2500,  93.4050,  90.6550,  91.9700,  92.2500,
    90.3450,  88.5000,  88.2500,  85.5000,  84.4400,  84.7500,  84.4400,  89.4050,  88.1250,  89.1250,
    87.1550,  87.2500,  87.3750,  88.9700,  90.0000,  89.8450,  86.9700,  85.9400,  84.7500,  85.4700,
    84.4700,  88.5000,  89.4700,  90.0000,  92.4400,  91.4400,  92.9700,  91.7200,  91.1550,  91.7500,
    90.0000,  88.8750,  89.0000,  85.2500,  83.8150,  85.2500,  86.6250,  87.9400,  89.3750,  90.6250,
    90.7500,  88.8450,  91.9700,  93.3750,  93.8150,  94.0300,  94.0300,  91.8150,  92.0000,  91.9400,
    89.7500,  88.7500,  86.1550,  84.8750,  85.9400,  99.3750,  103.2800, 105.3750, 107.6250, 105.2500,
    104.5000, 105.5000, 106.1250, 107.9400, 106.2500, 107.0000, 108.7500, 110.9400, 110.9400, 114.2200,
    123.0000, 121.7500, 119.8150, 120.3150, 119.3750, 118.1900, 116.6900, 115.3450, 113.0000, 118.3150,
    116.8700, 116.7500, 113.8700, 114.6200, 115.3100, 116.0000, 121.6900, 119.8700, 120.8700, 116.7500,
    116.5000, 116.0000, 118.3100, 121.5000, 122.0000, 121.4400, 125.7500, 127.7500, 124.1900, 124.4400,
    125.7500, 124.6900, 125.3100, 132.0000, 131.3100, 132.2500, 133.8800, 133.5000, 135.5000, 137.4400,
    138.6900, 139.1900, 138.5000, 138.1300, 137.5000, 138.8800, 132.1300, 129.7500, 128.5000, 125.4400,
    125.1200, 126.5000, 128.6900, 126.6200, 126.6900, 126.0000, 123.1200, 121.8700, 124.0000, 127.0000,
    124.4400, 122.5000, 123.7500, 123.8100, 124.5000, 127.8700, 128.5600, 129.6300, 124.8700, 124.3700,
    124.8700, 123.6200, 124.0600, 125.8700, 125.1900, 125.6200, 126.0000, 128.5000, 126.7500, 129.7500,
    132.6900, 133.9400, 136.5000, 137.6900, 135.5600, 133.5600, 135.0000, 132.3800, 131.4400, 130.8800,
    129.6300, 127.2500, 127.8100, 125.0000, 126.8100, 124.7500, 122.8100, 122.2500, 121.0600, 120.0000,
    123.2500, 122.7500, 119.1900, 115.0600, 116.6900, 114.8700, 110.8700, 107.2500, 108.8700, 109.0000,
    108.5000, 113.0600, 93.0000,  94.6200,  95.1200,  96.0000,  95.5600,  95.3100,  99.0000,  98.8100,
    96.8100,  95.9400,  94.4400,  92.9400,  93.9400,  95.5000,  97.0600,  97.5000,  96.2500,  96.3700,
    95.0000,  94.8700,  98.2500,  105.1200, 108.4400, 109.8700, 105.0000, 106.0000, 104.9400, 104.5000,
    104.4400, 106.3100, 112.8700, 116.5000, 119.1900, 121.0000, 122.1200, 111.9400, 112.7500, 110.1900,
    107.9400, 109.6900, 111.0600, 110.4400, 110.1200, 110.3100, 110.4400, 110.0000, 110.7500, 110.5000,
    110.5000, 109.5000,
};

// 252-entry low price input data from test_FRAMA.xls.
const test_input_low = [_]f64{
    90.7500,  91.4050,  94.2500,  93.5000,  92.8150,  93.5000,  92.0000,  89.7500,  89.4400,  90.6250,
    92.7500,  96.3150,  96.0300,  88.8150,  86.7500,  90.9400,  88.9050,  88.7800,  89.2500,  89.7500,
    87.5000,  86.5300,  84.6250,  82.2800,  81.5650,  80.8750,  81.2500,  84.0650,  85.5950,  85.9700,
    84.4050,  85.0950,  85.5000,  85.5300,  87.8750,  86.5650,  84.6550,  83.2500,  82.5650,  83.4400,
    82.5300,  85.0650,  86.8750,  88.5300,  89.2800,  90.1250,  90.7500,  89.0000,  88.5650,  90.0950,
    89.0000,  86.4700,  84.0000,  83.3150,  82.0000,  83.2500,  84.7500,  85.2800,  87.1900,  88.4400,
    88.2500,  87.3450,  89.2800,  91.0950,  89.5300,  91.1550,  92.0000,  90.5300,  89.9700,  88.8150,
    86.7500,  85.0650,  82.0300,  81.5000,  82.5650,  96.3450,  96.4700,  101.1550, 104.2500, 101.7500,
    101.7200, 101.7200, 103.1550, 105.6900, 103.6550, 104.0000, 105.5300, 108.5300, 108.7500, 107.7500,
    117.0000, 118.0000, 116.0000, 118.5000, 116.5300, 116.2500, 114.5950, 110.8750, 110.5000, 110.7200,
    112.6200, 114.1900, 111.1900, 109.4400, 111.5600, 112.4400, 117.5000, 116.0600, 116.5600, 113.3100,
    112.5600, 114.0000, 114.7500, 118.8700, 119.0000, 119.7500, 122.6200, 123.0000, 121.7500, 121.5600,
    123.1200, 122.1900, 122.7500, 124.3700, 128.0000, 129.5000, 130.8100, 130.6300, 132.1300, 133.8800,
    135.3800, 135.7500, 136.1900, 134.5000, 135.3800, 133.6900, 126.0600, 126.8700, 123.5000, 122.6200,
    122.7500, 123.5600, 125.8100, 124.6200, 124.3700, 121.8100, 118.1900, 118.0600, 117.5600, 121.0000,
    121.1200, 118.9400, 119.8100, 121.0000, 122.0000, 124.5000, 126.5600, 123.5000, 121.2500, 121.0600,
    122.3100, 121.0000, 120.8700, 122.0600, 122.7500, 122.6900, 122.8700, 125.5000, 124.2500, 128.0000,
    128.3800, 130.6900, 131.6300, 134.3800, 132.0000, 131.9400, 131.9400, 129.5600, 123.7500, 126.0000,
    126.2500, 124.3700, 121.4400, 120.4400, 121.3700, 121.6900, 120.0000, 119.6200, 115.5000, 116.7500,
    119.0600, 119.0600, 115.0600, 111.0600, 113.1200, 110.0000, 105.0000, 104.6900, 103.8700, 104.6900,
    105.4400, 107.0000, 89.0000,  92.5000,  92.1200,  94.6200,  92.8100,  94.2500,  96.2500,  96.3700,
    93.6900,  93.5000,  90.0000,  90.1900,  90.5000,  92.1200,  94.1200,  94.8700,  93.0000,  93.8700,
    93.0000,  92.6200,  93.5600,  98.3700,  104.4400, 106.0000, 101.8100, 104.1200, 103.3700, 102.1200,
    102.2500, 103.3700, 107.9400, 112.5000, 115.4400, 115.5000, 112.2500, 107.5600, 106.5600, 106.8700,
    104.5000, 105.7500, 108.6200, 107.7500, 108.0600, 108.0000, 108.1900, 108.1200, 109.0600, 108.7500,
    108.5600, 106.6200,
};

// Expected FRAMA values, 252 entries.
const test_expected_frama = [_]f64{
    math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
    math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
    math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
    89.22671050342360,  89.33682876326700,  89.35856764336560,  89.43310239684000,  89.47082931633650,
    89.37869910788870,  89.12556219544840,  88.89115876150430,  88.31300745552690,  87.40507669494460,
    86.57356547394590,  85.89848822739120,  86.15180198300710,  86.34933228089390,  86.65306604987650,
    86.55106409845000,  86.53204808278100,  86.52914459592330,  86.54901457229990,  86.69578691869190,
    86.76484617172940,  86.71259620319170,  86.64367592990800,  86.57105142033130,  86.51958854599050,
    86.44721216106140,  86.45433109997260,  86.49247643474000,  86.56005691800830,  86.77979715068570,
    86.90670558143720,  87.04578122644150,  87.13883588353810,  87.34672192917600,  87.97454299697020,
    88.66811941238330,  88.52434793414250,  88.48482846945310,  88.42008690365100,  88.32432325619980,
    88.25354486073350,  88.14484283089590,  87.85455822478500,  88.11593829889930,  88.28362254494940,
    88.33390751615900,  88.32623104955050,  88.36446581314600,  88.52442732438340,  88.69739412610610,
    89.66302327692510,  90.85256886435120,  90.91704236139110,  90.93001715919770,  90.72490665226460,
    90.54537981954970,  90.49341244181010,  90.31305220222090,  89.91864626179560,  89.22400447234590,
    90.27011547796660,  91.89238069442760,  94.95304271855450,  97.09846388459310,  98.00008978496790,
    98.41485341671240,  98.87804648283650,  104.64000000000000, 105.72811350054100, 105.26403371285900,
    105.36386040471100, 105.84640410018900, 107.00074975605000, 107.84507711684500, 109.15061836503300,
    113.46379915259300, 115.83475327982200, 116.52354569655000, 117.93747690109700, 117.94854969872800,
    117.66052662433900, 116.86272486344900, 115.52953436643100, 115.34974336234200, 115.29526922539300,
    115.26068054267300, 115.27434242238600, 115.09869321289600, 114.95297492047600, 114.88089319121700,
    114.84474663070300, 114.90946935433000, 114.97072281745300, 115.04578341115000, 115.04536087712600,
    115.02593804158700, 115.02453605896500, 115.13603528837900, 115.52684975794600, 115.65874963847700,
    115.78967084561300, 116.22754967761900, 116.84985380134300, 117.44551886973900, 117.96470004242600,
    120.06903626149100, 121.21008540014300, 122.33645161525400, 127.37689813615500, 127.99975796005200,
    128.46741276146300, 129.10776056689700, 129.65064396790600, 130.26342435221600, 131.35952501733000,
    133.26586703797100, 134.34648967636400, 135.58893175985700, 135.99381098938800, 136.13880177237000,
    136.21505011519200, 135.83651324279900, 135.60807459099700, 135.47882165029800, 135.07860449875800,
    134.51275124440700, 133.86549636620300, 133.41394890024100, 125.62000000000000, 125.59633419039200,
    124.93313611371900, 123.94685405180400, 123.21926532001500, 122.75754883008000, 122.99272627813500,
    122.93125117031900, 122.54555605939100, 122.44596369450800, 122.44274244630400, 122.45951458647800,
    122.51543458717600, 122.59588129397600, 122.67354949331500, 122.68646914748500, 122.68774838354900,
    122.74954803256700, 122.71944122683400, 122.70282387051800, 122.76326904385600, 122.79621943896900,
    122.90238467427100, 123.01574410872400, 123.10313838895400, 123.14249285231400, 123.20547966041500,
    123.47894986686500, 124.00054570690400, 124.87676528249000, 128.62681459660400, 131.29664422670300,
    131.62199116417700, 133.14317462344700, 131.85148268159700, 131.64067292249000, 131.53846982622200,
    131.47806560020400, 131.36260843074200, 131.13168585701800, 130.79500213890900, 129.30384190341500,
    127.24799786761200, 126.63211575201500, 125.75552528396400, 123.30744191111000, 121.85785126900600,
    121.71821486177200, 121.59387984799400, 120.57291871850100, 117.38790612905300, 116.48251086688600,
    115.17150035054500, 114.13499298155200, 112.93872729544800, 111.87930160220700, 110.01049157382800,
    108.16524199695300, 108.50781213775200, 99.82043389495540,  96.27485062109970,  95.62284596868710,
    95.54965165331060,  95.27413209875520,  95.18959285515430,  95.55235701361380,  97.59000000000000,
    97.10158819524640,  96.65367530552770,  96.11681200042330,  95.56564380104830,  95.16052766940610,
    94.99699545329450,  95.07160396820560,  95.09863980555460,  95.07758677330600,  95.07947201820910,
    95.05868981255450,  95.02995933565580,  95.05770518819190,  95.70260675064800,  97.27204023791490,
    99.75062893645490,  100.60008190453100, 101.63678304891200, 102.21478920369900, 102.53486115334100,
    103.34500000000000, 103.81075056175000, 104.81523430115500, 106.21793872319500, 108.24424614042100,
    110.25613730939300, 111.72713168704800, 111.31632004813300, 110.98332479082400, 110.08358361255600,
    109.69338946921400, 109.60310110509600, 109.60920382381800, 109.59975576966400, 109.54836792354200,
    109.50921271876300, 109.48274942826400, 109.38946522506500, 109.47904564711100, 109.52555071439800,
    109.52711612931300, 109.16498860274600,
};

// Expected fractal dimension values, 252 entries.
const test_expected_fdim = [_]f64{
    math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),
    math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),
    math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),
    1.598901691679030, 1.621656447931540, 1.621656447931540, 1.612524280201100, 1.809194635440850,
    1.387319706108450, 1.433510159350320, 1.529737671684330, 1.468514565209500, 1.383541646349600,
    1.371094151691250, 1.371094151691250, 1.259406641789480, 1.277260493845600, 1.298012267717490,
    1.466219310044780, 1.649509949102230, 1.756366431089380, 1.779825403913370, 1.605739179393220,
    1.669764342126080, 1.630354403535170, 1.743748079381010, 1.807016088915350, 1.807016088915350,
    1.810175441119980, 1.836501267717120, 1.826812211548690, 1.806526976530780, 1.645776563426130,
    1.749431438179060, 1.775821529983230, 1.775821529983230, 1.558464698544080, 1.377767335711980,
    1.171152827083140, 1.420210308839580, 1.854737052791960, 1.906153398330910, 1.880077384928280,
    1.880077384928280, 1.686513295356410, 1.361619952167230, 1.107056103570650, 1.463371616646130,
    1.691815078151830, 1.746534223980870, 1.889518456584240, 1.691877704637670, 1.630040997367520,
    1.302854449444730, 1.224959951896460, 1.347837640323710, 1.359568840439050, 1.215178931103870,
    1.569714727022220, 1.922558772344390, 1.775050541483170, 1.628437539281390, 1.455763437050440,
    1.458367318817070, 1.386185145674240, 1.285022563289830, 1.354633021653340, 1.425628910546560,
    1.545306300369370, 1.524917878187410, 1.000000000000000, 1.150392531269620, 1.111526353076260,
    1.186801639893640, 1.282970225101590, 1.263728470502570, 1.263728470502570, 1.190564192051970,
    1.200303662027460, 1.216008115539910, 1.239228939755000, 1.154780047739040, 1.066251081838840,
    1.201515950578200, 1.201515950578200, 1.224727257646170, 1.661335170391540, 1.592029990883720,
    1.600820601168370, 1.592649571456720, 1.596896049728040, 1.661719699705190, 1.661719699705190,
    1.631032149417140, 1.932829994967900, 1.848977998086050, 1.848977998086050, 1.786169592166980,
    1.711899406890060, 1.633597275972030, 1.565199245899730, 1.555615850300350, 1.788193588780650,
    1.788193588780650, 1.641341888572470, 1.583648690375480, 1.505879832118480, 1.514662280333620,
    1.243904642244370, 1.235224866681230, 1.199278170274840, 1.032289537956510, 1.281591410368070,
    1.394374506090950, 1.391072825973270, 1.368089962796670, 1.416121498909740, 1.346133913992910,
    1.236900696360800, 1.295001190641800, 1.191314697601160, 1.126825971149440, 1.244089237249700,
    1.141355849245540, 1.637187430482630, 1.758912120860560, 1.935597966630190, 1.728232546671110,
    1.647160936023790, 1.582929512908910, 1.582929512908910, 1.000000000000000, 1.290060571579690,
    1.203293085091340, 1.318626731797110, 1.369099686573460, 1.361441827946200, 1.361441827946200,
    1.269560921322060, 1.379196970993940, 1.442875476601240, 1.552187464629540, 1.841211817056240,
    1.911807849588850, 1.898657743779690, 1.854225425568590, 1.737921496166320, 1.674182027088440,
    1.582170830109580, 1.582170830109580, 1.592512774331040, 1.659879167557740, 1.781874913568540,
    1.553583499247610, 1.565487746752300, 1.829432106782320, 1.892324482473850, 1.979547502083450,
    1.714082767929260, 1.614462026110310, 1.530088640192170, 1.236779256931340, 1.142796109680280,
    1.325012608562160, 1.042261224471580, 1.112967837396340, 1.652580079362500, 1.747888621351000,
    1.887525270741590, 1.845506980904780, 1.732519689135010, 1.698830465279440, 1.326437303409080,
    1.235593877743260, 1.488569067439900, 1.406429149758940, 1.242407755243840, 1.265908301287900,
    1.350932383579930, 1.407805930047770, 1.320594735235250, 1.186348768480190, 1.219061041496490,
    1.244790924007050, 1.421978096723610, 1.417064085114450, 1.396205365497800, 1.215187037194790,
    1.108444344884270, 1.367936486765890, 1.152171573538770, 1.123458372081530, 1.304894697621100,
    1.315426605047450, 1.347434689033220, 1.383392354387930, 1.413473602062410, 1.000000000000000,
    1.340214852627430, 1.362836554272960, 1.458450082116240, 1.458450082116240, 1.458450082116240,
    1.458450082116240, 1.450134813904230, 1.807354922057600, 1.676066769222260, 1.676066769222260,
    1.857759875494850, 1.830074998557690, 1.749415080999700, 1.507878520255750, 1.417577972069620,
    1.316836652604660, 1.316836652604660, 1.316836652604660, 1.319580340030940, 1.267125050525610,
    1.000000000000000, 1.253243902640980, 1.408611230119410, 1.419561476146430, 1.369251325009040,
    1.348322665371890, 1.336525471100320, 1.341196428592140, 1.349007628943650, 1.217818725101000,
    1.497854784405750, 1.669790510969210, 1.794519840624300, 1.867896463992650, 1.498250867527830,
    1.501004721822870, 1.432816911840080, 1.328137461314290, 1.380022430660380, 1.248358387335300,
    1.226830408447690, 1.303801476703040,
};

test "FRAMA update value" {
    const tolerance = 1e-9;
    const l_primed = 15;

    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    for (0..l_primed) |i| {
        try testing.expect(math.isNan(frama.update(test_input_mid[i], test_input_high[i], test_input_low[i])));
    }

    for (l_primed..test_input_mid.len) |i| {
        const act = frama.update(test_input_mid[i], test_input_high[i], test_input_low[i]);
        try testing.expect(almostEqual(act, test_expected_frama[i], tolerance));
    }

    // NaN passthrough.
    try testing.expect(math.isNan(frama.update(math.nan(f64), math.nan(f64), math.nan(f64))));
}

test "FRAMA update fdim" {
    const tolerance = 1e-9;
    const l_primed = 15;

    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    for (0..l_primed) |i| {
        _ = frama.update(test_input_mid[i], test_input_high[i], test_input_low[i]);
        try testing.expect(math.isNan(frama.fractal_dimension));
    }

    for (l_primed..test_input_mid.len) |i| {
        _ = frama.update(test_input_mid[i], test_input_high[i], test_input_low[i]);
        try testing.expect(almostEqual(frama.fractal_dimension, test_expected_fdim[i], tolerance));
    }
}

test "FRAMA isPrimed" {
    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    try testing.expect(!frama.isPrimed());

    const l_primed = 15;
    for (0..l_primed) |i| {
        _ = frama.update(test_input_mid[i], test_input_high[i], test_input_low[i]);
        try testing.expect(!frama.isPrimed());
    }

    _ = frama.update(test_input_mid[l_primed], test_input_high[l_primed], test_input_low[l_primed]);
    try testing.expect(frama.isPrimed());
}

test "FRAMA metadata" {
    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    var meta: Metadata = undefined;
    frama.getMetadata(&meta);

    try testing.expectEqual(Identifier.fractal_adaptive_moving_average, meta.identifier);
    try testing.expectEqualStrings("frama(16, 0.010)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqualStrings("frama(16, 0.010)", meta.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("framaDim(16, 0.010)", meta.outputs_buf[1].mnemonic);
}

test "FRAMA constructor" {
    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();

    // Odd length rounds up.
    var frama2 = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 17, .slowest_smoothing_factor = 0.01 });
    defer frama2.deinit();
    try testing.expectEqual(@as(usize, 18), frama2.length);

    try testing.expectError(error.InvalidLength, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 1, .slowest_smoothing_factor = 0.01 }));
    try testing.expectError(error.InvalidLength, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 0, .slowest_smoothing_factor = 0.01 }));
    try testing.expectError(error.InvalidLength, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = -1, .slowest_smoothing_factor = 0.01 }));
    try testing.expectError(error.InvalidSmoothingFactor, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = -0.01 }));
    try testing.expectError(error.InvalidSmoothingFactor, FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 1.01 }));
}

test "FRAMA updateEntity" {
    const tolerance = 1e-9;

    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    // Prime with 15 zeros.
    for (0..15) |_| {
        _ = frama.update(0.0, 0.0, 0.0);
    }

    // 16th update with value 3.0.
    const inp: f64 = 3.0;
    const s = Scalar{ .time = 1, .value = inp };
    const out = frama.updateScalar(&s);
    const outputs = out.slice();
    try testing.expectEqual(@as(usize, 2), outputs.len);
    try testing.expect(almostEqual(outputs[0].scalar.value, 2.999999999999997, tolerance));
    try testing.expect(almostEqual(outputs[1].scalar.value, 1.0000000000000002, tolerance));
}

test "FRAMA updateBar" {
    const tolerance = 1e-9;

    var frama = try FractalAdaptiveMovingAverage.init(testing.allocator, .{ .length = 16, .slowest_smoothing_factor = 0.01 });
    defer frama.deinit();
    frama.fixSlices();

    for (0..test_input_mid.len) |i| {
        const b = Bar{
            .time = @intCast(i),
            .open = 0,
            .high = test_input_high[i],
            .low = test_input_low[i],
            .close = test_input_mid[i],
            .volume = 0,
        };
        const out = frama.updateBar(&b);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 2), outputs.len);

        if (i < 15) {
            try testing.expect(math.isNan(outputs[0].scalar.value));
            try testing.expect(math.isNan(outputs[1].scalar.value));
        } else {
            try testing.expect(almostEqual(outputs[0].scalar.value, test_expected_frama[i], tolerance));
            try testing.expect(almostEqual(outputs[1].scalar.value, test_expected_fdim[i], tolerance));
        }
    }
}
