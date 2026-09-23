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

/// Enumerates the outputs of the Stochastic Momentum Index indicator.
pub const StochasticMomentumIndexOutput = enum(u8) {
    /// Stochastic Momentum Index oscillator value (range [-100, +100]).
    smi = 1,
    /// Signal-line value: the ul-period EMA of the oscillator.
    signal = 2,
};

/// Parameters to create a Stochastic Momentum Index indicator.
///
/// The field names q, r, s, u and ul are the canonical symbols from William Blau's
/// Momentum, Direction, and Divergence (Wiley, 1995), chapter 3.
///
/// The indicator consumes the high, low and close prices of a bar, so it has no
/// configurable price-component fields.
pub const StochasticMomentumIndexParams = struct {
    q: usize = 5,
    r: usize = 20,
    s: usize = 5,
    u: usize = 3,
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
    prev: f64 = 0.0,
    primed: bool = false,

    fn init(period: usize) Ema {
        return .{ .alpha = 2.0 / (@as(f64, @floatFromInt(period)) + 1.0) };
    }

    fn update(self: *Ema, x: f64) f64 {
        if (!self.primed) {
            self.prev = x;
            self.primed = true;
            return self.prev;
        }
        self.prev = self.alpha * x + (1.0 - self.alpha) * self.prev;
        return self.prev;
    }
};

