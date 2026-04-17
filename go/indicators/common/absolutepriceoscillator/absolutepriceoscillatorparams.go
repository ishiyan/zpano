package absolutepriceoscillator

import "zpano/entities"

// MovingAverageType specifies the type of moving average to use in the APO calculation.
type MovingAverageType int

const (
	// SMA uses a Simple Moving Average.
	SMA MovingAverageType = iota

	// EMA uses an Exponential Moving Average.
	EMA
)

// AbsolutePriceOscillatorParams describes parameters to create an instance of the indicator.
type AbsolutePriceOscillatorParams struct {
	// FastLength is the number of periods for the fast moving average.
	//
	// The value should be greater than 1.
	FastLength int

	// SlowLength is the number of periods for the slow moving average.
	//
	// The value should be greater than 1.
	SlowLength int

	// MovingAverageType specifies the type of moving average (SMA or EMA).
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
