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

// arnaud_legoux
const alma_mod = @import("../arnaud_legoux/arnaud_legoux_moving_average/arnaud_legoux_moving_average.zig");

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

// manfred_durschner
const nma_mod = @import("../manfred_durschner/new_moving_average/new_moving_average.zig");

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

// custom
const goertzel_mod = @import("../custom/goertzel_spectrum/goertzel_spectrum.zig");
const maxent_mod = @import("../custom/maximum_entropy_spectrum/maximum_entropy_spectrum.zig");

// jean_philippe_poton
const fdi_mod = @import("../jean_philippe_poton/fractal_dimension_index/fractal_dimension_index.zig");
const fgdi_mod = @import("../jean_philippe_poton/fractal_graph_dimension_index/fractal_graph_dimension_index.zig");
const frasma_mod = @import("../jean_philippe_poton/fractal_adaptive_simple_moving_average/fractal_adaptive_simple_moving_average.zig");
const frasma2_mod = @import("../jean_philippe_poton/fractal_adaptive_simple_moving_average_2/fractal_adaptive_simple_moving_average_2.zig");
const rsfrasma_mod = @import("../jean_philippe_poton/rescaled_fractal_adaptive_simple_moving_average/rescaled_fractal_adaptive_simple_moving_average.zig");
const fban_mod = @import("../jean_philippe_poton/fractal_bands/fractal_bands.zig");
const fbanha_mod = @import("../jean_philippe_poton/fractal_bands_hybride_adaptive/fractal_bands_hybride_adaptive.zig");
const fctban_mod = @import("../jean_philippe_poton/fractional_bands/fractional_bands.zig");
const hurdif_mod = @import("../jean_philippe_poton/hurst_difference/hurst_difference.zig");

// doug_schaff
const stc_mod = @import("../doug_schaff/schaff_trend_cycle/schaff_trend_cycle.zig");

