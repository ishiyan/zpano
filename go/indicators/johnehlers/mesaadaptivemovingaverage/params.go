package mesaadaptivemovingaverage

import (
	"zpano/entities"
	"zpano/indicators/johnehlers/hilberttransformer"
)

// LengthParams describes parameters to create an instance of the indicator
// based on lengths.
type LengthParams struct {
	// EstimatorType is the type of cycle estimator to use.
	// The default value is hilberttransformer.HomodyneDiscriminator.
	EstimatorType hilberttransformer.CycleEstimatorType

	// EstimatorParams describes parameters to create an instance
	// of the Hilbert transformer cycle estimator.
	EstimatorParams hilberttransformer.CycleEstimatorParams

	// FastLimitLength is the fastest boundary length, ℓf.
	// The equivalent smoothing factor αf is
	//
	//   αf = 2/(ℓf + 1), 2 ≤ ℓ
	//
	// The value should be greater than 1.
	// The default value is 3 (αf=0.5).
	FastLimitLength int

	// SlowLimitLength is the slowest boundary length, ℓs.
	// The equivalent smoothing factor αs is
	//
	//   αs = 2/(ℓs + 1), 2 ≤ ℓ
	//
	// The value should be greater than 1.
	// The default value is 39 (αs=0.05).
	SlowLimitLength int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// The original MAMA indicator uses the median price (high+low)/2.
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

// SmoothingFactorParams describes parameters to create an instance of the indicator
// based on smoothing factors.
type SmoothingFactorParams struct {
	// EstimatorType is the type of cycle estimator to use.
	// The default value is hilberttransformer.HomodyneDiscriminator.
	EstimatorType hilberttransformer.CycleEstimatorType

	// EstimatorParams describes parameters to create an instance
	// of the Hilbert transformer cycle estimator.
	EstimatorParams hilberttransformer.CycleEstimatorParams

	// FastLimitSmoothingFactor is the fastest boundary smoothing factor, αf in (0,1].
	// The equivalent length ℓf is
	//
	//   ℓf = 2/αf - 1, 0 < αf ≤ 1, 1 ≤ ℓf
	//
	// The default value is 0.5 (ℓf=3).
	FastLimitSmoothingFactor float64

	// SlowLimitSmoothingFactor is the slowest boundary smoothing factor, αs in (0,1].
	// The equivalent length ℓs is
	//
	//   ℓs = 2/αs - 1, 0 < αs ≤ 1, 1 ≤ ℓs
	//
	// The default value is 0.05 (ℓs=39).
	SlowLimitSmoothingFactor float64

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
