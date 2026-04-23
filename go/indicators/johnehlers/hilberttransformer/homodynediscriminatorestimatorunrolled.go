package hilberttransformer

import (
	"math"
	"sync"
)

// HomodyneDiscriminatorEstimator implements the Hilbert transformer of the
// WMA-smoothed and detrended data with the Homodyne Discriminator applied.
//
// Copied from the TA-Lib implementation with unrolled loops.
//
// John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 52-77.
type HomodyneDiscriminatorEstimatorUnrolled struct {
	mu                              sync.RWMutex
	smoothingLength                 int
	minPeriod                       int
	maxPeriod                       int
	alphaEmaQuadratureInPhase       float64
	alphaEmaPeriod                  float64
	warmUpPeriod                    int
	oneMinAlphaEmaQuadratureInPhase float64
	oneMinAlphaEmaPeriod            float64
	smoothed                        float64
	detrended                       float64
	inPhase                         float64
	quadrature                      float64
	smoothingMultiplier             float64
	adjustedPeriod                  float64
	count                           int
	index                           int
	i2Previous                      float64
	q2Previous                      float64
	re                              float64
	im                              float64
	period                          float64
	isPrimed                        bool
	isWarmedUp                      bool

	// WMA smoother private members.
	wmaSum    float64
	wmaSub    float64
	wmaInput1 float64
	wmaInput2 float64
	wmaInput3 float64
	wmaInput4 float64

	// Detrender private members.
	detrenderOdd0              float64
	detrenderOdd1              float64
	detrenderOdd2              float64
	detrenderPreviousOdd       float64
	detrenderPreviousInputOdd  float64
	detrenderEven0             float64
	detrenderEven1             float64
	detrenderEven2             float64
	detrenderPreviousEven      float64
	detrenderPreviousInputEven float64

	// Quadrature (Q1) component private members.
	q1Odd0              float64
	q1Odd1              float64
	q1Odd2              float64
	q1PreviousOdd       float64
	q1PreviousInputOdd  float64
	q1Even0             float64
	q1Even1             float64
	q1Even2             float64
	q1PreviousEven      float64
	q1PreviousInputEven float64

	// InPhase (I1) private members.
	i1Previous1Odd  float64
	i1Previous2Odd  float64
	i1Previous1Even float64
	i1Previous2Even float64

	// jI private members
	jiOdd0              float64
	jiOdd1              float64
	jiOdd2              float64
	jiPreviousOdd       float64
	jiPreviousInputOdd  float64
	jiEven0             float64
	jiEven1             float64
	jiEven2             float64
	jiPreviousEven      float64
	jiPreviousInputEven float64

	// jQ private members.
	jqOdd0              float64
	jqOdd1              float64
	jqOdd2              float64
	jqPreviousOdd       float64
	jqPreviousInputOdd  float64
	jqEven0             float64
	jqEven1             float64
	jqEven2             float64
	jqPreviousEven      float64
	jqPreviousInputEven float64
}

// NewHomodyneDiscriminatorEstimator returns an instnce of the estimator using supplied parameters.
func NewHomodyneDiscriminatorEstimatorUnrolled(
	p *CycleEstimatorParams,
) (*HomodyneDiscriminatorEstimatorUnrolled, error) {
	err := verifyParameters(p)
	if err != nil {
		return nil, err
	}

	length := p.SmoothingLength
	alphaQuad := p.AlphaEmaQuadratureInPhase
	alphaPeriod := p.AlphaEmaPeriod

	const (
		c4   = 4
		c3   = 3
		c110 = 1. / 10.
		c16  = 1. / 6.
		c13  = 1. / 3.
	)

	smoothingMultiplier := c13 // length == 2
	if length == c4 {
		smoothingMultiplier = c110
	} else if length == c3 {
		smoothingMultiplier = c16
	}

	// The TA-Lib implementation uses the following lookback value with hardcoded smoothingLength=4.
	//
	// The fixed lookback is 32 and is establish as follows:
	// 12 price bar to be compatible with the implementation of Tradestation found in John Ehlers book,
	// 6 price bars for the Detrender,
	// 6 price bars for Q1,
	// 3 price bars for jI,
	// 3 price bars for jQ,
	// 1 price bar for Re/Im,
	// 1 price bar for the Delta Phase,
	// --------
	// 32 Total.
	//
	// The first 9 bars are not used by TA-Lib, they are just skipped for the compatibility with the Tradestation.
	// We do not skip them. Thus, we use the fixed lookback value 32 - 9 = 23.
	const primedCount = 23

	warmUpPeriod := primedCount
	if p.WarmUpPeriod > primedCount {
		warmUpPeriod = p.WarmUpPeriod
	}

	return &HomodyneDiscriminatorEstimatorUnrolled{
		smoothingLength:                 length,
		minPeriod:                       defaultMinPeriod,
		maxPeriod:                       defaultMaxPeriod,
		alphaEmaQuadratureInPhase:       alphaQuad,
		alphaEmaPeriod:                  alphaPeriod,
		warmUpPeriod:                    warmUpPeriod,
		oneMinAlphaEmaQuadratureInPhase: 1 - alphaQuad,
		oneMinAlphaEmaPeriod:            1 - alphaPeriod,
		smoothingMultiplier:             smoothingMultiplier,
		period:                          defaultMinPeriod,
	}, nil
}

