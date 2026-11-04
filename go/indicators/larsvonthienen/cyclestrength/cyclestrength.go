// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclestrength

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// CycleStrength implements the streaming prepared research indicator.
type CycleStrength struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewCycleStrength creates a validated streaming indicator.
func NewCycleStrength(p *Params) (*CycleStrength, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Window, "window", 6); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.Period, "period", 2.0000000000000004); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.Smoothing, "smoothing", 0); err != nil {
		return nil, err
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
	values := []string{strconv.Itoa(p.Window), strconv.FormatFloat(p.Period, 'g', 12, 64), strconv.FormatFloat(p.Smoothing, 'g', 12, 64)}
	mnemonic := "cstr(" + strings.Join(values, ",") + suffix + ")"
	description := "Cycle strength " + mnemonic
	s := &CycleStrength{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *CycleStrength) IsPrimed() bool { s.mu.RLock(); defer s.mu.RUnlock(); return s.engine.primed() }

// Reset clears history while retaining parameters.
func (s *CycleStrength) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *CycleStrength) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *CycleStrength) Metadata() core.Metadata {
	return core.BuildMetadata(core.CycleStrength, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " strength", Description: s.description + " strength"}, {Mnemonic: s.mnemonic + " amplitude", Description: s.description + " amplitude"}, {Mnemonic: s.mnemonic + " phase", Description: s.description + " phase"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *CycleStrength) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *CycleStrength) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 3)
	output = append(output, entities.Scalar{Time: day, Value: result.Strength})
	output = append(output, entities.Scalar{Time: day, Value: result.Amplitude})
	output = append(output, entities.Scalar{Time: day, Value: result.Phase})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *CycleStrength) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *CycleStrength) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *CycleStrength) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *CycleStrength) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
