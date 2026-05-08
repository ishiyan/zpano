package newmovingaverage

import "zpano/entities"

// NewMovingAverageParams holds the parameters for the NewMovingAverage indicator.
type NewMovingAverageParams struct {
	// PrimaryPeriod is the period for the primary (outer) moving average.
	//
	// If 0 or too small, it is auto-resolved via Nyquist constraint.
	PrimaryPeriod int `json:"primary_period"`

	// SecondaryPeriod is the period for the secondary (inner) moving average.
	//
	// The value should be greater than or equal to 2.
	SecondaryPeriod int `json:"secondary_period"`

	// MAType selects the moving average type (SMA=0, EMA=1, SMMA=2, LWMA=3).
	MAType MAType `json:"ma_type"`

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarClosePrice) is used and the component is not shown in the indicator mnemonic.
	BarComponent entities.BarComponent `json:"bar_component"`

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	//
	// If zero, the default (QuoteMidPrice) is used and the component is not shown in the indicator mnemonic.
	QuoteComponent entities.QuoteComponent `json:"quote_component"`

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	//
	// If zero, the default (TradePrice) is used and the component is not shown in the indicator mnemonic.
	TradeComponent entities.TradeComponent `json:"trade_component"`
}

// DefaultParams returns a NewMovingAverageParams value populated with conventional defaults.
func DefaultParams() *NewMovingAverageParams {
	return &NewMovingAverageParams{
		PrimaryPeriod:   0,
		SecondaryPeriod: 8,
		MAType:          LWMA,
	}
}