// don_mak
const aema_mod = @import("../don_mak/adaptive_exponential_moving_average/adaptive_exponential_moving_average.zig");
const iswp_mod = @import("../don_mak/instantaneous_sine_wave_period/instantaneous_sine_wave_period.zig");
const pfd_mod = @import("../don_mak/polynomial_fit_derivative/polynomial_fit_derivative.zig");
const mhw_mod = @import("../don_mak/mexican_hat_wavelet/mexican_hat_wavelet.zig");
const swb_mod = @import("../don_mak/sinc_wavelet_bandpass/sinc_wavelet_bandpass.zig");
const mema_mod = @import("../don_mak/modified_exponential_moving_average/modified_exponential_moving_average.zig");
const vcema_mod = @import("../don_mak/velocity_corrected_exponential_moving_average/velocity_corrected_exponential_moving_average.zig");
const pof_mod = @import("../don_mak/polynomial_forecast/polynomial_forecast.zig");
const pvtx_mod = @import("../don_mak/parabolic_vertex/parabolic_vertex.zig");
const cvtx_mod = @import("../don_mak/cubic_vertex/cubic_vertex.zig");
const qpl_mod = @import("../raymond_lee/quantum_price_levels/quantum_price_levels.zig");
const mmm_mod = @import("../zurab_silagadze/moving_mini_max/moving_mini_max.zig");
const tsi_mod = @import("../william_blau/true_strength_index/true_strength_index.zig");
const ergodic_mod = @import("../william_blau/ergodic_oscillator/ergodic_oscillator.zig");
const mdi_mod = @import("../william_blau/mean_deviation_index/mean_deviation_index.zig");
const macdi_mod = @import("../william_blau/macd_index/macd_index.zig");
const dm_mod = @import("../william_blau/double_smoothed_momenta/double_smoothed_momenta.zig");
const cmi_mod = @import("../william_blau/candlestick_momentum_index/candlestick_momentum_index.zig");
const csi_mod = @import("../william_blau/candlestick_strength_index/candlestick_strength_index.zig");
const smi_mod = @import("../william_blau/stochastic_momentum_index/stochastic_momentum_index.zig");
const dss_mod = @import("../william_blau/double_smoothed_stochastic/double_smoothed_stochastic.zig");
const dti_mod = @import("../william_blau/directional_trend_index/directional_trend_index.zig");
const tvi_mod = @import("../william_blau/tick_volume_indicator/tick_volume_indicator.zig");
const sdtsi_mod = @import("../william_blau/slope_divergence_tsi_filter/slope_divergence_tsi_filter.zig");
const atf_mod = @import("../william_blau/adx_type_filter/adx_type_filter.zig");
const ntf_mod = @import("../william_blau/nonambiguous_trend_filter/nonambiguous_trend_filter.zig");

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

        // ── arnaud legoux ────────────────────────────────────────────────

        .arnaud_legoux_moving_average => createWithAllocParams(alma_mod.ArnaudLegouxMovingAverage, alma_mod.ArnaudLegouxMovingAverageParams, allocator, obj, .{
            .window = getUsize(obj, "window", 9),
            .sigma = getF64(obj, "sigma", 6.0),
            .offset = getF64(obj, "offset", 0.85),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── donald lambert ───────────────────────────────────────────────

        .commodity_channel_index => createWithAllocParams(cci_mod.CommodityChannelIndex, cci_mod.CommodityChannelIndexParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
            .inverse_scaling_factor = getF64(obj, "inverseScalingFactor", 0),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── gene quong ───────────────────────────────────────────────────

        .money_flow_index => createWithAllocParams(mfi_mod.MoneyFlowIndex, mfi_mod.MoneyFlowIndexParams, allocator, obj, .{
            .length = @as(u32, @intCast(getUsize(obj, "length", 14))),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── george lane ──────────────────────────────────────────────────

        .stochastic => createWithAllocParams(stoch_mod.Stochastic, stoch_mod.StochasticParams, allocator, obj, .{
            .fast_k_length = getUsize(obj, "fastKLength", 5),
            .slow_k_length = getUsize(obj, "slowKLength", 3),
            .slow_d_length = getUsize(obj, "slowDLength", 3),
            .first_is_average = getBool(obj, "firstIsAverage", false),
        }),

        // ── gerald appel ─────────────────────────────────────────────────

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

        // ── igor livshin ─────────────────────────────────────────────────

        .balance_of_power => blk: {
            const ind = bop_mod.BalanceOfPower.init();
            const ptr = heapAlloc(bop_mod.BalanceOfPower, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(bop_mod.BalanceOfPower) };
        },

        // ── jack hutson ──────────────────────────────────────────────────

        .triple_exponential_moving_average_oscillator => createWithParams(trix_mod.TripleExponentialMovingAverageOscillator, allocator, trix_mod.TripleExponentialMovingAverageOscillator.init(.{
            .length = getUsize(obj, "length", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),

        // ── john bollinger ───────────────────────────────────────────────

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

        // ── john ehlers ──────────────────────────────────────────────────

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

        // ── joseph granville ─────────────────────────────────────────────

        .on_balance_volume => createWithParams(obv_mod.OnBalanceVolume, allocator, obv_mod.OnBalanceVolume.init(.{})),

        // ── larry williams ───────────────────────────────────────────────

        .williams_percent_r => createWithAllocParams(wpr_mod.WilliamsPercentR, wpr_mod.WilliamsPercentRParams, allocator, obj, .{
            .length = getUsize(obj, "length", 14),
        }),

        .ultimate_oscillator => createWithAllocParams(uo_mod.UltimateOscillator, uo_mod.UltimateOscillatorParams, allocator, obj, .{
            .length1 = getUsize(obj, "length1", 7),
            .length2 = getUsize(obj, "length2", 14),
            .length3 = getUsize(obj, "length3", 28),
        }),

        // ── manfred durschner ────────────────────────────────────────────

        // ── lars von thienen ──────────────────────────────────────────────────
        .bartels_periodicity_score => createWithAllocParams(lvt_bps.BartelsPeriodicityScore,lvt_bps.BartelsPeriodicityScoreParams,allocator,obj,.{.period=getInt(obj,"period",20),.segments=getInt(obj,"segments",5),.smoothing=getF64(obj,"smoothing",0.0),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .complete_cycle_scanner => createWithAllocParams(lvt_cscan.CompleteCycleScanner,lvt_cscan.CompleteCycleScannerParams,allocator,obj,.{.window=getInt(obj,"window",200),.min_period=getInt(obj,"min_period",5),.max_period=getInt(obj,"max_period",80),.smoothing=getF64(obj,"smoothing",1600.0),.genuine_threshold=getF64(obj,"genuine_threshold",49.0),.phase_cycles=getF64(obj,"phase_cycles",2.0),.top_k=getInt(obj,"top_k",3),.peak_distance=getInt(obj,"peak_distance",2),.max_segments=getInt(obj,"max_segments",5),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .composite_cycle => createWithAllocParams(lvt_ccyc.CompositeCycle,lvt_ccyc.CompositeCycleParams,allocator,obj,.{.window=getInt(obj,"window",100),.periods=@constCast(&[_]f64{20,40}),.smoothing=getF64(obj,"smoothing",1600.0),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .cycle_phase_score => createWithAllocParams(lvt_cphase.CyclePhaseScore,lvt_cphase.CyclePhaseScoreParams,allocator,obj,.{.window=getInt(obj,"window",200),.min_period=getInt(obj,"min_period",5),.max_period=getInt(obj,"max_period",80),.smoothing=getF64(obj,"smoothing",1600.0),.genuine_threshold=getF64(obj,"genuine_threshold",49.0),.phase_cycles=getF64(obj,"phase_cycles",2.0),.top_k=getInt(obj,"top_k",3),.peak_distance=getInt(obj,"peak_distance",2),.max_segments=getInt(obj,"max_segments",5),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .cycle_polarity => createWithAllocParams(lvt_cpol.CyclePolarity,lvt_cpol.CyclePolarityParams,allocator,obj,.{.window=getInt(obj,"window",100),.short_period=getF64(obj,"short_period",20.0),.long_period=getF64(obj,"long_period",40.0),.smoothing=getF64(obj,"smoothing",1600.0),.tolerance=getF64(obj,"tolerance",0.0),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .cycle_strength => createWithAllocParams(lvt_cstr.CycleStrength,lvt_cstr.CycleStrengthParams,allocator,obj,.{.window=getInt(obj,"window",100),.period=getF64(obj,"period",20.0),.smoothing=getF64(obj,"smoothing",1600.0),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .cycle_swing_momentum => createWithAllocParams(lvt_csmo.CycleSwingMomentum,lvt_csmo.CycleSwingMomentumParams,allocator,obj,.{.window=getInt(obj,"window",50),.fast_smoothing=getF64(obj,"fast_smoothing",160.0),.slow_smoothing=getF64(obj,"slow_smoothing",1600.0),.band_memory=getInt(obj,"band_memory",34),.leveling=getF64(obj,"leveling",10.0),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .cyclic_smoothed_relative_strength_index => createWithAllocParams(lvt_crsi.CyclicSmoothedRelativeStrengthIndex,lvt_crsi.CyclicSmoothedRelativeStrengthIndexParams,allocator,obj,.{.dominant_cycle=getInt(obj,"dominant_cycle",20),.vibration=getInt(obj,"vibration",10),.leveling=getF64(obj,"leveling",10.0),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .detrended_rhythm_oscillators => createWithAllocParams(lvt_dro.DetrendedRhythmOscillators,lvt_dro.DetrendedRhythmOscillatorsParams,allocator,obj,.{.fast_length=getInt(obj,"fast_length",3),.slow_length=getInt(obj,"slow_length",50),.zig_period=getInt(obj,"zig_period",30),.cycle_memory=getInt(obj,"cycle_memory",15),.detrend=getBool(obj,"detrend",true),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .dominant_market_vibration => createWithAllocParams(lvt_dmv.DominantMarketVibration,lvt_dmv.DominantMarketVibrationParams,allocator,obj,.{.window=getInt(obj,"window",200),.min_period=getInt(obj,"min_period",5),.max_period=getInt(obj,"max_period",80),.smoothing=getF64(obj,"smoothing",1600.0),.genuine_threshold=getF64(obj,"genuine_threshold",49.0),.phase_cycles=getF64(obj,"phase_cycles",2.0),.top_k=getInt(obj,"top_k",3),.peak_distance=getInt(obj,"peak_distance",2),.max_segments=getInt(obj,"max_segments",5),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .dynamic_cycle_explorer => createWithAllocParams(lvt_dce.DynamicCycleExplorer,lvt_dce.DynamicCycleExplorerParams,allocator,obj,.{.window=getInt(obj,"window",200),.min_period=getInt(obj,"min_period",5),.max_period=getInt(obj,"max_period",80),.smoothing=getF64(obj,"smoothing",1600.0),.genuine_threshold=getF64(obj,"genuine_threshold",49.0),.phase_cycles=getF64(obj,"phase_cycles",2.0),.top_k=getInt(obj,"top_k",3),.peak_distance=getInt(obj,"peak_distance",2),.max_segments=getInt(obj,"max_segments",5),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .goertzel_cycle_spectrum => createWithAllocParams(lvt_gcs.GoertzelCycleSpectrum,lvt_gcs.GoertzelCycleSpectrumParams,allocator,obj,.{.window=getInt(obj,"window",128),.min_period=getInt(obj,"min_period",5),.max_period=getInt(obj,"max_period",50),.step=getInt(obj,"step",1),.smoothing=getF64(obj,"smoothing",1600.0),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .hodrick_prescott_smoothing_and_detrending => createWithAllocParams(lvt_hpsd.HodrickPrescottSmoothingAndDetrending,lvt_hpsd.HodrickPrescottSmoothingAndDetrendingParams,allocator,obj,.{.window=getInt(obj,"window",50),.smoothing=getF64(obj,"smoothing",1600.0),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .metonic_relative_strength_index => createWithAllocParams(lvt_mrsi.MetonicRelativeStrengthIndex,lvt_mrsi.MetonicRelativeStrengthIndexParams,allocator,obj,.{.length=getInt(obj,"length",13),.delay1=getInt(obj,"delay1",6939),.delay2=getInt(obj,"delay2",4016),.delay3=getInt(obj,"delay3",2922),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .metonic_sentiment_index => createWithAllocParams(lvt_msent.MetonicSentimentIndex,lvt_msent.MetonicSentimentIndexParams,allocator,obj,.{.length=getInt(obj,"length",13),.delay1=getInt(obj,"delay1",6939),.delay2=getInt(obj,"delay2",4016),.delay3=getInt(obj,"delay3",2922),.mode=getInt(obj,"mode",0)}),
        .metonic_williams_percent_r => createWithAllocParams(lvt_mwillr.MetonicWilliamsPercentR,lvt_mwillr.MetonicWilliamsPercentRParams,allocator,obj,.{.length=getInt(obj,"length",13),.delay1=getInt(obj,"delay1",6939),.delay2=getInt(obj,"delay2",4016),.delay3=getInt(obj,"delay3",2922)}),
        .phasing_oscillator => createWithAllocParams(lvt_phosc.PhasingOscillator,lvt_phosc.PhasingOscillatorParams,allocator,obj,.{.window=getInt(obj,"window",200),.short_min=getInt(obj,"short_min",5),.short_max=getInt(obj,"short_max",30),.long_min=getInt(obj,"long_min",31),.long_max=getInt(obj,"long_max",80),.smoothing=getF64(obj,"smoothing",1600.0),.genuine_threshold=getF64(obj,"genuine_threshold",49.0),.phase_cycles=getF64(obj,"phase_cycles",2.0),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),
        .lars_von_thienen_zero_lag_exponential_moving_average => createWithAllocParams(lvt_lzlema.LarsVonThienenZeroLagExponentialMovingAverage,lvt_lzlema.LarsVonThienenZeroLagExponentialMovingAverageParams,allocator,obj,.{.length=getInt(obj,"length",13),.bar_component=getBarComponent(obj),.quote_component=getQuoteComponent(obj),.trade_component=getTradeComponent(obj)}),

        .new_moving_average => createWithAllocParams(nma_mod.NewMovingAverage, nma_mod.NewMovingAverageParams, allocator, obj, .{
            .primary_period = getUsize(obj, "primaryPeriod", 0),
            .secondary_period = getUsize(obj, "secondaryPeriod", 8),
            .ma_type = @enumFromInt(getUsize(obj, "maType", 3)),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── marc chaikin ─────────────────────────────────────────────────

        .advance_decline => blk: {
            const ind = ad_mod.AdvanceDecline.init();
            const ptr = heapAlloc(ad_mod.AdvanceDecline, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(ad_mod.AdvanceDecline) };
        },

        .advance_decline_oscillator => createWithAllocParams(ado_mod.AdvanceDeclineOscillator, ado_mod.AdvanceDeclineOscillatorParams, allocator, obj, .{
            .fast_length = @as(u32, @intCast(getUsize(obj, "fastLength", 3))),
            .slow_length = @as(u32, @intCast(getUsize(obj, "slowLength", 10))),
        }),

        // ── mark jurik ───────────────────────────────────────────────────

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

        // ── patrick mulloy ───────────────────────────────────────────────

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

        // ── perry kaufman ────────────────────────────────────────────────

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

        // ── tim tillson ──────────────────────────────────────────────────

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

        // ── tushar chande ────────────────────────────────────────────────

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

        // ── vladimir kravchuk ─────────────────────────────────────────────

        .adaptive_trend_and_cycle_filter => blk: {
            const ind = atcf_mod.AdaptiveTrendAndCycleFilter.init(allocator) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(atcf_mod.AdaptiveTrendAndCycleFilter, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(atcf_mod.AdaptiveTrendAndCycleFilter) };
        },

        // ── welles wilder ────────────────────────────────────────────────

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

        .fractal_dimension_index => blk: {
            const p = fdi_mod.FractalDimensionIndexParams{
                .period = @as(usize, @intCast(if (getInt(obj, "period", 0) > 0) getInt(obj, "period", 0) else 30)),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = fdi_mod.FractalDimensionIndex.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(fdi_mod.FractalDimensionIndex, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(fdi_mod.FractalDimensionIndex) };
        },

        .fractal_graph_dimension_index => blk: {
            const p = fgdi_mod.FractalGraphDimensionIndexParams{
                .period = @as(usize, @intCast(if (getInt(obj, "period", 0) > 0) getInt(obj, "period", 0) else 30)),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = fgdi_mod.FractalGraphDimensionIndex.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(fgdi_mod.FractalGraphDimensionIndex, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(fgdi_mod.FractalGraphDimensionIndex) };
        },

        .fractal_adaptive_simple_moving_average => blk: {
            const p = frasma_mod.FractalAdaptiveSimpleMovingAverageParams{
                .period = @as(usize, @intCast(if (getInt(obj, "period", 0) > 0) getInt(obj, "period", 0) else 30)),
                .normal_speed = @as(usize, @intCast(if (getInt(obj, "normalSpeed", 0) > 0) getInt(obj, "normalSpeed", 0) else 20)),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = frasma_mod.FractalAdaptiveSimpleMovingAverage.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(frasma_mod.FractalAdaptiveSimpleMovingAverage, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(frasma_mod.FractalAdaptiveSimpleMovingAverage) };
        },
        .fractal_adaptive_simple_moving_average_2 => blk: {
            const p = frasma2_mod.FractalAdaptiveSimpleMovingAverage2Params{
                .period = @as(usize, @intCast(if (getInt(obj, "period", 0) > 0) getInt(obj, "period", 0) else 30)),
                .normal_speed = @as(usize, @intCast(if (getInt(obj, "normalSpeed", 0) > 0) getInt(obj, "normalSpeed", 0) else 20)),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = frasma2_mod.FractalAdaptiveSimpleMovingAverage2.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(frasma2_mod.FractalAdaptiveSimpleMovingAverage2, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(frasma2_mod.FractalAdaptiveSimpleMovingAverage2) };
        },
        .rescaled_fractal_adaptive_simple_moving_average => blk: {
            const p = rsfrasma_mod.RescaledFractalAdaptiveSimpleMovingAverageParams{
                .period = @as(usize, @intCast(if (getInt(obj, "period", 0) > 0) getInt(obj, "period", 0) else 64)),
                .normal_speed = @as(usize, @intCast(if (getInt(obj, "normal_speed", 0) > 0) getInt(obj, "normal_speed", 0) else 30)),
                .price_scale = getF64(obj, "price_scale", 1.0),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = rsfrasma_mod.RescaledFractalAdaptiveSimpleMovingAverage.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(rsfrasma_mod.RescaledFractalAdaptiveSimpleMovingAverage, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(rsfrasma_mod.RescaledFractalAdaptiveSimpleMovingAverage) };
        },
        .fractal_bands => blk: {
            const p = fban_mod.FractalBandsParams{
                .period = @as(usize, @intCast(if (getInt(obj, "period", 0) > 0) getInt(obj, "period", 0) else 30)),
                .normal_speed = @as(usize, @intCast(if (getInt(obj, "normalSpeed", 0) > 0) getInt(obj, "normalSpeed", 0) else 20)),
                .alpha = getF64(obj, "alpha", 2.0),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = fban_mod.FractalBands.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(fban_mod.FractalBands, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(fban_mod.FractalBands) };
        },
        .fractal_bands_hybride_adaptive => blk: {
            const p = fbanha_mod.FractalBandsHybrideAdaptiveParams{
                .period = @as(usize, @intCast(if (getInt(obj, "period", 0) > 0) getInt(obj, "period", 0) else 30)),
                .normal_speed_fallback = @as(usize, @intCast(if (getInt(obj, "normalSpeedFallback", 0) > 0) getInt(obj, "normalSpeedFallback", 0) else 30)),
                .alpha = getF64(obj, "alpha", 2.0),
                .nyquist = getF64(obj, "nyquist", 0.5),
                .alpha_hp = getF64(obj, "alphaHP", 0.07),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = fbanha_mod.FractalBandsHybrideAdaptive.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(fbanha_mod.FractalBandsHybrideAdaptive, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(fbanha_mod.FractalBandsHybrideAdaptive) };
        },
        .fractional_bands => blk: {
            const p = fctban_mod.FractionalBandsParams{
                .period = @as(usize, @intCast(if (getInt(obj, "period", 0) > 0) getInt(obj, "period", 0) else 30)),
                .price_scale = getF64(obj, "priceScale", 1.0),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = fctban_mod.FractionalBands.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(fctban_mod.FractionalBands, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(fctban_mod.FractionalBands) };
        },
        .hurst_difference => blk: {
            const p = hurdif_mod.HurstDifferenceParams{
                .period = @as(usize, @intCast(if (getInt(obj, "period", 0) > 0) getInt(obj, "period", 0) else 30)),
                .bar_component = getBarComponent(obj),
                .quote_component = getQuoteComponent(obj),
                .trade_component = getTradeComponent(obj),
            };
            const ind = hurdif_mod.HurstDifference.init(allocator, p) catch return FactoryError.IndicatorInitFailed;
            const ptr = heapAlloc(hurdif_mod.HurstDifference, allocator, ind) catch return FactoryError.OutOfMemory;
            break :blk .{ .indicator = ptr.indicator(), .ctx = ptr, .deinit_fn = DeinitFn(hurdif_mod.HurstDifference) };
        },

        // ── doug schaff ───────────────────────────────────────────────────
        .schaff_trend_cycle => createWithAllocParams(stc_mod.SchaffTrendCycle, stc_mod.SchaffTrendCycleParams, allocator, obj, .{
            .fast = getUsize(obj, "fast", 23),
            .slow = getUsize(obj, "slow", 50),
            .tclen = getUsize(obj, "tclen", 10),
            .factor = getF64(obj, "factor", 0.5),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── don mak ────────────────────────────────────────────────────────
        .adaptive_exponential_moving_average => createWithAllocParams(aema_mod.AdaptiveExponentialMovingAverage, aema_mod.AdaptiveExponentialMovingAverageParams, allocator, obj, .{
            .alpha_max = getF64(obj, "alphaMax", 0.5),
            .alpha_min = getF64(obj, "alphaMin", 0.05),
            .omega0 = getF64(obj, "omega0", 1.0),
            .smoothing = @as(i64, getInt(obj, "smoothing", 3)),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .instantaneous_sine_wave_period => createWithAllocParams(iswp_mod.InstantaneousSineWavePeriod, iswp_mod.InstantaneousSineWavePeriodParams, allocator, obj, .{
            .smoothing = @as(i64, getInt(obj, "smoothing", 0)),
            .min_period = getF64(obj, "minPeriod", 4.0),
            .max_period = getF64(obj, "maxPeriod", 50.0),
            .error_threshold = getF64(obj, "errorThreshold", 20.0),
            .dx = getF64(obj, "dx", 0.01),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .polynomial_fit_derivative => createWithAllocParams(pfd_mod.PolynomialFitDerivative, pfd_mod.PolynomialFitDerivativeParams, allocator, obj, .{
            .degree = getUsize(obj, "degree", 3),
            .order = getUsize(obj, "order", 1),
            .smoothing = @as(i64, getInt(obj, "smoothing", 6)),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .mexican_hat_wavelet => createWithAllocParams(mhw_mod.MexicanHatWavelet, mhw_mod.MexicanHatWaveletParams, allocator, obj, .{
            .band = band_blk: {
                const bi = getInt(obj, "band", 1);
                break :band_blk if (bi >= 0 and bi <= 3) @as(mhw_mod.Band, @enumFromInt(@as(u8, @intCast(bi)))) else mhw_mod.Band.mid;
            },
            .dilation = getF64(obj, "dilation", 0.0),
            .period = getF64(obj, "period", 0.0),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .sinc_wavelet_bandpass => createWithAllocParams(swb_mod.SincWaveletBandpass, swb_mod.SincWaveletBandpassParams, allocator, obj, .{
            .band = swb_band_blk: {
                const bi = getInt(obj, "band", 1);
                break :swb_band_blk if (bi >= 0 and bi <= 3) @as(swb_mod.Band, @enumFromInt(@as(u8, @intCast(bi)))) else swb_mod.Band.mid;
            },
            .velocity = getBool(obj, "velocity", false),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        .modified_exponential_moving_average => createWithAllocParams(mema_mod.ModifiedExponentialMovingAverage, mema_mod.ModifiedExponentialMovingAverageParams, allocator, obj, .{
            .period = getUsize(obj, "period", 6),
            .degree = getUsize(obj, "degree", 3),
            .skip = getUsize(obj, "skip", 1),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),
        .velocity_corrected_exponential_moving_average => createWithAllocParams(vcema_mod.VelocityCorrectedExponentialMovingAverage, vcema_mod.VelocityCorrectedExponentialMovingAverageParams, allocator, obj, .{
            .period = getUsize(obj, "period", 6),
            .degree = getUsize(obj, "degree", 3),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),
        .polynomial_forecast => createWithAllocParams(pof_mod.PolynomialForecast, pof_mod.PolynomialForecastParams, allocator, obj, .{
            .degree = getUsize(obj, "degree", 3),
            .order = getUsize(obj, "order", 1),
            .smoothing = getUsize(obj, "smoothing", 0),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),
        .parabolic_vertex => createWithParams(pvtx_mod.ParabolicVertex, allocator, pvtx_mod.ParabolicVertex.init(.{
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),
        .cubic_vertex => createWithParams(cvtx_mod.CubicVertex, allocator, cvtx_mod.CubicVertex.init(.{
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),
        .quantum_price_levels => createWithAllocParams(qpl_mod.QuantumPriceLevels, qpl_mod.QuantumPriceLevelsParams, allocator, obj, .{
            .lookback = getUsize(obj, "lookback", 2048),
            .num_levels = getUsize(obj, "numLevels", 21),
            .num_bins = getUsize(obj, "numBins", 100),
            .scale_factor = getF64(obj, "scaleFactor", 0.21),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),
        .moving_mini_max => createWithAllocParams(mmm_mod.MovingMiniMax, mmm_mod.MovingMiniMaxParams, allocator, obj, .{
            .m = getUsize(obj, "m", 5),
            .n = getUsize(obj, "n", 50),
            .num_extrema = getUsize(obj, "numExtrema", 3),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),

        // ── william blau ───────────────────────────────────────────────────
        .true_strength_index => createWithAllocParams(tsi_mod.TrueStrengthIndex, tsi_mod.TrueStrengthIndexParams, allocator, obj, .{
            .q = getUsize(obj, "q", 2),
            .r = getUsize(obj, "r", 20),
            .s = getUsize(obj, "s", 5),
            .u = getUsize(obj, "u", 3),
            .ul = getUsize(obj, "ul", 3),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),
        .ergodic_oscillator => createWithAllocParams(ergodic_mod.ErgodicOscillator, ergodic_mod.ErgodicOscillatorParams, allocator, obj, .{
            .q = getUsize(obj, "q", 2),
            .r = getUsize(obj, "r", 20),
            .s = getUsize(obj, "s", 5),
            .u = getUsize(obj, "u", 3),
            .ul = getUsize(obj, "ul", 3),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),
        .mean_deviation_index => createWithParams(mdi_mod.MeanDeviationIndex, allocator, mdi_mod.MeanDeviationIndex.init(.{
            .r = getUsize(obj, "r", 20),
            .s = getUsize(obj, "s", 5),
            .u = getUsize(obj, "u", 3),
            .ul = getUsize(obj, "ul", 3),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),
        .macd_index => createWithParams(macdi_mod.MacdIndex, allocator, macdi_mod.MacdIndex.init(.{
            .r = getUsize(obj, "r", 20),
            .s = getUsize(obj, "s", 5),
            .u = getUsize(obj, "u", 3),
            .ul = getUsize(obj, "ul", 3),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        })),
        .double_smoothed_momenta => createWithAllocParams(dm_mod.DoubleSmoothedMomenta, dm_mod.DoubleSmoothedMomentaParams, allocator, obj, .{
            .a = getUsize(obj, "a", 2),
            .y = getUsize(obj, "y", 2),
            .z = getUsize(obj, "z", 14),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),
        .candlestick_momentum_index => createWithParams(cmi_mod.CandlestickMomentumIndex, allocator, cmi_mod.CandlestickMomentumIndex.init(.{
            .r = getUsize(obj, "r", 20),
            .s = getUsize(obj, "s", 5),
            .u = getUsize(obj, "u", 3),
            .ul = getUsize(obj, "ul", 3),
        })),
        .candlestick_strength_index => createWithParams(csi_mod.CandlestickStrengthIndex, allocator, csi_mod.CandlestickStrengthIndex.init(.{
            .r = getUsize(obj, "r", 20),
            .s = getUsize(obj, "s", 5),
            .u = getUsize(obj, "u", 3),
            .ul = getUsize(obj, "ul", 3),
        })),
        .stochastic_momentum_index => createWithAllocParams(smi_mod.StochasticMomentumIndex, smi_mod.StochasticMomentumIndexParams, allocator, obj, .{
            .q = getUsize(obj, "q", 5),
            .r = getUsize(obj, "r", 20),
            .s = getUsize(obj, "s", 5),
            .u = getUsize(obj, "u", 3),
            .ul = getUsize(obj, "ul", 3),
        }),
        .double_smoothed_stochastic => createWithAllocParams(dss_mod.DoubleSmoothedStochastic, dss_mod.DoubleSmoothedStochasticParams, allocator, obj, .{
            .q = getUsize(obj, "q", 5),
            .r = getUsize(obj, "r", 7),
            .s = getUsize(obj, "s", 3),
            .g = getUsize(obj, "g", 3),
        }),
        .directional_trend_index => createWithAllocParams(dti_mod.DirectionalTrendIndex, dti_mod.DirectionalTrendIndexParams, allocator, obj, .{
            .q = getUsize(obj, "q", 2),
            .r = getUsize(obj, "r", 20),
            .s = getUsize(obj, "s", 5),
            .u = getUsize(obj, "u", 3),
            .ul = getUsize(obj, "ul", 3),
        }),
        .tick_volume_indicator => createWithParams(tvi_mod.TickVolumeIndicator, allocator, tvi_mod.TickVolumeIndicator.init(.{
            .r = getUsize(obj, "r", 12),
            .s = getUsize(obj, "s", 12),
            .u = getUsize(obj, "u", 1),
        })),
        .slope_divergence_tsi_filter => createWithAllocParams(sdtsi_mod.SlopeDivergenceTsiFilter, sdtsi_mod.SlopeDivergenceTsiFilterParams, allocator, obj, .{
            .q = getUsize(obj, "q", 2),
            .r = getUsize(obj, "r", 32),
            .s = getUsize(obj, "s", 32),
            .u = getUsize(obj, "u", 7),
            .x = getUsize(obj, "x", 32),
            .y = getUsize(obj, "y", 7),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),
        .adx_type_filter => createWithAllocParams(atf_mod.AdxTypeFilter, atf_mod.AdxTypeFilterParams, allocator, obj, .{
            .source = std.enums.fromInt(atf_mod.AdxTypeFilterSource, getInt(obj, "source", 0)) orelse
                return FactoryError.IndicatorInitFailed,
            .q = getUsize(obj, "q", 0),
            .r = getUsize(obj, "r", 32),
            .s = getUsize(obj, "s", 32),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),
        .nonambiguous_trend_filter => createWithAllocParams(ntf_mod.NonambiguousTrendFilter, ntf_mod.NonambiguousTrendFilterParams, allocator, obj, .{
            .base = std.enums.fromInt(ntf_mod.NonambiguousTrendFilterBase, getInt(obj, "base", 0)) orelse
                return FactoryError.IndicatorInitFailed,
            .q = getUsize(obj, "q", 0),
            .r = getUsize(obj, "r", 0),
            .s = getUsize(obj, "s", 0),
            .u = getUsize(obj, "u", 0),
            .bar_component = getBarComponent(obj),
            .quote_component = getQuoteComponent(obj),
            .trade_component = getTradeComponent(obj),
        }),
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

const lvt_bps=@import("../lars_von_thienen/bartels_periodicity_score/bartels_periodicity_score.zig");
const lvt_cscan=@import("../lars_von_thienen/complete_cycle_scanner/complete_cycle_scanner.zig");
const lvt_ccyc=@import("../lars_von_thienen/composite_cycle/composite_cycle.zig");
const lvt_cphase=@import("../lars_von_thienen/cycle_phase_score/cycle_phase_score.zig");
const lvt_cpol=@import("../lars_von_thienen/cycle_polarity/cycle_polarity.zig");
const lvt_cstr=@import("../lars_von_thienen/cycle_strength/cycle_strength.zig");
const lvt_csmo=@import("../lars_von_thienen/cycle_swing_momentum/cycle_swing_momentum.zig");
const lvt_crsi=@import("../lars_von_thienen/cyclic_smoothed_relative_strength_index/cyclic_smoothed_relative_strength_index.zig");
const lvt_dro=@import("../lars_von_thienen/detrended_rhythm_oscillators/detrended_rhythm_oscillators.zig");
const lvt_dmv=@import("../lars_von_thienen/dominant_market_vibration/dominant_market_vibration.zig");
const lvt_dce=@import("../lars_von_thienen/dynamic_cycle_explorer/dynamic_cycle_explorer.zig");
const lvt_gcs=@import("../lars_von_thienen/goertzel_cycle_spectrum/goertzel_cycle_spectrum.zig");
const lvt_hpsd=@import("../lars_von_thienen/hodrick_prescott_smoothing_and_detrending/hodrick_prescott_smoothing_and_detrending.zig");
const lvt_mrsi=@import("../lars_von_thienen/metonic_relative_strength_index/metonic_relative_strength_index.zig");
const lvt_msent=@import("../lars_von_thienen/metonic_sentiment_index/metonic_sentiment_index.zig");
const lvt_mwillr=@import("../lars_von_thienen/metonic_williams_percent_r/metonic_williams_percent_r.zig");
const lvt_phosc=@import("../lars_von_thienen/phasing_oscillator/phasing_oscillator.zig");
const lvt_lzlema=@import("../lars_von_thienen/zero_lag_exponential_moving_average/zero_lag_exponential_moving_average.zig");
