package jurikdirectionalmovementindex

// JurikDirectionalMovementIndexParams describes parameters to create an instance of the indicator.
type JurikDirectionalMovementIndexParams struct {
	// Length is the smoothing length parameter for the internal JMA instances.
	//
	// The value should be greater than 0. Typical values range from 2 to 20.
	Length int
}

// DefaultParams returns a [JurikDirectionalMovementIndexParams] value populated with conventional defaults.
func DefaultParams() *JurikDirectionalMovementIndexParams {
	return &JurikDirectionalMovementIndexParams{
		Length: 14,
	}
}
