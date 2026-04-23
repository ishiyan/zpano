package t2exponentialmovingaverage

import "zpano/entities"

// T2ExponentialMovingAverageLengthParams describes parameters to create an instance of the indicator
// based on length.
type T2ExponentialMovingAverageLengthParams struct {
	// Length is the length (the number of time periods, l) of the moving window to calculate the average.
	//
	// The value should be greater than 1.
	Length int

	// VolumeFactor is the volume factor, v (0 <= v <= 1), of the exponential moving average.
	// The default value is 0.7.
	// When v=0, T2 is just an EMA, and when v=1, T2 is DEMA.
	// In between, T2 is a cooler DEMA.
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

// T2ExponentialMovingAverageSmoothingFactorParams describes parameters to create an instance of the indicator
// based on smoothing factor.
type T2ExponentialMovingAverageSmoothingFactorParams struct {
	// SmoothingFactor is the smoothing factor, a (0 < a < 1), of the exponential moving average.
	//
	// The equivalent length l is:
	//    l = 2/a - 1, 0<a<1, 1<=l.
	SmoothingFactor float64

	// VolumeFactor is the volume factor, v (0 <= v <= 1), of the exponential moving average.
	// The default value is 0.7.
	// When v=0, T2 is just an EMA, and when v=1, T2 is DEMA.
	// In between, T2 is a cooler DEMA.
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
