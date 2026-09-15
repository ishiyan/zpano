package doublesmoothedmomenta

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
//
// The parameter names A, Y and Z are the canonical symbols from William Blau's
// double-smoothed momentum family. They are kept verbatim for fidelity with the
// catalog definition and the test-data naming.
type Params struct {
	// A is the highest/lowest close look-back.
	//
	// A=2 gives the one-bar momentum of the RSI family; A>2 gives a
	// double-smoothed stochastic of the close.
	//
	// The value should be greater than 0. The default value is 2.
	A int

	// Y is the period of the inner (1st) smoothing EMA of the cascade.
	//
	// Setting Y=1 makes the inner stage a passthrough, so DM(2,1,z) is the
	// EMA-form RSI(z).
	//
	// The value should be greater than 0. The default value is 2.
	Y int

	// Z is the period of the outer (2nd) smoothing EMA of the cascade.
	//
	// This is the dominant smoothing (the RSI-style default 14).
	//
	// The value should be greater than 0. The default value is 14.
	Z int

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
		A: 2,
		Y: 2,
		Z: 14,
	}
}
