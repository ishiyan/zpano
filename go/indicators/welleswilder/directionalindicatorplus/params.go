package directionalindicatorplus

// DirectionalIndicatorPlusParams describes parameters to create an instance of the indicator.
type DirectionalIndicatorPlusParams struct {
	// Length is the smoothing length (the number of time periods). Must be >= 1. The default value is 14.
	Length int
}

// DefaultParams returns a [DirectionalIndicatorPlusParams] value populated with conventional defaults.
func DefaultParams() *DirectionalIndicatorPlusParams {
	return &DirectionalIndicatorPlusParams{
		Length: 14,
	}
}
