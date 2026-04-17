package averagedirectionalmovementindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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
func (a *AverageDirectionalMovementIndex) Length() int {
	return a.length
}

// IsPrimed indicates whether the indicator is primed.
func (a *AverageDirectionalMovementIndex) IsPrimed() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.primed
}

// Metadata describes the output data of the indicator.
func (a *AverageDirectionalMovementIndex) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.AverageDirectionalMovementIndex,
		Mnemonic:    adxMnemonic,
		Description: adxDescription,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(AverageDirectionalMovementIndexValue),
				Type:        outputs.ScalarType,
				Mnemonic:    adxMnemonic,
				Description: adxDescription,
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

// Update updates the Average Directional Movement Index given the next bar's close, high, and low values.
func (a *AverageDirectionalMovementIndex) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	dxValue := a.directionalMovementIndex.Update(close, high, low)

	if !a.directionalMovementIndex.IsPrimed() {
		return math.NaN()
	}

	if a.primed {
		a.value = (a.value*a.lengthMinusOne + dxValue) / float64(a.length)
		return a.value
	}

	a.count++
	a.sum += dxValue

	if a.count == a.length {
		a.value = a.sum / float64(a.length)
		a.primed = true

		return a.value
	}

	return math.NaN()
}

// UpdateSample updates the Average Directional Movement Index using a single sample value
// as a substitute for close, high, and low.
func (a *AverageDirectionalMovementIndex) UpdateSample(sample float64) float64 {
	return a.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (a *AverageDirectionalMovementIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (a *AverageDirectionalMovementIndex) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (a *AverageDirectionalMovementIndex) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (a *AverageDirectionalMovementIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: a.Update(v, v, v)}

	return output
}
