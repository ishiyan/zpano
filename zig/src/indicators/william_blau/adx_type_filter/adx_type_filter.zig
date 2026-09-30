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

/// Enumerates the outputs of the ADX-Type Filter indicator.
pub const AdxTypeFilterOutput = enum(u8) {
    /// The ADX-Type Filter value (non-negative).
    value = 1,
};

/// Specifies the bipolar momentum the ADX-Type Filter is applied to.
pub const AdxTypeFilterSource = enum(u8) {
    /// The TSI numerator C - C[q-1] (named instance TSI_ATF).
    tsi_momentum = 0,
    /// The SMI raw stochastic momentum C - 0.5*(HH(q) + LL(q)) (named instance SMI_ATF).
    smi_momentum = 1,
    /// The DTI numerator max(H - H[q-1], 0) - max(L[q-1] - L, 0).
    dti_momentum = 2,
    /// The TVI tick balance upticks - downticks (2C - H - L for a bar).
    tvi_balance = 3,
    /// The single-smoothed normalized TSI(q, r, 1, 1), replacing the inner EMA.
    tsi_normalized = 4,
};

/// Parameters to create an ADX-Type Filter indicator.
///
/// The field names q, r and s are the canonical symbols from William Blau's
/// Momentum, Direction, and Divergence (Wiley, 1995), Appendix B, Figure B-24.
pub const AdxTypeFilterParams = struct {
    /// The bipolar momentum the filter is applied to. Default tsi_momentum.
    source: AdxTypeFilterSource = .tsi_momentum,
    /// The momentum look-back period. Zero selects the source default: 2 for
    /// tsi_momentum, dti_momentum and tsi_normalized, 32 for smi_momentum. Not
    /// used by tvi_balance. Default 0.
    q: usize = 0,
    /// The period of the inner EMA (for tsi_normalized, the smoothing period of the
    /// normalized TSI, which replaces the inner EMA). Must be > 0. Default 32.
    r: usize = 32,
    /// The period of the outer EMA. Must be > 0. Default 32.
    s: usize = 32,
    /// Bar component to extract (TSI sources only). `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract (TSI sources only). `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract (TSI sources only). `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
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

/// A rolling window holding at most `buffer.len` values.
const Window = struct {
    buffer: []f64,
    count: usize = 0,

    fn push(self: *Window, sample: f64) void {
        if (self.count < self.buffer.len) {
            self.buffer[self.count] = sample;
            self.count += 1;
            return;
        }

        std.mem.copyForwards(f64, self.buffer[0 .. self.buffer.len - 1], self.buffer[1..]);
        self.buffer[self.buffer.len - 1] = sample;
    }

    fn isFull(self: *const Window) bool {
        return self.count == self.buffer.len;
    }
};

/// ADX-Type Filter (ATF) by William Blau.
///
/// A non-negative trend-strength filter, analogous to Wilder's ADX, built by
/// rectifying and double-smoothing a bipolar momentum series (book Fig. B-24):
///
///   ATF(Price, r, s) = EMA(|EMA(Price, r)|, s)
///
/// The inner EMA(r) smooths the signed momentum, the absolute value discards the
/// direction and keeps the amplitude, and the outer EMA(s) smooths the amplitude.
/// A rising ATF signals a strengthening trend, a falling ATF a ranging market.
///
/// The bipolar momentum is selected by the source:
///   - tsi_momentum:   C - C[q-1]                               (TSI_ATF);
///   - smi_momentum:   C - 0.5*(HH(q) + LL(q))                  (SMI_ATF);
///   - dti_momentum:   max(H - H[q-1], 0) - max(L[q-1] - L, 0);
///   - tvi_balance:    upticks - downticks (2C - H - L for a bar);
///   - tsi_normalized: TSI(q, r, 1, 1), which replaces the inner EMA (r = 1).
///
/// Priming: each EMA stage seeds on its first finite momentum value. A NaN momentum
/// (the q-bar look-back warm-up) is propagated: the output is NaN and the EMAs do not
/// advance. The output is always >= 0.
pub const AdxTypeFilter = struct {
    source: AdxTypeFilterSource,
    uses_components: bool,

    // Rolling windows of the last q values (for the q-bar look-back).
    closes: Window,
    highs: Window,
    lows: Window,

    // Tick-rule state for single-valued samples of the tvi_balance source.
    previous: f64,
    has_previous: bool,

    // The normalized TSI (tsi_normalized only) replaces the inner EMA.
    tsi: ?tsi_mod.TrueStrengthIndex,
    inner: Ema,
    outer: Ema,

    primed: bool,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: AdxTypeFilterParams) !AdxTypeFilter {
        const source = params.source;
        const q = if (params.q != 0) params.q else if (source == .smi_momentum) @as(usize, 32) else @as(usize, 2);
        const r = params.r;
        const s = params.s;

        if (r < 1) return error.InvalidR;
        if (s < 1) return error.InvalidS;

        // Price components are meaningful only for the single-price TSI sources; the
        // other sources use the bar's high/low/close or the single value of the
        // sample (scalar value, quote mid price, trade price).
        const uses_components = source == .tsi_momentum or source == .tsi_normalized;

        var bc = bar_component.default_bar_component;
        var qc = quote_component.default_quote_component;
        var tc = trade_component.default_trade_component;
        if (uses_components) {
            bc = params.bar_component orelse bc;
            qc = params.quote_component orelse qc;
            tc = params.trade_component orelse tc;
        }

        const name = switch (source) {
            .tsi_momentum => "tsi",
            .smi_momentum => "smi",
            .dti_momentum => "dti",
            .tvi_balance => "tvi",
            .tsi_normalized => "tsin",
        };

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = (if (source == .tvi_balance)
            std.fmt.bufPrint(&mnemonic_buf, "atf.{s}({d},{d})", .{ name, r, s })
        else
            std.fmt.bufPrint(&mnemonic_buf, "atf.{s}({d},{d},{d}{s})", .{ name, q, r, s, triple })) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "ADX-Type Filter {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const closes = try allocator.alloc(f64, q);
        errdefer allocator.free(closes);
        const highs = try allocator.alloc(f64, q);
        errdefer allocator.free(highs);
        const lows = try allocator.alloc(f64, q);
        errdefer allocator.free(lows);

        const tsi: ?tsi_mod.TrueStrengthIndex = if (source == .tsi_normalized)
            try tsi_mod.TrueStrengthIndex.init(allocator, .{ .q = q, .r = r, .s = 1, .u = 1, .ul = 1 })
        else
            null;

        return .{
            .source = source,
            .uses_components = uses_components,
            .closes = .{ .buffer = closes },
            .highs = .{ .buffer = highs },
            .lows = .{ .buffer = lows },
            .previous = 0.0,
            .has_previous = false,
            .tsi = tsi,
            .inner = Ema.init(if (source == .tsi_normalized) 1 else r),
            .outer = Ema.init(s),
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *AdxTypeFilter) void {
        self.allocator.free(self.closes.buffer);
        self.allocator.free(self.highs.buffer);
        self.allocator.free(self.lows.buffer);
        if (self.tsi) |*tsi| tsi.deinit();
    }

    pub fn fixSlices(self: *AdxTypeFilter) void {
        // ATF doesn't use LineIndicator; mnemonic/description are read from the
        // owned buffers directly, so only the wrapped TSI needs a fixup.
        if (self.tsi) |*tsi| tsi.fixSlices();
    }

    /// Inner smooth -> rectify -> outer smooth; a NaN momentum propagates without
    /// advancing the EMAs.
    fn filter(self: *AdxTypeFilter, momentum: f64) f64 {
        if (math.isNan(momentum)) return math.nan(f64);

        self.primed = true;

        return self.outer.update(@abs(self.inner.update(momentum)));
    }

    /// Updates the indicator given the next single sample value.
    ///
    /// The smi_momentum and dti_momentum sources use the value as the high, the low
    /// and the close; the tvi_balance source applies the tick rule
    /// (balance = value - previous value).
    pub fn update(self: *AdxTypeFilter, sample: f64) f64 {
        switch (self.source) {
            .tsi_momentum => {
                self.closes.push(sample);
                if (!self.closes.isFull()) return math.nan(f64);

                // mtm_k = C_k - C_(k-(q-1)); the leftmost element is C_(k-(q-1)).
                return self.filter(sample - self.closes.buffer[0]);
            },
            .tsi_normalized => {
                const tsi = self.tsi.?.updateValues(sample).tsi;
                return self.filter(tsi);
            },
            .tvi_balance => {
                const balance = if (self.has_previous) sample - self.previous else 0.0;
                self.previous = sample;
                self.has_previous = true;
                return self.filter(balance);
            },
            else => return self.updateHighLowClose(sample, sample, sample),
        }
    }

    /// Updates the indicator given the next bar's high, low and close.
    ///
    /// The tsi_momentum and tsi_normalized sources use the close only.
    pub fn updateHighLowClose(self: *AdxTypeFilter, high: f64, low: f64, close: f64) f64 {
        switch (self.source) {
            .smi_momentum => {
                self.highs.push(high);
                self.lows.push(low);
                if (!self.highs.isFull()) return math.nan(f64);

                // sm = C - 0.5*(HH(q) + LL(q)).
                const hh = std.mem.max(f64, self.highs.buffer);
                const ll = std.mem.min(f64, self.lows.buffer);
                return self.filter(close - 0.5 * (hh + ll));
            },
            .dti_momentum => {
                self.highs.push(high);
                self.lows.push(low);
                if (!self.highs.isFull()) return math.nan(f64);

                // HMU - LMD; the leftmost elements are H_(k-(q-1)) and L_(k-(q-1)).
                const hmu = @max(high - self.highs.buffer[0], 0.0);
                const lmd = @max(self.lows.buffer[0] - low, 0.0);
                return self.filter(hmu - lmd);
            },
            // up - down = (C - L) - (H - C) = 2C - H - L.
            .tvi_balance => return self.filter(2.0 * close - high - low),
            else => return self.update(close),
        }
    }

    /// Returns whether the indicator has accumulated enough data.
    pub fn isPrimed(self: *const AdxTypeFilter) bool {
        return self.primed;
    }

    fn mnemonic(self: *const AdxTypeFilter) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const AdxTypeFilter) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    /// Returns metadata for this indicator.
    pub fn getMetadata(self: *const AdxTypeFilter, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        build_metadata_mod.buildMetadata(
            out,
            .adx_type_filter,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    pub fn updateScalar(self: *AdxTypeFilter, sample: *const Scalar) OutputArray {
        return wrap(sample.time, self.update(sample.value));
    }

    /// The TSI sources use the bar component; the other sources use the bar's high,
    /// low and close.
    pub fn updateBar(self: *AdxTypeFilter, sample: *const Bar) OutputArray {
        const value = if (self.uses_components)
            self.update(self.bar_func(sample.*))
        else
            self.updateHighLowClose(sample.high, sample.low, sample.close);
        return wrap(sample.time, value);
    }

    pub fn updateQuote(self: *AdxTypeFilter, sample: *const Quote) OutputArray {
        return wrap(sample.time, self.update(self.quote_func(sample.*)));
    }

    pub fn updateTrade(self: *AdxTypeFilter, sample: *const Trade) OutputArray {
        return wrap(sample.time, self.update(self.trade_func(sample.*)));
    }

    fn wrap(time: i64, value: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = value } });
        return out;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *AdxTypeFilter) indicator_mod.Indicator {
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
        const self: *AdxTypeFilter = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const AdxTypeFilter = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AdxTypeFilter = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AdxTypeFilter = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AdxTypeFilter = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AdxTypeFilter = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidR,
        InvalidS,
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
    source: AdxTypeFilterSource,
    q: usize,
    r: usize,
    s: usize,
    expected: [252]f64,
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }

    try testing.expect(@abs(act - exp) <= tolerance);
}

