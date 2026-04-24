package pearsonscorrelationcoefficient

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Length is the number of time periods in the rolling window.
	//
	// The value should be greater than 0.
	Length int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarHighPrice) is used for the X series
	// and BarLowPrice is used for the Y series.
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

// DefaultParams returns a Params value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		Length: 20,
	}
}
