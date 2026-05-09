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

/// Enumerates the outputs of the Relative Strength Index indicator.
pub const RelativeStrengthIndexOutput = enum(u8) {
    /// The RSI value (0..100).
    value = 1,
};

/// Parameters for the Relative Strength Index indicator.
pub const RelativeStrengthIndexParams = struct {
    /// The smoothing length. Must be >= 2. Default is 14.
    length: usize = 14,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Welles Wilder's Relative Strength Index (RSI).
///
/// RSI = 100 * avg_gain / (avg_gain + avg_loss)
///
/// Uses Wilder's smoothing:
///   avg_gain(n) = avg_gain(n-1) * (n-1)/n + gain_today / n
///   avg_loss(n) = avg_loss(n-1) * (n-1)/n + loss_today / n
///
/// The indicator is not primed until `length` samples have been processed.
pub const RelativeStrengthIndex = struct {
    line: LineIndicator,
    length: usize,
    count: isize,
    previous_sample: f64,
    previous_gain: f64,
    previous_loss: f64,
    value: f64,
    primed: bool,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub const Error = error{ InvalidLength, MnemonicTooLong, DescriptionTooLong };

    pub fn init(params: RelativeStrengthIndexParams) Error!RelativeStrengthIndex {
        if (params.length < 2) return error.InvalidLength;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "rsi({d}{s})", .{ params.length, triple }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Relative Strength Index {s}", .{mnemonic_slice}) catch return error.DescriptionTooLong;
        const description_len = desc_slice.len;

        return .{
            .line = LineIndicator.new(mnemonic_slice, desc_slice, bc, qc, tc),
            .length = params.length,
            .count = -1,
            .previous_sample = 0,
            .previous_gain = 0,
            .previous_loss = 0,
            .value = math.nan(f64),
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(_: *RelativeStrengthIndex) void {}
    pub fn fixSlices(self: *RelativeStrengthIndex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update given a single sample value.
    pub fn update(self: *RelativeStrengthIndex, sample: f64) f64 {
        if (math.isNan(sample)) return math.nan(f64);

        self.count += 1;

        if (self.count == 0) {
            self.previous_sample = sample;
            return self.value;
        }

        const temp = sample - self.previous_sample;
        self.previous_sample = sample;

        if (!self.primed) {
            if (temp < 0) {
                self.previous_loss -= temp;
            } else {
                self.previous_gain += temp;
            }

            if (self.count < self.length) {
                return self.value;
            }

            self.previous_gain /= @floatFromInt(self.length);
            self.previous_loss /= @floatFromInt(self.length);
            self.primed = true;
        } else {
            const n: f64 = @floatFromInt(self.length);
            self.previous_gain *= n - 1;
            self.previous_loss *= n - 1;

            if (temp < 0) {
                self.previous_loss -= temp;
            } else {
                self.previous_gain += temp;
            }

            self.previous_gain /= n;
            self.previous_loss /= n;
        }

        const sum = self.previous_gain + self.previous_loss;
        if (sum > 1e-8) {
            self.value = 100.0 * self.previous_gain / sum;
        } else {
            self.value = 0.0;
        }

        return self.value;
    }

    /// Update using a single sample value (convenience method).
    pub fn updateSample(self: *RelativeStrengthIndex, sample: f64) f64 {
        return self.update(sample);
    }

    pub fn isPrimed(self: *const RelativeStrengthIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const RelativeStrengthIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.relative_strength_index, self.line.mnemonic, self.line.description, &.{
            .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
        });
    }

    fn makeOutput(self: *const RelativeStrengthIndex, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *RelativeStrengthIndex, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *RelativeStrengthIndex, sample: *const Bar) OutputArray {
        const v = self.line.extractBar(sample);
        _ = self.update(v);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *RelativeStrengthIndex, sample: *const Quote) OutputArray {
        const v = self.line.extractQuote(sample);
        _ = self.update(v);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *RelativeStrengthIndex, sample: *const Trade) OutputArray {
        const v = self.line.extractTrade(sample);
        _ = self.update(v);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *RelativeStrengthIndex) indicator_mod.Indicator {
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
        const self: *const RelativeStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const RelativeStrengthIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *RelativeStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *RelativeStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *RelativeStrengthIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *RelativeStrengthIndex = @ptrCast(@alignCast(ptr));
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

// Test data from TA-Lib reference (length=9, 25 entries).
// Test data from TA-Lib reference (length=14, 252 entries, same as DM+).
test "RelativeStrengthIndex update length=9" {
    const tolerance = 1e-9;
    var rsi: RelativeStrengthIndex = try RelativeStrengthIndex.init(.{ .length = 9 });
    rsi.fixSlices();

    for (0..testdata.test_input_1.len) |i| {
        const act = rsi.update(testdata.test_input_1[i]);
        const exp = testdata.test_expected_1[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "RelativeStrengthIndex update length=14" {
    // Matches Go's TestRelativeStrengthIndexUpdate2: check NaN during warmup,
    // then spot-check final value is in valid RSI range [0, 100].
    var rsi: RelativeStrengthIndex = try RelativeStrengthIndex.init(.{ .length = 14 });
    rsi.fixSlices();

    var act: f64 = math.nan(f64);
    for (0..testdata.test_input_2.len) |i| {
        act = rsi.update(testdata.test_input_2[i]);
        if (i < 14) {
            try testing.expect(math.isNan(act));
        }
    }

    // Final value should be in valid RSI range.
    try testing.expect(act >= 0.0 and act <= 100.0);
}

test "RelativeStrengthIndex constructor validation" {
    const result = RelativeStrengthIndex.init(.{ .length = 1 });
    try testing.expectError(error.InvalidLength, result);
}

test "RelativeStrengthIndex isPrimed length=9" {
    var rsi: RelativeStrengthIndex = try RelativeStrengthIndex.init(.{ .length = 9 });
    rsi.fixSlices();

    try testing.expect(!rsi.isPrimed());

    for (0..9) |_| {
        _ = rsi.update(100.0);
    }
    try testing.expect(!rsi.isPrimed());

    _ = rsi.update(100.0);
    try testing.expect(rsi.isPrimed());
}

test "RelativeStrengthIndex NaN passthrough" {
    var rsi: RelativeStrengthIndex = try RelativeStrengthIndex.init(.{ .length = 14 });
    rsi.fixSlices();

    try testing.expect(math.isNan(rsi.update(math.nan(f64))));
    try testing.expect(math.isNan(rsi.updateSample(math.nan(f64))));
}

test "RelativeStrengthIndex metadata" {
    var rsi: RelativeStrengthIndex = try RelativeStrengthIndex.init(.{ .length = 14 });
    rsi.fixSlices();
    var meta: Metadata = undefined;
    rsi.getMetadata(&meta);

    try testing.expectEqual(Identifier.relative_strength_index, meta.identifier);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "RelativeStrengthIndex updateBar" {
    var rsi: RelativeStrengthIndex = try RelativeStrengthIndex.init(.{ .length = 9 });
    rsi.fixSlices();

    for (0..8) |i| {
        const bar = Bar{ .time = @intCast(i), .open = 0, .high = 0, .low = 0, .close = testdata.test_input_1[i], .volume = 0 };
        _ = rsi.updateBar(&bar);
    }

    const bar = Bar{ .time = 42, .open = 0, .high = 0, .low = 0, .close = testdata.test_input_1[8], .volume = 0 };
    const out = rsi.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
}
