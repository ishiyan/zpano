// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package bartelsperiodicityscore

import (
	"zpano/entities"
)

// Params configures the indicator; explicit zeros retain their numerical meaning.
type Params struct {
	// Period configures period; the reference default is 20.
	Period int
	// Segments configures segments; the reference default is 5.
	Segments int
	// Smoothing configures smoothing; the reference default is 0.0.
	Smoothing float64
	// BarComponent defaults to the entities default when zero and is omitted from the mnemonic.
	BarComponent entities.BarComponent
	// QuoteComponent defaults to the entities default when zero and is omitted from the mnemonic.
	QuoteComponent entities.QuoteComponent
	// TradeComponent defaults to the entities default when zero and is omitted from the mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns the prepared reference defaults.
func DefaultParams() *Params { return &Params{Period: 20, Segments: 5, Smoothing: 0.0} }
