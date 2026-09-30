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
const line_indicator_mod = @import("../../core/line_indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");
const tsi_mod = @import("../true_strength_index/true_strength_index.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Slope Divergence TSI Filter indicator.
pub const SlopeDivergenceTsiFilterOutput = enum(u8) {
    /// The Slope Divergence TSI Filter value (range [-100, +100]).
    value = 1,
};

/// Parameters to create a Slope Divergence TSI Filter indicator.
///
/// The field names q, r, s, u, x and y are the canonical symbols from William
/// Blau's Momentum, Direction, and Divergence (Wiley, 1995), chapter 12 and
/// Appendix B, Figure B-25.
pub const SlopeDivergenceTsiFilterParams = struct {
    /// The TSI momentum look-back period; momentum is C_k - C_(k-(q-1)). Must be > 0. Default 2.
    q: usize = 2,
    /// The period of the 1st (innermost) EMA of the TSI smoothing cascade. Must be > 0. Default 32.
    r: usize = 32,
    /// The period of the 2nd EMA of the TSI smoothing cascade. Must be > 0. Default 32.
    s: usize = 32,
    /// The period of the 3rd (outermost) EMA of the TSI smoothing cascade. Must be > 0. Default 7.
    u: usize = 7,
    /// The period of the 1st EMA of the price reference DEMA(close, x, y). Must be > 0. Default 32.
    x: usize = 32,
    /// The period of the 2nd EMA of the price reference DEMA(close, x, y). Must be > 0. Default 7.
    y: usize = 7,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
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

/// Slope Divergence TSI Filter (SD_TSI) by William Blau.
///
/// A trend/congestion prefilter built on the True Strength Index. It keeps the
/// TSI value only when the slope of the TSI agrees in sign with the slope of a
/// separate double EMA of price; otherwise it outputs 0 (a slope divergence, or
/// congestion zone):
///
///   ind_k = TSI(close, q, r, s, u)_k
///   ref_k = DEMA(close, x, y)_k = EMA(EMA(close, x), y)_k
///
///   SD_TSI_k = ind_k   if ind_k - ind_(k-1) > 0 and ref_k - ref_(k-1) > 0
///            = ind_k   if ind_k - ind_(k-1) < 0 and ref_k - ref_(k-1) < 0
///            = 0       otherwise
///
/// The gate is strict (book Fig. B-25): a flat slope on either series yields 0.
/// The output range is [-100, +100].
///
/// Priming: the TSI is NaN for bars 0..q-2 (momentum look-back), so SD_TSI is NaN
/// there. The price DEMA seeds at bar 0 and advances every bar, including through
/// the TSI warm-up. At the first finite TSI bar there is no prior TSI value, hence
/// no slope, so the output is 0.0.
pub const SlopeDivergenceTsiFilter = struct {
    line: LineIndicator,

    // The TSI oscillator (its signal line is unused, so ul=1).
    tsi: tsi_mod.TrueStrengthIndex,

    // Price reference: DEMA(close, x, y) = EMA(EMA(close, x), y).
    reference_x: Ema,
    reference_y: Ema,

    // Slope state: the previous finite TSI and the previous-bar reference.
    previous_tsi: f64,
    has_previous_tsi: bool,
    previous_reference: f64,

    primed: bool,

    // Fixed buffers for mnemonic and description strings.
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: SlopeDivergenceTsiFilterParams) !SlopeDivergenceTsiFilter {
        const q = params.q;
        const r = params.r;
        const s = params.s;
        const u = params.u;
        const x = params.x;
        const y = params.y;

        if (q < 1) return error.InvalidQ;
        if (r < 1) return error.InvalidR;
        if (s < 1) return error.InvalidS;
        if (u < 1) return error.InvalidU;
        if (x < 1) return error.InvalidX;
        if (y < 1) return error.InvalidY;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "sdtsi({d},{d},{d},{d},{d},{d}{s})", .{ q, r, s, u, x, y, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Slope Divergence TSI Filter {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const tsi = try tsi_mod.TrueStrengthIndex.init(allocator, .{ .q = q, .r = r, .s = s, .u = u, .ul = 1 });

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .tsi = tsi,
            .reference_x = Ema.init(x),
            .reference_y = Ema.init(y),
            .previous_tsi = 0.0,
            .has_previous_tsi = false,
            .previous_reference = 0.0,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *SlopeDivergenceTsiFilter) void {
        self.tsi.deinit();
    }

    /// After init, fix up the line's mnemonic/description slices to point into
    /// `self`'s own buffers (not the stack-local ones from `init`).
    pub fn fixSlices(self: *SlopeDivergenceTsiFilter) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
        self.tsi.fixSlices();
    }

    /// Core update logic. Returns the SD_TSI value or NaN during the TSI warm-up.
    pub fn update(self: *SlopeDivergenceTsiFilter, sample: f64) f64 {
        // The price reference advances every bar (it has no NaN warm-up).
        const reference = self.reference_y.update(self.reference_x.update(sample));

        const tsi = self.tsi.updateValues(sample).tsi;
        if (math.isNan(tsi)) {
            // TSI momentum warm-up: keep the previous-bar reference current.
            self.previous_reference = reference;
            return math.nan(f64);
        }

        var result: f64 = 0.0;

        // At the first finite TSI there is no prior TSI, hence no slope.
        if (self.has_previous_tsi) {
            const delta_tsi = tsi - self.previous_tsi;
            const delta_reference = reference - self.previous_reference;

            // Keep the TSI only when both slopes are strictly same-signed.
            if ((delta_tsi > 0.0 and delta_reference > 0.0) or (delta_tsi < 0.0 and delta_reference < 0.0)) {
                result = tsi;
            }
        }

        self.previous_tsi = tsi;
        self.has_previous_tsi = true;
        self.previous_reference = reference;
        self.primed = true;

        return result;
    }

    /// Returns whether the indicator has accumulated enough data.
    pub fn isPrimed(self: *const SlopeDivergenceTsiFilter) bool {
        return self.primed;
    }

    /// Returns metadata for this indicator.
    pub fn getMetadata(self: *const SlopeDivergenceTsiFilter, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .slope_divergence_tsi_filter,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *SlopeDivergenceTsiFilter, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *SlopeDivergenceTsiFilter, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *SlopeDivergenceTsiFilter, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *SlopeDivergenceTsiFilter, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *SlopeDivergenceTsiFilter) indicator_mod.Indicator {
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
        const self: *SlopeDivergenceTsiFilter = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const SlopeDivergenceTsiFilter = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *SlopeDivergenceTsiFilter = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *SlopeDivergenceTsiFilter = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *SlopeDivergenceTsiFilter = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *SlopeDivergenceTsiFilter = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidQ,
        InvalidR,
        InvalidS,
        InvalidU,
        InvalidX,
        InvalidY,
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
    r: usize,
    s: usize,
    u: usize,
    x: usize,
    y: usize,
    expected: [252]f64,
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }

    try testing.expect(@abs(act - exp) <= tolerance);
}

test "SD_TSI reference data all combos" {
    const allocator = testing.allocator;
    const tolerance = 1e-13;
    const input = testdata.testInput();

    // Every expected array uses the book momentum q=2.
    const combos = [_]Combo{
        .{ .name = "R32_S32_U7_X32_Y7", .r = 32, .s = 32, .u = 7, .x = 32, .y = 7, .expected = testdata.expectedR32_S32_U7_X32_Y7() },
        .{ .name = "R32_S32_U1_X32_Y1", .r = 32, .s = 32, .u = 1, .x = 32, .y = 1, .expected = testdata.expectedR32_S32_U1_X32_Y1() },
        .{ .name = "R32_S32_U7_X32_Y1", .r = 32, .s = 32, .u = 7, .x = 32, .y = 1, .expected = testdata.expectedR32_S32_U7_X32_Y1() },
        .{ .name = "R32_S32_U1_X32_Y7", .r = 32, .s = 32, .u = 1, .x = 32, .y = 7, .expected = testdata.expectedR32_S32_U1_X32_Y7() },
        .{ .name = "R1_S1_U1_X1_Y1", .r = 1, .s = 1, .u = 1, .x = 1, .y = 1, .expected = testdata.expectedR1_S1_U1_X1_Y1() },
        .{ .name = "R20_S5_U3_X20_Y3", .r = 20, .s = 5, .u = 3, .x = 20, .y = 3, .expected = testdata.expectedR20_S5_U3_X20_Y3() },
        .{ .name = "R32_S13_U3_X32_Y7", .r = 32, .s = 13, .u = 3, .x = 32, .y = 7, .expected = testdata.expectedR32_S13_U3_X32_Y7() },
        .{ .name = "R12_S12_U1_X12_Y1", .r = 12, .s = 12, .u = 1, .x = 12, .y = 1, .expected = testdata.expectedR12_S12_U1_X12_Y1() },
        .{ .name = "R25_S13_U1_X25_Y1", .r = 25, .s = 13, .u = 1, .x = 25, .y = 1, .expected = testdata.expectedR25_S13_U1_X25_Y1() },
        .{ .name = "R64_S64_U7_X32_Y7", .r = 64, .s = 64, .u = 7, .x = 32, .y = 7, .expected = testdata.expectedR64_S64_U7_X32_Y7() },
        .{ .name = "R32_S32_U7_X16_Y3", .r = 32, .s = 32, .u = 7, .x = 16, .y = 3, .expected = testdata.expectedR32_S32_U7_X16_Y3() },
        .{ .name = "R5_S5_U5_X5_Y5", .r = 5, .s = 5, .u = 5, .x = 5, .y = 5, .expected = testdata.expectedR5_S5_U5_X5_Y5() },
        .{ .name = "R10_S10_U1_X10_Y1", .r = 10, .s = 10, .u = 1, .x = 10, .y = 1, .expected = testdata.expectedR10_S10_U1_X10_Y1() },
        .{ .name = "R40_S20_U5_X32_Y7", .r = 40, .s = 20, .u = 5, .x = 32, .y = 7, .expected = testdata.expectedR40_S20_U5_X32_Y7() },
        .{ .name = "R32_S5_U1_X32_Y1", .r = 32, .s = 5, .u = 1, .x = 32, .y = 1, .expected = testdata.expectedR32_S5_U1_X32_Y1() },
        .{ .name = "R50_S25_U1_X50_Y1", .r = 50, .s = 25, .u = 1, .x = 50, .y = 1, .expected = testdata.expectedR50_S25_U1_X50_Y1() },
    };

    for (combos) |combo| {
        var ind = try SlopeDivergenceTsiFilter.init(allocator, .{
            .q = 2,
            .r = combo.r,
            .s = combo.s,
            .u = combo.u,
            .x = combo.x,
            .y = combo.y,
        });
        defer ind.deinit();

        for (0..252) |i| {
            try checkVal(combo.expected[i], ind.update(input[i]), tolerance);
        }
    }
}

test "SD_TSI passthrough keeps +/-100 only when the slopes agree" {
    const allocator = testing.allocator;

    var ind = try SlopeDivergenceTsiFilter.init(allocator, .{ .q = 2, .r = 1, .s = 1, .u = 1, .x = 1, .y = 1 });
    defer ind.deinit();

    try testing.expect(math.isNan(ind.update(10.0))); // momentum undefined
    try testing.expectEqual(@as(f64, 0.0), ind.update(12.0)); // first finite TSI, no slope
    try testing.expectEqual(@as(f64, -100.0), ind.update(11.0)); // both falling
    try testing.expectEqual(@as(f64, 100.0), ind.update(13.0)); // both rising
    try testing.expectEqual(@as(f64, 0.0), ind.update(14.0)); // TSI flat
}

test "SD_TSI warm-up region" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const q = 5;

    var ind = try SlopeDivergenceTsiFilter.init(allocator, .{ .q = q });
    defer ind.deinit();

    for (0..q - 1) |i| {
        try testing.expect(math.isNan(ind.update(input[i])));
    }

    try testing.expectEqual(@as(f64, 0.0), ind.update(input[q - 1]));

    for (q..252) |i| {
        try testing.expect(!math.isNan(ind.update(input[i])));
    }
}

test "SD_TSI values are bounded to [-100, 100]" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    var ind = try SlopeDivergenceTsiFilter.init(allocator, .{});
    defer ind.deinit();

    for (0..252) |i| {
        const value = ind.update(input[i]);

        if (!math.isNan(value)) {
            try testing.expect(value >= -100.0);
            try testing.expect(value <= 100.0);
        }
    }
}

