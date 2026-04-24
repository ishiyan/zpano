package directionalindicatorminus

// DirectionalIndicatorMinusParams describes parameters to create an instance of the indicator.
type DirectionalIndicatorMinusParams struct {
	// Length is the smoothing length (the number of time periods). Must be >= 1. The default value is 14.
	Length int
}

// DefaultParams returns a [DirectionalIndicatorMinusParams] value populated with conventional defaults.
func DefaultParams() *DirectionalIndicatorMinusParams {
	return &DirectionalIndicatorMinusParams{
		Length: 14,
	}
}
