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
const corona_mod = @import("../corona/corona.zig");

const OutputArray = indicator_mod.OutputArray;
const OutputValue = indicator_mod.OutputValue;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Heatmap = heatmap_mod.Heatmap;
const Corona = corona_mod.Corona;

/// Enumerates the outputs of the CoronaSignalToNoiseRatio indicator.
pub const CoronaSignalToNoiseRatioOutput = enum(u8) {
    value = 1,
    signal_to_noise_ratio = 2,
};

/// Parameters to create a CoronaSignalToNoiseRatio indicator.
pub const Params = struct {
    /// Number of elements in the heatmap raster. Default: 50.
    raster_length: i32 = 0,
    /// Maximal raster value (z) of the heatmap. Default: 20.
    max_raster_value: f64 = 0,
    /// Minimal ordinate (y) value of the heatmap. Default: 1.
    min_parameter_value: f64 = 0,
    /// Maximal ordinate (y) value of the heatmap. Default: 11.
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
    InvalidMinParameterValue,
    InvalidMaxParameterValue,
    InvalidHighPassFilterCutoff,
    InvalidMinimalPeriod,
    InvalidMaximalPeriod,
    MnemonicTooLong,
    OutOfMemory,
};

// SNR-specific constants.
const high_low_buffer_size = 5;
const high_low_buffer_size_min_one = high_low_buffer_size - 1;
const high_low_median_index = 2;
const average_sample_alpha = 0.1;
const average_sample_one_minus = 0.9;
const signal_ema_alpha = 0.2;
const signal_ema_one_minus = 0.9; // Intentional: sums to 1.1, per Ehlers.
const noise_ema_alpha = 0.1;
const noise_ema_one_minus = 0.9;
const ratio_offset_db = 3.5;
const ratio_upper_db = 10.0;
const db_gain = 20.0;
const width_low_ratio_threshold = 0.5;
const width_baseline = 0.2;
const width_slope = 0.4;
const raster_blend_exponent = 0.8;
const raster_blend_half = 0.5;
const raster_negative_arg_cutoff = 1.0;

