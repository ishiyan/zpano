package directionalmovementindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
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
func (s *DirectionalMovementIndex) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *DirectionalMovementIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.directionalIndicatorPlus.IsPrimed() && s.directionalIndicatorMinus.IsPrimed()
}

// Metadata describes the output data of the indicator.
func (s *DirectionalMovementIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DirectionalMovementIndex,
		dxMnemonic,
		dxDescription,
		[]core.OutputText{
			{Mnemonic: dxMnemonic, Description: dxDescription},
			{Mnemonic: "+di", Description: "Directional Indicator Plus"},
			{Mnemonic: "-di", Description: "Directional Indicator Minus"},
			{Mnemonic: "+dm", Description: "Directional Movement Plus"},
			{Mnemonic: "-dm", Description: "Directional Movement Minus"},
			{Mnemonic: "atr", Description: "Average True Range"},
			{Mnemonic: "tr", Description: "True Range"},
		},
	)
}

// Update updates the Directional Movement Index given the next bar's close, high, and low values.
func (s *DirectionalMovementIndex) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	dipValue := s.directionalIndicatorPlus.Update(close, high, low)
	dimValue := s.directionalIndicatorMinus.Update(close, high, low)

	if s.directionalIndicatorPlus.IsPrimed() && s.directionalIndicatorMinus.IsPrimed() {
		sum := dipValue + dimValue

		if math.Abs(sum) < epsilon {
			s.value = 0
		} else {
			s.value = 100 * math.Abs(dipValue-dimValue) / sum //nolint:mnd
		}

		return s.value
	}

	return math.NaN()
}

// UpdateSample updates the Directional Movement Index using a single sample value
// as a substitute for close, high, and low.
func (s *DirectionalMovementIndex) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *DirectionalMovementIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DirectionalMovementIndex) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *DirectionalMovementIndex) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *DirectionalMovementIndex) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}
