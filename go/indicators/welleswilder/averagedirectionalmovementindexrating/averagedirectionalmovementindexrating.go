package averagedirectionalmovementindexrating

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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
func (a *AverageDirectionalMovementIndexRating) Length() int {
	return a.length
}

// IsPrimed indicates whether the indicator is primed.
func (a *AverageDirectionalMovementIndexRating) IsPrimed() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.primed
}

// Metadata describes the output data of the indicator.
func (a *AverageDirectionalMovementIndexRating) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.AverageDirectionalMovementIndexRating,
		Mnemonic:    adxrMnemonic,
		Description: adxrDescription,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(AverageDirectionalMovementIndexRatingValue),
				Type:        outputs.ScalarType,
				Mnemonic:    adxrMnemonic,
				Description: adxrDescription,
			},
			{
				Kind:        int(AverageDirectionalMovementIndexValue),
				Type:        outputs.ScalarType,
				Mnemonic:    "adx",
				Description: "Average Directional Movement Index",
			},
			{
				Kind:        int(DirectionalMovementIndexValue),
				Type:        outputs.ScalarType,
				Mnemonic:    "dx",
				Description: "Directional Movement Index",
			},
			{
				Kind:        int(DirectionalIndicatorPlusValue),
				Type:        outputs.ScalarType,
				Mnemonic:    "+di",
				Description: "Directional Indicator Plus",
			},
			{
				Kind:        int(DirectionalIndicatorMinusValue),
				Type:        outputs.ScalarType,
				Mnemonic:    "-di",
				Description: "Directional Indicator Minus",
			},
			{
				Kind:        int(DirectionalMovementPlusValue),
				Type:        outputs.ScalarType,
				Mnemonic:    "+dm",
				Description: "Directional Movement Plus",
			},
			{
				Kind:        int(DirectionalMovementMinusValue),
				Type:        outputs.ScalarType,
				Mnemonic:    "-dm",
				Description: "Directional Movement Minus",
			},
			{
				Kind:        int(AverageTrueRangeValue),
				Type:        outputs.ScalarType,
				Mnemonic:    "atr",
				Description: "Average True Range",
			},
			{
				Kind:        int(TrueRangeValue),
				Type:        outputs.ScalarType,
				Mnemonic:    "tr",
				Description: "True Range",
			},
		},
	}
}

// Update updates the Average Directional Movement Index Rating given the next bar's close, high, and low values.
func (a *AverageDirectionalMovementIndexRating) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	adxValue := a.averageDirectionalMovementIndex.Update(close, high, low)

	if !a.averageDirectionalMovementIndex.IsPrimed() {
		return math.NaN()
	}

	// Store ADX value in circular buffer.
	a.buffer[a.bufferIndex] = adxValue
	a.bufferIndex = (a.bufferIndex + 1) % a.bufferSize
	a.bufferCount++

	if a.bufferCount < a.bufferSize {
		return math.NaN()
	}

	// The oldest value in the buffer is at bufferIndex (since we just advanced it).
	oldADX := a.buffer[a.bufferIndex%a.bufferSize]
	a.value = (adxValue + oldADX) / 2 //nolint:mnd
	a.primed = true

	return a.value
}

// UpdateSample updates the Average Directional Movement Index Rating using a single sample value
// as a substitute for close, high, and low.
func (a *AverageDirectionalMovementIndexRating) UpdateSample(sample float64) float64 {
	return a.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (a *AverageDirectionalMovementIndexRating) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (a *AverageDirectionalMovementIndexRating) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (a *AverageDirectionalMovementIndexRating) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (a *AverageDirectionalMovementIndexRating) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(v, v, v)}

	return output
}
