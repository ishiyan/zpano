// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package metonicsentimentindex

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// MetonicSentimentIndex implements the streaming prepared research indicator.
type MetonicSentimentIndex struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewMetonicSentimentIndex creates a validated streaming indicator.
func NewMetonicSentimentIndex(p *Params) (*MetonicSentimentIndex, error) {
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
	if err := harmoniccycle.Integer(p.Mode, "mode", 0); err != nil {
		return nil, err
	}
	if p.Delay1 == p.Delay2 || p.Delay1 == p.Delay3 || p.Delay2 == p.Delay3 {
		return nil, fmt.Errorf("three delays must be distinct")
	}
	if p.Mode > 1 {
		return nil, fmt.Errorf("mode must be 0 or 1")
	}
	suffix := ""
	values := []string{strconv.Itoa(p.Length), strconv.Itoa(p.Delay1), strconv.Itoa(p.Delay2), strconv.Itoa(p.Delay3), strconv.Itoa(p.Mode)}
	mnemonic := "msent(" + strings.Join(values, ",") + suffix + ")"
	description := "Metonic sentiment index " + mnemonic
	s := &MetonicSentimentIndex{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *MetonicSentimentIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *MetonicSentimentIndex) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *MetonicSentimentIndex) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *MetonicSentimentIndex) Metadata() core.Metadata {
	return core.BuildMetadata(core.MetonicSentimentIndex, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " value", Description: s.description + " value"}, {Mnemonic: s.mnemonic + " daily_sentiment", Description: s.description + " daily_sentiment"}, {Mnemonic: s.mnemonic + " smoothed_close", Description: s.description + " smoothed_close"}, {Mnemonic: s.mnemonic + " smoothed_open", Description: s.description + " smoothed_open"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *MetonicSentimentIndex) Update(price float64, day time.Time, openPrice float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price, day, openPrice)
}
func (s *MetonicSentimentIndex) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 4)
	output = append(output, entities.Scalar{Time: day, Value: result.Value})
	output = append(output, entities.Scalar{Time: day, Value: result.DailySentiment})
	output = append(output, entities.Scalar{Time: day, Value: result.SmoothedClose})
	output = append(output, entities.Scalar{Time: day, Value: result.SmoothedOpen})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *MetonicSentimentIndex) UpdateScalar(sample *entities.Scalar) core.Output {
	if s.engine.p.Mode == 0 {
		panic("original sentiment requires a completed OHLC bar")
	}
	return s.wrap(sample.Time, s.Update(sample.Value, sample.Time, math.NaN()))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *MetonicSentimentIndex) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Close, sample.Time, sample.Open))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *MetonicSentimentIndex) UpdateQuote(sample *entities.Quote) core.Output {
	if s.engine.p.Mode == 0 {
		panic("original sentiment requires a completed OHLC bar")
	}
	return s.wrap(sample.Time, s.Update(sample.Mid(), sample.Time, math.NaN()))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *MetonicSentimentIndex) UpdateTrade(sample *entities.Trade) core.Output {
	if s.engine.p.Mode == 0 {
		panic("original sentiment requires a completed OHLC bar")
	}
	return s.wrap(sample.Time, s.Update(sample.Price, sample.Time, math.NaN()))
}
