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

/// Enumerates the outputs of the Tick Volume Indicator.
pub const TickVolumeIndicatorOutput = enum(u8) {
    /// Tick Volume Indicator oscillator value (range [-100, +100]).
    value = 1,
};

/// Parameters to create a Tick Volume Indicator.
///
/// The field names r, s and u are the canonical symbols from William Blau's
/// Momentum, Direction, and Divergence (Wiley, 1995), chapters 4 and 10.
///
/// The indicator consumes upticks and downticks (derived from the bar range,
/// or from consecutive prices for single-valued samples), so it has no
/// configurable price-component fields.
pub const TickVolumeIndicatorParams = struct {
    r: usize = 12,
    s: usize = 12,
    u: usize = 1,
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

/// Tick Volume Indicator (TVI) by William Blau.
///
/// A normalized, double-/triple-smoothed oscillator built from the balance of
/// upticks vs downticks inside each bar, bounded to [-100, +100] (Blau ch.4, ch.10):
///
///   tvi_k = 100 * (TEMA(up, r, s, u) - TEMA(down, r, s, u))
///               / (TEMA(up, r, s, u) + TEMA(down, r, s, u))
///
/// where TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u). Setting u=1 recovers the
/// book's double-smoothed TVI(r, s), because EMA(., 1) is a passthrough. Because it
/// is built from intra-bar tick direction rather than from the close vs a previous
/// close, the TVI is immune to opening gaps.
///
/// The inputs are two non-negative series, upticks and downticks. Genuine tick
/// counts are fed through `update`. The entity updates derive a deterministic proxy:
///   - Bar: up = close - low, down = high - close (the intra-bar range split);
///   - Scalar, Trade, Quote: a magnitude tick rule against the previous value
///     (value, price or mid price): up = max(x - previous, 0),
///     down = max(previous - x, 0). The first sample yields (0, 0). On such a
///     single-valued series the TVI reduces to a True Strength Index of the
///     one-step momentum.
///
/// Priming convention (book / EasyLanguage): each EMA stage seeds on its first
/// received value, so there is no NaN warm-up region and the output is finite for
/// every update. Division guard: denominator 0 (a fully flat market) -> 0.0.
pub const TickVolumeIndicator = struct {
    up_r: Ema,
    up_s: Ema,
    up_u: Ema,
    down_r: Ema,
    down_s: Ema,
    down_u: Ema,

    previous: f64,
    has_previous: bool,

    primed: bool,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(params: TickVolumeIndicatorParams) !TickVolumeIndicator {
        const r = params.r;
        const s = params.s;
        const u = params.u;

        if (r < 1) return error.InvalidR;
        if (s < 1) return error.InvalidS;
        if (u < 1) return error.InvalidU;

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "tvi({d},{d},{d})", .{
            r,
            s,
            u,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Tick Volume Indicator {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .up_r = Ema.init(r),
            .up_s = Ema.init(s),
            .up_u = Ema.init(u),
            .down_r = Ema.init(r),
            .down_s = Ema.init(s),
            .down_u = Ema.init(u),
            .previous = 0.0,
            .has_previous = false,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *TickVolumeIndicator) void {
        _ = self;
        // TVI doesn't use LineIndicator; mnemonic/description are read from the
        // owned buffers directly, so no slice fixup is needed.
    }

    /// Updates the indicator given the next bar's upticks and downticks.
    /// Returns the TVI value.
    pub fn update(self: *TickVolumeIndicator, upticks: f64, downticks: f64) f64 {
        // Upticks cascade: TEMA(up, r, s, u).
        const up = self.up_u.update(self.up_s.update(self.up_r.update(upticks)));
        // Downticks cascade: TEMA(down, r, s, u).
        const down = self.down_u.update(self.down_s.update(self.down_r.update(downticks)));

        self.primed = true;

        // Division guard (Appendix B): fully flat smoothed volume -> 0.0.
        const denominator = up + down;
        if (denominator == 0.0) return 0.0;

        return 100.0 * (up - down) / denominator;
    }

    /// Derives the upticks and downticks from the change of the value vs the
    /// previous value and updates the indicator.
    fn updateTickRule(self: *TickVolumeIndicator, value: f64) f64 {
        var upticks: f64 = 0.0;
        var downticks: f64 = 0.0;
        if (self.has_previous) {
            const diff = value - self.previous;
            upticks = @max(diff, 0.0);
            downticks = @max(-diff, 0.0);
        }

        self.previous = value;
        self.has_previous = true;

        return self.update(upticks, downticks);
    }

    pub fn isPrimed(self: *const TickVolumeIndicator) bool {
        return self.primed;
    }

    fn mnemonic(self: *const TickVolumeIndicator) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const TickVolumeIndicator) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const TickVolumeIndicator, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        build_metadata_mod.buildMetadata(
            out,
            .tick_volume_indicator,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    /// Updates the indicator given the next scalar sample.
    ///
    /// A scalar carries a single value, so the tick rule is applied to it.
    pub fn updateScalar(self: *TickVolumeIndicator, sample: *const Scalar) OutputArray {
        return wrap(sample.time, self.updateTickRule(sample.value));
    }

    /// Updates the indicator given the next bar sample.
    ///
    /// A bar splits its range: up = close - low, down = high - close.
    pub fn updateBar(self: *TickVolumeIndicator, sample: *const Bar) OutputArray {
        return wrap(sample.time, self.update(sample.close - sample.low, sample.high - sample.close));
    }

    /// Updates the indicator given the next quote sample.
    ///
    /// A quote applies the tick rule to its mid price.
    pub fn updateQuote(self: *TickVolumeIndicator, sample: *const Quote) OutputArray {
        return wrap(sample.time, self.updateTickRule(sample.mid()));
    }

    /// Updates the indicator given the next trade sample.
    ///
    /// A trade applies the tick rule to its price.
    pub fn updateTrade(self: *TickVolumeIndicator, sample: *const Trade) OutputArray {
        return wrap(sample.time, self.updateTickRule(sample.price));
    }

    /// Wraps the TVI value into the output.
    fn wrap(time: i64, value: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = value } });
        return out;
    }

    pub fn indicator(self: *TickVolumeIndicator) indicator_mod.Indicator {
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
        const self: *TickVolumeIndicator = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const TickVolumeIndicator = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *TickVolumeIndicator = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *TickVolumeIndicator = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *TickVolumeIndicator = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *TickVolumeIndicator = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidR,
        InvalidS,
        InvalidU,
        MnemonicTooLong,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

const tolerance = 1e-10;

const Combo = struct {
    name: []const u8,
    r: usize,
    s: usize,
    u: usize,
    expected: [252]f64,
};

fn checkVal(exp: f64, act: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }
    try testing.expect(@abs(act - exp) <= tolerance);
}

fn checkOutput(out: OutputArray, time: i64, exp: f64) !void {
    const items = out.slice();
    try testing.expectEqual(@as(usize, 1), items.len);
    try testing.expectEqual(time, items[0].scalar.time);
    try checkVal(exp, items[0].scalar.value);
}

// Tick rule with passthrough stages: first -> 0, up -> +100, down -> -100, flat -> 0.
const tick_values = [_]f64{ 10.0, 12.0, 11.0, 11.0 };
const tick_expected = [_]f64{ 0.0, 100.0, -100.0, 0.0 };

test "TVI reference data all combos" {
    const upticks = testdata.testUpticks();
    const downticks = testdata.testDownticks();

    const combos = [_]Combo{
        .{ .name = "R12_S12_U1", .r = 12, .s = 12, .u = 1, .expected = testdata.expectedR12_S12_U1() },
        .{ .name = "R25_S13_U1", .r = 25, .s = 13, .u = 1, .expected = testdata.expectedR25_S13_U1() },
        .{ .name = "R32_S32_U5", .r = 32, .s = 32, .u = 5, .expected = testdata.expectedR32_S32_U5() },
        .{ .name = "R1_S1_U1", .r = 1, .s = 1, .u = 1, .expected = testdata.expectedR1_S1_U1() },
        .{ .name = "R32_S5_U1", .r = 32, .s = 5, .u = 1, .expected = testdata.expectedR32_S5_U1() },
        .{ .name = "R12_S12_U5", .r = 12, .s = 12, .u = 5, .expected = testdata.expectedR12_S12_U5() },
        .{ .name = "R20_S5_U3", .r = 20, .s = 5, .u = 3, .expected = testdata.expectedR20_S5_U3() },
        .{ .name = "R5_S5_U5", .r = 5, .s = 5, .u = 5, .expected = testdata.expectedR5_S5_U5() },
        .{ .name = "R32_S32_U1", .r = 32, .s = 32, .u = 1, .expected = testdata.expectedR32_S32_U1() },
        .{ .name = "R10_S10_U1", .r = 10, .s = 10, .u = 1, .expected = testdata.expectedR10_S10_U1() },
        .{ .name = "R50_S25_U1", .r = 50, .s = 25, .u = 1, .expected = testdata.expectedR50_S25_U1() },
        .{ .name = "R12_S26_U9", .r = 12, .s = 26, .u = 9, .expected = testdata.expectedR12_S26_U9() },
        .{ .name = "R3_S3_U3", .r = 3, .s = 3, .u = 3, .expected = testdata.expectedR3_S3_U3() },
        .{ .name = "R7_S4_U2", .r = 7, .s = 4, .u = 2, .expected = testdata.expectedR7_S4_U2() },
        .{ .name = "R64_S1_U1", .r = 64, .s = 1, .u = 1, .expected = testdata.expectedR64_S1_U1() },
        .{ .name = "R12_S12_U3", .r = 12, .s = 12, .u = 3, .expected = testdata.expectedR12_S12_U3() },
    };

    for (combos) |combo| {
        var ind = try TickVolumeIndicator.init(.{ .r = combo.r, .s = combo.s, .u = combo.u });

        for (0..252) |i| {
            try checkVal(combo.expected[i], ind.update(upticks[i], downticks[i]));
        }
    }
}

test "TVI passthrough reduces to the normalized tick balance" {
    var ind = try TickVolumeIndicator.init(.{ .r = 1, .s = 1, .u = 1 });

    try checkVal(60.0, ind.update(8.0, 2.0));
    try checkVal(-100.0, ind.update(0.0, 5.0));
    // Flat market: denominator 0 -> division guard 0.0.
    try checkVal(0.0, ind.update(0.0, 0.0));
}

test "TVI is primed after the first update" {
    const upticks = testdata.testUpticks();
    const downticks = testdata.testDownticks();

    var ind = try TickVolumeIndicator.init(.{});
    try testing.expect(!ind.isPrimed());
    _ = ind.update(upticks[0], downticks[0]);
    try testing.expect(ind.isPrimed());

    var tick_ind = try TickVolumeIndicator.init(.{});
    try testing.expect(!tick_ind.isPrimed());
    const scalar = Scalar{ .time = 0, .value = 10.0 };
    _ = tick_ind.updateScalar(&scalar);
    try testing.expect(tick_ind.isPrimed());
}

test "TVI metadata default" {
    var ind = try TickVolumeIndicator.init(.{});

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.tick_volume_indicator, meta.identifier);
    try testing.expectEqualStrings("tvi(12,12,1)", meta.mnemonic);
    try testing.expectEqualStrings("Tick Volume Indicator tvi(12,12,1)", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@intFromEnum(TickVolumeIndicatorOutput.value), meta.outputs_buf[0].kind);
    try testing.expectEqualStrings("tvi(12,12,1)", meta.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Tick Volume Indicator tvi(12,12,1)", meta.outputs_buf[0].description);
}

test "TVI custom mnemonic" {
    var ind = try TickVolumeIndicator.init(.{ .r = 32, .s = 32, .u = 5 });

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("tvi(32,32,5)", meta.mnemonic);
}

test "TVI invalid params" {
    try testing.expectError(error.InvalidR, TickVolumeIndicator.init(.{ .r = 0 }));
    try testing.expectError(error.InvalidS, TickVolumeIndicator.init(.{ .s = 0 }));
    try testing.expectError(error.InvalidU, TickVolumeIndicator.init(.{ .u = 0 }));
}

test "TVI bar maps up = close - low and down = high - close" {
    const input = testdata.testInput();
    const open = testdata.testOpen();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const expected = testdata.expectedR12_S12_U1();

    var ind = try TickVolumeIndicator.init(.{ .r = 12, .s = 12, .u = 1 });

    for (0..252) |i| {
        const bar = Bar{
            .time = 42,
            .open = open[i],
            .high = high[i],
            .low = low[i],
            .close = input[i],
            .volume = 0.0,
        };
        try checkOutput(ind.updateBar(&bar), 42, expected[i]);
    }
}

test "TVI scalar applies the tick rule to the value" {
    var ind = try TickVolumeIndicator.init(.{ .r = 1, .s = 1, .u = 1 });

    for (tick_values, 0..) |value, i| {
        const scalar = Scalar{ .time = 42, .value = value };
        try checkOutput(ind.updateScalar(&scalar), 42, tick_expected[i]);
    }
}

test "TVI trade applies the tick rule to the price" {
    var ind = try TickVolumeIndicator.init(.{ .r = 1, .s = 1, .u = 1 });

    for (tick_values, 0..) |price, i| {
        const trade = Trade{ .time = 42, .price = price, .volume = 1.0 };
        try checkOutput(ind.updateTrade(&trade), 42, tick_expected[i]);
    }
}

test "TVI quote applies the tick rule to the mid price" {
    var ind = try TickVolumeIndicator.init(.{ .r = 1, .s = 1, .u = 1 });

    const quotes = [_]Quote{
        .{ .time = 42, .bid_price = 9.0, .ask_price = 11.0, .bid_size = 1.0, .ask_size = 1.0 },
        .{ .time = 42, .bid_price = 11.0, .ask_price = 13.0, .bid_size = 1.0, .ask_size = 1.0 },
        .{ .time = 42, .bid_price = 10.0, .ask_price = 12.0, .bid_size = 1.0, .ask_size = 1.0 },
        .{ .time = 42, .bid_price = 10.5, .ask_price = 11.5, .bid_size = 1.0, .ask_size = 1.0 },
    };

    for (quotes, 0..) |quote, i| {
        try checkOutput(ind.updateQuote(&quote), 42, tick_expected[i]);
    }
}

test "TVI scalar matches update of the value changes" {
    const input = testdata.testInput();

    var ind = try TickVolumeIndicator.init(.{ .r = 12, .s = 12, .u = 3 });
    var ref = try TickVolumeIndicator.init(.{ .r = 12, .s = 12, .u = 3 });

    for (0..252) |i| {
        const diff: f64 = if (i == 0) 0.0 else input[i] - input[i - 1];
        const exp = ref.update(@max(diff, 0.0), @max(-diff, 0.0));
        const scalar = Scalar{ .time = 42, .value = input[i] };
        try checkOutput(ind.updateScalar(&scalar), 42, exp);
    }
}
