package coronasignaltonoiseratio

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

// Snapshot owns all 13 diagnostic fields for one CoronaSignalToNoiseRatio update.
type Snapshot struct {
	Heatmap             *outputs.Heatmap
	SignalToNoiseRatio  float64
	PhysicalDecibels    float64
	CalibratedDecibels  float64
	NormalizedPosition  float64
	Signal              float64
	Noise               float64
	MedianBarRange      float64
	Width               float64
	DominantCycle       float64
	DominantCycleMedian float64
	HasCycleEstimate    float64
	RatioIsFinite       float64
}

// SignalToNoiseOutput aliases [Snapshot].
type SignalToNoiseOutput = Snapshot

func validateHighLow(low, high float64) (float64, error) {
	if math.IsNaN(low) || math.IsNaN(high) {
		return math.NaN(), nil
	}
	if math.IsInf(low, 0) || math.IsInf(high, 0) {
		return 0, errors.New("high and low must be finite")
	}
	return validateBarRange(high - low)
}

func validateBarRange(barRange float64) (float64, error) {
	if math.IsNaN(barRange) {
		return math.NaN(), nil
	}
	if math.IsInf(barRange, 0) || barRange < 0.0 {
		return 0, errors.New("barRange must be finite and nonnegative; require high >= low")
	}
	return barRange, nil
}

// CoronaSignalToNoiseRatio is Ehlers' Corona Signal-to-Noise Ratio indicator.
type CoronaSignalToNoiseRatio struct {
	mu             sync.RWMutex
	mnemonic       string
	description    string
	mnemonicSNR    string
	descriptionSNR string
	engine         *corona.Corona
	engineParams   corona.Params
	raster         *corona.Raster
	sampleCount    int
	ranges         [5]float64
	average        float64
	signal         float64
	noise          float64
	barFunc        entities.BarFunc
	quoteFunc      entities.QuoteFunc
	tradeFunc      entities.TradeFunc
}

// NewCoronaSignalToNoiseRatioDefault returns an instance created with default parameters.
func NewCoronaSignalToNoiseRatioDefault() (*CoronaSignalToNoiseRatio, error) {
	return NewCoronaSignalToNoiseRatioParams(DefaultParams())
}

// NewCoronaSignalToNoiseRatio aliases [NewCoronaSignalToNoiseRatioParams].
func NewCoronaSignalToNoiseRatio(p *Params) (*CoronaSignalToNoiseRatio, error) {
	return NewCoronaSignalToNoiseRatioParams(p)
}

// NewCoronaSignalToNoiseRatioParams returns an instance created with the supplied parameters.
func NewCoronaSignalToNoiseRatioParams(p *Params) (*CoronaSignalToNoiseRatio, error) {
	const invalid = "invalid corona signal to noise ratio parameters"

	source := DefaultParams()
	if p != nil {
		copied := *p
		source = &copied
	}
	raster, err := corona.NewRaster(source.RasterLength, source.MinParameterValue, source.MaxParameterValue, source.MaxRasterValue)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	engine, err := corona.NewCorona(&corona.Params{
		HighPassFilterCutoff:   source.HighPassFilterCutoff,
		MinimalPeriod:          source.MinimalPeriod,
		MaximalPeriod:          source.MaximalPeriod,
		DecibelsLowerThreshold: 6.0,
		DecibelsUpperThreshold: raster.Ceiling(),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
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

	comp := core.ComponentTripleMnemonic(bc, qc, tc)
	flags := snrFlags(*source)
	mnemonicValue := fmt.Sprintf("csnr(%d, %s, %s, %s, %d%s%s)",
		raster.Length(),
		corona.FormatNumber(raster.Ceiling()),
		corona.FormatNumber(raster.Lower()),
		corona.FormatNumber(raster.Upper()),
		source.HighPassFilterCutoff,
		flags,
		comp,
	)
	mnemonicSNR := fmt.Sprintf("csnr-snr(%d%s%s)", source.HighPassFilterCutoff, flags, comp)

	instance := &CoronaSignalToNoiseRatio{
		mnemonic:       mnemonicValue,
		description:    "Corona signal to noise ratio " + mnemonicValue,
		mnemonicSNR:    mnemonicSNR,
		descriptionSNR: "Corona signal to noise ratio scalar " + mnemonicSNR,
		engine:         engine,
		engineParams:   engine.Parameters(),
		raster:         raster,
		barFunc:        barFunc,
		quoteFunc:      quoteFunc,
		tradeFunc:      tradeFunc,
	}
	instance.resetLocked()
	return instance, nil
}

func (s *CoronaSignalToNoiseRatio) resetLocked() {
	s.engine.Reset()
	s.sampleCount = 0
	s.raster.Reset()
	s.ranges = [5]float64{}
	s.average = 0
	s.signal = 0
	s.noise = 0
}

// Reset clears streaming state while retaining parameters.
func (s *CoronaSignalToNoiseRatio) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// IsPrimed indicates whether the indicator is primed.
func (s *CoronaSignalToNoiseRatio) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sampleCount >= corona.FirstBankSample
}

// Metadata describes the output data of the indicator.
func (s *CoronaSignalToNoiseRatio) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CoronaSignalToNoiseRatio,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicSNR, Description: s.descriptionSNR},
		},
	)
}

