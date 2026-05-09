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

/// Enumerates the outputs of the momentum indicator.
pub const MomentumOutput = enum(u8) {
    /// The scalar value of the momentum.
    value = 1,
};

/// Parameters to create an instance of the momentum indicator.
pub const MomentumParams = struct {
    /// The length (number of time periods). Must be >= 1.
    length: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the momentum (MOM).
///
/// MOMi = Pi - Pi-l
///
/// where l is the length.
///
/// The indicator is not primed during the first l updates.
pub const Momentum = struct {
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

    pub fn init(allocator: std.mem.Allocator, params: MomentumParams) !Momentum {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "mom({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Momentum {s}", .{mnemonic_slice}) catch
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

    pub fn deinit(self: *Momentum) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *Momentum) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update logic. Returns the momentum value or NaN if not yet primed.
    pub fn update(self: *Momentum, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            var i: usize = 0;
            while (i < self.last_index) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }

            self.window[self.last_index] = sample;

            return sample - self.window[0];
        }

        self.window[self.window_count] = sample;
        self.window_count += 1;

        if (self.window_length == self.window_count) {
            self.primed = true;

            return sample - self.window[0];
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const Momentum) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const Momentum, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .momentum,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *Momentum, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *Momentum, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *Momentum, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *Momentum, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *Momentum) indicator_mod.Indicator {
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
        const self: *Momentum = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const Momentum = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *Momentum = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *Momentum = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *Momentum = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *Momentum = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidLength,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");


fn createMomentum(allocator: std.mem.Allocator, length: usize) !Momentum {
    var mom = try Momentum.init(allocator, .{ .length = length });
    mom.fixSlices();
    return mom;
}

test "momentum update length 14" {
    const input = testdata.testInput();
    var mom = try createMomentum(testing.allocator, 14);
    defer mom.deinit();

    // First 13 updates (index 0..12) produce NaN.
    for (0..13) |i| {
        const act = mom.update(input[i]);
        try testing.expect(math.isNan(act));
    }

    // From index 13 onward, primed.
    for (13..252) |i| {
        const act = mom.update(input[i]);

        if (i == 14) try testing.expect(@abs(act - (-0.50)) < 1e-13);
        if (i == 15) try testing.expect(@abs(act - (-2.00)) < 1e-13);
        if (i == 16) try testing.expect(@abs(act - (-5.22)) < 1e-13);
        if (i == 251) try testing.expect(@abs(act - (-1.13)) < 1e-13);
    }

    // NaN passthrough
    try testing.expect(math.isNan(mom.update(math.nan(f64))));
}

test "momentum is primed" {
    const input = testdata.testInput();

    inline for ([_]usize{ 1, 2, 3, 5, 10 }) |length| {
        var mom = try createMomentum(testing.allocator, length);
        defer mom.deinit();

        try testing.expect(!mom.isPrimed());

        for (0..length) |i| {
            _ = mom.update(input[i]);
            try testing.expect(!mom.isPrimed());
        }

        for (length..252) |i| {
            _ = mom.update(input[i]);
            try testing.expect(mom.isPrimed());
        }
    }
}

test "momentum metadata" {
    var mom = try createMomentum(testing.allocator, 5);
    defer mom.deinit();
    var m: Metadata = undefined;
    mom.getMetadata(&m);

    try testing.expectEqual(Identifier.momentum, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("mom(5)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Momentum mom(5)", m.outputs_buf[0].description);
}

test "momentum update entity" {
    const length: usize = 2;
    const inp: f64 = 3.0;
    const exp: f64 = 3.0; // mom = 3.0 - 0.0 = 3.0
    const time: i64 = 1617235200;

    // scalar
    {
        var mom = try createMomentum(testing.allocator, length);
        defer mom.deinit();
        _ = mom.update(0.0);
        _ = mom.update(0.0);
        const out = mom.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expectEqual(exp, s.value);
    }

    // bar
    {
        var mom = try createMomentum(testing.allocator, length);
        defer mom.deinit();
        _ = mom.update(0.0);
        _ = mom.update(0.0);
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = mom.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(exp, s.value);
    }

    // quote
    {
        var mom = try createMomentum(testing.allocator, length);
        defer mom.deinit();
        _ = mom.update(0.0);
        _ = mom.update(0.0);
        const quote = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };
        const out = mom.updateQuote(&quote);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(exp, s.value);
    }

    // trade
    {
        var mom = try createMomentum(testing.allocator, length);
        defer mom.deinit();
        _ = mom.update(0.0);
        _ = mom.update(0.0);
        const trade = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = mom.updateTrade(&trade);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(exp, s.value);
    }
}

test "momentum init invalid length" {
    const result = Momentum.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, result);
}

test "momentum mnemonic components" {
    // all defaults -> no component suffix
    {
        var mom = try createMomentum(testing.allocator, 5);
        defer mom.deinit();
        try testing.expectEqualStrings("mom(5)", mom.line.mnemonic);
    }

    // bar component set to Median
    {
        var mom = try Momentum.init(testing.allocator, .{
            .length = 5,
            .bar_component = .median,
        });
        defer mom.deinit();
        mom.fixSlices();
        try testing.expectEqualStrings("mom(5, hl/2)", mom.line.mnemonic);
        try testing.expectEqualStrings("Momentum mom(5, hl/2)", mom.line.description);
    }

    // bar=high, trade=volume
    {
        var mom = try Momentum.init(testing.allocator, .{
            .length = 5,
            .bar_component = .high,
            .trade_component = .volume,
        });
        defer mom.deinit();
        mom.fixSlices();
        try testing.expectEqualStrings("mom(5, h, v)", mom.line.mnemonic);
    }
}
