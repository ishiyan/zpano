package stochastic

// MovingAverageType specifies the type of moving average to use for smoothing.
type MovingAverageType int

const (
	// SMA uses a Simple Moving Average.
	SMA MovingAverageType = iota

	// EMA uses an Exponential Moving Average.
	EMA
)

// StochasticParams describes parameters to create an instance of the indicator.
type StochasticParams struct {
	// FastKLength is the lookback period for the raw %K calculation (highest high / lowest low).
	//
	// The value should be greater than 0. The default value is 5.
	FastKLength int

	// SlowKLength is the smoothing period for Slow-K (also known as Fast-D).
	//
	// The value should be greater than 0. The default value is 3.
	SlowKLength int

	// SlowDLength is the smoothing period for Slow-D.
	//
	// The value should be greater than 0. The default value is 3.
	SlowDLength int

	// SlowKMAType specifies the type of moving average for Slow-K smoothing (SMA or EMA).
	//
	// If zero (SMA), the Simple Moving Average is used.
	SlowKMAType MovingAverageType

	// SlowDMAType specifies the type of moving average for Slow-D smoothing (SMA or EMA).
	//
	// If zero (SMA), the Simple Moving Average is used.
	SlowDMAType MovingAverageType

	// FirstIsAverage controls the EMA seeding algorithm.
	// When true, the first EMA value is the simple average of the first period values.
	// When false (default), the first input value is used directly (Metastock style).
	// Only relevant when an MA type is EMA.
	FirstIsAverage bool
}

// DefaultParams returns a [StochasticParams] value populated with conventional defaults.
func DefaultParams() *StochasticParams {
	return &StochasticParams{
		FastKLength: 5,
		SlowKLength: 3,
		SlowDLength: 3,
	}
}
