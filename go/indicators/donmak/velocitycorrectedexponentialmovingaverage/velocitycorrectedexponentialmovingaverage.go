package velocitycorrectedexponentialmovingaverage

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

// VelocityCorrectedExponentialMovingAverage is Don Mak's Velocity-Corrected Exponential Moving Average (VCEMA).
//
// It is a reduced-lag EMA that pre-corrects price by adding its polynomial velocity
// before smoothing:
//
//	corrected = price + PFD(price, degree, order=1)
//	VCEMA(n)  = EMA(corrected, n)
//
// Reference:
//
// Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading. Ch 4.1.
type VelocityCorrectedExponentialMovingAverage struct {
	mu sync.RWMutex

	core.LineIndicator

	degree  int
	nPoints int

	coefficients []float64

	emaAlpha       float64
	emaValue       float64
	emaInitialized bool

	buf      []float64
	bufPos   int
	bufCount int

	primed bool
}

// NewVelocityCorrectedExponentialMovingAverage returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen
func NewVelocityCorrectedExponentialMovingAverage(p *Params) (*VelocityCorrectedExponentialMovingAverage, error) {
	const (
		invalid       = "invalid velocity-corrected exponential moving average parameters"
		fmts          = "%s: %s"
		fmtw          = "%s: %w"
		defaultPeriod = 6
		defaultDegree = 3
	)

	period := p.Period
	if period == 0 {
		period = defaultPeriod
	}

	degree := p.Degree
	if degree == 0 {
		degree = defaultDegree
	}

	if period < 2 {
		return nil, fmt.Errorf(fmts, invalid, "period should be >= 2")
	}

	if degree < 2 {
		return nil, fmt.Errorf(fmts, invalid, "degree should be >= 2")
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

	mnemonic := fmt.Sprintf("vcema(%d,%d%s)", period, degree, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Velocity-corrected exponential moving average " + mnemonic

	nPoints := degree + 1

	vcema := &VelocityCorrectedExponentialMovingAverage{
		degree:       degree,
		nPoints:      nPoints,
		coefficients: computeVelocityCoefficients(degree),
		emaAlpha:     2.0 / (float64(period) + 1.0),
		buf:          make([]float64, nPoints),
	}

	vcema.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, vcema.Update)

	return vcema, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *VelocityCorrectedExponentialMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *VelocityCorrectedExponentialMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.VelocityCorrectedExponentialMovingAverage,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample value and returns the filter output.
func (s *VelocityCorrectedExponentialMovingAverage) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store the raw price in the ring buffer.
	s.buf[s.bufPos] = sample
	s.bufPos = (s.bufPos + 1) % s.nPoints
	s.bufCount++

	if s.bufCount < s.nPoints {
		s.primed = false

		return math.NaN()
	}

	s.primed = true

	// Compute the velocity from the raw prices.
	velocity := 0.0
	for k := 0; k < s.nPoints; k++ {
		idx := ((s.bufPos-1-k)%s.nPoints + s.nPoints) % s.nPoints
		velocity += s.coefficients[k] * s.buf[idx]
	}

	// Corrected price = price + velocity.
	corrected := sample + velocity

	// Apply the EMA to the corrected price (seed at the first corrected value).
	if !s.emaInitialized {
		s.emaValue = corrected
		s.emaInitialized = true
	} else {
		s.emaValue = s.emaAlpha*corrected + (1.0-s.emaAlpha)*s.emaValue
	}

	return s.emaValue
}
