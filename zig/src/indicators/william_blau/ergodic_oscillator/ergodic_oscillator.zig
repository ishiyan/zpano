const std = @import("std");
const math = std.math;

const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const bar_component = entities.bar_component;
const quote_component = entities.quote_component;
const trade_component = entities.trade_component;
const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");
const tsi_mod = @import("../true_strength_index/true_strength_index.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const TrueStrengthIndex = tsi_mod.TrueStrengthIndex;

/// Enumerates the outputs of the Ergodic Oscillator indicator.
pub const ErgodicOscillatorOutput = enum(u8) {
    /// The Ergodic oscillator value (the True Strength Index, range [-100, +100]).
    ergodic = 1,
    /// Signal-line value: the ul-period EMA of the oscillator.
    signal = 2,
};

/// Parameters to create an Ergodic Oscillator indicator.
///
/// The field names q, r, s, u and ul are the canonical symbols from William
/// Blau's Momentum, Direction, and Divergence (Wiley, 1995), chapter 2.
pub const ErgodicOscillatorParams = struct {
    q: usize = 2,
    r: usize = 20,
    s: usize = 5,
    u: usize = 3,
    ul: usize = 3,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ergodic Oscillator by William Blau.
///
/// The Ergodic is the True Strength Index plotted together with a signal line --
/// the EMA of the oscillator that Blau introduces as the trading vehicle for the
/// TSI (ch. 2, Fig. 2-14):
///
///   ergodic_k = TSI(q, r, s, u)_k                    (the oscillator)
///   signal_k  = EMA(ergodic, ul)_k                   (ul-period EMA of it)
///
/// The indicator produces two outputs:
///   - Ergodic: the oscillator, range [-100, +100], NaN during warm-up (bars 0..q-2);
///   - Signal: the ul-period EMA of the oscillator.
///
/// The numerics are exactly those of the True Strength Index, so this indicator
/// wraps a TrueStrengthIndex instance instead of duplicating the triple EMA
/// cascade. Only the mnemonic, the identifier, and the output naming differ.
pub const ErgodicOscillator = struct {
    tsi: TrueStrengthIndex,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: ErgodicOscillatorParams) !ErgodicOscillator {
        const q = params.q;
        const r = params.r;
        const s = params.s;
        const u = params.u;
        const ul = params.ul;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // The oscillator and its signal line are exactly the True Strength Index
        // outputs; wrap an instance rather than duplicating its numerics. The
        // resolved components are passed through so both agree on the mnemonic.
        // Parameter validation is performed by the wrapped indicator.
        var tsi = try TrueStrengthIndex.init(allocator, .{
            .q = q,
            .r = r,
            .s = s,
            .u = u,
            .ul = ul,
            .bar_component = bc,
            .quote_component = qc,
            .trade_component = tc,
        });
        errdefer tsi.deinit();

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "ergodic({d},{d},{d},{d},{d}{s})", .{
            q,
            r,
            s,
            u,
            ul,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Ergodic Oscillator {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .tsi = tsi,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *ErgodicOscillator) void {
        self.tsi.deinit();
    }

    pub fn fixSlices(self: *ErgodicOscillator) void {
        self.tsi.fixSlices();
        // The Ergodic doesn't use LineIndicator; mnemonic/description are read
        // from owned buffers directly, so no further slice fixup is needed.
    }

    /// Returns ergodic, signal.
    pub fn updateValues(self: *ErgodicOscillator, sample: f64) struct { ergodic: f64, signal: f64 } {
        const result = self.tsi.updateValues(sample);
        return .{ .ergodic = result.tsi, .signal = result.signal };
    }

    pub fn isPrimed(self: *const ErgodicOscillator) bool {
        return self.tsi.isPrimed();
    }

    fn mnemonic(self: *const ErgodicOscillator) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const ErgodicOscillator) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const ErgodicOscillator, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var ergodic_mn_buf: [160]u8 = undefined;
        const ergodic_mn = std.fmt.bufPrint(&ergodic_mn_buf, "{s} ergodic", .{mn}) catch mn;
        var signal_mn_buf: [160]u8 = undefined;
        const signal_mn = std.fmt.bufPrint(&signal_mn_buf, "{s} signal", .{mn}) catch mn;

        var ergodic_desc_buf: [256]u8 = undefined;
        const ergodic_desc = std.fmt.bufPrint(&ergodic_desc_buf, "{s} ergodic", .{desc}) catch desc;
        var signal_desc_buf: [256]u8 = undefined;
        const signal_desc = std.fmt.bufPrint(&signal_desc_buf, "{s} signal", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .ergodic_oscillator,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = ergodic_mn, .description = ergodic_desc },
                .{ .mnemonic = signal_mn, .description = signal_desc },
            },
        );
    }

    pub fn updateScalar(self: *ErgodicOscillator, sample: *const Scalar) OutputArray {
        const result = self.updateValues(sample.value);
        return makeOutput(sample.time, result.ergodic, result.signal);
    }

    pub fn updateBar(self: *ErgodicOscillator, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *ErgodicOscillator, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *ErgodicOscillator, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    fn makeOutput(time: i64, ergodic_v: f64, signal_v: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = ergodic_v } });
        out.append(.{ .scalar = .{ .time = time, .value = signal_v } });
        return out;
    }

    pub fn indicator(self: *ErgodicOscillator) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.VTable{
        .isPrimed = vtableIsPrimed,
        .metadata = vtableMetadata,
        .updateScalar = vtableUpdateScalar,
        .updateBar = vtableUpdateBar,
        .updateQuote = vtableUpdateQuote,
        .updateTrade = vtableUpdateTrade,
    };

    fn vtableIsPrimed(ptr: *anyopaque) bool {
        const self: *ErgodicOscillator = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const ErgodicOscillator = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *ErgodicOscillator = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *ErgodicOscillator = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *ErgodicOscillator = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *ErgodicOscillator = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidQ,
        InvalidR,
        InvalidS,
        InvalidU,
        InvalidUl,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

const Combo = struct {
    name: []const u8,
    q: usize,
    r: usize,
    s: usize,
    u: usize,
    ul: usize,
    ergodic: [252]f64,
    signal: [252]f64,
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }
    try testing.expect(@abs(act - exp) <= tolerance);
}

