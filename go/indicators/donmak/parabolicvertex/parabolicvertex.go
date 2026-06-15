package parabolicvertex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// ParabolicVertex is Don Mak's Parabolic Vertex (PVTX).
//
// It predicts turning points by fitting a parabola to the 3 most recent price
// points and computing where the vertex (extremum) occurs relative to the current
// bar. Given three consecutive prices x(n), x(n-1), x(n-2) (most recent first)
// fitted to x(t) = d*t^2 + e*t + f at t = 0, -1, -2, the vertex is at:
//
//	t_v = -(1.5*x(n) - 2*x(n-1) + 0.5*x(n-2)) / (x(n) - 2*x(n-1) + x(n-2))
//
// The output is the number of bars from the current bar to the predicted turning
// point. It works best on pre-smoothed prices.
//
// Reference:
//
// Mak, Don K. (2003). The Science of Financial Market Trading. Ch 7, Appendix 5.
type ParabolicVertex struct {
	mu sync.RWMutex

	core.LineIndicator

	buffer [3]float64
	index  int
	count  int

	primed bool
}

// NewParabolicVertex returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen
func NewParabolicVertex(p *Params) (*ParabolicVertex, error) {
	const (
		invalid = "invalid parabolic vertex parameters"
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

	mnemonic := "pvtx"
	if suffix := core.ComponentTripleMnemonic(bc, qc, tc); suffix != "" {
		mnemonic = "pvtx(" + suffix[2:] + ")" // strip leading ", "
	}

	desc := "Parabolic vertex " + mnemonic

	pvtx := &ParabolicVertex{}

	pvtx.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, pvtx.Update)

	return pvtx, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *ParabolicVertex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *ParabolicVertex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.ParabolicVertex,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample value and returns the filter output.
func (s *ParabolicVertex) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store the price in the ring buffer.
	s.buffer[s.index] = sample
	s.index = (s.index + 1) % 3
	s.count++

	if s.count < 3 {
		s.primed = false

		return math.NaN()
	}

	s.primed = true

	// Extract prices: x[n] (newest), x[n-1], x[n-2] (oldest).
	xn := s.buffer[((s.index-1)%3+3)%3]
	xn1 := s.buffer[((s.index-2)%3+3)%3]
	xn2 := s.buffer[((s.index-3)%3+3)%3]

	// Denominator = second-order finite difference (proportional to curvature).
	denom := xn - 2.0*xn1 + xn2
	if denom == 0.0 {
		return math.NaN()
	}

	numer := 1.5*xn - 2.0*xn1 + 0.5*xn2

	return -numer / denom
}
