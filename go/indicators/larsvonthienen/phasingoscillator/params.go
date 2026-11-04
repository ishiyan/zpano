// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package phasingoscillator

import (
	"zpano/entities"
)

// Params configures the indicator; explicit zeros retain their numerical meaning.
type Params struct {
	// Window configures window; the reference default is 200.
	Window int
	// ShortMin configures short min; the reference default is 5.
	ShortMin int
	// ShortMax configures short max; the reference default is 30.
	ShortMax int
	// LongMin configures long min; the reference default is 31.
	LongMin int
	// LongMax configures long max; the reference default is 80.
	LongMax int
	// Smoothing configures smoothing; the reference default is 1600.0.
	Smoothing float64
	// GenuineThreshold configures genuine threshold; the reference default is 49.0.
	GenuineThreshold float64
	// PhaseCycles configures phase cycles; the reference default is 2.0.
	PhaseCycles float64
	// BarComponent defaults to the entities default when zero and is omitted from the mnemonic.
	BarComponent entities.BarComponent
	// QuoteComponent defaults to the entities default when zero and is omitted from the mnemonic.
	QuoteComponent entities.QuoteComponent
	// TradeComponent defaults to the entities default when zero and is omitted from the mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns the prepared reference defaults.
func DefaultParams() *Params {
	return &Params{Window: 200, ShortMin: 5, ShortMax: 30, LongMin: 31, LongMax: 80, Smoothing: 1600.0, GenuineThreshold: 49.0, PhaseCycles: 2.0}
}