/// Stochastic Momentum Index (SMI) by William Blau.
///
/// A double-/triple-smoothed stochastic oscillator bounded to [-100, +100],
/// paired with an EMA signal line (the Ergodic form, Blau ch.3.4):
///
///   smi_k    = 100 * TEMA(sm, r, s, u) / TEMA(hr, r, s, u)   (the oscillator)
///   signal_k = EMA(smi, ul)_k                                (ul-period EMA)
///
/// where, over the last q bars, HH_k is the highest high and LL_k is the lowest low,
/// sm_k = close_k - 0.5*(HH_k + LL_k) is the distance of the close from the range
/// midpoint, hr_k = 0.5*(HH_k - LL_k) >= 0 is the half-range, and
/// TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u).
///
/// Where the ordinary stochastic measures where the close sits inside the recent
/// high-low range, the SMI measures the close relative to the midpoint of that range.
/// Because |sm| <= hr on every bar, the ratio is bounded to [-100, +100]. With q = 1 it
/// is Blau's one-day stochastic (sentiment indicator). The inputs are the high, low and
/// close prices.
///
/// The indicator produces two outputs:
///   - SMI: the oscillator, range [-100, +100];
///   - Signal: the ul-period EMA of the oscillator (Blau's Ergodic signal line).
///
/// Priming convention (book / EasyLanguage): sm and hr become valid once q bars of
/// high/low exist, i.e. at bar q-1. All six cascade stages seed there together, so
/// both outputs are NaN for bars 0..q-2 and finite from bar q-1; for q = 1 there is
/// no NaN warm-up. Division guard: denominator <= 0 -> oscillator 0.0.
pub const StochasticMomentumIndex = struct {
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

    pub fn init(allocator: std.mem.Allocator, params: StochasticMomentumIndexParams) InitError!StochasticMomentumIndex {
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
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "smi({d},{d},{d},{d},{d})", .{
            q,
            r,
            s,
            u,
            ul,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Stochastic Momentum Index {s}", .{mnemonic_slice}) catch
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

    pub fn deinit(self: *StochasticMomentumIndex) void {
        self.allocator.free(self.highs);
        self.allocator.free(self.lows);
    }

    pub fn fixSlices(self: *StochasticMomentumIndex) void {
        _ = self;
        // SMI doesn't use LineIndicator; mnemonic/description are read from the
        // owned buffers directly, so no slice fixup is needed.
    }

    /// Updates the indicator given the next bar's high, low and close values.
    /// Returns smi, signal; both are NaN until q bars have been seen.
    pub fn updateValues(
        self: *StochasticMomentumIndex,
        high: f64,
        low: f64,
        close: f64,
    ) struct { smi: f64, signal: f64 } {
        self.highs[self.window_index] = high;
        self.lows[self.window_index] = low;
        self.window_index = (self.window_index + 1) % self.q;

        if (self.window_count < self.q) {
            self.window_count += 1;
        }

        // Need q bars of high/low before the stochastic is defined. Until then
        // neither output exists -- do NOT advance the EMA cascades.
        if (self.window_count < self.q) {
            return .{ .smi = math.nan(f64), .signal = math.nan(f64) };
        }

        // Rolling extremes over the last q bars.
        var hh = self.highs[0];
        var ll = self.lows[0];
        for (1..self.q) |i| {
            hh = @max(hh, self.highs[i]);
            ll = @min(ll, self.lows[i]);
        }

        // Stochastic momentum (signed) and half-range (non-negative).
        const sm = close - 0.5 * (hh + ll);
        const hr = 0.5 * (hh - ll);

        // Numerator cascade: TEMA(sm, r, s, u).
        const n = self.num_u.update(self.num_s.update(self.num_r.update(sm)));
        // Denominator cascade: TEMA(hr, r, s, u).
        const d = self.den_u.update(self.den_s.update(self.den_r.update(hr)));

        // Division guard: flat window so far -> oscillator 0.0.
        const smi: f64 = if (d > 0.0) 100.0 * n / d else 0.0;

        // Signal line = EMA(smi, ul); seeds on the first finite oscillator value.
        const signal = self.signal_ema.update(smi);
        self.primed = true;

        return .{ .smi = smi, .signal = signal };
    }

    pub fn isPrimed(self: *const StochasticMomentumIndex) bool {
        return self.primed;
    }

    fn mnemonic(self: *const StochasticMomentumIndex) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const StochasticMomentumIndex) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const StochasticMomentumIndex, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var smi_mn_buf: [160]u8 = undefined;
        const smi_mn = std.fmt.bufPrint(&smi_mn_buf, "{s} smi", .{mn}) catch mn;
        var signal_mn_buf: [160]u8 = undefined;
        const signal_mn = std.fmt.bufPrint(&signal_mn_buf, "{s} signal", .{mn}) catch mn;

        var smi_desc_buf: [256]u8 = undefined;
        const smi_desc = std.fmt.bufPrint(&smi_desc_buf, "{s} SMI", .{desc}) catch desc;
        var signal_desc_buf: [256]u8 = undefined;
        const signal_desc = std.fmt.bufPrint(&signal_desc_buf, "{s} signal", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .stochastic_momentum_index,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = smi_mn, .description = smi_desc },
                .{ .mnemonic = signal_mn, .description = signal_desc },
            },
        );
    }

    /// Updates the indicator given the next scalar sample.
    ///
    /// A scalar carries a single value, used as the high, the low and the close.
    pub fn updateScalar(self: *StochasticMomentumIndex, sample: *const Scalar) OutputArray {
        const v = sample.value;
        return self.updateEntity(sample.time, v, v, v);
    }

    pub fn updateBar(self: *StochasticMomentumIndex, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, sample.high, sample.low, sample.close);
    }

    /// Updates the indicator given the next quote sample.
    ///
    /// A quote maps the mid price to the high, the low and the close.
    pub fn updateQuote(self: *StochasticMomentumIndex, sample: *const Quote) OutputArray {
        const v = (sample.bid_price + sample.ask_price) / 2.0;
        return self.updateEntity(sample.time, v, v, v);
    }

    /// Updates the indicator given the next trade sample.
    ///
    /// A trade carries a single price, used as the high, the low and the close.
    pub fn updateTrade(self: *StochasticMomentumIndex, sample: *const Trade) OutputArray {
        const v = sample.price;
        return self.updateEntity(sample.time, v, v, v);
    }

    fn updateEntity(
        self: *StochasticMomentumIndex,
        time: i64,
        high: f64,
        low: f64,
        close: f64,
    ) OutputArray {
        const result = self.updateValues(high, low, close);

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = result.smi } });
        out.append(.{ .scalar = .{ .time = time, .value = result.signal } });
        return out;
    }

    pub fn indicator(self: *StochasticMomentumIndex) indicator_mod.Indicator {
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
        const self: *StochasticMomentumIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const StochasticMomentumIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *StochasticMomentumIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *StochasticMomentumIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *StochasticMomentumIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *StochasticMomentumIndex = @ptrCast(@alignCast(ptr));
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

// signal_ul is the signal-line EMA period used for every expected signal array.
const signal_ul = 3;

const Combo = struct {
    name: []const u8,
    q: usize,
    r: usize,
    s: usize,
    u: usize,
    smi: [252]f64,
    signal: [252]f64,
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }
    try testing.expect(@abs(act - exp) <= tolerance);
}

