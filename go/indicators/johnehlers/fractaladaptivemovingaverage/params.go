package fractaladaptivemovingaverage

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Length is the length, ℓ, (the number of time periods) of the Fractal Adaptive Moving Average.
	//
	// The value should be an even integer, greater or equal to 2.
	// The default value is 16.
	Length int

	// SlowestSmoothingFactor is the slowest boundary smoothing factor, αs in [0,1].
	// The equivalent length ℓs is
	//
	//   ℓs = 2/αs - 1, 0 < αs ≤ 1, 1 ≤ ℓs
	//
	// The default value is 0.01 (equivalent ℓs = 199).
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

// DefaultParams returns a [Params] value populated with Ehlers defaults.
func DefaultParams() *Params {
	return &Params{
		Length:                 16,
		SlowestSmoothingFactor: 0.01,
	}
}
