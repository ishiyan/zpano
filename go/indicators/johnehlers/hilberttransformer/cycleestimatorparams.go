package hilberttransformer

// CycleEstimatorParams describes parameters to create an instance
// of the Hilbert transformer cycle estimator.
type CycleEstimatorParams struct {
	// The smoothing length (the number of time periods) of the underlying
	// linear-Weighted Moving Average (WMA).
	//
	// The valid values are 2, 3, 4.
	// The default value is 4.
	SmoothingLength int

	// The value of α (0 < α ≤ 1) used in EMA to smooth the in-phase
	// and quadrature components.
	//
	// The default values per estimator type are:
	//
	// - homodyne discriminator: 0.2
	// - phase accumulator: 0.15
	// - dual differentiator: 0.15
	AlphaEmaQuadratureInPhase float64

	// The value of α (0 < α ≤ 1) used in EMA to smooth the instantaneous period.
	//
	// The default values per estimator type are:
	//
	// - homodyne discriminator: 0.2
	// - phase accumulator: 0.25
	// - dual differentiator: 0.15
	AlphaEmaPeriod float64

	// The number of updates before the estimator is primed.
	//
	// If less than the implementation-specific primed length, it will be overridden
	// by the implementation-specific primed length.
	WarmUpPeriod int
}
