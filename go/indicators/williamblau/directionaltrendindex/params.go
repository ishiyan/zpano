package directionaltrendindex

// Params describes parameters to create an instance of the indicator.
//
// The field names Q, R, S, U and Ul are the canonical symbols from William
// Blau's Momentum, Direction, and Divergence (Wiley, 1995). They are kept
// verbatim for fidelity with the book, the MQL5 reference, and the test-data
// naming.
//
// The indicator consumes the high and low prices of a bar, so it has no
// configurable price-component fields.
type Params struct {
	// Q is the high-low momentum look-back period; the high and the low are
	// compared with their values q-1 bars ago.
	//
	// The look-back distance is q-1 bars, so q=2 is the one-bar form Blau uses
	// in the book. The value should be greater than 0 (q >= 2 is meaningful).
	// The default value is 2.
	Q int

	// R is the period of the 1st (innermost) EMA in the smoothing cascade,
	// applied to the high-low momentum.
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
	// Setting u=1 switches the 3rd stage off (passthrough), yielding a
	// double-smoothed DTI. The value should be greater than 0. The default
	// value is 3.
	U int

	// Ul is the period of the signal-line EMA, applied to the oscillator to
	// produce the second output (Blau's Ergodic signal line).
	//
	// Setting ul=1 makes the signal a passthrough (signal == dti every bar).
	// The value should be greater than 0. The default value is 3.
	Ul int
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
