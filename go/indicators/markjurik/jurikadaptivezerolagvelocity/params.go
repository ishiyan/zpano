package jurikadaptivezerolagvelocity

import "zpano/entities"

// JurikAdaptiveZeroLagVelocityParams describes the parameters of the indicator.
type JurikAdaptiveZeroLagVelocityParams struct {
	// LoLength is the minimum adaptive depth.
	LoLength int `json:"lo_length"`

	// HiLength is the maximum adaptive depth.
	HiLength int `json:"hi_length"`

	// Sensitivity controls the volatility regime detection sensitivity.
	Sensitivity float64 `json:"sensitivity"`

	// Period controls the adaptive smoother period.
	Period float64 `json:"period"`

	// BarComponent specifies the bar component to use.
	BarComponent entities.BarComponent `json:"bar_component"`

	// QuoteComponent specifies the quote component to use.
	QuoteComponent entities.QuoteComponent `json:"quote_component"`

	// TradeComponent specifies the trade component to use.
	TradeComponent entities.TradeComponent `json:"trade_component"`
}

// DefaultParams returns default parameters for the indicator.
func DefaultParams() *JurikAdaptiveZeroLagVelocityParams {
	return &JurikAdaptiveZeroLagVelocityParams{
		LoLength:    5,
		HiLength:    30,
		Sensitivity: 1.0,
		Period:      3.0,
	}
}
