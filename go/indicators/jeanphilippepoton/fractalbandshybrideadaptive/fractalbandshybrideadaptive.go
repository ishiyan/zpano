package fractalbandshybrideadaptive

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// FractalBandsHybrideAdaptive computes the Fractal Bands Hybride Adaptive indicator.
//
// Hybrid variant of Fractal Bands that replaces fixed normal_speed with
// Ehlers' CyclePeriod indicator output multiplied by a Nyquist factor,
// making the FRASMA2 doubly adaptive to both fractal dimension and
// dominant market cycle.
//
// The indicator is not primed during the first `period` updates.
type FractalBandsHybrideAdaptive struct {
	mu sync.RWMutex
	core.LineIndicator
	barFunc            entities.BarFunc
	quoteFunc          entities.QuoteFunc
	tradeFunc          entities.TradeFunc
	window             []float64
	closes             []float64
	period             int
	windowSize         int
	normalSpeedFallback int
	alpha              float64
	nyquist            float64
	alphaHP            float64
	windowCount        int
	primed             bool
	logDenom           float64
	ln2                float64
	invPeriodSq        float64
	// Ehlers CyclePeriod buffers.
	smoothBuf    []float64
	cycleBuf     []float64
	q1Buf        []float64
	i1Buf        []float64
	dpBuf        []float64
	instPeriodBuf []float64
	// Last computed values.
	frasma2   float64
	upperBand float64
	lowerBand float64
}

// NewFractalBandsHybrideAdaptive returns an instance of the indicator created using supplied parameters.
func NewFractalBandsHybrideAdaptive(p *Params) (*FractalBandsHybrideAdaptive, error) {
	const (
		invalid = "invalid fractal bands hybride adaptive parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "fbanha(%d,%d,%g,%g,%g%s)"
	)

	period := p.Period
	if period < 2 {
		return nil, fmt.Errorf(fmts, invalid, "period should be greater than 1")
	}

	normalSpeedFallback := p.NormalSpeedFallback
	if normalSpeedFallback < 1 {
		return nil, fmt.Errorf(fmts, invalid, "normal_speed_fallback should be greater than 0")
	}

	alpha := p.Alpha
	if alpha <= 0.0 {
		return nil, fmt.Errorf(fmts, invalid, "alpha should be greater than 0")
	}

	nyquist := p.Nyquist
	if nyquist <= 0.0 {
		return nil, fmt.Errorf(fmts, invalid, "nyquist should be greater than 0")
	}

	alphaHP := p.AlphaHP
	if alphaHP <= 0.0 || alphaHP >= 1.0 {
		return nil, fmt.Errorf(fmts, invalid, "alpha_hp should be between 0 and 1")
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

	mnemonic := fmt.Sprintf(fmtn, period, normalSpeedFallback, alpha, nyquist, alphaHP,
		core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Fractal bands hybride adaptive " + mnemonic

	windowSize := period + 1

	f := &FractalBandsHybrideAdaptive{
		barFunc:            barFunc,
		quoteFunc:          quoteFunc,
		tradeFunc:          tradeFunc,
		window:             make([]float64, windowSize),
		closes:             make([]float64, 0, 256),
		period:             period,
		windowSize:         windowSize,
		normalSpeedFallback: normalSpeedFallback,
		alpha:              alpha,
		nyquist:            nyquist,
		alphaHP:            alphaHP,
		logDenom:           math.Log(2.0 * float64(period-1)),
		ln2:               math.Log(2.0),
		invPeriodSq:        1.0 / float64(period*period),
		smoothBuf:          make([]float64, 0, 256),
		cycleBuf:           make([]float64, 0, 256),
		q1Buf:              make([]float64, 0, 256),
		i1Buf:              make([]float64, 0, 256),
		dpBuf:              make([]float64, 0, 256),
		instPeriodBuf:      make([]float64, 0, 256),
		frasma2:           math.NaN(),
		upperBand:         math.NaN(),
		lowerBand:         math.NaN(),
	}

	f.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, f.Update)

	return f, nil
}

// IsPrimed indicates whether the indicator is primed.
func (f *FractalBandsHybrideAdaptive) IsPrimed() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.primed
}

// Metadata describes the output data of the indicator.
func (f *FractalBandsHybrideAdaptive) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.FractalBandsHybrideAdaptive,
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

