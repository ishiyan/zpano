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

/// Enumerates the outputs of the Polynomial Fit Derivative indicator.
pub const PolynomialFitDerivativeOutput = enum(u8) {
    /// The order-th derivative of the polynomial fit at the current bar.
    value = 1,
};

/// Parameters to create a Polynomial Fit Derivative indicator.
pub const PolynomialFitDerivativeParams = struct {
    degree: usize = 3,
    order: usize = 1,
    smoothing: i64 = 6,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the FIR filter coefficients for the order-th derivative of a
/// degree-`degree` polynomial fit, evaluated at the most recent point.
///
/// Uses the Lagrange basis with the elementary-symmetric-polynomial identity:
///   c_i = order! * e_{degree-order}(others) / prod_{j != i} (j - i)
/// where `others` is the set of point positions {0..degree} excluding i.
fn computeCoefficients(allocator: std.mem.Allocator, degree: usize, order: usize) ![]f64 {
    const n_points = degree + 1;

    var factorial_order: f64 = 1.0;
    var f: usize = 2;
    while (f <= order) : (f += 1) {
        factorial_order *= @floatFromInt(f);
    }

    const coefficients = try allocator.alloc(f64, n_points);
    errdefer allocator.free(coefficients);

    // Scratch array for elementary symmetric polynomials e[0..degree].
    const e = try allocator.alloc(f64, degree + 1);
    defer allocator.free(e);

    var i: usize = 0;
    while (i < n_points) : (i += 1) {
        var denom: f64 = 1.0;
        var j: usize = 0;
        while (j < n_points) : (j += 1) {
            if (j != i) {
                denom *= @as(f64, @floatFromInt(@as(i64, @intCast(j)) - @as(i64, @intCast(i))));
            }
        }

        // Elementary symmetric polynomials of the values {0..degree} \ {i}.
        @memset(e, 0.0);
        e[0] = 1.0;
        j = 0;
        while (j < n_points) : (j += 1) {
            if (j == i) continue;
            const v: f64 = @floatFromInt(j);
            var k: usize = degree;
            while (k >= 1) : (k -= 1) {
                e[k] += v * e[k - 1];
            }
        }

        const numerator = factorial_order * e[degree - order];
        coefficients[i] = numerator / denom;
    }

    return coefficients;
}

/// Polynomial Fit Derivative (PFD) by Don Mak.
///
/// Fits a polynomial of degree `degree` to the most recent `degree + 1`
/// (optionally EMA-smoothed) prices and evaluates its `order`-th derivative at
/// the current bar. This is a FIR filter: a dot product of fixed Lagrange-derived
/// coefficients with the last `degree + 1` smoothed prices.
pub const PolynomialFitDerivative = struct {
    line: LineIndicator,

    coefficients: []f64,
    n_points: usize,

    smoothing: i64,
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

    pub fn init(allocator: std.mem.Allocator, params: PolynomialFitDerivativeParams) !PolynomialFitDerivative {
        const degree = params.degree;
        const order = params.order;
        const smoothing = params.smoothing;

        if (degree < 2) return error.InvalidDegree;
        if (order < 1 or order > degree) return error.InvalidOrder;
        if (smoothing < 0) return error.InvalidSmoothing;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "pfd({d},{d},{d}{s})", .{
            degree,
            order,
            smoothing,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Polynomial fit derivative {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const coefficients = try computeCoefficients(allocator, degree, order);
        errdefer allocator.free(coefficients);

        const buf = try allocator.alloc(f64, degree + 1);
        @memset(buf, 0.0);

        const ema_alpha: f64 = if (smoothing > 0)
            2.0 / (@as(f64, @floatFromInt(smoothing)) + 1.0)
        else
            0.0;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .coefficients = coefficients,
            .n_points = degree + 1,
            .smoothing = smoothing,
            .ema_alpha = ema_alpha,
            .buf = buf,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *PolynomialFitDerivative) void {
        self.allocator.free(self.coefficients);
        self.allocator.free(self.buf);
    }

    pub fn fixSlices(self: *PolynomialFitDerivative) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *PolynomialFitDerivative, sample: f64) f64 {
        // Step 1: optional EMA smoothing.
        var smoothed = sample;
        if (self.smoothing > 0) {
            if (!self.ema_initialized) {
                self.ema_value = sample;
                self.ema_initialized = true;
            } else {
                self.ema_value = self.ema_alpha * sample + (1.0 - self.ema_alpha) * self.ema_value;
            }
            smoothed = self.ema_value;
        }

        // Step 2: push into the ring buffer.
        self.buf[self.buf_pos] = smoothed;
        self.buf_pos = (self.buf_pos + 1) % self.n_points;
        self.buf_count += 1;

        // Step 3: not enough data yet.
        if (self.buf_count < self.n_points) {
            self.primed = false;
            return math.nan(f64);
        }

        // Step 4: FIR dot product (coefficients[j] multiplies the j-th most recent).
        var result: f64 = 0.0;
        var j: usize = 0;
        while (j < self.n_points) : (j += 1) {
            const offset = @as(i64, @intCast(self.buf_pos)) - 1 - @as(i64, @intCast(j));
            const n: i64 = @intCast(self.n_points);
            const buf_idx: usize = @intCast(@mod(offset, n));
            result += self.coefficients[j] * self.buf[buf_idx];
        }

        self.primed = true;
        return result;
    }

    pub fn isPrimed(self: *const PolynomialFitDerivative) bool {
        return self.primed;
    }

    fn mnemonic(self: *const PolynomialFitDerivative) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const PolynomialFitDerivative) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const PolynomialFitDerivative, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        build_metadata_mod.buildMetadata(
            out,
            .polynomial_fit_derivative,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    pub fn updateScalar(self: *PolynomialFitDerivative, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *PolynomialFitDerivative, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *PolynomialFitDerivative, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *PolynomialFitDerivative, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *PolynomialFitDerivative) indicator_mod.Indicator {
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
        const self: *PolynomialFitDerivative = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const PolynomialFitDerivative = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *PolynomialFitDerivative = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *PolynomialFitDerivative = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *PolynomialFitDerivative = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *PolynomialFitDerivative = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidDegree,
        InvalidOrder,
        InvalidSmoothing,
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
    smoothing: i64,
    expected: [252]f64,
};

test "PFD reference data all combos" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    const combos = [_]Combo{
        .{ .degree = 2, .order = 1, .smoothing = 0, .expected = testdata.expectedD2_O1_S0() },
        .{ .degree = 2, .order = 1, .smoothing = 3, .expected = testdata.expectedD2_O1_S3() },
        .{ .degree = 2, .order = 1, .smoothing = 6, .expected = testdata.expectedD2_O1_S6() },
        .{ .degree = 2, .order = 2, .smoothing = 0, .expected = testdata.expectedD2_O2_S0() },
        .{ .degree = 2, .order = 2, .smoothing = 3, .expected = testdata.expectedD2_O2_S3() },
        .{ .degree = 2, .order = 2, .smoothing = 6, .expected = testdata.expectedD2_O2_S6() },
        .{ .degree = 3, .order = 1, .smoothing = 0, .expected = testdata.expectedD3_O1_S0() },
        .{ .degree = 3, .order = 1, .smoothing = 3, .expected = testdata.expectedD3_O1_S3() },
        .{ .degree = 3, .order = 1, .smoothing = 6, .expected = testdata.expectedD3_O1_S6() },
        .{ .degree = 3, .order = 2, .smoothing = 0, .expected = testdata.expectedD3_O2_S0() },
        .{ .degree = 3, .order = 2, .smoothing = 3, .expected = testdata.expectedD3_O2_S3() },
        .{ .degree = 3, .order = 2, .smoothing = 6, .expected = testdata.expectedD3_O2_S6() },
        .{ .degree = 4, .order = 1, .smoothing = 0, .expected = testdata.expectedD4_O1_S0() },
        .{ .degree = 4, .order = 1, .smoothing = 3, .expected = testdata.expectedD4_O1_S3() },
        .{ .degree = 4, .order = 1, .smoothing = 6, .expected = testdata.expectedD4_O1_S6() },
        .{ .degree = 4, .order = 2, .smoothing = 0, .expected = testdata.expectedD4_O2_S0() },
        .{ .degree = 4, .order = 2, .smoothing = 3, .expected = testdata.expectedD4_O2_S3() },
        .{ .degree = 4, .order = 2, .smoothing = 6, .expected = testdata.expectedD4_O2_S6() },
        .{ .degree = 5, .order = 1, .smoothing = 0, .expected = testdata.expectedD5_O1_S0() },
        .{ .degree = 5, .order = 1, .smoothing = 3, .expected = testdata.expectedD5_O1_S3() },
        .{ .degree = 5, .order = 1, .smoothing = 6, .expected = testdata.expectedD5_O1_S6() },
        .{ .degree = 5, .order = 2, .smoothing = 0, .expected = testdata.expectedD5_O2_S0() },
        .{ .degree = 5, .order = 2, .smoothing = 3, .expected = testdata.expectedD5_O2_S3() },
        .{ .degree = 5, .order = 2, .smoothing = 6, .expected = testdata.expectedD5_O2_S6() },
        .{ .degree = 6, .order = 1, .smoothing = 0, .expected = testdata.expectedD6_O1_S0() },
        .{ .degree = 6, .order = 1, .smoothing = 3, .expected = testdata.expectedD6_O1_S3() },
        .{ .degree = 6, .order = 1, .smoothing = 6, .expected = testdata.expectedD6_O1_S6() },
        .{ .degree = 6, .order = 2, .smoothing = 0, .expected = testdata.expectedD6_O2_S0() },
        .{ .degree = 6, .order = 2, .smoothing = 3, .expected = testdata.expectedD6_O2_S3() },
        .{ .degree = 6, .order = 2, .smoothing = 6, .expected = testdata.expectedD6_O2_S6() },
        .{ .degree = 4, .order = 3, .smoothing = 6, .expected = testdata.expectedD4_O3_S6() },
        .{ .degree = 5, .order = 3, .smoothing = 6, .expected = testdata.expectedD5_O3_S6() },
        .{ .degree = 6, .order = 3, .smoothing = 6, .expected = testdata.expectedD6_O3_S6() },
        .{ .degree = 6, .order = 5, .smoothing = 6, .expected = testdata.expectedD6_O5_S6() },
    };

    for (combos) |combo| {
        var pfd = try PolynomialFitDerivative.init(allocator, .{
            .degree = combo.degree,
            .order = combo.order,
            .smoothing = combo.smoothing,
        });
        defer pfd.deinit();

        for (0..252) |idx| {
            const value = pfd.update(input[idx]);
            const exp = combo.expected[idx];
            if (math.isNan(exp)) {
                try testing.expect(math.isNan(value));
            } else {
                try testing.expect(@abs(value - exp) <= tolerance);
            }
        }
    }
}

test "PFD metadata default" {
    const allocator = testing.allocator;

    var pfd = try PolynomialFitDerivative.init(allocator, .{});
    defer pfd.deinit();
    pfd.fixSlices();

    var meta: Metadata = undefined;
    pfd.getMetadata(&meta);

    try testing.expectEqual(Identifier.polynomial_fit_derivative, meta.identifier);
    try testing.expectEqualStrings("pfd(3,1,6)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "PFD custom mnemonic" {
    const allocator = testing.allocator;

    var pfd = try PolynomialFitDerivative.init(allocator, .{ .degree = 4, .order = 2, .smoothing = 3 });
    defer pfd.deinit();

    var meta: Metadata = undefined;
    pfd.getMetadata(&meta);

    try testing.expectEqualStrings("pfd(4,2,3)", meta.mnemonic);
}

test "PFD invalid params" {
    const allocator = testing.allocator;

    const r1 = PolynomialFitDerivative.init(allocator, .{ .degree = 1 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = PolynomialFitDerivative.init(allocator, .{ .order = 0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = PolynomialFitDerivative.init(allocator, .{ .degree = 3, .order = 4 });
    try testing.expect(if (r3) |_| false else |_| true);

    const r4 = PolynomialFitDerivative.init(allocator, .{ .smoothing = -1 });
    try testing.expect(if (r4) |_| false else |_| true);
}
