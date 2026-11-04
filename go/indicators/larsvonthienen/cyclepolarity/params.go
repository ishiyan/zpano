// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclepolarity

import (
	"zpano/entities"
)

// Params configures the indicator; explicit zeros retain their numerical meaning.
type Params struct {
	// Window configures window; the reference default is 100.
	Window int
	// ShortPeriod configures short period; the reference default is 20.0.
	ShortPeriod float64
	// LongPeriod configures long period; the reference default is 40.0.
	LongPeriod float64
	// Smoothing configures smoothing; the reference default is 1600.0.
	Smoothing float64
	// Tolerance configures tolerance; the reference default is 0.0.
	Tolerance float64
	// BarComponent defaults to the entities default when zero and is omitted from the mnemonic.
	BarComponent entities.BarComponent
	// QuoteComponent defaults to the entities default when zero and is omitted from the mnemonic.
	QuoteComponent entities.QuoteComponent
	// TradeComponent defaults to the entities default when zero and is omitted from the mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns the prepared reference defaults.
func DefaultParams() *Params {
	return &Params{Window: 100, ShortPeriod: 20.0, LongPeriod: 40.0, Smoothing: 1600.0, Tolerance: 0.0}
}
