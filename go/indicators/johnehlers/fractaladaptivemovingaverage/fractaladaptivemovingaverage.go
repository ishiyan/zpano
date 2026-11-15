package fractaladaptivemovingaverage

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
)

var logTwo = math.Log(2.0)

func boxRange(highs, lows []float64, start, end int) (float64, error) {
	highest := highs[start]
	lowest := lows[start]
	for index := start + 1; index < end; index++ {
		if highs[index] > highest {
			highest = highs[index]
		}
		if lows[index] < lowest {
			lowest = lows[index]
		}
	}
	span := highest - lowest
	if math.IsNaN(span) || math.IsInf(span, 0) {
		return 0, errors.New("price-range arithmetic overflowed")
	}
	return span, nil
}

// FractalAdaptiveMovingAverage implements John F. Ehlers' 2005 Fractal Adaptive
// Moving Average (FRAMA), returning both the adaptive EMA and the raw guarded
// box-range fractal-dimension estimate used to adapt it.
type FractalAdaptiveMovingAverage struct {
	mu                     sync.RWMutex
	mnemonic               string
	description            string
	mnemonicFdim           string
	descriptionFdim        string
	length                 int
	halfLength             int
	warmupCount            int
	slowestSmoothingFactor float64
	coefficientMode        int
	exponentScale          float64
	count                  int
	highs                  []float64
	lows                   []float64
	dimensionState         float64
	alphaState             float64
	filterState            float64
	currentValue           float64
	currentDimension       float64
	barFunc                entities.BarFunc
	quoteFunc              entities.QuoteFunc
	tradeFunc              entities.TradeFunc
}

// NewFractalAdaptiveMovingAverageDefault creates a FractalAdaptiveMovingAverage with Ehlers defaults.
func NewFractalAdaptiveMovingAverageDefault() (*FractalAdaptiveMovingAverage, error) {
	return NewFractalAdaptiveMovingAverage(DefaultParams())
}

// NewFractalAdaptiveMovingAverage returns an instance of the indicator created using supplied parameters.
func NewFractalAdaptiveMovingAverage(params *Params) (*FractalAdaptiveMovingAverage, error) {
	const (
		invalid = "invalid fractal adaptive moving average parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		descr   = "Fractal adaptive moving average "
	)

	if params == nil {
		params = DefaultParams()
	}

	if params.Length < 2 || params.Length%2 != 0 {
		return nil, fmt.Errorf(fmts, invalid, "length must be an even integer >= 2")
	}

	slowest := params.SlowestSmoothingFactor
	if math.IsNaN(slowest) || math.IsInf(slowest, 0) || slowest <= 0.0 || slowest > 1.0 {
		return nil, fmt.Errorf(fmts, invalid, "slowest_smoothing_factor must satisfy 0 < A <= 1")
	}

	if params.CoefficientMode != 0 && params.CoefficientMode != 1 {
		return nil, fmt.Errorf(fmts, invalid, "coefficient_mode must be 0 (article) or 1 (full precision)")
	}
	if params.CoefficientMode == 0 && slowest != 0.01 {
		return nil, fmt.Errorf(fmts, invalid, "literal article mode requires slowest_smoothing_factor=0.01")
	}

	barComp := params.BarComponent
	if barComp == 0 {
		barComp = entities.DefaultBarComponent
	}
	quoteComp := params.QuoteComponent
	if quoteComp == 0 {
		quoteComp = entities.DefaultQuoteComponent
	}
	tradeComp := params.TradeComponent
	if tradeComp == 0 {
		tradeComp = entities.DefaultTradeComponent
	}

	barFunc, err := entities.BarComponentFunc(barComp)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}
	quoteFunc, err := entities.QuoteComponentFunc(quoteComp)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}
	tradeFunc, err := entities.TradeComponentFunc(tradeComp)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	exponentScale := -4.6
	if params.CoefficientMode == 1 {
		exponentScale = math.Log(slowest)
	}

	mnemonic := params.Mnemonic()
	mnemonicFdim := params.DimensionMnemonic()

	indicator := &FractalAdaptiveMovingAverage{
		mnemonic:               mnemonic,
		description:            descr + mnemonic,
		mnemonicFdim:           mnemonicFdim,
		descriptionFdim:        descr + mnemonicFdim,
		length:                 params.Length,
		halfLength:             params.Length / 2,
		warmupCount:            params.Length - 1,
		slowestSmoothingFactor: slowest,
		coefficientMode:        params.CoefficientMode,
		exponentScale:          exponentScale,
		highs:                  make([]float64, 0, params.Length),
		lows:                   make([]float64, 0, params.Length),
		barFunc:                barFunc,
		quoteFunc:              quoteFunc,
		tradeFunc:              tradeFunc,
	}
	indicator.Reset()
	return indicator, nil
}

