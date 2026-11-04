// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package metonicsentimentindex

import (
	"math"
	"time"

	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/metonic"
	"zpano/indicators/larsvonthienen/zerolagexponentialmovingaverage"
)

// Snapshot owns the numerical diagnostics for one accepted observation.
type Snapshot struct {
	Value          float64
	DailySentiment float64
	SmoothedClose  float64
	SmoothedOpen   float64
}

func emptySnapshot() Snapshot { nan := math.NaN(); return Snapshot{nan, nan, nan, nan} }

type engine struct {
	p                           Params
	calendar                    metonic.Calendar
	delay                       *metonic.DelayMean
	lastValue                   float64
	samplesSeen                 int
	closeSmoother, openSmoother *zerolagexponentialmovingaverage.ZLEMA
	previousClose               float64
}

func (s *engine) reset() {
	s.calendar = metonic.Calendar{}
	s.delay = metonic.NewDelayMean([]int{s.p.Delay1, s.p.Delay2, s.p.Delay3})
	s.lastValue = math.NaN()
	s.samplesSeen = 0
	s.closeSmoother = zerolagexponentialmovingaverage.NewZLEMA(s.p.Length)
	s.openSmoother = zerolagexponentialmovingaverage.NewZLEMA(s.p.Length)
	s.previousClose = math.NaN()
}
func (s *engine) primed() bool { return !math.IsNaN(s.lastValue) }
func (s *engine) warmup() int {
	return max(s.p.Delay1, s.p.Delay2, s.p.Delay3) + (s.p.Length-1)/2 + s.p.Mode
}
func (s *engine) update(price float64, day time.Time, openPrice float64) Snapshot {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return emptySnapshot()
	}
	if s.p.Mode == 0 {
		harmoniccycle.Sample(openPrice)
		if math.IsNaN(openPrice) {
			return emptySnapshot()
		}
	}
	if err := s.calendar.Validate(day); err != nil {
		panic(err)
	}
	smoothedClose := s.closeSmoother.Update(price)
	smoothedOpen := math.NaN()
	comparison := s.previousClose
	if s.p.Mode == 0 {
		smoothedOpen = s.openSmoother.Update(openPrice)
		comparison = smoothedOpen
	}
	sentiment := math.NaN()
	if s.samplesSeen >= (s.p.Length-1)/2+s.p.Mode && !math.IsNaN(comparison) {
		sentiment = -1
		if smoothedClose > comparison {
			sentiment = 1
		}
	}
	s.previousClose = smoothedClose
	mean := s.delay.Update(sentiment)
	s.lastValue = 3 * mean
	s.samplesSeen++
	s.calendar.Commit(day)
	return Snapshot{s.lastValue, sentiment, smoothedClose, smoothedOpen}
}
