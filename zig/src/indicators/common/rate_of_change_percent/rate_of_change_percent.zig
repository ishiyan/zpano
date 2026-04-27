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

pub const RateOfChangePercentOutput = enum(u8) {
    value = 1,
};

pub const RateOfChangePercentParams = struct {
    length: usize,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Rate of Change Percent (ROCP).
///
/// ROCP_i = P_i / P_{i-l} - 1
///
/// The indicator is not primed during the first l updates.
pub const RateOfChangePercent = struct {
    line: LineIndicator,
    window: []f64,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    const epsilon = 1e-13;

    pub fn init(allocator: std.mem.Allocator, params: RateOfChangePercentParams) !RateOfChangePercent {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "rocp({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Rate of Change Percent {s}", .{mnemonic_slice}) catch
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
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *RateOfChangePercent) void {
        self.allocator.free(self.window);
    }

    fn fixSlices(self: *RateOfChangePercent) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    fn compute(sample: f64, previous: f64) f64 {
        if (@abs(previous) > epsilon) {
            return sample / previous - 1.0;
        }
        return 0.0;
    }

    pub fn update(self: *RateOfChangePercent, sample: f64) f64 {
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
            return compute(sample, self.window[0]);
        }

        self.window[self.window_count] = sample;
        self.window_count += 1;

        if (self.window_length == self.window_count) {
            self.primed = true;
            return compute(sample, self.window[0]);
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const RateOfChangePercent) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const RateOfChangePercent) Metadata {
        return build_metadata_mod.buildMetadata(
            .rate_of_change_percent,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *RateOfChangePercent, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *RateOfChangePercent, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *RateOfChangePercent, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *RateOfChangePercent, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *RateOfChangePercent) indicator_mod.Indicator {
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
        const self: *RateOfChangePercent = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque) Metadata {
        const self: *const RateOfChangePercent = @ptrCast(@alignCast(ptr));
        return self.getMetadata();
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *RateOfChangePercent = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *RateOfChangePercent = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *RateOfChangePercent = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *RateOfChangePercent = @ptrCast(@alignCast(ptr));
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

fn createRocp(allocator: std.mem.Allocator, length: usize) !RateOfChangePercent {
    var rocp = try RateOfChangePercent.init(allocator, .{ .length = length });
    rocp.fixSlices();
    return rocp;
}

test "rocp update length 14" {
    const input = testInput();
    var rocp = try createRocp(testing.allocator, 14);
    defer rocp.deinit();

    for (0..13) |i| {
        try testing.expect(math.isNan(rocp.update(input[i])));
    }

    for (13..252) |i| {
        const act = rocp.update(input[i]);
        if (i == 14) try testing.expect(@abs(act - (-0.00546)) < 1e-4);
        if (i == 15) try testing.expect(@abs(act - (-0.02109)) < 1e-4);
        if (i == 16) try testing.expect(@abs(act - (-0.0553)) < 1e-4);
        if (i == 251) try testing.expect(@abs(act - (-0.010367)) < 1e-4);
    }

    try testing.expect(math.isNan(rocp.update(math.nan(f64))));
}

test "rocp is primed" {
    const input = testInput();
    inline for ([_]usize{ 1, 2, 5, 10 }) |length| {
        var rocp = try createRocp(testing.allocator, length);
        defer rocp.deinit();
        try testing.expect(!rocp.isPrimed());
        for (0..length) |i| {
            _ = rocp.update(input[i]);
            try testing.expect(!rocp.isPrimed());
        }
        for (length..252) |i| {
            _ = rocp.update(input[i]);
            try testing.expect(rocp.isPrimed());
        }
    }
}

test "rocp metadata" {
    var rocp = try createRocp(testing.allocator, 5);
    defer rocp.deinit();
    const m = rocp.getMetadata();
    try testing.expectEqual(Identifier.rate_of_change_percent, m.identifier);
    try testing.expectEqualStrings("rocp(5)", m.outputs[0].mnemonic);
    try testing.expectEqualStrings("Rate of Change Percent rocp(5)", m.outputs[0].description);
}

test "rocp init invalid length" {
    try testing.expectError(error.InvalidLength, RateOfChangePercent.init(testing.allocator, .{ .length = 0 }));
}
