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
const band_mod = @import("../../core/outputs/band.zig");

const ht = @import("../hilbert_transformer/hilbert_transformer.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const CycleEstimatorType = ht.CycleEstimatorType;
const CycleEstimatorParams = ht.CycleEstimatorParams;
const CycleEstimator = ht.CycleEstimator;

/// Enumerates the outputs of the indicator.
pub const MamaOutput = enum(u8) {
    value = 1,
    fama = 2,
    band = 3,
};

/// Parameters to create a MAMA indicator based on lengths.
pub const LengthParams = struct {
    estimator_type: CycleEstimatorType = .homodyne_discriminator,
    estimator_params: CycleEstimatorParams = .{},
    fast_limit_length: i32 = 3,
    slow_limit_length: i32 = 39,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create a MAMA indicator based on smoothing factors.
pub const SmoothingFactorParams = struct {
    estimator_type: CycleEstimatorType = .homodyne_discriminator,
    estimator_params: CycleEstimatorParams = .{},
    fast_limit_smoothing_factor: f64 = 0.5,
    slow_limit_smoothing_factor: f64 = 0.05,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehlers' Mesa Adaptive Moving Average (MAMA) indicator.
///
/// Three outputs:
///   - Value: the MAMA line
///   - Fama: the Following Adaptive Moving Average
///   - Band: upper=MAMA, lower=FAMA
pub const MesaAdaptiveMovingAverage = struct {
    alpha_fast_limit: f64,
    alpha_slow_limit: f64,
    previous_phase: f64,
    mama: f64,
    fama: f64,
    htce: CycleEstimator,
    is_phase_cached: bool,
    primed: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    // Mnemonic/description buffers for 3 outputs.
    mn_value_buf: [128]u8,
    mn_value_len: usize,
    mn_fama_buf: [128]u8,
    mn_fama_len: usize,
    mn_band_buf: [128]u8,
    mn_band_len: usize,
    desc_value_buf: [256]u8,
    desc_value_len: usize,
    desc_fama_buf: [256]u8,
    desc_fama_len: usize,
    desc_band_buf: [256]u8,
    desc_band_len: usize,

    pub const InitError = error{
        InvalidFastLimitLength,
        InvalidSlowLimitLength,
        InvalidFastLimitSmoothingFactor,
        InvalidSlowLimitSmoothingFactor,
    } || ht.VerifyError;

    /// Creates an instance with default parameters.
    pub fn initDefault() InitError!MesaAdaptiveMovingAverage {
        return initLength(.{});
    }

    /// Creates an instance from length parameters.
    pub fn initLength(params: LengthParams) InitError!MesaAdaptiveMovingAverage {
        if (params.fast_limit_length < 2) {
            return InitError.InvalidFastLimitLength;
        }
        if (params.slow_limit_length < 2) {
            return InitError.InvalidSlowLimitLength;
        }

        const alpha_fast = 2.0 / @as(f64, @floatFromInt(1 + params.fast_limit_length));
        const alpha_slow = 2.0 / @as(f64, @floatFromInt(1 + params.slow_limit_length));

        // Build estimator moniker.
        var est_moniker_buf: [64]u8 = undefined;
        var est_moniker: []const u8 = "";
        var estimator = ht.newCycleEstimator(params.estimator_type, &params.estimator_params) catch |err| return err;

        if (params.estimator_type != .homodyne_discriminator or
            params.estimator_params.smoothing_length != 4 or
            params.estimator_params.alpha_ema_quadrature_in_phase != 0.2 or
            params.estimator_params.alpha_ema_period != 0.2)
        {
            const m = ht.estimatorMoniker(&est_moniker_buf, params.estimator_type, &estimator);
            if (m.len > 0) {
                var tmp: [66]u8 = undefined;
                const full = std.fmt.bufPrint(&tmp, ", {s}", .{m}) catch "";
                @memcpy(est_moniker_buf[0..full.len], full);
                est_moniker = est_moniker_buf[0..full.len];
            }
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc,
            qc,
            tc,
        );

        // Build mnemonics: mama(fast, slow, est, comp), fama(...), mama-fama(...)
        var mn_value_buf: [128]u8 = undefined;
        const mn_value = std.fmt.bufPrint(&mn_value_buf, "mama({d}, {d}{s}{s})", .{
            params.fast_limit_length, params.slow_limit_length, est_moniker, triple,
        }) catch return InitError.InvalidFastLimitLength;

        var mn_fama_buf: [128]u8 = undefined;
        const mn_fama = std.fmt.bufPrint(&mn_fama_buf, "fama({d}, {d}{s}{s})", .{
            params.fast_limit_length, params.slow_limit_length, est_moniker, triple,
        }) catch return InitError.InvalidFastLimitLength;

        var mn_band_buf: [128]u8 = undefined;
        const mn_band = std.fmt.bufPrint(&mn_band_buf, "mama-fama({d}, {d}{s}{s})", .{
            params.fast_limit_length, params.slow_limit_length, est_moniker, triple,
        }) catch return InitError.InvalidFastLimitLength;

        return buildResult(
            alpha_fast,
            alpha_slow,
            estimator,
            bc,
            qc,
            tc,
            mn_value_buf,
            mn_value.len,
            mn_fama_buf,
            mn_fama.len,
            mn_band_buf,
            mn_band.len,
        );
    }

    /// Creates an instance from smoothing factor parameters.
    pub fn initSmoothingFactor(params: SmoothingFactorParams) InitError!MesaAdaptiveMovingAverage {
        var alpha_fast = params.fast_limit_smoothing_factor;
        var alpha_slow = params.slow_limit_smoothing_factor;

        if (alpha_fast < 0.0 or alpha_fast > 1.0) {
            return InitError.InvalidFastLimitSmoothingFactor;
        }
        if (alpha_slow < 0.0 or alpha_slow > 1.0) {
            return InitError.InvalidSlowLimitSmoothingFactor;
        }

        const epsilon = 0.00000001;
        if (alpha_fast < epsilon) alpha_fast = epsilon;
        if (alpha_slow < epsilon) alpha_slow = epsilon;

        var est_moniker_buf: [64]u8 = undefined;
        var est_moniker: []const u8 = "";
        var estimator = ht.newCycleEstimator(params.estimator_type, &params.estimator_params) catch |err| return err;

        if (params.estimator_type != .homodyne_discriminator or
            params.estimator_params.smoothing_length != 4 or
            params.estimator_params.alpha_ema_quadrature_in_phase != 0.2 or
            params.estimator_params.alpha_ema_period != 0.2)
        {
            const m = ht.estimatorMoniker(&est_moniker_buf, params.estimator_type, &estimator);
            if (m.len > 0) {
                var tmp: [66]u8 = undefined;
                const full = std.fmt.bufPrint(&tmp, ", {s}", .{m}) catch "";
                @memcpy(est_moniker_buf[0..full.len], full);
                est_moniker = est_moniker_buf[0..full.len];
            }
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc,
            qc,
            tc,
        );

        var mn_value_buf: [128]u8 = undefined;
        const mn_value = std.fmt.bufPrint(&mn_value_buf, "mama({d:.3}, {d:.3}{s}{s})", .{
            alpha_fast, alpha_slow, est_moniker, triple,
        }) catch return InitError.InvalidFastLimitSmoothingFactor;

        var mn_fama_buf: [128]u8 = undefined;
        const mn_fama = std.fmt.bufPrint(&mn_fama_buf, "fama({d:.3}, {d:.3}{s}{s})", .{
            alpha_fast, alpha_slow, est_moniker, triple,
        }) catch return InitError.InvalidFastLimitSmoothingFactor;

        var mn_band_buf: [128]u8 = undefined;
        const mn_band = std.fmt.bufPrint(&mn_band_buf, "mama-fama({d:.3}, {d:.3}{s}{s})", .{
            alpha_fast, alpha_slow, est_moniker, triple,
        }) catch return InitError.InvalidFastLimitSmoothingFactor;

        return buildResult(
            alpha_fast,
            alpha_slow,
            estimator,
            bc,
            qc,
            tc,
            mn_value_buf,
            mn_value.len,
            mn_fama_buf,
            mn_fama.len,
            mn_band_buf,
            mn_band.len,
        );
    }

    fn buildResult(
        alpha_fast: f64,
        alpha_slow: f64,
        estimator: CycleEstimator,
        bc: bar_component.BarComponent,
        qc: quote_component.QuoteComponent,
        tc: trade_component.TradeComponent,
        mn_value_buf: [128]u8,
        mn_value_len: usize,
        mn_fama_buf: [128]u8,
        mn_fama_len: usize,
        mn_band_buf: [128]u8,
        mn_band_len: usize,
    ) InitError!MesaAdaptiveMovingAverage {
        const descr = "Mesa adaptive moving average ";

        var desc_value_buf: [256]u8 = undefined;
        const desc_value = std.fmt.bufPrint(&desc_value_buf, "{s}{s}", .{
            descr, mn_value_buf[0..mn_value_len],
        }) catch return InitError.InvalidFastLimitLength;

        var desc_fama_buf: [256]u8 = undefined;
        const desc_fama = std.fmt.bufPrint(&desc_fama_buf, "{s}{s}", .{
            descr, mn_fama_buf[0..mn_fama_len],
        }) catch return InitError.InvalidFastLimitLength;

        var desc_band_buf: [256]u8 = undefined;
        const desc_band = std.fmt.bufPrint(&desc_band_buf, "{s}{s}", .{
            descr, mn_band_buf[0..mn_band_len],
        }) catch return InitError.InvalidFastLimitLength;

        return .{
            .alpha_fast_limit = alpha_fast,
            .alpha_slow_limit = alpha_slow,
            .previous_phase = 0,
            .mama = 0,
            .fama = 0,
            .htce = estimator,
            .is_phase_cached = false,
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mn_value_buf = mn_value_buf,
            .mn_value_len = mn_value_len,
            .mn_fama_buf = mn_fama_buf,
            .mn_fama_len = mn_fama_len,
            .mn_band_buf = mn_band_buf,
            .mn_band_len = mn_band_len,
            .desc_value_buf = desc_value_buf,
            .desc_value_len = desc_value.len,
            .desc_fama_buf = desc_fama_buf,
            .desc_fama_len = desc_fama.len,
            .desc_band_buf = desc_band_buf,
            .desc_band_len = desc_band.len,
        };
    }

    /// Must be called after init to fix internal slice pointers.
    pub fn fixSlices(self: *MesaAdaptiveMovingAverage) void {
        _ = self;
        // No internal slices to fix (htce has no slices).
    }

    /// Update the indicator given the next sample.
    /// Returns [2]f64{mama, fama}. Returns NaN if not yet primed.
    pub fn update(self: *MesaAdaptiveMovingAverage, sample: f64) [2]f64 {
        const nan = math.nan(f64);

        if (math.isNan(sample)) {
            return .{ nan, nan };
        }

        self.htce.update(sample);

        if (self.primed) {
            const mama_val = self.calculate(sample);
            return .{ mama_val, self.fama };
        }

        if (self.htce.primed()) {
            if (self.is_phase_cached) {
                self.primed = true;
                const mama_val = self.calculate(sample);
                return .{ mama_val, self.fama };
            }

            self.is_phase_cached = true;
            self.previous_phase = self.calculatePhase();
            self.mama = sample;
            self.fama = sample;
        }

        return .{ nan, nan };
    }

    pub fn isPrimed(self: *const MesaAdaptiveMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const MesaAdaptiveMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .mesa_adaptive_moving_average,
            self.mn_value_buf[0..self.mn_value_len],
            self.desc_value_buf[0..self.desc_value_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mn_value_buf[0..self.mn_value_len], .description = self.desc_value_buf[0..self.desc_value_len] },
                .{ .mnemonic = self.mn_fama_buf[0..self.mn_fama_len], .description = self.desc_fama_buf[0..self.desc_fama_len] },
                .{ .mnemonic = self.mn_band_buf[0..self.mn_band_len], .description = self.desc_band_buf[0..self.desc_band_len] },
            },
        );
    }

    pub fn updateScalar(self: *MesaAdaptiveMovingAverage, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *MesaAdaptiveMovingAverage, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *MesaAdaptiveMovingAverage, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *MesaAdaptiveMovingAverage, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *MesaAdaptiveMovingAverage, time: i64, sample: f64) OutputArray {
        var out = OutputArray{};
        const result = self.update(sample);
        const mama_val = result[0];
        const fama_val = result[1];

        out.append(.{ .scalar = .{ .time = time, .value = mama_val } });
        out.append(.{ .scalar = .{ .time = time, .value = fama_val } });
        out.append(.{ .band = .{ .time = time, .upper = mama_val, .lower = fama_val } });
        return out;
    }

    fn calculatePhase(self: *MesaAdaptiveMovingAverage) f64 {
        if (self.htce.inPhase() == 0) {
            return self.previous_phase;
        }

        const rad2deg = 180.0 / math.pi;
        const phase = math.atan(self.htce.quadrature() / self.htce.inPhase()) * rad2deg;
        if (!math.isNan(phase) and !math.isInf(phase)) {
            return phase;
        }

        return self.previous_phase;
    }

    fn calculateMama(self: *MesaAdaptiveMovingAverage, sample: f64) f64 {
        const phase = self.calculatePhase();

        // Phase rate of change.
        var phase_rate_of_change = self.previous_phase - phase;
        self.previous_phase = phase;

        if (phase_rate_of_change < 1) {
            phase_rate_of_change = 1;
        }

        // Alpha = fast_limit / phase_rate_of_change, clamped to [slow, fast].
        const alpha = @min(@max(self.alpha_fast_limit / phase_rate_of_change, self.alpha_slow_limit), self.alpha_fast_limit);

        self.mama = alpha * sample + (1.0 - alpha) * self.mama;

        return alpha;
    }

    fn calculate(self: *MesaAdaptiveMovingAverage, sample: f64) f64 {
        const alpha = self.calculateMama(sample) / 2.0;
        self.fama = alpha * self.mama + (1.0 - alpha) * self.fama;
        return self.mama;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *MesaAdaptiveMovingAverage) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(MesaAdaptiveMovingAverage);
};

// ─── Tests ───────────────────────────────────────────────────────────────────

const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    if (math.isNan(a) and math.isNan(b)) return true;
    if (math.isNan(a) or math.isNan(b)) return false;
    return @abs(a - b) <= epsilon;
}

