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

/// Enumerates the outputs of the rate of change indicator.
pub const RateOfChangeOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the rate of change indicator.
pub const RateOfChangeParams = struct {
    length: usize,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the Rate of Change (ROC).
///
/// ROCi = 100 * (Pi / Pi-l - 1)
///
/// The indicator is not primed during the first l updates.
pub const RateOfChange = struct {
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

    pub fn init(allocator: std.mem.Allocator, params: RateOfChangeParams) !RateOfChange {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "roc({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Rate of Change {s}", .{mnemonic_slice}) catch
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

    pub fn deinit(self: *RateOfChange) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *RateOfChange) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    fn compute(sample: f64, previous: f64) f64 {
        if (@abs(previous) > epsilon) {
            return (sample / previous - 1.0) * 100.0;
        }
        return 0.0;
    }

    pub fn update(self: *RateOfChange, sample: f64) f64 {
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

    pub fn isPrimed(self: *const RateOfChange) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const RateOfChange, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .rate_of_change,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *RateOfChange, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *RateOfChange, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *RateOfChange, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *RateOfChange, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *RateOfChange) indicator_mod.Indicator {
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
        const self: *RateOfChange = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const RateOfChange = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *RateOfChange = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *RateOfChange = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *RateOfChange = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *RateOfChange = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{ InvalidLength, MnemonicTooLong, OutOfMemory };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn createRoc(allocator: std.mem.Allocator, length: usize) !RateOfChange {
    var roc = try RateOfChange.init(allocator, .{ .length = length });
    roc.fixSlices();
    return roc;
}

test "roc update length 14" {
    const input = testdata.testInput();
    var roc = try createRoc(testing.allocator, 14);
    defer roc.deinit();

    for (0..13) |i| {
        try testing.expect(math.isNan(roc.update(input[i])));
    }

    for (13..252) |i| {
        const act = roc.update(input[i]);
        if (i == 14) try testing.expect(@abs(act - (-0.546)) < 1e-2);
        if (i == 15) try testing.expect(@abs(act - (-2.109)) < 1e-2);
        if (i == 16) try testing.expect(@abs(act - (-5.53)) < 1e-2);
        if (i == 251) try testing.expect(@abs(act - (-1.0367)) < 1e-2);
    }

    try testing.expect(math.isNan(roc.update(math.nan(f64))));
}

test "roc is primed" {
    const input = testdata.testInput();
    inline for ([_]usize{ 1, 2, 5, 10 }) |length| {
        var roc = try createRoc(testing.allocator, length);
        defer roc.deinit();
        try testing.expect(!roc.isPrimed());
        for (0..length) |i| {
            _ = roc.update(input[i]);
            try testing.expect(!roc.isPrimed());
        }
        for (length..252) |i| {
            _ = roc.update(input[i]);
            try testing.expect(roc.isPrimed());
        }
    }
}

test "roc metadata" {
    var roc = try createRoc(testing.allocator, 5);
    defer roc.deinit();
    var m: Metadata = undefined;
    roc.getMetadata(&m);
    try testing.expectEqual(Identifier.rate_of_change, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqualStrings("roc(5)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Rate of Change roc(5)", m.outputs_buf[0].description);
}

test "roc init invalid length" {
    try testing.expectError(error.InvalidLength, RateOfChange.init(testing.allocator, .{ .length = 0 }));
}

test "roc update entity" {
    const length: usize = 2;
    const inp: f64 = 3.0;
    const exp: f64 = 0.0; // (3/3 - 1) * 100 = 0
    const time: i64 = 1617235200;

    // scalar
    {
        var roc = try createRoc(testing.allocator, length);
        defer roc.deinit();
        _ = roc.update(inp);
        _ = roc.update(inp);
        const out = roc.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(@abs(s.value - exp) < 1e-13);
    }
}
