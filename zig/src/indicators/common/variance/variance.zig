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

/// Enumerates the outputs of the variance indicator.
pub const VarianceOutput = enum(u8) {
    /// The scalar value of the variance.
    value = 1,
};

/// Parameters to create an instance of the variance indicator.
pub const VarianceParams = struct {
    /// The length (number of time periods). Must be >= 2.
    length: usize,
    /// Whether to compute the unbiased sample variance (true) or population variance (false).
    is_unbiased: bool = false,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the variance within a moving window of length l.
///
/// Population variance: σ² = (Σx² - (Σx)²/l) / l
/// Sample variance:     σ² = (Σx² - (Σx)²/l) / (l-1)
///
/// The indicator is not primed during the first l-1 updates.
pub const Variance = struct {
    line: LineIndicator,
    window: []f64,
    window_sum: f64,
    window_squared_sum: f64,
    window_length: usize,
    window_count: usize,
    last_index: usize,
    primed: bool,
    unbiased: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: VarianceParams) !Variance {
        if (params.length < 2) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        const c: u8 = if (params.is_unbiased) 's' else 'p';

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "var.{c}({d}{s})", .{ c, params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_prefix: []const u8 = if (params.is_unbiased)
            "Unbiased estimation of the sample variance "
        else
            "Estimation of the population variance ";
        const desc_slice = std.fmt.bufPrint(&description_buf, "{s}{s}", .{ desc_prefix, mnemonic_slice }) catch
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
            .window = window,
            .window_sum = 0.0,
            .window_squared_sum = 0.0,
            .window_length = params.length,
            .window_count = 0,
            .last_index = params.length - 1,
            .primed = false,
            .unbiased = params.is_unbiased,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *Variance) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *Variance) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update logic. Returns the variance value or NaN if not yet primed.
    pub fn update(self: *Variance, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        var temp: f64 = sample;
        const wlen: f64 = @floatFromInt(self.window_length);

        if (self.primed) {
            self.window_sum += temp;
            temp *= temp;
            self.window_squared_sum += temp;
            temp = self.window[0];
            self.window_sum -= temp;
            temp *= temp;
            self.window_squared_sum -= temp;

            var value: f64 = undefined;
            if (self.unbiased) {
                temp = self.window_sum;
                temp *= temp;
                temp /= wlen;
                value = self.window_squared_sum - temp;
                value /= @floatFromInt(self.last_index);
            } else {
                temp = self.window_sum / wlen;
                temp *= temp;
                value = self.window_squared_sum / wlen - temp;
            }

            var i: usize = 0;
            while (i < self.last_index) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }
            self.window[self.last_index] = sample;

            return value;
        }

        self.window_sum += temp;
        self.window[self.window_count] = temp;
        temp *= temp;
        self.window_squared_sum += temp;

        self.window_count += 1;
        if (self.window_length == self.window_count) {
            self.primed = true;

            var value: f64 = undefined;
            if (self.unbiased) {
                temp = self.window_sum;
                temp *= temp;
                temp /= wlen;
                value = self.window_squared_sum - temp;
                value /= @floatFromInt(self.last_index);
            } else {
                temp = self.window_sum / wlen;
                temp *= temp;
                value = self.window_squared_sum / wlen - temp;
            }

            return value;
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const Variance) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const Variance, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .variance,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *Variance, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *Variance, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *Variance, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *Variance, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *Variance) indicator_mod.Indicator {
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
        const self: *Variance = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const Variance = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *Variance = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *Variance = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *Variance = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *Variance = @ptrCast(@alignCast(ptr));
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

fn testInput() [12]f64 {
    return .{ 1, 2, 8, 4, 9, 6, 7, 13, 9, 10, 3, 12 };
}

fn createVariance(allocator: std.mem.Allocator, length: usize, unbiased: bool) !Variance {
    var v = try Variance.init(allocator, .{ .length = length, .is_unbiased = unbiased });
    v.fixSlices();
    return v;
}

fn almostEqual(a: f64, b: f64, eps: f64) bool {
    return @abs(a - b) < eps;
}

test "variance population length 3" {
    const input = testInput();
    var v = try createVariance(testing.allocator, 3, false);
    defer v.deinit();

    const expected = [_]f64{
        0,                   0, // NaN placeholders (unused)
        9.55555555555556000, 6.22222222222222000,
        4.66666666666667000, 4.22222222222222000,
        1.55555555555556000, 9.55555555555556000,
        6.22222222222222000, 2.88888888888889000,
        9.55555555555556000, 14.88888888888890000,
    };

    for (0..2) |i| {
        try testing.expect(math.isNan(v.update(input[i])));
    }
    for (2..12) |i| {
        const act = v.update(input[i]);
        try testing.expect(almostEqual(act, expected[i], 1e-13));
    }
    try testing.expect(math.isNan(v.update(math.nan(f64))));
}

test "variance population length 5" {
    const input = testInput();
    var v = try createVariance(testing.allocator, 5, false);
    defer v.deinit();

    const expected = [_]f64{
        0,        0,       0,        0, // NaN placeholders
        10.16000, 6.56000, 2.96000,  9.36000,
        5.76000,  6.00000, 11.04000, 12.24000,
    };

    for (0..4) |i| {
        try testing.expect(math.isNan(v.update(input[i])));
    }
    for (4..12) |i| {
        const act = v.update(input[i]);
        try testing.expect(almostEqual(act, expected[i], 1e-13));
    }
}

test "variance sample length 3" {
    const input = testInput();
    var v = try createVariance(testing.allocator, 3, true);
    defer v.deinit();

    const expected = [_]f64{
        0,                   0,
        14.3333333333333000, 9.3333333333333400,
        7.0000000000000000,  6.3333333333333400,
        2.3333333333333300,  14.3333333333333000,
        9.3333333333333400,  4.3333333333333400,
        14.3333333333333000, 22.3333333333333000,
    };

    for (0..2) |i| {
        try testing.expect(math.isNan(v.update(input[i])));
    }
    for (2..12) |i| {
        const act = v.update(input[i]);
        try testing.expect(almostEqual(act, expected[i], 1e-13));
    }
    try testing.expect(math.isNan(v.update(math.nan(f64))));
}

test "variance is primed" {
    const input = testInput();
    var v = try createVariance(testing.allocator, 3, false);
    defer v.deinit();

    try testing.expect(!v.isPrimed());
    for (0..2) |_| {
        _ = v.update(input[0]);
        try testing.expect(!v.isPrimed());
    }
    _ = v.update(input[2]);
    try testing.expect(v.isPrimed());
}

test "variance metadata population" {
    var v = try createVariance(testing.allocator, 7, false);
    defer v.deinit();
    var m: Metadata = undefined;
    v.getMetadata(&m);

    try testing.expectEqual(Identifier.variance, m.identifier);
    try testing.expectEqualStrings("var.p(7)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Estimation of the population variance var.p(7)", m.outputs_buf[0].description);
}

test "variance metadata sample" {
    var v = try createVariance(testing.allocator, 7, true);
    defer v.deinit();
    var m: Metadata = undefined;
    v.getMetadata(&m);

    try testing.expectEqual(Identifier.variance, m.identifier);
    try testing.expectEqualStrings("var.s(7)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Unbiased estimation of the sample variance var.s(7)", m.outputs_buf[0].description);
}

test "variance update entity" {
    const length: usize = 3;
    const inp: f64 = 3.0;
    const exp: f64 = inp * inp / @as(f64, @floatFromInt(length));
    const time: i64 = 1617235200;

    // scalar
    {
        var v = try createVariance(testing.allocator, length, true);
        defer v.deinit();
        _ = v.update(0.0);
        _ = v.update(0.0);
        const out = v.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(almostEqual(s.value, exp, 1e-13));
    }

    // bar
    {
        var v = try createVariance(testing.allocator, length, true);
        defer v.deinit();
        _ = v.update(0.0);
        _ = v.update(0.0);
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = v.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(almostEqual(s.value, exp, 1e-13));
    }
}

test "variance init invalid length" {
    const r1 = Variance.init(testing.allocator, .{ .length = 1 });
    try testing.expectError(error.InvalidLength, r1);
    const r0 = Variance.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, r0);
}

test "variance mnemonic components" {
    {
        var v = try createVariance(testing.allocator, 5, true);
        defer v.deinit();
        try testing.expectEqualStrings("var.s(5)", v.line.mnemonic);
    }
    {
        var v = try Variance.init(testing.allocator, .{
            .length = 5,
            .is_unbiased = true,
            .bar_component = .median,
        });
        defer v.deinit();
        v.fixSlices();
        try testing.expectEqualStrings("var.s(5, hl/2)", v.line.mnemonic);
    }
}
