package stochasticrelativestrengthindex

import "zpano/entities"

// MovingAverageType specifies the type of moving average to use for the Fast-D smoothing.
type MovingAverageType int

const (
	// SMA uses a Simple Moving Average.
	SMA MovingAverageType = iota

	// EMA uses an Exponential Moving Average.
	EMA
)

// StochasticRelativeStrengthIndexParams describes parameters to create an instance of the indicator.
type StochasticRelativeStrengthIndexParams struct {
	// Length is the number of periods for the RSI calculation.
	//
	// The value should be greater than 1. The default value is 14.
	Length int

	// FastKLength is the number of periods for the Fast-K stochastic calculation.
	//
	// The value should be greater than 0. The default value is 5.
	FastKLength int

	// FastDLength is the number of periods for the Fast-D smoothing.
	//
	// The value should be greater than 0. The default value is 3.
	FastDLength int

	// MovingAverageType specifies the type of moving average for Fast-D (SMA or EMA).
	//
	// If zero (SMA), the Simple Moving Average is used.
	MovingAverageType MovingAverageType

	// FirstIsAverage controls the EMA seeding algorithm.
	// When true, the first EMA value is the simple average of the first period values.
	// When false (default), the first input value is used directly (Metastock style).
	// Only relevant when MovingAverageType is EMA.
	FirstIsAverage bool

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

// DefaultParams returns a [StochasticRelativeStrengthIndexParams] value populated with conventional defaults.
func DefaultParams() *StochasticRelativeStrengthIndexParams {
	return &StochasticRelativeStrengthIndexParams{
		Length:      14,
		FastKLength: 5,
		FastDLength: 3,
	}
}
