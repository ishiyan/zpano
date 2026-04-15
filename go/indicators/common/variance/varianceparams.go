package variance

import "zpano/entities"

// VarianceParams describes parameters to create an instance of the indicator.
type VarianceParams struct {
	// Length is the length (the number of time periods, ℓ) of the moving window to calculate the variance.
	//
	// The value should be greater than 1.
	Length int

	// IsUnbiased indicates whether the estimate of the variance is the unbiased sample variance or
	// the population variance.
	//
	// When in doubt, use the unbiased sample variance (value is true).
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
