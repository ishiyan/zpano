package arnaudlegouxmovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// ArnaudLegouxMovingAverage computes the Arnaud Legoux Moving Average (ALMA).
//
// ALMA is a Gaussian-weighted moving average that reduces lag while maintaining
// smoothness. It applies a Gaussian bell curve as its kernel, shifted toward
// recent bars via an adjustable offset parameter.
//
// The indicator is not primed during the first (window - 1) updates.
type ArnaudLegouxMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	weights      []float64
	windowLength int
	buffer       []float64
	bufferCount  int
	bufferIndex  int
	primed       bool
}

// NewArnaudLegouxMovingAverage returns an instance of the indicator created using supplied parameters.
func NewArnaudLegouxMovingAverage(p *ArnaudLegouxMovingAverageParams) (*ArnaudLegouxMovingAverage, error) {
	const (
		invalid = "invalid Arnaud Legoux moving average parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "alma(%d, %g, %g%s)"
	)

	window := p.Window
	if window < 1 {
		return nil, fmt.Errorf(fmts, invalid, "window should be greater than 0")
	}

	sigma := p.Sigma
	if sigma <= 0 {
		return nil, fmt.Errorf(fmts, invalid, "sigma should be greater than 0")
	}

	offset := p.Offset
	if offset < 0 || offset > 1 {
		return nil, fmt.Errorf(fmts, invalid, "offset should be between 0 and 1")
	}

	// Resolve defaults for component functions.
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

	mnemonic := fmt.Sprintf(fmtn, window, sigma, offset, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Arnaud Legoux moving average " + mnemonic

	// Precompute Gaussian weights.
	m := offset * float64(window-1)
	s := float64(window) / sigma

	weights := make([]float64, window)
	norm := 0.0

	for i := 0; i < window; i++ {
		diff := float64(i) - m
		w := math.Exp(-(diff * diff) / (2.0 * s * s))
		weights[i] = w
		norm += w
	}

	for i := range weights {
		weights[i] /= norm
	}

	alma := &ArnaudLegouxMovingAverage{
		weights:      weights,
		windowLength: window,
		buffer:       make([]float64, window),
	}

	alma.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, alma.Update)

	return alma, nil
}

// IsPrimed indicates whether the indicator is primed.
func (a *ArnaudLegouxMovingAverage) IsPrimed() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.primed
}

// Metadata describes the output data of the indicator.
func (a *ArnaudLegouxMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.ArnaudLegouxMovingAverage,
		a.LineIndicator.Mnemonic,
		a.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: a.LineIndicator.Mnemonic, Description: a.LineIndicator.Description},
		},
	)
}

// Update updates the value of the Arnaud Legoux moving average given the next sample.
func (a *ArnaudLegouxMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	window := a.windowLength

	if window == 1 {
		a.primed = true
		return sample
	}

	// Fill the circular buffer.
	a.buffer[a.bufferIndex] = sample
	a.bufferIndex = (a.bufferIndex + 1) % window

	if !a.primed {
		a.bufferCount++
		if a.bufferCount < window {
			return math.NaN()
		}

		a.primed = true
	}

	// Compute weighted sum.
	// Weight[0] applies to oldest sample, weight[N-1] to newest.
	// The oldest sample is at a.bufferIndex (circular buffer).
	result := 0.0
	index := a.bufferIndex

	for i := 0; i < window; i++ {
		result += a.weights[i] * a.buffer[index]
		index = (index + 1) % window
	}

	return result
}
