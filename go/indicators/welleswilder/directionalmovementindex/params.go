package directionalmovementindex

// DirectionalMovementIndexParams describes parameters to create an instance of the indicator.
type DirectionalMovementIndexParams struct {
	// Length is the smoothing length (the number of time periods). Must be >= 1. The default value is 14.
	Length int
}

// DefaultParams returns a [DirectionalMovementIndexParams] value populated with conventional defaults.
func DefaultParams() *DirectionalMovementIndexParams {
	return &DirectionalMovementIndexParams{
		Length: 14,
	}
}