test "SD_TSI is primed from the first finite TSI bar" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const q = 5;

    var ind = try SlopeDivergenceTsiFilter.init(allocator, .{ .q = q });
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

test "SD_TSI entity updates" {
    const allocator = testing.allocator;
    const tolerance = 1e-13;
    const input = testdata.testInput();
    const expected = testdata.expectedR32_S32_U7_X32_Y7();

    var scalar_ind = try SlopeDivergenceTsiFilter.init(allocator, .{});
    defer scalar_ind.deinit();
    scalar_ind.fixSlices();

    var bar_ind = try SlopeDivergenceTsiFilter.init(allocator, .{});
    defer bar_ind.deinit();
    bar_ind.fixSlices();

    var quote_ind = try SlopeDivergenceTsiFilter.init(allocator, .{});
    defer quote_ind.deinit();
    quote_ind.fixSlices();

    var trade_ind = try SlopeDivergenceTsiFilter.init(allocator, .{});
    defer trade_ind.deinit();
    trade_ind.fixSlices();

    for (0..252) |i| {
        const time: i64 = @intCast(i);

        const scalar_out = scalar_ind.updateScalar(&Scalar{ .time = time, .value = input[i] });
        try testing.expectEqual(@as(usize, 1), scalar_out.slice().len);
        try checkVal(expected[i], scalar_out.slice()[0].scalar.value, tolerance);

        const bar_out = bar_ind.updateBar(&Bar{
            .time = time,
            .open = input[i],
            .high = input[i],
            .low = input[i],
            .close = input[i],
            .volume = 0.0,
        });
        try checkVal(expected[i], bar_out.slice()[0].scalar.value, tolerance);

        const quote_out = quote_ind.updateQuote(&Quote{
            .time = time,
            .bid_price = input[i],
            .bid_size = 0.0,
            .ask_price = input[i],
            .ask_size = 0.0,
        });
        try checkVal(expected[i], quote_out.slice()[0].scalar.value, tolerance);

        const trade_out = trade_ind.updateTrade(&Trade{
            .time = time,
            .price = input[i],
            .volume = 0.0,
        });
        try checkVal(expected[i], trade_out.slice()[0].scalar.value, tolerance);
    }
}

