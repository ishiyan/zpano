package bollingerbands

import "zpano/entities"

// MovingAverageType specifies the type of moving average to use in the Bollinger Bands calculation.
type MovingAverageType int

const (
	// SMA uses a Simple Moving Average.
	SMA MovingAverageType = iota

	// EMA uses an Exponential Moving Average.
	EMA
)

// BollingerBandsParams describes parameters to create an instance of the indicator.
type BollingerBandsParams struct {
	// Length is the number of periods for the moving average and standard deviation.
	//
	// The value should be greater than 1. The default value is 5.
	Length int

	// UpperMultiplier is the number of standard deviations above the middle band.
	//
	// The default value is 2.0.
	UpperMultiplier float64

	// LowerMultiplier is the number of standard deviations below the middle band.
	//
	// The default value is 2.0.
	LowerMultiplier float64

	// IsUnbiased indicates whether to use the unbiased sample standard deviation (true)
	// or the population standard deviation (false).
	//
	// If nil, defaults to true (unbiased sample standard deviation).
	IsUnbiased *bool

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

// DefaultParams returns a [BollingerBandsParams] value populated with conventional defaults.
func DefaultParams() *BollingerBandsParams {
	return &BollingerBandsParams{
		Length:          5,
		UpperMultiplier: 2.0,
		LowerMultiplier: 2.0,
	}
}
