package advancedeclineoscillator

// MovingAverageType specifies the type of moving average to use in the ADOSC calculation.
type MovingAverageType int

const (
	// SMA uses a Simple Moving Average.
	SMA MovingAverageType = iota

	// EMA uses an Exponential Moving Average.
	EMA
)

// AdvanceDeclineOscillatorParams describes parameters to create an instance of the indicator.
type AdvanceDeclineOscillatorParams struct {
	// FastLength is the number of periods for the fast moving average.
	//
	// The value should be greater than 1. Default is 3.
	FastLength int

	// SlowLength is the number of periods for the slow moving average.
	//
	// The value should be greater than 1. Default is 10.
	SlowLength int

	// MovingAverageType specifies the type of moving average (SMA or EMA).
	//
	// If zero (SMA), the Simple Moving Average is used.
	// Use EMA for the Exponential Moving Average (TaLib default).
	MovingAverageType MovingAverageType

	// FirstIsAverage controls the EMA seeding algorithm.
	// When true, the first EMA value is the simple average of the first period values.
	// When false (default), the first input value is used directly (Metastock style).
	// Only relevant when MovingAverageType is EMA.
	FirstIsAverage bool
}
