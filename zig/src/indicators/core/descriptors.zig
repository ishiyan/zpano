const descriptor_mod = @import("descriptor.zig");
const output_descriptor_mod = @import("output_descriptor.zig");
const identifier_mod = @import("identifier.zig");
const adaptivity_mod = @import("adaptivity.zig");
const input_requirement_mod = @import("input_requirement.zig");
const volume_usage_mod = @import("volume_usage.zig");
const shape_mod = @import("outputs/shape.zig");
const pane_mod = @import("pane.zig");
const role_mod = @import("role.zig");

const Descriptor = descriptor_mod.Descriptor;
const OD = output_descriptor_mod.OutputDescriptor;
const Id = identifier_mod.Identifier;
const A = adaptivity_mod.Adaptivity;
const IR = input_requirement_mod.InputRequirement;
const VU = volume_usage_mod.VolumeUsage;
const S = shape_mod.Shape;
const P = pane_mod.Pane;
const R = role_mod.Role;

/// Static registry of taxonomic descriptors for all implemented indicators.
/// Output Kind values are 1-based (matching Go's iota+1 per-indicator output enums).
pub const descriptors = [_]Descriptor{
    .{ .identifier = .simple_moving_average, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .weighted_moving_average, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .triangular_moving_average, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .exponential_moving_average, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .double_exponential_moving_average, .family = "Patrick Mulloy", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .triple_exponential_moving_average, .family = "Patrick Mulloy", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .t2_exponential_moving_average, .family = "Tim Tillson", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .t3_exponential_moving_average, .family = "Tim Tillson", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .kaufman_adaptive_moving_average, .family = "Perry Kaufman", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .jurik_moving_average, .family = "Mark Jurik", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .mesa_adaptive_moving_average, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 2, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 3, .shape = .band, .role = .envelope, .pane = .price },
    } },
    .{ .identifier = .fractal_adaptive_moving_average, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 2, .shape = .scalar, .role = .fractal_dimension, .pane = .own },
    } },
    .{ .identifier = .dominant_cycle, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .cycle_period, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .cycle_period, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .cycle_phase, .pane = .own },
    } },
    .{ .identifier = .momentum, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .rate_of_change, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .rate_of_change_percent, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .relative_strength_index, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own }} },
    .{ .identifier = .chande_momentum_oscillator, .family = "Tushar Chande", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own }} },
    .{ .identifier = .bollinger_bands, .family = "John Bollinger", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .envelope, .pane = .price },
        .{ .kind = 2, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 3, .shape = .scalar, .role = .envelope, .pane = .price },
        .{ .kind = 4, .shape = .scalar, .role = .volatility, .pane = .own },
        .{ .kind = 5, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 6, .shape = .band, .role = .envelope, .pane = .price },
    } },
    .{ .identifier = .variance, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .volatility, .pane = .own }} },
    .{ .identifier = .standard_deviation, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .volatility, .pane = .own }} },
    .{ .identifier = .goertzel_spectrum, .family = "Custom", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .heatmap, .role = .spectrum, .pane = .own }} },
    .{ .identifier = .center_of_gravity_oscillator, .family = "John Ehlers", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .signal, .pane = .own },
    } },
    .{ .identifier = .cyber_cycle, .family = "John Ehlers", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .signal, .pane = .own },
    } },
    .{ .identifier = .instantaneous_trend_line, .family = "John Ehlers", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 2, .shape = .scalar, .role = .signal, .pane = .price },
    } },
    .{ .identifier = .super_smoother, .family = "John Ehlers", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .zero_lag_exponential_moving_average, .family = "John Ehlers", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .zero_lag_error_correcting_exponential_moving_average, .family = "John Ehlers", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .roofing_filter, .family = "John Ehlers", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .true_range, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .volatility, .pane = .own }} },
    .{ .identifier = .average_true_range, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .volatility, .pane = .own }} },
    .{ .identifier = .normalized_average_true_range, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .volatility, .pane = .own }} },
    .{ .identifier = .directional_movement_minus, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .directional, .pane = .own }} },
    .{ .identifier = .directional_movement_plus, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .directional, .pane = .own }} },
    .{ .identifier = .directional_indicator_minus, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .volatility, .pane = .own },
        .{ .kind = 4, .shape = .scalar, .role = .volatility, .pane = .own },
    } },
    .{ .identifier = .directional_indicator_plus, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .volatility, .pane = .own },
        .{ .kind = 4, .shape = .scalar, .role = .volatility, .pane = .own },
    } },
    .{ .identifier = .directional_movement_index, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 4, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 5, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 6, .shape = .scalar, .role = .volatility, .pane = .own },
        .{ .kind = 7, .shape = .scalar, .role = .volatility, .pane = .own },
    } },
    .{ .identifier = .average_directional_movement_index, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 4, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 5, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 6, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 7, .shape = .scalar, .role = .volatility, .pane = .own },
        .{ .kind = 8, .shape = .scalar, .role = .volatility, .pane = .own },
    } },
    .{ .identifier = .average_directional_movement_index_rating, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 4, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 5, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 6, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 7, .shape = .scalar, .role = .directional, .pane = .own },
        .{ .kind = 8, .shape = .scalar, .role = .volatility, .pane = .own },
        .{ .kind = 9, .shape = .scalar, .role = .volatility, .pane = .own },
    } },
    .{ .identifier = .williams_percent_r, .family = "Larry Williams", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own }} },
    .{ .identifier = .percentage_price_oscillator, .family = "Gerald Appel", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .absolute_price_oscillator, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .commodity_channel_index, .family = "Donald Lambert", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own }} },
    .{ .identifier = .money_flow_index, .family = "Gene Quong", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .aggregate_bar_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own }} },
    .{ .identifier = .on_balance_volume, .family = "Joseph Granville", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .aggregate_bar_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .volume_flow, .pane = .own }} },
    .{ .identifier = .balance_of_power, .family = "Igor Livshin", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own }} },
    .{ .identifier = .rate_of_change_ratio, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .pearsons_correlation_coefficient, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .correlation, .pane = .own }} },
    .{ .identifier = .linear_regression, .family = "Common", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 2, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 3, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 4, .shape = .scalar, .role = .oscillator, .pane = .own },
        .{ .kind = 5, .shape = .scalar, .role = .oscillator, .pane = .own },
    } },
    .{ .identifier = .ultimate_oscillator, .family = "Larry Williams", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own }} },
    .{ .identifier = .stochastic_relative_strength_index, .family = "Tushar Chande", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .signal, .pane = .own },
    } },
    .{ .identifier = .stochastic, .family = "George Lane", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .signal, .pane = .own },
    } },
    .{ .identifier = .aroon, .family = "Tushar Chande", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .oscillator, .pane = .own },
    } },
    .{ .identifier = .advance_decline, .family = "Marc Chaikin", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .aggregate_bar_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .volume_flow, .pane = .own }} },
    .{ .identifier = .advance_decline_oscillator, .family = "Marc Chaikin", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .aggregate_bar_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .volume_flow, .pane = .own }} },
    .{ .identifier = .parabolic_stop_and_reverse, .family = "Welles Wilder", .adaptivity = .static_, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .overlay, .pane = .price }} },
    .{ .identifier = .triple_exponential_moving_average_oscillator, .family = "Jack Hutson", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .bollinger_bands_trend, .family = "John Bollinger", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .moving_average_convergence_divergence, .family = "Gerald Appel", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .signal, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .histogram, .pane = .own },
    } },
    .{ .identifier = .sine_wave, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .signal, .pane = .own },
        .{ .kind = 3, .shape = .band, .role = .envelope, .pane = .own },
        .{ .kind = 4, .shape = .scalar, .role = .cycle_period, .pane = .own },
        .{ .kind = 5, .shape = .scalar, .role = .cycle_phase, .pane = .own },
    } },
    .{ .identifier = .hilbert_transformer_instantaneous_trend_line, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 2, .shape = .scalar, .role = .cycle_period, .pane = .own },
    } },
    .{ .identifier = .trend_cycle_mode, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .regime_flag, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .regime_flag, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .regime_flag, .pane = .own },
        .{ .kind = 4, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 5, .shape = .scalar, .role = .oscillator, .pane = .own },
        .{ .kind = 6, .shape = .scalar, .role = .signal, .pane = .own },
        .{ .kind = 7, .shape = .scalar, .role = .cycle_period, .pane = .own },
        .{ .kind = 8, .shape = .scalar, .role = .cycle_phase, .pane = .own },
    } },
    .{ .identifier = .corona_spectrum, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .heatmap, .role = .spectrum, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .cycle_period, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .cycle_period, .pane = .own },
    } },
    .{ .identifier = .corona_signal_to_noise_ratio, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .heatmap, .role = .spectrum, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
    } },
    .{ .identifier = .corona_swing_position, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .heatmap, .role = .spectrum, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .bounded_oscillator, .pane = .own },
    } },
    .{ .identifier = .corona_trend_vigor, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .heatmap, .role = .spectrum, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .oscillator, .pane = .own },
    } },
    .{ .identifier = .adaptive_trend_and_cycle_filter, .family = "Vladimir Kravchuk", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 2, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 3, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 4, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 5, .shape = .scalar, .role = .smoother, .pane = .price },
        .{ .kind = 6, .shape = .scalar, .role = .oscillator, .pane = .own },
        .{ .kind = 7, .shape = .scalar, .role = .oscillator, .pane = .own },
        .{ .kind = 8, .shape = .scalar, .role = .oscillator, .pane = .own },
    } },
    .{ .identifier = .maximum_entropy_spectrum, .family = "Custom", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .heatmap, .role = .spectrum, .pane = .own }} },
    .{ .identifier = .discrete_fourier_transform_spectrum, .family = "John Ehlers", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .heatmap, .role = .spectrum, .pane = .own }} },
    .{ .identifier = .comb_band_pass_spectrum, .family = "John Ehlers", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .heatmap, .role = .spectrum, .pane = .own }} },
    .{ .identifier = .auto_correlation_indicator, .family = "John Ehlers", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .heatmap, .role = .correlation, .pane = .own }} },
    .{ .identifier = .auto_correlation_periodogram, .family = "John Ehlers", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .heatmap, .role = .spectrum, .pane = .own }} },
    .{ .identifier = .jurik_relative_trend_strength_index, .family = "Mark Jurik", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .jurik_composite_fractal_behavior_index, .family = "Mark Jurik", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .jurik_zero_lag_velocity, .family = "Mark Jurik", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .jurik_directional_movement_index, .family = "Mark Jurik", .adaptivity = .adaptive, .input_requirement = .bar_input, .volume_usage = .no_volume, .outputs = &[_]OD{
        .{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own },
        .{ .kind = 2, .shape = .scalar, .role = .oscillator, .pane = .own },
        .{ .kind = 3, .shape = .scalar, .role = .oscillator, .pane = .own },
    } },
    .{ .identifier = .jurik_turning_point_oscillator, .family = "Mark Jurik", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .jurik_commodity_channel_index, .family = "Mark Jurik", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .jurik_wavelet_sampler, .family = "Mark Jurik", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .jurik_adaptive_zero_lag_velocity, .family = "Mark Jurik", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .jurik_fractal_adaptive_zero_lag_velocity, .family = "Mark Jurik", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .jurik_adaptive_relative_trend_strength_index, .family = "Mark Jurik", .adaptivity = .adaptive, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .oscillator, .pane = .own }} },
    .{ .identifier = .arnaud_legoux_moving_average, .family = "Arnaud Legoux", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
    .{ .identifier = .new_moving_average, .family = "Manfred Dürschner", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }} },
};
