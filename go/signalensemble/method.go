package signalensemble

// AggregationMethod selects the aggregation strategy for combining multiple signal sources.
// Stateless methods (Fixed, Equal) ignore feedback entirely.
// Adaptive methods update weights based on observed outcomes.
type AggregationMethod int

const (
	Fixed               AggregationMethod = iota // User-supplied static weights
	Equal                                        // Uniform 1/n weights
	InverseVariance                              // Weight by 1/variance of errors
	ExponentialDecay                             // EMA of accuracy
	MultiplicativeWeights                        // Hedge algorithm (online learning)
	RankBased                                    // Weight by rank of rolling accuracy
	Bayesian                                     // Bayesian model averaging
)
