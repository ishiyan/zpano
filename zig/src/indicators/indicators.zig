// Root module for the indicators library.
// Re-exports all core types and sub-modules.

// --- Core enums ---
pub const shape = @import("core/outputs/shape.zig");
pub const role = @import("core/role.zig");
pub const pane = @import("core/pane.zig");
pub const adaptivity = @import("core/adaptivity.zig");
pub const input_requirement = @import("core/input_requirement.zig");
pub const volume_usage = @import("core/volume_usage.zig");

// --- Core types ---
pub const identifier = @import("core/identifier.zig");
pub const output_descriptor = @import("core/output_descriptor.zig");
pub const output_metadata = @import("core/output_metadata.zig");
pub const metadata = @import("core/metadata.zig");
pub const descriptor = @import("core/descriptor.zig");
pub const descriptors = @import("core/descriptors.zig");
pub const build_metadata = @import("core/build_metadata.zig");
pub const component_triple_mnemonic = @import("core/component_triple_mnemonic.zig");

// --- Indicator interface and outputs ---
pub const indicator = @import("core/indicator.zig");
pub const line_indicator = @import("core/line_indicator.zig");

// --- Output types ---
pub const band = @import("core/outputs/band.zig");
pub const heatmap = @import("core/outputs/heatmap.zig");
pub const polyline = @import("core/outputs/polyline.zig");

// --- Factory ---
pub const factory = @import("factory/factory.zig");

// --- Frequency Response ---
pub const frequency_response = @import("core/frequency_response.zig");

// --- Convenience type aliases ---
pub const Identifier = identifier.Identifier;
pub const Shape = shape.Shape;
pub const Role = role.Role;
pub const Pane = pane.Pane;
pub const Adaptivity = adaptivity.Adaptivity;
pub const InputRequirement = input_requirement.InputRequirement;
pub const VolumeUsage = volume_usage.VolumeUsage;
pub const OutputDescriptor = output_descriptor.OutputDescriptor;
pub const OutputMetadata = output_metadata.OutputMetadata;
pub const Metadata = metadata.Metadata;
pub const Descriptor = descriptor.Descriptor;
pub const Indicator = indicator.Indicator;
pub const OutputValue = indicator.OutputValue;
pub const OutputArray = indicator.OutputArray;
pub const LineIndicator = line_indicator.LineIndicator;
pub const Band = band.Band;
pub const Heatmap = heatmap.Heatmap;
pub const Polyline = polyline.Polyline;
pub const Point = polyline.Point;

// --- Functions ---
pub const descriptorOf = descriptor.descriptorOf;
pub const allDescriptors = descriptor.allDescriptors;
pub const buildMetadata = build_metadata.buildMetadata;
pub const componentTripleMnemonic = component_triple_mnemonic.componentTripleMnemonic;

