package hilberttransformer

// CycleEstimator describes a common Hilbert transformer cycle estimator functionality.
type CycleEstimator interface { //nolint:interfacebloat
	// SmoothingLength returns the underlying WMA smoothing length in samples.
	SmoothingLength() int

	// Smoothed returns the current WMA-smoothed value used by underlying Hilbert transformer.
	//
	// The linear-Weighted Moving Average has a window size of SmoothingLength.
	Smoothed() float64

	// Detrended returns the current detrended value.
	Detrended() float64

	// Quadrature returns the current Quadrature component value.
	Quadrature() float64

	// InPhase returns the current InPhase component value.
	InPhase() float64

	// PeriodValue returns the current period value.
	Period() float64

	// Count returns the current count value.
	Count() int

	// Primed indicates whether an instance is primed.
	Primed() bool

	// MinPeriod returns the minimal cycle period.
	MinPeriod() int

	// MaxPeriod returns the maximual cycle period.
	MaxPeriod() int

	// AlphaEmaQuadratureInPhase returns the value of α (0 < α ≤ 1)
	// used in EMA to smooth the in-phase and quadrature components.
	AlphaEmaQuadratureInPhase() float64

	// AlphaEmaPeriod returns the value of α (0 < α ≤ 1)
	// used in EMA to smooth the instantaneous period.
	AlphaEmaPeriod() float64

	// WarmUpPeriod returns the number of updates before the estimator
	// is primed (MaxPeriod * 2 = 100).
	WarmUpPeriod() int

	// Update updates the estimator given the next sample value.
	Update(sample float64)
}
