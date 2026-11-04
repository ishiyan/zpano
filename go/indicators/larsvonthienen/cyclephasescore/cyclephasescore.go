// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclephasescore

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// CyclePhaseScore implements the streaming prepared research indicator.
type CyclePhaseScore struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewCyclePhaseScore creates a validated streaming indicator.
func NewCyclePhaseScore(p *Params) (*CyclePhaseScore, error) {
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
	mnemonic := "cphase(" + strings.Join(values, ",") + suffix + ")"
	description := "Cycle phase score " + mnemonic
	s := &CyclePhaseScore{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *CyclePhaseScore) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *CyclePhaseScore) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *CyclePhaseScore) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *CyclePhaseScore) Metadata() core.Metadata {
	return core.BuildMetadata(core.CyclePhaseScore, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " value", Description: s.description + " value"}, {Mnemonic: s.mnemonic + " period", Description: s.description + " period"}, {Mnemonic: s.mnemonic + " phase", Description: s.description + " phase"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *CyclePhaseScore) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *CyclePhaseScore) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 3)
	output = append(output, entities.Scalar{Time: day, Value: result.Value})
	output = append(output, entities.Scalar{Time: day, Value: result.Period})
	output = append(output, entities.Scalar{Time: day, Value: result.Phase})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *CyclePhaseScore) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *CyclePhaseScore) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *CyclePhaseScore) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *CyclePhaseScore) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
