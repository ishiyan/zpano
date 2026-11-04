// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
// Numerical source: Lars von Thienen / WhenToTrade (2017), MPL-2.0.
// Research adaptations are documented in description.md.
package detrendedrhythmoscillators

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

// DetrendedRhythmOscillators implements the streaming prepared research indicator.
type DetrendedRhythmOscillators struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewDetrendedRhythmOscillators creates a validated streaming indicator.
func NewDetrendedRhythmOscillators(p *Params) (*DetrendedRhythmOscillators, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.FastLength, "fast_length", 1); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.SlowLength, "slow_length", 2); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.ZigPeriod, "zig_period", 5); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.CycleMemory, "cycle_memory", 3); err != nil {
		return nil, err
	}
	if p.FastLength >= p.SlowLength {
		return nil, fmt.Errorf("fast_length must be smaller than slow_length")
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
	detrend := "0"
	if p.Detrend {
		detrend = "1"
	}
	values := []string{strconv.Itoa(p.FastLength), strconv.Itoa(p.SlowLength), strconv.Itoa(p.ZigPeriod), strconv.Itoa(p.CycleMemory), detrend}
	mnemonic := "dro(" + strings.Join(values, ",") + suffix + ")"
	description := "Detrended rhythm oscillators " + mnemonic
	s := &DetrendedRhythmOscillators{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *DetrendedRhythmOscillators) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *DetrendedRhythmOscillators) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *DetrendedRhythmOscillators) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *DetrendedRhythmOscillators) Metadata() core.Metadata {
	return core.BuildMetadata(core.DetrendedRhythmOscillators, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " value", Description: s.description + " value"}, {Mnemonic: s.mnemonic + " direction", Description: s.description + " direction"}, {Mnemonic: s.mnemonic + " cycle_length", Description: s.description + " cycle_length"}, {Mnemonic: s.mnemonic + " pivot_value", Description: s.description + " pivot_value"}, {Mnemonic: s.mnemonic + " pivot_offset", Description: s.description + " pivot_offset"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *DetrendedRhythmOscillators) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *DetrendedRhythmOscillators) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 5)
	output = append(output, entities.Scalar{Time: day, Value: result.Value})
	output = append(output, entities.Scalar{Time: day, Value: result.Direction})
	output = append(output, entities.Scalar{Time: day, Value: result.CycleLength})
	output = append(output, entities.Scalar{Time: day, Value: result.PivotValue})
	output = append(output, entities.Scalar{Time: day, Value: result.PivotOffset})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *DetrendedRhythmOscillators) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *DetrendedRhythmOscillators) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *DetrendedRhythmOscillators) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *DetrendedRhythmOscillators) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
