package movingminimax

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// M is the smoothing window width controlling the quantum tunnelling ability.
	//
	// Larger values produce smoother output, suppressing smaller peaks. The value should be
	// >= 1. The default value is 5.
	M int

	// N is the lookback window size: the number of price bars over which the indicator is computed.
	//
	// Priming requires N prices. The value should be > 2*M. The default value is 50.
	N int

	// NumExtrema is the number of distinct support/resistance levels to detect and return.
	//
	// The value should be >= 1. The default value is 3.
	NumExtrema int

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

// DefaultParams returns a [Params] value populated with conventional defaults.
func DefaultParams() *Params {
	return &Params{
		M:          5,
		N:          50,
		NumExtrema: 3,
	}
}
