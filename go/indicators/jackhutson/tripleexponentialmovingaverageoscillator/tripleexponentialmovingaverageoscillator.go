package tripleexponentialmovingaverageoscillator

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/common/exponentialmovingaverage"
	"zpano/indicators/core"
)

// TripleExponentialMovingAverageOscillator is Jack Hutson's Triple Exponential Moving Average
// Oscillator (TRIX).
//
// TRIX is a 1-day rate-of-change of a triple-smoothed exponential moving average. It applies
// EMA three times in series (all with the same period and SMA-seeded), then computes:
//
//	TRIX = ((EMA3[i] - EMA3[i-1]) / EMA3[i-1]) * 100
//
// The indicator oscillates around zero. Positive values indicate upward momentum, negative
// values indicate downward momentum.
//
// Reference:
//
// Hutson, Jack K. (1983). "Good TRIX". Technical Analysis of Stocks and Commodities.
type TripleExponentialMovingAverageOscillator struct {
	mu sync.RWMutex
	core.LineIndicator
	ema1           *exponentialmovingaverage.ExponentialMovingAverage
	ema2           *exponentialmovingaverage.ExponentialMovingAverage
	ema3           *exponentialmovingaverage.ExponentialMovingAverage
	previousEMA3   float64
	hasPreviousEMA bool
	primed         bool
}

// NewTripleExponentialMovingAverageOscillator returns an instance of the indicator
// created using supplied parameters.
func NewTripleExponentialMovingAverageOscillator(
	p *TripleExponentialMovingAverageOscillatorParams,
) (*TripleExponentialMovingAverageOscillator, error) {
	const (
		invalid = "invalid triple exponential moving average oscillator parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtm    = "trix(%d%s)"
		minlen  = 1
	)

	if p.Length < minlen {
		return nil, fmt.Errorf(fmts, invalid, "length should be positive")
	}

	// Resolve defaults for component functions.
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

	barFunc, err := entities.BarComponentFunc(bc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	quoteFunc, err := entities.QuoteComponentFunc(qc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	tradeFunc, err := entities.TradeComponentFunc(tc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	emaParams := &exponentialmovingaverage.ExponentialMovingAverageLengthParams{
		Length:         p.Length,
		FirstIsAverage: true,
	}

	ema1, err := exponentialmovingaverage.NewExponentialMovingAverageLength(emaParams)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	ema2, err := exponentialmovingaverage.NewExponentialMovingAverageLength(emaParams)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	ema3, err := exponentialmovingaverage.NewExponentialMovingAverageLength(emaParams)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	mnemonic := fmt.Sprintf(fmtm, p.Length, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Triple exponential moving average oscillator " + mnemonic

	s := &TripleExponentialMovingAverageOscillator{
		ema1:         ema1,
		ema2:         ema2,
		ema3:         ema3,
		previousEMA3: math.NaN(),
	}

	s.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, s.Update)

	return s, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *TripleExponentialMovingAverageOscillator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *TripleExponentialMovingAverageOscillator) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.TripleExponentialMovingAverageOscillator,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample.
func (s *TripleExponentialMovingAverageOscillator) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	v1 := s.ema1.Update(sample)
	if math.IsNaN(v1) {
		return math.NaN()
	}

	v2 := s.ema2.Update(v1)
	if math.IsNaN(v2) {
		return math.NaN()
	}

	v3 := s.ema3.Update(v2)
	if math.IsNaN(v3) {
		return math.NaN()
	}

	if !s.hasPreviousEMA {
		s.previousEMA3 = v3
		s.hasPreviousEMA = true

		return math.NaN()
	}

	const hundred = 100.0

	result := ((v3 - s.previousEMA3) / s.previousEMA3) * hundred
	s.previousEMA3 = v3

	if !s.primed {
		s.primed = true
	}

	return result
}
