package stochasticrelativestrengthindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/common/exponentialmovingaverage"
	"zpano/indicators/common/simplemovingaverage"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/welleswilder/relativestrengthindex"
)

// lineUpdater is an interface for indicators that accept a single scalar and return a value.
type lineUpdater interface {
	Update(float64) float64
	IsPrimed() bool
}

// passthrough is a no-op smoother for FastD period of 1.
type passthrough struct{}

func (p *passthrough) Update(v float64) float64 { return v }
func (p *passthrough) IsPrimed() bool           { return true }

// StochasticRelativeStrengthIndex is Tushar Chande's Stochastic RSI.
//
// Stochastic RSI applies the Stochastic oscillator formula to RSI values
// instead of price data. It oscillates between 0 and 100.
//
// The indicator first computes RSI, then applies a stochastic calculation
// over a rolling window of RSI values to produce Fast-K. Fast-D is a
// moving average of Fast-K.
//
// Reference:
//
// Chande, Tushar S. and Kroll, Stanley (1993). "Stochastic RSI and Dynamic
// Momentum Index". Stock & Commodities V.11:5 (189-199).
type StochasticRelativeStrengthIndex struct {
	mu sync.RWMutex

	rsi *relativestrengthindex.RelativeStrengthIndex

	// Circular buffer for RSI values (size = fastKLength).
	rsiBuf    []float64
	rsiBufIdx int
	rsiCount  int

	fastKLength int
	fastDMA     lineUpdater

	fastK  float64
	fastD  float64
	primed bool

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc
	mnemonic  string
}

// NewStochasticRelativeStrengthIndex returns an instance of the indicator created using supplied parameters.
func NewStochasticRelativeStrengthIndex(p *StochasticRelativeStrengthIndexParams) (*StochasticRelativeStrengthIndex, error) {
	const (
		invalid       = "invalid stochastic relative strength index parameters"
		fmts          = "%s: %s"
		fmtw          = "%s: %w"
		minRSILength  = 2
		minFastLength = 1
	)

	if p.Length < minRSILength {
		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 1")
	}

	if p.FastKLength < minFastLength {
		return nil, fmt.Errorf(fmts, invalid, "fast K length should be greater than 0")
	}

	if p.FastDLength < minFastLength {
		return nil, fmt.Errorf(fmts, invalid, "fast D length should be greater than 0")
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

	// Create internal RSI.
	rsi, rsiErr := relativestrengthindex.NewRelativeStrengthIndex(&relativestrengthindex.RelativeStrengthIndexParams{
		Length: p.Length,
	})
	if rsiErr != nil {
		return nil, fmt.Errorf(fmtw, invalid, rsiErr)
	}

	// Create Fast-D smoother.
	var fastDMA lineUpdater

	var maLabel string

	if p.FastDLength < 2 { //nolint:mnd
		fastDMA = &passthrough{}
		maLabel = "SMA"
	} else {
		switch p.MovingAverageType {
		case EMA:
			maLabel = "EMA"

			ema, e := exponentialmovingaverage.NewExponentialMovingAverageLength(
				&exponentialmovingaverage.ExponentialMovingAverageLengthParams{
					Length:         p.FastDLength,
					FirstIsAverage: p.FirstIsAverage,
				})
			if e != nil {
				return nil, fmt.Errorf(fmtw, invalid, e)
			}

			fastDMA = ema
		default:
			maLabel = "SMA"

			sma, e := simplemovingaverage.NewSimpleMovingAverage(
				&simplemovingaverage.SimpleMovingAverageParams{Length: p.FastDLength})
			if e != nil {
				return nil, fmt.Errorf(fmtw, invalid, e)
			}

			fastDMA = sma
		}
	}

	mnemonic := fmt.Sprintf("stochrsi(%d/%d/%s%d%s)", p.Length, p.FastKLength,
		maLabel, p.FastDLength, core.ComponentTripleMnemonic(bc, qc, tc))

	return &StochasticRelativeStrengthIndex{
		rsi:         rsi,
		rsiBuf:      make([]float64, p.FastKLength),
		fastKLength: p.FastKLength,
		fastDMA:     fastDMA,
		fastK:       math.NaN(),
		fastD:       math.NaN(),
		barFunc:     barFunc,
		quoteFunc:   quoteFunc,
		tradeFunc:   tradeFunc,
		mnemonic:    mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *StochasticRelativeStrengthIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *StochasticRelativeStrengthIndex) Metadata() core.Metadata {
	desc := "Stochastic Relative Strength Index " + s.mnemonic

	return core.Metadata{
		Type:        core.StochasticRelativeStrengthIndex,
		Mnemonic:    s.mnemonic,
		Description: desc,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(StochasticRelativeStrengthIndexFastK),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonic + " fastK",
				Description: desc + " Fast-K",
			},
			{
				Kind:        int(StochasticRelativeStrengthIndexFastD),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonic + " fastD",
				Description: desc + " Fast-D",
			},
		},
	}
}

// Update updates the indicator given the next sample and returns both FastK and FastD values.
func (s *StochasticRelativeStrengthIndex) Update(sample float64) (float64, float64) {
	if math.IsNaN(sample) {
		return math.NaN(), math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Feed to internal RSI.
	rsiVal := s.rsi.Update(sample)
	if math.IsNaN(rsiVal) {
		return s.fastK, s.fastD
	}

	// Store RSI value in circular buffer.
	s.rsiBuf[s.rsiBufIdx] = rsiVal
	s.rsiBufIdx = (s.rsiBufIdx + 1) % s.fastKLength
	s.rsiCount++

	// Need at least fastKLength RSI values for stochastic calculation.
	if s.rsiCount < s.fastKLength {
		return s.fastK, s.fastD
	}

	// Find min and max of RSI values in the window.
	minRSI := s.rsiBuf[0]
	maxRSI := s.rsiBuf[0]

	for i := 1; i < s.fastKLength; i++ {
		if s.rsiBuf[i] < minRSI {
			minRSI = s.rsiBuf[i]
		}

		if s.rsiBuf[i] > maxRSI {
			maxRSI = s.rsiBuf[i]
		}
	}

	// Calculate Fast-K.
	diff := maxRSI - minRSI
	if diff > 0 {
		s.fastK = 100 * (rsiVal - minRSI) / diff //nolint:mnd
	} else {
		s.fastK = 0
	}

	// Feed Fast-K to Fast-D smoother.
	s.fastD = s.fastDMA.Update(s.fastK)

	if !s.primed && s.fastDMA.IsPrimed() {
		s.primed = true
	}

	return s.fastK, s.fastD
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *StochasticRelativeStrengthIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	fastK, fastD := s.Update(sample.Value)

	output := make([]any, 2) //nolint:mnd
	output[0] = entities.Scalar{Time: sample.Time, Value: fastK}
	output[1] = entities.Scalar{Time: sample.Time, Value: fastD}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *StochasticRelativeStrengthIndex) UpdateBar(sample *entities.Bar) core.Output {
	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: s.barFunc(sample)})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *StochasticRelativeStrengthIndex) UpdateQuote(sample *entities.Quote) core.Output {
	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: s.quoteFunc(sample)})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *StochasticRelativeStrengthIndex) UpdateTrade(sample *entities.Trade) core.Output {
	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: s.tradeFunc(sample)})
}