fn testBar(i: usize, input: *const [252]f64, high: *const [252]f64, low: *const [252]f64) Bar {
    return .{
        .time = @intCast(i),
        .open = input[i],
        .high = high[i],
        .low = low[i],
        .close = input[i],
        .volume = 0.0,
    };
}

const test_tolerance = 1e-13;

test "ATF reference data close sources" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    const combos = [_]Combo{
        .{ .source = .tsi_momentum, .q = 2, .r = 32, .s = 32, .expected = testdata.expectedTSIMTM_Q2_R32_S32() },
        .{ .source = .tsi_momentum, .q = 2, .r = 20, .s = 5, .expected = testdata.expectedTSIMTM_Q2_R20_S5() },
        .{ .source = .tsi_momentum, .q = 2, .r = 13, .s = 1, .expected = testdata.expectedTSIMTM_Q2_R13_S1() },
        .{ .source = .tsi_momentum, .q = 2, .r = 1, .s = 1, .expected = testdata.expectedTSIMTM_Q2_R1_S1() },
        .{ .source = .tsi_momentum, .q = 5, .r = 32, .s = 32, .expected = testdata.expectedTSIMTM_Q5_R32_S32() },
        .{ .source = .tsi_normalized, .q = 2, .r = 32, .s = 32, .expected = testdata.expectedTSINORM_R32_S32() },
        .{ .source = .tsi_normalized, .q = 2, .r = 20, .s = 20, .expected = testdata.expectedTSINORM_R20_S20() },
    };

    for (combos) |combo| {
        var ind = try AdxTypeFilter.init(allocator, .{ .source = combo.source, .q = combo.q, .r = combo.r, .s = combo.s });
        defer ind.deinit();

        for (0..252) |i| {
            try checkVal(combo.expected[i], ind.update(input[i]), test_tolerance);
        }
    }
}

