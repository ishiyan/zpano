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

/// Enumerates the outputs of the autocorrelation indicator.
pub const AutoCorrelationIndicatorOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an AutoCorrelationIndicator.
pub const Params = struct {
    /// Minimum correlation lag. Must be >= 1. Default is 3.
    min_lag: i32 = 3,
    /// Maximum correlation lag. Must be > min_lag. Default is 48.
    max_lag: i32 = 48,
    /// SuperSmoother cutoff period. Must be >= 2. Default is 10.
    smoothing_period: i32 = 10,
    /// Pearson averaging length. When 0, M = lag. Default is 0.
    averaging_length: i32 = 0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehlers autocorrelation indicator estimator.
const Estimator = struct {
    allocator: std.mem.Allocator,
    min_lag: usize,
    max_lag: usize,
    averaging_length: usize,
    length_spectrum: usize,
    filt_buffer_len: usize,

    // Pre-filter coefficients.
    coeff_hp0: f64,
    coeff_hp1: f64,
    coeff_hp2: f64,
    ss_c1: f64,
    ss_c2: f64,
    ss_c3: f64,

    // Pre-filter state.
    close0: f64,
    close1: f64,
    close2: f64,
    hp0: f64,
    hp1: f64,
    hp2: f64,

    // Filt history. filt[k] = Filt k bars ago (0 = current).
    filt: []f64,

    // Spectrum values indexed [0..lengthSpectrum), already scaled to [0,1].
    spectrum: []f64,

    spectrum_min: f64,
    spectrum_max: f64,

    fn init(
        allocator: std.mem.Allocator,
        min_lag: usize,
        max_lag: usize,
        smoothing_period: usize,
        averaging_length: usize,
    ) !Estimator {
        const two_pi = 2.0 * math.pi;

        const length_spectrum = max_lag - min_lag + 1;

        var m_max = averaging_length;
        if (averaging_length == 0) {
            m_max = max_lag;
        }

        const filt_buffer_len = max_lag + m_max;

        // Highpass coefficients, cutoff at MaxLag.
        const omega_hp = 0.707 * two_pi / @as(f64, @floatFromInt(max_lag));
        const alpha_hp = (@cos(omega_hp) + @sin(omega_hp) - 1.0) / @cos(omega_hp);
        const c_hp0 = (1.0 - alpha_hp / 2.0) * (1.0 - alpha_hp / 2.0);
        const c_hp1 = 2.0 * (1.0 - alpha_hp);
        const c_hp2 = (1.0 - alpha_hp) * (1.0 - alpha_hp);

        // SuperSmoother coefficients, period = SmoothingPeriod.
        const sp_f: f64 = @floatFromInt(smoothing_period);
        const a1 = @exp(-1.414 * math.pi / sp_f);
        const b1 = 2.0 * a1 * @cos(1.414 * math.pi / sp_f);
        const ss_c2 = b1;
        const ss_c3 = -a1 * a1;
        const ss_c1 = 1.0 - ss_c2 - ss_c3;

        const filt_buf = try allocator.alloc(f64, filt_buffer_len);
        @memset(filt_buf, 0.0);
        const spectrum_buf = try allocator.alloc(f64, length_spectrum);
        @memset(spectrum_buf, 0.0);

        return .{
            .allocator = allocator,
            .min_lag = min_lag,
            .max_lag = max_lag,
            .averaging_length = averaging_length,
            .length_spectrum = length_spectrum,
            .filt_buffer_len = filt_buffer_len,
            .coeff_hp0 = c_hp0,
            .coeff_hp1 = c_hp1,
            .coeff_hp2 = c_hp2,
            .ss_c1 = ss_c1,
            .ss_c2 = ss_c2,
            .ss_c3 = ss_c3,
            .close0 = 0.0,
            .close1 = 0.0,
            .close2 = 0.0,
            .hp0 = 0.0,
            .hp1 = 0.0,
            .hp2 = 0.0,
            .filt = filt_buf,
            .spectrum = spectrum_buf,
            .spectrum_min = 0.0,
            .spectrum_max = 0.0,
        };
    }

    fn deinit(self: *Estimator) void {
        self.allocator.free(self.filt);
        self.allocator.free(self.spectrum);
    }

    fn update(self: *Estimator, sample: f64) void {
        // Shift close history.
        self.close2 = self.close1;
        self.close1 = self.close0;
        self.close0 = sample;

        // Shift HP history and compute new HP.
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

        // Compute new Filt (SuperSmoother on HP).
        self.filt[0] = self.ss_c1 * (self.hp0 + self.hp1) / 2.0 + self.ss_c2 * self.filt[1] + self.ss_c3 * self.filt[2];

        // Pearson correlation per lag.
        self.spectrum_min = math.floatMax(f64);
        self.spectrum_max = -math.floatMax(f64);

        for (0..self.length_spectrum) |i| {
            const lag = self.min_lag + i;

            var m = self.averaging_length;
            if (m == 0) {
                m = lag;
            }

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

            const v = 0.5 * (r + 1.0);
            self.spectrum[i] = v;

            if (v < self.spectrum_min) {
                self.spectrum_min = v;
            }
            if (v > self.spectrum_max) {
                self.spectrum_max = v;
            }
        }
    }
};

/// Ehlers' Autocorrelation Indicator heatmap.
pub const AutoCorrelationIndicator = struct {
    allocator: std.mem.Allocator,
    estimator: Estimator,
    window_count: usize,
    prime_count: usize,
    primed: bool,
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
        InvalidMinLag,
        InvalidMaxLag,
        InvalidSmoothingPeriod,
        InvalidAveragingLength,
        MnemonicTooLong,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!AutoCorrelationIndicator {
        if (params.min_lag < 1) return error.InvalidMinLag;
        if (params.max_lag <= params.min_lag) return error.InvalidMaxLag;
        if (params.smoothing_period < 2) return error.InvalidSmoothingPeriod;
        if (params.averaging_length < 0) return error.InvalidAveragingLength;

        const bc = params.bar_component orelse bar_component.BarComponent.median;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        const min_lag: usize = @intCast(params.min_lag);
        const max_lag: usize = @intCast(params.max_lag);
        const smoothing_period: usize = @intCast(params.smoothing_period);
        const averaging_length: usize = @intCast(params.averaging_length);

        var estimator = Estimator.init(
            allocator,
            min_lag,
            max_lag,
            smoothing_period,
            averaging_length,
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

        if (params.averaging_length != 0) {
            const tag = std.fmt.bufPrint(flags_buf[flags_len..], ", average={d}", .{params.averaging_length}) catch
                return error.MnemonicTooLong;
            flags_len += tag.len;
        }

        const flags = flags_buf[0..flags_len];

        var mnemonic_buf: [256]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "aci({d}, {d}, {d}{s}{s})", .{
            params.min_lag,
            params.max_lag,
            params.smoothing_period,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Autocorrelation indicator {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        return .{
            .allocator = allocator,
            .estimator = estimator,
            .window_count = 0,
            .prime_count = estimator.filt_buffer_len,
            .primed = false,
            .min_parameter_value = @floatFromInt(params.min_lag),
            .max_parameter_value = @floatFromInt(params.max_lag),
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

    pub fn deinit(self: *AutoCorrelationIndicator) void {
        self.estimator.deinit();
    }

    pub fn fixSlices(self: *AutoCorrelationIndicator) void {
        _ = self;
    }

    fn mnemonic(self: *const AutoCorrelationIndicator) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const AutoCorrelationIndicator) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    /// Update with a new sample value and return the heatmap column.
    pub fn update(self: *AutoCorrelationIndicator, sample: f64, time: i64) Heatmap {
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

        // Spectrum is already in axis order (bin 0 = MinLag) and scaled to [0,1].
        // No additional normalization.
        var values: [heatmap_mod.max_heatmap_values]f64 = undefined;
        var value_min: f64 = math.inf(f64);
        var value_max: f64 = -math.inf(f64);

        for (0..length_spectrum) |i| {
            const v = self.estimator.spectrum[i];
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

    pub fn updateBar(self: *AutoCorrelationIndicator, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *AutoCorrelationIndicator, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *AutoCorrelationIndicator, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    pub fn updateScalar(self: *AutoCorrelationIndicator, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    fn updateEntity(self: *AutoCorrelationIndicator, time: i64, sample: f64) OutputArray {
        const h = self.update(sample, time);
        var out = OutputArray{};
        out.append(.{ .heatmap = h });
        return out;
    }

    pub fn isPrimed(self: *const AutoCorrelationIndicator) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const AutoCorrelationIndicator, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },

        };
        build_metadata_mod.buildMetadata(out, .auto_correlation_indicator, mn, desc, &texts);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *AutoCorrelationIndicator) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(AutoCorrelationIndicator);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

test "AutoCorrelationIndicator update" {
    const tolerance = 1e-12;
    const min_max_tol = 1e-10;

    var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
    defer x.deinit();

    var si: usize = 0;

    for (0..testdata.test_input.len) |i| {
        const h = x.update(testdata.test_input[i], @intCast(i));

        try testing.expectEqual(@as(f64, 3.0), h.parameter_first);
        try testing.expectEqual(@as(f64, 48.0), h.parameter_last);
        try testing.expectEqual(@as(f64, 1.0), h.parameter_resolution);

        if (!x.primed) {
            try testing.expect(h.isEmpty());
            continue;
        }

        try testing.expectEqual(@as(usize, 46), h.values_len);

        if (si < testdata.aci_snapshots.len and testdata.aci_snapshots[si].i == i) {
            const snap = testdata.aci_snapshots[si];
            try testing.expect(almostEqual(h.value_min, snap.value_min, min_max_tol));
            try testing.expect(almostEqual(h.value_max, snap.value_max, min_max_tol));

            const vals = h.valuesSlice();
            for (snap.spots) |sp| {
                try testing.expect(almostEqual(vals[sp.i], sp.v, tolerance));
            }

            si += 1;
        }
    }

    try testing.expectEqual(testdata.aci_snapshots.len, si);
}

test "AutoCorrelationIndicator primes at bar 95" {
    // primeCount = filtBufferLen = maxLag + max(averagingLength, maxLag) = 48 + 48 = 96
    // Uses >= comparison, so primes when windowCount >= 96, i.e. at bar index 95.
    var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
    defer x.deinit();

    try testing.expect(!x.isPrimed());

    var primed_at: ?usize = null;

    for (0..testdata.test_input.len) |i| {
        _ = x.update(testdata.test_input[i], @intCast(i));
        if (x.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expectEqual(@as(usize, 95), primed_at.?);
}

test "AutoCorrelationIndicator NaN input" {
    var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
    defer x.deinit();

    const h = x.update(math.nan(f64), 0);
    try testing.expect(h.isEmpty());
    try testing.expect(!x.isPrimed());
}

test "AutoCorrelationIndicator synthetic sine" {
    const period = 35.0;
    const bars = 600;

    var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
    defer x.deinit();

    var last: Heatmap = undefined;

    for (0..bars) |i| {
        const sample = 100.0 + @sin(2.0 * math.pi * @as(f64, @floatFromInt(i)) / period);
        last = x.update(sample, @intCast(i));
    }

    try testing.expect(!last.isEmpty());

    // Peak bin should correspond to lag=35. Bin k = lag - MinLag = 35 - 3 = 32.
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

test "AutoCorrelationIndicator metadata" {
    var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
    defer x.deinit();

    var md: Metadata = undefined;
    x.getMetadata(&md);

    const mn = "aci(3, 48, 10, hl/2)";
    try testing.expectEqualStrings(mn, x.mnemonic());
    try testing.expectEqual(Identifier.auto_correlation_indicator, md.identifier);
    try testing.expectEqualStrings(mn, md.mnemonic);
    try testing.expectEqual(@as(usize, 1), md.outputs_len);
}

test "AutoCorrelationIndicator mnemonic flags" {
    const TestCase = struct {
        params: Params,
        expected: []const u8,
    };

    const cases = [_]TestCase{
        .{ .params = .{}, .expected = "aci(3, 48, 10, hl/2)" },
        .{ .params = .{ .averaging_length = 5 }, .expected = "aci(3, 48, 10, average=5, hl/2)" },
        .{ .params = .{ .min_lag = 5, .max_lag = 30, .smoothing_period = 8 }, .expected = "aci(5, 30, 8, hl/2)" },
    };

    for (cases) |tc| {
        var x = try AutoCorrelationIndicator.init(testing.allocator, tc.params);
        defer x.deinit();
        try testing.expectEqualStrings(tc.expected, x.mnemonic());
    }
}

test "AutoCorrelationIndicator validation" {
    try testing.expectError(error.InvalidMinLag, AutoCorrelationIndicator.init(testing.allocator, .{ .min_lag = -1 }));
    try testing.expectError(error.InvalidMaxLag, AutoCorrelationIndicator.init(testing.allocator, .{ .min_lag = 10, .max_lag = 10 }));
    try testing.expectError(error.InvalidSmoothingPeriod, AutoCorrelationIndicator.init(testing.allocator, .{ .smoothing_period = 1 }));
    try testing.expectError(error.InvalidAveragingLength, AutoCorrelationIndicator.init(testing.allocator, .{ .averaging_length = -1 }));
}

test "AutoCorrelationIndicator updateEntity" {
    const prime_count = 200;
    const inp: f64 = 100.0;
    const time: i64 = 0;

    // Update scalar
    {
        var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
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
        var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
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
        var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
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
        var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |i| {
            _ = x.update(testdata.test_input[i % testdata.test_input.len], time);
        }
        const t = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = x.updateTrade(&t);
        try testing.expectEqual(@as(usize, 1), out.len);
    }
}
