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

/// Default inverse scaling factor. The value of 0.015 ensures that approximately
/// 70 to 80 percent of CCI values fall between -100 and +100.
pub const default_inverse_scaling_factor: f64 = 0.015;

/// Enumerates the outputs of the Commodity Channel Index indicator.
pub const CommodityChannelIndexOutput = enum(u8) {
    /// The scalar value of the commodity channel index.
    value = 1,
};

/// Parameters to create an instance of the Commodity Channel Index indicator.
pub const CommodityChannelIndexParams = struct {
    /// The length (number of time periods). Must be >= 2.
    length: usize,
    /// Inverse scaling factor. 0 means use default (0.015).
    inverse_scaling_factor: f64 = 0,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Donald Lambert's Commodity Channel Index (CCI).
///
/// CCI measures the deviation of the price from its statistical mean.
///
///   CCI = (typicalPrice - SMA) / (scalingFactor * meanDeviation)
///
/// where scalingFactor defaults to 0.015 so that approximately 70-80% of CCI values
/// fall between -100 and +100.
pub const CommodityChannelIndex = struct {
    line: LineIndicator,
    length: usize,
    scaling_factor: f64,
    window: []f64,
    window_count: usize,
    window_sum: f64,
    value: f64,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: CommodityChannelIndexParams) !CommodityChannelIndex {
        if (params.length < 2) {
            return error.InvalidLength;
        }

        const inverse_factor = if (params.inverse_scaling_factor == 0)
            default_inverse_scaling_factor
        else
            params.inverse_scaling_factor;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "cci({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Commodity Channel Index {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window = try allocator.alloc(f64, params.length);
        @memset(window, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .length = params.length,
            .scaling_factor = @as(f64, @floatFromInt(params.length)) / inverse_factor,
            .window = window,
            .window_count = 0,
            .window_sum = 0.0,
            .value = math.nan(f64),
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *CommodityChannelIndex) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *CommodityChannelIndex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update logic. Returns the CCI value or NaN if not yet primed.
    pub fn update(self: *CommodityChannelIndex, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const last_index = self.length - 1;
        const len_f: f64 = @floatFromInt(self.length);

        if (self.primed) {
            self.window_sum += sample - self.window[0];

            // Shift window left.
            var i: usize = 0;
            while (i < last_index) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }
            self.window[last_index] = sample;

            const average = self.window_sum / len_f;

            var temp: f64 = 0;
            for (self.window) |w| {
                temp += @abs(w - average);
            }

            if (@abs(temp) < math.floatMin(f64)) {
                self.value = 0;
            } else {
                self.value = self.scaling_factor * (sample - average) / temp;
            }
        } else {
            self.window_sum += sample;
            self.window[self.window_count] = sample;
            self.window_count += 1;

            if (self.window_count == self.length) {
                self.primed = true;

                const average = self.window_sum / len_f;

                var temp: f64 = 0;
                for (self.window) |w| {
                    temp += @abs(w - average);
                }

                if (@abs(temp) < math.floatMin(f64)) {
                    self.value = 0;
                } else {
                    self.value = self.scaling_factor * (sample - average) / temp;
                }
            }
        }

        return self.value;
    }

    pub fn isPrimed(self: *const CommodityChannelIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const CommodityChannelIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .commodity_channel_index,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *CommodityChannelIndex, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *CommodityChannelIndex, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *CommodityChannelIndex, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *CommodityChannelIndex, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *CommodityChannelIndex) indicator_mod.Indicator {
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
        const self: *CommodityChannelIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const CommodityChannelIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *CommodityChannelIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *CommodityChannelIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *CommodityChannelIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *CommodityChannelIndex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn createCci(allocator: std.mem.Allocator, length: usize) !CommodityChannelIndex {
    var cci = try CommodityChannelIndex.init(allocator, .{ .length = length });
    cci.fixSlices();
    return cci;
}

fn createCciWithInverse(allocator: std.mem.Allocator, length: usize, inverse: f64) !CommodityChannelIndex {
    var cci = try CommodityChannelIndex.init(allocator, .{ .length = length, .inverse_scaling_factor = inverse });
    cci.fixSlices();
    return cci;
}

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) < tolerance;
}