/// Ehlers' Corona Signal-to-Noise Ratio heatmap indicator.
pub const CoronaSignalToNoiseRatio = struct {
    allocator: std.mem.Allocator,
    corona: Corona,
    raster_length: usize,
    raster_step: f64,
    max_raster_value: f64,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    raster: []f64,
    high_low_buffer: [high_low_buffer_size]f64,
    hl_sorted: [high_low_buffer_size]f64,
    average_sample_previous: f64,
    signal_previous: f64,
    noise_previous: f64,
    signal_to_noise_ratio: f64,
    is_started: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_buf: [256]u8,
    mnemonic_len: usize,
    description_buf: [320]u8,
    description_len: usize,
    mnemonic_snr_buf: [128]u8,
    mnemonic_snr_len: usize,
    description_snr_buf: [256]u8,
    description_snr_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!CoronaSignalToNoiseRatio {
        const def_raster_length: i32 = 50;
        const def_max_raster_value = 20.0;
        const def_min_param = 1.0;
        const def_max_param = 11.0;
        const def_hp_cutoff: i32 = 30;
        const def_min_period: i32 = 6;
        const def_max_period: i32 = 30;

        var raster_length = params.raster_length;
        if (raster_length == 0) raster_length = def_raster_length;

        var max_raster_value = params.max_raster_value;
        if (max_raster_value == 0) max_raster_value = def_max_raster_value;

        var min_param = params.min_parameter_value;
        if (min_param == 0) min_param = def_min_param;

        var max_param = params.max_parameter_value;
        if (max_param == 0) max_param = def_max_param;

        var hp_cutoff = params.high_pass_filter_cutoff;
        if (hp_cutoff == 0) hp_cutoff = def_hp_cutoff;

        var min_period = params.minimal_period;
        if (min_period == 0) min_period = def_min_period;

        var max_period = params.maximal_period;
        if (max_period == 0) max_period = def_max_period;

        if (raster_length < 2) return error.InvalidRasterLength;
        if (max_raster_value <= 0) return error.InvalidMaxRasterValue;
        if (min_param < 0) return error.InvalidMinParameterValue;
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

        const raster_len_usize: usize = @intCast(raster_length);
        const raster = allocator.alloc(f64, raster_len_usize) catch {
            corona.deinit();
            return error.OutOfMemory;
        };
        @memset(raster, 0);

        // Build component mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc,
            qc,
            tc,
        );

        // parameterResolution = (rasterLength-1) / (maxParam - minParam).
        const parameter_resolution = @as(f64, @floatFromInt(raster_length - 1)) / (max_param - min_param);

        // Build mnemonics.
        var mnemonic_buf: [256]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "csnr({d}, {d}, {d}, {d}, {d}{s})", .{
            raster_length,
            max_raster_value,
            min_param,
            max_param,
            hp_cutoff,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Corona signal to noise ratio {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_snr_buf: [128]u8 = undefined;
        const mn_snr = std.fmt.bufPrint(&mnemonic_snr_buf, "csnr-snr({d}{s})", .{
            hp_cutoff,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_snr_len = mn_snr.len;

        var description_snr_buf: [256]u8 = undefined;
        const desc_snr = std.fmt.bufPrint(&description_snr_buf, "Corona signal to noise ratio scalar {s}", .{mn_snr}) catch
            return error.MnemonicTooLong;
        const description_snr_len = desc_snr.len;

        return .{
            .allocator = allocator,
            .corona = corona,
            .raster_length = raster_len_usize,
            .raster_step = max_raster_value / @as(f64, @floatFromInt(raster_length)),
            .max_raster_value = max_raster_value,
            .min_parameter_value = min_param,
            .max_parameter_value = max_param,
            .parameter_resolution = parameter_resolution,
            .raster = raster,
            .high_low_buffer = .{ 0, 0, 0, 0, 0 },
            .hl_sorted = .{ 0, 0, 0, 0, 0 },
            .average_sample_previous = 0,
            .signal_previous = 0,
            .noise_previous = 0,
            .signal_to_noise_ratio = math.nan(f64),
            .is_started = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .mnemonic_snr_buf = mnemonic_snr_buf,
            .mnemonic_snr_len = mnemonic_snr_len,
            .description_snr_buf = description_snr_buf,
            .description_snr_len = description_snr_len,
        };
    }

    pub fn deinit(self: *CoronaSignalToNoiseRatio) void {
        self.allocator.free(self.raster);
        self.corona.deinit();
    }

    pub fn fixSlices(self: *CoronaSignalToNoiseRatio) void {
        _ = self;
    }

    fn mnemonic(self: *const CoronaSignalToNoiseRatio) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const CoronaSignalToNoiseRatio) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    fn mnemonicSNR(self: *const CoronaSignalToNoiseRatio) []const u8 {
        return self.mnemonic_snr_buf[0..self.mnemonic_snr_len];
    }

    fn descriptionSNR(self: *const CoronaSignalToNoiseRatio) []const u8 {
        return self.description_snr_buf[0..self.description_snr_len];
    }

    /// Update with a new sample plus bar extremes. Returns heatmap and SNR scalar.
    pub fn updateSample(self: *CoronaSignalToNoiseRatio, sample: f64, sample_low: f64, sample_high: f64, time: i64) struct { heatmap: Heatmap, snr: f64 } {
        if (math.isNan(sample)) {
            return .{
                .heatmap = Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                .snr = math.nan(f64),
            };
        }

        const primed = self.corona.update(sample);

        if (!self.is_started) {
            self.average_sample_previous = sample;
            self.high_low_buffer[high_low_buffer_size_min_one] = sample_high - sample_low;
            self.is_started = true;

            return .{
                .heatmap = Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                .snr = math.nan(f64),
            };
        }

        const max_amp_sq = self.corona.getMaximalAmplitudeSquared();

        const average_sample = average_sample_alpha * sample + average_sample_one_minus * self.average_sample_previous;
        self.average_sample_previous = average_sample;

        if (@abs(average_sample) > 0 or max_amp_sq > 0) {
            self.signal_previous = signal_ema_alpha * @sqrt(max_amp_sq) + signal_ema_one_minus * self.signal_previous;
        }

        // Shift H-L ring buffer left; push new value.
        for (0..high_low_buffer_size_min_one) |i| {
            self.high_low_buffer[i] = self.high_low_buffer[i + 1];
        }
        self.high_low_buffer[high_low_buffer_size_min_one] = sample_high - sample_low;

        var ratio: f64 = 0.0;
        if (@abs(average_sample) > 0) {
            for (0..high_low_buffer_size) |i| {
                self.hl_sorted[i] = self.high_low_buffer[i];
            }

            std.mem.sort(f64, &self.hl_sorted, {}, std.sort.asc(f64));
            self.noise_previous = noise_ema_alpha * self.hl_sorted[high_low_median_index] + noise_ema_one_minus * self.noise_previous;

            if (@abs(self.noise_previous) > 0) {
                ratio = db_gain * math.log10(self.signal_previous / self.noise_previous) + ratio_offset_db;
                if (ratio < 0) {
                    ratio = 0;
                } else if (ratio > ratio_upper_db) {
                    ratio = ratio_upper_db;
                }

                ratio /= ratio_upper_db; // ∈ [0, 1]
            }
        }

        self.signal_to_noise_ratio = (self.max_parameter_value - self.min_parameter_value) * ratio + self.min_parameter_value;

        // Raster update.
        var width: f64 = 0.0;
        if (ratio <= width_low_ratio_threshold) {
            width = width_baseline - width_slope * ratio;
        }

        const ratio_scaled_to_raster_length: i64 = @intFromFloat(@round(ratio * @as(f64, @floatFromInt(self.raster_length))));
        const ratio_scaled_to_max_raster_value = ratio * self.max_raster_value;

        for (0..self.raster_length) |i| {
            var value = self.raster[i];
            const i_i64: i64 = @intCast(i);

            if (i_i64 == ratio_scaled_to_raster_length) {
                value *= 0.5;
            } else if (width == 0) {
                // Above the high-ratio threshold: handled by the ratio>0.5 override below.
            } else {
                const argument_base = (ratio_scaled_to_max_raster_value - self.raster_step * @as(f64, @floatFromInt(i))) / width;
                if (i_i64 < ratio_scaled_to_raster_length) {
                    value = raster_blend_half * (math.pow(f64, argument_base, raster_blend_exponent) + value);
                } else {
                    const argument = -argument_base;
                    if (argument > raster_negative_arg_cutoff) {
                        value = raster_blend_half * (math.pow(f64, argument, raster_blend_exponent) + value);
                    } else {
                        value = self.max_raster_value;
                    }
                }
            }

            if (value < 0) {
                value = 0;
            } else if (value > self.max_raster_value) {
                value = self.max_raster_value;
            }

            if (ratio > width_low_ratio_threshold) {
                value = self.max_raster_value;
            }

            self.raster[i] = value;
        }

        if (!primed) {
            return .{
                .heatmap = Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                .snr = math.nan(f64),
            };
        }

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
            .snr = self.signal_to_noise_ratio,
        };
    }

    // --- Entity update methods ---

    pub fn updateBar(self: *CoronaSignalToNoiseRatio, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*), sample.low, sample.high);
    }

    pub fn updateQuote(self: *CoronaSignalToNoiseRatio, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateEntity(sample.time, v, v, v);
    }

    pub fn updateTrade(self: *CoronaSignalToNoiseRatio, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateEntity(sample.time, v, v, v);
    }

    pub fn updateScalar(self: *CoronaSignalToNoiseRatio, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value, sample.value, sample.value);
    }

    fn updateEntity(self: *CoronaSignalToNoiseRatio, time: i64, sample: f64, low: f64, high: f64) OutputArray {
        const result = self.updateSample(sample, low, high, time);
        var out = OutputArray{};
        out.append(.{ .heatmap = result.heatmap });
        out.append(.{ .scalar = .{ .time = time, .value = result.snr } });
        return out;
    }

    pub fn isPrimed(self: *const CoronaSignalToNoiseRatio) bool {
        return self.corona.isPrimed();
    }

    pub fn getMetadata(self: *const CoronaSignalToNoiseRatio, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
            .{ .mnemonic = self.mnemonicSNR(), .description = self.descriptionSNR() },
        };
        build_metadata_mod.buildMetadata(out, .corona_signal_to_noise_ratio, mn, desc, &texts);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *CoronaSignalToNoiseRatio) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(CoronaSignalToNoiseRatio);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