test "SMI reference data all combos" {
    const tolerance = 1e-9;
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    const combos = [_]Combo{
        .{ .name = "Q5_R20_S5_U3", .q = 5, .r = 20, .s = 5, .u = 3, .smi = testdata.expectedQ5_R20_S5_U3(), .signal = testdata.expectedQ5_R20_S5_U3_SIG_UL3() },
        .{ .name = "Q13_R25_S2_U1", .q = 13, .r = 25, .s = 2, .u = 1, .smi = testdata.expectedQ13_R25_S2_U1(), .signal = testdata.expectedQ13_R25_S2_U1_SIG_UL3() },
        .{ .name = "Q2_R20_S20_U1", .q = 2, .r = 20, .s = 20, .u = 1, .smi = testdata.expectedQ2_R20_S20_U1(), .signal = testdata.expectedQ2_R20_S20_U1_SIG_UL3() },
        .{ .name = "Q13_R25_S2_U3", .q = 13, .r = 25, .s = 2, .u = 3, .smi = testdata.expectedQ13_R25_S2_U3(), .signal = testdata.expectedQ13_R25_S2_U3_SIG_UL3() },
        .{ .name = "Q5_R20_S5_U1", .q = 5, .r = 20, .s = 5, .u = 1, .smi = testdata.expectedQ5_R20_S5_U1(), .signal = testdata.expectedQ5_R20_S5_U1_SIG_UL3() },
        .{ .name = "Q8_R5_S3_U1", .q = 8, .r = 5, .s = 3, .u = 1, .smi = testdata.expectedQ8_R5_S3_U1(), .signal = testdata.expectedQ8_R5_S3_U1_SIG_UL3() },
        .{ .name = "Q21_R13_S4_U1", .q = 21, .r = 13, .s = 4, .u = 1, .smi = testdata.expectedQ21_R13_S4_U1(), .signal = testdata.expectedQ21_R13_S4_U1_SIG_UL3() },
        .{ .name = "Q1_R20_S5_U3", .q = 1, .r = 20, .s = 5, .u = 3, .smi = testdata.expectedQ1_R20_S5_U3(), .signal = testdata.expectedQ1_R20_S5_U3_SIG_UL3() },
        .{ .name = "Q1_R40_S20_U1", .q = 1, .r = 40, .s = 20, .u = 1, .smi = testdata.expectedQ1_R40_S20_U1(), .signal = testdata.expectedQ1_R40_S20_U1_SIG_UL3() },
        .{ .name = "Q1_R100_S20_U1", .q = 1, .r = 100, .s = 20, .u = 1, .smi = testdata.expectedQ1_R100_S20_U1(), .signal = testdata.expectedQ1_R100_S20_U1_SIG_UL3() },
        .{ .name = "Q1_R1_S1_U1", .q = 1, .r = 1, .s = 1, .u = 1, .smi = testdata.expectedQ1_R1_S1_U1(), .signal = testdata.expectedQ1_R1_S1_U1_SIG_UL3() },
        .{ .name = "Q5_R1_S1_U1", .q = 5, .r = 1, .s = 1, .u = 1, .smi = testdata.expectedQ5_R1_S1_U1(), .signal = testdata.expectedQ5_R1_S1_U1_SIG_UL3() },
        .{ .name = "Q3_R10_S10_U1", .q = 3, .r = 10, .s = 10, .u = 1, .smi = testdata.expectedQ3_R10_S10_U1(), .signal = testdata.expectedQ3_R10_S10_U1_SIG_UL3() },
        .{ .name = "Q34_R5_S5_U1", .q = 34, .r = 5, .s = 5, .u = 1, .smi = testdata.expectedQ34_R5_S5_U1(), .signal = testdata.expectedQ34_R5_S5_U1_SIG_UL3() },
        .{ .name = "Q2_R2_S2_U2", .q = 2, .r = 2, .s = 2, .u = 2, .smi = testdata.expectedQ2_R2_S2_U2(), .signal = testdata.expectedQ2_R2_S2_U2_SIG_UL3() },
        .{ .name = "Q50_R20_S5_U3", .q = 50, .r = 20, .s = 5, .u = 3, .smi = testdata.expectedQ50_R20_S5_U3(), .signal = testdata.expectedQ50_R20_S5_U3_SIG_UL3() },
    };

    for (combos) |combo| {
        var ind = try StochasticMomentumIndex.init(testing.allocator, .{
            .q = combo.q,
            .r = combo.r,
            .s = combo.s,
            .u = combo.u,
            .ul = signal_ul,
        });
        defer ind.deinit();

        for (0..252) |i| {
            const result = ind.updateValues(high[i], low[i], input[i]);
            try checkVal(combo.smi[i], result.smi, tolerance);
            try checkVal(combo.signal[i], result.signal, tolerance);
        }
    }
}

