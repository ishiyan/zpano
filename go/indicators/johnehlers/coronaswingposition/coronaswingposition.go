package coronaswingposition

import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/johnehlers/corona"
)

// Snapshot owns all 12 diagnostic fields for one CoronaSwingPosition update.
type Snapshot struct {
	Heatmap             *outputs.Heatmap
	SwingPosition       float64
	NormalizedPosition  float64
	DominantCycle       float64
	DominantCycleMedian float64
	BandPass            float64
	Quadrature          float64
	CycleAmplitude      float64
	Lead60              float64
	PositionSpan        float64
	Width               float64
	HasCycleEstimate    float64
}

// SwingOutput aliases [Snapshot].
type SwingOutput = Snapshot

// CoronaSwingPosition is Ehlers' Corona Swing Position indicator.
type CoronaSwingPosition struct {
	mu             sync.RWMutex
	mnemonic       string
	description    string
	mnemonicSP     string
	descriptionSP  string
	engine         *corona.Corona
	engineParams   corona.Params
	bandpass       *corona.AdaptiveBandPass
	raster         *corona.Raster
	leadWindow     int
	positionWindow int
	sampleCount    int
	leads          []float64
	positions      []float64
	barFunc        entities.BarFunc
	quoteFunc      entities.QuoteFunc
	tradeFunc      entities.TradeFunc
}

// NewCoronaSwingPositionDefault returns an instance created with default parameters.
func NewCoronaSwingPositionDefault() (*CoronaSwingPosition, error) {
	return NewCoronaSwingPositionParams(DefaultParams())
}

// NewCoronaSwingPosition aliases [NewCoronaSwingPositionParams].
func NewCoronaSwingPosition(p *Params) (*CoronaSwingPosition, error) {
	return NewCoronaSwingPositionParams(p)
}

// NewCoronaSwingPositionParams returns an instance created with the supplied parameters.
func NewCoronaSwingPositionParams(p *Params) (*CoronaSwingPosition, error) {
	const invalid = "invalid corona swing position parameters"

	source := DefaultParams()
	if p != nil {
		copied := *p
		source = &copied
	}
	raster, err := corona.NewRaster(source.RasterLength, source.MinParameterValue, source.MaxParameterValue, source.MaxRasterValue)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	if err := corona.ValidateInteger("LeadWindow", source.LeadWindow, 2); err != nil {
		return nil, fmt.Errorf("%s: %w", invalid, err)
	}
	if err := corona.ValidateInteger("PositionWindow", source.PositionWindow, 2); err != nil {
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
	flags := swingFlags(*source)
	mnemonicValue := fmt.Sprintf("cswing(%d, %s, %s, %s, %d%s%s)",
		raster.Length(),
		corona.FormatNumber(raster.Ceiling()),
		corona.FormatNumber(raster.Lower()),
		corona.FormatNumber(raster.Upper()),
		source.HighPassFilterCutoff,
		flags,
		comp,
	)
	mnemonicSP := fmt.Sprintf("cswing-sp(%d%s%s)", source.HighPassFilterCutoff, flags, comp)

	instance := &CoronaSwingPosition{
		mnemonic:       mnemonicValue,
		description:    "Corona swing position " + mnemonicValue,
		mnemonicSP:     mnemonicSP,
		descriptionSP:  "Corona swing position scalar " + mnemonicSP,
		engine:         engine,
		engineParams:   engine.Parameters(),
		bandpass:       corona.NewAdaptiveBandPass(),
		raster:         raster,
		leadWindow:     source.LeadWindow,
		positionWindow: source.PositionWindow,
		barFunc:        barFunc,
		quoteFunc:      quoteFunc,
		tradeFunc:      tradeFunc,
	}
	instance.resetLocked()
	return instance, nil
}

func (s *CoronaSwingPosition) resetLocked() {
	s.engine.Reset()
	s.sampleCount = 0
	s.bandpass.Reset()
	s.raster.Reset()
	s.leads = make([]float64, s.leadWindow)
	s.positions = make([]float64, s.positionWindow)
}

// Reset clears streaming state while retaining parameters.
func (s *CoronaSwingPosition) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// IsPrimed indicates whether the indicator is primed.
func (s *CoronaSwingPosition) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sampleCount >= corona.FirstBankSample
}

