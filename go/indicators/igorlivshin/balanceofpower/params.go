package balanceofpower

// Params describes parameters to create an instance of the indicator.
// Balance of Power requires OHLC bar data and has no configurable parameters.
type Params struct{}

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params { return &Params{} }
