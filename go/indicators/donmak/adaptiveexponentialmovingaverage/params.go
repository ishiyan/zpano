package adaptiveexponentialmovingaverage

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// AlphaMax is the smoothing factor for trending data (low frequency).
	//
	// The value should be in (0, 1] and greater than AlphaMin. The default value is 0.5.
	AlphaMax float64

	// AlphaMin is the smoothing factor for noisy data (high frequency).
	//
	// The value should be in (0, AlphaMax). The default value is 0.05.
	AlphaMin float64

	// Omega0 is the crossover frequency in radians/bar. Below this, alpha = AlphaMax.
	//
	// The value should be in (0, pi). The default value is 1.0.
	Omega0 float64

	// Smoothing is the embedded ISWP internal smoothing parameter.
	//
	// The value should be >= 0. The default value is 3.
	Smoothing int

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

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		AlphaMax:  0.5,
		AlphaMin:  0.05,
		Omega0:    1.0,
		Smoothing: 3,
	}
}
