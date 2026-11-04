// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclicsmoothedrelativestrengthindex

import (
	"zpano/entities"
)

// Params configures the indicator; explicit zeros retain their numerical meaning.
type Params struct {
	// DominantCycle configures dominant cycle; the reference default is 20.
	DominantCycle int
	// Vibration configures vibration; the reference default is 10.
	Vibration int
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
func DefaultParams() *Params { return &Params{DominantCycle: 20, Vibration: 10, Leveling: 10.0} }