// --- Indicator implementations ---
pub const simple_moving_average = @import("common/simple_moving_average/simple_moving_average.zig");
pub const weighted_moving_average = @import("common/weighted_moving_average/weighted_moving_average.zig");
pub const triangular_moving_average = @import("common/triangular_moving_average/triangular_moving_average.zig");
pub const exponential_moving_average = @import("common/exponential_moving_average/exponential_moving_average.zig");
pub const momentum = @import("common/momentum/momentum.zig");
pub const rate_of_change = @import("common/rate_of_change/rate_of_change.zig");
pub const rate_of_change_percent = @import("common/rate_of_change_percent/rate_of_change_percent.zig");
pub const rate_of_change_ratio = @import("common/rate_of_change_ratio/rate_of_change_ratio.zig");
pub const variance = @import("common/variance/variance.zig");
pub const standard_deviation = @import("common/standard_deviation/standard_deviation.zig");
pub const absolute_price_oscillator = @import("common/absolute_price_oscillator/absolute_price_oscillator.zig");
pub const linear_regression = @import("common/linear_regression/linear_regression.zig");
pub const pearsons_correlation_coefficient = @import("common/pearsons_correlation_coefficient/pearsons_correlation_coefficient.zig");
pub const balance_of_power = @import("igor_livshin/balance_of_power/balance_of_power.zig");
pub const on_balance_volume = @import("joseph_granville/on_balance_volume/on_balance_volume.zig");
pub const commodity_channel_index = @import("donald_lambert/commodity_channel_index/commodity_channel_index.zig");
pub const triple_exponential_moving_average_oscillator = @import("jack_hutson/triple_exponential_moving_average_oscillator/triple_exponential_moving_average_oscillator.zig");
pub const money_flow_index = @import("gene_quong/money_flow_index/money_flow_index.zig");
pub const kaufman_adaptive_moving_average = @import("perry_kaufman/kaufman_adaptive_moving_average/kaufman_adaptive_moving_average.zig");
pub const jurik_moving_average = @import("mark_jurik/jurik_moving_average/jurik_moving_average.zig");
pub const jurik_relative_trend_strength_index = @import("mark_jurik/jurik_relative_trend_strength_index/jurik_relative_trend_strength_index.zig");
pub const jurik_zero_lag_velocity = @import("mark_jurik/jurik_zero_lag_velocity/jurik_zero_lag_velocity.zig");
pub const jurik_composite_fractal_behavior_index = @import("mark_jurik/jurik_composite_fractal_behavior_index/jurik_composite_fractal_behavior_index.zig");
pub const jurik_directional_movement_index = @import("mark_jurik/jurik_directional_movement_index/jurik_directional_movement_index.zig");
pub const jurik_turning_point_oscillator = @import("mark_jurik/jurik_turning_point_oscillator/jurik_turning_point_oscillator.zig");
pub const jurik_adaptive_relative_trend_strength_index = @import("mark_jurik/jurik_adaptive_relative_trend_strength_index/jurik_adaptive_relative_trend_strength_index.zig");
pub const jurik_adaptive_zero_lag_velocity = @import("mark_jurik/jurik_adaptive_zero_lag_velocity/jurik_adaptive_zero_lag_velocity.zig");
pub const jurik_commodity_channel_index = @import("mark_jurik/jurik_commodity_channel_index/jurik_commodity_channel_index.zig");
pub const jurik_fractal_adaptive_zero_lag_velocity = @import("mark_jurik/jurik_fractal_adaptive_zero_lag_velocity/jurik_fractal_adaptive_zero_lag_velocity.zig");
pub const jurik_wavelet_sampler = @import("mark_jurik/jurik_wavelet_sampler/jurik_wavelet_sampler.zig");
pub const stochastic = @import("george_lane/stochastic/stochastic.zig");
pub const adaptive_trend_and_cycle_filter = @import("vladimir_kravchuk/adaptive_trend_and_cycle_filter/adaptive_trend_and_cycle_filter.zig");
pub const double_exponential_moving_average = @import("patrick_mulloy/double_exponential_moving_average/double_exponential_moving_average.zig");
pub const triple_exponential_moving_average = @import("patrick_mulloy/triple_exponential_moving_average/triple_exponential_moving_average.zig");
pub const t2_exponential_moving_average = @import("tim_tillson/t2_exponential_moving_average/t2_exponential_moving_average.zig");
pub const t3_exponential_moving_average = @import("tim_tillson/t3_exponential_moving_average/t3_exponential_moving_average.zig");
pub const percentage_price_oscillator = @import("gerald_appel/percentage_price_oscillator/percentage_price_oscillator.zig");
pub const moving_average_convergence_divergence = @import("gerald_appel/moving_average_convergence_divergence/moving_average_convergence_divergence.zig");
pub const bollinger_bands = @import("john_bollinger/bollinger_bands/bollinger_bands.zig");
pub const bollinger_bands_trend = @import("john_bollinger/bollinger_bands_trend/bollinger_bands_trend.zig");
pub const williams_percent_r = @import("larry_williams/williams_percent_r/williams_percent_r.zig");
pub const ultimate_oscillator = @import("larry_williams/ultimate_oscillator/ultimate_oscillator.zig");
pub const advance_decline = @import("marc_chaikin/advance_decline/advance_decline.zig");
pub const advance_decline_oscillator = @import("marc_chaikin/advance_decline_oscillator/advance_decline_oscillator.zig");
pub const aroon = @import("tushar_chande/aroon/aroon.zig");
pub const chande_momentum_oscillator = @import("tushar_chande/chande_momentum_oscillator/chande_momentum_oscillator.zig");
pub const stochastic_relative_strength_index = @import("tushar_chande/stochastic_relative_strength_index/stochastic_relative_strength_index.zig");
pub const true_range = @import("welles_wilder/true_range/true_range.zig");
pub const directional_movement_plus = @import("welles_wilder/directional_movement_plus/directional_movement_plus.zig");
pub const directional_movement_minus = @import("welles_wilder/directional_movement_minus/directional_movement_minus.zig");
pub const relative_strength_index = @import("welles_wilder/relative_strength_index/relative_strength_index.zig");
pub const parabolic_stop_and_reverse = @import("welles_wilder/parabolic_stop_and_reverse/parabolic_stop_and_reverse.zig");
pub const average_true_range = @import("welles_wilder/average_true_range/average_true_range.zig");
pub const normalized_average_true_range = @import("welles_wilder/normalized_average_true_range/normalized_average_true_range.zig");
pub const directional_indicator_plus = @import("welles_wilder/directional_indicator_plus/directional_indicator_plus.zig");
pub const directional_indicator_minus = @import("welles_wilder/directional_indicator_minus/directional_indicator_minus.zig");
pub const directional_movement_index = @import("welles_wilder/directional_movement_index/directional_movement_index.zig");
pub const average_directional_movement_index = @import("welles_wilder/average_directional_movement_index/average_directional_movement_index.zig");
pub const average_directional_movement_index_rating = @import("welles_wilder/average_directional_movement_index_rating/average_directional_movement_index_rating.zig");