const tolerance = 1e-4;

fn almostEqual(a: f64, b: f64, eps: f64) bool {
    return @abs(a - b) <= eps;
}

/// Produce synthetic High/Low around the sample.
fn makeHL(i: usize, sample: f64) struct { low: f64, high: f64 } {
    const frac = 0.005 + 0.03 * (1.0 + @sin(@as(f64, @floatFromInt(i)) * 0.37));
    const half = sample * frac;
    return .{ .low = sample - half, .high = sample + half };
}

test "CoronaSignalToNoiseRatio update" {
    const Snap = struct { i: usize, snr: f64, vmn: f64, vmx: f64 };
    const snapshots = [_]Snap{
        .{ .i = 11, .snr = 1.0, .vmn = 0.0, .vmx = 20.0 },
        .{ .i = 12, .snr = 1.0, .vmn = 0.0, .vmx = 20.0 },
        .{ .i = 50, .snr = 1.0, .vmn = 0.0, .vmx = 20.0 },
        .{ .i = 100, .snr = 2.9986583538, .vmn = 4.2011609652, .vmx = 20.0 },
        .{ .i = 150, .snr = 1.0, .vmn = 0.0000000035, .vmx = 20.0 },
        .{ .i = 200, .snr = 1.0, .vmn = 0.0, .vmx = 20.0 },
        .{ .i = 251, .snr = 1.0, .vmn = 0.0000000026, .vmx = 20.0 },
    };

    var x = try CoronaSignalToNoiseRatio.init(testing.allocator, .{});
    defer x.deinit();

    var si: usize = 0;
    for (0..testdata.test_input.len) |i| {
        const hl = makeHL(i, testdata.test_input[i]);
        const result = x.updateSample(testdata.test_input[i], hl.low, hl.high, @intCast(i));

        try testing.expectEqual(@as(f64, 1.0), result.heatmap.parameter_first);
        try testing.expectEqual(@as(f64, 11.0), result.heatmap.parameter_last);
        try testing.expect(almostEqual(4.9, result.heatmap.parameter_resolution, 1e-9));

        if (!x.isPrimed()) {
            try testing.expect(result.heatmap.isEmpty());
            try testing.expect(math.isNan(result.snr));
            continue;
        }

        try testing.expectEqual(@as(usize, 50), result.heatmap.values_len);

        if (si < snapshots.len and snapshots[si].i == i) {
            try testing.expect(almostEqual(snapshots[si].snr, result.snr, tolerance));
            try testing.expect(almostEqual(snapshots[si].vmn, result.heatmap.value_min, tolerance));
            try testing.expect(almostEqual(snapshots[si].vmx, result.heatmap.value_max, tolerance));
            si += 1;
        }
    }

    try testing.expectEqual(snapshots.len, si);
}

