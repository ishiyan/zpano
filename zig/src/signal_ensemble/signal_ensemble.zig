/// Signal ensemble: weighted blending of multiple signal sources.
///
/// Provides adaptive weighted blending of independent signal sources
/// with delayed feedback and online weight learning.
pub const method = @import("method.zig");
pub const error_metric = @import("error_metric.zig");
pub const aggregator = @import("aggregator.zig");

// Re-export primary types for convenience.
pub const AggregationMethod = method.AggregationMethod;
pub const ErrorMetric = error_metric.ErrorMetric;
pub const Aggregator = aggregator.Aggregator;
pub const AggregatorParams = aggregator.AggregatorParams;
pub const AggregatorError = aggregator.AggregatorError;
pub const HistoryEntry = aggregator.HistoryEntry;

// Pull in tests from sub-modules.
comptime {
    _ = @import("method.zig");
    _ = @import("error_metric.zig");
    _ = @import("aggregator.zig");
}
