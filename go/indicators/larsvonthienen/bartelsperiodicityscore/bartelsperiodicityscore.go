// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package bartelsperiodicityscore

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// BartelsPeriodicityScore implements the streaming prepared research indicator.
type BartelsPeriodicityScore struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewBartelsPeriodicityScore creates a validated streaming indicator.
func NewBartelsPeriodicityScore(p *Params) (*BartelsPeriodicityScore, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Period, "period", 3); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.Segments, "segments", 2); err != nil {
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
	values := []string{strconv.Itoa(p.Period), strconv.Itoa(p.Segments), strconv.FormatFloat(p.Smoothing, 'g', 12, 64)}
	mnemonic := "bps(" + strings.Join(values, ",") + suffix + ")"
	description := "Bartels periodicity score " + mnemonic
	s := &BartelsPeriodicityScore{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *BartelsPeriodicityScore) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *BartelsPeriodicityScore) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *BartelsPeriodicityScore) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *BartelsPeriodicityScore) Metadata() core.Metadata {
	return core.BuildMetadata(core.BartelsPeriodicityScore, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " score", Description: s.description + " score"}, {Mnemonic: s.mnemonic + " genuine_percent", Description: s.description + " genuine_percent"}, {Mnemonic: s.mnemonic + " segments", Description: s.description + " segments"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *BartelsPeriodicityScore) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *BartelsPeriodicityScore) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 3)
	output = append(output, entities.Scalar{Time: day, Value: result.Score})
	output = append(output, entities.Scalar{Time: day, Value: result.GenuinePercent})
	output = append(output, entities.Scalar{Time: day, Value: result.Segments})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *BartelsPeriodicityScore) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *BartelsPeriodicityScore) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *BartelsPeriodicityScore) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *BartelsPeriodicityScore) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
