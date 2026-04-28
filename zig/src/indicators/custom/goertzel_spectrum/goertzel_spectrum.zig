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

/// Enumerates the outputs of the GoertzelSpectrum indicator.
pub const GoertzelSpectrumOutput = enum(u8) {
    value = 1,
};

/// Parameters to create a GoertzelSpectrum indicator.
pub const Params = struct {
    length: i32 = 0,
    min_period: f64 = 0,
    max_period: f64 = 0,
    spectrum_resolution: i32 = 0,
    is_first_order: bool = false,
    disable_spectral_dilation_compensation: bool = false,
    disable_automatic_gain_control: bool = false,
    automatic_gain_control_decay_factor: f64 = 0,
    fixed_normalization: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Error = error{
    InvalidLength,
    InvalidMinPeriod,
    InvalidMaxPeriod,
    InvalidNyquist,
    InvalidResolution,
    InvalidAgcDecay,
    MnemonicTooLong,
    OutOfMemory,
};

/// Goertzel Spectrum heatmap indicator.
pub const GoertzelSpectrum = struct {
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
    spectrum_resolution: i32,
    length_spectrum: i32,
    min_period: f64,
    max_period: f64,
    is_first_order: bool,
    is_spectral_dilation_compensation: bool,
    is_automatic_gain_control: bool,
    automatic_gain_control_decay_factor: f64,

    input_series: []f64,
    input_series_minus_mean: []f64,
    spectrum: []f64,
    period: []f64,

    // Trig tables.
    frequency_sin: ?[]f64, // first-order only
    frequency_cos: ?[]f64, // first-order only
    frequency_cos2: ?[]f64, // second-order only

    spectrum_min: f64,
    spectrum_max: f64,
    previous_spectrum_max: f64,

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!GoertzelSpectrum {
        const def_length: i32 = 64;
        const def_min_period: f64 = 2.0;
        const def_max_period: f64 = 64.0;
        const def_spectrum_resolution: i32 = 1;
        const def_agc_decay: f64 = 0.991;
        const agc_decay_epsilon: f64 = 1e-12;
        const two_pi: comptime_float = 2.0 * math.pi;

        var length = params.length;
        if (length == 0) length = def_length;

        var min_period = params.min_period;
        if (min_period == 0) min_period = def_min_period;

        var max_period = params.max_period;
        if (max_period == 0) max_period = def_max_period;

        var spectrum_resolution = params.spectrum_resolution;
        if (spectrum_resolution == 0) spectrum_resolution = def_spectrum_resolution;

        var agc_decay = params.automatic_gain_control_decay_factor;
        if (agc_decay == 0) agc_decay = def_agc_decay;

        const sdc_on = !params.disable_spectral_dilation_compensation;
        const agc_on = !params.disable_automatic_gain_control;
        const floating_norm = !params.fixed_normalization;

        // Validation.
        if (length < 2) return error.InvalidLength;
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

        if (params.is_first_order) {
            const tag = ", fo";
            @memcpy(flags_buf[flags_len .. flags_len + tag.len], tag);
            flags_len += tag.len;
        }
        if (!sdc_on) {
            const tag = ", no-sdc";
            @memcpy(flags_buf[flags_len .. flags_len + tag.len], tag);
            flags_len += tag.len;
        }
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
        const mn = std.fmt.bufPrint(&mnemonic_buf, "gspect({d}, {d}, {d}, {d}{s}{s})", .{
            length,
            min_period,
            max_period,
            spectrum_resolution,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Goertzel spectrum {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        // Allocate estimator arrays.
        const length_u: usize = @intCast(length);
        const length_spectrum: i32 = @intFromFloat((max_period - min_period) * @as(f64, @floatFromInt(spectrum_resolution)) + 1);
        const ls_u: usize = @intCast(length_spectrum);
        const result: f64 = @floatFromInt(spectrum_resolution);

        const input_series = allocator.alloc(f64, length_u) catch return error.OutOfMemory;
        errdefer allocator.free(input_series);
        const input_series_minus_mean = allocator.alloc(f64, length_u) catch return error.OutOfMemory;
        errdefer allocator.free(input_series_minus_mean);
        const spectrum_arr = allocator.alloc(f64, ls_u) catch return error.OutOfMemory;
        errdefer allocator.free(spectrum_arr);
        const period_arr = allocator.alloc(f64, ls_u) catch return error.OutOfMemory;
        errdefer allocator.free(period_arr);

        var frequency_sin: ?[]f64 = null;
        var frequency_cos: ?[]f64 = null;
        var frequency_cos2: ?[]f64 = null;

        if (params.is_first_order) {
            frequency_sin = allocator.alloc(f64, ls_u) catch return error.OutOfMemory;
            errdefer allocator.free(frequency_sin.?);
            frequency_cos = allocator.alloc(f64, ls_u) catch return error.OutOfMemory;
            errdefer allocator.free(frequency_cos.?);

            for (0..ls_u) |i| {
                const p = max_period - @as(f64, @floatFromInt(i)) / result;
                period_arr[i] = p;
                const theta = two_pi / p;
                frequency_sin.?[i] = @sin(theta);
                frequency_cos.?[i] = @cos(theta);
            }
        } else {
            frequency_cos2 = allocator.alloc(f64, ls_u) catch return error.OutOfMemory;
            errdefer allocator.free(frequency_cos2.?);

            for (0..ls_u) |i| {
                const p = max_period - @as(f64, @floatFromInt(i)) / result;
                period_arr[i] = p;
                frequency_cos2.?[i] = 2.0 * @cos(two_pi / p);
            }
        }

        @memset(input_series, 0);
        @memset(input_series_minus_mean, 0);
        @memset(spectrum_arr, 0);

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
            .spectrum_resolution = spectrum_resolution,
            .length_spectrum = length_spectrum,
            .min_period = min_period,
            .max_period = max_period,
            .is_first_order = params.is_first_order,
            .is_spectral_dilation_compensation = sdc_on,
            .is_automatic_gain_control = agc_on,
            .automatic_gain_control_decay_factor = agc_decay,
            .input_series = input_series,
            .input_series_minus_mean = input_series_minus_mean,
            .spectrum = spectrum_arr,
            .period = period_arr,
            .frequency_sin = frequency_sin,
            .frequency_cos = frequency_cos,
            .frequency_cos2 = frequency_cos2,
            .spectrum_min = 0,
            .spectrum_max = 0,
            .previous_spectrum_max = 0,
        };
    }

    pub fn deinit(self: *GoertzelSpectrum) void {
        self.allocator.free(self.input_series);
        self.allocator.free(self.input_series_minus_mean);
        self.allocator.free(self.spectrum);
        self.allocator.free(self.period);
        if (self.frequency_sin) |s| self.allocator.free(s);
        if (self.frequency_cos) |c| self.allocator.free(c);
        if (self.frequency_cos2) |c| self.allocator.free(c);
    }

    pub fn fixSlices(self: *GoertzelSpectrum) void {
        _ = self;
    }

    fn mnemonicSlice(self: *const GoertzelSpectrum) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn descriptionSlice(self: *const GoertzelSpectrum) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    // --- Estimator methods ---

    fn goertzelSecondOrderEstimate(self: *const GoertzelSpectrum, j: usize) f64 {
        const cos2 = self.frequency_cos2.?[j];
        var s1: f64 = 0;
        var s2: f64 = 0;
        const len_u: usize = @intCast(self.length);

        for (0..len_u) |i| {
            const s0 = self.input_series_minus_mean[i] + cos2 * s1 - s2;
            s2 = s1;
            s1 = s0;
        }

        const sp = s1 * s1 + s2 * s2 - cos2 * s1 * s2;
        if (sp < 0) return 0;
        return sp;
    }

    fn goertzelFirstOrderEstimate(self: *const GoertzelSpectrum, j: usize) f64 {
        const cos_theta = self.frequency_cos.?[j];
        const sin_theta = self.frequency_sin.?[j];
        var yre: f64 = 0;
        var yim: f64 = 0;
        const len_u: usize = @intCast(self.length);

        for (0..len_u) |i| {
            const re = self.input_series_minus_mean[i] + cos_theta * yre - sin_theta * yim;
            const im = self.input_series_minus_mean[i] + cos_theta * yim + sin_theta * yre;
            yre = re;
            yim = im;
        }

        return yre * yre + yim * yim;
    }

    fn goertzelEstimate(self: *const GoertzelSpectrum, j: usize) f64 {
        if (self.is_first_order) return self.goertzelFirstOrderEstimate(j);
        return self.goertzelSecondOrderEstimate(j);
    }

    fn calculate(self: *GoertzelSpectrum) void {
        const len_u: usize = @intCast(self.length);
        const ls_u: usize = @intCast(self.length_spectrum);

        // Subtract mean.
        var mean: f64 = 0;
        for (0..len_u) |i| {
            mean += self.input_series[i];
        }
        mean /= @floatFromInt(self.length);

        for (0..len_u) |i| {
            self.input_series_minus_mean[i] = self.input_series[i] - mean;
        }

        // Seed with first bin.
        var sp = self.goertzelEstimate(0);
        if (self.is_spectral_dilation_compensation) {
            sp /= self.period[0];
        }
        self.spectrum[0] = sp;
        self.spectrum_min = sp;

        if (self.is_automatic_gain_control) {
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max;
            if (self.spectrum_max < sp) {
                self.spectrum_max = sp;
            }
        } else {
            self.spectrum_max = sp;
        }

        for (1..ls_u) |i| {
            sp = self.goertzelEstimate(i);
            if (self.is_spectral_dilation_compensation) {
                sp /= self.period[i];
            }
            self.spectrum[i] = sp;

            if (self.spectrum_max < sp) {
                self.spectrum_max = sp;
            } else if (self.spectrum_min > sp) {
                self.spectrum_min = sp;
            }
        }

        self.previous_spectrum_max = self.spectrum_max;
    }

    // --- Public interface ---

    pub fn updateSample(self: *GoertzelSpectrum, sample: f64, time: i64) Heatmap {
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

    pub fn updateBar(self: *GoertzelSpectrum, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *GoertzelSpectrum, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *GoertzelSpectrum, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    pub fn updateScalar(self: *GoertzelSpectrum, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    fn updateEntity(self: *GoertzelSpectrum, time: i64, sample: f64) OutputArray {
        const hm = self.updateSample(sample, time);
        var out = OutputArray{};
        out.append(.{ .heatmap = hm });
        return out;
    }

    pub fn isPrimed(self: *const GoertzelSpectrum) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const GoertzelSpectrum, out: *Metadata) void {
        const mn = self.mnemonicSlice();
        const desc = self.descriptionSlice();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
        };
        build_metadata_mod.buildMetadata(out, .goertzel_spectrum, mn, desc, &texts);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *GoertzelSpectrum) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(GoertzelSpectrum);
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

const test_tolerance = 1e-10;
const test_min_max_tol = 1e-9;

fn almostEqual(a: f64, b: f64, eps: f64) bool {
    return @abs(a - b) <= eps;
}

const SpotValue = struct { i: usize, v: f64 };
const GsSnap = struct {
    i: usize,
    value_min: f64,
    value_max: f64,
    spots: []const SpotValue,
};

const goertzel_snapshots = [_]GsSnap{
    .{
        .i = 63,
        .value_min = 0,
        .value_max = 1,
        .spots = &[_]SpotValue{
            .{ .i = 0, .v = 0.002212390126817 },
            .{ .i = 15, .v = 0.393689637083521 },
            .{ .i = 31, .v = 0.561558825583766 },
            .{ .i = 47, .v = 0.486814514368002 },
            .{ .i = 62, .v = 0.487856217300954 },
        },
    },
    .{
        .i = 64,
        .value_min = 0,
        .value_max = 0.9945044963,
        .spots = &[_]SpotValue{
            .{ .i = 0, .v = 0.006731833921830 },
            .{ .i = 15, .v = 0.435945652220356 },
            .{ .i = 31, .v = 0.554419782890674 },
            .{ .i = 47, .v = 0.489761317874540 },
            .{ .i = 62, .v = 0.490802995079533 },
        },
    },
    .{
        .i = 100,
        .value_min = 0,
        .value_max = 1,
        .spots = &[_]SpotValue{
            .{ .i = 0, .v = 0.008211812272033 },
            .{ .i = 15, .v = 0.454499290767355 },
            .{ .i = 31, .v = 0.450815700228196 },
            .{ .i = 47, .v = 0.432349912501093 },
            .{ .i = 62, .v = 1.0 },
        },
    },
    .{
        .i = 150,
        .value_min = 0,
        .value_max = 0.4526639264,
        .spots = &[_]SpotValue{
            .{ .i = 0, .v = 0.003721075091811 },
            .{ .i = 15, .v = 0.050467362919035 },
            .{ .i = 31, .v = 0.053328277804150 },
            .{ .i = 47, .v = 0.351864884608844 },
            .{ .i = 62, .v = 0.451342692411903 },
        },
    },
    .{
        .i = 200,
        .value_min = 0,
        .value_max = 0.5590969243,
        .spots = &[_]SpotValue{
            .{ .i = 0, .v = 0.041810380001389 },
            .{ .i = 15, .v = 0.388762084039364 },
            .{ .i = 31, .v = 0.412461432112096 },
            .{ .i = 47, .v = 0.446271463994143 },
            .{ .i = 62, .v = 0.280061782526868 },
        },
    },
};

test "GoertzelSpectrum update" {
    var x = try GoertzelSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    var si: usize = 0;
    for (0..test_input.len) |i| {
        const h = x.updateSample(test_input[i], @intCast(i));

        try testing.expectEqual(@as(f64, 2.0), h.parameter_first);
        try testing.expectEqual(@as(f64, 64.0), h.parameter_last);
        try testing.expectEqual(@as(f64, 1.0), h.parameter_resolution);

        if (!x.isPrimed()) {
            try testing.expect(h.isEmpty());
            continue;
        }

        try testing.expectEqual(@as(usize, 63), h.values_len);

        if (si < goertzel_snapshots.len and goertzel_snapshots[si].i == i) {
            const snap = goertzel_snapshots[si];
            try testing.expect(almostEqual(h.value_min, snap.value_min, test_min_max_tol));
            try testing.expect(almostEqual(h.value_max, snap.value_max, test_min_max_tol));

            const vals = h.valuesSlice();
            for (snap.spots) |sp| {
                try testing.expect(almostEqual(vals[sp.i], sp.v, test_tolerance));
            }

            si += 1;
        }
    }

    try testing.expectEqual(goertzel_snapshots.len, si);
}

test "GoertzelSpectrum primes at bar 63" {
    var x = try GoertzelSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    try testing.expect(!x.isPrimed());

    var primed_at: ?usize = null;
    for (0..test_input.len) |i| {
        _ = x.updateSample(test_input[i], @intCast(i));
        if (x.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expectEqual(@as(usize, 63), primed_at.?);
}

test "GoertzelSpectrum NaN input" {
    var x = try GoertzelSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    const h = x.updateSample(math.nan(f64), 0);
    try testing.expect(h.isEmpty());
    try testing.expect(!x.isPrimed());
}

test "GoertzelSpectrum metadata" {
    var x = try GoertzelSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    var md: Metadata = undefined;
    x.getMetadata(&md);

    const mn = "gspect(64, 2, 64, 1, hl/2)";

    try testing.expectEqualStrings(mn, x.mnemonicSlice());
    try testing.expectEqual(Identifier.goertzel_spectrum, md.identifier);
    try testing.expectEqualStrings(mn, md.mnemonic);
    try testing.expectEqual(@as(usize, 1), md.outputs_len);

    const outputs = md.outputs_buf[0..md.outputs_len];
    try testing.expectEqualStrings(mn, outputs[0].mnemonic);
}

test "GoertzelSpectrum mnemonic flags" {
    const TestCase = struct { params: Params, mn: []const u8 };
    const cases = [_]TestCase{
        .{ .params = .{}, .mn = "gspect(64, 2, 64, 1, hl/2)" },
        .{ .params = .{ .is_first_order = true }, .mn = "gspect(64, 2, 64, 1, fo, hl/2)" },
        .{ .params = .{ .disable_spectral_dilation_compensation = true }, .mn = "gspect(64, 2, 64, 1, no-sdc, hl/2)" },
        .{ .params = .{ .disable_automatic_gain_control = true }, .mn = "gspect(64, 2, 64, 1, no-agc, hl/2)" },
        .{ .params = .{ .automatic_gain_control_decay_factor = 0.8 }, .mn = "gspect(64, 2, 64, 1, agc=0.8, hl/2)" },
        .{ .params = .{ .fixed_normalization = true }, .mn = "gspect(64, 2, 64, 1, no-fn, hl/2)" },
        .{ .params = .{
            .is_first_order = true,
            .disable_spectral_dilation_compensation = true,
            .disable_automatic_gain_control = true,
            .fixed_normalization = true,
        }, .mn = "gspect(64, 2, 64, 1, fo, no-sdc, no-agc, no-fn, hl/2)" },
    };

    for (cases) |tc| {
        var x = try GoertzelSpectrum.init(testing.allocator, tc.params);
        defer x.deinit();
        try testing.expectEqualStrings(tc.mn, x.mnemonicSlice());
    }
}

test "GoertzelSpectrum validation" {
    try testing.expectError(error.InvalidLength, GoertzelSpectrum.init(testing.allocator, .{
        .length = 1,
        .min_period = 2,
        .max_period = 64,
        .spectrum_resolution = 1,
    }));
    try testing.expectError(error.InvalidMinPeriod, GoertzelSpectrum.init(testing.allocator, .{
        .length = 64,
        .min_period = 1,
        .max_period = 64,
        .spectrum_resolution = 1,
    }));
    try testing.expectError(error.InvalidMaxPeriod, GoertzelSpectrum.init(testing.allocator, .{
        .length = 64,
        .min_period = 10,
        .max_period = 10,
        .spectrum_resolution = 1,
    }));
    try testing.expectError(error.InvalidNyquist, GoertzelSpectrum.init(testing.allocator, .{
        .length = 16,
        .min_period = 2,
        .max_period = 64,
        .spectrum_resolution = 1,
    }));
    try testing.expectError(error.InvalidAgcDecay, GoertzelSpectrum.init(testing.allocator, .{
        .automatic_gain_control_decay_factor = -0.1,
    }));
    try testing.expectError(error.InvalidAgcDecay, GoertzelSpectrum.init(testing.allocator, .{
        .automatic_gain_control_decay_factor = 1.0,
    }));
}

test "GoertzelSpectrum updateEntity" {
    const prime_count = 70;
    const inp: f64 = 100.0;
    const time: i64 = 0;

    // Update scalar
    {
        var x = try GoertzelSpectrum.init(testing.allocator, .{});
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
        var x = try GoertzelSpectrum.init(testing.allocator, .{});
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
        var x = try GoertzelSpectrum.init(testing.allocator, .{});
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
        var x = try GoertzelSpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const t = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = x.updateTrade(&t);
        try testing.expectEqual(@as(usize, 1), out.len);
    }
}