func (s *CoronaSignalToNoiseRatio) emptySnapshot(t time.Time) Snapshot {
	nan := math.NaN()
	return Snapshot{
		Heatmap:             s.raster.Empty(t),
		SignalToNoiseRatio:  nan,
		PhysicalDecibels:    nan,
		CalibratedDecibels:  nan,
		NormalizedPosition:  nan,
		Signal:              nan,
		Noise:               nan,
		MedianBarRange:      nan,
		Width:               nan,
		DominantCycle:       nan,
		DominantCycleMedian: nan,
		HasCycleEstimate:    nan,
		RatioIsFinite:       nan,
	}
}

func (s *CoronaSignalToNoiseRatio) updateFromSnapshotRangeLocked(snapshot corona.Snapshot, barRange float64) (Snapshot, error) {
	if !snapshot.Accepted {
		if _, err := corona.AcceptSnapshot(snapshot, s.engineParams, s.sampleCount+1); err != nil {
			return Snapshot{}, err
		}
		return s.emptySnapshot(snapshot.Time), nil
	}
	rangeValue, err := validateBarRange(barRange)
	if err != nil {
		return Snapshot{}, err
	}
	if math.IsNaN(rangeValue) {
		return Snapshot{}, errors.New("accepted shared snapshot requires a finite bar range")
	}
	if _, err := corona.AcceptSnapshot(snapshot, s.engineParams, s.sampleCount+1); err != nil {
		return Snapshot{}, err
	}
	s.sampleCount = snapshot.SampleCount

	copy(s.ranges[:4], s.ranges[1:])
	s.ranges[4] = rangeValue
	sortedRanges := s.ranges
	slices.Sort(sortedRanges[:])
	medianRange := sortedRanges[2]

	if s.sampleCount == 1 {
		s.average = snapshot.Price
	} else {
		s.average = 0.1*snapshot.Price + 0.9*s.average
		if s.average != 0.0 || snapshot.MaximumPower > 0.0 {
			s.signal = 0.2*math.Sqrt(snapshot.MaximumPower) + 0.9*s.signal
		}
		if s.average != 0.0 && s.sampleCount > 5 {
			s.noise = 0.1*medianRange + 0.9*s.noise
		}
	}

	var physical float64
	switch {
	case s.signal > 0.0 && s.noise > 0.0:
		physical = 20.0 * (math.Log10(s.signal) - math.Log10(s.noise))
	case s.signal > 0.0:
		physical = math.Inf(1)
	case s.noise > 0.0:
		physical = math.Inf(-1)
	default:
		physical = math.NaN()
	}

	calibrated := physical + 3.5
	position := 0.0
	if !math.IsNaN(calibrated) {
		position = corona.Clip(calibrated, 0.0, 10.0) / 10.0
	}
	width := max(0.0, 0.2-0.4*position)
	s.raster.Update(position, width, corona.RasterSNR, position > 0.5)

	if !snapshot.Primed {
		return s.emptySnapshot(snapshot.Time), nil
	}

	mapped := s.raster.Lower() + (s.raster.Upper()-s.raster.Lower())*position
	hasEstimate := 0.0
	if snapshot.HasCycleEstimate {
		hasEstimate = 1.0
	}
	ratioIsFinite := 0.0
	if !math.IsNaN(physical) && !math.IsInf(physical, 0) {
		ratioIsFinite = 1.0
	}
	return Snapshot{
		Heatmap:             s.raster.Column(snapshot.Time),
		SignalToNoiseRatio:  mapped,
		PhysicalDecibels:    physical,
		CalibratedDecibels:  calibrated,
		NormalizedPosition:  position,
		Signal:              s.signal,
		Noise:               s.noise,
		MedianBarRange:      medianRange,
		Width:               width,
		DominantCycle:       snapshot.DominantCycle,
		DominantCycleMedian: snapshot.DominantCycleMedian,
		HasCycleEstimate:    hasEstimate,
		RatioIsFinite:       ratioIsFinite,
	}, nil
}

