// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package compositecycle

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// CompositeCycle implements the streaming prepared research indicator.
type CompositeCycle struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewCompositeCycle creates a validated streaming indicator.
func NewCompositeCycle(p *Params) (*CompositeCycle, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Window, "window", 6); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.Smoothing, "smoothing", 0); err != nil {
		return nil, err
	}
	if p.Periods == nil {
		return nil, fmt.Errorf("periods must be nonempty")
	}
	seen := map[float64]bool{}
	for _, period := range p.Periods {
		if err := harmoniccycle.Real(period, "period", 2.0000000000000004); err != nil {
			return nil, err
		}
		if seen[period] {
			return nil, fmt.Errorf("periods must be distinct")
		}
		seen[period] = true
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
	values := []string{strconv.Itoa(p.Window), formatPeriods(p.Periods), strconv.FormatFloat(p.Smoothing, 'g', 12, 64)}
	mnemonic := "ccyc(" + strings.Join(values, ",") + suffix + ")"
	description := "Composite cycle " + mnemonic
	s := &CompositeCycle{engine: &engine{p: *p}}
	s.engine.p.Periods = append([]float64{}, p.Periods...)
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}
func formatPeriods(periods []float64) string {
	values := make([]string, len(periods))
	for i, period := range periods {
		values[i] = strconv.FormatFloat(period, 'g', 12, 64)
	}
	return strings.Join(values, "_")
}

// IsPrimed reports numerical readiness.
func (s *CompositeCycle) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *CompositeCycle) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *CompositeCycle) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *CompositeCycle) Metadata() core.Metadata {
	return core.BuildMetadata(core.CompositeCycle, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " value", Description: s.description + " value"}, {Mnemonic: s.mnemonic + " slope", Description: s.description + " slope"}, {Mnemonic: s.mnemonic + " next_value", Description: s.description + " next_value"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *CompositeCycle) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}

// Models returns a copy of the latest harmonic fit.
func (s *CompositeCycle) Models() []harmoniccycle.CycleModel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]harmoniccycle.CycleModel{}, s.engine.models...)
}

// Project evaluates the frozen fit without advancing observations.
func (s *CompositeCycle) Project(barsAhead float64) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := harmoniccycle.Real(barsAhead, "bars_ahead", 0); err != nil {
		panic(err)
	}
	if !s.engine.primed() {
		return math.NaN()
	}
	return harmoniccycle.CompositeValue(s.engine.models, float64(s.engine.p.Window-1)+barsAhead)
}
func (s *CompositeCycle) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 3)
	output = append(output, entities.Scalar{Time: day, Value: result.Value})
	output = append(output, entities.Scalar{Time: day, Value: result.Slope})
	output = append(output, entities.Scalar{Time: day, Value: result.NextValue})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *CompositeCycle) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *CompositeCycle) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *CompositeCycle) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *CompositeCycle) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
