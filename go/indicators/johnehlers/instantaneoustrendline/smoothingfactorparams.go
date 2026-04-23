package instantaneoustrendline

import "zpano/entities"

// SmoothingFactorParams describes parameters to create an instance of the Instantaneous Trend Line indicator
// based on smoothing factor.
type SmoothingFactorParams struct {
	// SmoothingFactor is the smoothing factor, α, of the instantaneous trend line.
	// The equivalent length ℓ is:
	//
	//	ℓ = 2/α - 1, 0 < α ≤ 1, 1 ≤ ℓ
	//
	// The default value used in the Ehler's book is 0.07.
	SmoothingFactor float64

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarMedianPrice) is used and the component is not shown in the indicator mnemonic.
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
