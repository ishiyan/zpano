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

const AciSpot = struct {
    i: usize,
    v: f64,
};

const AciSnap = struct {
    i: usize,
    value_min: f64,
    value_max: f64,
    spots: []const AciSpot,
};

const aci_snapshots = [_]AciSnap{
    .{
        .i = 120,
        .value_min = 0.001836240969535,
        .value_max = 0.978182738969974,
        .spots = &[_]AciSpot{
            .{ .i = 0, .v = 0.001836240969535 },
            .{ .i = 9, .v = 0.689811238830704 },
            .{ .i = 19, .v = 0.129262528620445 },
            .{ .i = 28, .v = 0.548114630718731 },
            .{ .i = 44, .v = 0.403543031074008 },
        },
    },
    .{
        .i = 150,
        .value_min = 0.022833774167889,
        .value_max = 0.921882996195786,
        .spots = &[_]AciSpot{
            .{ .i = 0, .v = 0.022833774167889 },
            .{ .i = 9, .v = 0.162612446752813 },
            .{ .i = 19, .v = 0.101575974264094 },
            .{ .i = 28, .v = 0.517980565603365 },
            .{ .i = 44, .v = 0.619542434241698 },
        },
    },
    .{
        .i = 200,
        .value_min = 0.003686200641809,
        .value_max = 0.938289513902131,
        .spots = &[_]AciSpot{
            .{ .i = 0, .v = 0.013541709048057 },
            .{ .i = 9, .v = 0.673798419631138 },
            .{ .i = 19, .v = 0.485813016278695 },
            .{ .i = 28, .v = 0.060602734575409 },
            .{ .i = 44, .v = 0.858273375222992 },
        },
    },
    .{
        .i = 250,
        .value_min = 0.005505877822992,
        .value_max = 0.997937273618358,
        .spots = &[_]AciSpot{
            .{ .i = 0, .v = 0.997937273618358 },
            .{ .i = 9, .v = 0.833419934062773 },
            .{ .i = 19, .v = 0.135869513044420 },
            .{ .i = 28, .v = 0.057916227928612 },
            .{ .i = 44, .v = 0.488503533472072 },
        },
    },
};

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

test "AutoCorrelationIndicator update" {
    const tolerance = 1e-12;
    const min_max_tol = 1e-10;

    var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
    defer x.deinit();

    var si: usize = 0;

    for (0..test_input.len) |i| {
        const h = x.update(test_input[i], @intCast(i));

        try testing.expectEqual(@as(f64, 3.0), h.parameter_first);
        try testing.expectEqual(@as(f64, 48.0), h.parameter_last);
        try testing.expectEqual(@as(f64, 1.0), h.parameter_resolution);

        if (!x.primed) {
            try testing.expect(h.isEmpty());
            continue;
        }

        try testing.expectEqual(@as(usize, 46), h.values_len);

        if (si < aci_snapshots.len and aci_snapshots[si].i == i) {
            const snap = aci_snapshots[si];
            try testing.expect(almostEqual(h.value_min, snap.value_min, min_max_tol));
            try testing.expect(almostEqual(h.value_max, snap.value_max, min_max_tol));

            const vals = h.valuesSlice();
            for (snap.spots) |sp| {
                try testing.expect(almostEqual(vals[sp.i], sp.v, tolerance));
            }

            si += 1;
        }
    }

    try testing.expectEqual(aci_snapshots.len, si);
}

test "AutoCorrelationIndicator primes at bar 95" {
    // primeCount = filtBufferLen = maxLag + max(averagingLength, maxLag) = 48 + 48 = 96
    // Uses >= comparison, so primes when windowCount >= 96, i.e. at bar index 95.
    var x = try AutoCorrelationIndicator.init(testing.allocator, .{});
    defer x.deinit();

    try testing.expect(!x.isPrimed());

    var primed_at: ?usize = null;

    for (0..test_input.len) |i| {
        _ = x.update(test_input[i], @intCast(i));
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
            _ = x.update(test_input[i % test_input.len], time);
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
            _ = x.update(test_input[i % test_input.len], time);
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
            _ = x.update(test_input[i % test_input.len], time);
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
            _ = x.update(test_input[i % test_input.len], time);
        }
        const t = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = x.updateTrade(&t);
        try testing.expectEqual(@as(usize, 1), out.len);
    }
}
