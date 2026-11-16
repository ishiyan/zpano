// Package autocorrelationperiodogram implements John F. Ehlers' Autocorrelation
// Periodogram with dominant cycle estimation.
//
// Default profile ("article") implements the EasyLanguage sidebar of "Measuring
// Market Cycles", TASC September 2016. Profile "book" implements the corrected
// Listing 8-3 from "Cycle Analytics for Traders" (2013), Chapter 8.
package autocorrelationperiodogram

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
	twoPi            = 2.0 * math.Pi
	sqrtTwo          = math.Sqrt2
	inverseSqrtTwo   = 1.0 / sqrtTwo
	dftLagStart      = 3
	powerSmoothing   = 0.2
	powerRetention   = 0.8
	bookAgcDecay     = 0.995
	articleThreshold = 0.25
	bookThreshold    = 0.5
)

// PeriodogramOutput holds one heatmap column and the real-valued dominant cycle period in bars.
type PeriodogramOutput struct {
	Heatmap       *outputs.Heatmap
	DominantCycle float64
}

// PeriodogramState holds a read-only snapshot of the committed internal recursive state.
type PeriodogramState struct {
	HighPass              float64
	Filtered              float64
	Correlations          []float64
	SmoothedPower         []float64
	NormalizationMaximum  float64
	InternalDominantCycle float64
}

// AutoCorrelationPeriodogram implements Ehlers' Autocorrelation Periodogram with dominant cycle.
type AutoCorrelationPeriodogram struct {
	mu            sync.RWMutex
	params        Params
	mnemonic      string
	description   string
	mnemonicDC    string
	descriptionDC string
	hp0           float64
	hp1           float64
	hp2           float64
	ss1           float64
	ss2           float64
	ss3           float64
	historyLength int
	binCount      int
	cosines       [][]float64
	sines         [][]float64
	samplesSeen   int
	close1        float64
	close2        float64
	high1         float64
	high2         float64
	filtered      []float64
	correlations  []float64
	power         []float64
	maximum       float64
	dominant      float64
	hasDominant   bool
	lastOutput    PeriodogramOutput
	barFunc       entities.BarFunc
	quoteFunc     entities.QuoteFunc
	tradeFunc     entities.TradeFunc
}

// NewAutoCorrelationPeriodogramDefault returns an instance created with default parameters.
func NewAutoCorrelationPeriodogramDefault() (*AutoCorrelationPeriodogram, error) {
	return NewAutoCorrelationPeriodogramParams(DefaultParams())
}

// NewAutoCorrelationPeriodogram aliases [NewAutoCorrelationPeriodogramParams].
func NewAutoCorrelationPeriodogram(p *Params) (*AutoCorrelationPeriodogram, error) {
	return NewAutoCorrelationPeriodogramParams(p)
}

