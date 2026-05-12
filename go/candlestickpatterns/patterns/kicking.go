package patterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// Kicking detects the Kicking pattern (2-candle).
//
// Must have:
//   - first candle: marubozu (long body, very short shadows),
//   - second candle: opposite-color marubozu with a high-low gap,
//   - bullish: black marubozu followed by white marubozu gapping up,
//   - bearish: white marubozu followed by black marubozu gapping down.
//
// Category B: direction from second candle's color.
func Kicking(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(2, cp.VeryShortShadow, cp.LongBody) {
		return 0.0
	}

	b1 := cp.Bar(2)
	b2 := cp.Bar(1)

	color1 := 1
	if b1.C < b1.O {
		color1 = -1
	}
	color2 := 1
	if b2.C < b2.O {
		color2 = -1
	}
	// Crisp: opposite colors.
	if color1 == color2 {
		return 0.0
	}

	// Crisp: gap check.
	if color1 == -1 && !core.IsHighLowGapUp(b1.H, b2.L) {
		return 0.0
	}
	if color1 == 1 && !core.IsHighLowGapDown(b1.L, b2.H) {
		return 0.0
	}

	// Fuzzy: both are marubozu (long body, very short shadows).
	muLong1 := cp.MuGreater(core.RealBodyLen(b1.O, b1.C), cp.LongBody, 2)
	muVSUS1 := cp.MuLess(core.UpperShadow(b1.O, b1.H, b1.C), cp.VeryShortShadow, 2)
	muVSLS1 := cp.MuLess(core.LowerShadow(b1.O, b1.L, b1.C), cp.VeryShortShadow, 2)
	muLong2 := cp.MuGreater(core.RealBodyLen(b2.O, b2.C), cp.LongBody, 1)
	muVSUS2 := cp.MuLess(core.UpperShadow(b2.O, b2.H, b2.C), cp.VeryShortShadow, 1)
	muVSLS2 := cp.MuLess(core.LowerShadow(b2.O, b2.L, b2.C), cp.VeryShortShadow, 1)

	confidence := fuzzy.TProductAll(muLong1, muVSUS1, muVSLS1, muLong2, muVSUS2, muVSLS2)
	return float64(color2) * confidence * 100.0
}
