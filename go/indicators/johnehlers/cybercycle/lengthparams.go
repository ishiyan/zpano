package cybercycle

import "zpano/entities"

// LengthParams describes parameters to create an instance of the Cyber Cycle indicator
// based on length.
type LengthParams struct {
	// Length is the length, ℓ, of the cyber cycle.
	// The equivalent smoothing factor α is:
	//
	//	α = 2/(ℓ + 1), 0 < α ≤ 1, 1 ≤ ℓ
	//
	// The value should be a positive integer, greater or equal to 1.
	Length int

	// SignalLag is the lag, ℓ, of the signal line, which is an exponential moving average
	// of the cyber cycle value. The equivalent EMA smoothing factor α is:
	//
	//	α = 1/(ℓ + 1), 0 < α ≤ 1, 0 ≤ ℓ
	//
	// The value should be a positive integer, greater or equal to 1.
	// There are two default values used in the Ehler's book: 9 (EasyLanguage code) and 20 (EFS code).
	SignalLag int

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

// DefaultLengthParams returns a [LengthParams] value populated with Ehlers defaults.
func DefaultLengthParams() *LengthParams {
	return &LengthParams{
		Length:    28,
		SignalLag: 9,
	}
}
