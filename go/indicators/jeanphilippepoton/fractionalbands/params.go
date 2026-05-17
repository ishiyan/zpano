package fractionalbands

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Period is the lookback period for the FDI computation.
	//
	// The value should be greater than 1.
	Period int

	// PriceScale is the multiplier converting price to a working numeric space
	// for the deviation exponentiation.
	//
	// The value should be greater than 0.
	PriceScale float64

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

// DefaultParams returns a Params value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		Period:     30,
		PriceScale: 1.0,
	}
}
