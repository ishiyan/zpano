const std = @import("std");
const math = std.math;

const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Directional Trend Index indicator.
pub const DirectionalTrendIndexOutput = enum(u8) {
    /// Directional Trend Index oscillator value (range [-100, +100]).
    dti = 1,
    /// Signal-line value: the ul-period EMA of the oscillator.
    signal = 2,
};

/// Parameters to create a Directional Trend Index indicator.
///
/// The field names q, r, s, u and ul are the canonical symbols from William Blau's
/// Momentum, Direction, and Divergence (Wiley, 1995).
///
/// The indicator consumes the high and low prices of a bar, so it has no
/// configurable price-component fields.
pub const DirectionalTrendIndexParams = struct {
    /// High-low momentum look-back period; the high and the low are compared with
    /// their values q-1 bars ago. Must be > 0 (q >= 2 is meaningful). Default 2.
    q: usize = 2,
    /// Period of the 1st (innermost) EMA in the smoothing cascade. Must be > 0. Default 20.
    r: usize = 20,
    /// Period of the 2nd EMA in the smoothing cascade. Must be > 0. Default 5.
    s: usize = 5,
    /// Period of the 3rd (outermost) EMA in the smoothing cascade; u = 1 switches it
    /// off (passthrough). Must be > 0. Default 3.
    u: usize = 3,
    /// Period of the signal-line EMA; ul = 1 makes the signal a passthrough.
    /// Must be > 0. Default 3.
    ul: usize = 3,
};

/// Stateful streaming EMA: alpha = 2/(period+1), seeds e0 = x0.
///
/// Inlined verbatim from the Blau exponential moving average so the indicator is a
/// standalone porting unit. Do NOT change its numerics.
///
/// period == 1 -> alpha == 1 -> pure passthrough (output == input).
const Ema = struct {
    alpha: f64,
    previous: f64 = 0.0,
    primed: bool = false,

    fn init(period: usize) Ema {
        return .{ .alpha = 2.0 / (@as(f64, @floatFromInt(period)) + 1.0) };
    }

    fn update(self: *Ema, x: f64) f64 {
        if (!self.primed) {
            self.previous = x;
            self.primed = true;
            return self.previous;
        }
        self.previous = self.alpha * x + (1.0 - self.alpha) * self.previous;
        return self.previous;
    }
};