const bar_combos = [_]Combo{
    .{ .source = .smi_momentum, .q = 32, .r = 32, .s = 32, .expected = testdata.expectedSMIRAW_Q32_R32_S32() },
    .{ .source = .smi_momentum, .q = 32, .r = 20, .s = 5, .expected = testdata.expectedSMIRAW_Q32_R20_S5() },
    .{ .source = .smi_momentum, .q = 5, .r = 32, .s = 32, .expected = testdata.expectedSMIRAW_Q5_R32_S32() },
    .{ .source = .dti_momentum, .q = 2, .r = 32, .s = 32, .expected = testdata.expectedDTINUM_Q2_R32_S32() },
    .{ .source = .dti_momentum, .q = 2, .r = 28, .s = 28, .expected = testdata.expectedDTINUM_Q2_R28_S28() },
    .{ .source = .dti_momentum, .q = 5, .r = 32, .s = 32, .expected = testdata.expectedDTINUM_Q5_R32_S32() },
    .{ .source = .tvi_balance, .q = 0, .r = 32, .s = 32, .expected = testdata.expectedTVI_R32_S32() },
    .{ .source = .tvi_balance, .q = 0, .r = 12, .s = 12, .expected = testdata.expectedTVI_R12_S12() },
    .{ .source = .tvi_balance, .q = 0, .r = 1, .s = 1, .expected = testdata.expectedTVI_R1_S1() },
};

