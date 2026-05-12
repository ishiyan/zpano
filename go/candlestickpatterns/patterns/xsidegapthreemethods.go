package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// XSideGapThreeMethods: a three-candle continuation pattern.
//
// Must have:
//   - first and second candles are the same color with a gap between them,
//   - third candle is opposite color, opens within the second candle's
//     real body and closes within the first candle's real body (fills the
//     gap).
//
// Upside gap: two white candles with gap up, third is black = bullish.
// Downside gap: two black candles with gap down, third is white = bearish.
//
// Category C: both branches evaluated, return stronger signal.
func XSideGapThreeMethods(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Upside gap: two whites gap up, third black fills.
	bullSignal := 0.0
	if core.IsWhite(b1.O, b1.C) && core.IsWhite(b2.O, b2.C) && core.IsBlack(b3.O, b3.C) &&
		core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) {
		rb2 := core.RealBodyLen(b2.O, b2.C)
		width := 0.0
		if rb2 > 0.0 {
			width = cp.FuzzRatio * rb2
		}
		// o3 within 2nd body: o3 < c2 and o3 > o2
		muO3LtC2 := cp.MuLtRaw(b3.O, b2.C, width)
		muO3GtO2 := cp.MuGtRaw(b3.O, b2.O, width)
		// c3 within 1st body: c3 > o1 and c3 < c1
		rb1 := core.RealBodyLen(b1.O, b1.C)
		width1 := 0.0
		if rb1 > 0.0 {
			width1 = cp.FuzzRatio * rb1
		}
		muC3GtO1 := cp.MuGtRaw(b3.C, b1.O, width1)
		muC3LtC1 := cp.MuLtRaw(b3.C, b1.C, width1)
		conf := fuzzy.TProductAll(muO3LtC2, muO3GtO2, muC3GtO1, muC3LtC1)
		bullSignal = conf * 100.0
	}

	// Downside gap: two blacks gap down, third white fills.
	bearSignal := 0.0
	if core.IsBlack(b1.O, b1.C) && core.IsBlack(b2.O, b2.C) && core.IsWhite(b3.O, b3.C) &&
		core.IsRealBodyGapDown(b1.O, b1.C, b2.O, b2.C) {
		rb2 := core.RealBodyLen(b2.O, b2.C)
		width := 0.0
		if rb2 > 0.0 {
			width = cp.FuzzRatio * rb2
		}
		// o3 within 2nd body: o3 > c2 and o3 < o2
		muO3GtC2 := cp.MuGtRaw(b3.O, b2.C, width)
		muO3LtO2 := cp.MuLtRaw(b3.O, b2.O, width)
		// c3 within 1st body: c3 < o1 and c3 > c1
		rb1 := core.RealBodyLen(b1.O, b1.C)
		width1 := 0.0
		if rb1 > 0.0 {
			width1 = cp.FuzzRatio * rb1
		}
		muC3LtO1 := cp.MuLtRaw(b3.C, b1.O, width1)
		muC3GtC1 := cp.MuGtRaw(b3.C, b1.C, width1)
		conf := fuzzy.TProductAll(muO3GtC2, muO3LtO2, muC3LtO1, muC3GtC1)
		bearSignal = -conf * 100.0
	}

	if math.Abs(bullSignal) >= math.Abs(bearSignal) {
		return bullSignal
	}
	return bearSignal
}
