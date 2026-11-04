// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package hodrickprescottsmoothinganddetrending

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Trend float64
	Cycle float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan} }

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
		trend := HPTrend(prices, s.p.Smoothing)
		output = Snapshot{trend[len(trend)-1], harmoniccycle.Finite(price - trend[len(trend)-1])}
	}
	s.prices = prices
	return output
}
