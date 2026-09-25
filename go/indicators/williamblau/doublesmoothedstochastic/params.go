package doublesmoothedstochastic

// Params describes parameters to create an instance of the indicator.
//
// The field names Q, R, S and G are the canonical symbols from William Blau's
// Momentum, Direction, and Divergence (Wiley, 1995). They are kept verbatim for
// fidelity with the book, the MQL5 reference, and the test-data naming.
//
// The indicator consumes the high, low and close prices of a bar, so it has no
// configurable price-component fields.
type Params struct {
	// Q is the stochastic look-back period: the number of bars over which the
	// highest high and the lowest low are taken.
	//
	// Setting q=1 yields the book's one-bar HLC index. The value should be
	// greater than 0. The default value is 5.
	Q int

	// R is the period of the 1st (inner) EMA in the smoothing cascade, applied
	// to the raw stochastic and the range.
	//
	// The value should be greater than 0. The default value is 7.
	R int

	// S is the period of the 2nd (outer) EMA in the smoothing cascade, applied
	// to the output of the 1st EMA.
	//
	// The value should be greater than 0. The default value is 3.
	S int

	// G is the period of the signal-line SMA, applied to the oscillator to
	// produce the second output.
	//
	// Setting g=1 makes the signal a passthrough (signal == dss every bar).
	// The value should be greater than 0. The default value is 3.
	G int
}

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		Q: 5,
		R: 7,
		S: 3,
		G: 3,
	}
}
