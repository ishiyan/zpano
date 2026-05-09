const std = @import("std");
const math = std.math;


const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
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

/// Enumerates the outputs of the Advance-Decline indicator.
pub const AdvanceDeclineOutput = enum(u8) {
    /// The scalar value of the A/D line.
    value = 1,
};

/// Marc Chaikin's Advance-Decline (A/D) Line.
///
/// The Accumulation/Distribution Line is a cumulative indicator that uses volume
/// and price to assess whether a stock is being accumulated or distributed.
///
/// CLV = ((Close - Low) - (High - Close)) / (High - Low)
/// AD  = AD_previous + CLV × Volume
///
/// When High equals Low, the A/D value is unchanged (no division by zero).
pub const AdvanceDecline = struct {
    line: LineIndicator,
    ad: f64,
    value: f64,
    primed: bool,

    const mnemonic_str = "ad";
    const description_str = "Advance-Decline";

    pub fn init() AdvanceDecline {
        return .{
            .line = LineIndicator.new(
                mnemonic_str,
                description_str,
                null,
                null,
                null,
            ),
            .ad = 0,
            .value = math.nan(f64),
            .primed = false,
        };
    }

    pub fn deinit(_: *AdvanceDecline) void {}

    pub fn fixSlices(_: *AdvanceDecline) void {}

    /// Core update logic. For scalar/quote/trade, H=L=C so AD is unchanged.
    pub fn update(self: *AdvanceDecline, sample: f64) f64 {
        if (math.isNan(sample)) {
            return math.nan(f64);
        }
        return self.updateHlcv(sample, sample, sample, 1);
    }

    /// Updates the indicator with the given high, low, close, and volume values.
    pub fn updateHlcv(self: *AdvanceDecline, high: f64, low: f64, close: f64, volume: f64) f64 {
        if (math.isNan(high) or math.isNan(low) or math.isNan(close) or math.isNan(volume)) {
            return math.nan(f64);
        }

        const temp = high - low;
        if (temp > 0) {
            self.ad += ((close - low) - (high - close)) / temp * volume;
        }

        self.value = self.ad;
        self.primed = true;

        return self.value;
    }

    pub fn isPrimed(self: *const AdvanceDecline) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const AdvanceDecline, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .advance_decline,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *AdvanceDecline, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Shadows LineIndicator.updateBar to extract HLCV directly from the bar.
    pub fn updateBar(self: *AdvanceDecline, sample: *const Bar) OutputArray {
        const value = self.updateHlcv(sample.high, sample.low, sample.close, sample.volume);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *AdvanceDecline, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *AdvanceDecline, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *AdvanceDecline) indicator_mod.Indicator {
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
        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const AdvanceDecline = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AdvanceDecline = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn roundTo(v: f64, comptime digits: comptime_int) f64 {
    const p = comptime blk: {
        var result: f64 = 1.0;
        for (0..digits) |_| result *= 10.0;
        break :blk result;
    };
    return @round(v * p) / p;
}

// High test data, 252 entries.
// Low test data, 252 entries.
// Close test data, 252 entries.
// Volume test data, 252 entries.
// Expected AD output, 252 entries.
test "AdvanceDecline with volume" {
    const highs = testdata.testHighs();
    const lows = testdata.testLows();
    const closes = testdata.testCloses();
    const volumes = testdata.testVolumes();
    const expected = testdata.testExpectedAD();

    var ad = AdvanceDecline.init();

    for (0..252) |i| {
        const v = ad.updateHlcv(highs[i], lows[i], closes[i], volumes[i]);
        try testing.expect(!math.isNan(v));
        try testing.expect(ad.isPrimed());

        const got = roundTo(v, 2);
        const exp = roundTo(expected[i], 2);
        try testing.expectEqual(exp, got);
    }
}

test "AdvanceDecline TA-Lib spot checks" {
    const highs = testdata.testHighs();
    const lows = testdata.testLows();
    const closes = testdata.testCloses();
    const volumes = testdata.testVolumes();

    var ad = AdvanceDecline.init();
    var values: [252]f64 = undefined;

    for (0..252) |i| {
        values[i] = ad.updateHlcv(highs[i], lows[i], closes[i], volumes[i]);
    }

    try testing.expectEqual(roundTo(-1631000.00, 2), roundTo(values[0], 2));
    try testing.expectEqual(roundTo(2974412.02, 2), roundTo(values[1], 2));
    try testing.expectEqual(roundTo(8707691.07, 2), roundTo(values[250], 2));
    try testing.expectEqual(roundTo(8328944.54, 2), roundTo(values[251], 2));
}

test "AdvanceDecline updateBar" {
    var ad = AdvanceDecline.init();

    const highs = testdata.testHighs();
    const lows = testdata.testLows();
    const closes = testdata.testCloses();
    const volumes = testdata.testVolumes();
    const expected = testdata.testExpectedAD();

    for (0..10) |i| {
        const bar = Bar{
            .time = 0,
            .open = highs[i],
            .high = highs[i],
            .low = lows[i],
            .close = closes[i],
            .volume = volumes[i],
        };

        const output = ad.updateBar(&bar);
        const scalar_val = output.slice()[0].scalar.value;

        const got = roundTo(scalar_val, 2);
        const exp = roundTo(expected[i], 2);
        try testing.expectEqual(exp, got);
    }
}

test "AdvanceDecline scalar update" {
    var ad = AdvanceDecline.init();

    // Scalar update: H=L=C, so range=0, AD unchanged (remains 0 after primed).
    const v = ad.update(100.0);
    try testing.expectEqual(@as(f64, 0), v);
    try testing.expect(ad.isPrimed());
}

test "AdvanceDecline NaN" {
    var ad = AdvanceDecline.init();

    try testing.expect(math.isNan(ad.update(math.nan(f64))));
    try testing.expect(math.isNan(ad.updateHlcv(math.nan(f64), 1, 2, 3)));
    try testing.expect(math.isNan(ad.updateHlcv(1, math.nan(f64), 2, 3)));
    try testing.expect(math.isNan(ad.updateHlcv(1, 2, math.nan(f64), 3)));
    try testing.expect(math.isNan(ad.updateHlcv(1, 2, 3, math.nan(f64))));
}

test "AdvanceDecline not primed initially" {
    var ad = AdvanceDecline.init();

    try testing.expect(!ad.isPrimed());
    try testing.expect(math.isNan(ad.update(math.nan(f64))));
    try testing.expect(!ad.isPrimed());
}

test "AdvanceDecline metadata" {
    var ad = AdvanceDecline.init();
    var meta: Metadata = undefined;
    ad.getMetadata(&meta);

    try testing.expectEqual(Identifier.advance_decline, meta.identifier);
    try testing.expectEqualStrings("ad", meta.mnemonic);
    try testing.expectEqualStrings("Advance-Decline", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}
