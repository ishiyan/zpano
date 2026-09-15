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

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Double-Smoothed Momenta indicator.
pub const DoubleSmoothedMomentaOutput = enum(u8) {
    /// The Double-Smoothed Momenta oscillator value (range [0, 100]).
    value = 1,
};

/// Parameters to create a Double-Smoothed Momenta indicator.
///
/// The field names a, y and z are the canonical symbols from William Blau's
/// double-smoothed momentum family.
pub const DoubleSmoothedMomentaParams = struct {
    /// The highest/lowest close look-back. Must be > 0. Default 2.
    a: usize = 2,
    /// The period of the inner (1st) smoothing EMA. Must be > 0. Default 2.
    y: usize = 2,
    /// The period of the outer (2nd) smoothing EMA. Must be > 0. Default 14.
    z: usize = 14,
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

/// Double-Smoothed Momenta (DM) by William Blau, also known as the
/// Double-Smoothed RSI (DRSI) when the look-back is fixed at 2.
///
/// A close-based, double-smoothed momentum oscillator bounded to [0, 100]:
///
///   LCa_k = min(Close over the last a bars)          (lowest close)
///   HCa_k = max(Close over the last a bars)          (highest close)
///   st_k  = Close_k - LCa_k                          (close above the low close)
///   rng_k = HCa_k - LCa_k                            (a-bar close range)
///
///   DM(a, y, z) = 100 * EMA(EMA(st, y), z) / EMA(EMA(rng, y), z)
///
/// Each of the numerator (st) and denominator (rng) series is double-smoothed by an
/// inner EMA of period y then an outer EMA of period z (Blau's Ez(Ey(.)) ), and the
/// ratio is scaled by 100.
///
/// This is structurally the Double-Smoothed Stochastic computed on the CLOSE -- it
/// uses the highest/lowest close over a bars instead of the high/low of the bar --
/// and it has no signal line, so it produces a single scalar output per bar.
///
/// Named instances:
///   - RSI equivalence:     DM(2, 1, z) == RSI(z), the EMA-form RSI;
///   - Double-smoothed RSI: DRSI(y, z) = DM(2, y, z).
///
/// Priming: st/rng are valid once a closes exist (bar a-1); all four EMA stages seed
/// there. DM is NaN for bars 0..a-2 and finite from bar a-1. For a == 1 there is no
/// NaN warm-up, but the a-bar close range is then always 0, so DM is 0.0 on every
/// bar via the guard (a degenerate setting).
///
/// Division guard: EMA(EMA(rng)) <= 0 -> DM = 0.0.
pub const DoubleSmoothedMomenta = struct {
    line: LineIndicator,

    // Rolling window of the last a closes (for the highest/lowest close).
    window: []f64,
    window_length: usize,
    window_count: usize,
    last_index: usize,

    // Two independent 2-stage EMA cascades (double smoothing), each wired
    // inner(y) -> outer(z): EMA(EMA(x, y), z).
    numerator_y: Ema,
    numerator_z: Ema,
    denominator_y: Ema,
    denominator_z: Ema,

    primed: bool,
    allocator: std.mem.Allocator,

    // Fixed buffers for mnemonic and description strings.
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: DoubleSmoothedMomentaParams) !DoubleSmoothedMomenta {
        const a = params.a;
        const y = params.y;
        const z = params.z;

        if (a < 1) return error.InvalidA;
        if (y < 1) return error.InvalidY;
        if (z < 1) return error.InvalidZ;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "dm({d},{d},{d}{s})", .{ a, y, z, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Double-Smoothed Momenta {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window = try allocator.alloc(f64, a);
        @memset(window, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .window = window,
            .window_length = a,
            .window_count = 0,
            .last_index = a - 1,
            .numerator_y = Ema.init(y),
            .numerator_z = Ema.init(z),
            .denominator_y = Ema.init(y),
            .denominator_z = Ema.init(z),
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *DoubleSmoothedMomenta) void {
        self.allocator.free(self.window);
    }

    /// After init, fix up the line's mnemonic/description slices to point into
    /// `self`'s own buffers (not the stack-local ones from `init`).
    pub fn fixSlices(self: *DoubleSmoothedMomenta) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update logic. Returns the DM value or NaN during the a-bar warm-up.
    pub fn update(self: *DoubleSmoothedMomenta, sample: f64) f64 {
        if (self.primed) {
            var i: usize = 0;
            while (i < self.last_index) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = sample;
        } else {
            self.window[self.window_count] = sample;
            self.window_count += 1;

            // Need a closes before the highest/lowest close is defined. While
            // unprimed, the EMA cascades must NOT advance (they seed at bar a-1).
            if (self.window_length > self.window_count) {
                return math.nan(f64);
            }

            self.primed = true;
        }

        // Highest/lowest close over the last a bars.
        var hc = self.window[0];
        var lc = self.window[0];

        var i: usize = 1;
        while (i < self.window_length) : (i += 1) {
            const v = self.window[i];

            if (v > hc) {
                hc = v;
            }

            if (v < lc) {
                lc = v;
            }
        }

        // Raw close-above-low (>= 0) and a-bar close range (>= 0).
        const st = sample - lc;
        const rng = hc - lc;

        // Double-smooth each separately (inner y, then outer z), then divide.
        const num = self.numerator_z.update(self.numerator_y.update(st));
        const den = self.denominator_z.update(self.denominator_y.update(rng));

        // Division guard: smoothed range <= 0 -> DM = 0.0.
        if (den <= 0.0) {
            return 0.0;
        }

        return 100.0 * num / den;
    }

    /// Returns whether the indicator has accumulated enough data.
    pub fn isPrimed(self: *const DoubleSmoothedMomenta) bool {
        return self.primed;
    }

    /// Returns metadata for this indicator.
    pub fn getMetadata(self: *const DoubleSmoothedMomenta, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .double_smoothed_momenta,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *DoubleSmoothedMomenta, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *DoubleSmoothedMomenta, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *DoubleSmoothedMomenta, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *DoubleSmoothedMomenta, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *DoubleSmoothedMomenta) indicator_mod.Indicator {
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
        const self: *DoubleSmoothedMomenta = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const DoubleSmoothedMomenta = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DoubleSmoothedMomenta = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DoubleSmoothedMomenta = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DoubleSmoothedMomenta = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DoubleSmoothedMomenta = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidA,
        InvalidY,
        InvalidZ,
        MnemonicTooLong,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

const Combo = struct {
    name: []const u8,
    a: usize,
    y: usize,
    z: usize,
    expected: [252]f64,
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }

    try testing.expect(@abs(act - exp) <= tolerance);
}

/// Independently-coded EMA-form RSI: 100 * EMA(up, z) / EMA(up + dn, z).
fn emaFormRsi(closes: []const f64, z: usize, out: []f64) void {
    const alpha = 2.0 / (@as(f64, @floatFromInt(z)) + 1.0);
    out[0] = math.nan(f64);

    var numerator: f64 = 0.0;
    var denominator: f64 = 0.0;
    var primed = false;

    var k: usize = 1;
    while (k < closes.len) : (k += 1) {
        const diff = closes[k] - closes[k - 1];
        const up: f64 = if (diff > 0.0) diff else 0.0;
        const dn: f64 = if (diff < 0.0) -diff else 0.0;

        if (primed) {
            numerator = alpha * up + (1.0 - alpha) * numerator;
            denominator = alpha * (up + dn) + (1.0 - alpha) * denominator;
        } else {
            numerator = up;
            denominator = up + dn;
            primed = true;
        }

        out[k] = if (denominator <= 0.0) 0.0 else 100.0 * numerator / denominator;
    }
}

test "DM reference data all combos" {
    const allocator = testing.allocator;
    const tolerance = 1e-9;
    const input = testdata.testInput();

    const combos = [_]Combo{
        .{ .name = "A2_Y2_Z14", .a = 2, .y = 2, .z = 14, .expected = testdata.expectedA2_Y2_Z14() },
        .{ .name = "A2_Y1_Z14", .a = 2, .y = 1, .z = 14, .expected = testdata.expectedA2_Y1_Z14() },
        .{ .name = "A2_Y1_Z9", .a = 2, .y = 1, .z = 9, .expected = testdata.expectedA2_Y1_Z9() },
        .{ .name = "A2_Y1_Z2", .a = 2, .y = 1, .z = 2, .expected = testdata.expectedA2_Y1_Z2() },
        .{ .name = "A2_Y3_Z9", .a = 2, .y = 3, .z = 9, .expected = testdata.expectedA2_Y3_Z9() },
        .{ .name = "A2_Y5_Z5", .a = 2, .y = 5, .z = 5, .expected = testdata.expectedA2_Y5_Z5() },
        .{ .name = "A2_Y2_Z5", .a = 2, .y = 2, .z = 5, .expected = testdata.expectedA2_Y2_Z5() },
        .{ .name = "A2_Y1_Z1", .a = 2, .y = 1, .z = 1, .expected = testdata.expectedA2_Y1_Z1() },
        .{ .name = "A1_Y1_Z1", .a = 1, .y = 1, .z = 1, .expected = testdata.expectedA1_Y1_Z1() },
        .{ .name = "A5_Y2_Z14", .a = 5, .y = 2, .z = 14, .expected = testdata.expectedA5_Y2_Z14() },
        .{ .name = "A10_Y3_Z5", .a = 10, .y = 3, .z = 5, .expected = testdata.expectedA10_Y3_Z5() },
        .{ .name = "A14_Y2_Z9", .a = 14, .y = 2, .z = 9, .expected = testdata.expectedA14_Y2_Z9() },
        .{ .name = "A20_Y5_Z3", .a = 20, .y = 5, .z = 3, .expected = testdata.expectedA20_Y5_Z3() },
        .{ .name = "A3_Y3_Z3", .a = 3, .y = 3, .z = 3, .expected = testdata.expectedA3_Y3_Z3() },
        .{ .name = "A7_Y4_Z2", .a = 7, .y = 4, .z = 2, .expected = testdata.expectedA7_Y4_Z2() },
        .{ .name = "A32_Y2_Z7", .a = 32, .y = 2, .z = 7, .expected = testdata.expectedA32_Y2_Z7() },
    };

    for (combos) |combo| {
        var ind = try DoubleSmoothedMomenta.init(allocator, .{
            .a = combo.a,
            .y = combo.y,
            .z = combo.z,
        });
        defer ind.deinit();

        for (0..252) |i| {
            try checkVal(combo.expected[i], ind.update(input[i]), tolerance);
        }
    }
}

test "DM warm-up region" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const a = 5;

    var ind = try DoubleSmoothedMomenta.init(allocator, .{ .a = a, .y = 2, .z = 14 });
    defer ind.deinit();

    for (0..a - 1) |i| {
        try testing.expect(math.isNan(ind.update(input[i])));
    }

    for (a - 1..252) |i| {
        try testing.expect(!math.isNan(ind.update(input[i])));
    }
}

test "DM has no warm-up when a = 1" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    var ind = try DoubleSmoothedMomenta.init(allocator, .{ .a = 1, .y = 1, .z = 1 });
    defer ind.deinit();

    for (0..252) |i| {
        try testing.expect(!math.isNan(ind.update(input[i])));
    }
}

test "DM values are bounded to [0, 100]" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    var ind = try DoubleSmoothedMomenta.init(allocator, .{});
    defer ind.deinit();

    for (0..252) |i| {
        const value = ind.update(input[i]);

        if (!math.isNan(value)) {
            try testing.expect(value >= 0.0);
            try testing.expect(value <= 100.0);
        }
    }
}

test "DM degenerate a = 1 is identically zero" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    var ind = try DoubleSmoothedMomenta.init(allocator, .{ .a = 1, .y = 1, .z = 1 });
    defer ind.deinit();

    for (0..252) |i| {
        try testing.expectEqual(@as(f64, 0.0), ind.update(input[i]));
    }
}

