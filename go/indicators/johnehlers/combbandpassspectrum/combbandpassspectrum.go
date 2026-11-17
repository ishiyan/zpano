// Package combbandpassspectrum implements Ehlers' Comb Band-Pass Spectrum
// indicator from "Cycle Analytics for Traders" (2013), Code Listing 10-1.
package combbandpassspectrum

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
	twoPi             = 2.0 * math.Pi
	sqrtTwo           = math.Sqrt2
	agcDecay          = 0.995
	centroidThreshold = 0.5
	bankStartupBars   = 12
)

// SpectrumOutput holds the heatmap column and center-of-gravity dominant cycle period.
type SpectrumOutput struct {
	Heatmap       *outputs.Heatmap
	DominantCycle float64
}

// SpectrumState holds a snapshot of the committed internal spectral state.
type SpectrumState struct {
	HighPass              float64
	Filtered              float64
	RawPower              []float64
	NormalizationMaximum  float64
	InternalDominantCycle float64
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

func newRoofFilter(highpassPeriod, smoothingPeriod float64) (*roofFilter, error) {
	if math.IsNaN(highpassPeriod) || math.IsInf(highpassPeriod, 0) || highpassPeriod < 5.0 {
		return nil, errors.New("highpassPeriod must be a finite real >= 5")
	}
	if math.IsNaN(smoothingPeriod) || math.IsInf(smoothingPeriod, 0) || smoothingPeriod < 2.0 {
		return nil, errors.New("smoothingPeriod must be a finite real >= 2")
	}

	angle := twoPi / (sqrtTwo * highpassPeriod)
	cosine := math.Cos(angle)
	alpha := (cosine + math.Sin(angle) - 1.0) / cosine
	pole := 1.0 - alpha
	a := math.Exp(-sqrtTwo * math.Pi / smoothingPeriod)
	hp0 := (1.0 - alpha/2.0) * (1.0 - alpha/2.0)
	hp1 := 2.0 * pole
	hp2 := pole * pole
	ss2 := 2.0 * a * math.Cos(sqrtTwo*math.Pi/smoothingPeriod)
	ss3 := -a * a
	ss1 := 1.0 - ss2 - ss3
	if !(pole > 0.0 && pole < 1.0 && a > 0.0 && a < 1.0 && ss1 > 0.0) ||
		math.IsNaN(hp0) || math.IsInf(hp0, 0) || math.IsNaN(ss1) || math.IsInf(ss1, 0) {
		return nil, errors.New("periods must produce finite resolvable stable coefficients")
	}

	return &roofFilter{
		hp0: hp0,
		hp1: hp1,
		hp2: hp2,
		ss1: ss1,
		ss2: ss2,
		ss3: ss3,
	}, nil
}

func (r *roofFilter) reset() {
	r.x1 = 0
	r.x2 = 0
	r.h1 = 0
	r.h2 = 0
	r.f1 = 0
	r.f2 = 0
}

func (r *roofFilter) candidate(price float64) (float64, float64, error) {
	high := r.hp0*(price-2.0*r.x1+r.x2) + r.hp1*r.h1 - r.hp2*r.h2
	filtered := r.ss1*(high+r.h1)/2.0 + r.ss2*r.f1 + r.ss3*r.f2
	if math.IsNaN(high) || math.IsInf(high, 0) || math.IsNaN(filtered) || math.IsInf(filtered, 0) {
		return 0, 0, errors.New("roofing arithmetic overflowed")
	}
	return high, filtered, nil
}

func (r *roofFilter) commit(price, high, filtered float64) {
	r.x2, r.x1 = r.x1, price
	r.h2, r.h1 = r.h1, high
	r.f2, r.f1 = r.f1, filtered
}

func normalizeAndCentroid(
	periods []int,
	raw []float64,
	previousMax float64,
	previousDC float64,
	hasDC bool,
	enhance bool,
) (float64, float64, bool, []float64, error) {
	maximum := agcDecay * previousMax
	for _, power := range raw {
		if math.IsNaN(power) || math.IsInf(power, 0) || power < 0.0 {
			return 0, 0, false, nil, errors.New("spectral power arithmetic overflowed")
		}
		if power > maximum {
			maximum = power
		}
	}

	normalized := make([]float64, len(raw))
	if maximum > 0.0 {
		for index, power := range raw {
			normalized[index] = power / maximum
		}
	}

	sp := 0.0
	spx := 0.0
	for index, period := range periods {
		power := normalized[index]
		if power >= centroidThreshold {
			spx += float64(period) * power
			sp += power
		}
	}

	dominant := previousDC
	if sp != 0.0 {
		dominant = spx / sp
		hasDC = true
	}
	if math.IsNaN(dominant) || math.IsInf(dominant, 0) {
		return 0, 0, false, nil, errors.New("centroid arithmetic overflowed")
	}

	values := normalized
	if enhance {
		values = make([]float64, len(normalized))
		for index, value := range normalized {
			values[index] = value * value * value
		}
	}
	return maximum, dominant, hasDC, values, nil
}

// CombBandPassSpectrum implements Ehlers' 2013 Book Listing 10-1 comb band-pass spectrum indicator.
type CombBandPassSpectrum struct {
	mu            sync.RWMutex
	params        Params
	mnemonic      string
	description   string
	mnemonicDC    string
	descriptionDC string
	roof          *roofFilter
	periods       []int
	alpha         []float64
	beta          []float64
	count         int
	bp            [][]float64
	raw           []float64
	maximum       float64
	dominant      float64
	hasDominant   bool
	lastOutput    SpectrumOutput
	barFunc       entities.BarFunc
	quoteFunc     entities.QuoteFunc
	tradeFunc     entities.TradeFunc
}

// NewCombBandPassSpectrumDefault returns an instance created with default parameters.
func NewCombBandPassSpectrumDefault() (*CombBandPassSpectrum, error) {
	return NewCombBandPassSpectrumParams(DefaultParams())
}

// NewCombBandPassSpectrum aliases [NewCombBandPassSpectrumParams].
func NewCombBandPassSpectrum(p *Params) (*CombBandPassSpectrum, error) {
	return NewCombBandPassSpectrumParams(p)
}

// NewCombBandPassSpectrumParams returns an instance created with the supplied parameters.
func NewCombBandPassSpectrumParams(p *Params) (*CombBandPassSpectrum, error) {
	const invalid = "invalid comb band-pass spectrum parameters"

	source := DefaultParams()
	if p != nil {
		copied := *p
		source = &copied
	}

	if source.MinPeriod < 3 {
		return nil, fmt.Errorf("%s: minPeriod must be >= 3", invalid)
	}
	if source.MaxPeriod < source.MinPeriod {
		return nil, fmt.Errorf("%s: maxPeriod must be >= minPeriod", invalid)
	}
	if math.IsNaN(source.Bandwidth) || math.IsInf(source.Bandwidth, 0) ||
		!(source.Bandwidth > 0.0 && source.Bandwidth < float64(source.MinPeriod)/4.0) {
		return nil, fmt.Errorf("%s: bandwidth must satisfy 0 < bandwidth < minPeriod/4", invalid)
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

	cells := source.MaxPeriod - source.MinPeriod + 1
	periods := make([]int, cells)
	alpha := make([]float64, cells)
	beta := make([]float64, cells)
	for index := range cells {
		period := source.MinPeriod + index
		periods[index] = period
		betaValue := math.Cos(twoPi / float64(period))
		gammaValue := 1.0 / math.Cos(twoPi*source.Bandwidth/float64(period))
		alphaValue := gammaValue - math.Sqrt(gammaValue*gammaValue-1.0)
		if math.IsNaN(alphaValue) || math.IsInf(alphaValue, 0) || !(alphaValue > 0.0 && alphaValue < 1.0) {
			return nil, fmt.Errorf("%s: bandwidth must yield resolvable stable band-pass poles", invalid)
		}
		alpha[index] = alphaValue
		beta[index] = betaValue
	}

	mnemonicValue := source.Mnemonic()
	mnemonicDC := source.DominantCycleMnemonic()

	instance := &CombBandPassSpectrum{
		params:        *source,
		mnemonic:      mnemonicValue,
		description:   "Comb band-pass spectrum " + mnemonicValue,
		mnemonicDC:    mnemonicDC,
		descriptionDC: "Comb band-pass spectrum dominant cycle " + mnemonicDC,
		roof:          roof,
		periods:       periods,
		alpha:         alpha,
		beta:          beta,
		barFunc:       barFunc,
		quoteFunc:     quoteFunc,
		tradeFunc:     tradeFunc,
	}
	instance.resetLocked()
	return instance, nil
}

func (s *CombBandPassSpectrum) resetLocked() {
	s.roof.reset()
	s.count = 0
	s.bp = make([][]float64, len(s.periods))
	for index, period := range s.periods {
		s.bp[index] = make([]float64, period)
	}
	s.raw = make([]float64, len(s.periods))
	s.maximum = 0
	s.dominant = 0
	s.hasDominant = false
	s.lastOutput = SpectrumOutput{
		Heatmap:       outputs.NewEmptyHeatmap(time.Time{}, float64(s.params.MinPeriod), float64(s.params.MaxPeriod), 1.0),
		DominantCycle: math.NaN(),
	}
}

// Reset resets streaming state while preserving configuration.
func (s *CombBandPassSpectrum) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// Parameters returns a copy of the indicator parameters.
func (s *CombBandPassSpectrum) Parameters() Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// MinPeriod returns the shortest evaluated cycle period.
func (s *CombBandPassSpectrum) MinPeriod() int {
	return s.params.MinPeriod
}

// MaxPeriod returns the longest evaluated cycle period.
func (s *CombBandPassSpectrum) MaxPeriod() int {
	return s.params.MaxPeriod
}

// WarmupBars returns the number of initial unprimed samples (12 + MaxPeriod - 1).
func (s *CombBandPassSpectrum) WarmupBars() int {
	return bankStartupBars + s.params.MaxPeriod - 1
}

// SamplesSeen returns the number of accepted finite samples processed so far.
func (s *CombBandPassSpectrum) SamplesSeen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count
}

// IsPrimed indicates whether the indicator is primed.
func (s *CombBandPassSpectrum) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count > bankStartupBars+s.params.MaxPeriod-1
}

