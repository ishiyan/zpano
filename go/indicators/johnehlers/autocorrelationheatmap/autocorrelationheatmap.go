// Package autocorrelationheatmap implements John F. Ehlers' February 2025
// Autocorrelation Heatmap indicator ("Drunkard's Walk: Theory and Measurement
// by Autocorrelation", TASC Feb 2025).
package autocorrelationheatmap

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// AutocorrelationState holds an immutable diagnostic snapshot of committed internal state.
type AutocorrelationState struct {
	AcceptedSamples int
	Filtered        float64
	Correlations    []float64
}

// AutoCorrelationHeatmap implements Ehlers' February 2025 Autocorrelation Heatmap (ACH).
type AutoCorrelationHeatmap struct {
	mu            sync.RWMutex
	params        Params
	mnemonic      string
	description   string
	historyLength int
	primeCount    int
	coeffC1       float64
	coeffC2       float64
	coeffC3       float64
	samplesSeen   int
	price1        float64
	price2        float64
	filtered1     float64
	filtered2     float64
	history       []float64
	correlations  []float64
	lastOutput    *outputs.Heatmap
	barFunc       entities.BarFunc
	quoteFunc     entities.QuoteFunc
	tradeFunc     entities.TradeFunc
}

// NewAutoCorrelationHeatmapDefault returns an instance created with default parameters.
func NewAutoCorrelationHeatmapDefault() (*AutoCorrelationHeatmap, error) {
	return NewAutoCorrelationHeatmapParams(DefaultParams())
}

// NewAutoCorrelationHeatmap aliases [NewAutoCorrelationHeatmapParams].
func NewAutoCorrelationHeatmap(p *Params) (*AutoCorrelationHeatmap, error) {
	return NewAutoCorrelationHeatmapParams(p)
}

// NewAutoCorrelationHeatmapParams returns an instance created with the supplied parameters.
func NewAutoCorrelationHeatmapParams(p *Params) (*AutoCorrelationHeatmap, error) {
	const invalid = "invalid autocorrelation heatmap parameters"

	source := DefaultParams()
	if p != nil {
		copied := *p
		source = &copied
	}

	if source.Length < 2 {
		return nil, fmt.Errorf("%s: length must be >= 2", invalid)
	}
	if math.IsNaN(source.SmoothingPeriod) || math.IsInf(source.SmoothingPeriod, 0) || source.SmoothingPeriod < 1.0 {
		return nil, fmt.Errorf("%s: smoothingPeriod must be a finite real >= 1", invalid)
	}
	if source.MaxLag < 0 || source.MaxLag > 99 {
		return nil, fmt.Errorf("%s: maxLag must be in [0, 99]", invalid)
	}

	angle := math.Sqrt2 * math.Pi / source.SmoothingPeriod
	attenuation := math.Exp(-angle)
	coeffC2 := 2.0 * attenuation * math.Cos(angle)
	coeffC3 := -attenuation * attenuation
	coeffC1 := (1.0 + coeffC2 - coeffC3) / 4.0
	if !isFiniteFloat(coeffC1) || !isFiniteFloat(coeffC2) || !isFiniteFloat(coeffC3) {
		return nil, fmt.Errorf("%s: smoothingPeriod produces non-finite filter coefficients", invalid)
	}

	barComp, quoteComp, tradeComp := source.BarComponent, source.QuoteComponent, source.TradeComponent
	if barComp == 0 {
		barComp = entities.DefaultBarComponent
	}
	if quoteComp == 0 {
		quoteComp = entities.DefaultQuoteComponent
	}
	if tradeComp == 0 {
		tradeComp = entities.DefaultTradeComponent
	}

	barFunc, err := entities.BarComponentFunc(barComp)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	quoteFunc, err := entities.QuoteComponentFunc(quoteComp)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	tradeFunc, err := entities.TradeComponentFunc(tradeComp)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}

	historyLength := source.Length + source.MaxLag
	primeCount := 4
	if historyLength > primeCount {
		primeCount = historyLength
	}

	mnemonicValue := source.Mnemonic()
	instance := &AutoCorrelationHeatmap{
		params:        *source,
		mnemonic:      mnemonicValue,
		description:   "Autocorrelation heatmap " + mnemonicValue,
		historyLength: historyLength,
		primeCount:    primeCount,
		coeffC1:       coeffC1,
		coeffC2:       coeffC2,
		coeffC3:       coeffC3,
		barFunc:       barFunc,
		quoteFunc:     quoteFunc,
		tradeFunc:     tradeFunc,
	}
	instance.resetLocked()
	return instance, nil
}