test "ATF reference data bar sources" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    for (bar_combos) |combo| {
        var ind = try AdxTypeFilter.init(allocator, .{ .source = combo.source, .q = combo.q, .r = combo.r, .s = combo.s });
        defer ind.deinit();

        for (0..252) |i| {
            const out = ind.updateBar(&testBar(i, &input, &high, &low));
            try testing.expectEqual(@as(usize, 1), out.slice().len);
            try checkVal(combo.expected[i], out.slice()[0].scalar.value, test_tolerance);
        }
    }
}

test "ATF reference data bar sources from high, low and close" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    for (bar_combos) |combo| {
        var ind = try AdxTypeFilter.init(allocator, .{ .source = combo.source, .q = combo.q, .r = combo.r, .s = combo.s });
        defer ind.deinit();

        for (0..252) |i| {
            try checkVal(combo.expected[i], ind.updateHighLowClose(high[i], low[i], input[i]), test_tolerance);
        }
    }
}

test "ATF passthrough equals the absolute momentum" {
    const allocator = testing.allocator;

    var tsi = try AdxTypeFilter.init(allocator, .{ .source = .tsi_momentum, .q = 2, .r = 1, .s = 1 });
    defer tsi.deinit();
    try testing.expect(math.isNan(tsi.update(10.0)));
    try testing.expectEqual(@as(f64, 2.0), tsi.update(12.0));
    try testing.expectEqual(@as(f64, 1.0), tsi.update(11.0));

    var smi = try AdxTypeFilter.init(allocator, .{ .source = .smi_momentum, .q = 1, .r = 1, .s = 1 });
    defer smi.deinit();
    try testing.expectEqual(@as(f64, 0.5), smi.updateHighLowClose(11.0, 9.0, 10.5));

    var tvi = try AdxTypeFilter.init(allocator, .{ .source = .tvi_balance, .r = 1, .s = 1 });
    defer tvi.deinit();
    try testing.expectEqual(@as(f64, 0.0), tvi.update(10.0));
    try testing.expectEqual(@as(f64, 2.0), tvi.update(12.0));
    try testing.expectEqual(@as(f64, 3.0), tvi.update(9.0));
}

