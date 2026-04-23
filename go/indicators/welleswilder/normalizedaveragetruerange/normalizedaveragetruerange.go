package normalizedaveragetruerange

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/welleswilder/averagetruerange"
)

const (
	natrMnemonic    = "natr"
	natrDescription = "Normalized Average True Range"
)

// NormalizedAverageTrueRange is Welles Wilder's Normalized Average True Range indicator.
//
// NATR is calculated as (ATR / close) * 100, where ATR is the Average True Range.
// If close == 0, the result is 0 (not division by zero).
// The indicator is not primed during the first length updates.
//
// Reference:
//
// Forman, John (2006). "Cross-Market Evaluations With Normalized Average True Range",
// Technical Analysis of Stocks & Commodities (TASC), May 2006, pp. 60-63.
type NormalizedAverageTrueRange struct {
	mu               sync.RWMutex
	length           int
	value            float64
	primed           bool
	averageTrueRange *averagetruerange.AverageTrueRange
}

// NewNormalizedAverageTrueRange returns a new instance of the Normalized Average True Range indicator.
func NewNormalizedAverageTrueRange(length int) (*NormalizedAverageTrueRange, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	atr, _ := averagetruerange.NewAverageTrueRange(length)

	return &NormalizedAverageTrueRange{
		length:           length,
		value:            math.NaN(),
		averageTrueRange: atr,
	}, nil
}

// Length returns the length parameter.
func (s *NormalizedAverageTrueRange) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *NormalizedAverageTrueRange) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *NormalizedAverageTrueRange) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.NormalizedAverageTrueRange,
		natrMnemonic,
		natrDescription,
		[]core.OutputText{
			{Mnemonic: natrMnemonic, Description: natrDescription},
		},
	)
}

// Update updates the Normalized Average True Range given the next bar's close, high, and low values.
func (s *NormalizedAverageTrueRange) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	atrValue := s.averageTrueRange.Update(close, high, low)

	if s.averageTrueRange.IsPrimed() {
		s.primed = true

		if close == 0 {
			s.value = 0
		} else {
			s.value = (atrValue / close) * 100 //nolint:mnd
		}
	}

	if s.primed {
		return s.value
	}

	return math.NaN()
}

// UpdateSample updates the Normalized Average True Range using a single sample value
// as a substitute for high, low, and close.
func (s *NormalizedAverageTrueRange) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *NormalizedAverageTrueRange) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *NormalizedAverageTrueRange) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *NormalizedAverageTrueRange) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *NormalizedAverageTrueRange) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}
