package discretefouriertransformspectrum

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

// DiscreteFourierTransformSpectrum implements Ehlers' 2013 Book Listing 9-1 DFT spectrum indicator.
type DiscreteFourierTransformSpectrum struct {
	mu            sync.RWMutex
	params        Params
	mnemonic      string
	description   string
	mnemonicDC    string
	descriptionDC string
	roof          *roofFilter
	periods       []int
	cosines       [][]float64
	sines         [][]float64
	count         int
	history       []float64
	raw           []float64
	maximum       float64
	dominant      float64
	hasDominant   bool
	lastOutput    SpectrumOutput
	barFunc       entities.BarFunc
	quoteFunc     entities.QuoteFunc
	tradeFunc     entities.TradeFunc
}

// NewDiscreteFourierTransformSpectrumDefault returns an instance created with default parameters.
func NewDiscreteFourierTransformSpectrumDefault() (*DiscreteFourierTransformSpectrum, error) {
	return NewDiscreteFourierTransformSpectrumParams(DefaultParams())
}

// NewDiscreteFourierTransformSpectrum aliases [NewDiscreteFourierTransformSpectrumParams].
func NewDiscreteFourierTransformSpectrum(p *Params) (*DiscreteFourierTransformSpectrum, error) {
	return NewDiscreteFourierTransformSpectrumParams(p)
}

// NewDiscreteFourierTransformSpectrumParams returns an instance created with the supplied parameters.
func NewDiscreteFourierTransformSpectrumParams(p *Params) (*DiscreteFourierTransformSpectrum, error) {
	const invalid = "invalid discrete Fourier transform spectrum parameters"

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
	if source.WindowLength < 2 {
		return nil, fmt.Errorf("%s: windowLength must be >= 2", invalid)
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
	cosines := make([][]float64, cells)
	sines := make([][]float64, cells)
	for index := range cells {
		period := source.MinPeriod + index
		periods[index] = period
		cosRow := make([]float64, source.WindowLength)
		sinRow := make([]float64, source.WindowLength)
		for n := range source.WindowLength {
			angle := twoPi * float64(n) / float64(period)
			cosRow[n] = math.Cos(angle)
			sinRow[n] = math.Sin(angle)
		}
		cosines[index] = cosRow
		sines[index] = sinRow
	}

	mnemonicValue := source.Mnemonic()
	mnemonicDC := source.DominantCycleMnemonic()

	instance := &DiscreteFourierTransformSpectrum{
		params:        *source,
		mnemonic:      mnemonicValue,
		description:   "Discrete Fourier transform spectrum " + mnemonicValue,
		mnemonicDC:    mnemonicDC,
		descriptionDC: "Discrete Fourier transform spectrum dominant cycle " + mnemonicDC,
		roof:          roof,
		periods:       periods,
		cosines:       cosines,
		sines:         sines,
		barFunc:       barFunc,
		quoteFunc:     quoteFunc,
		tradeFunc:     tradeFunc,
	}
	instance.resetLocked()
	return instance, nil
}

func (s *DiscreteFourierTransformSpectrum) resetLocked() {
	s.roof.reset()
	s.count = 0
	s.history = make([]float64, s.params.WindowLength)
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
func (s *DiscreteFourierTransformSpectrum) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// Parameters returns a copy of the indicator parameters.
func (s *DiscreteFourierTransformSpectrum) Parameters() Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// MinPeriod returns the shortest evaluated cycle period.
func (s *DiscreteFourierTransformSpectrum) MinPeriod() int {
	return s.params.MinPeriod
}

// MaxPeriod returns the longest evaluated cycle period.
func (s *DiscreteFourierTransformSpectrum) MaxPeriod() int {
	return s.params.MaxPeriod
}

// WarmupBars returns the number of initial unprimed samples (WindowLength - 1).
func (s *DiscreteFourierTransformSpectrum) WarmupBars() int {
	return s.params.WindowLength - 1
}

// SamplesSeen returns the number of accepted finite samples processed so far.
func (s *DiscreteFourierTransformSpectrum) SamplesSeen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count
}

// IsPrimed indicates whether the indicator has accumulated WindowLength valid observations.
func (s *DiscreteFourierTransformSpectrum) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count >= s.params.WindowLength
}

// State returns a snapshot of the committed internal spectral state.
func (s *DiscreteFourierTransformSpectrum) State() SpectrumState {
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

// Output returns the latest committed SpectrumOutput.
func (s *DiscreteFourierTransformSpectrum) Output() SpectrumOutput {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastOutput
}

// Metadata describes the output data of the indicator.
func (s *DiscreteFourierTransformSpectrum) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DiscreteFourierTransformSpectrum,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicDC, Description: s.descriptionDC},
		},
	)
}

// TryUpdate feeds the next scalar sample and returns SpectrumOutput or an error on invalid/overflowing input.
func (s *DiscreteFourierTransformSpectrum) TryUpdate(sample float64, t time.Time) (SpectrumOutput, error) {
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

	history := make([]float64, s.params.WindowLength)
	history[0] = filtered
	copy(history[1:], s.history[:s.params.WindowLength-1])

	raw := make([]float64, len(s.periods))
	for index, period := range s.periods {
		compensation := 1.0
		if s.params.SpectralDilationCompensation {
			compensation = float64(period)
		}
		cosinePart := 0.0
		sinePart := 0.0
		cosRow := s.cosines[index]
		sinRow := s.sines[index]
		for n := range s.params.WindowLength {
			cosinePart += history[n] * cosRow[n] / compensation
			sinePart += history[n] * sinRow[n] / compensation
		}
		raw[index] = cosinePart*cosinePart + sinePart*sinePart
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
	s.history = history
	s.raw = raw
	s.maximum = maximum
	s.dominant = dominant
	s.hasDominant = hasDC
	s.count++

	if s.count >= s.params.WindowLength {
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
func (s *DiscreteFourierTransformSpectrum) Update(sample float64, t time.Time) SpectrumOutput {
	result, err := s.TryUpdate(sample, t)
	if err != nil {
		panic(err)
	}
	return result
}

func (s *DiscreteFourierTransformSpectrum) wrapEntity(t time.Time, result SpectrumOutput) core.Output {
	return core.Output{
		result.Heatmap,
		entities.Scalar{Time: t, Value: result.DominantCycle},
	}
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *DiscreteFourierTransformSpectrum) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(sample.Value, sample.Time))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DiscreteFourierTransformSpectrum) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.barFunc(sample), sample.Time))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *DiscreteFourierTransformSpectrum) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.quoteFunc(sample), sample.Time))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *DiscreteFourierTransformSpectrum) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.tradeFunc(sample), sample.Time))
}
