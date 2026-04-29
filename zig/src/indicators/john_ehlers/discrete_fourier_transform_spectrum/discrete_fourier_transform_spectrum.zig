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

/// Enumerates the outputs of the discrete Fourier transform spectrum.
pub const DiscreteFourierTransformSpectrumOutput = enum(u8) {
    value = 1,
};

/// Parameters to create a DiscreteFourierTransformSpectrum.
pub const Params = struct {
    /// Length of the spectrum window. Must be >= 2. Default is 48.
    length: i32 = 48,
    /// Minimum cycle period. Must be >= 2. Default is 10.
    min_period: f64 = 10.0,
    /// Maximum cycle period. Must be > min_period and <= 2*length. Default is 48.
    max_period: f64 = 48.0,
    /// Spectrum resolution. Must be >= 1. Default is 1.
    spectrum_resolution: i32 = 1,
    /// Disable spectral dilation compensation.
    disable_spectral_dilation_compensation: bool = false,
    /// Disable automatic gain control.
    disable_automatic_gain_control: bool = false,
    /// AGC decay factor in (0, 1). Default is 0.995.
    automatic_gain_control_decay_factor: f64 = 0.995,
    /// Use fixed normalization (min clamped to 0) instead of floating.
    fixed_normalization: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// DFT power spectrum estimator.
const Estimator = struct {
    allocator: std.mem.Allocator,
    length: usize,
    spectrum_resolution: usize,
    length_spectrum: usize,
    max_omega_length: usize,
    min_period: f64,
    max_period: f64,
    is_spectral_dilation_compensation: bool,
    is_automatic_gain_control: bool,
    automatic_gain_control_decay_factor: f64,

    input_series: []f64,
    input_series_minus_mean: []f64,
    spectrum: []f64,
    period: []f64,

    // Pre-computed trig tables: flattened [length_spectrum][max_omega_length].
    frequency_sin_omega: []f64,
    frequency_cos_omega: []f64,

    mean: f64,
    spectrum_min: f64,
    spectrum_max: f64,
    previous_spectrum_max: f64,

    fn init(
        allocator: std.mem.Allocator,
        length: usize,
        min_period: f64,
        max_period: f64,
        spectrum_resolution: usize,
        is_sdc: bool,
        is_agc: bool,
        agc_decay: f64,
    ) !Estimator {
        const two_pi = 2.0 * math.pi;
        const length_spectrum: usize = @intFromFloat((max_period - min_period) * @as(f64, @floatFromInt(spectrum_resolution)) + 1.0);
        const max_omega_length = length;

        const input_series = try allocator.alloc(f64, length);
        @memset(input_series, 0.0);
        const input_series_minus_mean = try allocator.alloc(f64, length);
        @memset(input_series_minus_mean, 0.0);
        const spectrum_buf = try allocator.alloc(f64, length_spectrum);
        @memset(spectrum_buf, 0.0);
        const period_buf = try allocator.alloc(f64, length_spectrum);

        const trig_size = length_spectrum * max_omega_length;
        const sin_buf = try allocator.alloc(f64, trig_size);
        const cos_buf = try allocator.alloc(f64, trig_size);

        const result: f64 = @floatFromInt(spectrum_resolution);

        for (0..length_spectrum) |i| {
            const fi: f64 = @floatFromInt(i);
            const p = max_period - fi / result;
            period_buf[i] = p;
            const theta = two_pi / p;

            const row_offset = i * max_omega_length;
            for (0..max_omega_length) |j| {
                const omega = @as(f64, @floatFromInt(j)) * theta;
                sin_buf[row_offset + j] = @sin(omega);
                cos_buf[row_offset + j] = @cos(omega);
            }
        }

        return .{
            .allocator = allocator,
            .length = length,
            .spectrum_resolution = spectrum_resolution,
            .length_spectrum = length_spectrum,
            .max_omega_length = max_omega_length,
            .min_period = min_period,
            .max_period = max_period,
            .is_spectral_dilation_compensation = is_sdc,
            .is_automatic_gain_control = is_agc,
            .automatic_gain_control_decay_factor = agc_decay,
            .input_series = input_series,
            .input_series_minus_mean = input_series_minus_mean,
            .spectrum = spectrum_buf,
            .period = period_buf,
            .frequency_sin_omega = sin_buf,
            .frequency_cos_omega = cos_buf,
            .mean = 0.0,
            .spectrum_min = 0.0,
            .spectrum_max = 0.0,
            .previous_spectrum_max = 0.0,
        };
    }

    fn deinit(self: *Estimator) void {
        self.allocator.free(self.input_series);
        self.allocator.free(self.input_series_minus_mean);
        self.allocator.free(self.spectrum);
        self.allocator.free(self.period);
        self.allocator.free(self.frequency_sin_omega);
        self.allocator.free(self.frequency_cos_omega);
    }

    fn calculate(self: *Estimator) void {
        // Subtract the mean from the input series.
        var mean: f64 = 0.0;
        for (0..self.length) |i| {
            mean += self.input_series[i];
        }
        mean /= @as(f64, @floatFromInt(self.length));

        for (0..self.length) |i| {
            self.input_series_minus_mean[i] = self.input_series[i] - mean;
        }
        self.mean = mean;

        // Evaluate the DFT power spectrum.
        self.spectrum_min = math.floatMax(f64);
        if (self.is_automatic_gain_control) {
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max;
        } else {
            self.spectrum_max = -math.floatMax(f64);
        }

        for (0..self.length_spectrum) |i| {
            const row_offset = i * self.max_omega_length;

            var sum_sin: f64 = 0.0;
            var sum_cos: f64 = 0.0;

            for (0..self.max_omega_length) |j| {
                const sample = self.input_series_minus_mean[j];
                sum_sin += sample * self.frequency_sin_omega[row_offset + j];
                sum_cos += sample * self.frequency_cos_omega[row_offset + j];
            }

            var s = sum_sin * sum_sin + sum_cos * sum_cos;
            if (self.is_spectral_dilation_compensation) {
                s /= self.period[i];
            }

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
};

/// MBST's Discrete Fourier Transform Spectrum heatmap indicator.
pub const DiscreteFourierTransformSpectrum = struct {
    allocator: std.mem.Allocator,
    estimator: Estimator,
    window_count: usize,
    last_index: usize,
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

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!DiscreteFourierTransformSpectrum {
        const agc_decay_epsilon: f64 = 1e-12;
        const def_agc_decay: f64 = 0.995;

        if (params.length < 2) return error.InvalidLength;
        if (params.min_period < 2.0) return error.InvalidMinPeriod;
        if (params.max_period <= params.min_period) return error.InvalidMaxPeriod;
        if (params.max_period > 2.0 * @as(f64, @floatFromInt(params.length))) return error.InvalidNyquist;
        if (params.spectrum_resolution < 1) return error.InvalidResolution;

        const agc_on = !params.disable_automatic_gain_control;
        if (agc_on and (params.automatic_gain_control_decay_factor <= 0.0 or params.automatic_gain_control_decay_factor >= 1.0)) {
            return error.InvalidAgcDecay;
        }

        const sdc_on = !params.disable_spectral_dilation_compensation;
        const floating_norm = !params.fixed_normalization;

        const bc = params.bar_component orelse bar_component.BarComponent.median;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        const length: usize = @intCast(params.length);

        var estimator = Estimator.init(
            allocator,
            length,
            params.min_period,
            params.max_period,
            @intCast(params.spectrum_resolution),
            sdc_on,
            agc_on,
            params.automatic_gain_control_decay_factor,
        ) catch return error.OutOfMemory;
        errdefer estimator.deinit();

        // Build mnemonic with flag tags.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc,
            qc,
            tc,
        );

        // Build flags string.
        var flags_buf: [128]u8 = undefined;
        var flags_len: usize = 0;

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
        if (agc_on and @abs(params.automatic_gain_control_decay_factor - def_agc_decay) > agc_decay_epsilon) {
            const agc_tag = std.fmt.bufPrint(flags_buf[flags_len..], ", agc={d}", .{params.automatic_gain_control_decay_factor}) catch
                return error.MnemonicTooLong;
            flags_len += agc_tag.len;
        }
        if (!floating_norm) {
            const tag = ", no-fn";
            @memcpy(flags_buf[flags_len .. flags_len + tag.len], tag);
            flags_len += tag.len;
        }

        const flags = flags_buf[0..flags_len];

        var mnemonic_buf: [256]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "dftps({d}, {d}, {d}, {d}{s}{s})", .{
            params.length,
            params.min_period,
            params.max_period,
            params.spectrum_resolution,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Discrete Fourier transform spectrum {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        return .{
            .allocator = allocator,
            .estimator = estimator,
            .window_count = 0,
            .last_index = length - 1,
            .primed = false,
            .floating_normalization = floating_norm,
            .min_parameter_value = params.min_period,
            .max_parameter_value = params.max_period,
            .parameter_resolution = @floatFromInt(params.spectrum_resolution),
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *DiscreteFourierTransformSpectrum) void {
        self.estimator.deinit();
    }

    pub fn fixSlices(self: *DiscreteFourierTransformSpectrum) void {
        _ = self;
    }

    fn mnemonic(self: *const DiscreteFourierTransformSpectrum) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const DiscreteFourierTransformSpectrum) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    /// Update with a new sample value and return the heatmap column.
    pub fn update(self: *DiscreteFourierTransformSpectrum, sample: f64, time: i64) Heatmap {
        if (math.isNan(sample)) {
            return Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution);
        }

        const window = self.estimator.input_series;

        if (self.primed) {
            std.mem.copyForwards(f64, window[0..self.last_index], window[1..self.estimator.length]);
            window[self.last_index] = sample;
        } else {
            window[self.window_count] = sample;
            self.window_count += 1;

            if (self.window_count == self.estimator.length) {
                self.primed = true;
            }
        }

        if (!self.primed) {
            return Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution);
        }

        self.estimator.calculate();

        const length_spectrum = self.estimator.length_spectrum;

        var min_ref: f64 = 0.0;
        if (self.floating_normalization) {
            min_ref = self.estimator.spectrum_min;
        }

        const max_ref = self.estimator.spectrum_max;
        const spectrum_range = max_ref - min_ref;

        // Reverse: spectrum[0] is at MaxPeriod, heatmap axis is MinPeriod -> MaxPeriod.
        var values: [heatmap_mod.max_heatmap_values]f64 = undefined;
        var value_min: f64 = math.inf(f64);
        var value_max: f64 = -math.inf(f64);

        for (0..length_spectrum) |i| {
            const v = (self.estimator.spectrum[length_spectrum - 1 - i] - min_ref) / spectrum_range;
            values[i] = v;

            if (v < value_min) {
                value_min = v;
            }
            if (v > value_max) {
                value_max = v;
            }
        }

        return Heatmap.new(
            time,
            self.min_parameter_value,
            self.max_parameter_value,
            self.parameter_resolution,
            value_min,
            value_max,
            values[0..length_spectrum],
        );
    }

    // --- Entity update methods ---

    pub fn updateBar(self: *DiscreteFourierTransformSpectrum, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *DiscreteFourierTransformSpectrum, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *DiscreteFourierTransformSpectrum, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    pub fn updateScalar(self: *DiscreteFourierTransformSpectrum, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    fn updateEntity(self: *DiscreteFourierTransformSpectrum, time: i64, sample: f64) OutputArray {
        const h = self.update(sample, time);
        var out = OutputArray{};
        out.append(.{ .heatmap = h });
        return out;
    }

    pub fn isPrimed(self: *const DiscreteFourierTransformSpectrum) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const DiscreteFourierTransformSpectrum, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },

        };
        build_metadata_mod.buildMetadata(out, .discrete_fourier_transform_spectrum, mn, desc, &texts);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *DiscreteFourierTransformSpectrum) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(DiscreteFourierTransformSpectrum);
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

const DftsSpot = struct {
    i: usize,
    v: f64,
};

const DftsSnap = struct {
    i: usize,
    value_min: f64,
    value_max: f64,
    spots: []const DftsSpot,
};

const dfts_snapshots = [_]DftsSnap{
    .{
        .i = 47,
        .value_min = 0,
        .value_max = 1,
        .spots = &[_]DftsSpot{
            .{ .i = 0, .v = 0.092590275198666 },
            .{ .i = 9, .v = 0.027548278511410 },
            .{ .i = 19, .v = 0.157582436454944 },
            .{ .i = 28, .v = 0.818873586056752 },
            .{ .i = 38, .v = 0.994469163657834 },
        },
    },
    .{
        .i = 60,
        .value_min = 0,
        .value_max = 1,
        .spots = &[_]DftsSpot{
            .{ .i = 0, .v = 0.051096698360059 },
            .{ .i = 9, .v = 0.058377239289306 },
            .{ .i = 19, .v = 0.248187339190831 },
            .{ .i = 28, .v = 0.136096222727665 },
            .{ .i = 38, .v = 0.059069344951534 },
        },
    },
    .{
        .i = 100,
        .value_min = 0,
        .value_max = 1,
        .spots = &[_]DftsSpot{
            .{ .i = 0, .v = 0.150638336509665 },
            .{ .i = 9, .v = 0.046435987728045 },
            .{ .i = 19, .v = 0.103832850895319 },
            .{ .i = 28, .v = 0.402801132284104 },
            .{ .i = 38, .v = 1.000000000000000 },
        },
    },
    .{
        .i = 150,
        .value_min = 0,
        .value_max = 0.5272269971142493,
        .spots = &[_]DftsSpot{
            .{ .i = 0, .v = 0.000000000000000 },
            .{ .i = 9, .v = 0.091857989427651 },
            .{ .i = 19, .v = 0.219395988856534 },
            .{ .i = 28, .v = 0.516960894560452 },
            .{ .i = 38, .v = 0.468682020733700 },
        },
    },
    .{
        .i = 200,
        .value_min = 0,
        .value_max = 0.6015223942655807,
        .spots = &[_]DftsSpot{
            .{ .i = 0, .v = 0.107853213261092 },
            .{ .i = 9, .v = 0.164118955219278 },
            .{ .i = 19, .v = 0.306440501928972 },
            .{ .i = 28, .v = 0.569768020155262 },
            .{ .i = 38, .v = 0.585690371992475 },
        },
    },
};

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

test "DiscreteFourierTransformSpectrum update" {
    const tolerance = 1e-12;
    const min_max_tol = 1e-10;

    var x = try DiscreteFourierTransformSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    var si: usize = 0;

    for (0..test_input.len) |i| {
        const h = x.update(test_input[i], @intCast(i));

        try testing.expectEqual(@as(f64, 10.0), h.parameter_first);
        try testing.expectEqual(@as(f64, 48.0), h.parameter_last);
        try testing.expectEqual(@as(f64, 1.0), h.parameter_resolution);

        if (!x.primed) {
            try testing.expect(h.isEmpty());
            continue;
        }

        try testing.expectEqual(@as(usize, 39), h.values_len);

        if (si < dfts_snapshots.len and dfts_snapshots[si].i == i) {
            const snap = dfts_snapshots[si];
            try testing.expect(almostEqual(h.value_min, snap.value_min, min_max_tol));
            try testing.expect(almostEqual(h.value_max, snap.value_max, min_max_tol));

            const vals = h.valuesSlice();
            for (snap.spots) |sp| {
                try testing.expect(almostEqual(vals[sp.i], sp.v, tolerance));
            }

            si += 1;
        }
    }

    try testing.expectEqual(dfts_snapshots.len, si);
}

test "DiscreteFourierTransformSpectrum primes at bar 47" {
    var x = try DiscreteFourierTransformSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    try testing.expect(!x.isPrimed());

    var primed_at: ?usize = null;

    for (0..test_input.len) |i| {
        _ = x.update(test_input[i], @intCast(i));
        if (x.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expectEqual(@as(usize, 47), primed_at.?);
}

test "DiscreteFourierTransformSpectrum NaN input" {
    var x = try DiscreteFourierTransformSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    const h = x.update(math.nan(f64), 0);
    try testing.expect(h.isEmpty());
    try testing.expect(!x.isPrimed());
}

test "DiscreteFourierTransformSpectrum synthetic sine" {
    const period = 16.0;
    const bars = 200;

    var x = try DiscreteFourierTransformSpectrum.init(testing.allocator, .{
        .disable_spectral_dilation_compensation = true,
        .disable_automatic_gain_control = true,
        .fixed_normalization = true,
    });
    defer x.deinit();

    var last: Heatmap = undefined;

    for (0..bars) |i| {
        const sample = 100.0 + @sin(2.0 * math.pi * @as(f64, @floatFromInt(i)) / period);
        last = x.update(sample, @intCast(i));
    }

    try testing.expect(!last.isEmpty());

    // Peak bin should correspond to period=16. Bin k corresponds to period MinPeriod+k.
    // With default MinPeriod=10, period=16 -> bin index 6.
    var peak_bin: usize = 0;
    const vals = last.valuesSlice();
    for (1..vals.len) |i| {
        if (vals[i] > vals[peak_bin]) {
            peak_bin = i;
        }
    }

    const expected_bin: usize = @intFromFloat(period - last.parameter_first);
    try testing.expectEqual(expected_bin, peak_bin);
}

test "DiscreteFourierTransformSpectrum metadata" {
    var x = try DiscreteFourierTransformSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    var md: Metadata = undefined;
    x.getMetadata(&md);

    const mn = "dftps(48, 10, 48, 1, hl/2)";
    try testing.expectEqualStrings(mn, x.mnemonic());
    try testing.expectEqual(Identifier.discrete_fourier_transform_spectrum, md.identifier);
    try testing.expectEqualStrings(mn, md.mnemonic);
    try testing.expectEqual(@as(usize, 1), md.outputs_len);
}

test "DiscreteFourierTransformSpectrum mnemonic flags" {
    const TestCase = struct {
        params: Params,
        expected: []const u8,
    };

    const cases = [_]TestCase{
        .{ .params = .{}, .expected = "dftps(48, 10, 48, 1, hl/2)" },
        .{ .params = .{ .disable_spectral_dilation_compensation = true }, .expected = "dftps(48, 10, 48, 1, no-sdc, hl/2)" },
        .{ .params = .{ .disable_automatic_gain_control = true }, .expected = "dftps(48, 10, 48, 1, no-agc, hl/2)" },
        .{ .params = .{ .automatic_gain_control_decay_factor = 0.8 }, .expected = "dftps(48, 10, 48, 1, agc=0.8, hl/2)" },
        .{ .params = .{ .fixed_normalization = true }, .expected = "dftps(48, 10, 48, 1, no-fn, hl/2)" },
        .{ .params = .{ .disable_spectral_dilation_compensation = true, .disable_automatic_gain_control = true, .fixed_normalization = true }, .expected = "dftps(48, 10, 48, 1, no-sdc, no-agc, no-fn, hl/2)" },
    };

    for (cases) |tc| {
        var x = try DiscreteFourierTransformSpectrum.init(testing.allocator, tc.params);
        defer x.deinit();
        try testing.expectEqualStrings(tc.expected, x.mnemonic());
    }
}

test "DiscreteFourierTransformSpectrum validation" {
    // Length < 2
    try testing.expectError(error.InvalidLength, DiscreteFourierTransformSpectrum.init(testing.allocator, .{ .length = 1 }));
    // MinPeriod < 2
    try testing.expectError(error.InvalidMinPeriod, DiscreteFourierTransformSpectrum.init(testing.allocator, .{ .min_period = 1.0 }));
    // MaxPeriod <= MinPeriod
    try testing.expectError(error.InvalidMaxPeriod, DiscreteFourierTransformSpectrum.init(testing.allocator, .{ .min_period = 10.0, .max_period = 10.0 }));
    // MaxPeriod > 2*Length
    try testing.expectError(error.InvalidNyquist, DiscreteFourierTransformSpectrum.init(testing.allocator, .{ .length = 10, .min_period = 2.0, .max_period = 48.0 }));
    // AGC decay <= 0
    try testing.expectError(error.InvalidAgcDecay, DiscreteFourierTransformSpectrum.init(testing.allocator, .{ .automatic_gain_control_decay_factor = -0.1 }));
    // AGC decay >= 1
    try testing.expectError(error.InvalidAgcDecay, DiscreteFourierTransformSpectrum.init(testing.allocator, .{ .automatic_gain_control_decay_factor = 1.0 }));
}

test "DiscreteFourierTransformSpectrum updateEntity" {
    const prime_count = 60;
    const inp: f64 = 100.0;
    const time: i64 = 0;

    // Update scalar
    {
        var x = try DiscreteFourierTransformSpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |i| {
            _ = x.update(test_input[i % test_input.len], time);
        }
        const s = Scalar{ .time = time, .value = inp };
        const out = x.updateScalar(&s);
        try testing.expectEqual(@as(usize, 1), out.len);
    }

    // Update bar
    {
        var x = try DiscreteFourierTransformSpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |i| {
            _ = x.update(test_input[i % test_input.len], time);
        }
        const b = Bar{ .time = time, .open = inp, .high = inp, .low = inp, .close = inp, .volume = 0 };
        const out = x.updateBar(&b);
        try testing.expectEqual(@as(usize, 1), out.len);
    }

    // Update quote
    {
        var x = try DiscreteFourierTransformSpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |i| {
            _ = x.update(test_input[i % test_input.len], time);
        }
        const q = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };
        const out = x.updateQuote(&q);
        try testing.expectEqual(@as(usize, 1), out.len);
    }

    // Update trade
    {
        var x = try DiscreteFourierTransformSpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |i| {
            _ = x.update(test_input[i % test_input.len], time);
        }
        const t = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = x.updateTrade(&t);
        try testing.expectEqual(@as(usize, 1), out.len);
    }
}
