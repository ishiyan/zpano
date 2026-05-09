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

/// Enumerates the outputs of the True Range indicator.
pub const TrueRangeOutput = enum(u8) {
    /// The scalar value of the True Range.
    value = 1,
};

/// Welles Wilder's True Range indicator.
///
/// The True Range is defined as the largest of:
///   - the distance from today's high to today's low
///   - the distance from yesterday's close to today's high
///   - the distance from yesterday's close to today's low
///
/// The first update stores the close and returns NaN (not primed).
/// The indicator is primed from the second update onward.
pub const TrueRange = struct {
    previous_close: f64,
    value: f64,
    primed: bool,

    const mnemonic_str = "tr";
    const description_str = "True Range";

    pub fn init() TrueRange {
        return .{
            .previous_close = math.nan(f64),
            .value = math.nan(f64),
            .primed = false,
        };
    }

    pub fn deinit(_: *TrueRange) void {}
    pub fn fixSlices(_: *TrueRange) void {}

    /// Core update given close, high, low values.
    pub fn update(self: *TrueRange, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) {
            return math.nan(f64);
        }

        if (!self.primed) {
            if (math.isNan(self.previous_close)) {
                self.previous_close = close;
                return math.nan(f64);
            }
            self.primed = true;
        }

        var greatest = high - low;

        const temp1 = @abs(high - self.previous_close);
        if (greatest < temp1) greatest = temp1;

        const temp2 = @abs(low - self.previous_close);
        if (greatest < temp2) greatest = temp2;

        self.value = greatest;
        self.previous_close = close;

        return self.value;
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *TrueRange, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const TrueRange) bool {
        return self.primed;
    }

    pub fn getMetadata(_: *const TrueRange, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.true_range, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
        });
    }

    fn makeOutput(self: *const TrueRange, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *TrueRange, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *TrueRange, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *TrueRange, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *TrueRange, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *TrueRange) indicator_mod.Indicator {
        return indicator_mod.Indicator{
            .ptr = @ptrCast(self),
            .vtable = &.{
                .isPrimed = vtableIsPrimed,
                .metadata = vtableMetadata,
                .updateScalar = vtableUpdateScalar,
                .updateBar = vtableUpdateBar,
                .updateQuote = vtableUpdateQuote,
                .updateTrade = vtableUpdateTrade,
            },
        };
    }

    fn vtableIsPrimed(ptr: *const anyopaque) bool {
        const self: *const TrueRange = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const TrueRange = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *TrueRange = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *TrueRange = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *TrueRange = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *TrueRange = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ============================================================================
// Tests
// ============================================================================

const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    if (math.isNan(a) and math.isNan(b)) return true;
    if (math.isNan(a) or math.isNan(b)) return false;
    return @abs(a - b) <= tolerance;
}

test "TrueRange update" {
    const tolerance = 1e-3;
    var tr = TrueRange.init();

    for (0..testdata.test_close.len) |i| {
        const act = tr.update(testdata.test_close[i], testdata.test_high[i], testdata.test_low[i]);
        const exp = testdata.test_expected_tr[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "TrueRange NaN passthrough" {
    var tr = TrueRange.init();

    try testing.expect(math.isNan(tr.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(tr.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(tr.update(1, 1, math.nan(f64))));
}

test "TrueRange isPrimed" {
    var tr = TrueRange.init();

    try testing.expect(!tr.isPrimed());

    _ = tr.update(testdata.test_close[0], testdata.test_high[0], testdata.test_low[0]);
    try testing.expect(!tr.isPrimed());

    _ = tr.update(testdata.test_close[1], testdata.test_high[1], testdata.test_low[1]);
    try testing.expect(tr.isPrimed());

    _ = tr.update(testdata.test_close[2], testdata.test_high[2], testdata.test_low[2]);
    try testing.expect(tr.isPrimed());
}

test "TrueRange updateSample" {
    var tr = TrueRange.init();

    const v0 = tr.updateSample(100.0);
    try testing.expect(math.isNan(v0));

    const v1 = tr.updateSample(105.0);
    try testing.expect(almostEqual(v1, 5.0, 1e-10));

    const v2 = tr.updateSample(102.0);
    try testing.expect(almostEqual(v2, 3.0, 1e-10));
}

test "TrueRange metadata" {
    var tr = TrueRange.init();
    var meta: Metadata = undefined;
    tr.getMetadata(&meta);

    try testing.expectEqual(Identifier.true_range, meta.identifier);
    try testing.expectEqualStrings("tr", meta.mnemonic);
    try testing.expectEqualStrings("True Range", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "TrueRange updateBar" {
    var tr = TrueRange.init();

    // Prime with first bar.
    _ = tr.update(100, 105, 95);

    const bar = Bar{ .time = 42, .open = 0, .high = 110, .low = 98, .close = 108, .volume = 0 };
    const out = tr.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
    try testing.expect(!math.isNan(out.slice()[0].scalar.value));
}
