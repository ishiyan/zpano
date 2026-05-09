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
const ema_mod = @import("../../common/exponential_moving_average/exponential_moving_average.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const ExponentialMovingAverage = ema_mod.ExponentialMovingAverage;

/// Enumerates the outputs of the TRIX indicator.
pub const TripleExponentialMovingAverageOscillatorOutput = enum(u8) {
    /// The scalar value of the TRIX oscillator.
    value = 1,
};

/// Parameters to create an instance of the TRIX indicator.
pub const TripleExponentialMovingAverageOscillatorParams = struct {
    /// The length (number of time periods). Must be >= 1.
    length: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Jack Hutson's Triple Exponential Moving Average Oscillator (TRIX).
///
/// TRIX is a 1-day rate-of-change of a triple-smoothed exponential moving average:
///
///   TRIX = ((EMA3[i] - EMA3[i-1]) / EMA3[i-1]) * 100
///
/// The indicator oscillates around zero. Positive values indicate upward momentum,
/// negative values indicate downward momentum.
pub const TripleExponentialMovingAverageOscillator = struct {
    line: LineIndicator,
    ema1: ExponentialMovingAverage,
    ema2: ExponentialMovingAverage,
    ema3: ExponentialMovingAverage,
    previous_ema3: f64,
    has_previous_ema: bool,
    primed: bool,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(params: TripleExponentialMovingAverageOscillatorParams) !TripleExponentialMovingAverageOscillator {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "trix({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Triple exponential moving average oscillator {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const ema_params = ema_mod.ExponentialMovingAverageLengthParams{
            .length = params.length,
            .first_is_average = true,
        };

        var ema1 = try ExponentialMovingAverage.initLength(ema_params);
        ema1.fixSlices();
        var ema2 = try ExponentialMovingAverage.initLength(ema_params);
        ema2.fixSlices();
        var ema3 = try ExponentialMovingAverage.initLength(ema_params);
        ema3.fixSlices();

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .ema1 = ema1,
            .ema2 = ema2,
            .ema3 = ema3,
            .previous_ema3 = math.nan(f64),
            .has_previous_ema = false,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *TripleExponentialMovingAverageOscillator) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
        // Internal EMAs don't need fixSlices since we don't use their mnemonics.
    }

    /// Core update logic.
    pub fn update(self: *TripleExponentialMovingAverageOscillator, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const v1 = self.ema1.update(sample);
        if (math.isNan(v1)) return math.nan(f64);

        const v2 = self.ema2.update(v1);
        if (math.isNan(v2)) return math.nan(f64);

        const v3 = self.ema3.update(v2);
        if (math.isNan(v3)) return math.nan(f64);

        if (!self.has_previous_ema) {
            self.previous_ema3 = v3;
            self.has_previous_ema = true;
            return math.nan(f64);
        }

        const result = ((v3 - self.previous_ema3) / self.previous_ema3) * 100.0;
        self.previous_ema3 = v3;

        if (!self.primed) {
            self.primed = true;
        }

        return result;
    }

    pub fn isPrimed(self: *const TripleExponentialMovingAverageOscillator) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const TripleExponentialMovingAverageOscillator, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .triple_exponential_moving_average_oscillator,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *TripleExponentialMovingAverageOscillator, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *TripleExponentialMovingAverageOscillator, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *TripleExponentialMovingAverageOscillator, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *TripleExponentialMovingAverageOscillator, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *TripleExponentialMovingAverageOscillator) indicator_mod.Indicator {
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
        const self: *TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *TripleExponentialMovingAverageOscillator = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn createTrix(length: usize) !TripleExponentialMovingAverageOscillator {
    var trix = try TripleExponentialMovingAverageOscillator.init(.{ .length = length });
    trix.fixSlices();
    return trix;
}

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

test "trix values" {
    const tolerance = 1e-10;
    const closes = testdata.testCloses();
    const expected = testdata.testExpected();

    var trix = try createTrix(5);

    for (closes, expected, 0..) |c, exp, i| {
        const result = trix.update(c);
        _ = i;

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(result));
        } else {
            try testing.expect(!math.isNan(result));
            try testing.expect(almostEqual(result, exp, tolerance));
        }
    }
}

test "trix is primed" {
    const closes = testdata.testCloses();

    var trix = try createTrix(5);

    // Lookback = 3*(5-1) + 1 = 13. First primed at index 13.
    for (0..13) |i| {
        _ = trix.update(closes[i]);
        try testing.expect(!trix.isPrimed());
    }

    _ = trix.update(closes[13]);
    try testing.expect(trix.isPrimed());
}

test "trix metadata" {
    var trix = try TripleExponentialMovingAverageOscillator.init(.{ .length = 30 });
    trix.fixSlices();

    var m: Metadata = undefined;
    trix.getMetadata(&m);

    try testing.expectEqual(Identifier.triple_exponential_moving_average_oscillator, m.identifier);
    try testing.expectEqualStrings("trix(30)", m.mnemonic);
    try testing.expectEqualStrings("Triple exponential moving average oscillator trix(30)", m.description);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
}

test "trix invalid params" {
    const result = TripleExponentialMovingAverageOscillator.init(.{ .length = 0 });
    try testing.expect(result == error.InvalidLength);
}

test "trix NaN" {
    var trix = try createTrix(5);
    const result = trix.update(math.nan(f64));
    try testing.expect(math.isNan(result));
}

test "trix update entity" {
    var trix = try createTrix(5);
    const time: i64 = 1617235200;

    const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = 100.0, .volume = 0 };
    const out = trix.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
}
