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

/// Enumerates the outputs of the autocorrelation periodogram.
pub const AutoCorrelationPeriodogramOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an AutoCorrelationPeriodogram.
pub const Params = struct {
    /// Minimum cycle period. Must be >= 2. Default is 10.
    min_period: i32 = 10,
    /// Maximum cycle period. Must be > min_period. Default is 48.
    max_period: i32 = 48,
    /// Fixed averaging length for Pearson correlation. Must be >= 1. Default is 3.
    averaging_length: i32 = 3,
    /// Disable spectral squaring (SqSum^2 -> SqSum).
    disable_spectral_squaring: bool = false,
    /// Disable per-bin exponential smoothing.
    disable_smoothing: bool = false,
    /// Disable automatic gain control.
    disable_automatic_gain_control: bool = false,
    /// AGC decay factor in (0, 1). Default is 0.995.
    automatic_gain_control_decay_factor: f64 = 0.995,
    /// Use fixed normalization (min clamped to 0).
    fixed_normalization: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehlers autocorrelation periodogram estimator.
const Estimator = struct {
    allocator: std.mem.Allocator,
    min_period: usize,
    max_period: usize,
    averaging_length: usize,
    length_spectrum: usize,
    filt_buffer_len: usize,

    is_spectral_squaring: bool,
    is_smoothing: bool,
    is_automatic_gain_control: bool,
    automatic_gain_control_decay_factor: f64,

    // Pre-filter coefficients.
    coeff_hp0: f64,
    coeff_hp1: f64,
    coeff_hp2: f64,
    ss_c1: f64,
    ss_c2: f64,
    ss_c3: f64,

    // DFT basis tables: flattened [length_spectrum][corr_len].
    cos_tab: []f64,
    sin_tab: []f64,
    corr_len: usize,

    // Pre-filter state.
    close0: f64,
    close1: f64,
    close2: f64,
    hp0: f64,
    hp1: f64,
    hp2: f64,

    filt: []f64,
    corr: []f64,
    r_previous: []f64,
    spectrum: []f64,

    spectrum_min: f64,
    spectrum_max: f64,
    previous_spectrum_max: f64,

    fn init(
        allocator: std.mem.Allocator,
        min_period: usize,
        max_period: usize,
        averaging_length: usize,
        is_spectral_squaring: bool,
        is_smoothing: bool,
        is_agc: bool,
        agc_decay: f64,
    ) !Estimator {
        const two_pi = 2.0 * math.pi;
        const dft_lag_start: usize = 3;

        const length_spectrum = max_period - min_period + 1;
        const filt_buffer_len = max_period + averaging_length;
        const corr_len = max_period + 1;

        // Highpass coefficients, cutoff at MaxPeriod.
        const omega_hp = 0.707 * two_pi / @as(f64, @floatFromInt(max_period));
        const alpha_hp = (@cos(omega_hp) + @sin(omega_hp) - 1.0) / @cos(omega_hp);
        const c_hp0 = (1.0 - alpha_hp / 2.0) * (1.0 - alpha_hp / 2.0);
        const c_hp1 = 2.0 * (1.0 - alpha_hp);
        const c_hp2 = (1.0 - alpha_hp) * (1.0 - alpha_hp);

        // SuperSmoother coefficients, period = MinPeriod.
        const mp_f: f64 = @floatFromInt(min_period);
        const a1 = @exp(-1.414 * math.pi / mp_f);
        const b1 = 2.0 * a1 * @cos(1.414 * math.pi / mp_f);
        const ss_c2 = b1;
        const ss_c3 = -a1 * a1;
        const ss_c1 = 1.0 - ss_c2 - ss_c3;

        // DFT basis tables: flattened [length_spectrum][corr_len].
        const trig_size = length_spectrum * corr_len;
        const cos_buf = try allocator.alloc(f64, trig_size);
        @memset(cos_buf, 0.0);
        const sin_buf = try allocator.alloc(f64, trig_size);
        @memset(sin_buf, 0.0);

        for (0..length_spectrum) |i| {
            const period = min_period + i;
            const row_offset = i * corr_len;

            for (dft_lag_start..corr_len) |n| {
                const angle = two_pi * @as(f64, @floatFromInt(n)) / @as(f64, @floatFromInt(period));
                cos_buf[row_offset + n] = @cos(angle);
                sin_buf[row_offset + n] = @sin(angle);
            }
        }

        const filt_buf = try allocator.alloc(f64, filt_buffer_len);
        @memset(filt_buf, 0.0);
        const corr_buf = try allocator.alloc(f64, corr_len);
        @memset(corr_buf, 0.0);
        const r_prev_buf = try allocator.alloc(f64, length_spectrum);
        @memset(r_prev_buf, 0.0);
        const spectrum_buf = try allocator.alloc(f64, length_spectrum);
        @memset(spectrum_buf, 0.0);

        return .{
            .allocator = allocator,
            .min_period = min_period,
            .max_period = max_period,
            .averaging_length = averaging_length,
            .length_spectrum = length_spectrum,
            .filt_buffer_len = filt_buffer_len,
            .is_spectral_squaring = is_spectral_squaring,
            .is_smoothing = is_smoothing,
            .is_automatic_gain_control = is_agc,
            .automatic_gain_control_decay_factor = agc_decay,
            .coeff_hp0 = c_hp0,
            .coeff_hp1 = c_hp1,
            .coeff_hp2 = c_hp2,
            .ss_c1 = ss_c1,
            .ss_c2 = ss_c2,
            .ss_c3 = ss_c3,
            .cos_tab = cos_buf,
            .sin_tab = sin_buf,
            .corr_len = corr_len,
            .close0 = 0.0,
            .close1 = 0.0,
            .close2 = 0.0,
            .hp0 = 0.0,
            .hp1 = 0.0,
            .hp2 = 0.0,
            .filt = filt_buf,
            .corr = corr_buf,
            .r_previous = r_prev_buf,
            .spectrum = spectrum_buf,
            .spectrum_min = 0.0,
            .spectrum_max = 0.0,
            .previous_spectrum_max = 0.0,
        };
    }

    fn deinit(self: *Estimator) void {
        self.allocator.free(self.cos_tab);
        self.allocator.free(self.sin_tab);
        self.allocator.free(self.filt);
        self.allocator.free(self.corr);
        self.allocator.free(self.r_previous);
        self.allocator.free(self.spectrum);
    }

    fn update(self: *Estimator, sample: f64) void {
        const dft_lag_start: usize = 3;

        // Pre-filter cascade.
        self.close2 = self.close1;
        self.close1 = self.close0;
        self.close0 = sample;

        self.hp2 = self.hp1;
        self.hp1 = self.hp0;
        self.hp0 = self.coeff_hp0 * (self.close0 - 2.0 * self.close1 + self.close2) +
            self.coeff_hp1 * self.hp1 -
            self.coeff_hp2 * self.hp2;

        // Shift Filt history rightward.
        var k: usize = self.filt_buffer_len - 1;
        while (k >= 1) : (k -= 1) {
            self.filt[k] = self.filt[k - 1];
        }

        self.filt[0] = self.ss_c1 * (self.hp0 + self.hp1) / 2.0 + self.ss_c2 * self.filt[1] + self.ss_c3 * self.filt[2];

        // Pearson correlation per lag [0..maxPeriod], fixed M = averagingLength.
        const m = self.averaging_length;

        for (0..self.corr_len) |lag| {
            var sx: f64 = 0.0;
            var sy: f64 = 0.0;
            var sxx: f64 = 0.0;
            var syy: f64 = 0.0;
            var sxy: f64 = 0.0;

            for (0..m) |c| {
                const x = self.filt[c];
                const y = self.filt[lag + c];
                sx += x;
                sy += y;
                sxx += x * x;
                syy += y * y;
                sxy += x * y;
            }

            const mf: f64 = @floatFromInt(m);
            const denom = (mf * sxx - sx * sx) * (mf * syy - sy * sy);

            var r: f64 = 0.0;
            if (denom > 0.0) {
                r = (mf * sxy - sx * sy) / @sqrt(denom);
            }

            self.corr[lag] = r;
        }

        // DFT of correlation, smooth, AGC normalize.
        self.spectrum_min = math.floatMax(f64);
        if (self.is_automatic_gain_control) {
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max;
        } else {
            self.spectrum_max = -math.floatMax(f64);
        }

        // Pass 1: compute raw R values and track running max.
        for (0..self.length_spectrum) |i| {
            const row_offset = i * self.corr_len;

            var cos_part: f64 = 0.0;
            var sin_part: f64 = 0.0;

            for (dft_lag_start..self.corr_len) |n| {
                cos_part += self.corr[n] * self.cos_tab[row_offset + n];
                sin_part += self.corr[n] * self.sin_tab[row_offset + n];
            }

            const sq_sum = cos_part * cos_part + sin_part * sin_part;

            var raw = sq_sum;
            if (self.is_spectral_squaring) {
                raw = sq_sum * sq_sum;
            }

            var r: f64 = undefined;
            if (self.is_smoothing) {
                r = 0.2 * raw + 0.8 * self.r_previous[i];
            } else {
                r = raw;
            }

            self.r_previous[i] = r;
            self.spectrum[i] = r;

            if (self.spectrum_max < r) {
                self.spectrum_max = r;
            }
        }

        self.previous_spectrum_max = self.spectrum_max;

        // Pass 2: normalize against running max and track normalized min.
        if (self.spectrum_max > 0.0) {
            for (0..self.length_spectrum) |i| {
                const v = self.spectrum[i] / self.spectrum_max;
                self.spectrum[i] = v;

                if (self.spectrum_min > v) {
                    self.spectrum_min = v;
                }
            }
        } else {
            for (0..self.length_spectrum) |i| {
                self.spectrum[i] = 0.0;
            }
            self.spectrum_min = 0.0;
        }
    }
};

/// Ehlers' Autocorrelation Periodogram heatmap indicator.
pub const AutoCorrelationPeriodogram = struct {
    allocator: std.mem.Allocator,
    estimator: Estimator,
    window_count: usize,
    prime_count: usize,
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
        InvalidMinPeriod,
        InvalidMaxPeriod,
        InvalidAveragingLength,
        InvalidAgcDecay,
        MnemonicTooLong,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!AutoCorrelationPeriodogram {
        const agc_decay_epsilon: f64 = 1e-12;
        const def_agc_decay: f64 = 0.995;
        const def_averaging_len: i32 = 3;

        if (params.min_period < 2) return error.InvalidMinPeriod;
        if (params.max_period <= params.min_period) return error.InvalidMaxPeriod;
        if (params.averaging_length < 1) return error.InvalidAveragingLength;

        const agc_on = !params.disable_automatic_gain_control;
        if (agc_on and (params.automatic_gain_control_decay_factor <= 0.0 or params.automatic_gain_control_decay_factor >= 1.0)) {
            return error.InvalidAgcDecay;
        }

        const squaring_on = !params.disable_spectral_squaring;
        const smoothing_on = !params.disable_smoothing;
        const floating_norm = !params.fixed_normalization;

        const bc = params.bar_component orelse bar_component.BarComponent.median;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        const min_period: usize = @intCast(params.min_period);
        const max_period: usize = @intCast(params.max_period);
        const averaging_length: usize = @intCast(params.averaging_length);

        var estimator = Estimator.init(
            allocator,
            min_period,
            max_period,
            averaging_length,
            squaring_on,
            smoothing_on,
            agc_on,
            params.automatic_gain_control_decay_factor,
        ) catch return error.OutOfMemory;
        errdefer estimator.deinit();

        // Build mnemonic.
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

        if (params.averaging_length != def_averaging_len) {
            const tag = std.fmt.bufPrint(flags_buf[flags_len..], ", average={d}", .{params.averaging_length}) catch
                return error.MnemonicTooLong;
            flags_len += tag.len;
        }

        if (!squaring_on) {
            const tag = ", no-sqr";
            @memcpy(flags_buf[flags_len .. flags_len + tag.len], tag);
            flags_len += tag.len;
        }

        if (!smoothing_on) {
            const tag = ", no-smooth";
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
        const mn = std.fmt.bufPrint(&mnemonic_buf, "acp({d}, {d}{s}{s})", .{
            params.min_period,
            params.max_period,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Autocorrelation periodogram {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        return .{
            .allocator = allocator,
            .estimator = estimator,
            .window_count = 0,
            .prime_count = estimator.filt_buffer_len,
            .primed = false,
            .floating_normalization = floating_norm,
            .min_parameter_value = @floatFromInt(params.min_period),
            .max_parameter_value = @floatFromInt(params.max_period),
            .parameter_resolution = 1.0,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *AutoCorrelationPeriodogram) void {
        self.estimator.deinit();
    }

    pub fn fixSlices(self: *AutoCorrelationPeriodogram) void {
        _ = self;
    }

    fn mnemonic(self: *const AutoCorrelationPeriodogram) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const AutoCorrelationPeriodogram) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    /// Update with a new sample value and return the heatmap column.
    pub fn update(self: *AutoCorrelationPeriodogram, sample: f64, time: i64) Heatmap {
        if (math.isNan(sample)) {
            return Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution);
        }

        self.estimator.update(sample);

        if (!self.primed) {
            self.window_count += 1;

            if (self.window_count >= self.prime_count) {
                self.primed = true;
            } else {
                return Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution);
            }
        }

        const length_spectrum = self.estimator.length_spectrum;

        var min_ref: f64 = 0.0;
        if (self.floating_normalization) {
            min_ref = self.estimator.spectrum_min;
        }

        // Estimator spectrum is already AGC-normalized in [0, 1].
        const max_ref: f64 = 1.0;
        const spectrum_range = max_ref - min_ref;

        var values: [heatmap_mod.max_heatmap_values]f64 = undefined;
        var value_min: f64 = math.inf(f64);
        var value_max: f64 = -math.inf(f64);

        for (0..length_spectrum) |i| {
            var v: f64 = 0.0;
            if (spectrum_range > 0.0) {
                v = (self.estimator.spectrum[i] - min_ref) / spectrum_range;
            }

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

    pub fn updateBar(self: *AutoCorrelationPeriodogram, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *AutoCorrelationPeriodogram, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *AutoCorrelationPeriodogram, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    pub fn updateScalar(self: *AutoCorrelationPeriodogram, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    fn updateEntity(self: *AutoCorrelationPeriodogram, time: i64, sample: f64) OutputArray {
        const h = self.update(sample, time);
        var out = OutputArray{};
        out.append(.{ .heatmap = h });
        return out;
    }

    pub fn isPrimed(self: *const AutoCorrelationPeriodogram) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const AutoCorrelationPeriodogram, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },

        };
        build_metadata_mod.buildMetadata(out, .auto_correlation_periodogram, mn, desc, &texts);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *AutoCorrelationPeriodogram) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(AutoCorrelationPeriodogram);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

test "AutoCorrelationPeriodogram update" {
    const tolerance = 1e-12;
    const min_max_tol = 1e-10;

    var x = try AutoCorrelationPeriodogram.init(testing.allocator, .{});
    defer x.deinit();

    var si: usize = 0;

    for (0..testdata.test_input.len) |i| {
        const h = x.update(testdata.test_input[i], @intCast(i));

        try testing.expectEqual(@as(f64, 10.0), h.parameter_first);
        try testing.expectEqual(@as(f64, 48.0), h.parameter_last);
        try testing.expectEqual(@as(f64, 1.0), h.parameter_resolution);

        if (!x.primed) {
            try testing.expect(h.isEmpty());
            continue;
        }

        try testing.expectEqual(@as(usize, 39), h.values_len);

        if (si < testdata.acp_snapshots.len and testdata.acp_snapshots[si].i == i) {
            const snap = testdata.acp_snapshots[si];
            try testing.expect(almostEqual(h.value_min, snap.value_min, min_max_tol));
            try testing.expect(almostEqual(h.value_max, snap.value_max, min_max_tol));

            const vals = h.valuesSlice();
            for (snap.spots) |sp| {
                try testing.expect(almostEqual(vals[sp.i], sp.v, tolerance));
            }

            si += 1;
        }
    }

    try testing.expectEqual(testdata.acp_snapshots.len, si);
}

test "AutoCorrelationPeriodogram primes at bar 50" {
    // primeCount = filtBufferLen = maxPeriod + averagingLength = 48 + 3 = 51
    // Primes when windowCount >= 51, i.e. at bar index 50.
    var x = try AutoCorrelationPeriodogram.init(testing.allocator, .{});
    defer x.deinit();

    try testing.expect(!x.isPrimed());

    var primed_at: ?usize = null;

    for (0..testdata.test_input.len) |i| {
        _ = x.update(testdata.test_input[i], @intCast(i));
        if (x.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expectEqual(@as(usize, 50), primed_at.?);
}

test "AutoCorrelationPeriodogram NaN input" {
    var x = try AutoCorrelationPeriodogram.init(testing.allocator, .{});
    defer x.deinit();

    const h = x.update(math.nan(f64), 0);
    try testing.expect(h.isEmpty());
    try testing.expect(!x.isPrimed());
}

test "AutoCorrelationPeriodogram synthetic sine" {
    const period = 20.0;
    const bars = 600;

    var x = try AutoCorrelationPeriodogram.init(testing.allocator, .{
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

    var peak_bin: usize = 0;
    const vals = last.valuesSlice();
    for (1..vals.len) |i| {
        if (vals[i] > vals[peak_bin]) {
            peak_bin = i;
        }
    }

    // Bin k corresponds to period MinPeriod+k. MinPeriod=10, period=20 -> bin 10.
    const expected_bin: usize = @intFromFloat(period - last.parameter_first);
    try testing.expectEqual(expected_bin, peak_bin);
}

test "AutoCorrelationPeriodogram metadata" {
    var x = try AutoCorrelationPeriodogram.init(testing.allocator, .{});
    defer x.deinit();

    var md: Metadata = undefined;
    x.getMetadata(&md);

    const mn = "acp(10, 48, hl/2)";
    try testing.expectEqualStrings(mn, x.mnemonic());
    try testing.expectEqual(Identifier.auto_correlation_periodogram, md.identifier);
    try testing.expectEqualStrings(mn, md.mnemonic);
    try testing.expectEqual(@as(usize, 1), md.outputs_len);
}

test "AutoCorrelationPeriodogram mnemonic flags" {
    const TestCase = struct {
        params: Params,
        expected: []const u8,
    };

    const cases = [_]TestCase{
        .{ .params = .{}, .expected = "acp(10, 48, hl/2)" },
        .{ .params = .{ .averaging_length = 5 }, .expected = "acp(10, 48, average=5, hl/2)" },
        .{ .params = .{ .disable_spectral_squaring = true }, .expected = "acp(10, 48, no-sqr, hl/2)" },
        .{ .params = .{ .disable_smoothing = true }, .expected = "acp(10, 48, no-smooth, hl/2)" },
        .{ .params = .{ .disable_automatic_gain_control = true }, .expected = "acp(10, 48, no-agc, hl/2)" },
        .{ .params = .{ .automatic_gain_control_decay_factor = 0.8 }, .expected = "acp(10, 48, agc=0.8, hl/2)" },
        .{ .params = .{ .fixed_normalization = true }, .expected = "acp(10, 48, no-fn, hl/2)" },
        .{
            .params = .{
                .averaging_length = 5,
                .disable_spectral_squaring = true,
                .disable_smoothing = true,
                .disable_automatic_gain_control = true,
                .fixed_normalization = true,
            },
            .expected = "acp(10, 48, average=5, no-sqr, no-smooth, no-agc, no-fn, hl/2)",
        },
    };

    for (cases) |tc| {
        var x = try AutoCorrelationPeriodogram.init(testing.allocator, tc.params);
        defer x.deinit();
        try testing.expectEqualStrings(tc.expected, x.mnemonic());
    }
}

test "AutoCorrelationPeriodogram validation" {
    try testing.expectError(error.InvalidMinPeriod, AutoCorrelationPeriodogram.init(testing.allocator, .{ .min_period = 1 }));
    try testing.expectError(error.InvalidMaxPeriod, AutoCorrelationPeriodogram.init(testing.allocator, .{ .min_period = 10, .max_period = 10 }));
    try testing.expectError(error.InvalidAveragingLength, AutoCorrelationPeriodogram.init(testing.allocator, .{ .averaging_length = -1 }));
    try testing.expectError(error.InvalidAgcDecay, AutoCorrelationPeriodogram.init(testing.allocator, .{ .automatic_gain_control_decay_factor = -0.1 }));
    try testing.expectError(error.InvalidAgcDecay, AutoCorrelationPeriodogram.init(testing.allocator, .{ .automatic_gain_control_decay_factor = 1.0 }));
}

test "AutoCorrelationPeriodogram updateEntity" {
    const prime_count = 100;
    const inp: f64 = 100.0;
    const time: i64 = 0;

    // Update scalar
    {
        var x = try AutoCorrelationPeriodogram.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |i| {
            _ = x.update(testdata.test_input[i % testdata.test_input.len], time);
        }
        const s = Scalar{ .time = time, .value = inp };
        const out = x.updateScalar(&s);
        try testing.expectEqual(@as(usize, 1), out.len);
    }

    // Update bar
    {
        var x = try AutoCorrelationPeriodogram.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |i| {
            _ = x.update(testdata.test_input[i % testdata.test_input.len], time);
        }
        const b = Bar{ .time = time, .open = inp, .high = inp, .low = inp, .close = inp, .volume = 0 };
        const out = x.updateBar(&b);
        try testing.expectEqual(@as(usize, 1), out.len);
    }

    // Update quote
    {
        var x = try AutoCorrelationPeriodogram.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |i| {
            _ = x.update(testdata.test_input[i % testdata.test_input.len], time);
        }
        const q = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };
        const out = x.updateQuote(&q);
        try testing.expectEqual(@as(usize, 1), out.len);
    }

    // Update trade
    {
        var x = try AutoCorrelationPeriodogram.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |i| {
            _ = x.update(testdata.test_input[i % testdata.test_input.len], time);
        }
        const t = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = x.updateTrade(&t);
        try testing.expectEqual(@as(usize, 1), out.len);
    }
}
