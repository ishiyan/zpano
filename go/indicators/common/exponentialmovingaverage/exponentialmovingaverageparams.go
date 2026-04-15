package exponentialmovingaverage

import "zpano/entities"

// ExponentialMovingAverageLengthParams describes parameters to create an instance of the indicator
// based on length.
type ExponentialMovingAverageLengthParams struct {
	// Length is the length (the number of time periods, ℓ) of the moving window to calculate the average.
	//
	// The value should be greater than or equal to 1.
	Length int

	// FirstIsAverage indicates whether the very first exponential moving average value is
	// a simple average of the first 'period' (the most widely documented approach) or
	// the first input value (used in Metastock).
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

// ExponentialMovingAverageSmoothingFactorParams describes parameters to create an instance of the indicator
// based on smoothing factor.
type ExponentialMovingAverageSmoothingFactorParams struct {
	// SmoothingFactor is the smoothing factor, α in (0,1), of the exponential moving average.
	//
	// The equivalent length ℓ is:
	//    ℓ = 2/α - 1, 0<α≤1, 1≤ℓ.
	SmoothingFactor float64

	// FirstIsAverage indicates whether the very first exponential moving average value is
	// a simple average of the first 'period' (the most widely documented approach) or
	// the first input value (used in Metastock).
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
