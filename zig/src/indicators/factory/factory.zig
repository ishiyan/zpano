/// Factory creates indicator instances from an Identifier and a JSON parameter string.
/// This avoids the need for callers to import individual indicator modules directly.
const std = @import("std");

const entities = @import("entities");
const bar_component = entities.bar_component;
const quote_component = entities.quote_component;
const trade_component = entities.trade_component;
const json = std.json;

// -- Core types --
const indicator_mod = @import("../core/indicator.zig");
const Indicator = indicator_mod.Indicator;
const identifier_mod = @import("../core/identifier.zig");
const Identifier = identifier_mod.Identifier;

// -- Entity component types --

// -- Indicator imports (alphabetical by author/family) --

// common
const sma_mod = @import("../common/simple_moving_average/simple_moving_average.zig");
const wma_mod = @import("../common/weighted_moving_average/weighted_moving_average.zig");
const trima_mod = @import("../common/triangular_moving_average/triangular_moving_average.zig");
const ema_mod = @import("../common/exponential_moving_average/exponential_moving_average.zig");
const variance_mod = @import("../common/variance/variance.zig");
const stddev_mod = @import("../common/standard_deviation/standard_deviation.zig");
const momentum_mod = @import("../common/momentum/momentum.zig");
const roc_mod = @import("../common/rate_of_change/rate_of_change.zig");
const rocp_mod = @import("../common/rate_of_change_percent/rate_of_change_percent.zig");
const rocr_mod = @import("../common/rate_of_change_ratio/rate_of_change_ratio.zig");
const apo_mod = @import("../common/absolute_price_oscillator/absolute_price_oscillator.zig");
const pcc_mod = @import("../common/pearsons_correlation_coefficient/pearsons_correlation_coefficient.zig");
const linreg_mod = @import("../common/linear_regression/linear_regression.zig");

// custom
const goertzel_mod = @import("../custom/goertzel_spectrum/goertzel_spectrum.zig");
const maxent_mod = @import("../custom/maximum_entropy_spectrum/maximum_entropy_spectrum.zig");

// donald_lambert
const cci_mod = @import("../donald_lambert/commodity_channel_index/commodity_channel_index.zig");

// gene_quong
const mfi_mod = @import("../gene_quong/money_flow_index/money_flow_index.zig");

// george_lane
const stoch_mod = @import("../george_lane/stochastic/stochastic.zig");

// gerald_appel
const ppo_mod = @import("../gerald_appel/percentage_price_oscillator/percentage_price_oscillator.zig");
const macd_mod = @import("../gerald_appel/moving_average_convergence_divergence/moving_average_convergence_divergence.zig");

// igor_livshin
const bop_mod = @import("../igor_livshin/balance_of_power/balance_of_power.zig");

// jack_hutson
const trix_mod = @import("../jack_hutson/triple_exponential_moving_average_oscillator/triple_exponential_moving_average_oscillator.zig");

// john_bollinger
const bb_mod = @import("../john_bollinger/bollinger_bands/bollinger_bands.zig");
const bbt_mod = @import("../john_bollinger/bollinger_bands_trend/bollinger_bands_trend.zig");

// john_ehlers
const ss_mod = @import("../john_ehlers/super_smoother/super_smoother.zig");
const cog_mod = @import("../john_ehlers/center_of_gravity_oscillator/center_of_gravity_oscillator.zig");
const cc_mod = @import("../john_ehlers/cyber_cycle/cyber_cycle.zig");
const itl_mod = @import("../john_ehlers/instantaneous_trend_line/instantaneous_trend_line.zig");
const zlema_mod = @import("../john_ehlers/zero_lag_exponential_moving_average/zero_lag_exponential_moving_average.zig");
const zlecema_mod = @import("../john_ehlers/zero_lag_error_correcting_exponential_moving_average/zero_lag_error_correcting_exponential_moving_average.zig");
const rf_mod = @import("../john_ehlers/roofing_filter/roofing_filter.zig");
const mama_mod = @import("../john_ehlers/mesa_adaptive_moving_average/mesa_adaptive_moving_average.zig");
const frama_mod = @import("../john_ehlers/fractal_adaptive_moving_average/fractal_adaptive_moving_average.zig");
const dc_mod = @import("../john_ehlers/dominant_cycle/dominant_cycle.zig");
const sw_mod = @import("../john_ehlers/sinewave/sinewave.zig");
const htitl_mod = @import("../john_ehlers/hilbert_transformer_instantaneous_trend_line/hilbert_transformer_instantaneous_trend_line.zig");
const tcm_mod = @import("../john_ehlers/trend_cycle_mode/trend_cycle_mode.zig");
const cs_mod = @import("../john_ehlers/corona_spectrum/corona_spectrum.zig");
const csnr_mod = @import("../john_ehlers/corona_signal_to_noise_ratio/corona_signal_to_noise_ratio.zig");
const cswp_mod = @import("../john_ehlers/corona_swing_position/corona_swing_position.zig");
const ctv_mod = @import("../john_ehlers/corona_trend_vigor/corona_trend_vigor.zig");
const aci_mod = @import("../john_ehlers/autocorrelation_indicator/autocorrelation_indicator.zig");
const acp_mod = @import("../john_ehlers/autocorrelation_periodogram/autocorrelation_periodogram.zig");
const cbps_mod = @import("../john_ehlers/comb_band_pass_spectrum/comb_band_pass_spectrum.zig");
const dfts_mod = @import("../john_ehlers/discrete_fourier_transform_spectrum/discrete_fourier_transform_spectrum.zig");

// joseph_granville
const obv_mod = @import("../joseph_granville/on_balance_volume/on_balance_volume.zig");

// larry_williams
const wpr_mod = @import("../larry_williams/williams_percent_r/williams_percent_r.zig");
const uo_mod = @import("../larry_williams/ultimate_oscillator/ultimate_oscillator.zig");

// marc_chaikin
const ad_mod = @import("../marc_chaikin/advance_decline/advance_decline.zig");
const ado_mod = @import("../marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig");

// mark_jurik
const jma_mod = @import("../mark_jurik/jurik_moving_average/jurik_moving_average.zig");
const jrsx_mod = @import("../mark_jurik/jurik_relative_trend_strength_index/jurik_relative_trend_strength_index.zig");
const jcfb_mod = @import("../mark_jurik/jurik_composite_fractal_behavior_index/jurik_composite_fractal_behavior_index.zig");
const jvel_mod = @import("../mark_jurik/jurik_zero_lag_velocity/jurik_zero_lag_velocity.zig");
const jdmx_mod = @import("../mark_jurik/jurik_directional_movement_index/jurik_directional_movement_index.zig");
const jtpo_mod = @import("../mark_jurik/jurik_turning_point_oscillator/jurik_turning_point_oscillator.zig");
const jarsx_mod = @import("../mark_jurik/jurik_adaptive_relative_trend_strength_index/jurik_adaptive_relative_trend_strength_index.zig");
const javel_mod = @import("../mark_jurik/jurik_adaptive_zero_lag_velocity/jurik_adaptive_zero_lag_velocity.zig");
const jccx_mod = @import("../mark_jurik/jurik_commodity_channel_index/jurik_commodity_channel_index.zig");
const jvelcfb_mod = @import("../mark_jurik/jurik_fractal_adaptive_zero_lag_velocity/jurik_fractal_adaptive_zero_lag_velocity.zig");
const wav_mod = @import("../mark_jurik/jurik_wavelet_sampler/jurik_wavelet_sampler.zig");
const alma_mod = @import("../arnaud_legoux/arnaud_legoux_moving_average/arnaud_legoux_moving_average.zig");
const nma_mod = @import("../manfred_durschner/new_moving_average/new_moving_average.zig");

