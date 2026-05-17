package rescaledfractaladaptivesimplemovingaverage

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// RescaledFractalAdaptiveSimpleMovingAverage computes the RS-FRASMA indicator.
//
// Uses Rescaled Range (R/S) analysis to estimate the Hurst exponent,
// then adapts the SMA period accordingly.
//
// The indicator is not primed during the first `period` updates.
type RescaledFractalAdaptiveSimpleMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	closes      []float64
	period      int
	normalSpeed int
	priceScale  float64
	primed      bool
	nIter       int
	blockSizes  []int
	blockCounts []int
}

// NewRescaledFractalAdaptiveSimpleMovingAverage returns an instance of the indicator created using supplied parameters.
func NewRescaledFractalAdaptiveSimpleMovingAverage(p *Params) (*RescaledFractalAdaptiveSimpleMovingAverage, error) {
	const (
		invalid = "invalid RS fractal adaptive simple moving average parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "rsfrasma(%d,%d,%.1f%s)"
	)

	period := p.Period
	if period < 4 {
		return nil, fmt.Errorf(fmts, invalid, "period should be greater than 3")
	}

	if period&(period-1) != 0 {
		return nil, fmt.Errorf(fmts, invalid, "period must be a power of 2")
	}

	normalSpeed := p.NormalSpeed
	if normalSpeed < 1 {
		return nil, fmt.Errorf(fmts, invalid, "normal_speed should be greater than 0")
	}

	priceScale := p.PriceScale
	if priceScale == 0 {
		priceScale = 1.0
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

	mnemonic := fmt.Sprintf(fmtn, period, normalSpeed, priceScale, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "RS fractal adaptive simple moving average " + mnemonic

	// Precompute R/S parameters.
	k0 := period / 4
	nIter := 0

	if k0 >= 2 {
		nIter = int(math.Floor(math.Log(float64(k0)) / math.Log(2)))
	}

	blockSizes := make([]int, nIter+1)
	blockCounts := make([]int, nIter+1)

	for u := 1; u <= nIter; u++ {
		blockSizes[u] = 1 << (u + 1)
		blockCounts[u] = period / blockSizes[u]
	}

	f := &RescaledFractalAdaptiveSimpleMovingAverage{
		closes:      make([]float64, 0, 256),
		period:      period,
		normalSpeed: normalSpeed,
		priceScale:  priceScale,
		nIter:       nIter,
		blockSizes:  blockSizes,
		blockCounts: blockCounts,
	}

	f.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, f.Update)

	return f, nil
}

// IsPrimed indicates whether the indicator is primed.
func (f *RescaledFractalAdaptiveSimpleMovingAverage) IsPrimed() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.primed
}

// Metadata describes the output data of the indicator.
func (f *RescaledFractalAdaptiveSimpleMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.RescaledFractalAdaptiveSimpleMovingAverage,
		f.LineIndicator.Mnemonic,
		f.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: f.LineIndicator.Mnemonic, Description: f.LineIndicator.Description},
		},
	)
}

// Update updates the value of the indicator given the next sample.
func (f *RescaledFractalAdaptiveSimpleMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	period := f.period
	priceScale := f.priceScale

	// Accumulate close history.
	f.closes = append(f.closes, sample)
	nCloses := len(f.closes)

	// Need at least period+1 closes.
	if nCloses <= period {
		return math.NaN()
	}

	if !f.primed {
		f.primed = true
	}

	pos := nCloses - 1

	// R/S analysis.
	nIter := f.nIter
	sumx := 0.0
	sumy := 0.0
	sumx2 := 0.0
	sumxy := 0.0
	validScales := 0

	for u := 1; u <= nIter; u++ {
		blockSize := f.blockSizes[u]
		nBlocksU := f.blockCounts[u]

		if nBlocksU < 1 {
			continue
		}

		rsSum := 0.0
		t := 0
		blockCount := 0

		for t <= period-blockSize {
			// Block mean.
			mu := 0.0
			for j := 1; j <= blockSize; j++ {
				mu += priceScale * f.closes[pos-(t+j)]
			}

			mu /= float64(blockSize)

			// Population std.
			sumSq := 0.0
			for j := 1; j <= blockSize; j++ {
				diff := priceScale*f.closes[pos-(t+j)] - mu
				sumSq += diff * diff
			}

			std := math.Sqrt(sumSq / float64(blockSize))
			if std <= 0.0 {
				std = 0.1
			}

			// Cumulative deviations and range.
			cumDev := 0.0
			wMax := 0.0
			wMin := 9999999999.0

			for k := 1; k <= blockSize; k++ {
				cumDev += priceScale*f.closes[pos-(t+k)] - mu
				if cumDev > wMax {
					wMax = cumDev
				}

				if cumDev < wMin {
					wMin = cumDev
				}
			}

			if wMax < 0.0 {
				wMax = 0.0
			}

			if wMin > 0.0 {
				wMin = 0.0
			}

			rVal := wMax - wMin
			rsSum += rVal / std
			t += blockSize
			blockCount++
		}

		// Average R/S for this scale.
		rsAvg := 1.0
		if blockCount > 0 {
			rsAvg = rsSum / float64(blockCount)
		}

		if rsAvg <= 0.0 {
			rsAvg = 1e-10
		}

		log2D := math.Log(float64(blockSize)) / math.Log(2)
		log2Rs := math.Log(rsAvg) / math.Log(2)

		sumx += log2D
		sumy += log2Rs
		sumx2 += log2D * log2D
		sumxy += log2D * log2Rs
		validScales++
	}

	// Linear regression slope = Hurst exponent.
	h := 0.5

	if validScales >= 2 {
		h1 := float64(validScales)*sumxy - sumx*sumy
		h2 := float64(validScales)*sumx2 - sumx*sumx

		if h2 <= 0.0 {
			h2 = 0.1
		}

		h = h1 / h2
	}

	// Guard H.
	if 2.0*h <= 0.0 {
		h = 0.001
	}

	alpha := 1.0 / (2.0 * h)
	spd := int(math.Round(float64(f.normalSpeed) * alpha))

	if spd < 1 {
		spd = 1
	}

	// Compute SMA with adapted speed.
	smaStart := pos - spd + 1
	if smaStart < 0 {
		smaStart = 0
	}

	total := 0.0
	count := pos - smaStart + 1

	for i := smaStart; i <= pos; i++ {
		total += f.closes[i]
	}

	return total / float64(count)
}
