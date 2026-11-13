package coronatrendvigor

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/johnehlers/corona"
)

// Snapshot owns all 13 diagnostic fields for one CoronaTrendVigor update.
type Snapshot struct {
	Heatmap             *outputs.Heatmap
	TrendVigor          float64
	NormalizedVigor     float64
	DominantCycle       float64
	DominantCycleMedian float64
	BandPass            float64
	Quadrature          float64
	CycleAmplitude      float64
	TrendDisplacement   float64
	RawRatio            float64
	Lookback            float64
	Width               float64
	HasCycleEstimate    float64
}

// TrendOutput aliases [Snapshot].
type TrendOutput = Snapshot

// CoronaTrendVigor is Ehlers' Corona Trend Vigor indicator.
type CoronaTrendVigor struct {
	mu             sync.RWMutex
	mnemonic       string
	description    string
	mnemonicTV     string
	descriptionTV  string
	engine         *corona.Corona
	engineParams   corona.Params
	bandpass       *corona.AdaptiveBandPass
	raster         *corona.Raster
	lookbackOffset int
	sampleCount    int
	prices         []float64
	ratio          float64
	ready          bool
	barFunc        entities.BarFunc
	quoteFunc      entities.QuoteFunc
	tradeFunc      entities.TradeFunc
}

// NewCoronaTrendVigorDefault returns an instance created with default parameters.
func NewCoronaTrendVigorDefault() (*CoronaTrendVigor, error) {
	return NewCoronaTrendVigorParams(DefaultParams())
}

// NewCoronaTrendVigor aliases [NewCoronaTrendVigorParams].
func NewCoronaTrendVigor(p *Params) (*CoronaTrendVigor, error) {
	return NewCoronaTrendVigorParams(p)
}

// NewCoronaTrendVigorParams returns an instance created with the supplied parameters.
func NewCoronaTrendVigorParams(p *Params) (*CoronaTrendVigor, error) {
	const invalid = "invalid corona trend vigor parameters"

	source := DefaultParams()
	if p != nil {
		copied := *p
		source = &copied
	}
	raster, err := corona.NewRaster(source.RasterLength, source.MinParameterValue, source.MaxParameterValue, source.MaxRasterValue)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	if source.LookbackOffset != 0 && source.LookbackOffset != -1 {
		return nil, fmt.Errorf("%s: %w", invalid, errors.New("LookbackOffset must be 0 or -1"))
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
	flags := trendFlags(*source)
	mnemonicValue := fmt.Sprintf("ctv(%d, %s, %s, %s, %d%s%s)",
		raster.Length(),
		corona.FormatNumber(raster.Ceiling()),
		corona.FormatNumber(raster.Lower()),
		corona.FormatNumber(raster.Upper()),
		source.HighPassFilterCutoff,
		flags,
		comp,
	)
	mnemonicTV := fmt.Sprintf("ctv-tv(%d%s%s)", source.HighPassFilterCutoff, flags, comp)

	instance := &CoronaTrendVigor{
		mnemonic:       mnemonicValue,
		description:    "Corona trend vigor " + mnemonicValue,
		mnemonicTV:     mnemonicTV,
		descriptionTV:  "Corona trend vigor scalar " + mnemonicTV,
		engine:         engine,
		engineParams:   engine.Parameters(),
		bandpass:       corona.NewAdaptiveBandPass(),
		raster:         raster,
		lookbackOffset: source.LookbackOffset,
		barFunc:        barFunc,
		quoteFunc:      quoteFunc,
		tradeFunc:      tradeFunc,
	}
	instance.resetLocked()
	return instance, nil
}

func (s *CoronaTrendVigor) resetLocked() {
	s.engine.Reset()
	s.sampleCount = 0
	s.bandpass.Reset()
	s.raster.Reset()
	s.prices = s.prices[:0]
	s.ratio = 0
	s.ready = false
}

// Reset clears streaming state while retaining parameters.
func (s *CoronaTrendVigor) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// IsPrimed indicates whether the indicator has both spectral readiness and cycle lookback history.
func (s *CoronaTrendVigor) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready
}

// Metadata describes the output data of the indicator.
func (s *CoronaTrendVigor) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CoronaTrendVigor,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicTV, Description: s.descriptionTV},
		},
	)
}

func (s *CoronaTrendVigor) emptySnapshot(t time.Time) Snapshot {
	nan := math.NaN()
	return Snapshot{
		Heatmap:             s.raster.Empty(t),
		TrendVigor:          nan,
		NormalizedVigor:     nan,
		DominantCycle:       nan,
		DominantCycleMedian: nan,
		BandPass:            nan,
		Quadrature:          nan,
		CycleAmplitude:      nan,
		TrendDisplacement:   nan,
		RawRatio:            nan,
		Lookback:            nan,
		Width:               nan,
		HasCycleEstimate:    nan,
	}
}

