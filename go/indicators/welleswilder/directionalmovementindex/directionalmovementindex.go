package directionalmovementindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/welleswilder/directionalindicatorminus"
	"zpano/indicators/welleswilder/directionalindicatorplus"
)

const (
	dxMnemonic    = "dx"
	dxDescription = "Directional Movement Index"
	epsilon       = 1e-8
)

// DirectionalMovementIndex is Welles Wilder's Directional Movement Index (DX).
//
// The directional movement index measures the strength of a trend by comparing
// the positive and negative directional indicators. It is calculated as:
//
//	DX = 100 * |+DI - -DI| / (+DI + -DI)
//
// where +DI is the directional indicator plus and -DI is the directional
// indicator minus, both computed over the same length.
//
// The indicator requires close, high, and low values.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type DirectionalMovementIndex struct {
	mu                        sync.RWMutex
	length                    int
	value                     float64
	directionalIndicatorPlus  *directionalindicatorplus.DirectionalIndicatorPlus
	directionalIndicatorMinus *directionalindicatorminus.DirectionalIndicatorMinus
}

// NewDirectionalMovementIndex returns a new instance of the Directional Movement Index indicator.
func NewDirectionalMovementIndex(length int) (*DirectionalMovementIndex, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	dip, err := directionalindicatorplus.NewDirectionalIndicatorPlus(length)
	if err != nil {
		return nil, fmt.Errorf("failed to create directional indicator plus: %w", err)
	}

	dim, err := directionalindicatorminus.NewDirectionalIndicatorMinus(length)
	if err != nil {
		return nil, fmt.Errorf("failed to create directional indicator minus: %w", err)
	}

	return &DirectionalMovementIndex{
		length:                    length,
		value:                     math.NaN(),
		directionalIndicatorPlus:  dip,
		directionalIndicatorMinus: dim,
	}, nil
}

// Length returns the length parameter.
func (d *DirectionalMovementIndex) Length() int {
	return d.length
}

// IsPrimed indicates whether the indicator is primed.
func (d *DirectionalMovementIndex) IsPrimed() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.directionalIndicatorPlus.IsPrimed() && d.directionalIndicatorMinus.IsPrimed()
}

// Metadata describes the output data of the indicator.
func (d *DirectionalMovementIndex) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.DirectionalMovementIndex,
		Mnemonic:    dxMnemonic,
		Description: dxDescription,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(DirectionalMovementIndexValue),
				Type:        outputs.ScalarType,
				Mnemonic:    dxMnemonic,
				Description: dxDescription,
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

// Update updates the Directional Movement Index given the next bar's close, high, and low values.
func (d *DirectionalMovementIndex) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	dipValue := d.directionalIndicatorPlus.Update(close, high, low)
	dimValue := d.directionalIndicatorMinus.Update(close, high, low)

	if d.directionalIndicatorPlus.IsPrimed() && d.directionalIndicatorMinus.IsPrimed() {
		sum := dipValue + dimValue

		if math.Abs(sum) < epsilon {
			d.value = 0
		} else {
			d.value = 100 * math.Abs(dipValue-dimValue) / sum //nolint:mnd
		}

		return d.value
	}

	return math.NaN()
}

// UpdateSample updates the Directional Movement Index using a single sample value
// as a substitute for close, high, and low.
func (d *DirectionalMovementIndex) UpdateSample(sample float64) float64 {
	return d.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (d *DirectionalMovementIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (d *DirectionalMovementIndex) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (d *DirectionalMovementIndex) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (d *DirectionalMovementIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(v, v, v)}

	return output
}