// SmoothingLength returns the underlying WMA smoothing length in samples.
func (s *HomodyneDiscriminatorEstimatorUnrolled) SmoothingLength() int {
	return s.smoothingLength
}

// MinPeriod returns the minimal cycle period.
func (s *HomodyneDiscriminatorEstimatorUnrolled) MinPeriod() int {
	return s.minPeriod
}

// MaxPeriod returns the maximual cycle period.
func (s *HomodyneDiscriminatorEstimatorUnrolled) MaxPeriod() int {
	return s.maxPeriod
}

// WarmUpPeriod returns the number of updates before the estimator
// is primed (MaxPeriod * 2 = 100).
func (s *HomodyneDiscriminatorEstimatorUnrolled) WarmUpPeriod() int {
	return s.warmUpPeriod
}

// AlphaEmaQuadratureInPhase returns the value of α (0 < α ≤ 1)
// used in EMA to smooth the in-phase and quadrature components.
func (s *HomodyneDiscriminatorEstimatorUnrolled) AlphaEmaQuadratureInPhase() float64 {
	return s.alphaEmaQuadratureInPhase
}

// AlphaEmaPeriod returns the value of α (0 < α ≤ 1)
// used in EMA to smooth the instantaneous period.
func (s *HomodyneDiscriminatorEstimatorUnrolled) AlphaEmaPeriod() float64 {
	return s.alphaEmaPeriod
}

// Count returns the current count value.
func (s *HomodyneDiscriminatorEstimatorUnrolled) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count
}

// Primed indicates whether an instance is primed.
func (s *HomodyneDiscriminatorEstimatorUnrolled) Primed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.isWarmedUp
}

// Period returns the current period value.
func (s *HomodyneDiscriminatorEstimatorUnrolled) Period() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.period
}

// InPhase returns the current InPhase component value.
func (s *HomodyneDiscriminatorEstimatorUnrolled) InPhase() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.inPhase
}

// Quadrature returns the current Quadrature component value.
func (s *HomodyneDiscriminatorEstimatorUnrolled) Quadrature() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.quadrature
}

// Detrended returns the current detrended value.
func (s *HomodyneDiscriminatorEstimatorUnrolled) Detrended() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.detrended
}

// Smoothed returns the current WMA-smoothed value used by underlying Hilbert transformer.
//
// The linear-Weighted Moving Average has a window size of SmoothingLength.
func (s *HomodyneDiscriminatorEstimatorUnrolled) Smoothed() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.smoothed
}

