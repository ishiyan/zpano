package macdindex

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
//
// The field names R, S, U and Ul are the canonical symbols from William Blau's
// Momentum, Direction, and Divergence (Wiley, 1995), chapter 5. They are kept
// verbatim for fidelity with the book, the MQL5 reference, and the test-data
// naming.
type Params struct {
	// R is the period of the slow EMA in the MACD line (EMA(close, s) - EMA(close, r)).
	//
	// The value should be greater than 0 and strictly greater than S. The
	// default value is 20.
	R int

	// S is the period of the fast EMA in the MACD line.
	//
	// The value should be greater than 0 and strictly less than R -- the MACD
	// line is fast minus slow, so the fast period must be shorter. The
	// default value is 5.
	S int

	// U is the period of the smoothing EMA applied to the MACD line.
	//
	// Setting u=1 switches this stage off (passthrough), yielding the book's
	// pure two-EMA MACD line EMA(close, s) - EMA(close, r). The value should
	// be greater than 0. The default value is 3.
	U int

	// Ul is the period of the signal-line EMA, applied to the index to produce
	// the second output (Blau's Ergodic signal line).
	//
	// Setting ul=1 makes the signal a passthrough (signal == macdi every bar).
	// The value should be greater than 0. This parameter is not shown in the
	// indicator mnemonic. The default value is 3.
	Ul int

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
		R:  20,
		S:  5,
		U:  3,
		Ul: 3,
	}
}
