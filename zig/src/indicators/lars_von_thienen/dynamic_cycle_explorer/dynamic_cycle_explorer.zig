const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const DynamicCycleExplorerParams = struct {
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
pub const DynamicCycleExplorerOutput = enum(u8) { value = 1, period = 2, phase = 3, eta_top = 4, eta_bottom = 5, highlighter = 6, highlighter_index = 7, fitted_at_highlighter = 8, delayed_highlighter = 9 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const DynamicCycleExplorer = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: DynamicCycleExplorerParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: DynamicCycleExplorerParams) !DynamicCycleExplorer {
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
        var self = DynamicCycleExplorer{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "dce({d},{d},{d},{d},{d},{d},{d},{d},{d}{s})", .{ p.window, p.min_period, p.max_period, p.smoothing, p.genuine_threshold, p.phase_cycles, p.top_k, p.peak_distance, p.max_segments, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Dynamic cycle explorer {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *DynamicCycleExplorer) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *DynamicCycleExplorer) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const DynamicCycleExplorer) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const DynamicCycleExplorer) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *DynamicCycleExplorer, price: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const DynamicCycleExplorer, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [9][320]u8 = undefined;
        var description_text: [9][448]u8 = undefined;
        var texts: [9]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} value", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} period", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} period", .{description}) catch unreachable };
        texts[2] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[2], "{s} phase", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[2], "{s} phase", .{description}) catch unreachable };
        texts[3] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[3], "{s} eta_top", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[3], "{s} eta_top", .{description}) catch unreachable };
        texts[4] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[4], "{s} eta_bottom", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[4], "{s} eta_bottom", .{description}) catch unreachable };
        texts[5] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[5], "{s} highlighter", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[5], "{s} highlighter", .{description}) catch unreachable };
        texts[6] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[6], "{s} highlighter_index", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[6], "{s} highlighter_index", .{description}) catch unreachable };
        texts[7] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[7], "{s} fitted_at_highlighter", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[7], "{s} fitted_at_highlighter", .{description}) catch unreachable };
        texts[8] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[8], "{s} delayed_highlighter", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[8], "{s} delayed_highlighter", .{description}) catch unreachable };
        indicators.buildMetadata(out, .dynamic_cycle_explorer, mnemonic, description, &texts);
    }
    fn wrap(self: *const DynamicCycleExplorer, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.period } });
        output.append(.{ .scalar = .{ .time = time, .value = result.phase } });
        output.append(.{ .scalar = .{ .time = time, .value = result.eta_top } });
        output.append(.{ .scalar = .{ .time = time, .value = result.eta_bottom } });
        output.append(.{ .scalar = .{ .time = time, .value = result.highlighter } });
        output.append(.{ .scalar = .{ .time = time, .value = result.highlighter_index } });
        output.append(.{ .scalar = .{ .time = time, .value = result.fitted_at_highlighter } });
        if (std.math.isNan(result.highlighter_index)) {
            output.append(.{ .polyline = indicators.Polyline.empty(time) });
        } else {
            const offset: @TypeOf(@as(i32, 0)) = @intFromFloat(@as(f64, @floatFromInt((self.engine.scanner orelse unreachable).samples_seen - 1)) - result.highlighter_index);
            output.append(.{ .polyline = indicators.Polyline.new(time, &.{.{ .offset = offset, .value = result.highlighter }}) });
        }
        return output;
    }
    pub fn updateScalar(self: *DynamicCycleExplorer, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *DynamicCycleExplorer, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *DynamicCycleExplorer, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *DynamicCycleExplorer, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *DynamicCycleExplorer) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(DynamicCycleExplorer);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "DynamicCycleExplorer prepared case 0" {
    var ind = try DynamicCycleExplorer.init(std.testing.allocator, .{ .genuine_threshold = 49.0, .max_period = 80, .max_segments = 5, .min_period = 5, .peak_distance = 2, .phase_cycles = 2.0, .smoothing = 1600.0, .top_k = 3, .window = 200 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_VALUE();
    const e1 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_PERIOD();
    const e2 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_PHASE();
    const e3 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_TOP();
    const e4 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_BOTTOM();
    const e5 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_HIGHLIGHTER();
    const e6 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_HIGHLIGHTER_INDEX();
    const e7 = testdata.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_FITTED_AT_HIGHLIGHTER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.period);
        try checkValue(e2[index], result.phase);
        try checkValue(e3[index], result.eta_top);
        try checkValue(e4[index], result.eta_bottom);
        try checkValue(e5[index], result.highlighter);
        try checkValue(e6[index], result.highlighter_index);
        try checkValue(e7[index], result.fitted_at_highlighter);
    }
}
test "DynamicCycleExplorer prepared case 1" {
    var ind = try DynamicCycleExplorer.init(std.testing.allocator, .{ .genuine_threshold = 49.0, .max_period = 12, .max_segments = 5, .min_period = 12, .peak_distance = 2, .phase_cycles = 0.0, .smoothing = 1600.0, .top_k = 1, .window = 60 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_VALUE();
    const e1 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_PERIOD();
    const e2 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_PHASE();
    const e3 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_TOP();
    const e4 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_BOTTOM();
    const e5 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_HIGHLIGHTER();
    const e6 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_HIGHLIGHTER_INDEX();
    const e7 = testdata.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_FITTED_AT_HIGHLIGHTER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.period);
        try checkValue(e2[index], result.phase);
        try checkValue(e3[index], result.eta_top);
        try checkValue(e4[index], result.eta_bottom);
        try checkValue(e5[index], result.highlighter);
        try checkValue(e6[index], result.highlighter_index);
        try checkValue(e7[index], result.fitted_at_highlighter);
    }
}
test "DynamicCycleExplorer prepared case 2" {
    var ind = try DynamicCycleExplorer.init(std.testing.allocator, .{ .genuine_threshold = 49.0, .max_period = 30, .max_segments = 5, .min_period = 5, .peak_distance = 2, .phase_cycles = 2.0, .smoothing = 0.0, .top_k = 2, .window = 100 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_VALUE();
    const e1 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_PERIOD();
    const e2 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_PHASE();
    const e3 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_TOP();
    const e4 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_BOTTOM();
    const e5 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_HIGHLIGHTER();
    const e6 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_HIGHLIGHTER_INDEX();
    const e7 = testdata.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_FITTED_AT_HIGHLIGHTER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.period);
        try checkValue(e2[index], result.phase);
        try checkValue(e3[index], result.eta_top);
        try checkValue(e4[index], result.eta_bottom);
        try checkValue(e5[index], result.highlighter);
        try checkValue(e6[index], result.highlighter_index);
        try checkValue(e7[index], result.fitted_at_highlighter);
    }
}
test "DynamicCycleExplorer prepared case 3" {
    var ind = try DynamicCycleExplorer.init(std.testing.allocator, .{ .genuine_threshold = 100.0, .max_period = 30, .max_segments = 5, .min_period = 5, .peak_distance = 2, .phase_cycles = 2.0, .smoothing = 1600.0, .top_k = 2, .window = 100 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_VALUE();
    const e1 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_PERIOD();
    const e2 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_PHASE();
    const e3 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_TOP();
    const e4 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_BOTTOM();
    const e5 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_HIGHLIGHTER();
    const e6 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_HIGHLIGHTER_INDEX();
    const e7 = testdata.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_FITTED_AT_HIGHLIGHTER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.period);
        try checkValue(e2[index], result.phase);
        try checkValue(e3[index], result.eta_top);
        try checkValue(e4[index], result.eta_bottom);
        try checkValue(e5[index], result.highlighter);
        try checkValue(e6[index], result.highlighter_index);
        try checkValue(e7[index], result.fitted_at_highlighter);
    }
}
test "DynamicCycleExplorer lifecycle and metadata" {
    var ind = try DynamicCycleExplorer.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(DynamicCycleExplorer, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 9), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "DynamicCycleExplorer component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try DynamicCycleExplorer.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
