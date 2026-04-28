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

/// Enumerates the outputs of the comb band-pass spectrum.
pub const CombBandPassSpectrumOutput = enum(u8) {
    value = 1,
};

/// Parameters to create a CombBandPassSpectrum.
pub const Params = struct {
    /// Minimum cycle period. Must be >= 2. Default is 10.
    min_period: i32 = 10,
    /// Maximum cycle period. Must be > min_period. Default is 48.
    max_period: i32 = 48,
    /// Fractional bandwidth of each band-pass filter. Must be in (0, 1). Default is 0.3.
    bandwidth: f64 = 0.3,
    /// Disable spectral dilation compensation.
    disable_spectral_dilation_compensation: bool = false,
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

/// Ehlers comb band-pass spectrum estimator (listing 10-1).
const Estimator = struct {
    allocator: std.mem.Allocator,
    min_period: usize,
    max_period: usize,
    length_spectrum: usize,
    is_spectral_dilation_compensation: bool,
    is_automatic_gain_control: bool,
    automatic_gain_control_decay_factor: f64,

    // Pre-filter coefficients.
    coeff_hp0: f64,
    coeff_hp1: f64,
    coeff_hp2: f64,
    ss_c1: f64,
    ss_c2: f64,
    ss_c3: f64,

    // Per-bin band-pass coefficients.
    periods: []usize,
    beta: []f64,
    alpha: []f64,
    comp: []f64,

    // Pre-filter state (scalar).
    close0: f64,
    close1: f64,
    close2: f64,
    hp0: f64,
    hp1: f64,
    hp2: f64,
    filt0: f64,
    filt1: f64,
    filt2: f64,

    // Band-pass state: flattened [length_spectrum][max_period].
    // bp[i * max_period + m] = BP output for bin i at lag m.
    bp: []f64,
    bp_stride: usize,

    spectrum: []f64,

    spectrum_min: f64,
    spectrum_max: f64,
    previous_spectrum_max: f64,

    fn init(
        allocator: std.mem.Allocator,
        min_period: usize,
        max_period: usize,
        bandwidth: f64,
        is_sdc: bool,
        is_agc: bool,
        agc_decay: f64,
    ) !Estimator {
        const two_pi = 2.0 * math.pi;

        const length_spectrum = max_period - min_period + 1;

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

        // Per-bin coefficients.
        const periods_buf = try allocator.alloc(usize, length_spectrum);
        const beta_buf = try allocator.alloc(f64, length_spectrum);
        const alpha_buf = try allocator.alloc(f64, length_spectrum);
        const comp_buf = try allocator.alloc(f64, length_spectrum);

        for (0..length_spectrum) |i| {
            const n = min_period + i;
            const nf: f64 = @floatFromInt(n);
            const b = @cos(two_pi / nf);
            const gamma = 1.0 / @cos(two_pi * bandwidth / nf);
            const a = gamma - @sqrt(gamma * gamma - 1.0);

            periods_buf[i] = n;
            beta_buf[i] = b;
            alpha_buf[i] = a;

            if (is_sdc) {
                comp_buf[i] = nf;
            } else {
                comp_buf[i] = 1.0;
            }
        }

        // BP state: flattened [length_spectrum][max_period].
        const bp_buf = try allocator.alloc(f64, length_spectrum * max_period);
        @memset(bp_buf, 0.0);

        const spectrum_buf = try allocator.alloc(f64, length_spectrum);
        @memset(spectrum_buf, 0.0);

        return .{
            .allocator = allocator,
            .min_period = min_period,
            .max_period = max_period,
            .length_spectrum = length_spectrum,
            .is_spectral_dilation_compensation = is_sdc,
            .is_automatic_gain_control = is_agc,
            .automatic_gain_control_decay_factor = agc_decay,
            .coeff_hp0 = c_hp0,
            .coeff_hp1 = c_hp1,
            .coeff_hp2 = c_hp2,
            .ss_c1 = ss_c1,
            .ss_c2 = ss_c2,
            .ss_c3 = ss_c3,
            .periods = periods_buf,
            .beta = beta_buf,
            .alpha = alpha_buf,
            .comp = comp_buf,
            .close0 = 0.0,
            .close1 = 0.0,
            .close2 = 0.0,
            .hp0 = 0.0,
            .hp1 = 0.0,
            .hp2 = 0.0,
            .filt0 = 0.0,
            .filt1 = 0.0,
            .filt2 = 0.0,
            .bp = bp_buf,
            .bp_stride = max_period,
            .spectrum = spectrum_buf,
            .spectrum_min = 0.0,
            .spectrum_max = 0.0,
            .previous_spectrum_max = 0.0,
        };
    }

    fn deinit(self: *Estimator) void {
        self.allocator.free(self.periods);
        self.allocator.free(self.beta);
        self.allocator.free(self.alpha);
        self.allocator.free(self.comp);
        self.allocator.free(self.bp);
        self.allocator.free(self.spectrum);
    }

    fn update(self: *Estimator, sample: f64) void {
        // Shift close history.
        self.close2 = self.close1;
        self.close1 = self.close0;
        self.close0 = sample;

        // HP filter.
        self.hp2 = self.hp1;
        self.hp1 = self.hp0;
        self.hp0 = self.coeff_hp0 * (self.close0 - 2.0 * self.close1 + self.close2) +
            self.coeff_hp1 * self.hp1 -
            self.coeff_hp2 * self.hp2;

        // SuperSmoother on HP (scalar state).
        self.filt2 = self.filt1;
        self.filt1 = self.filt0;
        self.filt0 = self.ss_c1 * (self.hp0 + self.hp1) / 2.0 + self.ss_c2 * self.filt1 + self.ss_c3 * self.filt2;

        const diff_filt = self.filt0 - self.filt2;

        // AGC seeds the running max.
        self.spectrum_min = math.floatMax(f64);
        if (self.is_automatic_gain_control) {
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max;
        } else {
            self.spectrum_max = -math.floatMax(f64);
        }

        for (0..self.length_spectrum) |i| {
            const row = self.bp[i * self.bp_stride .. (i + 1) * self.bp_stride];

            // Rightward shift: bp[m] = bp[m-1] for m from maxPeriod-1 down to 1.
            var m: usize = self.max_period - 1;
            while (m >= 1) : (m -= 1) {
                row[m] = row[m - 1];
            }

            const a = self.alpha[i];
            const b = self.beta[i];
            row[0] = 0.5 * (1.0 - a) * diff_filt + b * (1.0 + a) * row[1] - a * row[2];

            // Power = sum of (bp[m] / comp)^2 for m in [0..period).
            const n = self.periods[i];
            const c = self.comp[i];
            var pwr: f64 = 0.0;

            for (0..n) |j| {
                const v = row[j] / c;
                pwr += v * v;
            }

            self.spectrum[i] = pwr;

            if (self.spectrum_max < pwr) {
                self.spectrum_max = pwr;
            }
            if (self.spectrum_min > pwr) {
                self.spectrum_min = pwr;
            }
        }

        self.previous_spectrum_max = self.spectrum_max;
    }
};

/// Ehlers' Comb Band-Pass Spectrum heatmap indicator.
pub const CombBandPassSpectrum = struct {
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
        InvalidBandwidth,
        InvalidAgcDecay,
        MnemonicTooLong,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!CombBandPassSpectrum {
        const agc_decay_epsilon: f64 = 1e-12;
        const def_agc_decay: f64 = 0.995;
        const def_bandwidth: f64 = 0.3;
        const bandwidth_epsilon: f64 = 1e-12;

        if (params.min_period < 2) return error.InvalidMinPeriod;
        if (params.max_period <= params.min_period) return error.InvalidMaxPeriod;
        if (params.bandwidth <= 0.0 or params.bandwidth >= 1.0) return error.InvalidBandwidth;

        const agc_on = !params.disable_automatic_gain_control;
        if (agc_on and (params.automatic_gain_control_decay_factor <= 0.0 or params.automatic_gain_control_decay_factor >= 1.0)) {
            return error.InvalidAgcDecay;
        }

        const sdc_on = !params.disable_spectral_dilation_compensation;
        const floating_norm = !params.fixed_normalization;

        const bc = params.bar_component orelse bar_component.BarComponent.median;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        const min_period: usize = @intCast(params.min_period);
        const max_period: usize = @intCast(params.max_period);

        var estimator = Estimator.init(
            allocator,
            min_period,
            max_period,
            params.bandwidth,
            sdc_on,
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

        // Build flags string. Order: bw, no-sdc, no-agc, agc=X, no-fn.
        var flags_buf: [128]u8 = undefined;
        var flags_len: usize = 0;

        if (@abs(params.bandwidth - def_bandwidth) > bandwidth_epsilon) {
            const tag = std.fmt.bufPrint(flags_buf[flags_len..], ", bw={d}", .{params.bandwidth}) catch
                return error.MnemonicTooLong;
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
        const mn = std.fmt.bufPrint(&mnemonic_buf, "cbps({d}, {d}{s}{s})", .{
            params.min_period,
            params.max_period,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Comb band-pass spectrum {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        return .{
            .allocator = allocator,
            .estimator = estimator,
            .window_count = 0,
            .prime_count = max_period,
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

    pub fn deinit(self: *CombBandPassSpectrum) void {
        self.estimator.deinit();
    }

    pub fn fixSlices(self: *CombBandPassSpectrum) void {
        _ = self;
    }

    fn mnemonic(self: *const CombBandPassSpectrum) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const CombBandPassSpectrum) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    /// Update with a new sample value and return the heatmap column.
    pub fn update(self: *CombBandPassSpectrum, sample: f64, time: i64) Heatmap {
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

        const max_ref = self.estimator.spectrum_max;
        const spectrum_range = max_ref - min_ref;

        // Spectrum is already in axis order (bin 0 = MinPeriod).
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

    pub fn updateBar(self: *CombBandPassSpectrum, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *CombBandPassSpectrum, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *CombBandPassSpectrum, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    pub fn updateScalar(self: *CombBandPassSpectrum, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    fn updateEntity(self: *CombBandPassSpectrum, time: i64, sample: f64) OutputArray {
        const h = self.update(sample, time);
        var out = OutputArray{};
        out.append(.{ .heatmap = h });
        return out;
    }

    pub fn isPrimed(self: *const CombBandPassSpectrum) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const CombBandPassSpectrum, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },

        };
        build_metadata_mod.buildMetadata(out, .comb_band_pass_spectrum, mn, desc, &texts);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *CombBandPassSpectrum) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(CombBandPassSpectrum);
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

const CbpsSpot = struct {
    i: usize,
    v: f64,
};

const CbpsSnap = struct {
    i: usize,
    value_min: f64,
    value_max: f64,
    spots: []const CbpsSpot,
};

const cbps_snapshots = [_]CbpsSnap{
    .{
        .i = 47,
        .value_min = 0,
        .value_max = 0.351344643038070,
        .spots = &[_]CbpsSpot{
            .{ .i = 0, .v = 0.004676953354739 },
            .{ .i = 9, .v = 0.032804657174884 },
            .{ .i = 19, .v = 0.298241001617233 },
            .{ .i = 28, .v = 0.269179028265479 },
            .{ .i = 38, .v = 0.145584088643502 },
        },
    },
    .{
        .i = 60,
        .value_min = 0,
        .value_max = 0.233415131482019,
        .spots = &[_]CbpsSpot{
            .{ .i = 0, .v = 0.003611349016608 },
            .{ .i = 9, .v = 0.021460554913141 },
            .{ .i = 19, .v = 0.159313027547382 },
            .{ .i = 28, .v = 0.219799344776603 },
            .{ .i = 38, .v = 0.171081964194873 },
        },
    },
    .{
        .i = 100,
        .value_min = 0,
        .value_max = 0.064066532878879,
        .spots = &[_]CbpsSpot{
            .{ .i = 0, .v = 0.015789490651889 },
            .{ .i = 9, .v = 0.030957048077702 },
            .{ .i = 19, .v = 0.004154893462836 },
            .{ .i = 28, .v = 0.042739584630981 },
            .{ .i = 38, .v = 0.048070192646483 },
        },
    },
    .{
        .i = 150,
        .value_min = 0,
        .value_max = 0.044774991014571,
        .spots = &[_]CbpsSpot{
            .{ .i = 0, .v = 0.010977897375080 },
            .{ .i = 9, .v = 0.022161976000123 },
            .{ .i = 19, .v = 0.005434298746720 },
            .{ .i = 28, .v = 0.041109264147755 },
            .{ .i = 38, .v = 0.000028252306207 },
        },
    },
    .{
        .i = 200,
        .value_min = 0,
        .value_max = 0.056007975310479,
        .spots = &[_]CbpsSpot{
            .{ .i = 0, .v = 0.002054905622165 },
            .{ .i = 9, .v = 0.042579171063316 },
            .{ .i = 19, .v = 0.003278307476910 },
            .{ .i = 28, .v = 0.033557809407585 },
            .{ .i = 38, .v = 0.018072829155854 },
        },
    },
};

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

test "CombBandPassSpectrum update" {
    const tolerance = 1e-12;
    const min_max_tol = 1e-10;

    var x = try CombBandPassSpectrum.init(testing.allocator, .{});
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

        if (si < cbps_snapshots.len and cbps_snapshots[si].i == i) {
            const snap = cbps_snapshots[si];
            try testing.expect(almostEqual(h.value_min, snap.value_min, min_max_tol));
            try testing.expect(almostEqual(h.value_max, snap.value_max, min_max_tol));

            const vals = h.valuesSlice();
            for (snap.spots) |sp| {
                try testing.expect(almostEqual(vals[sp.i], sp.v, tolerance));
            }

            si += 1;
        }
    }

    try testing.expectEqual(cbps_snapshots.len, si);
}

test "CombBandPassSpectrum primes at bar 47" {
    var x = try CombBandPassSpectrum.init(testing.allocator, .{});
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

test "CombBandPassSpectrum NaN input" {
    var x = try CombBandPassSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    const h = x.update(math.nan(f64), 0);
    try testing.expect(h.isEmpty());
    try testing.expect(!x.isPrimed());
}

test "CombBandPassSpectrum synthetic sine" {
    const period = 20.0;
    const bars = 400;

    var x = try CombBandPassSpectrum.init(testing.allocator, .{
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

test "CombBandPassSpectrum metadata" {
    var x = try CombBandPassSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    var md: Metadata = undefined;
    x.getMetadata(&md);

    const mn = "cbps(10, 48, hl/2)";
    try testing.expectEqualStrings(mn, x.mnemonic());
    try testing.expectEqual(Identifier.comb_band_pass_spectrum, md.identifier);
    try testing.expectEqualStrings(mn, md.mnemonic);
    try testing.expectEqual(@as(usize, 1), md.outputs_len);
}

test "CombBandPassSpectrum mnemonic flags" {
    const TestCase = struct {
        params: Params,
        expected: []const u8,
    };

    const cases = [_]TestCase{
        .{ .params = .{}, .expected = "cbps(10, 48, hl/2)" },
        .{ .params = .{ .bandwidth = 0.5 }, .expected = "cbps(10, 48, bw=0.5, hl/2)" },
        .{ .params = .{ .disable_spectral_dilation_compensation = true }, .expected = "cbps(10, 48, no-sdc, hl/2)" },
        .{ .params = .{ .disable_automatic_gain_control = true }, .expected = "cbps(10, 48, no-agc, hl/2)" },
        .{ .params = .{ .automatic_gain_control_decay_factor = 0.8 }, .expected = "cbps(10, 48, agc=0.8, hl/2)" },
        .{ .params = .{ .fixed_normalization = true }, .expected = "cbps(10, 48, no-fn, hl/2)" },
        .{
            .params = .{
                .bandwidth = 0.5,
                .disable_spectral_dilation_compensation = true,
                .disable_automatic_gain_control = true,
                .fixed_normalization = true,
            },
            .expected = "cbps(10, 48, bw=0.5, no-sdc, no-agc, no-fn, hl/2)",
        },
    };

    for (cases) |tc| {
        var x = try CombBandPassSpectrum.init(testing.allocator, tc.params);
        defer x.deinit();
        try testing.expectEqualStrings(tc.expected, x.mnemonic());
    }
}

test "CombBandPassSpectrum validation" {
    try testing.expectError(error.InvalidMinPeriod, CombBandPassSpectrum.init(testing.allocator, .{ .min_period = 1 }));
    try testing.expectError(error.InvalidMaxPeriod, CombBandPassSpectrum.init(testing.allocator, .{ .min_period = 10, .max_period = 10 }));
    try testing.expectError(error.InvalidBandwidth, CombBandPassSpectrum.init(testing.allocator, .{ .bandwidth = -0.1 }));
    try testing.expectError(error.InvalidBandwidth, CombBandPassSpectrum.init(testing.allocator, .{ .bandwidth = 1.0 }));
    try testing.expectError(error.InvalidAgcDecay, CombBandPassSpectrum.init(testing.allocator, .{ .automatic_gain_control_decay_factor = -0.1 }));
    try testing.expectError(error.InvalidAgcDecay, CombBandPassSpectrum.init(testing.allocator, .{ .automatic_gain_control_decay_factor = 1.0 }));
}

test "CombBandPassSpectrum updateEntity" {
    const prime_count = 60;
    const inp: f64 = 100.0;
    const time: i64 = 0;

    // Update scalar
    {
        var x = try CombBandPassSpectrum.init(testing.allocator, .{});
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
        var x = try CombBandPassSpectrum.init(testing.allocator, .{});
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
        var x = try CombBandPassSpectrum.init(testing.allocator, .{});
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
        var x = try CombBandPassSpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |i| {
            _ = x.update(test_input[i % test_input.len], time);
        }
        const t = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = x.updateTrade(&t);
        try testing.expectEqual(@as(usize, 1), out.len);
    }
}
