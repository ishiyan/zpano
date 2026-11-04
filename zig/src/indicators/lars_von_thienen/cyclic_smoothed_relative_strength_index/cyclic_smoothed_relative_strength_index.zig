const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const CyclicSmoothedRelativeStrengthIndexParams = struct {
    dominant_cycle: isize = 20,
    vibration: isize = 10,
    leveling: f64 = 10.0,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const CyclicSmoothedRelativeStrengthIndexOutput = enum(u8) { value = 1, lower = 2, upper = 3, band = 4 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const CyclicSmoothedRelativeStrengthIndex = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: CyclicSmoothedRelativeStrengthIndexParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: CyclicSmoothedRelativeStrengthIndexParams) !CyclicSmoothedRelativeStrengthIndex {
        try validation.integer(allocator, p.dominant_cycle, "dominant_cycle", 10);
        try validation.integer(allocator, p.vibration, "vibration", 2);
        try validation.real(allocator, p.leveling, "leveling", 0.0);
        if (p.leveling <= 0 or p.leveling >= 50) return error.InvalidLeveling;
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
        var self = CyclicSmoothedRelativeStrengthIndex{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "crsi({d},{d},{d}{s})", .{ p.dominant_cycle, p.vibration, p.leveling, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Cyclic smoothed relative strength index {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *CyclicSmoothedRelativeStrengthIndex) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *CyclicSmoothedRelativeStrengthIndex) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const CyclicSmoothedRelativeStrengthIndex) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const CyclicSmoothedRelativeStrengthIndex) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *CyclicSmoothedRelativeStrengthIndex, price: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const CyclicSmoothedRelativeStrengthIndex, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [4][320]u8 = undefined;
        var description_text: [4][448]u8 = undefined;
        var texts: [4]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} value", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} lower", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} lower", .{description}) catch unreachable };
        texts[2] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[2], "{s} upper", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[2], "{s} upper", .{description}) catch unreachable };
        texts[3] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[3], "{s} band", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[3], "{s} band", .{description}) catch unreachable };
        indicators.buildMetadata(out, .cyclic_smoothed_relative_strength_index, mnemonic, description, &texts);
    }
    fn wrap(_: *const CyclicSmoothedRelativeStrengthIndex, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.lower } });
        output.append(.{ .scalar = .{ .time = time, .value = result.upper } });
        output.append(.{ .band = indicators.Band.new(time, result.lower, result.upper) });
        return output;
    }
    pub fn updateScalar(self: *CyclicSmoothedRelativeStrengthIndex, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *CyclicSmoothedRelativeStrengthIndex, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *CyclicSmoothedRelativeStrengthIndex, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *CyclicSmoothedRelativeStrengthIndex, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *CyclicSmoothedRelativeStrengthIndex) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(CyclicSmoothedRelativeStrengthIndex);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "CyclicSmoothedRelativeStrengthIndex prepared case 0" {
    var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{ .dominant_cycle = 20, .vibration = 10, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedD20_V10_Q10_VALUE();
    const e1 = testdata.expectedD20_V10_Q10_LOWER();
    const e2 = testdata.expectedD20_V10_Q10_UPPER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
    }
}
test "CyclicSmoothedRelativeStrengthIndex prepared case 1" {
    var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{ .dominant_cycle = 10, .vibration = 10, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedD10_V10_Q10_VALUE();
    const e1 = testdata.expectedD10_V10_Q10_LOWER();
    const e2 = testdata.expectedD10_V10_Q10_UPPER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
    }
}
test "CyclicSmoothedRelativeStrengthIndex prepared case 2" {
    var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{ .dominant_cycle = 21, .vibration = 10, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedD21_V10_Q10_VALUE();
    const e1 = testdata.expectedD21_V10_Q10_LOWER();
    const e2 = testdata.expectedD21_V10_Q10_UPPER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
    }
}
test "CyclicSmoothedRelativeStrengthIndex prepared case 3" {
    var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{ .dominant_cycle = 40, .vibration = 10, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedD40_V10_Q10_VALUE();
    const e1 = testdata.expectedD40_V10_Q10_LOWER();
    const e2 = testdata.expectedD40_V10_Q10_UPPER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
    }
}
test "CyclicSmoothedRelativeStrengthIndex prepared case 4" {
    var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{ .dominant_cycle = 100, .vibration = 10, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedD100_V10_Q10_VALUE();
    const e1 = testdata.expectedD100_V10_Q10_LOWER();
    const e2 = testdata.expectedD100_V10_Q10_UPPER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
    }
}
test "CyclicSmoothedRelativeStrengthIndex prepared case 5" {
    var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{ .dominant_cycle = 20, .vibration = 3, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedD20_V3_Q10_VALUE();
    const e1 = testdata.expectedD20_V3_Q10_LOWER();
    const e2 = testdata.expectedD20_V3_Q10_UPPER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
    }
}
test "CyclicSmoothedRelativeStrengthIndex prepared case 6" {
    var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{ .dominant_cycle = 20, .vibration = 6, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedD20_V6_Q10_VALUE();
    const e1 = testdata.expectedD20_V6_Q10_LOWER();
    const e2 = testdata.expectedD20_V6_Q10_UPPER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
    }
}
test "CyclicSmoothedRelativeStrengthIndex prepared case 7" {
    var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{ .dominant_cycle = 20, .vibration = 10, .leveling = 5.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedD20_V10_Q5_VALUE();
    const e1 = testdata.expectedD20_V10_Q5_LOWER();
    const e2 = testdata.expectedD20_V10_Q5_UPPER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
    }
}
test "CyclicSmoothedRelativeStrengthIndex prepared case 8" {
    var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{ .dominant_cycle = 20, .vibration = 10, .leveling = 20.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedD20_V10_Q20_VALUE();
    const e1 = testdata.expectedD20_V10_Q20_LOWER();
    const e2 = testdata.expectedD20_V10_Q20_UPPER();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
    }
}
test "CyclicSmoothedRelativeStrengthIndex lifecycle and metadata" {
    var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(CyclicSmoothedRelativeStrengthIndex, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 4), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "CyclicSmoothedRelativeStrengthIndex component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try CyclicSmoothedRelativeStrengthIndex.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
