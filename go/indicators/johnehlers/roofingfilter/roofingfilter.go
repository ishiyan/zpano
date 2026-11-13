package roofingfilter

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// RoofingFilter implements Ehlers' Roofing Filter: high-pass detrending followed by a two-pole SuperSmoother.
type RoofingFilter struct {
	mu sync.RWMutex
	core.LineIndicator

	poles      int
	zeroMean   bool
	warmupBars int

	gain  float64
	decay float64
	c1    float64
	c2    float64
	c3    float64

	count        int
	input1       float64
	input2       float64
	hp1          float64
	hp2          float64
	filt1        float64
	filt2        float64
	zero1        float64
	currentValue float64
}

func smootherCoefficients(period float64) (float64, float64, float64) {
	smootherAngle := math.Sqrt2 * math.Pi / period
	a1 := math.Exp(-smootherAngle)
	c2 := 2.0 * a1 * math.Cos(smootherAngle)
	c3 := -a1 * a1
	return 1.0 - c2 - c3, c2, c3
}

// NewRoofingFilterDefault creates a RoofingFilter with Ehlers' default parameters.
func NewRoofingFilterDefault() (*RoofingFilter, error) {
	return NewRoofingFilter(DefaultParams())
}

// NewRoofingFilter returns an instance of the indicator created using supplied parameters.
func NewRoofingFilter(params *Params) (*RoofingFilter, error) {
	const (
		invalid = "invalid roofing filter parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
	)

	if params == nil {
		params = DefaultParams()
	}

	shortest := params.ShortestCyclePeriod
	if math.IsNaN(shortest) || math.IsInf(shortest, 0) || shortest < 2.0 {
		return nil, fmt.Errorf(fmts, invalid, "shortest_cycle_period must be finite and at least 2 bars")
	}

	longest := params.LongestCyclePeriod
	if math.IsNaN(longest) || math.IsInf(longest, 0) || longest <= shortest {
		return nil, fmt.Errorf(fmts, invalid, "longest_cycle_period must be finite and greater than shortest")
	}

	poles := params.HighpassPoles
	if poles != 1 && poles != 2 {
		return nil, fmt.Errorf(fmts, invalid, "highpass_poles must be integer 1 or 2")
	}

	if params.ZeroMean && poles != 1 {
		return nil, fmt.Errorf(fmts, invalid, "zero_mean requires highpass_poles=1")
	}

	barComp := params.BarComponent
	if barComp == 0 {
		barComp = entities.DefaultBarComponent
	}
	quoteComp := params.QuoteComponent
	if quoteComp == 0 {
		quoteComp = entities.DefaultQuoteComponent
	}
	tradeComp := params.TradeComponent
	if tradeComp == 0 {
		tradeComp = entities.DefaultTradeComponent
	}

	barFunc, err := entities.BarComponentFunc(barComp)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}
	quoteFunc, err := entities.QuoteComponentFunc(quoteComp)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}
	tradeFunc, err := entities.TradeComponentFunc(tradeComp)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	angle := math.Sqrt2 * math.Pi / longest
	if poles == 1 {
		angle = 2.0 * math.Pi / longest
	}
	tangent := math.Tan(angle / 2.0)
	alpha := 2.0 * tangent / (1.0 + tangent)
	gain := 1.0 - alpha/2.0
	decay := 1.0 - alpha
	c1, c2, c3 := smootherCoefficients(shortest)

	warmupBars := 3
	if poles == 2 || params.ZeroMean {
		warmupBars = 4
	}

	mnemonic := params.Mnemonic()
	description := "Roofing Filter " + mnemonic

	filter := &RoofingFilter{
		poles:        poles,
		zeroMean:     params.ZeroMean,
		warmupBars:   warmupBars,
		gain:         gain,
		decay:        decay,
		c1:           c1,
		c2:           c2,
		c3:           c3,
		currentValue: math.NaN(),
	}
	filter.LineIndicator = core.NewLineIndicator(
		mnemonic,
		description,
		barFunc,
		quoteFunc,
		tradeFunc,
		filter.Update,
	)
	return filter, nil
}

// Reset clears all streaming histories and priming state while preserving configuration.
func (s *RoofingFilter) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.count = 0
	s.input1 = 0
	s.input2 = 0
	s.hp1 = 0
	s.hp2 = 0
	s.filt1 = 0
	s.filt2 = 0
	s.zero1 = 0
	s.currentValue = math.NaN()
}

// WarmupBars returns the number of initial NaN outputs before priming.
func (s *RoofingFilter) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.warmupBars
}

// Value returns the latest calculated scalar value (NaN before priming).
func (s *RoofingFilter) Value() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentValue
}

// IsPrimed indicates whether the indicator has produced at least one primed output.
func (s *RoofingFilter) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count > s.warmupBars
}

// Metadata describes the output data of the indicator.
func (s *RoofingFilter) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.RoofingFilter,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// TryUpdate consumes one scalar sample and returns an error if sample is infinite.
func (s *RoofingFilter) TryUpdate(sample float64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsNaN(sample) {
		return math.NaN(), nil
	}
	if math.IsInf(sample, 0) {
		return math.NaN(), errors.New("sample must be finite or NaN")
	}

	var hp, filt, zero float64
	if s.count >= s.poles {
		if s.poles == 2 {
			difference := sample - 2.0*s.input1 + s.input2
			hp = s.gain*s.gain*difference + 2.0*s.decay*s.hp1 - s.decay*s.decay*s.hp2
		} else {
			hp = s.gain*(sample-s.input1) + s.decay*s.hp1
		}
		filt = s.c1*(hp+s.hp1)/2.0 + s.c2*s.filt1 + s.c3*s.filt2
		if s.zeroMean {
			zero = s.gain*(filt-s.filt1) + s.decay*s.zero1
		}
	}

	s.input2 = s.input1
	s.input1 = sample
	s.hp2 = s.hp1
	s.hp1 = hp
	s.filt2 = s.filt1
	s.filt1 = filt
	s.zero1 = zero
	s.count++

	output := filt
	if s.zeroMean {
		output = zero
	}
	if s.count > s.warmupBars {
		s.currentValue = output
		return output, nil
	}
	return math.NaN(), nil
}

// Update updates the value of the roofing filter given the next sample and panics on infinite input.
func (s *RoofingFilter) Update(sample float64) float64 {
	result, err := s.TryUpdate(sample)
	if err != nil {
		panic(err)
	}
	return result
}
