package pearsonscorrelationcoefficient

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// PearsonsCorrelationCoefficient computes Pearson's Correlation Coefficient (r) over a rolling window.
//
// Given two input series X and Y, it computes:
//
//	r = (n*sumXY - sumX*sumY) / sqrt((n*sumX2 - sumX^2) * (n*sumY2 - sumY^2))
//
// The indicator is not primed during the first length-1 updates.
type PearsonsCorrelationCoefficient struct {
	mu sync.RWMutex
	core.LineIndicator
	barFunc entities.BarFunc
	length  int
	windowX []float64
	windowY []float64
	count   int
	pos     int
	sumX    float64
	sumY    float64
	sumX2   float64
	sumY2   float64
	sumXY   float64
	primed  bool
}

// New returns an instance of the indicator created using supplied parameters.
func New(p *Params) (*PearsonsCorrelationCoefficient, error) {
	const (
		invalid = "invalid pearsons correlation coefficient parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "correl(%d%s)"
		minlen  = 1
	)

	length := p.Length
	if length < minlen {
		return nil, fmt.Errorf(fmts, invalid, "length should be positive")
	}

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	bc := p.BarComponent
	qc := p.QuoteComponent
	tc := p.TradeComponent

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

	mnemonic := fmt.Sprintf(fmtn, length, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Pearsons Correlation Coefficient " + mnemonic

	c := &PearsonsCorrelationCoefficient{
		barFunc: barFunc,
		length:  length,
		windowX: make([]float64, length),
		windowY: make([]float64, length),
	}

	c.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, c.Update)

	return c, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *PearsonsCorrelationCoefficient) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *PearsonsCorrelationCoefficient) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.PearsonsCorrelationCoefficient,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given a single scalar sample.
// For a single-input update, both X and Y are set to the same value (degenerate case, always returns 1 or 0).
func (s *PearsonsCorrelationCoefficient) Update(sample float64) float64 {
	return s.UpdatePair(sample, sample)
}

// UpdatePair updates the indicator given an (x, y) pair.
func (s *PearsonsCorrelationCoefficient) UpdatePair(x, y float64) float64 {
	if math.IsNaN(x) || math.IsNaN(y) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	n := float64(s.length)

	if s.primed {
		// Remove the oldest values.
		oldX := s.windowX[s.pos]
		oldY := s.windowY[s.pos]

		s.sumX -= oldX
		s.sumY -= oldY
		s.sumX2 -= oldX * oldX
		s.sumY2 -= oldY * oldY
		s.sumXY -= oldX * oldY

		// Add new values.
		s.windowX[s.pos] = x
		s.windowY[s.pos] = y
		s.pos = (s.pos + 1) % s.length

		s.sumX += x
		s.sumY += y
		s.sumX2 += x * x
		s.sumY2 += y * y
		s.sumXY += x * y

		return s.correlate(n)
	}

	// Accumulating phase.
	s.windowX[s.count] = x
	s.windowY[s.count] = y

	s.sumX += x
	s.sumY += y
	s.sumX2 += x * x
	s.sumY2 += y * y
	s.sumXY += x * y

	s.count++

	if s.count == s.length {
		s.primed = true
		s.pos = 0

		return s.correlate(n)
	}

	return math.NaN()
}

// correlate computes the Pearson correlation from the running sums.
func (s *PearsonsCorrelationCoefficient) correlate(n float64) float64 {
	varX := s.sumX2 - (s.sumX*s.sumX)/n
	varY := s.sumY2 - (s.sumY*s.sumY)/n
	tempReal := varX * varY

	if tempReal <= 0 {
		return 0
	}

	return (s.sumXY - (s.sumX*s.sumY)/n) / math.Sqrt(tempReal)
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to extract both high (X) and low (Y) from the bar.
func (s *PearsonsCorrelationCoefficient) UpdateBar(sample *entities.Bar) core.Output {
	x := sample.High
	y := sample.Low
	value := s.UpdatePair(x, y)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
