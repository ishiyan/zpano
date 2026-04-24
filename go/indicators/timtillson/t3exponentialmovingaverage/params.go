package t3exponentialmovingaverage

import "zpano/entities"

// T3ExponentialMovingAverageLengthParams describes parameters to create an instance of the indicator
// based on length.
type T3ExponentialMovingAverageLengthParams struct {
	// Length is the length (the number of time periods, l) of the moving window to calculate the average.
	//
	// The value should be greater than 1.
	Length int

	// VolumeFactor is the volume factor, v (0 <= v <= 1), of the exponential moving average.
	// The default value is 0.7.
	// When v=0, T3 is just an EMA, and when v=1, T3 is TEMA.
	// In between, T3 is a cooler TEMA.
	VolumeFactor float64

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

// T3ExponentialMovingAverageSmoothingFactorParams describes parameters to create an instance of the indicator
// based on smoothing factor.
type T3ExponentialMovingAverageSmoothingFactorParams struct {
	// SmoothingFactor is the smoothing factor, a (0 < a < 1), of the exponential moving average.
	//
	// The equivalent length l is:
	//    l = 2/a - 1, 0<a<1, 1<=l.
	SmoothingFactor float64

	// VolumeFactor is the volume factor, v (0 <= v <= 1), of the exponential moving average.
	// The default value is 0.7.
	// When v=0, T3 is just an EMA, and when v=1, T3 is TEMA.
	// In between, T3 is a cooler TEMA.
	VolumeFactor float64

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

// DefaultLengthParams returns a [T3ExponentialMovingAverageLengthParams] value populated with conventional defaults.
func DefaultLengthParams() *T3ExponentialMovingAverageLengthParams {
	return &T3ExponentialMovingAverageLengthParams{
		Length:       5,
		VolumeFactor: 0.7,
	}
}

// DefaultSmoothingFactorParams returns a [T3ExponentialMovingAverageSmoothingFactorParams] value populated with conventional defaults.
func DefaultSmoothingFactorParams() *T3ExponentialMovingAverageSmoothingFactorParams {
	return &T3ExponentialMovingAverageSmoothingFactorParams{
		SmoothingFactor: 0.3333,
		VolumeFactor:    0.7,
	}
}
