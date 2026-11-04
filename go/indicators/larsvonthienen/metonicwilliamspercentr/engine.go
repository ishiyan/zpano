// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package metonicwilliamspercentr

import (
	"math"
	"time"

	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/metonic"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value          float64
	LinearPercentR float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan} }

type engine struct {
	p           Params
	calendar    metonic.Calendar
	delay       *metonic.DelayMean
	lastValue   float64
	samplesSeen int
	highs, lows []float64
}

func (s *engine) reset() {
	s.calendar = metonic.Calendar{}
	s.delay = metonic.NewDelayMean([]int{s.p.Delay1, s.p.Delay2, s.p.Delay3})
	s.lastValue = math.NaN()
	s.samplesSeen = 0
	s.highs = nil
	s.lows = nil
}
func (s *engine) primed() bool { return !math.IsNaN(s.lastValue) }
func (s *engine) warmup() int  { return max(s.p.Delay1, s.p.Delay2, s.p.Delay3) + s.p.Length - 1 }
func (s *engine) update(price float64, day time.Time, high, low float64) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	harmoniccycle.Sample(high)
	harmoniccycle.Sample(low)
	if math.IsNaN(high) || math.IsNaN(low) {
		return emptySnapshot()
	}
	if low > price || price > high {
		panic("completed HLC bar must satisfy low <= close <= high")
	}
	if err := s.calendar.Validate(day); err != nil {
		panic(err)
	}
	s.highs = harmoniccycle.Push(s.highs, high, s.p.Length)
	s.lows = harmoniccycle.Push(s.lows, low, s.p.Length)
	linear := math.NaN()
	if len(s.highs) == s.p.Length {
		maximum, minimum := s.highs[0], s.lows[0]
		for _, value := range s.highs {
			maximum = max(maximum, value)
		}
		for _, value := range s.lows {
			minimum = min(minimum, value)
		}
		if maximum > minimum {
			linear = -100 * (maximum - price) / (maximum - minimum)
		}
	}
	s.lastValue = s.delay.Update(linear)
	s.samplesSeen++
	s.calendar.Commit(day)
	return Snapshot{s.lastValue, linear}
}
