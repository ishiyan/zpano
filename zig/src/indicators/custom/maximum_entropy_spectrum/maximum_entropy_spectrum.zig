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
const heatmap_mod = @import("../../core/outputs/heatmap.zig");

const OutputArray = indicator_mod.OutputArray;
const OutputValue = indicator_mod.OutputValue;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Heatmap = heatmap_mod.Heatmap;

/// Enumerates the outputs of the MaximumEntropySpectrum indicator.
pub const MaximumEntropySpectrumOutput = enum(u8) {
    value = 1,
};

/// Parameters to create a MaximumEntropySpectrum indicator.
pub const Params = struct {
    length: i32 = 0,
    degree: i32 = 0,
    min_period: f64 = 0,
    max_period: f64 = 0,
    spectrum_resolution: i32 = 0,
    disable_automatic_gain_control: bool = false,
    automatic_gain_control_decay_factor: f64 = 0,
    fixed_normalization: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Error = error{
    InvalidLength,
    InvalidDegree,
    InvalidMinPeriod,
    InvalidMaxPeriod,
    InvalidNyquist,
    InvalidResolution,
    InvalidAgcDecay,
    MnemonicTooLong,
    OutOfMemory,
};

/// Maximum Entropy Spectrum heatmap indicator (Burg method).
pub const MaximumEntropySpectrum = struct {
    allocator: std.mem.Allocator,
    window_count: i32,
    last_index: i32,
    primed: bool,
    floating_normalization: bool,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_buf: [256]u8,
    mnemonic_len: usize,
    description_buf: [320]u8,
    description_len: usize,

    // Estimator state (inlined).
    length: i32,
    degree: i32,
    spectrum_resolution: i32,
    length_spectrum: i32,
    min_period: f64,
    max_period: f64,
    is_automatic_gain_control: bool,
    automatic_gain_control_decay_factor: f64,

    input_series: []f64,
    input_series_minus_mean: []f64,
    coefficients: []f64,
    spectrum: []f64,
    period: []f64,

    // 2D trig tables stored as flat arrays: [length_spectrum * degree].
    frequency_sin_omega: []f64,
    frequency_cos_omega: []f64,

    // Burg working buffers.
    h: []f64, // degree + 1
    g: []f64, // degree + 2
    per: []f64, // length + 1
    pef: []f64, // length + 1

    spectrum_min: f64,
    spectrum_max: f64,
    previous_spectrum_max: f64,

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!MaximumEntropySpectrum {
        const def_length: i32 = 60;
        const def_degree: i32 = 30;
        const def_min_period: f64 = 2.0;
        const def_max_period: f64 = 59.0;
        const def_spectrum_resolution: i32 = 1;
        const def_agc_decay: f64 = 0.995;
        const agc_decay_epsilon: f64 = 1e-12;
        const two_pi: comptime_float = 2.0 * math.pi;

        var length = params.length;
        if (length == 0) length = def_length;

        var degree = params.degree;
        if (degree == 0) degree = def_degree;

        var min_period = params.min_period;
        if (min_period == 0) min_period = def_min_period;

        var max_period = params.max_period;
        if (max_period == 0) max_period = def_max_period;

        var spectrum_resolution = params.spectrum_resolution;
        if (spectrum_resolution == 0) spectrum_resolution = def_spectrum_resolution;

        var agc_decay = params.automatic_gain_control_decay_factor;
        if (agc_decay == 0) agc_decay = def_agc_decay;

        const agc_on = !params.disable_automatic_gain_control;
        const floating_norm = !params.fixed_normalization;

        // Validation.
        if (length < 2) return error.InvalidLength;
        if (degree <= 0 or degree >= length) return error.InvalidDegree;
        if (min_period < 2) return error.InvalidMinPeriod;
        if (max_period <= min_period) return error.InvalidMaxPeriod;
        if (max_period > 2.0 * @as(f64, @floatFromInt(length))) return error.InvalidNyquist;
        if (spectrum_resolution < 1) return error.InvalidResolution;
        if (agc_on and (agc_decay <= 0 or agc_decay >= 1)) return error.InvalidAgcDecay;

        // Components.
        const bc = params.bar_component orelse bar_component.BarComponent.median;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Component mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc,
            qc,
            tc,
        );

        // Build flag tags.
        var flags_buf: [128]u8 = undefined;
        var flags_len: usize = 0;

        if (!agc_on) {
            const tag = ", no-agc";
            @memcpy(flags_buf[flags_len .. flags_len + tag.len], tag);
            flags_len += tag.len;
        }
        if (agc_on and @abs(agc_decay - def_agc_decay) > agc_decay_epsilon) {
            const agc_tag = std.fmt.bufPrint(flags_buf[flags_len..], ", agc={d}", .{agc_decay}) catch
                return error.MnemonicTooLong;
            flags_len += agc_tag.len;
        }
        if (!floating_norm) {
            const tag = ", no-fn";
            @memcpy(flags_buf[flags_len .. flags_len + tag.len], tag);
            flags_len += tag.len;
        }

        const flags = flags_buf[0..flags_len];

        // Build mnemonic.
        var mnemonic_buf: [256]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "mespect({d}, {d}, {d}, {d}, {d}{s}{s})", .{
            length,
            degree,
            min_period,
            max_period,
            spectrum_resolution,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Maximum entropy spectrum {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        // Allocate estimator arrays.
        const length_u: usize = @intCast(length);
        const degree_u: usize = @intCast(degree);
        const length_spectrum: i32 = @intFromFloat((max_period - min_period) * @as(f64, @floatFromInt(spectrum_resolution)) + 1);
        const ls_u: usize = @intCast(length_spectrum);
        const result: f64 = @floatFromInt(spectrum_resolution);

        const input_series = allocator.alloc(f64, length_u) catch return error.OutOfMemory;
        errdefer allocator.free(input_series);
        const input_series_minus_mean = allocator.alloc(f64, length_u) catch return error.OutOfMemory;
        errdefer allocator.free(input_series_minus_mean);
        const coefficients = allocator.alloc(f64, degree_u) catch return error.OutOfMemory;
        errdefer allocator.free(coefficients);
        const spectrum_arr = allocator.alloc(f64, ls_u) catch return error.OutOfMemory;
        errdefer allocator.free(spectrum_arr);
        const period_arr = allocator.alloc(f64, ls_u) catch return error.OutOfMemory;
        errdefer allocator.free(period_arr);

        // 2D trig tables as flat arrays [ls_u * degree_u].
        const trig_size = ls_u * degree_u;
        const freq_sin = allocator.alloc(f64, trig_size) catch return error.OutOfMemory;
        errdefer allocator.free(freq_sin);
        const freq_cos = allocator.alloc(f64, trig_size) catch return error.OutOfMemory;
        errdefer allocator.free(freq_cos);

        for (0..ls_u) |i| {
            const p = max_period - @as(f64, @floatFromInt(i)) / result;
            period_arr[i] = p;
            const theta = two_pi / p;

            for (0..degree_u) |j| {
                const omega = -@as(f64, @floatFromInt(j + 1)) * theta;
                freq_sin[i * degree_u + j] = @sin(omega);
                freq_cos[i * degree_u + j] = @cos(omega);
            }
        }

        // Burg working buffers.
        const h_buf = allocator.alloc(f64, degree_u + 1) catch return error.OutOfMemory;
        errdefer allocator.free(h_buf);
        const g_buf = allocator.alloc(f64, degree_u + 2) catch return error.OutOfMemory;
        errdefer allocator.free(g_buf);
        const per_buf = allocator.alloc(f64, length_u + 1) catch return error.OutOfMemory;
        errdefer allocator.free(per_buf);
        const pef_buf = allocator.alloc(f64, length_u + 1) catch return error.OutOfMemory;
        errdefer allocator.free(pef_buf);

        @memset(input_series, 0);
        @memset(input_series_minus_mean, 0);
        @memset(coefficients, 0);
        @memset(spectrum_arr, 0);
        @memset(h_buf, 0);
        @memset(g_buf, 0);
        @memset(per_buf, 0);
        @memset(pef_buf, 0);

        return .{
            .allocator = allocator,
            .window_count = 0,
            .last_index = length - 1,
            .primed = false,
            .floating_normalization = floating_norm,
            .min_parameter_value = min_period,
            .max_parameter_value = max_period,
            .parameter_resolution = @floatFromInt(spectrum_resolution),
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .length = length,
            .degree = degree,
            .spectrum_resolution = spectrum_resolution,
            .length_spectrum = length_spectrum,
            .min_period = min_period,
            .max_period = max_period,
            .is_automatic_gain_control = agc_on,
            .automatic_gain_control_decay_factor = agc_decay,
            .input_series = input_series,
            .input_series_minus_mean = input_series_minus_mean,
            .coefficients = coefficients,
            .spectrum = spectrum_arr,
            .period = period_arr,
            .frequency_sin_omega = freq_sin,
            .frequency_cos_omega = freq_cos,
            .h = h_buf,
            .g = g_buf,
            .per = per_buf,
            .pef = pef_buf,
            .spectrum_min = 0,
            .spectrum_max = 0,
            .previous_spectrum_max = 0,
        };
    }

    pub fn deinit(self: *MaximumEntropySpectrum) void {
        self.allocator.free(self.input_series);
        self.allocator.free(self.input_series_minus_mean);
        self.allocator.free(self.coefficients);
        self.allocator.free(self.spectrum);
        self.allocator.free(self.period);
        self.allocator.free(self.frequency_sin_omega);
        self.allocator.free(self.frequency_cos_omega);
        self.allocator.free(self.h);
        self.allocator.free(self.g);
        self.allocator.free(self.per);
        self.allocator.free(self.pef);
    }

    pub fn fixSlices(self: *MaximumEntropySpectrum) void {
        _ = self;
    }

    fn mnemonicSlice(self: *const MaximumEntropySpectrum) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn descriptionSlice(self: *const MaximumEntropySpectrum) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    // --- Burg estimator ---

    fn burgEstimate(self: *MaximumEntropySpectrum, series: []const f64) void {
        const length_u: usize = @intCast(self.length);
        const degree_u: usize = @intCast(self.degree);

        // Zero per and pef (1-based indexing).
        for (1..length_u + 1) |i| {
            self.pef[i] = 0;
            self.per[i] = 0;
        }

        for (1..degree_u + 1) |i| {
            var sn: f64 = 0;
            var sd: f64 = 0;
            var jj: usize = length_u - i;

            for (0..jj) |j| {
                const t1 = series[j + i] + self.pef[j];
                const t2 = series[j] + self.per[j];
                sn -= 2.0 * t1 * t2;
                sd += t1 * t1 + t2 * t2;
            }

            const t = sn / sd;
            self.g[i] = t;

            if (i != 1) {
                for (1..i) |j| {
                    self.h[j] = self.g[j] + t * self.g[i - j];
                }
                for (1..i) |j| {
                    self.g[j] = self.h[j];
                }
                jj -= 1;
            }

            for (0..jj) |j| {
                self.per[j] += t * self.pef[j] + t * series[j + i];
                self.pef[j] = self.pef[j + 1] + t * self.per[j + 1] + t * series[j + 1];
            }
        }

        for (0..degree_u) |i| {
            self.coefficients[i] = -self.g[i + 1];
        }
    }

    fn calculate(self: *MaximumEntropySpectrum) void {
        const len_u: usize = @intCast(self.length);
        const ls_u: usize = @intCast(self.length_spectrum);
        const deg_u: usize = @intCast(self.degree);

        // Subtract mean.
        var mean: f64 = 0;
        for (0..len_u) |i| {
            mean += self.input_series[i];
        }
        mean /= @floatFromInt(self.length);

        for (0..len_u) |i| {
            self.input_series_minus_mean[i] = self.input_series[i] - mean;
        }

        self.burgEstimate(self.input_series_minus_mean);

        // Evaluate spectrum from AR coefficients.
        self.spectrum_min = math.floatMax(f64);
        if (self.is_automatic_gain_control) {
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max;
        } else {
            self.spectrum_max = -math.floatMax(f64);
        }

        for (0..ls_u) |i| {
            var real: f64 = 1.0;
            var imag: f64 = 0.0;

            for (0..deg_u) |j| {
                real -= self.coefficients[j] * self.frequency_cos_omega[i * deg_u + j];
                imag -= self.coefficients[j] * self.frequency_sin_omega[i * deg_u + j];
            }

            const s = 1.0 / (real * real + imag * imag);
            self.spectrum[i] = s;

            if (self.spectrum_max < s) {
                self.spectrum_max = s;
            }
            if (self.spectrum_min > s) {
                self.spectrum_min = s;
            }
        }

        self.previous_spectrum_max = self.spectrum_max;
    }

    // --- Public interface ---

    pub fn updateSample(self: *MaximumEntropySpectrum, sample: f64, time: i64) Heatmap {
        if (math.isNan(sample)) {
            return Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution);
        }

        const last_idx: usize = @intCast(self.last_index);

        if (self.primed) {
            const len_u: usize = @intCast(self.length);
            std.mem.copyForwards(f64, self.input_series[0..last_idx], self.input_series[1..len_u]);
            self.input_series[last_idx] = sample;
        } else {
            const wc: usize = @intCast(self.window_count);
            self.input_series[wc] = sample;
            self.window_count += 1;
            if (self.window_count == self.length) {
                self.primed = true;
            }
        }

        if (!self.primed) {
            return Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution);
        }

        self.calculate();

        const ls_u: usize = @intCast(self.length_spectrum);

        var min_ref: f64 = 0;
        if (self.floating_normalization) {
            min_ref = self.spectrum_min;
        }
        const max_ref = self.spectrum_max;
        const spectrum_range = max_ref - min_ref;

        var values: [heatmap_mod.max_heatmap_values]f64 = undefined;
        var value_min: f64 = math.inf(f64);
        var value_max: f64 = -math.inf(f64);

        for (0..ls_u) |i| {
            const v = (self.spectrum[ls_u - 1 - i] - min_ref) / spectrum_range;
            values[i] = v;
            if (v < value_min) value_min = v;
            if (v > value_max) value_max = v;
        }

        return Heatmap.new(
            time,
            self.min_parameter_value,
            self.max_parameter_value,
            self.parameter_resolution,
            value_min,
            value_max,
            values[0..ls_u],
        );
    }

    // --- Entity update methods ---

    pub fn updateBar(self: *MaximumEntropySpectrum, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *MaximumEntropySpectrum, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *MaximumEntropySpectrum, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    pub fn updateScalar(self: *MaximumEntropySpectrum, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    fn updateEntity(self: *MaximumEntropySpectrum, time: i64, sample: f64) OutputArray {
        const hm = self.updateSample(sample, time);
        var out = OutputArray{};
        out.append(.{ .heatmap = hm });
        return out;
    }

    pub fn isPrimed(self: *const MaximumEntropySpectrum) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const MaximumEntropySpectrum, out: *Metadata) void {
        const mn = self.mnemonicSlice();
        const desc = self.descriptionSlice();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
        };
        build_metadata_mod.buildMetadata(out, .maximum_entropy_spectrum, mn, desc, &texts);
    }

    /// Expose coefficients for testing (Burg coefficient validation).
    pub fn getCoefficients(self: *const MaximumEntropySpectrum) []const f64 {
        return self.coefficients;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *MaximumEntropySpectrum) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(MaximumEntropySpectrum);
};

// --- Tests ---
const testing = std.testing;

const test_input = [_]f64{
    92.0000,  93.1725,  95.3125,  94.8450,  94.4075,  94.1100,  93.5000,  91.7350,  90.9550,  91.6875,
    94.5000,  97.9700,  97.5775,  90.7825,  89.0325,  92.0950,  91.1550,  89.7175,  90.6100,  91.0000,
    88.9225,  87.5150,  86.4375,  83.8900,  83.0025,  82.8125,  82.8450,  86.7350,  86.8600,  87.5475,
    85.7800,  86.1725,  86.4375,  87.2500,  88.9375,  88.2050,  85.8125,  84.5950,  83.6575,  84.4550,
    83.5000,  86.7825,  88.1725,  89.2650,  90.8600,  90.7825,  91.8600,  90.3600,  89.8600,  90.9225,
    89.5000,  87.6725,  86.5000,  84.2825,  82.9075,  84.2500,  85.6875,  86.6100,  88.2825,  89.5325,
    89.5000,  88.0950,  90.6250,  92.2350,  91.6725,  92.5925,  93.0150,  91.1725,  90.9850,  90.3775,
    88.2500,  86.9075,  84.0925,  83.1875,  84.2525,  97.8600,  99.8750,  103.2650, 105.9375, 103.5000,
    103.1100, 103.6100, 104.6400, 106.8150, 104.9525, 105.5000, 107.1400, 109.7350, 109.8450, 110.9850,
    120.0000, 119.8750, 117.9075, 119.4075, 117.9525, 117.2200, 115.6425, 113.1100, 111.7500, 114.5175,
    114.7450, 115.4700, 112.5300, 112.0300, 113.4350, 114.2200, 119.5950, 117.9650, 118.7150, 115.0300,
    114.5300, 115.0000, 116.5300, 120.1850, 120.5000, 120.5950, 124.1850, 125.3750, 122.9700, 123.0000,
    124.4350, 123.4400, 124.0300, 128.1850, 129.6550, 130.8750, 132.3450, 132.0650, 133.8150, 135.6600,
    137.0350, 137.4700, 137.3450, 136.3150, 136.4400, 136.2850, 129.0950, 128.3100, 126.0000, 124.0300,
    123.9350, 125.0300, 127.2500, 125.6200, 125.5300, 123.9050, 120.6550, 119.9650, 120.7800, 124.0000,
    122.7800, 120.7200, 121.7800, 122.4050, 123.2500, 126.1850, 127.5600, 126.5650, 123.0600, 122.7150,
    123.5900, 122.3100, 122.4650, 123.9650, 123.9700, 124.1550, 124.4350, 127.0000, 125.5000, 128.8750,
    130.5350, 132.3150, 134.0650, 136.0350, 133.7800, 132.7500, 133.4700, 130.9700, 127.5950, 128.4400,
    127.9400, 125.8100, 124.6250, 122.7200, 124.0900, 123.2200, 121.4050, 120.9350, 118.2800, 118.3750,
    121.1550, 120.9050, 117.1250, 113.0600, 114.9050, 112.4350, 107.9350, 105.9700, 106.3700, 106.8450,
    106.9700, 110.0300, 91.0000,  93.5600,  93.6200,  95.3100,  94.1850,  94.7800,  97.6250,  97.5900,
    95.2500,  94.7200,  92.2200,  91.5650,  92.2200,  93.8100,  95.5900,  96.1850,  94.6250,  95.1200,
    94.0000,  93.7450,  95.9050,  101.7450, 106.4400, 107.9350, 103.4050, 105.0600, 104.1550, 103.3100,
    103.3450, 104.8400, 110.4050, 114.5000, 117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300,
    106.2200, 107.7200, 109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
    109.5300, 108.0600,
};

