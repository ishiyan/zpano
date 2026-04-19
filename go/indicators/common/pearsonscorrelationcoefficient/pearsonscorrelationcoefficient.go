package pearsonscorrelationcoefficient

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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
func (c *PearsonsCorrelationCoefficient) IsPrimed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.primed
}

// Metadata describes an output data of the indicator.
func (c *PearsonsCorrelationCoefficient) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.PearsonsCorrelationCoefficient,
		Mnemonic:    c.LineIndicator.Mnemonic,
		Description: c.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(Value),
				Type:        outputs.ScalarType,
				Mnemonic:    c.LineIndicator.Mnemonic,
				Description: c.LineIndicator.Description,
			},
		},
	}
}

// Update updates the indicator given a single scalar sample.
// For a single-input update, both X and Y are set to the same value (degenerate case, always returns 1 or 0).
func (c *PearsonsCorrelationCoefficient) Update(sample float64) float64 {
	return c.UpdatePair(sample, sample)
}

// UpdatePair updates the indicator given an (x, y) pair.
func (c *PearsonsCorrelationCoefficient) UpdatePair(x, y float64) float64 {
	if math.IsNaN(x) || math.IsNaN(y) {
		return math.NaN()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	n := float64(c.length)

	if c.primed {
		// Remove the oldest values.
		oldX := c.windowX[c.pos]
		oldY := c.windowY[c.pos]

		c.sumX -= oldX
		c.sumY -= oldY
		c.sumX2 -= oldX * oldX
		c.sumY2 -= oldY * oldY
		c.sumXY -= oldX * oldY

		// Add new values.
		c.windowX[c.pos] = x
		c.windowY[c.pos] = y
		c.pos = (c.pos + 1) % c.length

		c.sumX += x
		c.sumY += y
		c.sumX2 += x * x
		c.sumY2 += y * y
		c.sumXY += x * y

		return c.correlate(n)
	}

	// Accumulating phase.
	c.windowX[c.count] = x
	c.windowY[c.count] = y

	c.sumX += x
	c.sumY += y
	c.sumX2 += x * x
	c.sumY2 += y * y
	c.sumXY += x * y

	c.count++

	if c.count == c.length {
		c.primed = true
		c.pos = 0

		return c.correlate(n)
	}

	return math.NaN()
}

// correlate computes the Pearson correlation from the running sums.
func (c *PearsonsCorrelationCoefficient) correlate(n float64) float64 {
	varX := c.sumX2 - (c.sumX*c.sumX)/n
	varY := c.sumY2 - (c.sumY*c.sumY)/n
	tempReal := varX * varY

	if tempReal <= 0 {
		return 0
	}

	return (c.sumXY - (c.sumX*c.sumY)/n) / math.Sqrt(tempReal)
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to extract both high (X) and low (Y) from the bar.
func (c *PearsonsCorrelationCoefficient) UpdateBar(sample *entities.Bar) core.Output {
	x := sample.High
	y := sample.Low
	value := c.UpdatePair(x, y)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
