package truerange

// Params describes parameters to create an instance of the indicator.
//
// The True Range indicator has no configurable parameters.
// This struct exists for consistency with other indicators.
type Params struct{}

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{}
}
