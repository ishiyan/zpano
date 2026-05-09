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

    pub fn fixSlices(self: *RateOfChangePercent) void {
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

    pub fn getMetadata(self: *const RateOfChangePercent, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
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
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const RateOfChangePercent = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
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
const testdata = @import("testdata.zig");


fn createRocp(allocator: std.mem.Allocator, length: usize) !RateOfChangePercent {
    var rocp = try RateOfChangePercent.init(allocator, .{ .length = length });
    rocp.fixSlices();
    return rocp;
}

test "rocp update length 14" {
    const input = testdata.testInput();
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
    const input = testdata.testInput();
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
    var m: Metadata = undefined;
    rocp.getMetadata(&m);
    try testing.expectEqual(Identifier.rate_of_change_percent, m.identifier);
    try testing.expectEqualStrings("rocp(5)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Rate of Change Percent rocp(5)", m.outputs_buf[0].description);
}

test "rocp init invalid length" {
    try testing.expectError(error.InvalidLength, RateOfChangePercent.init(testing.allocator, .{ .length = 0 }));
}
