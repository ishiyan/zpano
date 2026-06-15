package polynomialforecast

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// computeCoefficients computes FIR coefficients for the order-th derivative of a
// degree-degree polynomial fit evaluated at the most recent point (Lagrange basis).
func computeCoefficients(degree, order int) []float64 {
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

		if order == 1 {
			for ell := 0; ell < len(others); ell++ {
				term := 1.0

				for m := 0; m < len(others); m++ {
					if m != ell {
						term *= float64(others[m])
					}
				}

				numerator += term
			}
		} else {
			for ell := 0; ell < len(others); ell++ {
				for r := ell + 1; r < len(others); r++ {
					term := 2.0

					for m := 0; m < len(others); m++ {
						if m != ell && m != r {
							term *= float64(others[m])
						}
					}

					numerator += term
				}
			}
		}

		coefficients = append(coefficients, numerator/denom)
	}

	return coefficients
}

// PolynomialForecast is Don Mak's Polynomial Forecast (POF).
//
// It is a one-step-ahead price forecast using a Taylor series expansion built on
// polynomial fit derivatives (PFD):
//
//	velocity     = PFD(price, degree, order=1)
//	acceleration = PFD(price, degree, order=2)
//	order=1:  forecast = price + velocity
//	order=2:  forecast = price + velocity + 0.5*acceleration
//
// Reference:
//
// Mak, Don K. (2003). The Science of Financial Market Trading. Ch 10.2.
type PolynomialForecast struct {
	mu sync.RWMutex

	core.LineIndicator

	degree    int
	order     int
	smoothing int
	nPoints   int

	coeffVel []float64
	coeffAcc []float64

	emaAlpha       float64
	emaValue       float64
	emaInitialized bool

	buf      []float64
	bufPos   int
	bufCount int

	primed bool
}

// NewPolynomialForecast returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen
func NewPolynomialForecast(p *Params) (*PolynomialForecast, error) {
	const (
		invalid       = "invalid polynomial forecast parameters"
		fmts          = "%s: %s"
		fmtw          = "%s: %w"
		defaultDegree = 3
		defaultOrder  = 1
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

	if order < 1 || order > 2 {
		return nil, fmt.Errorf(fmts, invalid, "order should be 1 or 2")
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

	mnemonic := fmt.Sprintf("pof(%d,%d,%d%s)", degree, order, smoothing, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Polynomial forecast " + mnemonic

	nPoints := degree + 1

	emaAlpha := 0.0
	if smoothing > 0 {
		emaAlpha = 2.0 / (float64(smoothing) + 1.0)
	}

	var coeffAcc []float64
	if order == 2 {
		coeffAcc = computeCoefficients(degree, 2)
	}

	pof := &PolynomialForecast{
		degree:    degree,
		order:     order,
		smoothing: smoothing,
		nPoints:   nPoints,
		coeffVel:  computeCoefficients(degree, 1),
		coeffAcc:  coeffAcc,
		emaAlpha:  emaAlpha,
		buf:       make([]float64, nPoints),
	}

	pof.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, pof.Update)

	return pof, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *PolynomialForecast) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *PolynomialForecast) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.PolynomialForecast,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample value and returns the filter output.
func (s *PolynomialForecast) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Optional EMA pre-smoothing.
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

	// Store the smoothed price in the ring buffer.
	s.buf[s.bufPos] = smoothed
	s.bufPos = (s.bufPos + 1) % s.nPoints
	s.bufCount++

	if s.bufCount < s.nPoints {
		s.primed = false

		return math.NaN()
	}

	s.primed = true

	// Read buffer most-recent-first and compute velocity (and acceleration).
	velocity := 0.0
	acceleration := 0.0

	for k := 0; k < s.nPoints; k++ {
		idx := ((s.bufPos-1-k)%s.nPoints + s.nPoints) % s.nPoints
		value := s.buf[idx]
		velocity += s.coeffVel[k] * value

		if s.coeffAcc != nil {
			acceleration += s.coeffAcc[k] * value
		}
	}

	forecast := smoothed + velocity
	if s.order == 2 {
		forecast += 0.5 * acceleration
	}

	return forecast
}
