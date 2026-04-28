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

pub const RateOfChangeRatioOutput = enum(u8) {
    value = 1,
};

pub const RateOfChangeRatioParams = struct {
    length: usize,
    hundred_scale: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Rate of Change Ratio (ROCR / ROCR100).
///
/// ROCR_i = P_i / P_{i-l}  (or * 100 when hundred_scale is true)
///
/// The indicator is not primed during the first l updates.
pub const RateOfChangeRatio = struct {
    line: LineIndicator,
    window: []f64,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    hundred_scale: bool,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    const epsilon = 1e-13;

    pub fn init(allocator: std.mem.Allocator, params: RateOfChangeRatioParams) !RateOfChangeRatio {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = if (params.hundred_scale)
            std.fmt.bufPrint(&mnemonic_buf, "rocr100({d}{s})", .{ params.length, triple }) catch
                return error.MnemonicTooLong
        else
            std.fmt.bufPrint(&mnemonic_buf, "rocr({d}{s})", .{ params.length, triple }) catch
                return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        const desc_prefix: []const u8 = if (params.hundred_scale) "Rate of Change Ratio 100 Scale " else "Rate of Change Ratio ";

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "{s}{s}", .{ desc_prefix, mnemonic_slice }) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window_length = params.length + 1;
        const window = try allocator.alloc(f64, window_length);
        @memset(window, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .window = window,
            .window_length = window_length,
            .window_count = 0,
            .last_index = params.length,
            .hundred_scale = params.hundred_scale,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *RateOfChangeRatio) void {
        self.allocator.free(self.window);
    }

    fn fixSlices(self: *RateOfChangeRatio) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    fn compute(self: *const RateOfChangeRatio, sample: f64, previous: f64) f64 {
        const scale: f64 = if (self.hundred_scale) 100.0 else 1.0;
        if (@abs(previous) > epsilon) {
            return (sample / previous) * scale;
        }
        return 0.0;
    }

    pub fn update(self: *RateOfChangeRatio, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            if (self.last_index > 1) {
                var i: usize = 0;
                while (i < self.last_index) : (i += 1) {
                    self.window[i] = self.window[i + 1];
                }
            }
            self.window[self.last_index] = sample;
            return self.compute(sample, self.window[0]);
        }

        self.window[self.window_count] = sample;
        self.window_count += 1;

        if (self.window_length == self.window_count) {
            self.primed = true;
            return self.compute(sample, self.window[0]);
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const RateOfChangeRatio) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const RateOfChangeRatio, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .rate_of_change_ratio,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *RateOfChangeRatio, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *RateOfChangeRatio, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *RateOfChangeRatio, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *RateOfChangeRatio, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *RateOfChangeRatio) indicator_mod.Indicator {
        return .{ .ptr = @ptrCast(self), .vtable = &vtable };
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
        const self: *RateOfChangeRatio = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const RateOfChangeRatio = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *RateOfChangeRatio = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *RateOfChangeRatio = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *RateOfChangeRatio = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *RateOfChangeRatio = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{ InvalidLength, MnemonicTooLong, OutOfMemory };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn testInput() [252]f64 {
    return .{
        91.500000,  94.815000,  94.375000,  95.095000,  93.780000,  94.625000,  92.530000,  92.750000,  90.315000,  92.470000,
        96.125000,  97.250000,  98.500000,  89.875000,  91.000000,  92.815000,  89.155000,  89.345000,  91.625000,  89.875000,
        88.375000,  87.625000,  84.780000,  83.000000,  83.500000,  81.375000,  84.440000,  89.250000,  86.375000,  86.250000,
        85.250000,  87.125000,  85.815000,  88.970000,  88.470000,  86.875000,  86.815000,  84.875000,  84.190000,  83.875000,
        83.375000,  85.500000,  89.190000,  89.440000,  91.095000,  90.750000,  91.440000,  89.000000,  91.000000,  90.500000,
        89.030000,  88.815000,  84.280000,  83.500000,  82.690000,  84.750000,  85.655000,  86.190000,  88.940000,  89.280000,
        88.625000,  88.500000,  91.970000,  91.500000,  93.250000,  93.500000,  93.155000,  91.720000,  90.000000,  89.690000,
        88.875000,  85.190000,  83.375000,  84.875000,  85.940000,  97.250000,  99.875000,  104.940000, 106.000000, 102.500000,
        102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000, 112.750000,
        123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000, 110.595000, 118.125000,
        116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000, 116.620000, 117.000000, 115.250000,
        114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000, 123.370000, 122.940000, 122.560000,
        123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000, 132.810000, 134.000000, 137.380000,
        137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000,
        123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000,
        122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000,
        124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
        132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000, 130.130000,
        127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000, 117.750000, 119.870000,
        122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000, 107.000000, 107.870000, 107.000000,
        107.120000, 107.000000, 91.000000,  93.940000,  93.870000,  95.500000,  93.000000,  94.940000,  98.250000,  96.750000,
        94.810000,  94.370000,  91.560000,  90.250000,  93.940000,  93.620000,  97.000000,  95.000000,  95.870000,  94.060000,
        94.620000,  93.750000,  98.000000,  103.940000, 107.870000, 106.060000, 104.500000, 105.000000, 104.190000, 103.060000,
        103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000, 109.700000, 109.250000,
        107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000, 109.810000, 109.000000,
        108.750000, 107.870000,
    };
}

fn createRocr(allocator: std.mem.Allocator, length: usize, hundred_scale: bool) !RateOfChangeRatio {
    var rocr = try RateOfChangeRatio.init(allocator, .{ .length = length, .hundred_scale = hundred_scale });
    rocr.fixSlices();
    return rocr;
}

test "rocr update length 14" {
    const input = testInput();
    var rocr = try createRocr(testing.allocator, 14, false);
    defer rocr.deinit();

    for (0..13) |i| {
        try testing.expect(math.isNan(rocr.update(input[i])));
    }

    for (13..252) |i| {
        const act = rocr.update(input[i]);
        if (i == 14) try testing.expect(@abs(act - 0.994536) < 1e-4);
        if (i == 15) try testing.expect(@abs(act - 0.978906) < 1e-4);
        if (i == 16) try testing.expect(@abs(act - 0.944689) < 1e-4);
        if (i == 251) try testing.expect(@abs(act - 0.989633) < 1e-4);
    }

    try testing.expect(math.isNan(rocr.update(math.nan(f64))));
}

test "rocr100 update length 14" {
    const input = testInput();
    var rocr = try createRocr(testing.allocator, 14, true);
    defer rocr.deinit();

    for (0..13) |i| {
        try testing.expect(math.isNan(rocr.update(input[i])));
    }

    for (13..252) |i| {
        const act = rocr.update(input[i]);
        if (i == 14) try testing.expect(@abs(act - 99.4536) < 1e-2);
        if (i == 15) try testing.expect(@abs(act - 97.8906) < 1e-2);
        if (i == 16) try testing.expect(@abs(act - 94.4689) < 1e-2);
        if (i == 251) try testing.expect(@abs(act - 98.9633) < 1e-2);
    }

    try testing.expect(math.isNan(rocr.update(math.nan(f64))));
}

test "rocr is primed" {
    const input = testInput();
    inline for ([_]usize{ 1, 2, 5, 10 }) |length| {
        var rocr = try createRocr(testing.allocator, length, false);
        defer rocr.deinit();
        try testing.expect(!rocr.isPrimed());
        for (0..length) |i| {
            _ = rocr.update(input[i]);
            try testing.expect(!rocr.isPrimed());
        }
        for (length..252) |i| {
            _ = rocr.update(input[i]);
            try testing.expect(rocr.isPrimed());
        }
    }
}

test "rocr metadata" {
    var rocr = try createRocr(testing.allocator, 5, false);
    defer rocr.deinit();
    var m: Metadata = undefined;
    rocr.getMetadata(&m);
    try testing.expectEqual(Identifier.rate_of_change_ratio, m.identifier);
    try testing.expectEqualStrings("rocr(5)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Rate of Change Ratio rocr(5)", m.outputs_buf[0].description);
}

test "rocr100 metadata" {
    var rocr = try createRocr(testing.allocator, 5, true);
    defer rocr.deinit();
    var m: Metadata = undefined;
    rocr.getMetadata(&m);
    try testing.expectEqualStrings("rocr100(5)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Rate of Change Ratio 100 Scale rocr100(5)", m.outputs_buf[0].description);
}

test "rocr init invalid length" {
    try testing.expectError(error.InvalidLength, RateOfChangeRatio.init(testing.allocator, .{ .length = 0 }));
}

test "rocr update entity" {
    const length: usize = 2;
    const inp: f64 = 3.0;
    const exp: f64 = 1.0; // 3/3 = 1.0
    const time: i64 = 1617235200;

    var rocr = try createRocr(testing.allocator, length, false);
    defer rocr.deinit();
    _ = rocr.update(inp);
    _ = rocr.update(inp);
    const out = rocr.updateScalar(&.{ .time = time, .value = inp });
    try testing.expectEqual(@as(usize, 1), out.len);
    const s = out.slice()[0].scalar;
    try testing.expectEqual(time, s.time);
    try testing.expect(@abs(s.value - exp) < 1e-13);
}
