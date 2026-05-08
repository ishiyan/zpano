package arnaudlegouxmovingaverage

import "zpano/entities"

// ArnaudLegouxMovingAverageParams describes parameters to create an instance of the indicator.
type ArnaudLegouxMovingAverageParams struct {
	// Window is the number of bars in the lookback window.
	//
	// The value should be greater than 0.
	Window int

	// Sigma controls the Gaussian width; larger values produce smoother output.
	//
	// The value should be greater than 0.
	Sigma float64

	// Offset shifts the Gaussian peak; 0 = centered, 1 = newest bar.
	//
	// The value should be between 0 and 1.
	Offset float64

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

// DefaultParams returns an ArnaudLegouxMovingAverageParams value populated with conventional defaults.
func DefaultParams() *ArnaudLegouxMovingAverageParams {
	return &ArnaudLegouxMovingAverageParams{
		Window: 9,
		Sigma:  6.0,
		Offset: 0.85,
	}
}