// Update updates the estimator given the next sample value.
func (s *HomodyneDiscriminatorEstimatorUnrolled) Update(sample float64) { //nolint:funlen, cyclop, gocognit, maintidx
	if math.IsNaN(sample) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	const (
		c2                      = 2
		c3                      = 3
		c4                      = 4
		primedCount             = 23
		a                       = 0.0962
		b                       = 0.5769
		minPreviousPeriodFactor = 0.67
		maxPreviousPeriodFactor = 1.5
		c0075                   = 0.075
		c054                    = 0.54
	)

	var value float64

	// WMA smoothing.
	s.count++

	// We need (smoothingLength - 1) bars to accumulate the WMA sub and sum.
	// On (smoothingLength)-th bar we calculate the first WMA smoothed value and begin with detrender.
	if s.smoothingLength >= s.count { //nolint:nestif
		if 1 == s.count {
			s.wmaSub = sample
			s.wmaInput1 = sample
			s.wmaSum = sample
		} else if c2 == s.count {
			s.wmaSub += sample
			s.wmaInput2 = sample
			s.wmaSum += sample * c2

			if c2 == s.smoothingLength {
				value = s.wmaSum * s.smoothingMultiplier

				goto detrendLabel
			}
		} else if c3 == s.count {
			s.wmaSub += sample
			s.wmaInput3 = sample
			s.wmaSum += sample * c3

			if c3 == s.smoothingLength {
				value = s.wmaSum * s.smoothingMultiplier

				goto detrendLabel
			}
		} else { // 4 == s.count
			s.wmaSub += sample
			s.wmaInput4 = sample
			s.wmaSum += sample * c4
			value = s.wmaSum * s.smoothingMultiplier

			goto detrendLabel
		}

		return
	}

	s.wmaSum -= s.wmaSub
	s.wmaSum += sample * float64(s.smoothingLength)
	value = s.wmaSum * s.smoothingMultiplier
	s.wmaSub += sample
	s.wmaSub -= s.wmaInput1
	s.wmaInput1 = s.wmaInput2

	if c4 == s.smoothingLength {
		s.wmaInput2 = s.wmaInput3
		s.wmaInput3 = s.wmaInput4
		s.wmaInput4 = sample
	} else if c3 == s.smoothingLength {
		s.wmaInput2 = s.wmaInput3
		s.wmaInput3 = sample
	} else { // 2 == ss.smoothingLength
		s.wmaInput2 = sample
	}

	// Detrender.
detrendLabel:
	s.smoothed = value

	if !s.isWarmedUp {
		s.isWarmedUp = s.count > s.warmUpPeriod
		if !s.isPrimed {
			s.isPrimed = s.count > primedCount
		}
	}

	var detrender, ji, jq float64

	temp := a * s.smoothed
	s.adjustedPeriod = c0075*s.period + c054

	// Even value count.
	if 0 == s.count%2 { //nolint:dupl, nestif
		// Explicitly expanded index.
		if 0 == s.index { //nolint:dupl
			s.index = 1
			detrender = -s.detrenderEven0
			s.detrenderEven0 = temp
			detrender += temp
			detrender -= s.detrenderPreviousEven
			s.detrenderPreviousEven = b * s.detrenderPreviousInputEven
			s.detrenderPreviousInputEven = value
			detrender += s.detrenderPreviousEven
			detrender *= s.adjustedPeriod

			// Quadrature component.
			temp = a * detrender
			s.quadrature = -s.q1Even0
			s.q1Even0 = temp
			s.quadrature += temp
			s.quadrature -= s.q1PreviousEven
			s.q1PreviousEven = b * s.q1PreviousInputEven
			s.q1PreviousInputEven = detrender
			s.quadrature += s.q1PreviousEven
			s.quadrature *= s.adjustedPeriod

			// Advance the phase of the InPhase component by 90°.
			temp = a * s.i1Previous2Even
			ji = -s.jiEven0
			s.jiEven0 = temp
			ji += temp
			ji -= s.jiPreviousEven
			s.jiPreviousEven = b * s.jiPreviousInputEven
			s.jiPreviousInputEven = s.i1Previous2Even
			ji += s.jiPreviousEven
			ji *= s.adjustedPeriod

			// Advance the phase of the Quadrature component by 90°.
			temp = a * s.quadrature
			jq = -s.jqEven0
			s.jqEven0 = temp
		} else if 1 == s.index { //nolint:dupl
			s.index = 2
			detrender = -s.detrenderEven1
			s.detrenderEven1 = temp
			detrender += temp
			detrender -= s.detrenderPreviousEven
			s.detrenderPreviousEven = b * s.detrenderPreviousInputEven
			s.detrenderPreviousInputEven = value
			detrender += s.detrenderPreviousEven
			detrender *= s.adjustedPeriod

			// Quadrature component.
			temp = a * detrender
			s.quadrature = -s.q1Even1
			s.q1Even1 = temp
			s.quadrature += temp
			s.quadrature -= s.q1PreviousEven
			s.q1PreviousEven = b * s.q1PreviousInputEven
			s.q1PreviousInputEven = detrender
			s.quadrature += s.q1PreviousEven
			s.quadrature *= s.adjustedPeriod

			// Advance the phase of the InPhase component by 90°.
			temp = a * s.i1Previous2Even
			ji = -s.jiEven1
			s.jiEven1 = temp
			ji += temp
			ji -= s.jiPreviousEven
			s.jiPreviousEven = b * s.jiPreviousInputEven
			s.jiPreviousInputEven = s.i1Previous2Even
			ji += s.jiPreviousEven
			ji *= s.adjustedPeriod

			// Advance the phase of the Quadrature component by 90°.
			temp = a * s.quadrature
			jq = -s.jqEven1
			s.jqEven1 = temp
		} else { //nolint:dupl // 2 == s.index
			s.index = 0
			detrender = -s.detrenderEven2
			s.detrenderEven2 = temp
			detrender += temp
			detrender -= s.detrenderPreviousEven
			s.detrenderPreviousEven = b * s.detrenderPreviousInputEven
			s.detrenderPreviousInputEven = value
			detrender += s.detrenderPreviousEven
			detrender *= s.adjustedPeriod

			// Quadrature component.
			temp = a * detrender
			s.quadrature = -s.q1Even2
			s.q1Even2 = temp
			s.quadrature += temp
			s.quadrature -= s.q1PreviousEven
			s.q1PreviousEven = b * s.q1PreviousInputEven
			s.q1PreviousInputEven = detrender
			s.quadrature += s.q1PreviousEven
			s.quadrature *= s.adjustedPeriod

			// Advance the phase of the InPhase component by 90°.
			temp = a * s.i1Previous2Even
			ji = -s.jiEven2
			s.jiEven2 = temp
			ji += temp
			ji -= s.jiPreviousEven
			s.jiPreviousEven = b * s.jiPreviousInputEven
			s.jiPreviousInputEven = s.i1Previous2Even
			ji += s.jiPreviousEven
			ji *= s.adjustedPeriod

			// Advance the phase of the Quadrature component by 90°.
			temp = a * s.quadrature
			jq = -s.jqEven2
			s.jqEven2 = temp
		}

		// Advance the phase of the Quadrature component by 90° (continued).
		jq += temp
		jq -= s.jqPreviousEven
		s.jqPreviousEven = b * s.jqPreviousInputEven
		s.jqPreviousInputEven = s.quadrature
		jq += s.jqPreviousEven
		jq *= s.adjustedPeriod

		// InPhase component.
		s.inPhase = s.i1Previous2Even

		// The current detrender value will be used by the "odd" logic later.
		s.i1Previous2Odd = s.i1Previous1Odd
		s.i1Previous1Odd = detrender
	} else { //nolint:dupl // Odd value count.
		if 0 == s.index { //nolint:dupl
			s.index = 1
			detrender = -s.detrenderOdd0
			s.detrenderOdd0 = temp
			detrender += temp
			detrender -= s.detrenderPreviousOdd
			s.detrenderPreviousOdd = b * s.detrenderPreviousInputOdd
			s.detrenderPreviousInputOdd = value
			detrender += s.detrenderPreviousOdd
			detrender *= s.adjustedPeriod

			// Quadrature component.
			temp = a * detrender
			s.quadrature = -s.q1Odd0
			s.q1Odd0 = temp
			s.quadrature += temp
			s.quadrature -= s.q1PreviousOdd
			s.q1PreviousOdd = b * s.q1PreviousInputOdd
			s.q1PreviousInputOdd = detrender
			s.quadrature += s.q1PreviousOdd
			s.quadrature *= s.adjustedPeriod

			// Advance the phase of the InPhase component by 90°.
			temp = a * s.i1Previous2Odd
			ji = -s.jiOdd0
			s.jiOdd0 = temp
			ji += temp
			ji -= s.jiPreviousOdd
			s.jiPreviousOdd = b * s.jiPreviousInputOdd
			s.jiPreviousInputOdd = s.i1Previous2Odd
			ji += s.jiPreviousOdd
			ji *= s.adjustedPeriod

			// Advance the phase of the Quadrature component by 90°.
			temp = a * s.quadrature
			jq = -s.jqOdd0
			s.jqOdd0 = temp
		} else if 1 == s.index { //nolint:dupl
			s.index = 2

			// Quadrature component.
			detrender = -s.detrenderOdd1
			s.detrenderOdd1 = temp
			detrender += temp
			detrender -= s.detrenderPreviousOdd
			s.detrenderPreviousOdd = b * s.detrenderPreviousInputOdd
			s.detrenderPreviousInputOdd = value
			detrender += s.detrenderPreviousOdd
			detrender *= s.adjustedPeriod
			temp = a * detrender
			s.quadrature = -s.q1Odd1
			s.q1Odd1 = temp
			s.quadrature += temp
			s.quadrature -= s.q1PreviousOdd
			s.q1PreviousOdd = b * s.q1PreviousInputOdd
			s.q1PreviousInputOdd = detrender
			s.quadrature += s.q1PreviousOdd
			s.quadrature *= s.adjustedPeriod

			// Advance the phase of the InPhase component by 90°.
			temp = a * s.i1Previous2Odd
			ji = -s.jiOdd1
			s.jiOdd1 = temp
			ji += temp
			ji -= s.jiPreviousOdd
			s.jiPreviousOdd = b * s.jiPreviousInputOdd
			s.jiPreviousInputOdd = s.i1Previous2Odd
			ji += s.jiPreviousOdd
			ji *= s.adjustedPeriod

			// Advance the phase of the Quadrature component by 90°.
			temp = a * s.quadrature
			jq = -s.jqOdd1
			s.jqOdd1 = temp
		} else { //nolint:dupl // 2 == s.index
			s.index = 0
			detrender = -s.detrenderOdd2
			s.detrenderOdd2 = temp
			detrender += temp
			detrender -= s.detrenderPreviousOdd
			s.detrenderPreviousOdd = b * s.detrenderPreviousInputOdd
			s.detrenderPreviousInputOdd = value
			detrender += s.detrenderPreviousOdd
			detrender *= s.adjustedPeriod

			// Quadrature component.
			temp = a * detrender
			s.quadrature = -s.q1Odd2
			s.q1Odd2 = temp
			s.quadrature += temp
			s.quadrature -= s.q1PreviousOdd
			s.q1PreviousOdd = b * s.q1PreviousInputOdd
			s.q1PreviousInputOdd = detrender
			s.quadrature += s.q1PreviousOdd
			s.quadrature *= s.adjustedPeriod

			// Advance the phase of the InPhase component by 90°.
			temp = a * s.i1Previous2Odd
			ji = -s.jiOdd2
			s.jiOdd2 = temp
			ji += temp
			ji -= s.jiPreviousOdd
			s.jiPreviousOdd = b * s.jiPreviousInputOdd
			s.jiPreviousInputOdd = s.i1Previous2Odd
			ji += s.jiPreviousOdd
			ji *= s.adjustedPeriod

			// Advance the phase of the Quadrature component by 90°.
			temp = a * s.quadrature
			jq = -s.jqOdd2
			s.jqOdd2 = temp
		}

		// Advance the phase of the Quadrature component by 90° (continued).
		jq += temp
		jq -= s.jqPreviousOdd
		s.jqPreviousOdd = b * s.jqPreviousInputOdd
		s.jqPreviousInputOdd = s.quadrature
		jq += s.jqPreviousOdd
		jq *= s.adjustedPeriod

		// InPhase component.
		s.inPhase = s.i1Previous2Odd

		// The current detrender value will be used by the "even" logic later.
		s.i1Previous2Even = s.i1Previous1Even
		s.i1Previous1Even = detrender
	}

	s.detrended = detrender

	// Phasor addition for 3 bar averaging.
	i2 := s.inPhase - jq
	q2 := s.quadrature + ji

	// Smooth the InPhase and the Quadrature components before applying the discriminator.
	i2 = s.alphaEmaQuadratureInPhase*i2 + s.oneMinAlphaEmaQuadratureInPhase*s.i2Previous
	q2 = s.alphaEmaQuadratureInPhase*q2 + s.oneMinAlphaEmaQuadratureInPhase*s.q2Previous

	// Homodyne discriminator.
	// Homodyne means we are multiplying the signal by itself.
	// We multiply the signal of the current bar with the complex conjugate of the signal 1 bar ago.
	s.re = s.alphaEmaQuadratureInPhase*(i2*s.i2Previous+q2*s.q2Previous) + s.oneMinAlphaEmaQuadratureInPhase*s.re
	s.im = s.alphaEmaQuadratureInPhase*(i2*s.q2Previous-q2*s.i2Previous) + s.oneMinAlphaEmaQuadratureInPhase*s.im
	s.q2Previous = q2
	s.i2Previous = i2
	temp = s.period

	periodNew := c2 * math.Pi / math.Atan2(s.im, s.re)
	if !math.IsNaN(periodNew) && !math.IsInf(periodNew, 0) {
		s.period = periodNew
	}

	value = maxPreviousPeriodFactor * temp
	if s.period > value {
		s.period = value
	} else {
		value = minPreviousPeriodFactor * temp
		if s.period < value {
			s.period = value
		}
	}

	if s.period < defaultMinPeriod {
		s.period = defaultMinPeriod
	} else if s.period > defaultMaxPeriod {
		s.period = defaultMaxPeriod
	}

	s.period = s.alphaEmaPeriod*s.period + s.oneMinAlphaEmaPeriod*temp
}
