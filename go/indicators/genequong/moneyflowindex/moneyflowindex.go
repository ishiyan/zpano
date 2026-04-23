package moneyflowindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// MoneyFlowIndex is Gene Quong's Money Flow Index (MFI).
//
// MFI is a volume-weighted oscillator calculated over ℓ periods, showing money flow
// on up days as a percentage of the total of up and down days.
//
//	TypicalPrice = (High + Low + Close) / 3
//	MoneyFlow = TypicalPrice × Volume
//	MFI = 100 × PositiveMoneyFlow / (PositiveMoneyFlow + NegativeMoneyFlow)
//
// A value of 80 is generally considered overbought, or a value of 20 oversold.
//
// Reference:
//
// Quong, Gene, and Soudack, Avrum (1989). "Volume-Weighted RSI: Money Flow Index".
// Technical Analysis of Stocks and Commodities.
type MoneyFlowIndex struct {
	mu sync.RWMutex
	core.LineIndicator
	barFunc        entities.BarFunc
	length         int
	negativeBuffer []float64
	positiveBuffer []float64
	negativeSum    float64
	positiveSum    float64
	previousSample float64
	bufferIndex    int
	bufferLowIndex int
	bufferCount    int
	value          float64
	primed         bool
}

// NewMoneyFlowIndex returns an instance of the indicator created using supplied parameters.
func NewMoneyFlowIndex(p *MoneyFlowIndexParams) (*MoneyFlowIndex, error) {
	const (
		invalid   = "invalid money flow index parameters"
		fmts      = "%s: %s"
		fmtw      = "%s: %w"
		minLength = 1
	)

	if p.Length < minLength {
		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 0")
	}

	bc := p.BarComponent
	if bc == 0 {
		bc = entities.BarTypicalPrice
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

	mnemonic := fmt.Sprintf("mfi(%d%s)", p.Length, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Money Flow Index " + mnemonic

	m := &MoneyFlowIndex{
		barFunc:        barFunc,
		length:         p.Length,
		negativeBuffer: make([]float64, p.Length),
		positiveBuffer: make([]float64, p.Length),
		value:          math.NaN(),
	}

	// LineIndicator's Update uses volume=1 (scalar-only path).
	m.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, m.Update)

	return m, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *MoneyFlowIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *MoneyFlowIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.MoneyFlowIndex,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator with the given sample using volume = 1.
// This satisfies the LineIndicator updateFn signature.
func (s *MoneyFlowIndex) Update(sample float64) float64 {
	return s.UpdateWithVolume(sample, 1)
}

// UpdateWithVolume updates the indicator with the given sample and volume.
func (s *MoneyFlowIndex) UpdateWithVolume(sample, volume float64) float64 {
	if math.IsNaN(sample) || math.IsNaN(volume) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	lengthMinOne := s.length - 1

	if s.primed {
		s.negativeSum -= s.negativeBuffer[s.bufferLowIndex]
		s.positiveSum -= s.positiveBuffer[s.bufferLowIndex]

		amount := sample * volume
		diff := sample - s.previousSample

		if diff < 0 {
			s.negativeBuffer[s.bufferIndex] = amount
			s.positiveBuffer[s.bufferIndex] = 0
			s.negativeSum += amount
		} else if diff > 0 {
			s.negativeBuffer[s.bufferIndex] = 0
			s.positiveBuffer[s.bufferIndex] = amount
			s.positiveSum += amount
		} else {
			s.negativeBuffer[s.bufferIndex] = 0
			s.positiveBuffer[s.bufferIndex] = 0
		}

		sum := s.positiveSum + s.negativeSum
		if sum < 1 {
			s.value = 0
		} else {
			s.value = 100 * s.positiveSum / sum
		}

		s.bufferIndex++
		if s.bufferIndex > lengthMinOne {
			s.bufferIndex = 0
		}

		s.bufferLowIndex++
		if s.bufferLowIndex > lengthMinOne {
			s.bufferLowIndex = 0
		}
	} else if s.bufferCount == 0 {
		s.bufferCount++
	} else {
		amount := sample * volume
		diff := sample - s.previousSample

		if diff < 0 {
			s.negativeBuffer[s.bufferIndex] = amount
			s.positiveBuffer[s.bufferIndex] = 0
			s.negativeSum += amount
		} else if diff > 0 {
			s.negativeBuffer[s.bufferIndex] = 0
			s.positiveBuffer[s.bufferIndex] = amount
			s.positiveSum += amount
		} else {
			s.negativeBuffer[s.bufferIndex] = 0
			s.positiveBuffer[s.bufferIndex] = 0
		}

		if s.length == s.bufferCount {
			sum := s.positiveSum + s.negativeSum
			if sum < 1 {
				s.value = 0
			} else {
				s.value = 100 * s.positiveSum / sum
			}

			s.primed = true
		}

		s.bufferIndex++
		if s.bufferIndex > lengthMinOne {
			s.bufferIndex = 0
		}

		s.bufferCount++
	}

	s.previousSample = sample

	return s.value
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to use bar volume.
func (s *MoneyFlowIndex) UpdateBar(sample *entities.Bar) core.Output {
	price := s.barFunc(sample)
	value := s.UpdateWithVolume(price, sample.Volume)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
