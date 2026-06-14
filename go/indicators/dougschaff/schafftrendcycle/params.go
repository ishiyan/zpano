package schafftrendcycle

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Fast is the number of periods for the fast EMA of the MACD line.
	//
	// The value should be greater than 0. The default value is 23.
	Fast int

	// Slow is the number of periods for the slow EMA of the MACD line.
	// It also sets the warm-up gate (barindex > slow).
	//
	// The value should be greater than 0. The default value is 50.
	Slow int

	// Tclen is the cycle length — the look-back for both stochastics.
	//
	// The value should be greater than 0. The default value is 10.
	Tclen int

	// Factor is the EMA smoothing alpha for both %D stages.
	//
	// The value should be in (0, 1]. The default value is 0.5.
	Factor float64

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
		Fast:   23,
		Slow:   50,
		Tclen:  10,
		Factor: 0.5,
	}
}