// Input data from TA-Lib test_MAMA.xsl, 252 entries.
// Expected MAMA values from TA-Lib test_MAMA_new.xsl, 252 entries.
// Expected FAMA values from TA-Lib test_MAMA_new.xsl, 252 entries.
test "MAMA update reference MAMA values" {
    const lprimed = 26;
    var mama = try MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 3,
        .slow_limit_length = 39,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
            .warm_up_period = 0,
        },
    });

    for (0..lprimed) |i| {
        const result = mama.update(testdata.test_input[i]);
        try testing.expect(math.isNan(result[0]));
    }

    for (lprimed..testdata.test_input.len) |i| {
        const result = mama.update(testdata.test_input[i]);
        try testing.expect(almostEqual(result[0], testdata.test_expected_mama[i], 1e-9));
    }

    // NaN input should return NaN.
    const nan_result = mama.update(math.nan(f64));
    try testing.expect(math.isNan(nan_result[0]));
}

test "MAMA update reference FAMA values" {
    const lprimed = 26;
    var mama = try MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 3,
        .slow_limit_length = 39,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
            .warm_up_period = 0,
        },
    });

    for (0..lprimed) |i| {
        const result = mama.update(testdata.test_input[i]);
        try testing.expect(math.isNan(result[1]));
    }

    for (lprimed..testdata.test_input.len) |i| {
        const result = mama.update(testdata.test_input[i]);
        try testing.expect(almostEqual(result[1], testdata.test_expected_fama[i], 1e-9));
    }
}

