// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package goertzelcyclespectrum

import (
	"zpano/entities"
)

// Params configures the indicator; explicit zeros retain their numerical meaning.
type Params struct {
	// Window configures window; the reference default is 128.
	Window int
	// MinPeriod configures min period; the reference default is 5.
	MinPeriod int
	// MaxPeriod configures max period; the reference default is 50.
	MaxPeriod int
	// Step configures step; the reference default is 1.
	Step int
	// Smoothing configures smoothing; the reference default is 1600.0.
	Smoothing float64
	// BarComponent defaults to the entities default when zero and is omitted from the mnemonic.
	BarComponent entities.BarComponent
	// QuoteComponent defaults to the entities default when zero and is omitted from the mnemonic.
	QuoteComponent entities.QuoteComponent
	// TradeComponent defaults to the entities default when zero and is omitted from the mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns the prepared reference defaults.
func DefaultParams() *Params {
	return &Params{Window: 128, MinPeriod: 5, MaxPeriod: 50, Step: 1, Smoothing: 1600.0}
}
