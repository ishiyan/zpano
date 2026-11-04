// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package hodrickprescottsmoothinganddetrending

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// HodrickPrescottSmoothingAndDetrending implements the streaming prepared research indicator.
type HodrickPrescottSmoothingAndDetrending struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewHodrickPrescottSmoothingAndDetrending creates a validated streaming indicator.
func NewHodrickPrescottSmoothingAndDetrending(p *Params) (*HodrickPrescottSmoothingAndDetrending, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Window, "window", 6); err != nil {
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
	values := []string{strconv.Itoa(p.Window), strconv.FormatFloat(p.Smoothing, 'g', 12, 64)}
	mnemonic := "hpsd(" + strings.Join(values, ",") + suffix + ")"
	description := "Hodrick prescott smoothing and detrending " + mnemonic
	s := &HodrickPrescottSmoothingAndDetrending{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *HodrickPrescottSmoothingAndDetrending) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *HodrickPrescottSmoothingAndDetrending) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.engine.reset()
}

// WarmupBars returns the minimum number of initial observations.
func (s *HodrickPrescottSmoothingAndDetrending) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *HodrickPrescottSmoothingAndDetrending) Metadata() core.Metadata {
	return core.BuildMetadata(core.HodrickPrescottSmoothingAndDetrending, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " trend", Description: s.description + " trend"}, {Mnemonic: s.mnemonic + " cycle", Description: s.description + " cycle"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *HodrickPrescottSmoothingAndDetrending) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *HodrickPrescottSmoothingAndDetrending) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 2)
	output = append(output, entities.Scalar{Time: day, Value: result.Trend})
	output = append(output, entities.Scalar{Time: day, Value: result.Cycle})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *HodrickPrescottSmoothingAndDetrending) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *HodrickPrescottSmoothingAndDetrending) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *HodrickPrescottSmoothingAndDetrending) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *HodrickPrescottSmoothingAndDetrending) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
