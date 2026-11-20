package transformeddiscretefouriertransformspectrum

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
	twoPi           = 2.0 * math.Pi
	logTen          = math.Ln10
	dbCap           = 20.0
	centroidDBLimit = 3.0
)

// SpectrumOutput holds the transformed dB heatmap column and 3 dB center-of-gravity dominant cycle period.
type SpectrumOutput struct {
	Heatmap       *outputs.Heatmap
	DominantCycle float64
}

// SpectrumState holds an immutable diagnostic snapshot of the committed internal spectral state.
type SpectrumState struct {
	HighPass      float64
	CleanedData   float64
	MaximumPower  float64
	DominantCycle float64
	CosineParts   []float64
	SineParts     []float64
	Powers        []float64
	Decibels      []float64
}

// TransformedDiscreteFourierTransformSpectrum implements Ehlers' January 2007 transformed DFT spectrum indicator.
type TransformedDiscreteFourierTransformSpectrum struct {
	mu            sync.RWMutex
	params        Params
	mnemonic      string
	description   string
	mnemonicDC    string
	descriptionDC string
	highpassAlpha float64
	warmupBars    int
	periods       []int
	cosines       [][]float64
	sines         [][]float64
	samplesSeen   int
	previousPrice float64
	hpHistory     [6]float64
	cleaned       float64
	window        []float64
	cosineParts   []float64
	sineParts     []float64
	powers        []float64
	decibels      []float64
	maximum       float64
	dominant      float64
	lastOutput    SpectrumOutput
	barFunc       entities.BarFunc
	quoteFunc     entities.QuoteFunc
	tradeFunc     entities.TradeFunc
}

// NewTransformedDiscreteFourierTransformSpectrumDefault returns an instance created with default parameters.
func NewTransformedDiscreteFourierTransformSpectrumDefault() (*TransformedDiscreteFourierTransformSpectrum, error) {
	return NewTransformedDiscreteFourierTransformSpectrumParams(DefaultParams())
}

// NewTransformedDiscreteFourierTransformSpectrum aliases [NewTransformedDiscreteFourierTransformSpectrumParams].
func NewTransformedDiscreteFourierTransformSpectrum(p *Params) (*TransformedDiscreteFourierTransformSpectrum, error) {
	return NewTransformedDiscreteFourierTransformSpectrumParams(p)
}