// NewAutoCorrelationPeriodogramParams returns an instance created with the supplied parameters.
func NewAutoCorrelationPeriodogramParams(p *Params) (*AutoCorrelationPeriodogram, error) {
	const invalid = "invalid autocorrelation periodogram parameters"

	cfg := DefaultParams()
	if p != nil {
		copied := *p
		cfg = &copied
	}

	if !cfg.Source.IsKnown() {
		return nil, fmt.Errorf("%s: source must be 'article' or 'book'", invalid)
	}
	if cfg.MinPeriod != 0 && cfg.MinPeriod < 3 {
		return nil, fmt.Errorf("%s: minPeriod must be >= 3", invalid)
	}
	resolvedMin := resolveMinPeriod(cfg.Source, cfg.MinPeriod)
	if resolvedMin < 3 {
		return nil, fmt.Errorf("%s: minPeriod must be >= 3", invalid)
	}
	if cfg.MaxPeriod < 5 || cfg.MaxPeriod <= resolvedMin {
		return nil, fmt.Errorf("%s: maxPeriod must be >= 5 and > minPeriod", invalid)
	}
	if cfg.AveragingLength < 0 || cfg.AveragingLength == 1 {
		return nil, fmt.Errorf("%s: averagingLength must be 0 or >= 2", invalid)
	}

	var pole, hp0, hp1, hp2 float64
	if cfg.Source == Article {
		theta := twoPi / float64(cfg.MaxPeriod)
		cosine := math.Cos(theta)
		pole = (1.0 - math.Sin(theta)) / cosine
		hp0 = 0.5 * (1.0 + pole)
		hp1 = pole
		hp2 = 0.0
	} else {
		theta := inverseSqrtTwo * twoPi / float64(cfg.MaxPeriod)
		cosine := math.Cos(theta)
		alpha := (cosine + math.Sin(theta) - 1.0) / cosine
		pole = 1.0 - alpha
		hp0 = (1.0 - alpha/2.0) * (1.0 - alpha/2.0)
		hp1 = 2.0 * pole
		hp2 = pole * pole
	}

	a := math.Exp(-sqrtTwo * math.Pi / float64(resolvedMin))
	ss2 := 2.0 * a * math.Cos(sqrtTwo*math.Pi/float64(resolvedMin))
	ss3 := -a * a
	ss1 := 1.0 - ss2 - ss3
	if !(pole > 0.0 && pole < 1.0 && a > 0.0 && a < 1.0 && ss1 > 0.0) ||
		math.IsNaN(hp0) || math.IsInf(hp0, 0) || math.IsNaN(hp1) || math.IsInf(hp1, 0) ||
		math.IsNaN(hp2) || math.IsInf(hp2, 0) || math.IsNaN(ss1) || math.IsInf(ss1, 0) ||
		math.IsNaN(ss2) || math.IsInf(ss2, 0) || math.IsNaN(ss3) || math.IsInf(ss3, 0) {
		return nil, fmt.Errorf("%s: periods must yield finite stable coefficients", invalid)
	}

	barComp, quoteComp, tradeComp := cfg.BarComponent, cfg.QuoteComponent, cfg.TradeComponent
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

	cfg.MinPeriod = resolvedMin
	effectiveMaxWindow := cfg.AveragingLength
	if effectiveMaxWindow == 0 {
		effectiveMaxWindow = cfg.MaxPeriod
	}
	historyLength := max(3, cfg.MaxPeriod+effectiveMaxWindow)
	binCount := cfg.MaxPeriod - cfg.MinPeriod + 1
	lagSpan := cfg.MaxPeriod - dftLagStart + 1

	cosines := make([][]float64, binCount)
	sines := make([][]float64, binCount)
	for binIndex := range binCount {
		period := cfg.MinPeriod + binIndex
		cosRow := make([]float64, lagSpan)
		sinRow := make([]float64, lagSpan)
		for offset := range lagSpan {
			lag := dftLagStart + offset
			angle := twoPi * float64(lag) / float64(period)
			cosRow[offset] = math.Cos(angle)
			sinRow[offset] = math.Sin(angle)
		}
		cosines[binIndex] = cosRow
		sines[binIndex] = sinRow
	}

	mnemonicValue := cfg.Mnemonic()
	mnemonicDC := cfg.DominantCycleMnemonic()

	instance := &AutoCorrelationPeriodogram{
		params:        *cfg,
		mnemonic:      mnemonicValue,
		description:   "Autocorrelation periodogram " + mnemonicValue,
		mnemonicDC:    mnemonicDC,
		descriptionDC: "Autocorrelation periodogram dominant cycle " + mnemonicDC,
		hp0:           hp0,
		hp1:           hp1,
		hp2:           hp2,
		ss1:           ss1,
		ss2:           ss2,
		ss3:           ss3,
		historyLength: historyLength,
		binCount:      binCount,
		cosines:       cosines,
		sines:         sines,
		barFunc:       barFunc,
		quoteFunc:     quoteFunc,
		tradeFunc:     tradeFunc,
	}
	instance.resetLocked()
	return instance, nil
}

func (s *AutoCorrelationPeriodogram) resetLocked() {
	s.samplesSeen = 0
	s.close1 = 0
	s.close2 = 0
	s.high1 = 0
	s.high2 = 0
	s.filtered = make([]float64, s.historyLength)
	s.correlations = make([]float64, s.params.MaxPeriod+1)
	s.power = make([]float64, s.binCount)
	s.maximum = 0
	s.dominant = 0
	s.hasDominant = false
	s.lastOutput = PeriodogramOutput{
		Heatmap:       outputs.NewEmptyHeatmap(time.Time{}, float64(s.params.MinPeriod), float64(s.params.MaxPeriod), 1.0),
		DominantCycle: math.NaN(),
	}
}

// Reset restores initial state while preserving configuration.
func (s *AutoCorrelationPeriodogram) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// Parameters returns a copy of the resolved indicator parameters.
func (s *AutoCorrelationPeriodogram) Parameters() Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// WarmupBars returns the number of leading unavailable samples (historyLength - 1).
func (s *AutoCorrelationPeriodogram) WarmupBars() int {
	return s.historyLength - 1
}

// SamplesSeen returns the number of accepted finite samples processed so far.
func (s *AutoCorrelationPeriodogram) SamplesSeen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.samplesSeen
}

// IsPrimed reports whether the full Pearson filtered history is populated.
func (s *AutoCorrelationPeriodogram) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.samplesSeen >= s.historyLength
}

