package modifiedexponentialmovingaverage

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// computeVelocityCoefficients computes FIR coefficients for the first derivative
// of a degree-degree polynomial fit evaluated at the most recent point (Lagrange
// basis, order=1).
func computeVelocityCoefficients(degree int) []float64 {
	nPoints := degree + 1
	coefficients := make([]float64, 0, nPoints)

	for i := 0; i < nPoints; i++ {
		denom := 1.0

		for j := 0; j < nPoints; j++ {
			if j != i {
				denom *= float64(j - i)
			}
		}

		others := make([]int, 0, degree)

		for j := 0; j < nPoints; j++ {
			if j != i {
				others = append(others, j)
			}
		}

		numerator := 0.0

		for ell := 0; ell < len(others); ell++ {
			term := 1.0

			for m := 0; m < len(others); m++ {
				if m != ell {
					term *= float64(others[m])
				}
			}

			numerator += term
		}

		coefficients = append(coefficients, numerator/denom)
	}

	return coefficients
}

// ModifiedExponentialMovingAverage is Don Mak's Modified Exponential Moving Average (MEMA / MEMA-D).
//
// It is a reduced-lag EMA that adds the EMA's own polynomial velocity back to its
// output, compensating for smoothing delay:
//
//	MEMA(n) = EMA(n) + PFD(EMA, degree, order=1, stride=skip)
//
// Reference:
//
// Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading. Ch 4.2.
type ModifiedExponentialMovingAverage struct {
	mu sync.RWMutex

	core.LineIndicator

	degree  int
	skip    int
	nPoints int

	coefficients []float64

	emaAlpha       float64
	emaValue       float64
	emaInitialized bool

	buf      []float64
	bufSize  int
	bufPos   int
	bufCount int

	primed bool
}

// NewModifiedExponentialMovingAverage returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen
func NewModifiedExponentialMovingAverage(p *Params) (*ModifiedExponentialMovingAverage, error) {
	const (
		invalid        = "invalid modified exponential moving average parameters"
		fmts           = "%s: %s"
		fmtw           = "%s: %w"
		defaultPeriod  = 6
		defaultDegree  = 3
		defaultSkip    = 1
	)

	period := p.Period
	if period == 0 {
		period = defaultPeriod
	}

	degree := p.Degree
	if degree == 0 {
		degree = defaultDegree
	}

	skip := p.Skip
	if skip == 0 {
		skip = defaultSkip
	}

	if period < 2 {
		return nil, fmt.Errorf(fmts, invalid, "period should be >= 2")
	}

	if degree < 2 {
		return nil, fmt.Errorf(fmts, invalid, "degree should be >= 2")
	}

	if skip < 1 {
		return nil, fmt.Errorf(fmts, invalid, "skip should be >= 1")
	}

	bc := p.BarComponent
	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	qc := p.QuoteComponent
	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	tc := p.TradeComponent
	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	mnemonic := fmt.Sprintf("mema(%d,%d,%d%s)", period, degree, skip, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Modified exponential moving average " + mnemonic

	bufSize := degree*skip + 1

	mema := &ModifiedExponentialMovingAverage{
		degree:       degree,
		skip:         skip,
		nPoints:      degree + 1,
		coefficients: computeVelocityCoefficients(degree),
		emaAlpha:     2.0 / (float64(period) + 1.0),
		buf:          make([]float64, bufSize),
		bufSize:      bufSize,
	}

	mema.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, mema.Update)

	return mema, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *ModifiedExponentialMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *ModifiedExponentialMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.ModifiedExponentialMovingAverage,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample value and returns the filter output.
func (s *ModifiedExponentialMovingAverage) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// EMA recursion (seed at first sample).
	if !s.emaInitialized {
		s.emaValue = sample
		s.emaInitialized = true
	} else {
		s.emaValue = s.emaAlpha*sample + (1.0-s.emaAlpha)*s.emaValue
	}

	// Store EMA value in the ring buffer.
	s.buf[s.bufPos] = s.emaValue
	s.bufPos = (s.bufPos + 1) % s.bufSize
	s.bufCount++

	if s.bufCount < s.bufSize {
		s.primed = false

		return math.NaN()
	}

	s.primed = true

	// Read EMA values at stride positions and compute the velocity correction.
	velocity := 0.0
	for k := 0; k < s.nPoints; k++ {
		offset := k * s.skip
		idx := ((s.bufPos-1-offset)%s.bufSize + s.bufSize) % s.bufSize
		velocity += s.coefficients[k] * s.buf[idx]
	}

	return s.emaValue + velocity
}