test "SD_TSI metadata default" {
    const allocator = testing.allocator;

    var ind = try SlopeDivergenceTsiFilter.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.slope_divergence_tsi_filter, meta.identifier);
    try testing.expectEqualStrings("sdtsi(2,32,32,7,32,7)", meta.mnemonic);
    try testing.expectEqualStrings("Slope Divergence TSI Filter sdtsi(2,32,32,7,32,7)", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "SD_TSI mnemonic component combinations" {
    const allocator = testing.allocator;

    const Case = struct {
        bc: ?bar_component.BarComponent,
        qc: ?quote_component.QuoteComponent,
        tc: ?trade_component.TradeComponent,
        expected: []const u8,
    };

    const cases = [_]Case{
        .{ .bc = null, .qc = null, .tc = null, .expected = "sdtsi(2,32,32,7,32,7)" },
        .{ .bc = .median, .qc = null, .tc = null, .expected = "sdtsi(2,32,32,7,32,7, hl/2)" },
        .{ .bc = null, .qc = .bid, .tc = null, .expected = "sdtsi(2,32,32,7,32,7, b)" },
        .{ .bc = null, .qc = null, .tc = .volume, .expected = "sdtsi(2,32,32,7,32,7, v)" },
        .{ .bc = .open, .qc = .bid, .tc = null, .expected = "sdtsi(2,32,32,7,32,7, o, b)" },
        .{ .bc = .high, .qc = null, .tc = .volume, .expected = "sdtsi(2,32,32,7,32,7, h, v)" },
        .{ .bc = null, .qc = .ask, .tc = .volume, .expected = "sdtsi(2,32,32,7,32,7, a, v)" },
    };

    for (cases) |case| {
        var ind = try SlopeDivergenceTsiFilter.init(allocator, .{
            .bar_component = case.bc,
            .quote_component = case.qc,
            .trade_component = case.tc,
        });
        defer ind.deinit();
        ind.fixSlices();

        try testing.expectEqualStrings(case.expected, ind.line.mnemonic);
    }
}

test "SD_TSI custom mnemonic" {
    const allocator = testing.allocator;

    var ind = try SlopeDivergenceTsiFilter.init(allocator, .{ .q = 3, .r = 20, .s = 5, .u = 3, .x = 20, .y = 3 });
    defer ind.deinit();
    ind.fixSlices();

    try testing.expectEqualStrings("sdtsi(3,20,5,3,20,3)", ind.line.mnemonic);
}

test "SD_TSI invalid params" {
    const allocator = testing.allocator;

    try testing.expectError(error.InvalidQ, SlopeDivergenceTsiFilter.init(allocator, .{ .q = 0 }));
    try testing.expectError(error.InvalidR, SlopeDivergenceTsiFilter.init(allocator, .{ .r = 0 }));
    try testing.expectError(error.InvalidS, SlopeDivergenceTsiFilter.init(allocator, .{ .s = 0 }));
    try testing.expectError(error.InvalidU, SlopeDivergenceTsiFilter.init(allocator, .{ .u = 0 }));
    try testing.expectError(error.InvalidX, SlopeDivergenceTsiFilter.init(allocator, .{ .x = 0 }));
    try testing.expectError(error.InvalidY, SlopeDivergenceTsiFilter.init(allocator, .{ .y = 0 }));
}
