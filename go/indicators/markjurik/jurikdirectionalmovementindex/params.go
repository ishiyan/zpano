package jurikdirectionalmovementindex

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Length is the smoothing length parameter for the internal JMA instances.
	//
	// The value should be greater than 0. Typical values range from 2 to 20.
	Length int
}

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		Length: 14,
	}
}
