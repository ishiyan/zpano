// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclicsmoothedrelativestrengthindex

import (
	"math"

	"zpano/indicators/larsvonthienen/adaptivebands"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value float64
	Lower float64
	Upper float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan, nan} }

type engine struct {
	p                  Params
	rsi                *WilderRSI
	rsiHistory, values []float64
	previousValue      float64
}

func (s *engine) reset() {
	s.rsi = NewWilderRSI(s.p.DominantCycle / 2)
	s.rsiHistory = nil
	s.values = nil
	s.previousValue = 0
}
func (s *engine) primed() bool { return len(s.values) == 2*s.p.DominantCycle }
func (s *engine) warmup() int {
	return s.p.DominantCycle/2 + (s.p.Vibration-1)/2 + 2*s.p.DominantCycle - 1
}
func (s *engine) update(price float64) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	rsi := s.rsi.Update(price)
	if math.IsNaN(rsi) {
		return emptySnapshot()
	}
	lag := (s.p.Vibration - 1) / 2
	s.rsiHistory = harmoniccycle.Push(s.rsiHistory, rsi, lag+1)
	if len(s.rsiHistory) < lag+1 {
		return emptySnapshot()
	}
	torque := 2 / float64(s.p.Vibration+1)
	value := harmoniccycle.Finite(torque*(2*rsi-s.rsiHistory[0]) + (1-torque)*s.previousValue)
	s.previousValue = value
	s.values = harmoniccycle.Push(s.values, value, 2*s.p.DominantCycle)
	output := emptySnapshot()
	output.Value = value
	if s.primed() {
		output.Lower, output.Upper = adaptivebands.GridBands(s.values, s.p.Leveling)
	}
	return output
}
