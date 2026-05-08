package jurikturningpointoscillator

import "zpano/entities"

// JurikTurningPointOscillatorParams describes the parameters of the indicator.
type JurikTurningPointOscillatorParams struct {
	// Length controls the lookback window for the Spearman rank correlation.
	Length int `json:"length"`

	// BarComponent specifies the bar component to use.
	BarComponent entities.BarComponent `json:"bar_component"`

	// QuoteComponent specifies the quote component to use.
	QuoteComponent entities.QuoteComponent `json:"quote_component"`

	// TradeComponent specifies the trade component to use.
	TradeComponent entities.TradeComponent `json:"trade_component"`
}

// DefaultParams returns default parameters for the indicator.
func DefaultParams() *JurikTurningPointOscillatorParams {
	return &JurikTurningPointOscillatorParams{
		Length: 14,
	}
}
