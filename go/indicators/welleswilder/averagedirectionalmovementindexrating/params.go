package averagedirectionalmovementindexrating

// AverageDirectionalMovementIndexRatingParams describes parameters to create an instance of the indicator.
type AverageDirectionalMovementIndexRatingParams struct {
	// Length is the smoothing length (the number of time periods). Must be >= 1. The default value is 14.
	Length int
}

// DefaultParams returns a [AverageDirectionalMovementIndexRatingParams] value populated with conventional defaults.
func DefaultParams() *AverageDirectionalMovementIndexRatingParams {
	return &AverageDirectionalMovementIndexRatingParams{
		Length: 14,
	}
}
