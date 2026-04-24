package triangularmovingaverage

import "zpano/entities"

// TriangularMovingAverageParams describes parameters to create an instance of the indicator.
type TriangularMovingAverageParams struct {
	// Length is the length (the number of time periods, ℓ) of the moving window to calculate the average.
	//
	// The value should be greater than 1.
	Length int

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

// DefaultParams returns a TriangularMovingAverageParams value populated with conventional defaults.
func DefaultParams() *TriangularMovingAverageParams {
	return &TriangularMovingAverageParams{
		Length: 20,
	}
}
