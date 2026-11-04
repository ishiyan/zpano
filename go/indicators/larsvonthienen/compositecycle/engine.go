// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package compositecycle

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/hodrickprescottsmoothinganddetrending"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value     float64
	Slope     float64
	NextValue float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan, nan} }

type engine struct {
	p      Params
	prices []float64
	models []harmoniccycle.CycleModel
}

func (s *engine) reset()       { s.prices = nil; s.models = nil }
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
		models := make([]harmoniccycle.CycleModel, len(s.p.Periods))
		var sum float64
		for i, period := range s.p.Periods {
			models[i] = harmoniccycle.Goertzel(values, period)
			sum += harmoniccycle.CycleSlope(models[i], float64(s.p.Window-1))
		}
		output = Snapshot{harmoniccycle.CompositeValue(models, float64(s.p.Window-1)), harmoniccycle.Finite(sum), harmoniccycle.CompositeValue(models, float64(s.p.Window))}
		s.models = models
	}
	s.prices = prices
	return output
}