/// Directional Trend Index (DTI) by William Blau.
///
/// A double-/triple-smoothed High-Low Momentum oscillator bounded to [-100, +100],
/// paired with an EMA signal line (the Ergodic form):
///
///   dti_k    = 100 * TEMA(HLM, r, s, u)_k / TEMA(|HLM|, r, s, u)_k   (the oscillator)
///   signal_k = EMA(dti, ul)_k                                       (ul-period EMA)
///
/// where the High-Low Momentum is built from how far the high rose and the low
/// fell relative to q-1 bars ago:
///
///   HMU_k = max(high_k - high_(k-(q-1)), 0)      (upward high movement)
///   LMD_k = max(low_(k-(q-1)) - low_k, 0)        (downward low movement)
///   HLM_k = HMU_k - LMD_k                         (composite high-low momentum)
///   TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u)  (triple EMA cascade)
///
/// This is the True Strength Index structure applied to HLM instead of price
/// momentum. The inputs are the high and low prices only (no close).
///
/// The indicator produces two outputs:
///   - DTI: the oscillator, range [-100, +100];
///   - Signal: the ul-period EMA of the oscillator.
///
/// Priming convention (book / EasyLanguage): HLM is valid from bar q-1 (it needs a
/// high/low from q-1 bars ago), so all cascade stages seed there together; both
/// outputs are NaN for bars 0..q-2 and finite from bar q-1. For q = 1 there is no
/// NaN warm-up, but HLM == 0 on every bar, so the division guard yields dti == 0.0
/// for all bars. The signal EMA seeds on the first finite oscillator value.
/// Division guard: denominator == 0 -> oscillator 0.0.
pub const DirectionalTrendIndex = struct {
    q: usize,
    highs: []f64,
    lows: []f64,
    window_count: usize,
    window_index: usize,

    num_r: Ema,
    num_s: Ema,
    num_u: Ema,
    den_r: Ema,
    den_s: Ema,
    den_u: Ema,

    signal_ema: Ema,

    primed: bool,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    allocator: std.mem.Allocator,

    pub fn init(allocator: std.mem.Allocator, params: DirectionalTrendIndexParams) InitError!DirectionalTrendIndex {
        const q = params.q;
        const r = params.r;
        const s = params.s;
        const u = params.u;
        const ul = params.ul;

        if (q < 1) return error.InvalidQ;
        if (r < 1) return error.InvalidR;
        if (s < 1) return error.InvalidS;
        if (u < 1) return error.InvalidU;
        if (ul < 1) return error.InvalidUl;

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "dti({d},{d},{d},{d},{d})", .{
            q,
            r,
            s,
            u,
            ul,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Directional Trend Index {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const highs = allocator.alloc(f64, q) catch return error.OutOfMemory;
        errdefer allocator.free(highs);
        const lows = allocator.alloc(f64, q) catch return error.OutOfMemory;

        return .{
            .q = q,
            .highs = highs,
            .lows = lows,
            .window_count = 0,
            .window_index = 0,
            .num_r = Ema.init(r),
            .num_s = Ema.init(s),
            .num_u = Ema.init(u),
            .den_r = Ema.init(r),
            .den_s = Ema.init(s),
            .den_u = Ema.init(u),
            .signal_ema = Ema.init(ul),
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .allocator = allocator,
        };
    }

    pub fn deinit(self: *DirectionalTrendIndex) void {
        self.allocator.free(self.highs);
        self.allocator.free(self.lows);
    }

    pub fn fixSlices(self: *DirectionalTrendIndex) void {
        _ = self;
        // DTI doesn't use LineIndicator; mnemonic/description are read from the
        // owned buffers directly, so no slice fixup is needed.
    }

    /// Updates the indicator given the next bar's high and low values.
    /// Returns dti, signal; both are NaN until q bars have been seen.
    pub fn updateValues(
        self: *DirectionalTrendIndex,
        high: f64,
        low: f64,
    ) struct { dti: f64, signal: f64 } {
        self.highs[self.window_index] = high;
        self.lows[self.window_index] = low;
        self.window_index = (self.window_index + 1) % self.q;

        if (self.window_count < self.q) {
            self.window_count += 1;
        }

        // HLM needs a high/low from q-1 bars ago. Until then neither output
        // exists -- do NOT advance the EMA cascades.
        if (self.window_count < self.q) {
            return .{ .dti = math.nan(f64), .signal = math.nan(f64) };
        }

        // The oldest value in the full window sits at the next write position:
        // high_(k-(q-1)) and low_(k-(q-1)).
        const previous_high = self.highs[self.window_index];
        const previous_low = self.lows[self.window_index];

        // Upward high movement and downward low movement, each floored at 0.
        const hmu = @max(high - previous_high, 0.0);
        const lmd = @max(previous_low - low, 0.0);

        // Composite high-low momentum and its magnitude.
        const hlm = hmu - lmd;
        const abs_hlm = @abs(hlm);

        // Numerator cascade: TEMA(HLM, r, s, u).
        const n = self.num_u.update(self.num_s.update(self.num_r.update(hlm)));
        // Denominator cascade: TEMA(|HLM|, r, s, u).
        const d = self.den_u.update(self.den_s.update(self.den_r.update(abs_hlm)));

        // Division guard: denominator 0 -> oscillator 0.0.
        const dti: f64 = if (d != 0.0) 100.0 * n / d else 0.0;

        // Signal line = EMA(dti, ul); seeds on the first finite oscillator value.
        const signal = self.signal_ema.update(dti);
        self.primed = true;

        return .{ .dti = dti, .signal = signal };
    }

    pub fn isPrimed(self: *const DirectionalTrendIndex) bool {
        return self.primed;
    }

    fn mnemonic(self: *const DirectionalTrendIndex) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const DirectionalTrendIndex) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const DirectionalTrendIndex, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var dti_mn_buf: [160]u8 = undefined;
        const dti_mn = std.fmt.bufPrint(&dti_mn_buf, "{s} dti", .{mn}) catch mn;
        var signal_mn_buf: [160]u8 = undefined;
        const signal_mn = std.fmt.bufPrint(&signal_mn_buf, "{s} signal", .{mn}) catch mn;

        var dti_desc_buf: [256]u8 = undefined;
        const dti_desc = std.fmt.bufPrint(&dti_desc_buf, "{s} DTI", .{desc}) catch desc;
        var signal_desc_buf: [256]u8 = undefined;
        const signal_desc = std.fmt.bufPrint(&signal_desc_buf, "{s} signal", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .directional_trend_index,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = dti_mn, .description = dti_desc },
                .{ .mnemonic = signal_mn, .description = signal_desc },
            },
        );
    }

    /// Updates the indicator given the next scalar sample.
    ///
    /// A scalar carries a single value, used as both the high and the low.
    pub fn updateScalar(self: *DirectionalTrendIndex, sample: *const Scalar) OutputArray {
        const v = sample.value;
        return self.updateEntity(sample.time, v, v);
    }

    pub fn updateBar(self: *DirectionalTrendIndex, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, sample.high, sample.low);
    }

    /// Updates the indicator given the next quote sample.
    ///
    /// A quote maps the mid price to both the high and the low.
    pub fn updateQuote(self: *DirectionalTrendIndex, sample: *const Quote) OutputArray {
        const v = (sample.bid_price + sample.ask_price) / 2.0;
        return self.updateEntity(sample.time, v, v);
    }

    /// Updates the indicator given the next trade sample.
    ///
    /// A trade carries a single price, used as both the high and the low.
    pub fn updateTrade(self: *DirectionalTrendIndex, sample: *const Trade) OutputArray {
        const v = sample.price;
        return self.updateEntity(sample.time, v, v);
    }

    fn updateEntity(
        self: *DirectionalTrendIndex,
        time: i64,
        high: f64,
        low: f64,
    ) OutputArray {
        const result = self.updateValues(high, low);

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = result.dti } });
        out.append(.{ .scalar = .{ .time = time, .value = result.signal } });
        return out;
    }

    pub fn indicator(self: *DirectionalTrendIndex) indicator_mod.Indicator {
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
        const self: *DirectionalTrendIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const DirectionalTrendIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DirectionalTrendIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DirectionalTrendIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DirectionalTrendIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DirectionalTrendIndex = @ptrCast(@alignCast(ptr));
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

const tolerance = 1e-13;

const Combo = struct {
    name: []const u8,
    q: usize,
    r: usize,
    s: usize,
    u: usize,
    ul: usize,
    dti: [252]f64,
    signal: [252]f64,
};

fn checkVal(exp: f64, act: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }
    try testing.expect(@abs(act - exp) <= tolerance);
}

