package coronaspectrum

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
	"zpano/indicators/johnehlers/corona"
)

// Snapshot owns all 9 diagnostic fields for one CoronaSpectrum update.
type Snapshot struct {
	Heatmap             *outputs.Heatmap
	DominantCycle       float64
	DominantCycleMedian float64
	MaximumPower        float64
	NormalizationPower  float64
	SelectedBins        float64
	HasCycleEstimate    float64
	Power               []float64
	RelativePower       []float64
}

// SpectrumOutput aliases [Snapshot].
type SpectrumOutput = Snapshot

// CoronaSpectrum is Ehlers' Corona Spectrum heatmap indicator.
type CoronaSpectrum struct {
	mu                  sync.RWMutex
	mnemonic            string
	description         string
	mnemonicDC          string
	descriptionDC       string
	mnemonicDCM         string
	descriptionDCM      string
	engine              *corona.Corona
	engineParams        corona.Params
	sampleCount         int
	minParameterValue   float64
	maxParameterValue   float64
	parameterResolution float64
	barFunc             entities.BarFunc
	quoteFunc           entities.QuoteFunc
	tradeFunc           entities.TradeFunc
}

// NewCoronaSpectrumDefault returns an instance created with default parameters.
func NewCoronaSpectrumDefault() (*CoronaSpectrum, error) {
	return NewCoronaSpectrumParams(DefaultParams())
}

// NewCoronaSpectrum aliases [NewCoronaSpectrumParams].
func NewCoronaSpectrum(p *Params) (*CoronaSpectrum, error) {
	return NewCoronaSpectrumParams(p)
}

// NewCoronaSpectrumParams returns an instance created with the supplied parameters.
func NewCoronaSpectrumParams(p *Params) (*CoronaSpectrum, error) {
	const invalid = "invalid corona spectrum parameters"

	source := DefaultParams()
	if p != nil {
		copied := *p
		source = &copied
	}
	if err := corona.ValidateFinite("MinRasterValue", source.MinRasterValue); err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	if err := corona.ValidateFinite("MaxRasterValue", source.MaxRasterValue); err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	if err := corona.ValidateFinite("MinParameterValue", source.MinParameterValue); err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	if err := corona.ValidateFinite("MaxParameterValue", source.MaxParameterValue); err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	if err := corona.ValidateInteger("HighPassFilterCutoff", source.HighPassFilterCutoff, 5); err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	if !(0.0 <= source.MinRasterValue && source.MinRasterValue < source.MaxRasterValue && source.MaxRasterValue > 0.0) {
		return nil, fmt.Errorf("%s: %w", invalid, errors.New("require 0 <= MinRasterValue < MaxRasterValue"))
	}

	minPeriod := int(math.Ceil(source.MinParameterValue))
	maxPeriod := int(math.Floor(source.MaxParameterValue))
	if minPeriod < 6 {
		return nil, fmt.Errorf("%s: MinParameterValue must ceil to >= 6", invalid)
	}
	if maxPeriod <= minPeriod {
		return nil, fmt.Errorf("%s: MaxParameterValue floor must be > MinParameterValue ceil", invalid)
	}

	bc, qc, tc := source.BarComponent, source.QuoteComponent, source.TradeComponent
	if bc == 0 {
		bc = entities.DefaultBarComponent
	}
	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}
	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	barFunc, err := entities.BarComponentFunc(bc)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	quoteFunc, err := entities.QuoteComponentFunc(qc)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	tradeFunc, err := entities.TradeComponentFunc(tc)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}

	engine, err := corona.NewCorona(&corona.Params{
		HighPassFilterCutoff:   source.HighPassFilterCutoff,
		MinimalPeriod:          minPeriod,
		MaximalPeriod:          maxPeriod,
		DecibelsLowerThreshold: source.MinRasterValue,
		DecibelsUpperThreshold: source.MaxRasterValue,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}

	comp := core.ComponentTripleMnemonic(bc, qc, tc)
	mnemonicValue := fmt.Sprintf("cspect(%s, %s, %d, %d, %d%s)",
		corona.FormatNumber(source.MinRasterValue),
		corona.FormatNumber(source.MaxRasterValue),
		minPeriod,
		maxPeriod,
		source.HighPassFilterCutoff,
		comp,
	)
	mnemonicDC := fmt.Sprintf("cspect-dc(%d%s)", source.HighPassFilterCutoff, comp)
	mnemonicDCM := fmt.Sprintf("cspect-dcm(%d%s)", source.HighPassFilterCutoff, comp)

	return &CoronaSpectrum{
		mnemonic:            mnemonicValue,
		description:         "Corona spectrum " + mnemonicValue,
		mnemonicDC:          mnemonicDC,
		descriptionDC:       "Corona spectrum dominant cycle " + mnemonicDC,
		mnemonicDCM:         mnemonicDCM,
		descriptionDCM:      "Corona spectrum dominant cycle median " + mnemonicDCM,
		engine:              engine,
		engineParams:        engine.Parameters(),
		minParameterValue:   float64(minPeriod),
		maxParameterValue:   float64(maxPeriod),
		parameterResolution: 2.0,
		barFunc:             barFunc,
		quoteFunc:           quoteFunc,
		tradeFunc:           tradeFunc,
	}, nil
}

