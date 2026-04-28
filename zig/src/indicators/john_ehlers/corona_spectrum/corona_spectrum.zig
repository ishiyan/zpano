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

/// Enumerates the outputs of the CoronaSpectrum indicator.
pub const CoronaSpectrumOutput = enum(u8) {
    value = 1,
    dominant_cycle = 2,
    dominant_cycle_median = 3,
};

/// Parameters to create a CoronaSpectrum indicator.
pub const Params = struct {
    /// Minimal raster value (z) of the heatmap, in decibels. Default: 6.
    min_raster_value: f64 = 0,
    /// Maximal raster value (z) of the heatmap, in decibels. Default: 20.
    max_raster_value: f64 = 0,
    /// Minimal ordinate (y) value — minimal cycle period. Default: 6.
    min_parameter_value: f64 = 0,
    /// Maximal ordinate (y) value — maximal cycle period. Default: 30.
    max_parameter_value: f64 = 0,
    /// High-pass filter cutoff. Default: 30.
    high_pass_filter_cutoff: i32 = 0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Error = error{
    InvalidMinRasterValue,
    InvalidMaxRasterValue,
    InvalidMinParameterValue,
    InvalidMaxParameterValue,
    InvalidHighPassFilterCutoff,
    MnemonicTooLong,
    OutOfMemory,
};

/// Ehlers' Corona Spectrum heatmap indicator.
pub const CoronaSpectrum = struct {
    corona: Corona,
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
    mnemonic_dc_buf: [128]u8,
    mnemonic_dc_len: usize,
    description_dc_buf: [256]u8,
    description_dc_len: usize,
    mnemonic_dcm_buf: [128]u8,
    mnemonic_dcm_len: usize,
    description_dcm_buf: [256]u8,
    description_dcm_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!CoronaSpectrum {
        const def_min_raster = 6.0;
        const def_max_raster = 20.0;
        const def_min_param = 6.0;
        const def_max_param = 30.0;
        const def_hp_cutoff: i32 = 30;

        var min_raster = params.min_raster_value;
        if (min_raster == 0) min_raster = def_min_raster;

        var max_raster = params.max_raster_value;
        if (max_raster == 0) max_raster = def_max_raster;

        var min_param_raw = params.min_parameter_value;
        if (min_param_raw == 0) min_param_raw = def_min_param;

        var max_param_raw = params.max_parameter_value;
        if (max_param_raw == 0) max_param_raw = def_max_param;

        var hp_cutoff = params.high_pass_filter_cutoff;
        if (hp_cutoff == 0) hp_cutoff = def_hp_cutoff;

        if (min_raster < 0) return error.InvalidMinRasterValue;
        if (max_raster <= min_raster) return error.InvalidMaxRasterValue;

        // MBST rounds min up and max down to integers.
        const min_param = @ceil(min_param_raw);
        const max_param = @floor(max_param_raw);

        if (min_param < 2) return error.InvalidMinParameterValue;
        if (max_param <= min_param) return error.InvalidMaxParameterValue;
        if (hp_cutoff < 2) return error.InvalidHighPassFilterCutoff;

        // Default bar component is BarMedianPrice for corona indicators.
        const bc = params.bar_component orelse bar_component.BarComponent.median;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var corona = Corona.init(allocator, .{
            .high_pass_filter_cutoff = hp_cutoff,
            .minimal_period = @intFromFloat(min_param),
            .maximal_period = @intFromFloat(max_param),
            .decibels_lower_threshold = min_raster,
            .decibels_upper_threshold = max_raster,
        }) catch return error.OutOfMemory;
        errdefer corona.deinit();

        // Build component mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc,
            qc,
            tc,
        );

        // Parameter resolution: (filterBankLength-1) / (maxParam - minParam).
        const parameter_resolution = @as(f64, @floatFromInt(corona.getFilterBankLength() - 1)) / (max_param - min_param);

        // Build mnemonics.
        var mnemonic_buf: [256]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "cspect({d}, {d}, {d}, {d}, {d}{s})", .{
            min_raster,
            max_raster,
            min_param,
            max_param,
            hp_cutoff,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Corona spectrum {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_dc_buf: [128]u8 = undefined;
        const mn_dc = std.fmt.bufPrint(&mnemonic_dc_buf, "cspect-dc({d}{s})", .{
            hp_cutoff,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_dc_len = mn_dc.len;

        var description_dc_buf: [256]u8 = undefined;
        const desc_dc = std.fmt.bufPrint(&description_dc_buf, "Corona spectrum dominant cycle {s}", .{mn_dc}) catch
            return error.MnemonicTooLong;
        const description_dc_len = desc_dc.len;

        var mnemonic_dcm_buf: [128]u8 = undefined;
        const mn_dcm = std.fmt.bufPrint(&mnemonic_dcm_buf, "cspect-dcm({d}{s})", .{
            hp_cutoff,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_dcm_len = mn_dcm.len;

        var description_dcm_buf: [256]u8 = undefined;
        const desc_dcm = std.fmt.bufPrint(&description_dcm_buf, "Corona spectrum dominant cycle median {s}", .{mn_dcm}) catch
            return error.MnemonicTooLong;
        const description_dcm_len = desc_dcm.len;

        return .{
            .corona = corona,
            .min_parameter_value = min_param,
            .max_parameter_value = max_param,
            .parameter_resolution = parameter_resolution,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .mnemonic_dc_buf = mnemonic_dc_buf,
            .mnemonic_dc_len = mnemonic_dc_len,
            .description_dc_buf = description_dc_buf,
            .description_dc_len = description_dc_len,
            .mnemonic_dcm_buf = mnemonic_dcm_buf,
            .mnemonic_dcm_len = mnemonic_dcm_len,
            .description_dcm_buf = description_dcm_buf,
            .description_dcm_len = description_dcm_len,
        };
    }

    pub fn deinit(self: *CoronaSpectrum) void {
        self.corona.deinit();
    }

    pub fn fixSlices(self: *CoronaSpectrum) void {
        _ = self;
    }

    fn mnemonic(self: *const CoronaSpectrum) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const CoronaSpectrum) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    fn mnemonicDC(self: *const CoronaSpectrum) []const u8 {
        return self.mnemonic_dc_buf[0..self.mnemonic_dc_len];
    }

    fn descriptionDC(self: *const CoronaSpectrum) []const u8 {
        return self.description_dc_buf[0..self.description_dc_len];
    }

    fn mnemonicDCM(self: *const CoronaSpectrum) []const u8 {
        return self.mnemonic_dcm_buf[0..self.mnemonic_dcm_len];
    }

    fn descriptionDCM(self: *const CoronaSpectrum) []const u8 {
        return self.description_dcm_buf[0..self.description_dcm_len];
    }

    /// Update with a new sample. Returns heatmap, dominant cycle, dominant cycle median.
    pub fn updateSample(self: *CoronaSpectrum, sample: f64, time: i64) struct { heatmap: Heatmap, dc: f64, dcm: f64 } {
        if (math.isNan(sample)) {
            return .{
                .heatmap = Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                .dc = math.nan(f64),
                .dcm = math.nan(f64),
            };
        }

        const primed = self.corona.update(sample);
        if (!primed) {
            return .{
                .heatmap = Heatmap.empty(time, self.min_parameter_value, self.max_parameter_value, self.parameter_resolution),
                .dc = math.nan(f64),
                .dcm = math.nan(f64),
            };
        }

        const bank = self.corona.getFilterBank();
        var values: [heatmap_mod.max_heatmap_values]f64 = undefined;
        var value_min: f64 = math.inf(f64);
        var value_max: f64 = -math.inf(f64);

        for (0..bank.len) |i| {
            const v = bank[i].decibels;
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
                values[0..bank.len],
            ),
            .dc = self.corona.getDominantCycle(),
            .dcm = self.corona.getDominantCycleMedian(),
        };
    }

    // --- Entity update methods ---

    pub fn updateBar(self: *CoronaSpectrum, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *CoronaSpectrum, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *CoronaSpectrum, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    pub fn updateScalar(self: *CoronaSpectrum, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    fn updateEntity(self: *CoronaSpectrum, time: i64, sample: f64) OutputArray {
        const result = self.updateSample(sample, time);
        var out = OutputArray{};
        out.append(.{ .heatmap = result.heatmap });
        out.append(.{ .scalar = .{ .time = time, .value = result.dc } });
        out.append(.{ .scalar = .{ .time = time, .value = result.dcm } });
        return out;
    }

    pub fn isPrimed(self: *const CoronaSpectrum) bool {
        return self.corona.isPrimed();
    }

    pub fn getMetadata(self: *const CoronaSpectrum, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
            .{ .mnemonic = self.mnemonicDC(), .description = self.descriptionDC() },
            .{ .mnemonic = self.mnemonicDCM(), .description = self.descriptionDCM() },

        };
        build_metadata_mod.buildMetadata(out, .corona_spectrum, mn, desc, &texts);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *CoronaSpectrum) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(CoronaSpectrum);
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

test "CoronaSpectrum update" {
    const Snap = struct { i: usize, dc: f64, dcm: f64 };
    const snapshots = [_]Snap{
        .{ .i = 11, .dc = 17.7604672565, .dcm = 17.7604672565 },
        .{ .i = 12, .dc = 6.0000000000, .dcm = 6.0000000000 },
        .{ .i = 50, .dc = 15.9989078712, .dcm = 15.9989078712 },
        .{ .i = 100, .dc = 14.7455497547, .dcm = 14.7455497547 },
        .{ .i = 150, .dc = 17.5000000000, .dcm = 17.2826036069 },
        .{ .i = 200, .dc = 19.7557338512, .dcm = 20.0000000000 },
        .{ .i = 251, .dc = 6.0000000000, .dcm = 6.0000000000 },
    };

    var x = try CoronaSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    var si: usize = 0;
    for (0..test_input.len) |i| {
        const result = x.updateSample(test_input[i], @intCast(i));

        try testing.expectEqual(@as(f64, 6.0), result.heatmap.parameter_first);
        try testing.expectEqual(@as(f64, 30.0), result.heatmap.parameter_last);
        try testing.expectEqual(@as(f64, 2.0), result.heatmap.parameter_resolution);

        if (!x.isPrimed()) {
            try testing.expect(result.heatmap.isEmpty());
            try testing.expect(math.isNan(result.dc));
            try testing.expect(math.isNan(result.dcm));
            continue;
        }

        try testing.expectEqual(@as(usize, 49), result.heatmap.values_len);

        if (si < snapshots.len and snapshots[si].i == i) {
            try testing.expect(almostEqual(snapshots[si].dc, result.dc, tolerance));
            try testing.expect(almostEqual(snapshots[si].dcm, result.dcm, tolerance));
            si += 1;
        }
    }

    try testing.expectEqual(snapshots.len, si);
}

test "CoronaSpectrum primes at bar 11" {
    var x = try CoronaSpectrum.init(testing.allocator, .{});
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

test "CoronaSpectrum NaN input" {
    var x = try CoronaSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    const result = x.updateSample(math.nan(f64), 0);
    try testing.expect(result.heatmap.isEmpty());
    try testing.expect(math.isNan(result.dc));
    try testing.expect(math.isNan(result.dcm));
    try testing.expect(!x.isPrimed());
}

test "CoronaSpectrum metadata" {
    var x = try CoronaSpectrum.init(testing.allocator, .{});
    defer x.deinit();

    var md: Metadata = undefined;
    x.getMetadata(&md);

    const mn_value = "cspect(6, 20, 6, 30, 30, hl/2)";
    const mn_dc = "cspect-dc(30, hl/2)";
    const mn_dcm = "cspect-dcm(30, hl/2)";

    try testing.expectEqualStrings(mn_value, x.mnemonic());
    try testing.expectEqual(Identifier.corona_spectrum, md.identifier);
    try testing.expectEqualStrings(mn_value, md.mnemonic);
    try testing.expectEqual(@as(usize, 3), md.outputs_len);

    const outputs = md.outputs_buf[0..md.outputs_len];
    try testing.expectEqualStrings(mn_value, outputs[0].mnemonic);
    try testing.expectEqualStrings(mn_dc, outputs[1].mnemonic);
    try testing.expectEqualStrings(mn_dcm, outputs[2].mnemonic);
}

test "CoronaSpectrum custom ranges round to integers" {
    var x = try CoronaSpectrum.init(testing.allocator, .{
        .min_raster_value = 4,
        .max_raster_value = 25,
        .min_parameter_value = 8.7, // ceils to 9
        .max_parameter_value = 40.4, // floors to 40
        .high_pass_filter_cutoff = 20,
    });
    defer x.deinit();

    try testing.expectEqual(@as(f64, 9.0), x.min_parameter_value);
    try testing.expectEqual(@as(f64, 40.0), x.max_parameter_value);
    try testing.expectEqualStrings("cspect(4, 25, 9, 40, 20, hl/2)", x.mnemonic());
}

test "CoronaSpectrum validation" {
    try testing.expectError(error.InvalidMaxRasterValue, CoronaSpectrum.init(testing.allocator, .{
        .min_raster_value = 10,
        .max_raster_value = 10,
    }));
    try testing.expectError(error.InvalidMinParameterValue, CoronaSpectrum.init(testing.allocator, .{
        .min_parameter_value = 1,
    }));
    try testing.expectError(error.InvalidMaxParameterValue, CoronaSpectrum.init(testing.allocator, .{
        .min_parameter_value = 20,
        .max_parameter_value = 20,
    }));
    try testing.expectError(error.InvalidHighPassFilterCutoff, CoronaSpectrum.init(testing.allocator, .{
        .high_pass_filter_cutoff = 1,
    }));
}

test "CoronaSpectrum updateEntity" {
    const prime_count = 50;
    const inp: f64 = 100.0;
    const time: i64 = 0;

    // Update scalar
    {
        var x = try CoronaSpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const s = Scalar{ .time = time, .value = inp };
        const out = x.updateScalar(&s);
        try testing.expectEqual(@as(usize, 3), out.len);
    }

    // Update bar
    {
        var x = try CoronaSpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const b = Bar{ .time = time, .open = inp, .high = inp, .low = inp, .close = inp, .volume = 0 };
        const out = x.updateBar(&b);
        try testing.expectEqual(@as(usize, 3), out.len);
    }

    // Update quote
    {
        var x = try CoronaSpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const q = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };
        const out = x.updateQuote(&q);
        try testing.expectEqual(@as(usize, 3), out.len);
    }

    // Update trade
    {
        var x = try CoronaSpectrum.init(testing.allocator, .{});
        defer x.deinit();
        for (0..prime_count) |idx| {
            _ = x.updateSample(test_input[idx % test_input.len], time);
        }
        const t = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = x.updateTrade(&t);
        try testing.expectEqual(@as(usize, 3), out.len);
    }
}
