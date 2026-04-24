package ultimateoscillator

// UltimateOscillatorParams describes parameters to create an instance of the indicator.
//
// Length1 is the first (shortest) period. Default is 7.
// Length2 is the second (medium) period. Default is 14.
// Length3 is the third (longest) period. Default is 28.
//
// All three lengths must be >= 2. If a length is 0, the default is used.
type UltimateOscillatorParams struct {
	Length1 int
	Length2 int
	Length3 int
}

// DefaultParams returns a [UltimateOscillatorParams] value populated with conventional defaults.
func DefaultParams() *UltimateOscillatorParams {
	return &UltimateOscillatorParams{
		Length1: 7,
		Length2: 14,
		Length3: 28,
	}
}
