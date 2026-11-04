// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package zerolagexponentialmovingaverage

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan} }

type engine struct {
	p           Params
	smoother    *ZLEMA
	samplesSeen int
}

func (s *engine) reset()       { s.smoother = NewZLEMA(s.p.Length); s.samplesSeen = 0 }
func (s *engine) primed() bool { return s.samplesSeen >= (s.p.Length-1)/2+1 }
func (s *engine) warmup() int  { return (s.p.Length - 1) / 2 }
func (s *engine) update(price float64) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	s.samplesSeen++
	return Snapshot{s.smoother.Update(price)}
}
