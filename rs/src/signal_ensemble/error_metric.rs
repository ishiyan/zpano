/// Error metric used by inverse-variance and rank-based methods.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum ErrorMetric {
    /// |signal_i - outcome|
    Absolute = 0,
    /// (signal_i - outcome)^2
    Squared = 1,
}
