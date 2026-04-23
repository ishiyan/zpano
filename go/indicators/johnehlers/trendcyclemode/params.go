package trendcyclemode

import (
	"zpano/entities"
	"zpano/indicators/johnehlers/hilberttransformer"
)

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// EstimatorType is the type of cycle estimator to use.
	// The default value is hilberttransformer.HomodyneDiscriminator.
	EstimatorType hilberttransformer.CycleEstimatorType

	// EstimatorParams describes parameters to create an instance
	// of the Hilbert transformer cycle estimator.
	EstimatorParams hilberttransformer.CycleEstimatorParams

	// AlphaEmaPeriodAdditional is the value of α (0 < α ≤ 1) used in EMA
	// for additional smoothing of the instantaneous period.
	//
	// The default value is 0.33.
	AlphaEmaPeriodAdditional float64

	// TrendLineSmoothingLength is the additional WMA smoothing length used to smooth the trend line.
	// The valid values are 2, 3, 4. The default value is 4.
	TrendLineSmoothingLength int

	// CyclePartMultiplier is the multiplier to the dominant cycle period used to determine
	// the window length to calculate the trend line. The typical values are in [0.5, 1.5].
	// The default value is 1.0. Valid range is (0, 10].
	CyclePartMultiplier float64

	// SeparationPercentage is the threshold (in percent) above which a wide separation between
	// the WMA-smoothed price and the instantaneous trend line forces the trend mode.
	//
	// The default value is 1.5. Valid range is (0, 100].
	SeparationPercentage float64

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarMedianPrice) is used. Since the default differs from the framework
	// default bar component, it is always shown in the indicator mnemonic.
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