// Reset clears all streaming histories and priming state while preserving configuration.
func (s *FractalAdaptiveMovingAverage) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.count = 0
	s.highs = s.highs[:0]
	s.lows = s.lows[:0]
	s.dimensionState = 0.0
	s.alphaState = 1.0
	s.filterState = 0.0
	s.currentValue = math.NaN()
	s.currentDimension = math.NaN()
}

// WarmupBars returns the number of initial NaN outputs before priming.
func (s *FractalAdaptiveMovingAverage) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.warmupCount
}

// SamplesSeen returns the number of accepted finite observations seen since start or reset.
func (s *FractalAdaptiveMovingAverage) SamplesSeen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count
}

// Value returns the latest accepted FRAMA value (NaN before priming).
func (s *FractalAdaptiveMovingAverage) Value() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentValue
}

// FractalDimension returns the latest accepted raw fractal-dimension estimate (NaN before priming).
func (s *FractalAdaptiveMovingAverage) FractalDimension() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentDimension
}

// Alpha returns the latest diagnostic smoothing factor (NaN before priming).
func (s *FractalAdaptiveMovingAverage) Alpha() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.count < s.length {
		return math.NaN()
	}
	return s.alphaState
}

// IsPrimed indicates whether an indicator is primed.
func (s *FractalAdaptiveMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count >= s.length
}

// Metadata describes a requested output data of an indicator.
func (s *FractalAdaptiveMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.FractalAdaptiveMovingAverage,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicFdim, Description: s.descriptionFdim},
		},
	)
}

func (s *FractalAdaptiveMovingAverage) tryUpdateLocked(
	sample, sampleHigh, sampleLow float64,
) (float64, error) {
	if math.IsInf(sample, 0) {
		return math.NaN(), errors.New("price must be finite or NaN")
	}
	if math.IsInf(sampleHigh, 0) {
		return math.NaN(), errors.New("high must be finite or NaN")
	}
	if math.IsInf(sampleLow, 0) {
		return math.NaN(), errors.New("low must be finite or NaN")
	}
	if math.IsNaN(sample) || math.IsNaN(sampleHigh) || math.IsNaN(sampleLow) {
		return math.NaN(), nil
	}
	if sampleHigh < sampleLow {
		return math.NaN(), errors.New("high must be >= low")
	}

	keep := len(s.highs)
	if keep > s.length-1 {
		keep = s.length - 1
	}
	nextHighs := make([]float64, keep+1)
	nextLows := make([]float64, keep+1)
	nextHighs[0] = sampleHigh
	nextLows[0] = sampleLow
	copy(nextHighs[1:], s.highs[:keep])
	copy(nextLows[1:], s.lows[:keep])

	nextCount := s.count + 1
	dimension := s.dimensionState
	alpha := s.alphaState

	if nextCount >= s.length {
		range1, err := boxRange(nextHighs, nextLows, 0, s.halfLength)
		if err != nil {
			return math.NaN(), err
		}
		range2, err := boxRange(nextHighs, nextLows, s.halfLength, s.length)
		if err != nil {
			return math.NaN(), err
		}
		range3, err := boxRange(nextHighs, nextLows, 0, s.length)
		if err != nil {
			return math.NaN(), err
		}

		halfFloat := float64(s.halfLength)
		lengthFloat := float64(s.length)
		n1 := range1 / halfFloat
		n2 := range2 / halfFloat
		n3 := range3 / lengthFloat

		if n1 > 0.0 && n2 > 0.0 && n3 > 0.0 {
			combined := n1 + n2
			if math.IsNaN(combined) || math.IsInf(combined, 0) {
				return math.NaN(), errors.New("combined box-range arithmetic overflowed")
			}
			dimension = (math.Log(combined) - math.Log(n3)) / logTwo
			if math.IsNaN(dimension) || math.IsInf(dimension, 0) {
				return math.NaN(), errors.New("dimension arithmetic overflowed")
			}
		}

		if dimension <= 1.0 {
			alpha = 1.0
		} else {
			alpha = math.Exp(s.exponentScale * (dimension - 1.0))
			if alpha < s.slowestSmoothingFactor {
				alpha = s.slowestSmoothingFactor
			}
			if alpha > 1.0 {
				alpha = 1.0
			}
		}
	}

	var nextValue float64
	if nextCount <= s.length {
		nextValue = sample
	} else {
		nextValue = alpha*sample + (1.0-alpha)*s.filterState
		if math.IsNaN(nextValue) || math.IsInf(nextValue, 0) {
			return math.NaN(), errors.New("FRAMA arithmetic overflowed")
		}
	}

	s.highs = nextHighs
	s.lows = nextLows
	s.count = nextCount
	s.dimensionState = dimension
	s.alphaState = alpha
	s.filterState = nextValue
	if nextCount >= s.length {
		s.currentValue = nextValue
		s.currentDimension = dimension
		return nextValue, nil
	}
	return math.NaN(), nil
}

