package averagedirectionalmovementindexrating

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/welleswilder/averagedirectionalmovementindex"
)

const (
	adxrMnemonic    = "adxr"
	adxrDescription = "Average Directional Movement Index Rating"
)

// AverageDirectionalMovementIndexRating is Welles Wilder's Average Directional Movement Index Rating (ADXR).
//
// The average directional movement index rating averages the current ADX value with
// the ADX value from (length - 1) periods ago. It is calculated as:
//
//	ADXR = (ADX[current] + ADX[current - (length - 1)]) / 2
//
// The indicator requires close, high, and low values.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type AverageDirectionalMovementIndexRating struct {
	mu                              sync.RWMutex
	length                          int
	bufferSize                      int
	buffer                          []float64
	bufferIndex                     int
	bufferCount                     int
	primed                          bool
	value                           float64
	averageDirectionalMovementIndex *averagedirectionalmovementindex.AverageDirectionalMovementIndex
}

// NewAverageDirectionalMovementIndexRating returns a new instance of the Average Directional Movement Index Rating indicator.
func NewAverageDirectionalMovementIndexRating(length int) (*AverageDirectionalMovementIndexRating, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	adx, err := averagedirectionalmovementindex.NewAverageDirectionalMovementIndex(length)
	if err != nil {
		return nil, fmt.Errorf("failed to create average directional movement index: %w", err)
	}

	// Need to store `length` ADX values to look back (length-1) periods.
	bufferSize := length

	return &AverageDirectionalMovementIndexRating{
		length:                          length,
		bufferSize:                      bufferSize,
		buffer:                          make([]float64, bufferSize),
		value:                           math.NaN(),
		averageDirectionalMovementIndex: adx,
	}, nil
}

// Length returns the length parameter.
func (s *AverageDirectionalMovementIndexRating) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *AverageDirectionalMovementIndexRating) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *AverageDirectionalMovementIndexRating) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AverageDirectionalMovementIndexRating,
		adxrMnemonic,
		adxrDescription,
		[]core.OutputText{
			{Mnemonic: adxrMnemonic, Description: adxrDescription},
			{Mnemonic: "adx", Description: "Average Directional Movement Index"},
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

// Update updates the Average Directional Movement Index Rating given the next bar's close, high, and low values.
func (s *AverageDirectionalMovementIndexRating) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	adxValue := s.averageDirectionalMovementIndex.Update(close, high, low)

	if !s.averageDirectionalMovementIndex.IsPrimed() {
		return math.NaN()
	}

	// Store ADX value in circular buffer.
	s.buffer[s.bufferIndex] = adxValue
	s.bufferIndex = (s.bufferIndex + 1) % s.bufferSize
	s.bufferCount++

	if s.bufferCount < s.bufferSize {
		return math.NaN()
	}

	// The oldest value in the buffer is at bufferIndex (since we just advanced it).
	oldADX := s.buffer[s.bufferIndex%s.bufferSize]
	s.value = (adxValue + oldADX) / 2 //nolint:mnd
	s.primed = true

	return s.value
}

// UpdateSample updates the Average Directional Movement Index Rating using a single sample value
// as a substitute for close, high, and low.
func (s *AverageDirectionalMovementIndexRating) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *AverageDirectionalMovementIndexRating) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *AverageDirectionalMovementIndexRating) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *AverageDirectionalMovementIndexRating) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *AverageDirectionalMovementIndexRating) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}
