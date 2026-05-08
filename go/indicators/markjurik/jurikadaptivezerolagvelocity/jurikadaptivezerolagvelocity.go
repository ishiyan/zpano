package jurikadaptivezerolagvelocity

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// velSmooth is the adaptive smoother (Stage 2) for JAVEL.
type velSmooth struct {
	jrc03       float64
	jrc06       int
	jrc07       int
	emaFactor   float64
	damping     float64
	eps2        float64
	bufferSize  int
	buffer      []float64
	head        int
	length      int
	barCount    int
	velocity    float64
	position    float64
	smoothedMAD float64
	initialized bool
}

func newVelSmooth(period float64) *velSmooth {
	const eps2 = 0.0001

	jrc03 := math.Min(500.0, math.Max(eps2, period))
	jrc06 := int(math.Max(31, math.Ceil(2*period)))
	jrc07 := int(math.Min(30, math.Ceil(period)))
	emaFactor := 1.0 - math.Exp(-math.Log(4.0)/(period/2.0))
	damping := 0.86 - 0.55/math.Sqrt(jrc03)

	return &velSmooth{
		jrc03:      jrc03,
		jrc06:      jrc06,
		jrc07:      jrc07,
		emaFactor:  emaFactor,
		damping:    damping,
		eps2:       eps2,
		bufferSize: 1001,
		buffer:     make([]float64, 1001),
	}
}

func (s *velSmooth) update(value float64) float64 {
	s.barCount++

	// Store in circular buffer.
	oldIndex := s.head % s.bufferSize
	s.buffer[oldIndex] = value
	s.head++

	if s.length < s.jrc06 {
		s.length++
	}

	length := s.length

	// First bar: initialize position.
	if length < 2 {
		if !s.initialized {
			s.position = value
			s.initialized = true
		}

		return s.position
	}

	if !s.initialized {
		s.position = value
		s.initialized = true
	}

	// Linear regression over buffer (forward: k=0 oldest, k=length-1 newest).
	var sumValues, sumWeighted float64

	for k := 0; k < length; k++ {
		idx := (s.head - length + k) % s.bufferSize
		if idx < 0 {
			idx += s.bufferSize
		}

		sumValues += s.buffer[idx]
		sumWeighted += s.buffer[idx] * float64(k)
	}

	midpoint := float64(length-1) / 2.0
	sumXSq := float64(length) * float64(length-1) * float64(2*length-1) / 6.0
	regressionDenom := sumXSq - float64(length)*midpoint*midpoint

	var regressionSlope float64
	if math.Abs(regressionDenom) >= s.eps2 {
		regressionSlope = (sumWeighted - midpoint*sumValues) / regressionDenom
	}

	intercept := sumValues/float64(length) - regressionSlope*midpoint

	// Compute MAD from regression residuals.
	var sumAbsDev float64

	for k := 0; k < length; k++ {
		idx := (s.head - length + k) % s.bufferSize
		if idx < 0 {
			idx += s.bufferSize
		}

		predicted := intercept + regressionSlope*float64(k)
		sumAbsDev += math.Abs(s.buffer[idx] - predicted)
	}

	rawMAD := sumAbsDev / float64(length)
	scale := 1.2 * math.Pow(float64(s.jrc06)/float64(length), 0.25)
	rawMAD *= scale

	// Smooth MAD with EMA (seed for first jrc07+1 bars).
	if s.barCount <= s.jrc07+1 {
		s.smoothedMAD = rawMAD
	} else {
		s.smoothedMAD += s.emaFactor * (rawMAD - s.smoothedMAD)
	}

	// Adaptive velocity/position dynamics.
	predictionError := value - s.position

	var responseFactor float64
	if s.smoothedMAD*s.jrc03 < s.eps2 {
		responseFactor = 1.0
	} else {
		responseFactor = 1.0 - math.Exp(-math.Abs(predictionError)/(s.smoothedMAD*s.jrc03))
	}

	s.velocity = responseFactor*predictionError + s.velocity*s.damping
	s.position += s.velocity

	return s.position
}

// JurikAdaptiveZeroLagVelocity computes the Jurik Adaptive Zero Lag Velocity (JAVEL) indicator.
type JurikAdaptiveZeroLagVelocity struct {
	mu sync.RWMutex
	core.LineIndicator
	primed      bool
	loLength    int
	hiLength    int
	sensitivity float64
	eps         float64

	prices   []float64
	value1   []float64
	barCount int
	smooth   *velSmooth
}