test "DTI reference data all combos" {
    const high = testdata.testHigh();
    const low = testdata.testLow();

    const combos = [_]Combo{
        .{ .name = "Q2_R20_S5_U3", .q = 2, .r = 20, .s = 5, .u = 3, .ul = 3, .dti = testdata.expectedQ2_R20_S5_U3(), .signal = testdata.expectedQ2_R20_S5_U3_SIG_UL3() },
        .{ .name = "Q2_R25_S13_U1", .q = 2, .r = 25, .s = 13, .u = 1, .ul = 3, .dti = testdata.expectedQ2_R25_S13_U1(), .signal = testdata.expectedQ2_R25_S13_U1_SIG_UL3() },
        .{ .name = "Q2_R20_S5_U1", .q = 2, .r = 20, .s = 5, .u = 1, .ul = 3, .dti = testdata.expectedQ2_R20_S5_U1(), .signal = testdata.expectedQ2_R20_S5_U1_SIG_UL3() },
        .{ .name = "Q2_R28_S28_U5", .q = 2, .r = 28, .s = 28, .u = 5, .ul = 3, .dti = testdata.expectedQ2_R28_S28_U5(), .signal = testdata.expectedQ2_R28_S28_U5_SIG_UL3() },
        .{ .name = "Q2_R1_S1_U1", .q = 2, .r = 1, .s = 1, .u = 1, .ul = 3, .dti = testdata.expectedQ2_R1_S1_U1(), .signal = testdata.expectedQ2_R1_S1_U1_SIG_UL3() },
        .{ .name = "Q3_R20_S5_U3", .q = 3, .r = 20, .s = 5, .u = 3, .ul = 3, .dti = testdata.expectedQ3_R20_S5_U3(), .signal = testdata.expectedQ3_R20_S5_U3_SIG_UL3() },
        .{ .name = "Q5_R20_S5_U3", .q = 5, .r = 20, .s = 5, .u = 3, .ul = 3, .dti = testdata.expectedQ5_R20_S5_U3(), .signal = testdata.expectedQ5_R20_S5_U3_SIG_UL3() },
        .{ .name = "Q2_R13_S13_U1", .q = 2, .r = 13, .s = 13, .u = 1, .ul = 3, .dti = testdata.expectedQ2_R13_S13_U1(), .signal = testdata.expectedQ2_R13_S13_U1_SIG_UL3() },
        .{ .name = "Q2_R40_S20_U1", .q = 2, .r = 40, .s = 20, .u = 1, .ul = 3, .dti = testdata.expectedQ2_R40_S20_U1(), .signal = testdata.expectedQ2_R40_S20_U1_SIG_UL3() },
        .{ .name = "Q2_R5_S5_U5", .q = 2, .r = 5, .s = 5, .u = 5, .ul = 3, .dti = testdata.expectedQ2_R5_S5_U5(), .signal = testdata.expectedQ2_R5_S5_U5_SIG_UL3() },
        .{ .name = "Q1_R20_S5_U3", .q = 1, .r = 20, .s = 5, .u = 3, .ul = 3, .dti = testdata.expectedQ1_R20_S5_U3(), .signal = testdata.expectedQ1_R20_S5_U3_SIG_UL3() },
        .{ .name = "Q10_R20_S5_U1", .q = 10, .r = 20, .s = 5, .u = 1, .ul = 3, .dti = testdata.expectedQ10_R20_S5_U1(), .signal = testdata.expectedQ10_R20_S5_U1_SIG_UL3() },
        .{ .name = "Q2_R9_S3_U1", .q = 2, .r = 9, .s = 3, .u = 1, .ul = 3, .dti = testdata.expectedQ2_R9_S3_U1(), .signal = testdata.expectedQ2_R9_S3_U1_SIG_UL3() },
        .{ .name = "Q2_R64_S64_U1", .q = 2, .r = 64, .s = 64, .u = 1, .ul = 3, .dti = testdata.expectedQ2_R64_S64_U1(), .signal = testdata.expectedQ2_R64_S64_U1_SIG_UL3() },
        .{ .name = "Q4_R28_S28_U5", .q = 4, .r = 28, .s = 28, .u = 5, .ul = 3, .dti = testdata.expectedQ4_R28_S28_U5(), .signal = testdata.expectedQ4_R28_S28_U5_SIG_UL3() },
        .{ .name = "Q2_R7_S4_U2", .q = 2, .r = 7, .s = 4, .u = 2, .ul = 3, .dti = testdata.expectedQ2_R7_S4_U2(), .signal = testdata.expectedQ2_R7_S4_U2_SIG_UL3() },
    };

    for (combos) |combo| {
        var ind = try DirectionalTrendIndex.init(testing.allocator, .{
            .q = combo.q,
            .r = combo.r,
            .s = combo.s,
            .u = combo.u,
            .ul = combo.ul,
        });
        defer ind.deinit();

        for (0..252) |i| {
            const result = ind.updateValues(high[i], low[i]);
            try checkVal(combo.dti[i], result.dti);
            try checkVal(combo.signal[i], result.signal);
        }
    }
}

