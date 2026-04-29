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
fn testInput() [252]f64 {
    return .{
        91.83333333333330,  93.72000000000000,  95.00000000000000,  94.92833333333330,  94.19833333333330,  94.28166666666670,  93.17666666666670,  92.07333333333330,  90.74166666666670,  91.94833333333330,
        95.04166666666670,  97.73000000000000,  97.88500000000000,  90.48000000000000,  89.68833333333330,  92.33500000000000,  90.48833333333330,  89.59333333333330,  90.94833333333330,  90.62500000000000,
        88.74000000000000,  87.55166666666670,  85.88500000000000,  83.59333333333330,  83.16833333333330,  82.33333333333330,  83.37666666666670,  87.57333333333330,  86.69833333333330,  87.11500000000000,
        85.60333333333330,  86.49000000000000,  86.23000000000000,  87.82333333333330,  88.78166666666670,  87.76166666666670,  86.14666666666670,  84.68833333333330,  83.83500000000000,  84.26166666666670,
        83.45833333333330,  86.35500000000000,  88.51166666666670,  89.32333333333330,  90.93833333333330,  90.77166666666670,  91.72000000000000,  89.90666666666670,  90.24000000000000,  90.78166666666670,
        89.34333333333330,  88.05333333333330,  85.76000000000000,  84.02166666666670,  82.83500000000000,  84.41666666666670,  85.67666666666670,  86.47000000000000,  88.50166666666670,  89.44833333333330,
        89.20833333333330,  88.23000000000000,  91.07333333333330,  91.99000000000000,  92.19833333333330,  92.89500000000000,  93.06166666666670,  91.35500000000000,  90.65666666666670,  90.14833333333330,
        88.45833333333330,  86.33500000000000,  83.85333333333330,  83.75000000000000,  84.81500000000000,  97.65666666666670,  99.87500000000000,  103.82333333333300, 105.95833333333300, 103.16666666666700,
        102.87500000000000, 103.93833333333300, 105.13500000000000, 106.54333333333300, 105.32333333333300, 105.20833333333300, 107.63500000000000, 109.59500000000000, 110.06333333333300, 111.57333333333300,
        121.00000000000000, 119.79166666666700, 118.18833333333300, 119.35500000000000, 117.94833333333300, 116.96000000000000, 115.49166666666700, 112.69833333333300, 111.36500000000000, 115.72000000000000,
        115.16333333333300, 115.64666666666700, 112.35333333333300, 112.60333333333300, 113.27000000000000, 114.81333333333300, 119.89666666666700, 117.51666666666700, 118.14333333333300, 115.10333333333300,
        114.45666666666700, 115.16666666666700, 116.31000000000000, 120.35333333333300, 120.39666666666700, 120.64666666666700, 124.37333333333300, 124.70666666666700, 122.96000000000000, 122.85333333333300,
        123.99666666666700, 123.14666666666700, 124.22666666666700, 128.54000000000000, 130.10333333333300, 131.33333333333300, 131.89666666666700, 132.31333333333300, 133.87666666666700, 136.23333333333300,
        137.29333333333300, 137.60666666666700, 137.31333333333300, 136.31333333333300, 136.37666666666700, 135.73333333333300, 128.81333333333300, 128.54000000000000, 125.29000000000000, 124.29000000000000,
        123.62333333333300, 125.43666666666700, 127.62666666666700, 125.53666666666700, 125.58333333333300, 123.35333333333300, 120.22666666666700, 119.47666666666700, 121.58333333333300, 123.83333333333300,
        122.58333333333300, 120.25000000000000, 122.29000000000000, 121.97666666666700, 123.29000000000000, 126.58000000000000, 127.87333333333300, 125.66666666666700, 123.02000000000000, 122.39333333333300,
        123.87333333333300, 122.20666666666700, 122.43333333333300, 123.62333333333300, 123.98000000000000, 123.83333333333300, 124.47666666666700, 127.08333333333300, 125.62333333333300, 128.87000000000000,
        131.02333333333300, 131.79333333333300, 134.29333333333300, 135.69000000000000, 133.31333333333300, 132.93666666666700, 132.96000000000000, 130.64666666666700, 126.85333333333300, 129.00333333333300,
        127.66666666666700, 125.60333333333300, 123.75000000000000, 123.48000000000000, 123.72666666666700, 123.31333333333300, 120.95666666666700, 120.95666666666700, 118.10333333333300, 118.87333333333300,
        121.43666666666700, 120.33333333333300, 116.87333333333300, 113.20666666666700, 114.68666666666700, 111.62333333333300, 106.97666666666700, 106.31333333333300, 106.87000000000000, 106.89666666666700,
        107.02000000000000, 109.02000000000000, 91.00000000000000,  93.68666666666670,  93.70333333333330,  95.37333333333330,  93.79000000000000,  94.83333333333330,  97.83333333333330,  97.31000000000000,
        95.10333333333330,  94.60333333333330,  92.00000000000000,  91.12666666666670,  92.79333333333330,  93.74666666666670,  96.06000000000000,  95.79000000000000,  95.04000000000000,  94.76666666666670,
        94.20666666666670,  93.74666666666670,  96.60333333333330,  102.47666666666700, 106.91666666666700, 107.31000000000000, 103.77000000000000, 105.04000000000000, 104.16666666666700, 103.22666666666700,
        103.37000000000000, 104.98333333333300, 110.89333333333300, 115.00000000000000, 117.08333333333300, 118.26000000000000, 115.91333333333300, 109.50000000000000, 109.67000000000000, 108.77000000000000,
        106.48000000000000, 108.21000000000000, 109.89333333333300, 109.13000000000000, 109.43333333333300, 108.77000000000000, 109.08333333333300, 109.29000000000000, 109.87333333333300, 109.41666666666700,
        109.27000000000000, 107.99666666666700,
    };
}

test "cci length 11" {
    const tolerance = 5e-8;
    const input = testInput();

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
    const input = testInput();

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
    const input = testInput();

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
