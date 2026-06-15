package polynomialfitderivative

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// computeCoefficients computes the FIR filter coefficients for the order-th
// derivative of a degree-degree polynomial fit, evaluated at the most recent
// point. Uses the Lagrange basis with the elementary-symmetric-polynomial
// identity:
//
//	c_i = order! * e_{degree-order}(others) / prod_{j != i} (j - i)
//
// where others is the set of point positions {0..degree} excluding i.
func computeCoefficients(degree, order int) []float64 {
	nPoints := degree + 1

	factorialOrder := 1
	for f := 2; f <= order; f++ {
		factorialOrder *= f
	}

	coefficients := make([]float64, 0, nPoints)

	for i := 0; i < nPoints; i++ {
		denom := 1.0

		for j := 0; j < nPoints; j++ {
			if j != i {
				denom *= float64(j - i)
			}
		}

		// Build the "others" list: {0..degree} excluding i.
		others := make([]int, 0, degree)

		for j := 0; j < nPoints; j++ {
			if j != i {
				others = append(others, j)
			}
		}

		m := len(others) // equals degree

		// Elementary symmetric polynomials e[0..m] of the values in others.
		e := make([]float64, m+1)
		e[0] = 1.0

		for _, v := range others {
			for k := m; k >= 1; k-- {
				e[k] += float64(v) * e[k-1]
			}
		}

		numerator := float64(factorialOrder) * e[m-order]
		coefficients = append(coefficients, numerator/denom)
	}

	return coefficients
}

// PolynomialFitDerivative is Don Mak's Polynomial Fit Derivative (PFD) indicator.
//
// It fits a polynomial of degree Degree to the most recent Degree+1 (optionally
// EMA-smoothed) prices and evaluates its Order-th derivative at the current bar.
// This is a FIR filter: a dot product of fixed Lagrange-derived coefficients
// with the last Degree+1 smoothed prices.
//
// Reference:
//
// Mak, Don K. (2003). The Science of Financial Market Trading. Ch 6.
// Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading. Ch 8.
type PolynomialFitDerivative struct {
	mu sync.RWMutex

	core.LineIndicator

	coefficients []float64
	nPoints      int

	smoothing      int
	emaAlpha       float64
	emaValue       float64
	emaInitialized bool

	buf      []float64
	bufPos   int
	bufCount int

	primed bool
}

// NewPolynomialFitDerivative returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewPolynomialFitDerivative(p *Params) (*PolynomialFitDerivative, error) {
	const (
		invalid          = "invalid polynomial fit derivative parameters"
		fmts             = "%s: %s"
		fmtw             = "%s: %w"
		fmtn             = "pfd(%d,%d,%d%s)"
		defaultDegree    = 3
		defaultOrder     = 1
		defaultSmoothing = 6
	)

	degree := p.Degree
	if degree == 0 {
		degree = defaultDegree
	}

	order := p.Order
	if order == 0 {
		order = defaultOrder
	}

	smoothing := p.Smoothing

	if degree < 2 {
		return nil, fmt.Errorf(fmts, invalid, "degree should be >= 2")
	}

	if order < 1 || order > degree {
		return nil, fmt.Errorf(fmts, invalid, "order should be >= 1 and <= degree")
	}

	if smoothing < 0 {
		return nil, fmt.Errorf(fmts, invalid, "smoothing should be >= 0")
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

	emaAlpha := 0.0
	if smoothing > 0 {
		emaAlpha = 2.0 / (float64(smoothing) + 1.0)
	}

	mnemonic := fmt.Sprintf(fmtn, degree, order, smoothing, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Polynomial fit derivative " + mnemonic

	pfd := &PolynomialFitDerivative{
		coefficients: computeCoefficients(degree, order),
		nPoints:      degree + 1,
		smoothing:    smoothing,
		emaAlpha:     emaAlpha,
		buf:          make([]float64, degree+1),
	}

	pfd.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, pfd.Update)

	return pfd, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *PolynomialFitDerivative) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *PolynomialFitDerivative) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.PolynomialFitDerivative,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample value and returns the FIR output.
func (s *PolynomialFitDerivative) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Step 1: optional EMA smoothing.
	smoothed := sample
	if s.smoothing > 0 {
		if !s.emaInitialized {
			s.emaValue = sample
			s.emaInitialized = true
		} else {
			s.emaValue = s.emaAlpha*sample + (1.0-s.emaAlpha)*s.emaValue
		}

		smoothed = s.emaValue
	}

	// Step 2: push into the ring buffer.
	s.buf[s.bufPos] = smoothed
	s.bufPos = (s.bufPos + 1) % s.nPoints
	s.bufCount++

	// Step 3: not enough data yet.
	if s.bufCount < s.nPoints {
		s.primed = false

		return math.NaN()
	}

	// Step 4: FIR dot product (coefficients[j] multiplies the j-th most recent).
	result := 0.0
	for j := 0; j < s.nPoints; j++ {
		bufIdx := ((s.bufPos-1-j)%s.nPoints + s.nPoints) % s.nPoints
		result += s.coefficients[j] * s.buf[bufIdx]
	}

	s.primed = true

	return result
}
