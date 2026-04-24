package parabolicstopandreverse

// ParabolicStopAndReverseParams describes parameters to create an instance of the indicator.
//
// The Parabolic SAR Extended supports separate acceleration factor parameters for long
// and short directions. For the standard Parabolic SAR behavior, use the same values
// for both long and short parameters (which is the default).
type ParabolicStopAndReverseParams struct {
	// StartValue controls the initial direction and SAR value.
	//
	//  0  = Auto-detect direction using the first two bars (default).
	//  >0 = Force long at the specified SAR value.
	//  <0 = Force short at abs(StartValue) as the initial SAR value.
	//
	// Default is 0.0.
	StartValue float64

	// OffsetOnReverse is a percent offset added/removed to the initial stop
	// on short/long reversal.
	//
	// Default is 0.0.
	OffsetOnReverse float64

	// AccelerationInitLong is the initial acceleration factor for the long direction.
	//
	// Default is 0.02.
	AccelerationInitLong float64

	// AccelerationLong is the acceleration factor increment for the long direction.
	//
	// Default is 0.02.
	AccelerationLong float64

	// AccelerationMaxLong is the maximum acceleration factor for the long direction.
	//
	// Default is 0.20.
	AccelerationMaxLong float64

	// AccelerationInitShort is the initial acceleration factor for the short direction.
	//
	// Default is 0.02.
	AccelerationInitShort float64

	// AccelerationShort is the acceleration factor increment for the short direction.
	//
	// Default is 0.02.
	AccelerationShort float64

	// AccelerationMaxShort is the maximum acceleration factor for the short direction.
	//
	// Default is 0.20.
	AccelerationMaxShort float64
}

// DefaultParams returns a [ParabolicStopAndReverseParams] value populated with conventional defaults.
func DefaultParams() *ParabolicStopAndReverseParams {
	return &ParabolicStopAndReverseParams{
		StartValue:            0,
		OffsetOnReverse:       0,
		AccelerationInitLong:  0.02,
		AccelerationLong:      0.02,
		AccelerationMaxLong:   0.20,
		AccelerationInitShort: 0.02,
		AccelerationShort:     0.02,
		AccelerationMaxShort:  0.20,
	}
}
