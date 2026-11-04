// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclepolarity

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

// CyclePolarity implements the streaming prepared research indicator.
type CyclePolarity struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewCyclePolarity creates a validated streaming indicator.
func NewCyclePolarity(p *Params) (*CyclePolarity, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Window, "window", 6); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.ShortPeriod, "short_period", 2.0000000000000004); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.LongPeriod, "long_period", 2.0000000000000004); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.Smoothing, "smoothing", 0); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.Tolerance, "tolerance", 0); err != nil {
		return nil, err
	}
	if p.ShortPeriod >= p.LongPeriod {
		return nil, fmt.Errorf("short_period must be smaller than long_period")
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
	values := []string{strconv.Itoa(p.Window), strconv.FormatFloat(p.ShortPeriod, 'g', 12, 64), strconv.FormatFloat(p.LongPeriod, 'g', 12, 64), strconv.FormatFloat(p.Smoothing, 'g', 12, 64), strconv.FormatFloat(p.Tolerance, 'g', 12, 64)}
	mnemonic := "cpol(" + strings.Join(values, ",") + suffix + ")"
	description := "Cycle polarity " + mnemonic
	s := &CyclePolarity{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *CyclePolarity) IsPrimed() bool { s.mu.RLock(); defer s.mu.RUnlock(); return s.engine.primed() }

// Reset clears history while retaining parameters.
func (s *CyclePolarity) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *CyclePolarity) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *CyclePolarity) Metadata() core.Metadata {
	return core.BuildMetadata(core.CyclePolarity, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " polarity", Description: s.description + " polarity"}, {Mnemonic: s.mnemonic + " short_value", Description: s.description + " short_value"}, {Mnemonic: s.mnemonic + " long_value", Description: s.description + " long_value"}, {Mnemonic: s.mnemonic + " short_slope", Description: s.description + " short_slope"}, {Mnemonic: s.mnemonic + " long_slope", Description: s.description + " long_slope"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *CyclePolarity) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *CyclePolarity) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 5)
	output = append(output, entities.Scalar{Time: day, Value: result.Polarity})
	output = append(output, entities.Scalar{Time: day, Value: result.ShortValue})
	output = append(output, entities.Scalar{Time: day, Value: result.LongValue})
	output = append(output, entities.Scalar{Time: day, Value: result.ShortSlope})
	output = append(output, entities.Scalar{Time: day, Value: result.LongSlope})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *CyclePolarity) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *CyclePolarity) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *CyclePolarity) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *CyclePolarity) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
