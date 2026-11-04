// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclestrength

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/hodrickprescottsmoothinganddetrending"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Strength  float64
	Amplitude float64
	Phase     float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan, nan} }

type engine struct {
	p      Params
	prices []float64
}

func (s *engine) reset()       { s.prices = nil }
func (s *engine) primed() bool { return len(s.prices) == s.p.Window }
func (s *engine) warmup() int  { return s.p.Window - 1 }
func (s *engine) update(price float64) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	prices := harmoniccycle.Push(append([]float64{}, s.prices...), price, s.p.Window)
	output := emptySnapshot()
	if len(prices) == s.p.Window {
		values := hodrickprescottsmoothinganddetrending.CycleData(prices, s.p.Smoothing)
		model := harmoniccycle.Goertzel(values, s.p.Period)
		output = Snapshot{harmoniccycle.CycleStrength(model.Amplitude, s.p.Period), model.Amplitude, model.Phase}
	}
	s.prices = prices
	return output
}
