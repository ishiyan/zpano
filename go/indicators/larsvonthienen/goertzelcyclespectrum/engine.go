// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package goertzelcyclespectrum

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/hodrickprescottsmoothinganddetrending"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Amplitude         []float64
	Phases            []float64
	TroughOffsets     []float64
	DominantPeriod    float64
	DominantAmplitude float64
	DominantPhase     float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nil, nil, nil, nan, nan, nan} }

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
		count := (s.p.MaxPeriod-s.p.MinPeriod)/s.p.Step + 1
		output.Amplitude = make([]float64, count)
		output.Phases = make([]float64, count)
		output.TroughOffsets = make([]float64, count)
		best := 0
		for i := range count {
			model := harmoniccycle.Goertzel(values, float64(s.p.MinPeriod+i*s.p.Step))
			output.Amplitude[i] = model.Amplitude
			output.Phases[i] = model.Phase
			output.TroughOffsets[i] = model.TroughOffset
			if model.Amplitude > output.Amplitude[best] {
				best = i
			}
		}
		output.DominantAmplitude = output.Amplitude[best]
		output.DominantPhase = output.Phases[best]
		if output.DominantAmplitude > 0 {
			output.DominantPeriod = float64(s.p.MinPeriod + best*s.p.Step)
		}
	}
	s.prices = prices
	return output
}
