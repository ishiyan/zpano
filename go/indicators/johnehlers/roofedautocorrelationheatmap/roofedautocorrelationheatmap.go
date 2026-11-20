// Package roofedautocorrelationheatmap implements John F. Ehlers' 2013
// Roofed Autocorrelation Heatmap indicator ("Cycle Analytics for Traders",
// Wiley 2013, Code Listing 8-2).
package roofedautocorrelationheatmap

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

const (
	twoPi   = 2.0 * math.Pi
	sqrtTwo = math.Sqrt2
)

// AutocorrelationState holds an immutable diagnostic snapshot of committed internal state.
type AutocorrelationState struct {
	HighPass     float64
	Filtered     float64
	Correlations []float64
}

type roofFilter struct {
	hp0 float64
	hp1 float64
	hp2 float64
	ss1 float64
	ss2 float64
	ss3 float64
	x1  float64
	x2  float64
	h1  float64
	h2  float64
	f1  float64
	f2  float64
}

func newRoofFilter(highpassPeriod, smoothingPeriod float64) (roofFilter, error) {
	if !isFiniteFloat(highpassPeriod) || highpassPeriod < 5.0 {
		return roofFilter{}, errors.New("highpassPeriod must be a finite real >= 5.0")
	}
	if !isFiniteFloat(smoothingPeriod) || smoothingPeriod < 2.0 {
		return roofFilter{}, errors.New("smoothingPeriod must be a finite real >= 2.0")
	}

	angle := twoPi / (sqrtTwo * highpassPeriod)
	cosine := math.Cos(angle)
	alpha := (cosine + math.Sin(angle) - 1.0) / cosine
	pole := 1.0 - alpha
	attenuation := math.Exp(-sqrtTwo * math.Pi / smoothingPeriod)
	hp0 := (1.0 - alpha/2.0) * (1.0 - alpha/2.0)
	hp1 := 2.0 * pole
	hp2 := pole * pole
	ss2 := 2.0 * attenuation * math.Cos(sqrtTwo*math.Pi/smoothingPeriod)
	ss3 := -attenuation * attenuation
	ss1 := 1.0 - ss2 - ss3

	if !(pole > 0.0 && pole < 1.0 && attenuation > 0.0 && attenuation < 1.0 && ss1 > 0.0) ||
		!isFiniteFloat(hp0) || !isFiniteFloat(hp1) || !isFiniteFloat(hp2) ||
		!isFiniteFloat(ss1) || !isFiniteFloat(ss2) || !isFiniteFloat(ss3) {
		return roofFilter{}, errors.New("periods must produce finite resolvable stable coefficients")
	}

	return roofFilter{
		hp0: hp0,
		hp1: hp1,
		hp2: hp2,
		ss1: ss1,
		ss2: ss2,
		ss3: ss3,
	}, nil
}

func (r *roofFilter) reset() {
	r.x1 = 0.0
	r.x2 = 0.0
	r.h1 = 0.0
	r.h2 = 0.0
	r.f1 = 0.0
	r.f2 = 0.0
}

func (r *roofFilter) candidate(price float64) (float64, float64, error) {
	high := float64(float64(r.hp0*(price-float64(2.0*r.x1)+r.x2))+float64(r.hp1*r.h1)) - float64(r.hp2*r.h2)
	filtered := float64(float64(r.ss1*(high+r.h1)/2.0)+float64(r.ss2*r.f1)) + float64(r.ss3*r.f2)
	if !isFiniteFloat(high) || !isFiniteFloat(filtered) {
		return 0.0, 0.0, errors.New("roofing arithmetic overflowed")
	}
	return high, filtered, nil
}

func (r *roofFilter) commit(price, high, filtered float64) {
	r.x2 = r.x1
	r.x1 = price
	r.h2 = r.h1
	r.h1 = high
	r.f2 = r.f1
	r.f1 = filtered
}

// RoofedAutoCorrelationHeatmap implements Ehlers' 2013 Roofed Autocorrelation Heatmap (RACH).
type RoofedAutoCorrelationHeatmap struct {
	mu            sync.RWMutex
	params        Params
	mnemonic      string
	description   string
	historyLength int
	roof          roofFilter
	samplesSeen   int
	history       []float64
	correlations  []float64
	lastOutput    *outputs.Heatmap
	barFunc       entities.BarFunc
	quoteFunc     entities.QuoteFunc
	tradeFunc     entities.TradeFunc
}

// NewRoofedAutoCorrelationHeatmapDefault returns an instance created with default parameters.
func NewRoofedAutoCorrelationHeatmapDefault() (*RoofedAutoCorrelationHeatmap, error) {
	return NewRoofedAutoCorrelationHeatmapParams(DefaultParams())
}

