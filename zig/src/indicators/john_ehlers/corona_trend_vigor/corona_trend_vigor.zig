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
const AdaptiveBandPass = corona_mod.AdaptiveBandPass;
const Raster = corona_mod.Raster;

pub const CoronaTrendVigorOutput = enum(u8) {
    value = 1,
    trend_vigor = 2,
};

pub const Params = struct {
    raster_length: i32 = 50,
    max_raster_value: f64 = 20.0,
    min_parameter_value: f64 = -10.0,
    max_parameter_value: f64 = 10.0,
    high_pass_filter_cutoff: i32 = 30,
    minimal_period: i32 = 6,
    maximal_period: i32 = 30,
    lookback_offset: i32 = 0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub fn defaultParams() Params {
    return .{};
}

pub const Error = corona_mod.Error;

pub const CoronaTrendVigorSnapshot = struct {
    heatmap: Heatmap,
    trend_vigor: f64,
    normalized_vigor: f64,
    dominant_cycle: f64,
    dominant_cycle_median: f64,
    band_pass: f64,
    quadrature: f64,
    cycle_amplitude: f64,
    trend_displacement: f64,
    raw_ratio: f64,
    lookback: f64,
    width: f64,
    has_cycle_estimate: f64,
};

pub const Snapshot = CoronaTrendVigorSnapshot;

fn formatTrendFlags(buffer: []u8, params: Params) Error![]const u8 {
    var pos: usize = 0;
    if (params.minimal_period != 6) {
        const chunk = std.fmt.bufPrint(buffer[pos..], ", pmin={d}", .{params.minimal_period}) catch return error.MnemonicTooLong;
        pos += chunk.len;
    }
    if (params.maximal_period != 30) {
        const chunk = std.fmt.bufPrint(buffer[pos..], ", pmax={d}", .{params.maximal_period}) catch return error.MnemonicTooLong;
        pos += chunk.len;
    }
    if (params.lookback_offset != 0) {
        const chunk = std.fmt.bufPrint(buffer[pos..], ", offset={d}", .{params.lookback_offset}) catch return error.MnemonicTooLong;
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
    var flags_buffer: [96]u8 = undefined;
    const flags = try formatTrendFlags(&flags_buffer, params);

    return std.fmt.bufPrint(buffer, "ctv({d}, {d}, {d}, {d}, {d}{s}{s})", .{
        params.raster_length,
        corona_mod.normalizeZero(params.max_raster_value),
        corona_mod.normalizeZero(params.min_parameter_value),
        corona_mod.normalizeZero(params.max_parameter_value),
        params.high_pass_filter_cutoff,
        flags,
        triple,
    }) catch return error.MnemonicTooLong;
}

pub const CoronaTrendVigor = struct {
    allocator: std.mem.Allocator,
    engine: Corona,
    engine_params: corona_mod.Params,
    bandpass: AdaptiveBandPass,
    raster: Raster,
    lookback_offset: i32,
    sample_count: usize,
    prices: []f64,
    prices_len: usize,
    ratio: f64,
    ready: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_storage: [256]u8,
    mnemonic_len: usize,
    description_storage: [320]u8,
    description_len: usize,
    mnemonic_tv_storage: [160]u8,
    mnemonic_tv_len: usize,
    description_tv_storage: [256]u8,
    description_tv_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!CoronaTrendVigor {
        if (params.lookback_offset != 0 and params.lookback_offset != -1) {
            return error.InvalidLookbackOffset;
        }
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

        const prices_cap: usize = @intCast(params.maximal_period + 1);
        const prices = allocator.alloc(f64, prices_cap) catch return error.OutOfMemory;
        errdefer allocator.free(prices);

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buffer: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buffer, bc, qc, tc);
        var flags_buffer: [96]u8 = undefined;
        const flags = try formatTrendFlags(&flags_buffer, params);

        var mnemonic_storage: [256]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_storage, "ctv({d}, {d}, {d}, {d}, {d}{s}{s})", .{
            raster.length,
            corona_mod.normalizeZero(raster.ceiling),
            corona_mod.normalizeZero(raster.lower),
            corona_mod.normalizeZero(raster.upper),
            params.high_pass_filter_cutoff,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;

        var description_storage: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Corona trend vigor {s}", .{mn}) catch
            return error.MnemonicTooLong;

        var mnemonic_tv_storage: [160]u8 = undefined;
        const mn_tv = std.fmt.bufPrint(&mnemonic_tv_storage, "ctv-tv({d}{s}{s})", .{
            params.high_pass_filter_cutoff,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;

        var description_tv_storage: [256]u8 = undefined;
        const desc_tv = std.fmt.bufPrint(&description_tv_storage, "Corona trend vigor scalar {s}", .{mn_tv}) catch
            return error.MnemonicTooLong;

        return .{
            .allocator = allocator,
            .engine = engine,
            .engine_params = engine.parameters,
            .bandpass = .{},
            .raster = raster,
            .lookback_offset = params.lookback_offset,
            .sample_count = 0,
            .prices = prices,
            .prices_len = 0,
            .ratio = 0.0,
            .ready = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mn.len,
            .description_storage = description_storage,
            .description_len = desc.len,
            .mnemonic_tv_storage = mnemonic_tv_storage,
            .mnemonic_tv_len = mn_tv.len,
            .description_tv_storage = description_tv_storage,
            .description_tv_len = desc_tv.len,
        };
    }

    pub fn deinit(self: *CoronaTrendVigor) void {
        self.engine.deinit();
        self.raster.deinit();
        self.allocator.free(self.prices);
    }

    pub fn reset(self: *CoronaTrendVigor) void {
        self.engine.reset();
        self.bandpass.reset();
        self.raster.reset();
        self.sample_count = 0;
        self.prices_len = 0;
        self.ratio = 0.0;
        self.ready = false;
    }

    pub fn fixSlices(self: *CoronaTrendVigor) void {
        _ = self;
    }

    pub fn mnemonic(self: *const CoronaTrendVigor) []const u8 {
        return self.mnemonic_storage[0..self.mnemonic_len];
    }

    fn description(self: *const CoronaTrendVigor) []const u8 {
        return self.description_storage[0..self.description_len];
    }

    fn mnemonicTV(self: *const CoronaTrendVigor) []const u8 {
        return self.mnemonic_tv_storage[0..self.mnemonic_tv_len];
    }

    fn descriptionTV(self: *const CoronaTrendVigor) []const u8 {
        return self.description_tv_storage[0..self.description_tv_len];
    }

    fn emptySnapshot(self: *const CoronaTrendVigor, time: i64) CoronaTrendVigorSnapshot {
        const nan = math.nan(f64);
        return .{
            .heatmap = self.raster.empty(time),
            .trend_vigor = nan,
            .normalized_vigor = nan,
            .dominant_cycle = nan,
            .dominant_cycle_median = nan,
            .band_pass = nan,
            .quadrature = nan,
            .cycle_amplitude = nan,
            .trend_displacement = nan,
            .raw_ratio = nan,
            .lookback = nan,
            .width = nan,
            .has_cycle_estimate = nan,
        };
    }

    pub fn tryUpdateFromSnapshot(self: *CoronaTrendVigor, snapshot: CoronaSnapshot) Error!CoronaTrendVigorSnapshot {
        const accepted = try corona_mod.acceptSnapshot(snapshot, self.engine_params, self.sample_count + 1);
        if (!accepted) {
            return self.emptySnapshot(snapshot.time);
        }
        const cycle = try self.bandpass.update(snapshot.price, snapshot.dominant_cycle_median);
        self.sample_count = snapshot.sample_count;

        _ = corona_mod.rollingAppend(self.prices, &self.prices_len, snapshot.price);

        const floored_dcm: i32 = @intFromFloat(@floor(snapshot.dominant_cycle_median));
        const lag_signed = @max(@as(i32, 1), floored_dcm + self.lookback_offset);
        const lag: usize = @intCast(lag_signed);
        const has_history = self.prices_len > lag;
        const trend = if (has_history) snapshot.price - self.prices[self.prices_len - 1 - lag] else math.nan(f64);
        const raw_ratio = if (has_history and cycle.amplitude > 0.0) trend / cycle.amplitude else math.nan(f64);
        if (has_history and trend != 0.0 and cycle.amplitude > 0.0) {
            const smoothed = try corona_mod.checked(0.33 * raw_ratio + 0.67 * self.ratio);
            self.ratio = corona_mod.clip(smoothed, -10.0, 10.0);
        }

        const vigor = 0.05 * (self.ratio + 10.0);
        var width: f64 = if (0.3 <= vigor and vigor <= 0.7) @min(vigor - 0.3, 0.7 - vigor) else 0.01;
        if (width == 0.0) {
            width = 0.01;
        }

        self.raster.update(vigor, width, .trend, vigor < 0.3 or vigor > 0.7);
        self.ready = snapshot.primed and has_history;
        if (!self.ready) {
            return self.emptySnapshot(snapshot.time);
        }

        const mapped = self.raster.lower + (self.raster.upper - self.raster.lower) * vigor;
        return .{
            .heatmap = self.raster.column(snapshot.time),
            .trend_vigor = mapped,
            .normalized_vigor = vigor,
            .dominant_cycle = snapshot.dominant_cycle,
            .dominant_cycle_median = snapshot.dominant_cycle_median,
            .band_pass = cycle.band_pass,
            .quadrature = cycle.quadrature,
            .cycle_amplitude = cycle.amplitude,
            .trend_displacement = trend,
            .raw_ratio = raw_ratio,
            .lookback = @floatFromInt(lag),
            .width = width,
            .has_cycle_estimate = if (snapshot.has_cycle_estimate) 1.0 else 0.0,
        };
    }

    pub fn updateFromSnapshot(self: *CoronaTrendVigor, snapshot: CoronaSnapshot) CoronaTrendVigorSnapshot {
        return self.tryUpdateFromSnapshot(snapshot) catch |err| {
            std.debug.panic("CoronaTrendVigor.updateFromSnapshot failed: {s}", .{@errorName(err)});
        };
    }

    pub fn tryUpdate(self: *CoronaTrendVigor, sample: f64, time: i64) Error!CoronaTrendVigorSnapshot {
        const engine_snap = try self.engine.tryUpdate(sample, time);
        return self.tryUpdateFromSnapshot(engine_snap);
    }

    pub fn update(self: *CoronaTrendVigor, sample: f64, time: i64) CoronaTrendVigorSnapshot {
        return self.tryUpdate(sample, time) catch |err| {
            std.debug.panic("CoronaTrendVigor.update failed: {s}", .{@errorName(err)});
        };
    }

    pub const updateSample = update;

    fn wrapEntity(time: i64, snapshot: CoronaTrendVigorSnapshot) OutputArray {
        var out = OutputArray{};
        out.append(.{ .heatmap = snapshot.heatmap });
        out.append(.{ .scalar = .{ .time = time, .value = snapshot.trend_vigor } });
        return out;
    }

    pub fn updateScalar(self: *CoronaTrendVigor, sample: *const Scalar) OutputArray {
        return wrapEntity(sample.time, self.update(sample.value, sample.time));
    }

    pub fn updateBar(self: *CoronaTrendVigor, sample: *const Bar) OutputArray {
        return wrapEntity(sample.time, self.update(self.bar_func(sample.*), sample.time));
    }

    pub fn updateQuote(self: *CoronaTrendVigor, sample: *const Quote) OutputArray {
        return wrapEntity(sample.time, self.update(self.quote_func(sample.*), sample.time));
    }

    pub fn updateTrade(self: *CoronaTrendVigor, sample: *const Trade) OutputArray {
        return wrapEntity(sample.time, self.update(self.trade_func(sample.*), sample.time));
    }

    pub fn isPrimed(self: *const CoronaTrendVigor) bool {
        return self.ready;
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
const testdata = @import("testdata.zig");

const test_tolerance: f64 = 5e-12;

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

fn verifyTrendCase(
    comptime raster_len: usize,
    hp_cutoff: i32,
    min_period: i32,
    max_period: i32,
    expected_heatmap: *const [252][raster_len]f64,
    expected_min: *const [252]f64,
    expected_max: *const [252]f64,
    expected_tv: *const [252]f64,
    expected_norm: *const [252]f64,
    expected_dc: *const [252]f64,
    expected_dcm: *const [252]f64,
    expected_bp: *const [252]f64,
    expected_quad: *const [252]f64,
    expected_amp: *const [252]f64,
    expected_disp: *const [252]f64,
    expected_raw_ratio: *const [252]f64,
    expected_lookback: *const [252]f64,
    expected_width: *const [252]f64,
    expected_has_est: *const [252]f64,
) !void {
    const p = Params{
        .raster_length = @intCast(raster_len),
        .high_pass_filter_cutoff = hp_cutoff,
        .minimal_period = min_period,
        .maximal_period = max_period,
    };
    var instance = try CoronaTrendVigor.init(testing.allocator, p);
    defer instance.deinit();
    var shared_consumer = try CoronaTrendVigor.init(testing.allocator, p);
    defer shared_consumer.deinit();
    var shared_engine = try Corona.init(testing.allocator, .{
        .high_pass_filter_cutoff = hp_cutoff,
        .minimal_period = min_period,
        .maximal_period = max_period,
        .decibels_lower_threshold = 6.0,
        .decibels_upper_threshold = 20.0,
    });
    defer shared_engine.deinit();

    const step = 20.0 / @as(f64, @floatFromInt(raster_len));
    const expected_first = -10.0 + step;
    const expected_resolution = 1.0 / step;
    const prices = testdata.testInput();
    var saw_lag_tie = false;

    for (prices, 0..) |price, index| {
        const standalone_snap = instance.update(price, @intCast(index));
        var engine_snap = shared_engine.update(price, @intCast(index));
        if (engine_snap.primed) {
            if (!math.isNan(expected_dcm[index])) {
                try expectClose(standalone_snap.dominant_cycle, expected_dc[index]);
                try expectClose(standalone_snap.dominant_cycle_median, expected_dcm[index]);
                try expectClose(standalone_snap.band_pass, expected_bp[index]);
                try expectClose(standalone_snap.quadrature, expected_quad[index]);
                try expectClose(standalone_snap.cycle_amplitude, expected_amp[index]);
                try expectClose(standalone_snap.has_cycle_estimate, expected_has_est[index]);
            }
            const nearest_integer = @round(engine_snap.dominant_cycle_median);
            if (@abs(engine_snap.dominant_cycle_median - nearest_integer) <= 5e-12) {
                if (math.isNan(expected_lookback[index])) {
                    if (@as(f64, @floatFromInt(index + 1)) == nearest_integer and engine_snap.dominant_cycle_median < nearest_integer) {
                        engine_snap.dominant_cycle_median = nearest_integer;
                        saw_lag_tie = true;
                    }
                } else if (@floor(engine_snap.dominant_cycle_median) != expected_lookback[index]) {
                    if (expected_lookback[index] == nearest_integer) {
                        engine_snap.dominant_cycle_median = nearest_integer;
                    } else if (expected_lookback[index] == nearest_integer - 1.0) {
                        engine_snap.dominant_cycle_median = math.nextAfter(f64, nearest_integer, 0.0);
                    }
                    saw_lag_tie = true;
                }
            }
        }
        const snap = shared_consumer.updateFromSnapshot(engine_snap);

        try expectClose(snap.heatmap.parameter_first, expected_first);
        try expectClose(snap.heatmap.parameter_last, 10.0);
        try expectClose(snap.heatmap.parameter_resolution, expected_resolution);

        const row_empty = math.isNan(expected_tv[index]);
        try testing.expectEqual(!row_empty, shared_consumer.isPrimed());
        if (!saw_lag_tie) {
            try testing.expectEqual(!row_empty, instance.isPrimed());
        }
        if (row_empty) {
            try testing.expect(snap.heatmap.isEmpty());
        } else {
            const values = snap.heatmap.valuesSlice();
            try testing.expectEqual(raster_len, values.len);
            for (0..raster_len) |bin_index| {
                try expectClose(values[bin_index], expected_heatmap[index][bin_index]);
                if (!saw_lag_tie) {
                    try expectClose(standalone_snap.heatmap.valuesSlice()[bin_index], expected_heatmap[index][bin_index]);
                }
            }
        }

        try expectClose(snap.heatmap.value_min, expected_min[index]);
        try expectClose(snap.heatmap.value_max, expected_max[index]);
        try expectClose(snap.trend_vigor, expected_tv[index]);
        try expectClose(snap.normalized_vigor, expected_norm[index]);
        try expectClose(snap.dominant_cycle, expected_dc[index]);
        try expectClose(snap.dominant_cycle_median, expected_dcm[index]);
        try expectClose(snap.band_pass, expected_bp[index]);
        try expectClose(snap.quadrature, expected_quad[index]);
        try expectClose(snap.cycle_amplitude, expected_amp[index]);
        try expectClose(snap.trend_displacement, expected_disp[index]);
        try expectClose(snap.raw_ratio, expected_raw_ratio[index]);
        try expectClose(snap.lookback, expected_lookback[index]);
        try expectClose(snap.width, expected_width[index]);
        try expectClose(snap.has_cycle_estimate, expected_has_est[index]);
        if (!saw_lag_tie) {
            try expectClose(standalone_snap.trend_vigor, expected_tv[index]);
        }
    }

    shared_consumer.reset();
    try testing.expect(!shared_consumer.isPrimed());
    instance.reset();
    try testing.expect(!instance.isPrimed());
}

test "CoronaTrendVigor matches all 16 reference cases" {
    try verifyTrendCase(50, 30, 6, 30, &testdata.expectedH30_PMIN6_PMAX30_R50_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX30_R50_TREND_VIGOR(), &testdata.expectedH30_PMIN6_PMAX30_R50_NORMALIZED_VIGOR(), &testdata.expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX30_R50_BAND_PASS(), &testdata.expectedH30_PMIN6_PMAX30_R50_QUADRATURE(), &testdata.expectedH30_PMIN6_PMAX30_R50_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN6_PMAX30_R50_TREND_DISPLACEMENT(), &testdata.expectedH30_PMIN6_PMAX30_R50_RAW_RATIO(), &testdata.expectedH30_PMIN6_PMAX30_R50_LOOKBACK(), &testdata.expectedH30_PMIN6_PMAX30_R50_WIDTH(), &testdata.expectedH30_PMIN6_PMAX30_R50_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(25, 30, 6, 30, &testdata.expectedH30_PMIN6_PMAX30_R25_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX30_R25_TREND_VIGOR(), &testdata.expectedH30_PMIN6_PMAX30_R25_NORMALIZED_VIGOR(), &testdata.expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX30_R25_BAND_PASS(), &testdata.expectedH30_PMIN6_PMAX30_R25_QUADRATURE(), &testdata.expectedH30_PMIN6_PMAX30_R25_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN6_PMAX30_R25_TREND_DISPLACEMENT(), &testdata.expectedH30_PMIN6_PMAX30_R25_RAW_RATIO(), &testdata.expectedH30_PMIN6_PMAX30_R25_LOOKBACK(), &testdata.expectedH30_PMIN6_PMAX30_R25_WIDTH(), &testdata.expectedH30_PMIN6_PMAX30_R25_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(50, 30, 6, 20, &testdata.expectedH30_PMIN6_PMAX20_R50_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX20_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX20_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX20_R50_TREND_VIGOR(), &testdata.expectedH30_PMIN6_PMAX20_R50_NORMALIZED_VIGOR(), &testdata.expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX20_R50_BAND_PASS(), &testdata.expectedH30_PMIN6_PMAX20_R50_QUADRATURE(), &testdata.expectedH30_PMIN6_PMAX20_R50_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN6_PMAX20_R50_TREND_DISPLACEMENT(), &testdata.expectedH30_PMIN6_PMAX20_R50_RAW_RATIO(), &testdata.expectedH30_PMIN6_PMAX20_R50_LOOKBACK(), &testdata.expectedH30_PMIN6_PMAX20_R50_WIDTH(), &testdata.expectedH30_PMIN6_PMAX20_R50_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(25, 30, 6, 20, &testdata.expectedH30_PMIN6_PMAX20_R25_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX20_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX20_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX20_R25_TREND_VIGOR(), &testdata.expectedH30_PMIN6_PMAX20_R25_NORMALIZED_VIGOR(), &testdata.expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX20_R25_BAND_PASS(), &testdata.expectedH30_PMIN6_PMAX20_R25_QUADRATURE(), &testdata.expectedH30_PMIN6_PMAX20_R25_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN6_PMAX20_R25_TREND_DISPLACEMENT(), &testdata.expectedH30_PMIN6_PMAX20_R25_RAW_RATIO(), &testdata.expectedH30_PMIN6_PMAX20_R25_LOOKBACK(), &testdata.expectedH30_PMIN6_PMAX20_R25_WIDTH(), &testdata.expectedH30_PMIN6_PMAX20_R25_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(50, 30, 10, 30, &testdata.expectedH30_PMIN10_PMAX30_R50_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX30_R50_TREND_VIGOR(), &testdata.expectedH30_PMIN10_PMAX30_R50_NORMALIZED_VIGOR(), &testdata.expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX30_R50_BAND_PASS(), &testdata.expectedH30_PMIN10_PMAX30_R50_QUADRATURE(), &testdata.expectedH30_PMIN10_PMAX30_R50_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN10_PMAX30_R50_TREND_DISPLACEMENT(), &testdata.expectedH30_PMIN10_PMAX30_R50_RAW_RATIO(), &testdata.expectedH30_PMIN10_PMAX30_R50_LOOKBACK(), &testdata.expectedH30_PMIN10_PMAX30_R50_WIDTH(), &testdata.expectedH30_PMIN10_PMAX30_R50_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(25, 30, 10, 30, &testdata.expectedH30_PMIN10_PMAX30_R25_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX30_R25_TREND_VIGOR(), &testdata.expectedH30_PMIN10_PMAX30_R25_NORMALIZED_VIGOR(), &testdata.expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX30_R25_BAND_PASS(), &testdata.expectedH30_PMIN10_PMAX30_R25_QUADRATURE(), &testdata.expectedH30_PMIN10_PMAX30_R25_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN10_PMAX30_R25_TREND_DISPLACEMENT(), &testdata.expectedH30_PMIN10_PMAX30_R25_RAW_RATIO(), &testdata.expectedH30_PMIN10_PMAX30_R25_LOOKBACK(), &testdata.expectedH30_PMIN10_PMAX30_R25_WIDTH(), &testdata.expectedH30_PMIN10_PMAX30_R25_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(50, 30, 10, 48, &testdata.expectedH30_PMIN10_PMAX48_R50_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX48_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX48_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX48_R50_TREND_VIGOR(), &testdata.expectedH30_PMIN10_PMAX48_R50_NORMALIZED_VIGOR(), &testdata.expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX48_R50_BAND_PASS(), &testdata.expectedH30_PMIN10_PMAX48_R50_QUADRATURE(), &testdata.expectedH30_PMIN10_PMAX48_R50_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN10_PMAX48_R50_TREND_DISPLACEMENT(), &testdata.expectedH30_PMIN10_PMAX48_R50_RAW_RATIO(), &testdata.expectedH30_PMIN10_PMAX48_R50_LOOKBACK(), &testdata.expectedH30_PMIN10_PMAX48_R50_WIDTH(), &testdata.expectedH30_PMIN10_PMAX48_R50_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(25, 30, 10, 48, &testdata.expectedH30_PMIN10_PMAX48_R25_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX48_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX48_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX48_R25_TREND_VIGOR(), &testdata.expectedH30_PMIN10_PMAX48_R25_NORMALIZED_VIGOR(), &testdata.expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX48_R25_BAND_PASS(), &testdata.expectedH30_PMIN10_PMAX48_R25_QUADRATURE(), &testdata.expectedH30_PMIN10_PMAX48_R25_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN10_PMAX48_R25_TREND_DISPLACEMENT(), &testdata.expectedH30_PMIN10_PMAX48_R25_RAW_RATIO(), &testdata.expectedH30_PMIN10_PMAX48_R25_LOOKBACK(), &testdata.expectedH30_PMIN10_PMAX48_R25_WIDTH(), &testdata.expectedH30_PMIN10_PMAX48_R25_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(50, 48, 6, 30, &testdata.expectedH48_PMIN6_PMAX30_R50_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX30_R50_TREND_VIGOR(), &testdata.expectedH48_PMIN6_PMAX30_R50_NORMALIZED_VIGOR(), &testdata.expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX30_R50_BAND_PASS(), &testdata.expectedH48_PMIN6_PMAX30_R50_QUADRATURE(), &testdata.expectedH48_PMIN6_PMAX30_R50_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN6_PMAX30_R50_TREND_DISPLACEMENT(), &testdata.expectedH48_PMIN6_PMAX30_R50_RAW_RATIO(), &testdata.expectedH48_PMIN6_PMAX30_R50_LOOKBACK(), &testdata.expectedH48_PMIN6_PMAX30_R50_WIDTH(), &testdata.expectedH48_PMIN6_PMAX30_R50_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(25, 48, 6, 30, &testdata.expectedH48_PMIN6_PMAX30_R25_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX30_R25_TREND_VIGOR(), &testdata.expectedH48_PMIN6_PMAX30_R25_NORMALIZED_VIGOR(), &testdata.expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX30_R25_BAND_PASS(), &testdata.expectedH48_PMIN6_PMAX30_R25_QUADRATURE(), &testdata.expectedH48_PMIN6_PMAX30_R25_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN6_PMAX30_R25_TREND_DISPLACEMENT(), &testdata.expectedH48_PMIN6_PMAX30_R25_RAW_RATIO(), &testdata.expectedH48_PMIN6_PMAX30_R25_LOOKBACK(), &testdata.expectedH48_PMIN6_PMAX30_R25_WIDTH(), &testdata.expectedH48_PMIN6_PMAX30_R25_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(50, 48, 6, 20, &testdata.expectedH48_PMIN6_PMAX20_R50_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX20_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX20_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX20_R50_TREND_VIGOR(), &testdata.expectedH48_PMIN6_PMAX20_R50_NORMALIZED_VIGOR(), &testdata.expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX20_R50_BAND_PASS(), &testdata.expectedH48_PMIN6_PMAX20_R50_QUADRATURE(), &testdata.expectedH48_PMIN6_PMAX20_R50_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN6_PMAX20_R50_TREND_DISPLACEMENT(), &testdata.expectedH48_PMIN6_PMAX20_R50_RAW_RATIO(), &testdata.expectedH48_PMIN6_PMAX20_R50_LOOKBACK(), &testdata.expectedH48_PMIN6_PMAX20_R50_WIDTH(), &testdata.expectedH48_PMIN6_PMAX20_R50_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(25, 48, 6, 20, &testdata.expectedH48_PMIN6_PMAX20_R25_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX20_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX20_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX20_R25_TREND_VIGOR(), &testdata.expectedH48_PMIN6_PMAX20_R25_NORMALIZED_VIGOR(), &testdata.expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX20_R25_BAND_PASS(), &testdata.expectedH48_PMIN6_PMAX20_R25_QUADRATURE(), &testdata.expectedH48_PMIN6_PMAX20_R25_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN6_PMAX20_R25_TREND_DISPLACEMENT(), &testdata.expectedH48_PMIN6_PMAX20_R25_RAW_RATIO(), &testdata.expectedH48_PMIN6_PMAX20_R25_LOOKBACK(), &testdata.expectedH48_PMIN6_PMAX20_R25_WIDTH(), &testdata.expectedH48_PMIN6_PMAX20_R25_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(50, 48, 10, 30, &testdata.expectedH48_PMIN10_PMAX30_R50_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX30_R50_TREND_VIGOR(), &testdata.expectedH48_PMIN10_PMAX30_R50_NORMALIZED_VIGOR(), &testdata.expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX30_R50_BAND_PASS(), &testdata.expectedH48_PMIN10_PMAX30_R50_QUADRATURE(), &testdata.expectedH48_PMIN10_PMAX30_R50_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN10_PMAX30_R50_TREND_DISPLACEMENT(), &testdata.expectedH48_PMIN10_PMAX30_R50_RAW_RATIO(), &testdata.expectedH48_PMIN10_PMAX30_R50_LOOKBACK(), &testdata.expectedH48_PMIN10_PMAX30_R50_WIDTH(), &testdata.expectedH48_PMIN10_PMAX30_R50_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(25, 48, 10, 30, &testdata.expectedH48_PMIN10_PMAX30_R25_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX30_R25_TREND_VIGOR(), &testdata.expectedH48_PMIN10_PMAX30_R25_NORMALIZED_VIGOR(), &testdata.expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX30_R25_BAND_PASS(), &testdata.expectedH48_PMIN10_PMAX30_R25_QUADRATURE(), &testdata.expectedH48_PMIN10_PMAX30_R25_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN10_PMAX30_R25_TREND_DISPLACEMENT(), &testdata.expectedH48_PMIN10_PMAX30_R25_RAW_RATIO(), &testdata.expectedH48_PMIN10_PMAX30_R25_LOOKBACK(), &testdata.expectedH48_PMIN10_PMAX30_R25_WIDTH(), &testdata.expectedH48_PMIN10_PMAX30_R25_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(50, 48, 10, 48, &testdata.expectedH48_PMIN10_PMAX48_R50_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX48_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX48_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX48_R50_TREND_VIGOR(), &testdata.expectedH48_PMIN10_PMAX48_R50_NORMALIZED_VIGOR(), &testdata.expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX48_R50_BAND_PASS(), &testdata.expectedH48_PMIN10_PMAX48_R50_QUADRATURE(), &testdata.expectedH48_PMIN10_PMAX48_R50_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN10_PMAX48_R50_TREND_DISPLACEMENT(), &testdata.expectedH48_PMIN10_PMAX48_R50_RAW_RATIO(), &testdata.expectedH48_PMIN10_PMAX48_R50_LOOKBACK(), &testdata.expectedH48_PMIN10_PMAX48_R50_WIDTH(), &testdata.expectedH48_PMIN10_PMAX48_R50_HAS_CYCLE_ESTIMATE());
    try verifyTrendCase(25, 48, 10, 48, &testdata.expectedH48_PMIN10_PMAX48_R25_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX48_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX48_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX48_R25_TREND_VIGOR(), &testdata.expectedH48_PMIN10_PMAX48_R25_NORMALIZED_VIGOR(), &testdata.expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX48_R25_BAND_PASS(), &testdata.expectedH48_PMIN10_PMAX48_R25_QUADRATURE(), &testdata.expectedH48_PMIN10_PMAX48_R25_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN10_PMAX48_R25_TREND_DISPLACEMENT(), &testdata.expectedH48_PMIN10_PMAX48_R25_RAW_RATIO(), &testdata.expectedH48_PMIN10_PMAX48_R25_LOOKBACK(), &testdata.expectedH48_PMIN10_PMAX48_R25_WIDTH(), &testdata.expectedH48_PMIN10_PMAX48_R25_HAS_CYCLE_ESTIMATE());
}

test "CoronaTrendVigor metadata, mnemonics, sinusoid cancellation, validation, and entities" {
    var instance = try CoronaTrendVigor.init(testing.allocator, defaultParams());
    defer instance.deinit();

    var md: Metadata = undefined;
    instance.getMetadata(&md);
    try testing.expectEqual(Identifier.corona_trend_vigor, md.identifier);
    try testing.expectEqualStrings("ctv(50, 20, -10, 10, 30)", md.mnemonic);
    try testing.expectEqual(@as(usize, 2), md.outputs_len);
    try testing.expectEqual(shape_mod.Shape.heatmap, md.outputs_buf[0].shape);
    try testing.expectEqual(shape_mod.Shape.scalar, md.outputs_buf[1].shape);
    try testing.expectEqualStrings("ctv-tv(30)", md.outputs_buf[1].mnemonic);

    var flagged = try CoronaTrendVigor.init(testing.allocator, .{
        .minimal_period = 10,
        .maximal_period = 48,
        .lookback_offset = -1,
        .bar_component = .median,
    });
    defer flagged.deinit();
    var flagged_md: Metadata = undefined;
    flagged.getMetadata(&flagged_md);
    try testing.expectEqualStrings("ctv(50, 20, -10, 10, 30, pmin=10, pmax=48, offset=-1, hl/2)", flagged_md.mnemonic);
    try testing.expectEqualStrings("ctv-tv(30, pmin=10, pmax=48, offset=-1, hl/2)", flagged_md.outputs_buf[1].mnemonic);

    var engine = try Corona.init(testing.allocator, null);
    defer engine.deinit();
    var consumer = try CoronaTrendVigor.init(testing.allocator, defaultParams());
    defer consumer.deinit();
    for (0..120) |index| {
        const index_f: f64 = @floatFromInt(index);
        const price = 100.0 + 0.5 * index_f + 3.0 * @sin(corona_mod.two_pi * index_f / 20.0);
        var snap = engine.update(price, @intCast(index));
        snap.dominant_cycle = 20.0;
        snap.dominant_cycle_median = 20.0;
        const output = consumer.updateFromSnapshot(snap);
        if (index < 20) {
            try testing.expect(output.heatmap.isEmpty());
        } else {
            try testing.expectEqual(@as(f64, 20.0), output.lookback);
            try testing.expect(@abs(output.trend_displacement - 10.0) <= 1e-12);
        }
    }
    try testing.expect(consumer.isPrimed());

    try testing.expectError(error.InvalidRasterLength, CoronaTrendVigor.init(testing.allocator, .{ .raster_length = 1 }));
    try testing.expectError(error.InvalidMaxRasterValue, CoronaTrendVigor.init(testing.allocator, .{ .max_raster_value = 6.0 }));
    try testing.expectError(error.InvalidMaxParameterValue, CoronaTrendVigor.init(testing.allocator, .{ .min_parameter_value = 10.0, .max_parameter_value = 10.0 }));
    try testing.expectError(error.InvalidHighPassFilterCutoff, CoronaTrendVigor.init(testing.allocator, .{ .high_pass_filter_cutoff = 4 }));
    try testing.expectError(error.InvalidMinimalPeriod, CoronaTrendVigor.init(testing.allocator, .{ .minimal_period = 5 }));
    try testing.expectError(error.InvalidMaximalPeriod, CoronaTrendVigor.init(testing.allocator, .{ .minimal_period = 10, .maximal_period = 10 }));
    try testing.expectError(error.InvalidLookbackOffset, CoronaTrendVigor.init(testing.allocator, .{ .lookback_offset = 1 }));
    try testing.expectError(error.InvalidLookbackOffset, CoronaTrendVigor.init(testing.allocator, .{ .lookback_offset = -2 }));

    const prices = testdata.testInput();
    const opens = testdata.testOpen();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();
    var bar_ind = try CoronaTrendVigor.init(testing.allocator, defaultParams());
    defer bar_ind.deinit();
    var scalar_ind = try CoronaTrendVigor.init(testing.allocator, defaultParams());
    defer scalar_ind.deinit();
    var quote_ind = try CoronaTrendVigor.init(testing.allocator, defaultParams());
    defer quote_ind.deinit();
    var trade_ind = try CoronaTrendVigor.init(testing.allocator, defaultParams());
    defer trade_ind.deinit();
    for (0..35) |index| {
        const bar = Bar{ .time = @intCast(index), .open = opens[index], .high = highs[index], .low = lows[index], .close = prices[index], .volume = 1000.0 };
        const scalar = Scalar{ .time = @intCast(index), .value = prices[index] };
        const quote = Quote{ .time = @intCast(index), .bid_price = prices[index] - 0.5, .ask_price = prices[index] + 0.5, .bid_size = 10.0, .ask_size = 10.0 };
        const trade = Trade{ .time = @intCast(index), .price = prices[index], .volume = 10.0 };
        try testing.expectEqual(@as(usize, 2), bar_ind.updateBar(&bar).len);
        try testing.expectEqual(@as(usize, 2), scalar_ind.updateScalar(&scalar).len);
        try testing.expectEqual(@as(usize, 2), quote_ind.updateQuote(&quote).len);
        try testing.expectEqual(@as(usize, 2), trade_ind.updateTrade(&trade).len);
    }
}