test "MAMA isPrimed" {
    const lprimed = 26;
    var mama = try MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 3,
        .slow_limit_length = 39,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
            .warm_up_period = 0,
        },
    });

    try testing.expect(!mama.isPrimed());

    for (0..lprimed) |_| {
        _ = mama.update(testdata.test_input[0]);
        try testing.expect(!mama.isPrimed());
    }

    // Feed actual data to prime.
    var mama2 = try MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 3,
        .slow_limit_length = 39,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
            .warm_up_period = 0,
        },
    });

    for (0..lprimed) |i| {
        _ = mama2.update(testdata.test_input[i]);
        try testing.expect(!mama2.isPrimed());
    }

    _ = mama2.update(testdata.test_input[lprimed]);
    try testing.expect(mama2.isPrimed());
}

test "MAMA metadata length params" {
    const mama = try MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 2,
        .slow_limit_length = 40,
    });

    var out: Metadata = undefined;
    mama.getMetadata(&out);

    try testing.expectEqualStrings("mama(2, 40)", out.mnemonic);
    try testing.expectEqualStrings("Mesa adaptive moving average mama(2, 40)", out.description);
    try testing.expectEqual(@as(usize, 3), out.outputs_len);
    try testing.expectEqualStrings("mama(2, 40)", out.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("fama(2, 40)", out.outputs_buf[1].mnemonic);
    try testing.expectEqualStrings("mama-fama(2, 40)", out.outputs_buf[2].mnemonic);
}

