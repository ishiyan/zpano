package normalizedaveragetruerange

// NormalizedAverageTrueRangeParams describes parameters to create an instance of the indicator.
type NormalizedAverageTrueRangeParams struct {
	// Length is the number of time periods. Must be >= 1. The default value is 14.
	Length int
}

// DefaultParams returns a [NormalizedAverageTrueRangeParams] value populated with conventional defaults.
func DefaultParams() *NormalizedAverageTrueRangeParams {
	return &NormalizedAverageTrueRangeParams{
		Length: 14,
	}
}
