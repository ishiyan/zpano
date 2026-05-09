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

const dc_mod = @import("../dominant_cycle/dominant_cycle.zig");
const ht = @import("../hilbert_transformer/hilbert_transformer.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const DominantCycle = dc_mod.DominantCycle;
const CycleEstimatorType = ht.CycleEstimatorType;
const CycleEstimatorParams = ht.CycleEstimatorParams;

/// Enumerates the outputs of the indicator.
pub const HtitlOutput = enum(u8) {
    value = 1,
    dominant_cycle_period = 2,
};

/// Parameters to create a HilbertTransformerInstantaneousTrendLine indicator.
pub const Params = struct {
    estimator_type: CycleEstimatorType = .homodyne_discriminator,
    estimator_params: CycleEstimatorParams = .{},
    alpha_ema_period_additional: f64 = 0.33,
    trend_line_smoothing_length: u8 = 4,
    cycle_part_multiplier: f64 = 1.0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehlers' Hilbert Transformer Instantaneous Trend Line indicator.
///
/// Two outputs:
///   - Value: the instantaneous trend line value, computed as a WMA of simple averages
///     over windows whose length tracks the smoothed dominant cycle period.
///   - DominantCyclePeriod: the additionally EMA-smoothed dominant cycle period.
pub const HilbertTransformerInstantaneousTrendLine = struct {
    dc: DominantCycle,
    alpha: f64,
    one_min_alpha: f64,
    cycle_part_multiplier: f64,
    trend_line_smoothing_length: u8,
    coeff0: f64,
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    smoothed_period: f64,
    value: f64,
    average1: f64,
    average2: f64,
    average3: f64,
    input: [50]f64,
    input_length: usize,
    input_length_min1: usize,
    primed: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    // Mnemonic/description buffers for 2 outputs.
    mn_value_buf: [128]u8,
    mn_value_len: usize,
    mn_dcp_buf: [128]u8,
    mn_dcp_len: usize,
    desc_value_buf: [256]u8,
    desc_value_len: usize,
    desc_dcp_buf: [192]u8,
    desc_dcp_len: usize,

    pub const InitError = error{
        InvalidAlphaEmaPeriodAdditional,
        InvalidTrendLineSmoothingLength,
        InvalidCyclePartMultiplier,
    } || ht.VerifyError;

    /// Creates an instance with default parameters.
    pub fn initDefault() InitError!HilbertTransformerInstantaneousTrendLine {
        return init(.{
            .estimator_type = .homodyne_discriminator,
            .estimator_params = .{
                .smoothing_length = 4,
                .alpha_ema_quadrature_in_phase = 0.2,
                .alpha_ema_period = 0.2,
                .warm_up_period = 100,
            },
            .alpha_ema_period_additional = 0.33,
            .trend_line_smoothing_length = 4,
            .cycle_part_multiplier = 1.0,
        });
    }

    /// Creates an instance with the given parameters.
    pub fn init(params: Params) InitError!HilbertTransformerInstantaneousTrendLine {
        const alpha = params.alpha_ema_period_additional;
        if (alpha <= 0.0 or alpha > 1.0) {
            return InitError.InvalidAlphaEmaPeriodAdditional;
        }

        const tlsl = params.trend_line_smoothing_length;
        if (tlsl < 2 or tlsl > 4) {
            return InitError.InvalidTrendLineSmoothingLength;
        }

        const cpm = params.cycle_part_multiplier;
        if (cpm <= 0.0 or cpm > 10.0) {
            return InitError.InvalidCyclePartMultiplier;
        }

        // Default to BarMedianPrice (same as SineWave/TCM).
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

        // WMA coefficients.
        var c0: f64 = 0;
        var c1: f64 = 0;
        var c2: f64 = 0;
        var c3: f64 = 0;
        switch (tlsl) {
            2 => {
                c0 = 2.0 / 3.0;
                c1 = 1.0 / 3.0;
            },
            3 => {
                c0 = 3.0 / 6.0;
                c1 = 2.0 / 6.0;
                c2 = 1.0 / 6.0;
            },
            else => { // 4
                c0 = 4.0 / 10.0;
                c1 = 3.0 / 10.0;
                c2 = 2.0 / 10.0;
                c3 = 1.0 / 10.0;
            },
        }

        const max_period = dc.maxPeriod();

        // Build mnemonics: htitl(alpha, tlsl, cpm, est, comp) and dcp(alpha, est, comp).
        var mn_value_buf: [128]u8 = undefined;
        const mn_value = std.fmt.bufPrint(&mn_value_buf, "htitl({d:.3}, {d}, {d:.3}{s}{s})", .{
            alpha, tlsl, cpm, est_moniker, triple,
        }) catch return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_dcp_buf: [128]u8 = undefined;
        const mn_dcp = std.fmt.bufPrint(&mn_dcp_buf, "dcp({d:.3}{s}{s})", .{
            alpha, est_moniker, triple,
        }) catch return InitError.InvalidAlphaEmaPeriodAdditional;

        var desc_value_buf: [256]u8 = undefined;
        const desc_value = std.fmt.bufPrint(&desc_value_buf, "Hilbert transformer instantaneous trend line {s}", .{mn_value}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var desc_dcp_buf: [192]u8 = undefined;
        const desc_dcp = std.fmt.bufPrint(&desc_dcp_buf, "Dominant cycle period {s}", .{mn_dcp}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        return .{
            .dc = dc,
            .alpha = alpha,
            .one_min_alpha = 1.0 - alpha,
            .cycle_part_multiplier = cpm,
            .trend_line_smoothing_length = tlsl,
            .coeff0 = c0,
            .coeff1 = c1,
            .coeff2 = c2,
            .coeff3 = c3,
            .smoothed_period = 0,
            .value = math.nan(f64),
            .average1 = 0,
            .average2 = 0,
            .average3 = 0,
            .input = [_]f64{0} ** 50,
            .input_length = max_period,
            .input_length_min1 = max_period - 1,
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mn_value_buf = mn_value_buf,
            .mn_value_len = mn_value.len,
            .mn_dcp_buf = mn_dcp_buf,
            .mn_dcp_len = mn_dcp.len,
            .desc_value_buf = desc_value_buf,
            .desc_value_len = desc_value.len,
            .desc_dcp_buf = desc_dcp_buf,
            .desc_dcp_len = desc_dcp.len,
        };
    }

    /// Must be called after init to fix internal slice pointers.
    pub fn fixSlices(self: *HilbertTransformerInstantaneousTrendLine) void {
        self.dc.fixSlices();
    }

    /// Update the indicator given the next sample.
    /// Returns (value, period). Returns NaN if not yet primed.
    pub fn update(self: *HilbertTransformerInstantaneousTrendLine, sample: f64) [2]f64 {
        const nan = math.nan(f64);

        if (math.isNan(sample)) {
            return .{ nan, nan };
        }

        const dc_result = self.dc.update(sample);
        const dc_period = dc_result[1]; // Already EMA-smoothed by DominantCycle.
        self.pushInput(sample);

        if (self.primed) {
            self.smoothed_period = dc_period;
            const average = self.calculateAverage();
            self.value = self.coeff0 * average + self.coeff1 * self.average1 +
                self.coeff2 * self.average2 + self.coeff3 * self.average3;
            self.average3 = self.average2;
            self.average2 = self.average1;
            self.average1 = average;

            return .{ self.value, self.smoothed_period };
        }

        if (self.dc.isPrimed()) {
            self.primed = true;
            self.smoothed_period = dc_period;
            const average = self.calculateAverage();
            self.value = average;
            self.average1 = average;
            self.average2 = average;
            self.average3 = average;

            return .{ self.value, self.smoothed_period };
        }

        return .{ nan, nan };
    }

    pub fn isPrimed(self: *const HilbertTransformerInstantaneousTrendLine) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const HilbertTransformerInstantaneousTrendLine, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .hilbert_transformer_instantaneous_trend_line,
            self.mn_value_buf[0..self.mn_value_len],
            self.desc_value_buf[0..self.desc_value_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mn_value_buf[0..self.mn_value_len], .description = self.desc_value_buf[0..self.desc_value_len] },
                .{ .mnemonic = self.mn_dcp_buf[0..self.mn_dcp_len], .description = self.desc_dcp_buf[0..self.desc_dcp_len] },
            },
        );
    }

    pub fn updateScalar(self: *HilbertTransformerInstantaneousTrendLine, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *HilbertTransformerInstantaneousTrendLine, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *HilbertTransformerInstantaneousTrendLine, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *HilbertTransformerInstantaneousTrendLine, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *HilbertTransformerInstantaneousTrendLine, time: i64, sample: f64) OutputArray {
        const result = self.update(sample);
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = result[0] } }); // value
        out.append(.{ .scalar = .{ .time = time, .value = result[1] } }); // period
        return out;
    }

    fn pushInput(self: *HilbertTransformerInstantaneousTrendLine, val: f64) void {
        // Shift right by 1, newest at [0].
        var i: usize = self.input_length_min1;
        while (i > 0) : (i -= 1) {
            self.input[i] = self.input[i - 1];
        }
        self.input[0] = val;
    }

    fn calculateAverage(self: *const HilbertTransformerInstantaneousTrendLine) f64 {
        var length: usize = @intFromFloat(@floor(self.smoothed_period * self.cycle_part_multiplier + 0.5));
        if (length > self.input_length) {
            length = self.input_length;
        } else if (length < 1) {
            length = 1;
        }

        var sum: f64 = 0;
        for (0..length) |i| {
            sum += self.input[i];
        }

        return sum / @as(f64, @floatFromInt(length));
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *HilbertTransformerInstantaneousTrendLine) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(HilbertTransformerInstantaneousTrendLine);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tol: f64) bool {
    return @abs(a - b) <= tol;
}

