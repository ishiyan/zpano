package quantumpricelevels

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Lookback is the number of price-return ratios maintained in the sliding window.
	//
	// Priming requires Lookback+1 prices. The value should be >= 2. The default value is 2048.
	Lookback int

	// NumLevels is the number of quantum energy levels to compute (n = 0..NumLevels-1).
	//
	// The value should be >= 1. The default value is 21.
	NumLevels int

	// NumBins is the number of histogram bins for the wavefunction distribution.
	//
	// The value should be >= 2. The default value is 100.
	NumBins int

	// ScaleFactor is the empirical scaling constant in the NQPR formula (1 + ScaleFactor*sigma*QPR).
	//
	// The value should be > 0. The default value is 0.21.
	ScaleFactor float64

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
		Lookback:    2048,
		NumLevels:   21,
		NumBins:     100,
		ScaleFactor: 0.21,
	}
}