// Reset clears streaming state while retaining parameters.
func (s *CoronaSpectrum) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.engine.Reset()
	s.sampleCount = 0
}

// IsPrimed indicates whether the indicator is primed.
func (s *CoronaSpectrum) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sampleCount >= corona.FirstBankSample
}

// Metadata describes the output data of the indicator.
func (s *CoronaSpectrum) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CoronaSpectrum,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicDC, Description: s.descriptionDC},
			{Mnemonic: s.mnemonicDCM, Description: s.descriptionDCM},
		},
	)
}

func (s *CoronaSpectrum) emptySnapshot(t time.Time) Snapshot {
	nan := math.NaN()
	bankLength := s.engine.FilterBankLength()
	power := make([]float64, bankLength)
	relativePower := make([]float64, bankLength)
	for i := range bankLength {
		power[i] = nan
		relativePower[i] = nan
	}
	return Snapshot{
		Heatmap:             outputs.NewEmptyHeatmap(t, s.minParameterValue, s.maxParameterValue, s.parameterResolution),
		DominantCycle:       nan,
		DominantCycleMedian: nan,
		MaximumPower:        nan,
		NormalizationPower:  nan,
		SelectedBins:        nan,
		HasCycleEstimate:    nan,
		Power:               power,
		RelativePower:       relativePower,
	}
}

func (s *CoronaSpectrum) updateFromSnapshotLocked(snapshot corona.Snapshot) (Snapshot, error) {
	accepted, err := corona.AcceptSnapshot(snapshot, s.engineParams, s.sampleCount+1)
	if err != nil {
		return Snapshot{}, err
	}
	if !accepted {
		return s.emptySnapshot(snapshot.Time), nil
	}
	s.sampleCount = snapshot.SampleCount
	if !snapshot.Primed {
		return s.emptySnapshot(snapshot.Time), nil
	}

	values := slices.Clone(snapshot.Decibels)
	valueMin := values[0]
	valueMax := values[0]
	for _, value := range values[1:] {
		if value < valueMin {
			valueMin = value
		}
		if value > valueMax {
			valueMax = value
		}
	}
	hasEstimate := 0.0
	if snapshot.HasCycleEstimate {
		hasEstimate = 1.0
	}
	return Snapshot{
		Heatmap:             outputs.NewHeatmap(snapshot.Time, s.minParameterValue, s.maxParameterValue, s.parameterResolution, valueMin, valueMax, values),
		DominantCycle:       snapshot.DominantCycle,
		DominantCycleMedian: snapshot.DominantCycleMedian,
		MaximumPower:        snapshot.MaximumPower,
		NormalizationPower:  snapshot.NormalizationPower,
		SelectedBins:        float64(snapshot.SelectedBins),
		HasCycleEstimate:    hasEstimate,
		Power:               slices.Clone(snapshot.Power),
		RelativePower:       slices.Clone(snapshot.RelativePower),
	}, nil
}

// TryUpdate feeds the next sample to the standalone engine and returns a full snapshot or error.
func (s *CoronaSpectrum) TryUpdate(sample float64, t time.Time) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot, err := s.engine.TryUpdate(sample, t)
	if err != nil {
		return Snapshot{}, err
	}
	return s.updateFromSnapshotLocked(snapshot)
}

// Update feeds the next sample to the standalone engine and returns a full snapshot.
func (s *CoronaSpectrum) Update(sample float64, t time.Time) Snapshot {
	result, err := s.TryUpdate(sample, t)
	if err != nil {
		panic(err)
	}
	return result
}

// TryUpdateFromSnapshot consumes one shared engine snapshot in accepted-sample order.
func (s *CoronaSpectrum) TryUpdateFromSnapshot(snapshot corona.Snapshot) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.updateFromSnapshotLocked(snapshot)
}

// UpdateFromSnapshot consumes one shared engine snapshot and panics on invalid ordering/parameters.
func (s *CoronaSpectrum) UpdateFromSnapshot(snapshot corona.Snapshot) Snapshot {
	result, err := s.TryUpdateFromSnapshot(snapshot)
	if err != nil {
		panic(err)
	}
	return result
}

func (s *CoronaSpectrum) wrapEntity(t time.Time, result Snapshot) core.Output {
	return core.Output{
		result.Heatmap,
		entities.Scalar{Time: t, Value: result.DominantCycle},
		entities.Scalar{Time: t, Value: result.DominantCycleMedian},
	}
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CoronaSpectrum) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(sample.Value, sample.Time))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CoronaSpectrum) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.barFunc(sample), sample.Time))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CoronaSpectrum) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.quoteFunc(sample), sample.Time))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CoronaSpectrum) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.tradeFunc(sample), sample.Time))
}
