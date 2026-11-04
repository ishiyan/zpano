// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cycleswingmomentum

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

// CycleSwingMomentum implements the streaming prepared research indicator.
type CycleSwingMomentum struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewCycleSwingMomentum creates a validated streaming indicator.
func NewCycleSwingMomentum(p *Params) (*CycleSwingMomentum, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Window, "window", 6); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.FastSmoothing, "fast_smoothing", 0); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.SlowSmoothing, "slow_smoothing", 0); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.BandMemory, "band_memory", 2); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.Leveling, "leveling", 0); err != nil {
		return nil, err
	}
	if p.Leveling <= 0 || p.Leveling >= 50 {
		return nil, fmt.Errorf("leveling must be in (0,50)")
	}
	if p.FastSmoothing >= p.SlowSmoothing {
		return nil, fmt.Errorf("fast_smoothing must be smaller than slow_smoothing")
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
	values := []string{strconv.Itoa(p.Window), strconv.FormatFloat(p.FastSmoothing, 'g', 12, 64), strconv.FormatFloat(p.SlowSmoothing, 'g', 12, 64), strconv.Itoa(p.BandMemory), strconv.FormatFloat(p.Leveling, 'g', 12, 64)}
	mnemonic := "csmo(" + strings.Join(values, ",") + suffix + ")"
	description := "Cycle swing momentum " + mnemonic
	s := &CycleSwingMomentum{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *CycleSwingMomentum) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *CycleSwingMomentum) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *CycleSwingMomentum) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *CycleSwingMomentum) Metadata() core.Metadata {
	return core.BuildMetadata(core.CycleSwingMomentum, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " value", Description: s.description + " value"}, {Mnemonic: s.mnemonic + " lower", Description: s.description + " lower"}, {Mnemonic: s.mnemonic + " upper", Description: s.description + " upper"}, {Mnemonic: s.mnemonic + " score", Description: s.description + " score"}, {Mnemonic: s.mnemonic + " band", Description: s.description + " band"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *CycleSwingMomentum) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *CycleSwingMomentum) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 5)
	output = append(output, entities.Scalar{Time: day, Value: result.Value})
	output = append(output, entities.Scalar{Time: day, Value: result.Lower})
	output = append(output, entities.Scalar{Time: day, Value: result.Upper})
	output = append(output, entities.Scalar{Time: day, Value: result.Score})
	output = append(output, outputs.Band{Time: day, Lower: result.Lower, Upper: result.Upper})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *CycleSwingMomentum) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *CycleSwingMomentum) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *CycleSwingMomentum) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *CycleSwingMomentum) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
