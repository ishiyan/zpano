package directionalindicatorminus

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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
func (d *DirectionalIndicatorMinus) Length() int {
	return d.length
}

// IsPrimed indicates whether the indicator is primed.
func (d *DirectionalIndicatorMinus) IsPrimed() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.averageTrueRange.IsPrimed() && d.directionalMovementMinus.IsPrimed()
}

// Metadata describes the output data of the indicator.
func (d *DirectionalIndicatorMinus) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.DirectionalIndicatorMinus,
		Mnemonic:    dimMnemonic,
		Description: dimDescription,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(DirectionalIndicatorMinusValue),
				Type:        outputs.ScalarType,
				Mnemonic:    dimMnemonic,
				Description: dimDescription,
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

// Update updates the Directional Indicator Minus given the next bar's close, high, and low values.
func (d *DirectionalIndicatorMinus) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	atrValue := d.averageTrueRange.Update(close, high, low)
	dmmValue := d.directionalMovementMinus.Update(high, low)

	if d.averageTrueRange.IsPrimed() && d.directionalMovementMinus.IsPrimed() {
		atrScaled := atrValue * float64(d.length)

		if math.Abs(atrScaled) < epsilon {
			d.value = 0
		} else {
			d.value = 100 * dmmValue / atrScaled //nolint:mnd
		}

		return d.value
	}

	return math.NaN()
}

// UpdateSample updates the Directional Indicator Minus using a single sample value
// as a substitute for close, high, and low.
func (d *DirectionalIndicatorMinus) UpdateSample(sample float64) float64 {
	return d.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (d *DirectionalIndicatorMinus) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (d *DirectionalIndicatorMinus) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (d *DirectionalIndicatorMinus) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (d *DirectionalIndicatorMinus) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(v, v, v)}

	return output
}
