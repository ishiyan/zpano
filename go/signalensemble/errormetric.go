package signalensemble

// ErrorMetric selects the error metric used by inverse-variance and rank-based methods.
type ErrorMetric int

const (
	Absolute ErrorMetric = iota // |signal_i - outcome|
	Squared                     // (signal_i - outcome)^2
)