test "DTI passthrough reduces to 100*sign(HLM)" {
    var ind = try DirectionalTrendIndex.init(testing.allocator, .{ .q = 2, .r = 1, .s = 1, .u = 1, .ul = 1 });
    defer ind.deinit();

    const first = ind.updateValues(10.0, 9.0);
    try testing.expect(math.isNan(first.dti));
    try testing.expect(math.isNan(first.signal));

    const rising = ind.updateValues(12.0, 11.0); // HMU=+2, LMD=0
    try checkVal(100.0, rising.dti);
    try checkVal(100.0, rising.signal);

    const falling = ind.updateValues(11.0, 8.0); // HMU=0, LMD=3
    try checkVal(-100.0, falling.dti);
    try checkVal(-100.0, falling.signal);

    const inside = ind.updateValues(10.0, 9.0); // HMU=0, LMD=0 -> division guard
    try checkVal(0.0, inside.dti);
    try checkVal(0.0, inside.signal);
}

test "DTI with q = 1 yields zero on every bar" {
    const high = testdata.testHigh();
    const low = testdata.testLow();

    var ind = try DirectionalTrendIndex.init(testing.allocator, .{ .q = 1 });
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateValues(high[i], low[i]);
        try testing.expectEqual(@as(f64, 0.0), result.dti);
        try testing.expectEqual(@as(f64, 0.0), result.signal);
    }
}

