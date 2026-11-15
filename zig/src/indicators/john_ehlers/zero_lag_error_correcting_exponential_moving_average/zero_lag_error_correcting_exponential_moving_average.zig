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
const adaptivity_mod = @import("../../core/adaptivity.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const descriptor_mod = @import("../../core/descriptor.zig");
const identifier_mod = @import("../../core/identifier.zig");
const indicator_mod = @import("../../core/indicator.zig");
const metadata_mod = @import("../../core/metadata.zig");
const shape_mod = @import("../../core/outputs/shape.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

const alpha_epsilon: f64 = 1e-8;
const least_error_sentinel: f64 = 1_000_000.0;

pub const Error = error{
    InvalidLength,
    InvalidSmoothingFactor,
    InvalidGainLimit,
    InvalidSample,
    ArithmeticOverflow,
    MnemonicTooLong,
};

/// Enumerates the outputs of the zero-lag error-correcting exponential moving average indicator.
pub const ZeroLagErrorCorrectingExponentialMovingAverageOutput = enum(u8) {
    value = 1,
    ema = 2,
};

/// Length-based parameters to create an instance of ZeroLagErrorCorrectingExponentialMovingAverage.
pub const ZeroLagErrorCorrectingExponentialMovingAverageLengthParams = struct {
    /// Equivalent EMA length L (>= 1), defining alpha = 2 / (L + 1). Default is 20.
    length: i64 = 20,
    /// Nonnegative integer gain bound in tenths (candidates span [-G/10, +G/10]). Default is 50.
    gain_limit: i32 = 50,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Smoothing-factor-based parameters to create an instance of ZeroLagErrorCorrectingExponentialMovingAverage.
pub const ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams = struct {
    /// Baseline EMA smoothing factor alpha in (0, 1]. Default is 2 / 21.
    smoothing_factor: f64 = 2.0 / 21.0,
    /// Nonnegative integer gain bound in tenths (candidates span [-G/10, +G/10]). Default is 50.
    gain_limit: i32 = 50,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const LengthParams = ZeroLagErrorCorrectingExponentialMovingAverageLengthParams;
pub const SmoothingFactorParams = ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams;
pub const ZeroLagErrorCorrectingExponentialMovingAverageParams = ZeroLagErrorCorrectingExponentialMovingAverageLengthParams;
pub const Params = ZeroLagErrorCorrectingExponentialMovingAverageLengthParams;

pub fn defaultLengthParams() ZeroLagErrorCorrectingExponentialMovingAverageLengthParams {
    return .{};
}

pub fn defaultSmoothingFactorParams() ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams {
    return .{};
}

pub fn defaultParams() ZeroLagErrorCorrectingExponentialMovingAverageLengthParams {
    return .{};
}

pub fn equivalentLengthFromAlpha(alpha: f64) i64 {
    if (!(alpha >= alpha_epsilon)) {
        return math.maxInt(i64);
    }
    const rounded = @round(2.0 / alpha);
    if (rounded >= @as(f64, @floatFromInt(math.maxInt(i64)))) {
        return math.maxInt(i64);
    }
    return @as(i64, @intFromFloat(rounded)) - 1;
}

fn formatLengthMnemonicWithPrefix(
    prefix: []const u8,
    buffer: []u8,
    params: ZeroLagErrorCorrectingExponentialMovingAverageLengthParams,
) Error![]const u8 {
    var triple_storage: [64]u8 = undefined;
    const suffix = component_triple_mnemonic_mod.componentTripleMnemonic(
        &triple_storage,
        params.bar_component orelse bar_component.default_bar_component,
        params.quote_component orelse quote_component.default_quote_component,
        params.trade_component orelse trade_component.default_trade_component,
    );
    return std.fmt.bufPrint(buffer, "{s}({d}, {d}{s})", .{
        prefix,
        params.length,
        params.gain_limit,
        suffix,
    }) catch return error.MnemonicTooLong;
}

fn formatSmoothingFactorMnemonicWithPrefix(
    prefix: []const u8,
    buffer: []u8,
    params: ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
) Error![]const u8 {
    var triple_storage: [64]u8 = undefined;
    const suffix = component_triple_mnemonic_mod.componentTripleMnemonic(
        &triple_storage,
        params.bar_component orelse bar_component.default_bar_component,
        params.quote_component orelse quote_component.default_quote_component,
        params.trade_component orelse trade_component.default_trade_component,
    );
    const equiv_length = equivalentLengthFromAlpha(params.smoothing_factor);
    return std.fmt.bufPrint(buffer, "{s}({d}, {d:.8}, {d}{s})", .{
        prefix,
        equiv_length,
        params.smoothing_factor,
        params.gain_limit,
        suffix,
    }) catch return error.MnemonicTooLong;
}

pub fn mnemonic(buffer: []u8, params: ZeroLagErrorCorrectingExponentialMovingAverageLengthParams) Error![]const u8 {
    return formatLengthMnemonicWithPrefix("zecema", buffer, params);
}

pub fn emaMnemonic(buffer: []u8, params: ZeroLagErrorCorrectingExponentialMovingAverageLengthParams) Error![]const u8 {
    return formatLengthMnemonicWithPrefix("zecemaEma", buffer, params);
}

pub fn smoothingFactorMnemonic(
    buffer: []u8,
    params: ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
) Error![]const u8 {
    return formatSmoothingFactorMnemonicWithPrefix("zecema", buffer, params);
}

pub fn smoothingFactorEmaMnemonic(
    buffer: []u8,
    params: ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
) Error![]const u8 {
    return formatSmoothingFactorMnemonicWithPrefix("zecemaEma", buffer, params);
}

/// Ehlers and Way's Zero-Lag Error-Correcting Exponential Moving Average (ZECEMA).
///
/// Follows the November 2010 TASC EasyLanguage listing ("Zero Lag (Well, Almost)"):
/// starting from explicit zero EMA and EC states, each accepted finite sample updates
/// the baseline EMA once and searches integer-tenths gains `index / 10` for inclusive
/// integers `[-gain_limit, +gain_limit]` against the frozen previous-bar EC using
/// strict improvement (`error_value < least_error`) from the literal `1_000_000`
/// sentinel. Both EC (`value`) and baseline EMA (`ema`) outputs are available from
/// the first accepted finite observation.
pub const ZeroLagErrorCorrectingExponentialMovingAverage = struct {
    configured_length: i64,
    configured_gain_limit: i32,
    configured_alpha: f64,
    complement: f64,

    count: usize,
    ema_state: f64,
    ec_state: f64,
    best_gain_state: f64,
    least_error_state: f64,
    current_value: f64,
    current_ema: f64,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    mnemonic_storage: [128]u8,
    mnemonic_len: usize,
    description_storage: [256]u8,
    description_len: usize,
    mnemonic_ema_storage: [128]u8,
    mnemonic_ema_len: usize,
    description_ema_storage: [256]u8,
    description_ema_len: usize,

    pub fn initDefault() Error!ZeroLagErrorCorrectingExponentialMovingAverage {
        return initFromLength(defaultLengthParams());
    }

    pub fn init(params: ZeroLagErrorCorrectingExponentialMovingAverageLengthParams) Error!ZeroLagErrorCorrectingExponentialMovingAverage {
        return initFromLength(params);
    }

    pub fn initLength(params: ZeroLagErrorCorrectingExponentialMovingAverageLengthParams) Error!ZeroLagErrorCorrectingExponentialMovingAverage {
        return initFromLength(params);
    }

    pub fn initFromLength(params: ZeroLagErrorCorrectingExponentialMovingAverageLengthParams) Error!ZeroLagErrorCorrectingExponentialMovingAverage {
        if (params.length < 1) {
            return error.InvalidLength;
        }
        if (params.gain_limit < 0) {
            return error.InvalidGainLimit;
        }

        const alpha_value = 2.0 / (@as(f64, @floatFromInt(params.length)) + 1.0);
        const complement_value = 1.0 - alpha_value;
        if (!math.isFinite(alpha_value) or !(alpha_value > 0.0 and alpha_value <= 1.0) or !(complement_value >= 0.0 and complement_value < 1.0)) {
            return error.InvalidLength;
        }

        var mnemonic_storage: [128]u8 = undefined;
        const primary_mnemonic = try mnemonic(&mnemonic_storage, params);
        const mnemonic_len = primary_mnemonic.len;

        var mnemonic_ema_storage: [128]u8 = undefined;
        const secondary_mnemonic = try emaMnemonic(&mnemonic_ema_storage, params);
        const mnemonic_ema_len = secondary_mnemonic.len;

        return initCommon(
            params.length,
            params.gain_limit,
            alpha_value,
            complement_value,
            mnemonic_storage,
            mnemonic_len,
            mnemonic_ema_storage,
            mnemonic_ema_len,
            params.bar_component,
            params.quote_component,
            params.trade_component,
        );
    }

    pub fn initSmoothingFactor(params: ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams) Error!ZeroLagErrorCorrectingExponentialMovingAverage {
        return initFromSmoothingFactor(params);
    }

    pub fn initFromSmoothingFactor(params: ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams) Error!ZeroLagErrorCorrectingExponentialMovingAverage {
        const alpha_value = params.smoothing_factor;
        if (!math.isFinite(alpha_value) or !(alpha_value > 0.0 and alpha_value <= 1.0)) {
            return error.InvalidSmoothingFactor;
        }
        const complement_value = 1.0 - alpha_value;
        if (!(complement_value >= 0.0 and complement_value < 1.0)) {
            return error.InvalidSmoothingFactor;
        }
        if (params.gain_limit < 0) {
            return error.InvalidGainLimit;
        }

        var mnemonic_storage: [128]u8 = undefined;
        const primary_mnemonic = try smoothingFactorMnemonic(&mnemonic_storage, params);
        const mnemonic_len = primary_mnemonic.len;

        var mnemonic_ema_storage: [128]u8 = undefined;
        const secondary_mnemonic = try smoothingFactorEmaMnemonic(&mnemonic_ema_storage, params);
        const mnemonic_ema_len = secondary_mnemonic.len;

        return initCommon(
            equivalentLengthFromAlpha(alpha_value),
            params.gain_limit,
            alpha_value,
            complement_value,
            mnemonic_storage,
            mnemonic_len,
            mnemonic_ema_storage,
            mnemonic_ema_len,
            params.bar_component,
            params.quote_component,
            params.trade_component,
        );
    }

    fn initCommon(
        length_value: i64,
        gain_limit_value: i32,
        alpha_value: f64,
        complement_value: f64,
        mnemonic_storage: [128]u8,
        mnemonic_len: usize,
        mnemonic_ema_storage: [128]u8,
        mnemonic_ema_len: usize,
        bar_comp_opt: ?bar_component.BarComponent,
        quote_comp_opt: ?quote_component.QuoteComponent,
        trade_comp_opt: ?trade_component.TradeComponent,
    ) Error!ZeroLagErrorCorrectingExponentialMovingAverage {
        const bar_comp = bar_comp_opt orelse bar_component.default_bar_component;
        const quote_comp = quote_comp_opt orelse quote_component.default_quote_component;
        const trade_comp = trade_comp_opt orelse trade_component.default_trade_component;

        var description_storage: [256]u8 = undefined;
        const primary_description = std.fmt.bufPrint(
            &description_storage,
            "Zero-lag Error-Correcting Exponential Moving Average {s}",
            .{mnemonic_storage[0..mnemonic_len]},
        ) catch return error.MnemonicTooLong;
        const description_len = primary_description.len;

        var description_ema_storage: [256]u8 = undefined;
        const secondary_description = std.fmt.bufPrint(
            &description_ema_storage,
            "Zero-lag Error-Correcting Exponential Moving Average baseline EMA {s}",
            .{mnemonic_ema_storage[0..mnemonic_ema_len]},
        ) catch return error.MnemonicTooLong;
        const description_ema_len = secondary_description.len;

        var instance = ZeroLagErrorCorrectingExponentialMovingAverage{
            .configured_length = length_value,
            .configured_gain_limit = gain_limit_value,
            .configured_alpha = alpha_value,
            .complement = complement_value,
            .count = 0,
            .ema_state = 0.0,
            .ec_state = 0.0,
            .best_gain_state = 0.0,
            .least_error_state = 0.0,
            .current_value = math.nan(f64),
            .current_ema = math.nan(f64),
            .bar_func = bar_component.componentValue(bar_comp),
            .quote_func = quote_component.componentValue(quote_comp),
            .trade_func = trade_component.componentValue(trade_comp),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mnemonic_len,
            .description_storage = description_storage,
            .description_len = description_len,
            .mnemonic_ema_storage = mnemonic_ema_storage,
            .mnemonic_ema_len = mnemonic_ema_len,
            .description_ema_storage = description_ema_storage,
            .description_ema_len = description_ema_len,
        };
        instance.reset();
        return instance;
    }

    pub fn fixSlices(self: *ZeroLagErrorCorrectingExponentialMovingAverage) void {
        _ = self;
    }

    pub fn reset(self: *ZeroLagErrorCorrectingExponentialMovingAverage) void {
        self.count = 0;
        self.ema_state = 0.0;
        self.ec_state = 0.0;
        self.best_gain_state = 0.0;
        self.least_error_state = 0.0;
        self.current_value = math.nan(f64);
        self.current_ema = math.nan(f64);
    }

    pub fn length(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) i64 {
        return self.configured_length;
    }

    pub fn gainLimit(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) i32 {
        return self.configured_gain_limit;
    }

    pub fn alpha(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) f64 {
        return self.configured_alpha;
    }

    pub fn samplesSeen(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) usize {
        return self.count;
    }

    pub fn warmupBars(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) usize {
        _ = self;
        return 0;
    }

    pub fn isPrimed(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) bool {
        return self.count > 0;
    }

    pub fn value(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) f64 {
        return self.current_value;
    }

    pub fn ema(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) f64 {
        return self.current_ema;
    }

    pub fn bestGain(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) f64 {
        return if (self.count > 0) self.best_gain_state else math.nan(f64);
    }

    pub fn leastError(self: *const ZeroLagErrorCorrectingExponentialMovingAverage) f64 {
        return if (self.count > 0) self.least_error_state else math.nan(f64);
    }

    pub fn getMetadata(self: *const ZeroLagErrorCorrectingExponentialMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .zero_lag_error_correcting_exponential_moving_average,
            self.mnemonic_storage[0..self.mnemonic_len],
            self.description_storage[0..self.description_len],
            &[_]build_metadata_mod.OutputText{
                .{
                    .mnemonic = self.mnemonic_storage[0..self.mnemonic_len],
                    .description = self.description_storage[0..self.description_len],
                },
                .{
                    .mnemonic = self.mnemonic_ema_storage[0..self.mnemonic_ema_len],
                    .description = self.description_ema_storage[0..self.description_ema_len],
                },
            },
        );
    }

    fn step(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: f64) Error!struct { ec: f64, ema: f64 } {
        if (math.isInf(sample)) {
            return error.InvalidSample;
        }
        if (math.isNan(sample)) {
            return .{ .ec = math.nan(f64), .ema = math.nan(f64) };
        }

        const alpha_value = self.configured_alpha;
        const complement_value = self.complement;
        const previous_ec = self.ec_state;

        const ema_value = alpha_value * sample + complement_value * self.ema_state;
        if (!math.isFinite(ema_value)) {
            return error.ArithmeticOverflow;
        }

        var least_error_value: f64 = least_error_sentinel;
        var best_gain_value: f64 = self.best_gain_state;
        const limit: i64 = @as(i64, self.configured_gain_limit);
        var index: i64 = -limit;
        while (index <= limit) : (index += 1) {
            const gain: f64 = @as(f64, @floatFromInt(index)) / 10.0;
            const candidate = alpha_value * (ema_value + gain * (sample - previous_ec)) + complement_value * previous_ec;
            const error_value = @abs(sample - candidate);
            if (!math.isFinite(candidate) or !math.isFinite(error_value)) {
                return error.ArithmeticOverflow;
            }
            if (error_value < least_error_value) {
                least_error_value = error_value;
                best_gain_value = gain;
            }
        }

        const ec_value = alpha_value * (ema_value + best_gain_value * (sample - previous_ec)) + complement_value * previous_ec;
        if (!math.isFinite(ec_value)) {
            return error.ArithmeticOverflow;
        }

        self.ema_state = ema_value;
        self.ec_state = ec_value;
        self.best_gain_state = best_gain_value;
        self.least_error_state = least_error_value;
        self.count += 1;
        self.current_value = ec_value;
        self.current_ema = ema_value;
        return .{ .ec = ec_value, .ema = ema_value };
    }

    pub fn tryUpdate(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: f64) Error!f64 {
        const pair = try self.step(sample);
        return pair.ec;
    }

    pub fn update(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: f64) f64 {
        return self.tryUpdate(sample) catch @panic("invalid zero-lag error-correcting exponential moving average sample");
    }

    pub fn updateScalar(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *ZeroLagErrorCorrectingExponentialMovingAverage, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *ZeroLagErrorCorrectingExponentialMovingAverage, stamp: i64, sample: f64) OutputArray {
        const pair = self.step(sample) catch @panic("invalid zero-lag error-correcting exponential moving average sample");
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = stamp, .value = pair.ec } });
        out.append(.{ .scalar = .{ .time = stamp, .value = pair.ema } });
        return out;
    }

    pub fn indicator(self: *ZeroLagErrorCorrectingExponentialMovingAverage) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(ZeroLagErrorCorrectingExponentialMovingAverage);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");

const ReferenceCase = struct {
    length: i64,
    gain_limit: i32,
    expected_ec: *const fn () [252]f64,
    expected_ema: *const fn () [252]f64,
};

const reference_cases = [_]ReferenceCase{
    .{ .length = 20, .gain_limit = 50, .expected_ec = &testdata.expectedL20_G50_EC, .expected_ema = &testdata.expectedL20_G50_EMA },
    .{ .length = 12, .gain_limit = 50, .expected_ec = &testdata.expectedL12_G50_EC, .expected_ema = &testdata.expectedL12_G50_EMA },
    .{ .length = 32, .gain_limit = 22, .expected_ec = &testdata.expectedL32_G22_EC, .expected_ema = &testdata.expectedL32_G22_EMA },
    .{ .length = 20, .gain_limit = 0, .expected_ec = &testdata.expectedL20_G0_EC, .expected_ema = &testdata.expectedL20_G0_EMA },
    .{ .length = 20, .gain_limit = 1, .expected_ec = &testdata.expectedL20_G1_EC, .expected_ema = &testdata.expectedL20_G1_EMA },
    .{ .length = 20, .gain_limit = 10, .expected_ec = &testdata.expectedL20_G10_EC, .expected_ema = &testdata.expectedL20_G10_EMA },
    .{ .length = 20, .gain_limit = 22, .expected_ec = &testdata.expectedL20_G22_EC, .expected_ema = &testdata.expectedL20_G22_EMA },
    .{ .length = 20, .gain_limit = 100, .expected_ec = &testdata.expectedL20_G100_EC, .expected_ema = &testdata.expectedL20_G100_EMA },
    .{ .length = 1, .gain_limit = 50, .expected_ec = &testdata.expectedL1_G50_EC, .expected_ema = &testdata.expectedL1_G50_EMA },
    .{ .length = 2, .gain_limit = 50, .expected_ec = &testdata.expectedL2_G50_EC, .expected_ema = &testdata.expectedL2_G50_EMA },
    .{ .length = 5, .gain_limit = 50, .expected_ec = &testdata.expectedL5_G50_EC, .expected_ema = &testdata.expectedL5_G50_EMA },
    .{ .length = 10, .gain_limit = 50, .expected_ec = &testdata.expectedL10_G50_EC, .expected_ema = &testdata.expectedL10_G50_EMA },
    .{ .length = 40, .gain_limit = 50, .expected_ec = &testdata.expectedL40_G50_EC, .expected_ema = &testdata.expectedL40_G50_EMA },
    .{ .length = 100, .gain_limit = 50, .expected_ec = &testdata.expectedL100_G50_EC, .expected_ema = &testdata.expectedL100_G50_EMA },
    .{ .length = 252, .gain_limit = 50, .expected_ec = &testdata.expectedL252_G50_EC, .expected_ema = &testdata.expectedL252_G50_EMA },
    .{ .length = 20, .gain_limit = 49, .expected_ec = &testdata.expectedL20_G49_EC, .expected_ema = &testdata.expectedL20_G49_EMA },
};

const BatchRow = struct {
    ec: f64,
    ema: f64,
    best_gain: f64,
    least_error: f64,
};

fn articleBatch(prices: *const [252]f64, length_value: i64, gain_limit_value: i32) [252]BatchRow {
    const alpha_value = 2.0 / (@as(f64, @floatFromInt(length_value)) + 1.0);
    var ema_series = [_]f64{0.0} ** 253;
    var ec_series = [_]f64{0.0} ** 253;
    var best_gain: f64 = 0.0;
    var outputs: [252]BatchRow = undefined;

    for (prices, 0..) |price, offset| {
        const step_index = offset + 1;
        ema_series[step_index] = alpha_value * price + (1.0 - alpha_value) * ema_series[step_index - 1];
        var least_error: f64 = 1_000_000.0;
        var integer_gain: i64 = -@as(i64, gain_limit_value);
        const limit: i64 = @as(i64, gain_limit_value);
        while (integer_gain <= limit) : (integer_gain += 1) {
            const gain = @as(f64, @floatFromInt(integer_gain)) / 10.0;
            const trial = alpha_value * (ema_series[step_index] + gain * (price - ec_series[step_index - 1])) + (1.0 - alpha_value) * ec_series[step_index - 1];
            const error_value = price - trial;
            if (@abs(error_value) < least_error) {
                least_error = @abs(error_value);
                best_gain = gain;
            }
        }
        ec_series[step_index] = alpha_value * (ema_series[step_index] + best_gain * (price - ec_series[step_index - 1])) + (1.0 - alpha_value) * ec_series[step_index - 1];
        outputs[offset] = .{
            .ec = ec_series[step_index],
            .ema = ema_series[step_index],
            .best_gain = best_gain,
            .least_error = least_error,
        };
    }
    return outputs;
}

fn expectCloseOrNan(actual: f64, expected: f64) !void {
    if (math.isNan(expected)) {
        try testing.expect(math.isNan(actual));
        return;
    }
    const scale = @max(@as(f64, 1.0), @abs(expected));
    try testing.expect(@abs(actual - expected) <= 2e-12 * scale);
}

test "zecema matches all 16 reference fixtures and article batch" {
    const prices = testdata.testInput();
    for (reference_cases) |case_item| {
        const expected_ec = case_item.expected_ec();
        const expected_ema = case_item.expected_ema();
        const reference = articleBatch(&prices, case_item.length, case_item.gain_limit);

        var stream = try ZeroLagErrorCorrectingExponentialMovingAverage.initFromLength(.{
            .length = case_item.length,
            .gain_limit = case_item.gain_limit,
        });
        var entity_stream = try ZeroLagErrorCorrectingExponentialMovingAverage.initFromLength(.{
            .length = case_item.length,
            .gain_limit = case_item.gain_limit,
        });

        for (prices, 0..) |price, bar_index| {
            const ec = stream.update(price);
            const ema_value = stream.ema();
            try expectCloseOrNan(ec, expected_ec[bar_index]);
            try expectCloseOrNan(ema_value, expected_ema[bar_index]);
            try expectCloseOrNan(ec, reference[bar_index].ec);
            try expectCloseOrNan(ema_value, reference[bar_index].ema);
            try testing.expectEqual(reference[bar_index].best_gain, stream.bestGain());
            try testing.expectEqual(reference[bar_index].least_error, stream.leastError());

            const stamp: i64 = @intCast(1_600_000_000 + bar_index * 86_400);
            const entity_out = entity_stream.updateScalar(&.{ .time = stamp, .value = price });
            try testing.expectEqual(@as(usize, 2), entity_out.len);
            try testing.expectEqual(stamp, entity_out.slice()[0].scalar.time);
            try testing.expectEqual(stamp, entity_out.slice()[1].scalar.time);
            try testing.expectEqual(ec, entity_out.slice()[0].scalar.value);
            try testing.expectEqual(ema_value, entity_out.slice()[1].scalar.value);
        }
    }
}

test "zecema zero-state startup, first-observation priming, and reset" {
    var stream = try ZeroLagErrorCorrectingExponentialMovingAverage.initDefault();
    try testing.expect(!stream.isPrimed());
    try testing.expectEqual(@as(usize, 0), stream.warmupBars());
    try testing.expectEqual(@as(usize, 0), stream.samplesSeen());
    try testing.expect(math.isNan(stream.value()));
    try testing.expect(math.isNan(stream.ema()));
    try testing.expect(math.isNan(stream.bestGain()));
    try testing.expect(math.isNan(stream.leastError()));
    try testing.expectEqual(2.0 / 21.0, stream.alpha());

    const ec_first = stream.update(100.0);
    const alpha_value: f64 = 2.0 / 21.0;
    const expected_ema = alpha_value * 100.0;
    const expected_ec = alpha_value * (expected_ema + 5.0 * 100.0);
    try testing.expect(stream.isPrimed());
    try testing.expectEqual(@as(usize, 1), stream.samplesSeen());
    try testing.expectEqual(expected_ec, ec_first);
    try testing.expectEqual(expected_ema, stream.ema());
    try testing.expectEqual(@as(f64, 5.0), stream.bestGain());
    try testing.expectEqual(@abs(100.0 - expected_ec), stream.leastError());

    stream.reset();
    try testing.expect(!stream.isPrimed());
    try testing.expectEqual(@as(usize, 0), stream.samplesSeen());
    try testing.expect(math.isNan(stream.value()));
    try testing.expect(math.isNan(stream.ema()));
    try testing.expect(math.isNan(stream.bestGain()));
    try testing.expect(math.isNan(stream.leastError()));
    try testing.expectEqual(expected_ec, stream.update(100.0));
}

test "zecema zero gain cascaded EMA, tie rule, sentinel, NaN skip, and atomic error" {
    const prices = testdata.testInput();
    var zero_gain = try ZeroLagErrorCorrectingExponentialMovingAverage.initFromLength(.{ .length = 20, .gain_limit = 0 });
    const alpha_value: f64 = 2.0 / 21.0;
    var ema_acc: f64 = 0.0;
    var cascaded_acc: f64 = 0.0;
    for (prices) |price| {
        ema_acc = alpha_value * price + (1.0 - alpha_value) * ema_acc;
        cascaded_acc = alpha_value * ema_acc + (1.0 - alpha_value) * cascaded_acc;
        try testing.expectEqual(cascaded_acc, zero_gain.update(price));
        try testing.expectEqual(ema_acc, zero_gain.ema());
        try testing.expectEqual(@as(f64, 0.0), zero_gain.bestGain());
    }

    var tie_stream = try ZeroLagErrorCorrectingExponentialMovingAverage.initDefault();
    try testing.expectEqual(@as(f64, 0.0), tie_stream.update(0.0));
    try testing.expectEqual(@as(f64, -5.0), tie_stream.bestGain());
    try testing.expectEqual(@as(f64, 0.0), tie_stream.leastError());

    var huge_stream = try ZeroLagErrorCorrectingExponentialMovingAverage.initDefault();
    const huge_price: f64 = 1e9;
    const huge_ec = huge_stream.update(huge_price);
    try testing.expectEqual(@as(f64, 0.0), huge_stream.bestGain());
    try testing.expectEqual(@as(f64, 1_000_000.0), huge_stream.leastError());
    try testing.expectEqual(alpha_value * (alpha_value * huge_price), huge_ec);

    var nan_stream = try ZeroLagErrorCorrectingExponentialMovingAverage.initDefault();
    const pre_nan = nan_stream.updateScalar(&.{ .time = 10, .value = math.nan(f64) });
    try testing.expect(math.isNan(pre_nan.slice()[0].scalar.value));
    try testing.expect(math.isNan(pre_nan.slice()[1].scalar.value));
    try testing.expect(!nan_stream.isPrimed());

    const accepted_ec = nan_stream.update(prices[0]);
    const accepted_ema = nan_stream.ema();
    const accepted_gain = nan_stream.bestGain();
    const accepted_error = nan_stream.leastError();

    try testing.expect(math.isNan(nan_stream.update(math.nan(f64))));
    try testing.expectEqual(@as(usize, 1), nan_stream.samplesSeen());
    try testing.expectEqual(accepted_ec, nan_stream.value());
    try testing.expectEqual(accepted_ema, nan_stream.ema());
    try testing.expectEqual(accepted_gain, nan_stream.bestGain());
    try testing.expectEqual(accepted_error, nan_stream.leastError());

    try testing.expectError(error.InvalidSample, nan_stream.tryUpdate(math.inf(f64)));
    try testing.expectError(error.InvalidSample, nan_stream.tryUpdate(-math.inf(f64)));
    try testing.expectError(error.ArithmeticOverflow, nan_stream.tryUpdate(1e308));
    try testing.expectEqual(@as(usize, 1), nan_stream.samplesSeen());
    try testing.expectEqual(accepted_ec, nan_stream.value());
    try testing.expectEqual(accepted_ema, nan_stream.ema());
    try testing.expectEqual(accepted_gain, nan_stream.bestGain());
    try testing.expectEqual(accepted_error, nan_stream.leastError());
}

test "zecema smoothing factor variant, metadata, entities, and validation" {
    const prices = testdata.testInput();
    var from_length = try ZeroLagErrorCorrectingExponentialMovingAverage.initFromLength(.{ .length = 20, .gain_limit = 50 });
    var from_alpha = try ZeroLagErrorCorrectingExponentialMovingAverage.initFromSmoothingFactor(.{
        .smoothing_factor = 2.0 / 21.0,
        .gain_limit = 50,
    });
    try testing.expectEqual(@as(i64, 20), from_alpha.length());
    for (prices) |price| {
        try testing.expectEqual(from_length.update(price), from_alpha.update(price));
        try testing.expectEqual(from_length.ema(), from_alpha.ema());
    }

    var unit_alpha = try ZeroLagErrorCorrectingExponentialMovingAverage.initFromSmoothingFactor(.{
        .smoothing_factor = 1.0,
        .gain_limit = 50,
    });
    try testing.expectEqual(@as(i64, 1), unit_alpha.length());
    try testing.expectEqual(@as(f64, 91.5), unit_alpha.update(91.5));
    try testing.expectEqual(math.maxInt(i64), equivalentLengthFromAlpha(1e-9));

    var meta: Metadata = undefined;
    from_length.getMetadata(&meta);
    const desc = descriptor_mod.descriptorOf(.zero_lag_error_correcting_exponential_moving_average).?;
    try testing.expectEqual(adaptivity_mod.Adaptivity.adaptive, desc.adaptivity);
    try testing.expectEqual(Identifier.zero_lag_error_correcting_exponential_moving_average, meta.identifier);
    try testing.expectEqualStrings("zecema(20, 50)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(u8, @intFromEnum(ZeroLagErrorCorrectingExponentialMovingAverageOutput.value)), meta.outputs_buf[0].kind);
    try testing.expectEqual(shape_mod.Shape.scalar, meta.outputs_buf[0].shape);
    try testing.expectEqualStrings("zecema(20, 50)", meta.outputs_buf[0].mnemonic);
    try testing.expectEqual(@as(u8, @intFromEnum(ZeroLagErrorCorrectingExponentialMovingAverageOutput.ema)), meta.outputs_buf[1].kind);
    try testing.expectEqual(shape_mod.Shape.scalar, meta.outputs_buf[1].shape);
    try testing.expectEqualStrings("zecemaEma(20, 50)", meta.outputs_buf[1].mnemonic);

    var bar_stream = try ZeroLagErrorCorrectingExponentialMovingAverage.initDefault();
    const bar_out = bar_stream.updateBar(&.{ .time = 100, .open = 100.0, .high = 101.0, .low = 99.0, .close = 100.0, .volume = 10.0 });
    try testing.expectEqual(@as(usize, 2), bar_out.len);
    try testing.expect(math.isFinite(bar_out.slice()[0].scalar.value));
    try testing.expect(math.isFinite(bar_out.slice()[1].scalar.value));

    var quote_stream = try ZeroLagErrorCorrectingExponentialMovingAverage.initDefault();
    const quote_out = quote_stream.updateQuote(&.{ .time = 100, .bid_price = 100.0, .ask_price = 100.0, .bid_size = 1.0, .ask_size = 1.0 });
    try testing.expectEqual(@as(usize, 2), quote_out.len);

    var trade_stream = try ZeroLagErrorCorrectingExponentialMovingAverage.initDefault();
    const trade_out = trade_stream.updateTrade(&.{ .time = 100, .price = 100.0, .volume = 1.0 });
    try testing.expectEqual(@as(usize, 2), trade_out.len);

    var mn_storage: [128]u8 = undefined;
    try testing.expectEqualStrings(
        "zecema(32, 22, o, a)",
        try mnemonic(&mn_storage, .{ .length = 32, .gain_limit = 22, .bar_component = .open, .quote_component = .ask }),
    );
    try testing.expectEqualStrings(
        "zecemaEma(9, 0.20000000, 22, hlc/3, v)",
        try smoothingFactorEmaMnemonic(&mn_storage, .{ .smoothing_factor = 0.2, .gain_limit = 22, .bar_component = .typical, .trade_component = .volume }),
    );

    try testing.expectError(error.InvalidLength, ZeroLagErrorCorrectingExponentialMovingAverage.initFromLength(.{ .length = 0 }));
    try testing.expectError(error.InvalidLength, ZeroLagErrorCorrectingExponentialMovingAverage.initFromLength(.{ .length = math.maxInt(i64) }));
    try testing.expectError(error.InvalidGainLimit, ZeroLagErrorCorrectingExponentialMovingAverage.initFromLength(.{ .gain_limit = -1 }));
    try testing.expectError(error.InvalidSmoothingFactor, ZeroLagErrorCorrectingExponentialMovingAverage.initFromSmoothingFactor(.{ .smoothing_factor = 0.0 }));
    try testing.expectError(error.InvalidSmoothingFactor, ZeroLagErrorCorrectingExponentialMovingAverage.initFromSmoothingFactor(.{ .smoothing_factor = -0.1 }));
    try testing.expectError(error.InvalidSmoothingFactor, ZeroLagErrorCorrectingExponentialMovingAverage.initFromSmoothingFactor(.{ .smoothing_factor = 1.1 }));
    try testing.expectError(error.InvalidSmoothingFactor, ZeroLagErrorCorrectingExponentialMovingAverage.initFromSmoothingFactor(.{ .smoothing_factor = 1e-30 }));
    try testing.expectError(error.InvalidGainLimit, ZeroLagErrorCorrectingExponentialMovingAverage.initFromSmoothingFactor(.{ .gain_limit = -1 }));
}
