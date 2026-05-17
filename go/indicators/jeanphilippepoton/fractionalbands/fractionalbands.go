package fractionalbands

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// FractionalBands computes the Fractional Bands indicator.
//
// Fractal-adaptive moving average with FBM-scaled volatility bands.
// Uses fractional Brownian motion power law: band_width = 2 * deviation^(2*H)
// where H is the Hurst exponent derived from the Fractal Graph Dimension Index.
//
// The indicator is not primed during the first `period` updates.
type FractionalBands struct {
	mu sync.RWMutex
	core.LineIndicator
	barFunc     entities.BarFunc
	quoteFunc   entities.QuoteFunc
	tradeFunc   entities.TradeFunc
	window      []float64
	closes      []float64
	period      int
	windowSize  int
	priceScale  float64
	windowCount int
	primed      bool
	logDenom    float64
	ln2         float64
	invPeriodSq float64
	frasma2     float64
	upperBand   float64
	lowerBand   float64
}

// NewFractionalBands returns an instance of the indicator created using supplied parameters.
func NewFractionalBands(p *Params) (*FractionalBands, error) {
	const (
		invalid = "invalid fractional bands parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "fctban(%d,%g%s)"
	)

	period := p.Period
	if period < 2 {
		return nil, fmt.Errorf(fmts, invalid, "period should be greater than 1")
	}

	priceScale := p.PriceScale
	if priceScale <= 0.0 {
		return nil, fmt.Errorf(fmts, invalid, "price_scale should be greater than 0")
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

	mnemonic := fmt.Sprintf(fmtn, period, priceScale, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Fractional bands " + mnemonic

	windowSize := period + 1

	f := &FractionalBands{
		barFunc:     barFunc,
		quoteFunc:   quoteFunc,
		tradeFunc:   tradeFunc,
		window:      make([]float64, windowSize),
		closes:      make([]float64, 0, 256),
		period:      period,
		windowSize:  windowSize,
		priceScale:  priceScale,
		logDenom:    math.Log(2.0 * float64(period-1)),
		ln2:         math.Log(2.0),
		invPeriodSq: 1.0 / float64(period*period),
		frasma2:     math.NaN(),
		upperBand:   math.NaN(),
		lowerBand:   math.NaN(),
	}

	f.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, f.Update)

	return f, nil
}

// IsPrimed indicates whether the indicator is primed.
func (f *FractionalBands) IsPrimed() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.primed
}

// Metadata describes the output data of the indicator.
func (f *FractionalBands) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.FractionalBands,
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
func (f *FractionalBands) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	period := f.period
	windowSize := f.windowSize
	p := f.priceScale

	// Accumulate close history.
	f.closes = append(f.closes, sample)

	// Fill the FGDI window (period+1 elements).
	if f.windowCount < windowSize {
		f.window[f.windowCount] = sample
		f.windowCount++

		if f.windowCount < windowSize {
			return math.NaN()
		}

		f.primed = true
	} else {
		for i := 0; i < windowSize-1; i++ {
			f.window[i] = f.window[i+1]
		}

		f.window[windowSize-1] = sample
	}

	// FGDI computation over period+1 points.
	priceMax := f.window[0]
	priceMin := f.window[0]

	for k := 1; k < windowSize; k++ {
		if f.window[k] > priceMax {
			priceMax = f.window[k]
		}

		if f.window[k] < priceMin {
			priceMin = f.window[k]
		}
	}

	priceRange := priceMax - priceMin

	var fgdi float64

	if priceRange < 1e-10 {
		fgdi = 1.0
	} else {
		invRange := 1.0 / priceRange
		prevNorm := (f.window[0] - priceMin) * invRange
		length := 0.0

		for i := 1; i < period; i++ { // period-1 segments
			curNorm := (f.window[i] - priceMin) * invRange
			diff := curNorm - prevNorm
			length += math.Sqrt(diff*diff + f.invPeriodSq)
			prevNorm = curNorm
		}

		if length > 0.0 {
			fgdi = 1.0 + (math.Log(length)+f.ln2)/f.logDenom
		} else {
			fgdi = 1.0
		}
	}

	// Hurst exponent and adaptive speed.
	hurst := 2.0 - fgdi
	if hurst < 0.01 {
		hurst = 0.01
	}

	trailDim := 1.0 / hurst
	beta := trailDim / 2.0
	speed := int(math.Round(float64(period) * beta))

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

	// Deviation in scaled space over last *period* closes.
	devStart := nCloses - period
	frasma2Scaled := p * frasma2Val
	sqSum := 0.0

	for k := devStart; k < nCloses; k++ {
		res := p*f.closes[k] - frasma2Scaled
		sqSum += res * res
	}

	deviation := math.Sqrt(sqSum / float64(period))

	// FBM band offset: 2 * sigma^(2H).
	twoH := 2.0 * hurst
	bandOffset := 2.0 * math.Pow(deviation, twoH)
	upperBand := (frasma2Scaled + bandOffset) / p
	lowerBand := (frasma2Scaled - bandOffset) / p

	f.frasma2 = frasma2Val
	f.upperBand = upperBand
	f.lowerBand = lowerBand

	return frasma2Val
}

// UpdateAll updates the indicator and returns all three outputs: frasma2, upperBand, lowerBand.
func (f *FractionalBands) UpdateAll(sample float64) (float64, float64, float64) {
	frasma2 := f.Update(sample)

	f.mu.RLock()
	defer f.mu.RUnlock()

	return frasma2, f.upperBand, f.lowerBand
}

// Frasma2Value returns the last computed FRASMA2 value.
func (f *FractionalBands) Frasma2Value() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.frasma2
}

// UpperBandValue returns the last computed upper band value.
func (f *FractionalBands) UpperBandValue() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.upperBand
}

// LowerBandValue returns the last computed lower band value.
func (f *FractionalBands) LowerBandValue() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.lowerBand
}

// UpdateScalar updates the indicator given the next scalar sample.
func (f *FractionalBands) UpdateScalar(sample *entities.Scalar) core.Output {
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
func (f *FractionalBands) UpdateBar(sample *entities.Bar) core.Output {
	v := f.barFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (f *FractionalBands) UpdateQuote(sample *entities.Quote) core.Output {
	v := f.quoteFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (f *FractionalBands) UpdateTrade(sample *entities.Trade) core.Output {
	v := f.tradeFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
