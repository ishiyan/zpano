package tickvolumeindicator

// Params describes parameters to create an instance of the indicator.
//
// The field names R, S and U are the canonical symbols from William Blau's
// Momentum, Direction, and Divergence (Wiley, 1995), chapters 4 and 10. They
// are kept verbatim for fidelity with the book and the test-data naming.
//
// The indicator consumes upticks and downticks (derived from the bar range,
// or from consecutive prices for single-valued samples), so it has no
// configurable price-component fields.
type Params struct {
	// R is the period of the 1st (innermost) EMA in the smoothing cascade,
	// applied to the upticks and downticks.
	//
	// The value should be greater than 0. The default value is 12.
	R int

	// S is the period of the 2nd EMA in the smoothing cascade, applied to the
	// output of the 1st EMA.
	//
	// The value should be greater than 0. The default value is 12.
	S int

	// U is the period of the 3rd (outermost) EMA in the smoothing cascade,
	// applied to the output of the 2nd EMA.
	//
	// The default u=1 switches the 3rd stage off (passthrough), yielding the
	// book's classic double-smoothed TVI(r, s); chapter 10 uses TVI(32, 32, 5).
	// The value should be greater than 0. The default value is 1.
	U int
}

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		R: 12,
		S: 12,
		U: 1,
	}
}
