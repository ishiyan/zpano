// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclephasescore

import (
	"math"

	"zpano/indicators/larsvonthienen/cyclescanner"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value  float64
	Period float64
	Phase  float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan, nan} }

type engine struct {
	p       Params
	scanner *cyclescanner.RollingScanner
}

func (s *engine) reset() {
	s.scanner, _ = cyclescanner.NewRollingScanner(cyclescanner.ScanConfig{s.p.Window, s.p.MinPeriod, s.p.MaxPeriod, s.p.Smoothing, s.p.GenuineThreshold, s.p.PhaseCycles, s.p.TopK, s.p.PeakDistance, s.p.MaxSegments})
}
func (s *engine) primed() bool { return s.scanner.IsPrimed() }
func (s *engine) warmup() int  { return s.p.Window - 1 }
func (s *engine) update(price float64) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	scan := s.scanner.Update(price)
	cycle := scan.Cycles[0]
	return Snapshot{harmoniccycle.PhaseState(cycle.Phase), cycle.Period, cycle.Phase}
}
