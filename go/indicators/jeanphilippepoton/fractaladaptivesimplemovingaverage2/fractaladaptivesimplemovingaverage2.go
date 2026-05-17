package fractaladaptivesimplemovingaverage2

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// FractalAdaptiveSimpleMovingAverage2 computes the FRASMA2 indicator.
//
// Uses the corrected Fractal Graph Dimension Index (FGDI) formula to adaptively modify
// an SMA's period. The corrected formula uses ln(2*(N-1)) denominator and N-1
// path segments. When the market is trending (FGDI near 1.0), the SMA speeds up;
// when erratic (FGDI near 2.0), the SMA slows down.
//
// The indicator is not primed during the first `period` updates.
type FractalAdaptiveSimpleMovingAverage2 struct {
	mu sync.RWMutex
	core.LineIndicator
	window      []float64
	closes      []float64
	period      int
	normalSpeed int
	windowCount int
	primed      bool
	log2Pm1     float64
	ln2         float64
	invPSq      float64
}

// NewFractalAdaptiveSimpleMovingAverage2 returns an instance of the indicator created using supplied parameters.
func NewFractalAdaptiveSimpleMovingAverage2(p *Params) (*FractalAdaptiveSimpleMovingAverage2, error) {
	const (
		invalid = "invalid fractal adaptive simple moving average 2 parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "frasma2(%d,%d%s)"
	)

	period := p.Period
	if period < 2 {
		return nil, fmt.Errorf(fmts, invalid, "period should be greater than 1")
	}

	normalSpeed := p.NormalSpeed
	if normalSpeed < 1 {
		return nil, fmt.Errorf(fmts, invalid, "normal_speed should be greater than 0")
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

	mnemonic := fmt.Sprintf(fmtn, period, normalSpeed, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Fractal adaptive simple moving average 2 " + mnemonic

	f := &FractalAdaptiveSimpleMovingAverage2{
		window:      make([]float64, period),
		closes:      make([]float64, 0, 256),
		period:      period,
		normalSpeed: normalSpeed,
		log2Pm1:     math.Log(2.0 * float64(period-1)),
		ln2:         math.Log(2.0),
		invPSq:      1.0 / float64(period*period),
	}

	f.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, f.Update)

	return f, nil
}

// IsPrimed indicates whether the indicator is primed.
func (f *FractalAdaptiveSimpleMovingAverage2) IsPrimed() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.primed
}

// Metadata describes the output data of the indicator.
func (f *FractalAdaptiveSimpleMovingAverage2) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.FractalAdaptiveSimpleMovingAverage2,
		f.LineIndicator.Mnemonic,
		f.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: f.LineIndicator.Mnemonic, Description: f.LineIndicator.Description},
		},
	)
}

// Update updates the value of the indicator given the next sample.
func (f *FractalAdaptiveSimpleMovingAverage2) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	period := f.period

	// Accumulate close history for SMA computation.
	f.closes = append(f.closes, sample)

	// Fill the FGDI window. FRASMA2 requires `period+1` samples before first output
	// (first `period` values are NaN, output starts at index `period`).
	if f.windowCount < period {
		f.window[f.windowCount] = sample
		f.windowCount++

		if f.windowCount < period {
			return math.NaN()
		}

		// windowCount == period: we have exactly `period` samples (indices 0..period-1).
		// FRASMA2 outputs NaN for the first `period` samples, so return NaN here
		// and mark primed. The next call will produce the first valid output.
		f.primed = true

		return math.NaN()
	}

	// Shift window.
	for i := 0; i < period-1; i++ {
		f.window[i] = f.window[i+1]
	}

	f.window[period-1] = sample

	// --- Compute FGDI using corrected formula (N-1 segments, ln(2*(N-1)) denominator) ---
	priceMax := f.window[0]
	priceMin := f.window[0]

	for k := 1; k < period; k++ {
		if f.window[k] > priceMax {
			priceMax = f.window[k]
		}

		if f.window[k] < priceMin {
			priceMin = f.window[k]
		}
	}

	priceRange := priceMax - priceMin
	if priceRange <= 0.0 {
		return math.NaN()
	}

	// N-1 segments: iterate from index 1 to period-1.
	priorNorm := (f.window[0] - priceMin) / priceRange
	length := 0.0

	for k := 1; k < period; k++ {
		currNorm := (f.window[k] - priceMin) / priceRange
		diff := currNorm - priorNorm
		length += math.Sqrt(diff*diff + f.invPSq)
		priorNorm = currNorm
	}

	if length <= 0.0 {
		return math.NaN()
	}

	fgdi := 1.0 + (math.Log(length)+f.ln2)/f.log2Pm1

	// --- Adaptive speed ---
	denom := 2.0 - fgdi
	if math.Abs(denom) < 1e-10 {
		return math.NaN()
	}

	trailDim := 1.0 / denom
	alpha := trailDim / 2.0
	speed := int(math.Round(float64(f.normalSpeed) * alpha))

	if speed < 1 {
		speed = 1
	}

	// --- SMA of length `speed` ending at current position ---
	nCloses := len(f.closes)
	if speed > nCloses {
		return math.NaN()
	}

	smaSum := 0.0
	for k := nCloses - speed; k < nCloses; k++ {
		smaSum += f.closes[k]
	}

	return smaSum / float64(speed)
}
