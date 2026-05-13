/// Aggregation method for combining multiple signal sources.
///
/// Stateless methods (fixed, equal) ignore feedback entirely.
/// Adaptive methods update weights based on observed outcomes.
pub const AggregationMethod = enum(u8) {
    /// User-supplied static weights.
    fixed = 0,
    /// Uniform 1/n weights.
    equal = 1,
    /// Weight by 1/variance of errors.
    inverse_variance = 2,
    /// EMA of accuracy.
    exponential_decay = 3,
    /// Hedge algorithm (online learning).
    multiplicative_weights = 4,
    /// Weight by rank of rolling accuracy.
    rank_based = 5,
    /// Bayesian model averaging.
    bayesian = 6,
};
