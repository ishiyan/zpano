package patterns

import (
	"math"

	"zpano/candlestickpatterns/core"
	"zpano/fuzzy"
)

// UpDownGapSideBySideWhiteLines: a three-candle pattern.
//
// Must have:
//   - first candle: white (for up gap) or black (for down gap),
//   - gap (up or down) between the first and second candle — both 2nd AND
//     3rd must gap from the 1st,
//   - second and third candles are both white with similar size and
//     approximately the same open.
//
// Up gap = bullish continuation, down gap = bearish continuation.
//
// Category C: both branches evaluated, return stronger signal.
func UpDownGapSideBySideWhiteLines(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3, cp.Near, cp.Equal) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Crisp: both 2nd and 3rd must be white.
	if !(core.IsWhite(b2.O, b2.C) && core.IsWhite(b3.O, b3.C)) {
		return 0.0
	}

	// Both 2nd and 3rd must gap from 1st in the same direction — crisp.
	gapUp := core.IsRealBodyGapUp(b1.O, b1.C, b2.O, b2.C) && core.IsRealBodyGapUp(b1.O, b1.C, b3.O, b3.C)
	gapDown := core.IsRealBodyGapDown(b1.O, b1.C, b2.O, b2.C) && core.IsRealBodyGapDown(b1.O, b1.C, b3.O, b3.C)

	if !(gapUp || gapDown) {
		return 0.0
	}

	rb2 := core.RealBodyLen(b2.O, b2.C)
	rb3 := core.RealBodyLen(b3.O, b3.C)

	// Fuzzy: similar size and same open.
	muNearSize := cp.MuLess(math.Abs(rb2-rb3), cp.Near, 2)
	muEqualOpen := cp.MuLess(math.Abs(b3.O-b2.O), cp.Equal, 2)

	conf := fuzzy.TProductAll(muNearSize, muEqualOpen)

	if gapUp {
		return conf * 100.0
	}
	return -conf * 100.0
}
