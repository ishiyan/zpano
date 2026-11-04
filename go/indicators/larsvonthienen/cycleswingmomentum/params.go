// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cycleswingmomentum

import (
	"zpano/entities"
)

// Params configures the indicator; explicit zeros retain their numerical meaning.
type Params struct {
	// Window configures window; the reference default is 50.
	Window int
	// FastSmoothing configures fast smoothing; the reference default is 160.0.
	FastSmoothing float64
	// SlowSmoothing configures slow smoothing; the reference default is 1600.0.
	SlowSmoothing float64
	// BandMemory configures band memory; the reference default is 34.
	BandMemory int
	// Leveling configures leveling; the reference default is 10.0.
	Leveling float64
	// BarComponent defaults to the entities default when zero and is omitted from the mnemonic.
	BarComponent entities.BarComponent
	// QuoteComponent defaults to the entities default when zero and is omitted from the mnemonic.
	QuoteComponent entities.QuoteComponent
	// TradeComponent defaults to the entities default when zero and is omitted from the mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns the prepared reference defaults.
func DefaultParams() *Params {
	return &Params{Window: 50, FastSmoothing: 160.0, SlowSmoothing: 1600.0, BandMemory: 34, Leveling: 10.0}
}
