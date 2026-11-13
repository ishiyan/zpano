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
const Raster = corona_mod.Raster;

pub const CoronaSignalToNoiseRatioOutput = enum(u8) {
    value = 1,
    signal_to_noise_ratio = 2,
};

pub const Params = struct {
    raster_length: i32 = 50,
    max_raster_value: f64 = 20.0,
    min_parameter_value: f64 = 1.0,
    max_parameter_value: f64 = 11.0,
    high_pass_filter_cutoff: i32 = 30,
    minimal_period: i32 = 6,
    maximal_period: i32 = 30,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub fn defaultParams() Params {
    return .{};
}

pub const Error = corona_mod.Error;

pub const CoronaSignalToNoiseRatioSnapshot = struct {
    heatmap: Heatmap,
    signal_to_noise_ratio: f64,
    physical_decibels: f64,
    calibrated_decibels: f64,
    normalized_position: f64,
    signal: f64,
    noise: f64,
    median_bar_range: f64,
    width: f64,
    dominant_cycle: f64,
    dominant_cycle_median: f64,
    has_cycle_estimate: f64,
    ratio_is_finite: f64,
};

pub const Snapshot = CoronaSignalToNoiseRatioSnapshot;

fn validateHighLow(low: f64, high: f64) Error!f64 {
    if (math.isNan(low) or math.isNan(high)) return math.nan(f64);
    if (math.isInf(low) or math.isInf(high) or high < low) return error.InvalidBarRange;
    return corona_mod.checked(high - low);
}

fn validateBarRange(bar_range: f64) Error!f64 {
    if (math.isNan(bar_range)) return math.nan(f64);
    if (math.isInf(bar_range) or bar_range < 0.0) return error.InvalidBarRange;
    return bar_range;
}

fn formatSnrFlags(buffer: []u8, params: Params) Error![]const u8 {
    var pos: usize = 0;
    if (params.minimal_period != 6) {
        const chunk = std.fmt.bufPrint(buffer[pos..], ", pmin={d}", .{params.minimal_period}) catch return error.MnemonicTooLong;
        pos += chunk.len;
    }
    if (params.maximal_period != 30) {
        const chunk = std.fmt.bufPrint(buffer[pos..], ", pmax={d}", .{params.maximal_period}) catch return error.MnemonicTooLong;
        pos += chunk.len;
    }
    return buffer[0..pos];
}

pub fn formatMnemonic(buffer: []u8, params: Params) Error![]const u8 {
    const bc = params.bar_component orelse bar_component.default_bar_component;
    const qc = params.quote_component orelse quote_component.default_quote_component;
    const tc = params.trade_component orelse trade_component.default_trade_component;

    var triple_buffer: [64]u8 = undefined;
    const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buffer, bc, qc, tc);
    var flags_buffer: [64]u8 = undefined;
    const flags = try formatSnrFlags(&flags_buffer, params);

    return std.fmt.bufPrint(buffer, "csnr({d}, {d}, {d}, {d}, {d}{s}{s})", .{
        params.raster_length,
        corona_mod.normalizeZero(params.max_raster_value),
        corona_mod.normalizeZero(params.min_parameter_value),
        corona_mod.normalizeZero(params.max_parameter_value),
        params.high_pass_filter_cutoff,
        flags,
        triple,
    }) catch return error.MnemonicTooLong;
}

