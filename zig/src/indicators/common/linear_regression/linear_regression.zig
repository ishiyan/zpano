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

/// Enumerates the outputs of the linear regression indicator.
pub const LinearRegressionOutput = enum(u8) {
    /// The regression value: b + m*(period-1).
    value = 1,
    /// The time series forecast: b + m*period.
    forecast = 2,
    /// The y-intercept of the regression line: b.
    intercept = 3,
    /// The slope of the regression line: m.
    slope_rad = 4,
    /// The slope in degrees: atan(m) * 180/pi.
    slope_deg = 5,
};

/// Parameters to create an instance of the linear regression indicator.
pub const LinearRegressionParams = struct {
    /// The lookback period. Must be >= 2.
    length: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the least-squares regression line over a rolling window and produces
/// five outputs per sample: Value, Forecast, Intercept, SlopeRad, SlopeDeg.
pub const LinearRegression = struct {
    line: LineIndicator,
    window: []f64,
    length: usize,
    length_f: f64,
    sum_x: f64,
    divisor: f64,
    window_count: usize,
    primed: bool,
    cur_value: f64,
    cur_forecast: f64,
    cur_intercept: f64,
    cur_slope_rad: f64,
    cur_slope_deg: f64,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    const rad_to_deg = 180.0 / math.pi;

    pub fn init(allocator: std.mem.Allocator, params: LinearRegressionParams) !LinearRegression {
        if (params.length < 2) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "linreg({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Linear Regression {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const n: f64 = @floatFromInt(params.length);
        const sum_x = n * (n - 1.0) * 0.5;
        const sum_x_sqr = n * (n - 1.0) * (2.0 * n - 1.0) / 6.0;
        const divisor = sum_x * sum_x - n * sum_x_sqr;

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
            .length = params.length,
            .length_f = n,
            .sum_x = sum_x,
            .divisor = divisor,
            .window_count = 0,
            .primed = false,
            .cur_value = 0.0,
            .cur_forecast = 0.0,
            .cur_intercept = 0.0,
            .cur_slope_rad = 0.0,
            .cur_slope_deg = 0.0,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *LinearRegression) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *LinearRegression) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    fn computeFromWindow(self: *LinearRegression) void {
        var sum_xy: f64 = 0.0;
        var sum_y: f64 = 0.0;

        var i: usize = self.length;
        while (i > 0) : (i -= 1) {
            const v = self.window[self.length - i];
            sum_y += v;
            sum_xy += @as(f64, @floatFromInt(i - 1)) * v;
        }

        const m = (self.length_f * sum_xy - self.sum_x * sum_y) / self.divisor;
        const b = (sum_y - m * self.sum_x) / self.length_f;

        self.cur_slope_rad = m;
        self.cur_slope_deg = math.atan(m) * rad_to_deg;
        self.cur_intercept = b;
        self.cur_value = b + m * (self.length_f - 1.0);
        self.cur_forecast = b + m * self.length_f;
    }

    fn calculate(self: *LinearRegression, sample: f64) void {
        // Shift window.
        var i: usize = 0;
        while (i < self.length - 1) : (i += 1) {
            self.window[i] = self.window[i + 1];
        }
        self.window[self.length - 1] = sample;
        self.computeFromWindow();
    }

    /// Core update. Returns the Value output or NaN if not primed.
    pub fn update(self: *LinearRegression, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            self.calculate(sample);
            return self.cur_value;
        }

        self.window[self.window_count] = sample;
        self.window_count += 1;

        if (self.window_count == self.length) {
            self.primed = true;
            self.computeFromWindow();
            return self.cur_value;
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const LinearRegression) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const LinearRegression, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .linear_regression,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    fn makeMultiOutput(self: *LinearRegression, time: i64, value: f64) OutputArray {
        var out = OutputArray{};
        if (math.isNan(value)) {
            const nan = math.nan(f64);
            out.append(.{ .scalar = .{ .time = time, .value = nan } });
            out.append(.{ .scalar = .{ .time = time, .value = nan } });
            out.append(.{ .scalar = .{ .time = time, .value = nan } });
            out.append(.{ .scalar = .{ .time = time, .value = nan } });
            out.append(.{ .scalar = .{ .time = time, .value = nan } });
        } else {
            out.append(.{ .scalar = .{ .time = time, .value = self.cur_value } });
            out.append(.{ .scalar = .{ .time = time, .value = self.cur_forecast } });
            out.append(.{ .scalar = .{ .time = time, .value = self.cur_intercept } });
            out.append(.{ .scalar = .{ .time = time, .value = self.cur_slope_rad } });
            out.append(.{ .scalar = .{ .time = time, .value = self.cur_slope_deg } });
        }
        return out;
    }

    pub fn updateScalar(self: *LinearRegression, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return self.makeMultiOutput(sample.time, value);
    }

    pub fn updateBar(self: *LinearRegression, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return self.makeMultiOutput(sample.time, value);
    }

    pub fn updateQuote(self: *LinearRegression, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return self.makeMultiOutput(sample.time, value);
    }

    pub fn updateTrade(self: *LinearRegression, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return self.makeMultiOutput(sample.time, value);
    }

    pub fn indicator(self: *LinearRegression) indicator_mod.Indicator {
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
        const self: *LinearRegression = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const LinearRegression = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *LinearRegression = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *LinearRegression = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *LinearRegression = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *LinearRegression = @ptrCast(@alignCast(ptr));
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


fn createLinreg(allocator: std.mem.Allocator, length: usize) !LinearRegression {
    var lr = try LinearRegression.init(allocator, .{ .length = length });
    lr.fixSlices();
    return lr;
}

fn almostEqual(a: f64, b: f64, eps: f64) bool {
    return @abs(a - b) < eps;
}

// 252-entry input data from Excel verification (period 14).
test "linear regression value output period 14 all 252 rows" {
    const tolerance = 1e-4;
    const input = testdata.testInput();
    const exp_value = testdata.testExpectedValue();

    var lr = try createLinreg(testing.allocator, 14);
    defer lr.deinit();

    for (0..13) |i| {
        try testing.expect(math.isNan(lr.update(input[i])));
    }
    for (13..252) |i| {
        const value = lr.update(input[i]);
        try testing.expect(almostEqual(value, exp_value[i], tolerance));
    }
    try testing.expect(math.isNan(lr.update(math.nan(f64))));
}

test "linear regression all 5 outputs period 14 all 252 rows" {
    const tolerance = 1e-4;
    const input = testdata.testInput();
    const exp_value = testdata.testExpectedValue();
    const exp_forecast = testdata.testExpectedForecast();
    const exp_intercept = testdata.testExpectedIntercept();
    const exp_slope_rad = testdata.testExpectedSlopeRad();
    const exp_slope_deg = testdata.testExpectedSlopeDeg();

    var lr = try createLinreg(testing.allocator, 14);
    defer lr.deinit();

    // Feed first 12 samples via update.
    for (0..12) |i| {
        _ = lr.update(input[i]);
    }

    // Feed index 12 via updateScalar to get NaN outputs.
    const time: i64 = 1617235200;
    const out_nan = lr.updateScalar(&.{ .time = time, .value = input[12] });
    try testing.expectEqual(@as(usize, 5), out_nan.len);
    for (0..5) |j| {
        try testing.expect(math.isNan(out_nan.slice()[j].scalar.value));
    }

    // Feed indices 13-251 via updateScalar and verify all 5 outputs.
    for (13..252) |i| {
        const out = lr.updateScalar(&.{ .time = time, .value = input[i] });
        try testing.expectEqual(@as(usize, 5), out.len);
        const s = out.slice();
        try testing.expect(almostEqual(s[0].scalar.value, exp_value[i], tolerance));
        try testing.expect(almostEqual(s[1].scalar.value, exp_forecast[i], tolerance));
        try testing.expect(almostEqual(s[2].scalar.value, exp_intercept[i], tolerance));
        try testing.expect(almostEqual(s[3].scalar.value, exp_slope_rad[i], tolerance));
        try testing.expect(almostEqual(s[4].scalar.value, exp_slope_deg[i], tolerance));
    }
}

test "linear regression is primed" {
    const input = testdata.testInput();
    var lr = try createLinreg(testing.allocator, 14);
    defer lr.deinit();

    try testing.expect(!lr.isPrimed());
    for (0..13) |_| {
        _ = lr.update(input[0]);
        try testing.expect(!lr.isPrimed());
    }
    _ = lr.update(input[13]);
    try testing.expect(lr.isPrimed());
}

test "linear regression is primed length 2" {
    const input = testdata.testInput();
    var lr = try createLinreg(testing.allocator, 2);
    defer lr.deinit();

    try testing.expect(!lr.isPrimed());
    _ = lr.update(input[0]);
    try testing.expect(!lr.isPrimed());
    _ = lr.update(input[1]);
    try testing.expect(lr.isPrimed());
}

test "linear regression metadata" {
    var lr = try createLinreg(testing.allocator, 14);
    defer lr.deinit();
    var m: Metadata = undefined;
    lr.getMetadata(&m);

    try testing.expectEqual(Identifier.linear_regression, m.identifier);
    try testing.expectEqualStrings("linreg(14)", m.mnemonic);
    try testing.expectEqualStrings("Linear Regression linreg(14)", m.description);
    try testing.expectEqual(@as(usize, 5), m.outputs_len);
}

test "linear regression init invalid length" {
    const r1 = LinearRegression.init(testing.allocator, .{ .length = 1 });
    try testing.expectError(error.InvalidLength, r1);
    const r0 = LinearRegression.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, r0);
}

test "linear regression mnemonic with components" {
    var lr = try LinearRegression.init(testing.allocator, .{
        .length = 14,
        .bar_component = .median,
    });
    defer lr.deinit();
    lr.fixSlices();
    try testing.expectEqualStrings("linreg(14, hl/2)", lr.line.mnemonic);
}

test "linear regression update entity" {
    const time: i64 = 1617235200;
    const input = testdata.testInput();

    // bar
    {
        var lr = try createLinreg(testing.allocator, 14);
        defer lr.deinit();
        for (0..13) |i| {
            _ = lr.update(input[i]);
        }
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = input[13], .volume = 0 };
        const out = lr.updateBar(&bar);
        try testing.expectEqual(@as(usize, 5), out.len);
        try testing.expectEqual(time, out.slice()[0].scalar.time);
    }

    // quote
    {
        var lr = try createLinreg(testing.allocator, 14);
        defer lr.deinit();
        for (0..13) |i| {
            _ = lr.update(input[i]);
        }
        const quote = Quote{ .time = time, .bid_price = input[13], .ask_price = input[13], .bid_size = 0, .ask_size = 0 };
        const out = lr.updateQuote(&quote);
        try testing.expectEqual(@as(usize, 5), out.len);
    }

    // trade
    {
        var lr = try createLinreg(testing.allocator, 14);
        defer lr.deinit();
        for (0..13) |i| {
            _ = lr.update(input[i]);
        }
        const trade = Trade{ .time = time, .price = input[13], .volume = 0 };
        const out = lr.updateTrade(&trade);
        try testing.expectEqual(@as(usize, 5), out.len);
    }
}
