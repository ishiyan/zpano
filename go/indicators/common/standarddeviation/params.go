package standarddeviation

import "zpano/entities"

// StandardDeviationParams describes parameters to create an instance of the indicator.
type StandardDeviationParams struct {
	// Length is the length (the number of time periods, ℓ) of the moving window to calculate the standard deviation.
	//
	// The value should be greater than 1.
	Length int

	// IsUnbiased indicates whether the estimate of the standard deviation is the unbiased sample standard deviation
	// or the population standard deviation.
	//
	// When in doubt, use the unbiased sample standard deviation (value is true).
	IsUnbiased bool

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

// DefaultParams returns a StandardDeviationParams value populated with conventional defaults.
func DefaultParams() *StandardDeviationParams {
	return &StandardDeviationParams{
		Length:     20,
		IsUnbiased: true,
	}
}
