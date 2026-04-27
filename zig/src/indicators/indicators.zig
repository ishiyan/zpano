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
}