test "DTI signal equals dti when ul = 1" {
    const high = testdata.testHigh();
    const low = testdata.testLow();

    var ind = try DirectionalTrendIndex.init(testing.allocator, .{ .ul = 1 });
    defer ind.deinit();

    for (0..252) |i| {
        const result = ind.updateValues(high[i], low[i]);
        if (math.isNan(result.dti)) {
            try testing.expect(math.isNan(result.signal));
        } else {
            try testing.expectEqual(result.dti, result.signal);
        }
    }
}

test "DTI is primed from bar q-1" {
    const high = testdata.testHigh();
    const low = testdata.testLow();

    for ([_]usize{ 2, 3, 5, 10 }) |q| {
        var ind = try DirectionalTrendIndex.init(testing.allocator, .{ .q = q });
        defer ind.deinit();
        try testing.expect(!ind.isPrimed());

        for (0..q - 1) |i| {
            const result = ind.updateValues(high[i], low[i]);
            try testing.expect(!ind.isPrimed());
            try testing.expect(math.isNan(result.dti));
            try testing.expect(math.isNan(result.signal));
        }

        for (q - 1..252) |i| {
            _ = ind.updateValues(high[i], low[i]);
            try testing.expect(ind.isPrimed());
        }
    }
}

test "DTI with q = 1 is primed after the first bar" {
    const high = testdata.testHigh();
    const low = testdata.testLow();

    var ind = try DirectionalTrendIndex.init(testing.allocator, .{ .q = 1 });
    defer ind.deinit();
    try testing.expect(!ind.isPrimed());

    _ = ind.updateValues(high[0], low[0]);
    try testing.expect(ind.isPrimed());
}

test "DTI metadata default" {
    var ind = try DirectionalTrendIndex.init(testing.allocator, .{});
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.directional_trend_index, meta.identifier);
    try testing.expectEqualStrings("dti(2,20,5,3,3)", meta.mnemonic);
    try testing.expectEqualStrings("Directional Trend Index dti(2,20,5,3,3)", meta.description);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
    try testing.expectEqualStrings("dti(2,20,5,3,3) dti", meta.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Directional Trend Index dti(2,20,5,3,3) DTI", meta.outputs_buf[0].description);
    try testing.expectEqualStrings("dti(2,20,5,3,3) signal", meta.outputs_buf[1].mnemonic);
    try testing.expectEqualStrings("Directional Trend Index dti(2,20,5,3,3) signal", meta.outputs_buf[1].description);
}

