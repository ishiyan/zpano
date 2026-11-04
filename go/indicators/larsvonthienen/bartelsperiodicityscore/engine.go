// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package bartelsperiodicityscore

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/hodrickprescottsmoothinganddetrending"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Score          float64
	GenuinePercent float64
	Segments       float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan, nan} }

type engine struct {
	p      Params
	prices []float64
}

func (s *engine) reset()       { s.prices = nil }
func (s *engine) primed() bool { return len(s.prices) == s.p.Period*s.p.Segments }
func (s *engine) warmup() int  { return s.p.Period*s.p.Segments - 1 }
func (s *engine) update(price float64) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	prices := harmoniccycle.Push(append([]float64{}, s.prices...), price, s.p.Period*s.p.Segments)
	output := emptySnapshot()
	if len(prices) == s.p.Period*s.p.Segments {
		score := BartelsScore(hodrickprescottsmoothinganddetrending.CycleData(prices, s.p.Smoothing), s.p.Period, s.p.Segments, s.p.Segments)
		output = Snapshot{score.Score, score.GenuinePercent, score.Segments}
	}
	s.prices = prices
	return output
}
