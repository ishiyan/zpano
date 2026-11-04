const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const PhasingOscillatorParams = struct {
    window: isize = 200,
    short_min: isize = 5,
    short_max: isize = 30,
    long_min: isize = 31,
    long_max: isize = 80,
    smoothing: f64 = 1600.0,
    genuine_threshold: f64 = 49.0,
    phase_cycles: f64 = 2.0,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const PhasingOscillatorOutput = enum(u8) { value = 1, short_score = 2, long_score = 3, short_period = 4, long_period = 5 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const PhasingOscillator = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: PhasingOscillatorParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: PhasingOscillatorParams) !PhasingOscillator {
        try validation.integer(allocator, p.window, "window", 6);
        try validation.integer(allocator, p.short_min, "short_min", 3);
        try validation.integer(allocator, p.short_max, "short_max", 3);
        try validation.integer(allocator, p.long_min, "long_min", 3);
        try validation.integer(allocator, p.long_max, "long_max", 3);
        try validation.real(allocator, p.smoothing, "smoothing", 0.0);
        try validation.real(allocator, p.genuine_threshold, "genuine_threshold", 0.0);
        try validation.real(allocator, p.phase_cycles, "phase_cycles", 0.0);
        if (p.short_max >= p.long_min) return error.InvalidOrder;
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
        var self = PhasingOscillator{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "phosc({d},{d},{d},{d},{d},{d},{d},{d}{s})", .{ p.window, p.short_min, p.short_max, p.long_min, p.long_max, p.smoothing, p.genuine_threshold, p.phase_cycles, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Phasing oscillator {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *PhasingOscillator) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *PhasingOscillator) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const PhasingOscillator) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const PhasingOscillator) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *PhasingOscillator, price: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const PhasingOscillator, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [5][320]u8 = undefined;
        var description_text: [5][448]u8 = undefined;
        var texts: [5]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} value", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} value", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} short_score", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} short_score", .{description}) catch unreachable };
        texts[2] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[2], "{s} long_score", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[2], "{s} long_score", .{description}) catch unreachable };
        texts[3] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[3], "{s} short_period", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[3], "{s} short_period", .{description}) catch unreachable };
        texts[4] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[4], "{s} long_period", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[4], "{s} long_period", .{description}) catch unreachable };
        indicators.buildMetadata(out, .phasing_oscillator, mnemonic, description, &texts);
    }
    fn wrap(self: *const PhasingOscillator, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.value } });
        output.append(.{ .scalar = .{ .time = time, .value = result.short_score } });
        output.append(.{ .scalar = .{ .time = time, .value = result.long_score } });
        output.append(.{ .scalar = .{ .time = time, .value = result.short_period } });
        output.append(.{ .scalar = .{ .time = time, .value = result.long_period } });
        _ = self;
        return output;
    }
    pub fn updateScalar(self: *PhasingOscillator, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *PhasingOscillator, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *PhasingOscillator, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *PhasingOscillator, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *PhasingOscillator) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(PhasingOscillator);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "PhasingOscillator prepared case 0" {
    var ind = try PhasingOscillator.init(std.testing.allocator, .{ .genuine_threshold = 49.0, .long_max = 80, .long_min = 31, .phase_cycles = 2.0, .short_max = 30, .short_min = 5, .smoothing = 1600.0, .window = 200 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW200_S05_S130_L031_L180_HP1600_G49_F2_VALUE();
    const e1 = testdata.expectedW200_S05_S130_L031_L180_HP1600_G49_F2_SHORT_SCORE();
    const e2 = testdata.expectedW200_S05_S130_L031_L180_HP1600_G49_F2_LONG_SCORE();
    const e3 = testdata.expectedW200_S05_S130_L031_L180_HP1600_G49_F2_SHORT_PERIOD();
    const e4 = testdata.expectedW200_S05_S130_L031_L180_HP1600_G49_F2_LONG_PERIOD();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.short_score);
        try checkValue(e2[index], result.long_score);
        try checkValue(e3[index], result.short_period);
        try checkValue(e4[index], result.long_period);
    }
}
test "PhasingOscillator prepared case 1" {
    var ind = try PhasingOscillator.init(std.testing.allocator, .{ .genuine_threshold = 49.0, .long_max = 40, .long_min = 16, .phase_cycles = 0.0, .short_max = 15, .short_min = 5, .smoothing = 1600.0, .window = 100 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_S05_S115_L016_L140_HP1600_G49_F0_VALUE();
    const e1 = testdata.expectedW100_S05_S115_L016_L140_HP1600_G49_F0_SHORT_SCORE();
    const e2 = testdata.expectedW100_S05_S115_L016_L140_HP1600_G49_F0_LONG_SCORE();
    const e3 = testdata.expectedW100_S05_S115_L016_L140_HP1600_G49_F0_SHORT_PERIOD();
    const e4 = testdata.expectedW100_S05_S115_L016_L140_HP1600_G49_F0_LONG_PERIOD();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.short_score);
        try checkValue(e2[index], result.long_score);
        try checkValue(e3[index], result.short_period);
        try checkValue(e4[index], result.long_period);
    }
}
test "PhasingOscillator prepared case 2" {
    var ind = try PhasingOscillator.init(std.testing.allocator, .{ .genuine_threshold = 49.0, .long_max = 40, .long_min = 16, .phase_cycles = 2.0, .short_max = 15, .short_min = 5, .smoothing = 0.0, .window = 100 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_S05_S115_L016_L140_HP0_G49_F2_VALUE();
    const e1 = testdata.expectedW100_S05_S115_L016_L140_HP0_G49_F2_SHORT_SCORE();
    const e2 = testdata.expectedW100_S05_S115_L016_L140_HP0_G49_F2_LONG_SCORE();
    const e3 = testdata.expectedW100_S05_S115_L016_L140_HP0_G49_F2_SHORT_PERIOD();
    const e4 = testdata.expectedW100_S05_S115_L016_L140_HP0_G49_F2_LONG_PERIOD();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.short_score);
        try checkValue(e2[index], result.long_score);
        try checkValue(e3[index], result.short_period);
        try checkValue(e4[index], result.long_period);
    }
}
test "PhasingOscillator prepared case 3" {
    var ind = try PhasingOscillator.init(std.testing.allocator, .{ .genuine_threshold = 100.0, .long_max = 80, .long_min = 31, .phase_cycles = 2.0, .short_max = 30, .short_min = 5, .smoothing = 1600.0, .window = 200 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW200_S05_S130_L031_L180_HP1600_G100_F2_VALUE();
    const e1 = testdata.expectedW200_S05_S130_L031_L180_HP1600_G100_F2_SHORT_SCORE();
    const e2 = testdata.expectedW200_S05_S130_L031_L180_HP1600_G100_F2_LONG_SCORE();
    const e3 = testdata.expectedW200_S05_S130_L031_L180_HP1600_G100_F2_SHORT_PERIOD();
    const e4 = testdata.expectedW200_S05_S130_L031_L180_HP1600_G100_F2_LONG_PERIOD();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.value);
        try checkValue(e1[index], result.short_score);
        try checkValue(e2[index], result.long_score);
        try checkValue(e3[index], result.short_period);
        try checkValue(e4[index], result.long_period);
    }
}
test "PhasingOscillator lifecycle and metadata" {
    var ind = try PhasingOscillator.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(PhasingOscillator, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 5), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "PhasingOscillator component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try PhasingOscillator.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
