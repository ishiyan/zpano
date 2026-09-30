package slopedivergencetsifilter

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
//
// The field names Q, R, S, U, X and Y are the canonical symbols from William
// Blau's Momentum, Direction, and Divergence (Wiley, 1995), chapter 12 and
// Appendix B, Figure B-25. They are kept verbatim for fidelity with the book
// and the test-data naming.
type Params struct {
	// Q is the TSI momentum look-back period; momentum is C_k - C_(k-(q-1)).
	//
	// The look-back distance is q-1 bars, so q=2 is the one-bar momentum Blau
	// uses throughout the book. The value should be greater than 0. The default
	// value is 2.
	Q int

	// R is the period of the 1st (innermost) EMA of the TSI smoothing cascade.
	//
	// The value should be greater than 0. The default value is 32.
	R int

	// S is the period of the 2nd EMA of the TSI smoothing cascade.
	//
	// The value should be greater than 0. The default value is 32.
	S int

	// U is the period of the 3rd (outermost) EMA of the TSI smoothing cascade.
	//
	// Setting u=1 switches the 3rd stage off (passthrough), yielding the
	// double-smoothed TSI of the book's raw form (Fig. 12-1). The value should
	// be greater than 0. The default value is 7.
	U int

	// X is the period of the 1st EMA of the price reference DEMA(close, x, y).
	//
	// The value should be greater than 0. The default value is 32.
	X int

	// Y is the period of the 2nd EMA of the price reference DEMA(close, x, y).
	//
	// Setting y=1 makes the price reference a single EMA. The value should be
	// greater than 0. The default value is 7.
	Y int

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
		Q: 2,
		R: 32,
		S: 32,
		U: 7,
		X: 32,
		Y: 7,
	}
}
