package polynomialfitderivative

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Degree is the polynomial degree. The number of data points used is Degree + 1.
	//
	// The value should be >= 2. The default value is 3 (cubic).
	Degree int

	// Order is the derivative order (1 = velocity, 2 = acceleration).
	//
	// The value should be >= 1 and <= Degree. The default value is 1.
	Order int

	// Smoothing is the EMA pre-smoothing length applied before the FIR filter.
	//
	// The value should be >= 0 (0 means no smoothing). The default value is 6.
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
		Degree:    3,
		Order:     1,
		Smoothing: 6,
	}
}
