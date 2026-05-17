package rescaledfractaladaptivesimplemovingaverage

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Period is the lookback window for R/S analysis. Must be a power of 2, >= 4.
	Period int

	// NormalSpeed is the base SMA period before fractal adaptation. Must be >= 1.
	NormalSpeed int

	// PriceScale is the multiplier applied to prices before R/S calculation.
	// Originally named PIP_Convertor. Default is 1.0.
	PriceScale float64

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
		Period:      64,
		NormalSpeed: 30,
		PriceScale:  1.0,
	}
}
