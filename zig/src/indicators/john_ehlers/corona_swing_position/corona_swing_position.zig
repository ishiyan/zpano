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

pub const CoronaSwingPositionOutput = enum(u8) {
    value = 1,
    swing_position = 2,
};

pub const Params = struct {
    raster_length: i32 = 50,
    max_raster_value: f64 = 20.0,
    min_parameter_value: f64 = -5.0,
    max_parameter_value: f64 = 5.0,
    high_pass_filter_cutoff: i32 = 30,
    minimal_period: i32 = 6,
    maximal_period: i32 = 30,
    lead_window: i32 = 51,
    position_window: i32 = 21,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub fn defaultParams() Params {
    return .{};
}

pub const Error = corona_mod.Error;

pub const CoronaSwingPositionSnapshot = struct {
    heatmap: Heatmap,
    swing_position: f64,
    normalized_position: f64,
    dominant_cycle: f64,
    dominant_cycle_median: f64,
    band_pass: f64,
    quadrature: f64,
    cycle_amplitude: f64,
    lead60: f64,
    position_span: f64,
    width: f64,
    has_cycle_estimate: f64,
};

pub const Snapshot = CoronaSwingPositionSnapshot;

fn formatSwingFlags(buffer: []u8, params: Params) Error![]const u8 {
    var pos: usize = 0;
    if (params.minimal_period != 6) {
        const chunk = std.fmt.bufPrint(buffer[pos..], ", pmin={d}", .{params.minimal_period}) catch return error.MnemonicTooLong;
        pos += chunk.len;
    }
    if (params.maximal_period != 30) {
        const chunk = std.fmt.bufPrint(buffer[pos..], ", pmax={d}", .{params.maximal_period}) catch return error.MnemonicTooLong;
        pos += chunk.len;
    }
    if (params.lead_window != 51) {
        const chunk = std.fmt.bufPrint(buffer[pos..], ", lead={d}", .{params.lead_window}) catch return error.MnemonicTooLong;
        pos += chunk.len;
    }
    if (params.position_window != 21) {
        const chunk = std.fmt.bufPrint(buffer[pos..], ", pos={d}", .{params.position_window}) catch return error.MnemonicTooLong;
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
    const flags = try formatSwingFlags(&flags_buffer, params);

    return std.fmt.bufPrint(buffer, "cswing({d}, {d}, {d}, {d}, {d}{s}{s})", .{
        params.raster_length,
        corona_mod.normalizeZero(params.max_raster_value),
        corona_mod.normalizeZero(params.min_parameter_value),
        corona_mod.normalizeZero(params.max_parameter_value),
        params.high_pass_filter_cutoff,
        flags,
        triple,
    }) catch return error.MnemonicTooLong;
}

pub const CoronaSwingPosition = struct {
    allocator: std.mem.Allocator,
    engine: Corona,
    engine_params: corona_mod.Params,
    bandpass: AdaptiveBandPass,
    raster: Raster,
    sample_count: usize,
    leads: []f64,
    leads_len: usize,
    positions: []f64,
    positions_len: usize,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_storage: [256]u8,
    mnemonic_len: usize,
    description_storage: [320]u8,
    description_len: usize,
    mnemonic_sp_storage: [160]u8,
    mnemonic_sp_len: usize,
    description_sp_storage: [256]u8,
    description_sp_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: Params) Error!CoronaSwingPosition {
        if (params.lead_window < 2) return error.InvalidLeadWindow;
        if (params.position_window < 2) return error.InvalidPositionWindow;
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

        const lead_cap: usize = @intCast(params.lead_window);
        const pos_cap: usize = @intCast(params.position_window);
        const leads = allocator.alloc(f64, lead_cap) catch return error.OutOfMemory;
        errdefer allocator.free(leads);
        @memset(leads, 0.0);
        const positions = allocator.alloc(f64, pos_cap) catch return error.OutOfMemory;
        errdefer allocator.free(positions);
        @memset(positions, 0.0);

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buffer: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buffer, bc, qc, tc);
        var flags_buffer: [96]u8 = undefined;
        const flags = try formatSwingFlags(&flags_buffer, params);

        var mnemonic_storage: [256]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_storage, "cswing({d}, {d}, {d}, {d}, {d}{s}{s})", .{
            raster.length,
            corona_mod.normalizeZero(raster.ceiling),
            corona_mod.normalizeZero(raster.lower),
            corona_mod.normalizeZero(raster.upper),
            params.high_pass_filter_cutoff,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;

        var description_storage: [320]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Corona swing position {s}", .{mn}) catch
            return error.MnemonicTooLong;

        var mnemonic_sp_storage: [160]u8 = undefined;
        const mn_sp = std.fmt.bufPrint(&mnemonic_sp_storage, "cswing-sp({d}{s}{s})", .{
            params.high_pass_filter_cutoff,
            flags,
            triple,
        }) catch return error.MnemonicTooLong;

        var description_sp_storage: [256]u8 = undefined;
        const desc_sp = std.fmt.bufPrint(&description_sp_storage, "Corona swing position scalar {s}", .{mn_sp}) catch
            return error.MnemonicTooLong;

        return .{
            .allocator = allocator,
            .engine = engine,
            .engine_params = engine.parameters,
            .bandpass = .{},
            .raster = raster,
            .sample_count = 0,
            .leads = leads,
            .leads_len = leads.len,
            .positions = positions,
            .positions_len = positions.len,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mn.len,
            .description_storage = description_storage,
            .description_len = desc.len,
            .mnemonic_sp_storage = mnemonic_sp_storage,
            .mnemonic_sp_len = mn_sp.len,
            .description_sp_storage = description_sp_storage,
            .description_sp_len = desc_sp.len,
        };
    }

    pub fn deinit(self: *CoronaSwingPosition) void {
        self.engine.deinit();
        self.raster.deinit();
        self.allocator.free(self.leads);
        self.allocator.free(self.positions);
    }

    pub fn reset(self: *CoronaSwingPosition) void {
        self.engine.reset();
        self.bandpass.reset();
        self.raster.reset();
        self.sample_count = 0;
        @memset(self.leads, 0.0);
        self.leads_len = self.leads.len;
        @memset(self.positions, 0.0);
        self.positions_len = self.positions.len;
    }

    pub fn fixSlices(self: *CoronaSwingPosition) void {
        _ = self;
    }

    pub fn mnemonic(self: *const CoronaSwingPosition) []const u8 {
        return self.mnemonic_storage[0..self.mnemonic_len];
    }

    fn description(self: *const CoronaSwingPosition) []const u8 {
        return self.description_storage[0..self.description_len];
    }

    fn mnemonicSP(self: *const CoronaSwingPosition) []const u8 {
        return self.mnemonic_sp_storage[0..self.mnemonic_sp_len];
    }

    fn descriptionSP(self: *const CoronaSwingPosition) []const u8 {
        return self.description_sp_storage[0..self.description_sp_len];
    }

    fn emptySnapshot(self: *const CoronaSwingPosition, time: i64) CoronaSwingPositionSnapshot {
        const nan = math.nan(f64);
        return .{
            .heatmap = self.raster.empty(time),
            .swing_position = nan,
            .normalized_position = nan,
            .dominant_cycle = nan,
            .dominant_cycle_median = nan,
            .band_pass = nan,
            .quadrature = nan,
            .cycle_amplitude = nan,
            .lead60 = nan,
            .position_span = nan,
            .width = nan,
            .has_cycle_estimate = nan,
        };
    }

    pub fn tryUpdateFromSnapshot(self: *CoronaSwingPosition, snapshot: CoronaSnapshot) Error!CoronaSwingPositionSnapshot {
        const accepted = try corona_mod.acceptSnapshot(snapshot, self.engine_params, self.sample_count + 1);
        if (!accepted) {
            return self.emptySnapshot(snapshot.time);
        }
        const cycle = try self.bandpass.update(snapshot.price, snapshot.dominant_cycle_median);
        self.sample_count = snapshot.sample_count;

        const lead_bounds = corona_mod.rollingAppend(self.leads, &self.leads_len, cycle.lead60);
        const flat = lead_bounds.maximum == lead_bounds.minimum;
        const position = corona_mod.clip(
            if (flat) 0.5 else (cycle.lead60 - lead_bounds.minimum) / (lead_bounds.maximum - lead_bounds.minimum),
            0.0,
            1.0,
        );
        const pos_bounds = corona_mod.rollingAppend(self.positions, &self.positions_len, position);
        const span = pos_bounds.maximum - pos_bounds.minimum;
        var width = if (span > 0.85) 0.01 else 0.15 * span;
        if (width == 0.0) {
            width = 0.01;
        }

        self.raster.update(position, width, .swing, span > 0.8 or flat);
        if (!snapshot.primed) {
            return self.emptySnapshot(snapshot.time);
        }

        const mapped = self.raster.lower + (self.raster.upper - self.raster.lower) * position;
        return .{
            .heatmap = self.raster.column(snapshot.time),
            .swing_position = mapped,
            .normalized_position = position,
            .dominant_cycle = snapshot.dominant_cycle,
            .dominant_cycle_median = snapshot.dominant_cycle_median,
            .band_pass = cycle.band_pass,
            .quadrature = cycle.quadrature,
            .cycle_amplitude = cycle.amplitude,
            .lead60 = cycle.lead60,
            .position_span = span,
            .width = width,
            .has_cycle_estimate = if (snapshot.has_cycle_estimate) 1.0 else 0.0,
        };
    }

    pub fn updateFromSnapshot(self: *CoronaSwingPosition, snapshot: CoronaSnapshot) CoronaSwingPositionSnapshot {
        return self.tryUpdateFromSnapshot(snapshot) catch |err| {
            std.debug.panic("CoronaSwingPosition.updateFromSnapshot failed: {s}", .{@errorName(err)});
        };
    }

    pub fn tryUpdate(self: *CoronaSwingPosition, sample: f64, time: i64) Error!CoronaSwingPositionSnapshot {
        const engine_snap = try self.engine.tryUpdate(sample, time);
        return self.tryUpdateFromSnapshot(engine_snap);
    }

    pub fn update(self: *CoronaSwingPosition, sample: f64, time: i64) CoronaSwingPositionSnapshot {
        return self.tryUpdate(sample, time) catch |err| {
            std.debug.panic("CoronaSwingPosition.update failed: {s}", .{@errorName(err)});
        };
    }

    pub const updateSample = update;

    fn wrapEntity(time: i64, snapshot: CoronaSwingPositionSnapshot) OutputArray {
        var out = OutputArray{};
        out.append(.{ .heatmap = snapshot.heatmap });
        out.append(.{ .scalar = .{ .time = time, .value = snapshot.swing_position } });
        return out;
    }

    pub fn updateScalar(self: *CoronaSwingPosition, sample: *const Scalar) OutputArray {
        return wrapEntity(sample.time, self.update(sample.value, sample.time));
    }

    pub fn updateBar(self: *CoronaSwingPosition, sample: *const Bar) OutputArray {
        return wrapEntity(sample.time, self.update(self.bar_func(sample.*), sample.time));
    }

    pub fn updateQuote(self: *CoronaSwingPosition, sample: *const Quote) OutputArray {
        return wrapEntity(sample.time, self.update(self.quote_func(sample.*), sample.time));
    }

    pub fn updateTrade(self: *CoronaSwingPosition, sample: *const Trade) OutputArray {
        return wrapEntity(sample.time, self.update(self.trade_func(sample.*), sample.time));
    }

    pub fn isPrimed(self: *const CoronaSwingPosition) bool {
        return self.sample_count >= corona_mod.first_bank_sample;
    }

    pub fn getMetadata(self: *const CoronaSwingPosition, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
            .{ .mnemonic = self.mnemonicSP(), .description = self.descriptionSP() },
        };
        build_metadata_mod.buildMetadata(out, .corona_swing_position, mn, desc, &texts);
    }

    pub fn indicator(self: *CoronaSwingPosition) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(CoronaSwingPosition);
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

fn verifySwingCase(
    comptime raster_len: usize,
    hp_cutoff: i32,
    min_period: i32,
    max_period: i32,
    expected_heatmap: *const [252][raster_len]f64,
    expected_min: *const [252]f64,
    expected_max: *const [252]f64,
    expected_sp: *const [252]f64,
    expected_norm: *const [252]f64,
    expected_dc: *const [252]f64,
    expected_dcm: *const [252]f64,
    expected_bp: *const [252]f64,
    expected_quad: *const [252]f64,
    expected_amp: *const [252]f64,
    expected_lead: *const [252]f64,
    expected_span: *const [252]f64,
    expected_width: *const [252]f64,
    expected_has_est: *const [252]f64,
) !void {
    var instance = try CoronaSwingPosition.init(testing.allocator, .{
        .raster_length = @intCast(raster_len),
        .high_pass_filter_cutoff = hp_cutoff,
        .minimal_period = min_period,
        .maximal_period = max_period,
    });
    defer instance.deinit();

    const step = 10.0 / @as(f64, @floatFromInt(raster_len));
    const expected_first = -5.0 + step;
    const expected_resolution = 1.0 / step;
    const prices = testdata.testInput();

    for (prices, 0..) |price, index| {
        const snap = instance.update(price, @intCast(index));
        try expectClose(snap.heatmap.parameter_first, expected_first);
        try expectClose(snap.heatmap.parameter_last, 5.0);
        try expectClose(snap.heatmap.parameter_resolution, expected_resolution);

        const row_empty = math.isNan(expected_sp[index]);
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
        try expectClose(snap.swing_position, expected_sp[index]);
        try expectClose(snap.normalized_position, expected_norm[index]);
        try expectClose(snap.dominant_cycle, expected_dc[index]);
        try expectClose(snap.dominant_cycle_median, expected_dcm[index]);
        try expectClose(snap.band_pass, expected_bp[index]);
        try expectClose(snap.quadrature, expected_quad[index]);
        try expectClose(snap.cycle_amplitude, expected_amp[index]);
        try expectClose(snap.lead60, expected_lead[index]);
        try expectClose(snap.position_span, expected_span[index]);
        try expectClose(snap.width, expected_width[index]);
        try expectClose(snap.has_cycle_estimate, expected_has_est[index]);
    }

    instance.reset();
    try testing.expect(!instance.isPrimed());
}

test "CoronaSwingPosition matches all 16 reference cases" {
    try verifySwingCase(50, 30, 6, 30, &testdata.expectedH30_PMIN6_PMAX30_R50_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX30_R50_SWING_POSITION(), &testdata.expectedH30_PMIN6_PMAX30_R50_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX30_R50_BAND_PASS(), &testdata.expectedH30_PMIN6_PMAX30_R50_QUADRATURE(), &testdata.expectedH30_PMIN6_PMAX30_R50_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN6_PMAX30_R50_LEAD60(), &testdata.expectedH30_PMIN6_PMAX30_R50_POSITION_SPAN(), &testdata.expectedH30_PMIN6_PMAX30_R50_WIDTH(), &testdata.expectedH30_PMIN6_PMAX30_R50_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(25, 30, 6, 30, &testdata.expectedH30_PMIN6_PMAX30_R25_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX30_R25_SWING_POSITION(), &testdata.expectedH30_PMIN6_PMAX30_R25_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX30_R25_BAND_PASS(), &testdata.expectedH30_PMIN6_PMAX30_R25_QUADRATURE(), &testdata.expectedH30_PMIN6_PMAX30_R25_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN6_PMAX30_R25_LEAD60(), &testdata.expectedH30_PMIN6_PMAX30_R25_POSITION_SPAN(), &testdata.expectedH30_PMIN6_PMAX30_R25_WIDTH(), &testdata.expectedH30_PMIN6_PMAX30_R25_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(50, 30, 6, 20, &testdata.expectedH30_PMIN6_PMAX20_R50_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX20_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX20_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX20_R50_SWING_POSITION(), &testdata.expectedH30_PMIN6_PMAX20_R50_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX20_R50_BAND_PASS(), &testdata.expectedH30_PMIN6_PMAX20_R50_QUADRATURE(), &testdata.expectedH30_PMIN6_PMAX20_R50_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN6_PMAX20_R50_LEAD60(), &testdata.expectedH30_PMIN6_PMAX20_R50_POSITION_SPAN(), &testdata.expectedH30_PMIN6_PMAX20_R50_WIDTH(), &testdata.expectedH30_PMIN6_PMAX20_R50_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(25, 30, 6, 20, &testdata.expectedH30_PMIN6_PMAX20_R25_HEATMAP(), &testdata.expectedH30_PMIN6_PMAX20_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN6_PMAX20_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN6_PMAX20_R25_SWING_POSITION(), &testdata.expectedH30_PMIN6_PMAX20_R25_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN6_PMAX20_R25_BAND_PASS(), &testdata.expectedH30_PMIN6_PMAX20_R25_QUADRATURE(), &testdata.expectedH30_PMIN6_PMAX20_R25_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN6_PMAX20_R25_LEAD60(), &testdata.expectedH30_PMIN6_PMAX20_R25_POSITION_SPAN(), &testdata.expectedH30_PMIN6_PMAX20_R25_WIDTH(), &testdata.expectedH30_PMIN6_PMAX20_R25_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(50, 30, 10, 30, &testdata.expectedH30_PMIN10_PMAX30_R50_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX30_R50_SWING_POSITION(), &testdata.expectedH30_PMIN10_PMAX30_R50_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX30_R50_BAND_PASS(), &testdata.expectedH30_PMIN10_PMAX30_R50_QUADRATURE(), &testdata.expectedH30_PMIN10_PMAX30_R50_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN10_PMAX30_R50_LEAD60(), &testdata.expectedH30_PMIN10_PMAX30_R50_POSITION_SPAN(), &testdata.expectedH30_PMIN10_PMAX30_R50_WIDTH(), &testdata.expectedH30_PMIN10_PMAX30_R50_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(25, 30, 10, 30, &testdata.expectedH30_PMIN10_PMAX30_R25_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX30_R25_SWING_POSITION(), &testdata.expectedH30_PMIN10_PMAX30_R25_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX30_R25_BAND_PASS(), &testdata.expectedH30_PMIN10_PMAX30_R25_QUADRATURE(), &testdata.expectedH30_PMIN10_PMAX30_R25_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN10_PMAX30_R25_LEAD60(), &testdata.expectedH30_PMIN10_PMAX30_R25_POSITION_SPAN(), &testdata.expectedH30_PMIN10_PMAX30_R25_WIDTH(), &testdata.expectedH30_PMIN10_PMAX30_R25_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(50, 30, 10, 48, &testdata.expectedH30_PMIN10_PMAX48_R50_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX48_R50_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX48_R50_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX48_R50_SWING_POSITION(), &testdata.expectedH30_PMIN10_PMAX48_R50_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX48_R50_BAND_PASS(), &testdata.expectedH30_PMIN10_PMAX48_R50_QUADRATURE(), &testdata.expectedH30_PMIN10_PMAX48_R50_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN10_PMAX48_R50_LEAD60(), &testdata.expectedH30_PMIN10_PMAX48_R50_POSITION_SPAN(), &testdata.expectedH30_PMIN10_PMAX48_R50_WIDTH(), &testdata.expectedH30_PMIN10_PMAX48_R50_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(25, 30, 10, 48, &testdata.expectedH30_PMIN10_PMAX48_R25_HEATMAP(), &testdata.expectedH30_PMIN10_PMAX48_R25_HEATMAP_MIN(), &testdata.expectedH30_PMIN10_PMAX48_R25_HEATMAP_MAX(), &testdata.expectedH30_PMIN10_PMAX48_R25_SWING_POSITION(), &testdata.expectedH30_PMIN10_PMAX48_R25_NORMALIZED_POSITION(), &testdata.expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE(), &testdata.expectedH30_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH30_PMIN10_PMAX48_R25_BAND_PASS(), &testdata.expectedH30_PMIN10_PMAX48_R25_QUADRATURE(), &testdata.expectedH30_PMIN10_PMAX48_R25_CYCLE_AMPLITUDE(), &testdata.expectedH30_PMIN10_PMAX48_R25_LEAD60(), &testdata.expectedH30_PMIN10_PMAX48_R25_POSITION_SPAN(), &testdata.expectedH30_PMIN10_PMAX48_R25_WIDTH(), &testdata.expectedH30_PMIN10_PMAX48_R25_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(50, 48, 6, 30, &testdata.expectedH48_PMIN6_PMAX30_R50_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX30_R50_SWING_POSITION(), &testdata.expectedH48_PMIN6_PMAX30_R50_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX30_R50_BAND_PASS(), &testdata.expectedH48_PMIN6_PMAX30_R50_QUADRATURE(), &testdata.expectedH48_PMIN6_PMAX30_R50_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN6_PMAX30_R50_LEAD60(), &testdata.expectedH48_PMIN6_PMAX30_R50_POSITION_SPAN(), &testdata.expectedH48_PMIN6_PMAX30_R50_WIDTH(), &testdata.expectedH48_PMIN6_PMAX30_R50_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(25, 48, 6, 30, &testdata.expectedH48_PMIN6_PMAX30_R25_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX30_R25_SWING_POSITION(), &testdata.expectedH48_PMIN6_PMAX30_R25_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX30_R25_BAND_PASS(), &testdata.expectedH48_PMIN6_PMAX30_R25_QUADRATURE(), &testdata.expectedH48_PMIN6_PMAX30_R25_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN6_PMAX30_R25_LEAD60(), &testdata.expectedH48_PMIN6_PMAX30_R25_POSITION_SPAN(), &testdata.expectedH48_PMIN6_PMAX30_R25_WIDTH(), &testdata.expectedH48_PMIN6_PMAX30_R25_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(50, 48, 6, 20, &testdata.expectedH48_PMIN6_PMAX20_R50_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX20_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX20_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX20_R50_SWING_POSITION(), &testdata.expectedH48_PMIN6_PMAX20_R50_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX20_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX20_R50_BAND_PASS(), &testdata.expectedH48_PMIN6_PMAX20_R50_QUADRATURE(), &testdata.expectedH48_PMIN6_PMAX20_R50_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN6_PMAX20_R50_LEAD60(), &testdata.expectedH48_PMIN6_PMAX20_R50_POSITION_SPAN(), &testdata.expectedH48_PMIN6_PMAX20_R50_WIDTH(), &testdata.expectedH48_PMIN6_PMAX20_R50_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(25, 48, 6, 20, &testdata.expectedH48_PMIN6_PMAX20_R25_HEATMAP(), &testdata.expectedH48_PMIN6_PMAX20_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN6_PMAX20_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN6_PMAX20_R25_SWING_POSITION(), &testdata.expectedH48_PMIN6_PMAX20_R25_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN6_PMAX20_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN6_PMAX20_R25_BAND_PASS(), &testdata.expectedH48_PMIN6_PMAX20_R25_QUADRATURE(), &testdata.expectedH48_PMIN6_PMAX20_R25_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN6_PMAX20_R25_LEAD60(), &testdata.expectedH48_PMIN6_PMAX20_R25_POSITION_SPAN(), &testdata.expectedH48_PMIN6_PMAX20_R25_WIDTH(), &testdata.expectedH48_PMIN6_PMAX20_R25_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(50, 48, 10, 30, &testdata.expectedH48_PMIN10_PMAX30_R50_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX30_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX30_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX30_R50_SWING_POSITION(), &testdata.expectedH48_PMIN10_PMAX30_R50_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX30_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX30_R50_BAND_PASS(), &testdata.expectedH48_PMIN10_PMAX30_R50_QUADRATURE(), &testdata.expectedH48_PMIN10_PMAX30_R50_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN10_PMAX30_R50_LEAD60(), &testdata.expectedH48_PMIN10_PMAX30_R50_POSITION_SPAN(), &testdata.expectedH48_PMIN10_PMAX30_R50_WIDTH(), &testdata.expectedH48_PMIN10_PMAX30_R50_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(25, 48, 10, 30, &testdata.expectedH48_PMIN10_PMAX30_R25_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX30_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX30_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX30_R25_SWING_POSITION(), &testdata.expectedH48_PMIN10_PMAX30_R25_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX30_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX30_R25_BAND_PASS(), &testdata.expectedH48_PMIN10_PMAX30_R25_QUADRATURE(), &testdata.expectedH48_PMIN10_PMAX30_R25_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN10_PMAX30_R25_LEAD60(), &testdata.expectedH48_PMIN10_PMAX30_R25_POSITION_SPAN(), &testdata.expectedH48_PMIN10_PMAX30_R25_WIDTH(), &testdata.expectedH48_PMIN10_PMAX30_R25_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(50, 48, 10, 48, &testdata.expectedH48_PMIN10_PMAX48_R50_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX48_R50_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX48_R50_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX48_R50_SWING_POSITION(), &testdata.expectedH48_PMIN10_PMAX48_R50_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX48_R50_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX48_R50_BAND_PASS(), &testdata.expectedH48_PMIN10_PMAX48_R50_QUADRATURE(), &testdata.expectedH48_PMIN10_PMAX48_R50_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN10_PMAX48_R50_LEAD60(), &testdata.expectedH48_PMIN10_PMAX48_R50_POSITION_SPAN(), &testdata.expectedH48_PMIN10_PMAX48_R50_WIDTH(), &testdata.expectedH48_PMIN10_PMAX48_R50_HAS_CYCLE_ESTIMATE());
    try verifySwingCase(25, 48, 10, 48, &testdata.expectedH48_PMIN10_PMAX48_R25_HEATMAP(), &testdata.expectedH48_PMIN10_PMAX48_R25_HEATMAP_MIN(), &testdata.expectedH48_PMIN10_PMAX48_R25_HEATMAP_MAX(), &testdata.expectedH48_PMIN10_PMAX48_R25_SWING_POSITION(), &testdata.expectedH48_PMIN10_PMAX48_R25_NORMALIZED_POSITION(), &testdata.expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE(), &testdata.expectedH48_PMIN10_PMAX48_R25_DOMINANT_CYCLE_MEDIAN(), &testdata.expectedH48_PMIN10_PMAX48_R25_BAND_PASS(), &testdata.expectedH48_PMIN10_PMAX48_R25_QUADRATURE(), &testdata.expectedH48_PMIN10_PMAX48_R25_CYCLE_AMPLITUDE(), &testdata.expectedH48_PMIN10_PMAX48_R25_LEAD60(), &testdata.expectedH48_PMIN10_PMAX48_R25_POSITION_SPAN(), &testdata.expectedH48_PMIN10_PMAX48_R25_WIDTH(), &testdata.expectedH48_PMIN10_PMAX48_R25_HAS_CYCLE_ESTIMATE());
}

test "CoronaSwingPosition metadata, mnemonics, validation, and entities" {
    var instance = try CoronaSwingPosition.init(testing.allocator, defaultParams());
    defer instance.deinit();

    var md: Metadata = undefined;
    instance.getMetadata(&md);
    try testing.expectEqual(Identifier.corona_swing_position, md.identifier);
    try testing.expectEqualStrings("cswing(50, 20, -5, 5, 30)", md.mnemonic);
    try testing.expectEqual(@as(usize, 2), md.outputs_len);
    try testing.expectEqual(shape_mod.Shape.heatmap, md.outputs_buf[0].shape);
    try testing.expectEqual(shape_mod.Shape.scalar, md.outputs_buf[1].shape);
    try testing.expectEqualStrings("cswing-sp(30)", md.outputs_buf[1].mnemonic);

    var flagged = try CoronaSwingPosition.init(testing.allocator, .{
        .minimal_period = 10,
        .maximal_period = 48,
        .lead_window = 25,
        .position_window = 15,
        .bar_component = .median,
    });
    defer flagged.deinit();
    var flagged_md: Metadata = undefined;
    flagged.getMetadata(&flagged_md);
    try testing.expectEqualStrings("cswing(50, 20, -5, 5, 30, pmin=10, pmax=48, lead=25, pos=15, hl/2)", flagged_md.mnemonic);
    try testing.expectEqualStrings("cswing-sp(30, pmin=10, pmax=48, lead=25, pos=15, hl/2)", flagged_md.outputs_buf[1].mnemonic);

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
        var inst = try CoronaSwingPosition.init(testing.allocator, p);
        defer inst.deinit();
        var mn_buffer: [160]u8 = undefined;
        const formatted = try formatMnemonic(&mn_buffer, p);
        try testing.expectEqualStrings(formatted, inst.mnemonic());
    }

    try testing.expectError(error.InvalidRasterLength, CoronaSwingPosition.init(testing.allocator, .{ .raster_length = 1 }));
    try testing.expectError(error.InvalidMaxRasterValue, CoronaSwingPosition.init(testing.allocator, .{ .max_raster_value = 6.0 }));
    try testing.expectError(error.InvalidMaxParameterValue, CoronaSwingPosition.init(testing.allocator, .{ .min_parameter_value = 5.0, .max_parameter_value = 5.0 }));
    try testing.expectError(error.InvalidHighPassFilterCutoff, CoronaSwingPosition.init(testing.allocator, .{ .high_pass_filter_cutoff = 4 }));
    try testing.expectError(error.InvalidMinimalPeriod, CoronaSwingPosition.init(testing.allocator, .{ .minimal_period = 5 }));
    try testing.expectError(error.InvalidMaximalPeriod, CoronaSwingPosition.init(testing.allocator, .{ .minimal_period = 10, .maximal_period = 10 }));
    try testing.expectError(error.InvalidLeadWindow, CoronaSwingPosition.init(testing.allocator, .{ .lead_window = 1 }));
    try testing.expectError(error.InvalidPositionWindow, CoronaSwingPosition.init(testing.allocator, .{ .position_window = 1 }));

    const prices = testdata.testInput();
    const opens = testdata.testOpen();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();
    var bar_ind = try CoronaSwingPosition.init(testing.allocator, defaultParams());
    defer bar_ind.deinit();
    var scalar_ind = try CoronaSwingPosition.init(testing.allocator, defaultParams());
    defer scalar_ind.deinit();
    var quote_ind = try CoronaSwingPosition.init(testing.allocator, defaultParams());
    defer quote_ind.deinit();
    var trade_ind = try CoronaSwingPosition.init(testing.allocator, defaultParams());
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
}
