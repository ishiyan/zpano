const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const BartelsPeriodicityScoreParams = struct {
    period: isize = 20,
    segments: isize = 5,
    smoothing: f64 = 0.0,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const BartelsPeriodicityScoreOutput = enum(u8) { score = 1, genuine_percent = 2, segments = 3 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const BartelsPeriodicityScore = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: BartelsPeriodicityScoreParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: BartelsPeriodicityScoreParams) !BartelsPeriodicityScore {
        try validation.integer(allocator, p.period, "period", 3);
        try validation.integer(allocator, p.segments, "segments", 2);
        try validation.real(allocator, p.smoothing, "smoothing", 0.0);
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
        var self = BartelsPeriodicityScore{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "bps({d},{d},{d}{s})", .{ p.period, p.segments, p.smoothing, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Bartels periodicity score {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *BartelsPeriodicityScore) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *BartelsPeriodicityScore) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const BartelsPeriodicityScore) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const BartelsPeriodicityScore) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *BartelsPeriodicityScore, price: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const BartelsPeriodicityScore, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [3][320]u8 = undefined;
        var description_text: [3][448]u8 = undefined;
        var texts: [3]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} score", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} score", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} genuine_percent", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} genuine_percent", .{description}) catch unreachable };
        texts[2] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[2], "{s} segments", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[2], "{s} segments", .{description}) catch unreachable };
        indicators.buildMetadata(out, .bartels_periodicity_score, mnemonic, description, &texts);
    }
    fn wrap(self: *const BartelsPeriodicityScore, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .scalar = .{ .time = time, .value = result.score } });
        output.append(.{ .scalar = .{ .time = time, .value = result.genuine_percent } });
        output.append(.{ .scalar = .{ .time = time, .value = result.segments } });
        _ = self;
        return output;
    }
    pub fn updateScalar(self: *BartelsPeriodicityScore, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *BartelsPeriodicityScore, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *BartelsPeriodicityScore, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *BartelsPeriodicityScore, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *BartelsPeriodicityScore) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(BartelsPeriodicityScore);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "BartelsPeriodicityScore prepared case 0" {
    var ind = try BartelsPeriodicityScore.init(std.testing.allocator, .{ .period = 20, .segments = 5, .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedP20_M5_L0_SCORE();
    const e1 = testdata.expectedP20_M5_L0_GENUINE_PERCENT();
    const e2 = testdata.expectedP20_M5_L0_SEGMENTS();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.score);
        try checkValue(e1[index], result.genuine_percent);
        try checkValue(e2[index], result.segments);
    }
}
test "BartelsPeriodicityScore prepared case 1" {
    var ind = try BartelsPeriodicityScore.init(std.testing.allocator, .{ .period = 3, .segments = 2, .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedP3_M2_L0_SCORE();
    const e1 = testdata.expectedP3_M2_L0_GENUINE_PERCENT();
    const e2 = testdata.expectedP3_M2_L0_SEGMENTS();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.score);
        try checkValue(e1[index], result.genuine_percent);
        try checkValue(e2[index], result.segments);
    }
}
test "BartelsPeriodicityScore prepared case 2" {
    var ind = try BartelsPeriodicityScore.init(std.testing.allocator, .{ .period = 10, .segments = 5, .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedP10_M5_L0_SCORE();
    const e1 = testdata.expectedP10_M5_L0_GENUINE_PERCENT();
    const e2 = testdata.expectedP10_M5_L0_SEGMENTS();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.score);
        try checkValue(e1[index], result.genuine_percent);
        try checkValue(e2[index], result.segments);
    }
}
test "BartelsPeriodicityScore prepared case 3" {
    var ind = try BartelsPeriodicityScore.init(std.testing.allocator, .{ .period = 20, .segments = 2, .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedP20_M2_L0_SCORE();
    const e1 = testdata.expectedP20_M2_L0_GENUINE_PERCENT();
    const e2 = testdata.expectedP20_M2_L0_SEGMENTS();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.score);
        try checkValue(e1[index], result.genuine_percent);
        try checkValue(e2[index], result.segments);
    }
}
test "BartelsPeriodicityScore prepared case 4" {
    var ind = try BartelsPeriodicityScore.init(std.testing.allocator, .{ .period = 20, .segments = 5, .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedP20_M5_L1600_SCORE();
    const e1 = testdata.expectedP20_M5_L1600_GENUINE_PERCENT();
    const e2 = testdata.expectedP20_M5_L1600_SEGMENTS();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.score);
        try checkValue(e1[index], result.genuine_percent);
        try checkValue(e2[index], result.segments);
    }
}
test "BartelsPeriodicityScore prepared case 5" {
    var ind = try BartelsPeriodicityScore.init(std.testing.allocator, .{ .period = 40, .segments = 5, .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedP40_M5_L0_SCORE();
    const e1 = testdata.expectedP40_M5_L0_GENUINE_PERCENT();
    const e2 = testdata.expectedP40_M5_L0_SEGMENTS();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.score);
        try checkValue(e1[index], result.genuine_percent);
        try checkValue(e2[index], result.segments);
    }
}
test "BartelsPeriodicityScore prepared case 6" {
    var ind = try BartelsPeriodicityScore.init(std.testing.allocator, .{ .period = 42, .segments = 6, .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedP42_M6_L0_SCORE();
    const e1 = testdata.expectedP42_M6_L0_GENUINE_PERCENT();
    const e2 = testdata.expectedP42_M6_L0_SEGMENTS();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.score);
        try checkValue(e1[index], result.genuine_percent);
        try checkValue(e2[index], result.segments);
    }
}
test "BartelsPeriodicityScore prepared case 7" {
    var ind = try BartelsPeriodicityScore.init(std.testing.allocator, .{ .period = 60, .segments = 5, .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedP60_M5_L0_SCORE();
    const e1 = testdata.expectedP60_M5_L0_GENUINE_PERCENT();
    const e2 = testdata.expectedP60_M5_L0_SEGMENTS();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], result.score);
        try checkValue(e1[index], result.genuine_percent);
        try checkValue(e2[index], result.segments);
    }
}
test "BartelsPeriodicityScore lifecycle and metadata" {
    var ind = try BartelsPeriodicityScore.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(BartelsPeriodicityScore, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 3), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "BartelsPeriodicityScore component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try BartelsPeriodicityScore.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
