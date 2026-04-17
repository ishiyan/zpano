package moneyflowindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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

	mnemonic := fmt.Sprintf("mfi(%d)", p.Length)
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
func (m *MoneyFlowIndex) IsPrimed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.primed
}

// Metadata describes the output data of the indicator.
func (m *MoneyFlowIndex) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.MoneyFlowIndex,
		Mnemonic:    m.LineIndicator.Mnemonic,
		Description: m.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(MoneyFlowIndexValue),
				Type:        outputs.ScalarType,
				Mnemonic:    m.LineIndicator.Mnemonic,
				Description: m.LineIndicator.Description,
			},
		},
	}
}

// Update updates the indicator with the given sample using volume = 1.
// This satisfies the LineIndicator updateFn signature.
func (m *MoneyFlowIndex) Update(sample float64) float64 {
	return m.UpdateWithVolume(sample, 1)
}

// UpdateWithVolume updates the indicator with the given sample and volume.
func (m *MoneyFlowIndex) UpdateWithVolume(sample, volume float64) float64 {
	if math.IsNaN(sample) || math.IsNaN(volume) {
		return math.NaN()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	lengthMinOne := m.length - 1

	if m.primed {
		m.negativeSum -= m.negativeBuffer[m.bufferLowIndex]
		m.positiveSum -= m.positiveBuffer[m.bufferLowIndex]

		amount := sample * volume
		diff := sample - m.previousSample

		if diff < 0 {
			m.negativeBuffer[m.bufferIndex] = amount
			m.positiveBuffer[m.bufferIndex] = 0
			m.negativeSum += amount
		} else if diff > 0 {
			m.negativeBuffer[m.bufferIndex] = 0
			m.positiveBuffer[m.bufferIndex] = amount
			m.positiveSum += amount
		} else {
			m.negativeBuffer[m.bufferIndex] = 0
			m.positiveBuffer[m.bufferIndex] = 0
		}

		sum := m.positiveSum + m.negativeSum
		if sum < 1 {
			m.value = 0
		} else {
			m.value = 100 * m.positiveSum / sum
		}

		m.bufferIndex++
		if m.bufferIndex > lengthMinOne {
			m.bufferIndex = 0
		}

		m.bufferLowIndex++
		if m.bufferLowIndex > lengthMinOne {
			m.bufferLowIndex = 0
		}
	} else if m.bufferCount == 0 {
		m.bufferCount++
	} else {
		amount := sample * volume
		diff := sample - m.previousSample

		if diff < 0 {
			m.negativeBuffer[m.bufferIndex] = amount
			m.positiveBuffer[m.bufferIndex] = 0
			m.negativeSum += amount
		} else if diff > 0 {
			m.negativeBuffer[m.bufferIndex] = 0
			m.positiveBuffer[m.bufferIndex] = amount
			m.positiveSum += amount
		} else {
			m.negativeBuffer[m.bufferIndex] = 0
			m.positiveBuffer[m.bufferIndex] = 0
		}

		if m.length == m.bufferCount {
			sum := m.positiveSum + m.negativeSum
			if sum < 1 {
				m.value = 0
			} else {
				m.value = 100 * m.positiveSum / sum
			}

			m.primed = true
		}

		m.bufferIndex++
		if m.bufferIndex > lengthMinOne {
			m.bufferIndex = 0
		}

		m.bufferCount++
	}

	m.previousSample = sample

	return m.value
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to use bar volume.
func (m *MoneyFlowIndex) UpdateBar(sample *entities.Bar) core.Output {
	price := m.barFunc(sample)
	value := m.UpdateWithVolume(price, sample.Volume)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
