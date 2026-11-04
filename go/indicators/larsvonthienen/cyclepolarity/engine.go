// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclepolarity

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/hodrickprescottsmoothinganddetrending"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Polarity   float64
	ShortValue float64
	LongValue  float64
	ShortSlope float64
	LongSlope  float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan, nan, nan, nan} }

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
		short := harmoniccycle.Goertzel(values, s.p.ShortPeriod)
		long := harmoniccycle.Goertzel(values, s.p.LongPeriod)
		shortSlope := harmoniccycle.CycleSlope(short, float64(s.p.Window-1))
		longSlope := harmoniccycle.CycleSlope(long, float64(s.p.Window-1))
		output = Snapshot{harmoniccycle.CyclePolarity(shortSlope, longSlope, s.p.Tolerance), harmoniccycle.CycleValue(short, float64(s.p.Window-1)), harmoniccycle.CycleValue(long, float64(s.p.Window-1)), shortSlope, longSlope}
	}
	s.prices = prices
	return output
}
