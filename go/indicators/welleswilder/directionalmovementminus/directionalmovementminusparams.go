package directionalmovementminus

// DirectionalMovementMinusParams describes parameters to create an instance of the indicator.
type DirectionalMovementMinusParams struct {
	// Length is the smoothing length (the number of time periods). Must be >= 1. The default value is 14.
	Length int
}
