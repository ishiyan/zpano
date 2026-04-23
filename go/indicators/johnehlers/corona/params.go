package corona

// Params configures a Corona spectral analysis engine.
//
// All fields have zero-value defaults: a zero value means "use the default".
// The defaults follow Ehlers' original TASC article (November 2008).
type Params struct {
	// HighPassFilterCutoff is the cutoff period of the detrending high-pass
	// filter, in bars. Default: 30. Zero or negative => default.
	HighPassFilterCutoff int

	// MinimalPeriod is the minimum cycle period (in bars) covered by the
	// bandpass filter bank. Must be >= 2. Default: 6. Zero => default.
	MinimalPeriod int

	// MaximalPeriod is the maximum cycle period (in bars) covered by the
	// bandpass filter bank. Must be > MinimalPeriod. Default: 30. Zero => default.
	MaximalPeriod int

	// DecibelsLowerThreshold: filter bins with smoothed dB value at or below
	// this threshold contribute to the weighted dominant-cycle estimate.
	// Default: 6. Zero => default.
	DecibelsLowerThreshold float64

	// DecibelsUpperThreshold: upper clamp on the smoothed dB value and
	// reference value for the dominant-cycle weighting (weight = upper - dB).
	// Default: 20. Zero => default.
	DecibelsUpperThreshold float64
}

// DefaultParams returns a Params value populated with Ehlers defaults.
func DefaultParams() *Params {
	return &Params{
		HighPassFilterCutoff:   30,
		MinimalPeriod:          6,
		MaximalPeriod:          30,
		DecibelsLowerThreshold: 6,
		DecibelsUpperThreshold: 20,
	}
}
