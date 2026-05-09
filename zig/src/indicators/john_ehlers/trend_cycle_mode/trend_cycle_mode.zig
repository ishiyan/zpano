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

const deg2rad = math.pi / 180.0;
const epsilon = 1e-308;

/// Enumerates the outputs of the trend cycle mode indicator.
pub const TrendCycleModeOutput = enum(u8) {
    value = 1,
    is_trend_mode = 2,
    is_cycle_mode = 3,
    instantaneous_trend_line = 4,
    sine_wave = 5,
    sine_wave_lead = 6,
    dominant_cycle_period = 7,
    dominant_cycle_phase = 8,
};

/// Parameters to create a TrendCycleMode indicator.
pub const Params = struct {
    estimator_type: CycleEstimatorType = .homodyne_discriminator,
    estimator_params: CycleEstimatorParams = .{},
    alpha_ema_period_additional: f64 = 0.33,
    trend_line_smoothing_length: u8 = 4,
    cycle_part_multiplier: f64 = 1.0,
    separation_percentage: f64 = 1.5,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehlers' Trend-versus-Cycle Mode indicator.
///
/// Eight outputs:
///   - Value: +1 in trend mode, -1 in cycle mode.
///   - IsTrendMode: 1 if trend mode, 0 otherwise.
///   - IsCycleMode: 1 if cycle mode, 0 otherwise.
///   - InstantaneousTrendLine: WMA-smoothed trend line.
///   - SineWave: sin(phase * deg2rad).
///   - SineWaveLead: sin((phase + 45) * deg2rad).
///   - DominantCyclePeriod: smoothed dominant cycle period.
///   - DominantCyclePhase: dominant cycle phase, in degrees.
pub const TrendCycleMode = struct {
    dc: DominantCycle,
    cycle_part_multiplier: f64,
    separation_factor: f64,
    trend_line_smoothing_length: u8,
    coeff0: f64,
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    trendline: f64,
    trend_average1: f64,
    trend_average2: f64,
    trend_average3: f64,
    sin_wave: f64,
    sin_wave_lead: f64,
    previous_dc_phase: f64,
    previous_sine_lead_wave_difference: f64,
    samples_in_trend: i32,
    is_trend_mode: bool,
    input: [50]f64,
    input_length: usize,
    input_length_min1: usize,
    primed: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    // Mnemonic/description buffers for 8 outputs.
    mn_value_buf: [192]u8,
    mn_value_len: usize,
    mn_trend_buf: [192]u8,
    mn_trend_len: usize,
    mn_cycle_buf: [192]u8,
    mn_cycle_len: usize,
    mn_itl_buf: [192]u8,
    mn_itl_len: usize,
    mn_sine_buf: [192]u8,
    mn_sine_len: usize,
    mn_sine_lead_buf: [192]u8,
    mn_sine_lead_len: usize,
    mn_dcp_buf: [128]u8,
    mn_dcp_len: usize,
    mn_dcph_buf: [128]u8,
    mn_dcph_len: usize,
    desc_value_buf: [256]u8,
    desc_value_len: usize,
    desc_trend_buf: [256]u8,
    desc_trend_len: usize,
    desc_cycle_buf: [256]u8,
    desc_cycle_len: usize,
    desc_itl_buf: [256]u8,
    desc_itl_len: usize,
    desc_sine_buf: [256]u8,
    desc_sine_len: usize,
    desc_sine_lead_buf: [256]u8,
    desc_sine_lead_len: usize,
    desc_dcp_buf: [256]u8,
    desc_dcp_len: usize,
    desc_dcph_buf: [256]u8,
    desc_dcph_len: usize,

    pub const InitError = error{
        InvalidAlphaEmaPeriodAdditional,
        InvalidTrendLineSmoothingLength,
        InvalidCyclePartMultiplier,
        InvalidSeparationPercentage,
    } || ht.VerifyError;

    /// Creates a TrendCycleMode with default parameters.
    pub fn initDefault() InitError!TrendCycleMode {
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
            .separation_percentage = 1.5,
        });
    }

    /// Creates a TrendCycleMode with the given parameters.
    pub fn init(params: Params) InitError!TrendCycleMode {
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

        const sep = params.separation_percentage;
        if (sep <= 0.0 or sep > 100.0) {
            return InitError.InvalidSeparationPercentage;
        }

        // Default to BarMedianPrice (not the framework default).
        const bc = params.bar_component orelse .median;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Build DominantCycle.
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
        const nan = math.nan(f64);

        // Build mnemonics for all 8 outputs.
        // TCM mnemonics: tcm(alpha, tlsl, cpm, sep%[, estimator][, triple])
        var mn_value_buf: [192]u8 = undefined;
        const mn_value = std.fmt.bufPrint(&mn_value_buf, "tcm({d:.3}, {d}, {d:.3}, {d:.3}%{s}{s})", .{ alpha, tlsl, cpm, sep, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_trend_buf: [192]u8 = undefined;
        const mn_trend = std.fmt.bufPrint(&mn_trend_buf, "tcm-trend({d:.3}, {d}, {d:.3}, {d:.3}%{s}{s})", .{ alpha, tlsl, cpm, sep, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_cycle_buf: [192]u8 = undefined;
        const mn_cycle = std.fmt.bufPrint(&mn_cycle_buf, "tcm-cycle({d:.3}, {d}, {d:.3}, {d:.3}%{s}{s})", .{ alpha, tlsl, cpm, sep, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_itl_buf: [192]u8 = undefined;
        const mn_itl = std.fmt.bufPrint(&mn_itl_buf, "tcm-itl({d:.3}, {d}, {d:.3}, {d:.3}%{s}{s})", .{ alpha, tlsl, cpm, sep, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_sine_buf: [192]u8 = undefined;
        const mn_sine = std.fmt.bufPrint(&mn_sine_buf, "tcm-sine({d:.3}, {d}, {d:.3}, {d:.3}%{s}{s})", .{ alpha, tlsl, cpm, sep, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_sine_lead_buf: [192]u8 = undefined;
        const mn_sine_lead = std.fmt.bufPrint(&mn_sine_lead_buf, "tcm-sineLead({d:.3}, {d}, {d:.3}, {d:.3}%{s}{s})", .{ alpha, tlsl, cpm, sep, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_dcp_buf: [128]u8 = undefined;
        const mn_dcp = std.fmt.bufPrint(&mn_dcp_buf, "dcp({d:.3}{s}{s})", .{ alpha, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var mn_dcph_buf: [128]u8 = undefined;
        const mn_dcph = std.fmt.bufPrint(&mn_dcph_buf, "dcph({d:.3}{s}{s})", .{ alpha, est_moniker, triple }) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        // Descriptions.
        var desc_value_buf: [256]u8 = undefined;
        const desc_value = std.fmt.bufPrint(&desc_value_buf, "Trend versus cycle mode {s}", .{mn_value}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        var desc_trend_buf: [256]u8 = undefined;
        const desc_trend = std.fmt.bufPrint(&desc_trend_buf, "Trend versus cycle mode, is-trend flag {s}", .{mn_trend}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        var desc_cycle_buf: [256]u8 = undefined;
        const desc_cycle = std.fmt.bufPrint(&desc_cycle_buf, "Trend versus cycle mode, is-cycle flag {s}", .{mn_cycle}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        var desc_itl_buf: [256]u8 = undefined;
        const desc_itl = std.fmt.bufPrint(&desc_itl_buf, "Trend versus cycle mode instantaneous trend line {s}", .{mn_itl}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        var desc_sine_buf: [256]u8 = undefined;
        const desc_sine = std.fmt.bufPrint(&desc_sine_buf, "Trend versus cycle mode sine wave {s}", .{mn_sine}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        var desc_sine_lead_buf: [256]u8 = undefined;
        const desc_sine_lead = std.fmt.bufPrint(&desc_sine_lead_buf, "Trend versus cycle mode sine wave lead {s}", .{mn_sine_lead}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        var desc_dcp_buf: [256]u8 = undefined;
        const desc_dcp = std.fmt.bufPrint(&desc_dcp_buf, "Dominant cycle period {s}", .{mn_dcp}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;
        var desc_dcph_buf: [256]u8 = undefined;
        const desc_dcph = std.fmt.bufPrint(&desc_dcph_buf, "Dominant cycle phase {s}", .{mn_dcph}) catch
            return InitError.InvalidAlphaEmaPeriodAdditional;

        var input_arr: [50]f64 = undefined;
        @memset(&input_arr, 0);

        return .{
            .dc = dc,
            .cycle_part_multiplier = cpm,
            .separation_factor = sep / 100.0,
            .trend_line_smoothing_length = tlsl,
            .coeff0 = c0,
            .coeff1 = c1,
            .coeff2 = c2,
            .coeff3 = c3,
            .trendline = nan,
            .trend_average1 = 0,
            .trend_average2 = 0,
            .trend_average3 = 0,
            .sin_wave = nan,
            .sin_wave_lead = nan,
            .previous_dc_phase = 0,
            .previous_sine_lead_wave_difference = 0,
            .samples_in_trend = 0,
            .is_trend_mode = true,
            .input = input_arr,
            .input_length = max_period,
            .input_length_min1 = max_period - 1,
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mn_value_buf = mn_value_buf,
            .mn_value_len = mn_value.len,
            .mn_trend_buf = mn_trend_buf,
            .mn_trend_len = mn_trend.len,
            .mn_cycle_buf = mn_cycle_buf,
            .mn_cycle_len = mn_cycle.len,
            .mn_itl_buf = mn_itl_buf,
            .mn_itl_len = mn_itl.len,
            .mn_sine_buf = mn_sine_buf,
            .mn_sine_len = mn_sine.len,
            .mn_sine_lead_buf = mn_sine_lead_buf,
            .mn_sine_lead_len = mn_sine_lead.len,
            .mn_dcp_buf = mn_dcp_buf,
            .mn_dcp_len = mn_dcp.len,
            .mn_dcph_buf = mn_dcph_buf,
            .mn_dcph_len = mn_dcph.len,
            .desc_value_buf = desc_value_buf,
            .desc_value_len = desc_value.len,
            .desc_trend_buf = desc_trend_buf,
            .desc_trend_len = desc_trend.len,
            .desc_cycle_buf = desc_cycle_buf,
            .desc_cycle_len = desc_cycle.len,
            .desc_itl_buf = desc_itl_buf,
            .desc_itl_len = desc_itl.len,
            .desc_sine_buf = desc_sine_buf,
            .desc_sine_len = desc_sine.len,
            .desc_sine_lead_buf = desc_sine_lead_buf,
            .desc_sine_lead_len = desc_sine_lead.len,
            .desc_dcp_buf = desc_dcp_buf,
            .desc_dcp_len = desc_dcp.len,
            .desc_dcph_buf = desc_dcph_buf,
            .desc_dcph_len = desc_dcph.len,
        };
    }

    /// Must be called after init to fix internal slice pointers.
    pub fn fixSlices(self: *TrendCycleMode) void {
        self.dc.fixSlices();
    }

    /// Update the indicator given the next sample.
    /// Returns (value, isTrend, isCycle, trendline, sine, sineLead, period, phase).
    /// Returns NaN for all outputs if not yet primed.
    pub fn update(self: *TrendCycleMode, sample: f64) [8]f64 {
        const nan = math.nan(f64);

        if (math.isNan(sample)) {
            return .{ nan, nan, nan, nan, nan, nan, nan, nan };
        }

        const dc_result = self.dc.update(sample);
        const period = dc_result[1];
        const phase = dc_result[2];
        const smoothed_price = self.dc.smoothedPrice();

        self.pushInput(sample);

        if (self.primed) {
            const smoothed_period = period;
            const average = self.calculateTrendAverage(smoothed_period);
            self.trendline = self.coeff0 * average + self.coeff1 * self.trend_average1 +
                self.coeff2 * self.trend_average2 + self.coeff3 * self.trend_average3;
            self.trend_average3 = self.trend_average2;
            self.trend_average2 = self.trend_average1;
            self.trend_average1 = average;

            const diff = self.calculateSineLeadWaveDifference(phase);

            // Condition 1: cycle mode for half-period after crossing.
            self.is_trend_mode = true;

            if ((diff > 0 and self.previous_sine_lead_wave_difference < 0) or
                (diff < 0 and self.previous_sine_lead_wave_difference > 0))
            {
                self.is_trend_mode = false;
                self.samples_in_trend = 0;
            }

            self.previous_sine_lead_wave_difference = diff;
            self.samples_in_trend += 1;

            if (@as(f64, @floatFromInt(self.samples_in_trend)) < 0.5 * smoothed_period) {
                self.is_trend_mode = false;
            }

            // Condition 2: cycle mode if phase rate is in [2/3, 1.5] of DC rate.
            const phase_delta = phase - self.previous_dc_phase;
            self.previous_dc_phase = phase;

            if (@abs(smoothed_period) > epsilon) {
                const dc_rate = 360.0 / smoothed_period;
                if (phase_delta > (2.0 / 3.0) * dc_rate and phase_delta < 1.5 * dc_rate) {
                    self.is_trend_mode = false;
                }
            }

            // Condition 3: separation override to trend.
            if (@abs(self.trendline) > epsilon and
                @abs((smoothed_price - self.trendline) / self.trendline) >= self.separation_factor)
            {
                self.is_trend_mode = true;
            }

            return .{
                self.mode(),    self.isTrendFloat(), self.isCycleFloat(),
                self.trendline, self.sin_wave,       self.sin_wave_lead,
                period,         phase,
            };
        }

        if (self.dc.isPrimed()) {
            self.primed = true;
            const smoothed_period = period;
            self.trendline = self.calculateTrendAverage(smoothed_period);
            self.trend_average1 = self.trendline;
            self.trend_average2 = self.trendline;
            self.trend_average3 = self.trendline;

            self.previous_dc_phase = phase;
            self.previous_sine_lead_wave_difference = self.calculateSineLeadWaveDifference(phase);

            self.is_trend_mode = true;
            self.samples_in_trend += 1;

            if (@as(f64, @floatFromInt(self.samples_in_trend)) < 0.5 * smoothed_period) {
                self.is_trend_mode = false;
            }

            return .{
                self.mode(),    self.isTrendFloat(), self.isCycleFloat(),
                self.trendline, self.sin_wave,       self.sin_wave_lead,
                period,         phase,
            };
        }

        return .{ nan, nan, nan, nan, nan, nan, nan, nan };
    }

    pub fn isPrimed(self: *const TrendCycleMode) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const TrendCycleMode, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .trend_cycle_mode,
            self.mn_value_buf[0..self.mn_value_len],
            self.desc_value_buf[0..self.desc_value_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mn_value_buf[0..self.mn_value_len], .description = self.desc_value_buf[0..self.desc_value_len] },
                .{ .mnemonic = self.mn_trend_buf[0..self.mn_trend_len], .description = self.desc_trend_buf[0..self.desc_trend_len] },
                .{ .mnemonic = self.mn_cycle_buf[0..self.mn_cycle_len], .description = self.desc_cycle_buf[0..self.desc_cycle_len] },
                .{ .mnemonic = self.mn_itl_buf[0..self.mn_itl_len], .description = self.desc_itl_buf[0..self.desc_itl_len] },
                .{ .mnemonic = self.mn_sine_buf[0..self.mn_sine_len], .description = self.desc_sine_buf[0..self.desc_sine_len] },
                .{ .mnemonic = self.mn_sine_lead_buf[0..self.mn_sine_lead_len], .description = self.desc_sine_lead_buf[0..self.desc_sine_lead_len] },
                .{ .mnemonic = self.mn_dcp_buf[0..self.mn_dcp_len], .description = self.desc_dcp_buf[0..self.desc_dcp_len] },
                .{ .mnemonic = self.mn_dcph_buf[0..self.mn_dcph_len], .description = self.desc_dcph_buf[0..self.desc_dcph_len] },
            },
        );
    }

    pub fn updateScalar(self: *TrendCycleMode, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *TrendCycleMode, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *TrendCycleMode, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *TrendCycleMode, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *TrendCycleMode, time: i64, sample: f64) OutputArray {
        const result = self.update(sample);
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = result[0] } }); // value
        out.append(.{ .scalar = .{ .time = time, .value = result[1] } }); // isTrend
        out.append(.{ .scalar = .{ .time = time, .value = result[2] } }); // isCycle
        out.append(.{ .scalar = .{ .time = time, .value = result[3] } }); // ITL
        out.append(.{ .scalar = .{ .time = time, .value = result[4] } }); // sine
        out.append(.{ .scalar = .{ .time = time, .value = result[5] } }); // sineLead
        out.append(.{ .scalar = .{ .time = time, .value = result[6] } }); // period
        out.append(.{ .scalar = .{ .time = time, .value = result[7] } }); // phase
        return out;
    }

    fn pushInput(self: *TrendCycleMode, value: f64) void {
        // Shift right by 1, newest at [0].
        var i: usize = self.input_length_min1;
        while (i > 0) : (i -= 1) {
            self.input[i] = self.input[i - 1];
        }
        self.input[0] = value;
    }

    fn calculateTrendAverage(self: *const TrendCycleMode, smoothed_period: f64) f64 {
        var length: usize = @intFromFloat(@floor(smoothed_period * self.cycle_part_multiplier + 0.5));
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

    fn calculateSineLeadWaveDifference(self: *TrendCycleMode, phase: f64) f64 {
        const lead_offset = 45.0;
        const p = phase * deg2rad;
        self.sin_wave = @sin(p);
        self.sin_wave_lead = @sin(p + lead_offset * deg2rad);
        return self.sin_wave - self.sin_wave_lead;
    }

    fn mode(self: *const TrendCycleMode) f64 {
        if (self.is_trend_mode) return 1.0;
        return -1.0;
    }

    fn isTrendFloat(self: *const TrendCycleMode) f64 {
        if (self.is_trend_mode) return 1.0;
        return 0.0;
    }

    fn isCycleFloat(self: *const TrendCycleMode) f64 {
        if (self.is_trend_mode) return 0.0;
        return 1.0;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *TrendCycleMode) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(TrendCycleMode);
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

// 252-entry input data.
// Expected period data, 252 entries.
// Expected phase data, 252 entries.
// Expected sine wave values, 252 entries.
// Expected sine wave lead values, 252 entries.
// Expected ITL data, 252 entries.
// Expected value data, 252 entries (first 63 are NaN).
const tolerance = 1e-4;

test "TCM update period" {
    const skip = 9;
    const settle_skip = 177;

    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = tcm.update(testdata.test_input[i]);
        const period = result[6];
        if (math.isNan(period) or i < settle_skip) continue;
        try testing.expect(almostEqual(testdata.test_expected_period[i], period, tolerance));
    }
}

test "TCM update phase" {
    const skip = 9;
    const settle_skip = 177;

    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = tcm.update(testdata.test_input[i]);
        const phase = result[7];
        if (math.isNan(phase) or i < settle_skip) continue;
        if (math.isNan(testdata.test_expected_phase[i])) continue;
        try testing.expect(@abs(phaseDiff(testdata.test_expected_phase[i], phase)) <= tolerance);
    }
}

test "TCM update sine" {
    const skip = 9;
    const settle_skip = 177;

    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = tcm.update(testdata.test_input[i]);
        const sine = result[4];
        if (math.isNan(sine) or i < settle_skip) continue;
        if (math.isNan(testdata.test_expected_sine[i])) continue;
        try testing.expect(almostEqual(testdata.test_expected_sine[i], sine, tolerance));
    }
}

test "TCM update sine lead" {
    const skip = 9;
    const settle_skip = 177;

    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = tcm.update(testdata.test_input[i]);
        const sine_lead = result[5];
        if (math.isNan(sine_lead) or i < settle_skip) continue;
        if (math.isNan(testdata.test_expected_sine_lead[i])) continue;
        try testing.expect(almostEqual(testdata.test_expected_sine_lead[i], sine_lead, tolerance));
    }
}

test "TCM update ITL" {
    const skip = 9;
    const settle_skip = 177;

    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = tcm.update(testdata.test_input[i]);
        const itl = result[3];
        if (math.isNan(itl) or i < settle_skip) continue;
        if (math.isNan(testdata.test_expected_itl[i])) continue;
        try testing.expect(almostEqual(testdata.test_expected_itl[i], itl, tolerance));
    }
}

test "TCM update value" {
    const skip = 9;

    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = tcm.update(testdata.test_input[i]);
        const value = result[0];
        if (i >= testdata.test_expected_value.len) continue;
        // MBST known mismatches.
        if (i == 70 or i == 71) continue;
        if (math.isNan(value) or math.isNan(testdata.test_expected_value[i])) continue;
        try testing.expect(almostEqual(testdata.test_expected_value[i], value, tolerance));
    }
}

test "TCM trend+cycle complementary" {
    const skip = 9;

    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    for (skip..testdata.test_input.len) |i| {
        const result = tcm.update(testdata.test_input[i]);
        const value = result[0];
        const trend = result[1];
        const cycle = result[2];
        if (math.isNan(value)) continue;
        try testing.expect(almostEqual(trend + cycle, 1.0, 1e-15));
        if (value > 0) {
            try testing.expect(trend == 1.0);
        } else {
            try testing.expect(trend == 0.0);
        }
    }
}

test "TCM NaN input returns NaN tuple" {
    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    const result = tcm.update(math.nan(f64));
    for (0..8) |i| {
        try testing.expect(math.isNan(result[i]));
    }
}

test "TCM isPrimed" {
    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    try testing.expect(!tcm.isPrimed());

    var primed_at: ?usize = null;
    for (0..testdata.test_input.len) |i| {
        _ = tcm.update(testdata.test_input[i]);
        if (tcm.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expect(primed_at != null);
    try testing.expect(tcm.isPrimed());
}

test "TCM metadata default" {
    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    var meta: Metadata = undefined;
    tcm.getMetadata(&meta);

    try testing.expectEqual(Identifier.trend_cycle_mode, meta.identifier);
    try testing.expectEqualStrings("tcm(0.330, 4, 1.000, 1.500%, hl/2)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 8), meta.outputs_len);
}

test "TCM metadata phase accumulator" {
    var tcm = try TrendCycleMode.init(.{
        .alpha_ema_period_additional = 0.5,
        .estimator_type = .phase_accumulator,
        .estimator_params = .{
            .smoothing_length = 4,
            .alpha_ema_quadrature_in_phase = 0.2,
            .alpha_ema_period = 0.2,
        },
        .trend_line_smoothing_length = 3,
        .cycle_part_multiplier = 0.5,
        .separation_percentage = 2.0,
    });
    tcm.fixSlices();

    var meta: Metadata = undefined;
    tcm.getMetadata(&meta);

    try testing.expectEqualStrings("tcm(0.500, 3, 0.500, 2.000%, pa(4, 0.200, 0.200), hl/2)", meta.mnemonic);
}

test "TCM constructor errors" {
    // Alpha <= 0
    try testing.expectError(error.InvalidAlphaEmaPeriodAdditional, TrendCycleMode.init(.{
        .alpha_ema_period_additional = 0.0,
    }));
    // Alpha > 1
    try testing.expectError(error.InvalidAlphaEmaPeriodAdditional, TrendCycleMode.init(.{
        .alpha_ema_period_additional = 1.00000001,
    }));
    // TLSL < 2
    try testing.expectError(error.InvalidTrendLineSmoothingLength, TrendCycleMode.init(.{
        .alpha_ema_period_additional = 0.33,
        .trend_line_smoothing_length = 1,
    }));
    // TLSL > 4
    try testing.expectError(error.InvalidTrendLineSmoothingLength, TrendCycleMode.init(.{
        .alpha_ema_period_additional = 0.33,
        .trend_line_smoothing_length = 5,
    }));
    // CPM <= 0
    try testing.expectError(error.InvalidCyclePartMultiplier, TrendCycleMode.init(.{
        .alpha_ema_period_additional = 0.33,
        .cycle_part_multiplier = 0.0,
    }));
    // CPM > 10
    try testing.expectError(error.InvalidCyclePartMultiplier, TrendCycleMode.init(.{
        .alpha_ema_period_additional = 0.33,
        .cycle_part_multiplier = 10.00001,
    }));
    // Sep <= 0
    try testing.expectError(error.InvalidSeparationPercentage, TrendCycleMode.init(.{
        .alpha_ema_period_additional = 0.33,
        .separation_percentage = 0.0,
    }));
    // Sep > 100
    try testing.expectError(error.InvalidSeparationPercentage, TrendCycleMode.init(.{
        .alpha_ema_period_additional = 0.33,
        .separation_percentage = 100.00001,
    }));
}

test "TCM updateScalar" {
    var tcm = try TrendCycleMode.initDefault();
    tcm.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const s = Scalar{ .time = @intCast(i), .value = testdata.test_input[i] };
        const out = tcm.updateScalar(&s);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 8), outputs.len);
    }
}
