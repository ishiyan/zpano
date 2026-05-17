package aroon

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Length is the lookback period for the Aroon calculation.
	//
	// The value should be greater than 1. The default value is 14.
	Length int
}

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		Length: 14,
	}
}
