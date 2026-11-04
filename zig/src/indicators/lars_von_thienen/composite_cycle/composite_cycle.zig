const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const CompositeCycleParams = struct {
    window: isize = 100,
    periods: []f64 = @constCast(&[_]f64{ 20.0, 40.0 }),
    smoothing: f64 = 1600.0,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const CompositeCycleOutput = enum(u8) { value = 1, slope = 2, next_value = 3 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const CompositeCycle = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: CompositeCycleParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: CompositeCycleParams) !CompositeCycle {
        try validation.integer(allocator, p.window, "window", 6);
        try validation.real(allocator, p.smoothing, "smoothing", 0.0);
        if (p.periods.len == 0) return error.NoPeriods;
        for (p.periods, 0..) |period, i| {
            try validation.real(allocator, period, "period", 2.0000000000000004);
            for (p.periods[0..i]) |prior| {
                if (prior == period) return error.DuplicatePeriods;
            }
        }
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
        var self = CompositeCycle{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        var periods_buf: [128]u8 = undefined;
        var periods_len: usize = 0;
        for (p.periods, 0..) |period, i| {
            const s = try std.fmt.bufPrint(periods_buf[periods_len..], "{s}{d}", .{ if (i == 0) "" else "_", period });
            periods_len += s.len;
        }
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "ccyc({d},{s},{d}{s})", .{ p.window, periods_buf[0..periods_len], p.smoothing, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Composite cycle {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *CompositeCycle) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *CompositeCycle) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const CompositeCycle) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const CompositeCycle) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *CompositeCycle, price: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const CompositeCycle, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [3][320]u8 = undefined;
        var description_text: [3][448]u8 = undefined;
        var texts: [3]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} value", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} slope", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} slope", .{description}) catch unreachable };
        texts[2] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[2], "{s} next_value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[2], "{s} next_value", .{description}) catch unreachable };
        indicators.buildMetadata(out, .composite_cycle, mnemonic, description, &texts);
    }
    fn wrap(self: *const CompositeCycle, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.slope } });
        output.append(.{ .scalar = .{ .time = time, .value = result.next_value } });
        _ = self;
        return output;
    }
    pub fn updateScalar(self: *CompositeCycle, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *CompositeCycle, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *CompositeCycle, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *CompositeCycle, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *CompositeCycle) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(CompositeCycle);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "CompositeCycle prepared case 0" {
    var ind = try CompositeCycle.init(std.testing.allocator, .{ .window = 100, .periods = @constCast(&[_]f64{ 20.0, 40.0 }), .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_P20_40_L1600_VALUE();
    const e1 = testdata.expectedW100_P20_40_L1600_SLOPE();
    const e2 = testdata.expectedW100_P20_40_L1600_NEXT_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.slope);
        try checkValue(e2[index], result.next_value);
    }
}
test "CompositeCycle prepared case 1" {
    var ind = try CompositeCycle.init(std.testing.allocator, .{ .window = 100, .periods = @constCast(&[_]f64{20.0}), .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_P20_L0_VALUE();
    const e1 = testdata.expectedW100_P20_L0_SLOPE();
    const e2 = testdata.expectedW100_P20_L0_NEXT_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.slope);
        try checkValue(e2[index], result.next_value);
    }
}
test "CompositeCycle prepared case 2" {
    var ind = try CompositeCycle.init(std.testing.allocator, .{ .window = 100, .periods = @constCast(&[_]f64{ 20.0, 40.0 }), .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_P20_40_L0_VALUE();
    const e1 = testdata.expectedW100_P20_40_L0_SLOPE();
    const e2 = testdata.expectedW100_P20_40_L0_NEXT_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.slope);
        try checkValue(e2[index], result.next_value);
    }
}
test "CompositeCycle prepared case 3" {
    var ind = try CompositeCycle.init(std.testing.allocator, .{ .window = 120, .periods = @constCast(&[_]f64{ 10.0, 20.0, 40.0 }), .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW120_P10_20_40_L1600_VALUE();
    const e1 = testdata.expectedW120_P10_20_40_L1600_SLOPE();
    const e2 = testdata.expectedW120_P10_20_40_L1600_NEXT_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.slope);
        try checkValue(e2[index], result.next_value);
    }
}
test "CompositeCycle prepared case 4" {
    var ind = try CompositeCycle.init(std.testing.allocator, .{ .window = 100, .periods = @constCast(&[_]f64{ 17.5, 31.0 }), .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_P17_5_31_L1600_VALUE();
    const e1 = testdata.expectedW100_P17_5_31_L1600_SLOPE();
    const e2 = testdata.expectedW100_P17_5_31_L1600_NEXT_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.slope);
        try checkValue(e2[index], result.next_value);
    }
}
test "CompositeCycle prepared case 5" {
    var ind = try CompositeCycle.init(std.testing.allocator, .{ .window = 200, .periods = @constCast(&[_]f64{ 20.0, 40.0 }), .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW200_P20_40_L1600_VALUE();
    const e1 = testdata.expectedW200_P20_40_L1600_SLOPE();
    const e2 = testdata.expectedW200_P20_40_L1600_NEXT_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.slope);
        try checkValue(e2[index], result.next_value);
    }
}
test "CompositeCycle prepared case 6" {
    var ind = try CompositeCycle.init(std.testing.allocator, .{ .window = 100, .periods = @constCast(&[_]f64{ 20.0, 40.0 }), .smoothing = 160.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_P20_40_L160_VALUE();
    const e1 = testdata.expectedW100_P20_40_L160_SLOPE();
    const e2 = testdata.expectedW100_P20_40_L160_NEXT_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.slope);
        try checkValue(e2[index], result.next_value);
    }
}
test "CompositeCycle prepared case 7" {
    var ind = try CompositeCycle.init(std.testing.allocator, .{ .window = 253, .periods = @constCast(&[_]f64{ 20.0, 40.0 }), .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW253_P20_40_L1600_VALUE();
    const e1 = testdata.expectedW253_P20_40_L1600_SLOPE();
    const e2 = testdata.expectedW253_P20_40_L1600_NEXT_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.slope);
        try checkValue(e2[index], result.next_value);
    }
}
test "CompositeCycle lifecycle and metadata" {
    var ind = try CompositeCycle.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(CompositeCycle, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 3), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "CompositeCycle component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try CompositeCycle.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
