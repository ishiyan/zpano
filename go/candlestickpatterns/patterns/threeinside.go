package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// ThreeInside: a three-candle reversal pattern.
//
// Three Inside Up (bullish):
//   - first candle: long black,
//   - second candle: short, engulfed by the first candle's real body,
//   - third candle: white, closes above the first candle's open.
//
// Three Inside Down (bearish):
//   - first candle: long white,
//   - second candle: short, engulfed by the first candle's real body,
//   - third candle: black, closes below the first candle's open.
//
// The meaning of "long" is specified with LongBody.
// The meaning of "short" is specified with ShortBody.
//
// Category C: both branches evaluated, return stronger signal.
func ThreeInside(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.LongBody, cp.ShortBody) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Shared fuzzy conditions.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 3)
	muShort2 := cp.MuLess(core.RealBodyLen(b2.O, b2.C), cp.ShortBody, 2)

	// Fuzzy containment: 1st body encloses 2nd body.
	eqAvg := cp.AvgCS(cp.Equal, 2)
	eqWidth := 0.0
	if eqAvg > 0.0 {
		eqWidth = cp.FuzzRatio * eqAvg
	}
	muEncUpper := cp.MuGeRaw(max(b1.O, b1.C), max(b2.O, b2.C), eqWidth)
	muEncLower := cp.MuLtRaw(min(b1.O, b1.C), min(b2.O, b2.C), eqWidth)

	// Three Inside Up: long black, short engulfed, white closes above 1st open.
	bullSignal := 0.0
	if core.IsBlack(b1.O, b1.C) && core.IsWhite(b3.O, b3.C) {
		rb1 := core.RealBodyLen(b1.O, b1.C)
		width := 0.0
		if rb1 > 0.0 {
			width = cp.FuzzRatio * rb1
		}
		muCloseAbove := cp.MuGtRaw(b3.C, b1.O, width)
		conf := fuzzy.TProductAll(muLong1, muShort2, muEncUpper, muEncLower, muCloseAbove)
		bullSignal = conf * 100.0
	}

	// Three Inside Down: long white, short engulfed, black closes below 1st open.
	bearSignal := 0.0
	if core.IsWhite(b1.O, b1.C) && core.IsBlack(b3.O, b3.C) {
		rb1 := core.RealBodyLen(b1.O, b1.C)
		width := 0.0
		if rb1 > 0.0 {
			width = cp.FuzzRatio * rb1
		}
		muCloseBelow := cp.MuLtRaw(b3.C, b1.O, width)
		conf := fuzzy.TProductAll(muLong1, muShort2, muEncUpper, muEncLower, muCloseBelow)
		bearSignal = -conf * 100.0
	}

	if math.Abs(bullSignal) >= math.Abs(bearSignal) {
		return bullSignal
	}
	return bearSignal
}
