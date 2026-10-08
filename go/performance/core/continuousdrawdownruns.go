package core

import (
	"math"

	"zpano/streamingkbn"
)

// DdPercent converts a compounded log return into a percentage drawdown,
// expm1(logsum)·100.
func DdPercent(logsum float64) float64 {
	return math.Expm1(logsum) * 100.0
}

// drawdownRun is one continuous losing run: its log-sum and length.
type drawdownRun struct {
	logsum float64
	count  int
}

// ContinuousDrawdownRuns tracks streaming 'continuous' drawdown runs for
// Burke-type measures.
//
// A continuous drawdown is the compounded loss over a maximal run of
// consecutive negative returns. Following PerformanceAnalytics
// BurkeRatio, returns are compounded as if they were percentages:
//
//	DD = (prod(1 + r_i * 0.01) - 1) * 100
//
// For decimal returns this is close to the sum of the run's returns,
// not their compounded return; the quirk is kept to match R.
//
// The Burke denominator is
//
//	sqrt(sum(DD_j^2))
//
// where the sum is taken over all continuous losing runs in the current
// window.
//
// This type is a pure accumulator: the caller owns the rolling window
// and feeds evicted values to Revert and new values to Update. Within one
// step, call Revert(old) BEFORE Update(new) so run adjacency stays
// correct.
//
// The complexity is O(1) per call. The zero value is ready to use.
type ContinuousDrawdownRuns struct {
	// Runs, oldest first.
	runs []drawdownRun
	// Sum of squared continuous drawdowns.
	sumSq streamingkbn.KleinKBNAccumulator
	// Whether the most recent return is negative.
	lastWasNegative bool
}

// NewContinuousDrawdownRuns returns a new empty ContinuousDrawdownRuns.
func NewContinuousDrawdownRuns() *ContinuousDrawdownRuns {
	return &ContinuousDrawdownRuns{}
}

// Reset clears all accumulated state.
func (c *ContinuousDrawdownRuns) Reset() {
	c.runs = c.runs[:0]
	c.sumSq.Reset()
	c.lastWasNegative = false
}

// Revert removes the oldest return from the left edge of the window.
func (c *ContinuousDrawdownRuns) Revert(oldRet float64) {
	if oldRet < 0 {
		// The oldest negative is the front of the left-most run, shrink it.
		run := &c.runs[0]
		d := DdPercent(run.logsum)
		c.sumSq.Revert(d * d)
		run.logsum -= math.Log1p(oldRet * 0.01)
		run.count--
		if run.count == 0 {
			c.runs = c.runs[1:] // run fully evicted
		} else {
			d = DdPercent(run.logsum)
			c.sumSq.Update(d * d)
		}
	}
	// oldRet >= 0 is a separator, nothing to update.
}

// Update adds a new (most recent) return at the right edge of the window.
func (c *ContinuousDrawdownRuns) Update(ret float64) {
	if ret < 0 {
		logr := math.Log1p(ret * 0.01)
		if c.lastWasNegative && len(c.runs) > 0 {
			// Extend the currently-open (right-most) run.
			run := &c.runs[len(c.runs)-1]
			d := DdPercent(run.logsum)
			c.sumSq.Revert(d * d)
			run.logsum += logr
			run.count++
			d = DdPercent(run.logsum)
			c.sumSq.Update(d * d)
		} else {
			// Start a new run.
			c.runs = append(c.runs, drawdownRun{logsum: logr, count: 1})
			d := DdPercent(logr)
			c.sumSq.Update(d * d)
		}
		c.lastWasNegative = true
	} else {
		// A non-negative return closes any open run (already counted).
		c.lastWasNegative = false
	}
}

// Drawdowns returns the continuous drawdowns (negative percentages), one
// value for each losing run, oldest first. The result is a newly
// allocated slice.
func (c *ContinuousDrawdownRuns) Drawdowns() []float64 {
	out := make([]float64, len(c.runs))
	for i, r := range c.runs {
		out[i] = DdPercent(r.logsum)
	}
	return out
}

// SumDrawdownsSquared returns the sum of squared continuous drawdowns
// (clamped at zero).
func (c *ContinuousDrawdownRuns) SumDrawdownsSquared() float64 {
	return c.clampedSumSq()
}

// clampedSumSq mirrors Python's max(value, 0.0).
func (c *ContinuousDrawdownRuns) clampedSumSq() float64 {
	v := c.sumSq.Value()
	if 0.0 > v {
		return 0.0
	}
	return v
}

// SqrtSumDrawdownsSquared returns the square root of the sum of squared
// continuous drawdowns. This is the denominator used by the Burke ratio.
func (c *ContinuousDrawdownRuns) SqrtSumDrawdownsSquared() float64 {
	return math.Sqrt(c.clampedSumSq())
}

// RunCount returns the number of continuous losing runs in the current
// window.
func (c *ContinuousDrawdownRuns) RunCount() int {
	return len(c.runs)
}