// patrick_mulloy
const dema_mod = @import("../patrick_mulloy/double_exponential_moving_average/double_exponential_moving_average.zig");
const tema_mod = @import("../patrick_mulloy/triple_exponential_moving_average/triple_exponential_moving_average.zig");

// perry_kaufman
const kama_mod = @import("../perry_kaufman/kaufman_adaptive_moving_average/kaufman_adaptive_moving_average.zig");

// tim_tillson
const t2_mod = @import("../tim_tillson/t2_exponential_moving_average/t2_exponential_moving_average.zig");
const t3_mod = @import("../tim_tillson/t3_exponential_moving_average/t3_exponential_moving_average.zig");

// tushar_chande
const cmo_mod = @import("../tushar_chande/chande_momentum_oscillator/chande_momentum_oscillator.zig");
const srsi_mod = @import("../tushar_chande/stochastic_relative_strength_index/stochastic_relative_strength_index.zig");
const aroon_mod = @import("../tushar_chande/aroon/aroon.zig");

// vladimir_kravchuk
const atcf_mod = @import("../vladimir_kravchuk/adaptive_trend_and_cycle_filter/adaptive_trend_and_cycle_filter.zig");

// welles_wilder
const tr_mod = @import("../welles_wilder/true_range/true_range.zig");
const atr_mod = @import("../welles_wilder/average_true_range/average_true_range.zig");
const natr_mod = @import("../welles_wilder/normalized_average_true_range/normalized_average_true_range.zig");
const dmm_mod = @import("../welles_wilder/directional_movement_minus/directional_movement_minus.zig");
const dmp_mod = @import("../welles_wilder/directional_movement_plus/directional_movement_plus.zig");
const dim_mod = @import("../welles_wilder/directional_indicator_minus/directional_indicator_minus.zig");
const dip_mod = @import("../welles_wilder/directional_indicator_plus/directional_indicator_plus.zig");
const dmx_mod = @import("../welles_wilder/directional_movement_index/directional_movement_index.zig");
const adx_mod = @import("../welles_wilder/average_directional_movement_index/average_directional_movement_index.zig");
const adxr_mod = @import("../welles_wilder/average_directional_movement_index_rating/average_directional_movement_index_rating.zig");
const rsi_mod = @import("../welles_wilder/relative_strength_index/relative_strength_index.zig");
const psar_mod = @import("../welles_wilder/parabolic_stop_and_reverse/parabolic_stop_and_reverse.zig");

pub const FactoryError = error{
    UnsupportedIndicator,
    InvalidParams,
    OutOfMemory,
    IndicatorInitFailed,
};

/// Result of a factory creation: an Indicator interface plus a cleanup mechanism.
pub const FactoryResult = struct {
    indicator: Indicator,
    /// Opaque pointer to the heap-allocated indicator (used for deinit).
    ctx: *anyopaque,
    /// Type-erased deinit function. Call as `result.deinit_fn(result.ctx, allocator)`.
    deinit_fn: *const fn (*anyopaque, std.mem.Allocator) void,

    /// Convenience method to free the indicator.
    pub fn deinit(self: FactoryResult, allocator: std.mem.Allocator) void {
        self.deinit_fn(self.ctx, allocator);
    }
};

// ── JSON helpers ────────────────────────────────────────────────────────────

const ObjectMap = json.ObjectMap;

fn getInt(obj: ObjectMap, key: []const u8, default: i32) i32 {
    const val = obj.get(key) orelse return default;
    return switch (val) {
        .integer => |i| @intCast(i),
        .float => |f| @intFromFloat(f),
        else => default,
    };
}

fn getUsize(obj: ObjectMap, key: []const u8, default: usize) usize {
    const val = obj.get(key) orelse return default;
    return switch (val) {
        .integer => |i| if (i >= 0) @intCast(i) else default,
        .float => |f| if (f >= 0) @intFromFloat(f) else default,
        else => default,
    };
}

fn getF64(obj: ObjectMap, key: []const u8, default: f64) f64 {
    const val = obj.get(key) orelse return default;
    return switch (val) {
        .float => |f| f,
        .integer => |i| @floatFromInt(i),
        else => default,
    };
}

fn getBool(obj: ObjectMap, key: []const u8, default: bool) bool {
    const val = obj.get(key) orelse return default;
    return switch (val) {
        .bool => |b| b,
        else => default,
    };
}

fn getU8(obj: ObjectMap, key: []const u8, default: u8) u8 {
    const val = obj.get(key) orelse return default;
    return switch (val) {
        .integer => |i| if (i >= 0 and i <= 255) @intCast(i) else default,
        .float => |f| if (f >= 0 and f <= 255) @intFromFloat(f) else default,
        else => default,
    };
}

fn hasKey(obj: ObjectMap, key: []const u8) bool {
    return obj.get(key) != null;
}

fn getBarComponent(obj: ObjectMap) ?bar_component.BarComponent {
    const val = obj.get("barComponent") orelse return null;
    const i: u8 = switch (val) {
        .integer => |n| if (n >= 0 and n <= 8) @intCast(n) else return null,
        .float => |f| if (f >= 0 and f <= 8) @intFromFloat(f) else return null,
        else => return null,
    };
    return @enumFromInt(i);
}

fn getQuoteComponent(obj: ObjectMap) ?quote_component.QuoteComponent {
    const val = obj.get("quoteComponent") orelse return null;
    const i: u8 = switch (val) {
        .integer => |n| if (n >= 0 and n <= 7) @intCast(n) else return null,
        .float => |f| if (f >= 0 and f <= 7) @intFromFloat(f) else return null,
        else => return null,
    };
    return @enumFromInt(i);
}

fn getTradeComponent(obj: ObjectMap) ?trade_component.TradeComponent {
    const val = obj.get("tradeComponent") orelse return null;
    const i: u8 = switch (val) {
        .integer => |n| if (n >= 0 and n <= 1) @intCast(n) else return null,
        .float => |f| if (f >= 0 and f <= 1) @intFromFloat(f) else return null,
        else => return null,
    };
    return @enumFromInt(i);
}

/// Parse a JSON string into an ObjectMap. Returns an empty map for empty/null input.
fn parseParams(allocator: std.mem.Allocator, params: []const u8) !json.Parsed(json.Value) {
    const input = if (params.len == 0) "{}" else params;
    return json.parseFromSlice(json.Value, allocator, input, .{});
}

fn getObject(parsed: json.Parsed(json.Value)) ?ObjectMap {
    return switch (parsed.value) {
        .object => |obj| obj,
        else => null,
    };
}

