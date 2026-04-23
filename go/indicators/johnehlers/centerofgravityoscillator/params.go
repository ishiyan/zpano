package centerofgravityoscillator

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Length is the length, ℓ, (the number of time periods) of the Center of Gravity oscillator.
	//
	// The value should be a positive integer, greater or equal to 1.
	// The default value used by Ehlers is 10.
	Length int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarMedianPrice) is used and the component is not shown in the indicator mnemonic.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	//
	// If zero, the default (QuoteMidPrice) is used and the component is not shown in the indicator mnemonic.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	//
	// If zero, the default (TradePrice) is used and the component is not shown in the indicator mnemonic.
	TradeComponent entities.TradeComponent
}
