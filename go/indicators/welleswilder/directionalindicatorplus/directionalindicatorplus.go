package directionalindicatorplus

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/welleswilder/averagetruerange"
	"zpano/indicators/welleswilder/directionalmovementplus"
)

const (
	dipMnemonic    = "+di"
	dipDescription = "Directional Indicator Plus"
	epsilon        = 1e-8
)

// DirectionalIndicatorPlus is Welles Wilder's Directional Indicator Plus (+DI).
//
// The directional indicator plus measures the percentage of the average true range
// that is attributable to upward movement. It is calculated as:
//
//	+DI = 100 * +DM(n) / (ATR * length)
//
// where +DM(n) is the Wilder-smoothed directional movement plus and ATR is the
// average true range over the same length.
//
// The indicator requires close, high, and low values. ATR uses all three;
// directional movement plus uses high and low.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type DirectionalIndicatorPlus struct {
	mu                      sync.RWMutex
	length                  int
	value                   float64
	averageTrueRange        *averagetruerange.AverageTrueRange
	directionalMovementPlus *directionalmovementplus.DirectionalMovementPlus
}

// NewDirectionalIndicatorPlus returns a new instance of the Directional Indicator Plus indicator.
func NewDirectionalIndicatorPlus(length int) (*DirectionalIndicatorPlus, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	atr, err := averagetruerange.NewAverageTrueRange(length)
	if err != nil {
		return nil, fmt.Errorf("failed to create average true range: %w", err)
	}

	dmp, err := directionalmovementplus.NewDirectionalMovementPlus(length)
	if err != nil {
		return nil, fmt.Errorf("failed to create directional movement plus: %w", err)
	}

	return &DirectionalIndicatorPlus{
		length:                  length,
		value:                   math.NaN(),
		averageTrueRange:        atr,
		directionalMovementPlus: dmp,
	}, nil
}

// Length returns the length parameter.
func (s *DirectionalIndicatorPlus) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *DirectionalIndicatorPlus) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.averageTrueRange.IsPrimed() && s.directionalMovementPlus.IsPrimed()
}

// Metadata describes the output data of the indicator.
func (s *DirectionalIndicatorPlus) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DirectionalIndicatorPlus,
		dipMnemonic,
		dipDescription,
		[]core.OutputText{
			{Mnemonic: dipMnemonic, Description: dipDescription},
			{Mnemonic: "+dm", Description: "Directional Movement Plus"},
			{Mnemonic: "atr", Description: "Average True Range"},
			{Mnemonic: "tr", Description: "True Range"},
		},
	)
}

// Update updates the Directional Indicator Plus given the next bar's close, high, and low values.
func (s *DirectionalIndicatorPlus) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	atrValue := s.averageTrueRange.Update(close, high, low)
	dmpValue := s.directionalMovementPlus.Update(high, low)

	if s.averageTrueRange.IsPrimed() && s.directionalMovementPlus.IsPrimed() {
		atrScaled := atrValue * float64(s.length)

		if math.Abs(atrScaled) < epsilon {
			s.value = 0
		} else {
			s.value = 100 * dmpValue / atrScaled //nolint:mnd
		}

		return s.value
	}

	return math.NaN()
}

// UpdateSample updates the Directional Indicator Plus using a single sample value
// as a substitute for close, high, and low.
func (s *DirectionalIndicatorPlus) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *DirectionalIndicatorPlus) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DirectionalIndicatorPlus) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *DirectionalIndicatorPlus) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *DirectionalIndicatorPlus) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}