// Metadata describes the output data of the indicator.
func (s *CoronaSwingPosition) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.CoronaSwingPosition,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicSP, Description: s.descriptionSP},
		},
	)
}

func (s *CoronaSwingPosition) emptySnapshot(t time.Time) Snapshot {
	nan := math.NaN()
	return Snapshot{
		Heatmap:             s.raster.Empty(t),
		SwingPosition:       nan,
		NormalizedPosition:  nan,
		DominantCycle:       nan,
		DominantCycleMedian: nan,
		BandPass:            nan,
		Quadrature:          nan,
		CycleAmplitude:      nan,
		Lead60:              nan,
		PositionSpan:        nan,
		Width:               nan,
		HasCycleEstimate:    nan,
	}
}

func (s *CoronaSwingPosition) updateFromSnapshotLocked(snapshot corona.Snapshot) (Snapshot, error) {
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

	var lowest, highest float64
	s.leads, lowest, highest = corona.RollingAppend(s.leads, cycle.Lead60, s.leadWindow)
	flat := highest == lowest
	position := 0.5
	if !flat {
		position = (cycle.Lead60 - lowest) / (highest - lowest)
	}
	position = corona.Clip(position, 0.0, 1.0)

	var lowPosition, highPosition float64
	s.positions, lowPosition, highPosition = corona.RollingAppend(s.positions, position, s.positionWindow)
	span := highPosition - lowPosition
	width := 0.15 * span
	if span > 0.85 {
		width = 0.01
	}
	if width == 0.0 {
		width = 0.01
	}

	s.raster.Update(position, width, corona.RasterSwing, span > 0.8 || flat)
	if !snapshot.Primed {
		return s.emptySnapshot(snapshot.Time), nil
	}

	mapped := s.raster.Lower() + (s.raster.Upper()-s.raster.Lower())*position
	hasEstimate := 0.0
	if snapshot.HasCycleEstimate {
		hasEstimate = 1.0
	}
	return Snapshot{
		Heatmap:             s.raster.Column(snapshot.Time),
		SwingPosition:       mapped,
		NormalizedPosition:  position,
		DominantCycle:       snapshot.DominantCycle,
		DominantCycleMedian: snapshot.DominantCycleMedian,
		BandPass:            cycle.BandPass,
		Quadrature:          cycle.Quadrature,
		CycleAmplitude:      cycle.Amplitude,
		Lead60:              cycle.Lead60,
		PositionSpan:        span,
		Width:               width,
		HasCycleEstimate:    hasEstimate,
	}, nil
}

// TryUpdate feeds the next sample to the standalone engine and returns a full snapshot or error.
func (s *CoronaSwingPosition) TryUpdate(sample float64, t time.Time) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot, err := s.engine.TryUpdate(sample, t)
	if err != nil {
		return Snapshot{}, err
	}
	return s.updateFromSnapshotLocked(snapshot)
}

// Update feeds the next sample to the standalone engine and returns a full snapshot.
func (s *CoronaSwingPosition) Update(sample float64, t time.Time) Snapshot {
	result, err := s.TryUpdate(sample, t)
	if err != nil {
		panic(err)
	}
	return result
}

// TryUpdateFromSnapshot consumes one shared engine snapshot in accepted-sample order.
func (s *CoronaSwingPosition) TryUpdateFromSnapshot(snapshot corona.Snapshot) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.updateFromSnapshotLocked(snapshot)
}

// UpdateFromSnapshot consumes one shared engine snapshot and panics on invalid ordering/parameters.
func (s *CoronaSwingPosition) UpdateFromSnapshot(snapshot corona.Snapshot) Snapshot {
	result, err := s.TryUpdateFromSnapshot(snapshot)
	if err != nil {
		panic(err)
	}
	return result
}

func (s *CoronaSwingPosition) wrapEntity(t time.Time, result Snapshot) core.Output {
	return core.Output{
		result.Heatmap,
		entities.Scalar{Time: t, Value: result.SwingPosition},
	}
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CoronaSwingPosition) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(sample.Value, sample.Time))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CoronaSwingPosition) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.barFunc(sample), sample.Time))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CoronaSwingPosition) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.quoteFunc(sample), sample.Time))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CoronaSwingPosition) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrapEntity(sample.Time, s.Update(s.tradeFunc(sample), sample.Time))
}
