/// Aggregation method for combining multiple signal sources.
///
/// Stateless methods (`Fixed`, `Equal`) ignore feedback entirely.
/// Adaptive methods update weights based on observed outcomes.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AggregationMethod {
    /// User-supplied static weights.
    Fixed = 0,
    /// Uniform 1/n weights.
    Equal = 1,
    /// Weight by 1/variance of errors.
    InverseVariance = 2,
    /// EMA of accuracy.
    ExponentialDecay = 3,
    /// Hedge algorithm (online learning).
    MultiplicativeWeights = 4,
    /// Weight by rank of rolling accuracy.
    RankBased = 5,
    /// Bayesian model averaging.
    Bayesian = 6,
}
