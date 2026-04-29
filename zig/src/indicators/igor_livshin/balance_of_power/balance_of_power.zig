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
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

const epsilon = 1e-8;

/// Enumerates the outputs of the Balance of Power indicator.
pub const BalanceOfPowerOutput = enum(u8) {
    /// The scalar value of the balance of power.
    value = 1,
};

/// Igor Livshin's Balance of Power (BOP).
///
/// The Balance of Market Power captures the struggles of bulls vs. bears
/// throughout the trading day. It assigns scores to both bulls and bears
/// based on how much they were able to move prices throughout the trading day.
///
/// BOP = (Close - Open) / (High - Low)
///
/// When the range (High - Low) is less than epsilon, the value is 0.
/// The indicator is always primed.
pub const BalanceOfPower = struct {
    line: LineIndicator,
    value: f64,

    const mnemonic_str = "bop";
    const description_str = "Balance of Power";

    pub fn init() BalanceOfPower {
        return .{
            .line = LineIndicator.new(
                mnemonic_str,
                description_str,
                null,
                null,
                null,
            ),
            .value = math.nan(f64),
        };
    }

    /// Core update logic. For scalar/quote/trade, O=H=L=C so BOP is always 0.
    pub fn update(self: *BalanceOfPower, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }
        return self.updateOhlc(sample, sample, sample, sample);
    }

    /// Updates the indicator with the given OHLC values.
    pub fn updateOhlc(self: *BalanceOfPower, open: f64, high: f64, low: f64, close: f64) f64 {
        if (math.isNan(open) or math.isNan(high) or math.isNan(low) or math.isNan(close)) {
            return math.nan(f64);
        }

        const r = high - low;
        if (r < epsilon) {
            self.value = 0;
        } else {
            self.value = (close - open) / r;
        }

        return self.value;
    }

    pub fn isPrimed(_: *const BalanceOfPower) bool {
        return true;
    }

    pub fn getMetadata(self: *const BalanceOfPower, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .balance_of_power,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *BalanceOfPower, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Shadows LineIndicator.updateBar to extract OHLC directly from the bar.
    pub fn updateBar(self: *BalanceOfPower, sample: *const Bar) OutputArray {
        const value = self.updateOhlc(sample.open, sample.high, sample.low, sample.close);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *BalanceOfPower, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *BalanceOfPower, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *BalanceOfPower) indicator_mod.Indicator {
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
        const self: *BalanceOfPower = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const BalanceOfPower = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *BalanceOfPower = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *BalanceOfPower = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *BalanceOfPower = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *BalanceOfPower = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn testOpen() [20]f64 {
    return .{
        92.500, 91.500, 95.155, 93.970, 95.500, 94.500, 95.000, 91.500, 91.815, 91.125,
        93.875, 97.500, 98.815, 92.000, 91.125, 91.875, 93.405, 89.750, 89.345, 92.250,
    };
}

fn testHigh() [20]f64 {
    return .{
        93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000,
        96.250000, 99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000,
    };
}

fn testLow() [20]f64 {
    return .{
        90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000,
        92.750000, 96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000,
    };
}

fn testClose() [20]f64 {
    return .{
        91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
        96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
    };
}

fn testExpected() [20]f64 {
    return .{
        -0.400000000000000, 0.937765205091938,  -0.367058823529412, 0.418215613382900,  -0.540031397174254,
        0.102459016393443,  -0.823333333333333, 0.314861460957179,  -0.495049504950495, 0.632941176470588,
        0.642857142857143,  -0.075528700906344, -0.101777059773828, -0.540025412960610, -0.027382256297919,
        0.406926406926406,  -0.944444444444444, -0.216000000000001, 0.838235294117648,  -0.950000000000000,
    };
}

fn roundTo(v: f64, comptime digits: comptime_int) f64 {
    const p = comptime std.math.pow(f64, 10.0, @as(f64, @floatFromInt(digits)));
    return @round(v * p) / p;
}

test "balance of power OHLC" {
    const open = testOpen();
    const high = testHigh();
    const low = testLow();
    const close = testClose();
    const expected = testExpected();

    var bop = BalanceOfPower.init();

    for (0..20) |i| {
        const v = bop.updateOhlc(open[i], high[i], low[i], close[i]);
        try testing.expect(!math.isNan(v));
        try testing.expect(bop.isPrimed());

        const got = roundTo(v, 9);
        const exp = roundTo(expected[i], 9);
        try testing.expect(got == exp);
    }
}

test "balance of power is primed" {
    var bop = BalanceOfPower.init();

    // Always primed.
    try testing.expect(bop.isPrimed());

    _ = bop.updateOhlc(92.5, 93.25, 90.75, 91.5);
    try testing.expect(bop.isPrimed());
}

test "balance of power NaN" {
    var bop = BalanceOfPower.init();

    try testing.expect(math.isNan(bop.update(math.nan(f64))));
    try testing.expect(math.isNan(bop.updateOhlc(math.nan(f64), 1.0, 2.0, 3.0)));
    try testing.expect(math.isNan(bop.updateOhlc(1.0, math.nan(f64), 2.0, 3.0)));
    try testing.expect(math.isNan(bop.updateOhlc(1.0, 2.0, math.nan(f64), 3.0)));
    try testing.expect(math.isNan(bop.updateOhlc(1.0, 2.0, 3.0, math.nan(f64))));
}

test "balance of power zero range" {
    var bop = BalanceOfPower.init();
    const v = bop.updateOhlc(0.001, 0.001, 0.001, 0.001);
    try testing.expectEqual(@as(f64, 0), v);
}

test "balance of power scalar always zero" {
    var bop = BalanceOfPower.init();
    try testing.expectEqual(@as(f64, 0), bop.update(50.0));
    try testing.expectEqual(@as(f64, 0), bop.update(100.0));
}

test "balance of power metadata" {
    var bop = BalanceOfPower.init();
    var m: Metadata = undefined;
    bop.getMetadata(&m);

    try testing.expectEqual(Identifier.balance_of_power, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("bop", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Balance of Power", m.outputs_buf[0].description);
}

test "balance of power update bar" {
    const open = testOpen();
    const high = testHigh();
    const low = testLow();
    const close = testClose();
    const expected = testExpected();
    const time: i64 = 1617235200;

    var bop = BalanceOfPower.init();

    for (0..20) |i| {
        const bar = Bar{ .time = time, .open = open[i], .high = high[i], .low = low[i], .close = close[i], .volume = 0 };
        const out = bop.updateBar(&bar);

        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);

        const got = roundTo(s.value, 9);
        const exp = roundTo(expected[i], 9);
        try testing.expect(got == exp);
    }
}
