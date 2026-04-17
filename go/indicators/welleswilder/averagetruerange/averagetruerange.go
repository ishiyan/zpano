package averagetruerange

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/welleswilder/truerange"
)

const (
	atrMnemonic    = "atr"
	atrDescription = "Average True Range"
)

// AverageTrueRange is Welles Wilder's Average True Range indicator.
//
// ATR averages True Range (TR) values over the specified length using the Wilder method:
//   - multiply the previous value by (length - 1)
//   - add the current TR value
//   - divide by length
//
// The initial ATR value is a simple average of the first length TR values.
// The indicator is not primed during the first length updates.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type AverageTrueRange struct {
	mu          sync.RWMutex
	length      int
	lastIndex   int
	stage       int
	windowCount int
	window      []float64
	windowSum   float64
	value       float64
	primed      bool
	trueRange   *truerange.TrueRange
}

// NewAverageTrueRange returns a new instance of the Average True Range indicator.
func NewAverageTrueRange(length int) (*AverageTrueRange, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	atr := &AverageTrueRange{
		length:    length,
		lastIndex: length - 1,
		value:     math.NaN(),
		trueRange: truerange.NewTrueRange(),
	}

	if atr.lastIndex > 0 {
		atr.window = make([]float64, length)
	}

	return atr, nil
}

// Length returns the length parameter.
func (a *AverageTrueRange) Length() int {
	return a.length
}

// IsPrimed indicates whether the indicator is primed.
func (a *AverageTrueRange) IsPrimed() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.primed
}

// Metadata describes the output data of the indicator.
func (a *AverageTrueRange) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.AverageTrueRange,
		Mnemonic:    atrMnemonic,
		Description: atrDescription,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(AverageTrueRangeValue),
				Type:        outputs.ScalarType,
				Mnemonic:    atrMnemonic,
				Description: atrDescription,
			},
		},
	}
}

// Update updates the Average True Range given the next bar's close, high, and low values.
func (a *AverageTrueRange) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	trueRangeValue := a.trueRange.Update(close, high, low)

	if a.lastIndex == 0 {
		a.value = trueRangeValue

		if a.stage == 0 {
			a.stage++
		} else if a.stage == 1 {
			a.stage++
			a.primed = true
		}

		return a.value
	}

	if a.stage > 1 {
		// Wilder smoothing method.
		a.value *= float64(a.lastIndex)
		a.value += trueRangeValue
		a.value /= float64(a.length)

		return a.value
	}

	if a.stage == 1 {
		a.windowSum += trueRangeValue
		a.window[a.windowCount] = trueRangeValue
		a.windowCount++

		if a.windowCount == a.length {
			a.stage++
			a.primed = true
			a.value = a.windowSum / float64(a.length)
		}

		if a.primed {
			return a.value
		}

		return math.NaN()
	}

	// The very first sample is used by the True Range.
	a.stage++

	return math.NaN()
}

// UpdateSample updates the Average True Range using a single sample value
// as a substitute for high, low, and close.
func (a *AverageTrueRange) UpdateSample(sample float64) float64 {
	return a.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (a *AverageTrueRange) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (a *AverageTrueRange) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (a *AverageTrueRange) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (a *AverageTrueRange) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(v, v, v)}

	return output
}
