package averagetruerange

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
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
func (s *AverageTrueRange) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *AverageTrueRange) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *AverageTrueRange) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AverageTrueRange,
		atrMnemonic,
		atrDescription,
		[]core.OutputText{
			{Mnemonic: atrMnemonic, Description: atrDescription},
		},
	)
}

// Update updates the Average True Range given the next bar's close, high, and low values.
func (s *AverageTrueRange) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	trueRangeValue := s.trueRange.Update(close, high, low)

	if s.lastIndex == 0 {
		s.value = trueRangeValue

		if s.stage == 0 {
			s.stage++
		} else if s.stage == 1 {
			s.stage++
			s.primed = true
		}

		return s.value
	}

	if s.stage > 1 {
		// Wilder smoothing method.
		s.value *= float64(s.lastIndex)
		s.value += trueRangeValue
		s.value /= float64(s.length)

		return s.value
	}

	if s.stage == 1 {
		s.windowSum += trueRangeValue
		s.window[s.windowCount] = trueRangeValue
		s.windowCount++

		if s.windowCount == s.length {
			s.stage++
			s.primed = true
			s.value = s.windowSum / float64(s.length)
		}

		if s.primed {
			return s.value
		}

		return math.NaN()
	}

	// The very first sample is used by the True Range.
	s.stage++

	return math.NaN()
}

// UpdateSample updates the Average True Range using a single sample value
// as a substitute for high, low, and close.
func (s *AverageTrueRange) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *AverageTrueRange) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *AverageTrueRange) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *AverageTrueRange) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *AverageTrueRange) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}
