package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Breakaway detects a five-candle reversal pattern.
//
// Bullish: first candle is long black, second candle is black gapping down,
// third and fourth candles have consecutively lower highs and lows, fifth
// candle is white closing into the gap (between first and second candle's
// real bodies).
//
// Bearish: mirror image with colors reversed and gaps reversed.
//
// Category C: both branches evaluated, return stronger signal.
func Breakaway(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(5, cp.LongBody) {
		return 0.0
	}

	b1 := cp.Bar(5)
	b2 := cp.Bar(4)
	b3 := cp.Bar(3)
	b4 := cp.Bar(2)
	b5 := cp.Bar(1)

	// Fuzzy: 1st candle is long.
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 5)

	// Bullish breakaway.
	bullSignal := 0.0
	if core.IsBlack(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) &&
		core.IsBlack(b4.O, b4.C) && core.IsWhite(b5.O, b5.C) &&
		b3.H < b2.H && b3.L < b2.L &&
		b4.H < b3.H && b4.L < b3.L &&
		core.IsRealBodyGapDown(b1.O, b1.C, b2.O, b2.C) {
		rb1 := core.RealBodyLen(b1.O, b1.C)
		width := cp.FuzzRatio * rb1
		if rb1 <= 0.0 {
			width = 0.0
		}
		// Fuzzy: c5 > o2 and c5 < c1 (closing into the gap).
		muC5AboveO2 := cp.MuGtRaw(b5.C, b2.O, width)
		muC5BelowC1 := cp.MuLtRaw(b5.C, b1.C, width)
		conf := fuzzy.TProductAll(muLong1, muC5AboveO2, muC5BelowC1)
		bullSignal = conf * 100.0
	}

	// Bearish breakaway.
	bearSignal := 0.0
	if core.IsWhite(b1.O, b1.C) && core.IsWhite(b2.O, b2.C) &&
		core.IsWhite(b4.O, b4.C) && core.IsBlack(b5.O, b5.C) &&
		b3.H > b2.H && b3.L > b2.L &&
		b4.H > b3.H && b4.L > b3.L &&
		core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) {
		rb1 := core.RealBodyLen(b1.O, b1.C)
		width := cp.FuzzRatio * rb1
		if rb1 <= 0.0 {
			width = 0.0
		}
		muC5BelowO2 := cp.MuLtRaw(b5.C, b2.O, width)
		muC5AboveC1 := cp.MuGtRaw(b5.C, b1.C, width)
		conf := fuzzy.TProductAll(muLong1, muC5BelowO2, muC5AboveC1)
		bearSignal = -conf * 100.0
	}

	if math.Abs(bullSignal) >= math.Abs(bearSignal) {
		return bullSignal
	}
	return bearSignal
}
