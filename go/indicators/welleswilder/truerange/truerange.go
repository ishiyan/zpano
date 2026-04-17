package truerange

import (
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

const (
	trMnemonic    = "tr"
	trDescription = "True Range"
)

// TrueRange is Welles Wilder's True Range indicator.
//
// The True Range is defined as the largest of:
//   - the distance from today's high to today's low
//   - the distance from yesterday's close to today's high
//   - the distance from yesterday's close to today's low
//
// The first update stores the close and returns NaN (not primed).
// The indicator is primed from the second update onward.
//
// Unlike most indicators, TrueRange requires bar data (high, low, close)
// and does not use a single bar component. For scalar, quote, and trade updates,
// the single value is used as a substitute for high, low, and close.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type TrueRange struct {
	mu            sync.RWMutex
	previousClose float64
	value         float64
	primed        bool
}

// NewTrueRange returns a new instance of the True Range indicator.
func NewTrueRange() *TrueRange {
	return &TrueRange{
		previousClose: math.NaN(),
		value:         math.NaN(),
	}
}

// IsPrimed indicates whether the indicator is primed.
func (tr *TrueRange) IsPrimed() bool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	return tr.primed
}

// Metadata describes the output data of the indicator.
func (tr *TrueRange) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.TrueRange,
		Mnemonic:    trMnemonic,
		Description: trDescription,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(TrueRangeValue),
				Type:        outputs.ScalarType,
				Mnemonic:    trMnemonic,
				Description: trDescription,
			},
		},
	}
}

// Update updates the True Range given the next bar's close, high, and low values.
func (tr *TrueRange) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	if !tr.primed {
		if math.IsNaN(tr.previousClose) {
			tr.previousClose = close

			return math.NaN()
		}

		tr.primed = true
	}

	greatest := high - low

	if temp := math.Abs(high - tr.previousClose); greatest < temp {
		greatest = temp
	}

	if temp := math.Abs(low - tr.previousClose); greatest < temp {
		greatest = temp
	}

	tr.value = greatest
	tr.previousClose = close

	return tr.value
}

// UpdateSample updates the True Range using a single sample value
// as a substitute for high, low, and close.
func (tr *TrueRange) UpdateSample(sample float64) float64 {
	return tr.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (tr *TrueRange) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: tr.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (tr *TrueRange) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: tr.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (tr *TrueRange) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: tr.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (tr *TrueRange) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: tr.Update(v, v, v)}

	return output
}
