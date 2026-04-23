package maximumentropyspectrum

import "zpano/entities"

// Params describes parameters to create an instance of the MaximumEntropySpectrum indicator.
//
// All boolean fields are named so the Go zero value (false) corresponds to the MBST
// default behavior. This lets a zero-valued Params{} produce the default indicator.
type Params struct {
	// Length is the number of time periods in the spectrum window. It determines the minimum
	// and maximum spectrum periods. The default value is 60.
	Length int

	// Degree is the order of the auto-regression model (Burg maximum-entropy estimation).
	// Must be positive and less than Length. The default value is 30.
	Degree int

	// MinPeriod is the minimum cycle period covered by the spectrum, must be >= 2
	// (2 corresponds to the Nyquist frequency). The default value is 2.
	MinPeriod float64

	// MaxPeriod is the maximum cycle period covered by the spectrum, must be > MinPeriod and
	// <= 2 * Length. The default value is 59.
	MaxPeriod float64

	// SpectrumResolution is the spectrum resolution (positive integer). A value of 10 means that
	// the spectrum is evaluated at every 0.1 of period amplitude. The default value is 1.
	SpectrumResolution int

	// DisableAutomaticGainControl disables the fast-attack slow-decay automatic gain control
	// when true. MBST default behavior is enabled, so the default value is false (AGC on).
	DisableAutomaticGainControl bool

	// AutomaticGainControlDecayFactor is the decay factor used by the fast-attack slow-decay
	// automatic gain control. Must be in the open interval (0, 1) when AGC is enabled. If zero,
	// the default value 0.995 is used.
	AutomaticGainControlDecayFactor float64

	// FixedNormalization selects fixed (min clamped to 0) normalization when true. MBST default
	// is floating normalization, so the default value is false (floating normalization).
	FixedNormalization bool

	// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
	//
	// If zero, the default (BarMedianPrice) is used, matching the MBST reference which operates on
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
