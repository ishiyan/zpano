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
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Adaptive Exponential Moving Average indicator.
pub const AdaptiveExponentialMovingAverageOutput = enum(u8) {
    /// The adaptively smoothed price value.
    value = 1,
    /// The instantaneous frequency estimate (may be NaN).
    omega = 2,
    /// The smoothing factor used for the bar.
    alpha = 3,
};

/// Parameters to create an Adaptive Exponential Moving Average indicator.
pub const AdaptiveExponentialMovingAverageParams = struct {
    alpha_max: f64 = 0.5,
    alpha_min: f64 = 0.05,
    omega0: f64 = 1.0,
    smoothing: i64 = 3,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

const iswp_min_period = 4.0;
const iswp_max_period = 50.0;
const iswp_error_threshold = 20.0;
const iswp_dx = 0.01;

/// Embedded Instantaneous Sine Wave Period omega estimator (omega-only reduction).
///
/// Estimates the dominant circular frequency omega of price data by modeling it
/// locally as a single sine wave, combining a 4-point and a 5-point method and
/// selecting the one with the lower estimation error. Inlined so the indicator is a
/// standalone porting unit. Do NOT change its numerics.
const Iswp = struct {
    smoothing: i64,
    ema_alpha: f64,
    ema_value: f64 = 0.0,
    ema_primed: bool = false,
    buffer: [5]f64 = .{ 0.0, 0.0, 0.0, 0.0, 0.0 },
    count: usize = 0,

    fn init(smoothing: i64) Iswp {
        const ema_alpha: f64 = if (smoothing > 0)
            2.0 / (@as(f64, @floatFromInt(smoothing)) + 1.0)
        else
            1.0;
        return .{ .smoothing = smoothing, .ema_alpha = ema_alpha };
    }

    fn applyEma(self: *Iswp, price: f64) f64 {
        if (!self.ema_primed) {
            self.ema_value = price;
            self.ema_primed = true;
        } else {
            self.ema_value = self.ema_alpha * price + (1.0 - self.ema_alpha) * self.ema_value;
        }
        return self.ema_value;
    }

    fn pushBuffer(self: *Iswp, value: f64) void {
        var i: usize = 4;
        while (i > 0) : (i -= 1) {
            self.buffer[i] = self.buffer[i - 1];
        }
        self.buffer[0] = value;
    }

    fn calcOmega4(self: *const Iswp) struct { omega: f64, err: f64 } {
        const x0 = self.buffer[0];
        const xm1 = self.buffer[1];
        const xm2 = self.buffer[2];
        const xm3 = self.buffer[3];

        const den = xm1 - xm2;
        if (den == 0.0) return .{ .omega = math.nan(f64), .err = iswp_error_threshold };

        const ratio = (x0 - xm3) / den;

        const sqrt_arg = 3.0 - ratio;
        if (sqrt_arg < 0.0) return .{ .omega = math.nan(f64), .err = iswp_error_threshold };

        const arg = 0.5 * math.sqrt(sqrt_arg);
        if (arg > 1.0) return .{ .omega = math.nan(f64), .err = iswp_error_threshold };

        const omega4 = 2.0 * math.asin(arg);

        const dx2 = iswp_dx * iswp_dx;

        const denom1 = 1.0 - 0.25 * sqrt_arg;
        if (denom1 <= 0.0 or sqrt_arg == 0.0) return .{ .omega = omega4, .err = iswp_error_threshold };

        const f1 = 1.0 / (denom1 * sqrt_arg);
        const inv_den2 = 1.0 / (den * den);
        const q2 = inv_den2 * (dx2 + dx2) + (ratio * ratio) * inv_den2 * (dx2 + dx2);

        const product = f1 * q2;
        if (product < 0.0) return .{ .omega = omega4, .err = iswp_error_threshold };

        return .{ .omega = omega4, .err = 0.5 * math.sqrt(product) };
    }

    fn calcOmega5(self: *const Iswp) struct { omega: f64, err: f64 } {
        const x0 = self.buffer[0];
        const xm1 = self.buffer[1];
        const xm3 = self.buffer[3];
        const xm4 = self.buffer[4];

        const den1 = xm1 - xm3;
        if (den1 == 0.0) return .{ .omega = math.nan(f64), .err = iswp_error_threshold };

        const arg = 0.5 * (x0 - xm4) / den1;
        if (@abs(arg) > 1.0) return .{ .omega = math.nan(f64), .err = iswp_error_threshold };

        const omega5 = math.acos(arg);

        const dx2 = iswp_dx * iswp_dx;

        const denom = 1.0 - arg * arg;
        if (denom <= 0.0) return .{ .omega = omega5, .err = iswp_error_threshold };

        const f1 = 1.0 / denom;
        const inv_den1_sq = 1.0 / (den1 * den1);
        const numerator_ratio = (x0 - xm4) / (den1 * den1);
        const r2 = inv_den1_sq * (dx2 + dx2) + (numerator_ratio * numerator_ratio) * (dx2 + dx2);

        const product = f1 * r2;
        if (product < 0.0) return .{ .omega = omega5, .err = iswp_error_threshold };

        return .{ .omega = omega5, .err = 0.5 * math.sqrt(product) };
    }

    fn update(self: *Iswp, price: f64) f64 {
        const smoothed = if (self.smoothing > 0) self.applyEma(price) else price;

        self.pushBuffer(smoothed);
        self.count += 1;

        if (self.count < 5) return math.nan(f64);

        const r4 = self.calcOmega4();
        const r5 = self.calcOmega5();

        if (r4.err >= iswp_error_threshold and r5.err >= iswp_error_threshold) {
            return math.nan(f64);
        }

        const omega = if (r5.err < r4.err) r5.omega else r4.omega;

        if (math.isNan(omega) or omega <= 0.0) return math.nan(f64);

        const period = (2.0 * math.pi) / omega;
        if (period < iswp_min_period or period > iswp_max_period) return math.nan(f64);

        return omega;
    }
};

/// Adaptive Exponential Moving Average (AEMA) by Don Mak.
///
/// An EMA with a time-varying smoothing factor alpha that adapts based on the
/// instantaneous frequency of the price data, estimated by an embedded ISWP.
///
/// The indicator produces three outputs:
///   - value: the adaptively smoothed price (never NaN);
///   - omega: the instantaneous frequency estimate (may be NaN);
///   - alpha: the smoothing factor used for this bar.
pub const AdaptiveExponentialMovingAverage = struct {
    alpha_max: f64,
    alpha_min: f64,
    omega0: f64,
    a: f64,
    b: f64,

    iswp: Iswp,

    ema_value: f64,
    initialized: bool,
    primed: bool,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: AdaptiveExponentialMovingAverageParams) !AdaptiveExponentialMovingAverage {
        _ = allocator;

        const alpha_max = params.alpha_max;
        const alpha_min = params.alpha_min;
        const omega0 = params.omega0;
        const smoothing = params.smoothing;

        if (!(alpha_min > 0.0 and alpha_min < alpha_max and alpha_max <= 1.0)) return error.InvalidAlpha;
        if (!(omega0 > 0.0 and omega0 < math.pi)) return error.InvalidOmega0;
        if (smoothing < 0) return error.InvalidSmoothing;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        const a = (alpha_max - alpha_min) * omega0 * math.pi / (math.pi - omega0);
        const b = alpha_min - a / math.pi;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "aema({d:.2},{d:.2},{d:.2},{d}{s})", .{
            alpha_max,
            alpha_min,
            omega0,
            smoothing,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Adaptive Exponential Moving Average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .alpha_max = alpha_max,
            .alpha_min = alpha_min,
            .omega0 = omega0,
            .a = a,
            .b = b,
            .iswp = Iswp.init(smoothing),
            .ema_value = 0.0,
            .initialized = false,
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *AdaptiveExponentialMovingAverage) void {
        _ = self;
    }

    pub fn fixSlices(self: *AdaptiveExponentialMovingAverage) void {
        _ = self;
        // AEMA doesn't use LineIndicator; mnemonic/description are read from buffers
        // directly and all state is inline, so no slice fixup is needed.
    }

    fn computeAlpha(self: *const AdaptiveExponentialMovingAverage, omega: f64) f64 {
        if (math.isNan(omega)) return self.alpha_min;
        if (omega <= self.omega0) return self.alpha_max;
        if (omega >= math.pi) return self.alpha_min;

        const alpha = self.a / omega + self.b;
        if (alpha > self.alpha_max) return self.alpha_max;
        if (alpha < self.alpha_min) return self.alpha_min;
        return alpha;
    }

    /// Returns value, omega, alpha.
    pub fn updateValues(self: *AdaptiveExponentialMovingAverage, sample: f64) struct { value: f64, omega: f64, alpha: f64 } {
        const omega = self.iswp.update(sample);
        const alpha = self.computeAlpha(omega);

        if (!self.initialized) {
            self.ema_value = sample;
            self.initialized = true;
        } else {
            self.ema_value = alpha * sample + (1.0 - alpha) * self.ema_value;
        }

        if (!math.isNan(omega)) self.primed = true;

        return .{ .value = self.ema_value, .omega = omega, .alpha = alpha };
    }

    pub fn isPrimed(self: *const AdaptiveExponentialMovingAverage) bool {
        return self.primed;
    }

    fn mnemonic(self: *const AdaptiveExponentialMovingAverage) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const AdaptiveExponentialMovingAverage) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const AdaptiveExponentialMovingAverage, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var value_mn_buf: [160]u8 = undefined;
        const value_mn = std.fmt.bufPrint(&value_mn_buf, "{s} value", .{mn}) catch mn;
        var omega_mn_buf: [160]u8 = undefined;
        const omega_mn = std.fmt.bufPrint(&omega_mn_buf, "{s} omega", .{mn}) catch mn;
        var alpha_mn_buf: [160]u8 = undefined;
        const alpha_mn = std.fmt.bufPrint(&alpha_mn_buf, "{s} alpha", .{mn}) catch mn;

        var value_desc_buf: [256]u8 = undefined;
        const value_desc = std.fmt.bufPrint(&value_desc_buf, "{s} Value", .{desc}) catch desc;
        var omega_desc_buf: [256]u8 = undefined;
        const omega_desc = std.fmt.bufPrint(&omega_desc_buf, "{s} Omega", .{desc}) catch desc;
        var alpha_desc_buf: [256]u8 = undefined;
        const alpha_desc = std.fmt.bufPrint(&alpha_desc_buf, "{s} Alpha", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .adaptive_exponential_moving_average,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = value_mn, .description = value_desc },
                .{ .mnemonic = omega_mn, .description = omega_desc },
                .{ .mnemonic = alpha_mn, .description = alpha_desc },
            },
        );
    }

    pub fn updateScalar(self: *AdaptiveExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const result = self.updateValues(sample.value);
        return makeOutput(sample.time, result.value, result.omega, result.alpha);
    }

    pub fn updateBar(self: *AdaptiveExponentialMovingAverage, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *AdaptiveExponentialMovingAverage, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *AdaptiveExponentialMovingAverage, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    fn makeOutput(time: i64, value_v: f64, omega_v: f64, alpha_v: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = value_v } });
        out.append(.{ .scalar = .{ .time = time, .value = omega_v } });
        out.append(.{ .scalar = .{ .time = time, .value = alpha_v } });
        return out;
    }

    pub fn indicator(self: *AdaptiveExponentialMovingAverage) indicator_mod.Indicator {
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
        const self: *AdaptiveExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const AdaptiveExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AdaptiveExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AdaptiveExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AdaptiveExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AdaptiveExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidAlpha,
        InvalidOmega0,
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

fn checkVal(exp: f64, act: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }
    try testing.expect(@abs(act - exp) <= tolerance);
}

