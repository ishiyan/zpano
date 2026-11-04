const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const CycleSwingMomentumParams = struct {
    window: isize = 50,
    fast_smoothing: f64 = 160.0,
    slow_smoothing: f64 = 1600.0,
    band_memory: isize = 34,
    leveling: f64 = 10.0,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const CycleSwingMomentumOutput = enum(u8) { value = 1, lower = 2, upper = 3, score = 4, band = 5 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const CycleSwingMomentum = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: CycleSwingMomentumParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: CycleSwingMomentumParams) !CycleSwingMomentum {
        try validation.integer(allocator, p.window, "window", 6);
        try validation.real(allocator, p.fast_smoothing, "fast_smoothing", 0.0);
        try validation.real(allocator, p.slow_smoothing, "slow_smoothing", 0.0);
        try validation.integer(allocator, p.band_memory, "band_memory", 2);
        try validation.real(allocator, p.leveling, "leveling", 0.0);
        if (p.leveling <= 0 or p.leveling >= 50) return error.InvalidLeveling;
        if (p.fast_smoothing >= p.slow_smoothing) return error.InvalidOrder;
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
        var self = CycleSwingMomentum{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "csmo({d},{d},{d},{d},{d}{s})", .{ p.window, p.fast_smoothing, p.slow_smoothing, p.band_memory, p.leveling, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Cycle swing momentum {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *CycleSwingMomentum) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *CycleSwingMomentum) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const CycleSwingMomentum) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const CycleSwingMomentum) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *CycleSwingMomentum, price: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const CycleSwingMomentum, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [5][320]u8 = undefined;
        var description_text: [5][448]u8 = undefined;
        var texts: [5]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} value", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} lower", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} lower", .{description}) catch unreachable };
        texts[2] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[2], "{s} upper", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[2], "{s} upper", .{description}) catch unreachable };
        texts[3] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[3], "{s} score", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[3], "{s} score", .{description}) catch unreachable };
        texts[4] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[4], "{s} band", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[4], "{s} band", .{description}) catch unreachable };
        indicators.buildMetadata(out, .cycle_swing_momentum, mnemonic, description, &texts);
    }
    fn wrap(_: *const CycleSwingMomentum, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.lower } });
        output.append(.{ .scalar = .{ .time = time, .value = result.upper } });
        output.append(.{ .scalar = .{ .time = time, .value = result.score } });
        output.append(.{ .band = indicators.Band.new(time, result.lower, result.upper) });
        return output;
    }
    pub fn updateScalar(self: *CycleSwingMomentum, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *CycleSwingMomentum, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *CycleSwingMomentum, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *CycleSwingMomentum, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *CycleSwingMomentum) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(CycleSwingMomentum);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "CycleSwingMomentum prepared case 0" {
    var ind = try CycleSwingMomentum.init(std.testing.allocator, .{ .window = 50, .fast_smoothing = 160.0, .slow_smoothing = 1600.0, .band_memory = 34, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW50_F160_S1600_M34_Q10_VALUE();
    const e1 = testdata.expectedW50_F160_S1600_M34_Q10_LOWER();
    const e2 = testdata.expectedW50_F160_S1600_M34_Q10_UPPER();
    const e3 = testdata.expectedW50_F160_S1600_M34_Q10_SCORE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
        try checkValue(e3[index], result.score);
    }
}
test "CycleSwingMomentum prepared case 1" {
    var ind = try CycleSwingMomentum.init(std.testing.allocator, .{ .window = 6, .fast_smoothing = 16.0, .slow_smoothing = 160.0, .band_memory = 5, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW6_F16_S160_M5_Q10_VALUE();
    const e1 = testdata.expectedW6_F16_S160_M5_Q10_LOWER();
    const e2 = testdata.expectedW6_F16_S160_M5_Q10_UPPER();
    const e3 = testdata.expectedW6_F16_S160_M5_Q10_SCORE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
        try checkValue(e3[index], result.score);
    }
}
test "CycleSwingMomentum prepared case 2" {
    var ind = try CycleSwingMomentum.init(std.testing.allocator, .{ .window = 20, .fast_smoothing = 160.0, .slow_smoothing = 1600.0, .band_memory = 20, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW20_F160_S1600_M20_Q10_VALUE();
    const e1 = testdata.expectedW20_F160_S1600_M20_Q10_LOWER();
    const e2 = testdata.expectedW20_F160_S1600_M20_Q10_UPPER();
    const e3 = testdata.expectedW20_F160_S1600_M20_Q10_SCORE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
        try checkValue(e3[index], result.score);
    }
}
test "CycleSwingMomentum prepared case 3" {
    var ind = try CycleSwingMomentum.init(std.testing.allocator, .{ .window = 100, .fast_smoothing = 160.0, .slow_smoothing = 1600.0, .band_memory = 34, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_F160_S1600_M34_Q10_VALUE();
    const e1 = testdata.expectedW100_F160_S1600_M34_Q10_LOWER();
    const e2 = testdata.expectedW100_F160_S1600_M34_Q10_UPPER();
    const e3 = testdata.expectedW100_F160_S1600_M34_Q10_SCORE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
        try checkValue(e3[index], result.score);
    }
}
test "CycleSwingMomentum prepared case 4" {
    var ind = try CycleSwingMomentum.init(std.testing.allocator, .{ .window = 50, .fast_smoothing = 0.0, .slow_smoothing = 1600.0, .band_memory = 34, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW50_F0_S1600_M34_Q10_VALUE();
    const e1 = testdata.expectedW50_F0_S1600_M34_Q10_LOWER();
    const e2 = testdata.expectedW50_F0_S1600_M34_Q10_UPPER();
    const e3 = testdata.expectedW50_F0_S1600_M34_Q10_SCORE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
        try checkValue(e3[index], result.score);
    }
}
test "CycleSwingMomentum prepared case 5" {
    var ind = try CycleSwingMomentum.init(std.testing.allocator, .{ .window = 50, .fast_smoothing = 1600.0, .slow_smoothing = 16000.0, .band_memory = 34, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW50_F1600_S16000_M34_Q10_VALUE();
    const e1 = testdata.expectedW50_F1600_S16000_M34_Q10_LOWER();
    const e2 = testdata.expectedW50_F1600_S16000_M34_Q10_UPPER();
    const e3 = testdata.expectedW50_F1600_S16000_M34_Q10_SCORE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
        try checkValue(e3[index], result.score);
    }
}
test "CycleSwingMomentum prepared case 6" {
    var ind = try CycleSwingMomentum.init(std.testing.allocator, .{ .window = 50, .fast_smoothing = 160.0, .slow_smoothing = 1600.0, .band_memory = 34, .leveling = 5.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW50_F160_S1600_M34_Q5_VALUE();
    const e1 = testdata.expectedW50_F160_S1600_M34_Q5_LOWER();
    const e2 = testdata.expectedW50_F160_S1600_M34_Q5_UPPER();
    const e3 = testdata.expectedW50_F160_S1600_M34_Q5_SCORE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
        try checkValue(e3[index], result.score);
    }
}
test "CycleSwingMomentum prepared case 7" {
    var ind = try CycleSwingMomentum.init(std.testing.allocator, .{ .window = 252, .fast_smoothing = 160.0, .slow_smoothing = 1600.0, .band_memory = 34, .leveling = 10.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW252_F160_S1600_M34_Q10_VALUE();
    const e1 = testdata.expectedW252_F160_S1600_M34_Q10_LOWER();
    const e2 = testdata.expectedW252_F160_S1600_M34_Q10_UPPER();
    const e3 = testdata.expectedW252_F160_S1600_M34_Q10_SCORE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.lower);
        try checkValue(e2[index], result.upper);
        try checkValue(e3[index], result.score);
    }
}
test "CycleSwingMomentum lifecycle and metadata" {
    var ind = try CycleSwingMomentum.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(CycleSwingMomentum, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 5), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "CycleSwingMomentum component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try CycleSwingMomentum.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
