package fractaldimensionindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// FractalDimensionIndex computes the Fractal Dimension Index (FDI).
//
// Measures the fractal dimension of a price time series using normalized
// path length. Values near 1.5 indicate a random market, near 1.0 a
// trending market, and near 2.0 a highly volatile market.
//
// The indicator is not primed during the first `period` updates.
type FractalDimensionIndex struct {
	mu sync.RWMutex
	core.LineIndicator
	window      []float64
	period      int
	windowCount int
	primed      bool
	log2N       float64
	ln2         float64
	invNSq      float64
}

// NewFractalDimensionIndex returns an instance of the indicator created using supplied parameters.
func NewFractalDimensionIndex(p *Params) (*FractalDimensionIndex, error) {
	const (
		invalid = "invalid fractal dimension parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "fdi(%d%s)"
		minper  = 2
	)

	period := p.Period
	if period < minper {
		return nil, fmt.Errorf(fmts, invalid, "period should be greater than 1")
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

	mnemonic := fmt.Sprintf(fmtn, period, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Fractal dimension index " + mnemonic

	fdi := &FractalDimensionIndex{
		window: make([]float64, period+1),
		period: period,
		log2N:  math.Log(2.0 * float64(period)),
		ln2:    math.Log(2.0),
		invNSq: 1.0 / float64(period*period),
	}

	fdi.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, fdi.Update)

	return fdi, nil
}

// IsPrimed indicates whether the indicator is primed.
func (f *FractalDimensionIndex) IsPrimed() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.primed
}

// Metadata describes the output data of the indicator.
func (f *FractalDimensionIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.FractalDimensionIndex,
		f.LineIndicator.Mnemonic,
		f.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: f.LineIndicator.Mnemonic, Description: f.LineIndicator.Description},
		},
	)
}

// Update updates the value of the fractal dimension given the next sample.
func (f *FractalDimensionIndex) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	period := f.period

	if f.primed {
		for i := 0; i < period; i++ {
			f.window[i] = f.window[i+1]
		}

		f.window[period] = sample
	} else {
		f.window[f.windowCount] = sample
		f.windowCount++

		if f.windowCount <= period {
			return math.NaN()
		}

		f.primed = true
	}

	// Find min/max for normalization.
	priceMax := f.window[0]
	priceMin := f.window[0]

	for k := 1; k <= period; k++ {
		if f.window[k] > priceMax {
			priceMax = f.window[k]
		}

		if f.window[k] < priceMin {
			priceMin = f.window[k]
		}
	}

	priceRange := priceMax - priceMin
	if priceRange < 1e-10 {
		return 1.0
	}

	// Normalize and compute path length.
	priorNorm := (f.window[0] - priceMin) / priceRange
	length := 0.0

	for k := 1; k <= period; k++ {
		currNorm := (f.window[k] - priceMin) / priceRange
		diff := currNorm - priorNorm
		length += math.Sqrt(diff*diff + f.invNSq)
		priorNorm = currNorm
	}

	return 1.0 + (math.Log(length)+f.ln2)/f.log2N
}