func (s *CoronaTrendVigor) updateFromSnapshotLocked(snapshot corona.Snapshot) (Snapshot, error) {
	accepted, err := corona.AcceptSnapshot(snapshot, s.engineParams, s.sampleCount+1)
	if err != nil {
		return Snapshot{}, err
	}
	if !accepted {
		return s.emptySnapshot(snapshot.Time), nil
	}

	cycle, err := s.bandpass.TryUpdate(snapshot.Price, snapshot.DominantCycleMedian)
	if err != nil {
		return Snapshot{}, err
	}
	s.sampleCount = snapshot.SampleCount

	capacity := s.engineParams.MaximalPeriod + 1
	if len(s.prices) < capacity {
		s.prices = append(s.prices, snapshot.Price)
	} else {
		copy(s.prices, s.prices[1:])
		s.prices[len(s.prices)-1] = snapshot.Price
	}

	lag := max(1, int(math.Floor(snapshot.DominantCycleMedian))+s.lookbackOffset)
	hasHistory := len(s.prices) > lag
	trend := math.NaN()
	if hasHistory {
		trend = snapshot.Price - s.prices[len(s.prices)-1-lag]
	}
	rawRatio := math.NaN()
	if hasHistory && cycle.Amplitude > 0.0 {
		rawRatio = trend / cycle.Amplitude
	}
	if hasHistory && trend != 0.0 && cycle.Amplitude > 0.0 {
		blended := 0.33*rawRatio + 0.67*s.ratio
		if math.IsNaN(blended) || math.IsInf(blended, 0) {
			return Snapshot{}, errors.New("calculation overflow; use finite, reasonably scaled prices")
		}
		s.ratio = corona.Clip(blended, -10.0, 10.0)
	}

	vigor := 0.05 * (s.ratio + 10.0)
	width := 0.01
	if 0.3 <= vigor && vigor <= 0.7 {
		width = min(vigor-0.3, 0.7-vigor)
	}
	if width == 0.0 {
		width = 0.01
	}

	s.raster.Update(vigor, width, corona.RasterTrend, vigor < 0.3 || vigor > 0.7)
	s.ready = snapshot.Primed && hasHistory
	if !s.ready {
		return s.emptySnapshot(snapshot.Time), nil
	}

	mapped := s.raster.Lower() + (s.raster.Upper()-s.raster.Lower())*vigor
	hasEstimate := 0.0
	if snapshot.HasCycleEstimate {
		hasEstimate = 1.0
	}
	return Snapshot{
		Heatmap:             s.raster.Column(snapshot.Time),
		TrendVigor:          mapped,
		NormalizedVigor:     vigor,
		DominantCycle:       snapshot.DominantCycle,
		DominantCycleMedian: snapshot.DominantCycleMedian,
		BandPass:            cycle.BandPass,
		Quadrature:          cycle.Quadrature,
		CycleAmplitude:      cycle.Amplitude,
		TrendDisplacement:   trend,
		RawRatio:            rawRatio,
		Lookback:            float64(lag),
		Width:               width,
		HasCycleEstimate:    hasEstimate,
	}, nil
}

// TryUpdate feeds the next sample to the standalone engine and returns a full snapshot or error.
func (s *CoronaTrendVigor) TryUpdate(sample float64, t time.Time) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot, err := s.engine.TryUpdate(sample, t)
	if err != nil {
		return Snapshot{}, err
	}
	return s.updateFromSnapshotLocked(snapshot)
}

// Update feeds the next sample to the standalone engine and returns a full snapshot.
func (s *CoronaTrendVigor) Update(sample float64, t time.Time) Snapshot {
	result, err := s.TryUpdate(sample, t)
	if err != nil {
		panic(err)
	}
	return result
}

// TryUpdateFromSnapshot consumes one shared engine snapshot in accepted-sample order.
func (s *CoronaTrendVigor) TryUpdateFromSnapshot(snapshot corona.Snapshot) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.updateFromSnapshotLocked(snapshot)
}

// UpdateFromSnapshot consumes one shared engine snapshot and panics on invalid ordering/parameters.
func (s *CoronaTrendVigor) UpdateFromSnapshot(snapshot corona.Snapshot) Snapshot {
	result, err := s.TryUpdateFromSnapshot(snapshot)
	if err != nil {
		panic(err)
	}
	return result
}

func (s *CoronaTrendVigor) wrapEntity(t time.Time, result Snapshot) core.Output {
	return core.Output{
		result.Heatmap,
		entities.Scalar{Time: t, Value: result.TrendVigor},
	}
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CoronaTrendVigor) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(sample.Value, sample.Time))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CoronaTrendVigor) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.barFunc(sample), sample.Time))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CoronaTrendVigor) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.quoteFunc(sample), sample.Time))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CoronaTrendVigor) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.tradeFunc(sample), sample.Time))
}