test "CoronaSignalToNoiseRatio primes at bar 11" {
    var x = try CoronaSignalToNoiseRatio.init(testing.allocator, .{});
    defer x.deinit();

    try testing.expect(!x.isPrimed());

    var primed_at: ?usize = null;
    for (0..testdata.test_input.len) |i| {
        const hl = makeHL(i, testdata.test_input[i]);
        _ = x.updateSample(testdata.test_input[i], hl.low, hl.high, @intCast(i));
        if (x.isPrimed() and primed_at == null) {
            primed_at = i;
        }
    }

    try testing.expectEqual(@as(usize, 11), primed_at.?);
}

test "CoronaSignalToNoiseRatio NaN input" {
    var x = try CoronaSignalToNoiseRatio.init(testing.allocator, .{});
    defer x.deinit();

    const result = x.updateSample(math.nan(f64), math.nan(f64), math.nan(f64), 0);
    try testing.expect(result.heatmap.isEmpty());
    try testing.expect(math.isNan(result.snr));
    try testing.expect(!x.isPrimed());
}

test "CoronaSignalToNoiseRatio metadata" {
    var x = try CoronaSignalToNoiseRatio.init(testing.allocator, .{});
    defer x.deinit();

    var md: Metadata = undefined;
    x.getMetadata(&md);

    const mn_value = "csnr(50, 20, 1, 11, 30, hl/2)";
    const mn_snr = "csnr-snr(30, hl/2)";

    try testing.expectEqualStrings(mn_value, x.mnemonic());
    try testing.expectEqual(Identifier.corona_signal_to_noise_ratio, md.identifier);
    try testing.expectEqualStrings(mn_value, md.mnemonic);
    try testing.expectEqual(@as(usize, 2), md.outputs_len);

    const outputs = md.outputs_buf[0..md.outputs_len];
    try testing.expectEqualStrings(mn_value, outputs[0].mnemonic);
    try testing.expectEqualStrings(mn_snr, outputs[1].mnemonic);
}

