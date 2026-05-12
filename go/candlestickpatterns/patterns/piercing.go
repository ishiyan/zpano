package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Piercing detects a Piercing pattern: a two-candle bullish reversal pattern.
//
// Must have:
//   - first candle: long black,
//   - second candle: long white that opens below the prior low and closes
//     above the midpoint of the first candle's real body but within the body.
func Piercing(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.LongBody) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	// Color checks stay crisp.
	if !core.IsBlack(b1.O, b1.C) || !core.IsWhite(b2.O, b2.C) {
		return 0.0
	}

	rb1 := core.RealBodyLen(b1.O, b1.C)
	eqAvg := cp.AvgCS(cp.Equal, 1)
	eqWidth := 0.0
	if eqAvg > 0.0 {
		eqWidth = cp.FuzzRatio * eqAvg
	}

	muLong1 := cp.MuGreater(rb1, cp.LongBody, 2)
	muLong2 := cp.MuGreater(core.RealBodyLen(b2.O, b2.C), cp.LongBody, 1)
	muOpenBelow := cp.MuLtRaw(b2.O, b1.L, eqWidth)
	penThreshold := b1.C + rb1*0.5
	penWidth := 0.0
	if rb1 > 0.0 {
		penWidth = cp.FuzzRatio * rb1 * 0.5
	}
	muPen := cp.MuGtRaw(b2.C, penThreshold, penWidth)
	muBelowOpen1 := cp.MuLtRaw(b2.C, b1.O, eqWidth)

	confidence := fuzzy.TProductAll(muLong1, muLong2, muOpenBelow, muPen, muBelowOpen1)
	return confidence * 100.0
}