// NewTransformedDiscreteFourierTransformSpectrumParams returns an instance created with the supplied parameters.
func NewTransformedDiscreteFourierTransformSpectrumParams(p *Params) (*TransformedDiscreteFourierTransformSpectrum, error) {
	const invalid = "invalid transformed discrete Fourier transform spectrum parameters"

	source := DefaultParams()
	if p != nil {
		copied := *p
		source = &copied
	}

	if source.MinPeriod < 8 {
		return nil, fmt.Errorf("%s: minPeriod must be >= 8", invalid)
	}
	if source.MaxPeriod < source.MinPeriod {
		return nil, fmt.Errorf("%s: maxPeriod must be >= minPeriod", invalid)
	}
	if source.WindowLength < 1 {
		return nil, fmt.Errorf("%s: windowLength must be >= 1", invalid)
	}
	if math.IsNaN(source.HighpassPeriod) || math.IsInf(source.HighpassPeriod, 0) || source.HighpassPeriod <= 4.0 {
		return nil, fmt.Errorf("%s: highpassPeriod must be a finite real > 4", invalid)
	}

	omega := twoPi / source.HighpassPeriod
	cosineOmega := math.Cos(omega)
	if cosineOmega == 0.0 {
		return nil, fmt.Errorf("%s: highpassPeriod produces singular high-pass coefficient", invalid)
	}
	highpassAlpha := (1.0 - math.Sin(omega)) / cosineOmega
	if math.IsNaN(highpassAlpha) || math.IsInf(highpassAlpha, 0) {
		return nil, fmt.Errorf("%s: highpassPeriod produces non-finite high-pass coefficient", invalid)
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

	channelCount := source.MaxPeriod - source.MinPeriod + 1
	periods := make([]int, channelCount)
	cosines := make([][]float64, channelCount)
	sines := make([][]float64, channelCount)
	for channelIndex := range channelCount {
		period := source.MinPeriod + channelIndex
		periods[channelIndex] = period
		cosineRow := make([]float64, source.WindowLength)
		sineRow := make([]float64, source.WindowLength)
		for lag := range source.WindowLength {
			angle := twoPi * float64(lag) / float64(period)
			cosineRow[lag] = math.Cos(angle)
			sineRow[lag] = math.Sin(angle)
		}
		cosines[channelIndex] = cosineRow
		sines[channelIndex] = sineRow
	}

	warmupBars := 5
	if source.WindowLength-1 > warmupBars {
		warmupBars = source.WindowLength - 1
	}

	mnemonicValue := source.Mnemonic()
	mnemonicDC := source.DominantCycleMnemonic()

	instance := &TransformedDiscreteFourierTransformSpectrum{
		params:        *source,
		mnemonic:      mnemonicValue,
		description:   "Transformed discrete Fourier transform spectrum " + mnemonicValue,
		mnemonicDC:    mnemonicDC,
		descriptionDC: "Transformed discrete Fourier transform spectrum dominant cycle " + mnemonicDC,
		highpassAlpha: highpassAlpha,
		warmupBars:    warmupBars,
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

func (s *TransformedDiscreteFourierTransformSpectrum) unavailableOutput(t time.Time) SpectrumOutput {
	return SpectrumOutput{
		Heatmap:       outputs.NewEmptyHeatmap(t, float64(s.params.MinPeriod), float64(s.params.MaxPeriod), 1.0),
		DominantCycle: math.NaN(),
	}
}

func (s *TransformedDiscreteFourierTransformSpectrum) resetLocked() {
	channelCount := len(s.periods)
	s.samplesSeen = 0
	s.previousPrice = 0
	s.hpHistory = [6]float64{}
	s.cleaned = 0
	s.window = make([]float64, s.params.WindowLength)
	s.cosineParts = make([]float64, channelCount)
	s.sineParts = make([]float64, channelCount)
	s.powers = make([]float64, channelCount)
	s.decibels = make([]float64, channelCount)
	s.maximum = 0
	s.dominant = 0
	s.lastOutput = s.unavailableOutput(time.Time{})
}

// Reset resets streaming state while preserving configuration.
func (s *TransformedDiscreteFourierTransformSpectrum) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// Parameters returns a copy of the indicator parameters.
func (s *TransformedDiscreteFourierTransformSpectrum) Parameters() Params {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.params
}

// MinPeriod returns the shortest evaluated cycle period.
func (s *TransformedDiscreteFourierTransformSpectrum) MinPeriod() int {
	return s.params.MinPeriod
}

// MaxPeriod returns the longest evaluated cycle period.
func (s *TransformedDiscreteFourierTransformSpectrum) MaxPeriod() int {
	return s.params.MaxPeriod
}

// WarmupBars returns the number of initial unprimed samples: max(5, WindowLength - 1).
func (s *TransformedDiscreteFourierTransformSpectrum) WarmupBars() int {
	return s.warmupBars
}

// SamplesSeen returns the number of accepted finite samples processed so far.
func (s *TransformedDiscreteFourierTransformSpectrum) SamplesSeen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.samplesSeen
}

// IsPrimed indicates whether the indicator has accumulated max(6, WindowLength) valid observations.
func (s *TransformedDiscreteFourierTransformSpectrum) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.samplesSeen > s.warmupBars
}

// State returns a snapshot of the committed internal spectral state.
func (s *TransformedDiscreteFourierTransformSpectrum) State() SpectrumState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return SpectrumState{
		HighPass:      s.hpHistory[0],
		CleanedData:   s.cleaned,
		MaximumPower:  s.maximum,
		DominantCycle: s.dominant,
		CosineParts:   slices.Clone(s.cosineParts),
		SineParts:     slices.Clone(s.sineParts),
		Powers:        slices.Clone(s.powers),
		Decibels:      slices.Clone(s.decibels),
	}
}

// Output returns the latest committed SpectrumOutput.
func (s *TransformedDiscreteFourierTransformSpectrum) Output() SpectrumOutput {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastOutput
}

// Metadata describes the output data of the indicator.
func (s *TransformedDiscreteFourierTransformSpectrum) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.TransformedDiscreteFourierTransformSpectrum,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicDC, Description: s.descriptionDC},
		},
	)
}

func isFiniteFloat(number float64) bool {
	return !math.IsNaN(number) && !math.IsInf(number, 0)
}