pub const super_smoother = @import("john_ehlers/super_smoother/super_smoother.zig");
pub const roofing_filter = @import("john_ehlers/roofing_filter/roofing_filter.zig");
pub const instantaneous_trend_line = @import("john_ehlers/instantaneous_trend_line/instantaneous_trend_line.zig");
pub const cyber_cycle = @import("john_ehlers/cyber_cycle/cyber_cycle.zig");
pub const zero_lag_error_correcting_exponential_moving_average = @import("john_ehlers/zero_lag_error_correcting_exponential_moving_average/zero_lag_error_correcting_exponential_moving_average.zig");
pub const zero_lag_exponential_moving_average = @import("john_ehlers/zero_lag_exponential_moving_average/zero_lag_exponential_moving_average.zig");
pub const center_of_gravity_oscillator = @import("john_ehlers/center_of_gravity_oscillator/center_of_gravity_oscillator.zig");
pub const fractal_adaptive_moving_average = @import("john_ehlers/fractal_adaptive_moving_average/fractal_adaptive_moving_average.zig");
pub const discrete_fourier_transform_spectrum = @import("john_ehlers/discrete_fourier_transform_spectrum/discrete_fourier_transform_spectrum.zig");
pub const autocorrelation_indicator = @import("john_ehlers/autocorrelation_indicator/autocorrelation_indicator.zig");
pub const autocorrelation_periodogram = @import("john_ehlers/autocorrelation_periodogram/autocorrelation_periodogram.zig");
pub const comb_band_pass_spectrum = @import("john_ehlers/comb_band_pass_spectrum/comb_band_pass_spectrum.zig");

// --- Hilbert transformer (helper, not a registered indicator) ---
pub const dominant_cycle = @import("john_ehlers/dominant_cycle/dominant_cycle.zig");
pub const sine_wave = @import("john_ehlers/sinewave/sinewave.zig");
pub const trend_cycle_mode = @import("john_ehlers/trend_cycle_mode/trend_cycle_mode.zig");
pub const hilbert_transformer_instantaneous_trend_line = @import("john_ehlers/hilbert_transformer_instantaneous_trend_line/hilbert_transformer_instantaneous_trend_line.zig");
pub const mesa_adaptive_moving_average = @import("john_ehlers/mesa_adaptive_moving_average/mesa_adaptive_moving_average.zig");

pub const corona = @import("john_ehlers/corona/corona.zig");
pub const corona_spectrum = @import("john_ehlers/corona_spectrum/corona_spectrum.zig");
pub const corona_swing_position = @import("john_ehlers/corona_swing_position/corona_swing_position.zig");
pub const corona_trend_vigor = @import("john_ehlers/corona_trend_vigor/corona_trend_vigor.zig");

pub const goertzel_spectrum = @import("custom/goertzel_spectrum/goertzel_spectrum.zig");
pub const maximum_entropy_spectrum = @import("custom/maximum_entropy_spectrum/maximum_entropy_spectrum.zig");

