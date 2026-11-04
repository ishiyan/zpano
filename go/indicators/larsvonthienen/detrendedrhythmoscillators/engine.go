// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
// Numerical source: Lars von Thienen / WhenToTrade (2017), MPL-2.0.
// Research adaptations are documented in description.md.
package detrendedrhythmoscillators

import (
	"math"
	"sort"

	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value       float64
	Direction   float64
	CycleLength float64
	PivotValue  float64
	PivotOffset float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan, nan, nan, nan} }

type engine struct {
	p                      Params
	fast, slow             float64
	window                 []float64
	pivots                 []pivot
	direction, samplesSeen int
	lastCycleLength        float64
}

func (s *engine) reset() {
	s.fast = math.NaN()
	s.slow = math.NaN()
	s.window = nil
	s.pivots = nil
	s.direction = 0
	s.samplesSeen = 0
	s.lastCycleLength = math.NaN()
}
func (s *engine) primed() bool { return s.samplesSeen > s.warmup() }
func (s *engine) warmup() int  { return rhythmWarmup(s.p) }
func (s *engine) update(price float64) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	value := price
	if s.p.Detrend {
		s.fast = ema(price, s.fast, s.p.FastLength)
		s.slow = ema(price, s.slow, s.p.SlowLength)
		value = math.NaN()
		if s.slow != 0 {
			value = harmoniccycle.Finite(100 * (s.fast - s.slow) / s.slow)
		}
	}
	index := s.samplesSeen
	s.samplesSeen++
	if math.IsNaN(value) {
		return emptySnapshot()
	}
	s.window = harmoniccycle.Push(s.window, value, s.p.ZigPeriod)
	maximum, minimum := s.window[0], s.window[0]
	for _, item := range s.window {
		maximum = max(maximum, item)
		minimum = min(minimum, item)
	}
	high, low := value == maximum, value == minimum
	direction := s.direction
	if high && !low {
		direction = 1
	} else if low && !high {
		direction = -1
	}
	if direction != 0 && (high || low) {
		if direction != s.direction || s.pivots == nil {
			s.pivots = append([]pivot{{value, index}}, s.pivots...)
			if len(s.pivots) > s.p.CycleMemory+3 {
				s.pivots = s.pivots[:s.p.CycleMemory+3]
			}
		} else if direction == 1 && value > s.pivots[0].value || direction == -1 && value < s.pivots[0].value {
			s.pivots[0] = pivot{value, index}
		}
	}
	s.direction = direction
	cycles := make([]float64, 0, len(s.pivots))
	if len(s.pivots) >= 5 {
		for i := 1; i < len(s.pivots)-2; i++ {
			cycles = append(cycles, float64(s.pivots[i].index-s.pivots[i+2].index))
		}
	}
	if len(cycles) > s.p.CycleMemory {
		cycles = cycles[len(cycles)-s.p.CycleMemory:]
	}
	s.lastCycleLength = median(cycles)
	if !s.primed() {
		return emptySnapshot()
	}
	pivotValue, offset := math.NaN(), math.NaN()
	if s.pivots != nil {
		pivotValue = s.pivots[0].value
		offset = float64(s.pivots[0].index - index)
	}
	return Snapshot{value, float64(direction), s.lastCycleLength, pivotValue, offset}
}

type pivot struct {
	value float64
	index int
}

func ema(value, previous float64, length int) float64 {
	if math.IsNaN(previous) {
		return value
	}
	alpha := 2 / float64(length+1)
	return harmoniccycle.Finite(alpha*value + (1-alpha)*previous)
}
func rhythmWarmup(p Params) int {
	if p.Detrend {
		return p.SlowLength - 1
	}
	return 0
}
func median(values []float64) float64 {
	if values == nil || len(values) < 1 {
		return math.NaN()
	}
	sort.Float64s(values)
	middle := len(values) / 2
	if len(values)%2 != 0 {
		return values[middle]
	}
	return (values[middle-1] + values[middle]) / 2
}
