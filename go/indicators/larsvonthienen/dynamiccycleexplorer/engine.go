// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package dynamiccycleexplorer

import (
	"math"

	"zpano/indicators/larsvonthienen/cyclescanner"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value               float64
	Period              float64
	Phase               float64
	EtaTop              float64
	EtaBottom           float64
	Highlighter         float64
	HighlighterIndex    float64
	FittedAtHighlighter float64
}

func emptySnapshot() Snapshot {
	nan := math.NaN()
	return Snapshot{nan, nan, nan, nan, nan, nan, nan, nan}
}

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
	if math.IsNaN(cycle.Period) {
		return emptySnapshot()
	}
	longLength := int(cycle.Period)
	shortLength := max(2, int(math.Floor(cycle.Period/2+0.5)))
	delay := longLength / 2
	center := len(s.scanner.Prices) - 1 - delay
	highlighter := CenteredMean(s.scanner.Prices, center, shortLength) - CenteredMean(s.scanner.Prices, center, longLength)
	index, fitted := math.NaN(), math.NaN()
	if !math.IsNaN(highlighter) && !math.IsInf(highlighter, 0) {
		index = float64(s.scanner.SamplesSeen - 1 - delay)
		fitted = cyclescanner.ModeledValue(cycle, -float64(delay))
	}
	return Snapshot{cycle.Value, cycle.Period, cycle.Phase, cycle.EtaTop, cycle.EtaBottom, highlighter, index, fitted}
}

// CenteredMean returns a retrospective centered average, or NaN at incomplete edges.
func CenteredMean(values []float64, center, length int) float64 {
	half := length / 2
	if center-half < 0 || center+half >= len(values) {
		return math.NaN()
	}
	var sum float64
	if length%2 != 0 {
		for _, value := range values[center-half : center+half+1] {
			sum += value
		}
		return sum / float64(length)
	}
	for _, value := range values[center-half+1 : center+half] {
		sum += value
	}
	return (0.5*values[center-half] + sum + 0.5*values[center+half]) / float64(length)
}
