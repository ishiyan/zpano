// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package zerolagexponentialmovingaverage

import (
	"strconv"
	"strings"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// LarsVonThienenZeroLagExponentialMovingAverage implements the streaming prepared research indicator.
type LarsVonThienenZeroLagExponentialMovingAverage struct {
	mu     sync.RWMutex
	engine *engine
	core.LineIndicator
}

// NewLarsVonThienenZeroLagExponentialMovingAverage creates a validated streaming indicator.
func NewLarsVonThienenZeroLagExponentialMovingAverage(p *Params) (*LarsVonThienenZeroLagExponentialMovingAverage, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Length, "length", 1); err != nil {
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
	values := []string{strconv.Itoa(p.Length)}
	mnemonic := "lzlema(" + strings.Join(values, ",") + suffix + ")"
	description := "Zero lag exponential moving average " + mnemonic
	s := &LarsVonThienenZeroLagExponentialMovingAverage{engine: &engine{p: *p}}
	s.engine.reset()
	s.LineIndicator = core.NewLineIndicator(mnemonic, description, barFunc, quoteFunc, tradeFunc, s.Update)
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *LarsVonThienenZeroLagExponentialMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *LarsVonThienenZeroLagExponentialMovingAverage) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.engine.reset()
}

// WarmupBars returns the minimum number of initial observations.
func (s *LarsVonThienenZeroLagExponentialMovingAverage) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *LarsVonThienenZeroLagExponentialMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(core.LarsVonThienenZeroLagExponentialMovingAverage, s.LineIndicator.Mnemonic, s.LineIndicator.Description, []core.OutputText{{Mnemonic: s.LineIndicator.Mnemonic + " value", Description: s.LineIndicator.Description + " value"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *LarsVonThienenZeroLagExponentialMovingAverage) Update(price float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price).Value
}
