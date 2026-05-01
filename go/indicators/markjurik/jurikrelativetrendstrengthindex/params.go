package jurikrelativetrendstrengthindex

import "zpano/entities" //nolint:depguard

// JurikRelativeTrendStrengthIndexParams describes parameters to create an instance of the indicator.
type JurikRelativeTrendStrengthIndexParams struct {
	// Length is the smoothing period. Values below 2 are clamped to produce
	// a minimum internal smoothing window of 5 bars.
	//
	// The value should be >= 2. Typical values range from 5 to 20.
	Length int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	BarComponent entities.BarComponent

	// QuoteComponent indicates the component of a quote to use when updating the indicator with a quote sample.
	QuoteComponent entities.QuoteComponent

	// TradeComponent indicates the component of a trade to use when updating the indicator with a trade sample.
	TradeComponent entities.TradeComponent
}

// DefaultParams returns a [JurikRelativeTrendStrengthIndexParams] value populated with conventional defaults.
func DefaultParams() *JurikRelativeTrendStrengthIndexParams {
	return &JurikRelativeTrendStrengthIndexParams{
		Length: 14,
	}
}
