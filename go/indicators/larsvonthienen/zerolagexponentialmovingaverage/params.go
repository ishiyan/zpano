// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package zerolagexponentialmovingaverage

import (
	"zpano/entities"
)

// Params configures the indicator; explicit zeros retain their numerical meaning.
type Params struct {
	// Length configures length; the reference default is 13.
	Length int
	// BarComponent defaults to the entities default when zero and is omitted from the mnemonic.
	BarComponent entities.BarComponent
	// QuoteComponent defaults to the entities default when zero and is omitted from the mnemonic.
	QuoteComponent entities.QuoteComponent
	// TradeComponent defaults to the entities default when zero and is omitted from the mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns the prepared reference defaults.
func DefaultParams() *Params { return &Params{Length: 13} }