const test_tolerance = 1e-12;
const test_min_max_tol = 1e-10;

fn almostEqual(a: f64, b: f64, eps: f64) bool {
    return @abs(a - b) <= eps;
}

const SpotValue = struct { i: usize, v: f64 };
const MesSnap = struct {
    i: usize,
    value_min: f64,
    value_max: f64,
    spots: []const SpotValue,
};

const mes_snapshots = [_]MesSnap{
    .{
        .i = 59,
        .value_min = 0,
        .value_max = 1,
        .spots = &[_]SpotValue{
            .{ .i = 0, .v = 0.000000000000000 },
            .{ .i = 14, .v = 0.124709393535801 },
            .{ .i = 28, .v = 0.021259483287733 },
            .{ .i = 42, .v = 0.726759100473496 },
            .{ .i = 57, .v = 0.260829244402141 },
        },
    },
    .{
        .i = 60,
        .value_min = 0,
        .value_max = 0.3803558166,
        .spots = &[_]SpotValue{
            .{ .i = 0, .v = 0.000000000000000 },
            .{ .i = 14, .v = 0.047532484316402 },
            .{ .i = 28, .v = 0.156007210177695 },
            .{ .i = 42, .v = 0.204392941920655 },
            .{ .i = 57, .v = 0.099988829337396 },
        },
    },
    .{
        .i = 100,
        .value_min = 0,
        .value_max = 0.7767627734,
        .spots = &[_]SpotValue{
            .{ .i = 0, .v = 0.000000000000000 },
            .{ .i = 14, .v = 0.005541589459818 },
            .{ .i = 28, .v = 0.019544065000896 },
            .{ .i = 42, .v = 0.045342308770863 },
            .{ .i = 57, .v = 0.776762773404885 },
        },
    },
    .{
        .i = 150,
        .value_min = 0,
        .value_max = 0.0126783313,
        .spots = &[_]SpotValue{
            .{ .i = 0, .v = 0.000347619185321 },
            .{ .i = 14, .v = 0.001211800388686 },
            .{ .i = 28, .v = 0.001749939543675 },
            .{ .i = 42, .v = 0.010949450171300 },
            .{ .i = 57, .v = 0.001418701588812 },
        },
    },
    .{
        .i = 200,
        .value_min = 0,
        .value_max = 0.5729940203,
        .spots = &[_]SpotValue{
            .{ .i = 0, .v = 0.000000000000000 },
            .{ .i = 14, .v = 0.047607367831419 },
            .{ .i = 28, .v = 0.013304430092822 },
            .{ .i = 42, .v = 0.137193402225458 },
            .{ .i = 57, .v = 0.506646287515276 },
        },
    },
};

test "MaximumEntropySpectrum update" {
    var x = try MaximumEntropySpectrum.init(testing.allocator, .{});
    defer x.deinit();

    var si: usize = 0;
    for (0..test_input.len) |i| {
        const h = x.updateSample(test_input[i], @intCast(i));

        try testing.expectEqual(@as(f64, 2.0), h.parameter_first);
        try testing.expectEqual(@as(f64, 59.0), h.parameter_last);
        try testing.expectEqual(@as(f64, 1.0), h.parameter_resolution);

        if (!x.isPrimed()) {
            try testing.expect(h.isEmpty());
            continue;
        }

        try testing.expectEqual(@as(usize, 58), h.values_len);

        if (si < mes_snapshots.len and mes_snapshots[si].i == i) {
            const snap = mes_snapshots[si];
            try testing.expect(almostEqual(h.value_min, snap.value_min, test_min_max_tol));
            try testing.expect(almostEqual(h.value_max, snap.value_max, test_min_max_tol));

            const vals = h.valuesSlice();
            for (snap.spots) |sp| {
                try testing.expect(almostEqual(vals[sp.i], sp.v, test_tolerance));
            }

            si += 1;
        }
    }

    try testing.expectEqual(mes_snapshots.len, si);
}

