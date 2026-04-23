package rateofchangeratio

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Length is the length (the number of time periods, ℓ) between today's sample and the sample ℓ periods ago.
	//
	// The value should be greater than 0.
	Length int

	// HundredScale indicates whether to multiply the ratio by 100.
	//
	// If false, the result is price/previousPrice (centered at 1).
	// If true, the result is (price/previousPrice)*100 (centered at 100).
	HundredScale bool

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarClosePrice) is used and the component is not shown in the indicator mnemonic.
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
