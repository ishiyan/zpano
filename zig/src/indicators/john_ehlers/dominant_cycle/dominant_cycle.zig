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

const ht = @import("../hilbert_transformer/hilbert_transformer.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const CycleEstimator = ht.CycleEstimator;
const CycleEstimatorType = ht.CycleEstimatorType;
const CycleEstimatorParams = ht.CycleEstimatorParams;

/// Enumerates the outputs of the dominant cycle indicator.
pub const DominantCycleOutput = enum(u8) {
    raw_period = 1,
    period = 2,
    phase = 3,
};

/// Parameters to create a DominantCycle indicator.
pub const Params = struct {
    estimator_type: CycleEstimatorType = .homodyne_discriminator,
    estimator_params: CycleEstimatorParams = .{},
    alpha_ema_period_additional: f64 = 0.33,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehlers' Dominant Cycle computes the instantaneous cycle period and phase
/// derived from a Hilbert transformer cycle estimator.
///
/// Three outputs:
///   - RawPeriod: the raw instantaneous cycle period produced by the Hilbert transformer estimator.
///   - Period: the dominant cycle period obtained by additional EMA smoothing of the raw period.
///   - Phase: the dominant cycle phase, in degrees.
pub const DominantCycle = struct {
    alpha_ema_period_additional: f64,
    one_min_alpha_ema_period_additional: f64,
    smoothed_period: f64,
    smoothed_phase: f64,
    smoothed_input: []f64,
    smoothed_input_length_min1: usize,
    htce: CycleEstimator,
    primed: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_raw_buf: [128]u8,
    mnemonic_raw_len: usize,
    mnemonic_period_buf: [128]u8,
    mnemonic_period_len: usize,
    mnemonic_phase_buf: [128]u8,
    mnemonic_phase_len: usize,
    description_raw_buf: [192]u8,
    description_raw_len: usize,
    description_period_buf: [192]u8,
    description_period_len: usize,
    description_phase_buf: [192]u8,
    description_phase_len: usize,
    // Fixed-size buffer for smoothed input (max 50 elements).
    smoothed_input_storage: [50]f64,

    pub const InitError = error{
        InvalidAlphaEmaPeriodAdditional,
    } || ht.VerifyError;

    /// Creates a DominantCycle with default parameters.
    pub fn initDefault() InitError!DominantCycle {
        return init(.{
            .estimator_type = .homodyne_discriminator,
            .estimator_params = .{
                .smoothing_length = 4,
                .alpha_ema_quadrature_in_phase = 0.2,
                .alpha_ema_period = 0.2,
                .warm_up_period = 100,
            },
            .alpha_ema_period_additional = 0.33,
        });
    }

    /// Creates a DominantCycle with the given parameters.
    pub fn init(params: Params) InitError!DominantCycle {
        const alpha = params.alpha_ema_period_additional;
        if (alpha <= 0.0 or alpha > 1.0) {
            return InitError.InvalidAlphaEmaPeriodAdditional;
        }

        var estimator = try ht.newCycleEstimator(params.estimator_type, &params.estimator_params);

        // Build estimator moniker (only if non-default).
        var est_moniker_buf: [64]u8 = undefined;
        var est_moniker: []const u8 = "";
        if (params.estimator_type != .homodyne_discriminator or
            params.estimator_params.smoothing_length != 4 or
            params.estimator_params.alpha_ema_quadrature_in_phase != 0.2 or
            params.estimator_params.alpha_ema_period != 0.2)
        {
            const m = ht.estimatorMoniker(&est_moniker_buf, params.estimator_type, &estimator);
            if (m.len > 0) {
                // Prepend ", "
                var tmp: [66]u8 = undefined;
                const full = std.fmt.bufPrint(&tmp, ", {s}", .{m}) catch "";
                @memcpy(est_moniker_buf[0..full.len], full);
                est_moniker = est_moniker_buf[0..full.len];
            }
        }

        // Resolve component defaults.
        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            params.bar_component orelse bar_component.default_bar_component,
            params.quote_component orelse quote_component.default_quote_component,
            params.trade_component orelse trade_component.default_trade_component,
        );

        // Build mnemonics.
        var mnemonic_raw_buf: [128]u8 = undefined;
        const mn_raw = std.fmt.bufPrint(&mnemonic_raw_buf, "dcp-raw({d:.3}{s}{s})", .{ alpha, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        const mnemonic_raw_len = mn_raw.len;

        var mnemonic_period_buf: [128]u8 = undefined;
        const mn_per = std.fmt.bufPrint(&mnemonic_period_buf, "dcp({d:.3}{s}{s})", .{ alpha, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        const mnemonic_period_len = mn_per.len;

        var mnemonic_phase_buf: [128]u8 = undefined;
        const mn_pha = std.fmt.bufPrint(&mnemonic_phase_buf, "dcph({d:.3}{s}{s})", .{ alpha, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        const mnemonic_phase_len = mn_pha.len;

        var description_raw_buf: [192]u8 = undefined;
        const desc_raw = std.fmt.bufPrint(&description_raw_buf, "Dominant cycle raw period {s}", .{mn_raw}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        const description_raw_len = desc_raw.len;

        var description_period_buf: [192]u8 = undefined;
        const desc_per = std.fmt.bufPrint(&description_period_buf, "Dominant cycle period {s}", .{mn_per}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        const description_period_len = desc_per.len;

        var description_phase_buf: [192]u8 = undefined;
        const desc_pha = std.fmt.bufPrint(&description_phase_buf, "Dominant cycle phase {s}", .{mn_pha}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        const description_phase_len = desc_pha.len;

        const max_period = estimator.maxPeriod();

        return .{
            .alpha_ema_period_additional = alpha,
            .one_min_alpha_ema_period_additional = 1.0 - alpha,
            .smoothed_period = 0.0,
            .smoothed_phase = 0.0,
            .smoothed_input = undefined, // Set by fixSlices.
            .smoothed_input_length_min1 = max_period - 1,
            .htce = estimator,
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_raw_buf = mnemonic_raw_buf,
            .mnemonic_raw_len = mnemonic_raw_len,
            .mnemonic_period_buf = mnemonic_period_buf,
            .mnemonic_period_len = mnemonic_period_len,
            .mnemonic_phase_buf = mnemonic_phase_buf,
            .mnemonic_phase_len = mnemonic_phase_len,
            .description_raw_buf = description_raw_buf,
            .description_raw_len = description_raw_len,
            .description_period_buf = description_period_buf,
            .description_period_len = description_period_len,
            .description_phase_buf = description_phase_buf,
            .description_phase_len = description_phase_len,
            .smoothed_input_storage = [_]f64{0} ** 50,
        };
    }

    /// Must be called after init to fix internal slice pointers.
    pub fn fixSlices(self: *DominantCycle) void {
        const max_period = self.smoothed_input_length_min1 + 1;
        self.smoothed_input = self.smoothed_input_storage[0..max_period];
    }

    /// Update the indicator given the next sample.
    /// Returns (rawPeriod, period, phase). Returns NaN values if not yet primed.
    pub fn update(self: *DominantCycle, sample: f64) [3]f64 {
        const nan = math.nan(f64);

        if (math.isNan(sample)) {
            return .{ nan, nan, nan };
        }

        self.htce.update(sample);
        self.pushSmoothedInput(self.htce.smoothed());

        if (self.primed) {
            self.smoothed_period = self.alpha_ema_period_additional * self.htce.period() +
                self.one_min_alpha_ema_period_additional * self.smoothed_period;
            self.calculateSmoothedPhase();
            return .{ self.htce.period(), self.smoothed_period, self.smoothed_phase };
        }

        if (self.htce.primed()) {
            self.primed = true;
            self.smoothed_period = self.htce.period();
            self.calculateSmoothedPhase();
            return .{ self.htce.period(), self.smoothed_period, self.smoothed_phase };
        }

        return .{ nan, nan, nan };
    }

    /// Returns the current WMA-smoothed price value. Returns NaN if not primed.
    pub fn smoothedPrice(self: *const DominantCycle) f64 {
        if (!self.primed) {
            return math.nan(f64);
        }
        return self.htce.smoothed();
    }

    /// Returns the maximum cycle period supported by the underlying estimator.
    pub fn maxPeriod(self: *const DominantCycle) usize {
        return self.htce.maxPeriod();
    }

    pub fn isPrimed(self: *const DominantCycle) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const DominantCycle, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .dominant_cycle,
            self.mnemonic_period_buf[0..self.mnemonic_period_len],
            self.description_period_buf[0..self.description_period_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mnemonic_raw_buf[0..self.mnemonic_raw_len], .description = self.description_raw_buf[0..self.description_raw_len] },
                .{ .mnemonic = self.mnemonic_period_buf[0..self.mnemonic_period_len], .description = self.description_period_buf[0..self.description_period_len] },
                .{ .mnemonic = self.mnemonic_phase_buf[0..self.mnemonic_phase_len], .description = self.description_phase_buf[0..self.description_phase_len] },
            },
        );
    }

    pub fn updateScalar(self: *DominantCycle, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *DominantCycle, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *DominantCycle, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *DominantCycle, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *DominantCycle, time: i64, sample: f64) OutputArray {
        const result = self.update(sample);
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = result[0] } });
        out.append(.{ .scalar = .{ .time = time, .value = result[1] } });
        out.append(.{ .scalar = .{ .time = time, .value = result[2] } });
        return out;
    }

    fn pushSmoothedInput(self: *DominantCycle, value: f64) void {
        var i: usize = self.smoothed_input_length_min1;
        while (i > 0) : (i -= 1) {
            self.smoothed_input[i] = self.smoothed_input[i - 1];
        }
        self.smoothed_input[0] = value;
    }

    fn calculateSmoothedPhase(self: *DominantCycle) void {
        const rad2deg = 180.0 / math.pi;
        const two_pi = 2.0 * math.pi;
        const epsilon = 0.01;
        const ninety = 90.0;
        const one_eighty = 180.0;
        const three_sixty = 360.0;

        // Sum over one full dominant cycle.
        var length = @as(usize, @intFromFloat(@floor(self.smoothed_period + 0.5)));
        if (length > self.smoothed_input_length_min1) {
            length = self.smoothed_input_length_min1;
        }

        var real_part: f64 = 0;
        var imag_part: f64 = 0;
        const f_length: f64 = @floatFromInt(length);

        for (0..length) |i| {
            const fi: f64 = @floatFromInt(i);
            const temp = two_pi * fi / f_length;
            const smoothed_val = self.smoothed_input[i];
            real_part += smoothed_val * @sin(temp);
            imag_part += smoothed_val * @cos(temp);
        }

        const previous = self.smoothed_phase;
        var phase = math.atan(real_part / imag_part) * rad2deg;
        if (math.isNan(phase) or math.isInf(phase)) {
            phase = previous;
        }

        if (@abs(imag_part) <= epsilon) {
            if (real_part > 0) {
                phase += ninety;
            } else if (real_part < 0) {
                phase -= ninety;
            }
        }

        // Introduce the 90 degree reference shift.
        phase += ninety;
        // Compensate for one bar lag.
        phase += three_sixty / self.smoothed_period;
        // Resolve phase ambiguity.
        if (imag_part < 0) {
            phase += one_eighty;
        }
        // Cycle wraparound.
        if (phase > three_sixty) {
            phase -= three_sixty;
        }

        self.smoothed_phase = phase;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *DominantCycle) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(DominantCycle);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tol: f64) bool {
    return @abs(a - b) <= tol;
}

/// Returns the shortest signed angular difference between two angles, in (-180, 180].
fn phaseDiff(a: f64, b: f64) f64 {
    var d = @mod(a - b, 360.0);
    if (d > 180.0) {
        d -= 360.0;
    } else if (d <= -180.0) {
        d += 360.0;
    }
    return d;
}

// 252-entry input data from TA-Lib tests, test_MAMA.xsl.
// Expected period data, 252 entries, smoothed as AI18=0.33*X18+0.67*AI17.
// Expected phase data, 252 entries.
const tolerance = 1e-4;

test "DC update period" {
    const skip = 9;
    const settle_skip = 177;

    var dc = try DominantCycle.initDefault();
    dc.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = dc.update(testdata.test_input[i]);
        const period_val = result[1];
        if (math.isNan(period_val) or i < settle_skip) {
            continue;
        }
        try testing.expect(almostEqual(testdata.test_expected_period[i], period_val, tolerance));
    }
}

test "DC update phase" {
    const skip = 9;
    const settle_skip = 177;

    var dc = try DominantCycle.initDefault();
    dc.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = dc.update(testdata.test_input[i]);
        const phase_val = result[2];
        if (math.isNan(phase_val) or i < settle_skip) {
            continue;
        }
        if (math.isNan(testdata.test_expected_phase[i])) {
            continue;
        }
        try testing.expect(@abs(phaseDiff(testdata.test_expected_phase[i], phase_val)) <= tolerance);
    }
}

test "DC NaN input returns NaN triple" {
    var dc = try DominantCycle.initDefault();
    dc.fixSlices();

    const result = dc.update(math.nan(f64));
    try testing.expect(math.isNan(result[0]));
    try testing.expect(math.isNan(result[1]));
    try testing.expect(math.isNan(result[2]));
}

test "DC isPrimed" {
    var dc = try DominantCycle.initDefault();
    dc.fixSlices();

    try testing.expect(!dc.isPrimed());

    var primed_at: ?usize = null;
    for (0..testdata.test_input.len) |i| {
        _ = dc.update(testdata.test_input[i]);
        if (dc.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expect(primed_at != null);
    try testing.expect(dc.isPrimed());
}

test "DC metadata default" {
    var dc = try DominantCycle.initDefault();
    dc.fixSlices();

    var meta: Metadata = undefined;
    dc.getMetadata(&meta);

    try testing.expectEqual(Identifier.dominant_cycle, meta.identifier);
    try testing.expectEqualStrings("dcp(0.330)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 3), meta.outputs_len);
}

test "DC metadata phase accumulator" {
    var dc = try DominantCycle.init(.{
        .alpha_ema_period_additional = 0.5,
        .estimator_type = .phase_accumulator,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
        },
    });
    dc.fixSlices();

    var meta: Metadata = undefined;
    dc.getMetadata(&meta);

    try testing.expectEqualStrings("dcp(0.500, pa(4, 0.200, 0.200))", meta.mnemonic);
}

test "DC constructor errors" {
    // Alpha <= 0
    try testing.expectError(error.InvalidAlphaEmaPeriodAdditional, DominantCycle.init(.{
        .alpha_ema_period_additional = 0.0,
    }));
    // Alpha > 1
    try testing.expectError(error.InvalidAlphaEmaPeriodAdditional, DominantCycle.init(.{
        .alpha_ema_period_additional = 1.00000001,
    }));
}

test "DC smoothedPrice" {
    var dc = try DominantCycle.initDefault();
    dc.fixSlices();

    // Before any updates: NaN.
    try testing.expect(math.isNan(dc.smoothedPrice()));

    for (0..testdata.test_input.len) |i| {
        _ = dc.update(testdata.test_input[i]);
        const v = dc.smoothedPrice();
        if (dc.isPrimed()) {
            try testing.expect(!math.isNan(v));
            break;
        } else {
            try testing.expect(math.isNan(v));
        }
    }
}

test "DC maxPeriod" {
    var dc = try DominantCycle.initDefault();
    dc.fixSlices();
    try testing.expectEqual(dc.smoothed_input.len, dc.maxPeriod());
}

test "DC updateScalar" {
    var dc = try DominantCycle.initDefault();
    dc.fixSlices();

    // Prime the indicator.
    for (0..testdata.test_input.len) |i| {
        const s = Scalar{ .time = @intCast(i), .value = testdata.test_input[i] };
        const out = dc.updateScalar(&s);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 3), outputs.len);
    }
}