func isFiniteFloat(number float64) bool {
	return !math.IsNaN(number) && !math.IsInf(number, 0)
}

func (s *AutoCorrelationHeatmap) emptyHeatmap(t time.Time) *outputs.Heatmap {
	return outputs.NewEmptyHeatmap(t, 0.0, float64(s.params.MaxLag), 1.0)
}

func (s *AutoCorrelationHeatmap) resetLocked() {
	s.samplesSeen = 0
	s.price1 = 0
	s.price2 = 0
	s.filtered1 = 0
	s.filtered2 = 0
	s.history = make([]float64, s.historyLength)
	s.correlations = make([]float64, s.params.MaxLag+1)
	s.lastOutput = s.emptyHeatmap(time.Time{})
}

// Reset restores explicit initial conditions while preserving configuration.
func (s *AutoCorrelationHeatmap) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// Parameters returns a copy of the indicator parameters.
func (s *AutoCorrelationHeatmap) Parameters() Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// Length returns the Pearson correlation sample pair count.
func (s *AutoCorrelationHeatmap) Length() int {
	return s.params.Length
}

// SmoothingPeriod returns the Ultimate Smoother critical period.
func (s *AutoCorrelationHeatmap) SmoothingPeriod() float64 {
	return s.params.SmoothingPeriod
}

// MaxLag returns the inclusive upper correlation lag.
func (s *AutoCorrelationHeatmap) MaxLag() int {
	return s.params.MaxLag
}

// WarmupBars returns the number of leading empty columns: max(4, Length+MaxLag) - 1.
func (s *AutoCorrelationHeatmap) WarmupBars() int {
	return s.primeCount - 1
}

// SamplesSeen returns the number of accepted finite samples processed since creation or reset.
func (s *AutoCorrelationHeatmap) SamplesSeen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.samplesSeen
}

// IsPrimed indicates whether sufficient history has accumulated for all lags.
func (s *AutoCorrelationHeatmap) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.samplesSeen >= s.primeCount
}

// State returns an immutable diagnostic snapshot of committed internal filter and correlation state.
func (s *AutoCorrelationHeatmap) State() AutocorrelationState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return AutocorrelationState{
		AcceptedSamples: s.samplesSeen,
		Filtered:        s.filtered1,
		Correlations:    slices.Clone(s.correlations),
	}
}

// Output returns a copy of the latest committed heatmap column.
func (s *AutoCorrelationHeatmap) Output() *outputs.Heatmap {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.lastOutput.IsEmpty() {
		return outputs.NewEmptyHeatmap(
			s.lastOutput.Time,
			s.lastOutput.ParameterFirst,
			s.lastOutput.ParameterLast,
			s.lastOutput.ParameterResolution,
		)
	}
	return outputs.NewHeatmap(
		s.lastOutput.Time,
		s.lastOutput.ParameterFirst,
		s.lastOutput.ParameterLast,
		s.lastOutput.ParameterResolution,
		s.lastOutput.ValueMin,
		s.lastOutput.ValueMax,
		slices.Clone(s.lastOutput.Values),
	)
}

// Metadata describes the output data of the indicator.
func (s *AutoCorrelationHeatmap) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AutoCorrelationHeatmap,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
		},
	)
}

