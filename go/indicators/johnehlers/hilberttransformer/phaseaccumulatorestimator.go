package hilberttransformer

import (
	"math"
	"sync"
)

const accumulationLength = 40

// PhaseAccumulatorEstimatorEstimator implements the Hilbert transformer of the
// WMA-smoothed and detrended data followed by the Phase Accumulation to determine
// the instant period.
//
// John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 63-66.
type PhaseAccumulatorEstimator struct {
	mu                               sync.RWMutex
	smoothingLength                  int
	minPeriod                        int
	maxPeriod                        int
	alphaEmaQuadratureInPhase        float64
	alphaEmaPeriod                   float64
	warmUpPeriod                     int
	smoothingLengthPlusHtLengthMin1  int
	smoothingLengthPlus2HtLengthMin2 int
	smoothingLengthPlus2HtLengthMin1 int
	smoothingLengthPlus2HtLength     int
	oneMinAlphaEmaQuadratureInPhase  float64
	oneMinAlphaEmaPeriod             float64
	rawValues                        []float64
	wmaFactors                       []float64
	wmaSmoothed                      []float64
	detrended                        []float64
	deltaPhase                       []float64
	inPhase                          float64
	quadrature                       float64
	count                            int
	smoothedInPhasePrevious          float64
	smoothedQuadraturePrevious       float64
	phasePrevious                    float64
	period                           float64
	isPrimed                         bool
	isWarmedUp                       bool
}

// NewhaseAccumulatorEstimator returns an instnce of the estimator using supplied parameters.
func NewPhaseAccumulatorEstimator(p *CycleEstimatorParams) (*PhaseAccumulatorEstimator, error) {
	err := verifyParameters(p)
	if err != nil {
		return nil, err
	}

	length := p.SmoothingLength
	alphaQuad := p.AlphaEmaQuadratureInPhase
	alphaPeriod := p.AlphaEmaPeriod

	smoothingLengthPlusHtLengthMin1 := length + htLength - 1
	smoothingLengthPlus2HtLengthMin2 := smoothingLengthPlusHtLengthMin1 + htLength - 1
	smoothingLengthPlus2HtLengthMin1 := smoothingLengthPlus2HtLengthMin2 + 1
	smoothingLengthPlus2HtLength := smoothingLengthPlus2HtLengthMin1 + 1

	// These slices will be automatically filled with zeroes.
	wmaSmoothed := make([]float64, htLength)
	detrended := make([]float64, htLength)
	deltaPhase := make([]float64, accumulationLength)
	rawValues := make([]float64, length)
	wmaFactors := make([]float64, length)

	fillWmaFactors(length, wmaFactors)

	return &PhaseAccumulatorEstimator{
		smoothingLength:                  length,
		minPeriod:                        defaultMinPeriod,
		maxPeriod:                        defaultMaxPeriod,
		alphaEmaQuadratureInPhase:        alphaQuad,
		alphaEmaPeriod:                   alphaPeriod,
		warmUpPeriod:                     max(p.WarmUpPeriod, smoothingLengthPlus2HtLength),
		smoothingLengthPlusHtLengthMin1:  smoothingLengthPlusHtLengthMin1,
		smoothingLengthPlus2HtLengthMin2: smoothingLengthPlus2HtLengthMin2,
		smoothingLengthPlus2HtLengthMin1: smoothingLengthPlus2HtLengthMin1,
		smoothingLengthPlus2HtLength:     smoothingLengthPlus2HtLength,
		oneMinAlphaEmaQuadratureInPhase:  1 - alphaQuad,
		oneMinAlphaEmaPeriod:             1 - alphaPeriod,
		rawValues:                        rawValues,
		wmaFactors:                       wmaFactors,
		wmaSmoothed:                      wmaSmoothed,
		detrended:                        detrended,
		deltaPhase:                       deltaPhase,
		period:                           defaultMinPeriod,
	}, nil
}

// SmoothingLength returns the underlying WMA smoothing length in samples.
func (s *PhaseAccumulatorEstimator) SmoothingLength() int {
	return s.smoothingLength
}

// MinPeriod returns the minimal cycle period.
func (s *PhaseAccumulatorEstimator) MinPeriod() int {
	return s.minPeriod
}

// MaxPeriod returns the maximual cycle period.
func (s *PhaseAccumulatorEstimator) MaxPeriod() int {
	return s.maxPeriod
}

