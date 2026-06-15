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

/// Enumerates the outputs of the Polynomial Forecast indicator.
pub const PolynomialForecastOutput = enum(u8) {
    /// The 1-bar-ahead price forecast value.
    value = 1,
};

/// Parameters to create a Polynomial Forecast indicator.
pub const PolynomialForecastParams = struct {
    degree: usize = 3,
    order: usize = 1,
    smoothing: usize = 0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes FIR coefficients for the order-th derivative of a degree-`degree`
/// polynomial fit evaluated at the most recent point (Lagrange basis).
fn computeCoefficients(allocator: std.mem.Allocator, degree: usize, order: usize) ![]f64 {
    const n_points = degree + 1;

    const coefficients = try allocator.alloc(f64, n_points);
    errdefer allocator.free(coefficients);

    const others = try allocator.alloc(f64, degree);
    defer allocator.free(others);

    var i: usize = 0;
    while (i < n_points) : (i += 1) {
        var denom: f64 = 1.0;
        var j: usize = 0;
        while (j < n_points) : (j += 1) {
            if (j != i) {
                denom *= @as(f64, @floatFromInt(@as(i64, @intCast(j)) - @as(i64, @intCast(i))));
            }
        }

        // Fill the "others" list: {0..degree} \ {i}.
        var cnt: usize = 0;
        j = 0;
        while (j < n_points) : (j += 1) {
            if (j != i) {
                others[cnt] = @floatFromInt(j);
                cnt += 1;
            }
        }

        var numerator: f64 = 0.0;
        if (order == 1) {
            var ell: usize = 0;
            while (ell < degree) : (ell += 1) {
                var term: f64 = 1.0;
                var m: usize = 0;
                while (m < degree) : (m += 1) {
                    if (m != ell) {
                        term *= others[m];
                    }
                }
                numerator += term;
            }
        } else {
            var ell: usize = 0;
            while (ell < degree) : (ell += 1) {
                var r: usize = ell + 1;
                while (r < degree) : (r += 1) {
                    var term: f64 = 2.0;
                    var m: usize = 0;
                    while (m < degree) : (m += 1) {
                        if (m != ell and m != r) {
                            term *= others[m];
                        }
                    }
                    numerator += term;
                }
            }
        }

        coefficients[i] = numerator / denom;
    }

    return coefficients;
}

/// Polynomial Forecast (POF) by Don Mak.
///
/// A one-step-ahead price forecast using a Taylor series expansion built on
/// polynomial fit derivatives (PFD):
///   velocity     = PFD(price, degree, order=1)
///   acceleration = PFD(price, degree, order=2)
///   order=1:  forecast = price + velocity
///   order=2:  forecast = price + velocity + 0.5*acceleration
pub const PolynomialForecast = struct {
    line: LineIndicator,

    degree: usize,
    order: usize,
    smoothing: usize,
    n_points: usize,
    coeff_vel: []f64,
    coeff_acc: ?[]f64,

    ema_alpha: f64,
    ema_value: f64 = 0.0,
    ema_initialized: bool = false,

    buf: []f64,
    buf_pos: usize = 0,
    buf_count: usize = 0,

    primed: bool = false,

    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: PolynomialForecastParams) !PolynomialForecast {
        const degree = params.degree;
        const order = params.order;
        const smoothing = params.smoothing;

        if (degree < 2) return error.InvalidDegree;
        if (order < 1 or order > 2) return error.InvalidOrder;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "pof({d},{d},{d}{s})", .{ degree, order, smoothing, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Polynomial forecast {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const coeff_vel = try computeCoefficients(allocator, degree, 1);
        errdefer allocator.free(coeff_vel);

        var coeff_acc: ?[]f64 = null;
        if (order == 2) {
            coeff_acc = try computeCoefficients(allocator, degree, 2);
        }
        errdefer if (coeff_acc) |ca| allocator.free(ca);

        const n_points = degree + 1;
        const buf = try allocator.alloc(f64, n_points);
        @memset(buf, 0.0);

        const ema_alpha: f64 = if (smoothing > 0) 2.0 / (@as(f64, @floatFromInt(smoothing)) + 1.0) else 0.0;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .degree = degree,
            .order = order,
            .smoothing = smoothing,
            .n_points = n_points,
            .coeff_vel = coeff_vel,
            .coeff_acc = coeff_acc,
            .ema_alpha = ema_alpha,
            .buf = buf,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *PolynomialForecast) void {
        self.allocator.free(self.coeff_vel);
        if (self.coeff_acc) |ca| self.allocator.free(ca);
        self.allocator.free(self.buf);
    }

    pub fn fixSlices(self: *PolynomialForecast) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *PolynomialForecast, sample: f64) f64 {
        // Optional EMA pre-smoothing.
        var smoothed: f64 = sample;
        if (self.smoothing > 0) {
            if (!self.ema_initialized) {
                self.ema_value = sample;
                self.ema_initialized = true;
            } else {
                self.ema_value = self.ema_alpha * sample + (1.0 - self.ema_alpha) * self.ema_value;
            }
            smoothed = self.ema_value;
        }

        // Store the smoothed price in the ring buffer.
        self.buf[self.buf_pos] = smoothed;
        self.buf_pos = (self.buf_pos + 1) % self.n_points;
        self.buf_count += 1;

        if (self.buf_count < self.n_points) {
            self.primed = false;
            return math.nan(f64);
        }

        self.primed = true;

        // Read buffer most-recent-first and compute velocity (and acceleration).
        var velocity: f64 = 0.0;
        var acceleration: f64 = 0.0;
        const n: i64 = @intCast(self.n_points);
        var k: usize = 0;
        while (k < self.n_points) : (k += 1) {
            const idx_signed = @as(i64, @intCast(self.buf_pos)) - 1 - @as(i64, @intCast(k));
            const idx: usize = @intCast(@mod(idx_signed, n));
            const value = self.buf[idx];
            velocity += self.coeff_vel[k] * value;
            if (self.coeff_acc) |ca| {
                acceleration += ca[k] * value;
            }
        }

        var forecast = smoothed + velocity;
        if (self.order == 2) {
            forecast += 0.5 * acceleration;
        }

        return forecast;
    }

    pub fn isPrimed(self: *const PolynomialForecast) bool {
        return self.primed;
    }

    fn mnemonic(self: *const PolynomialForecast) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const PolynomialForecast) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const PolynomialForecast, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        build_metadata_mod.buildMetadata(
            out,
            .polynomial_forecast,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    pub fn updateScalar(self: *PolynomialForecast, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *PolynomialForecast, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *PolynomialForecast, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *PolynomialForecast, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *PolynomialForecast) indicator_mod.Indicator {
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
        const self: *PolynomialForecast = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const PolynomialForecast = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *PolynomialForecast = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *PolynomialForecast = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *PolynomialForecast = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *PolynomialForecast = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidDegree,
        InvalidOrder,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

const tolerance = 1e-9;

const Combo = struct {
    degree: usize,
    order: usize,
    smoothing: usize,
    expected: []const f64,
};

fn checkSeries(params: PolynomialForecastParams, inputs: []const f64, expected: []const f64) !void {
    const allocator = testing.allocator;

    var pof = try PolynomialForecast.init(allocator, params);
    defer pof.deinit();

    try testing.expectEqual(inputs.len, expected.len);

    for (0..inputs.len) |i| {
        const value = pof.update(inputs[i]);
        const exp = expected[i];
        if (math.isNan(exp)) {
            try testing.expect(math.isNan(value));
        } else {
            try testing.expect(@abs(value - exp) <= tolerance);
        }
    }
}

test "POF reference data all combos" {
    const input = &testdata.input_close;

    const combos = [_]Combo{
        .{ .degree = 2, .order = 1, .smoothing = 0, .expected = &testdata.expected_d2_o1_s0 },
        .{ .degree = 2, .order = 1, .smoothing = 3, .expected = &testdata.expected_d2_o1_s3 },
        .{ .degree = 2, .order = 1, .smoothing = 6, .expected = &testdata.expected_d2_o1_s6 },
        .{ .degree = 2, .order = 2, .smoothing = 0, .expected = &testdata.expected_d2_o2_s0 },
        .{ .degree = 2, .order = 2, .smoothing = 3, .expected = &testdata.expected_d2_o2_s3 },
        .{ .degree = 2, .order = 2, .smoothing = 6, .expected = &testdata.expected_d2_o2_s6 },
        .{ .degree = 3, .order = 1, .smoothing = 0, .expected = &testdata.expected_d3_o1_s0 },
        .{ .degree = 3, .order = 1, .smoothing = 3, .expected = &testdata.expected_d3_o1_s3 },
        .{ .degree = 3, .order = 1, .smoothing = 6, .expected = &testdata.expected_d3_o1_s6 },
        .{ .degree = 3, .order = 2, .smoothing = 0, .expected = &testdata.expected_d3_o2_s0 },
        .{ .degree = 3, .order = 2, .smoothing = 3, .expected = &testdata.expected_d3_o2_s3 },
        .{ .degree = 3, .order = 2, .smoothing = 6, .expected = &testdata.expected_d3_o2_s6 },
        .{ .degree = 4, .order = 1, .smoothing = 0, .expected = &testdata.expected_d4_o1_s0 },
        .{ .degree = 4, .order = 1, .smoothing = 3, .expected = &testdata.expected_d4_o1_s3 },
        .{ .degree = 4, .order = 1, .smoothing = 6, .expected = &testdata.expected_d4_o1_s6 },
        .{ .degree = 4, .order = 2, .smoothing = 0, .expected = &testdata.expected_d4_o2_s0 },
        .{ .degree = 4, .order = 2, .smoothing = 3, .expected = &testdata.expected_d4_o2_s3 },
        .{ .degree = 4, .order = 2, .smoothing = 6, .expected = &testdata.expected_d4_o2_s6 },
        .{ .degree = 5, .order = 1, .smoothing = 0, .expected = &testdata.expected_d5_o1_s0 },
        .{ .degree = 5, .order = 1, .smoothing = 3, .expected = &testdata.expected_d5_o1_s3 },
        .{ .degree = 5, .order = 1, .smoothing = 6, .expected = &testdata.expected_d5_o1_s6 },
        .{ .degree = 5, .order = 2, .smoothing = 0, .expected = &testdata.expected_d5_o2_s0 },
        .{ .degree = 5, .order = 2, .smoothing = 3, .expected = &testdata.expected_d5_o2_s3 },
        .{ .degree = 5, .order = 2, .smoothing = 6, .expected = &testdata.expected_d5_o2_s6 },
    };

    for (combos) |combo| {
        try checkSeries(.{ .degree = combo.degree, .order = combo.order, .smoothing = combo.smoothing }, input, combo.expected);
    }

    try checkSeries(.{ .degree = 3, .order = 1, .smoothing = 0 }, &testdata.test1_input_linear, &testdata.test1_expected_d3_o1_s0);
    try checkSeries(.{ .degree = 3, .order = 2, .smoothing = 0 }, &testdata.test1_input_linear, &testdata.test1_expected_d3_o2_s0);
}

test "POF metadata default" {
    const allocator = testing.allocator;

    var pof = try PolynomialForecast.init(allocator, .{});
    defer pof.deinit();
    pof.fixSlices();

    var meta: Metadata = undefined;
    pof.getMetadata(&meta);

    try testing.expectEqual(Identifier.polynomial_forecast, meta.identifier);
    try testing.expectEqualStrings("pof(3,1,0)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "POF custom mnemonic" {
    const allocator = testing.allocator;

    var pof = try PolynomialForecast.init(allocator, .{ .degree = 5, .order = 2, .smoothing = 6 });
    defer pof.deinit();

    var meta: Metadata = undefined;
    pof.getMetadata(&meta);

    try testing.expectEqualStrings("pof(5,2,6)", meta.mnemonic);
}

test "POF invalid params" {
    const allocator = testing.allocator;

    const r1 = PolynomialForecast.init(allocator, .{ .degree = 1 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = PolynomialForecast.init(allocator, .{ .order = 0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = PolynomialForecast.init(allocator, .{ .order = 3 });
    try testing.expect(if (r3) |_| false else |_| true);
}
