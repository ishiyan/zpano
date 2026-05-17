package zerolagerrorcorrectingexponentialmovingaverage

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// SmoothingFactor is the smoothing factor (alpha) of the EMA.
	//
	// alpha = 2/(length + 1), 0 < alpha <= 1, 1 <= length.
	// The default value is 0.095 (equivalent to length 20).
	SmoothingFactor float64

	// GainLimit defines the range [-g, g] for finding the best gain factor.
	//
	// The value should be positive. The default value is 5.
	GainLimit float64

	// GainStep defines the iteration step for finding the best gain factor.
	//
	// The value should be positive. The default value is 0.1.
	GainStep float64

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

// DefaultParams returns a [Params] value populated with Ehlers defaults.
func DefaultParams() *Params {
	return &Params{
		SmoothingFactor: 0.095,
		GainLimit:       5,
		GainStep:        0.1,
	}
}
