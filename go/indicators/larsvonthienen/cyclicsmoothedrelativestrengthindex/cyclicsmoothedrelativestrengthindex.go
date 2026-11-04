// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclicsmoothedrelativestrengthindex

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// CyclicSmoothedRelativeStrengthIndex implements the streaming prepared research indicator.
type CyclicSmoothedRelativeStrengthIndex struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewCyclicSmoothedRelativeStrengthIndex creates a validated streaming indicator.
func NewCyclicSmoothedRelativeStrengthIndex(p *Params) (*CyclicSmoothedRelativeStrengthIndex, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.DominantCycle, "dominant_cycle", 10); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.Vibration, "vibration", 2); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.Leveling, "leveling", 0); err != nil {
		return nil, err
	}
	if p.Leveling <= 0 || p.Leveling >= 50 {
		return nil, fmt.Errorf("leveling must be in (0,50)")
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
	values := []string{strconv.Itoa(p.DominantCycle), strconv.Itoa(p.Vibration), strconv.FormatFloat(p.Leveling, 'g', 12, 64)}
	mnemonic := "crsi(" + strings.Join(values, ",") + suffix + ")"
	description := "Cyclic smoothed relative strength index " + mnemonic
	s := &CyclicSmoothedRelativeStrengthIndex{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *CyclicSmoothedRelativeStrengthIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *CyclicSmoothedRelativeStrengthIndex) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.engine.reset()
}

// WarmupBars returns the minimum number of initial observations.
func (s *CyclicSmoothedRelativeStrengthIndex) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *CyclicSmoothedRelativeStrengthIndex) Metadata() core.Metadata {
	return core.BuildMetadata(core.CyclicSmoothedRelativeStrengthIndex, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " value", Description: s.description + " value"}, {Mnemonic: s.mnemonic + " lower", Description: s.description + " lower"}, {Mnemonic: s.mnemonic + " upper", Description: s.description + " upper"}, {Mnemonic: s.mnemonic + " band", Description: s.description + " band"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *CyclicSmoothedRelativeStrengthIndex) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *CyclicSmoothedRelativeStrengthIndex) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 4)
	output = append(output, entities.Scalar{Time: day, Value: result.Value})
	output = append(output, entities.Scalar{Time: day, Value: result.Lower})
	output = append(output, entities.Scalar{Time: day, Value: result.Upper})
	output = append(output, outputs.Band{Time: day, Lower: result.Lower, Upper: result.Upper})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *CyclicSmoothedRelativeStrengthIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *CyclicSmoothedRelativeStrengthIndex) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *CyclicSmoothedRelativeStrengthIndex) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *CyclicSmoothedRelativeStrengthIndex) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