// getCyclePeriod computes the Ehlers CyclePeriod for the current bar.
func (f *FractalBandsHybrideAdaptive) getCyclePeriod() float64 {
	t := len(f.closes) - 1

	// Extend buffers to index t.
	for len(f.smoothBuf) <= t {
		f.smoothBuf = append(f.smoothBuf, 0.0)
	}

	for len(f.cycleBuf) <= t {
		f.cycleBuf = append(f.cycleBuf, 0.0)
	}

	for len(f.q1Buf) <= t {
		f.q1Buf = append(f.q1Buf, 0.0)
	}

	for len(f.i1Buf) <= t {
		f.i1Buf = append(f.i1Buf, 0.0)
	}

	for len(f.dpBuf) <= t {
		f.dpBuf = append(f.dpBuf, 0.0)
	}

	for len(f.instPeriodBuf) <= t {
		f.instPeriodBuf = append(f.instPeriodBuf, 6.0)
	}

	if t < 6 {
		return math.NaN()
	}

	prices := f.closes

	// 4-bar weighted smoother.
	f.smoothBuf[t] = (prices[t] + 2.0*prices[t-1] + 2.0*prices[t-2] + prices[t-3]) / 6.0

	// High-pass filter.
	alphaHP := f.alphaHP
	hpCoeff := (1.0 - 0.5*alphaHP) * (1.0 - 0.5*alphaHP)
	oneMinusAlpha := 1.0 - alphaHP

	f.cycleBuf[t] = hpCoeff*(f.smoothBuf[t]-2.0*f.smoothBuf[t-1]+f.smoothBuf[t-2]) +
		2.0*oneMinusAlpha*f.cycleBuf[t-1] - oneMinusAlpha*oneMinusAlpha*f.cycleBuf[t-2]

	// Quadrature component.
	f.q1Buf[t] = (0.0962*f.cycleBuf[t] + 0.5769*f.cycleBuf[t-2] -
		0.5769*f.cycleBuf[t-4] - 0.0962*f.cycleBuf[t-6]) *
		(0.5 + 0.08*f.instPeriodBuf[t-1])

	// In-phase component.
	f.i1Buf[t] = f.cycleBuf[t-3]

	// Smooth I and Q with EMA.
	if t > 6 {
		f.i1Buf[t] = 0.15*f.i1Buf[t] + 0.85*f.i1Buf[t-1]
		f.q1Buf[t] = 0.15*f.q1Buf[t] + 0.85*f.q1Buf[t-1]
	}

	// Compute delta phase.
	var dp float64

	if math.Abs(f.i1Buf[t]) > 1e-10 {
		dp = math.Atan(f.q1Buf[t] / f.i1Buf[t])
	} else {
		dp = f.dpBuf[t-1]
	}

	// Clamp delta phase.
	if dp < 0.1 {
		dp = 0.1
	}

	if dp > 1.1 {
		dp = 1.1
	}

	f.dpBuf[t] = dp

	// Median delta phase over 5 bars.
	var medianDP float64

	if t >= 10 {
		w := [5]float64{f.dpBuf[t-4], f.dpBuf[t-3], f.dpBuf[t-2], f.dpBuf[t-1], f.dpBuf[t]}
		// Sort 5 elements to find median.
		for i := 0; i < 4; i++ {
			for j := i + 1; j < 5; j++ {
				if w[j] < w[i] {
					w[i], w[j] = w[j], w[i]
				}
			}
		}

		medianDP = w[2]
	} else {
		medianDP = dp
	}

	// Instantaneous period.
	var dc float64

	if math.Abs(medianDP) > 1e-10 {
		dc = 6.2832/medianDP + 0.5
	} else {
		dc = f.instPeriodBuf[t-1]
	}

	// Clamp and smooth.
	if dc < 6.0 {
		dc = 6.0
	}

	if dc > 50.0 {
		dc = 50.0
	}

	f.instPeriodBuf[t] = 0.33*dc + 0.67*f.instPeriodBuf[t-1]

	return f.instPeriodBuf[t]
}

// Update updates the value of the indicator given the next sample.
// Returns the FRASMA2 value. Use UpdateAll() for all three outputs.
func (f *FractalBandsHybrideAdaptive) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	period := f.period
	windowSize := f.windowSize

	// Accumulate close history.
	f.closes = append(f.closes, sample)

	// Update Ehlers CyclePeriod.
	cp := f.getCyclePeriod()

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

	// FGDI computation over period+1 points (period segments).
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
		length := 0.0

		for i := 1; i < windowSize; i++ {
			normCur := (f.window[i] - priceMin) / priceRange
			normPrev := (f.window[i-1] - priceMin) / priceRange
			diff := normCur - normPrev
			length += math.Sqrt(diff*diff + f.invPeriodSq)
		}

		fgdi = 1.0 + (math.Log(length)+f.ln2)/f.logDenom
	}

	// Hurst exponent.
	hurst := 2.0 - fgdi
	if hurst < 0.01 {
		hurst = 0.01
	}

	trailDim := 1.0 / hurst
	beta := trailDim / 2.0

	// Adaptive normal_speed from CyclePeriod.
	var ns float64

	if math.IsNaN(cp) || cp < 1.0 {
		ns = float64(f.normalSpeedFallback)
	} else {
		ns = cp * f.nyquist
	}

	speed := int(math.Round(ns * beta))
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

	// Deviation over the last period closes.
	sqSum := 0.0
	devStart := nCloses - period
	if devStart < 0 {
		devStart = 0
	}

	for k := devStart; k < nCloses; k++ {
		res := f.closes[k] - frasma2Val
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
func (f *FractalBandsHybrideAdaptive) UpdateAll(sample float64) (float64, float64, float64) {
	frasma2 := f.Update(sample)

	f.mu.RLock()
	defer f.mu.RUnlock()

	return frasma2, f.upperBand, f.lowerBand
}

// Frasma2Value returns the last computed FRASMA2 value.
func (f *FractalBandsHybrideAdaptive) Frasma2Value() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.frasma2
}

// UpperBandValue returns the last computed upper band value.
func (f *FractalBandsHybrideAdaptive) UpperBandValue() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.upperBand
}

// LowerBandValue returns the last computed lower band value.
func (f *FractalBandsHybrideAdaptive) LowerBandValue() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.lowerBand
}

// UpdateScalar updates the indicator given the next scalar sample.
func (f *FractalBandsHybrideAdaptive) UpdateScalar(sample *entities.Scalar) core.Output {
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
func (f *FractalBandsHybrideAdaptive) UpdateBar(sample *entities.Bar) core.Output {
	v := f.barFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (f *FractalBandsHybrideAdaptive) UpdateQuote(sample *entities.Quote) core.Output {
	v := f.quoteFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (f *FractalBandsHybrideAdaptive) UpdateTrade(sample *entities.Trade) core.Output {
	v := f.tradeFunc(sample)

	return f.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
