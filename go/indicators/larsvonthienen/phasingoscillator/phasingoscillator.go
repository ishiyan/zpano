// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package phasingoscillator

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

// PhasingOscillator implements the streaming prepared research indicator.
type PhasingOscillator struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewPhasingOscillator creates a validated streaming indicator.
func NewPhasingOscillator(p *Params) (*PhasingOscillator, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Window, "window", 6); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.ShortMin, "short_min", 3); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.ShortMax, "short_max", 3); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.LongMin, "long_min", 3); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.LongMax, "long_max", 3); err != nil {
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
	if p.ShortMax >= p.LongMin {
		return nil, fmt.Errorf("short_max must be smaller than long_min")
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
	values := []string{strconv.Itoa(p.Window), strconv.Itoa(p.ShortMin), strconv.Itoa(p.ShortMax), strconv.Itoa(p.LongMin), strconv.Itoa(p.LongMax), strconv.FormatFloat(p.Smoothing, 'g', 12, 64), strconv.FormatFloat(p.GenuineThreshold, 'g', 12, 64), strconv.FormatFloat(p.PhaseCycles, 'g', 12, 64)}
	mnemonic := "phosc(" + strings.Join(values, ",") + suffix + ")"
	description := "Phasing oscillator " + mnemonic
	s := &PhasingOscillator{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *PhasingOscillator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *PhasingOscillator) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *PhasingOscillator) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *PhasingOscillator) Metadata() core.Metadata {
	return core.BuildMetadata(core.PhasingOscillator, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " value", Description: s.description + " value"}, {Mnemonic: s.mnemonic + " short_score", Description: s.description + " short_score"}, {Mnemonic: s.mnemonic + " long_score", Description: s.description + " long_score"}, {Mnemonic: s.mnemonic + " short_period", Description: s.description + " short_period"}, {Mnemonic: s.mnemonic + " long_period", Description: s.description + " long_period"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *PhasingOscillator) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *PhasingOscillator) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 5)
	output = append(output, entities.Scalar{Time: day, Value: result.Value})
	output = append(output, entities.Scalar{Time: day, Value: result.ShortScore})
	output = append(output, entities.Scalar{Time: day, Value: result.LongScore})
	output = append(output, entities.Scalar{Time: day, Value: result.ShortPeriod})
	output = append(output, entities.Scalar{Time: day, Value: result.LongPeriod})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *PhasingOscillator) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *PhasingOscillator) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *PhasingOscillator) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *PhasingOscillator) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
