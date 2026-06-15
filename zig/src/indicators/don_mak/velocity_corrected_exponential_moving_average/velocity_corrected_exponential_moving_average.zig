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

/// Enumerates the outputs of the Velocity-Corrected Exponential Moving Average indicator.
pub const VelocityCorrectedExponentialMovingAverageOutput = enum(u8) {
    /// The velocity-corrected EMA value.
    value = 1,
};

/// Parameters to create a Velocity-Corrected Exponential Moving Average indicator.
pub const VelocityCorrectedExponentialMovingAverageParams = struct {
    period: usize = 6,
    degree: usize = 3,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes FIR coefficients for the first derivative of a degree-`degree`
/// polynomial fit evaluated at the most recent point (Lagrange basis, order=1).
fn computeVelocityCoefficients(allocator: std.mem.Allocator, degree: usize) ![]f64 {
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

        // First derivative of the numerator at t=0 (sum over each removed "other").
        var numerator: f64 = 0.0;
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

        coefficients[i] = numerator / denom;
    }

    return coefficients;
}

/// Velocity-Corrected Exponential Moving Average (VCEMA) by Don Mak.
///
/// A reduced-lag EMA that pre-corrects price by adding its polynomial velocity
/// before smoothing:
///   corrected = price + PFD(price, degree, order=1)
///   VCEMA(n)  = EMA(corrected, n)
pub const VelocityCorrectedExponentialMovingAverage = struct {
    line: LineIndicator,

    degree: usize,
    n_points: usize,
    coefficients: []f64,

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

    pub fn init(allocator: std.mem.Allocator, params: VelocityCorrectedExponentialMovingAverageParams) !VelocityCorrectedExponentialMovingAverage {
        const period = params.period;
        const degree = params.degree;

        if (period < 2) return error.InvalidPeriod;
        if (degree < 2) return error.InvalidDegree;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "vcema({d},{d}{s})", .{ period, degree, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Velocity-corrected exponential moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const coefficients = try computeVelocityCoefficients(allocator, degree);
        errdefer allocator.free(coefficients);

        const n_points = degree + 1;
        const buf = try allocator.alloc(f64, n_points);
        @memset(buf, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .degree = degree,
            .n_points = n_points,
            .coefficients = coefficients,
            .ema_alpha = 2.0 / (@as(f64, @floatFromInt(period)) + 1.0),
            .buf = buf,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *VelocityCorrectedExponentialMovingAverage) void {
        self.allocator.free(self.coefficients);
        self.allocator.free(self.buf);
    }

    pub fn fixSlices(self: *VelocityCorrectedExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *VelocityCorrectedExponentialMovingAverage, sample: f64) f64 {
        // Store the raw price in the ring buffer.
        self.buf[self.buf_pos] = sample;
        self.buf_pos = (self.buf_pos + 1) % self.n_points;
        self.buf_count += 1;

        if (self.buf_count < self.n_points) {
            self.primed = false;
            return math.nan(f64);
        }

        self.primed = true;

        // Compute the velocity from the raw prices.
        var velocity: f64 = 0.0;
        const n: i64 = @intCast(self.n_points);
        var k: usize = 0;
        while (k < self.n_points) : (k += 1) {
            const idx_signed = @as(i64, @intCast(self.buf_pos)) - 1 - @as(i64, @intCast(k));
            const idx: usize = @intCast(@mod(idx_signed, n));
            velocity += self.coefficients[k] * self.buf[idx];
        }

        // Corrected price = price + velocity.
        const corrected = sample + velocity;

        // Apply the EMA to the corrected price (seed at the first corrected value).
        if (!self.ema_initialized) {
            self.ema_value = corrected;
            self.ema_initialized = true;
        } else {
            self.ema_value = self.ema_alpha * corrected + (1.0 - self.ema_alpha) * self.ema_value;
        }

        return self.ema_value;
    }

    pub fn isPrimed(self: *const VelocityCorrectedExponentialMovingAverage) bool {
        return self.primed;
    }

    fn mnemonic(self: *const VelocityCorrectedExponentialMovingAverage) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const VelocityCorrectedExponentialMovingAverage) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const VelocityCorrectedExponentialMovingAverage, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        build_metadata_mod.buildMetadata(
            out,
            .velocity_corrected_exponential_moving_average,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    pub fn updateScalar(self: *VelocityCorrectedExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *VelocityCorrectedExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *VelocityCorrectedExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *VelocityCorrectedExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *VelocityCorrectedExponentialMovingAverage) indicator_mod.Indicator {
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
        const self: *VelocityCorrectedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const VelocityCorrectedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *VelocityCorrectedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *VelocityCorrectedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *VelocityCorrectedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *VelocityCorrectedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidPeriod,
        InvalidDegree,
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
    period: usize,
    degree: usize,
    expected: []const f64,
};

fn checkSeries(params: VelocityCorrectedExponentialMovingAverageParams, inputs: []const f64, expected: []const f64) !void {
    const allocator = testing.allocator;

    var vcema = try VelocityCorrectedExponentialMovingAverage.init(allocator, params);
    defer vcema.deinit();

    try testing.expectEqual(inputs.len, expected.len);

    for (0..inputs.len) |i| {
        const value = vcema.update(inputs[i]);
        const exp = expected[i];
        if (math.isNan(exp)) {
            try testing.expect(math.isNan(value));
        } else {
            try testing.expect(@abs(value - exp) <= tolerance);
        }
    }
}

test "VCEMA reference data all combos" {
    const input = &testdata.input_close;

    const combos = [_]Combo{
        .{ .period = 3, .degree = 2, .expected = &testdata.expected_p3_d2 },
        .{ .period = 3, .degree = 3, .expected = &testdata.expected_p3_d3 },
        .{ .period = 3, .degree = 4, .expected = &testdata.expected_p3_d4 },
        .{ .period = 3, .degree = 5, .expected = &testdata.expected_p3_d5 },
        .{ .period = 6, .degree = 2, .expected = &testdata.expected_p6_d2 },
        .{ .period = 6, .degree = 3, .expected = &testdata.expected_p6_d3 },
        .{ .period = 6, .degree = 4, .expected = &testdata.expected_p6_d4 },
        .{ .period = 6, .degree = 5, .expected = &testdata.expected_p6_d5 },
        .{ .period = 12, .degree = 2, .expected = &testdata.expected_p12_d2 },
        .{ .period = 12, .degree = 3, .expected = &testdata.expected_p12_d3 },
        .{ .period = 12, .degree = 4, .expected = &testdata.expected_p12_d4 },
        .{ .period = 12, .degree = 5, .expected = &testdata.expected_p12_d5 },
    };

    for (combos) |combo| {
        try checkSeries(.{ .period = combo.period, .degree = combo.degree }, input, combo.expected);
    }

    try checkSeries(.{ .period = 6, .degree = 3 }, &testdata.test1_input_linear, &testdata.test1_expected_p6_d3);
}

test "VCEMA metadata default" {
    const allocator = testing.allocator;

    var vcema = try VelocityCorrectedExponentialMovingAverage.init(allocator, .{});
    defer vcema.deinit();
    vcema.fixSlices();

    var meta: Metadata = undefined;
    vcema.getMetadata(&meta);

    try testing.expectEqual(Identifier.velocity_corrected_exponential_moving_average, meta.identifier);
    try testing.expectEqualStrings("vcema(6,3)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "VCEMA custom mnemonic" {
    const allocator = testing.allocator;

    var vcema = try VelocityCorrectedExponentialMovingAverage.init(allocator, .{ .period = 12, .degree = 5 });
    defer vcema.deinit();

    var meta: Metadata = undefined;
    vcema.getMetadata(&meta);

    try testing.expectEqualStrings("vcema(12,5)", meta.mnemonic);
}

test "VCEMA invalid params" {
    const allocator = testing.allocator;

    const r1 = VelocityCorrectedExponentialMovingAverage.init(allocator, .{ .period = 1 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = VelocityCorrectedExponentialMovingAverage.init(allocator, .{ .degree = 1 });
    try testing.expect(if (r2) |_| false else |_| true);
}
