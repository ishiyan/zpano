package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

const darkCloudCoverPenetrationFactor = 0.5

// DarkCloudCover detects a two-candle bearish reversal pattern.
//
// Must have:
//   - first candle: long white candle,
//   - second candle: black candle that opens above the prior high and
//     closes well within the first candle's real body (below the midpoint).
func DarkCloudCover(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.LongBody) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	// Color checks stay crisp.
	if !core.IsWhite(b1.O, b1.C) || !core.IsBlack(b2.O, b2.C) {
		return 0.0
	}

	rb1 := core.RealBodyLen(b1.O, b1.C)
	eqAvg := cp.AvgCS(cp.Equal, 1)
	eqWidth := cp.FuzzRatio * eqAvg
	if eqAvg <= 0.0 {
		eqWidth = 0.0
	}

	muLong := cp.MuGreater(rb1, cp.LongBody, 2)
	muOpenAbove := cp.MuGtRaw(b2.O, b1.H, eqWidth)
	penThreshold := b1.C - rb1*darkCloudCoverPenetrationFactor
	penProduct := rb1 * darkCloudCoverPenetrationFactor
	penWidth := cp.FuzzRatio * penProduct
	if penProduct <= 0.0 {
		penWidth = 0.0
	}
	muPen := cp.MuLtRaw(b2.C, penThreshold, penWidth)
	muAboveOpen1 := cp.MuGtRaw(b2.C, b1.O, eqWidth)

	confidence := fuzzy.TProductAll(muLong, muOpenAbove, muPen, muAboveOpen1)
	return -confidence * 100.0
}
