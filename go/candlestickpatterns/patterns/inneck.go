package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// InNeck detects the In Neck pattern (2-candle bearish continuation).
//
// Must have:
//   - first candle: long black,
//   - second candle: white, opens below the prior low, closes slightly
//     into the prior real body (close near the prior close).
//
// The meaning of "long" is specified with cp.LongBody.
// The meaning of "near" is specified with cp.Near.
//
// Category A: always bearish (continuous).
func InNeck(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.LongBody, cp.Near) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	// Crisp gates: color checks and open below prior low.
	if !(core.IsBlack(b1.O, b1.C) && core.IsWhite(b2.O, b2.C) && b2.O < b1.L) {
		return 0.0
	}

	// Fuzzy conditions.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 2)
	// Close near prior close: mu_less(abs_diff, near_avg) — crossover at near boundary.
	muNearClose := cp.MuLess(math.Abs(b2.C-b1.C), cp.Near, 1)

	confidence := fuzzy.TProductAll(muLong1, muNearClose)
	return -confidence * 100.0
}
