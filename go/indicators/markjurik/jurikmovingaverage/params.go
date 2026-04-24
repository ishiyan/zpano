package jurikmovingaverage

import "zpano/entities" //nolint:depguard

// JurikMovingAverageParams describes parameters to create an instance of the indicator.
type JurikMovingAverageParams struct {
	// Length is the length (the number of time periods, ℓ) determines
	// the degree of smoothness and it can be any positive value.
	//
	// Small values make the moving average respond rapidly to price change
	// and larger values produce smoother, flatter curves.
	//
	// The value should be greater than 1. Typical values range from 5 to 80.
	//
	// Irrespective from the value, the indicator needs at 30 first values to be primed.
	Length int

	// Phase affects the amount of lag (delay).
	// Lower lag tends to produce larger overshoot during price gaps, so you need
	// to consider the trade-off between lag and overshoot and select a value for
	// phase that balances your trading system's needs.
	//
	// The phase values should be in [-100, 100].
	//
	// - The value of -100 results in maximum lag and no overshoot.
	//
	// - The value of 0 results in some lag and some overshoot.
	//
	// - The value of 100 results in minimum lag and maximum overshoot.
	Phase int

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

// DefaultParams returns a [JurikMovingAverageParams] value populated with conventional defaults.
func DefaultParams() *JurikMovingAverageParams {
	return &JurikMovingAverageParams{
		Length: 14,
		Phase:  0,
	}
}
