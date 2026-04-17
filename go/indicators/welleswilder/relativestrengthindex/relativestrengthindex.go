package relativestrengthindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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
func (r *RelativeStrengthIndex) IsPrimed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.primed
}

// Metadata describes an output data of the indicator.
func (r *RelativeStrengthIndex) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.RelativeStrengthIndex,
		Mnemonic:    r.LineIndicator.Mnemonic,
		Description: r.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(RelativeStrengthIndexValue),
				Type:        outputs.ScalarType,
				Mnemonic:    r.LineIndicator.Mnemonic,
				Description: r.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the indicator given the next sample.
func (r *RelativeStrengthIndex) Update(sample float64) float64 {
	const epsilon = 1e-8

	if math.IsNaN(sample) {
		return sample
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.count++

	if r.count == 0 {
		r.previousSample = sample

		return r.value
	}

	temp := sample - r.previousSample
	r.previousSample = sample

	if !r.primed {
		// Accumulation phase: count 1..length-1.
		if temp < 0 {
			r.previousLoss -= temp
		} else {
			r.previousGain += temp
		}

		if r.count < r.length {
			return r.value
		}

		// Priming: count == length.
		r.previousGain /= float64(r.length)
		r.previousLoss /= float64(r.length)
		r.primed = true
	} else {
		// Wilder's smoothing.
		r.previousGain *= float64(r.length - 1)
		r.previousLoss *= float64(r.length - 1)

		if temp < 0 {
			r.previousLoss -= temp
		} else {
			r.previousGain += temp
		}

		r.previousGain /= float64(r.length)
		r.previousLoss /= float64(r.length)
	}

	sum := r.previousGain + r.previousLoss
	if sum > epsilon {
		r.value = 100 * r.previousGain / sum
	} else {
		r.value = 0
	}

	return r.value
}
