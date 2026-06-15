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

/// Enumerates the outputs of the Instantaneous Sine Wave Period indicator.
pub const InstantaneousSineWavePeriodOutput = enum(u8) {
    /// The estimated cycle period in bars (may be NaN).
    period = 1,
    /// The circular frequency in radians/bar (may be NaN).
    omega = 2,
    /// The wave velocity (may be NaN).
    velocity = 3,
    /// The wave acceleration (may be NaN).
    acceleration = 4,
    /// The estimated sine wave amplitude (may be NaN).
    amplitude = 5,
    /// The phase angle in radians (may be NaN).
    phase = 6,
    /// The constant level D (may be NaN).
    dc_level = 7,
};

/// Parameters to create an Instantaneous Sine Wave Period indicator.
pub const InstantaneousSineWavePeriodParams = struct {
    smoothing: i64 = 0,
    min_period: f64 = 4.0,
    max_period: f64 = 50.0,
    error_threshold: f64 = 20.0,
    dx: f64 = 0.01,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

const ModelParams = struct {
    amplitude: f64,
    phase: f64,
    velocity: f64,
    acceleration: f64,
    dc_level: f64,
};

const UpdateResult = struct {
    period: f64,
    omega: f64,
    velocity: f64,
    acceleration: f64,
    amplitude: f64,
    phase: f64,
    dc_level: f64,
};

/// Instantaneous Sine Wave Period (ISWP) by Don Mak.
///
/// Estimates the dominant cycle period of price data by modeling it locally as a
/// single sine wave superimposed on a constant level, combining a 4-point method
/// (IF4) and a 5-point method (IF5) and selecting the one with the lower
/// estimation error at each bar.
pub const InstantaneousSineWavePeriod = struct {
    smoothing: i64,
    min_period: f64,
    max_period: f64,
    error_threshold: f64,
    dx: f64,

    ema_alpha: f64,
    ema_value: f64 = 0.0,
    ema_primed: bool = false,

    buffer: [5]f64 = .{ 0.0, 0.0, 0.0, 0.0, 0.0 },
    count: usize = 0,

    primed: bool = false,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: InstantaneousSineWavePeriodParams) !InstantaneousSineWavePeriod {
        _ = allocator;

        const smoothing = params.smoothing;
        const min_period = params.min_period;
        const max_period = params.max_period;
        const error_threshold = params.error_threshold;
        const dx = params.dx;

        if (smoothing < 0) return error.InvalidSmoothing;
        if (min_period <= 0.0) return error.InvalidMinPeriod;
        if (max_period <= min_period) return error.InvalidMaxPeriod;
        if (error_threshold <= 0.0) return error.InvalidErrorThreshold;
        if (dx <= 0.0) return error.InvalidDx;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        const ema_alpha: f64 = if (smoothing > 0)
            2.0 / (@as(f64, @floatFromInt(smoothing)) + 1.0)
        else
            1.0;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "iswp({d},{d:.2},{d:.2},{d:.2},{d:.2}{s})", .{
            smoothing,
            min_period,
            max_period,
            error_threshold,
            dx,
            triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Instantaneous Sine Wave Period {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .smoothing = smoothing,
            .min_period = min_period,
            .max_period = max_period,
            .error_threshold = error_threshold,
            .dx = dx,
            .ema_alpha = ema_alpha,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *InstantaneousSineWavePeriod) void {
        _ = self;
    }

    pub fn fixSlices(self: *InstantaneousSineWavePeriod) void {
        _ = self;
        // ISWP doesn't use LineIndicator; mnemonic/description are read from buffers
        // directly and all state is inline, so no slice fixup is needed.
    }

    fn applyEma(self: *InstantaneousSineWavePeriod, price: f64) f64 {
        if (!self.ema_primed) {
            self.ema_value = price;
            self.ema_primed = true;
        } else {
            self.ema_value = self.ema_alpha * price + (1.0 - self.ema_alpha) * self.ema_value;
        }
        return self.ema_value;
    }

    fn pushBuffer(self: *InstantaneousSineWavePeriod, value: f64) void {
        var i: usize = 4;
        while (i > 0) : (i -= 1) {
            self.buffer[i] = self.buffer[i - 1];
        }
        self.buffer[0] = value;
    }

    fn calcOmega4(self: *const InstantaneousSineWavePeriod) struct { omega: f64, err: f64 } {
        const x0 = self.buffer[0];
        const xm1 = self.buffer[1];
        const xm2 = self.buffer[2];
        const xm3 = self.buffer[3];

        const den = xm1 - xm2;
        if (den == 0.0) return .{ .omega = math.nan(f64), .err = self.error_threshold };

        const ratio = (x0 - xm3) / den;

        const sqrt_arg = 3.0 - ratio;
        if (sqrt_arg < 0.0) return .{ .omega = math.nan(f64), .err = self.error_threshold };

        const arg = 0.5 * math.sqrt(sqrt_arg);
        if (arg > 1.0) return .{ .omega = math.nan(f64), .err = self.error_threshold };

        const omega4 = 2.0 * math.asin(arg);

        const dx2 = self.dx * self.dx;

        const denom1 = 1.0 - 0.25 * sqrt_arg;
        if (denom1 <= 0.0 or sqrt_arg == 0.0) return .{ .omega = omega4, .err = self.error_threshold };

        const f1 = 1.0 / (denom1 * sqrt_arg);
        const inv_den2 = 1.0 / (den * den);
        const q2 = inv_den2 * (dx2 + dx2) + (ratio * ratio) * inv_den2 * (dx2 + dx2);

        const product = f1 * q2;
        if (product < 0.0) return .{ .omega = omega4, .err = self.error_threshold };

        return .{ .omega = omega4, .err = 0.5 * math.sqrt(product) };
    }

    fn calcOmega5(self: *const InstantaneousSineWavePeriod) struct { omega: f64, err: f64 } {
        const x0 = self.buffer[0];
        const xm1 = self.buffer[1];
        const xm3 = self.buffer[3];
        const xm4 = self.buffer[4];

        const den1 = xm1 - xm3;
        if (den1 == 0.0) return .{ .omega = math.nan(f64), .err = self.error_threshold };

        const arg = 0.5 * (x0 - xm4) / den1;
        if (@abs(arg) > 1.0) return .{ .omega = math.nan(f64), .err = self.error_threshold };

        const omega5 = math.acos(arg);

        const dx2 = self.dx * self.dx;

        const denom = 1.0 - arg * arg;
        if (denom <= 0.0) return .{ .omega = omega5, .err = self.error_threshold };

        const f1 = 1.0 / denom;
        const inv_den1_sq = 1.0 / (den1 * den1);
        const numerator_ratio = (x0 - xm4) / (den1 * den1);
        const r2 = inv_den1_sq * (dx2 + dx2) + (numerator_ratio * numerator_ratio) * (dx2 + dx2);

        const product = f1 * r2;
        if (product < 0.0) return .{ .omega = omega5, .err = self.error_threshold };

        return .{ .omega = omega5, .err = 0.5 * math.sqrt(product) };
    }

    fn calcModelParams(self: *const InstantaneousSineWavePeriod, omega: f64) ModelParams {
        const x0 = self.buffer[0];
        const xm1 = self.buffer[1];
        const xm2 = self.buffer[2];

        const half_w = omega / 2.0;
        const three_half_w = 1.5 * omega;

        const sin_hw = math.sin(half_w);
        const cos_hw = math.cos(half_w);
        const sin_3hw = math.sin(three_half_w);
        const cos_3hw = math.cos(three_half_w);

        const d0 = sin_hw * sin_hw * cos_hw * sin_3hw - sin_hw * sin_hw * sin_hw * cos_3hw;

        const nan = math.nan(f64);
        if (@abs(d0) < 1e-15) {
            return .{ .amplitude = nan, .phase = nan, .velocity = nan, .acceleration = nan, .dc_level = nan };
        }

        const inv_d0 = 1.0 / d0;

        const dx0_m1 = x0 - xm1;
        const dxm1_m2 = xm1 - xm2;

        const c = inv_d0 * (dx0_m1 * sin_hw * sin_3hw - dxm1_m2 * sin_hw * sin_hw);
        const s = inv_d0 * (dxm1_m2 * sin_hw * cos_hw - dx0_m1 * sin_hw * cos_3hw);

        const amplitude = 0.5 * math.sqrt(c * c + s * s);
        const phase = math.atan2(s, c);
        const velocity = amplitude * omega * math.cos(phase);
        const acceleration = -amplitude * omega * omega * math.sin(phase);
        const dc_level = x0 - s / 2.0;

        return .{ .amplitude = amplitude, .phase = phase, .velocity = velocity, .acceleration = acceleration, .dc_level = dc_level };
    }

    pub fn updateValues(self: *InstantaneousSineWavePeriod, sample: f64) UpdateResult {
        const nan = math.nan(f64);
        const invalid = UpdateResult{ .period = nan, .omega = nan, .velocity = nan, .acceleration = nan, .amplitude = nan, .phase = nan, .dc_level = nan };

        const smoothed = if (self.smoothing > 0) self.applyEma(sample) else sample;

        self.pushBuffer(smoothed);
        self.count += 1;

        if (self.count < 5) return invalid;

        const r4 = self.calcOmega4();
        const r5 = self.calcOmega5();

        if (r4.err >= self.error_threshold and r5.err >= self.error_threshold) return invalid;

        const omega = if (r5.err < r4.err) r5.omega else r4.omega;

        if (math.isNan(omega) or omega <= 0.0) return invalid;

        const period = (2.0 * math.pi) / omega;
        if (period < self.min_period or period > self.max_period) return invalid;

        const mp = self.calcModelParams(omega);

        self.primed = true;

        return .{
            .period = period,
            .omega = omega,
            .velocity = mp.velocity,
            .acceleration = mp.acceleration,
            .amplitude = mp.amplitude,
            .phase = mp.phase,
            .dc_level = mp.dc_level,
        };
    }

    pub fn isPrimed(self: *const InstantaneousSineWavePeriod) bool {
        return self.primed;
    }

    fn mnemonic(self: *const InstantaneousSineWavePeriod) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const InstantaneousSineWavePeriod) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const InstantaneousSineWavePeriod, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        const labels = [_][]const u8{ "period", "omega", "velocity", "acceleration", "amplitude", "phase", "dcLevel" };
        const desc_labels = [_][]const u8{ "Period", "Omega", "Velocity", "Acceleration", "Amplitude", "Phase", "DC Level" };

        var mn_bufs: [7][160]u8 = undefined;
        var desc_bufs: [7][256]u8 = undefined;
        var texts: [7]build_metadata_mod.OutputText = undefined;

        for (0..7) |i| {
            const out_mn = std.fmt.bufPrint(&mn_bufs[i], "{s} {s}", .{ mn, labels[i] }) catch mn;
            const out_desc = std.fmt.bufPrint(&desc_bufs[i], "{s} {s}", .{ desc, desc_labels[i] }) catch desc;
            texts[i] = .{ .mnemonic = out_mn, .description = out_desc };
        }

        build_metadata_mod.buildMetadata(
            out,
            .instantaneous_sine_wave_period,
            mn,
            desc,
            &texts,
        );
    }

    pub fn updateScalar(self: *InstantaneousSineWavePeriod, sample: *const Scalar) OutputArray {
        const r = self.updateValues(sample.value);
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.period } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.omega } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.velocity } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.acceleration } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.amplitude } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.phase } });
        out.append(.{ .scalar = .{ .time = sample.time, .value = r.dc_level } });
        return out;
    }

    pub fn updateBar(self: *InstantaneousSineWavePeriod, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *InstantaneousSineWavePeriod, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *InstantaneousSineWavePeriod, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn indicator(self: *InstantaneousSineWavePeriod) indicator_mod.Indicator {
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
        const self: *InstantaneousSineWavePeriod = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const InstantaneousSineWavePeriod = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *InstantaneousSineWavePeriod = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *InstantaneousSineWavePeriod = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *InstantaneousSineWavePeriod = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *InstantaneousSineWavePeriod = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidSmoothing,
        InvalidMinPeriod,
        InvalidMaxPeriod,
        InvalidErrorThreshold,
        InvalidDx,
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

const Combo = struct {
    smoothing: i64,
    period: [252]f64,
    omega: [252]f64,
    velocity: [252]f64,
    acceleration: [252]f64,
};

test "ISWP reference data all combos" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    const combos = [_]Combo{
        .{ .smoothing = 0, .period = testdata.expectedS0_PERIOD(), .omega = testdata.expectedS0_OMEGA(), .velocity = testdata.expectedS0_VELOCITY(), .acceleration = testdata.expectedS0_ACCELERATION() },
        .{ .smoothing = 3, .period = testdata.expectedS3_PERIOD(), .omega = testdata.expectedS3_OMEGA(), .velocity = testdata.expectedS3_VELOCITY(), .acceleration = testdata.expectedS3_ACCELERATION() },
        .{ .smoothing = 6, .period = testdata.expectedS6_PERIOD(), .omega = testdata.expectedS6_OMEGA(), .velocity = testdata.expectedS6_VELOCITY(), .acceleration = testdata.expectedS6_ACCELERATION() },
        .{ .smoothing = 12, .period = testdata.expectedS12_PERIOD(), .omega = testdata.expectedS12_OMEGA(), .velocity = testdata.expectedS12_VELOCITY(), .acceleration = testdata.expectedS12_ACCELERATION() },
    };

    for (combos) |combo| {
        var iswp = try InstantaneousSineWavePeriod.init(allocator, .{ .smoothing = combo.smoothing });
        defer iswp.deinit();

        for (0..252) |i| {
            const r = iswp.updateValues(input[i]);
            try checkVal(combo.period[i], r.period);
            try checkVal(combo.omega[i], r.omega);
            try checkVal(combo.velocity[i], r.velocity);
            try checkVal(combo.acceleration[i], r.acceleration);
        }
    }
}

test "ISWP metadata default" {
    const allocator = testing.allocator;

    var iswp = try InstantaneousSineWavePeriod.init(allocator, .{});
    defer iswp.deinit();

    var meta: Metadata = undefined;
    iswp.getMetadata(&meta);

    try testing.expectEqual(Identifier.instantaneous_sine_wave_period, meta.identifier);
    try testing.expectEqualStrings("iswp(0,4.00,50.00,20.00,0.01)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 7), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 7), meta.outputs_buf[6].kind);
}

test "ISWP custom mnemonic" {
    const allocator = testing.allocator;

    var iswp = try InstantaneousSineWavePeriod.init(allocator, .{ .smoothing = 6 });
    defer iswp.deinit();

    var meta: Metadata = undefined;
    iswp.getMetadata(&meta);

    try testing.expectEqualStrings("iswp(6,4.00,50.00,20.00,0.01)", meta.mnemonic);
}

test "ISWP invalid params" {
    const allocator = testing.allocator;

    const r1 = InstantaneousSineWavePeriod.init(allocator, .{ .smoothing = -1 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = InstantaneousSineWavePeriod.init(allocator, .{ .min_period = 0.0 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = InstantaneousSineWavePeriod.init(allocator, .{ .min_period = 50.0, .max_period = 50.0 });
    try testing.expect(if (r3) |_| false else |_| true);

    const r4 = InstantaneousSineWavePeriod.init(allocator, .{ .error_threshold = 0.0 });
    try testing.expect(if (r4) |_| false else |_| true);

    const r5 = InstantaneousSineWavePeriod.init(allocator, .{ .dx = 0.0 });
    try testing.expect(if (r5) |_| false else |_| true);
}

test "ISWP entity update ordering" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    const exp_period = testdata.expectedS0_PERIOD();
    const exp_omega = testdata.expectedS0_OMEGA();

    var iswp = try InstantaneousSineWavePeriod.init(allocator, .{});
    defer iswp.deinit();

    var last_out: OutputArray = undefined;
    for (0..252) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        last_out = iswp.updateScalar(&scalar);
    }
    const items = last_out.slice();

    try testing.expectEqual(@as(usize, 7), items.len);
    try checkVal(exp_period[251], items[0].scalar.value);
    try checkVal(exp_omega[251], items[1].scalar.value);
}
