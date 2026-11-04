const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const DominantMarketVibrationParams = struct {
    window: isize = 200,
    min_period: isize = 5,
    max_period: isize = 80,
    smoothing: f64 = 1600.0,
    genuine_threshold: f64 = 49.0,
    phase_cycles: f64 = 2.0,
    top_k: isize = 3,
    peak_distance: isize = 2,
    max_segments: isize = 5,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const DominantMarketVibrationOutput = enum(u8) { period = 1, amplitude = 2, phase = 3, strength = 4, genuine_percent = 5, value = 6, slope = 7, eta_top = 8, eta_bottom = 9 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const DominantMarketVibration = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: DominantMarketVibrationParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: DominantMarketVibrationParams) !DominantMarketVibration {
        try validation.integer(allocator, p.window, "window", 6);
        try validation.integer(allocator, p.min_period, "min_period", 3);
        try validation.integer(allocator, p.max_period, "max_period", 3);
        try validation.real(allocator, p.smoothing, "smoothing", 0.0);
        try validation.real(allocator, p.genuine_threshold, "genuine_threshold", 0.0);
        try validation.real(allocator, p.phase_cycles, "phase_cycles", 0.0);
        try validation.integer(allocator, p.top_k, "top_k", 1);
        try validation.integer(allocator, p.peak_distance, "peak_distance", 1);
        try validation.integer(allocator, p.max_segments, "max_segments", 2);
        if (p.min_period > p.max_period) return error.InvalidRange;
        if (p.genuine_threshold > 100 or (p.phase_cycles != 0 and p.phase_cycles < 1)) return error.InvalidPhaseConfig;
        const bc = p.bar_component orelse entities.bar_component.default_bar_component;
        const qc = p.quote_component orelse entities.quote_component.default_quote_component;
        const tc = p.trade_component orelse entities.trade_component.default_trade_component;
        var suffix_buf: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&suffix_buf, bc, qc, tc);
        var arena = std.heap.ArenaAllocator.init(allocator);
        errdefer arena.deinit();
        const a = arena.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, p) });
        try engine.reset(a);
        var self = DominantMarketVibration{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "dmv({d},{d},{d},{d},{d},{d},{d},{d},{d}{s})", .{ p.window, p.min_period, p.max_period, p.smoothing, p.genuine_threshold, p.phase_cycles, p.top_k, p.peak_distance, p.max_segments, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Dominant market vibration {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *DominantMarketVibration) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *DominantMarketVibration) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const DominantMarketVibration) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const DominantMarketVibration) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *DominantMarketVibration, price: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const DominantMarketVibration, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [9][320]u8 = undefined;
        var description_text: [9][448]u8 = undefined;
        var texts: [9]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} period", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} period", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} amplitude", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} amplitude", .{description}) catch unreachable };
        texts[2] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[2], "{s} phase", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[2], "{s} phase", .{description}) catch unreachable };
        texts[3] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[3], "{s} strength", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[3], "{s} strength", .{description}) catch unreachable };
        texts[4] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[4], "{s} genuine_percent", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[4], "{s} genuine_percent", .{description}) catch unreachable };
        texts[5] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[5], "{s} value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[5], "{s} value", .{description}) catch unreachable };
        texts[6] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[6], "{s} slope", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[6], "{s} slope", .{description}) catch unreachable };
        texts[7] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[7], "{s} eta_top", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[7], "{s} eta_top", .{description}) catch unreachable };
        texts[8] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[8], "{s} eta_bottom", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[8], "{s} eta_bottom", .{description}) catch unreachable };
        indicators.buildMetadata(out, .dominant_market_vibration, mnemonic, description, &texts);
    }
    fn wrap(self: *const DominantMarketVibration, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.period } });
        output.append(.{ .scalar = .{ .time = time, .value = result.amplitude } });
        output.append(.{ .scalar = .{ .time = time, .value = result.phase } });
        output.append(.{ .scalar = .{ .time = time, .value = result.strength } });
        output.append(.{ .scalar = .{ .time = time, .value = result.genuine_percent } });
        output.append(.{ .scalar = .{ .time = time, .value = result.value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.slope } });
        output.append(.{ .scalar = .{ .time = time, .value = result.eta_top } });
        output.append(.{ .scalar = .{ .time = time, .value = result.eta_bottom } });
        _ = self;
        return output;
    }
    pub fn updateScalar(self: *DominantMarketVibration, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *DominantMarketVibration, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *DominantMarketVibration, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *DominantMarketVibration, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *DominantMarketVibration) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(DominantMarketVibration);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "DominantMarketVibration prepared case 0" {
    var ind = try DominantMarketVibration.init(std.testing.allocator, .{ .genuine_threshold = 49.0, .max_period = 80, .max_segments = 5, .min_period = 5, .peak_distance = 2, .phase_cycles = 2.0, .smoothing = 1600.0, .top_k = 3, .window = 200 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_PERIOD();
    const e1 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE();
    const e2 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_PHASE();
    const e3 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_STRENGTH();
    const e4 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_PERCENT();
    const e5 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_SEGMENTS();
    const e6 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_FIT_AMPLITUDE();
    const e7 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_FIT_SAMPLES();
    const e8 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_VALUE();
    const e9 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_SLOPE();
    const e10 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_TOP();
    const e11 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_BOTTOM();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.period);
        try checkValue(e1[index], result.amplitude);
        try checkValue(e2[index], result.phase);
        try checkValue(e3[index], result.strength);
        try checkValue(e4[index], result.genuine_percent);
        try checkValue(e5[index], result.segments);
        try checkValue(e6[index], result.fit_amplitude);
        try checkValue(e7[index], result.fit_samples);
        try checkValue(e8[index], result.value);
        try checkValue(e9[index], result.slope);
        try checkValue(e10[index], result.eta_top);
        try checkValue(e11[index], result.eta_bottom);
    }
}
test "DominantMarketVibration prepared case 1" {
    var ind = try DominantMarketVibration.init(std.testing.allocator, .{ .genuine_threshold = 49.0, .max_period = 12, .max_segments = 5, .min_period = 12, .peak_distance = 2, .phase_cycles = 0.0, .smoothing = 1600.0, .top_k = 1, .window = 60 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_PERIOD();
    const e1 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_AMPLITUDE();
    const e2 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_PHASE();
    const e3 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_STRENGTH();
    const e4 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_GENUINE_PERCENT();
    const e5 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_SEGMENTS();
    const e6 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_FIT_AMPLITUDE();
    const e7 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_FIT_SAMPLES();
    const e8 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_VALUE();
    const e9 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_SLOPE();
    const e10 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_TOP();
    const e11 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_BOTTOM();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.period);
        try checkValue(e1[index], result.amplitude);
        try checkValue(e2[index], result.phase);
        try checkValue(e3[index], result.strength);
        try checkValue(e4[index], result.genuine_percent);
        try checkValue(e5[index], result.segments);
        try checkValue(e6[index], result.fit_amplitude);
        try checkValue(e7[index], result.fit_samples);
        try checkValue(e8[index], result.value);
        try checkValue(e9[index], result.slope);
        try checkValue(e10[index], result.eta_top);
        try checkValue(e11[index], result.eta_bottom);
    }
}
test "DominantMarketVibration prepared case 2" {
    var ind = try DominantMarketVibration.init(std.testing.allocator, .{ .genuine_threshold = 49.0, .max_period = 30, .max_segments = 5, .min_period = 5, .peak_distance = 2, .phase_cycles = 2.0, .smoothing = 0.0, .top_k = 2, .window = 100 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_PERIOD();
    const e1 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE();
    const e2 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_PHASE();
    const e3 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_STRENGTH();
    const e4 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_PERCENT();
    const e5 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_SEGMENTS();
    const e6 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_FIT_AMPLITUDE();
    const e7 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_FIT_SAMPLES();
    const e8 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_VALUE();
    const e9 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_SLOPE();
    const e10 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_TOP();
    const e11 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_BOTTOM();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.period);
        try checkValue(e1[index], result.amplitude);
        try checkValue(e2[index], result.phase);
        try checkValue(e3[index], result.strength);
        try checkValue(e4[index], result.genuine_percent);
        try checkValue(e5[index], result.segments);
        try checkValue(e6[index], result.fit_amplitude);
        try checkValue(e7[index], result.fit_samples);
        try checkValue(e8[index], result.value);
        try checkValue(e9[index], result.slope);
        try checkValue(e10[index], result.eta_top);
        try checkValue(e11[index], result.eta_bottom);
    }
}
test "DominantMarketVibration prepared case 3" {
    var ind = try DominantMarketVibration.init(std.testing.allocator, .{ .genuine_threshold = 100.0, .max_period = 30, .max_segments = 5, .min_period = 5, .peak_distance = 2, .phase_cycles = 2.0, .smoothing = 1600.0, .top_k = 2, .window = 100 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_PERIOD();
    const e1 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE();
    const e2 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_PHASE();
    const e3 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_STRENGTH();
    const e4 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_PERCENT();
    const e5 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_SEGMENTS();
    const e6 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_FIT_AMPLITUDE();
    const e7 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_FIT_SAMPLES();
    const e8 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_VALUE();
    const e9 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_SLOPE();
    const e10 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_TOP();
    const e11 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_BOTTOM();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.period);
        try checkValue(e1[index], result.amplitude);
        try checkValue(e2[index], result.phase);
        try checkValue(e3[index], result.strength);
        try checkValue(e4[index], result.genuine_percent);
        try checkValue(e5[index], result.segments);
        try checkValue(e6[index], result.fit_amplitude);
        try checkValue(e7[index], result.fit_samples);
        try checkValue(e8[index], result.value);
        try checkValue(e9[index], result.slope);
        try checkValue(e10[index], result.eta_top);
        try checkValue(e11[index], result.eta_bottom);
    }
}
test "DominantMarketVibration lifecycle and metadata" {
    var ind = try DominantMarketVibration.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(DominantMarketVibration, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 9), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "DominantMarketVibration component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try DominantMarketVibration.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
