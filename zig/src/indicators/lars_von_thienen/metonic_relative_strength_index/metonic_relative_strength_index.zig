const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const MetonicRelativeStrengthIndexParams = struct {
    length: isize = 13,
    delay1: isize = 6939,
    delay2: isize = 4016,
    delay3: isize = 2922,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const MetonicRelativeStrengthIndexOutput = enum(u8) { value = 1, linear_rsi = 2 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const MetonicRelativeStrengthIndex = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: MetonicRelativeStrengthIndexParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: MetonicRelativeStrengthIndexParams) !MetonicRelativeStrengthIndex {
        try validation.integer(allocator, p.length, "length", 1);
        try validation.integer(allocator, p.delay1, "delay1", 1);
        try validation.integer(allocator, p.delay2, "delay2", 1);
        try validation.integer(allocator, p.delay3, "delay3", 1);
        if (p.delay1 == p.delay2 or p.delay1 == p.delay3 or p.delay2 == p.delay3) return error.DuplicateDelays;
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
        var self = MetonicRelativeStrengthIndex{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "mrsi({d},{d},{d},{d}{s})", .{ p.length, p.delay1, p.delay2, p.delay3, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Metonic relative strength index {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *MetonicRelativeStrengthIndex) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *MetonicRelativeStrengthIndex) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const MetonicRelativeStrengthIndex) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const MetonicRelativeStrengthIndex) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *MetonicRelativeStrengthIndex, price: f64, day: ?i64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price, day);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const MetonicRelativeStrengthIndex, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [2][320]u8 = undefined;
        var description_text: [2][448]u8 = undefined;
        var texts: [2]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} value", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} linear_rsi", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} linear_rsi", .{description}) catch unreachable };
        indicators.buildMetadata(out, .metonic_relative_strength_index, mnemonic, description, &texts);
    }
    fn wrap(self: *const MetonicRelativeStrengthIndex, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.linear_rsi } });
        _ = self;
        return output;
    }
    pub fn updateScalar(self: *MetonicRelativeStrengthIndex, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value, (indicators.lvt_metonic.dayFromUnixSeconds(sample.time) catch @panic("use UTC midnight calendar timestamps"))) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *MetonicRelativeStrengthIndex, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*), (indicators.lvt_metonic.dayFromUnixSeconds(sample.time) catch @panic("use UTC midnight calendar timestamps"))) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *MetonicRelativeStrengthIndex, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*), (indicators.lvt_metonic.dayFromUnixSeconds(sample.time) catch @panic("use UTC midnight calendar timestamps"))) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *MetonicRelativeStrengthIndex, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*), (indicators.lvt_metonic.dayFromUnixSeconds(sample.time) catch @panic("use UTC midnight calendar timestamps"))) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *MetonicRelativeStrengthIndex) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(MetonicRelativeStrengthIndex);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "MetonicRelativeStrengthIndex prepared case 0" {
    var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{ .length = 13, .delay1 = 6939, .delay2 = 4016, .delay3 = 2922 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL13_D16939_D24016_D32922_VALUE();
    const e1 = testdata.expectedL13_D16939_D24016_D32922_LINEAR_RSI();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index));
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_rsi);
    }
}
test "MetonicRelativeStrengthIndex prepared case 1" {
    var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{ .length = 13, .delay1 = 19, .delay2 = 11, .delay3 = 8 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL13_D119_D211_D38_VALUE();
    const e1 = testdata.expectedL13_D119_D211_D38_LINEAR_RSI();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index));
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_rsi);
    }
}
test "MetonicRelativeStrengthIndex prepared case 2" {
    var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{ .length = 1, .delay1 = 19, .delay2 = 11, .delay3 = 8 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL1_D119_D211_D38_VALUE();
    const e1 = testdata.expectedL1_D119_D211_D38_LINEAR_RSI();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index));
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_rsi);
    }
}
test "MetonicRelativeStrengthIndex prepared case 3" {
    var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{ .length = 14, .delay1 = 19, .delay2 = 11, .delay3 = 8 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL14_D119_D211_D38_VALUE();
    const e1 = testdata.expectedL14_D119_D211_D38_LINEAR_RSI();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index));
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_rsi);
    }
}
test "MetonicRelativeStrengthIndex prepared case 4" {
    var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{ .length = 50, .delay1 = 100, .delay2 = 60, .delay3 = 40 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL50_D1100_D260_D340_VALUE();
    const e1 = testdata.expectedL50_D1100_D260_D340_LINEAR_RSI();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index));
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_rsi);
    }
}
test "MetonicRelativeStrengthIndex prepared case 5" {
    var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{ .length = 13, .delay1 = 8, .delay2 = 19, .delay3 = 11 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL13_D18_D219_D311_VALUE();
    const e1 = testdata.expectedL13_D18_D219_D311_LINEAR_RSI();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index));
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_rsi);
    }
}
test "MetonicRelativeStrengthIndex prepared case 6" {
    var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{ .length = 13, .delay1 = 238, .delay2 = 100, .delay3 = 50 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL13_D1238_D2100_D350_VALUE();
    const e1 = testdata.expectedL13_D1238_D2100_D350_LINEAR_RSI();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index));
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_rsi);
    }
}
test "MetonicRelativeStrengthIndex prepared case 7" {
    var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{ .length = 13, .delay1 = 239, .delay2 = 100, .delay3 = 50 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL13_D1239_D2100_D350_VALUE();
    const e1 = testdata.expectedL13_D1239_D2100_D350_LINEAR_RSI();
    for (input, 0..) |price, index| {
        const result = try ind.update(price, @intCast(index));
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.linear_rsi);
    }
}
test "MetonicRelativeStrengthIndex lifecycle and metadata" {
    var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(MetonicRelativeStrengthIndex, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "MetonicRelativeStrengthIndex component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}

test "production Metonic calendar delays" {
    const input = testdata.testInput();
    var ind = try MetonicRelativeStrengthIndex.init(std.testing.allocator, .{ .length = 1 });
    defer ind.deinit();
    for (0..6945) |i| {
        const actual = (try ind.update(input[i % input.len], @intCast(i))).value;
        if (i <= 6939) {
            try std.testing.expect(std.math.isNan(actual));
        } else {
            var sum: f64 = 0;
            for ([_]usize{ 6939, 4016, 2922 }) |lag| {
                if (input[(i - lag) % input.len] >= input[(i - lag - 1) % input.len]) sum += 100;
            }
            try std.testing.expect(@abs(actual - sum / 3) < 1e-13);
        }
    }
}
