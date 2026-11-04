// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package metonicrelativestrengthindex

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

// MetonicRelativeStrengthIndex implements the streaming prepared research indicator.
type MetonicRelativeStrengthIndex struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewMetonicRelativeStrengthIndex creates a validated streaming indicator.
func NewMetonicRelativeStrengthIndex(p *Params) (*MetonicRelativeStrengthIndex, error) {
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
	values := []string{strconv.Itoa(p.Length), strconv.Itoa(p.Delay1), strconv.Itoa(p.Delay2), strconv.Itoa(p.Delay3)}
	mnemonic := "mrsi(" + strings.Join(values, ",") + suffix + ")"
	description := "Metonic relative strength index " + mnemonic
	s := &MetonicRelativeStrengthIndex{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *MetonicRelativeStrengthIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *MetonicRelativeStrengthIndex) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *MetonicRelativeStrengthIndex) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *MetonicRelativeStrengthIndex) Metadata() core.Metadata {
	return core.BuildMetadata(core.MetonicRelativeStrengthIndex, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " value", Description: s.description + " value"}, {Mnemonic: s.mnemonic + " linear_rsi", Description: s.description + " linear_rsi"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *MetonicRelativeStrengthIndex) Update(price float64, day time.Time) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price, day)
}
func (s *MetonicRelativeStrengthIndex) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 2)
	output = append(output, entities.Scalar{Time: day, Value: result.Value})
	output = append(output, entities.Scalar{Time: day, Value: result.LinearRsi})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *MetonicRelativeStrengthIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value, sample.Time))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *MetonicRelativeStrengthIndex) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample), sample.Time))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *MetonicRelativeStrengthIndex) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample), sample.Time))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *MetonicRelativeStrengthIndex) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample), sample.Time))
}
