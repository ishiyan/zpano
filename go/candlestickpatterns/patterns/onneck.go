package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// OnNeck detects an On Neck pattern: a two-candle bearish continuation pattern.
//
// Must have:
//   - first candle: long black,
//   - second candle: white that opens below the prior low and closes
//     equal to the prior candle's low.
//
// The meaning of "long" is specified with LongBody.
// The meaning of "equal" is specified with Equal.
//
// Category A: always bearish (continuous).
func OnNeck(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.LongBody, cp.Equal) {
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

	// Close equal to prior low: crisp was abs(c2-l1) <= eq.
	// Model as mu_less(abs_diff, eq_avg) — crossover at eq boundary.
	muNearLow := cp.MuLess(math.Abs(b2.C-b1.L), cp.Equal, 2)

	confidence := fuzzy.TProductAll(muLong1, muNearLow)

	return -1.0 * confidence * 100.0
}
