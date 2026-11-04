// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package metonic

import (
	"fmt"
	"math"
	"time"

	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// Calendar enforces a consecutive daily grid, including weekends and holidays.
type Calendar struct {
	last        time.Time
	dated       bool
	initialized bool
}

// Validate rejects gaps, intraday timestamps and mixed dated/undated modes before state mutation.
func (c *Calendar) Validate(day time.Time) error {
	dated := !day.IsZero()
	if c.initialized && c.dated != dated {
		return fmt.Errorf("cannot mix dated and undated observations")
	}
	if dated {
		_, offset := day.Zone()
		if offset != 0 || day.Hour() != 0 || day.Minute() != 0 || day.Second() != 0 || day.Nanosecond() != 0 {
			return fmt.Errorf("use UTC calendar dates")
		}
		if c.initialized && !day.Equal(c.last.AddDate(0, 0, 1)) {
			return fmt.Errorf("calendar gap: fill full OHLC before update")
		}
	}
	return nil
}

// Commit records a successfully accepted date.
func (c *Calendar) Commit(day time.Time) { c.last = day; c.dated = !day.IsZero(); c.initialized = true }

// DailyRecord holds the complete daily payload so filling preserves OHLC.
type DailyRecord[T any] struct {
	Day     time.Time
	Payload T
}

// FillCalendarDays copies the preceding complete payload into missing calendar slots.
func FillCalendarDays[T any](records []DailyRecord[T]) ([]DailyRecord[T], error) {
	output := make([]DailyRecord[T], 0, len(records))
	for i, record := range records {
		if record.Day.IsZero() {
			return nil, fmt.Errorf("daily records need dates")
		}
		if i > 0 {
			previous := records[i-1]
			if !record.Day.After(previous.Day) {
				return nil, fmt.Errorf("dates must increase strictly")
			}
			for day := previous.Day.AddDate(0, 0, 1); day.Before(record.Day); day = day.AddDate(0, 0, 1) {
				output = append(output, DailyRecord[T]{day, previous.Payload})
			}
		}
		output = append(output, record)
	}
	return output, nil
}

// DelayMean retains NaN warmup slots in a bounded calendar ring.
type DelayMean struct {
	delays   []int
	history  []float64
	capacity int
}

// NewDelayMean builds a reusable delay-average engine.
func NewDelayMean(delays []int) *DelayMean {
	capacity := 0
	for _, delay := range delays {
		if delay < 1 {
			panic("delay must be positive")
		}
		capacity = max(capacity, delay+1)
	}
	return &DelayMean{append([]int{}, delays...), make([]float64, 0, capacity), capacity}
}

// Update advances one calendar slot, even if the underlying value is unavailable.
func (s *DelayMean) Update(value float64) float64 {
	s.history = harmoniccycle.Push(s.history, value, s.capacity)
	if len(s.history) < s.capacity {
		return math.NaN()
	}
	var sum float64
	for _, delay := range s.delays {
		value := s.history[len(s.history)-1-delay]
		if math.IsNaN(value) {
			return math.NaN()
		}
		sum += value
	}
	return sum / float64(len(s.delays))
}