// NewJurikAdaptiveZeroLagVelocity returns an instance of the indicator.
func NewJurikAdaptiveZeroLagVelocity(p *JurikAdaptiveZeroLagVelocityParams) (*JurikAdaptiveZeroLagVelocity, error) {
	return newJurikAdaptiveZeroLagVelocity(p.LoLength, p.HiLength, p.Sensitivity, p.Period,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

func newJurikAdaptiveZeroLagVelocity(loLength, hiLength int, sensitivity, period float64,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*JurikAdaptiveZeroLagVelocity, error) {
	const (
		invalid = "invalid jurik adaptive zero lag velocity parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "javel(%d, %d, %v, %v%s)"
	)

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if loLength < 2 {
		return nil, fmt.Errorf(fmts, invalid, "lo_length should be at least 2")
	}

	if hiLength < loLength {
		return nil, fmt.Errorf(fmts, invalid, "hi_length should be at least lo_length")
	}

	if period <= 0 {
		return nil, fmt.Errorf(fmts, invalid, "period should be positive")
	}

	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	mnemonic := fmt.Sprintf(fmtn, loLength, hiLength, sensitivity, period,
		core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Jurik adaptive zero lag velocity " + mnemonic

	ind := &JurikAdaptiveZeroLagVelocity{
		loLength:    loLength,
		hiLength:    hiLength,
		sensitivity: sensitivity,
		eps:         0.001,
		smooth:      newVelSmooth(period),
	}

	ind.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ind.Update)

	return ind, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *JurikAdaptiveZeroLagVelocity) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *JurikAdaptiveZeroLagVelocity) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikAdaptiveZeroLagVelocity,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

func (s *JurikAdaptiveZeroLagVelocity) computeAdaptiveDepth(bar int) float64 {
	longWindow := bar
	if longWindow > 99 {
		longWindow = 99
	}

	longWindow++

	shortWindow := bar
	if shortWindow > 9 {
		shortWindow = 9
	}

	shortWindow++

	var avg1 float64
	for i := bar - longWindow + 1; i <= bar; i++ {
		avg1 += s.value1[i]
	}

	avg1 /= float64(longWindow)

	var avg2 float64
	for i := bar - shortWindow + 1; i <= bar; i++ {
		avg2 += s.value1[i]
	}

	avg2 /= float64(shortWindow)

	value2 := s.sensitivity * math.Log((s.eps+avg1)/(s.eps+avg2))
	value3 := value2 / (1.0 + math.Abs(value2))

	return float64(s.loLength) + float64(s.hiLength-s.loLength)*(1.0+value3)/2.0
}

func (s *JurikAdaptiveZeroLagVelocity) computeWLSSlope(bar, depth int) float64 {
	n := float64(depth + 1)
	s1 := n * (n + 1) / 2.0
	s2 := s1 * (2*n + 1) / 3.0
	denom := s1*s1*s1 - s2*s2

	var sumXW, sumXW2 float64

	for i := 0; i <= depth; i++ {
		w := n - float64(i)
		p := s.prices[bar-i]
		sumXW += p * w
		sumXW2 += p * w * w
	}

	return (sumXW2*s1 - sumXW*s2) / denom
}

// Update updates the indicator given the next sample.
func (s *JurikAdaptiveZeroLagVelocity) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bar := s.barCount
	s.barCount++

	s.prices = append(s.prices, sample)

	// Compute value1 (abs diff).
	if bar == 0 {
		s.value1 = append(s.value1, 0.0)
	} else {
		s.value1 = append(s.value1, math.Abs(sample-s.prices[bar-1]))
	}

	// Compute adaptive depth.
	adaptiveDepth := s.computeAdaptiveDepth(bar)
	depth := int(math.Ceil(adaptiveDepth))

	// Check if we have enough prices for WLS.
	if bar < depth {
		return math.NaN()
	}

	// Stage 1: WLS slope.
	slope := s.computeWLSSlope(bar, depth)

	// Stage 2: adaptive smoother.
	result := s.smooth.update(slope)

	if !s.primed {
		s.primed = true
	}

	return result
}