test "ATF warm-up region" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    var smi = try AdxTypeFilter.init(allocator, .{ .source = .smi_momentum });
    defer smi.deinit();

    var tvi = try AdxTypeFilter.init(allocator, .{ .source = .tvi_balance });
    defer tvi.deinit();

    for (0..252) |i| {
        try testing.expectEqual(i < 31, math.isNan(smi.updateHighLowClose(high[i], low[i], input[i])));
        try testing.expect(!math.isNan(tvi.updateHighLowClose(high[i], low[i], input[i])));
    }
}

test "ATF values are non-negative" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();

    for ([_]AdxTypeFilterSource{ .tsi_momentum, .smi_momentum, .dti_momentum, .tvi_balance, .tsi_normalized }) |source| {
        var ind = try AdxTypeFilter.init(allocator, .{ .source = source });
        defer ind.deinit();

        for (0..252) |i| {
            const value = ind.updateBar(&testBar(i, &input, &high, &low)).slice()[0].scalar.value;
            if (!math.isNan(value)) {
                try testing.expect(value >= 0.0);
            }
        }
    }
}

test "ATF is primed from the first finite momentum" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const q = 5;

    var ind = try AdxTypeFilter.init(allocator, .{ .source = .tsi_momentum, .q = q });
    defer ind.deinit();

    for (0..q - 1) |i| {
        _ = ind.update(input[i]);
        try testing.expect(!ind.isPrimed());
    }

    for (q - 1..252) |i| {
        _ = ind.update(input[i]);
        try testing.expect(ind.isPrimed());
    }
}

test "ATF entity updates" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const high = testdata.testHigh();
    const low = testdata.testLow();
    const expected = testdata.expectedTSIMTM_Q2_R32_S32();

    var scalar_ind = try AdxTypeFilter.init(allocator, .{});
    defer scalar_ind.deinit();
    var bar_ind = try AdxTypeFilter.init(allocator, .{});
    defer bar_ind.deinit();
    var quote_ind = try AdxTypeFilter.init(allocator, .{});
    defer quote_ind.deinit();
    var trade_ind = try AdxTypeFilter.init(allocator, .{});
    defer trade_ind.deinit();

    for (0..252) |i| {
        const time: i64 = @intCast(i);

        const scalar_out = scalar_ind.updateScalar(&Scalar{ .time = time, .value = input[i] });
        try testing.expectEqual(@as(usize, 1), scalar_out.slice().len);
        try checkVal(expected[i], scalar_out.slice()[0].scalar.value, test_tolerance);

        const bar_out = bar_ind.updateBar(&testBar(i, &input, &high, &low));
        try checkVal(expected[i], bar_out.slice()[0].scalar.value, test_tolerance);

        const quote_out = quote_ind.updateQuote(&Quote{
            .time = time,
            .bid_price = input[i],
            .bid_size = 0.0,
            .ask_price = input[i],
            .ask_size = 0.0,
        });
        try checkVal(expected[i], quote_out.slice()[0].scalar.value, test_tolerance);

        const trade_out = trade_ind.updateTrade(&Trade{ .time = time, .price = input[i], .volume = 0.0 });
        try checkVal(expected[i], trade_out.slice()[0].scalar.value, test_tolerance);
    }
}

