package corona

// Params configures a Corona spectral analysis engine.
//
// Explicit zero values are NOT substituted with defaults; use [DefaultParams]
// to obtain Ehlers' November 2008 TASC article settings.
type Params struct {
	// HighPassFilterCutoff is the cutoff period of the detrending high-pass
	// filter, in bars. Must be >= 5. Default: 30.
	HighPassFilterCutoff int

	// MinimalPeriod is the minimum cycle period (in bars) covered by the
	// bandpass filter bank. Must be >= 6. Default: 6.
	MinimalPeriod int

	// MaximalPeriod is the maximum cycle period (in bars) covered by the
	// bandpass filter bank. Must be > MinimalPeriod. Default: 30.
	MaximalPeriod int

	// DecibelsLowerThreshold: filter bins with smoothed dB value at or below
	// this threshold contribute to the weighted dominant-cycle estimate.
	// Must satisfy 0 <= DecibelsLowerThreshold < DecibelsUpperThreshold. Default: 6.
	DecibelsLowerThreshold float64

	// DecibelsUpperThreshold: upper clamp on the smoothed dB value and
	// reference value for the dominant-cycle weighting (weight = upper - dB).
	// Must be > 0 and > DecibelsLowerThreshold. Default: 20.
	DecibelsUpperThreshold float64
}

// DefaultParams returns a [Params] value populated with Ehlers defaults.
func DefaultParams() *Params {
	return &Params{
		HighPassFilterCutoff:   30,
		MinimalPeriod:          6,
		MaximalPeriod:          30,
		DecibelsLowerThreshold: 6,
		DecibelsUpperThreshold: 20,
	}
}
