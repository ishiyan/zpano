// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package metonicwilliamspercentr

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

// MetonicWilliamsPercentR implements the streaming prepared research indicator.
type MetonicWilliamsPercentR struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewMetonicWilliamsPercentR creates a validated streaming indicator.
func NewMetonicWilliamsPercentR(p *Params) (*MetonicWilliamsPercentR, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Length, "length", 1); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.Delay1, "delay1", 1); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.Delay2, "delay2", 1); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.Delay3, "delay3", 1); err != nil {
		return nil, err
	}
	if p.Delay1 == p.Delay2 || p.Delay1 == p.Delay3 || p.Delay2 == p.Delay3 {
		return nil, fmt.Errorf("three delays must be distinct")
	}
	suffix := ""
	values := []string{strconv.Itoa(p.Length), strconv.Itoa(p.Delay1), strconv.Itoa(p.Delay2), strconv.Itoa(p.Delay3)}
	mnemonic := "mwillr(" + strings.Join(values, ",") + suffix + ")"
	description := "Metonic williams percent r " + mnemonic
	s := &MetonicWilliamsPercentR{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *MetonicWilliamsPercentR) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *MetonicWilliamsPercentR) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *MetonicWilliamsPercentR) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *MetonicWilliamsPercentR) Metadata() core.Metadata {
	return core.BuildMetadata(core.MetonicWilliamsPercentR, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " value", Description: s.description + " value"}, {Mnemonic: s.mnemonic + " linear_percent_r", Description: s.description + " linear_percent_r"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *MetonicWilliamsPercentR) Update(price float64, day time.Time, high, low float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price, day, high, low)
}
func (s *MetonicWilliamsPercentR) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 2)
	output = append(output, entities.Scalar{Time: day, Value: result.Value})
	output = append(output, entities.Scalar{Time: day, Value: result.LinearPercentR})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *MetonicWilliamsPercentR) UpdateScalar(sample *entities.Scalar) core.Output {
	panic("conventional Metonic Williams %R requires a completed HLC bar")
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *MetonicWilliamsPercentR) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Close, sample.Time, sample.High, sample.Low))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *MetonicWilliamsPercentR) UpdateQuote(sample *entities.Quote) core.Output {
	panic("conventional Metonic Williams %R requires a completed HLC bar")
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *MetonicWilliamsPercentR) UpdateTrade(sample *entities.Trade) core.Output {
	panic("conventional Metonic Williams %R requires a completed HLC bar")
}
