package advancedecline

// Params describes parameters to create an instance of the indicator.
// Advance-Decline requires HLCV bar data and has no configurable parameters.
type Params struct{}

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params { return &Params{} }
