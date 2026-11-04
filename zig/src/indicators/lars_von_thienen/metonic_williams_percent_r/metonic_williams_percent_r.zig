const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const MetonicWilliamsPercentRParams = struct {
    length: isize = 13,
    delay1: isize = 6939,
    delay2: isize = 4016,
    delay3: isize = 2922,
};
pub const MetonicWilliamsPercentROutput = enum(u8) { value = 1, linear_percent_r = 2 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const MetonicWilliamsPercentR = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: MetonicWilliamsPercentRParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: MetonicWilliamsPercentRParams) !MetonicWilliamsPercentR {
        try validation.integer(allocator, p.length, "length", 1);
        try validation.integer(allocator, p.delay1, "delay1", 1);
        try validation.integer(allocator, p.delay2, "delay2", 1);
        try validation.integer(allocator, p.delay3, "delay3", 1);
        if (p.delay1 == p.delay2 or p.delay1 == p.delay3 or p.delay2 == p.delay3) return error.DuplicateDelays;
        const bc = entities.bar_component.default_bar_component;
        const qc = entities.quote_component.default_quote_component;
        const tc = entities.trade_component.default_trade_component;
        const suffix = "";
        var arena = std.heap.ArenaAllocator.init(allocator);
        errdefer arena.deinit();
        const a = arena.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, p) });
        try engine.reset(a);
        var self = MetonicWilliamsPercentR{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "mwillr({d},{d},{d},{d}{s})", .{ p.length, p.delay1, p.delay2, p.delay3, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Metonic williams percent r {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *MetonicWilliamsPercentR) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *MetonicWilliamsPercentR) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const MetonicWilliamsPercentR) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const MetonicWilliamsPercentR) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *MetonicWilliamsPercentR, price: f64, day: ?i64, high: f64, low: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price, day, high, low);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const MetonicWilliamsPercentR, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [2][320]u8 = undefined;
        var description_text: [2][448]u8 = undefined;
        var texts: [2]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} value", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} linear_percent_r", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} linear_percent_r", .{description}) catch unreachable };
        indicators.buildMetadata(out, .metonic_williams_percent_r, mnemonic, description, &texts);
    }
    fn wrap(self: *const MetonicWilliamsPercentR, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.linear_percent_r } });
        _ = self;
        return output;
    }
    pub fn updateScalar(self: *MetonicWilliamsPercentR, sample: *const entities.Scalar) indicators.OutputArray {
        return self.wrap(sample.time, numerical.emptySnapshot(self.arena.allocator()) catch unreachable);
    }
    pub fn updateBar(self: *MetonicWilliamsPercentR, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(sample.close, (indicators.lvt_metonic.dayFromUnixSeconds(sample.time) catch @panic("use UTC midnight calendar timestamps")), sample.high, sample.low) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *MetonicWilliamsPercentR, sample: *const entities.Quote) indicators.OutputArray {
        return self.wrap(sample.time, numerical.emptySnapshot(self.arena.allocator()) catch unreachable);
    }
    pub fn updateTrade(self: *MetonicWilliamsPercentR, sample: *const entities.Trade) indicators.OutputArray {
        return self.wrap(sample.time, numerical.emptySnapshot(self.arena.allocator()) catch unreachable);
    }
    pub fn indicator(self: *MetonicWilliamsPercentR) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(MetonicWilliamsPercentR);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "MetonicWilliamsPercentR prepared case 0" {
    var ind = try MetonicWilliamsPercentR.init(std.testing.allocator, .{ .length = 13, .delay1 = 6939, .delay2 = 4016, .delay3 = 2922 });
    defer ind.deinit();
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const e0 = testdata.expectedL13_D16939_D24016_D32922_VALUE();
    const e1 = testdata.expectedL13_D16939_D24016_D32922_LINEAR_PERCENT_R();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index), high[index], low[index]);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_percent_r);
    }
}
test "MetonicWilliamsPercentR prepared case 1" {
    var ind = try MetonicWilliamsPercentR.init(std.testing.allocator, .{ .length = 13, .delay1 = 19, .delay2 = 11, .delay3 = 8 });
    defer ind.deinit();
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const e0 = testdata.expectedL13_D119_D211_D38_VALUE();
    const e1 = testdata.expectedL13_D119_D211_D38_LINEAR_PERCENT_R();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index), high[index], low[index]);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_percent_r);
    }
}
test "MetonicWilliamsPercentR prepared case 2" {
    var ind = try MetonicWilliamsPercentR.init(std.testing.allocator, .{ .length = 1, .delay1 = 19, .delay2 = 11, .delay3 = 8 });
    defer ind.deinit();
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const e0 = testdata.expectedL1_D119_D211_D38_VALUE();
    const e1 = testdata.expectedL1_D119_D211_D38_LINEAR_PERCENT_R();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index), high[index], low[index]);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_percent_r);
    }
}
test "MetonicWilliamsPercentR prepared case 3" {
    var ind = try MetonicWilliamsPercentR.init(std.testing.allocator, .{ .length = 14, .delay1 = 19, .delay2 = 11, .delay3 = 8 });
    defer ind.deinit();
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const e0 = testdata.expectedL14_D119_D211_D38_VALUE();
    const e1 = testdata.expectedL14_D119_D211_D38_LINEAR_PERCENT_R();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index), high[index], low[index]);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_percent_r);
    }
}
test "MetonicWilliamsPercentR prepared case 4" {
    var ind = try MetonicWilliamsPercentR.init(std.testing.allocator, .{ .length = 50, .delay1 = 100, .delay2 = 60, .delay3 = 40 });
    defer ind.deinit();
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const e0 = testdata.expectedL50_D1100_D260_D340_VALUE();
    const e1 = testdata.expectedL50_D1100_D260_D340_LINEAR_PERCENT_R();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index), high[index], low[index]);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_percent_r);
    }
}
test "MetonicWilliamsPercentR prepared case 5" {
    var ind = try MetonicWilliamsPercentR.init(std.testing.allocator, .{ .length = 13, .delay1 = 8, .delay2 = 19, .delay3 = 11 });
    defer ind.deinit();
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const e0 = testdata.expectedL13_D18_D219_D311_VALUE();
    const e1 = testdata.expectedL13_D18_D219_D311_LINEAR_PERCENT_R();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index), high[index], low[index]);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_percent_r);
    }
}
test "MetonicWilliamsPercentR prepared case 6" {
    var ind = try MetonicWilliamsPercentR.init(std.testing.allocator, .{ .length = 13, .delay1 = 239, .delay2 = 100, .delay3 = 50 });
    defer ind.deinit();
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const e0 = testdata.expectedL13_D1239_D2100_D350_VALUE();
    const e1 = testdata.expectedL13_D1239_D2100_D350_LINEAR_PERCENT_R();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index), high[index], low[index]);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_percent_r);
    }
}
test "MetonicWilliamsPercentR prepared case 7" {
    var ind = try MetonicWilliamsPercentR.init(std.testing.allocator, .{ .length = 13, .delay1 = 240, .delay2 = 100, .delay3 = 50 });
    defer ind.deinit();
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const e0 = testdata.expectedL13_D1240_D2100_D350_VALUE();
    const e1 = testdata.expectedL13_D1240_D2100_D350_LINEAR_PERCENT_R();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index), high[index], low[index]);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_percent_r);
    }
}
test "MetonicWilliamsPercentR lifecycle and metadata" {
    var ind = try MetonicWilliamsPercentR.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(MetonicWilliamsPercentR, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}
