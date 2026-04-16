package supersmoother

import "zpano/entities"

// SuperSmootherParams describes parameters to create an instance of the indicator.
type SuperSmootherParams struct {
	// ShortestCyclePeriod is the shortest cycle period in bars.
	// The Super Smoother attenuates all cycle periods shorter than this one.
	//
	// The value should be greater than 1. The default value is 10.
	ShortestCyclePeriod int

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
