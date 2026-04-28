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
pub const true_range = @import("welles_wilder/true_range/true_range.zig");
pub const directional_movement_plus = @import("welles_wilder/directional_movement_plus/directional_movement_plus.zig");
pub const directional_movement_minus = @import("welles_wilder/directional_movement_minus/directional_movement_minus.zig");

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
    _ = true_range;
    _ = directional_movement_plus;
    _ = directional_movement_minus;
}
