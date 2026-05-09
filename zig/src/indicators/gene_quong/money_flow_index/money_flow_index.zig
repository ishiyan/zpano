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
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Money Flow Index indicator.
pub const MoneyFlowIndexOutput = enum(u8) {
    /// The scalar value of the money flow index.
    value = 1,
};

/// Parameters to create an instance of the Money Flow Index indicator.
pub const MoneyFlowIndexParams = struct {
    /// The number of time periods. Must be >= 1.
    length: u32 = 14,
    /// Bar component to extract. `null` means use default (Typical).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Gene Quong's Money Flow Index (MFI).
///
/// MFI is a volume-weighted oscillator calculated over ℓ periods, showing money flow
/// on up days as a percentage of the total of up and down days.
///
///   TypicalPrice = (High + Low + Close) / 3
///   MoneyFlow = TypicalPrice × Volume
///   MFI = 100 × PositiveMoneyFlow / (PositiveMoneyFlow + NegativeMoneyFlow)
pub const MoneyFlowIndex = struct {
    line: LineIndicator,
    bar_func: *const fn (Bar) f64,
    length: u32,
    negative_buffer: []f64,
    positive_buffer: []f64,
    negative_sum: f64,
    positive_sum: f64,
    previous_sample: f64,
    buffer_index: u32,
    buffer_low_index: u32,
    buffer_count: u32,
    value: f64,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: MoneyFlowIndexParams) !MoneyFlowIndex {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse .typical;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "mfi({d}{s})", .{ params.length, triple }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        const desc = "Money Flow Index " ++ "mfi";

        const neg_buf = try allocator.alloc(f64, params.length);
        @memset(neg_buf, 0);
        const pos_buf = try allocator.alloc(f64, params.length);
        @memset(pos_buf, 0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                desc,
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .bar_func = bar_component.componentValue(bc),
            .length = params.length,
            .negative_buffer = neg_buf,
            .positive_buffer = pos_buf,
            .negative_sum = 0,
            .positive_sum = 0,
            .previous_sample = 0,
            .buffer_index = 0,
            .buffer_low_index = 0,
            .buffer_count = 0,
            .value = math.nan(f64),
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn deinit(self: *MoneyFlowIndex) void {
        self.allocator.free(self.negative_buffer);
        self.allocator.free(self.positive_buffer);
    }

    pub fn fixSlices(self: *MoneyFlowIndex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    /// Update with volume = 1 (scalar path).
    pub fn update(self: *MoneyFlowIndex, sample: f64) f64 {
        return self.updateWithVolume(sample, 1);
    }

    /// Update with the given sample and volume.
    pub fn updateWithVolume(self: *MoneyFlowIndex, sample: f64, volume: f64) f64 {
        if (math.isNan(sample) or math.isNan(volume)) {
            return math.nan(f64);
        }

        const length_min_one = self.length - 1;

        if (self.primed) {
            self.negative_sum -= self.negative_buffer[self.buffer_low_index];
            self.positive_sum -= self.positive_buffer[self.buffer_low_index];

            const amount = sample * volume;
            const diff = sample - self.previous_sample;

            if (diff < 0) {
                self.negative_buffer[self.buffer_index] = amount;
                self.positive_buffer[self.buffer_index] = 0;
                self.negative_sum += amount;
            } else if (diff > 0) {
                self.negative_buffer[self.buffer_index] = 0;
                self.positive_buffer[self.buffer_index] = amount;
                self.positive_sum += amount;
            } else {
                self.negative_buffer[self.buffer_index] = 0;
                self.positive_buffer[self.buffer_index] = 0;
            }

            const sum = self.positive_sum + self.negative_sum;
            if (sum < 1) {
                self.value = 0;
            } else {
                self.value = 100 * self.positive_sum / sum;
            }

            self.buffer_index += 1;
            if (self.buffer_index > length_min_one) {
                self.buffer_index = 0;
            }

            self.buffer_low_index += 1;
            if (self.buffer_low_index > length_min_one) {
                self.buffer_low_index = 0;
            }
        } else if (self.buffer_count == 0) {
            self.buffer_count += 1;
        } else {
            const amount = sample * volume;
            const diff = sample - self.previous_sample;

            if (diff < 0) {
                self.negative_buffer[self.buffer_index] = amount;
                self.positive_buffer[self.buffer_index] = 0;
                self.negative_sum += amount;
            } else if (diff > 0) {
                self.negative_buffer[self.buffer_index] = 0;
                self.positive_buffer[self.buffer_index] = amount;
                self.positive_sum += amount;
            } else {
                self.negative_buffer[self.buffer_index] = 0;
                self.positive_buffer[self.buffer_index] = 0;
            }

            if (self.length == self.buffer_count) {
                const sum = self.positive_sum + self.negative_sum;
                if (sum < 1) {
                    self.value = 0;
                } else {
                    self.value = 100 * self.positive_sum / sum;
                }

                self.primed = true;
            }

            self.buffer_index += 1;
            if (self.buffer_index > length_min_one) {
                self.buffer_index = 0;
            }

            self.buffer_count += 1;
        }

        self.previous_sample = sample;

        return self.value;
    }

    pub fn isPrimed(self: *const MoneyFlowIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const MoneyFlowIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .money_flow_index,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *MoneyFlowIndex, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Shadows LineIndicator.updateBar to use bar volume.
    pub fn updateBar(self: *MoneyFlowIndex, sample: *const Bar) OutputArray {
        const price = self.bar_func(sample.*);
        const value = self.updateWithVolume(price, sample.volume);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *MoneyFlowIndex, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *MoneyFlowIndex, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *MoneyFlowIndex) indicator_mod.Indicator {
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
        const self: *MoneyFlowIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const MoneyFlowIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *MoneyFlowIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *MoneyFlowIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *MoneyFlowIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *MoneyFlowIndex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{InvalidLength} || std.mem.Allocator.Error;
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn roundTo(v: f64, comptime digits: comptime_int) f64 {
    const p = comptime blk: {
        var result: f64 = 1.0;
        for (0..digits) |_| {
            result *= 10.0;
        }
        break :blk result;
    };
    return @round(v * p) / p;
}

// Typical price test data: (high + low + close) / 3, 252 entries.
fn createMfi(allocator: std.mem.Allocator) !MoneyFlowIndex {
    var mfi = try MoneyFlowIndex.init(allocator, .{ .length = 14 });
    mfi.fixSlices();
    return mfi;
}

test "money flow index with volume" {
    const tp = testdata.testTypicalPrices();
    const vol = testdata.testVolumes();
    const expected = testdata.testExpectedMfi();
    const digits = 9;

    var mfi = try createMfi(testing.allocator);
    defer mfi.deinit();

    for (0..14) |i| {
        const v = mfi.updateWithVolume(tp[i], vol[i]);
        try testing.expect(math.isNan(v));
        try testing.expect(!mfi.isPrimed());
    }

    for (14..252) |i| {
        const v = mfi.updateWithVolume(tp[i], vol[i]);
        try testing.expect(!math.isNan(v));
        try testing.expect(mfi.isPrimed());

        const got = roundTo(v, digits);
        const exp = roundTo(expected[i], digits);
        try testing.expectEqual(exp, got);
    }
}

test "money flow index volume 1" {
    const tp = testdata.testTypicalPrices();
    const expected = testdata.testExpectedMfiVolume1();
    const digits = 9;

    var mfi = try createMfi(testing.allocator);
    defer mfi.deinit();

    for (0..14) |i| {
        const v = mfi.update(tp[i]);
        try testing.expect(math.isNan(v));
    }

    for (14..252) |i| {
        const v = mfi.update(tp[i]);
        try testing.expect(!math.isNan(v));

        const got = roundTo(v, digits);
        const exp = roundTo(expected[i], digits);
        try testing.expectEqual(exp, got);
    }
}

test "money flow index is primed" {
    var mfi = try MoneyFlowIndex.init(testing.allocator, .{ .length = 5 });
    defer mfi.deinit();
    mfi.fixSlices();

    try testing.expect(!mfi.isPrimed());

    // Feed 6 samples (5+1): first stores previousSample, next 5 fill buffer.
    for (1..6) |i| {
        _ = mfi.update(@floatFromInt(i));
        try testing.expect(!mfi.isPrimed());
    }

    _ = mfi.update(5);
    try testing.expect(mfi.isPrimed());

    _ = mfi.update(6);
    try testing.expect(mfi.isPrimed());
}

test "money flow index NaN" {
    var mfi = try MoneyFlowIndex.init(testing.allocator, .{ .length = 5 });
    defer mfi.deinit();
    mfi.fixSlices();

    try testing.expect(math.isNan(mfi.update(math.nan(f64))));
    try testing.expect(math.isNan(mfi.updateWithVolume(1.0, math.nan(f64))));
    try testing.expect(math.isNan(mfi.updateWithVolume(math.nan(f64), math.nan(f64))));
}

test "money flow index metadata" {
    var mfi = try createMfi(testing.allocator);
    defer mfi.deinit();

    var m: Metadata = undefined;
    mfi.getMetadata(&m);

    try testing.expectEqual(@import("../../core/identifier.zig").Identifier.money_flow_index, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("mfi(14, hlc/3)", m.mnemonic);
}

test "money flow index update bar" {
    const digits = 9;
    const input_high = [_]f64{
        93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000, 96.250000,
        99.625000, 99.125000, 92.750000, 91.315000,
    };
    const input_low = [_]f64{
        90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000, 92.750000,
        96.315000, 96.030000, 88.815000, 86.750000,
    };
    const input_close = [_]f64{
        91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000, 96.125000,
        97.250000, 98.500000, 89.875000, 91.000000,
    };
    const input_volume = [_]f64{
        4077500, 4955900, 4775300,  4155300,  4593100,  3631300, 3382800, 4954200, 4500000, 3397500,
        4204500, 6321400, 10203600, 19043900, 11692000,
    };

    var mfi = try createMfi(testing.allocator);
    defer mfi.deinit();

    const time: i64 = 1617235200;

    for (0..14) |i| {
        const bar = Bar{ .time = time, .open = 0, .high = input_high[i], .low = input_low[i], .close = input_close[i], .volume = input_volume[i] };
        const out = mfi.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(math.isNan(s.value));
    }

    // Index 14: first value with real volume via UpdateBar.
    const bar = Bar{ .time = time, .open = 0, .high = input_high[14], .low = input_low[14], .close = input_close[14], .volume = input_volume[14] };
    const out = mfi.updateBar(&bar);
    const s = out.slice()[0].scalar;
    try testing.expect(!math.isNan(s.value));

    const expected = testdata.testExpectedMfi();
    const got = roundTo(s.value, digits);
    const exp = roundTo(expected[14], digits);
    try testing.expectEqual(exp, got);
}

test "money flow index invalid params" {
    const result = MoneyFlowIndex.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, result);
}

test "money flow index small sum" {
    var mfi = try MoneyFlowIndex.init(testing.allocator, .{ .length = 2 });
    defer mfi.deinit();
    mfi.fixSlices();

    for (0..10) |_| {
        _ = mfi.updateWithVolume(0.001, 0.5);
    }

    try testing.expect(mfi.isPrimed());

    const v = mfi.updateWithVolume(0.001, 0.5);
    try testing.expectEqual(@as(f64, 0), v);
}