test "DM(2,1,z) equals the EMA-form RSI(z)" {
    const allocator = testing.allocator;
    const tolerance = 1e-9;
    const input = testdata.testInput();

    var expected: [252]f64 = undefined;

    for ([_]usize{ 1, 2, 9, 14 }) |z| {
        emaFormRsi(&input, z, &expected);

        var ind = try DoubleSmoothedMomenta.init(allocator, .{ .a = 2, .y = 1, .z = z });
        defer ind.deinit();

        for (0..252) |i| {
            try checkVal(expected[i], ind.update(input[i]), tolerance);
        }
    }
}

test "DM is primed once a closes have been seen" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const a = 5;

    var ind = try DoubleSmoothedMomenta.init(allocator, .{ .a = a, .y = 2, .z = 14 });
    defer ind.deinit();

    for (0..a - 1) |i| {
        _ = ind.update(input[i]);
        try testing.expect(!ind.isPrimed());
    }

    for (a - 1..252) |i| {
        _ = ind.update(input[i]);
        try testing.expect(ind.isPrimed());
    }
}

test "DM entity updates" {
    const allocator = testing.allocator;
    const tolerance = 1e-9;
    const input = testdata.testInput();
    const expected = testdata.expectedA2_Y2_Z14();

    var scalar_ind = try DoubleSmoothedMomenta.init(allocator, .{});
    defer scalar_ind.deinit();
    scalar_ind.fixSlices();

    var bar_ind = try DoubleSmoothedMomenta.init(allocator, .{});
    defer bar_ind.deinit();
    bar_ind.fixSlices();

    var quote_ind = try DoubleSmoothedMomenta.init(allocator, .{});
    defer quote_ind.deinit();
    quote_ind.fixSlices();

    var trade_ind = try DoubleSmoothedMomenta.init(allocator, .{});
    defer trade_ind.deinit();
    trade_ind.fixSlices();

    for (0..252) |i| {
        const time: i64 = @intCast(i);

        const scalar_out = scalar_ind.updateScalar(&Scalar{ .time = time, .value = input[i] });
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

test "DM metadata default" {
    const allocator = testing.allocator;

    var ind = try DoubleSmoothedMomenta.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.double_smoothed_momenta, meta.identifier);
    try testing.expectEqualStrings("dm(2,2,14)", meta.mnemonic);
    try testing.expectEqualStrings("Double-Smoothed Momenta dm(2,2,14)", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "DM mnemonic component combinations" {
    const allocator = testing.allocator;

    const Case = struct {
        bc: ?bar_component.BarComponent,
        qc: ?quote_component.QuoteComponent,
        tc: ?trade_component.TradeComponent,
        expected: []const u8,
    };

    const cases = [_]Case{
        .{ .bc = null, .qc = null, .tc = null, .expected = "dm(2,2,14)" },
        .{ .bc = .median, .qc = null, .tc = null, .expected = "dm(2,2,14, hl/2)" },
        .{ .bc = null, .qc = .bid, .tc = null, .expected = "dm(2,2,14, b)" },
        .{ .bc = null, .qc = null, .tc = .volume, .expected = "dm(2,2,14, v)" },
        .{ .bc = .open, .qc = .bid, .tc = null, .expected = "dm(2,2,14, o, b)" },
        .{ .bc = .high, .qc = null, .tc = .volume, .expected = "dm(2,2,14, h, v)" },
        .{ .bc = null, .qc = .ask, .tc = .volume, .expected = "dm(2,2,14, a, v)" },
    };

    for (cases) |case| {
        var ind = try DoubleSmoothedMomenta.init(allocator, .{
            .bar_component = case.bc,
            .quote_component = case.qc,
            .trade_component = case.tc,
        });
        defer ind.deinit();
        ind.fixSlices();

        try testing.expectEqualStrings(case.expected, ind.line.mnemonic);
    }
}

test "DM custom mnemonic" {
    const allocator = testing.allocator;

    var ind = try DoubleSmoothedMomenta.init(allocator, .{ .a = 10, .y = 3, .z = 5 });
    defer ind.deinit();
    ind.fixSlices();

    try testing.expectEqualStrings("dm(10,3,5)", ind.line.mnemonic);
}

test "DM invalid params" {
    const allocator = testing.allocator;

    const r1 = DoubleSmoothedMomenta.init(allocator, .{ .a = 0 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = DoubleSmoothedMomenta.init(allocator, .{ .y = 0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = DoubleSmoothedMomenta.init(allocator, .{ .z = 0 });
    try testing.expect(if (r3) |_| false else |_| true);
}
