package movingaverageconvergencedivergence

import "zpano/entities"

// MovingAverageType specifies the type of moving average to use in the MACD calculation.
type MovingAverageType int

const (
	// EMA uses an Exponential Moving Average (default for classic MACD).
	EMA MovingAverageType = iota

	// SMA uses a Simple Moving Average.
	SMA
)

// MovingAverageConvergenceDivergenceParams describes parameters to create an instance of the indicator.
type MovingAverageConvergenceDivergenceParams struct {
	// FastLength is the number of periods for the fast moving average.
	//
	// The value should be greater than 1. The default value is 12.
	FastLength int

	// SlowLength is the number of periods for the slow moving average.
	//
	// The value should be greater than 1. The default value is 26.
	SlowLength int

	// SignalLength is the number of periods for the signal line moving average.
	//
	// The value should be greater than 0. The default value is 9.
	SignalLength int

	// MovingAverageType specifies the type of moving average for the fast and slow lines (EMA or SMA).
	//
	// If zero (EMA), the Exponential Moving Average is used.
	MovingAverageType MovingAverageType

	// SignalMovingAverageType specifies the type of moving average for the signal line (EMA or SMA).
	//
	// If zero (EMA), the Exponential Moving Average is used.
	SignalMovingAverageType MovingAverageType

	// FirstIsAverage controls the EMA seeding algorithm.
	// When true (default nil), the first EMA value is the simple average of the first period values
	// (TA-Lib compatible). When set to false, the first input value is used directly (Metastock style).
	// Only relevant when MovingAverageType is EMA.
	FirstIsAverage *bool

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarClosePrice) is used and the component is not shown in the indicator mnemonic.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	//
	// If zero, the default (QuoteMidPrice) is used and the component is not shown in the indicator mnemonic.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	//
	// If zero, the default (TradePrice) is used and the component is not shown in the indicator mnemonic.
	TradeComponent entities.TradeComponent
}
