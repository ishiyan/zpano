package bollingerbands

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/common/exponentialmovingaverage"
	"zpano/indicators/common/simplemovingaverage"
	"zpano/indicators/common/variance"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// lineUpdater is an interface for indicators that accept a single scalar and return a value.
type lineUpdater interface {
	Update(float64) float64
	IsPrimed() bool
}

// BollingerBands is John Bollinger's Bollinger Bands indicator.
//
// Bollinger Bands consist of a middle band (moving average) and upper/lower bands
// placed a specified number of standard deviations above and below the middle band.
//
// The indicator produces six outputs:
//   - LowerValue: middleValue - lowerMultiplier * stddev
//   - MiddleValue: moving average of the input
//   - UpperValue: middleValue + upperMultiplier * stddev
//   - BandWidth: (upperValue - lowerValue) / middleValue
//   - PercentBand: (sample - lowerValue) / (upperValue - lowerValue)
//   - Band: lower/upper band pair
//
// Reference:
//
// Bollinger, John (2002). Bollinger on Bollinger Bands. McGraw-Hill.
type BollingerBands struct {
	mu sync.RWMutex

	ma       lineUpdater
	variance *variance.Variance

	upperMultiplier float64
	lowerMultiplier float64

	middleValue float64
	upperValue  float64
	lowerValue  float64
	bandWidth   float64
	percentBand float64
	primed      bool

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	mnemonic string
}

// NewBollingerBands returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewBollingerBands(p *BollingerBandsParams) (*BollingerBands, error) {
	const (
		invalid          = "invalid bollinger bands parameters"
		fmts             = "%s: %s"
		fmtw             = "%s: %w"
		minLength        = 2
		defaultLength    = 5
		defaultMultipler = 2.0
	)

	length := p.Length
	if length == 0 {
		length = defaultLength
	}

	if length < minLength {
		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 1")
	}

	upperMultiplier := p.UpperMultiplier
	if upperMultiplier == 0 {
		upperMultiplier = defaultMultipler
	}

	lowerMultiplier := p.LowerMultiplier
	if lowerMultiplier == 0 {
		lowerMultiplier = defaultMultipler
	}

	isUnbiased := true
	if p.IsUnbiased != nil {
		isUnbiased = *p.IsUnbiased
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

	// Create variance sub-indicator.
	v, err := variance.NewVariance(&variance.VarianceParams{
		Length:         length,
		IsUnbiased:     isUnbiased,
		BarComponent:   bc,
		QuoteComponent: qc,
		TradeComponent: tc,
	})
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	// Create moving average sub-indicator.
	var ma lineUpdater

	var maLabel string

	switch p.MovingAverageType {
	case EMA:
		maLabel = "EMA"

		ema, e := exponentialmovingaverage.NewExponentialMovingAverageLength(
			&exponentialmovingaverage.ExponentialMovingAverageLengthParams{
				Length:         length,
				FirstIsAverage: p.FirstIsAverage,
			})
		if e != nil {
			return nil, fmt.Errorf(fmtw, invalid, e)
		}

		ma = ema
	default:
		maLabel = "SMA"

		sma, e := simplemovingaverage.NewSimpleMovingAverage(
			&simplemovingaverage.SimpleMovingAverageParams{Length: length})
		if e != nil {
			return nil, fmt.Errorf(fmtw, invalid, e)
		}

		ma = sma
	}

	_ = maLabel // Reserved for future mnemonic use.

	mnemonic := fmt.Sprintf("bb(%d,%.0f,%.0f%s)", length, upperMultiplier, lowerMultiplier,
		core.ComponentTripleMnemonic(bc, qc, tc))

	return &BollingerBands{
		ma:              ma,
		variance:        v,
		upperMultiplier: upperMultiplier,
		lowerMultiplier: lowerMultiplier,
		middleValue:     math.NaN(),
		upperValue:      math.NaN(),
		lowerValue:      math.NaN(),
		bandWidth:       math.NaN(),
		percentBand:     math.NaN(),
		barFunc:         barFunc,
		quoteFunc:       quoteFunc,
		tradeFunc:       tradeFunc,
		mnemonic:        mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *BollingerBands) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *BollingerBands) Metadata() core.Metadata {
	desc := "Bollinger Bands " + s.mnemonic

	return core.BuildMetadata(
		core.BollingerBands,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " lower", Description: desc + " Lower"},
			{Mnemonic: s.mnemonic + " middle", Description: desc + " Middle"},
			{Mnemonic: s.mnemonic + " upper", Description: desc + " Upper"},
			{Mnemonic: s.mnemonic + " bandWidth", Description: desc + " Band Width"},
			{Mnemonic: s.mnemonic + " percentBand", Description: desc + " Percent Band"},
			{Mnemonic: s.mnemonic + " band", Description: desc + " Band"},
		},
	)
}

// Update updates the indicator given the next sample value.
// Returns lower, middle, upper, bandWidth, percentBand values.
//
//nolint:nonamedreturns
func (s *BollingerBands) Update(sample float64) (lower, middle, upper, bw, pctB float64) {
	nan := math.NaN()

	if math.IsNaN(sample) {
		return nan, nan, nan, nan, nan
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	middle = s.ma.Update(sample)
	v := s.variance.Update(sample)

	s.primed = s.ma.IsPrimed() && s.variance.IsPrimed()

	if math.IsNaN(middle) || math.IsNaN(v) {
		s.middleValue = nan
		s.upperValue = nan
		s.lowerValue = nan
		s.bandWidth = nan
		s.percentBand = nan

		return nan, nan, nan, nan, nan
	}

	stddev := math.Sqrt(v)
	upper = middle + s.upperMultiplier*stddev
	lower = middle - s.lowerMultiplier*stddev

	const epsilon = 1e-10

	if math.Abs(middle) < epsilon {
		bw = 0
	} else {
		bw = (upper - lower) / middle
	}

	spread := upper - lower
	if math.Abs(spread) < epsilon {
		pctB = 0
	} else {
		pctB = (sample - lower) / spread
	}

	s.middleValue = middle
	s.upperValue = upper
	s.lowerValue = lower
	s.bandWidth = bw
	s.percentBand = pctB

	return lower, middle, upper, bw, pctB
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *BollingerBands) UpdateScalar(sample *entities.Scalar) core.Output {
	lower, middle, upper, bw, pctB := s.Update(sample.Value)

	const outputCount = 6

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: lower}
	output[1] = entities.Scalar{Time: sample.Time, Value: middle}
	output[2] = entities.Scalar{Time: sample.Time, Value: upper}
	output[3] = entities.Scalar{Time: sample.Time, Value: bw}
	output[4] = entities.Scalar{Time: sample.Time, Value: pctB}

	if math.IsNaN(lower) || math.IsNaN(upper) {
		output[5] = outputs.NewEmptyBand(sample.Time)
	} else {
		output[5] = outputs.NewBand(sample.Time, lower, upper)
	}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *BollingerBands) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *BollingerBands) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *BollingerBands) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
