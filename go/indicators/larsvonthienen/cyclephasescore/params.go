// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclephasescore

import (
	"zpano/entities"
)

// Params configures the indicator; explicit zeros retain their numerical meaning.
type Params struct {
	// Window configures window; the reference default is 200.
	Window int
	// MinPeriod configures min period; the reference default is 5.
	MinPeriod int
	// MaxPeriod configures max period; the reference default is 80.
	MaxPeriod int
	// Smoothing configures smoothing; the reference default is 1600.0.
	Smoothing float64
	// GenuineThreshold configures genuine threshold; the reference default is 49.0.
	GenuineThreshold float64
	// PhaseCycles configures phase cycles; the reference default is 2.0.
	PhaseCycles float64
	// TopK configures top k; the reference default is 3.
	TopK int
	// PeakDistance configures peak distance; the reference default is 2.
	PeakDistance int
	// MaxSegments configures max segments; the reference default is 5.
	MaxSegments int
	// BarComponent defaults to the entities default when zero and is omitted from the mnemonic.
	BarComponent entities.BarComponent
	// QuoteComponent defaults to the entities default when zero and is omitted from the mnemonic.
	QuoteComponent entities.QuoteComponent
	// TradeComponent defaults to the entities default when zero and is omitted from the mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns the prepared reference defaults.
func DefaultParams() *Params {
	return &Params{Window: 200, MinPeriod: 5, MaxPeriod: 80, Smoothing: 1600.0, GenuineThreshold: 49.0, PhaseCycles: 2.0, TopK: 3, PeakDistance: 2, MaxSegments: 5}
}
