package truestrengthindex

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
//
// The field names Q, R, S, U and Ul are the canonical symbols from William
// Blau's Momentum, Direction, and Divergence (Wiley, 1995), chapter 2. They are
// kept verbatim for fidelity with the book, the MQL5 reference, and the
// test-data naming.
type Params struct {
	// Q is the momentum look-back period; momentum is C_k - C_(k-(q-1)).
	//
	// The look-back distance is q-1 bars, so q=2 is the one-bar momentum Blau
	// uses throughout the book. The value should be greater than 0 (q >= 2 is
	// meaningful). The default value is 2.
	Q int

	// R is the period of the 1st (innermost) EMA in the smoothing cascade,
	// applied to the momentum.
	//
	// The value should be greater than 0. The default value is 20.
	R int

	// S is the period of the 2nd EMA in the smoothing cascade, applied to the
	// output of the 1st EMA.
	//
	// The value should be greater than 0. The default value is 5.
	S int

	// U is the period of the 3rd (outermost) EMA in the smoothing cascade,
	// applied to the output of the 2nd EMA.
	//
	// Setting u=1 switches the 3rd stage off (passthrough), yielding the book's
	// classic double-smoothed TSI. The value should be greater than 0. The
	// default value is 3.
	U int

	// Ul is the period of the signal-line EMA, applied to the oscillator to
	// produce the second output (Blau's Ergodic signal line).
	//
	// Setting ul=1 makes the signal a passthrough (signal == tsi every bar).
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
		Q:  2,
		R:  20,
		S:  5,
		U:  3,
		Ul: 3,
	}
}