pub const CoronaSignalToNoiseRatio = struct {
    engine: Corona,
    engine_params: corona_mod.Params,
    raster: Raster,
    sample_count: usize,
    average_price: f64,
    signal: f64,
    noise: f64,
    ranges: [5]f64,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_storage: [256]u8,
    mnemonic_len: usize,
    description_storage: [320]u8,
    description_len: usize,
    mnemonic_snr_storage: [160]u8,
    mnemonic_snr_len: usize,
    description_snr_storage: [256]u8,
    description_snr_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!CoronaSignalToNoiseRatio {
        if (!corona_mod.validateFinite(params.max_raster_value) or params.max_raster_value <= 6.0) {
            return error.InvalidMaxRasterValue;
        }

        var raster = try Raster.init(
            allocator,
            params.raster_length,
            params.min_parameter_value,
            params.max_parameter_value,
            params.max_raster_value,
        );
        errdefer raster.deinit();

        var engine = try Corona.init(allocator, .{
            .high_pass_filter_cutoff = params.high_pass_filter_cutoff,
            .minimal_period = params.minimal_period,
            .maximal_period = params.maximal_period,
            .decibels_lower_threshold = 6.0,
            .decibels_upper_threshold = raster.ceiling,
        });
        errdefer engine.deinit();

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buffer: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buffer, bc, qc, tc);
        var flags_buffer: [64]u8 = undefined;
        const flags = try formatSnrFlags(&flags_buffer, params);

        var mnemonic_storage: [256]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_storage, "csnr({d}, {d}, {d}, {d}, {d}{s}{s})", .{
            raster.length,
            corona_mod.normalizeZero(raster.ceiling),
            corona_mod.normalizeZero(raster.lower),
            corona_mod.normalizeZero(raster.upper),
            params.high_pass_filter_cutoff,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;

        var description_storage: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Corona signal to noise ratio {s}", .{mn}) catch
            return error.MnemonicTooLong;

        var mnemonic_snr_storage: [160]u8 = undefined;
        const mn_snr = std.fmt.bufPrint(&mnemonic_snr_storage, "csnr-snr({d}{s}{s})", .{
            params.high_pass_filter_cutoff,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;

        var description_snr_storage: [256]u8 = undefined;
        const desc_snr = std.fmt.bufPrint(&description_snr_storage, "Corona signal to noise ratio scalar {s}", .{mn_snr}) catch
            return error.MnemonicTooLong;

        return .{
            .engine = engine,
            .engine_params = engine.parameters,
            .raster = raster,
            .sample_count = 0,
            .average_price = 0.0,
            .signal = 0.0,
            .noise = 0.0,
            .ranges = [_]f64{0.0} ** 5,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mn.len,
            .description_storage = description_storage,
            .description_len = desc.len,
            .mnemonic_snr_storage = mnemonic_snr_storage,
            .mnemonic_snr_len = mn_snr.len,
            .description_snr_storage = description_snr_storage,
            .description_snr_len = desc_snr.len,
        };
    }

    pub fn deinit(self: *CoronaSignalToNoiseRatio) void {
        self.engine.deinit();
        self.raster.deinit();
    }

    pub fn reset(self: *CoronaSignalToNoiseRatio) void {
        self.engine.reset();
        self.raster.reset();
        self.sample_count = 0;
        self.average_price = 0.0;
        self.signal = 0.0;
        self.noise = 0.0;
        self.ranges = [_]f64{0.0} ** 5;
    }

    pub fn fixSlices(self: *CoronaSignalToNoiseRatio) void {
        _ = self;
    }

    pub fn mnemonic(self: *const CoronaSignalToNoiseRatio) []const u8 {
        return self.mnemonic_storage[0..self.mnemonic_len];
    }

    fn description(self: *const CoronaSignalToNoiseRatio) []const u8 {
        return self.description_storage[0..self.description_len];
    }

    fn mnemonicSNR(self: *const CoronaSignalToNoiseRatio) []const u8 {
        return self.mnemonic_snr_storage[0..self.mnemonic_snr_len];
    }

    fn descriptionSNR(self: *const CoronaSignalToNoiseRatio) []const u8 {
        return self.description_snr_storage[0..self.description_snr_len];
    }

    fn emptySnapshot(self: *const CoronaSignalToNoiseRatio, time: i64) CoronaSignalToNoiseRatioSnapshot {
        const nan = math.nan(f64);
        return .{
            .heatmap = self.raster.empty(time),
            .signal_to_noise_ratio = nan,
            .physical_decibels = nan,
            .calibrated_decibels = nan,
            .normalized_position = nan,
            .signal = nan,
            .noise = nan,
            .median_bar_range = nan,
            .width = nan,
            .dominant_cycle = nan,
            .dominant_cycle_median = nan,
            .has_cycle_estimate = nan,
            .ratio_is_finite = nan,
        };
    }

    pub fn tryUpdateFromSnapshotRange(
        self: *CoronaSignalToNoiseRatio,
        snapshot: CoronaSnapshot,
        bar_range: f64,
    ) Error!CoronaSignalToNoiseRatioSnapshot {
        const accepted = try corona_mod.acceptSnapshot(snapshot, self.engine_params, self.sample_count + 1);
        if (!accepted) {
            return self.emptySnapshot(snapshot.time);
        }
        const range_value = try validateBarRange(bar_range);
        if (math.isNan(range_value)) {
            return error.InvalidBarRange;
        }
        self.sample_count = snapshot.sample_count;

        std.mem.copyForwards(f64, self.ranges[0..4], self.ranges[1..5]);
        self.ranges[4] = range_value;
        var sorted_ranges = self.ranges;
        std.mem.sort(f64, &sorted_ranges, {}, std.sort.asc(f64));
        const median_range = sorted_ranges[2];

        if (snapshot.sample_count == 1) {
            self.average_price = snapshot.price;
        } else {
            self.average_price = try corona_mod.checked(0.1 * snapshot.price + 0.9 * self.average_price);
            if (self.average_price != 0.0 or snapshot.maximum_power > 0.0) {
                self.signal = try corona_mod.checked(0.2 * @sqrt(snapshot.maximum_power) + 0.9 * self.signal);
            }
            if (self.average_price != 0.0 and snapshot.sample_count > 5) {
                self.noise = try corona_mod.checked(0.1 * median_range + 0.9 * self.noise);
            }
        }

        var physical_decibels: f64 = undefined;
        if (self.signal > 0.0 and self.noise > 0.0) {
            physical_decibels = 20.0 * (@log10(self.signal) - @log10(self.noise));
        } else if (self.signal > 0.0) {
            physical_decibels = math.inf(f64);
        } else if (self.noise > 0.0) {
            physical_decibels = -math.inf(f64);
        } else {
            physical_decibels = math.nan(f64);
        }

        const calibrated_decibels = physical_decibels + 3.5;
        const position = if (math.isNan(calibrated_decibels))
            0.0
        else
            corona_mod.clip(calibrated_decibels, 0.0, 10.0) / 10.0;
        const width = @max(0.0, 0.2 - 0.4 * position);
        self.raster.update(position, width, .snr, position > 0.5);

        if (!snapshot.primed) {
            return self.emptySnapshot(snapshot.time);
        }

        const mapped = self.raster.lower + (self.raster.upper - self.raster.lower) * position;
        return .{
            .heatmap = self.raster.column(snapshot.time),
            .signal_to_noise_ratio = mapped,
            .physical_decibels = physical_decibels,
            .calibrated_decibels = calibrated_decibels,
            .normalized_position = position,
            .signal = self.signal,
            .noise = self.noise,
            .median_bar_range = median_range,
            .width = width,
            .dominant_cycle = snapshot.dominant_cycle,
            .dominant_cycle_median = snapshot.dominant_cycle_median,
            .has_cycle_estimate = if (snapshot.has_cycle_estimate) 1.0 else 0.0,
            .ratio_is_finite = if (corona_mod.validateFinite(physical_decibels)) 1.0 else 0.0,
        };
    }

    pub fn updateFromSnapshotRange(
        self: *CoronaSignalToNoiseRatio,
        snapshot: CoronaSnapshot,
        bar_range: f64,
    ) CoronaSignalToNoiseRatioSnapshot {
        return self.tryUpdateFromSnapshotRange(snapshot, bar_range) catch |err| {
            std.debug.panic("CoronaSignalToNoiseRatio.updateFromSnapshotRange failed: {s}", .{@errorName(err)});
        };
    }

    pub fn tryUpdateFromSnapshot(
        self: *CoronaSignalToNoiseRatio,
        snapshot: CoronaSnapshot,
        low: f64,
        high: f64,
    ) Error!CoronaSignalToNoiseRatioSnapshot {
        if (!snapshot.accepted) {
            return self.tryUpdateFromSnapshotRange(snapshot, 0.0);
        }
        const range_value = try validateHighLow(low, high);
        return self.tryUpdateFromSnapshotRange(snapshot, range_value);
    }

    pub fn updateFromSnapshot(
        self: *CoronaSignalToNoiseRatio,
        snapshot: CoronaSnapshot,
        low: f64,
        high: f64,
    ) CoronaSignalToNoiseRatioSnapshot {
        return self.tryUpdateFromSnapshot(snapshot, low, high) catch |err| {
            std.debug.panic("CoronaSignalToNoiseRatio.updateFromSnapshot failed: {s}", .{@errorName(err)});
        };
    }

    pub fn tryUpdateRange(
        self: *CoronaSignalToNoiseRatio,
        sample: f64,
        bar_range: f64,
        time: i64,
    ) Error!CoronaSignalToNoiseRatioSnapshot {
        const ok = try corona_mod.validPrice(sample);
        if (!ok) return self.emptySnapshot(time);
        const range_value = try validateBarRange(bar_range);
        if (math.isNan(range_value)) return self.emptySnapshot(time);
        const engine_snap = try self.engine.tryUpdate(sample, time);
        return self.tryUpdateFromSnapshotRange(engine_snap, range_value);
    }

    pub fn updateRange(
        self: *CoronaSignalToNoiseRatio,
        sample: f64,
        bar_range: f64,
        time: i64,
    ) CoronaSignalToNoiseRatioSnapshot {
        return self.tryUpdateRange(sample, bar_range, time) catch |err| {
            std.debug.panic("CoronaSignalToNoiseRatio.updateRange failed: {s}", .{@errorName(err)});
        };
    }

    pub fn tryUpdate(
        self: *CoronaSignalToNoiseRatio,
        sample: f64,
        low: f64,
        high: f64,
        time: i64,
    ) Error!CoronaSignalToNoiseRatioSnapshot {
        const ok = try corona_mod.validPrice(sample);
        if (!ok) return self.emptySnapshot(time);
        const range_value = try validateHighLow(low, high);
        if (math.isNan(range_value)) return self.emptySnapshot(time);
        const engine_snap = try self.engine.tryUpdate(sample, time);
        return self.tryUpdateFromSnapshotRange(engine_snap, range_value);
    }

    pub fn update(
        self: *CoronaSignalToNoiseRatio,
        sample: f64,
        low: f64,
        high: f64,
        time: i64,
    ) CoronaSignalToNoiseRatioSnapshot {
        return self.tryUpdate(sample, low, high, time) catch |err| {
            std.debug.panic("CoronaSignalToNoiseRatio.update failed: {s}", .{@errorName(err)});
        };
    }

    pub const updateSample = update;

    fn wrapEntity(time: i64, snapshot: CoronaSignalToNoiseRatioSnapshot) OutputArray {
        var out = OutputArray{};
        out.append(.{ .heatmap = snapshot.heatmap });
        out.append(.{ .scalar = .{ .time = time, .value = snapshot.signal_to_noise_ratio } });
        return out;
    }

    pub fn updateScalar(self: *CoronaSignalToNoiseRatio, sample: *const Scalar) OutputArray {
        return wrapEntity(sample.time, self.update(sample.value, sample.value, sample.value, sample.time));
    }

    pub fn updateBar(self: *CoronaSignalToNoiseRatio, sample: *const Bar) OutputArray {
        return wrapEntity(sample.time, self.update(self.bar_func(sample.*), sample.low, sample.high, sample.time));
    }

    pub fn updateQuote(self: *CoronaSignalToNoiseRatio, sample: *const Quote) OutputArray {
        const value = self.quote_func(sample.*);
        return wrapEntity(sample.time, self.update(value, value, value, sample.time));
    }

    pub fn updateTrade(self: *CoronaSignalToNoiseRatio, sample: *const Trade) OutputArray {
        const value = self.trade_func(sample.*);
        return wrapEntity(sample.time, self.update(value, value, value, sample.time));
    }

    pub fn isPrimed(self: *const CoronaSignalToNoiseRatio) bool {
        return self.sample_count >= corona_mod.first_bank_sample;
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

fn verifySnrCase(
    comptime raster_len: usize,
    hp_cutoff: i32,
    min_period: i32,
    max_period: i32,
    expected_heatmap: *const [252][raster_len]f64,
    expected_min: *const [252]f64,
    expected_max: *const [252]f64,
    expected_snr: *const [252]f64,
    expected_phys: *const [252]f64,
    expected_cal: *const [252]f64,
    expected_norm: *const [252]f64,
    expected_signal: *const [252]f64,
    expected_noise: *const [252]f64,
    expected_median_range: *const [252]f64,
    expected_width: *const [252]f64,
    expected_dc: *const [252]f64,
    expected_dcm: *const [252]f64,
    expected_has_est: *const [252]f64,
    expected_finite: *const [252]f64,
) !void {
    var instance = try CoronaSignalToNoiseRatio.init(testing.allocator, .{
        .raster_length = @intCast(raster_len),
        .high_pass_filter_cutoff = hp_cutoff,
        .minimal_period = min_period,
        .maximal_period = max_period,
    });
    defer instance.deinit();

    const step = 10.0 / @as(f64, @floatFromInt(raster_len));
    const expected_first = 1.0 + step;
    const expected_resolution = 1.0 / step;
    const prices = testdata.testInput();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();

    for (prices, 0..) |price, index| {
        const snap = instance.update(price, lows[index], highs[index], @intCast(index));
        try expectClose(snap.heatmap.parameter_first, expected_first);
        try expectClose(snap.heatmap.parameter_last, 11.0);
        try expectClose(snap.heatmap.parameter_resolution, expected_resolution);

        const row_empty = math.isNan(expected_snr[index]);
        try testing.expectEqual(!row_empty, instance.isPrimed());
        if (row_empty) {
            try testing.expect(snap.heatmap.isEmpty());
        } else {
            const values = snap.heatmap.valuesSlice();
            try testing.expectEqual(raster_len, values.len);
            for (0..raster_len) |bin_index| {
                try expectClose(values[bin_index], expected_heatmap[index][bin_index]);
            }
        }

        try expectClose(snap.heatmap.value_min, expected_min[index]);
        try expectClose(snap.heatmap.value_max, expected_max[index]);
        try expectClose(snap.signal_to_noise_ratio, expected_snr[index]);
        try expectClose(snap.physical_decibels, expected_phys[index]);
        try expectClose(snap.calibrated_decibels, expected_cal[index]);
        try expectClose(snap.normalized_position, expected_norm[index]);
        try expectClose(snap.signal, expected_signal[index]);
        try expectClose(snap.noise, expected_noise[index]);
        try expectClose(snap.median_bar_range, expected_median_range[index]);
        try expectClose(snap.width, expected_width[index]);
        try expectClose(snap.dominant_cycle, expected_dc[index]);
        try expectClose(snap.dominant_cycle_median, expected_dcm[index]);
        try expectClose(snap.has_cycle_estimate, expected_has_est[index]);
        try expectClose(snap.ratio_is_finite, expected_finite[index]);
    }

    instance.reset();
    try testing.expect(!instance.isPrimed());
}

test "CoronaSignalToNoiseRatio matches all 16 reference cases" {
    try verifySnrCase(50, 30, 6, 30, &testdata.expectedH30_PMIN6_PMAX30_R50_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX30_R50_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH30_PMIN6_PMAX30_R50_PHYSICAL_DECIBELS(), &testdata.expectedH30_PMIN6_PMAX30_R50_CALIBRATED_DECIBELS(), &testdata.expectedH30_PMIN6_PMAX30_R50_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN6_PMAX30_R50_SIGNAL(), &testdata.expectedH30_PMIN6_PMAX30_R50_NOISE(), &testdata.expectedH30_PMIN6_PMAX30_R50_MEDIAN_BAR_RANGE(), &testdata.expectedH30_PMIN6_PMAX30_R50_WIDTH(), &testdata.expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX30_R50_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN6_PMAX30_R50_RATIO_IS_FINITE());
    try verifySnrCase(25, 30, 6, 30, &testdata.expectedH30_PMIN6_PMAX30_R25_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX30_R25_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH30_PMIN6_PMAX30_R25_PHYSICAL_DECIBELS(), &testdata.expectedH30_PMIN6_PMAX30_R25_CALIBRATED_DECIBELS(), &testdata.expectedH30_PMIN6_PMAX30_R25_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN6_PMAX30_R25_SIGNAL(), &testdata.expectedH30_PMIN6_PMAX30_R25_NOISE(), &testdata.expectedH30_PMIN6_PMAX30_R25_MEDIAN_BAR_RANGE(), &testdata.expectedH30_PMIN6_PMAX30_R25_WIDTH(), &testdata.expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX30_R25_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN6_PMAX30_R25_RATIO_IS_FINITE());
    try verifySnrCase(50, 30, 6, 20, &testdata.expectedH30_PMIN6_PMAX20_R50_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX20_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX20_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX20_R50_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH30_PMIN6_PMAX20_R50_PHYSICAL_DECIBELS(), &testdata.expectedH30_PMIN6_PMAX20_R50_CALIBRATED_DECIBELS(), &testdata.expectedH30_PMIN6_PMAX20_R50_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN6_PMAX20_R50_SIGNAL(), &testdata.expectedH30_PMIN6_PMAX20_R50_NOISE(), &testdata.expectedH30_PMIN6_PMAX20_R50_MEDIAN_BAR_RANGE(), &testdata.expectedH30_PMIN6_PMAX20_R50_WIDTH(), &testdata.expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX20_R50_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN6_PMAX20_R50_RATIO_IS_FINITE());
    try verifySnrCase(25, 30, 6, 20, &testdata.expectedH30_PMIN6_PMAX20_R25_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX20_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX20_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX20_R25_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH30_PMIN6_PMAX20_R25_PHYSICAL_DECIBELS(), &testdata.expectedH30_PMIN6_PMAX20_R25_CALIBRATED_DECIBELS(), &testdata.expectedH30_PMIN6_PMAX20_R25_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN6_PMAX20_R25_SIGNAL(), &testdata.expectedH30_PMIN6_PMAX20_R25_NOISE(), &testdata.expectedH30_PMIN6_PMAX20_R25_MEDIAN_BAR_RANGE(), &testdata.expectedH30_PMIN6_PMAX20_R25_WIDTH(), &testdata.expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX20_R25_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN6_PMAX20_R25_RATIO_IS_FINITE());
    try verifySnrCase(50, 30, 10, 30, &testdata.expectedH30_PMIN10_PMAX30_R50_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX30_R50_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH30_PMIN10_PMAX30_R50_PHYSICAL_DECIBELS(), &testdata.expectedH30_PMIN10_PMAX30_R50_CALIBRATED_DECIBELS(), &testdata.expectedH30_PMIN10_PMAX30_R50_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN10_PMAX30_R50_SIGNAL(), &testdata.expectedH30_PMIN10_PMAX30_R50_NOISE(), &testdata.expectedH30_PMIN10_PMAX30_R50_MEDIAN_BAR_RANGE(), &testdata.expectedH30_PMIN10_PMAX30_R50_WIDTH(), &testdata.expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX30_R50_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN10_PMAX30_R50_RATIO_IS_FINITE());
    try verifySnrCase(25, 30, 10, 30, &testdata.expectedH30_PMIN10_PMAX30_R25_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX30_R25_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH30_PMIN10_PMAX30_R25_PHYSICAL_DECIBELS(), &testdata.expectedH30_PMIN10_PMAX30_R25_CALIBRATED_DECIBELS(), &testdata.expectedH30_PMIN10_PMAX30_R25_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN10_PMAX30_R25_SIGNAL(), &testdata.expectedH30_PMIN10_PMAX30_R25_NOISE(), &testdata.expectedH30_PMIN10_PMAX30_R25_MEDIAN_BAR_RANGE(), &testdata.expectedH30_PMIN10_PMAX30_R25_WIDTH(), &testdata.expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX30_R25_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN10_PMAX30_R25_RATIO_IS_FINITE());
    try verifySnrCase(50, 30, 10, 48, &testdata.expectedH30_PMIN10_PMAX48_R50_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX48_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX48_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX48_R50_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH30_PMIN10_PMAX48_R50_PHYSICAL_DECIBELS(), &testdata.expectedH30_PMIN10_PMAX48_R50_CALIBRATED_DECIBELS(), &testdata.expectedH30_PMIN10_PMAX48_R50_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN10_PMAX48_R50_SIGNAL(), &testdata.expectedH30_PMIN10_PMAX48_R50_NOISE(), &testdata.expectedH30_PMIN10_PMAX48_R50_MEDIAN_BAR_RANGE(), &testdata.expectedH30_PMIN10_PMAX48_R50_WIDTH(), &testdata.expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX48_R50_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN10_PMAX48_R50_RATIO_IS_FINITE());
    try verifySnrCase(25, 30, 10, 48, &testdata.expectedH30_PMIN10_PMAX48_R25_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX48_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX48_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX48_R25_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH30_PMIN10_PMAX48_R25_PHYSICAL_DECIBELS(), &testdata.expectedH30_PMIN10_PMAX48_R25_CALIBRATED_DECIBELS(), &testdata.expectedH30_PMIN10_PMAX48_R25_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN10_PMAX48_R25_SIGNAL(), &testdata.expectedH30_PMIN10_PMAX48_R25_NOISE(), &testdata.expectedH30_PMIN10_PMAX48_R25_MEDIAN_BAR_RANGE(), &testdata.expectedH30_PMIN10_PMAX48_R25_WIDTH(), &testdata.expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX48_R25_HAS_CYCLE_ESTIMATE(), &testdata.expectedH30_PMIN10_PMAX48_R25_RATIO_IS_FINITE());
    try verifySnrCase(50, 48, 6, 30, &testdata.expectedH48_PMIN6_PMAX30_R50_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX30_R50_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH48_PMIN6_PMAX30_R50_PHYSICAL_DECIBELS(), &testdata.expectedH48_PMIN6_PMAX30_R50_CALIBRATED_DECIBELS(), &testdata.expectedH48_PMIN6_PMAX30_R50_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN6_PMAX30_R50_SIGNAL(), &testdata.expectedH48_PMIN6_PMAX30_R50_NOISE(), &testdata.expectedH48_PMIN6_PMAX30_R50_MEDIAN_BAR_RANGE(), &testdata.expectedH48_PMIN6_PMAX30_R50_WIDTH(), &testdata.expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX30_R50_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN6_PMAX30_R50_RATIO_IS_FINITE());
    try verifySnrCase(25, 48, 6, 30, &testdata.expectedH48_PMIN6_PMAX30_R25_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX30_R25_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH48_PMIN6_PMAX30_R25_PHYSICAL_DECIBELS(), &testdata.expectedH48_PMIN6_PMAX30_R25_CALIBRATED_DECIBELS(), &testdata.expectedH48_PMIN6_PMAX30_R25_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN6_PMAX30_R25_SIGNAL(), &testdata.expectedH48_PMIN6_PMAX30_R25_NOISE(), &testdata.expectedH48_PMIN6_PMAX30_R25_MEDIAN_BAR_RANGE(), &testdata.expectedH48_PMIN6_PMAX30_R25_WIDTH(), &testdata.expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX30_R25_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN6_PMAX30_R25_RATIO_IS_FINITE());
    try verifySnrCase(50, 48, 6, 20, &testdata.expectedH48_PMIN6_PMAX20_R50_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX20_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX20_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX20_R50_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH48_PMIN6_PMAX20_R50_PHYSICAL_DECIBELS(), &testdata.expectedH48_PMIN6_PMAX20_R50_CALIBRATED_DECIBELS(), &testdata.expectedH48_PMIN6_PMAX20_R50_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN6_PMAX20_R50_SIGNAL(), &testdata.expectedH48_PMIN6_PMAX20_R50_NOISE(), &testdata.expectedH48_PMIN6_PMAX20_R50_MEDIAN_BAR_RANGE(), &testdata.expectedH48_PMIN6_PMAX20_R50_WIDTH(), &testdata.expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX20_R50_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN6_PMAX20_R50_RATIO_IS_FINITE());
    try verifySnrCase(25, 48, 6, 20, &testdata.expectedH48_PMIN6_PMAX20_R25_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX20_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX20_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX20_R25_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH48_PMIN6_PMAX20_R25_PHYSICAL_DECIBELS(), &testdata.expectedH48_PMIN6_PMAX20_R25_CALIBRATED_DECIBELS(), &testdata.expectedH48_PMIN6_PMAX20_R25_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN6_PMAX20_R25_SIGNAL(), &testdata.expectedH48_PMIN6_PMAX20_R25_NOISE(), &testdata.expectedH48_PMIN6_PMAX20_R25_MEDIAN_BAR_RANGE(), &testdata.expectedH48_PMIN6_PMAX20_R25_WIDTH(), &testdata.expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX20_R25_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN6_PMAX20_R25_RATIO_IS_FINITE());
    try verifySnrCase(50, 48, 10, 30, &testdata.expectedH48_PMIN10_PMAX30_R50_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX30_R50_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH48_PMIN10_PMAX30_R50_PHYSICAL_DECIBELS(), &testdata.expectedH48_PMIN10_PMAX30_R50_CALIBRATED_DECIBELS(), &testdata.expectedH48_PMIN10_PMAX30_R50_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN10_PMAX30_R50_SIGNAL(), &testdata.expectedH48_PMIN10_PMAX30_R50_NOISE(), &testdata.expectedH48_PMIN10_PMAX30_R50_MEDIAN_BAR_RANGE(), &testdata.expectedH48_PMIN10_PMAX30_R50_WIDTH(), &testdata.expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX30_R50_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN10_PMAX30_R50_RATIO_IS_FINITE());
    try verifySnrCase(25, 48, 10, 30, &testdata.expectedH48_PMIN10_PMAX30_R25_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX30_R25_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH48_PMIN10_PMAX30_R25_PHYSICAL_DECIBELS(), &testdata.expectedH48_PMIN10_PMAX30_R25_CALIBRATED_DECIBELS(), &testdata.expectedH48_PMIN10_PMAX30_R25_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN10_PMAX30_R25_SIGNAL(), &testdata.expectedH48_PMIN10_PMAX30_R25_NOISE(), &testdata.expectedH48_PMIN10_PMAX30_R25_MEDIAN_BAR_RANGE(), &testdata.expectedH48_PMIN10_PMAX30_R25_WIDTH(), &testdata.expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX30_R25_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN10_PMAX30_R25_RATIO_IS_FINITE());
    try verifySnrCase(50, 48, 10, 48, &testdata.expectedH48_PMIN10_PMAX48_R50_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX48_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX48_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX48_R50_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH48_PMIN10_PMAX48_R50_PHYSICAL_DECIBELS(), &testdata.expectedH48_PMIN10_PMAX48_R50_CALIBRATED_DECIBELS(), &testdata.expectedH48_PMIN10_PMAX48_R50_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN10_PMAX48_R50_SIGNAL(), &testdata.expectedH48_PMIN10_PMAX48_R50_NOISE(), &testdata.expectedH48_PMIN10_PMAX48_R50_MEDIAN_BAR_RANGE(), &testdata.expectedH48_PMIN10_PMAX48_R50_WIDTH(), &testdata.expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX48_R50_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN10_PMAX48_R50_RATIO_IS_FINITE());
    try verifySnrCase(25, 48, 10, 48, &testdata.expectedH48_PMIN10_PMAX48_R25_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX48_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX48_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX48_R25_SIGNAL_TO_NOISE_RATIO(), &testdata.expectedH48_PMIN10_PMAX48_R25_PHYSICAL_DECIBELS(), &testdata.expectedH48_PMIN10_PMAX48_R25_CALIBRATED_DECIBELS(), &testdata.expectedH48_PMIN10_PMAX48_R25_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN10_PMAX48_R25_SIGNAL(), &testdata.expectedH48_PMIN10_PMAX48_R25_NOISE(), &testdata.expectedH48_PMIN10_PMAX48_R25_MEDIAN_BAR_RANGE(), &testdata.expectedH48_PMIN10_PMAX48_R25_WIDTH(), &testdata.expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX48_R25_HAS_CYCLE_ESTIMATE(), &testdata.expectedH48_PMIN10_PMAX48_R25_RATIO_IS_FINITE());
}

test "CoronaSignalToNoiseRatio metadata, mnemonics, negative min, validation, and entities" {
    var instance = try CoronaSignalToNoiseRatio.init(testing.allocator, defaultParams());
    defer instance.deinit();

    var md: Metadata = undefined;
    instance.getMetadata(&md);
    try testing.expectEqual(Identifier.corona_signal_to_noise_ratio, md.identifier);
    try testing.expectEqualStrings("csnr(50, 20, 1, 11, 30)", md.mnemonic);
    try testing.expectEqual(@as(usize, 2), md.outputs_len);
    try testing.expectEqual(shape_mod.Shape.heatmap, md.outputs_buf[0].shape);
    try testing.expectEqual(shape_mod.Shape.scalar, md.outputs_buf[1].shape);
    try testing.expectEqualStrings("csnr-snr(30)", md.outputs_buf[1].mnemonic);

    var neg_bounds = try CoronaSignalToNoiseRatio.init(testing.allocator, .{
        .min_parameter_value = -5.0,
        .max_parameter_value = 5.0,
    });
    defer neg_bounds.deinit();
    try testing.expectEqualStrings("csnr(50, 20, -5, 5, 30)", neg_bounds.mnemonic());

    var flagged = try CoronaSignalToNoiseRatio.init(testing.allocator, .{
        .minimal_period = 10,
        .maximal_period = 48,
        .bar_component = .median,
    });
    defer flagged.deinit();
    var flagged_md: Metadata = undefined;
    flagged.getMetadata(&flagged_md);
    try testing.expectEqualStrings("csnr(50, 20, 1, 11, 30, pmin=10, pmax=48, hl/2)", flagged_md.mnemonic);
    try testing.expectEqualStrings("csnr-snr(30, pmin=10, pmax=48, hl/2)", flagged_md.outputs_buf[1].mnemonic);

    try testing.expectError(error.InvalidRasterLength, CoronaSignalToNoiseRatio.init(testing.allocator, .{ .raster_length = 1 }));
    try testing.expectError(error.InvalidMaxRasterValue, CoronaSignalToNoiseRatio.init(testing.allocator, .{ .max_raster_value = 6.0 }));
    try testing.expectError(error.InvalidMaxParameterValue, CoronaSignalToNoiseRatio.init(testing.allocator, .{ .min_parameter_value = 11.0, .max_parameter_value = 11.0 }));
    try testing.expectError(error.InvalidHighPassFilterCutoff, CoronaSignalToNoiseRatio.init(testing.allocator, .{ .high_pass_filter_cutoff = 4 }));
    try testing.expectError(error.InvalidMinimalPeriod, CoronaSignalToNoiseRatio.init(testing.allocator, .{ .minimal_period = 5 }));
    try testing.expectError(error.InvalidMaximalPeriod, CoronaSignalToNoiseRatio.init(testing.allocator, .{ .minimal_period = 10, .maximal_period = 10 }));
    try testing.expectError(error.InvalidBarRange, instance.tryUpdate(100.0, 101.0, 99.0, 0));
    try testing.expectError(error.InvalidBarRange, instance.tryUpdateRange(100.0, -1.0, 0));
    try testing.expect(!instance.isPrimed());

    const prices = testdata.testInput();
    const opens = testdata.testOpen();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();
    var bar_ind = try CoronaSignalToNoiseRatio.init(testing.allocator, defaultParams());
    defer bar_ind.deinit();
    var scalar_ind = try CoronaSignalToNoiseRatio.init(testing.allocator, defaultParams());
    defer scalar_ind.deinit();
    var quote_ind = try CoronaSignalToNoiseRatio.init(testing.allocator, defaultParams());
    defer quote_ind.deinit();
    var trade_ind = try CoronaSignalToNoiseRatio.init(testing.allocator, defaultParams());
    defer trade_ind.deinit();
    for (0..15) |index| {
        const bar = Bar{ .time = @intCast(index), .open = opens[index], .high = highs[index], .low = lows[index], .close = prices[index], .volume = 1000.0 };
        const scalar = Scalar{ .time = @intCast(index), .value = prices[index] };
        const quote = Quote{ .time = @intCast(index), .bid_price = prices[index] - 0.5, .ask_price = prices[index] + 0.5, .bid_size = 10.0, .ask_size = 10.0 };
        const trade = Trade{ .time = @intCast(index), .price = prices[index], .volume = 10.0 };
        try testing.expectEqual(@as(usize, 2), bar_ind.updateBar(&bar).len);
        try testing.expectEqual(@as(usize, 2), scalar_ind.updateScalar(&scalar).len);
        try testing.expectEqual(@as(usize, 2), quote_ind.updateQuote(&quote).len);
        try testing.expectEqual(@as(usize, 2), trade_ind.updateTrade(&trade).len);
    }
    const final_scalar = Scalar{ .time = 15, .value = 100.0 };
    const final_out = scalar_ind.updateScalar(&final_scalar);
    try testing.expectEqual(@as(f64, 11.0), final_out.slice()[1].scalar.value);
}
