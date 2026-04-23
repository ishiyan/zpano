package directionalindicatorminus

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/welleswilder/averagetruerange"
	"zpano/indicators/welleswilder/directionalmovementminus"
)

const (
	dimMnemonic    = "-di"
	dimDescription = "Directional Indicator Minus"
	epsilon        = 1e-8
)

// DirectionalIndicatorMinus is Welles Wilder's Directional Indicator Minus (-DI).
//
// The directional indicator minus measures the percentage of the average true range
// that is attributable to downward movement. It is calculated as:
//
//	-DI = 100 * -DM(n) / (ATR * length)
//
// where -DM(n) is the Wilder-smoothed directional movement minus and ATR is the
// average true range over the same length.
//
// The indicator requires close, high, and low values. ATR uses all three;
// directional movement minus uses high and low.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type DirectionalIndicatorMinus struct {
	mu                       sync.RWMutex
	length                   int
	value                    float64
	averageTrueRange         *averagetruerange.AverageTrueRange
	directionalMovementMinus *directionalmovementminus.DirectionalMovementMinus
}

// NewDirectionalIndicatorMinus returns a new instance of the Directional Indicator Minus indicator.
func NewDirectionalIndicatorMinus(length int) (*DirectionalIndicatorMinus, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	atr, err := averagetruerange.NewAverageTrueRange(length)
	if err != nil {
		return nil, fmt.Errorf("failed to create average true range: %w", err)
	}

	dmm, err := directionalmovementminus.NewDirectionalMovementMinus(length)
	if err != nil {
		return nil, fmt.Errorf("failed to create directional movement minus: %w", err)
	}

	return &DirectionalIndicatorMinus{
		length:                   length,
		value:                    math.NaN(),
		averageTrueRange:         atr,
		directionalMovementMinus: dmm,
	}, nil
}

// Length returns the length parameter.
func (s *DirectionalIndicatorMinus) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *DirectionalIndicatorMinus) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.averageTrueRange.IsPrimed() && s.directionalMovementMinus.IsPrimed()
}

// Metadata describes the output data of the indicator.
func (s *DirectionalIndicatorMinus) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DirectionalIndicatorMinus,
		dimMnemonic,
		dimDescription,
		[]core.OutputText{
			{Mnemonic: dimMnemonic, Description: dimDescription},
			{Mnemonic: "-dm", Description: "Directional Movement Minus"},
			{Mnemonic: "atr", Description: "Average True Range"},
			{Mnemonic: "tr", Description: "True Range"},
		},
	)
}

// Update updates the Directional Indicator Minus given the next bar's close, high, and low values.
func (s *DirectionalIndicatorMinus) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	atrValue := s.averageTrueRange.Update(close, high, low)
	dmmValue := s.directionalMovementMinus.Update(high, low)

	if s.averageTrueRange.IsPrimed() && s.directionalMovementMinus.IsPrimed() {
		atrScaled := atrValue * float64(s.length)

		if math.Abs(atrScaled) < epsilon {
			s.value = 0
		} else {
			s.value = 100 * dmmValue / atrScaled //nolint:mnd
		}

		return s.value
	}

	return math.NaN()
}

// UpdateSample updates the Directional Indicator Minus using a single sample value
// as a substitute for close, high, and low.
func (s *DirectionalIndicatorMinus) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *DirectionalIndicatorMinus) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DirectionalIndicatorMinus) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *DirectionalIndicatorMinus) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *DirectionalIndicatorMinus) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}