// Test data from TA-Lib (252 entries), typical price input.
test "cci length 11" {
    const tolerance = 5e-8;
    const input = testdata.testInput();

    var cci = try createCci(testing.allocator, 11);
    defer cci.deinit();

    // First 10 values should be NaN.
    for (0..10) |i| {
        const v = cci.update(input[i]);
        try testing.expect(math.isNan(v));
    }

    // Index 10: first value.
    var v = cci.update(input[10]);
    try testing.expect(!math.isNan(v));
    try testing.expect(almostEqual(v, 87.92686612269590, tolerance));

    // Index 11.
    v = cci.update(input[11]);
    try testing.expect(almostEqual(v, 180.00543014506300, tolerance));

    // Feed remaining and check last.
    for (12..251) |i| {
        _ = cci.update(input[i]);
    }

    v = cci.update(input[251]);
    try testing.expect(almostEqual(v, -169.65514382823800, tolerance));

    try testing.expect(cci.isPrimed());
}

test "cci length 2" {
    const tolerance = 5e-7;
    const input = testdata.testInput();

    var cci = try createCci(testing.allocator, 2);
    defer cci.deinit();

    // First value should be NaN.
    var v = cci.update(input[0]);
    try testing.expect(math.isNan(v));

    // Index 1: first value.
    v = cci.update(input[1]);
    try testing.expect(!math.isNan(v));
    try testing.expect(almostEqual(v, 66.66666666666670, tolerance));

    // Feed remaining and check last.
    for (2..251) |i| {
        _ = cci.update(input[i]);
    }

    v = cci.update(input[251]);
    try testing.expect(almostEqual(v, -66.66666666666590, tolerance));
}

test "cci is primed" {
    var cci = try createCci(testing.allocator, 5);
    defer cci.deinit();

    try testing.expect(!cci.isPrimed());

    for (1..5) |i| {
        _ = cci.update(@floatFromInt(i));
        try testing.expect(!cci.isPrimed());
    }

    _ = cci.update(5);
    try testing.expect(cci.isPrimed());

    _ = cci.update(6);
    try testing.expect(cci.isPrimed());
}

test "cci NaN" {
    var cci = try createCci(testing.allocator, 5);
    defer cci.deinit();

    const v = cci.update(math.nan(f64));
    try testing.expect(math.isNan(v));
}

test "cci metadata" {
    var cci = try CommodityChannelIndex.init(testing.allocator, .{ .length = 20 });
    defer cci.deinit();
    cci.fixSlices();

    var m: Metadata = undefined;
    cci.getMetadata(&m);

    try testing.expectEqual(Identifier.commodity_channel_index, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("cci(20)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Commodity Channel Index cci(20)", m.outputs_buf[0].description);
}

test "cci update entity" {
    const input = testdata.testInput();

    var cci = try createCci(testing.allocator, 11);
    defer cci.deinit();

    const time: i64 = 1617235200;

    for (0..10) |i| {
        const scalar = Scalar{ .time = time, .value = input[i] };
        const out = cci.updateScalar(&scalar);
        const v = out.slice()[0].scalar.value;
        try testing.expect(math.isNan(v));
    }

    const scalar = Scalar{ .time = time, .value = input[10] };
    const out = cci.updateScalar(&scalar);
    const v = out.slice()[0].scalar.value;
    try testing.expect(!math.isNan(v));
}

test "cci invalid params" {
    const result1 = CommodityChannelIndex.init(testing.allocator, .{ .length = 1 });
    try testing.expect(result1 == error.InvalidLength);

    const result2 = CommodityChannelIndex.init(testing.allocator, .{ .length = 0 });
    try testing.expect(result2 == error.InvalidLength);
}

test "cci custom scaling factor" {
    var cci = try createCciWithInverse(testing.allocator, 5, 0.03);
    defer cci.deinit();

    for (1..6) |i| {
        _ = cci.update(@floatFromInt(i));
    }

    try testing.expect(cci.isPrimed());
}
