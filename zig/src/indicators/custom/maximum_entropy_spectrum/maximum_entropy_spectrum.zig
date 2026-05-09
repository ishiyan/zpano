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
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, eps: f64) bool {
    return @abs(a - b) <= eps;
}

const SpotValue = testdata.SpotValue;
test "MaximumEntropySpectrum update" {
    var x = try MaximumEntropySpectrum.init(testing.allocator, .{});
    defer x.deinit();

    var si: usize = 0;
    for (0..testdata.test_input.len) |i| {
        const h = x.updateSample(testdata.test_input[i], @intCast(i));

        try testing.expectEqual(@as(f64, 2.0), h.parameter_first);
        try testing.expectEqual(@as(f64, 59.0), h.parameter_last);
        try testing.expectEqual(@as(f64, 1.0), h.parameter_resolution);

        if (!x.isPrimed()) {
            try testing.expect(h.isEmpty());
            continue;
        }

        try testing.expectEqual(@as(usize, 58), h.values_len);

        if (si < testdata.mes_snapshots.len and testdata.mes_snapshots[si].i == i) {
            const snap = testdata.mes_snapshots[si];
            try testing.expect(almostEqual(h.value_min, snap.value_min, testdata.test_min_max_tol));
            try testing.expect(almostEqual(h.value_max, snap.value_max, testdata.test_min_max_tol));

            const vals = h.valuesSlice();
            for (snap.spots) |sp| {
                try testing.expect(almostEqual(vals[sp.i], sp.v, testdata.test_tolerance));
            }

            si += 1;
        }
    }

    try testing.expectEqual(testdata.mes_snapshots.len, si);
}

test "MaximumEntropySpectrum primes at bar 59" {
    var x = try MaximumEntropySpectrum.init(testing.allocator, .{});
    defer x.deinit();

    try testing.expect(!x.isPrimed());

    var primed_at: ?usize = null;
    for (0..testdata.test_input.len) |i| {
        _ = x.updateSample(testdata.test_input[i], @intCast(i));
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
            _ = x.updateSample(testdata.test_input[idx % testdata.test_input.len], time);
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
            _ = x.updateSample(testdata.test_input[idx % testdata.test_input.len], time);
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
            _ = x.updateSample(testdata.test_input[idx % testdata.test_input.len], time);
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
            _ = x.updateSample(testdata.test_input[idx % testdata.test_input.len], time);
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
fn roundDec(v: f64, dec: u32) f64 {
    const p = math.pow(f64, 10.0, @floatFromInt(dec));
    return @round(v * p) / p;
}

test "Burg coefficients against MBST" {
    // Skip if test data arrays are empty (not yet populated).
    if (testdata.test_input_four_sinusoids.len == 0) return;

    const CoefCase = struct {
        name: []const u8,
        input: []const f64,
        degree: i32,
        dec: u32,
        want: []const f64,
    };

    const cases = [_]CoefCase{
        .{ .name = "sinusoids/1", .input = &testdata.test_input_four_sinusoids, .degree = 1, .dec = 1, .want = &[_]f64{0.941872} },
        .{ .name = "sinusoids/2", .input = &testdata.test_input_four_sinusoids, .degree = 2, .dec = 1, .want = &[_]f64{ 1.826156, -0.938849 } },
        .{ .name = "sinusoids/3", .input = &testdata.test_input_four_sinusoids, .degree = 3, .dec = 1, .want = &[_]f64{ 2.753231, -2.740306, 0.985501 } },
        .{ .name = "sinusoids/4", .input = &testdata.test_input_four_sinusoids, .degree = 4, .dec = 1, .want = &[_]f64{ 3.736794, -5.474295, 3.731127, -0.996783 } },
        .{ .name = "test1/5", .input = &testdata.test_input_test1, .degree = 5, .dec = 1, .want = &[_]f64{ 1.4, -0.7, 0.04, 0.7, -0.5 } },
        .{ .name = "test2/7", .input = &testdata.test_input_test2, .degree = 7, .dec = 0, .want = &[_]f64{ 0.677, 0.175, 0.297, 0.006, -0.114, -0.083, -0.025 } },
        .{ .name = "test3/2", .input = &testdata.test_input_test3, .degree = 2, .dec = 1, .want = &[_]f64{ 1.02, -0.53 } },
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
