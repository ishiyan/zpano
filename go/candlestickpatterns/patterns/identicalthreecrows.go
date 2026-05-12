package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// IdenticalThreeCrows detects the Identical Three Crows pattern (3-candle bearish).
//
// Must have:
//   - three consecutive declining black candles,
//   - each opens very close to the prior candle's close (equal criterion),
//   - very short lower shadows.
//
// The meaning of "equal" is specified with cp.Equal.
// The meaning of "very short" for shadows is specified with cp.VeryShortShadow.
//
// Category A: always bearish (continuous).
func IdenticalThreeCrows(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.Equal, cp.VeryShortShadow) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: all black, declining closes.
	if !(core.IsBlack(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) && core.IsBlack(b3.O, b3.C)) {
		return 0.0
	}
	if !(b1.C > b2.C && b2.C > b3.C) {
		return 0.0
	}

	// Fuzzy conditions.
	muLS1 := cp.MuLess(core.LowerShadow(b1.O, b1.L, b1.C), cp.VeryShortShadow, 3)
	muLS2 := cp.MuLess(core.LowerShadow(b2.O, b2.L, b2.C), cp.VeryShortShadow, 2)
	muLS3 := cp.MuLess(core.LowerShadow(b3.O, b3.L, b3.C), cp.VeryShortShadow, 1)
	// Opens near prior close (equal criterion, two-sided band).
	muEq2 := cp.MuLess(math.Abs(b2.O-b1.C), cp.Equal, 3)
	muEq3 := cp.MuLess(math.Abs(b3.O-b2.C), cp.Equal, 2)

	confidence := fuzzy.TProductAll(muLS1, muLS2, muLS3, muEq2, muEq3)
	return -confidence * 100.0
}
