package combbandpassspectrum

import "zpano/entities"

// Params describes parameters to create an instance of the CombBandPassSpectrum indicator.
//
// All boolean fields are named so the Go zero value (false) corresponds to the
// Ehlers reference default behavior. This lets a zero-valued Params{} produce
// the default indicator.
type Params struct {
	// MinPeriod is the minimum (shortest) cycle period covered by the spectrum, must be >= 2
	// (2 corresponds to the Nyquist frequency). Also drives the cutoff of the Super Smoother
	// pre-filter. The default value is 10.
	MinPeriod int

	// MaxPeriod is the maximum (longest) cycle period covered by the spectrum, must be > MinPeriod.
	// Also drives the cutoff of the 2-pole Butterworth highpass pre-filter and the length of
	// the band-pass output history kept per filter. The default value is 48.
	MaxPeriod int

	// Bandwidth is the fractional bandwidth of each band-pass filter in the comb. Must be in
	// (0, 1). Typical Ehlers values are 0.3 (default) for medium selectivity.
	Bandwidth float64

	// DisableSpectralDilationCompensation disables the spectral dilation compensation
	// (division of each band-pass output by its evaluated period before squaring)
	// when true. Ehlers' default behavior is enabled, so the default value is false (SDC on).
	DisableSpectralDilationCompensation bool

	// DisableAutomaticGainControl disables the fast-attack slow-decay automatic gain control
	// when true. Ehlers' default behavior is enabled, so the default value is false (AGC on).
	DisableAutomaticGainControl bool

	// AutomaticGainControlDecayFactor is the decay factor used by the fast-attack slow-decay
	// automatic gain control. Must be in the open interval (0, 1) when AGC is enabled. If zero,
	// the default value 0.995 is used (the value in Ehlers' EasyLanguage listing 10-1).
	AutomaticGainControlDecayFactor float64

	// FixedNormalization selects fixed (min clamped to 0) normalization when true. The default
	// is floating normalization (consistent with the other zpano spectrum heatmaps). Note that
	// Ehlers' listing 10-1 uses fixed normalization (MaxPwr only); set this to true for exact
	// EL-faithful behavior.
	FixedNormalization bool

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
