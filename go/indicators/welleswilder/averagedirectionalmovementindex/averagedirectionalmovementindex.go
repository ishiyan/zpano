package averagedirectionalmovementindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/welleswilder/directionalmovementindex"
)

const (
	adxMnemonic    = "adx"
	adxDescription = "Average Directional Movement Index"
)

// AverageDirectionalMovementIndex is Welles Wilder's Average Directional Movement Index (ADX).
//
// The average directional movement index smooths the directional movement index (DX)
// using Wilder's smoothing technique. It is calculated as:
//
//	Initial ADX = SMA of first `length` DX values
//	Subsequent ADX = (previousADX * (length-1) + DX) / length
//
// The indicator requires close, high, and low values.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type AverageDirectionalMovementIndex struct {
	mu                       sync.RWMutex
	length                   int
	lengthMinusOne           float64
	count                    int
	sum                      float64
	primed                   bool
	value                    float64
	directionalMovementIndex *directionalmovementindex.DirectionalMovementIndex
}

// NewAverageDirectionalMovementIndex returns a new instance of the Average Directional Movement Index indicator.
func NewAverageDirectionalMovementIndex(length int) (*AverageDirectionalMovementIndex, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	dx, err := directionalmovementindex.NewDirectionalMovementIndex(length)
	if err != nil {
		return nil, fmt.Errorf("failed to create directional movement index: %w", err)
	}

	return &AverageDirectionalMovementIndex{
		length:                   length,
		lengthMinusOne:           float64(length - 1),
		value:                    math.NaN(),
		directionalMovementIndex: dx,
	}, nil
}

// Length returns the length parameter.
func (s *AverageDirectionalMovementIndex) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *AverageDirectionalMovementIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *AverageDirectionalMovementIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AverageDirectionalMovementIndex,
		adxMnemonic,
		adxDescription,
		[]core.OutputText{
			{Mnemonic: adxMnemonic, Description: adxDescription},
			{Mnemonic: "dx", Description: "Directional Movement Index"},
			{Mnemonic: "+di", Description: "Directional Indicator Plus"},
			{Mnemonic: "-di", Description: "Directional Indicator Minus"},
			{Mnemonic: "+dm", Description: "Directional Movement Plus"},
			{Mnemonic: "-dm", Description: "Directional Movement Minus"},
			{Mnemonic: "atr", Description: "Average True Range"},
			{Mnemonic: "tr", Description: "True Range"},
		},
	)
}

// Update updates the Average Directional Movement Index given the next bar's close, high, and low values.
func (s *AverageDirectionalMovementIndex) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	dxValue := s.directionalMovementIndex.Update(close, high, low)

	if !s.directionalMovementIndex.IsPrimed() {
		return math.NaN()
	}

	if s.primed {
		s.value = (s.value*s.lengthMinusOne + dxValue) / float64(s.length)
		return s.value
	}

	s.count++
	s.sum += dxValue

	if s.count == s.length {
		s.value = s.sum / float64(s.length)
		s.primed = true

		return s.value
	}

	return math.NaN()
}

// UpdateSample updates the Average Directional Movement Index using a single sample value
// as a substitute for close, high, and low.
func (s *AverageDirectionalMovementIndex) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *AverageDirectionalMovementIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *AverageDirectionalMovementIndex) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *AverageDirectionalMovementIndex) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *AverageDirectionalMovementIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}