const ValueCombo = struct {
    alpha_max: f64,
    alpha_min: f64,
    omega0: f64,
    smoothing: i64,
    expected: [252]f64,
};

test "AEMA value output all combos" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    const combos = [_]ValueCombo{
        .{ .alpha_max = 0.5, .alpha_min = 0.05, .omega0 = 1.0, .smoothing = 3, .expected = testdata.expectedDEFAULT() },
        .{ .alpha_max = 0.8, .alpha_min = 0.02, .omega0 = 1.0, .smoothing = 3, .expected = testdata.expectedA0_8_A0_02() },
        .{ .alpha_max = 0.5, .alpha_min = 0.05, .omega0 = 0.5, .smoothing = 3, .expected = testdata.expectedW0_5() },
        .{ .alpha_max = 0.5, .alpha_min = 0.05, .omega0 = 1.5, .smoothing = 3, .expected = testdata.expectedW1_5() },
        .{ .alpha_max = 0.5, .alpha_min = 0.05, .omega0 = 1.0, .smoothing = 0, .expected = testdata.expectedS0() },
        .{ .alpha_max = 0.5, .alpha_min = 0.05, .omega0 = 1.0, .smoothing = 6, .expected = testdata.expectedS6() },
    };

    for (combos) |combo| {
        var aema = try AdaptiveExponentialMovingAverage.init(allocator, .{
            .alpha_max = combo.alpha_max,
            .alpha_min = combo.alpha_min,
            .omega0 = combo.omega0,
            .smoothing = combo.smoothing,
        });
        defer aema.deinit();

        for (0..252) |i| {
            const result = aema.updateValues(input[i]);
            try checkVal(combo.expected[i], result.value);
        }
    }
}