test "MAMA metadata smoothing factor params" {
    const mama = try MesaAdaptiveMovingAverage.initSmoothingFactor(.{
        .fast_limit_smoothing_factor = 0.666666666,
        .slow_limit_smoothing_factor = 0.064516129,
    });

    var out: Metadata = undefined;
    mama.getMetadata(&out);

    try testing.expectEqualStrings("mama(0.667, 0.065)", out.mnemonic);
}

test "MAMA metadata with component" {
    const mama = try MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 3,
        .slow_limit_length = 39,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
            .warm_up_period = 0,
        },
        .bar_component = .median,
        .quote_component = .mid,
        .trade_component = .price,
    });

    var out: Metadata = undefined;
    mama.getMetadata(&out);

    try testing.expectEqualStrings("mama(3, 39, hl/2)", out.mnemonic);
}

test "MAMA metadata with non-default estimator" {
    const mama = try MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 2,
        .slow_limit_length = 40,
        .estimator_type = .homodyne_discriminator,
        .estimator_params = .{
            .smoothing_length = 3,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
            .warm_up_period = 0,
        },
        .bar_component = .median,
        .quote_component = .mid,
        .trade_component = .price,
    });

    var out: Metadata = undefined;
    mama.getMetadata(&out);

    try testing.expectEqualStrings("mama(2, 40, hd(3, 0.200, 0.200), hl/2)", out.mnemonic);
}

