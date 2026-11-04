const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const LarsVonThienenZeroLagExponentialMovingAverageParams = struct {
    length: isize = 13,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const LarsVonThienenZeroLagExponentialMovingAverageOutput = enum(u8) { value = 1 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const LarsVonThienenZeroLagExponentialMovingAverage = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: LarsVonThienenZeroLagExponentialMovingAverageParams,
    line: indicators.LineIndicator,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: LarsVonThienenZeroLagExponentialMovingAverageParams) !LarsVonThienenZeroLagExponentialMovingAverage {
        try validation.integer(allocator, p.length, "length", 1);
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
        var self = LarsVonThienenZeroLagExponentialMovingAverage{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .line = indicators.LineIndicator.new("", "", bc, qc, tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "lzlema({d}{s})", .{ p.length, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Zero lag exponential moving average {s}", .{mnemonic});
        self.description_len = description.len;
        self.fixSlices();
        return self;
    }
    pub fn fixSlices(self: *LarsVonThienenZeroLagExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }
    pub fn deinit(self: *LarsVonThienenZeroLagExponentialMovingAverage) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *LarsVonThienenZeroLagExponentialMovingAverage) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const LarsVonThienenZeroLagExponentialMovingAverage) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const LarsVonThienenZeroLagExponentialMovingAverage) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *LarsVonThienenZeroLagExponentialMovingAverage, price: f64) !f64 {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output.value;
    }
    pub fn getMetadata(self: *const LarsVonThienenZeroLagExponentialMovingAverage, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [1][320]u8 = undefined;
        var description_text: [1][448]u8 = undefined;
        var texts: [1]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} value", .{description}) catch unreachable };
        indicators.buildMetadata(out, .lars_von_thienen_zero_lag_exponential_moving_average, mnemonic, description, &texts);
    }
    fn wrap(self: *const LarsVonThienenZeroLagExponentialMovingAverage, time: i64, result: f64) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result } });
        _ = self;
        return output;
    }
    pub fn updateScalar(self: *LarsVonThienenZeroLagExponentialMovingAverage, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *LarsVonThienenZeroLagExponentialMovingAverage, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.line.extractBar(sample)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *LarsVonThienenZeroLagExponentialMovingAverage, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.line.extractQuote(sample)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *LarsVonThienenZeroLagExponentialMovingAverage, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.line.extractTrade(sample)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *LarsVonThienenZeroLagExponentialMovingAverage) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(LarsVonThienenZeroLagExponentialMovingAverage);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "LarsVonThienenZeroLagExponentialMovingAverage prepared case 0" {
    var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{ .length = 13 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL13_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result);
    }
}
test "LarsVonThienenZeroLagExponentialMovingAverage prepared case 1" {
    var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{ .length = 1 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL1_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result);
    }
}
test "LarsVonThienenZeroLagExponentialMovingAverage prepared case 2" {
    var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{ .length = 2 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL2_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result);
    }
}
test "LarsVonThienenZeroLagExponentialMovingAverage prepared case 3" {
    var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{ .length = 3 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL3_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result);
    }
}
test "LarsVonThienenZeroLagExponentialMovingAverage prepared case 4" {
    var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{ .length = 10 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL10_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result);
    }
}
test "LarsVonThienenZeroLagExponentialMovingAverage prepared case 5" {
    var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{ .length = 20 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL20_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result);
    }
}
test "LarsVonThienenZeroLagExponentialMovingAverage prepared case 6" {
    var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{ .length = 50 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL50_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result);
    }
}
test "LarsVonThienenZeroLagExponentialMovingAverage prepared case 7" {
    var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{ .length = 252 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL252_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result);
    }
}
test "LarsVonThienenZeroLagExponentialMovingAverage prepared case 8" {
    var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{ .length = 600 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedL600_VALUE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result);
    }
}
test "LarsVonThienenZeroLagExponentialMovingAverage lifecycle and metadata" {
    var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(LarsVonThienenZeroLagExponentialMovingAverage, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "LarsVonThienenZeroLagExponentialMovingAverage component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try LarsVonThienenZeroLagExponentialMovingAverage.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