test "AEMA omega and alpha default" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const exp_omega = testdata.expectedDEFAULT_OMEGA();
    const exp_alpha = testdata.expectedDEFAULT_ALPHA();

    var aema = try AdaptiveExponentialMovingAverage.init(allocator, .{});
    defer aema.deinit();

    for (0..252) |i| {
        const result = aema.updateValues(input[i]);
        try checkVal(exp_omega[i], result.omega);
        try checkVal(exp_alpha[i], result.alpha);
    }
}

test "AEMA sine wave default" {
    const allocator = testing.allocator;
    const input = testdata.test1InputSine();
    const exp_value = testdata.test1Expected();
    const exp_omega = testdata.test1ExpectedOmega();
    const exp_alpha = testdata.test1ExpectedAlpha();

    var aema = try AdaptiveExponentialMovingAverage.init(allocator, .{});
    defer aema.deinit();

    for (0..100) |i| {
        const result = aema.updateValues(input[i]);
        try checkVal(exp_value[i], result.value);
        try checkVal(exp_omega[i], result.omega);
        try checkVal(exp_alpha[i], result.alpha);
    }
}

test "AEMA metadata default" {
    const allocator = testing.allocator;

    var aema = try AdaptiveExponentialMovingAverage.init(allocator, .{});
    defer aema.deinit();

    var meta: Metadata = undefined;
    aema.getMetadata(&meta);

    try testing.expectEqual(Identifier.adaptive_exponential_moving_average, meta.identifier);
    try testing.expectEqualStrings("aema(0.50,0.05,1.00,3)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 3), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
    try testing.expectEqual(@as(u8, 3), meta.outputs_buf[2].kind);
}

test "AEMA custom mnemonic" {
    const allocator = testing.allocator;

    var aema = try AdaptiveExponentialMovingAverage.init(allocator, .{
        .alpha_max = 0.8,
        .alpha_min = 0.02,
        .omega0 = 1.5,
        .smoothing = 6,
    });
    defer aema.deinit();

    var meta: Metadata = undefined;
    aema.getMetadata(&meta);

    try testing.expectEqualStrings("aema(0.80,0.02,1.50,6)", meta.mnemonic);
}

test "AEMA invalid params" {
    const allocator = testing.allocator;

    const r1 = AdaptiveExponentialMovingAverage.init(allocator, .{ .alpha_max = 0.05, .alpha_min = 0.5 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = AdaptiveExponentialMovingAverage.init(allocator, .{ .alpha_max = 1.5 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = AdaptiveExponentialMovingAverage.init(allocator, .{ .omega0 = 4.0 });
    try testing.expect(if (r3) |_| false else |_| true);

    const r4 = AdaptiveExponentialMovingAverage.init(allocator, .{ .smoothing = -1 });
    try testing.expect(if (r4) |_| false else |_| true);
}

test "AEMA entity update ordering" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const exp_value = testdata.expectedDEFAULT();
    const exp_omega = testdata.expectedDEFAULT_OMEGA();
    const exp_alpha = testdata.expectedDEFAULT_ALPHA();

    var aema = try AdaptiveExponentialMovingAverage.init(allocator, .{});
    defer aema.deinit();

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        last_out = aema.updateScalar(&scalar);
    }
    const items = last_out.slice();

    try checkVal(exp_value[251], items[0].scalar.value);
    try checkVal(exp_omega[251], items[1].scalar.value);
    try checkVal(exp_alpha[251], items[2].scalar.value);
}
