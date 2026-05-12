package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// ThreeOutside: a three-candle reversal pattern.
//
// Must have:
//   - first and second candles form an engulfing pattern,
//   - third candle confirms the direction by closing higher (up) or
//     lower (down).
//
// Three Outside Up: first candle is black, second is white engulfing
// the first, third closes higher than the second.
//
// Three Outside Down: first candle is white, second is black engulfing
// the first, third closes lower than the second.
//
// Category C: both branches evaluated, return stronger signal.
func ThreeOutside(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Fuzzy engulfment width.
	eqAvg := cp.AvgCS(cp.Equal, 1)
	eqWidth := 0.0
	if eqAvg > 0.0 {
		eqWidth = cp.FuzzRatio * eqAvg
	}

	// Three Outside Up: black + white engulfing + 3rd closes higher.
	bullSignal := 0.0
	if core.IsBlack(b1.O, b1.C) && core.IsWhite(b2.O, b2.C) {
		muEncUpper := cp.MuGeRaw(max(b2.O, b2.C), max(b1.O, b1.C), eqWidth)
		muEncLower := cp.MuLtRaw(min(b2.O, b2.C), min(b1.O, b1.C), eqWidth)
		rb2 := core.RealBodyLen(b2.O, b2.C)
		width := 0.0
		if rb2 > 0.0 {
			width = cp.FuzzRatio * rb2
		}
		muCloseHigher := cp.MuGtRaw(b3.C, b2.C, width)
		conf := fuzzy.TProductAll(muEncUpper, muEncLower, muCloseHigher)
		bullSignal = conf * 100.0
	}

	// Three Outside Down: white + black engulfing + 3rd closes lower.
	bearSignal := 0.0
	if core.IsWhite(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) {
		muEncUpper := cp.MuGeRaw(max(b2.O, b2.C), max(b1.O, b1.C), eqWidth)
		muEncLower := cp.MuLtRaw(min(b2.O, b2.C), min(b1.O, b1.C), eqWidth)
		rb2 := core.RealBodyLen(b2.O, b2.C)
		width := 0.0
		if rb2 > 0.0 {
			width = cp.FuzzRatio * rb2
		}
		muCloseLower := cp.MuLtRaw(b3.C, b2.C, width)
		conf := fuzzy.TProductAll(muEncUpper, muEncLower, muCloseLower)
		bearSignal = -conf * 100.0
	}

	if math.Abs(bullSignal) >= math.Abs(bearSignal) {
		return bullSignal
	}
	return bearSignal
}
