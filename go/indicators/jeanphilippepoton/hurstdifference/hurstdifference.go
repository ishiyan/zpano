package hurstdifference

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// HurstDifference computes the Hurst Difference (first difference of the corrected FGDI).
//
// Positive values indicate rising volatility (potential trade entry);
// negative values indicate declining volatility.
//
// The FGDI is computed using the corrected FGDI formula with (period-1)
// segments and denominator ln(2*(period-1)).
//
// The indicator is not primed during the first `period` updates.
// The hurst_diff output requires one additional update beyond FGDI priming.
type HurstDifference struct {
	mu sync.RWMutex
	core.LineIndicator
	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc
	window    []float64
	period    int
	nMinus1   int
	winCount  int
	primed    bool
	log2PM1   float64
	ln2       float64
	invNSq    float64
	prevFgdi  float64
	lastFgdi  float64
}

// NewHurstDifference returns an instance of the indicator created using supplied parameters.
func NewHurstDifference(p *Params) (*HurstDifference, error) {
	const (
		invalid = "invalid hurst difference parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "hurdif(%d%s)"
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
	desc := "Hurst difference " + mnemonic

	nMinus1 := period - 1

	ind := &HurstDifference{
		barFunc:   barFunc,
		quoteFunc: quoteFunc,
		tradeFunc: tradeFunc,
		window:    make([]float64, period+1),
		period:    period,
		nMinus1:   nMinus1,
		log2PM1:   math.Log(2.0 * float64(nMinus1)),
		ln2:       math.Log(2.0),
		invNSq:    1.0 / float64(period*period),
		prevFgdi:  math.NaN(),
		lastFgdi:  math.NaN(),
	}

	ind.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ind.Update)

	return ind, nil
}

// IsPrimed indicates whether the indicator is primed.
func (h *HurstDifference) IsPrimed() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.primed
}

// Metadata describes the output data of the indicator.
func (h *HurstDifference) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.HurstDifference,
		h.LineIndicator.Mnemonic,
		h.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: h.LineIndicator.Mnemonic, Description: h.LineIndicator.Description},
			{Mnemonic: h.LineIndicator.Mnemonic + " fgdi", Description: h.LineIndicator.Description + " FGDI"},
		},
	)
}

// Update updates the value of the hurst difference given the next sample.
// Returns the hurst_diff value. Use FgdiValue() for the FGDI output.
func (h *HurstDifference) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	period := h.period

	if h.primed {
		for i := 0; i < period; i++ {
			h.window[i] = h.window[i+1]
		}

		h.window[period] = sample
	} else {
		h.window[h.winCount] = sample
		h.winCount++

		if h.winCount <= period {
			return math.NaN()
		}

		h.primed = true
	}

	// Use the last `period` elements of the window (indices 1..period inclusive).
	// Find min/max for normalization.
	priceMax := h.window[1]
	priceMin := h.window[1]

	for k := 2; k <= period; k++ {
		if h.window[k] > priceMax {
			priceMax = h.window[k]
		}

		if h.window[k] < priceMin {
			priceMin = h.window[k]
		}
	}

	priceRange := priceMax - priceMin

	var fgdi float64

	if priceRange <= 0.0 {
		fgdi = 0.0
	} else {
		// Normalize and compute path length.
		priorNorm := (h.window[1] - priceMin) / priceRange
		length := 0.0

		for k := 2; k <= period; k++ {
			currNorm := (h.window[k] - priceMin) / priceRange
			diff := currNorm - priorNorm
			length += math.Sqrt(diff*diff + h.invNSq)
			priorNorm = currNorm
		}

		if length > 0.0 {
			fgdi = 1.0 + (math.Log(length)+h.ln2)/h.log2PM1
		} else {
			fgdi = 0.0
		}
	}

	// First difference.
	var hurstDiff float64
	if math.IsNaN(h.prevFgdi) {
		hurstDiff = math.NaN()
	} else {
		hurstDiff = fgdi - h.prevFgdi
	}

	h.prevFgdi = fgdi
	h.lastFgdi = fgdi

	return hurstDiff
}

// UpdateAll updates the indicator and returns both outputs: hurstDiff, fgdi.
func (h *HurstDifference) UpdateAll(sample float64) (float64, float64) {
	hurstDiff := h.Update(sample)

	h.mu.RLock()
	defer h.mu.RUnlock()

	return hurstDiff, h.lastFgdi
}

// FgdiValue returns the last computed FGDI value.
func (h *HurstDifference) FgdiValue() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.lastFgdi
}
