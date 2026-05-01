package jurikcompositefractalbehaviorindex

import "zpano/entities" //nolint:depguard

// JurikCompositeFractalBehaviorIndexParams describes parameters to create an instance of the indicator.
type JurikCompositeFractalBehaviorIndexParams struct {
	// FractalType controls the maximum fractal depth. Valid values are 1–4:
	//   1 = JCFB24 (8 depths: 2,3,4,6,8,12,16,24)
	//   2 = JCFB48 (10 depths: +32,48)
	//   3 = JCFB96 (12 depths: +64,96)
	//   4 = JCFB192 (14 depths: +128,192)
	FractalType int

	// Smooth is the smoothing window for the running averages.
	// The value should be >= 1.
	Smooth int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns a [JurikCompositeFractalBehaviorIndexParams] value populated with conventional defaults.
func DefaultParams() *JurikCompositeFractalBehaviorIndexParams {
	return &JurikCompositeFractalBehaviorIndexParams{
		FractalType: 1,
		Smooth:      10,
	}
}
