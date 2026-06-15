package cubicvertex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// CubicVertex is Don Mak's Cubic Vertex (CVTX).
//
// It predicts turning points by fitting a cubic polynomial to the 4 most recent
// price points and computing where the two vertices (extrema) occur relative to the
// current bar. Given four consecutive prices x(n), x(n-1), x(n-2), x(n-3) (most
// recent first), the cubic coefficients are (Eq 7.2a-c):
//
//	c = (x(n) - 3*x(n-1) + 3*x(n-2) - x(n-3)) / 6
//	d = (2*x(n) - 5*x(n-1) + 4*x(n-2) - x(n-3)) / 2
//	e = (11*x(n) - 18*x(n-1) + 9*x(n-2) - 2*x(n-3)) / 6
//
// The vertex locations are the roots of 3c*t^2 + 2d*t + e = 0. The near root has the
// smaller absolute value (more imminent turn); the far root the larger. It works best
// on pre-smoothed prices.
//
// Reference:
//
// Mak, Don K. (2003). The Science of Financial Market Trading. Ch 7, Appendix 5.
type CubicVertex struct {
	mu sync.RWMutex

	mnemonic    string
	description string

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	buffer [4]float64
	index  int
	count  int

	primed bool
}

// NewCubicVertex returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen
func NewCubicVertex(p *Params) (*CubicVertex, error) {
	const (
		invalid = "invalid cubic vertex parameters"
		fmtw    = "%s: %w"
	)

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

	mnemonic := "cvtx"
	if suffix := core.ComponentTripleMnemonic(bc, qc, tc); suffix != "" {
		mnemonic = "cvtx(" + suffix[2:] + ")" // strip leading ", "
	}

	return &CubicVertex{
		mnemonic:    mnemonic,
		description: "Cubic vertex " + mnemonic,
		barFunc:     barFunc,
		quoteFunc:   quoteFunc,
		tradeFunc:   tradeFunc,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *CubicVertex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *CubicVertex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CubicVertex,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " near", Description: s.description + " near turn"},
			{Mnemonic: s.mnemonic + " far", Description: s.description + " far turn"},
		},
	)
}

// Update updates the indicator given the next sample value and returns
// (bars to near turn, bars to far turn).
func (s *CubicVertex) Update(sample float64) (float64, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nan := math.NaN()

	// Store the price in the ring buffer.
	s.buffer[s.index] = sample
	s.index = (s.index + 1) % 4
	s.count++

	if s.count < 4 {
		s.primed = false

		return nan, nan
	}

	s.primed = true

	// Extract prices: x[n] (newest), x[n-1], x[n-2], x[n-3] (oldest).
	xn := s.buffer[((s.index-1)%4+4)%4]
	xn1 := s.buffer[((s.index-2)%4+4)%4]
	xn2 := s.buffer[((s.index-3)%4+4)%4]
	xn3 := s.buffer[((s.index-4)%4+4)%4]

	// Cubic polynomial coefficients (Eq 7.2a-c).
	c := (xn - 3.0*xn1 + 3.0*xn2 - xn3) / 6.0
	d := (2.0*xn - 5.0*xn1 + 4.0*xn2 - xn3) / 2.0
	e := (11.0*xn - 18.0*xn1 + 9.0*xn2 - 2.0*xn3) / 6.0

	// Case: c == 0 -- cubic term vanishes, reduces to parabola or line.
	if c == 0.0 {
		if d == 0.0 {
			return nan, nan
		}

		vertex := -e / (2.0 * d)

		return vertex, nan
	}

	// Full cubic: solve quadratic 3c*t^2 + 2d*t + e = 0.
	disc := d*d - 3.0*c*e

	if disc < 0.0 {
		return nan, nan
	}

	if disc == 0.0 {
		vertex := -d / (3.0 * c)

		return vertex, vertex
	}

	sqrtDisc := math.Sqrt(disc)
	threeC := 3.0 * c

	tPlus := (-d + sqrtDisc) / threeC
	tMinus := (-d - sqrtDisc) / threeC

	if math.Abs(tPlus) <= math.Abs(tMinus) {
		return tPlus, tMinus
	}

	return tMinus, tPlus
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CubicVertex) UpdateScalar(sample *entities.Scalar) core.Output {
	near, far := s.Update(sample.Value)

	const outputCount = 2

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: near}
	output[1] = entities.Scalar{Time: sample.Time, Value: far}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CubicVertex) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CubicVertex) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CubicVertex) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
