package advancedecline

// AdvanceDeclineParams describes parameters to create an instance of the indicator.
// Advance-Decline requires HLCV bar data and has no configurable parameters.
type AdvanceDeclineParams struct{}

// DefaultParams returns a [AdvanceDeclineParams] value populated with conventional defaults.
func DefaultParams() *AdvanceDeclineParams { return &AdvanceDeclineParams{} }
