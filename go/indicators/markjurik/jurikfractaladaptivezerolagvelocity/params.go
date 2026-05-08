package jurikfractaladaptivezerolagvelocity

import "zpano/entities"

// JurikFractalAdaptiveZeroLagVelocityParams describes the parameters of the indicator.
type JurikFractalAdaptiveZeroLagVelocityParams struct {
	// LoDepth is the minimum depth for the velocity computation.
	LoDepth int `json:"lo_depth"`

	// HiDepth is the maximum depth for the velocity computation.
	HiDepth int `json:"hi_depth"`

	// FractalType selects the scale set (1-4).
	FractalType int `json:"fractal_type"`

	// Smooth is the smoothing window for CFB channel averages.
	Smooth int `json:"smooth"`

	// BarComponent specifies the bar component to use.
	BarComponent entities.BarComponent `json:"bar_component"`

	// QuoteComponent specifies the quote component to use.
	QuoteComponent entities.QuoteComponent `json:"quote_component"`

	// TradeComponent specifies the trade component to use.
	TradeComponent entities.TradeComponent `json:"trade_component"`
}

// DefaultParams returns default parameters for the indicator.
func DefaultParams() *JurikFractalAdaptiveZeroLagVelocityParams {
	return &JurikFractalAdaptiveZeroLagVelocityParams{
		LoDepth:     5,
		HiDepth:     30,
		FractalType: 1,
		Smooth:      10,
	}
}
