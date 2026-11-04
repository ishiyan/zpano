const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const CyclePolarityParams = struct {
    window: isize = 100,
    short_period: f64 = 20.0,
    long_period: f64 = 40.0,
    smoothing: f64 = 1600.0,
    tolerance: f64 = 0.0,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const CyclePolarityOutput = enum(u8) { polarity = 1, short_value = 2, long_value = 3, short_slope = 4, long_slope = 5 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const CyclePolarity = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: CyclePolarityParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: CyclePolarityParams) !CyclePolarity {
        try validation.integer(allocator, p.window, "window", 6);
        try validation.real(allocator, p.short_period, "short_period", 2.0000000000000004);
        try validation.real(allocator, p.long_period, "long_period", 2.0000000000000004);
        try validation.real(allocator, p.smoothing, "smoothing", 0.0);
        try validation.real(allocator, p.tolerance, "tolerance", 0.0);
        if (p.short_period >= p.long_period) return error.InvalidOrder;
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
        var self = CyclePolarity{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "cpol({d},{d},{d},{d},{d}{s})", .{ p.window, p.short_period, p.long_period, p.smoothing, p.tolerance, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Cycle polarity {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *CyclePolarity) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *CyclePolarity) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const CyclePolarity) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const CyclePolarity) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *CyclePolarity, price: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const CyclePolarity, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [5][320]u8 = undefined;
        var description_text: [5][448]u8 = undefined;
        var texts: [5]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} polarity", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} polarity", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} short_value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} short_value", .{description}) catch unreachable };
        texts[2] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[2], "{s} long_value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[2], "{s} long_value", .{description}) catch unreachable };
        texts[3] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[3], "{s} short_slope", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[3], "{s} short_slope", .{description}) catch unreachable };
        texts[4] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[4], "{s} long_slope", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[4], "{s} long_slope", .{description}) catch unreachable };
        indicators.buildMetadata(out, .cycle_polarity, mnemonic, description, &texts);
    }
    fn wrap(self: *const CyclePolarity, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.polarity } });
        output.append(.{ .scalar = .{ .time = time, .value = result.short_value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.long_value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.short_slope } });
        output.append(.{ .scalar = .{ .time = time, .value = result.long_slope } });
        _ = self;
        return output;
    }
    pub fn updateScalar(self: *CyclePolarity, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *CyclePolarity, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *CyclePolarity, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *CyclePolarity, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *CyclePolarity) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(CyclePolarity);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "CyclePolarity prepared case 0" {
    var ind = try CyclePolarity.init(std.testing.allocator, .{ .window = 100, .short_period = 20.0, .long_period = 40.0, .smoothing = 1600.0, .tolerance = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_S20_L40_HP1600_T0_POLARITY();
    const e1 = testdata.expectedW100_S20_L40_HP1600_T0_SHORT_VALUE();
    const e2 = testdata.expectedW100_S20_L40_HP1600_T0_LONG_VALUE();
    const e3 = testdata.expectedW100_S20_L40_HP1600_T0_SHORT_SLOPE();
    const e4 = testdata.expectedW100_S20_L40_HP1600_T0_LONG_SLOPE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.polarity);
        try checkValue(e1[index], result.short_value);
        try checkValue(e2[index], result.long_value);
        try checkValue(e3[index], result.short_slope);
        try checkValue(e4[index], result.long_slope);
    }
}
test "CyclePolarity prepared case 1" {
    var ind = try CyclePolarity.init(std.testing.allocator, .{ .window = 100, .short_period = 20.0, .long_period = 40.0, .smoothing = 0.0, .tolerance = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_S20_L40_HP0_T0_POLARITY();
    const e1 = testdata.expectedW100_S20_L40_HP0_T0_SHORT_VALUE();
    const e2 = testdata.expectedW100_S20_L40_HP0_T0_LONG_VALUE();
    const e3 = testdata.expectedW100_S20_L40_HP0_T0_SHORT_SLOPE();
    const e4 = testdata.expectedW100_S20_L40_HP0_T0_LONG_SLOPE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.polarity);
        try checkValue(e1[index], result.short_value);
        try checkValue(e2[index], result.long_value);
        try checkValue(e3[index], result.short_slope);
        try checkValue(e4[index], result.long_slope);
    }
}
test "CyclePolarity prepared case 2" {
    var ind = try CyclePolarity.init(std.testing.allocator, .{ .window = 60, .short_period = 10.0, .long_period = 20.0, .smoothing = 1600.0, .tolerance = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW60_S10_L20_HP1600_T0_POLARITY();
    const e1 = testdata.expectedW60_S10_L20_HP1600_T0_SHORT_VALUE();
    const e2 = testdata.expectedW60_S10_L20_HP1600_T0_LONG_VALUE();
    const e3 = testdata.expectedW60_S10_L20_HP1600_T0_SHORT_SLOPE();
    const e4 = testdata.expectedW60_S10_L20_HP1600_T0_LONG_SLOPE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.polarity);
        try checkValue(e1[index], result.short_value);
        try checkValue(e2[index], result.long_value);
        try checkValue(e3[index], result.short_slope);
        try checkValue(e4[index], result.long_slope);
    }
}
test "CyclePolarity prepared case 3" {
    var ind = try CyclePolarity.init(std.testing.allocator, .{ .window = 200, .short_period = 40.0, .long_period = 80.0, .smoothing = 1600.0, .tolerance = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW200_S40_L80_HP1600_T0_POLARITY();
    const e1 = testdata.expectedW200_S40_L80_HP1600_T0_SHORT_VALUE();
    const e2 = testdata.expectedW200_S40_L80_HP1600_T0_LONG_VALUE();
    const e3 = testdata.expectedW200_S40_L80_HP1600_T0_SHORT_SLOPE();
    const e4 = testdata.expectedW200_S40_L80_HP1600_T0_LONG_SLOPE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.polarity);
        try checkValue(e1[index], result.short_value);
        try checkValue(e2[index], result.long_value);
        try checkValue(e3[index], result.short_slope);
        try checkValue(e4[index], result.long_slope);
    }
}
test "CyclePolarity prepared case 4" {
    var ind = try CyclePolarity.init(std.testing.allocator, .{ .window = 100, .short_period = 17.5, .long_period = 31.0, .smoothing = 1600.0, .tolerance = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_S17_5_L31_HP1600_T0_POLARITY();
    const e1 = testdata.expectedW100_S17_5_L31_HP1600_T0_SHORT_VALUE();
    const e2 = testdata.expectedW100_S17_5_L31_HP1600_T0_LONG_VALUE();
    const e3 = testdata.expectedW100_S17_5_L31_HP1600_T0_SHORT_SLOPE();
    const e4 = testdata.expectedW100_S17_5_L31_HP1600_T0_LONG_SLOPE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.polarity);
        try checkValue(e1[index], result.short_value);
        try checkValue(e2[index], result.long_value);
        try checkValue(e3[index], result.short_slope);
        try checkValue(e4[index], result.long_slope);
    }
}
test "CyclePolarity prepared case 5" {
    var ind = try CyclePolarity.init(std.testing.allocator, .{ .window = 100, .short_period = 20.0, .long_period = 40.0, .smoothing = 1600.0, .tolerance = 0.1 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_S20_L40_HP1600_T0_1_POLARITY();
    const e1 = testdata.expectedW100_S20_L40_HP1600_T0_1_SHORT_VALUE();
    const e2 = testdata.expectedW100_S20_L40_HP1600_T0_1_LONG_VALUE();
    const e3 = testdata.expectedW100_S20_L40_HP1600_T0_1_SHORT_SLOPE();
    const e4 = testdata.expectedW100_S20_L40_HP1600_T0_1_LONG_SLOPE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.polarity);
        try checkValue(e1[index], result.short_value);
        try checkValue(e2[index], result.long_value);
        try checkValue(e3[index], result.short_slope);
        try checkValue(e4[index], result.long_slope);
    }
}
test "CyclePolarity prepared case 6" {
    var ind = try CyclePolarity.init(std.testing.allocator, .{ .window = 100, .short_period = 20.0, .long_period = 40.0, .smoothing = 1600.0, .tolerance = 1000.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_S20_L40_HP1600_T1000_POLARITY();
    const e1 = testdata.expectedW100_S20_L40_HP1600_T1000_SHORT_VALUE();
    const e2 = testdata.expectedW100_S20_L40_HP1600_T1000_LONG_VALUE();
    const e3 = testdata.expectedW100_S20_L40_HP1600_T1000_SHORT_SLOPE();
    const e4 = testdata.expectedW100_S20_L40_HP1600_T1000_LONG_SLOPE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.polarity);
        try checkValue(e1[index], result.short_value);
        try checkValue(e2[index], result.long_value);
        try checkValue(e3[index], result.short_slope);
        try checkValue(e4[index], result.long_slope);
    }
}
test "CyclePolarity prepared case 7" {
    var ind = try CyclePolarity.init(std.testing.allocator, .{ .window = 253, .short_period = 20.0, .long_period = 40.0, .smoothing = 1600.0, .tolerance = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW253_S20_L40_HP1600_T0_POLARITY();
    const e1 = testdata.expectedW253_S20_L40_HP1600_T0_SHORT_VALUE();
    const e2 = testdata.expectedW253_S20_L40_HP1600_T0_LONG_VALUE();
    const e3 = testdata.expectedW253_S20_L40_HP1600_T0_SHORT_SLOPE();
    const e4 = testdata.expectedW253_S20_L40_HP1600_T0_LONG_SLOPE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.polarity);
        try checkValue(e1[index], result.short_value);
        try checkValue(e2[index], result.long_value);
        try checkValue(e3[index], result.short_slope);
        try checkValue(e4[index], result.long_slope);
    }
}
test "CyclePolarity lifecycle and metadata" {
    var ind = try CyclePolarity.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(CyclePolarity, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 5), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "CyclePolarity component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try CyclePolarity.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
