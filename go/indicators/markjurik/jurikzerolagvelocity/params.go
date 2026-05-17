package jurikzerolagvelocity

import "zpano/entities"

// Params describes the parameters of the indicator.
type Params struct {
	// Depth controls the linear regression window (window = Depth+1).
	Depth int `json:"depth"`

	// BarComponent specifies the bar component to use.
	BarComponent entities.BarComponent `json:"bar_component"`

	// QuoteComponent specifies the quote component to use.
	QuoteComponent entities.QuoteComponent `json:"quote_component"`

	// TradeComponent specifies the trade component to use.
	TradeComponent entities.TradeComponent `json:"trade_component"`
}

// DefaultParams returns default parameters for the indicator.
func DefaultParams() *Params {
	return &Params{
		Depth: 10,
	}
}
