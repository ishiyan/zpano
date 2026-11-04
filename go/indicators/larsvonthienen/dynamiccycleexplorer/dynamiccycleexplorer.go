// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package dynamiccycleexplorer

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// DynamicCycleExplorer implements the streaming prepared research indicator.
type DynamicCycleExplorer struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewDynamicCycleExplorer creates a validated streaming indicator.
func NewDynamicCycleExplorer(p *Params) (*DynamicCycleExplorer, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Window, "window", 6); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.MinPeriod, "min_period", 3); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.MaxPeriod, "max_period", 3); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.Smoothing, "smoothing", 0); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.GenuineThreshold, "genuine_threshold", 0); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.PhaseCycles, "phase_cycles", 0); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.TopK, "top_k", 1); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.PeakDistance, "peak_distance", 1); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.MaxSegments, "max_segments", 2); err != nil {
		return nil, err
	}
	if p.MinPeriod > p.MaxPeriod {
		return nil, fmt.Errorf("min_period exceeds max_period")
	}
	if p.GenuineThreshold > 100 || p.PhaseCycles != 0 && p.PhaseCycles < 1 {
		return nil, fmt.Errorf("invalid scanner threshold or phase cycles")
	}
	bc, qc, tc := p.BarComponent, p.QuoteComponent, p.TradeComponent
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
		return nil, err
	}
	quoteFunc, err := entities.QuoteComponentFunc(qc)
	if err != nil {
		return nil, err
	}
	tradeFunc, err := entities.TradeComponentFunc(tc)
	if err != nil {
		return nil, err
	}
	suffix := core.ComponentTripleMnemonic(bc, qc, tc)
	values := []string{strconv.Itoa(p.Window), strconv.Itoa(p.MinPeriod), strconv.Itoa(p.MaxPeriod), strconv.FormatFloat(p.Smoothing, 'g', 12, 64), strconv.FormatFloat(p.GenuineThreshold, 'g', 12, 64), strconv.FormatFloat(p.PhaseCycles, 'g', 12, 64), strconv.Itoa(p.TopK), strconv.Itoa(p.PeakDistance), strconv.Itoa(p.MaxSegments)}
	mnemonic := "dce(" + strings.Join(values, ",") + suffix + ")"
	description := "Dynamic cycle explorer " + mnemonic
	s := &DynamicCycleExplorer{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *DynamicCycleExplorer) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *DynamicCycleExplorer) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *DynamicCycleExplorer) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *DynamicCycleExplorer) Metadata() core.Metadata {
	return core.BuildMetadata(core.DynamicCycleExplorer, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " value", Description: s.description + " value"}, {Mnemonic: s.mnemonic + " period", Description: s.description + " period"}, {Mnemonic: s.mnemonic + " phase", Description: s.description + " phase"}, {Mnemonic: s.mnemonic + " eta_top", Description: s.description + " eta_top"}, {Mnemonic: s.mnemonic + " eta_bottom", Description: s.description + " eta_bottom"}, {Mnemonic: s.mnemonic + " highlighter", Description: s.description + " highlighter"}, {Mnemonic: s.mnemonic + " highlighter_index", Description: s.description + " highlighter_index"}, {Mnemonic: s.mnemonic + " fitted_at_highlighter", Description: s.description + " fitted_at_highlighter"}, {Mnemonic: s.mnemonic + " delayed_highlighter", Description: s.description + " delayed_highlighter"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *DynamicCycleExplorer) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *DynamicCycleExplorer) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 9)
	output = append(output, entities.Scalar{Time: day, Value: result.Value})
	output = append(output, entities.Scalar{Time: day, Value: result.Period})
	output = append(output, entities.Scalar{Time: day, Value: result.Phase})
	output = append(output, entities.Scalar{Time: day, Value: result.EtaTop})
	output = append(output, entities.Scalar{Time: day, Value: result.EtaBottom})
	output = append(output, entities.Scalar{Time: day, Value: result.Highlighter})
	output = append(output, entities.Scalar{Time: day, Value: result.HighlighterIndex})
	output = append(output, entities.Scalar{Time: day, Value: result.FittedAtHighlighter})
	points := []outputs.Point{}
	if !math.IsNaN(result.HighlighterIndex) {
		points = append(points, outputs.Point{Offset: int(math.Round(result.Period)) / 2, Value: result.Highlighter})
	}
	output = append(output, outputs.Polyline{Time: day, Points: points})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *DynamicCycleExplorer) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *DynamicCycleExplorer) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *DynamicCycleExplorer) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *DynamicCycleExplorer) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
