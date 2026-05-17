package fractalgeneralizeddimensionindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// FractalGraphDimensionIndex computes the Fractal Graph Dimension Index (FGDI).
//
// This is Poton's corrected and enhanced version of the Fractal Dimension
// Index (FDI). It fixes loop boundary and denominator bugs in the original
// and adds standard deviation bands around the estimated dimension.
//
// The indicator is not primed during the first `period - 1` updates.
type FractalGraphDimensionIndex struct {
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
	log2N1    float64
	ln2       float64
	invNSq    float64
	fgdi      float64
	upper     float64
	lower     float64
	stddev    float64
}

// NewFractalGraphDimensionIndex returns an instance of the indicator created using supplied parameters.
func NewFractalGraphDimensionIndex(p *Params) (*FractalGraphDimensionIndex, error) {
	const (
		invalid = "invalid fractal graph dimension index parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "fgdi(%d%s)"
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
	desc := "Fractal graph dimension index " + mnemonic

	nMinus1 := period - 1

	ind := &FractalGraphDimensionIndex{
		barFunc:   barFunc,
		quoteFunc: quoteFunc,
		tradeFunc: tradeFunc,
		window:    make([]float64, period),
		period:    period,
		nMinus1:   nMinus1,
		log2N1:    math.Log(2.0 * float64(nMinus1)),
		ln2:       math.Log(2.0),
		invNSq:    1.0 / float64(period*period),
		fgdi:       math.NaN(),
		upper:     math.NaN(),
		lower:     math.NaN(),
		stddev:    math.NaN(),
	}

	ind.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ind.Update)

	return ind, nil
}

// IsPrimed indicates whether the indicator is primed.
func (f *FractalGraphDimensionIndex) IsPrimed() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.primed
}

// Metadata describes the output data of the indicator.
func (f *FractalGraphDimensionIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.FractalGraphDimensionIndex,
		f.LineIndicator.Mnemonic,
		f.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: f.LineIndicator.Mnemonic, Description: f.LineIndicator.Description},
			{Mnemonic: f.LineIndicator.Mnemonic + " upper", Description: f.LineIndicator.Description + " Upper"},
			{Mnemonic: f.LineIndicator.Mnemonic + " lower", Description: f.LineIndicator.Description + " Lower"},
			{Mnemonic: f.LineIndicator.Mnemonic + " stddev", Description: f.LineIndicator.Description + " Stddev"},
			{Mnemonic: f.LineIndicator.Mnemonic + " band", Description: f.LineIndicator.Description + " Band"},
		},
	)
}

// Update updates the value of the fractal graph dimension index given the next sample.
// Returns the FGDI value. Use Fgdi(), Upper(), Lower(), Stddev() for all outputs.
func (f *FractalGraphDimensionIndex) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	period := f.period
	nMinus1 := f.nMinus1

	if f.primed {
		for i := 0; i < nMinus1; i++ {
			f.window[i] = f.window[i+1]
		}

		f.window[nMinus1] = sample
	} else {
		f.window[f.winCount] = sample
		f.winCount++

		if f.winCount < period {
			return math.NaN()
		}

		f.primed = true
	}

	// Find min/max for normalization.
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
	if priceRange < 1e-10 {
		f.fgdi = 1.0
		f.stddev = 0.0
		f.upper = 1.0
		f.lower = 1.0

		return 1.0
	}

	// Normalize and compute path segments.
	priorNorm := (f.window[0] - priceMin) / priceRange
	length := 0.0
	segments := make([]float64, nMinus1)

	for k := 1; k < period; k++ {
		currNorm := (f.window[k] - priceMin) / priceRange
		diff := currNorm - priorNorm
		seg := math.Sqrt(diff*diff + f.invNSq)
		segments[k-1] = seg
		length += seg
		priorNorm = currNorm
	}

	// FGDI = 1 + (ln(L) + ln(2)) / ln(2*(N-1))
	fgdi := 1.0 + (math.Log(length)+f.ln2)/f.log2N1

	// Standard deviation of the estimate.
	meanSeg := length / float64(nMinus1)
	sumSq := 0.0

	for k := 0; k < nMinus1; k++ {
		d := segments[k] - meanSeg
		sumSq += d * d
	}

	variance := sumSq / (length * length * f.log2N1 * f.log2N1)
	stddev := math.Sqrt(variance)

	f.fgdi = fgdi
	f.upper = fgdi + stddev
	f.lower = fgdi - stddev
	f.stddev = stddev

	return fgdi
}

// UpdateAll updates the indicator and returns all four outputs: fgdi, upper, lower, stddev.
func (f *FractalGraphDimensionIndex) UpdateAll(sample float64) (float64, float64, float64, float64) {
	fgdi := f.Update(sample)

	f.mu.RLock()
	defer f.mu.RUnlock()

	return fgdi, f.upper, f.lower, f.stddev
}

// FgdiValue returns the last computed FGDI value.
func (f *FractalGraphDimensionIndex) FgdiValue() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.fgdi
}

// UpperValue returns the last computed upper band value.
func (f *FractalGraphDimensionIndex) UpperValue() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.upper
}

// LowerValue returns the last computed lower band value.
func (f *FractalGraphDimensionIndex) LowerValue() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.lower
}

// StddevValue returns the last computed stddev value.
func (f *FractalGraphDimensionIndex) StddevValue() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.stddev
}

// UpdateScalar updates the indicator given the next scalar sample.
func (f *FractalGraphDimensionIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	fgdi, upper, lower, stddev := f.UpdateAll(sample.Value)

	const outputCount = 5

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: fgdi}
	output[1] = entities.Scalar{Time: sample.Time, Value: upper}
	output[2] = entities.Scalar{Time: sample.Time, Value: lower}
	output[3] = entities.Scalar{Time: sample.Time, Value: stddev}

	if math.IsNaN(lower) || math.IsNaN(upper) {
		output[4] = outputs.NewEmptyBand(sample.Time)
	} else {
		output[4] = outputs.NewBand(sample.Time, lower, upper)
	}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (f *FractalGraphDimensionIndex) UpdateBar(sample *entities.Bar) core.Output {
	v := f.barFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (f *FractalGraphDimensionIndex) UpdateQuote(sample *entities.Quote) core.Output {
	v := f.quoteFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (f *FractalGraphDimensionIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := f.tradeFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
