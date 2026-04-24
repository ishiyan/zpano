package directionalmovementplus

// DirectionalMovementPlusParams describes parameters to create an instance of the indicator.
type DirectionalMovementPlusParams struct {
	// Length is the smoothing length (the number of time periods). Must be >= 1. The default value is 14.
	Length int
}

// DefaultParams returns a [DirectionalMovementPlusParams] value populated with conventional defaults.
func DefaultParams() *DirectionalMovementPlusParams {
	return &DirectionalMovementPlusParams{
		Length: 14,
	}
}