test "MAMA validation length" {
    try testing.expectError(error.InvalidFastLimitLength, MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 1,
        .slow_limit_length = 39,
    }));
    try testing.expectError(error.InvalidFastLimitLength, MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 0,
        .slow_limit_length = 39,
    }));
    try testing.expectError(error.InvalidFastLimitLength, MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = -1,
        .slow_limit_length = 39,
    }));
    try testing.expectError(error.InvalidSlowLimitLength, MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 3,
        .slow_limit_length = 1,
    }));
    try testing.expectError(error.InvalidSlowLimitLength, MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 3,
        .slow_limit_length = 0,
    }));
    try testing.expectError(error.InvalidSlowLimitLength, MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 3,
        .slow_limit_length = -1,
    }));
}

test "MAMA validation smoothing factor" {
    try testing.expectError(error.InvalidFastLimitSmoothingFactor, MesaAdaptiveMovingAverage.initSmoothingFactor(.{
        .fast_limit_smoothing_factor = -0.00000001,
        .slow_limit_smoothing_factor = 0.33333333,
    }));
    try testing.expectError(error.InvalidFastLimitSmoothingFactor, MesaAdaptiveMovingAverage.initSmoothingFactor(.{
        .fast_limit_smoothing_factor = 1.00000001,
        .slow_limit_smoothing_factor = 0.33333333,
    }));
    try testing.expectError(error.InvalidSlowLimitSmoothingFactor, MesaAdaptiveMovingAverage.initSmoothingFactor(.{
        .fast_limit_smoothing_factor = 0.66666666,
        .slow_limit_smoothing_factor = -0.00000001,
    }));
    try testing.expectError(error.InvalidSlowLimitSmoothingFactor, MesaAdaptiveMovingAverage.initSmoothingFactor(.{
        .fast_limit_smoothing_factor = 0.66666666,
        .slow_limit_smoothing_factor = 1.00000001,
    }));
}

