package centerofgravityoscillator

import (
	"fmt"

	"zpano/entities"
	"zpano/indicators/core"
)

// Params describes parameters to create an instance of the CenterOfGravityOscillator indicator.
//
// Explicit zero values are NOT substituted with defaults; use [DefaultParams]
// to obtain Ehlers' default settings.
type Params struct {
	// Length is the length, ℓ, (the number of time periods) of the Center of Gravity oscillator.
	//
	// The value should be a positive integer, greater or equal to 1.
	// The default value used by Ehlers is 10.
	Length int

	// Centered selects the zero-centered baseline from Chapter 5 of Ehlers' 2004 book
	// (-Num/Denom + (Length+1)/2). When false (default), the May 2002 article / Mesa
	// baseline (-Num/Denom) is used.
	Centered bool

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, DefaultBarComponent (Close) is used and omitted from the indicator mnemonic.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	//
	// If zero, DefaultQuoteComponent (Mid) is used and omitted from the indicator mnemonic.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	//
	// If zero, DefaultTradeComponent (Price) is used and omitted from the indicator mnemonic.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns a [Params] value populated with Ehlers defaults.
func DefaultParams() *Params {
	return &Params{
		Length:   10,
		Centered: false,
	}
}

func formatMnemonic(prefix string, params Params) string {
	barComp := params.BarComponent
	if barComp == 0 {
		barComp = entities.DefaultBarComponent
	}
	quoteComp := params.QuoteComponent
	if quoteComp == 0 {
		quoteComp = entities.DefaultQuoteComponent
	}
	tradeComp := params.TradeComponent
	if tradeComp == 0 {
		tradeComp = entities.DefaultTradeComponent
	}
	centeredTag := ""
	if params.Centered {
		centeredTag = ", centered"
	}
	suffix := core.ComponentTripleMnemonic(barComp, quoteComp, tradeComp)
	return fmt.Sprintf("%s(%d%s%s)", prefix, params.Length, centeredTag, suffix)
}

// Mnemonic returns the primary indicator mnemonic for p.
func (p Params) Mnemonic() string {
	return formatMnemonic("cog", p)
}

// TriggerMnemonic returns the trigger line mnemonic for p.
func (p Params) TriggerMnemonic() string {
	return formatMnemonic("cogTrig", p)
}