// State returns a snapshot of the committed internal spectral state.
func (s *CombBandPassSpectrum) State() SpectrumState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return SpectrumState{
		HighPass:              s.roof.h1,
		Filtered:              s.roof.f1,
		RawPower:              slices.Clone(s.raw),
		NormalizationMaximum:  s.maximum,
		InternalDominantCycle: s.dominant,
	}
}

// BandPassHistory returns a deep snapshot of newest-first P-sample band-pass rows per period.
func (s *CombBandPassSpectrum) BandPassHistory() [][]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([][]float64, len(s.bp))
	for index, row := range s.bp {
		rows[index] = slices.Clone(row)
	}
	return rows
}

// Output returns the latest committed SpectrumOutput.
func (s *CombBandPassSpectrum) Output() SpectrumOutput {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastOutput
}

// Metadata describes the output data of the indicator.
func (s *CombBandPassSpectrum) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CombBandPassSpectrum,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicDC, Description: s.descriptionDC},
		},
	)
}

// TryUpdate feeds the next scalar sample and returns SpectrumOutput or an error on invalid/overflowing input.
func (s *CombBandPassSpectrum) TryUpdate(sample float64, t time.Time) (SpectrumOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsInf(sample, 0) {
		return SpectrumOutput{}, errors.New("price must be finite or NaN (missing)")
	}
	if math.IsNaN(sample) {
		return SpectrumOutput{
			Heatmap:       outputs.NewEmptyHeatmap(t, float64(s.params.MinPeriod), float64(s.params.MaxPeriod), 1.0),
			DominantCycle: math.NaN(),
		}, nil
	}

	high, filtered, err := s.roof.candidate(sample)
	if err != nil {
		return SpectrumOutput{}, err
	}

	difference := filtered - s.roof.f2
	nextBP := make([][]float64, len(s.periods))
	raw := make([]float64, len(s.periods))
	for index, period := range s.periods {
		old := s.bp[index]
		current := 0.0
		if s.count+1 > bankStartupBars {
			alpha := s.alpha[index]
			beta := s.beta[index]
			current = 0.5*(1.0-alpha)*difference + beta*(1.0+alpha)*old[0] - alpha*old[1]
		}
		if math.IsNaN(current) || math.IsInf(current, 0) {
			return SpectrumOutput{}, errors.New("band-pass arithmetic overflowed")
		}
		row := make([]float64, period)
		row[0] = current
		copy(row[1:], old[:period-1])
		nextBP[index] = row

		power := 0.0
		if s.count+1 > bankStartupBars {
			compensation := 1.0
			if s.params.SpectralDilationCompensation {
				compensation = float64(period)
			}
			for m := range period {
				scaled := row[m] / compensation
				power += scaled * scaled
			}
		}
		raw[index] = power
	}

	maximum, dominant, hasDC, values, err := normalizeAndCentroid(
		s.periods,
		raw,
		s.maximum,
		s.dominant,
		s.hasDominant,
		s.params.EnhanceResolution,
	)
	if err != nil {
		return SpectrumOutput{}, err
	}

	s.roof.commit(sample, high, filtered)
	s.bp = nextBP
	s.raw = raw
	s.maximum = maximum
	s.dominant = dominant
	s.hasDominant = hasDC
	s.count++

	if s.count > bankStartupBars+s.params.MaxPeriod-1 {
		clonedValues := slices.Clone(values)
		valueMin := clonedValues[0]
		valueMax := clonedValues[0]
		for _, value := range clonedValues[1:] {
			if value < valueMin {
				valueMin = value
			}
			if value > valueMax {
				valueMax = value
			}
		}
		dominantOut := math.NaN()
		if hasDC {
			dominantOut = dominant
		}
		s.lastOutput = SpectrumOutput{
			Heatmap: outputs.NewHeatmap(
				t,
				float64(s.params.MinPeriod),
				float64(s.params.MaxPeriod),
				1.0,
				valueMin,
				valueMax,
				clonedValues,
			),
			DominantCycle: dominantOut,
		}
	} else {
		s.lastOutput = SpectrumOutput{
			Heatmap:       outputs.NewEmptyHeatmap(t, float64(s.params.MinPeriod), float64(s.params.MaxPeriod), 1.0),
			DominantCycle: math.NaN(),
		}
	}

	return s.lastOutput, nil
}

// Update feeds the next sample and returns SpectrumOutput, panicking on non-NaN invalid/overflowing input.
func (s *CombBandPassSpectrum) Update(sample float64, t time.Time) SpectrumOutput {
	result, err := s.TryUpdate(sample, t)
	if err != nil {
		panic(err)
	}
	return result
}

func (s *CombBandPassSpectrum) wrapEntity(t time.Time, result SpectrumOutput) core.Output {
	return core.Output{
		result.Heatmap,
		entities.Scalar{Time: t, Value: result.DominantCycle},
	}
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CombBandPassSpectrum) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(sample.Value, sample.Time))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CombBandPassSpectrum) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.barFunc(sample), sample.Time))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CombBandPassSpectrum) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.quoteFunc(sample), sample.Time))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CombBandPassSpectrum) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.tradeFunc(sample), sample.Time))
}