// TryUpdate feeds the next scalar sample and returns SpectrumOutput or an error on invalid/overflowing input.
func (s *TransformedDiscreteFourierTransformSpectrum) TryUpdate(sample float64, t time.Time) (SpectrumOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if math.IsInf(sample, 0) {
		return SpectrumOutput{}, errors.New("input must not be infinite")
	}
	if math.IsNaN(sample) {
		return s.unavailableOutput(t), nil
	}

	barNumber := s.samplesSeen + 1
	var highPass, cleaned float64
	if barNumber <= 5 {
		highPass = sample
		cleaned = sample
	} else {
		highPass = 0.5*(1.0+s.highpassAlpha)*(sample-s.previousPrice) + s.highpassAlpha*s.hpHistory[0]
		cleaned = (highPass +
			2.0*s.hpHistory[0] +
			3.0*s.hpHistory[1] +
			3.0*s.hpHistory[2] +
			2.0*s.hpHistory[3] +
			s.hpHistory[4]) / 12.0
	}
	if !isFiniteFloat(highPass) || !isFiniteFloat(cleaned) {
		return SpectrumOutput{}, errors.New("input causes non-finite filter arithmetic; state was not changed")
	}

	windowLength := s.params.WindowLength
	nextWindow := make([]float64, windowLength)
	nextWindow[0] = cleaned
	copy(nextWindow[1:], s.window[:windowLength-1])

	channelCount := len(s.periods)
	cosineParts := make([]float64, channelCount)
	sineParts := make([]float64, channelCount)
	powers := make([]float64, channelCount)
	maximum := 0.0
	for channelIndex := range channelCount {
		cosineSum := 0.0
		sineSum := 0.0
		cosineRow := s.cosines[channelIndex]
		sineRow := s.sines[channelIndex]
		for lag := range windowLength {
			cosineSum += nextWindow[lag] * cosineRow[lag]
			sineSum += nextWindow[lag] * sineRow[lag]
		}
		power := cosineSum*cosineSum + sineSum*sineSum
		if !isFiniteFloat(cosineSum) || !isFiniteFloat(sineSum) || !isFiniteFloat(power) {
			return SpectrumOutput{}, errors.New("input causes non-finite filter arithmetic; state was not changed")
		}
		cosineParts[channelIndex] = cosineSum
		sineParts[channelIndex] = sineSum
		powers[channelIndex] = power
		if channelIndex == 0 || power > maximum {
			maximum = power
		}
	}

	decibels := slices.Clone(s.decibels)
	for channelIndex, power := range powers {
		if maximum > 0.0 && power > 0.0 {
			decibels[channelIndex] = -10.0 * math.Log(0.01/(1.0-0.99*power/maximum)) / logTen
		}
		if decibels[channelIndex] > dbCap {
			decibels[channelIndex] = dbCap
		}
		if !isFiniteFloat(decibels[channelIndex]) {
			return SpectrumOutput{}, errors.New("input causes non-finite filter arithmetic; state was not changed")
		}
	}

	numerator := 0.0
	denominator := 0.0
	for channelIndex, period := range s.periods {
		channelDB := decibels[channelIndex]
		if channelDB < centroidDBLimit {
			weight := centroidDBLimit - channelDB
			numerator += float64(period) * weight
			denominator += weight
		}
	}
	dominant := s.dominant
	if denominator != 0.0 {
		dominant = numerator / denominator
	}
	if !isFiniteFloat(maximum) || !isFiniteFloat(dominant) {
		return SpectrumOutput{}, errors.New("input causes non-finite filter arithmetic; state was not changed")
	}

	s.previousPrice = sample
	for lag := 5; lag >= 1; lag-- {
		s.hpHistory[lag] = s.hpHistory[lag-1]
	}
	s.hpHistory[0] = highPass
	s.cleaned = cleaned
	s.window = nextWindow
	s.cosineParts = cosineParts
	s.sineParts = sineParts
	s.powers = powers
	s.decibels = decibels
	s.maximum = maximum
	s.dominant = dominant
	s.samplesSeen = barNumber

	if s.samplesSeen > s.warmupBars {
		clonedValues := slices.Clone(decibels)
		valueMin := clonedValues[0]
		valueMax := clonedValues[0]
		for _, channelValue := range clonedValues[1:] {
			if channelValue < valueMin {
				valueMin = channelValue
			}
			if channelValue > valueMax {
				valueMax = channelValue
			}
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
			DominantCycle: dominant,
		}
	} else {
		s.lastOutput = s.unavailableOutput(t)
	}

	return s.lastOutput, nil
}

// Update feeds the next sample and returns SpectrumOutput, panicking on non-NaN invalid/overflowing input.
func (s *TransformedDiscreteFourierTransformSpectrum) Update(sample float64, t time.Time) SpectrumOutput {
	result, err := s.TryUpdate(sample, t)
	if err != nil {
		panic(err)
	}
	return result
}

func (s *TransformedDiscreteFourierTransformSpectrum) wrapEntity(t time.Time, result SpectrumOutput) core.Output {
	return core.Output{
		result.Heatmap,
		entities.Scalar{Time: t, Value: result.DominantCycle},
	}
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *TransformedDiscreteFourierTransformSpectrum) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(sample.Value, sample.Time))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *TransformedDiscreteFourierTransformSpectrum) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.barFunc(sample), sample.Time))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *TransformedDiscreteFourierTransformSpectrum) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.quoteFunc(sample), sample.Time))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *TransformedDiscreteFourierTransformSpectrum) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.tradeFunc(sample), sample.Time))
}