// ── Heap allocation helper ──────────────────────────────────────────────────

/// Allocates a copy of `val` on the heap and returns a pointer.
/// After copying, calls fixSlices() if available to fix up self-referential
/// slice pointers that were invalidated by the move from stack to heap.
fn heapAlloc(comptime T: type, allocator: std.mem.Allocator, val: T) !*T {
    const ptr = try allocator.create(T);
    ptr.* = val;
    if (@hasDecl(T, "fixSlices")) ptr.fixSlices();
    return ptr;
}

/// Creates a type-erased deinit function for a heap-allocated indicator.
fn DeinitFn(comptime T: type) *const fn (*anyopaque, std.mem.Allocator) void {
    return &struct {
        fn deinit(ctx: *anyopaque, alloc: std.mem.Allocator) void {
            const ptr: *T = @ptrCast(@alignCast(ctx));
            if (@hasDecl(T, "deinit")) {
                ptr.deinit();
            }
            alloc.destroy(ptr);
        }
    }.deinit;
}

// ── Public API ──────────────────────────────────────────────────────────────

/// Creates an indicator from its identifier and a JSON-encoded parameter string.
/// The caller must call `result.deinit_fn(allocator)` when done to free the indicator.
pub fn create(allocator: std.mem.Allocator, id: Identifier, params_json: []const u8) FactoryError!FactoryResult {
    const parsed = parseParams(allocator, params_json) catch return FactoryError.InvalidParams;
    defer parsed.deinit();

    const obj = getObject(parsed) orelse return FactoryError.InvalidParams;

    return switch (id) {
        // ── common ──────────────────────────────────────────────────────

        .simple_moving_average => createWithAllocParams(sma_mod.SimpleMovingAverage, sma_mod.SimpleMovingAverageParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .weighted_moving_average => createWithAllocParams(wma_mod.WeightedMovingAverage, wma_mod.WeightedMovingAverageParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .triangular_moving_average => createWithAllocParams(trima_mod.TriangularMovingAverage, trima_mod.TriangularMovingAverageParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .exponential_moving_average => blk: {
            if (hasKey(obj, "smoothingFactor")) {
                break :blk createWithParams(ema_mod.ExponentialMovingAverage, allocator, ema_mod.ExponentialMovingAverage.initSmoothingFactor(.{
                    .smoothing_factor = getF64(obj, "smoothingFactor", 0.1),
                    .first_is_average = getBool(obj, "firstIsAverage", false),
                    .bar_component = getBarComponent(obj),
                    .quote_component = getQuoteComponent(obj),
                    .trade_component = getTradeComponent(obj),
                }));
            }
            break :blk createWithParams(ema_mod.ExponentialMovingAverage, allocator, ema_mod.ExponentialMovingAverage.initLength(.{
                .length = getUsize(obj, "length", 14),
                .first_is_average = getBool(obj, "firstIsAverage", false),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }));
        },

        .variance => createWithAllocParams(variance_mod.Variance, variance_mod.VarianceParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .is_unbiased = getBool(obj, "isUnbiased", true),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .standard_deviation => createWithAllocParams(stddev_mod.StandardDeviation, stddev_mod.StandardDeviationParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .is_unbiased = getBool(obj, "isUnbiased", true),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .momentum => createWithAllocParams(momentum_mod.Momentum, momentum_mod.MomentumParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .rate_of_change => createWithAllocParams(roc_mod.RateOfChange, roc_mod.RateOfChangeParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .rate_of_change_percent => createWithAllocParams(rocp_mod.RateOfChangePercent, rocp_mod.RateOfChangePercentParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .rate_of_change_ratio => createWithAllocParams(rocr_mod.RateOfChangeRatio, rocr_mod.RateOfChangeRatioParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .hundred_scale = getBool(obj, "hundredScale", false),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .absolute_price_oscillator => createWithAllocParams(apo_mod.AbsolutePriceOscillator, apo_mod.AbsolutePriceOscillatorParams, allocator, obj, .{
            .fast_length = getUsize(obj, "fastLength", 12),
            .slow_length = getUsize(obj, "slowLength", 26),
            .first_is_average = getBool(obj, "firstIsAverage", false),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .pearsons_correlation_coefficient => createWithAllocParams(pcc_mod.PearsonsCorrelationCoefficient, pcc_mod.PearsonsCorrelationCoefficientParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .linear_regression => createWithAllocParams(linreg_mod.LinearRegression, linreg_mod.LinearRegressionParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── custom ──────────────────────────────────────────────────────

        .goertzel_spectrum => blk: {
            const p = goertzel_mod.Params{
                .length = getInt(obj, "length", 0),
                .min_period = getF64(obj, "minPeriod", 0),
                .max_period = getF64(obj, "maxPeriod", 0),
                .spectrum_resolution = getInt(obj, "spectrumResolution", 0),
                .is_first_order = getBool(obj, "isFirstOrder", false),
                .disable_spectral_dilation_compensation = getBool(obj, "disableSpectralDilationCompensation", false),
                .disable_automatic_gain_control = getBool(obj, "disableAutomaticGainControl", false),
                .automatic_gain_control_decay_factor = getF64(obj, "automaticGainControlDecayFactor", 0),
                .fixed_normalization = getBool(obj, "fixedNormalization", false),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = goertzel_mod.GoertzelSpectrum.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(goertzel_mod.GoertzelSpectrum, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(goertzel_mod.GoertzelSpectrum) };
        },

        .maximum_entropy_spectrum => blk: {
            const p = maxent_mod.Params{
                .length = getInt(obj, "length", 0),
                .degree = getInt(obj, "degree", 0),
                .min_period = getF64(obj, "minPeriod", 0),
                .max_period = getF64(obj, "maxPeriod", 0),
                .spectrum_resolution = getInt(obj, "spectrumResolution", 0),
                .disable_automatic_gain_control = getBool(obj, "disableAutomaticGainControl", false),
                .automatic_gain_control_decay_factor = getF64(obj, "automaticGainControlDecayFactor", 0),
                .fixed_normalization = getBool(obj, "fixedNormalization", false),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = maxent_mod.MaximumEntropySpectrum.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(maxent_mod.MaximumEntropySpectrum, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(maxent_mod.MaximumEntropySpectrum) };
        },

        // ── donald_lambert ──────────────────────────────────────────────

        .commodity_channel_index => createWithAllocParams(cci_mod.CommodityChannelIndex, cci_mod.CommodityChannelIndexParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .inverse_scaling_factor = getF64(obj, "inverseScalingFactor", 0),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── gene_quong ──────────────────────────────────────────────────

        .money_flow_index => createWithAllocParams(mfi_mod.MoneyFlowIndex, mfi_mod.MoneyFlowIndexParams, allocator, obj, .{
            .length = @as(u32, @intCast(getUsize(obj, "length", 14))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── george_lane ─────────────────────────────────────────────────

        .stochastic => createWithAllocParams(stoch_mod.Stochastic, stoch_mod.StochasticParams, allocator, obj, .{
            .fast_k_length = getUsize(obj, "fastKLength", 5),
            .slow_k_length = getUsize(obj, "slowKLength", 3),
            .slow_d_length = getUsize(obj, "slowDLength", 3),
            .first_is_average = getBool(obj, "firstIsAverage", false),
        }),

        // ── gerald_appel ────────────────────────────────────────────────

        .percentage_price_oscillator => createWithAllocParams(ppo_mod.PercentagePriceOscillator, ppo_mod.PercentagePriceOscillatorParams, allocator, obj, .{
            .fast_length = getUsize(obj, "fastLength", 12),
            .slow_length = getUsize(obj, "slowLength", 26),
            .first_is_average = getBool(obj, "firstIsAverage", false),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .moving_average_convergence_divergence => createWithAllocParams(macd_mod.MovingAverageConvergenceDivergence, macd_mod.MacdParams, allocator, obj, .{
            .fast_length = getUsize(obj, "fastLength", 12),
            .slow_length = getUsize(obj, "slowLength", 26),
            .signal_length = getUsize(obj, "signalLength", 9),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── igor_livshin ────────────────────────────────────────────────

        .balance_of_power => blk: {
            const ind = bop_mod.BalanceOfPower.init();
            const ptr = heapAlloc(bop_mod.BalanceOfPower, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(bop_mod.BalanceOfPower) };
        },

        // ── jack_hutson ─────────────────────────────────────────────────

        .triple_exponential_moving_average_oscillator => createWithParams(trix_mod.TripleExponentialMovingAverageOscillator, allocator, trix_mod.TripleExponentialMovingAverageOscillator.init(.{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        // ── john_bollinger ──────────────────────────────────────────────

        .bollinger_bands => createWithAllocParams(bb_mod.BollingerBands, bb_mod.BollingerBandsParams, allocator, obj, .{
            .length = getUsize(obj, "length", 20),
            .upper_multiplier = getF64(obj, "upperMultiplier", 2.0),
            .lower_multiplier = getF64(obj, "lowerMultiplier", 2.0),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .bollinger_bands_trend => createWithAllocParams(bbt_mod.BollingerBandsTrend, bbt_mod.BollingerBandsTrendParams, allocator, obj, .{
            .fast_length = getUsize(obj, "fastLength", 20),
            .slow_length = getUsize(obj, "slowLength", 50),
            .upper_multiplier = getF64(obj, "upperMultiplier", 2.0),
            .lower_multiplier = getF64(obj, "lowerMultiplier", 2.0),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── john_ehlers ─────────────────────────────────────────────────

        .super_smoother => createWithParams(ss_mod.SuperSmoother, allocator, ss_mod.SuperSmoother.init(.{
            .shortest_cycle_period = @as(i32, @intCast(getUsize(obj, "shortestCyclePeriod", 10))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .center_of_gravity_oscillator => blk: {
            const p = cog_mod.Params{
                .length = getInt(obj, "length", 10),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = cog_mod.CenterOfGravityOscillator.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(cog_mod.CenterOfGravityOscillator, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(cog_mod.CenterOfGravityOscillator) };
        },

        .cyber_cycle => blk: {
            if (hasKey(obj, "smoothingFactor")) {
                const ind = cc_mod.CyberCycle.initSmoothingFactor(.{
                    .smoothing_factor = getF64(obj, "smoothingFactor", 0.07),
                    .signal_lag = getInt(obj, "signalLag", 9),
                    .bar_component = getBarComponent(obj),
                    .quote_component = getQuoteComponent(obj),
                    .trade_component = getTradeComponent(obj),
                }) catch return FactoryError.IndicatorInitFailed;
                const ptr = heapAlloc(cc_mod.CyberCycle, allocator, ind) catch return FactoryError.OutOfMemory;
                break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(cc_mod.CyberCycle) };
            }
            const ind = cc_mod.CyberCycle.initLength(.{
                .length = getInt(obj, "length", 28),
                .signal_lag = getInt(obj, "signalLag", 9),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(cc_mod.CyberCycle, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(cc_mod.CyberCycle) };
        },

        .instantaneous_trend_line => blk: {
            if (hasKey(obj, "smoothingFactor")) {
                const ind = itl_mod.InstantaneousTrendLine.initSmoothingFactor(.{
                    .smoothing_factor = getF64(obj, "smoothingFactor", 0.07),
                    .bar_component = getBarComponent(obj),
                    .quote_component = getQuoteComponent(obj),
                    .trade_component = getTradeComponent(obj),
                }) catch return FactoryError.IndicatorInitFailed;
                const ptr = heapAlloc(itl_mod.InstantaneousTrendLine, allocator, ind) catch return FactoryError.OutOfMemory;
                break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(itl_mod.InstantaneousTrendLine) };
            }
            const ind = itl_mod.InstantaneousTrendLine.initLength(.{
                .length = getInt(obj, "length", 14),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(itl_mod.InstantaneousTrendLine, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(itl_mod.InstantaneousTrendLine) };
        },

        .zero_lag_exponential_moving_average => createWithAllocParams(zlema_mod.ZeroLagExponentialMovingAverage, zlema_mod.ZeroLagExponentialMovingAverageParams, allocator, obj, .{
            .smoothing_factor = getF64(obj, "smoothingFactor", 0.25),
            .velocity_gain_factor = getF64(obj, "velocityGainFactor", 0.5),
            .velocity_momentum_length = getInt(obj, "velocityMomentumLength", 3),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .zero_lag_error_correcting_exponential_moving_average => createWithParams(zlecema_mod.ZeroLagErrorCorrectingExponentialMovingAverage, allocator, zlecema_mod.ZeroLagErrorCorrectingExponentialMovingAverage.init(.{
            .smoothing_factor = getF64(obj, "smoothingFactor", 0.095),
            .gain_limit = getF64(obj, "gainLimit", 5),
            .gain_step = getF64(obj, "gainStep", 0.1),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .roofing_filter => createWithParams(rf_mod.RoofingFilter, allocator, rf_mod.RoofingFilter.init(.{
            .shortest_cycle_period = getInt(obj, "shortestCyclePeriod", 10),
            .longest_cycle_period = getInt(obj, "longestCyclePeriod", 48),
            .has_two_pole_highpass_filter = getBool(obj, "hasTwoPoleHighpassFilter", false),
            .has_zero_mean = getBool(obj, "hasZeroMean", false),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .mesa_adaptive_moving_average => blk: {
            if (hasKey(obj, "fastLimitSmoothingFactor") or hasKey(obj, "slowLimitSmoothingFactor")) {
                const ind = mama_mod.MesaAdaptiveMovingAverage.initSmoothingFactor(.{
                    .fast_limit_smoothing_factor = getF64(obj, "fastLimitSmoothingFactor", 0.5),
                    .slow_limit_smoothing_factor = getF64(obj, "slowLimitSmoothingFactor", 0.05),
                    .bar_component = getBarComponent(obj),
                    .quote_component = getQuoteComponent(obj),
                    .trade_component = getTradeComponent(obj),
                }) catch return FactoryError.IndicatorInitFailed;
                const ptr = heapAlloc(mama_mod.MesaAdaptiveMovingAverage, allocator, ind) catch return FactoryError.OutOfMemory;
                break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(mama_mod.MesaAdaptiveMovingAverage) };
            }
            if (obj.count() == 0) {
                const ind = mama_mod.MesaAdaptiveMovingAverage.initDefault() catch return FactoryError.IndicatorInitFailed;
                const ptr = heapAlloc(mama_mod.MesaAdaptiveMovingAverage, allocator, ind) catch return FactoryError.OutOfMemory;
                break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(mama_mod.MesaAdaptiveMovingAverage) };
            }
            const ind = mama_mod.MesaAdaptiveMovingAverage.initLength(.{
                .fast_limit_length = getInt(obj, "fastLimitLength", 3),
                .slow_limit_length = getInt(obj, "slowLimitLength", 39),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(mama_mod.MesaAdaptiveMovingAverage, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(mama_mod.MesaAdaptiveMovingAverage) };
        },

        .fractal_adaptive_moving_average => blk: {
            const p = frama_mod.Params{
                .length = getInt(obj, "length", 16),
                .slowest_smoothing_factor = getF64(obj, "slowestSmoothingFactor", 0.01),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = frama_mod.FractalAdaptiveMovingAverage.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(frama_mod.FractalAdaptiveMovingAverage, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(frama_mod.FractalAdaptiveMovingAverage) };
        },

        .dominant_cycle => blk: {
            if (obj.count() == 0) {
                const ind = dc_mod.DominantCycle.initDefault() catch return FactoryError.IndicatorInitFailed;
                const ptr = heapAlloc(dc_mod.DominantCycle, allocator, ind) catch return FactoryError.OutOfMemory;
                break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(dc_mod.DominantCycle) };
            }
            const ind = dc_mod.DominantCycle.init(.{
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(dc_mod.DominantCycle, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(dc_mod.DominantCycle) };
        },

        .sine_wave => blk: {
            if (obj.count() == 0) {
                const ind = sw_mod.SineWave.initDefault() catch return FactoryError.IndicatorInitFailed;
                const ptr = heapAlloc(sw_mod.SineWave, allocator, ind) catch return FactoryError.OutOfMemory;
                break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(sw_mod.SineWave) };
            }
            const ind = sw_mod.SineWave.init(.{
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(sw_mod.SineWave, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(sw_mod.SineWave) };
        },

        .hilbert_transformer_instantaneous_trend_line => blk: {
            if (obj.count() == 0) {
                const ind = htitl_mod.HilbertTransformerInstantaneousTrendLine.initDefault() catch return FactoryError.IndicatorInitFailed;
                const ptr = heapAlloc(htitl_mod.HilbertTransformerInstantaneousTrendLine, allocator, ind) catch return FactoryError.OutOfMemory;
                break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(htitl_mod.HilbertTransformerInstantaneousTrendLine) };
            }
            const ind = htitl_mod.HilbertTransformerInstantaneousTrendLine.init(.{
                .trend_line_smoothing_length = getU8(obj, "trendLineSmoothingLength", 4),
                .cycle_part_multiplier = getF64(obj, "cyclePartMultiplier", 1.0),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(htitl_mod.HilbertTransformerInstantaneousTrendLine, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(htitl_mod.HilbertTransformerInstantaneousTrendLine) };
        },

        .trend_cycle_mode => blk: {
            if (obj.count() == 0) {
                const ind = tcm_mod.TrendCycleMode.initDefault() catch return FactoryError.IndicatorInitFailed;
                const ptr = heapAlloc(tcm_mod.TrendCycleMode, allocator, ind) catch return FactoryError.OutOfMemory;
                break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(tcm_mod.TrendCycleMode) };
            }
            const ind = tcm_mod.TrendCycleMode.init(.{
                .trend_line_smoothing_length = getU8(obj, "trendLineSmoothingLength", 4),
                .cycle_part_multiplier = getF64(obj, "cyclePartMultiplier", 1.0),
                .separation_percentage = getF64(obj, "separationPercentage", 1.5),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(tcm_mod.TrendCycleMode, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(tcm_mod.TrendCycleMode) };
        },

        .corona_spectrum => blk: {
            const p = cs_mod.Params{
                .min_raster_value = getF64(obj, "minRasterValue", 0),
                .max_raster_value = getF64(obj, "maxRasterValue", 0),
                .min_parameter_value = getF64(obj, "minParameterValue", 0),
                .max_parameter_value = getF64(obj, "maxParameterValue", 0),
                .high_pass_filter_cutoff = getInt(obj, "highPassFilterCutoff", 0),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = cs_mod.CoronaSpectrum.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(cs_mod.CoronaSpectrum, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(cs_mod.CoronaSpectrum) };
        },

        .corona_signal_to_noise_ratio => blk: {
            const p = csnr_mod.Params{
                .raster_length = getInt(obj, "rasterLength", 0),
                .max_raster_value = getF64(obj, "maxRasterValue", 0),
                .min_parameter_value = getF64(obj, "minParameterValue", 0),
                .max_parameter_value = getF64(obj, "maxParameterValue", 0),
                .high_pass_filter_cutoff = getInt(obj, "highPassFilterCutoff", 0),
                .minimal_period = getInt(obj, "minimalPeriod", 0),
                .maximal_period = getInt(obj, "maximalPeriod", 0),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = csnr_mod.CoronaSignalToNoiseRatio.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(csnr_mod.CoronaSignalToNoiseRatio, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(csnr_mod.CoronaSignalToNoiseRatio) };
        },

        .corona_swing_position => blk: {
            const p = cswp_mod.Params{
                .raster_length = getInt(obj, "rasterLength", 0),
                .max_raster_value = getF64(obj, "maxRasterValue", 0),
                .min_parameter_value = getF64(obj, "minParameterValue", 0),
                .max_parameter_value = getF64(obj, "maxParameterValue", 0),
                .high_pass_filter_cutoff = getInt(obj, "highPassFilterCutoff", 0),
                .minimal_period = getInt(obj, "minimalPeriod", 0),
                .maximal_period = getInt(obj, "maximalPeriod", 0),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = cswp_mod.CoronaSwingPosition.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(cswp_mod.CoronaSwingPosition, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(cswp_mod.CoronaSwingPosition) };
        },

        .corona_trend_vigor => blk: {
            const p = ctv_mod.Params{
                .raster_length = getInt(obj, "rasterLength", 0),
                .max_raster_value = getF64(obj, "maxRasterValue", 0),
                .min_parameter_value = getF64(obj, "minParameterValue", 0),
                .max_parameter_value = getF64(obj, "maxParameterValue", 0),
                .high_pass_filter_cutoff = getInt(obj, "highPassFilterCutoff", 0),
                .minimal_period = getInt(obj, "minimalPeriod", 0),
                .maximal_period = getInt(obj, "maximalPeriod", 0),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = ctv_mod.CoronaTrendVigor.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(ctv_mod.CoronaTrendVigor, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(ctv_mod.CoronaTrendVigor) };
        },

        .auto_correlation_indicator => blk: {
            const p = aci_mod.Params{
                .min_lag = getInt(obj, "minLag", 3),
                .max_lag = getInt(obj, "maxLag", 48),
                .smoothing_period = getInt(obj, "smoothingPeriod", 10),
                .averaging_length = getInt(obj, "averagingLength", 0),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = aci_mod.AutoCorrelationIndicator.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(aci_mod.AutoCorrelationIndicator, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(aci_mod.AutoCorrelationIndicator) };
        },

        .auto_correlation_periodogram => blk: {
            const p = acp_mod.Params{
                .min_period = getInt(obj, "minPeriod", 10),
                .max_period = getInt(obj, "maxPeriod", 48),
                .averaging_length = getInt(obj, "averagingLength", 3),
                .disable_spectral_squaring = getBool(obj, "disableSpectralSquaring", false),
                .disable_smoothing = getBool(obj, "disableSmoothing", false),
                .disable_automatic_gain_control = getBool(obj, "disableAutomaticGainControl", false),
                .automatic_gain_control_decay_factor = getF64(obj, "automaticGainControlDecayFactor", 0.995),
                .fixed_normalization = getBool(obj, "fixedNormalization", false),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = acp_mod.AutoCorrelationPeriodogram.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(acp_mod.AutoCorrelationPeriodogram, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(acp_mod.AutoCorrelationPeriodogram) };
        },

        .comb_band_pass_spectrum => blk: {
            const p = cbps_mod.Params{
                .min_period = getInt(obj, "minPeriod", 10),
                .max_period = getInt(obj, "maxPeriod", 48),
                .bandwidth = getF64(obj, "bandwidth", 0.3),
                .disable_spectral_dilation_compensation = getBool(obj, "disableSpectralDilationCompensation", false),
                .disable_automatic_gain_control = getBool(obj, "disableAutomaticGainControl", false),
                .automatic_gain_control_decay_factor = getF64(obj, "automaticGainControlDecayFactor", 0.995),
                .fixed_normalization = getBool(obj, "fixedNormalization", false),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = cbps_mod.CombBandPassSpectrum.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(cbps_mod.CombBandPassSpectrum, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(cbps_mod.CombBandPassSpectrum) };
        },

        .discrete_fourier_transform_spectrum => blk: {
            const p = dfts_mod.Params{
                .length = getInt(obj, "length", 48),
                .min_period = getF64(obj, "minPeriod", 10.0),
                .max_period = getF64(obj, "maxPeriod", 48.0),
                .spectrum_resolution = getInt(obj, "spectrumResolution", 1),
                .disable_spectral_dilation_compensation = getBool(obj, "disableSpectralDilationCompensation", false),
                .disable_automatic_gain_control = getBool(obj, "disableAutomaticGainControl", false),
                .automatic_gain_control_decay_factor = getF64(obj, "automaticGainControlDecayFactor", 0.995),
                .fixed_normalization = getBool(obj, "fixedNormalization", false),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = dfts_mod.DiscreteFourierTransformSpectrum.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(dfts_mod.DiscreteFourierTransformSpectrum, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(dfts_mod.DiscreteFourierTransformSpectrum) };
        },

        // ── joseph_granville ────────────────────────────────────────────

        .on_balance_volume => createWithParams(obv_mod.OnBalanceVolume, allocator, obv_mod.OnBalanceVolume.init(.{})),

        // ── larry_williams ──────────────────────────────────────────────

        .williams_percent_r => createWithAllocParams(wpr_mod.WilliamsPercentR, wpr_mod.WilliamsPercentRParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
        }),

        .ultimate_oscillator => createWithAllocParams(uo_mod.UltimateOscillator, uo_mod.UltimateOscillatorParams, allocator, obj, .{
            .length1 = getUsize(obj, "length1", 7),
            .length2 = getUsize(obj, "length2", 14),
            .length3 = getUsize(obj, "length3", 28),
        }),

        // ── marc_chaikin ────────────────────────────────────────────────

        .advance_decline => blk: {
            const ind = ad_mod.AdvanceDecline.init();
            const ptr = heapAlloc(ad_mod.AdvanceDecline, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(ad_mod.AdvanceDecline) };
        },

        .advance_decline_oscillator => createWithAllocParams(ado_mod.AdvanceDeclineOscillator, ado_mod.AdvanceDeclineOscillatorParams, allocator, obj, .{
            .fast_length = @as(u32, @intCast(getUsize(obj, "fastLength", 3))),
            .slow_length = @as(u32, @intCast(getUsize(obj, "slowLength", 10))),
        }),

        // ── mark_jurik ──────────────────────────────────────────────────

        .jurik_moving_average => createWithParams(jma_mod.JurikMovingAverage, allocator, jma_mod.JurikMovingAverage.init(.{
            .length = @as(u32, @intCast(getUsize(obj, "length", 14))),
            .phase = getInt(obj, "phase", 0),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .jurik_relative_trend_strength_index => createWithParams(jrsx_mod.JurikRelativeTrendStrengthIndex, allocator, jrsx_mod.JurikRelativeTrendStrengthIndex.init(.{
            .length = @as(u32, @intCast(getUsize(obj, "length", 14))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .jurik_composite_fractal_behavior_index => createWithParams(jcfb_mod.JurikCompositeFractalBehaviorIndex, allocator, jcfb_mod.JurikCompositeFractalBehaviorIndex.init(.{
            .fractal_type = @as(u32, @intCast(getUsize(obj, "fractalType", 1))),
            .smooth = @as(u32, @intCast(getUsize(obj, "smooth", 10))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .jurik_zero_lag_velocity => createWithParams(jvel_mod.JurikZeroLagVelocity, allocator, jvel_mod.JurikZeroLagVelocity.init(.{
            .depth = @as(u32, @intCast(getUsize(obj, "depth", 10))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .jurik_directional_movement_index => createWithParams(jdmx_mod.JurikDirectionalMovementIndex, allocator, jdmx_mod.JurikDirectionalMovementIndex.init(.{
            .length = @as(u32, @intCast(getUsize(obj, "length", 14))),
        })),

        .jurik_turning_point_oscillator => createWithParams(jtpo_mod.JurikTurningPointOscillator, allocator, jtpo_mod.JurikTurningPointOscillator.init(.{
            .length = @as(u32, @intCast(getUsize(obj, "length", 14))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .jurik_commodity_channel_index => createWithParams(jccx_mod.JurikCommodityChannelIndex, allocator, jccx_mod.JurikCommodityChannelIndex.init(.{
            .length = @as(u32, @intCast(getUsize(obj, "length", 20))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .jurik_wavelet_sampler => createWithParams(wav_mod.JurikWaveletSampler, allocator, wav_mod.JurikWaveletSampler.init(.{
            .index = @as(u32, @intCast(getUsize(obj, "index", 12))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .jurik_adaptive_zero_lag_velocity => createWithParams(javel_mod.JurikAdaptiveZeroLagVelocity, allocator, javel_mod.JurikAdaptiveZeroLagVelocity.init(.{
            .lo_length = @as(u32, @intCast(getUsize(obj, "loLength", 5))),
            .hi_length = @as(u32, @intCast(getUsize(obj, "hiLength", 30))),
            .sensitivity = getF64(obj, "sensitivity", 1.0),
            .period = getF64(obj, "period", 3.0),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .jurik_fractal_adaptive_zero_lag_velocity => createWithParams(jvelcfb_mod.JurikFractalAdaptiveZeroLagVelocity, allocator, jvelcfb_mod.JurikFractalAdaptiveZeroLagVelocity.init(.{
            .lo_depth = @as(u32, @intCast(getUsize(obj, "loDepth", 5))),
            .hi_depth = @as(u32, @intCast(getUsize(obj, "hiDepth", 30))),
            .fractal_type = @as(u32, @intCast(getUsize(obj, "fractalType", 1))),
            .smooth = @as(u32, @intCast(getUsize(obj, "smooth", 10))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .jurik_adaptive_relative_trend_strength_index => createWithParams(jarsx_mod.JurikAdaptiveRelativeTrendStrengthIndex, allocator, jarsx_mod.JurikAdaptiveRelativeTrendStrengthIndex.init(.{
            .lo_length = @as(u32, @intCast(getUsize(obj, "loLength", 5))),
            .hi_length = @as(u32, @intCast(getUsize(obj, "hiLength", 30))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        // ── arnaud_legoux ───────────────────────────────────────────────

        .arnaud_legoux_moving_average => createWithAllocParams(alma_mod.ArnaudLegouxMovingAverage, alma_mod.ArnaudLegouxMovingAverageParams, allocator, obj, .{
            .window = getUsize(obj, "window", 9),
            .sigma = getF64(obj, "sigma", 6.0),
            .offset = getF64(obj, "offset", 0.85),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── manfred_durschner ───────────────────────────────────────────

        .new_moving_average => createWithAllocParams(nma_mod.NewMovingAverage, nma_mod.NewMovingAverageParams, allocator, obj, .{
            .primary_period = getUsize(obj, "primaryPeriod", 0),
            .secondary_period = getUsize(obj, "secondaryPeriod", 8),
            .ma_type = @enumFromInt(getUsize(obj, "maType", 3)),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── patrick_mulloy ──────────────────────────────────────────────

        .double_exponential_moving_average => blk: {
            if (hasKey(obj, "smoothingFactor")) {
                break :blk createWithParams(dema_mod.DoubleExponentialMovingAverage, allocator, dema_mod.DoubleExponentialMovingAverage.initSmoothingFactor(.{
                    .smoothing_factor = getF64(obj, "smoothingFactor", 0.1),
                    .bar_component = getBarComponent(obj),
                    .quote_component = getQuoteComponent(obj),
                    .trade_component = getTradeComponent(obj),
                }));
            }
            break :blk createWithParams(dema_mod.DoubleExponentialMovingAverage, allocator, dema_mod.DoubleExponentialMovingAverage.initLength(.{
                .length = getUsize(obj, "length", 14),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }));
        },

        .triple_exponential_moving_average => blk: {
            if (hasKey(obj, "smoothingFactor")) {
                break :blk createWithParams(tema_mod.TripleExponentialMovingAverage, allocator, tema_mod.TripleExponentialMovingAverage.initSmoothingFactor(.{
                    .smoothing_factor = getF64(obj, "smoothingFactor", 0.1),
                    .bar_component = getBarComponent(obj),
                    .quote_component = getQuoteComponent(obj),
                    .trade_component = getTradeComponent(obj),
                }));
            }
            break :blk createWithParams(tema_mod.TripleExponentialMovingAverage, allocator, tema_mod.TripleExponentialMovingAverage.initLength(.{
                .length = getUsize(obj, "length", 14),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }));
        },

        // ── perry_kaufman ───────────────────────────────────────────────

        .kaufman_adaptive_moving_average => blk: {
            if (hasKey(obj, "fastestSmoothingFactor") or hasKey(obj, "slowestSmoothingFactor")) {
                const ind = kama_mod.KaufmanAdaptiveMovingAverage.initSmoothingFactor(allocator, .{
                    .efficiency_ratio_length = @as(u32, @intCast(getUsize(obj, "efficiencyRatioLength", 10))),
                    .fastest_smoothing_factor = getF64(obj, "fastestSmoothingFactor", 2.0 / 3.0),
                    .slowest_smoothing_factor = getF64(obj, "slowestSmoothingFactor", 2.0 / 31.0),
                    .bar_component = getBarComponent(obj),
                    .quote_component = getQuoteComponent(obj),
                    .trade_component = getTradeComponent(obj),
                }) catch return FactoryError.IndicatorInitFailed;
                const ptr = heapAlloc(kama_mod.KaufmanAdaptiveMovingAverage, allocator, ind) catch return FactoryError.OutOfMemory;
                break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(kama_mod.KaufmanAdaptiveMovingAverage) };
            }
            const ind = kama_mod.KaufmanAdaptiveMovingAverage.initLength(allocator, .{
                .efficiency_ratio_length = @as(u32, @intCast(getUsize(obj, "efficiencyRatioLength", 10))),
                .fastest_length = @as(u32, @intCast(getUsize(obj, "fastestLength", 2))),
                .slowest_length = @as(u32, @intCast(getUsize(obj, "slowestLength", 30))),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(kama_mod.KaufmanAdaptiveMovingAverage, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(kama_mod.KaufmanAdaptiveMovingAverage) };
        },

        // ── tim_tillson ─────────────────────────────────────────────────

        .t2_exponential_moving_average => blk: {
            if (hasKey(obj, "smoothingFactor")) {
                break :blk createWithParams(t2_mod.T2ExponentialMovingAverage, allocator, t2_mod.T2ExponentialMovingAverage.initSmoothingFactor(.{
                    .smoothing_factor = getF64(obj, "smoothingFactor", 0.1),
                    .volume_factor = getF64(obj, "volumeFactor", 0.7),
                    .bar_component = getBarComponent(obj),
                    .quote_component = getQuoteComponent(obj),
                    .trade_component = getTradeComponent(obj),
                }));
            }
            break :blk createWithParams(t2_mod.T2ExponentialMovingAverage, allocator, t2_mod.T2ExponentialMovingAverage.initLength(.{
                .length = getUsize(obj, "length", 14),
                .volume_factor = getF64(obj, "volumeFactor", 0.7),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }));
        },

        .t3_exponential_moving_average => blk: {
            if (hasKey(obj, "smoothingFactor")) {
                break :blk createWithParams(t3_mod.T3ExponentialMovingAverage, allocator, t3_mod.T3ExponentialMovingAverage.initSmoothingFactor(.{
                    .smoothing_factor = getF64(obj, "smoothingFactor", 0.1),
                    .volume_factor = getF64(obj, "volumeFactor", 0.7),
                    .bar_component = getBarComponent(obj),
                    .quote_component = getQuoteComponent(obj),
                    .trade_component = getTradeComponent(obj),
                }));
            }
            break :blk createWithParams(t3_mod.T3ExponentialMovingAverage, allocator, t3_mod.T3ExponentialMovingAverage.initLength(.{
                .length = getUsize(obj, "length", 14),
                .volume_factor = getF64(obj, "volumeFactor", 0.7),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            }));
        },

        // ── tushar_chande ───────────────────────────────────────────────

        .chande_momentum_oscillator => createWithAllocParams(cmo_mod.ChandeMomentumOscillator, cmo_mod.ChandeMomentumOscillatorParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .stochastic_relative_strength_index => createWithAllocParams(srsi_mod.StochasticRelativeStrengthIndex, srsi_mod.StochasticRelativeStrengthIndexParams, allocator, obj, .{
            .length = getUsize(obj, "rsiLength", 14),
            .fast_k_length = getUsize(obj, "fastKLength", 5),
            .fast_d_length = getUsize(obj, "fastDLength", 3),
            .first_is_average = getBool(obj, "firstIsAverage", false),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .aroon => createWithAllocParams(aroon_mod.Aroon, aroon_mod.AroonParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
        }),

        // ── vladimir_kravchuk ───────────────────────────────────────────

        .adaptive_trend_and_cycle_filter => blk: {
            const ind = atcf_mod.AdaptiveTrendAndCycleFilter.init(allocator) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(atcf_mod.AdaptiveTrendAndCycleFilter, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(atcf_mod.AdaptiveTrendAndCycleFilter) };
        },

        // ── welles_wilder ───────────────────────────────────────────────

        .true_range => blk: {
            const ind = tr_mod.TrueRange.init();
            const ptr = heapAlloc(tr_mod.TrueRange, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(tr_mod.TrueRange) };
        },

        .average_true_range => blk: {
            const ind = atr_mod.AverageTrueRange.init(allocator, .{
                .length = getInt(obj, "length", 14),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(atr_mod.AverageTrueRange, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(atr_mod.AverageTrueRange) };
        },

        .normalized_average_true_range => blk: {
            const ind = natr_mod.NormalizedAverageTrueRange.init(allocator, .{
                .length = getInt(obj, "length", 14),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(natr_mod.NormalizedAverageTrueRange, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(natr_mod.NormalizedAverageTrueRange) };
        },

        .directional_movement_minus => createWithParams(dmm_mod.DirectionalMovementMinus, allocator, dmm_mod.DirectionalMovementMinus.init(.{
            .length = getUsize(obj, "length", 14),
        })),

        .directional_movement_plus => createWithParams(dmp_mod.DirectionalMovementPlus, allocator, dmp_mod.DirectionalMovementPlus.init(.{
            .length = getUsize(obj, "length", 14),
        })),

        .directional_indicator_minus => blk: {
            const ind = dim_mod.DirectionalIndicatorMinus.init(allocator, .{
                .length = getInt(obj, "length", 14),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(dim_mod.DirectionalIndicatorMinus, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(dim_mod.DirectionalIndicatorMinus) };
        },

        .directional_indicator_plus => blk: {
            const ind = dip_mod.DirectionalIndicatorPlus.init(allocator, .{
                .length = getInt(obj, "length", 14),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(dip_mod.DirectionalIndicatorPlus, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(dip_mod.DirectionalIndicatorPlus) };
        },

        .directional_movement_index => blk: {
            const ind = dmx_mod.DirectionalMovementIndex.init(allocator, .{
                .length = getInt(obj, "length", 14),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(dmx_mod.DirectionalMovementIndex, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(dmx_mod.DirectionalMovementIndex) };
        },

        .average_directional_movement_index => blk: {
            const ind = adx_mod.AverageDirectionalMovementIndex.init(allocator, .{
                .length = getInt(obj, "length", 14),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(adx_mod.AverageDirectionalMovementIndex, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(adx_mod.AverageDirectionalMovementIndex) };
        },

        .average_directional_movement_index_rating => blk: {
            const ind = adxr_mod.AverageDirectionalMovementIndexRating.init(allocator, .{
                .length = getInt(obj, "length", 14),
            }) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(adxr_mod.AverageDirectionalMovementIndexRating, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(adxr_mod.AverageDirectionalMovementIndexRating) };
        },

        .relative_strength_index => createWithParams(rsi_mod.RelativeStrengthIndex, allocator, rsi_mod.RelativeStrengthIndex.init(.{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        .parabolic_stop_and_reverse => createWithParams(psar_mod.ParabolicStopAndReverse, allocator, psar_mod.ParabolicStopAndReverse.init(.{
            .start_value = getF64(obj, "startValue", 0),
            .offset_on_reverse = getF64(obj, "offsetOnReverse", 0),
            .acceleration_init_long = getF64(obj, "accelerationInitLong", 0),
            .acceleration_long = getF64(obj, "accelerationLong", 0),
            .acceleration_max_long = getF64(obj, "accelerationMaxLong", 0),
            .acceleration_init_short = getF64(obj, "accelerationInitShort", 0),
            .acceleration_short = getF64(obj, "accelerationShort", 0),
            .acceleration_max_short = getF64(obj, "accelerationMaxShort", 0),
        })),
    };
}

// ── Generic construction helpers ────────────────────────────────────────────

/// For indicators with `init(allocator, params)` pattern (Pattern A with named params).
fn createWithAllocParams(
    comptime T: type,
    comptime P: type,
    allocator: std.mem.Allocator,
    _: ObjectMap,
    params: P,
) FactoryError!FactoryResult {
    const ind = T.init(allocator, params) catch return FactoryError.IndicatorInitFailed;
    const ptr = heapAlloc(T, allocator, ind) catch return FactoryError.OutOfMemory;
    return .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(T) };
}

/// For indicators with `init(params)` or `initLength(params)` pattern (no allocator).
/// Takes the result of the init call directly (which may be error union).
fn createWithParams(
    comptime T: type,
    allocator: std.mem.Allocator,
    init_result: anytype,
) FactoryError!FactoryResult {
    const ind = switch (@typeInfo(@TypeOf(init_result))) {
        .error_union => init_result catch return FactoryError.IndicatorInitFailed,
        else => init_result,
    };
    const ptr = heapAlloc(T, allocator, ind) catch return FactoryError.OutOfMemory;
    return .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(T) };
}

// ── Tests ───────────────────────────────────────────────────────────────────

const metadata_mod = @import("../core/metadata.zig");

test "create simple moving average with default params" {
    const allocator = std.testing.allocator;
    const result = try create(allocator, .simple_moving_average, "{}");
    defer result.deinit(allocator);
    var md: metadata_mod.Metadata = undefined;
    result.indicator.metadata(&md);
    try std.testing.expect(md.outputs_len > 0);
}

test "create simple moving average with length" {
    const allocator = std.testing.allocator;
    const result = try create(allocator, .simple_moving_average,
        \\{"length": 20}
    );
    defer result.deinit(allocator);
    var md: metadata_mod.Metadata = undefined;
    result.indicator.metadata(&md);
    try std.testing.expect(md.outputs_len > 0);
}

test "create exponential moving average with smoothing factor" {
    const allocator = std.testing.allocator;
    const result = try create(allocator, .exponential_moving_average,
        \\{"smoothingFactor": 0.1}
    );
    defer result.deinit(allocator);
    var md: metadata_mod.Metadata = undefined;
    result.indicator.metadata(&md);
    try std.testing.expect(md.outputs_len > 0);
}

test "create true range no params" {
    const allocator = std.testing.allocator;
    const result = try create(allocator, .true_range, "");
    defer result.deinit(allocator);
    var md: metadata_mod.Metadata = undefined;
    result.indicator.metadata(&md);
    try std.testing.expect(md.outputs_len > 0);
}
