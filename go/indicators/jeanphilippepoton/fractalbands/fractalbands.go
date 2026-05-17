package fractalbands

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// FractalBands computes the Fractal Bands indicator.
//
// FRASMA2 center line with upper/lower bands scaled by alpha^H where H is
// the local Hurst exponent estimated from the Fractal Graph Dimension Index.
//
// The indicator is not primed during the first `period - 1` updates.
type FractalBands struct {
	mu sync.RWMutex
	core.LineIndicator
	barFunc      entities.BarFunc
	quoteFunc    entities.QuoteFunc
	tradeFunc    entities.TradeFunc
	window       []float64
	closes       []float64
	period       int
	periodMinus1 int
	normalSpeed  int
	alpha        float64
	windowCount  int
	primed       bool
	logDenom     float64
	ln2          float64
	invPeriodSq  float64
	frasma2      float64
	upperBand    float64
	lowerBand    float64
}

// NewFractalBands returns an instance of the indicator created using supplied parameters.
func NewFractalBands(p *Params) (*FractalBands, error) {
	const (
		invalid = "invalid fractal bands parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "fban(%d,%d,%g%s)"
	)

	period := p.Period
	if period < 2 {
		return nil, fmt.Errorf(fmts, invalid, "period should be greater than 1")
	}

	normalSpeed := p.NormalSpeed
	if normalSpeed < 1 {
		return nil, fmt.Errorf(fmts, invalid, "normal_speed should be greater than 0")
	}

	alpha := p.Alpha
	if alpha <= 0.0 {
		return nil, fmt.Errorf(fmts, invalid, "alpha should be greater than 0")
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

	mnemonic := fmt.Sprintf(fmtn, period, normalSpeed, alpha, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Fractal bands " + mnemonic

	periodMinus1 := period - 1

	f := &FractalBands{
		barFunc:      barFunc,
		quoteFunc:    quoteFunc,
		tradeFunc:    tradeFunc,
		window:       make([]float64, period),
		closes:       make([]float64, 0, 256),
		period:       period,
		periodMinus1: periodMinus1,
		normalSpeed:  normalSpeed,
		alpha:        alpha,
		logDenom:     math.Log(2.0 * float64(periodMinus1)),
		ln2:          math.Log(2.0),
		invPeriodSq:  1.0 / float64(period*period),
		frasma2:      math.NaN(),
		upperBand:    math.NaN(),
		lowerBand:    math.NaN(),
	}

	f.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, f.Update)

	return f, nil
}

// IsPrimed indicates whether the indicator is primed.
func (f *FractalBands) IsPrimed() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.primed
}

// Metadata describes the output data of the indicator.
func (f *FractalBands) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.FractalBands,
		f.LineIndicator.Mnemonic,
		f.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: f.LineIndicator.Mnemonic, Description: f.LineIndicator.Description},
			{Mnemonic: f.LineIndicator.Mnemonic + " upper", Description: f.LineIndicator.Description + " Upper Band"},
			{Mnemonic: f.LineIndicator.Mnemonic + " lower", Description: f.LineIndicator.Description + " Lower Band"},
			{Mnemonic: f.LineIndicator.Mnemonic + " band", Description: f.LineIndicator.Description + " Band"},
		},
	)
}

// Update updates the value of the indicator given the next sample.
// Returns the FRASMA2 value. Use UpdateAll() for all three outputs.
func (f *FractalBands) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	period := f.period
	periodMinus1 := f.periodMinus1

	// Accumulate close history for SMA computation.
	f.closes = append(f.closes, sample)

	// Fill the FGDI window.
	if f.windowCount < period {
		f.window[f.windowCount] = sample
		f.windowCount++

		if f.windowCount < period {
			return math.NaN()
		}

		f.primed = true
	} else {
		for i := 0; i < periodMinus1; i++ {
			f.window[i] = f.window[i+1]
		}

		f.window[periodMinus1] = sample
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

	var fgdi float64

	if priceRange <= 0.0 {
		fgdi = 0.0
	} else {
		// Compute normalized path length: period points, period-1 segments.
		priorNorm := (f.window[0] - priceMin) / priceRange
		length := 0.0

		for k := 1; k < period; k++ {
			currNorm := (f.window[k] - priceMin) / priceRange
			diff := currNorm - priorNorm
			length += math.Sqrt(diff*diff + f.invPeriodSq)
			priorNorm = currNorm
		}

		if length > 0.0 {
			fgdi = 1.0 + (math.Log(length)+f.ln2)/f.logDenom
		} else {
			fgdi = 0.0
		}
	}

	// Hurst exponent.
	hurst := 2.0 - fgdi
	if hurst < 0.01 {
		hurst = 0.01
	}

	trailDim := 1.0 / hurst
	beta := trailDim / 2.0
	speed := int(math.Round(float64(f.normalSpeed) * beta))

	if speed < 1 {
		speed = 1
	}

	// FRASMA2: SMA of close over 'speed' bars ending at current position.
	nCloses := len(f.closes)
	if speed > nCloses {
		f.frasma2 = math.NaN()
		f.upperBand = math.NaN()
		f.lowerBand = math.NaN()

		return math.NaN()
	}

	smaSum := 0.0
	for k := nCloses - speed; k < nCloses; k++ {
		smaSum += f.closes[k]
	}

	frasma2Val := smaSum / float64(speed)

	// Deviation over the FGDI lookback window (period bars).
	sqSum := 0.0
	for k := 0; k < period; k++ {
		res := f.window[k] - frasma2Val
		sqSum += res * res
	}

	deviation := 2.0 * math.Sqrt(sqSum/float64(period))

	// Fractal bands.
	bandMult := deviation * math.Pow(f.alpha, hurst)
	upperBand := frasma2Val + bandMult
	lowerBand := frasma2Val - bandMult

	f.frasma2 = frasma2Val
	f.upperBand = upperBand
	f.lowerBand = lowerBand

	return frasma2Val
}

// UpdateAll updates the indicator and returns all three outputs: frasma2, upperBand, lowerBand.
func (f *FractalBands) UpdateAll(sample float64) (float64, float64, float64) {
	frasma2 := f.Update(sample)

	f.mu.RLock()
	defer f.mu.RUnlock()

	return frasma2, f.upperBand, f.lowerBand
}

// Frasma2Value returns the last computed FRASMA2 value.
func (f *FractalBands) Frasma2Value() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.frasma2
}

// UpperBandValue returns the last computed upper band value.
func (f *FractalBands) UpperBandValue() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.upperBand
}

// LowerBandValue returns the last computed lower band value.
func (f *FractalBands) LowerBandValue() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.lowerBand
}

// UpdateScalar updates the indicator given the next scalar sample.
func (f *FractalBands) UpdateScalar(sample *entities.Scalar) core.Output {
	frasma2, upper, lower := f.UpdateAll(sample.Value)

	const outputCount = 4

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: frasma2}
	output[1] = entities.Scalar{Time: sample.Time, Value: upper}
	output[2] = entities.Scalar{Time: sample.Time, Value: lower}

	if math.IsNaN(lower) || math.IsNaN(upper) {
		output[3] = outputs.NewEmptyBand(sample.Time)
	} else {
		output[3] = outputs.NewBand(sample.Time, lower, upper)
	}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (f *FractalBands) UpdateBar(sample *entities.Bar) core.Output {
	v := f.barFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (f *FractalBands) UpdateQuote(sample *entities.Quote) core.Output {
	v := f.quoteFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (f *FractalBands) UpdateTrade(sample *entities.Trade) core.Output {
	v := f.tradeFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