// 252-entry input data from TA-Lib tests, test_MAMA.xsl.
// Expected period data, 252 entries, smoothed as AI18=0.33*X18+0.67*AI17.
// Expected instantaneous trend line values from MBST InstantaneousTrendLineTest.cs.
// 252 entries.
const tolerance = 1e-4;

test "HTITL update value" {
    const skip = 9;
    const settle_skip = 177;

    var ind = try HilbertTransformerInstantaneousTrendLine.initDefault();
    ind.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = ind.update(testdata.test_input[i]);
        const val = result[0];
        if (math.isNan(val) or i < settle_skip) {
            continue;
        }
        if (math.isNan(testdata.test_expected_value[i])) {
            continue;
        }
        try testing.expect(almostEqual(testdata.test_expected_value[i], val, tolerance));
    }
}

test "HTITL update period" {
    const skip = 9;
    const settle_skip = 177;

    var ind = try HilbertTransformerInstantaneousTrendLine.initDefault();
    ind.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = ind.update(testdata.test_input[i]);
        const period_val = result[1];
        if (math.isNan(period_val) or i < settle_skip) {
            continue;
        }
        try testing.expect(almostEqual(testdata.test_expected_period[i], period_val, tolerance));
    }
}

test "HTITL NaN input returns NaN pair" {
    var ind = try HilbertTransformerInstantaneousTrendLine.initDefault();
    ind.fixSlices();

    const result = ind.update(math.nan(f64));
    try testing.expect(math.isNan(result[0]));
    try testing.expect(math.isNan(result[1]));
}

