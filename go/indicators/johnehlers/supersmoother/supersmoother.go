package supersmoother

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// SuperSmoother implements Ehlers' two-pole SuperSmoother low-pass filter with a zero at the two-bar Nyquist period.
type SuperSmoother struct {
	mu sync.RWMutex
	core.LineIndicator

	warmupBars int
	c1         float64
	c2         float64
	c3         float64

	count        int
	input1       float64
	output1      float64
	output2      float64
	currentValue float64
}

func smootherCoefficients(period float64) (float64, float64, float64) {
	smootherAngle := math.Sqrt2 * math.Pi / period
	a1 := math.Exp(-smootherAngle)
	c2 := 2.0 * a1 * math.Cos(smootherAngle)
	c3 := -a1 * a1
	return 1.0 - c2 - c3, c2, c3
}

// NewSuperSmootherDefault creates a SuperSmoother with Ehlers' default parameters.
func NewSuperSmootherDefault() (*SuperSmoother, error) {
	return NewSuperSmoother(DefaultParams())
}

// NewSuperSmoother returns an instance of the indicator created using supplied parameters.
func NewSuperSmoother(params *Params) (*SuperSmoother, error) {
	const (
		invalid = "invalid super smoother parameters"
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

	c1, c2, c3 := smootherCoefficients(shortest)
	mnemonic := params.Mnemonic()
	description := "Super Smoother " + mnemonic

	filter := &SuperSmoother{
		warmupBars:   2,
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
func (s *SuperSmoother) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.count = 0
	s.input1 = 0
	s.output1 = 0
	s.output2 = 0
	s.currentValue = math.NaN()
}

// WarmupBars returns the number of initial NaN outputs before priming.
func (s *SuperSmoother) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.warmupBars
}

// Value returns the latest calculated scalar value (NaN before priming).
func (s *SuperSmoother) Value() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentValue
}

// IsPrimed indicates whether the indicator has produced at least one primed output.
func (s *SuperSmoother) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count > s.warmupBars
}

// Metadata describes the output data of the indicator.
func (s *SuperSmoother) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.SuperSmoother,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// TryUpdate consumes one scalar sample and returns an error if sample is infinite.
func (s *SuperSmoother) TryUpdate(sample float64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsNaN(sample) {
		return math.NaN(), nil
	}
	if math.IsInf(sample, 0) {
		return math.NaN(), errors.New("sample must be finite or NaN")
	}

	if s.count == 0 {
		s.input1 = sample
		s.output1 = sample
		s.output2 = sample
	}

	output := s.c1*(sample+s.input1)/2.0 + s.c2*s.output1 + s.c3*s.output2
	s.input1 = sample
	s.output2 = s.output1
	s.output1 = output
	s.count++

	if s.count > s.warmupBars {
		s.currentValue = output
		return output, nil
	}
	return math.NaN(), nil
}

// Update updates the value of the super smoother given the next sample and panics on infinite input.
func (s *SuperSmoother) Update(sample float64) float64 {
	result, err := s.TryUpdate(sample)
	if err != nil {
		panic(err)
	}
	return result
}
