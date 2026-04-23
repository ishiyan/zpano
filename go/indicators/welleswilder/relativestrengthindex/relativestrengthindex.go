package relativestrengthindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// RelativeStrengthIndex is Welles Wilder's Relative Strength Index (RSI).
//
// RSI measures the magnitude of recent price changes to evaluate overbought
// or oversold conditions. It oscillates between 0 and 100.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type RelativeStrengthIndex struct {
	mu sync.RWMutex
	core.LineIndicator
	length         int
	count          int
	previousSample float64
	previousGain   float64
	previousLoss   float64
	value          float64
	primed         bool
}

// NewRelativeStrengthIndex returns an instance of the indicator created using supplied parameters.
func NewRelativeStrengthIndex(p *RelativeStrengthIndexParams) (*RelativeStrengthIndex, error) {
	const (
		invalid   = "invalid relative strength index parameters"
		fmts      = "%s: %s"
		fmtw      = "%s: %w"
		minLength = 2
	)

	length := p.Length
	if length < minLength {
		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 1")
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

	mnemonic := fmt.Sprintf("rsi(%d%s)", length, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Relative Strength Index " + mnemonic

	rsi := &RelativeStrengthIndex{
		length: length,
		count:  -1,
		value:  math.NaN(),
	}

	rsi.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, rsi.Update)

	return rsi, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *RelativeStrengthIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *RelativeStrengthIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.RelativeStrengthIndex,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the indicator given the next sample.
func (s *RelativeStrengthIndex) Update(sample float64) float64 {
	const epsilon = 1e-8

	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.count++

	if s.count == 0 {
		s.previousSample = sample

		return s.value
	}

	temp := sample - s.previousSample
	s.previousSample = sample

	if !s.primed {
		// Accumulation phase: count 1..length-1.
		if temp < 0 {
			s.previousLoss -= temp
		} else {
			s.previousGain += temp
		}

		if s.count < s.length {
			return s.value
		}

		// Priming: count == length.
		s.previousGain /= float64(s.length)
		s.previousLoss /= float64(s.length)
		s.primed = true
	} else {
		// Wilder's smoothing.
		s.previousGain *= float64(s.length - 1)
		s.previousLoss *= float64(s.length - 1)

		if temp < 0 {
			s.previousLoss -= temp
		} else {
			s.previousGain += temp
		}

		s.previousGain /= float64(s.length)
		s.previousLoss /= float64(s.length)
	}

	sum := s.previousGain + s.previousLoss
	if sum > epsilon {
		s.value = 100 * s.previousGain / sum
	} else {
		s.value = 0
	}

	return s.value
}