// WarmUpPeriod returns the number of updates before the estimator
// is primed (MaxPeriod * 2 = 100).
func (s *PhaseAccumulatorEstimator) WarmUpPeriod() int {
	return s.warmUpPeriod
}

// AlphaEmaQuadratureInPhase returns the value of α (0 < α ≤ 1)
// used in EMA to smooth the in-phase and quadrature components.
func (s *PhaseAccumulatorEstimator) AlphaEmaQuadratureInPhase() float64 {
	return s.alphaEmaQuadratureInPhase
}

// AlphaEmaPeriod returns the value of α (0 < α ≤ 1)
// used in EMA to smooth the instantaneous period.
func (s *PhaseAccumulatorEstimator) AlphaEmaPeriod() float64 {
	return s.alphaEmaPeriod
}

// Count returns the current count value.
func (s *PhaseAccumulatorEstimator) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count
}

// Primed indicates whether an instance is primed.
func (s *PhaseAccumulatorEstimator) Primed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.isWarmedUp
}

// Period returns the current period value.
func (s *PhaseAccumulatorEstimator) Period() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.period
}

// InPhase returns the current InPhase component value.
func (s *PhaseAccumulatorEstimator) InPhase() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.inPhase
}

// Quadrature returns the current Quadrature component value.
func (s *PhaseAccumulatorEstimator) Quadrature() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.quadrature
}

// Detrended returns the current detrended value.
func (s *PhaseAccumulatorEstimator) Detrended() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.detrended[0]
}

// Smoothed returns the current WMA-smoothed value used by underlying Hilbert transformer.
//
// The linear-Weighted Moving Average has a window size of SmoothingLength.
func (s *PhaseAccumulatorEstimator) Smoothed() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.wmaSmoothed[0]
}

// Update updates the estimator given the next sample value.
func (s *PhaseAccumulatorEstimator) Update(sample float64) { //nolint:funlen
	if math.IsNaN(sample) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	push(s.rawValues, sample)

	if s.isPrimed { //nolint:nestif
		if !s.isWarmedUp {
			s.count++
			if s.warmUpPeriod < s.count {
				s.isWarmedUp = true
			}
		}

		// The WMA is used to remove some high-frequency components before detrending the signal.
		push(s.wmaSmoothed, s.wma(s.rawValues))

		amplitudeCorrectionFactor := correctAmplitude(s.period)

		// Since we have an amplitude-corrected Hilbert transformer, and since we want to detrend
		// over its length, we simply use the Hilbert transformer itself as the detrender.
		push(s.detrended, ht(s.wmaSmoothed)*amplitudeCorrectionFactor)

		// Compute both the in-phase and quadrature components of the detrended signal.
		s.quadrature = ht(s.detrended) * amplitudeCorrectionFactor
		s.inPhase = s.detrended[quadratureIndex]

		// Exponential moving average smoothing.
		smoothedInPhase := s.emaQuadratureInPhase(s.inPhase, s.smoothedInPhasePrevious)
		smoothedQuadrature := s.emaQuadratureInPhase(s.quadrature, s.smoothedQuadraturePrevious)
		s.smoothedInPhasePrevious = smoothedInPhase
		s.smoothedQuadraturePrevious = smoothedQuadrature

		// Compute an instantaneous phase.
		phase := instantaneousPhase(smoothedInPhase, smoothedQuadrature, s.phasePrevious)

		// Compute a differential phase.
		push(s.deltaPhase, calculateDifferentialPhase(phase, s.phasePrevious))
		s.phasePrevious = phase

		// Compute an instantaneous period.
		periodPrevious := s.period
		s.period = instantaneousPeriod(s.deltaPhase, periodPrevious)

		// Exponential moving average smoothing of the period.
		s.period = s.emaPeriod(s.period, periodPrevious)
	} else {
		// On (smoothingLength)-th sample we calculate the first
		// WMA smoothed value and begin with the detrender.
		s.count++
		if s.smoothingLength > s.count { // count < 4
			return
		}

		push(s.wmaSmoothed, s.wma(s.rawValues)) // count >= 4

		if s.smoothingLengthPlusHtLengthMin1 > s.count { // count < 10
			return
		}

		amplitudeCorrectionFactor := correctAmplitude(s.period) // count >= 10
		push(s.detrended, ht(s.wmaSmoothed)*amplitudeCorrectionFactor)

		if s.smoothingLengthPlus2HtLengthMin2 > s.count { // count < 16
			return
		}

		s.quadrature = ht(s.detrended) * amplitudeCorrectionFactor // count >= 16
		s.inPhase = s.detrended[quadratureIndex]

		if s.smoothingLengthPlus2HtLengthMin2 == s.count { // count == 16
			s.smoothedInPhasePrevious = s.inPhase
			s.smoothedQuadraturePrevious = s.quadrature

			return
		}

		smoothedInPhase := s.emaQuadratureInPhase(s.inPhase, s.smoothedInPhasePrevious) // count >= 17
		smoothedQuadrature := s.emaQuadratureInPhase(s.quadrature, s.smoothedQuadraturePrevious)
		s.smoothedInPhasePrevious = smoothedInPhase
		s.smoothedQuadraturePrevious = smoothedQuadrature

		phase := instantaneousPhase(smoothedInPhase, smoothedQuadrature, s.phasePrevious)
		push(s.deltaPhase, calculateDifferentialPhase(phase, s.phasePrevious))
		s.phasePrevious = phase

		periodPrevious := s.period
		s.period = instantaneousPeriod(s.deltaPhase, periodPrevious)

		if s.smoothingLengthPlus2HtLengthMin1 < s.count { // count >= 18
			s.period = s.emaPeriod(s.period, periodPrevious)
			s.isPrimed = true
		}
	}
}

