package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// BeltHold detects a one-candle pattern.
//
// A long candle with a very short shadow on the opening side:
//   - bullish: long white candle with very short lower shadow,
//   - bearish: long black candle with very short upper shadow.
//
// The meaning of "long" is specified with LongBody.
// The meaning of "very short" for shadows is specified with VeryShortShadow.
//
// Category C: both branches evaluated, return stronger signal.
func BeltHold(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(1, cp.LongBody, cp.VeryShortShadow) {
		return 0.0
	}

	b := cp.Bar(1)
	muLong := cp.MuGreater(core.RealBodyLen(b.O, b.C), cp.LongBody, 1)

	// Bullish: white + very short lower shadow.
	bullSignal := 0.0
	if core.IsWhite(b.O, b.C) {
		muVS := cp.MuLess(core.LowerShadow(b.O, b.L, b.C), cp.VeryShortShadow, 1)
		conf := fuzzy.TProductAll(muLong, muVS)
		bullSignal = conf * 100.0
	}

	// Bearish: black + very short upper shadow.
	bearSignal := 0.0
	if core.IsBlack(b.O, b.C) {
		muVS := cp.MuLess(core.UpperShadow(b.O, b.H, b.C), cp.VeryShortShadow, 1)
		conf := fuzzy.TProductAll(muLong, muVS)
		bearSignal = -conf * 100.0
	}

	if math.Abs(bullSignal) >= math.Abs(bearSignal) {
		return bullSignal
	}
	return bearSignal
}
