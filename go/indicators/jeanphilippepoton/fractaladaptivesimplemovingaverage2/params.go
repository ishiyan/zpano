package fractaladaptivesimplemovingaverage2

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Period is the lookback period N for the FGDI computation.
	//
	// The value should be greater than 1.
	Period int

	// NormalSpeed is the base SMA period before fractal adaptation.
	//
	// The value should be greater than 0.
	NormalSpeed int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns a Params value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		Period:      30,
		NormalSpeed: 20,
	}
}