test "SMI passthrough reduces to the raw one-day stochastic" {
    const tolerance = 1e-9;

    var ind = try StochasticMomentumIndex.init(testing.allocator, .{ .q = 1, .r = 1, .s = 1, .u = 1, .ul = 1 });
    defer ind.deinit();

    const at_high = ind.updateValues(12.0, 10.0, 12.0);
    try checkVal(100.0, at_high.smi, tolerance);
    try checkVal(100.0, at_high.signal, tolerance);

    const at_low = ind.updateValues(12.0, 10.0, 10.0);
    try checkVal(-100.0, at_low.smi, tolerance);
    try checkVal(-100.0, at_low.signal, tolerance);

    const at_mid = ind.updateValues(12.0, 10.0, 11.0);
    try checkVal(0.0, at_mid.smi, tolerance);
    try checkVal(0.0, at_mid.signal, tolerance);

    const flat = ind.updateValues(11.0, 11.0, 11.0);
    try checkVal(0.0, flat.smi, tolerance);
    try checkVal(0.0, flat.signal, tolerance);
}

test "SMI is primed from bar q-1" {
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const q = 5;

    var ind = try StochasticMomentumIndex.init(testing.allocator, .{ .q = q });
    defer ind.deinit();
    try testing.expect(!ind.isPrimed());

    for (0..q - 1) |i| {
        const result = ind.updateValues(high[i], low[i], input[i]);
        try testing.expect(!ind.isPrimed());
        try testing.expect(math.isNan(result.smi));
        try testing.expect(math.isNan(result.signal));
    }

    for (q - 1..252) |i| {
        _ = ind.updateValues(high[i], low[i], input[i]);
        try testing.expect(ind.isPrimed());
    }
}

test "SMI with q = 1 is primed after the first bar" {
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    var ind = try StochasticMomentumIndex.init(testing.allocator, .{ .q = 1 });
    defer ind.deinit();
    try testing.expect(!ind.isPrimed());

    _ = ind.updateValues(high[0], low[0], input[0]);
    try testing.expect(ind.isPrimed());
}

test "SMI metadata default" {
    var ind = try StochasticMomentumIndex.init(testing.allocator, .{});
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.stochastic_momentum_index, meta.identifier);
    try testing.expectEqualStrings("smi(5,20,5,3,3)", meta.mnemonic);
    try testing.expectEqualStrings("Stochastic Momentum Index smi(5,20,5,3,3)", meta.description);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
    try testing.expectEqualStrings("smi(5,20,5,3,3) smi", meta.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("smi(5,20,5,3,3) signal", meta.outputs_buf[1].mnemonic);
}

