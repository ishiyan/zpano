package polynomialforecast

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Degree is the polynomial degree for the local fit (uses degree+1 bars).
	//
	// The value should be >= 2. The default value is 3.
	Degree int

	// Order is the Taylor expansion order: 1 = velocity only (F1V), 2 = velocity + acceleration (F1VA).
	//
	// The value should be 1 or 2. The default value is 1.
	Order int

	// Smoothing is the EMA pre-smoothing period applied to price before fitting (0 = none).
	//
	// The value should be >= 0. The default value is 0.
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
		Degree: 3,
		Order:  1,
	}
}
