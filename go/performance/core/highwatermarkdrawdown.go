package core

import (
	"math"

	"zpano/streamingkbn"
)

// HighWaterMarkDrawdown computes rolling high-water-mark drawdowns.
//
// Drawdown at each observation is measured from the high-water mark, the
// highest equity value reached up to that observation within the current
// rolling window, including the equity at the start of the window:
//
//	drawdown_t = equity_t / max(equity_start, equity_1, ..., equity_t) - 1
//
// This matches R PerformanceAnalytics Drawdowns(), which uses
// cummax(c(1, cumprod(1 + R))): a first negative return already produces
// a drawdown. For a rolling window, equity_start is the equity just
// before the first observation in the window, so the result equals a
// fresh calculation over the window's returns.
//
// Drawdowns are expressed as decimals and are non-positive.
//
// Cumulative log-equity is maintained internally so that returns can be
// accumulated accurately. Running sums of drawdowns and squared drawdowns
// are maintained using compensated floating-point accumulation.
//
// When an observation leaves the window, the window's starting equity
// changes. The drawdowns in the window are recomputed only when this
// changes their high-water marks; otherwise, the update is O(1).
//
// A window size of zero means an expanding (unbounded) window.
type HighWaterMarkDrawdown struct {
	windowSize int

	// Cumulative log-equity at each observation, oldest first.
	cumlog []float64

	// Drawdown at each observation, as a decimal (<= 0), oldest first.
	dd []float64

	// Cumulative log return.
	c streamingkbn.KleinKBNAccumulator

	// Log-equity just before the first observation in the window.
	base float64

	// Current high-water mark in log-equity space.
	peak float64

	// Running drawdown aggregates.
	sumDD  streamingkbn.KleinKBNAccumulator
	sumDD2 streamingkbn.KleinKBNAccumulator
}

// NewHighWaterMarkDrawdown returns a new empty HighWaterMarkDrawdown.
// A non-positive windowSize means an expanding (unbounded) window.
func NewHighWaterMarkDrawdown(windowSize int) *HighWaterMarkDrawdown {
	if windowSize < 0 {
		windowSize = 0
	}
	return &HighWaterMarkDrawdown{windowSize: windowSize}
}

// WindowSize returns the rolling window size (0 for an expanding window).
func (h *HighWaterMarkDrawdown) WindowSize() int {
	return h.windowSize
}

// Reset resets the accumulator to its initial empty state.
func (h *HighWaterMarkDrawdown) Reset() {
	h.cumlog = h.cumlog[:0]
	h.dd = h.dd[:0]
	h.sumDD.Reset()
	h.sumDD2.Reset()
	h.c.Reset()
	h.base = 0.0
	h.peak = 0.0
}

// recompute recomputes all drawdowns from the cumulative log-equity values.
//
// This is required when an observation leaving the rolling window changes
// the high-water marks of the remaining observations.
func (h *HighWaterMarkDrawdown) recompute() {
	h.dd = h.dd[:0]
	h.sumDD.Reset()
	h.sumDD2.Reset()
	peak := h.base
	for _, c := range h.cumlog {
		var dd float64
		if c >= peak {
			peak = c
			dd = 0.0
		} else {
			dd = math.Expm1(c - peak)
		}
		h.dd = append(h.dd, dd)
		h.sumDD.Update(dd)
		h.sumDD2.Update(dd * dd)
	}
	h.peak = peak
}

// Update adds a return observation, expressed as a decimal (for example
// 0.02 for a 2% return and -0.015 for a -1.5% return).
//
// If the rolling window is full, the oldest observation is removed before
// the new observation is added.
//
// Returns true if the rolling window required a drawdown recomputation,
// otherwise false.
func (h *HighWaterMarkDrawdown) Update(ret float64) bool {
	hasOldBase := false
	oldBase := 0.0
	if h.windowSize > 0 && len(h.cumlog) == h.windowSize {
		oldC := h.cumlog[0]
		oldDD := h.dd[0]
		h.cumlog = h.cumlog[1:]
		h.dd = h.dd[1:]

		h.sumDD.Revert(oldDD)
		h.sumDD2.Revert(oldDD * oldDD)

		// The evicted observation's equity is the new starting equity.
		// High-water marks of the remaining observations can only
		// change if the old starting equity was above the evicted one.
		if oldC < h.base {
			hasOldBase = true
			oldBase = h.base
		}
		h.base = oldC
	}

	// Global cumulative log-equity.
	h.c.Update(math.Log1p(ret))
	c := h.c.Value()
	h.cumlog = append(h.cumlog, c)

	// Peaks of all remaining observations were max(oldBase, c0, ..., cj);
	// without oldBase they are max(c0, c1, ..., cj). They differ only if
	// the new first observation is also below oldBase.
	if hasOldBase && h.cumlog[0] < oldBase {
		h.recompute()
		return true
	}

	var dd float64
	if c >= h.peak {
		h.peak = c
		dd = 0.0
	} else {
		dd = math.Expm1(c - h.peak)
	}

	h.dd = append(h.dd, dd)
	h.sumDD.Update(dd)
	h.sumDD2.Update(dd * dd)
	return false
}

// Drawdowns returns the drawdowns for observations currently in the
// window, oldest first.
//
// Returns the internal slice without a defensive copy; callers must not
// modify it, and it is only valid until the next call to Update or Reset.
func (h *HighWaterMarkDrawdown) Drawdowns() []float64 {
	return h.dd
}

// Drawdown returns the most recent drawdown in the current window
// (NaN when empty).
func (h *HighWaterMarkDrawdown) Drawdown() float64 {
	if len(h.dd) > 0 {
		return h.dd[len(h.dd)-1]
	}
	return math.NaN()
}

// MaximumDrawdown returns the maximum drawdown in the current window
// (NaN when empty).
//
// Drawdowns are non-positive, so the largest loss is the minimum
// drawdown value.
func (h *HighWaterMarkDrawdown) MaximumDrawdown() float64 {
	if len(h.dd) == 0 {
		return math.NaN()
	}
	m := h.dd[0]
	for _, v := range h.dd[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

// DrawdownsMean returns the arithmetic mean of drawdowns in the current
// window (NaN when empty).
func (h *HighWaterMarkDrawdown) DrawdownsMean() float64 {
	n := len(h.dd)
	if n > 0 {
		return h.sumDD.Value() / float64(n)
	}
	return math.NaN()
}

// DrawdownsSquaredMean returns the mean squared drawdown in the current
// window (NaN when empty).
func (h *HighWaterMarkDrawdown) DrawdownsSquaredMean() float64 {
	n := len(h.dd)
	if n > 0 {
		return h.sumDD2.Value() / float64(n)
	}
	return math.NaN()
}

// DrawdownsCount returns the number of observations in the current window.
func (h *HighWaterMarkDrawdown) DrawdownsCount() int {
	return len(h.dd)
}
