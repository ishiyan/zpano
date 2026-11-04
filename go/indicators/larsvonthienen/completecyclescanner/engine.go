// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package completecyclescanner

import (
	"math"

	"zpano/indicators/larsvonthienen/cyclescanner"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Spectrum       []float64
	GenuinePercent []float64
	DominantPeriod float64
	AcceptedCount  float64
	Cycles         []cyclescanner.RankedCycle
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nil, nil, nan, nan, nil} }

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
	return Snapshot{scan.Amplitude, scan.GenuinePercent, scan.Cycles[0].Period, scan.AcceptedCount, scan.Cycles}
}
