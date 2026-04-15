package fractaladaptivemovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// FractalAdaptiveMovingAverage (Ehler's fractal adaptive moving average, FRAMA)
// is an EMA with the smoothing factor, α, being changed with each new sample:
//
// FRAMAᵢ = αᵢPᵢ + (1 - αᵢ)*FRAMAᵢ₋₁,  αs ≤ αᵢ ≤ 1
//
// Here the αs is the slowest α (default suggested value is 0.01 or equivalent
// length of 199 samples).
//
// The concept of FRAMA is to relate the fractal dimension FDᵢ, calculated on a window
// samples, to the EMA smoothing factor αᵢ, thus making the EMA adaptive.
//
// This dependency is defined as follows:
//
//	αᵢ = exp(-w(FDᵢ - 1)),  1 ≤ FDᵢ ≤ 2,
//
//	w = ln(αs)
//
//	or, given the length ℓs = 2/αs - 1,
//
//	w = ln(2/(ℓs + 1))
//
// The fractal dimension varies over the range from 1 to 2.
//
// When FDᵢ = 1 (series forms a straight line), the exponent is zero – which means that
// αᵢ = 1, and the output of the exponential moving average is equal to the input.
//
// When FDᵢ = 2 (series fills all plane, excibiting extreme volatility), the exponent
// is -w, which means that αᵢ = αs, and the output of the exponential moving average
// is equal to the output of the slowest moving average with αs.
//
// The fractal dimension is estimated by using a "box count" method.
// Since price samples are typically uniformly spaced, the box count is approximated
// by the average slope of the price curve. This is calculated as the highest price
// minus the lowest price within an interval, divided by the length of that interval.
//
// FDᵢ = (ln(N1+N2) − ln(N3)) / ln(2)
//
// N1 is calculated over the first half of the total lookback period ℓ as the
// (highest price - lowest price) during the first ℓ/2 bars, divided by ℓ/2.
//
// N2 is calculated over the second half of the total lookback period ℓ as the
// (lighest price - lowest price) during the second ℓ/2 bars (from ℓ/2 to ℓ-1 bars ago),
// divided by ℓ/2.
//
// N3 is calculated over the entire lookback period ℓ as the
// (highest price - lowest price) during the full ℓ bars, divided by ℓ.
//
// Reference:
//
//	Falconer, K. (2014). Fractal geometry: Mathematical foundations and applications (3rd ed) Wiley.
//	Ehlers, John F. (2005). Fractal Adaptive Moving Average. Technical Analysis of Stocks & Commodities, 23(10), 81–82.
//	Ehlers, John F. (2006). FRAMA – Fractal Adaptive Moving Average, https://www.mesasoftware.com/papers/FRAMA.pdf.
type FractalAdaptiveMovingAverage struct {
	mu               sync.RWMutex
	mnemonic         string
	description      string
	mnemonicFdim     string
	descriptionFdim  string
	alphaSlowest     float64
	scalingFactor    float64
	fractalDimension float64
	value            float64
	length           int
	lengthMinOne     int
	halfLength       int
	windowCount      int
	windowHigh       []float64
	windowLow        []float64
	primed           bool
	barFunc          entities.BarFunc
	quoteFunc        entities.QuoteFunc
	tradeFunc        entities.TradeFunc
}

// NewFractalAdaptiveMovingAverage returns an instance of the indicator
// created using supplied parameters.
func NewFractalAdaptiveMovingAverage( //nolint:funlen
	params *Params,
) (*FractalAdaptiveMovingAverage, error) {
	const (
		invalid = "invalid fractal adaptive moving average parameters"
		fmtl    = "%s: length should be an even integer larger than 1"
		fmta    = "%s: slowest smoothing factor should be in range [0, 1]"
		fmtw    = "%s: %w"
		fmtn    = "frama(%d, %.3f%s)"
		fmtnd   = "framaDim(%d, %.3f%s)"
		two     = 2
		descr   = "Fractal adaptive moving average "
	)

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if params.Length < two {
		return nil, fmt.Errorf(fmtl, invalid)
	}

	if params.SlowestSmoothingFactor < 0. || params.SlowestSmoothingFactor > 1. {
		return nil, fmt.Errorf(fmta, invalid)
	}

	length := params.Length
	if length%2 != 0 {
		length++
	}

	// Resolve defaults for component functions.
	// A zero value means "use default, don't show in mnemonic".
	bc := params.BarComponent
	qc := params.QuoteComponent
	tc := params.TradeComponent

	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	componentMnemonic := core.ComponentTripleMnemonic(bc, qc, tc)
	mnemonic := fmt.Sprintf(fmtn, length, params.SlowestSmoothingFactor, componentMnemonic)
	mnemonicFdim := fmt.Sprintf(fmtnd, length, params.SlowestSmoothingFactor, componentMnemonic)

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	return &FractalAdaptiveMovingAverage{
		mnemonic:         mnemonic,
		description:      descr + mnemonic,
		mnemonicFdim:     mnemonicFdim,
		descriptionFdim:  descr + mnemonicFdim,
		length:           length,
		lengthMinOne:     length - 1,
		halfLength:       length / two,
		windowHigh:       make([]float64, length),
		windowLow:        make([]float64, length),
		windowCount:      0,
		alphaSlowest:     params.SlowestSmoothingFactor,
		scalingFactor:    math.Log(params.SlowestSmoothingFactor),
		fractalDimension: math.NaN(),
		value:            math.NaN(),
		primed:           false,
		barFunc:          barFunc,
		quoteFunc:        quoteFunc,
		tradeFunc:        tradeFunc,
	}, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *FractalAdaptiveMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *FractalAdaptiveMovingAverage) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.FractalAdaptiveMovingAverage,
		Mnemonic:    s.mnemonic,
		Description: s.description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(Value),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonic,
				Description: s.description,
			},
			{
				Kind:        int(Fdim),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonicFdim,
				Description: s.descriptionFdim,
			},
		},
	}
}

