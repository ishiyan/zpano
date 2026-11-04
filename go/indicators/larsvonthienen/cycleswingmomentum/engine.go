// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cycleswingmomentum

import (
	"math"

	"zpano/indicators/larsvonthienen/adaptivebands"
	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/hodrickprescottsmoothinganddetrending"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value float64
	Lower float64
	Upper float64
	Score float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan, nan, nan} }

type engine struct {
	p              Params
	prices, values []float64
}

func (s *engine) reset()       { s.prices = nil; s.values = nil }
func (s *engine) primed() bool { return len(s.values) == s.p.BandMemory }
func (s *engine) warmup() int  { return s.p.Window + s.p.BandMemory - 2 }
func (s *engine) update(price float64) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	prices := harmoniccycle.Push(append([]float64{}, s.prices...), price, s.p.Window)
	output := emptySnapshot()
	if len(prices) == s.p.Window {
		fast := hodrickprescottsmoothinganddetrending.HPTrend(prices, s.p.FastSmoothing)
		slow := hodrickprescottsmoothinganddetrending.HPTrend(prices, s.p.SlowSmoothing)
		value := harmoniccycle.Finite(fast[len(fast)-1] - slow[len(slow)-1])
		s.values = harmoniccycle.Push(s.values, value, s.p.BandMemory)
		output.Value = value
		if s.primed() {
			output.Lower, output.Upper = adaptivebands.GridBands(s.values, s.p.Leveling)
			output.Score = 0
			if value >= output.Upper {
				output.Score = 1
			} else if value <= output.Lower {
				output.Score = -1
			}
		}
	}
	s.prices = prices
	return output
}