// TryUpdate feeds price and (low, high) to the standalone engine before mutating state on error.
func (s *CoronaSignalToNoiseRatio) TryUpdate(sample, low, high float64, t time.Time) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ok, err := corona.ValidPrice(sample)
	if err != nil {
		return Snapshot{}, err
	}
	if !ok {
		return s.emptySnapshot(t), nil
	}
	rangeValue, err := validateHighLow(low, high)
	if err != nil {
		return Snapshot{}, err
	}
	if math.IsNaN(rangeValue) {
		return s.emptySnapshot(t), nil
	}
	snapshot, err := s.engine.TryUpdate(sample, t)
	if err != nil {
		return Snapshot{}, err
	}
	return s.updateFromSnapshotRangeLocked(snapshot, rangeValue)
}

// Update feeds price and (low, high) to the standalone engine and returns a full snapshot.
func (s *CoronaSignalToNoiseRatio) Update(sample, low, high float64, t time.Time) Snapshot {
	result, err := s.TryUpdate(sample, low, high, t)
	if err != nil {
		panic(err)
	}
	return result
}

// TryUpdateRange feeds price and an explicit barRange to the standalone engine.
func (s *CoronaSignalToNoiseRatio) TryUpdateRange(sample, barRange float64, t time.Time) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ok, err := corona.ValidPrice(sample)
	if err != nil {
		return Snapshot{}, err
	}
	if !ok {
		return s.emptySnapshot(t), nil
	}
	rangeValue, err := validateBarRange(barRange)
	if err != nil {
		return Snapshot{}, err
	}
	if math.IsNaN(rangeValue) {
		return s.emptySnapshot(t), nil
	}
	snapshot, err := s.engine.TryUpdate(sample, t)
	if err != nil {
		return Snapshot{}, err
	}
	return s.updateFromSnapshotRangeLocked(snapshot, rangeValue)
}

// UpdateRange feeds price and an explicit barRange to the standalone engine.
func (s *CoronaSignalToNoiseRatio) UpdateRange(sample, barRange float64, t time.Time) Snapshot {
	result, err := s.TryUpdateRange(sample, barRange, t)
	if err != nil {
		panic(err)
	}
	return result
}

// TryUpdateFromSnapshot consumes a shared engine snapshot alongside (low, high).
func (s *CoronaSignalToNoiseRatio) TryUpdateFromSnapshot(snapshot corona.Snapshot, low, high float64) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !snapshot.Accepted {
		return s.updateFromSnapshotRangeLocked(snapshot, 0)
	}
	rangeValue, err := validateHighLow(low, high)
	if err != nil {
		return Snapshot{}, err
	}
	return s.updateFromSnapshotRangeLocked(snapshot, rangeValue)
}

// UpdateFromSnapshot consumes a shared engine snapshot alongside (low, high).
func (s *CoronaSignalToNoiseRatio) UpdateFromSnapshot(snapshot corona.Snapshot, low, high float64) Snapshot {
	result, err := s.TryUpdateFromSnapshot(snapshot, low, high)
	if err != nil {
		panic(err)
	}
	return result
}

// TryUpdateFromSnapshotRange consumes a shared engine snapshot alongside an explicit barRange.
func (s *CoronaSignalToNoiseRatio) TryUpdateFromSnapshotRange(snapshot corona.Snapshot, barRange float64) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.updateFromSnapshotRangeLocked(snapshot, barRange)
}

// UpdateFromSnapshotRange consumes a shared engine snapshot alongside an explicit barRange.
func (s *CoronaSignalToNoiseRatio) UpdateFromSnapshotRange(snapshot corona.Snapshot, barRange float64) Snapshot {
	result, err := s.TryUpdateFromSnapshotRange(snapshot, barRange)
	if err != nil {
		panic(err)
	}
	return result
}

func (s *CoronaSignalToNoiseRatio) wrapEntity(t time.Time, result Snapshot) core.Output {
	return core.Output{
		result.Heatmap,
		entities.Scalar{Time: t, Value: result.SignalToNoiseRatio},
	}
}

// UpdateScalar updates the indicator given a scalar sample (zero bar range).
func (s *CoronaSignalToNoiseRatio) UpdateScalar(sample *entities.Scalar) core.Output {
	value := sample.Value
	return s.wrapEntity(sample.Time, s.Update(value, value, value, sample.Time))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CoronaSignalToNoiseRatio) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.barFunc(sample), sample.Low, sample.High, sample.Time))
}

// UpdateQuote updates the indicator given a quote sample (zero bar range).
func (s *CoronaSignalToNoiseRatio) UpdateQuote(sample *entities.Quote) core.Output {
	value := s.quoteFunc(sample)
	return s.wrapEntity(sample.Time, s.Update(value, value, value, sample.Time))
}

// UpdateTrade updates the indicator given a trade sample (zero bar range).
func (s *CoronaSignalToNoiseRatio) UpdateTrade(sample *entities.Trade) core.Output {
	value := s.tradeFunc(sample)
	return s.wrapEntity(sample.Time, s.Update(value, value, value, sample.Time))
}
