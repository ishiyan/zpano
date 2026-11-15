package centerofgravityoscillator

//nolint: gofumpt
import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
)

// CenterOfGravityOscillator implements John F. Ehlers' Center of Gravity (CG)
// oscillator from the May 2002 TASC article / Mesa paper (Centered: false) and
// Chapter 5 of "Cybernetic Analysis for Stocks and Futures" (2004, Centered: true).
//
// With newest-first lag i = 0..Length-1:
//
//	Num = Σ((1 + i) * Price[t - i]), Denom = Σ(Price[t - i])
//	CG  = -Num / Denom + offset   (when Denom != 0; otherwise retain previous CG, initially 0)
//
// where offset is 0 when Centered is false and (Length + 1) / 2 when Centered is true.
// The indicator exposes two outputs: the oscillator value (primed after Length finite
// observations) and the one-observation delayed trigger line (primed after Length + 1
// finite observations).
type CenterOfGravityOscillator struct {
	mu              sync.RWMutex
	mnemonic        string
	description     string
	mnemonicTrig    string
	descriptionTrig string
	length          int
	centered        bool
	offset          float64
	window          []float64
	next            int
	count           int
	lastValue       float64
	currentValue    float64
	currentTrigger  float64
	barFunc         entities.BarFunc
	quoteFunc       entities.QuoteFunc
	tradeFunc       entities.TradeFunc
}

// NewCenterOfGravityOscillatorDefault creates a CenterOfGravityOscillator with Ehlers defaults.
func NewCenterOfGravityOscillatorDefault() (*CenterOfGravityOscillator, error) {
	return NewCenterOfGravityOscillator(DefaultParams())
}

// NewCenterOfGravityOscillator returns an instance of the indicator
// created using supplied parameters.
func NewCenterOfGravityOscillator(
	params *Params,
) (*CenterOfGravityOscillator, error) {
	const (
		invalid   = "invalid center of gravity oscillator parameters"
		fmts      = "%s: %s"
		fmtl      = "%s: length should be a positive integer"
		fmtw      = "%s: %w"
		descr     = "Center of Gravity oscillator "
		descrTrig = "Center of Gravity trigger "
	)

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if params == nil {
		return nil, fmt.Errorf(fmts, invalid, "params must not be nil")
	}

	if params.Length < 1 {
		return nil, fmt.Errorf(fmtl, invalid)
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

	if barFunc, err = entities.BarComponentFunc(barComp); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(quoteComp); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tradeComp); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	offset := 0.0
	if params.Centered {
		offset = float64(params.Length+1) / 2.0
	}

	mnemonic := params.Mnemonic()
	mnemonicTrig := params.TriggerMnemonic()

	instance := &CenterOfGravityOscillator{
		mnemonic:        mnemonic,
		description:     descr + mnemonic,
		mnemonicTrig:    mnemonicTrig,
		descriptionTrig: descrTrig + mnemonicTrig,
		length:          params.Length,
		centered:        params.Centered,
		offset:          offset,
		window:          make([]float64, params.Length),
		barFunc:         barFunc,
		quoteFunc:       quoteFunc,
		tradeFunc:       tradeFunc,
	}
	instance.Reset()

	return instance, nil
}

// Reset restores startup state while preserving configuration.
func (s *CenterOfGravityOscillator) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	clear(s.window)
	s.next = 0
	s.count = 0
	s.lastValue = 0.0
	s.currentValue = math.NaN()
	s.currentTrigger = math.NaN()
}

// Length returns the configured observation window length.
func (s *CenterOfGravityOscillator) Length() int {
	return s.length
}

// Centered reports whether the zero-centered book baseline is enabled.
func (s *CenterOfGravityOscillator) Centered() bool {
	return s.centered
}

// Value returns the latest accepted oscillator value (NaN before priming).
func (s *CenterOfGravityOscillator) Value() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentValue
}

// Trigger returns the latest accepted trigger value (NaN until observation Length+1).
func (s *CenterOfGravityOscillator) Trigger() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentTrigger
}

// IsPrimed indicates whether at least Length finite observations have been seen.
func (s *CenterOfGravityOscillator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count >= s.length
}

// IsTriggerPrimed indicates whether a preceding full-window oscillator value exists.
func (s *CenterOfGravityOscillator) IsTriggerPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count > s.length
}

// Metadata describes an output data of the indicator.
func (s *CenterOfGravityOscillator) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CenterOfGravityOscillator,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicTrig, Description: s.descriptionTrig},
		},
	)
}

func (s *CenterOfGravityOscillator) tryUpdateLocked(sample float64) (float64, error) {
	if math.IsInf(sample, 0) {
		return math.NaN(), errors.New("price must be finite or NaN")
	}
	if math.IsNaN(sample) {
		return math.NaN(), nil
	}

	trigger := math.NaN()
	if s.count >= s.length {
		trigger = s.lastValue
	}

	s.window[s.next] = sample
	s.next = (s.next + 1) % s.length
	if s.count <= s.length {
		s.count++
	}
	if s.count < s.length {
		return math.NaN(), nil
	}

	numerator := 0.0
	denominator := 0.0
	for lag := 0; lag < s.length; lag++ {
		index := (s.next + s.length - 1 - lag) % s.length
		price := s.window[index]
		numerator += float64(1+lag) * price
		denominator += price
	}

	if denominator != 0.0 {
		s.lastValue = -numerator/denominator + s.offset
	}
	s.currentValue = s.lastValue
	s.currentTrigger = trigger
	return s.currentValue, nil
}

// TryUpdate consumes one scalar sample and returns an error if sample is infinite.
func (s *CenterOfGravityOscillator) TryUpdate(sample float64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.tryUpdateLocked(sample)
}

// Update updates the value of the center of gravity oscillator given the next sample
// and panics on infinite input.
func (s *CenterOfGravityOscillator) Update(sample float64) float64 {
	result, err := s.TryUpdate(sample)
	if err != nil {
		panic(err)
	}
	return result
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CenterOfGravityOscillator) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CenterOfGravityOscillator) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CenterOfGravityOscillator) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CenterOfGravityOscillator) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *CenterOfGravityOscillator) updateEntity(
	stamp time.Time, sample float64,
) core.Output {
	const length = 2

	s.mu.Lock()
	defer s.mu.Unlock()

	cog, err := s.tryUpdateLocked(sample)
	if err != nil {
		panic(err)
	}
	trig := math.NaN()
	if !math.IsNaN(cog) {
		trig = s.currentTrigger
	}

	output := make(core.Output, length)
	output[0] = entities.Scalar{Time: stamp, Value: cog}
	output[1] = entities.Scalar{Time: stamp, Value: trig}

	return output
}
