const std = @import("std");
const math = std.math;

const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;
const bar_component = @import("bar_component");
const quote_component = @import("quote_component");
const trade_component = @import("trade_component");

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

/// Enumerates the outputs of the On-Balance Volume indicator.
pub const OnBalanceVolumeOutput = enum(u8) {
    /// The scalar value of the on-balance volume.
    value = 1,
};

/// Parameters to create an instance of the On-Balance Volume indicator.
pub const OnBalanceVolumeParams = struct {
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Joseph Granville's On-Balance Volume (OBV).
///
/// OBV is a cumulative volume indicator. On each update, if the price is higher
/// than the previous price, the volume is added to the running total; if the price
/// is lower, the volume is subtracted. If the price is unchanged, the total remains
/// the same.
pub const OnBalanceVolume = struct {
    line: LineIndicator,
    bar_func: *const fn (Bar) f64,
    previous_sample: f64,
    value: f64,
    primed: bool,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,

    pub fn init(params: OnBalanceVolumeParams) OnBalanceVolume {
        const bc = params.bar_component orelse .close;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        var mnemonic_len: usize = undefined;
        if (triple.len == 0) {
            const s = std.fmt.bufPrint(&mnemonic_buf, "obv", .{}) catch unreachable;
            mnemonic_len = s.len;
        } else {
            // Strip leading ", " from triple.
            const suffix = if (triple.len > 2) triple[2..] else triple;
            const s = std.fmt.bufPrint(&mnemonic_buf, "obv({s})", .{suffix}) catch unreachable;
            mnemonic_len = s.len;
        }

        const desc_str = "On-Balance Volume OBV";

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                desc_str,
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .bar_func = bar_component.componentValue(bc),
            .previous_sample = math.nan(f64),
            .value = math.nan(f64),
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *OnBalanceVolume) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    /// Update with volume = 1 (scalar path).
    pub fn update(self: *OnBalanceVolume, sample: f64) f64 {
        return self.updateWithVolume(sample, 1);
    }

    /// Update with the given sample and volume.
    pub fn updateWithVolume(self: *OnBalanceVolume, sample: f64, volume: f64) f64 {
        if (math.isNan(sample) or math.isNan(volume)) {
            return math.nan(f64);
        }

        if (!self.primed) {
            self.value = volume;
            self.primed = true;
        } else {
            if (sample > self.previous_sample) {
                self.value += volume;
            } else if (sample < self.previous_sample) {
                self.value -= volume;
            }
        }

        self.previous_sample = sample;
        return self.value;
    }

    pub fn isPrimed(self: *const OnBalanceVolume) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const OnBalanceVolume, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .on_balance_volume,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *OnBalanceVolume, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Shadows LineIndicator.updateBar to use bar volume.
    pub fn updateBar(self: *OnBalanceVolume, sample: *const Bar) OutputArray {
        const price = self.bar_func(sample.*);
        const value = self.updateWithVolume(price, sample.volume);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *OnBalanceVolume, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *OnBalanceVolume, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *OnBalanceVolume) indicator_mod.Indicator {
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
        const self: *OnBalanceVolume = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const OnBalanceVolume = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *OnBalanceVolume = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *OnBalanceVolume = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *OnBalanceVolume = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *OnBalanceVolume = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn testPrices() [12]f64 {
    return .{ 1, 2, 8, 4, 9, 6, 7, 13, 9, 10, 3, 12 };
}

fn testVolumes() [12]f64 {
    return .{ 100, 90, 200, 150, 500, 100, 300, 150, 100, 300, 200, 100 };
}

fn testExpected() [12]f64 {
    return .{ 100, 190, 390, 240, 740, 640, 940, 1090, 990, 1290, 1090, 1190 };
}

fn createObv() OnBalanceVolume {
    var obv = OnBalanceVolume.init(.{});
    obv.fixSlices();
    return obv;
}

test "on balance volume with volume" {
    const prices = testPrices();
    const vol = testVolumes();
    const expected = testExpected();

    var obv = createObv();

    for (0..12) |i| {
        const v = obv.updateWithVolume(prices[i], vol[i]);
        try testing.expect(!math.isNan(v));
        try testing.expect(obv.isPrimed());
        try testing.expectEqual(expected[i], v);
    }
}

test "on balance volume is primed" {
    var obv = createObv();

    try testing.expect(!obv.isPrimed());

    _ = obv.updateWithVolume(1.0, 100.0);
    try testing.expect(obv.isPrimed());

    _ = obv.updateWithVolume(2.0, 50.0);
    try testing.expect(obv.isPrimed());
}

test "on balance volume NaN" {
    var obv = createObv();

    try testing.expect(math.isNan(obv.update(math.nan(f64))));
    try testing.expect(math.isNan(obv.updateWithVolume(1.0, math.nan(f64))));
    try testing.expect(math.isNan(obv.updateWithVolume(math.nan(f64), math.nan(f64))));
}

test "on balance volume metadata" {
    var obv = createObv();
    obv.fixSlices();
    var m: Metadata = undefined;
    obv.getMetadata(&m);

    try testing.expectEqual(Identifier.on_balance_volume, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("obv", m.outputs_buf[0].mnemonic);
}

test "on balance volume update scalar" {
    var obv = createObv();
    const time: i64 = 1617235200;

    const out = obv.updateScalar(&.{ .time = time, .value = 10.0 });
    try testing.expectEqual(@as(usize, 1), out.len);
    const s = out.slice()[0].scalar;
    try testing.expectEqual(time, s.time);
    try testing.expectEqual(@as(f64, 1.0), s.value);
}

test "on balance volume update bar" {
    const prices = testPrices();
    const vol = testVolumes();
    const expected = testExpected();
    const time: i64 = 1617235200;

    var obv = createObv();

    for (0..12) |i| {
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = prices[i], .volume = vol[i] };
        const out = obv.updateBar(&bar);

        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(expected[i], s.value);
    }
}

test "on balance volume equal prices" {
    var obv = createObv();

    var v = obv.updateWithVolume(10.0, 100.0);
    try testing.expectEqual(@as(f64, 100.0), v);

    // Same price: value unchanged.
    v = obv.updateWithVolume(10.0, 200.0);
    try testing.expectEqual(@as(f64, 100.0), v);
}
