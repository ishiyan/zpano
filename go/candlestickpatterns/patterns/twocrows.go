package patterns

import (
	"zpano/candlestickpatterns/core"
)

// TwoCrows: a three-candle bearish pattern.
//
// Must have:
//   - first candle: long white,
//   - second candle: black, gaps up (real body gap up from the first),
//   - third candle: black, opens within the second candle's real body,
//     closes within the first candle's real body.
//
// The meaning of "long" is specified with LongBody.
//
// Category A: always bearish (continuous).
func TwoCrows(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.LongBody) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: colors.
	if !(core.IsWhite(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) && core.IsBlack(b3.O, b3.C)) {
		return 0.0
	}

	// Crisp: gap up.
	if !core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) {
		return 0.0
	}

	// Crisp: third opens within second body (o3 < o2 and o3 > c2).
	if !(b3.O < b2.O && b3.O > b2.C) {
		return 0.0
	}

	// Crisp: third closes within first body (c3 > o1 and c3 < c1).
	if !(b3.C > b1.O && b3.C < b1.C) {
		return 0.0
	}

	// Fuzzy: first candle is long.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 3)

	confidence := muLong1

	return -confidence * 100.0
}