// State returns a snapshot of the committed internal recursive state.
func (s *AutoCorrelationPeriodogram) State() PeriodogramState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return PeriodogramState{
		HighPass:              s.high1,
		Filtered:              s.filtered[0],
		Correlations:          slices.Clone(s.correlations),
		SmoothedPower:         slices.Clone(s.power),
		NormalizationMaximum:  s.maximum,
		InternalDominantCycle: s.dominant,
	}
}

// Output returns the latest committed PeriodogramOutput.
func (s *AutoCorrelationPeriodogram) Output() PeriodogramOutput {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.lastOutput.Heatmap == nil || s.lastOutput.Heatmap.IsEmpty() {
		return s.lastOutput
	}
	return PeriodogramOutput{
		Heatmap: outputs.NewHeatmap(
			s.lastOutput.Heatmap.Time,
			s.lastOutput.Heatmap.ParameterFirst,
			s.lastOutput.Heatmap.ParameterLast,
			s.lastOutput.Heatmap.ParameterResolution,
			s.lastOutput.Heatmap.ValueMin,
			s.lastOutput.Heatmap.ValueMax,
			slices.Clone(s.lastOutput.Heatmap.Values),
		),
		DominantCycle: s.lastOutput.DominantCycle,
	}
}

// Metadata describes the output data of the indicator.
func (s *AutoCorrelationPeriodogram) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AutoCorrelationPeriodogram,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicDC, Description: s.descriptionDC},
		},
	)
}