test "HTITL isPrimed" {
    var ind = try HilbertTransformerInstantaneousTrendLine.initDefault();
    ind.fixSlices();

    try testing.expect(!ind.isPrimed());

    var primed_at: ?usize = null;
    for (0..testdata.test_input.len) |i| {
        _ = ind.update(testdata.test_input[i]);
        if (ind.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expect(primed_at != null);
    try testing.expect(ind.isPrimed());
}

test "HTITL metadata default" {
    var ind = try HilbertTransformerInstantaneousTrendLine.initDefault();
    ind.fixSlices();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.hilbert_transformer_instantaneous_trend_line, meta.identifier);
    try testing.expectEqualStrings("htitl(0.330, 4, 1.000, hl/2)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
}

test "HTITL metadata phase accumulator" {
    var ind = try HilbertTransformerInstantaneousTrendLine.init(.{
        .alpha_ema_period_additional = 0.5,
        .estimator_type = .phase_accumulator,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
        },
        .trend_line_smoothing_length = 3,
        .cycle_part_multiplier = 0.5,
    });
    ind.fixSlices();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("htitl(0.500, 3, 0.500, pa(4, 0.200, 0.200), hl/2)", meta.mnemonic);
}

test "HTITL metadata tlsl=2" {
    var ind = try HilbertTransformerInstantaneousTrendLine.init(.{
        .alpha_ema_period_additional = 0.33,
        .trend_line_smoothing_length = 2,
        .cycle_part_multiplier = 1.0,
        .bar_component = .close,
        .quote_component = .mid,
        .trade_component = .price,
    });
    ind.fixSlices();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("htitl(0.330, 2, 1.000)", meta.mnemonic);
}

test "HTITL constructor errors" {
    // Alpha <= 0
    try testing.expectError(error.InvalidAlphaEmaPeriodAdditional, HilbertTransformerInstantaneousTrendLine.init(.{
        .alpha_ema_period_additional = 0.0,
    }));
    // Alpha > 1
    try testing.expectError(error.InvalidAlphaEmaPeriodAdditional, HilbertTransformerInstantaneousTrendLine.init(.{
        .alpha_ema_period_additional = 1.00000001,
    }));
    // TLSL < 2
    try testing.expectError(error.InvalidTrendLineSmoothingLength, HilbertTransformerInstantaneousTrendLine.init(.{
        .alpha_ema_period_additional = 0.33,
        .trend_line_smoothing_length = 1,
    }));
    // TLSL > 4
    try testing.expectError(error.InvalidTrendLineSmoothingLength, HilbertTransformerInstantaneousTrendLine.init(.{
        .alpha_ema_period_additional = 0.33,
        .trend_line_smoothing_length = 5,
    }));
    // CPM <= 0
    try testing.expectError(error.InvalidCyclePartMultiplier, HilbertTransformerInstantaneousTrendLine.init(.{
        .alpha_ema_period_additional = 0.33,
        .cycle_part_multiplier = 0.0,
    }));
    // CPM > 10
    try testing.expectError(error.InvalidCyclePartMultiplier, HilbertTransformerInstantaneousTrendLine.init(.{
        .alpha_ema_period_additional = 0.33,
        .cycle_part_multiplier = 10.00001,
    }));
}

test "HTITL updateScalar" {
    var ind = try HilbertTransformerInstantaneousTrendLine.initDefault();
    ind.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const s = Scalar{ .time = @intCast(i), .value = testdata.test_input[i] };
        const out = ind.updateScalar(&s);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 2), outputs.len);
    }
}