test "CoronaSignalToNoiseRatio validation" {
    try testing.expectError(error.InvalidRasterLength, CoronaSignalToNoiseRatio.init(testing.allocator, .{
        .raster_length = 1,
    }));
    try testing.expectError(error.InvalidMaxRasterValue, CoronaSignalToNoiseRatio.init(testing.allocator, .{
        .max_raster_value = -1,
    }));
    try testing.expectError(error.InvalidMaxParameterValue, CoronaSignalToNoiseRatio.init(testing.allocator, .{
        .min_parameter_value = 5,
        .max_parameter_value = 5,
    }));
    try testing.expectError(error.InvalidHighPassFilterCutoff, CoronaSignalToNoiseRatio.init(testing.allocator, .{
        .high_pass_filter_cutoff = 1,
    }));
    try testing.expectError(error.InvalidMinimalPeriod, CoronaSignalToNoiseRatio.init(testing.allocator, .{
        .minimal_period = 1,
    }));
    try testing.expectError(error.InvalidMaximalPeriod, CoronaSignalToNoiseRatio.init(testing.allocator, .{
        .minimal_period = 10,
        .maximal_period = 10,
    }));
}

test "CoronaSignalToNoiseRatio updateEntity" {
    const prime_count = 50;
    const inp: f64 = 100.0;
    const time: i64 = 0;

    // Update scalar
    {
        var x = try CoronaSignalToNoiseRatio.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            const hl = makeHL(idx, testdata.test_input[idx % testdata.test_input.len]);
            _ = x.updateSample(testdata.test_input[idx % testdata.test_input.len], hl.low, hl.high, time);
        }
        const s = Scalar{ .time = time, .value = inp };
        const out = x.updateScalar(&s);
        try testing.expectEqual(@as(usize, 2), out.len);
    }

    // Update bar
    {
        var x = try CoronaSignalToNoiseRatio.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            const hl = makeHL(idx, testdata.test_input[idx % testdata.test_input.len]);
            _ = x.updateSample(testdata.test_input[idx % testdata.test_input.len], hl.low, hl.high, time);
        }
        const b = Bar{ .time = time, .open = inp, .high = inp, .low = inp, .close = inp, .volume = 0 };
        const out = x.updateBar(&b);
        try testing.expectEqual(@as(usize, 2), out.len);
    }

    // Update quote
    {
        var x = try CoronaSignalToNoiseRatio.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            const hl = makeHL(idx, testdata.test_input[idx % testdata.test_input.len]);
            _ = x.updateSample(testdata.test_input[idx % testdata.test_input.len], hl.low, hl.high, time);
        }
        const q = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };
        const out = x.updateQuote(&q);
        try testing.expectEqual(@as(usize, 2), out.len);
    }

    // Update trade
    {
        var x = try CoronaSignalToNoiseRatio.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            const hl = makeHL(idx, testdata.test_input[idx % testdata.test_input.len]);
            _ = x.updateSample(testdata.test_input[idx % testdata.test_input.len], hl.low, hl.high, time);
        }
        const t = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = x.updateTrade(&t);
        try testing.expectEqual(@as(usize, 2), out.len);
    }
}
