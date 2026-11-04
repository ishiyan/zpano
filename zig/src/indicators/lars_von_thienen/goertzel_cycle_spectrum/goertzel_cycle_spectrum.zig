const nan = std.math.nan(f64);
const std = @import("std");
const entities = @import("entities");
const indicators = @import("indicators");
const numerical = @import("engine.zig");
const memory = @import("../harmonic_cycle/memory.zig");
const validation = @import("../harmonic_cycle/validation.zig");
pub const Snapshot = numerical.Snapshot;
/// Prepared research parameters; omitted components use entity defaults.
pub const GoertzelCycleSpectrumParams = struct {
    window: isize = 128,
    min_period: isize = 5,
    max_period: isize = 50,
    step: isize = 1,
    smoothing: f64 = 1600.0,
    bar_component: ?entities.BarComponent = null,
    quote_component: ?entities.QuoteComponent = null,
    trade_component: ?entities.TradeComponent = null,
};
pub const GoertzelCycleSpectrumOutput = enum(u8) { amplitude = 1, phases = 2, trough_offsets = 3, dominant_period = 4, dominant_amplitude = 5, dominant_phase = 6 };
/// Arrays in a Snapshot remain valid until the next update, reset or deinit.
/// Entity outputs own their heatmap/polyline buffers. Updates clone state into
/// a reusable scratch arena and commit only after successful calculation.
pub const GoertzelCycleSpectrum = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    scratch: std.heap.ArenaAllocator,
    engine: *numerical.Engine,
    p: GoertzelCycleSpectrumParams,
    bar_func: *const fn (entities.Bar) f64,
    quote_func: *const fn (entities.Quote) f64,
    trade_func: *const fn (entities.Trade) f64,
    mnemonic_buf: [256]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [384]u8 = undefined,
    description_len: usize = 0,
    pub fn init(allocator: std.mem.Allocator, p: GoertzelCycleSpectrumParams) !GoertzelCycleSpectrum {
        try validation.integer(allocator, p.window, "window", 6);
        try validation.integer(allocator, p.min_period, "min_period", 3);
        try validation.integer(allocator, p.max_period, "max_period", 3);
        try validation.integer(allocator, p.step, "step", 1);
        try validation.real(allocator, p.smoothing, "smoothing", 0.0);
        if (p.min_period > p.max_period) return error.InvalidRange;
        if (@rem(p.max_period - p.min_period, p.step) != 0) return error.InvalidStep;
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
        var self = GoertzelCycleSpectrum{ .allocator = allocator, .arena = arena, .scratch = std.heap.ArenaAllocator.init(allocator), .engine = engine, .p = p, .bar_func = entities.bar_component.componentValue(bc), .quote_func = entities.quote_component.componentValue(qc), .trade_func = entities.trade_component.componentValue(tc) };
        const mnemonic = try std.fmt.bufPrint(&self.mnemonic_buf, "gcs({d},{d},{d},{d},{d}{s})", .{ p.window, p.min_period, p.max_period, p.step, p.smoothing, suffix });
        self.mnemonic_len = mnemonic.len;
        const description = try std.fmt.bufPrint(&self.description_buf, "Goertzel cycle spectrum {s}", .{mnemonic});
        self.description_len = description.len;
        return self;
    }
    pub fn deinit(self: *GoertzelCycleSpectrum) void {
        self.arena.deinit();
        self.scratch.deinit();
    }
    pub fn reset(self: *GoertzelCycleSpectrum) !void {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.create(a, numerical.Engine{ .p = try memory.clone(a, self.engine.p) });
        try engine.reset(a);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
    }
    pub fn isPrimed(self: *const GoertzelCycleSpectrum) bool {
        return self.engine.primed(self.allocator) catch false;
    }
    pub fn warmupBars(self: *const GoertzelCycleSpectrum) isize {
        return self.engine.warmup(self.allocator) catch unreachable;
    }
    pub fn update(self: *GoertzelCycleSpectrum, price: f64) !Snapshot {
        _ = self.scratch.reset(.retain_capacity);
        const a = self.scratch.allocator();
        const engine = try memory.clone(a, self.engine);
        const output = try engine.update(a, price);
        std.mem.swap(std.heap.ArenaAllocator, &self.arena, &self.scratch);
        self.engine = engine;
        return output;
    }
    pub fn getMetadata(self: *const GoertzelCycleSpectrum, out: *indicators.Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];
        var mnemonic_text: [6][320]u8 = undefined;
        var description_text: [6][448]u8 = undefined;
        var texts: [6]indicators.OutputText = undefined;
        texts[0] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[0], "{s} amplitude", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[0], "{s} amplitude", .{description}) catch unreachable };
        texts[1] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[1], "{s} phases", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[1], "{s} phases", .{description}) catch unreachable };
        texts[2] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[2], "{s} trough_offsets", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[2], "{s} trough_offsets", .{description}) catch unreachable };
        texts[3] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[3], "{s} dominant_period", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[3], "{s} dominant_period", .{description}) catch unreachable };
        texts[4] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[4], "{s} dominant_amplitude", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[4], "{s} dominant_amplitude", .{description}) catch unreachable };
        texts[5] = .{ .mnemonic = std.fmt.bufPrint(&mnemonic_text[5], "{s} dominant_phase", .{mnemonic}) catch unreachable, .description = std.fmt.bufPrint(&description_text[5], "{s} dominant_phase", .{description}) catch unreachable };
        indicators.buildMetadata(out, .goertzel_cycle_spectrum, mnemonic, description, &texts);
    }
    fn wrap(self: *const GoertzelCycleSpectrum, time: i64, result: Snapshot) indicators.OutputArray {
        var output = indicators.OutputArray{};
        output.append(.{ .heatmap = indicators.Heatmap.new(time, @floatFromInt(self.p.min_period), @floatFromInt(self.p.max_period), 1.0 / @as(f64, @floatFromInt(self.p.step)), nan, nan, result.amplitude) });
        output.append(.{ .heatmap = indicators.Heatmap.new(time, @floatFromInt(self.p.min_period), @floatFromInt(self.p.max_period), 1.0 / @as(f64, @floatFromInt(self.p.step)), nan, nan, result.phases) });
        output.append(.{ .heatmap = indicators.Heatmap.new(time, @floatFromInt(self.p.min_period), @floatFromInt(self.p.max_period), 1.0 / @as(f64, @floatFromInt(self.p.step)), nan, nan, result.trough_offsets) });
        output.append(.{ .scalar = .{ .time = time, .value = result.dominant_period } });
        output.append(.{ .scalar = .{ .time = time, .value = result.dominant_amplitude } });
        output.append(.{ .scalar = .{ .time = time, .value = result.dominant_phase } });
        return output;
    }
    pub fn updateScalar(self: *GoertzelCycleSpectrum, sample: *const entities.Scalar) indicators.OutputArray {
        const result = self.update(sample.value) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateBar(self: *GoertzelCycleSpectrum, sample: *const entities.Bar) indicators.OutputArray {
        const result = self.update(self.bar_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateQuote(self: *GoertzelCycleSpectrum, sample: *const entities.Quote) indicators.OutputArray {
        const result = self.update(self.quote_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn updateTrade(self: *GoertzelCycleSpectrum, sample: *const entities.Trade) indicators.OutputArray {
        const result = self.update(self.trade_func(sample.*)) catch @panic("invalid observation or allocation failure");
        return self.wrap(sample.time, result);
    }
    pub fn indicator(self: *GoertzelCycleSpectrum) indicators.Indicator {
        return .{ .ptr = self, .vtable = &vtable };
    }
    const vtable = indicators.Indicator.GenVTable(GoertzelCycleSpectrum);
};
const testdata = @import("testdata.zig");
fn checkValue(expected: f64, actual: f64) !void {
    if (std.math.isNan(expected)) {
        try std.testing.expect(std.math.isNan(actual));
    } else {
        try std.testing.expect(@abs(actual - expected) <= @max(1e-10, @abs(expected) * 5e-15));
    }
}
test "GoertzelCycleSpectrum prepared case 0" {
    var ind = try GoertzelCycleSpectrum.init(std.testing.allocator, .{ .window = 128, .min_period = 5, .max_period = 50, .step = 1, .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A5();
    const e1 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH5();
    const e2 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO5();
    const e3 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A6();
    const e4 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH6();
    const e5 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO6();
    const e6 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A7();
    const e7 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH7();
    const e8 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO7();
    const e9 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A8();
    const e10 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH8();
    const e11 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO8();
    const e12 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A9();
    const e13 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH9();
    const e14 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO9();
    const e15 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A10();
    const e16 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH10();
    const e17 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO10();
    const e18 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A11();
    const e19 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH11();
    const e20 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO11();
    const e21 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A12();
    const e22 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH12();
    const e23 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO12();
    const e24 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A13();
    const e25 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH13();
    const e26 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO13();
    const e27 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A14();
    const e28 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH14();
    const e29 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO14();
    const e30 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A15();
    const e31 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH15();
    const e32 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO15();
    const e33 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A16();
    const e34 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH16();
    const e35 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO16();
    const e36 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A17();
    const e37 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH17();
    const e38 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO17();
    const e39 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A18();
    const e40 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH18();
    const e41 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO18();
    const e42 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A19();
    const e43 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH19();
    const e44 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO19();
    const e45 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A20();
    const e46 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH20();
    const e47 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO20();
    const e48 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A21();
    const e49 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH21();
    const e50 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO21();
    const e51 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A22();
    const e52 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH22();
    const e53 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO22();
    const e54 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A23();
    const e55 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH23();
    const e56 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO23();
    const e57 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A24();
    const e58 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH24();
    const e59 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO24();
    const e60 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A25();
    const e61 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH25();
    const e62 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO25();
    const e63 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A26();
    const e64 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH26();
    const e65 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO26();
    const e66 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A27();
    const e67 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH27();
    const e68 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO27();
    const e69 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A28();
    const e70 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH28();
    const e71 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO28();
    const e72 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A29();
    const e73 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH29();
    const e74 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO29();
    const e75 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A30();
    const e76 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH30();
    const e77 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO30();
    const e78 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A31();
    const e79 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH31();
    const e80 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO31();
    const e81 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A32();
    const e82 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH32();
    const e83 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO32();
    const e84 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A33();
    const e85 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH33();
    const e86 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO33();
    const e87 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A34();
    const e88 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH34();
    const e89 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO34();
    const e90 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A35();
    const e91 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH35();
    const e92 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO35();
    const e93 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A36();
    const e94 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH36();
    const e95 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO36();
    const e96 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A37();
    const e97 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH37();
    const e98 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO37();
    const e99 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A38();
    const e100 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH38();
    const e101 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO38();
    const e102 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A39();
    const e103 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH39();
    const e104 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO39();
    const e105 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A40();
    const e106 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH40();
    const e107 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO40();
    const e108 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A41();
    const e109 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH41();
    const e110 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO41();
    const e111 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A42();
    const e112 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH42();
    const e113 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO42();
    const e114 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A43();
    const e115 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH43();
    const e116 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO43();
    const e117 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A44();
    const e118 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH44();
    const e119 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO44();
    const e120 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A45();
    const e121 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH45();
    const e122 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO45();
    const e123 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A46();
    const e124 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH46();
    const e125 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO46();
    const e126 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A47();
    const e127 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH47();
    const e128 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO47();
    const e129 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A48();
    const e130 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH48();
    const e131 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO48();
    const e132 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A49();
    const e133 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH49();
    const e134 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO49();
    const e135 = testdata.expectedW128_MIN5_MAX50_S1_L1600_A50();
    const e136 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PH50();
    const e137 = testdata.expectedW128_MIN5_MAX50_S1_L1600_TO50();
    const e138 = testdata.expectedW128_MIN5_MAX50_S1_L1600_DOM();
    const e139 = testdata.expectedW128_MIN5_MAX50_S1_L1600_AMP();
    const e140 = testdata.expectedW128_MIN5_MAX50_S1_L1600_PHASE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], if (result.amplitude.len == 0) nan else result.amplitude[0]);
        try checkValue(e1[index], if (result.phases.len == 0) nan else result.phases[0]);
        try checkValue(e2[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[0]);
        try checkValue(e3[index], if (result.amplitude.len == 0) nan else result.amplitude[1]);
        try checkValue(e4[index], if (result.phases.len == 0) nan else result.phases[1]);
        try checkValue(e5[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[1]);
        try checkValue(e6[index], if (result.amplitude.len == 0) nan else result.amplitude[2]);
        try checkValue(e7[index], if (result.phases.len == 0) nan else result.phases[2]);
        try checkValue(e8[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[2]);
        try checkValue(e9[index], if (result.amplitude.len == 0) nan else result.amplitude[3]);
        try checkValue(e10[index], if (result.phases.len == 0) nan else result.phases[3]);
        try checkValue(e11[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[3]);
        try checkValue(e12[index], if (result.amplitude.len == 0) nan else result.amplitude[4]);
        try checkValue(e13[index], if (result.phases.len == 0) nan else result.phases[4]);
        try checkValue(e14[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[4]);
        try checkValue(e15[index], if (result.amplitude.len == 0) nan else result.amplitude[5]);
        try checkValue(e16[index], if (result.phases.len == 0) nan else result.phases[5]);
        try checkValue(e17[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[5]);
        try checkValue(e18[index], if (result.amplitude.len == 0) nan else result.amplitude[6]);
        try checkValue(e19[index], if (result.phases.len == 0) nan else result.phases[6]);
        try checkValue(e20[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[6]);
        try checkValue(e21[index], if (result.amplitude.len == 0) nan else result.amplitude[7]);
        try checkValue(e22[index], if (result.phases.len == 0) nan else result.phases[7]);
        try checkValue(e23[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[7]);
        try checkValue(e24[index], if (result.amplitude.len == 0) nan else result.amplitude[8]);
        try checkValue(e25[index], if (result.phases.len == 0) nan else result.phases[8]);
        try checkValue(e26[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[8]);
        try checkValue(e27[index], if (result.amplitude.len == 0) nan else result.amplitude[9]);
        try checkValue(e28[index], if (result.phases.len == 0) nan else result.phases[9]);
        try checkValue(e29[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[9]);
        try checkValue(e30[index], if (result.amplitude.len == 0) nan else result.amplitude[10]);
        try checkValue(e31[index], if (result.phases.len == 0) nan else result.phases[10]);
        try checkValue(e32[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[10]);
        try checkValue(e33[index], if (result.amplitude.len == 0) nan else result.amplitude[11]);
        try checkValue(e34[index], if (result.phases.len == 0) nan else result.phases[11]);
        try checkValue(e35[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[11]);
        try checkValue(e36[index], if (result.amplitude.len == 0) nan else result.amplitude[12]);
        try checkValue(e37[index], if (result.phases.len == 0) nan else result.phases[12]);
        try checkValue(e38[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[12]);
        try checkValue(e39[index], if (result.amplitude.len == 0) nan else result.amplitude[13]);
        try checkValue(e40[index], if (result.phases.len == 0) nan else result.phases[13]);
        try checkValue(e41[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[13]);
        try checkValue(e42[index], if (result.amplitude.len == 0) nan else result.amplitude[14]);
        try checkValue(e43[index], if (result.phases.len == 0) nan else result.phases[14]);
        try checkValue(e44[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[14]);
        try checkValue(e45[index], if (result.amplitude.len == 0) nan else result.amplitude[15]);
        try checkValue(e46[index], if (result.phases.len == 0) nan else result.phases[15]);
        try checkValue(e47[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[15]);
        try checkValue(e48[index], if (result.amplitude.len == 0) nan else result.amplitude[16]);
        try checkValue(e49[index], if (result.phases.len == 0) nan else result.phases[16]);
        try checkValue(e50[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[16]);
        try checkValue(e51[index], if (result.amplitude.len == 0) nan else result.amplitude[17]);
        try checkValue(e52[index], if (result.phases.len == 0) nan else result.phases[17]);
        try checkValue(e53[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[17]);
        try checkValue(e54[index], if (result.amplitude.len == 0) nan else result.amplitude[18]);
        try checkValue(e55[index], if (result.phases.len == 0) nan else result.phases[18]);
        try checkValue(e56[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[18]);
        try checkValue(e57[index], if (result.amplitude.len == 0) nan else result.amplitude[19]);
        try checkValue(e58[index], if (result.phases.len == 0) nan else result.phases[19]);
        try checkValue(e59[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[19]);
        try checkValue(e60[index], if (result.amplitude.len == 0) nan else result.amplitude[20]);
        try checkValue(e61[index], if (result.phases.len == 0) nan else result.phases[20]);
        try checkValue(e62[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[20]);
        try checkValue(e63[index], if (result.amplitude.len == 0) nan else result.amplitude[21]);
        try checkValue(e64[index], if (result.phases.len == 0) nan else result.phases[21]);
        try checkValue(e65[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[21]);
        try checkValue(e66[index], if (result.amplitude.len == 0) nan else result.amplitude[22]);
        try checkValue(e67[index], if (result.phases.len == 0) nan else result.phases[22]);
        try checkValue(e68[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[22]);
        try checkValue(e69[index], if (result.amplitude.len == 0) nan else result.amplitude[23]);
        try checkValue(e70[index], if (result.phases.len == 0) nan else result.phases[23]);
        try checkValue(e71[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[23]);
        try checkValue(e72[index], if (result.amplitude.len == 0) nan else result.amplitude[24]);
        try checkValue(e73[index], if (result.phases.len == 0) nan else result.phases[24]);
        try checkValue(e74[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[24]);
        try checkValue(e75[index], if (result.amplitude.len == 0) nan else result.amplitude[25]);
        try checkValue(e76[index], if (result.phases.len == 0) nan else result.phases[25]);
        try checkValue(e77[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[25]);
        try checkValue(e78[index], if (result.amplitude.len == 0) nan else result.amplitude[26]);
        try checkValue(e79[index], if (result.phases.len == 0) nan else result.phases[26]);
        try checkValue(e80[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[26]);
        try checkValue(e81[index], if (result.amplitude.len == 0) nan else result.amplitude[27]);
        try checkValue(e82[index], if (result.phases.len == 0) nan else result.phases[27]);
        try checkValue(e83[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[27]);
        try checkValue(e84[index], if (result.amplitude.len == 0) nan else result.amplitude[28]);
        try checkValue(e85[index], if (result.phases.len == 0) nan else result.phases[28]);
        try checkValue(e86[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[28]);
        try checkValue(e87[index], if (result.amplitude.len == 0) nan else result.amplitude[29]);
        try checkValue(e88[index], if (result.phases.len == 0) nan else result.phases[29]);
        try checkValue(e89[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[29]);
        try checkValue(e90[index], if (result.amplitude.len == 0) nan else result.amplitude[30]);
        try checkValue(e91[index], if (result.phases.len == 0) nan else result.phases[30]);
        try checkValue(e92[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[30]);
        try checkValue(e93[index], if (result.amplitude.len == 0) nan else result.amplitude[31]);
        try checkValue(e94[index], if (result.phases.len == 0) nan else result.phases[31]);
        try checkValue(e95[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[31]);
        try checkValue(e96[index], if (result.amplitude.len == 0) nan else result.amplitude[32]);
        try checkValue(e97[index], if (result.phases.len == 0) nan else result.phases[32]);
        try checkValue(e98[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[32]);
        try checkValue(e99[index], if (result.amplitude.len == 0) nan else result.amplitude[33]);
        try checkValue(e100[index], if (result.phases.len == 0) nan else result.phases[33]);
        try checkValue(e101[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[33]);
        try checkValue(e102[index], if (result.amplitude.len == 0) nan else result.amplitude[34]);
        try checkValue(e103[index], if (result.phases.len == 0) nan else result.phases[34]);
        try checkValue(e104[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[34]);
        try checkValue(e105[index], if (result.amplitude.len == 0) nan else result.amplitude[35]);
        try checkValue(e106[index], if (result.phases.len == 0) nan else result.phases[35]);
        try checkValue(e107[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[35]);
        try checkValue(e108[index], if (result.amplitude.len == 0) nan else result.amplitude[36]);
        try checkValue(e109[index], if (result.phases.len == 0) nan else result.phases[36]);
        try checkValue(e110[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[36]);
        try checkValue(e111[index], if (result.amplitude.len == 0) nan else result.amplitude[37]);
        try checkValue(e112[index], if (result.phases.len == 0) nan else result.phases[37]);
        try checkValue(e113[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[37]);
        try checkValue(e114[index], if (result.amplitude.len == 0) nan else result.amplitude[38]);
        try checkValue(e115[index], if (result.phases.len == 0) nan else result.phases[38]);
        try checkValue(e116[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[38]);
        try checkValue(e117[index], if (result.amplitude.len == 0) nan else result.amplitude[39]);
        try checkValue(e118[index], if (result.phases.len == 0) nan else result.phases[39]);
        try checkValue(e119[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[39]);
        try checkValue(e120[index], if (result.amplitude.len == 0) nan else result.amplitude[40]);
        try checkValue(e121[index], if (result.phases.len == 0) nan else result.phases[40]);
        try checkValue(e122[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[40]);
        try checkValue(e123[index], if (result.amplitude.len == 0) nan else result.amplitude[41]);
        try checkValue(e124[index], if (result.phases.len == 0) nan else result.phases[41]);
        try checkValue(e125[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[41]);
        try checkValue(e126[index], if (result.amplitude.len == 0) nan else result.amplitude[42]);
        try checkValue(e127[index], if (result.phases.len == 0) nan else result.phases[42]);
        try checkValue(e128[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[42]);
        try checkValue(e129[index], if (result.amplitude.len == 0) nan else result.amplitude[43]);
        try checkValue(e130[index], if (result.phases.len == 0) nan else result.phases[43]);
        try checkValue(e131[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[43]);
        try checkValue(e132[index], if (result.amplitude.len == 0) nan else result.amplitude[44]);
        try checkValue(e133[index], if (result.phases.len == 0) nan else result.phases[44]);
        try checkValue(e134[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[44]);
        try checkValue(e135[index], if (result.amplitude.len == 0) nan else result.amplitude[45]);
        try checkValue(e136[index], if (result.phases.len == 0) nan else result.phases[45]);
        try checkValue(e137[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[45]);
        try checkValue(e138[index], result.dominant_period);
        try checkValue(e139[index], result.dominant_amplitude);
        try checkValue(e140[index], result.dominant_phase);
    }
}
test "GoertzelCycleSpectrum prepared case 1" {
    var ind = try GoertzelCycleSpectrum.init(std.testing.allocator, .{ .window = 100, .min_period = 20, .max_period = 20, .step = 1, .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW100_MIN20_MAX20_S1_L0_A20();
    const e1 = testdata.expectedW100_MIN20_MAX20_S1_L0_PH20();
    const e2 = testdata.expectedW100_MIN20_MAX20_S1_L0_TO20();
    const e3 = testdata.expectedW100_MIN20_MAX20_S1_L0_DOM();
    const e4 = testdata.expectedW100_MIN20_MAX20_S1_L0_AMP();
    const e5 = testdata.expectedW100_MIN20_MAX20_S1_L0_PHASE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], if (result.amplitude.len == 0) nan else result.amplitude[0]);
        try checkValue(e1[index], if (result.phases.len == 0) nan else result.phases[0]);
        try checkValue(e2[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[0]);
        try checkValue(e3[index], result.dominant_period);
        try checkValue(e4[index], result.dominant_amplitude);
        try checkValue(e5[index], result.dominant_phase);
    }
}
test "GoertzelCycleSpectrum prepared case 2" {
    var ind = try GoertzelCycleSpectrum.init(std.testing.allocator, .{ .window = 40, .min_period = 5, .max_period = 25, .step = 5, .smoothing = 0.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW40_MIN5_MAX25_S5_L0_A5();
    const e1 = testdata.expectedW40_MIN5_MAX25_S5_L0_PH5();
    const e2 = testdata.expectedW40_MIN5_MAX25_S5_L0_TO5();
    const e3 = testdata.expectedW40_MIN5_MAX25_S5_L0_A10();
    const e4 = testdata.expectedW40_MIN5_MAX25_S5_L0_PH10();
    const e5 = testdata.expectedW40_MIN5_MAX25_S5_L0_TO10();
    const e6 = testdata.expectedW40_MIN5_MAX25_S5_L0_A15();
    const e7 = testdata.expectedW40_MIN5_MAX25_S5_L0_PH15();
    const e8 = testdata.expectedW40_MIN5_MAX25_S5_L0_TO15();
    const e9 = testdata.expectedW40_MIN5_MAX25_S5_L0_A20();
    const e10 = testdata.expectedW40_MIN5_MAX25_S5_L0_PH20();
    const e11 = testdata.expectedW40_MIN5_MAX25_S5_L0_TO20();
    const e12 = testdata.expectedW40_MIN5_MAX25_S5_L0_A25();
    const e13 = testdata.expectedW40_MIN5_MAX25_S5_L0_PH25();
    const e14 = testdata.expectedW40_MIN5_MAX25_S5_L0_TO25();
    const e15 = testdata.expectedW40_MIN5_MAX25_S5_L0_DOM();
    const e16 = testdata.expectedW40_MIN5_MAX25_S5_L0_AMP();
    const e17 = testdata.expectedW40_MIN5_MAX25_S5_L0_PHASE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], if (result.amplitude.len == 0) nan else result.amplitude[0]);
        try checkValue(e1[index], if (result.phases.len == 0) nan else result.phases[0]);
        try checkValue(e2[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[0]);
        try checkValue(e3[index], if (result.amplitude.len == 0) nan else result.amplitude[1]);
        try checkValue(e4[index], if (result.phases.len == 0) nan else result.phases[1]);
        try checkValue(e5[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[1]);
        try checkValue(e6[index], if (result.amplitude.len == 0) nan else result.amplitude[2]);
        try checkValue(e7[index], if (result.phases.len == 0) nan else result.phases[2]);
        try checkValue(e8[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[2]);
        try checkValue(e9[index], if (result.amplitude.len == 0) nan else result.amplitude[3]);
        try checkValue(e10[index], if (result.phases.len == 0) nan else result.phases[3]);
        try checkValue(e11[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[3]);
        try checkValue(e12[index], if (result.amplitude.len == 0) nan else result.amplitude[4]);
        try checkValue(e13[index], if (result.phases.len == 0) nan else result.phases[4]);
        try checkValue(e14[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[4]);
        try checkValue(e15[index], result.dominant_period);
        try checkValue(e16[index], result.dominant_amplitude);
        try checkValue(e17[index], result.dominant_phase);
    }
}
test "GoertzelCycleSpectrum prepared case 3" {
    var ind = try GoertzelCycleSpectrum.init(std.testing.allocator, .{ .window = 40, .min_period = 5, .max_period = 25, .step = 5, .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW40_MIN5_MAX25_S5_L1600_A5();
    const e1 = testdata.expectedW40_MIN5_MAX25_S5_L1600_PH5();
    const e2 = testdata.expectedW40_MIN5_MAX25_S5_L1600_TO5();
    const e3 = testdata.expectedW40_MIN5_MAX25_S5_L1600_A10();
    const e4 = testdata.expectedW40_MIN5_MAX25_S5_L1600_PH10();
    const e5 = testdata.expectedW40_MIN5_MAX25_S5_L1600_TO10();
    const e6 = testdata.expectedW40_MIN5_MAX25_S5_L1600_A15();
    const e7 = testdata.expectedW40_MIN5_MAX25_S5_L1600_PH15();
    const e8 = testdata.expectedW40_MIN5_MAX25_S5_L1600_TO15();
    const e9 = testdata.expectedW40_MIN5_MAX25_S5_L1600_A20();
    const e10 = testdata.expectedW40_MIN5_MAX25_S5_L1600_PH20();
    const e11 = testdata.expectedW40_MIN5_MAX25_S5_L1600_TO20();
    const e12 = testdata.expectedW40_MIN5_MAX25_S5_L1600_A25();
    const e13 = testdata.expectedW40_MIN5_MAX25_S5_L1600_PH25();
    const e14 = testdata.expectedW40_MIN5_MAX25_S5_L1600_TO25();
    const e15 = testdata.expectedW40_MIN5_MAX25_S5_L1600_DOM();
    const e16 = testdata.expectedW40_MIN5_MAX25_S5_L1600_AMP();
    const e17 = testdata.expectedW40_MIN5_MAX25_S5_L1600_PHASE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], if (result.amplitude.len == 0) nan else result.amplitude[0]);
        try checkValue(e1[index], if (result.phases.len == 0) nan else result.phases[0]);
        try checkValue(e2[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[0]);
        try checkValue(e3[index], if (result.amplitude.len == 0) nan else result.amplitude[1]);
        try checkValue(e4[index], if (result.phases.len == 0) nan else result.phases[1]);
        try checkValue(e5[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[1]);
        try checkValue(e6[index], if (result.amplitude.len == 0) nan else result.amplitude[2]);
        try checkValue(e7[index], if (result.phases.len == 0) nan else result.phases[2]);
        try checkValue(e8[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[2]);
        try checkValue(e9[index], if (result.amplitude.len == 0) nan else result.amplitude[3]);
        try checkValue(e10[index], if (result.phases.len == 0) nan else result.phases[3]);
        try checkValue(e11[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[3]);
        try checkValue(e12[index], if (result.amplitude.len == 0) nan else result.amplitude[4]);
        try checkValue(e13[index], if (result.phases.len == 0) nan else result.phases[4]);
        try checkValue(e14[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[4]);
        try checkValue(e15[index], result.dominant_period);
        try checkValue(e16[index], result.dominant_amplitude);
        try checkValue(e17[index], result.dominant_phase);
    }
}
test "GoertzelCycleSpectrum prepared case 4" {
    var ind = try GoertzelCycleSpectrum.init(std.testing.allocator, .{ .window = 200, .min_period = 10, .max_period = 50, .step = 10, .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW200_MIN10_MAX50_S10_L1600_A10();
    const e1 = testdata.expectedW200_MIN10_MAX50_S10_L1600_PH10();
    const e2 = testdata.expectedW200_MIN10_MAX50_S10_L1600_TO10();
    const e3 = testdata.expectedW200_MIN10_MAX50_S10_L1600_A20();
    const e4 = testdata.expectedW200_MIN10_MAX50_S10_L1600_PH20();
    const e5 = testdata.expectedW200_MIN10_MAX50_S10_L1600_TO20();
    const e6 = testdata.expectedW200_MIN10_MAX50_S10_L1600_A30();
    const e7 = testdata.expectedW200_MIN10_MAX50_S10_L1600_PH30();
    const e8 = testdata.expectedW200_MIN10_MAX50_S10_L1600_TO30();
    const e9 = testdata.expectedW200_MIN10_MAX50_S10_L1600_A40();
    const e10 = testdata.expectedW200_MIN10_MAX50_S10_L1600_PH40();
    const e11 = testdata.expectedW200_MIN10_MAX50_S10_L1600_TO40();
    const e12 = testdata.expectedW200_MIN10_MAX50_S10_L1600_A50();
    const e13 = testdata.expectedW200_MIN10_MAX50_S10_L1600_PH50();
    const e14 = testdata.expectedW200_MIN10_MAX50_S10_L1600_TO50();
    const e15 = testdata.expectedW200_MIN10_MAX50_S10_L1600_DOM();
    const e16 = testdata.expectedW200_MIN10_MAX50_S10_L1600_AMP();
    const e17 = testdata.expectedW200_MIN10_MAX50_S10_L1600_PHASE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], if (result.amplitude.len == 0) nan else result.amplitude[0]);
        try checkValue(e1[index], if (result.phases.len == 0) nan else result.phases[0]);
        try checkValue(e2[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[0]);
        try checkValue(e3[index], if (result.amplitude.len == 0) nan else result.amplitude[1]);
        try checkValue(e4[index], if (result.phases.len == 0) nan else result.phases[1]);
        try checkValue(e5[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[1]);
        try checkValue(e6[index], if (result.amplitude.len == 0) nan else result.amplitude[2]);
        try checkValue(e7[index], if (result.phases.len == 0) nan else result.phases[2]);
        try checkValue(e8[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[2]);
        try checkValue(e9[index], if (result.amplitude.len == 0) nan else result.amplitude[3]);
        try checkValue(e10[index], if (result.phases.len == 0) nan else result.phases[3]);
        try checkValue(e11[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[3]);
        try checkValue(e12[index], if (result.amplitude.len == 0) nan else result.amplitude[4]);
        try checkValue(e13[index], if (result.phases.len == 0) nan else result.phases[4]);
        try checkValue(e14[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[4]);
        try checkValue(e15[index], result.dominant_period);
        try checkValue(e16[index], result.dominant_amplitude);
        try checkValue(e17[index], result.dominant_phase);
    }
}
test "GoertzelCycleSpectrum prepared case 5" {
    var ind = try GoertzelCycleSpectrum.init(std.testing.allocator, .{ .window = 20, .min_period = 10, .max_period = 50, .step = 10, .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW20_MIN10_MAX50_S10_L1600_A10();
    const e1 = testdata.expectedW20_MIN10_MAX50_S10_L1600_PH10();
    const e2 = testdata.expectedW20_MIN10_MAX50_S10_L1600_TO10();
    const e3 = testdata.expectedW20_MIN10_MAX50_S10_L1600_A20();
    const e4 = testdata.expectedW20_MIN10_MAX50_S10_L1600_PH20();
    const e5 = testdata.expectedW20_MIN10_MAX50_S10_L1600_TO20();
    const e6 = testdata.expectedW20_MIN10_MAX50_S10_L1600_A30();
    const e7 = testdata.expectedW20_MIN10_MAX50_S10_L1600_PH30();
    const e8 = testdata.expectedW20_MIN10_MAX50_S10_L1600_TO30();
    const e9 = testdata.expectedW20_MIN10_MAX50_S10_L1600_A40();
    const e10 = testdata.expectedW20_MIN10_MAX50_S10_L1600_PH40();
    const e11 = testdata.expectedW20_MIN10_MAX50_S10_L1600_TO40();
    const e12 = testdata.expectedW20_MIN10_MAX50_S10_L1600_A50();
    const e13 = testdata.expectedW20_MIN10_MAX50_S10_L1600_PH50();
    const e14 = testdata.expectedW20_MIN10_MAX50_S10_L1600_TO50();
    const e15 = testdata.expectedW20_MIN10_MAX50_S10_L1600_DOM();
    const e16 = testdata.expectedW20_MIN10_MAX50_S10_L1600_AMP();
    const e17 = testdata.expectedW20_MIN10_MAX50_S10_L1600_PHASE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], if (result.amplitude.len == 0) nan else result.amplitude[0]);
        try checkValue(e1[index], if (result.phases.len == 0) nan else result.phases[0]);
        try checkValue(e2[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[0]);
        try checkValue(e3[index], if (result.amplitude.len == 0) nan else result.amplitude[1]);
        try checkValue(e4[index], if (result.phases.len == 0) nan else result.phases[1]);
        try checkValue(e5[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[1]);
        try checkValue(e6[index], if (result.amplitude.len == 0) nan else result.amplitude[2]);
        try checkValue(e7[index], if (result.phases.len == 0) nan else result.phases[2]);
        try checkValue(e8[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[2]);
        try checkValue(e9[index], if (result.amplitude.len == 0) nan else result.amplitude[3]);
        try checkValue(e10[index], if (result.phases.len == 0) nan else result.phases[3]);
        try checkValue(e11[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[3]);
        try checkValue(e12[index], if (result.amplitude.len == 0) nan else result.amplitude[4]);
        try checkValue(e13[index], if (result.phases.len == 0) nan else result.phases[4]);
        try checkValue(e14[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[4]);
        try checkValue(e15[index], result.dominant_period);
        try checkValue(e16[index], result.dominant_amplitude);
        try checkValue(e17[index], result.dominant_phase);
    }
}
test "GoertzelCycleSpectrum prepared case 6" {
    var ind = try GoertzelCycleSpectrum.init(std.testing.allocator, .{ .window = 252, .min_period = 8, .max_period = 32, .step = 8, .smoothing = 160.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW252_MIN8_MAX32_S8_L160_A8();
    const e1 = testdata.expectedW252_MIN8_MAX32_S8_L160_PH8();
    const e2 = testdata.expectedW252_MIN8_MAX32_S8_L160_TO8();
    const e3 = testdata.expectedW252_MIN8_MAX32_S8_L160_A16();
    const e4 = testdata.expectedW252_MIN8_MAX32_S8_L160_PH16();
    const e5 = testdata.expectedW252_MIN8_MAX32_S8_L160_TO16();
    const e6 = testdata.expectedW252_MIN8_MAX32_S8_L160_A24();
    const e7 = testdata.expectedW252_MIN8_MAX32_S8_L160_PH24();
    const e8 = testdata.expectedW252_MIN8_MAX32_S8_L160_TO24();
    const e9 = testdata.expectedW252_MIN8_MAX32_S8_L160_A32();
    const e10 = testdata.expectedW252_MIN8_MAX32_S8_L160_PH32();
    const e11 = testdata.expectedW252_MIN8_MAX32_S8_L160_TO32();
    const e12 = testdata.expectedW252_MIN8_MAX32_S8_L160_DOM();
    const e13 = testdata.expectedW252_MIN8_MAX32_S8_L160_AMP();
    const e14 = testdata.expectedW252_MIN8_MAX32_S8_L160_PHASE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], if (result.amplitude.len == 0) nan else result.amplitude[0]);
        try checkValue(e1[index], if (result.phases.len == 0) nan else result.phases[0]);
        try checkValue(e2[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[0]);
        try checkValue(e3[index], if (result.amplitude.len == 0) nan else result.amplitude[1]);
        try checkValue(e4[index], if (result.phases.len == 0) nan else result.phases[1]);
        try checkValue(e5[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[1]);
        try checkValue(e6[index], if (result.amplitude.len == 0) nan else result.amplitude[2]);
        try checkValue(e7[index], if (result.phases.len == 0) nan else result.phases[2]);
        try checkValue(e8[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[2]);
        try checkValue(e9[index], if (result.amplitude.len == 0) nan else result.amplitude[3]);
        try checkValue(e10[index], if (result.phases.len == 0) nan else result.phases[3]);
        try checkValue(e11[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[3]);
        try checkValue(e12[index], result.dominant_period);
        try checkValue(e13[index], result.dominant_amplitude);
        try checkValue(e14[index], result.dominant_phase);
    }
}
test "GoertzelCycleSpectrum prepared case 7" {
    var ind = try GoertzelCycleSpectrum.init(std.testing.allocator, .{ .window = 253, .min_period = 10, .max_period = 30, .step = 10, .smoothing = 1600.0 });
    defer ind.deinit();
    const input = testdata.testInput();
    const e0 = testdata.expectedW253_MIN10_MAX30_S10_L1600_A10();
    const e1 = testdata.expectedW253_MIN10_MAX30_S10_L1600_PH10();
    const e2 = testdata.expectedW253_MIN10_MAX30_S10_L1600_TO10();
    const e3 = testdata.expectedW253_MIN10_MAX30_S10_L1600_A20();
    const e4 = testdata.expectedW253_MIN10_MAX30_S10_L1600_PH20();
    const e5 = testdata.expectedW253_MIN10_MAX30_S10_L1600_TO20();
    const e6 = testdata.expectedW253_MIN10_MAX30_S10_L1600_A30();
    const e7 = testdata.expectedW253_MIN10_MAX30_S10_L1600_PH30();
    const e8 = testdata.expectedW253_MIN10_MAX30_S10_L1600_TO30();
    const e9 = testdata.expectedW253_MIN10_MAX30_S10_L1600_DOM();
    const e10 = testdata.expectedW253_MIN10_MAX30_S10_L1600_AMP();
    const e11 = testdata.expectedW253_MIN10_MAX30_S10_L1600_PHASE();
    for (input, 0..) |price, index| {
        const result = try ind.update(price);
        try checkValue(e0[index], if (result.amplitude.len == 0) nan else result.amplitude[0]);
        try checkValue(e1[index], if (result.phases.len == 0) nan else result.phases[0]);
        try checkValue(e2[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[0]);
        try checkValue(e3[index], if (result.amplitude.len == 0) nan else result.amplitude[1]);
        try checkValue(e4[index], if (result.phases.len == 0) nan else result.phases[1]);
        try checkValue(e5[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[1]);
        try checkValue(e6[index], if (result.amplitude.len == 0) nan else result.amplitude[2]);
        try checkValue(e7[index], if (result.phases.len == 0) nan else result.phases[2]);
        try checkValue(e8[index], if (result.trough_offsets.len == 0) nan else result.trough_offsets[2]);
        try checkValue(e9[index], result.dominant_period);
        try checkValue(e10[index], result.dominant_amplitude);
        try checkValue(e11[index], result.dominant_phase);
    }
}
test "GoertzelCycleSpectrum lifecycle and metadata" {
    var ind = try GoertzelCycleSpectrum.init(std.testing.allocator, .{});
    defer ind.deinit();
    if (@hasDecl(GoertzelCycleSpectrum, "fixSlices")) ind.fixSlices();
    var meta: indicators.Metadata = undefined;
    ind.getMetadata(&meta);
    try std.testing.expectEqual(@as(usize, 6), meta.outputs_len);
    try std.testing.expect(!ind.isPrimed());
    try ind.reset();
    try std.testing.expect(!ind.isPrimed());
}

test "GoertzelCycleSpectrum component mnemonic suffixes" {
    const combinations = [_][3]?u8{ .{ null, null, null }, .{ 0, null, null }, .{ null, 0, null }, .{ null, null, 1 }, .{ 0, 0, null }, .{ 0, null, 1 }, .{ null, 0, 1 } };
    for (combinations) |c| {
        const bc: ?entities.BarComponent = if (c[0]) |v| @enumFromInt(v) else null;
        const qc: ?entities.QuoteComponent = if (c[1]) |v| @enumFromInt(v) else null;
        const tc: ?entities.TradeComponent = if (c[2]) |v| @enumFromInt(v) else null;
        var ind = try GoertzelCycleSpectrum.init(std.testing.allocator, .{ .bar_component = bc, .quote_component = qc, .trade_component = tc });
        defer ind.deinit();
        var metadata: indicators.Metadata = undefined;
        ind.getMetadata(&metadata);
        var buffer: [128]u8 = undefined;
        const suffix = indicators.componentTripleMnemonic(&buffer, bc orelse entities.bar_component.default_bar_component, qc orelse entities.quote_component.default_quote_component, tc orelse entities.trade_component.default_trade_component);
        try std.testing.expect(std.mem.endsWith(u8, metadata.mnemonic[0 .. metadata.mnemonic.len - 1], suffix));
    }
}