// TryUpdate consumes one (sample, sampleHigh, sampleLow) observation and returns an error on invalid input.
func (s *FractalAdaptiveMovingAverage) TryUpdate(
	sample, sampleHigh, sampleLow float64,
) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.tryUpdateLocked(sample, sampleHigh, sampleLow)
}

// Update updates the value of the indicator given (sample, sampleHigh, sampleLow) and panics on invalid input.
func (s *FractalAdaptiveMovingAverage) Update(sample, sampleHigh, sampleLow float64) float64 {
	result, err := s.TryUpdate(sample, sampleHigh, sampleLow)
	if err != nil {
		panic(err)
	}
	return result
}

// TryUpdatePrice consumes one scalar price observation (using price for high and low).
func (s *FractalAdaptiveMovingAverage) TryUpdatePrice(price float64) (float64, error) {
	return s.TryUpdate(price, price, price)
}

// UpdatePrice updates the indicator with a scalar price observation and panics on invalid input.
func (s *FractalAdaptiveMovingAverage) UpdatePrice(price float64) float64 {
	return s.Update(price, price, price)
}

// UpdateScalar updates an indicator given the next scalar sample.
func (s *FractalAdaptiveMovingAverage) UpdateScalar(sample *entities.Scalar) core.Output {
	extracted := sample.Value
	return s.updateEntity(sample.Time, extracted, extracted, extracted)
}

// UpdateBar updates an indicator given the next bar sample.
func (s *FractalAdaptiveMovingAverage) UpdateBar(sample *entities.Bar) core.Output {
	extracted := s.barFunc(sample)
	return s.updateEntity(sample.Time, extracted, sample.High, sample.Low)
}

// UpdateQuote updates an indicator given the next quote sample.
func (s *FractalAdaptiveMovingAverage) UpdateQuote(sample *entities.Quote) core.Output {
	extracted := s.quoteFunc(sample)
	return s.updateEntity(sample.Time, extracted, sample.Ask, sample.Bid)
}

// UpdateTrade updates an indicator given the next trade sample.
func (s *FractalAdaptiveMovingAverage) UpdateTrade(sample *entities.Trade) core.Output {
	extracted := s.tradeFunc(sample)
	return s.updateEntity(sample.Time, extracted, extracted, extracted)
}

func (s *FractalAdaptiveMovingAverage) updateEntity(
	stamp time.Time,
	sample, sampleHigh, sampleLow float64,
) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	frama, err := s.tryUpdateLocked(sample, sampleHigh, sampleLow)
	if err != nil {
		panic(err)
	}
	fdim := math.NaN()
	if !math.IsNaN(frama) {
		fdim = s.currentDimension
	}

	output := make(core.Output, 2)
	output[0] = entities.Scalar{Time: stamp, Value: frama}
	output[1] = entities.Scalar{Time: stamp, Value: fdim}
	return output
}
