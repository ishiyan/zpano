package autocorrelationperiodogram

import "zpano/entities"

// Params describes parameters to create an instance of the AutoCorrelationPeriodogram.
//
// All boolean fields are named so the Go zero value (false) corresponds to the
// Ehlers reference default behavior. This lets a zero-valued Params{} produce
// the default indicator.
type Params struct {
	// MinPeriod is the minimum (shortest) cycle period shown on the heatmap axis.
	// Must be >= 2. Also drives the cutoff of the 2-pole Super Smoother pre-filter.
	// The default value is 10 (matching Ehlers' EasyLanguage listing 8-3, which
	// hardcodes the Super Smoother period at 10).
	MinPeriod int

	// MaxPeriod is the maximum (longest) cycle period shown on the heatmap axis.
	// Must be > MinPeriod. Also drives the cutoff of the 2-pole Butterworth highpass
	// pre-filter, the upper bound of the Pearson correlation lag range, and the upper
	// bound of the DFT inner sum. The default value is 48.
	MaxPeriod int

	// AveragingLength is the fixed number of samples (M) used in each Pearson
	// correlation accumulation across lags 0..MaxPeriod. Must be >= 1. The default
	// value is 3 (matching Ehlers' EasyLanguage listing 8-3, which initializes
	// AverageLength = 3).
	AveragingLength int

	// DisableSpectralSquaring disables the EL spectral squaring step. By default
	// (false), the smoothing recursion is R[P] = 0.2*SqSum[P]^2 + 0.8*R_previous[P],
	// matching Ehlers' listing 8-3. When true, R[P] = 0.2*SqSum[P] + 0.8*R_previous[P],
	// matching MBST's AutoCorrelationSpectrumEstimator. Exposed for investigation;
	// default keeps the EL behavior.
	DisableSpectralSquaring bool

	// DisableSmoothing disables the per-bin exponential smoothing. By default
	// (false), SqSum is smoothed via R[P] = 0.2*SqSum[P]^(1 or 2) + 0.8*R_previous[P].
	// When true, R[P] = SqSum[P]^(1 or 2) directly, with no memory of previous bars.
	DisableSmoothing bool

	// DisableAutomaticGainControl disables the fast-attack / slow-decay AGC used
	// to normalize R[P] for display. By default (false), MaxPwr is fed back and
	// decayed by AutomaticGainControlDecayFactor each bar.
	DisableAutomaticGainControl bool

	// AutomaticGainControlDecayFactor is the per-bar decay factor applied to the
	// running MaxPwr when AGC is enabled. Must be in (0, 1). The default value
	// is 0.995 (matching Ehlers' EasyLanguage listing 8-3).
	AutomaticGainControlDecayFactor float64

	// FixedNormalization disables floating minimum subtraction during heatmap
	// normalization. By default (false), each bin is rescaled by subtracting the
	// current minimum power across the spectrum before dividing by the (AGC-adjusted)
	// peak-to-peak range. When true, the minimum reference is locked at 0 and only
	// the peak power is used as the normalizing reference.
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
