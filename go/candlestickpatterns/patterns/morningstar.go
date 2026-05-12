package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

const morningStarPenetrationFactor = 0.3

// MorningStar detects the Morning Star pattern (3-candle bullish reversal).
//
// Must have:
//   - first candle: long black real body,
//   - second candle: short real body that gaps down (real body gap down from
//     the first),
//   - third candle: white real body that closes well within the first candle's
//     real body.
//
// The meaning of "long" is specified with cp.LongBody.
// The meaning of "short" is specified with cp.ShortBody.
//
// Category A: always bullish (continuous).
func MorningStar(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.LongBody, cp.ShortBody) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp gates: color checks and gap.
	if !(core.IsBlack(b1.O, b1.C) &&
		core.IsRealBodyGapDown(b1.O, b1.C, b2.O, b2.C) &&
		core.IsWhite(b3.O, b3.C)) {
		return 0.0
	}

	// Fuzzy conditions.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 3)
	muShort2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.ShortBody, 2)

	// c3 > c1 + rb1 * penetration  →  c3 > threshold
	rb1 := core.RealBodyLen(b1.O, b1.C)
	threshold := b1.C + rb1*morningStarPenetrationFactor
	width := cp.FuzzRatio * rb1 * morningStarPenetrationFactor
	muPenetration := cp.MuGtRaw(b3.C, threshold, width)

	confidence := fuzzy.TProductAll(muLong1, muShort2, muPenetration)
	return confidence * 100.0
}