// NewRoofedAutoCorrelationHeatmap aliases [NewRoofedAutoCorrelationHeatmapParams].
func NewRoofedAutoCorrelationHeatmap(p *Params) (*RoofedAutoCorrelationHeatmap, error) {
	return NewRoofedAutoCorrelationHeatmapParams(p)
}

// NewRoofedAutoCorrelationHeatmapParams returns an instance created with the supplied parameters.
func NewRoofedAutoCorrelationHeatmapParams(p *Params) (*RoofedAutoCorrelationHeatmap, error) {
	const invalid = "invalid roofed autocorrelation heatmap parameters"

	source := DefaultParams()
	if p != nil {
		copied := *p
		source = &copied
	}

	if source.MinLag < 0 {
		return nil, fmt.Errorf("%s: minLag must be >= 0", invalid)
	}
	if source.MaxLag < source.MinLag {
		return nil, fmt.Errorf("%s: maxLag must be >= minLag", invalid)
	}
	if source.AveragingLength < 0 {
		return nil, fmt.Errorf("%s: averagingLength must be >= 0", invalid)
	}

	roof, err := newRoofFilter(source.HighpassPeriod, source.SmoothingPeriod)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
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

	effectiveWindow := source.AveragingLength
	if effectiveWindow == 0 {
		effectiveWindow = source.MaxLag
	}
	historyLength := source.MaxLag + effectiveWindow
	if historyLength < 1 {
		historyLength = 1
	}

	mnemonicValue := source.Mnemonic()
	instance := &RoofedAutoCorrelationHeatmap{
		params:        *source,
		mnemonic:      mnemonicValue,
		description:   "Roofed autocorrelation heatmap " + mnemonicValue,
		historyLength: historyLength,
		roof:          roof,
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

func (s *RoofedAutoCorrelationHeatmap) emptyHeatmap(t time.Time) *outputs.Heatmap {
	return outputs.NewEmptyHeatmap(t, float64(s.params.MinLag), float64(s.params.MaxLag), 1.0)
}

func (s *RoofedAutoCorrelationHeatmap) resetLocked() {
	s.roof.reset()
	s.samplesSeen = 0
	s.history = make([]float64, s.historyLength)
	s.correlations = make([]float64, s.params.MaxLag+1)
	s.lastOutput = s.emptyHeatmap(time.Time{})
}

// Reset restores explicit initial conditions while preserving configuration.
func (s *RoofedAutoCorrelationHeatmap) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// Parameters returns a copy of the indicator parameters.
func (s *RoofedAutoCorrelationHeatmap) Parameters() Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// MinLag returns the first displayed correlation lag.
func (s *RoofedAutoCorrelationHeatmap) MinLag() int {
	return s.params.MinLag
}

// MaxLag returns the last displayed correlation lag.
func (s *RoofedAutoCorrelationHeatmap) MaxLag() int {
	return s.params.MaxLag
}

// MinPeriod aliases [RoofedAutoCorrelationHeatmap.MinLag] for heatmap axis parity.
func (s *RoofedAutoCorrelationHeatmap) MinPeriod() int {
	return s.params.MinLag
}

// MaxPeriod aliases [RoofedAutoCorrelationHeatmap.MaxLag] for heatmap axis parity.
func (s *RoofedAutoCorrelationHeatmap) MaxPeriod() int {
	return s.params.MaxLag
}

// AveragingLength returns the configured sample-pair window length (0 when M = lag).
func (s *RoofedAutoCorrelationHeatmap) AveragingLength() int {
	return s.params.AveragingLength
}

// HighpassPeriod returns the two-pole roofing high-pass cutoff period.
func (s *RoofedAutoCorrelationHeatmap) HighpassPeriod() float64 {
	return s.params.HighpassPeriod
}

// SmoothingPeriod returns the two-pole SuperSmoother cutoff period.
func (s *RoofedAutoCorrelationHeatmap) SmoothingPeriod() float64 {
	return s.params.SmoothingPeriod
}

// WarmupBars returns the number of accepted samples masked before the first non-empty heatmap column.
func (s *RoofedAutoCorrelationHeatmap) WarmupBars() int {
	return s.historyLength - 1
}

// SamplesSeen returns the count of accepted finite samples processed since creation or reset.
func (s *RoofedAutoCorrelationHeatmap) SamplesSeen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.samplesSeen
}

// IsPrimed reports whether the full lag-correlation window is populated with accepted roofed samples.
func (s *RoofedAutoCorrelationHeatmap) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.samplesSeen >= s.historyLength
}

// State returns an immutable diagnostic snapshot of internal roofing and correlation state.
func (s *RoofedAutoCorrelationHeatmap) State() AutocorrelationState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return AutocorrelationState{
		HighPass:     s.roof.h1,
		Filtered:     s.roof.f1,
		Correlations: slices.Clone(s.correlations),
	}
}