test "DTI custom mnemonic" {
    var ind = try DirectionalTrendIndex.init(testing.allocator, .{ .q = 4, .r = 28, .s = 28, .u = 5, .ul = 1 });
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("dti(4,28,28,5,1)", meta.mnemonic);
}

test "DTI invalid params" {
    try testing.expectError(error.InvalidQ, DirectionalTrendIndex.init(testing.allocator, .{ .q = 0 }));
    try testing.expectError(error.InvalidR, DirectionalTrendIndex.init(testing.allocator, .{ .r = 0 }));
    try testing.expectError(error.InvalidS, DirectionalTrendIndex.init(testing.allocator, .{ .s = 0 }));
    try testing.expectError(error.InvalidU, DirectionalTrendIndex.init(testing.allocator, .{ .u = 0 }));
    try testing.expectError(error.InvalidUl, DirectionalTrendIndex.init(testing.allocator, .{ .ul = 0 }));
}

test "DTI bar update ordering" {
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const exp_dti = testdata.expectedQ2_R20_S5_U3();
    const exp_signal = testdata.expectedQ2_R20_S5_U3_SIG_UL3();

    var ind = try DirectionalTrendIndex.init(testing.allocator, .{});
    defer ind.deinit();

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const bar = Bar{
            .time = 0,
            .open = 0.0,
            .high = high[i],
            .low = low[i],
            .close = 0.0,
            .volume = 0.0,
        };
        last_out = ind.updateBar(&bar);
    }
    const items = last_out.slice();

    try testing.expectEqual(@as(usize, 2), items.len);
    try checkVal(exp_dti[251], items[0].scalar.value);
    try checkVal(exp_signal[251], items[1].scalar.value);
}

test "DTI scalar is used as high and low" {
    var ind = try DirectionalTrendIndex.init(testing.allocator, .{ .q = 2, .r = 1, .s = 1, .u = 1, .ul = 1 });
    defer ind.deinit();

    const s0 = Scalar{ .time = 0, .value = 10.0 };
    const first = ind.updateScalar(&s0).slice();
    try testing.expect(math.isNan(first[0].scalar.value));
    try testing.expect(math.isNan(first[1].scalar.value));

    // Rising value: HLM is the plain one-bar momentum.
    const s1 = Scalar{ .time = 0, .value = 12.0 };
    const items = ind.updateScalar(&s1).slice();
    try checkVal(100.0, items[0].scalar.value);
    try checkVal(100.0, items[1].scalar.value);
}

test "DTI quote mid price is used as high and low" {
    var ind = try DirectionalTrendIndex.init(testing.allocator, .{ .q = 2, .r = 1, .s = 1, .u = 1, .ul = 1 });
    defer ind.deinit();

    const q0 = Quote{ .time = 0, .bid_price = 12.0, .ask_price = 14.0, .bid_size = 1.0, .ask_size = 1.0 };
    _ = ind.updateQuote(&q0);

    // Mid falls from 13 to 11.
    const q1 = Quote{ .time = 0, .bid_price = 10.0, .ask_price = 12.0, .bid_size = 1.0, .ask_size = 1.0 };
    const items = ind.updateQuote(&q1).slice();
    try checkVal(-100.0, items[0].scalar.value);
    try checkVal(-100.0, items[1].scalar.value);
}

test "DTI trade price is used as high and low" {
    var ind = try DirectionalTrendIndex.init(testing.allocator, .{ .q = 2, .r = 1, .s = 1, .u = 1, .ul = 1 });
    defer ind.deinit();

    const t0 = Trade{ .time = 0, .price = 10.0, .volume = 1.0 };
    _ = ind.updateTrade(&t0);

    const t1 = Trade{ .time = 0, .price = 12.0, .volume = 1.0 };
    const items = ind.updateTrade(&t1).slice();
    try checkVal(100.0, items[0].scalar.value);
    try checkVal(100.0, items[1].scalar.value);
}