func isFiniteFloat(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// TryUpdate feeds the next scalar sample and returns PeriodogramOutput or an error on invalid/overflowing input.
func (s *AutoCorrelationPeriodogram) TryUpdate(sample float64, t time.Time) (PeriodogramOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsNaN(sample) {
		return PeriodogramOutput{
			Heatmap:       outputs.NewEmptyHeatmap(t, float64(s.params.MinPeriod), float64(s.params.MaxPeriod), 1.0),
			DominantCycle: math.NaN(),
		}, nil
	}
	if math.IsInf(sample, 0) {
		return PeriodogramOutput{}, errors.New("sample must be finite or NaN (skipped)")
	}

	var high float64
	if s.params.Source == Article {
		high = s.hp0*(sample-s.close1) + s.hp1*s.high1
	} else {
		high = s.hp0*(sample-2.0*s.close1+s.close2) + s.hp1*s.high1 - s.hp2*s.high2
	}
	filteredValue := s.ss1*(high+s.high1)/2.0 + s.ss2*s.filtered[0] + s.ss3*s.filtered[1]
	if !isFiniteFloat(high) || !isFiniteFloat(filteredValue) {
		return PeriodogramOutput{}, errors.New("filter arithmetic overflowed")
	}

	nextFiltered := make([]float64, s.historyLength)
	nextFiltered[0] = filteredValue
	copy(nextFiltered[1:], s.filtered[:s.historyLength-1])

	nextCorrelations := slices.Clone(s.correlations)
	for lag := 0; lag <= s.params.MaxPeriod; lag++ {
		windowLen := s.params.AveragingLength
		if windowLen == 0 {
			windowLen = lag
		}
		var sx, sy, sxx, syy, sxy float64
		for count := range windowLen {
			x := nextFiltered[count]
			y := nextFiltered[lag+count]
			sx += x
			sy += y
			sxx += x * x
			sxy += x * y
			syy += y * y
		}
		windowFloat := float64(windowLen)
		denominator := (windowFloat*sxx - sx*sx) * (windowFloat*syy - sy*sy)
		if !isFiniteFloat(sx) || !isFiniteFloat(sy) || !isFiniteFloat(sxx) ||
			!isFiniteFloat(syy) || !isFiniteFloat(sxy) || !isFiniteFloat(denominator) {
			return PeriodogramOutput{}, errors.New("Pearson arithmetic overflowed")
		}
		if denominator > 0.0 {
			nextCorrelations[lag] = (windowFloat*sxy - sx*sy) / math.Sqrt(denominator)
		}
		if !isFiniteFloat(nextCorrelations[lag]) {
			return PeriodogramOutput{}, errors.New("Pearson correlation overflowed")
		}
	}

	nextPower := make([]float64, s.binCount)
	lagSpan := s.params.MaxPeriod - dftLagStart + 1
	for binIndex := range s.binCount {
		cosRow := s.cosines[binIndex]
		sinRow := s.sines[binIndex]
		var cosinePart, sinePart float64
		for offset := range lagSpan {
			corrValue := nextCorrelations[dftLagStart+offset]
			cosinePart += corrValue * cosRow[offset]
			sinePart += corrValue * sinRow[offset]
		}
		squaredSum := cosinePart*cosinePart + sinePart*sinePart
		smoothed := powerSmoothing*squaredSum*squaredSum + powerRetention*s.power[binIndex]
		if !isFiniteFloat(smoothed) {
			return PeriodogramOutput{}, errors.New("spectral arithmetic overflowed")
		}
		nextPower[binIndex] = smoothed
	}

	maximum := 0.0
	if s.params.Source == Book {
		maximum = bookAgcDecay * s.maximum
	}
	for _, powerValue := range nextPower {
		if powerValue > maximum {
			maximum = powerValue
		}
	}

	normalized := make([]float64, s.binCount)
	if maximum > 0.0 {
		for binIndex, powerValue := range nextPower {
			normalized[binIndex] = powerValue / maximum
		}
	}

	weights := normalized
	if s.params.Source == Article && s.params.EnhanceResolution {
		weights = make([]float64, s.binCount)
		for binIndex, normValue := range normalized {
			weights[binIndex] = normValue * normValue * normValue
		}
	}

	threshold := articleThreshold
	if s.params.Source == Book {
		threshold = bookThreshold
	}
	peak := slices.Max(weights)
	var sp, spx float64
	for binIndex, weight := range weights {
		if weight >= threshold && (s.params.Source == Book || peak >= articleThreshold) {
			spx += float64(s.params.MinPeriod+binIndex) * weight
			sp += weight
		}
	}

	dominant := s.dominant
	hasDominant := s.hasDominant
	if sp != 0.0 && (s.params.Source == Book || sp >= articleThreshold) {
		dominant = spx / sp
		hasDominant = true
	}
	if !isFiniteFloat(dominant) {
		return PeriodogramOutput{}, errors.New("dominant-cycle arithmetic overflowed")
	}

	values := slices.Clone(weights)
	if s.params.Source == Book && s.params.EnhanceResolution {
		for binIndex, normValue := range normalized {
			values[binIndex] = normValue * normValue
		}
	}

	s.close2, s.close1 = s.close1, sample
	s.high2, s.high1 = s.high1, high
	s.filtered = nextFiltered
	s.correlations = nextCorrelations
	s.power = nextPower
	s.maximum = maximum
	s.dominant = dominant
	s.hasDominant = hasDominant
	s.samplesSeen++

	var emitted PeriodogramOutput
	if s.samplesSeen >= s.historyLength {
		valueMin := values[0]
		valueMax := values[0]
		for _, binValue := range values[1:] {
			if binValue < valueMin {
				valueMin = binValue
			}
			if binValue > valueMax {
				valueMax = binValue
			}
		}
		dominantOut := math.NaN()
		if hasDominant {
			dominantOut = dominant
		}
		s.lastOutput = PeriodogramOutput{
			Heatmap: outputs.NewHeatmap(
				t,
				float64(s.params.MinPeriod),
				float64(s.params.MaxPeriod),
				1.0,
				valueMin,
				valueMax,
				slices.Clone(values),
			),
			DominantCycle: dominantOut,
		}
		emitted = PeriodogramOutput{
			Heatmap: outputs.NewHeatmap(
				t,
				float64(s.params.MinPeriod),
				float64(s.params.MaxPeriod),
				1.0,
				valueMin,
				valueMax,
				values,
			),
			DominantCycle: dominantOut,
		}
	} else {
		emptyOut := PeriodogramOutput{
			Heatmap:       outputs.NewEmptyHeatmap(t, float64(s.params.MinPeriod), float64(s.params.MaxPeriod), 1.0),
			DominantCycle: math.NaN(),
		}
		s.lastOutput = emptyOut
		emitted = emptyOut
	}

	return emitted, nil
}

// Update feeds the next sample and returns PeriodogramOutput, panicking on non-NaN invalid/overflowing input.
func (s *AutoCorrelationPeriodogram) Update(sample float64, t time.Time) PeriodogramOutput {
	result, err := s.TryUpdate(sample, t)
	if err != nil {
		panic(err)
	}
	return result
}

func (s *AutoCorrelationPeriodogram) wrapEntity(t time.Time, result PeriodogramOutput) core.Output {
	return core.Output{
		result.Heatmap,
		entities.Scalar{Time: t, Value: result.DominantCycle},
	}
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *AutoCorrelationPeriodogram) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(sample.Value, sample.Time))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *AutoCorrelationPeriodogram) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.barFunc(sample), sample.Time))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *AutoCorrelationPeriodogram) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.quoteFunc(sample), sample.Time))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *AutoCorrelationPeriodogram) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.tradeFunc(sample), sample.Time))
}