// Output returns a copy of the latest committed heatmap column.
func (s *RoofedAutoCorrelationHeatmap) Output() *outputs.Heatmap {
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
func (s *RoofedAutoCorrelationHeatmap) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.RoofedAutoCorrelationHeatmap,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
		},
	)
}

// TryUpdate feeds the next scalar sample and returns the heatmap column or an error on non-finite/overflowing arithmetic.
func (s *RoofedAutoCorrelationHeatmap) TryUpdate(sample float64, t time.Time) (*outputs.Heatmap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsInf(sample, 0) {
		return nil, errors.New("price must be finite or NaN (missing)")
	}
	if math.IsNaN(sample) {
		return s.emptyHeatmap(t), nil
	}

	high, filtered, err := s.roof.candidate(sample)
	if err != nil {
		return nil, err
	}

	nextHistory := make([]float64, s.historyLength)
	nextHistory[0] = filtered
	copy(nextHistory[1:], s.history[:s.historyLength-1])

	nextCorrelations := slices.Clone(s.correlations)
	binCount := s.params.MaxLag + 1

	for lag := range binCount {
		window := s.params.AveragingLength
		if window == 0 {
			window = lag
		}
		windowFloat := float64(window)

		var sumX, sumY, sumXX, sumXY, sumYY float64
		for offset := range window {
			sampleX := nextHistory[offset]
			sampleY := nextHistory[lag+offset]
			sumX += sampleX
			sumY += sampleY
			sumXX += float64(sampleX * sampleX)
			sumXY += float64(sampleX * sampleY)
			sumYY += float64(sampleY * sampleY)
		}

		varX := float64(windowFloat*sumXX) - float64(sumX*sumX)
		varY := float64(windowFloat*sumYY) - float64(sumY*sumY)
		denominator := float64(varX * varY)
		if !isFiniteFloat(sumX) || !isFiniteFloat(sumY) ||
			!isFiniteFloat(sumXX) || !isFiniteFloat(sumXY) || !isFiniteFloat(sumYY) ||
			!isFiniteFloat(denominator) {
			return nil, errors.New("Pearson arithmetic overflowed")
		}
		if denominator > 0.0 {
			numerator := float64(windowFloat*sumXY) - float64(sumX*sumY)
			nextCorrelations[lag] = numerator / math.Sqrt(denominator)
		}
		nextCorrelations[lag] = float64(0.5 * float64(nextCorrelations[lag]+1.0))
		if !isFiniteFloat(nextCorrelations[lag]) {
			return nil, errors.New("correlation arithmetic overflowed")
		}
	}

	s.roof.commit(sample, high, filtered)
	s.history = nextHistory
	s.correlations = nextCorrelations
	s.samplesSeen++

	if s.samplesSeen < s.historyLength {
		s.lastOutput = s.emptyHeatmap(t)
		return s.emptyHeatmap(t), nil
	}

	displayed := slices.Clone(nextCorrelations[s.params.MinLag : s.params.MaxLag+1])
	valueMin := displayed[0]
	valueMax := displayed[0]
	for _, item := range displayed[1:] {
		if item < valueMin {
			valueMin = item
		}
		if item > valueMax {
			valueMax = item
		}
	}

	minLagFloat := float64(s.params.MinLag)
	maxLagFloat := float64(s.params.MaxLag)
	s.lastOutput = outputs.NewHeatmap(t, minLagFloat, maxLagFloat, 1.0, valueMin, valueMax, slices.Clone(displayed))
	return outputs.NewHeatmap(t, minLagFloat, maxLagFloat, 1.0, valueMin, valueMax, displayed), nil
}

// Update feeds the next sample and returns the heatmap column, panicking on non-NaN invalid/overflowing input.
func (s *RoofedAutoCorrelationHeatmap) Update(sample float64, t time.Time) *outputs.Heatmap {
	result, err := s.TryUpdate(sample, t)
	if err != nil {
		panic(err)
	}
	return result
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *RoofedAutoCorrelationHeatmap) UpdateScalar(sample *entities.Scalar) core.Output {
	return core.Output{s.Update(sample.Value, sample.Time)}
}

// UpdateBar updates the indicator given the next bar sample.
func (s *RoofedAutoCorrelationHeatmap) UpdateBar(sample *entities.Bar) core.Output {
	return core.Output{s.Update(s.barFunc(sample), sample.Time)}
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *RoofedAutoCorrelationHeatmap) UpdateQuote(sample *entities.Quote) core.Output {
	return core.Output{s.Update(s.quoteFunc(sample), sample.Time)}
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *RoofedAutoCorrelationHeatmap) UpdateTrade(sample *entities.Trade) core.Output {
	return core.Output{s.Update(s.tradeFunc(sample), sample.Time)}
}
