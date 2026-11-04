// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const DetrendedRhythmOscillatorsParams = struct {
    fast_length: isize = 3,
    slow_length: isize = 50,
    zig_period: isize = 30,
    cycle_memory: isize = 15,
    detrend: bool = true,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const DetrendedRhythmOscillatorsOutput = enum(u8) { value = 1, direction = 2, cycle_length = 3, pivot_value = 4, pivot_offset = 5 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const DetrendedRhythmOscillators = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: DetrendedRhythmOscillatorsParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: DetrendedRhythmOscillatorsParams) !DetrendedRhythmOscillators {
        try validation.integer(allocator, p.fast_length, "fast_length", 1);
        try validation.integer(allocator, p.slow_length, "slow_length", 2);
        try validation.integer(allocator, p.zig_period, "zig_period", 5);
        try validation.integer(allocator, p.cycle_memory, "cycle_memory", 3);
        if (p.fast_length >= p.slow_length) return error.InvalidOrder;
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
        var self = DetrendedRhythmOscillators{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "dro({d},{d},{d},{d},{d}{s})", .{ p.fast_length, p.slow_length, p.zig_period, p.cycle_memory, @intFromBool(p.detrend), suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Detrended rhythm oscillators {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *DetrendedRhythmOscillators) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *DetrendedRhythmOscillators) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const DetrendedRhythmOscillators) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const DetrendedRhythmOscillators) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *DetrendedRhythmOscillators, price: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const DetrendedRhythmOscillators, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [5][320]u8 = undefined;
        var description_text: [5][448]u8 = undefined;
        var texts: [5]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} value", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} direction", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} direction", .{description}) catch unreachable };
        texts[2] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[2], "{s} cycle_length", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[2], "{s} cycle_length", .{description}) catch unreachable };
        texts[3] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[3], "{s} pivot_value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[3], "{s} pivot_value", .{description}) catch unreachable };
        texts[4] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[4], "{s} pivot_offset", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[4], "{s} pivot_offset", .{description}) catch unreachable };
        indicators.buildMetadata(out, .detrended_rhythm_oscillators, mnemonic, description, &texts);
    }
    fn wrap(self: *const DetrendedRhythmOscillators, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.direction } });
        output.append(.{ .scalar = .{ .time = time, .value = result.cycle_length } });
        output.append(.{ .scalar = .{ .time = time, .value = result.pivot_value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.pivot_offset } });
        _ = self;
        return output;
    }
    pub fn updateScalar(self: *DetrendedRhythmOscillators, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *DetrendedRhythmOscillators, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *DetrendedRhythmOscillators, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *DetrendedRhythmOscillators, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *DetrendedRhythmOscillators) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(DetrendedRhythmOscillators);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "DetrendedRhythmOscillators prepared case 0" {
    var ind = try DetrendedRhythmOscillators.init(std.testing.allocator, .{ .fast_length = 3, .slow_length = 50, .zig_period = 30, .cycle_memory = 15, .detrend = true });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedF3_S50_Z30_M15_D1_VALUE();
    const e1 = testdata.expectedF3_S50_Z30_M15_D1_DIRECTION();
    const e2 = testdata.expectedF3_S50_Z30_M15_D1_CYCLE_LENGTH();
    const e3 = testdata.expectedF3_S50_Z30_M15_D1_PIVOT_VALUE();
    const e4 = testdata.expectedF3_S50_Z30_M15_D1_PIVOT_OFFSET();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.direction);
        try checkValue(e2[index], result.cycle_length);
        try checkValue(e3[index], result.pivot_value);
        try checkValue(e4[index], result.pivot_offset);
    }
}
test "DetrendedRhythmOscillators prepared case 1" {
    var ind = try DetrendedRhythmOscillators.init(std.testing.allocator, .{ .fast_length = 3, .slow_length = 50, .zig_period = 5, .cycle_memory = 15, .detrend = true });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedF3_S50_Z5_M15_D1_VALUE();
    const e1 = testdata.expectedF3_S50_Z5_M15_D1_DIRECTION();
    const e2 = testdata.expectedF3_S50_Z5_M15_D1_CYCLE_LENGTH();
    const e3 = testdata.expectedF3_S50_Z5_M15_D1_PIVOT_VALUE();
    const e4 = testdata.expectedF3_S50_Z5_M15_D1_PIVOT_OFFSET();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.direction);
        try checkValue(e2[index], result.cycle_length);
        try checkValue(e3[index], result.pivot_value);
        try checkValue(e4[index], result.pivot_offset);
    }
}
test "DetrendedRhythmOscillators prepared case 2" {
    var ind = try DetrendedRhythmOscillators.init(std.testing.allocator, .{ .fast_length = 3, .slow_length = 50, .zig_period = 10, .cycle_memory = 15, .detrend = true });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedF3_S50_Z10_M15_D1_VALUE();
    const e1 = testdata.expectedF3_S50_Z10_M15_D1_DIRECTION();
    const e2 = testdata.expectedF3_S50_Z10_M15_D1_CYCLE_LENGTH();
    const e3 = testdata.expectedF3_S50_Z10_M15_D1_PIVOT_VALUE();
    const e4 = testdata.expectedF3_S50_Z10_M15_D1_PIVOT_OFFSET();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.direction);
        try checkValue(e2[index], result.cycle_length);
        try checkValue(e3[index], result.pivot_value);
        try checkValue(e4[index], result.pivot_offset);
    }
}
test "DetrendedRhythmOscillators prepared case 3" {
    var ind = try DetrendedRhythmOscillators.init(std.testing.allocator, .{ .fast_length = 2, .slow_length = 20, .zig_period = 10, .cycle_memory = 15, .detrend = true });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedF2_S20_Z10_M15_D1_VALUE();
    const e1 = testdata.expectedF2_S20_Z10_M15_D1_DIRECTION();
    const e2 = testdata.expectedF2_S20_Z10_M15_D1_CYCLE_LENGTH();
    const e3 = testdata.expectedF2_S20_Z10_M15_D1_PIVOT_VALUE();
    const e4 = testdata.expectedF2_S20_Z10_M15_D1_PIVOT_OFFSET();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.direction);
        try checkValue(e2[index], result.cycle_length);
        try checkValue(e3[index], result.pivot_value);
        try checkValue(e4[index], result.pivot_offset);
    }
}
test "DetrendedRhythmOscillators prepared case 4" {
    var ind = try DetrendedRhythmOscillators.init(std.testing.allocator, .{ .fast_length = 5, .slow_length = 100, .zig_period = 30, .cycle_memory = 15, .detrend = true });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedF5_S100_Z30_M15_D1_VALUE();
    const e1 = testdata.expectedF5_S100_Z30_M15_D1_DIRECTION();
    const e2 = testdata.expectedF5_S100_Z30_M15_D1_CYCLE_LENGTH();
    const e3 = testdata.expectedF5_S100_Z30_M15_D1_PIVOT_VALUE();
    const e4 = testdata.expectedF5_S100_Z30_M15_D1_PIVOT_OFFSET();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.direction);
        try checkValue(e2[index], result.cycle_length);
        try checkValue(e3[index], result.pivot_value);
        try checkValue(e4[index], result.pivot_offset);
    }
}
test "DetrendedRhythmOscillators prepared case 5" {
    var ind = try DetrendedRhythmOscillators.init(std.testing.allocator, .{ .fast_length = 3, .slow_length = 50, .zig_period = 5, .cycle_memory = 3, .detrend = true });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedF3_S50_Z5_M3_D1_VALUE();
    const e1 = testdata.expectedF3_S50_Z5_M3_D1_DIRECTION();
    const e2 = testdata.expectedF3_S50_Z5_M3_D1_CYCLE_LENGTH();
    const e3 = testdata.expectedF3_S50_Z5_M3_D1_PIVOT_VALUE();
    const e4 = testdata.expectedF3_S50_Z5_M3_D1_PIVOT_OFFSET();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.direction);
        try checkValue(e2[index], result.cycle_length);
        try checkValue(e3[index], result.pivot_value);
        try checkValue(e4[index], result.pivot_offset);
    }
}
test "DetrendedRhythmOscillators prepared case 6" {
    var ind = try DetrendedRhythmOscillators.init(std.testing.allocator, .{ .fast_length = 3, .slow_length = 50, .zig_period = 10, .cycle_memory = 15, .detrend = false });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedF3_S50_Z10_M15_D0_VALUE();
    const e1 = testdata.expectedF3_S50_Z10_M15_D0_DIRECTION();
    const e2 = testdata.expectedF3_S50_Z10_M15_D0_CYCLE_LENGTH();
    const e3 = testdata.expectedF3_S50_Z10_M15_D0_PIVOT_VALUE();
    const e4 = testdata.expectedF3_S50_Z10_M15_D0_PIVOT_OFFSET();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.direction);
        try checkValue(e2[index], result.cycle_length);
        try checkValue(e3[index], result.pivot_value);
        try checkValue(e4[index], result.pivot_offset);
    }
}
test "DetrendedRhythmOscillators prepared case 7" {
    var ind = try DetrendedRhythmOscillators.init(std.testing.allocator, .{ .fast_length = 3, .slow_length = 300, .zig_period = 30, .cycle_memory = 15, .detrend = true });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedF3_S300_Z30_M15_D1_VALUE();
    const e1 = testdata.expectedF3_S300_Z30_M15_D1_DIRECTION();
    const e2 = testdata.expectedF3_S300_Z30_M15_D1_CYCLE_LENGTH();
    const e3 = testdata.expectedF3_S300_Z30_M15_D1_PIVOT_VALUE();
    const e4 = testdata.expectedF3_S300_Z30_M15_D1_PIVOT_OFFSET();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.direction);
        try checkValue(e2[index], result.cycle_length);
        try checkValue(e3[index], result.pivot_value);
        try checkValue(e4[index], result.pivot_offset);
    }
}
test "DetrendedRhythmOscillators lifecycle and metadata" {
    var ind = try DetrendedRhythmOscillators.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(DetrendedRhythmOscillators, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 5), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "DetrendedRhythmOscillators component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try DetrendedRhythmOscillators.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
