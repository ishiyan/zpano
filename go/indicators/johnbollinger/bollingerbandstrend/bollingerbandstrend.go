package bollingerbandstrend

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/common/exponentialmovingaverage"
	"zpano/indicators/common/simplemovingaverage"
	"zpano/indicators/common/variance"
	"zpano/indicators/core"
)

// lineUpdater is an interface for indicators that accept a single scalar and return a value.
type lineUpdater interface {
	Update(float64) float64
	IsPrimed() bool
}

// bbLine holds the sub-components for one Bollinger Band calculation.
type bbLine struct {
	ma              lineUpdater
	variance        *variance.Variance
	upperMultiplier float64
	lowerMultiplier float64
}

// update feeds a sample and returns lower, middle, upper, primed.
//
//nolint:nonamedreturns
func (s *bbLine) update(sample float64) (lower, middle, upper float64, primed bool) {
	nan := math.NaN()

	middle = s.ma.Update(sample)
	v := s.variance.Update(sample)

	primed = s.ma.IsPrimed() && s.variance.IsPrimed()

	if math.IsNaN(middle) || math.IsNaN(v) {
		return nan, nan, nan, primed
	}

	stddev := math.Sqrt(v)
	upper = middle + s.upperMultiplier*stddev
	lower = middle - s.lowerMultiplier*stddev

	return lower, middle, upper, primed
}

// BollingerBandsTrend is John Bollinger's Bollinger Bands Trend indicator.
//
// BBTrend measures the difference between the widths of fast and slow Bollinger Bands
// relative to the fast middle band, indicating trend strength and direction.
//
// The indicator produces a single output:
//
//	bbtrend = (|fastLower - slowLower| - |fastUpper - slowUpper|) / fastMiddle
//
// Reference:
//
// Bollinger, John (2002). Bollinger on Bollinger Bands. McGraw-Hill.
type BollingerBandsTrend struct {
	mu sync.RWMutex

	fastBB *bbLine
	slowBB *bbLine

	value  float64
	primed bool

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	mnemonic string
}

// NewBollingerBandsTrend returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewBollingerBandsTrend(p *BollingerBandsTrendParams) (*BollingerBandsTrend, error) {
	const (
		invalid           = "invalid bollinger bands trend parameters"
		fmts              = "%s: %s"
		fmtw              = "%s: %w"
		minLength         = 2
		defaultFastLength = 20
		defaultSlowLength = 50
		defaultMultiplier = 2.0
	)

	fastLength := p.FastLength
	if fastLength == 0 {
		fastLength = defaultFastLength
	}

	slowLength := p.SlowLength
	if slowLength == 0 {
		slowLength = defaultSlowLength
	}

	if fastLength < minLength {
		return nil, fmt.Errorf(fmts, invalid, "fast length should be greater than 1")
	}

	if slowLength < minLength {
		return nil, fmt.Errorf(fmts, invalid, "slow length should be greater than 1")
	}

	if slowLength <= fastLength {
		return nil, fmt.Errorf(fmts, invalid, "slow length should be greater than fast length")
	}

	upperMultiplier := p.UpperMultiplier
	if upperMultiplier == 0 {
		upperMultiplier = defaultMultiplier
	}

	lowerMultiplier := p.LowerMultiplier
	if lowerMultiplier == 0 {
		lowerMultiplier = defaultMultiplier
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

	// Create fast and slow BB sub-components.
	fastBB, err := newBBLine(fastLength, upperMultiplier, lowerMultiplier, isUnbiased, p.MovingAverageType, p.FirstIsAverage, bc, qc, tc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	slowBB, err := newBBLine(slowLength, upperMultiplier, lowerMultiplier, isUnbiased, p.MovingAverageType, p.FirstIsAverage, bc, qc, tc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	mnemonic := fmt.Sprintf("bbtrend(%d,%d,%.0f,%.0f%s)", fastLength, slowLength, upperMultiplier, lowerMultiplier,
		core.ComponentTripleMnemonic(bc, qc, tc))

	return &BollingerBandsTrend{
		fastBB:    fastBB,
		slowBB:    slowBB,
		value:     math.NaN(),
		barFunc:   barFunc,
		quoteFunc: quoteFunc,
		tradeFunc: tradeFunc,
		mnemonic:  mnemonic,
	}, nil
}

//nolint:cyclop
func newBBLine(
	length int, upperMultiplier, lowerMultiplier float64, isUnbiased bool,
	maType MovingAverageType, firstIsAverage bool,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*bbLine, error) {
	v, err := variance.NewVariance(&variance.VarianceParams{
		Length:         length,
		IsUnbiased:     isUnbiased,
		BarComponent:   bc,
		QuoteComponent: qc,
		TradeComponent: tc,
	})
	if err != nil {
		return nil, err
	}

	var ma lineUpdater

	switch maType {
	case EMA:
		ema, e := exponentialmovingaverage.NewExponentialMovingAverageLength(
			&exponentialmovingaverage.ExponentialMovingAverageLengthParams{
				Length:         length,
				FirstIsAverage: firstIsAverage,
			})
		if e != nil {
			return nil, e
		}

		ma = ema
	default:
		sma, e := simplemovingaverage.NewSimpleMovingAverage(
			&simplemovingaverage.SimpleMovingAverageParams{Length: length})
		if e != nil {
			return nil, e
		}

		ma = sma
	}

	return &bbLine{
		ma:              ma,
		variance:        v,
		upperMultiplier: upperMultiplier,
		lowerMultiplier: lowerMultiplier,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *BollingerBandsTrend) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *BollingerBandsTrend) Metadata() core.Metadata {
	desc := "Bollinger Bands Trend " + s.mnemonic

	return core.BuildMetadata(
		core.BollingerBandsTrend,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: desc},
		},
	)
}

// Update updates the indicator given the next sample value and returns the BBTrend value.
func (s *BollingerBandsTrend) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fastLower, fastMiddle, fastUpper, fastPrimed := s.fastBB.update(sample)
	slowLower, _, slowUpper, slowPrimed := s.slowBB.update(sample)

	s.primed = fastPrimed && slowPrimed

	if !s.primed || math.IsNaN(fastMiddle) || math.IsNaN(fastLower) || math.IsNaN(slowLower) {
		s.value = math.NaN()
		return math.NaN()
	}

	const epsilon = 1e-10

	lowerDiff := math.Abs(fastLower - slowLower)
	upperDiff := math.Abs(fastUpper - slowUpper)

	if math.Abs(fastMiddle) < epsilon {
		s.value = 0
		return 0
	}

	result := (lowerDiff - upperDiff) / fastMiddle
	s.value = result

	return result
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *BollingerBandsTrend) UpdateScalar(sample *entities.Scalar) core.Output {
	v := s.Update(sample.Value)

	return []any{entities.Scalar{Time: sample.Time, Value: v}}
}

// UpdateBar updates the indicator given the next bar sample.
func (s *BollingerBandsTrend) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *BollingerBandsTrend) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *BollingerBandsTrend) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