func (s *PhaseAccumulatorEstimator) wma(array []float64) float64 {
	value := 0.
	for i := range s.smoothingLength {
		value += s.wmaFactors[i] * array[i]
	}

	return value
}

func (s *PhaseAccumulatorEstimator) emaQuadratureInPhase(value, valuePrevious float64) float64 {
	return s.alphaEmaQuadratureInPhase*value + s.oneMinAlphaEmaQuadratureInPhase*valuePrevious
}

func (s *PhaseAccumulatorEstimator) emaPeriod(value, valuePrevious float64) float64 {
	return s.alphaEmaPeriod*value + s.oneMinAlphaEmaPeriod*valuePrevious
}

func calculateDifferentialPhase(phase, phasePrevious float64) float64 {
	const (
		twoPi         = 2 * math.Pi
		piOver2       = math.Pi / 2
		threePiOver4  = 3 * math.Pi / 4
		minDeltaPhase = twoPi / defaultMaxPeriod
		maxDeltaPhase = twoPi / defaultMinPeriod
	)

	// Compute a differential phase.
	deltaPhase := phasePrevious - phase

	// Resolve phase wraparound from 1st quadrant to 4th quadrant.
	if phasePrevious < piOver2 && phase > threePiOver4 {
		deltaPhase += twoPi
	}

	/*for deltaPhase < 0 {
		deltaPhase += twoPi
	}*/

	// Limit deltaPhase to be within [minDeltaPhase, maxDeltaPhase],
	// i.e. within the bounds of [minPeriod, maxPeriod] sample cycles.
	if deltaPhase < minDeltaPhase {
		deltaPhase = minDeltaPhase
	} else if deltaPhase > maxDeltaPhase {
		deltaPhase = maxDeltaPhase
	}

	return deltaPhase
}

func instantaneousPhase(smoothedInPhase, smoothedQuadrature, phasePrevious float64) float64 {
	// Use arctangent to compute the instantaneous phase in radians.
	phase := math.Atan(math.Abs(smoothedQuadrature / smoothedInPhase))
	if math.IsNaN(phase) || math.IsInf(phase, 0) {
		return phasePrevious
	}

	// Resolve the ambiguity for quadrants 2, 3, and 4.
	if smoothedInPhase < 0 {
		const pi = math.Pi

		if smoothedQuadrature > 0 {
			phase = pi - phase // 2nd quadrant.
		} else if smoothedQuadrature < 0 {
			phase = pi + phase // 3rd quadrant.
		}
	} else if smoothedInPhase > 0 && smoothedQuadrature < 0 {
		const twoPi = 2 * math.Pi

		phase = twoPi - phase // 4th quadrant.
	}

	return phase
}

func instantaneousPeriod(deltaPhase []float64, periodPrevious float64) float64 {
	const twoPi = 2 * math.Pi

	sumPhase := 0.
	period := 0

	for i := range accumulationLength {
		sumPhase += deltaPhase[i]
		if sumPhase >= twoPi {
			period = i + 1

			break
		}
	}

	// Resolve instantaneous period errors.
	if period == 0 {
		return periodPrevious
	}

	return float64(period)
}