pub const hilbert_transformer = @import("john_ehlers/hilbert_transformer/hilbert_transformer.zig");
pub const homodyne_discriminator = @import("john_ehlers/hilbert_transformer/homodyne_discriminator.zig");
pub const homodyne_discriminator_unrolled = @import("john_ehlers/hilbert_transformer/homodyne_discriminator_unrolled.zig");
pub const phase_accumulator = @import("john_ehlers/hilbert_transformer/phase_accumulator.zig");
pub const dual_differentiator = @import("john_ehlers/hilbert_transformer/dual_differentiator.zig");

// Force-include tests from sub-modules.
comptime {
    _ = identifier;
    _ = simple_moving_average;
    _ = weighted_moving_average;
    _ = triangular_moving_average;
    _ = exponential_moving_average;
    _ = momentum;
    _ = rate_of_change;
    _ = rate_of_change_percent;
    _ = rate_of_change_ratio;
    _ = variance;
    _ = standard_deviation;
    _ = absolute_price_oscillator;
    _ = linear_regression;
    _ = pearsons_correlation_coefficient;
    _ = balance_of_power;
    _ = on_balance_volume;
    _ = commodity_channel_index;
    _ = triple_exponential_moving_average_oscillator;
    _ = money_flow_index;
    _ = kaufman_adaptive_moving_average;
    _ = jurik_moving_average;
    _ = jurik_relative_trend_strength_index;
    _ = jurik_zero_lag_velocity;
    _ = jurik_composite_fractal_behavior_index;
    _ = jurik_directional_movement_index;
    _ = jurik_turning_point_oscillator;
    _ = jurik_adaptive_relative_trend_strength_index;
    _ = jurik_adaptive_zero_lag_velocity;
    _ = jurik_commodity_channel_index;
    _ = jurik_fractal_adaptive_zero_lag_velocity;
    _ = jurik_wavelet_sampler;
    _ = stochastic;
    _ = adaptive_trend_and_cycle_filter;
    _ = double_exponential_moving_average;
    _ = triple_exponential_moving_average;
    _ = t2_exponential_moving_average;
    _ = t3_exponential_moving_average;
    _ = percentage_price_oscillator;
    _ = moving_average_convergence_divergence;
    _ = bollinger_bands;
    _ = bollinger_bands_trend;
    _ = williams_percent_r;
    _ = ultimate_oscillator;
    _ = advance_decline;
    _ = advance_decline_oscillator;
    _ = aroon;
    _ = chande_momentum_oscillator;
    _ = stochastic_relative_strength_index;
    _ = true_range;
    _ = directional_movement_plus;
    _ = directional_movement_minus;
    _ = relative_strength_index;
    _ = parabolic_stop_and_reverse;
    _ = average_true_range;
    _ = normalized_average_true_range;
    _ = directional_indicator_plus;
    _ = directional_indicator_minus;
    _ = directional_movement_index;
    _ = average_directional_movement_index;
    _ = average_directional_movement_index_rating;
    _ = super_smoother;
    _ = roofing_filter;
    _ = instantaneous_trend_line;
    _ = cyber_cycle;
    _ = zero_lag_error_correcting_exponential_moving_average;
    _ = zero_lag_exponential_moving_average;
    _ = center_of_gravity_oscillator;
    _ = fractal_adaptive_moving_average;
    _ = discrete_fourier_transform_spectrum;
    _ = autocorrelation_indicator;
    _ = autocorrelation_periodogram;
    _ = comb_band_pass_spectrum;
    _ = hilbert_transformer;
    _ = homodyne_discriminator;
    _ = homodyne_discriminator_unrolled;
    _ = phase_accumulator;
    _ = dual_differentiator;
    _ = dominant_cycle;
    _ = sine_wave;
    _ = trend_cycle_mode;
    _ = hilbert_transformer_instantaneous_trend_line;
    _ = mesa_adaptive_moving_average;
    _ = corona;
    _ = corona_spectrum;
    _ = corona_swing_position;
    _ = corona_trend_vigor;
    _ = goertzel_spectrum;
    _ = maximum_entropy_spectrum;
    _ = frequency_response;
    _ = factory;
}
