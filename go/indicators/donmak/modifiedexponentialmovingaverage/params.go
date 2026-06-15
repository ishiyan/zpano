package modifiedexponentialmovingaverage

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Period is the EMA smoothing period.
	//
	// The value should be >= 2. The default value is 6.
	Period int

	// Degree is the polynomial degree for the velocity correction.
	//
	// The value should be >= 2. The default value is 3.
	Degree int

	// Skip is the stride for sampling the EMA history (1 = MEMA, >1 = MEMA-D).
	//
	// The value should be >= 1. The default value is 1.
	Skip int

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
		Period: 6,
		Degree: 3,
		Skip:   1,
	}
}