// Update updates the value of the moving average given the next sample.
func (s *FractalAdaptiveMovingAverage) Update(sample, sampleHigh, sampleLow float64) float64 {
	if math.IsNaN(sampleHigh) || math.IsNaN(sampleLow) || math.IsNaN(sample) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		for i := range s.lengthMinOne {
			j := i + 1
			s.windowHigh[i] = s.windowHigh[j]
			s.windowLow[i] = s.windowLow[j]
		}

		s.windowHigh[s.lengthMinOne] = sampleHigh
		s.windowLow[s.lengthMinOne] = sampleLow

		s.fractalDimension = s.estimateFractalDimension()
		alpha := s.estimateAlpha()
		s.value += (sample - s.value) * alpha

		return s.value
	} else {
		s.windowHigh[s.windowCount] = sampleHigh
		s.windowLow[s.windowCount] = sampleLow

		s.windowCount++
		if s.windowCount == s.lengthMinOne {
			s.value = sample
		} else if s.windowCount == s.length {
			s.fractalDimension = s.estimateFractalDimension()
			alpha := s.estimateAlpha()
			s.value += (sample - s.value) * alpha
			s.primed = true

			return s.value
		}
	}

	return math.NaN()
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *FractalAdaptiveMovingAverage) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	return s.updateEntity(sample.Time, v, v, v)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *FractalAdaptiveMovingAverage) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.updateEntity(sample.Time, v, sample.High, sample.Low)
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *FractalAdaptiveMovingAverage) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.updateEntity(sample.Time, v, sample.Ask, sample.Bid)
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *FractalAdaptiveMovingAverage) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.updateEntity(sample.Time, v, v, v)
}

func (s *FractalAdaptiveMovingAverage) updateEntity(
	time time.Time, sample, sampleHigh, sampleLow float64,
) core.Output {
	const length = 2

	output := make([]any, length)
	frama := s.Update(sample, sampleHigh, sampleLow)

	fdim := s.fractalDimension
	if math.IsNaN(frama) {
		fdim = math.NaN()
	}

	i := 0
	output[i] = entities.Scalar{Time: time, Value: frama}
	i++
	output[i] = entities.Scalar{Time: time, Value: fdim}

	return output
}

func (s *FractalAdaptiveMovingAverage) estimateFractalDimension() float64 {
	minLowHalf := math.MaxFloat64
	maxHighHalf := math.SmallestNonzeroFloat64

	for i := range s.halfLength {
		l := s.windowLow[i]
		if minLowHalf > l {
			minLowHalf = l
		}

		h := s.windowHigh[i]
		if maxHighHalf < h {
			maxHighHalf = h
		}
	}

	rangeN1 := maxHighHalf - minLowHalf
	minLowFull := minLowHalf
	maxHighFull := maxHighHalf
	minLowHalf = math.MaxFloat64
	maxHighHalf = math.SmallestNonzeroFloat64

	for j := range s.halfLength {
		i := j + s.halfLength
		l := s.windowLow[i]

		if minLowFull > l {
			minLowFull = l
		}

		if minLowHalf > l {
			minLowHalf = l
		}

		h := s.windowHigh[i]
		if maxHighFull < h {
			maxHighFull = h
		}

		if maxHighHalf < h {
			maxHighHalf = h
		}
	}

	rangeN2 := maxHighHalf - minLowHalf
	rangeN3 := maxHighFull - minLowFull

	fdim := (math.Log((rangeN1+rangeN2)/float64(s.halfLength)) -
		math.Log(rangeN3/float64(s.length))) * math.Log2E

	const two = 2

	return math.Min(math.Max(fdim, 1), two)
}

func (s *FractalAdaptiveMovingAverage) estimateAlpha() float64 {
	factor := s.scalingFactor

	// We use the fractal dimension to dynamically change the alpha of an exponential moving average.
	// The fractal dimension varies over the range from 1 to 2.
	// Since the prices are log-normal, it seems reasonable to use an exponential function to relate
	// the fractal dimension to alpha.

	// An empirically chosen scaling in Ehlers's method to map fractal dimension (1–2)
	// to the exponential α.
	alpha := math.Exp(factor * (s.fractalDimension - 1))

	// When the fractal dimension is 1, the exponent is zero – which means that alpha is 1, and
	// the output of the exponential moving average is equal to the input.

	// Limit alpha to vary only from αs to 1.
	return math.Min(math.Max(alpha, s.alphaSlowest), 1)
}