test "MAMA updateEntity" {
    const lprimed = 26;
    var mama = try MesaAdaptiveMovingAverage.initLength(.{
        .fast_limit_length = 3,
        .slow_limit_length = 39,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
            .warm_up_period = 0,
        },
    });

    // Feed lprimed samples with value 0 to prime.
    for (0..lprimed) |_| {
        _ = mama.update(0.0);
    }

    // Now test entity updates.
    const time: i64 = 1617235200;
    const inp: f64 = 3.0;
    const expected_mama_val: f64 = 1.5;
    const expected_fama_val: f64 = 0.375;

    // Scalar
    const scalar = Scalar{ .time = time, .value = inp };
    const outputs = mama.updateScalar(&scalar);
    try testing.expectEqual(@as(usize, 3), outputs.len);
    try testing.expect(almostEqual(outputs.values[0].scalar.value, expected_mama_val, 1e-15));
    try testing.expect(almostEqual(outputs.values[1].scalar.value, expected_fama_val, 1e-15));
    try testing.expect(almostEqual(outputs.values[2].band.upper, expected_mama_val, 1e-15));
    try testing.expect(almostEqual(outputs.values[2].band.lower, expected_fama_val, 1e-15));
}

test "MAMA smoothing factor with estimator types" {
    // Test with phase accumulator
    const mama_pa = try MesaAdaptiveMovingAverage.initSmoothingFactor(.{
        .fast_limit_smoothing_factor = 0.66666666,
        .slow_limit_smoothing_factor = 0.33333333,
        .estimator_type = .phase_accumulator,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
            .warm_up_period = 0,
        },
        .bar_component = .median,
        .quote_component = .mid,
        .trade_component = .price,
    });
    var out: Metadata = undefined;
    mama_pa.getMetadata(&out);
    try testing.expectEqualStrings("mama(0.667, 0.333, pa(4, 0.200, 0.200), hl/2)", out.mnemonic);

    // Test with dual differentiator
    const mama_dd = try MesaAdaptiveMovingAverage.initSmoothingFactor(.{
        .fast_limit_smoothing_factor = 0.66666666,
        .slow_limit_smoothing_factor = 0.33333333,
        .estimator_type = .dual_differentiator,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
            .warm_up_period = 0,
        },
        .bar_component = .median,
        .quote_component = .mid,
        .trade_component = .price,
    });
    var out2: Metadata = undefined;
    mama_dd.getMetadata(&out2);
    try testing.expectEqualStrings("mama(0.667, 0.333, dd(4, 0.200, 0.200), hl/2)", out2.mnemonic);
}
