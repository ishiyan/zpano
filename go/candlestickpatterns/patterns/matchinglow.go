package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
)

// MatchingLow detects the Matching Low pattern (2-candle bullish).
//
// Must have:
//   - first candle: black,
//   - second candle: black with close equal to the first candle's close.
//
// The meaning of "equal" is specified with cp.Equal.
//
// Category A: always bullish (continuous).
func MatchingLow(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.Equal) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	// Crisp gates: both black.
	if !(core.IsBlack(b1.O, b1.C) && core.IsBlack(b2.O, b2.C)) {
		return 0.0
	}

	// Fuzzy: close equal to prior close (two-sided band).
	muEq := cp.MuLess(math.Abs(b2.C-b1.C), cp.Equal, 2)
	return muEq * 100.0
}
