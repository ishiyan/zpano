// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package phasingoscillator

import (
	"math"

	"zpano/indicators/larsvonthienen/cyclescanner"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value       float64
	ShortScore  float64
	LongScore   float64
	ShortPeriod float64
	LongPeriod  float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan, nan, nan, nan} }

type engine struct {
	p           Params
	short, long *cyclescanner.RollingScanner
}

func (s *engine) reset() {
	s.short, _ = cyclescanner.NewRollingScanner(cyclescanner.ScanConfig{s.p.Window, s.p.ShortMin, s.p.ShortMax, s.p.Smoothing, s.p.GenuineThreshold, s.p.PhaseCycles, 1, 2, 5})
	s.long, _ = cyclescanner.NewRollingScanner(cyclescanner.ScanConfig{s.p.Window, s.p.LongMin, s.p.LongMax, s.p.Smoothing, s.p.GenuineThreshold, s.p.PhaseCycles, 1, 2, 5})
}
func (s *engine) primed() bool { return s.short.IsPrimed() && s.long.IsPrimed() }
func (s *engine) warmup() int  { return s.p.Window - 1 }
func (s *engine) update(price float64) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	short := s.short.Update(price).Cycles[0]
	long := s.long.Update(price).Cycles[0]
	shortScore, longScore := harmoniccycle.PhaseState(short.Phase), harmoniccycle.PhaseState(long.Phase)
	return Snapshot{shortScore + longScore, shortScore, longScore, short.Period, long.Period}
}
