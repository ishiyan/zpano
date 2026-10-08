package core

import (
	"zpano/streamingkbn"
)

// WinLoss computes streaming winning/losing return sums, means and counts.
//
// The zero value is ready to use.
type WinLoss struct {
	nonZeroSum streamingkbn.KleinKBNSummator
	winSum     streamingkbn.KleinKBNSummator
	lossSum    streamingkbn.KleinKBNSummator
}

// NewWinLoss returns a new empty WinLoss.
func NewWinLoss() *WinLoss {
	return &WinLoss{}
}

// Reset clears all accumulated state.
func (w *WinLoss) Reset() {
	w.nonZeroSum.Reset()
	w.winSum.Reset()
	w.lossSum.Reset()
}

// Revert removes a previously added return.
func (w *WinLoss) Revert(ret float64) {
	if ret != 0 {
		w.nonZeroSum.Revert(ret)
	}
	if ret > 0 {
		w.winSum.Revert(ret)
	}
	if ret < 0 {
		w.lossSum.Revert(ret)
	}
}

// Update adds a return.
func (w *WinLoss) Update(ret float64) {
	if ret != 0 {
		w.nonZeroSum.Update(ret)
	}
	if ret > 0 {
		w.winSum.Update(ret)
	}
	if ret < 0 {
		w.lossSum.Update(ret)
	}
}

// NonZeroReturnsMean returns the arithmetic mean (average) of non-zero
// returns (NaN when there are none).
func (w *WinLoss) NonZeroReturnsMean() float64 { return w.nonZeroSum.Mean() }

// NonZeroReturnsCount returns the number of non-zero returns.
func (w *WinLoss) NonZeroReturnsCount() int { return w.nonZeroSum.N() }

// WinningReturnsSum returns the sum of winning (positive) returns.
func (w *WinLoss) WinningReturnsSum() float64 { return w.winSum.Value() }

// WinningReturnsMean returns the arithmetic mean (average) of winning
// (positive) returns (NaN when there are none).
func (w *WinLoss) WinningReturnsMean() float64 { return w.winSum.Mean() }

// WinningReturnsCount returns the number of winning (positive) returns.
func (w *WinLoss) WinningReturnsCount() int { return w.winSum.N() }

// LosingReturnsSum returns the sum of losing (negative) returns.
func (w *WinLoss) LosingReturnsSum() float64 { return w.lossSum.Value() }

// LosingReturnsMean returns the arithmetic mean (average) of losing
// (negative) returns (NaN when there are none).
func (w *WinLoss) LosingReturnsMean() float64 { return w.lossSum.Mean() }

// LosingReturnsCount returns the number of losing (negative) returns.
func (w *WinLoss) LosingReturnsCount() int { return w.lossSum.N() }
