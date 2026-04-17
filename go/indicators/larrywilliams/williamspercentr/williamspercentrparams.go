package williamspercentr

// WilliamsPercentRParams describes parameters to create an instance of the indicator.
type WilliamsPercentRParams struct {
	// Length is the number of time periods. The typical values are 5, 9 or 14. The default is 14. Must be >= 2.
	Length int
}
