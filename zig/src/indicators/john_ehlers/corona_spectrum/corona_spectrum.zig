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
const shape_mod = @import("../../core/outputs/shape.zig");
const heatmap_mod = @import("../../core/outputs/heatmap.zig");
const corona_mod = @import("../corona/corona.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Heatmap = heatmap_mod.Heatmap;
const Corona = corona_mod.Corona;
const CoronaSnapshot = corona_mod.CoronaSnapshot;

pub const CoronaSpectrumOutput = enum(u8) {
    value = 1,
    dominant_cycle = 2,
    dominant_cycle_median = 3,
};

pub const Params = struct {
    min_raster_value: f64 = 6.0,
    max_raster_value: f64 = 20.0,
    min_parameter_value: f64 = 6.0,
    max_parameter_value: f64 = 30.0,
    high_pass_filter_cutoff: i32 = 30,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub fn defaultParams() Params {
    return .{};
}

pub const Error = corona_mod.Error;

pub const CoronaSpectrumSnapshot = struct {
    heatmap: Heatmap,
    dominant_cycle: f64,
    dominant_cycle_median: f64,
    maximum_power: f64,
    normalization_power: f64,
    selected_bins: f64,
    has_cycle_estimate: f64,
    bank_length: usize,
    power_storage: [corona_mod.max_bank_length]f64 = [_]f64{math.nan(f64)} ** corona_mod.max_bank_length,
    relative_power_storage: [corona_mod.max_bank_length]f64 = [_]f64{math.nan(f64)} ** corona_mod.max_bank_length,

    pub fn powerSlice(self: *const CoronaSpectrumSnapshot) []const f64 {
        return self.power_storage[0..self.bank_length];
    }

    pub fn relativePowerSlice(self: *const CoronaSpectrumSnapshot) []const f64 {
        return self.relative_power_storage[0..self.bank_length];
    }
};

pub const Snapshot = CoronaSpectrumSnapshot;

pub fn formatMnemonic(buffer: []u8, params: Params) Error![]const u8 {
    const bc = params.bar_component orelse bar_component.default_bar_component;
    const qc = params.quote_component orelse quote_component.default_quote_component;
    const tc = params.trade_component orelse trade_component.default_trade_component;

    var triple_buffer: [64]u8 = undefined;
    const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buffer, bc, qc, tc);

    const min_period: i32 = @intFromFloat(@ceil(params.min_parameter_value));
    const max_period: i32 = @intFromFloat(@floor(params.max_parameter_value));

    return std.fmt.bufPrint(buffer, "cspect({d}, {d}, {d}, {d}, {d}{s})", .{
        corona_mod.normalizeZero(params.min_raster_value),
        corona_mod.normalizeZero(params.max_raster_value),
        min_period,
        max_period,
        params.high_pass_filter_cutoff,
        triple,
    }) catch return error.MnemonicTooLong;
}

pub const CoronaSpectrum = struct {
    engine: Corona,
    engine_params: corona_mod.Params,
    sample_count: usize,
    min_parameter_value: f64,
    max_parameter_value: f64,
    parameter_resolution: f64,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_storage: [256]u8,
    mnemonic_len: usize,
    description_storage: [320]u8,
    description_len: usize,
    mnemonic_dc_storage: [128]u8,
    mnemonic_dc_len: usize,
    description_dc_storage: [256]u8,
    description_dc_len: usize,
    mnemonic_dcm_storage: [128]u8,
    mnemonic_dcm_len: usize,
    description_dcm_storage: [256]u8,
    description_dcm_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!CoronaSpectrum {
        if (!corona_mod.validateFinite(params.min_raster_value) or params.min_raster_value < 0.0) {
            return error.InvalidDecibelsLowerThreshold;
        }
        if (!corona_mod.validateFinite(params.max_raster_value) or
            params.max_raster_value <= 0.0 or
            params.max_raster_value <= params.min_raster_value)
        {
            return error.InvalidDecibelsUpperThreshold;
        }
        if (!corona_mod.validateFinite(params.min_parameter_value)) {
            return error.InvalidMinParameterValue;
        }
        if (!corona_mod.validateFinite(params.max_parameter_value)) {
            return error.InvalidMaxParameterValue;
        }
        if (params.high_pass_filter_cutoff < 5) {
            return error.InvalidHighPassFilterCutoff;
        }

        const min_period_f = @ceil(params.min_parameter_value);
        const max_period_f = @floor(params.max_parameter_value);
        if (min_period_f < 6.0 or min_period_f > @as(f64, @floatFromInt(math.maxInt(i32)))) {
            return error.InvalidMinParameterValue;
        }
        if (max_period_f <= min_period_f or max_period_f > @as(f64, @floatFromInt(math.maxInt(i32)))) {
            return error.InvalidMaxParameterValue;
        }
        const min_period: i32 = @intFromFloat(min_period_f);
        const max_period: i32 = @intFromFloat(max_period_f);

        var engine = try Corona.init(allocator, .{
            .high_pass_filter_cutoff = params.high_pass_filter_cutoff,
            .minimal_period = min_period,
            .maximal_period = max_period,
            .decibels_lower_threshold = params.min_raster_value,
            .decibels_upper_threshold = params.max_raster_value,
        });
        errdefer engine.deinit();

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buffer: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buffer, bc, qc, tc);

        var mnemonic_storage: [256]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_storage, "cspect({d}, {d}, {d}, {d}, {d}{s})", .{
            corona_mod.normalizeZero(params.min_raster_value),
            corona_mod.normalizeZero(params.max_raster_value),
            min_period,
            max_period,
            params.high_pass_filter_cutoff,
            triple,
        }) catch return error.MnemonicTooLong;

        var description_storage: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Corona spectrum {s}", .{mn}) catch
            return error.MnemonicTooLong;

        var mnemonic_dc_storage: [128]u8 = undefined;
        const mn_dc = std.fmt.bufPrint(&mnemonic_dc_storage, "cspect-dc({d}{s})", .{
            params.high_pass_filter_cutoff,
            triple,
        }) catch return error.MnemonicTooLong;

        var description_dc_storage: [256]u8 = undefined;
        const desc_dc = std.fmt.bufPrint(&description_dc_storage, "Corona spectrum dominant cycle {s}", .{mn_dc}) catch
            return error.MnemonicTooLong;

        var mnemonic_dcm_storage: [128]u8 = undefined;
        const mn_dcm = std.fmt.bufPrint(&mnemonic_dcm_storage, "cspect-dcm({d}{s})", .{
            params.high_pass_filter_cutoff,
            triple,
        }) catch return error.MnemonicTooLong;

        var description_dcm_storage: [256]u8 = undefined;
        const desc_dcm = std.fmt.bufPrint(&description_dcm_storage, "Corona spectrum dominant cycle median {s}", .{mn_dcm}) catch
            return error.MnemonicTooLong;

        return .{
            .engine = engine,
            .engine_params = engine.parameters,
            .sample_count = 0,
            .min_parameter_value = min_period_f,
            .max_parameter_value = max_period_f,
            .parameter_resolution = 2.0,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mn.len,
            .description_storage = description_storage,
            .description_len = desc.len,
            .mnemonic_dc_storage = mnemonic_dc_storage,
            .mnemonic_dc_len = mn_dc.len,
            .description_dc_storage = description_dc_storage,
            .description_dc_len = desc_dc.len,
            .mnemonic_dcm_storage = mnemonic_dcm_storage,
            .mnemonic_dcm_len = mn_dcm.len,
            .description_dcm_storage = description_dcm_storage,
            .description_dcm_len = desc_dcm.len,
        };
    }

    pub fn deinit(self: *CoronaSpectrum) void {
        self.engine.deinit();
    }

    pub fn reset(self: *CoronaSpectrum) void {
        self.engine.reset();
        self.sample_count = 0;
    }

    pub fn fixSlices(self: *CoronaSpectrum) void {
        _ = self;
    }

    pub fn mnemonic(self: *const CoronaSpectrum) []const u8 {
        return self.mnemonic_storage[0..self.mnemonic_len];
    }

    fn description(self: *const CoronaSpectrum) []const u8 {
        return self.description_storage[0..self.description_len];
    }

    fn mnemonicDC(self: *const CoronaSpectrum) []const u8 {
        return self.mnemonic_dc_storage[0..self.mnemonic_dc_len];
    }

    fn descriptionDC(self: *const CoronaSpectrum) []const u8 {
        return self.description_dc_storage[0..self.description_dc_len];
    }

    fn mnemonicDCM(self: *const CoronaSpectrum) []const u8 {
        return self.mnemonic_dcm_storage[0..self.mnemonic_dcm_len];
    }

    fn descriptionDCM(self: *const CoronaSpectrum) []const u8 {
        return self.description_dcm_storage[0..self.description_dcm_len];
    }

    fn emptySnapshot(self: *const CoronaSpectrum, time: i64) CoronaSpectrumSnapshot {
        const nan = math.nan(f64);
        return .{
            .heatmap = Heatmap.empty(
                time,
                self.min_parameter_value,
                self.max_parameter_value,
                self.parameter_resolution,
            ),
            .dominant_cycle = nan,
            .dominant_cycle_median = nan,
            .maximum_power = nan,
            .normalization_power = nan,
            .selected_bins = nan,
            .has_cycle_estimate = nan,
            .bank_length = self.engine.filter_bank_length,
        };
    }

    pub fn tryUpdateFromSnapshot(self: *CoronaSpectrum, snapshot: CoronaSnapshot) Error!CoronaSpectrumSnapshot {
        const accepted = try corona_mod.acceptSnapshot(snapshot, self.engine_params, self.sample_count + 1);
        if (!accepted) {
            return self.emptySnapshot(snapshot.time);
        }
        self.sample_count = snapshot.sample_count;
        if (!snapshot.primed) {
            return self.emptySnapshot(snapshot.time);
        }

        const decibels = snapshot.decibelsSlice();
        var minimum = decibels[0];
        var maximum = decibels[0];
        for (decibels[1..]) |item| {
            if (item < minimum) minimum = item;
            if (item > maximum) maximum = item;
        }

        var result = CoronaSpectrumSnapshot{
            .heatmap = Heatmap.new(
                snapshot.time,
                self.min_parameter_value,
                self.max_parameter_value,
                self.parameter_resolution,
                minimum,
                maximum,
                decibels,
            ),
            .dominant_cycle = snapshot.dominant_cycle,
            .dominant_cycle_median = snapshot.dominant_cycle_median,
            .maximum_power = snapshot.maximum_power,
            .normalization_power = snapshot.normalization_power,
            .selected_bins = @floatFromInt(snapshot.selected_bins),
            .has_cycle_estimate = if (snapshot.has_cycle_estimate) 1.0 else 0.0,
            .bank_length = snapshot.bank_length,
        };
        @memcpy(result.power_storage[0..snapshot.bank_length], snapshot.powerSlice());
        @memcpy(result.relative_power_storage[0..snapshot.bank_length], snapshot.relativePowerSlice());
        return result;
    }

    pub fn updateFromSnapshot(self: *CoronaSpectrum, snapshot: CoronaSnapshot) CoronaSpectrumSnapshot {
        return self.tryUpdateFromSnapshot(snapshot) catch |err| {
            std.debug.panic("CoronaSpectrum.updateFromSnapshot failed: {s}", .{@errorName(err)});
        };
    }

    pub fn tryUpdate(self: *CoronaSpectrum, sample: f64, time: i64) Error!CoronaSpectrumSnapshot {
        const engine_snap = try self.engine.tryUpdate(sample, time);
        return self.tryUpdateFromSnapshot(engine_snap);
    }

    pub fn update(self: *CoronaSpectrum, sample: f64, time: i64) CoronaSpectrumSnapshot {
        return self.tryUpdate(sample, time) catch |err| {
            std.debug.panic("CoronaSpectrum.update failed: {s}", .{@errorName(err)});
        };
    }

    pub const updateSample = update;

    fn wrapEntity(time: i64, snapshot: CoronaSpectrumSnapshot) OutputArray {
        var out = OutputArray{};
        out.append(.{ .heatmap = snapshot.heatmap });
        out.append(.{ .scalar = .{ .time = time, .value = snapshot.dominant_cycle } });
        out.append(.{ .scalar = .{ .time = time, .value = snapshot.dominant_cycle_median } });
        return out;
    }

    pub fn updateScalar(self: *CoronaSpectrum, sample: *const Scalar) OutputArray {
        return wrapEntity(sample.time, self.update(sample.value, sample.time));
    }

    pub fn updateBar(self: *CoronaSpectrum, sample: *const Bar) OutputArray {
        return wrapEntity(sample.time, self.update(self.bar_func(sample.*), sample.time));
    }

    pub fn updateQuote(self: *CoronaSpectrum, sample: *const Quote) OutputArray {
        return wrapEntity(sample.time, self.update(self.quote_func(sample.*), sample.time));
    }

    pub fn updateTrade(self: *CoronaSpectrum, sample: *const Trade) OutputArray {
        return wrapEntity(sample.time, self.update(self.trade_func(sample.*), sample.time));
    }

    pub fn isPrimed(self: *const CoronaSpectrum) bool {
        return self.sample_count >= corona_mod.first_bank_sample;
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
const testdata = @import("testdata.zig");

const test_tolerance: f64 = 1e-10;

fn expectClose(actual: f64, expected: f64) !void {
    if (math.isNan(expected)) {
        try testing.expect(math.isNan(actual));
        return;
    }
    if (math.isInf(expected)) {
        try testing.expectEqual(expected, actual);
        return;
    }
    try testing.expect(!math.isNan(actual) and !math.isInf(actual));
    const bound = @max(test_tolerance, @abs(expected) * test_tolerance);
    try testing.expect(@abs(actual - expected) <= bound);
}

fn verifySpectrumCase(
    comptime bank_len: usize,
    hp_cutoff: i32,
    min_period: i32,
    max_period: i32,
    threshold: f64,
    expected_heatmap: *const [252][bank_len]f64,
    expected_min: *const [252]f64,
    expected_max: *const [252]f64,
    expected_dc: *const [252]f64,
    expected_dcm: *const [252]f64,
    expected_max_power: *const [252]f64,
    expected_norm_power: *const [252]f64,
    expected_selected_bins: *const [252]f64,
    expected_has_est: *const [252]f64,
    expected_power: *const [252][bank_len]f64,
    expected_rel_power: *const [252][bank_len]f64,
) !void {
    var instance = try CoronaSpectrum.init(testing.allocator, .{
        .min_raster_value = threshold,
        .max_raster_value = 20.0,
        .min_parameter_value = @floatFromInt(min_period),
        .max_parameter_value = @floatFromInt(max_period),
        .high_pass_filter_cutoff = hp_cutoff,
    });
    defer instance.deinit();

    const prices = testdata.testInput();
    for (prices, 0..) |price, index| {
        const snap = instance.update(price, @intCast(index));
        try expectClose(snap.heatmap.parameter_first, @floatFromInt(min_period));
        try expectClose(snap.heatmap.parameter_last, @floatFromInt(max_period));
        try expectClose(snap.heatmap.parameter_resolution, 2.0);

        const row_empty = math.isNan(expected_dc[index]);
        try testing.expectEqual(!row_empty, instance.isPrimed());
        try testing.expectEqual(bank_len, snap.powerSlice().len);
        try testing.expectEqual(bank_len, snap.relativePowerSlice().len);

        if (row_empty) {
            try testing.expect(snap.heatmap.isEmpty());
            for (snap.powerSlice()) |item| try testing.expect(math.isNan(item));
            for (snap.relativePowerSlice()) |item| try testing.expect(math.isNan(item));
        } else {
            const values = snap.heatmap.valuesSlice();
            try testing.expectEqual(bank_len, values.len);
            for (0..bank_len) |bin_index| {
                try expectClose(values[bin_index], expected_heatmap[index][bin_index]);
                try expectClose(snap.powerSlice()[bin_index], expected_power[index][bin_index]);
                try expectClose(snap.relativePowerSlice()[bin_index], expected_rel_power[index][bin_index]);
            }
        }

        try expectClose(snap.heatmap.value_min, expected_min[index]);
        try expectClose(snap.heatmap.value_max, expected_max[index]);
        try expectClose(snap.dominant_cycle, expected_dc[index]);
        try expectClose(snap.dominant_cycle_median, expected_dcm[index]);
        try expectClose(snap.maximum_power, expected_max_power[index]);
        try expectClose(snap.normalization_power, expected_norm_power[index]);
        try expectClose(snap.selected_bins, expected_selected_bins[index]);
        try expectClose(snap.has_cycle_estimate, expected_has_est[index]);
    }

    instance.reset();
    try testing.expect(!instance.isPrimed());
}

test "CoronaSpectrum matches all 16 reference cases" {
    try verifySpectrumCase(49, 30, 6, 30, 6.0, &testdata.expectedH30_PMIN6_PMAX30_DB6_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX30_DB6_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX30_DB6_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX30_DB6_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX30_DB6_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX30_DB6_MAXIMUM_POWER(), &testdata.expectedH30_PMIN6_PMAX30_DB6_NORMALIZATION_POWER(), &testdata.expectedH30_PMIN6_PMAX30_DB6_SELECTED_BINS(), &testdata.expectedH30_PMIN6_PMAX30_DB6_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN6_PMAX30_DB6_POWER(), &testdata.expectedH30_PMIN6_PMAX30_DB6_RELATIVE_POWER());
    try verifySpectrumCase(49, 30, 6, 30, 3.0, &testdata.expectedH30_PMIN6_PMAX30_DB3_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX30_DB3_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX30_DB3_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX30_DB3_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX30_DB3_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX30_DB3_MAXIMUM_POWER(), &testdata.expectedH30_PMIN6_PMAX30_DB3_NORMALIZATION_POWER(), &testdata.expectedH30_PMIN6_PMAX30_DB3_SELECTED_BINS(), &testdata.expectedH30_PMIN6_PMAX30_DB3_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN6_PMAX30_DB3_POWER(), &testdata.expectedH30_PMIN6_PMAX30_DB3_RELATIVE_POWER());
    try verifySpectrumCase(29, 30, 6, 20, 6.0, &testdata.expectedH30_PMIN6_PMAX20_DB6_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX20_DB6_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX20_DB6_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX20_DB6_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX20_DB6_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX20_DB6_MAXIMUM_POWER(), &testdata.expectedH30_PMIN6_PMAX20_DB6_NORMALIZATION_POWER(), &testdata.expectedH30_PMIN6_PMAX20_DB6_SELECTED_BINS(), &testdata.expectedH30_PMIN6_PMAX20_DB6_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN6_PMAX20_DB6_POWER(), &testdata.expectedH30_PMIN6_PMAX20_DB6_RELATIVE_POWER());
    try verifySpectrumCase(29, 30, 6, 20, 3.0, &testdata.expectedH30_PMIN6_PMAX20_DB3_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX20_DB3_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX20_DB3_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX20_DB3_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX20_DB3_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX20_DB3_MAXIMUM_POWER(), &testdata.expectedH30_PMIN6_PMAX20_DB3_NORMALIZATION_POWER(), &testdata.expectedH30_PMIN6_PMAX20_DB3_SELECTED_BINS(), &testdata.expectedH30_PMIN6_PMAX20_DB3_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN6_PMAX20_DB3_POWER(), &testdata.expectedH30_PMIN6_PMAX20_DB3_RELATIVE_POWER());
    try verifySpectrumCase(41, 30, 10, 30, 6.0, &testdata.expectedH30_PMIN10_PMAX30_DB6_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX30_DB6_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX30_DB6_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX30_DB6_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX30_DB6_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX30_DB6_MAXIMUM_POWER(), &testdata.expectedH30_PMIN10_PMAX30_DB6_NORMALIZATION_POWER(), &testdata.expectedH30_PMIN10_PMAX30_DB6_SELECTED_BINS(), &testdata.expectedH30_PMIN10_PMAX30_DB6_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN10_PMAX30_DB6_POWER(), &testdata.expectedH30_PMIN10_PMAX30_DB6_RELATIVE_POWER());
    try verifySpectrumCase(41, 30, 10, 30, 3.0, &testdata.expectedH30_PMIN10_PMAX30_DB3_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX30_DB3_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX30_DB3_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX30_DB3_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX30_DB3_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX30_DB3_MAXIMUM_POWER(), &testdata.expectedH30_PMIN10_PMAX30_DB3_NORMALIZATION_POWER(), &testdata.expectedH30_PMIN10_PMAX30_DB3_SELECTED_BINS(), &testdata.expectedH30_PMIN10_PMAX30_DB3_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN10_PMAX30_DB3_POWER(), &testdata.expectedH30_PMIN10_PMAX30_DB3_RELATIVE_POWER());
    try verifySpectrumCase(77, 30, 10, 48, 6.0, &testdata.expectedH30_PMIN10_PMAX48_DB6_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX48_DB6_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX48_DB6_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX48_DB6_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX48_DB6_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX48_DB6_MAXIMUM_POWER(), &testdata.expectedH30_PMIN10_PMAX48_DB6_NORMALIZATION_POWER(), &testdata.expectedH30_PMIN10_PMAX48_DB6_SELECTED_BINS(), &testdata.expectedH30_PMIN10_PMAX48_DB6_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN10_PMAX48_DB6_POWER(), &testdata.expectedH30_PMIN10_PMAX48_DB6_RELATIVE_POWER());
    try verifySpectrumCase(77, 30, 10, 48, 3.0, &testdata.expectedH30_PMIN10_PMAX48_DB3_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX48_DB3_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX48_DB3_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX48_DB3_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX48_DB3_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX48_DB3_MAXIMUM_POWER(), &testdata.expectedH30_PMIN10_PMAX48_DB3_NORMALIZATION_POWER(), &testdata.expectedH30_PMIN10_PMAX48_DB3_SELECTED_BINS(), &testdata.expectedH30_PMIN10_PMAX48_DB3_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN10_PMAX48_DB3_POWER(), &testdata.expectedH30_PMIN10_PMAX48_DB3_RELATIVE_POWER());
    try verifySpectrumCase(49, 48, 6, 30, 6.0, &testdata.expectedH48_PMIN6_PMAX30_DB6_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX30_DB6_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX30_DB6_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX30_DB6_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX30_DB6_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX30_DB6_MAXIMUM_POWER(), &testdata.expectedH48_PMIN6_PMAX30_DB6_NORMALIZATION_POWER(), &testdata.expectedH48_PMIN6_PMAX30_DB6_SELECTED_BINS(), &testdata.expectedH48_PMIN6_PMAX30_DB6_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN6_PMAX30_DB6_POWER(), &testdata.expectedH48_PMIN6_PMAX30_DB6_RELATIVE_POWER());
    try verifySpectrumCase(49, 48, 6, 30, 3.0, &testdata.expectedH48_PMIN6_PMAX30_DB3_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX30_DB3_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX30_DB3_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX30_DB3_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX30_DB3_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX30_DB3_MAXIMUM_POWER(), &testdata.expectedH48_PMIN6_PMAX30_DB3_NORMALIZATION_POWER(), &testdata.expectedH48_PMIN6_PMAX30_DB3_SELECTED_BINS(), &testdata.expectedH48_PMIN6_PMAX30_DB3_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN6_PMAX30_DB3_POWER(), &testdata.expectedH48_PMIN6_PMAX30_DB3_RELATIVE_POWER());
    try verifySpectrumCase(29, 48, 6, 20, 6.0, &testdata.expectedH48_PMIN6_PMAX20_DB6_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX20_DB6_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX20_DB6_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX20_DB6_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX20_DB6_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX20_DB6_MAXIMUM_POWER(), &testdata.expectedH48_PMIN6_PMAX20_DB6_NORMALIZATION_POWER(), &testdata.expectedH48_PMIN6_PMAX20_DB6_SELECTED_BINS(), &testdata.expectedH48_PMIN6_PMAX20_DB6_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN6_PMAX20_DB6_POWER(), &testdata.expectedH48_PMIN6_PMAX20_DB6_RELATIVE_POWER());
    try verifySpectrumCase(29, 48, 6, 20, 3.0, &testdata.expectedH48_PMIN6_PMAX20_DB3_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX20_DB3_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX20_DB3_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX20_DB3_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX20_DB3_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX20_DB3_MAXIMUM_POWER(), &testdata.expectedH48_PMIN6_PMAX20_DB3_NORMALIZATION_POWER(), &testdata.expectedH48_PMIN6_PMAX20_DB3_SELECTED_BINS(), &testdata.expectedH48_PMIN6_PMAX20_DB3_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN6_PMAX20_DB3_POWER(), &testdata.expectedH48_PMIN6_PMAX20_DB3_RELATIVE_POWER());
    try verifySpectrumCase(41, 48, 10, 30, 6.0, &testdata.expectedH48_PMIN10_PMAX30_DB6_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX30_DB6_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX30_DB6_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX30_DB6_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX30_DB6_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX30_DB6_MAXIMUM_POWER(), &testdata.expectedH48_PMIN10_PMAX30_DB6_NORMALIZATION_POWER(), &testdata.expectedH48_PMIN10_PMAX30_DB6_SELECTED_BINS(), &testdata.expectedH48_PMIN10_PMAX30_DB6_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN10_PMAX30_DB6_POWER(), &testdata.expectedH48_PMIN10_PMAX30_DB6_RELATIVE_POWER());
    try verifySpectrumCase(41, 48, 10, 30, 3.0, &testdata.expectedH48_PMIN10_PMAX30_DB3_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX30_DB3_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX30_DB3_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX30_DB3_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX30_DB3_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX30_DB3_MAXIMUM_POWER(), &testdata.expectedH48_PMIN10_PMAX30_DB3_NORMALIZATION_POWER(), &testdata.expectedH48_PMIN10_PMAX30_DB3_SELECTED_BINS(), &testdata.expectedH48_PMIN10_PMAX30_DB3_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN10_PMAX30_DB3_POWER(), &testdata.expectedH48_PMIN10_PMAX30_DB3_RELATIVE_POWER());
    try verifySpectrumCase(77, 48, 10, 48, 6.0, &testdata.expectedH48_PMIN10_PMAX48_DB6_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX48_DB6_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX48_DB6_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX48_DB6_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX48_DB6_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX48_DB6_MAXIMUM_POWER(), &testdata.expectedH48_PMIN10_PMAX48_DB6_NORMALIZATION_POWER(), &testdata.expectedH48_PMIN10_PMAX48_DB6_SELECTED_BINS(), &testdata.expectedH48_PMIN10_PMAX48_DB6_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN10_PMAX48_DB6_POWER(), &testdata.expectedH48_PMIN10_PMAX48_DB6_RELATIVE_POWER());
    try verifySpectrumCase(77, 48, 10, 48, 3.0, &testdata.expectedH48_PMIN10_PMAX48_DB3_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX48_DB3_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX48_DB3_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX48_DB3_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX48_DB3_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX48_DB3_MAXIMUM_POWER(), &testdata.expectedH48_PMIN10_PMAX48_DB3_NORMALIZATION_POWER(), &testdata.expectedH48_PMIN10_PMAX48_DB3_SELECTED_BINS(), &testdata.expectedH48_PMIN10_PMAX48_DB3_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN10_PMAX48_DB3_POWER(), &testdata.expectedH48_PMIN10_PMAX48_DB3_RELATIVE_POWER());
}

test "CoronaSpectrum metadata, mnemonics, validation, and entities" {
    var instance = try CoronaSpectrum.init(testing.allocator, defaultParams());
    defer instance.deinit();

    var md: Metadata = undefined;
    instance.getMetadata(&md);
    try testing.expectEqual(Identifier.corona_spectrum, md.identifier);
    try testing.expectEqualStrings("cspect(6, 20, 6, 30, 30)", md.mnemonic);
    try testing.expectEqual(@as(usize, 3), md.outputs_len);
    try testing.expectEqual(shape_mod.Shape.heatmap, md.outputs_buf[0].shape);
    try testing.expectEqual(shape_mod.Shape.scalar, md.outputs_buf[1].shape);
    try testing.expectEqualStrings("cspect-dc(30)", md.outputs_buf[1].mnemonic);
    try testing.expectEqual(shape_mod.Shape.scalar, md.outputs_buf[2].shape);
    try testing.expectEqualStrings("cspect-dcm(30)", md.outputs_buf[2].mnemonic);

    var zero_lower = try CoronaSpectrum.init(testing.allocator, .{ .min_raster_value = 0.0 });
    defer zero_lower.deinit();
    try testing.expectEqualStrings("cspect(0, 20, 6, 30, 30)", zero_lower.mnemonic());

    const combos = [_]struct {
        bc: ?bar_component.BarComponent,
        qc: ?quote_component.QuoteComponent,
        tc: ?trade_component.TradeComponent,
    }{
        .{ .bc = null, .qc = null, .tc = null },
        .{ .bc = .open, .qc = null, .tc = null },
        .{ .bc = null, .qc = .bid, .tc = null },
        .{ .bc = null, .qc = null, .tc = .volume },
        .{ .bc = .median, .qc = .ask, .tc = null },
        .{ .bc = .median, .qc = null, .tc = .volume },
        .{ .bc = null, .qc = .ask, .tc = .volume },
        .{ .bc = .open, .qc = .bid, .tc = .volume },
    };
    for (combos) |combo| {
        const p = Params{ .bar_component = combo.bc, .quote_component = combo.qc, .trade_component = combo.tc };
        var inst = try CoronaSpectrum.init(testing.allocator, p);
        defer inst.deinit();
        var mn_buffer: [128]u8 = undefined;
        const formatted = try formatMnemonic(&mn_buffer, p);
        try testing.expectEqualStrings(formatted, inst.mnemonic());
    }

    try testing.expectError(error.InvalidDecibelsLowerThreshold, CoronaSpectrum.init(testing.allocator, .{ .min_raster_value = -1.0 }));
    try testing.expectError(error.InvalidDecibelsUpperThreshold, CoronaSpectrum.init(testing.allocator, .{ .min_raster_value = 20.0, .max_raster_value = 20.0 }));
    try testing.expectError(error.InvalidMinParameterValue, CoronaSpectrum.init(testing.allocator, .{ .min_parameter_value = 5.0 }));
    try testing.expectError(error.InvalidMaxParameterValue, CoronaSpectrum.init(testing.allocator, .{ .min_parameter_value = 20.0, .max_parameter_value = 20.0 }));
    try testing.expectError(error.InvalidHighPassFilterCutoff, CoronaSpectrum.init(testing.allocator, .{ .high_pass_filter_cutoff = 4 }));

    const prices = testdata.testInput();
    const opens = testdata.testOpen();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();
    var bar_ind = try CoronaSpectrum.init(testing.allocator, defaultParams());
    defer bar_ind.deinit();
    var scalar_ind = try CoronaSpectrum.init(testing.allocator, defaultParams());
    defer scalar_ind.deinit();
    var quote_ind = try CoronaSpectrum.init(testing.allocator, defaultParams());
    defer quote_ind.deinit();
    var trade_ind = try CoronaSpectrum.init(testing.allocator, defaultParams());
    defer trade_ind.deinit();
    for (0..15) |index| {
        const bar = Bar{ .time = @intCast(index), .open = opens[index], .high = highs[index], .low = lows[index], .close = prices[index], .volume = 1000.0 };
        const scalar = Scalar{ .time = @intCast(index), .value = prices[index] };
        const quote = Quote{ .time = @intCast(index), .bid_price = prices[index] - 0.5, .ask_price = prices[index] + 0.5, .bid_size = 10.0, .ask_size = 10.0 };
        const trade = Trade{ .time = @intCast(index), .price = prices[index], .volume = 10.0 };
        try testing.expectEqual(@as(usize, 3), bar_ind.updateBar(&bar).len);
        try testing.expectEqual(@as(usize, 3), scalar_ind.updateScalar(&scalar).len);
        try testing.expectEqual(@as(usize, 3), quote_ind.updateQuote(&quote).len);
        try testing.expectEqual(@as(usize, 3), trade_ind.updateTrade(&trade).len);
    }
}
