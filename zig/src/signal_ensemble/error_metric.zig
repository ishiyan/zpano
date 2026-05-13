/// Error metric used by inverse-variance and rank-based methods.
pub const ErrorMetric = enum(u8) {
    /// |signal_i - outcome|
    absolute = 0,
    /// (signal_i - outcome)^2
    squared = 1,
};
