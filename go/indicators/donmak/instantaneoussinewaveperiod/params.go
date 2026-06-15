package instantaneoussinewaveperiod

import "zpano/entities"

// Params describes parameters to create an instance of the indicator.
type Params struct {
	// Smoothing is the EMA smoothing length applied to input prices before frequency estimation.
	//
	// The value should be >= 0 (0 means no smoothing). The default value is 0.
	Smoothing int

	// MinPeriod is the minimum allowed period in bars. Estimates below this are rejected.
	//
	// The value should be > 0. The default value is 4.0.
	MinPeriod float64

	// MaxPeriod is the maximum allowed period in bars. Estimates above this are rejected.
	//
	// The value should be > MinPeriod. The default value is 50.0.
	MaxPeriod float64

	// ErrorThreshold is the maximum tolerated error for the omega estimate.
	// If both methods exceed this, the output is NaN.
	//
	// The value should be > 0. The default value is 20.0.
	ErrorThreshold float64

	// Dx is the assumed measurement error for each price point (used in error propagation).
	//
	// The value should be > 0. The default value is 0.01.
	Dx float64

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
		Smoothing:      0,
		MinPeriod:      4.0,
		MaxPeriod:      50.0,
		ErrorThreshold: 20.0,
		Dx:             0.01,
	}
}