test "Ergodic reference data all combos" {
    const allocator = testing.allocator;
    const tolerance = 1e-9;
    const input = testdata.testInput();

    const combos = [_]Combo{
        .{ .name = "Q2_R20_S5_U3_L3", .q = 2, .r = 20, .s = 5, .u = 3, .ul = 3, .ergodic = testdata.expectedErgQ2_R20_S5_U3_L3(), .signal = testdata.expectedSigQ2_R20_S5_U3_L3() },
        .{ .name = "Q2_R32_S5_U1_L5", .q = 2, .r = 32, .s = 5, .u = 1, .ul = 5, .ergodic = testdata.expectedErgQ2_R32_S5_U1_L5(), .signal = testdata.expectedSigQ2_R32_S5_U1_L5() },
        .{ .name = "Q2_R20_S5_U1_L5", .q = 2, .r = 20, .s = 5, .u = 1, .ul = 5, .ergodic = testdata.expectedErgQ2_R20_S5_U1_L5(), .signal = testdata.expectedSigQ2_R20_S5_U1_L5() },
        .{ .name = "Q2_R32_S5_U1_L7", .q = 2, .r = 32, .s = 5, .u = 1, .ul = 7, .ergodic = testdata.expectedErgQ2_R32_S5_U1_L7(), .signal = testdata.expectedSigQ2_R32_S5_U1_L7() },
        .{ .name = "Q2_R25_S13_U1_L5", .q = 2, .r = 25, .s = 13, .u = 1, .ul = 5, .ergodic = testdata.expectedErgQ2_R25_S13_U1_L5(), .signal = testdata.expectedSigQ2_R25_S13_U1_L5() },
        .{ .name = "Q2_R20_S5_U3_L1", .q = 2, .r = 20, .s = 5, .u = 3, .ul = 1, .ergodic = testdata.expectedErgQ2_R20_S5_U3_L1(), .signal = testdata.expectedSigQ2_R20_S5_U3_L1() },
        .{ .name = "Q2_R1_S1_U1_L1", .q = 2, .r = 1, .s = 1, .u = 1, .ul = 1, .ergodic = testdata.expectedErgQ2_R1_S1_U1_L1(), .signal = testdata.expectedSigQ2_R1_S1_U1_L1() },
        .{ .name = "Q2_R20_S5_U3_L9", .q = 2, .r = 20, .s = 5, .u = 3, .ul = 9, .ergodic = testdata.expectedErgQ2_R20_S5_U3_L9(), .signal = testdata.expectedSigQ2_R20_S5_U3_L9() },
        .{ .name = "Q2_R64_S64_U1_L5", .q = 2, .r = 64, .s = 64, .u = 1, .ul = 5, .ergodic = testdata.expectedErgQ2_R64_S64_U1_L5(), .signal = testdata.expectedSigQ2_R64_S64_U1_L5() },
        .{ .name = "Q2_R9_S3_U1_L3", .q = 2, .r = 9, .s = 3, .u = 1, .ul = 3, .ergodic = testdata.expectedErgQ2_R9_S3_U1_L3(), .signal = testdata.expectedSigQ2_R9_S3_U1_L3() },
        .{ .name = "Q3_R20_S5_U3_L3", .q = 3, .r = 20, .s = 5, .u = 3, .ul = 3, .ergodic = testdata.expectedErgQ3_R20_S5_U3_L3(), .signal = testdata.expectedSigQ3_R20_S5_U3_L3() },
        .{ .name = "Q5_R20_S5_U3_L5", .q = 5, .r = 20, .s = 5, .u = 3, .ul = 5, .ergodic = testdata.expectedErgQ5_R20_S5_U3_L5(), .signal = testdata.expectedSigQ5_R20_S5_U3_L5() },
        .{ .name = "Q2_R13_S7_U1_L7", .q = 2, .r = 13, .s = 7, .u = 1, .ul = 7, .ergodic = testdata.expectedErgQ2_R13_S7_U1_L7(), .signal = testdata.expectedSigQ2_R13_S7_U1_L7() },
        .{ .name = "Q2_R20_S5_U1_L3", .q = 2, .r = 20, .s = 5, .u = 1, .ul = 3, .ergodic = testdata.expectedErgQ2_R20_S5_U1_L3(), .signal = testdata.expectedSigQ2_R20_S5_U1_L3() },
    };

    for (combos) |combo| {
        var ind = try ErgodicOscillator.init(allocator, .{
            .q = combo.q,
            .r = combo.r,
            .s = combo.s,
            .u = combo.u,
            .ul = combo.ul,
        });
        defer ind.deinit();

        for (0..252) |i| {
            const result = ind.updateValues(input[i]);
            try checkVal(combo.ergodic[i], result.ergodic, tolerance);
            try checkVal(combo.signal[i], result.signal, tolerance);
        }
    }
}

