package hilberttransformer

import "fmt"

const (
	defaultMinPeriod = 6
	defaultMaxPeriod = 50
	htLength         = 7
	quadratureIndex  = htLength / 2
)

// NewCycleEstimator creates a new cycle estimator based on the specified type and parameters.
func NewCycleEstimator(typ CycleEstimatorType, params *CycleEstimatorParams) (CycleEstimator, error) {
	switch typ {
	case HomodyneDiscriminator:
		return NewHomodyneDiscriminatorEstimator(params)
	case HomodyneDiscriminatorUnrolled:
		return NewHomodyneDiscriminatorEstimatorUnrolled(params)
	case PhaseAccumulator:
		return NewPhaseAccumulatorEstimator(params)
	case DualDifferentiator:
		return NewDualDifferentiatorEstimator(params)
	}

	return nil, fmt.Errorf("invalid cycle estimator type: %s", typ)
}

// EstimatorMoniker returns the moniker of the cycle estimator.
func EstimatorMoniker(typ CycleEstimatorType, estimator CycleEstimator) string {
	namer := func(s string, e CycleEstimator) string {
		const f = "%s(%d, %.3f, %.3f)"

		return fmt.Sprintf(f, s, e.SmoothingLength(), e.AlphaEmaQuadratureInPhase(), e.AlphaEmaPeriod())
	}

	switch typ {
	case HomodyneDiscriminator:
		return namer("hd", estimator)
	case HomodyneDiscriminatorUnrolled:
		return namer("hdu", estimator)
	case PhaseAccumulator:
		return namer("pa", estimator)
	case DualDifferentiator:
		return namer("dd", estimator)
	}

	return ""
}

// Push shifts all elements to the right and place the new value at index zero.
func push(array []float64, value float64) {
	for i := len(array) - 1; i > 0; i-- {
		array[i] = array[i-1]
	}

	array[0] = value
}

func correctAmplitude(previousPeriod float64) float64 {
	const (
		a = 0.54
		b = 0.075
	)

	return a + b*previousPeriod
}

func ht(array []float64) float64 {
	const (
		a = 0.0962
		b = 0.5769
	)

	value := 0.
	value += a * array[0]
	value += b * array[2]
	value -= b * array[4]
	value -= a * array[6]

	return value
}

func adjustPeriod(period, periodPrevious float64) float64 {
	const (
		minPreviousPeriodFactor = 0.67
		maxPreviousPeriodFactor = 1.5
	)

	temp := maxPreviousPeriodFactor * periodPrevious
	if period > temp {
		period = temp
	} else {
		temp = minPreviousPeriodFactor * periodPrevious
		if period < temp {
			period = temp
		}
	}

	if period < defaultMinPeriod {
		period = defaultMinPeriod
	} else if period > defaultMaxPeriod {
		period = defaultMaxPeriod
	}

	return period
}

func fillWmaFactors(length int, factors []float64) {
	const (
		c4   = 4
		c3   = 3
		c410 = 4. / 10.
		c310 = 3. / 10.
		c210 = 2. / 10.
		c110 = 1. / 10.
		c36  = 3. / 6.
		c26  = 2. / 6.
		c16  = 1. / 6.
		c23  = 2. / 3.
		c13  = 1. / 3.
	)

	if length == c4 {
		factors[0] = c410
		factors[1] = c310
		factors[2] = c210
		factors[3] = c110
	} else if length == c3 {
		factors[0] = c36
		factors[1] = c26
		factors[2] = c16
	} else { // if length == 2
		factors[0] = c23
		factors[1] = c13
	}
}

func verifyParameters(p *CycleEstimatorParams) error {
	const (
		invalid = "invalid cycle estimator parameters"
		fmts    = "%s: %s"
		minLen  = 2
		maxLen  = 4
	)

	length := p.SmoothingLength
	if length < minLen || length > maxLen {
		return fmt.Errorf(fmts, invalid, "SmoothingLength should be in range [2, 4]")
	}

	alphaQuad := p.AlphaEmaQuadratureInPhase
	if alphaQuad <= 0 || alphaQuad >= 1 {
		return fmt.Errorf(fmts, invalid, "AlphaEmaQuadratureInPhase should be in range (0, 1)")
	}

	alphaPeriod := p.AlphaEmaPeriod
	if alphaPeriod <= 0 || alphaPeriod >= 1 {
		return fmt.Errorf(fmts, invalid, "AlphaEmaPeriod should be in range (0, 1)")
	}

	return nil
}
