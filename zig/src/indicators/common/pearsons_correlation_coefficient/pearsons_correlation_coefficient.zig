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

/// Enumerates the outputs of the Pearson's Correlation Coefficient indicator.
pub const PearsonsCorrelationCoefficientOutput = enum(u8) {
    /// The scalar value of the Pearson's Correlation Coefficient.
    value = 1,
};

/// Parameters to create an instance of the indicator.
pub const PearsonsCorrelationCoefficientParams = struct {
    /// The length (number of time periods). Must be >= 1.
    length: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes Pearson's Correlation Coefficient (r) over a rolling window.
///
/// Given two input series X and Y, it computes:
///   r = (n*sumXY - sumX*sumY) / sqrt((n*sumX2 - sumX^2) * (n*sumY2 - sumY^2))
///
/// The indicator is not primed during the first length-1 updates.
/// For single-input updates, both X and Y are set to the same value (always returns 0 when primed).
/// For bar updates, X = High and Y = Low.
pub const PearsonsCorrelationCoefficient = struct {
    line: LineIndicator,
    window_x: []f64,
    window_y: []f64,
    length: usize,
    count: usize,
    pos: usize,
    sum_x: f64,
    sum_y: f64,
    sum_x2: f64,
    sum_y2: f64,
    sum_xy: f64,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: PearsonsCorrelationCoefficientParams) !PearsonsCorrelationCoefficient {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "correl({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Pearsons Correlation Coefficient {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window_x = try allocator.alloc(f64, params.length);
        @memset(window_x, 0.0);
        const window_y = try allocator.alloc(f64, params.length);
        @memset(window_y, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .window_x = window_x,
            .window_y = window_y,
            .length = params.length,
            .count = 0,
            .pos = 0,
            .sum_x = 0.0,
            .sum_y = 0.0,
            .sum_x2 = 0.0,
            .sum_y2 = 0.0,
            .sum_xy = 0.0,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *PearsonsCorrelationCoefficient) void {
        self.allocator.free(self.window_x);
        self.allocator.free(self.window_y);
    }

    pub fn fixSlices(self: *PearsonsCorrelationCoefficient) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update with a single scalar. Calls updatePair(sample, sample).
    pub fn update(self: *PearsonsCorrelationCoefficient, sample: f64) f64 {
        return self.updatePair(sample, sample);
    }

    /// Core update with an (x, y) pair.
    pub fn updatePair(self: *PearsonsCorrelationCoefficient, x: f64, y: f64) f64 {
        if (math.isNan(x) or math.isNan(y)) {
            return math.nan(f64);
        }

        const n: f64 = @floatFromInt(self.length);

        if (self.primed) {
            // Remove oldest values.
            const old_x = self.window_x[self.pos];
            const old_y = self.window_y[self.pos];

            self.sum_x -= old_x;
            self.sum_y -= old_y;
            self.sum_x2 -= old_x * old_x;
            self.sum_y2 -= old_y * old_y;
            self.sum_xy -= old_x * old_y;

            // Add new values.
            self.window_x[self.pos] = x;
            self.window_y[self.pos] = y;
            self.pos = (self.pos + 1) % self.length;

            self.sum_x += x;
            self.sum_y += y;
            self.sum_x2 += x * x;
            self.sum_y2 += y * y;
            self.sum_xy += x * y;

            return self.correlate(n);
        }

        // Accumulating phase.
        self.window_x[self.count] = x;
        self.window_y[self.count] = y;

        self.sum_x += x;
        self.sum_y += y;
        self.sum_x2 += x * x;
        self.sum_y2 += y * y;
        self.sum_xy += x * y;

        self.count += 1;

        if (self.count == self.length) {
            self.primed = true;
            self.pos = 0;
            return self.correlate(n);
        }

        return math.nan(f64);
    }

    /// Computes the Pearson correlation from the running sums.
    fn correlate(self: *const PearsonsCorrelationCoefficient, n: f64) f64 {
        const var_x = self.sum_x2 - (self.sum_x * self.sum_x) / n;
        const var_y = self.sum_y2 - (self.sum_y * self.sum_y) / n;
        const temp_real = var_x * var_y;

        if (temp_real <= 0) {
            return 0;
        }

        return (self.sum_xy - (self.sum_x * self.sum_y) / n) / @sqrt(temp_real);
    }

    pub fn isPrimed(self: *const PearsonsCorrelationCoefficient) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const PearsonsCorrelationCoefficient, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .pearsons_correlation_coefficient,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *PearsonsCorrelationCoefficient, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// UpdateBar overrides to extract High (X) and Low (Y) from the bar.
    pub fn updateBar(self: *PearsonsCorrelationCoefficient, sample: *const Bar) OutputArray {
        const x = sample.high;
        const y = sample.low;
        const value = self.updatePair(x, y);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *PearsonsCorrelationCoefficient, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *PearsonsCorrelationCoefficient, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *PearsonsCorrelationCoefficient) indicator_mod.Indicator {
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
        const self: *PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidLength,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn createCorrel(allocator: std.mem.Allocator, length: usize) !PearsonsCorrelationCoefficient {
    var c = try PearsonsCorrelationCoefficient.init(allocator, .{ .length = length });
    c.fixSlices();
    return c;
}

fn almostEqual(a: f64, b: f64, eps: f64) bool {
    return @abs(a - b) < eps;
}

test "pearsons correlation coefficient talib spot checks period=20" {
    const high = testdata.testHighInput();
    const low = testdata.testLowInput();

    var c = try createCorrel(testing.allocator, 20);
    defer c.deinit();

    // First 18 values should be NaN (lookback = 19, primed at index 19).
    for (0..19) |i| {
        const v = c.updatePair(high[i], low[i]);
        if (i < 19) {
            try testing.expect(math.isNan(v));
        }
    }

    // Feed remaining.
    for (19..252) |i| {
        const act = c.updatePair(high[i], low[i]);
        switch (i) {
            19 => try testing.expect(almostEqual(act, 0.9401569, 1e-4)),
            20 => try testing.expect(almostEqual(act, 0.9471812, 1e-4)),
            251 => try testing.expect(almostEqual(act, 0.8866901, 1e-4)),
            else => {},
        }
    }

    // NaN passthrough.
    try testing.expect(math.isNan(c.updatePair(math.nan(f64), 1.0)));
    try testing.expect(math.isNan(c.updatePair(1.0, math.nan(f64))));
}

test "pearsons correlation coefficient excel verification period=20" {
    const high = testdata.testHighInput();
    const low = testdata.testLowInput();
    const expected = testdata.testExcelExpected();
    const eps: f64 = 1e-10;

    var c = try createCorrel(testing.allocator, 20);
    defer c.deinit();

    for (0..252) |i| {
        const act = c.updatePair(high[i], low[i]);
        if (i >= 19) {
            if (math.isNan(expected[i])) {
                try testing.expect(math.isNan(act));
            } else {
                try testing.expect(almostEqual(act, expected[i], eps));
            }
        } else {
            try testing.expect(math.isNan(act));
        }
    }
}

test "pearsons correlation coefficient is primed length=1" {
    const high = testdata.testHighInput();
    const low = testdata.testLowInput();

    var c = try createCorrel(testing.allocator, 1);
    defer c.deinit();

    try testing.expect(!c.isPrimed());
    _ = c.updatePair(high[0], low[0]);
    try testing.expect(c.isPrimed());
}

test "pearsons correlation coefficient is primed length=2" {
    const high = testdata.testHighInput();
    const low = testdata.testLowInput();

    var c = try createCorrel(testing.allocator, 2);
    defer c.deinit();

    try testing.expect(!c.isPrimed());
    _ = c.updatePair(high[0], low[0]);
    try testing.expect(!c.isPrimed());
    _ = c.updatePair(high[1], low[1]);
    try testing.expect(c.isPrimed());
}

test "pearsons correlation coefficient is primed length=20" {
    const high = testdata.testHighInput();
    const low = testdata.testLowInput();

    var c = try createCorrel(testing.allocator, 20);
    defer c.deinit();

    try testing.expect(!c.isPrimed());
    for (0..19) |i| {
        _ = c.updatePair(high[i], low[i]);
        try testing.expect(!c.isPrimed());
    }
    _ = c.updatePair(high[19], low[19]);
    try testing.expect(c.isPrimed());
}

test "pearsons correlation coefficient metadata" {
    var c = try createCorrel(testing.allocator, 20);
    defer c.deinit();
    var m: Metadata = undefined;
    c.getMetadata(&m);

    try testing.expectEqual(Identifier.pearsons_correlation_coefficient, m.identifier);
    try testing.expectEqualStrings("correl(20)", m.mnemonic);
    try testing.expectEqualStrings("Pearsons Correlation Coefficient correl(20)", m.description);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqualStrings("correl(20)", m.outputs_buf[0].mnemonic);
}

test "pearsons correlation coefficient init invalid length" {
    const r = PearsonsCorrelationCoefficient.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, r);
}

test "pearsons correlation coefficient mnemonic with components" {
    var c = try PearsonsCorrelationCoefficient.init(testing.allocator, .{
        .length = 20,
        .bar_component = .median,
    });
    defer c.deinit();
    c.fixSlices();
    try testing.expectEqualStrings("correl(20, hl/2)", c.line.mnemonic);
}

test "pearsons correlation coefficient update entity bar" {
    var c = try createCorrel(testing.allocator, 2);
    defer c.deinit();

    // Feed one pair via updatePair, then one bar.
    _ = c.updatePair(10, 5);
    _ = c.updatePair(20, 10);
    const bar = Bar{ .time = 1617235200, .open = 0, .high = 30, .low = 15, .close = 0, .volume = 0 };
    const out = c.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
    const s = out.slice()[0].scalar;
    try testing.expectEqual(@as(i64, 1617235200), s.time);
    try testing.expect(!math.isNan(s.value));
}

test "pearsons correlation coefficient update entity scalar" {
    const inp: f64 = 3.0;
    const time: i64 = 1617235200;

    var c = try createCorrel(testing.allocator, 2);
    defer c.deinit();

    _ = c.update(inp);
    _ = c.update(inp);
    const out = c.updateScalar(&.{ .time = time, .value = inp });
    try testing.expectEqual(@as(usize, 1), out.len);
    const s = out.slice()[0].scalar;
    try testing.expectEqual(time, s.time);
    // correl(x,x) with constant value returns 0.
    try testing.expect(almostEqual(s.value, 0.0, 1e-10));
}

test "pearsons correlation coefficient update entity quote" {
    const inp: f64 = 3.0;
    const time: i64 = 1617235200;

    var c = try createCorrel(testing.allocator, 2);
    defer c.deinit();

    _ = c.update(inp);
    _ = c.update(inp);
    const q = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 1, .ask_size = 1 };
    const out = c.updateQuote(&q);
    const s = out.slice()[0].scalar;
    try testing.expect(almostEqual(s.value, 0.0, 1e-10));
}

test "pearsons correlation coefficient update entity trade" {
    const inp: f64 = 3.0;
    const time: i64 = 1617235200;

    var c = try createCorrel(testing.allocator, 2);
    defer c.deinit();

    _ = c.update(inp);
    _ = c.update(inp);
    const t = Trade{ .time = time, .price = inp, .volume = 1 };
    const out = c.updateTrade(&t);
    const s = out.slice()[0].scalar;
    try testing.expect(almostEqual(s.value, 0.0, 1e-10));
}
