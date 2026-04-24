package aroon

// AroonParams describes parameters to create an instance of the indicator.
type AroonParams struct {
	// Length is the lookback period for the Aroon calculation.
	//
	// The value should be greater than 1. The default value is 14.
	Length int
}

// DefaultParams returns a [AroonParams] value populated with conventional defaults.
func DefaultParams() *AroonParams {
	return &AroonParams{
		Length: 14,
	}
}