test "SMI custom mnemonic" {
    var ind = try StochasticMomentumIndex.init(testing.allocator, .{ .q = 13, .r = 25, .s = 2, .u = 1, .ul = 7 });
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("smi(13,25,2,1,7)", meta.mnemonic);
}

test "SMI invalid params" {
    try testing.expectError(error.InvalidQ, StochasticMomentumIndex.init(testing.allocator, .{ .q = 0 }));
    try testing.expectError(error.InvalidR, StochasticMomentumIndex.init(testing.allocator, .{ .r = 0 }));
    try testing.expectError(error.InvalidS, StochasticMomentumIndex.init(testing.allocator, .{ .s = 0 }));
    try testing.expectError(error.InvalidU, StochasticMomentumIndex.init(testing.allocator, .{ .u = 0 }));
    try testing.expectError(error.InvalidUl, StochasticMomentumIndex.init(testing.allocator, .{ .ul = 0 }));
}

test "SMI bar update ordering" {
    const tolerance = 1e-9;
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const exp_smi = testdata.expectedQ5_R20_S5_U3();
    const exp_signal = testdata.expectedQ5_R20_S5_U3_SIG_UL3();

    var ind = try StochasticMomentumIndex.init(testing.allocator, .{});
    defer ind.deinit();

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const bar = Bar{
            .time = 0,
            .open = 0.0,
            .high = high[i],
            .low = low[i],
            .close = input[i],
            .volume = 0.0,
        };
        last_out = ind.updateBar(&bar);
    }
    const items = last_out.slice();

    try testing.expectEqual(@as(usize, 2), items.len);
    try checkVal(exp_smi[251], items[0].scalar.value, tolerance);
    try checkVal(exp_signal[251], items[1].scalar.value, tolerance);
}

test "SMI scalar is used as high, low and close" {
    const tolerance = 1e-9;

    var ind = try StochasticMomentumIndex.init(testing.allocator, .{ .q = 2, .r = 1, .s = 1, .u = 1, .ul = 1 });
    defer ind.deinit();

    const s0 = Scalar{ .time = 0, .value = 10.0 };
    const first = ind.updateScalar(&s0).slice();
    try testing.expect(math.isNan(first[0].scalar.value));
    try testing.expect(math.isNan(first[1].scalar.value));

    // Close at the 2-bar high.
    const s1 = Scalar{ .time = 0, .value = 12.0 };
    const items = ind.updateScalar(&s1).slice();
    try checkVal(100.0, items[0].scalar.value, tolerance);
    try checkVal(100.0, items[1].scalar.value, tolerance);
}

test "SMI quote mid price is used as high, low and close" {
    const tolerance = 1e-9;

    var ind = try StochasticMomentumIndex.init(testing.allocator, .{ .q = 2, .r = 1, .s = 1, .u = 1, .ul = 1 });
    defer ind.deinit();

    const q0 = Quote{ .time = 0, .bid_price = 12.0, .ask_price = 14.0, .bid_size = 1.0, .ask_size = 1.0 };
    _ = ind.updateQuote(&q0);

    // Mid 11 is at the 2-bar low.
    const q1 = Quote{ .time = 0, .bid_price = 10.0, .ask_price = 12.0, .bid_size = 1.0, .ask_size = 1.0 };
    const items = ind.updateQuote(&q1).slice();
    try checkVal(-100.0, items[0].scalar.value, tolerance);
    try checkVal(-100.0, items[1].scalar.value, tolerance);
}

test "SMI trade price is used as high, low and close" {
    const tolerance = 1e-9;

    var ind = try StochasticMomentumIndex.init(testing.allocator, .{ .q = 2, .r = 1, .s = 1, .u = 1, .ul = 1 });
    defer ind.deinit();

    const t0 = Trade{ .time = 0, .price = 10.0, .volume = 1.0 };
    _ = ind.updateTrade(&t0);

    const t1 = Trade{ .time = 0, .price = 12.0, .volume = 1.0 };
    const items = ind.updateTrade(&t1).slice();
    try checkVal(100.0, items[0].scalar.value, tolerance);
    try checkVal(100.0, items[1].scalar.value, tolerance);
}
