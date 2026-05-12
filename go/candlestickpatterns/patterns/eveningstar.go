package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

const eveningStarPenetrationFactor = 0.3

// EveningStar: a three-candle bearish reversal pattern.
//
// Must have:
// - first candle: long white real body,
// - second candle: short real body that gaps up (real body gap up from the
//   first),
// - third candle: black real body that moves well within the first candle's
//   real body.
//
// The meaning of "long" is specified with cp.LongBody.
// The meaning of "short" is specified with cp.ShortBody.
//
// Category A: always bearish (continuous).
func EveningStar(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.LongBody, cp.ShortBody) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: color checks and gap.
	if !(core.IsWhite(b1.O, b1.C) &&
		core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) &&
		core.IsBlack(b3.O, b3.C)) {
		return 0.0
	}

	// Fuzzy conditions.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 3)
	muShort2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.ShortBody, 2)

	// c3 < c1 - rb1 * penetration  →  c3 < threshold
	rb1 := core.RealBodyLen(b1.O, b1.C)
	threshold := b1.C - rb1*eveningStarPenetrationFactor
	width := cp.FuzzRatio * rb1 * eveningStarPenetrationFactor
	muPenetration := cp.MuLtRaw(b3.C, threshold, width)

	confidence := fuzzy.TProductAll(muLong1, muShort2, muPenetration)
	return -confidence * 100.0
}