test "ATF SMI single values use the value as high, low and close" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const params = AdxTypeFilterParams{ .source = .smi_momentum, .q = 5 };

    var reference = try AdxTypeFilter.init(allocator, params);
    defer reference.deinit();
    var scalar_ind = try AdxTypeFilter.init(allocator, params);
    defer scalar_ind.deinit();
    var quote_ind = try AdxTypeFilter.init(allocator, params);
    defer quote_ind.deinit();
    var trade_ind = try AdxTypeFilter.init(allocator, params);
    defer trade_ind.deinit();

    for (0..252) |i| {
        const v = input[i];
        const exp = reference.updateHighLowClose(v, v, v);

        try checkVal(exp, scalar_ind.updateScalar(&Scalar{ .time = 0, .value = v }).slice()[0].scalar.value, 0.0);
        try checkVal(exp, quote_ind.updateQuote(&Quote{
            .time = 0,
            .bid_price = v,
            .bid_size = 0.0,
            .ask_price = v,
            .ask_size = 0.0,
        }).slice()[0].scalar.value, 0.0);
        try checkVal(exp, trade_ind.updateTrade(&Trade{ .time = 0, .price = v, .volume = 0.0 }).slice()[0].scalar.value, 0.0);
    }
}

test "ATF metadata default" {
    const allocator = testing.allocator;

    var ind = try AdxTypeFilter.init(allocator, .{});
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.adx_type_filter, meta.identifier);
    try testing.expectEqualStrings("atf.tsi(2,32,32)", meta.mnemonic);
    try testing.expectEqualStrings("ADX-Type Filter atf.tsi(2,32,32)", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "ATF mnemonics" {
    const allocator = testing.allocator;

    const Case = struct {
        params: AdxTypeFilterParams,
        expected: []const u8,
    };

    const cases = [_]Case{
        .{ .params = .{}, .expected = "atf.tsi(2,32,32)" },
        .{ .params = .{ .source = .smi_momentum }, .expected = "atf.smi(32,32,32)" },
        .{ .params = .{ .source = .dti_momentum }, .expected = "atf.dti(2,32,32)" },
        .{ .params = .{ .source = .tvi_balance }, .expected = "atf.tvi(32,32)" },
        .{ .params = .{ .source = .tsi_normalized }, .expected = "atf.tsin(2,32,32)" },
        .{ .params = .{ .source = .smi_momentum, .q = 5, .r = 20, .s = 5 }, .expected = "atf.smi(5,20,5)" },
        .{ .params = .{ .bar_component = .median }, .expected = "atf.tsi(2,32,32, hl/2)" },
        .{ .params = .{ .quote_component = .bid }, .expected = "atf.tsi(2,32,32, b)" },
        .{ .params = .{ .trade_component = .volume }, .expected = "atf.tsi(2,32,32, v)" },
        .{ .params = .{ .bar_component = .open, .quote_component = .bid }, .expected = "atf.tsi(2,32,32, o, b)" },
        .{ .params = .{ .bar_component = .high, .trade_component = .volume }, .expected = "atf.tsi(2,32,32, h, v)" },
        .{ .params = .{ .quote_component = .ask, .trade_component = .volume }, .expected = "atf.tsi(2,32,32, a, v)" },
        .{ .params = .{ .source = .tsi_normalized, .bar_component = .median }, .expected = "atf.tsin(2,32,32, hl/2)" },
        .{ .params = .{ .source = .smi_momentum, .bar_component = .median }, .expected = "atf.smi(32,32,32)" },
    };

    for (cases) |case| {
        var ind = try AdxTypeFilter.init(allocator, case.params);
        defer ind.deinit();

        try testing.expectEqualStrings(case.expected, ind.mnemonic());
    }
}

test "ATF invalid params" {
    const allocator = testing.allocator;

    try testing.expectError(error.InvalidR, AdxTypeFilter.init(allocator, .{ .r = 0 }));
    try testing.expectError(error.InvalidS, AdxTypeFilter.init(allocator, .{ .s = 0 }));
}
