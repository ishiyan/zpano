package stochastic

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/common/exponentialmovingaverage"
	"zpano/indicators/common/simplemovingaverage"
	"zpano/indicators/core"
)

// lineUpdater is an interface for indicators that accept a single scalar and return a value.
type lineUpdater interface {
	Update(float64) float64
	IsPrimed() bool
}

// passthrough is a no-op smoother for period of 1.
type passthrough struct{}

// Update returns v unchanged (no smoothing).
func (p *passthrough) Update(v float64) float64 { return v }

// IsPrimed always reports true; a passthrough has no warmup.
func (p *passthrough) IsPrimed() bool { return true }

// Stochastic is George Lane's Stochastic Oscillator.
//
// The Stochastic Oscillator measures the position of the close relative to the
// high-low range over a lookback period. It produces three outputs:
//   - Fast-K: the raw stochastic value
//   - Slow-K: a moving average of Fast-K (also known as Fast-D)
//   - Slow-D: a moving average of Slow-K
//
// The indicator requires bar data (high, low, close). For scalar, quote, and
// trade updates, the single value substitutes for all three.
//
// Reference:
//
// Lane, George C. (1984). "Lane's Stochastics". Technical Analysis of Stocks & Commodities.
type Stochastic struct {
	mu sync.RWMutex

	fastKLength int

	// Circular buffers for high and low values (size = fastKLength).
	highBuf     []float64
	lowBuf      []float64
	bufferIndex int
	count       int

	slowKMA lineUpdater
	slowDMA lineUpdater

	fastK  float64
	slowK  float64
	slowD  float64
	primed bool

	mnemonic string
}

// NewStochastic returns an instance of the indicator created using supplied parameters.
func NewStochastic(p *StochasticParams) (*Stochastic, error) {
	const (
		invalid   = "invalid stochastic parameters"
		fmts      = "%s: %s"
		fmtw      = "%s: %w"
		minLength = 1
	)

	if p.FastKLength < minLength {
		return nil, fmt.Errorf(fmts, invalid, "fast K length should be greater than 0")
	}

	if p.SlowKLength < minLength {
		return nil, fmt.Errorf(fmts, invalid, "slow K length should be greater than 0")
	}

	if p.SlowDLength < minLength {
		return nil, fmt.Errorf(fmts, invalid, "slow D length should be greater than 0")
	}

	// Create Slow-K smoother.
	slowKMA, slowKLabel, err := createMA(p.SlowKMAType, p.SlowKLength, p.FirstIsAverage)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	// Create Slow-D smoother.
	slowDMA, slowDLabel, err := createMA(p.SlowDMAType, p.SlowDLength, p.FirstIsAverage)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	mnemonic := fmt.Sprintf("stoch(%d/%s%d/%s%d)", p.FastKLength,
		slowKLabel, p.SlowKLength, slowDLabel, p.SlowDLength)

	return &Stochastic{
		fastKLength: p.FastKLength,
		highBuf:     make([]float64, p.FastKLength),
		lowBuf:      make([]float64, p.FastKLength),
		slowKMA:     slowKMA,
		slowDMA:     slowDMA,
		fastK:       math.NaN(),
		slowK:       math.NaN(),
		slowD:       math.NaN(),
		mnemonic:    mnemonic,
	}, nil
}

func createMA(maType MovingAverageType, length int, firstIsAverage bool) (lineUpdater, string, error) {
	if length < 2 { //nolint:mnd
		return &passthrough{}, "SMA", nil
	}

	switch maType {
	case EMA:
		ema, e := exponentialmovingaverage.NewExponentialMovingAverageLength(
			&exponentialmovingaverage.ExponentialMovingAverageLengthParams{
				Length:         length,
				FirstIsAverage: firstIsAverage,
			})
		if e != nil {
			return nil, "", e
		}

		return ema, "EMA", nil
	default:
		sma, e := simplemovingaverage.NewSimpleMovingAverage(
			&simplemovingaverage.SimpleMovingAverageParams{Length: length})
		if e != nil {
			return nil, "", e
		}

		return sma, "SMA", nil
	}
}

// IsPrimed indicates whether the indicator is primed.
func (s *Stochastic) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *Stochastic) Metadata() core.Metadata {
	desc := "Stochastic Oscillator " + s.mnemonic

	return core.BuildMetadata(
		core.Stochastic,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " fastK", Description: desc + " Fast-K"},
			{Mnemonic: s.mnemonic + " slowK", Description: desc + " Slow-K"},
			{Mnemonic: s.mnemonic + " slowD", Description: desc + " Slow-D"},
		},
	)
}

// Update updates the indicator given the next bar's close, high, and low values.
// Returns FastK, SlowK, and SlowD.
func (s *Stochastic) Update(close, high, low float64) (float64, float64, float64) {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN(), math.NaN(), math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Store high and low in circular buffer.
	s.highBuf[s.bufferIndex] = high
	s.lowBuf[s.bufferIndex] = low
	s.bufferIndex = (s.bufferIndex + 1) % s.fastKLength
	s.count++

	// Need at least fastKLength bars.
	if s.count < s.fastKLength {
		return s.fastK, s.slowK, s.slowD
	}

	// Find highest high and lowest low in the window.
	hh := s.highBuf[0]
	ll := s.lowBuf[0]

	for i := 1; i < s.fastKLength; i++ {
		if s.highBuf[i] > hh {
			hh = s.highBuf[i]
		}

		if s.lowBuf[i] < ll {
			ll = s.lowBuf[i]
		}
	}

	// Calculate Fast-K.
	diff := hh - ll
	if diff > 0 {
		s.fastK = 100 * (close - ll) / diff //nolint:mnd
	} else {
		s.fastK = 0
	}

	// Feed Fast-K to Slow-K smoother.
	s.slowK = s.slowKMA.Update(s.fastK)

	// Feed Slow-K to Slow-D smoother (only when Slow-K MA is primed).
	if s.slowKMA.IsPrimed() {
		s.slowD = s.slowDMA.Update(s.slowK)

		if !s.primed && s.slowDMA.IsPrimed() {
			s.primed = true
		}
	}

	return s.fastK, s.slowK, s.slowD
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *Stochastic) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value
	fastK, slowK, slowD := s.Update(v, v, v)

	output := make([]any, 3) //nolint:mnd
	output[0] = entities.Scalar{Time: sample.Time, Value: fastK}
	output[1] = entities.Scalar{Time: sample.Time, Value: slowK}
	output[2] = entities.Scalar{Time: sample.Time, Value: slowD}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *Stochastic) UpdateBar(sample *entities.Bar) core.Output {
	fastK, slowK, slowD := s.Update(sample.Close, sample.High, sample.Low)

	output := make([]any, 3) //nolint:mnd
	output[0] = entities.Scalar{Time: sample.Time, Value: fastK}
	output[1] = entities.Scalar{Time: sample.Time, Value: slowK}
	output[2] = entities.Scalar{Time: sample.Time, Value: slowD}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *Stochastic) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *Stochastic) UpdateTrade(sample *entities.Trade) core.Output {
	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: sample.Price})
}
