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

const dc_mod = @import("../dominant_cycle/dominant_cycle.zig");
const ht = @import("../hilbert_transformer/hilbert_transformer.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const DominantCycle = dc_mod.DominantCycle;
const CycleEstimatorType = ht.CycleEstimatorType;
const CycleEstimatorParams = ht.CycleEstimatorParams;

const deg2rad = math.pi / 180.0;

/// Enumerates the outputs of the sine wave indicator.
pub const SineWaveOutput = enum(u8) {
    value = 1,
    lead = 2,
    band = 3,
    dominant_cycle_period = 4,
    dominant_cycle_phase = 5,
};

/// Parameters to create a SineWave indicator.
pub const Params = struct {
    estimator_type: CycleEstimatorType = .homodyne_discriminator,
    estimator_params: CycleEstimatorParams = .{},
    alpha_ema_period_additional: f64 = 0.33,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehlers' Sine Wave indicator.
///
/// Five outputs:
///   - Value: sin(phase * deg2rad).
///   - Lead: sin((phase + 45) * deg2rad).
///   - Band: upper=Value, lower=Lead.
///   - DominantCyclePeriod: the smoothed dominant cycle period.
///   - DominantCyclePhase: the dominant cycle phase, in degrees.
pub const SineWave = struct {
    dc: DominantCycle,
    primed: bool,
    sine_value: f64,
    lead_value: f64,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    // Mnemonic/description buffers for 5 outputs.
    mn_value_buf: [128]u8,
    mn_value_len: usize,
    mn_lead_buf: [128]u8,
    mn_lead_len: usize,
    mn_band_buf: [128]u8,
    mn_band_len: usize,
    mn_dcp_buf: [128]u8,
    mn_dcp_len: usize,
    mn_dcph_buf: [128]u8,
    mn_dcph_len: usize,
    desc_value_buf: [192]u8,
    desc_value_len: usize,
    desc_lead_buf: [192]u8,
    desc_lead_len: usize,
    desc_band_buf: [192]u8,
    desc_band_len: usize,
    desc_dcp_buf: [192]u8,
    desc_dcp_len: usize,
    desc_dcph_buf: [192]u8,
    desc_dcph_len: usize,

    pub const InitError = error{
        InvalidAlphaEmaPeriodAdditional,
    } || ht.VerifyError;

    /// Creates a SineWave with default parameters.
    pub fn initDefault() InitError!SineWave {
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

    /// Creates a SineWave with the given parameters.
    pub fn init(params: Params) InitError!SineWave {
        const alpha = params.alpha_ema_period_additional;
        if (alpha <= 0.0 or alpha > 1.0) {
            return InitError.InvalidAlphaEmaPeriodAdditional;
        }

        // SineWave defaults to BarMedianPrice (not the framework default).
        const bc = params.bar_component orelse .median;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Build DominantCycle with explicit components.
        var dc = DominantCycle.init(.{
            .estimator_type = params.estimator_type,
            .estimator_params = params.estimator_params,
            .alpha_ema_period_additional = alpha,
            .bar_component = bc,
            .quote_component = qc,
            .trade_component = tc,
        }) catch |err| switch (err) {
            error.InvalidAlphaEmaPeriodAdditional => return InitError.InvalidAlphaEmaPeriodAdditional,
            else => return err,
        };
        dc.fixSlices();

        // Build estimator moniker (only if non-default).
        var estimator = ht.newCycleEstimator(params.estimator_type, &params.estimator_params) catch |err| return err;
        var est_moniker_buf: [64]u8 = undefined;
        var est_moniker: []const u8 = "";
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

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc,
            qc,
            tc,
        );

        // Build mnemonics for all 5 outputs.
        var mn_value_buf: [128]u8 = undefined;
        const mn_value = std.fmt.bufPrint(&mn_value_buf, "sw({d:.3}{s}{s})", .{ alpha, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_lead_buf: [128]u8 = undefined;
        const mn_lead = std.fmt.bufPrint(&mn_lead_buf, "sw-lead({d:.3}{s}{s})", .{ alpha, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_band_buf: [128]u8 = undefined;
        const mn_band = std.fmt.bufPrint(&mn_band_buf, "sw-band({d:.3}{s}{s})", .{ alpha, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_dcp_buf: [128]u8 = undefined;
        const mn_dcp = std.fmt.bufPrint(&mn_dcp_buf, "dcp({d:.3}{s}{s})", .{ alpha, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_dcph_buf: [128]u8 = undefined;
        const mn_dcph = std.fmt.bufPrint(&mn_dcph_buf, "dcph({d:.3}{s}{s})", .{ alpha, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var desc_value_buf: [192]u8 = undefined;
        const desc_value = std.fmt.bufPrint(&desc_value_buf, "Sine wave {s}", .{mn_value}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var desc_lead_buf: [192]u8 = undefined;
        const desc_lead = std.fmt.bufPrint(&desc_lead_buf, "Sine wave lead {s}", .{mn_lead}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var desc_band_buf: [192]u8 = undefined;
        const desc_band = std.fmt.bufPrint(&desc_band_buf, "Sine wave band {s}", .{mn_band}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var desc_dcp_buf: [192]u8 = undefined;
        const desc_dcp = std.fmt.bufPrint(&desc_dcp_buf, "Dominant cycle period {s}", .{mn_dcp}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var desc_dcph_buf: [192]u8 = undefined;
        const desc_dcph = std.fmt.bufPrint(&desc_dcph_buf, "Dominant cycle phase {s}", .{mn_dcph}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        const nan = math.nan(f64);

        return .{
            .dc = dc,
            .primed = false,
            .sine_value = nan,
            .lead_value = nan,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mn_value_buf = mn_value_buf,
            .mn_value_len = mn_value.len,
            .mn_lead_buf = mn_lead_buf,
            .mn_lead_len = mn_lead.len,
            .mn_band_buf = mn_band_buf,
            .mn_band_len = mn_band.len,
            .mn_dcp_buf = mn_dcp_buf,
            .mn_dcp_len = mn_dcp.len,
            .mn_dcph_buf = mn_dcph_buf,
            .mn_dcph_len = mn_dcph.len,
            .desc_value_buf = desc_value_buf,
            .desc_value_len = desc_value.len,
            .desc_lead_buf = desc_lead_buf,
            .desc_lead_len = desc_lead.len,
            .desc_band_buf = desc_band_buf,
            .desc_band_len = desc_band.len,
            .desc_dcp_buf = desc_dcp_buf,
            .desc_dcp_len = desc_dcp.len,
            .desc_dcph_buf = desc_dcph_buf,
            .desc_dcph_len = desc_dcph.len,
        };
    }

    /// Must be called after init to fix internal slice pointers.
    pub fn fixSlices(self: *SineWave) void {
        self.dc.fixSlices();
    }

    /// Update the indicator given the next sample.
    /// Returns (value, lead, period, phase). Returns NaN if not yet primed.
    pub fn update(self: *SineWave, sample: f64) [4]f64 {
        const nan = math.nan(f64);

        if (math.isNan(sample)) {
            return .{ nan, nan, nan, nan };
        }

        const dc_result = self.dc.update(sample);
        const period = dc_result[1];
        const phase = dc_result[2];

        if (math.isNan(phase)) {
            return .{ nan, nan, nan, nan };
        }

        const lead_offset = 45.0;

        self.primed = true;
        self.sine_value = @sin(phase * deg2rad);
        self.lead_value = @sin((phase + lead_offset) * deg2rad);

        return .{ self.sine_value, self.lead_value, period, phase };
    }

    pub fn isPrimed(self: *const SineWave) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const SineWave, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .sine_wave,
            self.mn_value_buf[0..self.mn_value_len],
            self.desc_value_buf[0..self.desc_value_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mn_value_buf[0..self.mn_value_len], .description = self.desc_value_buf[0..self.desc_value_len] },
                .{ .mnemonic = self.mn_lead_buf[0..self.mn_lead_len], .description = self.desc_lead_buf[0..self.desc_lead_len] },
                .{ .mnemonic = self.mn_band_buf[0..self.mn_band_len], .description = self.desc_band_buf[0..self.desc_band_len] },
                .{ .mnemonic = self.mn_dcp_buf[0..self.mn_dcp_len], .description = self.desc_dcp_buf[0..self.desc_dcp_len] },
                .{ .mnemonic = self.mn_dcph_buf[0..self.mn_dcph_len], .description = self.desc_dcph_buf[0..self.desc_dcph_len] },
            },
        );
    }

    pub fn updateScalar(self: *SineWave, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *SineWave, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *SineWave, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *SineWave, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *SineWave, time: i64, sample: f64) OutputArray {
        const result = self.update(sample);
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = result[0] } }); // value
        out.append(.{ .scalar = .{ .time = time, .value = result[1] } }); // lead
        out.append(.{ .band = .{ .time = time, .upper = result[0], .lower = result[1] } }); // band
        out.append(.{ .scalar = .{ .time = time, .value = result[2] } }); // period
        out.append(.{ .scalar = .{ .time = time, .value = result[3] } }); // phase
        return out;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *SineWave) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(SineWave);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tol: f64) bool {
    return @abs(a - b) <= tol;
}

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
// Expected sine wave values, 252 entries.
// Expected sine wave lead values, 252 entries.
const tolerance = 1e-4;

test "SW update sine" {
    const skip = 9;
    const settle_skip = 177;

    var sw = try SineWave.initDefault();
    sw.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = sw.update(testdata.test_input[i]);
        const val = result[0];
        if (math.isNan(val) or i < settle_skip) {
            continue;
        }
        if (math.isNan(testdata.test_expected_sine[i])) {
            continue;
        }
        try testing.expect(almostEqual(testdata.test_expected_sine[i], val, tolerance));
    }
}

test "SW update sine lead" {
    const skip = 9;
    const settle_skip = 177;

    var sw = try SineWave.initDefault();
    sw.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = sw.update(testdata.test_input[i]);
        const lead_val = result[1];
        if (math.isNan(lead_val) or i < settle_skip) {
            continue;
        }
        if (math.isNan(testdata.test_expected_sine_lead[i])) {
            continue;
        }
        try testing.expect(almostEqual(testdata.test_expected_sine_lead[i], lead_val, tolerance));
    }
}

test "SW update period" {
    const skip = 9;
    const settle_skip = 177;

    var sw = try SineWave.initDefault();
    sw.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = sw.update(testdata.test_input[i]);
        const period_val = result[2];
        if (math.isNan(period_val) or i < settle_skip) {
            continue;
        }
        try testing.expect(almostEqual(testdata.test_expected_period[i], period_val, tolerance));
    }
}

test "SW update phase" {
    const skip = 9;
    const settle_skip = 177;

    var sw = try SineWave.initDefault();
    sw.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = sw.update(testdata.test_input[i]);
        const phase_val = result[3];
        if (math.isNan(phase_val) or i < settle_skip) {
            continue;
        }
        if (math.isNan(testdata.test_expected_phase[i])) {
            continue;
        }
        try testing.expect(@abs(phaseDiff(testdata.test_expected_phase[i], phase_val)) <= tolerance);
    }
}

test "SW NaN input returns NaN quadruple" {
    var sw = try SineWave.initDefault();
    sw.fixSlices();

    const result = sw.update(math.nan(f64));
    try testing.expect(math.isNan(result[0]));
    try testing.expect(math.isNan(result[1]));
    try testing.expect(math.isNan(result[2]));
    try testing.expect(math.isNan(result[3]));
}

test "SW isPrimed" {
    var sw = try SineWave.initDefault();
    sw.fixSlices();

    try testing.expect(!sw.isPrimed());

    var primed_at: ?usize = null;
    for (0..testdata.test_input.len) |i| {
        _ = sw.update(testdata.test_input[i]);
        if (sw.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expect(primed_at != null);
    try testing.expect(sw.isPrimed());
}

test "SW metadata default" {
    var sw = try SineWave.initDefault();
    sw.fixSlices();

    var meta: Metadata = undefined;
    sw.getMetadata(&meta);

    try testing.expectEqual(Identifier.sine_wave, meta.identifier);
    try testing.expectEqualStrings("sw(0.330, hl/2)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 5), meta.outputs_len);
}

test "SW metadata phase accumulator" {
    var sw = try SineWave.init(.{
        .alpha_ema_period_additional = 0.5,
        .estimator_type = .phase_accumulator,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
        },
    });
    sw.fixSlices();

    var meta: Metadata = undefined;
    sw.getMetadata(&meta);

    try testing.expectEqualStrings("sw(0.500, pa(4, 0.200, 0.200), hl/2)", meta.mnemonic);
}

test "SW constructor errors" {
    // Alpha <= 0
    try testing.expectError(error.InvalidAlphaEmaPeriodAdditional, SineWave.init(.{
        .alpha_ema_period_additional = 0.0,
    }));
    // Alpha > 1
    try testing.expectError(error.InvalidAlphaEmaPeriodAdditional, SineWave.init(.{
        .alpha_ema_period_additional = 1.00000001,
    }));
}

test "SW updateScalar" {
    var sw = try SineWave.initDefault();
    sw.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const s = Scalar{ .time = @intCast(i), .value = testdata.test_input[i] };
        const out = sw.updateScalar(&s);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 5), outputs.len);
    }
}

test "SW band ordering" {
    var sw = try SineWave.initDefault();
    sw.fixSlices();

    // Prime the indicator.
    for (0..200) |i| {
        _ = sw.update(testdata.test_input[i % testdata.test_input.len]);
    }

    const s = Scalar{ .time = 0, .value = testdata.test_input[0] };
    const out = sw.updateScalar(&s);
    const outputs = out.slice();

    // Output 0 = value (scalar), output 1 = lead (scalar), output 2 = band.
    const value_val = outputs[0].scalar.value;
    const lead_val = outputs[1].scalar.value;
    const band_upper = outputs[2].band.upper;
    const band_lower = outputs[2].band.lower;

    try testing.expect(almostEqual(band_upper, value_val, 1e-15));
    try testing.expect(almostEqual(band_lower, lead_val, 1e-15));
}
