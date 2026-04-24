package tripleexponentialmovingaverageoscillator

import "zpano/entities"

// TripleExponentialMovingAverageOscillatorParams describes parameters to create an instance of the indicator.
type TripleExponentialMovingAverageOscillatorParams struct {
	// Length is the number of time periods for the three chained EMA calculations.
	//
	// The value should be greater than or equal to 1. The default value is 30.
	Length int

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

// DefaultParams returns a [TripleExponentialMovingAverageOscillatorParams] value populated with conventional defaults.
func DefaultParams() *TripleExponentialMovingAverageOscillatorParams {
	return &TripleExponentialMovingAverageOscillatorParams{
		Length: 30,
	}
}
