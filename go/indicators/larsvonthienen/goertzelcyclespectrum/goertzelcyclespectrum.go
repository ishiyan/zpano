// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package goertzelcyclespectrum

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

// GoertzelCycleSpectrum implements the streaming prepared research indicator.
type GoertzelCycleSpectrum struct {
	mu                    sync.RWMutex
	engine                *engine
	mnemonic, description string
	barFunc               entities.BarFunc
	quoteFunc             entities.QuoteFunc
	tradeFunc             entities.TradeFunc
}

// NewGoertzelCycleSpectrum creates a validated streaming indicator.
func NewGoertzelCycleSpectrum(p *Params) (*GoertzelCycleSpectrum, error) {
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
	if err := harmoniccycle.Integer(p.Step, "step", 1); err != nil {
		return nil, err
	}
	if err := harmoniccycle.Real(p.Smoothing, "smoothing", 0); err != nil {
		return nil, err
	}
	if p.MinPeriod > p.MaxPeriod {
		return nil, fmt.Errorf("min_period exceeds max_period")
	}
	if (p.MaxPeriod-p.MinPeriod)%p.Step != 0 {
		return nil, fmt.Errorf("step must land on max_period")
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
	values := []string{strconv.Itoa(p.Window), strconv.Itoa(p.MinPeriod), strconv.Itoa(p.MaxPeriod), strconv.Itoa(p.Step), strconv.FormatFloat(p.Smoothing, 'g', 12, 64)}
	mnemonic := "gcs(" + strings.Join(values, ",") + suffix + ")"
	description := "Goertzel cycle spectrum " + mnemonic
	s := &GoertzelCycleSpectrum{engine: &engine{p: *p}}
	s.engine.reset()
	s.mnemonic = mnemonic
	s.description = description
	s.barFunc = barFunc
	s.quoteFunc = quoteFunc
	s.tradeFunc = tradeFunc
	return s, nil
}

// IsPrimed reports numerical readiness.
func (s *GoertzelCycleSpectrum) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.primed()
}

// Reset clears history while retaining parameters.
func (s *GoertzelCycleSpectrum) Reset() { s.mu.Lock(); defer s.mu.Unlock(); s.engine.reset() }

// WarmupBars returns the minimum number of initial observations.
func (s *GoertzelCycleSpectrum) WarmupBars() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.engine.warmup()
}

// Metadata returns registered plot semantics.
func (s *GoertzelCycleSpectrum) Metadata() core.Metadata {
	return core.BuildMetadata(core.GoertzelCycleSpectrum, s.mnemonic, s.description, []core.OutputText{{Mnemonic: s.mnemonic + " amplitude", Description: s.description + " amplitude"}, {Mnemonic: s.mnemonic + " phases", Description: s.description + " phases"}, {Mnemonic: s.mnemonic + " trough_offsets", Description: s.description + " trough_offsets"}, {Mnemonic: s.mnemonic + " dominant_period", Description: s.description + " dominant_period"}, {Mnemonic: s.mnemonic + " dominant_amplitude", Description: s.description + " dominant_amplitude"}, {Mnemonic: s.mnemonic + " dominant_phase", Description: s.description + " dominant_phase"}})
}

// Update accepts one observation and returns a fresh numerical snapshot.
func (s *GoertzelCycleSpectrum) Update(price float64) Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.update(price)
}
func (s *GoertzelCycleSpectrum) wrap(day time.Time, result Snapshot) core.Output {
	output := make(core.Output, 0, 6)
	{
		minimum, maximum := math.NaN(), math.NaN()
		if len(result.Amplitude) > 0 {
			minimum, maximum = result.Amplitude[0], result.Amplitude[0]
			for _, value := range result.Amplitude {
				minimum = min(minimum, value)
				maximum = max(maximum, value)
			}
		}
		output = append(output, outputs.Heatmap{Time: day, ParameterFirst: float64(s.engine.p.MinPeriod), ParameterLast: float64(s.engine.p.MaxPeriod), ParameterResolution: 1 / float64(s.engine.p.Step), ValueMin: minimum, ValueMax: maximum, Values: result.Amplitude})
	}
	{
		minimum, maximum := math.NaN(), math.NaN()
		if len(result.Phases) > 0 {
			minimum, maximum = result.Phases[0], result.Phases[0]
			for _, value := range result.Phases {
				minimum = min(minimum, value)
				maximum = max(maximum, value)
			}
		}
		output = append(output, outputs.Heatmap{Time: day, ParameterFirst: float64(s.engine.p.MinPeriod), ParameterLast: float64(s.engine.p.MaxPeriod), ParameterResolution: 1 / float64(s.engine.p.Step), ValueMin: minimum, ValueMax: maximum, Values: result.Phases})
	}
	{
		minimum, maximum := math.NaN(), math.NaN()
		if len(result.TroughOffsets) > 0 {
			minimum, maximum = result.TroughOffsets[0], result.TroughOffsets[0]
			for _, value := range result.TroughOffsets {
				minimum = min(minimum, value)
				maximum = max(maximum, value)
			}
		}
		output = append(output, outputs.Heatmap{Time: day, ParameterFirst: float64(s.engine.p.MinPeriod), ParameterLast: float64(s.engine.p.MaxPeriod), ParameterResolution: 1 / float64(s.engine.p.Step), ValueMin: minimum, ValueMax: maximum, Values: result.TroughOffsets})
	}
	output = append(output, entities.Scalar{Time: day, Value: result.DominantPeriod})
	output = append(output, entities.Scalar{Time: day, Value: result.DominantAmplitude})
	output = append(output, entities.Scalar{Time: day, Value: result.DominantPhase})
	return output
}

// UpdateScalar extracts the observation and emits its plot outputs.
func (s *GoertzelCycleSpectrum) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.wrap(sample.Time, s.Update(sample.Value))
}

// UpdateBar extracts the observation and emits its plot outputs.
func (s *GoertzelCycleSpectrum) UpdateBar(sample *entities.Bar) core.Output {
	return s.wrap(sample.Time, s.Update(s.barFunc(sample)))
}

// UpdateQuote extracts the observation and emits its plot outputs.
func (s *GoertzelCycleSpectrum) UpdateQuote(sample *entities.Quote) core.Output {
	return s.wrap(sample.Time, s.Update(s.quoteFunc(sample)))
}

// UpdateTrade extracts the observation and emits its plot outputs.
func (s *GoertzelCycleSpectrum) UpdateTrade(sample *entities.Trade) core.Output {
	return s.wrap(sample.Time, s.Update(s.tradeFunc(sample)))
}
