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

/// Enumerates the outputs of the triangular moving average indicator.
pub const TriangularMovingAverageOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the triangular moving average indicator.
pub const TriangularMovingAverageParams = struct {
    length: usize,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the triangular moving average (TRIMA).
///
/// The TRIMA puts more weight on the data in the middle of the window,
/// equivalent to doing a SMA of a SMA.
///
/// The indicator is not primed during the first ℓ−1 updates.
pub const TriangularMovingAverage = struct {
    line: LineIndicator,
    factor: f64,
    numerator: f64,
    numerator_sub: f64,
    numerator_add: f64,
    window: []f64,
    window_length: usize,
    window_length_half: usize,
    window_count: usize,
    is_odd: bool,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: TriangularMovingAverageParams) !TriangularMovingAverage {
        if (params.length < 2) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "trima({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Triangular moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        var length_half = params.length >> 1;
        const l = 1 + length_half;
        const is_odd = params.length % 2 == 1;

        var factor: f64 = undefined;
        if (is_odd) {
            factor = 1.0 / @as(f64, @floatFromInt(l * l));
        } else {
            factor = 1.0 / @as(f64, @floatFromInt(length_half * l));
            length_half -= 1;
        }

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
            .factor = factor,
            .numerator = 0.0,
            .numerator_sub = 0.0,
            .numerator_add = 0.0,
            .window = window,
            .window_length = params.length,
            .window_length_half = length_half,
            .window_count = 0,
            .is_odd = is_odd,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *TriangularMovingAverage) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *TriangularMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *TriangularMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        var temp = sample;

        if (self.primed) {
            self.numerator -= self.numerator_sub;
            self.numerator_sub -= self.window[0];

            const j = self.window_length - 1;
            var i: usize = 0;
            while (i < j) : (i += 1) {
                self.window[i] = self.window[i + 1];
            }

            self.window[j] = temp;
            temp = self.window[self.window_length_half];
            self.numerator_sub += temp;

            if (self.is_odd) {
                self.numerator += self.numerator_add;
                self.numerator_add -= temp;
            } else {
                self.numerator_add -= temp;
                self.numerator += self.numerator_add;
            }

            temp = sample;
            self.numerator_add += temp;
            self.numerator += temp;
        } else {
            self.window[self.window_count] = temp;
            self.window_count += 1;

            if (self.window_length > self.window_count) {
                return math.nan(f64);
            }

            // Priming: compute initial numerator_sub and numerator
            {
                var ii: usize = self.window_length_half + 1;
                while (ii > 0) {
                    ii -= 1;
                    self.numerator_sub += self.window[ii];
                    self.numerator += self.numerator_sub;
                }
            }

            {
                var ii: usize = self.window_length_half + 1;
                while (ii < self.window_length) : (ii += 1) {
                    self.numerator_add += self.window[ii];
                    self.numerator += self.numerator_add;
                }
            }

            self.primed = true;
        }

        return self.numerator * self.factor;
    }

    pub fn isPrimed(self: *const TriangularMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const TriangularMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .triangular_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *TriangularMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *TriangularMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *TriangularMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *TriangularMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *TriangularMovingAverage) indicator_mod.Indicator {
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
        const self: *TriangularMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const TriangularMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *TriangularMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *TriangularMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *TriangularMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *TriangularMovingAverage = @ptrCast(@alignCast(ptr));
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


fn createTrima(allocator: std.mem.Allocator, length: usize) !TriangularMovingAverage {
    var trima = try TriangularMovingAverage.init(allocator, .{ .length = length });
    trima.fixSlices();
    return trima;
}

fn runAll(trima: *TriangularMovingAverage, input: *const [252]f64) [252]f64 {
    var results: [252]f64 = undefined;
    for (0..252) |i| {
        results[i] = trima.update(input[i]);
    }
    return results;
}

test "trima update length 9" {
    const input = testdata.testInput();
    var trima = try createTrima(testing.allocator, 9);
    defer trima.deinit();
    const results = runAll(&trima, &input);

    for (0..8) |i| {
        try testing.expect(math.isNan(results[i]));
    }
    try testing.expect(@abs(93.8176 - results[8]) < 1e-4);
    try testing.expect(@abs(109.1312 - results[251]) < 1e-4);
    try testing.expect(math.isNan(trima.update(math.nan(f64))));
}

test "trima update length 10" {
    const input = testdata.testInput();
    var trima = try createTrima(testing.allocator, 10);
    defer trima.deinit();
    const results = runAll(&trima, &input);

    for (0..9) |i| {
        try testing.expect(math.isNan(results[i]));
    }
    try testing.expect(@abs(93.6043 - results[9]) < 1e-4);
    try testing.expect(@abs(93.4252 - results[10]) < 1e-4);
    try testing.expect(@abs(109.1850 - results[250]) < 1e-4);
    try testing.expect(@abs(109.1407 - results[251]) < 1e-4);
    try testing.expect(math.isNan(trima.update(math.nan(f64))));
}

test "trima update length 12" {
    const input = testdata.testInput();
    var trima = try createTrima(testing.allocator, 12);
    defer trima.deinit();
    const results = runAll(&trima, &input);

    for (0..10) |i| {
        try testing.expect(math.isNan(results[i]));
    }
    try testing.expect(@abs(93.5329 - results[11]) < 1e-4);
    try testing.expect(@abs(109.1157 - results[251]) < 1e-4);
    try testing.expect(math.isNan(trima.update(math.nan(f64))));
}

test "trima is primed length 9" {
    const input = testdata.testInput();
    var trima = try createTrima(testing.allocator, 9);
    defer trima.deinit();

    try testing.expect(!trima.isPrimed());
    for (0..8) |i| {
        _ = trima.update(input[i]);
        try testing.expect(!trima.isPrimed());
    }
    _ = trima.update(input[8]);
    try testing.expect(trima.isPrimed());
}

test "trima is primed length 12" {
    const input = testdata.testInput();
    var trima = try createTrima(testing.allocator, 12);
    defer trima.deinit();

    try testing.expect(!trima.isPrimed());
    for (0..11) |i| {
        _ = trima.update(input[i]);
        try testing.expect(!trima.isPrimed());
    }
    _ = trima.update(input[11]);
    try testing.expect(trima.isPrimed());
}

test "trima metadata" {
    var trima = try createTrima(testing.allocator, 5);
    defer trima.deinit();
    var m: Metadata = undefined;
    trima.getMetadata(&m);

    try testing.expectEqual(Identifier.triangular_moving_average, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("trima(5)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Triangular moving average trima(5)", m.outputs_buf[0].description);
}

test "trima init invalid length" {
    const result = TriangularMovingAverage.init(testing.allocator, .{ .length = 1 });
    try testing.expectError(error.InvalidLength, result);

    const result0 = TriangularMovingAverage.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, result0);
}

test "trima update entity" {
    const input = testdata.testInput();
    const time: i64 = 1617235200;
    const exp: f64 = 93.5329761904762;

    // scalar
    {
        var trima = try createTrima(testing.allocator, 12);
        defer trima.deinit();
        for (0..11) |i| {
            _ = trima.update(input[i]);
        }
        const out = trima.updateScalar(&.{ .time = time, .value = input[11] });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(@abs(exp - s.value) < 1e-12);
    }

    // bar
    {
        var trima = try createTrima(testing.allocator, 12);
        defer trima.deinit();
        for (0..11) |i| {
            _ = trima.update(input[i]);
        }
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = input[11], .volume = 0 };
        const out = trima.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(exp - s.value) < 1e-12);
    }
}

test "trima mnemonic components" {
    {
        var trima = try createTrima(testing.allocator, 5);
        defer trima.deinit();
        try testing.expectEqualStrings("trima(5)", trima.line.mnemonic);
    }
    {
        var trima = try TriangularMovingAverage.init(testing.allocator, .{
            .length = 5,
            .bar_component = .median,
        });
        defer trima.deinit();
        trima.fixSlices();
        try testing.expectEqualStrings("trima(5, hl/2)", trima.line.mnemonic);
    }
}
