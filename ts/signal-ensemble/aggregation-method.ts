/** Aggregation method for combining multiple signal sources.
 *
 * Stateless methods (FIXED, EQUAL) ignore feedback entirely.
 * Adaptive methods update weights based on observed outcomes.
 */
export enum AggregationMethod {
    /** User-supplied static weights. */
    FIXED = 0,
    /** Uniform 1/n weights. */
    EQUAL = 1,
    /** Weight by 1/variance of errors. */
    INVERSE_VARIANCE = 2,
    /** EMA of accuracy. */
    EXPONENTIAL_DECAY = 3,
    /** Hedge algorithm (online learning). */
    MULTIPLICATIVE_WEIGHTS = 4,
    /** Weight by rank of rolling accuracy. */
    RANK_BASED = 5,
    /** Bayesian model averaging. */
    BAYESIAN = 6,
}