test "MaximumEntropySpectrum primes at bar 59" {
    var x = try MaximumEntropySpectrum.init(testing.allocator, .{});
    defer x.deinit();

    try testing.expect(!x.isPrimed());

    var primed_at: ?usize = null;
    for (0..test_input.len) |i| {
        _ = x.updateSample(test_input[i], @intCast(i));
        if (x.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expectEqual(@as(usize, 59), primed_at.?);
}

test "MaximumEntropySpectrum NaN input" {
    var x = try MaximumEntropySpectrum.init(testing.allocator, .{});
    defer x.deinit();

    const h = x.updateSample(math.nan(f64), 0);
    try testing.expect(h.isEmpty());
    try testing.expect(!x.isPrimed());
}

test "MaximumEntropySpectrum metadata" {
    var x = try MaximumEntropySpectrum.init(testing.allocator, .{});
    defer x.deinit();

    var md: Metadata = undefined;
    x.getMetadata(&md);

    const mn = "mespect(60, 30, 2, 59, 1, hl/2)";

    try testing.expectEqualStrings(mn, x.mnemonicSlice());
    try testing.expectEqual(Identifier.maximum_entropy_spectrum, md.identifier);
    try testing.expectEqualStrings(mn, md.mnemonic);
    try testing.expectEqual(@as(usize, 1), md.outputs_len);

    const outputs = md.outputs_buf[0..md.outputs_len];
    try testing.expectEqualStrings(mn, outputs[0].mnemonic);
}

test "MaximumEntropySpectrum mnemonic flags" {
    const TestCase = struct { params: Params, mn: []const u8 };
    const cases = [_]TestCase{
        .{ .params = .{}, .mn = "mespect(60, 30, 2, 59, 1, hl/2)" },
        .{ .params = .{ .disable_automatic_gain_control = true }, .mn = "mespect(60, 30, 2, 59, 1, no-agc, hl/2)" },
        .{ .params = .{ .automatic_gain_control_decay_factor = 0.8 }, .mn = "mespect(60, 30, 2, 59, 1, agc=0.8, hl/2)" },
        .{ .params = .{ .fixed_normalization = true }, .mn = "mespect(60, 30, 2, 59, 1, no-fn, hl/2)" },
        .{ .params = .{
            .disable_automatic_gain_control = true,
            .fixed_normalization = true,
        }, .mn = "mespect(60, 30, 2, 59, 1, no-agc, no-fn, hl/2)" },
    };

    for (cases) |tc| {
        var x = try MaximumEntropySpectrum.init(testing.allocator, tc.params);
        defer x.deinit();
        try testing.expectEqualStrings(tc.mn, x.mnemonicSlice());
    }
}

test "MaximumEntropySpectrum validation" {
    try testing.expectError(error.InvalidLength, MaximumEntropySpectrum.init(testing.allocator, .{
        .length = 1,
        .degree = 1,
        .min_period = 2,
        .max_period = 4,
        .spectrum_resolution = 1,
    }));
    try testing.expectError(error.InvalidDegree, MaximumEntropySpectrum.init(testing.allocator, .{
        .length = 4,
        .degree = 4,
        .min_period = 2,
        .max_period = 4,
        .spectrum_resolution = 1,
    }));
    try testing.expectError(error.InvalidMinPeriod, MaximumEntropySpectrum.init(testing.allocator, .{
        .length = 60,
        .degree = 30,
        .min_period = 1,
        .max_period = 59,
        .spectrum_resolution = 1,
    }));
    try testing.expectError(error.InvalidMaxPeriod, MaximumEntropySpectrum.init(testing.allocator, .{
        .length = 60,
        .degree = 30,
        .min_period = 10,
        .max_period = 10,
        .spectrum_resolution = 1,
    }));
    try testing.expectError(error.InvalidNyquist, MaximumEntropySpectrum.init(testing.allocator, .{
        .length = 10,
        .degree = 5,
        .min_period = 2,
        .max_period = 59,
        .spectrum_resolution = 1,
    }));
    try testing.expectError(error.InvalidAgcDecay, MaximumEntropySpectrum.init(testing.allocator, .{
        .automatic_gain_control_decay_factor = -0.1,
    }));
    try testing.expectError(error.InvalidAgcDecay, MaximumEntropySpectrum.init(testing.allocator, .{
        .automatic_gain_control_decay_factor = 1.0,
    }));
}

test "MaximumEntropySpectrum updateEntity" {
    const prime_count = 70;
    const inp: f64 = 100.0;
    const time: i64 = 0;

    // Update scalar
    {
        var x = try MaximumEntropySpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const s = Scalar{ .time = time, .value = inp };
        const out = x.updateScalar(&s);
        try testing.expectEqual(@as(usize, 1), out.len);
    }

    // Update bar
    {
        var x = try MaximumEntropySpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const b = Bar{ .time = time, .open = inp, .high = inp, .low = inp, .close = inp, .volume = 0 };
        const out = x.updateBar(&b);
        try testing.expectEqual(@as(usize, 1), out.len);
    }

    // Update quote
    {
        var x = try MaximumEntropySpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const q = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };
        const out = x.updateQuote(&q);
        try testing.expectEqual(@as(usize, 1), out.len);
    }

    // Update trade
    {
        var x = try MaximumEntropySpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const t = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = x.updateTrade(&t);
        try testing.expectEqual(@as(usize, 1), out.len);
    }
}

// --- Burg coefficient test data ---
// TODO: Copy test data from go/indicators/custom/maximumentropyspectrum/data_test.go
// These are placeholder empty arrays. Replace with the actual data from Go.
const nan = math.nan(f64);
const test_input_four_sinusoids = [_]f64{
    0,          0.423468,    0.773358,    0.994145,   1.06138,    0.987356,   0.817725,   0.619848,    0.46583,      0.414411,
    0.496197,   0.705906,    1.00353,     1.32416,    1.59398,    1.74856,    1.74886,    1.59099,     1.30736,      0.958791,
    0.619794,   0.360352,    0.228939,    0.240858,   0.374817,   0.578574,   0.782249,   0.916001,    0.9277,       0.796193,
    0.536904,   0.198347,    -0.149568,   -0.432559,  -0.590119,  -0.590286,  -0.437408,  -0.170958,   0.14429,      0.43429,
    0.634006,   0.703413,    0.637257,    0.466174,   0.248881,   0.0573801,  -0.0411547, -0.00132771, 0.188895,     0.504843,
    0.890862,   1.27371,     1.58028,     1.75536,    1.77488,    1.65156,    1.43145,    1.18225,     0.976284,     0.872449,
    0.90143,    1.05801,     1.30229,     1.56954,    1.78624,    1.88836,    1.83732,    1.6297,      1.29833,      0.904381,
    0.522552,   0.222876,    0.0537461,   0.0303063,  0.131076,   0.303651,   0.478056,   0.584438,    0.570723,     0.41585,
    0.13531,    -0.222383,   -0.587437,   -0.885816,  -1.05737,   -1.0706,    -0.930302,  -0.676386,   -0.373967,    -0.097304,
    0.0884771,  0.143364,    0.0621751,   -0.124379,  -0.357554,  -0.565395,  -0.68086,   -0.658745,   -0.487223,    -0.191133,
    0.173801,   0.534393,    0.817753,    0.96898,    0.964429,   0.817267,   0.573948,   0.302494,    0.0754659,    -0.0481357,
    -0.0375786, 0.101938,    0.330558,    0.583661,   0.787938,   0.87965,    0.820574,   0.607662,    0.274043,     -0.1189,
    -0.496418,  -0.788562,   -0.947153,   -0.957347,  -0.84095,   -0.650658,  -0.456669,  -0.328971,   -0.319693,    -0.449915,
    -0.70417,   -1.03404,    -1.36991,    -1.63806,   -1.77877,   -1.761,     -1.59006,   -1.30629,    -0.975154,    -0.671118,
    -0.45938,   -0.379911,   -0.43779,    -0.602191,  -0.814298,  -1.00216,   -1.09881,   -1.05919,    -0.871635,    -0.561098,
    -0.183334,  0.188571,    0.48196,     0.64229,    0.646347,   0.507745,   0.273345,   0.0114816,   -0.205088,    -0.317282,
    -0.294368,  -0.141584,   0.101195,    0.36939,    0.589822,   0.698965,   0.65886,    0.466728,    0.155912,     -0.212101,
    -0.562597,  -0.825809,   -0.953864,   -0.932301,  -0.783328,  -0.560005,  -0.332811,  -0.171915,   -0.129546,    -0.226824,
    -0.448326,  -0.74572,    -1.04958,    -1.28648,   -1.39706,   -1.35072,   -1.15322,   -0.84527,    -0.492591,    -0.169772,
    0.0580117,  0.150927,    0.104103,    -0.0514171, -0.256638,  -0.439503,  -0.533032,  -0.492206,   -0.305419,    0.00233809,
    0.375365,   0.740745,    1.02612,     1.17736,    1.17172,    1.02329,    0.779355,   0.508575,    0.283898,     0.164476,
    0.181015,   0.328212,    0.566131,    0.830195,   1.0473,     1.15409,    1.1128,     0.920872,    0.611805,     0.247137,
    -0.0985058, -0.355585,   -0.476573,   -0.447425,  -0.290772,  -0.0600507, 0.173975,   0.340961,    0.388602,     0.295764,
    0.077871,   -0.216787,   -0.518906,   -0.755263,  -0.866809,  -0.823292,  -0.630797,  -0.330316,   0.0122797,    0.32239,
    0.534976,   0.610456,    0.544277,    0.367807,   0.140323,   -0.0659218, -0.183836,  -0.168364,   -0.00789333,  0.272616,
    0.617564,   0.954243,    1.21062,     1.33299,    1.29907,    1.12343,    0.853751,   0.558978,    0.312208,     0.172609,
    0.170803,   0.301349,    0.524182,    0.774649,   0.979648,   1.07589,    1.02577,    0.826845,    0.512698,     0.144852,
    -0.202215,  -0.459248,   -0.579111,   -0.54821,   -0.389627,  -0.157189,  0.0780519,  0.245582,    0.293038,     0.199299,
    -0.0201741, -0.317086,   -0.622197,   -0.862439,  -0.978994,  -0.941885,  -0.757459,  -0.4669,     -0.136156,    0.160241,
    0.357464,   0.416251,    0.332433,    0.13776,    -0.108161,  -0.332837,  -0.46904,   -0.471656,   -0.329057,    -0.0661795,
    0.261465,   0.581352,    0.821735,    0.92927,    0.882105,   0.695194,   0.416544,   0.115293,    -0.135406,    -0.276469,
    -0.277464,  -0.144069,   0.0834135,   0.340148,   0.552935,   0.658472,   0.61918,    0.432677,    0.132547,     -0.219775,
    -0.550019,  -0.789263,   -0.8908,     -0.841514,  -0.66495,   -0.41533,   -0.163986,  0.0184232,   0.0795033,    -0.0018122,
    -0.20998,   -0.496643,   -0.792567,   -1.02477,   -1.13461,   -1.09229,   -0.904373,  -0.612152,   -0.281586,    0.012925,
    0.206821,   0.261213,    0.172357,    -0.0275801, -0.27864,   -0.508079,  -0.648526,  -0.654821,   -0.515342,    -0.255027,
    0.0706568,  0.389315,    0.629428,    0.737956,   0.69338,    0.510971,   0.238965,   -0.0533937,  -0.292992,    -0.420925,
    -0.40704,   -0.257349,   -0.0124734,  0.26249,    0.494168,   0.619175,   0.599917,   0.434011,    0.155014,     -0.175754,
    -0.484244,  -0.701869,   -0.782339,   -0.712991,  -0.517802,  -0.251338,  0.0148447,  0.210004,    0.28179,      0.20927,
    0.00815507, -0.273049,   -0.565025,   -0.794788,  -0.903759,  -0.86228,   -0.676993,  -0.38925,    -0.0649408,   0.221735,
    0.406543,   0.451014,    0.351866,    0.14171,    -0.11913,   -0.357653,  -0.506358,  -0.520055,   -0.387154,    -0.132637,
    0.188028,   0.502523,    0.739494,    0.846135,   0.801193,   0.620175,   0.351473,   0.0643611,   -0.168164,    -0.287447,
    -0.263685,  -0.10328,    0.152774,    0.439073,   0.682036,   0.818172,   0.809853,   0.654697,    0.386238,     0.0657007,
    -0.233056,  -0.441736,   -0.514418,   -0.438828,  -0.239302,  0.0293255,  0.295364,   0.488078,    0.555258,     0.476201,
    0.266884,   -0.0239422,  -0.326781,   -0.568556,  -0.690674,  -0.663516,  -0.493758,  -0.22273,    0.083798,     0.351866,
    0.517597,   0.542963,    0.425153,    0.197214,   -0.0801778, -0.333795,  -0.496042,  -0.521746,   -0.399408,    -0.154116,
    0.15859,    0.466396,    0.698033,    0.800848,   0.753772,   0.572468,   0.305403,   0.0218043,   -0.205517,    -0.318222,
    -0.286909,  -0.118421,   0.145839,    0.440132,   0.690656,   0.833807,   0.831932,   0.682666,    0.419546,     0.103743,
    -0.19103,   -0.396711,   -0.467679,   -0.391983,  -0.194237,  0.0705383,  0.330601,   0.51531,     0.572681,     0.482325,
    0.260563,   -0.0435349,  -0.360224,   -0.616274,  -0.753031,  -0.740865,  -0.586457,  -0.331093,   -0.0404128,   0.211864,
    0.362206,   0.373002,    0.241872,    0.00225745, -0.284874,  -0.546134,  -0.713905,  -0.743116,   -0.622443,    -0.377169,
    -0.062997,  0.247667,    0.483548,    0.592058,   0.552221,   0.379776,   0.123187,   -0.148434,   -0.362523,    -0.461106,
    -0.415228,  -0.232196,   0.0461557,   0.353752,   0.616572,   0.770918,   0.77914,    0.638927,    0.383866,     0.0751297,
    -0.213651,  -0.41458,    -0.482264,   -0.404995,  -0.207585,  0.05484,    0.310565,   0.489124,    0.538829,     0.439671,
    0.208379,   -0.105607,   -0.43225,    -0.698135,  -0.84452,   -0.841758,  -0.696527,  -0.450083,   -0.167956,    0.0763453,
    0.21959,    0.22453,     0.089159,    -0.152763,  -0.440052,  -0.69924,   -0.862779,  -0.885786,   -0.757205,    -0.502605,
    -0.177939,  0.144211,    0.392483,    0.514279,   0.488653,   0.331356,   0.0908046,  -0.164046,   -0.360908,    -0.442188,
    -0.379381,  -0.180254,   0.112949,    0.433846,   0.708241,   0.872392,   0.888712,   0.755013,    0.505007,     0.199946,
    -0.086537,  -0.286626,   -0.355069,   -0.28031,   -0.0872786, 0.168879,   0.416553,   0.585521,    0.624457,     0.513788,
    0.270694,   -0.0549787,  -0.392881,   -0.669399,  -0.825702,  -0.832134,  -0.695397,  -0.456743,   -0.181635,    0.0565751,
    0.194901,   0.196391,    0.0593363,   -0.18226,   -0.46708,   -0.721664,  -0.878613,  -0.893318,   -0.755067,    -0.489793,
    -0.153767,  0.18019,     0.440576,    0.574738,   0.561722,   0.41727,    0.189733,   -0.0520924,  -0.236184,    -0.305313,
    -0.23139,   -0.0226004,  0.278457,    0.605154,   0.883191,   1.04886,    1.06472,    0.92879,     0.675001,     0.364773,
    0.0718325,  -0.135998,   -0.213517,   -0.149246,  0.031849,   0.274668,   0.507768,   0.66122,     0.684103,     0.557308,
    0.298482,   -0.0421149,  -0.393826,   -0.682858,  -0.850318,  -0.866583,  -0.738423,  -0.507147,   -0.238214,    -0.00488759,
    0.13001,    0.129744,    -0.00717942, -0.246601,  -0.527153,  -0.775465,  -0.924368,  -0.929599,   -0.780861,    -0.504506,
    -0.157177,  0.188027,    0.459435,    0.604317,   0.601704,   0.467329,   0.249494,   0.0168588,   -0.158785,    -0.220524,
    -0.140632,  0.0723513,   0.375543,    0.70215,    0.977847,   1.13905,    1.14855,    1.00466,     0.741609,     0.421078,
    0.116961,   -0.102839,   -0.19311,    -0.142387,  0.0243726,  0.252156,   0.469715,   0.607435,    0.6148,       0.473161,
    0.200619,   -0.152188,   -0.514335,   -0.811897,  -0.985984,  -1.00707,   -0.882067,  -0.652423,   -0.383677,    -0.149093,
    -0.0114134, -0.00724835, -0.137963,   -0.369327,  -0.640001,  -0.87678,   -1.01279,   -1.00418,    -0.841112,    -0.550393,
    -0.189065,  0.169476,    0.453384,    0.609863,   0.617945,   0.493385,   0.28447,    0.0597765,   -0.109079,    -0.165434,
    -0.0818547, 0.132846,    0.435582,    0.759476,   1.03027,    1.18458,    1.18551,    1.03176,     0.757923,     0.426,
    0.110113,   -0.121703,   -0.224184,   -0.185852,  -0.0316328, 0.183554,   0.388655,   0.51435,     0.510507,     0.358901,
    0.0780314,  -0.281136,   -0.647475,   -0.947005,  -1.12092,   -1.13986,   -1.011,     -0.775978,   -0.50052,     -0.257973,
    -0.111088,  -0.0964275,  -0.215304,   -0.433483,  -0.689719,  -0.911019,  -1.03085,   -1.0058,     -0.826502,    -0.52024,
    -0.144445,  0.22721,     0.522731,    0.689293,   0.705976,   0.588616,   0.38555,    0.165343,    -0.000487246, -0.0554472,
    0.0276926,  0.23993,     0.538064,    0.855221,   1.11729,    1.26116,    1.25034,    1.08396,     0.79704,      0.451947,
    0.123068,   -0.121406,   -0.236146,   -0.209655,  -0.0668393, 0.137458,   0.332345,   0.448758,    0.436899,     0.278902,
    -0.0063955, -0.367779,   -0.734002,   -1.03111,   -1.20046,   -1.21298,   -1.07613,   -0.831877,   -0.546189,    -0.292574,
    -0.133856,  -0.106612,   -0.212153,   -0.416283,  -0.657878,  -0.864185,  -0.969019,  -0.929391,   -0.736414,    -0.417806,
    -0.0313555, 0.349036,    0.651283,    0.822597,   0.842183,   0.726031,   0.522614,   0.30057,     0.131407,     0.0715384,
    0.148075,   0.351913,    0.639815,    0.944989,   1.19353,    1.3227,     1.29641,    1.11428,     0.811808,     0.451726,
    0.108699,   -0.148933,   -0.275791,   -0.260386,  -0.127641,  0.0676308,  0.254642,   0.364523,    0.347736,     0.1867,
    -0.0994952, -0.45947,    -0.821944,   -1.11308,   -1.27448,   -1.27742,   -1.12975,   -0.873839,   -0.575941,    -0.309785,
    -0.138314,  -0.0981532,  -0.190637,   -0.38162,   -0.610107,  -0.803567,  -0.896146,  -0.845255,   -0.642431,    -0.31578,
    0.0766167,  0.460697,    0.764362,    0.934941,   0.951845,   0.831309,   0.622029,   0.3928,      0.215206,     0.14566,
    0.211231,   0.40278,     0.677099,    0.967527,   1.20042,    1.31341,    1.27088,    1.07293,     0.755518,     0.381747,
    0.0265274,  -0.241686,   -0.377511,   -0.369524,  -0.242726,  -0.0519748, 0.131973,   0.240359,    0.22382,      0.0649731,
    -0.216928,  -0.570425,   -0.92429,    -1.20488,   -1.35412,   -1.34371,   -1.18195,   -0.911625,   -0.599352,    -0.319102,
    -0.133951,  -0.0805809,  -0.160344,   -0.339131,  -0.556045,  -0.738745,  -0.821659,  -0.762539,   -0.553284,    -0.222313,
    0.172077,   0.555747,    0.856671,    1.02238,    1.03258,    0.903838,   0.685148,   0.445541,    0.256747,     0.17524,
    0.228101,   0.406203,    0.6664,      0.94219,    1.1602,     1.25844,    1.20175,    0.990692,    0.661646,     0.278046,
    -0.0850079, -0.359008,   -0.498639,   -0.492619,  -0.366108,  -0.174092,  0.0126146,  0.125278,    0.114622,     -0.0366226,
    -0.30912,   -0.651412,   -0.99239,    -1.25867,   -1.39256,   -1.36621,   -1.18842,   -0.902411,   -0.575169,    -0.280901,
    -0.0828112, -0.0176075,  -0.0866286,  -0.255755,  -0.464134,  -0.639553,  -0.716649,  -0.653444,   -0.442113,    -0.111304,
    0.280559,   0.659345,    0.953188,    1.10991,    1.10959,    0.969184,   0.73807,    0.485567,    0.283598,     0.188737,
    0.228107,   0.392609,    0.639168,    0.901434,   1.10629,    1.1921,     1.1241,     0.903303,    0.56643,      0.17718,
    -0.189229,  -0.464315,   -0.602912,   -0.593968,  -0.462888,  -0.264872,  -0.0708764, 0.0503074,   0.0494151,    -0.0907822,
    -0.350928,  -0.679621,   -1.00592,    -1.25673,   -1.37477,   -1.33266,   -1.1397,    -0.839534,   -0.499489,    -0.193985,
    0.013698,   0.0868723,   0.0242778,   -0.139888,  -0.344742,  -0.518119,  -0.59478,   -0.53293,    -0.324931,
};
const test_input_test1 = [_]f64{
    -0.154493,   0.267746, -0.467592, 0.590125,  1.66508,    2.69146,   2.74053,   2.83912,    2.13444,   2.84587,
    3.75228,     3.87505,  3.09342,   2.98394,   2.6341,     2.44277,   2.28914,   2.52084,    2.47011,   3.75575,
    4.62112,     3.20226,  1.81495,   0.426689,  2.20248,    3.35074,   2.64228,   0.89022,    1.0758,    1.49673,
    2.43165,     1.12177,  1.19072,   0.556267,  1.42577,    1.1862,    1.33601,   -0.461124,  2.39259,   2.79946,
    2.38925,     2.50728,  2.74062,   3.22706,   2.39703,    2.53098,   3.10527,   3.84903,    2.98466,   2.96785,
    2.83976,     3.92499,  3.39933,   1.21796,   0.148419,   0.306743,  2.5188,    1.38492,    0.327015,  -1.71652,
    0.115546,    1.05554,  1.14376,   0.960662,  -0.695491,  0.106176,  2.72558,   5.41535,    5.20301,   4.12257,
    4.18681,     6.57888,  7.84211,   7.01215,   6.25565,    5.96839,   6.52322,   5.79923,    3.49086,   2.47817,
    3.29932,     5.74942,  3.6973,    2.41183,   1.34383,    2.61797,   4.75977,   5.44577,    3.37387,   0.919866,
    1.05551,     1.70275,  3.12487,   1.8639,    2.12439,    3.34981,   6.75375,   7.7149,     5.83674,   4.29095,
    5.34984,     7.81248,  7.60164,   4.25137,   4.45683,    6.66606,   7.34578,   4.78136,    2.01286,   3.49902,
    7.52819,     7.63999,  4.86152,   3.28031,   5.24192,    6.51857,   5.93301,   1.99266,    1.59152,   4.3494,
    6.66992,     4.93506,  3.88175,   5.57268,   9.24591,    10.2713,   9.4236,    8.06345,    11.2273,   13.3064,
    12.8033,     10.4785,  9.68791,   10.9026,   11.0786,    9.24774,   6.06011,   6.99377,    7.06191,   6.73005,
    4.75445,     3.75562,  4.47948,   3.88179,   2.57925,    -0.241455, -0.292703, 0.451028,   -0.183554, -4.7693,
    -7.35775,    -9.05055, -8.55567,  -9.15969,  -10.1821,   -9.56669,  -8.51355,  -8.01721,   -10.1904,  -11.6768,
    -11.2362,    -8.30647, -9.93273,  -11.2495,  -11.7595,   -11.8598,  -12.2361,  -14.0704,   -14.0706,  -12.6549,
    -9.81637,    -10.2899, -15.2639,  -16.0539,  -13.291,    -9.87149,  -10.9089,  -13.9898,   -14.7389,  -12.1001,
    -8.99634,    -9.42988, -11.3537,  -12.9699,  -9.21516,   -6.72457,  -5.96632,  -10.7561,   -13.3011,  -11.811,
    -7.84478,    -7.32789, -8.66873,  -10.8726,  -9.76712,   -8.74054,  -7.59645,  -9.21074,   -10.3317,  -8.95261,
    -6.09563,    -5.97103, -7.63809,  -7.76908,  -6.79108,   -5.53455,  -5.50734,  -7.03587,   -6.9796,   -5.83728,
    -7.04902,    -8.13646, -10.2796,  -9.89936,  -9.3416,    -8.55272,  -9.26914,  -10.5683,   -11.7162,  -9.12182,
    -8.94192,    -9.75476, -11.701,   -10.1027,  -7.96955,   -7.26784,  -8.5111,   -8.62474,   -7.42749,  -4.65164,
    -6.23204,    -9.04893, -10.2988,  -5.76752,  -3.93007,   -6.03845,  -7.96042,  -7.73502,   -3.89031,  -2.45114,
    -5.68922,    -6.36723, -3.46444,  -1.85108,  -2.62985,   -4.18099,  -3.76635,  -2.39744,   -1.81296,  -3.52414,
    -5.21996,    -4.20541, -0.964009, 0.0791309, -0.468074,  -1.0992,   0.519417,  2.66487,    1.66409,   0.543407,
    1.09234,     3.38011,  2.95083,   0.819819,  -0.52162,   1.4174,    0.605319,  0.022261,   -0.62426,  -0.684749,
    -1.01914,    -2.1876,  -3.11528,  -1.58306,  0.114128,   -1.11779,  -4.11866,  -4.89873,   -3.92071,  -1.89906,
    -2.68446,    -4.22725, -6.43118,  -4.53554,  -4.39757,   -1.47731,  -2.13026,  -3.12187,   -4.12669,  -3.82813,
    -2.18146,    -1.59104, -2.04932,  -1.57833,  0.349718,   1.04036,   -0.105499, 0.219319,   4.29623,   6.74877,
    6.21717,     4.73313,  6.12465,   9.16759,   11.3755,    9.86437,   8.50226,   10.32,      14.754,    17.1337,
    14.1212,     11.2528,  12.525,    14.3051,   13.2894,    9.20449,   9.03667,   10.3547,    11.7348,   10.6885,
    8.73636,     7.60529,  8.6305,    10.9681,   11.3156,    8.83943,   8.72507,   12.1551,    13.6017,   10.8666,
    8.10075,     9.22222,  12.1008,   11.2902,   10.0623,    9.68097,   13.4675,   14.7869,    10.9685,   8.84422,
    11.7487,     14.7135,  14.5873,   12.2641,   13.8908,    15.2516,   14.5796,   10.7641,    9.66932,   10.6614,
    10.1978,     7.98695,  6.93255,   6.98925,   8.74184,    10.0769,   10.2966,   6.78023,    5.67814,   6.54964,
    8.21926,     7.78765,  4.4364,    3.48574,   4.42144,    5.9231,    3.35941,   -0.0522906, -0.979249, 1.21499,
    3.45366,     3.40318,  2.73096,   2.27258,   2.42335,    2.13037,   1.25271,   -0.711089,  -0.87326,  1.94523,
    2.6787,      0.444597, -2.06283,  -0.741667, 1.21719,    1.59351,   -1.26551,  -4.92902,   -5.41489,  -3.75084,
    -3.67814,    -5.59454, -6.24593,  -5.80943,  -5.18162,   -6.96681,  -9.00567,  -9.18818,   -8.71499,  -7.53588,
    -8.81469,    -8.41229, -5.23676,  -2.18041,  -3.56639,   -4.67908,  -4.84244,  -2.7607,    -3.91998,  -5.09383,
    -5.78113,    -5.02455, -4.64618,  -5.82443,  -7.27838,   -8.2092,   -8.22793,  -6.85023,   -6.12949,  -8.1441,
    -11.0565,    -10.8742, -7.49999,  -5.37865,  -7.60538,   -9.1533,   -8.59465,  -7.09866,   -7.73716,  -8.51578,
    -7.14109,    -5.77103, -5.80195,  -6.19756,  -4.89473,   -4.04671,  -4.85963,  -6.38138,   -5.94177,  -5.02805,
    -5.70675,    -7.52055, -8.10659,  -6.47269,  -4.94867,   -5.8885,   -8.53523,  -6.299,     -4.55272,  -5.06223,
    -6.50954,    -6.3651,  -3.8294,   -1.33969,  -0.690447,  -1.66949,  -1.77415,  -0.863641,  -0.445446, -2.38383,
    -2.70314,    -4.58643, -4.07359,  -3.42418,  -2.46067,   -3.98641,  -4.24995,  -3.14493,   -2.45151,  -3.08614,
    -4.46827,    -5.38232, -5.00127,  -5.60612,  -6.55096,   -8.17847,  -9.33258,  -8.92086,   -9.74632,  -9.9634,
    -9.49974,    -6.99243, -6.5621,   -7.43194,  -8.32843,   -6.27971,  -5.78272,  -7.00901,   -8.37082,  -7.33603,
    -4.6288,     -4.73265, -5.44714,  -5.71592,  -3.07072,   -2.59515,  -2.05666,  -4.21139,   -3.8456,   -2.96751,
    -1.12855,    -3.32555, -4.66631,  -4.33927,  -2.16567,   -1.67574,  -3.25436,  -3.74978,   -2.45332,  0.960707,
    1.02715,     -2.44935, -2.66762,  0.166573,  3.62077,    1.60331,   -1.02528,  -0.41703,   3.60018,   3.21778,
    0.0532309,   -1.55976, 0.819045,  1.1755,    -0.388379,  -0.365386, 0.434221,  0.849924,   1.34936,   1.89925,
    2.87332,     0.742747, -0.922765, 0.0124654, 4.18292,    5.55936,   3.72489,   3.32731,    3.7997,    4.7594,
    3.99982,     0.393849, -2.21144,  -2.22113,  -2.45447,   -5.40628,  -8.15082,  -6.84737,   -3.61632,  -4.19149,
    -7.42681,    -9.72228, -9.86708,  -9.33874,  -10.1524,   -12,       -14.353,   -14.3232,   -11.5522,  -11.8308,
    -14.4272,    -15.8485, -13.8451,  -12.1385,  -11.502,    -13.3125,  -13.9237,  -13.9327,   -14.1488,  -14.1093,
    -14.9721,    -16.2411, -16.1087,  -14.8598,  -14.0859,   -14.4906,  -14.2942,  -13.7718,   -13.9925,  -14.229,
    -13.0456,    -11.9646, -11.9596,  -12.2878,  -12.0963,   -11.1245,  -11.4945,  -10.4767,   -12.3236,  -13.7937,
    -14.8698,    -12.4803, -12.5159,  -13.1128,  -13.5027,   -12.4741,  -12.1013,  -11.9719,   -13.1094,  -12.6115,
    -11.6415,    -9.82913, -11.855,   -12.4263,  -12.8788,   -11.725,   -10.4145,  -8.93947,   -10.1003,  -10.7874,
    -10.7658,    -10.74,   -9.26528,  -8.96879,  -8.60259,   -6.37513,  -2.98524,  -3.67814,   -6.03185,  -7.81372,
    -7.00944,    -6.59131, -7.04624,  -7.93765,  -8.52649,   -9.45049,  -7.82704,  -8.22837,   -9.54005,  -12.8874,
    -11.232,     -10.3391, -10.5255,  -13.1384,  -12.4522,   -11.2992,  -10.3612,  -11.6451,   -11.4215,  -9.20567,
    -7.13699,    -7.6926,  -9.94848,  -10.4309,  -7.31069,   -5.98831,  -9.00164,  -11.3881,   -8.38032,  -5.61742,
    -6.16111,    -7.30488, -8.76897,  -7.75789,  -6.20106,   -6.09247,  -8.95997,  -11.8575,   -10.3722,  -9.20117,
    -8.65694,    -10.4463, -11.8931,  -11.8163,  -8.80956,   -6.60651,  -9.55294,  -11.1903,   -10.7632,  -9.59246,
    -10.9303,    -14.1307, -14.6162,  -11.9256,  -11.4294,   -14.1227,  -15.3324,  -13.3052,   -12.0878,  -12.1335,
    -12.6285,    -13.6016, -13.9441,  -14.4245,  -13.5043,   -16.075,   -17.619,   -15.5123,   -11.9789,  -10.4461,
    -14.4694,    -14.6635, -13.276,   -10.9654,  -13.1188,   -16.3155,  -16.4632,  -14.1056,   -15.6607,  -18.7849,
    -20.9127,    -20.6696, -19.3832,  -19.652,   -20.7142,   -21.6051,  -22.5018,  -22.6008,   -20.7462,  -20.6104,
    -21.8306,    -25.7877, -25.8696,  -23.4782,  -21.8196,   -23.4579,  -23.1486,  -22.4924,   -19.3276,  -17.3143,
    -18.4517,    -19.7227, -20.5887,  -19.8319,  -19.9942,   -21.6976,  -24.7583,  -26.3022,   -24.1429,  -23.0637,
    -25.596,     -28.9256, -27.4639,  -24.6531,  -23.2179,   -25.4542,  -25.7244,  -26.9354,   -28.0725,  -27.7764,
    -29.1182,    -29.645,  -30.4409,  -30.9854,  -31.7086,   -33.0169,  -33.9927,  -33.9239,   -33.0698,  -34.525,
    -37.143,     -37.9698, -36.6539,  -36.5242,  -37.6408,   -37.9838,  -38.4443,  -38.4391,   -38.7109,  -41.0617,
    -41.9755,    -41.1273, -39.2596,  -41.554,   -42.5789,   -42.151,   -39.534,   -41.1185,   -41.5864,  -41.8831,
    -39.5701,    -40.6223, -41.5535,  -44.1914,  -43.849,    -42.5745,  -41.9454,  -43.8638,   -44.6531,  -43.2043,
    -42.5599,    -45.1062, -46.5762,  -45.5266,  -43.863,    -44.6714,  -46.8966,  -46.6627,   -45.0059,  -44.0229,
    -44.7762,    -45.3116, -44.0622,  -44.3306,  -44.8996,   -45.2513,  -43.8117,  -43.1199,   -42.2809,  -42.1324,
    -40.3382,    -40.1806, -42.4776,  -45.2179,  -44.8535,   -43.8437,  -42.9161,  -43.0175,   -41.5753,  -41.8477,
    -41.3694,    -40.6872, -41.1838,  -41.1628,  -41.747,    -41.0803,  -41.0728,  -41.8368,   -42.5375,  -40.7291,
    -38.68,      -38.3749, -39.5626,  -40.0342,  -40.895,    -38.6894,  -38.1381,  -37.3086,   -38.4922,  -38.1503,
    -39.2915,    -39.7951, -39.8553,  -38.6247,  -39.127,    -40.7875,  -41.147,   -39.4582,   -38.6229,  -38.3117,
    -39.0824,    -36.5746, -35.8961,  -35.7241,  -35.4128,   -34.1662,  -33.6578,  -34.8008,   -34.5951,  -32.8775,
    -31.7472,    -30.4881, -31.2414,  -30.8585,  -29.5218,   -27.4727,  -28.1311,  -30.0443,   -33.0005,  -33.0222,
    -32.4741,    -31.6032, -33.4837,  -35.7992,  -35.5987,   -34.4458,  -33.1233,  -36.4592,   -38.8968,  -38.7359,
    -36.2382,    -35.3646, -37.8735,  -38.3303,  -37.8054,   -36.3328,  -37.7996,  -38.7921,   -39.3408,  -38.2412,
    -36.2579,    -35.5304, -38.484,   -39.4946,  -39.228,    -36.9026,  -38.1743,  -40.7769,   -41.7075,  -38.6996,
    -35.0928,    -36.6918, -40.0478,  -40.1989,  -35.5285,   -33.8812,  -35.3921,  -38.1138,   -37.1166,  -35.4768,
    -35.3462,    -36.5135, -37.519,   -37.5886,  -38.7189,   -39.457,   -40.5375,  -40.0317,   -39.2927,  -37.8168,
    -39.1008,    -41.3449, -41.2154,  -38.7638,  -37.4997,   -40.2999,  -42.9002,  -42.8801,   -39.5922,  -38.5705,
    -40.8961,    -42.6095, -40.8147,  -38.6909,  -38.7715,   -39.344,   -39.504,   -38.3313,   -38.4319,  -38.3984,
    -38.2426,    -38.4949, -37.4668,  -37.5437,  -36.1589,   -36.2023,  -35.4132,  -36.5653,   -34.8031,  -34.3762,
    -34.1696,    -34.7241, -32.7471,  -31.5448,  -31.9062,   -34.7414,  -34.5286,  -31.1617,   -30.6527,  -32.014,
    -34.0209,    -33.7951, -30.8681,  -29.5866,  -31.2953,   -33.1206,  -31.707,   -30.3741,   -31.2128,  -32.4106,
    -32.3049,    -29.919,  -27.3308,  -25.8176,  -26.0414,   -25.077,   -24.7511,  -24.9018,   -25.0369,  -25.4418,
    -25.4346,    -25.9081, -27.5487,  -28.0477,  -29.2525,   -30.5617,  -32.4195,  -32.0946,   -31.0442,  -30.1395,
    -30.9546,    -32.1209, -30.4246,  -30.3696,  -31.3718,   -32.0118,  -29.5235,  -29.6165,   -29.5384,  -30.0034,
    -28.7398,    -28.8582, -28.7726,  -29.1246,  -27.3887,   -26.996,   -28.6934,  -30.0959,   -28.7692,  -27.635,
    -29.1321,    -30.067,  -29.4092,  -27.2232,  -27.2461,   -28.8812,  -29.6944,  -27.8106,   -27.1538,  -28.4559,
    -29.8039,    -29.4136, -27.6357,  -26.6882,  -26.851,    -27.2102,  -26.074,   -23.2096,   -21.8121,  -22.2624,
    -23.011,     -22.9349, -21.646,   -20.887,   -21.4815,   -22.9732,  -24.7179,  -25.7694,   -24.7753,  -24.961,
    -25.7089,    -25.8072, -24.9761,  -24.016,   -24.1293,   -25.3014,  -24.5138,  -22.8391,   -23.56,    -24.405,
    -23.9387,    -22.7575, -22.4235,  -23.0278,  -23.3293,   -24.0534,  -23.7359,  -23.6546,   -23.0506,  -22.6675,
    -22.1875,    -24.3732, -26.0688,  -27.1981,  -26.8777,   -26.5299,  -26.9193,  -26.3544,   -24.6056,  -23.8174,
    -24.6947,    -25.1752, -22.1718,  -21.3082,  -22.1013,   -23.0958,  -21.8591,  -20.805,    -19.2282,  -18.8907,
    -18.3436,    -17.7483, -17.8502,  -18.9091,  -19.6056,   -19.9548,  -19.1727,  -19.3822,   -19.7211,  -19.4691,
    -19.911,     -18.3067, -17.3392,  -18.1031,  -18.9671,   -18.7352,  -17.1791,  -16.2531,   -16.5707,  -16.3838,
    -16.7903,    -15.4905, -14.5882,  -12.5137,  -12.5954,   -14.0892,  -13.7921,  -10.7175,   -9.74835,  -10.9641,
    -12.5628,    -11.4739, -8.1905,   -6.96551,  -9.26433,   -12.0003,  -11.177,   -9.23484,   -10.0663,  -12.9438,
    -13.5278,    -13.6318, -13.1926,  -12.5553,  -12.8283,   -13.8336,  -15.8128,  -15.1939,   -16.3083,  -16.084,
    -17.1351,    -17.886,  -18.3539,  -18.0413,  -16.5043,   -16.9358,  -17.4124,  -18.3266,   -17.6637,  -14.9154,
    -12.9043,    -14.6949, -15.7046,  -14.2006,  -10.9709,   -13.2923,  -16.4497,  -16.7456,   -13.5034,  -12.7048,
    -13.8489,    -15.2316, -14.7754,  -12.9142,  -11.9259,   -12.6738,  -14.147,   -12.7889,   -11.4146,  -11.0985,
    -10.7991,    -8.67082, -7.06612,  -5.70541,  -4.92032,   -5.01411,  -5.45519,  -4.96863,   -5.11225,  -4.44247,
    -3.94859,    -5.14282, -6.18437,  -6.68643,  -7.65277,   -6.62485,  -5.27168,  -6.64129,   -8.13625,  -7.01422,
    -6.55956,    -6.70695, -8.25354,  -7.4485,   -5.39595,   -3.031,    -4.56192,  -8.01759,   -7.41791,  -4.58084,
    -2.64115,    -3.22322, -4.14206,  -2.63547,  -1.51366,   -1.96097,  -3.67429,  -4.91386,   -4.37214,  -4.00668,
    -4.68029,    -4.83448, -5.46184,  -5.93854,  -5.53964,   -4.60502,  -4.489,    -4.81776,   -5.37529,  -6.403,
    -7.56326,    -8.01028, -8.23111,  -8.58005,  -7.31983,   -8.27996,  -9.49992,  -9.81422,   -9.12967,  -9.83094,
    -10.2212,    -10.2133, -9.30801,  -8.17061,  -10.1693,   -11.2662,  -9.91311,  -7.36292,   -5.90298,  -8.56156,
    -10.6509,    -9.27851, -6.60209,  -6.66612,  -9.25043,   -10.8588,  -9.14266,  -9.0778,    -11.9409,  -14.0469,
    -12.8047,    -10.0932, -10.246,   -13.0942,  -14.8219,   -12.7719,  -10.2393,  -10.3921,   -14.0638,  -14.7263,
    -12.5264,    -10.7295, -12.0031,  -14.4443,  -14.0444,   -11.5233,  -10.4894,  -10.5924,   -10.1904,  -10.7243,
    -9.7598,     -7.39853, -7.56473,  -8.04634,  -9.64142,   -8.27349,  -7.09967,  -6.29337,   -7.55408,  -7.14421,
    -4.17835,    -3.17082, -4.49929,  -5.062,    -5.1336,    -4.20957,  -3.23286,  -5.30119,   -7.94957,  -8.53484,
    -7.69721,    -7.33513, -7.86607,  -8.53747,  -8.39934,   -7.39065,  -6.77183,  -7.47785,   -8.15569,  -6.73406,
    -6.7177,     -6.86579, -8.11411,  -9.5214,   -9.31914,   -9.64458,  -10.8366,  -11.6734,   -10.4953,  -9.08093,
    -7.04891,    -6.07143, -6.60442,  -4.89427,  -2.86787,   0.0245173, -1.23144,  -2.60081,   -3.64113,  -2.15619,
    -0.44086,    1.06484,  2.23457,   1.98713,   3.4035,     5.99616,   6.70247,   6.44403,    5.73699,   5.98192,
    7.73938,     8.65574,  7.65906,   5.73063,   7.35171,    9.55862,   8.69226,   7.31855,    5.31466,   5.93893,
    6.50922,     6.32165,  6.14192,   7.681,     8.7921,     9.09865,   8.53079,   8.65088,    8.98312,   10.7953,
    11.9114,     11.6342,  13.4874,   16.1642,   17.6873,    16.5841,   17.1418,   18.6208,    18.9647,   19.278,
    18.4277,     17.1536,  17.785,    20.4695,   21.425,     20.9517,   20.4442,   20.6907,    21.1447,   21.2028,
    19.6834,     19.8478,  19.964,    20.8286,   21.55,      21.5696,   20.442,    17.558,     15.845,    18.7524,
    19.4562,     14.6714,  9.84557,   11.7498,   14.8807,    14.8,      11.4846,   11.2387,    15.8297,   19.1495,
    17.2534,     15.0219,  18.3861,   21.6146,   21.0892,    16.598,    16.8823,   18.9278,    21.1896,   20.2354,
    18.9252,     18.0077,  20.0135,   19.6243,   18.4322,    18.7351,   18.4568,   17.5194,    15.9895,   14.4362,
    14.3524,     13.4177,  12.6057,   9.85806,   9.66949,    9.52823,   9.2093,    9.25841,    9.17535,   11.4052,
    12.8777,     12.5382,  10.6089,   10.0143,   12.2386,    13.4646,   13.2936,   11.0971,    9.29366,   10.2019,
    13.8131,     13.7541,  11.1554,   9.27561,   10.8554,    13.4614,   13.4117,   10.3289,    7.55042,   9.08554,
    11.2329,     10.6593,  8.41733,   5.19226,   4.55358,    7.0979,    7.68245,   6.32427,    3.37364,   4.34866,
    6.80715,     7.91944,  6.66081,   5.45392,   6.0856,     8.65793,   9.34422,   10.3886,    9.38338,   7.25136,
    4.83495,     5.50941,  6.63024,   8.01388,   6.39507,    4.94575,   4.04152,   5.97255,    6.9239,    5.32913,
    4.20286,     5.70123,  4.60204,   2.5888,    -0.0851128, 0.551474,  3.92616,   5.50466,    4.06752,   2.13222,
    3.31378,     8.06114,  9.45448,   6.66955,   3.02567,    6.38744,   11.5072,   13.0002,    10.2397,   9.46404,
    10.24,       12.086,   10.6068,   8.94754,   9.16604,    11.5748,   12.2609,   12.0632,    10.2852,   9.96771,
    10.2177,     9.45447,  8.87991,   8.32066,   9.61608,    8.03225,   6.31031,   3.68858,    2.25231,   2.42579,
    4.34949,     4.25534,  2.70124,   1.97779,   -0.97026,   -0.316197, 1.05901,   1.88641,    -2.33558,  -2.32636,
    -0.00701054, 1.39435,  -0.170101, -1.75681,  0.146831,   1.24915,   1.86487,   -0.52854,   -0.376534, 1.43194,
    3.10168,     1.48054,  -0.636448, 0.548911,  0.570599,   0.453317,  -1.10184,  0.178239,   0.968689,  2.00547,
    1.25585,     1.58132,  1.74383,   3.31385,   4.23086,    5.21691,   5.23764,   6.37995,    6.67323,   5.34751,
    4.22562,     5.49398,  6.32383,   4.26022,   3.73278,    4.04188,   3.07312,   2.79903,    2.91956,   4.72303,
    6.12711,     6.5146,   6.19003,   5.54209,   5.35546,    5.9083,    5.22238,   5.20156,    4.67671,   3.56179,
    1.51764,     2.26265,  5.15405,   5.59138,   5.0983,     4.21278,   3.79838,   3.47377,    3.7424,    4.1367,
    4.71237,     4.21459,  3.98038,   4.55196,   4.66077,    4.04647,   1.68832,   0.746771,   1.08009,   3.40762,
    4.46355,     3.80912,  2.62563,   5.44603,   9.0098,     10.3509,   8.92618,   8.03205,    7.85415,   8.94455,
    8.76311,     7.20169,  6.54734,   7.91414,   11.542,     11.5622,   9.31215,   9.25019,    11.8917,   12.1215,
    12.6295,     10.378,   9.93374,   10.7987,   12.4366,    13.1182,   13.5185,   13.2395,    12.097,    10.1466,
    9.56719,     11.1964,  12.7763,   11.5461,   8.93911,    8.16588,   9.42133,   9.21427,    7.37058,   7.2322,
    6.60058,     7.56028,  8.44081,   8.5424,    7.02759,    8.00339,   8.3952,    8.83128,    9.0106,    8.20166,
    7.46689,     5.9864,   5.36052,   3.10384,   2.08333,    1.38316,   1.50204,   -0.509614,  -0.678868, 0.31758,
    2.12909,     0.964061, -0.202979, -1.6362,   -0.963685,  -0.465444, -0.35077,  -0.932439,  -0.663251, 1.37306,
    1.4846,      0.790636, 2.70086,   5.84894,   6.79392,    5.21905,   5.28933,   6.32482,    6.29194,   6.50455,
    7.1641,      7.89487,  8.12179,   8.6264,    7.24988,    7.59397,   7.47298,   7.92312,    7.41166,   6.8137,
    5.64875,     5.49847,  5.63509,   6.8609,    5.71708,    5.83148,   6.66015,   6.70486,    8.25118,   7.67602,
    6.8512,      4.97026,  5.09031,   6.61807,   7.76647,    5.27236,   3.94877,   5.5479,     8.2798,    7.34271,
    5.51075,     5.09451,  7.9287,    9.72442,   9.35057,    7.92384,   8.43373,   10.2695,    10.2984,   10.0723,
    8.92473,     9.16515,  10.8267,   12.4844,   11.9284,    9.87895,   8.24865,   7.52319,    8.16953,   8.71872,
    8.68082,     7.0557,   5.73073,   7.96759,   9.73285,    9.85042,   6.91783,   5.83563,    8.8987,    10.7485,
    10.4347,     8.35009,  8.1633,    10.1092,   12.3422,    11.2553,   9.73668,   8.97662,    10.8012,   12.8139,
    13.7135,     13.4129,  13.399,    13.2934,   13.7535,    13.9797,   14.1308,   16.9558,    17.0897,   16.1521,
    15.5898,     18.4834,  20.4694,   19.3741,   17.5733,    18.0174,   22.6723,   25.0229,    23.4303,   18.9261,
    20.1408,     24.5085,  25.0821,   20.4004,   15.2774,    15.7088,   18.1611,   16.9331,    12.258,    10.9202,
    12.7436,     15.358,   12.3043,   9.82003,   10.6448,    14.0971,   14.1754,   12.0648,    9.44142,   9.85165,
    11.4642,     10.181,   8.64067,   9.56851,   12.0086,    12.2899,   11.0388,   10.916,     13.334,    14.3392,
    15.6946,     14.5396,  16.072,    18.421,    19.2834,    18.7442,   18.7254,   19.7741,    21.4662,   22.5682,
    21.5819,     21.5586,  23.543,    24.9287,   23.4236,    22.4159,   22.5847,   23.5157,    25.0636,   23.8645,
    22.1762,     24.8175,  26.4259,   25.0512,   23.886,     23.8657,   24.63,     24.842,     25.2279,   24.235,
    24.9064,     26.0435,  25.7331,   25.538,    24.601,     24.8048,   23.562,    22.4102,    22.1563,   23.1836,
    21.3748,     20.1023,  20.62,     21.6032,   21.1153,    18.8774,   16.4887,   16.2532,    19.2569,   19.9517,
    18.341,      17.1907,  18.4415,   21.1543,   20.781,     19.2705,   17.1176,   19.1996,    19.9059,   19.6404,
    17.8959,     15.9296,  15.3587,   16.3487,   15.6155,    15.0532,   14.7167,   16.1373,    14.6943,   13.4991,
    12.907,      14.2237,  15.597,    15.7522,   12.734,     11.6437,   14.6732,   17.036,     15.5026,   13.9768,
    14.1935,     15.6295,  17.5567,   17.0834,   16.0999,    16.0426,   16.0299,   15.1157,    13.1831,   11.7549,
    13.9427,     13.9369,  14.0228,   10.547,    9.64562,    10.1519,   12.0165,   12.3618,    11.1648,   10.216,
    12.299,      14.6001,  15.1221,   14.4729,   16.3023,    18.1483,   17.2407,   14.4654,    13.1347,   14.2578,
    15.0341,     14.8712,  12.8008,   12.9213,   15.7037,    17.513,    17.4291,   15.6866,    15.331,    16.8136,
    18.7538,     18.9723,  19.0773,   17.3151,   16.117,     16.6881,   15.2966,   14.4803,    11.223,    10.0756,
    11.1215,     12.0666,  10.3887,   9.28164,   10.9436,    14.568,    16.2398,   15.068,     14.4526,   14.8001,
    16.2408,     16.2231,  15.7972,   15.0108,   14.3914,    13.2985,   12.2118,   12.5711,    12.602,    11.6622,
    11.5277,     11.8235,  13.3378,   14.1418,   13.8304,    12.978,    13.3,      14.7098,    15.027,    12.737,
    11.659,      12.3973,  13.1499,   11.1758,   9.59453,    9.20775,   11.491,    11.843,     11.4916,   11.0308,
    11.4102,     12.1163,  10.6817,   9.14883,   8.63285,    9.64305,   9.85436,   10.1939,    10.2271,   10.4643,
    10.6668,     8.78037,  8.38194,   7.83248,   7.967,      6.475,     3.98821,   4.06583,    5.37295,   5.69773,
    2.92853,     0.311514, 1.18225,   3.89707,   4.21536,    -1.59948,  -5.66164,  -4.83722,   -2.16486,  -5.47697,
    -8.62283,    -9.25048, -6.76258,  -6.27412,  -7.3027,    -9.48235,  -7.01672,  -4.66398,   -3.753,    -5.33697,
    -5.38453,    -3.96925, -3.03644,  -5.78968,  -8.40427,   -8.97228,  -6.13778,  -4.71308,   -6.5068,   -7.9825,
    -6.29268,    -5.62025, -4.59369,  -6.10704,  -5.32264,   -3.44181,  0.916561,  -1.16683,   -2.37116,
};
const test_input_test2 = [_]f64{
    0.124663,   1.2274,     -0.692594,  0.816833,    0.0454217,  -1.44616,   0.426407,    -0.219062,  -0.871184,  -0.420529,
    -1.875,     -1.79552,   -2.02713,   -1.69483,    -1.12224,   -3.57315,   -1.81474,    -2.64564,   -3.67606,   -1.73465,
    -2.72511,   -3.35671,   -2.55614,   -1.25729,    -2.20018,   -2.04335,   -2.6433,     -0.674693,  -1.90935,   -1.06685,
    -1.34139,   0.944873,   1.16278,    0.586521,    1.90126,    1.64907,    1.41359,     1.73666,    1.42526,    0.850299,
    1.42375,    2.01816,    3.14526,    1.71761,     2.97834,    3.65839,    4.00663,     2.83459,    2.86493,    3.43688,
    2.50931,    3.47046,    3.22459,    1.78915,     0.924592,   0.822544,   0.386951,    0.650993,   -0.736471,  -0.250501,
    0.865228,   -0.135077,  0.295066,   1.71046,     0.566839,   0.76785,    1.26391,     -0.271361,  -1.3708,    -1.31563,
    -0.32655,   -2.84248,   -3.30454,   -3.00202,    -5.03529,   -5.52404,   -4.61448,    -5.06921,   -5.53117,   -5.75022,
    -6.60642,   -6.04153,   -4.95923,   -5.16244,    -5.94849,   -3.89675,   -4.28679,    -3.6313,    -1.79424,   -2.916,
    -1.45487,   -0.461141,  -0.767656,  -0.666949,   -0.212166,  -0.251976,  1.35035,     1.76689,    2.28385,    2.7887,
    3.49312,    2.75701,    4.03158,    2.51268,     2.83262,    3.59451,    2.38647,     2.19838,    0.629051,   0.136845,
    0.244548,   -1.02417,   -0.545968,  -1.09755,    -2.06444,   -1.45569,   -3.04734,    -3.49761,   -1.93698,   -3.47086,
    -2.89898,   -2.9294,    -2.40258,   -2.79061,    -2.06659,   -3.49424,   -1.77158,    -3.66325,   -3.17298,   -2.5045,
    -2.86094,   -3.45203,   -2.17179,   -1.4939,     -2.26222,   -1.94471,   -2.39292,    -0.119844,  -1.46277,   -1.45441,
    0.453447,   0.717021,   0.0100505,  2.07862,     2.31047,    2.11683,    2.4839,      0.555062,   -0.522529,  0.271041,
    -0.523389,  -0.464596,  -0.329427,  -1.36334,    -1.99477,   -3.04258,   -3.73112,    -4.71278,   -5.25836,   -5.54619,
    -6.79229,   -6.64601,   -5.69137,   -5.51775,    -6.05358,   -5.15735,   -4.74923,    -4.42598,   -3.08362,   -3.52618,
    -4.19838,   -3.56504,   -3.8582,    -3.13806,    -4.12002,   -2.75363,   -0.83476,    -2.78346,   -2.48539,   -1.98539,
    -3.21122,   -2.10886,   -3.04553,   -3.70863,    -3.85759,   -2.13841,   -2.39966,    -2.22813,   -2.7304,    -2.90696,
    -2.92965,   -4.00097,   -4.85197,   -4.41936,    -3.39712,   -5.94229,   -5.8189,     -4.87809,   -4.68869,   -4.00493,
    -5.26481,   -5.04943,   -5.88372,   -5.1805,     -4.50725,   -5.24913,   -4.30755,    -3.61449,   -2.57362,   -1.29398,
    -3.59565,   -2.341,     -0.490767,  0.641621,    0.650178,   1.52301,    2.1378,      2.96823,    3.29036,    3.97937,
    3.35866,    1.25498,    0.101147,   -0.686592,   -1.4285,    -0.810703,  0.124401,    -0.0625823, -0.780775,  -1.0948,
    -1.24241,   -1.72731,   -1.42197,   -3.85501,    -6.23869,   -4.74888,   -5.28131,    -5.66609,   -6.3137,    -7.67891,
    -8.04995,   -7.62861,   -9.44627,   -8.38112,    -7.53564,   -7.4172,    -7.05515,    -5.20631,   -5.73981,   -6.26952,
    -5.43563,   -4.64058,   -3.15017,   -2.91189,    -2.48896,   -2.31891,   -3.81081,    -2.46811,   -2.23903,   -1.35855,
    -1.2648,    -1.61202,   -2.76833,   -2.07232,    -2.29518,   -3.46669,   -3.90809,    -5.12107,   -4.15208,   -4.21242,
    -5.10437,   -3.28664,   -3.65553,   -4.29934,    -2.99153,   -2.30094,   -2.89703,    -3.15488,   -2.78629,   -2.83885,
    -1.22123,   -2.39696,   -1.09096,   -1.0865,     0.378697,   -0.571742,  -0.56909,    -2.47041,   -2.02389,   -2.1707,
    -2.4977,    -2.15665,   -0.864566,  -0.381262,   1.27016,    -0.153786,  1.07099,     1.61328,    2.21382,    1.20446,
    2.0276,     2.50673,    3.11014,    1.38509,     2.0228,     0.603682,   -0.645378,   -1.22506,   -0.456994,  -0.520576,
    -1.26325,   -0.590333,  -0.751763,  1.18829,     0.88484,    0.850131,   1.9167,      2.83579,    1.87688,    3.66388,
    2.40409,    2.57957,    4.13613,    1.00817,     0.863573,   0.681111,   0.0822985,   -0.697664,  -0.389217,  -0.0014408,
    0.24244,    0.123114,   0.247544,   1.04629,     1.36748,    1.36561,    2.50722,     2.16657,    1.4567,     3.90587,
    1.05891,    2.38499,    2.90996,    2.85391,     2.68811,    0.0479036,  2.08724,     0.959835,   1.75001,    3.14473,
    1.05326,    0.382868,   2.44425,    2.5329,      3.48683,    4.72862,    4.07186,     4.12753,    4.3293,     4.36487,
    4.25866,    5.13765,    3.71671,    4.11783,     4.64705,    3.96776,    3.12788,     3.12717,    3.21838,    3.23705,
    1.45164,    1.43073,    0.864648,   2.41228,     1.97082,    0.916343,   -0.737916,   -0.116669,  -0.819045,  -0.914337,
    -1.96925,   -0.66679,   -0.868121,  -0.79945,    -2.13597,   -0.976,     -0.829096,   0.506593,   -0.534234,  2.25336,
    2.13538,    1.27013,    -0.181085,  1.22678,     2.07398,    1.14908,    1.36445,     3.00618,    1.99053,    1.42726,
    1.77402,    2.86498,    1.1506,     2.6245,      -0.198325,  0.579936,   0.763666,    -0.999984,  -0.0476686, 1.14122,
    0.349778,   3.99493,    2.59256,    2.54155,     1.97066,    1.0348,     1.18556,     0.640136,   -1.05838,   -2.2893,
    -1.87342,   -1.74431,   -1.46176,   -1.79641,    -2.0907,    -2.23393,   -1.50393,    -0.809592,  -1.05804,   -1.02313,
    0.285105,   0.308347,   0.688473,   1.52432,     0.516637,   0.863049,   1.22561,     -0.201735,  -0.820267,  0.87083,
    1.15549,    0.528681,   1.83274,    1.48228,     -0.119458,  -0.393215,  -0.723904,   -1.02844,   -2.62099,   -3.60425,
    -3.66127,   -4.72935,   -4.42068,   -3.76768,    -5.68684,   -7.2439,    -7.37909,    -7.59458,   -6.50544,   -7.52451,
    -8.74701,   -8.69002,   -8.60633,   -10.2387,    -9.72939,   -8.57978,   -8.33085,    -7.34534,   -6.32593,   -5.53717,
    -4.89497,   -5.23343,   -2.11626,   -3.31056,    -3.09087,   0.279461,   1.21582,     2.52775,    1.89094,    3.54942,
    4.45283,    3.50697,    3.71926,    4.64674,     5.92162,    4.68789,    5.19091,     5.00786,    4.95714,    4.12983,
    3.47302,    3.97759,    2.30699,    2.92682,     2.08196,    2.802,      1.98154,     1.84464,    3.25609,    2.79831,
    3.32527,    3.43583,    2.37444,    4.92924,     4.11577,    3.36849,    2.22038,     1.1731,     2.29488,    3.08411,
    2.2566,     1.25295,    1.11747,    1.93322,     1.13703,    1.27601,    3.34346,     2.89357,    2.52023,    3.08026,
    2.9888,     1.97745,    2.14192,    4.02048,     2.51823,    2.6318,     4.0564,      4.1979,     2.17617,    4.78495,
    3.48851,    1.67941,    1.35678,    0.861481,    0.57843,    -1.27457,   -1.75009,    -1.05192,   -1.33995,   -1.92317,
    -2.7526,    -1.87693,   -0.950385,  -0.808734,   -2.31412,   -2.33807,   -1.62452,    -1.79889,   -1.8698,    -1.80609,
    -1.25224,   -3.07173,   -3.06324,   -3.2687,     -6.16966,   -4.03204,   -5.54081,    -4.98149,   -4.69483,   -4.62679,
    -4.3356,    -2.76886,   -2.90221,   -1.51092,    -2.70017,   -1.98773,   -1.67944,    -2.71885,   -2.48582,   -3.02129,
    -0.984558,  -2.36131,   -3.20492,   -3.33627,    -5.07912,   -6.05464,   -5.23979,    -4.99535,   -5.07247,   -5.6082,
    -6.41289,   -4.96647,   -5.65815,   -6.00418,    -5.37464,   -4.78401,   -3.85079,    -2.37171,   -0.858724,  -1.25051,
    -1.02279,   0.248885,   2.26627,    2.67699,     2.84858,    3.67387,    2.86443,     3.51869,    4.97141,    5.53464,
    5.32618,    7.15634,    6.95079,    5.13129,     4.46494,    6.03911,    6.5791,      5.93832,    5.88814,    4.24776,
    4.02594,    3.57956,    2.74616,    2.34403,     1.95862,    1.452,      2.34242,     1.75357,    2.3747,     2.29117,
    0.858087,   2.15734,    1.49284,    0.429382,    -0.821501,  0.128974,   -2.05757,    0.185967,   -1.80779,   -0.694627,
    -1.67347,   -1.41612,   -1.93447,   -0.30756,    -0.745519,  -1.60251,   -1.1138,     0.511493,   0.310006,   -2.01983,
    0.429704,   -0.181718,  -0.165501,  0.544536,    2.24012,    0.734362,   0.814747,    -0.485931,  0.0169074,  -1.61299,
    -2.86141,   -3.51069,   -3.71292,   -4.13011,    -4.51397,   -4.13468,   -4.84782,    -4.72924,   -6.2162,    -5.10142,
    -6.39661,   -5.72454,   -5.94098,   -4.43024,    -4.36171,   -4.69801,   -5.44651,    -4.17203,   -3.79125,   -3.8855,
    -3.80357,   -4.03717,   -3.44513,   -3.24059,    -2.43602,   -1.69565,   -2.8617,     -1.4951,    -1.08035,   -1.35414,
    -1.13551,   1.27297,    0.920265,   0.0817683,   0.663191,   0.0645392,  0.334252,    1.14518,    1.05073,    0.667472,
    0.959366,   -0.114634,  -0.399112,  -0.798392,   -1.47523,   -0.450166,  -1.56215,    0.17393,    -1.88649,   -3.2802,
    -3.56995,   -2.80063,   -2.9865,    -3.88737,    -3.87736,   -3.86625,   -2.80121,    -3.00797,   -2.25853,   -2.46431,
    -1.6452,    -2.34875,   -3.12171,   -2.45358,    -1.92622,   -0.420573,  -0.530553,   0.425036,   -1.04139,   0.572198,
    0.988191,   0.314044,   1.62325,    1.3525,      2.7014,     1.47489,    1.31745,     1.64853,    2.84432,    1.46796,
    2.73116,    2.32019,    0.497506,   0.824845,    -0.323752,  0.267162,   -0.810307,   -0.860576,  -0.267437,  -0.269231,
    -2.05546,   -1.04036,   -0.524941,  -2.20624,    -3.78387,   -1.34063,   -3.46238,    -5.6335,    -6.69317,   -5.64519,
    -7.0368,    -8.43301,   -6.90779,   -6.93485,    -7.59501,   -8.55323,   -5.55448,    -5.60833,   -6.99958,   -6.83132,
    -6.74279,   -7.87083,   -7.77663,   -6.51524,    -10.0847,   -9.48705,   -9.66964,    -9.51961,   -8.87294,   -10.0373,
    -10.3787,   -8.53661,   -7.92937,   -8.29631,    -8.3452,    -8.03802,   -7.43616,    -6.30265,   -3.66988,   -5.37564,
    -3.50797,   -3.29794,   -2.59083,   -1.80662,    -0.455945,  0.667073,   -1.02528,    2.2529,     1.14713,    0.967698,
    -1.04286,   0.139082,   -1.94139,   -0.74614,    -2.92438,   -2.64037,   -0.92817,    -2.09567,   -3.38793,   -2.44725,
    -1.77104,   -1.28327,   -0.0308506, 1.18562,     1.17435,    2.46073,    2.88905,     3.1333,     2.95196,    3.14068,
    3.1419,     3.39124,    2.79351,    2.84715,     3.86642,    1.90981,    3.15856,     1.68297,    1.86942,    0.808669,
    0.52633,    0.174001,   0.106628,   -0.0650736,  -0.996778,  1.18094,    -0.728581,   -0.13467,   1.56157,    -0.427568,
    -2.41026,   -0.948607,  -0.535662,  -1.58723,    -2.48176,   -2.72354,   -2.58106,    -2.20292,   -1.3793,    -3.44984,
    -4.35337,   -2.97232,   -1.18639,   -1.60502,    -2.24355,   -1.05066,   -0.826239,   1.45586,    0.142256,   -1.00905,
    -1.01113,   -1.51511,   -3.15114,   -4.76483,    -3.31638,   -3.4355,    -2.99527,    -2.29775,   -0.146159,  -2.1971,
    -1.53888,   -0.612851,  -0.109464,  -1.064,      -1.30181,   -0.463903,  -1.1001,     -0.964888,  -1.62934,   0.0650192,
    0.28109,    -1.11924,   -1.28631,   0.13242,     -1.37704,   -1.45007,   -1.05484,    -1.91156,   -1.56362,   -0.57056,
    -2.17267,   -1.64846,   -2.02883,   -3.0942,     -1.63649,   -0.922648,  -0.900849,   -1.81894,   -0.669106,  -0.0446142,
    0.513145,   2.32494,    4.68771,    2.55638,     2.58132,    3.3386,     3.0351,      2.22609,    1.66112,    1.48127,
    1.1987,     0.512088,   -0.903826,  -0.205688,   -0.662401,  -1.72861,   -2.34967,    -1.68638,   -2.50387,   -2.62542,
    -2.22319,   -2.41047,   -4.37974,   -4.24234,    -4.91018,   -5.40641,   -4.91505,    -3.67589,   -3.481,     -0.972795,
    -0.163184,  -1.6778,    -0.127815,  0.381345,    0.686862,   1.45517,    0.854655,    -0.231213,  -0.180148,  1.5678,
    2.69581,    1.4602,     3.1119,     1.63887,     2.432,      2.99902,    1.23545,     1.24725,    2.44967,    1.65934,
    1.34842,    2.00835,    2.30586,    4.07405,     3.19319,    4.46962,    5.64359,     4.77058,    5.37035,    3.8606,
    3.92654,    5.06098,    4.14272,    3.53102,     3.40392,    4.428,      3.49752,     3.63622,    4.4226,     3.80611,
    3.94223,    2.77559,    2.8696,     3.29378,     3.95881,    2.6234,     1.30081,     1.34819,    1.67095,    3.33721,
    1.24125,    2.38076,    1.84183,    2.09007,     0.688936,   0.752887,   1.79279,     1.90172,    1.89804,    2.15224,
    0.244202,   2.04685,    0.0778621,  -0.743374,   0.335718,   -0.135506,  0.343895,    -1.77606,   1.13117,    0.884085,
    1.43732,    0.331897,   -0.33921,   0.15265,     -2.28274,   -3.13967,   -1.1835,     -2.0129,    -2.4137,    -2.76429,
    -1.34883,   -0.265375,  -1.992,     -1.23043,    -0.750651,  1.65839,    0.580579,    1.36225,    1.71932,    1.3373,
    1.01644,    -1.37736,   0.511318,   -1.22545,    -1.02524,   -2.23441,   -2.29634,    -2.27667,   -2.95939,   -1.20387,
    0.259413,   1.71316,    1.0546,     0.745618,    3.44892,    3.54017,    4.83027,     5.5771,     6.13045,    6.56832,
    6.23618,    6.78732,    7.21144,    5.95925,     5.75205,    5.6882,     3.9803,      3.08466,    3.02085,    1.0316,
    0.307744,   -1.72158,   -2.65128,   -4.28926,    -2.51676,   -3.57747,   -2.79226,    -1.4127,    -3.26879,   -2.51663,
    -1.3254,    -1.34941,   -2.40995,   -1.56137,    -1.58315,   -2.26872,   -2.11977,    -1.56968,   -1.12081,   -1.6941,
    -0.410528,  -0.237689,  -2.79139,   -1.62749,    -0.957536,  -3.40256,   -2.0517,     -2.25323,   -1.6298,    -1.3513,
    -1.08208,   -0.504109,  -0.207053,  -0.985225,   0.093082,   -0.920326,  0.350392,    -0.428694,  0.295824,   -1.79183,
    -0.395086,  0.175808,   0.196176,   1.79581,     1.59902,    2.6132,     1.87241,     1.97422,    2.14574,    2.38772,
    2.97482,    1.7773,     2.35894,    2.1242,      1.54728,    1.46275,    2.08353,     -0.349797,  -0.767157,  -1.93123,
    -1.22863,   -1.33987,   -0.726219,  -0.997329,   0.28024,    0.103255,   0.623784,    0.626352,   1.12694,    1.49839,
    0.432037,   -0.0577674, -0.314986,  -0.496102,   -1.18837,   -2.0017,    -1.19159,    -0.947219,  -2.5949,    -2.38403,
    -0.784756,  -0.749927,  -1.85061,   -1.41954,    -0.588824,  -0.175078,  -2.03226,    0.401457,   0.570709,   0.173063,
    -0.800612,  -1.00578,   -0.413284,  -0.461997,   -1.54799,   -0.360349,  -0.172057,   -1.51289,   -0.163397,  -1.03356,
    -1.01099,   -0.839831,  -0.173977,  -1.05253,    0.857238,   0.406543,   -0.127678,   -2.50787,   -0.192743,  -0.830097,
    -1.30048,   -2.75772,   -0.463618,  -1.44991,    -2.48523,   -2.63131,   -2.19342,    -4.99053,   -5.03019,   -4.7758,
    -6.10028,   -7.49506,   -6.74455,   -7.53935,    -9.17242,   -8.2121,    -9.05083,    -8.47347,   -9.2221,    -8.30437,
    -6.72223,   -8.09677,   -7.84726,   -7.9011,     -8.04381,   -6.02154,   -6.48422,    -5.62021,   -6.37062,   -4.07518,
    -4.49868,   -2.02396,   -2.83132,   -3.44968,    -1.11981,   -0.618368,  -0.162212,   -0.101713,  1.39877,    0.47995,
    -0.792802,  1.83778,    0.695494,   -0.686329,   -0.872519,  -0.385498,  -1.25907,    -0.182559,  -0.355677,  -0.502172,
    -1.70178,   -1.52549,   0.730219,   -2.54871,    -3.29928,   -3.78205,   -4.11442,    -4.71521,   -5.06028,   -6.42411,
    -6.40948,   -4.71256,   -5.15601,   -5.56004,    -5.03484,   -4.49651,   -3.42939,    -4.64712,   -1.74578,   -2.10365,
    -2.38689,   -1.58235,   -1.55119,   -1.81705,    -1.77639,   -2.14831,   -2.71625,    -0.835904,  0.232315,   -1.57776,
    -1.074,     -1.18816,   -0.729887,  -0.283184,   -1.02982,   -0.655291,  1.92221,     1.76335,    1.89159,    2.32637,
    1.17562,    2.15871,    3.41404,    1.39403,     1.64602,    0.131289,   2.05947,     1.37178,    0.600073,   -0.818363,
    -0.15139,   -0.327754,  0.20629,    1.52005,     1.75385,    1.39786,    2.35723,     4.4407,     4.32663,    3.72477,
    3.7502,     6.0969,     6.32516,    4.5256,      3.88512,    6.20565,    4.49911,     3.35248,    4.61861,    2.14201,
    2.65148,    1.8262,     3.36833,    1.14889,     0.783012,   1.76356,    -0.413986,   0.0380319,  2.32746,    0.998746,
    -0.146673,  -1.18509,   -0.0854994, -0.836262,   -0.104841,  1.18441,    1.5793,      1.40416,    0.853623,   1.03881,
    2.50426,    2.36345,    3.89176,    2.36855,     2.05758,    1.97743,    2.51925,     1.35837,    -0.468224,  1.31126,
    -0.0536328, -1.10551,   -1.05995,   -0.757794,   -2.95614,   -4.44322,   -1.53319,    -1.41206,   -1.55843,   -1.46667,
    -0.503415,  2.1726,     1.72633,    1.6532,      3.36023,    3.76351,    2.8815,      2.40403,    4.87144,    3.69786,
    2.93861,    2.36309,    2.83423,    3.28151,     1.23417,    1.47765,    1.23267,     -1.34257,   -0.879411,  -2.42685,
    -2.63232,   -2.13055,   -2.0761,    -4.18267,    -4.1932,    -2.9697,    -4.30188,    -4.52833,   -3.05069,   -3.89795,
    -5.79304,   -4.67604,   -2.95545,   -2.64836,    -5.21766,   -3.99477,   -4.14274,    -4.37704,   -5.14836,   -4.87867,
    -2.55316,   -2.01118,   -1.09166,   0.169728,    0.306804,   0.203494,   2.74422,     1.2596,     2.82607,    3.95238,
    3.08856,    2.92532,    2.95216,    2.98351,     3.27574,    3.83624,    3.56491,     3.4243,     5.65413,    5.86373,
    5.71241,    4.74456,    5.48754,    5.51935,     5.6011,     5.07448,    5.2435,      4.90717,    4.9053,     5.25382,
    4.71262,    3.66772,    4.54057,    4.40167,     3.10862,    2.0922,     1.43759,     1.26013,    0.542764,   -0.615635,
    -0.20371,   -1.41603,   -0.640463,  -2.36183,    -1.07904,   0.108492,   -0.00483943, 0.118943,   0.340149,   2.01451,
    1.77352,    1.33798,    1.48513,    1.17519,     0.638168,   1.77321,    1.50569,     2.06973,    2.99459,    1.66726,
    2.1701,     3.32507,    2.07512,    2.6121,      1.93265,    1.45051,    0.694373,    0.105531,   -0.122746,  -0.707029,
    -1.2026,    -1.29989,   -1.36948,   -1.23336,    -2.61736,   -2.97324,   -2.75428,    -2.65801,   -2.60803,   -2.5874,
    -2.83738,   -1.19442,   -1.64937,   -0.267087,   -1.02556,   -1.51594,   -1.42184,    -1.37806,   -1.13483,   -2.51613,
    -2.12999,   -0.0887563, -1.34618,   0.502561,    1.04249,    -0.336828,  -0.737351,   0.237569,   -2.08945,   -2.40268,
    -1.95955,   -1.86755,   -1.5193,    -0.941992,   -0.359466,  -0.250101,  0.827395,    -1.03458,   0.253577,   0.321781,
    0.244794,   -1.01929,   -1.04375,   -0.762494,   -0.946706,  0.850249,   1.16029,     0.627367,   -0.0201622, 1.52268,
    2.52693,    2.21582,    1.07064,    2.74362,     1.65172,    2.23112,    0.760692,    -0.843152,  0.587993,   0.500434,
    -1.75993,   -0.894391,  0.275741,   -0.126272,   -0.723669,  0.00552378, -1.6073,     -1.53553,   -2.27241,   -2.31811,
    -2.6314,    -0.721988,  -2.38727,   -2.59316,    -1.88813,   -1.80762,   -3.48412,    -2.8516,    -2.80831,   -2.89959,
    -3.88683,   -3.72133,   -3.21136,   -5.69953,    -3.50825,   -3.78205,   -6.07771,    -4.8186,    -5.83896,   -4.5723,
    -6.10098,   -7.48248,   -8.5059,    -6.08352,    -6.85725,   -7.92316,   -5.72047,    -3.89037,   -5.06524,   -5.10203,
    -2.45226,   -1.96651,   -3.28933,   -1.31653,    -0.431499,  -0.864968,  1.23426,     2.67244,    2.41618,    2.4993,
    0.779123,   2.18458,    2.14045,    -0.00982022, 0.395157,   1.75845,    0.877253,    0.425317,   1.0865,     2.10208,
    1.89591,    1.50378,    1.25879,    2.72496,     2.89075,    2.44907,    3.95955,     2.93156,    1.86313,    2.3991,
    2.91305,    2.24837,    2.35581,    2.04967,     1.71856,    1.35094,    0.14658,     -1.62729,   -1.66687,   -1.95843,
    -2.87105,   -4.64242,   -4.17111,   -5.58713,    -3.57128,   -4.87012,   -4.44892,    -4.64971,   -3.96096,   -3.33089,
    -2.80228,   -2.76871,   -1.12805,   -0.601573,   -0.408903,  0.912948,   1.17209,     1.65941,    0.0940405,  0.606716,
    -0.270879,  0.37008,    1.1598,     0.993844,    -0.041602,  0.159974,   0.228533,    -0.300169,  1.96407,    0.146443,
    0.971167,   0.774351,   -0.450588,  -2.39818,    -2.24962,   -2.80438,   -3.69162,    -3.49721,   -3.04317,   -2.32207,
    -2.02777,   -2.50316,   -0.76183,   -0.446839,   0.528612,   1.66703,    1.28377,     0.721126,   1.02808,    2.37046,
    2.61081,    3.27347,    3.82783,    3.56662,     3.54194,    4.12914,    4.95509,     4.24875,    4.43105,    5.32149,
    5.19531,    5.04977,    6.15917,    6.79615,     6.91719,    7.88135,    8.9234,      9.36803,    8.92763,    8.46495,
    8.68096,    8.99316,    7.75767,    7.93296,     8.08309,    7.06845,    5.02183,     5.95177,    5.918,      4.08205,
    4.18502,    3.80814,    0.74113,    1.48761,     2.62065,    0.694486,   -1.30164,    -2.3588,    -0.302352,  -1.5671,
    -2.25472,   -1.55649,   -2.01963,   -2.46818,    -1.96716,   -3.10205,   -3.66526,    -3.7383,    -2.80941,   -1.92896,
    -1.03419,   -2.02455,   -2.52477,   -1.88958,    -2.32281,   -1.49101,   -2.04423,    -1.04067,   -0.390769,  -1.28354,
    1.14986,    0.944644,   2.03978,    2.69361,     3.58363,    5.05962,    4.96357,     5.31835,    5.59672,    6.02626,
    4.8763,     5.21567,    3.81637,    4.0286,      3.10567,    3.88898,    3.51899,     1.96984,    1.87145,    2.47895,
    2.60059,    0.623883,   -0.759829,  0.250432,    -0.308767,  -1.12839,   -0.514404,   -0.187402,  0.685532,   0.891604,
    0.447684,   0.851473,   1.30557,    0.715818,    1.36684,    1.77307,    1.42213,     2.68344,    4.34813,    4.79591,
    3.28098,    4.73084,    3.52001,    2.69288,     1.97565,    1.58072,    1.58801,     1.05489,    0.0934115,  0.503142,
    1.0207,     -0.86498,   0.00948239, 1.55816,     -0.140937,  -0.84314,   -0.356313,   -0.308752,  -0.10002,   0.887657,
    -2.24582,   -1.40254,   -0.522946,  -0.437792,   -1.83832,   -2.33374,   -1.23715,    -2.77073,   -2.49501,   -0.639706,
    -1.9953,    -1.76139,   -2.10304,   -1.01736,    -2.25265,   -3.83242,   -3.84946,    -2.83332,   -2.91157,   -1.86383,
    -1.98457,   -0.972304,  -0.596844,  0.023088,    -1.90602,   0.940504,   2.22432,     -0.087193,  1.31193,    3.24166,
    3.31753,    4.54197,    2.82701,    4.5772,      1.71757,    3.24668,    3.54623,     3.26764,    3.18704,    4.60377,
    3.30578,    4.02865,    3.15578,    3.56049,     3.11847,    3.58835,    3.51343,     3.89529,    2.19765,    3.0212,
    2.95397,    1.77176,    2.08433,    0.632517,    -1.38789,   -0.0466202, -0.66302,    -1.94383,   -2.43592,   -1.97494,
    -2.9667,    -3.66948,   -3.99846,   -3.15893,    -3.89861,   -4.25904,   -3.35715,    -3.07267,   -2.90583,   -3.29303,
    -2.57061,   -2.28862,   -3.05817,   -3.35807,    -2.97317,   -2.63703,   -2.20714,    -2.52367,   -2.67913,   -2.87945,
    -3.53052,   -2.67748,   -4.7277,    -4.60305,    -4.23769,   -4.41664,   -3.09841,    -3.58655,   -1.99997,   -2.29179,
    -2.37228,   -1.95746,   -1.10968,   -2.03033,    -1.12787,   -0.722337,  -0.871017,   -0.175467,  -0.666555,  -0.142878,
    0.715347,   1.73103,    0.891649,   0.269296,    1.34527,    0.17199,    0.70655,     0.158569,   0.15807,    -0.268695,
    -0.941435,  -1.53553,   -1.35602,   -2.48284,    -1.84137,   -2.10203,   -1.27396,    -2.15113,   -1.93538,   -1.86993,
    -0.465532,  -1.76653,   -1.16782,   -1.09781,    -0.369869,  0.307955,   0.916135,    2.06731,    1.16544,    2.04148,
    3.65309,    3.03764,    3.12468,    5.75846,     5.41254,    6.17844,    7.35262,     6.41507,    6.98319,    8.00242,
    7.84885,    7.46472,    6.29113,    8.74066,     6.01542,    5.05445,    6.20665,     5.07805,    4.46484,    3.85593,
    2.13265,    2.91793,    2.30644,    1.58126,     2.12065,    1.74069,    1.53482,     1.55467,    1.2146,     1.16406,
    1.85242,    1.27934,    -0.133984,  -0.10858,    -0.0287076, -0.175439,  -1.07144,    -1.35367,   -1.76136,   0.399808,
    -0.0401004, -1.98723,   0.465618,   0.386514,    -0.792003,  -0.040655,  0.435747,    1.87909,    1.99424,    0.291803,
    1.05795,    0.928135,   1.32179,    0.34205,     -0.698922,  0.835252,   0.651524,    0.478915,   1.78065,    0.0166886,
    -0.98963,   2.08124,    -1.08279,   1.6028,      -0.333253,  -1.24324,   -0.327525,   0.0709702,  1.02527,    0.933132,
    0.198952,   0.27715,    0.494888,   0.122397,    -0.855442,  -0.94372,   -1.16785,    -1.29985,   -0.21662,   -2.1779,
    -2.87792,   -3.93545,   -3.37824,   -2.3343,     -4.0299,    -3.71468,   -2.69175,    -4.44614,   -3.75257,   -1.90843,
    -3.61893,   -2.03373,   -1.96216,   -2.97554,    -3.49618,   -3.80133,   -2.21707,    -1.40791,   -2.61216,   -1.95516,
    -2.88288,   -0.458705,  -0.410249,  -0.701522,   -1.48361,   -0.925277,  -1.37351,    -0.544615,  0.357697,   -0.337192,
    0.956326,   -0.719959,  0.554526,   2.6672,      1.28572,    1.68263,    1.63036,     1.23999,    0.97693,    1.38059,
    0.860477,   0.362572,   2.49184,    3.02845,     2.6842,     3.09969,    1.82906,     3.28209,    3.14493,    2.19544,
    2.89953,    2.12132,    2.39033,    2.29989,     2.45484,    3.95956,    3.20252,     3.05118,    3.64348,
};
const test_input_test3 = [_]f64{
    0.350394,    -0.994945,  0.0830646,  0.0243834,  -0.276853,  0.338712,   1.29058,    2.40158,     2.47705,    2.7253,
    1.9412,      3.17409,    1.46332,    0.351798,   -0.95905,   -0.972537,  -0.264139,  0.333564,    -0.232584,  -1.85144,
    -1.68603,    -1.47082,   0.503685,   0.941981,   0.709138,   0.328586,   0.14561,    -0.89354,    0.165513,   1.89499,
    1.46075,     0.795057,   0.463917,   1.46021,    0.756041,   1.1322,     0.0854401,  0.742616,    0.895651,   0.688415,
    -0.29735,    -0.931195,  0.0802377,  -0.353879,  0.0596443,  2.6116,     2.17562,    1.58392,     0.229735,   0.0524373,
    -0.6638,     -1.31992,   -0.270991,  -0.508754,  -0.785489,  0.887231,   0.175184,   -1.22644,    -0.538809,  0.223763,
    1.57789,     0.119277,   -1.08899,   -1.77386,   -0.751811,  0.334621,   -0.0197696, 0.68539,     0.970724,   1.99246,
    1.50552,     -1.61879,   -0.912679,  1.3745,     0.615091,   1.73953,    1.47907,    -0.118823,   -1.18657,   0.00935715,
    -0.176975,   -0.977452,  0.676699,   -0.954164,  0.0117292,  -0.169237,  -1.51202,   -2.92928,    -1.94937,   -0.830859,
    1.07357,     1.21583,    0.175918,   0.00717284, 0.579716,   1.23021,    -0.997656,  -2.67963,    -2.39341,   -1.84935,
    -0.210849,   0.352916,   -0.620031,  0.283497,   3.13396,    0.775568,   -2.38201,   -2.32055,    -1.17635,   1.4157,
    2.18008,     0.743249,   1.43521,    2.08238,    -0.336817,  -1.53841,   -0.508269,  0.689133,    0.725952,   1.27647,
    1.47785,     1.44368,    0.0855124,  1.30954,    1.54069,    1.27615,    0.0475867,  -0.261098,   -0.628487,  0.12761,
    0.709917,    -1.32665,   -2.86025,   -3.20077,   -1.00524,   0.300356,   1.11709,    1.65961,     1.42697,    0.133947,
    -1.49438,    -3.62868,   -2.05359,   -2.86484,   -1.28913,   1.44799,    3.80552,    2.98314,     1.65235,    0.147991,
    -1.58968,    -2.28718,   -0.454357,  -1.32221,   -3.20661,   -2.48267,   1.0704,     2.51867,     1.00305,    0.118798,
    -1.06368,    0.25387,    0.585367,   0.0510902,  -0.241737,  -0.478254,  -0.798845,  -0.306564,   0.182102,   -1.11655,
    -1.90283,    0.106595,   -0.755149,  -1.53277,   1.30156,    2.09875,    1.62922,    0.181934,    0.218448,   -0.0945462,
    -0.510536,   -0.483002,  0.799698,   2.71068,    1.61035,    0.306949,   -1.16859,   0.142485,    -1.47791,   -1.04836,
    1.75855,     3.71051,    3.96235,    0.955907,   -3.36659,   -6.45396,   -4.91175,   -1.53268,    1.59625,    0.275631,
    -0.702491,   0.0457979,  1.03686,    -0.0595258, -1.01349,   -1.29486,   -0.891654,  1.53058,     1.83114,    1.09994,
    -0.118083,   -0.990172,  -2.16804,   -0.521333,  0.782263,   1.02952,    1.47002,    0.904681,    -0.888374,  -3.38803,
    -2.57893,    -0.311696,  0.756091,   0.0839908,  -0.418234,  -1.05725,   0.865292,   1.53433,     0.0983097,  -1.9712,
    -1.60268,    -2.4828,    -0.397514,  1.97381,    2.57765,    0.685065,   0.0403296,  -0.00381523, 0.359175,   0.419552,
    0.652383,    -0.115027,  0.64872,    -0.197061,  -1.91513,   -1.49565,   -0.241031,  0.65625,     1.45756,    -0.962275,
    -2.0555,     -0.979419,  -0.886128,  -0.55976,   0.429835,   0.748921,   3.81132,    3.2622,      0.430525,   -1.69639,
    -1.44985,    -0.829957,  -0.213325,  -1.48572,   -2.79371,   0.402135,   -0.437192,  -1.76234,    -1.80716,   -0.730218,
    -0.696048,   0.44687,    2.5505,     2.89594,    0.0695752,  -2.29347,   -2.01177,   -1.13921,    0.373649,   3.03244,
    3.19671,     2.77501,    -0.102918,  -1.91628,   -2.59037,   -1.05973,   1.3565,     1.78239,     1.6367,     1.39002,
    3.23711,     2.28674,    0.71814,    -0.683011,  -0.454855,  -0.787193,  0.659416,   1.10634,     1.69555,    0.0489239,
    0.00351194,  -1.01139,   -2.58154,   -2.12811,   -1.65158,   -0.689077,  0.239711,   1.53822,     1.97632,    1.57505,
    0.125305,    -0.736488,  0.290505,   1.00934,    -0.350446,  -2.33503,   -1.29783,   -0.953738,   0.757308,   2.49916,
    2.15559,     0.929058,   0.841539,   0.967926,   0.96271,    -0.296048,  -0.177014,  0.0357363,   0.212491,   -0.502381,
    -1.18462,    -0.439052,  0.850869,   3.41399,    2.31638,    0.786068,   -0.416693,  -2.95245,    -2.87259,   -0.910645,
    0.0949596,   2.05793,    2.23376,    0.789032,   0.984995,   0.238795,   -1.77356,   -0.51297,    1.19926,    1.32393,
    1.19871,     -0.851978,  -2.76332,   -1.42654,   0.703565,   0.704314,   -0.245266,  -1.09573,    -2.04125,   -0.117155,
    1.01944,     -0.473266,  -1.41485,   -1.96922,   -1.09176,   1.02566,    2.97805,    1.32206,     -0.504806,  -2.68472,
    -2.6387,     -1.48914,   1.49939,    0.963183,   -0.486002,  -1.58647,   -2.44592,   -2.06598,    0.413799,   2.61114,
    1.56862,     -0.671295,  -1.4464,    -1.48052,   -1.04347,   1.21231,    1.97464,    0.860268,    -0.724972,  -1.39934,
    0.258313,    1.54678,    0.926663,   -1.37717,   -0.108739,  1.22335,    2.10713,    0.969644,    0.498324,   -0.348261,
    -1.29209,    -0.396477,  1.78748,    -0.139855,  -0.639945,  -0.436158,  -1.08697,   -0.829314,   1.09032,    1.83857,
    3.23051,     2.40743,    -0.146377,  -1.76209,   -2.16412,   -0.967389,  -0.449612,  0.898793,    0.225034,   -1.70705,
    -1.5463,     -0.230929,  0.774714,   -0.161805,  0.701703,   1.53575,    2.50323,    2.44486,     0.178724,   -1.41996,
    -1.85519,    -0.886616,  -0.640971,  -0.285058,  -0.58321,   0.277486,   -0.024647,  -2.4506,     -3.5649,    -2.60292,
    -0.525108,   0.211518,   1.43811,    2.03871,    -1.59283,   -2.4924,    -2.5469,    -0.876554,   -1.95057,   -0.870956,
    -0.725461,   0.161009,   -0.299884,  1.44007,    2.53452,    1.07046,    -0.168643,  0.366231,    0.471405,   1.16026,
    1.18031,     0.370433,   -0.371246,  -0.878241,  -2.2225,    -2.4012,    -1.8655,    0.139126,    1.70443,    1.27044,
    2.86755,     2.26922,    0.641286,   0.511121,   2.1079,     1.66005,    -0.193291,  -0.645104,   -1.00237,   -0.240258,
    1.12901,     3.74733,    3.11519,    1.27917,    -0.94544,   -2.088,     -1.52059,   1.8725,      2.53985,    3.60936,
    2.15138,     -0.873494,  -2.96124,   -0.80804,   -1.3485,    -1.0876,    0.0896258,  -0.510814,   0.971944,   0.963723,
    1.47508,     1.75542,    1.25003,    1.27496,    -0.12987,   1.18948,    1.56222,    -0.229649,   -1.01995,   -0.775149,
    0.0295828,   -1.03336,   -0.135574,  0.513219,   1.3027,     1.31747,    0.701467,   -0.239215,   -1.29851,   -0.689983,
    0.965063,    0.297986,   0.869065,   1.56821,    0.713878,   0.149979,   0.419237,   -0.348942,   -1.70841,   -1.6934,
    0.223811,    2.50628,    3.25829,    1.94515,    0.652415,   -0.677049,  0.98269,    1.64844,     -1.28723,   -2.77558,
    -2.28501,    -1.37988,   0.303028,   1.22028,    2.25543,    1.53494,    1.16003,    1.41307,     0.760364,   0.381282,
    0.266839,    -0.780353,  -1.04988,   -1.05736,   -1.58408,   -0.512953,  0.298307,   0.75725,     -1.51771,   -1.22512,
    0.936476,    2.01442,    0.151472,   -1.27355,   -1.53626,   -2.3014,    -1.30096,   -0.121766,   0.289569,   0.2951,
    -1.23063,    -1.51436,   -0.143298,  -0.481271,  0.742558,   1.27975,    1.64536,    1.60432,     1.97974,    -0.194577,
    0.165702,    -0.242437,  0.267815,   -0.561484,  -0.833247,  0.508914,   0.0170574,  -1.01972,    -0.701191,  -0.265911,
    0.614303,    1.02981,    -0.0153402, -1.05396,   -0.463525,  -0.0771127, -0.266761,  0.755106,    -0.606005,  -0.668653,
    -1.88461,    1.13069,    1.64629,    0.85677,    -3.16229,   -2.16467,   -0.575665,  0.853431,    0.233166,   1.29066,
    0.017058,    -0.682934,  -0.868807,  0.527979,   0.0847427,  -1.30269,   -2.25589,   -2.72647,    -0.15471,   0.548402,
    0.901107,    -0.787187,  -1.06975,   -1.75194,   -1.6215,    0.899486,   0.399767,   0.638693,    0.820902,   -0.0659783,
    -0.612889,   -1.75969,   -3.26585,   -0.39885,   2.44311,    1.85808,    2.4733,     2.42359,     2.14708,    1.88832,
    -0.881713,   -2.33676,   -1.30057,   -0.0832507, 0.0103129,  0.141132,   -0.357172,  -2.52832,    -1.83888,   1.63641,
    3.81838,     5.4401,     3.70728,    2.26087,    -0.525245,  -0.315589,  0.0713773,  -1.29461,    1.15205,    1.38099,
    1.60961,     1.35342,    2.55396,    2.61346,    1.92579,    1.15556,    0.695707,   -0.611525,   0.860986,   1.75433,
    1.22977,     0.872744,   -0.717571,  -0.116683,  -0.816318,  -2.52219,   -0.339713,  2.88508,     3.32979,    1.44294,
    -1.26033,    -2.11941,   -0.651147,  -0.556536,  0.0925199,  0.877394,   2.07304,    0.387788,    -2.3965,    -3.63704,
    -2.46795,    -0.492742,  0.83505,    0.154256,   0.11443,    0.603752,   -0.200106,  -0.621551,   -1.36911,   0.627258,
    1.69433,     1.04396,    0.33159,    1.11949,    0.252581,   0.719045,   1.45538,    0.863798,    0.297695,   -1.07994,
    -0.308294,   -2.04022,   -1.1567,    -0.385271,  -0.113161,  0.601259,   1.60698,    0.678505,    -1.70961,   0.181124,
    1.64101,     0.988023,   -0.920889,  -0.651477,  0.965302,   0.93574,    -0.0506845, -1.29392,    -0.852926,  0.176206,
    -1.78382,    -2.97127,   -2.78086,   -0.826899,  0.13968,    0.633139,   0.8858,     -0.031284,   -2.74759,   -3.21973,
    -3.14759,    -1.26779,   1.32115,    2.35068,    2.25769,    1.16543,    1.70102,    -0.227704,   -1.69536,   -1.06955,
    0.541685,    0.521986,   1.85456,    3.18642,    2.02108,    2.32803,    2.48636,    0.30342,     -1.41168,   -3.96927,
    -2.29673,    -1.11283,   -0.473674,  -0.0263915, -0.0203612, -0.454894,  -0.535035,  0.26708,     -0.123484,  -1.66052,
    -2.04597,    0.383889,   1.78814,    -0.028753,  -0.562072,  0.277397,   0.996438,   0.265221,    -1.20651,   -2.47102,
    -2.80444,    -3.26579,   -2.28519,   -0.347454,  0.718006,   0.0331211,  0.537687,   0.23318,     -0.395945,  -0.359903,
    0.201522,    0.900449,   0.417311,   -0.418626,  0.428208,   -1.55375,   -1.65635,   -1.74367,    0.945494,   3.61454,
    4.39519,     2.53118,    1.27597,    -0.503489,  -0.0731608, -0.46304,   -0.303052,  1.57432,     2.28335,    0.319761,
    -2.54583,    -2.55886,   -1.93898,   0.837351,   1.45073,    1.47624,    -0.0385718, 1.06825,     1.06229,    0.607738,
    0.496884,    -0.690183,  -2.95153,   -3.43803,   -0.674535,  1.02431,    2.51353,    2.88059,     1.00163,    -0.331386,
    -0.967325,   -0.675148,  -0.825043,  -1.65261,   -1.58269,   -0.484533,  0.260661,   0.270199,    0.382151,   0.875523,
    1.72353,     0.658275,   -0.716472,  -0.395446,  -0.806623,  -0.321939,  1.12801,    2.30621,     2.24066,    0.573076,
    0.153987,    0.393299,   0.69623,    2.21051,    2.93321,    2.23408,    0.0277941,  -1.1591,     0.807246,   -0.751037,
    -3.6941,     -2.86332,   -1.14196,   -0.168101,  0.763001,   0.634989,   0.0184467,  -0.863873,   -2.3591,    -1.43378,
    1.62969,     0.517144,   -0.0918909, 0.178299,   0.0417185,  -1.518,     -1.10262,   1.42844,     1.85117,    1.29285,
    -0.338269,   -1.26251,   -2.92452,   -4.76441,   -4.53164,   -0.479243,  3.64693,    3.63882,     2.73094,    2.47235,
    0.218523,    -1.0341,    -1.04555,   0.599369,   0.14861,    -0.0497009, -0.108285,  0.949331,    1.38095,    1.24929,
    2.04599,     1.6581,     1.81736,    1.95071,    1.95745,    0.886377,   1.01734,    0.251949,    -0.0692864, -0.767318,
    -2.42824,    -1.78177,   0.792931,   2.34607,    1.92224,    0.349315,   -1.05787,   -0.415069,   0.258202,   3.74402,
    3.95821,     1.86308,    1.15033,    0.404044,   -2.03756,   -2.13741,   -0.699396,  0.939633,    0.798321,   1.05505,
    0.0175672,   0.117639,   1.16502,    -0.51223,   -0.764164,  0.28745,    -0.762807,  -0.0167352,  1.19635,    0.601538,
    -1.06842,    -0.97237,   -0.884902,  -1.18606,   -0.336458,  0.891664,   1.36417,    0.612332,    -0.0232285, -1.57927,
    -2.85816,    -2.94748,   -1.9198,    -1.10158,   -1.70931,   -0.285317,  1.2791,     -0.369614,   -0.22532,   0.556208,
    1.39956,     -0.14921,   -0.88712,   -0.430832,  -0.469464,  0.489405,   0.0114141,  -0.360822,   -1.16864,   -0.501263,
    -0.442682,   -0.136456,  0.913832,   1.544,      0.530454,   -0.427743,  0.27476,    2.01259,     1.80179,    -0.179191,
    -0.894698,   1.08999,    2.13967,    0.22948,    -2.16396,   -3.24463,   -2.30551,   -1.35839,    0.604641,   -0.49962,
    -0.338188,   -0.2261,    -0.474152,  2.40971,    3.2689,     1.17169,    0.810745,   0.368212,    0.277324,   -0.713314,
    -3.29998,    -2.1363,    -1.69666,   -0.0780577, 1.50062,    1.66284,    1.94125,    0.99076,     1.08213,    1.90861,
    -0.00745845, -2.04717,   -1.54353,   0.616858,   1.85244,    1.69321,    -1.65637,   -4.38997,    -3.70481,   -2.85322,
    -1.10362,    1.53888,    1.3282,     1.20235,    0.456759,   -2.72878,   -2.8458,    -1.50762,    2.16464,    3.58837,
    2.71405,     -0.0235837, -2.38268,   -0.224704,  2.28242,    3.15696,    3.09183,    1.59718,     0.253724,   -0.366983,
    0.14117,     0.271041,   -0.512044,  -2.7243,    -2.71373,   -1.00258,   0.419721,   1.05451,     0.255145,   0.0963533,
    -0.0915133,  0.0706681,  -0.241783,  -0.481656,  1.23714,    2.27465,    0.49698,    -0.892252,   -0.734435,  1.58514,
    3.62273,     2.77677,    1.08206,    -1.66719,   -2.71641,   -0.880798,  2.4568,     2.55912,     2.58238,    0.540715,
    -0.525318,   -0.201316,  1.27144,    3.59509,    3.75435,    1.99014,    1.17178,    -0.0745491,  1.42006,    2.64843,
    2.18761,     2.2601,     3.03284,    1.63285,    -0.836224,  -4.97826,   -4.69087,   -4.06512,    -1.93297,   -0.569392,
    1.64817,     2.20976,    1.44052,    -0.441525,  -1.1077,    -0.9067,    0.808907,   1.32859,     -0.391506,  -0.22805,
    -1.65057,    -1.94783,   -0.652961,  0.686181,   0.738194,   -0.373017,  0.423315,   0.705433,    3.01396,    1.95594,
    0.154556,    -1.73165,   -2.1316,    0.308874,   1.65308,    0.856851,   2.26479,    1.50252,     -0.242407,  -0.631625,
    1.07142,     0.784683,   1.40818,    1.36783,    -0.349315,  0.24055,    1.43307,    1.43558,     1.1911,     1.42473,
    1.62812,     2.20315,    1.72325,    -0.160773,  -1.66633,   -1.20979,   0.875979,   1.90407,     3.52365,    0.927726,
    0.0660131,   -1.12077,   -1.31021,   -2.2012,    -0.871637,  0.803285,   1.60678,    0.473275,    1.18901,    0.756831,
    0.24208,     -0.8407,    -0.868867,  -1.99745,   -2.06837,   -0.60693,   1.17388,    -0.602159,   -0.933359,  -2.59997,
    -2.98605,    -3.35109,   -2.41737,   -1.17402,   -1.2597,    -2.07937,   -3.7636,    -4.11414,    -2.36531,   0.351199,
    2.17545,     0.788777,   -0.526223,  -0.310068,  -0.822152,  -0.0664202, 0.596,      1.39208,     1.64761,    1.10912,
    -0.239386,   -2.09029,   -2.047,     -0.951495,  -0.142458,  0.530582,   0.931158,   1.0445,      1.53815,    1.72214,
    1.12397,     0.742954,   0.0784585,  -2.35663,   -1.71938,   -2.29894,   -3.39423,   -2.60009,    -1.69133,   -0.247403,
    1.88947,     3.22451,    0.581142,   0.206573,   -0.547193,  1.24006,    0.545286,   0.319965,    0.81048,    0.879248,
    1.37932,     0.0377395,  -0.305449,  -0.365883,  1.34316,    2.07674,    1.22927,    -0.973028,   -1.57902,   -2.43877,
    -2.78902,    -3.10076,   -0.911808,  1.99745,    3.58996,    0.96406,    0.187949,   -0.334081,   -1.048,     -2.51116,
    -2.55287,    -1.56718,   -2.02848,   -0.191623,  0.750667,   0.809453,   -1.29643,   -0.00240562, 0.931175,   1.71883,
    2.22716,     0.670824,   0.425751,   0.713757,   0.649782,   -1.23618,   -1.96555,   -4.10769,    -3.77012,   0.347588,
    3.90971,     3.22432,    -1.06498,   -1.72528,   -0.636288,  -2.35821,   -1.61682,   0.0575011,   -0.121344,  -1.80324,
    -2.93419,    -3.66878,   -3.36376,   -2.05442,   -0.338332,  0.630213,   0.868143,   -0.496026,   -1.72688,   -0.445335,
    -1.47566,    -1.86117,   -0.397891,  -0.572852,  -0.146414,  0.787799,   -0.305385,  -1.43836,    -0.586497,  -1.35617,
    0.0515099,   -0.063515,  1.19774,    2.95177,    4.67966,    2.67998,    1.13762,    0.421357,    -0.285617,  -0.433429,
    -1.10943,    -1.71429,   -2.0094,    -0.613174,  0.964303,   1.20264,    0.561777,   -2.46652,    -4.87941,   -4.74653,
    -3.0576,     2.40022,    5.21541,    5.30611,    3.61742,    0.402864,   -1.15385,   -3.12256,    -2.79765,   -0.873934,
    0.600155,    1.31709,    -0.432879,  -2.16618,   -0.745124,  0.675823,   0.968999,   -0.415967,   -0.641048,  -0.818807,
    -1.59465,    -1.34153,   -1.11182,   -2.23907,   -2.77084,   -1.57883,   -0.198134,  1.59829,     2.66188,    1.56013,
    0.0246934,   1.10986,    -0.508234,  0.142804,   0.340175,   -1.3531,    -1.69196,   -1.36129,    -0.369551,  -0.290009,
    0.0816389,   0.728858,   1.2617,     0.0969199,  -0.984639,  -1.69891,   -0.0797756, 0.409987,    -0.0492129, 0.0817645,
    0.658307,    0.779798,   0.48878,    1.05822,    -0.287185,  -0.419674,  0.783793,   -0.101397,   -0.250113,  -0.105448,
    -0.983997,   -2.50489,   -1.97752,   -1.33597,   0.660614,   0.631164,   -1.17512,   -0.551888,   -1.63686,   -1.92594,
    -1.11961,    0.886976,   1.61059,    1.47957,    0.674509,   0.79399,    -0.224661,  -0.826678,   -1.21519,   -1.23265,
    -0.38264,    0.743695,   -1.7133,    -0.589349,  0.367045,   1.72872,    2.77358,    2.47084,     0.238369,   -2.5734,
    -3.46608,    -1.4142,    0.857312,   0.459876,   2.06929,    1.90066,    0.522686,   0.818562,    0.224351,   -1.84505,
    0.151831,    0.845365,   0.883929,   -1.82026,   -2.78075,   -1.57976,   1.13538,    1.80362,     1.36293,    0.956672,
    -0.834427,   -1.01039,   0.442665,   1.96,       2.00756,    0.861134,   -0.681689,  -3.63195,    -4.70501,   -3.39243,
    -4.00423,    -3.66301,   -2.37643,   0.399091,   1.3345,     -0.223762,  -1.07472,   -1.16386,    -0.558001,  0.800879,
    0.683407,    0.00802989, -1.80119,   -1.92539,   1.38514,    2.41631,    0.808309,   -0.942814,   -1.29795,   -0.2456,
    0.472615,    -1.03069,   0.0877001,  0.433334,   0.586251,   2.85619,    2.05073,    0.487148,    -0.75031,   -1.67166,
    -1.54711,    -1.13952,   -0.744194,  -0.703277,  -1.81665,   -0.743744,  -1.67172,   -1.40669,    0.599991,   2.97614,
    1.91765,     1.849,      0.908509,   -2.49106,   -4.01539,   -3.05029,   -1.64632,   0.0412988,   1.14809,    2.53736,
    1.21601,     0.320714,   -1.70135,   -1.85999,   -1.27764,   -2.34433,   -0.801495,  1.64419,     4.20665,    2.51813,
    0.267778,    -0.660118,  -0.353292,  -0.51965,   -0.215395,  -0.125234,  0.819757,   -1.18638,    -1.50958,   -2.12984,
    -1.26863,    0.0128985,  1.23683,    0.243243,   1.25699,    1.45689,    0.634342,   -0.478249,   0.628941,   1.82411,
    -0.39081,    -1.49805,   -2.80539,   -0.601223,  1.12979,    1.85722,    0.686521,   -0.124908,   0.0292058,  -0.0879855,
    0.0477534,   1.69009,    1.97969,    1.47389,    -0.256925,  -0.354442,  -0.277271,  0.468019,    0.111144,   0.95725,
    0.637306,    -2.10955,   -3.68658,   -3.61921,   -1.25491,   -1.10973,   0.265964,   1.23642,     0.712907,   0.41885,
    -0.0321915,  -0.15304,   0.359397,   0.0018072,  0.00571137, -0.0651676, -0.231226,  -2.21134,    -2.32561,   -1.2921,
    -0.0128674,  1.10632,    0.798583,   0.424296,   1.10468,    0.655798,   0.0900038,  1.08755,     -0.347324,  -3.01641,
    -5.05757,    -4.3069,    -1.132,     -0.435724,  0.69001,    0.571405,   -0.285412,  -0.412467,   -0.950448,  0.271736,
    1.50071,     0.640354,   0.880327,   -0.660447,  -0.377172,  0.447573,   0.405253,   1.37222,     2.70432,    1.88083,
    0.0153696,   -1.38772,   -0.477076,  1.16158,    0.359148,   -1.40635,   -2.50118,   -2.87859,    1.05221,    1.17551,
    0.620353,    0.0282931,  1.2747,     0.136337,   -1.07642,   -1.25198,   -0.656905,  -1.02827,    -0.421461,  -2.00213,
    -2.87019,    -2.52586,   -0.742898,  -0.721218,  -1.29751,   -2.62071,   -4.16561,   -3.74605,    -1.52889,   -0.580517,
    1.33498,     0.70311,    1.0688,     2.05647,    1.06507,    0.922993,   -1.42647,   -2.01458,    -0.730222,  0.360031,
    1.46789,     2.20883,    1.54596,    1.29493,    -2.18299,   -1.71233,   -0.545828,  -0.199916,   -1.30566,   -0.503581,
    -1.04548,    -0.570325,  -1.79655,   -1.10697,   -0.26767,   -1.02998,   -0.274751,  0.297703,    0.596199,   1.52761,
    1.43282,     -0.119734,  -0.698072,  -1.07443,   -1.30248,   -2.85199,   -2.83738,   -1.36555,    -0.253211,  -0.592401,
    -0.254604,   0.870241,   0.529716,   0.557594,   -0.492974,  -0.61407,   1.81175,    1.66165,     0.476889,   1.45033,
    0.906701,    0.546041,   -2.69148,   -5.06727,   -2.73825,   2.58644,    4.1646,     2.30969,     0.15349,    -1.70049,
    -1.35762,    0.259314,   1.66349,    1.58203,    0.826297,   -1.04694,   1.11674,    2.64735,     3.52354,    1.96944,
    -0.829926,   -0.237787,  0.0989646,  1.09625,    1.43413,    1.42581,    1.62174,    -0.769273,   0.795298,   1.70205,
    1.46088,     0.864468,   1.64119,    1.63032,    0.289593,   0.705699,   0.11528,    -0.225244,   -0.193662,  -0.874684,
    -0.808812,   0.847286,   1.68612,    1.58822,    0.711058,   -2.19268,   -2.54332,   -1.19311,    -1.24113,   -0.407527,
    1.61744,     2.57789,    2.87281,    1.17751,    -0.0199512, -1.72454,   -2.64544,   -0.865285,   0.972146,   1.59338,
    2.25124,     0.892856,   0.555755,   -0.709081,  -1.97664,   -3.07171,   -0.240414,  -0.851707,   -1.20559,   -1.80447,
    -1.77751,    -1.24209,   0.396708,   1.98461,    1.91623,    0.788943,   0.44833,    1.83281,     1.43467,    -0.38179,
    -0.715888,   0.933653,   1.10297,    2.15212,    0.55351,    -0.10016,   0.758645,   0.377576,    -0.39391,   -0.17329,
    -1.67712,    0.275202,   3.06215,    5.25967,    2.94903,    -0.248445,  -2.08941,   -3.20115,    -2.87377,   -1.71701,
    -2.35983,    -1.5481,    0.763542,   1.97175,    1.86532,    -0.178708,  -0.904734,  -2.23692,    -1.67292,   0.0414194,
    1.78036,     1.85852,    0.972545,   1.2227,     0.901226,   -0.0422479, -1.53866,   -2.96681,    -0.50756,   2.1498,
    2.29655,     2.21426,    2.5972,     1.33258,    -0.69775,   -1.70654,   -1.82389,   0.160923,    3.8918,     3.92515,
    1.83527,     -0.623351,  -2.77112,   -3.37197,   -0.94031,   0.261254,   0.398152,   -0.00176634, -1.20792,   0.0574421,
    1.93945,     2.40246,    0.543309,   0.403454,   -1.23303,   -0.238959,  0.749974,   1.52034,     1.54963,    1.89634,
    -0.844431,   -0.0656159, 0.0522065,  0.667727,   1.15392,    1.53596,    1.27626,    1.39794,     1.66912,    -0.0490651,
    -0.570802,   -0.0174721, 0.285613,   0.424505,   1.92272,    2.83157,    1.64462,    -0.812433,   -0.862819,  -1.98419,
    -3.33761,    -4.73927,   -4.53636,   -1.17715,   3.13808,    4.1427,     2.20532,    1.33263,     -1.4101,    -1.98934,
    -0.650215,   1.41336,    0.894967,   1.14182,    -0.0148319, -0.070722,  0.387088,   3.04793,     3.15509,    2.6637,
    2.63899,     1.01676,    2.24454,    2.33202,    3.79954,    2.84919,    -0.271568,  -1.49824,    0.324245,   0.653432,
    0.976307,    -0.933812,  -2.03912,   -2.38387,   -3.02954,   -3.73678,   -5.64166,   -2.94081,    -0.892362,  2.31501,
    4.68225,     3.02522,    1.38029,    0.928889,   2.65243,    3.32423,    1.98558,    1.04931,     -0.0249684, -0.545424,
    -0.274306,   0.935793,   1.28893,    -1.15287,   -0.91913,   -1.66251,   -1.73483,   0.150566,    0.75552,    0.737795,
    0.77091,     0.625633,   0.308936,   0.822506,   1.65079,    1.37363,    1.91799,    1.49711,     2.13865,    1.76601,
    1.25938,     -0.314031,  -0.311162,  -0.0508888, 0.338047,   0.783406,   1.70656,    1.8982,      -0.0844638, -2.08863,
    -2.55223,    0.440977,   1.58503,    1.54108,    0.561238,   -0.239909,  -1.43723,   -1.3308,     0.294347,   0.65602,
    0.292816,    0.385973,   -1.12941,   0.229344,   1.01003,    1.71846,    1.59786,    -0.505482,   -0.827442,  -1.78924,
    -1.69769,    -2.12668,   -0.370628,  -0.0875976, 0.0451396,  -3.67245,   -3.70737,   -2.53714,    -3.38953,   -0.708919,
    2.66921,     3.30998,    2.42922,    1.78827,    0.71375,    -1.48955,   -3.18567,   -2.5368,     -0.750597,  0.869904,
    2.08549,     -0.0559989, -3.14748,   -3.30494,   -2.9489,    -1.18031,   0.953482,   0.141021,    0.161389,   1.65526,
    3.65408,     3.4205,     1.50415,    0.132104,   -0.812322,  -0.0903128, 1.57541,    2.27973,     1.18118,    -0.19181,
    -2.13683,    -1.261,     -1.10156,   -1.81544,   -1.39755,   -0.1052,    2.01694,    3.36947,     2.6659,     1.95996,
    -0.106662,   -1.94171,   -1.17879,   -1.06241,   -0.546038,  -0.13255,   -0.731366,  -1.46033,    -0.769755,  -0.664746,
    -0.888176,   -0.305558,  -0.576032,  -0.263171,  -0.297043,  2.25192,    2.19337,    1.66279,     0.615048,   -0.213033,
    -1.38304,    -2.45712,   -0.744146,  -1.25376,   -1.011,     -0.278214,  0.886382,   1.33709,     2.46804,
};

fn roundDec(v: f64, dec: u32) f64 {
    const p = math.pow(f64, 10.0, @floatFromInt(dec));
    return @round(v * p) / p;
}

test "Burg coefficients against MBST" {
    // Skip if test data arrays are empty (not yet populated).
    if (test_input_four_sinusoids.len == 0) return;

    const CoefCase = struct {
        name: []const u8,
        input: []const f64,
        degree: i32,
        dec: u32,
        want: []const f64,
    };

    const cases = [_]CoefCase{
        .{ .name = "sinusoids/1", .input = &test_input_four_sinusoids, .degree = 1, .dec = 1, .want = &[_]f64{0.941872} },
        .{ .name = "sinusoids/2", .input = &test_input_four_sinusoids, .degree = 2, .dec = 1, .want = &[_]f64{ 1.826156, -0.938849 } },
        .{ .name = "sinusoids/3", .input = &test_input_four_sinusoids, .degree = 3, .dec = 1, .want = &[_]f64{ 2.753231, -2.740306, 0.985501 } },
        .{ .name = "sinusoids/4", .input = &test_input_four_sinusoids, .degree = 4, .dec = 1, .want = &[_]f64{ 3.736794, -5.474295, 3.731127, -0.996783 } },
        .{ .name = "test1/5", .input = &test_input_test1, .degree = 5, .dec = 1, .want = &[_]f64{ 1.4, -0.7, 0.04, 0.7, -0.5 } },
        .{ .name = "test2/7", .input = &test_input_test2, .degree = 7, .dec = 0, .want = &[_]f64{ 0.677, 0.175, 0.297, 0.006, -0.114, -0.083, -0.025 } },
        .{ .name = "test3/2", .input = &test_input_test3, .degree = 2, .dec = 1, .want = &[_]f64{ 1.02, -0.53 } },
    };

    for (cases) |tc| {
        if (tc.input.len == 0) continue;

        const length: i32 = @intCast(tc.input.len);
        var x = try MaximumEntropySpectrum.init(testing.allocator, .{
            .length = length,
            .degree = tc.degree,
            .min_period = 2,
            .max_period = @floatFromInt(length * 2),
            .spectrum_resolution = 1,
        });
        defer x.deinit();

        // Copy input data into the estimator's input_series.
        @memcpy(x.input_series, tc.input);
        // Set mean to 0, copy as-is for input_series_minus_mean.
        @memcpy(x.input_series_minus_mean, tc.input);

        // Run calculate which calls burgEstimate internally.
        x.calculate();

        const coefs = x.getCoefficients();
        try testing.expectEqual(tc.want.len, coefs.len);

        for (tc.want, 0..) |w, i| {
            const got = roundDec(coefs[i], tc.dec);
            const exp = roundDec(w, tc.dec);
            try testing.expect(got == exp);
        }
    }
}
