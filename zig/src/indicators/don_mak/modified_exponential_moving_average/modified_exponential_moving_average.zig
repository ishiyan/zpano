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

/// Enumerates the outputs of the Modified Exponential Moving Average indicator.
pub const ModifiedExponentialMovingAverageOutput = enum(u8) {
    /// The velocity-corrected EMA value.
    value = 1,
};

/// Parameters to create a Modified Exponential Moving Average indicator.
pub const ModifiedExponentialMovingAverageParams = struct {
    period: usize = 6,
    degree: usize = 3,
    skip: usize = 1,
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

/// Modified Exponential Moving Average (MEMA / MEMA-D) by Don Mak.
///
/// A reduced-lag EMA that adds the EMA's own polynomial velocity back to its
/// output, compensating for smoothing delay:
///   MEMA(n) = EMA(n) + PFD(EMA, degree, order=1, stride=skip)
pub const ModifiedExponentialMovingAverage = struct {
    line: LineIndicator,

    degree: usize,
    skip: usize,
    n_points: usize,
    coefficients: []f64,

    ema_alpha: f64,
    ema_value: f64 = 0.0,
    ema_initialized: bool = false,

    buf: []f64,
    buf_size: usize,
    buf_pos: usize = 0,
    buf_count: usize = 0,

    primed: bool = false,

    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: ModifiedExponentialMovingAverageParams) !ModifiedExponentialMovingAverage {
        const period = params.period;
        const degree = params.degree;
        const skip = params.skip;

        if (period < 2) return error.InvalidPeriod;
        if (degree < 2) return error.InvalidDegree;
        if (skip < 1) return error.InvalidSkip;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "mema({d},{d},{d}{s})", .{ period, degree, skip, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Modified exponential moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const coefficients = try computeVelocityCoefficients(allocator, degree);
        errdefer allocator.free(coefficients);

        const buf_size = degree * skip + 1;
        const buf = try allocator.alloc(f64, buf_size);
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
            .skip = skip,
            .n_points = degree + 1,
            .coefficients = coefficients,
            .ema_alpha = 2.0 / (@as(f64, @floatFromInt(period)) + 1.0),
            .buf = buf,
            .buf_size = buf_size,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *ModifiedExponentialMovingAverage) void {
        self.allocator.free(self.coefficients);
        self.allocator.free(self.buf);
    }

    pub fn fixSlices(self: *ModifiedExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *ModifiedExponentialMovingAverage, sample: f64) f64 {
        // EMA recursion (seed at first sample).
        if (!self.ema_initialized) {
            self.ema_value = sample;
            self.ema_initialized = true;
        } else {
            self.ema_value = self.ema_alpha * sample + (1.0 - self.ema_alpha) * self.ema_value;
        }

        // Store EMA value in the ring buffer.
        self.buf[self.buf_pos] = self.ema_value;
        self.buf_pos = (self.buf_pos + 1) % self.buf_size;
        self.buf_count += 1;

        if (self.buf_count < self.buf_size) {
            self.primed = false;
            return math.nan(f64);
        }

        self.primed = true;

        // Read EMA values at stride positions and compute the velocity correction.
        var velocity: f64 = 0.0;
        const n: i64 = @intCast(self.buf_size);
        var k: usize = 0;
        while (k < self.n_points) : (k += 1) {
            const offset: i64 = @intCast(k * self.skip);
            const idx_signed = @as(i64, @intCast(self.buf_pos)) - 1 - offset;
            const idx: usize = @intCast(@mod(idx_signed, n));
            velocity += self.coefficients[k] * self.buf[idx];
        }

        return self.ema_value + velocity;
    }

    pub fn isPrimed(self: *const ModifiedExponentialMovingAverage) bool {
        return self.primed;
    }

    fn mnemonic(self: *const ModifiedExponentialMovingAverage) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const ModifiedExponentialMovingAverage) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const ModifiedExponentialMovingAverage, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        build_metadata_mod.buildMetadata(
            out,
            .modified_exponential_moving_average,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    pub fn updateScalar(self: *ModifiedExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *ModifiedExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *ModifiedExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *ModifiedExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *ModifiedExponentialMovingAverage) indicator_mod.Indicator {
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
        const self: *ModifiedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const ModifiedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *ModifiedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *ModifiedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *ModifiedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *ModifiedExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidPeriod,
        InvalidDegree,
        InvalidSkip,
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
    skip: usize,
    expected: []const f64,
};

fn checkSeries(params: ModifiedExponentialMovingAverageParams, inputs: []const f64, expected: []const f64) !void {
    const allocator = testing.allocator;

    var mema = try ModifiedExponentialMovingAverage.init(allocator, params);
    defer mema.deinit();

    try testing.expectEqual(inputs.len, expected.len);

    for (0..inputs.len) |i| {
        const value = mema.update(inputs[i]);
        const exp = expected[i];
        if (math.isNan(exp)) {
            try testing.expect(math.isNan(value));
        } else {
            try testing.expect(@abs(value - exp) <= tolerance);
        }
    }
}

test "MEMA reference data all combos" {
    const input = &testdata.input_close;

    const combos = [_]Combo{
        .{ .period = 3, .degree = 3, .skip = 1, .expected = &testdata.expected_p3_d3_sk1 },
        .{ .period = 3, .degree = 3, .skip = 2, .expected = &testdata.expected_p3_d3_sk2 },
        .{ .period = 3, .degree = 3, .skip = 4, .expected = &testdata.expected_p3_d3_sk4 },
        .{ .period = 3, .degree = 4, .skip = 1, .expected = &testdata.expected_p3_d4_sk1 },
        .{ .period = 3, .degree = 4, .skip = 2, .expected = &testdata.expected_p3_d4_sk2 },
        .{ .period = 3, .degree = 4, .skip = 4, .expected = &testdata.expected_p3_d4_sk4 },
        .{ .period = 6, .degree = 3, .skip = 1, .expected = &testdata.expected_p6_d3_sk1 },
        .{ .period = 6, .degree = 3, .skip = 2, .expected = &testdata.expected_p6_d3_sk2 },
        .{ .period = 6, .degree = 3, .skip = 4, .expected = &testdata.expected_p6_d3_sk4 },
        .{ .period = 6, .degree = 4, .skip = 1, .expected = &testdata.expected_p6_d4_sk1 },
        .{ .period = 6, .degree = 4, .skip = 2, .expected = &testdata.expected_p6_d4_sk2 },
        .{ .period = 6, .degree = 4, .skip = 4, .expected = &testdata.expected_p6_d4_sk4 },
        .{ .period = 12, .degree = 3, .skip = 1, .expected = &testdata.expected_p12_d3_sk1 },
        .{ .period = 12, .degree = 3, .skip = 2, .expected = &testdata.expected_p12_d3_sk2 },
        .{ .period = 12, .degree = 3, .skip = 4, .expected = &testdata.expected_p12_d3_sk4 },
        .{ .period = 12, .degree = 4, .skip = 1, .expected = &testdata.expected_p12_d4_sk1 },
        .{ .period = 12, .degree = 4, .skip = 2, .expected = &testdata.expected_p12_d4_sk2 },
        .{ .period = 12, .degree = 4, .skip = 4, .expected = &testdata.expected_p12_d4_sk4 },
    };

    for (combos) |combo| {
        try checkSeries(.{ .period = combo.period, .degree = combo.degree, .skip = combo.skip }, input, combo.expected);
    }

    try checkSeries(.{ .period = 6, .degree = 3, .skip = 1 }, &testdata.test1_input_linear, &testdata.test1_expected_p6_d3_sk1);
}

test "MEMA metadata default" {
    const allocator = testing.allocator;

    var mema = try ModifiedExponentialMovingAverage.init(allocator, .{});
    defer mema.deinit();
    mema.fixSlices();

    var meta: Metadata = undefined;
    mema.getMetadata(&meta);

    try testing.expectEqual(Identifier.modified_exponential_moving_average, meta.identifier);
    try testing.expectEqualStrings("mema(6,3,1)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "MEMA custom mnemonic" {
    const allocator = testing.allocator;

    var mema = try ModifiedExponentialMovingAverage.init(allocator, .{ .period = 12, .degree = 4, .skip = 2 });
    defer mema.deinit();

    var meta: Metadata = undefined;
    mema.getMetadata(&meta);

    try testing.expectEqualStrings("mema(12,4,2)", meta.mnemonic);
}

test "MEMA invalid params" {
    const allocator = testing.allocator;

    const r1 = ModifiedExponentialMovingAverage.init(allocator, .{ .period = 1 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = ModifiedExponentialMovingAverage.init(allocator, .{ .degree = 1 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = ModifiedExponentialMovingAverage.init(allocator, .{ .skip = 0 });
    try testing.expect(if (r3) |_| false else |_| true);
}