test "Ergodic signal passthrough when ul = 1" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    var ind = try ErgodicOscillator.init(allocator, .{ .q = 2, .r = 20, .s = 5, .u = 3, .ul = 1 });
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateValues(input[i]);
        if (math.isNan(result.ergodic)) {
            try testing.expect(math.isNan(result.signal));
        } else {
            try testing.expectEqual(result.ergodic, result.signal);
        }
    }
}

test "Ergodic is primed on the first finite oscillator value" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    var ind = try ErgodicOscillator.init(allocator, .{ .q = 3, .r = 20, .s = 5, .u = 3, .ul = 3 });
    defer ind.deinit();

    try testing.expect(!ind.isPrimed());
    _ = ind.updateValues(input[0]);
    try testing.expect(!ind.isPrimed());
    _ = ind.updateValues(input[1]);
    try testing.expect(!ind.isPrimed());
    _ = ind.updateValues(input[2]);
    try testing.expect(ind.isPrimed());
}

test "Ergodic metadata default" {
    const allocator = testing.allocator;

    var ind = try ErgodicOscillator.init(allocator, .{});
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.ergodic_oscillator, meta.identifier);
    try testing.expectEqualStrings("ergodic(2,20,5,3,3)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
}

test "Ergodic custom mnemonic" {
    const allocator = testing.allocator;

    var ind = try ErgodicOscillator.init(allocator, .{ .q = 2, .r = 25, .s = 13, .u = 1, .ul = 7 });
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("ergodic(2,25,13,1,7)", meta.mnemonic);
}

test "Ergodic invalid params" {
    const allocator = testing.allocator;

    const r1 = ErgodicOscillator.init(allocator, .{ .q = 0 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = ErgodicOscillator.init(allocator, .{ .r = 0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = ErgodicOscillator.init(allocator, .{ .s = 0 });
    try testing.expect(if (r3) |_| false else |_| true);

    const r4 = ErgodicOscillator.init(allocator, .{ .u = 0 });
    try testing.expect(if (r4) |_| false else |_| true);

    const r5 = ErgodicOscillator.init(allocator, .{ .ul = 0 });
    try testing.expect(if (r5) |_| false else |_| true);
}

test "Ergodic entity update ordering" {
    const allocator = testing.allocator;
    const tolerance = 1e-9;
    const input = testdata.testInput();
    const exp_ergodic = testdata.expectedErgQ2_R20_S5_U3_L3();
    const exp_signal = testdata.expectedSigQ2_R20_S5_U3_L3();

    var ind = try ErgodicOscillator.init(allocator, .{});
    defer ind.deinit();

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        last_out = ind.updateScalar(&scalar);
    }
    const items = last_out.slice();

    try checkVal(exp_ergodic[251], items[0].scalar.value, tolerance);
    try checkVal(exp_signal[251], items[1].scalar.value, tolerance);
}
