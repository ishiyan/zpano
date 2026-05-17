package directionalmovementminus

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Length is the smoothing length (the number of time periods). Must be >= 1. The default value is 14.
	Length int
}

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		Length: 14,
	}
}
