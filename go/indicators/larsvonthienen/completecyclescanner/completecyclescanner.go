// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package completecyclescanner

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// CompleteCycleScanner implements the streaming prepared research indicator.
type CompleteCycleScanner struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewCompleteCycleScanner creates a validated streaming indicator.
func NewCompleteCycleScanner(p *Params) (*CompleteCycleScanner, error) {
	if p == nil {
		p = DefaultParams()
	}
	if err := harmoniccycle.Integer(p.Window, "window", 6); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.MinPeriod, "min_period", 3); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.MaxPeriod, "max_period", 3); err != nil {
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
	if err := harmoniccycle.Integer(p.TopK, "top_k", 1); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.PeakDistance, "peak_distance", 1); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Integer(p.MaxSegments, "max_segments", 2); err != nil {
		return nil, err
	}
	if p.MinPeriod > p.MaxPeriod {
		return nil, fmt.Errorf("min_period exceeds max_period")
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
	values := []string{strconv.Itoa(p.Window), strconv.Itoa(p.MinPeriod), strconv.Itoa(p.MaxPeriod), strconv.FormatFloat(p.Smoothing, 'g', 12, 64), strconv.FormatFloat(p.GenuineThreshold, 'g', 12, 64), strconv.FormatFloat(p.PhaseCycles, 'g', 12, 64), strconv.Itoa(p.TopK), strconv.Itoa(p.PeakDistance), strconv.Itoa(p.MaxSegments)}
	mnemonic := "cscan(" + strings.Join(values, ",") + suffix + ")"
	description := "Complete cycle scanner " + mnemonic
	s := &CompleteCycleScanner{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *CompleteCycleScanner) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *CompleteCycleScanner) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *CompleteCycleScanner) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *CompleteCycleScanner) Metadata() core.Metadata {
	return core.BuildMetadata(core.CompleteCycleScanner, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " spectrum", Description: s.description + " spectrum"}, {Mnemonic: s.mnemonic + " genuine_percent", Description: s.description + " genuine_percent"}, {Mnemonic: s.mnemonic + " dominant_period", Description: s.description + " dominant_period"}, {Mnemonic: s.mnemonic + " accepted_count", Description: s.description + " accepted_count"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *CompleteCycleScanner) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *CompleteCycleScanner) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 4)
	{
		minimum, maximum := math.NaN(), math.NaN()
		if len(result.Spectrum) > 0 {
			minimum, maximum = result.Spectrum[0], result.Spectrum[0]
			for _, value := range result.Spectrum {
				minimum = min(minimum, value)
				maximum = max(maximum, value)
			}
		}
		output = append(output, outputs.Heatmap{Time: day, ParameterFirst: float64(s.engine.p.MinPeriod), ParameterLast: float64(s.engine.p.MaxPeriod), ParameterResolution: 1, ValueMin: minimum, ValueMax: maximum, Values: result.Spectrum})
	}
	{
		minimum, maximum := math.NaN(), math.NaN()
		if len(result.GenuinePercent) > 0 {
			minimum, maximum = result.GenuinePercent[0], result.GenuinePercent[0]
			for _, value := range result.GenuinePercent {
				minimum = min(minimum, value)
				maximum = max(maximum, value)
			}
		}
		output = append(output, outputs.Heatmap{Time: day, ParameterFirst: float64(s.engine.p.MinPeriod), ParameterLast: float64(s.engine.p.MaxPeriod), ParameterResolution: 1, ValueMin: minimum, ValueMax: maximum, Values: result.GenuinePercent})
	}
	output = append(output, entities.Scalar{Time: day, Value: result.DominantPeriod})
	output = append(output, entities.Scalar{Time: day, Value: result.AcceptedCount})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *CompleteCycleScanner) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *CompleteCycleScanner) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *CompleteCycleScanner) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *CompleteCycleScanner) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
