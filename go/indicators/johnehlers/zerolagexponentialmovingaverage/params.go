package zerolagexponentialmovingaverage

import "zpano/entities"

// ZeroLagExponentialMovingAverageParams describes parameters to create an instance of the indicator.
type ZeroLagExponentialMovingAverageParams struct {
	// SmoothingFactor is the smoothing factor (alpha) of the EMA.
	//
	// alpha = 2/(length + 1), 0 < alpha <= 1, 1 <= length.
	// The default value is 0.25.
	SmoothingFactor float64

	// VelocityGainFactor is the gain factor used to estimate the velocity.
	//
	// The default value is 0.5.
	VelocityGainFactor float64

	// VelocityMomentumLength is the length of the momentum used to estimate the velocity.
	//
	// The value should be positive. The default value is 3.
	VelocityMomentumLength int

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

// DefaultParams returns a [ZeroLagExponentialMovingAverageParams] value populated with Ehlers defaults.
func DefaultParams() *ZeroLagExponentialMovingAverageParams {
	return &ZeroLagExponentialMovingAverageParams{
		SmoothingFactor:        0.25,
		VelocityGainFactor:     0.5,
		VelocityMomentumLength: 3,
	}
}
