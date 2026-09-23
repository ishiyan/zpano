package stochasticmomentumindex

// Params describes parameters to create an instance of the indicator.
//
// The field names Q, R, S, U and Ul are the canonical symbols from William
// Blau's Momentum, Direction, and Divergence (Wiley, 1995), chapter 3. They are
// kept verbatim for fidelity with the book, the MQL5 reference, and the
// test-data naming.
//
// The indicator consumes the high, low and close prices of a bar, so it has no
// configurable price-component fields.
type Params struct {
	// Q is the stochastic look-back period: the number of bars over which the
	// highest high and the lowest low are taken.
	//
	// Setting q=1 yields the book's one-day stochastic. The value should be
	// greater than 0. The default value is 5.
	Q int

	// R is the period of the 1st (innermost) EMA in the smoothing cascade,
	// applied to the stochastic momentum and the half-range.
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
	// classic double-smoothed oscillator. The value should be greater than 0.
	// The default value is 3.
	U int

	// Ul is the period of the signal-line EMA, applied to the oscillator to
	// produce the second output.
	//
	// Setting ul=1 makes the signal a passthrough (signal == smi every bar).
	// The value should be greater than 0. The default value is 3.
	Ul int
}

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		Q:  5,
		R:  20,
		S:  5,
		U:  3,
		Ul: 3,
	}
}