// TryUpdate feeds the next scalar sample and returns the heatmap column or an error on non-finite/overflowing arithmetic.
func (s *AutoCorrelationHeatmap) TryUpdate(sample float64, t time.Time) (*outputs.Heatmap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsInf(sample, 0) {
		return nil, errors.New("price must be finite or NaN (missing sample)")
	}
	if math.IsNaN(sample) {
		return s.emptyHeatmap(t), nil
	}

	var filtered float64
	if s.samplesSeen < 3 {
		filtered = sample
	} else {
		filtered = (1.0-s.coeffC1)*sample +
			(2.0*s.coeffC1-s.coeffC2)*s.price1 -
			(s.coeffC1+s.coeffC3)*s.price2 +
			s.coeffC2*s.filtered1 +
			s.coeffC3*s.filtered2
	}
	if !isFiniteFloat(filtered) {
		return nil, errors.New("Ultimate Smoother calculation overflowed")
	}

	nextHistory := make([]float64, s.historyLength)
	nextHistory[0] = filtered
	copy(nextHistory[1:], s.history[:s.historyLength-1])

	nextCorrelations := slices.Clone(s.correlations)
	lengthFloat := float64(s.params.Length)
	binCount := s.params.MaxLag + 1

	for lag := range binCount {
		var sumX, sumY, sumXX, sumXY, sumYY float64
		for offset := range s.params.Length {
			sampleX := nextHistory[offset]
			sampleY := nextHistory[lag+offset]
			sumX += sampleX
			sumY += sampleY
			sumXX += sampleX * sampleX
			sumXY += sampleX * sampleY
			sumYY += sampleY * sampleY
		}
		varX := lengthFloat*sumXX - sumX*sumX
		varY := lengthFloat*sumYY - sumY*sumY
		if !isFiniteFloat(sumX) || !isFiniteFloat(sumY) ||
			!isFiniteFloat(sumXX) || !isFiniteFloat(sumXY) || !isFiniteFloat(sumYY) ||
			!isFiniteFloat(varX) || !isFiniteFloat(varY) {
			return nil, errors.New("Pearson sums overflowed; reduce input magnitude")
		}
		if varX > 0.0 && varY > 0.0 {
			product := varX * varY
			numerator := lengthFloat*sumXY - sumX*sumY
			if product <= 0.0 || !isFiniteFloat(product) || !isFiniteFloat(numerator) {
				return nil, errors.New("Pearson denominator or numerator overflowed/underflowed")
			}
			correlation := numerator / math.Sqrt(product)
			if !isFiniteFloat(correlation) {
				return nil, errors.New("Pearson correlation is not finite")
			}
			nextCorrelations[lag] = correlation
		}
	}

	s.price2 = s.price1
	s.price1 = sample
	s.filtered2 = s.filtered1
	s.filtered1 = filtered
	s.history = nextHistory
	s.correlations = nextCorrelations
	s.samplesSeen++

	if s.samplesSeen < s.primeCount {
		empty := s.emptyHeatmap(t)
		s.lastOutput = empty
		return s.emptyHeatmap(t), nil
	}

	valueMin := nextCorrelations[0]
	valueMax := nextCorrelations[0]
	for _, item := range nextCorrelations[1:] {
		if item < valueMin {
			valueMin = item
		}
		if item > valueMax {
			valueMax = item
		}
	}

	maxLagFloat := float64(s.params.MaxLag)
	s.lastOutput = outputs.NewHeatmap(t, 0.0, maxLagFloat, 1.0, valueMin, valueMax, slices.Clone(nextCorrelations))
	return outputs.NewHeatmap(t, 0.0, maxLagFloat, 1.0, valueMin, valueMax, slices.Clone(nextCorrelations)), nil
}

// Update feeds the next sample and returns the heatmap column, panicking on non-NaN invalid/overflowing input.
func (s *AutoCorrelationHeatmap) Update(sample float64, t time.Time) *outputs.Heatmap {
	result, err := s.TryUpdate(sample, t)
	if err != nil {
		panic(err)
	}
	return result
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *AutoCorrelationHeatmap) UpdateScalar(sample *entities.Scalar) core.Output {
	return core.Output{s.Update(sample.Value, sample.Time)}
}

// UpdateBar updates the indicator given the next bar sample.
func (s *AutoCorrelationHeatmap) UpdateBar(sample *entities.Bar) core.Output {
	return core.Output{s.Update(s.barFunc(sample), sample.Time)}
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *AutoCorrelationHeatmap) UpdateQuote(sample *entities.Quote) core.Output {
	return core.Output{s.Update(s.quoteFunc(sample), sample.Time)}
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *AutoCorrelationHeatmap) UpdateTrade(sample *entities.Trade) core.Output {
	return core.Output{s.Update(s.tradeFunc(sample), sample.Time)}
}
