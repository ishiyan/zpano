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
const corona_mod = @import("../corona/corona.zig");

const OutputArray = indicator_mod.OutputArray;
const OutputValue = indicator_mod.OutputValue;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Heatmap = heatmap_mod.Heatmap;
const Corona = corona_mod.Corona;

// Constants.
const bp_delta = 0.1;
const ratio_new_coef = 0.33;
const ratio_previous_coef = 0.67;
const vigor_mid_low = 0.3;
const vigor_mid_high = 0.7;
const vigor_mid = 0.5;
const width_edge = 0.01;
const raster_blend_scale = 0.8;
const raster_blend_previous = 0.2;
const raster_blend_half = 0.5;
const raster_blend_exponent = 0.85;
const ratio_limit = 10.0;
const vigor_scale = 0.05;

/// Enumerates the outputs of the CoronaTrendVigor indicator.
pub const CoronaTrendVigorOutput = enum(u8) {
    value = 1,
    trend_vigor = 2,
};

/// Parameters to create a CoronaTrendVigor indicator.
pub const Params = struct {
    /// Number of elements in the heatmap raster. Default: 50.
    raster_length: i32 = 0,
    /// Maximal raster value (z) of the heatmap. Default: 20.
    max_raster_value: f64 = 0,
    /// Minimal ordinate (y) value. Default: -10.
    min_parameter_value: f64 = 0,
    /// Maximal ordinate (y) value. Default: 10.
    max_parameter_value: f64 = 0,
    /// High-pass filter cutoff. Default: 30.
    high_pass_filter_cutoff: i32 = 0,
    /// Minimal period of the inner Corona engine. Default: 6.
    minimal_period: i32 = 0,
    /// Maximal period of the inner Corona engine. Default: 30.
    maximal_period: i32 = 0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Error = error{
    InvalidRasterLength,
    InvalidMaxRasterValue,
    InvalidMaxParameterValue,
    InvalidHighPassFilterCutoff,
    InvalidMinimalPeriod,
    InvalidMaximalPeriod,
    MnemonicTooLong,
    OutOfMemory,
};

/// Ehlers' Corona Trend Vigor heatmap indicator.
pub const CoronaTrendVigor = struct {
    allocator: std.mem.Allocator,
    corona: Corona,
    raster_length: usize,
    raster_step: f64,
    max_raster_value: f64,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    raster: []f64,
    sample_buffer: []f64,
    sample_count: usize,
    sample_previous: f64,
    sample_previous2: f64,
    band_pass_previous: f64,
    band_pass_previous2: f64,
    ratio_previous: f64,
    trend_vigor: f64,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_buf: [256]u8,
    mnemonic_len: usize,
    description_buf: [320]u8,
    description_len: usize,
    mnemonic_tv_buf: [128]u8,
    mnemonic_tv_len: usize,
    description_tv_buf: [256]u8,
    description_tv_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!CoronaTrendVigor {
        const def_raster_length: i32 = 50;
        const def_max_raster = 20.0;
        const def_min_param = -10.0;
        const def_max_param = 10.0;
        const def_hp_cutoff: i32 = 30;
        const def_min_period: i32 = 6;
        const def_max_period: i32 = 30;

        var raster_length = params.raster_length;
        if (raster_length == 0) raster_length = def_raster_length;

        var max_raster = params.max_raster_value;
        if (max_raster == 0) max_raster = def_max_raster;

        var min_param = params.min_parameter_value;
        var max_param = params.max_parameter_value;
        // "Both zero" default: only substitute when both are zero.
        if (min_param == 0 and max_param == 0) {
            min_param = def_min_param;
            max_param = def_max_param;
        }

        var hp_cutoff = params.high_pass_filter_cutoff;
        if (hp_cutoff == 0) hp_cutoff = def_hp_cutoff;

        var min_period = params.minimal_period;
        if (min_period == 0) min_period = def_min_period;

        var max_period = params.maximal_period;
        if (max_period == 0) max_period = def_max_period;

        // Validation.
        if (raster_length < 2) return error.InvalidRasterLength;
        if (max_raster <= 0) return error.InvalidMaxRasterValue;
        if (max_param <= min_param) return error.InvalidMaxParameterValue;
        if (hp_cutoff < 2) return error.InvalidHighPassFilterCutoff;
        if (min_period < 2) return error.InvalidMinimalPeriod;
        if (max_period <= min_period) return error.InvalidMaximalPeriod;

        const bc = params.bar_component orelse bar_component.BarComponent.median;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var corona = Corona.init(allocator, .{
            .high_pass_filter_cutoff = hp_cutoff,
            .minimal_period = min_period,
            .maximal_period = max_period,
        }) catch return error.OutOfMemory;
        errdefer corona.deinit();

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc,
            qc,
            tc,
        );

        const raster_len_usize: usize = @intCast(raster_length);
        const parameter_resolution = @as(f64, @floatFromInt(raster_length - 1)) / (max_param - min_param);
        const raster_step_val = max_raster / @as(f64, @floatFromInt(raster_length));

        // Build mnemonics.
        var mnemonic_buf: [256]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "ctv({d}, {d}, {d}, {d}, {d}{s})", .{
            raster_length,
            max_raster,
            min_param,
            max_param,
            hp_cutoff,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Corona trend vigor {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_tv_buf: [128]u8 = undefined;
        const mn_tv = std.fmt.bufPrint(&mnemonic_tv_buf, "ctv-tv({d}{s})", .{
            hp_cutoff,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_tv_len = mn_tv.len;

        var description_tv_buf: [256]u8 = undefined;
        const desc_tv = std.fmt.bufPrint(&description_tv_buf, "Corona trend vigor scalar {s}", .{mn_tv}) catch
            return error.MnemonicTooLong;
        const description_tv_len = desc_tv.len;

        // Allocate raster and sample buffer.
        const raster = allocator.alloc(f64, raster_len_usize) catch return error.OutOfMemory;
        @memset(raster, 0);

        const buf_size: usize = @intCast(corona.getMaximalPeriodTimesTwo());
        const sample_buffer = allocator.alloc(f64, buf_size) catch {
            allocator.free(raster);
            return error.OutOfMemory;
        };
        @memset(sample_buffer, 0);

        return .{
            .allocator = allocator,
            .corona = corona,
            .raster_length = raster_len_usize,
            .raster_step = raster_step_val,
            .max_raster_value = max_raster,
            .min_parameter_value = min_param,
            .max_parameter_value = max_param,
            .parameter_resolution = parameter_resolution,
            .raster = raster,
            .sample_buffer = sample_buffer,
            .sample_count = 0,
            .sample_previous = 0,
            .sample_previous2 = 0,
            .band_pass_previous = 0,
            .band_pass_previous2 = 0,
            .ratio_previous = 0,
            .trend_vigor = math.nan(f64),
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .mnemonic_tv_buf = mnemonic_tv_buf,
            .mnemonic_tv_len = mnemonic_tv_len,
            .description_tv_buf = description_tv_buf,
            .description_tv_len = description_tv_len,
        };
    }

    pub fn deinit(self: *CoronaTrendVigor) void {
        self.allocator.free(self.raster);
        self.allocator.free(self.sample_buffer);
        self.corona.deinit();
    }

    pub fn fixSlices(self: *CoronaTrendVigor) void {
        _ = self;
    }

    fn mnemonic(self: *const CoronaTrendVigor) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const CoronaTrendVigor) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    fn mnemonicTV(self: *const CoronaTrendVigor) []const u8 {
        return self.mnemonic_tv_buf[0..self.mnemonic_tv_len];
    }

    fn descriptionTV(self: *const CoronaTrendVigor) []const u8 {
        return self.description_tv_buf[0..self.description_tv_len];
    }

    /// Update with a new sample. Returns heatmap and trend vigor scalar.
    pub fn updateSample(self: *CoronaTrendVigor, sample: f64, time: i64) struct { heatmap: Heatmap, tv: f64 } {
        if (math.isNan(sample)) {
            return .{
                .heatmap = Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                .tv = math.nan(f64),
            };
        }

        const primed = self.corona.update(sample);
        self.sample_count += 1;

        const buf_last = self.sample_buffer.len - 1;

        if (self.sample_count == 1) {
            self.sample_previous = sample;
            self.sample_buffer[buf_last] = sample;
            return .{
                .heatmap = Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                .tv = math.nan(f64),
            };
        }

        // Bandpass InPhase filter at the dominant cycle median period.
        const omega = 2.0 * math.pi / self.corona.getDominantCycleMedian();
        const beta2 = math.cos(omega);
        const gamma2 = 1.0 / math.cos(omega * 2.0 * bp_delta);
        const alpha2 = gamma2 - @sqrt(gamma2 * gamma2 - 1.0);
        const band_pass = 0.5 * (1.0 - alpha2) * (sample - self.sample_previous2) +
            beta2 * (1.0 + alpha2) * self.band_pass_previous -
            alpha2 * self.band_pass_previous2;

        // Quadrature = derivative / omega.
        const quadrature2 = (band_pass - self.band_pass_previous) / omega;

        self.band_pass_previous2 = self.band_pass_previous;
        self.band_pass_previous = band_pass;
        self.sample_previous2 = self.sample_previous;
        self.sample_previous = sample;

        // Left-shift sample buffer and append new sample.
        for (0..buf_last) |i| {
            self.sample_buffer[i] = self.sample_buffer[i + 1];
        }
        self.sample_buffer[buf_last] = sample;

        // Cycle amplitude.
        const amplitude2 = @sqrt(band_pass * band_pass + quadrature2 * quadrature2);

        // Trend amplitude taken over the cycle period.
        const dcm_raw = self.corona.getDominantCycleMedian() - 1.0;
        const buf_len_i64 = @as(i64, @intCast(self.sample_buffer.len));
        var cycle_period: i64 = 1;
        if (math.isFinite(dcm_raw) and dcm_raw >= 1.0) {
            if (dcm_raw > @as(f64, @floatFromInt(buf_len_i64))) {
                cycle_period = buf_len_i64;
            } else {
                cycle_period = @intFromFloat(dcm_raw);
                if (cycle_period > buf_len_i64) {
                    cycle_period = buf_len_i64;
                }
            }
        }

        var lookback: usize = @intCast(cycle_period);
        if (self.sample_count < lookback) {
            lookback = self.sample_count;
        }

        const trend = sample - self.sample_buffer[self.sample_buffer.len - lookback];

        var ratio: f64 = 0.0;
        if (@abs(trend) > 0 and amplitude2 > 0) {
            ratio = ratio_new_coef * trend / amplitude2 + ratio_previous_coef * self.ratio_previous;
        }

        if (ratio > ratio_limit) {
            ratio = ratio_limit;
        } else if (ratio < -ratio_limit) {
            ratio = -ratio_limit;
        }

        self.ratio_previous = ratio;

        // ratio in [-10, 10] => vigor in [0, 1].
        const vigor = vigor_scale * (ratio + ratio_limit);

        var width: f64 = undefined;
        if (vigor >= vigor_mid_low and vigor < vigor_mid) {
            width = vigor - (vigor_mid_low - width_edge);
        } else if (vigor >= vigor_mid and vigor <= vigor_mid_high) {
            width = (vigor_mid_high + width_edge) - vigor;
        } else {
            width = width_edge;
        }

        self.trend_vigor = (self.max_parameter_value - self.min_parameter_value) * vigor + self.min_parameter_value;

        const vigor_scaled_to_raster_length: i64 = @intFromFloat(@round(@as(f64, @floatFromInt(self.raster_length)) * vigor));
        const vigor_scaled_to_max_raster_value = vigor * self.max_raster_value;

        for (0..self.raster_length) |i| {
            var value = self.raster[i];

            if (@as(i64, @intCast(i)) == vigor_scaled_to_raster_length) {
                value *= raster_blend_half;
            } else {
                var argument = vigor_scaled_to_max_raster_value - self.raster_step * @as(f64, @floatFromInt(i));
                if (@as(i64, @intCast(i)) > vigor_scaled_to_raster_length) {
                    argument = -argument;
                }

                if (width > 0) {
                    value = raster_blend_scale *
                        (std.math.pow(f64, argument / width, raster_blend_exponent) + raster_blend_previous * value);
                }
            }

            if (value < 0) {
                value = 0;
            } else if (value > self.max_raster_value or vigor < vigor_mid_low or vigor > vigor_mid_high) {
                value = self.max_raster_value;
            }

            if (math.isNan(value)) {
                value = 0;
            }

            self.raster[i] = value;
        }

        if (!primed) {
            return .{
                .heatmap = Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                .tv = math.nan(f64),
            };
        }

        // Build heatmap from raster.
        var values: [heatmap_mod.max_heatmap_values]f64 = undefined;
        var value_min: f64 = math.inf(f64);
        var value_max: f64 = -math.inf(f64);

        for (0..self.raster_length) |i| {
            const v = self.raster[i];
            values[i] = v;
            if (v < value_min) value_min = v;
            if (v > value_max) value_max = v;
        }

        return .{
            .heatmap = Heatmap.new(
                time,
                self.min_parameter_value,
                self.max_parameter_value,
                self.parameter_resolution,
                value_min,
                value_max,
                values[0..self.raster_length],
            ),
            .tv = self.trend_vigor,
        };
    }

    // --- Entity update methods ---

    pub fn updateBar(self: *CoronaTrendVigor, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *CoronaTrendVigor, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *CoronaTrendVigor, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    pub fn updateScalar(self: *CoronaTrendVigor, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    fn updateEntity(self: *CoronaTrendVigor, time: i64, sample: f64) OutputArray {
        const result = self.updateSample(sample, time);
        var out = OutputArray{};
        out.append(.{ .heatmap = result.heatmap });
        out.append(.{ .scalar = .{ .time = time, .value = result.tv } });
        return out;
    }

    pub fn isPrimed(self: *const CoronaTrendVigor) bool {
        return self.corona.isPrimed();
    }

    pub fn getMetadata(self: *const CoronaTrendVigor, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
            .{ .mnemonic = self.mnemonicTV(), .description = self.descriptionTV() },

        };
        build_metadata_mod.buildMetadata(out, .corona_trend_vigor, mn, desc, &texts);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *CoronaTrendVigor) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(CoronaTrendVigor);
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

const tolerance = 1e-4;

fn almostEqual(a: f64, b: f64, eps: f64) bool {
    return @abs(a - b) <= eps;
}

test "CoronaTrendVigor update" {
    const Snap = struct { i: usize, tv: f64, vmin: f64, vmax: f64 };
    const snapshots = [_]Snap{
        .{ .i = 11, .tv = 5.6512200755, .vmin = 20.0, .vmax = 20.0 },
        .{ .i = 12, .tv = 6.8379492897, .vmin = 20.0, .vmax = 20.0 },
        .{ .i = 50, .tv = 2.6145116709, .vmin = 2.3773561485, .vmax = 20.0 },
        .{ .i = 100, .tv = 2.7536803664, .vmin = 2.4892742850, .vmax = 20.0 },
        .{ .i = 150, .tv = -6.4606404251, .vmin = 20.0, .vmax = 20.0 },
        .{ .i = 200, .tv = -10.0, .vmin = 20.0, .vmax = 20.0 },
        .{ .i = 251, .tv = -0.1894989954, .vmin = 0.5847573715, .vmax = 20.0 },
    };

    var x = try CoronaTrendVigor.init(testing.allocator, .{});
    defer x.deinit();

    var si: usize = 0;
    for (0..test_input.len) |i| {
        const result = x.updateSample(test_input[i], @intCast(i));

        try testing.expectEqual(@as(f64, -10.0), result.heatmap.parameter_first);
        try testing.expectEqual(@as(f64, 10.0), result.heatmap.parameter_last);
        try testing.expect(almostEqual(2.45, result.heatmap.parameter_resolution, 1e-9));

        if (!x.isPrimed()) {
            try testing.expect(result.heatmap.isEmpty());
            try testing.expect(math.isNan(result.tv));
            continue;
        }

        try testing.expectEqual(@as(usize, 50), result.heatmap.values_len);

        if (si < snapshots.len and snapshots[si].i == i) {
            try testing.expect(almostEqual(snapshots[si].tv, result.tv, tolerance));
            try testing.expect(almostEqual(snapshots[si].vmin, result.heatmap.value_min, tolerance));
            try testing.expect(almostEqual(snapshots[si].vmax, result.heatmap.value_max, tolerance));
            si += 1;
        }
    }

    try testing.expectEqual(snapshots.len, si);
}

test "CoronaTrendVigor primes at bar 11" {
    var x = try CoronaTrendVigor.init(testing.allocator, .{});
    defer x.deinit();

    try testing.expect(!x.isPrimed());

    var primed_at: ?usize = null;
    for (0..test_input.len) |i| {
        _ = x.updateSample(test_input[i], @intCast(i));
        if (x.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expectEqual(@as(usize, 11), primed_at.?);
}

test "CoronaTrendVigor NaN input" {
    var x = try CoronaTrendVigor.init(testing.allocator, .{});
    defer x.deinit();

    const result = x.updateSample(math.nan(f64), 0);
    try testing.expect(result.heatmap.isEmpty());
    try testing.expect(math.isNan(result.tv));
    try testing.expect(!x.isPrimed());
}

test "CoronaTrendVigor metadata" {
    var x = try CoronaTrendVigor.init(testing.allocator, .{});
    defer x.deinit();

    var md: Metadata = undefined;
    x.getMetadata(&md);

    const mn_value = "ctv(50, 20, -10, 10, 30, hl/2)";
    const mn_tv = "ctv-tv(30, hl/2)";

    try testing.expectEqualStrings(mn_value, x.mnemonic());
    try testing.expectEqual(Identifier.corona_trend_vigor, md.identifier);
    try testing.expectEqualStrings(mn_value, md.mnemonic);
    try testing.expectEqual(@as(usize, 2), md.outputs_len);

    const outputs = md.outputs_buf[0..md.outputs_len];
    try testing.expectEqualStrings(mn_value, outputs[0].mnemonic);
    try testing.expectEqualStrings(mn_tv, outputs[1].mnemonic);
}

test "CoronaTrendVigor validation" {
    try testing.expectError(error.InvalidRasterLength, CoronaTrendVigor.init(testing.allocator, .{
        .raster_length = 1,
    }));
    try testing.expectError(error.InvalidMaxRasterValue, CoronaTrendVigor.init(testing.allocator, .{
        .max_raster_value = -1,
    }));
    try testing.expectError(error.InvalidMaxParameterValue, CoronaTrendVigor.init(testing.allocator, .{
        .min_parameter_value = 10,
        .max_parameter_value = 10,
    }));
    try testing.expectError(error.InvalidHighPassFilterCutoff, CoronaTrendVigor.init(testing.allocator, .{
        .high_pass_filter_cutoff = 1,
    }));
    try testing.expectError(error.InvalidMinimalPeriod, CoronaTrendVigor.init(testing.allocator, .{
        .minimal_period = 1,
    }));
    try testing.expectError(error.InvalidMaximalPeriod, CoronaTrendVigor.init(testing.allocator, .{
        .minimal_period = 10,
        .maximal_period = 10,
    }));
}

test "CoronaTrendVigor updateEntity" {
    const prime_count = 50;
    const inp: f64 = 100.0;
    const time: i64 = 0;

    // Update scalar
    {
        var x = try CoronaTrendVigor.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const s = Scalar{ .time = time, .value = inp };
        const out = x.updateScalar(&s);
        try testing.expectEqual(@as(usize, 2), out.len);
    }

    // Update bar
    {
        var x = try CoronaTrendVigor.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const b = Bar{ .time = time, .open = inp, .high = inp, .low = inp, .close = inp, .volume = 0 };
        const out = x.updateBar(&b);
        try testing.expectEqual(@as(usize, 2), out.len);
    }

    // Update quote
    {
        var x = try CoronaTrendVigor.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const q = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };
        const out = x.updateQuote(&q);
        try testing.expectEqual(@as(usize, 2), out.len);
    }

    // Update trade
    {
        var x = try CoronaTrendVigor.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const t = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = x.updateTrade(&t);
        try testing.expectEqual(@as(usize, 2), out.len);
    }
}
