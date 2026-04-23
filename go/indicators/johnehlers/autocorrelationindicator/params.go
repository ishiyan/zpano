package autocorrelationindicator

import "zpano/entities"

// Params describes parameters to create an instance of the AutoCorrelationIndicator.
//
// All boolean fields are named so the Go zero value (false) corresponds to the
// Ehlers reference default behavior. This lets a zero-valued Params{} produce
// the default indicator.
type Params struct {
	// MinLag is the minimum (shortest) correlation lag shown on the heatmap axis.
	// Must be >= 1. The default value is 3 (matching Ehlers' EasyLanguage listing 8-2,
	// which plots lags 3..48).
	MinLag int

	// MaxLag is the maximum (longest) correlation lag shown on the heatmap axis.
	// Must be > MinLag. Also drives the cutoff of the 2-pole Butterworth highpass
	// pre-filter. The default value is 48.
	MaxLag int

	// SmoothingPeriod is the cutoff period of the 2-pole Super Smoother pre-filter
	// applied after the highpass. Must be >= 2. The default value is 10 (matching
	// Ehlers' EasyLanguage listing 8-2, which hardcodes 10).
	SmoothingPeriod int

	// AveragingLength is the number of samples (M) used in each Pearson correlation
	// accumulation. When zero (the Ehlers default), M equals the current lag, making
	// each correlation use the same number of samples as its lag distance. When
	// positive, the same M is used for all lags. Must be >= 0.
	AveragingLength int

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarMedianPrice) is used, matching the Ehlers reference which operates on
	// (High+Low)/2. Since this differs from the framework-wide default, it is always shown in the
	// indicator mnemonic.
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
