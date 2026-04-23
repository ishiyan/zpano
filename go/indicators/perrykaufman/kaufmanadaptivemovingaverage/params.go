package kaufmanadaptivemovingaverage

import "zpano/entities"

// KaufmanAdaptiveMovingAverageLengthParams describes parameters to create an instance of the indicator
// based on lengths.
type KaufmanAdaptiveMovingAverageLengthParams struct {
	// EfficiencyRatioLength is the number of last samples used to calculate the efficiency ratio.
	//
	// The value should be greater than 1.
	// The default value is 10.
	EfficiencyRatioLength int

	// FastestLength is the fastest boundary length, ℓf.
	// The equivalent smoothing factor αf is
	//
	//   αf = 2/(ℓf + 1), 2 ≤ ℓ
	//
	// The value should be greater than 1.
	// The default value is 2.
	FastestLength int

	// SlowestLength is the slowest boundary length, ℓs.
	// The equivalent smoothing factor αs is
	//
	//   αs = 2/(ℓs + 1), 2 ≤ ℓ
	//
	// The value should be greater than 1.
	// The default value is 30.
	SlowestLength int

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

// KaufmanAdaptiveMovingAverageSmoothingFactorParams describes parameters to create an instance of the indicator
// based on smoothing factors.
type KaufmanAdaptiveMovingAverageSmoothingFactorParams struct {
	// EfficiencyRatioLength is the number of last samples used to calculate the efficiency ratio.
	//
	// The value should be greater than 1.
	// The default value is 10.
	EfficiencyRatioLength int

	// FastestSmoothingFactor is the fastest boundary smoothing factor, αf in (0,1).
	// The equivalent length ℓf is
	//
	//   ℓf = 2/αf - 1, 0 < αf ≤ 1, 1 ≤ ℓf
	//
	// The default value is 2/3 (0.6666...).
	FastestSmoothingFactor float64

	// SlowestSmoothingFactor is the slowest boundary smoothing factor, αs in (0,1).
	// The equivalent length ℓs is
	//
	//   ℓs = 2/αs - 1, 0 < αs ≤ 1, 1 ≤ ℓs
	//
	// The default value is 2/31 (0.06451612903225806451612903225806).
	SlowestSmoothingFactor float64

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
