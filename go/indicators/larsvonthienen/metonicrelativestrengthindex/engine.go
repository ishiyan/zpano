// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package metonicrelativestrengthindex

import (
	"math"
	"time"

	"zpano/indicators/larsvonthienen/cyclicsmoothedrelativestrengthindex"
	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/metonic"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value     float64
	LinearRsi float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan} }

type engine struct {
	p           Params
	calendar    metonic.Calendar
	delay       *metonic.DelayMean
	lastValue   float64
	samplesSeen int
	rsi         *cyclicsmoothedrelativestrengthindex.WilderRSI
}

func (s *engine) reset() {
	s.calendar = metonic.Calendar{}
	s.delay = metonic.NewDelayMean([]int{s.p.Delay1, s.p.Delay2, s.p.Delay3})
	s.lastValue = math.NaN()
	s.samplesSeen = 0
	s.rsi = cyclicsmoothedrelativestrengthindex.NewWilderRSI(s.p.Length)
}
func (s *engine) primed() bool { return !math.IsNaN(s.lastValue) }
func (s *engine) warmup() int  { return max(s.p.Delay1, s.p.Delay2, s.p.Delay3) + s.p.Length }
func (s *engine) update(price float64, day time.Time) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	if err := s.calendar.Validate(day); err != nil {
		panic(err)
	}
	linear := s.rsi.Update(price)
	s.lastValue = s.delay.Update(linear)
	s.samplesSeen++
	s.calendar.Commit(day)
	return Snapshot{s.lastValue, linear}
}
